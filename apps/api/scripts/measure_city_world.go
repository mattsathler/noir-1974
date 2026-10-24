package main

import (
	"context"
	"encoding/csv"
	"flag"
	"fmt"
	"log"
	"net/http"
	"net/http/httptest"
	"os"
	"path/filepath"
	"sort"
	"strconv"
	"strings"
	"sync"
	"time"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/joho/godotenv"
	"noir-1974/apps/api/internal/observability"
	"noir-1974/apps/api/internal/world"
)

type measurementSample struct {
	phase, status, elapsed, err string
	iteration                   int
}

func main() {
	scope := flag.String("scope", "city", "reconstruction to measure: city or world")
	cityID := flag.String("city", "fr/lyon", "city ID to reconstruct")
	serverID := flag.String("server", "", "world server ID; defaults to NOIR_WORLD_SERVER_ID")
	samples := flag.Int("samples", 100, "measured requests after warmups")
	warmups := flag.Int("warmups", 5, "warmup requests")
	csvPath := flag.String("csv", "", "optional path for raw sample CSV")
	coldOnly := flag.Bool("cold-only", false, "measure only the first cold request and print its SQL timings")
	coldRepeats := flag.Int("cold-repeats", 0, "measure this many cold cache misses using fresh store instances and return")
	concurrency := flag.Int("concurrency", 1, "number of concurrent workers for measured samples")
	prewarmPoolConnections := flag.Int("prewarm-pool-connections", 0, "open and release this many DB connections before measuring, without warming the world cache")
	flag.Parse()
	if *samples < 1 || *warmups < 0 || *coldRepeats < 0 || *concurrency < 1 {
		log.Fatal("samples and concurrency must be positive; warmups and cold-repeats cannot be negative")
	}
	for _, envFile := range []string{"../../.env.local", "../../.env"} {
		if err := godotenv.Load(envFile); err == nil {
			break
		}
	}
	databaseURL := os.Getenv("DATABASE_URL")
	if databaseURL == "" {
		log.Fatal("DATABASE_URL is required")
	}
	if *serverID == "" {
		*serverID = os.Getenv("NOIR_WORLD_SERVER_ID")
	}
	if *serverID == "" {
		*serverID = "00000000-0000-4000-8000-000000000001"
	}
	if *scope != "city" && *scope != "world" {
		log.Fatalf("unknown scope %q", *scope)
	}
	if *scope == "city" {
		if _, ok := world.CityByID(*cityID); !ok {
			log.Fatalf("unknown city %q", *cityID)
		}
	}
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Minute)
	defer cancel()
	poolConfig, err := pgxpool.ParseConfig(databaseURL)
	if err != nil {
		log.Fatal("parse database pool config: ", err)
	}
	recorder := observability.NewHTTPRecorder()
	poolConfig.ConnConfig.DefaultQueryExecMode = pgx.QueryExecModeExec
	poolConfig.ConnConfig.Tracer = observability.NewPGXQueryTracer(recorder)
	pool, err := pgxpool.NewWithConfig(ctx, poolConfig)
	if err != nil {
		log.Fatal("open database pool: ", err)
	}
	defer pool.Close()
	if err := pool.Ping(ctx); err != nil {
		log.Fatal("connect to database: ", err)
	}
	if *prewarmPoolConnections < 0 {
		log.Fatal("prewarm-pool-connections cannot be negative")
	}
	if *prewarmPoolConnections > 0 {
		connections := make([]*pgxpool.Conn, 0, *prewarmPoolConnections)
		for index := 0; index < *prewarmPoolConnections; index++ {
			connection, err := pool.Acquire(ctx)
			if err != nil {
				log.Fatal("prewarm database pool: ", err)
			}
			connections = append(connections, connection)
		}
		for _, connection := range connections {
			connection.Release()
		}
	}
	store := world.NewPostgresStore(pool)
	route := "/m14/city-world"
	if *scope == "world" {
		route = "/m14/world-for-cases"
	}
	handler := recorder.Middleware(func(*http.Request) string { return route }, http.HandlerFunc(func(writer http.ResponseWriter, request *http.Request) {
		var err error
		if *scope == "world" {
			_, err = store.WorldForCases(request.Context(), *serverID)
		} else {
			_, err = store.CityWorldForCases(request.Context(), *serverID, *cityID)
		}
		if err != nil {
			http.Error(writer, err.Error(), http.StatusInternalServerError)
			return
		}
		writer.WriteHeader(http.StatusOK)
	}))
	rows := make([]measurementSample, 0, *samples+*warmups+1)
	var rowsMu sync.Mutex
	request := func(phase string, iteration int) {
		started := time.Now()
		response := httptest.NewRecorder()
		handler.ServeHTTP(response, httptest.NewRequest(http.MethodGet, "/m14/city-world", nil))
		rowsMu.Lock()
		rows = append(rows, measurementSample{phase: phase, iteration: iteration, status: strconv.Itoa(response.Code), elapsed: strconv.FormatFloat(time.Since(started).Seconds(), 'f', 9, 64)})
		rowsMu.Unlock()
	}
	if *coldRepeats > 0 {
		for iteration := 1; iteration <= *coldRepeats; iteration++ {
			store = world.NewPostgresStore(pool)
			request("cold", iteration)
		}
		if *csvPath != "" {
			if err := writeCSV(*csvPath, rows); err != nil {
				log.Fatal("write CSV: ", err)
			}
		}
		values := make([]float64, 0, len(rows))
		for _, row := range rows {
			value, _ := strconv.ParseFloat(row.elapsed, 64)
			values = append(values, value)
		}
		sort.Float64s(values)
		mean := 0.0
		for _, value := range values {
			mean += value
		}
		mean /= float64(len(values))
		fmt.Printf("scope=%s phase=cold samples=%d min_ms=%.2f p50_ms=%.2f p95_ms=%.2f mean_ms=%.2f max_ms=%.2f\n", *scope, len(values), values[0]*1000, percentile(values, .50)*1000, percentile(values, .95)*1000, mean*1000, values[len(values)-1]*1000)
		printQueryMetrics(recorder, route)
		return
	}
	request("cold", 1)
	if *coldOnly {
		if *csvPath != "" {
			if err := writeCSV(*csvPath, rows); err != nil {
				log.Fatal("write CSV: ", err)
			}
		}
		elapsed, _ := strconv.ParseFloat(rows[0].elapsed, 64)
		fmt.Printf("scope=%s phase=cold status=%s elapsed_ms=%.2f\n", *scope, rows[0].status, elapsed*1000)
		printQueryMetrics(recorder, route)
		return
	}
	for iteration := 1; iteration <= *warmups; iteration++ {
		request("warmup", iteration)
	}
	if *concurrency == 1 {
		for iteration := 1; iteration <= *samples; iteration++ {
			request("sample", iteration)
		}
	} else {
		jobs := make(chan int)
		workers := min(*concurrency, *samples)
		var wait sync.WaitGroup
		wait.Add(workers)
		for worker := 0; worker < workers; worker++ {
			go func() {
				defer wait.Done()
				for iteration := range jobs {
					request("sample", iteration)
				}
			}()
		}
		for iteration := 1; iteration <= *samples; iteration++ {
			jobs <- iteration
		}
		close(jobs)
		wait.Wait()
	}
	if *csvPath != "" {
		if err := writeCSV(*csvPath, rows); err != nil {
			log.Fatal("write CSV: ", err)
		}
	}
	values := make([]float64, 0, *samples)
	for _, row := range rows {
		if row.phase == "sample" && row.status == "200" {
			value, _ := strconv.ParseFloat(row.elapsed, 64)
			values = append(values, value)
		}
	}
	if len(values) != *samples {
		log.Fatalf("expected %d successful samples, got %d", *samples, len(values))
	}
	sort.Float64s(values)
	mean := 0.0
	for _, value := range values {
		mean += value
	}
	mean /= float64(len(values))
	fmt.Printf("scope=%s city=%s samples=%d concurrency=%d min_ms=%.2f p50_ms=%.2f p95_ms=%.2f mean_ms=%.2f max_ms=%.2f\n",
		*scope, *cityID, len(values), *concurrency, values[0]*1000, percentile(values, .50)*1000, percentile(values, .95)*1000, mean*1000, values[len(values)-1]*1000)
	printQueryMetrics(recorder, route)
}

func printQueryMetrics(recorder *observability.HTTPRecorder, route string) {
	metrics := httptest.NewRecorder()
	recorder.ServeHTTP(metrics, httptest.NewRequest(http.MethodGet, "/metrics", nil))
	for _, line := range strings.Split(metrics.Body.String(), "\n") {
		if strings.Contains(line, `route="`+route+`"`) && (strings.HasPrefix(line, "noir_db_queries_total") || strings.HasPrefix(line, "noir_db_query_duration_seconds_sum")) {
			fmt.Println(line)
		}
	}
}

func percentile(sorted []float64, quantile float64) float64 {
	index := int(float64(len(sorted))*quantile+.999999) - 1
	if index < 0 {
		index = 0
	}
	if index >= len(sorted) {
		index = len(sorted) - 1
	}
	return sorted[index]
}

func writeCSV(path string, rows []measurementSample) error {
	if err := os.MkdirAll(filepath.Dir(path), 0o755); err != nil {
		return err
	}
	file, err := os.Create(path)
	if err != nil {
		return err
	}
	defer file.Close()
	writer := csv.NewWriter(file)
	if err := writer.Write([]string{"phase", "iteration", "status", "seconds", "error"}); err != nil {
		return err
	}
	for _, row := range rows {
		if err := writer.Write([]string{row.phase, strconv.Itoa(row.iteration), row.status, row.elapsed, row.err}); err != nil {
			return err
		}
	}
	writer.Flush()
	return writer.Error()
}

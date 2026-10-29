package main

import (
	"context"
	"encoding/csv"
	"encoding/json"
	"flag"
	"fmt"
	"log"
	"net/http"
	"net/http/httptest"
	"os"
	"path/filepath"
	"runtime"
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

type sample struct {
	phase, route, elapsed, status string
	iteration, results, payload   int
}

func main() {
	countryID := flag.String("country", "fr", "country ID to measure")
	query := flag.String("query", "a", "company search term")
	serverID := flag.String("server", "", "world server ID; defaults to NOIR_WORLD_SERVER_ID")
	samples := flag.Int("samples", 100, "measured warm requests for each route")
	warmups := flag.Int("warmups", 5, "warmup requests for each route")
	concurrency := flag.Int("concurrency", 8, "workers for the concurrent phase; zero disables it")
	csvPath := flag.String("csv", "", "optional path for raw sample CSV")
	flag.Parse()
	if *samples < 1 || *warmups < 0 || *concurrency < 0 || strings.TrimSpace(*query) == "" {
		log.Fatal("samples must be positive, warmups and concurrency nonnegative, and query nonempty")
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
	var cities []world.WorldCity
	for _, city := range world.Cities() {
		if city.CountryID == *countryID {
			cities = append(cities, city)
		}
	}
	if len(cities) == 0 {
		log.Fatalf("unknown country %q", *countryID)
	}
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Minute)
	defer cancel()
	poolConfig, err := pgxpool.ParseConfig(databaseURL)
	if err != nil {
		log.Fatal("parse database pool config: ", err)
	}
	poolConfig.ConnConfig.DefaultQueryExecMode = pgx.QueryExecModeExec
	recorder := observability.NewHTTPRecorder()
	poolConfig.ConnConfig.Tracer = observability.NewPGXQueryTracer(recorder)
	pool, err := pgxpool.NewWithConfig(ctx, poolConfig)
	if err != nil {
		log.Fatal("open database pool: ", err)
	}
	defer pool.Close()
	if err := pool.Ping(ctx); err != nil {
		log.Fatal("connect to database: ", err)
	}

	var target world.CompanyPortalRecord
	rows := make([]sample, 0, 1+(*samples+*warmups)*2)
	var rowsMu sync.Mutex
	request := func(activeStore *world.PostgresStore, route, phase string, iteration int) []byte {
		started := time.Now()
		response := httptest.NewRecorder()
		resultCount := 0
		handler := http.HandlerFunc(func(writer http.ResponseWriter, request *http.Request) {
			var payload any
			if route == "company_portal" {
				results, err := activeStore.SearchCompaniesInCountry(request.Context(), *serverID, *countryID, *query)
				if err != nil {
					http.Error(writer, err.Error(), http.StatusInternalServerError)
					return
				}
				resultCount = len(results)
				if len(results) > 0 && target.Company.ID == "" {
					target = results[0]
				}
				payload = map[string]any{"results": results, "countryId": *countryID, "limit": 10}
			} else {
				results, err := activeStore.CompanyDirectoryAtLocation(request.Context(), *serverID, target.Address.LocationID, target.Company.ID)
				if err != nil {
					http.Error(writer, err.Error(), http.StatusInternalServerError)
					return
				}
				resultCount = len(results)
				payload = map[string]any{"results": results}
			}
			writer.Header().Set("Content-Type", "application/json")
			_ = json.NewEncoder(writer).Encode(payload)
		})
		wrapped := recorder.Middleware(func(*http.Request) string { return route }, handler)
		wrapped.ServeHTTP(response, httptest.NewRequest(http.MethodGet, "/m23/"+route, nil))
		body := response.Body.Bytes()
		rowsMu.Lock()
		rows = append(rows, sample{phase: phase, route: route, iteration: iteration, elapsed: strconv.FormatFloat(time.Since(started).Seconds(), 'f', 9, 64), status: strconv.Itoa(response.Code), results: resultCount, payload: len(body)})
		rowsMu.Unlock()
		if response.Code != http.StatusOK {
			log.Fatalf("%s %s returned %d: %s", phase, route, response.Code, strings.TrimSpace(string(body)))
		}
		return body
	}

	var before, after runtime.MemStats
	runtime.GC()
	runtime.ReadMemStats(&before)
	coldStore := world.NewPostgresStoreWithCacheObserver(pool, recorder)
	request(coldStore, "company_portal", "cold", 1)
	runtime.GC()
	runtime.ReadMemStats(&after)
	fmt.Printf("country=%s cities=%d cold_index_heap_delta_bytes=%d\n", *countryID, len(cities), int64(after.HeapAlloc)-int64(before.HeapAlloc))
	if target.Company.ID == "" {
		log.Fatalf("query %q returned no companies; choose a broader query", *query)
	}
	for _, route := range []string{"company_portal", "company_directory"} {
		for iteration := 1; iteration <= *warmups; iteration++ {
			request(coldStore, route, "warmup", iteration)
		}
		for iteration := 1; iteration <= *samples; iteration++ {
			request(coldStore, route, "sample", iteration)
		}
	}
	if *concurrency > 0 {
		for _, route := range []string{"company_portal", "company_directory"} {
			jobs := make(chan int)
			var workers sync.WaitGroup
			workerCount := *concurrency
			if workerCount > *samples {
				workerCount = *samples
			}
			workers.Add(workerCount)
			for worker := 0; worker < workerCount; worker++ {
				go func() {
					defer workers.Done()
					for iteration := range jobs {
						request(coldStore, route, "concurrent", iteration)
					}
				}()
			}
			for iteration := 1; iteration <= *samples; iteration++ {
				jobs <- iteration
			}
			close(jobs)
			workers.Wait()
		}
	}
	if *csvPath != "" {
		if err := writeCSV(*csvPath, rows); err != nil {
			log.Fatal("write CSV: ", err)
		}
	}
	for _, route := range []string{"company_portal", "company_directory"} {
		for _, phase := range []string{"sample", "concurrent"} {
			values := make([]float64, 0, *samples)
			for _, row := range rows {
				if row.route == route && row.phase == phase && row.status == "200" {
					value, _ := strconv.ParseFloat(row.elapsed, 64)
					values = append(values, value)
				}
			}
			if len(values) == 0 {
				continue
			}
			sort.Float64s(values)
			fmt.Printf("route=%s phase=%s samples=%d concurrency=%d p50_ms=%.2f p95_ms=%.2f", route, phase, len(values), func() int {
				if phase == "concurrent" {
					return *concurrency
				}
				return 1
			}(), percentile(values, .50)*1000, percentile(values, .95)*1000)
			for _, row := range rows {
				if row.route == route && row.phase == phase {
					fmt.Printf(" results=%d payload_bytes=%d", row.results, row.payload)
					break
				}
			}
			fmt.Println()
		}
	}
	printQueryMetrics(recorder)
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

func printQueryMetrics(recorder *observability.HTTPRecorder) {
	metrics := httptest.NewRecorder()
	recorder.ServeHTTP(metrics, httptest.NewRequest(http.MethodGet, "/metrics", nil))
	for _, line := range strings.Split(metrics.Body.String(), "\n") {
		if strings.Contains(line, `route="company_portal"`) || strings.Contains(line, `route="company_directory"`) {
			if strings.HasPrefix(line, "noir_db_queries_total") || strings.HasPrefix(line, "noir_db_query_duration_seconds_sum") {
				fmt.Println(line)
			}
		}
		if strings.HasPrefix(line, "noir_cache_requests_total") {
			fmt.Println(line)
		}
	}
}

func writeCSV(path string, rows []sample) error {
	if err := os.MkdirAll(filepath.Dir(path), 0o755); err != nil {
		return err
	}
	file, err := os.Create(path)
	if err != nil {
		return err
	}
	defer file.Close()
	writer := csv.NewWriter(file)
	if err := writer.Write([]string{"phase", "route", "iteration", "status", "seconds", "results", "payload_bytes"}); err != nil {
		return err
	}
	for _, row := range rows {
		if err := writer.Write([]string{row.phase, row.route, strconv.Itoa(row.iteration), row.status, row.elapsed, strconv.Itoa(row.results), strconv.Itoa(row.payload)}); err != nil {
			return err
		}
	}
	writer.Flush()
	return writer.Error()
}

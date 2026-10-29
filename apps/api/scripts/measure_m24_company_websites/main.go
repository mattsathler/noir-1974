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
	phase, route, status string
	iteration, payload   int
	elapsed              float64
}

func main() {
	countryID := flag.String("country", "fr", "country ID to measure")
	query := flag.String("query", "a", "company query used to select a website")
	serverID := flag.String("server", "", "world server ID; defaults to NOIR_WORLD_SERVER_ID")
	samples := flag.Int("samples", 100, "requests per route and phase")
	coldSamples := flag.Int("cold-samples", 5, "cold-cache requests per route (a new store cache for each request)")
	warmups := flag.Int("warmups", 5, "sequential warmup requests per route")
	concurrency := flag.Int("concurrency", 8, "concurrent workers; zero disables concurrent samples")
	csvPath := flag.String("csv", "", "optional path for raw CSV samples")
	flag.Parse()
	if *samples < 1 || *coldSamples < 1 || *warmups < 0 || *concurrency < 0 || strings.TrimSpace(*query) == "" {
		log.Fatal("samples and cold-samples must be positive, warmups and concurrency nonnegative, and query nonempty")
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
	var cityCount int
	for _, city := range world.Cities() {
		if city.CountryID == *countryID {
			cityCount++
		}
	}
	if cityCount == 0 {
		log.Fatalf("unknown country %q", *countryID)
	}
	ctx, cancel := context.WithTimeout(context.Background(), 15*time.Minute)
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
	newStore := func() *world.PostgresStore { return world.NewPostgresStoreWithCacheObserver(pool, recorder) }
	store := newStore()
	catalog, err := store.SearchCompaniesInCountry(ctx, *serverID, *countryID, *query)
	if err != nil || len(catalog) == 0 {
		log.Fatalf("select benchmark company: results=%d error=%v", len(catalog), err)
	}
	target := catalog[0]
	routes := map[string]func(context.Context) (any, error){
		"company_search": func(ctx context.Context) (any, error) {
			return store.SearchCompaniesInCountry(ctx, *serverID, *countryID, *query)
		},
		"company_website": func(ctx context.Context) (any, error) {
			return store.CompanyWebsite(ctx, *serverID, *countryID, target.Company.ID)
		},
	}
	rows := make([]sample, 0, (*coldSamples+*warmups+2*(*samples))*len(routes))
	var rowsMu sync.Mutex
	request := func(route, phase string, iteration int) {
		operation := routes[route]
		started := time.Now()
		response := httptest.NewRecorder()
		handler := http.HandlerFunc(func(writer http.ResponseWriter, request *http.Request) {
			payload, err := operation(request.Context())
			if err != nil {
				http.Error(writer, err.Error(), http.StatusInternalServerError)
				return
			}
			writer.Header().Set("Content-Type", "application/json")
			_ = json.NewEncoder(writer).Encode(map[string]any{"results": payload})
		})
		wrapped := recorder.Middleware(func(*http.Request) string { return route }, handler)
		wrapped.ServeHTTP(response, httptest.NewRequest(http.MethodGet, "/api/world/companies/benchmark", nil))
		elapsed := time.Since(started).Seconds()
		row := sample{phase: phase, route: route, iteration: iteration, status: strconv.Itoa(response.Code), payload: response.Body.Len(), elapsed: elapsed}
		rowsMu.Lock()
		rows = append(rows, row)
		rowsMu.Unlock()
		if response.Code != http.StatusOK {
			log.Fatalf("%s %s returned %d: %s", phase, route, response.Code, strings.TrimSpace(response.Body.String()))
		}
	}
	var websiteMemoryDelta int64
	for route := range routes {
		for iteration := 1; iteration <= *coldSamples; iteration++ {
			store = newStore()
			if route == "company_website" && iteration == 1 {
				var before, after runtime.MemStats
				runtime.GC()
				runtime.ReadMemStats(&before)
				request(route, "cold", iteration)
				runtime.GC()
				runtime.ReadMemStats(&after)
				websiteMemoryDelta = int64(after.HeapAlloc) - int64(before.HeapAlloc)
			} else {
				request(route, "cold", iteration)
			}
		}
	}
	store = newStore()
	fmt.Printf("country=%s cities=%d website_cold_heap_delta_bytes=%d company=%s\n", *countryID, cityCount, websiteMemoryDelta, target.Company.ID)
	for route := range routes {
		for iteration := 1; iteration <= *warmups; iteration++ {
			request(route, "warmup", iteration)
		}
		for iteration := 1; iteration <= *samples; iteration++ {
			request(route, "sample", iteration)
		}
	}
	if *concurrency > 0 {
		for route := range routes {
			jobs := make(chan int)
			var workers sync.WaitGroup
			workerCount := min(*concurrency, *samples)
			workers.Add(workerCount)
			for worker := 0; worker < workerCount; worker++ {
				go func() {
					defer workers.Done()
					for iteration := range jobs {
						request(route, "concurrent", iteration)
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
	for route := range routes {
		for _, phase := range []string{"cold", "sample", "concurrent"} {
			values := make([]float64, 0, *samples)
			for _, row := range rows {
				if row.route == route && row.phase == phase && row.status == "200" {
					values = append(values, row.elapsed)
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
					fmt.Printf(" payload_bytes=%d", row.payload)
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
		if strings.Contains(line, `route="company_search"`) || strings.Contains(line, `route="company_website"`) {
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
	if err := writer.Write([]string{"phase", "route", "iteration", "status", "seconds", "payload_bytes"}); err != nil {
		return err
	}
	for _, row := range rows {
		if err := writer.Write([]string{row.phase, row.route, strconv.Itoa(row.iteration), row.status, strconv.FormatFloat(row.elapsed, 'f', 9, 64), strconv.Itoa(row.payload)}); err != nil {
			return err
		}
	}
	writer.Flush()
	return writer.Error()
}

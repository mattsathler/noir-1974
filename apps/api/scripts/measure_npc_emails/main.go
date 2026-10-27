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
	"noir-1974/apps/api/internal/cases"
	"noir-1974/apps/api/internal/mail"
	"noir-1974/apps/api/internal/observability"
	"noir-1974/apps/api/internal/world"
)

type sample struct {
	phase, status, seconds string
	iteration              int
}

type worldReader interface {
	WorldConfig(context.Context, string) (string, int, error)
	GetNPC(context.Context, string, string) (world.NPCProfile, error)
	EmployeesOfCompany(context.Context, string, string) ([]world.CompanyEmployee, error)
}

type legacyWorldReader struct{ store *world.PostgresStore }

func (reader legacyWorldReader) WorldConfig(ctx context.Context, serverID string) (string, int, error) {
	return reader.store.WorldConfig(ctx, serverID)
}

func (reader legacyWorldReader) GetNPC(ctx context.Context, serverID, npcID string) (world.NPCProfile, error) {
	return reader.store.GetNPC(ctx, serverID, npcID)
}

func (reader legacyWorldReader) EmployeesOfCompany(ctx context.Context, serverID, companyID string) ([]world.CompanyEmployee, error) {
	return reader.store.EmployeesOfCompany(ctx, serverID, companyID)
}

func (reader legacyWorldReader) EmployeesOfCompanies(ctx context.Context, serverID, seed string, version int, companyIDs []string) (map[string][]world.CompanyEmployee, error) {
	return reader.store.EmployeesOfCompanies(ctx, serverID, seed, version, companyIDs)
}

func main() {
	serverID := flag.String("server", "", "world server ID; defaults to NOIR_WORLD_SERVER_ID")
	npcID := flag.String("npc", "", "NPC ID; defaults to first employed Lyon NPC")
	fromValue := flag.String("from", "", "first mailbox date, YYYY-MM-DD; defaults to six days before --to")
	throughValue := flag.String("to", "", "last mailbox date, YYYY-MM-DD; defaults to today")
	samples := flag.Int("samples", 100, "measured requests after warmups")
	warmups := flag.Int("warmups", 5, "warmup requests")
	csvPath := flag.String("csv", "", "optional path for raw sample CSV")
	concurrency := flag.Int("concurrency", 1, "number of concurrent workers for measured samples")
	legacyConfigRead := flag.Bool("legacy-config-read", false, "emulate the previous service flow with a separate WorldConfig query")
	flag.Parse()
	for _, path := range []string{"../../.env.local", "../../.env"} {
		if err := godotenv.Load(path); err == nil {
			break
		}
	}
	if *serverID == "" {
		*serverID = os.Getenv("NOIR_WORLD_SERVER_ID")
	}
	if *serverID == "" {
		*serverID = "00000000-0000-4000-8000-000000000001"
	}
	if *samples < 1 || *warmups < 0 || *concurrency < 1 {
		log.Fatal("samples and concurrency must be positive; warmups cannot be negative")
	}
	var err error
	throughDate := time.Now().UTC().Truncate(24 * time.Hour)
	if *throughValue != "" {
		throughDate, err = time.Parse("2006-01-02", *throughValue)
		if err != nil {
			log.Fatal("parse --to: ", err)
		}
	}
	fromDate := throughDate.AddDate(0, 0, -6)
	if *fromValue != "" {
		fromDate, err = time.Parse("2006-01-02", *fromValue)
		if err != nil {
			log.Fatal("parse --from: ", err)
		}
	}
	from := fromDate
	if throughDate.Before(from) || throughDate.Sub(from) > 6*24*time.Hour {
		log.Fatal("--to must be within a valid seven-day interval")
	}
	through := throughDate.AddDate(0, 0, 1).Add(-time.Nanosecond)
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Minute)
	defer cancel()
	poolConfig, err := pgxpool.ParseConfig(os.Getenv("DATABASE_URL"))
	if err != nil {
		log.Fatal("parse DATABASE_URL: ", err)
	}
	if os.Getenv("DATABASE_URL") == "" {
		log.Fatal("DATABASE_URL is required")
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
	if *npcID == "" {
		if err := pool.QueryRow(ctx, `SELECT e.entity_id FROM world_entities e JOIN employment_relationships r ON r.server_id=e.server_id AND r.employee_npc_id=e.entity_id WHERE e.server_id=$1::uuid AND e.entity_type='npc' AND e.region_id LIKE 'fr/lyon/district/%' ORDER BY e.entity_id LIMIT 1`, *serverID).Scan(npcID); err != nil {
			log.Fatal("select QA NPC: ", err)
		}
	}
	worldStore := world.NewPostgresStore(pool)
	var reader worldReader = worldStore
	if *legacyConfigRead {
		reader = legacyWorldReader{store: worldStore}
	}
	service := mail.NewService(reader, cases.NewPostgresStore(pool))
	const route = "/m14/npc-emails-service"
	handler := recorder.Middleware(func(*http.Request) string { return route }, http.HandlerFunc(func(writer http.ResponseWriter, request *http.Request) {
		if _, err := service.NPCEmails(request.Context(), *serverID, *npcID, from, through); err != nil {
			http.Error(writer, err.Error(), http.StatusInternalServerError)
			return
		}
		writer.WriteHeader(http.StatusOK)
	}))
	rows := make([]sample, 0, *samples+*warmups+1)
	var rowsMu sync.Mutex
	request := func(phase string, iteration int) {
		started := time.Now()
		response := httptest.NewRecorder()
		handler.ServeHTTP(response, httptest.NewRequest(http.MethodGet, route, nil))
		rowsMu.Lock()
		rows = append(rows, sample{phase: phase, iteration: iteration, status: strconv.Itoa(response.Code), seconds: strconv.FormatFloat(time.Since(started).Seconds(), 'f', 9, 64)})
		rowsMu.Unlock()
	}
	request("cold", 1)
	for i := 1; i <= *warmups; i++ {
		request("warmup", i)
	}
	if *concurrency == 1 {
		for i := 1; i <= *samples; i++ {
			request("sample", i)
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
			value, _ := strconv.ParseFloat(row.seconds, 64)
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
	fmt.Printf("npc=%s samples=%d concurrency=%d min_ms=%.2f p50_ms=%.2f p95_ms=%.2f mean_ms=%.2f max_ms=%.2f\n", *npcID, len(values), *concurrency, values[0]*1000, percentile(values, .50)*1000, percentile(values, .95)*1000, mean*1000, values[len(values)-1]*1000)
	metrics := httptest.NewRecorder()
	recorder.ServeHTTP(metrics, httptest.NewRequest(http.MethodGet, "/metrics", nil))
	for _, line := range strings.Split(metrics.Body.String(), "\n") {
		if strings.Contains(line, `route="`+route+`"`) && (strings.HasPrefix(line, "noir_db_queries_total") || strings.HasPrefix(line, "noir_db_query_duration_seconds_sum")) {
			fmt.Println(line)
		}
	}
}

func percentile(values []float64, quantile float64) float64 {
	index := int(float64(len(values))*quantile+.999999) - 1
	if index < 0 {
		index = 0
	}
	if index >= len(values) {
		index = len(values) - 1
	}
	return values[index]
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
	if err := writer.Write([]string{"phase", "iteration", "status", "seconds"}); err != nil {
		return err
	}
	for _, row := range rows {
		if err := writer.Write([]string{row.phase, strconv.Itoa(row.iteration), row.status, row.seconds}); err != nil {
			return err
		}
	}
	writer.Flush()
	return writer.Error()
}

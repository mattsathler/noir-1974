package main

import (
	"context"
	"errors"
	"fmt"
	"log"
	"net/http"
	"os"
	"os/signal"
	"strconv"
	"syscall"
	"time"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/joho/godotenv"
	"noir-1974/apps/api/internal/cases"
	"noir-1974/apps/api/internal/health"
	"noir-1974/apps/api/internal/identity"
	"noir-1974/apps/api/internal/mail"
	"noir-1974/apps/api/internal/newspaper"
	"noir-1974/apps/api/internal/observability"
	"noir-1974/apps/api/internal/records"
	"noir-1974/apps/api/internal/world"
)

const defaultWorldServerID = "00000000-0000-4000-8000-000000000001"
const startupTimeout = 90 * time.Second

func main() {
	for _, envFile := range []string{".env.local", ".env", "../../.env.local", "../../.env"} {
		if err := godotenv.Load(envFile); err == nil {
			break
		}
	}

	databaseURL := os.Getenv("DATABASE_URL")
	if databaseURL == "" {
		log.Fatal("DATABASE_URL is required")
	}

	ctx, cancel := context.WithTimeout(context.Background(), startupTimeout)
	defer cancel()

	httpMetrics := observability.NewHTTPRecorder()
	poolConfig, err := pgxpool.ParseConfig(databaseURL)
	if err != nil {
		log.Fatalf("parse database pool config: %v", err)
	}
	// Neon pooled connections use PgBouncer transaction pooling. Use unnamed
	// extended-protocol statements so server-side prepared statement state is
	// not tied to a backend connection that PgBouncer may later reassign.
	poolConfig.ConnConfig.DefaultQueryExecMode = pgx.QueryExecModeExec
	poolConfig.ConnConfig.Tracer = observability.NewPGXQueryTracer(httpMetrics)
	pool, err := pgxpool.NewWithConfig(ctx, poolConfig)
	if err != nil {
		log.Fatalf("create database pool: %v", err)
	}
	defer pool.Close()

	if err := pool.Ping(ctx); err != nil {
		log.Fatalf("connect to database: %v", err)
	}

	identityStore := identity.NewPostgresStore(pool)
	if err := identityStore.Migrate(ctx); err != nil {
		log.Fatalf("migrate identity schema: %v", err)
	}
	worldStore := world.NewPostgresStoreWithCacheObserver(pool, httpMetrics)
	if err := worldStore.Migrate(ctx); err != nil {
		log.Fatalf("migrate world schema: %v", err)
	}
	worldServerID := os.Getenv("NOIR_WORLD_SERVER_ID")
	if worldServerID == "" {
		worldServerID = defaultWorldServerID
	}
	worldSeed := os.Getenv("NOIR_WORLD_SEED")
	if worldSeed == "" {
		worldSeed = "noir-1974-shared-city"
	}
	worldVersion := world.LatestGeneratorVersion
	storedSeed, storedVersion, configErr := worldStore.WorldConfig(ctx, worldServerID)
	if configErr == nil {
		if storedSeed != worldSeed {
			log.Fatalf("configured world seed conflicts with the existing server world")
		}
		worldVersion = storedVersion
	} else if !errors.Is(configErr, world.ErrEntityNotFound) {
		log.Fatalf("read shared world configuration: %v", configErr)
	}
	if err := worldStore.Materialize(ctx, worldServerID, worldSeed, worldVersion); err != nil {
		log.Fatalf("materialize shared world: %v", err)
	}
	caseStore := cases.NewPostgresStore(pool)
	if err := caseStore.Migrate(ctx); err != nil {
		log.Fatalf("migrate cases schema: %v", err)
	}
	caseInterval := 30 * time.Minute
	if configured := os.Getenv("NOIR_CASE_GENERATION_INTERVAL"); configured != "" {
		caseInterval, err = time.ParseDuration(configured)
		if err != nil || caseInterval < time.Minute {
			log.Fatalf("NOIR_CASE_GENERATION_INTERVAL must be a duration of at least 1m")
		}
	}
	maxOpenCases := 10
	if configured := os.Getenv("NOIR_MAX_OPEN_CASES"); configured != "" {
		maxOpenCases, err = strconv.Atoi(configured)
		if err != nil || maxOpenCases < 1 {
			log.Fatalf("NOIR_MAX_OPEN_CASES must be a positive integer")
		}
	}
	if err := caseStore.EnsureGenerationSchedule(ctx, worldServerID, caseInterval); err != nil {
		log.Fatalf("initialize recurring case schedule: %v", err)
	}
	generatedWorld, err := world.Generate(worldSeed, worldVersion)
	if err != nil {
		log.Fatalf("generate shared world: %v", err)
	}
	openingCase, err := cases.Generate(generatedWorld, worldSeed+"/opening-case", cases.GeneratorVersion)
	if err != nil {
		log.Fatalf("generate opening case: %v", err)
	}
	if err := caseStore.Create(ctx, worldServerID, generatedWorld, openingCase); err != nil {
		log.Fatalf("persist opening case: %v", err)
	}
	if openingCase.ArchetypeID != "financial_extortion" {
		if _, err := ensureExtortionCase(ctx, caseStore, worldServerID, worldSeed, generatedWorld); err != nil {
			log.Fatalf("persist opening extortion case: %v", err)
		}
	}
	mux := http.NewServeMux()
	newsService := newspaper.NewService(worldStore, caseStore)
	recordsService := records.NewService(worldStore, caseStore)
	mailService := mail.NewService(worldStore, caseStore)
	appEnvironment := os.Getenv("APP_ENV")
	exposeErrorDetails := appEnvironment == "" || appEnvironment == "development"
	mux.Handle("/api/", identity.NewHandlerWithGameAndMailMetrics(identityStore, appEnvironment == "production", worldStore, newsService, caseStore, recordsService, mailService, worldServerID, exposeErrorDetails, httpMetrics))
	mux.Handle("GET /api/health", httpMetrics.Middleware(func(*http.Request) string { return "/api/health" }, health.NewHandler(pool)))
	mux.Handle("GET /metrics", httpMetrics)
	server := &http.Server{
		Addr:              ":8080",
		Handler:           mux,
		ReadHeaderTimeout: 5 * time.Second,
	}

	serverCtx, stop := signal.NotifyContext(context.Background(), os.Interrupt, syscall.SIGTERM)
	defer stop()
	recurring := cases.NewRecurringGenerator(databaseURL, worldServerID, caseStore, worldStore, caseInterval, maxOpenCases)
	go recurring.Run(serverCtx)
	asyncRequests := cases.NewAsyncRequestProcessor(caseStore, worldServerID, 5*time.Second)
	asyncRequests.SetAccusationAcceptedHandler(func(cityID string) { newsService.InvalidateCity(worldServerID, cityID) })
	asyncRequests.SetRequestCompletedHandler(newsService.InvalidateAll)
	go asyncRequests.Run(serverCtx)

	go func() {
		<-serverCtx.Done()
		shutdownCtx, shutdownCancel := context.WithTimeout(context.Background(), 5*time.Second)
		defer shutdownCancel()
		if err := server.Shutdown(shutdownCtx); err != nil && !errors.Is(err, http.ErrServerClosed) {
			log.Printf("shutdown server: %v", err)
		}
	}()

	log.Printf("Noir API listening on %s", server.Addr)
	if err := server.ListenAndServe(); err != nil && !errors.Is(err, http.ErrServerClosed) {
		log.Fatalf("serve API: %v", err)
	}
}

func ensureExtortionCase(ctx context.Context, store *cases.PostgresStore, serverID, worldSeed string, base world.World) (cases.Definition, error) {
	for attempt := 0; attempt < 128; attempt++ {
		seed := fmt.Sprintf("%s/opening-extortion-%d", worldSeed, attempt)
		definition, err := cases.Generate(base, seed, cases.GeneratorVersion)
		if err != nil {
			return cases.Definition{}, err
		}
		if definition.ArchetypeID != "financial_extortion" {
			continue
		}
		if err := store.Create(ctx, serverID, base, definition); err != nil {
			return cases.Definition{}, err
		}
		return definition, nil
	}
	return cases.Definition{}, errors.New("could not derive an extortion case from the generated world")
}

package main

import (
	"context"
	"errors"
	"fmt"
	"log"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/joho/godotenv"
	"noir-1974/apps/api/internal/cases"
	"noir-1974/apps/api/internal/health"
	"noir-1974/apps/api/internal/identity"
	"noir-1974/apps/api/internal/mail"
	"noir-1974/apps/api/internal/newspaper"
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

	pool, err := pgxpool.New(ctx, databaseURL)
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
	worldStore := world.NewPostgresStore(pool)
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
	if err := worldStore.Materialize(ctx, worldServerID, worldSeed, world.LatestGeneratorVersion); err != nil {
		log.Fatalf("materialize shared world: %v", err)
	}
	caseStore := cases.NewPostgresStore(pool)
	if err := caseStore.Migrate(ctx); err != nil {
		log.Fatalf("migrate cases schema: %v", err)
	}
	generatedWorld, err := world.Generate(worldSeed, world.LatestGeneratorVersion)
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
	openingCases := []cases.Definition{openingCase}
	if openingCase.ArchetypeID != "financial_extortion" {
		extortionCase, err := ensureExtortionCase(ctx, caseStore, worldServerID, worldSeed, generatedWorld)
		if err != nil {
			log.Fatalf("persist opening extortion case: %v", err)
		}
		openingCases = append(openingCases, extortionCase)
	}
	for _, definition := range openingCases {
		victim, err := worldStore.GetNPC(ctx, worldServerID, definition.VictimID)
		if err != nil {
			log.Fatalf("load opening case victim: %v", err)
		}
		if victim.NPC.Status != "dead" {
			if err := worldStore.SetNPCStatus(ctx, worldServerID, definition.VictimID, "dead"); err != nil {
				log.Fatalf("record opening case victim status: %v", err)
			}
		}
	}

	mux := http.NewServeMux()
	newsService := newspaper.NewService(worldStore, caseStore)
	recordsService := records.NewService(worldStore, caseStore)
	mailService := mail.NewService(worldStore, caseStore)
	mux.Handle("/api/", identity.NewHandlerWithGameAndMail(identityStore, os.Getenv("APP_ENV") == "production", worldStore, newsService, caseStore, recordsService, mailService, worldServerID))
	mux.Handle("GET /api/health", health.NewHandler(pool))
	server := &http.Server{
		Addr:              ":8080",
		Handler:           mux,
		ReadHeaderTimeout: 5 * time.Second,
	}

	serverCtx, stop := signal.NotifyContext(context.Background(), os.Interrupt, syscall.SIGTERM)
	defer stop()

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

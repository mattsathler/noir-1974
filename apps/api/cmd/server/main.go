package main

import (
	"context"
	"errors"
	"log"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/joho/godotenv"
	"noir-1974/apps/api/internal/health"
	"noir-1974/apps/api/internal/identity"
)

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

	ctx, cancel := context.WithTimeout(context.Background(), 15*time.Second)
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

	mux := http.NewServeMux()
	mux.Handle("/api/", identity.NewHandler(identityStore, os.Getenv("APP_ENV") == "production"))
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

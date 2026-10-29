package main

import (
	"context"
	"log"
	"os"
	"time"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/joho/godotenv"
	"noir-1974/apps/api/internal/world"
)

const defaultWorldServerID = "00000000-0000-4000-8000-000000000001"

func main() {
	for _, path := range []string{".env.local", ".env", "../../.env.local", "../../.env"} {
		if err := godotenv.Load(path); err == nil {
			break
		}
	}
	databaseURL := os.Getenv("DATABASE_URL")
	if databaseURL == "" {
		log.Fatal("DATABASE_URL is required")
	}
	serverID := os.Getenv("NOIR_WORLD_SERVER_ID")
	if serverID == "" {
		serverID = defaultWorldServerID
	}
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Minute)
	defer cancel()
	config, err := pgxpool.ParseConfig(databaseURL)
	if err != nil {
		log.Fatalf("parse database pool config: %v", err)
	}
	config.ConnConfig.DefaultQueryExecMode = pgx.QueryExecModeExec
	pool, err := pgxpool.NewWithConfig(ctx, config)
	if err != nil {
		log.Fatalf("create database pool: %v", err)
	}
	defer pool.Close()
	if err := pool.Ping(ctx); err != nil {
		log.Fatalf("connect to database: %v", err)
	}
	store := world.NewPostgresStore(pool)
	if err := store.UpgradeMaterializedWorld(ctx, serverID); err != nil {
		log.Fatalf("upgrade shared world content: %v", err)
	}
	var version, cities, districtsWithSixBuildings, companies, employees int
	if err := pool.QueryRow(ctx, `
		SELECT w.generator_version,
		       (SELECT count(*) FROM world_city_generators c WHERE c.server_id=w.server_id),
	       (SELECT count(*) FROM (
	          SELECT region_id FROM world_entities
	          WHERE server_id=w.server_id AND entity_type='location' AND location_kind='building'
	          GROUP BY region_id HAVING count(*) >= 6
	        ) districts),
	       (SELECT count(*) FROM world_entities WHERE server_id=w.server_id AND entity_type='company' AND generation_key LIKE '%/building/%/company/%'),
	       (SELECT count(*) FROM world_entities WHERE server_id=w.server_id AND entity_type='npc' AND generation_key LIKE '%/company/%/person/%')
		FROM worlds w WHERE w.server_id=$1::uuid`, serverID).Scan(&version, &cities, &districtsWithSixBuildings, &companies, &employees); err != nil {
		log.Fatalf("verify upgraded world content: %v", err)
	}
	log.Printf("shared world %s: version %d, %d materialized cities, %d districts with at least six buildings, %d expanded companies, %d expanded employees", serverID, version, cities, districtsWithSixBuildings, companies, employees)
}

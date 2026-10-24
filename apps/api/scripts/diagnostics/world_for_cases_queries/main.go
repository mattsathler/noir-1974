package main

import (
	"context"
	"flag"
	"fmt"
	"log"
	"os"
	"strings"
	"time"

	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/joho/godotenv"
)

func main() {
	serverID := flag.String("server", "", "world server ID; defaults to NOIR_WORLD_SERVER_ID")
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
	if os.Getenv("DATABASE_URL") == "" {
		log.Fatal("DATABASE_URL is required")
	}
	ctx, cancel := context.WithTimeout(context.Background(), 2*time.Minute)
	defer cancel()
	pool, err := pgxpool.New(ctx, os.Getenv("DATABASE_URL"))
	if err != nil {
		log.Fatal("open database pool: ", err)
	}
	defer pool.Close()
	queries := []struct {
		name string
		sql  string
	}{
		{"npc_entities", `SELECT entity_id, generation_key FROM world_entities WHERE server_id=$1::uuid AND entity_type='npc' ORDER BY entity_id`},
		{"employment_relationships", `SELECT relationship_id, employee_npc_id, employer_company_id, role FROM employment_relationships WHERE server_id=$1::uuid ORDER BY relationship_id`},
		{"personal_relationships", `SELECT relationship_id, person_a_npc_id, person_b_npc_id, relationship_type FROM personal_relationships WHERE server_id=$1::uuid ORDER BY relationship_id`},
		{"latest_npc_statuses", `SELECT DISTINCT ON (m.entity_id) m.entity_id, m.value #>> '{}' FROM world_mutations m JOIN world_entities e ON e.server_id=m.server_id AND e.entity_id=m.entity_id WHERE m.server_id=$1::uuid AND m.field_key='status' AND e.entity_type='npc' ORDER BY m.entity_id, m.mutation_order DESC`},
	}
	for _, query := range queries {
		rows, err := pool.Query(ctx, "EXPLAIN (ANALYZE, BUFFERS, FORMAT TEXT) "+query.sql, *serverID)
		if err != nil {
			log.Fatalf("explain %s: %v", query.name, err)
		}
		var plan []string
		for rows.Next() {
			var line string
			if err := rows.Scan(&line); err != nil {
				rows.Close()
				log.Fatal("scan plan: ", err)
			}
			plan = append(plan, line)
		}
		if err := rows.Err(); err != nil {
			rows.Close()
			log.Fatalf("read %s plan: %v", query.name, err)
		}
		rows.Close()
		fmt.Printf("--- %s ---\n%s\n", query.name, strings.Join(plan, "\n"))
	}
}

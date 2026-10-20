package world

import (
	"context"
	"embed"
	"encoding/json"
	"errors"
	"fmt"
	"strings"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgconn"
)

//go:embed schema.sql
var worldSchema embed.FS

var (
	ErrWorldConflict     = errors.New("server already has a different world seed or generator version")
	ErrEntityNotFound    = errors.New("world entity not found")
	ErrGeneratedConflict = errors.New("persisted generated entity conflicts with deterministic output")
	ErrInvalidMutation   = errors.New("invalid world mutation")
)

type worldDatabase interface {
	Exec(context.Context, string, ...any) (pgconn.CommandTag, error)
	Query(context.Context, string, ...any) (pgx.Rows, error)
	QueryRow(context.Context, string, ...any) pgx.Row
	Begin(context.Context) (pgx.Tx, error)
}

type PostgresStore struct {
	db worldDatabase
}

type EmploymentView struct {
	RelationshipID string  `json:"relationshipId"`
	Role           string  `json:"role"`
	Company        Company `json:"company"`
}

type PersonalRelationshipView struct {
	RelationshipType string `json:"relationshipType"`
	NPC              NPC    `json:"npc"`
}

type NPCProfile struct {
	NPC                   NPC                        `json:"npc"`
	Employment            []EmploymentView           `json:"employment"`
	PersonalRelationships []PersonalRelationshipView `json:"personalRelationships"`
}

func NewPostgresStore(db worldDatabase) *PostgresStore {
	return &PostgresStore{db: db}
}

func (store *PostgresStore) Migrate(ctx context.Context) error {
	contents, err := worldSchema.ReadFile("schema.sql")
	if err != nil {
		return err
	}
	for _, statement := range strings.Split(string(contents), ";") {
		if strings.TrimSpace(statement) == "" {
			continue
		}
		if _, err := store.db.Exec(ctx, statement); err != nil {
			return fmt.Errorf("apply world schema: %w", err)
		}
	}
	return nil
}

// Materialize generates a world once and persists identity anchors and queryable relationships.
func (store *PostgresStore) Materialize(ctx context.Context, serverID, seed string, version int) error {
	generated, err := Generate(seed, version)
	if err != nil {
		return err
	}
	tx, err := store.db.Begin(ctx)
	if err != nil {
		return fmt.Errorf("begin world materialization: %w", err)
	}
	defer func() { _ = tx.Rollback(context.Background()) }()

	if _, err := tx.Exec(ctx,
		`INSERT INTO game_servers (server_id) VALUES ($1::uuid) ON CONFLICT (server_id) DO NOTHING`,
		serverID,
	); err != nil {
		return fmt.Errorf("ensure game server: %w", err)
	}
	if _, err := tx.Exec(ctx,
		`INSERT INTO worlds (server_id, world_seed, generator_version)
		 VALUES ($1::uuid, $2, $3) ON CONFLICT (server_id) DO NOTHING`,
		serverID, seed, version,
	); err != nil {
		return fmt.Errorf("ensure world record: %w", err)
	}
	var storedSeed string
	var storedVersion int
	if err := tx.QueryRow(ctx,
		`SELECT world_seed, generator_version FROM worlds WHERE server_id = $1::uuid FOR UPDATE`,
		serverID,
	).Scan(&storedSeed, &storedVersion); err != nil {
		return fmt.Errorf("read world record: %w", err)
	}
	if storedSeed != seed || storedVersion != version {
		return ErrWorldConflict
	}

	for _, region := range generated.Regions {
		if err := execConsistent(ctx, tx,
			`INSERT INTO world_regions (server_id, region_id, parent_region_id, region_kind, display_name)
			 VALUES ($1::uuid, $2, $3, $4, $5)
			 ON CONFLICT (server_id, region_id) DO UPDATE
			 SET display_name = EXCLUDED.display_name
			 WHERE world_regions.parent_region_id IS NOT DISTINCT FROM EXCLUDED.parent_region_id
			   AND world_regions.region_kind = EXCLUDED.region_kind
			   AND world_regions.display_name = EXCLUDED.display_name`,
			serverID, region.ID, optionalString(region.ParentID), region.Kind, region.Name,
		); err != nil {
			return fmt.Errorf("persist region %q: %w", region.ID, err)
		}
	}
	for _, entity := range generated.Entities {
		if err := execConsistent(ctx, tx,
			`INSERT INTO world_entities (server_id, entity_id, entity_type, region_id, generation_key, location_kind)
			 VALUES ($1::uuid, $2, $3, $4, $5, $6)
			 ON CONFLICT (server_id, entity_id) DO UPDATE
			 SET generation_key = EXCLUDED.generation_key
			 WHERE world_entities.entity_type = EXCLUDED.entity_type
			   AND world_entities.region_id = EXCLUDED.region_id
			   AND world_entities.generation_key = EXCLUDED.generation_key
			   AND world_entities.location_kind IS NOT DISTINCT FROM EXCLUDED.location_kind`,
			serverID, entity.ID, entity.EntityType, entity.RegionID, entity.GenerationKey, optionalString(entity.LocationKind),
		); err != nil {
			return fmt.Errorf("persist entity %q: %w", entity.ID, err)
		}
	}
	for _, relationship := range generated.EmploymentRelationships {
		if err := execConsistent(ctx, tx,
			`INSERT INTO employment_relationships (server_id, relationship_id, employee_npc_id, employer_company_id, role)
			 VALUES ($1::uuid, $2, $3, $4, $5)
			 ON CONFLICT (server_id, relationship_id) DO UPDATE SET role = EXCLUDED.role
			 WHERE employment_relationships.employee_npc_id = EXCLUDED.employee_npc_id
			   AND employment_relationships.employer_company_id = EXCLUDED.employer_company_id
			   AND employment_relationships.role IS NOT DISTINCT FROM EXCLUDED.role`,
			serverID, relationship.ID, relationship.EmployeeNPCID, relationship.EmployerCompanyID, relationship.Role,
		); err != nil {
			return fmt.Errorf("persist employment relationship %q: %w", relationship.ID, err)
		}
	}
	for _, relationship := range generated.PersonalRelationships {
		if err := execConsistent(ctx, tx,
			`INSERT INTO personal_relationships (server_id, relationship_id, person_a_npc_id, person_b_npc_id, relationship_type)
			 VALUES ($1::uuid, $2, $3, $4, $5)
			 ON CONFLICT (server_id, relationship_id) DO UPDATE SET relationship_type = EXCLUDED.relationship_type
			 WHERE personal_relationships.person_a_npc_id = EXCLUDED.person_a_npc_id
			   AND personal_relationships.person_b_npc_id = EXCLUDED.person_b_npc_id
			   AND personal_relationships.relationship_type = EXCLUDED.relationship_type`,
			serverID, relationship.ID, relationship.PersonANPCID, relationship.PersonBNPCID, relationship.RelationshipType,
		); err != nil {
			return fmt.Errorf("persist personal relationship %q: %w", relationship.ID, err)
		}
	}
	for _, relationship := range generated.CompanyLocations {
		if err := execConsistent(ctx, tx,
			`INSERT INTO company_locations (server_id, relationship_id, company_id, location_id, location_role)
			 VALUES ($1::uuid, $2, $3, $4, $5)
			 ON CONFLICT (server_id, relationship_id) DO UPDATE SET location_role = EXCLUDED.location_role
			 WHERE company_locations.company_id = EXCLUDED.company_id
			   AND company_locations.location_id = EXCLUDED.location_id
			   AND company_locations.location_role = EXCLUDED.location_role`,
			serverID, relationship.ID, relationship.CompanyID, relationship.LocationID, relationship.LocationRole,
		); err != nil {
			return fmt.Errorf("persist company location %q: %w", relationship.ID, err)
		}
	}
	if err := tx.Commit(ctx); err != nil {
		return fmt.Errorf("commit world materialization: %w", err)
	}
	return nil
}

// GetNPC rebuilds one person's generated fields and loads persisted relationships by indexed IDs.
func (store *PostgresStore) GetNPC(ctx context.Context, serverID, npcID string) (NPCProfile, error) {
	var seed string
	var version int
	var generationKey string
	err := store.db.QueryRow(ctx,
		`SELECT w.world_seed, w.generator_version, e.generation_key
		 FROM world_entities e JOIN worlds w ON w.server_id = e.server_id
		 WHERE e.server_id = $1::uuid AND e.entity_id = $2 AND e.entity_type = 'npc'`,
		serverID, npcID,
	).Scan(&seed, &version, &generationKey)
	if errors.Is(err, pgx.ErrNoRows) {
		return NPCProfile{}, ErrEntityNotFound
	}
	if err != nil {
		return NPCProfile{}, err
	}
	npc, err := GenerateNPC(seed, version, generationKey)
	if err != nil {
		return NPCProfile{}, err
	}
	if npc.ID != npcID {
		return NPCProfile{}, ErrGeneratedConflict
	}
	if err := store.applyNPCStatus(ctx, serverID, npcID, &npc); err != nil {
		return NPCProfile{}, err
	}
	profile := NPCProfile{
		NPC:                   npc,
		Employment:            make([]EmploymentView, 0),
		PersonalRelationships: make([]PersonalRelationshipView, 0),
	}
	if err := store.loadEmployment(ctx, serverID, seed, version, npcID, &profile); err != nil {
		return NPCProfile{}, err
	}
	if err := store.loadPersonalRelationships(ctx, serverID, seed, version, npcID, &profile); err != nil {
		return NPCProfile{}, err
	}
	return profile, nil
}

// SetNPCStatus appends a persistent status change scoped to one server.
func (store *PostgresStore) SetNPCStatus(ctx context.Context, serverID, npcID, status string) error {
	if status != "alive" && status != "dead" {
		return ErrInvalidMutation
	}
	value, err := json.Marshal(status)
	if err != nil {
		return err
	}
	result, err := store.db.Exec(ctx,
		`INSERT INTO world_mutations (server_id, entity_id, field_key, value)
		 SELECT $1::uuid, e.entity_id, 'status', $3::jsonb
		 FROM world_entities e
		 WHERE e.server_id = $1::uuid AND e.entity_id = $2 AND e.entity_type = 'npc'`,
		serverID, npcID, string(value),
	)
	if err != nil {
		return err
	}
	if result.RowsAffected() != 1 {
		return ErrEntityNotFound
	}
	return nil
}

func (store *PostgresStore) applyNPCStatus(ctx context.Context, serverID, npcID string, npc *NPC) error {
	var status string
	err := store.db.QueryRow(ctx,
		`SELECT value #>> '{}' FROM world_mutations
		 WHERE server_id = $1::uuid AND entity_id = $2 AND field_key = 'status'
	 ORDER BY mutation_order DESC LIMIT 1`,
		serverID, npcID,
	).Scan(&status)
	if errors.Is(err, pgx.ErrNoRows) {
		return nil
	}
	if err != nil {
		return err
	}
	if status != "alive" && status != "dead" {
		return ErrInvalidMutation
	}
	npc.Status = status
	return nil
}

// EmployeesOfCompany uses the employer relationship index and reconstructs only matching NPCs.
func (store *PostgresStore) EmployeesOfCompany(ctx context.Context, serverID, companyID string) ([]NPC, error) {
	seed, version, err := store.worldSeedAndVersion(ctx, serverID)
	if err != nil {
		return nil, err
	}
	var companyKey string
	err = store.db.QueryRow(ctx,
		`SELECT generation_key FROM world_entities
		 WHERE server_id = $1::uuid AND entity_id = $2 AND entity_type = 'company'`,
		serverID, companyID,
	).Scan(&companyKey)
	if errors.Is(err, pgx.ErrNoRows) {
		return nil, ErrEntityNotFound
	}
	if err != nil {
		return nil, err
	}
	company, err := GenerateCompany(seed, version, companyKey)
	if err != nil {
		return nil, err
	}
	if company.ID != companyID {
		return nil, ErrGeneratedConflict
	}

	rows, err := store.db.Query(ctx,
		`SELECT e.generation_key
		 FROM employment_relationships r
		 JOIN world_entities e
		   ON e.server_id = r.server_id AND e.entity_id = r.employee_npc_id
		 WHERE r.server_id = $1::uuid AND r.employer_company_id = $2
		 ORDER BY r.employee_npc_id`,
		serverID, companyID,
	)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	employees := make([]NPC, 0)
	for rows.Next() {
		var key string
		if err := rows.Scan(&key); err != nil {
			return nil, err
		}
		npc, err := GenerateNPC(seed, version, key)
		if err != nil {
			return nil, err
		}
		employees = append(employees, npc)
	}
	if err := rows.Err(); err != nil {
		return nil, err
	}
	return employees, nil
}

func (store *PostgresStore) loadEmployment(ctx context.Context, serverID, seed string, version int, npcID string, profile *NPCProfile) error {
	rows, err := store.db.Query(ctx,
		`SELECT r.relationship_id, r.role, c.entity_id, c.generation_key
		 FROM employment_relationships r
		 JOIN world_entities c
		   ON c.server_id = r.server_id AND c.entity_id = r.employer_company_id
		 WHERE r.server_id = $1::uuid AND r.employee_npc_id = $2
		 ORDER BY r.relationship_id`,
		serverID, npcID,
	)
	if err != nil {
		return err
	}
	defer rows.Close()
	for rows.Next() {
		var view EmploymentView
		var companyID, companyKey string
		if err := rows.Scan(&view.RelationshipID, &view.Role, &companyID, &companyKey); err != nil {
			return err
		}
		company, err := GenerateCompany(seed, version, companyKey)
		if err != nil {
			return err
		}
		if company.ID != companyID {
			return ErrGeneratedConflict
		}
		view.Company = company
		profile.Employment = append(profile.Employment, view)
	}
	return rows.Err()
}

func (store *PostgresStore) loadPersonalRelationships(ctx context.Context, serverID, seed string, version int, npcID string, profile *NPCProfile) error {
	rows, err := store.db.Query(ctx,
		`SELECT r.relationship_type, other.entity_id, other.generation_key
		 FROM personal_relationships r
		 JOIN world_entities other
		   ON other.server_id = r.server_id
		  AND other.entity_id = CASE WHEN r.person_a_npc_id = $2 THEN r.person_b_npc_id ELSE r.person_a_npc_id END
		 WHERE r.server_id = $1::uuid AND (r.person_a_npc_id = $2 OR r.person_b_npc_id = $2)
		 ORDER BY r.relationship_type, other.entity_id`,
		serverID, npcID,
	)
	if err != nil {
		return err
	}
	defer rows.Close()
	for rows.Next() {
		var view PersonalRelationshipView
		var otherID, otherKey string
		if err := rows.Scan(&view.RelationshipType, &otherID, &otherKey); err != nil {
			return err
		}
		otherNPC, err := GenerateNPC(seed, version, otherKey)
		if err != nil {
			return err
		}
		if otherNPC.ID != otherID {
			return ErrGeneratedConflict
		}
		view.NPC = otherNPC
		profile.PersonalRelationships = append(profile.PersonalRelationships, view)
	}
	return rows.Err()
}

func (store *PostgresStore) worldSeedAndVersion(ctx context.Context, serverID string) (string, int, error) {
	var seed string
	var version int
	err := store.db.QueryRow(ctx,
		`SELECT world_seed, generator_version FROM worlds WHERE server_id = $1::uuid`,
		serverID,
	).Scan(&seed, &version)
	if errors.Is(err, pgx.ErrNoRows) {
		return "", 0, ErrEntityNotFound
	}
	return seed, version, err
}

func execConsistent(ctx context.Context, tx pgx.Tx, query string, args ...any) error {
	result, err := tx.Exec(ctx, query, args...)
	if err != nil {
		return err
	}
	if result.RowsAffected() != 1 {
		return ErrGeneratedConflict
	}
	return nil
}

func optionalString(value string) any {
	if value == "" {
		return nil
	}
	return value
}

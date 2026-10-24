package world

import (
	"context"
	"embed"
	"encoding/json"
	"errors"
	"fmt"
	"sort"
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

const MaxSearchResults = 5
const MaxNPCSearchResults = 10

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

// NPCRecordsProfile contains only the generated facts needed for a records query.
type NPCRecordsProfile struct {
	NPC              NPC
	WorldSeed        string
	GeneratorVersion int
	Employer         string
}

type PersonSearchRecord struct {
	NPC           PersonSearchNPC         `json:"npc"`
	Workplaces    []PersonSearchWorkplace `json:"workplaces"`
	MaritalStatus string                  `json:"maritalStatus"`
	ReportedScene *PersonSearchScene      `json:"reportedScene,omitempty"`
}

type PersonSearchScene struct {
	Name    string `json:"name"`
	Address string `json:"address"`
}

type PersonSearchNPC struct {
	ID         string          `json:"id"`
	Name       string          `json:"name"`
	Age        int             `json:"age"`
	Appearance string          `json:"appearance"`
	Status     string          `json:"status"`
	Visual     *VisualIdentity `json:"visual,omitempty"`
}

type PersonSearchWorkplace struct {
	Company string `json:"company"`
	Address string `json:"address"`
}

type Address struct {
	LocationID string `json:"locationId"`
	Name       string `json:"name"`
	Kind       string `json:"kind"`
	Status     string `json:"status,omitempty"`
	District   string `json:"district"`
	City       string `json:"city"`
	Country    string `json:"country"`
	Full       string `json:"full"`
}

type CompanyRecord struct {
	Company   Company           `json:"company"`
	Address   Address           `json:"address"`
	Employees []CompanyEmployee `json:"employees"`
}

type CompanyEmployee struct {
	ID     string          `json:"id"`
	Name   string          `json:"name"`
	Role   string          `json:"role"`
	Status string          `json:"status"`
	Visual *VisualIdentity `json:"visual,omitempty"`
}

type PhoneContact struct {
	ID      string          `json:"id"`
	Name    string          `json:"name"`
	Company string          `json:"company"`
	Status  string          `json:"status"`
	Visual  *VisualIdentity `json:"visual,omitempty"`
}

type LocationRecord struct {
	Address Address `json:"address"`
}

type DistrictDestination struct {
	ID        string           `json:"id"`
	Name      string           `json:"name"`
	Locations []LocationRecord `json:"locations"`
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

	regionArgs := make([][]any, 0, len(generated.Regions))
	for _, region := range generated.Regions {
		regionArgs = append(regionArgs, []any{serverID, region.ID, optionalString(region.ParentID), region.Kind, region.Name})
	}
	regionIndex, err := execConsistentBatch(ctx, tx,
		`INSERT INTO world_regions (server_id, region_id, parent_region_id, region_kind, display_name)
		 VALUES ($1::uuid, $2, $3, $4, $5)
		 ON CONFLICT (server_id, region_id) DO UPDATE SET display_name = EXCLUDED.display_name
		 WHERE world_regions.parent_region_id IS NOT DISTINCT FROM EXCLUDED.parent_region_id
		   AND world_regions.region_kind = EXCLUDED.region_kind
		   AND world_regions.display_name = EXCLUDED.display_name`, regionArgs)
	if err != nil {
		return fmt.Errorf("persist region %q: %w", generated.Regions[max(regionIndex, 0)].ID, err)
	}

	entityArgs := make([][]any, 0, len(generated.Entities))
	for _, entity := range generated.Entities {
		entityArgs = append(entityArgs, []any{serverID, entity.ID, entity.EntityType, entity.RegionID, entity.GenerationKey, optionalString(entity.LocationKind)})
	}
	entityIndex, err := execConsistentBatch(ctx, tx,
		`INSERT INTO world_entities (server_id, entity_id, entity_type, region_id, generation_key, location_kind)
		 VALUES ($1::uuid, $2, $3, $4, $5, $6)
		 ON CONFLICT (server_id, entity_id) DO UPDATE SET generation_key = EXCLUDED.generation_key
		 WHERE world_entities.entity_type = EXCLUDED.entity_type
		   AND world_entities.region_id = EXCLUDED.region_id
		   AND world_entities.generation_key = EXCLUDED.generation_key
		   AND world_entities.location_kind IS NOT DISTINCT FROM EXCLUDED.location_kind`, entityArgs)
	if err != nil {
		return fmt.Errorf("persist entity %q: %w", generated.Entities[max(entityIndex, 0)].ID, err)
	}

	employmentArgs := make([][]any, 0, len(generated.EmploymentRelationships))
	for _, relationship := range generated.EmploymentRelationships {
		employmentArgs = append(employmentArgs, []any{serverID, relationship.ID, relationship.EmployeeNPCID, relationship.EmployerCompanyID, relationship.Role})
	}
	employmentIndex, err := execConsistentBatch(ctx, tx,
		`INSERT INTO employment_relationships (server_id, relationship_id, employee_npc_id, employer_company_id, role)
		 VALUES ($1::uuid, $2, $3, $4, $5)
		 ON CONFLICT (server_id, relationship_id) DO UPDATE SET role = EXCLUDED.role
		 WHERE employment_relationships.employee_npc_id = EXCLUDED.employee_npc_id
		   AND employment_relationships.employer_company_id = EXCLUDED.employer_company_id
		   AND employment_relationships.role IS NOT DISTINCT FROM EXCLUDED.role`, employmentArgs)
	if err != nil {
		return fmt.Errorf("persist employment relationship %q: %w", generated.EmploymentRelationships[max(employmentIndex, 0)].ID, err)
	}

	personalArgs := make([][]any, 0, len(generated.PersonalRelationships))
	for _, relationship := range generated.PersonalRelationships {
		personalArgs = append(personalArgs, []any{serverID, relationship.ID, relationship.PersonANPCID, relationship.PersonBNPCID, relationship.RelationshipType})
	}
	personalIndex, err := execConsistentBatch(ctx, tx,
		`INSERT INTO personal_relationships (server_id, relationship_id, person_a_npc_id, person_b_npc_id, relationship_type)
		 VALUES ($1::uuid, $2, $3, $4, $5)
		 ON CONFLICT (server_id, relationship_id) DO UPDATE SET relationship_type = EXCLUDED.relationship_type
		 WHERE personal_relationships.person_a_npc_id = EXCLUDED.person_a_npc_id
		   AND personal_relationships.person_b_npc_id = EXCLUDED.person_b_npc_id
		   AND personal_relationships.relationship_type = EXCLUDED.relationship_type`, personalArgs)
	if err != nil {
		return fmt.Errorf("persist personal relationship %q: %w", generated.PersonalRelationships[max(personalIndex, 0)].ID, err)
	}

	companyLocationArgs := make([][]any, 0, len(generated.CompanyLocations))
	for _, relationship := range generated.CompanyLocations {
		companyLocationArgs = append(companyLocationArgs, []any{serverID, relationship.ID, relationship.CompanyID, relationship.LocationID, relationship.LocationRole})
	}
	companyLocationIndex, err := execConsistentBatch(ctx, tx,
		`INSERT INTO company_locations (server_id, relationship_id, company_id, location_id, location_role)
		 VALUES ($1::uuid, $2, $3, $4, $5)
		 ON CONFLICT (server_id, relationship_id) DO UPDATE SET location_role = EXCLUDED.location_role
		 WHERE company_locations.company_id = EXCLUDED.company_id
		   AND company_locations.location_id = EXCLUDED.location_id
		   AND company_locations.location_role = EXCLUDED.location_role`, companyLocationArgs)
	if err != nil {
		return fmt.Errorf("persist company location %q: %w", generated.CompanyLocations[max(companyLocationIndex, 0)].ID, err)
	}
	if err := tx.Commit(ctx); err != nil {
		return fmt.Errorf("commit world materialization: %w", err)
	}
	return nil
}

// MaterializeCity adds a city's deterministic shared content the first time a
// player reaches it. New cities use at least the current city-generation rules;
// an existing city's stored version is never replaced.
func (store *PostgresStore) MaterializeCity(ctx context.Context, serverID, cityID string) error {
	city, ok := CityByID(cityID)
	if !ok {
		return ErrEntityNotFound
	}
	if cityID == "fr/lyon" {
		return nil
	}
	var seed string
	var worldVersion int
	var alreadyMaterialized bool
	err := store.db.QueryRow(ctx,
		`SELECT world_seed, generator_version,
			EXISTS (SELECT 1 FROM world_city_generators c WHERE c.server_id = w.server_id AND c.city_id = $2)
		 FROM worlds w WHERE w.server_id = $1::uuid`,
		serverID, city.ID,
	).Scan(&seed, &worldVersion, &alreadyMaterialized)
	if errors.Is(err, pgx.ErrNoRows) {
		return ErrEntityNotFound
	}
	if err != nil {
		return err
	}
	if alreadyMaterialized {
		return nil
	}
	version := worldVersion
	if version < GeneratorVersion5 {
		version = GeneratorVersion5
	}
	tx, err := store.db.Begin(ctx)
	if err != nil {
		return fmt.Errorf("begin city materialization: %w", err)
	}
	defer func() { _ = tx.Rollback(context.Background()) }()
	inserted, err := tx.Exec(ctx,
		`INSERT INTO world_city_generators(server_id, city_id, generator_version) VALUES($1::uuid, $2, $3) ON CONFLICT(server_id, city_id) DO NOTHING`,
		serverID, city.ID, version,
	)
	if err != nil {
		return fmt.Errorf("record city generator: %w", err)
	}
	if inserted.RowsAffected() == 0 {
		return nil
	}
	generated, err := GenerateCity(seed, version, city.ID)
	if err != nil {
		return err
	}
	if err := persistGeneratedCity(ctx, tx, serverID, generated); err != nil {
		return err
	}
	if err := tx.Commit(ctx); err != nil {
		return fmt.Errorf("commit city materialization: %w", err)
	}
	return nil
}

func (store *PostgresStore) cityGeneratorVersion(ctx context.Context, serverID, cityID string, fallback int) (int, error) {
	if cityID == "fr/lyon" {
		return fallback, nil
	}
	var version int
	err := store.db.QueryRow(ctx, `SELECT generator_version FROM world_city_generators WHERE server_id=$1::uuid AND city_id=$2`, serverID, cityID).Scan(&version)
	if errors.Is(err, pgx.ErrNoRows) {
		if fallback < GeneratorVersion5 {
			return GeneratorVersion5, nil
		}
		return fallback, nil
	}
	return version, err
}

func persistGeneratedCity(ctx context.Context, tx pgx.Tx, serverID string, generated World) error {
	regionArgs := make([][]any, 0, len(generated.Regions))
	for _, region := range generated.Regions {
		regionArgs = append(regionArgs, []any{serverID, region.ID, optionalString(region.ParentID), region.Kind, region.Name})
	}
	regionIndex, err := execConsistentBatch(ctx, tx, `INSERT INTO world_regions(server_id,region_id,parent_region_id,region_kind,display_name) VALUES($1::uuid,$2,$3,$4,$5) ON CONFLICT(server_id,region_id) DO UPDATE SET display_name=EXCLUDED.display_name WHERE world_regions.parent_region_id IS NOT DISTINCT FROM EXCLUDED.parent_region_id AND world_regions.region_kind=EXCLUDED.region_kind AND world_regions.display_name=EXCLUDED.display_name`, regionArgs)
	if err != nil {
		return fmt.Errorf("persist city region %q: %w", generated.Regions[max(regionIndex, 0)].ID, err)
	}
	entityArgs := make([][]any, 0, len(generated.Entities))
	for _, entity := range generated.Entities {
		entityArgs = append(entityArgs, []any{serverID, entity.ID, entity.EntityType, entity.RegionID, entity.GenerationKey, optionalString(entity.LocationKind)})
	}
	entityIndex, err := execConsistentBatch(ctx, tx, `INSERT INTO world_entities(server_id,entity_id,entity_type,region_id,generation_key,location_kind) VALUES($1::uuid,$2,$3,$4,$5,$6) ON CONFLICT(server_id,entity_id) DO UPDATE SET generation_key=EXCLUDED.generation_key WHERE world_entities.entity_type=EXCLUDED.entity_type AND world_entities.region_id=EXCLUDED.region_id AND world_entities.generation_key=EXCLUDED.generation_key AND world_entities.location_kind IS NOT DISTINCT FROM EXCLUDED.location_kind`, entityArgs)
	if err != nil {
		return fmt.Errorf("persist city entity %q: %w", generated.Entities[max(entityIndex, 0)].ID, err)
	}
	employmentArgs := make([][]any, 0, len(generated.EmploymentRelationships))
	for _, relation := range generated.EmploymentRelationships {
		employmentArgs = append(employmentArgs, []any{serverID, relation.ID, relation.EmployeeNPCID, relation.EmployerCompanyID, relation.Role})
	}
	if _, err := execConsistentBatch(ctx, tx, `INSERT INTO employment_relationships(server_id,relationship_id,employee_npc_id,employer_company_id,role) VALUES($1::uuid,$2,$3,$4,$5) ON CONFLICT(server_id,relationship_id) DO UPDATE SET role=EXCLUDED.role WHERE employment_relationships.employee_npc_id=EXCLUDED.employee_npc_id AND employment_relationships.employer_company_id=EXCLUDED.employer_company_id AND employment_relationships.role IS NOT DISTINCT FROM EXCLUDED.role`, employmentArgs); err != nil {
		return fmt.Errorf("persist city employment: %w", err)
	}
	personalArgs := make([][]any, 0, len(generated.PersonalRelationships))
	for _, relation := range generated.PersonalRelationships {
		personalArgs = append(personalArgs, []any{serverID, relation.ID, relation.PersonANPCID, relation.PersonBNPCID, relation.RelationshipType})
	}
	if _, err := execConsistentBatch(ctx, tx, `INSERT INTO personal_relationships(server_id,relationship_id,person_a_npc_id,person_b_npc_id,relationship_type) VALUES($1::uuid,$2,$3,$4,$5) ON CONFLICT(server_id,relationship_id) DO UPDATE SET relationship_type=EXCLUDED.relationship_type WHERE personal_relationships.person_a_npc_id=EXCLUDED.person_a_npc_id AND personal_relationships.person_b_npc_id=EXCLUDED.person_b_npc_id AND personal_relationships.relationship_type=EXCLUDED.relationship_type`, personalArgs); err != nil {
		return fmt.Errorf("persist city personal relationships: %w", err)
	}
	companyLocationArgs := make([][]any, 0, len(generated.CompanyLocations))
	for _, relation := range generated.CompanyLocations {
		companyLocationArgs = append(companyLocationArgs, []any{serverID, relation.ID, relation.CompanyID, relation.LocationID, relation.LocationRole})
	}
	if _, err := execConsistentBatch(ctx, tx, `INSERT INTO company_locations(server_id,relationship_id,company_id,location_id,location_role) VALUES($1::uuid,$2,$3,$4,$5) ON CONFLICT(server_id,relationship_id) DO UPDATE SET location_role=EXCLUDED.location_role WHERE company_locations.company_id=EXCLUDED.company_id AND company_locations.location_id=EXCLUDED.location_id AND company_locations.location_role=EXCLUDED.location_role`, companyLocationArgs); err != nil {
		return fmt.Errorf("persist city headquarters: %w", err)
	}
	return nil
}

// GetNPC rebuilds one person's generated fields and loads persisted relationships by indexed IDs.
func (store *PostgresStore) GetNPC(ctx context.Context, serverID, npcID string) (NPCProfile, error) {
	var seed string
	var version int
	var generationKey string
	err := store.db.QueryRow(ctx,
		`SELECT w.world_seed, COALESCE(city_generator.generator_version, w.generator_version), e.generation_key
		 FROM world_entities e JOIN worlds w ON w.server_id = e.server_id
		 LEFT JOIN world_city_generators city_generator
		   ON city_generator.server_id=e.server_id
		  AND city_generator.city_id=split_part(e.region_id, '/district/', 1)
		 WHERE e.server_id = $1::uuid AND e.entity_id = $2 AND e.entity_type = 'npc'`,
		serverID, npcID,
	).Scan(&seed, &version, &generationKey)
	if errors.Is(err, pgx.ErrNoRows) {
		return NPCProfile{}, ErrEntityNotFound
	}
	if err != nil {
		return NPCProfile{}, err
	}
	npc, err := GenerateNPCFromKey(seed, version, generationKey)
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

// NPCBelongsToCity validates the canonical generated ID prefix against a player's current city.
func (store *PostgresStore) NPCBelongsToCity(_ context.Context, _ string, cityID, npcID string) (bool, error) {
	return npcBelongsToCity(cityID, npcID), nil
}

func npcBelongsToCity(cityID, npcID string) bool {
	if cityID == "" || npcID == "" {
		return false
	}
	if strings.HasPrefix(npcID, "npc:"+cityID+"/") {
		return true
	}
	return cityID == "fr/lyon" && strings.HasPrefix(npcID, "npc:lyon/")
}

// GetNPCForRecords reconstructs the requested NPC and optional employer with a
// single indexed query, without loading unrelated relationships or mutations.
func (store *PostgresStore) GetNPCForRecords(ctx context.Context, serverID, npcID string, includeEmployer bool) (NPCRecordsProfile, error) {
	var result NPCRecordsProfile
	var generationKey, companyID, companyKey string
	err := store.db.QueryRow(ctx,
		`SELECT w.world_seed, COALESCE(city_generator.generator_version, w.generator_version), e.generation_key,
		        COALESCE(employer.entity_id, ''), COALESCE(employer.generation_key, '')
		 FROM world_entities e
		 JOIN worlds w ON w.server_id = e.server_id
		 LEFT JOIN world_city_generators city_generator
		   ON city_generator.server_id=e.server_id
		  AND city_generator.city_id=split_part(e.region_id, '/district/', 1)
		 LEFT JOIN LATERAL (
		     SELECT c.entity_id, c.generation_key
		     FROM employment_relationships r
		     JOIN world_entities c
		       ON c.server_id = r.server_id AND c.entity_id = r.employer_company_id
		     WHERE r.server_id = e.server_id AND r.employee_npc_id = e.entity_id AND $3::boolean
		     ORDER BY r.relationship_id
		     LIMIT 1
		 ) employer ON true
		 WHERE e.server_id = $1::uuid AND e.entity_id = $2 AND e.entity_type = 'npc'`,
		serverID, npcID, includeEmployer,
	).Scan(&result.WorldSeed, &result.GeneratorVersion, &generationKey, &companyID, &companyKey)
	if errors.Is(err, pgx.ErrNoRows) {
		return NPCRecordsProfile{}, ErrEntityNotFound
	}
	if err != nil {
		return NPCRecordsProfile{}, err
	}
	result.NPC, err = GenerateNPCFromKey(result.WorldSeed, result.GeneratorVersion, generationKey)
	if err != nil {
		return NPCRecordsProfile{}, err
	}
	if result.NPC.ID != npcID {
		return NPCRecordsProfile{}, ErrGeneratedConflict
	}
	if companyKey != "" {
		company, err := GenerateCompanyFromKey(result.WorldSeed, result.GeneratorVersion, companyKey)
		if err != nil {
			return NPCRecordsProfile{}, err
		}
		if company.ID != companyID {
			return NPCRecordsProfile{}, ErrGeneratedConflict
		}
		result.Employer = company.Name
	}
	return result, nil
}

// SearchNPCs reconstructs matching NPCs and their persisted relations from the server's world seed.
func (store *PostgresStore) SearchNPCs(ctx context.Context, serverID, query string) ([]PersonSearchRecord, error) {
	return store.searchNPCs(ctx, serverID, "", query)
}

func (store *PostgresStore) SearchNPCsInCity(ctx context.Context, serverID, cityID, query string) ([]PersonSearchRecord, error) {
	return store.searchNPCs(ctx, serverID, cityID, query)
}

func (store *PostgresStore) searchNPCs(ctx context.Context, serverID, cityID, query string) ([]PersonSearchRecord, error) {
	base, err := store.WorldForCases(ctx, serverID)
	if err != nil {
		return nil, fmt.Errorf("reconstruct world for person search: %w", err)
	}
	query = strings.ToLower(strings.TrimSpace(query))
	regions := make(map[string]string, len(base.Regions))
	entities := make(map[string]Entity, len(base.Entities))
	for _, region := range base.Regions {
		regions[region.ID] = region.Name
	}
	for _, entity := range base.Entities {
		entities[entity.ID] = entity
	}
	workplacesByCompany := make(map[string]PersonSearchWorkplace, len(base.CompanyLocations))
	for _, relation := range base.CompanyLocations {
		if relation.LocationRole == "headquarters" {
			workplacesByCompany[relation.CompanyID] = PersonSearchWorkplace{
				Company: companyName(base.Companies, relation.CompanyID),
				Address: makeAddress(entities[relation.LocationID], regions).Full,
			}
		}
	}
	matches := make([]NPC, 0)
	for _, npc := range base.NPCs {
		if cityID != "" && !npcBelongsToCity(cityID, npc.ID) {
			continue
		}
		if !strings.Contains(strings.ToLower(npc.Name), query) {
			continue
		}
		matches = append(matches, npc)
	}
	sort.Slice(matches, func(i, j int) bool {
		if matches[i].Name == matches[j].Name {
			return matches[i].ID < matches[j].ID
		}
		return matches[i].Name < matches[j].Name
	})
	if len(matches) > MaxNPCSearchResults {
		matches = matches[:MaxNPCSearchResults]
	}
	results := make([]PersonSearchRecord, 0, len(matches))
	married := make(map[string]bool)
	for _, relation := range base.PersonalRelationships {
		if relation.RelationshipType == "marriage" {
			married[relation.PersonANPCID] = true
			married[relation.PersonBNPCID] = true
		}
	}
	employmentByNPC := make(map[string][]EmploymentRelationship)
	for _, relation := range base.EmploymentRelationships {
		employmentByNPC[relation.EmployeeNPCID] = append(employmentByNPC[relation.EmployeeNPCID], relation)
	}
	for _, npc := range matches {
		workplaces := make([]PersonSearchWorkplace, 0, len(employmentByNPC[npc.ID]))
		for _, employment := range employmentByNPC[npc.ID] {
			if workplace, ok := workplacesByCompany[employment.EmployerCompanyID]; ok {
				workplaces = append(workplaces, workplace)
			}
		}
		maritalStatus := npc.MaritalStatus
		if married[npc.ID] {
			maritalStatus = "married"
		} else if maritalStatus != "single" && maritalStatus != "divorced" && maritalStatus != "widowed" {
			maritalStatus = "single"
		}
		results = append(results, PersonSearchRecord{
			NPC:        PersonSearchNPC{ID: npc.ID, Name: npc.Name, Age: npc.Age, Appearance: npc.Appearance, Status: npc.Status, Visual: npc.Visual},
			Workplaces: workplaces, MaritalStatus: maritalStatus,
		})
	}
	return results, nil
}

func personMaritalStatus(profile NPCProfile) string {
	for _, relation := range profile.PersonalRelationships {
		if relation.RelationshipType == "marriage" {
			return "married"
		}
	}
	switch profile.NPC.MaritalStatus {
	case "single", "divorced", "widowed":
		return profile.NPC.MaritalStatus
	default:
		return "single"
	}
}

func companyName(companies []Company, companyID string) string {
	for _, company := range companies {
		if company.ID == companyID {
			return company.Name
		}
	}
	return ""
}

// SearchCompanies finds companies by name or headquarters address and includes their employees.
func (store *PostgresStore) SearchCompanies(ctx context.Context, serverID, query string) ([]CompanyRecord, error) {
	base, err := store.generatedWorld(ctx, serverID)
	if err != nil {
		return nil, err
	}
	return searchCompaniesInWorld(ctx, store, serverID, base, query)
}

// SearchCompaniesInCity limits company records to the city where the player is currently using City Records.
func (store *PostgresStore) SearchCompaniesInCity(ctx context.Context, serverID, cityID, query string) ([]CompanyRecord, error) {
	base, err := store.generatedWorld(ctx, serverID)
	if err != nil {
		return nil, err
	}
	city, ok := CityByID(cityID)
	if !ok && cityID == "lyon" {
		city, ok = CityByID("fr/lyon")
	}
	if !ok {
		return []CompanyRecord{}, nil
	}
	if city.ID != "fr/lyon" {
		seed, version, err := store.worldSeedAndVersion(ctx, serverID)
		if err != nil {
			return nil, err
		}
		version, err = store.cityGeneratorVersion(ctx, serverID, city.ID, version)
		if err != nil {
			return nil, err
		}
		base, err = GenerateCity(seed, version, city.ID)
		if err != nil {
			return nil, err
		}
	}
	return searchCompaniesInWorld(ctx, store, serverID, base, query)
}

func searchCompaniesInWorld(ctx context.Context, store *PostgresStore, serverID string, base World, query string) ([]CompanyRecord, error) {
	query = strings.ToLower(strings.TrimSpace(query))
	if query == "" {
		return []CompanyRecord{}, nil
	}
	entities := make(map[string]Entity, len(base.Entities))
	regions := make(map[string]string, len(base.Regions))
	for _, entity := range base.Entities {
		entities[entity.ID] = entity
	}
	for _, region := range base.Regions {
		regions[region.ID] = region.Name
	}
	results := make([]CompanyRecord, 0)
	exactResults := make([]CompanyRecord, 0)
	for _, company := range base.Companies {
		var address Address
		for _, relation := range base.CompanyLocations {
			if relation.CompanyID == company.ID && relation.LocationRole == "headquarters" {
				address = makeAddress(entities[relation.LocationID], regions)
				break
			}
		}
		record := CompanyRecord{Company: company, Address: address}
		if strings.EqualFold(strings.TrimSpace(company.Name), query) ||
			strings.EqualFold(strings.TrimSpace(address.Name), query) ||
			strings.EqualFold(strings.TrimSpace(address.Full), query) {
			exactResults = append(exactResults, record)
		} else if strings.Contains(strings.ToLower(company.Name+" "+address.Full), query) {
			results = append(results, record)
		}
	}
	if len(exactResults) > 0 {
		results = exactResults
	}
	sort.Slice(results, func(i, j int) bool {
		if results[i].Company.Name == results[j].Company.Name {
			return results[i].Company.ID < results[j].Company.ID
		}
		return results[i].Company.Name < results[j].Company.Name
	})
	if len(results) > MaxSearchResults {
		results = results[:MaxSearchResults]
	}
	for index := range results {
		employees, err := store.EmployeesOfCompany(ctx, serverID, results[index].Company.ID)
		if err != nil {
			return nil, err
		}
		results[index].Employees = employees
	}
	return results, nil
}

// CompanyAtLocation returns every company whose location relation points to the selected building.
func (store *PostgresStore) CompanyAtLocation(ctx context.Context, serverID, locationID string) ([]CompanyRecord, error) {
	base, err := store.generatedWorld(ctx, serverID)
	if err != nil {
		return nil, err
	}
	cityID := strings.TrimPrefix(locationID, "location:")
	if districtIndex := strings.Index(cityID, "/district/"); districtIndex >= 0 {
		cityID = cityID[:districtIndex]
	}
	if cityID != "lyon" && cityID != "fr/lyon" {
		var cityVersion int
		err := store.db.QueryRow(ctx, `SELECT generator_version FROM world_city_generators WHERE server_id=$1::uuid AND city_id=$2`, serverID, cityID).Scan(&cityVersion)
		if err != nil && !errors.Is(err, pgx.ErrNoRows) {
			return nil, err
		}
		if err == nil {
			seed, _, err := store.worldSeedAndVersion(ctx, serverID)
			if err != nil {
				return nil, err
			}
			cityWorld, err := GenerateCity(seed, cityVersion, cityID)
			if err != nil {
				return nil, err
			}
			base.Entities = append(base.Entities, cityWorld.Entities...)
			base.Regions = append(base.Regions, cityWorld.Regions...)
			base.Companies = append(base.Companies, cityWorld.Companies...)
			base.CompanyLocations = append(base.CompanyLocations, cityWorld.CompanyLocations...)
		}
	}
	entities := make(map[string]Entity, len(base.Entities))
	regions := make(map[string]string, len(base.Regions))
	companies := make(map[string]Company, len(base.Companies))
	for _, entity := range base.Entities {
		entities[entity.ID] = entity
	}
	for _, region := range base.Regions {
		regions[region.ID] = region.Name
	}
	for _, company := range base.Companies {
		companies[company.ID] = company
	}

	location, exists := entities[locationID]
	if !exists {
		return []CompanyRecord{}, nil
	}
	results := make([]CompanyRecord, 0)
	for _, relation := range base.CompanyLocations {
		if relation.LocationID != locationID {
			continue
		}
		company, exists := companies[relation.CompanyID]
		if !exists {
			continue
		}
		results = append(results, CompanyRecord{Company: company, Address: makeAddress(location, regions)})
	}
	sort.Slice(results, func(i, j int) bool {
		if results[i].Company.Name == results[j].Company.Name {
			return results[i].Company.ID < results[j].Company.ID
		}
		return results[i].Company.Name < results[j].Company.Name
	})
	return results, nil
}

// CompanyDirectoryAtLocation returns staff only for a company occupying the selected building.
func (store *PostgresStore) CompanyDirectoryAtLocation(ctx context.Context, serverID, locationID, companyID string) ([]CompanyRecord, error) {
	companies, err := store.CompanyAtLocation(ctx, serverID, locationID)
	if err != nil {
		return nil, err
	}
	for _, record := range companies {
		if record.Company.ID != companyID {
			continue
		}
		employees, err := store.EmployeesOfCompany(ctx, serverID, companyID)
		if err != nil {
			return nil, err
		}
		record.Employees = employees
		return []CompanyRecord{record}, nil
	}
	return []CompanyRecord{}, nil
}

// SearchLocations finds persisted city locations by name or their complete address.
func (store *PostgresStore) SearchLocations(ctx context.Context, serverID, districtID, query string) ([]LocationRecord, error) {
	base, err := store.generatedWorld(ctx, serverID)
	if err != nil {
		return nil, err
	}
	city, cityFound := CityForLocation(districtID)
	if cityFound && city.ID != "fr/lyon" {
		seed, version, err := store.worldSeedAndVersion(ctx, serverID)
		if err != nil {
			return nil, err
		}
		version, err = store.cityGeneratorVersion(ctx, serverID, city.ID, version)
		if err != nil {
			return nil, err
		}
		base, err = GenerateCity(seed, version, city.ID)
		if err != nil {
			return nil, err
		}
	}
	query = strings.ToLower(strings.TrimSpace(query))
	regions := make(map[string]string, len(base.Regions))
	for _, region := range base.Regions {
		regions[region.ID] = region.Name
	}
	results := make([]LocationRecord, 0)
	for _, entity := range base.Entities {
		if entity.EntityType != "location" || entity.RegionID != districtID {
			continue
		}
		address := makeAddress(entity, regions)
		if strings.Contains(strings.ToLower(address.Full+" "+entity.LocationKind), query) {
			results = append(results, LocationRecord{Address: address})
		}
	}
	for _, region := range base.Regions {
		if region.Kind != "city" || districtID != region.ID+"/district/central" {
			continue
		}
		for _, service := range centralServiceLocations(region) {
			address := service.Address
			if strings.Contains(strings.ToLower(address.Full+" "+address.Kind), query) {
				results = append(results, service)
			}
		}
	}
	sort.Slice(results, func(i, j int) bool {
		if results[i].Address.Full == results[j].Address.Full {
			return results[i].Address.LocationID < results[j].Address.LocationID
		}
		return results[i].Address.Full < results[j].Address.Full
	})
	if len(results) > MaxSearchResults {
		results = results[:MaxSearchResults]
	}
	return results, nil
}

// TransitDistricts returns the generated neighborhoods and visitable locations for the GATE map.
func (store *PostgresStore) TransitDistricts(ctx context.Context, serverID string) ([]DistrictDestination, error) {
	return store.TransitDistrictsForCity(ctx, serverID, "fr/lyon")
}

// TransitDistrictsForCity returns local navigation data for one city, never GATE city routes.
func (store *PostgresStore) TransitDistrictsForCity(ctx context.Context, serverID, cityID string) ([]DistrictDestination, error) {
	city, ok := CityByID(cityID)
	if !ok {
		return nil, ErrEntityNotFound
	}
	seed, version, err := store.worldSeedAndVersion(ctx, serverID)
	if err != nil {
		return nil, err
	}
	version, err = store.cityGeneratorVersion(ctx, serverID, city.ID, version)
	if err != nil {
		return nil, err
	}
	base, err := generateCityGeography(seed, version, city.ID)
	if err != nil {
		return nil, err
	}
	regions := make(map[string]Region, len(base.Regions))
	for _, region := range base.Regions {
		regions[region.ID] = region
	}
	byID := make(map[string]*DistrictDestination)
	centralIDs := make([]string, 0)
	for _, region := range base.Regions {
		if region.Kind != "city" {
			continue
		}
		centralID := region.ID + "/district/central"
		central := &DistrictDestination{ID: centralID, Name: "Central", Locations: centralServiceLocations(region)}
		byID[centralID] = central
		centralIDs = append(centralIDs, centralID)
	}
	for _, region := range base.Regions {
		if region.Kind == "district" {
			byID[region.ID] = &DistrictDestination{ID: region.ID, Name: region.Name, Locations: []LocationRecord{}}
		}
	}
	for _, entity := range base.Entities {
		if entity.EntityType != "location" {
			continue
		}
		if district, ok := byID[entity.RegionID]; ok {
			district.Locations = append(district.Locations, LocationRecord{Address: makeAddress(entity, regionNames(regions))})
		}
	}
	result := make([]DistrictDestination, 0, len(byID))
	for _, centralID := range centralIDs {
		result = append(result, *byID[centralID])
	}
	for _, region := range base.Regions {
		if district := byID[region.ID]; district != nil && region.Kind == "district" {
			sort.Slice(district.Locations, func(i, j int) bool { return district.Locations[i].Address.Full < district.Locations[j].Address.Full })
			result = append(result, *district)
		}
	}
	return result, nil
}

func generatedCityLocations(seed string, city WorldCity, districtID, districtName string) []LocationRecord {
	streets := []string{"Rue de la Gare", "Market Street", "Foundry Lane", "Exchange Road", "Riverside Walk", "Station Street"}
	buildings := []string{"The Lantern Café", "Meridian Office", "Northstar Tailors", "Civic Exchange", "The Evening Ledger", "Bell & Finch Rooms"}
	stream := newStream(seed, GeneratorVersion5, "city/"+city.ID+"/district/"+districtID+"/locations")
	selectedStreets := shuffledCopy(streets, stream)[:2]
	selectedBuildings := shuffledCopy(buildings, stream)[:2]
	locations := make([]LocationRecord, 0, 4)
	for _, street := range selectedStreets {
		name := street + " — " + districtName
		locations = append(locations, LocationRecord{Address: Address{
			LocationID: "location:" + city.ID + "/district/" + strings.TrimPrefix(districtID, city.ID+"/district/") + "/street/" + slug(street),
			Name:       name, Kind: "street", District: districtName, City: city.Name, Country: city.Country,
			Full: strings.Join([]string{name, city.Name, city.Country}, ", "),
		}})
	}
	for _, building := range selectedBuildings {
		name := building + " — " + districtName
		locations = append(locations, LocationRecord{Address: Address{
			LocationID: "location:" + city.ID + "/district/" + strings.TrimPrefix(districtID, city.ID+"/district/") + "/building/" + slug(building),
			Name:       name, Kind: "building", District: districtName, City: city.Name, Country: city.Country,
			Full: strings.Join([]string{name, city.Name, city.Country}, ", "),
		}})
	}
	sort.Slice(locations, func(i, j int) bool { return locations[i].Address.LocationID < locations[j].Address.LocationID })
	return locations
}

func centralServiceLocations(city Region) []LocationRecord {
	cityData, ok := CityByID(city.ID)
	if !ok {
		cityData, _ = CityByID("fr/lyon")
	}
	cityKey := cityData.ID
	if cityData.ID == "fr/lyon" {
		cityKey = "lyon"
	}
	services := []struct{ id, name, status string }{
		{"police-station", city.Name + " Police Station", ""},
		{"meridian-bank", "Meridian Bank", ""},
		{"gate-station", "Central GATE Station", ""},
		{"hospital", city.Name + " Hospital", "closed"},
	}
	locations := make([]LocationRecord, 0, len(services))
	for _, service := range services {
		fullAddress := strings.Join([]string{service.name, "Central", city.Name, cityData.Country}, ", ")
		locations = append(locations, LocationRecord{Address: Address{
			LocationID: "service:" + cityKey + "/" + service.id,
			Name:       service.name,
			Kind:       "building",
			Status:     service.status,
			District:   "Central",
			City:       city.Name,
			Country:    cityData.Country,
			Full:       fullAddress,
		}})
	}
	return locations
}

func regionNames(regions map[string]Region) map[string]string {
	names := make(map[string]string, len(regions))
	for id, region := range regions {
		names[id] = region.Name
	}
	return names
}

func (store *PostgresStore) generatedWorld(ctx context.Context, serverID string) (World, error) {
	var seed string
	var version int
	err := store.db.QueryRow(ctx,
		`SELECT world_seed, generator_version FROM worlds WHERE server_id = $1::uuid`, serverID,
	).Scan(&seed, &version)
	if errors.Is(err, pgx.ErrNoRows) {
		return World{}, ErrEntityNotFound
	}
	if err != nil {
		return World{}, err
	}
	return Generate(seed, version)
}

// GeneratedWorld reconstructs the small, deterministic world assigned to a server.
func (store *PostgresStore) GeneratedWorld(ctx context.Context, serverID string) (World, error) {
	return store.generatedWorld(ctx, serverID)
}

// WorldForCases reconstructs the seeded world plus NPCs added by on-demand growth.
func (store *PostgresStore) WorldForCases(ctx context.Context, serverID string) (World, error) {
	base, err := store.generatedWorld(ctx, serverID)
	if err != nil {
		return World{}, err
	}
	seed, version, err := store.worldSeedAndVersion(ctx, serverID)
	if err != nil {
		return World{}, err
	}
	cityRows, err := store.db.Query(ctx, `SELECT city_id, generator_version FROM world_city_generators WHERE server_id=$1::uuid ORDER BY city_id`, serverID)
	if err != nil {
		return World{}, err
	}
	for cityRows.Next() {
		var cityID string
		var cityVersion int
		if err := cityRows.Scan(&cityID, &cityVersion); err != nil {
			cityRows.Close()
			return World{}, err
		}
		cityWorld, err := GenerateCity(seed, cityVersion, cityID)
		if err != nil {
			cityRows.Close()
			return World{}, err
		}
		base.Regions = append(base.Regions, cityWorld.Regions...)
		base.Entities = append(base.Entities, cityWorld.Entities...)
		base.NPCs = append(base.NPCs, cityWorld.NPCs...)
		base.Companies = append(base.Companies, cityWorld.Companies...)
		base.CompanyLocations = append(base.CompanyLocations, cityWorld.CompanyLocations...)
		base.EmploymentRelationships = append(base.EmploymentRelationships, cityWorld.EmploymentRelationships...)
		base.PersonalRelationships = append(base.PersonalRelationships, cityWorld.PersonalRelationships...)
	}
	cityRows.Close()
	if err := cityRows.Err(); err != nil {
		return World{}, err
	}
	known := make(map[string]bool, len(base.NPCs))
	for _, npc := range base.NPCs {
		known[npc.ID] = true
	}
	rows, err := store.db.Query(ctx, `SELECT entity_id, generation_key FROM world_entities WHERE server_id=$1::uuid AND entity_type='npc' ORDER BY entity_id`, serverID)
	if err != nil {
		return World{}, err
	}
	for rows.Next() {
		var id, key string
		if err := rows.Scan(&id, &key); err != nil {
			rows.Close()
			return World{}, err
		}
		if known[id] {
			continue
		}
		npc, err := GenerateNPCFromKey(seed, version, key)
		if err != nil {
			rows.Close()
			return World{}, err
		}
		if npc.ID != id {
			rows.Close()
			return World{}, ErrGeneratedConflict
		}
		base.NPCs = append(base.NPCs, npc)
		known[id] = true
	}
	rows.Close()
	if err := rows.Err(); err != nil {
		return World{}, err
	}
	if err := store.applyNPCStatuses(ctx, serverID, base.NPCs); err != nil {
		return World{}, err
	}
	rows, err = store.db.Query(ctx, `SELECT relationship_id, employee_npc_id, employer_company_id, role FROM employment_relationships WHERE server_id=$1::uuid ORDER BY relationship_id`, serverID)
	if err != nil {
		return World{}, err
	}
	base.EmploymentRelationships = base.EmploymentRelationships[:0]
	for rows.Next() {
		var r EmploymentRelationship
		if err := rows.Scan(&r.ID, &r.EmployeeNPCID, &r.EmployerCompanyID, &r.Role); err != nil {
			rows.Close()
			return World{}, err
		}
		base.EmploymentRelationships = append(base.EmploymentRelationships, r)
	}
	rows.Close()
	if err := rows.Err(); err != nil {
		return World{}, err
	}
	rows, err = store.db.Query(ctx, `SELECT relationship_id, person_a_npc_id, person_b_npc_id, relationship_type FROM personal_relationships WHERE server_id=$1::uuid ORDER BY relationship_id`, serverID)
	if err != nil {
		return World{}, err
	}
	base.PersonalRelationships = base.PersonalRelationships[:0]
	for rows.Next() {
		var r PersonalRelationship
		if err := rows.Scan(&r.ID, &r.PersonANPCID, &r.PersonBNPCID, &r.RelationshipType); err != nil {
			rows.Close()
			return World{}, err
		}
		base.PersonalRelationships = append(base.PersonalRelationships, r)
	}
	rows.Close()
	if err := rows.Err(); err != nil {
		return World{}, err
	}
	sort.Slice(base.NPCs, func(i, j int) bool { return base.NPCs[i].ID < base.NPCs[j].ID })
	return base, nil
}

// CityWorldForCases reconstructs one city's case data without rebuilding every
// city visited by the server or querying each NPC mutation separately.
func (store *PostgresStore) CityWorldForCases(ctx context.Context, serverID, cityID string) (World, error) {
	if _, ok := CityByID(cityID); !ok {
		return World{}, ErrEntityNotFound
	}
	seed, worldVersion, err := store.worldSeedAndVersion(ctx, serverID)
	if err != nil {
		return World{}, err
	}
	cityVersion, err := store.cityGeneratorVersion(ctx, serverID, cityID, worldVersion)
	if err != nil {
		return World{}, err
	}
	base, err := GenerateCity(seed, cityVersion, cityID)
	if err != nil {
		return World{}, err
	}
	known := make(map[string]bool, len(base.NPCs))
	for _, npc := range base.NPCs {
		known[npc.ID] = true
	}
	rows, err := store.db.Query(ctx,
		`SELECT entity_id, generation_key FROM world_entities
		 WHERE server_id=$1::uuid AND entity_type='npc' AND region_id LIKE $2 ORDER BY entity_id`,
		serverID, cityID+"/district/%",
	)
	if err != nil {
		return World{}, err
	}
	for rows.Next() {
		var id, key string
		if err := rows.Scan(&id, &key); err != nil {
			rows.Close()
			return World{}, err
		}
		if known[id] {
			continue
		}
		npc, err := GenerateNPCFromKey(seed, worldVersion, key)
		if err != nil {
			rows.Close()
			return World{}, err
		}
		if npc.ID != id {
			rows.Close()
			return World{}, ErrGeneratedConflict
		}
		base.NPCs = append(base.NPCs, npc)
		known[id] = true
	}
	rows.Close()
	if err := rows.Err(); err != nil {
		return World{}, err
	}

	npcIDs := make([]string, 0, len(base.NPCs))
	for _, npc := range base.NPCs {
		npcIDs = append(npcIDs, npc.ID)
	}
	rows, err = store.db.Query(ctx,
		`SELECT DISTINCT ON (entity_id) entity_id, value #>> '{}'
		 FROM world_mutations WHERE server_id=$1::uuid AND field_key='status'
		 AND entity_id=ANY($2::text[]) ORDER BY entity_id, mutation_order DESC`, serverID, npcIDs,
	)
	if err != nil {
		return World{}, err
	}
	statuses := make(map[string]string)
	for rows.Next() {
		var id, status string
		if err := rows.Scan(&id, &status); err != nil {
			rows.Close()
			return World{}, err
		}
		statuses[id] = status
	}
	rows.Close()
	if err := rows.Err(); err != nil {
		return World{}, err
	}
	for index := range base.NPCs {
		if status, ok := statuses[base.NPCs[index].ID]; ok {
			if status != "alive" && status != "dead" && status != "arrested" {
				return World{}, ErrInvalidMutation
			}
			base.NPCs[index].Status = status
		}
	}

	employmentIDs := make(map[string]bool, len(base.EmploymentRelationships))
	for _, relation := range base.EmploymentRelationships {
		employmentIDs[relation.ID] = true
	}
	rows, err = store.db.Query(ctx,
		`SELECT relationship_id, employee_npc_id, employer_company_id, COALESCE(role,'')
		 FROM employment_relationships WHERE server_id=$1::uuid AND employee_npc_id=ANY($2::text[])
		 ORDER BY relationship_id`, serverID, npcIDs,
	)
	if err != nil {
		return World{}, err
	}
	for rows.Next() {
		var relation EmploymentRelationship
		if err := rows.Scan(&relation.ID, &relation.EmployeeNPCID, &relation.EmployerCompanyID, &relation.Role); err != nil {
			rows.Close()
			return World{}, err
		}
		if !employmentIDs[relation.ID] {
			base.EmploymentRelationships = append(base.EmploymentRelationships, relation)
			employmentIDs[relation.ID] = true
		}
	}
	rows.Close()
	if err := rows.Err(); err != nil {
		return World{}, err
	}

	personalIDs := make(map[string]bool, len(base.PersonalRelationships))
	for _, relation := range base.PersonalRelationships {
		personalIDs[relation.ID] = true
	}
	rows, err = store.db.Query(ctx,
		`SELECT relationship_id, person_a_npc_id, person_b_npc_id, relationship_type
		 FROM personal_relationships WHERE server_id=$1::uuid
		 AND (person_a_npc_id=ANY($2::text[]) OR person_b_npc_id=ANY($2::text[]))
		 ORDER BY relationship_id`, serverID, npcIDs,
	)
	if err != nil {
		return World{}, err
	}
	for rows.Next() {
		var relation PersonalRelationship
		if err := rows.Scan(&relation.ID, &relation.PersonANPCID, &relation.PersonBNPCID, &relation.RelationshipType); err != nil {
			rows.Close()
			return World{}, err
		}
		if !personalIDs[relation.ID] {
			base.PersonalRelationships = append(base.PersonalRelationships, relation)
			personalIDs[relation.ID] = true
		}
	}
	rows.Close()
	if err := rows.Err(); err != nil {
		return World{}, err
	}
	sort.Slice(base.NPCs, func(i, j int) bool { return base.NPCs[i].ID < base.NPCs[j].ID })
	return base, nil
}

// PhoneContactByNumber resolves a generated employee telephone number within the caller's city.
func (store *PostgresStore) PhoneContactByNumber(ctx context.Context, serverID, cityID, number string) (PhoneContact, error) {
	digits := strings.Map(func(character rune) rune {
		if character >= '0' && character <= '9' {
			return character
		}
		return -1
	}, number)
	if len(digits) != 7 {
		return PhoneContact{}, ErrEntityNotFound
	}
	city, err := store.CityWorldForCases(ctx, serverID, cityID)
	if err != nil {
		return PhoneContact{}, err
	}
	npcs := make(map[string]NPC, len(city.NPCs))
	for _, npc := range city.NPCs {
		npcs[npc.ID] = npc
	}
	companies := make(map[string]string, len(city.Companies))
	for _, company := range city.Companies {
		companies[company.ID] = company.Name
	}
	sort.Slice(city.EmploymentRelationships, func(i, j int) bool {
		return city.EmploymentRelationships[i].ID < city.EmploymentRelationships[j].ID
	})
	for _, relationship := range city.EmploymentRelationships {
		npc, exists := npcs[relationship.EmployeeNPCID]
		if !exists || phoneNumberForNPC(npc.ID) != digits {
			continue
		}
		return PhoneContact{
			ID: npc.ID, Name: npc.Name, Company: companies[relationship.EmployerCompanyID],
			Status: npc.Status, Visual: npc.Visual,
		}, nil
	}
	return PhoneContact{}, ErrEntityNotFound
}

func phoneNumberForNPC(npcID string) string {
	hash := uint32(2166136261)
	for index := 0; index < len(npcID); index++ {
		hash = (hash ^ uint32(npcID[index])) * 16777619
	}
	digits := fmt.Sprintf("%07d", 1000000+hash%9000000)
	return digits
}

// ExpandDistrictSociety adds a deterministic pair to an existing district under a row lock.
func (store *PostgresStore) ExpandDistrictSociety(ctx context.Context, serverID, districtID string) error {
	seed, version, err := store.worldSeedAndVersion(ctx, serverID)
	if err != nil {
		return err
	}
	if city, _, _, ok := cityForDistrict(seed, districtID); ok {
		version, err = store.cityGeneratorVersion(ctx, serverID, city.ID, version)
		if err != nil {
			return err
		}
	}
	tx, err := store.db.Begin(ctx)
	if err != nil {
		return err
	}
	defer func() { _ = tx.Rollback(context.Background()) }()
	var districtExists int
	if err := tx.QueryRow(ctx, `SELECT 1 FROM world_regions WHERE server_id=$1::uuid AND region_id=$2 AND region_kind='district' FOR UPDATE`, serverID, districtID).Scan(&districtExists); err != nil {
		return err
	}
	var companyKey string
	if err := tx.QueryRow(ctx, `SELECT generation_key FROM world_entities WHERE server_id=$1::uuid AND region_id=$2 AND entity_type='company' ORDER BY entity_id LIMIT 1`, serverID, districtID).Scan(&companyKey); err != nil {
		return err
	}
	company, err := GenerateCompanyFromKey(seed, version, companyKey)
	if err != nil {
		return err
	}
	var slot int
	if err := tx.QueryRow(ctx, `SELECT COALESCE(MAX(split_part(generation_key,'/person/',2)::int),8) FROM world_entities WHERE server_id=$1::uuid AND region_id=$2 AND entity_type='npc'`, serverID, districtID).Scan(&slot); err != nil {
		return err
	}
	npcs, jobs, marriage, err := GenerateNPCPair(seed, version, districtID, slot+1, company)
	if err != nil {
		return err
	}
	entityRows := make([][]any, 0, len(npcs))
	for _, n := range npcs {
		entityRows = append(entityRows, []any{serverID, n.ID, "npc", n.RegionID, n.GenerationKey, nil})
	}
	if _, err := execConsistentBatch(ctx, tx, `INSERT INTO world_entities(server_id,entity_id,entity_type,region_id,generation_key,location_kind) VALUES($1::uuid,$2,$3,$4,$5,$6) ON CONFLICT(server_id,entity_id) DO UPDATE SET generation_key=EXCLUDED.generation_key WHERE world_entities.entity_type=EXCLUDED.entity_type AND world_entities.region_id=EXCLUDED.region_id AND world_entities.generation_key=EXCLUDED.generation_key`, entityRows); err != nil {
		return err
	}
	jobRows := make([][]any, 0, len(jobs))
	for _, r := range jobs {
		jobRows = append(jobRows, []any{serverID, r.ID, r.EmployeeNPCID, r.EmployerCompanyID, r.Role})
	}
	if _, err := execConsistentBatch(ctx, tx, `INSERT INTO employment_relationships(server_id,relationship_id,employee_npc_id,employer_company_id,role) VALUES($1::uuid,$2,$3,$4,$5) ON CONFLICT(server_id,relationship_id) DO UPDATE SET role=EXCLUDED.role WHERE employment_relationships.employee_npc_id=EXCLUDED.employee_npc_id AND employment_relationships.employer_company_id=EXCLUDED.employer_company_id AND employment_relationships.role=EXCLUDED.role`, jobRows); err != nil {
		return err
	}
	if _, err := execConsistentBatch(ctx, tx, `INSERT INTO personal_relationships(server_id,relationship_id,person_a_npc_id,person_b_npc_id,relationship_type) VALUES($1::uuid,$2,$3,$4,$5) ON CONFLICT(server_id,relationship_id) DO UPDATE SET relationship_type=EXCLUDED.relationship_type WHERE personal_relationships.person_a_npc_id=EXCLUDED.person_a_npc_id AND personal_relationships.person_b_npc_id=EXCLUDED.person_b_npc_id AND personal_relationships.relationship_type=EXCLUDED.relationship_type`, [][]any{{serverID, marriage.ID, marriage.PersonANPCID, marriage.PersonBNPCID, marriage.RelationshipType}}); err != nil {
		return err
	}
	_ = districtExists
	return tx.Commit(ctx)
}

// WorldConfig returns the seed and generator version for one server's world.
func (store *PostgresStore) WorldConfig(ctx context.Context, serverID string) (string, int, error) {
	return store.worldSeedAndVersion(ctx, serverID)
}

func makeAddress(entity Entity, regions map[string]string) Address {
	district := regions[entity.RegionID]
	cityData, ok := CityForLocation(entity.RegionID)
	if !ok {
		cityData, _ = CityByID("fr/lyon")
	}
	city := cityData.Name
	country := cityData.Country
	full := strings.Trim(strings.Join([]string{entity.Name, district, city, country}, ", "), ", ")
	return Address{LocationID: entity.ID, Name: entity.Name, Kind: entity.LocationKind, District: district, City: city, Country: country, Full: full}
}

// SetNPCStatus appends a persistent status change scoped to one server.
func (store *PostgresStore) SetNPCStatus(ctx context.Context, serverID, npcID, status string) error {
	if status != "alive" && status != "dead" && status != "arrested" {
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
	if status != "alive" && status != "dead" && status != "arrested" {
		return ErrInvalidMutation
	}
	npc.Status = status
	return nil
}

func (store *PostgresStore) applyNPCStatuses(ctx context.Context, serverID string, npcs []NPC) error {
	if len(npcs) == 0 {
		return nil
	}
	ids := make([]string, 0, len(npcs))
	byID := make(map[string]*NPC, len(npcs))
	for index := range npcs {
		ids = append(ids, npcs[index].ID)
		byID[npcs[index].ID] = &npcs[index]
	}
	rows, err := store.db.Query(ctx,
		`SELECT DISTINCT ON (entity_id) entity_id, value #>> '{}'
		 FROM world_mutations WHERE server_id=$1::uuid AND field_key='status'
		 AND entity_id=ANY($2::text[]) ORDER BY entity_id, mutation_order DESC`, serverID, ids,
	)
	if err != nil {
		return err
	}
	defer rows.Close()
	for rows.Next() {
		var id, status string
		if err := rows.Scan(&id, &status); err != nil {
			return err
		}
		if status != "alive" && status != "dead" && status != "arrested" {
			return ErrInvalidMutation
		}
		byID[id].Status = status
	}
	return rows.Err()
}

// EmployeesOfCompany uses the employer relationship index and reconstructs only matching NPCs.
func (store *PostgresStore) EmployeesOfCompany(ctx context.Context, serverID, companyID string) ([]CompanyEmployee, error) {
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
		`SELECT e.entity_id, e.generation_key, r.role,
		 COALESCE((SELECT value #>> '{}' FROM world_mutations m
		   WHERE m.server_id = e.server_id AND m.entity_id = e.entity_id AND m.field_key = 'status'
		   ORDER BY m.mutation_order DESC LIMIT 1), 'alive')
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

	employees := make([]CompanyEmployee, 0)
	for rows.Next() {
		var employeeID, key, role, status string
		if err := rows.Scan(&employeeID, &key, &role, &status); err != nil {
			return nil, err
		}
		npc, err := GenerateNPC(seed, version, key)
		if err != nil {
			return nil, err
		}
		employees = append(employees, CompanyEmployee{ID: employeeID, Name: npc.Name, Role: role, Status: status, Visual: npc.Visual})
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

func execConsistentBatch(ctx context.Context, tx pgx.Tx, query string, rows [][]any) (int, error) {
	if len(rows) == 0 {
		return -1, nil
	}
	var batch pgx.Batch
	for _, args := range rows {
		batch.Queue(query, args...)
	}
	results := tx.SendBatch(ctx, &batch)
	for index := range rows {
		result, err := results.Exec()
		if err != nil {
			_ = results.Close()
			return index, err
		}
		if result.RowsAffected() != 1 {
			_ = results.Close()
			return index, ErrGeneratedConflict
		}
	}
	if err := results.Close(); err != nil {
		return len(rows) - 1, err
	}
	return -1, nil
}

func optionalString(value string) any {
	if value == "" {
		return nil
	}
	return value
}

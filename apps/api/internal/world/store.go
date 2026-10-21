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
}

type PersonSearchNPC struct {
	ID         string `json:"id"`
	Name       string `json:"name"`
	Age        int    `json:"age"`
	Appearance string `json:"appearance"`
}

type PersonSearchWorkplace struct {
	Company string `json:"company"`
	Address string `json:"address"`
}

type Address struct {
	LocationID string `json:"locationId"`
	Name       string `json:"name"`
	Kind       string `json:"kind"`
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
	ID   string `json:"id"`
	Name string `json:"name"`
	Role string `json:"role"`
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

// GetNPCForRecords reconstructs the requested NPC and optional employer with a
// single indexed query, without loading unrelated relationships or mutations.
func (store *PostgresStore) GetNPCForRecords(ctx context.Context, serverID, npcID string, includeEmployer bool) (NPCRecordsProfile, error) {
	var result NPCRecordsProfile
	var generationKey, companyID, companyKey string
	err := store.db.QueryRow(ctx,
		`SELECT w.world_seed, w.generator_version, e.generation_key,
		        COALESCE(employer.entity_id, ''), COALESCE(employer.generation_key, '')
		 FROM world_entities e
		 JOIN worlds w ON w.server_id = e.server_id
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
	result.NPC, err = GenerateNPC(result.WorldSeed, result.GeneratorVersion, generationKey)
	if err != nil {
		return NPCRecordsProfile{}, err
	}
	if result.NPC.ID != npcID {
		return NPCRecordsProfile{}, ErrGeneratedConflict
	}
	if companyKey != "" {
		company, err := GenerateCompany(result.WorldSeed, result.GeneratorVersion, companyKey)
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
	base, err := store.WorldForCases(ctx, serverID)
	if err != nil {
		return nil, err
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
	results := make([]PersonSearchRecord, 0)
	for _, npc := range base.NPCs {
		if !strings.Contains(strings.ToLower(npc.Name), query) {
			continue
		}
		profile, err := store.GetNPC(ctx, serverID, npc.ID)
		if err != nil {
			return nil, err
		}
		workplaces := make([]PersonSearchWorkplace, 0, len(profile.Employment))
		maritalStatus := personMaritalStatus(profile)
		for _, employment := range profile.Employment {
			if workplace, ok := workplacesByCompany[employment.Company.ID]; ok {
				workplaces = append(workplaces, workplace)
			}
		}
		results = append(results, PersonSearchRecord{
			NPC:        PersonSearchNPC{ID: profile.NPC.ID, Name: profile.NPC.Name, Age: profile.NPC.Age, Appearance: profile.NPC.Appearance},
			Workplaces: workplaces, MaritalStatus: maritalStatus,
		})
	}
	sort.Slice(results, func(i, j int) bool { return results[i].NPC.Name < results[j].NPC.Name })
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
	query = strings.ToLower(strings.TrimSpace(query))
	entities := make(map[string]Entity, len(base.Entities))
	regions := make(map[string]string, len(base.Regions))
	for _, entity := range base.Entities {
		entities[entity.ID] = entity
	}
	for _, region := range base.Regions {
		regions[region.ID] = region.Name
	}
	results := make([]CompanyRecord, 0)
	for _, company := range base.Companies {
		var address Address
		for _, relation := range base.CompanyLocations {
			if relation.CompanyID == company.ID && relation.LocationRole == "headquarters" {
				address = makeAddress(entities[relation.LocationID], regions)
				break
			}
		}
		if !strings.Contains(strings.ToLower(company.Name+" "+address.Full), query) {
			continue
		}
		employees, err := store.EmployeesOfCompany(ctx, serverID, company.ID)
		if err != nil {
			return nil, err
		}
		results = append(results, CompanyRecord{Company: company, Address: address, Employees: employees})
	}
	sort.Slice(results, func(i, j int) bool { return results[i].Company.Name < results[j].Company.Name })
	return results, nil
}

// SearchLocations finds persisted city locations by name or their complete address.
func (store *PostgresStore) SearchLocations(ctx context.Context, serverID, districtID, query string) ([]LocationRecord, error) {
	base, err := store.generatedWorld(ctx, serverID)
	if err != nil {
		return nil, err
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
	sort.Slice(results, func(i, j int) bool { return results[i].Address.Full < results[j].Address.Full })
	return results, nil
}

// TransitDistricts returns the generated neighborhoods and visitable locations for the GATE map.
func (store *PostgresStore) TransitDistricts(ctx context.Context, serverID string) ([]DistrictDestination, error) {
	base, err := store.generatedWorld(ctx, serverID)
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

func centralServiceLocations(city Region) []LocationRecord {
	cityKey := strings.TrimPrefix(city.ID, "fr/")
	services := []struct{ id, name string }{
		{"police-station", city.Name + " Police Station"},
		{"meridian-bank", "Meridian Bank"},
		{"gate-station", "Central GATE Station"},
	}
	locations := make([]LocationRecord, 0, len(services))
	for _, service := range services {
		fullAddress := strings.Join([]string{service.name, "Central", city.Name, "France"}, ", ")
		locations = append(locations, LocationRecord{Address: Address{
			LocationID: "service:" + cityKey + "/" + service.id,
			Name:       service.name,
			Kind:       "building",
			District:   "Central",
			City:       city.Name,
			Country:    "France",
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
		npc, err := GenerateNPC(seed, version, key)
		if err != nil {
			rows.Close()
			return World{}, err
		}
		if npc.ID != id {
			rows.Close()
			return World{}, ErrGeneratedConflict
		}
		if err := store.applyNPCStatus(ctx, serverID, id, &npc); err != nil {
			rows.Close()
			return World{}, err
		}
		base.NPCs = append(base.NPCs, npc)
		known[id] = true
	}
	rows.Close()
	if err := rows.Err(); err != nil {
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

// ExpandDistrictSociety adds a deterministic pair to an existing district under a row lock.
func (store *PostgresStore) ExpandDistrictSociety(ctx context.Context, serverID, districtID string) error {
	seed, version, err := store.worldSeedAndVersion(ctx, serverID)
	if err != nil {
		return err
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
	company, err := GenerateCompany(seed, version, companyKey)
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
	city := "Lyon"
	country := "France"
	full := strings.Trim(strings.Join([]string{entity.Name, district, city, country}, ", "), ", ")
	return Address{LocationID: entity.ID, Name: entity.Name, Kind: entity.LocationKind, District: district, City: city, Country: country, Full: full}
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
		`SELECT e.entity_id, e.generation_key, r.role
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
		var employeeID, key, role string
		if err := rows.Scan(&employeeID, &key, &role); err != nil {
			return nil, err
		}
		npc, err := GenerateNPC(seed, version, key)
		if err != nil {
			return nil, err
		}
		employees = append(employees, CompanyEmployee{ID: employeeID, Name: npc.Name, Role: role})
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

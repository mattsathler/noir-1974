package world

import (
	"context"
	"crypto/sha256"
	"embed"
	"encoding/hex"
	"encoding/json"
	"errors"
	"fmt"
	"sort"
	"strconv"
	"strings"
	"sync"
	"time"
	"unicode/utf16"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgconn"
	"github.com/jackc/pgx/v5/pgtype"
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

type CacheObserver interface {
	ObserveCache(cache, outcome string)
}

type PostgresStore struct {
	db                       worldDatabase
	cacheObserver            CacheObserver
	countryWorldCacheMu      sync.Mutex
	countryWorldCache        map[string]countryWorldCacheEntry
	cityWorldCache           map[string]countryWorldCacheEntry
	countryExpansionCache    map[string]countryExpansionCacheEntry
	cityExpansionCache       map[string]cityWorldExpansionCacheEntry
	countryExpansionRevision map[string]uint64
	worldForCasesCache       map[string]worldForCasesCacheEntry
	worldCasesExpansionCache map[string]worldCasesExpansionCacheEntry
	worldCasesRevision       map[string]uint64
	countryCompanyCache      map[string]countryCompanyCacheEntry
}

type countryCompanyCacheEntry struct {
	companies []CompanyPortalRecord
	seed      string
	version   int
	expiresAt time.Time
}

const (
	countryWorldCacheTTL      = 10 * time.Minute
	countryWorldCacheSize     = 16
	countryCompanyCacheSize   = 16
	cityWorldCacheSize        = 32
	countryExpansionCacheSize = 64
	cityExpansionCacheSize    = 64
	worldForCasesCacheSize    = 8
	worldCasesExpansionSize   = 32
)

type countryWorldCacheEntry struct {
	world       World
	districtIDs []string
	knownNPCs   map[string]bool
	expiresAt   time.Time
}

type countryExpansionCacheEntry struct {
	npcs      []NPC
	revision  uint64
	expiresAt time.Time
}

type cityWorldExpansionCacheEntry struct {
	npcs       []NPC
	employment []EmploymentRelationship
	personal   []PersonalRelationship
	revision   uint64
	expiresAt  time.Time
}

type worldForCasesCacheEntry struct {
	world     World
	knownNPCs map[string]bool
	expiresAt time.Time
}

type worldCasesExpansionCacheEntry struct {
	npcs       []NPC
	employment []EmploymentRelationship
	personal   []PersonalRelationship
	revision   uint64
	expiresAt  time.Time
}

type persistedNPCGeneration struct {
	ID  string `json:"id"`
	Key string `json:"key"`
}

type persistedNPCStatus struct {
	ID     string `json:"id"`
	Status string `json:"status"`
}

type generatedMailEmployment struct {
	RelationshipID string `json:"relationshipId"`
	Role           string `json:"role"`
	CompanyID      string `json:"companyId"`
	CompanyKey     string `json:"companyKey"`
}

type generatedMailPersonalRelationship struct {
	RelationshipType string    `json:"relationshipType"`
	NPCID            string    `json:"npcId"`
	GenerationKey    string    `json:"generationKey"`
	Status           string    `json:"status"`
	StatusChangedAt  time.Time `json:"statusChangedAt"`
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

type CompanyPortalRecord struct {
	Company            Company `json:"company"`
	Address            Address `json:"address"`
	WebsiteDomain      string  `json:"websiteDomain"`
	WebsiteDescription string  `json:"websiteDescription"`
	SiteAddress        string  `json:"siteAddress"`
}

type CompanyEmployee struct {
	ID              string          `json:"id"`
	Name            string          `json:"name"`
	Email           string          `json:"email"`
	Role            string          `json:"role"`
	Status          string          `json:"status"`
	StatusChangedAt time.Time       `json:"-"`
	Visual          *VisualIdentity `json:"visual,omitempty"`
}

// EmailAddressForNPC produces a stable directory address without storing it.
func EmailAddressForNPC(id, name string) string {
	slug := employeeEmailNameSlug(name)
	return fmt.Sprintf("%s.%s@meridian-mail.net", slug, emailAddressIDHash(id))
}

func legacyEmailAddressForNPC(id, name string) string {
	slug := employeeEmailNameSlug(name)
	hash := sha256.Sum256([]byte(id))
	return fmt.Sprintf("%s.%x@meridian-mail.net", slug, hash[:3])
}

func employeeEmailNameSlug(name string) string {
	name = strings.ToLower(name)
	replacer := strings.NewReplacer("é", "e", "è", "e", "ê", "e", "ë", "e", "à", "a", "â", "a", "ä", "a", "î", "i", "ï", "i", "ô", "o", "ö", "o", "ù", "u", "û", "u", "ü", "u", "ç", "c", "œ", "oe", "'", "", "-", " ")
	parts := strings.Fields(replacer.Replace(name))
	slug := strings.Join(parts, ".")
	slug = strings.Map(func(r rune) rune {
		if (r >= 'a' && r <= 'z') || (r >= '0' && r <= '9') || r == '.' {
			return r
		}
		return -1
	}, slug)
	if slug == "" {
		slug = "employee"
	}
	return slug
}

func emailAddressIDHash(id string) string {
	hash := sha256.Sum256([]byte(id))
	return hex.EncodeToString(hash[:6])
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

func NewPostgresStoreWithCacheObserver(db worldDatabase, observer CacheObserver) *PostgresStore {
	return &PostgresStore{db: db, cacheObserver: observer}
}

func (store *PostgresStore) observeCache(cache string, hit bool) {
	if store.cacheObserver == nil {
		return
	}
	outcome := "miss"
	if hit {
		outcome = "hit"
	}
	store.cacheObserver.ObserveCache(cache, outcome)
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

// UpgradeMaterializedWorld adds the v7 generated content to an existing v5
// shared world while retaining its seed, IDs, mutations, and prior entities.
func (store *PostgresStore) UpgradeMaterializedWorld(ctx context.Context, serverID string) error {
	tx, err := store.db.Begin(ctx)
	if err != nil {
		return fmt.Errorf("begin world content upgrade: %w", err)
	}
	defer func() { _ = tx.Rollback(context.Background()) }()
	var seed string
	var version int
	if err := tx.QueryRow(ctx, `SELECT world_seed, generator_version FROM worlds WHERE server_id=$1::uuid FOR UPDATE`, serverID).Scan(&seed, &version); err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return ErrEntityNotFound
		}
		return fmt.Errorf("read world for content upgrade: %w", err)
	}
	if version >= GeneratorVersion7 {
		return nil
	}
	if version != GeneratorVersion5 {
		return fmt.Errorf("world content upgrade requires generator version %d; found %d", GeneratorVersion5, version)
	}
	base, err := Generate(seed, GeneratorVersion7)
	if err != nil {
		return err
	}
	if err := persistGeneratedCity(ctx, tx, serverID, base); err != nil {
		return err
	}
	rows, err := tx.Query(ctx, `SELECT city_id, generator_version FROM world_city_generators WHERE server_id=$1::uuid ORDER BY city_id FOR UPDATE`, serverID)
	if err != nil {
		return fmt.Errorf("list materialized cities: %w", err)
	}
	type cityVersion struct {
		id      string
		version int
	}
	cities := make([]cityVersion, 0)
	for rows.Next() {
		var city cityVersion
		if err := rows.Scan(&city.id, &city.version); err != nil {
			rows.Close()
			return err
		}
		cities = append(cities, city)
	}
	if err := rows.Err(); err != nil {
		rows.Close()
		return err
	}
	rows.Close()
	for _, city := range cities {
		if city.version >= GeneratorVersion7 {
			continue
		}
		if city.version != GeneratorVersion5 {
			return fmt.Errorf("city %q content upgrade requires generator version %d; found %d", city.id, GeneratorVersion5, city.version)
		}
		generated, err := GenerateCity(seed, GeneratorVersion7, city.id)
		if err != nil {
			return fmt.Errorf("generate upgraded city %q: %w", city.id, err)
		}
		if err := persistGeneratedCity(ctx, tx, serverID, generated); err != nil {
			return fmt.Errorf("persist upgraded city %q: %w", city.id, err)
		}
		if _, err := tx.Exec(ctx, `UPDATE world_city_generators SET generator_version=$3 WHERE server_id=$1::uuid AND city_id=$2`, serverID, city.id, GeneratorVersion7); err != nil {
			return fmt.Errorf("record upgraded city %q: %w", city.id, err)
		}
	}
	if _, err := tx.Exec(ctx, `UPDATE worlds SET generator_version=$2 WHERE server_id=$1::uuid`, serverID, GeneratorVersion7); err != nil {
		return fmt.Errorf("record upgraded world version: %w", err)
	}
	if err := tx.Commit(ctx); err != nil {
		return fmt.Errorf("commit world content upgrade: %w", err)
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
	if version < LatestGeneratorVersion {
		version = LatestGeneratorVersion
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
	// The city marker is inserted before persistence in this transaction. It
	// serializes first-time materialization for this city, so no rows for this
	// city can already exist when this transaction wins the marker insert.
	// COPY avoids executing one INSERT per generated entity on this cold path.
	if err := copyGeneratedCity(ctx, tx, serverID, generated); err != nil {
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
		if fallback < LatestGeneratorVersion {
			return LatestGeneratorVersion, nil
		}
		return fallback, nil
	}
	return version, err
}

func copyGeneratedCity(ctx context.Context, tx pgx.Tx, serverID string, generated World) error {
	for _, npc := range generated.NPCs {
		if generated.GeneratorVersion >= GeneratorVersion6 {
			if npc.Visual == nil || ValidateVisualIdentityForPresentation(*npc.Visual, npc.Presentation) != nil {
				return fmt.Errorf("persist NPC %q: %w", npc.ID, ErrInvalidVisualIdentity)
			}
		}
	}

	regionRows := make([][]any, 0, len(generated.Regions))
	for _, region := range generated.Regions {
		regionRows = append(regionRows, []any{serverID, region.ID, optionalString(region.ParentID), region.Kind, region.Name})
	}
	if _, err := tx.CopyFrom(ctx, pgx.Identifier{"world_regions"}, []string{"server_id", "region_id", "parent_region_id", "region_kind", "display_name"}, pgx.CopyFromRows(regionRows)); err != nil {
		return fmt.Errorf("copy city regions: %w", err)
	}

	entityRows := make([][]any, 0, len(generated.Entities))
	for _, entity := range generated.Entities {
		entityRows = append(entityRows, []any{serverID, entity.ID, entity.EntityType, entity.RegionID, entity.GenerationKey, optionalString(entity.LocationKind)})
	}
	if _, err := tx.CopyFrom(ctx, pgx.Identifier{"world_entities"}, []string{"server_id", "entity_id", "entity_type", "region_id", "generation_key", "location_kind"}, pgx.CopyFromRows(entityRows)); err != nil {
		return fmt.Errorf("copy city entities: %w", err)
	}

	employmentRows := make([][]any, 0, len(generated.EmploymentRelationships))
	for _, relation := range generated.EmploymentRelationships {
		employmentRows = append(employmentRows, []any{serverID, relation.ID, relation.EmployeeNPCID, relation.EmployerCompanyID, relation.Role})
	}
	if _, err := tx.CopyFrom(ctx, pgx.Identifier{"employment_relationships"}, []string{"server_id", "relationship_id", "employee_npc_id", "employer_company_id", "role"}, pgx.CopyFromRows(employmentRows)); err != nil {
		return fmt.Errorf("copy city employment: %w", err)
	}

	personalRows := make([][]any, 0, len(generated.PersonalRelationships))
	for _, relation := range generated.PersonalRelationships {
		personalRows = append(personalRows, []any{serverID, relation.ID, relation.PersonANPCID, relation.PersonBNPCID, relation.RelationshipType})
	}
	if _, err := tx.CopyFrom(ctx, pgx.Identifier{"personal_relationships"}, []string{"server_id", "relationship_id", "person_a_npc_id", "person_b_npc_id", "relationship_type"}, pgx.CopyFromRows(personalRows)); err != nil {
		return fmt.Errorf("copy city personal relationships: %w", err)
	}

	companyLocationRows := make([][]any, 0, len(generated.CompanyLocations))
	for _, relation := range generated.CompanyLocations {
		companyLocationRows = append(companyLocationRows, []any{serverID, relation.ID, relation.CompanyID, relation.LocationID, relation.LocationRole})
	}
	if _, err := tx.CopyFrom(ctx, pgx.Identifier{"company_locations"}, []string{"server_id", "relationship_id", "company_id", "location_id", "location_role"}, pgx.CopyFromRows(companyLocationRows)); err != nil {
		return fmt.Errorf("copy city headquarters: %w", err)
	}
	return nil
}

func persistGeneratedCity(ctx context.Context, tx pgx.Tx, serverID string, generated World) error {
	for _, npc := range generated.NPCs {
		if generated.GeneratorVersion >= GeneratorVersion6 {
			if npc.Visual == nil || ValidateVisualIdentityForPresentation(*npc.Visual, npc.Presentation) != nil {
				return fmt.Errorf("persist NPC %q: %w", npc.ID, ErrInvalidVisualIdentity)
			}
		}
	}
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
	profile, _, _, err := store.getNPCWithWorldConfig(ctx, serverID, npcID, false)
	return profile, err
}

// GetNPCWithWorldConfig returns the NPC profile and the world configuration
// needed by deterministic consumers without issuing a second worlds query.
func (store *PostgresStore) GetNPCWithWorldConfig(ctx context.Context, serverID, npcID string) (NPCProfile, string, int, error) {
	return store.getNPCWithWorldConfig(ctx, serverID, npcID, true)
}

func (store *PostgresStore) getNPCWithWorldConfig(ctx context.Context, serverID, npcID string, combineWorldAndStatus bool) (NPCProfile, string, int, error) {
	var seed string
	var version, worldVersion int
	var generationKey, status string
	var statusChangedAt pgtype.Timestamptz
	var employmentJSON, personalJSON []byte
	var err error
	if combineWorldAndStatus {
		err = store.db.QueryRow(ctx,
			`WITH target AS (
			 SELECT w.world_seed AS seed, w.generator_version AS world_version,
			        COALESCE(city_generator.generator_version, w.generator_version) AS city_version,
			        e.entity_id, e.generation_key,
				 COALESCE(status.value #>> '{}', 'alive') AS status, status.created_at
			 FROM world_entities e JOIN worlds w ON w.server_id=e.server_id
			 LEFT JOIN world_city_generators city_generator
			   ON city_generator.server_id=e.server_id
			  AND city_generator.city_id=split_part(e.region_id, '/district/', 1)
			 LEFT JOIN LATERAL (
			   SELECT m.value, m.created_at FROM world_mutations m
			   WHERE m.server_id=e.server_id AND m.entity_id=e.entity_id AND m.field_key='status'
			   ORDER BY m.mutation_order DESC LIMIT 1
			 ) status ON true
			 WHERE e.server_id=$1::uuid AND e.entity_id=$2 AND e.entity_type='npc'
			)
			SELECT target.seed, target.world_version, target.city_version, target.generation_key, target.status, target.created_at,
		        COALESCE((SELECT jsonb_agg(jsonb_build_object(
		          'relationshipId', r.relationship_id, 'role', COALESCE(r.role,''),
		          'companyId', c.entity_id, 'companyKey', c.generation_key) ORDER BY r.relationship_id)
		          FROM employment_relationships r JOIN world_entities c
		            ON c.server_id=r.server_id AND c.entity_id=r.employer_company_id
		          WHERE r.server_id=$1::uuid AND r.employee_npc_id=target.entity_id), '[]'::jsonb),
		        COALESCE((SELECT jsonb_agg(jsonb_build_object(
		          'relationshipType', r.relationship_type, 'npcId', other.entity_id,
		          'generationKey', other.generation_key,
		          'status', COALESCE((SELECT m.value #>> '{}' FROM world_mutations m
		            WHERE m.server_id=other.server_id AND m.entity_id=other.entity_id AND m.field_key='status'
		            ORDER BY m.mutation_order DESC LIMIT 1), 'alive'),
		          'statusChangedAt', (SELECT m.created_at FROM world_mutations m
		            WHERE m.server_id=other.server_id AND m.entity_id=other.entity_id AND m.field_key='status'
		            ORDER BY m.mutation_order DESC LIMIT 1)) ORDER BY r.relationship_type, other.entity_id)
		          FROM personal_relationships r JOIN world_entities other
		            ON other.server_id=r.server_id AND other.entity_id=CASE
		              WHEN r.person_a_npc_id=target.entity_id THEN r.person_b_npc_id ELSE r.person_a_npc_id END
		          WHERE r.server_id=$1::uuid AND (r.person_a_npc_id=target.entity_id OR r.person_b_npc_id=target.entity_id)), '[]'::jsonb)
			FROM target`,
			serverID, npcID,
		).Scan(&seed, &worldVersion, &version, &generationKey, &status, &statusChangedAt, &employmentJSON, &personalJSON)
	} else {
		err = store.db.QueryRow(ctx,
			`SELECT w.world_seed, COALESCE(city_generator.generator_version, w.generator_version), e.generation_key
			 FROM world_entities e JOIN worlds w ON w.server_id = e.server_id
			 LEFT JOIN world_city_generators city_generator
			   ON city_generator.server_id=e.server_id
			  AND city_generator.city_id=split_part(e.region_id, '/district/', 1)
			 WHERE e.server_id = $1::uuid AND e.entity_id = $2 AND e.entity_type = 'npc'`,
			serverID, npcID,
		).Scan(&seed, &version, &generationKey)
		worldVersion = version
	}
	if errors.Is(err, pgx.ErrNoRows) {
		return NPCProfile{}, "", 0, ErrEntityNotFound
	}
	if err != nil {
		return NPCProfile{}, "", 0, err
	}
	npc, err := GenerateNPCFromKey(seed, version, generationKey)
	if err != nil {
		return NPCProfile{}, "", 0, err
	}
	if npc.ID != npcID {
		return NPCProfile{}, "", 0, ErrGeneratedConflict
	}
	if combineWorldAndStatus {
		if status != "alive" && status != "missing" && status != "dead" && status != "arrested" {
			return NPCProfile{}, "", 0, ErrInvalidMutation
		}
		npc.Status = status
		if statusChangedAt.Valid {
			npc.StatusChangedAt = statusChangedAt.Time
		}
	}
	profile := NPCProfile{
		NPC:                   npc,
		Employment:            make([]EmploymentView, 0),
		PersonalRelationships: make([]PersonalRelationshipView, 0),
	}
	if !combineWorldAndStatus {
		if err := store.applyNPCStatus(ctx, serverID, npcID, &profile.NPC); err != nil {
			return NPCProfile{}, "", 0, err
		}
		if err := store.loadEmployment(ctx, serverID, seed, version, npcID, &profile); err != nil {
			return NPCProfile{}, "", 0, err
		}
		if err := store.loadPersonalRelationships(ctx, serverID, seed, version, npcID, &profile); err != nil {
			return NPCProfile{}, "", 0, err
		}
		return profile, seed, worldVersion, nil
	}
	var generatedEmployment []generatedMailEmployment
	if err := json.Unmarshal(employmentJSON, &generatedEmployment); err != nil {
		return NPCProfile{}, "", 0, fmt.Errorf("decode NPC employment profile: %w", err)
	}
	for _, relation := range generatedEmployment {
		company, err := GenerateCompany(seed, version, relation.CompanyKey)
		if err != nil {
			return NPCProfile{}, "", 0, err
		}
		if company.ID != relation.CompanyID {
			return NPCProfile{}, "", 0, ErrGeneratedConflict
		}
		profile.Employment = append(profile.Employment, EmploymentView{
			RelationshipID: relation.RelationshipID, Role: relation.Role, Company: company,
		})
	}
	var generatedPersonal []generatedMailPersonalRelationship
	if err := json.Unmarshal(personalJSON, &generatedPersonal); err != nil {
		return NPCProfile{}, "", 0, fmt.Errorf("decode NPC personal relationships: %w", err)
	}
	for _, relation := range generatedPersonal {
		other, err := GenerateNPC(seed, version, relation.GenerationKey)
		if err != nil {
			return NPCProfile{}, "", 0, err
		}
		if other.ID != relation.NPCID {
			return NPCProfile{}, "", 0, ErrGeneratedConflict
		}
		if relation.Status != "alive" && relation.Status != "missing" && relation.Status != "dead" && relation.Status != "arrested" {
			return NPCProfile{}, "", 0, ErrInvalidMutation
		}
		other.Status = relation.Status
		other.StatusChangedAt = relation.StatusChangedAt
		profile.PersonalRelationships = append(profile.PersonalRelationships, PersonalRelationshipView{
			RelationshipType: relation.RelationshipType, NPC: other,
		})
	}
	return profile, seed, worldVersion, nil
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

// SearchNPCsInCountry searches generated and persisted NPCs within one country.
func (store *PostgresStore) SearchNPCsInCountry(ctx context.Context, serverID, countryID, query string) ([]PersonSearchRecord, error) {
	cities := make([]WorldCity, 0)
	for _, city := range Cities() {
		if city.CountryID == countryID {
			cities = append(cities, city)
		}
	}
	if len(cities) == 0 {
		return nil, ErrEntityNotFound
	}

	cityIDs := make([]string, 0, len(cities))
	for _, city := range cities {
		cityIDs = append(cityIDs, city.ID)
	}
	cityVersions := make(map[string]int, len(cityIDs))
	rows, err := store.db.Query(ctx,
		`SELECT w.world_seed, w.generator_version, COALESCE(city_generator.city_id, ''),
		        COALESCE(city_generator.generator_version, 0)
		 FROM worlds w
		 LEFT JOIN world_city_generators city_generator
		   ON city_generator.server_id = w.server_id AND city_generator.city_id=ANY($2::text[])
		 WHERE w.server_id=$1::uuid ORDER BY city_generator.city_id`, serverID, cityIDs,
	)
	if err != nil {
		return nil, err
	}
	var seed string
	var worldVersion int
	scanCityVersion := func() error {
		var cityID string
		var version int
		if err := rows.Scan(&seed, &worldVersion, &cityID, &version); err != nil {
			return err
		}
		if cityID != "" {
			cityVersions[cityID] = version
		}
		return nil
	}
	if !rows.Next() {
		rows.Close()
		if err := rows.Err(); err != nil {
			return nil, err
		}
		return nil, ErrEntityNotFound
	}
	if err := scanCityVersion(); err != nil {
		rows.Close()
		return nil, err
	}
	for rows.Next() {
		if err := scanCityVersion(); err != nil {
			rows.Close()
			return nil, err
		}
	}
	rows.Close()
	if err := rows.Err(); err != nil {
		return nil, err
	}

	cacheKey := countryWorldCacheKey(seed, worldVersion, cities, cityVersions)
	base, districtIDs, known, cached := store.cachedCountryWorld(cacheKey)
	store.observeCache("country_world_base", cached)
	if !cached {
		base = World{GeneratorVersion: worldVersion}
		districtIDs = make([]string, 0, len(cities)*20)
		known = make(map[string]bool)
		for _, city := range cities {
			version := countryCityGeneratorVersion(city, worldVersion, cityVersions)
			cityWorld, err := GenerateCity(seed, version, city.ID)
			if err != nil {
				return nil, err
			}
			base.Regions = append(base.Regions, cityWorld.Regions...)
			base.Entities = append(base.Entities, cityWorld.Entities...)
			base.NPCs = append(base.NPCs, cityWorld.NPCs...)
			base.Companies = append(base.Companies, cityWorld.Companies...)
			base.CompanyLocations = append(base.CompanyLocations, cityWorld.CompanyLocations...)
			base.EmploymentRelationships = append(base.EmploymentRelationships, cityWorld.EmploymentRelationships...)
			base.PersonalRelationships = append(base.PersonalRelationships, cityWorld.PersonalRelationships...)
			for _, region := range cityWorld.Regions {
				if region.Kind == "district" {
					districtIDs = append(districtIDs, region.ID)
				}
			}
			for _, npc := range cityWorld.NPCs {
				known[npc.ID] = true
			}
		}
		store.cacheCountryWorld(cacheKey, base, districtIDs, known)
	}
	if len(districtIDs) == 0 {
		return nil, ErrEntityNotFound
	}
	expansionKey := serverID + "\x00" + countryID + "\x00" + cacheKey
	expandedNPCs, expansionsCached, expansionRevision := store.cachedCountryExpansions(serverID, countryID, expansionKey)
	store.observeCache("country_world_expansions", expansionsCached)
	if !expansionsCached {
		rows, err = store.db.Query(ctx,
			`SELECT entity_id, generation_key FROM world_entities
			 WHERE server_id=$1::uuid AND entity_type='npc' AND region_id=ANY($2::text[])
			 ORDER BY entity_id`, serverID, districtIDs,
		)
		if err != nil {
			return nil, err
		}
		expandedNPCs = make([]NPC, 0)
		for rows.Next() {
			var id, key string
			if err := rows.Scan(&id, &key); err != nil {
				rows.Close()
				return nil, err
			}
			if known[id] {
				continue
			}
			npc, err := GenerateNPCFromKey(seed, worldVersion, key)
			if err != nil {
				rows.Close()
				return nil, err
			}
			if npc.ID != id {
				rows.Close()
				return nil, ErrGeneratedConflict
			}
			expandedNPCs = append(expandedNPCs, npc)
		}
		rows.Close()
		if err := rows.Err(); err != nil {
			return nil, err
		}
		store.cacheCountryExpansions(expansionKey, serverID, countryID, expansionRevision, expandedNPCs)
	}

	matches := personSearchMatchesWithAdditional(base.NPCs, expandedNPCs, "", query)
	base.NPCs = matches
	npcIDs := make([]string, 0, len(matches))
	matchedIDs := make(map[string]bool, len(matches))
	for _, npc := range matches {
		npcIDs = append(npcIDs, npc.ID)
		matchedIDs[npc.ID] = true
	}
	var employmentRows []EmploymentRelationship
	var personalRows []PersonalRelationship
	var statusErr, employmentErr, personalErr error
	var reads sync.WaitGroup
	reads.Add(3)
	go func() {
		defer reads.Done()
		statusErr = store.applyNPCStatuses(ctx, serverID, base.NPCs)
	}()
	go func() {
		defer reads.Done()
		employmentRows, employmentErr = store.cityEmploymentRelationships(ctx, serverID, npcIDs)
	}()
	go func() {
		defer reads.Done()
		personalRows, personalErr = store.cityPersonalRelationships(ctx, serverID, npcIDs)
	}()
	reads.Wait()
	if statusErr != nil {
		return nil, statusErr
	}
	if employmentErr != nil {
		return nil, employmentErr
	}
	if personalErr != nil {
		return nil, personalErr
	}

	employmentByID := make(map[string]bool)
	employment := make([]EmploymentRelationship, 0)
	for _, relation := range base.EmploymentRelationships {
		if matchedIDs[relation.EmployeeNPCID] {
			employment = append(employment, relation)
			employmentByID[relation.ID] = true
		}
	}
	for _, relation := range employmentRows {
		if !employmentByID[relation.ID] {
			employment = append(employment, relation)
			employmentByID[relation.ID] = true
		}
	}
	base.EmploymentRelationships = employment

	personalByID := make(map[string]bool)
	personal := make([]PersonalRelationship, 0)
	for _, relation := range base.PersonalRelationships {
		if relation.RelationshipType == "marriage" && (matchedIDs[relation.PersonANPCID] || matchedIDs[relation.PersonBNPCID]) {
			personal = append(personal, relation)
			personalByID[relation.ID] = true
		}
	}
	for _, relation := range personalRows {
		if relation.RelationshipType == "marriage" && !personalByID[relation.ID] {
			personal = append(personal, relation)
			personalByID[relation.ID] = true
		}
	}
	base.PersonalRelationships = personal
	return personSearchRecords(base, matches), nil
}

func countryCityGeneratorVersion(city WorldCity, worldVersion int, cityVersions map[string]int) int {
	if city.ID == "fr/lyon" {
		return worldVersion
	}
	if version, ok := cityVersions[city.ID]; ok {
		return version
	}
	if worldVersion < LatestGeneratorVersion {
		return LatestGeneratorVersion
	}
	return worldVersion
}

func countryWorldCacheKey(seed string, worldVersion int, cities []WorldCity, cityVersions map[string]int) string {
	var scope strings.Builder
	for _, city := range cities {
		fmt.Fprintf(&scope, "%s:%d;", city.ID, countryCityGeneratorVersion(city, worldVersion, cityVersions))
	}
	seedHash := sha256.Sum256([]byte(seed))
	return fmt.Sprintf("%x:%d:%s", seedHash, worldVersion, scope.String())
}

func (store *PostgresStore) cachedCountryWorld(key string) (World, []string, map[string]bool, bool) {
	now := time.Now()
	store.countryWorldCacheMu.Lock()
	defer store.countryWorldCacheMu.Unlock()
	entry, exists := store.countryWorldCache[key]
	if !exists {
		return World{}, nil, nil, false
	}
	if !now.Before(entry.expiresAt) {
		delete(store.countryWorldCache, key)
		return World{}, nil, nil, false
	}
	generated := entry.world
	generated.NPCs = append([]NPC(nil), entry.world.NPCs...)
	return generated, entry.districtIDs, entry.knownNPCs, true
}

func (store *PostgresStore) cacheCountryWorld(key string, generated World, districtIDs []string, knownNPCs map[string]bool) {
	generated.NPCs = append([]NPC(nil), generated.NPCs...)
	entry := countryWorldCacheEntry{
		world: generated, districtIDs: districtIDs,
		knownNPCs: knownNPCs, expiresAt: time.Now().Add(countryWorldCacheTTL),
	}
	store.countryWorldCacheMu.Lock()
	defer store.countryWorldCacheMu.Unlock()
	if store.countryWorldCache == nil {
		store.countryWorldCache = make(map[string]countryWorldCacheEntry)
	}
	if _, exists := store.countryWorldCache[key]; !exists && len(store.countryWorldCache) >= countryWorldCacheSize {
		oldestKey := ""
		var oldestExpiry time.Time
		for cachedKey, cachedEntry := range store.countryWorldCache {
			if oldestKey == "" || cachedEntry.expiresAt.Before(oldestExpiry) {
				oldestKey, oldestExpiry = cachedKey, cachedEntry.expiresAt
			}
		}
		delete(store.countryWorldCache, oldestKey)
	}
	store.countryWorldCache[key] = entry
}

func cityWorldCacheKey(seed string, worldVersion, cityVersion int, cityID string) string {
	seedHash := sha256.Sum256([]byte(seed))
	return fmt.Sprintf("city:%x:%d:%d:%s", seedHash, worldVersion, cityVersion, cityID)
}

func cloneCityBase(generated World) World {
	generated.NPCs = append([]NPC(nil), generated.NPCs...)
	generated.EmploymentRelationships = append([]EmploymentRelationship(nil), generated.EmploymentRelationships...)
	generated.PersonalRelationships = append([]PersonalRelationship(nil), generated.PersonalRelationships...)
	return generated
}

func (store *PostgresStore) cachedCityWorld(key string) (World, []string, map[string]bool, bool) {
	now := time.Now()
	store.countryWorldCacheMu.Lock()
	defer store.countryWorldCacheMu.Unlock()
	entry, exists := store.cityWorldCache[key]
	if !exists {
		return World{}, nil, nil, false
	}
	if !now.Before(entry.expiresAt) {
		delete(store.cityWorldCache, key)
		return World{}, nil, nil, false
	}
	return cloneCityBase(entry.world), entry.districtIDs, entry.knownNPCs, true
}

func (store *PostgresStore) cacheCityWorld(key string, generated World, districtIDs []string, knownNPCs map[string]bool) {
	entry := countryWorldCacheEntry{
		world: cloneCityBase(generated), districtIDs: districtIDs,
		knownNPCs: knownNPCs, expiresAt: time.Now().Add(countryWorldCacheTTL),
	}
	store.countryWorldCacheMu.Lock()
	defer store.countryWorldCacheMu.Unlock()
	if store.cityWorldCache == nil {
		store.cityWorldCache = make(map[string]countryWorldCacheEntry)
	}
	if _, exists := store.cityWorldCache[key]; !exists && len(store.cityWorldCache) >= cityWorldCacheSize {
		oldestKey := ""
		var oldestExpiry time.Time
		for cachedKey, cachedEntry := range store.cityWorldCache {
			if oldestKey == "" || cachedEntry.expiresAt.Before(oldestExpiry) {
				oldestKey, oldestExpiry = cachedKey, cachedEntry.expiresAt
			}
		}
		delete(store.cityWorldCache, oldestKey)
	}
	store.cityWorldCache[key] = entry
}

func countryExpansionScope(serverID, countryID string) string {
	return serverID + "\x00" + countryID + "\x00"
}

func (store *PostgresStore) cachedCityExpansions(serverID, countryID, key string) ([]NPC, []EmploymentRelationship, []PersonalRelationship, bool, uint64) {
	now := time.Now()
	store.countryWorldCacheMu.Lock()
	defer store.countryWorldCacheMu.Unlock()
	scope := countryExpansionScope(serverID, countryID)
	revision := store.countryExpansionRevision[scope]
	entry, exists := store.cityExpansionCache[key]
	if !exists || entry.revision != revision || !now.Before(entry.expiresAt) {
		if exists && !now.Before(entry.expiresAt) {
			delete(store.cityExpansionCache, key)
		}
		return nil, nil, nil, false, revision
	}
	return append([]NPC(nil), entry.npcs...), append([]EmploymentRelationship(nil), entry.employment...), append([]PersonalRelationship(nil), entry.personal...), true, revision
}

func (store *PostgresStore) cacheCityExpansions(key, serverID, countryID string, revision uint64, npcs []NPC, employment []EmploymentRelationship, personal []PersonalRelationship) {
	scope := countryExpansionScope(serverID, countryID)
	entry := cityWorldExpansionCacheEntry{
		npcs:       append([]NPC(nil), npcs...),
		employment: append([]EmploymentRelationship(nil), employment...),
		personal:   append([]PersonalRelationship(nil), personal...),
		revision:   revision, expiresAt: time.Now().Add(countryWorldCacheTTL),
	}
	store.countryWorldCacheMu.Lock()
	defer store.countryWorldCacheMu.Unlock()
	if store.countryExpansionRevision[scope] != revision {
		return
	}
	if store.cityExpansionCache == nil {
		store.cityExpansionCache = make(map[string]cityWorldExpansionCacheEntry)
	}
	if _, exists := store.cityExpansionCache[key]; !exists && len(store.cityExpansionCache) >= cityExpansionCacheSize {
		oldestKey := ""
		var oldestExpiry time.Time
		for cachedKey, cachedEntry := range store.cityExpansionCache {
			if oldestKey == "" || cachedEntry.expiresAt.Before(oldestExpiry) {
				oldestKey, oldestExpiry = cachedKey, cachedEntry.expiresAt
			}
		}
		delete(store.cityExpansionCache, oldestKey)
	}
	store.cityExpansionCache[key] = entry
}

func worldForCasesCacheKey(seed string, worldVersion int, cityVersions map[string]int) string {
	cityIDs := make([]string, 0, len(cityVersions))
	for cityID := range cityVersions {
		cityIDs = append(cityIDs, cityID)
	}
	sort.Strings(cityIDs)
	var cityScope strings.Builder
	for _, cityID := range cityIDs {
		fmt.Fprintf(&cityScope, "%s:%d;", cityID, cityVersions[cityID])
	}
	seedHash := sha256.Sum256([]byte(seed))
	return fmt.Sprintf("world:%x:%d:%s", seedHash, worldVersion, cityScope.String())
}

func worldCasesScope(serverID string) string {
	return serverID + "\x00"
}

func cloneWorldForCasesBase(base World) World {
	base.NPCs = append([]NPC(nil), base.NPCs...)
	base.EmploymentRelationships = append([]EmploymentRelationship(nil), base.EmploymentRelationships...)
	base.PersonalRelationships = append([]PersonalRelationship(nil), base.PersonalRelationships...)
	return base
}

func (store *PostgresStore) cachedWorldForCases(key string) (World, map[string]bool, bool) {
	now := time.Now()
	store.countryWorldCacheMu.Lock()
	defer store.countryWorldCacheMu.Unlock()
	entry, exists := store.worldForCasesCache[key]
	if !exists {
		return World{}, nil, false
	}
	if !now.Before(entry.expiresAt) {
		delete(store.worldForCasesCache, key)
		return World{}, nil, false
	}
	return cloneWorldForCasesBase(entry.world), entry.knownNPCs, true
}

func (store *PostgresStore) cacheWorldForCases(key string, base World, knownNPCs map[string]bool) {
	entry := worldForCasesCacheEntry{
		world: cloneWorldForCasesBase(base), knownNPCs: knownNPCs,
		expiresAt: time.Now().Add(countryWorldCacheTTL),
	}
	store.countryWorldCacheMu.Lock()
	defer store.countryWorldCacheMu.Unlock()
	if store.worldForCasesCache == nil {
		store.worldForCasesCache = make(map[string]worldForCasesCacheEntry)
	}
	if _, exists := store.worldForCasesCache[key]; !exists && len(store.worldForCasesCache) >= worldForCasesCacheSize {
		oldestKey := ""
		var oldestExpiry time.Time
		for cachedKey, cachedEntry := range store.worldForCasesCache {
			if oldestKey == "" || cachedEntry.expiresAt.Before(oldestExpiry) {
				oldestKey, oldestExpiry = cachedKey, cachedEntry.expiresAt
			}
		}
		delete(store.worldForCasesCache, oldestKey)
	}
	store.worldForCasesCache[key] = entry
}

func (store *PostgresStore) cachedWorldCasesExpansions(serverID, key string) ([]NPC, []EmploymentRelationship, []PersonalRelationship, bool, uint64) {
	now := time.Now()
	store.countryWorldCacheMu.Lock()
	defer store.countryWorldCacheMu.Unlock()
	revision := store.worldCasesRevision[serverID]
	entry, exists := store.worldCasesExpansionCache[key]
	if !exists || entry.revision != revision || !now.Before(entry.expiresAt) {
		if exists && !now.Before(entry.expiresAt) {
			delete(store.worldCasesExpansionCache, key)
		}
		return nil, nil, nil, false, revision
	}
	return append([]NPC(nil), entry.npcs...), append([]EmploymentRelationship(nil), entry.employment...), append([]PersonalRelationship(nil), entry.personal...), true, revision
}

func (store *PostgresStore) cacheWorldCasesExpansions(serverID, key string, revision uint64, npcs []NPC, employment []EmploymentRelationship, personal []PersonalRelationship) {
	entry := worldCasesExpansionCacheEntry{
		npcs: append([]NPC(nil), npcs...), employment: append([]EmploymentRelationship(nil), employment...),
		personal: append([]PersonalRelationship(nil), personal...), revision: revision,
		expiresAt: time.Now().Add(countryWorldCacheTTL),
	}
	store.countryWorldCacheMu.Lock()
	defer store.countryWorldCacheMu.Unlock()
	if store.worldCasesRevision[serverID] != revision {
		return
	}
	if store.worldCasesExpansionCache == nil {
		store.worldCasesExpansionCache = make(map[string]worldCasesExpansionCacheEntry)
	}
	if _, exists := store.worldCasesExpansionCache[key]; !exists && len(store.worldCasesExpansionCache) >= worldCasesExpansionSize {
		oldestKey := ""
		var oldestExpiry time.Time
		for cachedKey, cachedEntry := range store.worldCasesExpansionCache {
			if oldestKey == "" || cachedEntry.expiresAt.Before(oldestExpiry) {
				oldestKey, oldestExpiry = cachedKey, cachedEntry.expiresAt
			}
		}
		delete(store.worldCasesExpansionCache, oldestKey)
	}
	store.worldCasesExpansionCache[key] = entry
}

func (store *PostgresStore) invalidateWorldForCases(serverID string) {
	scope := serverID + "\x00"
	store.countryWorldCacheMu.Lock()
	defer store.countryWorldCacheMu.Unlock()
	if store.worldCasesRevision == nil {
		store.worldCasesRevision = make(map[string]uint64)
	}
	store.worldCasesRevision[serverID]++
	for key := range store.worldCasesExpansionCache {
		if strings.HasPrefix(key, scope) {
			delete(store.worldCasesExpansionCache, key)
		}
	}
}

func (store *PostgresStore) cachedCountryExpansions(serverID, countryID, key string) ([]NPC, bool, uint64) {
	now := time.Now()
	store.countryWorldCacheMu.Lock()
	defer store.countryWorldCacheMu.Unlock()
	scope := countryExpansionScope(serverID, countryID)
	revision := store.countryExpansionRevision[scope]
	entry, exists := store.countryExpansionCache[key]
	if !exists || entry.revision != revision || !now.Before(entry.expiresAt) {
		if exists && !now.Before(entry.expiresAt) {
			delete(store.countryExpansionCache, key)
		}
		return nil, false, revision
	}
	return entry.npcs, true, revision
}

func (store *PostgresStore) cacheCountryExpansions(key, serverID, countryID string, revision uint64, npcs []NPC) {
	scope := countryExpansionScope(serverID, countryID)
	entry := countryExpansionCacheEntry{
		npcs: append([]NPC(nil), npcs...), revision: revision,
		expiresAt: time.Now().Add(countryWorldCacheTTL),
	}
	store.countryWorldCacheMu.Lock()
	defer store.countryWorldCacheMu.Unlock()
	if store.countryExpansionRevision[scope] != revision {
		return
	}
	if store.countryExpansionCache == nil {
		store.countryExpansionCache = make(map[string]countryExpansionCacheEntry)
	}
	if _, exists := store.countryExpansionCache[key]; !exists && len(store.countryExpansionCache) >= countryExpansionCacheSize {
		oldestKey := ""
		var oldestExpiry time.Time
		for cachedKey, cachedEntry := range store.countryExpansionCache {
			if oldestKey == "" || cachedEntry.expiresAt.Before(oldestExpiry) {
				oldestKey, oldestExpiry = cachedKey, cachedEntry.expiresAt
			}
		}
		delete(store.countryExpansionCache, oldestKey)
	}
	store.countryExpansionCache[key] = entry
}

func (store *PostgresStore) invalidateCountryExpansions(serverID, countryID string) {
	scope := countryExpansionScope(serverID, countryID)
	store.countryWorldCacheMu.Lock()
	defer store.countryWorldCacheMu.Unlock()
	if store.countryExpansionRevision == nil {
		store.countryExpansionRevision = make(map[string]uint64)
	}
	store.countryExpansionRevision[scope]++
	for key := range store.countryExpansionCache {
		if strings.HasPrefix(key, scope) {
			delete(store.countryExpansionCache, key)
		}
	}
	for key := range store.cityExpansionCache {
		if strings.HasPrefix(key, scope) {
			delete(store.cityExpansionCache, key)
		}
	}
}

func (store *PostgresStore) searchNPCs(ctx context.Context, serverID, cityID, query string) ([]PersonSearchRecord, error) {
	var base World
	var err error
	if cityID != "" {
		if cityID == "lyon" {
			cityID = "fr/lyon"
		}
		base, err = store.CityWorldForCases(ctx, serverID, cityID)
	} else {
		base, err = store.WorldForCases(ctx, serverID)
	}
	if err != nil {
		return nil, fmt.Errorf("reconstruct world for person search: %w", err)
	}
	return personSearchRecords(base, personSearchMatches(base.NPCs, cityID, query)), nil
}

func personSearchMatches(npcs []NPC, cityID, query string) []NPC {
	return personSearchMatchesWithAdditional(npcs, nil, cityID, query)
}

func personSearchMatchesWithAdditional(npcs, additional []NPC, cityID, query string) []NPC {
	query = strings.ToLower(strings.TrimSpace(query))
	matches := make([]NPC, 0, MaxNPCSearchResults)
	for _, group := range [][]NPC{npcs, additional} {
		for _, npc := range group {
			if cityID != "" && !npcBelongsToCity(cityID, npc.ID) {
				continue
			}
			if strings.Contains(strings.ToLower(npc.Name), query) {
				matches = append(matches, npc)
			}
		}
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
	return matches
}

func personSearchRecords(base World, matches []NPC) []PersonSearchRecord {
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
	return results
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

func (store *PostgresStore) countryCompanyIndex(ctx context.Context, serverID, countryID string) (countryCompanyCacheEntry, error) {
	cities := make([]WorldCity, 0, 4)
	cityIDs := make([]string, 0, 4)
	for _, city := range Cities() {
		if city.CountryID == countryID {
			cities = append(cities, city)
			cityIDs = append(cityIDs, city.ID)
		}
	}
	if len(cities) == 0 {
		return countryCompanyCacheEntry{}, nil
	}
	rows, err := store.db.Query(ctx,
		`SELECT w.world_seed, w.generator_version, COALESCE(g.city_id, ''), COALESCE(g.generator_version, 0)
		 FROM worlds w LEFT JOIN world_city_generators g
		 ON g.server_id=w.server_id AND g.city_id=ANY($2::text[])
		 WHERE w.server_id=$1::uuid ORDER BY g.city_id`, serverID, cityIDs)
	if err != nil {
		return countryCompanyCacheEntry{}, err
	}
	cityVersions := make(map[string]int, len(cityIDs))
	var seed string
	var worldVersion int
	for rows.Next() {
		var cityID string
		var version int
		if err := rows.Scan(&seed, &worldVersion, &cityID, &version); err != nil {
			rows.Close()
			return countryCompanyCacheEntry{}, err
		}
		if cityID != "" {
			cityVersions[cityID] = version
		}
	}
	rows.Close()
	if err := rows.Err(); err != nil {
		return countryCompanyCacheEntry{}, err
	}
	if seed == "" {
		return countryCompanyCacheEntry{}, ErrEntityNotFound
	}
	cacheKey := "companies:" + countryWorldCacheKey(seed, worldVersion, cities, cityVersions)
	store.countryWorldCacheMu.Lock()
	entry, cached := store.countryCompanyCache[cacheKey]
	if cached && time.Now().After(entry.expiresAt) {
		delete(store.countryCompanyCache, cacheKey)
		cached = false
	}
	store.countryWorldCacheMu.Unlock()
	store.observeCache("country_company_index", cached)
	companies := entry.companies
	if !cached {
		companies = make([]CompanyPortalRecord, 0, len(cities)*720)
		for _, city := range cities {
			version := countryCityGeneratorVersion(city, worldVersion, cityVersions)
			generated, err := GenerateCity(seed, version, city.ID)
			if err != nil {
				return countryCompanyCacheEntry{}, err
			}
			entities := make(map[string]Entity, len(generated.Entities))
			regions := make(map[string]string, len(generated.Regions))
			for _, entity := range generated.Entities {
				entities[entity.ID] = entity
			}
			for _, region := range generated.Regions {
				regions[region.ID] = region.Name
			}
			locations := make(map[string]string, len(generated.CompanyLocations))
			for _, relation := range generated.CompanyLocations {
				if relation.LocationRole == "headquarters" {
					locations[relation.CompanyID] = relation.LocationID
				}
			}
			for _, company := range generated.Companies {
				if locationID := locations[company.ID]; locationID != "" {
					record := CompanyPortalRecord{Company: company, Address: makeAddress(entities[locationID], regions)}
					record.WebsiteDomain, record.WebsiteDescription = companyWebsiteSearchMetadata(seed, worldVersion, record)
					companies = append(companies, record)
				}
			}
		}
		sort.Slice(companies, func(i, j int) bool { return companies[i].Company.ID < companies[j].Company.ID })
		for index := range companies {
			companies[index].SiteAddress = CompanySiteAddress(seed, countryID, index)
		}
		store.countryWorldCacheMu.Lock()
		if store.countryCompanyCache == nil {
			store.countryCompanyCache = make(map[string]countryCompanyCacheEntry)
		}
		if _, exists := store.countryCompanyCache[cacheKey]; !exists && len(store.countryCompanyCache) >= countryCompanyCacheSize {
			oldestKey := ""
			var oldestExpiry time.Time
			for cachedKey, cachedEntry := range store.countryCompanyCache {
				if oldestKey == "" || cachedEntry.expiresAt.Before(oldestExpiry) {
					oldestKey, oldestExpiry = cachedKey, cachedEntry.expiresAt
				}
			}
			delete(store.countryCompanyCache, oldestKey)
		}
		entry = countryCompanyCacheEntry{companies: companies, seed: seed, version: worldVersion, expiresAt: time.Now().Add(countryWorldCacheTTL)}
		store.countryCompanyCache[cacheKey] = entry
		store.countryWorldCacheMu.Unlock()
	}
	if cached {
		entry.seed = seed
		entry.version = worldVersion
	}
	return entry, nil
}

// SearchCompaniesInCountry returns lightweight company listings across a country's cities.
// Employee directories are deliberately loaded only by CompanyDirectoryAtLocation.
func (store *PostgresStore) SearchCompaniesInCountry(ctx context.Context, serverID, countryID, query string) ([]CompanyPortalRecord, error) {
	query = strings.ToLower(strings.TrimSpace(query))
	if query == "" {
		return []CompanyPortalRecord{}, nil
	}
	entry, err := store.countryCompanyIndex(ctx, serverID, countryID)
	if err != nil {
		return nil, err
	}
	if entry.companies == nil {
		return []CompanyPortalRecord{}, nil
	}
	companies := entry.companies
	results, exact := make([]CompanyPortalRecord, 0), make([]CompanyPortalRecord, 0)
	for _, record := range companies {
		searchable := record.Company.Name + " " + record.Address.Name + " " + record.Address.Full + " " + record.WebsiteDomain + " " + record.WebsiteDescription
		if strings.EqualFold(strings.TrimSpace(record.Company.Name), query) || strings.EqualFold(strings.TrimSpace(record.Address.Name), query) || strings.EqualFold(strings.TrimSpace(record.Address.Full), query) || strings.EqualFold(strings.TrimSpace(record.WebsiteDomain), query) {
			exact = append(exact, record)
		} else if strings.Contains(strings.ToLower(searchable), query) {
			results = append(results, record)
		}
	}
	if len(exact) > 0 {
		results = exact
	}
	sort.Slice(results, func(i, j int) bool {
		if results[i].Company.Name == results[j].Company.Name {
			return results[i].Company.ID < results[j].Company.ID
		}
		return results[i].Company.Name < results[j].Company.Name
	})
	if len(results) > 10 {
		results = results[:10]
	}
	return results, nil
}

// CompanySitesInCountry returns the lightweight seeded site index for a country's provider directory.
func (store *PostgresStore) CompanySitesInCountry(ctx context.Context, serverID, countryID string) ([]CompanyPortalRecord, error) {
	entry, err := store.countryCompanyIndex(ctx, serverID, countryID)
	if err != nil {
		return nil, err
	}
	companies := append([]CompanyPortalRecord(nil), entry.companies...)
	sort.Slice(companies, func(i, j int) bool {
		if companies[i].Company.Name == companies[j].Company.Name {
			return companies[i].Company.ID < companies[j].Company.ID
		}
		return companies[i].Company.Name < companies[j].Company.Name
	})
	return companies, nil
}

// CompanyWebsite returns the deterministic public site for one company in the active country.
func (store *PostgresStore) CompanyWebsite(ctx context.Context, serverID, countryID, companyID string) (CompanyWebsite, error) {
	entry, err := store.countryCompanyIndex(ctx, serverID, countryID)
	if err != nil {
		return CompanyWebsite{}, err
	}
	for _, company := range entry.companies {
		if company.Company.ID == companyID {
			site := GenerateCompanyWebsite(entry.seed, entry.version, company)
			site.SiteAddress = company.SiteAddress
			return site, nil
		}
	}
	return CompanyWebsite{}, ErrEntityNotFound
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
	if city, exists := CityForLocation(locationID); exists {
		entry, err := store.countryCompanyIndex(ctx, serverID, city.CountryID)
		if err != nil {
			return nil, err
		}
		for _, company := range entry.companies {
			if company.Address.LocationID != locationID || company.Company.ID != companyID {
				continue
			}
			store.observeCache("company_directory_index", true)
			cityVersion, err := store.cityGeneratorVersion(ctx, serverID, city.ID, entry.version)
			if err != nil {
				return nil, err
			}
			persistedEmployees, err := store.employeesOfCompany(ctx, serverID, companyID, entry.seed, cityVersion)
			if err != nil {
				return nil, err
			}
			generatedEmployees, err := generatedCompanyEmployees(entry.seed, cityVersion, company.Company)
			if err != nil {
				return nil, err
			}
			persistedByID := make(map[string]CompanyEmployee, len(persistedEmployees))
			for _, employee := range persistedEmployees {
				persistedByID[employee.ID] = employee
			}
			employees := make([]CompanyEmployee, 0, len(generatedEmployees)+len(persistedEmployees))
			for _, employee := range generatedEmployees {
				if persisted, ok := persistedByID[employee.ID]; ok {
					employees = append(employees, persisted)
					delete(persistedByID, employee.ID)
				} else {
					employees = append(employees, employee)
				}
			}
			for _, employee := range persistedEmployees {
				if _, alreadyIncluded := persistedByID[employee.ID]; alreadyIncluded {
					employees = append(employees, employee)
				}
			}
			if err := store.indexStaffEmailAddresses(ctx, serverID, employees); err != nil {
				return nil, fmt.Errorf("index company staff email addresses: %w", err)
			}
			return []CompanyRecord{{Company: company.Company, Address: company.Address, Employees: employees}}, nil
		}
		store.observeCache("company_directory_index", false)
	}
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
		if err := store.indexStaffEmailAddresses(ctx, serverID, employees); err != nil {
			return nil, fmt.Errorf("index company staff email addresses: %w", err)
		}
		return []CompanyRecord{record}, nil
	}
	return []CompanyRecord{}, nil
}

// CompanySiteAtLocation validates a site against the current country index without loading employee data.
func (store *PostgresStore) CompanySiteAtLocation(ctx context.Context, serverID, locationID, companyID string) ([]CompanyRecord, error) {
	if city, exists := CityForLocation(locationID); exists {
		entry, err := store.countryCompanyIndex(ctx, serverID, city.CountryID)
		if err != nil {
			return nil, err
		}
		for _, company := range entry.companies {
			if company.Address.LocationID == locationID && company.Company.ID == companyID {
				store.observeCache("company_site_index", true)
				return []CompanyRecord{{Company: company.Company, Address: company.Address}}, nil
			}
		}
		store.observeCache("company_site_index", false)
	}
	companies, err := store.CompanyAtLocation(ctx, serverID, locationID)
	if err != nil {
		return nil, err
	}
	for _, record := range companies {
		if record.Company.ID == companyID {
			return []CompanyRecord{record}, nil
		}
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

func generatedCityLocations(seed string, city WorldCity, districtID, districtName string, version int) []LocationRecord {
	streets := []string{"Rue de la Gare", "Market Street", "Foundry Lane", "Exchange Road", "Riverside Walk", "Station Street"}
	buildings := []string{"The Lantern Café", "Meridian Office", "Northstar Tailors", "Civic Exchange", "The Evening Ledger", "Bell & Finch Rooms"}
	stream := newStream(seed, GeneratorVersion5, "city/"+city.ID+"/district/"+districtID+"/locations")
	selectedStreets := shuffledCopy(streets, stream)[:2]
	selectedBuildings := shuffledCopy(buildings, stream)[:2]
	if version >= GeneratorVersion7 {
		slug := strings.TrimPrefix(districtID, city.ID+"/district/")
		selectedBuildings = make([]string, 0, 6)
		for _, building := range districtBuildingSet(seed, version, slug) {
			selectedBuildings = append(selectedBuildings, building.name)
		}
	}
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
		{"meridian-bank", "Union Mercantile Bank", ""},
		{"gate-station", "Central GATE Station", ""},
		{"hospital", city.Name + " Hospital", "open"},
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
	base, _, _, err := store.generatedWorldWithConfig(ctx, serverID)
	return base, err
}

func (store *PostgresStore) generatedWorldWithConfig(ctx context.Context, serverID string) (World, string, int, error) {
	var seed string
	var version int
	err := store.db.QueryRow(ctx,
		`SELECT world_seed, generator_version FROM worlds WHERE server_id = $1::uuid`, serverID,
	).Scan(&seed, &version)
	if errors.Is(err, pgx.ErrNoRows) {
		return World{}, "", 0, ErrEntityNotFound
	}
	if err != nil {
		return World{}, "", 0, err
	}
	base, err := Generate(seed, version)
	return base, seed, version, err
}

// GeneratedWorld reconstructs the small, deterministic world assigned to a server.
func (store *PostgresStore) GeneratedWorld(ctx context.Context, serverID string) (World, error) {
	return store.generatedWorld(ctx, serverID)
}

// WorldForCases reconstructs the seeded world plus NPCs added by on-demand growth.
func (store *PostgresStore) WorldForCases(ctx context.Context, serverID string) (World, error) {
	cityRows, err := store.db.Query(ctx,
		`SELECT w.world_seed, w.generator_version, COALESCE(c.city_id,''), COALESCE(c.generator_version,0)
		 FROM worlds w LEFT JOIN world_city_generators c ON c.server_id=w.server_id
		 WHERE w.server_id=$1::uuid ORDER BY c.city_id`, serverID,
	)
	if err != nil {
		return World{}, err
	}
	var seed string
	var version int
	cityVersions := make(map[string]int)
	if !cityRows.Next() {
		cityRows.Close()
		if err := cityRows.Err(); err != nil {
			return World{}, err
		}
		return World{}, ErrEntityNotFound
	}
	scanCityVersion := func() error {
		var cityID string
		var cityVersion int
		if err := cityRows.Scan(&seed, &version, &cityID, &cityVersion); err != nil {
			return err
		}
		if cityID != "" {
			cityVersions[cityID] = cityVersion
		}
		return nil
	}
	if err := scanCityVersion(); err != nil {
		cityRows.Close()
		return World{}, err
	}
	for cityRows.Next() {
		if err := scanCityVersion(); err != nil {
			cityRows.Close()
			return World{}, err
		}
	}
	cityRows.Close()
	if err := cityRows.Err(); err != nil {
		return World{}, err
	}
	cacheKey := worldForCasesCacheKey(seed, version, cityVersions)
	base, known, generatedCached := store.cachedWorldForCases(cacheKey)
	store.observeCache("world_for_cases_base", generatedCached)
	if !generatedCached {
		base, err = Generate(seed, version)
		if err != nil {
			return World{}, err
		}
		cityIDs := make([]string, 0, len(cityVersions))
		for cityID := range cityVersions {
			cityIDs = append(cityIDs, cityID)
		}
		sort.Strings(cityIDs)
		for _, cityID := range cityIDs {
			cityWorld, err := GenerateCity(seed, cityVersions[cityID], cityID)
			if err != nil {
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
		known = make(map[string]bool, len(base.NPCs))
		for _, npc := range base.NPCs {
			known[npc.ID] = true
		}
		store.cacheWorldForCases(cacheKey, base, known)
	}
	expansionKey := worldCasesScope(serverID) + cacheKey
	expandedNPCs, employmentRows, personalRows, expansionsCached, expansionRevision := store.cachedWorldCasesExpansions(serverID, expansionKey)
	store.observeCache("world_for_cases_expansions", expansionsCached)
	var expansionStatuses map[string]string
	if !expansionsCached {
		var entitiesJSON, employmentJSON, personalJSON, statusesJSON []byte
		err := store.db.QueryRow(ctx,
			`SELECT
			 COALESCE((SELECT jsonb_agg(jsonb_build_object('id', entity_id, 'key', generation_key) ORDER BY entity_id)
			           FROM world_entities WHERE server_id=$1::uuid AND entity_type='npc'), '[]'::jsonb),
			 COALESCE((SELECT jsonb_agg(jsonb_build_object('id', relationship_id, 'employeeNpcId', employee_npc_id,
			             'employerCompanyId', employer_company_id, 'role', COALESCE(role,'')) ORDER BY relationship_id)
			           FROM employment_relationships WHERE server_id=$1::uuid), '[]'::jsonb),
			 COALESCE((SELECT jsonb_agg(jsonb_build_object('id', relationship_id, 'personANpcId', person_a_npc_id,
			             'personBNpcId', person_b_npc_id, 'relationshipType', relationship_type) ORDER BY relationship_id)
			           FROM personal_relationships WHERE server_id=$1::uuid), '[]'::jsonb),
			 COALESCE((SELECT jsonb_agg(jsonb_build_object('id', entity_id, 'status', status) ORDER BY entity_id)
			           FROM (
			             SELECT DISTINCT ON (m.entity_id) m.entity_id, m.value #>> '{}' AS status
			             FROM world_mutations m JOIN world_entities e ON e.server_id=m.server_id AND e.entity_id=m.entity_id
			             WHERE m.server_id=$1::uuid AND m.field_key='status' AND e.entity_type='npc'
			             ORDER BY m.entity_id, m.mutation_order DESC
			           ) latest_statuses), '[]'::jsonb)`, serverID,
		).Scan(&entitiesJSON, &employmentJSON, &personalJSON, &statusesJSON)
		if err != nil {
			return World{}, err
		}
		var entityRows []persistedNPCGeneration
		var employmentRowsLoaded []EmploymentRelationship
		var personalRowsLoaded []PersonalRelationship
		var statusRows []persistedNPCStatus
		for _, result := range []struct {
			data   []byte
			target any
		}{{entitiesJSON, &entityRows}, {employmentJSON, &employmentRowsLoaded}, {personalJSON, &personalRowsLoaded}, {statusesJSON, &statusRows}} {
			if err := json.Unmarshal(result.data, result.target); err != nil {
				return World{}, fmt.Errorf("decode world expansion rows: %w", err)
			}
		}
		statusesLoaded := make(map[string]string, len(statusRows))
		for _, status := range statusRows {
			statusesLoaded[status.ID] = status.Status
		}
		expandedNPCs = make([]NPC, 0)
		for _, entity := range entityRows {
			if known[entity.ID] {
				continue
			}
			npc, err := GenerateNPCFromKey(seed, version, entity.Key)
			if err != nil {
				return World{}, err
			}
			if npc.ID != entity.ID {
				return World{}, ErrGeneratedConflict
			}
			expandedNPCs = append(expandedNPCs, npc)
		}
		employmentRows, personalRows = employmentRowsLoaded, personalRowsLoaded
		expansionStatuses = statusesLoaded
		store.cacheWorldCasesExpansions(serverID, expansionKey, expansionRevision, expandedNPCs, employmentRows, personalRows)
	}
	base.NPCs = append(base.NPCs, expandedNPCs...)
	if expansionsCached {
		if err := store.applyNPCStatuses(ctx, serverID, base.NPCs); err != nil {
			return World{}, err
		}
	} else if err := applyNPCStatusMap(base.NPCs, expansionStatuses); err != nil {
		return World{}, err
	}
	base.EmploymentRelationships = employmentRows
	base.PersonalRelationships = personalRows
	sort.Slice(base.NPCs, func(i, j int) bool { return base.NPCs[i].ID < base.NPCs[j].ID })
	return base, nil
}

// CityWorldForCases reconstructs one city's case data without rebuilding every
// city visited by the server or querying each NPC mutation separately.
func (store *PostgresStore) CityWorldForCases(ctx context.Context, serverID, cityID string) (World, error) {
	if _, ok := CityByID(cityID); !ok {
		return World{}, ErrEntityNotFound
	}
	var seed string
	var worldVersion, cityVersion int
	err := store.db.QueryRow(ctx,
		`SELECT w.world_seed, w.generator_version,
	        COALESCE(city_generator.generator_version,
	                 CASE WHEN $2::text = 'fr/lyon' OR w.generator_version >= $3::int
	                      THEN w.generator_version ELSE $3::int END)
		 FROM worlds w
		 LEFT JOIN world_city_generators city_generator
		   ON city_generator.server_id = w.server_id AND city_generator.city_id = $2
		 WHERE w.server_id = $1::uuid`,
		serverID, cityID, GeneratorVersion5,
	).Scan(&seed, &worldVersion, &cityVersion)
	if errors.Is(err, pgx.ErrNoRows) {
		return World{}, ErrEntityNotFound
	}
	if err != nil {
		return World{}, err
	}
	cityKey := cityWorldCacheKey(seed, worldVersion, cityVersion, cityID)
	base, districtIDs, known, generatedCached := store.cachedCityWorld(cityKey)
	store.observeCache("city_world_base", generatedCached)
	if !generatedCached {
		base, err = GenerateCity(seed, cityVersion, cityID)
		if err != nil {
			return World{}, err
		}
		known = make(map[string]bool, len(base.NPCs))
		districtIDs = make([]string, 0, len(base.Regions))
		for _, npc := range base.NPCs {
			known[npc.ID] = true
		}
		for _, region := range base.Regions {
			if region.Kind == "district" {
				districtIDs = append(districtIDs, region.ID)
			}
		}
		store.cacheCityWorld(cityKey, base, districtIDs, known)
	}
	if len(districtIDs) == 0 {
		return World{}, ErrEntityNotFound
	}
	city, _ := CityByID(cityID)
	cityExpansionKey := countryExpansionScope(serverID, city.CountryID) + "city:" + cityKey
	expandedNPCs, employmentRows, personalRows, expansionsCached, expansionRevision := store.cachedCityExpansions(serverID, city.CountryID, cityExpansionKey)
	store.observeCache("city_world_expansions", expansionsCached)
	if !expansionsCached {
		rows, err := store.db.Query(ctx,
			`SELECT entity_id, generation_key FROM world_entities
			 WHERE server_id=$1::uuid AND entity_type='npc' AND region_id=ANY($2::text[]) ORDER BY entity_id`,
			serverID, districtIDs,
		)
		if err != nil {
			return World{}, err
		}
		expandedNPCs = make([]NPC, 0)
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
			expandedNPCs = append(expandedNPCs, npc)
		}
		rows.Close()
		if err := rows.Err(); err != nil {
			return World{}, err
		}
	}
	base.NPCs = append(base.NPCs, expandedNPCs...)
	expandedNPCIDs := make([]string, 0, len(expandedNPCs))
	for _, npc := range expandedNPCs {
		expandedNPCIDs = append(expandedNPCIDs, npc.ID)
	}

	npcIDs := make([]string, 0, len(base.NPCs))
	for _, npc := range base.NPCs {
		npcIDs = append(npcIDs, npc.ID)
	}
	var statuses map[string]string
	var statusErr, employmentErr, personalErr error
	var reads sync.WaitGroup
	reads.Add(1)
	go func() {
		defer reads.Done()
		statuses, statusErr = store.cityNPCStatuses(ctx, serverID, npcIDs)
	}()
	if !expansionsCached {
		reads.Add(2)
		go func() {
			defer reads.Done()
			employmentRows, employmentErr = store.cityEmploymentRelationships(ctx, serverID, expandedNPCIDs)
		}()
		go func() {
			defer reads.Done()
			personalRows, personalErr = store.cityPersonalRelationships(ctx, serverID, expandedNPCIDs)
		}()
	}
	reads.Wait()
	if statusErr != nil {
		return World{}, statusErr
	}
	if employmentErr != nil {
		return World{}, employmentErr
	}
	if personalErr != nil {
		return World{}, personalErr
	}
	if !expansionsCached {
		store.cacheCityExpansions(cityExpansionKey, serverID, city.CountryID, expansionRevision, expandedNPCs, employmentRows, personalRows)
	}
	for index := range base.NPCs {
		if status, ok := statuses[base.NPCs[index].ID]; ok {
			if status != "alive" && status != "missing" && status != "dead" && status != "arrested" {
				return World{}, ErrInvalidMutation
			}
			base.NPCs[index].Status = status
		}
	}

	employmentIDs := make(map[string]bool, len(base.EmploymentRelationships))
	for _, relation := range base.EmploymentRelationships {
		employmentIDs[relation.ID] = true
	}
	for _, relation := range employmentRows {
		if !employmentIDs[relation.ID] {
			base.EmploymentRelationships = append(base.EmploymentRelationships, relation)
			employmentIDs[relation.ID] = true
		}
	}

	personalIDs := make(map[string]bool, len(base.PersonalRelationships))
	for _, relation := range base.PersonalRelationships {
		personalIDs[relation.ID] = true
	}
	for _, relation := range personalRows {
		if !personalIDs[relation.ID] {
			base.PersonalRelationships = append(base.PersonalRelationships, relation)
			personalIDs[relation.ID] = true
		}
	}
	sort.Slice(base.NPCs, func(i, j int) bool { return base.NPCs[i].ID < base.NPCs[j].ID })
	return base, nil
}

func (store *PostgresStore) cityNPCStatuses(ctx context.Context, serverID string, npcIDs []string) (map[string]string, error) {
	rows, err := store.db.Query(ctx,
		`SELECT DISTINCT ON (entity_id) entity_id, value #>> '{}'
		 FROM world_mutations WHERE server_id=$1::uuid AND field_key='status'
		 AND entity_id=ANY($2::text[]) ORDER BY entity_id, mutation_order DESC`, serverID, npcIDs,
	)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	statuses := make(map[string]string)
	for rows.Next() {
		var id, status string
		if err := rows.Scan(&id, &status); err != nil {
			return nil, err
		}
		statuses[id] = status
	}
	return statuses, rows.Err()
}

func (store *PostgresStore) cityEmploymentRelationships(ctx context.Context, serverID string, npcIDs []string) ([]EmploymentRelationship, error) {
	if len(npcIDs) == 0 {
		return []EmploymentRelationship{}, nil
	}
	rows, err := store.db.Query(ctx,
		`SELECT relationship_id, employee_npc_id, employer_company_id, COALESCE(role,'')
		 FROM employment_relationships WHERE server_id=$1::uuid AND employee_npc_id=ANY($2::text[])
		 ORDER BY relationship_id`, serverID, npcIDs,
	)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	relationships := make([]EmploymentRelationship, 0)
	for rows.Next() {
		var relation EmploymentRelationship
		if err := rows.Scan(&relation.ID, &relation.EmployeeNPCID, &relation.EmployerCompanyID, &relation.Role); err != nil {
			return nil, err
		}
		relationships = append(relationships, relation)
	}
	return relationships, rows.Err()
}

func (store *PostgresStore) cityPersonalRelationships(ctx context.Context, serverID string, npcIDs []string) ([]PersonalRelationship, error) {
	if len(npcIDs) == 0 {
		return []PersonalRelationship{}, nil
	}
	rows, err := store.db.Query(ctx,
		`SELECT relationship_id, person_a_npc_id, person_b_npc_id, relationship_type
		 FROM personal_relationships WHERE server_id=$1::uuid
		 AND (person_a_npc_id=ANY($2::text[]) OR person_b_npc_id=ANY($2::text[]))
		 ORDER BY relationship_id`, serverID, npcIDs,
	)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	relationships := make([]PersonalRelationship, 0)
	for rows.Next() {
		var relation PersonalRelationship
		if err := rows.Scan(&relation.ID, &relation.PersonANPCID, &relation.PersonBNPCID, &relation.RelationshipType); err != nil {
			return nil, err
		}
		relationships = append(relationships, relation)
	}
	return relationships, rows.Err()
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
	for _, codeUnit := range utf16.Encode([]rune(npcID)) {
		hash = (hash ^ uint32(codeUnit)) * 16777619
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
	countryID := ""
	if city, _, _, ok := cityForDistrict(seed, districtID); ok {
		countryID = city.CountryID
		version, err = store.cityGeneratorVersion(ctx, serverID, city.ID, version)
		if err != nil {
			return err
		}
	}
	if version >= GeneratorVersion7 {
		return nil
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
	if err := tx.Commit(ctx); err != nil {
		return err
	}
	if countryID != "" {
		store.invalidateCountryExpansions(serverID, countryID)
	}
	store.invalidateWorldForCases(serverID)
	return nil
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
	if status != "alive" && status != "missing" && status != "dead" && status != "arrested" {
		return ErrInvalidMutation
	}
	npc.Status = status
	return nil
}

func applyNPCStatusMap(npcs []NPC, statuses map[string]string) error {
	for index := range npcs {
		status, exists := statuses[npcs[index].ID]
		if !exists {
			continue
		}
		if status != "alive" && status != "missing" && status != "dead" && status != "arrested" {
			return ErrInvalidMutation
		}
		npcs[index].Status = status
	}
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
		if status != "alive" && status != "missing" && status != "dead" && status != "arrested" {
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
	return store.employeesOfCompany(ctx, serverID, companyID, seed, version)
}

// employeesOfCompany reuses the versioned country index config when available.
func (store *PostgresStore) employeesOfCompany(ctx context.Context, serverID, companyID, seed string, version int) ([]CompanyEmployee, error) {
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
		employees = append(employees, CompanyEmployee{ID: employeeID, Name: npc.Name, Email: EmailAddressForNPC(employeeID, npc.Name), Role: role, Status: status, Visual: npc.Visual})
	}
	if err := rows.Err(); err != nil {
		return nil, err
	}
	return employees, nil
}

// EmployeeByEmail resolves only addresses published in company staff directories.
func (store *PostgresStore) EmployeeByEmail(ctx context.Context, serverID, address string) (CompanyEmployee, error) {
	var employee CompanyEmployee
	err := store.db.QueryRow(ctx, `SELECT npc_id,name,email,role,status FROM world_staff_email_directory
	 WHERE server_id=$1::uuid AND email=$2`, serverID, strings.ToLower(strings.TrimSpace(address))).
		Scan(&employee.ID, &employee.Name, &employee.Email, &employee.Role, &employee.Status)
	if errors.Is(err, pgx.ErrNoRows) {
		return CompanyEmployee{}, ErrEntityNotFound
	}
	return employee, err
}

func (store *PostgresStore) indexStaffEmailAddresses(ctx context.Context, serverID string, employees []CompanyEmployee) error {
	if len(employees) == 0 {
		return nil
	}
	values := make([]string, 0, len(employees))
	args := make([]any, 0, 6*len(employees))
	for _, employee := range employees {
		if employee.ID == "" || employee.Email == "" || employee.Name == "" {
			continue
		}
		first := len(args) + 1
		values = append(values, fmt.Sprintf("($%d::uuid,$%d,$%d,$%d,$%d,$%d)", first, first+1, first+2, first+3, first+4, first+5))
		args = append(args, serverID, strings.ToLower(employee.Email), employee.ID, employee.Name, employee.Role, employee.Status)
		first = len(args) + 1
		values = append(values, fmt.Sprintf("($%d::uuid,$%d,$%d,$%d,$%d,$%d)", first, first+1, first+2, first+3, first+4, first+5))
		args = append(args, serverID, strings.ToLower(legacyEmailAddressForNPC(employee.ID, employee.Name)), employee.ID, employee.Name, employee.Role, employee.Status)
	}
	if len(values) == 0 {
		return nil
	}
	_, err := store.db.Exec(ctx, `INSERT INTO world_staff_email_directory(server_id,email,npc_id,name,role,status) VALUES `+strings.Join(values, ",")+
		` ON CONFLICT(server_id,email) DO UPDATE SET npc_id=EXCLUDED.npc_id,name=EXCLUDED.name,role=EXCLUDED.role,status=EXCLUDED.status`, args...)
	return err
}

// EmployeeByNPCID validates the result selected from an email-address search.
func (store *PostgresStore) EmployeeByNPCID(ctx context.Context, serverID, npcID, address string) (CompanyEmployee, error) {
	if !strings.HasPrefix(npcID, "npc:") {
		return CompanyEmployee{}, ErrEntityNotFound
	}
	seed, worldVersion, err := store.WorldConfig(ctx, serverID)
	if err != nil {
		return CompanyEmployee{}, err
	}
	var key, role, status string
	err = store.db.QueryRow(ctx, `SELECT e.generation_key,r.role,
	 COALESCE((SELECT value #>> '{}' FROM world_mutations m WHERE m.server_id=e.server_id AND m.entity_id=e.entity_id AND m.field_key='status' ORDER BY m.mutation_order DESC LIMIT 1),'alive')
	 FROM world_entities e JOIN employment_relationships r ON r.server_id=e.server_id AND r.employee_npc_id=e.entity_id
	 WHERE e.server_id=$1::uuid AND e.entity_id=$2 ORDER BY r.relationship_id LIMIT 1`, serverID, npcID).
		Scan(&key, &role, &status)
	if err == nil {
		npc, err := GenerateNPCFromKey(seed, worldVersion, key)
		if err != nil {
			return CompanyEmployee{}, err
		}
		if !employeeEmailMatches(npcID, npc.Name, address) {
			return CompanyEmployee{}, ErrEntityNotFound
		}
		return CompanyEmployee{ID: npcID, Name: npc.Name, Email: EmailAddressForNPC(npcID, npc.Name), Role: role, Status: status, Visual: npc.Visual}, nil
	}
	if !errors.Is(err, pgx.ErrNoRows) {
		return CompanyEmployee{}, err
	}
	key = strings.TrimPrefix(npcID, "npc:")
	parts := strings.Split(key, "/")
	if len(parts) != 10 || parts[2] != "people" || parts[4] != "district" || parts[6] != "company" || parts[8] != "person" {
		return CompanyEmployee{}, ErrEntityNotFound
	}
	cityID := parts[0] + "/" + parts[1]
	city, ok := CityByID(cityID)
	if !ok {
		return CompanyEmployee{}, ErrEntityNotFound
	}
	companyIndex, err := strconv.Atoi(parts[7])
	if err != nil || companyIndex < 1 || companyIndex > 36 {
		return CompanyEmployee{}, ErrEntityNotFound
	}
	slot, err := strconv.Atoi(parts[9])
	if err != nil || slot < 1 {
		return CompanyEmployee{}, ErrEntityNotFound
	}
	citySeed := CitySeed(seed, city.ID)
	generationVersion := 0
	for candidate := LatestGeneratorVersion; candidate >= GeneratorVersion7; candidate-- {
		if parts[3] == seedFingerprint(citySeed, candidate) {
			generationVersion = candidate
			break
		}
	}
	if generationVersion == 0 || slot > employeeCount(citySeed, generationVersion, companyIndex) {
		return CompanyEmployee{}, ErrEntityNotFound
	}
	npc, err := GenerateNPC(seed, generationVersion, key)
	if err != nil {
		return CompanyEmployee{}, err
	}
	if !employeeEmailMatches(npcID, npc.Name, address) {
		return CompanyEmployee{}, ErrEntityNotFound
	}
	return CompanyEmployee{ID: npcID, Name: npc.Name, Email: EmailAddressForNPC(npcID, npc.Name), Role: employmentRoles[(slot-1)%len(employmentRoles)], Status: npc.Status, Visual: npc.Visual}, nil
}

func employeeEmailMatches(id, name, address string) bool {
	address = strings.TrimSpace(address)
	return strings.EqualFold(EmailAddressForNPC(id, name), address) || strings.EqualFold(legacyEmailAddressForNPC(id, name), address)
}

// EmployeesOfCompanies loads staff for several employers in one query, reusing the caller's world config.
func (store *PostgresStore) EmployeesOfCompanies(ctx context.Context, serverID, seed string, version int, companyIDs []string) (map[string][]CompanyEmployee, error) {
	result := make(map[string][]CompanyEmployee, len(companyIDs))
	if len(companyIDs) == 0 {
		return result, nil
	}
	rows, err := store.db.Query(ctx,
		`SELECT r.employer_company_id, e.entity_id, e.generation_key, r.role,
		 COALESCE(status.value #>> '{}', 'alive'), status.created_at
		 FROM employment_relationships r
		 JOIN world_entities e
		   ON e.server_id = r.server_id AND e.entity_id = r.employee_npc_id AND e.entity_type = 'npc'
		 LEFT JOIN LATERAL (
			SELECT value, created_at FROM world_mutations m
		   WHERE m.server_id = e.server_id AND m.entity_id = e.entity_id AND m.field_key = 'status'
		   ORDER BY m.mutation_order DESC LIMIT 1
		 ) status ON true
		 WHERE r.server_id = $1::uuid AND r.employer_company_id = ANY($2::text[])
		 ORDER BY r.employer_company_id, e.entity_id`,
		serverID, companyIDs,
	)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	for rows.Next() {
		var companyID, employeeID, key, role, status string
		var statusChangedAt pgtype.Timestamptz
		if err := rows.Scan(&companyID, &employeeID, &key, &role, &status, &statusChangedAt); err != nil {
			return nil, err
		}
		npc, err := GenerateNPC(seed, version, key)
		if err != nil {
			return nil, err
		}
		employee := CompanyEmployee{ID: employeeID, Name: npc.Name, Email: EmailAddressForNPC(employeeID, npc.Name), Role: role, Status: status, Visual: npc.Visual}
		if statusChangedAt.Valid {
			employee.StatusChangedAt = statusChangedAt.Time
		}
		result[companyID] = append(result[companyID], employee)
	}
	if err := rows.Err(); err != nil {
		return nil, err
	}
	return result, nil
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

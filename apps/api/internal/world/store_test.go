package world

import (
	"context"
	"encoding/json"
	"errors"
	"strings"
	"testing"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgconn"
)

type fixtureRow struct {
	values []any
	err    error
}

func (row fixtureRow) Scan(destinations ...any) error {
	if row.err != nil {
		return row.err
	}
	if len(destinations) != len(row.values) {
		return errors.New("fixture row has unexpected column count")
	}
	for index, value := range row.values {
		switch destination := destinations[index].(type) {
		case *string:
			*destination = value.(string)
		case *int:
			*destination = value.(int)
		default:
			return errors.New("fixture row has an unsupported destination type")
		}
	}
	return nil
}

type fixtureRows struct {
	rows    [][]any
	current []any
	index   int
}

func (rows *fixtureRows) Close()                                       {}
func (rows *fixtureRows) Err() error                                   { return nil }
func (rows *fixtureRows) CommandTag() pgconn.CommandTag                { return pgconn.NewCommandTag("SELECT") }
func (rows *fixtureRows) FieldDescriptions() []pgconn.FieldDescription { return nil }
func (rows *fixtureRows) RawValues() [][]byte                          { return nil }
func (rows *fixtureRows) Conn() *pgx.Conn                              { return nil }
func (rows *fixtureRows) Values() ([]any, error)                       { return rows.current, nil }
func (rows *fixtureRows) Next() bool {
	if rows.index >= len(rows.rows) {
		return false
	}
	rows.current = rows.rows[rows.index]
	rows.index++
	return true
}
func (rows *fixtureRows) Scan(destinations ...any) error {
	return fixtureRow{values: rows.current}.Scan(destinations...)
}

type fixtureDatabase struct {
	queryRow func(string, []any) pgx.Row
	query    func(string, []any) pgx.Rows
	queries  []string
}

type durableMutationDatabase struct {
	seed       string
	version    int
	npcID      string
	generation string
	mutations  map[string][]string
}

func (database *durableMutationDatabase) Exec(_ context.Context, query string, args ...any) (pgconn.CommandTag, error) {
	if !strings.Contains(query, "INSERT INTO world_mutations") {
		return pgconn.NewCommandTag("INSERT 0 1"), nil
	}
	serverID, npcID, rawStatus := args[0].(string), args[1].(string), args[2].(string)
	if npcID != database.npcID || (serverID != "server-a" && serverID != "server-b") {
		return pgconn.NewCommandTag("INSERT 0 0"), nil
	}
	var status string
	if err := json.Unmarshal([]byte(rawStatus), &status); err != nil {
		return pgconn.CommandTag{}, err
	}
	database.mutations[serverID] = append(database.mutations[serverID], status)
	return pgconn.NewCommandTag("INSERT 0 1"), nil
}

func (database *durableMutationDatabase) QueryRow(_ context.Context, query string, args ...any) pgx.Row {
	if strings.Contains(query, "SELECT w.world_seed") {
		if args[1] != database.npcID || (args[0] != "server-a" && args[0] != "server-b") {
			return fixtureRow{err: pgx.ErrNoRows}
		}
		return fixtureRow{values: []any{database.seed, database.version, database.generation}}
	}
	if strings.Contains(query, "SELECT value #>> '{}' FROM world_mutations") {
		serverID := args[0].(string)
		values := database.mutations[serverID]
		if len(values) == 0 {
			return fixtureRow{err: pgx.ErrNoRows}
		}
		return fixtureRow{values: []any{values[len(values)-1]}}
	}
	return fixtureRow{err: errors.New("unexpected query row")}
}

func (database *durableMutationDatabase) Query(context.Context, string, ...any) (pgx.Rows, error) {
	return &fixtureRows{}, nil
}

func (database *durableMutationDatabase) Begin(context.Context) (pgx.Tx, error) {
	return nil, errors.New("transaction is not used by this fixture")
}

func (database *fixtureDatabase) Exec(context.Context, string, ...any) (pgconn.CommandTag, error) {
	return pgconn.NewCommandTag("INSERT 0 1"), nil
}
func (database *fixtureDatabase) Query(_ context.Context, query string, args ...any) (pgx.Rows, error) {
	database.queries = append(database.queries, query)
	return database.query(query, args), nil
}
func (database *fixtureDatabase) QueryRow(_ context.Context, query string, args ...any) pgx.Row {
	database.queries = append(database.queries, query)
	return database.queryRow(query, args)
}
func (database *fixtureDatabase) Begin(context.Context) (pgx.Tx, error) {
	return nil, errors.New("transaction is not used by this fixture")
}

func TestMutationSurvivesStoreRestartAndIsIsolatedByServer(t *testing.T) {
	const seed = "m04-restart-fixture"
	generated, err := Generate(seed, GeneratorVersion2)
	if err != nil {
		t.Fatalf("Generate() error = %v", err)
	}
	npc := generated.NPCs[0]
	database := &durableMutationDatabase{
		seed:       seed,
		version:    GeneratorVersion2,
		npcID:      npc.ID,
		generation: npc.GenerationKey,
		mutations:  make(map[string][]string),
	}

	firstProcess := NewPostgresStore(database)
	if err := firstProcess.SetNPCStatus(context.Background(), "server-a", npc.ID, "dead"); err != nil {
		t.Fatalf("SetNPCStatus() error = %v", err)
	}

	// A new store instance represents a process restart; only the database is shared.
	restartedProcess := NewPostgresStore(database)
	rebuilt, err := restartedProcess.GetNPC(context.Background(), "server-a", npc.ID)
	if err != nil {
		t.Fatalf("GetNPC() after restart error = %v", err)
	}
	if rebuilt.NPC.Status != "dead" {
		t.Fatalf("status after restart = %q, want dead", rebuilt.NPC.Status)
	}

	otherServer, err := restartedProcess.GetNPC(context.Background(), "server-b", npc.ID)
	if err != nil {
		t.Fatalf("GetNPC() for other server error = %v", err)
	}
	if otherServer.NPC.Status != "alive" {
		t.Fatalf("other server status = %q, want alive", otherServer.NPC.Status)
	}
	if err := restartedProcess.SetNPCStatus(context.Background(), "server-a", npc.ID, "missing"); !errors.Is(err, ErrInvalidMutation) {
		t.Fatalf("invalid status error = %v, want ErrInvalidMutation", err)
	}
}

func TestPostgresStoreReconstructsNPCProfilesAndIndexedRelations(t *testing.T) {
	const seed = "m03-store-fixture"
	generated, err := Generate(seed, GeneratorVersion2)
	if err != nil {
		t.Fatalf("Generate() error = %v", err)
	}
	firstNPC := generated.NPCs[0]
	var marriage PersonalRelationship
	for _, relationship := range generated.PersonalRelationships {
		if relationship.PersonANPCID == firstNPC.ID || relationship.PersonBNPCID == firstNPC.ID {
			marriage = relationship
			break
		}
	}
	partnerID := marriage.PersonANPCID
	if partnerID == firstNPC.ID {
		partnerID = marriage.PersonBNPCID
	}
	var partner NPC
	for _, npc := range generated.NPCs {
		if npc.ID == partnerID {
			partner = npc
			break
		}
	}
	var employment EmploymentRelationship
	for _, relationship := range generated.EmploymentRelationships {
		if relationship.EmployeeNPCID == firstNPC.ID {
			employment = relationship
			break
		}
	}
	var employer Company
	for _, company := range generated.Companies {
		if company.ID == employment.EmployerCompanyID {
			employer = company
			break
		}
	}
	entityByID := make(map[string]Entity, len(generated.Entities))
	for _, entity := range generated.Entities {
		entityByID[entity.ID] = entity
	}

	database := &fixtureDatabase{}
	database.queryRow = func(query string, args []any) pgx.Row {
		switch {
		case strings.Contains(query, "JOIN worlds w"):
			npcID := args[1].(string)
			for _, npc := range generated.NPCs {
				if npc.ID == npcID {
					return fixtureRow{values: []any{seed, GeneratorVersion2, npc.GenerationKey}}
				}
			}
		case strings.Contains(query, "FROM worlds WHERE"):
			return fixtureRow{values: []any{seed, GeneratorVersion2}}
		case strings.Contains(query, "entity_type = 'company'"):
			if employer.ID == args[1].(string) {
				return fixtureRow{values: []any{employer.GenerationKey}}
			}
		}
		return fixtureRow{err: pgx.ErrNoRows}
	}
	database.query = func(query string, args []any) pgx.Rows {
		if strings.Contains(query, "SELECT r.relationship_id, r.role") {
			npcID := args[1].(string)
			for _, relationship := range generated.EmploymentRelationships {
				if relationship.EmployeeNPCID == npcID {
					company := employer
					if company.ID != relationship.EmployerCompanyID {
						for _, candidate := range generated.Companies {
							if candidate.ID == relationship.EmployerCompanyID {
								company = candidate
							}
						}
					}
					return &fixtureRows{rows: [][]any{{relationship.ID, relationship.Role, company.ID, company.GenerationKey}}}
				}
			}
		}
		if strings.Contains(query, "SELECT r.relationship_type, other.entity_id") {
			npcID := args[1].(string)
			for _, relationship := range generated.PersonalRelationships {
				otherID := relationship.PersonANPCID
				if otherID == npcID {
					otherID = relationship.PersonBNPCID
				} else if relationship.PersonBNPCID != npcID {
					continue
				}
				other := entityByID[otherID]
				return &fixtureRows{rows: [][]any{{relationship.RelationshipType, other.ID, other.GenerationKey}}}
			}
		}
		if strings.Contains(query, "SELECT e.entity_id, e.generation_key, r.role") {
			companyID := args[1].(string)
			rows := &fixtureRows{}
			for _, relationship := range generated.EmploymentRelationships {
				if relationship.EmployerCompanyID == companyID {
					rows.rows = append(rows.rows, []any{relationship.EmployeeNPCID, entityByID[relationship.EmployeeNPCID].GenerationKey, relationship.Role})
				}
			}
			return rows
		}
		return &fixtureRows{}
	}

	store := NewPostgresStore(database)
	profile, err := store.GetNPC(context.Background(), "server-id", firstNPC.ID)
	if err != nil {
		t.Fatalf("GetNPC() error = %v", err)
	}
	if profile.NPC != firstNPC || len(profile.Employment) != 1 || profile.Employment[0].Company != employer || len(profile.PersonalRelationships) != 1 || profile.PersonalRelationships[0].NPC != partner {
		t.Fatalf("GetNPC() profile did not include generated attributes and relations: %+v", profile)
	}
	reverseProfile, err := store.GetNPC(context.Background(), "server-id", partner.ID)
	if err != nil {
		t.Fatalf("GetNPC() reverse relationship error = %v", err)
	}
	if len(reverseProfile.PersonalRelationships) != 1 || reverseProfile.PersonalRelationships[0].NPC != firstNPC {
		t.Fatalf("reverse personal relationship lookup failed: %+v", reverseProfile.PersonalRelationships)
	}
	for _, query := range database.queries {
		if strings.Contains(query, "personal_relationships r") && !strings.Contains(query, "r.person_a_npc_id = $2 OR r.person_b_npc_id = $2") {
			t.Fatal("personal relation query does not look up both relationship directions")
		}
	}

	employees, err := store.EmployeesOfCompany(context.Background(), "server-id", employer.ID)
	if err != nil {
		t.Fatalf("EmployeesOfCompany() error = %v", err)
	}
	if len(employees) != 8 {
		t.Fatalf("employee count = %d, want 8", len(employees))
	}
	if !strings.Contains(database.queries[len(database.queries)-1], "r.employer_company_id = $2") {
		t.Fatal("employee query is not filtered by company relationship")
	}
}

package cases

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"regexp"
	"strconv"
	"strings"
	"testing"
	"time"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgconn"
	"github.com/jackc/pgx/v5/pgtype"
	"noir-1974/apps/api/internal/records"
)

type testCaseRow struct {
	values []any
	err    error
}

type emptyCaseRows struct{ pgx.Rows }

func (*emptyCaseRows) Close()                                       {}
func (*emptyCaseRows) Err() error                                   { return nil }
func (*emptyCaseRows) CommandTag() pgconn.CommandTag                { return pgconn.CommandTag{} }
func (*emptyCaseRows) FieldDescriptions() []pgconn.FieldDescription { return nil }
func (*emptyCaseRows) Next() bool                                   { return false }
func (*emptyCaseRows) Scan(...any) error                            { return nil }
func (*emptyCaseRows) Values() ([]any, error)                       { return nil, nil }
func (*emptyCaseRows) RawValues() [][]byte                          { return nil }

type crimeReportQueryFixture struct{ queries int }

func (fixture *crimeReportQueryFixture) Query(_ context.Context, query string, args ...any) (pgx.Rows, error) {
	maxPlaceholder := 0
	for _, match := range regexp.MustCompile(`\$(\d+)`).FindAllStringSubmatch(query, -1) {
		index, err := strconv.Atoi(match[1])
		if err != nil {
			return nil, err
		}
		if index > maxPlaceholder {
			maxPlaceholder = index
		}
	}
	if maxPlaceholder != len(args) {
		return nil, fmt.Errorf("query uses $%d but received %d arguments", maxPlaceholder, len(args))
	}
	if strings.Contains(query, "v.display_name") || strings.Contains(query, "scene.display_name") {
		return nil, errors.New("world_entities does not contain display_name")
	}
	fixture.queries++
	return &emptyCaseRows{}, nil
}

func (*crimeReportQueryFixture) Exec(context.Context, string, ...any) (pgconn.CommandTag, error) {
	return pgconn.CommandTag{}, errors.New("unexpected crime report exec")
}

func (*crimeReportQueryFixture) QueryRow(context.Context, string, ...any) pgx.Row {
	return testCaseRow{err: errors.New("unexpected crime report query row")}
}

func (*crimeReportQueryFixture) Begin(context.Context) (pgx.Tx, error) {
	return nil, errors.New("unexpected crime report transaction")
}

func TestCityCrimeReportsBindsEverySQLParameter(t *testing.T) {
	database := &crimeReportQueryFixture{}
	store := NewPostgresStore(database)
	now := time.Date(1974, time.October, 27, 12, 0, 0, 0, time.UTC)
	if _, err := store.CityCrimeReports(context.Background(), "server-1", "fr/lyon", now.AddDate(0, 0, -30), now.Add(time.Nanosecond), now); err != nil {
		t.Fatalf("CityCrimeReports() error = %v", err)
	}
	if _, err := store.ListPublicBodyDiscoveries(context.Background(), "server-1", "fr/lyon", time.Now().UTC()); err != nil {
		t.Fatalf("ListPublicBodyDiscoveries() error = %v", err)
	}
	if database.queries != 3 {
		t.Fatalf("crime report methods executed %d queries, want 3", database.queries)
	}
}

func TestSplitSQLStatementsKeepsDollarQuotedFunctionBodies(t *testing.T) {
	sql := `CREATE FUNCTION guard() RETURNS trigger AS $$ BEGIN PERFORM 1; RETURN NEW; END; $$ LANGUAGE plpgsql; SELECT 1;`
	statements := splitSQLStatements(sql)
	if len(statements) != 2 {
		t.Fatalf("splitSQLStatements returned %d statements, want 2: %#v", len(statements), statements)
	}
	if !strings.Contains(statements[0], "PERFORM 1; RETURN NEW;") {
		t.Fatalf("function body was split at an internal semicolon: %q", statements[0])
	}
}

func (row testCaseRow) Scan(destinations ...any) error {
	if row.err != nil {
		return row.err
	}
	if len(destinations) != len(row.values) {
		return errors.New("test row has unexpected column count")
	}
	for index, destination := range destinations {
		switch target := destination.(type) {
		case *string:
			*target = row.values[index].(string)
		case *int:
			*target = row.values[index].(int)
		case *[]byte:
			*target = append([]byte(nil), row.values[index].([]byte)...)
		case *pgtype.Text:
			if row.values[index] != nil {
				*target = pgtype.Text{String: row.values[index].(string), Valid: true}
			}
		case *pgtype.Timestamptz:
			if row.values[index] != nil {
				*target = pgtype.Timestamptz{Time: row.values[index].(time.Time), Valid: true}
			}
		default:
			return errors.New("test row has unsupported destination")
		}
	}
	return nil
}

type casePersistenceFixture struct {
	sharedState   map[string]string
	progress      map[string][]byte
	evidence      map[string]map[string]bool
	activityRows  [][]any
	emailRows     [][]any
	activityQuery string
	activityArgs  []any
	sessionQuery  string
	sessionArgs   []any
	sessionRows   [][]any
}

func newCasePersistenceFixture() *casePersistenceFixture {
	return &casePersistenceFixture{
		sharedState: make(map[string]string),
		progress:    make(map[string][]byte),
		evidence:    make(map[string]map[string]bool),
	}
}

func (fixture *casePersistenceFixture) Exec(_ context.Context, query string, args ...any) (pgconn.CommandTag, error) {
	switch {
	case strings.Contains(query, "UPDATE case_states"):
		key := args[0].(string) + "/" + args[1].(string)
		if _, exists := fixture.sharedState[key]; !exists {
			return pgconn.NewCommandTag("UPDATE 0"), nil
		}
		fixture.sharedState[key] = args[2].(string)
		return pgconn.NewCommandTag("UPDATE 1"), nil
	case strings.Contains(query, "INSERT INTO player_case_progress"):
		key := args[0].(string) + "/" + args[1].(string) + "/" + args[2].(string)
		fixture.progress[key] = []byte(args[3].(string))
		return pgconn.NewCommandTag("INSERT 1"), nil
	default:
		return pgconn.CommandTag{}, errors.New("unexpected case exec")
	}
}

func (fixture *casePersistenceFixture) QueryRow(_ context.Context, query string, args ...any) pgx.Row {
	switch {
	case strings.Contains(query, "SELECT status FROM case_states"):
		key := args[0].(string) + "/" + args[1].(string)
		status, exists := fixture.sharedState[key]
		if !exists {
			return testCaseRow{err: pgx.ErrNoRows}
		}
		return testCaseRow{values: []any{status}}
	case strings.Contains(query, "SELECT COUNT(*) FROM ("):
		key := args[0].(string) + "/" + args[1].(string)
		count := 0
		for _, evidenceID := range args[2].([]string) {
			if fixture.evidence[key][evidenceID] {
				count++
			}
		}
		return testCaseRow{values: []any{count}}
	case strings.Contains(query, "SELECT progress FROM player_case_progress"):
		key := args[0].(string) + "/" + args[1].(string) + "/" + args[2].(string)
		progress, exists := fixture.progress[key]
		if !exists {
			return testCaseRow{err: pgx.ErrNoRows}
		}
		return testCaseRow{values: []any{progress}}
	default:
		return testCaseRow{err: errors.New("unexpected case query row")}
	}
}

func (fixture *casePersistenceFixture) Query(_ context.Context, query string, args ...any) (pgx.Rows, error) {
	if strings.Contains(query, "FROM player_sessions s") {
		fixture.sessionQuery = query
		fixture.sessionArgs = args
		return &caseFixtureRows{rows: fixture.sessionRows}, nil
	}
	if strings.Contains(query, "FROM case_activity_records") {
		fixture.activityQuery = query
		fixture.activityArgs = args
		return &caseFixtureRows{rows: fixture.activityRows}, nil
	}
	if strings.Contains(query, "FROM case_mail_messages") {
		return &caseFixtureRows{rows: fixture.emailRows}, nil
	}
	return nil, errors.New("unexpected case row query")
}

func TestAnonymousCaseEmailHidesSenderOnlyInRecipientMailbox(t *testing.T) {
	data, err := json.Marshal(map[string]string{"sender": "Ada Morel", "recipient": "Mara Vale", "subject": "A private warning", "body": "Keep the register closed."})
	if err != nil {
		t.Fatal(err)
	}
	fixture := newCasePersistenceFixture()
	fixture.emailRows = [][]any{{"case:one", "npc:victim", "npc:killer", "anonymous_warning", data}}
	store := NewPostgresStore(fixture)

	received, err := store.NPCEmails(context.Background(), "server-1", "npc:victim", time.Time{}, time.Time{})
	if err != nil {
		t.Fatalf("recipient NPCEmails() error = %v", err)
	}
	if len(received) != 1 || received[0].Sender != "Unknown sender" {
		t.Fatalf("recipient view should hide sender: %+v", received)
	}

	sent, err := store.NPCEmails(context.Background(), "server-1", "npc:killer", time.Time{}, time.Time{})
	if err != nil {
		t.Fatalf("sender NPCEmails() error = %v", err)
	}
	if len(sent) != 1 || sent[0].Sender != "Ada Morel" {
		t.Fatalf("sender view should retain the sender identity: %+v", sent)
	}
}

func TestPersistedCaseIdentityAllowsRegeneratedCrimeTime(t *testing.T) {
	identity := persistedCaseIdentity{
		Seed: "seed", CityID: "fr/lyon", GeneratorVersion: GeneratorVersion, CaseType: "homicide",
		ArchetypeID: "jealousy", Motive: "motive", VictimID: "npc:victim", KillerID: "npc:killer", LocationID: "location:scene",
	}
	definition := Definition{
		Seed: "seed", CityID: "fr/lyon", GeneratorVersion: GeneratorVersion, CaseType: "homicide",
		ArchetypeID: "jealousy", Motive: "motive", VictimID: "npc:victim", KillerID: "npc:killer", LocationID: "location:scene",
		CrimeAt: time.Date(1974, time.October, 27, 19, 30, 0, 0, time.UTC),
	}
	if !identity.matches(definition) {
		t.Fatal("same persisted case identity should remain idempotent when crime time is regenerated")
	}
	definition.CrimeAt = definition.CrimeAt.AddDate(0, 0, 1)
	if !identity.matches(definition) {
		t.Fatal("crime time should not change the persisted case identity")
	}
	definition.LocationID = "location:other-scene"
	if identity.matches(definition) {
		t.Fatal("changing the persisted scene should still produce a case conflict")
	}
}

func TestListOpenPublicCasesForSessionUsesOneScopedQuery(t *testing.T) {
	createdAt := time.Date(1974, time.October, 24, 12, 0, 0, 0, time.UTC)
	fixture := newCasePersistenceFixture()
	fixture.sessionRows = [][]any{{"fr/lyon", "case-1", "fr/lyon", "homicide", "npc:victim", "location:fr/lyon/central", "open", createdAt}}
	store := NewPostgresStore(fixture)

	results, err := store.ListOpenPublicCasesForSession(context.Background(), "server-1", "session-token", createdAt)
	if err != nil {
		t.Fatal(err)
	}
	if len(results) != 1 || results[0].ID != "case-1" || results[0].CityID != "fr/lyon" || results[0].Status != "open" || !results[0].CreatedAt.Equal(createdAt) {
		t.Fatalf("unexpected public cases: %+v", results)
	}
	if fixture.sessionQuery == "" || !strings.Contains(fixture.sessionQuery, "s2.status = 'open'") || !strings.Contains(fixture.sessionQuery, "c.city_id = resolved.city_id") {
		t.Fatalf("combined query did not scope open cases to the resolved city: %s", fixture.sessionQuery)
	}
	if len(fixture.sessionArgs) != 5 || fixture.sessionArgs[1] == nil || fixture.sessionArgs[2] != createdAt || fixture.sessionArgs[4] != "fr/lyon" {
		t.Fatalf("unexpected combined query arguments: %#v", fixture.sessionArgs)
	}
}

func TestListOpenPublicCasesForSessionRejectsMissingSession(t *testing.T) {
	store := NewPostgresStore(newCasePersistenceFixture())
	_, err := store.ListOpenPublicCasesForSession(context.Background(), "server-1", "missing-token", time.Now())
	if !errors.Is(err, ErrSessionNotFound) {
		t.Fatalf("missing session error = %v, want %v", err, ErrSessionNotFound)
	}
}

func (fixture *casePersistenceFixture) Begin(context.Context) (pgx.Tx, error) {
	return nil, errors.New("case fixture does not create definitions")
}

type caseFixtureRows struct {
	rows    [][]any
	current []any
	index   int
}

func (rows *caseFixtureRows) Close()                                       {}
func (rows *caseFixtureRows) Err() error                                   { return nil }
func (rows *caseFixtureRows) CommandTag() pgconn.CommandTag                { return pgconn.NewCommandTag("SELECT") }
func (rows *caseFixtureRows) FieldDescriptions() []pgconn.FieldDescription { return nil }
func (rows *caseFixtureRows) RawValues() [][]byte                          { return nil }
func (rows *caseFixtureRows) Conn() *pgx.Conn                              { return nil }
func (rows *caseFixtureRows) Values() ([]any, error)                       { return rows.current, nil }
func (rows *caseFixtureRows) Next() bool {
	if rows.index >= len(rows.rows) {
		return false
	}
	rows.current = rows.rows[rows.index]
	rows.index++
	return true
}
func (rows *caseFixtureRows) Scan(destinations ...any) error {
	return testCaseRow{values: rows.current}.Scan(destinations...)
}

func TestSharedCaseStateAndPlayerProgressPersistSeparately(t *testing.T) {
	fixture := newCasePersistenceFixture()
	fixture.sharedState["server/case-1"] = "open"
	fixture.evidence["server/case-1"] = map[string]bool{"evidence-a": true, "evidence-b": true}
	store := NewPostgresStore(fixture)

	if err := store.SetSharedStatus(context.Background(), "server", "case-1", "solved"); err != nil {
		t.Fatalf("SetSharedStatus() error = %v", err)
	}
	shared, err := store.GetSharedState(context.Background(), "server", "case-1")
	if err != nil {
		t.Fatalf("GetSharedState() error = %v", err)
	}
	if shared.Status != "solved" {
		t.Fatalf("shared status = %q, want solved", shared.Status)
	}

	progress := PlayerProgress{DiscoveredEvidenceIDs: []string{"evidence-b", "evidence-a", "evidence-a"}}
	if err := store.SavePlayerProgress(context.Background(), "server", "case-1", "account-a", progress); err != nil {
		t.Fatalf("SavePlayerProgress() error = %v", err)
	}
	savedProgress, err := store.GetPlayerProgress(context.Background(), "server", "case-1", "account-a")
	if err != nil {
		t.Fatalf("GetPlayerProgress() error = %v", err)
	}
	wantEvidence := []string{"evidence-a", "evidence-b"}
	if len(savedProgress.DiscoveredEvidenceIDs) != len(wantEvidence) || savedProgress.DiscoveredEvidenceIDs[0] != wantEvidence[0] || savedProgress.DiscoveredEvidenceIDs[1] != wantEvidence[1] {
		t.Fatalf("saved progress = %+v, want evidence IDs %v", savedProgress, wantEvidence)
	}

	otherPlayer, err := store.GetPlayerProgress(context.Background(), "server", "case-1", "account-b")
	if err != nil {
		t.Fatalf("GetPlayerProgress() for other account error = %v", err)
	}
	if len(otherPlayer.DiscoveredEvidenceIDs) != 0 {
		t.Fatalf("other player's progress = %+v, want empty", otherPlayer)
	}
	sharedAfterProgress, err := store.GetSharedState(context.Background(), "server", "case-1")
	if err != nil {
		t.Fatalf("GetSharedState() after player progress error = %v", err)
	}
	if sharedAfterProgress.Status != "solved" {
		t.Fatalf("player progress changed shared status to %q", sharedAfterProgress.Status)
	}
}

func TestCaseStoreRejectsInvalidSharedStateAndEvidence(t *testing.T) {
	fixture := newCasePersistenceFixture()
	fixture.sharedState["server/case-1"] = "open"
	store := NewPostgresStore(fixture)

	if err := store.SetSharedStatus(context.Background(), "server", "case-1", "unknown"); !errors.Is(err, ErrInvalidCaseState) {
		t.Fatalf("invalid shared status error = %v, want ErrInvalidCaseState", err)
	}
	if err := store.SetSharedStatus(context.Background(), "server", "missing", "solved"); !errors.Is(err, ErrCaseNotFound) {
		t.Fatalf("unknown case status error = %v, want ErrCaseNotFound", err)
	}
	if err := store.SavePlayerProgress(context.Background(), "server", "case-1", "account-a", PlayerProgress{DiscoveredEvidenceIDs: []string{"missing-evidence"}}); !errors.Is(err, ErrEvidenceNotFound) {
		t.Fatalf("unknown evidence error = %v, want ErrEvidenceNotFound", err)
	}
	if err := store.SavePlayerProgress(context.Background(), "server", "case-1", "account-a", PlayerProgress{DiscoveredEvidenceIDs: []string{" "}}); !errors.Is(err, ErrInvalidProgress) {
		t.Fatalf("blank evidence ID error = %v, want ErrInvalidProgress", err)
	}
}

func TestAccusationChanceRequiresCorrectSuspectAndWeightsEvidence(t *testing.T) {
	if got := accusationChance(false, 10, 10); got != 0 {
		t.Fatalf("chance for an innocent suspect = %d, want 0", got)
	}
	if got := accusationChance(true, 1, 1); got != 75 {
		t.Fatalf("chance for one direct proof = %d, want 75", got)
	}
	if got := accusationChance(true, 2, 1); got != 35 {
		t.Fatalf("chance for one direct and one unsupported proof = %d, want 35", got)
	}
	if got := accusationChance(true, 4, 4); got != 99 {
		t.Fatalf("chance for overwhelming evidence = %d, want cap 99", got)
	}
	if got := accusationChance(true, 1, 0); got != 0 {
		t.Fatalf("chance for unsupported proof = %d, want 0", got)
	}
}

func TestAccusationBoardPinsQueryAcceptsEvidencePinsAndExcludesNotesAsProof(t *testing.T) {
	if !strings.Contains(accusationBoardPinsQuery, "p.designation='accused' AND i.item_source='person'") {
		t.Fatal("accused pin query must require a person source")
	}
	if !strings.Contains(accusationBoardPinsQuery, "p.designation='victim' AND i.item_source='person'") {
		t.Fatal("victim pin query must require a person source")
	}
	if !strings.Contains(accusationBoardPinsQuery, "p.designation='proof' AND i.item_source <> 'note'") {
		t.Fatal("proof pin query must accept evidence sources while excluding blank notes")
	}
	whereClause := strings.SplitN(accusationBoardPinsQuery, "\nWHERE p.server_id", 2)
	if len(whereClause) != 2 || strings.Contains(whereClause[1], "item_source") {
		t.Fatal("proof selection must not be restricted to person pins")
	}
}

func TestAccusationDecisionRollIsStableForRetries(t *testing.T) {
	first := accusationAccepted("server", "account", "case", 75)
	for attempt := 0; attempt < 20; attempt++ {
		if got := accusationAccepted("server", "account", "case", 75); got != first {
			t.Fatalf("retry %d verdict = %t, want stable verdict %t", attempt, got, first)
		}
	}
	if accusationAccepted("server", "account", "case", 0) {
		t.Fatal("zero chance should never accept an accusation")
	}
	if !accusationAccepted("server", "account", "case", 99) {
		// Keep this fixture tied to the deterministic seed so the 99% ceiling stays testable.
		t.Fatal("fixture's deterministic roll should fall below the 99% ceiling")
	}
}

func TestAsyncRequestIDIsStableAndScoped(t *testing.T) {
	want := asyncRequestID("server-a", "account-a", "accusation", "case-1")
	if got := asyncRequestID("server-a", "account-a", "accusation", "case-1"); got != want {
		t.Fatalf("repeated request ID = %q, want %q", got, want)
	}
	if got := asyncRequestID("server-a", "account-b", "accusation", "case-1"); got == want {
		t.Fatal("different player received the same request ID")
	}
	if DefaultAsyncRequestDelay != 3*time.Minute {
		t.Fatalf("default asynchronous delay = %s, want 3 minutes", DefaultAsyncRequestDelay)
	}
}

func TestNPCRecordsQueryIsScopedAndOmitsBackendClassification(t *testing.T) {
	fixture := newCasePersistenceFixture()
	publicRecord := records.Record{
		ID: "record:extortion", NPCID: "npc:victim", Source: "bank",
		OccurredAt: time.Date(1974, 11, 10, 16, 42, 0, 0, time.UTC),
		Operator:   "Meridian Global Settlement Office", Description: "A private remittance cleared.",
	}
	data, err := json.Marshal(publicRecord)
	if err != nil {
		t.Fatal(err)
	}
	fixture.activityRows = [][]any{{data}}
	store := NewPostgresStore(fixture)
	from := time.Date(1974, 11, 1, 0, 0, 0, 0, time.UTC)
	through := time.Date(1974, 11, 12, 23, 59, 0, 0, time.UTC)
	result, err := store.NPCRecords(context.Background(), "server-1", "npc:victim", from, through, "")
	if err != nil {
		t.Fatalf("NPCRecords() error = %v", err)
	}
	if len(result) != 1 || result[0].ID != publicRecord.ID {
		t.Fatalf("NPCRecords() = %+v", result)
	}
	if strings.Contains(fixture.activityQuery, "internal_relation") || strings.Contains(fixture.activityQuery, "case_id") {
		t.Fatalf("public query selects internal case metadata: %s", fixture.activityQuery)
	}
	if len(fixture.activityArgs) != 4 || fixture.activityArgs[0] != "server-1" || fixture.activityArgs[1] != "npc:victim" || fixture.activityArgs[2] != from || fixture.activityArgs[3] != through {
		t.Fatalf("query did not constrain server, NPC and interval: %#v", fixture.activityArgs)
	}
}

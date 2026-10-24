package cases

import (
	"context"
	"encoding/json"
	"errors"
	"strings"
	"testing"
	"time"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgconn"
	"noir-1974/apps/api/internal/records"
)

type testCaseRow struct {
	values []any
	err    error
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
	activityQuery string
	activityArgs  []any
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
	if strings.Contains(query, "FROM case_activity_records") {
		fixture.activityQuery = query
		fixture.activityArgs = args
		return &caseFixtureRows{rows: fixture.activityRows}, nil
	}
	return nil, errors.New("unexpected case row query")
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
		OccurredAt: time.Date(2026, 11, 10, 16, 42, 0, 0, time.UTC),
		Operator:   "Meridian Global Settlement Office", Description: "A private remittance cleared.",
	}
	data, err := json.Marshal(publicRecord)
	if err != nil {
		t.Fatal(err)
	}
	fixture.activityRows = [][]any{{data}}
	store := NewPostgresStore(fixture)
	from := time.Date(2026, 11, 1, 0, 0, 0, 0, time.UTC)
	through := time.Date(2026, 11, 12, 23, 59, 0, 0, time.UTC)
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

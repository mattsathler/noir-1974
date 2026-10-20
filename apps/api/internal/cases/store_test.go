package cases

import (
	"context"
	"errors"
	"strings"
	"testing"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgconn"
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
	sharedState map[string]string
	progress    map[string][]byte
	evidence    map[string]map[string]bool
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
	case strings.Contains(query, "SELECT COUNT(*) FROM case_evidence"):
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

func (fixture *casePersistenceFixture) Query(context.Context, string, ...any) (pgx.Rows, error) {
	return nil, errors.New("case fixture does not support row queries")
}

func (fixture *casePersistenceFixture) Begin(context.Context) (pgx.Tx, error) {
	return nil, errors.New("case fixture does not create definitions")
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

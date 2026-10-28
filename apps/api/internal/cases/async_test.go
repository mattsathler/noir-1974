package cases

import (
	"context"
	"encoding/json"
	"errors"
	"strings"
	"testing"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgconn"
)

type asyncRequestFixture struct {
	requestID   string
	accountID   string
	requestType string
	payload     []byte
	claimed     bool
	completed   bool
	retried     bool
	emails      int
}

func (fixture *asyncRequestFixture) Begin(context.Context) (pgx.Tx, error) {
	return &asyncRequestTx{fixture: fixture, claimTx: !fixture.claimed}, nil
}

func (fixture *asyncRequestFixture) Exec(_ context.Context, query string, _ ...any) (pgconn.CommandTag, error) {
	switch {
	case strings.Contains(query, "SET status='completed'"):
		fixture.completed = true
	case strings.Contains(query, "SET status='pending'"):
		fixture.retried = true
	default:
		return pgconn.CommandTag{}, errors.New("unexpected async fixture exec")
	}
	return pgconn.NewCommandTag("UPDATE 1"), nil
}

func (*asyncRequestFixture) Query(context.Context, string, ...any) (pgx.Rows, error) {
	return nil, errors.New("unexpected async fixture query")
}

func (*asyncRequestFixture) QueryRow(context.Context, string, ...any) pgx.Row {
	return testCaseRow{err: errors.New("unexpected async fixture row query")}
}

type asyncRequestTx struct {
	pgx.Tx
	fixture *asyncRequestFixture
	claimTx bool
}

func (tx *asyncRequestTx) QueryRow(_ context.Context, query string, _ ...any) pgx.Row {
	if tx.claimTx && strings.Contains(query, "FROM player_async_requests") {
		return testCaseRow{values: []any{tx.fixture.requestID, tx.fixture.accountID, tx.fixture.requestType, tx.fixture.payload}}
	}
	return testCaseRow{err: errors.New("unexpected async transaction row query")}
}

func (tx *asyncRequestTx) Exec(_ context.Context, query string, _ ...any) (pgconn.CommandTag, error) {
	switch {
	case tx.claimTx && strings.Contains(query, "SET status='processing'"):
		tx.fixture.claimed = true
	case !tx.claimTx && strings.Contains(query, "INSERT INTO player_inbox_messages"):
		tx.fixture.emails++
	case !tx.claimTx && strings.Contains(query, "UPDATE player_whiteboards SET locked_at=NULL"):
		// Completed reviews release the board for another submission.
	default:
		return pgconn.CommandTag{}, errors.New("unexpected async transaction exec")
	}
	return pgconn.NewCommandTag("UPDATE 1"), nil
}

func (*asyncRequestTx) Commit(context.Context) error   { return nil }
func (*asyncRequestTx) Rollback(context.Context) error { return nil }

func TestAsyncWorkerDeliversUnmatchedAccusationAndCompletesJob(t *testing.T) {
	fixture := &asyncRequestFixture{
		requestID: "async:accusation-1", accountID: "account-1", requestType: "accusation",
		payload: []byte(`{"cityId":"br/rio-de-janeiro","boardId":"board-1","accusedId":"npc-1","victimId":"npc-2","proofIds":["record-1"]}`),
	}
	store := NewPostgresStore(fixture)
	worked, err := store.ProcessNextAsyncRequest(context.Background(), "server-1", nil)
	if err != nil {
		t.Fatalf("ProcessNextAsyncRequest() error = %v", err)
	}
	if !worked || !fixture.completed || fixture.emails != 1 {
		t.Fatalf("worker state = worked:%t completed:%t emails:%d; want true, true, 1", worked, fixture.completed, fixture.emails)
	}
}

func TestAsyncWorkerReschedulesMalformedJob(t *testing.T) {
	fixture := &asyncRequestFixture{
		requestID: "async:broken", accountID: "account-1", requestType: "residential_search", payload: []byte("{"),
	}
	store := NewPostgresStore(fixture)
	worked, err := store.ProcessNextAsyncRequest(context.Background(), "server-1", nil)
	if err == nil {
		t.Fatal("ProcessNextAsyncRequest() error = nil, want payload error")
	}
	if !worked || !fixture.retried || fixture.completed {
		t.Fatalf("worker state = worked:%t retried:%t completed:%t; want true, true, false", worked, fixture.retried, fixture.completed)
	}
}

func TestAccusationSnapshotRoundTripsWithoutBoardLookup(t *testing.T) {
	proofIDs := []string{"case:evidence-1", "mail:warning-2"}
	encoded, err := json.Marshal(accusationPayload{CityID: "br/rio-de-janeiro", BoardID: "board-1", CaseID: "case-1", AccusedID: "npc-killer", VictimID: "npc-victim", ProofIDs: proofIDs})
	if err != nil {
		t.Fatal(err)
	}
	proofIDs[0] = "changed-after-submit"
	var snapshot accusationPayload
	if err := json.Unmarshal(encoded, &snapshot); err != nil {
		t.Fatal(err)
	}
	if len(snapshot.ProofIDs) != 2 || snapshot.ProofIDs[0] != "case:evidence-1" || snapshot.AccusedID != "npc-killer" || snapshot.VictimID != "npc-victim" {
		t.Fatalf("stored accusation snapshot changed after submit: %+v", snapshot)
	}
}

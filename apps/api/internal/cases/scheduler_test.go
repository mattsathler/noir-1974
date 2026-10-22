package cases

import (
	"context"
	"sync"
	"testing"
	"time"

	"noir-1974/apps/api/internal/world"
)

type recurringStoreFixture struct {
	mu        sync.Mutex
	slot      int64
	open      int
	deferred  int
	advanced  int
	created   []Definition
	serverIDs []string
	createdCh chan struct{}
}

func (fixture *recurringStoreFixture) EnsureGenerationSchedule(context.Context, string, time.Duration) error {
	return nil
}
func (fixture *recurringStoreFixture) DueGeneration(_ context.Context, _ string) (int64, bool, error) {
	fixture.mu.Lock()
	defer fixture.mu.Unlock()
	return fixture.slot, true, nil
}
func (fixture *recurringStoreFixture) CountOpenCases(context.Context, string, string) (int, error) {
	return fixture.open, nil
}
func (fixture *recurringStoreFixture) UsedNPCIDs(context.Context, string, string, string) (map[string]struct{}, error) {
	return map[string]struct{}{}, nil
}
func (fixture *recurringStoreFixture) Create(_ context.Context, serverID string, base world.World, definition Definition) error {
	fixture.mu.Lock()
	defer fixture.mu.Unlock()
	fixture.created = append(fixture.created, definition)
	fixture.serverIDs = append(fixture.serverIDs, serverID)
	if fixture.createdCh != nil {
		select {
		case fixture.createdCh <- struct{}{}:
		default:
		}
	}
	return nil
}
func (fixture *recurringStoreFixture) AdvanceGeneration(_ context.Context, _ string, slot int64, _ time.Duration) error {
	fixture.mu.Lock()
	defer fixture.mu.Unlock()
	if slot != fixture.slot {
		return context.Canceled
	}
	fixture.slot++
	fixture.advanced++
	return nil
}
func (fixture *recurringStoreFixture) DeferGeneration(_ context.Context, _ string, slot int64, _ time.Duration) error {
	fixture.mu.Lock()
	defer fixture.mu.Unlock()
	if slot != fixture.slot {
		return context.Canceled
	}
	fixture.deferred++
	return nil
}

type recurringWorldFixture struct {
	base       world.World
	deadNPCIDs []string
}

func (fixture *recurringWorldFixture) WorldForCases(context.Context, string) (world.World, error) {
	return fixture.base, nil
}
func (fixture *recurringWorldFixture) ExpandDistrictSociety(context.Context, string, string) error {
	return nil
}
func (fixture *recurringWorldFixture) SetNPCStatus(_ context.Context, _, npcID, status string) error {
	if status == "dead" {
		fixture.deadNPCIDs = append(fixture.deadNPCIDs, npcID)
	}
	return nil
}

func newRecurringTestGenerator(t *testing.T, store *recurringStoreFixture, worldStore *recurringWorldFixture, lock generationLock) *RecurringGenerator {
	t.Helper()
	generator := NewRecurringGenerator("unused-in-test", "server-shared", store, worldStore, time.Minute, 3)
	generator.withGenerationLock = lock
	return generator
}

func recurringTestWorld(t *testing.T) world.World {
	t.Helper()
	generated, err := world.Generate("m10-recurring-test", world.GeneratorVersion5)
	if err != nil {
		t.Fatalf("Generate() test world error = %v", err)
	}
	return generated
}

func TestRecurringGeneratorPersistsCursorAcrossRestartAndSharesCreatedCase(t *testing.T) {
	store := &recurringStoreFixture{slot: 9}
	worldStore := &recurringWorldFixture{base: recurringTestWorld(t)}
	lock := func(_ context.Context, generate func() error) error { return generate() }

	firstProcess := newRecurringTestGenerator(t, store, worldStore, lock)
	if err := firstProcess.runOne(context.Background()); err != nil {
		t.Fatalf("first runOne() error = %v", err)
	}

	// A new generator uses the same store, representing the persisted schedule cursor.
	restartedProcess := newRecurringTestGenerator(t, store, worldStore, lock)
	if err := restartedProcess.runOne(context.Background()); err != nil {
		t.Fatalf("runOne() after restart error = %v", err)
	}
	if store.slot != 11 || store.advanced != 2 || len(store.created) != 2 {
		t.Fatalf("cursor=%d advanced=%d created=%d; want 11, 2, 2", store.slot, store.advanced, len(store.created))
	}
	if store.created[0].ID == store.created[1].ID {
		t.Fatalf("successive cursor slots produced duplicate case ID %q", store.created[0].ID)
	}
	if len(worldStore.deadNPCIDs) != 2 || worldStore.deadNPCIDs[0] != store.created[0].VictimID || worldStore.deadNPCIDs[1] != store.created[1].VictimID {
		t.Fatalf("shared victim status updates = %v, want generated victims %q and %q", worldStore.deadNPCIDs, store.created[0].VictimID, store.created[1].VictimID)
	}
	if len(store.serverIDs) != 2 || store.serverIDs[0] != "server-shared" || store.serverIDs[1] != "server-shared" {
		t.Fatalf("created cases used server IDs %v, want the shared server", store.serverIDs)
	}
}

func TestRecurringGeneratorDefersWhenOpenCaseLimitIsReached(t *testing.T) {
	store := &recurringStoreFixture{slot: 4, open: 3}
	worldStore := &recurringWorldFixture{base: recurringTestWorld(t)}
	generator := newRecurringTestGenerator(t, store, worldStore, func(_ context.Context, generate func() error) error { return generate() })

	if err := generator.runOne(context.Background()); err != nil {
		t.Fatalf("runOne() at open-case limit error = %v", err)
	}
	if store.deferred != 1 || store.advanced != 0 || len(store.created) != 0 {
		t.Fatalf("deferred=%d advanced=%d created=%d; want 1, 0, 0", store.deferred, store.advanced, len(store.created))
	}
}

func TestRecurringGeneratorRunsOnConfiguredIntervalUntilStopped(t *testing.T) {
	store := &recurringStoreFixture{createdCh: make(chan struct{}, 1)}
	worldStore := &recurringWorldFixture{base: recurringTestWorld(t)}
	generator := newRecurringTestGenerator(t, store, worldStore, func(_ context.Context, generate func() error) error { return generate() })
	generator.interval = time.Second
	ctx, cancel := context.WithCancel(context.Background())
	done := make(chan struct{})
	go func() {
		generator.Run(ctx)
		close(done)
	}()
	select {
	case <-store.createdCh:
	case <-time.After(3 * time.Second):
		cancel()
		<-done
		t.Fatal("recurring generator did not dispatch a due case")
	}
	cancel()
	select {
	case <-done:
	case <-time.After(time.Second):
		t.Fatal("recurring generator did not stop after context cancellation")
	}
	if store.advanced == 0 {
		t.Fatal("recurring generator created a case without advancing the persisted cursor")
	}
}

func TestRecurringGeneratorSerializesInstancesForOneServer(t *testing.T) {
	store := &recurringStoreFixture{}
	worldStore := &recurringWorldFixture{base: recurringTestWorld(t)}
	var advisoryLock sync.Mutex
	lock := func(_ context.Context, generate func() error) error {
		advisoryLock.Lock()
		defer advisoryLock.Unlock()
		return generate()
	}
	instances := []*RecurringGenerator{
		newRecurringTestGenerator(t, store, worldStore, lock),
		newRecurringTestGenerator(t, store, worldStore, lock),
	}
	var workers sync.WaitGroup
	errors := make(chan error, len(instances))
	for _, instance := range instances {
		workers.Add(1)
		go func(generator *RecurringGenerator) {
			defer workers.Done()
			errors <- generator.runOne(context.Background())
		}(instance)
	}
	workers.Wait()
	close(errors)
	for err := range errors {
		if err != nil {
			t.Fatalf("concurrent runOne() error = %v", err)
		}
	}
	if store.advanced != 2 || len(store.created) != 2 || store.slot != 2 {
		t.Fatalf("concurrent schedule state: cursor=%d advanced=%d created=%d; want 2, 2, 2", store.slot, store.advanced, len(store.created))
	}
	if store.created[0].ID == store.created[1].ID {
		t.Fatalf("concurrent instances created duplicate case %q", store.created[0].ID)
	}
}

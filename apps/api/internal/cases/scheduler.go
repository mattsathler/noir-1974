package cases

import (
	"context"
	"errors"
	"fmt"
	"log"
	"sort"
	"time"

	"github.com/jackc/pgx/v5"
	"noir-1974/apps/api/internal/world"
)

type recurringWorld interface {
	WorldForCases(context.Context, string) (world.World, error)
	ExpandDistrictSociety(context.Context, string, string) error
	SetNPCStatus(context.Context, string, string, string) error
}

type recurringStore interface {
	EnsureGenerationSchedule(context.Context, string, time.Duration) error
	DueGeneration(context.Context, string) (int64, bool, error)
	CountOpenCases(context.Context, string) (int, error)
	UsedNPCIDs(context.Context, string, string) (map[string]struct{}, error)
	Create(context.Context, string, world.World, Definition) error
	AdvanceGeneration(context.Context, string, int64, time.Duration) error
	DeferGeneration(context.Context, string, int64, time.Duration) error
}

type generationLock func(context.Context, func() error) error

type RecurringGenerator struct {
	databaseURL, serverID string
	store                 recurringStore
	world                 recurringWorld
	interval              time.Duration
	maxOpen               int
	withGenerationLock    generationLock
}

func NewRecurringGenerator(databaseURL, serverID string, store recurringStore, worldStore recurringWorld, interval time.Duration, maxOpen int) *RecurringGenerator {
	generator := &RecurringGenerator{databaseURL: databaseURL, serverID: serverID, store: store, world: worldStore, interval: interval, maxOpen: maxOpen}
	generator.withGenerationLock = generator.withAdvisoryLock
	return generator
}

func (g *RecurringGenerator) Run(ctx context.Context) {
	poll := 30 * time.Second
	if g.interval < poll {
		poll = g.interval
	}
	if poll < time.Second {
		poll = time.Second
	}
	ticker := time.NewTicker(poll)
	defer ticker.Stop()
	for {
		select {
		case <-ctx.Done():
			return
		case <-ticker.C:
			if err := g.runOne(ctx); err != nil {
				log.Printf("recurring case generation: %v", err)
			}
		}
	}
}

func (g *RecurringGenerator) runOne(ctx context.Context) error {
	return g.withGenerationLock(ctx, func() error { return g.generateOne(ctx) })
}

func (g *RecurringGenerator) withAdvisoryLock(ctx context.Context, generate func() error) error {
	conn, err := pgx.Connect(ctx, g.databaseURL)
	if err != nil {
		return fmt.Errorf("connect generation lock: %w", err)
	}
	defer conn.Close(context.Background())
	lockTx, err := conn.Begin(ctx)
	if err != nil {
		return err
	}
	defer func() { _ = lockTx.Rollback(context.Background()) }()
	var locked bool
	if err := lockTx.QueryRow(ctx, `SELECT pg_try_advisory_xact_lock(hashtextextended($1,0))`, `noir:case-generation:`+g.serverID).Scan(&locked); err != nil {
		return err
	}
	if !locked {
		return nil
	}
	return generate()
}

func (g *RecurringGenerator) generateOne(ctx context.Context) error {
	slot, due, err := g.store.DueGeneration(ctx, g.serverID)
	if err != nil || !due {
		return err
	}
	open, err := g.store.CountOpenCases(ctx, g.serverID)
	if err != nil {
		return err
	}
	if open >= g.maxOpen {
		return g.store.DeferGeneration(ctx, g.serverID, slot, g.interval)
	}
	seed := fmt.Sprintf("noir-1974/recurring-case/%s/%d", g.serverID, slot)
	base, err := g.world.WorldForCases(ctx, g.serverID)
	if err != nil {
		return err
	}
	used, err := g.store.UsedNPCIDs(ctx, g.serverID, seed)
	if err != nil {
		return err
	}
	var definition Definition
	for attempt := 0; attempt < 256; attempt++ {
		definition, err = GenerateExcluding(base, seed, GeneratorVersion, used)
		if err == nil {
			break
		}
		if !errors.Is(err, ErrNoValidCaseCandidates) {
			return err
		}
		district := leastPopulatedDistrict(base)
		if district == "" {
			_ = g.store.DeferGeneration(ctx, g.serverID, slot, g.interval)
			return err
		}
		if err := g.world.ExpandDistrictSociety(ctx, g.serverID, district); err != nil {
			_ = g.store.DeferGeneration(ctx, g.serverID, slot, g.interval)
			return fmt.Errorf("expand district population: %w", err)
		}
		base, err = g.world.WorldForCases(ctx, g.serverID)
		if err != nil {
			return err
		}
	}
	if err != nil {
		_ = g.store.DeferGeneration(ctx, g.serverID, slot, g.interval)
		return fmt.Errorf("could not find unused case candidates: %w", err)
	}
	if err := g.store.Create(ctx, g.serverID, base, definition); err != nil {
		return err
	}
	if err := g.world.SetNPCStatus(ctx, g.serverID, definition.VictimID, "dead"); err != nil {
		return err
	}
	return g.store.AdvanceGeneration(ctx, g.serverID, slot, g.interval)
}

func leastPopulatedDistrict(base world.World) string {
	counts := map[string]int{}
	for _, npc := range base.NPCs {
		counts[npc.RegionID]++
	}
	var districts []string
	for _, region := range base.Regions {
		if region.Kind == "district" {
			districts = append(districts, region.ID)
		}
	}
	sort.Strings(districts)
	selected := ""
	minimum := int(^uint(0) >> 1)
	for _, id := range districts {
		if counts[id] < minimum {
			selected = id
			minimum = counts[id]
		}
	}
	return selected
}

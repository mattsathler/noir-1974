package cases

import (
	"context"
	"errors"
	"fmt"
	"log"
	"sort"
	"strings"
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
	CountOpenCases(context.Context, string, string) (int, error)
	UsedNPCIDs(context.Context, string, string, string) (map[string]struct{}, error)
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
	base, err := g.world.WorldForCases(ctx, g.serverID)
	if err != nil {
		return err
	}
	cityIDs := materializedCityIDs(base)
	cityID := cityIDs[int(slot-1)%len(cityIDs)]
	base = worldForCity(base, cityID)
	open, err := g.store.CountOpenCases(ctx, g.serverID, cityID)
	if err != nil {
		return err
	}
	if open >= g.maxOpen {
		return g.store.DeferGeneration(ctx, g.serverID, slot, g.interval)
	}
	seed := fmt.Sprintf("noir-1974/recurring-case/%s/%s/%d", g.serverID, cityID, slot)
	used, err := g.store.UsedNPCIDs(ctx, g.serverID, cityID, seed)
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
		base = worldForCity(base, cityID)
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

func materializedCityIDs(base world.World) []string {
	ids := make([]string, 0)
	for _, city := range world.Cities() {
		prefix := city.ID + "/district/"
		for _, npc := range base.NPCs {
			if strings.HasPrefix(npc.RegionID, prefix) || city.ID == "fr/lyon" && !strings.Contains(npc.RegionID, "/district/") {
				ids = append(ids, city.ID)
				break
			}
		}
	}
	if len(ids) == 0 {
		return []string{"fr/lyon"}
	}
	return ids
}

func worldForCity(base world.World, cityID string) world.World {
	if cityID == "fr/lyon" {
		for _, npc := range base.NPCs {
			if !strings.Contains(npc.RegionID, "/district/") {
				return base
			}
		}
	}
	local := world.World{GeneratorVersion: base.GeneratorVersion}
	regions := map[string]bool{}
	prefix := cityID + "/district/"
	for _, region := range base.Regions {
		if region.ID == cityID || strings.HasPrefix(region.ID, prefix) {
			local.Regions = append(local.Regions, region)
			regions[region.ID] = true
		}
	}
	entities := map[string]bool{}
	for _, entity := range base.Entities {
		if regions[entity.RegionID] {
			local.Entities = append(local.Entities, entity)
			entities[entity.ID] = true
		}
	}
	allowedNPCs := map[string]bool{}
	for _, npc := range base.NPCs {
		if regions[npc.RegionID] {
			local.NPCs = append(local.NPCs, npc)
			allowedNPCs[npc.ID] = true
		}
	}
	allowedCompanies := map[string]bool{}
	for _, company := range base.Companies {
		if regions[company.RegionID] {
			local.Companies = append(local.Companies, company)
			allowedCompanies[company.ID] = true
		}
	}
	for _, relation := range base.EmploymentRelationships {
		if allowedNPCs[relation.EmployeeNPCID] && allowedCompanies[relation.EmployerCompanyID] {
			local.EmploymentRelationships = append(local.EmploymentRelationships, relation)
		}
	}
	for _, relation := range base.PersonalRelationships {
		if allowedNPCs[relation.PersonANPCID] && allowedNPCs[relation.PersonBNPCID] {
			local.PersonalRelationships = append(local.PersonalRelationships, relation)
		}
	}
	for _, relation := range base.CompanyLocations {
		if allowedCompanies[relation.CompanyID] && entities[relation.LocationID] {
			local.CompanyLocations = append(local.CompanyLocations, relation)
		}
	}
	return local
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

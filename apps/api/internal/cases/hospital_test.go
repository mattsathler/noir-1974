package cases

import (
	"context"
	"errors"
	"fmt"
	"reflect"
	"regexp"
	"strings"
	"testing"
	"time"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgconn"
	"noir-1974/apps/api/internal/world"
)

type hospitalTestDB struct {
	seed      string
	version   int
	base      world.World
	status    string
	hasCase   bool
	caseID    string
	archetype string
	killerID  string
	sceneID   string
	deathAt   time.Time
}

type hospitalTestRow struct {
	values []any
	err    error
}

func (row hospitalTestRow) Scan(destinations ...any) error {
	if row.err != nil {
		return row.err
	}
	if len(destinations) != len(row.values) {
		return fmt.Errorf("unexpected hospital test row width: %d", len(destinations))
	}
	for index, destination := range destinations {
		switch target := destination.(type) {
		case *string:
			*target = row.values[index].(string)
		case *int:
			*target = row.values[index].(int)
		case *time.Time:
			*target = row.values[index].(time.Time)
		default:
			return fmt.Errorf("unsupported hospital test scan target %T", destination)
		}
	}
	return nil
}

func (db *hospitalTestDB) Exec(context.Context, string, ...any) (pgconn.CommandTag, error) {
	return pgconn.CommandTag{}, errors.New("unexpected hospital test exec")
}

func (db *hospitalTestDB) Query(context.Context, string, ...any) (pgx.Rows, error) {
	return &caseFixtureRows{}, nil
}

func (db *hospitalTestDB) QueryRow(_ context.Context, query string, args ...any) pgx.Row {
	switch {
	case strings.Contains(query, "SELECT world_seed, generator_version FROM worlds"):
		return hospitalTestRow{values: []any{db.seed, db.version}}
	case strings.Contains(query, "SELECT value #>> '{}' FROM world_mutations"):
		if db.status == "" {
			return hospitalTestRow{err: pgx.ErrNoRows}
		}
		return hospitalTestRow{values: []any{db.status}}
	case strings.Contains(query, "SELECT c.case_id,c.archetype_id,c.killer_entity_id,c.location_entity_id"):
		if !db.hasCase {
			return hospitalTestRow{err: pgx.ErrNoRows}
		}
		return hospitalTestRow{values: []any{db.caseID, db.archetype, db.killerID, db.sceneID, db.deathAt}}
	case strings.Contains(query, "SELECT e.region_id FROM world_entities"):
		for _, entity := range db.base.Entities {
			if entity.ID == db.sceneID {
				return hospitalTestRow{values: []any{entity.RegionID}}
			}
		}
		return hospitalTestRow{err: pgx.ErrNoRows}
	case strings.Contains(query, hospitalRegionNameQuery):
		for _, region := range db.base.Regions {
			if region.ID == args[1].(string) {
				return hospitalTestRow{values: []any{region.Name}}
			}
		}
		return hospitalTestRow{err: pgx.ErrNoRows}
	case strings.Contains(query, "SELECT generation_key FROM world_entities"):
		for _, npc := range db.base.NPCs {
			if npc.ID == args[1].(string) {
				return hospitalTestRow{values: []any{npc.GenerationKey}}
			}
		}
		return hospitalTestRow{err: pgx.ErrNoRows}
	default:
		return hospitalTestRow{err: fmt.Errorf("unexpected hospital test query: %s", query)}
	}
}

func (*hospitalTestDB) Begin(context.Context) (pgx.Tx, error) {
	return nil, errors.New("unexpected hospital test transaction")
}

func TestHospitalRegionNameQueryUsesWorldRegionDisplayName(t *testing.T) {
	if !strings.Contains(hospitalRegionNameQuery, "SELECT display_name FROM world_regions") {
		t.Fatalf("hospital region query must read world_regions.display_name, got %q", hospitalRegionNameQuery)
	}
}

func TestFingerprintCodesAreStableUniqueAndFormatted(t *testing.T) {
	ids := []string{"npc:fr/lyon/person/one", "npc:fr/lyon/person/two", "npc:fr/lyon/person/three"}
	first := fingerprintCodes(ids)
	reversed := []string{ids[2], ids[1], ids[0]}
	second := fingerprintCodes(reversed)
	pattern := regexp.MustCompile(`^[A-Z]{2}[0-9]{5}$`)
	seen := make(map[string]string, len(ids))
	for _, id := range ids {
		code := first[id]
		if !pattern.MatchString(code) {
			t.Fatalf("fingerprint for %s has invalid format %q", id, code)
		}
		if prior, exists := seen[code]; exists {
			t.Fatalf("NPCs %s and %s share fingerprint %s", prior, id, code)
		}
		seen[code] = id
		if second[id] != code {
			t.Fatalf("fingerprint changed with input ordering for %s: %s != %s", id, code, second[id])
		}
	}
}

func TestResidentialSceneClassifierUsesWholeGenerationSegments(t *testing.T) {
	for _, key := range []string{"fr/lyon/district/a/building/immeuble-saint-paul", "jp/tokyo/district/a/building/apartment-12", "us/boston/home/house-4"} {
		if !isResidentialGenerationKey(key) {
			t.Errorf("expected residential key %q to match", key)
		}
	}
	for _, key := range []string{"fr/lyon/district/a/building/warehouse", "fr/lyon/district/a/building/hotel-des-voyageurs", "fr/lyon/district/a/street/house-martin"} {
		if isResidentialGenerationKey(key) {
			t.Errorf("non-residential key %q matched as a residence", key)
		}
	}
}

func TestPostMortemTraceIsCaseStableAndDoesNotIdentifyOwner(t *testing.T) {
	var fingerprintCase, hairCase string
	for index := 0; index < 100; index++ {
		caseID := fmt.Sprintf("case-%d", index)
		if postMortemUsesFingerprint(caseID) && fingerprintCase == "" {
			fingerprintCase = caseID
		}
		if !postMortemUsesFingerprint(caseID) && hairCase == "" {
			hairCase = caseID
		}
	}
	if fingerprintCase == "" || hairCase == "" {
		t.Fatal("expected deterministic case selection to cover both trace types")
	}
	for _, test := range []struct {
		caseID string
		want   string
	}{
		{caseID: fingerprintCase, want: "AB12"},
		{caseID: hairCase, want: "ginger"},
	} {
		first := postMortemTrace(test.caseID, "AB12345", "ginger")
		second := postMortemTrace(test.caseID, "AB12345", "ginger")
		if first != second || !strings.Contains(first.Content, test.want) {
			t.Fatalf("case %s did not reproduce its selected anonymous trace: %+v", test.caseID, first)
		}
		if strings.Contains(first.Content, "killer") || strings.Contains(first.Content, "suspect") {
			t.Fatalf("trace should not identify or classify its owner: %q", first.Content)
		}
	}
}

func TestResidentialFingerprintChanceIsStableAndNeverAppliesOffSite(t *testing.T) {
	const sampleCount = 1000
	selected := 0
	for index := 0; index < sampleCount; index++ {
		caseID := fmt.Sprintf("case-%d", index)
		first := residentialFingerprintEligible(caseID, true)
		if first != residentialFingerprintEligible(caseID, true) {
			t.Fatalf("residential warrant result changed on retry for %s", caseID)
		}
		if residentialFingerprintEligible(caseID, false) {
			t.Fatalf("non-residential case %s produced a residential trace", caseID)
		}
		if first {
			selected++
		}
	}
	if selected < 250 || selected > 450 {
		t.Fatalf("SHA-256 residential chance selected %d/%d cases, want approximately 35%%", selected, sampleCount)
	}
}

func TestHospitalReconstructsHomicideRecordAndOmitsItForNonCases(t *testing.T) {
	const seed = "m21-hospital-reconstruction"
	base, err := world.GenerateCity(seed, world.LatestGeneratorVersion, "fr/lyon")
	if err != nil {
		t.Fatal(err)
	}
	if len(base.NPCs) < 2 {
		t.Fatal("hospital fixture city needs at least two NPCs")
	}
	var scene world.Entity
	for _, entity := range base.Entities {
		if entity.EntityType == "location" {
			scene = entity
			break
		}
	}
	if scene.ID == "" {
		t.Fatal("hospital fixture city has no crime scene")
	}
	victim, killer := base.NPCs[0], base.NPCs[1]
	deathAt := time.Date(1974, time.October, 23, 22, 30, 0, 0, time.UTC)
	database := &hospitalTestDB{
		seed: seed, version: world.LatestGeneratorVersion, base: base, status: "dead", hasCase: true,
		caseID: "case:hospital-reconstruction", archetype: "jealousy", killerID: killer.ID,
		sceneID: scene.ID, deathAt: deathAt,
	}
	store := NewPostgresStore(database)
	first, firstItems, err := store.SearchHospitalPatients(context.Background(), "server-1", "fr/lyon", victim.Name)
	if err != nil {
		t.Fatalf("SearchHospitalPatients() error = %v", err)
	}
	if len(first) != 1 || first[0].PostMortem == nil {
		t.Fatalf("homicide victim should have one post-mortem file: %+v", first)
	}
	record := first[0].PostMortem
	if !record.EstimatedDeathAt.Equal(deathAt) || record.Cause == "" || !strings.Contains(record.BodyFoundAt, scene.Name) {
		t.Fatalf("post-mortem fields were not reconstructed from the case: %+v", record)
	}
	if len(record.Evidence) != 1 || len(firstItems) != 2 {
		t.Fatalf("case report and trace should both be discoverable: record=%+v items=%+v", record, firstItems)
	}
	second, secondItems, err := store.SearchHospitalPatients(context.Background(), "server-1", "fr/lyon", victim.Name)
	if err != nil {
		t.Fatalf("retry SearchHospitalPatients() error = %v", err)
	}
	if !reflect.DeepEqual(first, second) || !reflect.DeepEqual(firstItems, secondItems) {
		t.Fatalf("reconstructed hospital records changed on retry: first=%+v second=%+v", first, second)
	}

	database.status = "dead"
	database.hasCase = false
	nonCase, nonCaseItems, err := store.SearchHospitalPatients(context.Background(), "server-1", "fr/lyon", victim.Name)
	if err != nil {
		t.Fatalf("non-case SearchHospitalPatients() error = %v", err)
	}
	if len(nonCase) != 1 || nonCase[0].PostMortem != nil || len(nonCaseItems) != 0 {
		t.Fatalf("dead NPC without a homicide case must not receive invented forensics: patients=%+v items=%+v", nonCase, nonCaseItems)
	}

	database.status = "alive"
	database.hasCase = true
	living, livingItems, err := store.SearchHospitalPatients(context.Background(), "server-1", "fr/lyon", victim.Name)
	if err != nil {
		t.Fatalf("living SearchHospitalPatients() error = %v", err)
	}
	if len(living) != 1 || living[0].PostMortem != nil || len(livingItems) != 0 {
		t.Fatalf("living NPC must not receive homicide data: patients=%+v items=%+v", living, livingItems)
	}
}

func TestPostMortemReportIsPinnableWithoutYear(t *testing.T) {
	record := &PostMortemRecord{
		ReportID:         "case:postmortem:case-1",
		EstimatedDeathAt: time.Date(2026, time.October, 23, 23, 55, 0, 0, time.UTC),
		Cause:            "Fatal blunt force injuries",
		BodyFoundAt:      "Lyon, France",
	}
	victim := world.NPC{ID: "person-1", Name: "Pascal Poulin", Visual: &world.VisualIdentity{HairID: "hair_short_ginger", SkinID: "skin_white", EyesID: "eyes_brown"}}
	item := postMortemReportItem(record, world.WorldCity{TimeZone: "Europe/Paris"}, victim)
	if item.ID != record.ReportID || item.Source != "case" {
		t.Fatalf("post-mortem discovery must use its stable report ID and pinnable source, got %+v", item)
	}
	if !strings.Contains(item.Content, "ESTIMATED TIME OF DEATH · Oct 24, 1:55 AM") {
		t.Fatalf("report pin must show local estimated time without a year, got %q", item.Content)
	}
	if strings.Contains(item.Content, "2026") || !strings.Contains(item.Content, record.Cause) {
		t.Fatalf("report pin should omit the year and retain the cause, got %q", item.Content)
	}
	if !strings.Contains(string(item.Metadata), `"name":"Pascal Poulin"`) || !strings.Contains(string(item.Metadata), `"hairId":"hair_short_ginger"`) {
		t.Fatalf("report pin metadata should preserve the victim portrait, got %s", item.Metadata)
	}
}

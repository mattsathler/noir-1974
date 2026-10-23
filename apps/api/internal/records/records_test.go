package records

import (
	"encoding/json"
	"reflect"
	"testing"
	"time"

	"context"
	"noir-1974/apps/api/internal/world"
)

type fixtureWorldReader struct {
	base    world.World
	profile world.NPCRecordsProfile
}

func (reader fixtureWorldReader) GetNPCForRecords(context.Context, string, string, bool) (world.NPCRecordsProfile, error) {
	return reader.profile, nil
}

type fixtureCaseReader struct{ items []Record }

func (reader fixtureCaseReader) NPCRecords(context.Context, string, string, time.Time, time.Time, string) ([]Record, error) {
	return reader.items, nil
}

func TestGenerateIsDeterministicAndLimitedToNPCAndInterval(t *testing.T) {
	from := time.Date(2026, time.November, 1, 0, 0, 0, 0, time.UTC)
	through := time.Date(2026, time.November, 7, 23, 59, 59, 0, time.UTC)
	context := Context{RegionName: "Bellecour", HomeDistrict: "Bellecour", Employer: "Aurore Industries"}
	first, err := Generate("records-seed", 2, "npc:ada", context, from, through)
	if err != nil {
		t.Fatalf("Generate() error = %v", err)
	}
	second, err := Generate("records-seed", 2, "npc:ada", context, from, through)
	if err != nil {
		t.Fatalf("Generate() second call error = %v", err)
	}
	if !reflect.DeepEqual(first, second) {
		t.Fatal("same seed, world version, NPC and interval produced different records")
	}
	if len(first) < 2 {
		t.Fatalf("got %d records, want records from ordinary financial and GATE activity", len(first))
	}
	sources := map[string]bool{}
	for _, item := range first {
		if item.NPCID != "npc:ada" || item.OccurredAt.Before(from) || item.OccurredAt.After(through) {
			t.Fatalf("record escaped the requested NPC or interval: %+v", item)
		}
		if item.CaseID != "" || item.InternalRelation != "" {
			t.Fatalf("routine record has case classification: %+v", item)
		}
		sources[item.Source] = true
	}
	if !sources["bank"] || !sources["gate"] {
		t.Fatalf("query did not include both corporate record systems: %v", sources)
	}
}

func TestExtortionEventsKeepCaseLinksOutOfPublicJSON(t *testing.T) {
	events := ExtortionEvents("case-seed", 2, "case:123", "npc:killer", "npc:victim", "Bellecour", time.Date(2026, time.November, 12, 0, 0, 0, 0, time.UTC))
	if len(events) != 2 || events[0].InternalRelation != "extortion_payment" || events[1].InternalRelation != "precrime_route" {
		t.Fatalf("unexpected internal event relationships: %+v", events)
	}
	for _, event := range events {
		if event.CaseID != "case:123" || event.ID == "" {
			t.Fatalf("event missing deterministic identity or case link: %+v", event)
		}
		data, err := json.Marshal(event)
		if err != nil {
			t.Fatal(err)
		}
		var public map[string]any
		if err := json.Unmarshal(data, &public); err != nil {
			t.Fatal(err)
		}
		if _, exists := public["caseId"]; exists {
			t.Fatalf("public record revealed its case: %s", data)
		}
		if _, exists := public["internalRelation"]; exists {
			t.Fatalf("public record revealed its internal classification: %s", data)
		}
	}
}

func TestGenerateRejectsInvalidAndOversizedIntervals(t *testing.T) {
	from := time.Date(2026, time.January, 1, 0, 0, 0, 0, time.UTC)
	if _, err := Generate("seed", 2, "npc:ada", Context{}, from, from.AddDate(0, 0, MaxRangeDays)); err != ErrRangeTooLarge {
		t.Fatalf("Generate() eight calendar day interval error = %v, want %v", err, ErrRangeTooLarge)
	}
	if _, err := Generate("seed", 2, "npc:ada", Context{}, from, from.Add(367*24*time.Hour)); err != ErrRangeTooLarge {
		t.Fatalf("Generate() oversized interval error = %v, want %v", err, ErrRangeTooLarge)
	}
	if _, err := Generate("seed", 2, "npc:ada", Context{}, from, from.Add(-24*time.Hour)); err != ErrInvalidQuery {
		t.Fatalf("Generate() reversed interval error = %v, want %v", err, ErrInvalidQuery)
	}
}

func TestServiceMergesAndSortsRoutineAndCaseEvents(t *testing.T) {
	base, err := world.Generate("records-seed", world.GeneratorVersion2)
	if err != nil {
		t.Fatal(err)
	}
	npc := base.NPCs[0]
	profile := world.NPCRecordsProfile{NPC: npc, WorldSeed: "records-seed", GeneratorVersion: world.GeneratorVersion2, Employer: base.Companies[0].Name}
	linked := Record{
		ID: "record:linked", NPCID: npc.ID, Source: "bank",
		OccurredAt:  time.Date(2026, 11, 6, 12, 0, 0, 0, time.UTC),
		Description: "A private remittance cleared.", CaseID: "case:secret", InternalRelation: "extortion_payment",
	}
	service := NewService(fixtureWorldReader{base: base, profile: profile}, fixtureCaseReader{items: []Record{linked}})
	from := time.Date(2026, 11, 6, 0, 0, 0, 0, time.UTC)
	through := time.Date(2026, 11, 12, 23, 59, 0, 0, time.UTC)
	result, err := service.NPCRecords(context.Background(), "server", npc.ID, from, through, "")
	if err != nil {
		t.Fatalf("NPCRecords() error = %v", err)
	}
	if len(result) < 3 {
		t.Fatalf("got %d events, expected case and routine activity mixed", len(result))
	}
	foundLinked := false
	for index, item := range result {
		if item.ID == linked.ID {
			foundLinked = true
			if item.CaseID != "" || item.InternalRelation != "" {
				t.Fatalf("service leaked classification fields: %+v", item)
			}
		}
		if index > 0 && result[index-1].OccurredAt.After(item.OccurredAt) {
			t.Fatalf("records are out of time order: %+v", result)
		}
	}
	if !foundLinked {
		t.Fatal("the NPC's case-related record was missing from the merged activity")
	}
}

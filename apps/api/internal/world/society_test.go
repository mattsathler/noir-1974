package world

import (
	"reflect"
	"testing"
)

func TestGenerateSocietyIsDeterministic(t *testing.T) {
	first, err := Generate("m03-society-seed", GeneratorVersion2)
	if err != nil {
		t.Fatalf("Generate() error = %v", err)
	}
	second, err := Generate("m03-society-seed", GeneratorVersion2)
	if err != nil {
		t.Fatalf("Generate() second call error = %v", err)
	}
	if !reflect.DeepEqual(first, second) {
		t.Fatal("same seed and generator version produced different societies")
	}
}

func TestGenerateSocietyProducesNPCsCompaniesAndRelationships(t *testing.T) {
	generated, err := Generate("m03-shape-seed", GeneratorVersion2)
	if err != nil {
		t.Fatalf("Generate() error = %v", err)
	}
	if generated.GeneratorVersion != GeneratorVersion2 {
		t.Fatalf("generator version = %d, want %d", generated.GeneratorVersion, GeneratorVersion2)
	}
	if len(generated.NPCs) != 24 || len(generated.Companies) != 3 {
		t.Fatalf("society size = %d NPCs, %d companies; want 24 and 3", len(generated.NPCs), len(generated.Companies))
	}
	if len(generated.EmploymentRelationships) != 24 || len(generated.PersonalRelationships) != 12 || len(generated.CompanyLocations) != 3 {
		t.Fatalf("relationship counts = %d employment, %d personal, %d company locations; want 24, 12, 3", len(generated.EmploymentRelationships), len(generated.PersonalRelationships), len(generated.CompanyLocations))
	}

	entityByID := make(map[string]Entity, len(generated.Entities))
	for _, entity := range generated.Entities {
		entityByID[entity.ID] = entity
	}
	for _, relationship := range generated.EmploymentRelationships {
		if entityByID[relationship.EmployeeNPCID].EntityType != "npc" || entityByID[relationship.EmployerCompanyID].EntityType != "company" {
			t.Fatalf("employment relationship has invalid endpoints: %+v", relationship)
		}
	}
	for _, relationship := range generated.PersonalRelationships {
		if entityByID[relationship.PersonANPCID].EntityType != "npc" || entityByID[relationship.PersonBNPCID].EntityType != "npc" || relationship.PersonANPCID >= relationship.PersonBNPCID {
			t.Fatalf("personal relationship has invalid or non-canonical endpoints: %+v", relationship)
		}
	}
	for _, relationship := range generated.CompanyLocations {
		if entityByID[relationship.CompanyID].EntityType != "company" || entityByID[relationship.LocationID].LocationKind != "building" {
			t.Fatalf("company location has invalid endpoints: %+v", relationship)
		}
	}
}

func TestGeneratedProfilesReconstructFromGenerationKeys(t *testing.T) {
	generated, err := Generate("m03-profile-seed", GeneratorVersion2)
	if err != nil {
		t.Fatalf("Generate() error = %v", err)
	}
	for _, expected := range generated.NPCs {
		actual, err := GenerateNPC("m03-profile-seed", GeneratorVersion2, expected.GenerationKey)
		if err != nil {
			t.Fatalf("GenerateNPC() error = %v", err)
		}
		if !reflect.DeepEqual(expected, actual) {
			t.Fatalf("NPC reconstruction mismatch: expected %+v, got %+v", expected, actual)
		}
	}
	for _, expected := range generated.Companies {
		actual, err := GenerateCompany("m03-profile-seed", GeneratorVersion2, expected.GenerationKey)
		if err != nil {
			t.Fatalf("GenerateCompany() error = %v", err)
		}
		if !reflect.DeepEqual(expected, actual) {
			t.Fatalf("company reconstruction mismatch: expected %+v, got %+v", expected, actual)
		}
	}
}

func TestDifferentSeedsUseDisjointNPCIDs(t *testing.T) {
	first, err := Generate("m03-world-one", GeneratorVersion2)
	if err != nil {
		t.Fatalf("Generate() first world error = %v", err)
	}
	second, err := Generate("m03-world-two", GeneratorVersion2)
	if err != nil {
		t.Fatalf("Generate() second world error = %v", err)
	}
	firstIDs := make(map[string]struct{}, len(first.NPCs))
	for _, npc := range first.NPCs {
		firstIDs[npc.ID] = struct{}{}
	}
	for _, npc := range second.NPCs {
		if _, exists := firstIDs[npc.ID]; exists {
			t.Fatalf("NPC ID %q appears in worlds with different seeds", npc.ID)
		}
	}
}

func TestProfileReconstructionRejectsForeignKeys(t *testing.T) {
	if _, err := GenerateNPC("m03-seed", GeneratorVersion2, "lyon/people/foreign/district/bellecour/person/01"); err == nil {
		t.Fatal("GenerateNPC() accepted a generation key from another seed")
	}
	if _, err := GenerateCompany("m03-seed", GeneratorVersion2, "lyon/companies/foreign/district/bellecour/company"); err == nil {
		t.Fatal("GenerateCompany() accepted a generation key from another seed")
	}
}

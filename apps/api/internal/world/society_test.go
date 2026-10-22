package world

import (
	"fmt"
	"reflect"
	"strings"
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

func TestGeneratorVersion5CreatesTwentyDeterministicDistricts(t *testing.T) {
	const seed = "m10-twenty-districts"
	first, err := Generate(seed, GeneratorVersion5)
	if err != nil {
		t.Fatalf("Generate() error = %v", err)
	}
	second, err := Generate(seed, GeneratorVersion5)
	if err != nil {
		t.Fatalf("Generate() second call error = %v", err)
	}
	if !reflect.DeepEqual(first, second) {
		t.Fatal("same seed and version produced different 20-district worlds")
	}
	if first.GeneratorVersion != GeneratorVersion5 {
		t.Fatalf("generator version = %d, want %d", first.GeneratorVersion, GeneratorVersion5)
	}
	var districts int
	names := map[string]struct{}{}
	for _, region := range first.Regions {
		if region.Kind != "district" {
			continue
		}
		districts++
		if _, exists := names[region.Name]; exists {
			t.Fatalf("duplicate generated district name %q", region.Name)
		}
		names[region.Name] = struct{}{}
	}
	locations := 0
	for _, entity := range first.Entities {
		if entity.EntityType == "location" {
			locations++
		}
	}
	if districts != 20 || len(first.NPCs) != 160 || len(first.Companies) != 20 || locations != 80 {
		t.Fatalf("world contains %d districts, %d NPCs, %d companies and %d locations; want 20, 160, 20 and 80", districts, len(first.NPCs), len(first.Companies), locations)
	}
	for _, npc := range first.NPCs {
		if npc.Visual == nil {
			t.Fatalf("version 5 NPC %q has no visual identity", npc.ID)
		}
		if err := ValidateVisualIdentity(*npc.Visual); err != nil {
			t.Fatalf("NPC %q has invalid visual identity %+v: %v", npc.ID, *npc.Visual, err)
		}
		rebuilt, err := GenerateNPC(seed, GeneratorVersion5, npc.GenerationKey)
		if err != nil {
			t.Fatalf("GenerateNPC() reconstruction error = %v", err)
		}
		if !reflect.DeepEqual(npc, rebuilt) {
			t.Fatalf("NPC %q changed when reconstructed from its generation key", npc.ID)
		}
	}
}

func TestGenerateCityCreatesAnIndependentDeterministicPopulation(t *testing.T) {
	const seed = "m11-city-world"
	first, err := GenerateCity(seed, GeneratorVersion5, "jp/tokyo")
	if err != nil {
		t.Fatalf("GenerateCity() error = %v", err)
	}
	second, err := GenerateCity(seed, GeneratorVersion5, "jp/tokyo")
	if err != nil {
		t.Fatalf("GenerateCity() second call error = %v", err)
	}
	if !reflect.DeepEqual(first, second) {
		t.Fatal("same seed, city and generator version produced different city worlds")
	}
	if len(first.NPCs) != 160 || len(first.Companies) != 20 || len(first.Regions) != 21 || len(first.Entities) != 260 {
		t.Fatalf("Tokyo world shape: %d NPCs, %d companies, %d regions, %d entities", len(first.NPCs), len(first.Companies), len(first.Regions), len(first.Entities))
	}
	if first.Regions[0].ID == "fr/lyon" {
		t.Fatal("Tokyo generation retained Lyon's city identity")
	}
	for _, npc := range first.NPCs {
		rebuilt, err := GenerateNPC(seed, GeneratorVersion5, npc.GenerationKey)
		if err != nil {
			t.Fatalf("reconstruct NPC %q: %v", npc.ID, err)
		}
		if !reflect.DeepEqual(npc, rebuilt) {
			t.Fatalf("NPC %q changed when reconstructed", npc.ID)
		}
	}
	for _, company := range first.Companies {
		rebuilt, err := GenerateCompany(seed, GeneratorVersion5, company.GenerationKey)
		if err != nil {
			t.Fatalf("reconstruct company %q: %v", company.ID, err)
		}
		if company != rebuilt {
			t.Fatalf("company %q changed when reconstructed", company.ID)
		}
	}
	lyon, err := GenerateCity(seed, GeneratorVersion5, "fr/lyon")
	if err != nil {
		t.Fatal(err)
	}
	if reflect.DeepEqual(first.NPCs[0].ID, lyon.NPCs[0].ID) {
		t.Fatal("NPC IDs were reused between cities")
	}
}

func TestVersion5HairChoicesRespectAgeAndNPCStyleCatalog(t *testing.T) {
	for _, age := range []int{18, 25, 39} {
		for sample := 0; sample < 100; sample++ {
			stream := newStream("m10-age-hair", GeneratorVersion5, fmt.Sprintf("sample-%d", sample))
			hairID := generatedHairID(stream, age)
			if hairTone(hairID) == "grey" || strings.Contains(hairID, "balding") || strings.Contains(hairID, "sparse_sides") || strings.Contains(hairID, "pigtails") {
				t.Fatalf("age %d received implausible generated hair %q", age, hairID)
			}
			if !containsLayer(hairLayerIDs, hairID) {
				t.Fatalf("generated hair %q is missing from the visual catalog", hairID)
			}
		}
	}
	for _, age := range []int{40, 44} {
		for sample := 0; sample < 100; sample++ {
			stream := newStream("m10-middle-age-hair", GeneratorVersion5, fmt.Sprintf("sample-%d", sample))
			hairID := generatedHairID(stream, age)
			if strings.Contains(hairID, "balding") || strings.Contains(hairID, "sparse_sides") {
				t.Fatalf("age %d received age-restricted hair %q", age, hairID)
			}
		}
	}
}

func TestGeneratorVersion4VisualSnapshotRemainsReconstructible(t *testing.T) {
	generated, err := Generate("m10-v4-visual-snapshot", GeneratorVersion4)
	if err != nil {
		t.Fatal(err)
	}
	npc := generated.NPCs[0]
	if npc.Name != "Noëlle Delmas" || npc.Age != 30 {
		t.Fatalf("version 4 NPC snapshot changed: name=%q age=%d", npc.Name, npc.Age)
	}
	wantVisual := VisualIdentity{HairID: "hair_balding_blonde", SkinID: "skin_light", EyesID: "eyes_amber", AccessoryIDs: []string{}}
	if !reflect.DeepEqual(*npc.Visual, wantVisual) {
		t.Fatalf("version 4 visual identity = %+v, want %+v", *npc.Visual, wantVisual)
	}
	rebuilt, err := GenerateNPC("m10-v4-visual-snapshot", GeneratorVersion4, npc.GenerationKey)
	if err != nil {
		t.Fatalf("GenerateNPC() version 4 reconstruction error = %v", err)
	}
	if !reflect.DeepEqual(npc, rebuilt) {
		t.Fatal("version 4 NPC changed when reconstructed from its generation key")
	}
}

package world

import (
	"bytes"
	"encoding/json"
	"os"
	"os/exec"
	"reflect"
	"strings"
	"testing"
)

func TestGenerateIsDeterministic(t *testing.T) {
	first, err := Generate("m02-fixture-seed", GeneratorVersion1)
	if err != nil {
		t.Fatalf("Generate() error = %v", err)
	}
	second, err := Generate("m02-fixture-seed", GeneratorVersion1)
	if err != nil {
		t.Fatalf("Generate() second call error = %v", err)
	}
	if !reflect.DeepEqual(first, second) {
		t.Fatal("same seed and generator version produced different worlds")
	}
}

func TestGenerateChangesCityForDifferentSeeds(t *testing.T) {
	first, err := Generate("m02-seed-one", GeneratorVersion1)
	if err != nil {
		t.Fatalf("Generate() first seed error = %v", err)
	}
	second, err := Generate("m02-seed-two", GeneratorVersion1)
	if err != nil {
		t.Fatalf("Generate() second seed error = %v", err)
	}
	if reflect.DeepEqual(first, second) {
		t.Fatal("different seeds produced identical city slices")
	}
}

func TestGenerateVersionOneProducesMinimumCity(t *testing.T) {
	generated, err := Generate("m02-city-fixture", GeneratorVersion1)
	if err != nil {
		t.Fatalf("Generate() error = %v", err)
	}
	if generated.GeneratorVersion != GeneratorVersion1 {
		t.Fatalf("generator version = %d, want %d", generated.GeneratorVersion, GeneratorVersion1)
	}
	if len(generated.Regions) != 4 {
		t.Fatalf("region count = %d, want city plus three districts", len(generated.Regions))
	}
	if len(generated.Entities) != 12 {
		t.Fatalf("entity count = %d, want two streets and two buildings per district", len(generated.Entities))
	}
}

func TestGenerateReturnsStableUniqueIDs(t *testing.T) {
	generated, err := Generate("m02-id-fixture", GeneratorVersion1)
	if err != nil {
		t.Fatalf("Generate() error = %v", err)
	}
	regionIDs := make(map[string]struct{}, len(generated.Regions))
	for _, region := range generated.Regions {
		if _, exists := regionIDs[region.ID]; exists {
			t.Fatalf("duplicate region ID %q", region.ID)
		}
		regionIDs[region.ID] = struct{}{}
	}
	entityIDs := make(map[string]struct{}, len(generated.Entities))
	for _, entity := range generated.Entities {
		if _, exists := entityIDs[entity.ID]; exists {
			t.Fatalf("duplicate entity ID %q", entity.ID)
		}
		entityIDs[entity.ID] = struct{}{}
		if _, exists := regionIDs[entity.RegionID]; !exists {
			t.Fatalf("entity %q refers to missing region %q", entity.ID, entity.RegionID)
		}
		if entity.ID != "location:"+entity.GenerationKey {
			t.Fatalf("entity ID %q is not derived from generation key %q", entity.ID, entity.GenerationKey)
		}
	}
}

func TestGenerateSurvivesProcessRestart(t *testing.T) {
	if os.Getenv("NOIR_WORLD_GENERATOR_CHILD") == "1" {
		generated, err := Generate("m02-restart-fixture", GeneratorVersion1)
		if err != nil {
			t.Fatalf("Generate() in child process error = %v", err)
		}
		if err := json.NewEncoder(os.Stdout).Encode(generated); err != nil {
			t.Fatalf("encode generated world: %v", err)
		}
		return
	}

	generateInFreshProcess := func() []byte {
		t.Helper()
		command := exec.Command(os.Args[0], "-test.run=^TestGenerateSurvivesProcessRestart$")
		command.Env = append(os.Environ(), "NOIR_WORLD_GENERATOR_CHILD=1")
		output, err := command.CombinedOutput()
		if err != nil {
			t.Fatalf("generator subprocess failed: %v\n%s", err, output)
		}
		line, _, found := bytes.Cut(output, []byte("\n"))
		if !found || !json.Valid(line) {
			t.Fatalf("generator subprocess returned invalid JSON: %s", output)
		}
		return line
	}

	first := generateInFreshProcess()
	second := generateInFreshProcess()
	if !bytes.Equal(first, second) {
		t.Fatal("same seed and generator version changed after process restart")
	}
}

func TestGenerateRejectsInvalidSeedAndVersion(t *testing.T) {
	if _, err := Generate(" \t", GeneratorVersion1); err == nil || !strings.Contains(err.Error(), "seed") {
		t.Fatalf("Generate() empty seed error = %v, want seed error", err)
	}
	if _, err := Generate("valid-seed", 0); err == nil || !strings.Contains(err.Error(), ErrUnsupportedGeneratorVersion.Error()) {
		t.Fatalf("Generate() unsupported version error = %v", err)
	}
}

package cases

import (
	"errors"
	"reflect"
	"testing"

	"noir-1974/apps/api/internal/world"
)

func TestGenerateHomicideIsDeterministicAndUsesWorldRelationships(t *testing.T) {
	base, err := world.Generate("m05-world-fixture", world.GeneratorVersion2)
	if err != nil {
		t.Fatalf("world.Generate() error = %v", err)
	}
	first, err := Generate(base, "case-seed-one", GeneratorVersion)
	if err != nil {
		t.Fatalf("Generate() error = %v", err)
	}
	second, err := Generate(base, "case-seed-one", GeneratorVersion)
	if err != nil {
		t.Fatalf("Generate() second call error = %v", err)
	}
	if !reflect.DeepEqual(first, second) {
		t.Fatalf("same seed and version produced different definitions:\nfirst:  %+v\nsecond: %+v", first, second)
	}
	if first.CaseType != "homicide" || first.KillerID == first.VictimID {
		t.Fatalf("generated case has invalid participants or type: %+v", first)
	}
	if err := Validate(first, base); err != nil {
		t.Fatalf("Validate() generated definition error = %v", err)
	}
	if len(first.Relationships) == 0 || len(first.Evidence) == 0 || first.ArchetypeID == "" || first.Motive == "" {
		t.Fatalf("case omitted its archetype, motive, social links or evidence: %+v", first)
	}
	if first.Evidence[0].LocationID != first.LocationID {
		t.Fatalf("evidence location %q does not match case location %q", first.Evidence[0].LocationID, first.LocationID)
	}
	for _, reference := range []string{first.KillerID, first.VictimID, first.LocationID} {
		if reference == "" {
			t.Fatal("case contains an empty entity reference")
		}
	}
	other, err := Generate(base, "case-seed-two", GeneratorVersion)
	if err != nil {
		t.Fatalf("Generate() with another seed error = %v", err)
	}
	if first.ID == other.ID {
		t.Fatal("different case seeds produced the same case ID")
	}
}

func TestValidateRejectsUnsolvableOrInvalidCases(t *testing.T) {
	base, err := world.Generate("m05-validation-fixture", world.GeneratorVersion2)
	if err != nil {
		t.Fatalf("world.Generate() error = %v", err)
	}
	definition, err := Generate(base, "case-validation", GeneratorVersion)
	if err != nil {
		t.Fatalf("Generate() error = %v", err)
	}

	t.Run("same victim and killer", func(t *testing.T) {
		invalid := cloneDefinition(definition)
		invalid.KillerID = invalid.VictimID
		if err := Validate(invalid, base); !errors.Is(err, ErrInvalidDefinition) {
			t.Fatalf("Validate() error = %v, want ErrInvalidDefinition", err)
		}
	})

	t.Run("missing entity", func(t *testing.T) {
		invalid := cloneDefinition(definition)
		invalid.VictimID = "npc:missing"
		if err := Validate(invalid, base); !errors.Is(err, ErrInvalidDefinition) {
			t.Fatalf("Validate() error = %v, want ErrInvalidDefinition", err)
		}
	})

	t.Run("no evidence", func(t *testing.T) {
		invalid := cloneDefinition(definition)
		invalid.Evidence = nil
		if err := Validate(invalid, base); !errors.Is(err, ErrUnsolvableCase) {
			t.Fatalf("Validate() error = %v, want ErrUnsolvableCase", err)
		}
	})

	t.Run("evidence cannot be reached from either person", func(t *testing.T) {
		invalid := cloneDefinition(definition)
		invalid.Evidence[0].ReachableFromEntityIDs = []string{base.Companies[0].ID}
		if err := Validate(invalid, base); !errors.Is(err, ErrUnsolvableCase) {
			t.Fatalf("Validate() error = %v, want ErrUnsolvableCase", err)
		}
	})

	t.Run("missing required world relationship", func(t *testing.T) {
		invalid := cloneDefinition(definition)
		invalid.Relationships = nil
		if err := Validate(invalid, base); !errors.Is(err, ErrInvalidDefinition) {
			t.Fatalf("Validate() error = %v, want ErrInvalidDefinition", err)
		}
	})

	t.Run("wrong crime location", func(t *testing.T) {
		invalid := cloneDefinition(definition)
		for _, entity := range base.Entities {
			if entity.EntityType == "location" && entity.ID != definition.LocationID {
				invalid.LocationID = entity.ID
				break
			}
		}
		if err := Validate(invalid, base); !errors.Is(err, ErrInvalidDefinition) {
			t.Fatalf("Validate() error = %v, want ErrInvalidDefinition", err)
		}
	})
}

func TestGenerateRejectsInvalidSeedVersionAndWorld(t *testing.T) {
	base, err := world.Generate("m05-empty-fixture", world.GeneratorVersion2)
	if err != nil {
		t.Fatalf("world.Generate() error = %v", err)
	}
	if _, err := Generate(base, "  ", GeneratorVersion); !errors.Is(err, ErrInvalidCaseSeed) {
		t.Fatalf("empty seed error = %v, want ErrInvalidCaseSeed", err)
	}
	if _, err := Generate(base, "seed", GeneratorVersion+1); !errors.Is(err, ErrUnsupportedGenerator) {
		t.Fatalf("unsupported version error = %v, want ErrUnsupportedGenerator", err)
	}
	withoutSociety, err := world.Generate("m05-geography-only", world.GeneratorVersion1)
	if err != nil {
		t.Fatalf("world.Generate() v1 error = %v", err)
	}
	if _, err := Generate(withoutSociety, "seed", GeneratorVersion); !errors.Is(err, ErrNoValidCaseCandidates) {
		t.Fatalf("world without society error = %v, want ErrNoValidCaseCandidates", err)
	}
}

func cloneDefinition(definition Definition) Definition {
	copy := definition
	copy.Relationships = append([]RelationshipReference(nil), definition.Relationships...)
	copy.Evidence = make([]EvidenceDefinition, len(definition.Evidence))
	for index, evidence := range definition.Evidence {
		copy.Evidence[index] = evidence
		copy.Evidence[index].RelationshipIDs = append([]string(nil), evidence.RelationshipIDs...)
		copy.Evidence[index].ReachableFromEntityIDs = append([]string(nil), evidence.ReachableFromEntityIDs...)
	}
	return copy
}

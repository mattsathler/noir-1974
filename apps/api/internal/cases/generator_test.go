package cases

import (
	"errors"
	"fmt"
	"reflect"
	"testing"

	"noir-1974/apps/api/internal/records"
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

func TestFinancialExtortionIncludesMixedBankAndGATECaseRecords(t *testing.T) {
	base, err := world.Generate("m07-world-fixture", world.GeneratorVersion2)
	if err != nil {
		t.Fatalf("world.Generate() error = %v", err)
	}
	var definition Definition
	for index := 0; index < 200; index++ {
		definition, err = Generate(base, fmt.Sprintf("m07-case-seed-%d", index), GeneratorVersion)
		if err != nil {
			t.Fatalf("Generate() error = %v", err)
		}
		if definition.ArchetypeID == "financial_extortion" {
			break
		}
	}
	if definition.ArchetypeID != "financial_extortion" {
		t.Fatal("no deterministic seed selected the financial extortion archetype")
	}
	if err := Validate(definition, base); err != nil {
		t.Fatalf("Validate() error = %v", err)
	}
	var victimPayment, killerReceipt records.Record
	var sceneVictim, sceneKiller records.Record
	for _, item := range definition.Records {
		switch item.InternalRelation {
		case "extortion_payment":
			if item.NPCID == definition.VictimID {
				victimPayment = item
			}
			if item.NPCID == definition.KillerID {
				killerReceipt = item
			}
		case "scene-victim":
			sceneVictim = item
		case "scene-killer":
			sceneKiller = item
		}
	}
	if victimPayment.Source != "bank" || victimPayment.Direction != "outgoing" || victimPayment.NPCID != definition.VictimID {
		t.Fatalf("extortion debit is not linked to the victim: %+v", victimPayment)
	}
	if killerReceipt.Source != "bank" || killerReceipt.Direction != "incoming" || killerReceipt.NPCID != definition.KillerID {
		t.Fatalf("matching extortion credit is not linked to the killer: %+v", killerReceipt)
	}
	if victimPayment.Amount != killerReceipt.Amount || victimPayment.Reference == "" || victimPayment.Reference != killerReceipt.Reference || !victimPayment.OccurredAt.Equal(killerReceipt.OccurredAt) {
		t.Fatalf("sender and recipient transaction entries do not match: debit=%+v credit=%+v", victimPayment, killerReceipt)
	}
	if sceneVictim.Destination == "" || sceneVictim.Destination != sceneKiller.Destination || !sceneVictim.OccurredAt.Before(definition.CrimeAt) || !sceneKiller.OccurredAt.Before(definition.CrimeAt) {
		t.Fatalf("victim and killer GATE movements do not converge at the scene before the crime: victim=%+v killer=%+v crime=%s", sceneVictim, sceneKiller, definition.CrimeAt)
	}
	if !victimPayment.OccurredAt.Before(sceneKiller.OccurredAt) {
		t.Fatal("extortion transfer should precede the investigated GATE movements")
	}
}

func TestJealousyCaseConnectsSpousesParamourAndInvestigationChannels(t *testing.T) {
	base, err := world.Generate("m13-jealousy-v4", world.GeneratorVersion2)
	if err != nil {
		t.Fatalf("world.Generate() error = %v", err)
	}
	var definition Definition
	for index := 0; index < 300; index++ {
		definition, err = Generate(base, fmt.Sprintf("jealousy-v4-%d", index), GeneratorVersion)
		if err != nil {
			t.Fatalf("Generate() error = %v", err)
		}
		if definition.ArchetypeID == "jealousy" {
			break
		}
	}
	if definition.ArchetypeID != "jealousy" {
		t.Fatal("no seed selected the jealousy archetype")
	}
	if definition.ParamourID == "" || definition.ParamourID == definition.KillerID || definition.ParamourID == definition.VictimID {
		t.Fatalf("jealousy case has no distinct paramour: %+v", definition)
	}
	if err := Validate(definition, base); err != nil {
		t.Fatalf("Validate() error = %v", err)
	}
	var affairVictim, affairParamour bool
	for _, record := range definition.Records {
		if record.InternalRelation == "affair-victim" && record.NPCID == definition.VictimID {
			affairVictim = true
		}
		if record.InternalRelation == "affair-paramour" && record.NPCID == definition.ParamourID {
			affairParamour = true
		}
	}
	if !affairVictim || !affairParamour {
		t.Fatal("jealousy case omitted synchronized GATE passages for the affair pair")
	}
	var affairMail, remorseMail bool
	for _, message := range definition.Emails {
		if message.InternalRelation == "jealousy_exchange" && message.MailboxNPCID == definition.VictimID && message.RelatedNPCID == definition.ParamourID {
			affairMail = true
		}
		if message.InternalRelation == "killer_remorse" && message.MailboxNPCID == definition.KillerID {
			remorseMail = true
		}
	}
	if !affairMail || !remorseMail {
		t.Fatalf("jealousy case omitted affair or killer correspondence: %+v", definition.Emails)
	}
}

func TestLegacyCaseGeneratorVersionRemainsAvailable(t *testing.T) {
	base, err := world.Generate("legacy-case-v3", world.GeneratorVersion2)
	if err != nil {
		t.Fatalf("world.Generate() error = %v", err)
	}
	definition, err := Generate(base, "legacy-case-seed", legacyGeneratorVersion)
	if err != nil {
		t.Fatalf("legacy Generate() error = %v", err)
	}
	if definition.GeneratorVersion != legacyGeneratorVersion || !definition.CrimeAt.IsZero() || definition.ParamourID != "" {
		t.Fatalf("legacy case changed its v3 shape: %+v", definition)
	}
	if err := Validate(definition, base); err != nil {
		t.Fatalf("legacy Validate() error = %v", err)
	}
}

func TestAnonymousCaseMessagesAreRareAndRetainSenderIdentityInSourceData(t *testing.T) {
	base, err := world.Generate("m13-anonymous-mail-v4", world.GeneratorVersion2)
	if err != nil {
		t.Fatalf("world.Generate() error = %v", err)
	}
	anonymousCount := 0
	for index := 0; index < 100; index++ {
		definition, err := Generate(base, fmt.Sprintf("anonymous-mail-case-%d", index), GeneratorVersion)
		if err != nil {
			t.Fatalf("Generate() error = %v", err)
		}
		for _, message := range definition.Emails {
			if message.InternalRelation != "anonymous_warning" {
				continue
			}
			anonymousCount++
			if message.Sender != npcByID(base.NPCs, definition.KillerID).Name || message.MailboxNPCID != definition.VictimID || message.RelatedNPCID != definition.KillerID {
				t.Fatalf("anonymous message did not retain its true sender and participants: %+v", message)
			}
		}
	}
	if anonymousCount == 0 || anonymousCount == 100 {
		t.Fatalf("anonymous messages appeared %d/100 times; expected a rare, recurring subset", anonymousCount)
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

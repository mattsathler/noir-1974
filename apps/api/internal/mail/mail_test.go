package mail

import (
	"fmt"
	"strings"
	"testing"
	"time"
)

func TestRoleCluesAreOccasionalAndDoNotStateTheTitle(t *testing.T) {
	clues := 0
	for index := 0; index < 128; index++ {
		rng := newStream("role-hint-seed", 1, fmt.Sprintf("role-hint/%d", index), time.Date(1974, time.November, 1, 0, 0, 0, 0, time.UTC))
		item := addSubtleRoleClue(messageTemplate{Subject: "Office notice", Body: "Please return the revised file."}, "accountant", &rng)
		if item.Body != "Please return the revised file." {
			clues++
		}
		if strings.Contains(strings.ToLower(item.Body), "accountant") || strings.Contains(strings.ToLower(item.Body), "as manager") {
			t.Fatalf("role clue states a job title explicitly: %q", item.Body)
		}
	}
	if clues == 0 || clues > 32 {
		t.Fatalf("role clue count = %d of 128, want occasional clues", clues)
	}
}

func TestTemplateCatalogContainsFourHundredNewTemplatesPerType(t *testing.T) {
	catalog, err := loadTemplateCatalog()
	if err != nil {
		t.Fatal(err)
	}
	for name, group := range map[string][]messageTemplate{
		"personal":  catalog.Personal,
		"workplace": catalog.Workplace,
		"colleague": catalog.Colleague,
	} {
		if len(group) < 400 {
			t.Errorf("%s templates = %d, want at least 400", name, len(group))
		}
		seen := make(map[string]struct{}, len(group))
		for _, item := range group {
			key := item.Subject + "\x00" + item.Body
			if _, exists := seen[key]; exists {
				t.Errorf("%s template is duplicated: %q", name, item.Subject)
			}
			seen[key] = struct{}{}
		}
	}
}

func TestPairedCorrespondenceIsIdenticalInBothMailboxes(t *testing.T) {
	from := time.Date(1974, time.November, 1, 0, 0, 0, 0, time.UTC)
	for _, test := range []struct {
		name     string
		relation string
		contactA Correspondent
		contactB Correspondent
		roleA    string
		roleB    string
	}{
		{
			name: "partners", relation: "personal",
			contactA: Correspondent{ID: "npc:b", Name: "Beatrice", RelationshipType: "marriage"},
			contactB: Correspondent{ID: "npc:a", Name: "Adrien", RelationshipType: "marriage"},
		},
		{
			name: "colleagues", relation: "colleague",
			contactA: Correspondent{ID: "npc:b", Name: "Beatrice", Role: "clerk", CompanyID: "company:1"},
			contactB: Correspondent{ID: "npc:a", Name: "Adrien", Role: "manager", CompanyID: "company:1"},
			roleA:    "manager", roleB: "clerk",
		},
	} {
		t.Run(test.name, func(t *testing.T) {
			found := 0
			for offset := 0; offset < 200; offset += MaxRangeDays {
				start := from.AddDate(0, 0, offset)
				end := start.AddDate(0, 0, MaxRangeDays-1)
				messagesA, err := Generate("shared-seed", 1, "npc:a", Context{NPCName: "Adrien", EmploymentRole: test.roleA, Personal: contactsFor("personal", test.relation, test.contactA), Colleagues: contactsFor("colleague", test.relation, test.contactA)}, start, end)
				if err != nil {
					t.Fatal(err)
				}
				messagesB, err := Generate("shared-seed", 1, "npc:b", Context{NPCName: "Beatrice", EmploymentRole: test.roleB, Personal: contactsFor("personal", test.relation, test.contactB), Colleagues: contactsFor("colleague", test.relation, test.contactB)}, start, end)
				if err != nil {
					t.Fatal(err)
				}
				pairedA := pairedMessages(messagesA, test.relation)
				pairedB := pairedMessages(messagesB, test.relation)
				for id, messageA := range pairedA {
					messageB, ok := pairedB[id]
					if !ok {
						t.Fatalf("message %s is missing from the other mailbox", id)
					}
					if messageA.Sender != messageB.Sender || messageA.Recipient != messageB.Recipient || messageA.Subject != messageB.Subject || messageA.Body != messageB.Body || !messageA.SentAt.Equal(messageB.SentAt) {
						t.Fatalf("message differs between mailboxes: A=%+v B=%+v", messageA, messageB)
					}
					found++
				}
				for id := range pairedB {
					if _, ok := pairedA[id]; !ok {
						t.Fatalf("message %s is missing from the other mailbox", id)
					}
				}
			}
			if found == 0 {
				t.Fatal("no paired messages were generated in the deterministic sample")
			}
		})
	}
}

func contactsFor(kind, relation string, contact Correspondent) []Correspondent {
	if kind == relation {
		return []Correspondent{contact}
	}
	return nil
}

func pairedMessages(messages []Message, relation string) map[string]Message {
	result := make(map[string]Message)
	for _, message := range messages {
		if message.InternalRelation == relation {
			result[message.ID] = message
		}
	}
	return result
}

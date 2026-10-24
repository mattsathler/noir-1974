// Package mail generates mailbox-scoped correspondence and keeps case links
// as backend-only metadata.
package mail

import (
	"crypto/sha256"
	_ "embed"
	"encoding/hex"
	"encoding/json"
	"errors"
	"fmt"
	"sort"
	"strings"
	"sync"
	"time"
)

const GeneratorVersion = 4
const MaxRangeDays = 7

var (
	ErrInvalidQuery       = errors.New("mail query requires an NPC and a valid date range")
	ErrUnsupportedVersion = errors.New("unsupported mail generator version")
	ErrRangeTooLarge      = errors.New("mail query cannot exceed 7 calendar days")
	ErrInvalidTemplates   = errors.New("mail template catalog is invalid")
)

// Message is a player-visible message. Mailbox and case relationship fields
// are authoritative backend metadata and are deliberately omitted from JSON.
type Message struct {
	ID        string    `json:"id"`
	SentAt    time.Time `json:"sentAt"`
	Service   string    `json:"service"`
	Sender    string    `json:"sender"`
	Recipient string    `json:"recipient"`
	Subject   string    `json:"subject"`
	Body      string    `json:"body"`

	MailboxNPCID     string `json:"-"`
	RelatedNPCID     string `json:"-"`
	CaseID           string `json:"-"`
	InternalRelation string `json:"-"`
}

type Context struct {
	NPCName        string
	Employer       string
	EmploymentRole string
	DeathAt        time.Time
	Personal       []Correspondent
	Colleagues     []Correspondent
}

type Correspondent struct {
	ID               string
	Name             string
	Status           string
	DeathAt          time.Time
	Role             string
	RelationshipType string
	CompanyID        string
}

type messageTemplate struct {
	Subject string `json:"subject"`
	Body    string `json:"body"`
}

type templateCatalog struct {
	Personal  []messageTemplate `json:"personal"`
	Workplace []messageTemplate `json:"workplace"`
	Colleague []messageTemplate `json:"colleague"`
}

//go:embed templates.json
var templateData []byte

var (
	catalogOnce sync.Once
	catalog     templateCatalog
	catalogErr  error
)

func loadTemplateCatalog() (*templateCatalog, error) {
	catalogOnce.Do(func() {
		if err := json.Unmarshal(templateData, &catalog); err != nil {
			catalogErr = fmt.Errorf("%w: %v", ErrInvalidTemplates, err)
			return
		}
		if len(catalog.Personal) == 0 || len(catalog.Workplace) == 0 || len(catalog.Colleague) == 0 {
			catalogErr = ErrInvalidTemplates
			return
		}
		for _, group := range [][]messageTemplate{catalog.Personal, catalog.Workplace, catalog.Colleague} {
			for _, item := range group {
				if strings.TrimSpace(item.Subject) == "" || strings.TrimSpace(item.Body) == "" {
					catalogErr = ErrInvalidTemplates
					return
				}
			}
		}
	})
	if catalogErr != nil {
		return nil, catalogErr
	}
	return &catalog, nil
}

// Generate builds routine mail only for one mailbox and an inclusive interval.
// It does not enumerate other residents or materialize a global message store.
func Generate(seed string, worldVersion int, npcID string, context Context, from, through time.Time) ([]Message, error) {
	if seed == "" || npcID == "" || from.IsZero() || through.IsZero() || through.Before(from) {
		return nil, ErrInvalidQuery
	}
	if worldVersion <= 0 {
		return nil, ErrUnsupportedVersion
	}
	templates, err := loadTemplateCatalog()
	if err != nil {
		return nil, err
	}
	start, end := utcDate(from), utcDate(through)
	if end.Sub(start) > (MaxRangeDays-1)*24*time.Hour {
		return nil, ErrRangeTooLarge
	}
	if context.NPCName == "" {
		context.NPCName = "Account holder"
	}
	result := make([]Message, 0)
	for date := start; !date.After(end); date = date.AddDate(0, 0, 1) {
		rng := newStream(seed, worldVersion, npcID, date)
		if context.Employer != "" && rng.pick(2) == 0 {
			item := templates.Workplace[rng.pick(len(templates.Workplace))]
			item = personalize(item, context, &rng)
			sender := context.Employer
			if sender == "" {
				sender = "Meridian Global Correspondence Office"
			}
			result = append(result, Message{
				ID:      stableID(seed, worldVersion, npcID, date),
				SentAt:  date.Add(time.Duration(7*60+rng.pick(13*60)) * time.Minute),
				Service: "Meridian Global Correspondence Exchange",
				Sender:  sender, Recipient: context.NPCName, Subject: item.Subject, Body: item.Body,
				MailboxNPCID: npcID,
			})
		}
		for _, correspondent := range context.Personal {
			if message, ok := correspondence(seed, worldVersion, npcID, context.NPCName, context.EmploymentRole, correspondent, date, "personal", templates); ok {
				if beforeDeathDeadline(message.SentAt, correspondent.DeathAt) {
					result = append(result, message)
				}
			}
		}
		for _, colleague := range context.Colleagues {
			if message, ok := correspondence(seed, worldVersion, npcID, context.NPCName, context.EmploymentRole, colleague, date, "colleague", templates); ok {
				if beforeDeathDeadline(message.SentAt, colleague.DeathAt) {
					result = append(result, message)
				}
			}
		}
	}
	if !context.DeathAt.IsZero() {
		available := result[:0]
		for _, message := range result {
			if beforeDeathDeadline(message.SentAt, context.DeathAt) {
				available = append(available, message)
			}
		}
		result = available
	}
	SortMessages(result)
	return result, nil
}

func beforeDeathDeadline(sentAt, deathAt time.Time) bool {
	return deathAt.IsZero() || !sentAt.After(deathAt)
}

// correspondence generates a single shared message for both participants' mailboxes.
func correspondence(seed string, worldVersion int, mailboxID, mailboxName, mailboxRole string, contact Correspondent, date time.Time, kind string, templates *templateCatalog) (Message, bool) {
	if contact.ID == "" || contact.Name == "" || contact.ID == mailboxID {
		return Message{}, false
	}
	pairA, pairB := mailboxID, contact.ID
	nameA, nameB := mailboxName, contact.Name
	roleA, roleB := mailboxRole, contact.Role
	if pairB < pairA {
		pairA, pairB = pairB, pairA
		nameA, nameB = nameB, nameA
		roleA, roleB = roleB, roleA
	}
	key := fmt.Sprintf("noir-mail-pair/v%d/%s/%d/%s/%s/%s/%s", GeneratorVersion, seed, worldVersion, kind, contact.CompanyID, pairA, pairB)
	rng := newKeyedStream(key, date)
	chance := 16
	if kind == "personal" {
		chance = 5
	}
	if rng.pick(chance) != 0 {
		return Message{}, false
	}
	item := templates.Colleague[rng.pick(len(templates.Colleague))]
	senderIndex := rng.pick(2)
	if kind == "personal" {
		item = templates.Personal[rng.pick(len(templates.Personal))]
		if contact.RelationshipType == "marriage" && rng.pick(2) == 0 {
			item.Body += " See you tonight, my love."
		}
	} else {
		role := roleA
		if senderIndex == 1 {
			role = roleB
		}
		item = addSubtleRoleClue(item, role, &rng)
	}
	sender, recipient := nameA, nameB
	if senderIndex == 1 {
		sender, recipient = nameB, nameA
	}
	return Message{
		ID:      stablePairID(key, date),
		SentAt:  date.Add(time.Duration(7*60+rng.pick(13*60)) * time.Minute),
		Service: "Meridian Global Correspondence Exchange",
		Sender:  sender, Recipient: recipient, Subject: item.Subject, Body: item.Body,
		MailboxNPCID: mailboxID, RelatedNPCID: contact.ID,
		InternalRelation: kind,
	}, true
}

func personalize(item messageTemplate, context Context, rng *stream) messageTemplate {
	return addSubtleRoleClue(item, context.EmploymentRole, rng)
}

func addSubtleRoleClue(item messageTemplate, role string, rng *stream) messageTemplate {
	if role == "" || rng.pick(8) != 0 {
		return item
	}
	clues := map[string][]string{
		"accountant":   {"Accounts will need your eye on the final column before it goes upstairs.", "Please give the figures your usual careful review before the file moves on."},
		"clerk":        {"The counter copy should be ready before the morning queue forms.", "Please make sure the register is in order before the next customer is called."},
		"driver":       {"The route sheet needs your check before the next vehicle leaves.", "Please confirm the delivery order before the afternoon run begins."},
		"manager":      {"The department is waiting for your decision before it sends a reply.", "Please settle the matter for your section before it goes upstairs."},
		"messenger":    {"The sealed packet must leave with the first afternoon dispatch.", "Please see that the signed envelope reaches the branch before closing."},
		"receptionist": {"The visitor book should be brought up to date before the doors close.", "Please leave the front desk register ready for the morning staff."},
		"secretary":    {"The clean minutes are needed before the director's next appointment.", "Please have the correspondence ready for the director's review."},
		"typist":       {"The final typescript is due at the registry before closing.", "Please prepare a clean copy for the next dispatch."},
	}
	options := clues[role]
	if len(options) == 0 {
		options = []string{"Your section's usual check is required before the file moves on."}
	}
	item.Body += " " + options[rng.pick(len(options))]
	return item
}

// CaseMessage creates a stable message for a case. Its internal link is never
// serialized to the client and is persisted separately by the case store.
func CaseMessage(seed string, worldVersion int, caseID, mailboxNPCID, relatedNPCID, sender, recipient, subject, body, relation string, sentAt time.Time) Message {
	key := fmt.Sprintf("noir-mail/v%d/%s/%d/%s/%s", GeneratorVersion, seed, worldVersion, caseID, relation)
	return Message{
		ID:               stableCaseID(key, mailboxNPCID),
		SentAt:           sentAt.UTC(),
		Service:          "Meridian Global Correspondence Exchange",
		Sender:           sender,
		Recipient:        recipient,
		Subject:          subject,
		Body:             body,
		MailboxNPCID:     mailboxNPCID,
		RelatedNPCID:     relatedNPCID,
		CaseID:           caseID,
		InternalRelation: relation,
	}
}

func stableCaseID(key, mailboxNPCID string) string {
	digest := sha256.Sum256([]byte(key + "/" + mailboxNPCID))
	return "mail:" + hex.EncodeToString(digest[:12])
}

func stableID(seed string, worldVersion int, npcID string, date time.Time) string {
	digest := sha256.Sum256([]byte(fmt.Sprintf("noir-mail/v%d/%s/%d/%s/%s", GeneratorVersion, seed, worldVersion, npcID, date.Format("2006-01-02"))))
	return "mail:" + hex.EncodeToString(digest[:12])
}

func stablePairID(key string, date time.Time) string {
	digest := sha256.Sum256([]byte(fmt.Sprintf("%s/%s", key, date.Format("2006-01-02"))))
	return "mail:" + hex.EncodeToString(digest[:12])
}

func SortMessages(messages []Message) {
	sort.Slice(messages, func(i, j int) bool {
		if messages[i].SentAt.Equal(messages[j].SentAt) {
			return messages[i].ID < messages[j].ID
		}
		return messages[i].SentAt.Before(messages[j].SentAt)
	})
}

func utcDate(value time.Time) time.Time {
	year, month, day := value.UTC().Date()
	return time.Date(year, month, day, 0, 0, 0, 0, time.UTC)
}

type stream struct {
	key     [sha256.Size]byte
	counter uint64
}

func newStream(seed string, worldVersion int, npcID string, date time.Time) stream {
	key := sha256.Sum256([]byte(fmt.Sprintf("noir-mail-stream/v%d/%s/%d/%s/%s", GeneratorVersion, seed, worldVersion, npcID, date.Format("2006-01-02"))))
	return stream{key: key}
}

func newKeyedStream(key string, date time.Time) stream {
	return stream{key: sha256.Sum256([]byte(fmt.Sprintf("%s/%s", key, date.Format("2006-01-02"))))}
}

func (rng *stream) pick(limit int) int {
	var input [sha256.Size + 8]byte
	copy(input[:sha256.Size], rng.key[:])
	for i := 0; i < 8; i++ {
		input[sha256.Size+i] = byte(rng.counter >> (56 - 8*i))
	}
	rng.counter++
	digest := sha256.Sum256(input[:])
	var value uint64
	for _, part := range digest[:8] {
		value = value<<8 | uint64(part)
	}
	return int(value % uint64(limit))
}

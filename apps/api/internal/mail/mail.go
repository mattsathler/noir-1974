// Package mail generates mailbox-scoped correspondence and keeps case links
// as backend-only metadata.
package mail

import (
	"crypto/sha256"
	"encoding/hex"
	"errors"
	"fmt"
	"sort"
	"time"
)

const GeneratorVersion = 1
const MaxRangeDays = 7

var (
	ErrInvalidQuery       = errors.New("mail query requires an NPC and a valid date range")
	ErrUnsupportedVersion = errors.New("unsupported mail generator version")
	ErrRangeTooLarge      = errors.New("mail query cannot exceed 7 calendar days")
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
	NPCName         string
	Employer        string
	PersonalContact string
}

var ordinaryMessages = []struct {
	subject string
	body    string
}{
	{"Shift memorandum", "The revised office rota has been left with the evening clerk. Kindly initial your copy before the first bell."},
	{"A note for the household", "I shall keep supper beneath the cover until your return. The street relay was busy again after six."},
	{"Collection notice", "A sealed correspondence packet is being held at the Meridian exchange counter under your account name."},
	{"Counter hours", "The settlement office will close its public counter early on Thursday. Private account enquiries remain chargeable."},
	{"Sunday arrangements", "The downstairs room has been reserved for the family. Send word through the exchange if your plans change."},
	{"Copy returned", "The carbon copy you requested has been returned to the company registry. The original remains in the private file."},
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
	start, end := utcDate(from), utcDate(through)
	if end.Sub(start) > (MaxRangeDays-1)*24*time.Hour {
		return nil, ErrRangeTooLarge
	}
	if context.NPCName == "" {
		context.NPCName = "Account holder"
	}
	if context.Employer == "" {
		context.Employer = "Meridian Global Correspondence Office"
	}
	if context.PersonalContact == "" {
		context.PersonalContact = "Meridian Correspondence Exchange"
	}

	result := make([]Message, 0)
	for date := start; !date.After(end); date = date.AddDate(0, 0, 1) {
		rng := newStream(seed, worldVersion, npcID, date)
		if rng.pick(3) == 0 {
			continue
		}
		item := ordinaryMessages[rng.pick(len(ordinaryMessages))]
		sender := context.Employer
		if rng.pick(2) == 0 {
			sender = context.PersonalContact
		}
		result = append(result, Message{
			ID:           stableID(seed, worldVersion, npcID, date),
			SentAt:       date.Add(time.Duration(7*60+rng.pick(13*60)) * time.Minute),
			Service:      "Meridian Global Correspondence Exchange",
			Sender:       sender,
			Recipient:    context.NPCName,
			Subject:      item.subject,
			Body:         item.body,
			MailboxNPCID: npcID,
		})
	}
	SortMessages(result)
	return result, nil
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

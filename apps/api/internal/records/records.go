// Package records generates query-scoped personal activity and keeps case links
// as backend-only metadata.
package records

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
	ErrInvalidQuery       = errors.New("records query requires an NPC and a valid date range")
	ErrUnsupportedVersion = errors.New("unsupported records generator version")
	ErrRangeTooLarge      = errors.New("records query cannot exceed 7 calendar days")
)

// Record contains the player-visible face of a financial or GATE record.
// The case relationship fields are intentionally omitted from JSON.
type Record struct {
	ID          string    `json:"id"`
	NPCID       string    `json:"npcId"`
	Source      string    `json:"source"`
	OccurredAt  time.Time `json:"occurredAt"`
	Operator    string    `json:"operator"`
	Description string    `json:"description"`
	Reference   string    `json:"reference,omitempty"`
	Amount      int       `json:"amount,omitempty"`
	Direction   string    `json:"direction,omitempty"`
	Origin      string    `json:"origin,omitempty"`
	Destination string    `json:"destination,omitempty"`

	CaseID           string `json:"-"`
	InternalRelation string `json:"-"`
	RelatedNPCID     string `json:"-"`
}

// Context supplies only facts already generated for the queried person.
type Context struct {
	RegionName   string
	Employer     string
	HomeDistrict string
	Districts    []string
}

var bankActivities = []string{
	"A household account settled a routine service charge.",
	"A standing order was processed by the account holder.",
	"A cash withdrawal was entered at a company counter.",
	"A small payment for household goods was recorded.",
	"A personal transfer was received and cleared.",
	"A personal transfer was sent and cleared.",
	"A salary allocation was entered by the employer's accounts office.",
}

var gateActivities = []string{
	"A passenger entry was recorded at the district relay.",
	"A passenger exit was recorded at the district relay.",
	"A GATE reservation was amended at a public terminal.",
	"A relay entry was registered under the account holder's travel plan.",
	"A return passage was recorded at the local relay.",
}

// Generate returns routine records for one NPC and one inclusive interval.
// It never enumerates other residents or materializes a city's activity.
func Generate(seed string, worldVersion int, npcID string, context Context, from, through time.Time) ([]Record, error) {
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
	if context.RegionName == "" {
		context.RegionName = "Lyon"
	}
	if context.HomeDistrict == "" {
		context.HomeDistrict = context.RegionName
	}
	if context.Employer == "" {
		context.Employer = "Independent account"
	}
	districts := append([]string(nil), context.Districts...)
	if len(districts) == 0 {
		districts = []string{context.HomeDistrict, "Bellecour", "Part-Dieu"}
	}
	sort.Strings(districts)

	result := make([]Record, 0)
	for date := start; !date.After(end); date = date.AddDate(0, 0, 1) {
		for slot := 0; slot < 2; slot++ {
			rng := newStream(seed, worldVersion, npcID, date, slot)
			// Each source has a quiet day, so an interval contains ordinary gaps
			// as well as entries without every NPC appearing constantly online.
			if rng.pick(4) == 0 {
				continue
			}
			record := Record{
				ID:         stableID(seed, worldVersion, npcID, date, slot),
				NPCID:      npcID,
				OccurredAt: date.Add(time.Duration(6*60+rng.pick(12*60)) * time.Minute),
			}
			if slot == 0 {
				record.Source = "bank"
				record.Operator = "Meridian Global Settlement Office"
				record.Description = bankActivities[rng.pick(len(bankActivities))]
				record.Reference = context.Employer
				record.Amount = 2 + rng.pick(94)
				record.Direction = bankDirection(record.Description)
			} else {
				record.Source = "gate"
				record.Operator = "Meridian Global · GATE Network"
				record.Description = gateActivities[rng.pick(len(gateActivities))]
				choices := make([]string, 0, len(districts))
				for _, district := range districts {
					if district != context.HomeDistrict {
						choices = append(choices, district)
					}
				}
				record.Origin, record.Destination = context.HomeDistrict, context.HomeDistrict
				if len(choices) > 0 {
					record.Destination = choices[rng.pick(len(choices))]
					if record.Description == gateActivities[4] {
						record.Origin, record.Destination = record.Destination, record.Origin
					}
				}
			}
			result = append(result, record)
		}
	}
	sortRecords(result)
	return result, nil
}

// ExtortionEvents creates ordinary-looking ledger and movement entries for a
// murder scenario. Their links remain available only to server-side case logic.
func ExtortionEvents(seed string, worldVersion int, caseID, killerID, victimID, sceneDistrict string) []Record {
	key := fmt.Sprintf("noir-records/v%d/%s/%d/%s", GeneratorVersion, seed, worldVersion, caseID)
	transferID := stableEventID(key, "transfer", victimID)
	travelID := stableEventID(key, "gate", killerID)
	transferMinute := 11*60 + intFromKey(key, "transfer-time", 7*60)
	travelMinute := 15*60 + intFromKey(key, "gate-time", 4*60)
	origin := "Part-Dieu"
	if origin == sceneDistrict {
		origin = "Bellecour"
	}
	return []Record{
		{
			ID: transferID, NPCID: victimID, Source: "bank",
			OccurredAt:  time.Date(1974, time.November, 9+intFromKey(key, "transfer-day", 2), 0, 0, 0, 0, time.UTC).Add(time.Duration(transferMinute) * time.Minute),
			Operator:    "Meridian Global Settlement Office",
			Description: "A private remittance cleared to a personal account.",
			Reference:   "private account transfer", Amount: 45 + intFromKey(key, "transfer-amount", 46), Direction: "outgoing",
			CaseID: caseID, InternalRelation: "extortion_payment", RelatedNPCID: killerID,
		},
		{
			ID: travelID, NPCID: killerID, Source: "gate",
			OccurredAt:  time.Date(1974, time.November, 12, 0, 0, 0, 0, time.UTC).Add(time.Duration(travelMinute) * time.Minute),
			Operator:    "Meridian Global · GATE Network",
			Description: "A single passenger passage was recorded on the local relay.",
			Origin:      origin, Destination: sceneDistrict,
			CaseID: caseID, InternalRelation: "precrime_route",
		},
	}
}

func bankDirection(description string) string {
	switch description {
	case "A personal transfer was received and cleared.", "A salary allocation was entered by the employer's accounts office.":
		return "incoming"
	default:
		return "outgoing"
	}
}

func intFromKey(key, field string, limit int) int {
	digest := sha256.Sum256([]byte(key + "/" + field))
	var value uint64
	for _, part := range digest[:8] {
		value = value<<8 | uint64(part)
	}
	return int(value % uint64(limit))
}

func sortRecords(records []Record) {
	sort.Slice(records, func(i, j int) bool {
		if records[i].OccurredAt.Equal(records[j].OccurredAt) {
			return records[i].ID < records[j].ID
		}
		return records[i].OccurredAt.Before(records[j].OccurredAt)
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

func newStream(seed string, worldVersion int, npcID string, date time.Time, slot int) stream {
	key := sha256.Sum256([]byte(fmt.Sprintf("noir-records/v%d/%s/%d/%s/%s/%d", GeneratorVersion, seed, worldVersion, npcID, date.Format("2006-01-02"), slot)))
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

func stableID(seed string, worldVersion int, npcID string, date time.Time, slot int) string {
	return stableEventID(fmt.Sprintf("noir-record-id/v%d/%s/%d", GeneratorVersion, seed, worldVersion), npcID, date.Format("2006-01-02"), fmt.Sprint(slot))
}

func stableEventID(parts ...string) string {
	key := sha256.Sum256([]byte(fmt.Sprint(parts)))
	return "record:" + hex.EncodeToString(key[:12])
}

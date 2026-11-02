package market

import (
	"context"
	"fmt"
	"math"
	"sort"
	"time"

	"github.com/jackc/pgx/v5"
)

const (
	minimumScore = 500
	maximumScore = 1500
)

// ApplyCompanyImpact records one idempotent event and distributes its inverse
// movement across the affected company's deterministic rivals in the same tx.
func ApplyCompanyImpact(ctx context.Context, tx pgx.Tx, serverID, eventID, companyID, eventType string, requestedDelta int, occurredAt time.Time, rivalIDs []string) error {
	if requestedDelta == 0 || companyID == "" || eventID == "" {
		return nil
	}
	participants := append([]string(nil), rivalIDs...)
	participants = append(participants, companyID)
	sort.Strings(participants)
	unique := participants[:0]
	for _, id := range participants {
		if id == "" || (len(unique) > 0 && unique[len(unique)-1] == id) {
			continue
		}
		unique = append(unique, id)
	}
	participants = unique

	if _, err := tx.Exec(ctx, `INSERT INTO company_market_scores(server_id,company_id)
		SELECT $1::uuid, member_id FROM unnest($2::text[]) AS member(member_id) ON CONFLICT DO NOTHING`, serverID, participants); err != nil {
		return err
	}
	rows, err := tx.Query(ctx, `SELECT company_id,score,updated_at FROM company_market_scores
		WHERE server_id=$1::uuid AND company_id=ANY($2::text[]) ORDER BY company_id FOR UPDATE`, serverID, participants)
	if err != nil {
		return err
	}
	scores := make(map[string]int, len(participants))
	for rows.Next() {
		var id string
		var score int
		var updatedAt time.Time
		if err := rows.Scan(&id, &score, &updatedAt); err != nil {
			rows.Close()
			return err
		}
		if score < 1000 {
			days := int(time.Since(updatedAt).Hours() / 24)
			score = min(1000, score+max(days, 0))
		}
		scores[id] = score
	}
	if err := rows.Err(); err != nil {
		rows.Close()
		return err
	}
	rows.Close()

	var alreadyRecorded bool
	if err := tx.QueryRow(ctx, `SELECT EXISTS(SELECT 1 FROM company_market_events WHERE server_id=$1::uuid AND event_id=$2)`, serverID, eventID).Scan(&alreadyRecorded); err != nil {
		return err
	}
	if alreadyRecorded {
		return nil
	}
	if _, ok := scores[companyID]; !ok {
		return fmt.Errorf("market impact target %q is missing from the score lock set", companyID)
	}

	allocated := allocateOpposingDelta(companyID, participants, scores, requestedDelta)
	if allocated[companyID] == 0 {
		return nil
	}
	if occurredAt.IsZero() {
		occurredAt = time.Now().UTC()
	}
	for _, id := range participants {
		delta := allocated[id]
		if delta == 0 {
			continue
		}
		after := scores[id] + delta
		idForEvent, kind := eventID, eventType
		if id != companyID {
			idForEvent = eventID + ":rival:" + id
			kind = "rival_market_movement"
		}
		correlationID := eventID
		if _, err := tx.Exec(ctx, `INSERT INTO company_market_events
			(server_id,event_id,company_id,event_type,score_delta,score_before,score_after,occurred_at,correlation_id,source_company_id)
			VALUES($1::uuid,$2,$3,$4,$5,$6,$7,$8,$9,$10)`, serverID, idForEvent, id, kind, delta, scores[id], after, occurredAt.UTC(), correlationID, companyID); err != nil {
			return err
		}
		if _, err := tx.Exec(ctx, `UPDATE company_market_scores SET score=$3,updated_at=now() WHERE server_id=$1::uuid AND company_id=$2`, serverID, id, after); err != nil {
			return err
		}
	}
	return nil
}

func allocateOpposingDelta(target string, participants []string, scores map[string]int, requested int) map[string]int {
	result := make(map[string]int, len(participants))
	rivals := make([]string, 0, len(participants)-1)
	for _, id := range participants {
		if id != target {
			rivals = append(rivals, id)
		}
	}
	if len(rivals) == 0 {
		magnitude := min(abs(requested), allowedMove(scores[target], requested))
		result[target] = sign(requested) * magnitude
		return result
	}
	for magnitude := abs(requested); magnitude >= len(rivals); magnitude-- {
		if magnitude%len(rivals) != 0 {
			continue
		}
		candidate := make(map[string]int, len(participants))
		candidate[target] = sign(requested) * magnitude
		share := magnitude / len(rivals)
		for _, id := range rivals {
			candidate[id] = -sign(requested) * share
		}
		withinBounds := true
		for id, delta := range candidate {
			after := scores[id] + delta
			if after < minimumScore || after > maximumScore {
				withinBounds = false
				break
			}
		}
		if withinBounds {
			return candidate
		}
	}
	return result
}

func allowedMove(score, requested int) int {
	if requested > 0 {
		return maximumScore - score
	}
	return score - minimumScore
}

func abs(value int) int { return int(math.Abs(float64(value))) }

func sign(value int) int {
	if value < 0 {
		return -1
	}
	return 1
}

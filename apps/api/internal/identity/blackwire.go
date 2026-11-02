package identity

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"strings"

	"github.com/jackc/pgx/v5"
)

var errBlackwireUnavailable = errors.New("blackwire information is unavailable")

type blackwireOffer struct {
	BoardID    string `json:"boardId"`
	BoardTitle string `json:"boardTitle"`
	ItemID     string `json:"itemId"`
	Source     string `json:"-"`
	Title      string `json:"title"`
	Content    string `json:"content"`
	PriceCents int64  `json:"-"`
	Category   string `json:"-"`
	CompanyID  string `json:"-"`
}

type blackwireMarketplace struct {
	Offers []blackwireOffer `json:"offers"`
}

type blackwireStore interface {
	blackwireOffers(context.Context, string, string) (blackwireMarketplace, error)
	sellBlackwireItem(context.Context, string, string, string, string) (blackwireOffer, int64, error)
}

func classifyBlackwireOffer(item map[string]any) (int64, string, string) {
	title, _ := item["title"].(string)
	content, _ := item["content"].(string)
	source, _ := item["source"].(string)
	metadata, _ := item["metadata"].(map[string]any)
	text := strings.ToLower(strings.Join([]string{title, content, source}, " "))
	companyID := ""
	if metadata != nil {
		for _, key := range []string{"companyId", "employerCompanyId"} {
			if value, ok := metadata[key].(string); ok && value != "" {
				companyID = value
				break
			}
		}
		if company, ok := metadata["company"].(map[string]any); ok {
			if value, ok := company["id"].(string); ok && companyID == "" {
				companyID = value
			}
		}
	}
	category, price := "ROUTINE INFORMATION", int64(25)
	if metadata != nil {
		for _, key := range []string{"caseId", "crimeType", "crimeId"} {
			if value, ok := metadata[key].(string); ok && value != "" {
				category, price = "CASE INFORMATION", 2500
				break
			}
		}
	}
	financial := strings.Contains(text, "fraud") || strings.Contains(text, "extortion") || strings.Contains(text, "embezzle") || strings.Contains(text, "payroll") || strings.Contains(text, "accounting") || strings.Contains(text, "financial") || strings.Contains(text, "bank account") || strings.Contains(text, "transfer")
	executive := strings.Contains(text, "director") || strings.Contains(text, "chief executive") || strings.Contains(text, "president") || strings.Contains(text, "board member") || strings.Contains(text, "managing partner")
	switch {
	case financial && executive:
		category, price = "EXECUTIVE FINANCIAL EVIDENCE", 75000
	case financial:
		category, price = "FINANCIAL / CORPORATE EVIDENCE", 35000
	case executive && companyID != "":
		category, price = "CORPORATE LEADERSHIP", 20000
	case category == "CASE INFORMATION":
		price = 5000
	}
	return price, category, companyID
}

func (s *postgresStore) blackwireOffers(ctx context.Context, accountID, serverID string) (blackwireMarketplace, error) {
	queryer, ok := s.db.(interface {
		Query(context.Context, string, ...any) (pgx.Rows, error)
	})
	if !ok {
		return blackwireMarketplace{}, errors.New("whiteboard queries are unavailable")
	}
	rows, err := queryer.Query(ctx, `SELECT board.whiteboard_id, board.title, item.item_id, item.item_data
		FROM player_whiteboards board
		JOIN player_whiteboard_pins pin USING (server_id, account_id, whiteboard_id)
		JOIN player_discovered_items item USING (server_id, account_id, item_id)
		LEFT JOIN player_blackwire_sales sale USING (server_id, account_id, item_id)
		WHERE board.server_id=$1::uuid AND board.account_id=$2::uuid AND sale.item_id IS NULL
			AND NOT EXISTS (SELECT 1 FROM player_whiteboard_pins locked_pin JOIN player_whiteboards locked_board USING(server_id,account_id,whiteboard_id)
				WHERE locked_pin.server_id=item.server_id AND locked_pin.account_id=item.account_id AND locked_pin.item_id=item.item_id AND locked_board.locked_at IS NOT NULL)
			AND NOT EXISTS (
				SELECT 1 FROM player_whiteboard_pins case_pin
				JOIN (
					SELECT server_id,case_id,evidence_id AS item_id FROM case_evidence
					UNION SELECT server_id,case_id,record_id AS item_id FROM case_activity_records
					UNION SELECT server_id,case_id,message_id AS item_id FROM case_mail_messages
				) case_item ON case_item.server_id=case_pin.server_id AND case_item.item_id=case_pin.item_id
				JOIN case_states case_state ON case_state.server_id=case_item.server_id AND case_state.case_id=case_item.case_id
				WHERE case_pin.server_id=board.server_id AND case_pin.account_id=board.account_id
					AND case_pin.whiteboard_id=board.whiteboard_id AND case_state.status<>'open'
			)
		ORDER BY board.created_at, board.whiteboard_id, pin.pinned_at, item.item_id`, serverID, accountID)
	if err != nil {
		return blackwireMarketplace{}, err
	}
	defer rows.Close()
	result := blackwireMarketplace{Offers: []blackwireOffer{}}
	for rows.Next() {
		var boardID, boardTitle, itemID string
		var raw []byte
		if err := rows.Scan(&boardID, &boardTitle, &itemID, &raw); err != nil {
			return blackwireMarketplace{}, err
		}
		var item map[string]any
		if err := json.Unmarshal(raw, &item); err != nil {
			continue
		}
		price, category, companyID := classifyBlackwireOffer(item)
		offer := blackwireOffer{BoardID: boardID, BoardTitle: boardTitle, ItemID: itemID, PriceCents: price, Category: category, CompanyID: companyID}
		offer.Source, _ = item["source"].(string)
		offer.Title, _ = item["title"].(string)
		offer.Content, _ = item["content"].(string)
		result.Offers = append(result.Offers, offer)
	}
	if err := rows.Err(); err != nil {
		return blackwireMarketplace{}, err
	}
	return result, nil
}

func (s *postgresStore) sellBlackwireItem(ctx context.Context, accountID, serverID, boardID, itemID string) (blackwireOffer, int64, error) {
	transactional, ok := s.db.(interface {
		Begin(context.Context) (pgx.Tx, error)
	})
	if !ok {
		return blackwireOffer{}, 0, errors.New("blackwire transactions are unavailable")
	}
	tx, err := transactional.Begin(ctx)
	if err != nil {
		return blackwireOffer{}, 0, err
	}
	defer tx.Rollback(ctx)
	var boardTitle string
	var raw []byte
	err = tx.QueryRow(ctx, `SELECT board.title, item.item_data
		FROM player_whiteboards board
		JOIN player_whiteboard_pins pin USING (server_id, account_id, whiteboard_id)
		JOIN player_discovered_items item USING (server_id, account_id, item_id)
		LEFT JOIN player_blackwire_sales sale USING (server_id, account_id, item_id)
		WHERE board.server_id=$1::uuid AND board.account_id=$2::uuid AND board.whiteboard_id=$3 AND item.item_id=$4
			AND board.locked_at IS NULL AND sale.item_id IS NULL
			AND NOT EXISTS (SELECT 1 FROM player_whiteboard_pins locked_pin JOIN player_whiteboards locked_board USING(server_id,account_id,whiteboard_id)
				WHERE locked_pin.server_id=item.server_id AND locked_pin.account_id=item.account_id AND locked_pin.item_id=item.item_id AND locked_board.locked_at IS NOT NULL)
			AND NOT EXISTS (
				SELECT 1 FROM player_whiteboard_pins case_pin
				JOIN (
					SELECT server_id,case_id,evidence_id AS item_id FROM case_evidence
					UNION SELECT server_id,case_id,record_id AS item_id FROM case_activity_records
					UNION SELECT server_id,case_id,message_id AS item_id FROM case_mail_messages
				) case_item ON case_item.server_id=case_pin.server_id AND case_item.item_id=case_pin.item_id
				JOIN case_states case_state ON case_state.server_id=case_item.server_id AND case_state.case_id=case_item.case_id
				WHERE case_pin.server_id=board.server_id AND case_pin.account_id=board.account_id
					AND case_pin.whiteboard_id=board.whiteboard_id AND case_state.status<>'open'
			)
		FOR UPDATE OF item`, serverID, accountID, boardID, itemID).Scan(&boardTitle, &raw)
	if errors.Is(err, pgx.ErrNoRows) {
		return blackwireOffer{}, 0, errBlackwireUnavailable
	}
	if err != nil {
		return blackwireOffer{}, 0, err
	}
	var item map[string]any
	if err := json.Unmarshal(raw, &item); err != nil {
		return blackwireOffer{}, 0, fmt.Errorf("decode discovered information: %w", err)
	}
	price, category, companyID := classifyBlackwireOffer(item)
	var balance int64
	err = tx.QueryRow(ctx, `UPDATE player_economy SET credits_cents=credits_cents+$3, updated_at=now()
		WHERE server_id=$1::uuid AND account_id=$2::uuid RETURNING credits_cents`, serverID, accountID, price).Scan(&balance)
	if errors.Is(err, pgx.ErrNoRows) {
		return blackwireOffer{}, 0, ErrNoCharacter
	}
	if err != nil {
		return blackwireOffer{}, 0, err
	}
	_, err = tx.Exec(ctx, `INSERT INTO player_blackwire_sales(server_id,account_id,item_id,whiteboard_id,price_cents,category,company_id)
		VALUES($1::uuid,$2::uuid,$3,$4,$5,$6,NULLIF($7,''))`, serverID, accountID, itemID, boardID, price, category, companyID)
	if err != nil {
		return blackwireOffer{}, 0, err
	}
	messageID := "blackwire-sale:" + itemID
	title, _ := item["title"].(string)
	body := fmt.Sprintf("The Blackwire exchange has completed the sale of %q. Credits received: CR %d.%02d. The information has been withdrawn from your boards and cannot be used as case evidence.",
		title, price/100, price%100)
	if _, err := tx.Exec(ctx, `INSERT INTO player_inbox_messages(server_id,account_id,message_id,sent_at,sender,message_category,subject,body,attachments)
		VALUES($1::uuid,$2::uuid,$3,now(),'Blackwire Exchange','official','Blackwire sale · Settlement complete',$4,'[]'::jsonb)
		ON CONFLICT(server_id,account_id,message_id) DO NOTHING`, serverID, accountID, messageID, body); err != nil {
		return blackwireOffer{}, 0, err
	}
	if companyID != "" {
		delta := blackwireMarketDelta(category)
		if delta != 0 {
			if _, err := tx.Exec(ctx, `INSERT INTO company_market_scores(server_id,company_id) VALUES($1::uuid,$2) ON CONFLICT DO NOTHING`, serverID, companyID); err != nil {
				return blackwireOffer{}, 0, err
			}
			var before int
			if err := tx.QueryRow(ctx, `SELECT CASE WHEN score < 1000 THEN LEAST(1000, score + FLOOR(EXTRACT(EPOCH FROM (now()-updated_at))/86400)::integer) ELSE score END FROM company_market_scores WHERE server_id=$1::uuid AND company_id=$2 FOR UPDATE`, serverID, companyID).Scan(&before); err != nil {
				return blackwireOffer{}, 0, err
			}
			after := max(500, before+delta)
			eventID := "blackwire:" + accountID + ":" + itemID
			result, err := tx.Exec(ctx, `INSERT INTO company_market_events(server_id,event_id,company_id,event_type,score_delta,score_before,score_after,occurred_at)
				VALUES($1::uuid,$2,$3,'blackwire_leak',$4,$5,$6,now()) ON CONFLICT(server_id,event_id) DO NOTHING`, serverID, eventID, companyID, delta, before, after)
			if err != nil {
				return blackwireOffer{}, 0, err
			}
			if result.RowsAffected() > 0 {
				if _, err := tx.Exec(ctx, `UPDATE company_market_scores SET score=$3,updated_at=now() WHERE server_id=$1::uuid AND company_id=$2`, serverID, companyID, after); err != nil {
					return blackwireOffer{}, 0, err
				}
			}
		}
	}
	if _, err := tx.Exec(ctx, `DELETE FROM player_whiteboard_pins WHERE server_id=$1::uuid AND account_id=$2::uuid AND item_id=$3`, serverID, accountID, itemID); err != nil {
		return blackwireOffer{}, 0, err
	}
	if _, err := tx.Exec(ctx, `DELETE FROM player_whiteboard_connections WHERE server_id=$1::uuid AND account_id=$2::uuid AND (from_item_id=$3 OR to_item_id=$3)`, serverID, accountID, itemID); err != nil {
		return blackwireOffer{}, 0, err
	}
	if _, err := tx.Exec(ctx, `UPDATE player_discovered_items SET item_data=item_data || '{"blackwireSold":true}'::jsonb WHERE server_id=$1::uuid AND account_id=$2::uuid AND item_id=$3`, serverID, accountID, itemID); err != nil {
		return blackwireOffer{}, 0, err
	}
	if err := tx.Commit(ctx); err != nil {
		return blackwireOffer{}, 0, err
	}
	offer := blackwireOffer{BoardID: boardID, BoardTitle: boardTitle, ItemID: itemID, PriceCents: price, Category: category, CompanyID: companyID}
	offer.Source, _ = item["source"].(string)
	offer.Title, _ = item["title"].(string)
	offer.Content, _ = item["content"].(string)
	return offer, balance, nil
}

func blackwireMarketDelta(category string) int {
	switch category {
	case "EXECUTIVE FINANCIAL EVIDENCE":
		return -150
	case "FINANCIAL / CORPORATE EVIDENCE":
		return -100
	case "CORPORATE LEADERSHIP":
		return -60
	case "CASE INFORMATION":
		return -35
	default:
		return 0
	}
}

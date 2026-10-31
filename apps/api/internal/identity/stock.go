package identity

import (
	"context"
	"crypto/sha256"
	"encoding/binary"
	"errors"
	"fmt"
	"math"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgconn"
)

var (
	ErrNoCharacter         = errors.New("player character required")
	ErrInsufficientFunds   = errors.New("insufficient credits")
	ErrInsufficientShares  = errors.New("not enough shares")
	ErrDuplicateStockOrder = errors.New("stock order already processed")
	ErrInvalidStockOrder   = errors.New("invalid stock order")
)

const (
	stockPageSize       = 25
	stockOrderMaxShares = 10000
)

type playerProgression struct {
	Fame         int64 `json:"fame"`
	CreditsCents int64 `json:"creditsCents"`
}

type stockListing struct {
	CompanyID   string `json:"companyId"`
	CompanyName string `json:"companyName"`
	CountryID   string `json:"countryId"`
	CityName    string `json:"cityName"`
	PriceCents  int64  `json:"priceCents"`
	MarketScore int    `json:"marketScore"`
	ChangeBps   int64  `json:"changeBps"`
	LastEventAt string `json:"lastEventAt,omitempty"`
}

type stockPosition struct {
	CompanyID        string `json:"companyId"`
	CompanyName      string `json:"companyName"`
	CountryID        string `json:"countryId"`
	CityName         string `json:"cityName"`
	Shares           int64  `json:"shares"`
	AverageCostCents int64  `json:"averageCostCents"`
	PriceCents       int64  `json:"priceCents"`
	MarketValueCents int64  `json:"marketValueCents"`
	MarketScore      int    `json:"marketScore"`
	ChangeBps        int64  `json:"changeBps"`
	LastEventAt      string `json:"lastEventAt,omitempty"`
}

type companyMarketSnapshot struct {
	Score       int
	ChangeBps   int64
	LastEventAt string
}

type companyMarketEvent struct {
	EventType   string `json:"eventType"`
	ScoreBefore int    `json:"scoreBefore"`
	ScoreAfter  int    `json:"scoreAfter"`
	ChangeBps   int64  `json:"changeBps"`
	OccurredAt  string `json:"occurredAt"`
}

type stockOrderInput struct {
	ClientOrderID string `json:"clientOrderId"`
	CompanyID     string `json:"companyId"`
	Side          string `json:"side"`
	Shares        int64  `json:"shares"`
}

func stockQuoteCents(countryID, companyID string, marketScore int) int64 {
	digest := sha256.Sum256([]byte("noir-stock-quote-v1\x00" + countryID + "\x00" + companyID))
	base := 5000 + int64(binary.BigEndian.Uint64(digest[:8])%95001)
	if marketScore < 500 {
		marketScore = 500
	}
	if marketScore > 1500 {
		marketScore = 1500
	}
	return (base*int64(marketScore) + 500) / 1000
}

func (s *postgresStore) companyMarketSnapshots(ctx context.Context, serverID string, companyIDs []string) (map[string]companyMarketSnapshot, error) {
	queryer, ok := s.db.(interface {
		Query(context.Context, string, ...any) (pgx.Rows, error)
	})
	if !ok {
		return nil, errors.New("market quote queries are unavailable")
	}
	rows, err := queryer.Query(ctx, `SELECT company.company_id,
		CASE WHEN COALESCE(score.score,1000) < 1000 THEN LEAST(1000, score.score + FLOOR(EXTRACT(EPOCH FROM (now()-score.updated_at))/86400)::integer) ELSE COALESCE(score.score,1000) END,
		COALESCE(ROUND((event.score_after-event.score_before)*10000.0/NULLIF(event.score_before,0)),0)::bigint,
		COALESCE(to_char(event.occurred_at AT TIME ZONE 'UTC','YYYY-MM-DD"T"HH24:MI:SS"Z"'),'')
		FROM unnest($2::text[]) AS company(company_id)
		LEFT JOIN company_market_scores score ON score.server_id=$1::uuid AND score.company_id=company.company_id
		LEFT JOIN LATERAL (SELECT score_before,score_after,occurred_at FROM company_market_events
			WHERE server_id=$1::uuid AND company_id=company.company_id ORDER BY recorded_at DESC,event_id DESC LIMIT 1) event ON true`, serverID, companyIDs)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	result := make(map[string]companyMarketSnapshot, len(companyIDs))
	for rows.Next() {
		var companyID string
		var snapshot companyMarketSnapshot
		if err := rows.Scan(&companyID, &snapshot.Score, &snapshot.ChangeBps, &snapshot.LastEventAt); err != nil {
			return nil, err
		}
		result[companyID] = snapshot
	}
	return result, rows.Err()
}

func (s *postgresStore) companyMarketHistory(ctx context.Context, serverID, companyID string, offset, limit int) ([]companyMarketEvent, error) {
	queryer, ok := s.db.(interface {
		Query(context.Context, string, ...any) (pgx.Rows, error)
	})
	if !ok {
		return nil, errors.New("market history queries are unavailable")
	}
	rows, err := queryer.Query(ctx, `SELECT score_before,score_after,
		COALESCE(ROUND((score_after-score_before)*10000.0/NULLIF(score_before,0)),0)::bigint,
		to_char(occurred_at AT TIME ZONE 'UTC','YYYY-MM-DD"T"HH24:MI:SS"Z"')
		FROM company_market_events WHERE server_id=$1::uuid AND company_id=$2
		ORDER BY recorded_at DESC,event_id DESC LIMIT $3 OFFSET $4`, serverID, companyID, limit, offset)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	result := make([]companyMarketEvent, 0)
	for rows.Next() {
		var event companyMarketEvent
		event.EventType = "Company incident"
		if err := rows.Scan(&event.ScoreBefore, &event.ScoreAfter, &event.ChangeBps, &event.OccurredAt); err != nil {
			return nil, err
		}
		result = append(result, event)
	}
	return result, rows.Err()
}

func (s *postgresStore) playerStockPortfolio(ctx context.Context, accountID, serverID string) (playerProgression, []stockPosition, error) {
	_, err := s.db.Exec(ctx, `INSERT INTO player_economy (server_id, account_id)
		SELECT $1::uuid, account_id FROM player_characters WHERE account_id=$2::uuid
		ON CONFLICT (server_id, account_id) DO NOTHING`, serverID, accountID)
	if err != nil {
		return playerProgression{}, nil, err
	}
	var progression playerProgression
	err = s.db.QueryRow(ctx, `SELECT fame, credits_cents FROM player_economy WHERE server_id=$1::uuid AND account_id=$2::uuid`, serverID, accountID).Scan(&progression.Fame, &progression.CreditsCents)
	if errors.Is(err, pgx.ErrNoRows) {
		return playerProgression{}, nil, ErrNoCharacter
	}
	if err != nil {
		return playerProgression{}, nil, err
	}
	queryer, ok := s.db.(interface {
		Query(context.Context, string, ...any) (pgx.Rows, error)
	})
	if !ok {
		return playerProgression{}, nil, errors.New("stock portfolio queries are unavailable")
	}
	rows, err := queryer.Query(ctx, `SELECT position.company_id, position.company_name, position.country_id, position.city_name, position.shares, position.average_cost_cents,
		CASE WHEN COALESCE(score.score,1000) < 1000 THEN LEAST(1000, score.score + FLOOR(EXTRACT(EPOCH FROM (now()-score.updated_at))/86400)::integer) ELSE COALESCE(score.score,1000) END,
		COALESCE(ROUND((event.score_after-event.score_before)*10000.0/NULLIF(event.score_before,0)),0)::bigint,
		COALESCE(to_char(event.occurred_at AT TIME ZONE 'UTC','YYYY-MM-DD"T"HH24:MI:SS"Z"'),'')
		FROM player_stock_positions position
		LEFT JOIN company_market_scores score ON score.server_id=position.server_id AND score.company_id=position.company_id
		LEFT JOIN LATERAL (SELECT score_before,score_after,occurred_at FROM company_market_events
			WHERE server_id=position.server_id AND company_id=position.company_id ORDER BY recorded_at DESC,event_id DESC LIMIT 1) event ON true
		WHERE position.server_id=$1::uuid AND position.account_id=$2::uuid ORDER BY position.company_name, position.company_id`, serverID, accountID)
	if err != nil {
		return playerProgression{}, nil, err
	}
	defer rows.Close()
	positions := make([]stockPosition, 0)
	for rows.Next() {
		var position stockPosition
		if err := rows.Scan(&position.CompanyID, &position.CompanyName, &position.CountryID, &position.CityName, &position.Shares, &position.AverageCostCents, &position.MarketScore, &position.ChangeBps, &position.LastEventAt); err != nil {
			return playerProgression{}, nil, err
		}
		position.PriceCents = stockQuoteCents(position.CountryID, position.CompanyID, position.MarketScore)
		if position.Shares > math.MaxInt64/position.PriceCents {
			return playerProgression{}, nil, errors.New("stock position value overflow")
		}
		position.MarketValueCents = position.Shares * position.PriceCents
		positions = append(positions, position)
	}
	return progression, positions, rows.Err()
}

type stockTradeStore interface {
	Begin(context.Context) (pgx.Tx, error)
}

func (s *postgresStore) executeStockOrder(ctx context.Context, accountID, serverID string, listing *stockListing, input stockOrderInput) (int64, error) {
	if listing == nil || input.ClientOrderID == "" || len(input.ClientOrderID) > 80 || input.CompanyID != listing.CompanyID || input.Shares < 1 || input.Shares > stockOrderMaxShares || (input.Side != "buy" && input.Side != "sell") {
		return 0, ErrInvalidStockOrder
	}
	beginner, ok := s.db.(stockTradeStore)
	if !ok {
		return 0, errors.New("atomic stock orders are unavailable")
	}
	tx, err := beginner.Begin(ctx)
	if err != nil {
		return 0, err
	}
	defer tx.Rollback(ctx)
	var exists bool
	err = tx.QueryRow(ctx, `INSERT INTO player_economy (server_id, account_id)
		SELECT $1::uuid, account_id FROM player_characters WHERE account_id=$2::uuid
		ON CONFLICT (server_id, account_id) DO NOTHING RETURNING true`, serverID, accountID).Scan(&exists)
	if errors.Is(err, pgx.ErrNoRows) {
		var credits int64
		err = tx.QueryRow(ctx, `SELECT credits_cents FROM player_economy WHERE server_id=$1::uuid AND account_id=$2::uuid`, serverID, accountID).Scan(&credits)
	}
	if errors.Is(err, pgx.ErrNoRows) {
		return 0, ErrNoCharacter
	}
	if err != nil {
		return 0, err
	}
	if _, err := tx.Exec(ctx, `INSERT INTO company_market_scores(server_id, company_id) VALUES($1::uuid, $2) ON CONFLICT DO NOTHING`, serverID, listing.CompanyID); err != nil {
		return 0, err
	}
	if err := tx.QueryRow(ctx, `SELECT CASE WHEN score < 1000 THEN LEAST(1000, score + FLOOR(EXTRACT(EPOCH FROM (now()-updated_at))/86400)::integer) ELSE score END
		FROM company_market_scores WHERE server_id=$1::uuid AND company_id=$2 FOR SHARE`, serverID, listing.CompanyID).Scan(&listing.MarketScore); err != nil {
		return 0, err
	}
	listing.PriceCents = stockQuoteCents(listing.CountryID, listing.CompanyID, listing.MarketScore)
	var eventAt string
	err = tx.QueryRow(ctx, `SELECT score_before,score_after,
		COALESCE(ROUND((score_after-score_before)*10000.0/NULLIF(score_before,0)),0)::bigint,
		to_char(occurred_at AT TIME ZONE 'UTC','YYYY-MM-DD"T"HH24:MI:SS"Z"')
		FROM company_market_events WHERE server_id=$1::uuid AND company_id=$2
		ORDER BY recorded_at DESC,event_id DESC LIMIT 1`, serverID, listing.CompanyID).Scan(new(int), new(int), &listing.ChangeBps, &eventAt)
	if err != nil && !errors.Is(err, pgx.ErrNoRows) {
		return 0, err
	}
	listing.LastEventAt = eventAt
	if input.Shares > math.MaxInt64/listing.PriceCents {
		return 0, errors.New("stock order value is too large")
	}
	total := input.Shares * listing.PriceCents
	_, err = tx.Exec(ctx, `INSERT INTO player_stock_orders (server_id, account_id, client_order_id, company_id, side, shares, price_cents, total_cents)
		VALUES ($1::uuid, $2::uuid, $3, $4, $5, $6, $7, $8)`, serverID, accountID, input.ClientOrderID, listing.CompanyID, input.Side, input.Shares, listing.PriceCents, total)
	if err != nil {
		var pgErr *pgconn.PgError
		if errors.As(err, &pgErr) && pgErr.Code == "23505" {
			return 0, ErrDuplicateStockOrder
		}
		return 0, err
	}
	var credits int64
	if err := tx.QueryRow(ctx, `SELECT credits_cents FROM player_economy WHERE server_id=$1::uuid AND account_id=$2::uuid FOR UPDATE`, serverID, accountID).Scan(&credits); err != nil {
		return 0, err
	}
	var position stockPosition
	err = tx.QueryRow(ctx, `SELECT shares, average_cost_cents FROM player_stock_positions
		WHERE server_id=$1::uuid AND account_id=$2::uuid AND company_id=$3 FOR UPDATE`, serverID, accountID, listing.CompanyID).Scan(&position.Shares, &position.AverageCostCents)
	hasPosition := err == nil
	if err != nil && !errors.Is(err, pgx.ErrNoRows) {
		return 0, err
	}
	updatedCredits := credits
	if input.Side == "buy" {
		if credits < total {
			return 0, ErrInsufficientFunds
		}
		if hasPosition && position.Shares > math.MaxInt64-input.Shares {
			return 0, errors.New("stock position is too large")
		}
		newShares := input.Shares
		newAverage := listing.PriceCents
		if hasPosition {
			newShares += position.Shares
			if position.Shares > (math.MaxInt64-total)/position.AverageCostCents {
				return 0, errors.New("stock position cost overflow")
			}
			weightedCost := position.Shares*position.AverageCostCents + total
			if weightedCost > math.MaxInt64-newShares/2 {
				return 0, errors.New("stock average cost overflow")
			}
			newAverage = (weightedCost + newShares/2) / newShares
		}
		if _, err := tx.Exec(ctx, `UPDATE player_economy SET credits_cents=credits_cents-$3, updated_at=now() WHERE server_id=$1::uuid AND account_id=$2::uuid`, serverID, accountID, total); err != nil {
			return 0, err
		}
		updatedCredits -= total
		_, err = tx.Exec(ctx, `INSERT INTO player_stock_positions (server_id, account_id, company_id, country_id, company_name, city_name, shares, average_cost_cents)
			VALUES ($1::uuid, $2::uuid, $3, $4, $5, $6, $7, $8)
			ON CONFLICT (server_id, account_id, company_id) DO UPDATE SET country_id=EXCLUDED.country_id, company_name=EXCLUDED.company_name, city_name=EXCLUDED.city_name, shares=EXCLUDED.shares, average_cost_cents=EXCLUDED.average_cost_cents, updated_at=now()`, serverID, accountID, listing.CompanyID, listing.CountryID, listing.CompanyName, listing.CityName, newShares, newAverage)
	} else {
		if !hasPosition || position.Shares < input.Shares {
			return 0, ErrInsufficientShares
		}
		if credits > math.MaxInt64-total {
			return 0, errors.New("player credit balance overflow")
		}
		if position.Shares == input.Shares {
			_, err = tx.Exec(ctx, `DELETE FROM player_stock_positions WHERE server_id=$1::uuid AND account_id=$2::uuid AND company_id=$3`, serverID, accountID, listing.CompanyID)
		} else {
			_, err = tx.Exec(ctx, `UPDATE player_stock_positions SET shares=shares-$4, updated_at=now() WHERE server_id=$1::uuid AND account_id=$2::uuid AND company_id=$3`, serverID, accountID, listing.CompanyID, input.Shares)
		}
		if err == nil {
			_, err = tx.Exec(ctx, `UPDATE player_economy SET credits_cents=credits_cents+$3, updated_at=now() WHERE server_id=$1::uuid AND account_id=$2::uuid`, serverID, accountID, total)
			updatedCredits += total
		}
	}
	if err != nil {
		return 0, err
	}
	if err := tx.Commit(ctx); err != nil {
		return 0, fmt.Errorf("commit stock order: %w", err)
	}
	return updatedCredits, nil
}

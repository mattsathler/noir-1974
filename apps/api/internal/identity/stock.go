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
}

type stockOrderInput struct {
	ClientOrderID string `json:"clientOrderId"`
	CompanyID     string `json:"companyId"`
	Side          string `json:"side"`
	Shares        int64  `json:"shares"`
}

func stockQuoteCents(countryID, companyID string) int64 {
	digest := sha256.Sum256([]byte("noir-stock-quote-v1\x00" + countryID + "\x00" + companyID))
	return 5000 + int64(binary.BigEndian.Uint64(digest[:8])%95001)
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
	rows, err := queryer.Query(ctx, `SELECT company_id, company_name, country_id, city_name, shares, average_cost_cents
		FROM player_stock_positions WHERE server_id=$1::uuid AND account_id=$2::uuid ORDER BY company_name, company_id`, serverID, accountID)
	if err != nil {
		return playerProgression{}, nil, err
	}
	defer rows.Close()
	positions := make([]stockPosition, 0)
	for rows.Next() {
		var position stockPosition
		if err := rows.Scan(&position.CompanyID, &position.CompanyName, &position.CountryID, &position.CityName, &position.Shares, &position.AverageCostCents); err != nil {
			return playerProgression{}, nil, err
		}
		position.PriceCents = stockQuoteCents(position.CountryID, position.CompanyID)
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

func (s *postgresStore) executeStockOrder(ctx context.Context, accountID, serverID string, listing stockListing, input stockOrderInput) (int64, error) {
	if input.ClientOrderID == "" || len(input.ClientOrderID) > 80 || input.CompanyID != listing.CompanyID || input.Shares < 1 || input.Shares > stockOrderMaxShares || (input.Side != "buy" && input.Side != "sell") {
		return 0, ErrInvalidStockOrder
	}
	if input.Shares > math.MaxInt64/listing.PriceCents {
		return 0, errors.New("stock order value is too large")
	}
	total := input.Shares * listing.PriceCents
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

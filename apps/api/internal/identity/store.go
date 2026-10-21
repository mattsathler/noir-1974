package identity

import (
	"context"
	"crypto/sha256"
	"embed"
	"errors"
	"strings"
	"time"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgconn"
	"github.com/jackc/pgx/v5/pgtype"
)

//go:embed schema.sql
var schema embed.FS

var ErrConflict = errors.New("account or character already exists")
var ErrNotFound = errors.New("record not found")

type querier interface {
	Exec(context.Context, string, ...any) (pgconn.CommandTag, error)
	QueryRow(context.Context, string, ...any) pgx.Row
}

type account struct {
	ID           string     `json:"id"`
	Username     string     `json:"username"`
	Email        string     `json:"email"`
	PasswordHash string     `json:"-"`
	Character    *character `json:"-"`
}

type character struct {
	DisplayName string `json:"displayName"`
	PortraitID  int    `json:"portraitId"`
}

type store interface {
	createAccount(context.Context, string, string, string) (account, error)
	accountByEmail(context.Context, string) (account, error)
	createSession(context.Context, string, string, time.Time) error
	accountForSession(context.Context, string, time.Time) (account, *character, error)
	createCharacter(context.Context, string, string, int) (character, error)
	playerLocation(context.Context, string) (string, error)
	setPlayerLocation(context.Context, string, string) error
	deleteSession(context.Context, string) error
}

type postgresStore struct{ db querier }

func NewPostgresStore(db querier) *postgresStore { return &postgresStore{db: db} }

func (s *postgresStore) Migrate(ctx context.Context) error {
	contents, err := schema.ReadFile("schema.sql")
	if err != nil {
		return err
	}
	for _, statement := range strings.Split(string(contents), ";") {
		if strings.TrimSpace(statement) == "" {
			continue
		}
		if _, err := s.db.Exec(ctx, statement); err != nil {
			return err
		}
	}
	return nil
}

func (s *postgresStore) createAccount(ctx context.Context, username, email, passwordHash string) (account, error) {
	var result account
	err := s.db.QueryRow(ctx,
		`INSERT INTO player_accounts (username, email, password_hash)
		 VALUES ($1, $2, $3) RETURNING id::text, username, email, password_hash`,
		username, email, passwordHash,
	).Scan(&result.ID, &result.Username, &result.Email, &result.PasswordHash)
	if err != nil {
		var pgErr *pgconn.PgError
		if errors.As(err, &pgErr) && pgErr.Code == "23505" {
			return account{}, ErrConflict
		}
		return account{}, err
	}
	return result, nil
}

func (s *postgresStore) accountByEmail(ctx context.Context, email string) (account, error) {
	var result account
	var displayName pgtype.Text
	var portraitID pgtype.Int2
	err := s.db.QueryRow(ctx,
		`SELECT a.id::text, a.username, a.email, a.password_hash,
		       c.display_name, c.portrait_id
		 FROM player_accounts a
		 LEFT JOIN player_characters c ON c.account_id = a.id
		 WHERE lower(a.email) = lower($1)`, email,
	).Scan(&result.ID, &result.Username, &result.Email, &result.PasswordHash, &displayName, &portraitID)
	if errors.Is(err, pgx.ErrNoRows) {
		return account{}, ErrNotFound
	}
	if err == nil && displayName.Valid && portraitID.Valid {
		result.Character = &character{DisplayName: displayName.String, PortraitID: int(portraitID.Int16)}
	}
	return result, err
}

func (s *postgresStore) createSession(ctx context.Context, accountID, token string, expires time.Time) error {
	hash := sha256.Sum256([]byte(token))
	_, err := s.db.Exec(ctx,
		`INSERT INTO player_sessions (token_hash, account_id, expires_at) VALUES ($1, $2::text::uuid, $3)`,
		hash[:], accountID, expires,
	)
	return err
}

func (s *postgresStore) accountForSession(ctx context.Context, token string, now time.Time) (account, *character, error) {
	hash := sha256.Sum256([]byte(token))
	var result account
	var displayName pgtype.Text
	var portraitID pgtype.Int2
	err := s.db.QueryRow(ctx,
		`SELECT a.id::text, a.username, a.email,
		       c.display_name, c.portrait_id
		 FROM player_sessions s
		 JOIN player_accounts a ON a.id = s.account_id
		 LEFT JOIN player_characters c ON c.account_id = a.id
		 WHERE s.token_hash = $1 AND s.expires_at > $2`, hash[:], now,
	).Scan(&result.ID, &result.Username, &result.Email, &displayName, &portraitID)
	if errors.Is(err, pgx.ErrNoRows) {
		return account{}, nil, ErrNotFound
	}
	if err != nil {
		return account{}, nil, err
	}
	var characterResult *character
	if displayName.Valid && portraitID.Valid {
		characterResult = &character{DisplayName: displayName.String, PortraitID: int(portraitID.Int16)}
	}
	return result, characterResult, nil
}

func (s *postgresStore) createCharacter(ctx context.Context, accountID, name string, portraitID int) (character, error) {
	var result character
	err := s.db.QueryRow(ctx,
		`INSERT INTO player_characters (account_id, display_name, portrait_id)
		 VALUES ($1::text::uuid, $2, $3) RETURNING display_name, portrait_id`, accountID, name, portraitID,
	).Scan(&result.DisplayName, &result.PortraitID)
	if err != nil {
		var pgErr *pgconn.PgError
		if errors.As(err, &pgErr) && pgErr.Code == "23505" {
			return character{}, ErrConflict
		}
		return character{}, err
	}
	return result, nil
}

func (s *postgresStore) playerLocation(ctx context.Context, accountID string) (string, error) {
	var locationID string
	err := s.db.QueryRow(ctx, `SELECT location_id FROM player_locations WHERE account_id = $1::uuid`, accountID).Scan(&locationID)
	if errors.Is(err, pgx.ErrNoRows) {
		return "fr/lyon/district/central", nil
	}
	return locationID, err
}

func (s *postgresStore) setPlayerLocation(ctx context.Context, accountID, locationID string) error {
	_, err := s.db.Exec(ctx, `INSERT INTO player_locations (account_id, location_id) VALUES ($1::uuid, $2)
		ON CONFLICT (account_id) DO UPDATE SET location_id = EXCLUDED.location_id, updated_at = now()`, accountID, locationID)
	return err
}

func (s *postgresStore) deleteSession(ctx context.Context, token string) error {
	hash := sha256.Sum256([]byte(token))
	_, err := s.db.Exec(ctx, `DELETE FROM player_sessions WHERE token_hash = $1`, hash[:])
	return err
}

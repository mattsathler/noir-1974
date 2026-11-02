package identity

import (
	"context"
	"errors"
	"fmt"

	"github.com/jackc/pgx/v5"
)

type hackerExposureWriter interface {
	recordHackerAction(context.Context, string, string, string, string, string, int) error
	reduceHackerExposure(context.Context, string, string, string) (bool, error)
}

type hackerExposureMessage struct {
	subject string
	body    string
}

var hackerExposureWarnings = []struct {
	threshold int
	message   hackerExposureMessage
}{
	{20, hackerExposureMessage{"Unusual traffic on the company relay", "An anonymous note was routed through a company relay: ‘Someone has noticed the traffic. Change your route while you still can.’"}},
	{45, hackerExposureMessage{"A warning from an unknown sender", "‘Several offices are comparing records of recent intrusions. Stop using this route before they connect the incidents.’"}},
	{70, hackerExposureMessage{"Corporate security is asking questions", "‘The security desks are sharing more than they should. Your last few visits were not as quiet as you think.’"}},
	{90, hackerExposureMessage{"The Crow has seen the reports", "An encrypted note from The Crow: ‘You have one chance to disappear from this route. Change proxies now. I cannot keep them off you forever.’"}},
}

func (s *postgresStore) recordHackerAction(ctx context.Context, serverID, accountID, countryID, action, eventID string, delta int) error {
	if countryID == "" || eventID == "" || delta <= 0 {
		return nil
	}
	transactional, ok := s.db.(interface {
		Begin(context.Context) (pgx.Tx, error)
	})
	if !ok {
		return errors.New("hacker exposure transactions are unavailable")
	}
	tx, err := transactional.Begin(ctx)
	if err != nil {
		return err
	}
	defer tx.Rollback(ctx)
	var inserted string
	err = tx.QueryRow(ctx, `INSERT INTO player_hacker_exposure_events(server_id,account_id,event_id,country_id,action,exposure_delta)
		VALUES($1::uuid,$2::uuid,$3,$4,$5,$6)
		ON CONFLICT(server_id,account_id,event_id) DO NOTHING RETURNING event_id`, serverID, accountID, eventID, countryID, action, delta).Scan(&inserted)
	if errors.Is(err, pgx.ErrNoRows) {
		return tx.Commit(ctx)
	}
	if err != nil {
		return err
	}
	var exposure int
	err = tx.QueryRow(ctx, `INSERT INTO player_hacker_exposure(server_id,account_id,country_id,exposure)
		VALUES($1::uuid,$2::uuid,$3,LEAST(100,$4))
		ON CONFLICT(server_id,account_id,country_id) DO UPDATE
		SET exposure=LEAST(100,player_hacker_exposure.exposure+EXCLUDED.exposure),updated_at=now()
		RETURNING exposure`, serverID, accountID, countryID, delta).Scan(&exposure)
	if err != nil {
		return err
	}
	for _, warning := range hackerExposureWarnings {
		if exposure < warning.threshold {
			continue
		}
		messageID := fmt.Sprintf("hacker-exposure:%s:%d", countryID, warning.threshold)
		if _, err := tx.Exec(ctx, `INSERT INTO player_inbox_messages(server_id,account_id,message_id,sent_at,sender,message_category,subject,body,attachments)
			VALUES($1::uuid,$2::uuid,$3,now(),'Unknown correspondent','spam',$4,$5,'[]'::jsonb)
			ON CONFLICT(server_id,account_id,message_id) DO NOTHING`, serverID, accountID, messageID, warning.message.subject, warning.message.body); err != nil {
			return err
		}
	}
	return tx.Commit(ctx)
}

func (s *postgresStore) reduceHackerExposure(ctx context.Context, serverID, accountID, countryID string) (bool, error) {
	if countryID == "" {
		return false, nil
	}
	result, err := s.db.Exec(ctx, `UPDATE player_hacker_exposure
		SET exposure=(exposure*65)/100,updated_at=now()
		WHERE server_id=$1::uuid AND account_id=$2::uuid AND country_id=$3`, serverID, accountID, countryID)
	return err == nil && result.RowsAffected() > 0, err
}

func exposureDeltaForBlackwire(category string) int {
	switch category {
	case "POLITICAL / EXECUTIVE EVIDENCE":
		return 32
	case "POLITICAL INFORMATION", "EXECUTIVE FINANCIAL EVIDENCE":
		return 26
	case "FINANCIAL / CORPORATE EVIDENCE":
		return 20
	case "EXECUTIVE INFORMATION":
		return 15
	case "CASE INFORMATION":
		return 12
	default:
		return 3
	}
}

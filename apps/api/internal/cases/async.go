package cases

import (
	"context"
	"encoding/json"
	"errors"
	"log"
	"time"

	"github.com/jackc/pgx/v5"
)

type AsyncRequestProcessor struct {
	store                *PostgresStore
	serverID             string
	interval             time.Duration
	onAccusationAccepted func(string)
	onRequestCompleted   func()
}

func NewAsyncRequestProcessor(store *PostgresStore, serverID string, interval time.Duration) *AsyncRequestProcessor {
	if interval <= 0 {
		interval = 5 * time.Second
	}
	return &AsyncRequestProcessor{store: store, serverID: serverID, interval: interval}
}

func (processor *AsyncRequestProcessor) SetAccusationAcceptedHandler(handler func(cityID string)) {
	processor.onAccusationAccepted = handler
}

func (processor *AsyncRequestProcessor) SetRequestCompletedHandler(handler func()) {
	processor.onRequestCompleted = handler
}

func (processor *AsyncRequestProcessor) Run(ctx context.Context) {
	ticker := time.NewTicker(processor.interval)
	defer ticker.Stop()
	for {
		worked, err := processor.store.ProcessNextAsyncRequest(ctx, processor.serverID, processor.onAccusationAccepted)
		if worked && err == nil && processor.onRequestCompleted != nil {
			processor.onRequestCompleted()
		}
		if err != nil && !errors.Is(err, context.Canceled) {
			log.Printf("process asynchronous player request: %v", err)
		}
		if worked {
			continue
		}
		select {
		case <-ctx.Done():
			return
		case <-ticker.C:
		}
	}
}

func (store *PostgresStore) ProcessNextAsyncRequest(ctx context.Context, serverID string, onAccusationAccepted func(string)) (bool, error) {
	tx, err := store.db.Begin(ctx)
	if err != nil {
		return false, err
	}
	var request asyncRequest
	err = tx.QueryRow(ctx, `SELECT request_id,account_id::text,request_type,payload FROM player_async_requests
		WHERE server_id=$1::uuid AND available_at<=now()
		AND (status='pending' OR (status='processing' AND lease_until<=now()))
		ORDER BY available_at,created_at LIMIT 1 FOR UPDATE SKIP LOCKED`, serverID).
		Scan(&request.RequestID, &request.AccountID, &request.RequestType, &request.Payload)
	if errors.Is(err, pgx.ErrNoRows) {
		_ = tx.Rollback(ctx)
		return false, nil
	}
	if err != nil {
		_ = tx.Rollback(ctx)
		return false, err
	}
	if _, err := tx.Exec(ctx, `UPDATE player_async_requests SET status='processing',lease_until=now()+interval '1 minute',attempts=attempts+1
		WHERE server_id=$1::uuid AND request_id=$2`, serverID, request.RequestID); err != nil {
		_ = tx.Rollback(ctx)
		return false, err
	}
	if err := tx.Commit(ctx); err != nil {
		return false, err
	}

	var processingErr error
	switch request.RequestType {
	case "residential_search":
		var payload residentialSearchPayload
		if err := json.Unmarshal(request.Payload, &payload); err != nil {
			processingErr = err
		} else {
			_, processingErr = store.completeResidentialSearch(ctx, serverID, request.AccountID, payload.CityID, payload.StationLocationID, payload.TargetNPCID, time.Now().UTC())
		}
	case "accusation":
		var payload accusationPayload
		if err := json.Unmarshal(request.Payload, &payload); err != nil {
			processingErr = err
		} else {
			var result AccusationResult
			result, processingErr = store.resolveAccusation(ctx, serverID, request.AccountID, request.RequestID, payload, time.Now().UTC())
			if processingErr == nil && result.Accepted && onAccusationAccepted != nil {
				onAccusationAccepted(payload.CityID)
			}
		}
	default:
		processingErr = errors.New("unknown asynchronous request type")
	}

	if processingErr != nil {
		_, rescheduleErr := store.db.Exec(ctx, `UPDATE player_async_requests SET status='pending',lease_until=NULL,available_at=now()+interval '30 seconds'
			WHERE server_id=$1::uuid AND request_id=$2 AND status='processing'`, serverID, request.RequestID)
		if rescheduleErr != nil {
			return true, errors.Join(processingErr, rescheduleErr)
		}
		return true, processingErr
	}
	if _, err := store.db.Exec(ctx, `UPDATE player_async_requests SET status='completed',lease_until=NULL,completed_at=now()
		WHERE server_id=$1::uuid AND request_id=$2 AND status='processing'`, serverID, request.RequestID); err != nil {
		return true, err
	}
	return true, nil
}

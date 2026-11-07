package cases

import (
	"context"
	"encoding/json"
	"fmt"
	"time"

	"github.com/jackc/pgx/v5"
)

// ProcessDuePoliceWork assigns case-local agents, collects one deterministic
// item on days two, four and six, and performs the single day-seven review.
func (store *PostgresStore) ProcessDuePoliceWork(ctx context.Context, serverID string, now time.Time) error {
	tx, err := store.db.Begin(ctx)
	if err != nil {
		return err
	}
	defer func() { _ = tx.Rollback(context.Background()) }()
	if _, err := tx.Exec(ctx, `INSERT INTO police_case_investigations(server_id,case_id,agent_id,review_at)
		SELECT c.server_id,c.case_id,'police-agent:'||substr(md5(c.case_seed||'/police-agent'),1,16),c.created_at+interval '7 days'
		FROM cases c JOIN case_states s USING(server_id,case_id)
		WHERE c.server_id=$1::uuid AND s.status='open'
		ON CONFLICT(server_id,case_id) DO NOTHING`, serverID); err != nil {
		return fmt.Errorf("assign case detectives: %w", err)
	}
	rows, err := tx.Query(ctx, `SELECT p.case_id,p.collected_item_ids,p.review_at,c.case_type,c.archetype_id,c.generator_version,COALESCE(c.victim_entity_id,''),s.body_found_at
		FROM police_case_investigations p JOIN cases c USING(server_id,case_id) JOIN case_states s USING(server_id,case_id)
		WHERE p.server_id=$1::uuid AND s.status='open' AND p.reviewed_at IS NULL ORDER BY p.case_id FOR UPDATE OF p`, serverID)
	if err != nil {
		return err
	}
	type investigation struct {
		id, caseType, archetype, victim string
		generatorVersion                int
		ids                             []string
		review                          time.Time
		bodyFound                       *time.Time
	}
	work := []investigation{}
	for rows.Next() {
		var item investigation
		var bodyFound *time.Time
		if err := rows.Scan(&item.id, &item.ids, &item.review, &item.caseType, &item.archetype, &item.generatorVersion, &item.victim, &bodyFound); err != nil {
			rows.Close()
			return err
		}
		item.bodyFound = bodyFound
		work = append(work, item)
	}
	if err := rows.Err(); err != nil {
		rows.Close()
		return err
	}
	rows.Close()
	for _, item := range work {
		age := now.Sub(item.review.Add(-7 * 24 * time.Hour))
		for slot, day := range []int{2, 4, 6} {
			if age < time.Duration(day)*24*time.Hour || len(item.ids) > slot {
				continue
			}
			var id string
			err := tx.QueryRow(ctx, `SELECT candidate.item_id FROM (
				SELECT evidence_id AS item_id FROM case_evidence WHERE server_id=$1::uuid AND case_id=$2
				UNION ALL SELECT record_id FROM case_activity_records WHERE server_id=$1::uuid AND case_id=$2
				UNION ALL SELECT message_id FROM case_mail_messages WHERE server_id=$1::uuid AND case_id=$2
			) candidate WHERE NOT(candidate.item_id=ANY($3::text[]))
			AND NOT EXISTS(SELECT 1 FROM destroyed_case_evidence d WHERE d.server_id=$1::uuid AND d.case_id=$2 AND d.item_id=candidate.item_id)
			ORDER BY md5(candidate.item_id||'/'||$2) LIMIT 1`, serverID, item.id, item.ids).Scan(&id)
			if err == nil {
				item.ids = append(item.ids, id)
			} else if err != pgx.ErrNoRows {
				return err
			}
		}
		if _, err := tx.Exec(ctx, `UPDATE police_case_investigations SET collected_item_ids=$3 WHERE server_id=$1::uuid AND case_id=$2`, serverID, item.id, item.ids); err != nil {
			return err
		}
		if now.Before(item.review) {
			continue
		}
		if err := store.reviewPoliceCase(ctx, tx, serverID, item, now); err != nil {
			return err
		}
	}
	return tx.Commit(ctx)
}

func (store *PostgresStore) reviewPoliceCase(ctx context.Context, tx pgx.Tx, serverID string, item struct {
	id, caseType, archetype, victim string
	generatorVersion                int
	ids                             []string
	review                          time.Time
	bodyFound                       *time.Time
}, now time.Time) error {
	var responsible []string
	rows, err := tx.Query(ctx, `SELECT npc_entity_id FROM case_participants WHERE server_id=$1::uuid AND case_id=$2 AND participant_role IN ('principal','accomplice') ORDER BY npc_entity_id`, serverID, item.id)
	if err != nil {
		return err
	}
	for rows.Next() {
		var id string
		if err := rows.Scan(&id); err != nil {
			rows.Close()
			return err
		}
		responsible = append(responsible, id)
	}
	if err := rows.Err(); err != nil {
		rows.Close()
		return err
	}
	rows.Close()
	sufficient := len(item.ids) > 0 && len(responsible) > 0
	for _, npcID := range responsible {
		var relevant bool
		err := tx.QueryRow(ctx, `SELECT
			EXISTS(SELECT 1 FROM case_evidence e WHERE e.server_id=$1::uuid AND e.case_id=$2 AND e.evidence_id=ANY($3::text[])
				AND NOT EXISTS(SELECT 1 FROM destroyed_case_evidence d WHERE d.server_id=e.server_id AND d.case_id=e.case_id AND d.item_id=e.evidence_id)
				AND (COALESCE(jsonb_array_length(e.evidence_data->'supportsNpcIds'),0)=0 OR e.evidence_data->'supportsNpcIds' ? $4)
				AND ((e.evidence_type='office_log' AND $5 IN ('workplace_conflict','serial_killer'))
				 OR (e.evidence_type='travel_register' AND $5='financial_extortion')
				 OR (e.evidence_type='meeting_minutes' AND $5='workplace_conflict')
				 OR (e.evidence_type='account_record' AND $5='financial_extortion')
				 OR (e.evidence_type IN ('letter_fragment','witness_statement') AND $5 IN ('jealousy','serial_killer'))
				 OR ($6>=9 AND e.evidence_type IN ($5||'_document',$5||'_account',$5||'_corroboration'))))
			OR EXISTS(SELECT 1 FROM case_activity_records a WHERE a.server_id=$1::uuid AND a.case_id=$2 AND a.record_id=ANY($3::text[])
				AND NOT EXISTS(SELECT 1 FROM destroyed_case_evidence d WHERE d.server_id=a.server_id AND d.case_id=a.case_id AND d.item_id=a.record_id)
				AND a.npc_entity_id=$4 AND ((a.source='bank' AND a.internal_relation='extortion_payment')
				 OR (a.source='gate' AND a.internal_relation IN ('scene-victim','scene-killer','scene-killer-exit','affair-victim','affair-paramour'))))
			OR EXISTS(SELECT 1 FROM case_mail_messages m WHERE m.server_id=$1::uuid AND m.case_id=$2 AND m.message_id=ANY($3::text[])
				AND NOT EXISTS(SELECT 1 FROM destroyed_case_evidence d WHERE d.server_id=m.server_id AND d.case_id=m.case_id AND d.item_id=m.message_id)
				AND (m.mailbox_npc_entity_id=$4 OR m.related_npc_entity_id=$4)
				AND m.internal_relation IN ('anonymous_warning','anonymous_followup','killer_remorse','jealousy_exchange','case_correspondence'))`, serverID, item.id, item.ids, npcID, item.archetype, item.generatorVersion).Scan(&relevant)
		if err != nil {
			return err
		}
		if !relevant {
			sufficient = false
		}
	}
	var proofStrength float64
	if len(item.ids) > 0 {
		if err := tx.QueryRow(ctx, `SELECT COALESCE(sum(proof.strength),0)::float8 FROM (
			SELECT e.evidence_id AS item_id,CASE
				WHEN (c.archetype_id IN ('workplace_conflict','serial_killer') AND e.evidence_type='office_log')
				 OR (c.archetype_id='financial_extortion' AND e.evidence_type='travel_register')
				 OR (c.generator_version>=9 AND e.evidence_type=c.archetype_id||'_document') THEN 1.0
				WHEN (c.archetype_id='workplace_conflict' AND e.evidence_type='meeting_minutes')
				 OR (c.archetype_id='financial_extortion' AND e.evidence_type='account_record')
				 OR (c.archetype_id IN ('jealousy','serial_killer') AND e.evidence_type IN ('letter_fragment','witness_statement'))
				 OR (c.generator_version>=9 AND e.evidence_type IN (c.archetype_id||'_account',c.archetype_id||'_corroboration')) THEN 0.75 ELSE 0.0 END AS strength
			FROM case_evidence e JOIN cases c USING(server_id,case_id) WHERE e.server_id=$1::uuid AND e.case_id=$2
			UNION ALL
			SELECT a.record_id,CASE WHEN (a.source='bank' AND (a.internal_relation='extortion_payment' OR (c.generator_version<4 AND c.archetype_id='financial_extortion' AND a.internal_relation='precrime_route')))
				OR (a.source='gate' AND a.internal_relation IN ('scene-victim','scene-killer','scene-killer-exit')) THEN 1.0
				WHEN a.source='gate' AND a.internal_relation IN ('affair-victim','affair-paramour') THEN 0.75 ELSE 0.0 END
			FROM case_activity_records a JOIN cases c USING(server_id,case_id) WHERE a.server_id=$1::uuid AND a.case_id=$2
			UNION ALL
			SELECT m.message_id,CASE WHEN m.internal_relation IN ('anonymous_warning','anonymous_followup','killer_remorse','jealousy_exchange','case_correspondence')
				THEN CASE WHEN c.generator_version>=4 THEN 0.75 ELSE 0.25 END ELSE 0.0 END
			FROM case_mail_messages m JOIN cases c USING(server_id,case_id) WHERE m.server_id=$1::uuid AND m.case_id=$2
		) proof WHERE proof.item_id=ANY($3::text[])
		AND NOT EXISTS(SELECT 1 FROM destroyed_case_evidence d WHERE d.server_id=$1::uuid AND d.case_id=$2 AND d.item_id=proof.item_id)`, serverID, item.id, item.ids).Scan(&proofStrength); err != nil {
			return err
		}
	}
	if proofStrength < 2.75 {
		sufficient = false
	}
	if item.caseType == "homicide" && (item.bodyFound == nil || item.bodyFound.After(now)) {
		sufficient = false
	}
	if item.caseType == "homicide" && item.victim != "" {
		var identified bool
		if err := tx.QueryRow(ctx, `SELECT
			EXISTS(SELECT 1 FROM case_evidence e WHERE e.server_id=$1::uuid AND e.case_id=$2 AND e.evidence_id=ANY($3::text[]) AND e.evidence_data->'supportsNpcIds' ? $4)
			OR EXISTS(SELECT 1 FROM case_activity_records a WHERE a.server_id=$1::uuid AND a.case_id=$2 AND a.record_id=ANY($3::text[]) AND (a.npc_entity_id=$4 OR a.related_npc_entity_id=$4))
			OR EXISTS(SELECT 1 FROM case_mail_messages m WHERE m.server_id=$1::uuid AND m.case_id=$2 AND m.message_id=ANY($3::text[]) AND (m.mailbox_npc_entity_id=$4 OR m.related_npc_entity_id=$4))`, serverID, item.id, item.ids, item.victim).Scan(&identified); err != nil {
			return err
		}
		if !identified {
			sufficient = false
		}
	}
	if sufficient {
		for _, npcID := range responsible {
			payload, _ := json.Marshal("arrested")
			if _, err := tx.Exec(ctx, `INSERT INTO world_mutations(server_id,entity_id,field_key,value,created_at) VALUES($1::uuid,$2,'status',$3::jsonb,$4)`, serverID, npcID, string(payload), now.UTC()); err != nil {
				return err
			}
		}
		if _, err := tx.Exec(ctx, `UPDATE case_states SET status='solved',resolution_source='police',updated_at=$3 WHERE server_id=$1::uuid AND case_id=$2 AND status='open'`, serverID, item.id, now.UTC()); err != nil {
			return err
		}
		if err := store.applyCaseCompletionMarketImpact(ctx, tx, serverID, item.id, now.UTC()); err != nil {
			return fmt.Errorf("apply police case completion impact: %w", err)
		}
	}
	outcome := "insufficient"
	if sufficient {
		outcome = "solved"
	}
	_, err = tx.Exec(ctx, `UPDATE police_case_investigations SET reviewed_at=$3,outcome=$4 WHERE server_id=$1::uuid AND case_id=$2 AND reviewed_at IS NULL`, serverID, item.id, now.UTC(), outcome)
	return err
}

package cases

import (
	"context"
	"embed"
	"encoding/json"
	"errors"
	"fmt"
	"sort"
	"strings"
	"time"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgconn"
	"noir-1974/apps/api/internal/mail"
	"noir-1974/apps/api/internal/records"
	"noir-1974/apps/api/internal/world"
)

//go:embed schema.sql
var caseSchema embed.FS

var (
	ErrCaseNotFound      = errors.New("case not found")
	ErrInvalidCaseState  = errors.New("invalid shared case state")
	ErrInvalidProgress   = errors.New("invalid player case progress")
	ErrEvidenceNotFound  = errors.New("case evidence not found")
	ErrPersistedConflict = errors.New("persisted case does not match generated definition")
)

type caseDatabase interface {
	Exec(context.Context, string, ...any) (pgconn.CommandTag, error)
	Query(context.Context, string, ...any) (pgx.Rows, error)
	QueryRow(context.Context, string, ...any) pgx.Row
	Begin(context.Context) (pgx.Tx, error)
}

type PostgresStore struct {
	db caseDatabase
}

type SharedState struct {
	CaseID string `json:"caseId"`
	Status string `json:"status"`
}

type PlayerProgress struct {
	DiscoveredEvidenceIDs []string `json:"discoveredEvidenceIds"`
}

// PublicCase contains only facts safe for the newspaper headline.
type PublicCase struct {
	ID         string    `json:"id"`
	CaseType   string    `json:"caseType"`
	VictimID   string    `json:"victimId"`
	LocationID string    `json:"locationId"`
	Status     string    `json:"status"`
	CreatedAt  time.Time `json:"createdAt"`
}

type PublicEvidence struct {
	ID           string `json:"id"`
	Presentation string `json:"presentation"`
	Content      string `json:"content"`
	LocationID   string `json:"locationId"`
}

func NewPostgresStore(db caseDatabase) *PostgresStore {
	return &PostgresStore{db: db}
}

func (store *PostgresStore) Migrate(ctx context.Context) error {
	contents, err := caseSchema.ReadFile("schema.sql")
	if err != nil {
		return err
	}
	for _, statement := range strings.Split(string(contents), ";") {
		if strings.TrimSpace(statement) == "" {
			continue
		}
		if _, err := store.db.Exec(ctx, statement); err != nil {
			return fmt.Errorf("apply case schema: %w", err)
		}
	}
	return nil
}

// Create persists a compact case definition and its relationships and evidence references.
func (store *PostgresStore) Create(ctx context.Context, serverID string, base world.World, definition Definition) error {
	if err := Validate(definition, base); err != nil {
		return err
	}
	tx, err := store.db.Begin(ctx)
	if err != nil {
		return fmt.Errorf("begin case creation: %w", err)
	}
	defer func() { _ = tx.Rollback(context.Background()) }()

	var existingSeed, existingType, existingArchetype, existingMotive, existingVictim, existingKiller, existingLocation string
	var existingVersion int
	err = tx.QueryRow(ctx,
		`SELECT case_seed, generator_version, case_type, archetype_id, motive, victim_entity_id, killer_entity_id, location_entity_id
		 FROM cases WHERE server_id = $1::uuid AND case_id = $2 FOR UPDATE`, serverID, definition.ID,
	).Scan(&existingSeed, &existingVersion, &existingType, &existingArchetype, &existingMotive, &existingVictim, &existingKiller, &existingLocation)
	if err == nil {
		if existingSeed != definition.Seed || existingVersion != definition.GeneratorVersion || existingType != definition.CaseType || existingArchetype != definition.ArchetypeID || existingMotive != definition.Motive || existingVictim != definition.VictimID || existingKiller != definition.KillerID || existingLocation != definition.LocationID {
			return ErrPersistedConflict
		}
		if err := persistCaseEmails(ctx, tx, serverID, definition); err != nil {
			return err
		}
		if err := tx.Commit(ctx); err != nil {
			return fmt.Errorf("commit existing case check: %w", err)
		}
		return nil
	}
	if !errors.Is(err, pgx.ErrNoRows) {
		return fmt.Errorf("check existing case: %w", err)
	}

	var entityCount int
	if err := tx.QueryRow(ctx,
		`SELECT COUNT(*) FROM world_entities
		 WHERE server_id = $1::uuid
		   AND ((entity_id = $2 AND entity_type = 'npc')
		     OR (entity_id = $3 AND entity_type = 'npc')
		     OR (entity_id = $4 AND entity_type = 'location'))`,
		serverID, definition.KillerID, definition.VictimID, definition.LocationID,
	).Scan(&entityCount); err != nil {
		return fmt.Errorf("verify case entities: %w", err)
	}
	if entityCount != 3 {
		return world.ErrEntityNotFound
	}

	relationshipIDs := make([]string, 0, len(definition.Relationships))
	for _, relation := range definition.Relationships {
		relationshipIDs = append(relationshipIDs, relation.ID)
	}
	for _, relation := range definition.Relationships {
		var query string
		switch relation.Type {
		case "employment":
			query = `SELECT EXISTS (SELECT 1 FROM employment_relationships WHERE server_id = $1::uuid AND relationship_id = $2)`
		case "personal":
			query = `SELECT EXISTS (SELECT 1 FROM personal_relationships WHERE server_id = $1::uuid AND relationship_id = $2)`
		default:
			return ErrPersistedConflict
		}
		var exists bool
		if err := tx.QueryRow(ctx, query, serverID, relation.ID).Scan(&exists); err != nil {
			return fmt.Errorf("verify case relationship %s: %w", relation.ID, err)
		}
		if !exists {
			return ErrPersistedConflict
		}
	}

	if _, err := tx.Exec(ctx,
		`INSERT INTO cases (server_id, case_id, case_seed, generator_version, case_type, archetype_id, motive,
		 victim_entity_id, killer_entity_id, location_entity_id)
		 VALUES ($1::uuid, $2, $3, $4, $5, $6, $7, $8, $9, $10)`,
		serverID, definition.ID, definition.Seed, definition.GeneratorVersion, definition.CaseType, definition.ArchetypeID, definition.Motive,
		definition.VictimID, definition.KillerID, definition.LocationID,
	); err != nil {
		return fmt.Errorf("persist case definition: %w", err)
	}
	for _, relation := range definition.Relationships {
		if _, err := tx.Exec(ctx,
			`INSERT INTO case_relationships (server_id, case_id, relationship_id, relationship_type, relationship_role)
			 VALUES ($1::uuid, $2, $3, $4, $5)`,
			serverID, definition.ID, relation.ID, relation.Type, relation.Role,
		); err != nil {
			return fmt.Errorf("persist case relationship: %w", err)
		}
	}
	for _, evidence := range definition.Evidence {
		data, err := json.Marshal(struct {
			Presentation           string   `json:"presentation"`
			Content                string   `json:"content"`
			RelationshipIDs        []string `json:"relationshipIds"`
			ReachableFromEntityIDs []string `json:"reachableFromEntityIds"`
		}{evidence.Presentation, evidence.Content, evidence.RelationshipIDs, evidence.ReachableFromEntityIDs})
		if err != nil {
			return fmt.Errorf("encode case evidence: %w", err)
		}
		if _, err := tx.Exec(ctx,
			`INSERT INTO case_evidence (server_id, case_id, evidence_id, location_entity_id, evidence_type, evidence_data)
			 VALUES ($1::uuid, $2, $3, $4, $5, $6::jsonb)`,
			serverID, definition.ID, evidence.ID, evidence.LocationID, evidence.Type, string(data),
		); err != nil {
			return fmt.Errorf("persist case evidence: %w", err)
		}
	}
	for _, record := range definition.Records {
		data, err := json.Marshal(record)
		if err != nil {
			return fmt.Errorf("encode case activity record: %w", err)
		}
		if _, err := tx.Exec(ctx,
			`INSERT INTO case_activity_records (server_id, case_id, record_id, npc_entity_id, related_npc_entity_id, source, occurred_at, internal_relation, record_data)
			 VALUES ($1::uuid, $2, $3, $4, NULLIF($5, ''), $6, $7, $8, $9::jsonb)`,
			serverID, definition.ID, record.ID, record.NPCID, record.RelatedNPCID, record.Source, record.OccurredAt, record.InternalRelation, string(data),
		); err != nil {
			return fmt.Errorf("persist case activity record %s: %w", record.ID, err)
		}
	}
	if err := persistCaseEmails(ctx, tx, serverID, definition); err != nil {
		return err
	}
	if _, err := tx.Exec(ctx,
		`INSERT INTO case_states (server_id, case_id, status) VALUES ($1::uuid, $2, 'open')`,
		serverID, definition.ID,
	); err != nil {
		return fmt.Errorf("persist initial shared case state: %w", err)
	}
	if err := tx.Commit(ctx); err != nil {
		return fmt.Errorf("commit case creation: %w", err)
	}
	return nil
}

func persistCaseEmails(ctx context.Context, tx pgx.Tx, serverID string, definition Definition) error {
	for _, message := range definition.Emails {
		data, err := json.Marshal(message)
		if err != nil {
			return fmt.Errorf("encode case email: %w", err)
		}
		if _, err := tx.Exec(ctx,
			`INSERT INTO case_mail_messages (server_id, case_id, message_id, mailbox_npc_entity_id, related_npc_entity_id, sent_at, internal_relation, message_data)
			 VALUES ($1::uuid, $2, $3, $4, $5, $6, $7, $8::jsonb)
			 ON CONFLICT (server_id, message_id) DO NOTHING`,
			serverID, definition.ID, message.ID, message.MailboxNPCID, message.RelatedNPCID, message.SentAt, message.InternalRelation, string(data),
		); err != nil {
			return fmt.Errorf("persist case email %s: %w", message.ID, err)
		}
	}
	return nil
}

// NPCRecords loads only the selected person's case-related entries within the
// requested interval. It deliberately omits the internal relationship label.
func (store *PostgresStore) NPCRecords(ctx context.Context, serverID, npcID string, from, through time.Time) ([]records.Record, error) {
	rows, err := store.db.Query(ctx,
		`SELECT record_data FROM case_activity_records
		 WHERE server_id = $1::uuid AND npc_entity_id = $2 AND occurred_at >= $3 AND occurred_at <= $4
		 ORDER BY occurred_at, record_id`, serverID, npcID, from, through,
	)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	result := make([]records.Record, 0)
	for rows.Next() {
		var data []byte
		var item records.Record
		if err := rows.Scan(&data); err != nil {
			return nil, err
		}
		if err := json.Unmarshal(data, &item); err != nil {
			return nil, fmt.Errorf("decode case activity record: %w", err)
		}
		result = append(result, item)
	}
	return result, rows.Err()
}

// NPCEmails loads messages sent or received by the requested NPC and interval
// while retaining case classification as backend-only metadata.
func (store *PostgresStore) NPCEmails(ctx context.Context, serverID, npcID string, from, through time.Time) ([]mail.Message, error) {
	rows, err := store.db.Query(ctx,
		`SELECT case_id, related_npc_entity_id, internal_relation, message_data
		 FROM case_mail_messages
		 WHERE server_id = $1::uuid AND (mailbox_npc_entity_id = $2 OR related_npc_entity_id = $2)
		   AND sent_at >= $3 AND sent_at <= $4
		 ORDER BY sent_at, message_id`, serverID, npcID, from, through,
	)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	result := make([]mail.Message, 0)
	for rows.Next() {
		var item mail.Message
		var data []byte
		if err := rows.Scan(&item.CaseID, &item.RelatedNPCID, &item.InternalRelation, &data); err != nil {
			return nil, err
		}
		if err := json.Unmarshal(data, &item); err != nil {
			return nil, fmt.Errorf("decode case email: %w", err)
		}
		item.MailboxNPCID = npcID
		result = append(result, item)
	}
	return result, rows.Err()
}

// LatestPublicCase returns the newest case without disclosing its killer.
func (store *PostgresStore) LatestPublicCase(ctx context.Context, serverID string) (PublicCase, error) {
	var result PublicCase
	err := store.db.QueryRow(ctx,
		`SELECT c.case_id, c.case_type, c.victim_entity_id, c.location_entity_id, s.status, c.created_at
		 FROM cases c JOIN case_states s USING (server_id, case_id)
		 WHERE c.server_id = $1::uuid ORDER BY c.created_at DESC, c.case_id DESC LIMIT 1`, serverID,
	).Scan(&result.ID, &result.CaseType, &result.VictimID, &result.LocationID, &result.Status, &result.CreatedAt)
	if errors.Is(err, pgx.ErrNoRows) {
		return PublicCase{}, ErrCaseNotFound
	}
	return result, err
}

// ListOpenPublicCases returns active cases without exposing their solutions.
func (store *PostgresStore) ListOpenPublicCases(ctx context.Context, serverID string) ([]PublicCase, error) {
	rows, err := store.db.Query(ctx,
		`SELECT c.case_id, c.case_type, c.victim_entity_id, c.location_entity_id, s.status, c.created_at
		 FROM cases c JOIN case_states s USING (server_id, case_id)
		 WHERE c.server_id = $1::uuid AND s.status = 'open'
		 ORDER BY c.created_at DESC, c.case_id DESC`, serverID,
	)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	results := make([]PublicCase, 0)
	for rows.Next() {
		var result PublicCase
		if err := rows.Scan(&result.ID, &result.CaseType, &result.VictimID, &result.LocationID, &result.Status, &result.CreatedAt); err != nil {
			return nil, err
		}
		results = append(results, result)
	}
	return results, rows.Err()
}

// GetPublicCase returns narrative evidence without internal classifications or solution data.
func (store *PostgresStore) GetPublicCase(ctx context.Context, serverID, caseID string) (PublicCase, []PublicEvidence, error) {
	var result PublicCase
	err := store.db.QueryRow(ctx,
		`SELECT c.case_id, c.case_type, c.victim_entity_id, c.location_entity_id, s.status, c.created_at
		 FROM cases c JOIN case_states s USING (server_id, case_id)
		 WHERE c.server_id = $1::uuid AND c.case_id = $2`, serverID, caseID,
	).Scan(&result.ID, &result.CaseType, &result.VictimID, &result.LocationID, &result.Status, &result.CreatedAt)
	if errors.Is(err, pgx.ErrNoRows) {
		return PublicCase{}, nil, ErrCaseNotFound
	}
	if err != nil {
		return PublicCase{}, nil, err
	}
	rows, err := store.db.Query(ctx,
		`SELECT evidence_id, location_entity_id, evidence_data
		 FROM case_evidence WHERE server_id = $1::uuid AND case_id = $2 ORDER BY evidence_id`, serverID, caseID,
	)
	if err != nil {
		return PublicCase{}, nil, err
	}
	defer rows.Close()
	evidence := make([]PublicEvidence, 0)
	for rows.Next() {
		var item PublicEvidence
		var data []byte
		if err := rows.Scan(&item.ID, &item.LocationID, &data); err != nil {
			return PublicCase{}, nil, err
		}
		var narrative struct {
			Presentation string `json:"presentation"`
			Content      string `json:"content"`
		}
		if err := json.Unmarshal(data, &narrative); err != nil {
			return PublicCase{}, nil, fmt.Errorf("decode public evidence: %w", err)
		}
		item.Presentation, item.Content = narrative.Presentation, narrative.Content
		if item.Presentation == "" {
			item.Presentation = "CASE RECORD"
		}
		if item.Content == "" {
			item.Content = "A record has been filed in this case."
		}
		evidence = append(evidence, item)
	}
	if err := rows.Err(); err != nil {
		return PublicCase{}, nil, err
	}
	return result, evidence, nil
}

func (store *PostgresStore) GetSharedState(ctx context.Context, serverID, caseID string) (SharedState, error) {
	var state SharedState
	state.CaseID = caseID
	err := store.db.QueryRow(ctx,
		`SELECT status FROM case_states WHERE server_id = $1::uuid AND case_id = $2`,
		serverID, caseID,
	).Scan(&state.Status)
	if errors.Is(err, pgx.ErrNoRows) {
		return SharedState{}, ErrCaseNotFound
	}
	return state, err
}

func (store *PostgresStore) SetSharedStatus(ctx context.Context, serverID, caseID, status string) error {
	if status != "open" && status != "solved" && status != "archived" {
		return ErrInvalidCaseState
	}
	result, err := store.db.Exec(ctx,
		`UPDATE case_states SET status = $3, updated_at = now()
		 WHERE server_id = $1::uuid AND case_id = $2`,
		serverID, caseID, status,
	)
	if err != nil {
		return err
	}
	if result.RowsAffected() != 1 {
		return ErrCaseNotFound
	}
	return nil
}

func (store *PostgresStore) GetPlayerProgress(ctx context.Context, serverID, caseID, accountID string) (PlayerProgress, error) {
	progress := PlayerProgress{DiscoveredEvidenceIDs: make([]string, 0)}
	var data []byte
	err := store.db.QueryRow(ctx,
		`SELECT progress FROM player_case_progress
		 WHERE server_id = $1::uuid AND case_id = $2 AND account_id = $3::uuid`,
		serverID, caseID, accountID,
	).Scan(&data)
	if errors.Is(err, pgx.ErrNoRows) {
		return progress, nil
	}
	if err != nil {
		return PlayerProgress{}, err
	}
	if err := json.Unmarshal(data, &progress); err != nil {
		return PlayerProgress{}, fmt.Errorf("decode player case progress: %w", err)
	}
	if progress.DiscoveredEvidenceIDs == nil {
		progress.DiscoveredEvidenceIDs = make([]string, 0)
	}
	return progress, nil
}

func (store *PostgresStore) SavePlayerProgress(ctx context.Context, serverID, caseID, accountID string, progress PlayerProgress) error {
	normalized, err := normalizeProgress(progress)
	if err != nil {
		return err
	}
	if len(normalized.DiscoveredEvidenceIDs) > 0 {
		var count int
		if err := store.db.QueryRow(ctx,
			`SELECT COUNT(*) FROM case_evidence
			 WHERE server_id = $1::uuid AND case_id = $2 AND evidence_id = ANY($3::text[])`,
			serverID, caseID, normalized.DiscoveredEvidenceIDs,
		).Scan(&count); err != nil {
			return err
		}
		if count != len(normalized.DiscoveredEvidenceIDs) {
			return ErrEvidenceNotFound
		}
	}
	data, err := json.Marshal(normalized)
	if err != nil {
		return err
	}
	_, err = store.db.Exec(ctx,
		`INSERT INTO player_case_progress (server_id, case_id, account_id, progress)
		 VALUES ($1::uuid, $2, $3::uuid, $4::jsonb)
		 ON CONFLICT (server_id, case_id, account_id)
		 DO UPDATE SET progress = EXCLUDED.progress, updated_at = now()`,
		serverID, caseID, accountID, string(data),
	)
	return err
}

func normalizeProgress(progress PlayerProgress) (PlayerProgress, error) {
	ids := append([]string(nil), progress.DiscoveredEvidenceIDs...)
	for _, id := range ids {
		if strings.TrimSpace(id) == "" {
			return PlayerProgress{}, ErrInvalidProgress
		}
	}
	sort.Strings(ids)
	unique := ids[:0]
	for _, id := range ids {
		if len(unique) == 0 || unique[len(unique)-1] != id {
			unique = append(unique, id)
		}
	}
	return PlayerProgress{DiscoveredEvidenceIDs: unique}, nil
}

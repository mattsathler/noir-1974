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

//go:embed schema.sql interrogation/*.json
var caseSchema embed.FS

var (
	ErrCaseNotFound          = errors.New("case not found")
	ErrInvalidCaseState      = errors.New("invalid shared case state")
	ErrInvalidProgress       = errors.New("invalid player case progress")
	ErrInvalidWhiteboardNote = errors.New("invalid whiteboard note")
	ErrEvidenceNotFound      = errors.New("case evidence not found")
	ErrPersistedConflict     = errors.New("persisted case does not match generated definition")
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
	DiscoveredEvidenceIDs []string               `json:"discoveredEvidenceIds"`
	WhiteboardPins        []string               `json:"whiteboardPins"`
	WhiteboardConnections []WhiteboardConnection `json:"whiteboardConnections"`
}

type WhiteboardConnection struct {
	FromID string `json:"fromId"`
	ToID   string `json:"toId"`
}

type DiscoveredItem struct {
	ID       string          `json:"id"`
	Source   string          `json:"source"`
	Title    string          `json:"title"`
	Content  string          `json:"content"`
	Metadata json.RawMessage `json:"metadata,omitempty"`
}

type WhiteboardPin struct {
	DiscoveredItem
	X int `json:"x"`
	Y int `json:"y"`
}

type Whiteboard struct {
	ID          string                 `json:"id"`
	Title       string                 `json:"title"`
	CreatedAt   time.Time              `json:"createdAt"`
	Pins        []WhiteboardPin        `json:"pins"`
	Connections []WhiteboardConnection `json:"connections"`
}

// PublicCase contains only facts safe for the newspaper headline.
type PublicCase struct {
	ID         string    `json:"id"`
	CityID     string    `json:"cityId"`
	CaseType   string    `json:"caseType"`
	VictimID   string    `json:"victimId"`
	LocationID string    `json:"-"`
	Status     string    `json:"status"`
	CreatedAt  time.Time `json:"createdAt"`
}

// PersonSearchScenes returns scene addresses only for case victims in the requested city.
func (store *PostgresStore) PersonSearchScenes(ctx context.Context, serverID, cityID string, npcIDs []string) (map[string]world.PersonSearchScene, error) {
	scenes := make(map[string]world.PersonSearchScene)
	if len(npcIDs) == 0 {
		return scenes, nil
	}
	rows, err := store.db.Query(ctx, `
		SELECT DISTINCT ON (c.victim_entity_id) c.victim_entity_id, c.location_entity_id
		FROM cases c JOIN case_states s USING (server_id, case_id)
		WHERE c.server_id = $1::uuid AND c.city_id = $2
		  AND s.status IN ('open', 'solved') AND c.victim_entity_id = ANY($3::text[])
		ORDER BY c.victim_entity_id, c.created_at DESC, c.case_id DESC`, serverID, cityID, npcIDs)
	if err != nil {
		return nil, err
	}
	locationsByVictim := make(map[string]string)
	for rows.Next() {
		var victimID, locationID string
		if err := rows.Scan(&victimID, &locationID); err != nil {
			rows.Close()
			return nil, err
		}
		locationsByVictim[victimID] = locationID
	}
	if err := rows.Err(); err != nil {
		rows.Close()
		return nil, err
	}
	rows.Close()
	if len(locationsByVictim) == 0 {
		return scenes, nil
	}
	var seed string
	var version int
	if err := store.db.QueryRow(ctx, `SELECT world_seed, generator_version FROM worlds WHERE server_id = $1::uuid`, serverID).Scan(&seed, &version); err != nil {
		return nil, err
	}
	cityVersion := version
	if cityID != "fr/lyon" {
		err := store.db.QueryRow(ctx, `SELECT generator_version FROM world_city_generators WHERE server_id = $1::uuid AND city_id = $2`, serverID, cityID).Scan(&cityVersion)
		if errors.Is(err, pgx.ErrNoRows) {
			if version < world.LatestGeneratorVersion {
				cityVersion = world.LatestGeneratorVersion
			}
		} else if err != nil {
			return nil, err
		}
	}
	generated, err := world.GenerateCity(seed, cityVersion, cityID)
	if err != nil {
		return nil, err
	}
	entities := make(map[string]world.Entity, len(generated.Entities))
	for _, entity := range generated.Entities {
		entities[entity.ID] = entity
	}
	regions := make(map[string]string, len(generated.Regions))
	for _, region := range generated.Regions {
		regions[region.ID] = region.Name
	}
	city, _ := world.CityByID(cityID)
	for victimID, locationID := range locationsByVictim {
		entity, ok := entities[locationID]
		if !ok {
			continue
		}
		address := strings.Trim(strings.Join([]string{entity.Name, regions[entity.RegionID], city.Name, city.Country}, ", "), ", ")
		scenes[victimID] = world.PersonSearchScene{Name: entity.Name, Address: address}
	}
	return scenes, nil
}

type PublicEvidence struct {
	ID           string `json:"id"`
	Type         string `json:"type"`
	Presentation string `json:"presentation"`
	Content      string `json:"content"`
	LocationID   string `json:"locationId"`
}

type InterrogationAnswers map[string][]string

type interrogationDialogue struct {
	Routine    []string `json:"routine"`
	Associates []string `json:"associates"`
	Unusual    []string `json:"unusual"`
	Work       []string `json:"work"`
	Worries    []string `json:"worries"`
}

type interrogationCaseFile struct {
	Shared interrogationDialogue            `json:"shared"`
	Roles  map[string]interrogationDialogue `json:"roles"`
}

func (dialogue interrogationDialogue) answers() InterrogationAnswers {
	return InterrogationAnswers{
		"routine":    dialogue.Routine,
		"associates": dialogue.Associates,
		"unusual":    dialogue.Unusual,
		"work":       dialogue.Work,
		"worries":    dialogue.Worries,
	}
}

// InterrogationAnswersForNPC returns authored dialogue for open cases connected to this NPC.
// The response contains only spoken lines; internal case roles and culprit IDs stay server-side.
func (store *PostgresStore) InterrogationAnswersForNPC(ctx context.Context, serverID, cityID, npcID string) (InterrogationAnswers, error) {
	data, err := caseSchema.ReadFile("interrogation/default.json")
	if err != nil {
		return nil, err
	}
	var defaultDialogue interrogationDialogue
	if err := json.Unmarshal(data, &defaultDialogue); err != nil {
		return nil, err
	}
	entries, err := caseSchema.ReadDir("interrogation")
	if err != nil {
		return nil, err
	}
	caseFiles := make(map[string]interrogationCaseFile, len(entries)-1)
	for _, entry := range entries {
		name := entry.Name()
		if entry.IsDir() || name == "default.json" || !strings.HasSuffix(name, ".json") {
			continue
		}
		archetype := strings.TrimSuffix(name, ".json")
		caseData, err := caseSchema.ReadFile("interrogation/" + name)
		if err != nil {
			return nil, err
		}
		var caseFile interrogationCaseFile
		if err := json.Unmarshal(caseData, &caseFile); err != nil {
			return nil, err
		}
		caseFiles[archetype] = caseFile
	}
	answers := defaultDialogue.answers()
	var seed string
	var version int
	if err := store.db.QueryRow(ctx, `SELECT world_seed, generator_version FROM worlds WHERE server_id = $1::uuid`, serverID).Scan(&seed, &version); err != nil {
		return nil, err
	}
	rows, err := store.db.Query(ctx, `
		SELECT c.archetype_id,
		       victim.generation_key, killer.generation_key,
	       CASE WHEN c.killer_entity_id = $3 THEN 'perpetrator'
	            WHEN EXISTS (
	              SELECT 1 FROM employment_relationships victim_job
	              JOIN employment_relationships caller_job
	                ON caller_job.server_id = victim_job.server_id
	               AND caller_job.employer_company_id = victim_job.employer_company_id
	              WHERE victim_job.server_id = c.server_id
	                AND victim_job.employee_npc_id = c.victim_entity_id
	                AND caller_job.employee_npc_id = $3
	            ) THEN 'colleague' ELSE 'acquaintance' END
		FROM cases c
	JOIN case_states s USING (server_id, case_id)
	JOIN world_entities victim ON victim.server_id = c.server_id AND victim.entity_id = c.victim_entity_id
	JOIN world_entities killer ON killer.server_id = c.server_id AND killer.entity_id = c.killer_entity_id
	WHERE c.server_id = $1::uuid AND c.city_id = $2 AND s.status = 'open'
	  AND (c.killer_entity_id = $3 OR EXISTS (
	    SELECT 1 FROM employment_relationships victim_job
	    JOIN employment_relationships caller_job
	      ON caller_job.server_id = victim_job.server_id
	     AND caller_job.employer_company_id = victim_job.employer_company_id
	    WHERE victim_job.server_id = c.server_id
	      AND victim_job.employee_npc_id = c.victim_entity_id
	      AND caller_job.employee_npc_id = $3
	  ) OR EXISTS (
	    SELECT 1 FROM personal_relationships relation
	    WHERE relation.server_id = c.server_id
	      AND ((relation.person_a_npc_id = $3 AND relation.person_b_npc_id = c.victim_entity_id)
	        OR (relation.person_b_npc_id = $3 AND relation.person_a_npc_id = c.victim_entity_id))
	  ))
	ORDER BY c.created_at DESC, c.case_id`, serverID, cityID, npcID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	contextFound := false
	for rows.Next() {
		var archetype, victimKey, killerKey, role string
		if err := rows.Scan(&archetype, &victimKey, &killerKey, &role); err != nil {
			return nil, err
		}
		victim, err := world.GenerateNPC(seed, version, victimKey)
		if err != nil {
			return nil, err
		}
		killer, err := world.GenerateNPC(seed, version, killerKey)
		if err != nil {
			return nil, err
		}
		caseFile := caseFiles[archetype]
		branch := caseFile.Roles[role]
		if !contextFound {
			answers = make(InterrogationAnswers, 5)
			for question := range defaultDialogue.answers() {
				answers[question] = []string{}
			}
			contextFound = true
		}
		mergeInterrogationAnswers(answers, caseFile.Shared.answers(), map[string]string{"victim": victim.Name})
		mergeInterrogationAnswers(answers, branch.answers(), map[string]string{"victim": victim.Name, "suspect": killer.Name})
	}
	if err := rows.Err(); err != nil {
		return nil, err
	}
	return answers, nil
}

func mergeInterrogationAnswers(target, source InterrogationAnswers, names map[string]string) {
	for question, lines := range source {
		for _, line := range lines {
			for token, name := range names {
				line = strings.ReplaceAll(line, "{"+token+"}", name)
			}
			target[question] = append(target[question], line)
		}
	}
}

func NewPostgresStore(db caseDatabase) *PostgresStore {
	return &PostgresStore{db: db}
}

func (store *PostgresStore) EnsureGenerationSchedule(ctx context.Context, serverID string, interval time.Duration) error {
	_, err := store.db.Exec(ctx, `INSERT INTO case_generation_schedule(server_id,next_run_at) VALUES($1::uuid,now()+make_interval(secs=>$2::double precision)) ON CONFLICT(server_id) DO NOTHING`, serverID, interval.Seconds())
	return err
}

func (store *PostgresStore) DueGeneration(ctx context.Context, serverID string) (int64, bool, error) {
	var slot int64
	var due bool
	err := store.db.QueryRow(ctx, `SELECT next_slot,next_run_at<=now() FROM case_generation_schedule WHERE server_id=$1::uuid`, serverID).Scan(&slot, &due)
	return slot, due, err
}

func (store *PostgresStore) CountOpenCases(ctx context.Context, serverID, cityID string) (int, error) {
	var count int
	err := store.db.QueryRow(ctx, `SELECT count(*) FROM case_states s JOIN cases c USING(server_id,case_id) WHERE s.server_id=$1::uuid AND c.city_id=$2 AND s.status='open'`, serverID, cityID).Scan(&count)
	return count, err
}

func (store *PostgresStore) UsedNPCIDs(ctx context.Context, serverID, cityID, currentSeed string) (map[string]struct{}, error) {
	rows, err := store.db.Query(ctx, `SELECT victim_entity_id FROM cases WHERE server_id=$1::uuid AND city_id=$2 AND case_seed<>$3 UNION SELECT killer_entity_id FROM cases WHERE server_id=$1::uuid AND city_id=$2 AND case_seed<>$3`, serverID, cityID, currentSeed)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	used := map[string]struct{}{}
	for rows.Next() {
		var id string
		if err := rows.Scan(&id); err != nil {
			return nil, err
		}
		used[id] = struct{}{}
	}
	return used, rows.Err()
}

func (store *PostgresStore) AdvanceGeneration(ctx context.Context, serverID string, slot int64, interval time.Duration) error {
	result, err := store.db.Exec(ctx, `UPDATE case_generation_schedule SET next_slot=next_slot+1,next_run_at=GREATEST(next_run_at+make_interval(secs=>$3::double precision),now()+make_interval(secs=>$3::double precision)),updated_at=now() WHERE server_id=$1::uuid AND next_slot=$2`, serverID, slot, interval.Seconds())
	if err != nil {
		return err
	}
	if result.RowsAffected() != 1 {
		return errors.New("case generation slot changed")
	}
	return nil
}

func (store *PostgresStore) DeferGeneration(ctx context.Context, serverID string, slot int64, interval time.Duration) error {
	_, err := store.db.Exec(ctx, `UPDATE case_generation_schedule SET next_run_at=now()+make_interval(secs=>$3::double precision),updated_at=now() WHERE server_id=$1::uuid AND next_slot=$2`, serverID, slot, interval.Seconds())
	return err
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

	var existingSeed, existingCityID, existingType, existingArchetype, existingMotive, existingVictim, existingKiller, existingLocation string
	var existingVersion int
	err = tx.QueryRow(ctx,
		`SELECT case_seed, city_id, generator_version, case_type, archetype_id, motive, victim_entity_id, killer_entity_id, location_entity_id
		 FROM cases WHERE server_id = $1::uuid AND case_id = $2 FOR UPDATE`, serverID, definition.ID,
	).Scan(&existingSeed, &existingCityID, &existingVersion, &existingType, &existingArchetype, &existingMotive, &existingVictim, &existingKiller, &existingLocation)
	if err == nil {
		if existingSeed != definition.Seed || existingCityID != definition.CityID || existingVersion != definition.GeneratorVersion || existingType != definition.CaseType || existingArchetype != definition.ArchetypeID || existingMotive != definition.Motive || existingVictim != definition.VictimID || existingKiller != definition.KillerID || existingLocation != definition.LocationID {
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
		`INSERT INTO cases (server_id, case_id, case_seed, city_id, generator_version, case_type, archetype_id, motive,
		 victim_entity_id, killer_entity_id, location_entity_id)
		 VALUES ($1::uuid, $2, $3, $4, $5, $6, $7, $8, $9, $10, $11)`,
		serverID, definition.ID, definition.Seed, definition.CityID, definition.GeneratorVersion, definition.CaseType, definition.ArchetypeID, definition.Motive,
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
func (store *PostgresStore) NPCRecords(ctx context.Context, serverID, npcID string, from, through time.Time, source string) ([]records.Record, error) {
	query := `SELECT record_data FROM case_activity_records
		 WHERE server_id = $1::uuid AND npc_entity_id = $2 AND occurred_at >= $3 AND occurred_at <= $4
		 ORDER BY occurred_at, record_id`
	args := []any{serverID, npcID, from, through}
	if source != "" {
		query = `SELECT record_data FROM case_activity_records
			 WHERE server_id = $1::uuid AND npc_entity_id = $2 AND source = $5 AND occurred_at >= $3 AND occurred_at <= $4
			 ORDER BY occurred_at, record_id`
		args = append(args, source)
	}
	rows, err := store.db.Query(ctx, query, args...)
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
func (store *PostgresStore) LatestPublicCase(ctx context.Context, serverID, cityID string) (PublicCase, error) {
	var result PublicCase
	err := store.db.QueryRow(ctx,
		`SELECT c.case_id, c.city_id, c.case_type, c.victim_entity_id, c.location_entity_id, s.status, c.created_at
		 FROM cases c JOIN case_states s USING (server_id, case_id)
		 WHERE c.server_id = $1::uuid AND c.city_id = $2 ORDER BY c.created_at DESC, c.case_id DESC LIMIT 1`, serverID, cityID,
	).Scan(&result.ID, &result.CityID, &result.CaseType, &result.VictimID, &result.LocationID, &result.Status, &result.CreatedAt)
	if errors.Is(err, pgx.ErrNoRows) {
		return PublicCase{}, ErrCaseNotFound
	}
	return result, err
}

// ListOpenPublicCases returns active cases without exposing their solutions.
func (store *PostgresStore) ListOpenPublicCases(ctx context.Context, serverID, cityID string) ([]PublicCase, error) {
	rows, err := store.db.Query(ctx,
		`SELECT c.case_id, c.city_id, c.case_type, c.victim_entity_id, c.location_entity_id, s.status, c.created_at
		 FROM cases c JOIN case_states s USING (server_id, case_id)
		 WHERE c.server_id = $1::uuid AND c.city_id = $2 AND s.status = 'open'
		 ORDER BY c.created_at DESC, c.case_id DESC`, serverID, cityID,
	)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	results := make([]PublicCase, 0)
	for rows.Next() {
		var result PublicCase
		if err := rows.Scan(&result.ID, &result.CityID, &result.CaseType, &result.VictimID, &result.LocationID, &result.Status, &result.CreatedAt); err != nil {
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
		`SELECT c.case_id, c.city_id, c.case_type, c.victim_entity_id, c.location_entity_id, s.status, c.created_at
		 FROM cases c JOIN case_states s USING (server_id, case_id)
		 WHERE c.server_id = $1::uuid AND c.case_id = $2`, serverID, caseID,
	).Scan(&result.ID, &result.CityID, &result.CaseType, &result.VictimID, &result.LocationID, &result.Status, &result.CreatedAt)
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
		item.Type = "record"
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
	activityRows, err := store.db.Query(ctx,
		`SELECT record_id, record_data FROM case_activity_records
		 WHERE server_id = $1::uuid AND case_id = $2 ORDER BY occurred_at, record_id`, serverID, caseID,
	)
	if err != nil {
		return PublicCase{}, nil, err
	}
	for activityRows.Next() {
		var id string
		var data []byte
		if err := activityRows.Scan(&id, &data); err != nil {
			activityRows.Close()
			return PublicCase{}, nil, err
		}
		var record records.Record
		if err := json.Unmarshal(data, &record); err != nil {
			activityRows.Close()
			return PublicCase{}, nil, fmt.Errorf("decode public activity record: %w", err)
		}
		kind, presentation := "gate", "GATE ACCESS LOG"
		content := record.Description
		if record.Source == "bank" {
			kind, presentation = "bank", "FINANCIAL RECORD"
			if record.Amount > 0 {
				content += fmt.Sprintf(" Amount: %d Credits.", record.Amount)
			}
		}
		if record.Origin != "" || record.Destination != "" {
			content += fmt.Sprintf(" Recorded passage: %s to %s.", record.Origin, record.Destination)
		}
		evidence = append(evidence, PublicEvidence{ID: id, Type: kind, Presentation: presentation, Content: content})
	}
	if err := activityRows.Err(); err != nil {
		activityRows.Close()
		return PublicCase{}, nil, err
	}
	activityRows.Close()
	mailRows, err := store.db.Query(ctx,
		`SELECT message_id, message_data FROM case_mail_messages
		 WHERE server_id = $1::uuid AND case_id = $2 ORDER BY sent_at, message_id`, serverID, caseID,
	)
	if err != nil {
		return PublicCase{}, nil, err
	}
	for mailRows.Next() {
		var id string
		var data []byte
		if err := mailRows.Scan(&id, &data); err != nil {
			mailRows.Close()
			return PublicCase{}, nil, err
		}
		var message mail.Message
		if err := json.Unmarshal(data, &message); err != nil {
			mailRows.Close()
			return PublicCase{}, nil, fmt.Errorf("decode public case correspondence: %w", err)
		}
		evidence = append(evidence, PublicEvidence{ID: id, Type: "email", Presentation: "EMAIL · " + message.Subject, Content: message.Body})
	}
	if err := mailRows.Err(); err != nil {
		mailRows.Close()
		return PublicCase{}, nil, err
	}
	mailRows.Close()
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
	progress := PlayerProgress{DiscoveredEvidenceIDs: make([]string, 0), WhiteboardPins: make([]string, 0), WhiteboardConnections: make([]WhiteboardConnection, 0)}
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
	if progress.WhiteboardPins == nil {
		progress.WhiteboardPins = make([]string, 0)
	}
	if progress.WhiteboardConnections == nil {
		progress.WhiteboardConnections = make([]WhiteboardConnection, 0)
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
			`SELECT COUNT(*) FROM (
				SELECT evidence_id AS item_id FROM case_evidence WHERE server_id = $1::uuid AND case_id = $2
				UNION ALL SELECT record_id FROM case_activity_records WHERE server_id = $1::uuid AND case_id = $2
				UNION ALL SELECT message_id FROM case_mail_messages WHERE server_id = $1::uuid AND case_id = $2
			) AS case_items WHERE item_id = ANY($3::text[])`,
			serverID, caseID, normalized.DiscoveredEvidenceIDs,
		).Scan(&count); err != nil {
			return err
		}
		if count != len(normalized.DiscoveredEvidenceIDs) {
			return ErrEvidenceNotFound
		}
	}
	discovered := make(map[string]bool, len(normalized.DiscoveredEvidenceIDs))
	for _, id := range normalized.DiscoveredEvidenceIDs {
		discovered[id] = true
	}
	for _, id := range normalized.WhiteboardPins {
		if !discovered[id] {
			return ErrEvidenceNotFound
		}
	}
	pinned := make(map[string]bool, len(normalized.WhiteboardPins))
	for _, id := range normalized.WhiteboardPins {
		pinned[id] = true
	}
	for _, connection := range normalized.WhiteboardConnections {
		if !pinned[connection.FromID] || !pinned[connection.ToID] {
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
	normalizeIDs := func(source []string) ([]string, error) {
		ids := append([]string(nil), source...)
		for _, id := range ids {
			if strings.TrimSpace(id) == "" {
				return nil, ErrInvalidProgress
			}
		}
		sort.Strings(ids)
		unique := ids[:0]
		for _, id := range ids {
			if len(unique) == 0 || unique[len(unique)-1] != id {
				unique = append(unique, id)
			}
		}
		return unique, nil
	}
	discovered, err := normalizeIDs(progress.DiscoveredEvidenceIDs)
	if err != nil {
		return PlayerProgress{}, err
	}
	pins, err := normalizeIDs(progress.WhiteboardPins)
	if err != nil {
		return PlayerProgress{}, err
	}
	connections := append([]WhiteboardConnection(nil), progress.WhiteboardConnections...)
	for index := range connections {
		connection := &connections[index]
		if strings.TrimSpace(connection.FromID) == "" || strings.TrimSpace(connection.ToID) == "" || connection.FromID == connection.ToID {
			return PlayerProgress{}, ErrInvalidProgress
		}
		if connection.ToID < connection.FromID {
			connection.FromID, connection.ToID = connection.ToID, connection.FromID
		}
	}
	sort.Slice(connections, func(i, j int) bool {
		if connections[i].FromID == connections[j].FromID {
			return connections[i].ToID < connections[j].ToID
		}
		return connections[i].FromID < connections[j].FromID
	})
	uniqueConnections := connections[:0]
	for _, connection := range connections {
		if len(uniqueConnections) == 0 || uniqueConnections[len(uniqueConnections)-1] != connection {
			uniqueConnections = append(uniqueConnections, connection)
		}
	}
	return PlayerProgress{DiscoveredEvidenceIDs: discovered, WhiteboardPins: pins, WhiteboardConnections: uniqueConnections}, nil
}

func (store *PostgresStore) RecordDiscoveredItems(ctx context.Context, serverID, accountID string, items []DiscoveredItem) error {
	valid := make([]DiscoveredItem, 0, len(items))
	for _, item := range items {
		if item.ID == "" || item.Title == "" || (item.Source != "email" && item.Source != "bank" && item.Source != "gate" && item.Source != "case" && item.Source != "person" && item.Source != "company" && item.Source != "note") {
			continue
		}
		valid = append(valid, item)
	}
	if len(valid) == 0 {
		return nil
	}
	data, err := json.Marshal(valid)
	if err != nil {
		return err
	}
	if _, err := store.db.Exec(ctx,
		`INSERT INTO player_discovered_items (server_id, account_id, item_id, item_source, item_data)
		 SELECT $1::uuid, $2::uuid, item->>'id', item->>'source', item
		 FROM jsonb_array_elements($3::jsonb) AS items(item)
		 ON CONFLICT (server_id, account_id, item_id) DO UPDATE
		 SET item_source = EXCLUDED.item_source, item_data = EXCLUDED.item_data`,
		serverID, accountID, string(data)); err != nil {
		return fmt.Errorf("record discovered items: %w", err)
	}
	return nil
}

func (store *PostgresStore) ListWhiteboards(ctx context.Context, serverID, accountID string) ([]Whiteboard, error) {
	rows, err := store.db.Query(ctx,
		`SELECT whiteboard_id, title, created_at FROM player_whiteboards
		 WHERE server_id = $1::uuid AND account_id = $2::uuid ORDER BY created_at, whiteboard_id`, serverID, accountID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	boards := make([]Whiteboard, 0)
	for rows.Next() {
		var board Whiteboard
		if err := rows.Scan(&board.ID, &board.Title, &board.CreatedAt); err != nil {
			return nil, err
		}
		board.Pins = []WhiteboardPin{}
		board.Connections = []WhiteboardConnection{}
		boards = append(boards, board)
	}
	return boards, rows.Err()
}

func (store *PostgresStore) CreateWhiteboard(ctx context.Context, serverID, accountID, boardID, title string) (Whiteboard, error) {
	var board Whiteboard
	err := store.db.QueryRow(ctx,
		`INSERT INTO player_whiteboards (server_id, account_id, whiteboard_id, title)
		 VALUES (
		   $1::uuid, $2::uuid, $3,
		   COALESCE(NULLIF($4, ''), 'INVESTIGATION ' || (
		     SELECT (COUNT(*) + 1)::text FROM player_whiteboards
		     WHERE server_id = $1::uuid AND account_id = $2::uuid
		   ))
		 ) RETURNING whiteboard_id, title, created_at`,
		serverID, accountID, boardID, title).Scan(&board.ID, &board.Title, &board.CreatedAt)
	if err != nil {
		return Whiteboard{}, err
	}
	board.Pins = []WhiteboardPin{}
	board.Connections = []WhiteboardConnection{}
	return board, nil
}

func (store *PostgresStore) RenameWhiteboard(ctx context.Context, serverID, accountID, boardID, title string) error {
	tag, err := store.db.Exec(ctx,
		`UPDATE player_whiteboards SET title = $4
		 WHERE server_id = $1::uuid AND account_id = $2::uuid AND whiteboard_id = $3`,
		serverID, accountID, boardID, title)
	if err != nil {
		return err
	}
	if tag.RowsAffected() == 0 {
		return ErrCaseNotFound
	}
	return nil
}

func (store *PostgresStore) DeleteWhiteboard(ctx context.Context, serverID, accountID, boardID string) error {
	_, err := store.db.Exec(ctx,
		`DELETE FROM player_whiteboards
		 WHERE server_id = $1::uuid AND account_id = $2::uuid AND whiteboard_id = $3`,
		serverID, accountID, boardID)
	// DELETE is idempotent so clients can remove stale boards left in local cache.
	return err
}

func (store *PostgresStore) GetWhiteboard(ctx context.Context, serverID, accountID, boardID string) (Whiteboard, error) {
	var board Whiteboard
	var pinsJSON, connectionsJSON []byte
	err := store.db.QueryRow(ctx,
		`SELECT board.title, board.created_at,
		        COALESCE((
		          SELECT jsonb_agg(item.item_data || jsonb_build_object('x', pin.position_x, 'y', pin.position_y)
		                           ORDER BY pin.pinned_at, pin.item_id)
		          FROM player_whiteboard_pins pin
		          JOIN player_discovered_items item USING (server_id, account_id, item_id)
		          WHERE pin.server_id = board.server_id AND pin.account_id = board.account_id AND pin.whiteboard_id = board.whiteboard_id
		        ), '[]'::jsonb),
		        COALESCE((
		          SELECT jsonb_agg(jsonb_build_object('fromId', edge.from_item_id, 'toId', edge.to_item_id)
		                           ORDER BY edge.from_item_id, edge.to_item_id)
		          FROM player_whiteboard_connections edge
		          WHERE edge.server_id = board.server_id AND edge.account_id = board.account_id AND edge.whiteboard_id = board.whiteboard_id
		        ), '[]'::jsonb)
		 FROM player_whiteboards board
		 WHERE board.server_id = $1::uuid AND board.account_id = $2::uuid AND board.whiteboard_id = $3`,
		serverID, accountID, boardID).Scan(&board.Title, &board.CreatedAt, &pinsJSON, &connectionsJSON)
	if errors.Is(err, pgx.ErrNoRows) {
		return Whiteboard{}, ErrCaseNotFound
	}
	if err != nil {
		return Whiteboard{}, err
	}
	board.ID = boardID
	if err := json.Unmarshal(pinsJSON, &board.Pins); err != nil {
		return Whiteboard{}, fmt.Errorf("decode whiteboard pins: %w", err)
	}
	if err := json.Unmarshal(connectionsJSON, &board.Connections); err != nil {
		return Whiteboard{}, fmt.Errorf("decode whiteboard connections: %w", err)
	}
	return board, nil
}

func (store *PostgresStore) PinDiscoveredItem(ctx context.Context, serverID, accountID, boardID, itemID string, x, y int) error {
	if x < 0 || y < 0 || x > 10000 || y > 10000 {
		return ErrInvalidProgress
	}
	itemIDs := []string{itemID}
	if strings.HasPrefix(itemID, "company:company:") {
		itemIDs = append(itemIDs, strings.TrimPrefix(itemID, "company:"))
	} else if strings.HasPrefix(itemID, "company:") {
		itemIDs = append(itemIDs, "company:"+itemID)
	}
	var boardExists bool
	var discoveredItemID string
	if err := store.db.QueryRow(ctx,
		`SELECT EXISTS (SELECT 1 FROM player_whiteboards WHERE server_id = $1::uuid AND account_id = $2::uuid AND whiteboard_id = $3),
		        COALESCE((SELECT item_id FROM player_discovered_items
		                  WHERE server_id = $1::uuid AND account_id = $2::uuid AND item_id = ANY($4::text[])
		                  ORDER BY CASE WHEN item_id = $5 THEN 0 ELSE 1 END LIMIT 1), '')`,
		serverID, accountID, boardID, itemIDs, itemID).Scan(&boardExists, &discoveredItemID); err != nil {
		return err
	}
	if !boardExists || discoveredItemID == "" {
		return ErrEvidenceNotFound
	}
	_, err := store.db.Exec(ctx,
		`INSERT INTO player_whiteboard_pins (server_id, account_id, whiteboard_id, item_id, position_x, position_y)
		 VALUES ($1::uuid, $2::uuid, $3, $4, $5, $6) ON CONFLICT DO NOTHING`, serverID, accountID, boardID, discoveredItemID, x, y)
	return err
}

func (store *PostgresStore) CreateWhiteboardNote(ctx context.Context, serverID, accountID, boardID string, x, y int) (WhiteboardPin, error) {
	if x < 0 || y < 0 || x > 10000 || y > 10000 {
		return WhiteboardPin{}, ErrInvalidProgress
	}
	tx, err := store.db.Begin(ctx)
	if err != nil {
		return WhiteboardPin{}, err
	}
	defer tx.Rollback(ctx)

	var boardExists bool
	if err := tx.QueryRow(ctx,
		`SELECT EXISTS (SELECT 1 FROM player_whiteboards WHERE server_id = $1::uuid AND account_id = $2::uuid AND whiteboard_id = $3)`,
		serverID, accountID, boardID).Scan(&boardExists); err != nil {
		return WhiteboardPin{}, err
	}
	if !boardExists {
		return WhiteboardPin{}, ErrCaseNotFound
	}

	var itemID string
	if err := tx.QueryRow(ctx,
		`WITH note_id AS (SELECT 'note:' || gen_random_uuid()::text AS id)
		 INSERT INTO player_discovered_items (server_id, account_id, item_id, item_source, item_data)
		 SELECT $1::uuid, $2::uuid, note_id.id, 'note',
		        jsonb_build_object('id', note_id.id, 'source', 'note', 'title', 'UNTITLED NOTE', 'content', '')
		 FROM note_id
		 RETURNING item_id`, serverID, accountID).Scan(&itemID); err != nil {
		return WhiteboardPin{}, err
	}
	if _, err := tx.Exec(ctx,
		`INSERT INTO player_whiteboard_pins (server_id, account_id, whiteboard_id, item_id, position_x, position_y)
		 VALUES ($1::uuid, $2::uuid, $3, $4, $5, $6)`, serverID, accountID, boardID, itemID, x, y); err != nil {
		return WhiteboardPin{}, err
	}
	if err := tx.Commit(ctx); err != nil {
		return WhiteboardPin{}, err
	}
	return WhiteboardPin{DiscoveredItem: DiscoveredItem{ID: itemID, Source: "note", Title: "UNTITLED NOTE"}, X: x, Y: y}, nil
}

func (store *PostgresStore) UpdateWhiteboardNote(ctx context.Context, serverID, accountID, boardID, itemID, title, content string) error {
	title = strings.TrimSpace(title)
	if title == "" {
		title = "UNTITLED NOTE"
	}
	if len([]rune(title)) > 80 || len([]rune(content)) > 5000 {
		return ErrInvalidWhiteboardNote
	}
	result, err := store.db.Exec(ctx,
		`UPDATE player_discovered_items item
		 SET item_data = item.item_data || jsonb_build_object('title', $5::text, 'content', $6::text)
		 WHERE item.server_id = $1::uuid AND item.account_id = $2::uuid AND item.item_id = $4
		   AND item.item_source = 'note'
		   AND EXISTS (
		       SELECT 1 FROM player_whiteboard_pins pin
		       WHERE pin.server_id = item.server_id AND pin.account_id = item.account_id
		         AND pin.whiteboard_id = $3 AND pin.item_id = item.item_id
		   )`, serverID, accountID, boardID, itemID, title, content)
	if err != nil {
		return err
	}
	if result.RowsAffected() != 1 {
		return ErrEvidenceNotFound
	}
	return nil
}

func (store *PostgresStore) MoveWhiteboardPin(ctx context.Context, serverID, accountID, boardID, itemID string, x, y int) error {
	if x < 0 || y < 0 || x > 10000 || y > 10000 {
		return ErrInvalidProgress
	}
	result, err := store.db.Exec(ctx,
		`UPDATE player_whiteboard_pins SET position_x = $5, position_y = $6
		 WHERE server_id = $1::uuid AND account_id = $2::uuid AND whiteboard_id = $3 AND item_id = $4`,
		serverID, accountID, boardID, itemID, x, y)
	if err != nil {
		return err
	}
	if result.RowsAffected() != 1 {
		return ErrEvidenceNotFound
	}
	return nil
}

func (store *PostgresStore) RemoveWhiteboardPin(ctx context.Context, serverID, accountID, boardID, itemID string) error {
	_, err := store.db.Exec(ctx,
		`DELETE FROM player_whiteboard_pins WHERE server_id = $1::uuid AND account_id = $2::uuid AND whiteboard_id = $3 AND item_id = $4`,
		serverID, accountID, boardID, itemID)
	// DELETE is idempotent: a stale client may still show a pin already removed on the server.
	return err
}

func (store *PostgresStore) ConnectWhiteboardPins(ctx context.Context, serverID, accountID, boardID, fromID, toID string) error {
	if fromID == toID || fromID == "" || toID == "" {
		return ErrInvalidProgress
	}
	var count int
	if err := store.db.QueryRow(ctx,
		`SELECT COUNT(*) FROM player_whiteboard_pins
		 WHERE server_id = $1::uuid AND account_id = $2::uuid AND whiteboard_id = $3 AND item_id = ANY($4::text[])`,
		serverID, accountID, boardID, []string{fromID, toID}).Scan(&count); err != nil {
		return err
	}
	if count != 2 {
		return ErrEvidenceNotFound
	}
	if toID < fromID {
		fromID, toID = toID, fromID
	}
	_, err := store.db.Exec(ctx,
		`INSERT INTO player_whiteboard_connections (server_id, account_id, whiteboard_id, from_item_id, to_item_id)
		 VALUES ($1::uuid, $2::uuid, $3, $4, $5) ON CONFLICT DO NOTHING`, serverID, accountID, boardID, fromID, toID)
	return err
}

func (store *PostgresStore) DisconnectWhiteboardPins(ctx context.Context, serverID, accountID, boardID, fromID, toID string) error {
	if toID < fromID {
		fromID, toID = toID, fromID
	}
	_, err := store.db.Exec(ctx,
		`DELETE FROM player_whiteboard_connections WHERE server_id = $1::uuid AND account_id = $2::uuid AND whiteboard_id = $3 AND from_item_id = $4 AND to_item_id = $5`,
		serverID, accountID, boardID, fromID, toID)
	return err
}

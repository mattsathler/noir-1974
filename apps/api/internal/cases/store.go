package cases

import (
	"context"
	cryptorand "crypto/rand"
	"crypto/sha256"
	"embed"
	"encoding/binary"
	"encoding/hex"
	"encoding/json"
	"errors"
	"fmt"
	"log"
	"math"
	"math/rand"
	"sort"
	"strconv"
	"strings"
	"time"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgconn"
	"github.com/jackc/pgx/v5/pgtype"
	"noir-1974/apps/api/internal/mail"
	"noir-1974/apps/api/internal/market"
	"noir-1974/apps/api/internal/records"
	"noir-1974/apps/api/internal/world"
)

//go:embed schema.sql interrogation/*.json residential_search/*.json
var caseSchema embed.FS

var (
	ErrCaseNotFound          = errors.New("case not found")
	ErrSessionNotFound       = errors.New("player session not found")
	ErrInvalidCaseState      = errors.New("invalid shared case state")
	ErrInvalidProgress       = errors.New("invalid player case progress")
	ErrInvalidLocalAccess    = errors.New("invalid local access result")
	ErrInvalidWhiteboardNote = errors.New("invalid whiteboard note")
	ErrEvidenceNotFound      = errors.New("case evidence not found")
	ErrPlayerMailNotFound    = errors.New("personal email not found")
	ErrPersistedConflict     = errors.New("persisted case does not match generated definition")
)

const residentialSearchCatalogVersion = 7

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
	X           int    `json:"x"`
	Y           int    `json:"y"`
	Designation string `json:"designation,omitempty"`
}

type AccusationResult struct {
	Accepted bool      `json:"-"`
	Status   string    `json:"status"`
	ReadyAt  time.Time `json:"readyAt,omitempty"`
	Message  string    `json:"message"`
}

type Whiteboard struct {
	ID          string                 `json:"id"`
	Title       string                 `json:"title"`
	CreatedAt   time.Time              `json:"createdAt"`
	LockedAt    *time.Time             `json:"lockedAt,omitempty"`
	Pins        []WhiteboardPin        `json:"pins"`
	Connections []WhiteboardConnection `json:"connections"`
}

type PlayerMailAttachment struct {
	DiscoveredItem
	Kind string `json:"kind"`
}

type PlayerMailMessage struct {
	ID          string                 `json:"id"`
	CaseID      string                 `json:"caseId,omitempty"`
	CrimeID     string                 `json:"crimeId,omitempty"`
	CrimeType   string                 `json:"crimeType,omitempty"`
	SentAt      time.Time              `json:"sentAt"`
	Sender      string                 `json:"sender"`
	Category    string                 `json:"category"`
	Subject     string                 `json:"subject"`
	Body        string                 `json:"body"`
	Attachments []PlayerMailAttachment `json:"attachments"`
}

type PlayerEmailQuestionRequest struct {
	RequestID  string    `json:"requestId"`
	NPCID      string    `json:"npcId"`
	Name       string    `json:"name"`
	Email      string    `json:"email"`
	QuestionID string    `json:"questionId"`
	Subject    string    `json:"subject"`
	Status     string    `json:"status"`
	ReadyAt    time.Time `json:"readyAt"`
}

var playerEmailQuestions = map[string]struct {
	prompt string
}{
	"routine":    {"How has your routine been lately?"},
	"associates": {"Who do you usually spend time with?"},
	"unusual":    {"Has anything unusual happened around here recently?"},
	"work":       {"How are things at work?"},
	"worries":    {"Is anything worrying you?"},
}

func billDispatchPostscript(seed string) string {
	bulletins := [...]string{
		"BILL BOUER, CITY POLICE: The Crow left his skull mark on another breached terminal. He calls the intrusion resistance; this office continues to treat it as theft. His capture remains an active priority.",
		"BILL BOUER, CITY POLICE: Investigators are tracing the skull mark linked to The Crow. His messages claim the stolen files expose corporate surveillance. That claim does not make an unauthorized breach lawful.",
		"BILL BOUER, CITY POLICE: A skull painted beside the latest breach has been attributed to The Crow. His broadcasts call this a fight against the system. The department is still working to bring him in.",
	}
	digest := sha256.Sum256([]byte("noir-m17-bill-dispatch/" + seed))
	index := binary.LittleEndian.Uint64(digest[:8]) % uint64(len(bulletins))
	return "\n\n" + bulletins[index]
}

type WarrantRequest struct {
	WarrantID string             `json:"warrantId"`
	Status    string             `json:"status"`
	ReadyAt   time.Time          `json:"readyAt,omitempty"`
	Message   *PlayerMailMessage `json:"message,omitempty"`
	BodyFound bool               `json:"bodyFound,omitempty"`
}

type EmailBreakerAccess struct {
	RequestID string    `json:"requestId"`
	NPCID     string    `json:"npcId"`
	Name      string    `json:"name"`
	Email     string    `json:"email"`
	From      string    `json:"from"`
	Through   string    `json:"through"`
	Status    string    `json:"status"`
	ReadyAt   time.Time `json:"readyAt"`
}

type BankruptAccess struct {
	NPCID         string    `json:"npcId"`
	Name          string    `json:"name"`
	AccountNumber string    `json:"accountNumber"`
	OpenedAt      time.Time `json:"openedAt"`
}

func (store *PostgresStore) GrantBankruptAccess(ctx context.Context, serverID, accountID string, employee world.CompanyEmployee) (BankruptAccess, error) {
	var result BankruptAccess
	err := store.db.QueryRow(ctx, `INSERT INTO player_bankrupt_access(server_id,account_id,target_npc_id,target_account_number,target_name)
	 VALUES($1::uuid,$2::uuid,$3,$4,$5) ON CONFLICT(server_id,account_id,target_npc_id) DO UPDATE SET target_account_number=EXCLUDED.target_account_number,target_name=EXCLUDED.target_name
	 RETURNING target_npc_id,target_name,target_account_number,opened_at`, serverID, accountID, employee.ID, employee.BankAccount, employee.Name).
		Scan(&result.NPCID, &result.Name, &result.AccountNumber, &result.OpenedAt)
	return result, err
}

func (store *PostgresStore) HasBankruptAccess(ctx context.Context, serverID, accountID, npcID string) (bool, error) {
	var allowed bool
	err := store.db.QueryRow(ctx, `SELECT EXISTS(SELECT 1 FROM player_bankrupt_access WHERE server_id=$1::uuid AND account_id=$2::uuid AND target_npc_id=$3)`, serverID, accountID, npcID).Scan(&allowed)
	return allowed, err
}

func (store *PostgresStore) RequestEmailBreaker(ctx context.Context, serverID, accountID string, employee world.CompanyEmployee, from, through, requestID string) (EmailBreakerAccess, error) {
	_, err := store.db.Exec(ctx, `INSERT INTO player_email_breaker_access(server_id,account_id,request_id,target_npc_id,target_name,target_email,from_date,through_date,ready_at)
	 VALUES($1::uuid,$2::uuid,$3,$4,$5,$6,$7::date,$8::date,now()) ON CONFLICT(server_id,account_id,request_id) DO NOTHING`, serverID, accountID, requestID, employee.ID, employee.Name, employee.Email, from, through)
	if err != nil {
		return EmailBreakerAccess{}, err
	}
	return store.EmailBreakerRequest(ctx, serverID, accountID, requestID)
}

func (store *PostgresStore) EmailBreakerRequest(ctx context.Context, serverID, accountID, requestID string) (EmailBreakerAccess, error) {
	var result EmailBreakerAccess
	err := store.db.QueryRow(ctx, `SELECT request_id,target_npc_id,target_name,target_email,from_date::text,through_date::text,
	 CASE WHEN ready_at<=now() THEN 'ready' ELSE 'pending' END,ready_at
	 FROM player_email_breaker_access WHERE server_id=$1::uuid AND account_id=$2::uuid AND request_id=$3`, serverID, accountID, requestID).
		Scan(&result.RequestID, &result.NPCID, &result.Name, &result.Email, &result.From, &result.Through, &result.Status, &result.ReadyAt)
	return result, err
}

func (store *PostgresStore) ListEmailBreakerAccess(ctx context.Context, serverID, accountID string) ([]EmailBreakerAccess, error) {
	rows, err := store.db.Query(ctx, `SELECT request_id,target_npc_id,target_name,target_email,from_date::text,through_date::text,
	 CASE WHEN ready_at<=now() THEN 'ready' ELSE 'pending' END,ready_at
	 FROM player_email_breaker_access WHERE server_id=$1::uuid AND account_id=$2::uuid ORDER BY created_at DESC`, serverID, accountID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	results := make([]EmailBreakerAccess, 0)
	for rows.Next() {
		var item EmailBreakerAccess
		if err := rows.Scan(&item.RequestID, &item.NPCID, &item.Name, &item.Email, &item.From, &item.Through, &item.Status, &item.ReadyAt); err != nil {
			return nil, err
		}
		results = append(results, item)
	}
	return results, rows.Err()
}

func (store *PostgresStore) HasEmailBreakerAccess(ctx context.Context, serverID, accountID, npcID string) (bool, error) {
	var allowed bool
	err := store.db.QueryRow(ctx, `SELECT EXISTS(SELECT 1 FROM player_email_breaker_access WHERE server_id=$1::uuid AND account_id=$2::uuid AND target_npc_id=$3 AND ready_at<=now())`, serverID, accountID, npcID).Scan(&allowed)
	return allowed, err
}

type persistedCaseIdentity struct {
	Seed             string
	CityID           string
	GeneratorVersion int
	CaseType         string
	ArchetypeID      string
	CompanyID        string
	Motive           string
	VictimID         string
	KillerID         string
	LocationID       string
}

func (identity persistedCaseIdentity) matches(definition Definition) bool {
	return identity.Seed == definition.Seed && identity.CityID == definition.CityID && identity.GeneratorVersion == definition.GeneratorVersion && identity.CaseType == definition.CaseType && identity.ArchetypeID == definition.ArchetypeID && identity.CompanyID == definition.CompanyID && identity.Motive == definition.Motive && identity.VictimID == definition.VictimID && identity.KillerID == definition.KillerID && identity.LocationID == definition.LocationID
}

type asyncRequest struct {
	RequestID   string
	AccountID   string
	RequestType string
	Payload     []byte
	Status      string
	AvailableAt time.Time
}

type residentialSearchPayload struct {
	CityID            string `json:"cityId"`
	StationLocationID string `json:"stationLocationId"`
	TargetNPCID       string `json:"targetNpcId"`
}

type accusationPayload struct {
	CityID     string   `json:"cityId"`
	BoardID    string   `json:"boardId"`
	CaseID     string   `json:"caseId,omitempty"`
	CrimeIDs   []string `json:"crimeIds,omitempty"`
	AccusedID  string   `json:"accusedId,omitempty"`
	AccusedIDs []string `json:"accusedIds,omitempty"`
	VictimID   string   `json:"victimId,omitempty"`
	ProofIDs   []string `json:"proofIds"`
}

func (request accusationPayload) accusedNPCIDs() []string {
	if len(request.AccusedIDs) > 0 {
		return request.AccusedIDs
	}
	if request.AccusedID != "" {
		return []string{request.AccusedID}
	}
	return nil
}

type LocalAccessAttempt struct {
	AttemptID int64     `json:"attemptId"`
	Status    string    `json:"status"`
	Attempted time.Time `json:"attemptedAt"`
}

// PublicCase contains only facts safe for the newspaper headline.
type PublicCase struct {
	ID               string                `json:"id"`
	CityID           string                `json:"cityId"`
	CaseType         string                `json:"caseType"`
	ArchetypeID      string                `json:"-"`
	VictimID         string                `json:"victimId,omitempty"`
	VictimName       string                `json:"victimName,omitempty"`
	VictimVisual     *world.VisualIdentity `json:"victimVisual,omitempty"`
	OrganizationName string                `json:"organizationName,omitempty"`
	CrimeAt          time.Time             `json:"crimeAt,omitempty"`
	LocationID       string                `json:"-"`
	Status           string                `json:"status"`
	CreatedAt        time.Time             `json:"createdAt"`
}

// PublicBlackwireExposure contains only a company's name and the public
// consequences of a severe data leak; it never identifies the seller or source.
type PublicBlackwireExposure struct {
	CompanyID  string
	Category   string
	OccurredAt time.Time
}

func (store *PostgresStore) ListPublicBlackwireExposures(ctx context.Context, serverID, cityID string) ([]PublicBlackwireExposure, error) {
	rows, err := store.db.Query(ctx, `SELECT company_id,category,occurred_at FROM (
			SELECT DISTINCT event.company_id,sale.category,event.occurred_at
			FROM company_market_events event
			JOIN player_blackwire_sales sale ON sale.server_id=event.server_id
				AND event.event_id='blackwire:' || sale.account_id::text || ':' || sale.item_id
			WHERE event.server_id=$1::uuid AND event.event_type='blackwire_leak'
				AND event.company_id LIKE '%' || $2 || '%'
				AND event.recorded_at >= now() - interval '24 hours' AND event.recorded_at <= now()
				AND sale.category IN ('FINANCIAL / CORPORATE EVIDENCE','EXECUTIVE FINANCIAL EVIDENCE','CORPORATE LEADERSHIP','POLITICAL / EXECUTIVE EVIDENCE','POLITICAL INFORMATION')
			UNION ALL
			SELECT ''::text,sale.category,sale.sold_at
			FROM player_blackwire_sales sale
			WHERE sale.server_id=$1::uuid AND sale.company_id IS NULL
				AND sale.category IN ('POLITICAL / EXECUTIVE EVIDENCE','POLITICAL INFORMATION')
				AND sale.sold_at >= now() - interval '24 hours' AND sale.sold_at <= now()
		) exposure ORDER BY occurred_at DESC LIMIT 6`, serverID, cityID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	var exposures []PublicBlackwireExposure
	for rows.Next() {
		var exposure PublicBlackwireExposure
		if err := rows.Scan(&exposure.CompanyID, &exposure.Category, &exposure.OccurredAt); err != nil {
			return nil, err
		}
		exposures = append(exposures, exposure)
	}
	return exposures, rows.Err()
}

type MissingPerson struct {
	NPCID     string                `json:"npcId"`
	Name      string                `json:"name"`
	Visual    *world.VisualIdentity `json:"visual,omitempty"`
	MissingAt time.Time             `json:"missingAt"`
}

type CrimeReportDigest struct {
	Cases   []PublicCase    `json:"cases"`
	Missing []MissingPerson `json:"missing"`
}

type PublicBodyDiscovery struct {
	CaseID       string    `json:"caseId"`
	Status       string    `json:"status"`
	VictimName   string    `json:"victimName"`
	LocationName string    `json:"locationName"`
	DistrictName string    `json:"districtName"`
	FoundAt      time.Time `json:"foundAt"`
}

// PersonSearchScenes returns public scene addresses for case victims across one country.
func (store *PostgresStore) PersonSearchScenes(ctx context.Context, serverID, countryID string, npcIDs []string) (map[string]world.PersonSearchScene, error) {
	scenes := make(map[string]world.PersonSearchScene)
	if len(npcIDs) == 0 {
		return scenes, nil
	}
	cities := make([]world.WorldCity, 0, 4)
	cityIDs := make([]string, 0, 5)
	for _, city := range world.Cities() {
		if city.CountryID == countryID {
			cities = append(cities, city)
			cityIDs = append(cityIDs, city.ID)
		}
	}
	if len(cities) == 0 {
		return scenes, nil
	}
	if countryID == "fr" {
		cityIDs = append(cityIDs, "lyon")
	}
	rows, err := store.db.Query(ctx, `
		SELECT DISTINCT ON (c.victim_entity_id) c.victim_entity_id, c.location_entity_id, c.city_id
		FROM cases c JOIN case_states s USING (server_id, case_id)
		WHERE c.server_id = $1::uuid AND c.city_id = ANY($2::text[])
		  AND s.status IN ('open', 'solved') AND c.victim_entity_id = ANY($3::text[])
		ORDER BY c.victim_entity_id, c.created_at DESC, c.case_id DESC`, serverID, cityIDs, npcIDs)
	if err != nil {
		return nil, err
	}
	type victimScene struct{ locationID, cityID string }
	locationsByVictim := make(map[string]victimScene)
	for rows.Next() {
		var victimID, locationID, cityID string
		if err := rows.Scan(&victimID, &locationID, &cityID); err != nil {
			rows.Close()
			return nil, err
		}
		if cityID == "lyon" {
			cityID = "fr/lyon"
		}
		locationsByVictim[victimID] = victimScene{locationID: locationID, cityID: cityID}
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
	type sceneCityIndex struct {
		city     world.WorldCity
		entities map[string]world.Entity
		regions  map[string]string
	}
	byCity := make(map[string]sceneCityIndex)
	for _, scene := range locationsByVictim {
		if _, exists := byCity[scene.cityID]; exists {
			continue
		}
		city, exists := world.CityByID(scene.cityID)
		if !exists || city.CountryID != countryID {
			continue
		}
		cityVersion := version
		if city.ID != "fr/lyon" {
			err := store.db.QueryRow(ctx, `SELECT generator_version FROM world_city_generators WHERE server_id = $1::uuid AND city_id = $2`, serverID, city.ID).Scan(&cityVersion)
			if errors.Is(err, pgx.ErrNoRows) {
				if version < world.LatestGeneratorVersion {
					cityVersion = world.LatestGeneratorVersion
				}
			} else if err != nil {
				return nil, err
			}
		}
		generated, err := world.GenerateCity(seed, cityVersion, city.ID)
		if err != nil {
			return nil, err
		}
		index := sceneCityIndex{city: city, entities: make(map[string]world.Entity, len(generated.Entities)), regions: make(map[string]string, len(generated.Regions))}
		for _, entity := range generated.Entities {
			index.entities[entity.ID] = entity
		}
		for _, region := range generated.Regions {
			index.regions[region.ID] = region.Name
		}
		byCity[city.ID] = index
	}
	for victimID, scene := range locationsByVictim {
		index, exists := byCity[scene.cityID]
		if !exists {
			continue
		}
		entity, exists := index.entities[scene.locationID]
		if !exists {
			continue
		}
		address := strings.Trim(strings.Join([]string{entity.Name, index.regions[entity.RegionID], index.city.Name, index.city.Country}, ", "), ", ")
		scenes[victimID] = world.PersonSearchScene{Name: entity.Name, Address: address}
	}
	return scenes, nil
}

type PublicEvidence struct {
	ID           string `json:"id"`
	CrimeID      string `json:"crimeId,omitempty"`
	CrimeType    string `json:"crimeType,omitempty"`
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

type interrogationProfiles struct {
	Personalities map[string]interrogationDialogue `json:"personalities"`
	Occupations   map[string]interrogationDialogue `json:"occupations"`
	Routines      map[string]interrogationDialogue `json:"routines"`
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

func politicalInterrogationAnswers(seed, npcID string) InterrogationAnswers {
	statements := InterrogationAnswers{
		"unusual": {
			"The papers say Bill Bouer has a new lead on The Crow. At the tram stop, no one will say whether the skull on the station door was painted or printed.",
			"I heard The Crow says the skull is a warning, not a signature. The police radio called it evidence. I do not know which story to trust.",
			"The ministry says Ghostkey is a criminal tool. A friend says it showed them that a company had changed their file, and officials say The Crow uses it to steal records.",
			"Bill Bouer was on the radio again, promising The Crow will be caught. The announcer skipped the part about what The Crow says he stole.",
			"People say The Crow leaves a skull where a system has been opened. I think the frightening part is how quickly the police arrive.",
		},
		"worries": {
			"I keep hearing the state's warning about Ghostkey. The Crow calls it a key to locked files. It sounds less like a warning when they say everyone who uses it will be watched.",
			"Bouer says The Crow's thefts endanger residents. The Crow says the companies endanger us first. I worry both sides can make that sound true.",
			"They painted over The Crow's skull on the public terminal by morning. By evening someone had drawn it again. I worry the next visit will be from the police.",
			"I worry that if The Crow is caught, the ministry will say the quiet proves it was right all along.",
			"A neighbor says The Crow's skull means someone opened a locked file. The paper says it means a criminal was here. Neither explanation makes me feel safer.",
		},
	}
	return InterrogationAnswers{
		"unusual": selectInterrogationLines(statements["unusual"], 1, "politics/"+seed+"/"+npcID+"/unusual"),
		"worries": selectInterrogationLines(statements["worries"], 1, "politics/"+seed+"/"+npcID+"/worries"),
	}
}

// InterrogationAnswersForNPC returns authored reply text for open cases connected to this NPC.
// Internal case roles and culprit IDs stay server-side.
func (store *PostgresStore) InterrogationAnswersForNPC(ctx context.Context, serverID, npcID string) (InterrogationAnswers, error) {
	data, err := caseSchema.ReadFile("interrogation/default.json")
	if err != nil {
		return nil, err
	}
	var defaultDialogue interrogationDialogue
	if err := json.Unmarshal(data, &defaultDialogue); err != nil {
		return nil, err
	}
	var profiles interrogationProfiles
	if err := json.Unmarshal(data, &profiles); err != nil {
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
	var generationKey, occupation, employerKey string
	if err := store.db.QueryRow(ctx, `SELECT w.world_seed,w.generator_version,e.generation_key,COALESCE(j.role,''),COALESCE(company.generation_key,'')
		FROM worlds w JOIN world_entities e ON e.server_id=w.server_id AND e.entity_id=$2 AND e.entity_type='npc'
		LEFT JOIN LATERAL (SELECT role,employer_company_id FROM employment_relationships WHERE server_id=e.server_id AND employee_npc_id=e.entity_id ORDER BY relationship_id LIMIT 1) j ON true
		LEFT JOIN world_entities company ON company.server_id=e.server_id AND company.entity_id=j.employer_company_id
		WHERE w.server_id=$1::uuid`, serverID, npcID).Scan(&seed, &version, &generationKey, &occupation, &employerKey); err != nil {
		return nil, err
	}
	npc, err := world.GenerateNPCFromKey(seed, version, generationKey)
	if err != nil {
		return nil, err
	}
	employer := ""
	if employerKey != "" {
		company, err := world.GenerateCompanyFromKey(seed, version, employerKey)
		if err == nil {
			employer = company.Name
		}
	}
	routine := "independent"
	switch occupation {
	case "driver", "messenger":
		routine = "route"
	case "manager":
		routine = "supervisory"
	case "accountant", "clerk", "receptionist", "secretary", "typist":
		routine = "office"
	}
	profileAnswers := []InterrogationAnswers{profiles.Personalities[npc.Personality].answers(), profiles.Occupations[occupation].answers(), profiles.Routines[routine].answers()}
	rows, err := store.db.Query(ctx, `
		SELECT c.case_id,c.archetype_id,
		       victim.generation_key, killer.generation_key,
	       CASE WHEN c.killer_entity_id = $2 THEN 'perpetrator'
	            WHEN current_victim.entity_id = $2 THEN 'acquaintance'
	            WHEN EXISTS (
	              SELECT 1 FROM personal_relationships direct_relation
	              WHERE direct_relation.server_id = c.server_id
	                AND ((direct_relation.person_a_npc_id = $2 AND direct_relation.person_b_npc_id IN (current_victim.entity_id,c.killer_entity_id))
	                  OR (direct_relation.person_b_npc_id = $2 AND direct_relation.person_a_npc_id IN (current_victim.entity_id,c.killer_entity_id)))
	            ) THEN 'acquaintance'
	            WHEN EXISTS (
	              SELECT 1 FROM employment_relationships principal_job
	              JOIN employment_relationships caller_job
	                ON caller_job.server_id = principal_job.server_id
	               AND caller_job.employer_company_id = principal_job.employer_company_id
	              WHERE principal_job.server_id = c.server_id
	                AND principal_job.employee_npc_id IN (current_victim.entity_id,c.killer_entity_id)
	                AND caller_job.employee_npc_id = $2
	            ) THEN 'colleague' ELSE 'acquaintance' END
		FROM cases c
	JOIN case_states s USING (server_id, case_id)
	JOIN LATERAL (
	  SELECT participant.npc_entity_id AS entity_id
	  FROM case_crimes crime JOIN case_crime_participants participant USING(server_id,crime_id)
	  WHERE crime.server_id=c.server_id AND crime.case_id=c.case_id AND participant.participant_role='victim' AND crime.occurred_at<=now()
	  ORDER BY crime.crime_sequence DESC LIMIT 1
	) current_victim ON true
	JOIN world_entities victim ON victim.server_id = c.server_id AND victim.entity_id = current_victim.entity_id
	JOIN world_entities killer ON killer.server_id = c.server_id AND killer.entity_id = c.killer_entity_id
	WHERE c.server_id = $1::uuid AND s.status = 'open'
	  AND (c.killer_entity_id = $2 OR current_victim.entity_id = $2 OR EXISTS (
	    SELECT 1 FROM employment_relationships victim_job
	    JOIN employment_relationships caller_job
	      ON caller_job.server_id = victim_job.server_id
	     AND caller_job.employer_company_id = victim_job.employer_company_id
	    WHERE victim_job.server_id = c.server_id
	      AND victim_job.employee_npc_id IN (current_victim.entity_id,c.killer_entity_id)
	      AND caller_job.employee_npc_id = $2
	  ) OR EXISTS (
	    SELECT 1 FROM personal_relationships relation
	    WHERE relation.server_id = c.server_id
	      AND ((relation.person_a_npc_id = $2 AND relation.person_b_npc_id IN (current_victim.entity_id,c.killer_entity_id))
	        OR (relation.person_b_npc_id = $2 AND relation.person_a_npc_id IN (current_victim.entity_id,c.killer_entity_id)))
	  ))
	ORDER BY CASE
	           WHEN c.killer_entity_id = $2 THEN 0
	           WHEN current_victim.entity_id = $2 THEN 1
	           WHEN EXISTS (
	             SELECT 1 FROM personal_relationships relation
	             WHERE relation.server_id = c.server_id
	               AND ((relation.person_a_npc_id = $2 AND relation.person_b_npc_id IN (current_victim.entity_id,c.killer_entity_id))
	                 OR (relation.person_b_npc_id = $2 AND relation.person_a_npc_id IN (current_victim.entity_id,c.killer_entity_id)))
	           ) THEN 2
	           ELSE 3
	         END,
	         c.created_at DESC, c.case_id
		LIMIT 1`, serverID, npcID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	contextFound := false
	for rows.Next() {
		var caseID, archetype, victimKey, killerKey, role string
		if err := rows.Scan(&caseID, &archetype, &victimKey, &killerKey, &role); err != nil {
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
		mergeInterrogationAnswers(answers, caseFile.Shared.answers(), map[string]string{"victim": victim.Name, "appearance": killer.Appearance}, 1, caseID+"/shared/"+npcID)
		mergeInterrogationAnswers(answers, branch.answers(), map[string]string{"victim": victim.Name, "suspect": killer.Name, "appearance": killer.Appearance}, 2, caseID+"/role/"+npcID)
	}
	if err := rows.Err(); err != nil {
		return nil, err
	}
	if !contextFound {
		profileTokens := map[string]string{"npc": npc.Name, "occupation": occupation, "employer": employer, "personality": npc.Personality, "routine": routine}
		for index, profile := range profileAnswers {
			mergeInterrogationAnswers(answers, profile, profileTokens, 1, npcID+"/profile/"+fmt.Sprint(index))
		}
	}
	mergeInterrogationAnswers(answers, politicalInterrogationAnswers(seed, npcID), nil, 1, "politics/"+seed+"/"+npcID)
	return answers, nil
}

func mergeInterrogationAnswers(target, source InterrogationAnswers, names map[string]string, limit int, selectionKey string) {
	for question, lines := range source {
		for _, line := range selectInterrogationLines(lines, limit, selectionKey+"/"+question) {
			for token, name := range names {
				line = strings.ReplaceAll(line, "{"+token+"}", name)
			}
			target[question] = append(target[question], line)
		}
	}
}

func selectInterrogationLines(lines []string, limit int, selectionKey string) []string {
	if limit <= 0 || len(lines) <= limit {
		return lines
	}
	hash := sha256.Sum256([]byte(selectionKey))
	start := int(binary.BigEndian.Uint64(hash[:8]) % uint64(len(lines)))
	selected := make([]string, 0, limit)
	for offset := 0; offset < len(lines) && len(selected) < limit; offset++ {
		line := lines[(start+offset)%len(lines)]
		duplicate := false
		for _, existing := range selected {
			if line == existing {
				duplicate = true
				break
			}
		}
		if !duplicate {
			selected = append(selected, line)
		}
	}
	return selected
}

func NewPostgresStore(db caseDatabase) *PostgresStore {
	return &PostgresStore{db: db}
}

func (store *PostgresStore) worldSeedAndVersion(ctx context.Context, serverID string) (string, int, error) {
	var seed string
	var version int
	err := store.db.QueryRow(ctx,
		`SELECT world_seed, generator_version FROM worlds WHERE server_id = $1::uuid`,
		serverID,
	).Scan(&seed, &version)
	if errors.Is(err, pgx.ErrNoRows) {
		return "", 0, world.ErrEntityNotFound
	}
	return seed, version, err
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
	rows, err := store.db.Query(ctx, `SELECT participant.npc_entity_id FROM case_participants participant JOIN cases c USING(server_id,case_id)
		WHERE c.server_id=$1::uuid AND c.city_id=$2 AND c.case_seed<>$3`, serverID, cityID, currentSeed)
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
	for index, statement := range splitSQLStatements(string(contents)) {
		if strings.TrimSpace(statement) == "" {
			continue
		}
		if _, err := store.db.Exec(ctx, statement); err != nil {
			summary := strings.Join(strings.Fields(statement), " ")
			if len(summary) > 160 {
				summary = summary[:160] + "…"
			}
			return fmt.Errorf("apply case schema statement %d (%s): %w", index+1, summary, err)
		}
	}
	return nil
}

// splitSQLStatements keeps semicolons inside quoted strings, comments, and
// PostgreSQL dollar-quoted function bodies intact while applying embedded schema.
func splitSQLStatements(sql string) []string {
	statements := make([]string, 0)
	start := 0
	var dollarTag string
	singleQuoted, doubleQuoted, lineComment, blockComment := false, false, false, false
	for i := 0; i < len(sql); i++ {
		if lineComment {
			if sql[i] == '\n' {
				lineComment = false
			}
			continue
		}
		if blockComment {
			if i+1 < len(sql) && sql[i] == '*' && sql[i+1] == '/' {
				blockComment = false
				i++
			}
			continue
		}
		if dollarTag != "" {
			if strings.HasPrefix(sql[i:], dollarTag) {
				i += len(dollarTag) - 1
				dollarTag = ""
			}
			continue
		}
		if singleQuoted {
			if sql[i] == '\'' {
				if i+1 < len(sql) && sql[i+1] == '\'' {
					i++
					continue
				}
				singleQuoted = false
			}
			continue
		}
		if doubleQuoted {
			if sql[i] == '"' {
				if i+1 < len(sql) && sql[i+1] == '"' {
					i++
					continue
				}
				doubleQuoted = false
			}
			continue
		}
		if i+1 < len(sql) && sql[i] == '-' && sql[i+1] == '-' {
			lineComment = true
			i++
			continue
		}
		if i+1 < len(sql) && sql[i] == '/' && sql[i+1] == '*' {
			blockComment = true
			i++
			continue
		}
		switch sql[i] {
		case '\'':
			singleQuoted = true
		case '"':
			doubleQuoted = true
		case '$':
			if end := strings.IndexByte(sql[i+1:], '$'); end >= 0 {
				tagBody := sql[i+1 : i+1+end]
				valid := true
				for j, char := range tagBody {
					if !(char == '_' || char >= 'a' && char <= 'z' || char >= 'A' && char <= 'Z' || j > 0 && char >= '0' && char <= '9') {
						valid = false
						break
					}
				}
				if valid {
					dollarTag = sql[i : i+end+2]
					i += end + 1
				}
			}
		case ';':
			if statement := strings.TrimSpace(sql[start:i]); statement != "" {
				statements = append(statements, statement)
			}
			start = i + 1
		}
	}
	if statement := strings.TrimSpace(sql[start:]); statement != "" {
		statements = append(statements, statement)
	}
	return statements
}

const DefaultAsyncRequestDelay = 3 * time.Minute
const DefaultAccusationReviewDelay = 30 * time.Second

// RequestResidentialSearch queues a police search. Its inventory is delivered after review.
// NPC status is deliberately not checked here: a deceased person's residence remains searchable.
func (store *PostgresStore) RequestResidentialSearch(ctx context.Context, serverID, accountID, cityID, stationLocationID, npcID string, _ time.Time) (WarrantRequest, error) {
	cityKey := cityID
	if cityID == "fr/lyon" {
		cityKey = "lyon"
	}
	if stationLocationID != "service:"+cityKey+"/police-station" || npcID == "" {
		return WarrantRequest{}, ErrCaseNotFound
	}
	var targetExists bool
	if err := store.db.QueryRow(ctx, `SELECT EXISTS (
		SELECT 1 FROM world_entities e JOIN world_regions r USING (server_id, region_id)
		WHERE e.server_id=$1::uuid AND e.entity_id=$2 AND e.entity_type='npc'
		AND (r.region_id=$3 OR r.region_id LIKE $3 || '/%')
	)`, serverID, npcID, strings.SplitN(cityID, "/", 2)[0]).Scan(&targetExists); err != nil {
		return WarrantRequest{}, fmt.Errorf("validate residential search target: %w", err)
	}
	if !targetExists {
		return WarrantRequest{}, ErrCaseNotFound
	}
	identity := sha256.Sum256([]byte(fmt.Sprintf("%s|%s|%s", accountID, cityID, npcID)))
	warrantID := "warrant:" + hex.EncodeToString(identity[:12])
	payload, err := json.Marshal(residentialSearchPayload{CityID: cityID, StationLocationID: stationLocationID, TargetNPCID: npcID})
	if err != nil {
		return WarrantRequest{}, err
	}
	tx, err := store.db.Begin(ctx)
	if err != nil {
		return WarrantRequest{}, fmt.Errorf("begin residential search request: %w", err)
	}
	defer func() { _ = tx.Rollback(context.Background()) }()
	if _, err := tx.Exec(ctx, `INSERT INTO player_search_warrants(server_id,account_id,warrant_id,city_id,station_location_id,target_npc_id,status,catalog_version,ready_at)
		VALUES($1::uuid,$2::uuid,$3,$4,$5,$6,'pending',3,now()+make_interval(secs=>$7::double precision))
		ON CONFLICT(server_id,account_id,warrant_id) DO NOTHING`, serverID, accountID, warrantID, cityID, stationLocationID, npcID, DefaultAsyncRequestDelay.Seconds()); err != nil {
		return WarrantRequest{}, fmt.Errorf("record residential search request: %w", err)
	}
	var status string
	var readyAt time.Time
	if err := tx.QueryRow(ctx, `SELECT status,ready_at FROM player_search_warrants WHERE server_id=$1::uuid AND account_id=$2::uuid AND warrant_id=$3`, serverID, accountID, warrantID).Scan(&status, &readyAt); err != nil {
		return WarrantRequest{}, err
	}
	if status != "pending" {
		if err := tx.Commit(ctx); err != nil {
			return WarrantRequest{}, err
		}
		return WarrantRequest{WarrantID: warrantID, Status: status, ReadyAt: readyAt}, nil
	}
	job, err := enqueueAsyncRequest(ctx, tx, serverID, accountID, "residential_search", warrantID, payload, DefaultAsyncRequestDelay)
	if err != nil {
		return WarrantRequest{}, err
	}
	if _, err := tx.Exec(ctx, `UPDATE player_search_warrants SET ready_at=$4 WHERE server_id=$1::uuid AND account_id=$2::uuid AND warrant_id=$3`, serverID, accountID, warrantID, job.AvailableAt); err != nil {
		return WarrantRequest{}, err
	}
	if err := tx.Commit(ctx); err != nil {
		return WarrantRequest{}, err
	}
	return WarrantRequest{WarrantID: warrantID, Status: job.Status, ReadyAt: job.AvailableAt}, nil
}

func enqueueAsyncRequest(ctx context.Context, tx pgx.Tx, serverID, accountID, requestType, requestKey string, payload []byte, delay time.Duration) (asyncRequest, error) {
	requestID := asyncRequestID(serverID, accountID, requestType, requestKey)
	_, err := tx.Exec(ctx, `INSERT INTO player_async_requests(server_id,account_id,request_id,request_type,request_key,payload,available_at)
		VALUES($1::uuid,$2::uuid,$3,$4,$5,$6::jsonb,now()+make_interval(secs=>$7::double precision))
		ON CONFLICT(server_id,account_id,request_type,request_key) DO NOTHING`, serverID, accountID, requestID, requestType, requestKey, string(payload), delay.Seconds())
	if err != nil {
		return asyncRequest{}, fmt.Errorf("queue asynchronous %s request: %w", requestType, err)
	}
	var request asyncRequest
	err = tx.QueryRow(ctx, `SELECT request_id,request_type,payload,status,available_at FROM player_async_requests
		WHERE server_id=$1::uuid AND account_id=$2::uuid AND request_type=$3 AND request_key=$4`, serverID, accountID, requestType, requestKey).
		Scan(&request.RequestID, &request.RequestType, &request.Payload, &request.Status, &request.AvailableAt)
	return request, err
}

func asyncRequestID(serverID, accountID, requestType, requestKey string) string {
	identity := sha256.Sum256([]byte(fmt.Sprintf("%s|%s|%s|%s", serverID, accountID, requestType, requestKey)))
	return "async:" + hex.EncodeToString(identity[:16])
}

// completeResidentialSearch generates and mails the deterministic inventory for a queued search.
func (store *PostgresStore) completeResidentialSearch(ctx context.Context, serverID, accountID, cityID, stationLocationID, npcID string, now time.Time) (WarrantRequest, error) {
	var generationKey string
	if err := store.db.QueryRow(ctx, `SELECT generation_key FROM world_entities WHERE server_id=$1::uuid AND entity_id=$2`, serverID, npcID).Scan(&generationKey); err != nil {
		return WarrantRequest{}, fmt.Errorf("load residential subject: %w", err)
	}
	var priorID, priorStatus string
	var priorCatalogVersion int
	var priorMessage PlayerMailMessage
	var priorAttachments []byte
	err := store.db.QueryRow(ctx, `SELECT w.warrant_id,w.status,w.catalog_version,m.message_id,m.sent_at,m.subject,m.body,m.attachments
		FROM player_search_warrants w JOIN player_inbox_messages m USING (server_id,account_id,warrant_id)
		WHERE w.server_id=$1::uuid AND w.account_id=$2::uuid AND w.city_id=$3 AND w.target_npc_id=$4
		ORDER BY w.created_at LIMIT 1`, serverID, accountID, cityID, npcID).Scan(&priorID, &priorStatus, &priorCatalogVersion, &priorMessage.ID, &priorMessage.SentAt, &priorMessage.Subject, &priorMessage.Body, &priorAttachments)
	if err == nil {
		if priorCatalogVersion >= residentialSearchCatalogVersion {
			if err := json.Unmarshal(priorAttachments, &priorMessage.Attachments); err != nil {
				return WarrantRequest{}, err
			}
			return WarrantRequest{WarrantID: priorID, Status: priorStatus, Message: &priorMessage}, nil
		}
	}
	if err != nil && !errors.Is(err, pgx.ErrNoRows) {
		return WarrantRequest{}, fmt.Errorf("check prior residential search: %w", err)
	}
	seed, worldVersion, err := store.worldSeedAndVersion(ctx, serverID)
	if err != nil {
		return WarrantRequest{}, err
	}
	person, err := world.GenerateNPCFromKey(seed, worldVersion, generationKey)
	if err != nil {
		return WarrantRequest{}, fmt.Errorf("reconstruct residential search subject: %w", err)
	}
	var archetype string
	var isKiller, isVictim bool
	var caseRole string
	var caseID, caseSceneID string
	err = store.db.QueryRow(ctx, `SELECT c.case_id,c.archetype_id, c.killer_entity_id=$3, c.victim_entity_id=$3,
		CASE WHEN c.killer_entity_id=$3 THEN 'killer' WHEN c.victim_entity_id=$3 THEN 'victim'
		WHEN EXISTS (SELECT 1 FROM case_relationships cr JOIN employment_relationships er USING(server_id,relationship_id)
			WHERE cr.server_id=c.server_id AND cr.case_id=c.case_id AND cr.relationship_role='paramour_employment' AND er.employee_npc_id=$3) THEN 'paramour'
		WHEN EXISTS (SELECT 1 FROM employment_relationships a JOIN employment_relationships b ON b.server_id=a.server_id AND b.employer_company_id=a.employer_company_id WHERE a.server_id=c.server_id AND a.employee_npc_id=$3 AND b.employee_npc_id IN (c.killer_entity_id,c.victim_entity_id)) THEN 'colleague'
		ELSE 'associate' END,c.location_entity_id
		FROM cases c JOIN case_states s USING (server_id, case_id)
		WHERE c.server_id=$1::uuid AND c.city_id=$2 AND (c.killer_entity_id=$3 OR c.victim_entity_id=$3
		OR EXISTS (SELECT 1 FROM case_relationships cr JOIN employment_relationships er USING(server_id,relationship_id)
			WHERE cr.server_id=c.server_id AND cr.case_id=c.case_id AND cr.relationship_role='paramour_employment' AND er.employee_npc_id=$3)
		OR EXISTS (SELECT 1 FROM employment_relationships a JOIN employment_relationships b ON b.server_id=a.server_id AND b.employer_company_id=a.employer_company_id WHERE a.server_id=c.server_id AND a.employee_npc_id=$3 AND b.employee_npc_id IN (c.killer_entity_id,c.victim_entity_id))
		OR EXISTS (SELECT 1 FROM personal_relationships r WHERE r.server_id=c.server_id AND ((r.person_a_npc_id=$3 AND r.person_b_npc_id IN (c.killer_entity_id,c.victim_entity_id)) OR (r.person_b_npc_id=$3 AND r.person_a_npc_id IN (c.killer_entity_id,c.victim_entity_id)))) )
		AND s.status IN ('open','solved') AND COALESCE(c.crime_at,c.created_at)<=$4 ORDER BY c.created_at DESC LIMIT 1`, serverID, cityID, npcID, now.UTC()).Scan(&caseID, &archetype, &isKiller, &isVictim, &caseRole, &caseSceneID)
	if errors.Is(err, pgx.ErrNoRows) {
		archetype = ""
	} else if err != nil {
		return WarrantRequest{}, fmt.Errorf("load residential search context: %w", err)
	}
	stamp := now.UTC()
	identity := sha256.Sum256([]byte(fmt.Sprintf("%s|%s|%s", accountID, cityID, npcID)))
	warrantID := "warrant:" + hex.EncodeToString(identity[:12])
	existingReport := priorID != ""
	commonFindings := make([]residentialFinding, 0, 60)
	for _, path := range []string{"everyday_documents.json", "everyday_papers.json", "everyday_objects.json"} {
		findings, err := readResidentialFindingList("residential_search/" + path)
		if err != nil {
			return WarrantRequest{}, fmt.Errorf("load residential inventory catalog: %w", err)
		}
		commonFindings = append(commonFindings, findings...)
	}
	if len(commonFindings) < 8 {
		return WarrantRequest{}, errors.New("residential inventory catalog must contain at least eight findings")
	}
	random := rand.New(rand.NewSource(int64(binary.BigEndian.Uint64(identity[:8]))))
	order := random.Perm(len(commonFindings))
	items := make([]PlayerMailAttachment, 0, 8)
	for index := 0; index < 7; index++ {
		items = append(items, makeResidentialAttachment(warrantID, index, commonFindings[order[index]]))
	}
	var contextualFindings []residentialFinding
	if archetype != "" {
		caseFindings, err := readResidentialCaseFindings(archetype)
		if err != nil {
			return WarrantRequest{}, fmt.Errorf("load case-specific residential findings: %w", err)
		}
		switch caseRole {
		case "killer":
			contextualFindings = caseFindings.Killer
		case "victim":
			contextualFindings = caseFindings.Victim
		case "colleague":
			contextualFindings = caseFindings.Colleague
		case "associate":
			contextualFindings = caseFindings.Associate
		case "paramour":
			contextualFindings = caseFindings.Paramour
		}
	}
	lastFinding := commonFindings[order[7]]
	if len(contextualFindings) > 0 {
		lastFinding = contextualFindings[random.Intn(len(contextualFindings))]
	}
	if caseID != "" && archetype == "financial_extortion" {
		var recordData []byte
		err := store.db.QueryRow(ctx, `SELECT record_data FROM case_activity_records
			WHERE server_id=$1::uuid AND case_id=$2 AND (npc_entity_id=$3 OR related_npc_entity_id=$3)
			ORDER BY occurred_at,record_id LIMIT 1`, serverID, caseID, npcID).Scan(&recordData)
		if err == nil {
			var record records.Record
			if json.Unmarshal(recordData, &record) == nil {
				lastFinding = residentialFindingFromRecord(record)
			}
		} else if !errors.Is(err, pgx.ErrNoRows) {
			return WarrantRequest{}, fmt.Errorf("load residential record detail: %w", err)
		}
	}
	items = append(items, makeResidentialAttachment(warrantID, 7, lastFinding))
	bodyFound := false
	if isVictim && caseID != "" {
		residential, err := store.isResidentialCaseScene(ctx, serverID, caseSceneID)
		if err != nil {
			return WarrantRequest{}, fmt.Errorf("check residential body discovery: %w", err)
		}
		if residential {
			bodyFound = true
			items = append(items, makeResidentialAttachment(warrantID, len(items), residentialFinding{
				Title:   "Body located",
				Content: fmt.Sprintf("The search team located %s at the reported address. The body was secured and transferred for examination; this report does not identify a suspect.", person.Name),
				Kind:    "body_found",
			}))
		}
	}
	if isKiller && caseID != "" {
		residential, err := store.isResidentialCaseScene(ctx, serverID, caseSceneID)
		if err != nil {
			return WarrantRequest{}, fmt.Errorf("check residential case scene: %w", err)
		}
		if residentialFingerprintEligible(caseID, residential) {
			codes, err := store.cityFingerprintCodes(ctx, serverID, cityID)
			if err != nil {
				return WarrantRequest{}, fmt.Errorf("generate city fingerprint codes: %w", err)
			}
			if code := codes[npcID]; len(code) >= 4 {
				finding := residentialFinding{Title: "Partial fingerprint trace", Content: fmt.Sprintf("A partial fingerprint code recovered during the residential search begins %s. The report does not identify its owner.", code[:4]), Kind: "fingerprint"}
				items = append(items, makeResidentialAttachment(warrantID, len(items), finding))
			}
		}
	}
	if isKiller && caseID != "" && binary.BigEndian.Uint64(identity[8:16])%4 == 0 {
		weapons := []residentialFinding{
			{Title: "Unregistered revolver", Content: "A small-calibre revolver was recovered wrapped in oilcloth. Its serial plate is filed smooth; the report records no ballistic comparison.", Kind: "weapon"},
			{Title: "Heavy iron baton", Content: "A short iron baton was found beneath stored linens. One end carries a dark, unidentified stain; no laboratory comparison is included.", Kind: "weapon"},
			{Title: "Carpenter's mallet", Content: "A wooden mallet with a dense iron head was recovered from a locked cupboard. The handle has been recently scrubbed.", Kind: "weapon"},
			{Title: "Hunting knife", Content: "A broad-bladed hunting knife was found inside a tool roll. The blade has been cleaned and the report makes no forensic attribution.", Kind: "weapon"},
		}
		weaponIndex := int(identity[16]) % len(weapons)
		items = append(items, makeResidentialAttachment(warrantID, len(items), weapons[weaponIndex]))
	}
	attachmentsJSON, err := json.Marshal(items)
	if err != nil {
		return WarrantRequest{}, err
	}
	tx, err := store.db.Begin(ctx)
	if err != nil {
		return WarrantRequest{}, fmt.Errorf("begin residential search: %w", err)
	}
	defer func() { _ = tx.Rollback(context.Background()) }()
	if existingReport {
		warrantID = priorID
		if _, err := tx.Exec(ctx, `UPDATE player_search_warrants SET catalog_version=$7,station_location_id=$5,created_at=$6
			WHERE server_id=$1::uuid AND account_id=$2::uuid AND warrant_id=$3 AND city_id=$4`, serverID, accountID, warrantID, cityID, stationLocationID, stamp, residentialSearchCatalogVersion); err != nil {
			return WarrantRequest{}, fmt.Errorf("update residential search report: %w", err)
		}
	} else if _, err := tx.Exec(ctx, `INSERT INTO player_search_warrants (server_id,account_id,warrant_id,city_id,station_location_id,target_npc_id,status,catalog_version,created_at)
		VALUES ($1::uuid,$2::uuid,$3,$4,$5,$6,'granted',$8,$7)
		ON CONFLICT (server_id,account_id,warrant_id) DO UPDATE SET
		station_location_id=EXCLUDED.station_location_id,status='granted',catalog_version=EXCLUDED.catalog_version,created_at=EXCLUDED.created_at`, serverID, accountID, warrantID, cityID, stationLocationID, npcID, stamp, residentialSearchCatalogVersion); err != nil {
		return WarrantRequest{}, fmt.Errorf("record residential search authorization: %w", err)
	}
	subject := "Residential search report · " + person.Name
	bodyText := "The authorized residential search concerning " + person.Name + " is complete. The enclosed inventory lists all items recovered and secured. Review each item in context; routine household effects were included in the same report."
	if bodyFound {
		subject = "Body found · " + person.Name
		bodyText = "The search team located the missing resident at the reported address. The body was secured and transferred for examination. The city newspaper will report the discovery."
	}
	bodyText += billDispatchPostscript(warrantID)
	message := PlayerMailMessage{ID: warrantID + ":mail", SentAt: stamp, Subject: subject, Body: bodyText, Attachments: items}
	if _, err := tx.Exec(ctx, `INSERT INTO player_inbox_messages (server_id,account_id,message_id,warrant_id,sent_at,subject,body,attachments)
		VALUES ($1::uuid,$2::uuid,$3,$4,$5,$6,$7,$8::jsonb)
		ON CONFLICT (server_id,account_id,message_id) DO UPDATE SET sent_at=EXCLUDED.sent_at,subject=EXCLUDED.subject,body=EXCLUDED.body,attachments=EXCLUDED.attachments`, serverID, accountID, message.ID, warrantID, stamp, message.Subject, message.Body, string(attachmentsJSON)); err != nil {
		return WarrantRequest{}, fmt.Errorf("deliver residential search report: %w", err)
	}
	if _, err := tx.Exec(ctx, `INSERT INTO player_discovered_items (server_id,account_id,item_id,item_source,item_data)
		SELECT $1::uuid,$2::uuid,item->>'id',item->>'source',item FROM jsonb_array_elements($3::jsonb) AS items(item)
		ON CONFLICT (server_id,account_id,item_id) DO UPDATE SET item_source=EXCLUDED.item_source,item_data=EXCLUDED.item_data`, serverID, accountID, string(attachmentsJSON)); err != nil {
		return WarrantRequest{}, fmt.Errorf("save search attachments: %w", err)
	}
	if bodyFound {
		if _, err := tx.Exec(ctx, `UPDATE case_states SET body_found_at=LEAST(COALESCE(body_found_at,$3),$3),body_found_method='residential_search',updated_at=$3
			WHERE server_id=$1::uuid AND case_id=$2`, serverID, caseID, stamp); err != nil {
			return WarrantRequest{}, fmt.Errorf("record residential body discovery: %w", err)
		}
		if _, err := tx.Exec(ctx, `INSERT INTO world_mutations(server_id,entity_id,field_key,value)
			SELECT $1::uuid,$2,'status','"dead"'::jsonb WHERE EXISTS (
				SELECT 1 FROM world_entities WHERE server_id=$1::uuid AND entity_id=$2 AND entity_type='npc')`, serverID, npcID); err != nil {
			return WarrantRequest{}, fmt.Errorf("update discovered resident status: %w", err)
		}
	}
	if err := tx.Commit(ctx); err != nil {
		return WarrantRequest{}, fmt.Errorf("commit residential search: %w", err)
	}
	return WarrantRequest{WarrantID: warrantID, Status: "granted", Message: &message, BodyFound: bodyFound}, nil
}

func residentialFindingFromRecord(record records.Record) residentialFinding {
	date := record.OccurredAt.UTC().Format("2 January")
	content := fmt.Sprintf("A folded %s copy is dated %s. %s", record.Source, date, strings.TrimSpace(record.Description))
	if record.Reference != "" {
		content += " The reference is " + record.Reference + "."
	}
	if record.Amount != 0 {
		content += fmt.Sprintf(" The amount entered is %d Credits.", record.Amount)
	}
	if record.Origin != "" || record.Destination != "" {
		content += fmt.Sprintf(" The route fields read %s to %s.", record.Origin, record.Destination)
	}
	kind := "document"
	if record.Source == "gate" {
		kind = "paper"
	}
	return residentialFinding{Title: "Folded " + strings.ToUpper(record.Source[:1]) + record.Source[1:] + " copy", Content: content, Kind: kind}
}

// RecordLocalAccessAttempt stores the backend's authorization outcome for a restricted local system.
func (store *PostgresStore) RecordLocalAccessAttempt(ctx context.Context, serverID, accountID, cityID, locationID, companyID, facilityKind, status string, attemptedAt time.Time) (LocalAccessAttempt, error) {
	if status != "granted" && status != "denied" {
		return LocalAccessAttempt{}, ErrInvalidLocalAccess
	}
	var attempt LocalAccessAttempt
	err := store.db.QueryRow(ctx, `INSERT INTO player_local_access_attempts
		(server_id,account_id,city_id,location_id,company_id,facility_kind,status,attempted_at)
		VALUES ($1::uuid,$2::uuid,$3,$4,$5,$6,$7,$8)
		RETURNING attempt_id,status,attempted_at`, serverID, accountID, cityID, locationID, companyID, facilityKind, status, attemptedAt.UTC()).Scan(&attempt.AttemptID, &attempt.Status, &attempt.Attempted)
	if err != nil {
		return LocalAccessAttempt{}, fmt.Errorf("record local access attempt: %w", err)
	}
	return attempt, nil
}

// ValidateLocalAccessAttempt checks ownership, scope, result, and the short local-session lifetime.
func (store *PostgresStore) ValidateLocalAccessAttempt(ctx context.Context, serverID, accountID string, attemptID int64, cityID, locationID, companyID, facilityKind string, now time.Time) (bool, error) {
	var valid bool
	err := store.db.QueryRow(ctx, `SELECT EXISTS (
		SELECT 1 FROM player_local_access_attempts
		WHERE server_id=$1::uuid AND account_id=$2::uuid AND attempt_id=$3
		  AND city_id=$4 AND location_id=$5 AND company_id=$6 AND facility_kind=$7
		  AND status='granted' AND attempted_at >= $8::timestamptz - interval '30 minutes' AND attempted_at <= $8::timestamptz
	)`, pgx.QueryExecModeSimpleProtocol, serverID, accountID, attemptID, cityID, locationID, companyID, facilityKind, now.UTC()).Scan(&valid)
	if err != nil {
		return false, fmt.Errorf("validate local access attempt: %w", err)
	}
	return valid, nil
}

func makeResidentialAttachment(warrantID string, index int, finding residentialFinding) PlayerMailAttachment {
	metadata, _ := json.Marshal(map[string]string{"kind": finding.Kind, "origin": "residential search", "warrantId": warrantID})
	return PlayerMailAttachment{
		DiscoveredItem: DiscoveredItem{
			ID:       fmt.Sprintf("%s:%02d", warrantID, index+1),
			Source:   "case",
			Title:    finding.Title,
			Content:  finding.Content,
			Metadata: metadata,
		},
		Kind: finding.Kind,
	}
}

func (store *PostgresStore) QueueNPCEmailQuestion(ctx context.Context, serverID, accountID string, employee world.CompanyEmployee, questionID, requestID string, now time.Time) (PlayerEmailQuestionRequest, error) {
	question, valid := playerEmailQuestions[questionID]
	if !valid || requestID == "" || employee.ID == "" || employee.Email == "" {
		return PlayerEmailQuestionRequest{}, ErrInvalidProgress
	}
	now = now.UTC()
	var canRespond bool
	err := store.db.QueryRow(ctx, `SELECT EXISTS (
		SELECT 1 FROM world_entities e
		WHERE e.server_id=$1::uuid AND e.entity_id=$2 AND e.entity_type='npc'
		AND COALESCE((SELECT value #>> '{}' FROM world_mutations m
			WHERE m.server_id=e.server_id AND m.entity_id=e.entity_id AND m.field_key='status'
			ORDER BY m.mutation_order DESC LIMIT 1),'alive') NOT IN ('dead','missing')
	)`, serverID, employee.ID).Scan(&canRespond)
	if err != nil {
		return PlayerEmailQuestionRequest{}, err
	}
	answer := "[no reply generated]"
	replyGenerated := canRespond
	var deliveredAt any
	if canRespond {
		answers, err := store.InterrogationAnswersForNPC(ctx, serverID, employee.ID)
		if err != nil {
			return PlayerEmailQuestionRequest{}, err
		}
		pool := answers[questionID]
		if len(pool) == 0 {
			return PlayerEmailQuestionRequest{}, ErrInvalidProgress
		}
		digest := sha256.Sum256([]byte(serverID + "/" + accountID + "/" + employee.ID + "/" + requestID + "/" + questionID))
		answer = pool[binary.BigEndian.Uint64(digest[:8])%uint64(len(pool))]
	} else {
		// Persist the accepted request for idempotency, but mark it consumed without creating a reply.
		deliveredAt = now
	}
	readyAt := now.Add(15 * time.Second)
	subject := "Re: " + question.prompt
	_, err = store.db.Exec(ctx, `INSERT INTO player_npc_email_questions(server_id,account_id,request_id,target_npc_id,target_name,target_email,question_id,subject,reply_body,reply_generated,ready_at,delivered_at)
		VALUES($1::uuid,$2::uuid,$3,$4,$5,$6,$7,$8,$9,$10,$11,$12) ON CONFLICT(server_id,account_id,request_id) DO NOTHING`, serverID, accountID, requestID, employee.ID, employee.Name, strings.ToLower(employee.Email), questionID, subject, answer, replyGenerated, readyAt, deliveredAt)
	if err != nil {
		return PlayerEmailQuestionRequest{}, err
	}
	var result PlayerEmailQuestionRequest
	err = store.db.QueryRow(ctx, `SELECT request_id,target_npc_id,target_name,target_email,question_id,subject,
		CASE WHEN ready_at<=now() THEN 'ready' ELSE 'pending' END,ready_at
		FROM player_npc_email_questions WHERE server_id=$1::uuid AND account_id=$2::uuid AND request_id=$3`, serverID, accountID, requestID).
		Scan(&result.RequestID, &result.NPCID, &result.Name, &result.Email, &result.QuestionID, &result.Subject, &result.Status, &result.ReadyAt)
	if err != nil {
		return PlayerEmailQuestionRequest{}, err
	}
	if result.NPCID != employee.ID || result.Email != strings.ToLower(employee.Email) || result.QuestionID != questionID {
		return PlayerEmailQuestionRequest{}, ErrPersistedConflict
	}
	return result, nil
}

func (store *PostgresStore) deliverNPCEmailReplies(ctx context.Context, serverID, accountID string) error {
	return store.deliverNPCEmailRepliesMatching(ctx, serverID, accountID, "")
}

func (store *PostgresStore) deliverNPCEmailRepliesForNPC(ctx context.Context, serverID, npcID string) error {
	return store.deliverNPCEmailRepliesMatching(ctx, serverID, "", npcID)
}

func (store *PostgresStore) deliverNPCEmailRepliesMatching(ctx context.Context, serverID, accountID, npcID string) error {
	tx, err := store.db.Begin(ctx)
	if err != nil {
		return err
	}
	defer func() { _ = tx.Rollback(context.Background()) }()
	query := `SELECT q.account_id,q.request_id,q.target_npc_id,q.target_name,q.subject,q.reply_body,q.ready_at,
		EXISTS(SELECT 1 FROM world_entities e WHERE e.server_id=q.server_id AND e.entity_id=q.target_npc_id AND e.entity_type='npc'),
		COALESCE((SELECT value #>> '{}' FROM world_mutations m WHERE m.server_id=q.server_id AND m.entity_id=q.target_npc_id AND m.field_key='status' ORDER BY m.mutation_order DESC LIMIT 1),'alive')
		FROM player_npc_email_questions q WHERE q.server_id=$1::uuid AND q.delivered_at IS NULL AND q.ready_at<=now()`
	args := []any{serverID}
	if accountID != "" {
		query += ` AND q.account_id=$2::uuid`
		args = append(args, accountID)
	}
	if npcID != "" {
		query += fmt.Sprintf(` AND q.target_npc_id=$%d`, len(args)+1)
		args = append(args, npcID)
	}
	query += ` ORDER BY q.ready_at,q.request_id FOR UPDATE OF q SKIP LOCKED`
	rows, err := tx.Query(ctx, query, args...)
	if err != nil {
		return err
	}
	type dueReply struct {
		accountID string
		requestID string
		targetNPC string
		sender    string
		subject   string
		body      string
		readyAt   time.Time
		npcExists bool
		npcStatus string
	}
	due := make([]dueReply, 0)
	for rows.Next() {
		var item dueReply
		if err := rows.Scan(&item.accountID, &item.requestID, &item.targetNPC, &item.sender, &item.subject, &item.body, &item.readyAt, &item.npcExists, &item.npcStatus); err != nil {
			rows.Close()
			return err
		}
		due = append(due, item)
	}
	if err := rows.Err(); err != nil {
		rows.Close()
		return err
	}
	rows.Close()
	for _, item := range due {
		if !item.npcExists || item.npcStatus == "dead" || item.npcStatus == "missing" {
			if _, err := tx.Exec(ctx, `UPDATE player_npc_email_questions SET delivered_at=now(),reply_generated=false WHERE server_id=$1::uuid AND account_id=$2::uuid AND request_id=$3 AND delivered_at IS NULL`, serverID, item.accountID, item.requestID); err != nil {
				return err
			}
			continue
		}
		messageID := "npc-question:" + item.requestID + ":reply"
		if _, err := tx.Exec(ctx, `INSERT INTO player_inbox_messages(server_id,account_id,message_id,sent_at,sender,message_category,subject,body,attachments)
			VALUES($1::uuid,$2::uuid,$3,$4,$5,'official',$6,$7,'[]'::jsonb) ON CONFLICT(server_id,account_id,message_id) DO NOTHING`,
			serverID, item.accountID, messageID, item.readyAt, item.sender, item.subject, item.body); err != nil {
			return err
		}
		if _, err := tx.Exec(ctx, `UPDATE player_npc_email_questions SET delivered_at=now() WHERE server_id=$1::uuid AND account_id=$2::uuid AND request_id=$3 AND delivered_at IS NULL`, serverID, item.accountID, item.requestID); err != nil {
			return err
		}
	}
	return tx.Commit(ctx)
}

func (store *PostgresStore) ListPlayerMail(ctx context.Context, serverID, accountID string) ([]PlayerMailMessage, error) {
	if err := store.deliverNPCEmailReplies(ctx, serverID, accountID); err != nil {
		return nil, fmt.Errorf("deliver NPC email replies: %w", err)
	}
	if err := store.ensureSampleSpam(ctx, serverID, accountID); err != nil {
		return nil, fmt.Errorf("prepare junk mail: %w", err)
	}
	rows, err := store.db.Query(ctx, `SELECT message.message_id,COALESCE(message.case_id,''),COALESCE(message.crime_id,''),COALESCE(crime.crime_type,''),message.sent_at,message.sender,message.message_category,message.subject,message.body,message.attachments
		FROM player_inbox_messages message LEFT JOIN case_crimes crime ON crime.server_id=message.server_id AND crime.case_id=message.case_id AND crime.crime_id=message.crime_id
		WHERE message.server_id=$1::uuid AND message.account_id=$2::uuid AND message.deleted_at IS NULL ORDER BY message.sent_at DESC,message.message_id DESC`, serverID, accountID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	messages := make([]PlayerMailMessage, 0)
	for rows.Next() {
		var message PlayerMailMessage
		var attachments []byte
		if err := rows.Scan(&message.ID, &message.CaseID, &message.CrimeID, &message.CrimeType, &message.SentAt, &message.Sender, &message.Category, &message.Subject, &message.Body, &attachments); err != nil {
			return nil, err
		}
		if err := json.Unmarshal(attachments, &message.Attachments); err != nil {
			return nil, err
		}
		messages = append(messages, message)
	}
	return messages, rows.Err()
}

func (store *PostgresStore) ensureSampleSpam(ctx context.Context, serverID, accountID string) error {
	transactional, ok := store.db.(interface {
		Begin(context.Context) (pgx.Tx, error)
	})
	if !ok {
		return errors.New("spam mail transactions are unavailable")
	}
	tx, err := transactional.Begin(ctx)
	if err != nil {
		return err
	}
	defer tx.Rollback(ctx)
	stamp := time.Now().UTC()
	spam := []struct {
		sender, subject, body string
	}{
		{"Royal Pigeon Mutual", "Your complimentary pigeon is waiting", "Dear resident, your complimentary pigeon has been reserved. Please send three crumbs to confirm delivery. The bird has already declined to carry this notice, so we are sending it by post."},
		{"The Clock Recovery Bureau", "The town clock owes you eleven minutes", "Our patented clock audit found that the town clock may owe you eleven minutes, subject to the railway's objection. Reply with your preferred minute. We cannot guarantee it will be the correct one."},
		{"Café des Écoles Spoon Committee", "A teaspoon has selected you for a survey", "One teaspoon would like to ask whether the drawer is too small. A second teaspoon denies any complaint. A third has gone missing and is therefore unavailable for comment. A modest fee secures the full report."},
	}
	values := make([]string, 0, len(spam))
	args := []any{serverID, accountID}
	for index, message := range spam {
		first := len(args) + 1
		values = append(values, fmt.Sprintf("($1::uuid,$2::uuid,$%d,$%d,$%d,'spam',$%d,$%d,'[]'::jsonb)", first, first+1, first+2, first+3, first+4))
		args = append(args, fmt.Sprintf("spam:circular:%02d", index+1), stamp.Add(-time.Duration(index)*24*time.Hour), message.sender, message.subject, message.body)
	}
	_, err = tx.Exec(ctx, `INSERT INTO player_inbox_messages
		(server_id,account_id,message_id,sent_at,sender,message_category,subject,body,attachments) VALUES `+strings.Join(values, ",")+
		` ON CONFLICT(server_id,account_id,message_id) DO NOTHING`, args...)
	if err != nil {
		return err
	}
	if _, err := tx.Exec(ctx, `INSERT INTO player_spam_mail_schedule(server_id,account_id,next_delivery_at)
		VALUES($1::uuid,$2::uuid,now()+interval '6 hours') ON CONFLICT(server_id,account_id) DO NOTHING`, serverID, accountID); err != nil {
		return err
	}
	var sequence int64
	var due bool
	if err := tx.QueryRow(ctx, `SELECT sequence,next_delivery_at<=now()
		FROM player_spam_mail_schedule WHERE server_id=$1::uuid AND account_id=$2::uuid FOR UPDATE`, serverID, accountID).Scan(&sequence, &due); err != nil {
		return err
	}
	if due {
		message := recurringSpamMessage(sequence)
		messageID := fmt.Sprintf("spam:circular:%06d", sequence+1)
		if _, err := tx.Exec(ctx, `INSERT INTO player_inbox_messages(server_id,account_id,message_id,sent_at,sender,message_category,subject,body,attachments)
			VALUES($1::uuid,$2::uuid,$3,now(),$4,'spam',$5,$6,'[]'::jsonb)
			ON CONFLICT(server_id,account_id,message_id) DO NOTHING`, serverID, accountID, messageID, message.sender, message.subject, message.body); err != nil {
			return err
		}
		if _, err := tx.Exec(ctx, `UPDATE player_spam_mail_schedule
			SET sequence=sequence+1,next_delivery_at=now()+interval '6 hours',updated_at=now()
			WHERE server_id=$1::uuid AND account_id=$2::uuid`, serverID, accountID); err != nil {
			return err
		}
	}
	return tx.Commit(ctx)
}

func recurringSpamMessage(sequence int64) struct{ sender, subject, body string } {
	messages := [...]struct{ sender, subject, body string }{
		{"Royal Pigeon Mutual", "Your complimentary pigeon is waiting", "Dear resident, your complimentary pigeon has been reserved. Please send three crumbs to confirm delivery. The bird has already declined to carry this notice, so we are sending it by post."},
		{"The Clock Recovery Bureau", "The town clock owes you eleven minutes", "Our patented clock audit found that the town clock may owe you eleven minutes, subject to the railway's objection. Reply with your preferred minute. We cannot guarantee it will be the correct one."},
		{"Café des Écoles Spoon Committee", "A teaspoon has selected you for a survey", "One teaspoon would like to ask whether the drawer is too small. A second teaspoon denies any complaint. A third has gone missing and is therefore unavailable for comment. A modest fee secures the full report."},
		{"Maison du Parapluie", "A forecast specially prepared for you", "Our weather desk predicts a strong chance of umbrellas this week. Reply with your coat size and a small handling fee to receive the forecast before the clouds do."},
		{"National Turnip Register", "Your household may qualify for a turnip", "A turnip has been set aside in your name, pending proof that you have a suitable basket. The register accepts payment in advance and questions on Thursdays only."},
		{"Rue des Lanternes Society", "An invitation to our invisible lamp sale", "Our finest lamps are currently invisible to the public. Reserve one today and be the first to notice when it arrives. Delivery charges are calculated separately."},
		{"The Society for Retired Keys", "We found a lock that may suit you", "A retired key has asked us to find it a quiet lock. Send your address and a modest deposit. The key's references are sealed for its own protection."},
	}
	return messages[int(sequence%int64(len(messages)))]
}

func (store *PostgresStore) DeletePlayerMail(ctx context.Context, serverID, accountID, messageID string) error {
	result, err := store.db.Exec(ctx, `UPDATE player_inbox_messages SET deleted_at=now()
		WHERE server_id=$1::uuid AND account_id=$2::uuid AND message_id=$3 AND deleted_at IS NULL`, serverID, accountID, messageID)
	if err != nil {
		return err
	}
	if result.RowsAffected() != 1 {
		return ErrPlayerMailNotFound
	}
	return nil
}

// Create persists a compact case definition and its relationships and evidence references.
// Exists reports whether a deterministic case ID is already stored. Startup
// uses this to preserve opening cases created before a world-content upgrade.
func (store *PostgresStore) Exists(ctx context.Context, serverID, caseID string) (bool, error) {
	var exists bool
	err := store.db.QueryRow(ctx, `SELECT EXISTS(SELECT 1 FROM cases WHERE server_id=$1::uuid AND case_id=$2)`, serverID, caseID).Scan(&exists)
	return exists, err
}

func (store *PostgresStore) Create(ctx context.Context, serverID string, base world.World, definition Definition) error {
	if err := Validate(definition, base); err != nil {
		return err
	}
	tx, err := store.db.Begin(ctx)
	if err != nil {
		return fmt.Errorf("begin case creation: %w", err)
	}
	defer func() { _ = tx.Rollback(context.Background()) }()

	var existing persistedCaseIdentity
	var crimeAt any
	var bodyFoundAt any
	var bodyFoundMethod any
	if !definition.CrimeAt.IsZero() {
		crimeAt = definition.CrimeAt
	}
	if definition.CaseType == "homicide" && !caseSceneIsResidential(base, definition.LocationID) {
		foundAfter := definition.CrimeAt
		if foundAfter.IsZero() {
			foundAfter = time.Now().UTC()
		}
		digest := sha256.Sum256([]byte(definition.ID + "/automatic-body-discovery"))
		delay := 30 + int(binary.BigEndian.Uint32(digest[:4])%180)
		bodyFoundAt, bodyFoundMethod = foundAfter.Add(time.Duration(delay)*time.Minute), "automatic"
	}
	err = tx.QueryRow(ctx,
		`SELECT case_seed, city_id, generator_version, case_type, archetype_id, COALESCE(company_id,''), motive, COALESCE(victim_entity_id,''), killer_entity_id, COALESCE(location_entity_id,'')
		 FROM cases WHERE server_id = $1::uuid AND case_id = $2 FOR UPDATE`, serverID, definition.ID,
	).Scan(&existing.Seed, &existing.CityID, &existing.GeneratorVersion, &existing.CaseType, &existing.ArchetypeID, &existing.CompanyID, &existing.Motive, &existing.VictimID, &existing.KillerID, &existing.LocationID)
	if err == nil {
		if !existing.matches(definition) {
			return ErrPersistedConflict
		}
		if definition.GeneratorVersion >= connectedGeneratorVersion && crimeAt != nil {
			if _, err := tx.Exec(ctx, `UPDATE cases c SET crime_at=COALESCE(c.crime_at,
				(SELECT activity.occurred_at + interval '19 minutes' FROM case_activity_records activity
				 WHERE activity.server_id=c.server_id AND activity.case_id=c.case_id AND activity.internal_relation='scene-killer'
				 ORDER BY activity.occurred_at DESC LIMIT 1), $3)
				WHERE c.server_id=$1::uuid AND c.case_id=$2 AND c.generator_version >= $4 AND c.crime_at IS NULL`,
				serverID, definition.ID, crimeAt, connectedGeneratorVersion); err != nil {
				return fmt.Errorf("backfill persisted case crime time: %w", err)
			}
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
	rivalIDs := []string(nil)
	if definition.CompanyID != "" && (definition.ArchetypeID == "workplace_conflict" || definition.ArchetypeID == "financial_extortion") {
		rivalIDs, err = store.companyRivalIDs(ctx, serverID, definition.CompanyID)
		if err != nil {
			return fmt.Errorf("resolve company market rivals: %w", err)
		}
	}
	participants := definition.Participants
	if len(participants) == 0 {
		participants = []CaseParticipant{{NPCID: definition.KillerID, Role: "principal"}}
		if definition.VictimID != "" {
			participants = append(participants, CaseParticipant{NPCID: definition.VictimID, Role: "victim"})
		}
	}
	participantIDs := make([]string, 0, len(participants))
	for _, participant := range participants {
		participantIDs = append(participantIDs, participant.NPCID)
	}

	var entityCount int
	if err := tx.QueryRow(ctx,
		`SELECT COUNT(*) FROM world_entities
		 WHERE server_id = $1::uuid
		   AND ((entity_id = ANY($2::text[]) AND entity_type = 'npc')
	     OR ($3 <> '' AND entity_id = $3 AND entity_type = 'location'))`,
		serverID, participantIDs, definition.LocationID,
	).Scan(&entityCount); err != nil {
		return fmt.Errorf("verify case entities: %w", err)
	}
	wantEntityCount := len(participantIDs)
	if definition.LocationID != "" {
		wantEntityCount++
	}
	if entityCount != wantEntityCount {
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
		`INSERT INTO cases (server_id, case_id, case_seed, city_id, generator_version, case_type, archetype_id, company_id, motive,
		 victim_entity_id, killer_entity_id, location_entity_id, crime_at)
		 VALUES ($1::uuid, $2, $3, $4, $5, $6, $7, NULLIF($8,''), $9, NULLIF($10,''), $11, NULLIF($12,''), $13)`,
		serverID, definition.ID, definition.Seed, definition.CityID, definition.GeneratorVersion, definition.CaseType, definition.ArchetypeID, definition.CompanyID, definition.Motive,
		definition.VictimID, definition.KillerID, definition.LocationID, crimeAt,
	); err != nil {
		return fmt.Errorf("persist case definition: %w", err)
	}
	for _, participant := range participants {
		if _, err := tx.Exec(ctx, `INSERT INTO case_participants(server_id,case_id,npc_entity_id,participant_role) VALUES($1::uuid,$2,$3,$4)`, serverID, definition.ID, participant.NPCID, participant.Role); err != nil {
			return fmt.Errorf("persist case participant: %w", err)
		}
	}
	initialCrimeAt := definition.CrimeAt
	if initialCrimeAt.IsZero() {
		initialCrimeAt = time.Now().UTC()
	}
	if _, err := tx.Exec(ctx,
		`INSERT INTO case_crimes(server_id,case_id,crime_id,crime_type,occurred_at,location_entity_id,crime_sequence,source_key)
		 VALUES($1::uuid,$2,$3,$4,$5,NULLIF($6,''),1,'primary')`,
		serverID, definition.ID, definition.ID+"/crime/1", definition.CaseType, initialCrimeAt, definition.LocationID,
	); err != nil {
		return fmt.Errorf("persist initial case crime: %w", err)
	}
	for _, participant := range participants {
		if _, err := tx.Exec(ctx,
			`INSERT INTO case_crime_participants(server_id,crime_id,npc_entity_id,participant_role) VALUES($1::uuid,$2,$3,$4)`,
			serverID, definition.ID+"/crime/1", participant.NPCID, participant.Role,
		); err != nil {
			return fmt.Errorf("persist initial crime participant: %w", err)
		}
	}
	if err := applyCorporateMarketImpact(ctx, tx, serverID, definition, rivalIDs); err != nil {
		return fmt.Errorf("apply corporate market impact: %w", err)
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
			SupportsNPCIDs         []string `json:"supportsNpcIds,omitempty"`
		}{evidence.Presentation, evidence.Content, evidence.RelationshipIDs, evidence.ReachableFromEntityIDs, evidence.SupportsNPCIDs})
		if err != nil {
			return fmt.Errorf("encode case evidence: %w", err)
		}
		if _, err := tx.Exec(ctx,
			`INSERT INTO case_evidence (server_id, case_id, crime_id, evidence_id, location_entity_id, evidence_type, evidence_data)
			 VALUES ($1::uuid, $2, $3, $4, NULLIF($5,''), $6, $7::jsonb)`,
			serverID, definition.ID, definition.ID+"/crime/1", evidence.ID, evidence.LocationID, evidence.Type, string(data),
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
			`INSERT INTO case_activity_records (server_id, case_id, crime_id, record_id, npc_entity_id, related_npc_entity_id, source, occurred_at, internal_relation, record_data)
			 VALUES ($1::uuid, $2, $3, $4, $5, NULLIF($6, ''), $7, $8, $9, $10::jsonb)`,
			serverID, definition.ID, definition.ID+"/crime/1", record.ID, record.NPCID, record.RelatedNPCID, record.Source, record.OccurredAt, record.InternalRelation, string(data),
		); err != nil {
			return fmt.Errorf("persist case activity record %s: %w", record.ID, err)
		}
	}
	if err := persistCaseEmails(ctx, tx, serverID, definition); err != nil {
		return err
	}
	if _, err := tx.Exec(ctx,
		`INSERT INTO case_states (server_id, case_id, status, body_found_at, body_found_method)
		 VALUES ($1::uuid, $2, 'open', $3, $4)`,
		serverID, definition.ID, bodyFoundAt, bodyFoundMethod,
	); err != nil {
		return fmt.Errorf("persist initial shared case state: %w", err)
	}
	if contains([]string{"embezzlement", "financial_extortion", "accounting_fraud"}, definition.CaseType) && hasCompanyHeadquarters(base, definition.CompanyID) {
		if _, err := tx.Exec(ctx, `INSERT INTO case_evolution_jobs(server_id,case_id,event_key,event_type,available_at)
			VALUES($1::uuid,$2,'financial-cover-up','financial_cover_up',now()+interval '3 minutes') ON CONFLICT DO NOTHING`, serverID, definition.ID); err != nil {
			return fmt.Errorf("schedule financial case evolution: %w", err)
		}
	}
	if definition.ArchetypeID == "serial_killer" {
		for index, targetID := range definition.FutureVictimIDs {
			threatAt := time.Now().UTC().Add(time.Duration(index*2+1) * 24 * time.Hour)
			attackAt := threatAt.Add(24 * time.Hour)
			if _, err := tx.Exec(ctx, `INSERT INTO case_evolution_jobs(server_id,case_id,event_key,event_type,target_npc_entity_id,available_at)
				VALUES($1::uuid,$2,$3,'serial_threat',$4,$5) ON CONFLICT DO NOTHING`, serverID, definition.ID, fmt.Sprintf("serial-threat-%d", index+1), targetID, threatAt); err != nil {
				return fmt.Errorf("schedule serial threat: %w", err)
			}
			if _, err := tx.Exec(ctx, `INSERT INTO case_evolution_jobs(server_id,case_id,event_key,event_type,target_npc_entity_id,available_at)
				VALUES($1::uuid,$2,$3,'serial_homicide',$4,$5) ON CONFLICT DO NOTHING`, serverID, definition.ID, fmt.Sprintf("serial-homicide-%d", index+1), targetID, attackAt); err != nil {
				return fmt.Errorf("schedule serial homicide: %w", err)
			}
		}
	}
	if err := tx.Commit(ctx); err != nil {
		return fmt.Errorf("commit case creation: %w", err)
	}
	return nil
}

func hasCompanyHeadquarters(base world.World, companyID string) bool {
	for _, relation := range base.CompanyLocations {
		if relation.CompanyID == companyID && relation.LocationRole == "headquarters" {
			return true
		}
	}
	return false
}

// AppendCrime adds one occurred crime to an open investigation. sourceKey is
// the caller's stable event key, so replaying the same world event is safe.
func (store *PostgresStore) AppendCrime(ctx context.Context, serverID, caseID, crimeID, sourceKey, crimeType string, occurredAt time.Time, locationID string, participants []CaseParticipant) error {
	return store.AppendLinkedCrime(ctx, serverID, caseID, crimeID, sourceKey, crimeType, occurredAt, locationID, "", participants)
}

// AppendLinkedCrime records the causal parent when the new occurrence follows
// an earlier crime in the same investigation.
func (store *PostgresStore) AppendLinkedCrime(ctx context.Context, serverID, caseID, crimeID, sourceKey, crimeType string, occurredAt time.Time, locationID, parentCrimeID string, participants []CaseParticipant) error {
	if serverID == "" || caseID == "" || crimeID == "" || sourceKey == "" || crimeType == "" || occurredAt.IsZero() {
		return ErrInvalidDefinition
	}
	occurredAt = occurredAt.UTC().Truncate(time.Microsecond)
	if occurredAt.After(time.Now().UTC()) {
		return ErrInvalidDefinition
	}
	for _, participant := range participants {
		if participant.NPCID == "" || (participant.Role != "principal" && participant.Role != "accomplice" && participant.Role != "victim") {
			return ErrInvalidDefinition
		}
	}
	tx, err := store.db.Begin(ctx)
	if err != nil {
		return fmt.Errorf("begin case crime append: %w", err)
	}
	defer func() { _ = tx.Rollback(context.Background()) }()
	var status string
	if err := tx.QueryRow(ctx, `SELECT status FROM case_states WHERE server_id=$1::uuid AND case_id=$2 FOR UPDATE`, serverID, caseID).Scan(&status); errors.Is(err, pgx.ErrNoRows) {
		return ErrCaseNotFound
	} else if err != nil {
		return fmt.Errorf("lock case before crime append: %w", err)
	}
	if status != "open" {
		return ErrInvalidCaseState
	}
	if parentCrimeID != "" {
		var parentAt time.Time
		if err := tx.QueryRow(ctx, `SELECT occurred_at FROM case_crimes WHERE server_id=$1::uuid AND case_id=$2 AND crime_id=$3`, serverID, caseID, parentCrimeID).Scan(&parentAt); err != nil {
			if errors.Is(err, pgx.ErrNoRows) {
				return ErrCaseNotFound
			}
			return err
		}
		if parentAt.After(occurredAt) {
			return ErrInvalidDefinition
		}
	}
	var storedID, storedType, storedParentID string
	var storedAt time.Time
	var storedLocation string
	err = tx.QueryRow(ctx, `SELECT crime_id,crime_type,occurred_at,COALESCE(location_entity_id,''),COALESCE(parent_crime_id,'') FROM case_crimes WHERE server_id=$1::uuid AND case_id=$2 AND source_key=$3`, serverID, caseID, sourceKey).Scan(&storedID, &storedType, &storedAt, &storedLocation, &storedParentID)
	if err == nil {
		if storedID != crimeID || storedType != crimeType || !storedAt.Equal(occurredAt) || storedLocation != locationID || storedParentID != parentCrimeID {
			return ErrPersistedConflict
		}
		var storedParticipantCount int
		if err := tx.QueryRow(ctx, `SELECT count(*) FROM case_crime_participants WHERE server_id=$1::uuid AND crime_id=$2`, serverID, crimeID).Scan(&storedParticipantCount); err != nil {
			return err
		}
		if storedParticipantCount != len(participants) {
			return ErrPersistedConflict
		}
		for _, participant := range participants {
			var exists bool
			if err := tx.QueryRow(ctx, `SELECT EXISTS(SELECT 1 FROM case_crime_participants WHERE server_id=$1::uuid AND crime_id=$2 AND npc_entity_id=$3 AND participant_role=$4)`, serverID, crimeID, participant.NPCID, participant.Role).Scan(&exists); err != nil {
				return err
			}
			if !exists {
				return ErrPersistedConflict
			}
		}
		if err := tx.Commit(ctx); err != nil {
			return fmt.Errorf("commit repeated case crime: %w", err)
		}
		return nil
	}
	if !errors.Is(err, pgx.ErrNoRows) {
		return fmt.Errorf("check repeated case crime: %w", err)
	}
	if locationID != "" {
		var exists bool
		if err := tx.QueryRow(ctx, `SELECT EXISTS(SELECT 1 FROM world_entities WHERE server_id=$1::uuid AND entity_id=$2 AND entity_type='location')`, serverID, locationID).Scan(&exists); err != nil {
			return err
		}
		if !exists {
			return world.ErrEntityNotFound
		}
	}
	for _, participant := range participants {
		var exists bool
		if err := tx.QueryRow(ctx, `SELECT EXISTS(SELECT 1 FROM world_entities WHERE server_id=$1::uuid AND entity_id=$2 AND entity_type='npc')`, serverID, participant.NPCID).Scan(&exists); err != nil {
			return err
		}
		if !exists {
			return world.ErrEntityNotFound
		}
	}
	var sequence int
	if err := tx.QueryRow(ctx, `SELECT COALESCE(MAX(crime_sequence),0)+1 FROM case_crimes WHERE server_id=$1::uuid AND case_id=$2`, serverID, caseID).Scan(&sequence); err != nil {
		return fmt.Errorf("allocate case crime sequence: %w", err)
	}
	if _, err := tx.Exec(ctx,
		`INSERT INTO case_crimes(server_id,case_id,crime_id,crime_type,occurred_at,location_entity_id,parent_crime_id,crime_sequence,source_key)
		 VALUES($1::uuid,$2,$3,$4,$5,NULLIF($6,''),NULLIF($7,''),$8,$9)`, serverID, caseID, crimeID, crimeType, occurredAt, locationID, parentCrimeID, sequence, sourceKey,
	); err != nil {
		return fmt.Errorf("persist appended case crime: %w", err)
	}
	for _, participant := range participants {
		if _, err := tx.Exec(ctx, `INSERT INTO case_crime_participants(server_id,crime_id,npc_entity_id,participant_role) VALUES($1::uuid,$2,$3,$4)`, serverID, crimeID, participant.NPCID, participant.Role); err != nil {
			return fmt.Errorf("persist appended crime participant: %w", err)
		}
		if _, err := tx.Exec(ctx, `INSERT INTO case_participants(server_id,case_id,npc_entity_id,participant_role) VALUES($1::uuid,$2,$3,$4)
			ON CONFLICT(server_id,case_id,npc_entity_id) DO UPDATE SET participant_role=CASE WHEN EXCLUDED.participant_role IN ('principal','accomplice') THEN EXCLUDED.participant_role ELSE case_participants.participant_role END`, serverID, caseID, participant.NPCID, participant.Role); err != nil {
			return fmt.Errorf("update case participant index: %w", err)
		}
	}
	if err := tx.Commit(ctx); err != nil {
		return fmt.Errorf("commit appended case crime: %w", err)
	}
	return nil
}

// AttachCrimeEvidence stores evidence against the specific occurred crime that
// produced it. Existing evidence remains available to the case as a whole.
func (store *PostgresStore) AttachCrimeEvidence(ctx context.Context, serverID, caseID, crimeID string, evidence EvidenceDefinition) error {
	if serverID == "" || caseID == "" || crimeID == "" || evidence.ID == "" || evidence.Type == "" {
		return ErrInvalidDefinition
	}
	data, err := json.Marshal(struct {
		Presentation           string   `json:"presentation"`
		Content                string   `json:"content"`
		RelationshipIDs        []string `json:"relationshipIds"`
		ReachableFromEntityIDs []string `json:"reachableFromEntityIds"`
		SupportsNPCIDs         []string `json:"supportsNpcIds,omitempty"`
	}{evidence.Presentation, evidence.Content, evidence.RelationshipIDs, evidence.ReachableFromEntityIDs, evidence.SupportsNPCIDs})
	if err != nil {
		return fmt.Errorf("encode crime evidence: %w", err)
	}
	tx, err := store.db.Begin(ctx)
	if err != nil {
		return fmt.Errorf("begin crime evidence attachment: %w", err)
	}
	defer func() { _ = tx.Rollback(context.Background()) }()
	var belongs bool
	if err := tx.QueryRow(ctx, `SELECT EXISTS(SELECT 1 FROM case_crimes WHERE server_id=$1::uuid AND case_id=$2 AND crime_id=$3 AND occurred_at<=now())`, serverID, caseID, crimeID).Scan(&belongs); err != nil {
		return err
	}
	if !belongs {
		return ErrEvidenceNotFound
	}
	result, err := tx.Exec(ctx,
		`INSERT INTO case_evidence(server_id,case_id,crime_id,evidence_id,location_entity_id,evidence_type,evidence_data)
		 VALUES($1::uuid,$2,$3,$4,NULLIF($5,''),$6,$7::jsonb) ON CONFLICT(server_id,case_id,evidence_id) DO NOTHING`,
		serverID, caseID, crimeID, evidence.ID, evidence.LocationID, evidence.Type, string(data),
	)
	if err != nil {
		return fmt.Errorf("persist crime evidence: %w", err)
	}
	if result.RowsAffected() == 0 {
		var existingCrimeID, existingType string
		var existingData []byte
		if err := tx.QueryRow(ctx, `SELECT crime_id,evidence_type,evidence_data FROM case_evidence WHERE server_id=$1::uuid AND case_id=$2 AND evidence_id=$3`, serverID, caseID, evidence.ID).Scan(&existingCrimeID, &existingType, &existingData); err != nil {
			return err
		}
		var equal bool
		if err := tx.QueryRow(ctx, `SELECT $1::jsonb = $2::jsonb`, string(existingData), string(data)).Scan(&equal); err != nil {
			return err
		}
		if existingCrimeID != crimeID || existingType != evidence.Type || !equal {
			return ErrPersistedConflict
		}
	}
	if err := tx.Commit(ctx); err != nil {
		return fmt.Errorf("commit crime evidence attachment: %w", err)
	}
	return nil
}

type dueEvolutionJob struct {
	caseID, cityID, companyID, locationID string
	principalID, accompliceID             string
	eventKey, eventType, targetID         string
	seed                                  string
	worldVersion                          int
	planVersion                           int
	scheduledAt                           time.Time
}

// ProcessDueCaseEvolution advances one queued, backend-only event at a time.
// Plans remain invisible until the event and its evidence are committed.
func (store *PostgresStore) ProcessDueCaseEvolution(ctx context.Context, serverID string, now time.Time) error {
	tx, err := store.db.Begin(ctx)
	if err != nil {
		return fmt.Errorf("begin case evolution tick: %w", err)
	}
	defer func() { _ = tx.Rollback(context.Background()) }()
	if _, err := tx.Exec(ctx, `UPDATE case_evolution_jobs j SET status='cancelled',completed_at=$2
		FROM case_states s WHERE j.server_id=$1::uuid AND j.server_id=s.server_id AND j.case_id=s.case_id
		AND j.status='pending' AND s.status<>'open'`, serverID, now.UTC()); err != nil {
		return fmt.Errorf("cancel closed case evolution jobs: %w", err)
	}
	rows, err := tx.Query(ctx, `SELECT j.case_id,c.city_id,COALESCE(c.company_id,''),COALESCE(c.location_entity_id,
		(SELECT cl.location_id FROM company_locations cl WHERE cl.server_id=c.server_id AND cl.company_id=c.company_id AND cl.location_role='headquarters' ORDER BY cl.relationship_id LIMIT 1),''),j.event_key,j.event_type,COALESCE(j.target_npc_entity_id,''),
		COALESCE((SELECT npc_entity_id FROM case_participants p WHERE p.server_id=c.server_id AND p.case_id=c.case_id AND p.participant_role='principal' ORDER BY npc_entity_id LIMIT 1),''),
		COALESCE((SELECT npc_entity_id FROM case_participants p WHERE p.server_id=c.server_id AND p.case_id=c.case_id AND p.participant_role='accomplice' ORDER BY npc_entity_id LIMIT 1),''),
		w.world_seed,COALESCE(city_generator.generator_version,w.generator_version),j.plan_version,j.available_at
		FROM case_evolution_jobs j JOIN cases c USING(server_id,case_id) JOIN case_states s USING(server_id,case_id)
		JOIN worlds w USING(server_id) LEFT JOIN world_city_generators city_generator ON city_generator.server_id=c.server_id AND city_generator.city_id=c.city_id
		WHERE j.server_id=$1::uuid AND j.status='pending' AND j.available_at<=$2 AND s.status='open'
		ORDER BY j.available_at,j.case_id LIMIT 10 FOR UPDATE OF j SKIP LOCKED`, serverID, now.UTC())
	if err != nil {
		return fmt.Errorf("load due case evolution jobs: %w", err)
	}
	jobs := make([]dueEvolutionJob, 0, 10)
	for rows.Next() {
		var item dueEvolutionJob
		if err := rows.Scan(&item.caseID, &item.cityID, &item.companyID, &item.locationID, &item.eventKey, &item.eventType, &item.targetID, &item.principalID, &item.accompliceID, &item.seed, &item.worldVersion, &item.planVersion, &item.scheduledAt); err != nil {
			rows.Close()
			return err
		}
		jobs = append(jobs, item)
	}
	if err := rows.Err(); err != nil {
		rows.Close()
		return err
	}
	rows.Close()
	for _, job := range jobs {
		if job.planVersion != 1 {
			return fmt.Errorf("case %s uses unsupported evolution plan version %d", job.caseID, job.planVersion)
		}
		var caseStatus string
		if err := tx.QueryRow(ctx, `SELECT status FROM case_states WHERE server_id=$1::uuid AND case_id=$2 FOR UPDATE`, serverID, job.caseID).Scan(&caseStatus); err != nil {
			return fmt.Errorf("lock case state before evolution: %w", err)
		}
		if caseStatus != "open" {
			if _, err := tx.Exec(ctx, `UPDATE case_evolution_jobs SET status='cancelled',completed_at=$3 WHERE server_id=$1::uuid AND case_id=$2 AND event_key=$4 AND status='pending'`, serverID, job.caseID, now.UTC(), job.eventKey); err != nil {
				return err
			}
			continue
		}
		if strings.HasPrefix(job.eventType, "serial_") {
			if job.planVersion != 1 || job.principalID == "" || job.targetID == "" {
				return fmt.Errorf("case %s has an incomplete serial evolution plan", job.caseID)
			}
			if err := processSerialEvolutionJob(ctx, tx, serverID, job); err != nil {
				return err
			}
			continue
		}
		if job.eventType != "financial_cover_up" || job.planVersion != 1 {
			return fmt.Errorf("case %s uses unsupported evolution event %q version %d", job.caseID, job.eventType, job.planVersion)
		}
		if job.principalID == "" || job.accompliceID == "" || job.locationID == "" {
			if _, err := tx.Exec(ctx, `UPDATE case_evolution_jobs SET status='cancelled',completed_at=$3
				WHERE server_id=$1::uuid AND case_id=$2 AND event_key=$4 AND status='pending'`, serverID, job.caseID, now.UTC(), job.eventKey); err != nil {
				return fmt.Errorf("cancel incomplete financial evolution plan for case %s: %w", job.caseID, err)
			}
			continue
		}
		var actorStatus string
		if err := tx.QueryRow(ctx, `SELECT COALESCE((SELECT value #>> '{}' FROM world_mutations WHERE server_id=$1::uuid AND entity_id=$2 AND field_key='status' AND created_at<=$3 ORDER BY mutation_order DESC LIMIT 1),'alive')`, serverID, job.accompliceID, job.scheduledAt).Scan(&actorStatus); err != nil {
			return fmt.Errorf("read planned actor status: %w", err)
		}
		if actorStatus != "alive" {
			if _, err := tx.Exec(ctx, `UPDATE case_evolution_jobs SET status='cancelled',completed_at=$3 WHERE server_id=$1::uuid AND case_id=$2 AND event_key='financial-cover-up'`, serverID, job.caseID, now.UTC()); err != nil {
				return err
			}
			continue
		}
		actorKey, err := worldGenerationKeyForNPC(ctx, tx, serverID, job.accompliceID)
		if err != nil {
			return fmt.Errorf("load planned actor key: %w", err)
		}
		actor, err := world.GenerateNPCFromKey(job.seed, job.worldVersion, actorKey)
		if err != nil {
			return fmt.Errorf("reconstruct planned actor: %w", err)
		}
		principalKey, err := worldGenerationKeyForNPC(ctx, tx, serverID, job.principalID)
		if err != nil {
			return fmt.Errorf("load case principal key: %w", err)
		}
		principal, err := world.GenerateNPCFromKey(job.seed, job.worldVersion, principalKey)
		if err != nil {
			return fmt.Errorf("reconstruct case principal: %w", err)
		}
		crimeID := job.caseID + "/evolution/financial-cover-up"
		occurredAt := job.scheduledAt.UTC().Truncate(time.Microsecond)
		var sequence int
		if err := tx.QueryRow(ctx, `SELECT COALESCE(MAX(crime_sequence),0)+1 FROM case_crimes WHERE server_id=$1::uuid AND case_id=$2`, serverID, job.caseID).Scan(&sequence); err != nil {
			return err
		}
		if _, err := tx.Exec(ctx, `INSERT INTO case_crimes(server_id,case_id,crime_id,crime_type,occurred_at,location_entity_id,parent_crime_id,crime_sequence,source_key)
			VALUES($1::uuid,$2,$3,'evidence_tampering',$4,$5,$6,$7,'financial-cover-up')`, serverID, job.caseID, crimeID, occurredAt, job.locationID, job.caseID+"/crime/1", sequence); err != nil {
			return fmt.Errorf("persist evolved case crime: %w", err)
		}
		if _, err := tx.Exec(ctx, `INSERT INTO case_crime_participants(server_id,crime_id,npc_entity_id,participant_role) VALUES($1::uuid,$2,$3,'accomplice')`, serverID, crimeID, job.accompliceID); err != nil {
			return fmt.Errorf("persist evolved crime participant: %w", err)
		}
		evidence := EvidenceDefinition{ID: crimeID + "/evidence", Type: "document", Presentation: "Altered company register", Content: "A replacement register page was inserted after the original financial entry was challenged.", LocationID: job.locationID, SupportsNPCIDs: []string{job.accompliceID}}
		evidenceData, err := json.Marshal(struct {
			Presentation           string   `json:"presentation"`
			Content                string   `json:"content"`
			RelationshipIDs        []string `json:"relationshipIds"`
			ReachableFromEntityIDs []string `json:"reachableFromEntityIds"`
			SupportsNPCIDs         []string `json:"supportsNpcIds,omitempty"`
		}{evidence.Presentation, evidence.Content, nil, nil, evidence.SupportsNPCIDs})
		if err != nil {
			return err
		}
		if _, err := tx.Exec(ctx, `INSERT INTO case_evidence(server_id,case_id,crime_id,evidence_id,location_entity_id,evidence_type,evidence_data)
			VALUES($1::uuid,$2,$3,$4,$5,$6,$7::jsonb)`, serverID, job.caseID, crimeID, evidence.ID, job.locationID, evidence.Type, string(evidenceData)); err != nil {
			return fmt.Errorf("persist evolved crime evidence: %w", err)
		}
		record := records.Record{ID: crimeID + "/gate", NPCID: job.accompliceID, Source: "gate", OccurredAt: occurredAt, Operator: "Meridian GATE Network", Description: "A staff relay entry was recorded after closing.", Destination: "Company records office", CaseID: job.caseID, InternalRelation: "financial-cover-up", RelatedNPCID: job.principalID}
		recordData, err := json.Marshal(record)
		if err != nil {
			return err
		}
		if _, err := tx.Exec(ctx, `INSERT INTO case_activity_records(server_id,case_id,crime_id,record_id,npc_entity_id,related_npc_entity_id,source,occurred_at,internal_relation,record_data)
			VALUES($1::uuid,$2,$3,$4,$5,$6,'gate',$7,'financial-cover-up',$8::jsonb)`, serverID, job.caseID, crimeID, record.ID, job.accompliceID, job.principalID, occurredAt, string(recordData)); err != nil {
			return fmt.Errorf("persist evolved crime record: %w", err)
		}
		message := mail.CaseMessage(job.seed, job.worldVersion, job.caseID, job.accompliceID, job.principalID, actor.Name, principal.Name, "The replacement register", "I replaced the page after the figures were challenged. Keep the earlier copy out of the file.", "financial_cover_up", occurredAt)
		message.Service = "Meridian Personal Mail"
		message.Recipient = principal.Name
		message.Sender = actor.Name
		encodedMessage, err := json.Marshal(message)
		if err != nil {
			return err
		}
		if _, err := tx.Exec(ctx, `INSERT INTO case_mail_messages(server_id,case_id,crime_id,message_id,mailbox_npc_entity_id,related_npc_entity_id,sent_at,internal_relation,message_data)
			VALUES($1::uuid,$2,$3,$4,$5,$6,$7,'financial_cover_up',$8::jsonb)`, serverID, job.caseID, crimeID, message.ID, job.accompliceID, job.principalID, occurredAt, string(encodedMessage)); err != nil {
			return fmt.Errorf("persist evolved crime correspondence: %w", err)
		}
		var witnessKey string
		witnessErr := tx.QueryRow(ctx, `SELECT employee.generation_key
			FROM employment_relationships employment JOIN world_entities employee
			 ON employee.server_id=employment.server_id AND employee.entity_id=employment.employee_npc_id AND employee.entity_type='npc'
			WHERE employment.server_id=$1::uuid AND employment.employer_company_id=$2
			 AND employee.entity_id<>$3 AND employee.entity_id<>$4
			 AND COALESCE((SELECT value #>> '{}' FROM world_mutations mutation WHERE mutation.server_id=employee.server_id AND mutation.entity_id=employee.entity_id AND mutation.field_key='status' AND mutation.created_at<=$5 ORDER BY mutation_order DESC LIMIT 1),'alive')='alive'
			ORDER BY employment.relationship_id LIMIT 1`, serverID, job.companyID, job.principalID, job.accompliceID, occurredAt).Scan(&witnessKey)
		if witnessErr != nil && !errors.Is(witnessErr, pgx.ErrNoRows) {
			return fmt.Errorf("find employee reporting the cover-up: %w", witnessErr)
		}
		if witnessErr == nil {
			witness, err := world.GenerateNPCFromKey(job.seed, job.worldVersion, witnessKey)
			if err != nil {
				return fmt.Errorf("reconstruct reporting employee: %w", err)
			}
			messageID := "case-event:" + crimeID + ":report"
			_, err = tx.Exec(ctx, `INSERT INTO player_inbox_messages(server_id,account_id,case_id,crime_id,message_id,sent_at,sender,message_category,subject,body,attachments)
				SELECT $1::uuid,account.id,$2,$3,$4,$5,$6,'official','A changed page in the company register',$7,'[]'::jsonb
				FROM player_accounts account ON CONFLICT DO NOTHING`,
				serverID, job.caseID, crimeID, messageID, occurredAt, witness.Name,
				"Someone replaced a register page after the figures were challenged. The company records office should still have the earlier entry. Could you look into it?")
			if err != nil {
				return fmt.Errorf("deliver case lead to player mailboxes: %w", err)
			}
		}
		if _, err := tx.Exec(ctx, `UPDATE case_evolution_jobs SET status='completed',completed_at=$3 WHERE server_id=$1::uuid AND case_id=$2 AND event_key='financial-cover-up'`, serverID, job.caseID, occurredAt); err != nil {
			return err
		}
	}
	if err := tx.Commit(ctx); err != nil {
		return fmt.Errorf("commit case evolution tick: %w", err)
	}
	return nil
}

func worldGenerationKeyForNPC(ctx context.Context, tx pgx.Tx, serverID, npcID string) (string, error) {
	var key string
	err := tx.QueryRow(ctx, `SELECT generation_key FROM world_entities WHERE server_id=$1::uuid AND entity_id=$2 AND entity_type='npc'`, serverID, npcID).Scan(&key)
	return key, err
}

func processSerialEvolutionJob(ctx context.Context, tx pgx.Tx, serverID string, job dueEvolutionJob) error {
	var actorStatus, targetStatus string
	for npcID, status := range map[string]*string{job.principalID: &actorStatus, job.targetID: &targetStatus} {
		if err := tx.QueryRow(ctx, `SELECT COALESCE((SELECT value #>> '{}' FROM world_mutations WHERE server_id=$1::uuid AND entity_id=$2 AND field_key='status' AND created_at<=$3 ORDER BY mutation_order DESC LIMIT 1),'alive')`, serverID, npcID, job.scheduledAt).Scan(status); err != nil {
			return fmt.Errorf("read serial case participant status: %w", err)
		}
	}
	if actorStatus != "alive" || targetStatus != "alive" {
		if _, err := tx.Exec(ctx, `UPDATE case_evolution_jobs SET status='cancelled',completed_at=now() WHERE server_id=$1::uuid AND case_id=$2 AND event_key=$3`, serverID, job.caseID, job.eventKey); err != nil {
			return fmt.Errorf("cancel unavailable serial case event: %w", err)
		}
		return nil
	}
	actorKey, err := worldGenerationKeyForNPC(ctx, tx, serverID, job.principalID)
	if err != nil {
		return fmt.Errorf("load serial killer key: %w", err)
	}
	targetKey, err := worldGenerationKeyForNPC(ctx, tx, serverID, job.targetID)
	if err != nil {
		return fmt.Errorf("load serial victim key: %w", err)
	}
	actor, err := world.GenerateNPCFromKey(job.seed, job.worldVersion, actorKey)
	if err != nil {
		return err
	}
	target, err := world.GenerateNPCFromKey(job.seed, job.worldVersion, targetKey)
	if err != nil {
		return err
	}
	crimeID := job.caseID + "/crime/" + job.eventKey
	parentID := job.caseID + "/crime/1"
	if err := tx.QueryRow(ctx, `SELECT crime_id FROM case_crimes WHERE server_id=$1::uuid AND case_id=$2 AND occurred_at<=$3 ORDER BY occurred_at DESC,crime_sequence DESC LIMIT 1`, serverID, job.caseID, job.scheduledAt).Scan(&parentID); err != nil {
		return fmt.Errorf("find preceding serial case event: %w", err)
	}
	var sequence int
	if err := tx.QueryRow(ctx, `SELECT COALESCE(MAX(crime_sequence),0)+1 FROM case_crimes WHERE server_id=$1::uuid AND case_id=$2`, serverID, job.caseID).Scan(&sequence); err != nil {
		return err
	}
	crimeType := "criminal_threat"
	if job.eventType == "serial_homicide" {
		crimeType = "homicide"
	}
	if _, err := tx.Exec(ctx, `INSERT INTO case_crimes(server_id,case_id,crime_id,crime_type,occurred_at,location_entity_id,parent_crime_id,crime_sequence,source_key)
		VALUES($1::uuid,$2,$3,$4,$5,NULLIF($6,''),$7,$8,$9)`, serverID, job.caseID, crimeID, crimeType, job.scheduledAt, job.locationID, parentID, sequence, job.eventKey); err != nil {
		return fmt.Errorf("persist serial case event: %w", err)
	}
	for _, participant := range []CaseParticipant{{NPCID: job.principalID, Role: "principal"}, {NPCID: job.targetID, Role: "victim"}} {
		if _, err := tx.Exec(ctx, `INSERT INTO case_crime_participants(server_id,crime_id,npc_entity_id,participant_role) VALUES($1::uuid,$2,$3,$4)`, serverID, crimeID, participant.NPCID, participant.Role); err != nil {
			return err
		}
		if _, err := tx.Exec(ctx, `INSERT INTO case_participants(server_id,case_id,npc_entity_id,participant_role) VALUES($1::uuid,$2,$3,$4)
			ON CONFLICT(server_id,case_id,npc_entity_id) DO UPDATE SET participant_role=CASE WHEN EXCLUDED.participant_role='principal' THEN 'principal' ELSE case_participants.participant_role END`, serverID, job.caseID, participant.NPCID, participant.Role); err != nil {
			return err
		}
	}
	appearance := actor.Appearance
	if appearance == "" {
		appearance = "a distinctive figure in a dark coat"
	}
	if job.eventType == "serial_threat" {
		threats := []struct{ subject, body string }{
			{"You cannot hide", "I know where you work. I know your route, too."},
			{"I am still here", "You will recognize " + appearance + " when I am near. Do not show this to anyone."},
			{"One last warning", "The people at the office know better than to interfere. Keep quiet."},
		}
		for index, threat := range threats {
			message := mail.CaseMessage(job.seed, job.worldVersion, job.caseID, target.ID, actor.ID, "Unknown sender", target.Name, threat.subject, threat.body, "anonymous_warning", job.scheduledAt.Add(time.Duration(index-len(threats)+1)*time.Hour))
			// CaseMessage is stable per case, mailbox, and relation. A threat
			// event contains several messages with the same relation, so include
			// the event key and ordinal to keep their persisted IDs distinct.
			message.ID += "-" + job.eventKey + "-" + strconv.Itoa(index+1)
			message.Service, message.Sender, message.Recipient = "Meridian Global Correspondence Exchange", "Unknown sender", target.Name
			encoded, err := json.Marshal(message)
			if err != nil {
				return err
			}
			if _, err := tx.Exec(ctx, `INSERT INTO case_mail_messages(server_id,case_id,crime_id,message_id,mailbox_npc_entity_id,related_npc_entity_id,sent_at,internal_relation,message_data)
				VALUES($1::uuid,$2,$3,$4,$5,$6,$7,'anonymous_warning',$8::jsonb)`, serverID, job.caseID, crimeID, message.ID, target.ID, actor.ID, message.SentAt, string(encoded)); err != nil {
				return fmt.Errorf("persist serial killer threat: %w", err)
			}
		}
	} else if job.eventType != "serial_homicide" {
		return fmt.Errorf("unsupported serial evolution event %q", job.eventType)
	}
	content := "A set of anonymous threats was delivered to the next target. The messages refer to a workplace and repeat a physical description."
	presentation, evidenceType := "Anonymous threat correspondence", "letter_fragment"
	if job.eventType == "serial_homicide" {
		content = "A new victim was attacked after receiving repeated anonymous threats. The same workplace and physical description recur in the messages."
		presentation, evidenceType = "Serial attack scene report", "witness_statement"
		if _, err := tx.Exec(ctx, `INSERT INTO world_mutations(server_id,entity_id,field_key,value,created_at) VALUES($1::uuid,$2,'status','"missing"'::jsonb,$3)`, serverID, target.ID, job.scheduledAt); err != nil {
			return fmt.Errorf("mark serial victim missing: %w", err)
		}
	}
	evidence := EvidenceDefinition{ID: crimeID + "/evidence", Type: evidenceType, Presentation: presentation, Content: content, LocationID: job.locationID, RelationshipIDs: []string{}, ReachableFromEntityIDs: []string{target.ID, actor.ID}, SupportsNPCIDs: []string{actor.ID}}
	encodedEvidence, err := json.Marshal(struct {
		Presentation           string   `json:"presentation"`
		Content                string   `json:"content"`
		RelationshipIDs        []string `json:"relationshipIds"`
		ReachableFromEntityIDs []string `json:"reachableFromEntityIds"`
		SupportsNPCIDs         []string `json:"supportsNpcIds,omitempty"`
	}{evidence.Presentation, evidence.Content, evidence.RelationshipIDs, evidence.ReachableFromEntityIDs, evidence.SupportsNPCIDs})
	if err != nil {
		return err
	}
	if _, err := tx.Exec(ctx, `INSERT INTO case_evidence(server_id,case_id,crime_id,evidence_id,location_entity_id,evidence_type,evidence_data) VALUES($1::uuid,$2,$3,$4,NULLIF($5,''),$6,$7::jsonb)`, serverID, job.caseID, crimeID, evidence.ID, job.locationID, evidence.Type, string(encodedEvidence)); err != nil {
		return fmt.Errorf("persist serial event evidence: %w", err)
	}
	if job.eventType == "serial_threat" {
		var witnessKey string
		witnessErr := tx.QueryRow(ctx, `SELECT witness.generation_key FROM world_entities witness
			WHERE witness.server_id=$1::uuid AND witness.entity_type='npc' AND witness.entity_id<>$2 AND witness.entity_id<>$3
			AND (EXISTS (SELECT 1 FROM personal_relationships relation WHERE relation.server_id=witness.server_id AND relation.relationship_type='marriage'
				AND ((relation.person_a_npc_id=$2 AND relation.person_b_npc_id=witness.entity_id) OR (relation.person_b_npc_id=$2 AND relation.person_a_npc_id=witness.entity_id)))
			OR EXISTS (SELECT 1 FROM employment_relationships target_job JOIN employment_relationships witness_job
				ON witness_job.server_id=target_job.server_id AND witness_job.employer_company_id=target_job.employer_company_id
				WHERE target_job.server_id=witness.server_id AND target_job.employee_npc_id=$2 AND witness_job.employee_npc_id=witness.entity_id))
			ORDER BY CASE WHEN EXISTS (SELECT 1 FROM personal_relationships relation WHERE relation.server_id=witness.server_id AND relation.relationship_type='marriage'
				AND ((relation.person_a_npc_id=$2 AND relation.person_b_npc_id=witness.entity_id) OR (relation.person_b_npc_id=$2 AND relation.person_a_npc_id=witness.entity_id))) THEN 0 ELSE 1 END,witness.entity_id LIMIT 1`, serverID, target.ID, actor.ID).Scan(&witnessKey)
		if witnessErr != nil && !errors.Is(witnessErr, pgx.ErrNoRows) {
			return fmt.Errorf("find serial case correspondent: %w", witnessErr)
		}
		if witnessErr == nil {
			witness, err := world.GenerateNPCFromKey(job.seed, job.worldVersion, witnessKey)
			if err != nil {
				return fmt.Errorf("reconstruct serial case correspondent: %w", err)
			}
			messageID := "case-event:" + crimeID + ":help-request"
			body := fmt.Sprintf("%s has been receiving frightening anonymous messages. They mention a workplace and describe someone, but we cannot identify the sender. Could you ask around and compare the details with staff records before anyone else is hurt?", target.Name)
			// A case can evolve through several crimes, but it should not send
			// another unsolicited lead for every new event.
			if _, err := tx.Exec(ctx, `INSERT INTO player_inbox_messages(server_id,account_id,case_id,crime_id,message_id,sent_at,sender,message_category,subject,body,attachments)
				SELECT $1::uuid,account.id,$2,$3,$4,$5,$6,'official','Someone is receiving anonymous threats',$7,'[]'::jsonb
				FROM player_accounts account
				WHERE NOT EXISTS (
					SELECT 1 FROM player_inbox_messages prior
					WHERE prior.server_id=$1::uuid AND prior.account_id=account.id
					  AND prior.case_id=$2 AND prior.message_id LIKE 'case-event:%'
				)
				ON CONFLICT DO NOTHING`, serverID, job.caseID, crimeID, messageID, job.scheduledAt, witness.Name, body); err != nil {
				return fmt.Errorf("deliver serial case lead to player mailboxes: %w", err)
			}
		}
	}
	if _, err := tx.Exec(ctx, `UPDATE case_evolution_jobs SET status='completed',completed_at=$4 WHERE server_id=$1::uuid AND case_id=$2 AND event_key=$3`, serverID, job.caseID, job.eventKey, job.scheduledAt); err != nil {
		return err
	}
	if job.eventType == "serial_homicide" {
		if err := scheduleNextSerialVictim(ctx, tx, serverID, job); err != nil {
			return err
		}
	}
	return nil
}

func scheduleNextSerialVictim(ctx context.Context, tx pgx.Tx, serverID string, job dueEvolutionJob) error {
	sequence, err := strconv.Atoi(strings.TrimPrefix(job.eventKey, "serial-homicide-"))
	if err != nil || sequence < 1 {
		return fmt.Errorf("invalid serial evolution key %q", job.eventKey)
	}
	rows, err := tx.Query(ctx, `SELECT npc.entity_id FROM world_entities npc JOIN world_regions region ON region.server_id=npc.server_id AND region.region_id=npc.region_id
		WHERE npc.server_id=$1::uuid AND npc.entity_type='npc' AND region.region_id LIKE $2 || '/district/%'
		AND npc.entity_id<>$3 AND NOT EXISTS (SELECT 1 FROM case_participants participant WHERE participant.server_id=$1::uuid AND participant.case_id=$4 AND participant.npc_entity_id=npc.entity_id)
		AND COALESCE((SELECT mutation.value #>> '{}' FROM world_mutations mutation WHERE mutation.server_id=npc.server_id AND mutation.entity_id=npc.entity_id AND mutation.field_key='status' ORDER BY mutation.mutation_order DESC LIMIT 1),'alive')='alive'
		AND EXISTS (SELECT 1 FROM employment_relationships own_job JOIN employment_relationships colleague_job
			ON colleague_job.server_id=own_job.server_id AND colleague_job.employer_company_id=own_job.employer_company_id
			WHERE own_job.server_id=npc.server_id AND own_job.employee_npc_id=npc.entity_id AND colleague_job.employee_npc_id<>npc.entity_id AND colleague_job.employee_npc_id<>$3)
		ORDER BY npc.entity_id`, serverID, job.cityID, job.principalID, job.caseID)
	if err != nil {
		return fmt.Errorf("find next serial case target: %w", err)
	}
	targets := make([]string, 0)
	for rows.Next() {
		var targetID string
		if err := rows.Scan(&targetID); err != nil {
			rows.Close()
			return err
		}
		targets = append(targets, targetID)
	}
	if err := rows.Err(); err != nil {
		rows.Close()
		return err
	}
	rows.Close()
	if len(targets) == 0 {
		return nil
	}
	targetID := targets[keyIndex(job.seed, "serial-next-target/"+job.eventKey, len(targets))]
	nextSequence := sequence + 1
	threatAt := job.scheduledAt.Add(24 * time.Hour)
	attackAt := threatAt.Add(24 * time.Hour)
	if _, err := tx.Exec(ctx, `INSERT INTO case_evolution_jobs(server_id,case_id,event_key,event_type,target_npc_entity_id,available_at)
		VALUES($1::uuid,$2,$3,'serial_threat',$4,$5) ON CONFLICT DO NOTHING`, serverID, job.caseID, fmt.Sprintf("serial-threat-%d", nextSequence), targetID, threatAt); err != nil {
		return fmt.Errorf("schedule next serial threat: %w", err)
	}
	if _, err := tx.Exec(ctx, `INSERT INTO case_evolution_jobs(server_id,case_id,event_key,event_type,target_npc_entity_id,available_at)
		VALUES($1::uuid,$2,$3,'serial_homicide',$4,$5) ON CONFLICT DO NOTHING`, serverID, job.caseID, fmt.Sprintf("serial-homicide-%d", nextSequence), targetID, attackAt); err != nil {
		return fmt.Errorf("schedule next serial homicide: %w", err)
	}
	return nil
}

func caseSceneIsResidential(base world.World, locationID string) bool {
	for _, entity := range base.Entities {
		if entity.ID == locationID {
			return isResidentialGenerationKey(entity.GenerationKey)
		}
	}
	return false
}

func applyCorporateMarketImpact(ctx context.Context, tx pgx.Tx, serverID string, definition Definition, rivalIDs []string) error {
	delta := 0
	switch definition.ArchetypeID {
	case "workplace_conflict":
		delta = -25
	case "financial_extortion":
		delta = -75
	default:
		return nil
	}
	if definition.CompanyID == "" {
		return nil
	}
	occurredAt := definition.CrimeAt
	if occurredAt.IsZero() {
		occurredAt = time.Now().UTC()
	}
	return market.ApplyCompanyImpact(ctx, tx, serverID, definition.ID, definition.CompanyID, definition.ArchetypeID, delta, occurredAt.UTC(), rivalIDs)
}

func (store *PostgresStore) companyRivalIDs(ctx context.Context, serverID, companyID string) ([]string, error) {
	var seed string
	var version int
	err := store.db.QueryRow(ctx, `SELECT w.world_seed,COALESCE(city.generator_version,w.generator_version)
		FROM worlds w LEFT JOIN world_city_generators city ON city.server_id=w.server_id
		AND city.city_id=split_part(split_part($2,':',2),'/companies/',1)
		WHERE w.server_id=$1::uuid`, serverID, companyID).Scan(&seed, &version)
	if errors.Is(err, pgx.ErrNoRows) {
		return nil, world.ErrEntityNotFound
	}
	if err != nil {
		return nil, err
	}
	group, err := world.CompanyRivalGroupFor(seed, version, companyID)
	if err != nil {
		return nil, nil
	}
	rivalIDs := make([]string, 0, len(group.CompanyIDs))
	for _, rivalID := range group.CompanyIDs {
		if rivalID != companyID {
			rivalIDs = append(rivalIDs, rivalID)
		}
	}
	return rivalIDs, nil
}

func persistCaseEmails(ctx context.Context, tx pgx.Tx, serverID string, definition Definition) error {
	for _, message := range definition.Emails {
		data, err := json.Marshal(message)
		if err != nil {
			return fmt.Errorf("encode case email: %w", err)
		}
		if _, err := tx.Exec(ctx,
			`INSERT INTO case_mail_messages (server_id, case_id, crime_id, message_id, mailbox_npc_entity_id, related_npc_entity_id, sent_at, internal_relation, message_data)
			 VALUES ($1::uuid, $2, $3, $4, $5, $6, $7, $8, $9::jsonb)
			 ON CONFLICT (server_id, message_id) DO NOTHING`,
			serverID, definition.ID, definition.ID+"/crime/1", message.ID, message.MailboxNPCID, message.RelatedNPCID, message.SentAt, message.InternalRelation, string(data),
		); err != nil {
			return fmt.Errorf("persist case email %s: %w", message.ID, err)
		}
	}
	return nil
}

// NPCRecords loads only the selected person's case-related entries within the
// requested interval. It deliberately omits the internal relationship label.
func (store *PostgresStore) NPCRecords(ctx context.Context, serverID, npcID string, from, through time.Time, source string) ([]records.Record, error) {
	query := `SELECT activity.crime_id,crime.crime_type,activity.record_data FROM case_activity_records activity JOIN case_crimes crime USING(server_id,case_id,crime_id)
		 WHERE activity.server_id = $1::uuid AND activity.npc_entity_id = $2 AND activity.occurred_at >= $3 AND activity.occurred_at <= $4 AND activity.occurred_at <= now()
		 ORDER BY activity.occurred_at, activity.record_id`
	args := []any{serverID, npcID, from, through}
	if source != "" {
		query = `SELECT activity.crime_id,crime.crime_type,activity.record_data FROM case_activity_records activity JOIN case_crimes crime USING(server_id,case_id,crime_id)
			 WHERE activity.server_id = $1::uuid AND activity.npc_entity_id = $2 AND activity.source = $5 AND activity.occurred_at >= $3 AND activity.occurred_at <= $4 AND activity.occurred_at <= now()
			 ORDER BY activity.occurred_at, activity.record_id`
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
		if err := rows.Scan(&item.CrimeID, &item.CrimeType, &data); err != nil {
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
	if err := store.deliverNPCEmailRepliesForNPC(ctx, serverID, npcID); err != nil {
		return nil, fmt.Errorf("deliver NPC email replies: %w", err)
	}
	rows, err := store.db.Query(ctx,
		`SELECT message.case_id,message.crime_id,crime.crime_type,message.mailbox_npc_entity_id,message.related_npc_entity_id,message.internal_relation,message.message_data
		 FROM case_mail_messages message JOIN case_crimes crime USING(server_id,case_id,crime_id)
		 WHERE message.server_id = $1::uuid AND (message.mailbox_npc_entity_id = $2 OR message.related_npc_entity_id = $2)
		   AND message.sent_at >= $3 AND message.sent_at <= $4 AND message.sent_at <= now()
		 ORDER BY message.sent_at, message.message_id`, serverID, npcID, from, through,
	)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	result := make([]mail.Message, 0)
	for rows.Next() {
		var item mail.Message
		var data []byte
		var recipientMailboxID string
		if err := rows.Scan(&item.CaseID, &item.CrimeID, &item.CrimeType, &recipientMailboxID, &item.RelatedNPCID, &item.InternalRelation, &data); err != nil {
			return nil, err
		}
		if err := json.Unmarshal(data, &item); err != nil {
			return nil, fmt.Errorf("decode case email: %w", err)
		}
		item.MailboxNPCID = npcID
		if (item.InternalRelation == "anonymous_warning" || item.InternalRelation == "anonymous_followup") && recipientMailboxID == npcID {
			item.Sender = "Unknown sender"
		}
		result = append(result, item)
	}
	if err := rows.Err(); err != nil {
		return nil, err
	}
	questionRows, err := store.db.Query(ctx, `SELECT q.account_id::text,COALESCE(c.display_name,p.username),q.request_id,q.question_id,q.target_name,q.created_at
		FROM player_npc_email_questions q JOIN player_accounts p ON p.id=q.account_id
		LEFT JOIN player_characters c ON c.account_id=q.account_id
		WHERE q.server_id=$1::uuid AND q.target_npc_id=$2 AND q.created_at >= $3 AND q.created_at <= $4
		ORDER BY q.created_at,q.request_id`, serverID, npcID, from, through)
	if err != nil {
		return nil, err
	}
	defer questionRows.Close()
	for questionRows.Next() {
		var accountID, sender, requestID, questionID, recipient string
		var sentAt time.Time
		if err := questionRows.Scan(&accountID, &sender, &requestID, &questionID, &recipient, &sentAt); err != nil {
			return nil, err
		}
		question, ok := playerEmailQuestions[questionID]
		if !ok {
			continue
		}
		result = append(result, mail.Message{
			ID: "player-npc:" + accountID + ":" + requestID + ":question", SentAt: sentAt,
			Service: "Meridian Global Correspondence Exchange", Sender: sender, Recipient: recipient,
			Subject: question.prompt, Body: question.prompt, MailboxNPCID: npcID,
		})
	}
	if err := questionRows.Err(); err != nil {
		return nil, err
	}
	replyRows, err := store.db.Query(ctx, `SELECT q.account_id::text,COALESCE(c.display_name,p.username),q.request_id,q.target_name,q.subject,q.reply_body,q.ready_at
		FROM player_npc_email_questions q JOIN player_accounts p ON p.id=q.account_id
		LEFT JOIN player_characters c ON c.account_id=q.account_id
		WHERE q.server_id=$1::uuid AND q.target_npc_id=$2 AND q.reply_generated=true AND q.delivered_at IS NOT NULL
		AND q.ready_at >= $3 AND q.ready_at <= $4 ORDER BY q.ready_at,q.request_id`, serverID, npcID, from, through)
	if err != nil {
		return nil, err
	}
	defer replyRows.Close()
	for replyRows.Next() {
		var accountID, username, requestID, sender, subject, body string
		var sentAt time.Time
		if err := replyRows.Scan(&accountID, &username, &requestID, &sender, &subject, &body, &sentAt); err != nil {
			return nil, err
		}
		result = append(result, mail.Message{
			ID: "npc-player:" + accountID + ":" + requestID + ":reply", SentAt: sentAt,
			Service: "Meridian Global Correspondence Exchange", Sender: sender, Recipient: username,
			Subject: subject, Body: body, MailboxNPCID: npcID,
		})
	}
	if err := replyRows.Err(); err != nil {
		return nil, err
	}
	mail.SortMessages(result)
	return result, nil
}

// LatestPublicCase returns the newest active case without disclosing its killer.
func (store *PostgresStore) LatestPublicCase(ctx context.Context, serverID, cityID string) (PublicCase, error) {
	var result PublicCase
	err := store.db.QueryRow(ctx,
		`SELECT c.case_id, c.city_id, c.case_type, COALESCE(c.victim_entity_id,''), COALESCE(c.location_entity_id,''), s.status, c.created_at
		 FROM cases c JOIN case_states s USING (server_id, case_id)
		 WHERE c.server_id = $1::uuid AND c.city_id = $2 AND c.case_type='homicide' AND s.status='open' ORDER BY c.created_at DESC, c.case_id DESC LIMIT 1`, serverID, cityID,
	).Scan(&result.ID, &result.CityID, &result.CaseType, &result.VictimID, &result.LocationID, &result.Status, &result.CreatedAt)
	if errors.Is(err, pgx.ErrNoRows) {
		return PublicCase{}, ErrCaseNotFound
	}
	return result, err
}

func (store *PostgresStore) LatestPublicArrest(ctx context.Context, serverID, cityID string) (PublicCase, error) {
	var result PublicCase
	err := store.db.QueryRow(ctx, `SELECT c.case_id,c.city_id,c.case_type,COALESCE(c.victim_entity_id,''),COALESCE(c.location_entity_id,''),s.status,a.created_at
		FROM player_case_accusations a JOIN cases c USING(server_id,case_id) JOIN case_states s USING(server_id,case_id)
		WHERE a.server_id=$1::uuid AND c.city_id=$2 AND c.case_type='homicide' AND a.accepted=true ORDER BY a.created_at DESC,a.case_id DESC LIMIT 1`, serverID, cityID).
		Scan(&result.ID, &result.CityID, &result.CaseType, &result.VictimID, &result.LocationID, &result.Status, &result.CreatedAt)
	if errors.Is(err, pgx.ErrNoRows) {
		return PublicCase{}, ErrCaseNotFound
	}
	return result, err
}

// ListOpenPublicCases returns active cases without exposing their solutions.
func (store *PostgresStore) ListOpenPublicCases(ctx context.Context, serverID, cityID string) ([]PublicCase, error) {
	rows, err := store.db.Query(ctx,
		`SELECT c.case_id, c.city_id, c.case_type, COALESCE(c.victim_entity_id,''), COALESCE(c.location_entity_id,''), s.status, c.created_at
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

// CityCrimeReports provides the city overview without exposing case solutions.
func (store *PostgresStore) CityCrimeReports(ctx context.Context, serverID, cityID string, from, until, now time.Time) (CrimeReportDigest, error) {
	result := CrimeReportDigest{Cases: []PublicCase{}, Missing: []MissingPerson{}}
	rows, err := store.db.Query(ctx, `SELECT c.case_id,c.city_id,c.case_type,COALESCE(c.victim_entity_id,''),s.status,
		COALESCE(c.crime_at,c.created_at),c.created_at,COALESCE(v.generation_key,''),w.world_seed,
		COALESCE(organization.generation_key,''),
		COALESCE(city_generator.generator_version,w.generator_version)
		FROM cases c JOIN case_states s USING(server_id,case_id)
		LEFT JOIN world_entities v ON v.server_id=c.server_id AND v.entity_id=c.victim_entity_id
		LEFT JOIN world_entities organization ON organization.server_id=c.server_id AND organization.entity_id=c.company_id AND organization.entity_type='company'
		JOIN worlds w ON w.server_id=c.server_id
		LEFT JOIN world_city_generators city_generator ON city_generator.server_id=c.server_id AND city_generator.city_id=c.city_id
		WHERE c.server_id=$1::uuid AND c.city_id=$2 AND s.status='open'
		AND COALESCE(c.crime_at,c.created_at)>=$3 AND COALESCE(c.crime_at,c.created_at)<$4 AND COALESCE(c.crime_at,c.created_at)<=$5
		ORDER BY c.crime_at DESC,c.case_id DESC`, serverID, cityID, from.UTC(), until.UTC(), now.UTC())
	if err != nil {
		return CrimeReportDigest{}, err
	}
	for rows.Next() {
		var item PublicCase
		var generationKey, seed, organizationKey string
		var worldVersion int
		if err := rows.Scan(&item.ID, &item.CityID, &item.CaseType, &item.VictimID, &item.Status, &item.CrimeAt, &item.CreatedAt, &generationKey, &seed, &organizationKey, &worldVersion); err != nil {
			rows.Close()
			return CrimeReportDigest{}, err
		}
		if item.VictimID != "" && generationKey != "" {
			victim, err := world.GenerateNPCFromKey(seed, worldVersion, generationKey)
			if err != nil {
				rows.Close()
				return CrimeReportDigest{}, fmt.Errorf("reconstruct crime report victim: %w", err)
			}
			item.VictimName = victim.Name
			item.VictimVisual = victim.Visual
		}
		if organizationKey != "" {
			if organization, err := world.GenerateCompanyFromKey(seed, worldVersion, organizationKey); err == nil {
				item.OrganizationName = organization.Name
			}
		}
		result.Cases = append(result.Cases, item)
	}
	if err := rows.Err(); err != nil {
		rows.Close()
		return CrimeReportDigest{}, err
	}
	rows.Close()
	rows, err = store.db.Query(ctx, `SELECT DISTINCT ON (c.victim_entity_id) c.victim_entity_id,
		COALESCE(c.crime_at,c.created_at),v.generation_key,w.world_seed,
		COALESCE(city_generator.generator_version,w.generator_version)
		FROM cases c JOIN case_states s USING(server_id,case_id)
		JOIN world_entities v ON v.server_id=c.server_id AND v.entity_id=c.victim_entity_id
		JOIN worlds w ON w.server_id=c.server_id
		LEFT JOIN world_city_generators city_generator ON city_generator.server_id=c.server_id AND city_generator.city_id=c.city_id
		WHERE c.server_id=$1::uuid AND c.city_id=$2 AND c.victim_entity_id IS NOT NULL AND s.status IN ('open','solved')
		AND COALESCE(c.crime_at,c.created_at)<=$3 AND (s.body_found_at IS NULL OR s.body_found_at>$3)
		ORDER BY c.victim_entity_id,c.created_at DESC,c.case_id DESC`, serverID, cityID, now.UTC())
	if err != nil {
		return CrimeReportDigest{}, err
	}
	defer rows.Close()
	for rows.Next() {
		var person MissingPerson
		var generationKey, seed string
		var worldVersion int
		if err := rows.Scan(&person.NPCID, &person.MissingAt, &generationKey, &seed, &worldVersion); err != nil {
			return CrimeReportDigest{}, err
		}
		victim, err := world.GenerateNPCFromKey(seed, worldVersion, generationKey)
		if err != nil {
			return CrimeReportDigest{}, fmt.Errorf("reconstruct missing person: %w", err)
		}
		person.Name = victim.Name
		person.Visual = victim.Visual
		result.Missing = append(result.Missing, person)
	}
	return result, rows.Err()
}

// RevealDueBodies persists automatic public-scene discoveries once their scheduled time arrives.
func (store *PostgresStore) RevealDueBodies(ctx context.Context, serverID string, now time.Time) error {
	tx, err := store.db.Begin(ctx)
	if err != nil {
		return err
	}
	defer func() { _ = tx.Rollback(context.Background()) }()
	if _, err := tx.Exec(ctx, `INSERT INTO world_mutations(server_id,entity_id,field_key,value)
		SELECT c.server_id,c.victim_entity_id,'status','"missing"'::jsonb
		FROM cases c JOIN case_states s USING(server_id,case_id)
		WHERE c.server_id=$1::uuid AND c.victim_entity_id IS NOT NULL AND s.status IN ('open','solved') AND COALESCE(c.crime_at,c.created_at)<=$2
		AND (s.body_found_at IS NULL OR s.body_found_at>$2)
		AND COALESCE((SELECT m.value #>> '{}' FROM world_mutations m WHERE m.server_id=c.server_id AND m.entity_id=c.victim_entity_id AND m.field_key='status' ORDER BY m.mutation_order DESC LIMIT 1),'')<>'missing'`, serverID, now.UTC()); err != nil {
		return err
	}
	if _, err := tx.Exec(ctx, `INSERT INTO world_mutations(server_id,entity_id,field_key,value)
		SELECT c.server_id,c.victim_entity_id,'status','"dead"'::jsonb
		FROM cases c JOIN case_states s USING(server_id,case_id)
		WHERE c.server_id=$1::uuid AND c.victim_entity_id IS NOT NULL AND s.status IN ('open','solved') AND COALESCE(c.crime_at,c.created_at)<=$2
		AND s.body_found_at IS NOT NULL AND s.body_found_at<=$2
		AND COALESCE((SELECT m.value #>> '{}' FROM world_mutations m WHERE m.server_id=c.server_id AND m.entity_id=c.victim_entity_id AND m.field_key='status' ORDER BY m.mutation_order DESC LIMIT 1),'')<>'dead'`, serverID, now.UTC()); err != nil {
		return err
	}
	return tx.Commit(ctx)
}

func (store *PostgresStore) ListPublicBodyDiscoveries(ctx context.Context, serverID, cityID string, now time.Time) ([]PublicBodyDiscovery, error) {
	rows, err := store.db.Query(ctx, `SELECT c.case_id,s.status,v.generation_key,scene.generation_key,COALESCE(district.display_name,''),s.body_found_at,w.world_seed,
		COALESCE(city_generator.generator_version,w.generator_version)
		FROM cases c JOIN case_states s USING(server_id,case_id)
		JOIN world_entities v ON v.server_id=c.server_id AND v.entity_id=c.victim_entity_id
		JOIN world_entities scene ON scene.server_id=c.server_id AND scene.entity_id=c.location_entity_id
		JOIN world_regions district ON district.server_id=scene.server_id AND district.region_id=scene.region_id
		JOIN worlds w ON w.server_id=c.server_id
		LEFT JOIN world_city_generators city_generator ON city_generator.server_id=c.server_id AND city_generator.city_id=c.city_id
		WHERE c.server_id=$1::uuid AND c.city_id=$2 AND s.status='open' AND s.body_found_at<=$3
		ORDER BY s.body_found_at DESC,c.case_id DESC`, serverID, cityID, now.UTC())
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	result := make([]PublicBodyDiscovery, 0)
	var cityWorld world.World
	cityWorldLoaded := false
	for rows.Next() {
		var item PublicBodyDiscovery
		var victimKey, locationKey, seed string
		var worldVersion int
		if err := rows.Scan(&item.CaseID, &item.Status, &victimKey, &locationKey, &item.DistrictName, &item.FoundAt, &seed, &worldVersion); err != nil {
			return nil, err
		}
		victim, err := world.GenerateNPCFromKey(seed, worldVersion, victimKey)
		if err != nil {
			return nil, fmt.Errorf("reconstruct discovered victim: %w", err)
		}
		item.VictimName = victim.Name
		if !cityWorldLoaded {
			cityWorld, err = world.GenerateCity(seed, worldVersion, cityID)
			if err != nil {
				return nil, fmt.Errorf("reconstruct body discovery location: %w", err)
			}
			cityWorldLoaded = true
		}
		for _, entity := range cityWorld.Entities {
			if entity.GenerationKey == locationKey {
				item.LocationName = entity.Name
				break
			}
		}
		if item.LocationName == "" {
			item.LocationName = generatedNameFromKey(locationKey)
		}
		result = append(result, item)
	}
	return result, rows.Err()
}

func generatedNameFromKey(key string) string {
	parts := strings.Split(key, "/")
	name := strings.ReplaceAll(parts[len(parts)-1], "-", " ")
	if name == "" {
		return key
	}
	return strings.ToUpper(name[:1]) + name[1:]
}

// ListOpenPublicCasesForSession resolves the player's current city and lists its open cases in one database round-trip.
func (store *PostgresStore) ListOpenPublicCasesForSession(ctx context.Context, serverID, token string, now time.Time) ([]PublicCase, error) {
	tokenHash := sha256.Sum256([]byte(token))
	cityIDs := make([]string, 0, len(world.Cities()))
	for _, city := range world.Cities() {
		cityIDs = append(cityIDs, city.ID)
	}
	rows, err := store.db.Query(ctx, `
		SELECT resolved.city_id, open_cases.case_id, open_cases.city_id, open_cases.case_type,
		       open_cases.victim_entity_id, open_cases.location_entity_id, open_cases.status, open_cases.created_at
		FROM player_sessions s
		JOIN player_accounts a ON a.id = s.account_id
		LEFT JOIN player_locations pl ON pl.account_id = a.id
		CROSS JOIN LATERAL (
			SELECT COALESCE((
				SELECT city_id FROM unnest($4::text[]) AS city_catalog(city_id)
				WHERE COALESCE(pl.location_id, '') = city_id
				   OR COALESCE(pl.location_id, '') LIKE city_id || '/%'
				   OR COALESCE(pl.location_id, '') LIKE 'service:' || city_id || '/%'
				   OR COALESCE(pl.location_id, '') LIKE 'location:' || city_id || '/%'
				ORDER BY length(city_id) DESC LIMIT 1
			), CASE
				WHEN COALESCE(pl.location_id, '') IN ('lyon', 'service:lyon', 'location:lyon')
				  OR COALESCE(pl.location_id, '') LIKE 'lyon/%'
				  OR COALESCE(pl.location_id, '') LIKE 'service:lyon/%'
				  OR COALESCE(pl.location_id, '') LIKE 'location:lyon/%'
				THEN 'fr/lyon'
				ELSE $5
			END) AS city_id
		) resolved
		LEFT JOIN LATERAL (
			SELECT c.case_id, c.city_id, c.case_type, c.victim_entity_id, c.location_entity_id, s2.status, c.created_at
			FROM cases c JOIN case_states s2 USING (server_id, case_id)
			WHERE c.server_id = $1::uuid AND c.city_id = resolved.city_id AND s2.status = 'open'
			ORDER BY c.created_at DESC, c.case_id DESC
		) open_cases ON TRUE
		WHERE s.token_hash = $2 AND s.expires_at > $3
		ORDER BY open_cases.created_at DESC NULLS LAST, open_cases.case_id DESC NULLS LAST`,
		serverID, tokenHash[:], now, cityIDs, "fr/lyon")
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	results := make([]PublicCase, 0)
	sessionFound := false
	for rows.Next() {
		sessionFound = true
		var resolvedCity string
		var caseID, cityID, caseType, victimID, locationID, status pgtype.Text
		var createdAt pgtype.Timestamptz
		if err := rows.Scan(&resolvedCity, &caseID, &cityID, &caseType, &victimID, &locationID, &status, &createdAt); err != nil {
			return nil, err
		}
		if !caseID.Valid {
			continue
		}
		results = append(results, PublicCase{
			ID: caseID.String, CityID: cityID.String, CaseType: caseType.String, VictimID: victimID.String,
			LocationID: locationID.String, Status: status.String, CreatedAt: createdAt.Time,
		})
	}
	if err := rows.Err(); err != nil {
		return nil, err
	}
	if !sessionFound {
		return nil, ErrSessionNotFound
	}
	return results, nil
}

// GetPublicCase returns narrative evidence without internal classifications or solution data.
func (store *PostgresStore) GetPublicCase(ctx context.Context, serverID, caseID string) (PublicCase, []PublicEvidence, error) {
	var result PublicCase
	var generatorVersion int
	var worldSeed, organizationKey string
	err := store.db.QueryRow(ctx,
		`SELECT c.case_id, c.city_id, c.case_type, c.archetype_id, COALESCE(c.victim_entity_id,''), COALESCE(c.location_entity_id,''), s.status, c.created_at, c.generator_version,
		 w.world_seed, COALESCE(organization.generation_key,'')
		 FROM cases c JOIN case_states s USING (server_id, case_id) JOIN worlds w USING(server_id)
		 LEFT JOIN world_entities organization ON organization.server_id=c.server_id AND organization.entity_id=c.company_id AND organization.entity_type='company'
		 WHERE c.server_id = $1::uuid AND c.case_id = $2`, serverID, caseID,
	).Scan(&result.ID, &result.CityID, &result.CaseType, &result.ArchetypeID, &result.VictimID, &result.LocationID, &result.Status, &result.CreatedAt, &generatorVersion, &worldSeed, &organizationKey)
	if errors.Is(err, pgx.ErrNoRows) {
		return PublicCase{}, nil, ErrCaseNotFound
	}
	if err != nil {
		return PublicCase{}, nil, err
	}
	if organizationKey != "" && result.ArchetypeID != "serial_killer" {
		if organization, err := world.GenerateCompanyFromKey(worldSeed, generatorVersion, organizationKey); err == nil {
			result.OrganizationName = organization.Name
		}
	}
	rows, err := store.db.Query(ctx,
		`SELECT evidence.evidence_id,evidence.crime_id,crime.crime_type,COALESCE(evidence.location_entity_id,''),evidence.evidence_data
		 FROM case_evidence evidence JOIN case_crimes crime USING(server_id,case_id,crime_id)
		 WHERE evidence.server_id = $1::uuid AND evidence.case_id = $2 AND crime.occurred_at<=now()
		 ORDER BY evidence.evidence_id`, serverID, caseID,
	)
	if err != nil {
		return PublicCase{}, nil, err
	}
	defer rows.Close()
	evidence := make([]PublicEvidence, 0)
	for rows.Next() {
		var item PublicEvidence
		var data []byte
		if err := rows.Scan(&item.ID, &item.CrimeID, &item.CrimeType, &item.LocationID, &data); err != nil {
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
	// V4 records and correspondence are discovered through NPC Bank, GATE, and
	// Mail apps. Keep their associations out of the dossier so it cannot bypass
	// the investigation flow. Legacy v3 cases retain their existing presentation.
	if generatorVersion >= 4 {
		return result, evidence, nil
	}
	activityRows, err := store.db.Query(ctx,
		`SELECT record_id, crime_id, (SELECT crime_type FROM case_crimes crime WHERE crime.server_id=case_activity_records.server_id AND crime.case_id=case_activity_records.case_id AND crime.crime_id=case_activity_records.crime_id), record_data FROM case_activity_records
		 WHERE server_id = $1::uuid AND case_id = $2 AND occurred_at<=now() ORDER BY occurred_at, record_id`, serverID, caseID,
	)
	if err != nil {
		return PublicCase{}, nil, err
	}
	for activityRows.Next() {
		var id, crimeID, crimeType string
		var data []byte
		if err := activityRows.Scan(&id, &crimeID, &crimeType, &data); err != nil {
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
			if record.Direction != "" {
				content += " Direction: " + record.Direction + "."
			}
			if record.Reference != "" {
				content += " Account reference: " + record.Reference + "."
			}
			if record.Amount > 0 {
				content += fmt.Sprintf(" Amount: %d Credits.", record.Amount)
			}
		}
		if record.Origin != "" || record.Destination != "" {
			content += fmt.Sprintf(" Recorded passage: %s to %s.", record.Origin, record.Destination)
		}
		evidence = append(evidence, PublicEvidence{ID: id, CrimeID: crimeID, CrimeType: crimeType, Type: kind, Presentation: presentation, Content: content})
	}
	if err := activityRows.Err(); err != nil {
		activityRows.Close()
		return PublicCase{}, nil, err
	}
	activityRows.Close()
	mailRows, err := store.db.Query(ctx,
		`SELECT message_id, crime_id, (SELECT crime_type FROM case_crimes crime WHERE crime.server_id=case_mail_messages.server_id AND crime.case_id=case_mail_messages.case_id AND crime.crime_id=case_mail_messages.crime_id), message_data FROM case_mail_messages
		 WHERE server_id = $1::uuid AND case_id = $2 AND sent_at<=now() ORDER BY sent_at, message_id`, serverID, caseID,
	)
	if err != nil {
		return PublicCase{}, nil, err
	}
	for mailRows.Next() {
		var id, crimeID, crimeType string
		var data []byte
		if err := mailRows.Scan(&id, &crimeID, &crimeType, &data); err != nil {
			mailRows.Close()
			return PublicCase{}, nil, err
		}
		var message mail.Message
		if err := json.Unmarshal(data, &message); err != nil {
			mailRows.Close()
			return PublicCase{}, nil, fmt.Errorf("decode public case correspondence: %w", err)
		}
		evidence = append(evidence, PublicEvidence{ID: id, CrimeID: crimeID, CrimeType: crimeType, Type: "email", Presentation: "EMAIL · " + message.Subject, Content: message.Body})
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
	result, err := store.db.Exec(ctx, `UPDATE case_states SET status=$3,updated_at=now() WHERE server_id=$1::uuid AND case_id=$2`, serverID, caseID, status)
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
		 WHERE NOT EXISTS (SELECT 1 FROM player_blackwire_sales sale WHERE sale.server_id=$1::uuid AND sale.account_id=$2::uuid AND sale.item_id=item->>'id')
		 ON CONFLICT (server_id, account_id, item_id) DO UPDATE
		 SET item_source = EXCLUDED.item_source, item_data = EXCLUDED.item_data`,
		serverID, accountID, string(data)); err != nil {
		return fmt.Errorf("record discovered items: %w", err)
	}
	return nil
}

func (store *PostgresStore) ListWhiteboards(ctx context.Context, serverID, accountID string) ([]Whiteboard, error) {
	rows, err := store.db.Query(ctx,
		`SELECT whiteboard_id, title, created_at, locked_at FROM player_whiteboards
		 WHERE server_id = $1::uuid AND account_id = $2::uuid ORDER BY created_at, whiteboard_id`, serverID, accountID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	boards := make([]Whiteboard, 0)
	for rows.Next() {
		var board Whiteboard
		if err := rows.Scan(&board.ID, &board.Title, &board.CreatedAt, &board.LockedAt); err != nil {
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
		`SELECT board.title, board.created_at, board.locked_at,
		        COALESCE((
			          SELECT jsonb_agg(item.item_data || jsonb_build_object('x', pin.position_x, 'y', pin.position_y, 'designation', pin.designation)
		                           ORDER BY pin.pinned_at, pin.item_id)
		          FROM player_whiteboard_pins pin
		          JOIN player_discovered_items item USING (server_id, account_id, item_id)
			          LEFT JOIN player_blackwire_sales sale USING (server_id, account_id, item_id)
		          WHERE pin.server_id = board.server_id AND pin.account_id = board.account_id AND pin.whiteboard_id = board.whiteboard_id
			            AND sale.item_id IS NULL
		        ), '[]'::jsonb),
		        COALESCE((
		          SELECT jsonb_agg(jsonb_build_object('fromId', edge.from_item_id, 'toId', edge.to_item_id)
		                           ORDER BY edge.from_item_id, edge.to_item_id)
		          FROM player_whiteboard_connections edge
		          WHERE edge.server_id = board.server_id AND edge.account_id = board.account_id AND edge.whiteboard_id = board.whiteboard_id
		        ), '[]'::jsonb)
		 FROM player_whiteboards board
		 WHERE board.server_id = $1::uuid AND board.account_id = $2::uuid AND board.whiteboard_id = $3`,
		serverID, accountID, boardID).Scan(&board.Title, &board.CreatedAt, &board.LockedAt, &pinsJSON, &connectionsJSON)
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

func (store *PostgresStore) SetWhiteboardPinDesignation(ctx context.Context, serverID, accountID, boardID, itemID, designation string) error {
	if designation != "" && designation != "proof" && designation != "accused" && designation != "victim" {
		return ErrInvalidProgress
	}
	tx, err := store.db.Begin(ctx)
	if err != nil {
		return err
	}
	defer tx.Rollback(ctx)
	if designation == "accused" || designation == "victim" {
		var personPin bool
		err := tx.QueryRow(ctx, `SELECT EXISTS (SELECT 1 FROM player_whiteboard_pins p JOIN player_discovered_items i USING(server_id,account_id,item_id)
			WHERE p.server_id=$1::uuid AND p.account_id=$2::uuid AND p.whiteboard_id=$3 AND p.item_id=$4
			AND i.item_source='person' AND COALESCE(i.item_data #>> '{metadata,npc,id}','')<>'')`, serverID, accountID, boardID, itemID).Scan(&personPin)
		if err != nil {
			return err
		}
		if !personPin {
			return ErrInvalidProgress
		}
	}
	result, err := tx.Exec(ctx, `UPDATE player_whiteboard_pins SET designation=$5 WHERE server_id=$1::uuid AND account_id=$2::uuid AND whiteboard_id=$3 AND item_id=$4`, serverID, accountID, boardID, itemID, designation)
	if err != nil {
		return err
	}
	if result.RowsAffected() != 1 {
		return ErrEvidenceNotFound
	}
	return tx.Commit(ctx)
}

// SubmitAccusation snapshots a prepared board and queues its review.
func (store *PostgresStore) SubmitAccusation(ctx context.Context, serverID, accountID, cityID, boardID string, now time.Time) (AccusationResult, error) {
	tx, err := store.db.Begin(ctx)
	if err != nil {
		return AccusationResult{}, err
	}
	defer tx.Rollback(ctx)
	var lockedAt pgtype.Timestamptz
	if err := tx.QueryRow(ctx, `SELECT locked_at FROM player_whiteboards
		WHERE server_id=$1::uuid AND account_id=$2::uuid AND whiteboard_id=$3 FOR UPDATE`, serverID, accountID, boardID).Scan(&lockedAt); err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return AccusationResult{}, ErrCaseNotFound
		}
		return AccusationResult{}, err
	}
	var activeRequest asyncRequest
	err = tx.QueryRow(ctx, `SELECT request_id,request_type,payload,status,available_at FROM player_async_requests
		WHERE server_id=$1::uuid AND account_id=$2::uuid AND request_type='accusation'
		AND payload->>'boardId'=$3 AND status IN ('pending','processing')
		ORDER BY created_at DESC LIMIT 1`, serverID, accountID, boardID).
		Scan(&activeRequest.RequestID, &activeRequest.RequestType, &activeRequest.Payload, &activeRequest.Status, &activeRequest.AvailableAt)
	if err == nil {
		if !lockedAt.Valid {
			if _, err := tx.Exec(ctx, `UPDATE player_whiteboards SET locked_at=$4
				WHERE server_id=$1::uuid AND account_id=$2::uuid AND whiteboard_id=$3`, serverID, accountID, boardID, now.UTC()); err != nil {
				return AccusationResult{}, err
			}
		}
		if err := tx.Commit(ctx); err != nil {
			return AccusationResult{}, err
		}
		return AccusationResult{Status: activeRequest.Status, ReadyAt: activeRequest.AvailableAt, Message: "This board is already under review. The decision will arrive in the personal mailbox."}, nil
	}
	if !errors.Is(err, pgx.ErrNoRows) {
		return AccusationResult{}, err
	}
	if _, err := tx.Exec(ctx, `UPDATE player_whiteboards SET locked_at=$4
		WHERE server_id=$1::uuid AND account_id=$2::uuid AND whiteboard_id=$3`, serverID, accountID, boardID, now.UTC()); err != nil {
		return AccusationResult{}, err
	}
	var markedVictimID string
	var accusedIDs []string
	var proofIDs []string
	if err := tx.QueryRow(ctx, accusationBoardPinsQuery,
		serverID, accountID, boardID,
	).Scan(&accusedIDs, &markedVictimID, &proofIDs); err != nil {
		return AccusationResult{}, err
	}
	if len(accusedIDs) == 0 || len(proofIDs) == 0 {
		return AccusationResult{}, ErrInvalidProgress
	}
	var caseID string
	if markedVictimID != "" {
		err = tx.QueryRow(ctx, `SELECT c.case_id FROM cases c
			WHERE c.server_id=$1::uuid AND c.city_id=$2 AND (c.victim_entity_id=$3 OR EXISTS (
				SELECT 1 FROM case_crime_participants participant JOIN case_crimes crime USING(server_id,crime_id)
				WHERE participant.server_id=c.server_id AND crime.case_id=c.case_id AND participant.npc_entity_id=$3
				AND participant.participant_role='victim' AND crime.occurred_at<=$4))
			ORDER BY c.created_at DESC,c.case_id LIMIT 1`, serverID, cityID, markedVictimID, now.UTC()).Scan(&caseID)
	} else {
		err = tx.QueryRow(ctx, `SELECT c.case_id FROM cases c
			WHERE c.server_id=$1::uuid AND c.city_id=$2 AND c.case_type<>'homicide'
			AND (EXISTS (SELECT 1 FROM case_evidence evidence WHERE evidence.server_id=c.server_id AND evidence.case_id=c.case_id AND evidence.evidence_id=ANY($3::text[]))
				OR EXISTS (SELECT 1 FROM case_activity_records activity WHERE activity.server_id=c.server_id AND activity.case_id=c.case_id AND activity.record_id=ANY($3::text[]))
				OR EXISTS (SELECT 1 FROM case_mail_messages message WHERE message.server_id=c.server_id AND message.case_id=c.case_id AND message.message_id=ANY($3::text[])))
			ORDER BY c.created_at DESC,c.case_id LIMIT 1`, serverID, cityID, proofIDs).Scan(&caseID)
	}
	if err != nil && !errors.Is(err, pgx.ErrNoRows) {
		return AccusationResult{}, err
	}
	if errors.Is(err, pgx.ErrNoRows) {
		caseID = ""
	}
	var crimeIDs []string
	if caseID != "" {
		if err := tx.QueryRow(ctx, `SELECT COALESCE(array_agg(DISTINCT evidence.crime_id ORDER BY evidence.crime_id),'{}'::text[])
			FROM (
				SELECT server_id,case_id,crime_id FROM case_evidence WHERE server_id=$1::uuid AND case_id=$2 AND evidence_id=ANY($3::text[])
				UNION ALL SELECT server_id,case_id,crime_id FROM case_activity_records WHERE server_id=$1::uuid AND case_id=$2 AND record_id=ANY($3::text[])
				UNION ALL SELECT server_id,case_id,crime_id FROM case_mail_messages WHERE server_id=$1::uuid AND case_id=$2 AND message_id=ANY($3::text[])
			) evidence JOIN case_crimes crime USING(server_id,case_id,crime_id)
			WHERE crime.occurred_at<=$4`, serverID, caseID, proofIDs, now.UTC()).Scan(&crimeIDs); err != nil {
			return AccusationResult{}, fmt.Errorf("snapshot crimes represented on submitted board: %w", err)
		}
	}
	caseType := ""
	if caseID != "" {
		if err := tx.QueryRow(ctx, `SELECT case_type FROM cases WHERE server_id=$1::uuid AND case_id=$2`, serverID, caseID).Scan(&caseType); err != nil {
			return AccusationResult{}, err
		}
	}
	var submissionID [16]byte
	if _, err := cryptorand.Read(submissionID[:]); err != nil {
		return AccusationResult{}, fmt.Errorf("create accusation submission identity: %w", err)
	}
	requestKey := "case:" + caseID + ":board:" + boardID + ":review:" + hex.EncodeToString(submissionID[:])
	if caseID == "" {
		requestKey = "unmatched:" + boardID + ":" + hex.EncodeToString(submissionID[:])
	}
	payload := accusationPayload{CityID: cityID, BoardID: boardID, CaseID: caseID, CrimeIDs: crimeIDs, VictimID: markedVictimID, ProofIDs: proofIDs}
	if caseType == "homicide" && len(accusedIDs) == 1 {
		payload.AccusedID = accusedIDs[0]
	} else {
		payload.AccusedIDs = accusedIDs
	}
	payloadData, err := json.Marshal(payload)
	if err != nil {
		return AccusationResult{}, err
	}
	job, err := enqueueAsyncRequest(ctx, tx, serverID, accountID, "accusation", requestKey, payloadData, DefaultAccusationReviewDelay)
	if err != nil {
		return AccusationResult{}, err
	}
	if err := tx.Commit(ctx); err != nil {
		return AccusationResult{}, err
	}
	return AccusationResult{Status: job.Status, ReadyAt: job.AvailableAt, Message: "Your accusation has been received for review. Check the personal mailbox for the decision."}, nil
}

const accusationBoardPinsQuery = `SELECT COALESCE(array_agg(DISTINCT i.item_data #>> '{metadata,npc,id}' ORDER BY i.item_data #>> '{metadata,npc,id}') FILTER(WHERE p.designation='accused' AND i.item_source='person' AND COALESCE(i.item_data #>> '{metadata,npc,id}','')<>''),'{}'::text[]),
		COALESCE(max(i.item_data #>> '{metadata,npc,id}') FILTER(WHERE p.designation='victim' AND i.item_source='person'),''),
		COALESCE(array_agg(p.item_id ORDER BY p.item_id) FILTER(WHERE p.designation='proof' AND i.item_source <> 'note'),'{}'::text[])
		FROM player_whiteboard_pins p JOIN player_discovered_items i USING(server_id,account_id,item_id)
WHERE p.server_id=$1::uuid AND p.account_id=$2::uuid AND p.whiteboard_id=$3`

// resolveAccusation delivers the delayed review based on the immutable board snapshot.
func (store *PostgresStore) resolveAccusation(ctx context.Context, serverID, accountID, requestID string, request accusationPayload, now time.Time) (AccusationResult, error) {
	if err := store.validateAccusationCrimeSnapshot(ctx, serverID, request, now); err != nil {
		log.Printf("board_evaluation request_id=%q server_id=%q account_id=%q board_id=%q case_id=%q stage=validate_snapshot error=%q", requestID, serverID, accountID, request.BoardID, request.CaseID, err.Error())
		return AccusationResult{}, err
	}
	if len(request.AccusedIDs) > 0 {
		return store.resolveParticipantAccusation(ctx, serverID, accountID, requestID, request, now)
	}
	tx, err := store.db.Begin(ctx)
	if err != nil {
		return AccusationResult{}, err
	}
	defer tx.Rollback(ctx)
	unlockBoard := func() error {
		_, err := tx.Exec(ctx, `UPDATE player_whiteboards SET locked_at=NULL
			WHERE server_id=$1::uuid AND account_id=$2::uuid AND whiteboard_id=$3`, serverID, accountID, request.BoardID)
		return err
	}
	const genericFailure = "The corporate review has concluded. The available material did not meet the standard for an arrest. The decision does not identify which submissions were considered relevant."
	if request.CaseID == "" {
		log.Printf("board_evaluation request_id=%q server_id=%q account_id=%q board_id=%q stage=case_match result=no_matching_case accused_id=%q accused_ids=%q victim_id=%q proof_ids=%q", requestID, serverID, accountID, request.BoardID, request.AccusedID, request.AccusedIDs, request.VictimID, request.ProofIDs)
		messageID := "accusation:" + requestID
		if _, err := tx.Exec(ctx, `INSERT INTO player_inbox_messages(server_id,account_id,message_id,warrant_id,sent_at,subject,body,attachments)
			VALUES($1::uuid,$2::uuid,$3,NULL,$4,'Accusation review · Case decision',$5,'[]'::jsonb) ON CONFLICT DO NOTHING`, serverID, accountID, messageID, now.UTC(), genericFailure+billDispatchPostscript(requestID)); err != nil {
			return AccusationResult{}, err
		}
		if err := unlockBoard(); err != nil {
			return AccusationResult{}, err
		}
		if err := tx.Commit(ctx); err != nil {
			return AccusationResult{}, err
		}
		return AccusationResult{Status: "completed"}, nil
	}
	// Serialize verdicts for the shared case, including submissions from different players.
	if _, err := tx.Exec(ctx, `SELECT pg_advisory_xact_lock(hashtext($1),hashtext($2))`, serverID, request.CaseID); err != nil {
		return AccusationResult{}, err
	}
	var prior bool
	err = tx.QueryRow(ctx, `SELECT accepted FROM player_case_accusations WHERE server_id=$1::uuid AND account_id=$2::uuid AND case_id=$3 AND accused_npc_entity_id=$4`, serverID, accountID, request.CaseID, request.AccusedID).Scan(&prior)
	if err == nil && prior {
		result := AccusationResult{Accepted: prior, Status: "completed"}
		if _, err := tx.Exec(ctx, `INSERT INTO player_inbox_messages(server_id,account_id,message_id,warrant_id,sent_at,subject,body,attachments)
			VALUES($1::uuid,$2::uuid,$3,NULL,$4,'Accusation review · Case decision',$5,'[]'::jsonb) ON CONFLICT DO NOTHING`,
			serverID, accountID, "accusation:"+requestID, now.UTC(), "The corporate review confirms the prior arrest. The case remains resolved."+billDispatchPostscript(requestID)); err != nil {
			return AccusationResult{}, err
		}
		if err := unlockBoard(); err != nil {
			return AccusationResult{}, err
		}
		if err := tx.Commit(ctx); err != nil {
			return AccusationResult{}, err
		}
		return result, nil
	}
	if err != nil && !errors.Is(err, pgx.ErrNoRows) {
		return AccusationResult{}, err
	}
	var caseStatus string
	if err := tx.QueryRow(ctx, `SELECT status FROM case_states WHERE server_id=$1::uuid AND case_id=$2 FOR UPDATE`, serverID, request.CaseID).Scan(&caseStatus); err != nil {
		return AccusationResult{}, err
	}
	if caseStatus != "open" {
		log.Printf("board_evaluation request_id=%q server_id=%q account_id=%q board_id=%q case_id=%q stage=case_state case_status=%q result=case_not_open", requestID, serverID, accountID, request.BoardID, request.CaseID, caseStatus)
		if _, err := tx.Exec(ctx, `INSERT INTO player_case_accusations(server_id,account_id,case_id,accused_npc_entity_id,board_id,crime_ids,accepted,outcome,created_at)
			VALUES($1::uuid,$2::uuid,$3,$4,$5,$6,false,'insufficient',$7) ON CONFLICT(server_id,account_id,case_id,accused_npc_entity_id) DO UPDATE SET board_id=EXCLUDED.board_id,crime_ids=EXCLUDED.crime_ids,accepted=false,outcome='insufficient',created_at=EXCLUDED.created_at`, serverID, accountID, request.CaseID, request.AccusedID, request.BoardID, request.CrimeIDs, now.UTC()); err != nil {
			return AccusationResult{}, err
		}
		if _, err := tx.Exec(ctx, `INSERT INTO player_inbox_messages(server_id,account_id,message_id,warrant_id,sent_at,subject,body,attachments)
			VALUES($1::uuid,$2::uuid,$3,NULL,$4,'Accusation review · Case decision',$5,'[]'::jsonb) ON CONFLICT DO NOTHING`, serverID, accountID, "accusation:"+requestID, now.UTC(), genericFailure+billDispatchPostscript(requestID)); err != nil {
			return AccusationResult{}, err
		}
		if err := unlockBoard(); err != nil {
			return AccusationResult{}, err
		}
		if err := tx.Commit(ctx); err != nil {
			return AccusationResult{}, err
		}
		return AccusationResult{Status: "completed"}, nil
	}
	var killerID, victimID, caseType, archetypeID string
	var generatorVersion int
	if err := tx.QueryRow(ctx, `SELECT killer_entity_id,victim_entity_id,case_type,archetype_id,generator_version FROM cases WHERE server_id=$1::uuid AND case_id=$2`, serverID, request.CaseID).Scan(&killerID, &victimID, &caseType, &archetypeID, &generatorVersion); err != nil {
		return AccusationResult{}, err
	}
	var totalProofs int
	var proofStrength float64
	var proofBreakdown []byte
	if err := tx.QueryRow(ctx, `WITH proof_flags AS (
		SELECT proof.item_id,
			EXISTS (SELECT 1 FROM case_activity_records a JOIN cases c USING(server_id,case_id)
				WHERE a.server_id=$1::uuid AND a.case_id=$2 AND a.record_id=proof.item_id
				  AND ((a.source='bank' AND a.internal_relation='extortion_payment')
				    OR (a.source='gate' AND a.internal_relation IN ('scene-victim','scene-killer','scene-killer-exit'))
				    OR (c.generator_version<4 AND c.archetype_id='financial_extortion' AND a.internal_relation IN ('extortion_payment','precrime_route')))) AS extortion_or_scene_activity,
			EXISTS (SELECT 1 FROM cases c JOIN player_search_warrants w ON w.server_id=c.server_id AND w.city_id=c.city_id AND w.target_npc_id=c.killer_entity_id
				JOIN player_inbox_messages m ON m.server_id=w.server_id AND m.account_id=w.account_id AND m.warrant_id=w.warrant_id
				CROSS JOIN LATERAL jsonb_array_elements(m.attachments) AS attachment(item)
				WHERE c.server_id=$1::uuid AND c.case_id=$2 AND m.account_id=$3::uuid AND attachment.item->>'id'=proof.item_id
				  AND attachment.item->>'kind'='weapon') AS weapon_attachment,
			EXISTS (SELECT 1 FROM case_activity_records a WHERE a.server_id=$1::uuid AND a.case_id=$2 AND a.record_id=proof.item_id
				AND a.source='gate' AND a.internal_relation IN ('affair-victim','affair-paramour')) AS affair_gate_activity,
			EXISTS (SELECT 1 FROM case_evidence e JOIN cases c USING(server_id,case_id)
				WHERE e.server_id=$1::uuid AND e.case_id=$2 AND e.evidence_id=proof.item_id
				  AND ((c.archetype_id IN ('workplace_conflict','serial_killer') AND e.evidence_type='office_log')
				    OR (c.archetype_id='financial_extortion' AND e.evidence_type='travel_register'))) AS strong_case_evidence,
			EXISTS (SELECT 1 FROM case_evidence e JOIN cases c USING(server_id,case_id)
				WHERE e.server_id=$1::uuid AND e.case_id=$2 AND e.evidence_id=proof.item_id
				  AND ((c.archetype_id='workplace_conflict' AND e.evidence_type='meeting_minutes')
				    OR (c.archetype_id='financial_extortion' AND e.evidence_type='account_record')
				    OR (c.archetype_id IN ('jealousy','serial_killer') AND e.evidence_type IN ('letter_fragment','witness_statement')))) AS supporting_case_evidence,
			EXISTS (SELECT 1 FROM case_mail_messages m JOIN cases c USING(server_id,case_id)
				WHERE m.server_id=$1::uuid AND m.case_id=$2 AND m.message_id=proof.item_id
				  AND m.internal_relation IN ('anonymous_warning','anonymous_followup','killer_remorse','jealousy_exchange','case_correspondence')) AS case_correspondence,
			EXISTS (SELECT 1 FROM cases WHERE server_id=$1::uuid AND case_id=$2 AND generator_version>=4) AS current_generator
		FROM unnest($4::text[]) AS proof(item_id)
	), proof_evaluations AS (
		SELECT item_id,
			CASE WHEN extortion_or_scene_activity THEN 1.0
				WHEN weapon_attachment THEN 0.75
				WHEN affair_gate_activity THEN 0.75
				WHEN strong_case_evidence THEN 1.0
				WHEN supporting_case_evidence THEN 0.75
				WHEN case_correspondence AND current_generator THEN 0.75
				WHEN case_correspondence THEN 0.25
				ELSE 0.0 END AS strength,
			CASE WHEN extortion_or_scene_activity THEN 'case_activity_record'
				WHEN weapon_attachment THEN 'weapon_attachment'
				WHEN affair_gate_activity THEN 'affair_gate_activity'
				WHEN strong_case_evidence THEN 'strong_case_evidence'
				WHEN supporting_case_evidence THEN 'supporting_case_evidence'
				WHEN case_correspondence THEN 'case_correspondence'
				ELSE 'unmatched_or_context_only' END AS classification
		FROM proof_flags
	)
	SELECT count(*)::integer, COALESCE(sum(strength),0)::float8,
		COALESCE(jsonb_agg(jsonb_build_object('item_id',item_id,'classification',classification,'strength',strength) ORDER BY item_id),'[]'::jsonb)
	FROM proof_evaluations`, serverID, request.CaseID, accountID, request.ProofIDs).Scan(&totalProofs, &proofStrength, &proofBreakdown); err != nil {
		return AccusationResult{}, err
	}
	// Authenticated case evidence has different probative strength. Context-only pins
	// do not raise arrest odds; unrelated pins marked as proof lower them.
	victimMatches := request.VictimID != "" && request.VictimID == victimID
	if !victimMatches && request.VictimID != "" {
		if err := tx.QueryRow(ctx, `SELECT EXISTS(
			SELECT 1 FROM case_crime_participants participant JOIN case_crimes crime USING(server_id,crime_id)
			WHERE participant.server_id=$1::uuid AND crime.case_id=$2 AND participant.npc_entity_id=$3 AND participant.participant_role='victim'
			AND crime.occurred_at<=$4 AND (cardinality($5::text[])=0 OR crime.crime_id=ANY($5::text[])))`,
			serverID, request.CaseID, request.VictimID, now.UTC(), request.CrimeIDs).Scan(&victimMatches); err != nil {
			return AccusationResult{}, err
		}
	}
	correctAccused := request.AccusedID == killerID && victimMatches
	chance := accusationChance(correctAccused, totalProofs, proofStrength)
	// The deterministic review roll is useful while the submitted material is
	// inconclusive, but a board with authenticated financial and scene records
	// plus a related case message has met the corroboration threshold. This
	// combination is worth 2.75 points (1 + 1 + 0.75) and still requires the
	// correct accused person and victim.
	corroborated := correctAccused && proofStrength >= 2.75
	accepted := corroborated || accusationAccepted(serverID, accountID, request.CaseID, chance)
	log.Printf("board_evaluation request_id=%q server_id=%q account_id=%q board_id=%q case_id=%q case_status=%q case_type=%q archetype_id=%q generator_version=%d accused_id=%q killer_id=%q accused_matches_killer=%t victim_id=%q expected_victim_id=%q victim_matches=%t crime_ids=%q proof_count=%d proof_strength=%.2f chance=%d corroborated=%t accepted=%t proof_breakdown=%s", requestID, serverID, accountID, request.BoardID, request.CaseID, caseStatus, caseType, archetypeID, generatorVersion, request.AccusedID, killerID, request.AccusedID == killerID, request.VictimID, victimID, victimMatches, request.CrimeIDs, totalProofs, proofStrength, chance, corroborated, accepted, proofBreakdown)
	outcome, subject, body := "insufficient", "Accusation review · Case decision", "The corporate review has concluded. The available material did not meet the standard for an arrest. The decision does not identify which submissions were considered relevant."
	if accepted {
		outcome = "arrested"
		body = "The corporate review has authorized the arrest. The accused has been taken into custody, and the city police have been notified."
	}
	body += billDispatchPostscript(requestID)
	if _, err := tx.Exec(ctx, `INSERT INTO player_case_accusations(server_id,account_id,case_id,accused_npc_entity_id,board_id,crime_ids,accepted,outcome,created_at) VALUES($1::uuid,$2::uuid,$3,$4,$5,$6,$7,$8,$9)
		ON CONFLICT(server_id,account_id,case_id,accused_npc_entity_id) DO UPDATE SET board_id=EXCLUDED.board_id,crime_ids=EXCLUDED.crime_ids,accepted=EXCLUDED.accepted,outcome=EXCLUDED.outcome,created_at=EXCLUDED.created_at`, serverID, accountID, request.CaseID, request.AccusedID, request.BoardID, request.CrimeIDs, accepted, outcome, now.UTC()); err != nil {
		return AccusationResult{}, err
	}
	if accepted {
		statusJSON, _ := json.Marshal("arrested")
		if _, err := tx.Exec(ctx, `INSERT INTO world_mutations(server_id,entity_id,field_key,value) VALUES($1::uuid,$2,'status',$3::jsonb)`, serverID, request.AccusedID, string(statusJSON)); err != nil {
			return AccusationResult{}, err
		}
		if _, err := tx.Exec(ctx, `UPDATE case_states SET status='solved',updated_at=$3 WHERE server_id=$1::uuid AND case_id=$2`, serverID, request.CaseID, now.UTC()); err != nil {
			return AccusationResult{}, err
		}
	}
	if _, err := tx.Exec(ctx, `INSERT INTO player_inbox_messages(server_id,account_id,message_id,warrant_id,sent_at,subject,body,attachments)
		VALUES($1::uuid,$2::uuid,$3,NULL,$4,$5,$6,'[]'::jsonb) ON CONFLICT DO NOTHING`, serverID, accountID, "accusation:"+requestID, now.UTC(), subject, body); err != nil {
		return AccusationResult{}, err
	}
	if err := unlockBoard(); err != nil {
		return AccusationResult{}, err
	}
	if err := tx.Commit(ctx); err != nil {
		return AccusationResult{}, err
	}
	return AccusationResult{Accepted: accepted, Status: "completed"}, nil
}

func (store *PostgresStore) validateAccusationCrimeSnapshot(ctx context.Context, serverID string, request accusationPayload, now time.Time) error {
	if request.CaseID == "" || len(request.CrimeIDs) == 0 {
		return nil
	}
	var represented int
	err := store.db.QueryRow(ctx, `SELECT count(DISTINCT evidence.crime_id)::integer
		FROM (
			SELECT server_id,case_id,crime_id,evidence_id AS item_id FROM case_evidence WHERE server_id=$1::uuid AND case_id=$2
			UNION ALL SELECT server_id,case_id,crime_id,record_id FROM case_activity_records WHERE server_id=$1::uuid AND case_id=$2
			UNION ALL SELECT server_id,case_id,crime_id,message_id FROM case_mail_messages WHERE server_id=$1::uuid AND case_id=$2
		) evidence JOIN case_crimes crime USING(server_id,case_id,crime_id)
		WHERE evidence.item_id=ANY($3::text[]) AND crime.crime_id=ANY($4::text[]) AND crime.occurred_at<=$5`,
		serverID, request.CaseID, request.ProofIDs, request.CrimeIDs, now.UTC()).Scan(&represented)
	if err != nil {
		return err
	}
	if represented != len(request.CrimeIDs) {
		return ErrInvalidProgress
	}
	return nil
}

func (store *PostgresStore) resolveParticipantAccusation(ctx context.Context, serverID, accountID, requestID string, request accusationPayload, now time.Time) (AccusationResult, error) {
	tx, err := store.db.Begin(ctx)
	if err != nil {
		return AccusationResult{}, err
	}
	defer tx.Rollback(ctx)
	unlockBoard := func() error {
		_, err := tx.Exec(ctx, `UPDATE player_whiteboards SET locked_at=NULL WHERE server_id=$1::uuid AND account_id=$2::uuid AND whiteboard_id=$3`, serverID, accountID, request.BoardID)
		return err
	}
	complete := func(body string, accepted bool) (AccusationResult, error) {
		if _, err := tx.Exec(ctx, `INSERT INTO player_inbox_messages(server_id,account_id,message_id,warrant_id,sent_at,subject,body,attachments)
			VALUES($1::uuid,$2::uuid,$3,NULL,$4,'Accusation review · Case decision',$5,'[]'::jsonb) ON CONFLICT DO NOTHING`, serverID, accountID, "accusation:"+requestID, now.UTC(), body+billDispatchPostscript(requestID)); err != nil {
			return AccusationResult{}, err
		}
		if err := unlockBoard(); err != nil {
			return AccusationResult{}, err
		}
		if err := tx.Commit(ctx); err != nil {
			return AccusationResult{}, err
		}
		return AccusationResult{Accepted: accepted, Status: "completed"}, nil
	}
	if request.CaseID == "" {
		return complete("The corporate review has concluded. The available material did not meet the standard for an arrest. The decision does not identify which submissions were considered relevant.", false)
	}
	if _, err := tx.Exec(ctx, `SELECT pg_advisory_xact_lock(hashtext($1),hashtext($2))`, serverID, request.CaseID); err != nil {
		return AccusationResult{}, err
	}
	var caseType, caseStatus, seed, initialVictimID string
	var worldVersion int
	if err := tx.QueryRow(ctx, `SELECT c.case_type,s.status,w.world_seed,COALESCE(city_generator.generator_version,w.generator_version),COALESCE(c.victim_entity_id,'')
		FROM cases c JOIN case_states s USING(server_id,case_id) JOIN worlds w USING(server_id)
		LEFT JOIN world_city_generators city_generator ON city_generator.server_id=c.server_id AND city_generator.city_id=c.city_id
		WHERE c.server_id=$1::uuid AND c.case_id=$2`, serverID, request.CaseID).Scan(&caseType, &caseStatus, &seed, &worldVersion, &initialVictimID); err != nil {
		return AccusationResult{}, err
	}
	if caseStatus != "open" {
		log.Printf("board_evaluation request_id=%q server_id=%q account_id=%q board_id=%q case_id=%q case_status=%q route=participant result=case_not_open", requestID, serverID, accountID, request.BoardID, request.CaseID, caseStatus)
		return complete("The case file is no longer open for further accusations. No additional arrest was made.", false)
	}
	uniqueIDs := make([]string, 0, len(request.AccusedIDs))
	seen := make(map[string]struct{}, len(request.AccusedIDs))
	for _, npcID := range request.AccusedIDs {
		if npcID == "" {
			continue
		}
		if _, exists := seen[npcID]; exists {
			continue
		}
		seen[npcID] = struct{}{}
		uniqueIDs = append(uniqueIDs, npcID)
	}
	if len(uniqueIDs) == 0 || len(request.ProofIDs) == 0 {
		log.Printf("board_evaluation request_id=%q server_id=%q account_id=%q board_id=%q case_id=%q route=participant result=missing_accused_or_proof accused_ids=%q proof_ids=%q", requestID, serverID, accountID, request.BoardID, request.CaseID, uniqueIDs, request.ProofIDs)
		return complete("The corporate review has concluded. The available material did not meet the standard for an arrest.", false)
	}
	var principalID string
	if err := tx.QueryRow(ctx, `SELECT npc_entity_id FROM case_participants WHERE server_id=$1::uuid AND case_id=$2 AND participant_role='principal'`, serverID, request.CaseID).Scan(&principalID); err != nil {
		return AccusationResult{}, err
	}
	victimMatches := request.VictimID != "" && request.VictimID == initialVictimID
	if !victimMatches && request.VictimID != "" && caseType == "homicide" {
		if err := tx.QueryRow(ctx, `SELECT EXISTS(
			SELECT 1 FROM case_crime_participants participant JOIN case_crimes crime USING(server_id,crime_id)
			WHERE participant.server_id=$1::uuid AND crime.case_id=$2 AND participant.npc_entity_id=$3 AND participant.participant_role='victim'
			AND crime.occurred_at<=$4 AND (cardinality($5::text[])=0 OR crime.crime_id=ANY($5::text[])))`,
			serverID, request.CaseID, request.VictimID, now.UTC(), request.CrimeIDs).Scan(&victimMatches); err != nil {
			return AccusationResult{}, err
		}
	}
	lines := make([]string, 0, len(uniqueIDs))
	anyArrest := false
	for _, npcID := range uniqueIDs {
		var participantRole, generationKey string
		err := tx.QueryRow(ctx, `SELECT p.participant_role,e.generation_key FROM case_participants p JOIN world_entities e ON e.server_id=p.server_id AND e.entity_id=p.npc_entity_id
			WHERE p.server_id=$1::uuid AND p.case_id=$2 AND p.npc_entity_id=$3 AND p.participant_role IN ('principal','accomplice')`, serverID, request.CaseID, npcID).Scan(&participantRole, &generationKey)
		if errors.Is(err, pgx.ErrNoRows) {
			participantRole = ""
		} else if err != nil {
			return AccusationResult{}, err
		}
		var relevantProofs int
		var proofStrength float64
		if caseType == "homicide" {
			if npcID == principalID {
				err = tx.QueryRow(ctx, `SELECT count(*)::integer, COALESCE(sum(CASE WHEN evidence.evidence_id IS NOT NULL AND (COALESCE(jsonb_array_length(evidence.evidence_data->'supportsNpcIds'),0)=0 OR evidence.evidence_data->'supportsNpcIds' ? $4) THEN 1.0 ELSE 0.0 END),0)::float8
					FROM unnest($3::text[]) proof(id) LEFT JOIN case_evidence evidence ON evidence.server_id=$1::uuid AND evidence.case_id=$2 AND evidence.evidence_id=proof.id`, serverID, request.CaseID, request.ProofIDs, npcID).Scan(&relevantProofs, &proofStrength)
				if err != nil {
					return AccusationResult{}, err
				}
			} else {
				relevantProofs = len(request.ProofIDs)
			}
		} else {
			err = tx.QueryRow(ctx, `SELECT count(*) FILTER (WHERE
				EXISTS (SELECT 1 FROM case_evidence evidence WHERE evidence.server_id=$1::uuid AND evidence.case_id=$2 AND evidence.evidence_id=proof.id AND evidence.evidence_data->'supportsNpcIds' ? $4)
				OR EXISTS (SELECT 1 FROM case_activity_records activity WHERE activity.server_id=$1::uuid AND activity.case_id=$2 AND activity.record_id=proof.id AND activity.npc_entity_id=$4))::integer,
				COALESCE(sum(CASE WHEN EXISTS (SELECT 1 FROM case_evidence evidence WHERE evidence.server_id=$1::uuid AND evidence.case_id=$2 AND evidence.evidence_id=proof.id AND evidence.evidence_data->'supportsNpcIds' ? $4)
					OR EXISTS (SELECT 1 FROM case_activity_records activity WHERE activity.server_id=$1::uuid AND activity.case_id=$2 AND activity.record_id=proof.id AND activity.npc_entity_id=$4) THEN 1.0 ELSE 0.0 END),0)::float8
				FROM unnest($3::text[]) proof(id)`, serverID, request.CaseID, request.ProofIDs, npcID).Scan(&relevantProofs, &proofStrength)
			if err != nil {
				return AccusationResult{}, err
			}
		}
		correctParticipant := participantRole == "principal" || participantRole == "accomplice"
		if caseType == "homicide" {
			correctParticipant = npcID == principalID && participantRole == "principal" && victimMatches
		}
		var alreadyArrested bool
		if err := tx.QueryRow(ctx, `SELECT COALESCE((SELECT value #>> '{}' FROM world_mutations WHERE server_id=$1::uuid AND entity_id=$2 AND field_key='status' ORDER BY mutation_order DESC LIMIT 1),'')='arrested'`, serverID, npcID).Scan(&alreadyArrested); err != nil {
			return AccusationResult{}, err
		}
		chance := accusationChance(correctParticipant, relevantProofs, proofStrength)
		// Unlike the legacy homicide review, financial cases require at least one
		// submitted item that specifically supports this participant. The base
		// accusation chance must not let a guess or another person's evidence
		// authorize an arrest.
		accepted := participantAccusationEligible(correctParticipant, alreadyArrested, relevantProofs) &&
			(alreadyArrested || accusationAcceptedForParticipant(serverID, accountID, request.CaseID, npcID, chance))
		log.Printf("board_evaluation request_id=%q server_id=%q account_id=%q board_id=%q case_id=%q route=participant case_type=%q accused_id=%q principal_id=%q participant_role=%q correct_participant=%t victim_id=%q initial_victim_id=%q victim_matches=%t proof_ids=%q relevant_proofs=%d proof_strength=%.2f chance=%d already_arrested=%t accepted=%t", requestID, serverID, accountID, request.BoardID, request.CaseID, caseType, npcID, principalID, participantRole, correctParticipant, request.VictimID, initialVictimID, victimMatches, request.ProofIDs, relevantProofs, proofStrength, chance, alreadyArrested, accepted)
		outcome := "insufficient"
		if accepted {
			outcome = "arrested"
			anyArrest = true
			if !alreadyArrested {
				statusJSON, _ := json.Marshal("arrested")
				if _, err := tx.Exec(ctx, `INSERT INTO world_mutations(server_id,entity_id,field_key,value) VALUES($1::uuid,$2,'status',$3::jsonb)`, serverID, npcID, string(statusJSON)); err != nil {
					return AccusationResult{}, err
				}
			}
		}
		if _, err := tx.Exec(ctx, `INSERT INTO player_case_accusations(server_id,account_id,case_id,accused_npc_entity_id,board_id,crime_ids,accepted,outcome,created_at)
			VALUES($1::uuid,$2::uuid,$3,$4,$5,$6,$7,$8,$9) ON CONFLICT(server_id,account_id,case_id,accused_npc_entity_id)
			DO UPDATE SET board_id=EXCLUDED.board_id,crime_ids=EXCLUDED.crime_ids,accepted=EXCLUDED.accepted,outcome=EXCLUDED.outcome,created_at=EXCLUDED.created_at`, serverID, accountID, request.CaseID, npcID, request.BoardID, request.CrimeIDs, accepted, outcome, now.UTC()); err != nil {
			return AccusationResult{}, err
		}
		name := "the selected person"
		if npc, generateErr := world.GenerateNPCFromKey(seed, worldVersion, generationKey); generateErr == nil && npc.Name != "" {
			name = npc.Name
		}
		if accepted {
			if alreadyArrested {
				lines = append(lines, fmt.Sprintf("%s: already in custody.", name))
			} else {
				lines = append(lines, fmt.Sprintf("%s: arrest authorized.", name))
			}
		} else {
			lines = append(lines, fmt.Sprintf("%s: the evidence did not meet the standard for an arrest.", name))
		}
	}
	if _, err := tx.Exec(ctx, `UPDATE case_states SET status='solved',updated_at=$3 WHERE server_id=$1::uuid AND case_id=$2
		AND NOT EXISTS (SELECT 1 FROM case_participants p WHERE p.server_id=$1::uuid AND p.case_id=$2 AND p.participant_role IN ('principal','accomplice')
		AND COALESCE((SELECT value #>> '{}' FROM world_mutations m WHERE m.server_id=$1::uuid AND m.entity_id=p.npc_entity_id AND m.field_key='status' ORDER BY mutation_order DESC LIMIT 1),'')<>'arrested')`, serverID, request.CaseID, now.UTC()); err != nil {
		return AccusationResult{}, err
	}
	body := "The corporate review has concluded. Each named person was assessed separately.\n\n" + strings.Join(lines, "\n")
	return complete(body, anyArrest)
}

func accusationChance(correctSuspect bool, totalProofs int, proofStrength float64) int {
	if !correctSuspect {
		return 0
	}
	chance := int(math.Round(10 + proofStrength*65 - (float64(totalProofs)-proofStrength)*40))
	if chance < 0 {
		return 0
	}
	if chance > 99 {
		return 99
	}
	return chance
}

func accusationAccepted(serverID, accountID, caseID string, chance int) bool {
	roll := sha256.Sum256([]byte(fmt.Sprintf("judgment-v1|%s|%s|%s", serverID, accountID, caseID)))
	return int(binary.BigEndian.Uint64(roll[:8])%100) < chance
}

func accusationAcceptedForParticipant(serverID, accountID, caseID, npcID string, chance int) bool {
	roll := sha256.Sum256([]byte(fmt.Sprintf("judgment-participant-v1|%s|%s|%s|%s", serverID, accountID, caseID, npcID)))
	return int(binary.BigEndian.Uint64(roll[:8])%100) < chance
}

func participantAccusationEligible(correctParticipant, alreadyArrested bool, relevantProofs int) bool {
	return correctParticipant && (alreadyArrested || relevantProofs > 0)
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
		                    AND NOT EXISTS (SELECT 1 FROM player_blackwire_sales sale WHERE sale.server_id=$1::uuid AND sale.account_id=$2::uuid AND sale.item_id=player_discovered_items.item_id)
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

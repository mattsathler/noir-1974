package cases

import (
	"context"
	"crypto/sha256"
	"encoding/binary"
	"encoding/json"
	"fmt"
	"sort"
	"strings"
	"time"

	"github.com/jackc/pgx/v5"
	"noir-1974/apps/api/internal/world"
)

type HospitalEvidence struct {
	ID      string `json:"id"`
	Source  string `json:"source"`
	Title   string `json:"title"`
	Content string `json:"content"`
}

type PostMortemRecord struct {
	ReportID         string             `json:"reportId"`
	EstimatedDeathAt time.Time          `json:"estimatedDeathAt"`
	Cause            string             `json:"cause"`
	Weapon           string             `json:"weapon,omitempty"`
	BodyFoundAt      string             `json:"bodyFoundAt"`
	Evidence         []HospitalEvidence `json:"evidence"`
}

type HospitalPatient struct {
	ID              string                `json:"id"`
	Name            string                `json:"name"`
	Visual          *world.VisualIdentity `json:"visual,omitempty"`
	Status          string                `json:"status"`
	FingerprintCode string                `json:"fingerprintCode"`
	PostMortem      *PostMortemRecord     `json:"postMortem,omitempty"`
}

const hospitalRegionNameQuery = `SELECT display_name FROM world_regions WHERE server_id=$1::uuid AND region_id=$2`

// SearchHospitalPatients reconstructs city residents and case files from stable world data.
// The handler validates the account's current local access attempt before calling this method.
func (store *PostgresStore) SearchHospitalPatients(ctx context.Context, serverID, cityID, query string) ([]HospitalPatient, []DiscoveredItem, error) {
	city, ok := world.CityByID(cityID)
	if !ok {
		return nil, nil, ErrCaseNotFound
	}
	seed, worldVersion, err := store.worldSeedAndVersion(ctx, serverID)
	if err != nil {
		return nil, nil, err
	}
	cityVersion := worldVersion
	if cityID != "fr/lyon" {
		err := store.db.QueryRow(ctx, `SELECT generator_version FROM world_city_generators WHERE server_id=$1::uuid AND city_id=$2`, serverID, cityID).Scan(&cityVersion)
		if err == pgx.ErrNoRows && cityVersion < world.LatestGeneratorVersion {
			cityVersion = world.LatestGeneratorVersion
		} else if err != nil && err != pgx.ErrNoRows {
			return nil, nil, err
		}
	}
	generated, err := world.GenerateCity(seed, cityVersion, cityID)
	if err != nil {
		return nil, nil, err
	}
	ids := make([]string, 0, len(generated.NPCs))
	npcs := make(map[string]world.NPC, len(generated.NPCs))
	for _, npc := range generated.NPCs {
		ids = append(ids, npc.ID)
		npcs[npc.ID] = npc
	}
	// Include deterministic residents added after the base city generation.
	cityKey := cityID
	if cityID == "fr/lyon" {
		cityKey = "lyon"
	}
	rows, err := store.db.Query(ctx, `SELECT e.entity_id,e.generation_key FROM world_entities e JOIN world_regions r USING(server_id,region_id)
		WHERE e.server_id=$1::uuid AND e.entity_type='npc' AND r.region_id LIKE $2 ORDER BY e.entity_id`, serverID, cityKey+"/district/%")
	if err != nil {
		return nil, nil, err
	}
	for rows.Next() {
		var id, generationKey string
		if err := rows.Scan(&id, &generationKey); err != nil {
			rows.Close()
			return nil, nil, err
		}
		if _, exists := npcs[id]; exists {
			continue
		}
		npc, err := world.GenerateNPCFromKey(seed, cityVersion, generationKey)
		if err != nil {
			rows.Close()
			return nil, nil, err
		}
		npcs[id] = npc
		ids = append(ids, id)
	}
	if err := rows.Err(); err != nil {
		rows.Close()
		return nil, nil, err
	}
	rows.Close()
	codes := fingerprintCodes(ids)
	needle := strings.ToLower(strings.TrimSpace(query))
	if needle == "" {
		return []HospitalPatient{}, []DiscoveredItem{}, nil
	}
	patients := make([]HospitalPatient, 0)
	discovered := make([]DiscoveredItem, 0)
	npcIDs := make([]string, 0, len(npcs))
	for id := range npcs {
		npcIDs = append(npcIDs, id)
	}
	sort.Strings(npcIDs)
	for _, id := range npcIDs {
		npc := npcs[id]
		if !strings.Contains(strings.ToLower(npc.Name), needle) {
			continue
		}
		var status string
		err := store.db.QueryRow(ctx, `SELECT value #>> '{}' FROM world_mutations WHERE server_id=$1::uuid AND entity_id=$2 AND field_key='status' ORDER BY mutation_order DESC LIMIT 1`, serverID, id).Scan(&status)
		if err == pgx.ErrNoRows {
			status = "alive"
		} else if err != nil {
			return nil, nil, err
		}
		patient := HospitalPatient{ID: id, Name: npc.Name, Visual: npc.Visual, Status: status, FingerprintCode: codes[id]}
		if status == "dead" {
			record, items, err := store.postMortemForPatient(ctx, serverID, city, npc, codes)
			if err != nil {
				return nil, nil, err
			}
			patient.PostMortem = record
			discovered = append(discovered, items...)
		}
		patients = append(patients, patient)
		if len(patients) >= world.MaxNPCSearchResults {
			break
		}
	}
	return patients, discovered, nil
}

func (store *PostgresStore) postMortemForPatient(ctx context.Context, serverID string, city world.WorldCity, victim world.NPC, codes map[string]string) (*PostMortemRecord, []DiscoveredItem, error) {
	var caseID, archetype, killerID, sceneID string
	var deathAt time.Time
	err := store.db.QueryRow(ctx, `SELECT c.case_id,c.archetype_id,c.killer_entity_id,c.location_entity_id,COALESCE(c.crime_at,c.created_at)
		FROM cases c JOIN case_states s USING(server_id,case_id)
		WHERE c.server_id=$1::uuid AND c.city_id=$2 AND c.victim_entity_id=$3 AND s.status IN ('open','solved')
		ORDER BY c.created_at DESC,c.case_id DESC LIMIT 1`, serverID, city.ID, victim.ID).Scan(&caseID, &archetype, &killerID, &sceneID, &deathAt)
	if err == pgx.ErrNoRows {
		return nil, nil, nil
	}
	if err != nil {
		return nil, nil, err
	}
	var sceneName, district string
	err = store.db.QueryRow(ctx, `SELECT e.region_id FROM world_entities e WHERE e.server_id=$1::uuid AND e.entity_id=$2`, serverID, sceneID).Scan(&district)
	if err != nil {
		return nil, nil, err
	}
	seed, version, err := store.worldSeedAndVersion(ctx, serverID)
	if err != nil {
		return nil, nil, err
	}
	cityVersion := version
	if city.ID != "fr/lyon" {
		err := store.db.QueryRow(ctx, `SELECT generator_version FROM world_city_generators WHERE server_id=$1::uuid AND city_id=$2`, serverID, city.ID).Scan(&cityVersion)
		if err == pgx.ErrNoRows && cityVersion < world.LatestGeneratorVersion {
			cityVersion = world.LatestGeneratorVersion
		} else if err != nil && err != pgx.ErrNoRows {
			return nil, nil, err
		}
	}
	generated, err := world.GenerateCity(seed, cityVersion, city.ID)
	if err != nil {
		return nil, nil, err
	}
	for _, entity := range generated.Entities {
		if entity.ID == sceneID {
			sceneName = entity.Name
			break
		}
	}
	if sceneName == "" {
		sceneName = sceneID
	}
	var regionName string
	if err := store.db.QueryRow(ctx, hospitalRegionNameQuery, serverID, district).Scan(&regionName); err != nil {
		return nil, nil, err
	}
	address := strings.Trim(strings.Join([]string{sceneName, regionName, city.Name, city.Country}, ", "), ", ")
	causes := map[string]string{"workplace_conflict": "Fatal blunt force injuries", "jealousy": "Fatal penetrating injuries", "financial_extortion": "Fatal blunt force injuries"}
	record := &PostMortemRecord{ReportID: "case:postmortem:" + caseID, EstimatedDeathAt: deathAt.UTC(), Cause: causes[archetype], BodyFoundAt: address, Evidence: []HospitalEvidence{}}
	if record.Cause == "" {
		record.Cause = "Fatal injuries consistent with homicide"
	}
	// Stable, case-scoped trace: a partial fingerprint or hair matching the generated killer's hair color.
	hairColor := ""
	if !postMortemUsesFingerprint(caseID) {
		var key string
		if err := store.db.QueryRow(ctx, `SELECT generation_key FROM world_entities WHERE server_id=$1::uuid AND entity_id=$2`, serverID, killerID).Scan(&key); err != nil {
			return nil, nil, err
		}
		killer, err := world.GenerateNPCFromKey(seed, version, key)
		if err != nil {
			return nil, nil, err
		}
		hairColor = "dark"
		if killer.Visual != nil {
			parts := strings.Split(killer.Visual.HairID, "_")
			hairColor = parts[len(parts)-1]
		}
	}
	evidence := postMortemTrace(caseID, codes[killerID], hairColor)
	record.Evidence = append(record.Evidence, evidence)
	item := DiscoveredItem{ID: evidence.ID, Source: "case", Title: evidence.Title, Content: evidence.Content}
	reportItem := postMortemReportItem(record, city, victim)
	return record, []DiscoveredItem{item, reportItem}, nil
}

func postMortemUsesFingerprint(caseID string) bool {
	hash := sha256.Sum256([]byte("hospital-trace|" + caseID))
	return hash[0]%2 == 0
}

func postMortemTrace(caseID, code, hairColor string) HospitalEvidence {
	evidence := HospitalEvidence{ID: "hospital:" + caseID + ":trace", Source: "case", Title: "Post-mortem trace", Content: "A partial trace was recovered from the victim. Its owner is not identified in this report."}
	if postMortemUsesFingerprint(caseID) {
		if len(code) >= 4 {
			evidence.Content = fmt.Sprintf("A partial fingerprint was recovered from the victim. The visible pattern code is %s. The report does not identify its owner.", code[:4])
		}
	} else if hairColor != "" {
		evidence.Content = fmt.Sprintf("A strand of %s hair was recovered from the victim. The report does not identify its owner.", hairColor)
	}
	return evidence
}

func postMortemReportItem(record *PostMortemRecord, city world.WorldCity, victim world.NPC) DiscoveredItem {
	deathLocation, err := time.LoadLocation(city.TimeZone)
	if err != nil {
		deathLocation = time.UTC
	}
	deathTime := record.EstimatedDeathAt.In(deathLocation).Format("Jan 2, 3:04 PM")
	metadata, _ := json.Marshal(map[string]any{"npc": map[string]any{"id": victim.ID, "name": victim.Name, "visual": victim.Visual}})
	cause := record.Cause
	if record.Weapon != "" {
		cause += " · " + record.Weapon
	}
	return DiscoveredItem{
		ID:       record.ReportID,
		Source:   "case",
		Title:    "POST-MORTEM REPORT",
		Content:  fmt.Sprintf("ESTIMATED TIME OF DEATH · %s\nCAUSE · %s\nBODY FOUND · %s", deathTime, cause, record.BodyFoundAt),
		Metadata: metadata,
	}
}

func fingerprintCodes(ids []string) map[string]string {
	const capacity = 26 * 26 * 100000
	ordered := append([]string(nil), ids...)
	sort.Strings(ordered)
	used := make(map[int]struct{}, len(ordered))
	codes := make(map[string]string, len(ordered))
	for _, id := range ordered {
		hash := sha256.Sum256([]byte("noir-fingerprint-v1|" + id))
		value := int(binary.BigEndian.Uint64(hash[:8]) % capacity)
		for {
			if _, exists := used[value]; !exists {
				break
			}
			value = (value + 1) % capacity
		}
		used[value] = struct{}{}
		letters := value / 100000
		digits := value % 100000
		codes[id] = fmt.Sprintf("%c%c%05d", 'A'+letters/26, 'A'+letters%26, digits)
	}
	return codes
}

func (store *PostgresStore) cityFingerprintCodes(ctx context.Context, serverID, cityID string) (map[string]string, error) {
	seed, version, err := store.worldSeedAndVersion(ctx, serverID)
	if err != nil {
		return nil, err
	}
	cityVersion := version
	if cityID != "fr/lyon" {
		err := store.db.QueryRow(ctx, `SELECT generator_version FROM world_city_generators WHERE server_id=$1::uuid AND city_id=$2`, serverID, cityID).Scan(&cityVersion)
		if err == pgx.ErrNoRows && cityVersion < world.LatestGeneratorVersion {
			cityVersion = world.LatestGeneratorVersion
		} else if err != nil && err != pgx.ErrNoRows {
			return nil, err
		}
	}
	base, err := world.GenerateCity(seed, cityVersion, cityID)
	if err != nil {
		return nil, err
	}
	ids := make([]string, 0, len(base.NPCs))
	seen := make(map[string]struct{}, len(base.NPCs))
	for _, npc := range base.NPCs {
		ids = append(ids, npc.ID)
		seen[npc.ID] = struct{}{}
	}
	cityKey := cityID
	if cityID == "fr/lyon" {
		cityKey = "lyon"
	}
	rows, err := store.db.Query(ctx, `SELECT e.entity_id FROM world_entities e JOIN world_regions r USING(server_id,region_id)
		WHERE e.server_id=$1::uuid AND e.entity_type='npc' AND r.region_id LIKE $2 ORDER BY e.entity_id`, serverID, cityKey+"/district/%")
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	for rows.Next() {
		var id string
		if err := rows.Scan(&id); err != nil {
			return nil, err
		}
		if _, exists := seen[id]; !exists {
			ids = append(ids, id)
			seen[id] = struct{}{}
		}
	}
	if err := rows.Err(); err != nil {
		return nil, err
	}
	return fingerprintCodes(ids), nil
}

func (store *PostgresStore) isResidentialCaseScene(ctx context.Context, serverID, sceneID string) (bool, error) {
	var generationKey string
	err := store.db.QueryRow(ctx, `SELECT generation_key FROM world_entities WHERE server_id=$1::uuid AND entity_id=$2 AND entity_type='location'`, serverID, sceneID).Scan(&generationKey)
	if err == pgx.ErrNoRows {
		return false, nil
	}
	if err != nil {
		return false, err
	}
	return isResidentialGenerationKey(generationKey), nil
}

func isResidentialGenerationKey(generationKey string) bool {
	key := strings.ToLower(generationKey)
	segments := strings.Split(key, "/")
	for index, segment := range segments {
		if segment == "building" && index+1 < len(segments) && residentialName(segments[index+1]) {
			return true
		}
		if index == len(segments)-2 {
			switch segment {
			case "residence", "residential", "apartment", "house", "home", "immeuble", "flat":
				return true
			}
		}
	}
	return false
}

func residentialName(value string) bool {
	for _, marker := range []string{"residence", "residential", "apartment", "house", "home", "immeuble", "flat"} {
		if value == marker || strings.HasPrefix(value, marker+"-") || strings.HasPrefix(value, marker+"_") {
			return true
		}
	}
	return false
}

func deterministicChance(identity, salt string, percentage uint64) bool {
	hash := sha256.Sum256([]byte("noir-deterministic-chance-v1|" + identity + "|" + salt))
	return binary.BigEndian.Uint64(hash[:8])%100 < percentage
}

func residentialFingerprintEligible(caseID string, residential bool) bool {
	return caseID != "" && residential && deterministicChance(caseID, "residential-fingerprint", 35)
}

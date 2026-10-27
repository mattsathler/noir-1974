package cases

import (
	"crypto/sha256"
	"encoding/binary"
	"encoding/hex"
	"errors"
	"fmt"
	"sort"
	"strings"
	"time"

	"noir-1974/apps/api/internal/mail"
	"noir-1974/apps/api/internal/records"
	"noir-1974/apps/api/internal/world"
)

const GeneratorVersion = 4
const legacyGeneratorVersion = 3

var (
	ErrInvalidCaseSeed       = errors.New("case seed is required")
	ErrUnsupportedGenerator  = errors.New("unsupported case generator version")
	ErrNoValidCaseCandidates = errors.New("world has no valid homicide candidates")
	ErrInvalidDefinition     = errors.New("invalid case definition")
	ErrUnsolvableCase        = errors.New("case has no reachable clue")
)

type Definition struct {
	ID               string                  `json:"id"`
	CityID           string                  `json:"cityId"`
	Seed             string                  `json:"caseSeed"`
	GeneratorVersion int                     `json:"generatorVersion"`
	CaseType         string                  `json:"caseType"`
	ArchetypeID      string                  `json:"archetypeId"`
	Motive           string                  `json:"motive"`
	VictimID         string                  `json:"victimId"`
	KillerID         string                  `json:"killerId"`
	LocationID       string                  `json:"locationId"`
	CrimeAt          time.Time               `json:"-"`
	ParamourID       string                  `json:"-"`
	Relationships    []RelationshipReference `json:"relationships"`
	Evidence         []EvidenceDefinition    `json:"evidence"`
	Records          []records.Record        `json:"records,omitempty"`
	Emails           []mail.Message          `json:"-"`
}

type RelationshipReference struct {
	ID   string `json:"id"`
	Type string `json:"type"`
	Role string `json:"role"`
}

type EvidenceDefinition struct {
	ID                     string   `json:"id"`
	Type                   string   `json:"type"`
	Presentation           string   `json:"presentation"`
	Content                string   `json:"content"`
	LocationID             string   `json:"locationId"`
	RelationshipIDs        []string `json:"relationshipIds"`
	ReachableFromEntityIDs []string `json:"reachableFromEntityIds"`
}

type employmentLink struct {
	id       string
	employee string
}

type candidate struct {
	archetypeID   string
	motive        string
	relationships []RelationshipReference
	evidence      []evidencePlan
	killerID      string
	victimID      string
	locationID    string
	paramourID    string
}

type evidencePlan struct {
	typeID       string
	presentation string
	content      string
}

type archetypeGenerator struct {
	id            string
	candidates    func(world.World) []candidate
	validate      func(Definition, world.World) bool
	evidenceTypes []string
}

var archetypeGenerators = []archetypeGenerator{
	{id: "workplace_conflict", candidates: workplaceCandidates, validate: validateWorkplaceConflict, evidenceTypes: []string{"employment_record", "meeting_minutes", "office_log"}},
	{id: "jealousy", candidates: jealousyCandidates, validate: validateJealousy, evidenceTypes: []string{"civil_record", "letter_fragment", "witness_statement"}},
	{id: "financial_extortion", candidates: extortionCandidates, validate: validateExtortion, evidenceTypes: []string{"account_record", "travel_register", "personnel_record"}},
}

// Generate creates a deterministic homicide definition from existing world facts.
func Generate(base world.World, seed string, version int) (Definition, error) {
	return GenerateExcluding(base, seed, version, nil)
}

// GenerateExcluding creates a case without reusing NPCs assigned to prior cases.
func GenerateExcluding(base world.World, seed string, version int, excluded map[string]struct{}) (Definition, error) {
	if strings.TrimSpace(seed) == "" {
		return Definition{}, ErrInvalidCaseSeed
	}
	if version != GeneratorVersion && version != legacyGeneratorVersion {
		return Definition{}, fmt.Errorf("%w: %d", ErrUnsupportedGenerator, version)
	}
	candidateWorld := base
	if len(excluded) > 0 {
		candidateWorld.NPCs = make([]world.NPC, 0, len(base.NPCs))
		for _, npc := range base.NPCs {
			if _, used := excluded[npc.ID]; !used {
				candidateWorld.NPCs = append(candidateWorld.NPCs, npc)
			}
		}
		candidateWorld.EmploymentRelationships = make([]world.EmploymentRelationship, 0, len(base.EmploymentRelationships))
		for _, relation := range base.EmploymentRelationships {
			if _, used := excluded[relation.EmployeeNPCID]; !used {
				candidateWorld.EmploymentRelationships = append(candidateWorld.EmploymentRelationships, relation)
			}
		}
		candidateWorld.PersonalRelationships = make([]world.PersonalRelationship, 0, len(base.PersonalRelationships))
		for _, relation := range base.PersonalRelationships {
			_, usedA := excluded[relation.PersonANPCID]
			_, usedB := excluded[relation.PersonBNPCID]
			if !usedA && !usedB {
				candidateWorld.PersonalRelationships = append(candidateWorld.PersonalRelationships, relation)
			}
		}
	}
	available := availableArchetypesForVersion(candidateWorld, version)
	if len(available) == 0 {
		return Definition{}, ErrNoValidCaseCandidates
	}
	archetype := available[caseIndex(seed+"/archetype", version, len(available))]
	candidates := archetypeCandidates(archetype, candidateWorld, version)
	selected := candidates[caseIndex(seed+"/"+archetype.id, version, len(candidates))]
	caseID := makeCaseID(seed, version)
	caseDate := time.Now().UTC().Truncate(24 * time.Hour)
	crimeAt := caseDate.Add(time.Duration(18*60+caseIndex(seed+"/crime-time", version, 4*60)) * time.Minute)
	definition := Definition{
		ID:               caseID,
		CityID:           cityForLocation(base, selected.locationID),
		Seed:             seed,
		GeneratorVersion: version,
		CaseType:         "homicide",
		ArchetypeID:      selected.archetypeID,
		Motive:           selected.motive,
		VictimID:         selected.victimID,
		KillerID:         selected.killerID,
		LocationID:       selected.locationID,
		Relationships:    selected.relationships,
		Evidence:         buildEvidence(caseID, selected),
	}
	if version == GeneratorVersion {
		definition.CrimeAt = crimeAt
		definition.ParamourID = selected.paramourID
	}
	if version == legacyGeneratorVersion && selected.archetypeID == "financial_extortion" {
		region := regionName(base, npcByID(base.NPCs, selected.killerID).RegionID)
		definition.Records = records.ExtortionEvents(seed, base.GeneratorVersion, caseID, selected.killerID, selected.victimID, region, caseDate)
	}
	if version == GeneratorVersion {
		definition.Records = buildConnectedCaseRecords(seed, base.GeneratorVersion, caseID, selected, base, crimeAt)
		definition.Emails = buildConnectedCaseEmails(seed, base.GeneratorVersion, caseID, selected, base, crimeAt)
	} else {
		definition.Emails = buildCaseEmails(seed, base.GeneratorVersion, caseID, selected, base, caseDate)
	}
	if err := Validate(definition, base); err != nil {
		return Definition{}, err
	}
	return definition, nil
}

func availableArchetypesForVersion(base world.World, version int) []archetypeGenerator {
	available := make([]archetypeGenerator, 0, len(archetypeGenerators))
	for _, archetype := range archetypeGenerators {
		if len(archetypeCandidates(archetype, base, version)) > 0 {
			available = append(available, archetype)
		}
	}
	return available
}

func archetypeCandidates(archetype archetypeGenerator, base world.World, version int) []candidate {
	if version == GeneratorVersion && archetype.id == "jealousy" {
		return jealousyCandidatesV4(base)
	}
	return archetype.candidates(base)
}

func cityForLocation(base world.World, locationID string) string {
	for _, entity := range base.Entities {
		if entity.ID != locationID {
			continue
		}
		if index := strings.Index(entity.RegionID, "/district/"); index > 0 {
			return entity.RegionID[:index]
		}
	}
	return "fr/lyon"
}

func buildCaseEmails(seed string, worldVersion int, caseID string, selected candidate, base world.World, caseDate time.Time) []mail.Message {
	victim := npcByID(base.NPCs, selected.victimID)
	firstSubject, firstBody := "A private warning", "Leave the meeting notes as they are. If you take them to the registry, you will regret it."
	lastSubject, lastBody := "One last message", "This is your final chance to keep quiet. Do not make me come looking for you."
	switch selected.archetypeID {
	case "jealousy":
		firstSubject = "Do not ask about the meeting"
		firstBody = "You were told to leave that evening alone. Stop asking who was there, and do not show the letter to anyone."
		lastSubject = "You have been warned"
		lastBody = "Keep the letter to yourself. If this becomes a public matter, you will wish you had listened."
	case "financial_extortion":
		firstSubject = "The remittance"
		firstBody = "The next payment is due before Friday. I still have the duplicate settlement statement. Do not take the ledger to the registry."
		lastSubject = "Final notice"
		lastBody = "The payment is late. Send it as agreed, or the duplicate statement will reach people you would rather not involve."
	}
	messageDay := 3 + caseIndex(seed+"/correspondence/day", 1, 8)
	messageMinute := 7*60 + caseIndex(seed+"/correspondence/minute", 1, 13*60)
	lastSentAt := caseDate.AddDate(0, 0, messageDay-12).Add(time.Duration(messageMinute) * time.Minute)
	firstSentAt := lastSentAt.Add(-48 * time.Hour)
	return []mail.Message{
		mail.CaseMessage(seed, worldVersion, caseID, victim.ID, victim.ID, "Unknown sender", victim.Name, firstSubject, firstBody, "anonymous_warning", firstSentAt),
		mail.CaseMessage(seed, worldVersion, caseID, victim.ID, victim.ID, "Unknown sender", victim.Name, lastSubject, lastBody, "anonymous_followup", lastSentAt),
	}
}

func buildConnectedCaseRecords(seed string, worldVersion int, caseID string, selected candidate, base world.World, crimeAt time.Time) []records.Record {
	key := fmt.Sprintf("noir-case-records/v4/%s/%d/%s", seed, worldVersion, caseID)
	scene := regionName(base, locationRegion(base, selected.locationID))
	killerOrigin := regionName(base, npcByID(base.NPCs, selected.killerID).RegionID)
	victimOrigin := regionName(base, npcByID(base.NPCs, selected.victimID).RegionID)
	if killerOrigin == scene {
		killerOrigin = alternateDistrict(base, scene)
	}
	if victimOrigin == scene {
		victimOrigin = alternateDistrict(base, scene)
	}
	result := []records.Record{
		caseGateRecord(key, caseID, "scene-victim", selected.victimID, selected.killerID, crimeAt.Add(-38*time.Minute), victimOrigin, scene, "A passenger entry was recorded at the local relay."),
		caseGateRecord(key, caseID, "scene-killer", selected.killerID, selected.victimID, crimeAt.Add(-19*time.Minute), killerOrigin, scene, "A passenger entry was recorded at the local relay."),
	}
	if selected.archetypeID == "financial_extortion" {
		amount := 45 + keyIndex(key, "payment-amount", 46)
		reference := fmt.Sprintf("unregistered account %04d", keyIndex(key, "account-reference", 10000))
		paidAt := crimeAt.AddDate(0, 0, -2).Add(-70 * time.Minute)
		result = append(result,
			records.Record{ID: caseRecordID(key, "payment-out", selected.victimID), NPCID: selected.victimID, Source: "bank", OccurredAt: paidAt, Operator: "Meridian Global Settlement Office", Description: "An anonymous remittance was sent and cleared.", Reference: reference, Amount: amount, Direction: "outgoing", CaseID: caseID, InternalRelation: "extortion_payment", RelatedNPCID: selected.killerID},
			records.Record{ID: caseRecordID(key, "payment-in", selected.killerID), NPCID: selected.killerID, Source: "bank", OccurredAt: paidAt, Operator: "Meridian Global Settlement Office", Description: "An anonymous remittance was received and cleared.", Reference: reference, Amount: amount, Direction: "incoming", CaseID: caseID, InternalRelation: "extortion_payment", RelatedNPCID: selected.victimID},
		)
		result = append(result, buildTransferDecoys(key, caseID, base, selected, crimeAt)...)
	}
	if selected.archetypeID == "jealousy" && selected.paramourID != "" {
		paramourOrigin := regionName(base, npcByID(base.NPCs, selected.paramourID).RegionID)
		if paramourOrigin == scene {
			paramourOrigin = alternateDistrict(base, scene)
		}
		rendezvous := crimeAt.AddDate(0, 0, -3).Add(-42 * time.Minute)
		result = append(result,
			caseGateRecord(key, caseID, "affair-victim", selected.victimID, selected.paramourID, rendezvous, victimOrigin, scene, "A passenger entry was recorded at the local relay."),
			caseGateRecord(key, caseID, "affair-paramour", selected.paramourID, selected.victimID, rendezvous, paramourOrigin, scene, "A passenger entry was recorded at the local relay."),
		)
	}
	sort.Slice(result, func(i, j int) bool {
		if result[i].OccurredAt.Equal(result[j].OccurredAt) {
			return result[i].ID < result[j].ID
		}
		return result[i].OccurredAt.Before(result[j].OccurredAt)
	})
	return result
}

func caseGateRecord(key, caseID, role, npcID, relatedID string, at time.Time, origin, destination, description string) records.Record {
	return records.Record{ID: caseRecordID(key, role, npcID), NPCID: npcID, Source: "gate", OccurredAt: at.UTC(), Operator: "Meridian Global · GATE Network", Description: description, Origin: origin, Destination: destination, CaseID: caseID, InternalRelation: role, RelatedNPCID: relatedID}
}

func buildTransferDecoys(key, caseID string, base world.World, selected candidate, crimeAt time.Time) []records.Record {
	caseCityID := cityForLocation(base, selected.locationID)
	people := make([]world.NPC, 0, len(base.NPCs))
	for _, person := range base.NPCs {
		personCityID := person.RegionID
		if districtIndex := strings.Index(personCityID, "/district/"); districtIndex >= 0 {
			personCityID = personCityID[:districtIndex]
		}
		if person.ID != selected.killerID && person.ID != selected.victimID && personCityID == caseCityID {
			people = append(people, person)
		}
	}
	sort.Slice(people, func(i, j int) bool { return people[i].ID < people[j].ID })
	if len(people) < 2 {
		return nil
	}
	limit := min(4, len(people))
	result := make([]records.Record, 0, limit)
	for index := 0; index+1 < limit; index += 2 {
		from, to := people[index], people[index+1]
		amount := 20 + keyIndex(key, "decoy-amount/"+from.ID, 120)
		reference := fmt.Sprintf("unregistered account %04d", keyIndex(key, "decoy-account/"+from.ID, 10000))
		at := crimeAt.AddDate(0, 0, -4-index/2).Add(time.Duration(9*60+keyIndex(key, "decoy-time/"+from.ID, 8*60)) * time.Minute)
		result = append(result,
			records.Record{ID: caseRecordID(key, "decoy-out/"+from.ID, to.ID), NPCID: from.ID, Source: "bank", OccurredAt: at, Operator: "Meridian Global Settlement Office", Description: "An anonymous remittance was sent and cleared.", Reference: reference, Amount: amount, Direction: "outgoing", CaseID: caseID, InternalRelation: "unrelated_transfer", RelatedNPCID: to.ID},
			records.Record{ID: caseRecordID(key, "decoy-in/"+to.ID, from.ID), NPCID: to.ID, Source: "bank", OccurredAt: at, Operator: "Meridian Global Settlement Office", Description: "An anonymous remittance was received and cleared.", Reference: reference, Amount: amount, Direction: "incoming", CaseID: caseID, InternalRelation: "unrelated_transfer", RelatedNPCID: from.ID},
		)
	}
	return result
}

func buildConnectedCaseEmails(seed string, worldVersion int, caseID string, selected candidate, base world.World, crimeAt time.Time) []mail.Message {
	killer, victim := npcByID(base.NPCs, selected.killerID), npcByID(base.NPCs, selected.victimID)
	messageAt := crimeAt.Add(-26 * time.Hour)
	subject, body := "We need to talk", "I know you were at the GATE after the office closed. Please tell me who you met before this becomes something we cannot explain."
	if selected.archetypeID == "financial_extortion" {
		subject, body = "The last remittance", "The payment cleared. Use the same amount next time and keep the account reference out of your ledger. If anyone asks, you have never met me."
	} else if selected.archetypeID == "workplace_conflict" {
		subject, body = "About the office register", "The corrected entry is still there. I thought we agreed to leave the matter alone until the audit was finished."
	}
	mailbox := mail.CaseMessage(seed, worldVersion, caseID, victim.ID, killer.ID, killer.Name, victim.Name, subject, body, "case_correspondence", messageAt)
	if caseIndex(seed+"/anonymous-case-mail", 1, 5) == 0 {
		mailbox.InternalRelation = "anonymous_warning"
	}
	result := []mail.Message{mailbox}
	if selected.archetypeID == "jealousy" && selected.paramourID != "" {
		paramour := npcByID(base.NPCs, selected.paramourID)
		message := mail.CaseMessage(seed, worldVersion, caseID, victim.ID, paramour.ID, victim.Name, paramour.Name, "About last Tuesday", "It was only dinner, but my husband was asking where I had been. I cannot keep making excuses for these meetings.", "jealousy_exchange", crimeAt.AddDate(0, 0, -4))
		result = append(result, message)
	}
	if correspondent := killerCorrespondent(base, selected); correspondent.ID != "" {
		confession := mail.CaseMessage(seed, worldVersion, caseID, killer.ID, correspondent.ID, killer.Name, correspondent.Name, "I made a mistake", "I did something I cannot take back. If anyone asks, tell them I left before the last GATE call. Please do not ask me what happened.", "killer_remorse", crimeAt.Add(2*time.Hour))
		result = append(result, confession)
	}
	return result
}

func killerCorrespondent(base world.World, selected candidate) world.NPC {
	var candidates []world.NPC
	for _, relation := range base.PersonalRelationships {
		if relation.PersonANPCID == selected.killerID && relation.PersonBNPCID != selected.victimID {
			person := npcByID(base.NPCs, relation.PersonBNPCID)
			if person.ID != "" {
				candidates = append(candidates, person)
			}
		}
		if relation.PersonBNPCID == selected.killerID && relation.PersonANPCID != selected.victimID {
			person := npcByID(base.NPCs, relation.PersonANPCID)
			if person.ID != "" {
				candidates = append(candidates, person)
			}
		}
	}
	if len(candidates) > 0 {
		sort.Slice(candidates, func(i, j int) bool { return candidates[i].ID < candidates[j].ID })
		return candidates[0]
	}
	companyIDs := make(map[string]bool)
	for _, relation := range base.EmploymentRelationships {
		if relation.EmployeeNPCID == selected.killerID {
			companyIDs[relation.EmployerCompanyID] = true
		}
	}
	for _, relation := range base.EmploymentRelationships {
		if companyIDs[relation.EmployerCompanyID] && relation.EmployeeNPCID != selected.killerID && relation.EmployeeNPCID != selected.victimID && relation.EmployeeNPCID != selected.paramourID {
			person := npcByID(base.NPCs, relation.EmployeeNPCID)
			if person.ID != "" {
				candidates = append(candidates, person)
			}
		}
	}
	sort.Slice(candidates, func(i, j int) bool { return candidates[i].ID < candidates[j].ID })
	if len(candidates) == 0 {
		return world.NPC{}
	}
	return candidates[0]
}

func locationRegion(base world.World, locationID string) string {
	for _, entity := range base.Entities {
		if entity.ID == locationID {
			return entity.RegionID
		}
	}
	return ""
}

func alternateDistrict(base world.World, current string) string {
	var names []string
	for _, region := range base.Regions {
		if region.Name != current && strings.Contains(region.ID, "/district/") {
			names = append(names, region.Name)
		}
	}
	sort.Strings(names)
	if len(names) > 0 {
		return names[0]
	}
	return "Central District"
}

func caseRecordID(key, role, npcID string) string {
	digest := sha256.Sum256([]byte(key + "/" + role + "/" + npcID))
	return "record:" + hex.EncodeToString(digest[:12])
}

func keyIndex(key, field string, limit int) int {
	digest := sha256.Sum256([]byte(key + "/" + field))
	return int(binary.BigEndian.Uint64(digest[:8]) % uint64(limit))
}

// Validate checks the case archetype, world references, and reachable clues.
func Validate(definition Definition, base world.World) error {
	if !world.SupportsSocietyVersion(base.GeneratorVersion) {
		return ErrInvalidDefinition
	}
	if strings.TrimSpace(definition.Seed) == "" || definition.GeneratorVersion != GeneratorVersion && definition.GeneratorVersion != legacyGeneratorVersion || definition.ID != makeCaseID(definition.Seed, definition.GeneratorVersion) || definition.CaseType != "homicide" || definition.ArchetypeID == "" || definition.Motive == "" {
		return ErrInvalidDefinition
	}
	if definition.GeneratorVersion == GeneratorVersion && definition.CrimeAt.IsZero() {
		return ErrInvalidDefinition
	}
	if definition.KillerID == definition.VictimID {
		return ErrInvalidDefinition
	}
	entities := make(map[string]string, len(base.Entities))
	for _, entity := range base.Entities {
		entities[entity.ID] = entity.EntityType
	}
	if entities[definition.KillerID] != "npc" || entities[definition.VictimID] != "npc" || entities[definition.LocationID] != "location" {
		return ErrInvalidDefinition
	}
	if !containsNPC(base.NPCs, definition.KillerID) || !containsNPC(base.NPCs, definition.VictimID) {
		return ErrInvalidDefinition
	}
	links := make(map[string]RelationshipReference, len(definition.Relationships))
	for _, relation := range definition.Relationships {
		if relation.ID == "" || relation.Type == "" || relation.Role == "" || !worldRelationshipExists(base, relation.Type, relation.ID) {
			return ErrInvalidDefinition
		}
		key := relation.Type + ":" + relation.ID
		if _, exists := links[key]; exists {
			return ErrInvalidDefinition
		}
		links[key] = relation
	}
	if len(links) == 0 {
		return ErrInvalidDefinition
	}
	archetype, exists := archetypeByID(definition.ArchetypeID)
	validArchetype := exists && archetype.validate(definition, base)
	if definition.GeneratorVersion == GeneratorVersion {
		switch definition.ArchetypeID {
		case "jealousy":
			validArchetype = validateJealousyV4(definition, base)
		case "financial_extortion":
			validArchetype = validateExtortionV4(definition, base)
		}
	}
	if !validArchetype {
		return ErrInvalidDefinition
	}
	if !hasReachableEvidence(definition.Evidence, definition, links, entities) || !hasEvidenceTypes(definition.Evidence, archetype.evidenceTypes...) || !validateCaseRecords(definition, base) || definition.GeneratorVersion == GeneratorVersion && !validateConnectedCaseEmails(definition, base) {
		return ErrUnsolvableCase
	}
	return nil
}

func availableArchetypes(base world.World) []archetypeGenerator {
	var available []archetypeGenerator
	for _, archetype := range archetypeGenerators {
		if len(archetype.candidates(base)) > 0 {
			available = append(available, archetype)
		}
	}
	return available
}

func workplaceCandidates(base world.World) []candidate {
	if !world.SupportsSocietyVersion(base.GeneratorVersion) {
		return nil
	}
	entities := make(map[string]string, len(base.Entities))
	for _, entity := range base.Entities {
		entities[entity.ID] = entity.EntityType
	}
	companyLocations := make(map[string]string)
	for _, relation := range base.CompanyLocations {
		if relation.LocationRole == "headquarters" && entities[relation.LocationID] == "location" {
			companyLocations[relation.CompanyID] = relation.LocationID
		}
	}
	byCompany := make(map[string][]employmentLink)
	for _, relation := range base.EmploymentRelationships {
		if entities[relation.EmployeeNPCID] != "npc" || entities[relation.EmployerCompanyID] != "company" {
			continue
		}
		byCompany[relation.EmployerCompanyID] = append(byCompany[relation.EmployerCompanyID], employmentLink{
			id: relation.ID, employee: relation.EmployeeNPCID,
		})
	}
	companies := make([]string, 0, len(byCompany))
	for companyID := range byCompany {
		companies = append(companies, companyID)
	}
	sort.Strings(companies)
	result := make([]candidate, 0)
	for _, companyID := range companies {
		locationID := companyLocations[companyID]
		if locationID == "" {
			continue
		}
		links := byCompany[companyID]
		sort.Slice(links, func(i, j int) bool { return links[i].employee < links[j].employee })
		for i := 0; i < len(links); i++ {
			for j := i + 1; j < len(links); j++ {
				result = append(result,
					workplaceCandidate(links[i].employee, links[j].employee, locationID, links[i].id, links[j].id),
					workplaceCandidate(links[j].employee, links[i].employee, locationID, links[j].id, links[i].id),
				)
			}
		}
	}
	return result
}

func workplaceCandidate(killerID, victimID, locationID, killerLinkID, victimLinkID string) candidate {
	return candidate{
		archetypeID: "workplace_conflict",
		motive:      "A long-running dispute over status and routine office decisions.",
		killerID:    killerID,
		victimID:    victimID,
		locationID:  locationID,
		relationships: []RelationshipReference{
			{ID: killerLinkID, Type: "employment", Role: "killer_employment"},
			{ID: victimLinkID, Type: "employment", Role: "victim_employment"},
		},
		evidence: []evidencePlan{
			{typeID: "employment_record", presentation: "Personnel file", content: "A routine personnel entry places both people within the same company."},
			{typeID: "meeting_minutes", presentation: "Meeting minutes", content: "A brief note records a disagreement over responsibilities. The names are abbreviated."},
			{typeID: "office_log", presentation: "After-hours register", content: "Two familiar signatures appear on a late entry; the time has been corrected by hand."},
		},
	}
}

func extortionCandidates(base world.World) []candidate {
	workplace := workplaceCandidates(base)
	result := make([]candidate, 0, len(workplace))
	for _, possibility := range workplace {
		possibility.archetypeID = "financial_extortion"
		possibility.motive = "Financial extortion escalated after the victim threatened to expose a private arrangement."
		possibility.evidence = []evidencePlan{
			{typeID: "account_record", presentation: "Settlement statement", content: "A private remittance appears among routine account activity. The counterparty is listed only by ledger reference."},
			{typeID: "travel_register", presentation: "Passenger register", content: "A GATE passage was recorded shortly before the reported time of death."},
			{typeID: "personnel_record", presentation: "Personnel file", content: "Two staff records place the names within the same corporate office."},
		}
		result = append(result, possibility)
	}
	return result
}

func jealousyCandidates(base world.World) []candidate {
	locationsByRegion := make(map[string][]string)
	for _, entity := range base.Entities {
		if entity.EntityType == "location" {
			locationsByRegion[entity.RegionID] = append(locationsByRegion[entity.RegionID], entity.ID)
		}
	}
	var result []candidate
	for _, relation := range base.PersonalRelationships {
		if relation.RelationshipType != "marriage" {
			continue
		}
		first, second := npcByID(base.NPCs, relation.PersonANPCID), npcByID(base.NPCs, relation.PersonBNPCID)
		if first.ID == "" || second.ID == "" || first.RegionID != second.RegionID {
			continue
		}
		locations := locationsByRegion[first.RegionID]
		if len(locations) == 0 {
			continue
		}
		sort.Strings(locations)
		locationID := locations[0]
		for _, orientation := range [][2]string{{first.ID, second.ID}, {second.ID, first.ID}} {
			result = append(result, candidate{
				archetypeID: "jealousy",
				motive:      "Possessive jealousy tied to a private romantic suspicion.",
				killerID:    orientation[0], victimID: orientation[1], locationID: locationID,
				relationships: []RelationshipReference{{ID: relation.ID, Type: "personal", Role: "marriage"}},
				evidence: []evidencePlan{
					{typeID: "civil_record", presentation: "Civil registry extract", content: "The record confirms a private connection between the two names. It gives no reason for the death."},
					{typeID: "letter_fragment", presentation: "Unsent correspondence", content: "A torn sentence refers to a meeting and asks that it remain undisclosed."},
					{typeID: "witness_statement", presentation: "A neighbor's note", content: "A resident recalls seeing one spouse return unusually late, but cannot give an exact time."},
				},
			})
		}
	}
	return result
}

// jealousyCandidatesV4 requires a real spouse pair and a workplace acquaintance
// who can serve as the suspected lover in the generated paper and travel trails.
func jealousyCandidatesV4(base world.World) []candidate {
	var result []candidate
	locationsByRegion := make(map[string][]string)
	for _, entity := range base.Entities {
		if entity.EntityType == "location" {
			locationsByRegion[entity.RegionID] = append(locationsByRegion[entity.RegionID], entity.ID)
		}
	}
	for _, relation := range base.PersonalRelationships {
		if relation.RelationshipType != "marriage" {
			continue
		}
		first, second := npcByID(base.NPCs, relation.PersonANPCID), npcByID(base.NPCs, relation.PersonBNPCID)
		if first.ID == "" || second.ID == "" || first.RegionID != second.RegionID {
			continue
		}
		locations := append([]string(nil), locationsByRegion[first.RegionID]...)
		if len(locations) == 0 {
			continue
		}
		sort.Strings(locations)
		for _, spouseOrientation := range [][2]world.NPC{{first, second}, {second, first}} {
			killer, victim := spouseOrientation[0], spouseOrientation[1]
			var paramours []employmentLink
			victimEmploymentID := ""
			for _, employment := range base.EmploymentRelationships {
				if employment.EmployeeNPCID == victim.ID {
					victimEmploymentID = employment.ID
					for _, other := range base.EmploymentRelationships {
						if other.EmployerCompanyID == employment.EmployerCompanyID && other.EmployeeNPCID != killer.ID && other.EmployeeNPCID != victim.ID {
							person := npcByID(base.NPCs, other.EmployeeNPCID)
							if person.ID != "" && person.RegionID == victim.RegionID {
								paramours = append(paramours, employmentLink{id: other.ID, employee: person.ID})
							}
						}
					}
				}
			}
			sort.Slice(paramours, func(i, j int) bool { return paramours[i].employee < paramours[j].employee })
			for _, paramour := range paramours {
				result = append(result, candidate{
					archetypeID: "jealousy", motive: "A spouse's suspicion of a private affair escalated into violence.",
					killerID: killer.ID, victimID: victim.ID, locationID: locations[0], paramourID: paramour.employee,
					relationships: []RelationshipReference{{ID: relation.ID, Type: "personal", Role: "marriage"}, {ID: victimEmploymentID, Type: "employment", Role: "victim_employment"}, {ID: paramour.id, Type: "employment", Role: "paramour_employment"}},
					evidence: []evidencePlan{
						{typeID: "civil_record", presentation: "Civil registry extract", content: "The record confirms a marriage between two names. It gives no reason for the death."},
						{typeID: "letter_fragment", presentation: "Private correspondence", content: "A torn exchange refers to a repeated meeting and asks that the spouse not be told."},
						{typeID: "witness_statement", presentation: "A neighbor's note", content: "A resident recalls the victim meeting a colleague after work on several evenings."},
					},
				})
			}
		}
	}
	return result
}

func buildEvidence(caseID string, selected candidate) []EvidenceDefinition {
	evidence := make([]EvidenceDefinition, 0, len(selected.evidence))
	for index, plan := range selected.evidence {
		evidence = append(evidence, EvidenceDefinition{
			ID:                     fmt.Sprintf("evidence:%s:%02d:%s", caseID, index+1, plan.typeID),
			Type:                   plan.typeID,
			Presentation:           plan.presentation,
			Content:                plan.content,
			LocationID:             selected.locationID,
			RelationshipIDs:        relationshipIDs(selected.relationships),
			ReachableFromEntityIDs: []string{selected.killerID, selected.victimID},
		})
	}
	return evidence
}

func relationshipIDs(relations []RelationshipReference) []string {
	ids := make([]string, 0, len(relations))
	for _, relation := range relations {
		ids = append(ids, relation.ID)
	}
	return ids
}

func npcByID(npcs []world.NPC, id string) world.NPC {
	for _, npc := range npcs {
		if npc.ID == id {
			return npc
		}
	}
	return world.NPC{}
}

func archetypeByID(id string) (archetypeGenerator, bool) {
	for _, archetype := range archetypeGenerators {
		if archetype.id == id {
			return archetype, true
		}
	}
	return archetypeGenerator{}, false
}

func worldRelationshipExists(base world.World, kind, id string) bool {
	switch kind {
	case "employment":
		for _, relation := range base.EmploymentRelationships {
			if relation.ID == id {
				return true
			}
		}
	case "personal":
		for _, relation := range base.PersonalRelationships {
			if relation.ID == id {
				return true
			}
		}
	}
	return false
}

func validateWorkplaceConflict(definition Definition, base world.World) bool {
	var killerLink, victimLink *world.EmploymentRelationship
	for index := range base.EmploymentRelationships {
		relation := &base.EmploymentRelationships[index]
		for _, ref := range definition.Relationships {
			if ref.Type == "employment" && ref.ID == relation.ID {
				if ref.Role == "killer_employment" && relation.EmployeeNPCID == definition.KillerID {
					killerLink = relation
				}
				if ref.Role == "victim_employment" && relation.EmployeeNPCID == definition.VictimID {
					victimLink = relation
				}
			}
		}
	}
	return killerLink != nil && victimLink != nil && killerLink.EmployerCompanyID == victimLink.EmployerCompanyID && hasHeadquarters(base.CompanyLocations, killerLink.EmployerCompanyID, definition.LocationID)
}

func validateJealousy(definition Definition, base world.World) bool {
	for _, relation := range base.PersonalRelationships {
		referenced := false
		for _, ref := range definition.Relationships {
			if ref.Type == "personal" && ref.Role == "marriage" && ref.ID == relation.ID {
				referenced = true
			}
		}
		isCouple := relation.PersonANPCID == definition.KillerID && relation.PersonBNPCID == definition.VictimID || relation.PersonANPCID == definition.VictimID && relation.PersonBNPCID == definition.KillerID
		if referenced && relation.RelationshipType == "marriage" && isCouple {
			return containsLocationInRegion(base, definition.LocationID, npcByID(base.NPCs, definition.KillerID).RegionID)
		}
	}
	return false
}

func validateJealousyV4(definition Definition, base world.World) bool {
	if !validateJealousy(definition, base) || definition.KillerID == "" || definition.VictimID == "" {
		return false
	}
	var victimEmployment, paramourEmployment *world.EmploymentRelationship
	for index := range base.EmploymentRelationships {
		relation := &base.EmploymentRelationships[index]
		for _, reference := range definition.Relationships {
			if reference.Type == "employment" && reference.Role == "victim_employment" && reference.ID == relation.ID && relation.EmployeeNPCID == definition.VictimID {
				victimEmployment = relation
			}
			if reference.Type == "employment" && reference.Role == "paramour_employment" && reference.ID == relation.ID && relation.EmployeeNPCID == definition.ParamourID {
				paramourEmployment = relation
			}
		}
	}
	return victimEmployment != nil && paramourEmployment != nil && victimEmployment.EmployerCompanyID == paramourEmployment.EmployerCompanyID && definition.ParamourID != definition.KillerID && definition.ParamourID != definition.VictimID && npcByID(base.NPCs, definition.ParamourID).ID != ""
}

func validateExtortion(definition Definition, base world.World) bool {
	if !validateWorkplaceConflict(definition, base) || len(definition.Records) != 2 {
		return false
	}
	hasTransfer, hasTravel := false, false
	for _, item := range definition.Records {
		switch item.InternalRelation {
		case "extortion_payment":
			hasTransfer = item.CaseID == definition.ID && item.NPCID == definition.VictimID && item.RelatedNPCID == definition.KillerID && item.Source == "bank" && item.Amount > 0
		case "precrime_route":
			hasTravel = item.CaseID == definition.ID && item.NPCID == definition.KillerID && item.Source == "gate" && item.Destination != ""
		}
	}
	return hasTransfer && hasTravel
}

func validateExtortionV4(definition Definition, base world.World) bool {
	return validateWorkplaceConflict(definition, base)
}

func validateCaseRecords(definition Definition, base world.World) bool {
	if definition.GeneratorVersion == GeneratorVersion {
		return validateConnectedCaseRecords(definition, base)
	}
	if definition.ArchetypeID != "financial_extortion" {
		return len(definition.Records) == 0
	}
	if len(definition.Records) != 2 {
		return false
	}
	seen := make(map[string]bool, len(definition.Records))
	for _, item := range definition.Records {
		if item.ID == "" || item.NPCID == "" || item.OccurredAt.IsZero() || item.CaseID != definition.ID || seen[item.ID] {
			return false
		}
		seen[item.ID] = true
	}
	return true
}

func validateConnectedCaseRecords(definition Definition, base world.World) bool {
	if definition.CrimeAt.IsZero() {
		return false
	}
	expectedScene := regionName(base, locationRegion(base, definition.LocationID))
	scene, victimScene, killerScene := "", false, false
	var victimGate, killerGate, affairVictim, affairParamour *records.Record
	var victimPayment, killerReceipt *records.Record
	for index := range definition.Records {
		item := &definition.Records[index]
		if item.ID == "" || item.NPCID == "" || !containsNPC(base.NPCs, item.NPCID) || item.OccurredAt.IsZero() || item.CaseID != definition.ID || item.RelatedNPCID != "" && !containsNPC(base.NPCs, item.RelatedNPCID) {
			return false
		}
		switch item.InternalRelation {
		case "scene-victim":
			if item.Source != "gate" || item.NPCID != definition.VictimID || item.RelatedNPCID != definition.KillerID || !item.OccurredAt.Before(definition.CrimeAt) {
				return false
			}
			victimGate, scene, victimScene = item, item.Destination, true
		case "scene-killer":
			if item.Source != "gate" || item.NPCID != definition.KillerID || item.RelatedNPCID != definition.VictimID || !item.OccurredAt.Before(definition.CrimeAt) {
				return false
			}
			killerGate, scene, killerScene = item, item.Destination, true
		case "affair-victim":
			affairVictim = item
		case "affair-paramour":
			affairParamour = item
		case "extortion_payment":
			if item.Source == "bank" && item.NPCID == definition.VictimID && item.Direction == "outgoing" {
				victimPayment = item
			}
			if item.Source == "bank" && item.NPCID == definition.KillerID && item.Direction == "incoming" {
				killerReceipt = item
			}
		}
	}
	if !victimScene || !killerScene || victimGate == nil || killerGate == nil || scene == "" || scene != expectedScene || victimGate.Destination != expectedScene || killerGate.Destination != expectedScene || victimGate.OccurredAt.After(killerGate.OccurredAt) {
		return false
	}
	if definition.ArchetypeID == "financial_extortion" {
		if victimPayment == nil || killerReceipt == nil || victimPayment.Amount != killerReceipt.Amount || victimPayment.Reference == "" || victimPayment.Reference != killerReceipt.Reference || !victimPayment.OccurredAt.Equal(killerReceipt.OccurredAt) {
			return false
		}
	}
	if definition.ArchetypeID == "jealousy" {
		if affairVictim == nil || affairParamour == nil || affairVictim.Source != "gate" || affairParamour.Source != "gate" || affairVictim.NPCID != definition.VictimID || affairParamour.NPCID != definition.ParamourID || affairVictim.Destination != expectedScene || affairParamour.Destination != expectedScene || !affairVictim.OccurredAt.Equal(affairParamour.OccurredAt) {
			return false
		}
	}
	return true
}

func validateConnectedCaseEmails(definition Definition, base world.World) bool {
	remorse := false
	for _, message := range definition.Emails {
		if message.InternalRelation == "killer_remorse" && message.MailboxNPCID == definition.KillerID && message.RelatedNPCID != "" && message.Sender == npcByID(base.NPCs, definition.KillerID).Name {
			remorse = true
		}
	}
	return remorse
}

func regionName(base world.World, regionID string) string {
	for _, region := range base.Regions {
		if region.ID == regionID {
			return region.Name
		}
	}
	return "Lyon"
}

func hasEvidenceTypes(evidence []EvidenceDefinition, required ...string) bool {
	found := make(map[string]bool, len(evidence))
	for _, item := range evidence {
		found[item.Type] = true
	}
	for _, typeID := range required {
		if !found[typeID] {
			return false
		}
	}
	return true
}

func containsLocationInRegion(base world.World, locationID, regionID string) bool {
	for _, entity := range base.Entities {
		if entity.ID == locationID && entity.EntityType == "location" && entity.RegionID == regionID {
			return true
		}
	}
	return false
}

func hasReachableEvidence(evidence []EvidenceDefinition, definition Definition, relationships map[string]RelationshipReference, entities map[string]string) bool {
	if len(evidence) == 0 {
		return false
	}
	for _, clue := range evidence {
		if clue.ID == "" || clue.Type == "" || clue.Presentation == "" || clue.Content == "" || clue.LocationID != definition.LocationID || entities[clue.LocationID] != "location" {
			return false
		}
		if len(clue.RelationshipIDs) == 0 {
			return false
		}
		validLinks := true
		for _, relationshipID := range clue.RelationshipIDs {
			found := false
			for _, relationship := range relationships {
				if relationship.ID == relationshipID {
					found = true
					break
				}
			}
			if !found {
				validLinks = false
				break
			}
		}
		if !validLinks {
			return false
		}
		reachable := false
		for _, accessID := range clue.ReachableFromEntityIDs {
			if (accessID == definition.KillerID || accessID == definition.VictimID) && entities[accessID] == "npc" {
				reachable = true
				break
			}
		}
		if !reachable {
			return false
		}
	}
	return true
}

func hasHeadquarters(relations []world.CompanyLocation, companyID, locationID string) bool {
	for _, relation := range relations {
		if relation.CompanyID == companyID && relation.LocationID == locationID && relation.LocationRole == "headquarters" {
			return true
		}
	}
	return false
}

func containsNPC(npcs []world.NPC, id string) bool {
	for _, npc := range npcs {
		if npc.ID == id {
			return true
		}
	}
	return false
}

func contains(values []string, target string) bool {
	for _, value := range values {
		if value == target {
			return true
		}
	}
	return false
}

func caseIndex(seed string, version, count int) int {
	var versionBytes [8]byte
	binary.BigEndian.PutUint64(versionBytes[:], uint64(version))
	hash := sha256.New()
	writeField(hash, "noir-1974/case-generator")
	writeField(hash, string(versionBytes[:]))
	writeField(hash, seed)
	sum := hash.Sum(nil)
	return int(binary.BigEndian.Uint64(sum[:8]) % uint64(count))
}

func makeCaseID(seed string, version int) string {
	var versionBytes [8]byte
	binary.BigEndian.PutUint64(versionBytes[:], uint64(version))
	hash := sha256.New()
	writeField(hash, "noir-1974/case-id")
	writeField(hash, string(versionBytes[:]))
	writeField(hash, seed)
	return "case:" + hex.EncodeToString(hash.Sum(nil)[:12])
}

type byteWriter interface{ Write([]byte) (int, error) }

func writeField(writer byteWriter, value string) {
	var length [8]byte
	binary.BigEndian.PutUint64(length[:], uint64(len(value)))
	_, _ = writer.Write(length[:])
	_, _ = writer.Write([]byte(value))
}

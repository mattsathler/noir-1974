package cases

import (
	"crypto/sha256"
	"encoding/binary"
	"encoding/hex"
	"errors"
	"fmt"
	"sort"
	"strings"

	"noir-1974/apps/api/internal/world"
)

const GeneratorVersion = 2

var (
	ErrInvalidCaseSeed       = errors.New("case seed is required")
	ErrUnsupportedGenerator  = errors.New("unsupported case generator version")
	ErrNoValidCaseCandidates = errors.New("world has no valid homicide candidates")
	ErrInvalidDefinition     = errors.New("invalid case definition")
	ErrUnsolvableCase        = errors.New("case has no reachable clue")
)

type Definition struct {
	ID               string                  `json:"id"`
	Seed             string                  `json:"caseSeed"`
	GeneratorVersion int                     `json:"generatorVersion"`
	CaseType         string                  `json:"caseType"`
	ArchetypeID      string                  `json:"archetypeId"`
	Motive           string                  `json:"motive"`
	VictimID         string                  `json:"victimId"`
	KillerID         string                  `json:"killerId"`
	LocationID       string                  `json:"locationId"`
	Relationships    []RelationshipReference `json:"relationships"`
	Evidence         []EvidenceDefinition    `json:"evidence"`
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
}

// Generate creates a deterministic homicide definition from existing world facts.
func Generate(base world.World, seed string, version int) (Definition, error) {
	if strings.TrimSpace(seed) == "" {
		return Definition{}, ErrInvalidCaseSeed
	}
	if version != GeneratorVersion {
		return Definition{}, fmt.Errorf("%w: %d", ErrUnsupportedGenerator, version)
	}
	available := availableArchetypes(base)
	if len(available) == 0 {
		return Definition{}, ErrNoValidCaseCandidates
	}
	archetype := available[caseIndex(seed+"/archetype", version, len(available))]
	candidates := archetype.candidates(base)
	selected := candidates[caseIndex(seed+"/"+archetype.id, version, len(candidates))]
	caseID := makeCaseID(seed, version)
	definition := Definition{
		ID:               caseID,
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
	if err := Validate(definition, base); err != nil {
		return Definition{}, err
	}
	return definition, nil
}

// Validate checks the case archetype, world references, and reachable clues.
func Validate(definition Definition, base world.World) error {
	if base.GeneratorVersion != world.GeneratorVersion2 {
		return ErrInvalidDefinition
	}
	if strings.TrimSpace(definition.Seed) == "" || definition.GeneratorVersion != GeneratorVersion || definition.ID != makeCaseID(definition.Seed, definition.GeneratorVersion) || definition.CaseType != "homicide" || definition.ArchetypeID == "" || definition.Motive == "" {
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
	if !exists || !archetype.validate(definition, base) {
		return ErrInvalidDefinition
	}
	if !hasReachableEvidence(definition.Evidence, definition, links, entities) || !hasEvidenceTypes(definition.Evidence, archetype.evidenceTypes...) {
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
	if base.GeneratorVersion != world.GeneratorVersion2 {
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

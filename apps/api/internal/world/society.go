package world

import (
	"crypto/sha256"
	"encoding/hex"
	"errors"
	"fmt"
	"sort"
	"strconv"
	"strings"
)

var ErrInvalidGenerationKey = errors.New("invalid generated entity key")

type NPC struct {
	ID            string `json:"id"`
	GenerationKey string `json:"generationKey"`
	RegionID      string `json:"regionId"`
	Name          string `json:"name"`
	Age           int    `json:"age"`
	Appearance    string `json:"appearance"`
	Personality   string `json:"personality"`
	Status        string `json:"status"`
}

type Company struct {
	ID            string `json:"id"`
	GenerationKey string `json:"generationKey"`
	RegionID      string `json:"regionId"`
	Name          string `json:"name"`
}

type EmploymentRelationship struct {
	ID                string `json:"id"`
	EmployeeNPCID     string `json:"employeeNpcId"`
	EmployerCompanyID string `json:"employerCompanyId"`
	Role              string `json:"role"`
}

type PersonalRelationship struct {
	ID               string `json:"id"`
	PersonANPCID     string `json:"personANpcId"`
	PersonBNPCID     string `json:"personBNpcId"`
	RelationshipType string `json:"relationshipType"`
}

type CompanyLocation struct {
	ID           string `json:"id"`
	CompanyID    string `json:"companyId"`
	LocationID   string `json:"locationId"`
	LocationRole string `json:"locationRole"`
}

type personName struct {
	first string
	last  string
}

var firstNames = []string{
	"Alain", "André", "Bernard", "Brigitte", "Catherine", "Chantal", "Claude", "Dominique",
	"Françoise", "Isabelle", "Jean", "Louis", "Marie", "Michel", "Monique", "Pierre",
}

var lastNames = []string{
	"Bernard", "Bertrand", "David", "Dubois", "Durand", "Laurent", "Lefebvre", "Leroy",
	"Michel", "Moreau", "Morel", "Petit", "Richard", "Robert", "Roux", "Simon",
}

var appearances = []string{
	"dark wavy hair, brown eyes",
	"short auburn hair, green eyes",
	"black hair, round glasses",
	"shoulder-length brown hair, hazel eyes",
	"fair hair, blue eyes",
	"curly dark hair, a small scar on the chin",
	"silvering hair, heavy eyebrows",
	"close-cropped hair, a narrow moustache",
}

var personalities = []string{
	"patient and observant",
	"charming but guarded",
	"methodical and reserved",
	"quick-witted and impatient",
	"generous with a stubborn streak",
	"ambitious and watchful",
	"soft-spoken and perceptive",
	"outgoing but easily distracted",
}

var companyNames = []string{
	"Aurore Industries",
	"Cendre Group",
	"Delacroix Holdings",
	"Étoile Conglomerate",
	"Fournier & Fils",
	"Lumière Corporation",
	"Montclair Industries",
	"Rhône Commercial Group",
}

var employmentRoles = []string{
	"accountant", "clerk", "driver", "manager", "messenger", "receptionist", "secretary", "typist",
}

// GenerateNPC reconstructs one NPC directly from its stable generation key.
func GenerateNPC(seed string, version int, generationKey string) (NPC, error) {
	if strings.TrimSpace(seed) == "" {
		return NPC{}, errors.New("world seed is required")
	}
	if version != GeneratorVersion2 {
		return NPC{}, fmt.Errorf("%w: %d", ErrUnsupportedGeneratorVersion, version)
	}

	parts := strings.Split(generationKey, "/")
	if len(parts) != 7 || parts[0] != "lyon" || parts[1] != "people" || parts[3] != "district" || parts[5] != "person" || parts[2] != seedFingerprint(seed, version) {
		return NPC{}, ErrInvalidGenerationKey
	}
	slot, err := strconv.Atoi(parts[6])
	if err != nil || slot < 1 || slot > 8 {
		return NPC{}, ErrInvalidGenerationKey
	}

	districts := selectedDistricts(seed, version)
	districtIndex := -1
	for index, district := range districts {
		if district.slug == parts[4] {
			districtIndex = index
			break
		}
	}
	if districtIndex < 0 {
		return NPC{}, ErrInvalidGenerationKey
	}
	nameIndex := districtIndex*8 + slot - 1
	names := shuffledCopy(allPersonNames(), newStream(seed, version, "people/names"))
	if nameIndex >= len(names) {
		return NPC{}, ErrInvalidGenerationKey
	}
	regionID := "fr/lyon/district/" + parts[4]
	return npcFor(seed, version, regionID, generationKey, names[nameIndex]), nil
}

// GenerateCompany reconstructs one company directly from its stable generation key.
func GenerateCompany(seed string, version int, generationKey string) (Company, error) {
	if strings.TrimSpace(seed) == "" {
		return Company{}, errors.New("world seed is required")
	}
	if version != GeneratorVersion2 {
		return Company{}, fmt.Errorf("%w: %d", ErrUnsupportedGeneratorVersion, version)
	}

	parts := strings.Split(generationKey, "/")
	if len(parts) != 6 || parts[0] != "lyon" || parts[1] != "companies" || parts[2] != seedFingerprint(seed, version) || parts[3] != "district" || parts[5] != "company" {
		return Company{}, ErrInvalidGenerationKey
	}
	districts := selectedDistricts(seed, version)
	districtIndex := -1
	for index, district := range districts {
		if district.slug == parts[4] {
			districtIndex = index
			break
		}
	}
	if districtIndex < 0 {
		return Company{}, ErrInvalidGenerationKey
	}
	names := shuffledCopy(companyNames, newStream(seed, version, "company/names"))
	if districtIndex >= len(names) {
		return Company{}, ErrInvalidGenerationKey
	}
	return Company{
		ID:            "company:" + generationKey,
		GenerationKey: generationKey,
		RegionID:      "fr/lyon/district/" + parts[4],
		Name:          names[districtIndex],
	}, nil
}

func generateSociety(seed string) (World, error) {
	world := generateCity(seed, GeneratorVersion2)
	districts := selectedDistricts(seed, GeneratorVersion2)
	world.NPCs = make([]NPC, 0, len(districts)*8)
	world.Companies = make([]Company, 0, len(districts))
	world.EmploymentRelationships = make([]EmploymentRelationship, 0, len(districts)*8)
	world.PersonalRelationships = make([]PersonalRelationship, 0, len(districts)*4)
	world.CompanyLocations = make([]CompanyLocation, 0, len(districts))

	for _, district := range districts {
		regionID := "fr/lyon/district/" + district.slug
		companyKey := "lyon/companies/" + seedFingerprint(seed, GeneratorVersion2) + "/district/" + district.slug + "/company"
		company, err := GenerateCompany(seed, GeneratorVersion2, companyKey)
		if err != nil {
			return World{}, err
		}
		world.Companies = append(world.Companies, company)
		world.Entities = append(world.Entities, Entity{
			ID:            company.ID,
			EntityType:    "company",
			RegionID:      regionID,
			GenerationKey: company.GenerationKey,
			Name:          company.Name,
		})

		var headquartersID string
		for _, entity := range world.Entities {
			if entity.RegionID == regionID && entity.EntityType == "location" && entity.LocationKind == "building" {
				headquartersID = entity.ID
				break
			}
		}
		if headquartersID == "" {
			return World{}, errors.New("generated district has no headquarters location")
		}
		relationshipID := "company-location:" + company.ID + ":" + headquartersID
		world.CompanyLocations = append(world.CompanyLocations, CompanyLocation{
			ID:           relationshipID,
			CompanyID:    company.ID,
			LocationID:   headquartersID,
			LocationRole: "headquarters",
		})

		peopleBySlot := make([]NPC, 8)
		for slot := 1; slot <= len(peopleBySlot); slot++ {
			personKey := fmt.Sprintf("lyon/people/%s/district/%s/person/%02d", seedFingerprint(seed, GeneratorVersion2), district.slug, slot)
			person, err := GenerateNPC(seed, GeneratorVersion2, personKey)
			if err != nil {
				return World{}, err
			}
			peopleBySlot[slot-1] = person
			world.NPCs = append(world.NPCs, person)
			world.Entities = append(world.Entities, Entity{
				ID:            person.ID,
				EntityType:    "npc",
				RegionID:      regionID,
				GenerationKey: person.GenerationKey,
				Name:          person.Name,
			})
			world.EmploymentRelationships = append(world.EmploymentRelationships, EmploymentRelationship{
				ID:                "employment:" + person.ID + ":" + company.ID,
				EmployeeNPCID:     person.ID,
				EmployerCompanyID: company.ID,
				Role:              employmentRoles[slot-1],
			})
		}

		for slot := 0; slot < len(peopleBySlot); slot += 2 {
			personAID, personBID := peopleBySlot[slot].ID, peopleBySlot[slot+1].ID
			if personBID < personAID {
				personAID, personBID = personBID, personAID
			}
			world.PersonalRelationships = append(world.PersonalRelationships, PersonalRelationship{
				ID:               "personal:marriage:" + personAID + ":" + personBID,
				PersonANPCID:     personAID,
				PersonBNPCID:     personBID,
				RelationshipType: "marriage",
			})
		}
	}

	sort.Slice(world.Entities, func(i, j int) bool { return world.Entities[i].ID < world.Entities[j].ID })
	sort.Slice(world.NPCs, func(i, j int) bool { return world.NPCs[i].ID < world.NPCs[j].ID })
	sort.Slice(world.Companies, func(i, j int) bool { return world.Companies[i].ID < world.Companies[j].ID })
	sort.Slice(world.EmploymentRelationships, func(i, j int) bool {
		return world.EmploymentRelationships[i].ID < world.EmploymentRelationships[j].ID
	})
	sort.Slice(world.PersonalRelationships, func(i, j int) bool {
		return world.PersonalRelationships[i].ID < world.PersonalRelationships[j].ID
	})
	sort.Slice(world.CompanyLocations, func(i, j int) bool {
		return world.CompanyLocations[i].ID < world.CompanyLocations[j].ID
	})
	return world, nil
}

func npcFor(seed string, version int, regionID, generationKey string, name personName) NPC {
	stream := newStream(seed, version, "npc/attributes/"+generationKey)
	return NPC{
		ID:            "npc:" + generationKey,
		GenerationKey: generationKey,
		RegionID:      regionID,
		Name:          name.first + " " + name.last,
		Age:           18 + stream.intn(53),
		Appearance:    appearances[stream.intn(len(appearances))],
		Personality:   personalities[stream.intn(len(personalities))],
		Status:        "alive",
	}
}

func allPersonNames() []personName {
	result := make([]personName, 0, len(firstNames)*len(lastNames))
	for _, first := range firstNames {
		for _, last := range lastNames {
			result = append(result, personName{first: first, last: last})
		}
	}
	return result
}

func selectedDistricts(seed string, version int) []namedPlace {
	districts := shuffledCopy(lyonDistricts, newStream(seed, version, "city/districts"))[:3]
	sort.Slice(districts, func(i, j int) bool { return districts[i].slug < districts[j].slug })
	return districts
}

func seedFingerprint(seed string, version int) string {
	h := sha256.New()
	writeField(h, "noir-1974/entity-namespace")
	writeField(h, strconv.Itoa(version))
	writeField(h, seed)
	return hex.EncodeToString(h.Sum(nil))
}

package world

import (
	"crypto/sha256"
	"encoding/hex"
	"errors"
	"fmt"
	"sort"
	"strconv"
	"strings"
	"sync"

	"noir-1974/apps/api/internal/world/names"
)

var ErrInvalidGenerationKey = errors.New("invalid generated entity key")

var personNameOrderCache sync.Map

type NPC struct {
	ID            string          `json:"id"`
	GenerationKey string          `json:"generationKey"`
	RegionID      string          `json:"regionId"`
	Name          string          `json:"name"`
	Age           int             `json:"age"`
	Appearance    string          `json:"appearance"`
	Visual        *VisualIdentity `json:"visual,omitempty"`
	Personality   string          `json:"personality"`
	Status        string          `json:"status"`
	MaritalStatus string          `json:"-"`
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
	if !SupportsSocietyVersion(version) {
		return NPC{}, fmt.Errorf("%w: %d", ErrUnsupportedGeneratorVersion, version)
	}

	parts := strings.Split(generationKey, "/")
	countrySlug, citySlug, districtSlug := "fr", "lyon", ""
	cityID := "fr/lyon"
	citySeed := seed
	var slotPart string
	switch version {
	case GeneratorVersion2:
		if len(parts) != 7 || parts[0] != "lyon" || parts[1] != "people" || parts[3] != "district" || parts[5] != "person" || parts[2] != seedFingerprint(seed, version) {
			return NPC{}, ErrInvalidGenerationKey
		}
		districtSlug, slotPart = parts[4], parts[6]
	case GeneratorVersion3, GeneratorVersion4, GeneratorVersion5:
		if len(parts) != 8 || parts[2] != "people" || parts[4] != "district" || parts[6] != "person" {
			return NPC{}, ErrInvalidGenerationKey
		}
		countrySlug, citySlug = parts[0], parts[1]
		cityID = countrySlug + "/" + citySlug
		city, ok := CityByID(cityID)
		if !ok {
			return NPC{}, ErrInvalidGenerationKey
		}
		citySeed = CitySeed(seed, city.ID)
		if parts[3] != seedFingerprint(citySeed, version) {
			return NPC{}, ErrInvalidGenerationKey
		}
		districtSlug, slotPart = parts[5], parts[7]
	default:
		return NPC{}, ErrUnsupportedGeneratorVersion
	}
	slot, err := strconv.Atoi(slotPart)
	if err != nil || slot < 1 {
		return NPC{}, ErrInvalidGenerationKey
	}

	city, _ := CityByID(cityID)
	districts := districtsForCity(citySeed, version, city)
	districtIndex := -1
	for index, district := range districts {
		if district.slug == districtSlug {
			districtIndex = index
			break
		}
	}
	if districtIndex < 0 {
		return NPC{}, ErrInvalidGenerationKey
	}
	personNames, err := personNamesForVersion(version, countrySlug, citySlug)
	if err != nil {
		return NPC{}, err
	}
	nameIndex := districtIndex*8 + slot - 1
	if slot > 8 {
		nameIndex = len(districts)*8 + (slot-9)*len(districts) + districtIndex
	}
	cacheKey := fmt.Sprintf("%d:%s:%s", version, citySeed, personNameScope(version, countrySlug, citySlug))
	namesValue, ok := personNameOrderCache.Load(cacheKey)
	if !ok {
		namesValue, _ = personNameOrderCache.LoadOrStore(cacheKey, shuffledCopy(personNames, newStream(citySeed, version, personNameScope(version, countrySlug, citySlug))))
	}
	names := namesValue.([]personName)
	if len(names) == 0 {
		return NPC{}, ErrInvalidGenerationKey
	}
	nameVariant := nameIndex / len(names)
	name := names[nameIndex%len(names)]
	if nameVariant > 0 {
		middleNames := []string{"Arthur", "Beatrice", "Claude", "Diane", "Edwin", "Flora", "Gideon", "Helen", "Isaac", "June", "Laurent", "Margot", "Nolan", "Odette", "Pierre", "Ruth"}
		middle := middleNames[(nameVariant-1)%len(middleNames)]
		if nameVariant > len(middleNames) {
			middle = fmt.Sprintf("%s %d", middle, (nameVariant-1)/len(middleNames)+1)
		}
		name.first += " " + middle
	}
	regionID := cityID + "/district/" + districtSlug
	return npcFor(citySeed, version, regionID, generationKey, name), nil
}

// GenerateNPCFromKey chooses the city generator version embedded in a modern
// city-scoped key, falling back to the server's legacy version for Lyon.
func GenerateNPCFromKey(seed string, fallbackVersion int, generationKey string) (NPC, error) {
	parts := strings.Split(generationKey, "/")
	if len(parts) == 8 {
		cityID := parts[0] + "/" + parts[1]
		if city, ok := CityByID(cityID); ok && city.ID != "fr/lyon" {
			citySeed := CitySeed(seed, city.ID)
			for version := GeneratorVersion5; version >= GeneratorVersion3; version-- {
				if parts[3] == seedFingerprint(citySeed, version) {
					return GenerateNPC(seed, version, generationKey)
				}
			}
		}
	}
	return GenerateNPC(seed, fallbackVersion, generationKey)
}

// GenerateNPCPair creates two deterministic new residents in an existing
// district and connects them to that district's company and to each other.
func GenerateNPCPair(seed string, version int, districtID string, firstSlot int, company Company) ([]NPC, []EmploymentRelationship, PersonalRelationship, error) {
	if firstSlot < 9 || company.ID == "" || company.RegionID != districtID {
		return nil, nil, PersonalRelationship{}, ErrInvalidGenerationKey
	}
	city, citySeed, districtSlug, ok := cityForDistrict(seed, districtID)
	if !ok || strings.Contains(districtSlug, "/") {
		return nil, nil, PersonalRelationship{}, ErrInvalidGenerationKey
	}
	npcs := make([]NPC, 0, 2)
	employment := make([]EmploymentRelationship, 0, 2)
	for slot := firstSlot; slot < firstSlot+2; slot++ {
		key := npcGenerationKeyForCity(citySeed, version, city, districtSlug, slot)
		npc, err := GenerateNPC(seed, version, key)
		if err != nil {
			return nil, nil, PersonalRelationship{}, err
		}
		npcs = append(npcs, npc)
		employment = append(employment, EmploymentRelationship{
			ID:                "employment:" + npc.ID + ":" + company.ID,
			EmployeeNPCID:     npc.ID,
			EmployerCompanyID: company.ID,
			Role:              employmentRoles[(slot-1)%len(employmentRoles)],
		})
	}
	personAID, personBID := npcs[0].ID, npcs[1].ID
	if personBID < personAID {
		personAID, personBID = personBID, personAID
	}
	marriage := PersonalRelationship{
		ID:               "personal:marriage:" + personAID + ":" + personBID,
		PersonANPCID:     personAID,
		PersonBNPCID:     personBID,
		RelationshipType: "marriage",
	}
	return npcs, employment, marriage, nil
}

func npcGenerationKey(seed string, version int, districtSlug string, slot int) string {
	fingerprint := seedFingerprint(seed, version)
	if version == GeneratorVersion2 {
		return fmt.Sprintf("lyon/people/%s/district/%s/person/%02d", fingerprint, districtSlug, slot)
	}
	return fmt.Sprintf("fr/lyon/people/%s/district/%s/person/%02d", fingerprint, districtSlug, slot)
}

func npcGenerationKeyForCity(seed string, version int, city WorldCity, districtSlug string, slot int) string {
	fingerprint := seedFingerprint(seed, version)
	if version == GeneratorVersion2 {
		return fmt.Sprintf("lyon/people/%s/district/%s/person/%02d", fingerprint, districtSlug, slot)
	}
	return fmt.Sprintf("%s/people/%s/district/%s/person/%02d", city.ID, fingerprint, districtSlug, slot)
}

func cityForDistrict(seed, districtID string) (WorldCity, string, string, bool) {
	for _, city := range Cities() {
		prefix := city.ID + "/district/"
		if strings.HasPrefix(districtID, prefix) {
			return city, CitySeed(seed, city.ID), strings.TrimPrefix(districtID, prefix), true
		}
	}
	return WorldCity{}, "", "", false
}

func districtsForCity(seed string, version int, city WorldCity) []namedPlace {
	if city.ID == "fr/lyon" {
		return selectedDistricts(seed, version)
	}
	cityDistricts := CityDistricts(seed, city)
	result := make([]namedPlace, 0, len(cityDistricts))
	for _, district := range cityDistricts {
		result = append(result, namedPlace{slug: strings.TrimPrefix(district.ID, city.ID+"/district/"), name: district.Name})
	}
	return result
}

// GenerateCompany reconstructs one company directly from its stable generation key.
func GenerateCompany(seed string, version int, generationKey string) (Company, error) {
	if strings.TrimSpace(seed) == "" {
		return Company{}, errors.New("world seed is required")
	}
	if !SupportsSocietyVersion(version) {
		return Company{}, fmt.Errorf("%w: %d", ErrUnsupportedGeneratorVersion, version)
	}

	parts := strings.Split(generationKey, "/")
	districtSlug := ""
	switch version {
	case GeneratorVersion2:
		if len(parts) != 6 || parts[0] != "lyon" || parts[1] != "companies" || parts[2] != seedFingerprint(seed, version) || parts[3] != "district" || parts[5] != "company" {
			return Company{}, ErrInvalidGenerationKey
		}
		districtSlug = parts[4]
	case GeneratorVersion3, GeneratorVersion4, GeneratorVersion5:
		if len(parts) != 7 || parts[2] != "companies" || parts[4] != "district" || parts[6] != "company" {
			return Company{}, ErrInvalidGenerationKey
		}
		cityID := parts[0] + "/" + parts[1]
		city, ok := CityByID(cityID)
		if !ok {
			return Company{}, ErrInvalidGenerationKey
		}
		seed = CitySeed(seed, city.ID)
		if parts[3] != seedFingerprint(seed, version) {
			return Company{}, ErrInvalidGenerationKey
		}
		districtSlug = parts[5]
	default:
		return Company{}, ErrUnsupportedGeneratorVersion
	}
	cityID := "fr/lyon"
	if version >= GeneratorVersion3 {
		cityID = parts[0] + "/" + parts[1]
	}
	city, _ := CityByID(cityID)
	districts := districtsForCity(seed, version, city)
	districtIndex := -1
	for index, district := range districts {
		if district.slug == districtSlug {
			districtIndex = index
			break
		}
	}
	if districtIndex < 0 {
		return Company{}, ErrInvalidGenerationKey
	}
	names := shuffledCopy(companyNamesForVersion(version), newStream(seed, version, "company/names"))
	if districtIndex >= len(names) {
		return Company{}, ErrInvalidGenerationKey
	}
	return Company{
		ID:            "company:" + generationKey,
		GenerationKey: generationKey,
		RegionID:      city.ID + "/district/" + districtSlug,
		Name:          names[districtIndex],
	}, nil
}

// GenerateCompanyFromKey mirrors GenerateNPCFromKey for a city-scoped company.
func GenerateCompanyFromKey(seed string, fallbackVersion int, generationKey string) (Company, error) {
	parts := strings.Split(generationKey, "/")
	if len(parts) == 7 {
		cityID := parts[0] + "/" + parts[1]
		if city, ok := CityByID(cityID); ok && city.ID != "fr/lyon" {
			citySeed := CitySeed(seed, city.ID)
			for version := GeneratorVersion5; version >= GeneratorVersion3; version-- {
				if parts[3] == seedFingerprint(citySeed, version) {
					return GenerateCompany(seed, version, generationKey)
				}
			}
		}
	}
	return GenerateCompany(seed, fallbackVersion, generationKey)
}

func generateSociety(seed string, version int) (World, error) {
	world := generateCity(seed, version)
	districts := selectedDistricts(seed, version)
	world.NPCs = make([]NPC, 0, len(districts)*8)
	world.Companies = make([]Company, 0, len(districts))
	world.EmploymentRelationships = make([]EmploymentRelationship, 0, len(districts)*8)
	world.PersonalRelationships = make([]PersonalRelationship, 0, len(districts)*4)
	world.CompanyLocations = make([]CompanyLocation, 0, len(districts))

	for _, district := range districts {
		regionID := "fr/lyon/district/" + district.slug
		companyKey := "lyon/companies/" + seedFingerprint(seed, version) + "/district/" + district.slug + "/company"
		if version >= GeneratorVersion3 {
			companyKey = "fr/lyon/companies/" + seedFingerprint(seed, version) + "/district/" + district.slug + "/company"
		}
		company, err := GenerateCompany(seed, version, companyKey)
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
			personKey := fmt.Sprintf("lyon/people/%s/district/%s/person/%02d", seedFingerprint(seed, version), district.slug, slot)
			if version >= GeneratorVersion3 {
				personKey = fmt.Sprintf("fr/lyon/people/%s/district/%s/person/%02d", seedFingerprint(seed, version), district.slug, slot)
			}
			person, err := GenerateNPC(seed, version, personKey)
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
	maritalStream := newStream(seed, version, "npc/marital-status/"+generationKey)
	maritalStatus := "single"
	switch maritalStream.intn(10) {
	case 0:
		maritalStatus = "widowed"
	case 1, 2:
		maritalStatus = "divorced"
	}
	npc := NPC{
		ID:            "npc:" + generationKey,
		GenerationKey: generationKey,
		RegionID:      regionID,
		Name:          name.first + " " + name.last,
		Age:           18 + stream.intn(53),
		Appearance:    appearances[stream.intn(len(appearances))],
		Personality:   personalities[stream.intn(len(personalities))],
		Status:        "alive",
		MaritalStatus: maritalStatus,
	}
	if version >= GeneratorVersion4 {
		visual := visualIdentityForNPC(seed, version, generationKey, npc.Age)
		npc.Visual = &visual
		npc.Appearance = generatedAppearanceDescription(visual)
	}
	return npc
}

func visualIdentityForNPC(seed string, version int, generationKey string, age int) VisualIdentity {
	if version == GeneratorVersion4 {
		return legacyVisualIdentityForNPC(seed, generationKey)
	}
	stream := newStream(seed, version, "npc/visual/"+generationKey)
	hairStream := newStream(seed, version, "npc/visual/hair/"+generationKey)
	identity := VisualIdentity{
		HairID:       generatedHairID(hairStream, age),
		SkinID:       skinLayerIDs[stream.intn(len(skinLayerIDs))],
		EyesID:       eyeLayerIDs[stream.intn(len(eyeLayerIDs))],
		AccessoryIDs: []string{},
	}
	if stream.intn(4) == 0 {
		styles := []string{"moustache", "mutton", "soul_patch", "goatee", "full"}
		identity.BeardID = "beard_" + styles[stream.intn(len(styles))] + "_" + hairTone(identity.HairID)
	}
	if stream.intn(8) == 0 {
		identity.FaceID = "face_scar_short"
	}
	if stream.intn(5) == 0 {
		identity.AccessoryIDs = append(identity.AccessoryIDs, accessoryLayerIDs[stream.intn(len(accessoryLayerIDs))])
	}
	return identity
}

func legacyVisualIdentityForNPC(seed, generationKey string) VisualIdentity {
	stream := newStream(seed, GeneratorVersion4, "npc/visual/"+generationKey)
	identity := VisualIdentity{
		HairID:       hairLayerIDs[stream.intn(len(hairLayerIDs))],
		SkinID:       skinLayerIDs[stream.intn(len(skinLayerIDs))],
		EyesID:       eyeLayerIDs[stream.intn(len(eyeLayerIDs))],
		AccessoryIDs: []string{},
	}
	if stream.intn(4) == 0 {
		styles := []string{"moustache", "mutton", "soul_patch", "goatee", "full"}
		identity.BeardID = "beard_" + styles[stream.intn(len(styles))] + "_" + hairTone(identity.HairID)
	}
	if stream.intn(8) == 0 {
		identity.FaceID = "face_scar_short"
	}
	if stream.intn(5) == 0 {
		identity.AccessoryIDs = append(identity.AccessoryIDs, accessoryLayerIDs[stream.intn(len(accessoryLayerIDs))])
	}
	return identity
}

func generatedHairID(stream randomStream, age int) string {
	styles := []string{"short", "long", "long_wavy", "long_curly", "messy", "wavy", "bob", "slicked", "mohawk"}
	if age >= 45 {
		styles = append(styles, "balding", "sparse_sides")
	}
	style := styles[stream.intn(len(styles))]

	tones := []string{"black", "brown", "ginger", "blonde"}
	greyChance := 0
	switch {
	case age >= 60:
		greyChance = 40
	case age >= 50:
		greyChance = 20
	case age >= 40:
		greyChance = 5
	}
	if greyChance > 0 && stream.intn(100) < greyChance {
		return "hair_" + style + "_grey"
	}
	return "hair_" + style + "_" + tones[stream.intn(len(tones))]
}

func generatedAppearanceDescription(identity VisualIdentity) string {
	hairColor := hairTone(identity.HairID)
	hairStyle := hairStyle(identity.HairID)
	skinTone := strings.TrimPrefix(identity.SkinID, "skin_")
	eyeColor := strings.TrimPrefix(identity.EyesID, "eyes_")
	description := []string{
		hairColor + " " + hairStyle + " hair",
		strings.ReplaceAll(skinTone, "_", " ") + " skin",
		strings.ReplaceAll(eyeColor, "_", " ") + " eyes",
	}
	if beard := beardDescription(identity.BeardID); beard != "" {
		description = append(description, beard)
	}
	if face := faceDescription(identity.FaceID); face != "" {
		description = append(description, face)
	}
	for _, accessoryID := range identity.AccessoryIDs {
		if accessory := accessoryDescription(accessoryID); accessory != "" {
			description = append(description, accessory)
		}
	}
	return strings.Join(description, ", ")
}

func hairStyle(hairID string) string {
	style := strings.TrimPrefix(hairID, "hair_")
	style = strings.TrimSuffix(style, "_"+hairTone(hairID))
	switch style {
	case "short":
		return "short"
	case "short_alt":
		return "swept short"
	case "long":
		return "long"
	case "long_wavy":
		return "long wavy"
	case "long_curly":
		return "long curly"
	case "wavy":
		return "wavy"
	case "bob":
		return "bob"
	case "slicked":
		return "slicked-back"
	case "mohawk":
		return "mohawk"
	case "balding":
		return "balding"
	case "sparse_sides":
		return "sparse-sided"
	case "ponytail":
		return "ponytail"
	case "tied":
		return "tied-back"
	case "tied_long":
		return "long tied-back"
	case "messy":
		return "messy"
	case "pigtails":
		return "pigtails"
	case "emo":
		return "emo"
	default:
		return "short"
	}
}

func beardDescription(id string) string {
	switch {
	case id == "":
		return ""
	case strings.HasPrefix(id, "beard_moustache_"):
		return "a moustache"
	case strings.HasPrefix(id, "beard_mutton_"):
		return "mutton-chop sideburns"
	case strings.HasPrefix(id, "beard_soul_patch_"):
		return "a soul patch"
	case strings.HasPrefix(id, "beard_goatee_"):
		return "a goatee"
	case strings.HasPrefix(id, "beard_full_"):
		return "a full beard"
	default:
		return ""
	}
}

func faceDescription(id string) string {
	if id == "face_scar_short" {
		return "a short scar"
	}
	return ""
}

func accessoryDescription(id string) string {
	switch {
	case strings.HasPrefix(id, "accessories_glasses_round_"):
		return "round glasses"
	case strings.HasPrefix(id, "accessories_glasses_squared_"):
		return "squared glasses"
	case strings.HasPrefix(id, "accessories_monocle_"):
		return "a monocle"
	default:
		return ""
	}
}

func hairTone(hairID string) string {
	parts := strings.Split(hairID, "_")
	if len(parts) == 0 {
		return "black"
	}
	switch parts[len(parts)-1] {
	case "black":
		return "black"
	case "ginger":
		return "ginger"
	case "blonde":
		return "blonde"
	case "brown":
		return "brown"
	case "grey":
		return "grey"
	default:
		return "black"
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

func personNamesForVersion(version int, countrySlug, citySlug string) ([]personName, error) {
	if version == GeneratorVersion2 {
		return allPersonNames(), nil
	}
	if version != GeneratorVersion3 && version != GeneratorVersion4 && version != GeneratorVersion5 {
		return nil, ErrUnsupportedGeneratorVersion
	}
	catalog, ok := names.ForCity(countrySlug, citySlug)
	if !ok || len(catalog.GivenNames) == 0 || len(catalog.FamilyNames) == 0 {
		return allPersonNames(), nil
	}
	result := make([]personName, 0, len(catalog.GivenNames)*len(catalog.FamilyNames))
	for _, first := range catalog.GivenNames {
		for _, last := range catalog.FamilyNames {
			result = append(result, personName{first: first, last: last})
		}
	}
	return result, nil
}

func personNameScope(version int, countrySlug, citySlug string) string {
	if version == GeneratorVersion2 {
		return "people/names"
	}
	return "people/names/" + countrySlug + "/" + citySlug
}

func selectedDistricts(seed string, version int) []namedPlace {
	if version == GeneratorVersion5 {
		catalog, ok := names.ForCity("fr", "lyon")
		if !ok || len(catalog.DistrictPrefixes)*len(catalog.DistrictSuffixes) < 20 {
			return nil
		}
		candidates := make([]namedPlace, 0, len(catalog.DistrictPrefixes)*len(catalog.DistrictSuffixes))
		slugify := strings.NewReplacer(" ", "-", "é", "e", "è", "e", "ê", "e", "ô", "o", "û", "u", "ç", "c", "'", "")
		for _, prefix := range catalog.DistrictPrefixes {
			for _, suffix := range catalog.DistrictSuffixes {
				name := prefix + " " + suffix
				slug := strings.ToLower(slugify.Replace(name))
				candidates = append(candidates, namedPlace{slug: slug, name: name})
			}
		}
		selected := shuffledCopy(candidates, newStream(seed, version, "city/districts"))[:20]
		sort.Slice(selected, func(i, j int) bool { return selected[i].slug < selected[j].slug })
		return selected
	}
	districts := shuffledCopy(lyonDistricts, newStream(seed, version, "city/districts"))[:3]
	sort.Slice(districts, func(i, j int) bool { return districts[i].slug < districts[j].slug })
	return districts
}

var expandedCompanyNames = []string{
	"Aurore Industries", "Cendre Group", "Delacroix Holdings", "Étoile Conglomerate", "Fournier & Fils",
	"Lumière Corporation", "Montclair Industries", "Rhône Commercial Group", "Bastion Mutual", "Civic Assurance",
	"Continental Relay Company", "Dargent Shipping", "Eastbank Provisions", "Fallow & March", "Garnier Foundry",
	"Hearthstone Exchange", "Imperial Ledger Company", "Juno Telegraph Works", "Kingswell Security", "Lanterne Properties",
}

func companyNamesForVersion(version int) []string {
	if version >= GeneratorVersion5 {
		return expandedCompanyNames
	}
	return companyNames
}

func seedFingerprint(seed string, version int) string {
	h := sha256.New()
	writeField(h, "noir-1974/entity-namespace")
	writeField(h, strconv.Itoa(version))
	writeField(h, seed)
	return hex.EncodeToString(h.Sum(nil))
}

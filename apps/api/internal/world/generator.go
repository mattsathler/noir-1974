package world

import (
	"crypto/sha256"
	"encoding/binary"
	"errors"
	"fmt"
	"hash"
	"sort"
	"strconv"
	"strings"
)

const (
	GeneratorVersion1      = 1
	GeneratorVersion2      = 2
	GeneratorVersion3      = 3
	GeneratorVersion4      = 4
	GeneratorVersion5      = 5
	GeneratorVersion6      = 6
	GeneratorVersion7      = 7
	GeneratorVersion8      = 8
	LatestGeneratorVersion = GeneratorVersion8
)

var ErrUnsupportedGeneratorVersion = errors.New("unsupported world generator version")

type Region struct {
	ID       string `json:"id"`
	ParentID string `json:"parentRegionId,omitempty"`
	Kind     string `json:"kind"`
	Name     string `json:"name"`
}

type Entity struct {
	ID            string `json:"id"`
	EntityType    string `json:"entityType"`
	RegionID      string `json:"regionId"`
	GenerationKey string `json:"generationKey"`
	LocationKind  string `json:"locationKind"`
	Name          string `json:"name"`
}

type World struct {
	GeneratorVersion        int                      `json:"generatorVersion"`
	Regions                 []Region                 `json:"regions"`
	Entities                []Entity                 `json:"entities"`
	NPCs                    []NPC                    `json:"npcs,omitempty"`
	Companies               []Company                `json:"companies,omitempty"`
	EmploymentRelationships []EmploymentRelationship `json:"employmentRelationships,omitempty"`
	PersonalRelationships   []PersonalRelationship   `json:"personalRelationships,omitempty"`
	CompanyLocations        []CompanyLocation        `json:"companyLocations,omitempty"`
}

type namedPlace struct {
	slug string
	name string
}

var lyonDistricts = []namedPlace{
	{slug: "bellecour", name: "Bellecour"},
	{slug: "croix-rousse", name: "Croix-Rousse"},
	{slug: "guillotiere", name: "Guillotière"},
	{slug: "part-dieu", name: "Part-Dieu"},
	{slug: "perrache", name: "Perrache"},
	{slug: "vieux-lyon", name: "Vieux Lyon"},
}

var lyonStreets = []namedPlace{
	{slug: "cours-lafayette", name: "Cours Lafayette"},
	{slug: "rue-merciere", name: "Rue Mercière"},
	{slug: "rue-paul-chenavard", name: "Rue Paul Chenavard"},
	{slug: "rue-de-la-republique", name: "Rue de la République"},
	{slug: "rue-victor-hugo", name: "Rue Victor Hugo"},
	{slug: "rue-des-capucins", name: "Rue des Capucins"},
}

var lyonBuildings = []namedPlace{
	{slug: "boulangerie-fournier", name: "Boulangerie Fournier"},
	{slug: "cafe-des-ecoles", name: "Café des Écoles"},
	{slug: "cinema-lumiere", name: "Cinéma Lumière"},
	{slug: "hotel-des-voyageurs", name: "Hôtel des Voyageurs"},
	{slug: "immeuble-saint-paul", name: "Immeuble Saint-Paul"},
	{slug: "pharmacie-morel", name: "Pharmacie Morel"},
	{slug: "tailleur-bernard", name: "Tailleur Bernard"},
	{slug: "tabac-du-rhone", name: "Tabac du Rhône"},
}

var expandedBuildings = []namedPlace{
	{slug: "boulangerie-fournier", name: "Boulangerie Fournier"}, {slug: "cafe-des-ecoles", name: "Café des Écoles"}, {slug: "cinema-lumiere", name: "Cinéma Lumière"},
	{slug: "hotel-des-voyageurs", name: "Hôtel des Voyageurs"}, {slug: "immeuble-saint-paul", name: "Immeuble Saint-Paul"}, {slug: "pharmacie-morel", name: "Pharmacie Morel"},
	{slug: "tailleur-bernard", name: "Tailleur Bernard"}, {slug: "tabac-du-rhone", name: "Tabac du Rhône"}, {slug: "maison-des-artisans", name: "Maison des Artisans"},
	{slug: "atelier-du-progres", name: "Atelier du Progrès"}, {slug: "comptoir-central", name: "Comptoir Central"}, {slug: "imprimerie-nationale", name: "Imprimerie Nationale"},
	{slug: "grand-magasin", name: "Grand Magasin"}, {slug: "maison-de-la-presse", name: "Maison de la Presse"}, {slug: "entrepot-des-quais", name: "Entrepôt des Quais"},
	{slug: "garage-bellecour", name: "Garage Bellecour"}, {slug: "brasserie-du-marche", name: "Brasserie du Marché"}, {slug: "maison-de-couture", name: "Maison de Couture"},
	{slug: "bureau-des-transports", name: "Bureau des Transports"}, {slug: "clinique-saint-joseph", name: "Clinique Saint-Joseph"}, {slug: "librairie-du-centre", name: "Librairie du Centre"},
	{slug: "maison-des-fleurs", name: "Maison des Fleurs"}, {slug: "fonderie-rhone", name: "Fonderie Rhône"}, {slug: "entrepot-ferrand", name: "Entrepôt Ferrand"},
	{slug: "palais-du-commerce", name: "Palais du Commerce"}, {slug: "maison-montclair", name: "Maison Montclair"}, {slug: "garage-des-voutes", name: "Garage des Voûtes"},
	{slug: "hotel-de-la-gare", name: "Hôtel de la Gare"}, {slug: "cooperative-ouvriere", name: "Coopérative Ouvrière"}, {slug: "maison-des-tissus", name: "Maison des Tissus"},
	{slug: "entrepot-lafayette", name: "Entrepôt Lafayette"}, {slug: "atelier-saint-paul", name: "Atelier Saint-Paul"}, {slug: "courrier-du-soir", name: "Courrier du Soir"},
	{slug: "domaine-des-cedres", name: "Domaine des Cèdres"}, {slug: "halle-aux-grains", name: "Halle aux Grains"}, {slug: "maison-des-machines", name: "Maison des Machines"},
}

func districtBuildingSet(seed string, version int, districtSlug string) []namedPlace {
	if version < GeneratorVersion7 {
		return shuffledCopy(lyonBuildings, newStream(seed, version, "district/"+districtSlug+"/buildings"))[:2]
	}
	return shuffledCopy(expandedBuildings, newStream(seed, version, "district/"+districtSlug+"/buildings"))[:6]
}

// Generate returns a stable, in-memory city slice for a world seed and generator version.
func Generate(seed string, version int) (World, error) {
	if strings.TrimSpace(seed) == "" {
		return World{}, errors.New("world seed is required")
	}
	switch version {
	case GeneratorVersion1:
		return generateCity(seed, GeneratorVersion1), nil
	case GeneratorVersion2:
		return generateSociety(seed, GeneratorVersion2)
	case GeneratorVersion3:
		return generateSociety(seed, GeneratorVersion3)
	case GeneratorVersion4:
		return generateSociety(seed, GeneratorVersion4)
	case GeneratorVersion5:
		return generateSociety(seed, GeneratorVersion5)
	case GeneratorVersion6:
		return generateSociety(seed, GeneratorVersion6)
	case GeneratorVersion7:
		return generateExpandedSociety(seed, GeneratorVersion7)
	case GeneratorVersion8:
		return generateExpandedSociety(seed, GeneratorVersion8)
	default:
		return World{}, fmt.Errorf("%w: %d", ErrUnsupportedGeneratorVersion, version)
	}
}

// GenerateCity reconstructs one city inside a server's shared world. Lyon's
// existing identities remain untouched; additional cities receive their own
// deterministic seed scope and canonical IDs.
func GenerateCity(seed string, version int, cityID string) (World, error) {
	city, ok := CityByID(cityID)
	if !ok || strings.TrimSpace(seed) == "" {
		return World{}, ErrEntityNotFound
	}
	if cityID == "fr/lyon" {
		return Generate(seed, version)
	}
	if version < GeneratorVersion5 {
		return World{}, fmt.Errorf("city expansion requires generator version %d or later", GeneratorVersion5)
	}
	citySeed := CitySeed(seed, cityID)
	base, err := Generate(citySeed, version)
	if err != nil {
		return World{}, err
	}
	generated, err := generateCityGeography(seed, version, cityID)
	if err != nil {
		return World{}, err
	}
	oldDistricts := selectedDistricts(citySeed, version)
	newDistricts := CityDistricts(citySeed, city)
	if len(oldDistricts) != len(newDistricts) {
		return World{}, errors.New("city district catalog does not match generated world")
	}
	districtIDs := make(map[string]CityDistrict, len(oldDistricts))
	for index, oldDistrict := range oldDistricts {
		districtIDs["fr/lyon/district/"+oldDistrict.slug] = newDistricts[index]
	}

	for _, company := range base.Companies {
		targetDistrict := districtIDs[company.RegionID]
		newKey := cityGenerationKey(company.GenerationKey, city, targetDistrict)
		company, err = GenerateCompany(seed, version, newKey)
		if err != nil {
			return World{}, fmt.Errorf("rebuild company key %q for %s: %w", newKey, city.ID, err)
		}
		generated.Companies = append(generated.Companies, company)
		generated.Entities = append(generated.Entities, Entity{
			ID: company.ID, EntityType: "company", RegionID: company.RegionID,
			GenerationKey: company.GenerationKey, Name: company.Name,
		})
	}
	for _, npc := range base.NPCs {
		targetDistrict := districtIDs[npc.RegionID]
		newKey := cityGenerationKey(npc.GenerationKey, city, targetDistrict)
		npc, err = GenerateNPC(seed, version, newKey)
		if err != nil {
			return World{}, fmt.Errorf("rebuild NPC key %q for %s: %w", newKey, city.ID, err)
		}
		generated.NPCs = append(generated.NPCs, npc)
		generated.Entities = append(generated.Entities, Entity{
			ID: npc.ID, EntityType: "npc", RegionID: npc.RegionID,
			GenerationKey: newKey, Name: npc.Name,
		})
	}
	for _, relation := range base.EmploymentRelationships {
		relation.EmployeeNPCID = remapCityEntityID(relation.EmployeeNPCID, "npc:fr/lyon/", "npc:"+city.ID+"/", city, districtIDs)
		relation.EmployerCompanyID = remapCityEntityID(relation.EmployerCompanyID, "company:fr/lyon/", "company:"+city.ID+"/", city, districtIDs)
		relation.ID = "employment:" + relation.EmployeeNPCID + ":" + relation.EmployerCompanyID
		generated.EmploymentRelationships = append(generated.EmploymentRelationships, relation)
	}
	for _, relation := range base.PersonalRelationships {
		relation.PersonANPCID = remapCityEntityID(relation.PersonANPCID, "npc:fr/lyon/", "npc:"+city.ID+"/", city, districtIDs)
		relation.PersonBNPCID = remapCityEntityID(relation.PersonBNPCID, "npc:fr/lyon/", "npc:"+city.ID+"/", city, districtIDs)
		if relation.PersonBNPCID < relation.PersonANPCID {
			relation.PersonANPCID, relation.PersonBNPCID = relation.PersonBNPCID, relation.PersonANPCID
		}
		relation.ID = "personal:" + relation.RelationshipType + ":" + relation.PersonANPCID + ":" + relation.PersonBNPCID
		generated.PersonalRelationships = append(generated.PersonalRelationships, relation)
	}
	for _, company := range generated.Companies {
		buildingIndex := 1
		parts := strings.Split(company.GenerationKey, "/")
		if version >= GeneratorVersion7 && len(parts) == 10 {
			buildingIndex, _ = strconv.Atoi(parts[7])
		}
		districtSlug := strings.TrimPrefix(company.RegionID, city.ID+"/district/")
		building := districtBuildingSet(CitySeed(seed, city.ID), version, districtSlug)[(buildingIndex-1)%6]
		locationID := "location:" + city.ID + "/district/" + districtSlug + "/building/" + building.slug
		for _, entity := range generated.Entities {
			if entity.EntityType == "location" && entity.ID == locationID && entity.RegionID == company.RegionID && entity.LocationKind == "building" {
				generated.CompanyLocations = append(generated.CompanyLocations, CompanyLocation{
					ID:        "company-location:" + company.ID + ":" + entity.ID,
					CompanyID: company.ID, LocationID: entity.ID, LocationRole: "headquarters",
				})
				break
			}
		}
	}
	sortWorld(&generated)
	return generated, nil
}

// generateCityGeography builds only regions and addresses. Navigation endpoints
// use it without paying the cost of generating every NPC and relationship.
func generateCityGeography(seed string, version int, cityID string) (World, error) {
	city, ok := CityByID(cityID)
	if !ok || strings.TrimSpace(seed) == "" {
		return World{}, ErrEntityNotFound
	}
	if cityID == "fr/lyon" {
		generated := generateCity(seed, version)
		generated.GeneratorVersion = version
		return generated, nil
	}
	if version < GeneratorVersion5 {
		return World{}, fmt.Errorf("city expansion requires generator version %d or later", GeneratorVersion5)
	}
	citySeed := CitySeed(seed, cityID)
	generated := World{GeneratorVersion: version, Regions: []Region{{ID: city.ID, Kind: "city", Name: city.Name}}}
	for _, district := range CityDistricts(citySeed, city) {
		generated.Regions = append(generated.Regions, Region{ID: district.ID, ParentID: city.ID, Kind: "district", Name: district.Name})
		for _, location := range generatedCityLocations(citySeed, city, district.ID, district.Name, version) {
			address := location.Address
			key := strings.TrimPrefix(address.LocationID, "location:")
			generated.Entities = append(generated.Entities, Entity{
				ID: address.LocationID, EntityType: "location", RegionID: district.ID,
				GenerationKey: key, LocationKind: address.Kind, Name: address.Name,
			})
		}
	}
	sortWorld(&generated)
	return generated, nil
}

func remapCityEntityID(entityID, oldPrefix, newPrefix string, city WorldCity, districts map[string]CityDistrict) string {
	if !strings.HasPrefix(entityID, oldPrefix) {
		return entityID
	}
	remapped := strings.Replace(entityID, oldPrefix, newPrefix, 1)
	marker := city.ID + "/district/"
	cityRelative := strings.TrimPrefix(remapped, strings.TrimSuffix(newPrefix, "/")+"/")
	if !strings.HasPrefix(cityRelative, "people/") && !strings.HasPrefix(cityRelative, "companies/") {
		return remapped
	}
	parts := strings.Split(cityRelative, "/")
	for index, part := range parts {
		if part == "district" && index+1 < len(parts) {
			oldDistrictID := "fr/lyon/district/" + parts[index+1]
			if district, ok := districts[oldDistrictID]; ok {
				parts[index+1] = strings.TrimPrefix(district.ID, marker)
			}
			break
		}
	}
	return strings.TrimSuffix(newPrefix, "/") + "/" + strings.Join(parts, "/")
}

func sortWorld(generated *World) {
	sort.Slice(generated.Regions, func(i, j int) bool { return generated.Regions[i].ID < generated.Regions[j].ID })
	sort.Slice(generated.Entities, func(i, j int) bool { return generated.Entities[i].ID < generated.Entities[j].ID })
	sort.Slice(generated.NPCs, func(i, j int) bool { return generated.NPCs[i].ID < generated.NPCs[j].ID })
	sort.Slice(generated.Companies, func(i, j int) bool { return generated.Companies[i].ID < generated.Companies[j].ID })
	sort.Slice(generated.EmploymentRelationships, func(i, j int) bool {
		return generated.EmploymentRelationships[i].ID < generated.EmploymentRelationships[j].ID
	})
	sort.Slice(generated.PersonalRelationships, func(i, j int) bool {
		return generated.PersonalRelationships[i].ID < generated.PersonalRelationships[j].ID
	})
	sort.Slice(generated.CompanyLocations, func(i, j int) bool {
		return generated.CompanyLocations[i].ID < generated.CompanyLocations[j].ID
	})
}

func cityGenerationKey(legacyKey string, city WorldCity, district CityDistrict) string {
	parts := strings.Split(legacyKey, "/")
	for index := range parts {
		if parts[index] == "lyon" && index == 1 {
			parts[0], parts[1] = strings.SplitN(city.ID, "/", 2)[0], strings.SplitN(city.ID, "/", 2)[1]
			break
		}
	}
	for index := range parts {
		if parts[index] == "district" && index+1 < len(parts) {
			parts[index+1] = strings.TrimPrefix(district.ID, city.ID+"/district/")
			break
		}
	}
	return strings.Join(parts, "/")
}

// SupportsSocietyVersion reports whether a world version includes generated NPCs and relationships.
func SupportsSocietyVersion(version int) bool {
	return version >= GeneratorVersion2 && version <= LatestGeneratorVersion
}

func generateCity(seed string, version int) World {
	const cityRegionID = "fr/lyon"
	world := World{
		GeneratorVersion: version,
		Regions: []Region{{
			ID:   cityRegionID,
			Kind: "city",
			Name: "Lyon",
		}},
		Entities: make([]Entity, 0, 12),
	}

	districts := selectedDistricts(seed, version)

	for _, district := range districts {
		regionID := cityRegionID + "/district/" + district.slug
		world.Regions = append(world.Regions, Region{
			ID:       regionID,
			ParentID: cityRegionID,
			Kind:     "district",
			Name:     district.name,
		})

		streets := shuffledCopy(lyonStreets, newStream(seed, version, "district/"+district.slug+"/streets"))[:2]
		for _, street := range streets {
			key := "lyon/district/" + district.slug + "/street/" + street.slug
			world.Entities = append(world.Entities, Entity{
				ID:            "location:" + key,
				EntityType:    "location",
				RegionID:      regionID,
				GenerationKey: key,
				LocationKind:  "street",
				Name:          street.name,
			})
		}

		buildings := districtBuildingSet(seed, version, district.slug)
		for _, building := range buildings {
			key := "lyon/district/" + district.slug + "/building/" + building.slug
			world.Entities = append(world.Entities, Entity{
				ID:            "location:" + key,
				EntityType:    "location",
				RegionID:      regionID,
				GenerationKey: key,
				LocationKind:  "building",
				Name:          building.name,
			})
		}
	}

	sort.Slice(world.Regions, func(i, j int) bool { return world.Regions[i].ID < world.Regions[j].ID })
	sort.Slice(world.Entities, func(i, j int) bool { return world.Entities[i].ID < world.Entities[j].ID })
	return world
}

type randomStream struct {
	key     [sha256.Size]byte
	counter uint64
}

func newStream(seed string, version int, scope string) randomStream {
	h := sha256.New()
	writeField(h, "noir-1974/world-generator")
	writeField(h, fmt.Sprint(version))
	writeField(h, seed)
	writeField(h, scope)
	var stream randomStream
	copy(stream.key[:], h.Sum(nil))
	return stream
}

func writeField(h hash.Hash, value string) {
	var length [8]byte
	binary.BigEndian.PutUint64(length[:], uint64(len(value)))
	_, _ = h.Write(length[:])
	_, _ = h.Write([]byte(value))
}

func (stream *randomStream) uint64() uint64 {
	var input [sha256.Size + 8]byte
	copy(input[:sha256.Size], stream.key[:])
	binary.BigEndian.PutUint64(input[sha256.Size:], stream.counter)
	stream.counter++
	output := sha256.Sum256(input[:])
	return binary.BigEndian.Uint64(output[:8])
}

func (stream *randomStream) intn(limit int) int {
	bound := uint64(limit)
	threshold := -bound % bound
	for {
		value := stream.uint64()
		if value >= threshold {
			return int(value % bound)
		}
	}
}

func shuffledCopy[T any](values []T, stream randomStream) []T {
	copyOfValues := append([]T(nil), values...)
	for index := len(copyOfValues) - 1; index > 0; index-- {
		other := stream.intn(index + 1)
		copyOfValues[index], copyOfValues[other] = copyOfValues[other], copyOfValues[index]
	}
	return copyOfValues
}

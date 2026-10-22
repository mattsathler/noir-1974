package world

import (
	"crypto/sha256"
	"encoding/binary"
	"errors"
	"fmt"
	"hash"
	"sort"
	"strings"
)

const (
	GeneratorVersion1      = 1
	GeneratorVersion2      = 2
	GeneratorVersion3      = 3
	GeneratorVersion4      = 4
	LatestGeneratorVersion = GeneratorVersion4
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
	default:
		return World{}, fmt.Errorf("%w: %d", ErrUnsupportedGeneratorVersion, version)
	}
}

// SupportsSocietyVersion reports whether a world version includes generated NPCs and relationships.
func SupportsSocietyVersion(version int) bool {
	return version == GeneratorVersion2 || version == GeneratorVersion3 || version == GeneratorVersion4
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

		buildings := shuffledCopy(lyonBuildings, newStream(seed, version, "district/"+district.slug+"/buildings"))[:2]
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

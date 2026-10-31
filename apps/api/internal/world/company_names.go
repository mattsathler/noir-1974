package world

import (
	_ "embed"
	"encoding/json"
	"fmt"
	"strings"
)

const CompanyNameCatalogVersion = 1

//go:embed data/company_names.json
var companyNameCatalogData []byte

type companyNameCatalog struct {
	Version int      `json:"version"`
	Names   []string `json:"names"`
}

var companyNameCatalogNames = loadCompanyNameCatalog()

func loadCompanyNameCatalog() []string {
	var catalog companyNameCatalog
	if err := json.Unmarshal(companyNameCatalogData, &catalog); err != nil {
		panic(fmt.Errorf("decode company name catalog: %w", err))
	}
	if catalog.Version != CompanyNameCatalogVersion || len(catalog.Names) < 2880 {
		panic("company name catalog version or capacity is invalid")
	}
	seen := make(map[string]struct{}, len(catalog.Names))
	for _, name := range catalog.Names {
		name = strings.TrimSpace(name)
		if name == "" {
			panic("company name catalog contains an empty name")
		}
		key := strings.ToLower(name)
		if _, exists := seen[key]; exists {
			panic("company name catalog contains duplicate names")
		}
		seen[key] = struct{}{}
	}
	return catalog.Names
}

// companyNameForNationalIndex maps each national company slot to a unique
// catalog entry. The seed selects an affine permutation of the catalog, so
// every country receives stable but seed-specific names without collisions.
func companyNameForNationalIndex(seed, countryID string, index int) (string, bool) {
	count := len(companyNameCatalogNames)
	if count == 0 || index < 0 || index >= count {
		return "", false
	}
	stream := newStream(seed, CompanyNameCatalogVersion, "company/names/"+countryID)
	offset := stream.intn(count)
	step := stream.intn(count-1) + 1
	for greatestCommonDivisor(step, count) != 1 {
		step++
		if step >= count {
			step = 1
		}
	}
	return companyNameCatalogNames[(offset+index*step)%count], true
}

func companyNationalIndex(city WorldCity, companyOrdinal int) (int, bool) {
	if companyOrdinal < 1 || companyOrdinal > 720 {
		return 0, false
	}
	cityIndex := 0
	for _, candidate := range Cities() {
		if candidate.CountryID != city.CountryID {
			continue
		}
		if candidate.ID == city.ID {
			return cityIndex*720 + companyOrdinal - 1, true
		}
		cityIndex++
	}
	return 0, false
}

func greatestCommonDivisor(left, right int) int {
	for right != 0 {
		left, right = right, left%right
	}
	return left
}

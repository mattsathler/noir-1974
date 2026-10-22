// Package names provides deterministic, city-scoped name catalogs.
// Add one country file per city and register its canonical country/city slugs here.
package names

import france "noir-1974/apps/api/internal/world/names/france"

type Catalog struct {
	GivenNames       []string
	FamilyNames      []string
	DistrictPrefixes []string
	DistrictSuffixes []string
}

func ForCity(countrySlug, citySlug string) (Catalog, bool) {
	if countrySlug == "fr" && citySlug == "lyon" {
		return Catalog{
			GivenNames:       france.LyonGivenNames,
			FamilyNames:      france.LyonFamilyNames,
			DistrictPrefixes: france.LyonDistrictPrefixes,
			DistrictSuffixes: france.LyonDistrictSuffixes,
		}, true
	}
	catalog, ok := countryCatalogs[countrySlug]
	if banks, exists := countryDistrictBanks[countrySlug]; exists {
		catalog.DistrictPrefixes = banks[0]
		catalog.DistrictSuffixes = banks[1]
	}
	return catalog, ok
}

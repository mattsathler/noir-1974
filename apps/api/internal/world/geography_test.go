package world

import (
	"strings"
	"testing"

	"noir-1974/apps/api/internal/world/names"
)

func TestCityCatalogContainsOneHundredStableDestinationsAcrossTwentyFiveCountries(t *testing.T) {
	cities := Cities()
	if len(cities) != 100 {
		t.Fatalf("catalog contains %d cities, want 100", len(cities))
	}
	cityIDs := make(map[string]struct{}, len(cities))
	countryCities := make(map[string]int)
	for _, city := range cities {
		if city.ID == "" || city.CountryID == "" || city.Country == "" || city.Name == "" || city.TimeZone == "" {
			t.Fatalf("incomplete city catalog entry: %+v", city)
		}
		if _, exists := cityIDs[city.ID]; exists {
			t.Fatalf("duplicate city ID %q", city.ID)
		}
		cityIDs[city.ID] = struct{}{}
		countryCities[city.CountryID]++
		if got, ok := CityByID(city.ID); !ok || got != city {
			t.Fatalf("CityByID(%q) = %+v, %t; want %+v, true", city.ID, got, ok, city)
		}
	}
	if len(countryCities) != 25 {
		t.Fatalf("catalog contains %d countries, want 25", len(countryCities))
	}
	for countryID, count := range countryCities {
		if count != 4 {
			t.Errorf("country %q has %d cities, want 4", countryID, count)
		}
	}
	for _, city := range cities {
		parts := strings.Split(city.ID, "/")
		catalog, ok := names.ForCity(parts[0], parts[1])
		if !ok || len(catalog.GivenNames) < 12 || len(catalog.FamilyNames) < 12 || len(catalog.DistrictPrefixes) < 4 || len(catalog.DistrictSuffixes) < 5 {
			t.Errorf("city %q has no usable local country name catalog", city.ID)
		}
	}
}

func TestCityCatalogUsesEnglishDisplayNames(t *testing.T) {
	countries := map[string]string{
		"us": "United States", "cn": "China", "ru": "Russia", "gb": "United Kingdom", "de": "Germany",
		"jp": "Japan", "fr": "France", "in": "India", "kr": "South Korea", "sa": "Saudi Arabia",
		"ca": "Canada", "br": "Brazil", "it": "Italy", "au": "Australia", "es": "Spain",
		"tr": "Turkey", "mx": "Mexico", "id": "Indonesia", "ae": "United Arab Emirates", "ch": "Switzerland",
		"nl": "Netherlands", "sg": "Singapore", "il": "Israel", "za": "South Africa", "ar": "Argentina",
	}
	for _, city := range Cities() {
		if city.Country != countries[city.CountryID] {
			t.Errorf("country for %s is %q, want English catalog name %q", city.ID, city.Country, countries[city.CountryID])
		}
	}
	checks := map[string]WorldCity{
		"us/new-york":       {ID: "us/new-york", Country: "United States", Name: "New York"},
		"cn/beijing":        {ID: "cn/beijing", Country: "China", Name: "Beijing"},
		"ru/moscow":         {ID: "ru/moscow", Country: "Russia", Name: "Moscow"},
		"gb/london":         {ID: "gb/london", Country: "United Kingdom", Name: "London"},
		"br/rio-de-janeiro": {ID: "br/rio-de-janeiro", Country: "Brazil", Name: "Rio de Janeiro"},
		"za/cape-town":      {ID: "za/cape-town", Country: "South Africa", Name: "Cape Town"},
	}
	for id, expected := range checks {
		got, ok := CityByID(id)
		if !ok || got.Country != expected.Country || got.Name != expected.Name {
			t.Errorf("CityByID(%q) = %+v, %t; want country %q, name %q", id, got, ok, expected.Country, expected.Name)
		}
	}
}

func TestCityForLocationKeepsExistingLyonLocationIDs(t *testing.T) {
	for _, locationID := range []string{
		"fr/lyon/district/central",
		"service:lyon/meridian-bank",
		"location:jp/tokyo/district/old-quarter/building/the-lantern-cafe",
	} {
		city, ok := CityForLocation(locationID)
		want := "fr/lyon"
		if strings.Contains(locationID, "jp/tokyo") {
			want = "jp/tokyo"
		}
		if !ok || city.ID != want {
			t.Errorf("CityForLocation(%q) = %+v, %t; want %s", locationID, city, ok, want)
		}
	}
}

func TestCityDistrictsAndLocationsAreDeterministicAndCityScoped(t *testing.T) {
	city, ok := CityByID("jp/tokyo")
	if !ok {
		t.Fatal("Tokyo is missing from the city catalog")
	}
	first := CityDistricts("shared-world-seed", city)
	second := CityDistricts("shared-world-seed", city)
	if len(first) != 20 {
		t.Fatalf("district count = %d, want 20", len(first))
	}
	seen := make(map[string]struct{}, len(first))
	for i, district := range first {
		if district != second[i] {
			t.Fatalf("district %d changed between generation calls", i)
		}
		if _, exists := seen[district.ID]; exists {
			t.Fatalf("duplicate district ID %q", district.ID)
		}
		seen[district.ID] = struct{}{}
		if !strings.HasPrefix(district.ID, city.ID+"/district/") {
			t.Fatalf("district ID %q is not scoped to city %q", district.ID, city.ID)
		}
		locations := generatedCityLocations("shared-world-seed", city, district.ID, district.Name)
		if len(locations) != 4 {
			t.Errorf("district %q has %d locations, want four", district.ID, len(locations))
		}
	}
}

func TestEveryCityHasTheSameFourCentralBuildingCategories(t *testing.T) {
	wantCategories := []string{"police-station", "meridian-bank", "gate-station", "hospital"}
	for _, city := range Cities() {
		region := Region{ID: city.ID, Name: city.Name, Kind: "city"}
		first := centralServiceLocations(region)
		second := centralServiceLocations(region)
		if len(first) != len(wantCategories) || len(second) != len(wantCategories) {
			t.Fatalf("city %s has %d central buildings, want %d", city.ID, len(first), len(wantCategories))
		}
		for index, category := range wantCategories {
			firstAddress, secondAddress := first[index].Address, second[index].Address
			if !strings.HasSuffix(firstAddress.LocationID, "/"+category) || firstAddress.LocationID != secondAddress.LocationID {
				t.Errorf("city %s central category %s has unstable ID: %q / %q", city.ID, category, firstAddress.LocationID, secondAddress.LocationID)
			}
			if firstAddress.Kind != "building" || firstAddress.District != "Central" || firstAddress.City != city.Name || firstAddress.Country != city.Country {
				t.Errorf("city %s has inconsistent central building metadata: %+v", city.ID, firstAddress)
			}
			wantStatus := ""
			if category == "hospital" {
				wantStatus = "open"
			}
			if firstAddress.Status != wantStatus {
				t.Errorf("city %s %s status = %q, want %q", city.ID, category, firstAddress.Status, wantStatus)
			}
		}
	}
}

func TestGeneratedCityRelationshipsReferenceMaterializedEntities(t *testing.T) {
	for _, cityID := range []string{"us/los-angeles", "br/rio-de-janeiro"} {
		generated, err := GenerateCity("shared-world-seed", GeneratorVersion5, cityID)
		if err != nil {
			t.Fatalf("GenerateCity(%q): %v", cityID, err)
		}
		entities := make(map[string]Entity, len(generated.Entities))
		for _, entity := range generated.Entities {
			entities[entity.ID] = entity
		}
		for _, relation := range generated.EmploymentRelationships {
			if entity, ok := entities[relation.EmployeeNPCID]; !ok || entity.EntityType != "npc" {
				t.Errorf("city %s employment %s references missing employee %q", cityID, relation.ID, relation.EmployeeNPCID)
			}
			if entity, ok := entities[relation.EmployerCompanyID]; !ok || entity.EntityType != "company" {
				t.Errorf("city %s employment %s references missing employer %q", cityID, relation.ID, relation.EmployerCompanyID)
			}
		}
		for _, relation := range generated.PersonalRelationships {
			for _, npcID := range []string{relation.PersonANPCID, relation.PersonBNPCID} {
				if entity, ok := entities[npcID]; !ok || entity.EntityType != "npc" {
					t.Errorf("city %s personal relationship %s references missing NPC %q", cityID, relation.ID, npcID)
				}
			}
		}
		for _, relation := range generated.CompanyLocations {
			if entity, ok := entities[relation.CompanyID]; !ok || entity.EntityType != "company" {
				t.Errorf("city %s company location %s references missing company %q", cityID, relation.ID, relation.CompanyID)
			}
			if entity, ok := entities[relation.LocationID]; !ok || entity.EntityType != "location" {
				t.Errorf("city %s company location %s references missing location %q", cityID, relation.ID, relation.LocationID)
			}
		}
	}
}

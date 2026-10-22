package world

import (
	"crypto/sha256"
	"encoding/hex"
	"strings"

	"noir-1974/apps/api/internal/world/names"
)

const GeographyCatalogVersion = 1

// WorldCity is a canonical destination shared by every Noir server.
type WorldCity struct {
	ID        string `json:"id"`
	CountryID string `json:"countryId"`
	Country   string `json:"country"`
	Name      string `json:"name"`
	TimeZone  string `json:"timeZone"`
}

type CityDistrict struct {
	ID   string
	Name string
}

// Cities returns the fixed, version-independent city catalog in display order.
func Cities() []WorldCity {
	return append([]WorldCity(nil), worldCities...)
}

// CityByID resolves a canonical country/city ID.
func CityByID(id string) (WorldCity, bool) {
	for _, city := range worldCities {
		if city.ID == id {
			return city, true
		}
	}
	return WorldCity{}, false
}

// CityForLocation resolves a persisted local location ID to its city.
func CityForLocation(locationID string) (WorldCity, bool) {
	if strings.HasPrefix(locationID, "service:") || strings.HasPrefix(locationID, "location:") {
		legacy := strings.TrimPrefix(strings.TrimPrefix(locationID, "service:"), "location:")
		if strings.HasPrefix(legacy, "lyon/") {
			return CityByID("fr/lyon")
		}
		for _, city := range worldCities {
			if strings.HasPrefix(legacy, city.ID+"/") || legacy == city.ID {
				return city, true
			}
		}
	}
	for _, city := range worldCities {
		if locationID == city.ID || strings.HasPrefix(locationID, city.ID+"/") {
			return city, true
		}
	}
	return WorldCity{}, false
}

// CitySeed derives the deterministic content seed for one city in a shared world.
func CitySeed(seed, cityID string) string {
	if cityID == "fr/lyon" {
		return seed
	}
	hash := sha256.Sum256([]byte("noir-1974/city-world/v1\x00" + seed + "\x00" + cityID))
	return "city-v1:" + hex.EncodeToString(hash[:])
}

// CityDistricts creates twenty stable, city-scoped logical neighborhoods.
func CityDistricts(seed string, city WorldCity) []CityDistrict {
	catalog, hasCatalog := names.ForCity(city.CountryID, slug(city.Name))
	prefixes := []string{city.Name, "Old", "Upper", "Lower", "North", "South", "East", "West", "New", "Grand", "Little"}
	suffixes := []string{"Quarter", "Market Ward", "Riverside", "Station Ward", "Gardens", "Exchange", "Terrace", "Heights", "Foundry Ward", "Harbor Ward"}
	if hasCatalog && len(catalog.DistrictPrefixes) > 0 && len(catalog.DistrictSuffixes) > 0 {
		prefixes = append([]string{city.Name}, catalog.DistrictPrefixes...)
		suffixes = catalog.DistrictSuffixes
	}
	candidates := make([]CityDistrict, 0, len(prefixes)*len(suffixes))
	for _, prefix := range prefixes {
		for _, suffix := range suffixes {
			name := prefix + " " + suffix
			candidates = append(candidates, CityDistrict{
				ID: city.ID + "/district/" + slug(name), Name: name,
			})
		}
	}
	selected := shuffledCopy(candidates, newStream(seed, GeneratorVersion5, "city/"+city.ID+"/districts"))[:20]
	for i := 1; i < len(selected); i++ {
		for j := i; j > 0 && selected[j].ID < selected[j-1].ID; j-- {
			selected[j], selected[j-1] = selected[j-1], selected[j]
		}
	}
	return selected
}

func slug(value string) string {
	value = strings.ToLower(value)
	var builder strings.Builder
	previousDash := false
	for _, char := range value {
		if char >= 'a' && char <= 'z' || char >= '0' && char <= '9' {
			builder.WriteRune(char)
			previousDash = false
		} else if !previousDash && builder.Len() > 0 {
			builder.WriteByte('-')
			previousDash = true
		}
	}
	return strings.Trim(builder.String(), "-")
}

var worldCities = []WorldCity{
	{ID: "us/new-york", CountryID: "us", Country: "United States", Name: "New York", TimeZone: "America/New_York"},
	{ID: "us/washington-dc", CountryID: "us", Country: "United States", Name: "Washington, D.C.", TimeZone: "America/New_York"},
	{ID: "us/los-angeles", CountryID: "us", Country: "United States", Name: "Los Angeles", TimeZone: "America/Los_Angeles"},
	{ID: "us/san-francisco", CountryID: "us", Country: "United States", Name: "San Francisco", TimeZone: "America/Los_Angeles"},
	{ID: "cn/beijing", CountryID: "cn", Country: "China", Name: "Beijing", TimeZone: "Asia/Shanghai"},
	{ID: "cn/shanghai", CountryID: "cn", Country: "China", Name: "Shanghai", TimeZone: "Asia/Shanghai"},
	{ID: "cn/shenzhen", CountryID: "cn", Country: "China", Name: "Shenzhen", TimeZone: "Asia/Shanghai"},
	{ID: "cn/guangzhou", CountryID: "cn", Country: "China", Name: "Guangzhou", TimeZone: "Asia/Shanghai"},
	{ID: "ru/moscow", CountryID: "ru", Country: "Russia", Name: "Moscow", TimeZone: "Europe/Moscow"},
	{ID: "ru/saint-petersburg", CountryID: "ru", Country: "Russia", Name: "Saint Petersburg", TimeZone: "Europe/Moscow"},
	{ID: "ru/novosibirsk", CountryID: "ru", Country: "Russia", Name: "Novosibirsk", TimeZone: "Asia/Novosibirsk"},
	{ID: "ru/vladivostok", CountryID: "ru", Country: "Russia", Name: "Vladivostok", TimeZone: "Asia/Vladivostok"},
	{ID: "gb/london", CountryID: "gb", Country: "United Kingdom", Name: "London", TimeZone: "Europe/London"},
	{ID: "gb/manchester", CountryID: "gb", Country: "United Kingdom", Name: "Manchester", TimeZone: "Europe/London"},
	{ID: "gb/birmingham", CountryID: "gb", Country: "United Kingdom", Name: "Birmingham", TimeZone: "Europe/London"},
	{ID: "gb/edinburgh", CountryID: "gb", Country: "United Kingdom", Name: "Edinburgh", TimeZone: "Europe/London"},
	{ID: "de/berlin", CountryID: "de", Country: "Germany", Name: "Berlin", TimeZone: "Europe/Berlin"},
	{ID: "de/frankfurt", CountryID: "de", Country: "Germany", Name: "Frankfurt", TimeZone: "Europe/Berlin"},
	{ID: "de/munich", CountryID: "de", Country: "Germany", Name: "Munich", TimeZone: "Europe/Berlin"},
	{ID: "de/hamburg", CountryID: "de", Country: "Germany", Name: "Hamburg", TimeZone: "Europe/Berlin"},
	{ID: "jp/tokyo", CountryID: "jp", Country: "Japan", Name: "Tokyo", TimeZone: "Asia/Tokyo"},
	{ID: "jp/osaka", CountryID: "jp", Country: "Japan", Name: "Osaka", TimeZone: "Asia/Tokyo"},
	{ID: "jp/yokohama", CountryID: "jp", Country: "Japan", Name: "Yokohama", TimeZone: "Asia/Tokyo"},
	{ID: "jp/kyoto", CountryID: "jp", Country: "Japan", Name: "Kyoto", TimeZone: "Asia/Tokyo"},
	{ID: "fr/paris", CountryID: "fr", Country: "France", Name: "Paris", TimeZone: "Europe/Paris"},
	{ID: "fr/lyon", CountryID: "fr", Country: "France", Name: "Lyon", TimeZone: "Europe/Paris"},
	{ID: "fr/marseille", CountryID: "fr", Country: "France", Name: "Marseille", TimeZone: "Europe/Paris"},
	{ID: "fr/toulouse", CountryID: "fr", Country: "France", Name: "Toulouse", TimeZone: "Europe/Paris"},
	{ID: "in/new-delhi", CountryID: "in", Country: "India", Name: "New Delhi", TimeZone: "Asia/Kolkata"},
	{ID: "in/mumbai", CountryID: "in", Country: "India", Name: "Mumbai", TimeZone: "Asia/Kolkata"},
	{ID: "in/bengaluru", CountryID: "in", Country: "India", Name: "Bengaluru", TimeZone: "Asia/Kolkata"},
	{ID: "in/hyderabad", CountryID: "in", Country: "India", Name: "Hyderabad", TimeZone: "Asia/Kolkata"},
	{ID: "kr/seoul", CountryID: "kr", Country: "South Korea", Name: "Seoul", TimeZone: "Asia/Seoul"},
	{ID: "kr/busan", CountryID: "kr", Country: "South Korea", Name: "Busan", TimeZone: "Asia/Seoul"},
	{ID: "kr/incheon", CountryID: "kr", Country: "South Korea", Name: "Incheon", TimeZone: "Asia/Seoul"},
	{ID: "kr/daegu", CountryID: "kr", Country: "South Korea", Name: "Daegu", TimeZone: "Asia/Seoul"},
	{ID: "sa/riyadh", CountryID: "sa", Country: "Saudi Arabia", Name: "Riyadh", TimeZone: "Asia/Riyadh"},
	{ID: "sa/jeddah", CountryID: "sa", Country: "Saudi Arabia", Name: "Jeddah", TimeZone: "Asia/Riyadh"},
	{ID: "sa/mecca", CountryID: "sa", Country: "Saudi Arabia", Name: "Mecca", TimeZone: "Asia/Riyadh"},
	{ID: "sa/medina", CountryID: "sa", Country: "Saudi Arabia", Name: "Medina", TimeZone: "Asia/Riyadh"},
	{ID: "ca/toronto", CountryID: "ca", Country: "Canada", Name: "Toronto", TimeZone: "America/Toronto"},
	{ID: "ca/vancouver", CountryID: "ca", Country: "Canada", Name: "Vancouver", TimeZone: "America/Vancouver"},
	{ID: "ca/montreal", CountryID: "ca", Country: "Canada", Name: "Montreal", TimeZone: "America/Toronto"},
	{ID: "ca/ottawa", CountryID: "ca", Country: "Canada", Name: "Ottawa", TimeZone: "America/Toronto"},
	{ID: "br/sao-paulo", CountryID: "br", Country: "Brazil", Name: "São Paulo", TimeZone: "America/Sao_Paulo"},
	{ID: "br/rio-de-janeiro", CountryID: "br", Country: "Brazil", Name: "Rio de Janeiro", TimeZone: "America/Sao_Paulo"},
	{ID: "br/brasilia", CountryID: "br", Country: "Brazil", Name: "Brasília", TimeZone: "America/Sao_Paulo"},
	{ID: "br/belo-horizonte", CountryID: "br", Country: "Brazil", Name: "Belo Horizonte", TimeZone: "America/Sao_Paulo"},
	{ID: "it/rome", CountryID: "it", Country: "Italy", Name: "Rome", TimeZone: "Europe/Rome"},
	{ID: "it/milan", CountryID: "it", Country: "Italy", Name: "Milan", TimeZone: "Europe/Rome"},
	{ID: "it/venice", CountryID: "it", Country: "Italy", Name: "Venice", TimeZone: "Europe/Rome"},
	{ID: "it/turin", CountryID: "it", Country: "Italy", Name: "Turin", TimeZone: "Europe/Rome"},
	{ID: "au/sydney", CountryID: "au", Country: "Australia", Name: "Sydney", TimeZone: "Australia/Sydney"},
	{ID: "au/melbourne", CountryID: "au", Country: "Australia", Name: "Melbourne", TimeZone: "Australia/Melbourne"},
	{ID: "au/brisbane", CountryID: "au", Country: "Australia", Name: "Brisbane", TimeZone: "Australia/Brisbane"},
	{ID: "au/perth", CountryID: "au", Country: "Australia", Name: "Perth", TimeZone: "Australia/Perth"},
	{ID: "es/madrid", CountryID: "es", Country: "Spain", Name: "Madrid", TimeZone: "Europe/Madrid"},
	{ID: "es/barcelona", CountryID: "es", Country: "Spain", Name: "Barcelona", TimeZone: "Europe/Madrid"},
	{ID: "es/valencia", CountryID: "es", Country: "Spain", Name: "Valencia", TimeZone: "Europe/Madrid"},
	{ID: "es/seville", CountryID: "es", Country: "Spain", Name: "Seville", TimeZone: "Europe/Madrid"},
	{ID: "tr/istanbul", CountryID: "tr", Country: "Turkey", Name: "Istanbul", TimeZone: "Europe/Istanbul"},
	{ID: "tr/ankara", CountryID: "tr", Country: "Turkey", Name: "Ankara", TimeZone: "Europe/Istanbul"},
	{ID: "tr/izmir", CountryID: "tr", Country: "Turkey", Name: "Izmir", TimeZone: "Europe/Istanbul"},
	{ID: "tr/antalya", CountryID: "tr", Country: "Turkey", Name: "Antalya", TimeZone: "Europe/Istanbul"},
	{ID: "mx/mexico-city", CountryID: "mx", Country: "Mexico", Name: "Mexico City", TimeZone: "America/Mexico_City"},
	{ID: "mx/guadalajara", CountryID: "mx", Country: "Mexico", Name: "Guadalajara", TimeZone: "America/Mexico_City"},
	{ID: "mx/monterrey", CountryID: "mx", Country: "Mexico", Name: "Monterrey", TimeZone: "America/Monterrey"},
	{ID: "mx/cancun", CountryID: "mx", Country: "Mexico", Name: "Cancun", TimeZone: "America/Cancun"},
	{ID: "id/jakarta", CountryID: "id", Country: "Indonesia", Name: "Jakarta", TimeZone: "Asia/Jakarta"},
	{ID: "id/surabaya", CountryID: "id", Country: "Indonesia", Name: "Surabaya", TimeZone: "Asia/Jakarta"},
	{ID: "id/bandung", CountryID: "id", Country: "Indonesia", Name: "Bandung", TimeZone: "Asia/Jakarta"},
	{ID: "id/denpasar", CountryID: "id", Country: "Indonesia", Name: "Denpasar", TimeZone: "Asia/Makassar"},
	{ID: "ae/dubai", CountryID: "ae", Country: "United Arab Emirates", Name: "Dubai", TimeZone: "Asia/Dubai"},
	{ID: "ae/abu-dhabi", CountryID: "ae", Country: "United Arab Emirates", Name: "Abu Dhabi", TimeZone: "Asia/Dubai"},
	{ID: "ae/sharjah", CountryID: "ae", Country: "United Arab Emirates", Name: "Sharjah", TimeZone: "Asia/Dubai"},
	{ID: "ae/al-ain", CountryID: "ae", Country: "United Arab Emirates", Name: "Al Ain", TimeZone: "Asia/Dubai"},
	{ID: "ch/zurich", CountryID: "ch", Country: "Switzerland", Name: "Zurich", TimeZone: "Europe/Zurich"},
	{ID: "ch/geneva", CountryID: "ch", Country: "Switzerland", Name: "Geneva", TimeZone: "Europe/Zurich"},
	{ID: "ch/basel", CountryID: "ch", Country: "Switzerland", Name: "Basel", TimeZone: "Europe/Zurich"},
	{ID: "ch/bern", CountryID: "ch", Country: "Switzerland", Name: "Bern", TimeZone: "Europe/Zurich"},
	{ID: "nl/amsterdam", CountryID: "nl", Country: "Netherlands", Name: "Amsterdam", TimeZone: "Europe/Amsterdam"},
	{ID: "nl/rotterdam", CountryID: "nl", Country: "Netherlands", Name: "Rotterdam", TimeZone: "Europe/Amsterdam"},
	{ID: "nl/the-hague", CountryID: "nl", Country: "Netherlands", Name: "The Hague", TimeZone: "Europe/Amsterdam"},
	{ID: "nl/eindhoven", CountryID: "nl", Country: "Netherlands", Name: "Eindhoven", TimeZone: "Europe/Amsterdam"},
	{ID: "sg/singapore", CountryID: "sg", Country: "Singapore", Name: "Singapore", TimeZone: "Asia/Singapore"},
	{ID: "sg/jurong-east", CountryID: "sg", Country: "Singapore", Name: "Jurong East", TimeZone: "Asia/Singapore"},
	{ID: "sg/woodlands", CountryID: "sg", Country: "Singapore", Name: "Woodlands", TimeZone: "Asia/Singapore"},
	{ID: "sg/tampines", CountryID: "sg", Country: "Singapore", Name: "Tampines", TimeZone: "Asia/Singapore"},
	{ID: "il/tel-aviv", CountryID: "il", Country: "Israel", Name: "Tel Aviv", TimeZone: "Asia/Jerusalem"},
	{ID: "il/jerusalem", CountryID: "il", Country: "Israel", Name: "Jerusalem", TimeZone: "Asia/Jerusalem"},
	{ID: "il/haifa", CountryID: "il", Country: "Israel", Name: "Haifa", TimeZone: "Asia/Jerusalem"},
	{ID: "il/beersheba", CountryID: "il", Country: "Israel", Name: "Beersheba", TimeZone: "Asia/Jerusalem"},
	{ID: "za/johannesburg", CountryID: "za", Country: "South Africa", Name: "Johannesburg", TimeZone: "Africa/Johannesburg"},
	{ID: "za/cape-town", CountryID: "za", Country: "South Africa", Name: "Cape Town", TimeZone: "Africa/Johannesburg"},
	{ID: "za/durban", CountryID: "za", Country: "South Africa", Name: "Durban", TimeZone: "Africa/Johannesburg"},
	{ID: "za/pretoria", CountryID: "za", Country: "South Africa", Name: "Pretoria", TimeZone: "Africa/Johannesburg"},
	{ID: "ar/buenos-aires", CountryID: "ar", Country: "Argentina", Name: "Buenos Aires", TimeZone: "America/Argentina/Buenos_Aires"},
	{ID: "ar/cordoba", CountryID: "ar", Country: "Argentina", Name: "Cordoba", TimeZone: "America/Argentina/Cordoba"},
	{ID: "ar/rosario", CountryID: "ar", Country: "Argentina", Name: "Rosario", TimeZone: "America/Argentina/Cordoba"},
	{ID: "ar/mendoza", CountryID: "ar", Country: "Argentina", Name: "Mendoza", TimeZone: "America/Argentina/Mendoza"},
}

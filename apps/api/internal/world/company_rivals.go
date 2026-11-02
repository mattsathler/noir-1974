package world

import (
	"crypto/sha256"
	"encoding/hex"
	"sort"
	"strconv"
	"strings"
	"sync"
)

type CompanyRivalGroup struct {
	ID         string
	Industry   string
	CompanyIDs []string
}

var companyRivalGroupCache sync.Map

// CompanyRivalGroups builds stable, seed-specific rival groups from the
// generated company identities and business types. Company IDs include their
// generation fingerprint, so identical inputs always produce identical groups.
func CompanyRivalGroups(countryID string, companies []Company) []CompanyRivalGroup {
	cacheParts := make([]string, 0, len(companies))
	for _, company := range companies {
		if company.ID != "" && company.BusinessType.ID != "" {
			cacheParts = append(cacheParts, company.BusinessType.ID+"\x00"+company.ID)
		}
	}
	sort.Strings(cacheParts)
	cacheDigest := sha256.Sum256([]byte(countryID + "\x00" + strings.Join(cacheParts, "\x00")))
	cacheKey := "index:" + hex.EncodeToString(cacheDigest[:])
	if cached, ok := companyRivalGroupCache.Load(cacheKey); ok {
		return cloneCompanyRivalGroups(cached.([]CompanyRivalGroup))
	}
	byIndustry := make(map[string][]string)
	for _, company := range companies {
		industry := strings.TrimSpace(company.BusinessType.ID)
		if industry == "" || company.ID == "" {
			continue
		}
		byIndustry[industry] = append(byIndustry[industry], company.ID)
	}

	industries := make([]string, 0, len(byIndustry))
	for industry := range byIndustry {
		industries = append(industries, industry)
	}
	sort.Strings(industries)
	groups := make([]CompanyRivalGroup, 0)
	for _, industry := range industries {
		ids := byIndustry[industry]
		sort.Slice(ids, func(i, j int) bool {
			left := sha256.Sum256([]byte("noir-company-rival-order-v1\x00" + countryID + "\x00" + industry + "\x00" + ids[i]))
			right := sha256.Sum256([]byte("noir-company-rival-order-v1\x00" + countryID + "\x00" + industry + "\x00" + ids[j]))
			return hex.EncodeToString(left[:]) < hex.EncodeToString(right[:])
		})
		for len(ids) >= 2 {
			digest := sha256.Sum256([]byte("noir-company-rival-size-v1\x00" + countryID + "\x00" + industry + "\x00" + strings.Join(ids, "\x00")))
			size := min(4, 2+int(digest[0]%3))
			if len(ids)-size == 1 {
				if size == 2 {
					size = 3
				} else {
					size--
				}
			}
			members := append([]string(nil), ids[:size]...)
			ids = ids[size:]
			groupDigest := sha256.Sum256([]byte("noir-company-rival-group-v1\x00" + countryID + "\x00" + industry + "\x00" + strings.Join(members, "\x00")))
			groups = append(groups, CompanyRivalGroup{ID: hex.EncodeToString(groupDigest[:12]), Industry: industry, CompanyIDs: members})
		}
	}
	actual, _ := companyRivalGroupCache.LoadOrStore(cacheKey, groups)
	return cloneCompanyRivalGroups(actual.([]CompanyRivalGroup))
}

// NationalCompanyRivalGroups reconstructs the generated company catalogue for
// a country and returns the same rival groups used by market displays.
func NationalCompanyRivalGroups(seed string, version int, countryID string) ([]CompanyRivalGroup, error) {
	cacheKey := seed + "\x00" + strconv.Itoa(version) + "\x00" + countryID
	if cached, ok := companyRivalGroupCache.Load(cacheKey); ok {
		groups := cached.([]CompanyRivalGroup)
		return cloneCompanyRivalGroups(groups), nil
	}
	companies := make([]Company, 0)
	for _, city := range Cities() {
		if city.CountryID != countryID {
			continue
		}
		generated, err := GenerateCity(seed, version, city.ID)
		if err != nil {
			return nil, err
		}
		companies = append(companies, generated.Companies...)
	}
	if len(companies) == 0 {
		return nil, ErrEntityNotFound
	}
	groups := CompanyRivalGroups(countryID, companies)
	actual, _ := companyRivalGroupCache.LoadOrStore(cacheKey, groups)
	return cloneCompanyRivalGroups(actual.([]CompanyRivalGroup)), nil
}

func CompanyRivalGroupFor(seed string, version int, companyID string) (CompanyRivalGroup, error) {
	key := strings.TrimPrefix(companyID, "company:")
	separator := strings.Index(key, "/companies/")
	if separator < 0 {
		return CompanyRivalGroup{}, ErrInvalidGenerationKey
	}
	cityID := key[:separator]
	if cityID == "lyon" {
		cityID = "fr/lyon"
	}
	city, ok := CityByID(cityID)
	if !ok {
		return CompanyRivalGroup{}, ErrInvalidGenerationKey
	}
	groups, err := NationalCompanyRivalGroups(seed, version, city.CountryID)
	if err != nil {
		return CompanyRivalGroup{}, err
	}
	for _, group := range groups {
		for _, memberID := range group.CompanyIDs {
			if memberID == companyID {
				return group, nil
			}
		}
	}
	return CompanyRivalGroup{}, ErrEntityNotFound
}

func cloneCompanyRivalGroups(groups []CompanyRivalGroup) []CompanyRivalGroup {
	cloned := make([]CompanyRivalGroup, len(groups))
	for index, group := range groups {
		cloned[index] = CompanyRivalGroup{ID: group.ID, Industry: group.Industry, CompanyIDs: append([]string(nil), group.CompanyIDs...)}
	}
	return cloned
}

package world

import (
	"crypto/sha256"
	"encoding/binary"
	"fmt"
)

type NetworkProvider struct {
	ID          string              `json:"id"`
	Name        string              `json:"name"`
	Host        string              `json:"host"`
	ProviderIP  string              `json:"providerIp"`
	Directories []ProviderDirectory `json:"directories"`
}

type ProviderDirectory struct {
	ID             string         `json:"id"`
	Name           string         `json:"name"`
	IP             string         `json:"ip"`
	Address        string         `json:"address"`
	SiteAddress    string         `json:"siteAddress,omitempty"`
	SiteName       string         `json:"siteName,omitempty"`
	Company        *Company       `json:"company,omitempty"`
	CompanyAddress *Address       `json:"companyAddress,omitempty"`
	Files          []ProviderFile `json:"files"`
}

type ProviderFile struct {
	ID   string `json:"id"`
	Name string `json:"name"`
	Kind string `json:"kind"`
}

type providerIdentity struct {
	name string
	host string
}

var providerIdentities = []providerIdentity{
	{name: "Atlas Global Telecommunications", host: "atlasglobal"},
	{name: "Northstar Communications Group", host: "northstar"},
	{name: "Continental Data Systems", host: "continental"},
}

const companySiteProviderHost = "northstar"

// CompanySiteIP assigns collision-free addresses by stable position in the seeded country index.
func CompanySiteIP(seed, countryID string, ordinal int) string {
	const addressSpace = 254 * 254
	hash := sha256.Sum256([]byte(seed + "\x00" + countryID))
	start := int(binary.BigEndian.Uint16(hash[:2])) % addressSpace
	slot := (start + ordinal) % addressSpace
	return fmt.Sprintf("123.224.%d.%d", slot/254+1, slot%254+1)
}

func CompanySiteAddress(seed, countryID string, ordinal int) string {
	return fmt.Sprintf("merid//%s.%s:4000/", companySiteProviderHost, CompanySiteIP(seed, countryID, ordinal))
}

// HostingProviders returns the provider directory available in one country.
func HostingProviders(countryID string) ([]NetworkProvider, bool) {
	countryExists := false
	for _, city := range worldCities {
		if city.CountryID == countryID {
			countryExists = true
			break
		}
	}
	if !countryExists {
		return nil, false
	}

	countryHash := 0
	for _, character := range countryID {
		countryHash += int(character)
	}
	thirdOctet := 40 + countryHash%180
	providers := make([]NetworkProvider, 0, len(providerIdentities))
	for index, identity := range providerIdentities {
		providerIP := fmt.Sprintf("123.223.%d.%d", thirdOctet, index+1)
		baseIP := index*2 + 1
		directories := make([]ProviderDirectory, 0, 3)
		for folderIndex := 0; folderIndex < 3; folderIndex++ {
			ip := fmt.Sprintf("123.223.%d.%d", thirdOctet, baseIP+folderIndex)
			directory := ProviderDirectory{
				ID:      ip,
				Name:    ip,
				IP:      ip,
				Address: fmt.Sprintf("merid//%s.%s:4000/", identity.host, ip),
				Files:   []ProviderFile{},
			}
			if index == 0 && folderIndex == 0 {
				directory.SiteName = "MERIDIAN POLICE SERVICE"
				directory.Files = []ProviderFile{
					{ID: "person-search", Name: "person_search_app.exe", Kind: "person-search"},
					{ID: "warrant-request", Name: "search_warrant.exe", Kind: "warrant-request"},
					{ID: "murder-accusation", Name: "arrest_warrant_app.exe", Kind: "murder-accusation"},
				}
			} else if folderIndex == 0 && index == 2 {
				directory.SiteName = "MERIDIAN ENTERPRISES"
				directory.SiteAddress = directory.Address + "gate/access-logs"
				directory.Files = []ProviderFile{{ID: "gate-records", Name: "gate_access_log_app.exe", Kind: "gate-records"}}
			}
			directories = append(directories, directory)
		}
		providers = append(providers, NetworkProvider{
			ID:          identity.host,
			Name:        identity.name,
			Host:        identity.host,
			ProviderIP:  providerIP,
			Directories: directories,
		})
	}
	return providers, true
}

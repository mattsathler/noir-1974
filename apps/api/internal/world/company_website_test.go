package world

import (
	"encoding/json"
	"fmt"
	"reflect"
	"testing"
)

func companyWebsiteFixture(id, name string) CompanyPortalRecord {
	return CompanyPortalRecord{
		Company: Company{ID: id, Name: name},
		Address: Address{
			LocationID: "location:" + id,
			Name:       name + " Office",
			District:   "Central",
			City:       "Lyon",
			Country:    "France",
			Full:       name + " Office, Central, Lyon, France",
		},
	}
}

func TestGenerateCompanyWebsiteIsDeterministicAndKeepsCanonicalAddress(t *testing.T) {
	company := companyWebsiteFixture("company:atelier-1", "Ateliers de la Gare")
	first := GenerateCompanyWebsite("shared-world-seed", 4, company)
	second := GenerateCompanyWebsite("shared-world-seed", 4, company)
	if !reflect.DeepEqual(first, second) {
		t.Fatal("same world, version, and company produced different websites")
	}
	if first.Version != CompanyWebsiteProfileVersion {
		t.Fatalf("profile version = %d, want %d", first.Version, CompanyWebsiteProfileVersion)
	}
	changedWorldVersion := GenerateCompanyWebsite("shared-world-seed", 5, company)
	if reflect.DeepEqual(first, changedWorldVersion) {
		t.Fatal("changing the world generator version did not change the website profile")
	}
	if first.Address != company.Address {
		t.Fatalf("website address = %+v, want canonical address %+v", first.Address, company.Address)
	}
	if first.Account.Available {
		t.Fatal("diegetic account unexpectedly became functional")
	}
	for _, page := range first.Pages {
		if page.ID == "contact" {
			foundCanonicalAddress := false
			for _, section := range page.Sections {
				if section.Kind == "address" && section.Body == company.Address.Full {
					foundCanonicalAddress = true
				}
			}
			if !foundCanonicalAddress {
				t.Fatal("contact page does not show the canonical company address")
			}
		}
	}
	serialized, err := json.Marshal(first)
	if err != nil {
		t.Fatalf("marshal company website: %v", err)
	}
	var payload map[string]any
	if err := json.Unmarshal(serialized, &payload); err != nil {
		t.Fatalf("unmarshal company website: %v", err)
	}
	if _, exists := payload["employees"]; exists {
		t.Fatal("public website payload exposed an employee directory")
	}
}

func TestCompanyWebsiteProfilesVaryWithoutRepeatingVisualIdentity(t *testing.T) {
	seen := make(map[string]string, 4096)
	headers := make(map[string]bool)
	templates := make(map[string]bool)
	advertising := make(map[string]bool)
	for index := range 4096 {
		company := companyWebsiteFixture(fmt.Sprintf("company:%05d", index), fmt.Sprintf("Company %05d", index))
		site := GenerateCompanyWebsite("variety-world", 4, company)
		visualProfile, err := json.Marshal(struct {
			Template    string
			Advertising string
			Theme       CompanyWebsiteTheme
		}{site.Template, site.Advertising, site.Theme})
		if err != nil {
			t.Fatalf("marshal visual profile: %v", err)
		}
		key := string(visualProfile)
		if previous, exists := seen[key]; exists {
			t.Fatalf("companies %s and %s share a complete visual profile", previous, company.Company.ID)
		}
		seen[key] = company.Company.ID
		headers[site.Theme.Header] = true
		templates[site.Template] = true
		advertising[site.Advertising] = true
	}
	if len(headers) < 10 || len(templates) < 8 || len(advertising) != 4 {
		t.Fatalf("procedural catalog coverage = %d headers, %d templates, %d ad policies; want >=10, >=8, and 4", len(headers), len(templates), len(advertising))
	}
}

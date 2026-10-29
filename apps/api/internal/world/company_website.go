package world

import (
	"crypto/sha256"
	"encoding/binary"
	"encoding/hex"
	"fmt"
	"strings"
)

const CompanyWebsiteProfileVersion = 3

// CompanyWebsite is the deterministic public website profile generated for one company.
// Staff records are deliberately excluded and remain behind the existing access flow.
type CompanyWebsite struct {
	Version     int                   `json:"version"`
	Company     Company               `json:"company"`
	Address     Address               `json:"address"`
	Domain      string                `json:"domain"`
	Template    string                `json:"template"`
	Theme       CompanyWebsiteTheme   `json:"theme"`
	Advertising string                `json:"advertising"`
	Tagline     string                `json:"tagline"`
	About       string                `json:"about"`
	Pages       []CompanyWebsitePage  `json:"pages"`
	Account     CompanyWebsiteAccount `json:"account"`
}

type CompanyWebsiteTheme struct {
	Accent string `json:"accent"`
	Paper  string `json:"paper"`
	Ink    string `json:"ink"`
	Header string `json:"header"`
	Mark   string `json:"mark"`
	Layout string `json:"layout"`
}

type CompanyWebsitePage struct {
	ID       string                  `json:"id"`
	Title    string                  `json:"title"`
	Sections []CompanyWebsiteSection `json:"sections"`
}

type CompanyWebsiteSection struct {
	Kind    string   `json:"kind"`
	Heading string   `json:"heading"`
	Body    string   `json:"body"`
	Items   []string `json:"items,omitempty"`
}

type CompanyWebsiteAccount struct {
	Available bool   `json:"available"`
	Label     string `json:"label"`
	Message   string `json:"message"`
}

type companyWebsiteSector struct {
	name, tagline, about string
	offerings            []string
}

var companyWebsiteSectors = []companyWebsiteSector{
	{"TRADE & SUPPLY", "Practical goods, carefully selected.", "We serve local households and businesses with dependable goods and attentive service.", []string{"Household provisions", "Trade counter supplies", "Seasonal selections", "Delivery enquiries"}},
	{"PROFESSIONAL SERVICES", "Experience for the work that matters.", "Our office brings specialist knowledge and personal attention to the needs of every client.", []string{"Appointments and consultations", "Document preparation", "Business correspondence", "Client information"}},
	{"MANUFACTURING", "Made with care. Built to last.", "Our workshop combines established methods with reliable materials for customers across the region.", []string{"Workshop production", "Repair and maintenance", "Custom orders", "Trade enquiries"}},
	{"FOOD & HOSPITALITY", "A warm welcome, every day.", "We prepare and provide familiar local goods with care for our neighbours and visitors.", []string{"Daily selection", "Table reservations", "Private gatherings", "Local delivery"}},
	{"PROPERTY & BUILDING", "Building a better place to live and work.", "Our team supports property, maintenance and building projects throughout the surrounding districts.", []string{"Property enquiries", "Repairs and maintenance", "Building consultation", "Site appointments"}},
	{"TRANSPORT & LOGISTICS", "Keeping people and parcels in motion.", "We coordinate dependable local transport and delivery for private and commercial customers.", []string{"Local carriage", "Parcel collection", "Commercial schedules", "Route enquiries"}},
	{"HEALTH & PERSONAL CARE", "Personal service, close to home.", "Our local practice offers considerate service and practical advice to the community.", []string{"Appointments", "Personal care services", "Customer advice", "Opening information"}},
	{"GENERAL MERCHANDISE", "Serving the district since 1948.", "A locally operated business offering useful services and familiar assistance to our customers.", []string{"Customer service", "Product enquiries", "Local orders", "Office information"}},
}

var companyWebsiteBulletins = []string{
	"Our spring catalogue is available at the office.",
	"Please allow additional time for deliveries during market week.",
	"The customer desk will close early on the last Saturday of the month.",
	"New service hours have been posted at our local office.",
	"Telephone enquiries are answered during regular business hours.",
	"A printed list of current services may be requested in person.",
	"Our staff are preparing the next district notice for publication.",
	"The office will observe the municipal holiday schedule.",
}

// GenerateCompanyWebsite produces a stable, company-specific public site without persisting a copy.
func GenerateCompanyWebsite(seed string, generatorVersion int, company CompanyPortalRecord) CompanyWebsite {
	digest := companyWebsiteDigest(seed, generatorVersion, company.Company.ID)
	choose := func(length int, offset int) int {
		start := offset % (len(digest) - 7)
		return int(binary.BigEndian.Uint64(digest[start:start+8]) % uint64(length))
	}
	palette := []struct{ accent, paper, ink string }{
		{"#176b68", "#f4f2e8", "#172d30"}, {"#315c9a", "#f3f5f7", "#202b3e"},
		{"#9b4f32", "#f7f0e5", "#34251e"}, {"#6c568e", "#f4f0f6", "#30283a"},
		{"#8b671f", "#f6f1df", "#332d1c"}, {"#3f744f", "#f0f5ed", "#203126"},
		{"#a4424d", "#f8eeee", "#382326"}, {"#4b7180", "#eef4f4", "#203137"},
		{"#6f6f35", "#f4f3e4", "#303020"}, {"#9a5d7b", "#f6eff3", "#362632"},
		{"#475e47", "#f2f3ed", "#222b24"}, {"#8c5938", "#f6f0e8", "#35271f"},
	}
	templates := []string{"shop-window", "editorial", "corporate", "workshop", "catalogue", "directory", "classifieds", "bulletin", "ledger", "brochure"}
	headers := []string{
		"masthead-left", "masthead-centered", "split-banner", "compact-bar", "framed-title",
		"illustrated-banner", "catalogue-head", "office-sign", "ribbon", "split-logo",
		"ticker-strip", "notice-column", "letterhead", "serif-plaque", "double-rule",
		"side-rail", "shop-awning", "nameplate", "market-banner", "stacked-columns",
	}
	layouts := []string{"main-with-sidebar", "wide-columns", "stacked-sections", "feature-first", "index-and-detail", "catalogue-grid", "notice-board", "text-led"}
	accentSet := palette[choose(len(palette), 0)]
	slug := companyWebsiteSlug(company.Company.Name)
	if slug == "" {
		slug = "company"
	}
	companyToken := companyWebsiteDomainToken(digest)
	activity := companyWebsiteSectorForCompany(company.Company.Name, digest)
	template := companyWebsiteTemplate(activity, templates, choose(len(templates), 2))
	advertising := []string{"none", "sparse", "standard", "crowded"}[choose(4, 19)]
	bulletin := companyWebsiteBulletins[choose(len(companyWebsiteBulletins), 9)]
	offerings := rotateCompanyWebsiteItems(activity.offerings, choose(len(activity.offerings), 10))
	phone := fmt.Sprintf("%02d-%02d-%02d-%02d", 10+choose(80, 5), choose(100, 6), choose(100, 7), choose(100, 8))
	city := company.Address.City
	if city == "" {
		city = "the connected city"
	}
	return CompanyWebsite{
		Version:     CompanyWebsiteProfileVersion,
		Company:     company.Company,
		Address:     company.Address,
		Domain:      slug + "-" + companyToken + ".noir.local",
		Template:    template,
		Advertising: advertising,
		Theme: CompanyWebsiteTheme{
			Accent: accentSet.accent,
			Paper:  accentSet.paper,
			Ink:    accentSet.ink,
			Header: headers[choose(len(headers), 3)],
			Mark:   hex.EncodeToString(digest[:8]),
			Layout: layouts[choose(len(layouts), 4)],
		},
		Tagline: activity.tagline,
		About:   activity.about,
		Pages: []CompanyWebsitePage{
			{ID: "home", Title: "HOME", Sections: companyWebsiteHomeSections(template, company, activity, bulletin, offerings)},
			{ID: "about", Title: "ABOUT", Sections: []CompanyWebsiteSection{
				{Kind: "article", Heading: "OUR COMPANY", Body: activity.about},
				{Kind: "note", Heading: "SERVING " + strings.ToUpper(city), Body: fmt.Sprintf("%s is an independent company serving customers in %s and nearby districts.", company.Company.Name, city)},
			}},
			{ID: "services", Title: "SERVICES & PRODUCTS", Sections: []CompanyWebsiteSection{
				{Kind: "list", Heading: activity.name, Body: activity.tagline, Items: offerings},
				{Kind: "note", Heading: "CUSTOMER INFORMATION", Body: "For availability, prices and appointments, please contact the local office."},
			}},
			{ID: "contact", Title: "CONTACT & LOCATIONS", Sections: []CompanyWebsiteSection{
				{Kind: "address", Heading: "HEAD OFFICE", Body: company.Address.Full},
				{Kind: "fact", Heading: "TELEPHONE", Body: phone},
				{Kind: "fact", Heading: "OFFICE HOURS", Body: "Monday–Friday · 9:00 AM–5:00 PM"},
			}},
			{ID: "account", Title: "CUSTOMER ACCOUNT", Sections: []CompanyWebsiteSection{
				{Kind: "note", Heading: "CUSTOMER SIGN IN", Body: "Online account access is not available through the Meridian relay."},
			}},
		},
		Account: CompanyWebsiteAccount{Available: false, Label: "CUSTOMER ACCOUNT", Message: "Online account access is not available through the Meridian relay."},
	}
}

func companyWebsiteSearchMetadata(seed string, generatorVersion int, company CompanyPortalRecord) (string, string) {
	digest := companyWebsiteDigest(seed, generatorVersion, company.Company.ID)
	companyToken := companyWebsiteDomainToken(digest)
	sector := companyWebsiteSectorForCompany(company.Company.Name, digest)
	slug := companyWebsiteSlug(company.Company.Name)
	if slug == "" {
		slug = "company"
	}
	return slug + "-" + companyToken + ".noir.local", sector.tagline
}

func companyWebsiteSectorForCompany(name string, digest [32]byte) companyWebsiteSector {
	normalized := companyWebsiteSlug(name)
	sectors := map[string][]string{
		"MANUFACTURING":          {"atelier", "manufacture", "fonderie", "verrerie", "mecanique", "industrie", "works"},
		"TRANSPORT & LOGISTICS":  {"transport", "shipping", "messagerie", "courrier", "relay", "depot"},
		"PROFESSIONAL SERVICES":  {"service", "bureau", "edition", "exchange", "credit", "ledger", "assurance", "mutuelle"},
		"TRADE & SUPPLY":         {"maison", "comptoir", "magasin", "halles", "fourniture", "provision", "negociant", "entrepot", "cooperative"},
		"PROPERTY & BUILDING":    {"property", "properties", "building", "construction", "batiment"},
		"FOOD & HOSPITALITY":     {"hotel", "restaurant", "cafe", "boulangerie", "cuisine"},
		"HEALTH & PERSONAL CARE": {"pharmacie", "medical", "sante", "clinic"},
	}
	for _, sector := range companyWebsiteSectors {
		for _, keyword := range sectors[sector.name] {
			if strings.Contains(normalized, keyword) {
				return sector
			}
		}
	}
	index := binary.BigEndian.Uint64(digest[1:9]) % uint64(len(companyWebsiteSectors))
	return companyWebsiteSectors[index]
}

func companyWebsiteTemplate(sector companyWebsiteSector, templates []string, fallback int) string {
	preferred := map[string][]string{
		"TRADE & SUPPLY":         {"shop-window", "catalogue"},
		"PROFESSIONAL SERVICES":  {"corporate", "editorial"},
		"MANUFACTURING":          {"workshop", "catalogue"},
		"FOOD & HOSPITALITY":     {"shop-window", "editorial"},
		"PROPERTY & BUILDING":    {"corporate", "directory"},
		"TRANSPORT & LOGISTICS":  {"directory", "corporate"},
		"HEALTH & PERSONAL CARE": {"directory", "corporate"},
	}
	choices := preferred[sector.name]
	if len(choices) > 0 && fallback < 3 {
		return choices[fallback%len(choices)]
	}
	return templates[fallback]
}

func companyWebsiteDomainToken(digest [32]byte) string {
	return hex.EncodeToString(digest[24:32])
}

func companyWebsiteHomeSections(template string, company CompanyPortalRecord, sector companyWebsiteSector, bulletin string, offerings []string) []CompanyWebsiteSection {
	switch template {
	case "editorial":
		return []CompanyWebsiteSection{
			{Kind: "article", Heading: "A NOTE FROM THE OFFICE", Body: sector.about},
			{Kind: "bulletin", Heading: "FROM OUR NOTICE BOARD", Body: bulletin},
			{Kind: "address", Heading: "FIND US", Body: company.Address.Full},
		}
	case "corporate":
		return []CompanyWebsiteSection{
			{Kind: "article", Heading: "WHO WE ARE", Body: sector.about},
			{Kind: "list", Heading: "OUR SERVICES", Body: sector.tagline, Items: offerings},
			{Kind: "bulletin", Heading: "COMPANY NOTICE", Body: bulletin},
			{Kind: "address", Heading: "REGISTERED OFFICE", Body: company.Address.Full},
		}
	case "workshop":
		return []CompanyWebsiteSection{
			{Kind: "bulletin", Heading: "WORKSHOP NOTICE", Body: bulletin},
			{Kind: "list", Heading: "WORK UNDERTAKEN", Body: sector.tagline, Items: offerings},
			{Kind: "address", Heading: "WORKSHOP & COUNTER", Body: company.Address.Full},
		}
	case "catalogue":
		return []CompanyWebsiteSection{
			{Kind: "list", Heading: "THIS WEEK AT THE COUNTER", Body: sector.tagline, Items: offerings},
			{Kind: "bulletin", Heading: "CUSTOMER NOTICE", Body: bulletin},
			{Kind: "address", Heading: "VISIT OUR SHOP", Body: company.Address.Full},
		}
	case "directory":
		return []CompanyWebsiteSection{
			{Kind: "fact", Heading: "BUSINESS", Body: company.Company.Name},
			{Kind: "article", Heading: sector.name, Body: sector.about},
			{Kind: "bulletin", Heading: "LATEST NOTICE", Body: bulletin},
			{Kind: "address", Heading: "OFFICE ADDRESS", Body: company.Address.Full},
		}
	case "classifieds":
		return []CompanyWebsiteSection{
			{Kind: "bulletin", Heading: "NOTICES & SMALL ANNOUNCEMENTS", Body: bulletin},
			{Kind: "list", Heading: "GOODS AND SERVICES", Body: sector.tagline, Items: offerings},
			{Kind: "address", Heading: "LOCAL OFFICE", Body: company.Address.Full},
		}
	case "bulletin":
		return []CompanyWebsiteSection{
			{Kind: "bulletin", Heading: "NOTICE TO CUSTOMERS", Body: bulletin},
			{Kind: "article", Heading: "A WORD FROM THE MANAGER", Body: sector.about},
			{Kind: "address", Heading: "OFFICE & COUNTER", Body: company.Address.Full},
		}
	case "ledger":
		return []CompanyWebsiteSection{
			{Kind: "fact", Heading: "REGISTERED BUSINESS", Body: company.Company.Name},
			{Kind: "address", Heading: "REGISTERED ADDRESS", Body: company.Address.Full},
			{Kind: "list", Heading: "PUBLIC ENQUIRIES", Body: sector.tagline, Items: offerings[:3]},
		}
	case "brochure":
		return []CompanyWebsiteSection{
			{Kind: "article", Heading: "WELCOME TO OUR ESTABLISHMENT", Body: sector.about},
			{Kind: "list", Heading: "WHAT WE OFFER", Body: sector.tagline, Items: offerings},
			{Kind: "address", Heading: "PLAN YOUR VISIT", Body: company.Address.Full},
		}
	default:
		return []CompanyWebsiteSection{
			{Kind: "article", Heading: "WELCOME", Body: sector.tagline},
			{Kind: "bulletin", Heading: "LATEST INFORMATION", Body: bulletin},
			{Kind: "address", Heading: "VISIT OUR OFFICE", Body: company.Address.Full},
		}
	}
}

func rotateCompanyWebsiteItems(items []string, start int) []string {
	rotated := make([]string, 0, len(items))
	for offset := range items {
		rotated = append(rotated, items[(start+offset)%len(items)])
	}
	return rotated
}

func companyWebsiteDigest(seed string, generatorVersion int, companyID string) [32]byte {
	identity := fmt.Sprintf("noir-company-site\x00%d\x00%s\x00%d\x00%s", CompanyWebsiteProfileVersion, seed, generatorVersion, companyID)
	return sha256.Sum256([]byte(identity))
}

func companyWebsiteSlug(name string) string {
	name = strings.NewReplacer(
		"á", "a", "à", "a", "â", "a", "ä", "a", "ã", "a", "å", "a",
		"ç", "c", "é", "e", "è", "e", "ê", "e", "ë", "e",
		"í", "i", "ì", "i", "î", "i", "ï", "i", "ñ", "n",
		"ó", "o", "ò", "o", "ô", "o", "ö", "o", "õ", "o", "ø", "o",
		"ú", "u", "ù", "u", "û", "u", "ü", "u", "ý", "y", "ÿ", "y",
		"æ", "ae", "œ", "oe", "ß", "ss", "ł", "l", "đ", "d", "þ", "th",
	).Replace(strings.ToLower(name))
	var slug strings.Builder
	separator := false
	for _, char := range strings.ToLower(name) {
		if char >= 'a' && char <= 'z' || char >= '0' && char <= '9' {
			slug.WriteRune(char)
			separator = false
		} else if !separator && slug.Len() > 0 {
			slug.WriteByte('-')
			separator = true
		}
	}
	return strings.Trim(slug.String(), "-")
}

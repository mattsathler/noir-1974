package world

import (
	"crypto/sha256"
	"encoding/binary"
	"strconv"
	"strings"
)

const CompanyTypeCatalogVersion = 1

type CompanyBusinessType struct {
	ID          string   `json:"id"`
	Label       string   `json:"label"`
	Description string   `json:"description"`
	Category    string   `json:"category"`
	Roles       []string `json:"-"`
	Offerings   []string `json:"-"`
}

// Catalog entries are stable product/activity profiles. The ID is selected from
// the world seed and company identity, never from the generated display name.
var companyTypeCatalog = []CompanyBusinessType{
	{"laboratory", "Laboratory", "Scientific and industrial laboratory services.", "professional", []string{"laboratory director", "chemist", "laboratory technician", "research assistant", "clerk"}, []string{"Sample analysis", "Laboratory testing", "Research enquiries"}},
	{"accounting", "Accountancy Office", "Bookkeeping, payroll and business accounting.", "professional", []string{"accountant", "bookkeeper", "payroll clerk", "auditor", "secretary"}, []string{"Bookkeeping", "Payroll preparation", "Tax accounts"}},
	{"technology-office", "Technology Office", "Computing equipment, technical services and data processing.", "professional", []string{"technical director", "programmer", "systems operator", "repair technician", "clerk"}, []string{"Data processing", "Equipment repair", "Technical consultation"}},
	{"bank", "Bank", "Savings, credit and financial settlement services.", "professional", []string{"branch manager", "accountant", "bank clerk", "cashier", "messenger"}, []string{"Savings accounts", "Cashier services", "Business credit"}},
	{"insurance", "Insurance Office", "Insurance policies, claims and risk assessment.", "professional", []string{"office manager", "claims adjuster", "insurance clerk", "underwriter", "typist"}, []string{"Policy enquiries", "Claims assessment", "Premium collection"}},
	{"legal-office", "Legal Office", "Legal counsel, contracts and representation.", "professional", []string{"senior partner", "solicitor", "legal clerk", "research assistant", "secretary"}, []string{"Legal consultation", "Contract preparation", "Representation"}},
	{"consultancy", "Consultancy", "Independent business and organizational advice.", "professional", []string{"managing consultant", "business consultant", "research analyst", "office clerk", "secretary"}, []string{"Business advice", "Operational review", "Written reports"}},
	{"architecture-engineering", "Architecture & Engineering", "Building design, technical plans and engineering surveys.", "property", []string{"chief engineer", "architect", "draftsman", "site surveyor", "office clerk"}, []string{"Building plans", "Engineering surveys", "Project supervision"}},
	{"construction", "Construction Company", "Building, renovation and civil works.", "property", []string{"site manager", "civil engineer", "foreman", "bricklayer", "driver"}, []string{"Building contracts", "Renovation", "Road works"}},
	{"metalworks", "Metalworks", "Metal casting, fabrication and industrial components.", "manufacturing", []string{"works manager", "foundry foreman", "metalworker", "machinist", "storekeeper"}, []string{"Metal casting", "Fabricated components", "Machine parts"}},
	{"textile-mill", "Textile Mill", "Textile production, weaving and garment materials.", "manufacturing", []string{"mill manager", "weaving supervisor", "loom operator", "seamstress", "warehouse clerk"}, []string{"Woven cloth", "Dyeing services", "Workwear orders"}},
	{"chemical-works", "Chemical Works", "Industrial chemical production and material processing.", "manufacturing", []string{"plant manager", "chemist", "process operator", "safety inspector", "storekeeper"}, []string{"Industrial solvents", "Material treatment", "Chemical supplies"}},
	{"publishing-house", "Publishing House", "Books, periodicals and editorial publishing.", "professional", []string{"publisher", "editor", "proofreader", "typesetter", "rights clerk"}, []string{"Book publishing", "Editorial services", "Subscriptions"}},
	{"printing-house", "Printing House", "Commercial printing, typesetting and binding.", "manufacturing", []string{"print manager", "typesetter", "press operator", "binder", "delivery driver"}, []string{"Commercial printing", "Letterpress work", "Binding"}},
	{"newspaper", "Newspaper Office", "Local reporting, printing and newspaper circulation.", "professional", []string{"editor", "reporter", "copy editor", "press operator", "circulation clerk"}, []string{"Daily newspaper", "Local advertisements", "Press notices"}},
	{"radio-telecom", "Radio & Telecommunications", "Broadcasting and wired communication services.", "transport", []string{"station director", "broadcaster", "radio engineer", "switchboard operator", "newsreader"}, []string{"Radio programmes", "Commercial broadcasts", "Telegraph services"}},
	{"transport-company", "Transport Company", "Passenger transport and local freight services.", "transport", []string{"transport manager", "dispatcher", "driver", "mechanic", "booking clerk"}, []string{"Passenger carriage", "Freight delivery", "Route schedules"}},
	{"railway", "Railway Company", "Rail passenger, freight and station operations.", "transport", []string{"station master", "railway engineer", "signal operator", "porter", "ticket clerk"}, []string{"Passenger tickets", "Freight carriage", "Station services"}},
	{"shipping-cargo", "Shipping & Cargo", "Port handling, coastal shipping and freight forwarding.", "transport", []string{"harbour master", "shipping agent", "cargo supervisor", "dock worker", "freight clerk"}, []string{"Cargo forwarding", "Port handling", "Coastal freight"}},
	{"airline", "Airline", "Scheduled passenger and air freight services.", "transport", []string{"airport manager", "pilot", "air mechanic", "steward", "booking clerk"}, []string{"Passenger bookings", "Air freight", "Charter services"}},
	{"hotel", "Hotel", "Accommodation, dining and visitor services.", "hospitality", []string{"hotel manager", "receptionist", "cook", "porter", "housekeeper"}, []string{"Guest rooms", "Dining room", "Conference bookings"}},
	{"restaurant", "Restaurant", "Prepared meals, table service and catering.", "hospitality", []string{"restaurant manager", "head cook", "cook", "waiter", "cashier"}, []string{"Dining service", "Private suppers", "Outside catering"}},
	{"food-retail", "Food Retailer", "Groceries, household provisions and food delivery.", "trade", []string{"shop manager", "butcher", "grocer", "stock clerk", "delivery driver"}, []string{"Household provisions", "Fresh produce", "Local delivery"}},
	{"department-store", "Department Store", "General household goods, clothing and retail services.", "trade", []string{"store manager", "department supervisor", "sales clerk", "stock clerk", "cashier"}, []string{"Clothing", "Household goods", "Customer orders"}},
	{"mechanic-workshop", "Motor Workshop", "Vehicle repair, servicing and mechanical maintenance.", "manufacturing", []string{"workshop manager", "master mechanic", "mechanic", "apprentice", "parts clerk"}, []string{"Vehicle repairs", "Routine servicing", "Spare parts"}},
	{"automobile-dealer", "Motor Dealer", "New and used vehicle sales and customer servicing.", "trade", []string{"sales manager", "vehicle salesman", "mechanic", "accounts clerk", "receptionist"}, []string{"New vehicles", "Used cars", "Trade-in appraisals"}},
	{"electric-utility", "Electric Utility", "Electricity generation, distribution and maintenance.", "utilities", []string{"works superintendent", "electrical engineer", "line technician", "plant operator", "billing clerk"}, []string{"Electric supply", "Meter installation", "Line maintenance"}},
	{"water-utility", "Water Utility", "Water treatment, supply and municipal pipe maintenance.", "utilities", []string{"works superintendent", "water engineer", "plant operator", "pipe fitter", "billing clerk"}, []string{"Water supply", "Pipe repairs", "Water testing"}},
	{"hospital-clinic", "Hospital & Clinic", "Medical care, diagnosis and patient treatment.", "health", []string{"medical director", "physician", "nurse", "laboratory technician", "records clerk"}, []string{"Outpatient appointments", "Medical examinations", "Patient care"}},
	{"pharmacy", "Pharmacy", "Prescription preparation and household medical supplies.", "health", []string{"pharmacist", "assistant pharmacist", "dispensary clerk", "stock clerk", "cashier"}, []string{"Prescription service", "Medical supplies", "Health advice"}},
	{"school", "School", "Primary and secondary education and public instruction.", "professional", []string{"head teacher", "teacher", "school nurse", "librarian", "school secretary"}, []string{"Class enrolment", "Evening instruction", "School library"}},
	{"research-institute", "Research Institute", "Independent research, studies and technical publications.", "professional", []string{"research director", "researcher", "statistician", "laboratory assistant", "records clerk"}, []string{"Research studies", "Technical reports", "Public lectures"}},
	{"agricultural-cooperative", "Agricultural Cooperative", "Farm supplies, produce collection and cooperative sales.", "trade", []string{"cooperative manager", "agronomist", "produce buyer", "warehouse worker", "accounts clerk"}, []string{"Seed and equipment", "Produce collection", "Cooperative sales"}},
	{"mining-company", "Mining Company", "Mineral extraction, processing and ore transport.", "manufacturing", []string{"mine manager", "mining engineer", "shift foreman", "miner", "safety inspector"}, []string{"Mineral extraction", "Ore processing", "Industrial supply"}},
	{"cinema-theatre", "Cinema & Theatre", "Film exhibition, stage performances and public entertainment.", "hospitality", []string{"venue manager", "stage director", "projectionist", "usher", "ticket clerk"}, []string{"Film screenings", "Stage performances", "Private bookings"}},
	{"private-security", "Private Security", "Guarding, alarm response and commercial security services.", "professional", []string{"security director", "operations supervisor", "security guard", "alarm operator", "dispatch clerk"}, []string{"Premises guarding", "Alarm response", "Security consultation"}},
	{"cleaning-maintenance", "Cleaning & Maintenance", "Commercial cleaning, repairs and building upkeep.", "property", []string{"service manager", "maintenance foreman", "electrician", "cleaner", "supply clerk"}, []string{"Office cleaning", "Building repairs", "Scheduled maintenance"}},
}

func BusinessTypeForCompany(seed string, version int, companyID string) CompanyBusinessType {
	if len(companyTypeCatalog) == 0 {
		return CompanyBusinessType{}
	}
	identity := strings.Join([]string{"noir-company-type", strconv.Itoa(CompanyTypeCatalogVersion), seed, strconv.Itoa(version), companyID}, "\x00")
	digest := sha256.Sum256([]byte(identity))
	index := binary.BigEndian.Uint64(digest[:8]) % uint64(len(companyTypeCatalog))
	typeProfile := companyTypeCatalog[index]
	typeProfile.Roles = append([]string(nil), typeProfile.Roles...)
	typeProfile.Offerings = append([]string(nil), typeProfile.Offerings...)
	return typeProfile
}

func CompanyBusinessTypeByID(id string) (CompanyBusinessType, bool) {
	for _, typeProfile := range companyTypeCatalog {
		if typeProfile.ID == id {
			return typeProfile, true
		}
	}
	return CompanyBusinessType{}, false
}

func companyRoleAt(company Company, slot int) string {
	roles := company.BusinessType.Roles
	if len(roles) == 0 {
		roles = employmentRoles
	}
	return roles[slot%len(roles)]
}

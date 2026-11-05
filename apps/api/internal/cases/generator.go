package cases

import (
	"crypto/sha256"
	"encoding/binary"
	"encoding/hex"
	"errors"
	"fmt"
	"sort"
	"strings"
	"time"

	"noir-1974/apps/api/internal/mail"
	"noir-1974/apps/api/internal/records"
	"noir-1974/apps/api/internal/world"
)

const GeneratorVersion = 9
const previousSerialGeneratorVersion = 8
const previousFinancialGeneratorVersion = 7
const connectedGeneratorVersion = 4
const contextualGeneratorVersion = 6
const previousGeneratorVersion = 5
const legacyGeneratorVersion = 3

var (
	ErrInvalidCaseSeed       = errors.New("case seed is required")
	ErrUnsupportedGenerator  = errors.New("unsupported case generator version")
	ErrNoValidCaseCandidates = errors.New("world has no valid crime candidates")
	ErrInvalidDefinition     = errors.New("invalid case definition")
	ErrUnsolvableCase        = errors.New("case has no reachable clue")
)

type Definition struct {
	ID               string                  `json:"id"`
	CityID           string                  `json:"cityId"`
	Seed             string                  `json:"caseSeed"`
	GeneratorVersion int                     `json:"generatorVersion"`
	CaseType         string                  `json:"caseType"`
	ArchetypeID      string                  `json:"archetypeId"`
	CompanyID        string                  `json:"-"`
	Participants     []CaseParticipant       `json:"participants,omitempty"`
	Motive           string                  `json:"motive"`
	VictimID         string                  `json:"victimId"`
	KillerID         string                  `json:"killerId"`
	LocationID       string                  `json:"locationId"`
	CrimeAt          time.Time               `json:"-"`
	ParamourID       string                  `json:"-"`
	FutureVictimIDs  []string                `json:"-"`
	Relationships    []RelationshipReference `json:"relationships"`
	Evidence         []EvidenceDefinition    `json:"evidence"`
	Records          []records.Record        `json:"records,omitempty"`
	Emails           []mail.Message          `json:"-"`
}

type CaseParticipant struct {
	NPCID string `json:"npcId"`
	Role  string `json:"role"`
}

type RelationshipReference struct {
	ID   string `json:"id"`
	Type string `json:"type"`
	Role string `json:"role"`
}

type EvidenceDefinition struct {
	ID                     string   `json:"id"`
	Type                   string   `json:"type"`
	Presentation           string   `json:"presentation"`
	Content                string   `json:"content"`
	LocationID             string   `json:"locationId"`
	RelationshipIDs        []string `json:"relationshipIds"`
	ReachableFromEntityIDs []string `json:"reachableFromEntityIds"`
	SupportsNPCIDs         []string `json:"supportsNpcIds,omitempty"`
}

type employmentLink struct {
	id       string
	employee string
}

type candidate struct {
	caseType        string
	archetypeID     string
	companyID       string
	motive          string
	relationships   []RelationshipReference
	evidence        []evidencePlan
	killerID        string
	accompliceID    string
	victimID        string
	locationID      string
	paramourID      string
	futureVictimIDs []string
}

type evidencePlan struct {
	typeID       string
	presentation string
	content      string
	supports     []string
}

type archetypeGenerator struct {
	id            string
	candidates    func(world.World) []candidate
	validate      func(Definition, world.World) bool
	evidenceTypes []string
}

var archetypeGenerators = []archetypeGenerator{
	{id: "workplace_conflict", candidates: workplaceCandidates, validate: validateWorkplaceConflict, evidenceTypes: []string{"employment_record", "meeting_minutes", "office_log"}},
	{id: "jealousy", candidates: jealousyCandidates, validate: validateJealousy, evidenceTypes: []string{"civil_record", "letter_fragment", "witness_statement"}},
	{id: "financial_extortion", candidates: extortionCandidates, validate: validateExtortion, evidenceTypes: []string{"account_record", "travel_register", "personnel_record"}},
	{id: "serial_killer", candidates: serialKillerCandidates, validate: validateSerialKiller, evidenceTypes: []string{"letter_fragment", "witness_statement", "office_log"}},
}

func init() {
	archetypeGenerators = append(archetypeGenerators, newHomicideMotiveGenerators()...)
}

type homicideMotive struct {
	id     string
	motive string
	clues  [3]string
}

var newHomicideMotives = []homicideMotive{
	{id: "inheritance", motive: "The victim was killed after a disputed inheritance threatened to change who controlled a company estate.", clues: [3]string{"A sealed estate instruction was copied before it reached the company registrar.", "The victim's notes identify a recent challenge to the expected inheritance.", "The registrar's access book places a company colleague with the altered copy."}},
	{id: "revenge", motive: "A workplace grievance became a planned act of revenge against a colleague.", clues: [3]string{"A dated complaint records the principal's grievance against the victim.", "A private note refers to settling the dispute after the victim rejected an appeal.", "The office access register places the principal near the victim after hours."}},
	{id: "silence_witness", motive: "The victim was killed after discovering evidence of misconduct and preparing to report it.", clues: [3]string{"The victim's draft statement names an irregular company transfer.", "A missing page from the inquiry file was checked out by the principal.", "A copied ledger page connects the file and the transfer to the same employee."}},
	{id: "conceal_fraud", motive: "The victim was killed to prevent exposure of a financial scheme involving company records.", clues: [3]string{"A ledger copy shows figures changed after the victim's review.", "A correction slip bears the principal's authorization initials.", "Meeting minutes show the victim planned to present the discrepancy."}},
	{id: "power_struggle", motive: "The victim's removal would secure the principal's control of a company office and its records.", clues: [3]string{"A succession memorandum names the victim as the next office holder.", "A private agenda records the principal's objection to that appointment.", "A staff access log places the principal in the records room before the appointment vote."}},
	{id: "mistaken_identity", motive: "The killer mistook the victim for the intended target after acting on an incomplete description.", clues: [3]string{"A note describes a target by workplace clothing and shift rather than by name.", "A witness account places the victim in the target's borrowed coat.", "The staff roster shows the intended colleague had exchanged shifts with the victim."}},
	{id: "blackmail", motive: "The victim was killed after using private correspondence to demand money and threaten a disclosure.", clues: [3]string{"A draft demand requests payment before a named disclosure date.", "A retained envelope links the correspondence to the principal's office.", "The victim's appointment book records a meeting with the principal shortly before the deadline."}},
	{id: "debt", motive: "A private debt between colleagues escalated when the victim demanded repayment.", clues: [3]string{"A signed IOU records an overdue private loan to the victim.", "The principal's payroll advance request cites an urgent personal obligation.", "A witness statement places both colleagues together after a repayment demand."}},
	{id: "protect_reputation", motive: "The victim was killed to keep a damaging personal or professional scandal from becoming public.", clues: [3]string{"A sealed complaint describes a concealed company incident.", "The principal requested the complaint file shortly before it disappeared.", "The victim's appointment notes show an imminent meeting with the newspaper editor."}},
	{id: "conceal_accident", motive: "A company colleague killed the victim and staged the scene to resemble an accident.", clues: [3]string{"A maintenance report records that the equipment was inspected before the supposed failure.", "The principal amended the inspection log after the death.", "A shift register places the principal at the site during the staged failure and the later cleanup."}},
}

func newHomicideMotiveGenerators() []archetypeGenerator {
	result := make([]archetypeGenerator, 0, len(newHomicideMotives))
	for _, motive := range newHomicideMotives {
		selected := motive
		result = append(result, archetypeGenerator{
			id: selected.id,
			candidates: func(base world.World) []candidate {
				items := workplaceCandidates(base)
				for index := range items {
					items[index].archetypeID = selected.id
					items[index].motive = selected.motive
					principal, victim := items[index].killerID, items[index].victimID
					items[index].evidence = []evidencePlan{
						{typeID: selected.id + "_document", presentation: "Investigative document", content: selected.clues[0], supports: []string{principal}},
						{typeID: selected.id + "_account", presentation: "Company record", content: selected.clues[1], supports: []string{victim}},
						{typeID: selected.id + "_corroboration", presentation: "Corroborating record", content: selected.clues[2], supports: []string{principal, victim}},
					}
				}
				return items
			},
			validate: func(definition Definition, base world.World) bool {
				return definition.ArchetypeID == selected.id && validateWorkplaceConflict(definition, base)
			},
			evidenceTypes: []string{selected.id + "_document", selected.id + "_account", selected.id + "_corroboration"},
		})
	}
	return result
}

// Generate creates a deterministic crime definition from existing world facts.
func Generate(base world.World, seed string, version int) (Definition, error) {
	return GenerateExcluding(base, seed, version, nil)
}

// GenerateExcluding creates a case without reusing NPCs assigned to prior cases.
func GenerateExcluding(base world.World, seed string, version int, excluded map[string]struct{}) (Definition, error) {
	if strings.TrimSpace(seed) == "" {
		return Definition{}, ErrInvalidCaseSeed
	}
	if version != GeneratorVersion && version != previousSerialGeneratorVersion && version != previousFinancialGeneratorVersion && version != contextualGeneratorVersion && version != previousGeneratorVersion && version != connectedGeneratorVersion && version != legacyGeneratorVersion {
		return Definition{}, fmt.Errorf("%w: %d", ErrUnsupportedGenerator, version)
	}
	candidateWorld := base
	if len(excluded) > 0 {
		candidateWorld.NPCs = make([]world.NPC, 0, len(base.NPCs))
		for _, npc := range base.NPCs {
			if _, used := excluded[npc.ID]; !used {
				candidateWorld.NPCs = append(candidateWorld.NPCs, npc)
			}
		}
		candidateWorld.EmploymentRelationships = make([]world.EmploymentRelationship, 0, len(base.EmploymentRelationships))
		for _, relation := range base.EmploymentRelationships {
			if _, used := excluded[relation.EmployeeNPCID]; !used {
				candidateWorld.EmploymentRelationships = append(candidateWorld.EmploymentRelationships, relation)
			}
		}
		candidateWorld.PersonalRelationships = make([]world.PersonalRelationship, 0, len(base.PersonalRelationships))
		for _, relation := range base.PersonalRelationships {
			_, usedA := excluded[relation.PersonANPCID]
			_, usedB := excluded[relation.PersonBNPCID]
			if !usedA && !usedB {
				candidateWorld.PersonalRelationships = append(candidateWorld.PersonalRelationships, relation)
			}
		}
	}
	var selected candidate
	if version >= previousFinancialGeneratorVersion {
		choices := currentCaseChoices(candidateWorld, version)
		if len(choices) == 0 {
			return Definition{}, ErrNoValidCaseCandidates
		}
		choice := choices[caseIndex(seed+"/archetype", version, len(choices))]
		selected = choice.candidates[caseIndex(seed+"/"+choice.id, version, len(choice.candidates))]
	} else {
		available := availableArchetypesForVersion(candidateWorld, version)
		if len(available) == 0 {
			return Definition{}, ErrNoValidCaseCandidates
		}
		archetype := available[caseIndex(seed+"/archetype", version, len(available))]
		candidates := archetypeCandidates(archetype, candidateWorld, version)
		selected = candidates[caseIndex(seed+"/"+archetype.id, version, len(candidates))]
	}
	caseID := makeCaseID(seed, version)
	generatedAt := time.Now()
	caseDate := generatedAt.UTC().Truncate(24 * time.Hour)
	caseDay := caseDate
	caseCityID := cityForLocation(base, selected.locationID)
	if selected.caseType != "" {
		caseCityID = cityForCompany(base, selected.companyID)
	}
	var cityLocation *time.Location
	if version >= contextualGeneratorVersion {
		caseDay = cityLocalDayInCity(caseCityID, generatedAt)
		if city, ok := world.CityByID(caseCityID); ok {
			if location, err := time.LoadLocation(city.TimeZone); err == nil {
				cityLocation = location
			}
		}
	}
	crimeAt := caseDay.Add(caseTimeOffset(seed, selected.archetypeID, version))
	if cityLocation != nil && crimeAt.After(generatedAt.UTC()) {
		// Cases are generated before their evening crime time; keep generated records historical.
		crimeAt = crimeAt.In(cityLocation).AddDate(0, 0, -1).UTC()
	}
	definition := Definition{
		ID:               caseID,
		CityID:           caseCityID,
		Seed:             seed,
		GeneratorVersion: version,
		CaseType:         selected.caseType,
		ArchetypeID:      selected.archetypeID,
		CompanyID:        selected.companyID,
		Motive:           selected.motive,
		VictimID:         selected.victimID,
		KillerID:         selected.killerID,
		LocationID:       selected.locationID,
		Relationships:    selected.relationships,
		Evidence:         buildEvidence(caseID, selected),
	}
	if definition.CaseType == "" {
		definition.CaseType = "homicide"
	} else {
		definition.LocationID = ""
	}
	if selected.caseType == "" {
		definition.Participants = []CaseParticipant{{NPCID: selected.killerID, Role: "principal"}, {NPCID: selected.victimID, Role: "victim"}}
	} else {
		definition.Participants = []CaseParticipant{{NPCID: selected.killerID, Role: "principal"}, {NPCID: selected.accompliceID, Role: "accomplice"}}
	}
	if selected.caseType == "" && version >= connectedGeneratorVersion {
		definition.CrimeAt = crimeAt
		definition.ParamourID = selected.paramourID
		definition.FutureVictimIDs = append([]string(nil), selected.futureVictimIDs...)
	}
	if selected.caseType == "" && version == legacyGeneratorVersion && selected.archetypeID == "financial_extortion" {
		region := regionName(base, npcByID(base.NPCs, selected.killerID).RegionID)
		definition.Records = records.ExtortionEvents(seed, base.GeneratorVersion, caseID, selected.killerID, selected.victimID, region, caseDate)
	}
	if selected.caseType != "" {
		definition.CrimeAt = crimeAt
		definition.Records = buildFinancialCaseRecords(seed, caseID, selected, crimeAt)
	} else if version >= connectedGeneratorVersion {
		definition.Records = buildConnectedCaseRecords(seed, base.GeneratorVersion, caseID, selected, base, crimeAt, version)
		if selected.archetypeID == "serial_killer" {
			definition.Emails = buildSerialKillerEmails(seed, base.GeneratorVersion, caseID, selected, base, crimeAt)
		} else {
			definition.Emails = buildConnectedCaseEmails(seed, base.GeneratorVersion, caseID, selected, base, crimeAt)
		}
	} else {
		definition.Emails = buildCaseEmails(seed, base.GeneratorVersion, caseID, selected, base, caseDate)
	}
	if err := Validate(definition, base); err != nil {
		return Definition{}, err
	}
	return definition, nil
}

func buildFinancialCaseRecords(seed, caseID string, selected candidate, crimeAt time.Time) []records.Record {
	amount := 2750 + keyIndex(seed, "corporate-value", 9251)
	reference := fmt.Sprintf("company settlement %06d", keyIndex(seed, "corporate-reference", 1000000))
	principal, accomplice := selected.killerID, selected.accompliceID
	descriptionOut, descriptionIn := "A company settlement was issued outside the approved ledger.", "A matching company settlement cleared through a secondary account."
	internalRelation := "embezzlement_transfer"
	switch selected.caseType {
	case "financial_extortion":
		descriptionOut, descriptionIn = "A private company payment cleared under an unregistered settlement reference.", "A matching private remittance was received shortly after the company payment."
		internalRelation = "extortion_payment"
	case "accounting_fraud":
		descriptionOut, descriptionIn = "A revenue adjustment was credited after a revised company declaration was filed.", "A matching adjustment was transferred into a secondary account."
		internalRelation = "accounting_adjustment"
	default:
		descriptionOut, descriptionIn = "A payment connected to an open inquiry was issued outside the approved ledger.", "A matching payment connected to the inquiry cleared through a secondary account."
		internalRelation = selected.caseType + "_payment"
	}
	return []records.Record{
		{ID: caseRecordID(caseID, internalRelation+"-source", principal), NPCID: principal, RelatedNPCID: accomplice, Source: "bank", OccurredAt: crimeAt.Add(-24 * time.Hour), Operator: "Union Mercantile Settlement Office", Description: descriptionOut, Reference: reference, Amount: amount, Direction: "outgoing", CaseID: caseID, InternalRelation: internalRelation},
		{ID: caseRecordID(caseID, internalRelation+"-recipient", accomplice), NPCID: accomplice, RelatedNPCID: principal, Source: "bank", OccurredAt: crimeAt.Add(-24*time.Hour + 8*time.Minute), Operator: "Union Mercantile Settlement Office", Description: descriptionIn, Reference: reference, Amount: amount, Direction: "incoming", CaseID: caseID, InternalRelation: internalRelation},
	}
}

type caseChoice struct {
	id         string
	candidates []candidate
}

func currentCaseChoices(base world.World, version int) []caseChoice {
	choices := make([]caseChoice, 0, len(archetypeGenerators)+3)
	for _, archetype := range archetypeGenerators {
		// v8 historically had the serial option in both this registry and the
		// explicit append below. Preserve that choice set for old seeds; v9 fixes it.
		if version >= GeneratorVersion && archetype.id == "serial_killer" {
			continue
		}
		candidates := archetypeCandidates(archetype, base, version)
		if len(candidates) > 0 {
			choices = append(choices, caseChoice{id: "homicide/" + archetype.id, candidates: candidates})
		}
	}
	for _, caseType := range []string{"embezzlement", "financial_extortion", "accounting_fraud"} {
		candidates := financialCandidates(base, caseType)
		if len(candidates) > 0 {
			choices = append(choices, caseChoice{id: caseType, candidates: candidates})
		}
	}
	if version >= GeneratorVersion {
		for _, caseType := range expandedCrimeTypes {
			candidates := financialCandidates(base, caseType)
			if len(candidates) > 0 {
				choices = append(choices, caseChoice{id: caseType, candidates: candidates})
			}
		}
	}
	if version >= previousSerialGeneratorVersion {
		if candidates := serialKillerCandidates(base); len(candidates) > 0 {
			choices = append(choices, caseChoice{id: "homicide/serial_killer", candidates: candidates})
		}
	}
	return choices
}

// expandedCrimeTypes are non-homicide investigations that can be grounded in
// existing company, employee, and headquarters relationships.
var expandedCrimeTypes = []string{
	"insurance_fraud", "investment_fraud", "money_laundering", "financial_identity_theft",
	"procurement_fraud", "forged_will", "cargo_theft", "art_forgery", "antiquities_smuggling",
	"correspondence_blackmail", "industrial_sabotage", "arson", "staged_disappearance", "confidential_document_theft", "police_corruption",
}

func financialCrimeType(caseType string) bool {
	return contains(append([]string{"embezzlement", "financial_extortion", "accounting_fraud"}, expandedCrimeTypes...), caseType)
}

func financialCandidates(base world.World, caseType string) []candidate {
	entities := make(map[string]string, len(base.Entities))
	for _, entity := range base.Entities {
		entities[entity.ID] = entity.EntityType
	}
	hqByCompany := make(map[string]string)
	for _, location := range base.CompanyLocations {
		if location.LocationRole == "headquarters" && entities[location.LocationID] == "location" {
			hqByCompany[location.CompanyID] = location.LocationID
		}
	}
	companyEmployees := make(map[string][]employmentLink)
	for _, relation := range base.EmploymentRelationships {
		if entities[relation.EmployeeNPCID] == "npc" && entities[relation.EmployerCompanyID] == "company" {
			companyEmployees[relation.EmployerCompanyID] = append(companyEmployees[relation.EmployerCompanyID], employmentLink{id: relation.ID, employee: relation.EmployeeNPCID})
		}
	}
	companyIDs := make([]string, 0, len(companyEmployees))
	for companyID := range companyEmployees {
		companyIDs = append(companyIDs, companyID)
	}
	sort.Strings(companyIDs)
	result := make([]candidate, 0)
	for _, companyID := range companyIDs {
		locationID := hqByCompany[companyID]
		staff := companyEmployees[companyID]
		if locationID == "" || len(staff) < 2 {
			continue
		}
		sort.Slice(staff, func(i, j int) bool { return staff[i].employee < staff[j].employee })
		for principalIndex := range staff {
			for accompliceIndex := range staff {
				if principalIndex == accompliceIndex {
					continue
				}
				principal, accomplice := staff[principalIndex], staff[accompliceIndex]
				item := candidate{
					caseType: caseType, archetypeID: caseType, companyID: companyID,
					killerID: principal.employee, accompliceID: accomplice.employee, locationID: locationID,
					relationships: []RelationshipReference{
						{ID: principal.id, Type: "employment", Role: "principal_employment"},
						{ID: accomplice.id, Type: "employment", Role: "accomplice_employment"},
					},
				}
				switch caseType {
				case "embezzlement":
					item.motive = "Company funds were redirected through a false settlement and a second employee helped conceal the transfer."
					item.evidence = []evidencePlan{
						{typeID: "settlement_ledger", presentation: "Settlement ledger", content: "A numbered company transfer was entered under the principal's authorization code.", supports: []string{principal.employee}},
						{typeID: "duplicate_receipt", presentation: "Duplicate receipt", content: "A receipt copy was removed from the filing packet after the transfer cleared.", supports: []string{accomplice.employee}},
						{typeID: "account_reconciliation", presentation: "Account reconciliation", content: "The company total differs from the bank settlement by the same amount and reference.", supports: []string{principal.employee, accomplice.employee}},
					}
				case "financial_extortion":
					item.motive = "A company employee used confidential financial information to demand an unauthorized payment, with a colleague arranging collection."
					item.evidence = []evidencePlan{
						{typeID: "payment_demand", presentation: "Private payment demand", content: "A demand for an unrecorded payment refers to a restricted company account.", supports: []string{principal.employee}},
						{typeID: "collection_record", presentation: "Collection record", content: "A sealed envelope was logged under a second employee's initials.", supports: []string{accomplice.employee}},
						{typeID: "account_reconciliation", presentation: "Account reconciliation", content: "An irregular transfer matches the amount named in the private demand.", supports: []string{principal.employee, accomplice.employee}},
					}
				case "accounting_fraud":
					item.motive = "Filed company figures were altered before submission to the national revenue office, with two employees handling the changes."
					item.evidence = []evidencePlan{
						{typeID: "filing_copy", presentation: "Filed declaration copy", content: "The submitted revenue declaration carries the principal's revision initials.", supports: []string{principal.employee}},
						{typeID: "working_ledger", presentation: "Working ledger", content: "The retained ledger contains the original figures and a second employee's correction marks.", supports: []string{accomplice.employee}},
						{typeID: "submission_register", presentation: "Submission register", content: "A dispatch entry links both employees to the altered filing packet.", supports: []string{principal.employee, accomplice.employee}},
					}
				default:
					item.motive, item.evidence = expandedCrimeStory(caseType, principal.employee, accomplice.employee)
				}
				result = append(result, item)
			}
		}
	}
	return result
}

func expandedCrimeStory(caseType, principalID, accompliceID string) (string, []evidencePlan) {
	type story struct{ motive, first, second, third string }
	stories := map[string]story{
		"insurance_fraud":             {"A staged loss was filed for an insurance payment, with two employees preparing the claim and supporting records.", "The claim names the principal as the person who certified the loss.", "A retained inventory copy bears the accomplice's alterations.", "The dates and values on the claim conflict with the company's dispatch register."},
		"investment_fraud":            {"Two employees circulated a false investment offer and redirected the subscriptions.", "The principal signed the private investment circular.", "The accomplice logged deposits against a project absent from company minutes.", "A reconciliation links the subscriptions to a secondary account."},
		"money_laundering":            {"Company transactions were routed through ordinary accounts to disguise the source of funds.", "The principal approved an unexplained transfer sequence.", "The accomplice filed invoices for services absent from company records.", "Matching references connect the invoices and transfers."},
		"financial_identity_theft":    {"An employee used another person's financial identity to authorize unauthorized transactions.", "A payment instruction carries the principal's access code.", "The accomplice retained a duplicate form with altered identity details.", "The settlement register shows the resulting funds moving to a secondary account."},
		"procurement_fraud":           {"A purchasing contract was steered through concealed bid coordination.", "The principal approved the selected supplier outside the usual review.", "The accomplice's comparison sheet contains copied bid figures.", "Meeting notes connect both employees to the changed award record."},
		"forged_will":                 {"A forged estate document was used to redirect a company-held inheritance payment.", "The principal delivered the disputed instruction to the company registrar.", "The accomplice altered the filing copy after the original was logged.", "The dispatch book places both employees with the document before the transfer."},
		"cargo_theft":                 {"A shipment was diverted before delivery, with an employee changing its route and a colleague clearing the manifest.", "The principal authorized the altered route.", "The accomplice signed the amended cargo manifest.", "The dock ledger records the shipment as delivered despite its missing freight."},
		"art_forgery":                 {"A reproduction was substituted for an insured original while two employees handled the inventory paperwork.", "The principal signed the transfer certificate for the disputed work.", "The accomplice's inventory copy records a replacement after closing.", "The valuation sheet and movement log give incompatible descriptions of the same item."},
		"antiquities_smuggling":       {"A protected artifact was moved under a false commercial description.", "The principal approved the export declaration.", "The accomplice altered the crate inventory before dispatch.", "The retained packing list conflicts with the transport register."},
		"correspondence_blackmail":    {"A private letter was used to demand money from a company officer, with a colleague arranging collection.", "The principal's draft names the threatened disclosure.", "The accomplice logged a sealed collection envelope.", "A payment reference matches the amount demanded in the correspondence."},
		"industrial_sabotage":         {"Equipment was deliberately disabled and maintenance records were altered to hide the cause.", "The principal signed out the restricted maintenance key.", "The accomplice changed the inspection entry after the breakdown.", "The shift ledger places both employees in the plant during the unexplained shutdown."},
		"arson":                       {"A company building was deliberately set alight after records were removed and an insurance claim prepared.", "The principal filed an early loss estimate.", "The accomplice signed the after-hours access register.", "The inventory and access records show the same restricted room was cleared before the fire."},
		"staged_disappearance":        {"An employee's disappearance was staged to evade a debt and move company funds.", "The principal prepared a false departure notice.", "The accomplice recorded a company advance under a closed account.", "The travel and payroll ledgers conflict with the reported departure date."},
		"confidential_document_theft": {"Restricted company papers were copied and removed for an unauthorized sale.", "The principal requested access to the sealed archive.", "The accomplice's checkout slip lists the missing folder under another title.", "A duplicate index records both employees handling the file."},
		"police_corruption":           {"A company employee paid for confidential police information and a colleague concealed the expense.", "The principal authorized an unlisted consulting payment.", "The accomplice filed the receipt under a false account heading.", "The payment ledger and access log link the two employees to the same inquiry date."},
	}
	selected, ok := stories[caseType]
	if !ok {
		return "", nil
	}
	return selected.motive, []evidencePlan{
		{typeID: caseType + "_primary_document", presentation: "Primary document", content: selected.first, supports: []string{principalID}},
		{typeID: caseType + "_supporting_record", presentation: "Supporting record", content: selected.second, supports: []string{accompliceID}},
		{typeID: caseType + "_corroborating_entry", presentation: "Corroborating entry", content: selected.third, supports: []string{principalID, accompliceID}},
	}
}

// serialKillerCandidates requires the existing world to support the entire
// first victim clue chain: a working killer, a distinct victim, and someone
// close to that victim whom the player can write to.
func serialKillerCandidates(base world.World) []candidate {
	if !world.SupportsSocietyVersion(base.GeneratorVersion) {
		return nil
	}
	type serialWitness struct {
		npcID        string
		relationship RelationshipReference
	}
	locationsByCity := make(map[string][]string)
	for _, entity := range base.Entities {
		if entity.EntityType == "location" {
			city := cityForLocation(base, entity.ID)
			locationsByCity[city] = append(locationsByCity[city], entity.ID)
		}
	}
	for city := range locationsByCity {
		sort.Strings(locationsByCity[city])
	}
	peopleByID := make(map[string]world.NPC, len(base.NPCs))
	peopleByCity := make(map[string][]world.NPC)
	for _, person := range base.NPCs {
		peopleByID[person.ID] = person
		city := serialNPCcity(person)
		peopleByCity[city] = append(peopleByCity[city], person)
	}
	for city := range peopleByCity {
		sort.Slice(peopleByCity[city], func(i, j int) bool { return peopleByCity[city][i].ID < peopleByCity[city][j].ID })
	}
	jobsByNPC := make(map[string][]world.EmploymentRelationship)
	coWorkersByCompany := make(map[string][]world.EmploymentRelationship)
	for _, relation := range base.EmploymentRelationships {
		jobsByNPC[relation.EmployeeNPCID] = append(jobsByNPC[relation.EmployeeNPCID], relation)
		coWorkersByCompany[relation.EmployerCompanyID] = append(coWorkersByCompany[relation.EmployerCompanyID], relation)
	}
	for npcID := range jobsByNPC {
		sort.Slice(jobsByNPC[npcID], func(i, j int) bool { return jobsByNPC[npcID][i].ID < jobsByNPC[npcID][j].ID })
	}
	witnessesByVictim := make(map[string][]serialWitness, len(base.NPCs))
	for _, relation := range base.PersonalRelationships {
		if relation.RelationshipType != "marriage" {
			continue
		}
		if relation.PersonANPCID != "" && peopleByID[relation.PersonBNPCID].ID != "" {
			witnessesByVictim[relation.PersonANPCID] = append(witnessesByVictim[relation.PersonANPCID], serialWitness{npcID: relation.PersonBNPCID, relationship: RelationshipReference{ID: relation.ID, Type: "personal", Role: "victim_partner"}})
		}
		if relation.PersonBNPCID != "" && peopleByID[relation.PersonANPCID].ID != "" {
			witnessesByVictim[relation.PersonBNPCID] = append(witnessesByVictim[relation.PersonBNPCID], serialWitness{npcID: relation.PersonANPCID, relationship: RelationshipReference{ID: relation.ID, Type: "personal", Role: "victim_partner"}})
		}
	}
	for victimID, jobs := range jobsByNPC {
		for _, job := range jobs {
			for _, coworker := range coWorkersByCompany[job.EmployerCompanyID] {
				if coworker.EmployeeNPCID != victimID {
					witnessesByVictim[victimID] = append(witnessesByVictim[victimID], serialWitness{npcID: coworker.EmployeeNPCID, relationship: RelationshipReference{ID: job.ID, Type: "employment", Role: "victim_employment"}})
					break
				}
			}
		}
	}
	companiesByID := make(map[string]world.Company, len(base.Companies))
	for _, company := range base.Companies {
		companiesByID[company.ID] = company
	}
	var result []candidate
	witnessFor := func(victimID, killerID string) (RelationshipReference, bool) {
		for _, witness := range witnessesByVictim[victimID] {
			if witness.npcID != killerID && witness.npcID != victimID && peopleByID[witness.npcID].ID != "" {
				return witness.relationship, true
			}
		}
		return RelationshipReference{}, false
	}
	for city, cityPeople := range peopleByCity {
		locations := locationsByCity[city]
		if len(locations) == 0 {
			continue
		}
		for _, killer := range cityPeople {
			killerID := killer.ID
			if len(jobsByNPC[killerID]) == 0 || killer.Status != "" && killer.Status != "alive" {
				continue
			}
			killerJob := jobsByNPC[killerID][0]
			company := companiesByID[killerJob.EmployerCompanyID]
			if company.ID == "" {
				continue
			}
			var victimID, futureVictimID string
			var witnessRelation RelationshipReference
			for _, victim := range cityPeople {
				if victim.ID == killerID || victim.Status != "" && victim.Status != "alive" || !hasJobAtCompany(jobsByNPC[victim.ID], killerJob.EmployerCompanyID) {
					continue
				}
				if witness, ok := witnessFor(victim.ID, killerID); ok {
					victimID, witnessRelation = victim.ID, witness
					break
				}
			}
			if victimID == "" {
				continue
			}
			for _, possible := range cityPeople {
				if possible.ID == killerID || possible.ID == victimID || possible.Status != "" && possible.Status != "alive" {
					continue
				}
				if len(jobsByNPC[possible.ID]) == 0 {
					continue
				}
				if _, ok := witnessFor(possible.ID, killerID); ok {
					futureVictimID = possible.ID
					break
				}
			}
			if futureVictimID == "" {
				continue
			}
			result = append(result, candidate{
				archetypeID: "serial_killer", motive: "An unknown correspondent terrorized several people before a sequence of murders. The messages point toward the killer's workplace and appearance; people close to the victims can help identify the sender.",
				killerID: killerID, victimID: victimID, locationID: locations[0], companyID: company.ID,
				futureVictimIDs: []string{futureVictimID},
				relationships:   []RelationshipReference{{ID: killerJob.ID, Type: "employment", Role: "killer_employment"}, witnessRelation},
				evidence: []evidencePlan{
					{typeID: "letter_fragment", presentation: "Anonymous threat correspondence", content: "Several unsigned messages threaten the recipient. Their repeated details should be compared with testimony from people who knew the victim.", supports: []string{killerID}},
					{typeID: "witness_statement", presentation: "Statement from someone close to the victim", content: "The victim shared the threats with someone they trusted. That person may recall what the messages said about the sender.", supports: []string{killerID}},
					{typeID: "office_log", presentation: "Staff access ledger", content: "Company staff records can be compared with the workplace and appearance details described in the correspondence.", supports: []string{killerID}},
				},
			})
		}
	}
	return result
}

func serialNPCcity(npc world.NPC) string {
	if index := strings.Index(npc.RegionID, "/district/"); index >= 0 {
		return npc.RegionID[:index]
	}
	return npc.RegionID
}

func hasJobAtCompany(jobs []world.EmploymentRelationship, companyID string) bool {
	for _, job := range jobs {
		if job.EmployerCompanyID == companyID {
			return true
		}
	}
	return false
}

func serialVictimWitness(base world.World, victimID, killerID string, jobsByNPC map[string][]world.EmploymentRelationship) (string, RelationshipReference) {
	for _, relation := range base.PersonalRelationships {
		if relation.RelationshipType != "marriage" {
			continue
		}
		other := ""
		if relation.PersonANPCID == victimID {
			other = relation.PersonBNPCID
		} else if relation.PersonBNPCID == victimID {
			other = relation.PersonANPCID
		}
		if other != "" && other != killerID && npcByID(base.NPCs, other).ID != "" {
			return other, RelationshipReference{ID: relation.ID, Type: "personal", Role: "victim_partner"}
		}
	}
	for _, victimJob := range jobsByNPC[victimID] {
		for _, coworkerJob := range base.EmploymentRelationships {
			if coworkerJob.EmployerCompanyID == victimJob.EmployerCompanyID && coworkerJob.EmployeeNPCID != victimID && coworkerJob.EmployeeNPCID != killerID {
				return coworkerJob.EmployeeNPCID, RelationshipReference{ID: victimJob.ID, Type: "employment", Role: "victim_employment"}
			}
		}
	}
	return "", RelationshipReference{}
}

func companyByID(companies []world.Company, id string) world.Company {
	for _, company := range companies {
		if company.ID == id {
			return company
		}
	}
	return world.Company{}
}

func validateFinancialDefinition(definition Definition, base world.World) error {
	if definition.GeneratorVersion < previousFinancialGeneratorVersion || definition.CaseType != definition.ArchetypeID || !financialCrimeType(definition.CaseType) || definition.VictimID != "" || definition.KillerID == "" || definition.CompanyID == "" || definition.CrimeAt.IsZero() || len(definition.Participants) != 2 || len(definition.Relationships) != 2 || len(definition.Evidence) != 3 {
		return ErrInvalidDefinition
	}
	if !containsNPC(base.NPCs, definition.KillerID) || !containsNPC(base.NPCs, definition.Participants[1].NPCID) || definition.Participants[0] != (CaseParticipant{NPCID: definition.KillerID, Role: "principal"}) || definition.Participants[1].Role != "accomplice" || definition.Participants[0].NPCID == definition.Participants[1].NPCID || definition.LocationID != "" {
		return ErrInvalidDefinition
	}
	validPrincipal, validAccomplice := false, false
	for _, relation := range base.EmploymentRelationships {
		for _, reference := range definition.Relationships {
			if reference.ID != relation.ID || reference.Type != "employment" || relation.EmployerCompanyID != definition.CompanyID {
				continue
			}
			validPrincipal = validPrincipal || reference.Role == "principal_employment" && relation.EmployeeNPCID == definition.KillerID
			validAccomplice = validAccomplice || reference.Role == "accomplice_employment" && relation.EmployeeNPCID == definition.Participants[1].NPCID
		}
	}
	if !validPrincipal || !validAccomplice || !companyExists(base.Companies, definition.CompanyID) || cityForCompany(base, definition.CompanyID) != definition.CityID {
		return ErrInvalidDefinition
	}
	required := map[string]bool{}
	supportedParticipants := map[string]bool{}
	for _, evidence := range definition.Evidence {
		if len(evidence.SupportsNPCIDs) == 0 || evidence.ID == "" || evidence.Presentation == "" || evidence.Content == "" {
			return ErrUnsolvableCase
		}
		required[evidence.Type] = true
		for _, id := range evidence.SupportsNPCIDs {
			if id != definition.KillerID && id != definition.Participants[1].NPCID {
				return ErrInvalidDefinition
			}
			supportedParticipants[id] = true
		}
	}
	expectedTypes := map[string][]string{
		"embezzlement":        {"settlement_ledger", "duplicate_receipt", "account_reconciliation"},
		"financial_extortion": {"payment_demand", "collection_record", "account_reconciliation"},
		"accounting_fraud":    {"filing_copy", "working_ledger", "submission_register"},
	}
	if contains(expandedCrimeTypes, definition.CaseType) {
		prefix := definition.CaseType
		expectedTypes[definition.CaseType] = []string{prefix + "_primary_document", prefix + "_supporting_record", prefix + "_corroborating_entry"}
	}
	for _, evidenceType := range expectedTypes[definition.CaseType] {
		if !required[evidenceType] {
			return ErrUnsolvableCase
		}
	}
	if !supportedParticipants[definition.KillerID] || !supportedParticipants[definition.Participants[1].NPCID] {
		return ErrUnsolvableCase
	}
	if len(definition.Records) != 2 {
		return ErrUnsolvableCase
	}
	var sharedReference string
	var sharedAmount int
	seenNPCs := map[string]bool{}
	for _, record := range definition.Records {
		if record.CaseID != definition.ID || record.Source != "bank" || record.Amount <= 0 || record.Reference == "" || record.RelatedNPCID == "" || record.InternalRelation == "" || (record.NPCID != definition.KillerID && record.NPCID != definition.Participants[1].NPCID) {
			return ErrInvalidDefinition
		}
		if record.RelatedNPCID == record.NPCID || seenNPCs[record.NPCID] {
			return ErrInvalidDefinition
		}
		seenNPCs[record.NPCID] = true
		if sharedReference == "" {
			sharedReference, sharedAmount = record.Reference, record.Amount
		} else if sharedReference != record.Reference || sharedAmount != record.Amount {
			return ErrUnsolvableCase
		}
	}
	if !seenNPCs[definition.KillerID] || !seenNPCs[definition.Participants[1].NPCID] {
		return ErrUnsolvableCase
	}
	return nil
}

func cityLocalDay(locationID string, base world.World, now time.Time) time.Time {
	return cityLocalDayInCity(cityForLocation(base, locationID), now)
}

func cityLocalDayInCity(cityID string, now time.Time) time.Time {
	city, ok := world.CityByID(cityID)
	if !ok {
		return now.UTC().Truncate(24 * time.Hour)
	}
	zone, err := time.LoadLocation(city.TimeZone)
	if err != nil {
		return now.UTC().Truncate(24 * time.Hour)
	}
	localNow := now.In(zone)
	return time.Date(localNow.Year(), localNow.Month(), localNow.Day(), 0, 0, 0, 0, zone).UTC()
}

func cityForCompany(base world.World, companyID string) string {
	for _, company := range base.Companies {
		if company.ID == companyID {
			parts := strings.Split(company.RegionID, "/")
			if len(parts) >= 2 {
				return strings.Join(parts[:2], "/")
			}
		}
	}
	return "fr/lyon"
}

func caseTimeOffset(seed, archetypeID string, version int) time.Duration {
	startMinute := 18 * 60
	if version >= contextualGeneratorVersion {
		switch archetypeID {
		case "workplace_conflict":
			startMinute = 19 * 60
		case "financial_extortion":
			startMinute = 20 * 60
		}
	}
	return time.Duration(startMinute+caseIndex(seed+"/crime-time", version, 4*60)) * time.Minute
}

func availableArchetypesForVersion(base world.World, version int) []archetypeGenerator {
	available := make([]archetypeGenerator, 0, len(archetypeGenerators))
	for _, archetype := range archetypeGenerators {
		if len(archetypeCandidates(archetype, base, version)) > 0 {
			available = append(available, archetype)
		}
	}
	return available
}

func archetypeCandidates(archetype archetypeGenerator, base world.World, version int) []candidate {
	if archetype.id == "serial_killer" && version < previousSerialGeneratorVersion {
		return nil
	}
	for _, motive := range newHomicideMotives {
		if archetype.id == motive.id && version < GeneratorVersion {
			return nil
		}
	}
	if version >= contextualGeneratorVersion && archetype.id == "jealousy" {
		return jealousyCandidatesV6(base)
	}
	if version >= connectedGeneratorVersion && archetype.id == "jealousy" {
		return jealousyCandidatesV4(base)
	}
	return archetype.candidates(base)
}

func cityForLocation(base world.World, locationID string) string {
	for _, entity := range base.Entities {
		if entity.ID != locationID {
			continue
		}
		if index := strings.Index(entity.RegionID, "/district/"); index > 0 {
			return entity.RegionID[:index]
		}
	}
	return "fr/lyon"
}

func buildCaseEmails(seed string, worldVersion int, caseID string, selected candidate, base world.World, caseDate time.Time) []mail.Message {
	victim := npcByID(base.NPCs, selected.victimID)
	firstSubject, firstBody := "A private warning", "Leave the meeting notes as they are. If you take them to the registry, you will regret it."
	lastSubject, lastBody := "One last message", "This is your final chance to keep quiet. Do not make me come looking for you."
	switch selected.archetypeID {
	case "jealousy":
		firstSubject = "Do not ask about the meeting"
		firstBody = "You were told to leave that evening alone. Stop asking who was there, and do not show the letter to anyone."
		lastSubject = "You have been warned"
		lastBody = "Keep the letter to yourself. If this becomes a public matter, you will wish you had listened."
	case "financial_extortion":
		firstSubject = "The remittance"
		firstBody = "The next payment is due before Friday. I still have the duplicate settlement statement. Do not take the ledger to the registry."
		lastSubject = "Final notice"
		lastBody = "The payment is late. Send it as agreed, or the duplicate statement will reach people you would rather not involve."
	}
	messageDay := 3 + caseIndex(seed+"/correspondence/day", 1, 8)
	messageMinute := 7*60 + caseIndex(seed+"/correspondence/minute", 1, 13*60)
	lastSentAt := caseDate.AddDate(0, 0, messageDay-12).Add(time.Duration(messageMinute) * time.Minute)
	firstSentAt := lastSentAt.Add(-48 * time.Hour)
	return []mail.Message{
		mail.CaseMessage(seed, worldVersion, caseID, victim.ID, victim.ID, "Unknown sender", victim.Name, firstSubject, firstBody, "anonymous_warning", firstSentAt),
		mail.CaseMessage(seed, worldVersion, caseID, victim.ID, victim.ID, "Unknown sender", victim.Name, lastSubject, lastBody, "anonymous_followup", lastSentAt),
	}
}

func buildConnectedCaseRecords(seed string, worldVersion int, caseID string, selected candidate, base world.World, crimeAt time.Time, caseVersion int) []records.Record {
	keyVersion := connectedGeneratorVersion
	if caseVersion >= previousGeneratorVersion {
		keyVersion = caseVersion
	}
	key := fmt.Sprintf("noir-case-records/v%d/%s/%d/%s", keyVersion, seed, worldVersion, caseID)
	scene := regionName(base, locationRegion(base, selected.locationID))
	killerOrigin := regionName(base, npcByID(base.NPCs, selected.killerID).RegionID)
	victimOrigin := regionName(base, npcByID(base.NPCs, selected.victimID).RegionID)
	if killerOrigin == scene {
		killerOrigin = alternateDistrict(base, scene)
	}
	if victimOrigin == scene {
		victimOrigin = alternateDistrict(base, scene)
	}
	result := []records.Record{
		caseGateRecord(key, caseID, "scene-victim", selected.victimID, selected.killerID, crimeAt.Add(-38*time.Minute), victimOrigin, scene, "A passenger entry was recorded at the local relay."),
		caseGateRecord(key, caseID, "scene-killer", selected.killerID, selected.victimID, crimeAt.Add(-19*time.Minute), killerOrigin, scene, "A passenger entry was recorded at the local relay."),
	}
	if caseVersion >= previousGeneratorVersion && keyIndex(key, "killer-scene-exit", 2) == 0 {
		result = append(result, caseGateRecord(key, caseID, "scene-killer-exit", selected.killerID, selected.victimID, crimeAt.Add(14*time.Minute), scene, killerOrigin, "A passenger exit was recorded at the local relay."))
	}
	if selected.archetypeID == "financial_extortion" {
		amount := 45 + keyIndex(key, "payment-amount", 46)
		reference := fmt.Sprintf("unregistered account %04d", keyIndex(key, "account-reference", 10000))
		paidAt := crimeAt.AddDate(0, 0, -2).Add(-70 * time.Minute)
		result = append(result,
			records.Record{ID: caseRecordID(key, "payment-out", selected.victimID), NPCID: selected.victimID, Source: "bank", OccurredAt: paidAt, Operator: "Union Mercantile Settlement Office", Description: "An anonymous remittance was sent and cleared.", Reference: reference, Amount: amount, Direction: "outgoing", CaseID: caseID, InternalRelation: "extortion_payment", RelatedNPCID: selected.killerID},
			records.Record{ID: caseRecordID(key, "payment-in", selected.killerID), NPCID: selected.killerID, Source: "bank", OccurredAt: paidAt, Operator: "Union Mercantile Settlement Office", Description: "An anonymous remittance was received and cleared.", Reference: reference, Amount: amount, Direction: "incoming", CaseID: caseID, InternalRelation: "extortion_payment", RelatedNPCID: selected.victimID},
		)
		result = append(result, buildTransferDecoys(key, caseID, base, selected, crimeAt)...)
	}
	if selected.archetypeID == "jealousy" && selected.paramourID != "" {
		paramourOrigin := regionName(base, npcByID(base.NPCs, selected.paramourID).RegionID)
		if paramourOrigin == scene {
			paramourOrigin = alternateDistrict(base, scene)
		}
		rendezvous := crimeAt.AddDate(0, 0, -3).Add(-42 * time.Minute)
		result = append(result,
			caseGateRecord(key, caseID, "affair-victim", selected.victimID, selected.paramourID, rendezvous, victimOrigin, scene, "A passenger entry was recorded at the local relay."),
			caseGateRecord(key, caseID, "affair-paramour", selected.paramourID, selected.victimID, rendezvous, paramourOrigin, scene, "A passenger entry was recorded at the local relay."),
		)
	}
	sort.Slice(result, func(i, j int) bool {
		if result[i].OccurredAt.Equal(result[j].OccurredAt) {
			return result[i].ID < result[j].ID
		}
		return result[i].OccurredAt.Before(result[j].OccurredAt)
	})
	return result
}

func caseGateRecord(key, caseID, role, npcID, relatedID string, at time.Time, origin, destination, description string) records.Record {
	return records.Record{ID: caseRecordID(key, role, npcID), NPCID: npcID, Source: "gate", OccurredAt: at.UTC(), Operator: "Meridian Global · GATE Network", Description: description, Origin: origin, Destination: destination, CaseID: caseID, InternalRelation: role, RelatedNPCID: relatedID}
}

func buildTransferDecoys(key, caseID string, base world.World, selected candidate, crimeAt time.Time) []records.Record {
	caseCityID := cityForLocation(base, selected.locationID)
	people := make([]world.NPC, 0, len(base.NPCs))
	for _, person := range base.NPCs {
		personCityID := person.RegionID
		if districtIndex := strings.Index(personCityID, "/district/"); districtIndex >= 0 {
			personCityID = personCityID[:districtIndex]
		}
		if person.ID != selected.killerID && person.ID != selected.victimID && personCityID == caseCityID {
			people = append(people, person)
		}
	}
	sort.Slice(people, func(i, j int) bool { return people[i].ID < people[j].ID })
	if len(people) < 2 {
		return nil
	}
	limit := min(4, len(people))
	result := make([]records.Record, 0, limit)
	for index := 0; index+1 < limit; index += 2 {
		from, to := people[index], people[index+1]
		amount := 20 + keyIndex(key, "decoy-amount/"+from.ID, 120)
		reference := fmt.Sprintf("unregistered account %04d", keyIndex(key, "decoy-account/"+from.ID, 10000))
		at := crimeAt.AddDate(0, 0, -4-index/2).Add(time.Duration(9*60+keyIndex(key, "decoy-time/"+from.ID, 8*60)) * time.Minute)
		result = append(result,
			records.Record{ID: caseRecordID(key, "decoy-out/"+from.ID, to.ID), NPCID: from.ID, Source: "bank", OccurredAt: at, Operator: "Union Mercantile Settlement Office", Description: "An anonymous remittance was sent and cleared.", Reference: reference, Amount: amount, Direction: "outgoing", CaseID: caseID, InternalRelation: "unrelated_transfer", RelatedNPCID: to.ID},
			records.Record{ID: caseRecordID(key, "decoy-in/"+to.ID, from.ID), NPCID: to.ID, Source: "bank", OccurredAt: at, Operator: "Union Mercantile Settlement Office", Description: "An anonymous remittance was received and cleared.", Reference: reference, Amount: amount, Direction: "incoming", CaseID: caseID, InternalRelation: "unrelated_transfer", RelatedNPCID: from.ID},
		)
	}
	return result
}

func buildConnectedCaseEmails(seed string, worldVersion int, caseID string, selected candidate, base world.World, crimeAt time.Time) []mail.Message {
	killer, victim := npcByID(base.NPCs, selected.killerID), npcByID(base.NPCs, selected.victimID)
	messageAt := crimeAt.Add(-26 * time.Hour)
	subject, body := "We need to talk", "I know you were at the GATE after the office closed. Please tell me who you met before this becomes something we cannot explain."
	if selected.archetypeID == "financial_extortion" {
		subject, body = "The last remittance", "The payment cleared. Use the same amount next time and keep the account reference out of your ledger. If anyone asks, you have never met me."
	} else if selected.archetypeID == "workplace_conflict" {
		subject, body = "About the office register", "The corrected entry is still there. I thought we agreed to leave the matter alone until the audit was finished."
	}
	mailbox := mail.CaseMessage(seed, worldVersion, caseID, victim.ID, killer.ID, killer.Name, victim.Name, subject, body, "case_correspondence", messageAt)
	if caseIndex(seed+"/anonymous-case-mail", 1, 5) == 0 {
		mailbox.InternalRelation = "anonymous_warning"
	}
	result := []mail.Message{mailbox}
	if selected.archetypeID == "jealousy" && selected.paramourID != "" {
		paramour := npcByID(base.NPCs, selected.paramourID)
		message := mail.CaseMessage(seed, worldVersion, caseID, victim.ID, paramour.ID, victim.Name, paramour.Name, "About last Tuesday", "It was only dinner, but my husband was asking where I had been. I cannot keep making excuses for these meetings.", "jealousy_exchange", crimeAt.AddDate(0, 0, -4))
		result = append(result, message)
	}
	if correspondent := killerCorrespondent(base, selected); correspondent.ID != "" {
		confession := mail.CaseMessage(seed, worldVersion, caseID, killer.ID, correspondent.ID, killer.Name, correspondent.Name, "I made a mistake", "I did something I cannot take back. If anyone asks, tell them I left before the last GATE call. Please do not ask me what happened.", "killer_remorse", crimeAt.Add(2*time.Hour))
		result = append(result, confession)
	}
	return result
}

func buildSerialKillerEmails(seed string, worldVersion int, caseID string, selected candidate, base world.World, crimeAt time.Time) []mail.Message {
	victim, killer := npcByID(base.NPCs, selected.victimID), npcByID(base.NPCs, selected.killerID)
	company := companyByID(base.Companies, selected.companyID)
	appearance := killer.Appearance
	if appearance == "" {
		appearance = "a distinctive figure in a dark coat"
	}
	letters := []struct{ subject, body string }{
		{"You have been noticed", "I know when you leave. I know the route you take home. Do not ask who sent this."},
		{"You should have listened", "You saw me near the office. Remember " + appearance + ". You will see me again before this is finished."},
		{"The last warning", "The people at work know how to keep quiet. You should learn from them. I am closer than you think."},
		{"Still watching", "You told someone about my messages. That was a mistake. " + appearance + ". I am still there after the bell."},
		{"No one can keep you safe", "You can change your route, but not what you know about where I work. Stop talking, or the next letter will be the last."},
	}
	// The precise employer appears in only one deterministic letter for the case.
	workplaceClueIndex := keyIndex(seed, "serial-workplace-letter", len(letters))
	letters[workplaceClueIndex].body += " I work at " + company.Name + "."
	messages := make([]mail.Message, 0, len(letters))
	for index, letter := range letters {
		message := mail.CaseMessage(seed, worldVersion, caseID, victim.ID, killer.ID, "Unknown sender", victim.Name,
			letter.subject, letter.body, "anonymous_warning", crimeAt.AddDate(0, 0, index-len(letters)))
		message.ID += fmt.Sprintf("-serial-%02d", index+1)
		messages = append(messages, message)
	}
	return messages
}

func killerCorrespondent(base world.World, selected candidate) world.NPC {
	var candidates []world.NPC
	for _, relation := range base.PersonalRelationships {
		if relation.PersonANPCID == selected.killerID && relation.PersonBNPCID != selected.victimID {
			person := npcByID(base.NPCs, relation.PersonBNPCID)
			if person.ID != "" {
				candidates = append(candidates, person)
			}
		}
		if relation.PersonBNPCID == selected.killerID && relation.PersonANPCID != selected.victimID {
			person := npcByID(base.NPCs, relation.PersonANPCID)
			if person.ID != "" {
				candidates = append(candidates, person)
			}
		}
	}
	if len(candidates) > 0 {
		sort.Slice(candidates, func(i, j int) bool { return candidates[i].ID < candidates[j].ID })
		return candidates[0]
	}
	companyIDs := make(map[string]bool)
	for _, relation := range base.EmploymentRelationships {
		if relation.EmployeeNPCID == selected.killerID {
			companyIDs[relation.EmployerCompanyID] = true
		}
	}
	for _, relation := range base.EmploymentRelationships {
		if companyIDs[relation.EmployerCompanyID] && relation.EmployeeNPCID != selected.killerID && relation.EmployeeNPCID != selected.victimID && relation.EmployeeNPCID != selected.paramourID {
			person := npcByID(base.NPCs, relation.EmployeeNPCID)
			if person.ID != "" {
				candidates = append(candidates, person)
			}
		}
	}
	sort.Slice(candidates, func(i, j int) bool { return candidates[i].ID < candidates[j].ID })
	if len(candidates) == 0 {
		return world.NPC{}
	}
	return candidates[0]
}

func locationRegion(base world.World, locationID string) string {
	for _, entity := range base.Entities {
		if entity.ID == locationID {
			return entity.RegionID
		}
	}
	return ""
}

func alternateDistrict(base world.World, current string) string {
	var names []string
	for _, region := range base.Regions {
		if region.Name != current && strings.Contains(region.ID, "/district/") {
			names = append(names, region.Name)
		}
	}
	sort.Strings(names)
	if len(names) > 0 {
		return names[0]
	}
	return "Central District"
}

func caseRecordID(key, role, npcID string) string {
	digest := sha256.Sum256([]byte(key + "/" + role + "/" + npcID))
	return "record:" + hex.EncodeToString(digest[:12])
}

func keyIndex(key, field string, limit int) int {
	digest := sha256.Sum256([]byte(key + "/" + field))
	return int(binary.BigEndian.Uint64(digest[:8]) % uint64(limit))
}

// Validate checks the case archetype, world references, and reachable clues.
func Validate(definition Definition, base world.World) error {
	if !world.SupportsSocietyVersion(base.GeneratorVersion) {
		return ErrInvalidDefinition
	}
	if strings.TrimSpace(definition.Seed) == "" || !supportedCaseGeneratorVersion(definition.GeneratorVersion) || definition.ID != makeCaseID(definition.Seed, definition.GeneratorVersion) || definition.ArchetypeID == "" || definition.Motive == "" {
		return ErrInvalidDefinition
	}
	if definition.CaseType != "homicide" {
		return validateFinancialDefinition(definition, base)
	}
	if definition.GeneratorVersion >= connectedGeneratorVersion && definition.CrimeAt.IsZero() {
		return ErrInvalidDefinition
	}
	if definition.KillerID == definition.VictimID {
		return ErrInvalidDefinition
	}
	if definition.ArchetypeID == "workplace_conflict" || definition.ArchetypeID == "financial_extortion" {
		if strings.TrimSpace(definition.CompanyID) == "" {
			return ErrInvalidDefinition
		}
	}
	entities := make(map[string]string, len(base.Entities))
	for _, entity := range base.Entities {
		entities[entity.ID] = entity.EntityType
	}
	if entities[definition.KillerID] != "npc" || entities[definition.VictimID] != "npc" || entities[definition.LocationID] != "location" {
		return ErrInvalidDefinition
	}
	if !containsNPC(base.NPCs, definition.KillerID) || !containsNPC(base.NPCs, definition.VictimID) {
		return ErrInvalidDefinition
	}
	links := make(map[string]RelationshipReference, len(definition.Relationships))
	for _, relation := range definition.Relationships {
		if relation.ID == "" || relation.Type == "" || relation.Role == "" || !worldRelationshipExists(base, relation.Type, relation.ID) {
			return ErrInvalidDefinition
		}
		key := relation.Type + ":" + relation.ID
		if _, exists := links[key]; exists {
			return ErrInvalidDefinition
		}
		links[key] = relation
	}
	if len(links) == 0 {
		return ErrInvalidDefinition
	}
	archetype, exists := archetypeByID(definition.ArchetypeID)
	validArchetype := exists && archetype.validate(definition, base)
	if definition.GeneratorVersion >= connectedGeneratorVersion {
		switch definition.ArchetypeID {
		case "jealousy":
			validArchetype = validateJealousyV4(definition, base)
		case "financial_extortion":
			validArchetype = validateExtortionV4(definition, base)
		}
	}
	if !validateContextualScene(definition, base) || !validateContextualTiming(definition) {
		return ErrInvalidDefinition
	}
	if !validArchetype {
		return ErrInvalidDefinition
	}
	if !hasReachableEvidence(definition.Evidence, definition, links, entities) || !hasEvidenceTypes(definition.Evidence, archetype.evidenceTypes...) || !validateCaseRecords(definition, base) || definition.GeneratorVersion >= connectedGeneratorVersion && !validateConnectedCaseEmails(definition, base) {
		return ErrUnsolvableCase
	}
	return nil
}

func supportedCaseGeneratorVersion(version int) bool {
	return version == legacyGeneratorVersion || version == connectedGeneratorVersion || version == previousGeneratorVersion || version == contextualGeneratorVersion || version == previousFinancialGeneratorVersion || version == previousSerialGeneratorVersion || version == GeneratorVersion
}

func availableArchetypes(base world.World) []archetypeGenerator {
	var available []archetypeGenerator
	for _, archetype := range archetypeGenerators {
		if len(archetype.candidates(base)) > 0 {
			available = append(available, archetype)
		}
	}
	return available
}

func workplaceCandidates(base world.World) []candidate {
	if !world.SupportsSocietyVersion(base.GeneratorVersion) {
		return nil
	}
	entities := make(map[string]string, len(base.Entities))
	for _, entity := range base.Entities {
		entities[entity.ID] = entity.EntityType
	}
	companyLocations := make(map[string]string)
	for _, relation := range base.CompanyLocations {
		if relation.LocationRole == "headquarters" && entities[relation.LocationID] == "location" {
			companyLocations[relation.CompanyID] = relation.LocationID
		}
	}
	byCompany := make(map[string][]employmentLink)
	for _, relation := range base.EmploymentRelationships {
		if entities[relation.EmployeeNPCID] != "npc" || entities[relation.EmployerCompanyID] != "company" {
			continue
		}
		byCompany[relation.EmployerCompanyID] = append(byCompany[relation.EmployerCompanyID], employmentLink{
			id: relation.ID, employee: relation.EmployeeNPCID,
		})
	}
	companies := make([]string, 0, len(byCompany))
	for companyID := range byCompany {
		companies = append(companies, companyID)
	}
	sort.Strings(companies)
	result := make([]candidate, 0)
	for _, companyID := range companies {
		locationID := companyLocations[companyID]
		if locationID == "" {
			continue
		}
		links := byCompany[companyID]
		sort.Slice(links, func(i, j int) bool { return links[i].employee < links[j].employee })
		for i := 0; i < len(links); i++ {
			for j := i + 1; j < len(links); j++ {
				result = append(result,
					workplaceCandidate(companyID, links[i].employee, links[j].employee, locationID, links[i].id, links[j].id),
					workplaceCandidate(companyID, links[j].employee, links[i].employee, locationID, links[j].id, links[i].id),
				)
			}
		}
	}
	return result
}

func workplaceCandidate(companyID, killerID, victimID, locationID, killerLinkID, victimLinkID string) candidate {
	return candidate{
		archetypeID: "workplace_conflict",
		companyID:   companyID,
		motive:      "A long-running dispute over status and routine office decisions.",
		killerID:    killerID,
		victimID:    victimID,
		locationID:  locationID,
		relationships: []RelationshipReference{
			{ID: killerLinkID, Type: "employment", Role: "killer_employment"},
			{ID: victimLinkID, Type: "employment", Role: "victim_employment"},
		},
		evidence: []evidencePlan{
			{typeID: "employment_record", presentation: "Personnel file", content: "A routine personnel entry places both people within the same company."},
			{typeID: "meeting_minutes", presentation: "Meeting minutes", content: "A brief note records a disagreement over responsibilities. The names are abbreviated."},
			{typeID: "office_log", presentation: "After-hours register", content: "Two familiar signatures appear on a late entry; the time has been corrected by hand."},
		},
	}
}

func extortionCandidates(base world.World) []candidate {
	workplace := workplaceCandidates(base)
	result := make([]candidate, 0, len(workplace))
	for _, possibility := range workplace {
		possibility.archetypeID = "financial_extortion"
		possibility.motive = "Financial extortion escalated after the victim threatened to expose a private arrangement."
		possibility.evidence = []evidencePlan{
			{typeID: "account_record", presentation: "Settlement statement", content: "A private remittance appears among routine account activity. The counterparty is listed only by ledger reference."},
			{typeID: "travel_register", presentation: "Passenger register", content: "A GATE passage was recorded shortly before the reported time of death."},
			{typeID: "personnel_record", presentation: "Personnel file", content: "Two staff records place the names within the same corporate office."},
		}
		result = append(result, possibility)
	}
	return result
}

func jealousyCandidates(base world.World) []candidate {
	locationsByRegion := make(map[string][]string)
	for _, entity := range base.Entities {
		if entity.EntityType == "location" {
			locationsByRegion[entity.RegionID] = append(locationsByRegion[entity.RegionID], entity.ID)
		}
	}
	var result []candidate
	for _, relation := range base.PersonalRelationships {
		if relation.RelationshipType != "marriage" {
			continue
		}
		first, second := npcByID(base.NPCs, relation.PersonANPCID), npcByID(base.NPCs, relation.PersonBNPCID)
		if first.ID == "" || second.ID == "" || first.RegionID != second.RegionID {
			continue
		}
		locations := locationsByRegion[first.RegionID]
		if len(locations) == 0 {
			continue
		}
		sort.Strings(locations)
		locationID := locations[0]
		for _, orientation := range [][2]string{{first.ID, second.ID}, {second.ID, first.ID}} {
			result = append(result, candidate{
				archetypeID: "jealousy",
				motive:      "Possessive jealousy tied to a private romantic suspicion.",
				killerID:    orientation[0], victimID: orientation[1], locationID: locationID,
				relationships: []RelationshipReference{{ID: relation.ID, Type: "personal", Role: "marriage"}},
				evidence: []evidencePlan{
					{typeID: "civil_record", presentation: "Civil registry extract", content: "The record confirms a private connection between the two names. It gives no reason for the death."},
					{typeID: "letter_fragment", presentation: "Unsent correspondence", content: "A torn sentence refers to a meeting and asks that it remain undisclosed."},
					{typeID: "witness_statement", presentation: "A neighbor's note", content: "A resident recalls seeing one spouse return unusually late, but cannot give an exact time."},
				},
			})
		}
	}
	return result
}

// jealousyCandidatesV4 requires a real spouse pair and a workplace acquaintance
// who can serve as the suspected lover in the generated paper and travel trails.
func jealousyCandidatesV4(base world.World) []candidate {
	return jealousyCandidatesForVersion(base, false)
}

func jealousyCandidatesV6(base world.World) []candidate {
	return jealousyCandidatesForVersion(base, true)
}

func jealousyCandidatesForVersion(base world.World, privateScene bool) []candidate {
	var result []candidate
	locationsByRegion := make(map[string][]string)
	companyScenes := make(map[string]bool)
	if privateScene {
		for _, relation := range base.CompanyLocations {
			if relation.LocationRole == "headquarters" {
				companyScenes[relation.LocationID] = true
			}
		}
	}
	for _, entity := range base.Entities {
		if entity.EntityType == "location" && (!privateScene || entity.LocationKind == "building" && !companyScenes[entity.ID]) {
			locationsByRegion[entity.RegionID] = append(locationsByRegion[entity.RegionID], entity.ID)
		}
	}
	for _, relation := range base.PersonalRelationships {
		if relation.RelationshipType != "marriage" {
			continue
		}
		first, second := npcByID(base.NPCs, relation.PersonANPCID), npcByID(base.NPCs, relation.PersonBNPCID)
		if first.ID == "" || second.ID == "" || first.RegionID != second.RegionID {
			continue
		}
		locations := append([]string(nil), locationsByRegion[first.RegionID]...)
		if len(locations) == 0 {
			continue
		}
		sort.Strings(locations)
		for _, spouseOrientation := range [][2]world.NPC{{first, second}, {second, first}} {
			killer, victim := spouseOrientation[0], spouseOrientation[1]
			var paramours []employmentLink
			victimEmploymentID := ""
			for _, employment := range base.EmploymentRelationships {
				if employment.EmployeeNPCID == victim.ID {
					victimEmploymentID = employment.ID
					for _, other := range base.EmploymentRelationships {
						if other.EmployerCompanyID == employment.EmployerCompanyID && other.EmployeeNPCID != killer.ID && other.EmployeeNPCID != victim.ID {
							person := npcByID(base.NPCs, other.EmployeeNPCID)
							if person.ID != "" && person.RegionID == victim.RegionID {
								paramours = append(paramours, employmentLink{id: other.ID, employee: person.ID})
							}
						}
					}
				}
			}
			sort.Slice(paramours, func(i, j int) bool { return paramours[i].employee < paramours[j].employee })
			for _, paramour := range paramours {
				result = append(result, candidate{
					archetypeID: "jealousy", motive: "A spouse's suspicion of a private affair escalated into violence.",
					killerID: killer.ID, victimID: victim.ID, locationID: locations[0], paramourID: paramour.employee,
					relationships: []RelationshipReference{{ID: relation.ID, Type: "personal", Role: "marriage"}, {ID: victimEmploymentID, Type: "employment", Role: "victim_employment"}, {ID: paramour.id, Type: "employment", Role: "paramour_employment"}},
					evidence: []evidencePlan{
						{typeID: "civil_record", presentation: "Civil registry extract", content: "The record confirms a marriage between two names. It gives no reason for the death."},
						{typeID: "letter_fragment", presentation: "Private correspondence", content: "A torn exchange refers to a repeated meeting and asks that the spouse not be told."},
						{typeID: "witness_statement", presentation: "A neighbor's note", content: "A resident recalls the victim meeting a colleague after work on several evenings."},
					},
				})
			}
		}
	}
	return result
}

func validateContextualScene(definition Definition, base world.World) bool {
	if definition.GeneratorVersion < contextualGeneratorVersion || definition.ArchetypeID != "jealousy" {
		return true
	}
	for _, entity := range base.Entities {
		if entity.ID != definition.LocationID || entity.EntityType != "location" || entity.LocationKind != "building" {
			continue
		}
		for _, relation := range base.CompanyLocations {
			if relation.LocationID == entity.ID && relation.LocationRole == "headquarters" {
				return false
			}
		}
		return true
	}
	return false
}

func validateContextualTiming(definition Definition) bool {
	if definition.GeneratorVersion < contextualGeneratorVersion {
		return true
	}
	if definition.CrimeAt.IsZero() {
		return false
	}
	zone := time.UTC
	if city, ok := world.CityByID(definition.CityID); ok {
		if loaded, err := time.LoadLocation(city.TimeZone); err == nil {
			zone = loaded
		}
	}
	localTime := definition.CrimeAt.In(zone)
	minute := localTime.Hour()*60 + localTime.Minute()
	startMinute := 18 * 60
	switch definition.ArchetypeID {
	case "workplace_conflict":
		startMinute = 19 * 60
	case "financial_extortion":
		startMinute = 20 * 60
	}
	return minute >= startMinute && minute < startMinute+4*60
}

func buildEvidence(caseID string, selected candidate) []EvidenceDefinition {
	evidence := make([]EvidenceDefinition, 0, len(selected.evidence))
	for index, plan := range selected.evidence {
		reachable := []string{selected.killerID, selected.victimID}
		supports := append([]string(nil), plan.supports...)
		locationID := selected.locationID
		if selected.caseType != "" {
			reachable = []string{selected.killerID, selected.accompliceID}
			locationID = ""
		} else if len(supports) == 0 {
			supports = []string{selected.killerID}
		}
		evidence = append(evidence, EvidenceDefinition{
			ID:                     fmt.Sprintf("evidence:%s:%02d:%s", caseID, index+1, plan.typeID),
			Type:                   plan.typeID,
			Presentation:           plan.presentation,
			Content:                plan.content,
			LocationID:             locationID,
			RelationshipIDs:        relationshipIDs(selected.relationships),
			ReachableFromEntityIDs: reachable,
			SupportsNPCIDs:         supports,
		})
	}
	return evidence
}

func relationshipIDs(relations []RelationshipReference) []string {
	ids := make([]string, 0, len(relations))
	for _, relation := range relations {
		ids = append(ids, relation.ID)
	}
	return ids
}

func npcByID(npcs []world.NPC, id string) world.NPC {
	for _, npc := range npcs {
		if npc.ID == id {
			return npc
		}
	}
	return world.NPC{}
}

func archetypeByID(id string) (archetypeGenerator, bool) {
	for _, archetype := range archetypeGenerators {
		if archetype.id == id {
			return archetype, true
		}
	}
	return archetypeGenerator{}, false
}

func worldRelationshipExists(base world.World, kind, id string) bool {
	switch kind {
	case "employment":
		for _, relation := range base.EmploymentRelationships {
			if relation.ID == id {
				return true
			}
		}
	case "personal":
		for _, relation := range base.PersonalRelationships {
			if relation.ID == id {
				return true
			}
		}
	}
	return false
}

func validateWorkplaceConflict(definition Definition, base world.World) bool {
	var killerLink, victimLink *world.EmploymentRelationship
	for index := range base.EmploymentRelationships {
		relation := &base.EmploymentRelationships[index]
		for _, ref := range definition.Relationships {
			if ref.Type == "employment" && ref.ID == relation.ID {
				if ref.Role == "killer_employment" && relation.EmployeeNPCID == definition.KillerID {
					killerLink = relation
				}
				if ref.Role == "victim_employment" && relation.EmployeeNPCID == definition.VictimID {
					victimLink = relation
				}
			}
		}
	}
	return killerLink != nil && victimLink != nil && killerLink.EmployerCompanyID == victimLink.EmployerCompanyID && killerLink.EmployerCompanyID == definition.CompanyID && hasHeadquarters(base.CompanyLocations, killerLink.EmployerCompanyID, definition.LocationID)
}

func validateSerialKiller(definition Definition, base world.World) bool {
	if definition.GeneratorVersion < previousSerialGeneratorVersion || definition.CompanyID == "" || len(definition.FutureVictimIDs) != 1 {
		return false
	}
	killer, victim := npcByID(base.NPCs, definition.KillerID), npcByID(base.NPCs, definition.VictimID)
	if killer.ID == "" || victim.ID == "" || killer.ID == victim.ID || cityForLocation(base, definition.LocationID) != cityForNPC(killer) || cityForNPC(victim) != cityForNPC(killer) {
		return false
	}
	killerEmployed := false
	for _, relation := range base.EmploymentRelationships {
		if relation.EmployeeNPCID == killer.ID && relation.EmployerCompanyID == definition.CompanyID {
			for _, reference := range definition.Relationships {
				if reference.ID == relation.ID && reference.Type == "employment" && reference.Role == "killer_employment" {
					killerEmployed = true
				}
			}
		}
	}
	witnessID, witnessLink := serialVictimWitness(base, victim.ID, killer.ID, employmentByNPC(base.EmploymentRelationships))
	if witnessID == "" {
		return false
	}
	witnessLinked := false
	for _, reference := range definition.Relationships {
		if reference.ID == witnessLink.ID && reference.Type == witnessLink.Type && reference.Role == witnessLink.Role {
			witnessLinked = true
		}
	}
	if !killerEmployed || !witnessLinked {
		return false
	}
	seen := map[string]bool{killer.ID: true, victim.ID: true}
	for _, targetID := range definition.FutureVictimIDs {
		target := npcByID(base.NPCs, targetID)
		if target.ID == "" || target.Status != "" && target.Status != "alive" || cityForNPC(target) != cityForNPC(killer) || seen[targetID] || len(employmentByNPC(base.EmploymentRelationships)[targetID]) == 0 {
			return false
		}
		seen[targetID] = true
	}
	return true
}

func employmentByNPC(relationships []world.EmploymentRelationship) map[string][]world.EmploymentRelationship {
	result := make(map[string][]world.EmploymentRelationship)
	for _, relationship := range relationships {
		result[relationship.EmployeeNPCID] = append(result[relationship.EmployeeNPCID], relationship)
	}
	return result
}

func cityForNPC(npc world.NPC) string {
	if index := strings.Index(npc.RegionID, "/district/"); index >= 0 {
		return npc.RegionID[:index]
	}
	return npc.RegionID
}

func validateJealousy(definition Definition, base world.World) bool {
	for _, relation := range base.PersonalRelationships {
		referenced := false
		for _, ref := range definition.Relationships {
			if ref.Type == "personal" && ref.Role == "marriage" && ref.ID == relation.ID {
				referenced = true
			}
		}
		isCouple := relation.PersonANPCID == definition.KillerID && relation.PersonBNPCID == definition.VictimID || relation.PersonANPCID == definition.VictimID && relation.PersonBNPCID == definition.KillerID
		if referenced && relation.RelationshipType == "marriage" && isCouple {
			return containsLocationInRegion(base, definition.LocationID, npcByID(base.NPCs, definition.KillerID).RegionID)
		}
	}
	return false
}

func validateJealousyV4(definition Definition, base world.World) bool {
	if !validateJealousy(definition, base) || definition.KillerID == "" || definition.VictimID == "" {
		return false
	}
	var victimEmployment, paramourEmployment *world.EmploymentRelationship
	for index := range base.EmploymentRelationships {
		relation := &base.EmploymentRelationships[index]
		for _, reference := range definition.Relationships {
			if reference.Type == "employment" && reference.Role == "victim_employment" && reference.ID == relation.ID && relation.EmployeeNPCID == definition.VictimID {
				victimEmployment = relation
			}
			if reference.Type == "employment" && reference.Role == "paramour_employment" && reference.ID == relation.ID && relation.EmployeeNPCID == definition.ParamourID {
				paramourEmployment = relation
			}
		}
	}
	return victimEmployment != nil && paramourEmployment != nil && victimEmployment.EmployerCompanyID == paramourEmployment.EmployerCompanyID && definition.ParamourID != definition.KillerID && definition.ParamourID != definition.VictimID && npcByID(base.NPCs, definition.ParamourID).ID != ""
}

func validateExtortion(definition Definition, base world.World) bool {
	if !validateWorkplaceConflict(definition, base) || len(definition.Records) != 2 {
		return false
	}
	hasTransfer, hasTravel := false, false
	for _, item := range definition.Records {
		switch item.InternalRelation {
		case "extortion_payment":
			hasTransfer = item.CaseID == definition.ID && item.NPCID == definition.VictimID && item.RelatedNPCID == definition.KillerID && item.Source == "bank" && item.Amount > 0
		case "precrime_route":
			hasTravel = item.CaseID == definition.ID && item.NPCID == definition.KillerID && item.Source == "gate" && item.Destination != ""
		}
	}
	return hasTransfer && hasTravel
}

func validateExtortionV4(definition Definition, base world.World) bool {
	return validateWorkplaceConflict(definition, base)
}

func validateCaseRecords(definition Definition, base world.World) bool {
	if definition.GeneratorVersion >= connectedGeneratorVersion {
		return validateConnectedCaseRecords(definition, base)
	}
	if definition.ArchetypeID != "financial_extortion" {
		return len(definition.Records) == 0
	}
	if len(definition.Records) != 2 {
		return false
	}
	seen := make(map[string]bool, len(definition.Records))
	for _, item := range definition.Records {
		if item.ID == "" || item.NPCID == "" || item.OccurredAt.IsZero() || item.CaseID != definition.ID || seen[item.ID] {
			return false
		}
		seen[item.ID] = true
	}
	return true
}

func validateConnectedCaseRecords(definition Definition, base world.World) bool {
	if definition.CrimeAt.IsZero() {
		return false
	}
	expectedScene := regionName(base, locationRegion(base, definition.LocationID))
	scene, victimScene, killerScene := "", false, false
	var victimGate, killerGate, affairVictim, affairParamour *records.Record
	var victimPayment, killerReceipt *records.Record
	for index := range definition.Records {
		item := &definition.Records[index]
		if item.ID == "" || item.NPCID == "" || !containsNPC(base.NPCs, item.NPCID) || item.OccurredAt.IsZero() || item.CaseID != definition.ID || item.RelatedNPCID != "" && !containsNPC(base.NPCs, item.RelatedNPCID) {
			return false
		}
		switch item.InternalRelation {
		case "scene-victim":
			if item.Source != "gate" || item.NPCID != definition.VictimID || item.RelatedNPCID != definition.KillerID || !item.OccurredAt.Before(definition.CrimeAt) {
				return false
			}
			victimGate, scene, victimScene = item, item.Destination, true
		case "scene-killer":
			if item.Source != "gate" || item.NPCID != definition.KillerID || item.RelatedNPCID != definition.VictimID || !item.OccurredAt.Before(definition.CrimeAt) {
				return false
			}
			killerGate, scene, killerScene = item, item.Destination, true
		case "scene-killer-exit":
			if definition.GeneratorVersion < previousGeneratorVersion || item.Source != "gate" || item.NPCID != definition.KillerID || item.RelatedNPCID != definition.VictimID || !item.OccurredAt.After(definition.CrimeAt) || item.Origin != regionName(base, locationRegion(base, definition.LocationID)) {
				return false
			}
		case "affair-victim":
			affairVictim = item
		case "affair-paramour":
			affairParamour = item
		case "extortion_payment":
			if item.Source == "bank" && item.NPCID == definition.VictimID && item.Direction == "outgoing" {
				victimPayment = item
			}
			if item.Source == "bank" && item.NPCID == definition.KillerID && item.Direction == "incoming" {
				killerReceipt = item
			}
		}
	}
	if !victimScene || !killerScene || victimGate == nil || killerGate == nil || scene == "" || scene != expectedScene || victimGate.Destination != expectedScene || killerGate.Destination != expectedScene || victimGate.OccurredAt.After(killerGate.OccurredAt) {
		return false
	}
	if definition.ArchetypeID == "financial_extortion" {
		if victimPayment == nil || killerReceipt == nil || victimPayment.Amount != killerReceipt.Amount || victimPayment.Reference == "" || victimPayment.Reference != killerReceipt.Reference || !victimPayment.OccurredAt.Equal(killerReceipt.OccurredAt) {
			return false
		}
	}
	if definition.ArchetypeID == "jealousy" {
		if affairVictim == nil || affairParamour == nil || affairVictim.Source != "gate" || affairParamour.Source != "gate" || affairVictim.NPCID != definition.VictimID || affairParamour.NPCID != definition.ParamourID || affairVictim.Destination != expectedScene || affairParamour.Destination != expectedScene || !affairVictim.OccurredAt.Equal(affairParamour.OccurredAt) {
			return false
		}
	}
	return true
}

func validateConnectedCaseEmails(definition Definition, base world.World) bool {
	remorse := false
	serialThreat := false
	workplaceClueCount := 0
	workplaceName := companyByID(base.Companies, definition.CompanyID).Name
	for _, message := range definition.Emails {
		if message.InternalRelation == "killer_remorse" && message.MailboxNPCID == definition.KillerID && message.RelatedNPCID != "" && message.Sender == npcByID(base.NPCs, definition.KillerID).Name {
			remorse = true
		}
		if message.InternalRelation == "anonymous_warning" && message.MailboxNPCID == definition.VictimID && message.RelatedNPCID == definition.KillerID && message.Sender == "Unknown sender" && strings.Contains(message.Body, "work") {
			serialThreat = true
			if workplaceName != "" && strings.Contains(message.Body, workplaceName) {
				workplaceClueCount++
			}
		}
	}
	if definition.ArchetypeID == "serial_killer" {
		return serialThreat && workplaceClueCount == 1 && len(definition.Emails) == 5
	}
	return remorse
}

func regionName(base world.World, regionID string) string {
	for _, region := range base.Regions {
		if region.ID == regionID {
			return region.Name
		}
	}
	return "Lyon"
}

func hasEvidenceTypes(evidence []EvidenceDefinition, required ...string) bool {
	found := make(map[string]bool, len(evidence))
	for _, item := range evidence {
		found[item.Type] = true
	}
	for _, typeID := range required {
		if !found[typeID] {
			return false
		}
	}
	return true
}

func containsLocationInRegion(base world.World, locationID, regionID string) bool {
	for _, entity := range base.Entities {
		if entity.ID == locationID && entity.EntityType == "location" && entity.RegionID == regionID {
			return true
		}
	}
	return false
}

func hasReachableEvidence(evidence []EvidenceDefinition, definition Definition, relationships map[string]RelationshipReference, entities map[string]string) bool {
	if len(evidence) == 0 {
		return false
	}
	for _, clue := range evidence {
		if clue.ID == "" || clue.Type == "" || clue.Presentation == "" || clue.Content == "" || clue.LocationID != definition.LocationID || entities[clue.LocationID] != "location" {
			return false
		}
		if len(clue.RelationshipIDs) == 0 {
			return false
		}
		validLinks := true
		for _, relationshipID := range clue.RelationshipIDs {
			found := false
			for _, relationship := range relationships {
				if relationship.ID == relationshipID {
					found = true
					break
				}
			}
			if !found {
				validLinks = false
				break
			}
		}
		if !validLinks {
			return false
		}
		reachable := false
		for _, accessID := range clue.ReachableFromEntityIDs {
			if (accessID == definition.KillerID || accessID == definition.VictimID) && entities[accessID] == "npc" {
				reachable = true
				break
			}
		}
		if !reachable {
			return false
		}
	}
	return true
}

func hasHeadquarters(relations []world.CompanyLocation, companyID, locationID string) bool {
	for _, relation := range relations {
		if relation.CompanyID == companyID && relation.LocationID == locationID && relation.LocationRole == "headquarters" {
			return true
		}
	}
	return false
}

func companyExists(companies []world.Company, companyID string) bool {
	for _, company := range companies {
		if company.ID == companyID {
			return true
		}
	}
	return false
}

func containsNPC(npcs []world.NPC, id string) bool {
	for _, npc := range npcs {
		if npc.ID == id {
			return true
		}
	}
	return false
}

func contains(values []string, target string) bool {
	for _, value := range values {
		if value == target {
			return true
		}
	}
	return false
}

func caseIndex(seed string, version, count int) int {
	var versionBytes [8]byte
	binary.BigEndian.PutUint64(versionBytes[:], uint64(version))
	hash := sha256.New()
	writeField(hash, "noir-1974/case-generator")
	writeField(hash, string(versionBytes[:]))
	writeField(hash, seed)
	sum := hash.Sum(nil)
	return int(binary.BigEndian.Uint64(sum[:8]) % uint64(count))
}

func makeCaseID(seed string, version int) string {
	var versionBytes [8]byte
	binary.BigEndian.PutUint64(versionBytes[:], uint64(version))
	hash := sha256.New()
	writeField(hash, "noir-1974/case-id")
	writeField(hash, string(versionBytes[:]))
	writeField(hash, seed)
	return "case:" + hex.EncodeToString(hash.Sum(nil)[:12])
}

type byteWriter interface{ Write([]byte) (int, error) }

func writeField(writer byteWriter, value string) {
	var length [8]byte
	binary.BigEndian.PutUint64(length[:], uint64(len(value)))
	_, _ = writer.Write(length[:])
	_, _ = writer.Write([]byte(value))
}

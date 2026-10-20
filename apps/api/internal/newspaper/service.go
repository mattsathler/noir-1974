package newspaper

import (
	"context"
	"crypto/sha256"
	"encoding/binary"
	"fmt"
	"sort"
	"time"

	"noir-1974/apps/api/internal/cases"
	"noir-1974/apps/api/internal/world"
)

type worldReader interface {
	GeneratedWorld(context.Context, string) (world.World, error)
}

type caseReader interface {
	LatestPublicCase(context.Context, string) (cases.PublicCase, error)
}

type Service struct {
	worlds worldReader
	cases  caseReader
}

type Article struct {
	ID       string `json:"id"`
	Section  string `json:"section"`
	Headline string `json:"headline"`
	Summary  string `json:"summary"`
	CaseID   string `json:"caseId,omitempty"`
}

type Issue struct {
	Publication string    `json:"publication"`
	City        string    `json:"city"`
	Country     string    `json:"country"`
	IssueDate   string    `json:"issueDate"`
	Lead        Article   `json:"lead"`
	Articles    []Article `json:"articles"`
}

type newsTemplate struct {
	id       string
	headline string
	summary  string
}

func NewService(worlds worldReader, cases caseReader) *Service {
	return &Service{worlds: worlds, cases: cases}
}

// Issue creates a deterministic newspaper issue and keeps the latest case as its lead.
func (service *Service) Issue(ctx context.Context, serverID, city, country string, serverTime time.Time) (Issue, error) {
	base, err := service.worlds.GeneratedWorld(ctx, serverID)
	if err != nil {
		return Issue{}, err
	}
	latest, err := service.cases.LatestPublicCase(ctx, serverID)
	if err != nil {
		return Issue{}, err
	}
	if serverTime.IsZero() {
		return Issue{}, fmt.Errorf("invalid newspaper date")
	}
	localTime := serverTime.In(time.FixedZone("CET", 60*60))
	dateKey := localTime.Format("2006-01-02")
	dateLabel := localTime.Format("Monday, January 2, 2006")

	npcNames := make(map[string]string, len(base.NPCs))
	for _, npc := range base.NPCs {
		npcNames[npc.ID] = npc.Name
	}
	locations := make(map[string]world.Entity, len(base.Entities))
	for _, entity := range base.Entities {
		if entity.EntityType == "location" {
			locations[entity.ID] = entity
		}
	}
	victimName, exists := npcNames[latest.VictimID]
	if !exists {
		return Issue{}, fmt.Errorf("latest case victim is missing from generated world")
	}
	location, exists := locations[latest.LocationID]
	if !exists {
		return Issue{}, fmt.Errorf("latest case location is missing from generated world")
	}

	lead := Article{
		ID:       latest.ID,
		Section:  "SPECIAL REPORT",
		Headline: fmt.Sprintf("Police open homicide inquiry at %s", location.Name),
		Summary:  fmt.Sprintf("Authorities confirmed the death of %s at %s, in %s. Detectives are asking anyone with information to come forward. The investigation remains open.", victimName, location.Name, city),
		CaseID:   latest.ID,
	}
	articles := makeLightStories(base, serverID, dateKey)
	return Issue{
		Publication: fmt.Sprintf("THE %s GAZETTE", city),
		City:        city,
		Country:     country,
		IssueDate:   dateLabel,
		Lead:        lead,
		Articles:    articles,
	}, nil
}

func makeLightStories(base world.World, serverID, dateKey string) []Article {
	templates := []newsTemplate{
		{id: "town-clock", headline: "Council asked to settle the town clock dispute", summary: "Residents disagree over whether the clock is four minutes fast or the railway is three minutes slow."},
		{id: "bench", headline: "%s residents petition to keep a park bench exactly where it is", summary: "A neighborhood committee says the bench has been moved twice this month, though no one can agree on which direction."},
		{id: "pigeon", headline: "Pigeon delays morning tram inspection", summary: "The bird occupied the inspector's seat and declined several polite requests to leave."},
		{id: "spoons", headline: "%s café staff form committee over missing teaspoons", summary: "The staff insist the teaspoons are not missing, merely migrating toward the kitchen at an alarming rate."},
		{id: "croissant", headline: "Local baker denies reports of shrinking croissants", summary: "The baker blamed the measuring tape and offered no comment on the customers' breakfast scales."},
		{id: "umbrella", headline: "%s shopkeeper warns residents against wearing umbrellas as hats", summary: "The warning followed three damp customers and one very surprised umbrella salesman."},
		{id: "cat", headline: "Alley cat appointed unofficial supervisor of late deliveries", summary: "The cat has not acknowledged the appointment but continues to inspect every parcel in the lane."},
		{id: "bell", headline: "Bicycle bell mistaken for a new municipal signal", summary: "Officials confirmed that the sound means only that someone would like to pass."},
		{id: "chess", headline: "%s chess club bans advice from spectators", summary: "The rule was adopted unanimously after a match paused for twenty minutes of unsolicited strategy."},
	}

	districts := make([]string, 0)
	for _, region := range base.Regions {
		if region.Kind == "district" {
			districts = append(districts, region.Name)
		}
	}
	sort.Strings(districts)
	if len(districts) == 0 {
		districts = []string{"Bellecour"}
	}
	key := serverID + "/" + dateKey
	sort.Slice(templates, func(i, j int) bool {
		return storyRank(key, templates[i].id) < storyRank(key, templates[j].id)
	})
	articles := make([]Article, 0, 6)
	for _, template := range templates[:4] {
		headline := template.headline
		if containsFormatToken(headline) {
			headline = fmt.Sprintf(headline, districts[int(storyRank(key, template.id+"/district")%uint64(len(districts)))])
		}
		articles = append(articles, Article{ID: template.id, Section: "CITY NOTES", Headline: headline, Summary: template.summary})
	}
	articles = append(articles, makeCompanyStories(base, key)...)
	return articles
}

func makeCompanyStories(base world.World, key string) []Article {
	companies := append([]world.Company(nil), base.Companies...)
	sort.Slice(companies, func(i, j int) bool {
		return storyRank(key, "company/"+companies[i].ID) < storyRank(key, "company/"+companies[j].ID)
	})
	if len(companies) > 2 {
		companies = companies[:2]
	}

	npcNames := make(map[string]string, len(base.NPCs))
	for _, npc := range base.NPCs {
		npcNames[npc.ID] = npc.Name
	}
	companyRegions := make(map[string]string, len(base.Companies))
	for _, company := range base.Companies {
		companyRegions[company.ID] = company.RegionID
	}
	districtNames := make(map[string]string, len(base.Regions))
	for _, region := range base.Regions {
		if region.Kind == "district" {
			districtNames[region.ID] = region.Name
		}
	}
	var stories []Article
	for _, company := range companies {
		staff := make([]world.EmploymentRelationship, 0)
		for _, relation := range base.EmploymentRelationships {
			if relation.EmployerCompanyID == company.ID {
				staff = append(staff, relation)
			}
		}
		if len(staff) == 0 {
			continue
		}
		sort.Slice(staff, func(i, j int) bool { return staff[i].EmployeeNPCID < staff[j].EmployeeNPCID })
		staffRank := storyRank(key, "staff/"+company.ID)
		person := staff[int(staffRank%uint64(len(staff)))]
		name := npcNames[person.EmployeeNPCID]
		if name == "" {
			continue
		}
		district := districtNames[companyRegions[company.ID]]
		if district == "" {
			district = "the city"
		}
		variant := storyRank(key, "company-story/"+company.ID) % 2
		headline := fmt.Sprintf("%s reviews the review process after review", company.Name)
		summary := fmt.Sprintf("At its %s office in %s, %s said, ‘The new schedule should make routine decisions easier to track.’ The schedule will be reviewed again next month.", person.Role, district, name)
		if variant == 1 {
			headline = fmt.Sprintf("%s introduces a form to request fewer forms", company.Name)
			summary = fmt.Sprintf("%s, a %s at the %s headquarters, said, ‘The new form should simplify internal requests.’ A pilot group will first complete a form to join the pilot.", name, person.Role, district)
		}
		stories = append(stories, Article{
			ID:       "company-" + company.ID,
			Section:  "BUSINESS & COMMERCE",
			Headline: headline,
			Summary:  summary,
		})
	}
	return stories
}

func storyRank(key, id string) uint64 {
	digest := sha256.Sum256([]byte("noir-1974/newspaper/" + key + "/" + id))
	return binary.BigEndian.Uint64(digest[:8])
}

func containsFormatToken(value string) bool {
	for _, character := range value {
		if character == '%' {
			return true
		}
	}
	return false
}

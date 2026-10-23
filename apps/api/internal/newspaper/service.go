package newspaper

import (
	"context"
	"crypto/sha256"
	"encoding/binary"
	"errors"
	"fmt"
	"sort"
	"sync"
	"time"

	"noir-1974/apps/api/internal/cases"
	"noir-1974/apps/api/internal/world"
)

type worldReader interface {
	CityWorldForCases(context.Context, string, string) (world.World, error)
}

type caseReader interface {
	LatestPublicCase(context.Context, string, string) (cases.PublicCase, error)
}
type arrestReader interface {
	LatestPublicArrest(context.Context, string, string) (cases.PublicCase, error)
}

type Service struct {
	worlds worldReader
	cases  caseReader
	mu     sync.Mutex
	issues map[string]Issue
	flight map[string]*issueFlight
}

type issueFlight struct {
	done  chan struct{}
	issue Issue
	err   error
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
	return &Service{worlds: worlds, cases: cases, issues: make(map[string]Issue), flight: make(map[string]*issueFlight)}
}

func (service *Service) InvalidateCity(serverID, cityID string) {
	prefix := serverID + "/" + cityID + "/"
	service.mu.Lock()
	defer service.mu.Unlock()
	for key := range service.issues {
		if len(key) >= len(prefix) && key[:len(prefix)] == prefix {
			delete(service.issues, key)
		}
	}
}

// Issue caches one generated edition per server, city, and in-game calendar day.
func (service *Service) Issue(ctx context.Context, serverID, cityID, city, country string, serverTime time.Time) (Issue, error) {
	if serverTime.IsZero() {
		return Issue{}, fmt.Errorf("invalid newspaper date")
	}
	zone := time.FixedZone("CET", 60*60)
	if cityData, ok := world.CityByID(cityID); ok {
		if cityZone, err := time.LoadLocation(cityData.TimeZone); err == nil {
			zone = cityZone
		}
	}
	dateKey := serverTime.In(zone).Format("2006-01-02")
	cacheKey := serverID + "/" + cityID + "/" + dateKey
	service.mu.Lock()
	if issue, ok := service.issues[cacheKey]; ok {
		service.mu.Unlock()
		return issue, nil
	}
	if flight := service.flight[cacheKey]; flight != nil {
		service.mu.Unlock()
		select {
		case <-ctx.Done():
			return Issue{}, ctx.Err()
		case <-flight.done:
			return flight.issue, flight.err
		}
	}
	flight := &issueFlight{done: make(chan struct{})}
	service.flight[cacheKey] = flight
	service.mu.Unlock()

	issue, err := service.buildIssue(ctx, serverID, cityID, city, country, serverTime)
	service.mu.Lock()
	if err == nil {
		service.issues[cacheKey] = issue
		cityPrefix := serverID + "/" + cityID + "/"
		for key := range service.issues {
			if key != cacheKey && len(key) >= len(cityPrefix) && key[:len(cityPrefix)] == cityPrefix {
				delete(service.issues, key)
			}
		}
	}
	flight.issue, flight.err = issue, err
	delete(service.flight, cacheKey)
	close(flight.done)
	service.mu.Unlock()
	return issue, err
}

// buildIssue creates a deterministic city edition with its newest local case as lead.
func (service *Service) buildIssue(ctx context.Context, serverID, cityID, city, country string, serverTime time.Time) (Issue, error) {
	base, err := service.worlds.CityWorldForCases(ctx, serverID, cityID)
	if err != nil {
		return Issue{}, fmt.Errorf("reconstruct city newspaper data: %w", err)
	}
	latest, caseErr := service.cases.LatestPublicCase(ctx, serverID, cityID)
	if caseErr != nil && !errors.Is(caseErr, cases.ErrCaseNotFound) {
		return Issue{}, fmt.Errorf("query latest case for city %s: %w", cityID, caseErr)
	}
	if serverTime.IsZero() {
		return Issue{}, fmt.Errorf("invalid newspaper date")
	}
	location := time.FixedZone("CET", 60*60)
	if cityData, ok := world.CityByID(cityID); ok {
		if cityLocation, err := time.LoadLocation(cityData.TimeZone); err == nil {
			location = cityLocation
		}
	}
	localTime := serverTime.In(location)
	dateKey := localTime.Format("2006-01-02")
	dateLabel := localTime.Format("Monday, January 2")

	local := base
	npcNames := make(map[string]string, len(local.NPCs))
	for _, npc := range local.NPCs {
		npcNames[npc.ID] = npc.Name
	}
	lead := Article{ID: "city-desk-" + cityID + "-" + dateKey, Section: "CITY DESK", Headline: "A quiet day for the city police", Summary: fmt.Sprintf("No new homicide inquiry has been reported in %s. Local authorities remain on duty.", city)}
	if caseErr == nil {
		victimName, exists := npcNames[latest.VictimID]
		if !exists {
			return Issue{}, fmt.Errorf("latest case victim is missing from generated city")
		}
		headline := fmt.Sprintf("Police open homicide inquiry into death of %s", victimName)
		summary := fmt.Sprintf("Authorities confirmed the death of %s. Detectives are asking anyone with information to come forward. The circumstances remain under investigation.", victimName)
		if latest.Status == "solved" {
			headline = fmt.Sprintf("Police announce arrest in the death of %s", victimName)
			summary = fmt.Sprintf("City authorities confirmed that a suspect has been taken into custody in the case concerning %s. The corporate review is complete.", victimName)
		}
		lead = Article{
			ID: latest.ID, Section: "SPECIAL REPORT",
			Headline: headline,
			Summary:  summary,
			CaseID:   latest.ID,
		}
	}
	articles := makeLightStories(local, serverID+"/"+cityID, dateKey)
	if reader, ok := service.cases.(arrestReader); ok {
		arrest, err := reader.LatestPublicArrest(ctx, serverID, cityID)
		if err != nil && !errors.Is(err, cases.ErrCaseNotFound) {
			return Issue{}, fmt.Errorf("query city arrest notice: %w", err)
		}
		if err == nil && arrest.ID != latest.ID {
			if name, found := npcNames[arrest.VictimID]; found {
				articles = append([]Article{{ID: arrest.ID + "-arrest", Section: "CITY POLICE", Headline: fmt.Sprintf("Police announce arrest in the death of %s", name), Summary: fmt.Sprintf("A suspect has been taken into custody in the homicide inquiry concerning %s. The corporate review has concluded.", name), CaseID: arrest.ID}}, articles...)
			}
		}
	}
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

package newspaper

import (
	"context"
	"crypto/sha256"
	"encoding/binary"
	"errors"
	"fmt"
	"sort"
	"strings"
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
type bodyDiscoveryReader interface {
	ListPublicBodyDiscoveries(context.Context, string, string, time.Time) ([]cases.PublicBodyDiscovery, error)
}

type Service struct {
	worlds worldReader
	cases  caseReader
	mu     sync.Mutex
	issues map[string]cachedIssue
	flight map[string]*issueFlight
}

type cachedIssue struct {
	issue     Issue
	updatedAt time.Time
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
	return &Service{worlds: worlds, cases: cases, issues: make(map[string]cachedIssue), flight: make(map[string]*issueFlight)}
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

func (service *Service) InvalidateAll() {
	service.mu.Lock()
	defer service.mu.Unlock()
	clear(service.issues)
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
	if cached, ok := service.issues[cacheKey]; ok && time.Since(cached.updatedAt) < time.Minute {
		service.mu.Unlock()
		return cached.issue, nil
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
		service.issues[cacheKey] = cachedIssue{issue: issue, updatedAt: time.Now()}
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

// buildIssue creates a city edition whose crime reporting is limited to confirmed body discoveries.
func (service *Service) buildIssue(ctx context.Context, serverID, cityID, city, country string, serverTime time.Time) (Issue, error) {
	base, err := service.worlds.CityWorldForCases(ctx, serverID, cityID)
	if err != nil {
		return Issue{}, fmt.Errorf("reconstruct city newspaper data: %w", err)
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
	lead := Article{ID: "city-desk-" + cityID + "-" + dateKey, Section: "CITY DESK", Headline: "A quiet day for the city police", Summary: fmt.Sprintf("There are no current homicide inquiries or new arrests in %s. Local authorities remain on duty.", city)}
	articles := makeLightStories(local, serverID+"/"+cityID, dateKey)
	// The political conflict is part of the city's daily atmosphere, not a
	// one-off case event. Keep both viewpoints in every edition so the player
	// repeatedly encounters the regime's framing and the Crow's justification.
	articles = append(makeConflictStories(serverID+"/"+cityID, dateKey), articles...)
	var discoveries []cases.PublicBodyDiscovery
	if reader, ok := service.cases.(bodyDiscoveryReader); ok {
		discoveries, err = reader.ListPublicBodyDiscoveries(ctx, serverID, cityID, serverTime)
		if err != nil {
			return Issue{}, fmt.Errorf("query discovered bodies for city %s: %w", cityID, err)
		}
		bodyArticles := make([]Article, 0, len(discoveries))
		openDiscoveries := make([]cases.PublicBodyDiscovery, 0, len(discoveries))
		for _, discovery := range discoveries {
			if discovery.Status != "open" {
				continue
			}
			openDiscoveries = append(openDiscoveries, discovery)
			bodyArticles = append(bodyArticles, bodyDiscoveryArticle(discovery))
		}
		discoveries = openDiscoveries
		articles = append(bodyArticles, articles...)
	}
	activeCase, activeErr := service.cases.LatestPublicCase(ctx, serverID, cityID)
	if activeErr != nil && !errors.Is(activeErr, cases.ErrCaseNotFound) {
		return Issue{}, fmt.Errorf("query current city case: %w", activeErr)
	}
	hasCurrentCaseLead := false
	if activeErr == nil && activeCase.Status == "open" {
		if name, found := npcNames[activeCase.VictimID]; found {
			lead = Article{ID: activeCase.ID + "-current", Section: "CITY POLICE", Headline: fmt.Sprintf("Police continue inquiry into the death of %s", name), Summary: fmt.Sprintf("The homicide inquiry concerning %s remains open. Authorities continue to investigate.", name), CaseID: activeCase.ID}
			hasCurrentCaseLead = true
			for _, discovery := range discoveries {
				if discovery.CaseID == activeCase.ID {
					lead = bodyDiscoveryArticle(discovery)
					break
				}
			}
		}
	}
	if reader, ok := service.cases.(arrestReader); ok {
		arrest, err := reader.LatestPublicArrest(ctx, serverID, cityID)
		if err != nil && !errors.Is(err, cases.ErrCaseNotFound) {
			return Issue{}, fmt.Errorf("query city arrest notice: %w", err)
		}
		if err == nil {
			if name, found := npcNames[arrest.VictimID]; found {
				arrestArticle := Article{ID: arrest.ID + "-arrest", Section: "CITY POLICE", Headline: fmt.Sprintf("Police announce arrest in the death of %s", name), Summary: fmt.Sprintf("A suspect has been taken into custody in the homicide inquiry concerning %s. The corporate review has concluded.", name), CaseID: arrest.ID}
				if hasCurrentCaseLead {
					articles = append([]Article{arrestArticle}, articles...)
				} else {
					lead = arrestArticle
				}
			}
		}
	}
	filteredArticles := articles[:0]
	bodyStories := 0
	bodyStoryLimit := 3
	if strings.HasPrefix(lead.ID, "body-found-") {
		bodyStoryLimit--
	}
	for _, article := range articles {
		if article.ID == lead.ID {
			continue
		}
		if strings.HasPrefix(article.ID, "body-found-") {
			if bodyStories >= bodyStoryLimit {
				continue
			}
			bodyStories++
		}
		filteredArticles = append(filteredArticles, article)
	}
	articles = filteredArticles
	if len(articles) > 6 {
		articles = articles[:6]
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

func makeConflictStories(key, dateKey string) []Article {
	dayKey := key + "/" + dateKey
	billReports := []newsTemplate{
		{id: "bill-bouer-crow-01", headline: "Bill Bouer: The Crow's war on the system leaves ordinary residents exposed", summary: "In a broadcast carried by the state news service, investigator Bill Bouer said The Crow's attacks on corporate records put workers and residents at risk. A skull mark found on the breached terminal is being treated as the fugitive's signature. The Crow calls his campaign resistance; police say it is theft and sabotage."},
		{id: "bill-bouer-crow-02", headline: "Bill Bouer reports new lead in the hunt for The Crow", summary: "Bouer said investigators are tracing stolen corporate files and a skull mark left at the scene. He urged residents to report suspicious access. The Crow's supporters describe the raids as a strike against institutions that answer to no one."},
		{id: "bill-bouer-crow-03", headline: "Bouer warns that The Crow's latest raid was no harmless protest", summary: "Bill Bouer said the breach disrupted services used by the public and pointed to the skull painted near the access point. The Crow claims the target financed state surveillance; officials reject that account and promise a wider investigation."},
	}
	appWarnings := []newsTemplate{
		{id: "illegal-apps-01", headline: "Police warn residents against Ghostkey and other illicit applications", summary: "Authorities say bypass tools such as Ghostkey can expose private records and bring severe penalties. The notice calls their use a threat to public order; underground users say the same tools reveal what officials hide."},
		{id: "illegal-apps-02", headline: "Possession of unauthorized network tools may draw prosecution, ministry warns", summary: "The ministry named Ghostkey among applications used to break into protected systems. Investigators link such tools to The Crow's thefts, while the fugitive argues that locked corporate files conceal abuses."},
		{id: "illegal-apps-03", headline: "City bulletin: report Ghostkey and unlicensed access software", summary: "The public notice warns that installing illicit applications may be treated as intent to breach state or corporate networks. Critics say the rule makes even looking for evidence a crime."},
	}
	selectStory := func(group []newsTemplate, slot string) Article {
		index := int(storyRank(dayKey, slot) % uint64(len(group)))
		item := group[index]
		return Article{ID: item.id, Section: "STATE WATCH", Headline: item.headline, Summary: item.summary}
	}
	return []Article{selectStory(billReports, "bill-report"), selectStory(appWarnings, "illicit-app-warning")}
}

func bodyDiscoveryArticle(discovery cases.PublicBodyDiscovery) Article {
	place := discovery.LocationName
	if discovery.DistrictName != "" {
		place += ", " + discovery.DistrictName
	}
	return Article{
		ID:       "body-found-" + discovery.CaseID,
		Section:  "BODY FOUND",
		Headline: fmt.Sprintf("Body of %s discovered in %s", discovery.VictimName, place),
		Summary:  fmt.Sprintf("Authorities reported finding %s at %s. The circumstances remain under investigation.", discovery.VictimName, place),
		CaseID:   discovery.CaseID,
	}
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

package newspaper

import (
	"context"
	"errors"
	"strconv"
	"strings"
	"testing"
	"time"

	"noir-1974/apps/api/internal/cases"
	"noir-1974/apps/api/internal/world"
)

type newspaperWorldFixture struct{ base world.World }

func (fixture newspaperWorldFixture) CityWorldForCases(context.Context, string, string) (world.World, error) {
	return fixture.base, nil
}

type newspaperCaseFixture struct {
	active      cases.PublicCase
	activeErr   error
	arrest      cases.PublicCase
	arrestErr   error
	discoveries []cases.PublicBodyDiscovery
}

func (fixture newspaperCaseFixture) LatestPublicCase(context.Context, string, string) (cases.PublicCase, error) {
	return fixture.active, fixture.activeErr
}

func (fixture newspaperCaseFixture) LatestPublicArrest(context.Context, string, string) (cases.PublicCase, error) {
	return fixture.arrest, fixture.arrestErr
}

func (fixture newspaperCaseFixture) ListPublicBodyDiscoveries(context.Context, string, string, time.Time) ([]cases.PublicBodyDiscovery, error) {
	return fixture.discoveries, nil
}

func TestIssueLeadsWithArrestInsteadOfSolvedBodyDiscovery(t *testing.T) {
	fixture := newspaperCaseFixture{
		activeErr: cases.ErrCaseNotFound,
		arrest:    cases.PublicCase{ID: "solved-case", VictimID: "npc-victim"},
		discoveries: []cases.PublicBodyDiscovery{{
			CaseID: "solved-case", Status: "solved", VictimName: "Gaston Lefebvre", LocationName: "Café des Écoles", DistrictName: "Bas Théâtre",
			FoundAt: time.Date(2026, time.October, 28, 0, 54, 0, 0, time.UTC),
		}},
	}
	issue, err := NewService(newspaperWorldFixture{base: world.World{NPCs: []world.NPC{{ID: "npc-victim", Name: "Gaston Lefebvre"}}}}, fixture).
		buildIssue(context.Background(), "server", "fr/lyon", "Lyon", "France", time.Date(2026, time.October, 28, 12, 0, 0, 0, time.UTC))
	if err != nil {
		t.Fatal(err)
	}
	if !strings.Contains(issue.Lead.Headline, "arrest") || issue.Lead.CaseID != "solved-case" {
		t.Fatalf("lead = %+v; want the arrest notice for the solved case", issue.Lead)
	}
	for _, article := range issue.Articles {
		if article.ID == "body-found-solved-case" {
			t.Fatalf("solved case also appears as a body discovery: %+v", article)
		}
	}
	if strings.Contains(issue.Lead.Summary, "12:54") || strings.Contains(issue.Lead.Summary, "AM") || strings.Contains(issue.Lead.Summary, "PM") {
		t.Fatalf("lead mentions a time of death or discovery: %q", issue.Lead.Summary)
	}
}

func TestIssueDoesNotLeadWithSolvedCaseWithoutArrest(t *testing.T) {
	fixture := newspaperCaseFixture{
		activeErr: cases.ErrCaseNotFound,
		arrestErr: cases.ErrCaseNotFound,
		discoveries: []cases.PublicBodyDiscovery{{
			CaseID: "solved-case", Status: "solved", VictimName: "Gaston Lefebvre", LocationName: "Café des Écoles",
		}},
	}
	issue, err := NewService(newspaperWorldFixture{}, fixture).
		buildIssue(context.Background(), "server", "fr/lyon", "Lyon", "France", time.Date(2026, time.October, 28, 12, 0, 0, 0, time.UTC))
	if err != nil {
		t.Fatal(err)
	}
	if issue.Lead.ID == "body-found-solved-case" {
		t.Fatalf("solved case became the lead: %+v", issue.Lead)
	}
}

func TestIssueLeadsWithOpenCaseAndOmitsDiscoveryTime(t *testing.T) {
	fixture := newspaperCaseFixture{
		active:    cases.PublicCase{ID: "open-case", VictimID: "npc-victim", Status: "open"},
		arrestErr: cases.ErrCaseNotFound,
		discoveries: []cases.PublicBodyDiscovery{{
			CaseID: "open-case", Status: "open", VictimName: "Gaston Lefebvre", LocationName: "Café des Écoles", DistrictName: "Bas Théâtre",
			FoundAt: time.Date(2026, time.October, 28, 0, 54, 0, 0, time.UTC),
		}},
	}
	issue, err := NewService(newspaperWorldFixture{base: world.World{NPCs: []world.NPC{{ID: "npc-victim", Name: "Gaston Lefebvre"}}}}, fixture).
		buildIssue(context.Background(), "server", "fr/lyon", "Lyon", "France", time.Date(2026, time.October, 28, 12, 0, 0, 0, time.UTC))
	if err != nil {
		t.Fatal(err)
	}
	if issue.Lead.CaseID != "open-case" || !strings.Contains(issue.Lead.Headline, "Body of Gaston Lefebvre") {
		t.Fatalf("lead = %+v; want the discovery for the open case", issue.Lead)
	}
	if strings.Contains(issue.Lead.Summary, "12:54") || strings.Contains(issue.Lead.Summary, "AM") || strings.Contains(issue.Lead.Summary, "PM") {
		t.Fatalf("lead mentions a time of death or discovery: %q", issue.Lead.Summary)
	}
}

func TestIssueReportsOpenInquiryWithoutBodyDiscovery(t *testing.T) {
	fixture := newspaperCaseFixture{
		active:    cases.PublicCase{ID: "open-case", VictimID: "npc-victim", Status: "open"},
		arrestErr: cases.ErrCaseNotFound,
	}
	issue, err := NewService(newspaperWorldFixture{base: world.World{NPCs: []world.NPC{{ID: "npc-victim", Name: "Gaston Lefebvre"}}}}, fixture).
		buildIssue(context.Background(), "server", "fr/lyon", "Lyon", "France", time.Date(2026, time.October, 28, 12, 0, 0, 0, time.UTC))
	if err != nil {
		t.Fatal(err)
	}
	if issue.Lead.CaseID != "open-case" || !strings.Contains(issue.Lead.Headline, "Gaston Lefebvre") {
		t.Fatalf("lead = %+v; want the current open inquiry", issue.Lead)
	}
}

func TestIssueLimitsSubnewsToSixAndKeepsArrestNotice(t *testing.T) {
	discoveries := make([]cases.PublicBodyDiscovery, 8)
	for i := range discoveries {
		discoveries[i] = cases.PublicBodyDiscovery{CaseID: "open-case-" + strconv.Itoa(i), Status: "open", VictimName: "Resident", LocationName: "Café"}
	}
	fixture := newspaperCaseFixture{
		active:      cases.PublicCase{ID: "open-case-0", VictimID: "npc-victim", Status: "open"},
		arrest:      cases.PublicCase{ID: "arrest-case", VictimID: "npc-victim"},
		discoveries: discoveries,
	}
	issue, err := NewService(newspaperWorldFixture{base: world.World{NPCs: []world.NPC{{ID: "npc-victim", Name: "Resident"}}}}, fixture).
		buildIssue(context.Background(), "server", "fr/lyon", "Lyon", "France", time.Date(2026, time.October, 28, 12, 0, 0, 0, time.UTC))
	if err != nil {
		t.Fatal(err)
	}
	if len(issue.Articles) != 6 {
		t.Fatalf("subnews count = %d, want 6", len(issue.Articles))
	}
	if issue.Articles[0].ID != "arrest-case-arrest" {
		t.Fatalf("first subnews = %+v, want the arrest notice", issue.Articles[0])
	}
}

func TestIssueReturnsCaseReaderErrors(t *testing.T) {
	failure := errors.New("store unavailable")
	fixture := newspaperCaseFixture{activeErr: failure, arrestErr: cases.ErrCaseNotFound}
	_, err := NewService(newspaperWorldFixture{}, fixture).
		buildIssue(context.Background(), "server", "fr/lyon", "Lyon", "France", time.Date(2026, time.October, 28, 12, 0, 0, 0, time.UTC))
	if !errors.Is(err, failure) {
		t.Fatalf("buildIssue() error = %v, want %v", err, failure)
	}
}

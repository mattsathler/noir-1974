package records

import (
	"context"
	"time"

	"noir-1974/apps/api/internal/world"
)

type worldReader interface {
	GetNPCForRecords(context.Context, string, string, bool) (world.NPCRecordsProfile, error)
}

type caseReader interface {
	NPCRecords(context.Context, string, string, time.Time, time.Time, string) ([]Record, error)
}

type Service struct {
	worlds worldReader
	cases  caseReader
}

func NewService(worlds worldReader, cases caseReader) *Service {
	return &Service{worlds: worlds, cases: cases}
}

// NPCRecords creates routine activity on demand and mixes in the person's
// case-related ledger entries before sorting the complete, unmarked timeline.
func (service *Service) NPCRecords(ctx context.Context, serverID, npcID string, from, through time.Time, source string) ([]Record, error) {
	profile, err := service.worlds.GetNPCForRecords(ctx, serverID, npcID, source != "gate")
	if err != nil {
		return nil, err
	}
	regionName := ""
	districts := []string(nil)
	if source != "bank" {
		base, err := world.Generate(profile.WorldSeed, profile.GeneratorVersion)
		if err != nil {
			return nil, err
		}
		regionName = "Lyon"
		districts = make([]string, 0, len(base.Regions))
		for _, region := range base.Regions {
			if region.ID == profile.NPC.RegionID {
				regionName = region.Name
			}
			if region.Kind == "district" {
				districts = append(districts, region.Name)
			}
		}
	}
	employer := profile.Employer
	if employer == "" {
		employer = "Independent account"
	}
	ordinary, err := GenerateForSource(profile.WorldSeed, profile.GeneratorVersion, npcID, Context{RegionName: "Lyon", HomeDistrict: regionName, Employer: employer, Districts: districts, DailySalary: profile.DailySalary, Role: profile.Role, EmployerID: profile.EmployerID, Benefits: profile.EmployerBenefits}, from, through, source)
	if err != nil {
		return nil, err
	}
	linked, err := service.cases.NPCRecords(ctx, serverID, npcID, from, through, source)
	if err != nil {
		return nil, err
	}
	all := append(ordinary, linked...)
	availableAt := through
	if current := time.Now().UTC(); availableAt.After(current) {
		availableAt = current
	}
	for index := range all {
		if all[index].OccurredAt.After(availableAt) {
			continue
		}
		all[index].CaseID = ""
		all[index].InternalRelation = ""
		all[index].RelatedNPCID = ""
	}
	available := all[:0]
	for _, item := range all {
		if !item.OccurredAt.After(availableAt) {
			available = append(available, item)
		}
	}
	sortRecords(available)
	return available, nil
}

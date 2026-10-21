package records

import (
	"context"
	"time"

	"noir-1974/apps/api/internal/world"
)

type worldReader interface {
	WorldConfig(context.Context, string) (string, int, error)
	GeneratedWorld(context.Context, string) (world.World, error)
	GetNPC(context.Context, string, string) (world.NPCProfile, error)
}

type caseReader interface {
	NPCRecords(context.Context, string, string, time.Time, time.Time) ([]Record, error)
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
func (service *Service) NPCRecords(ctx context.Context, serverID, npcID string, from, through time.Time) ([]Record, error) {
	seed, version, err := service.worlds.WorldConfig(ctx, serverID)
	if err != nil {
		return nil, err
	}
	profile, err := service.worlds.GetNPC(ctx, serverID, npcID)
	if err != nil {
		return nil, err
	}
	base, err := service.worlds.GeneratedWorld(ctx, serverID)
	if err != nil {
		return nil, err
	}
	regionName := "Lyon"
	districts := make([]string, 0, len(base.Regions))
	for _, region := range base.Regions {
		if region.ID == profile.NPC.RegionID {
			regionName = region.Name
		}
		if region.Kind == "district" {
			districts = append(districts, region.Name)
		}
	}
	employer := "Independent account"
	if len(profile.Employment) > 0 {
		employer = profile.Employment[0].Company.Name
	}
	ordinary, err := Generate(seed, version, npcID, Context{RegionName: "Lyon", HomeDistrict: regionName, Employer: employer, Districts: districts}, from, through)
	if err != nil {
		return nil, err
	}
	linked, err := service.cases.NPCRecords(ctx, serverID, npcID, from, through)
	if err != nil {
		return nil, err
	}
	all := append(ordinary, linked...)
	for index := range all {
		all[index].CaseID = ""
		all[index].InternalRelation = ""
		all[index].RelatedNPCID = ""
	}
	sortRecords(all)
	return all, nil
}

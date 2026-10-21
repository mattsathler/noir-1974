package mail

import (
	"context"
	"time"

	"noir-1974/apps/api/internal/world"
)

type worldReader interface {
	WorldConfig(context.Context, string) (string, int, error)
	GetNPC(context.Context, string, string) (world.NPCProfile, error)
}

type caseReader interface {
	NPCEmails(context.Context, string, string, time.Time, time.Time) ([]Message, error)
}

type Service struct {
	worlds worldReader
	cases  caseReader
}

func NewService(worlds worldReader, cases caseReader) *Service {
	return &Service{worlds: worlds, cases: cases}
}

// NPCEmails generates ordinary mail for the requested mailbox and interval,
// then mixes in case-associated messages without exposing their classification.
func (service *Service) NPCEmails(ctx context.Context, serverID, npcID string, from, through time.Time) ([]Message, error) {
	seed, version, err := service.worlds.WorldConfig(ctx, serverID)
	if err != nil {
		return nil, err
	}
	profile, err := service.worlds.GetNPC(ctx, serverID, npcID)
	if err != nil {
		return nil, err
	}
	context := Context{NPCName: profile.NPC.Name}
	if len(profile.Employment) > 0 {
		context.Employer = profile.Employment[0].Company.Name
	}
	if len(profile.PersonalRelationships) > 0 {
		context.PersonalContact = profile.PersonalRelationships[0].NPC.Name
	}
	ordinary, err := Generate(seed, version, npcID, context, from, through)
	if err != nil {
		return nil, err
	}
	linked, err := service.cases.NPCEmails(ctx, serverID, npcID, from, through)
	if err != nil {
		return nil, err
	}
	all := append(ordinary, linked...)
	SortMessages(all)
	return all, nil
}

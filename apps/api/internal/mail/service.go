package mail

import (
	"context"
	"errors"
	"fmt"
	"time"

	"noir-1974/apps/api/internal/world"
)

type worldReader interface {
	WorldConfig(context.Context, string) (string, int, error)
	GetNPC(context.Context, string, string) (world.NPCProfile, error)
	EmployeesOfCompany(context.Context, string, string) ([]world.CompanyEmployee, error)
}

type generatedWorldReader interface {
	WorldForCases(context.Context, string) (world.World, error)
}

type batchEmployeesReader interface {
	EmployeesOfCompanies(context.Context, string, string, int, []string) (map[string][]world.CompanyEmployee, error)
}

type configuredNPCReader interface {
	GetNPCWithWorldConfig(context.Context, string, string) (world.NPCProfile, string, int, error)
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
	var seed string
	var version int
	var profile world.NPCProfile
	var err error
	if reader, ok := service.worlds.(configuredNPCReader); ok {
		profile, seed, version, err = reader.GetNPCWithWorldConfig(ctx, serverID, npcID)
		if errors.Is(err, world.ErrEntityNotFound) {
			seed, version, err = service.worlds.WorldConfig(ctx, serverID)
			if err != nil {
				return nil, err
			}
		} else if err != nil {
			return nil, err
		}
	} else {
		seed, version, err = service.worlds.WorldConfig(ctx, serverID)
		if err != nil {
			return nil, err
		}
		profile, err = service.worlds.GetNPC(ctx, serverID, npcID)
	}
	var mailboxContext Context
	var companyIDs []string
	if errors.Is(err, world.ErrEntityNotFound) {
		reader, ok := service.worlds.(generatedWorldReader)
		if !ok {
			return nil, err
		}
		generated, loadErr := reader.WorldForCases(ctx, serverID)
		if loadErr != nil {
			return nil, fmt.Errorf("reconstruct email subject from generated world: %w", loadErr)
		}
		mailboxContext, ok = mailContextFromWorld(generated, npcID)
		if !ok {
			return nil, err
		}
	} else if err != nil {
		return nil, err
	} else {
		mailboxContext = Context{NPCName: profile.NPC.Name}
		mailboxContext.DeathAt = npcDeathDeadline(profile.NPC)
		if len(profile.Employment) > 0 {
			mailboxContext.Employer = profile.Employment[0].Company.Name
			mailboxContext.EmploymentRole = profile.Employment[0].Role
			companyIDs = make([]string, 0, len(profile.Employment))
			for _, job := range profile.Employment {
				companyIDs = append(companyIDs, job.Company.ID)
			}
		}
		for _, relation := range profile.PersonalRelationships {
			mailboxContext.Personal = append(mailboxContext.Personal, Correspondent{
				ID: relation.NPC.ID, Name: relation.NPC.Name, Status: relation.NPC.Status,
				DeathAt: npcDeathDeadline(relation.NPC), RelationshipType: relation.RelationshipType,
			})
		}
	}
	type linkedResult struct {
		messages []Message
		err      error
	}
	if len(companyIDs) > 0 {
		if reader, ok := service.worlds.(batchEmployeesReader); ok {
			employeesByCompany, err := reader.EmployeesOfCompanies(ctx, serverID, seed, version, companyIDs)
			if err != nil {
				return nil, err
			}
			for _, job := range profile.Employment {
				for _, employee := range employeesByCompany[job.Company.ID] {
					if employee.ID == npcID {
						continue
					}
					mailboxContext.Colleagues = append(mailboxContext.Colleagues, Correspondent{
						ID: employee.ID, Name: employee.Name, Status: employee.Status,
						DeathAt: employeeDeathDeadline(employee), Role: employee.Role,
						CompanyID: job.Company.ID,
					})
				}
			}
		} else {
			for _, job := range profile.Employment {
				employees, err := service.worlds.EmployeesOfCompany(ctx, serverID, job.Company.ID)
				if err != nil {
					return nil, err
				}
				for _, employee := range employees {
					if employee.ID == npcID {
						continue
					}
					mailboxContext.Colleagues = append(mailboxContext.Colleagues, Correspondent{
						ID: employee.ID, Name: employee.Name, Status: employee.Status,
						DeathAt: employeeDeathDeadline(employee), Role: employee.Role,
						CompanyID: job.Company.ID,
					})
				}
			}
		}
	}
	linkedCtx, cancelLinked := context.WithCancel(ctx)
	defer cancelLinked()
	linkedCh := make(chan linkedResult, 1)
	go func() {
		messages, err := service.cases.NPCEmails(linkedCtx, serverID, npcID, from, through)
		linkedCh <- linkedResult{messages: messages, err: err}
	}()
	ordinary, err := Generate(seed, version, npcID, mailboxContext, from, through)
	if err != nil {
		return nil, err
	}
	linked := <-linkedCh
	if linked.err != nil {
		return nil, linked.err
	}
	all := append(ordinary, linked.messages...)
	if !mailboxContext.DeathAt.IsZero() {
		available := all[:0]
		for _, message := range all {
			if beforeDeathDeadline(message.SentAt, mailboxContext.DeathAt) {
				available = append(available, message)
			}
		}
		all = available
	}
	SortMessages(all)
	return all, nil
}

func npcDeathDeadline(npc world.NPC) time.Time {
	if npc.Status != "dead" {
		return time.Time{}
	}
	return npc.StatusChangedAt
}

func employeeDeathDeadline(employee world.CompanyEmployee) time.Time {
	if employee.Status != "dead" {
		return time.Time{}
	}
	return employee.StatusChangedAt
}

func mailContextFromWorld(generated world.World, npcID string) (Context, bool) {
	npcs := make(map[string]world.NPC, len(generated.NPCs))
	companies := make(map[string]world.Company, len(generated.Companies))
	for _, npc := range generated.NPCs {
		npcs[npc.ID] = npc
	}
	for _, company := range generated.Companies {
		companies[company.ID] = company
	}
	person, ok := npcs[npcID]
	if !ok {
		return Context{}, false
	}
	result := Context{NPCName: person.Name}
	employerIDs := make(map[string]bool)
	for _, job := range generated.EmploymentRelationships {
		if job.EmployeeNPCID == npcID {
			employerIDs[job.EmployerCompanyID] = true
			if company, exists := companies[job.EmployerCompanyID]; exists && result.Employer == "" {
				result.Employer = company.Name
				result.EmploymentRole = job.Role
			}
		}
	}
	for _, job := range generated.EmploymentRelationships {
		if job.EmployeeNPCID == npcID || !employerIDs[job.EmployerCompanyID] {
			continue
		}
		if colleague, exists := npcs[job.EmployeeNPCID]; exists {
			result.Colleagues = append(result.Colleagues, Correspondent{ID: colleague.ID, Name: colleague.Name, Status: colleague.Status, Role: job.Role, CompanyID: job.EmployerCompanyID})
		}
	}
	for _, relation := range generated.PersonalRelationships {
		otherID := relation.PersonANPCID
		if otherID == npcID {
			otherID = relation.PersonBNPCID
		} else if relation.PersonBNPCID != npcID {
			continue
		}
		if other, exists := npcs[otherID]; exists {
			result.Personal = append(result.Personal, Correspondent{ID: other.ID, Name: other.Name, Status: other.Status, RelationshipType: relation.RelationshipType})
		}
	}
	return result, true
}

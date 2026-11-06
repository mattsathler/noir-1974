package cases

import (
	"embed"
	"encoding/json"
	"fmt"
	"strings"
	"sync"
	"time"

	"noir-1974/apps/api/internal/mail"
	"noir-1974/apps/api/internal/world"
)

//go:embed case_mail_templates.json
var caseMailTemplateFS embed.FS

type caseMailTemplate struct {
	Subject string `json:"subject"`
	Body    string `json:"body"`
}

type caseMailCatalog struct {
	Ordinary   []caseMailTemplate `json:"ordinary"`
	AfterEvent []caseMailTemplate `json:"after_event"`
	Remorse    []caseMailTemplate `json:"remorse"`
}

var (
	caseMailCatalogOnce sync.Once
	caseMailCatalogData caseMailCatalog
	caseMailCatalogErr  error
)

func loadCaseMailCatalog() (caseMailCatalog, error) {
	caseMailCatalogOnce.Do(func() {
		data, err := caseMailTemplateFS.ReadFile("case_mail_templates.json")
		if err != nil {
			caseMailCatalogErr = err
			return
		}
		if err := json.Unmarshal(data, &caseMailCatalogData); err != nil {
			caseMailCatalogErr = err
			return
		}
		for _, group := range []struct {
			name      string
			templates []caseMailTemplate
		}{
			{name: "ordinary", templates: caseMailCatalogData.Ordinary},
			{name: "after_event", templates: caseMailCatalogData.AfterEvent},
			{name: "remorse", templates: caseMailCatalogData.Remorse},
		} {
			if len(group.templates) != 200 {
				caseMailCatalogErr = fmt.Errorf("case mail template group %s has %d entries; want 200", group.name, len(group.templates))
				return
			}
			for _, item := range group.templates {
				if strings.TrimSpace(item.Subject) == "" || strings.TrimSpace(item.Body) == "" {
					caseMailCatalogErr = fmt.Errorf("case mail template in group %s is incomplete", group.name)
					return
				}
			}
		}
	})
	return caseMailCatalogData, caseMailCatalogErr
}

func buildQuietCaseEmails(seed string, worldVersion int, caseID string, selected candidate, base world.World, crimeAt time.Time) ([]mail.Message, error) {
	templates, err := loadCaseMailCatalog()
	if err != nil {
		return nil, fmt.Errorf("load case mail templates: %w", err)
	}
	killer := npcByID(base.NPCs, selected.killerID)
	victim := npcByID(base.NPCs, selected.victimID)
	if killer.ID == "" || victim.ID == "" {
		return nil, ErrInvalidDefinition
	}
	message := func(mailbox, related, sender, recipient string, item caseMailTemplate, relation string, sentAt time.Time, suffix string) mail.Message {
		result := mail.CaseMessage(seed, worldVersion, caseID, mailbox, related, sender, recipient, item.Subject, item.Body, relation, sentAt)
		result.ID += "-" + suffix
		return result
	}
	ordinaryIndex := keyIndex(seed, "case-ordinary-mail", len(templates.Ordinary))
	preCrimeOne := templates.Ordinary[ordinaryIndex]
	preCrimeTwo := templates.Ordinary[(ordinaryIndex+1+keyIndex(seed, "case-ordinary-mail-two", len(templates.Ordinary)-1))%len(templates.Ordinary)]
	preOneAt := crimeAt.AddDate(0, 0, -(20 + keyIndex(seed, "case-mail-early-days", 30))).Add(time.Duration(8*60+keyIndex(seed, "case-mail-early-minute", 8*60)) * time.Minute)
	preTwoAt := crimeAt.AddDate(0, 0, -(14 + keyIndex(seed, "case-mail-late-days", 7))).Add(time.Duration(8*60+keyIndex(seed, "case-mail-late-minute", 8*60)) * time.Minute)
	result := []mail.Message{
		message(victim.ID, killer.ID, killer.Name, victim.Name, preCrimeOne, "case_everyday", preOneAt, "early"),
		message(victim.ID, killer.ID, killer.Name, victim.Name, preCrimeTwo, "case_everyday", preTwoAt, "later"),
	}
	if correspondent := killerCorrespondent(base, selected); correspondent.ID != "" {
		for index := 0; index < 2; index++ {
			item := templates.AfterEvent[(keyIndex(seed, fmt.Sprintf("case-after-event-%d", index), len(templates.AfterEvent)))]
			item.Body = strings.ReplaceAll(item.Body, "{{victim}}", victim.Name)
			daysAfter := 14 + keyIndex(seed, fmt.Sprintf("case-after-event-days-%d", index), 14)
			if index == 1 {
				daysAfter = 28 + keyIndex(seed, "case-after-event-days-second", 28)
			}
			sentAt := crimeAt.AddDate(0, 0, daysAfter).Add(time.Duration(9*60+keyIndex(seed, fmt.Sprintf("case-after-event-minute-%d", index), 7*60)) * time.Minute)
			result = append(result, message(killer.ID, correspondent.ID, correspondent.Name, killer.Name, item, "case_after_event", sentAt, fmt.Sprintf("after-%02d", index+1)))
		}
		if keyIndex(seed, "killer-remorse", 25) == 0 {
			item := templates.Remorse[keyIndex(seed, "killer-remorse-template", len(templates.Remorse))]
			item.Body = strings.ReplaceAll(item.Body, "{{victim}}", victim.Name)
			daysAfter := 35 + keyIndex(seed, "killer-remorse-days", 56)
			sentAt := crimeAt.AddDate(0, 0, daysAfter).Add(time.Duration(19*60+keyIndex(seed, "killer-remorse-minute", 3*60)) * time.Minute)
			result = append(result, message(killer.ID, correspondent.ID, killer.Name, correspondent.Name, item, "killer_remorse", sentAt, "remorse"))
		}
	}
	mail.SortMessages(result)
	return result, nil
}

func validateQuietCaseEmails(definition Definition, base world.World) bool {
	if definition.GeneratorVersion < GeneratorVersion || definition.CrimeAt.IsZero() {
		return false
	}
	killer := npcByID(base.NPCs, definition.KillerID)
	victim := npcByID(base.NPCs, definition.VictimID)
	if killer.ID == "" || victim.ID == "" {
		return false
	}
	everyday, afterEvent, remorse := 0, 0, 0
	for _, message := range definition.Emails {
		if message.ID == "" || message.CaseID != definition.ID || message.SentAt.IsZero() {
			return false
		}
		switch message.InternalRelation {
		case "case_everyday":
			everyday++
			if message.MailboxNPCID != victim.ID || message.RelatedNPCID != killer.ID || message.Sender != killer.Name || !message.SentAt.Before(definition.CrimeAt.AddDate(0, 0, -7)) {
				return false
			}
		case "case_after_event":
			afterEvent++
			if message.MailboxNPCID != killer.ID || message.RelatedNPCID == "" || message.Sender == "" || !message.SentAt.After(definition.CrimeAt.AddDate(0, 0, 7)) {
				return false
			}
		case "killer_remorse":
			remorse++
			if message.MailboxNPCID != killer.ID || message.RelatedNPCID == "" || message.Sender != killer.Name || !message.SentAt.After(definition.CrimeAt.AddDate(0, 0, 21)) {
				return false
			}
		default:
			return false
		}
	}
	if everyday != 2 || remorse > 1 {
		return false
	}
	if correspondent := killerCorrespondent(base, candidate{killerID: killer.ID, victimID: victim.ID, paramourID: definition.ParamourID, companyID: definition.CompanyID}); correspondent.ID != "" {
		if afterEvent != 2 {
			return false
		}
		for _, message := range definition.Emails {
			if message.InternalRelation == "case_after_event" && message.RelatedNPCID != correspondent.ID {
				return false
			}
			if message.InternalRelation == "killer_remorse" && message.RelatedNPCID != correspondent.ID {
				return false
			}
		}
	} else if afterEvent != 0 || remorse != 0 {
		return false
	}
	return true
}

func paceHomicideEvidence(seed string, selected candidate, crimeAt time.Time) candidate {
	softened := map[string][3]string{
		"inheritance": {
			"A sealed estate instruction was copied before the public notice appeared in the company register.",
			"The victim's notes refer to a recent change in who would oversee the estate.",
			"The registrar's access book records an earlier visit to the file room; the initials are difficult to read.",
		},
		"revenge": {
			"A dated complaint describes a disagreement that remained unresolved after the review.",
			"A private note asks for the matter to be left until after a later meeting.",
			"The office register has a corrected late entry from the week the complaint was reopened.",
		},
		"silence_witness": {
			"A draft statement concerns an irregular transfer, but its last page is missing.",
			"The inquiry file was checked out before the scheduled review; the return box is blank.",
			"A copied ledger page contains figures that do not match the filed summary.",
		},
		"conceal_fraud": {
			"The retained ledger copy differs from the figures later submitted for review.",
			"A correction slip carries initials also found on other routine authorizations.",
			"Meeting notes list a review of the discrepancy, without recording who attended.",
		},
		"power_struggle": {
			"A succession memorandum lists the next office holder, though the copy is marked provisional.",
			"An agenda margin contains an objection to the proposed appointment.",
			"A staff access entry places someone in the records room before the appointment vote.",
		},
		"mistaken_identity": {
			"A note describes a person by coat and shift time, without giving a name.",
			"A witness recalls the victim wearing a borrowed coat that evening.",
			"The shift roster shows two employees exchanged hours that week.",
		},
		"blackmail": {
			"A draft demand requests payment before a disclosure date; no recipient is named.",
			"An envelope from the same period bears an office mark, but no readable sender.",
			"The victim's appointment book records a meeting shortly before the stated date.",
		},
		"debt": {
			"A signed note records a private loan that remained unpaid beyond its original date.",
			"A payroll advance request mentions an urgent personal obligation.",
			"A witness recalls a conversation about repayment but could not hear the details.",
		},
		"protect_reputation": {
			"A sealed complaint describes an incident that was never entered in the public register.",
			"The complaint file was requested shortly before it could no longer be located.",
			"The victim's appointment notes list a newspaper meeting, without stating its subject.",
		},
		"conceal_accident": {
			"A maintenance report records an inspection before the equipment failure.",
			"The inspection log contains a later amendment in a different hand.",
			"The shift register places a staff member at the site during the repair and cleanup window.",
		},
		"jealousy": {
			"A torn exchange refers to a planned meeting but omits the other person's name.",
			"The civil record confirms a marriage between the names on the file, without further detail.",
			"A neighbor recalls the victim meeting a colleague after work on more than one evening.",
		},
		"serial_killer": {
			"An unsigned note repeats a small detail about the recipient's usual journey home.",
			"Someone close to the victim recalls that a few letters arrived over several weeks.",
			"The workplace and appearance details in the papers could fit more than one employee.",
		},
	}
	if selected.caseType != "" {
		return selected
	}
	if replacement, ok := softened[selected.archetypeID]; ok {
		for index := range selected.evidence {
			selected.evidence[index].content = replacement[index%len(replacement)]
		}
	}
	for index := range selected.evidence {
		daysBefore := 21 + keyIndex(seed, fmt.Sprintf("homicide-evidence-date/%s/%d", selected.archetypeID, index), 55)
		filedAt := crimeAt.AddDate(0, 0, -daysBefore)
		selected.evidence[index].content = fmt.Sprintf("Filed %s. %s", filedAt.Format("2 January 2006"), selected.evidence[index].content)
	}
	return selected
}

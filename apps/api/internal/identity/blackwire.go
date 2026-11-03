package identity

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"strings"
	"time"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgconn"
	"noir-1974/apps/api/internal/market"
	"noir-1974/apps/api/internal/world"
)

var errBlackwireUnavailable = errors.New("blackwire information is unavailable")
var errBlackwireBanned = errors.New("blackwire identity has been burned")
var errBlackwireIdentityNotBanned = errors.New("blackwire identity is not eligible for replacement")
var errBlackwireProfileActive = errors.New("blackwire profile is already registered")
var errBlackwireUsernameTaken = errors.New("blackwire username is already in use")
var errBlackwireProfileRequired = errors.New("a Blackwire profile is required")
var errBlackwirePackageUnavailable = errors.New("Blackwire package is unavailable")
var errBlackwirePackageNotFetched = errors.New("Blackwire package has not been fetched")
var errBlackwireAvatarUnavailable = errors.New("Blackwire portrait is unavailable")
var errBlackwireAppPreinstalled = errors.New("preinstalled Blackwire application cannot be removed")
var errBlackwireAppNotInstalled = errors.New("Blackwire application is not installed")

type blackwireOffer struct {
	BoardID    string `json:"boardId"`
	BoardTitle string `json:"boardTitle"`
	ItemID     string `json:"itemId"`
	Source     string `json:"-"`
	Title      string `json:"title"`
	Content    string `json:"content"`
	PriceCents int64  `json:"-"`
	Category   string `json:"-"`
	CompanyID  string `json:"-"`
}

type blackwireMarketplace struct {
	Offers       []blackwireOffer           `json:"offers"`
	Account      blackwireAccount           `json:"account"`
	Applications []blackwireApplicationInfo `json:"applications"`
}

type blackwireAccount struct {
	Handle                string   `json:"handle"`
	Username              string   `json:"username"`
	AvatarID              string   `json:"avatarId"`
	ProfileComplete       bool     `json:"profileComplete"`
	CryptoCents           int64    `json:"cryptoCents"`
	Reputation            int64    `json:"reputation"`
	Level                 int      `json:"level"`
	Banned                bool     `json:"banned"`
	NextLevelReputation   int64    `json:"nextLevelReputation"`
	UnlockedApplications  []string `json:"unlockedApplications"`
	InstalledApplications []string `json:"installedApplications"`
	FetchedApplications   []string `json:"fetchedApplications"`
}

type blackwireApplicationInfo struct {
	ID           string   `json:"id"`
	Executable   string   `json:"executable"`
	Summary      string   `json:"summary"`
	Description  string   `json:"description"`
	Version      string   `json:"version"`
	Creator      string   `json:"creator"`
	Origin       string   `json:"origin"`
	Category     string   `json:"category"`
	PackageURL   string   `json:"packageUrl,omitempty"`
	Available    bool     `json:"available"`
	Preinstalled bool     `json:"preinstalled"`
	Installed    bool     `json:"installed"`
	Fetched      bool     `json:"fetched"`
	Commands     []string `json:"commands"`
}

type blackwireStore interface {
	blackwireOffers(context.Context, string, string) (blackwireMarketplace, error)
	blackwireProfile(context.Context, string, string) (blackwireAccount, error)
	sellBlackwireItem(context.Context, string, string, string, string) (blackwireOffer, blackwireAccount, error)
}

type blackwireIdentityRegistrar interface {
	registerNewBlackwireIdentity(context.Context, string, string) (blackwireAccount, error)
}

type blackwireProfileWriter interface {
	registerBlackwireProfile(context.Context, string, string, string, string) (blackwireAccount, error)
}

type blackwireApplicationStore interface {
	fetchBlackwireApp(context.Context, string, string, string) (blackwireAccount, error)
	installBlackwireApp(context.Context, string, string, string) (blackwireAccount, error)
}

type blackwireApplicationUninstaller interface {
	uninstallBlackwireApp(context.Context, string, string, string) (blackwireAccount, error)
}

var blackwireLevelThresholds = []int64{0, 50, 150, 400, 900}

func blackwireApplications() []string {
	return []string{"ghostkey", "email-breaker", "bankrupt", "blackwire", "deadletter"}
}

func blackwireApplicationCatalog(account blackwireAccount) []blackwireApplicationInfo {
	installed := make(map[string]bool, len(account.InstalledApplications))
	fetched := make(map[string]bool, len(account.FetchedApplications))
	for _, appID := range account.InstalledApplications {
		installed[appID] = true
	}
	for _, appID := range account.FetchedApplications {
		fetched[appID] = true
	}
	return []blackwireApplicationInfo{
		{ID: "ghostkey", Executable: "GHOSTKEY.EXE", Summary: "Map network directories and exposed services.", Description: "A field utility for tracing public directories and finding compromised company services. Results depend on the route and protections in place.", Version: "4.7.2", Creator: "The Crow", Origin: "Unlisted relay · origin unverified", Category: "NETWORK RECONNAISSANCE", Available: true, Preinstalled: true, Installed: true, Commands: []string{"disable protection", "run ghostkey.exe"}},
		{ID: "email-breaker", Executable: "MAILBREAKER.EXE", Summary: "Force access to a published staff mailbox.", Description: "Tests a published company mailbox against the relay's breach routines. The recipient's company email address is required.", Version: "2.3.0", Creator: "Bai He · 白鹤", Origin: "Guangzhou relay · China (unverified)", Category: "MAIL ACCESS", Available: true, Preinstalled: true, Installed: true, Commands: []string{"disable protection", "run mailbreaker.exe"}},
		{ID: "bankrupt", Executable: "BANKRUPT.EXE", Summary: "Force access to a staff bank account ledger.", Description: "A targeted financial account breach utility. A staff member's published eight-digit account number is required.", Version: "1.6.1", Creator: "Moroz · Мороз", Origin: "Novosibirsk relay · Russia (unverified)", Category: "FINANCIAL ACCESS", Available: true, Preinstalled: true, Installed: true, Commands: []string{"disable protection", "run bankrupt.exe"}},
		{ID: "blackwire", Executable: "BLACKWIRE.EXE", Summary: "Open the Blackwire relay in the web browser.", Description: "The trusted relay client for opening this catalogue and information exchange in the Meridian web browser.", Version: "1.0.0", Creator: "The Crow", Origin: "Rotating relay · origin concealed", Category: "RELAY CLIENT", Available: true, Preinstalled: true, Installed: true, Commands: []string{"run blackwire.exe"}},
		{ID: "deadletter", Executable: "DEADLETTER.EXE", Summary: "Read a small offline archive left on the Blackwire relay.", Description: "A quiet, read-only archive reader. It stores a few fragments locally and does not connect to external services.", Version: "1.0.0", Creator: "Paper Kite · 纸鸢", Origin: "Shanghai relay · China (unverified)", Category: "OFFLINE READER", PackageURL: "merid://blackwire/packages/deadletter-1.pkg", Available: true, Installed: installed["deadletter"], Fetched: fetched["deadletter"], Commands: []string{"disable protection", "fetch merid://blackwire/packages/deadletter-1.pkg", "install merid://blackwire/packages/deadletter-1.pkg", "run deadletter.exe"}},
	}
}

func blackwireReputationFor(category string) int64 {
	switch category {
	case "CASE INFORMATION":
		return 10
	case "CORPORATE LEADERSHIP":
		return 25
	case "FINANCIAL / CORPORATE EVIDENCE":
		return 50
	case "POLITICAL INFORMATION":
		return 60
	case "EXECUTIVE FINANCIAL EVIDENCE":
		return 75
	case "POLITICAL / EXECUTIVE EVIDENCE":
		return 100
	default:
		return 1
	}
}

func ensureBlackwireAccount(ctx context.Context, db querier, accountID, serverID string) error {
	tag, err := db.Exec(ctx, `INSERT INTO player_economy(server_id,account_id)
		SELECT $1::uuid,account_id FROM player_characters WHERE account_id=$2::uuid
		ON CONFLICT(server_id,account_id) DO NOTHING`, serverID, accountID)
	if err != nil {
		return err
	}
	if tag.RowsAffected() == 0 {
		var exists bool
		if err := db.QueryRow(ctx, `SELECT EXISTS(SELECT 1 FROM player_economy WHERE server_id=$1::uuid AND account_id=$2::uuid)`, serverID, accountID).Scan(&exists); err != nil {
			return err
		}
		if !exists {
			return ErrNoCharacter
		}
	}
	_, err = db.Exec(ctx, `INSERT INTO player_blackwire_accounts(server_id,account_id,handle,avatar_id,profile_complete)
		VALUES($1::uuid,$2::uuid,'BW-'||upper(substr(md5($1::text||':'||$2::text),1,8)),'portrait-01',false)
		ON CONFLICT(server_id,account_id) DO NOTHING`, serverID, accountID)
	return err
}

func readBlackwireAccount(ctx context.Context, db querier, accountID, serverID string) (blackwireAccount, error) {
	var result blackwireAccount
	err := db.QueryRow(ctx, `SELECT handle,avatar_id,profile_complete,crypto_cents,reputation,level,banned_at IS NOT NULL,
		COALESCE((SELECT array_agg(app_id ORDER BY app_id) FROM player_blackwire_installed_apps installed
			WHERE installed.server_id=player_blackwire_accounts.server_id AND installed.account_id=player_blackwire_accounts.account_id),ARRAY[]::text[]),
		COALESCE((SELECT array_agg(app_id ORDER BY app_id) FROM player_blackwire_downloads downloaded
			WHERE downloaded.server_id=player_blackwire_accounts.server_id AND downloaded.account_id=player_blackwire_accounts.account_id),ARRAY[]::text[])
		FROM player_blackwire_accounts WHERE server_id=$1::uuid AND account_id=$2::uuid`, serverID, accountID).Scan(&result.Handle, &result.AvatarID, &result.ProfileComplete, &result.CryptoCents, &result.Reputation, &result.Level, &result.Banned, &result.InstalledApplications, &result.FetchedApplications)
	if err != nil {
		return blackwireAccount{}, err
	}
	result.Username = result.Handle
	for _, threshold := range blackwireLevelThresholds {
		if result.Reputation < threshold {
			result.NextLevelReputation = threshold
			break
		}
	}
	result.UnlockedApplications = blackwireApplications()
	result.InstalledApplications = append([]string{"ghostkey", "email-breaker", "bankrupt", "blackwire"}, result.InstalledApplications...)
	return result, nil
}

func (s *postgresStore) blackwireProfile(ctx context.Context, accountID, serverID string) (blackwireAccount, error) {
	if err := ensureBlackwireAccount(ctx, s.db, accountID, serverID); err != nil {
		return blackwireAccount{}, err
	}
	return readBlackwireAccount(ctx, s.db, accountID, serverID)
}

func blackwireAvatarExists(avatarID string) bool {
	switch avatarID {
	case "portrait-01", "portrait-02", "portrait-03", "portrait-04", "portrait-05", "portrait-06":
		return true
	default:
		return false
	}
}

func (s *postgresStore) registerBlackwireProfile(ctx context.Context, accountID, serverID, username, avatarID string) (blackwireAccount, error) {
	if !blackwireAvatarExists(avatarID) {
		return blackwireAccount{}, errBlackwireAvatarUnavailable
	}
	transactional, ok := s.db.(interface {
		Begin(context.Context) (pgx.Tx, error)
	})
	if !ok {
		return blackwireAccount{}, errors.New("Blackwire profile transactions are unavailable")
	}
	tx, err := transactional.Begin(ctx)
	if err != nil {
		return blackwireAccount{}, err
	}
	defer tx.Rollback(ctx)
	if err := ensureBlackwireAccount(ctx, tx, accountID, serverID); err != nil {
		return blackwireAccount{}, err
	}
	var complete, banned bool
	if err := tx.QueryRow(ctx, `SELECT profile_complete,banned_at IS NOT NULL FROM player_blackwire_accounts
		WHERE server_id=$1::uuid AND account_id=$2::uuid FOR UPDATE`, serverID, accountID).Scan(&complete, &banned); err != nil {
		return blackwireAccount{}, err
	}
	if banned {
		return blackwireAccount{}, errBlackwireBanned
	}
	if complete {
		return blackwireAccount{}, errBlackwireProfileActive
	}
	var usernameExists bool
	if err := tx.QueryRow(ctx, `SELECT EXISTS(SELECT 1 FROM player_blackwire_accounts
		WHERE server_id=$1::uuid AND lower(handle)=lower($2))`, serverID, username).Scan(&usernameExists); err != nil {
		return blackwireAccount{}, err
	}
	if usernameExists {
		return blackwireAccount{}, errBlackwireUsernameTaken
	}
	if _, err := tx.Exec(ctx, `UPDATE player_blackwire_accounts SET handle=$3,avatar_id=$4,profile_complete=true,updated_at=now()
		WHERE server_id=$1::uuid AND account_id=$2::uuid`, serverID, accountID, username, avatarID); err != nil {
		var postgresError *pgconn.PgError
		if errors.As(err, &postgresError) && postgresError.Code == "23505" {
			return blackwireAccount{}, errBlackwireUsernameTaken
		}
		return blackwireAccount{}, err
	}
	profile, err := readBlackwireAccount(ctx, tx, accountID, serverID)
	if err != nil {
		return blackwireAccount{}, err
	}
	if err := tx.Commit(ctx); err != nil {
		return blackwireAccount{}, err
	}
	return profile, nil
}

func (s *postgresStore) blackwireAppAccountState(ctx context.Context, accountID, serverID string) (bool, bool, error) {
	if err := ensureBlackwireAccount(ctx, s.db, accountID, serverID); err != nil {
		return false, false, err
	}
	var complete, banned bool
	err := s.db.QueryRow(ctx, `SELECT profile_complete,banned_at IS NOT NULL FROM player_blackwire_accounts
		WHERE server_id=$1::uuid AND account_id=$2::uuid`, serverID, accountID).Scan(&complete, &banned)
	return complete, banned, err
}

func (s *postgresStore) fetchBlackwireApp(ctx context.Context, accountID, serverID, appID string) (blackwireAccount, error) {
	if appID != "deadletter" {
		return blackwireAccount{}, errBlackwirePackageUnavailable
	}
	complete, banned, err := s.blackwireAppAccountState(ctx, accountID, serverID)
	if err != nil {
		return blackwireAccount{}, err
	}
	if banned {
		return blackwireAccount{}, errBlackwireBanned
	}
	if !complete {
		return blackwireAccount{}, errBlackwireProfileRequired
	}
	_, err = s.db.Exec(ctx, `INSERT INTO player_blackwire_downloads(server_id,account_id,app_id)
		VALUES($1::uuid,$2::uuid,$3) ON CONFLICT(server_id,account_id,app_id) DO UPDATE SET fetched_at=now()`, serverID, accountID, appID)
	if err != nil {
		return blackwireAccount{}, err
	}
	profile, err := readBlackwireAccount(ctx, s.db, accountID, serverID)
	return profile, err
}

func (s *postgresStore) installBlackwireApp(ctx context.Context, accountID, serverID, appID string) (blackwireAccount, error) {
	if appID != "deadletter" {
		return blackwireAccount{}, errBlackwirePackageUnavailable
	}
	transactional, ok := s.db.(interface {
		Begin(context.Context) (pgx.Tx, error)
	})
	if !ok {
		return blackwireAccount{}, errors.New("Blackwire installation transactions are unavailable")
	}
	tx, err := transactional.Begin(ctx)
	if err != nil {
		return blackwireAccount{}, err
	}
	defer tx.Rollback(ctx)
	if err := ensureBlackwireAccount(ctx, tx, accountID, serverID); err != nil {
		return blackwireAccount{}, err
	}
	var complete, banned bool
	if err := tx.QueryRow(ctx, `SELECT profile_complete,banned_at IS NOT NULL FROM player_blackwire_accounts
		WHERE server_id=$1::uuid AND account_id=$2::uuid FOR UPDATE`, serverID, accountID).Scan(&complete, &banned); err != nil {
		return blackwireAccount{}, err
	}
	if banned {
		return blackwireAccount{}, errBlackwireBanned
	}
	if !complete {
		return blackwireAccount{}, errBlackwireProfileRequired
	}
	var fetched bool
	if err := tx.QueryRow(ctx, `SELECT EXISTS(SELECT 1 FROM player_blackwire_downloads WHERE server_id=$1::uuid AND account_id=$2::uuid AND app_id=$3)`, serverID, accountID, appID).Scan(&fetched); err != nil {
		return blackwireAccount{}, err
	}
	if !fetched {
		return blackwireAccount{}, errBlackwirePackageNotFetched
	}
	if _, err := tx.Exec(ctx, `INSERT INTO player_blackwire_installed_apps(server_id,account_id,app_id)
		VALUES($1::uuid,$2::uuid,$3) ON CONFLICT(server_id,account_id,app_id) DO NOTHING`, serverID, accountID, appID); err != nil {
		return blackwireAccount{}, err
	}
	if _, err := tx.Exec(ctx, `DELETE FROM player_blackwire_downloads WHERE server_id=$1::uuid AND account_id=$2::uuid AND app_id=$3`, serverID, accountID, appID); err != nil {
		return blackwireAccount{}, err
	}
	profile, err := readBlackwireAccount(ctx, tx, accountID, serverID)
	if err != nil {
		return blackwireAccount{}, err
	}
	if err := tx.Commit(ctx); err != nil {
		return blackwireAccount{}, err
	}
	return profile, nil
}

func (s *postgresStore) uninstallBlackwireApp(ctx context.Context, accountID, serverID, appID string) (blackwireAccount, error) {
	if appID != "deadletter" {
		if appID == "ghostkey" || appID == "email-breaker" || appID == "bankrupt" || appID == "blackwire" {
			return blackwireAccount{}, errBlackwireAppPreinstalled
		}
		return blackwireAccount{}, errBlackwirePackageUnavailable
	}
	transactional, ok := s.db.(interface {
		Begin(context.Context) (pgx.Tx, error)
	})
	if !ok {
		return blackwireAccount{}, errors.New("Blackwire uninstall transactions are unavailable")
	}
	tx, err := transactional.Begin(ctx)
	if err != nil {
		return blackwireAccount{}, err
	}
	defer tx.Rollback(ctx)
	if err := ensureBlackwireAccount(ctx, tx, accountID, serverID); err != nil {
		return blackwireAccount{}, err
	}
	var complete, banned bool
	if err := tx.QueryRow(ctx, `SELECT profile_complete,banned_at IS NOT NULL FROM player_blackwire_accounts
		WHERE server_id=$1::uuid AND account_id=$2::uuid FOR UPDATE`, serverID, accountID).Scan(&complete, &banned); err != nil {
		return blackwireAccount{}, err
	}
	if banned {
		return blackwireAccount{}, errBlackwireBanned
	}
	if !complete {
		return blackwireAccount{}, errBlackwireProfileRequired
	}
	result, err := tx.Exec(ctx, `DELETE FROM player_blackwire_installed_apps
		WHERE server_id=$1::uuid AND account_id=$2::uuid AND app_id=$3`, serverID, accountID, appID)
	if err != nil {
		return blackwireAccount{}, err
	}
	if result.RowsAffected() == 0 {
		return blackwireAccount{}, errBlackwireAppNotInstalled
	}
	profile, err := readBlackwireAccount(ctx, tx, accountID, serverID)
	if err != nil {
		return blackwireAccount{}, err
	}
	if err := tx.Commit(ctx); err != nil {
		return blackwireAccount{}, err
	}
	return profile, nil
}

func (s *postgresStore) registerNewBlackwireIdentity(ctx context.Context, accountID, serverID string) (blackwireAccount, error) {
	transactional, ok := s.db.(interface {
		Begin(context.Context) (pgx.Tx, error)
	})
	if !ok {
		return blackwireAccount{}, errors.New("blackwire identity transactions are unavailable")
	}
	tx, err := transactional.Begin(ctx)
	if err != nil {
		return blackwireAccount{}, err
	}
	defer tx.Rollback(ctx)
	if err := ensureBlackwireAccount(ctx, tx, accountID, serverID); err != nil {
		return blackwireAccount{}, err
	}
	result, err := tx.Exec(ctx, `UPDATE player_blackwire_accounts
		SET identity_generation=identity_generation+1,
			handle='BW-'||upper(substr(md5(server_id::text||':'||account_id::text||':'||(identity_generation+1)::text),1,8)),
			avatar_id='portrait-01',profile_complete=false,crypto_cents=0,reputation=0,level=1,banned_at=NULL,updated_at=now()
		WHERE server_id=$1::uuid AND account_id=$2::uuid AND banned_at IS NOT NULL`, serverID, accountID)
	if err != nil {
		return blackwireAccount{}, err
	}
	if result.RowsAffected() == 0 {
		return blackwireAccount{}, errBlackwireIdentityNotBanned
	}
	if _, err := tx.Exec(ctx, `UPDATE player_hacker_exposure SET exposure=0,updated_at=now()
		WHERE server_id=$1::uuid AND account_id=$2::uuid`, serverID, accountID); err != nil {
		return blackwireAccount{}, err
	}
	profile, err := readBlackwireAccount(ctx, tx, accountID, serverID)
	if err != nil {
		return blackwireAccount{}, err
	}
	if err := tx.Commit(ctx); err != nil {
		return blackwireAccount{}, err
	}
	return profile, nil
}

func classifyBlackwireOffer(item map[string]any) (int64, string, string) {
	title, _ := item["title"].(string)
	content, _ := item["content"].(string)
	source, _ := item["source"].(string)
	metadata, _ := item["metadata"].(map[string]any)
	text := strings.ToLower(strings.Join([]string{title, content, source}, " "))
	companyID := ""
	explicitCompanyID := false
	if metadata != nil {
		if value, ok := metadata["companyId"].(string); ok && value != "" {
			companyID, explicitCompanyID = value, true
		}
		if companyID == "" {
			companyID, _ = metadata["employerCompanyId"].(string)
		}
		if company, ok := metadata["company"].(map[string]any); ok {
			if value, ok := company["id"].(string); ok && companyID == "" {
				companyID, explicitCompanyID = value, true
			}
		}
	}
	category, price := "ROUTINE INFORMATION", int64(25)
	if metadata != nil {
		for _, key := range []string{"caseId", "crimeType", "crimeId"} {
			if value, ok := metadata[key].(string); ok && value != "" {
				category, price = "CASE INFORMATION", 2500
				break
			}
		}
	}
	financial := strings.Contains(text, "fraud") || strings.Contains(text, "extortion") || strings.Contains(text, "embezzle") || strings.Contains(text, "payroll fraud") || strings.Contains(text, "false account") || strings.Contains(text, "accounting manipulation") || strings.Contains(text, "misappropriat") || strings.Contains(text, "kickback") || strings.Contains(text, "bribery") || strings.Contains(text, "concealed payment")
	executive := strings.Contains(text, "director") || strings.Contains(text, "chief executive") || strings.Contains(text, "president") || strings.Contains(text, "board member") || strings.Contains(text, "managing partner")
	political := false
	for _, term := range []string{"minister", "parliament", "state contract", "government tender", "public tender", "cabinet", "state secretary", "national commission", "campaign fund", "political contribution", "government procurement", "bribery"} {
		if strings.Contains(text, term) {
			political = true
			break
		}
	}
	if metadata != nil {
		political = political || metadata["politicalSensitivity"] == true
	}
	switch {
	case political && (executive || financial):
		category, price = "POLITICAL / EXECUTIVE EVIDENCE", 125000
	case political:
		category, price = "POLITICAL INFORMATION", 50000
	case financial && executive:
		category, price = "EXECUTIVE FINANCIAL EVIDENCE", 75000
	case financial:
		category, price = "FINANCIAL / CORPORATE EVIDENCE", 35000
	case executive && companyID != "":
		category, price = "CORPORATE LEADERSHIP", 20000
	case category == "CASE INFORMATION":
		price = 5000
	}
	if !explicitCompanyID && !financial && !executive {
		companyID = ""
	}
	return price, category, companyID
}

func blackwirePriceForMarketChange(priceCents, changeBps int64) int64 {
	const maxAdjustmentBps int64 = 5000
	changeBps = max(-maxAdjustmentBps, min(maxAdjustmentBps, changeBps))
	adjusted := (priceCents*(10000+changeBps) + 5000) / 10000
	return max(int64(1), adjusted)
}

func (s *postgresStore) blackwireOffers(ctx context.Context, accountID, serverID string) (blackwireMarketplace, error) {
	if err := ensureBlackwireAccount(ctx, s.db, accountID, serverID); err != nil {
		return blackwireMarketplace{}, err
	}
	profile, err := readBlackwireAccount(ctx, s.db, accountID, serverID)
	if err != nil {
		return blackwireMarketplace{}, err
	}
	if profile.Banned {
		return blackwireMarketplace{Offers: []blackwireOffer{}, Account: profile, Applications: blackwireApplicationCatalog(profile)}, nil
	}
	queryer, ok := s.db.(interface {
		Query(context.Context, string, ...any) (pgx.Rows, error)
	})
	if !ok {
		return blackwireMarketplace{}, errors.New("whiteboard queries are unavailable")
	}
	rows, err := queryer.Query(ctx, `SELECT board.whiteboard_id, board.title, item.item_id, item.item_data
		FROM player_whiteboards board
		JOIN player_whiteboard_pins pin USING (server_id, account_id, whiteboard_id)
		JOIN player_discovered_items item USING (server_id, account_id, item_id)
		LEFT JOIN player_blackwire_sales sale USING (server_id, account_id, item_id)
		WHERE board.server_id=$1::uuid AND board.account_id=$2::uuid AND sale.item_id IS NULL
			AND NOT EXISTS (SELECT 1 FROM player_whiteboard_pins locked_pin JOIN player_whiteboards locked_board USING(server_id,account_id,whiteboard_id)
				WHERE locked_pin.server_id=item.server_id AND locked_pin.account_id=item.account_id AND locked_pin.item_id=item.item_id AND locked_board.locked_at IS NOT NULL)
			AND NOT EXISTS (
				SELECT 1 FROM player_whiteboard_pins case_pin
				JOIN (
					SELECT server_id,case_id,evidence_id AS item_id FROM case_evidence
					UNION SELECT server_id,case_id,record_id AS item_id FROM case_activity_records
					UNION SELECT server_id,case_id,message_id AS item_id FROM case_mail_messages
				) case_item ON case_item.server_id=case_pin.server_id AND case_item.item_id=case_pin.item_id
				JOIN case_states case_state ON case_state.server_id=case_item.server_id AND case_state.case_id=case_item.case_id
				WHERE case_pin.server_id=board.server_id AND case_pin.account_id=board.account_id
					AND case_pin.whiteboard_id=board.whiteboard_id AND case_state.status<>'open'
			)
		ORDER BY board.created_at, board.whiteboard_id, pin.pinned_at, item.item_id`, serverID, accountID)
	if err != nil {
		return blackwireMarketplace{}, err
	}
	defer rows.Close()
	result := blackwireMarketplace{Offers: []blackwireOffer{}, Applications: blackwireApplicationCatalog(profile)}
	for rows.Next() {
		var boardID, boardTitle, itemID string
		var raw []byte
		if err := rows.Scan(&boardID, &boardTitle, &itemID, &raw); err != nil {
			return blackwireMarketplace{}, err
		}
		var item map[string]any
		if err := json.Unmarshal(raw, &item); err != nil {
			continue
		}
		price, category, companyID := classifyBlackwireOffer(item)
		offer := blackwireOffer{BoardID: boardID, BoardTitle: boardTitle, ItemID: itemID, PriceCents: price, Category: category, CompanyID: companyID}
		offer.Source, _ = item["source"].(string)
		offer.Title, _ = item["title"].(string)
		offer.Content, _ = item["content"].(string)
		result.Offers = append(result.Offers, offer)
	}
	if err := rows.Err(); err != nil {
		return blackwireMarketplace{}, err
	}
	result.Account = profile
	result.Applications = blackwireApplicationCatalog(profile)
	return result, nil
}

func (s *postgresStore) sellBlackwireItem(ctx context.Context, accountID, serverID, boardID, itemID string) (blackwireOffer, blackwireAccount, error) {
	transactional, ok := s.db.(interface {
		Begin(context.Context) (pgx.Tx, error)
	})
	if !ok {
		return blackwireOffer{}, blackwireAccount{}, errors.New("blackwire transactions are unavailable")
	}
	tx, err := transactional.Begin(ctx)
	if err != nil {
		return blackwireOffer{}, blackwireAccount{}, err
	}
	defer tx.Rollback(ctx)
	if err := ensureBlackwireAccount(ctx, tx, accountID, serverID); err != nil {
		return blackwireOffer{}, blackwireAccount{}, err
	}
	var banned bool
	if err := tx.QueryRow(ctx, `SELECT banned_at IS NOT NULL FROM player_blackwire_accounts
		WHERE server_id=$1::uuid AND account_id=$2::uuid FOR UPDATE`, serverID, accountID).Scan(&banned); err != nil {
		return blackwireOffer{}, blackwireAccount{}, err
	}
	if banned {
		return blackwireOffer{}, blackwireAccount{}, errBlackwireBanned
	}
	var boardTitle string
	var raw []byte
	err = tx.QueryRow(ctx, `SELECT board.title, item.item_data
		FROM player_whiteboards board
		JOIN player_whiteboard_pins pin USING (server_id, account_id, whiteboard_id)
		JOIN player_discovered_items item USING (server_id, account_id, item_id)
		LEFT JOIN player_blackwire_sales sale USING (server_id, account_id, item_id)
		WHERE board.server_id=$1::uuid AND board.account_id=$2::uuid AND board.whiteboard_id=$3 AND item.item_id=$4
			AND board.locked_at IS NULL AND sale.item_id IS NULL
			AND NOT EXISTS (SELECT 1 FROM player_whiteboard_pins locked_pin JOIN player_whiteboards locked_board USING(server_id,account_id,whiteboard_id)
				WHERE locked_pin.server_id=item.server_id AND locked_pin.account_id=item.account_id AND locked_pin.item_id=item.item_id AND locked_board.locked_at IS NOT NULL)
			AND NOT EXISTS (
				SELECT 1 FROM player_whiteboard_pins case_pin
				JOIN (
					SELECT server_id,case_id,evidence_id AS item_id FROM case_evidence
					UNION SELECT server_id,case_id,record_id AS item_id FROM case_activity_records
					UNION SELECT server_id,case_id,message_id AS item_id FROM case_mail_messages
				) case_item ON case_item.server_id=case_pin.server_id AND case_item.item_id=case_pin.item_id
				JOIN case_states case_state ON case_state.server_id=case_item.server_id AND case_state.case_id=case_item.case_id
				WHERE case_pin.server_id=board.server_id AND case_pin.account_id=board.account_id
					AND case_pin.whiteboard_id=board.whiteboard_id AND case_state.status<>'open'
			)
		FOR UPDATE OF item`, serverID, accountID, boardID, itemID).Scan(&boardTitle, &raw)
	if errors.Is(err, pgx.ErrNoRows) {
		return blackwireOffer{}, blackwireAccount{}, errBlackwireUnavailable
	}
	if err != nil {
		return blackwireOffer{}, blackwireAccount{}, err
	}
	var item map[string]any
	if err := json.Unmarshal(raw, &item); err != nil {
		return blackwireOffer{}, blackwireAccount{}, fmt.Errorf("decode discovered information: %w", err)
	}
	price, category, companyID := classifyBlackwireOffer(item)
	marketDelta := blackwireMarketDelta(category)
	marketEventID := "blackwire:" + accountID + ":" + itemID
	var rivalIDs []string
	if companyID != "" {
		var seed string
		var version int
		if err := tx.QueryRow(ctx, `SELECT w.world_seed,COALESCE(city.generator_version,w.generator_version)
			FROM worlds w LEFT JOIN world_city_generators city ON city.server_id=w.server_id
			AND city.city_id=split_part(split_part($2,':',2),'/companies/',1)
			WHERE w.server_id=$1::uuid`, serverID, companyID).Scan(&seed, &version); err != nil {
			return blackwireOffer{}, blackwireAccount{}, err
		}
		marketNow := time.Now().UTC()
		if group, groupErr := world.CompanyRivalGroupFor(seed, version, companyID); groupErr == nil {
			movements := rivalMarketMovements(serverID, []world.CompanyRivalGroup{group}, marketNow)
			var eventChangeBps int64
			var sourceCompanyID string
			eventErr := tx.QueryRow(ctx, `SELECT (score_after-score_before)*10,source_company_id
				FROM company_market_events WHERE server_id=$1::uuid AND company_id=$2
				ORDER BY recorded_at DESC,event_id DESC LIMIT 1`, serverID, companyID).Scan(&eventChangeBps, &sourceCompanyID)
			if eventErr != nil && !errors.Is(eventErr, pgx.ErrNoRows) {
				return blackwireOffer{}, blackwireAccount{}, eventErr
			}
			movement := movements[companyID].forAnchor(companyID, sourceCompanyID)
			price = blackwirePriceForMarketChange(price, combineMarketChanges(eventChangeBps, movement.changeBps))
			if marketDelta != 0 {
				rivalIDs = make([]string, 0, len(group.CompanyIDs))
				for _, rivalID := range group.CompanyIDs {
					if rivalID != companyID {
						rivalIDs = append(rivalIDs, rivalID)
					}
				}
			}
		}
	}
	_, err = tx.Exec(ctx, `UPDATE player_blackwire_accounts SET crypto_cents=crypto_cents+$3,
		reputation=reputation+$4, updated_at=now() WHERE server_id=$1::uuid AND account_id=$2::uuid`,
		serverID, accountID, price, blackwireReputationFor(category))
	if err != nil {
		return blackwireOffer{}, blackwireAccount{}, err
	}
	if _, err := tx.Exec(ctx, `UPDATE player_blackwire_accounts SET level=CASE
		WHEN reputation>=900 THEN 5 WHEN reputation>=400 THEN 4 WHEN reputation>=150 THEN 3
		WHEN reputation>=50 THEN 2 ELSE 1 END
		WHERE server_id=$1::uuid AND account_id=$2::uuid`, serverID, accountID); err != nil {
		return blackwireOffer{}, blackwireAccount{}, err
	}
	_, err = tx.Exec(ctx, `INSERT INTO player_blackwire_sales(server_id,account_id,item_id,whiteboard_id,price_cents,category,company_id)
		VALUES($1::uuid,$2::uuid,$3,$4,$5,$6,NULLIF($7,''))`, serverID, accountID, itemID, boardID, price, category, companyID)
	if err != nil {
		return blackwireOffer{}, blackwireAccount{}, err
	}
	messageID := "blackwire-sale:" + itemID
	title, _ := item["title"].(string)
	body := fmt.Sprintf("The Blackwire exchange has completed the sale of %q. Crypto received: %d.%02d. The information has been withdrawn from your boards and cannot be used as case evidence.",
		title, price/100, price%100)
	if _, err := tx.Exec(ctx, `INSERT INTO player_inbox_messages(server_id,account_id,message_id,sent_at,sender,message_category,subject,body,attachments)
		VALUES($1::uuid,$2::uuid,$3,now(),'Blackwire Exchange','official','Blackwire sale · Settlement complete',$4,'[]'::jsonb)
		ON CONFLICT(server_id,account_id,message_id) DO NOTHING`, serverID, accountID, messageID, body); err != nil {
		return blackwireOffer{}, blackwireAccount{}, err
	}
	if companyID != "" && marketDelta != 0 {
		if err := market.ApplyCompanyImpact(ctx, tx, serverID, marketEventID, companyID, "blackwire_leak", marketDelta, time.Now().UTC(), rivalIDs); err != nil {
			return blackwireOffer{}, blackwireAccount{}, err
		}
	}
	if _, err := tx.Exec(ctx, `DELETE FROM player_whiteboard_pins WHERE server_id=$1::uuid AND account_id=$2::uuid AND item_id=$3`, serverID, accountID, itemID); err != nil {
		return blackwireOffer{}, blackwireAccount{}, err
	}
	if _, err := tx.Exec(ctx, `DELETE FROM player_whiteboard_connections WHERE server_id=$1::uuid AND account_id=$2::uuid AND (from_item_id=$3 OR to_item_id=$3)`, serverID, accountID, itemID); err != nil {
		return blackwireOffer{}, blackwireAccount{}, err
	}
	if _, err := tx.Exec(ctx, `UPDATE player_discovered_items SET item_data=item_data || '{"blackwireSold":true}'::jsonb WHERE server_id=$1::uuid AND account_id=$2::uuid AND item_id=$3`, serverID, accountID, itemID); err != nil {
		return blackwireOffer{}, blackwireAccount{}, err
	}
	updatedAccount, err := readBlackwireAccount(ctx, tx, accountID, serverID)
	if err != nil {
		return blackwireOffer{}, blackwireAccount{}, err
	}
	if err := tx.Commit(ctx); err != nil {
		return blackwireOffer{}, blackwireAccount{}, err
	}
	offer := blackwireOffer{BoardID: boardID, BoardTitle: boardTitle, ItemID: itemID, PriceCents: price, Category: category, CompanyID: companyID}
	offer.Source, _ = item["source"].(string)
	offer.Title, _ = item["title"].(string)
	offer.Content, _ = item["content"].(string)
	return offer, updatedAccount, nil
}

func blackwireMarketDelta(category string) int {
	switch category {
	case "POLITICAL / EXECUTIVE EVIDENCE":
		return -240
	case "POLITICAL INFORMATION":
		return -120
	case "EXECUTIVE FINANCIAL EVIDENCE":
		return -150
	case "FINANCIAL / CORPORATE EVIDENCE":
		return -100
	case "CORPORATE LEADERSHIP":
		return -60
	case "CASE INFORMATION":
		return -15
	default:
		return 0
	}
}

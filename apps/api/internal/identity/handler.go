package identity

import (
	"context"
	"crypto/rand"
	"encoding/base64"
	"encoding/json"
	"errors"
	"io"
	"log"
	"net/http"
	"net/mail"
	"regexp"
	"strconv"
	"strings"
	"time"

	"golang.org/x/crypto/bcrypt"
	"noir-1974/apps/api/internal/cases"
	correspondence "noir-1974/apps/api/internal/mail"
	"noir-1974/apps/api/internal/newspaper"
	"noir-1974/apps/api/internal/records"
	"noir-1974/apps/api/internal/world"
)

const sessionCookie = "noir_session"
const sessionDuration = 30 * 24 * time.Hour

var usernamePattern = regexp.MustCompile(`^[a-zA-Z0-9_-]{3,24}$`)
var whiteboardIDPattern = regexp.MustCompile(`^board_[A-Za-z0-9_-]+$`)

type Handler struct {
	store               store
	secureCookie        bool
	now                 func() time.Time
	worldClockStartedAt time.Time
	worldReader         WorldDirectoryReader
	newspaperReader     NewspaperReader
	caseReader          CaseAPIReader
	whiteboardReader    WhiteboardAPIReader
	recordsReader       RecordsAPIReader
	emailsReader        EmailsAPIReader
	worldServerID       string
}

type WorldDirectoryReader interface {
	SearchNPCs(context.Context, string, string) ([]world.PersonSearchRecord, error)
	SearchCompanies(context.Context, string, string) ([]world.CompanyRecord, error)
	SearchLocations(context.Context, string, string, string) ([]world.LocationRecord, error)
}

type NewspaperReader interface {
	Issue(context.Context, string, string, string, time.Time) (newspaper.Issue, error)
}

type CaseAPIReader interface {
	ListOpenPublicCases(context.Context, string) ([]cases.PublicCase, error)
	GetPublicCase(context.Context, string, string) (cases.PublicCase, []cases.PublicEvidence, error)
	GetPlayerProgress(context.Context, string, string, string) (cases.PlayerProgress, error)
	SavePlayerProgress(context.Context, string, string, string, cases.PlayerProgress) error
}

type RecordsAPIReader interface {
	NPCRecords(context.Context, string, string, time.Time, time.Time, string) ([]records.Record, error)
}

type EmailsAPIReader interface {
	NPCEmails(context.Context, string, string, time.Time, time.Time) ([]correspondence.Message, error)
}

type WhiteboardAPIReader interface {
	RecordDiscoveredItems(context.Context, string, string, []cases.DiscoveredItem) error
	ListWhiteboards(context.Context, string, string) ([]cases.Whiteboard, error)
	CreateWhiteboard(context.Context, string, string, string, string) (cases.Whiteboard, error)
	RenameWhiteboard(context.Context, string, string, string, string) error
	DeleteWhiteboard(context.Context, string, string, string) error
	GetWhiteboard(context.Context, string, string, string) (cases.Whiteboard, error)
	PinDiscoveredItem(context.Context, string, string, string, string, int, int) error
	MoveWhiteboardPin(context.Context, string, string, string, string, int, int) error
	RemoveWhiteboardPin(context.Context, string, string, string, string) error
	ConnectWhiteboardPins(context.Context, string, string, string, string, string) error
	DisconnectWhiteboardPins(context.Context, string, string, string, string, string) error
}

// NewHandler creates the identity API. Cookies are secure in production deployments.
func NewHandler(repository store, secureCookie bool) http.Handler {
	return newHandler(repository, secureCookie, nil, "", nil, nil, nil, nil)
}

// NewHandlerWithWorld composes authenticated world directory routes into the identity API.
func NewHandlerWithWorld(repository store, secureCookie bool, reader WorldDirectoryReader, serverID string) http.Handler {
	return newHandler(repository, secureCookie, reader, serverID, nil, nil, nil, nil)
}

func NewHandlerWithNewspaper(repository store, secureCookie bool, reader WorldDirectoryReader, newspaperReader NewspaperReader, serverID string) http.Handler {
	return newHandler(repository, secureCookie, reader, serverID, newspaperReader, nil, nil, nil)
}

func NewHandlerWithGame(repository store, secureCookie bool, reader WorldDirectoryReader, newspaperReader NewspaperReader, caseReader CaseAPIReader, recordsReader RecordsAPIReader, serverID string) http.Handler {
	return newHandler(repository, secureCookie, reader, serverID, newspaperReader, caseReader, recordsReader, nil)
}

func NewHandlerWithGameAndMail(repository store, secureCookie bool, reader WorldDirectoryReader, newspaperReader NewspaperReader, caseReader CaseAPIReader, recordsReader RecordsAPIReader, emailsReader EmailsAPIReader, serverID string) http.Handler {
	return newHandler(repository, secureCookie, reader, serverID, newspaperReader, caseReader, recordsReader, emailsReader)
}

func newHandler(repository store, secureCookie bool, reader WorldDirectoryReader, serverID string, newspaperReader NewspaperReader, caseReader CaseAPIReader, recordsReader RecordsAPIReader, emailsReader EmailsAPIReader) http.Handler {
	now := time.Now
	handler := &Handler{store: repository, secureCookie: secureCookie, now: now, worldClockStartedAt: now(), worldReader: reader, newspaperReader: newspaperReader, caseReader: caseReader, recordsReader: recordsReader, emailsReader: emailsReader, worldServerID: serverID}
	if whiteboardReader, ok := caseReader.(WhiteboardAPIReader); ok {
		handler.whiteboardReader = whiteboardReader
	}
	mux := http.NewServeMux()
	mux.HandleFunc("POST /api/auth/register", handler.register)
	mux.HandleFunc("POST /api/auth/login", handler.login)
	mux.HandleFunc("POST /api/auth/logout", handler.logout)
	mux.HandleFunc("GET /api/me", handler.me)
	mux.HandleFunc("GET /api/world", handler.world)
	if reader != nil {
		mux.HandleFunc("GET /api/world/npcs", handler.searchNPCs)
		mux.HandleFunc("GET /api/world/companies", handler.searchCompanies)
		mux.HandleFunc("GET /api/world/locations", handler.searchLocations)
		mux.HandleFunc("GET /api/world/transit", handler.transitDistricts)
		mux.HandleFunc("POST /api/world/travel", handler.travel)
		mux.HandleFunc("POST /api/world/visit", handler.visit)
	}
	if recordsReader != nil {
		mux.HandleFunc("GET /api/world/npcs/{npcID}/records", handler.npcRecords)
	}
	if emailsReader != nil {
		mux.HandleFunc("GET /api/world/npcs/{npcID}/emails", handler.npcEmails)
	}
	if newspaperReader != nil {
		mux.HandleFunc("GET /api/world/newspaper", handler.newspaper)
	}
	if caseReader != nil {
		mux.HandleFunc("GET /api/cases", handler.listOpenCases)
		mux.HandleFunc("GET /api/cases/{caseID}", handler.getCase)
		mux.HandleFunc("POST /api/cases/{caseID}/progress", handler.saveCaseProgress)
	}
	if handler.whiteboardReader != nil {
		mux.HandleFunc("GET /api/whiteboards", handler.listWhiteboards)
		mux.HandleFunc("POST /api/whiteboards", handler.createWhiteboard)
		mux.HandleFunc("GET /api/whiteboards/{boardID}", handler.getWhiteboard)
		mux.HandleFunc("PATCH /api/whiteboards/{boardID}", handler.renameWhiteboard)
		mux.HandleFunc("DELETE /api/whiteboards/{boardID}", handler.deleteWhiteboard)
		mux.HandleFunc("POST /api/whiteboards/{boardID}/pins", handler.pinWhiteboardItem)
		mux.HandleFunc("PATCH /api/whiteboards/{boardID}/pins/{itemID}", handler.moveWhiteboardPin)
		mux.HandleFunc("DELETE /api/whiteboards/{boardID}/pins/{itemID}", handler.removeWhiteboardPin)
		mux.HandleFunc("POST /api/whiteboards/{boardID}/connections", handler.connectWhiteboardPins)
		mux.HandleFunc("DELETE /api/whiteboards/{boardID}/connections/{fromID}/{toID}", handler.disconnectWhiteboardPins)
	}
	mux.HandleFunc("POST /api/characters", handler.createCharacter)
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		ctx, cancel := context.WithTimeout(r.Context(), 30*time.Second)
		defer cancel()
		mux.ServeHTTP(w, r.WithContext(ctx))
	})
}

func (h *Handler) npcRecords(w http.ResponseWriter, r *http.Request) {
	current, ok := h.authorizedPlayer(w, r)
	if !ok {
		return
	}
	from, err := time.Parse("2006-01-02", r.URL.Query().Get("from"))
	if err != nil {
		http.Error(w, "from must be a date in YYYY-MM-DD format", http.StatusBadRequest)
		return
	}
	through, err := time.Parse("2006-01-02", r.URL.Query().Get("to"))
	if err != nil {
		http.Error(w, "to must be a date in YYYY-MM-DD format", http.StatusBadRequest)
		return
	}
	if through.Before(from) || through.Sub(from) > (records.MaxRangeDays-1)*24*time.Hour {
		http.Error(w, "date range must be valid and no longer than 7 calendar days", http.StatusBadRequest)
		return
	}
	source := r.URL.Query().Get("source")
	if source != "" && source != "bank" && source != "gate" {
		http.Error(w, "source must be bank or gate", http.StatusBadRequest)
		return
	}
	through = through.AddDate(0, 0, 1).Add(-time.Nanosecond)
	results, err := h.recordsReader.NPCRecords(r.Context(), h.worldServerID, r.PathValue("npcID"), from, through, source)
	if errors.Is(err, world.ErrEntityNotFound) {
		http.NotFound(w, r)
		return
	}
	if errors.Is(err, records.ErrInvalidSource) {
		http.Error(w, "source must be bank or gate", http.StatusBadRequest)
		return
	}
	if errors.Is(err, records.ErrInvalidQuery) || errors.Is(err, records.ErrRangeTooLarge) {
		http.Error(w, "date range must be valid and no longer than 7 calendar days", http.StatusBadRequest)
		return
	}
	if err != nil {
		log.Printf("retrieve personal records for NPC %q: %v", r.PathValue("npcID"), err)
		http.Error(w, "could not retrieve personal records", http.StatusInternalServerError)
		return
	}
	if source != "" {
		filtered := make([]records.Record, 0, len(results))
		for _, result := range results {
			if result.Source == source {
				filtered = append(filtered, result)
			}
		}
		results = filtered
	}
	items := make([]cases.DiscoveredItem, 0, len(results))
	for _, result := range results {
		metadata, _ := json.Marshal(result)
		title := "GATE ACCESS LOG"
		if result.Source == "bank" {
			title = "FINANCIAL RECORD"
		}
		content := result.Description
		if result.Amount > 0 {
			content += " · " + strconv.Itoa(result.Amount) + " Credits"
		}
		if result.Origin != "" || result.Destination != "" {
			content += " · " + result.Origin + " → " + result.Destination
		}
		items = append(items, cases.DiscoveredItem{ID: result.ID, Source: result.Source, Title: title, Content: content, Metadata: metadata})
	}
	if h.whiteboardReader != nil {
		if err := h.whiteboardReader.RecordDiscoveredItems(r.Context(), h.worldServerID, current.ID, items); err != nil {
			log.Printf("record discovered personal records for NPC %q: %v", r.PathValue("npcID"), err)
			http.Error(w, "could not save retrieved records", http.StatusInternalServerError)
			return
		}
	}
	writeJSON(w, http.StatusOK, map[string]any{"results": results})
}

func (h *Handler) npcEmails(w http.ResponseWriter, r *http.Request) {
	current, ok := h.authorizedPlayer(w, r)
	if !ok {
		return
	}
	from, err := time.Parse("2006-01-02", r.URL.Query().Get("from"))
	if err != nil {
		http.Error(w, "from must be a date in YYYY-MM-DD format", http.StatusBadRequest)
		return
	}
	through, err := time.Parse("2006-01-02", r.URL.Query().Get("to"))
	if err != nil {
		http.Error(w, "to must be a date in YYYY-MM-DD format", http.StatusBadRequest)
		return
	}
	if through.Before(from) || through.Sub(from) > (correspondence.MaxRangeDays-1)*24*time.Hour {
		http.Error(w, "date range must be valid and no longer than 7 calendar days", http.StatusBadRequest)
		return
	}
	through = through.AddDate(0, 0, 1).Add(-time.Nanosecond)
	results, err := h.emailsReader.NPCEmails(r.Context(), h.worldServerID, r.PathValue("npcID"), from, through)
	if errors.Is(err, world.ErrEntityNotFound) {
		http.NotFound(w, r)
		return
	}
	if errors.Is(err, correspondence.ErrInvalidQuery) || errors.Is(err, correspondence.ErrRangeTooLarge) {
		http.Error(w, "date range must be valid and no longer than 7 calendar days", http.StatusBadRequest)
		return
	}
	if err != nil {
		log.Printf("retrieve correspondence for NPC %q: %v", r.PathValue("npcID"), err)
		http.Error(w, "could not retrieve correspondence", http.StatusInternalServerError)
		return
	}
	items := make([]cases.DiscoveredItem, 0, len(results))
	for _, result := range results {
		metadata, _ := json.Marshal(result)
		items = append(items, cases.DiscoveredItem{ID: result.ID, Source: "email", Title: result.Subject, Content: result.Sender + " → " + result.Recipient + "\n" + result.Body, Metadata: metadata})
	}
	if h.whiteboardReader != nil {
		if err := h.whiteboardReader.RecordDiscoveredItems(r.Context(), h.worldServerID, current.ID, items); err != nil {
			log.Printf("record discovered correspondence for NPC %q: %v", r.PathValue("npcID"), err)
			http.Error(w, "could not save retrieved correspondence", http.StatusInternalServerError)
			return
		}
	}
	writeJSON(w, http.StatusOK, map[string]any{"results": results})
}

type caseDetailResponse struct {
	Case     cases.PublicCase       `json:"case"`
	Evidence []cases.PublicEvidence `json:"evidence"`
	Progress cases.PlayerProgress   `json:"progress"`
}

type caseProgressRequest struct {
	DiscoveredEvidenceIDs []string                     `json:"discoveredEvidenceIds"`
	WhiteboardPins        []string                     `json:"whiteboardPins"`
	WhiteboardConnections []cases.WhiteboardConnection `json:"whiteboardConnections"`
}

func (h *Handler) listOpenCases(w http.ResponseWriter, r *http.Request) {
	if _, ok := h.caseAccount(w, r); !ok {
		return
	}
	results, err := h.caseReader.ListOpenPublicCases(r.Context(), h.worldServerID)
	if err != nil {
		http.Error(w, "could not retrieve open cases", http.StatusInternalServerError)
		return
	}
	writeJSON(w, http.StatusOK, map[string]any{"results": results})
}

func (h *Handler) getCase(w http.ResponseWriter, r *http.Request) {
	current, ok := h.caseAccount(w, r)
	if !ok {
		return
	}
	publicCase, evidence, err := h.caseReader.GetPublicCase(r.Context(), h.worldServerID, r.PathValue("caseID"))
	if errors.Is(err, cases.ErrCaseNotFound) {
		http.NotFound(w, r)
		return
	}
	if err != nil {
		http.Error(w, "could not retrieve case file", http.StatusInternalServerError)
		return
	}
	progress, err := h.caseReader.GetPlayerProgress(r.Context(), h.worldServerID, publicCase.ID, current.ID)
	if err != nil {
		http.Error(w, "could not retrieve case progress", http.StatusInternalServerError)
		return
	}
	if h.whiteboardReader != nil {
		items := make([]cases.DiscoveredItem, 0, len(evidence))
		for _, item := range evidence {
			metadata, _ := json.Marshal(item)
			source := item.Type
			if source != "email" && source != "bank" && source != "gate" {
				source = "case"
			}
			items = append(items, cases.DiscoveredItem{ID: item.ID, Source: source, Title: item.Presentation, Content: item.Content, Metadata: metadata})
		}
		if err := h.whiteboardReader.RecordDiscoveredItems(r.Context(), h.worldServerID, current.ID, items); err != nil {
			http.Error(w, "could not record case details", http.StatusInternalServerError)
			return
		}
	}
	writeJSON(w, http.StatusOK, caseDetailResponse{Case: publicCase, Evidence: evidence, Progress: progress})
}

type whiteboardCreateRequest struct {
	Title string `json:"title"`
	ID    string `json:"id"`
}
type whiteboardRenameRequest struct {
	Title string `json:"title"`
}
type whiteboardPinRequest struct {
	ItemID string `json:"itemId"`
	X      int    `json:"x"`
	Y      int    `json:"y"`
}
type whiteboardPositionRequest struct {
	X int `json:"x"`
	Y int `json:"y"`
}
type whiteboardConnectionRequest struct {
	FromID string `json:"fromId"`
	ToID   string `json:"toId"`
}

func (h *Handler) listWhiteboards(w http.ResponseWriter, r *http.Request) {
	current, ok := h.caseAccount(w, r)
	if !ok {
		return
	}
	boards, err := h.whiteboardReader.ListWhiteboards(r.Context(), h.worldServerID, current.ID)
	if err != nil {
		http.Error(w, "could not retrieve whiteboards", http.StatusInternalServerError)
		return
	}
	writeJSON(w, http.StatusOK, map[string]any{"results": boards})
}

func (h *Handler) createWhiteboard(w http.ResponseWriter, r *http.Request) {
	current, ok := h.caseAccount(w, r)
	if !ok {
		return
	}
	var input whiteboardCreateRequest
	if !decodeJSON(w, r, &input) {
		return
	}
	title := strings.TrimSpace(input.Title)
	if len([]rune(title)) > 64 {
		http.Error(w, "whiteboard title cannot exceed 64 characters", http.StatusBadRequest)
		return
	}
	boardID := strings.TrimSpace(input.ID)
	if boardID == "" {
		identifier := make([]byte, 18)
		if _, err := rand.Read(identifier); err != nil {
			http.Error(w, "could not create whiteboard", http.StatusInternalServerError)
			return
		}
		boardID = "board_" + base64.RawURLEncoding.EncodeToString(identifier)
	} else if len(boardID) > 80 || !whiteboardIDPattern.MatchString(boardID) {
		http.Error(w, "invalid whiteboard id", http.StatusBadRequest)
		return
	}
	board, err := h.whiteboardReader.CreateWhiteboard(r.Context(), h.worldServerID, current.ID, boardID, title)
	if err != nil {
		http.Error(w, "could not create whiteboard", http.StatusInternalServerError)
		return
	}
	writeJSON(w, http.StatusCreated, map[string]any{"whiteboard": board})
}

func (h *Handler) renameWhiteboard(w http.ResponseWriter, r *http.Request) {
	current, ok := h.caseAccount(w, r)
	if !ok {
		return
	}
	var input whiteboardRenameRequest
	if !decodeJSON(w, r, &input) {
		return
	}
	title := strings.TrimSpace(input.Title)
	if title == "" || len([]rune(title)) > 64 {
		http.Error(w, "whiteboard title must contain 1 to 64 characters", http.StatusBadRequest)
		return
	}
	if err := h.whiteboardReader.RenameWhiteboard(r.Context(), h.worldServerID, current.ID, r.PathValue("boardID"), title); err != nil {
		if errors.Is(err, cases.ErrCaseNotFound) {
			http.NotFound(w, r)
			return
		}
		http.Error(w, "could not rename whiteboard", http.StatusInternalServerError)
		return
	}
	w.WriteHeader(http.StatusNoContent)
}

func (h *Handler) deleteWhiteboard(w http.ResponseWriter, r *http.Request) {
	current, ok := h.caseAccount(w, r)
	if !ok {
		return
	}
	if err := h.whiteboardReader.DeleteWhiteboard(r.Context(), h.worldServerID, current.ID, r.PathValue("boardID")); err != nil {
		if errors.Is(err, cases.ErrCaseNotFound) {
			http.NotFound(w, r)
			return
		}
		http.Error(w, "could not delete whiteboard", http.StatusInternalServerError)
		return
	}
	w.WriteHeader(http.StatusNoContent)
}

func (h *Handler) getWhiteboard(w http.ResponseWriter, r *http.Request) {
	current, ok := h.caseAccount(w, r)
	if !ok {
		return
	}
	board, err := h.whiteboardReader.GetWhiteboard(r.Context(), h.worldServerID, current.ID, r.PathValue("boardID"))
	if errors.Is(err, cases.ErrCaseNotFound) {
		http.NotFound(w, r)
		return
	}
	if err != nil {
		http.Error(w, "could not retrieve whiteboard", http.StatusInternalServerError)
		return
	}
	writeJSON(w, http.StatusOK, map[string]any{"whiteboard": board})
}

func (h *Handler) pinWhiteboardItem(w http.ResponseWriter, r *http.Request) {
	current, ok := h.caseAccount(w, r)
	if !ok {
		return
	}
	var input whiteboardPinRequest
	if !decodeJSON(w, r, &input) {
		return
	}
	if err := h.whiteboardReader.PinDiscoveredItem(r.Context(), h.worldServerID, current.ID, r.PathValue("boardID"), input.ItemID, input.X, input.Y); err != nil {
		if errors.Is(err, cases.ErrEvidenceNotFound) || errors.Is(err, cases.ErrInvalidProgress) {
			http.Error(w, err.Error(), http.StatusBadRequest)
			return
		}
		http.Error(w, "could not pin discovered item", http.StatusInternalServerError)
		return
	}
	w.WriteHeader(http.StatusNoContent)
}

func (h *Handler) moveWhiteboardPin(w http.ResponseWriter, r *http.Request) {
	current, ok := h.caseAccount(w, r)
	if !ok {
		return
	}
	var input whiteboardPositionRequest
	if !decodeJSON(w, r, &input) {
		return
	}
	if err := h.whiteboardReader.MoveWhiteboardPin(r.Context(), h.worldServerID, current.ID, r.PathValue("boardID"), r.PathValue("itemID"), input.X, input.Y); err != nil {
		if errors.Is(err, cases.ErrInvalidProgress) {
			http.Error(w, err.Error(), http.StatusBadRequest)
			return
		}
		if errors.Is(err, cases.ErrEvidenceNotFound) {
			http.NotFound(w, r)
			return
		}
		http.Error(w, "could not move whiteboard pin", http.StatusInternalServerError)
		return
	}
	w.WriteHeader(http.StatusNoContent)
}

func (h *Handler) removeWhiteboardPin(w http.ResponseWriter, r *http.Request) {
	current, ok := h.caseAccount(w, r)
	if !ok {
		return
	}
	if err := h.whiteboardReader.RemoveWhiteboardPin(r.Context(), h.worldServerID, current.ID, r.PathValue("boardID"), r.PathValue("itemID")); err != nil {
		if errors.Is(err, cases.ErrEvidenceNotFound) {
			http.NotFound(w, r)
			return
		}
		http.Error(w, "could not remove whiteboard pin", http.StatusInternalServerError)
		return
	}
	w.WriteHeader(http.StatusNoContent)
}

func (h *Handler) connectWhiteboardPins(w http.ResponseWriter, r *http.Request) {
	current, ok := h.caseAccount(w, r)
	if !ok {
		return
	}
	var input whiteboardConnectionRequest
	if !decodeJSON(w, r, &input) {
		return
	}
	if err := h.whiteboardReader.ConnectWhiteboardPins(r.Context(), h.worldServerID, current.ID, r.PathValue("boardID"), input.FromID, input.ToID); err != nil {
		if errors.Is(err, cases.ErrInvalidProgress) {
			http.Error(w, err.Error(), http.StatusBadRequest)
			return
		}
		if errors.Is(err, cases.ErrEvidenceNotFound) {
			http.Error(w, err.Error(), http.StatusBadRequest)
			return
		}
		http.Error(w, "could not connect whiteboard pins", http.StatusInternalServerError)
		return
	}
	w.WriteHeader(http.StatusNoContent)
}

func (h *Handler) disconnectWhiteboardPins(w http.ResponseWriter, r *http.Request) {
	current, ok := h.caseAccount(w, r)
	if !ok {
		return
	}
	if err := h.whiteboardReader.DisconnectWhiteboardPins(r.Context(), h.worldServerID, current.ID, r.PathValue("boardID"), r.PathValue("fromID"), r.PathValue("toID")); err != nil {
		http.Error(w, "could not remove whiteboard connection", http.StatusInternalServerError)
		return
	}
	w.WriteHeader(http.StatusNoContent)
}

func (h *Handler) saveCaseProgress(w http.ResponseWriter, r *http.Request) {
	current, ok := h.caseAccount(w, r)
	if !ok {
		return
	}
	var input caseProgressRequest
	if !decodeJSON(w, r, &input) {
		return
	}
	caseID := r.PathValue("caseID")
	if _, _, err := h.caseReader.GetPublicCase(r.Context(), h.worldServerID, caseID); errors.Is(err, cases.ErrCaseNotFound) {
		http.NotFound(w, r)
		return
	} else if err != nil {
		http.Error(w, "could not retrieve case file", http.StatusInternalServerError)
		return
	}
	progress := cases.PlayerProgress{DiscoveredEvidenceIDs: input.DiscoveredEvidenceIDs, WhiteboardPins: input.WhiteboardPins, WhiteboardConnections: input.WhiteboardConnections}
	if err := h.caseReader.SavePlayerProgress(r.Context(), h.worldServerID, caseID, current.ID, progress); err != nil {
		if errors.Is(err, cases.ErrInvalidProgress) || errors.Is(err, cases.ErrEvidenceNotFound) {
			http.Error(w, err.Error(), http.StatusBadRequest)
			return
		}
		http.Error(w, "could not save case progress", http.StatusInternalServerError)
		return
	}
	writeJSON(w, http.StatusOK, map[string]any{"progress": progress})
}

func (h *Handler) caseAccount(w http.ResponseWriter, r *http.Request) (account, bool) {
	if h.caseReader == nil || h.worldServerID == "" {
		http.Error(w, "case service unavailable", http.StatusServiceUnavailable)
		return account{}, false
	}
	token, err := r.Cookie(sessionCookie)
	if err != nil || token.Value == "" {
		http.Error(w, "authentication required", http.StatusUnauthorized)
		return account{}, false
	}
	current, _, err := h.store.accountForSession(r.Context(), token.Value, h.now())
	if err != nil {
		http.Error(w, "authentication required", http.StatusUnauthorized)
		return account{}, false
	}
	return current, true
}

func (h *Handler) newspaper(w http.ResponseWriter, r *http.Request) {
	if !h.authorizeWorldRequest(w, r) {
		return
	}
	worldState := h.worldAt()
	issue, err := h.newspaperReader.Issue(r.Context(), h.worldServerID, worldState.City, worldState.Country, worldState.ServerTime)
	if err != nil {
		http.Error(w, "could not retrieve the latest newspaper", http.StatusInternalServerError)
		return
	}
	writeJSON(w, http.StatusOK, issue)
}

func (h *Handler) searchNPCs(w http.ResponseWriter, r *http.Request) {
	if !h.authorizeWorldRequest(w, r) {
		return
	}
	query, ok := directoryQuery(w, r)
	if !ok {
		return
	}
	results, err := h.worldReader.SearchNPCs(r.Context(), h.worldServerID, query)
	if err != nil {
		http.Error(w, "could not search people", http.StatusInternalServerError)
		return
	}
	writeJSON(w, http.StatusOK, map[string]any{"results": results})
}

func (h *Handler) searchCompanies(w http.ResponseWriter, r *http.Request) {
	if !h.authorizeWorldRequest(w, r) {
		return
	}
	query, ok := directoryQuery(w, r)
	if !ok {
		return
	}
	results, err := h.worldReader.SearchCompanies(r.Context(), h.worldServerID, query)
	if err != nil {
		http.Error(w, "could not search companies", http.StatusInternalServerError)
		return
	}
	writeJSON(w, http.StatusOK, map[string]any{"results": results})
}

func (h *Handler) searchLocations(w http.ResponseWriter, r *http.Request) {
	current, ok := h.authorizedPlayer(w, r)
	if !ok {
		return
	}
	query, ok := directoryQuery(w, r)
	if !ok {
		return
	}
	reader, ok := h.worldReader.(transitReader)
	if !ok {
		http.Error(w, "GATE transit unavailable", http.StatusServiceUnavailable)
		return
	}
	destinations, err := reader.TransitDistricts(r.Context(), h.worldServerID)
	if err != nil {
		http.Error(w, "could not retrieve current district", http.StatusInternalServerError)
		return
	}
	locationID, err := h.store.playerLocation(r.Context(), current.ID)
	if err != nil {
		http.Error(w, "could not retrieve current location", http.StatusInternalServerError)
		return
	}
	districtID := districtForLocation(destinations, locationID)
	if districtID == "" {
		http.Error(w, "current location is unavailable", http.StatusConflict)
		return
	}
	results, err := h.worldReader.SearchLocations(r.Context(), h.worldServerID, districtID, query)
	if err != nil {
		http.Error(w, "could not search addresses", http.StatusInternalServerError)
		return
	}
	writeJSON(w, http.StatusOK, map[string]any{"results": results})
}

type transitReader interface {
	TransitDistricts(context.Context, string) ([]world.DistrictDestination, error)
}

func (h *Handler) transitDistricts(w http.ResponseWriter, r *http.Request) {
	if !h.authorizeWorldRequest(w, r) {
		return
	}
	reader, ok := h.worldReader.(transitReader)
	if !ok {
		http.Error(w, "GATE transit unavailable", http.StatusServiceUnavailable)
		return
	}
	districts, err := reader.TransitDistricts(r.Context(), h.worldServerID)
	if err != nil {
		http.Error(w, "could not retrieve GATE destinations", http.StatusInternalServerError)
		return
	}
	writeJSON(w, http.StatusOK, map[string]any{"destinations": districts})
}

func (h *Handler) travel(w http.ResponseWriter, r *http.Request) {
	current, ok := h.authorizedPlayer(w, r)
	if !ok {
		return
	}
	reader, ok := h.worldReader.(transitReader)
	if !ok {
		http.Error(w, "GATE transit unavailable", http.StatusServiceUnavailable)
		return
	}
	var input struct {
		DistrictID string `json:"districtId"`
	}
	if !decodeJSON(w, r, &input) {
		return
	}
	destinations, err := reader.TransitDistricts(r.Context(), h.worldServerID)
	if err != nil {
		http.Error(w, "could not retrieve GATE destinations", http.StatusInternalServerError)
		return
	}
	valid := false
	for _, destination := range destinations {
		if destination.ID == input.DistrictID {
			valid = true
			break
		}
	}
	if !valid {
		http.Error(w, "unknown GATE destination", http.StatusBadRequest)
		return
	}
	if err := h.store.setPlayerLocation(r.Context(), current.ID, input.DistrictID); err != nil {
		http.Error(w, "could not save player location", http.StatusInternalServerError)
		return
	}
	writeJSON(w, http.StatusOK, map[string]string{"locationId": input.DistrictID})
}

func (h *Handler) visit(w http.ResponseWriter, r *http.Request) {
	current, ok := h.authorizedPlayer(w, r)
	if !ok {
		return
	}
	reader, ok := h.worldReader.(transitReader)
	if !ok {
		http.Error(w, "GATE transit unavailable", http.StatusServiceUnavailable)
		return
	}
	var input struct {
		LocationID string `json:"locationId"`
	}
	if !decodeJSON(w, r, &input) {
		return
	}
	destinations, err := reader.TransitDistricts(r.Context(), h.worldServerID)
	if err != nil {
		http.Error(w, "could not retrieve current district", http.StatusInternalServerError)
		return
	}
	currentLocation, err := h.store.playerLocation(r.Context(), current.ID)
	if err != nil {
		http.Error(w, "could not retrieve current location", http.StatusInternalServerError)
		return
	}
	currentDistrictID := districtForLocation(destinations, currentLocation)
	if currentDistrictID == "" {
		http.Error(w, "current location is unavailable", http.StatusConflict)
		return
	}
	valid := input.LocationID == currentDistrictID
	for _, destination := range destinations {
		if destination.ID != currentDistrictID {
			continue
		}
		for _, place := range destination.Locations {
			if place.Address.LocationID == input.LocationID {
				valid = true
				break
			}
		}
	}
	if !valid {
		http.Error(w, "location is not in the current district", http.StatusBadRequest)
		return
	}
	if err := h.store.setPlayerLocation(r.Context(), current.ID, input.LocationID); err != nil {
		http.Error(w, "could not save player location", http.StatusInternalServerError)
		return
	}
	writeJSON(w, http.StatusOK, map[string]string{"locationId": input.LocationID, "districtId": currentDistrictID})
}

func districtForLocation(destinations []world.DistrictDestination, locationID string) string {
	for _, destination := range destinations {
		if destination.ID == locationID {
			return destination.ID
		}
		for _, place := range destination.Locations {
			if place.Address.LocationID == locationID {
				return destination.ID
			}
		}
	}
	return ""
}

func (h *Handler) authorizedPlayer(w http.ResponseWriter, r *http.Request) (account, bool) {
	if !h.authorizeWorldRequest(w, r) {
		return account{}, false
	}
	token, _ := r.Cookie(sessionCookie)
	current, _, err := h.store.accountForSession(r.Context(), token.Value, h.now())
	if err != nil {
		http.Error(w, "authentication required", http.StatusUnauthorized)
		return account{}, false
	}
	return current, true
}

func (h *Handler) authorizeWorldRequest(w http.ResponseWriter, r *http.Request) bool {
	if h.worldReader == nil || h.worldServerID == "" {
		http.Error(w, "world directory unavailable", http.StatusServiceUnavailable)
		return false
	}
	token, err := r.Cookie(sessionCookie)
	if err != nil || token.Value == "" {
		http.Error(w, "authentication required", http.StatusUnauthorized)
		return false
	}
	if _, _, err := h.store.accountForSession(r.Context(), token.Value, h.now()); err != nil {
		http.Error(w, "authentication required", http.StatusUnauthorized)
		return false
	}
	return true
}

func directoryQuery(w http.ResponseWriter, r *http.Request) (string, bool) {
	query := strings.TrimSpace(r.URL.Query().Get("q"))
	if query == "" || len([]rune(query)) > 80 {
		http.Error(w, "enter a search term of up to 80 characters", http.StatusBadRequest)
		return "", false
	}
	return query, true
}

type registerRequest struct {
	Username string `json:"username"`
	Email    string `json:"email"`
	Password string `json:"password"`
}

type loginRequest struct {
	Email    string `json:"email"`
	Password string `json:"password"`
}

type characterRequest struct {
	DisplayName string `json:"displayName"`
	PortraitID  int    `json:"portraitId"`
}

type identityResponse struct {
	Account   account       `json:"account"`
	Character *character    `json:"character"`
	World     worldResponse `json:"world"`
}

func (h *Handler) register(w http.ResponseWriter, r *http.Request) {
	var input registerRequest
	if !decodeJSON(w, r, &input) {
		return
	}
	input.Username = strings.TrimSpace(input.Username)
	input.Email = strings.ToLower(strings.TrimSpace(input.Email))
	if !usernamePattern.MatchString(input.Username) {
		http.Error(w, "username must be 3-24 characters using letters, numbers, underscores or hyphens", http.StatusBadRequest)
		return
	}
	if !validEmail(input.Email) {
		http.Error(w, "enter a valid email address", http.StatusBadRequest)
		return
	}
	if len(input.Password) < 12 || len(input.Password) > 72 {
		http.Error(w, "password must be between 12 and 72 bytes", http.StatusBadRequest)
		return
	}

	hash, err := hashPassword(input.Password)
	if err != nil {
		http.Error(w, "could not create account", http.StatusInternalServerError)
		return
	}
	created, err := h.store.createAccount(r.Context(), input.Username, input.Email, hash)
	if errors.Is(err, ErrConflict) {
		http.Error(w, "username or email already registered", http.StatusConflict)
		return
	}
	if err != nil {
		http.Error(w, "could not create account", http.StatusInternalServerError)
		return
	}
	if !h.startSession(w, r, created) {
		http.Error(w, "could not start session", http.StatusInternalServerError)
		return
	}
	writeJSON(w, http.StatusCreated, identityResponse{Account: publicAccount(created), Character: nil})
}

func (h *Handler) login(w http.ResponseWriter, r *http.Request) {
	var input loginRequest
	if !decodeJSON(w, r, &input) {
		return
	}
	input.Email = strings.ToLower(strings.TrimSpace(input.Email))
	if !validEmail(input.Email) || len(input.Password) == 0 || len(input.Password) > 72 {
		http.Error(w, "invalid email or password", http.StatusUnauthorized)
		return
	}
	found, err := h.store.accountByEmail(r.Context(), input.Email)
	if err != nil || !verifyPassword(found.PasswordHash, input.Password) {
		http.Error(w, "invalid email or password", http.StatusUnauthorized)
		return
	}
	if !h.startSession(w, r, found) {
		http.Error(w, "could not start session", http.StatusInternalServerError)
		return
	}
	writeJSON(w, http.StatusOK, identityResponse{
		Account:   publicAccount(found),
		Character: found.Character,
		World:     h.worldFor(r.Context(), found.ID),
	})
}

func (h *Handler) me(w http.ResponseWriter, r *http.Request) {
	token, err := r.Cookie(sessionCookie)
	if err != nil || token.Value == "" {
		http.Error(w, "authentication required", http.StatusUnauthorized)
		return
	}
	current, character, err := h.store.accountForSession(r.Context(), token.Value, h.now())
	if err != nil {
		http.Error(w, "authentication required", http.StatusUnauthorized)
		return
	}
	writeJSON(w, http.StatusOK, identityResponse{Account: publicAccount(current), Character: character, World: h.worldFor(r.Context(), current.ID)})
}

type worldResponse struct {
	City       string    `json:"city"`
	Country    string    `json:"country"`
	TimeZone   string    `json:"timeZone"`
	LocationID string    `json:"locationId"`
	DistrictID string    `json:"districtId"`
	District   string    `json:"district,omitempty"`
	PlaceName  string    `json:"placeName,omitempty"`
	ServerTime time.Time `json:"serverTime"`
}

func (h *Handler) world(w http.ResponseWriter, r *http.Request) {
	token, err := r.Cookie(sessionCookie)
	if err != nil || token.Value == "" {
		http.Error(w, "authentication required", http.StatusUnauthorized)
		return
	}
	current, _, err := h.store.accountForSession(r.Context(), token.Value, h.now())
	if err != nil {
		http.Error(w, "authentication required", http.StatusUnauthorized)
		return
	}
	writeJSON(w, http.StatusOK, h.worldFor(r.Context(), current.ID))
}

func (h *Handler) worldAt() worldResponse {
	parisTime := time.Date(1974, time.November, 12, 21, 15, 0, 0, time.FixedZone("CET", 60*60))
	gameTime := parisTime.Add(h.now().Sub(h.worldClockStartedAt))
	return worldResponse{
		City:       "Lyon",
		Country:    "France",
		TimeZone:   "Europe/Paris",
		LocationID: "lyon",
		ServerTime: gameTime.UTC(),
	}
}

func (h *Handler) worldFor(ctx context.Context, accountID string) worldResponse {
	state := h.worldAt()
	locationID, err := h.store.playerLocation(ctx, accountID)
	if err != nil {
		return state
	}
	state.LocationID = locationID
	if reader, ok := h.worldReader.(transitReader); ok {
		if destinations, err := reader.TransitDistricts(ctx, h.worldServerID); err == nil {
			for _, district := range destinations {
				if district.ID == locationID {
					state.DistrictID, state.District = district.ID, district.Name
				}
				for _, place := range district.Locations {
					if place.Address.LocationID == locationID {
						state.DistrictID, state.District = district.ID, district.Name
						state.PlaceName = place.Address.Name
					}
				}
			}
		}
	}
	return state
}

func (h *Handler) createCharacter(w http.ResponseWriter, r *http.Request) {
	token, err := r.Cookie(sessionCookie)
	if err != nil || token.Value == "" {
		http.Error(w, "authentication required", http.StatusUnauthorized)
		return
	}
	current, existing, err := h.store.accountForSession(r.Context(), token.Value, h.now())
	if err != nil {
		http.Error(w, "authentication required", http.StatusUnauthorized)
		return
	}
	if existing != nil {
		http.Error(w, "character already exists", http.StatusConflict)
		return
	}
	var input characterRequest
	if !decodeJSON(w, r, &input) {
		return
	}
	input.DisplayName = strings.TrimSpace(input.DisplayName)
	if len([]rune(input.DisplayName)) < 2 || len([]rune(input.DisplayName)) > 40 || input.PortraitID < 1 || input.PortraitID > 6 {
		http.Error(w, "invalid character details", http.StatusBadRequest)
		return
	}
	created, err := h.store.createCharacter(r.Context(), current.ID, input.DisplayName, input.PortraitID)
	if errors.Is(err, ErrConflict) {
		http.Error(w, "character already exists", http.StatusConflict)
		return
	}
	if err != nil {
		http.Error(w, "could not create character", http.StatusInternalServerError)
		return
	}
	writeJSON(w, http.StatusCreated, identityResponse{Account: publicAccount(current), Character: &created})
}

func (h *Handler) logout(w http.ResponseWriter, r *http.Request) {
	if token, err := r.Cookie(sessionCookie); err == nil && token.Value != "" {
		if err := h.store.deleteSession(r.Context(), token.Value); err != nil {
			http.Error(w, "could not end session", http.StatusInternalServerError)
			return
		}
	}
	http.SetCookie(w, &http.Cookie{Name: sessionCookie, Value: "", Path: "/", MaxAge: -1, HttpOnly: true, Secure: h.secureCookie, SameSite: http.SameSiteLaxMode})
	w.WriteHeader(http.StatusNoContent)
}

func (h *Handler) startSession(w http.ResponseWriter, r *http.Request, current account) bool {
	raw := make([]byte, 32)
	if _, err := rand.Read(raw); err != nil {
		return false
	}
	token := base64.RawURLEncoding.EncodeToString(raw)
	expires := h.now().Add(sessionDuration)
	if err := h.store.createSession(r.Context(), current.ID, token, expires); err != nil {
		return false
	}
	http.SetCookie(w, &http.Cookie{Name: sessionCookie, Value: token, Path: "/", Expires: expires, MaxAge: int(sessionDuration.Seconds()), HttpOnly: true, Secure: h.secureCookie, SameSite: http.SameSiteLaxMode})
	return true
}

func decodeJSON(w http.ResponseWriter, r *http.Request, target any) bool {
	r.Body = http.MaxBytesReader(w, r.Body, 16*1024)
	decoder := json.NewDecoder(r.Body)
	decoder.DisallowUnknownFields()
	if err := decoder.Decode(target); err != nil {
		http.Error(w, "invalid request", http.StatusBadRequest)
		return false
	}
	if err := decoder.Decode(new(any)); errors.Is(err, io.EOF) {
		return true
	}
	http.Error(w, "invalid request", http.StatusBadRequest)
	return false
}

func validEmail(value string) bool {
	if len(value) > 254 {
		return false
	}
	parsed, err := mail.ParseAddress(value)
	return err == nil && parsed.Address == value && strings.Contains(value, "@")
}

func hashPassword(password string) (string, error) {
	hash, err := bcrypt.GenerateFromPassword([]byte(password), bcrypt.DefaultCost)
	return string(hash), err
}

func verifyPassword(encoded, password string) bool {
	return bcrypt.CompareHashAndPassword([]byte(encoded), []byte(password)) == nil
}

func publicAccount(value account) account { value.PasswordHash = ""; return value }

func writeJSON(w http.ResponseWriter, status int, value any) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	_ = json.NewEncoder(w).Encode(value)
}

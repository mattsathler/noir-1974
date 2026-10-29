package identity

import (
	"bytes"
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
	"runtime/debug"
	"strconv"
	"strings"
	"time"

	"golang.org/x/crypto/bcrypt"
	"noir-1974/apps/api/internal/cases"
	correspondence "noir-1974/apps/api/internal/mail"
	"noir-1974/apps/api/internal/newspaper"
	"noir-1974/apps/api/internal/observability"
	"noir-1974/apps/api/internal/records"
	"noir-1974/apps/api/internal/world"
)

const sessionCookie = "noir_session"
const sessionDuration = 30 * 24 * time.Hour
const apiRequestTimeout = 30 * time.Second

var usernamePattern = regexp.MustCompile(`^[a-zA-Z0-9_-]{3,24}$`)
var whiteboardIDPattern = regexp.MustCompile(`^board_[A-Za-z0-9_-]+$`)

type Handler struct {
	store              store
	secureCookie       bool
	now                func() time.Time
	worldReader        WorldDirectoryReader
	newspaperReader    NewspaperReader
	caseReader         CaseAPIReader
	whiteboardReader   WhiteboardAPIReader
	recordsReader      RecordsAPIReader
	emailsReader       EmailsAPIReader
	worldServerID      string
	exposeErrorDetails bool
}

type requestIDContextKey struct{}

type apiErrorResponse struct {
	Code      string `json:"code"`
	Message   string `json:"message"`
	RequestID string `json:"requestId"`
	Details   string `json:"details,omitempty"`
}

type bufferedResponse struct {
	http.ResponseWriter
	status int
	body   bytes.Buffer
}

func (response *bufferedResponse) WriteHeader(status int) {
	if response.status == 0 {
		response.status = status
	}
}

func (response *bufferedResponse) Write(body []byte) (int, error) {
	if response.status == 0 {
		response.status = http.StatusOK
	}
	return response.body.Write(body)
}

type WorldDirectoryReader interface {
	SearchNPCs(context.Context, string, string) ([]world.PersonSearchRecord, error)
	SearchCompanies(context.Context, string, string) ([]world.CompanyRecord, error)
	SearchLocations(context.Context, string, string, string) ([]world.LocationRecord, error)
}

type cityCompanySearchReader interface {
	SearchCompaniesInCity(context.Context, string, string, string) ([]world.CompanyRecord, error)
}

type countryCompanySearchReader interface {
	SearchCompaniesInCountry(context.Context, string, string, string) ([]world.CompanyPortalRecord, error)
}

type cityNPCReader interface {
	NPCBelongsToCity(context.Context, string, string, string) (bool, error)
}

type cityPersonSearchReader interface {
	SearchNPCsInCity(context.Context, string, string, string) ([]world.PersonSearchRecord, error)
}

type countryPersonSearchReader interface {
	SearchNPCsInCountry(context.Context, string, string, string) ([]world.PersonSearchRecord, error)
}

type phoneDirectoryReader interface {
	PhoneContactByNumber(context.Context, string, string, string) (world.PhoneContact, error)
}

type companyLocationReader interface {
	CompanyAtLocation(context.Context, string, string) ([]world.CompanyRecord, error)
}

type companyDirectoryReader interface {
	CompanyDirectoryAtLocation(context.Context, string, string, string) ([]world.CompanyRecord, error)
}

type NewspaperReader interface {
	Issue(context.Context, string, string, string, string, time.Time) (newspaper.Issue, error)
}

type newspaperInvalidator interface{ InvalidateCity(string, string) }

type CaseAccusationReader interface {
	SubmitAccusation(context.Context, string, string, string, string, time.Time) (cases.AccusationResult, error)
}

type WhiteboardDesignationReader interface {
	SetWhiteboardPinDesignation(context.Context, string, string, string, string, string) error
}

type CaseAPIReader interface {
	ListOpenPublicCases(context.Context, string, string) ([]cases.PublicCase, error)
	GetPublicCase(context.Context, string, string) (cases.PublicCase, []cases.PublicEvidence, error)
	GetPlayerProgress(context.Context, string, string, string) (cases.PlayerProgress, error)
	SavePlayerProgress(context.Context, string, string, string, cases.PlayerProgress) error
}

type crimeReportsReader interface {
	CityCrimeReports(context.Context, string, string, time.Time, time.Time, time.Time) (cases.CrimeReportDigest, error)
}

type caseSessionLocationReader interface {
	accountForSessionAndLocation(context.Context, string, time.Time) (account, *character, string, error)
}

type sessionCaseListReader interface {
	ListOpenPublicCasesForSession(context.Context, string, string, time.Time) ([]cases.PublicCase, error)
}

type CaseInterrogationReader interface {
	InterrogationAnswersForNPC(context.Context, string, string, string) (cases.InterrogationAnswers, error)
}

type CasePersonSearchReader interface {
	PersonSearchScenes(context.Context, string, string, []string) (map[string]world.PersonSearchScene, error)
}

type ResidentialSearchReader interface {
	RequestResidentialSearch(context.Context, string, string, string, string, string, time.Time) (cases.WarrantRequest, error)
	ListPlayerMail(context.Context, string, string) ([]cases.PlayerMailMessage, error)
}

type PlayerMailDeleteReader interface {
	DeletePlayerMail(context.Context, string, string, string) error
}

type HospitalRecordsReader interface {
	SearchHospitalPatients(context.Context, string, string, string) ([]cases.HospitalPatient, []cases.DiscoveredItem, error)
}

type LocalAccessAttemptReader interface {
	RecordLocalAccessAttempt(context.Context, string, string, string, string, string, string, string, time.Time) (cases.LocalAccessAttempt, error)
	ValidateLocalAccessAttempt(context.Context, string, string, int64, string, string, string, string, time.Time) (bool, error)
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
	CreateWhiteboardNote(context.Context, string, string, string, int, int) (cases.WhiteboardPin, error)
	UpdateWhiteboardNote(context.Context, string, string, string, string, string, string) error
	PinDiscoveredItem(context.Context, string, string, string, string, int, int) error
	MoveWhiteboardPin(context.Context, string, string, string, string, int, int) error
	RemoveWhiteboardPin(context.Context, string, string, string, string) error
	ConnectWhiteboardPins(context.Context, string, string, string, string, string) error
	DisconnectWhiteboardPins(context.Context, string, string, string, string, string) error
}

// NewHandler creates the identity API. Cookies are secure in production deployments.
func NewHandler(repository store, secureCookie bool) http.Handler {
	return newHandler(repository, secureCookie, nil, "", nil, nil, nil, nil, false, nil)
}

// NewHandlerWithWorld composes authenticated world directory routes into the identity API.
func NewHandlerWithWorld(repository store, secureCookie bool, reader WorldDirectoryReader, serverID string) http.Handler {
	return newHandler(repository, secureCookie, reader, serverID, nil, nil, nil, nil, false, nil)
}

func NewHandlerWithNewspaper(repository store, secureCookie bool, reader WorldDirectoryReader, newspaperReader NewspaperReader, serverID string) http.Handler {
	return newHandler(repository, secureCookie, reader, serverID, newspaperReader, nil, nil, nil, false, nil)
}

func NewHandlerWithGame(repository store, secureCookie bool, reader WorldDirectoryReader, newspaperReader NewspaperReader, caseReader CaseAPIReader, recordsReader RecordsAPIReader, serverID string) http.Handler {
	return newHandler(repository, secureCookie, reader, serverID, newspaperReader, caseReader, recordsReader, nil, false, nil)
}

func NewHandlerWithGameAndMail(repository store, secureCookie bool, reader WorldDirectoryReader, newspaperReader NewspaperReader, caseReader CaseAPIReader, recordsReader RecordsAPIReader, emailsReader EmailsAPIReader, serverID string) http.Handler {
	return NewHandlerWithGameAndMailDiagnostics(repository, secureCookie, reader, newspaperReader, caseReader, recordsReader, emailsReader, serverID, false)
}

func NewHandlerWithGameAndMailDiagnostics(repository store, secureCookie bool, reader WorldDirectoryReader, newspaperReader NewspaperReader, caseReader CaseAPIReader, recordsReader RecordsAPIReader, emailsReader EmailsAPIReader, serverID string, exposeErrorDetails bool) http.Handler {
	return newHandler(repository, secureCookie, reader, serverID, newspaperReader, caseReader, recordsReader, emailsReader, exposeErrorDetails, nil)
}

// NewHandlerWithGameAndMailMetrics composes the game API with route-level latency recording.
func NewHandlerWithGameAndMailMetrics(repository store, secureCookie bool, reader WorldDirectoryReader, newspaperReader NewspaperReader, caseReader CaseAPIReader, recordsReader RecordsAPIReader, emailsReader EmailsAPIReader, serverID string, exposeErrorDetails bool, recorder *observability.HTTPRecorder) http.Handler {
	return newHandler(repository, secureCookie, reader, serverID, newspaperReader, caseReader, recordsReader, emailsReader, exposeErrorDetails, recorder)
}

func newHandler(repository store, secureCookie bool, reader WorldDirectoryReader, serverID string, newspaperReader NewspaperReader, caseReader CaseAPIReader, recordsReader RecordsAPIReader, emailsReader EmailsAPIReader, exposeErrorDetails bool, recorder *observability.HTTPRecorder) http.Handler {
	now := time.Now
	handler := &Handler{store: repository, secureCookie: secureCookie, now: now, worldReader: reader, newspaperReader: newspaperReader, caseReader: caseReader, recordsReader: recordsReader, emailsReader: emailsReader, worldServerID: serverID, exposeErrorDetails: exposeErrorDetails}
	if whiteboardReader, ok := caseReader.(WhiteboardAPIReader); ok {
		handler.whiteboardReader = whiteboardReader
	}
	mux := http.NewServeMux()
	mux.HandleFunc("POST /api/auth/register", handler.register)
	mux.HandleFunc("POST /api/auth/login", handler.login)
	mux.HandleFunc("POST /api/auth/logout", handler.logout)
	mux.HandleFunc("GET /api/me", handler.me)
	mux.HandleFunc("GET /api/character-visuals", handler.characterVisuals)
	mux.HandleFunc("GET /api/world", handler.world)
	if reader != nil {
		mux.HandleFunc("GET /api/world/npcs", handler.searchNPCs)
		mux.HandleFunc("GET /api/world/companies", handler.searchCompanies)
		mux.HandleFunc("GET /api/world/companies/national", handler.searchCompaniesInCountry)
		mux.HandleFunc("GET /api/world/locations", handler.searchLocations)
		mux.HandleFunc("GET /api/world/transit", handler.transitDistricts)
		mux.HandleFunc("GET /api/world/gate/destinations", handler.gateDestinations)
		mux.HandleFunc("POST /api/world/gate/travel", handler.gateTravel)
		mux.HandleFunc("GET /api/world/proxy/nodes", handler.proxyNodes)
		mux.HandleFunc("POST /api/world/proxy/connect", handler.proxyConnect)
		if _, ok := reader.(phoneDirectoryReader); ok && caseReader != nil {
			mux.HandleFunc("GET /api/world/phone/{number}", handler.lookupPhoneContact)
		}
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
		if _, ok := caseReader.(crimeReportsReader); ok {
			mux.HandleFunc("GET /api/crime-reports", handler.crimeReports)
		}
		mux.HandleFunc("GET /api/cases/{caseID}", handler.getCase)
		mux.HandleFunc("POST /api/cases/{caseID}/progress", handler.saveCaseProgress)
		if _, ok := caseReader.(CaseAccusationReader); ok {
			mux.HandleFunc("POST /api/accusations", handler.submitAccusation)
		}
		if _, ok := caseReader.(CaseInterrogationReader); ok {
			mux.HandleFunc("GET /api/cases/interrogation/{npcID}", handler.getInterrogationAnswers)
		}
		if _, ok := caseReader.(ResidentialSearchReader); ok {
			mux.HandleFunc("POST /api/police/warrants", handler.requestResidentialSearch)
			mux.HandleFunc("GET /api/player/emails", handler.playerEmails)
		}
		if _, ok := caseReader.(PlayerMailDeleteReader); ok {
			mux.HandleFunc("DELETE /api/player/emails/{messageID}", handler.deletePlayerEmail)
		}
		if _, ok := caseReader.(LocalAccessAttemptReader); ok {
			mux.HandleFunc("POST /api/facilities/access", handler.requestLocalAccess)
		}
		if _, ok := caseReader.(HospitalRecordsReader); ok {
			if _, secured := caseReader.(LocalAccessAttemptReader); secured && handler.whiteboardReader != nil {
				mux.HandleFunc("GET /api/hospital/patients", handler.searchHospitalPatients)
			}
		}
	}
	if handler.whiteboardReader != nil {
		mux.HandleFunc("GET /api/whiteboards", handler.listWhiteboards)
		mux.HandleFunc("POST /api/whiteboards", handler.createWhiteboard)
		mux.HandleFunc("GET /api/whiteboards/{boardID}", handler.getWhiteboard)
		mux.HandleFunc("PATCH /api/whiteboards/{boardID}", handler.renameWhiteboard)
		mux.HandleFunc("DELETE /api/whiteboards/{boardID}", handler.deleteWhiteboard)
		mux.HandleFunc("POST /api/whiteboards/{boardID}/notes", handler.createWhiteboardNote)
		mux.HandleFunc("PATCH /api/whiteboards/{boardID}/notes/{itemID}", handler.updateWhiteboardNote)
		mux.HandleFunc("POST /api/whiteboards/{boardID}/pins", handler.pinWhiteboardItem)
		mux.HandleFunc("PATCH /api/whiteboards/{boardID}/pins/{itemID}", handler.moveWhiteboardPin)
		if _, ok := handler.whiteboardReader.(WhiteboardDesignationReader); ok {
			mux.HandleFunc("PUT /api/whiteboards/{boardID}/pins/{itemID}/designation", handler.setWhiteboardPinDesignation)
		}
		mux.HandleFunc("DELETE /api/whiteboards/{boardID}/pins/{itemID}", handler.removeWhiteboardPin)
		mux.HandleFunc("POST /api/whiteboards/{boardID}/connections", handler.connectWhiteboardPins)
		mux.HandleFunc("DELETE /api/whiteboards/{boardID}/connections/{fromID}/{toID}", handler.disconnectWhiteboardPins)
	}
	mux.HandleFunc("POST /api/characters", handler.createCharacter)
	mux.HandleFunc("PUT /api/characters/appearance", handler.updateCharacterAppearance)
	measured := http.Handler(apiErrorMiddleware(mux))
	measured = requestTimeoutMiddleware(apiRequestTimeout, measured)
	if recorder != nil {
		measured = recorder.Middleware(func(request *http.Request) string {
			_, pattern := mux.Handler(request)
			if _, route, ok := strings.Cut(pattern, " "); ok {
				return route
			}
			return pattern
		}, measured)
	}
	return measured
}

func (h *Handler) searchHospitalPatients(w http.ResponseWriter, r *http.Request) {
	current, cityID, ok := h.caseAccountAndCity(w, r)
	if !ok {
		return
	}
	locationID := expectedServiceLocationID(cityID, "hospital")
	if locationID == "" || !h.requireLocalAccessAttemptInCity(w, r, current.ID, cityID, "hospital", locationID, "") {
		return
	}
	query, ok := directoryQuery(w, r)
	if !ok {
		return
	}
	reader, ok := h.caseReader.(HospitalRecordsReader)
	if !ok {
		http.Error(w, "hospital records unavailable", http.StatusServiceUnavailable)
		return
	}
	patients, discovered, err := reader.SearchHospitalPatients(r.Context(), h.worldServerID, cityID, query)
	if err != nil {
		h.writeAPIError(w, r, http.StatusInternalServerError, "HOSPITAL_SEARCH_FAILED", "could not search local patient records", err)
		return
	}
	if h.whiteboardReader != nil && len(discovered) > 0 {
		if err := h.whiteboardReader.RecordDiscoveredItems(r.Context(), h.worldServerID, current.ID, discovered); err != nil {
			h.writeAPIError(w, r, http.StatusInternalServerError, "HOSPITAL_RECORD_FAILED", "could not save forensic records", err)
			return
		}
	}
	writeJSON(w, http.StatusOK, map[string]any{"results": patients})
}

func requestTimeoutMiddleware(timeout time.Duration, next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		ctx, cancel := context.WithTimeout(r.Context(), timeout)
		defer cancel()
		next.ServeHTTP(w, r.WithContext(ctx))
	})
}

type residentialSearchRequest struct {
	TargetNPCID       string `json:"targetNpcId"`
	StationLocationID string `json:"stationLocationId"`
}

type localAccessRequest struct {
	LocationID string `json:"locationId"`
	CompanyID  string `json:"companyId"`
	Facility   string `json:"facility"`
}

func (h *Handler) requestLocalAccess(w http.ResponseWriter, r *http.Request) {
	current, ok := h.caseAccount(w, r)
	if !ok {
		return
	}
	var input localAccessRequest
	if !decodeJSON(w, r, &input) {
		return
	}
	reader, ok := h.caseReader.(LocalAccessAttemptReader)
	if !ok {
		http.Error(w, "local access service unavailable", http.StatusServiceUnavailable)
		return
	}
	state := h.worldFor(r.Context(), current.ID)
	if state.CityID == "" || strings.TrimSpace(input.LocationID) == "" || strings.TrimSpace(input.Facility) == "" {
		h.writeAPIError(w, r, http.StatusBadRequest, "LOCAL_ACCESS_CONTEXT_REQUIRED", "a local address and system are required", nil)
		return
	}
	valid := input.LocationID != "" && state.CityID != ""
	switch input.Facility {
	case "company":
		valid = valid && input.CompanyID != ""
		companyReader, available := h.worldReader.(companyDirectoryReader)
		if !available {
			valid = false
			break
		}
		records, err := companyReader.CompanyDirectoryAtLocation(r.Context(), h.worldServerID, input.LocationID, input.CompanyID)
		if err != nil {
			h.writeAPIError(w, r, http.StatusInternalServerError, "LOCAL_ACCESS_FAILED", "could not validate local system", err)
			return
		}
		city, exists := world.CityByID(state.CityID)
		valid = valid && exists && len(records) == 1 && records[0].Address.City == city.Name
	case "police", "bank", "gate":
		valid = valid && input.CompanyID == "" && input.LocationID == expectedServiceLocationID(state.CityID, input.Facility)
	case "hospital":
		valid = valid && input.CompanyID == "" && input.LocationID == expectedServiceLocationID(state.CityID, input.Facility)
	default:
		valid = false
	}
	status := "denied"
	if valid {
		status = "granted"
	}
	attempt, err := reader.RecordLocalAccessAttempt(r.Context(), h.worldServerID, current.ID, state.CityID, input.LocationID, input.CompanyID, input.Facility, status, h.now().UTC())
	if err != nil {
		h.writeAPIError(w, r, http.StatusInternalServerError, "LOCAL_ACCESS_FAILED", "could not record local system access", err)
		return
	}
	writeJSON(w, http.StatusOK, attempt)
}

func (h *Handler) requestResidentialSearch(w http.ResponseWriter, r *http.Request) {
	current, ok := h.caseAccount(w, r)
	if !ok {
		return
	}
	var input residentialSearchRequest
	if !decodeJSON(w, r, &input) {
		return
	}
	reader, ok := h.caseReader.(ResidentialSearchReader)
	if !ok {
		http.Error(w, "residential search unavailable", http.StatusServiceUnavailable)
		return
	}
	state := h.worldFor(r.Context(), current.ID)
	result, err := reader.RequestResidentialSearch(r.Context(), h.worldServerID, current.ID, state.CityID, input.StationLocationID, input.TargetNPCID, state.ServerTime)
	if errors.Is(err, cases.ErrCaseNotFound) {
		http.Error(w, "the target or police station is not valid for the current city", http.StatusBadRequest)
		return
	}
	if err != nil {
		h.writeAPIError(w, r, http.StatusInternalServerError, "RESIDENTIAL_SEARCH_FAILED", "could not complete the residential search", err)
		return
	}
	writeJSON(w, http.StatusCreated, result)
}

func (h *Handler) playerEmails(w http.ResponseWriter, r *http.Request) {
	current, ok := h.caseAccount(w, r)
	if !ok {
		return
	}
	reader, ok := h.caseReader.(ResidentialSearchReader)
	if !ok {
		http.Error(w, "personal mailbox unavailable", http.StatusServiceUnavailable)
		return
	}
	messages, err := reader.ListPlayerMail(r.Context(), h.worldServerID, current.ID)
	if err != nil {
		h.writeAPIError(w, r, http.StatusInternalServerError, "PLAYER_EMAIL_FAILED", "could not retrieve personal mail", err)
		return
	}
	writeJSON(w, http.StatusOK, map[string]any{"results": messages})
}

func (h *Handler) deletePlayerEmail(w http.ResponseWriter, r *http.Request) {
	current, ok := h.caseAccount(w, r)
	if !ok {
		return
	}
	reader, ok := h.caseReader.(PlayerMailDeleteReader)
	if !ok {
		http.Error(w, "personal mailbox unavailable", http.StatusServiceUnavailable)
		return
	}
	if err := reader.DeletePlayerMail(r.Context(), h.worldServerID, current.ID, r.PathValue("messageID")); err != nil {
		if errors.Is(err, cases.ErrPlayerMailNotFound) {
			http.NotFound(w, r)
			return
		}
		h.writeAPIError(w, r, http.StatusInternalServerError, "PLAYER_EMAIL_DELETE_FAILED", "could not delete personal mail", err)
		return
	}
	w.WriteHeader(http.StatusNoContent)
}

func apiErrorMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		requestID := newRequestID()
		w.Header().Set("X-Request-ID", requestID)
		ctx := context.WithValue(r.Context(), requestIDContextKey{}, requestID)
		response := &bufferedResponse{ResponseWriter: w}
		defer func() {
			if recovered := recover(); recovered != nil {
				log.Printf("api_panic request_id=%s method=%s path=%s panic=%v\n%s", requestID, r.Method, r.URL.Path, recovered, debug.Stack())
				response.body.Reset()
				w.Header().Set("Content-Type", "application/json; charset=utf-8")
				w.WriteHeader(http.StatusInternalServerError)
				_ = json.NewEncoder(w).Encode(apiErrorResponse{
					Code: "INTERNAL_SERVER_ERROR", Message: "An unexpected server error occurred.", RequestID: requestID,
				})
			}
		}()
		next.ServeHTTP(response, r.WithContext(ctx))
		if errors.Is(r.Context().Err(), context.DeadlineExceeded) {
			log.Printf("api_error request_id=%s method=%s path=%s code=REQUEST_TIMEOUT error=request deadline exceeded", requestID, r.Method, r.URL.Path)
			response.body.Reset()
			response.status = http.StatusGatewayTimeout
			w.Header().Set("Content-Type", "application/json; charset=utf-8")
			w.WriteHeader(response.status)
			_ = json.NewEncoder(w).Encode(apiErrorResponse{
				Code: "REQUEST_TIMEOUT", Message: "The request exceeded its time limit.", RequestID: requestID,
			})
			return
		}
		if response.status == 0 {
			response.status = http.StatusOK
		}
		if response.status >= http.StatusBadRequest && !strings.Contains(response.Header().Get("Content-Type"), "application/json") {
			message := strings.TrimSpace(response.body.String())
			log.Printf("api_error request_id=%s method=%s path=%s status=%d message=%q", requestID, r.Method, r.URL.Path, response.status, message)
			w.Header().Set("Content-Type", "application/json; charset=utf-8")
			w.WriteHeader(response.status)
			_ = json.NewEncoder(w).Encode(apiErrorResponse{Code: apiErrorCode(response.status), Message: message, RequestID: requestID})
			return
		}
		w.WriteHeader(response.status)
		_, _ = w.Write(response.body.Bytes())
	})
}

func newRequestID() string {
	value := make([]byte, 12)
	if _, err := rand.Read(value); err != nil {
		return strconv.FormatInt(time.Now().UnixNano(), 16)
	}
	return base64.RawURLEncoding.EncodeToString(value)
}

func apiErrorCode(status int) string {
	switch status {
	case http.StatusBadRequest:
		return "BAD_REQUEST"
	case http.StatusUnauthorized:
		return "UNAUTHORIZED"
	case http.StatusForbidden:
		return "FORBIDDEN"
	case http.StatusNotFound:
		return "NOT_FOUND"
	case http.StatusConflict:
		return "CONFLICT"
	case http.StatusServiceUnavailable:
		return "SERVICE_UNAVAILABLE"
	case http.StatusGatewayTimeout:
		return "REQUEST_TIMEOUT"
	default:
		if status >= http.StatusInternalServerError {
			return "INTERNAL_SERVER_ERROR"
		}
		return "REQUEST_FAILED"
	}
}

func (handler *Handler) writeAPIError(w http.ResponseWriter, r *http.Request, status int, code, message string, cause error) {
	requestID, _ := r.Context().Value(requestIDContextKey{}).(string)
	if cause != nil {
		log.Printf("api_error request_id=%s method=%s path=%s code=%s error=%v", requestID, r.Method, r.URL.Path, code, cause)
	}
	response := apiErrorResponse{Code: code, Message: message, RequestID: requestID}
	if handler.exposeErrorDetails && cause != nil {
		response.Details = cause.Error()
	}
	w.Header().Set("X-Request-ID", requestID)
	writeJSON(w, status, response)
}

func (h *Handler) npcRecords(w http.ResponseWriter, r *http.Request) {
	current, ok := h.authorizedPlayer(w, r)
	if !ok {
		return
	}
	if !h.validateNPCCity(w, r, current.ID, r.PathValue("npcID")) {
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
	if source != "" {
		locationID := strings.TrimSpace(r.URL.Query().Get("locationId"))
		_, canValidateCity := h.worldReader.(cityNPCReader)
		if (locationID == "" && canValidateCity) || (locationID != "" && locationID != expectedServiceLocationID(h.worldFor(r.Context(), current.ID).CityID, source)) {
			h.writeAPIError(w, r, http.StatusBadRequest, "RECORDS_LOCATION_REQUIRED", "select the matching local records service", nil)
			return
		}
		if locationID != "" && !h.requireLocalAccessAttempt(w, r, current.ID, source, locationID, "") {
			return
		}
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
	if len(results) > world.MaxSearchResults {
		results = results[:world.MaxSearchResults]
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

func (h *Handler) characterVisuals(w http.ResponseWriter, _ *http.Request) {
	writeJSON(w, http.StatusOK, world.VisualIdentityCatalog())
}

func (h *Handler) npcEmails(w http.ResponseWriter, r *http.Request) {
	current, ok := h.authorizedPlayer(w, r)
	if !ok {
		return
	}
	cityID := h.worldFor(r.Context(), current.ID).CityID
	if !h.validateNPCInCity(w, r, cityID, r.PathValue("npcID")) {
		return
	}
	locationID := strings.TrimSpace(r.URL.Query().Get("locationId"))
	_, canValidateCity := h.worldReader.(cityNPCReader)
	if (locationID == "" && canValidateCity) || (locationID != "" && locationID != expectedServiceLocationID(cityID, "police")) {
		h.writeAPIError(w, r, http.StatusBadRequest, "CORRESPONDENCE_LOCATION_REQUIRED", "select the local police station to retrieve correspondence", nil)
		return
	}
	if locationID != "" && !h.requireLocalAccessAttemptInCity(w, r, current.ID, cityID, "police", locationID, "") {
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
		h.writeAPIError(w, r, http.StatusNotFound, "NPC_EMAIL_PROFILE_NOT_FOUND", "the selected person has no correspondence profile in this city", err)
		return
	}
	if errors.Is(err, correspondence.ErrInvalidQuery) || errors.Is(err, correspondence.ErrRangeTooLarge) {
		http.Error(w, "date range must be valid and no longer than 7 calendar days", http.StatusBadRequest)
		return
	}
	if err != nil {
		requestID, _ := r.Context().Value(requestIDContextKey{}).(string)
		log.Printf("api_error request_id=%s method=%s path=%s code=NPC_CORRESPONDENCE_FAILED error=%v", requestID, r.Method, r.URL.Path, err)
		http.Error(w, "could not retrieve correspondence", http.StatusInternalServerError)
		return
	}
	now := h.now().UTC()
	available := results[:0]
	for _, result := range results {
		if !result.SentAt.After(now) {
			available = append(available, result)
		}
	}
	results = available
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
	if reader, ok := h.caseReader.(sessionCaseListReader); ok {
		token, err := r.Cookie(sessionCookie)
		if err != nil || token.Value == "" {
			http.Error(w, "authentication required", http.StatusUnauthorized)
			return
		}
		results, err := reader.ListOpenPublicCasesForSession(r.Context(), h.worldServerID, token.Value, h.now())
		if errors.Is(err, cases.ErrSessionNotFound) {
			http.Error(w, "authentication required", http.StatusUnauthorized)
			return
		}
		if err != nil {
			http.Error(w, "could not retrieve open cases", http.StatusInternalServerError)
			return
		}
		writeJSON(w, http.StatusOK, map[string]any{"results": results})
		return
	}
	_, cityID, ok := h.caseAccountAndCity(w, r)
	if !ok {
		return
	}
	results, err := h.caseReader.ListOpenPublicCases(r.Context(), h.worldServerID, cityID)
	if err != nil {
		http.Error(w, "could not retrieve open cases", http.StatusInternalServerError)
		return
	}
	writeJSON(w, http.StatusOK, map[string]any{"results": results})
}

func (h *Handler) crimeReports(w http.ResponseWriter, r *http.Request) {
	reader, ok := h.caseReader.(crimeReportsReader)
	if !ok {
		http.Error(w, "crime reports unavailable", http.StatusServiceUnavailable)
		return
	}
	_, cityID, ok := h.caseAccountAndCity(w, r)
	if !ok {
		return
	}
	now := h.now()
	from, until, err := crimeReportDateRange(cityID, r.URL.Query().Get("from"), r.URL.Query().Get("to"), now)
	if err != nil {
		http.Error(w, "invalid crime report date range", http.StatusBadRequest)
		return
	}
	result, err := reader.CityCrimeReports(r.Context(), h.worldServerID, cityID, from, until, now)
	if err != nil {
		h.writeAPIError(w, r, http.StatusInternalServerError, "CRIME_REPORTS_UNAVAILABLE", "could not retrieve city crime reports", err)
		return
	}
	writeJSON(w, http.StatusOK, result)
}

func crimeReportDateRange(cityID, fromText, toText string, now time.Time) (time.Time, time.Time, error) {
	zone := time.UTC
	if city, ok := world.CityByID(cityID); ok {
		if location, err := time.LoadLocation(city.TimeZone); err == nil {
			zone = location
		}
	}
	today := now.In(zone)
	if fromText == "" && toText == "" {
		fromText = today.AddDate(0, 0, -30).Format("2006-01-02")
		toText = today.Format("2006-01-02")
	} else if fromText == "" {
		fromText = toText
	} else if toText == "" {
		toText = fromText
	}
	from, err := time.ParseInLocation("2006-01-02", fromText, zone)
	if err != nil {
		return time.Time{}, time.Time{}, err
	}
	to, err := time.ParseInLocation("2006-01-02", toText, zone)
	if err != nil {
		return time.Time{}, time.Time{}, err
	}
	if from.After(to) {
		return time.Time{}, time.Time{}, errors.New("date range starts after it ends")
	}
	return from.UTC(), to.AddDate(0, 0, 1).UTC(), nil
}

func (h *Handler) caseAccountAndCity(w http.ResponseWriter, r *http.Request) (account, string, bool) {
	reader, ok := h.store.(caseSessionLocationReader)
	if !ok {
		current, valid := h.caseAccount(w, r)
		if !valid {
			return account{}, "", false
		}
		return current, h.worldFor(r.Context(), current.ID).CityID, true
	}
	token, err := r.Cookie(sessionCookie)
	if err != nil || token.Value == "" {
		http.Error(w, "authentication required", http.StatusUnauthorized)
		return account{}, "", false
	}
	current, _, locationID, err := reader.accountForSessionAndLocation(r.Context(), token.Value, h.now())
	if err != nil {
		http.Error(w, "authentication required", http.StatusUnauthorized)
		return account{}, "", false
	}
	cityID := h.worldAt().CityID
	if city, ok := world.CityForLocation(locationID); ok {
		cityID = city.ID
	}
	return current, cityID, true
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

func (h *Handler) getInterrogationAnswers(w http.ResponseWriter, r *http.Request) {
	current, ok := h.caseAccount(w, r)
	if !ok {
		return
	}
	if !h.validateNPCCity(w, r, current.ID, r.PathValue("npcID")) {
		return
	}
	reader, ok := h.caseReader.(CaseInterrogationReader)
	if !ok {
		http.Error(w, "interrogation service unavailable", http.StatusServiceUnavailable)
		return
	}
	answers, err := reader.InterrogationAnswersForNPC(r.Context(), h.worldServerID, h.worldFor(r.Context(), current.ID).CityID, r.PathValue("npcID"))
	if err != nil {
		http.Error(w, "could not retrieve interview testimony", http.StatusInternalServerError)
		return
	}
	writeJSON(w, http.StatusOK, map[string]any{"answers": answers})
}

func (h *Handler) validateNPCCity(w http.ResponseWriter, r *http.Request, accountID, npcID string) bool {
	return h.validateNPCInCity(w, r, h.worldFor(r.Context(), accountID).CityID, npcID)
}

func (h *Handler) validateNPCInCity(w http.ResponseWriter, r *http.Request, cityID, npcID string) bool {
	reader, ok := h.worldReader.(cityNPCReader)
	if !ok {
		return true
	}
	valid, err := reader.NPCBelongsToCity(r.Context(), h.worldServerID, cityID, npcID)
	if err != nil {
		h.writeAPIError(w, r, http.StatusInternalServerError, "NPC_CITY_VALIDATION_FAILED", "could not validate the selected person", err)
		return false
	}
	if !valid {
		h.writeAPIError(w, r, http.StatusNotFound, "NPC_NOT_IN_CURRENT_CITY", "the selected person is not available in the current city", nil)
		return false
	}
	return true
}

func expectedServiceLocationID(cityID, facility string) string {
	cityKey := cityID
	if cityID == "fr/lyon" {
		cityKey = "lyon"
	}
	suffix := map[string]string{"police": "police-station", "bank": "meridian-bank", "gate": "gate-station", "hospital": "hospital"}[facility]
	if suffix == "" || cityKey == "" {
		return ""
	}
	return "service:" + cityKey + "/" + suffix
}

func (h *Handler) requireLocalAccessAttempt(w http.ResponseWriter, r *http.Request, accountID, facility, locationID, companyID string) bool {
	cityID := h.worldFor(r.Context(), accountID).CityID
	return h.requireLocalAccessAttemptInCity(w, r, accountID, cityID, facility, locationID, companyID)
}

func (h *Handler) requireLocalAccessAttemptInCity(w http.ResponseWriter, r *http.Request, accountID, cityID, facility, locationID, companyID string) bool {
	attemptID, err := strconv.ParseInt(strings.TrimSpace(r.URL.Query().Get("accessAttemptId")), 10, 64)
	if err != nil || attemptID < 1 {
		h.writeAPIError(w, r, http.StatusForbidden, "LOCAL_ACCESS_REQUIRED", "connect to the institution's website before requesting these records", nil)
		return false
	}
	reader, ok := h.caseReader.(LocalAccessAttemptReader)
	if !ok {
		h.writeAPIError(w, r, http.StatusServiceUnavailable, "LOCAL_ACCESS_UNAVAILABLE", "local access validation is unavailable", nil)
		return false
	}
	valid, err := reader.ValidateLocalAccessAttempt(r.Context(), h.worldServerID, accountID, attemptID, cityID, locationID, companyID, facility, h.now().UTC())
	if err != nil {
		h.writeAPIError(w, r, http.StatusInternalServerError, "LOCAL_ACCESS_VALIDATION_FAILED", "could not validate institutional website access", err)
		return false
	}
	if !valid {
		h.writeAPIError(w, r, http.StatusForbidden, "LOCAL_ACCESS_REQUIRED", "connect to the institution's website before requesting these records", nil)
		return false
	}
	return true
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
type whiteboardDesignationRequest struct {
	Designation string `json:"designation"`
}
type accusationRequest struct {
	BoardID           string `json:"boardId"`
	StationLocationID string `json:"stationLocationId"`
}
type whiteboardNoteRequest struct {
	Title   string `json:"title"`
	Content string `json:"content"`
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

func (h *Handler) setWhiteboardPinDesignation(w http.ResponseWriter, r *http.Request) {
	current, ok := h.caseAccount(w, r)
	if !ok {
		return
	}
	var input whiteboardDesignationRequest
	if !decodeJSON(w, r, &input) {
		return
	}
	reader, ok := h.whiteboardReader.(WhiteboardDesignationReader)
	if !ok {
		http.Error(w, "whiteboard service unavailable", http.StatusServiceUnavailable)
		return
	}
	err := reader.SetWhiteboardPinDesignation(r.Context(), h.worldServerID, current.ID, r.PathValue("boardID"), r.PathValue("itemID"), input.Designation)
	if errors.Is(err, cases.ErrEvidenceNotFound) || errors.Is(err, cases.ErrInvalidProgress) {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}
	if err != nil {
		h.writeAPIError(w, r, http.StatusInternalServerError, "WHITEBOARD_DESIGNATION_FAILED", "could not update the pin designation", err)
		return
	}
	writeJSON(w, http.StatusOK, map[string]any{"designation": input.Designation})
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

func (h *Handler) createWhiteboardNote(w http.ResponseWriter, r *http.Request) {
	current, ok := h.caseAccount(w, r)
	if !ok {
		return
	}
	var input whiteboardPositionRequest
	if !decodeJSON(w, r, &input) {
		return
	}
	pin, err := h.whiteboardReader.CreateWhiteboardNote(r.Context(), h.worldServerID, current.ID, r.PathValue("boardID"), input.X, input.Y)
	if errors.Is(err, cases.ErrInvalidProgress) {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}
	if errors.Is(err, cases.ErrCaseNotFound) {
		http.NotFound(w, r)
		return
	}
	if err != nil {
		http.Error(w, "could not create whiteboard note", http.StatusInternalServerError)
		return
	}
	writeJSON(w, http.StatusCreated, map[string]any{"pin": pin})
}

func (h *Handler) updateWhiteboardNote(w http.ResponseWriter, r *http.Request) {
	current, ok := h.caseAccount(w, r)
	if !ok {
		return
	}
	var input whiteboardNoteRequest
	if !decodeJSON(w, r, &input) {
		return
	}
	err := h.whiteboardReader.UpdateWhiteboardNote(r.Context(), h.worldServerID, current.ID, r.PathValue("boardID"), r.PathValue("itemID"), input.Title, input.Content)
	if errors.Is(err, cases.ErrInvalidWhiteboardNote) {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}
	if errors.Is(err, cases.ErrEvidenceNotFound) {
		http.NotFound(w, r)
		return
	}
	if err != nil {
		http.Error(w, "could not save whiteboard note", http.StatusInternalServerError)
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

func (h *Handler) submitAccusation(w http.ResponseWriter, r *http.Request) {
	current, ok := h.caseAccount(w, r)
	if !ok {
		return
	}
	var input accusationRequest
	if !decodeJSON(w, r, &input) {
		return
	}
	if input.BoardID == "" {
		http.Error(w, "boardId is required", http.StatusBadRequest)
		return
	}
	reader, ok := h.caseReader.(CaseAccusationReader)
	if !ok {
		http.Error(w, "accusation service unavailable", http.StatusServiceUnavailable)
		return
	}
	cityID := h.worldFor(r.Context(), current.ID).CityID
	cityKey := cityID
	if cityID == "fr/lyon" {
		cityKey = "lyon"
	}
	if input.StationLocationID != "service:"+cityKey+"/police-station" {
		http.Error(w, "a valid police station in the current city is required", http.StatusBadRequest)
		return
	}
	result, err := reader.SubmitAccusation(r.Context(), h.worldServerID, current.ID, cityID, input.BoardID, h.now())
	if errors.Is(err, cases.ErrCaseNotFound) {
		http.NotFound(w, r)
		return
	}
	if errors.Is(err, cases.ErrInvalidProgress) || errors.Is(err, cases.ErrEvidenceNotFound) {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}
	if err != nil {
		h.writeAPIError(w, r, http.StatusInternalServerError, "ACCUSATION_FAILED", "could not submit the accusation", err)
		return
	}
	if result.Accepted {
		if invalidator, ok := h.newspaperReader.(newspaperInvalidator); ok {
			invalidator.InvalidateCity(h.worldServerID, cityID)
		}
	}
	writeJSON(w, http.StatusCreated, map[string]any{"result": result})
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

func (h *Handler) lookupPhoneContact(w http.ResponseWriter, r *http.Request) {
	current, ok := h.caseAccount(w, r)
	if !ok {
		return
	}
	reader, ok := h.worldReader.(phoneDirectoryReader)
	if !ok {
		http.Error(w, "phone directory unavailable", http.StatusServiceUnavailable)
		return
	}
	cityID := h.worldFor(r.Context(), current.ID).CityID
	contact, err := reader.PhoneContactByNumber(r.Context(), h.worldServerID, cityID, r.PathValue("number"))
	if errors.Is(err, world.ErrEntityNotFound) {
		h.writeAPIError(w, r, http.StatusNotFound, "PHONE_NUMBER_NOT_FOUND", "number not in directory", err)
		return
	}
	if err != nil {
		h.writeAPIError(w, r, http.StatusInternalServerError, "PHONE_LOOKUP_FAILED", "could not search the phone directory", err)
		return
	}
	writeJSON(w, http.StatusOK, map[string]any{"contact": contact})
}

func (h *Handler) newspaper(w http.ResponseWriter, r *http.Request) {
	if !h.authorizeWorldRequest(w, r) {
		return
	}
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
	worldState := h.worldFor(r.Context(), current.ID)
	issue, err := h.newspaperReader.Issue(r.Context(), h.worldServerID, worldState.CityID, worldState.City, worldState.Country, worldState.ServerTime)
	if err != nil {
		h.writeAPIError(w, r, http.StatusInternalServerError, "NEWSPAPER_UNAVAILABLE", "could not retrieve the latest newspaper", err)
		return
	}
	writeJSON(w, http.StatusOK, issue)
}

func (h *Handler) searchNPCs(w http.ResponseWriter, r *http.Request) {
	var current account
	var cityID string
	var ok bool
	if h.whiteboardReader != nil {
		current, cityID, ok = h.caseAccountAndCity(w, r)
	} else {
		current, valid := h.worldAccount(w, r)
		ok = valid
		if valid {
			cityID = h.worldFor(r.Context(), current.ID).CityID
		}
	}
	if !ok {
		return
	}
	city, exists := world.CityByID(cityID)
	if !exists {
		h.writeAPIError(w, r, http.StatusInternalServerError, "PERSON_SEARCH_FAILED", "could not resolve the current country", world.ErrEntityNotFound)
		return
	}
	query, ok := directoryQuery(w, r)
	if !ok {
		return
	}
	var results []world.PersonSearchRecord
	var err error
	if countryReader, ok := h.worldReader.(countryPersonSearchReader); ok {
		results, err = countryReader.SearchNPCsInCountry(r.Context(), h.worldServerID, city.CountryID, query)
	} else {
		results, err = h.worldReader.SearchNPCs(r.Context(), h.worldServerID, query)
		for index := 0; err == nil && index < len(results); {
			npc := results[index].NPC
			valid := strings.HasPrefix(npc.ID, "npc:"+city.CountryID+"/") || (city.CountryID == "fr" && strings.HasPrefix(npc.ID, "npc:lyon/"))
			if !valid {
				results = append(results[:index], results[index+1:]...)
				continue
			}
			index++
		}
	}
	if err != nil {
		h.writeAPIError(w, r, http.StatusInternalServerError, "PERSON_SEARCH_FAILED", "could not search people", err)
		return
	}
	if len(results) > world.MaxNPCSearchResults {
		results = results[:world.MaxNPCSearchResults]
	}
	if h.whiteboardReader != nil {
		if reader, ok := h.caseReader.(CasePersonSearchReader); ok {
			npcIDs := make([]string, 0, len(results))
			for _, result := range results {
				npcIDs = append(npcIDs, result.NPC.ID)
			}
			scenes, err := reader.PersonSearchScenes(r.Context(), h.worldServerID, city.CountryID, npcIDs)
			if err != nil {
				h.writeAPIError(w, r, http.StatusInternalServerError, "PERSON_SEARCH_FAILED", "could not retrieve case scene references", err)
				return
			}
			for index := range results {
				if scene, ok := scenes[results[index].NPC.ID]; ok {
					results[index].ReportedScene = &scene
				}
			}
		}
		items := make([]cases.DiscoveredItem, 0, len(results))
		for _, result := range results {
			metadata, _ := json.Marshal(result)
			workplaces := make([]string, 0, len(result.Workplaces))
			for _, workplace := range result.Workplaces {
				workplaces = append(workplaces, workplace.Company+" — "+workplace.Address)
			}
			if len(workplaces) == 0 {
				workplaces = append(workplaces, "NOT LISTED")
			}
			appearance := result.NPC.Appearance
			if appearance == "" {
				appearance = "NOT RECORDED"
			}
			maritalStatus := result.MaritalStatus
			if maritalStatus == "" {
				maritalStatus = "UNKNOWN"
			}
			personFacts := []string{
				"AGE · " + strconv.Itoa(result.NPC.Age) + " YEARS",
				"STATUS · " + strings.ToUpper(result.NPC.Status),
				"DESCRIPTION · " + appearance,
				"EMPLOYMENT · " + strings.Join(workplaces, "; "),
				"MARITAL STATUS · " + strings.ToUpper(maritalStatus),
			}
			if result.ReportedScene != nil {
				personFacts = append(personFacts, "REPORTED SCENE · "+result.ReportedScene.Address)
			}
			content := strings.Join(personFacts, "\n")
			items = append(items, cases.DiscoveredItem{
				ID:       "person:" + result.NPC.ID,
				Source:   "person",
				Title:    result.NPC.Name,
				Content:  content,
				Metadata: metadata,
			})
		}
		if err := h.whiteboardReader.RecordDiscoveredItems(r.Context(), h.worldServerID, current.ID, items); err != nil {
			log.Printf("record personnel files for whiteboard: %v", err)
			http.Error(w, "could not save personnel files", http.StatusInternalServerError)
			return
		}
	}
	writeJSON(w, http.StatusOK, map[string]any{"results": results})
}

func (h *Handler) searchCompanies(w http.ResponseWriter, r *http.Request) {
	locationID := strings.TrimSpace(r.URL.Query().Get("locationId"))
	locationLookup := locationID != ""
	requestedCityID := strings.TrimSpace(r.URL.Query().Get("cityId"))
	var current account
	if h.whiteboardReader != nil || locationLookup || requestedCityID != "" {
		var ok bool
		current, ok = h.caseAccount(w, r)
		if !ok {
			return
		}
		if locationLookup {
			locationCity, exists := world.CityForLocation(locationID)
			if !exists || locationCity.ID != h.worldFor(r.Context(), current.ID).CityID {
				h.writeAPIError(w, r, http.StatusBadRequest, "COMPANY_LOCATION_UNAVAILABLE", "the selected address is not in your current city", nil)
				return
			}
			if companyID := strings.TrimSpace(r.URL.Query().Get("companyId")); companyID != "" && !h.requireLocalAccessAttempt(w, r, current.ID, "company", locationID, companyID) {
				return
			}
		}
		if requestedCityID != "" && requestedCityID != h.worldFor(r.Context(), current.ID).CityID {
			h.writeAPIError(w, r, http.StatusBadRequest, "COMPANY_CITY_UNAVAILABLE", "the requested city is not your current city", nil)
			return
		}
	} else if !h.authorizeWorldRequest(w, r) {
		return
	}
	var results []world.CompanyRecord
	var err error
	if locationLookup {
		if companyID := strings.TrimSpace(r.URL.Query().Get("companyId")); companyID != "" {
			reader, ok := h.worldReader.(companyDirectoryReader)
			if !ok {
				h.writeAPIError(w, r, http.StatusServiceUnavailable, "COMPANY_DIRECTORY_UNAVAILABLE", "Company directory lookup is temporarily unavailable.", nil)
				return
			}
			results, err = reader.CompanyDirectoryAtLocation(r.Context(), h.worldServerID, locationID, companyID)
		} else if reader, ok := h.worldReader.(companyLocationReader); ok {
			results, err = reader.CompanyAtLocation(r.Context(), h.worldServerID, locationID)
		} else {
			h.writeAPIError(w, r, http.StatusServiceUnavailable, "COMPANY_DIRECTORY_UNAVAILABLE", "Company directory lookup is temporarily unavailable.", nil)
			return
		}
	} else {
		query, ok := directoryQuery(w, r)
		if !ok {
			return
		}
		if requestedCityID != "" {
			if reader, ok := h.worldReader.(cityCompanySearchReader); ok {
				results, err = reader.SearchCompaniesInCity(r.Context(), h.worldServerID, requestedCityID, query)
			} else {
				results, err = h.worldReader.SearchCompanies(r.Context(), h.worldServerID, query)
			}
		} else {
			results, err = h.worldReader.SearchCompanies(r.Context(), h.worldServerID, query)
		}
	}
	if err != nil {
		http.Error(w, "could not search companies", http.StatusInternalServerError)
		return
	}
	if !locationLookup && len(results) > world.MaxSearchResults {
		results = results[:world.MaxSearchResults]
	}
	if h.whiteboardReader != nil {
		items := make([]cases.DiscoveredItem, 0, len(results))
		for _, result := range results {
			metadata, _ := json.Marshal(result)
			items = append(items, cases.DiscoveredItem{
				ID:       result.Company.ID,
				Source:   "company",
				Title:    result.Company.Name,
				Content:  "HEADQUARTERS · " + result.Address.Full + "\nPERSONNEL ON FILE · " + strconv.Itoa(len(result.Employees)),
				Metadata: metadata,
			})
		}
		if err := h.whiteboardReader.RecordDiscoveredItems(r.Context(), h.worldServerID, current.ID, items); err != nil {
			log.Printf("record company files for whiteboard: %v", err)
			http.Error(w, "could not save company files", http.StatusInternalServerError)
			return
		}
	}
	writeJSON(w, http.StatusOK, map[string]any{"results": results})
}

func (h *Handler) searchCompaniesInCountry(w http.ResponseWriter, r *http.Request) {
	current, ok := h.authorizedPlayer(w, r)
	if !ok {
		return
	}
	reader, ok := h.worldReader.(countryCompanySearchReader)
	if !ok {
		h.writeAPIError(w, r, http.StatusServiceUnavailable, "COMPANY_PORTAL_UNAVAILABLE", "The national company portal is temporarily unavailable.", nil)
		return
	}
	query, ok := directoryQuery(w, r)
	if !ok {
		return
	}
	activeCity := h.worldFor(r.Context(), current.ID).CityID
	city, exists := world.CityByID(activeCity)
	if !exists {
		h.writeAPIError(w, r, http.StatusConflict, "PROXY_JURISDICTION_UNAVAILABLE", "Connect a valid Proxy city before using the company portal.", nil)
		return
	}
	results, err := reader.SearchCompaniesInCountry(r.Context(), h.worldServerID, city.CountryID, query)
	if err != nil {
		h.writeAPIError(w, r, http.StatusInternalServerError, "COMPANY_PORTAL_SEARCH_FAILED", "Company listings could not be searched.", err)
		return
	}
	if len(results) > 10 {
		results = results[:10]
	}
	if h.whiteboardReader != nil {
		items := make([]cases.DiscoveredItem, 0, len(results))
		for _, result := range results {
			metadata, _ := json.Marshal(result)
			items = append(items, cases.DiscoveredItem{ID: result.Company.ID, Source: "company", Title: result.Company.Name, Content: "HEADQUARTERS · " + result.Address.Full, Metadata: metadata})
		}
		if err := h.whiteboardReader.RecordDiscoveredItems(r.Context(), h.worldServerID, current.ID, items); err != nil {
			log.Printf("record national company files for whiteboard: %v", err)
			h.writeAPIError(w, r, http.StatusInternalServerError, "COMPANY_PORTAL_RECORD_FAILED", "Company search results could not be saved to your case files.", err)
			return
		}
	}
	writeJSON(w, http.StatusOK, map[string]any{"results": results, "countryId": city.CountryID, "limit": 10})
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
		http.Error(w, "local directory unavailable", http.StatusServiceUnavailable)
		return
	}
	locationID, err := h.store.playerLocation(r.Context(), current.ID)
	if err != nil {
		http.Error(w, "could not retrieve current district", http.StatusInternalServerError)
		return
	}
	city, ok := world.CityForLocation(locationID)
	if !ok {
		http.Error(w, "current city is unavailable", http.StatusConflict)
		return
	}
	destinations, err := h.transitDestinations(r.Context(), reader, city.ID)
	if err != nil {
		http.Error(w, "could not retrieve current district", http.StatusInternalServerError)
		return
	}
	districtID := districtForLocation(destinations, locationID)
	if districtID == "" {
		// Player locations intentionally persist only the city. When a local
		// directory lookup has no transient neighborhood context, use Central.
		for _, destination := range destinations {
			if strings.EqualFold(destination.Name, "Central") {
				districtID = destination.ID
				break
			}
		}
	}
	if districtID == "" {
		http.Error(w, "current city directory is unavailable", http.StatusConflict)
		return
	}
	results, err := h.worldReader.SearchLocations(r.Context(), h.worldServerID, districtID, query)
	if err != nil {
		http.Error(w, "could not search addresses", http.StatusInternalServerError)
		return
	}
	if len(results) > world.MaxSearchResults {
		results = results[:world.MaxSearchResults]
	}
	writeJSON(w, http.StatusOK, map[string]any{"results": results})
}

type transitReader interface {
	TransitDistricts(context.Context, string) ([]world.DistrictDestination, error)
}

type cityTransitReader interface {
	TransitDistrictsForCity(context.Context, string, string) ([]world.DistrictDestination, error)
}

type cityWorldMaterializer interface {
	MaterializeCity(context.Context, string, string) error
}

func (h *Handler) transitDestinations(ctx context.Context, reader transitReader, cityID string) ([]world.DistrictDestination, error) {
	if cityReader, ok := h.worldReader.(cityTransitReader); ok {
		return cityReader.TransitDistrictsForCity(ctx, h.worldServerID, cityID)
	}
	return reader.TransitDistricts(ctx, h.worldServerID)
}

func (h *Handler) transitDistricts(w http.ResponseWriter, r *http.Request) {
	current, ok := h.authorizedPlayer(w, r)
	if !ok {
		return
	}
	reader, ok := h.worldReader.(transitReader)
	if !ok {
		http.Error(w, "neighborhood directory unavailable", http.StatusServiceUnavailable)
		return
	}
	cityID := r.URL.Query().Get("cityId")
	if cityID == "" {
		locationID, err := h.store.playerLocation(r.Context(), current.ID)
		if err != nil {
			http.Error(w, "could not retrieve current location", http.StatusInternalServerError)
			return
		}
		city, exists := world.CityForLocation(locationID)
		if !exists {
			http.Error(w, "current city is unavailable", http.StatusConflict)
			return
		}
		cityID = city.ID
	}
	if _, exists := world.CityByID(cityID); !exists {
		http.Error(w, "unknown city", http.StatusBadRequest)
		return
	}
	districts, err := h.transitDestinations(r.Context(), reader, cityID)
	if err != nil {
		http.Error(w, "could not retrieve neighborhood destinations", http.StatusInternalServerError)
		return
	}
	writeJSON(w, http.StatusOK, map[string]any{"destinations": districts})
}

func (h *Handler) gateDestinations(w http.ResponseWriter, r *http.Request) {
	if !h.authorizeWorldRequest(w, r) {
		return
	}
	writeJSON(w, http.StatusOK, map[string]any{
		"catalogVersion": world.GeographyCatalogVersion,
		"destinations":   world.Cities(),
	})
}

func (h *Handler) proxyNodes(w http.ResponseWriter, r *http.Request) {
	if !h.authorizeWorldRequest(w, r) {
		return
	}
	writeJSON(w, http.StatusOK, map[string]any{
		"catalogVersion": world.GeographyCatalogVersion,
		"nodes":          world.Cities(),
	})
}

func (h *Handler) proxyConnect(w http.ResponseWriter, r *http.Request) {
	current, ok := h.authorizedPlayer(w, r)
	if !ok {
		return
	}
	var input struct {
		CityID string `json:"cityId"`
	}
	if !decodeJSON(w, r, &input) {
		return
	}
	node, exists := world.CityByID(input.CityID)
	if !exists {
		h.writeAPIError(w, r, http.StatusBadRequest, "UNKNOWN_PROXY_NODE", "The selected proxy node is not available.", nil)
		return
	}
	if materializer, ok := h.worldReader.(cityWorldMaterializer); ok {
		if err := materializer.MaterializeCity(r.Context(), h.worldServerID, node.ID); err != nil {
			h.writeAPIError(w, r, http.StatusInternalServerError, "PROXY_NODE_UNAVAILABLE", "The selected proxy node could not be prepared. Your jurisdiction has not changed.", err)
			return
		}
	}
	if err := h.store.setPlayerLocation(r.Context(), current.ID, node.ID); err != nil {
		h.writeAPIError(w, r, http.StatusInternalServerError, "PROXY_CONNECTION_FAILED", "The selected proxy node is ready, but your jurisdiction could not be updated.", err)
		return
	}
	state := h.worldAt()
	state.CityID = node.ID
	state.City = node.Name
	state.Country = node.Country
	state.TimeZone = node.TimeZone
	state.LocationID = node.ID
	state.PlaceName = ""
	writeJSON(w, http.StatusOK, map[string]any{"node": node, "world": state})
}

func (h *Handler) gateTravel(w http.ResponseWriter, r *http.Request) {
	current, ok := h.authorizedPlayer(w, r)
	if !ok {
		return
	}
	var input struct {
		CityID     string `json:"cityId"`
		DistrictID string `json:"districtId"`
	}
	if !decodeJSON(w, r, &input) {
		return
	}
	destination, exists := world.CityByID(input.CityID)
	if !exists {
		h.writeAPIError(w, r, http.StatusBadRequest, "UNKNOWN_GATE_DESTINATION", "The selected GATE destination is not available.", nil)
		return
	}
	reader, ok := h.worldReader.(transitReader)
	if !ok {
		h.writeAPIError(w, r, http.StatusServiceUnavailable, "GATE_DESTINATIONS_UNAVAILABLE", "Neighborhood destinations are temporarily unavailable.", nil)
		return
	}
	districts, err := h.transitDestinations(r.Context(), reader, destination.ID)
	if err != nil {
		h.writeAPIError(w, r, http.StatusInternalServerError, "GATE_DESTINATIONS_UNAVAILABLE", "Neighborhood destinations could not be loaded.", err)
		return
	}
	var selectedDistrict *world.DistrictDestination
	for index := range districts {
		if districts[index].ID == input.DistrictID {
			selectedDistrict = &districts[index]
			break
		}
	}
	if selectedDistrict == nil {
		h.writeAPIError(w, r, http.StatusBadRequest, "UNKNOWN_GATE_NEIGHBORHOOD", "The selected neighborhood is not part of that city.", nil)
		return
	}
	if materializer, ok := h.worldReader.(cityWorldMaterializer); ok {
		if err := materializer.MaterializeCity(r.Context(), h.worldServerID, destination.ID); err != nil {
			h.writeAPIError(w, r, http.StatusInternalServerError, "GATE_CITY_MATERIALIZATION_FAILED", "The destination could not be prepared. Your location has not changed.", err)
			return
		}
	}
	if err := h.store.setPlayerLocation(r.Context(), current.ID, destination.ID); err != nil {
		h.writeAPIError(w, r, http.StatusInternalServerError, "PLAYER_LOCATION_SAVE_FAILED", "The destination was prepared, but your location could not be updated.", err)
		return
	}
	state := h.worldAt()
	state.CityID = destination.ID
	state.City = destination.Name
	state.Country = destination.Country
	state.TimeZone = destination.TimeZone
	state.LocationID = destination.ID
	state.PlaceName = ""
	buildings := make([]world.Address, 0)
	for _, place := range selectedDistrict.Locations {
		if place.Address.Kind == "building" {
			buildings = append(buildings, place.Address)
		}
	}
	writeJSON(w, http.StatusOK, map[string]any{
		"cityId":    destination.ID,
		"district":  selectedDistrict,
		"buildings": buildings,
		"world":     state,
	})
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
	return h.worldAccount(w, r)
}

func (h *Handler) worldAccount(w http.ResponseWriter, r *http.Request) (account, bool) {
	if h.worldReader == nil || h.worldServerID == "" {
		http.Error(w, "world directory unavailable", http.StatusServiceUnavailable)
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

func (h *Handler) authorizeWorldRequest(w http.ResponseWriter, r *http.Request) bool {
	_, ok := h.worldAccount(w, r)
	return ok
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
	DisplayName string               `json:"displayName"`
	Visual      world.VisualIdentity `json:"visual"`
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
	CityID     string    `json:"cityId"`
	City       string    `json:"city"`
	Country    string    `json:"country"`
	TimeZone   string    `json:"timeZone"`
	LocationID string    `json:"locationId"`
	DistrictID string    `json:"districtId,omitempty"`
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
	gameTime := h.now()
	lyon, _ := world.CityByID("fr/lyon")
	return worldResponse{
		CityID:     lyon.ID,
		City:       lyon.Name,
		Country:    lyon.Country,
		TimeZone:   lyon.TimeZone,
		LocationID: lyon.ID,
		ServerTime: gameTime.UTC(),
	}
}

func (h *Handler) worldFor(ctx context.Context, accountID string) worldResponse {
	state := h.worldAt()
	locationID, err := h.store.playerLocation(ctx, accountID)
	if err != nil {
		return state
	}
	if city, ok := world.CityForLocation(locationID); ok {
		state.CityID = city.ID
		state.City = city.Name
		state.Country = city.Country
		state.TimeZone = city.TimeZone
		if isDistrictLocation(locationID, city.ID) {
			state.DistrictID = locationID
			state.District = districtNameFromID(locationID)
		}
		state.LocationID = city.ID
	} else {
		state.LocationID = locationID
	}
	return state
}

func districtNameFromID(districtID string) string {
	parts := strings.Split(districtID, "/")
	if len(parts) != 4 || parts[2] != "district" {
		return "Central"
	}
	words := strings.Split(parts[3], "-")
	for index, word := range words {
		if word != "" {
			words[index] = strings.ToUpper(word[:1]) + word[1:]
		}
	}
	return strings.Join(words, " ")
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
	if len([]rune(input.DisplayName)) < 2 || len([]rune(input.DisplayName)) > 40 {
		http.Error(w, "invalid character details", http.StatusBadRequest)
		return
	}
	if err := world.ValidateVisualIdentity(input.Visual); err != nil {
		http.Error(w, "invalid character appearance", http.StatusBadRequest)
		return
	}
	created, err := h.store.createCharacter(r.Context(), current.ID, input.DisplayName, input.Visual)
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

func (h *Handler) updateCharacterAppearance(w http.ResponseWriter, r *http.Request) {
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
	if existing == nil {
		http.Error(w, "character not found", http.StatusNotFound)
		return
	}
	var input struct {
		Visual world.VisualIdentity `json:"visual"`
	}
	if !decodeJSON(w, r, &input) {
		return
	}
	if err := world.ValidateVisualIdentity(input.Visual); err != nil {
		http.Error(w, "invalid character appearance", http.StatusBadRequest)
		return
	}
	updated, err := h.store.updateCharacterVisual(r.Context(), current.ID, input.Visual)
	if errors.Is(err, ErrNotFound) {
		http.Error(w, "character not found", http.StatusNotFound)
		return
	}
	if err != nil {
		http.Error(w, "could not update character appearance", http.StatusInternalServerError)
		return
	}
	writeJSON(w, http.StatusOK, identityResponse{Account: publicAccount(current), Character: &updated})
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

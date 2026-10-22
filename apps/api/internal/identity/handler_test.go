package identity

import (
	"context"
	"encoding/json"
	"errors"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
	"time"

	"noir-1974/apps/api/internal/newspaper"
	"noir-1974/apps/api/internal/records"
	"noir-1974/apps/api/internal/world"
)

type fakeStore struct {
	account          account
	accountErr       error
	createAccountErr error
	createSessionErr error
	sessionErr       error
	characterErr     error
	deleteSessionErr error
	character        *character
	locationID       string
	locationErr      error
	createdName      string
	createdVisual    world.VisualIdentity
	sessionToken     string
	sessionAccountID string
	deletedToken     string
}

type fakeDirectory struct{}

type locationDirectory struct {
	fakeDirectory
	queriedDistrict string
}

type failingCityMaterializer struct {
	fakeDirectory
	err error
}

func (reader failingCityMaterializer) MaterializeCity(context.Context, string, string) error {
	return reader.err
}

func (directory *locationDirectory) SearchLocations(_ context.Context, _, districtID, _ string) ([]world.LocationRecord, error) {
	directory.queriedDistrict = districtID
	return nil, nil
}

func (fakeDirectory) SearchNPCs(context.Context, string, string) ([]world.PersonSearchRecord, error) {
	return nil, nil
}
func (fakeDirectory) SearchCompanies(context.Context, string, string) ([]world.CompanyRecord, error) {
	return nil, nil
}
func (fakeDirectory) SearchLocations(context.Context, string, string, string) ([]world.LocationRecord, error) {
	return nil, nil
}
func (fakeDirectory) TransitDistricts(ctx context.Context, serverID string) ([]world.DistrictDestination, error) {
	return fakeDirectory{}.TransitDistrictsForCity(ctx, serverID, "fr/lyon")
}

func (fakeDirectory) TransitDistrictsForCity(_ context.Context, _, cityID string) ([]world.DistrictDestination, error) {
	city, ok := world.CityByID(cityID)
	if !ok {
		return nil, world.ErrEntityNotFound
	}
	cityKey := city.ID
	if city.ID == "fr/lyon" {
		cityKey = "lyon"
	}
	address := func(category, name, status string) world.Address {
		return world.Address{
			LocationID: "service:" + cityKey + "/" + category,
			Name:       name, Kind: "building", Status: status, District: "Central", City: city.Name, Country: city.Country,
			Full: name + ", Central, " + city.Name + ", " + city.Country,
		}
	}
	return []world.DistrictDestination{
		{ID: city.ID + "/district/central", Name: "Central", Locations: []world.LocationRecord{
			{Address: address("police-station", city.Name+" Police Station", "")},
			{Address: address("meridian-bank", "Meridian Bank", "")},
			{Address: address("gate-station", "Central GATE Station", "")},
			{Address: address("hospital", city.Name+" Hospital", "closed")},
		}},
		{ID: city.ID + "/district/bellecour", Name: "Bellecour", Locations: []world.LocationRecord{
			{Address: world.Address{LocationID: "location:bellecour-office", Name: "Bellecour Office", Kind: "building"}},
		}},
	}, nil
}

type fakeRecordReader struct {
	items []records.Record
	npcID string
	from  time.Time
	to    time.Time
}

func (reader *fakeRecordReader) NPCRecords(_ context.Context, _, npcID string, from, to time.Time, _ string) ([]records.Record, error) {
	reader.npcID, reader.from, reader.to = npcID, from, to
	return reader.items, nil
}

type fakeNews struct{}

func (fakeNews) Issue(context.Context, string, string, string, string, time.Time) (newspaper.Issue, error) {
	return newspaper.Issue{}, nil
}

func (f *fakeStore) createAccount(_ context.Context, username, email, hash string) (account, error) {
	if f.createAccountErr != nil {
		return account{}, f.createAccountErr
	}
	f.account = account{ID: "account-1", Username: username, Email: email, PasswordHash: hash}
	return f.account, nil
}

func (f *fakeStore) accountByEmail(context.Context, string) (account, error) {
	return f.account, f.accountErr
}

func (f *fakeStore) createSession(_ context.Context, accountID, token string, _ time.Time) error {
	f.sessionAccountID, f.sessionToken = accountID, token
	return f.createSessionErr
}

func (f *fakeStore) accountForSession(context.Context, string, time.Time) (account, *character, error) {
	return f.account, f.character, f.sessionErr
}

func (f *fakeStore) createCharacter(_ context.Context, _, name string, visual world.VisualIdentity) (character, error) {
	if f.characterErr != nil {
		return character{}, f.characterErr
	}
	f.createdName, f.createdVisual = name, visual
	return character{DisplayName: name, PortraitID: 1, Visual: visual}, nil
}

func (f *fakeStore) playerLocation(context.Context, string) (string, error) {
	if f.locationErr != nil {
		return "", f.locationErr
	}
	if f.locationID != "" {
		return f.locationID, nil
	}
	return "fr/lyon", nil
}
func (f *fakeStore) setPlayerLocation(_ context.Context, _ string, locationID string) error {
	f.locationID = locationID
	return nil
}

func (f *fakeStore) deleteSession(_ context.Context, token string) error {
	f.deletedToken = token
	return f.deleteSessionErr
}

func TestRegisterCreatesAccountAndHttpOnlySession(t *testing.T) {
	repository := &fakeStore{}
	handler := NewHandler(repository, true)
	request := httptest.NewRequest(http.MethodPost, "/api/auth/register", strings.NewReader(`{"username":"agent-01","email":"DETECTIVE@example.com","password":"a-long-secret"}`))
	recorder := httptest.NewRecorder()
	handler.ServeHTTP(recorder, request)

	if recorder.Code != http.StatusCreated {
		t.Fatalf("status = %d, body = %s", recorder.Code, recorder.Body.String())
	}
	if repository.account.Email != "detective@example.com" || repository.account.Username != "agent-01" || repository.sessionAccountID != "account-1" {
		t.Fatalf("account/session not created: %+v", repository)
	}
	if repository.account.PasswordHash == "a-long-secret" || !verifyPassword(repository.account.PasswordHash, "a-long-secret") {
		t.Fatal("password was not safely hashed")
	}
	cookie := recorder.Result().Cookies()[0]
	if !cookie.HttpOnly || !cookie.Secure || cookie.SameSite != http.SameSiteLaxMode {
		t.Fatalf("unexpected session cookie: %+v", cookie)
	}
	var response identityResponse
	if err := json.Unmarshal(recorder.Body.Bytes(), &response); err != nil {
		t.Fatal(err)
	}
	if response.Account.PasswordHash != "" || response.Character != nil {
		t.Fatalf("sensitive or unexpected registration response: %+v", response)
	}
}

func TestNPCRecordsRequiresSessionAndReturnsSelectedIntervalWithoutClassification(t *testing.T) {
	repository := &fakeStore{account: account{ID: "account-1"}, sessionToken: "valid", locationID: "service:lyon/meridian-bank"}
	reader := &fakeRecordReader{items: []records.Record{{
		ID: "record:1", NPCID: "npc:one", Source: "bank", OccurredAt: time.Date(1974, 11, 10, 16, 42, 0, 0, time.UTC),
		Operator: "Meridian Global Settlement Office", Description: "A private remittance cleared.", CaseID: "case:hidden", InternalRelation: "extortion_payment",
	}, {
		ID: "record:2", NPCID: "npc:one", Source: "gate", OccurredAt: time.Date(1974, 11, 10, 16, 42, 0, 0, time.UTC),
		Operator: "Meridian Global · GATE Network", Description: "A passenger entry was recorded.",
	}}}
	handler := NewHandlerWithGame(repository, false, fakeDirectory{}, fakeNews{}, nil, reader, "server-1")
	request := httptest.NewRequest(http.MethodGet, "/api/world/npcs/npc:one/records?from=1974-11-06&to=1974-11-12&source=bank", nil)
	request.AddCookie(&http.Cookie{Name: sessionCookie, Value: "valid"})
	recorder := httptest.NewRecorder()
	handler.ServeHTTP(recorder, request)
	if recorder.Code != http.StatusOK {
		t.Fatalf("status = %d, body = %s", recorder.Code, recorder.Body.String())
	}
	if reader.npcID != "npc:one" || reader.from.Format("2006-01-02") != "1974-11-06" || reader.to.Format("2006-01-02") != "1974-11-12" {
		t.Fatalf("query forwarded wrong target or interval: %+v", reader)
	}
	if strings.Contains(recorder.Body.String(), "case:hidden") || strings.Contains(recorder.Body.String(), "extortion_payment") {
		t.Fatalf("API leaked internal crime classification: %s", recorder.Body.String())
	}
	if strings.Contains(recorder.Body.String(), "record:2") || strings.Contains(recorder.Body.String(), "passenger entry") {
		t.Fatalf("bank app returned a GATE record: %s", recorder.Body.String())
	}
}

func TestBankRecordsAreAvailableBeforePhysicalVisitsAreEnabled(t *testing.T) {
	repository := &fakeStore{account: account{ID: "account-1"}, sessionToken: "valid"}
	reader := &fakeRecordReader{items: []records.Record{
		{ID: "bank:1", Source: "bank", Description: "A remittance was recorded."},
		{ID: "gate:1", Source: "gate", Description: "A GATE entry was recorded."},
	}}
	handler := NewHandlerWithGame(repository, false, fakeDirectory{}, fakeNews{}, nil, reader, "server-1")
	request := func(path string) *httptest.ResponseRecorder {
		req := httptest.NewRequest(http.MethodGet, path, nil)
		req.AddCookie(&http.Cookie{Name: sessionCookie, Value: "valid"})
		recorder := httptest.NewRecorder()
		handler.ServeHTTP(recorder, req)
		return recorder
	}

	bank := request("/api/world/npcs/npc:one/records?from=1974-11-06&to=1974-11-12&source=bank")
	if bank.Code != http.StatusOK || !strings.Contains(bank.Body.String(), "A remittance") {
		t.Fatalf("bank query without a physical visit: status=%d body=%s", bank.Code, bank.Body.String())
	}
	mixed := request("/api/world/npcs/npc:one/records?from=1974-11-06&to=1974-11-12")
	if mixed.Code != http.StatusOK || !strings.Contains(mixed.Body.String(), "bank:1") || !strings.Contains(mixed.Body.String(), "gate:1") {
		t.Fatalf("unfiltered records should include both sources: status=%d body=%s", mixed.Code, mixed.Body.String())
	}
}

func TestNeighborhoodTravelAndVisitDoNotPersistLocations(t *testing.T) {
	repository := &fakeStore{account: account{ID: "account-1"}, sessionToken: "valid"}
	handler := NewHandlerWithWorld(repository, false, fakeDirectory{}, "server-1")
	request := func(method, path, body string) *httptest.ResponseRecorder {
		req := httptest.NewRequest(method, path, strings.NewReader(body))
		req.AddCookie(&http.Cookie{Name: sessionCookie, Value: "valid"})
		recorder := httptest.NewRecorder()
		handler.ServeHTTP(recorder, req)
		return recorder
	}

	if got := request(http.MethodPost, "/api/world/travel", `{"districtId":"fr/lyon/district/bellecour"}`); got.Code != http.StatusNotFound {
		t.Fatalf("legacy neighborhood travel route status = %d, want 404", got.Code)
	}
	if got := request(http.MethodPost, "/api/world/visit", `{"locationId":"service:lyon/meridian-bank"}`); got.Code != http.StatusNotFound {
		t.Fatalf("legacy building visit route status = %d, want 404", got.Code)
	}
	if repository.locationID != "" {
		t.Fatalf("visual navigation unexpectedly persisted location %q", repository.locationID)
	}

	worldRequest := httptest.NewRequest(http.MethodGet, "/api/world", nil)
	worldRequest.AddCookie(&http.Cookie{Name: sessionCookie, Value: "valid"})
	worldResponseRecorder := httptest.NewRecorder()
	handler.ServeHTTP(worldResponseRecorder, worldRequest)
	var state worldResponse
	if err := json.Unmarshal(worldResponseRecorder.Body.Bytes(), &state); err != nil {
		t.Fatal(err)
	}
	if worldResponseRecorder.Code != http.StatusOK || state.CityID != "fr/lyon" || state.LocationID != "fr/lyon" || state.DistrictID != "" || state.District != "" || state.PlaceName != "" {
		t.Fatalf("world response should persist only the current city: status=%d state=%+v", worldResponseRecorder.Code, state)
	}
}

func TestGateCityCatalogAndInternationalTravel(t *testing.T) {
	repository := &fakeStore{account: account{ID: "account-1"}, sessionToken: "valid"}
	handler := NewHandlerWithWorld(repository, false, fakeDirectory{}, "server-1")
	request := func(method, path, body string) *httptest.ResponseRecorder {
		req := httptest.NewRequest(method, path, strings.NewReader(body))
		req.AddCookie(&http.Cookie{Name: sessionCookie, Value: "valid"})
		recorder := httptest.NewRecorder()
		handler.ServeHTTP(recorder, req)
		return recorder
	}

	catalog := request(http.MethodGet, "/api/world/gate/destinations", "")
	if catalog.Code != http.StatusOK {
		t.Fatalf("catalog status=%d body=%s", catalog.Code, catalog.Body.String())
	}
	var response struct {
		Destinations []world.WorldCity `json:"destinations"`
	}
	if err := json.Unmarshal(catalog.Body.Bytes(), &response); err != nil {
		t.Fatal(err)
	}
	if len(response.Destinations) != 100 {
		t.Fatalf("GATE returned %d city destinations, want 100", len(response.Destinations))
	}
	for _, destination := range response.Destinations {
		if strings.Contains(destination.ID, "district") || strings.Contains(destination.ID, "building") {
			t.Fatalf("GATE city catalog exposed a local destination: %+v", destination)
		}
	}

	if got := request(http.MethodPost, "/api/world/gate/travel", `{"cityId":"unknown/city"}`); got.Code != http.StatusBadRequest {
		t.Fatalf("unknown city status=%d body=%s", got.Code, got.Body.String())
	}
	if got := request(http.MethodPost, "/api/world/gate/travel", `{"cityId":"jp/tokyo","districtId":"br/rio-de-janeiro/district/central"}`); got.Code != http.StatusBadRequest {
		t.Fatalf("mismatched city and district status=%d body=%s", got.Code, got.Body.String())
	}
	arrived := request(http.MethodPost, "/api/world/gate/travel", `{"cityId":"jp/tokyo","districtId":"jp/tokyo/district/central"}`)
	if arrived.Code != http.StatusOK || repository.locationID != "jp/tokyo" {
		t.Fatalf("city travel status=%d location=%q body=%s", arrived.Code, repository.locationID, arrived.Body.String())
	}
	var travelResponse struct {
		CityID    string                    `json:"cityId"`
		District  world.DistrictDestination `json:"district"`
		Buildings []world.Address           `json:"buildings"`
		World     worldResponse             `json:"world"`
	}
	if err := json.Unmarshal(arrived.Body.Bytes(), &travelResponse); err != nil {
		t.Fatalf("travel response is invalid JSON: %v", err)
	}
	if travelResponse.CityID != "jp/tokyo" || travelResponse.District.ID != "jp/tokyo/district/central" || len(travelResponse.Buildings) != 4 {
		t.Fatalf("travel response missing selected district buildings: %+v", travelResponse)
	}
	if travelResponse.World.LocationID != "jp/tokyo" || travelResponse.World.DistrictID != "" {
		t.Fatalf("travel response should persist only city location: %+v", travelResponse.World)
	}
	for _, building := range travelResponse.Buildings {
		if building.Kind != "building" || building.City != "Tokyo" {
			t.Errorf("travel returned an invalid local building: %+v", building)
		}
	}

	worldRequest := request(http.MethodGet, "/api/world", "")
	var state worldResponse
	if err := json.Unmarshal(worldRequest.Body.Bytes(), &state); err != nil {
		t.Fatal(err)
	}
	if worldRequest.Code != http.StatusOK || state.CityID != "jp/tokyo" || state.City != "Tokyo" || state.Country != "Japan" || state.TimeZone != "Asia/Tokyo" || state.LocationID != "jp/tokyo" || state.DistrictID != "" {
		t.Fatalf("world response did not resolve city location: status=%d state=%+v", worldRequest.Code, state)
	}
}

func TestGateReturnsDistrictsForSelectedCity(t *testing.T) {
	repository := &fakeStore{account: account{ID: "account-1"}, sessionToken: "valid"}
	handler := NewHandlerWithWorld(repository, false, fakeDirectory{}, "server-1")
	request := httptest.NewRequest(http.MethodGet, "/api/world/transit?cityId=br/rio-de-janeiro", nil)
	request.AddCookie(&http.Cookie{Name: sessionCookie, Value: "valid"})
	response := httptest.NewRecorder()
	handler.ServeHTTP(response, request)
	if response.Code != http.StatusOK {
		t.Fatalf("district lookup status=%d body=%s", response.Code, response.Body.String())
	}
	var payload struct {
		Destinations []world.DistrictDestination `json:"destinations"`
	}
	if err := json.Unmarshal(response.Body.Bytes(), &payload); err != nil {
		t.Fatal(err)
	}
	if len(payload.Destinations) == 0 || payload.Destinations[0].ID != "br/rio-de-janeiro/district/central" {
		t.Fatalf("districts were not scoped to Rio de Janeiro: %+v", payload.Destinations)
	}
}

func TestGateTravelMaterializationFailureReturnsDiagnosticEnvelope(t *testing.T) {
	const session = "valid"
	materializationErr := errors.New(`persist city employment: ERROR: insert or update on table "employment_relationships" violates foreign key constraint`)
	repository := &fakeStore{account: account{ID: "account-1"}, sessionToken: session, locationID: "fr/lyon"}
	reader := failingCityMaterializer{err: materializationErr}
	handler := newHandler(repository, false, reader, "server-1", nil, nil, nil, nil, true)
	request := httptest.NewRequest(http.MethodPost, "/api/world/gate/travel", strings.NewReader(`{"cityId":"us/los-angeles","districtId":"us/los-angeles/district/central"}`))
	request.AddCookie(&http.Cookie{Name: sessionCookie, Value: session})
	response := httptest.NewRecorder()
	handler.ServeHTTP(response, request)

	var payload apiErrorResponse
	if err := json.Unmarshal(response.Body.Bytes(), &payload); err != nil {
		t.Fatalf("error response is not JSON: %v; body=%s", err, response.Body.String())
	}
	if response.Code != http.StatusInternalServerError || payload.Code != "GATE_CITY_MATERIALIZATION_FAILED" {
		t.Fatalf("status/code = %d/%q; body=%s", response.Code, payload.Code, response.Body.String())
	}
	if payload.RequestID == "" || response.Header().Get("X-Request-ID") != payload.RequestID {
		t.Fatalf("request ID missing or inconsistent: payload=%q header=%q", payload.RequestID, response.Header().Get("X-Request-ID"))
	}
	if payload.Details != materializationErr.Error() {
		t.Fatalf("diagnostic details = %q, want %q", payload.Details, materializationErr.Error())
	}
	if repository.locationID != "fr/lyon" {
		t.Fatalf("failed travel changed player location to %q", repository.locationID)
	}
}

func TestGateTravelMaterializationFailureHidesDiagnosticsWhenDisabled(t *testing.T) {
	const session = "valid"
	repository := &fakeStore{account: account{ID: "account-1"}, sessionToken: session}
	reader := failingCityMaterializer{err: errors.New("database constraint details")}
	handler := newHandler(repository, false, reader, "server-1", nil, nil, nil, nil, false)
	request := httptest.NewRequest(http.MethodPost, "/api/world/gate/travel", strings.NewReader(`{"cityId":"us/los-angeles","districtId":"us/los-angeles/district/central"}`))
	request.AddCookie(&http.Cookie{Name: sessionCookie, Value: session})
	response := httptest.NewRecorder()
	handler.ServeHTTP(response, request)

	var payload apiErrorResponse
	if err := json.Unmarshal(response.Body.Bytes(), &payload); err != nil {
		t.Fatalf("error response is not JSON: %v; body=%s", err, response.Body.String())
	}
	if payload.Details != "" || payload.RequestID == "" || payload.Code != "GATE_CITY_MATERIALIZATION_FAILED" {
		t.Fatalf("unexpected sanitized error payload: %+v", payload)
	}
}

func TestAPIErrorMiddlewareReturnsRequestReferenceForPanic(t *testing.T) {
	handler := apiErrorMiddleware(http.HandlerFunc(func(http.ResponseWriter, *http.Request) {
		panic("simulated handler panic")
	}))
	response := httptest.NewRecorder()
	handler.ServeHTTP(response, httptest.NewRequest(http.MethodGet, "/api/test", nil))

	var payload apiErrorResponse
	if err := json.Unmarshal(response.Body.Bytes(), &payload); err != nil {
		t.Fatalf("panic response is not JSON: %v; body=%s", err, response.Body.String())
	}
	if response.Code != http.StatusInternalServerError || payload.Code != "INTERNAL_SERVER_ERROR" || payload.RequestID == "" {
		t.Fatalf("unexpected panic response: status=%d payload=%+v", response.Code, payload)
	}
	if response.Header().Get("X-Request-ID") != payload.RequestID || strings.Contains(response.Body.String(), "simulated handler panic") {
		t.Fatalf("panic response has no matching reference or leaked internals: headers=%v body=%s", response.Header(), response.Body.String())
	}
}

func TestLocationSearchUsesPlayersCurrentDistrict(t *testing.T) {
	repository := &fakeStore{account: account{ID: "account-1"}, sessionToken: "valid", locationID: "service:lyon/meridian-bank"}
	directory := &locationDirectory{}
	handler := NewHandlerWithWorld(repository, false, directory, "server-1")
	search := func() *httptest.ResponseRecorder {
		request := httptest.NewRequest(http.MethodGet, "/api/world/locations?q=office", nil)
		request.AddCookie(&http.Cookie{Name: sessionCookie, Value: "valid"})
		recorder := httptest.NewRecorder()
		handler.ServeHTTP(recorder, request)
		return recorder
	}

	if got := search(); got.Code != http.StatusOK || directory.queriedDistrict != "fr/lyon/district/central" {
		t.Fatalf("search from a service point used wrong district: status=%d district=%q", got.Code, directory.queriedDistrict)
	}
	repository.locationID = "fr/lyon/district/bellecour"
	if got := search(); got.Code != http.StatusOK || directory.queriedDistrict != "fr/lyon/district/bellecour" {
		t.Fatalf("search after travel used wrong district: status=%d district=%q", got.Code, directory.queriedDistrict)
	}
}

func TestNPCRecordsRejectsMissingOrOversizedDateRange(t *testing.T) {
	repository := &fakeStore{account: account{ID: "account-1"}}
	reader := &fakeRecordReader{}
	handler := NewHandlerWithGame(repository, false, fakeDirectory{}, fakeNews{}, nil, reader, "server-1")
	for _, path := range []string{
		"/api/world/npcs/npc:one/records",
		"/api/world/npcs/npc:one/records?from=1974-01-01&to=1976-01-01",
		"/api/world/npcs/npc:one/records?from=1974-01-01&to=1974-01-08",
		"/api/world/npcs/npc:one/records?from=1974-01-01&to=1974-01-02&source=unknown",
	} {
		request := httptest.NewRequest(http.MethodGet, path, nil)
		request.AddCookie(&http.Cookie{Name: sessionCookie, Value: "valid"})
		recorder := httptest.NewRecorder()
		handler.ServeHTTP(recorder, request)
		if recorder.Code != http.StatusBadRequest {
			t.Errorf("GET %s status = %d, want 400", path, recorder.Code)
		}
	}
}

func TestRegisterRejectsInvalidAndDuplicateAccounts(t *testing.T) {
	tests := []struct {
		name string
		body string
		err  error
		want int
	}{
		{name: "invalid fields", body: `{"username":"x","email":"bad","password":"short"}`, want: http.StatusBadRequest},
		{name: "malformed body", body: `{`, want: http.StatusBadRequest},
		{name: "extra fields", body: `{"username":"detective","email":"d@example.com","password":"a-long-secret","extra":true}`, want: http.StatusBadRequest},
		{name: "duplicate", body: `{"username":"detective","email":"d@example.com","password":"a-long-secret"}`, err: ErrConflict, want: http.StatusConflict},
		{name: "database error", body: `{"username":"detective","email":"d@example.com","password":"a-long-secret"}`, err: errors.New("database unavailable"), want: http.StatusInternalServerError},
	}
	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			repository := &fakeStore{createAccountErr: test.err}
			recorder := httptest.NewRecorder()
			NewHandler(repository, false).ServeHTTP(recorder, httptest.NewRequest(http.MethodPost, "/api/auth/register", strings.NewReader(test.body)))
			if recorder.Code != test.want {
				t.Fatalf("status = %d, want %d", recorder.Code, test.want)
			}
		})
	}
}

func TestLoginAndSessionLookup(t *testing.T) {
	hash, err := hashPassword("a-long-secret")
	if err != nil {
		t.Fatal(err)
	}
	repository := &fakeStore{account: account{ID: "account-1", Username: "detective", Email: "d@example.com", PasswordHash: hash}}
	handler := NewHandler(repository, false)
	login := httptest.NewRecorder()
	handler.ServeHTTP(login, httptest.NewRequest(http.MethodPost, "/api/auth/login", strings.NewReader(`{"email":"D@example.com","password":"a-long-secret"}`)))
	if login.Code != http.StatusOK {
		t.Fatalf("login status = %d, body = %s", login.Code, login.Body.String())
	}

	cookie := login.Result().Cookies()[0]
	meRequest := httptest.NewRequest(http.MethodGet, "/api/me", nil)
	meRequest.AddCookie(cookie)
	me := httptest.NewRecorder()
	handler.ServeHTTP(me, meRequest)
	if me.Code != http.StatusOK {
		t.Fatalf("me status = %d, body = %s", me.Code, me.Body.String())
	}

	logoutRequest := httptest.NewRequest(http.MethodPost, "/api/auth/logout", nil)
	logoutRequest.AddCookie(cookie)
	logout := httptest.NewRecorder()
	handler.ServeHTTP(logout, logoutRequest)
	if logout.Code != http.StatusNoContent || repository.deletedToken == "" {
		t.Fatalf("logout failed: status=%d token=%q", logout.Code, repository.deletedToken)
	}
}

func TestLoginAndCharacterErrors(t *testing.T) {
	tests := []struct {
		name   string
		method string
		path   string
		body   string
		cookie bool
		store  *fakeStore
		want   int
	}{
		{name: "bad password", method: http.MethodPost, path: "/api/auth/login", body: `{"email":"d@example.com","password":"wrong"}`, store: &fakeStore{account: account{PasswordHash: "bad"}}, want: http.StatusUnauthorized},
		{name: "unknown account", method: http.MethodPost, path: "/api/auth/login", body: `{"email":"d@example.com","password":"secret"}`, store: &fakeStore{accountErr: ErrNotFound}, want: http.StatusUnauthorized},
		{name: "missing session", method: http.MethodGet, path: "/api/me", store: &fakeStore{}, want: http.StatusUnauthorized},
		{name: "expired session", method: http.MethodGet, path: "/api/me", cookie: true, store: &fakeStore{sessionErr: ErrNotFound}, want: http.StatusUnauthorized},
		{name: "character missing session", method: http.MethodPost, path: "/api/characters", body: `{"displayName":"Ada","visual":{"hairId":"hair_short_brown","skinId":"skin_light","eyesId":"eyes_brown"}}`, want: http.StatusUnauthorized},
		{name: "invalid character", method: http.MethodPost, path: "/api/characters", body: `{"displayName":"A","visual":{}}`, cookie: true, store: &fakeStore{}, want: http.StatusBadRequest},
		{name: "existing character", method: http.MethodPost, path: "/api/characters", body: `{"displayName":"Ada","visual":{"hairId":"hair_short_brown","skinId":"skin_light","eyesId":"eyes_brown"}}`, cookie: true, store: &fakeStore{character: &character{DisplayName: "Existing", PortraitID: 1}}, want: http.StatusConflict},
		{name: "character created", method: http.MethodPost, path: "/api/characters", body: `{"displayName":"  Ada  ","visual":{"hairId":"hair_short_brown","skinId":"skin_light","eyesId":"eyes_brown","faceId":"","beardId":"","accessoryIds":[]}}`, cookie: true, store: &fakeStore{}, want: http.StatusCreated},
		{name: "character database error", method: http.MethodPost, path: "/api/characters", body: `{"displayName":"Ada","visual":{"hairId":"hair_short_brown","skinId":"skin_light","eyesId":"eyes_brown"}}`, cookie: true, store: &fakeStore{characterErr: errors.New("database unavailable")}, want: http.StatusInternalServerError},
	}
	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			recorder := httptest.NewRecorder()
			request := httptest.NewRequest(test.method, test.path, strings.NewReader(test.body))
			if test.cookie {
				request.AddCookie(&http.Cookie{Name: sessionCookie, Value: "session-token"})
			}
			NewHandler(test.store, false).ServeHTTP(recorder, request)
			if recorder.Code != test.want {
				t.Fatalf("status = %d, want %d", recorder.Code, test.want)
			}
			if test.name == "character created" && (test.store.createdName != "Ada" || test.store.createdVisual.HairID != "hair_short_brown") {
				t.Fatalf("character input not normalized: %+v", test.store)
			}
		})
	}
}

func TestPasswordEncodingAndValidation(t *testing.T) {
	hash, err := hashPassword("correct horse")
	if err != nil {
		t.Fatal(err)
	}
	if !verifyPassword(hash, "correct horse") || verifyPassword(hash, "wrong") || verifyPassword("invalid", "correct horse") {
		t.Fatal("password verification did not enforce the encoded hash")
	}
	if validEmail("not-an-email") || validEmail(strings.Repeat("a", 250)+"@example.com") || !validEmail("detective@example.com") {
		t.Fatal("email validation returned an unexpected result")
	}
}

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
func (fakeDirectory) TransitDistricts(context.Context, string) ([]world.DistrictDestination, error) {
	return []world.DistrictDestination{
		{ID: "fr/lyon/district/central", Name: "Central", Locations: []world.LocationRecord{
			{Address: world.Address{LocationID: "service:lyon/police-station", Name: "Lyon Police Station"}},
			{Address: world.Address{LocationID: "service:lyon/meridian-bank", Name: "Meridian Bank"}},
			{Address: world.Address{LocationID: "service:lyon/gate-station", Name: "Central GATE Station"}},
			{Address: world.Address{LocationID: "location:central-office", Name: "Central Office"}},
		}},
		{ID: "fr/lyon/district/bellecour", Name: "Bellecour", Locations: []world.LocationRecord{
			{Address: world.Address{LocationID: "location:bellecour-office", Name: "Bellecour Office"}},
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

func (fakeNews) Issue(context.Context, string, string, string, time.Time) (newspaper.Issue, error) {
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
	return "fr/lyon/district/central", nil
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

func TestTravelAndVisitPersistValidatedPlayerLocation(t *testing.T) {
	repository := &fakeStore{account: account{ID: "account-1"}, sessionToken: "valid"}
	handler := NewHandlerWithWorld(repository, false, fakeDirectory{}, "server-1")
	request := func(method, path, body string) *httptest.ResponseRecorder {
		req := httptest.NewRequest(method, path, strings.NewReader(body))
		req.AddCookie(&http.Cookie{Name: sessionCookie, Value: "valid"})
		recorder := httptest.NewRecorder()
		handler.ServeHTTP(recorder, req)
		return recorder
	}

	if got := request(http.MethodPost, "/api/world/travel", `{"districtId":"fr/lyon/district/missing"}`); got.Code != http.StatusBadRequest {
		t.Fatalf("unknown travel destination status = %d, want 400", got.Code)
	}
	if got := request(http.MethodPost, "/api/world/travel", `{"districtId":"fr/lyon/district/bellecour"}`); got.Code != http.StatusOK || repository.locationID != "fr/lyon/district/bellecour" {
		t.Fatalf("travel did not persist valid district: status=%d location=%q body=%s", got.Code, repository.locationID, got.Body.String())
	}

	if got := request(http.MethodPost, "/api/world/visit", `{"locationId":"service:lyon/meridian-bank"}`); got.Code != http.StatusBadRequest || repository.locationID != "fr/lyon/district/bellecour" {
		t.Fatalf("visit allowed a building from another district: status=%d location=%q", got.Code, repository.locationID)
	}
	if got := request(http.MethodPost, "/api/world/travel", `{"districtId":"fr/lyon/district/central"}`); got.Code != http.StatusOK {
		t.Fatalf("return to Central status = %d body=%s", got.Code, got.Body.String())
	}
	if got := request(http.MethodPost, "/api/world/visit", `{"locationId":"service:lyon/meridian-bank"}`); got.Code != http.StatusOK || repository.locationID != "service:lyon/meridian-bank" {
		t.Fatalf("visit did not persist current service location: status=%d location=%q body=%s", got.Code, repository.locationID, got.Body.String())
	}

	worldRequest := httptest.NewRequest(http.MethodGet, "/api/world", nil)
	worldRequest.AddCookie(&http.Cookie{Name: sessionCookie, Value: "valid"})
	worldResponseRecorder := httptest.NewRecorder()
	handler.ServeHTTP(worldResponseRecorder, worldRequest)
	var state worldResponse
	if err := json.Unmarshal(worldResponseRecorder.Body.Bytes(), &state); err != nil {
		t.Fatal(err)
	}
	if worldResponseRecorder.Code != http.StatusOK || state.DistrictID != "fr/lyon/district/central" || state.District != "Central" || state.LocationID != "service:lyon/meridian-bank" || state.PlaceName != "Meridian Bank" {
		t.Fatalf("world response lost current place or district: status=%d state=%+v", worldResponseRecorder.Code, state)
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
		{name: "character missing session", method: http.MethodPost, path: "/api/characters", body: `{"displayName":"Ada","visual":{"hairId":"hair_short_brown","skinColorId":"skin_light","eyeColorId":"eyes_brown"}}`, want: http.StatusUnauthorized},
		{name: "invalid character", method: http.MethodPost, path: "/api/characters", body: `{"displayName":"A","visual":{}}`, cookie: true, store: &fakeStore{}, want: http.StatusBadRequest},
		{name: "existing character", method: http.MethodPost, path: "/api/characters", body: `{"displayName":"Ada","visual":{"hairId":"hair_short_brown","skinColorId":"skin_light","eyeColorId":"eyes_brown"}}`, cookie: true, store: &fakeStore{character: &character{DisplayName: "Existing", PortraitID: 1}}, want: http.StatusConflict},
		{name: "character created", method: http.MethodPost, path: "/api/characters", body: `{"displayName":"  Ada  ","visual":{"hairId":"hair_short_brown","skinColorId":"skin_light","eyeColorId":"eyes_brown","faceDetailId":"","accessoryIds":[]}}`, cookie: true, store: &fakeStore{}, want: http.StatusCreated},
		{name: "character database error", method: http.MethodPost, path: "/api/characters", body: `{"displayName":"Ada","visual":{"hairId":"hair_short_brown","skinColorId":"skin_light","eyeColorId":"eyes_brown"}}`, cookie: true, store: &fakeStore{characterErr: errors.New("database unavailable")}, want: http.StatusInternalServerError},
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

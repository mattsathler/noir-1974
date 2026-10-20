package health

import (
	"context"
	"errors"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/jackc/pgx/v5"
)

type fakeQueryer struct {
	value int
	err   error
}

func (f fakeQueryer) QueryRow(context.Context, string, ...any) pgx.Row {
	return fakeRow{value: f.value, err: f.err}
}

type fakeRow struct {
	value int
	err   error
}

func (r fakeRow) Scan(dest ...any) error {
	if r.err != nil {
		return r.err
	}
	*(dest[0].(*int)) = r.value
	return nil
}

func TestHealthHandler(t *testing.T) {
	tests := []struct {
		name       string
		queryer    fakeQueryer
		wantStatus int
		wantBody   string
	}{
		{name: "connected", queryer: fakeQueryer{value: 1}, wantStatus: http.StatusOK, wantBody: "{\"status\":\"ok\",\"database\":\"connected\"}\n"},
		{name: "unexpected query result", queryer: fakeQueryer{value: 0}, wantStatus: http.StatusServiceUnavailable, wantBody: "database unavailable\n"},
		{name: "query error", queryer: fakeQueryer{err: errors.New("query failed")}, wantStatus: http.StatusServiceUnavailable, wantBody: "database unavailable\n"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			recorder := httptest.NewRecorder()
			request := httptest.NewRequest(http.MethodGet, "/api/health", nil)
			NewHandler(tt.queryer).ServeHTTP(recorder, request)

			if recorder.Code != tt.wantStatus {
				t.Fatalf("status = %d, want %d", recorder.Code, tt.wantStatus)
			}
			if recorder.Body.String() != tt.wantBody {
				t.Fatalf("body = %q, want %q", recorder.Body.String(), tt.wantBody)
			}
		})
	}
}

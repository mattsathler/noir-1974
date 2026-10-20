package e2e

import (
	"context"
	"encoding/json"
	"errors"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/jackc/pgx/v5"
	"noir-1974/apps/api/internal/health"
)

type fakeDatabase struct{}

func (fakeDatabase) QueryRow(context.Context, string, ...any) pgx.Row { return fakeRow{} }

type fakeRow struct{}

func (fakeRow) Scan(dest ...any) error {
	if len(dest) != 1 {
		return errors.New("unexpected scan destination")
	}
	value, ok := dest[0].(*int)
	if !ok {
		return errors.New("unexpected scan type")
	}
	*value = 1
	return nil
}

func TestHealthEndpointEndToEnd(t *testing.T) {
	server := httptest.NewServer(health.NewHandler(fakeDatabase{}))
	defer server.Close()

	response, err := http.Get(server.URL + "/api/health")
	if err != nil {
		t.Fatalf("GET health: %v", err)
	}
	defer response.Body.Close()

	if response.StatusCode != http.StatusOK {
		t.Fatalf("status = %d, want %d", response.StatusCode, http.StatusOK)
	}

	var body struct {
		Status   string `json:"status"`
		Database string `json:"database"`
	}
	if err := json.NewDecoder(response.Body).Decode(&body); err != nil {
		t.Fatalf("decode response: %v", err)
	}
	if body.Status != "ok" || body.Database != "connected" {
		t.Fatalf("unexpected response: %+v", body)
	}
}

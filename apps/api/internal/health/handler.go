package health

import (
	"context"
	"encoding/json"
	"net/http"
	"time"

	"github.com/jackc/pgx/v5"
)

type queryer interface {
	QueryRow(context.Context, string, ...any) pgx.Row
}

type response struct {
	Status   string `json:"status"`
	Database string `json:"database"`
}

// NewHandler returns the health endpoint backed by a live database query.
func NewHandler(db queryer) http.Handler {
	mux := http.NewServeMux()
	mux.HandleFunc("GET /api/health", func(w http.ResponseWriter, r *http.Request) {
		ctx, cancel := context.WithTimeout(r.Context(), 3*time.Second)
		defer cancel()

		var result int
		if err := db.QueryRow(ctx, "SELECT 1").Scan(&result); err != nil || result != 1 {
			http.Error(w, "database unavailable", http.StatusServiceUnavailable)
			return
		}

		w.Header().Set("Content-Type", "application/json")
		_ = json.NewEncoder(w).Encode(response{Status: "ok", Database: "connected"})
	})
	return mux
}

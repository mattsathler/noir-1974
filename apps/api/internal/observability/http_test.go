package observability

import (
	"context"
	"errors"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/jackc/pgx/v5/pgxpool"
)

func TestPGXQueryTracerRecordsPoolAcquisitionByRoute(t *testing.T) {
	recorder := NewHTTPRecorder()
	tracer := NewPGXQueryTracer(recorder)
	ctx := context.WithValue(context.Background(), requestLabelsContextKey{}, requestLabels{
		method: "GET", route: "/api/cases",
	})
	ctx = tracer.TraceAcquireStart(ctx, nil, pgxpool.TraceAcquireStartData{})
	tracer.TraceAcquireEnd(ctx, nil, pgxpool.TraceAcquireEndData{})

	failedCtx := context.WithValue(context.Background(), requestLabelsContextKey{}, requestLabels{
		method: "GET", route: "/api/cases",
	})
	failedCtx = tracer.TraceAcquireStart(failedCtx, nil, pgxpool.TraceAcquireStartData{})
	tracer.TraceAcquireEnd(failedCtx, nil, pgxpool.TraceAcquireEndData{Err: errors.New("pool closed")})

	response := httptest.NewRecorder()
	recorder.ServeHTTP(response, httptest.NewRequest("GET", "/metrics", nil))
	metrics := response.Body.String()
	if !strings.Contains(metrics, `noir_db_pool_acquisitions_total{method="GET",route="/api/cases",status="ok"} 1`) {
		t.Fatalf("successful pool acquisition was not recorded: %s", metrics)
	}
	if !strings.Contains(metrics, `noir_db_pool_acquisitions_total{method="GET",route="/api/cases",status="error"} 1`) {
		t.Fatalf("failed pool acquisition was not recorded: %s", metrics)
	}
	if !strings.Contains(metrics, `noir_db_pool_acquire_duration_seconds_count{method="GET",route="/api/cases",status="ok"} 1`) {
		t.Fatalf("pool acquisition histogram was not recorded: %s", metrics)
	}
}

func TestCacheObserverExportsBoundedHitAndMissCounters(t *testing.T) {
	recorder := NewHTTPRecorder()
	recorder.ObserveCache("city_world_expansions", "hit")
	recorder.ObserveCache("city_world_expansions", "miss")
	recorder.ObserveCache("city_world_expansions", "unknown")

	response := httptest.NewRecorder()
	recorder.ServeHTTP(response, httptest.NewRequest("GET", "/metrics", nil))
	metrics := response.Body.String()
	if !strings.Contains(metrics, `noir_cache_requests_total{cache="city_world_expansions",outcome="hit"} 1`) {
		t.Fatalf("cache hit was not recorded: %s", metrics)
	}
	if !strings.Contains(metrics, `noir_cache_requests_total{cache="city_world_expansions",outcome="miss"} 1`) {
		t.Fatalf("cache miss was not recorded: %s", metrics)
	}
	if strings.Contains(metrics, `outcome="unknown"`) {
		t.Fatalf("unexpected cache outcome was exported: %s", metrics)
	}
}

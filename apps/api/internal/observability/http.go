package observability

import (
	"context"
	"fmt"
	"net/http"
	"sort"
	"strings"
	"sync"
	"time"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
)

var latencyBuckets = [...]float64{0.005, 0.01, 0.025, 0.05, 0.1, 0.25, 0.5, 1, 2.5, 5, 10}

type seriesKey struct {
	method string
	route  string
	status int
}

type latencySeries struct {
	count   uint64
	sum     float64
	buckets [len(latencyBuckets)]uint64
}

type queryKey struct {
	method string
	route  string
	query  string
	status string
}

type requestLabels struct {
	method string
	route  string
}

type requestLabelsContextKey struct{}
type queryTraceContextKey struct{}
type poolAcquireTraceContextKey struct{}

type queryTrace struct {
	started time.Time
	labels  requestLabels
	query   string
}

type querySeries struct {
	count   uint64
	sum     float64
	buckets [len(latencyBuckets)]uint64
}

type poolAcquireKey struct {
	method string
	route  string
	status string
}

type cacheKey struct {
	cache   string
	outcome string
}

type poolAcquireTrace struct {
	started time.Time
	labels  requestLabels
}

// HTTPRecorder stores bounded-cardinality HTTP request counts and latencies.
type HTTPRecorder struct {
	mu       sync.RWMutex
	series   map[seriesKey]latencySeries
	queries  map[queryKey]querySeries
	acquires map[poolAcquireKey]querySeries
	caches   map[cacheKey]uint64
	started  time.Time
}

func NewHTTPRecorder() *HTTPRecorder {
	return &HTTPRecorder{
		series:   make(map[seriesKey]latencySeries),
		queries:  make(map[queryKey]querySeries),
		acquires: make(map[poolAcquireKey]querySeries),
		caches:   make(map[cacheKey]uint64),
		started:  time.Now(),
	}
}

func (recorder *HTTPRecorder) ObserveCache(cache, outcome string) {
	if cache == "" || (outcome != "hit" && outcome != "miss") {
		return
	}
	recorder.mu.Lock()
	recorder.caches[cacheKey{cache: cache, outcome: outcome}]++
	recorder.mu.Unlock()
}

func (recorder *HTTPRecorder) observeAcquire(key poolAcquireKey, elapsed time.Duration) {
	seconds := elapsed.Seconds()
	recorder.mu.Lock()
	series := recorder.acquires[key]
	series.count++
	series.sum += seconds
	for index, bound := range latencyBuckets {
		if seconds <= bound {
			series.buckets[index]++
		}
	}
	recorder.acquires[key] = series
	recorder.mu.Unlock()
}

func (recorder *HTTPRecorder) Observe(method, route string, status int, elapsed time.Duration) {
	if route == "" {
		route = "unmatched"
	}
	if status == 0 {
		status = http.StatusOK
	}
	key := seriesKey{method: method, route: route, status: status}
	seconds := elapsed.Seconds()
	recorder.mu.Lock()
	series := recorder.series[key]
	series.count++
	series.sum += seconds
	for index, bound := range latencyBuckets {
		if seconds <= bound {
			series.buckets[index]++
		}
	}
	recorder.series[key] = series
	recorder.mu.Unlock()
}

func (recorder *HTTPRecorder) observeQuery(key queryKey, elapsed time.Duration) {
	seconds := elapsed.Seconds()
	recorder.mu.Lock()
	series := recorder.queries[key]
	series.count++
	series.sum += seconds
	for index, bound := range latencyBuckets {
		if seconds <= bound {
			series.buckets[index]++
		}
	}
	recorder.queries[key] = series
	recorder.mu.Unlock()
}

// Middleware records requests using a route template, never a raw URL path.
func (recorder *HTTPRecorder) Middleware(route func(*http.Request) string, next http.Handler) http.Handler {
	return http.HandlerFunc(func(writer http.ResponseWriter, request *http.Request) {
		started := time.Now()
		requestRoute := route(request)
		labels := requestLabels{method: request.Method, route: requestRoute}
		request = request.WithContext(context.WithValue(request.Context(), requestLabelsContextKey{}, labels))
		statusWriter := &responseStatusWriter{ResponseWriter: writer}
		next.ServeHTTP(statusWriter, request)
		recorder.Observe(labels.method, labels.route, statusWriter.status, time.Since(started))
	})
}

// PGXQueryTracer records query duration by HTTP route and a bounded SQL operation label.
type PGXQueryTracer struct {
	recorder *HTTPRecorder
}

func NewPGXQueryTracer(recorder *HTTPRecorder) *PGXQueryTracer {
	return &PGXQueryTracer{recorder: recorder}
}

func (tracer *PGXQueryTracer) TraceQueryStart(ctx context.Context, _ *pgx.Conn, data pgx.TraceQueryStartData) context.Context {
	labels, ok := ctx.Value(requestLabelsContextKey{}).(requestLabels)
	if !ok {
		return ctx
	}
	return context.WithValue(ctx, queryTraceContextKey{}, queryTrace{
		started: time.Now(), labels: labels, query: sqlOperation(data.SQL),
	})
}

func (tracer *PGXQueryTracer) TraceQueryEnd(ctx context.Context, _ *pgx.Conn, data pgx.TraceQueryEndData) {
	trace, ok := ctx.Value(queryTraceContextKey{}).(queryTrace)
	if !ok {
		return
	}
	status := "ok"
	if data.Err != nil {
		status = "error"
	}
	tracer.recorder.observeQuery(queryKey{
		method: trace.labels.method, route: trace.labels.route, query: trace.query, status: status,
	}, time.Since(trace.started))
}

// TraceAcquireStart captures the request label before pgxpool waits for an available connection.
func (tracer *PGXQueryTracer) TraceAcquireStart(ctx context.Context, _ *pgxpool.Pool, _ pgxpool.TraceAcquireStartData) context.Context {
	labels, ok := ctx.Value(requestLabelsContextKey{}).(requestLabels)
	if !ok {
		return ctx
	}
	return context.WithValue(ctx, poolAcquireTraceContextKey{}, poolAcquireTrace{started: time.Now(), labels: labels})
}

// TraceAcquireEnd records pool wait and connection acquisition time by HTTP route.
func (tracer *PGXQueryTracer) TraceAcquireEnd(ctx context.Context, _ *pgxpool.Pool, data pgxpool.TraceAcquireEndData) {
	trace, ok := ctx.Value(poolAcquireTraceContextKey{}).(poolAcquireTrace)
	if !ok {
		return
	}
	status := "ok"
	if data.Err != nil {
		status = "error"
	}
	tracer.recorder.observeAcquire(poolAcquireKey{method: trace.labels.method, route: trace.labels.route, status: status}, time.Since(trace.started))
}

func sqlOperation(query string) string {
	fields := strings.Fields(strings.ToLower(query))
	if len(fields) == 0 {
		return "other"
	}
	verbIndex := 0
	verb := strings.Trim(fields[verbIndex], "(;")
	if verb == "with" {
		for index, field := range fields {
			candidate := strings.Trim(field, "(;")
			if candidate == "select" || candidate == "insert" || candidate == "update" || candidate == "delete" {
				verbIndex, verb = index, candidate
				break
			}
		}
	}
	var targetIndex int
	switch verb {
	case "select", "delete":
		keyword := "from"
		for index := verbIndex + 1; index+1 < len(fields); index++ {
			if strings.Trim(fields[index], "(;") == keyword {
				targetIndex = index + 1
				break
			}
		}
	case "insert":
		for index := verbIndex + 1; index+1 < len(fields); index++ {
			if strings.Trim(fields[index], "(;") == "into" {
				targetIndex = index + 1
				break
			}
		}
	case "update":
		targetIndex = verbIndex + 1
	case "create", "alter", "drop", "truncate":
		for index := verbIndex + 1; index+1 < len(fields); index++ {
			if fields[index] == "table" || fields[index] == "index" {
				targetIndex = index + 1
				break
			}
		}
	}
	if targetIndex >= len(fields) {
		return verb + "_other"
	}
	target := strings.Trim(fields[targetIndex], "(),;\"`")
	if dot := strings.LastIndexByte(target, '.'); dot >= 0 {
		target = target[dot+1:]
	}
	for _, character := range target {
		if (character < 'a' || character > 'z') && (character < '0' || character > '9') && character != '_' {
			return verb + "_other"
		}
	}
	if target == "" {
		return verb + "_other"
	}
	return verb + "_" + target
}

func (recorder *HTTPRecorder) ServeHTTP(writer http.ResponseWriter, _ *http.Request) {
	type snapshot struct {
		key    seriesKey
		series latencySeries
	}
	type querySnapshot struct {
		key    queryKey
		series querySeries
	}
	type acquireSnapshot struct {
		key    poolAcquireKey
		series querySeries
	}
	type cacheSnapshot struct {
		key   cacheKey
		count uint64
	}
	recorder.mu.RLock()
	values := make([]snapshot, 0, len(recorder.series))
	for key, series := range recorder.series {
		values = append(values, snapshot{key: key, series: series})
	}
	queryValues := make([]querySnapshot, 0, len(recorder.queries))
	for key, series := range recorder.queries {
		queryValues = append(queryValues, querySnapshot{key: key, series: series})
	}
	acquireValues := make([]acquireSnapshot, 0, len(recorder.acquires))
	for key, series := range recorder.acquires {
		acquireValues = append(acquireValues, acquireSnapshot{key: key, series: series})
	}
	cacheValues := make([]cacheSnapshot, 0, len(recorder.caches))
	for key, count := range recorder.caches {
		cacheValues = append(cacheValues, cacheSnapshot{key: key, count: count})
	}
	recorder.mu.RUnlock()
	sort.Slice(values, func(i, j int) bool {
		a, b := values[i].key, values[j].key
		if a.route != b.route {
			return a.route < b.route
		}
		if a.method != b.method {
			return a.method < b.method
		}
		return a.status < b.status
	})
	sort.Slice(queryValues, func(i, j int) bool {
		a, b := queryValues[i].key, queryValues[j].key
		if a.route != b.route {
			return a.route < b.route
		}
		if a.method != b.method {
			return a.method < b.method
		}
		if a.query != b.query {
			return a.query < b.query
		}
		return a.status < b.status
	})
	sort.Slice(acquireValues, func(i, j int) bool {
		a, b := acquireValues[i].key, acquireValues[j].key
		if a.route != b.route {
			return a.route < b.route
		}
		if a.method != b.method {
			return a.method < b.method
		}
		return a.status < b.status
	})
	sort.Slice(cacheValues, func(i, j int) bool {
		a, b := cacheValues[i].key, cacheValues[j].key
		if a.cache != b.cache {
			return a.cache < b.cache
		}
		return a.outcome < b.outcome
	})

	writer.Header().Set("Content-Type", "text/plain; version=0.0.4; charset=utf-8")
	writer.WriteHeader(http.StatusOK)
	_, _ = fmt.Fprintln(writer, "# HELP noir_http_requests_total Total HTTP requests handled by the API.")
	_, _ = fmt.Fprintln(writer, "# TYPE noir_http_requests_total counter")
	_, _ = fmt.Fprintln(writer, "# HELP noir_http_request_duration_seconds HTTP request latency in seconds.")
	_, _ = fmt.Fprintln(writer, "# TYPE noir_http_request_duration_seconds histogram")
	_, _ = fmt.Fprintln(writer, "# HELP noir_db_queries_total Total database queries executed during HTTP requests.")
	_, _ = fmt.Fprintln(writer, "# TYPE noir_db_queries_total counter")
	_, _ = fmt.Fprintln(writer, "# HELP noir_db_query_duration_seconds Database query latency in seconds.")
	_, _ = fmt.Fprintln(writer, "# TYPE noir_db_query_duration_seconds histogram")
	_, _ = fmt.Fprintln(writer, "# HELP noir_db_pool_acquisitions_total Database pool connection acquisitions.")
	_, _ = fmt.Fprintln(writer, "# TYPE noir_db_pool_acquisitions_total counter")
	_, _ = fmt.Fprintln(writer, "# HELP noir_db_pool_acquire_duration_seconds Time spent acquiring a database pool connection in seconds.")
	_, _ = fmt.Fprintln(writer, "# TYPE noir_db_pool_acquire_duration_seconds histogram")
	_, _ = fmt.Fprintln(writer, "# HELP noir_cache_requests_total Cache lookups by cache and outcome.")
	_, _ = fmt.Fprintln(writer, "# TYPE noir_cache_requests_total counter")
	for _, value := range values {
		labels := metricLabels(value.key)
		_, _ = fmt.Fprintf(writer, "noir_http_requests_total%s %d\n", labels, value.series.count)
		for index, bound := range latencyBuckets {
			_, _ = fmt.Fprintf(writer, "noir_http_request_duration_seconds_bucket%s %d\n", withLabel(labels, "le", fmt.Sprintf("%g", bound)), value.series.buckets[index])
		}
		_, _ = fmt.Fprintf(writer, "noir_http_request_duration_seconds_bucket%s %d\n", withLabel(labels, "le", "+Inf"), value.series.count)
		_, _ = fmt.Fprintf(writer, "noir_http_request_duration_seconds_sum%s %g\n", labels, value.series.sum)
		_, _ = fmt.Fprintf(writer, "noir_http_request_duration_seconds_count%s %d\n", labels, value.series.count)
	}
	for _, value := range queryValues {
		labels := queryMetricLabels(value.key)
		_, _ = fmt.Fprintf(writer, "noir_db_queries_total%s %d\n", labels, value.series.count)
		for index, bound := range latencyBuckets {
			_, _ = fmt.Fprintf(writer, "noir_db_query_duration_seconds_bucket%s %d\n", withLabel(labels, "le", fmt.Sprintf("%g", bound)), value.series.buckets[index])
		}
		_, _ = fmt.Fprintf(writer, "noir_db_query_duration_seconds_bucket%s %d\n", withLabel(labels, "le", "+Inf"), value.series.count)
		_, _ = fmt.Fprintf(writer, "noir_db_query_duration_seconds_sum%s %g\n", labels, value.series.sum)
		_, _ = fmt.Fprintf(writer, "noir_db_query_duration_seconds_count%s %d\n", labels, value.series.count)
	}
	for _, value := range acquireValues {
		labels := poolAcquireMetricLabels(value.key)
		_, _ = fmt.Fprintf(writer, "noir_db_pool_acquisitions_total%s %d\n", labels, value.series.count)
		for index, bound := range latencyBuckets {
			_, _ = fmt.Fprintf(writer, "noir_db_pool_acquire_duration_seconds_bucket%s %d\n", withLabel(labels, "le", fmt.Sprintf("%g", bound)), value.series.buckets[index])
		}
		_, _ = fmt.Fprintf(writer, "noir_db_pool_acquire_duration_seconds_bucket%s %d\n", withLabel(labels, "le", "+Inf"), value.series.count)
		_, _ = fmt.Fprintf(writer, "noir_db_pool_acquire_duration_seconds_sum%s %g\n", labels, value.series.sum)
		_, _ = fmt.Fprintf(writer, "noir_db_pool_acquire_duration_seconds_count%s %d\n", labels, value.series.count)
	}
	for _, value := range cacheValues {
		labels := fmt.Sprintf("{cache=%q,outcome=%q}", value.key.cache, value.key.outcome)
		_, _ = fmt.Fprintf(writer, "noir_cache_requests_total%s %d\n", labels, value.count)
	}
}

func poolAcquireMetricLabels(key poolAcquireKey) string {
	return fmt.Sprintf("{method=%q,route=%q,status=%q}", key.method, key.route, key.status)
}

func metricLabels(key seriesKey) string {
	return fmt.Sprintf("{method=%q,route=%q,status=%q}", key.method, key.route, fmt.Sprint(key.status))
}

func queryMetricLabels(key queryKey) string {
	return fmt.Sprintf("{method=%q,route=%q,query=%q,status=%q}", key.method, key.route, key.query, key.status)
}

func withLabel(labels, name, value string) string {
	return strings.TrimSuffix(labels, "}") + fmt.Sprintf(",%s=%q}", name, value)
}

type responseStatusWriter struct {
	http.ResponseWriter
	status int
}

func (writer *responseStatusWriter) WriteHeader(status int) {
	if writer.status != 0 {
		return
	}
	writer.status = status
	writer.ResponseWriter.WriteHeader(status)
}

func (writer *responseStatusWriter) Write(body []byte) (int, error) {
	if writer.status == 0 {
		writer.WriteHeader(http.StatusOK)
	}
	return writer.ResponseWriter.Write(body)
}

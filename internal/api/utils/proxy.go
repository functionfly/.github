package utils

import (
	"context"
	"fmt"
	"io"
	"net/http"
	"net/textproto"
	"os"
	"strconv"
	"strings"
	"sync"
	"time"

	"github.com/functionfly/functionfly/internal/storage"
	"github.com/google/uuid"
	"github.com/gorilla/mux"
	"github.com/sirupsen/logrus"
	"github.com/functionfly/functionfly/internal/apierror"
)

type edgeContextKey int

const originalPathKey edgeContextKey = iota

// SetOriginalPath stores the original request path (before EdgeSlugMiddleware
// rewriting) in the context.
func SetOriginalPath(ctx context.Context, path string) context.Context {
	return context.WithValue(ctx, originalPathKey, path)
}

// OriginalPathFromContext returns the original request path stored by
// EdgeSlugMiddleware before rewriting. Returns empty string if not set.
func OriginalPathFromContext(ctx context.Context) string {
	if v, ok := ctx.Value(originalPathKey).(string); ok {
		return v
	}
	return ""
}

var (
	proxyTimeout time.Duration
	proxyOnce    sync.Once
	// clientPool reuses http.Client per backend URL to preserve connection pools.
	clientPool   sync.Map // map[string]*http.Client
)

func getProxyTimeout() time.Duration {
	proxyOnce.Do(func() {
		proxyTimeout = 30 * time.Second
		if v := os.Getenv("PROXY_TIMEOUT_MS"); v != "" {
			if ms, err := strconv.Atoi(v); err == nil && ms > 0 {
				proxyTimeout = time.Duration(ms) * time.Millisecond
			}
		}
		if v := os.Getenv("PROXY_TIMEOUT"); v != "" {
			if d, err := time.ParseDuration(v); err == nil {
				proxyTimeout = d
			}
		}
	})
	return proxyTimeout
}

func getProxyClient(backendURL string) *http.Client {
	if v, ok := clientPool.Load(backendURL); ok {
		return v.(*http.Client)
	}
	client := &http.Client{Timeout: getProxyTimeout()}
	actual, _ := clientPool.LoadOrStore(backendURL, client)
	return actual.(*http.Client)
}

func extractAppSlug(r *http.Request) string {
	if slug := mux.Vars(r)["appSlug"]; slug != "" {
		return slug
	}
	parts := strings.Split(strings.Trim(r.URL.Path, "/"), "/")
	if len(parts) >= 1 {
		return parts[0]
	}
	return ""
}

// ProxyResult captures the outcome of a proxy operation
type ProxyResult struct {
	Success      bool          // Whether the proxy operation succeeded
	StatusCode   int           // HTTP status code returned by backend
	LatencyMs    int64         // Response time in milliseconds
	BackendID    uuid.UUID     // ID of the backend that handled the request
	Outcome      string        // "success", "failure", "timeout", etc.
	ErrorMessage string        // Error message if failed
}

// CircuitRecorder records circuit breaker failures immediately on proxy errors.
type CircuitRecorder interface {
	RecordFailure(backendID uuid.UUID)
}

// ProxyToBackend proxies the request to the selected backend with failover support.
// If circuitRecorder is non-nil, failures are recorded immediately into the circuit breaker
// (in addition to the 5-second health monitor loop).
func ProxyToBackend(w http.ResponseWriter, r *http.Request, primaryBackend *storage.Backend, failoverBackends []*storage.Backend, requestID string, circuitRecorder CircuitRecorder) *ProxyResult {
	backends := []*storage.Backend{primaryBackend}
	backends = append(backends, failoverBackends...)

	// For non-idempotent methods, only try the primary backend
	isIdempotent := IsIdempotentMethod(r.Method)
	if !isIdempotent && len(backends) > 1 {
		backends = backends[:1]
	}

	var lastErr error
	result := &ProxyResult{
		Success:      false,
		BackendID:    primaryBackend.ID,
		Outcome:      "failure",
	}

	// Propagate W3C trace context
	traceparent := r.Header.Get("Traceparent")

	for i, backend := range backends {
		backendURL := backend.URL
		if !strings.HasSuffix(backendURL, "/") {
			backendURL += "/"
		}

		appSlug := extractAppSlug(r)

		// Use the original path (before EdgeSlugMiddleware rewrote it) so the
		// backend receives the path the client actually requested. Without this,
		// a root request "/" gets rewritten to "/{slug}/index" and the backend
		// receives "/index" instead of "/".
		origPath := OriginalPathFromContext(r.Context())
		if origPath == "" {
			origPath = r.URL.Path
		}
		// Strip the app slug prefix that EdgeSlugMiddleware added.
		trimmed := strings.TrimPrefix(origPath, "/"+appSlug+"/")
		if trimmed == origPath {
			// Prefix didn't match (direct path, not edge-rewritten); use as-is.
			trimmed = strings.TrimPrefix(origPath, "/")
		}
		targetURL := backendURL + trimmed
		if r.URL.RawQuery != "" {
			targetURL += "?" + r.URL.RawQuery
		}

		proxyReq, err := http.NewRequestWithContext(r.Context(), r.Method, targetURL, r.Body)
		if err != nil {
			lastErr = err
			continue
		}

		// Copy headers
		for key, values := range r.Header {
			for _, value := range values {
				proxyReq.Header.Add(key, value)
			}
		}

		// Strip Accept-Encoding so Go's http.Transport auto-decompresses
		// gzip responses. Without this, the Transport sees an explicit
		// Accept-Encoding and passes compressed bytes through unchanged,
		// which causes Caddy (or any upstream proxy) to see a mismatched
		// Content-Encoding/body and fail with "gzip: invalid header".
		proxyReq.Header.Del("Accept-Encoding")

		// Add FunctionFly headers
		proxyReq.Header.Set("X-FunctionFly-Request-ID", requestID)
		proxyReq.Header.Set("X-FunctionFly-App-Slug", appSlug)
		if traceparent != "" {
			proxyReq.Header.Set("Traceparent", traceparent)
		}

		// Reuse http.Client per backend (connection pooling)
		client := getProxyClient(backend.URL)

		start := time.Now()
		resp, err := client.Do(proxyReq)
		latency := time.Since(start)

		if err != nil {
			logrus.WithError(err).WithFields(logrus.Fields{
				"backend_url": backend.URL,
				"attempt":     i + 1,
				"latency_ms":  latency.Milliseconds(),
				"request_id":  requestID,
			}).Warn("Backend request failed")

			lastErr = err
			result.LatencyMs = latency.Milliseconds()
			result.BackendID = backend.ID
			result.Outcome = "backend_error"
			result.ErrorMessage = err.Error()

			// Immediate circuit breaker feedback
			if circuitRecorder != nil {
				circuitRecorder.RecordFailure(backend.ID)
			}
			continue
		}

		if resp.StatusCode >= 500 && i < len(backends)-1 {
			logrus.WithFields(logrus.Fields{
				"backend_url": backend.URL,
				"status_code": resp.StatusCode,
				"attempt":     i + 1,
				"request_id":  requestID,
			}).Warn("Backend returned 5xx, trying failover")

			resp.Body.Close()

			// Immediate circuit breaker feedback on 5xx
			if circuitRecorder != nil {
				circuitRecorder.RecordFailure(backend.ID)
			}
			continue
		}

		// Success
		result.Success = true
		result.StatusCode = resp.StatusCode
		result.LatencyMs = latency.Milliseconds()
		result.BackendID = backend.ID
		result.Outcome = "success"

		// Copy response headers, stripping hop-by-hop and encoding headers
		// that may be stale after Go's Transport auto-decompressed the body.
		for key, values := range resp.Header {
			if isHopByHopHeader(key) {
				continue
			}
			for _, value := range values {
				w.Header().Add(key, value)
			}
		}

		w.WriteHeader(resp.StatusCode)
		io.Copy(w, resp.Body)
		resp.Body.Close()

		logrus.WithFields(logrus.Fields{
			"backend_url": backend.URL,
			"status_code": resp.StatusCode,
			"latency_ms":  latency.Milliseconds(),
			"attempt":     i + 1,
			"request_id":  requestID,
		}).Info("Successfully proxied request")

		return result
	}

	// All backends failed
	result.Outcome = "all_backends_failed"
	if lastErr != nil {
		result.ErrorMessage = lastErr.Error()
	}

	logrus.WithError(lastErr).WithFields(logrus.Fields{
		"app_slug":   extractAppSlug(r),
		"backends":   len(backends),
		"request_id": requestID,
	}).Error("All backends failed")

	apierror.WriteError(w, apierror.NewServiceUnavailable("All backends unavailable"))
	return result
}

// IsIdempotentMethod checks if an HTTP method is considered idempotent for fast failover
func IsIdempotentMethod(method string) bool {
	switch method {
	case http.MethodGet, http.MethodHead, http.MethodPut, http.MethodDelete:
		return true
	default:
		return false
	}
}

// LatencyTracker maintains a rolling window of latencies for percentile computation.
type LatencyTracker struct {
	mu     sync.Mutex
	window []int64
	pos    int
	size   int
	full   bool
}

// NewLatencyTracker creates a tracker with the given window size.
func NewLatencyTracker(windowSize int) *LatencyTracker {
	if windowSize <= 0 {
		windowSize = 100
	}
	return &LatencyTracker{
		window: make([]int64, windowSize),
		size:   windowSize,
	}
}

// Record adds a latency sample to the tracker.
func (t *LatencyTracker) Record(latencyMs int64) {
	t.mu.Lock()
	defer t.mu.Unlock()
	t.window[t.pos] = latencyMs
	t.pos = (t.pos + 1) % t.size
	if t.pos == 0 {
		t.full = true
	}
}

// Percentiles returns P50, P95, P99 latency from the rolling window.
func (t *LatencyTracker) Percentiles() (p50, p95, p99 int64) {
	t.mu.Lock()
	defer t.mu.Unlock()

	count := t.size
	if !t.full {
		count = t.pos
	}
	if count == 0 {
		return 0, 0, 0
	}

	// Copy and sort
	sorted := make([]int64, count)
	copy(sorted, t.window[:count])
	sortInt64s(sorted)

	p50 = sorted[count*50/100]
	p95Idx := int(float64(count) * 0.95)
	if p95Idx >= count {
		p95Idx = count - 1
	}
	p95 = sorted[p95Idx]
	p99Idx := int(float64(count) * 0.99)
	if p99Idx >= count {
		p99Idx = count - 1
	}
	p99 = sorted[p99Idx]
	return
}

func sortInt64s(a []int64) {
	for i := 1; i < len(a); i++ {
		key := a[i]
		j := i - 1
		for j >= 0 && a[j] > key {
			a[j+1] = a[j]
			j--
		}
		a[j+1] = key
	}
}

// ErrCircuitOpen is a sentinel for circuit-breaker-gated routing decisions.
var ErrCircuitOpen = fmt.Errorf("circuit breaker open")

// hopByHopHeaders are headers that must not be forwarded by proxies (RFC 7230 §6.1)
// plus Content-Encoding/Content-Length which become stale when Go's Transport
// transparently decompresses the response body.
var hopByHopHeaders = map[string]bool{
	"Connection":          true,
	"Keep-Alive":          true,
	"Proxy-Authenticate":  true,
	"Proxy-Authorization": true,
	"Te":                  true,
	"Trailers":            true,
	"Transfer-Encoding":   true,
	"Upgrade":             true,
	"Content-Encoding":    true,
	"Content-Length":      true,
}

func isHopByHopHeader(key string) bool {
	return hopByHopHeaders[textproto.CanonicalMIMEHeaderKey(key)]
}

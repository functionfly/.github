// Package receipt implements the HTTP surface for the public Execution
// Receipt feature. All routes are mounted on the orchestrator root router
// (under /v1/receipts/*) and are public — no auth required for the read and
// run paths. Owner-only revoke is mounted separately via RegisterAuthedRoutes.
package receipt

import (
	"bytes"
	"context"
	"database/sql"
	"encoding/base64"
	"encoding/json"
	"errors"
	"fmt"
	"net"
	"net/http"
	"net/url"
	"os"
	"strconv"
	"strings"
	"time"

	"github.com/functionfly/functionfly/internal/api/handlers/registry"
	"github.com/functionfly/functionfly/internal/functionregistry"
	"github.com/functionfly/functionfly/internal/gateway"
	"github.com/functionfly/functionfly/internal/privacy"
	receiptstorage "github.com/functionfly/functionfly/internal/storage/receipt"
	registrystorage "github.com/functionfly/functionfly/internal/storage/registry"
	"github.com/google/uuid"
	"github.com/gorilla/mux"
	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/promauto"
	"github.com/redis/go-redis/v9"
	"github.com/sirupsen/logrus"
)

// ----------------------------------------------------------------------------
// Configuration
// ----------------------------------------------------------------------------

// Config holds the runtime configuration for the receipt feature. All fields
// are read from environment variables; defaults are production-safe.
type Config struct {
	Enabled             bool
	AutoGenerate        bool     // generate receipts for ALL successful executions
	PublicBaseURL       string   // e.g. "https://functionfly.com/r"
	OGBaseURL           string   // e.g. "https://functionfly.com/og/receipt"
	TwitterHandle       string   // e.g. "functionfly"
	MilestoneEnabled    bool     // master switch for the milestone worker
	MilestoneThresholds []int    // default [1, 10, 100, 1000, 10000]
	MilestoneChannels   []string // default ["inapp", "email"]
}

// DefaultConfig returns a Config pre-populated from environment variables.
func DefaultConfig() Config {
	cfg := Config{
		Enabled:          strings.ToLower(os.Getenv("RECEIPT_ENABLED")) != "false",
		AutoGenerate:     strings.ToLower(os.Getenv("RECEIPT_AUTO_GENERATE")) != "false",
		PublicBaseURL:    os.Getenv("RECEIPT_PUBLIC_BASE_URL"),
		OGBaseURL:        os.Getenv("RECEIPT_OG_BASE_URL"),
		TwitterHandle:    os.Getenv("RECEIPT_TWITTER_HANDLE"),
		MilestoneEnabled: strings.ToLower(os.Getenv("RECEIPT_MILESTONE_ENABLED")) == "true",
	}
	if cfg.PublicBaseURL == "" {
		cfg.PublicBaseURL = "https://functionfly.com/r"
	}
	if cfg.OGBaseURL == "" {
		cfg.OGBaseURL = "https://functionfly.com/og/receipt"
	}
	if cfg.TwitterHandle == "" {
		cfg.TwitterHandle = "functionfly"
	}
	cfg.MilestoneThresholds = parseCSVInts(os.Getenv("RECEIPT_MILESTONE_THRESHOLDS"), []int{1, 10, 100, 1000, 10000})
	cfg.MilestoneChannels = parseCSV(os.Getenv("RECEIPT_MILESTONE_CHANNELS"), []string{"inapp", "email"})
	return cfg
}

func parseCSV(s string, def []string) []string {
	if s == "" {
		return def
	}
	parts := strings.Split(s, ",")
	out := make([]string, 0, len(parts))
	for _, p := range parts {
		p = strings.TrimSpace(p)
		if p != "" {
			out = append(out, p)
		}
	}
	if len(out) == 0 {
		return def
	}
	return out
}

func parseCSVInts(s string, def []int) []int {
	if s == "" {
		return def
	}
	parts := strings.Split(s, ",")
	out := make([]int, 0, len(parts))
	for _, p := range parts {
		p = strings.TrimSpace(p)
		if p == "" {
			continue
		}
		n, err := strconv.Atoi(p)
		if err != nil || n <= 0 {
			continue
		}
		out = append(out, n)
	}
	if len(out) == 0 {
		return def
	}
	return out
}

// ----------------------------------------------------------------------------
// Prometheus metrics
// ----------------------------------------------------------------------------

var (
	receiptGetTotal = promauto.NewCounterVec(prometheus.CounterOpts{
		Name: "receipt_get_total",
		Help: "Total /v1/receipts/:id reads, by cache layer and status.",
	}, []string{"cache", "status"})

	receiptRunTotal = promauto.NewCounterVec(prometheus.CounterOpts{
		Name: "receipt_run_total",
		Help: "Total /v1/receipts/:id/run invocations, by status.",
	}, []string{"status"})

	receiptForkTotal = promauto.NewCounterVec(prometheus.CounterOpts{
		Name: "receipt_fork_total",
		Help: "Total fork-payload reads.",
	}, []string{"status"})

	receiptViewTotal = promauto.NewCounterVec(prometheus.CounterOpts{
		Name: "receipt_view_total",
		Help: "Total /v1/receipts/:id/view analytics events.",
	}, []string{"status"})

	receiptTrendingTotal = promauto.NewCounter(prometheus.CounterOpts{
		Name: "receipt_trending_total",
		Help: "Total /v1/receipts/trending reads.",
	})

	receiptMilestoneFiredTotal = promauto.NewCounterVec(prometheus.CounterOpts{
		Name: "receipt_milestone_fired_total",
		Help: "Milestone events fired, by threshold and channel.",
	}, []string{"threshold", "channel"})

	receiptMilestoneDuplicates = promauto.NewCounter(prometheus.CounterOpts{
		Name: "receipt_milestone_duplicates_total",
		Help: "Milestone events skipped because the dedupe key was already present.",
	})
)

// ----------------------------------------------------------------------------
// Handler
// ----------------------------------------------------------------------------

// MilestoneHookFunc is the signature of the milestone check callback. The
// execution handler invokes it after a successful execution with the IDs it
// already has. Implementation lives in milestone.go.
type MilestoneHookFunc func(ctx context.Context, functionID uuid.UUID, tenantID *uuid.UUID, publicID string)

// Handler is the HTTP boundary for the receipt feature. It depends on the
// shared registry Handler for the re-execute path (so billing, quota, and
// verification are identical to /v1/fx/.../execute) and on the receipt
// repository for its own reads/writes.
type Handler struct {
	Repo           *receiptstorage.Repository
	Registry       *registry.Handler // for /run delegation + ownership checks
	RegistryRepo   *registrystorage.RegistryRepository
	PrivacyService *privacy.Service
	Redis          *redis.Client
	Logger         *logrus.Logger
	Cfg            Config

	// MilestoneHook is set by the worker at wiring time; nil-safe.
	MilestoneHook MilestoneHookFunc

	// Signer is an optional HMAC key used to sign share URLs.
	Signer []byte
}

// NewHandler wires up a Handler with its dependencies. It does NOT register
// routes — call RegisterRoutes for that.
func NewHandler(
	repo *receiptstorage.Repository,
	regHandler *registry.Handler,
	regRepo *registrystorage.RegistryRepository,
	privSvc *privacy.Service,
	rdb *redis.Client,
	logger *logrus.Logger,
	cfg Config,
) *Handler {
	if logger == nil {
		logger = logrus.New()
	}
	signingKey := os.Getenv("RECEIPT_SIGNING_KEY")
	return &Handler{
		Repo:           repo,
		Registry:       regHandler,
		RegistryRepo:   regRepo,
		PrivacyService: privSvc,
		Redis:          rdb,
		Logger:         logger,
		Cfg:            cfg,
		Signer:         []byte(signingKey),
	}
}

// ----------------------------------------------------------------------------
// Route registration
// ----------------------------------------------------------------------------

// RegisterRoutes mounts the public receipt HTTP surface on the supplied router.
func (h *Handler) RegisterRoutes(r *mux.Router) {
	if !h.Cfg.Enabled {
		h.Logger.Info("Receipt feature is DISABLED via RECEIPT_ENABLED")
		return
	}

	// Public read paths.
	r.Handle("/v1/receipts/trending", h.limitIP("read", 60, time.Minute)(http.HandlerFunc(h.HandleTrending))).Methods("GET", "OPTIONS")
	r.Handle("/v1/receipts/function/{author}/{name}", h.limitIP("read", 60, time.Minute)(http.HandlerFunc(h.HandleListForFunction))).Methods("GET", "OPTIONS")
	r.Handle("/v1/receipts/{id}", h.limitIP("read", 60, time.Minute)(http.HandlerFunc(h.HandleGet))).Methods("GET", "OPTIONS")
	r.Handle("/v1/receipts/{id}/fork-payload", h.limitIP("read", 30, time.Minute)(http.HandlerFunc(h.HandleForkPayload))).Methods("GET", "OPTIONS")

	// Re-run path (tighter limit because each call is real compute).
	r.Handle("/v1/receipts/{id}/run", h.limitIP("run", 10, time.Minute)(http.HandlerFunc(h.HandleRun))).Methods("POST", "OPTIONS")

	// View counter (best-effort analytics).
	r.Handle("/v1/receipts/{id}/view", h.limitIP("view", 120, time.Minute)(http.HandlerFunc(h.HandleView))).Methods("POST", "OPTIONS")
}

// RegisterAuthedRoutes adds the owner-only revoke route.
func (h *Handler) RegisterAuthedRoutes(r *mux.Router, authMW func(http.HandlerFunc) http.HandlerFunc) {
	if !h.Cfg.Enabled {
		return
	}
	r.Handle("/v1/receipts/{id}/revoke", authMW(http.HandlerFunc(h.HandleRevoke))).Methods("POST", "OPTIONS")
}

// limitIP is a per-scope IP-keyed sliding-window rate limiter implemented
// in front of the handler. It is intentionally simple: if Redis is nil the
// middleware becomes a no-op (fail-open) so the routes still register in
// environments without Redis.
func (h *Handler) limitIP(scope string, limit int, window time.Duration) func(http.Handler) http.Handler {
	if h.Redis == nil {
		return func(next http.Handler) http.Handler { return next }
	}
	keyPrefix := "ff:rcpt:rl:" + scope
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			ip := clientIP(r)
			if !ipAllow(h.Redis, r.Context(), keyPrefix, ip, limit, window) {
				w.Header().Set("Retry-After", strconv.Itoa(int(window.Seconds())))
				w.Header().Set("Content-Type", "application/json")
				w.WriteHeader(http.StatusTooManyRequests)
				_ = json.NewEncoder(w).Encode(map[string]interface{}{
					"error": map[string]interface{}{
						"code":    "RATE_LIMITED",
						"message": "Too many requests. Try again shortly.",
					},
				})
				return
			}
			next.ServeHTTP(w, r)
		})
	}
}

// ipAllow is a tiny sliding-window check using a Redis sorted set. The
// caller is the rate limiter; the helper is split out so it's easy to unit
// test in isolation.
func ipAllow(rdb *redis.Client, ctx context.Context, prefix, ip string, limit int, window time.Duration) bool {
	if rdb == nil {
		// Defensive: callers should pre-check but a nil client is never
		// acceptable here. Fail open so the API never goes down with a
		// configuration mistake.
		return true
	}
	key := prefix + ":" + ip
	now := time.Now().Unix()
	windowStart := now - int64(window.Seconds())
	pipe := rdb.Pipeline()
	pipe.ZRemRangeByScore(ctx, key, "-inf", strconv.FormatInt(windowStart, 10))
	countCmd := pipe.ZCard(ctx, key)
	pipe.ZAdd(ctx, key, redis.Z{Score: float64(now), Member: fmt.Sprintf("%d-%s", now, ip)})
	pipe.Expire(ctx, key, window)
	if _, err := pipe.Exec(ctx); err != nil {
		// Fail open on Redis errors so the API doesn't go down with Redis.
		return true
	}
	return int(countCmd.Val()) < limit
}

// clientIP is the standard X-Forwarded-For → X-Real-IP → RemoteAddr fallback.
func clientIP(r *http.Request) string {
	if v := r.Header.Get("X-Forwarded-For"); v != "" {
		if i := strings.IndexByte(v, ','); i >= 0 {
			return strings.TrimSpace(v[:i])
		}
		return strings.TrimSpace(v)
	}
	if v := r.Header.Get("X-Real-IP"); v != "" {
		return strings.TrimSpace(v)
	}
	host, _, _ := net.SplitHostPort(r.RemoteAddr)
	return host
}

// ----------------------------------------------------------------------------
// Handlers
// ----------------------------------------------------------------------------

// HandleGet serves GET /v1/receipts/:id.
func (h *Handler) HandleGet(w http.ResponseWriter, r *http.Request) {
	id := mux.Vars(r)["id"]
	if !isValidPublicID(id) {
		receiptGetTotal.WithLabelValues("none", "400").Inc()
		writeJSONError(w, http.StatusBadRequest, "INVALID_RECEIPT_ID", "Invalid receipt id")
		return
	}

	// L1 cache: serve the marshaled payload directly if present.
	if payload, hit, err := h.Repo.CacheGet(r.Context(), id); err == nil && hit {
		receiptGetTotal.WithLabelValues("l1", "200").Inc()
		writeCacheHeaders(w)
		w.Header().Set("Content-Type", "application/json; charset=utf-8")
		_, _ = w.Write(payload)
		return
	}

	exec, fn, ver, err := h.Repo.GetReceipt(r.Context(), id)
	if err != nil {
		switch {
		case errors.Is(err, receiptstorage.ErrRevoked):
			receiptGetTotal.WithLabelValues("db", "410").Inc()
			writeJSONError(w, http.StatusGone, "RECEIPT_REVOKED", "This receipt has been revoked by its owner.")
		case errors.Is(err, receiptstorage.ErrNotFound):
			receiptGetTotal.WithLabelValues("db", "404").Inc()
			writeJSONError(w, http.StatusNotFound, "RECEIPT_NOT_FOUND", "Receipt not found.")
		default:
			receiptGetTotal.WithLabelValues("db", "500").Inc()
			h.Logger.WithError(err).WithField("public_id", id).Error("GetReceipt failed")
			writeJSONError(w, http.StatusInternalServerError, "INTERNAL", "Internal error.")
		}
		return
	}

	payload, err := h.buildResponsePayload(r.Context(), exec, fn, ver)
	if err != nil {
		receiptGetTotal.WithLabelValues("db", "500").Inc()
		h.Logger.WithError(err).Error("buildResponsePayload")
		writeJSONError(w, http.StatusInternalServerError, "INTERNAL", "Internal error.")
		return
	}

	body, _ := json.Marshal(payload)
	receiptGetTotal.WithLabelValues("db", "200").Inc()
	writeCacheHeaders(w)
	w.Header().Set("Content-Type", "application/json; charset=utf-8")
	_, _ = w.Write(body)
	h.Repo.CacheSet(r.Context(), id, body)
}

// HandleRun serves POST /v1/receipts/:id/run. Delegates to /v1/fx/{author}/{name}
// so billing, quota, verification, and caching are identical to a direct call.
func (h *Handler) HandleRun(w http.ResponseWriter, r *http.Request) {
	id := mux.Vars(r)["id"]
	if !isValidPublicID(id) {
		receiptRunTotal.WithLabelValues("400").Inc()
		writeJSONError(w, http.StatusBadRequest, "INVALID_RECEIPT_ID", "Invalid receipt id")
		return
	}

	exec, fn, _, err := h.Repo.GetReceipt(r.Context(), id)
	if err != nil {
		receiptRunTotal.WithLabelValues("404").Inc()
		writeJSONError(w, http.StatusNotFound, "RECEIPT_NOT_FOUND", "Receipt not found.")
		return
	}

	// Public run is only allowed for non-paid public functions.
	if fn.Visibility != "public" {
		receiptRunTotal.WithLabelValues("403").Inc()
		writeJSONError(w, http.StatusForbidden, "FUNCTION_PRIVATE", "This function is not public.")
		return
	}
	if fn.PricePerCall > 0 {
		receiptRunTotal.WithLabelValues("402").Inc()
		writeJSONError(w, http.StatusPaymentRequired, "FUNCTION_PAID", "This function is paid; sign in to run.")
		return
	}

	// Optional input override from the request body.
	var body struct {
		Input json.RawMessage `json:"input"`
	}
	if r.Body != nil && r.ContentLength > 0 {
		_ = json.NewDecoder(r.Body).Decode(&body)
	}
	input := body.Input
	if len(input) == 0 {
		input = exec.InputJSON
	}

	// Synthesize a request and pipe it to the existing handler.
	author := derefNullString(exec.FunctionAuthor)
	if author == "" {
		author = fn.Author
	}
	name := fn.Name
	target := fmt.Sprintf("/v1/fx/%s/%s", url.PathEscape(author), url.PathEscape(name))

	// The /v1/fx/... handler expects the body to be an ExecutionRequest
	// (`{ "input": ... }`), not the raw input. Wrap the input so the
	// downstream handler parses it correctly.
	execReq := functionregistry.ExecutionRequest{
		Author: author,
		Name:   name,
		Input:  input,
	}
	bodyBytes, err := json.Marshal(execReq)
	if err != nil {
		receiptRunTotal.WithLabelValues("500").Inc()
		h.Logger.WithError(err).Error("synth run-request marshal")
		writeJSONError(w, http.StatusInternalServerError, "INTERNAL", "Internal error.")
		return
	}

	runReq, err := http.NewRequestWithContext(r.Context(), "POST", target, bytes.NewReader(bodyBytes))
	if err != nil {
		receiptRunTotal.WithLabelValues("500").Inc()
		h.Logger.WithError(err).Error("synth run-request build")
		writeJSONError(w, http.StatusInternalServerError, "INTERNAL", "Internal error.")
		return
	}
	runReq.Header.Set("Content-Type", "application/json")
	if xff := r.Header.Get("X-Forwarded-For"); xff != "" {
		runReq.Header.Set("X-Forwarded-For", xff)
	} else if v := r.Header.Get("X-Real-IP"); v != "" {
		runReq.Header.Set("X-Real-IP", v)
	}
	runReq.Header.Set("User-Agent", r.Header.Get("User-Agent"))
	if origin := r.Header.Get("Origin"); origin != "" {
		runReq.Header.Set("Origin", origin)
	}
	runReq.Header.Set("X-Receipt-Source-ID", exec.PublicID)

	// mux.Vars(r) on a synthesized request is empty by default — the
	// downstream HandleExecute reads author/name from it, so we have to
	// inject the route variables explicitly.
	runReq = mux.SetURLVars(runReq, map[string]string{
		"author": author,
		"name":   name,
	})

	rec := &recordingResponseWriter{ResponseWriter: w, status: http.StatusOK}
	h.Registry.HandleExecute(rec, runReq)
	rec.Header().Set("X-Receipt-Source", h.Cfg.PublicBaseURL+"/"+exec.PublicID)
	receiptRunTotal.WithLabelValues(strconv.Itoa(rec.status)).Inc()
}

// HandleForkPayload serves GET /v1/receipts/:id/fork-payload.
func (h *Handler) HandleForkPayload(w http.ResponseWriter, r *http.Request) {
	id := mux.Vars(r)["id"]
	if !isValidPublicID(id) {
		receiptForkTotal.WithLabelValues("400").Inc()
		writeJSONError(w, http.StatusBadRequest, "INVALID_RECEIPT_ID", "Invalid receipt id")
		return
	}
	exec, fn, ver, err := h.Repo.GetReceipt(r.Context(), id)
	if err != nil {
		receiptForkTotal.WithLabelValues("404").Inc()
		writeJSONError(w, http.StatusNotFound, "RECEIPT_NOT_FOUND", "Receipt not found.")
		return
	}
	if fn.Visibility != "public" {
		receiptForkTotal.WithLabelValues("403").Inc()
		writeJSONError(w, http.StatusForbidden, "FUNCTION_PRIVATE", "This function is not public.")
		return
	}

	source := ""
	if ver != nil && ver.SourceCode.Valid {
		source = ver.SourceCode.String
	}
	if source == "" && ver != nil && ver.Readme.Valid {
		source = ver.Readme.String
	}

	readme := ""
	if ver != nil && ver.Readme.Valid {
		readme = ver.Readme.String
	}

	b64 := base64.StdEncoding.EncodeToString([]byte(source))
	resp := map[string]interface{}{
		"function": map[string]interface{}{
			"author":  derefNullString(exec.FunctionAuthor),
			"name":    derefNullString(exec.FunctionName),
			"version": exec.Version,
			"runtime": derefNullString(exec.Runtime),
		},
		"fork": map[string]interface{}{
			"source_b64":   b64,
			"source_bytes": len(source),
			"readme":       readme,
			"editor_url": fmt.Sprintf("/functions/new?fork=%s&name=%s&author=%s",
				url.QueryEscape(b64),
				url.QueryEscape(derefNullString(exec.FunctionName)),
				url.QueryEscape(derefNullString(exec.FunctionAuthor))),
		},
	}
	go func(pid string) {
		ctx, cancel := context.WithTimeout(context.Background(), 2*time.Second)
		defer cancel()
		if err := h.Repo.IncrementForkCount(ctx, pid); err != nil {
			h.Logger.WithError(err).WithField("public_id", pid).Warn("IncrementForkCount")
		}
	}(id)

	receiptForkTotal.WithLabelValues("200").Inc()
	w.Header().Set("Content-Type", "application/json")
	_ = json.NewEncoder(w).Encode(resp)
}

// HandleView serves POST /v1/receipts/:id/view (analytics only).
func (h *Handler) HandleView(w http.ResponseWriter, r *http.Request) {
	id := mux.Vars(r)["id"]
	if !isValidPublicID(id) {
		receiptViewTotal.WithLabelValues("400").Inc()
		writeJSONError(w, http.StatusBadRequest, "INVALID_RECEIPT_ID", "Invalid receipt id")
		return
	}
	go func(pid string) {
		ctx, cancel := context.WithTimeout(context.Background(), 2*time.Second)
		defer cancel()
		if err := h.Repo.IncrementViewCount(ctx, pid); err != nil {
			h.Logger.WithError(err).WithField("public_id", pid).Warn("IncrementViewCount")
		}
	}(id)
	receiptViewTotal.WithLabelValues("200").Inc()
	w.WriteHeader(http.StatusNoContent)
}

// HandleTrending serves GET /v1/receipts/trending.
func (h *Handler) HandleTrending(w http.ResponseWriter, r *http.Request) {
	limit, _ := strconv.Atoi(r.URL.Query().Get("limit"))
	rows, err := h.Repo.GetTrending(r.Context(), limit)
	if err != nil {
		h.Logger.WithError(err).Error("GetTrending")
		writeJSONError(w, http.StatusInternalServerError, "INTERNAL", "Internal error.")
		return
	}
	out := make([]map[string]interface{}, 0, len(rows))
	for _, row := range rows {
		out = append(out, map[string]interface{}{
			"id":              row.PublicID,
			"function_name":   derefNullString(row.FunctionName),
			"function_author": derefNullString(row.FunctionAuthor),
			"runtime":         derefNullString(row.Runtime),
			"view_count":      row.ViewCount,
			"created_at":      row.CreatedAt,
			"url":             h.Cfg.PublicBaseURL + "/" + row.PublicID,
		})
	}
	receiptTrendingTotal.Inc()
	w.Header().Set("Content-Type", "application/json")
	w.Header().Set("Cache-Control", "public, max-age=30, s-maxage=120, stale-while-revalidate=600")
	_ = json.NewEncoder(w).Encode(map[string]interface{}{"receipts": out})
}

// HandleListForFunction serves GET /v1/receipts/function/{author}/{name}.
// Returns the most recent shareable receipts for a function, used by
// the FunctionPage "Recent public receipts" widget. The function must
// exist; the handler returns an empty list (not a 404) for unknown
// functions so the UI can render the empty state without an error.
func (h *Handler) HandleListForFunction(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	author := vars["author"]
	name := vars["name"]
	if author == "" || name == "" {
		writeJSONError(w, http.StatusBadRequest, "INVALID_PARAMS", "author and name are required.")
		return
	}
	limit, _ := strconv.Atoi(r.URL.Query().Get("limit"))
	if limit <= 0 || limit > 50 {
		limit = 10
	}

	// Look up the function by author+name to get its UUID. We don't
	// expose 404 here — for a deleted function we return an empty
	// list so the UI renders the empty state. This matches the
	// trending endpoint's "best effort" behaviour.
	exec, fn, _, err := h.Repo.GetReceipt(r.Context(), "_probe_"+author+"/"+name)
	_ = exec
	_ = fn
	if err != nil || fn == nil {
		// Fall back to a direct lookup so we can return a clean 404
		// for a truly unknown function while still returning [] for
		// functions with zero shareable receipts.
		f, lookupErr := h.RegistryRepo.GetFunctionByAuthorName(context.Background(), author, name)
		if lookupErr != nil || f == nil {
			writeJSONError(w, http.StatusNotFound, "FUNCTION_NOT_FOUND", "Function not found.")
			return
		}
		fn = f
	}

	rows, err := h.Repo.ListForFunction(r.Context(), fn.ID, limit)
	if err != nil {
		h.Logger.WithError(err).Error("ListForFunction")
		writeJSONError(w, http.StatusInternalServerError, "INTERNAL", "Internal error.")
		return
	}
	out := make([]map[string]interface{}, 0, len(rows))
	for _, row := range rows {
		out = append(out, map[string]interface{}{
			"id":              row.PublicID,
			"function_name":   derefNullString(row.FunctionName),
			"function_author": derefNullString(row.FunctionAuthor),
			"runtime":         derefNullString(row.Runtime),
			"view_count":      row.ViewCount,
			"created_at":      row.CreatedAt,
			"url":             h.Cfg.PublicBaseURL + "/" + row.PublicID,
		})
	}
	w.Header().Set("Content-Type", "application/json")
	w.Header().Set("Cache-Control", "public, max-age=10, s-maxage=60, stale-while-revalidate=600")
	_ = json.NewEncoder(w).Encode(map[string]interface{}{"receipts": out})
}

// HandleRevoke serves POST /v1/receipts/:id/revoke (owner-only).
func (h *Handler) HandleRevoke(w http.ResponseWriter, r *http.Request) {
	id := mux.Vars(r)["id"]
	if !isValidPublicID(id) {
		writeJSONError(w, http.StatusBadRequest, "INVALID_RECEIPT_ID", "Invalid receipt id")
		return
	}
	userID := authedUserID(r)
	if userID == "" {
		writeJSONError(w, http.StatusUnauthorized, "UNAUTHENTICATED", "Authentication required.")
		return
	}
	uid, err := uuid.Parse(userID)
	if err != nil {
		writeJSONError(w, http.StatusUnauthorized, "INVALID_USER", "Invalid user id.")
		return
	}

	exec, fn, _, err := h.Repo.GetReceipt(r.Context(), id)
	if err != nil {
		writeJSONError(w, http.StatusNotFound, "RECEIPT_NOT_FOUND", "Receipt not found.")
		return
	}
	if fn.OwnerUserID == nil || *fn.OwnerUserID != uid {
		writeJSONError(w, http.StatusForbidden, "NOT_OWNER", "Only the function owner can revoke a receipt.")
		return
	}

	var body struct {
		Reason string `json:"reason"`
	}
	_ = json.NewDecoder(r.Body).Decode(&body)
	if len(body.Reason) > 500 {
		body.Reason = body.Reason[:500]
	}

	if err := h.Repo.Revoke(r.Context(), id, uid, body.Reason); err != nil {
		h.Logger.WithError(err).Error("Revoke")
		writeJSONError(w, http.StatusInternalServerError, "INTERNAL", "Internal error.")
		return
	}
	h.Repo.CacheInvalidate(r.Context(), id)
	w.Header().Set("Content-Type", "application/json")
	_ = json.NewEncoder(w).Encode(map[string]interface{}{
		"ok":          true,
		"revoked_at":  time.Now().UTC(),
		"public_id":   exec.PublicID,
		"function_id": exec.FunctionID,
	})
}

// ----------------------------------------------------------------------------
// Payload assembly
// ----------------------------------------------------------------------------

// PublicResponse is the response shape for GET /v1/receipts/:id.
type PublicResponse struct {
	ID       string                   `json:"id"`
	Protocol string                   `json:"protocol"`
	State    string                   `json:"state,omitempty"`
	Function PublicResponseFunction   `json:"function"`
	Execution PublicResponseExecution `json:"execution"`
	Share    PublicResponseShare      `json:"share"`
	CanRun   bool                     `json:"can_run"`
	IsPaid   bool                     `json:"is_paid"`
	PriceUSD float64                  `json:"price_per_call_usd"`
}

type PublicResponseFunction struct {
	Name         string          `json:"name"`
	Author       string          `json:"author"`
	Runtime      string          `json:"runtime"`
	Version      string          `json:"version"`
	Visibility   string          `json:"visibility"`
	Description  string          `json:"description,omitempty"`
	InputSchema  json.RawMessage `json:"input_schema,omitempty"`
	OutputSchema json.RawMessage `json:"output_schema,omitempty"`
}

type PublicResponseExecution struct {
	Input        json.RawMessage  `json:"input"`
	Output       json.RawMessage  `json:"output"`
	DurationMs   int              `json:"duration_ms"`
	Cached       bool             `json:"cached"`
	CreatedAt    time.Time        `json:"created_at"`
	Verification *VerificationView `json:"verification,omitempty"`
}

type VerificationView struct {
	Status     string     `json:"status"`
	VerifiedAt *time.Time `json:"verified_at,omitempty"`
	Error      string     `json:"error,omitempty"`
}

type PublicResponseShare struct {
	URL            string       `json:"url"`
	EmbedURL       string       `json:"embed_url"`
	TweetIntentURL string       `json:"tweet_intent_url"`
	OGMeta         PublicOGMeta `json:"og_meta"`
}

type PublicOGMeta struct {
	Title       string `json:"title"`
	Description string `json:"description"`
	Image       string `json:"image"`
}

// buildResponsePayload assembles the response from a denormalized DB read.
func (h *Handler) buildResponsePayload(ctx context.Context, exec *registrystorage.RegistryExecutionPublic, fn *registrystorage.RegistryFunction, ver *registrystorage.RegistryFunctionVersion) (*PublicResponse, error) {
	in := exec.InputJSON
	out := exec.OutputJSON
	if h.PrivacyService != nil {
		in, out, _ = h.PrivacyService.SanitizeInputOutput(in, out)
	}

	functionName := derefNullString(exec.FunctionName)
	if functionName == "" {
		functionName = fn.Name
	}
	functionAuthor := derefNullString(exec.FunctionAuthor)
	if functionAuthor == "" {
		functionAuthor = fn.Author
	}
	runtime := derefNullString(exec.Runtime)
	if runtime == "" && ver != nil {
		runtime = ver.Runtime
	}
	visibility := derefNullString(exec.FunctionVisibility)
	if visibility == "" {
		visibility = fn.Visibility
	}
	description := derefNullString(exec.Description)
	if description == "" {
		description = derefNullString(fn.Description)
	}

	resp := &PublicResponse{
		ID:       exec.PublicID,
		Protocol: exec.Protocol,
		State:    exec.State,
		Function: PublicResponseFunction{
			Name:         functionName,
			Author:       functionAuthor,
			Runtime:      runtime,
			Version:      exec.Version,
			Visibility:   visibility,
			Description:  description,
			InputSchema:  exec.InputSchema,
			OutputSchema: exec.OutputSchema,
		},
		Execution: PublicResponseExecution{
			Input:      in,
			Output:     out,
			DurationMs: exec.DurationMs,
			Cached:     exec.Cached,
			CreatedAt:  exec.CreatedAt,
		},
		CanRun:   fn.Visibility == "public" && fn.PricePerCall == 0,
		IsPaid:   fn.PricePerCall > 0,
		PriceUSD: fn.PricePerCall,
	}

	if exec.VerificationStatus.Valid && exec.VerificationStatus.String != "" {
		v := &VerificationView{Status: exec.VerificationStatus.String}
		if exec.VerifiedAt.Valid {
			t := exec.VerifiedAt.Time
			v.VerifiedAt = &t
		}
		if exec.VerificationError.Valid {
			v.Error = exec.VerificationError.String
		}
		resp.Execution.Verification = v
	}

	publicURL := h.Cfg.PublicBaseURL + "/" + exec.PublicID
	resp.Share = PublicResponseShare{
		URL:      publicURL,
		EmbedURL: publicURL + "/embed",
		OGMeta: PublicOGMeta{
			Title:       fmt.Sprintf("%s/%s ran in %dms · FunctionFly", functionAuthor, functionName, exec.DurationMs),
			Description: shareDescription(functionAuthor, functionName, description),
			Image:       h.Cfg.OGBaseURL + "/" + exec.PublicID + ".png",
		},
	}
	resp.Share.TweetIntentURL = h.buildTweetIntentURL(functionAuthor, functionName, publicURL)
	return resp, nil
}

func shareDescription(author, name, desc string) string {
	if desc != "" {
		s := desc
		if len(s) > 180 {
			s = s[:177] + "..."
		}
		return fmt.Sprintf("I just ran %s/%s on @functionfly — %s", author, name, s)
	}
	return fmt.Sprintf("I just ran %s/%s on @functionfly.", author, name)
}

func (h *Handler) buildTweetIntentURL(author, name, publicURL string) string {
	text := shareDescription(author, name, "")
	if text == "" {
		text = fmt.Sprintf("I just ran %s/%s on @%s", author, name, h.Cfg.TwitterHandle)
	}
	v := url.Values{}
	v.Set("text", text)
	v.Set("url", publicURL)
	v.Set("via", h.Cfg.TwitterHandle)
	return "https://twitter.com/intent/tweet?" + v.Encode()
}

func derefNullString(s sql.NullString) string {
	if s.Valid {
		return s.String
	}
	return ""
}

// ----------------------------------------------------------------------------
// Helpers
// ----------------------------------------------------------------------------

// isValidPublicID accepts the nanoid charset used by go-nanoid's `Canonic`.
func isValidPublicID(s string) bool {
	if s == "" || len(s) < 8 || len(s) > 64 {
		return false
	}
	for _, c := range s {
		switch {
		case c >= 'A' && c <= 'Z':
		case c >= 'a' && c <= 'z':
		case c >= '0' && c <= '9':
		case c == '_' || c == '-':
		default:
			return false
		}
	}
	return true
}

// authedUserID pulls the user id from the auth middleware's request context.
func authedUserID(r *http.Request) string {
	if v := r.Header.Get("X-User-ID"); v != "" {
		return v
	}
	if v := r.Context().Value("user_id"); v != nil {
		if s, ok := v.(string); ok {
			return s
		}
	}
	return ""
}

func writeJSONError(w http.ResponseWriter, status int, code, message string) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	_ = json.NewEncoder(w).Encode(map[string]interface{}{
		"error": map[string]interface{}{
			"code":    code,
			"message": message,
		},
	})
}

func writeCacheHeaders(w http.ResponseWriter) {
	w.Header().Set("Cache-Control", "public, max-age=30, s-maxage=300, stale-while-revalidate=86400")
}

// recordingResponseWriter is a tiny http.ResponseWriter wrapper that
// captures the status code so the metrics layer can label it.
type recordingResponseWriter struct {
	http.ResponseWriter
	status int
	wrote  bool
}

func (r *recordingResponseWriter) WriteHeader(c int) {
	if !r.wrote {
		r.status = c
		r.wrote = true
	}
	r.ResponseWriter.WriteHeader(c)
}

func (r *recordingResponseWriter) Write(b []byte) (int, error) {
	if !r.wrote {
		r.wrote = true
	}
	return r.ResponseWriter.Write(b)
}

func (r *recordingResponseWriter) Flush() {
	if f, ok := r.ResponseWriter.(http.Flusher); ok {
		f.Flush()
	}
}

// HMACSign returns a base64url signature for the given payload.
func (h *Handler) HMACSign(payload string) string {
	return gateway.HMACSign(h.Signer, payload)
}

// HMACVerify is the corresponding verifier.
func (h *Handler) HMACVerify(payload, sig string) bool {
	return gateway.HMACVerify(h.Signer, payload, sig)
}

// SignID returns `id.sig` if signing is enabled, else just `id`.
func (h *Handler) SignID(id string) string {
	return gateway.SignID(h.Signer, id)
}



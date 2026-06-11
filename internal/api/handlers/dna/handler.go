package dna

import (
	"encoding/json"
	"errors"
	"net/http"
	"strconv"
	"strings"
	"sync"
	"time"

	"github.com/functionfly/functionfly/internal/api/middleware"
	"github.com/functionfly/functionfly/internal/auth"
	"github.com/functionfly/functionfly/internal/dna"
	"github.com/gorilla/mux"
	"github.com/sirupsen/logrus"
)

// Handler exposes the Function DNA API over HTTP.
type Handler struct {
	svc    *dna.Service
	logger *logrus.Logger
	// Rate limiter for the manual analysis trigger endpoint
	analyzeLimiter *analyzeRateLimiter
}

// NewHandler creates a new DNA handler.
func NewHandler(svc *dna.Service, logger *logrus.Logger) *Handler {
	rateLimit := 10
	if v := os.Getenv("DNA_ANALYSIS_RATE_LIMIT"); v != "" {
		if n, err := strconv.Atoi(v); err == nil && n > 0 {
			rateLimit = n
		}
	}
	return &Handler{
		svc:            svc,
		logger:         logger,
		analyzeLimiter: newAnalyzeRateLimiter(rateLimit, time.Minute),
	}
}

// jsonOK writes a JSON 200 response.
func jsonOK(w http.ResponseWriter, data interface{}) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(data)
}

// jsonCreated writes a JSON 201 response.
func jsonCreated(w http.ResponseWriter, data interface{}) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusCreated)
	json.NewEncoder(w).Encode(data)
}

// writeError writes a structured JSON error response.
func writeError(w http.ResponseWriter, status int, code, message string) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	json.NewEncoder(w).Encode(map[string]string{
		"error":   message,
		"code":    code,
	})
}

// WriteError is the exported version of writeError for testing.
var WriteError = writeError

// requireAuth extracts and validates the JWT claims from the request.
func (h *Handler) requireAuth(w http.ResponseWriter, r *http.Request) *auth.Claims {
	claims := middleware.GetUserFromContext(r)
	if claims == nil {
		writeError(w, http.StatusUnauthorized, "UNAUTHORIZED", "authentication required")
		return nil
	}
	return claims
}

// GetProfile handles GET /v1/functions/{id}/dna
func (h *Handler) GetProfile(w http.ResponseWriter, r *http.Request) {
	claims := h.requireAuth(w, r)
	if claims == nil {
		return
	}

	functionID := mux.Vars(r)["id"]
	functionType := r.URL.Query().Get("type")
	if functionType == "" {
		functionType = "registry"
	}

	profile, err := h.svc.GetProfile(r.Context(), functionID, functionType, claims.TenantID.String())
	if err != nil {
		h.logger.WithError(err).Error("dna: get profile failed")
		writeError(w, http.StatusInternalServerError, "INTERNAL", "failed to get DNA profile")
		return
	}
	if profile.TenantID != claims.TenantID.String() {
		writeError(w, http.StatusForbidden, "FORBIDDEN", "access denied")
		return
	}

	jsonOK(w, profile)
}

// ListMutations handles GET /v1/functions/{id}/dna/mutations
func (h *Handler) ListMutations(w http.ResponseWriter, r *http.Request) {
	claims := h.requireAuth(w, r)
	if claims == nil {
		return
	}

	functionID := mux.Vars(r)["id"]
	functionType := r.URL.Query().Get("type")
	if functionType == "" {
		functionType = "registry"
	}

	if err := h.svc.CheckFunctionOwnership(r.Context(), functionID, functionType, claims.TenantID.String()); err != nil {
		writeError(w, http.StatusForbidden, "FORBIDDEN", "access denied")
		return
	}

	status := r.URL.Query().Get("status")
	limit := parseQueryInt(r, "limit", 20, 1, 100)
	offset := parseQueryInt(r, "offset", 0, 0, 10000)

	mutations, total, err := h.svc.ListMutations(r.Context(), functionID, claims.TenantID.String(), status, limit, offset)
	if err != nil {
		h.logger.WithError(err).Error("dna: list mutations failed")
		writeError(w, http.StatusInternalServerError, "INTERNAL", "failed to list mutations")
		return
	}

	jsonOK(w, map[string]interface{}{
		"mutations": mutations,
		"total":     total,
		"limit":     limit,
		"offset":    offset,
	})
}

// GetMutation handles GET /v1/functions/{id}/dna/mutations/{mutation_id}
func (h *Handler) GetMutation(w http.ResponseWriter, r *http.Request) {
	claims := h.requireAuth(w, r)
	if claims == nil {
		return
	}

	mutationID := mux.Vars(r)["mutation_id"]
	mutation, err := h.svc.GetMutation(r.Context(), mutationID)
	if err != nil {
		h.logger.WithError(err).Error("dna: get mutation failed")
		writeError(w, http.StatusInternalServerError, "INTERNAL", "failed to get mutation")
		return
	}
	if mutation == nil {
		writeError(w, http.StatusNotFound, "NOT_FOUND", "mutation not found")
		return
	}
	if mutation.TenantID != claims.TenantID.String() {
		writeError(w, http.StatusForbidden, "FORBIDDEN", "access denied")
		return
	}

	jsonOK(w, mutation)
}

// AcceptVariant handles POST /v1/functions/{id}/dna/variants/{mutation_id}/accept
func (h *Handler) AcceptVariant(w http.ResponseWriter, r *http.Request) {
	claims := h.requireAuth(w, r)
	if claims == nil {
		return
	}

	mutationID := mux.Vars(r)["mutation_id"]

	var req struct {
		CanaryPercentage int `json:"canary_percentage"`
	}
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		req.CanaryPercentage = 10
	}
	if req.CanaryPercentage <= 0 || req.CanaryPercentage > 100 {
		req.CanaryPercentage = 10
	}

	if err := h.svc.AcceptMutation(r.Context(), mutationID, claims.UserID.String(), claims.TenantID.String(), req.CanaryPercentage); err != nil {
		if errors.Is(err, dna.ErrAccessDenied) {
			writeError(w, http.StatusForbidden, "FORBIDDEN", "access denied")
			return
		}
		if errors.Is(err, dna.ErrMutationNotFound) {
			writeError(w, http.StatusNotFound, "NOT_FOUND", "mutation not found")
			return
		}
		if errors.Is(err, dna.ErrMutationNotProposed) {
			writeError(w, http.StatusConflict, "CONFLICT", "mutation has already been actioned")
			return
		}
		if errors.Is(err, dna.ErrInsufficientCredits) {
			writeError(w, http.StatusPaymentRequired, "INSUFFICIENT_CREDITS", "insufficient credits to accept this mutation")
			return
		}
		h.logger.WithError(err).Error("dna: accept mutation failed")
		writeError(w, http.StatusInternalServerError, "INTERNAL", "failed to accept mutation")
		return
	}

	jsonOK(w, map[string]interface{}{
		"mutation_id":       mutationID,
		"status":            "accepted",
		"canary_percentage": req.CanaryPercentage,
	})
}

// RejectVariant handles POST /v1/functions/{id}/dna/variants/{mutation_id}/reject
func (h *Handler) RejectVariant(w http.ResponseWriter, r *http.Request) {
	claims := h.requireAuth(w, r)
	if claims == nil {
		return
	}

	mutationID := mux.Vars(r)["mutation_id"]

	var req struct {
		Reason string `json:"reason"`
	}
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		writeError(w, http.StatusBadRequest, "BAD_REQUEST", "invalid request body")
		return
	}

	if err := h.svc.RejectMutation(r.Context(), mutationID, claims.TenantID.String(), req.Reason); err != nil {
		if errors.Is(err, dna.ErrAccessDenied) {
			writeError(w, http.StatusForbidden, "FORBIDDEN", "access denied")
			return
		}
		if errors.Is(err, dna.ErrMutationNotFound) {
			writeError(w, http.StatusNotFound, "NOT_FOUND", "mutation not found")
			return
		}
		if errors.Is(err, dna.ErrMutationNotProposed) {
			writeError(w, http.StatusConflict, "CONFLICT", "mutation has already been actioned")
			return
		}
		h.logger.WithError(err).Error("dna: reject mutation failed")
		writeError(w, http.StatusInternalServerError, "INTERNAL", "failed to reject mutation")
		return
	}

	jsonOK(w, map[string]interface{}{
		"mutation_id": mutationID,
		"status":      "rejected",
	})
}

// RollbackVariant handles POST /v1/functions/{id}/dna/variants/{mutation_id}/rollback
func (h *Handler) RollbackVariant(w http.ResponseWriter, r *http.Request) {
	claims := h.requireAuth(w, r)
	if claims == nil {
		return
	}

	mutationID := mux.Vars(r)["mutation_id"]

	var req struct {
		Reason string `json:"reason"`
	}
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		writeError(w, http.StatusBadRequest, "BAD_REQUEST", "invalid request body")
		return
	}

	if err := h.svc.RollbackMutation(r.Context(), mutationID, claims.TenantID.String(), req.Reason); err != nil {
		if errors.Is(err, dna.ErrAccessDenied) {
			writeError(w, http.StatusForbidden, "FORBIDDEN", "access denied")
			return
		}
		if errors.Is(err, dna.ErrMutationNotFound) {
			writeError(w, http.StatusNotFound, "NOT_FOUND", "mutation not found")
			return
		}
		if errors.Is(err, dna.ErrMutationNotRollbackable) {
			writeError(w, http.StatusConflict, "CONFLICT", "mutation is not in a rollback-eligible status")
			return
		}
		h.logger.WithError(err).Error("dna: rollback mutation failed")
		writeError(w, http.StatusInternalServerError, "INTERNAL", "failed to rollback mutation")
		return
	}

	jsonOK(w, map[string]interface{}{
		"mutation_id": mutationID,
		"status":     "rolled_back",
	})
}


// GetInsights handles GET /v1/functions/{id}/dna/insights
func (h *Handler) GetInsights(w http.ResponseWriter, r *http.Request) {
	claims := h.requireAuth(w, r)
	if claims == nil {
		return
	}

	functionID := mux.Vars(r)["id"]
	functionType := r.URL.Query().Get("type")
	if functionType == "" {
		functionType = "registry"
	}

	if err := h.svc.CheckFunctionOwnership(r.Context(), functionID, functionType, claims.TenantID.String()); err != nil {
		writeError(w, http.StatusForbidden, "FORBIDDEN", "access denied")
		return
	}

	period := r.URL.Query().Get("period")
	if period == "" {
		period = "30d"
	}

	insights, err := h.svc.GetInsights(r.Context(), functionID, claims.TenantID.String(), period)
	if err != nil {
		h.logger.WithError(err).Error("dna: get insights failed")
		writeError(w, http.StatusInternalServerError, "INTERNAL", "failed to get insights")
		return
	}

	jsonOK(w, insights)
}

// TriggerAnalysis handles POST /v1/functions/{id}/dna/analyze
func (h *Handler) TriggerAnalysis(w http.ResponseWriter, r *http.Request) {
	claims := h.requireAuth(w, r)
	if claims == nil {
		return
	}

	// Rate limit manual analysis triggers
	if !h.analyzeLimiter.Allow(claims.UserID.String()) {
		writeError(w, http.StatusTooManyRequests, "RATE_LIMITED", "too many analysis requests, please try again later")
		return
	}

	functionID := mux.Vars(r)["id"]
	functionType := r.URL.Query().Get("type")
	if functionType == "" {
		functionType = "registry"
	}

	if err := h.svc.CheckFunctionOwnership(r.Context(), functionID, functionType, claims.TenantID.String()); err != nil {
		if errors.Is(err, dna.ErrAccessDenied) {
			writeError(w, http.StatusForbidden, "FORBIDDEN", "access denied")
			return
		}
		h.logger.WithError(err).Error("dna: check ownership failed")
		writeError(w, http.StatusInternalServerError, "INTERNAL", "failed to check function ownership")
		return
	}

	var req struct {
		Priority int `json:"priority"`
	}
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil || req.Priority <= 0 {
		req.Priority = 1 // default priority for manual triggers
	}

	if err := h.svc.TriggerAnalysis(r.Context(), functionID, functionType, claims.TenantID.String(), req.Priority); err != nil {
		h.logger.WithError(err).Error("dna: trigger analysis failed")
		writeError(w, http.StatusInternalServerError, "INTERNAL", "failed to queue analysis")
		return
	}

	jsonOK(w, map[string]interface{}{
		"status":  "queued",
		"message": "analysis queued for processing",
	})
}

// ToggleEvolution handles POST /v1/functions/{id}/dna/evolution
func (h *Handler) ToggleEvolution(w http.ResponseWriter, r *http.Request) {
	claims := h.requireAuth(w, r)
	if claims == nil {
		return
	}

	functionID := mux.Vars(r)["id"]
	functionType := r.URL.Query().Get("type")
	if functionType == "" {
		functionType = "registry"
	}

	if err := h.svc.CheckFunctionOwnership(r.Context(), functionID, functionType, claims.TenantID.String()); err != nil {
		writeError(w, http.StatusForbidden, "FORBIDDEN", "access denied")
		return
	}

	var req struct {
		Enabled bool `json:"enabled"`
	}
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		writeError(w, http.StatusBadRequest, "BAD_REQUEST", "invalid request body")
		return
	}

	if err := h.svc.SetEvolutionEnabled(r.Context(), functionID, functionType, req.Enabled); err != nil {
		h.logger.WithError(err).Error("dna: toggle evolution failed")
		writeError(w, http.StatusInternalServerError, "INTERNAL", "failed to toggle evolution")
		return
	}

	jsonOK(w, map[string]interface{}{
		"evolution_enabled": req.Enabled,
	})
}

// GetEnterpriseInsights handles GET /v1/dna/enterprise/insights
func (h *Handler) GetEnterpriseInsights(w http.ResponseWriter, r *http.Request) {
	claims := h.requireAuth(w, r)
	if claims == nil {
		return
	}

	period := r.URL.Query().Get("period")
	if period == "" {
		period = "30d"
	}

	insights, err := h.svc.GetEnterpriseInsights(r.Context(), claims.TenantID.String(), period)
	if err != nil {
		h.logger.WithError(err).Error("dna: enterprise insights failed")
		writeError(w, http.StatusInternalServerError, "INTERNAL", "failed to get enterprise insights")
		return
	}

	jsonOK(w, insights)
}

// VerifyHash handles GET /v1/functions/{id}/dna/verify
func (h *Handler) VerifyHash(w http.ResponseWriter, r *http.Request) {
	claims := h.requireAuth(w, r)
	if claims == nil {
		return
	}

	functionID := mux.Vars(r)["id"]
	functionType := r.URL.Query().Get("type")
	if functionType == "" {
		functionType = "registry"
	}

	if err := h.svc.CheckFunctionOwnership(r.Context(), functionID, functionType, claims.TenantID.String()); err != nil {
		writeError(w, http.StatusForbidden, "FORBIDDEN", "access denied")
		return
	}

	matches, storedHash, computedHash, err := h.svc.VerifyDNAHash(r.Context(), functionID, functionType)
	if err != nil {
		h.logger.WithError(err).Error("dna: hash verification failed")
		writeError(w, http.StatusInternalServerError, "INTERNAL", "failed to verify DNA hash")
		return
	}

	jsonOK(w, map[string]interface{}{
		"function_id":   functionID,
		"matches":       matches,
		"stored_hash":   storedHash,
		"computed_hash": computedHash,
	})
}

// parseQueryInt extracts an integer query parameter with defaults and bounds.
func parseQueryInt(r *http.Request, key string, defaultVal, min, max int) int {
	s := r.URL.Query().Get(key)
	if s == "" {
		return defaultVal
	}
	v, err := strconv.Atoi(s)
	if err != nil || v < min || v > max {
		return defaultVal
	}
	return v
}

// analyzeRateLimiter is a per-user in-memory sliding window rate limiter with TTL-based cleanup.
type analyzeRateLimiter struct {
	mu       sync.Mutex
	entries  map[string][]time.Time
	limit    int
	window   time.Duration
	ttl      time.Duration // max time an entry can live without activity
	stopCh   chan struct{}
}

// Stop gracefully stops the background cleanup goroutine.
func (l *analyzeRateLimiter) Stop() {
	close(l.stopCh)
}

func newAnalyzeRateLimiter(limit int, window time.Duration) *analyzeRateLimiter {
	l := &analyzeRateLimiter{
		entries: make(map[string][]time.Time),
		limit:   limit,
		window:  window,
		ttl:     10 * time.Minute, // entries expire after 10 minutes of inactivity
		stopCh:  make(chan struct{}),
	}
	go l.cleanupLoop()
	return l
}

func (l *analyzeRateLimiter) cleanupLoop() {
	ticker := time.NewTicker(1 * time.Minute)
	defer ticker.Stop()
	for {
		select {
		case <-l.stopCh:
			return
		case <-ticker.C:
			l.cleanup()
		}
	}
}

func (l *analyzeRateLimiter) cleanup() {
	l.mu.Lock()
	defer l.mu.Unlock()

	now := time.Now()
	windowCutoff := now.Add(-l.window)
	ttlCutoff := now.Add(-l.ttl)

	for userID, timestamps := range l.entries {
		valid := timestamps[:0]
		for _, t := range timestamps {
			if t.After(windowCutoff) {
				valid = append(valid, t)
			}
		}
		// Remove entry if no valid timestamps OR all timestamps are older than TTL
		if len(valid) == 0 {
			delete(l.entries, userID)
		} else if len(timestamps) > 0 && timestamps[0].Before(ttlCutoff) && len(valid) == 0 {
			// All timestamps expired beyond TTL window
			delete(l.entries, userID)
		} else {
			l.entries[userID] = valid
		}
	}
}

func (l *analyzeRateLimiter) Allow(userID string) bool {
	l.mu.Lock()
	defer l.mu.Unlock()

	now := time.Now()
	windowCutoff := now.Add(-l.window)
	ttlCutoff := now.Add(-l.ttl)

	timestamps := l.entries[userID]
	valid := timestamps[:0]
	hasRecentActivity := false

	for _, t := range timestamps {
		if t.After(windowCutoff) {
			valid = append(valid, t)
			if t.After(ttlCutoff) {
				hasRecentActivity = true
			}
		}
	}

	// If no recent activity and entry exists, reset it
	if len(timestamps) > 0 && !hasRecentActivity {
		valid = timestamps[:0]
	}

	if len(valid) >= l.limit {
		l.entries[userID] = valid
		return false
	}

	l.entries[userID] = append(valid, now)
	return true
}

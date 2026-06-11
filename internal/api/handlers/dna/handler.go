package dna

import (
	"context"
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net/http"
	"strconv"
	"sync"
	"time"

	"github.com/functionfly/functionfly/internal/api/middleware"
	"github.com/functionfly/functionfly/internal/auth"
	"github.com/functionfly/functionfly/internal/dna"
	"github.com/gorilla/mux"
	"github.com/redis/go-redis/v9"
	"github.com/sirupsen/logrus"
)

const (
	analyzeRateLimitKeyPrefix = "ratelimit:dna:analyze:"
	analyzeRateLimitWindow   = 1 * time.Minute
	analyzeRateLimitMaxReqs  = 10
)

// Handler exposes the Function DNA API over HTTP.
type Handler struct {
	svc           *dna.Service
	logger        *logrus.Logger
	redisClient   *redis.Client
	analyzeLimiter *analyzeRateLimiter
}

// NewHandler creates a new DNA handler.
func NewHandler(svc *dna.Service, logger *logrus.Logger) *Handler {
	return &Handler{
		svc:            svc,
		logger:         logger,
		analyzeLimiter: newAnalyzeRateLimiter(analyzeRateLimitMaxReqs, analyzeRateLimitWindow),
	}
}

// NewHandlerWithRedis creates a new DNA handler with Redis-based distributed rate limiting.
func NewHandlerWithRedis(svc *dna.Service, logger *logrus.Logger, redisClient *redis.Client) *Handler {
	limiter := newAnalyzeRateLimiter(analyzeRateLimitMaxReqs, analyzeRateLimitWindow)
	if redisClient != nil {
		limiter.redisClient = redisClient
	}
	return &Handler{
		svc:            svc,
		logger:         logger,
		redisClient:    redisClient,
		analyzeLimiter: limiter,
	}
}

// Shutdown gracefully stops the handler's background goroutines (e.g., rate limiter cleanup).
func (h *Handler) Shutdown() {
	if h.analyzeLimiter != nil {
		h.analyzeLimiter.Stop()
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

	mutations, total, err := h.svc.ListMutations(r.Context(), functionID, status, limit, offset)
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

	// Enforce body size limit to prevent memory exhaustion
	if r.ContentLength > 4096 {
		writeError(w, http.StatusRequestEntityTooLarge, "PAYLOAD_TOO_LARGE", "request body too large")
		return
	}

	mutationID := mux.Vars(r)["mutation_id"]

	var req struct {
		CanaryPercentage int `json:"canary_percentage"`
	}
	if err := json.NewDecoder(io.LimitReader(r.Body, 4096)).Decode(&req); err != nil {
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

	// Enforce body size limit to prevent memory exhaustion
	if r.ContentLength > 4096 {
		writeError(w, http.StatusRequestEntityTooLarge, "PAYLOAD_TOO_LARGE", "request body too large")
		return
	}

	mutationID := mux.Vars(r)["mutation_id"]

	var req struct {
		Reason string `json:"reason"`
	}
	if err := json.NewDecoder(io.LimitReader(r.Body, 4096)).Decode(&req); err != nil {
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

	// Enforce body size limit to prevent memory exhaustion
	if r.ContentLength > 4096 {
		writeError(w, http.StatusRequestEntityTooLarge, "PAYLOAD_TOO_LARGE", "request body too large")
		return
	}

	mutationID := mux.Vars(r)["mutation_id"]

	var req struct {
		Reason string `json:"reason"`
	}
	if err := json.NewDecoder(io.LimitReader(r.Body, 4096)).Decode(&req); err != nil {
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
		if errors.Is(err, dna.ErrRollbackNotEligible) {
			writeError(w, http.StatusConflict, "CONFLICT", "mutation cannot be rolled back in current status")
			return
		}
		h.logger.WithError(err).Error("dna: rollback mutation failed")
		writeError(w, http.StatusInternalServerError, "INTERNAL", "failed to rollback mutation")
		return
	}

	jsonOK(w, map[string]interface{}{
		"mutation_id": mutationID,
		"status":      "rolled_back",
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

	insights, err := h.svc.GetInsights(r.Context(), functionID, period)
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
	if !h.analyzeLimiter.Allow(r.Context(), claims.UserID.String()) {
		writeError(w, http.StatusTooManyRequests, "RATE_LIMITED", "too many analysis requests, please try again later")
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

	if err := h.svc.TriggerAnalysis(r.Context(), functionID, functionType, claims.TenantID.String()); err != nil {
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

	// Enforce body size limit to prevent memory exhaustion
	if r.ContentLength > 4096 {
		writeError(w, http.StatusRequestEntityTooLarge, "PAYLOAD_TOO_LARGE", "request body too large")
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
	if err := json.NewDecoder(io.LimitReader(r.Body, 4096)).Decode(&req); err != nil {
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

// analyzeRateLimiter is a sliding window rate limiter that uses Redis when available,
// falling back to in-memory storage for single-instance deployments.
type analyzeRateLimiter struct {
	redisClient *redis.Client
	memStore    *inMemoryStore
	limit       int
	window      time.Duration
	stopCh      chan struct{}
}

// inMemoryStore provides thread-safe in-memory storage for rate limiting.
type inMemoryStore struct {
	mu    sync.Mutex
	entries map[string][]time.Time
}

func newInMemoryStore() *inMemoryStore {
	return &inMemoryStore{
		entries: make(map[string][]time.Time),
	}
}

func (s *inMemoryStore) getOrCreate(key string) *inMemoryEntry {
	s.mu.Lock()
	defer s.mu.Unlock()
	if e, ok := s.entries[key]; ok {
		return &inMemoryEntry{timestamps: e, store: s, key: key}
	}
	e := &inMemoryEntry{timestamps: []time.Time{}, store: s, key: key}
	s.entries[key] = e.timestamps
	return e
}

type inMemoryEntry struct {
	timestamps []time.Time
	mu         sync.Mutex
	store      *inMemoryStore
	key        string
}

func (e *inMemoryEntry) clean(cutoff time.Time) {
	valid := e.timestamps[:0]
	for _, t := range e.timestamps {
		if t.After(cutoff) {
			valid = append(valid, t)
		}
	}
	e.timestamps = valid
}

func newAnalyzeRateLimiter(limit int, window time.Duration) *analyzeRateLimiter {
	l := &analyzeRateLimiter{
		memStore: newInMemoryStore(),
		limit:    limit,
		window:   window,
		stopCh:   make(chan struct{}),
	}
	go l.cleanupLoop()
	return l
}

func (l *analyzeRateLimiter) cleanupLoop() {
	ticker := time.NewTicker(5 * time.Minute)
	defer ticker.Stop()
	for {
		select {
		case <-l.stopCh:
			return
		case <-ticker.C:
			l.cleanupMem()
		}
	}
}

// Stop gracefully stops the cleanup goroutine.
func (l *analyzeRateLimiter) Stop() {
	close(l.stopCh)
}

func (l *analyzeRateLimiter) cleanupMem() {
	l.memStore.mu.Lock()
	defer l.memStore.mu.Unlock()
	cutoff := time.Now().Add(-l.window)
	for userID, timestamps := range l.memStore.entries {
		valid := timestamps[:0]
		for _, t := range timestamps {
			if t.After(cutoff) {
				valid = append(valid, t)
			}
		}
		if len(valid) == 0 {
			delete(l.memStore.entries, userID)
		} else {
			l.memStore.entries[userID] = valid
		}
	}
}

// Allow checks if a request is allowed under the rate limit using Redis distributed
// sliding window. Falls back to in-memory if Redis is unavailable.
func (l *analyzeRateLimiter) Allow(ctx context.Context, userID string) bool {
	if l.redisClient != nil {
		return l.allowRedis(ctx, userID)
	}
	return l.allowMem(userID)
}

func (l *analyzeRateLimiter) allowRedis(ctx context.Context, userID string) bool {
	now := time.Now()
	windowStart := now.Add(-l.window).UnixMilli()
	currentTime := float64(now.UnixNano()) / float64(time.Millisecond)

	// Hash userID to prevent Redis key injection
	h := sha256.Sum256([]byte(userID))
	sanitizedKey := hex.EncodeToString(h[:8])
	redisKey := analyzeRateLimitKeyPrefix + sanitizedKey

	if err := l.redisClient.ZRemRangeByScore(ctx, redisKey, "-inf", strconv.FormatInt(windowStart, 10)).Err(); err != nil {
		return l.allowMem(userID)
	}

	count, err := l.redisClient.ZCard(ctx, redisKey).Result()
	if err != nil {
		return l.allowMem(userID)
	}

	if count >= int64(l.limit) {
		return false
	}

	if err := l.redisClient.ZAdd(ctx, redisKey, redis.Z{
		Score:  currentTime,
		Member: fmt.Sprintf("%.0f", currentTime),
	}).Err(); err != nil {
		return l.allowMem(userID)
	}

	l.redisClient.Expire(ctx, redisKey, l.window+time.Second)
	return true
}

func (l *analyzeRateLimiter) allowMem(userID string) bool {
	entry := l.memStore.getOrCreate(userID)
	entry.mu.Lock()
	defer entry.mu.Unlock()

	now := time.Now()
	cutoff := now.Add(-l.window)
	entry.clean(cutoff)

	if len(entry.timestamps) >= l.limit {
		return false
	}

	entry.timestamps = append(entry.timestamps, now)
	return true
}

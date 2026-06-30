package agent

import (
	"crypto/subtle"
	"encoding/json"
	"net/http"
	"os"
	"strings"
	"time"

	"github.com/functionfly/functionfly/internal/agent/strategist"
	"github.com/google/uuid"
	"github.com/gorilla/mux"
	"github.com/sirupsen/logrus"
	"gorm.io/gorm"
)

// OptimizationHandler handles optimization suggestions from the Rust runtime
type OptimizationHandler struct {
	db *gorm.DB
}

// NewOptimizationHandler creates a new optimization handler
func NewOptimizationHandler(db *gorm.DB) *OptimizationHandler {
	return &OptimizationHandler{db: db}
}

// ReceiveOptimizationSuggestion handles POST /api/optimizations
// This endpoint receives optimization suggestions from the Rust runtime's GraphOptimizer
func (h *OptimizationHandler) ReceiveOptimizationSuggestion(w http.ResponseWriter, r *http.Request) {
	// Validate that the request is from the runtime (internal auth)
	internalToken := r.Header.Get("X-Internal-Runtime-Token")
	expectedToken := getRuntimeToken()
	// Only enforce token validation when RUNTIME_API_TOKEN is configured.
	// When empty, the endpoint is unauthenticated (dev mode only).
	if expectedToken != "" && subtle.ConstantTimeCompare([]byte(internalToken), []byte(expectedToken)) != 1 {
		writeError(w, http.StatusUnauthorized, "UNAUTHORIZED", "invalid runtime token")
		return
	}

	var suggestion RuntimeOptimizationSuggestion
	if err := json.NewDecoder(r.Body).Decode(&suggestion); err != nil {
		writeErrorFromErr(r, w, http.StatusBadRequest, "INVALID_REQUEST", "decode optimization suggestion", err)
		return
	}

	// Validate required fields
	if suggestion.GraphID == "" || suggestion.NodeID == "" {
		writeError(w, http.StatusBadRequest, "MISSING_FIELDS", "graph_id and node_id are required")
		return
	}

	// Parse graph ID to extract tenant/agent information
	graphID, err := uuid.Parse(suggestion.GraphID)
	if err != nil {
		writeError(w, http.StatusBadRequest, "INVALID_GRAPH_ID", "invalid graph_id format")
		return
	}

	// Create a strategist proposal from the runtime suggestion
	proposal := strategist.ModificationProposal{
		ID:                uuid.New(),
		GraphID:           graphID.String(),
		ChangeType:        mapActionToChangeType(suggestion.Action),
		TargetNodeID:      suggestion.NodeID,
		TargetNodeName:    suggestion.NodeName,
		ExpectedLiftPct:   suggestion.ExpectedImpact,
		RiskScore:         calculateRiskScore(suggestion.PatternConfidence),
		Status:            strategist.StatusPending,
		GlobalPatternRefs: suggestion.PatternConfidence,
		CreatedAt:         time.Now().UTC(),
	}

	// Store the proposal in the database
	if err := h.db.WithContext(r.Context()).Create(&proposal).Error; err != nil {
		logrus.WithError(err).Error("failed to store optimization suggestion")
		writeError(w, http.StatusInternalServerError, "INTERNAL_ERROR", "failed to store suggestion")
		return
	}

	logrus.WithFields(logrus.Fields{
		"proposal_id": proposal.ID,
		"graph_id":    suggestion.GraphID,
		"node_id":     suggestion.NodeID,
		"action":      suggestion.Action,
		"confidence":  suggestion.PatternConfidence,
	}).Info("Optimization suggestion received from runtime")

	writeJSON(w, http.StatusOK, map[string]any{
		"ok":          true,
		"proposal_id": proposal.ID,
		"status":      "stored",
	})
}

// RuntimeOptimizationSuggestion represents an optimization suggestion from the Rust runtime
type RuntimeOptimizationSuggestion struct {
	ID                string  `json:"id"`
	GraphID           string  `json:"graph_id"`
	NodeID            string  `json:"node_id"`
	NodeName          string  `json:"node_name"`
	PatternConfidence string  `json:"pattern_confidence"`
	Action            string  `json:"action"`
	ExpectedImpact    float64 `json:"expected_impact"`
	Difficulty        string  `json:"difficulty"`
	Description       string  `json:"description"`
}

// mapActionToChangeType maps runtime action strings to strategist change types
func mapActionToChangeType(action string) string {
	// Parse the action string (format: "action_type: details")
	switch {
	case contains(action, "adjust_timeout"):
		return "adjust_timeout"
	case contains(action, "enable_caching"):
		return "enable_caching"
	case contains(action, "model_downgrade"):
		return "model_downgrade"
	case contains(action, "simplify_path"):
		return "simplify_path"
	case contains(action, "increase_quota"):
		return "increase_quota"
	case contains(action, "adjust_retry"):
		return "add_retry"
	default:
		return "optimize"
	}
}

// contains checks if a string contains a substring
func contains(s, substr string) bool {
	return strings.Contains(s, substr)
}

// calculateRiskScore converts pattern confidence to a risk score
func calculateRiskScore(confidence string) float64 {
	switch confidence {
	case "high":
		return 0.1 // Low risk - high confidence
	case "medium":
		return 0.3 // Medium risk
	default:
		return 0.5 // Higher risk for low confidence
	}
}

// getRuntimeToken retrieves the expected runtime token from environment.
// Reads from RUNTIME_API_TOKEN env var. Returns empty string if not set,
// which means the endpoint is unauthenticated (dev mode).
func getRuntimeToken() string {
	token := os.Getenv("RUNTIME_API_TOKEN")
	return token
}

// RegisterOptimizationRoutes registers the optimization routes
func (h *OptimizationHandler) RegisterOptimizationRoutes(router *mux.Router, basePath string) {
	// This endpoint is called by the Rust runtime, not by end users
	// It can be on a separate internal router or protected by internal auth
	router.HandleFunc(basePath+"/optimizations", h.ReceiveOptimizationSuggestion).Methods(http.MethodPost)
}

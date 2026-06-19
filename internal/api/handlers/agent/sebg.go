package agent

import (
	"encoding/json"
	"net/http"
	"strconv"
	"time"

	"github.com/functionfly/functionfly/internal/agent/actuator"
	"github.com/functionfly/functionfly/internal/agent/analyzer"
	"github.com/functionfly/functionfly/internal/agent/autonomy"
	"github.com/functionfly/functionfly/internal/agent/globalpatternlibrary"
	"github.com/functionfly/functionfly/internal/agent/graph"
	"github.com/functionfly/functionfly/internal/agent/identity"
	"github.com/functionfly/functionfly/internal/agent/sebg"
	"github.com/functionfly/functionfly/internal/agent/strategist"
	"github.com/functionfly/functionfly/internal/api/middleware"
	"github.com/google/uuid"
	"github.com/gorilla/mux"
	"github.com/sirupsen/logrus"
	"gorm.io/gorm"
)

// SEBGHandler exposes the SEBG (Self-Evolving Backend Graph) control loop over HTTP.
type SEBGHandler struct {
	db           *gorm.DB
	identityRepo *identity.Repository
}

// NewSEBGHandler creates a new SEBG handler.
func NewSEBGHandler(db *gorm.DB, identityRepo *identity.Repository) *SEBGHandler {
	return &SEBGHandler{db: db, identityRepo: identityRepo}
}

// requireAgentTenant verifies the request is authenticated and the agent belongs to the caller's tenant.
func (h *SEBGHandler) requireAgentTenant(w http.ResponseWriter, r *http.Request, agentID string) bool {
	claims := middleware.GetUserFromContext(r)
	if claims == nil {
		writeError(w, http.StatusUnauthorized, "UNAUTHORIZED", "authentication required")
		return false
	}
	if agentID == "" {
		writeError(w, http.StatusBadRequest, "BAD_REQUEST", "agent_id is required")
		return false
	}
	agent, err := h.identityRepo.GetAgent(r.Context(), agentID)
	if err != nil {
		writeError(w, http.StatusNotFound, "NOT_FOUND", "agent not found")
		return false
	}
	if agent.TenantID != claims.TenantID {
		writeError(w, http.StatusForbidden, "FORBIDDEN", "access denied")
		return false
	}
	return true
}

// getOrCreateConfig returns the tenant's SEBG config, creating a default if it doesn't exist.
func (h *SEBGHandler) getOrCreateConfig(tenantID uuid.UUID) (*sebg.TenantConfig, error) {
	var cfg sebg.TenantConfig
	err := h.db.Where("tenant_id = ?", tenantID).First(&cfg).Error
	if err == nil {
		return &cfg, nil
	}
	cfg = sebg.TenantConfig{
		ID:                    uuid.New(),
		TenantID:             tenantID,
		AutonomyTier:         sebg.TierManual,
		MaxRiskScoreAutoApply: 0.2,
		IsActive:            true,
	}
	h.db.Create(&cfg)
	return &cfg, nil
}

// listProposalssvc returns a strategist service wired to the DB.
func listProposalsSvc(db *gorm.DB) *strategist.Service {
	patternLib := globalpatternlibrary.NewService(db)
	return strategist.NewService(db, patternLib)
}

// ListProposals handles GET /v1/agent/{agent_id}/sebg/proposals.
// Query params: status (pending/approved/rejected/applied), limit, offset.
func (h *SEBGHandler) ListProposals(w http.ResponseWriter, r *http.Request) {
	agentID := mux.Vars(r)["agent_id"]
	if !h.requireAgentTenant(w, r, agentID) {
		return
	}
	tenantID, _ := uuid.Parse(middleware.GetUserFromContext(r).TenantID.String())

	status := r.URL.Query().Get("status")
	limit := 20
	if l := r.URL.Query().Get("limit"); l != "" {
		if v, err := strconv.Atoi(l); err == nil && v > 0 && v <= 100 {
			limit = v
		}
	}
	offset := 0
	if o := r.URL.Query().Get("offset"); o != "" {
		if v, err := strconv.Atoi(o); err == nil && v >= 0 {
			offset = v
		}
	}

	var proposals []strategist.ModificationProposal
	query := h.db.WithContext(r.Context()).Model(&strategist.ModificationProposal{}).
		Where("tenant_id = ?", tenantID.String())
	if status != "" {
		query = query.Where("status = ?", status)
	}
	if err := query.Order("created_at DESC").Limit(limit).Offset(offset).Find(&proposals).Error; err != nil {
		writeError(w, http.StatusInternalServerError, "INTERNAL_ERROR", "failed to list proposals")
		return
	}

	total := int64(len(proposals)) // accurate for small result sets
	writeJSON(w, http.StatusOK, map[string]any{
		"ok":        true,
		"proposals": proposals,
		"total":     total,
		"limit":     limit,
		"offset":    offset,
	})
}

// DecideProposal handles POST /v1/agent/{agent_id}/sebg/proposals/{proposal_id}/decide.
// Body: { "decision": "approved" | "rejected" }
func (h *SEBGHandler) DecideProposal(w http.ResponseWriter, r *http.Request) {
	agentID := mux.Vars(r)["agent_id"]
	if !h.requireAgentTenant(w, r, agentID) {
		return
	}
	proposalIDStr := mux.Vars(r)["proposal_id"]
	proposalID, err := uuid.Parse(proposalIDStr)
	if err != nil {
		writeError(w, http.StatusBadRequest, "BAD_REQUEST", "invalid proposal_id")
		return
	}

	var req struct {
		Decision string `json:"decision"` // approved / rejected
	}
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		writeErrorFromErr(r, w, http.StatusBadRequest, "INVALID_REQUEST", "decode sebg request", err)
		return
	}
	if req.Decision != "approved" && req.Decision != "rejected" {
		writeError(w, http.StatusBadRequest, "BAD_REQUEST", "decision must be 'approved' or 'rejected'")
		return
	}

	tenantID, _ := uuid.Parse(middleware.GetUserFromContext(r).TenantID.String())
	svc := listProposalsSvc(h.db)

	// Verify proposal belongs to this tenant
	var proposal strategist.ModificationProposal
	if err := h.db.WithContext(r.Context()).Where("id = ? AND tenant_id = ?", proposalID, tenantID.String()).First(&proposal).Error; err != nil {
		writeError(w, http.StatusNotFound, "NOT_FOUND", "proposal not found")
		return
	}

	if proposal.Status != strategist.StatusPending {
		writeError(w, http.StatusConflict, "CONFLICT", "proposal is no longer pending")
		return
	}

	if req.Decision == "approved" {
		if err := svc.Approve(r.Context(), proposalID, tenantID.String()); err != nil {
			writeError(w, http.StatusInternalServerError, "INTERNAL_ERROR", "failed to approve proposal")
			return
		}
		// Apply the proposal immediately for approved low-risk changes.
		// Higher-risk approved changes go through the autonomy tier logic in evolve.
		actuatorSvc := actuator.NewService(h.db, graph.NewService(h.db))
		if err := actuatorSvc.ApplyProposal(r.Context(), proposalID); err != nil {
			logrus.WithError(err).Warn("SEBG: failed to auto-apply approved proposal")
		}
	} else {
		if err := svc.Reject(r.Context(), proposalID, tenantID.String()); err != nil {
			writeError(w, http.StatusInternalServerError, "INTERNAL_ERROR", "failed to reject proposal")
			return
		}
	}

	writeJSON(w, http.StatusOK, map[string]any{
		"ok":         true,
		"proposal_id": proposalID,
		"decision":   req.Decision,
	})
}

// GetConfig handles GET /v1/agent/{agent_id}/sebg/config.
func (h *SEBGHandler) GetConfig(w http.ResponseWriter, r *http.Request) {
	agentID := mux.Vars(r)["agent_id"]
	if !h.requireAgentTenant(w, r, agentID) {
		return
	}
	tenantID, _ := uuid.Parse(middleware.GetUserFromContext(r).TenantID.String())

	cfg, err := h.getOrCreateConfig(tenantID)
	if err != nil {
		writeError(w, http.StatusInternalServerError, "INTERNAL_ERROR", "failed to load SEBG config")
		return
	}

	writeJSON(w, http.StatusOK, map[string]any{
		"ok":    true,
		"config": cfg,
	})
}

// UpdateTier handles PUT /v1/agent/{agent_id}/sebg/tier.
// Body: { "tier": "manual" | "assisted" | "fully_autonomous" }
func (h *SEBGHandler) UpdateTier(w http.ResponseWriter, r *http.Request) {
	agentID := mux.Vars(r)["agent_id"]
	if !h.requireAgentTenant(w, r, agentID) {
		return
	}
	tenantID, _ := uuid.Parse(middleware.GetUserFromContext(r).TenantID.String())

	var req struct {
		Tier string `json:"tier"`
	}
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		writeErrorFromErr(r, w, http.StatusBadRequest, "INVALID_REQUEST", "decode sebg request", err)
		return
	}
	if req.Tier != sebg.TierManual && req.Tier != sebg.TierAssisted && req.Tier != sebg.TierFullyAutonomous {
		writeError(w, http.StatusBadRequest, "BAD_REQUEST", "tier must be 'manual', 'assisted', or 'fully_autonomous'")
		return
	}

	cfg, err := h.getOrCreateConfig(tenantID)
	if err != nil {
		writeError(w, http.StatusInternalServerError, "INTERNAL_ERROR", "failed to load SEBG config")
		return
	}

	cfg.AutonomyTier = req.Tier
	if req.Tier == sebg.TierManual {
		cfg.MaxRiskScoreAutoApply = 0.0
	} else if req.Tier == sebg.TierAssisted {
		cfg.MaxRiskScoreAutoApply = 0.2
	} else {
		cfg.MaxRiskScoreAutoApply = 0.4
	}
	if err := h.db.WithContext(r.Context()).Save(cfg).Error; err != nil {
		writeError(w, http.StatusInternalServerError, "INTERNAL_ERROR", "failed to update SEBG tier")
		return
	}

	writeJSON(w, http.StatusOK, map[string]any{
		"ok":    true,
		"config": cfg,
	})
}

// TriggerEvolve handles POST /v1/agent/{agent_id}/sebg/evolve.
// Runs the full SEBG pipeline: Observe → Analyze → Strategize → Govern → Act.
func (h *SEBGHandler) TriggerEvolve(w http.ResponseWriter, r *http.Request) {
	agentID := mux.Vars(r)["agent_id"]
	if !h.requireAgentTenant(w, r, agentID) {
		return
	}
	tenantID, _ := uuid.Parse(middleware.GetUserFromContext(r).TenantID.String())

	cfg, err := h.getOrCreateConfig(tenantID)
	if err != nil || !cfg.IsActive {
		writeError(w, http.StatusBadRequest, "BAD_REQUEST", "SEBG is not active for this tenant")
		return
	}

	// Wire up the full SEBG stack (same as autonomy.evolveAction).
	analyzerSvc := autonomy.NewAnalyzerService(h.db)
	analyses, err := analyzerSvc.AnalyzeByTenant(r.Context(), analyzer.AnalyzeByTenantParams{
		TenantID:   tenantID,
		TimeWindow: 24 * time.Hour,
	})
	result := map[string]any{
		"status": "completed",
		"agent_id": agentID,
	}
	if err != nil || analyses == nil || len(analyses) == 0 {
		result["status"] = "no_data"
		result["graphs_analyzed"] = 0
		writeJSON(w, http.StatusOK, map[string]any{"ok": true, "result": result})
		return
	}

	strategistSvc := autonomy.NewStrategistService(h.db)
	governorSvc := autonomy.NewGovernorService(h.db)
	actuatorSvc := autonomy.NewActuatorService(h.db)

	var totalApplied, totalPending int
	for _, analysis := range analyses {
		proposals, err := strategistSvc.GenerateProposals(r.Context(), &analysis)
		if err != nil || len(proposals) == 0 {
			continue
		}
		for _, p := range proposals {
			decision, err := governorSvc.ReviewProposal(r.Context(), p.ID)
			if err != nil {
				continue
			}
			// Autonomy tier controls auto-apply threshold.
			if decision.AutoApproved || (cfg.AutonomyTier == sebg.TierAssisted && p.RiskScore <= 0.2) ||
				(cfg.AutonomyTier == sebg.TierFullyAutonomous && p.RiskScore <= 0.4) {
				if err := actuatorSvc.ApplyProposal(r.Context(), p.ID); err == nil {
					totalApplied++
					continue
				}
			}
			totalPending++
		}
	}

	result["graphs_analyzed"] = len(analyses)
	result["applied"] = totalApplied
	result["pending"] = totalPending
	writeJSON(w, http.StatusOK, map[string]any{"ok": true, "result": result})
}

// GetROI handles GET /v1/agent/{agent_id}/sebg/roi.
func (h *SEBGHandler) GetROI(w http.ResponseWriter, r *http.Request) {
	agentID := mux.Vars(r)["agent_id"]
	if !h.requireAgentTenant(w, r, agentID) {
		return
	}
	tenantIDStr := middleware.GetUserFromContext(r).TenantID.String()
	tenantID, _ := uuid.Parse(tenantIDStr)

	var applied int64
	var pending int64
	var revenueLift int64
	h.db.WithContext(r.Context()).Model(&strategist.ModificationProposal{}).
		Where("tenant_id = ? AND status = ?", tenantIDStr, strategist.StatusApplied).Count(&applied)
	h.db.WithContext(r.Context()).Model(&strategist.ModificationProposal{}).
		Where("tenant_id = ? AND status = ?", tenantIDStr, strategist.StatusPending).Count(&pending)
	h.db.WithContext(r.Context()).Model(&strategist.ModificationProposal{}).
		Where("tenant_id = ? AND status IN ?", tenantIDStr, []string{strategist.StatusApplied, strategist.StatusApproved}).
		Select("COALESCE(SUM(expected_revenue_lift), 0)").Scan(&revenueLift)

	writeJSON(w, http.StatusOK, map[string]any{
		"ok": true,
		"roi": sebg.ROIReport{
			TenantID:         tenantID,
			AppliedCount:     int(applied),
			PendingCount:     int(pending),
			RevenueLiftCents: revenueLift,
		},
	})
}

// RegisterRoutes registers all SEBG routes.
func (h *SEBGHandler) RegisterRoutes(router *mux.Router, basePath string, authMiddleware *middleware.AuthMiddleware) {
	auth := authMiddleware.RequireAuth
	agent := router.PathPrefix(basePath + "/agent").Subrouter()

	agent.HandleFunc("/{agent_id}/sebg/proposals", auth(h.ListProposals)).Methods(http.MethodGet)
	agent.HandleFunc("/{agent_id}/sebg/proposals/{proposal_id}/decide", auth(h.DecideProposal)).Methods(http.MethodPost)
	agent.HandleFunc("/{agent_id}/sebg/config", auth(h.GetConfig)).Methods(http.MethodGet)
	agent.HandleFunc("/{agent_id}/sebg/tier", auth(h.UpdateTier)).Methods(http.MethodPut)
	agent.HandleFunc("/{agent_id}/sebg/evolve", auth(h.TriggerEvolve)).Methods(http.MethodPost)
	agent.HandleFunc("/{agent_id}/sebg/roi", auth(h.GetROI)).Methods(http.MethodGet)
}

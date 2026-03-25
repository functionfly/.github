package agent

import (
	"encoding/json"
	"net/http"
	"strconv"
	"strings"
	"time"

	"github.com/functionfly/functionfly/internal/agent/autonomy"
	"github.com/functionfly/functionfly/internal/agent/economy"
	"github.com/functionfly/functionfly/internal/agent/evolution"
	"github.com/functionfly/functionfly/internal/agent/identity"
	"github.com/functionfly/functionfly/internal/agent/marketplace"
	"github.com/functionfly/functionfly/internal/agent/swarm"
	"github.com/functionfly/functionfly/internal/api/middleware"
	"github.com/gorilla/mux"
	"github.com/sirupsen/logrus"
)

// SwarmHandler handles agent swarm operations
type SwarmHandler struct {
	swarmService       *swarm.Service
	messageService     *swarm.MessageService
	walletService      *economy.Service
	marketplaceService *marketplace.Service
	evolutionService   *evolution.Service
	autonomyService    *autonomy.Service
	identityRepo       *identity.Repository
}

// NewSwarmHandler creates a new swarm handler
func NewSwarmHandler(
	swarmService *swarm.Service,
	messageService *swarm.MessageService,
	walletService *economy.Service,
	marketplaceService *marketplace.Service,
	evolutionService *evolution.Service,
	autonomyService *autonomy.Service,
	identityRepo *identity.Repository,
) *SwarmHandler {
	return &SwarmHandler{
		swarmService:       swarmService,
		messageService:     messageService,
		walletService:      walletService,
		marketplaceService: marketplaceService,
		evolutionService:   evolutionService,
		autonomyService:    autonomyService,
		identityRepo:       identityRepo,
	}
}

// requireAgentTenant ensures the request is authenticated and the agent_id belongs to the caller's tenant.
func (h *SwarmHandler) requireAgentTenant(w http.ResponseWriter, r *http.Request, agentID string) bool {
	logrus.Debugf("SwarmHandler: requireAgentTenant called with agentID=%s path=%s", agentID, r.URL.Path)
	if agentID == "" {
		writeError(w, http.StatusBadRequest, "BAD_REQUEST", "agent_id is required")
		return false
	}
	claims := middleware.GetUserFromContext(r)
	if claims == nil {
		logrus.Warnf("SwarmHandler: no claims in context for agentID=%s", agentID)
		writeError(w, http.StatusUnauthorized, "UNAUTHORIZED", "authentication required")
		return false
	}
	logrus.Debugf("SwarmHandler: claims.TenantID=%s for agentID=%s", claims.TenantID, agentID)
	agent, err := h.identityRepo.GetAgent(r.Context(), agentID)
	if err != nil {
		logrus.Warnf("SwarmHandler: agent not found agentID=%s err=%v", agentID, err)
		writeError(w, http.StatusNotFound, "NOT_FOUND", "agent not found")
		return false
	}
	logrus.Debugf("SwarmHandler: agent.TenantID=%s claims.TenantID=%s for agentID=%s", agent.TenantID, claims.TenantID, agentID)
	if agent.TenantID != claims.TenantID {
		logrus.Warnf("SwarmHandler: tenant mismatch for agentID=%s", agentID)
		writeError(w, http.StatusForbidden, "FORBIDDEN", "access denied")
		return false
	}
	return true
}

// SpawnChildRequest represents a request to spawn a child agent
type SpawnChildRequest struct {
	ChildAgentID     string         `json:"child_agent_id" validate:"required"`
	ChildName        string         `json:"child_name" validate:"required"`
	ChildDescription string         `json:"child_description"`
	SwarmRole        string         `json:"swarm_role"`
	MaxChildAgents   int            `json:"max_child_agents"`
	Capabilities     map[string]any `json:"capabilities"`
	InitialBudgetUSD float64        `json:"initial_budget_usd"`
}

// SpawnChild handles POST /v1/agent/:id/spawn
func (h *SwarmHandler) SpawnChild(w http.ResponseWriter, r *http.Request) {
	parentAgentID := mux.Vars(r)["id"]
	if !h.requireAgentTenant(w, r, parentAgentID) {
		return
	}

	var req SpawnChildRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		writeError(w, http.StatusBadRequest, "INVALID_REQUEST", err.Error())
		return
	}

	spawnReq := &swarm.SpawnChildRequest{
		ParentAgentID:    parentAgentID,
		ChildAgentID:     req.ChildAgentID,
		ChildName:        req.ChildName,
		ChildDescription: req.ChildDescription,
		SwarmRole:        req.SwarmRole,
		MaxChildAgents:   req.MaxChildAgents,
		Capabilities:     req.Capabilities,
		InitialBudgetUSD: req.InitialBudgetUSD,
	}

	agent, apiKey, err := h.swarmService.SpawnChild(r.Context(), spawnReq)
	if err != nil {
		writeError(w, http.StatusInternalServerError, "INTERNAL_ERROR", err.Error())
		return
	}

	writeJSON(w, http.StatusCreated, map[string]any{
		"ok":      true,
		"agent":   agent,
		"api_key": apiKey,
		"message": "Child agent spawned successfully",
	})
}

// GetChildren handles GET /v1/agent/:id/children
func (h *SwarmHandler) GetChildren(w http.ResponseWriter, r *http.Request) {
	agentID := mux.Vars(r)["id"]
	if !h.requireAgentTenant(w, r, agentID) {
		return
	}

	children, err := h.swarmService.GetChildren(r.Context(), agentID)
	if err != nil {
		writeError(w, http.StatusInternalServerError, "INTERNAL_ERROR", err.Error())
		return
	}

	writeJSON(w, http.StatusOK, map[string]any{
		"ok":       true,
		"children": children,
	})
}

// GetParent handles GET /v1/agent/:id/parent
func (h *SwarmHandler) GetParent(w http.ResponseWriter, r *http.Request) {
	agentID := mux.Vars(r)["id"]
	if !h.requireAgentTenant(w, r, agentID) {
		return
	}

	parent, err := h.swarmService.GetParent(r.Context(), agentID)
	if err != nil {
		writeError(w, http.StatusInternalServerError, "INTERNAL_ERROR", err.Error())
		return
	}

	writeJSON(w, http.StatusOK, map[string]any{
		"ok":     true,
		"parent": parent,
	})
}

// SendMessage handles POST /v1/agent/:id/message
func (h *SwarmHandler) SendMessage(w http.ResponseWriter, r *http.Request) {
	fromAgentID := mux.Vars(r)["id"]
	if !h.requireAgentTenant(w, r, fromAgentID) {
		return
	}

	var msg identity.AgentMessage
	if err := json.NewDecoder(r.Body).Decode(&msg); err != nil {
		writeError(w, http.StatusBadRequest, "INVALID_REQUEST", err.Error())
		return
	}

	msg.FromAgentID = fromAgentID

	if err := h.messageService.SendMessage(r.Context(), &msg); err != nil {
		writeError(w, http.StatusInternalServerError, "INTERNAL_ERROR", err.Error())
		return
	}

	writeJSON(w, http.StatusCreated, map[string]any{
		"ok":      true,
		"message": "Message sent successfully",
		"msg_id":  msg.ID,
	})
}

// GetInbox handles GET /v1/agent/:id/inbox
func (h *SwarmHandler) GetInbox(w http.ResponseWriter, r *http.Request) {
	agentID := mux.Vars(r)["id"]
	if !h.requireAgentTenant(w, r, agentID) {
		return
	}

	messages, err := h.messageService.GetInbox(r.Context(), agentID, 50)
	if err != nil {
		writeError(w, http.StatusInternalServerError, "INTERNAL_ERROR", err.Error())
		return
	}

	writeJSON(w, http.StatusOK, map[string]any{
		"ok":       true,
		"messages": messages,
	})
}

// GetWallet handles GET /v1/agent/:id/wallet
func (h *SwarmHandler) GetWallet(w http.ResponseWriter, r *http.Request) {
	agentID := mux.Vars(r)["id"]
	if !h.requireAgentTenant(w, r, agentID) {
		return
	}

	wallet, err := h.walletService.GetOrCreateWallet(r.Context(), agentID)
	if err != nil {
		writeError(w, http.StatusInternalServerError, "INTERNAL_ERROR", err.Error())
		return
	}

	writeJSON(w, http.StatusOK, map[string]any{
		"ok":     true,
		"wallet": wallet,
	})
}

// CreateListing handles POST /v1/marketplace/agent/list
func (h *SwarmHandler) CreateListing(w http.ResponseWriter, r *http.Request) {
	var req marketplace.CreateAgentListingRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		writeError(w, http.StatusBadRequest, "INVALID_REQUEST", err.Error())
		return
	}
	if req.AgentID == "" {
		writeError(w, http.StatusBadRequest, "BAD_REQUEST", "agent_id is required")
		return
	}
	if !h.requireAgentTenant(w, r, req.AgentID) {
		return
	}

	listing, err := h.marketplaceService.ListingAgent(r.Context(), &req)
	if err != nil {
		writeError(w, http.StatusInternalServerError, "INTERNAL_ERROR", err.Error())
		return
	}

	writeJSON(w, http.StatusCreated, map[string]any{
		"ok":      true,
		"listing": listing,
	})
}

// marketplaceAgentResponse is the API shape for one marketplace agent (camelCase for frontend).
type marketplaceAgentResponse struct {
	ID                     string            `json:"id"`
	AgentID                string            `json:"agentId"`
	Name                   string            `json:"name"`
	Description            string            `json:"description"`
	ListingType            string            `json:"listingType"`
	PricingModel           string            `json:"pricingModel"`
	PricePerCall           *float64          `json:"pricePerCall,omitempty"`
	SubscriptionMonthlyUsd *float64          `json:"subscriptionMonthlyUsd,omitempty"`
	RevenueSharePercent    *float64          `json:"revenueSharePercent,omitempty"`
	RatingScore            float64           `json:"ratingScore"`
	TotalCalls             int               `json:"totalCalls"`
	ROIScore               float64           `json:"roiScore"`
	TrustScore             *float64          `json:"trustScore,omitempty"`
	DeterministicVerified  bool              `json:"deterministicVerified"`
	Capabilities           []string          `json:"capabilities,omitempty"`
	Status                 string            `json:"status"`
	IsOfficial             bool              `json:"isOfficial"`
}

// officialAgentIDs are the FunctionFly-built default agents seeded at launch.
var officialAgentIDs = map[string]bool{
	"proofsmith":    true,
	"policymint":    true,
	"marginpilot":   true,
	"schemasheriff": true,
	"patchpulse":    true,
	"runbookweaver": true,
}

func marketplaceAgentFromResult(res marketplace.AgentSearchResult) marketplaceAgentResponse {
	l := res.Listing
	name, desc, status := "", "", "active"
	var trustScore *float64
	var capabilities []string
	deterministicVerified := false

	if l.Agent != nil {
		name = l.Agent.Name
		desc = l.Agent.Description
		status = l.Agent.Status
		ts := l.Agent.TrustScore
		trustScore = &ts
		// derive deterministic_verified: trust score >= 90
		deterministicVerified = l.Agent.TrustScore >= 90
		// flatten capabilities map keys into a string slice
		for k := range l.Agent.Capabilities {
			capabilities = append(capabilities, k)
		}
	}

	return marketplaceAgentResponse{
		ID:                     l.ID.String(),
		AgentID:                l.AgentID,
		Name:                   name,
		Description:            desc,
		ListingType:            l.ListingType,
		PricingModel:           l.PricingModel,
		PricePerCall:           l.PricePerCall,
		SubscriptionMonthlyUsd: l.SubscriptionMonthlyUSD,
		RevenueSharePercent:    l.RevenueSharePercent,
		RatingScore:            l.RatingScore,
		TotalCalls:             l.TotalCalls,
		ROIScore:               l.ROIScore,
		TrustScore:             trustScore,
		DeterministicVerified:  deterministicVerified,
		Capabilities:           capabilities,
		Status:                 status,
		IsOfficial:             officialAgentIDs[l.AgentID],
	}
}

// SearchAgents handles GET /v1/marketplace/agents
func (h *SwarmHandler) SearchAgents(w http.ResponseWriter, r *http.Request) {
	q := r.URL.Query()
	req := marketplace.SearchAgentsRequest{
		PricingModel: q.Get("pricing_model"),
		Limit:        20,
		Offset:       0,
	}
	if l := q.Get("limit"); l != "" {
		if n, _ := strconv.Atoi(l); n > 0 {
			req.Limit = n
		}
	}
	if o := q.Get("offset"); o != "" {
		if n, _ := strconv.Atoi(o); n >= 0 {
			req.Offset = n
		}
	}
	if v := q.Get("min_rating"); v != "" {
		if n, err := strconv.ParseFloat(v, 64); err == nil {
			req.MinRating = n
		}
	}
	if v := q.Get("max_price_per_call"); v != "" {
		if n, err := strconv.ParseFloat(v, 64); err == nil {
			req.MaxPricePerCall = n
		}
	}
	if v := q.Get("min_roi_score"); v != "" {
		if n, err := strconv.ParseFloat(v, 64); err == nil {
			req.MinROIScore = n
		}
	}
	if t := q.Get("listing_types"); t != "" {
		req.ListingTypes = strings.Split(t, ",")
		for i := range req.ListingTypes {
			req.ListingTypes[i] = strings.TrimSpace(req.ListingTypes[i])
		}
	}

	results, total, err := h.marketplaceService.SearchAgents(r.Context(), &req)
	if err != nil {
		writeError(w, http.StatusInternalServerError, "INTERNAL_ERROR", err.Error())
		return
	}

	agents := make([]marketplaceAgentResponse, len(results))
	for i, res := range results {
		agents[i] = marketplaceAgentFromResult(res)
	}
	writeJSON(w, http.StatusOK, map[string]any{
		"ok":     true,
		"agents": agents,
		"total":  total,
	})
}

// ProposeEvolution handles POST /v1/agent/:id/evolve
func (h *SwarmHandler) ProposeEvolution(w http.ResponseWriter, r *http.Request) {
	agentID := mux.Vars(r)["id"]
	if !h.requireAgentTenant(w, r, agentID) {
		return
	}

	analysis, err := h.evolutionService.AnalyzePerformance(r.Context(), agentID, 24*7*time.Hour)
	if err != nil {
		writeError(w, http.StatusInternalServerError, "INTERNAL_ERROR", err.Error())
		return
	}

	proposal, err := h.evolutionService.ProposeEvolution(r.Context(), agentID, analysis)
	if err != nil {
		writeError(w, http.StatusInternalServerError, "INTERNAL_ERROR", err.Error())
		return
	}

	writeJSON(w, http.StatusCreated, map[string]any{
		"ok":       true,
		"proposal": proposal,
		"analysis": analysis,
	})
}

// CreateSchedule handles POST /v1/agent/:id/schedule
func (h *SwarmHandler) CreateSchedule(w http.ResponseWriter, r *http.Request) {
	agentID := mux.Vars(r)["id"]
	if !h.requireAgentTenant(w, r, agentID) {
		return
	}

	var schedule identity.AutonomySchedule
	if err := json.NewDecoder(r.Body).Decode(&schedule); err != nil {
		writeError(w, http.StatusBadRequest, "INVALID_REQUEST", err.Error())
		return
	}

	schedule.AgentID = agentID

	if err := h.autonomyService.CreateSchedule(r.Context(), &schedule); err != nil {
		writeError(w, http.StatusInternalServerError, "INTERNAL_ERROR", err.Error())
		return
	}

	writeJSON(w, http.StatusCreated, map[string]any{
		"ok":       true,
		"schedule": schedule,
	})
}

// GetSchedules handles GET /v1/agent/:id/schedules
func (h *SwarmHandler) GetSchedules(w http.ResponseWriter, r *http.Request) {
	agentID := mux.Vars(r)["id"]
	if !h.requireAgentTenant(w, r, agentID) {
		return
	}

	schedules, err := h.autonomyService.GetSchedules(r.Context(), agentID)
	if err != nil {
		writeError(w, http.StatusInternalServerError, "INTERNAL_ERROR", err.Error())
		return
	}

	writeJSON(w, http.StatusOK, map[string]any{
		"ok":        true,
		"schedules": schedules,
	})
}

// RegisterRoutes registers swarm routes on a gorilla/mux router.
// authMiddleware is applied to every route so JWT claims are available in context.
func (h *SwarmHandler) RegisterRoutes(router *mux.Router, basePath string, authMiddleware *middleware.AuthMiddleware) {
	logrus.Infof("SwarmHandler: RegisterRoutes called with basePath=%s", basePath)
	auth := authMiddleware.RequireAuth
	agent := router.PathPrefix(basePath + "/agent").Subrouter()
	logrus.Infof("SwarmHandler: Registered agent subrouter at %s/agent", basePath)

	agent.HandleFunc("/{id}/spawn", auth(h.SpawnChild)).Methods(http.MethodPost)
	logrus.Infof("SwarmHandler: Registered route POST %s/agent/{id}/spawn", basePath)
	agent.HandleFunc("/{id}/children", auth(h.GetChildren)).Methods(http.MethodGet)
	agent.HandleFunc("/{id}/parent", auth(h.GetParent)).Methods(http.MethodGet)
	agent.HandleFunc("/{id}/message", auth(h.SendMessage)).Methods(http.MethodPost)
	agent.HandleFunc("/{id}/inbox", auth(h.GetInbox)).Methods(http.MethodGet)
	agent.HandleFunc("/{id}/wallet", auth(h.GetWallet)).Methods(http.MethodGet)
	agent.HandleFunc("/{id}/evolve", auth(h.ProposeEvolution)).Methods(http.MethodPost)
	agent.HandleFunc("/{id}/schedule", auth(h.CreateSchedule)).Methods(http.MethodPost)
	agent.HandleFunc("/{id}/schedules", auth(h.GetSchedules)).Methods(http.MethodGet)

	marketplace := router.PathPrefix(basePath + "/marketplace").Subrouter()
	marketplace.HandleFunc("/agents", h.SearchAgents).Methods(http.MethodGet)
	marketplace.HandleFunc("/agent/list", auth(h.CreateListing)).Methods(http.MethodPost)
	logrus.Infof("SwarmHandler: All routes registered successfully")
}

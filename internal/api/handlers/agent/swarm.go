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
	"github.com/gorilla/mux"
)

// SwarmHandler handles agent swarm operations
type SwarmHandler struct {
	swarmService       *swarm.Service
	messageService     *swarm.MessageService
	walletService      *economy.Service
	marketplaceService *marketplace.Service
	evolutionService   *evolution.Service
	autonomyService    *autonomy.Service
}

// NewSwarmHandler creates a new swarm handler
func NewSwarmHandler(
	swarmService *swarm.Service,
	messageService *swarm.MessageService,
	walletService *economy.Service,
	marketplaceService *marketplace.Service,
	evolutionService *evolution.Service,
	autonomyService *autonomy.Service,
) *SwarmHandler {
	return &SwarmHandler{
		swarmService:       swarmService,
		messageService:     messageService,
		walletService:      walletService,
		marketplaceService: marketplaceService,
		evolutionService:   evolutionService,
		autonomyService:    autonomyService,
	}
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
	if parentAgentID == "" {
		writeError(w, http.StatusBadRequest, "BAD_REQUEST", "agent_id is required")
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
	if agentID == "" {
		writeError(w, http.StatusBadRequest, "BAD_REQUEST", "agent_id is required")
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
	if agentID == "" {
		writeError(w, http.StatusBadRequest, "BAD_REQUEST", "agent_id is required")
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
	if fromAgentID == "" {
		writeError(w, http.StatusBadRequest, "BAD_REQUEST", "from_agent_id is required")
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
	if agentID == "" {
		writeError(w, http.StatusBadRequest, "BAD_REQUEST", "agent_id is required")
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
	if agentID == "" {
		writeError(w, http.StatusBadRequest, "BAD_REQUEST", "agent_id is required")
		return
	}

	wallet, err := h.walletService.GetWallet(r.Context(), agentID)
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
	ID                     string   `json:"id"`
	AgentID                string   `json:"agentId"`
	Name                   string   `json:"name"`
	Description            string   `json:"description"`
	ListingType            string   `json:"listingType"`
	PricingModel           string   `json:"pricingModel"`
	PricePerCall           *float64 `json:"pricePerCall,omitempty"`
	SubscriptionMonthlyUsd *float64 `json:"subscriptionMonthlyUsd,omitempty"`
	RevenueSharePercent    *float64 `json:"revenueSharePercent,omitempty"`
	RatingScore            float64  `json:"ratingScore"`
	TotalCalls             int      `json:"totalCalls"`
	ROIScore               float64  `json:"roiScore"`
}

func marketplaceAgentFromResult(res marketplace.AgentSearchResult) marketplaceAgentResponse {
	l := res.Listing
	name, desc := "", ""
	if l.Agent != nil {
		name, desc = l.Agent.Name, l.Agent.Description
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
	if agentID == "" {
		writeError(w, http.StatusBadRequest, "BAD_REQUEST", "agent_id is required")
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
	if agentID == "" {
		writeError(w, http.StatusBadRequest, "BAD_REQUEST", "agent_id is required")
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
	if agentID == "" {
		writeError(w, http.StatusBadRequest, "BAD_REQUEST", "agent_id is required")
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

// RegisterRoutes registers swarm routes on a gorilla/mux router
func (h *SwarmHandler) RegisterRoutes(router *mux.Router, basePath string) {
	agent := router.PathPrefix(basePath + "/agent").Subrouter()

	agent.HandleFunc("/{id}/spawn", h.SpawnChild).Methods(http.MethodPost)
	agent.HandleFunc("/{id}/children", h.GetChildren).Methods(http.MethodGet)
	agent.HandleFunc("/{id}/parent", h.GetParent).Methods(http.MethodGet)
	agent.HandleFunc("/{id}/message", h.SendMessage).Methods(http.MethodPost)
	agent.HandleFunc("/{id}/inbox", h.GetInbox).Methods(http.MethodGet)
	agent.HandleFunc("/{id}/wallet", h.GetWallet).Methods(http.MethodGet)
	agent.HandleFunc("/{id}/evolve", h.ProposeEvolution).Methods(http.MethodPost)
	agent.HandleFunc("/{id}/schedule", h.CreateSchedule).Methods(http.MethodPost)
	agent.HandleFunc("/{id}/schedules", h.GetSchedules).Methods(http.MethodGet)

	marketplace := router.PathPrefix(basePath + "/marketplace").Subrouter()
	marketplace.HandleFunc("/agents", h.SearchAgents).Methods(http.MethodGet)
	marketplace.HandleFunc("/agent/list", h.CreateListing).Methods(http.MethodPost)
}

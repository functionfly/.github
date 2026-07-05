package agent

import (
	"encoding/json"
	"fmt"
	"net/http"
	"strconv"
	"strings"
	"time"

	"github.com/functionfly/functionfly/internal/agent/autonomy"
	"github.com/functionfly/functionfly/internal/agent/deployment"
	"github.com/functionfly/functionfly/internal/agent/economy"
	"github.com/functionfly/functionfly/internal/agent/evolution"
	"github.com/functionfly/functionfly/internal/agent/identity"
	"github.com/functionfly/functionfly/internal/agent/learning"
	"github.com/functionfly/functionfly/internal/agent/marketplace"
	"github.com/functionfly/functionfly/internal/agent/security"
	"github.com/functionfly/functionfly/internal/agent/swarm"
	"github.com/functionfly/functionfly/internal/api/middleware"
	"github.com/functionfly/functionfly/internal/storage"
	"github.com/google/uuid"
	"github.com/gorilla/mux"
	"github.com/sirupsen/logrus"
)

// SwarmHandler handles agent swarm operations
type SwarmHandler struct {
	swarmService       *swarm.Service
	messageService     *swarm.MessageService
	walletService      *economy.Service
	financialTxRepo    *storage.FinancialTransactionRepository
	marketplaceService *marketplace.Service
	evolutionService   *evolution.Service
	autonomyService    *autonomy.Service
	identityRepo       *identity.Repository
	learningRepo       *learning.Repository
	analyzer           *learning.Analyzer
	optimizer          *learning.Optimizer
	generator          *deployment.Generator
	publisher          *deployment.Publisher
	securityService    *security.SwarmSecurityService
}

// NewSwarmHandler creates a new swarm handler
func NewSwarmHandler(
	swarmService *swarm.Service,
	messageService *swarm.MessageService,
	walletService *economy.Service,
	financialTxRepo *storage.FinancialTransactionRepository,
	marketplaceService *marketplace.Service,
	evolutionService *evolution.Service,
	autonomyService *autonomy.Service,
	identityRepo *identity.Repository,
	learningRepo *learning.Repository,
	analyzer *learning.Analyzer,
	optimizer *learning.Optimizer,
	generator *deployment.Generator,
	publisher *deployment.Publisher,
	securityService *security.SwarmSecurityService,
) *SwarmHandler {
	return &SwarmHandler{
		swarmService:       swarmService,
		messageService:     messageService,
		walletService:      walletService,
		financialTxRepo:    financialTxRepo,
		marketplaceService: marketplaceService,
		evolutionService:   evolutionService,
		autonomyService:    autonomyService,
		identityRepo:       identityRepo,
		learningRepo:       learningRepo,
		analyzer:           analyzer,
		optimizer:          optimizer,
		generator:          generator,
		publisher:          publisher,
		securityService:    securityService,
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
		writeErrorFromErr(r, w, http.StatusBadRequest, "INVALID_REQUEST", "decode swarm request", err)
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
		writeErrorFromErr(r, w, http.StatusInternalServerError, "INTERNAL_ERROR", "swarm handler", err)
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
		writeErrorFromErr(r, w, http.StatusInternalServerError, "INTERNAL_ERROR", "swarm handler", err)
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
		writeErrorFromErr(r, w, http.StatusInternalServerError, "INTERNAL_ERROR", "swarm handler", err)
		return
	}

	writeJSON(w, http.StatusOK, map[string]any{
		"ok":     true,
		"parent": parent,
	})
}

// ReassignRoleRequest represents a request to reassign an agent's swarm role
type ReassignRoleRequest struct {
	SwarmRole string `json:"swarm_role" validate:"required"`
}

// ReassignRole handles PUT /v1/agent/:id/role
func (h *SwarmHandler) ReassignRole(w http.ResponseWriter, r *http.Request) {
	agentID := mux.Vars(r)["id"]
	if !h.requireAgentTenant(w, r, agentID) {
		return
	}

	var req ReassignRoleRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		writeErrorFromErr(r, w, http.StatusBadRequest, "INVALID_REQUEST", "decode swarm request", err)
		return
	}

	if req.SwarmRole == "" {
		writeError(w, http.StatusBadRequest, "BAD_REQUEST", "swarm_role is required")
		return
	}

	if err := h.swarmService.ReassignRole(r.Context(), agentID, req.SwarmRole); err != nil {
		writeErrorFromErr(r, w, http.StatusInternalServerError, "INTERNAL_ERROR", "swarm handler", err)
		return
	}

	writeJSON(w, http.StatusOK, map[string]any{
		"ok":        true,
		"agent_id":  agentID,
		"swarm_role": req.SwarmRole,
	})
}

// ReshapeSwarmRequest represents a request to reshape the swarm topology
type ReshapeSwarmRequest struct {
	Topology string `json:"topology" validate:"required"`
}

// ReshapeSwarm handles PUT /v1/agent/:id/topology
func (h *SwarmHandler) ReshapeSwarm(w http.ResponseWriter, r *http.Request) {
	agentID := mux.Vars(r)["id"]
	if !h.requireAgentTenant(w, r, agentID) {
		return
	}

	var req ReshapeSwarmRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		writeErrorFromErr(r, w, http.StatusBadRequest, "INVALID_REQUEST", "decode swarm request", err)
		return
	}

	if req.Topology == "" {
		writeError(w, http.StatusBadRequest, "BAD_REQUEST", "topology is required")
		return
	}

	if err := h.swarmService.ReshapeSwarm(r.Context(), agentID, req.Topology); err != nil {
		writeErrorFromErr(r, w, http.StatusInternalServerError, "INTERNAL_ERROR", "swarm handler", err)
		return
	}

	writeJSON(w, http.StatusOK, map[string]any{
		"ok":        true,
		"agent_id":  agentID,
		"topology":  req.Topology,
	})
}

// SendMessage handles POST /v1/agent/:id/message
// Requires X-Agent-Signing-Key header for message authentication
func (h *SwarmHandler) SendMessage(w http.ResponseWriter, r *http.Request) {
	fromAgentID := mux.Vars(r)["id"]
	if !h.requireAgentTenant(w, r, fromAgentID) {
		return
	}

	signingKey := r.Header.Get("X-Agent-Signing-Key")
	if signingKey == "" {
		writeError(w, http.StatusUnauthorized, "MISSING_SIGNING_KEY", "X-Agent-Signing-Key header is required for agent-to-agent messages")
		return
	}

	agent, err := h.identityRepo.GetAgentBySigningKeyHash(r.Context(), signingKey)
	if err != nil || agent == nil {
		writeError(w, http.StatusUnauthorized, "INVALID_SIGNING_KEY", "invalid signing key")
		return
	}

	if agent.AgentID != fromAgentID {
		writeError(w, http.StatusForbidden, "SIGNING_KEY_MISMATCH", "signing key does not match the from agent")
		return
	}

	var msg identity.AgentMessage
	if err := json.NewDecoder(r.Body).Decode(&msg); err != nil {
		writeErrorFromErr(r, w, http.StatusBadRequest, "INVALID_REQUEST", "decode swarm request", err)
		return
	}

	msg.FromAgentID = fromAgentID

	if err := h.messageService.SendMessage(r.Context(), &msg, signingKey); err != nil {
		writeErrorFromErr(r, w, http.StatusInternalServerError, "INTERNAL_ERROR", "swarm handler", err)
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
		writeErrorFromErr(r, w, http.StatusInternalServerError, "INTERNAL_ERROR", "swarm handler", err)
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
	claims := middleware.GetUserFromContext(r)
	if claims == nil {
		writeError(w, http.StatusUnauthorized, "UNAUTHORIZED", "authentication required")
		return
	}

	wallet, err := h.walletService.GetOrCreateWallet(r.Context(), agentID)
	if err != nil {
		writeErrorFromErr(r, w, http.StatusInternalServerError, "INTERNAL_ERROR", "swarm handler", err)
		return
	}
	if h.financialTxRepo != nil {
		summary, err := h.financialTxRepo.GetAgentWalletSummary(r.Context(), claims.TenantID, agentID)
		if err != nil {
			writeError(w, http.StatusInternalServerError, "INTERNAL_ERROR", "failed to compute wallet summary")
			return
		}
		if summary != nil {
			wallet.BalanceUSD = summary.BalanceUSD
			wallet.TotalEarnedUSD = summary.TotalEarnedUSD
			wallet.TotalSpentUSD = summary.TotalSpentUSD
			wallet.LastEarningAt = summary.LastEarningAt
			wallet.LastSpendingAt = summary.LastSpendingAt
		}
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
		writeErrorFromErr(r, w, http.StatusBadRequest, "INVALID_REQUEST", "decode swarm request", err)
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
		writeErrorFromErr(r, w, http.StatusInternalServerError, "INTERNAL_ERROR", "swarm handler", err)
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
	TrustScore             *float64 `json:"trustScore,omitempty"`
	DeterministicVerified  bool     `json:"deterministicVerified"`
	Capabilities           []string `json:"capabilities,omitempty"`
	Status                 string   `json:"status"`
	IsOfficial             bool     `json:"isOfficial"`
	RankScore              float64  `json:"rankScore"`

	// Auth-gated fields (only populated when user is authenticated)
	WalletBalanceUSD    *float64 `json:"walletBalanceUsd,omitempty"`
	HiringHistoryCount  *int     `json:"hiringHistoryCount,omitempty"`
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
		deterministicVerified = l.Agent.TrustScore >= 90
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
		RankScore:              res.RankScore,
	}
}

// SearchAgents handles GET /v1/marketplace/agents
func (h *SwarmHandler) SearchAgents(w http.ResponseWriter, r *http.Request) {
	q := r.URL.Query()
	req := marketplace.SearchAgentsRequest{
		AgentID:      q.Get("agent_id"),
		PricingModel: q.Get("pricing_model"),
		SortBy:       q.Get("sort_by"),
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
	if c := q.Get("capabilities"); c != "" {
		req.Capabilities = strings.Split(c, ",")
		for i := range req.Capabilities {
			req.Capabilities[i] = strings.TrimSpace(req.Capabilities[i])
		}
	}

	if req.AgentID != "" {
		if parsedUUID, err := uuid.Parse(req.AgentID); err == nil {
			if agent, err := h.identityRepo.GetAgentByUUID(r.Context(), parsedUUID); err == nil {
				req.AgentID = agent.AgentID
			}
		}
	}

	results, total, err := h.marketplaceService.SearchAgents(r.Context(), &req)
	if err != nil {
		writeErrorFromErr(r, w, http.StatusInternalServerError, "INTERNAL_ERROR", "swarm handler", err)
		return
	}

	claims := middleware.GetUserFromContext(r)
	agents := make([]interface{}, len(results))
	for i, res := range results {
		agentResp := marketplaceAgentFromResult(res)
		if claims != nil {
			wallet, err := h.walletService.GetWallet(r.Context(), res.Listing.AgentID)
			if err == nil && wallet != nil {
				balance := wallet.BalanceUSD
				agentResp.WalletBalanceUSD = &balance
			}
			count, err := h.identityRepo.CountAgentHiring(r.Context(), res.Listing.AgentID)
			if err == nil {
				agentResp.HiringHistoryCount = &count
			}
		}
		agents[i] = agentResp
	}

	hasMore := int64(req.Offset+len(results)) < total
	writeJSON(w, http.StatusOK, map[string]any{
		"ok":     true,
		"agents": agents,
		"total":  total,
		"limit":  req.Limit,
		"offset": req.Offset,
		"has_more": hasMore,
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
		writeErrorFromErr(r, w, http.StatusInternalServerError, "INTERNAL_ERROR", "swarm handler", err)
		return
	}

	proposal, err := h.evolutionService.ProposeEvolution(r.Context(), agentID, analysis)
	if err != nil {
		writeErrorFromErr(r, w, http.StatusInternalServerError, "INTERNAL_ERROR", "swarm handler", err)
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
		writeErrorFromErr(r, w, http.StatusBadRequest, "INVALID_REQUEST", "decode swarm request", err)
		return
	}

	schedule.AgentID = agentID

	if err := h.autonomyService.CreateSchedule(r.Context(), &schedule); err != nil {
		writeErrorFromErr(r, w, http.StatusInternalServerError, "INTERNAL_ERROR", "swarm handler", err)
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
		writeErrorFromErr(r, w, http.StatusInternalServerError, "INTERNAL_ERROR", "swarm handler", err)
		return
	}

	writeJSON(w, http.StatusOK, map[string]any{
		"ok":        true,
		"schedules": schedules,
	})
}

// AnalyzeAgent handles GET /v1/agent/{id}/analyze - execution pattern analysis
func (h *SwarmHandler) AnalyzeAgent(w http.ResponseWriter, r *http.Request) {
	agentID := mux.Vars(r)["id"]
	if !h.requireAgentTenant(w, r, agentID) {
		return
	}

	// Parse time window parameter (default 7 days)
	timeWindow := 7 * 24 * time.Hour
	if days := r.URL.Query().Get("days"); days != "" {
		if d, err := strconv.Atoi(days); err == nil && d > 0 {
			timeWindow = time.Duration(d) * 24 * time.Hour
		}
	}

	result, err := h.analyzer.AnalyzePatterns(r.Context(), agentID, timeWindow)
	if err != nil {
		writeErrorFromErr(r, w, http.StatusInternalServerError, "INTERNAL_ERROR", "swarm handler", err)
		return
	}

	writeJSON(w, http.StatusOK, map[string]any{
		"ok":       true,
		"analysis": result,
	})
}

// OptimizeAgent handles POST /v1/agent/{id}/optimize - generate optimization recommendations
func (h *SwarmHandler) OptimizeAgent(w http.ResponseWriter, r *http.Request) {
	agentID := mux.Vars(r)["id"]
	if !h.requireAgentTenant(w, r, agentID) {
		return
	}

	result, err := h.optimizer.AutoOptimize(r.Context(), agentID)
	if err != nil {
		writeErrorFromErr(r, w, http.StatusInternalServerError, "INTERNAL_ERROR", "swarm handler", err)
		return
	}

	writeJSON(w, http.StatusOK, map[string]any{
		"ok":            true,
		"result":        result,
		"optimizations": result.Optimizations,
	})
}

// GetInsights handles GET /v1/agent/{id}/insights - get learning insights and patterns
func (h *SwarmHandler) GetInsights(w http.ResponseWriter, r *http.Request) {
	agentID := mux.Vars(r)["id"]
	if !h.requireAgentTenant(w, r, agentID) {
		return
	}

	// Get active patterns
	patterns, err := h.analyzer.GetActivePatterns(r.Context(), agentID)
	if err != nil {
		writeErrorFromErr(r, w, http.StatusInternalServerError, "INTERNAL_ERROR", "swarm handler", err)
		return
	}

	// Get optimizations
	optimizations, err := h.optimizer.GetOptimizations(r.Context(), agentID, "")
	if err != nil {
		writeErrorFromErr(r, w, http.StatusInternalServerError, "INTERNAL_ERROR", "swarm handler", err)
		return
	}

	// Get memory stats
	memories, total, err := h.learningRepo.GetMemories(r.Context(), agentID, "", 100, 0)
	if err != nil {
		writeErrorFromErr(r, w, http.StatusInternalServerError, "INTERNAL_ERROR", "swarm handler", err)
		return
	}

	writeJSON(w, http.StatusOK, map[string]any{
		"ok":            true,
		"patterns":      patterns,
		"optimizations": optimizations,
		"memories":      memories,
		"memory_count":  total,
	})
}

// SearchMemories handles GET /v1/agent/{id}/memories - search agent memories
func (h *SwarmHandler) SearchMemories(w http.ResponseWriter, r *http.Request) {
	agentID := mux.Vars(r)["id"]
	if !h.requireAgentTenant(w, r, agentID) {
		return
	}

	query := r.URL.Query().Get("q")
	limit := 50
	if l := r.URL.Query().Get("limit"); l != "" {
		if n, err := strconv.Atoi(l); err == nil && n > 0 {
			limit = n
		}
	}

	memories, err := h.learningRepo.SearchMemories(r.Context(), agentID, query, limit)
	if err != nil {
		writeErrorFromErr(r, w, http.StatusInternalServerError, "INTERNAL_ERROR", "swarm handler", err)
		return
	}

	writeJSON(w, http.StatusOK, map[string]any{
		"ok":       true,
		"query":    query,
		"memories": memories,
		"count":    len(memories),
	})
}

// GenerateCode handles POST /v1/agent/{id}/generate - generate code from specification
func (h *SwarmHandler) GenerateCode(w http.ResponseWriter, r *http.Request) {
	agentID := mux.Vars(r)["id"]
	if !h.requireAgentTenant(w, r, agentID) {
		return
	}

	var req deployment.GenerationRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		writeErrorFromErr(r, w, http.StatusBadRequest, "INVALID_REQUEST", "decode swarm request", err)
		return
	}

	req.AgentID = agentID

	generated, err := h.generator.Generate(r.Context(), &req)
	if err != nil {
		writeErrorFromErr(r, w, http.StatusInternalServerError, "INTERNAL_ERROR", "swarm handler", err)
		return
	}

	writeJSON(w, http.StatusCreated, map[string]any{
		"ok":      true,
		"code":    generated,
		"message": "Code generated successfully",
	})
}

// GetGenerations handles GET /v1/agent/{id}/generations - list generated code
func (h *SwarmHandler) GetGenerations(w http.ResponseWriter, r *http.Request) {
	agentID := mux.Vars(r)["id"]
	if !h.requireAgentTenant(w, r, agentID) {
		return
	}

	limit, offset := 20, 0
	if l := r.URL.Query().Get("limit"); l != "" {
		if n, err := strconv.Atoi(l); err == nil && n > 0 {
			limit = n
		}
	}
	if o := r.URL.Query().Get("offset"); o != "" {
		if n, err := strconv.Atoi(o); err == nil && n >= 0 {
			offset = n
		}
	}

	generations, total, err := h.generator.GetGenerations(r.Context(), agentID, limit, offset)
	if err != nil {
		writeErrorFromErr(r, w, http.StatusInternalServerError, "INTERNAL_ERROR", "swarm handler", err)
		return
	}

	writeJSON(w, http.StatusOK, map[string]any{
		"ok":          true,
		"generations": generations,
		"total":       total,
		"limit":       limit,
		"offset":      offset,
	})
}

// PublishFunction handles POST /v1/agent/{id}/publish - publish generated function
func (h *SwarmHandler) PublishFunction(w http.ResponseWriter, r *http.Request) {
	agentID := mux.Vars(r)["id"]
	if !h.requireAgentTenant(w, r, agentID) {
		return
	}

	var req deployment.PublishRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		writeErrorFromErr(r, w, http.StatusBadRequest, "INVALID_REQUEST", "decode swarm request", err)
		return
	}

	req.AgentID = agentID

	published, err := h.publisher.Publish(r.Context(), &req)
	if err != nil {
		writeErrorFromErr(r, w, http.StatusInternalServerError, "INTERNAL_ERROR", "swarm handler", err)
		return
	}

	writeJSON(w, http.StatusCreated, map[string]any{
		"ok":        true,
		"published": published,
		"message":   "Function published successfully",
	})
}

// GetPublishedFunctions handles GET /v1/agent/{id}/functions - list published functions
func (h *SwarmHandler) GetPublishedFunctions(w http.ResponseWriter, r *http.Request) {
	agentID := mux.Vars(r)["id"]
	if !h.requireAgentTenant(w, r, agentID) {
		return
	}

	limit, offset := 20, 0
	if l := r.URL.Query().Get("limit"); l != "" {
		if n, err := strconv.Atoi(l); err == nil && n > 0 {
			limit = n
		}
	}
	if o := r.URL.Query().Get("offset"); o != "" {
		if n, err := strconv.Atoi(o); err == nil && n >= 0 {
			offset = n
		}
	}

	functions, total, err := h.publisher.GetPublishedFunctions(r.Context(), agentID, limit, offset)
	if err != nil {
		writeErrorFromErr(r, w, http.StatusInternalServerError, "INTERNAL_ERROR", "swarm handler", err)
		return
	}

	writeJSON(w, http.StatusOK, map[string]any{
		"ok":        true,
		"functions": functions,
		"total":     total,
		"limit":     limit,
		"offset":    offset,
	})
}

// HireAgent handles POST /v1/marketplace/hire - hire an agent for a task
func (h *SwarmHandler) HireAgent(w http.ResponseWriter, r *http.Request) {
	var req struct {
		AgentID     string         `json:"agent_id" validate:"required"`
		TaskType    string         `json:"task_type" validate:"required"`
		TaskPayload map[string]any `json:"task_payload"`
		BudgetUSD   float64        `json:"budget_usd"`
	}
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		writeErrorFromErr(r, w, http.StatusBadRequest, "INVALID_REQUEST", "decode swarm request", err)
		return
	}

	claims := middleware.GetUserFromContext(r)
	if claims == nil {
		writeError(w, http.StatusUnauthorized, "UNAUTHORIZED", "authentication required")
		return
	}

	// Get the agent to verify it exists and is available
	agent, err := h.identityRepo.GetAgent(r.Context(), req.AgentID)
	if err != nil {
		writeError(w, http.StatusNotFound, "NOT_FOUND", "agent not found")
		return
	}

	hirerID := claims.UserID.String()

	// Validate the agent can be hired
	validation := h.securityService.ValidateDelegation(r.Context(), hirerID, req.AgentID, req.TaskType)
	if !validation.Valid {
		writeError(w, http.StatusForbidden, "FORBIDDEN", strings.Join(validation.Reasons, "; "))
		return
	}

	// Create the hiring record
	hiring := &identity.AgentHiring{
		ID:          uuid.New(),
		AgentID:     req.AgentID,
		HirerID:     hirerID,
		TenantID:    claims.TenantID.String(),
		TaskType:    req.TaskType,
		TaskPayload: req.TaskPayload,
		BudgetUSD:   req.BudgetUSD,
		Status:      "pending",
		CreatedAt:   time.Now(),
	}

	if err := h.identityRepo.CreateAgentHiring(r.Context(), hiring); err != nil {
		writeErrorFromErr(r, w, http.StatusInternalServerError, "INTERNAL_ERROR", "swarm handler", err)
		return
	}

	// Send message to the hired agent
	// Note: hiring message is sent without signing as it's initiated by the user via API
	// The hired agent can verify via its own signing key if needed
	hiringMsg := &identity.AgentMessage{
		ID:          uuid.New(),
		FromAgentID: hirerID,
		ToAgentID:   req.AgentID,
		MessageType: "hiring_request",
		Payload: map[string]any{
			"hiring_id":    hiring.ID.String(),
			"task_type":    req.TaskType,
			"task_payload": req.TaskPayload,
			"budget_usd":   req.BudgetUSD,
		},
		Status:    "pending",
		CreatedAt: time.Now(),
	}

	if err := h.messageService.ReceiveMessage(r.Context(), hiringMsg); err != nil {
		logrus.Warnf("Failed to send hiring message: %v", err)
	}

	writeJSON(w, http.StatusCreated, map[string]any{
		"ok":      true,
		"hiring":  hiring,
		"agent":   agent,
		"message": "Agent hired successfully",
	})
}

// PurchaseFunction handles POST /v1/marketplace/purchase - purchase a function
func (h *SwarmHandler) PurchaseFunction(w http.ResponseWriter, r *http.Request) {
	var req struct {
		FunctionAuthor string  `json:"function_author" validate:"required"`
		FunctionName   string  `json:"function_name" validate:"required"`
		AgentID        string  `json:"agent_id" validate:"required"`
		MaxPriceUSD    float64 `json:"max_price_usd"`
	}
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		writeErrorFromErr(r, w, http.StatusBadRequest, "INVALID_REQUEST", "decode swarm request", err)
		return
	}

	if !h.requireAgentTenant(w, r, req.AgentID) {
		return
	}

	// Check agent wallet balance
	wallet, err := h.walletService.GetOrCreateWallet(r.Context(), req.AgentID)
	if err != nil {
		writeError(w, http.StatusInternalServerError, "INTERNAL_ERROR", "failed to get wallet")
		return
	}

	// Get the published function to find the owner agent
	published, err := h.publisher.GetFunctionByURI(r.Context(), req.FunctionAuthor, req.FunctionName)
	if err != nil {
		writeError(w, http.StatusNotFound, "NOT_FOUND", "function not found")
		return
	}

	// Get the marketplace listing for pricing
	listing, err := h.marketplaceService.GetFunctionListingByURI(r.Context(), req.FunctionAuthor, req.FunctionName)
	if err != nil {
		writeError(w, http.StatusNotFound, "NOT_FOUND", "function listing not found")
		return
	}

	// Calculate the actual price based on pricing model
	pricePaid := 0.0
	switch listing.PricingModel {
	case "free":
		pricePaid = 0
	case "per_call":
		if listing.PricePerCall != nil {
			pricePaid = *listing.PricePerCall
		}
	case "subscription":
		if listing.SubscriptionMonthlyUSD != nil {
			// Use monthly rate as the purchase price for subscription
			pricePaid = *listing.SubscriptionMonthlyUSD
		}
	case "revenue_share":
		if listing.RevenueSharePercent != nil {
			// For revenue share, initial purchase is free, revenue is shared on usage
			pricePaid = 0
		}
	default:
		if listing.PricePerCall != nil {
			pricePaid = *listing.PricePerCall
		}
	}

	// Verify wallet has sufficient balance
	if wallet.BalanceUSD < pricePaid {
		writeError(w, http.StatusPaymentRequired, "INSUFFICIENT_FUNDS",
			fmt.Sprintf("insufficient wallet balance: $%.4f required, $%.4f available", pricePaid, wallet.BalanceUSD))
		return
	}

	// Check max price limit if specified
	if req.MaxPriceUSD > 0 && pricePaid > req.MaxPriceUSD {
		writeError(w, http.StatusBadRequest, "PRICE_EXCEEDS_LIMIT",
			fmt.Sprintf("function price $%.4f exceeds max price $%.4f", pricePaid, req.MaxPriceUSD))
		return
	}

	// Debit buyer's wallet
	if pricePaid > 0 {
		if _, err := h.walletService.Debit(r.Context(), req.AgentID, pricePaid, "function_purchase", map[string]any{
			"function_author": req.FunctionAuthor,
			"function_name":   req.FunctionName,
		}); err != nil {
			writeError(w, http.StatusInternalServerError, "PAYMENT_FAILED", "failed to process payment")
			return
		}

		// Credit seller's wallet (if published by an agent)
		if published.AgentID != "" && published.AgentID != req.AgentID {
			if _, err := h.walletService.Credit(r.Context(), published.AgentID, pricePaid, "function_sale", map[string]any{
				"function_author": req.FunctionAuthor,
				"function_name":   req.FunctionName,
				"buyer_agent_id":  req.AgentID,
			}); err != nil {
				// Log but don't fail - purchase is still valid
				logrus.WithError(err).Warnf("Failed to credit seller wallet for function sale: %s/%s", req.FunctionAuthor, req.FunctionName)
			}
		}
	}

	// Create purchase record
	purchase := &identity.FunctionPurchase{
		ID:             uuid.New(),
		AgentID:        req.AgentID,
		FunctionAuthor: req.FunctionAuthor,
		FunctionName:   req.FunctionName,
		PublishedID:    published.ID,
		PricePaidUSD:   pricePaid,
		Status:         "completed",
		CreatedAt:      time.Now(),
	}

	if err := h.identityRepo.CreateFunctionPurchase(r.Context(), purchase); err != nil {
		writeErrorFromErr(r, w, http.StatusInternalServerError, "INTERNAL_ERROR", "swarm handler", err)
		return
	}

	writeJSON(w, http.StatusCreated, map[string]any{
		"ok":       true,
		"purchase": purchase,
		"function": published,
		"message":  "Function purchased successfully",
	})
}

// TriggerKillSwitch handles POST /v1/agent/{id}/kill-switch - emergency agent suspension
func (h *SwarmHandler) TriggerKillSwitch(w http.ResponseWriter, r *http.Request) {
	agentID := mux.Vars(r)["id"]
	if !h.requireAgentTenant(w, r, agentID) {
		return
	}

	var req struct {
		Reason string `json:"reason"`
	}
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		logrus.WithError(err).Warn("failed to decode request body")
	} // Optional body

	if req.Reason == "" {
		req.Reason = "manual_trigger"
	}

	result, err := h.securityService.TriggerKillSwitch(r.Context(), agentID, req.Reason)
	if err != nil {
		writeErrorFromErr(r, w, http.StatusInternalServerError, "INTERNAL_ERROR", "swarm handler", err)
		return
	}

	writeJSON(w, http.StatusOK, map[string]any{
		"ok":            true,
		"result":        result,
		"agents_killed": result.AgentsKilled,
		"message":       fmt.Sprintf("Kill switch triggered: %d agents suspended", result.AgentsKilled),
	})
}

// CheckSwarmHealth handles GET /v1/agent/{id}/health - check swarm health and detect anomalies
func (h *SwarmHandler) CheckSwarmHealth(w http.ResponseWriter, r *http.Request) {
	agentID := mux.Vars(r)["id"]
	if !h.requireAgentTenant(w, r, agentID) {
		return
	}

	timeWindow := 24 * time.Hour
	if hours := r.URL.Query().Get("hours"); hours != "" {
		if h, err := strconv.Atoi(hours); err == nil && h > 0 {
			timeWindow = time.Duration(h) * time.Hour
		}
	}

	// Detect anomalies
	anomalies, err := h.securityService.DetectAnomaly(r.Context(), agentID, timeWindow)
	if err != nil {
		writeErrorFromErr(r, w, http.StatusInternalServerError, "INTERNAL_ERROR", "swarm handler", err)
		return
	}

	// Get children to check swarm health
	children, _ := h.swarmService.GetChildren(r.Context(), agentID)

	// Calculate health score
	healthScore := 100
	for _, a := range anomalies {
		switch a.Severity {
		case "high":
			healthScore -= 30
		case "medium":
			healthScore -= 15
		case "low":
			healthScore -= 5
		}
	}
	if healthScore < 0 {
		healthScore = 0
	}

	status := "healthy"
	if healthScore < 50 {
		status = "critical"
	} else if healthScore < 80 {
		status = "degraded"
	}

	writeJSON(w, http.StatusOK, map[string]any{
		"ok":           true,
		"agent_id":     agentID,
		"status":       status,
		"health_score": healthScore,
		"anomalies":    anomalies,
		"children":     len(children),
		"time_window":  timeWindow.String(),
	})
}

// RegisterRoutes registers swarm routes on a gorilla/mux router.
// authMiddleware is applied to every route so JWT claims are available in context.
func (h *SwarmHandler) RegisterRoutes(router *mux.Router, basePath string, authMiddleware *middleware.AuthMiddleware) {
	logrus.Infof("SwarmHandler: RegisterRoutes called with basePath=%s", basePath)
	auth := authMiddleware.RequireAuth
	agent := router.PathPrefix(basePath + "/agent").Subrouter()
	logrus.Infof("SwarmHandler: Registered agent subrouter at %s/agent", basePath)

	// Core swarm operations
	agent.HandleFunc("/{id}/spawn", auth(h.SpawnChild)).Methods(http.MethodPost)
	logrus.Infof("SwarmHandler: Registered route POST %s/agent/{id}/spawn", basePath)
	agent.HandleFunc("/{id}/children", auth(h.GetChildren)).Methods(http.MethodGet)
	agent.HandleFunc("/{id}/parent", auth(h.GetParent)).Methods(http.MethodGet)
	agent.HandleFunc("/{id}/role", auth(h.ReassignRole)).Methods(http.MethodPut)
	logrus.Infof("SwarmHandler: Registered route PUT %s/agent/{id}/role", basePath)
	agent.HandleFunc("/{id}/topology", auth(h.ReshapeSwarm)).Methods(http.MethodPut)
	logrus.Infof("SwarmHandler: Registered route PUT %s/agent/{id}/topology", basePath)
	agent.HandleFunc("/{id}/message", auth(h.SendMessage)).Methods(http.MethodPost)
	agent.HandleFunc("/{id}/inbox", auth(h.GetInbox)).Methods(http.MethodGet)
	agent.HandleFunc("/{id}/wallet", auth(h.GetWallet)).Methods(http.MethodGet)
	agent.HandleFunc("/{id}/evolve", auth(h.ProposeEvolution)).Methods(http.MethodPost)
	agent.HandleFunc("/{id}/schedule", auth(h.CreateSchedule)).Methods(http.MethodPost)
	agent.HandleFunc("/{id}/schedules", auth(h.GetSchedules)).Methods(http.MethodGet)

	// Learning & Analysis endpoints
	agent.HandleFunc("/{id}/analyze", auth(h.AnalyzeAgent)).Methods(http.MethodGet)
	agent.HandleFunc("/{id}/optimize", auth(h.OptimizeAgent)).Methods(http.MethodPost)
	agent.HandleFunc("/{id}/insights", auth(h.GetInsights)).Methods(http.MethodGet)
	agent.HandleFunc("/{id}/memories", auth(h.SearchMemories)).Methods(http.MethodGet)

	// Code generation & deployment endpoints
	agent.HandleFunc("/{id}/generate", auth(h.GenerateCode)).Methods(http.MethodPost)
	agent.HandleFunc("/{id}/generations", auth(h.GetGenerations)).Methods(http.MethodGet)
	agent.HandleFunc("/{id}/publish", auth(h.PublishFunction)).Methods(http.MethodPost)
	agent.HandleFunc("/{id}/functions", auth(h.GetPublishedFunctions)).Methods(http.MethodGet)

	// Security & health endpoints
	agent.HandleFunc("/{id}/kill-switch", auth(h.TriggerKillSwitch)).Methods(http.MethodPost)
	agent.HandleFunc("/{id}/health", auth(h.CheckSwarmHealth)).Methods(http.MethodGet)

	marketplaceBase := basePath + "/marketplace"
	router.HandleFunc(marketplaceBase+"/agents", h.SearchAgents).Methods(http.MethodGet)
	router.HandleFunc(marketplaceBase+"/agent/list", auth(h.CreateListing)).Methods(http.MethodPost)
	router.HandleFunc(marketplaceBase+"/hire", auth(h.HireAgent)).Methods(http.MethodPost)
	router.HandleFunc(marketplaceBase+"/purchase", auth(h.PurchaseFunction)).Methods(http.MethodPost)

	logrus.Infof("SwarmHandler: All routes registered successfully")
}

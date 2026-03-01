package agent

import (
	"net/http"
	"time"

	"github.com/functionfly/functionfly/internal/agent/autonomy"
	"github.com/functionfly/functionfly/internal/agent/evolution"
	"github.com/functionfly/functionfly/internal/agent/identity"
	"github.com/functionfly/functionfly/internal/agent/marketplace"
	"github.com/functionfly/functionfly/internal/agent/swarm"
	"github.com/functionfly/functionfly/internal/agent/economy"
	"github.com/google/uuid"
	"github.com/labstack/echo/v4"
)

// SwarmHandler handles agent swarm operations
type SwarmHandler struct {
	swarmService      *swarm.Service
	messageService    *swarm.MessageService
	walletService    *economy.Service
	marketplaceService *marketplace.Service
	evolutionService *evolution.Service
	autonomyService  *autonomy.Service
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
		evolutionService:  evolutionService,
		autonomyService:   autonomyService,
	}
}

// SpawnChildRequest represents a request to spawn a child agent
type SpawnChildRequest struct {
	ChildAgentID      string         `json:"child_agent_id" validate:"required"`
	ChildName         string         `json:"child_name" validate:"required"`
	ChildDescription string         `json:"child_description"`
	SwarmRole         string         `json:"swarm_role"`
	MaxChildAgents    int            `json:"max_child_agents"`
	Capabilities      map[string]any `json:"capabilities"`
	InitialBudgetUSD  float64        `json:"initial_budget_usd"`
}

// SpawnChild handles POST /v1/agent/:id/spawn
func (h *SwarmHandler) SpawnChild(c echo.Context) error {
	parentAgentID := c.Param("id")
	if parentAgentID == "" {
		return echo.NewHTTPError(http.StatusBadRequest, "agent_id is required")
	}

	var req SpawnChildRequest
	if err := c.Bind(&req); err != nil {
		return echo.NewHTTPError(http.StatusBadRequest, err.Error())
	}

	spawnReq := &swarm.SpawnChildRequest{
		ParentAgentID:     parentAgentID,
		ChildAgentID:      req.ChildAgentID,
		ChildName:         req.ChildName,
		ChildDescription:  req.ChildDescription,
		SwarmRole:         req.SwarmRole,
		MaxChildAgents:    req.MaxChildAgents,
		Capabilities:      req.Capabilities,
		InitialBudgetUSD:  req.InitialBudgetUSD,
	}

	agent, apiKey, err := h.swarmService.SpawnChild(c.Request().Context(), spawnReq)
	if err != nil {
		return echo.NewHTTPError(http.StatusInternalServerError, err.Error())
	}

	return c.JSON(http.StatusCreated, map[string]any{
		"ok":       true,
		"agent":    agent,
		"api_key":  apiKey,
		"message":  "Child agent spawned successfully",
	})
}

// GetChildren handles GET /v1/agent/:id/children
func (h *SwarmHandler) GetChildren(c echo.Context) error {
	agentID := c.Param("id")
	if agentID == "" {
		return echo.NewHTTPError(http.StatusBadRequest, "agent_id is required")
	}

	children, err := h.swarmService.GetChildren(c.Request().Context(), agentID)
	if err != nil {
		return echo.NewHTTPError(http.StatusInternalServerError, err.Error())
	}

	return c.JSON(http.StatusOK, map[string]any{
		"ok":       true,
		"children": children,
	})
}

// GetParent handles GET /v1/agent/:id/parent
func (h *SwarmHandler) GetParent(c echo.Context) error {
	agentID := c.Param("id")
	if agentID == "" {
		return echo.NewHTTPError(http.StatusBadRequest, "agent_id is required")
	}

	parent, err := h.swarmService.GetParent(c.Request().Context(), agentID)
	if err != nil {
		return echo.NewHTTPError(http.StatusInternalServerError, err.Error())
	}

	if parent == nil {
		return c.JSON(http.StatusOK, map[string]any{
			"ok":     true,
			"parent": nil,
		})
	}

	return c.JSON(http.StatusOK, map[string]any{
		"ok":     true,
		"parent": parent,
	})
}

// SendMessage handles POST /v1/agent/:id/message
func (h *SwarmHandler) SendMessage(c echo.Context) error {
	fromAgentID := c.Param("id")
	if fromAgentID == "" {
		return echo.NewHTTPError(http.StatusBadRequest, "from_agent_id is required")
	}

	var msg identity.AgentMessage
	if err := c.Bind(&msg); err != nil {
		return echo.NewHTTPError(http.StatusBadRequest, err.Error())
	}

	msg.FromAgentID = fromAgentID

	if err := h.messageService.SendMessage(c.Request().Context(), &msg); err != nil {
		return echo.NewHTTPError(http.StatusInternalServerError, err.Error())
	}

	return c.JSON(http.StatusCreated, map[string]any{
		"ok":      true,
		"message": "Message sent successfully",
		"msg_id":  msg.ID,
	})
}

// GetInbox handles GET /v1/agent/:id/inbox
func (h *SwarmHandler) GetInbox(c echo.Context) error {
	agentID := c.Param("id")
	if agentID == "" {
		return echo.NewHTTPError(http.StatusBadRequest, "agent_id is required")
	}

	messages, err := h.messageService.GetInbox(c.Request().Context(), agentID, 50)
	if err != nil {
		return echo.NewHTTPError(http.StatusInternalServerError, err.Error())
	}

	return c.JSON(http.StatusOK, map[string]any{
		"ok":       true,
		"messages": messages,
	})
}

// GetWallet handles GET /v1/agent/:id/wallet
func (h *SwarmHandler) GetWallet(c echo.Context) error {
	agentID := c.Param("id")
	if agentID == "" {
		return echo.NewHTTPError(http.StatusBadRequest, "agent_id is required")
	}

	wallet, err := h.walletService.GetWallet(c.Request().Context(), agentID)
	if err != nil {
		return echo.NewHTTPError(http.StatusInternalServerError, err.Error())
	}

	return c.JSON(http.StatusOK, map[string]any{
		"ok":      true,
		"wallet": wallet,
	})
}

// CreateListing handles POST /v1/marketplace/agent/list
func (h *SwarmHandler) CreateListing(c echo.Context) error {
	var req marketplace.CreateAgentListingRequest
	if err := c.Bind(&req); err != nil {
		return echo.NewHTTPError(http.StatusBadRequest, err.Error())
	}

	listing, err := h.marketplaceService.ListingAgent(c.Request().Context(), &req)
	if err != nil {
		return echo.NewHTTPError(http.StatusInternalServerError, err.Error())
	}

	return c.JSON(http.StatusCreated, map[string]any{
		"ok":      true,
		"listing": listing,
	})
}

// SearchAgents handles GET /v1/marketplace/agents
func (h *SwarmHandler) SearchAgents(c echo.Context) error {
	var req marketplace.SearchAgentsRequest
	if err := c.Bind(&req); err != nil {
		return echo.NewHTTPError(http.StatusBadRequest, err.Error())
	}

	if req.Limit == 0 {
		req.Limit = 20
	}

	results, total, err := h.marketplaceService.SearchAgents(c.Request().Context(), &req)
	if err != nil {
		return echo.NewHTTPError(http.StatusInternalServerError, err.Error())
	}

	return c.JSON(http.StatusOK, map[string]any{
		"ok":    true,
		"agents": results,
		"total": total,
	})
}

// ProposeEvolution handles POST /v1/agent/:id/evolve
func (h *SwarmHandler) ProposeEvolution(c echo.Context) error {
	agentID := c.Param("id")
	if agentID == "" {
		return echo.NewHTTPError(http.StatusBadRequest, "agent_id is required")
	}

	// First analyze performance
	analysis, err := h.evolutionService.AnalyzePerformance(c.Request().Context(), agentID, 24*7*time.Hour) // 1 week
	if err != nil {
		return echo.NewHTTPError(http.StatusInternalServerError, err.Error())
	}

	// Then propose evolution
	proposal, err := h.evolutionService.ProposeEvolution(c.Request().Context(), agentID, analysis)
	if err != nil {
		return echo.NewHTTPError(http.StatusInternalServerError, err.Error())
	}

	return c.JSON(http.StatusCreated, map[string]any{
		"ok":        true,
		"proposal":  proposal,
		"analysis":  analysis,
	})
}

// CreateSchedule handles POST /v1/agent/:id/schedule
func (h *SwarmHandler) CreateSchedule(c echo.Context) error {
	agentID := c.Param("id")
	if agentID == "" {
		return echo.NewHTTPError(http.StatusBadRequest, "agent_id is required")
	}

	var schedule identity.AutonomySchedule
	if err := c.Bind(&schedule); err != nil {
		return echo.NewHTTPError(http.StatusBadRequest, err.Error())
	}

	schedule.AgentID = agentID

	if err := h.autonomyService.CreateSchedule(c.Request().Context(), &schedule); err != nil {
		return echo.NewHTTPError(http.StatusInternalServerError, err.Error())
	}

	return c.JSON(http.StatusCreated, map[string]any{
		"ok":       true,
		"schedule": schedule,
	})
}

// GetSchedules handles GET /v1/agent/:id/schedules
func (h *SwarmHandler) GetSchedules(c echo.Context) error {
	agentID := c.Param("id")
	if agentID == "" {
		return echo.NewHTTPError(http.StatusBadRequest, "agent_id is required")
	}

	schedules, err := h.autonomyService.GetSchedules(c.Request().Context(), agentID)
	if err != nil {
		return echo.NewHTTPError(http.StatusInternalServerError, err.Error())
	}

	return c.JSON(http.StatusOK, map[string]any{
		"ok":        true,
		"schedules": schedules,
	})
}

// RegisterRoutes registers swarm routes
func (h *SwarmHandler) RegisterRoutes(e *echo.Echo, basePath string) {
	agent := e.Group(basePath + "/agent")

	// Spawn and children
	agent.POST("/:id/spawn", h.SpawnChild)
	agent.GET("/:id/children", h.GetChildren)
	agent.GET("/:id/parent", h.GetParent)

	// Messaging
	agent.POST("/:id/message", h.SendMessage)
	agent.GET("/:id/inbox", h.GetInbox)

	// Wallet
	agent.GET("/:id/wallet", h.GetWallet)

	// Evolution
	agent.POST("/:id/evolve", h.ProposeEvolution)

	// Autonomy
	agent.POST("/:id/schedule", h.CreateSchedule)
	agent.GET("/:id/schedules", h.GetSchedules)

	// Marketplace
	marketplace := e.Group(basePath + "/marketplace")
	marketplace.GET("/agents", h.SearchAgents)
	marketplace.POST("/agent/list", h.CreateListing)
}

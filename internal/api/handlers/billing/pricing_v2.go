package billing

import (
	"encoding/json"
	"fmt"
	"net/http"
	"time"

	"github.com/functionfly/functionfly/internal/api/middleware"
	"github.com/functionfly/functionfly/internal/storage"
	"github.com/google/uuid"
	"github.com/sirupsen/logrus"
)

// UsageBasedPricingResponse shows the metered pricing details for the 3-layer strategy
type UsageBasedPricingResponse struct {
	EventType     string      `json:"event_type"`
	DisplayName   string      `json:"display_name"`
	Description   string      `json:"description"`
	Unit          string      `json:"unit"`
	Tiers         []PriceTier `json:"tiers"`
	FreeAllowance int         `json:"free_allowance"`
	AIConfig      *AIConfig   `json:"ai_config,omitempty"`
}

type PriceTier struct {
	MinQuantity  *int   `json:"min_quantity"`
	MaxQuantity  *int   `json:"max_quantity"`
	PricePerUnit int    `json:"price_per_unit_cents"` // cents per million units
	DisplayPrice string `json:"display_price"`
}

type AIConfig struct {
	MarkupPercent int    `json:"markup_percent"`
	Description   string `json:"description"`
}

// CurrentUsageResponse shows current period usage with the 3-layer pricing breakdown
type CurrentUsageResponse struct {
	TenantID         uuid.UUID              `json:"tenant_id"`
	PeriodStart      time.Time              `json:"period_start"`
	PeriodEnd        time.Time              `json:"period_end"`
	Tier             string                 `json:"tier"`
	TierLimits       map[string]interface{} `json:"tier_limits"`
	Usage            map[string]UsageMetric `json:"usage"`
	Costs            UsageCosts             `json:"costs"`
	FreeAllowances   map[string]int         `json:"free_allowances"`
	IsOverLimit      bool                   `json:"is_over_limit"`
	ApproachingLimit bool                   `json:"approaching_limit"`
}

type UsageMetric struct {
	Used      int     `json:"used"`
	Limit     int     `json:"limit"`
	Remaining int     `json:"remaining"`
	Percent   float64 `json:"percent"`
}

type UsageCosts struct {
	BaseCostCents         int    `json:"base_cost_cents"` // Subscription cost
	ExecutionOverageCents int    `json:"execution_overage_cents"`
	AICallsCents          int    `json:"ai_calls_cents"`
	StateUsageCents       int    `json:"state_usage_cents"`
	WorkflowRunsCents     int    `json:"workflow_runs_cents"`
	TotalCents            int    `json:"total_cents"`
	TotalUSD              string `json:"total_usd"`
}

// HandleGetUsagePricing returns the usage-based pricing details for the 3-layer strategy
// GET /v1/billing/usage-pricing
func (h *Handler) HandleGetUsagePricing(w http.ResponseWriter, r *http.Request) {
	claims := middleware.GetUserFromContext(r)
	if claims == nil {
		writeJSONError(w, http.StatusUnauthorized, "Unauthorized")
		return
	}

	response := struct {
		UsageBased     []UsageBasedPricingResponse `json:"usage_based_pricing"`
		FreeAllowances map[string]int              `json:"free_allowances"`
		Description    string                      `json:"description"`
	}{
		UsageBased: getDefaultUsageBasedPricing(),
		FreeAllowances: map[string]int{
			"function_executions": 100000,
			"ai_calls":            1000,
			"storage_gb":          1,
			"workflow_runs":       1000,
		},
		Description: "Pay only for what you use above generous free allowances. All tiers include 100K executions, 1K AI calls, 1GB storage, and 1K workflow runs per month.",
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(response)
}

// getDefaultUsageBasedPricing returns the pay-as-you-go pricing structure
func getDefaultUsageBasedPricing() []UsageBasedPricingResponse {
	return []UsageBasedPricingResponse{
		{
			EventType:     "function_execution",
			DisplayName:   "Function Executions",
			Description:   "Pay only for what you use. First 100K included free every month.",
			Unit:          "per million executions",
			FreeAllowance: 100000,
			Tiers: []PriceTier{
				{MinQuantity: intPtr(0), MaxQuantity: intPtr(100000), PricePerUnit: 0, DisplayPrice: "Free (up to 100K)"},
				{MinQuantity: intPtr(100001), MaxQuantity: intPtr(1000000), PricePerUnit: 20, DisplayPrice: "$0.20 per million"},
				{MinQuantity: intPtr(1000001), MaxQuantity: intPtr(10000000), PricePerUnit: 15, DisplayPrice: "$0.15 per million"},
				{MinQuantity: intPtr(10000001), MaxQuantity: nil, PricePerUnit: 10, DisplayPrice: "$0.10 per million"},
			},
		},
		{
			EventType:     "ai_call",
			DisplayName:   "AI Model Calls",
			Description:   "Pass-through pricing from OpenRouter with 25% markup (20% on Pro/Team plans). First 1,000 calls free.",
			Unit:          "per 1K calls + actual token cost",
			FreeAllowance: 1000,
			AIConfig: &AIConfig{
				MarkupPercent: 25,
				Description:   "Pass-through from OpenRouter + 25% markup",
			},
			Tiers: []PriceTier{
				{MinQuantity: intPtr(0), MaxQuantity: intPtr(1000), PricePerUnit: 0, DisplayPrice: "Free (up to 1K)"},
				{MinQuantity: intPtr(1001), MaxQuantity: nil, PricePerUnit: 100, DisplayPrice: "$1.00 per 1K + token cost"},
			},
		},
		{
			EventType:     "state_usage",
			DisplayName:   "State Storage & Operations",
			Description:   "Database storage, reads, and writes. First 1GB and 10K writes included.",
			Unit:          "per GB storage + per 1K operations",
			FreeAllowance: 1, // GB
			Tiers: []PriceTier{
				{MinQuantity: intPtr(0), MaxQuantity: intPtr(1), PricePerUnit: 0, DisplayPrice: "Free (up to 1GB)"},
				{MinQuantity: intPtr(2), MaxQuantity: nil, PricePerUnit: 50, DisplayPrice: "$0.50 per GB"},
			},
		},
		{
			EventType:     "vector_search",
			DisplayName:   "Vector Search Queries",
			Description:   "Semantic search and vector similarity queries. First 1,000 included.",
			Unit:          "per 1K queries",
			FreeAllowance: 1000,
			Tiers: []PriceTier{
				{MinQuantity: intPtr(0), MaxQuantity: intPtr(1000), PricePerUnit: 0, DisplayPrice: "Free (up to 1K)"},
				{MinQuantity: intPtr(1001), MaxQuantity: nil, PricePerUnit: 500, DisplayPrice: "$5.00 per 1K"},
			},
		},
		{
			EventType:     "workflow_run",
			DisplayName:   "Workflow Runs",
			Description:   "Graph workflow executions. First 1,000 runs included.",
			Unit:          "per 100 runs",
			FreeAllowance: 1000,
			Tiers: []PriceTier{
				{MinQuantity: intPtr(0), MaxQuantity: intPtr(1000), PricePerUnit: 0, DisplayPrice: "Free (up to 1K)"},
				{MinQuantity: intPtr(1001), MaxQuantity: nil, PricePerUnit: 200, DisplayPrice: "$2.00 per 100"},
			},
		},
	}
}

// HandleGetCurrentUsage returns current period usage with the 3-layer pricing breakdown
// GET /v1/billing/usage/current
func (h *Handler) HandleGetCurrentUsage(w http.ResponseWriter, r *http.Request) {
	claims := middleware.GetUserFromContext(r)
	if claims == nil || claims.TenantID == uuid.Nil {
		writeJSONError(w, http.StatusUnauthorized, "Unauthorized")
		return
	}

	// Get subscription for tier info and period
	sub, err := h.repo.GetSubscriptionByTenantID(claims.TenantID)
	if err != nil {
		logrus.WithError(err).WithField("tenant_id", claims.TenantID).Warn("billing: failed to get subscription")
		writeJSONError(w, http.StatusInternalServerError, "Failed to get subscription")
		return
	}

	// Get usage from database for current period
	now := time.Now().UTC()
	periodStart := time.Date(now.Year(), now.Month(), 1, 0, 0, 0, 0, time.UTC)
	periodEnd := periodStart.AddDate(0, 1, 0).Add(-time.Second)

	// Pull usage metrics from existing rollups
	execRollups, _ := h.repo.GetUsageByTenant(claims.TenantID, "function_execution", periodStart, periodEnd)
	aiRollups, _ := h.repo.GetUsageByTenant(claims.TenantID, "ai_call", periodStart, periodEnd)
	workflowRollups, _ := h.repo.GetUsageByTenant(claims.TenantID, "workflow_run", periodStart, periodEnd)

	totalExecutions := 0
	for _, rollup := range execRollups {
		totalExecutions += rollup.TotalQuantity
	}

	totalAICalls := 0
	for _, rollup := range aiRollups {
		totalAICalls += rollup.TotalQuantity
	}

	totalWorkflows := 0
	for _, rollup := range workflowRollups {
		totalWorkflows += rollup.TotalQuantity
	}

	// Get tier limits from features JSON
	tierName := "Free"
	executionsLimit := 100000
	aiCallsLimit := 1000
	workflowLimit := 1000
	storageLimitGB := 1
	baseCostCents := 0

	if sub != nil && sub.PricingTier != nil {
		tierName = sub.PricingTier.Name
		baseCostCents = sub.PricingTier.PriceCents
		if sub.PricingTier.Features != nil {
			features := make(map[string]interface{})
			if f, ok := sub.PricingTier.Features.(map[string]interface{}); ok {
				features = f
			} else if data, err := json.Marshal(sub.PricingTier.Features); err == nil {
				_ = json.Unmarshal(data, &features)
			}

			if v, ok := features["requests"].(float64); ok {
				executionsLimit = int(v)
			}
			if v, ok := features["ai_calls_included"].(float64); ok {
				aiCallsLimit = int(v)
			}
			if v, ok := features["storage_gb"].(float64); ok {
				storageLimitGB = int(v)
			}
			if v, ok := features["workflows"].(float64); ok {
				workflowLimit = int(v)
			}
		}
	}

	// Calculate metrics
	execRemaining := executionsLimit - totalExecutions
	if execRemaining < 0 {
		execRemaining = 0
	}
	aiRemaining := aiCallsLimit - totalAICalls
	if aiRemaining < 0 {
		aiRemaining = 0
	}
	workflowRemaining := workflowLimit - totalWorkflows
	if workflowRemaining < 0 {
		workflowRemaining = 0
	}

	execPercent := float64(totalExecutions) / float64(executionsLimit) * 100
	if execPercent > 100 {
		execPercent = 100
	}
	aiPercent := float64(totalAICalls) / float64(aiCallsLimit) * 100
	if aiPercent > 100 {
		aiPercent = 100
	}

	workflowPercent := float64(totalWorkflows) / float64(workflowLimit) * 100
	if workflowPercent > 100 {
		workflowPercent = 100
	}

	// Calculate costs
	executionOverage := totalExecutions - executionsLimit
	if executionOverage < 0 {
		executionOverage = 0
	}
	executionOverageCents := executionOverage * 20 / 1000000 // $0.20 per million

	aiOverage := totalAICalls - aiCallsLimit
	if aiOverage < 0 {
		aiOverage = 0
	}
	aiOverageCents := aiOverage * 1 // $1 per 1K (simplified, actual uses pass-through)

	workflowOverage := totalWorkflows - workflowLimit
	if workflowOverage < 0 {
		workflowOverage = 0
	}
	workflowOverageCents := workflowOverage * 2 / 100 // $2 per 100

	totalCents := baseCostCents + executionOverageCents + aiOverageCents + workflowOverageCents

	// Build response
	response := CurrentUsageResponse{
		TenantID:    claims.TenantID,
		PeriodStart: periodStart,
		PeriodEnd:   periodEnd,
		Tier:        tierName,
		TierLimits: map[string]interface{}{
			"executions": executionsLimit,
			"ai_calls":   aiCallsLimit,
			"storage_gb": storageLimitGB,
			"workflows":  workflowLimit,
		},
		Usage: map[string]UsageMetric{
			"executions": {
				Used:      totalExecutions,
				Limit:     executionsLimit,
				Remaining: execRemaining,
				Percent:   execPercent,
			},
			"ai_calls": {
				Used:      totalAICalls,
				Limit:     aiCallsLimit,
				Remaining: aiRemaining,
				Percent:   aiPercent,
			},
			"workflows": {
				Used:      totalWorkflows,
				Limit:     workflowLimit,
				Remaining: workflowRemaining,
				Percent:   workflowPercent,
			},
		},
		Costs: UsageCosts{
			BaseCostCents:         baseCostCents,
			ExecutionOverageCents: executionOverageCents,
			AICallsCents:          aiOverageCents,
			StateUsageCents:       0,
			WorkflowRunsCents:     workflowOverageCents,
			TotalCents:            totalCents,
			TotalUSD:              formatCents(totalCents),
		},
		FreeAllowances: map[string]int{
			"function_executions": 100000,
			"ai_calls":            1000,
			"storage_gb":          1,
			"workflow_runs":       1000,
		},
		IsOverLimit:      executionOverage > 0 || aiOverage > 0 || workflowOverage > 0,
		ApproachingLimit: execPercent > 80 || aiPercent > 80 || workflowPercent > 80,
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(response)
}

// formatCents converts cents to a USD string
func formatCents(cents int) string {
	dollars := float64(cents) / 100.0
	if cents%100 == 0 {
		return fmt.Sprintf("$%.0f", dollars)
	}
	return fmt.Sprintf("$%.2f", dollars)
}

func intPtr(i int) *int {
	return &i
}

// RecordUsageEventRequest is used to record a metered usage event
type RecordUsageEventRequest struct {
	EventType          string                 `json:"event_type"` // function_execution, ai_call, state_read, state_write, vector_search, workflow_run
	Quantity           int                    `json:"quantity"`
	AIModel            string                 `json:"ai_model,omitempty"`
	AIInputTokens      int                    `json:"ai_input_tokens,omitempty"`
	AIOutputTokens     int                    `json:"ai_output_tokens,omitempty"`
	AICostUSD          float64                `json:"ai_cost_usd,omitempty"`
	StateOperation     string                 `json:"state_operation,omitempty"` // read, write, query, vector_search
	StorageBytes       int                    `json:"storage_bytes,omitempty"`
	WorkflowID         *uuid.UUID             `json:"workflow_id,omitempty"`
	WorkflowComplexity string                 `json:"workflow_complexity,omitempty"` // simple, standard, complex
	Metadata           map[string]interface{} `json:"metadata,omitempty"`
}

// HandleRecordUsage records a usage event for metered billing (internal/admin endpoint)
// POST /v1/billing/usage/record
func (h *Handler) HandleRecordUsage(w http.ResponseWriter, r *http.Request) {
	claims := middleware.GetUserFromContext(r)
	if claims == nil || claims.TenantID == uuid.Nil {
		writeJSONError(w, http.StatusUnauthorized, "Unauthorized")
		return
	}

	var req RecordUsageEventRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		writeJSONError(w, http.StatusBadRequest, "Invalid request body")
		return
	}

	if req.EventType == "" || req.Quantity <= 0 {
		writeJSONError(w, http.StatusBadRequest, "event_type and quantity are required")
		return
	}

	// Create usage event
	event := &storage.UsageEvent{
		ID:        uuid.New(),
		TenantID:  claims.TenantID,
		EventType: req.EventType,
		Quantity:  req.Quantity,
		Metadata:  req.Metadata,
		Timestamp: time.Now().UTC(),
	}

	if err := h.repo.RecordUsageEvent(r.Context(), event); err != nil {
		logrus.WithError(err).WithField("tenant_id", claims.TenantID).Error("billing: failed to record usage")
		writeJSONError(w, http.StatusInternalServerError, "Failed to record usage")
		return
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(map[string]interface{}{
		"event_id":  event.ID,
		"recorded":  true,
		"timestamp": event.Timestamp,
	})
}

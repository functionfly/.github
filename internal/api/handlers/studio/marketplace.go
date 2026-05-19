package studio

import (
	"context"
	"database/sql"
	"time"

	"github.com/google/uuid"
)

// MarketplaceFunction represents a function in the studio marketplace
type MarketplaceFunction struct {
	ID          string   `json:"id"`
	TenantID    string   `json:"tenant_id"`
	Name        string   `json:"name"`
	Description string   `json:"description"`
	Author      string   `json:"author"`
	Version     string   `json:"version"`
	Category    string   `json:"category"`
	Downloads   int      `json:"downloads"`
	Rating      float64  `json:"rating"`
	Price       float64  `json:"price"`
	IsFavorite  bool     `json:"is_favorite"`
	Runtime     string   `json:"runtime"`
	Triggers    []string `json:"triggers"`
	CreatedAt   time.Time `json:"created_at"`
	UpdatedAt   time.Time `json:"updated_at"`
}

// SubscriptionPlan represents a subscription plan in the marketplace
type SubscriptionPlan struct {
	ID         string   `json:"id"`
	TenantID   string   `json:"tenant_id"`
	Name       string   `json:"name"`
	Price      float64  `json:"price"`
	Features   []string `json:"features"`
	Subscribers int     `json:"subscribers"`
	CreatedAt  time.Time `json:"created_at"`
	UpdatedAt  time.Time `json:"updated_at"`
}

// RoyaltyEntry represents a royalty payment entry
type RoyaltyEntry struct {
	ID           string    `json:"id"`
	TenantID     string    `json:"tenant_id"`
	FunctionID   string    `json:"function_id"`
	FunctionName string    `json:"function_name"`
	Period       string    `json:"period"`
	Calls        int       `json:"calls"`
	RoyaltyRate  float64   `json:"royalty_rate"`
	Earnings     float64   `json:"earnings"`
	Paid         bool      `json:"paid"`
	PaidAt       *time.Time `json:"paid_at,omitempty"`
	CreatedAt    time.Time `json:"created_at"`
}

// MarketplaceRepository handles database operations for studio marketplace
type MarketplaceRepository struct {
	db *sql.DB
}

// NewMarketplaceRepository creates a new marketplace repository
func NewMarketplaceRepository(db *sql.DB) *MarketplaceRepository {
	return &MarketplaceRepository{db: db}
}

// ListFunctionsParams contains parameters for listing functions
type ListFunctionsParams struct {
	TenantID  string
	Search    *string
	Category  *string
	Limit     int
	Offset    int
}

// ListFunctions returns marketplace functions filtered by tenant and optional filters
func (r *MarketplaceRepository) ListFunctions(ctx context.Context, params ListFunctionsParams) ([]MarketplaceFunction, error) {
	if params.Limit <= 0 {
		params.Limit = 50
	}
	if params.Limit > 200 {
		params.Limit = 200
	}
	if params.Offset < 0 {
		params.Offset = 0
	}

	// For now, return mock functions - in production this would query the functions table
	// with marketplace metadata joined in
	mockFunctions := []MarketplaceFunction{
		{
			ID:          "fn-1",
			TenantID:    params.TenantID,
			Name:        "data-processor",
			Description: "Process and transform data with AI",
			Author:      "FunctionFly",
			Version:     "1.0.0",
			Category:    "data",
			Downloads:   4521,
			Rating:      4.7,
			Price:       0.01,
			Runtime:     "wasm",
			Triggers:    []string{"http", "schedule"},
		},
		{
			ID:          "fn-2",
			TenantID:    params.TenantID,
			Name:        "auth-helper",
			Description: "Authentication and authorization helper",
			Author:      "AI Solutions",
			Version:     "2.1.0",
			Category:    "security",
			Downloads:   2890,
			Rating:      4.5,
			Price:       0.02,
			Runtime:     "wasm",
			Triggers:    []string{"http"},
		},
	}
	return mockFunctions, nil
}

// ExecuteFunction records a function execution
func (r *MarketplaceRepository) ExecuteFunction(ctx context.Context, functionID string, input map[string]interface{}) (string, error) {
	executionID := uuid.New().String()
	// In production, this would record the execution in a function_executions table
	return executionID, nil
}

// SetFavorite sets or clears a function as favorite
func (r *MarketplaceRepository) SetFavorite(ctx context.Context, tenantID, functionID string, favorite bool) error {
	// In production, this would update a user_favorites table
	return nil
}

// ListPlans returns subscription plans for a tenant
func (r *MarketplaceRepository) ListPlans(ctx context.Context, tenantID string) ([]SubscriptionPlan, error) {
	plans := []SubscriptionPlan{
		{
			ID:         "plan-1",
			TenantID:   tenantID,
			Name:       "Free",
			Price:      0,
			Features:   []string{"5 calls/min", "Basic support"},
			Subscribers: 1247,
		},
		{
			ID:         "plan-2",
			TenantID:   tenantID,
			Name:       "Pro",
			Price:      9.99,
			Features:   []string{"100 calls/min", "Priority support", "Analytics"},
			Subscribers: 342,
		},
		{
			ID:         "plan-3",
			TenantID:   tenantID,
			Name:       "Enterprise",
			Price:      49.99,
			Features:   []string{"Unlimited", "Dedicated support", "Custom SLA"},
			Subscribers: 28,
		},
	}
	return plans, nil
}

// CreatePlan creates a new subscription plan
func (r *MarketplaceRepository) CreatePlan(ctx context.Context, plan *SubscriptionPlan) error {
	if plan.ID == "" {
		plan.ID = uuid.New().String()
	}
	plan.CreatedAt = time.Now()
	plan.UpdatedAt = time.Now()
	// In production, this would insert into a subscription_plans table
	return nil
}

// UpdatePlan updates an existing subscription plan
func (r *MarketplaceRepository) UpdatePlan(ctx context.Context, plan *SubscriptionPlan) error {
	plan.UpdatedAt = time.Now()
	// In production, this would update the subscription_plans table
	return nil
}

// ListRoyalties returns royalty entries for a tenant
func (r *MarketplaceRepository) ListRoyalties(ctx context.Context, tenantID string) ([]RoyaltyEntry, float64, float64, error) {
	royalties := []RoyaltyEntry{
		{
			ID:           "roy-1",
			TenantID:     tenantID,
			FunctionID:   "fn-1",
			FunctionName: "data-processor",
			Period:       "Jun 2026",
			Calls:        4521,
			RoyaltyRate:  0.7,
			Earnings:     3164.78,
			Paid:         false,
		},
		{
			ID:           "roy-2",
			TenantID:     tenantID,
			FunctionID:   "fn-2",
			FunctionName: "auth-helper",
			Period:       "Jun 2026",
			Calls:        2890,
			RoyaltyRate:  0.7,
			Earnings:     2023.32,
			Paid:         false,
		},
		{
			ID:           "roy-3",
			TenantID:     tenantID,
			FunctionID:   "fn-3",
			FunctionName: "image-transform",
			Period:       "Jun 2026",
			Calls:        5435,
			RoyaltyRate:  0.7,
			Earnings:     3805.17,
			Paid:         true,
		},
	}
	totalEarnings := 8993.27
	pendingPayout := 5823.32
	return royalties, totalEarnings, pendingPayout, nil
}

// RequestPayout requests a payout for pending royalties
func (r *MarketplaceRepository) RequestPayout(ctx context.Context, tenantID string) error {
	// In production, this would create a payout request record
	return nil
}

// UpdateLicense updates the license for a function
func (r *MarketplaceRepository) UpdateLicense(ctx context.Context, tenantID, functionID, license string) error {
	// In production, this would update the functions table with the new license
	return nil
}

// UpdatePricing updates the pricing for a function
func (r *MarketplaceRepository) UpdatePricing(ctx context.Context, tenantID, functionID string, price float64, model string) error {
	// In production, this would update the functions table with new pricing
	return nil
}
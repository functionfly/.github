package billing

import (
	"context"
	"fmt"
	"time"

	"github.com/google/uuid"
	"github.com/redis/go-redis/v9"
	"gorm.io/gorm"
)

// AgentBillingControls holds the economic controls for an agent
type AgentBillingControls struct {
	ID                 uuid.UUID          `json:"id" gorm:"type:uuid;primaryKey;default:gen_random_uuid()"`
	AgentID            string             `json:"agent_id" gorm:"uniqueIndex;not null"`
	SpendCapMonthlyUSD *float64           `json:"spend_cap_monthly_usd,omitempty" gorm:"type:decimal(10,2)"`
	SpendCapDailyUSD   *float64           `json:"spend_cap_daily_usd,omitempty" gorm:"type:decimal(10,2)"`
	CreditBalanceUSD   float64            `json:"credit_balance_usd" gorm:"type:decimal(10,2);not null;default:0"`
	BillingMode        string             `json:"billing_mode" gorm:"not null;default:'per_agent'"` // per_agent | per_tenant | per_team
	TeamID             *uuid.UUID         `json:"team_id,omitempty" gorm:"type:uuid"`
	AlertThresholds    []float64          `json:"alert_thresholds" gorm:"type:decimal[]"`
	CreatedAt          time.Time          `json:"created_at" gorm:"autoCreateTime"`
	UpdatedAt          time.Time          `json:"updated_at" gorm:"autoUpdateTime"`
}

// TableName returns the GORM table name
func (AgentBillingControls) TableName() string {
	return "agent_billing_controls"
}

// SpendSummary holds a summary of agent spending
type SpendSummary struct {
	AgentID          string    `json:"agent_id"`
	Period           string    `json:"period"`
	TotalCallsUSD    float64   `json:"total_calls_usd"`
	CreditBalance    float64   `json:"credit_balance_usd"`
	SpendCapMonthly  *float64  `json:"spend_cap_monthly_usd,omitempty"`
	SpendCapDaily    *float64  `json:"spend_cap_daily_usd,omitempty"`
	CapUtilization   float64   `json:"cap_utilization_pct"` // 0-100
	GeneratedAt      time.Time `json:"generated_at"`
}

// CreditPurchaseRequest is the request to pre-purchase execution credits
type CreditPurchaseRequest struct {
	AgentID         string  `json:"agent_id"`
	AmountUSD       float64 `json:"amount_usd"`
	PaymentMethodID string  `json:"payment_method_id"`
}

// BillingMode constants
const (
	BillingModePerAgent  = "per_agent"
	BillingModePerTenant = "per_tenant"
	BillingModePerTeam   = "per_team"
)

// Controller manages economic controls for agents
type Controller struct {
	redis *redis.Client
	db    *gorm.DB
}

// NewController creates a new economic controller
func NewController(db *gorm.DB, redisClient *redis.Client) *Controller {
	return &Controller{
		redis: redisClient,
		db:    db,
	}
}

// GetOrCreateControls retrieves or creates billing controls for an agent
func (c *Controller) GetOrCreateControls(ctx context.Context, agentID string) (*AgentBillingControls, error) {
	var controls AgentBillingControls
	err := c.db.WithContext(ctx).Where("agent_id = ?", agentID).First(&controls).Error
	if err == gorm.ErrRecordNotFound {
		// Create default controls
		controls = AgentBillingControls{
			ID:              uuid.New(),
			AgentID:         agentID,
			CreditBalanceUSD: 0,
			BillingMode:     BillingModePerAgent,
			AlertThresholds: []float64{0.5, 0.8, 0.95},
		}
		if err := c.db.WithContext(ctx).Create(&controls).Error; err != nil {
			return nil, fmt.Errorf("failed to create billing controls: %w", err)
		}
		return &controls, nil
	}
	if err != nil {
		return nil, fmt.Errorf("failed to get billing controls: %w", err)
	}
	return &controls, nil
}

// CheckSpendCap checks if an agent has sufficient budget for an estimated cost
func (c *Controller) CheckSpendCap(ctx context.Context, agentID string, estimatedCost float64) (bool, error) {
	controls, err := c.GetOrCreateControls(ctx, agentID)
	if err != nil {
		return true, nil // Non-fatal: allow execution if controls can't be loaded
	}

	// Check daily spend cap
	if controls.SpendCapDailyUSD != nil && *controls.SpendCapDailyUSD > 0 {
		dailySpend, err := c.getDailySpend(ctx, agentID)
		if err == nil && dailySpend+estimatedCost > *controls.SpendCapDailyUSD {
			return false, fmt.Errorf("daily spend cap of $%.2f would be exceeded (current: $%.4f, estimated: $%.6f)",
				*controls.SpendCapDailyUSD, dailySpend, estimatedCost)
		}
	}

	// Check monthly spend cap
	if controls.SpendCapMonthlyUSD != nil && *controls.SpendCapMonthlyUSD > 0 {
		monthlySpend, err := c.getMonthlySpend(ctx, agentID)
		if err == nil && monthlySpend+estimatedCost > *controls.SpendCapMonthlyUSD {
			return false, fmt.Errorf("monthly spend cap of $%.2f would be exceeded (current: $%.4f, estimated: $%.6f)",
				*controls.SpendCapMonthlyUSD, monthlySpend, estimatedCost)
		}
	}

	return true, nil
}

// ConsumeCredits deducts from the agent's credit balance
func (c *Controller) ConsumeCredits(ctx context.Context, agentID string, amount float64) error {
	if amount <= 0 {
		return nil
	}

	result := c.db.WithContext(ctx).Model(&AgentBillingControls{}).
		Where("agent_id = ? AND credit_balance_usd >= ?", agentID, amount).
		Update("credit_balance_usd", gorm.Expr("credit_balance_usd - ?", amount))

	if result.Error != nil {
		return fmt.Errorf("failed to consume credits: %w", result.Error)
	}
	if result.RowsAffected == 0 {
		return fmt.Errorf("insufficient credit balance for agent %s (required: $%.6f)", agentID, amount)
	}

	return nil
}

// AddCredits adds to the agent's credit balance (after purchase)
func (c *Controller) AddCredits(ctx context.Context, agentID string, amount float64) error {
	if amount <= 0 {
		return fmt.Errorf("credit amount must be positive")
	}

	result := c.db.WithContext(ctx).Model(&AgentBillingControls{}).
		Where("agent_id = ?", agentID).
		Update("credit_balance_usd", gorm.Expr("credit_balance_usd + ?", amount))

	if result.Error != nil {
		return fmt.Errorf("failed to add credits: %w", result.Error)
	}
	if result.RowsAffected == 0 {
		return fmt.Errorf("agent billing controls not found: %s", agentID)
	}

	return nil
}

// GetAgentSpend returns the spend summary for an agent for a given period
func (c *Controller) GetAgentSpend(ctx context.Context, agentID string, period string) (*SpendSummary, error) {
	controls, err := c.GetOrCreateControls(ctx, agentID)
	if err != nil {
		return nil, err
	}

	var totalSpend float64
	since := periodToTime(period)

	err = c.db.WithContext(ctx).
		Table("agent_execution_records").
		Where("agent_id = ? AND timestamp >= ?", agentID, since).
		Select("COALESCE(SUM(cost_usd), 0)").
		Scan(&totalSpend).Error
	if err != nil {
		return nil, fmt.Errorf("failed to calculate spend: %w", err)
	}

	summary := &SpendSummary{
		AgentID:         agentID,
		Period:          period,
		TotalCallsUSD:   totalSpend,
		CreditBalance:   controls.CreditBalanceUSD,
		SpendCapMonthly: controls.SpendCapMonthlyUSD,
		SpendCapDaily:   controls.SpendCapDailyUSD,
		GeneratedAt:     time.Now(),
	}

	// Calculate cap utilization
	if controls.SpendCapMonthlyUSD != nil && *controls.SpendCapMonthlyUSD > 0 {
		summary.CapUtilization = (totalSpend / *controls.SpendCapMonthlyUSD) * 100
	}

	return summary, nil
}

// UpdateSpendCap updates the spend cap for an agent
func (c *Controller) UpdateSpendCap(ctx context.Context, agentID string, dailyCap, monthlyCap *float64) error {
	updates := map[string]interface{}{
		"updated_at": time.Now(),
	}
	if dailyCap != nil {
		updates["spend_cap_daily_usd"] = *dailyCap
	}
	if monthlyCap != nil {
		updates["spend_cap_monthly_usd"] = *monthlyCap
	}

	result := c.db.WithContext(ctx).Model(&AgentBillingControls{}).
		Where("agent_id = ?", agentID).
		Updates(updates)

	if result.Error != nil {
		return fmt.Errorf("failed to update spend cap: %w", result.Error)
	}
	if result.RowsAffected == 0 {
		return fmt.Errorf("billing controls not found for agent: %s", agentID)
	}

	return nil
}

// getDailySpend returns the total spend for today from Redis cache
func (c *Controller) getDailySpend(ctx context.Context, agentID string) (float64, error) {
	if c.redis == nil {
		return 0, nil
	}
	now := time.Now().UTC()
	key := fmt.Sprintf("billing:spend:day:%s:%d%02d%02d", agentID, now.Year(), now.Month(), now.Day())
	val, err := c.redis.Get(ctx, key).Float64()
	if err == redis.Nil {
		return 0, nil
	}
	return val, err
}

// getMonthlySpend returns the total spend for this month from the DB
func (c *Controller) getMonthlySpend(ctx context.Context, agentID string) (float64, error) {
	now := time.Now().UTC()
	monthStart := time.Date(now.Year(), now.Month(), 1, 0, 0, 0, 0, time.UTC)

	var totalSpend float64
	err := c.db.WithContext(ctx).
		Table("agent_execution_records").
		Where("agent_id = ? AND timestamp >= ?", agentID, monthStart).
		Select("COALESCE(SUM(cost_usd), 0)").
		Scan(&totalSpend).Error

	return totalSpend, err
}

// periodToTime converts a period string (e.g. "2024-01") to a time.Time
func periodToTime(period string) time.Time {
	if period == "" {
		// Default: last 30 days
		return time.Now().UTC().AddDate(0, 0, -30)
	}

	// Try parsing as YYYY-MM
	var year, month int
	if n, _ := fmt.Sscanf(period, "%d-%d", &year, &month); n == 2 {
		return time.Date(year, time.Month(month), 1, 0, 0, 0, 0, time.UTC)
	}

	// Default: last 30 days
	return time.Now().UTC().AddDate(0, 0, -30)
}

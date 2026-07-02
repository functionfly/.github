package billing

import (
	"context"
	"fmt"
	"os"
	"strconv"
	"time"

	"github.com/google/uuid"
	"github.com/lib/pq"
	"github.com/redis/go-redis/v9"
	"gorm.io/gorm"
	"gorm.io/gorm/clause"
)

// AgentBillingControls holds the economic controls for an agent
type AgentBillingControls struct {
	ID                  uuid.UUID     `json:"id" gorm:"type:uuid;primaryKey;default:gen_random_uuid()"`
	AgentID             string        `json:"agent_id" gorm:"uniqueIndex;not null"`
	SpendCapMonthlyUSD  *float64      `json:"spend_cap_monthly_usd,omitempty" gorm:"type:decimal(10,2)"`
	SpendCapWeeklyUSD   *float64      `json:"spend_cap_weekly_usd,omitempty" gorm:"type:decimal(10,2)"`
	SpendCapDailyUSD    *float64      `json:"spend_cap_daily_usd,omitempty" gorm:"type:decimal(10,2)"`
	CreditBalanceUSD    float64       `json:"credit_balance_usd" gorm:"type:decimal(10,2);not null;default:0"`
	BillingMode         string        `json:"billing_mode" gorm:"not null;default:'per_agent'"` // per_agent | per_tenant | per_team
	TeamID              *uuid.UUID    `json:"team_id,omitempty" gorm:"type:uuid"`
	AlertThresholds     pq.Float64Array `json:"alert_thresholds" gorm:"type:decimal[]"`
	CreatedAt           time.Time     `json:"created_at" gorm:"autoCreateTime"`
	UpdatedAt           time.Time     `json:"updated_at" gorm:"autoUpdateTime"`
}

// TableName returns the GORM table name
func (AgentBillingControls) TableName() string {
	return "agent_billing_controls"
}

// SpendSummary holds a summary of agent spending
type SpendSummary struct {
	AgentID         string    `json:"agent_id"`
	Period          string    `json:"period"`
	TotalCallsUSD   float64   `json:"total_calls_usd"`
	CreditBalance   float64   `json:"credit_balance_usd"`
	SpendCapMonthly *float64  `json:"spend_cap_monthly_usd,omitempty"`
	SpendCapWeekly  *float64  `json:"spend_cap_weekly_usd,omitempty"`
	SpendCapDaily   *float64  `json:"spend_cap_daily_usd,omitempty"`
	CapUtilization  float64   `json:"cap_utilization_pct"` // 0-100
	GeneratedAt     time.Time `json:"generated_at"`
}

// CreditBalanceUpdate captures a credit balance mutation.
type CreditBalanceUpdate struct {
	AgentID     string  `json:"agent_id"`
	AmountUSD   float64 `json:"amount_usd"`
	PreviousUSD float64 `json:"previous_balance_usd"`
	CurrentUSD  float64 `json:"current_balance_usd"`
}

// CreditPurchaseRequest is the request to pre-purchase execution credits
type CreditPurchaseRequest struct {
	AgentID         string  `json:"agent_id"`
	AmountUSD       float64 `json:"amount_usd"`
	PaymentMethodID string  `json:"payment_method_id"`
}

// BrowserUsage records a browser automation usage event
type BrowserUsage struct {
	ID              uuid.UUID `json:"id" gorm:"type:uuid;primaryKey;default:gen_random_uuid()"`
	AgentID         string    `json:"agent_id" gorm:"not null;index"`
	SessionID       uuid.UUID `json:"session_id" gorm:"type:uuid;not null"`
	Action          string    `json:"action" gorm:"not null"` // navigate, click, fill, screenshot, extract
	Domain          string    `json:"domain" gorm:"index"`
	DurationMs      int       `json:"duration_ms"`
	BrowserMinutes  float64   `json:"browser_minutes" gorm:"type:decimal(10,4)"`
	CostUSD         float64   `json:"cost_usd" gorm:"type:decimal(10,6)"`
	CreatedAt       time.Time `json:"created_at" gorm:"autoCreateTime"`
}

// TableName returns the table name
func (BrowserUsage) TableName() string {
	return "agent_browser_usage"
}

// BrowserUsageStats holds browser usage statistics
type BrowserUsageStats struct {
	AgentID      string  `json:"agent_id"`
	Period       string  `json:"period"`
	TotalMinutes float64 `json:"total_minutes"`
	TotalCostUSD float64 `json:"total_cost_usd"`
	TotalActions int64   `json:"total_actions"`
}

// BillingMode constants
const (
	BillingModePerAgent  = "per_agent"
	BillingModePerTenant = "per_tenant"
	BillingModePerTeam   = "per_team"
)

// ControllerInterface defines the interface for billing controller operations.
// This is implemented by *Controller and *BillingControllerWrapper for flexibility.
type ControllerInterface interface {
	GetOrCreateControls(ctx context.Context, agentID string) (*AgentBillingControls, error)
	ConsumeCredits(ctx context.Context, agentID string, amount float64) (*CreditBalanceUpdate, error)
}

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
			ID:               uuid.New(),
			AgentID:          agentID,
			CreditBalanceUSD: 0,
			BillingMode:      BillingModePerAgent,
			AlertThresholds:  pq.Float64Array{0.5, 0.8, 0.95},
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
		// SECURITY FIX: Deny on error instead of allowing execution.
		// Allowing execution when controls can't be loaded enables budget bypass attacks.
		return false, fmt.Errorf("failed to verify spend cap for agent %s: %w", agentID, err)
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

// ConsumeCredits deducts from the agent's credit balance transactionally.
func (c *Controller) ConsumeCredits(ctx context.Context, agentID string, amount float64) (*CreditBalanceUpdate, error) {
	if amount <= 0 {
		return &CreditBalanceUpdate{
			AgentID:     agentID,
			AmountUSD:   amount,
			PreviousUSD: 0,
			CurrentUSD:  0,
		}, nil
	}

	var update CreditBalanceUpdate
	update.AgentID = agentID

	err := c.db.WithContext(ctx).Transaction(func(tx *gorm.DB) error {
		var controls AgentBillingControls
		if err := tx.Clauses(clause.Locking{Strength: "UPDATE"}).
			Where("agent_id = ?", agentID).
			First(&controls).Error; err != nil {
			if err == gorm.ErrRecordNotFound {
				return fmt.Errorf("billing controls not found for agent: %s", agentID)
			}
			return fmt.Errorf("failed to lock billing controls: %w", err)
		}

		update.PreviousUSD = controls.CreditBalanceUSD
		if controls.CreditBalanceUSD < amount {
			return fmt.Errorf("insufficient credit balance for %s (required: $%.6f, have: $%.6f)", agentID, amount, controls.CreditBalanceUSD)
		}

		newBalance := controls.CreditBalanceUSD - amount
		if err := tx.Model(&AgentBillingControls{}).
			Where("id = ?", controls.ID).
			Update("credit_balance_usd", newBalance).Error; err != nil {
			return fmt.Errorf("failed to consume credits: %w", err)
		}

		update.CurrentUSD = newBalance
		return nil
	})
	if err != nil {
		return nil, err
	}

	return &update, nil
}

// AddCredits adds to the agent's credit balance (after purchase).
func (c *Controller) AddCredits(ctx context.Context, agentID string, amount float64) error {
	if amount <= 0 {
		return fmt.Errorf("credit amount must be positive")
	}

	err := c.db.WithContext(ctx).Transaction(func(tx *gorm.DB) error {
		var controls AgentBillingControls
		err := tx.Clauses(clause.Locking{Strength: "UPDATE"}).
			Where("agent_id = ?", agentID).
			First(&controls).Error

		if err == gorm.ErrRecordNotFound {
			// No controls exist - create new
			controls = AgentBillingControls{
				ID:              uuid.New(),
				AgentID:         agentID,
				CreditBalanceUSD: amount,
				BillingMode:     BillingModePerAgent,
			}
			if err := tx.Create(&controls).Error; err != nil {
				return fmt.Errorf("failed to create billing controls for agent: %w", err)
			}
			return nil
		}
		if err != nil {
			return fmt.Errorf("failed to get billing controls: %w", err)
		}

		// Update existing
		if err := tx.Model(&AgentBillingControls{}).
			Where("id = ?", controls.ID).
			Updates(map[string]interface{}{
				"credit_balance_usd": gorm.Expr("credit_balance_usd + ?", amount),
				"updated_at":         time.Now(),
			}).Error; err != nil {
			return fmt.Errorf("failed to add credits: %w", err)
		}

		return nil
	})

	return err
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
func (c *Controller) UpdateSpendCap(ctx context.Context, agentID string, dailyCap, weeklyCap, monthlyCap *float64) error {
	updates := map[string]interface{}{
		"updated_at": time.Now(),
	}
	if dailyCap != nil {
		updates["spend_cap_daily_usd"] = *dailyCap
	}
	if weeklyCap != nil {
		updates["spend_cap_weekly_usd"] = *weeklyCap
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

// RecordBrowserUsage records browser automation usage for cost attribution
func (c *Controller) RecordBrowserUsage(ctx context.Context, agentID string, sessionID uuid.UUID, action, domain string, durationMs int) error {
	if c.redis == nil {
		return nil
	}

	// Calculate browser minutes and cost
	browserMinutes := float64(durationMs) / 60000.0
	costPerMinute := 0.01 // Default cost per minute
	if v := os.Getenv("BROWSER_COST_PER_MINUTE"); v != "" {
		if f, err := strconv.ParseFloat(v, 64); err == nil && f >= 0 {
			costPerMinute = f
		}
	}
	cost := browserMinutes * costPerMinute

	// Store in Redis for aggregation
	key := fmt.Sprintf("browser:usage:%s:%s", agentID, time.Now().UTC().Format("2006-01-02"))
	c.redis.HIncrByFloat(ctx, key, "total_minutes", browserMinutes)
	c.redis.HIncrByFloat(ctx, key, "total_cost", cost)
	c.redis.HIncrBy(ctx, key, "total_actions", 1)
	c.redis.Expire(ctx, key, 7*24*time.Hour) // Keep for 7 days

	// Also record in the database for permanent tracking
	usage := BrowserUsage{
		AgentID:        agentID,
		SessionID:       sessionID,
		Action:          action,
		Domain:          domain,
		DurationMs:      durationMs,
		BrowserMinutes:  browserMinutes,
		CostUSD:         cost,
	}
	return c.db.WithContext(ctx).Create(&usage).Error
}

// GetBrowserUsageStats returns browser usage statistics for an agent
func (c *Controller) GetBrowserUsageStats(ctx context.Context, agentID string, period string) (*BrowserUsageStats, error) {
	if c.redis == nil {
		return nil, fmt.Errorf("redis not available")
	}

	key := fmt.Sprintf("browser:usage:%s:%s", agentID, period)
	minutes, _ := c.redis.HGet(ctx, key, "total_minutes").Float64()
	cost, _ := c.redis.HGet(ctx, key, "total_cost").Float64()
	actions, _ := c.redis.HGet(ctx, key, "total_actions").Int64()

	return &BrowserUsageStats{
		AgentID:      agentID,
		Period:       period,
		TotalMinutes: minutes,
		TotalCostUSD: cost,
		TotalActions: actions,
	}, nil
}

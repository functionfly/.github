package quota

import (
	"time"
)

// AgentQuotaConfig holds the quota configuration for an agent
type AgentQuotaConfig struct {
	AgentID             string    `json:"agent_id"`
	MaxCallsPerMinute   int       `json:"max_calls_per_min"`
	MaxCallsPerDay      int       `json:"max_calls_per_day"`
	MaxStateWritesPerHr int       `json:"max_state_writes_per_hour"`
	MaxCostPerExecution float64   `json:"max_cost_per_execution"`
	MaxDailySpend       float64   `json:"max_daily_spend"`
	AllowedFunctions    []string  `json:"allowed_functions"`  // fx://org/* patterns; nil = all allowed
	ForbiddenFunctions  []string  `json:"forbidden_functions"`
	CreatedAt           time.Time `json:"created_at"`
	UpdatedAt           time.Time `json:"updated_at"`
}

// QuotaResult is the result of a quota check
type QuotaResult struct {
	Allowed        bool    `json:"allowed"`
	RemainingCalls int     `json:"remaining_calls"`
	RemainingSpend float64 `json:"remaining_spend_usd"`
	RetryAfterSecs int     `json:"retry_after_secs,omitempty"`
	Reason         string  `json:"reason,omitempty"`
}

// AgentUsage holds current usage counters for an agent
type AgentUsage struct {
	AgentID          string    `json:"agent_id"`
	CallsThisMinute  int64     `json:"calls_this_minute"`
	CallsToday       int64     `json:"calls_today"`
	StateWritesThisHr int64    `json:"state_writes_this_hour"`
	SpendTodayUSD    float64   `json:"spend_today_usd"`
	SpendThisMonthUSD float64  `json:"spend_this_month_usd"`
	LastUpdated      time.Time `json:"last_updated"`
}

// QuotaViolationCode represents the type of quota violation
type QuotaViolationCode string

const (
	ViolationRateLimitMinute  QuotaViolationCode = "RATE_LIMIT_MINUTE"
	ViolationRateLimitDay     QuotaViolationCode = "RATE_LIMIT_DAY"
	ViolationSpendCapDaily    QuotaViolationCode = "SPEND_CAP_DAILY"
	ViolationFunctionForbidden QuotaViolationCode = "FUNCTION_FORBIDDEN"
	ViolationFunctionNotAllowed QuotaViolationCode = "FUNCTION_NOT_ALLOWED"
)

// QuotaViolationError is returned when a quota is exceeded
type QuotaViolationError struct {
	Code           QuotaViolationCode `json:"code"`
	Message        string             `json:"message"`
	RetryAfterSecs int                `json:"retry_after_secs,omitempty"`
}

func (e *QuotaViolationError) Error() string {
	return string(e.Code) + ": " + e.Message
}

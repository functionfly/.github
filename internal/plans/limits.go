package plans

import (
	"fmt"
	"os"
	"strconv"
	"strings"
)

const (
	// App limits per tenant
	FreeMaxApps       = 1
	StarterMaxApps    = 3
	ProMaxApps        = 10
	EnterpriseMaxApps = -1 // Unlimited

	// Provider limits per app
	StarterMaxProvidersPerApp    = 2
	ProMaxProvidersPerApp        = 3
	EnterpriseMaxProvidersPerApp = 5

	// Default request limits per month
	// NOTE: These must match the frontend PLANS limits in web/dashboard/src/lib/constants.ts
	// Option B (Balanced): 25K Free, 250K Starter, 2.5M Pro, 25M Enterprise
	FreeMaxRequestsPerMonth                 = 25_000      // Free tier: 25K requests/month
	StarterMaxRequestsPerMonth              = 250_000     // 250K requests - matches frontend Starter
	DefaultProMaxRequestsPerMonth           = 2_500_000   // 2.5M requests - matches frontend Professional
	DefaultEnterpriseMaxRequestsPerMonth    = 25_000_000  // 25M requests - Enterprise tier

	// Function limits
	FreeMaxFunctions           = 3
	StarterMaxFunctions        = 5
	ProMaxFunctions            = 25
	EnterpriseMaxFunctions     = -1 // Unlimited

	// Secrets limits per tenant (unified with vault plans)
	FreeMaxSecrets     = 25
	StarterMaxSecrets  = 500
	ProMaxSecrets       = 5000
	EnterpriseMaxSecrets = 1000000 // Effectively unlimited

	// Token limits per secret
	FreeMaxTokensPerSecret     = 5
	StarterMaxTokensPerSecret  = 25
	ProMaxTokensPerSecret      = 100
	EnterpriseMaxTokensPerSecret = 1000

	// Dynamic credentials limits per tenant (30-day rolling window)
	FreeMaxDynamicCreds     = 100
	StarterMaxDynamicCreds  = 5000
	ProMaxDynamicCreds       = 50000
	EnterpriseMaxDynamicCreds = 1000000

	// Audit exports per day
	FreeMaxAuditExports     = 1
	StarterMaxAuditExports  = 10
	ProMaxAuditExports      = 50
	EnterpriseMaxAuditExports = 1000

	// Custom domains per tenant (starter/pro/enterprise)
	StarterMaxCustomDomains    = 1
	ProMaxCustomDomains        = 5
	EnterpriseMaxCustomDomains = -1 // Unlimited

	// MicroVM-specific limits (Enterprise tier only)
	EnterpriseMaxMicroVMs      = 100
	EnterpriseDefaultMemoryMB  = 512
	EnterpriseMaxMemoryMB      = 2048
	EnterpriseDefaultVCPU      = 2
	EnterpriseMaxVCPU          = 4
	EnterpriseDefaultTimeoutMs = 30_000  // 30 seconds
	EnterpriseMaxTimeoutMs     = 300_000 // 5 minutes

	// MicroVM Enterprise-specific limits (enhanced for heavy MicroVM workloads)
	MicroVMEnterpriseMaxMicroVMs      = 500                    // 5x standard Enterprise
	MicroVMEnterpriseDefaultMemoryMB  = 1024                   // 2x standard
	MicroVMEnterpriseMaxMemoryMB      = 8192                   // 4x standard Enterprise (8GB)
	MicroVMEnterpriseDefaultVCPU      = 4                       // 2x standard
	MicroVMEnterpriseMaxVCPU          = 16                      // 4x standard Enterprise
	MicroVMEnterpriseDefaultTimeoutMs = 60_000                  // 60 seconds
	MicroVMEnterpriseMaxTimeoutMs     = 600_000                 // 10 minutes
	MicroVMEnterpriseMaxConcurrentVMs  = 200                     // Max concurrent MicroVMs
	MicroVMEnterpriseComputeBudgetCents = 5000                  // Included compute budget per month

	// Enterprise tier pricing
	EnterpriseBaseFeeMonthly   = 99.00 // $/month
	EnterpriseRequestsPer10K   = 5000  // $0.50 per 10K requests (cents)
	EnterpriseMicroVMCpuSecond = 2     // $0.02 per vCPU-second (cents)
	EnterpriseMemoryGbSecond   = 2     // $0.002 per GB-second (cents)

	// MicroVM Enterprise tier pricing (enhanced compute budget included)
	MicroVMEnterpriseBaseFeeMonthly    = 299.00  // $/month
	MicroVMEnterpriseRequestsPer10K   = 3000    // $0.30 per 10K requests (cents) - cheaper for high volume
	MicroVMEnterpriseMicroVMCpuSecond  = 1      // $0.01 per vCPU-second (cents) - discounted
	MicroVMEnterpriseMemoryGbSecond    = 1      // $0.001 per GB-second (cents) - discounted

	// Time Machine replay window limits
	FreeReplayWindowHours              = 24
	StarterReplayWindowHours           = 72
	ProReplayWindowDays                = 30
	EnterpriseReplayWindowDays         = 90
	AgentEnterpriseReplayWindowDays    = -1 // Unlimited

	// Time Machine max executions per replay
	FreeMaxExecutionsPerReplay              = 100
	StarterMaxExecutionsPerReplay           = 1_000
	ProMaxExecutionsPerReplay               = 10_000
	EnterpriseMaxExecutionsPerReplay        = 100_000
	AgentEnterpriseMaxExecutionsPerReplay   = -1 // Unlimited

	// Time Machine concurrent replay jobs
	FreeMaxConcurrentReplays              = 1
	StarterMaxConcurrentReplays           = 1
	ProMaxConcurrentReplays               = 3
	EnterpriseMaxConcurrentReplays        = 10
	AgentEnterpriseMaxConcurrentReplays   = -1 // Unlimited

	// Time Machine replay data retention (days)
	FreeReplayDataRetentionDays           = 7
	StarterReplayDataRetentionDays        = 30
	ProReplayDataRetentionDays            = 90
	EnterpriseReplayDataRetentionDays     = 365
	AgentEnterpriseReplayDataRetentionDays = -1 // Unlimited

	// Function Consciousness lookback (days)
	FreeConsciousnessLookbackDays              = 0   // Not available
	StarterConsciousnessLookbackDays           = 0   // Not available
	ProConsciousnessLookbackDays               = 7
	EnterpriseConsciousnessLookbackDays        = 30
	AgentEnterpriseConsciousnessLookbackDays   = -1 // Unlimited

	// Function Consciousness analysis frequency (hours)
	ProConsciousnessFrequencyHours              = 24 // Daily digest
	EnterpriseConsciousnessFrequencyHours       = 1  // Hourly
	AgentEnterpriseConsciousnessFrequencyHours  = 0  // Real-time (5-min scheduler)

	// Function Consciousness max auto-fixes per day
	EnterpriseMaxAutoFixesPerDay        = 5
	AgentEnterpriseMaxAutoFixesPerDay   = -1 // Unlimited

	// Function Consciousness: max active insights per tenant
	ProConsciousnessMaxActiveInsights        = 50
	EnterpriseConsciousnessMaxActiveInsights = 500
	AgentEnterpriseConsciousnessMaxActiveInsights = -1 // Unlimited

	// Function Consciousness: max history retention (days)
	ProConsciousnessHistoryDays        = 30
	EnterpriseConsciousnessHistoryDays = 90
	AgentEnterpriseConsciousnessHistoryDays = 365
)

// GetConsciousnessLookbackDays returns the analysis lookback window (days) for a plan.
// Returns 0 for plans without consciousness access, -1 for unlimited.
func GetConsciousnessLookbackDays(plan string) int {
	switch plan {
	case PlanAgentEnterprise:
		return AgentEnterpriseConsciousnessLookbackDays
	case PlanEnterprise, PlanAgentPro, PlanMicroVMEnterprise:
		return EnterpriseConsciousnessLookbackDays
	case PlanPro, PlanAgentScale:
		return ProConsciousnessLookbackDays
	default:
		return 0
	}
}

// GetConsciousnessFrequencyHours returns how often scheduled analysis runs (hours).
// 0 means real-time / sub-hourly.
func GetConsciousnessFrequencyHours(plan string) int {
	switch plan {
	case PlanAgentEnterprise:
		return AgentEnterpriseConsciousnessFrequencyHours
	case PlanEnterprise, PlanAgentPro:
		return EnterpriseConsciousnessFrequencyHours
	case PlanPro, PlanAgentScale:
		return ProConsciousnessFrequencyHours
	default:
		return 0
	}
}

// GetConsciousnessMaxAutoFixesPerDay returns the maximum auto-applied actions per day for a plan.
// Returns 0 for plans without autonomous consciousness, -1 for unlimited.
func GetConsciousnessMaxAutoFixesPerDay(plan string) int {
	switch plan {
	case PlanAgentEnterprise:
		return AgentEnterpriseMaxAutoFixesPerDay
	case PlanEnterprise, PlanAgentPro:
		return EnterpriseMaxAutoFixesPerDay
	default:
		return 0
	}
}

// GetConsciousnessMaxActiveInsights returns the cap on simultaneously active insights per tenant.
// Returns -1 for unlimited.
func GetConsciousnessMaxActiveInsights(plan string) int {
	switch plan {
	case PlanAgentEnterprise:
		return AgentEnterpriseConsciousnessMaxActiveInsights
	case PlanEnterprise, PlanAgentPro:
		return EnterpriseConsciousnessMaxActiveInsights
	case PlanPro, PlanAgentScale:
		return ProConsciousnessMaxActiveInsights
	default:
		return 0
	}
}

// GetConsciousnessHistoryDays returns how long insight + score history is retained.
func GetConsciousnessHistoryDays(plan string) int {
	switch plan {
	case PlanAgentEnterprise:
		return AgentEnterpriseConsciousnessHistoryDays
	case PlanEnterprise, PlanAgentPro:
		return EnterpriseConsciousnessHistoryDays
	case PlanPro, PlanAgentScale:
		return ProConsciousnessHistoryDays
	default:
		return 0
	}
}

// HasConsciousnessFeature returns true if the plan has any consciousness capability (basic+).
func HasConsciousnessFeature(plan string) bool {
	return GetConsciousnessLookbackDays(plan) > 0 || GetConsciousnessLookbackDays(plan) == -1
}

// HasAdvancedConsciousnessFeature returns true if the plan has advanced consciousness (hourly+).
func HasAdvancedConsciousnessFeature(plan string) bool {
	return plan == PlanEnterprise || plan == PlanAgentPro || plan == PlanAgentEnterprise || plan == PlanMicroVMEnterprise
}

// HasAutonomousConsciousnessFeature returns true if the plan has autonomous (real-time, unlimited lookback).
func HasAutonomousConsciousnessFeature(plan string) bool {
	return plan == PlanAgentEnterprise
}

// ConsciousnessLimits provides a snapshot of all consciousness limits for a plan.
type ConsciousnessLimits struct {
	LookbackDays       int  `json:"lookback_days"`
	FrequencyHours     int  `json:"frequency_hours"`
	MaxActiveInsights  int  `json:"max_active_insights"`
	MaxAutoFixesPerDay int  `json:"max_auto_fixes_per_day"`
	HistoryDays        int  `json:"history_days"`
	Unlimited          bool `json:"unlimited"`
}

// GetConsciousnessLimits returns all consciousness limits for a plan.
func GetConsciousnessLimits(plan string) ConsciousnessLimits {
	lookback := GetConsciousnessLookbackDays(plan)
	return ConsciousnessLimits{
		LookbackDays:       lookback,
		FrequencyHours:     GetConsciousnessFrequencyHours(plan),
		MaxActiveInsights:  GetConsciousnessMaxActiveInsights(plan),
		MaxAutoFixesPerDay: GetConsciousnessMaxAutoFixesPerDay(plan),
		HistoryDays:        GetConsciousnessHistoryDays(plan),
		Unlimited:          lookback == -1,
	}
}

// StateFabric limits
const (
	FreeMaxStateFabrics       = 1  // Sandbox tier: 1 state object for experimentation
	StarterMaxStateFabrics    = 3
	ProMaxStateFabrics         = 10
	EnterpriseMaxStateFabrics  = -1 // Unlimited
)

// Enterprise SLA Plan specific limits
const (
	// SLA Target percentages by plan
	FreeSLATargetPercent       = 0.0   // No SLA guarantee
	StarterSLATargetPercent    = 99.5  // 99.5% uptime
	ProSLATargetPercent        = 99.9  // 99.9% uptime
	EnterpriseSLATargetPercent  = 99.99 // 99.99% uptime (four nines)
	EnterpriseSLASpecialTargetPercent = 99.999 // 99.999% uptime (five nines) for enterprise_sla

	// SLA response time targets (in minutes)
	EnterpriseSLAResponseTimeMinutes     = 15  // Critical: 15 min response
	EnterpriseSLAAcknowledgmentMinutes   = 5   // Acknowledge within 5 min
	EnterpriseSLAResolutionTimeMinutes   = 60  // Resolution target: 1 hour
	EnterpriseSLACriticalResolutionMin  = 30  // Critical issues: 30 min

	// SLA credits (percentage of monthly fee credited for downtime)
	EnterpriseSLACreditPercentBelow99   = 5  // 5% credit if below 99%
	EnterpriseSLACreditPercentBelow999  = 10 // 10% credit if below 99.9%
	EnterpriseSLACreditPercentBelow9999 = 25 // 25% credit if below 99.99%
	EnterpriseSLACreditPercentBelow99999 = 50 // 50% credit if below 99.999%

	// SLA incident thresholds
	EnterpriseSLAMaxIncidentsPerMonth       = 4   // Max incidents before escalation
	EnterpriseSLAMaxDowntimeMinutesPerMonth = 4.3 // ~4.3 min/month for 99.99%
	EnterpriseSLASpecialMaxDowntimeMinutes = 0.43 // ~0.43 min/month for 99.999%

	// SLA dashboard and reporting
	EnterpriseSLADashboardRefreshSeconds   = 60  // Real-time dashboard refresh
	EnterpriseSLAReportGenerationHours     = 24  // Daily SLA reports
)

// Plan type constants
// NOTE: "professional" must match frontend plan-utils.ts PlanTier type
const (
	PlanFree       = "free" // Default plan for new signups
	PlanStarter    = "starter"
	PlanPro        = "professional" // Was "pro" - changed for consistency with frontend
	PlanEnterprise = "enterprise"
	PlanEnterpriseSLA = "enterprise_sla" // Enterprise with enhanced SLA guarantees
	PlanMicroVMEnterprise = "microvm_enterprise" // MicroVM-focused Enterprise plan
)

// AEP (Agent Execution Plan) tier constants
const (
	PlanAgentStarter    = "agent_starter"
	PlanAgentScale      = "agent_scale"
	PlanAgentPro        = "agent_pro"
	PlanAgentEnterprise = "agent_enterprise"
)

// AEP Concurrency limits (reserved slots per agent)
const (
	AgentStarterMaxConcurrency    = 10
	AgentScaleMaxConcurrency      = 100
	AgentProMaxConcurrency        = 500
	AgentEnterpriseMaxConcurrency = -1 // Unlimited / dedicated pool
)

// AEP Burst ceiling (calls per second)
const (
	AgentStarterBurstCeiling    = 50
	AgentScaleBurstCeiling      = 500
	AgentProBurstCeiling        = 2000
	AgentEnterpriseBurstCeiling = -1 // Unlimited
)

// AEP Tool call limits per month
// These are now set based on market research to ensure profitability:
// - Starter: 100K calls at $29 = $0.00029/call base (covers ~$5-10 AI inference cost)
// - Scale: 1M calls at $149 = $0.000149/call base (covers ~$30-50 AI inference cost)
// - Pro: 10M calls at $399 = $0.00004/call base (covers ~$100-200 AI inference cost)
// Key principle: Don't profit on AI inference - profit on orchestration & reliability
const (
	AgentStarterMaxCallsPerMonth    = 100_000
	AgentScaleMaxCallsPerMonth      = 1_000_000
	AgentProMaxCallsPerMonth        = 10_000_000
	AgentEnterpriseMaxCallsPerMonth = -1 // Custom
)

// AEP Calls per minute limits
const (
	AgentStarterMaxCallsPerMinute    = 100
	AgentScaleMaxCallsPerMinute      = 500
	AgentProMaxCallsPerMinute        = 2000
	AgentEnterpriseMaxCallsPerMinute = -1 // Custom
)

// AEP Calls per day limits (derived from monthly limits: ~3.3% of monthly)
const (
	AgentStarterMaxCallsPerDay    = 3_333
	AgentScaleMaxCallsPerDay      = 33_333
	AgentProMaxCallsPerDay        = 333_333
	AgentEnterpriseMaxCallsPerDay = -1 // Custom
)

// AEP State writes per hour limits
const (
	AgentStarterMaxStateWritesPerHr    = 1_000
	AgentScaleMaxStateWritesPerHr      = 10_000
	AgentProMaxStateWritesPerHr        = 50_000
	AgentEnterpriseMaxStateWritesPerHr = -1 // Custom
)

// AEP Daily spend caps (USD)
const (
	AgentStarterDailySpendCapUSD    = 5.0
	AgentScaleDailySpendCapUSD      = 30.0
	AgentProDailySpendCapUSD        = 100.0
	AgentEnterpriseDailySpendCapUSD = -1.0 // Custom
)

// AEP Memory storage limits (GB)
const (
	AgentStarterMaxMemoryGB    = 10
	AgentScaleMaxMemoryGB      = 100
	AgentProMaxMemoryGB        = 500
	AgentEnterpriseMaxMemoryGB = -1 // Custom
)

// AEP Log retention days
const (
	AgentStarterLogRetentionDays    = 30
	AgentScaleLogRetentionDays      = 90
	AgentProLogRetentionDays        = 365
	AgentEnterpriseLogRetentionDays = -1 // Custom
)

// ==================== 2026 Unified Plan Pricing ====================
// Agent capabilities are now BUNDLED into main plans (no separate agent plans)
// Legacy agent tiers still exist as aliases for backward compatibility

// Main Plan Monthly Pricing (cents) - 2026 optimized (Option C Hybrid)
const (
	StarterPriceCents         = 2400  // $24/month
	ProPriceCents             = 7900  // $79/month
	EnterprisePriceCents      = 29900 // $299/month base (includes 5M AI calls)
	MicroVMEnterprisePriceCents = 39900 // $399/month - enhanced MicroVM capabilities
	AgentEnterprisePriceCents  = 49900 // $499/month - unlimited AI
)

// Annual Pricing (2 months free = 10 months billed)
const (
	StarterAnnualCents         = 24000  // $240/year ($24/mo equiv)
	ProAnnualCents             = 79000  // $790/year ($79/mo equiv)
	EnterpriseAnnualCents      = 299000 // $2990/year ($299/mo equiv)
	MicroVMEnterpriseAnnualCents = 399000 // $3990/year ($399/mo equiv)
	AgentEnterpriseAnnualCents  = 499000 // $4990/year ($499/mo equiv)
)

// Usage-Based Overage (cents per 1000 calls) - 2026 optimized
const (
	StarterOveragePer1000Cents     = 15  // $0.15/1K calls
	ProOveragePer1000Cents          = 8   // $0.08/1K calls
	EnterpriseOveragePer1000Cents   = 5   // $0.05/1K calls (lower for higher tier)
	// Agent Enterprise: no overage (unlimited)
)

// ==================== Agent Capabilities (Bundled in Main Plans) ====================
// Free: Basic 10K calls/mo, 3 concurrency, hard stop
// Starter: 100K calls/mo, 10 concurrency, $0.15/1K overage
// Professional: 1M calls/mo, 100 concurrency, $0.08/1K overage
// Enterprise: 5M calls/mo, 500 concurrency, $0.05/1K overage
// Agent Enterprise: Unlimited, $499/mo base

// Free tier agent limits
const (
	FreeMaxAgents             = 3
	FreeMaxAICallsPerMonth    = 10_000
	FreeMaxConcurrency        = 3
	FreeMaxCallsPerMinute     = 10
	FreeOverageAllowed        = false // Hard stop
)

// Starter agent limits (included in $24/mo plan)
const (
	StarterMaxAgents             = 10
	StarterMaxAICallsPerMonth    = 100_000
	StarterMaxConcurrency        = 10
	StarterMaxCallsPerMinute     = 100
	StarterMaxStateWritesPerHr   = 1_000
	StarterMaxMemoryGB           = 10
	StarterLogRetentionDays      = 30
)

// Professional agent limits (included in $79/mo plan)
const (
	ProMaxAgents             = 100
	ProMaxAICallsPerMonth    = 1_000_000
	ProMaxConcurrency        = 100
	ProMaxCallsPerMinute     = 500
	ProMaxStateWritesPerHr   = 10_000
	ProMaxMemoryGB           = 100
	ProLogRetentionDays      = 90
)

// Enterprise agent limits (included in $299/mo plan)
const (
	EnterpriseMaxAgents             = 500
	EnterpriseMaxAICallsPerMonth   = 5_000_000  // 5M calls included
	EnterpriseMaxConcurrency       = 500
	EnterpriseMaxCallsPerMinute    = 2000
	EnterpriseMaxStateWritesPerHr  = 50_000
	EnterpriseMaxMemoryGB          = 500
	EnterpriseLogRetentionDays     = 365
)

// Agent Enterprise limits ($499/mo - unlimited)
// Note: Many are -1 (unlimited), defined here to avoid reference issues
const (
	AgentEnterpriseMaxAgents   = -1 // Unlimited
	AgentEnterpriseMaxAICallsPerMonth = -1 // Unlimited
	// AgentEnterpriseMaxConcurrency, MaxCallsPerMinute etc. are defined in AEP section above
)

// User seat limits per plan
const (
	StarterMaxUsers           = 3
	ProMaxUsers              = 10
	EnterpriseMaxUsers       = -1 // Unlimited
	SeatWarningThresholdValue = 0.80 // Warn at 80% of limit
	SeatGracePeriodDays       = 30  // Days to remove users after downgrade
)

// Runtime type constants
const (
	RuntimeWasm          = "wasm"
	RuntimePython        = "python"
	RuntimeFlyMachines   = "fly-machines"    // CPython in Firecracker via Fly Machines API
	RuntimePythonMicroVM = "python-microvm" // Legacy name - kept for backward compat with existing DB entries
	RuntimePythonLight   = "python-light"   // Lightweight Python runtime: shared functionfly-python daemon
	RuntimePrism         = "prism"
	RuntimeBun           = "bun"
	RuntimeDeno          = "deno"
)

// IsEnterpriseTier returns true if the plan is enterprise or agent_enterprise
// Both plans have access to enterprise features (SLA, Audit, etc.)
func IsEnterpriseTier(plan string) bool {
	return plan == PlanEnterprise || plan == PlanAgentEnterprise || plan == PlanEnterpriseSLA || plan == PlanMicroVMEnterprise
}

// IsMicroVMEnterprisePlan returns true if the plan is MicroVM Enterprise
func IsMicroVMEnterprisePlan(plan string) bool {
	return plan == PlanMicroVMEnterprise
}

// IsSLAPlan returns true if the plan has SLA features (Enterprise or Enterprise SLA)
func IsSLAPlan(plan string) bool {
	return plan == PlanEnterprise || plan == PlanEnterpriseSLA || plan == PlanAgentEnterprise || plan == PlanMicroVMEnterprise
}

// GetSLATargetPercent returns the SLA target percentage for the given plan
func GetSLATargetPercent(plan string) float64 {
	switch plan {
	case PlanEnterpriseSLA:
		return EnterpriseSLASpecialTargetPercent
	case PlanEnterprise, PlanAgentEnterprise:
		return EnterpriseSLATargetPercent
	case PlanPro:
		return ProSLATargetPercent
	case PlanStarter:
		return StarterSLATargetPercent
	default:
		return FreeSLATargetPercent
	}
}

// GetSLAMaxDowntimeMinutes returns the maximum allowed downtime minutes per month for SLA compliance
func GetSLAMaxDowntimeMinutes(plan string) float64 {
	switch plan {
	case PlanEnterpriseSLA:
		return EnterpriseSLASpecialMaxDowntimeMinutes
	case PlanEnterprise, PlanAgentEnterprise:
		return EnterpriseSLAMaxDowntimeMinutesPerMonth
	case PlanPro:
		return 43.2 // ~43.2 min/month for 99.9%
	case PlanStarter:
		return 219 // ~219 min/month for 99.5%
	default:
		return 0 // No SLA guarantee
	}
}

// GetSLACreditPercent returns the credit percentage for SLA violations
func GetSLACreditPercent(plan string, uptimePercent float64) int {
	switch plan {
	case PlanEnterpriseSLA:
		if uptimePercent < 99.999 {
			return EnterpriseSLACreditPercentBelow99999
		}
	case PlanEnterprise, PlanAgentEnterprise:
		if uptimePercent < 99.99 {
			return EnterpriseSLACreditPercentBelow9999
		} else if uptimePercent < 99.9 {
			return EnterpriseSLACreditPercentBelow999
		} else if uptimePercent < 99.0 {
			return EnterpriseSLACreditPercentBelow99
		}
	case PlanPro:
		if uptimePercent < 99.9 {
			return EnterpriseSLACreditPercentBelow999
		} else if uptimePercent < 99.0 {
			return EnterpriseSLACreditPercentBelow99
		}
	}
	return 0 // No credit or not applicable
}

// IsAgentTier returns true if the plan is an AEP agent tier
func IsAgentTier(plan string) bool {
	switch plan {
	case PlanAgentStarter, PlanAgentScale, PlanAgentPro, PlanAgentEnterprise:
		return true
	}
	return false
}

// MaxUsersPerPlan returns the maximum number of users allowed for a plan
// Returns -1 for unlimited (enterprise)
func MaxUsersPerPlan(plan string) int {
	switch plan {
	case PlanPro:
		return ProMaxUsers
	case PlanEnterprise:
		return EnterpriseMaxUsers
	case PlanStarter:
		fallthrough
	default:
		return StarterMaxUsers
	}
}

// SeatWarningThreshold returns the percentage (0.0-1.0) at which to warn about seat usage
func SeatWarningThreshold() float64 {
	return SeatWarningThresholdValue
}

// GetSeatGracePeriodDays returns the number of days a tenant has to remove users
// after a downgrade causes them to exceed the new limit
func GetSeatGracePeriodDays() int {
	return SeatGracePeriodDays
}

// IsSeatLimitReached checks if the tenant has reached their seat limit
// currentUsers is the number of active (non-deactivated) users
func IsSeatLimitReached(plan string, currentUsers int) bool {
	maxUsers := MaxUsersPerPlan(plan)
	if maxUsers == -1 {
		return false // Unlimited
	}
	return currentUsers >= maxUsers
}

// IsSeatWarningThreshold checks if the tenant has reached the warning threshold
// currentUsers is the number of active (non-deactivated) users
func IsSeatWarningThreshold(plan string, currentUsers int) bool {
	maxUsers := MaxUsersPerPlan(plan)
	if maxUsers == -1 {
		return false // Unlimited
	}
	warningAt := float64(maxUsers) * SeatWarningThresholdValue
	return float64(currentUsers) >= warningAt
}

// SeatUsageInfo provides information about seat usage for a tenant
type SeatUsageInfo struct {
	Plan           string
	CurrentUsers   int
	MaxUsers       int
	WarningPercent float64
	IsUnlimited    bool
	IsAtLimit      bool
	IsAtWarning    bool
}

// GetSeatUsage returns detailed seat usage information
func GetSeatUsage(plan string, currentUsers int) *SeatUsageInfo {
	maxUsers := MaxUsersPerPlan(plan)
	isUnlimited := maxUsers == -1

	info := &SeatUsageInfo{
		Plan:           plan,
		CurrentUsers:   currentUsers,
		MaxUsers:       maxUsers,
		WarningPercent: SeatWarningThresholdValue * 100,
		IsUnlimited:    isUnlimited,
	}

	if !isUnlimited {
		info.IsAtLimit = currentUsers >= maxUsers
		info.IsAtWarning = float64(currentUsers) >= float64(maxUsers)*SeatWarningThresholdValue
	}

	return info
}

// AgentMaxCallsPerMinute returns the calls-per-minute limit for a plan
// Supports both legacy agent tiers and main plan names
func AgentMaxCallsPerMinute(plan string) int {
	switch plan {
	case PlanAgentScale, PlanPro:
		return ProMaxCallsPerMinute
	case PlanAgentPro, PlanEnterprise:
		return EnterpriseMaxCallsPerMinute
	case PlanAgentEnterprise:
		return -1 // Unlimited
	case PlanStarter, PlanAgentStarter:
		return StarterMaxCallsPerMinute
	default:
		return StarterMaxCallsPerMinute
	}
}

// AgentMaxCallsPerDay returns the calls-per-day limit for a plan
func AgentMaxCallsPerDay(plan string) int {
	switch plan {
	case PlanAgentScale, PlanPro:
		return 33_333
	case PlanAgentPro, PlanEnterprise:
		return 333_333
	case PlanAgentEnterprise:
		return -1 // Unlimited
	case PlanStarter, PlanAgentStarter:
		return 3_333
	default:
		return 3_333
	}
}

// AgentMaxConcurrency returns the max concurrency for a plan
func AgentMaxConcurrency(plan string) int {
	switch plan {
	case PlanAgentScale, PlanPro:
		return ProMaxConcurrency
	case PlanAgentPro, PlanEnterprise:
		return EnterpriseMaxConcurrency
	case PlanAgentEnterprise:
		return -1 // Unlimited
	case PlanStarter, PlanAgentStarter:
		return StarterMaxConcurrency
	default:
		return StarterMaxConcurrency
	}
}

// AgentLogRetentionDays returns the log retention days for a plan
func AgentLogRetentionDays(plan string) int {
	switch plan {
	case PlanAgentScale, PlanPro:
		return ProLogRetentionDays
	case PlanAgentPro, PlanEnterprise:
		return EnterpriseLogRetentionDays
	case PlanAgentEnterprise:
		return -1 // Custom
	case PlanStarter, PlanAgentStarter:
		return StarterLogRetentionDays
	default:
		return StarterLogRetentionDays
	}
}

// GetAgentLimitsForPlan returns the agent limits for any plan (main or agent tier)
func GetAgentLimitsForPlan(plan string) (maxAgents, maxCallsPerMonth, maxConcurrency, maxCallsPerMinute int) {
	switch plan {
	case PlanEnterprise, PlanAgentPro:
		return EnterpriseMaxAgents, EnterpriseMaxAICallsPerMonth, EnterpriseMaxConcurrency, EnterpriseMaxCallsPerMinute
	case PlanPro, PlanAgentScale:
		return ProMaxAgents, ProMaxAICallsPerMonth, ProMaxConcurrency, ProMaxCallsPerMinute
	case PlanStarter, PlanAgentStarter:
		return StarterMaxAgents, StarterMaxAICallsPerMonth, StarterMaxConcurrency, StarterMaxCallsPerMinute
	case PlanAgentEnterprise:
		return AgentEnterpriseMaxAgents, AgentEnterpriseMaxAICallsPerMonth, AgentEnterpriseMaxConcurrency, AgentEnterpriseMaxCallsPerMinute
	default: // Free
		return FreeMaxAgents, FreeMaxAICallsPerMonth, FreeMaxConcurrency, FreeMaxCallsPerMinute
	}
}

// GetAgentTierLimits returns the quota limits for an agent based on plan tier
// Returns: (maxCallsPerMinute, maxCallsPerDay, maxStateWritesPerHr, maxDailySpendUSD)
func GetAgentTierLimits(planTier string) (int, int, int, float64) {
	switch planTier {
	case PlanAgentEnterprise:
		return -1, -1, -1, -1.0 // Unlimited
	case PlanAgentPro, PlanEnterprise:
		return EnterpriseMaxCallsPerMinute, 333_333, EnterpriseMaxStateWritesPerHr, 100.0
	case PlanAgentScale, PlanPro:
		return ProMaxCallsPerMinute, 33_333, ProMaxStateWritesPerHr, 30.0
	case PlanAgentStarter, PlanStarter:
		return StarterMaxCallsPerMinute, 3_333, StarterMaxStateWritesPerHr, 5.0
	default: // Free or unknown
		return FreeMaxCallsPerMinute, 333, 100, 1.0
	}
}

// GetOverageRate returns the overage rate (cents per 1000 calls) for a plan
func GetOverageRate(plan string) int {
	switch plan {
	case PlanEnterprise, PlanAgentPro:
		return EnterpriseOveragePer1000Cents
	case PlanPro, PlanAgentScale:
		return ProOveragePer1000Cents
	case PlanStarter, PlanAgentStarter:
		return StarterOveragePer1000Cents
	default:
		return 0 // Free has hard stop
	}
}

// GetPlanPriceCents returns the monthly price in cents for a plan
func GetPlanPriceCents(plan string) int {
	switch plan {
	case PlanAgentEnterprise:
		return AgentEnterprisePriceCents
	case PlanEnterprise, PlanAgentPro:
		return EnterprisePriceCents
	case PlanPro, PlanAgentScale:
		return ProPriceCents
	case PlanStarter, PlanAgentStarter:
		return StarterPriceCents
	default:
		return 0 // Free
	}
}

// GetAnnualPriceCents returns the annual price in cents for a plan
func GetAnnualPriceCents(plan string) int {
	switch plan {
	case PlanAgentEnterprise:
		return AgentEnterpriseAnnualCents
	case PlanEnterprise, PlanAgentPro:
		return EnterpriseAnnualCents
	case PlanPro, PlanAgentScale:
		return ProAnnualCents
	case PlanStarter, PlanAgentStarter:
		return StarterAnnualCents
	default:
		return 0 // Free
	}
}

// IsValidRuntimeForPlan checks if the runtime is allowed for the given plan
func IsValidRuntimeForPlan(plan string, runtime string) bool {
	switch runtime {
	case RuntimeFlyMachines, RuntimePythonMicroVM:
		// fly-machines and legacy python-microvm are only available for enterprise tier
		return plan == PlanEnterprise || plan == PlanMicroVMEnterprise
	default:
		return true // All other runtimes are available for all plans
	}
}

// IsFlyMachinesRuntime returns true if the runtime is fly-machines or legacy python-microvm
func IsFlyMachinesRuntime(runtime string) bool {
	return runtime == RuntimeFlyMachines || runtime == RuntimePythonMicroVM
}

// MaxProviders returns the maximum number of providers allowed for the given plan
func MaxProviders(plan string) int {
	switch plan {
	case PlanPro:
		return ProMaxProvidersPerApp
	case PlanEnterprise:
		return EnterpriseMaxProvidersPerApp
	case PlanStarter:
		fallthrough
	default:
		return StarterMaxProvidersPerApp
	}
}

// MaxApps returns the maximum number of apps allowed for the given plan
func MaxApps(plan string) int {
	switch plan {
	case PlanFree:
		return FreeMaxApps
	case PlanStarter:
		return StarterMaxApps
	case PlanPro:
		return ProMaxApps
	case PlanEnterprise:
		return EnterpriseMaxApps
	default:
		return StarterMaxApps
	}
}

// GetMaxCustomDomains returns the maximum number of custom domains for the given plan.
// Returns -1 for unlimited (enterprise).
func GetMaxCustomDomains(plan string) int {
	switch strings.ToLower(plan) {
	case PlanPro:
		return ProMaxCustomDomains
	case PlanEnterprise:
		return EnterpriseMaxCustomDomains
	case PlanStarter:
		fallthrough
	default:
		return StarterMaxCustomDomains
	}
}


// GetMaxSecrets returns the maximum number of secrets allowed for the given plan
func GetMaxSecrets(plan string) int {
	switch plan {
	case PlanFree:
		return FreeMaxSecrets
	case PlanStarter:
		return StarterMaxSecrets
	case PlanPro:
		return ProMaxSecrets
	case PlanEnterprise:
		return EnterpriseMaxSecrets
	default:
		return FreeMaxSecrets
	}
}

// GetMaxTokensPerSecret returns the maximum number of access tokens allowed per secret for the given plan
func GetMaxTokensPerSecret(plan string) int {
	switch plan {
	case PlanFree:
		return FreeMaxTokensPerSecret
	case PlanStarter:
		return StarterMaxTokensPerSecret
	case PlanPro:
		return ProMaxTokensPerSecret
	case PlanEnterprise:
		return EnterpriseMaxTokensPerSecret
	default:
		return FreeMaxTokensPerSecret
	}
}


// MaxRequestsPerMonth returns the maximum number of requests allowed per month for the given plan
func MaxRequestsPerMonth(plan string) int {
	switch plan {
	case PlanPro:
		// Allow overriding via environment variable for flexibility
		if limitStr := os.Getenv("PRO_REQUEST_LIMIT"); limitStr != "" {
			if limit, err := strconv.Atoi(limitStr); err == nil && limit > 0 {
				return limit
			}
		}
		return DefaultProMaxRequestsPerMonth
	case PlanEnterprise:
		if limitStr := os.Getenv("ENTERPRISE_REQUEST_LIMIT"); limitStr != "" {
			if limit, err := strconv.Atoi(limitStr); err == nil && limit > 0 {
				return limit
			}
		}
		return DefaultEnterpriseMaxRequestsPerMonth
	case PlanStarter:
		fallthrough
	default:
		return StarterMaxRequestsPerMonth
	}
}

// MicroVMLimits represents the resource limits for a MicroVM execution
type MicroVMLimits struct {
	MaxMicroVMs      int
	MaxConcurrentVMs int  // Maximum concurrent VMs (0 = use MaxMicroVMs as limit)
	DefaultMemoryMB  int
	MaxMemoryMB      int
	DefaultVCPU      int
	MaxVCPU          int
	DefaultTimeout   int
	MaxTimeout       int
	IncludedBudget   int  // Included compute budget in cents (for MicroVM Enterprise)
}

// GetMicroVMLimits returns the MicroVM resource limits for the given plan
// Returns nil for non-enterprise plans (MicroVMs not available)
func GetMicroVMLimits(plan string) *MicroVMLimits {
	limits, _ := GetMicroVMLimitsForPath(plan, "")
	return limits
}

// GetMicroVMLimitsForPath returns the MicroVM resource limits for the given plan
// and deployment path. Hybrid and BYOAWS deployments have tighter caps because
// compute costs are customer-borne; Marketplace is the loosest since AWS
// handles capacity.
//
// deploymentPath is one of: "marketplace", "byoaws", "hybrid", or "" for default.
//
// Returns nil for non-enterprise plans (MicroVMs not available).
func GetMicroVMLimitsForPath(plan, deploymentPath string) (*MicroVMLimits, error) {
	if plan != PlanEnterprise && plan != PlanMicroVMEnterprise {
		return nil, nil
	}

	var base *MicroVMLimits
	if plan == PlanMicroVMEnterprise {
		base = &MicroVMLimits{
			MaxMicroVMs:      MicroVMEnterpriseMaxMicroVMs,
			MaxConcurrentVMs: MicroVMEnterpriseMaxConcurrentVMs,
			DefaultMemoryMB:  MicroVMEnterpriseDefaultMemoryMB,
			MaxMemoryMB:      MicroVMEnterpriseMaxMemoryMB,
			DefaultVCPU:      MicroVMEnterpriseDefaultVCPU,
			MaxVCPU:          MicroVMEnterpriseMaxVCPU,
			DefaultTimeout:   MicroVMEnterpriseDefaultTimeoutMs,
			MaxTimeout:       MicroVMEnterpriseMaxTimeoutMs,
			IncludedBudget:   MicroVMEnterpriseComputeBudgetCents,
		}
	} else {
		base = &MicroVMLimits{
			MaxMicroVMs:      EnterpriseMaxMicroVMs,
			MaxConcurrentVMs: EnterpriseMaxMicroVMs, // Concurrent = max for standard Enterprise
			DefaultMemoryMB:  EnterpriseDefaultMemoryMB,
			MaxMemoryMB:      EnterpriseMaxMemoryMB,
			DefaultVCPU:      EnterpriseDefaultVCPU,
			MaxVCPU:          EnterpriseMaxVCPU,
			DefaultTimeout:   EnterpriseDefaultTimeoutMs,
			MaxTimeout:       EnterpriseMaxTimeoutMs,
			IncludedBudget:   0, // No included budget for standard Enterprise
		}
	}

	// Apply per-path adjustments
	switch deploymentPath {
	case "byoaws":
		// Customer pays AWS directly; we still enforce a sane ceiling.
		if base.MaxMicroVMs > 200 {
			base.MaxMicroVMs = 200
		}
		if base.MaxConcurrentVMs > 100 {
			base.MaxConcurrentVMs = 100
		}
	case "hybrid":
		// FunctionFly-managed: tighter to keep ops surface manageable.
		if base.MaxMicroVMs > 100 {
			base.MaxMicroVMs = 100
		}
		if base.MaxConcurrentVMs > 50 {
			base.MaxConcurrentVMs = 50
		}
	case "marketplace", "":
		// No additional restriction; AWS Marketplace handles capacity.
	default:
		return nil, fmt.Errorf("unknown deployment path: %s", deploymentPath)
	}

	return base, nil
}

// ValidationError represents a validation error with code and message
type ValidationError struct {
	Code    string
	Message string
}

func (e *ValidationError) Error() string {
	return e.Message
}

// ValidateMicroVMResources validates the requested MicroVM resources
func ValidateMicroVMResources(plan string, memoryMB int, vCPU int, timeoutMs int) error {
	limits := GetMicroVMLimits(plan)
	if limits == nil {
		return &ValidationError{
			Code:    "MICROVM_NOT_AVAILABLE",
			Message: "MicroVMs are only available for Enterprise tier",
		}
	}

	if memoryMB < 256 || memoryMB > limits.MaxMemoryMB {
		return &ValidationError{
			Code:    "INVALID_MEMORY",
			Message: fmt.Sprintf("Memory must be between 256MB and %dMB", limits.MaxMemoryMB),
		}
	}

	if vCPU < 1 || vCPU > limits.MaxVCPU {
		return &ValidationError{
			Code:    "INVALID_VCPU",
			Message: fmt.Sprintf("vCPU must be between 1 and %d", limits.MaxVCPU),
		}
	}

	if timeoutMs < 1000 || timeoutMs > limits.MaxTimeout {
		return &ValidationError{
			Code:    "INVALID_TIMEOUT",
			Message: fmt.Sprintf("Timeout must be between 1s and %ds", limits.MaxTimeout/1000),
		}
	}

	return nil
}

// MicroVMBilling represents the billing breakdown for a MicroVM execution
type MicroVMBilling struct {
	BaseFeeMonthly   int
	RequestCharges   int
	ComputeCharges   int
	MemoryCharges    int
	TotalCents       int
	IncludedBudget   int  // Included compute budget (for MicroVM Enterprise)
	UsedBudget       int  // Used portion of included budget
}

// CalculateMicroVMBilling calculates the billing for MicroVM usage
func CalculateMicroVMBilling(plan string, requests int, computeSeconds float64, memoryMB int, memorySeconds float64) *MicroVMBilling {
	// Only PlanEnterprise and PlanMicroVMEnterprise have MicroVM billing
	if plan != PlanEnterprise && plan != PlanMicroVMEnterprise {
		return nil
	}

	// MicroVM Enterprise uses different pricing
	if plan == PlanMicroVMEnterprise {
		return calculateMicroVMEnterpriseBilling(requests, computeSeconds, memoryMB, memorySeconds)
	}

	// Standard Enterprise billing
	baseFeeCents := int(EnterpriseBaseFeeMonthly * 100)

	// Request charges: $0.50 per 10K requests
	requestCharges := (requests / 10000) * EnterpriseRequestsPer10K
	if requests%10000 > 0 {
		requestCharges += EnterpriseRequestsPer10K
	}

	// Compute charges: $0.02 per vCPU-second (using default 2 vCPU)
	computeCharges := int(computeSeconds * float64(EnterpriseMicroVMCpuSecond*EnterpriseDefaultVCPU))

	// Memory charges: $0.002 per GB-second
	memoryGBSeconds := memorySeconds * float64(memoryMB) / 1024.0
	memoryCharges := int(memoryGBSeconds * float64(EnterpriseMemoryGbSecond))

	total := baseFeeCents + requestCharges + computeCharges + memoryCharges

	return &MicroVMBilling{
		BaseFeeMonthly: baseFeeCents,
		RequestCharges: requestCharges,
		ComputeCharges: computeCharges,
		MemoryCharges:  memoryCharges,
		TotalCents:     total,
		IncludedBudget: 0,
		UsedBudget:     0,
	}
}

// calculateMicroVMEnterpriseBilling calculates billing for MicroVM Enterprise plan
// This plan has discounted rates and included compute budget
func calculateMicroVMEnterpriseBilling(requests int, computeSeconds float64, memoryMB int, memorySeconds float64) *MicroVMBilling {
	baseFeeCents := int(MicroVMEnterpriseBaseFeeMonthly * 100)

	// Request charges: $0.30 per 10K requests (discounted for high volume)
	requestCharges := (requests / 10000) * MicroVMEnterpriseRequestsPer10K
	if requests%10000 > 0 {
		requestCharges += MicroVMEnterpriseRequestsPer10K
	}

	// Compute charges: $0.01 per vCPU-second (discounted)
	// Using 4 vCPU as default for MicroVM Enterprise
	computeCharges := int(computeSeconds * float64(MicroVMEnterpriseMicroVMCpuSecond*MicroVMEnterpriseDefaultVCPU))

	// Memory charges: $0.001 per GB-second (discounted)
	memoryGBSeconds := memorySeconds * float64(memoryMB) / 1024.0
	memoryCharges := int(memoryGBSeconds * float64(MicroVMEnterpriseMemoryGbSecond))

	total := baseFeeCents + requestCharges + computeCharges + memoryCharges

	// Calculate used budget
	usedBudget := requestCharges + computeCharges + memoryCharges
	if usedBudget > MicroVMEnterpriseComputeBudgetCents {
		usedBudget = MicroVMEnterpriseComputeBudgetCents
	}

	return &MicroVMBilling{
		BaseFeeMonthly: baseFeeCents,
		RequestCharges: requestCharges,
		ComputeCharges: computeCharges,
		MemoryCharges:  memoryCharges,
		TotalCents:     total,
		IncludedBudget: MicroVMEnterpriseComputeBudgetCents,
		UsedBudget:     usedBudget,
	}
}

// ==================== Usage-Based Billing Tiers ====================

// UsagePricingTier defines usage-based pricing structure
type UsagePricingTier struct {
	Name                    string  // Tier name
	IncludedRequestsMonthly int     // Monthly request allowance
	MonthlyPriceCents       int     // Base monthly price
	OveragePricePer1000     int     // Overage price per 1000 requests (cents)
	AnnualDiscountPercent   float64 // Discount for annual commitment (0.0-1.0)
	MaxRequestsPerMonth     int     // -1 for unlimited
}

// Usage-based pricing tiers for the main platform - 2026 optimized (Option B Balanced)
var UsagePricingTiers = map[string]UsagePricingTier{
"free": {
		Name:                    "Free",
		IncludedRequestsMonthly: 25_000,
		MonthlyPriceCents:       0,
		OveragePricePer1000:     0, // Hard stop
		MaxRequestsPerMonth:     25_000,
	},
	"starter": {
		Name:                    "Starter",
		IncludedRequestsMonthly: 250_000,
		MonthlyPriceCents:       StarterPriceCents, // $24/month
		OveragePricePer1000:     StarterOveragePer1000Cents, // $0.15/1K
		AnnualDiscountPercent:   0.17, // 17% off (2 months free)
		MaxRequestsPerMonth:     -1,
	},
	"professional": {
		Name:                    "Professional",
		IncludedRequestsMonthly: 2_500_000,
		MonthlyPriceCents:       ProPriceCents, // $79/month
		OveragePricePer1000:     ProOveragePer1000Cents, // $0.08/1K
		AnnualDiscountPercent:   0.17, // 17% off
		MaxRequestsPerMonth:     -1,
	},
	"enterprise": {
		Name:                    "Enterprise",
		IncludedRequestsMonthly: 25_000_000,
		MonthlyPriceCents:       EnterprisePriceCents, // $299/month
		OveragePricePer1000:     EnterpriseOveragePer1000Cents, // $0.05/1K
		AnnualDiscountPercent:   0.17, // 17% off
		MaxRequestsPerMonth:     -1,
	},
	"agent_enterprise": {
		Name:                    "Agent Enterprise",
		IncludedRequestsMonthly: -1, // Unlimited
		MonthlyPriceCents:       AgentEnterprisePriceCents, // $499/month
		OveragePricePer1000:     0, // No overage (unlimited)
		AnnualDiscountPercent:   0.17, // 17% off
		MaxRequestsPerMonth:     -1,
	},
}

// GetUsagePricingTier returns the usage pricing for a plan
func GetUsagePricingTier(plan string) UsagePricingTier {
	if tier, ok := UsagePricingTiers[plan]; ok {
		return tier
	}
	// Default to starter if not found
	return UsagePricingTiers["starter"]
}

// CalculateUsageOverage calculates the overage charge for exceeding the included requests
func CalculateUsageOverage(plan string, requests int) int {
	tier := GetUsagePricingTier(plan)
	if tier.OveragePricePer1000 == 0 {
		return 0 // No overage billing
	}

	if requests <= tier.IncludedRequestsMonthly {
		return 0
	}

	overage := requests - tier.IncludedRequestsMonthly
	// Round up to nearest 1000
	units := (overage + 999) / 1000
	return units * tier.OveragePricePer1000
}

// GetAnnualPrice returns the price with annual commitment discount
func (t *UsagePricingTier) GetAnnualPrice() int {
	if t.AnnualDiscountPercent <= 0 {
		return t.MonthlyPriceCents * 12
	}
	return int(float64(t.MonthlyPriceCents)*(1-t.AnnualDiscountPercent)) * 12
}

// GetMonthlyWithAnnual returns monthly price when paying annually
func (t *UsagePricingTier) GetMonthlyWithAnnual() int {
	return t.GetAnnualPrice() / 12
}

// GetOverageRateDisplay returns a human-readable overage rate
func (t *UsagePricingTier) GetOverageRateDisplay() string {
	if t.OveragePricePer1000 == 0 {
		return "Hard stop"
	}
	return fmt.Sprintf("$%.4f/request", float64(t.OveragePricePer1000)/1000)
}

// ==================== Time Machine Limits ====================

// GetReplayWindowHours returns the maximum replay window in hours for a plan.
// Returns -1 for unlimited (agent enterprise).
func GetReplayWindowHours(plan string) int {
	switch plan {
	case PlanAgentEnterprise:
		return -1 // Unlimited
	case PlanEnterprise, PlanAgentPro:
		return EnterpriseReplayWindowDays * 24
	case PlanPro, PlanAgentScale:
		return ProReplayWindowDays * 24
	case PlanStarter, PlanAgentStarter:
		return StarterReplayWindowHours
	default: // Free
		return FreeReplayWindowHours
	}
}

// GetMaxExecutionsPerReplay returns the maximum number of executions per replay job.
// Returns -1 for unlimited.
func GetMaxExecutionsPerReplay(plan string) int {
	switch plan {
	case PlanAgentEnterprise:
		return -1
	case PlanEnterprise, PlanAgentPro:
		return EnterpriseMaxExecutionsPerReplay
	case PlanPro, PlanAgentScale:
		return ProMaxExecutionsPerReplay
	case PlanStarter, PlanAgentStarter:
		return StarterMaxExecutionsPerReplay
	default:
		return FreeMaxExecutionsPerReplay
	}
}

// GetMaxConcurrentReplays returns the maximum concurrent replay jobs per tenant.
// Returns -1 for unlimited.
func GetMaxConcurrentReplays(plan string) int {
	switch plan {
	case PlanAgentEnterprise:
		return -1
	case PlanEnterprise, PlanAgentPro:
		return EnterpriseMaxConcurrentReplays
	case PlanPro, PlanAgentScale:
		return ProMaxConcurrentReplays
	case PlanStarter, PlanAgentStarter:
		return StarterMaxConcurrentReplays
	default:
		return FreeMaxConcurrentReplays
	}
}

// GetReplayDataRetentionDays returns how long replay data is retained.
// Returns -1 for unlimited.
func GetReplayDataRetentionDays(plan string) int {
	switch plan {
	case PlanAgentEnterprise:
		return -1
	case PlanEnterprise, PlanAgentPro:
		return EnterpriseReplayDataRetentionDays
	case PlanPro, PlanAgentScale:
		return ProReplayDataRetentionDays
	case PlanStarter, PlanAgentStarter:
		return StarterReplayDataRetentionDays
	default:
		return FreeReplayDataRetentionDays
	}
}

// SupportsLiveReconciliation returns true if the plan supports live (non-dry-run) reconciliation.
func SupportsLiveReconciliation(plan string) bool {
	return plan == PlanEnterprise || plan == PlanAgentPro || plan == PlanAgentEnterprise
}

// SupportsAuditCertificates returns true if the plan supports audit certificate generation.
func SupportsAuditCertificates(plan string) bool {
	return plan == PlanEnterprise || plan == PlanAgentPro || plan == PlanAgentEnterprise
}

// SupportsReplayScheduling returns true if the plan supports scheduled replays.
func SupportsReplayScheduling(plan string) bool {
	return plan == PlanEnterprise || plan == PlanAgentPro || plan == PlanAgentEnterprise
}

// SupportsFullDiffReport returns true if the plan supports structured diff reports.
func SupportsFullDiffReport(plan string) bool {
	switch plan {
	case PlanPro, PlanAgentScale, PlanEnterprise, PlanAgentPro, PlanAgentEnterprise:
		return true
	default:
		return false
	}
}

// Support response time constants (in hours)
const (
	CommunitySupportResponseHours     = 72 // 72 hour response for free/starter
	PremiumSupportResponseHours        = 4  // 4 hour response for pro
	PrioritySupportResponseHours       = 1  // 1 hour response for enterprise
	PremiumSupportAvailabilityHours    = 24 // 24/7 availability for premium support
)

// GetSupportResponseHours returns the expected support response time in hours for a plan
func GetSupportResponseHours(plan string) int {
	switch plan {
	case PlanPro, PlanAgentScale:
		return PremiumSupportResponseHours
	case PlanEnterprise, PlanAgentPro, PlanAgentEnterprise:
		return PrioritySupportResponseHours
	default:
		return CommunitySupportResponseHours
	}
}

// IsPremiumSupportAvailable returns true if premium support (24/7 priority) is available for the plan
func IsPremiumSupportAvailable(plan string) bool {
	switch plan {
	case PlanPro, PlanAgentScale, PlanEnterprise, PlanAgentPro, PlanAgentEnterprise:
		return true
	default:
		return false
	}
}

// TimeMachineLimits provides a snapshot of all Time Machine limits for a plan.
type TimeMachineLimits struct {
	ReplayWindowHours      int  `json:"replay_window_hours"`
	MaxExecutionsPerReplay int  `json:"max_executions_per_replay"`
	MaxConcurrentReplays   int  `json:"max_concurrent_replays"`
	DataRetentionDays      int  `json:"data_retention_days"`
	AutoReconciliation     bool `json:"auto_reconciliation"`
	LiveReconciliation     bool `json:"live_reconciliation"`
	AuditCertificates      bool `json:"audit_certificates"`
	ReplayScheduling       bool `json:"replay_scheduling"`
	FullDiffReports        bool `json:"full_diff_reports"`
	IncidentInsurance      bool `json:"incident_insurance"`
	Unlimited              bool `json:"unlimited"`
}

// GetTimeMachineLimits returns all Time Machine limits for a plan.
func GetTimeMachineLimits(plan string) TimeMachineLimits {
	return TimeMachineLimits{
		ReplayWindowHours:      GetReplayWindowHours(plan),
		MaxExecutionsPerReplay: GetMaxExecutionsPerReplay(plan),
		MaxConcurrentReplays:   GetMaxConcurrentReplays(plan),
		DataRetentionDays:      GetReplayDataRetentionDays(plan),
		AutoReconciliation:     SupportsLiveReconciliation(plan),
		LiveReconciliation:     SupportsLiveReconciliation(plan),
		AuditCertificates:      SupportsAuditCertificates(plan),
		ReplayScheduling:       SupportsReplayScheduling(plan),
		FullDiffReports:        SupportsFullDiffReport(plan),
		IncidentInsurance:      plan == PlanAgentEnterprise,
		Unlimited:              plan == PlanAgentEnterprise,
	}
}

// MaxStateFabricsPerPlan returns the maximum number of state fabrics allowed for a plan
func MaxStateFabricsPerPlan(plan string) int {
	switch plan {
	case PlanEnterprise:
		return EnterpriseMaxStateFabrics
	case PlanPro:
		return ProMaxStateFabrics
	case PlanStarter:
		return StarterMaxStateFabrics
	default:
		return FreeMaxStateFabrics
	}
}

// PlanHasStateFabricFeature returns true if the plan includes the State Fabric feature.
// A plan has the feature if it is explicitly enabled (max > 0) or unlimited (max < 0).
func PlanHasStateFabricFeature(plan string) bool {
	return MaxStateFabricsPerPlan(plan) != 0
}

// ============================================================================
// Vault (Secrets) Plan Features
// ============================================================================
// Unified vault features mapped to main platform plans:
// - Free/Starter: Basic vault (expiration, namespaces)
// - Professional: + MFA, IP allowlist, break-glass, audit export
// - Enterprise: + Escrow, RBAC, shares, SIEM webhooks
// - Agent Enterprise: + SSO, HA status

// VaultFeature represents a vault-specific feature
type VaultFeature string

const (
	VaultFeatureMFA           VaultFeature = "mfa"
	VaultFeatureIPAllowlist    VaultFeature = "ip_allowlist"
	VaultFeatureExpiration     VaultFeature = "expiration"
	VaultFeatureBreakGlass     VaultFeature = "break_glass"
	VaultFeatureEscrow         VaultFeature = "escrow"
	VaultFeatureRBAC           VaultFeature = "rbac"
	VaultFeatureNamespaces     VaultFeature = "namespaces"
	VaultFeatureShares         VaultFeature = "shares"
	VaultFeatureSSO            VaultFeature = "sso"
	VaultFeatureSIEMWebhooks   VaultFeature = "siem_webhooks"
	VaultFeatureAuditExport    VaultFeature = "audit_export"
	VaultFeatureCacheStats     VaultFeature = "cache_stats"
	VaultFeatureHAStatus       VaultFeature = "ha_status"
	VaultFeatureDependencyGraph VaultFeature = "dependency_graph"
	VaultFeatureTokenMonitor   VaultFeature = "token_monitor"
)

// SupportsVaultMFA returns true if the plan supports vault MFA
func SupportsVaultMFA(plan string) bool {
	return plan == PlanPro || plan == PlanEnterprise || plan == PlanAgentEnterprise
}

// SupportsVaultIPAllowlist returns true if the plan supports IP allowlisting for vault tokens
func SupportsVaultIPAllowlist(plan string) bool {
	return plan == PlanPro || plan == PlanEnterprise || plan == PlanAgentEnterprise
}

// SupportsVaultBreakGlass returns true if the plan supports break-glass emergency access
func SupportsVaultBreakGlass(plan string) bool {
	return plan == PlanPro || plan == PlanEnterprise || plan == PlanAgentEnterprise
}

// SupportsVaultEscrow returns true if the plan supports key escrow
func SupportsVaultEscrow(plan string) bool {
	return plan == PlanEnterprise || plan == PlanAgentEnterprise
}

// SupportsVaultRBAC returns true if the plan supports RBAC for vault
func SupportsVaultRBAC(plan string) bool {
	return plan == PlanEnterprise || plan == PlanAgentEnterprise
}

// SupportsVaultShares returns true if the plan supports cross-tenant secret sharing
func SupportsVaultShares(plan string) bool {
	return plan == PlanEnterprise || plan == PlanAgentEnterprise
}

// SupportsVaultSSO returns true if the plan supports SSO for vault
func SupportsVaultSSO(plan string) bool {
	return plan == PlanAgentEnterprise
}

// SupportsVaultSIEMWebhooks returns true if the plan supports SIEM webhooks
func SupportsVaultSIEMWebhooks(plan string) bool {
	return plan == PlanEnterprise || plan == PlanAgentEnterprise
}

// SupportsVaultAuditExport returns true if the plan supports audit log export
func SupportsVaultAuditExport(plan string) bool {
	return plan == PlanPro || plan == PlanEnterprise || plan == PlanAgentEnterprise
}

// SupportsVaultHAStatus returns true if the plan supports HA status monitoring
func SupportsVaultHAStatus(plan string) bool {
	return plan == PlanAgentEnterprise
}

// SupportsVaultNamespaces returns true if the plan supports hierarchical namespaces for organizing secrets
func SupportsVaultNamespaces(plan string) bool {
	return plan == PlanPro || plan == PlanEnterprise || plan == PlanAgentEnterprise
}

// GetMaxDynamicCreds returns the maximum dynamic credentials for a plan (30-day rolling window)
func GetMaxDynamicCreds(plan string) int {
	switch plan {
	case PlanFree:
		return FreeMaxDynamicCreds
	case PlanStarter:
		return StarterMaxDynamicCreds
	case PlanPro:
		return ProMaxDynamicCreds
	case PlanEnterprise:
		return EnterpriseMaxDynamicCreds
	default:
		return FreeMaxDynamicCreds
	}
}

// GetMaxAuditExportsPerDay returns the maximum audit exports per day for a plan
func GetMaxAuditExportsPerDay(plan string) int {
	switch plan {
	case PlanFree:
		return FreeMaxAuditExports
	case PlanStarter:
		return StarterMaxAuditExports
	case PlanPro:
		return ProMaxAuditExports
	case PlanEnterprise:
		return EnterpriseMaxAuditExports
	default:
		return FreeMaxAuditExports
	}
}

// VaultLimits provides a snapshot of all vault limits for a plan
type VaultLimits struct {
	MaxSecrets         int  `json:"max_secrets"`
	MaxDynamicCreds    int  `json:"max_dynamic_creds"`
	MaxTokensPerSecret int  `json:"max_tokens_per_secret"`
	MaxAuditExports    int  `json:"max_audit_exports_per_day"`
	MFA                bool `json:"mfa"`
	IPAllowlist        bool `json:"ip_allowlist"`
	BreakGlass         bool `json:"break_glass"`
	Escrow             bool `json:"escrow"`
	RBAC               bool `json:"rbac"`
	Shares             bool `json:"shares"`
	SSO                bool `json:"sso"`
	SIEMWebhooks       bool `json:"siem_webhooks"`
	AuditExport        bool `json:"audit_export"`
	HAStatus           bool `json:"ha_status"`
}

// GetVaultLimits returns all vault limits for a plan
func GetVaultLimits(plan string) VaultLimits {
	return VaultLimits{
		MaxSecrets:         GetMaxSecrets(plan),
		MaxDynamicCreds:    GetMaxDynamicCreds(plan),
		MaxTokensPerSecret: GetMaxTokensPerSecret(plan),
		MaxAuditExports:    GetMaxAuditExportsPerDay(plan),
		MFA:                SupportsVaultMFA(plan),
		IPAllowlist:        SupportsVaultIPAllowlist(plan),
		BreakGlass:         SupportsVaultBreakGlass(plan),
		Escrow:             SupportsVaultEscrow(plan),
		RBAC:               SupportsVaultRBAC(plan),
		Shares:             SupportsVaultShares(plan),
		SSO:                SupportsVaultSSO(plan),
		SIEMWebhooks:       SupportsVaultSIEMWebhooks(plan),
		AuditExport:        SupportsVaultAuditExport(plan),
		HAStatus:           SupportsVaultHAStatus(plan),
	}
}

// ============================================================================
// State Fabric Add-ons (Premium Stackable Add-ons)
// ============================================================================
// SF is now a bundled feature in platform plans. Add-ons provide premium capabilities.
// Add-ons are available on any paid plan and stack on top of base SF limits.

// SF Add-on IDs
const (
	SFAddOnHotCache            = "sf_hot_cache"
	SFAddOnMultiRegion         = "sf_multi_region"
	SFAddOnAIRecall            = "sf_ai_recall"
	SFAddOnAdvancedInsights    = "sf_advanced_insights"
	SFAddOnAdvancedSecurity    = "sf_advanced_security"
)

// SF Add-on pricing (cents/month)
const (
	SFAddOnHotCachePriceCents          = 4900  // $49/mo per 5GB
	SFAddOnMultiRegionPriceCents       = 9900  // $99/mo
	SFAddOnAIRecallPriceCents          = 14900 // $149/mo
	SFAddOnAdvancedInsightsPriceCents  = 7900  // $79/mo
	SFAddOnAdvancedSecurityPriceCents  = 9900  // $99/mo
)

// SFAddOnInfo represents a State Fabric add-on's metadata
type SFAddOnInfo struct {
	ID          string `json:"id"`
	Name        string `json:"name"`
	Description string `json:"description"`
	PriceCents  int    `json:"price_cents"`
	StripePriceID string `json:"stripe_price_id,omitempty"`
}

// GetAllSFAddOns returns all available SF add-ons
func GetAllSFAddOns() []SFAddOnInfo {
	return []SFAddOnInfo{
		{
			ID:          SFAddOnHotCache,
			Name:        "Hot Cache Booster",
			Description: "Reduces replay and read costs. 5GB hot cache tier.",
			PriceCents:  SFAddOnHotCachePriceCents,
		},
		{
			ID:          SFAddOnMultiRegion,
			Name:        "Multi-Region Replication",
			Description: "Active-active replication across regions for HA and global latency.",
			PriceCents:  SFAddOnMultiRegionPriceCents,
		},
		{
			ID:          SFAddOnAIRecall,
			Name:        "AI Memory Pack",
			Description: "Vector index, embeddings storage, fast read engine for AI recall.",
			PriceCents:  SFAddOnAIRecallPriceCents,
		},
		{
			ID:          SFAddOnAdvancedInsights,
			Name:        "Advanced Insights",
			Description: "Cost forecasting, anomaly detection, hot path alerts.",
			PriceCents:  SFAddOnAdvancedInsightsPriceCents,
		},
		{
			ID:          SFAddOnAdvancedSecurity,
			Name:        "Advanced Security Pack",
			Description: "SOC2-friendly logs, key rotation, audit streams.",
			PriceCents:  SFAddOnAdvancedSecurityPriceCents,
		},
	}
}

// GetSFAddOnPrice returns the price in cents for an add-on
func GetSFAddOnPrice(addOnID string) int {
	switch addOnID {
	case SFAddOnHotCache:
		return SFAddOnHotCachePriceCents
	case SFAddOnMultiRegion:
		return SFAddOnMultiRegionPriceCents
	case SFAddOnAIRecall:
		return SFAddOnAIRecallPriceCents
	case SFAddOnAdvancedInsights:
		return SFAddOnAdvancedInsightsPriceCents
	case SFAddOnAdvancedSecurity:
		return SFAddOnAdvancedSecurityPriceCents
	default:
		return 0
	}
}

// IsValidSFAddOn returns true if the add-on ID is valid
func IsValidSFAddOn(addOnID string) bool {
	switch addOnID {
	case SFAddOnHotCache, SFAddOnMultiRegion, SFAddOnAIRecall,
		SFAddOnAdvancedInsights, SFAddOnAdvancedSecurity:
		return true
	}
	return false
}

// SupportsSFAddOn returns true if the plan can use SF add-ons
// All paid plans (Starter+) can purchase add-ons
func SupportsSFAddOn(plan string) bool {
	return plan != PlanFree
}

// SupportsSFAddOnPurchase returns true if the plan can purchase a specific SF add-on
// All paid plans can purchase any SF add-on
func SupportsSFAddOnPurchase(plan string, addOnID string) bool {
	if !IsValidSFAddOn(addOnID) {
		return false
	}
	return plan != PlanFree
}

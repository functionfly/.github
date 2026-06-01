package plans

import (
	"fmt"
	"os"
	"strconv"
	"strings"
)

const (
	// App limits per tenant
	StarterMaxApps    = 3
	ProMaxApps        = 10
	EnterpriseMaxApps = -1 // Unlimited

	// Provider limits per app
	StarterMaxProvidersPerApp    = 2
	ProMaxProvidersPerApp        = 3
	EnterpriseMaxProvidersPerApp = 5

	// Default request limits per month
	// NOTE: These must match the frontend PLANS limits in web/dashboard/src/lib/constants.ts
	StarterMaxRequestsPerMonth           = 1_000_000   // 1M requests - matches frontend Starter
	DefaultProMaxRequestsPerMonth        = 10_000_000 // 10M requests - matches frontend Professional
	DefaultEnterpriseMaxRequestsPerMonth = -1           // Unlimited (-1 = unlimited in our system)

	// Secrets limits per tenant
	StarterMaxSecrets    = 10
	ProMaxSecrets        = 50
	EnterpriseMaxSecrets = 10000 // Effectively unlimited

	// Token limits per secret
	StarterMaxTokensPerSecret    = 5
	ProMaxTokensPerSecret        = 20
	EnterpriseMaxTokensPerSecret = 100

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

	// Enterprise tier pricing
	EnterpriseBaseFeeMonthly   = 99.00 // $/month
	EnterpriseRequestsPer10K   = 5000  // $0.50 per 10K requests (cents)
	EnterpriseMicroVMCpuSecond = 2     // $0.02 per vCPU-second (cents)
	EnterpriseMemoryGbSecond   = 2     // $0.002 per GB-second (cents)

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
)

// StateFabric limits
const (
	FreeMaxStateFabrics       = 0
	StarterMaxStateFabrics    = 3
	ProMaxStateFabrics         = 10
	EnterpriseMaxStateFabrics  = -1 // Unlimited
)

// Plan type constants
// NOTE: "professional" must match frontend plan-utils.ts PlanTier type
const (
	PlanFree       = "free" // Default plan for new signups
	PlanStarter    = "starter"
	PlanPro        = "professional" // Was "pro" - changed for consistency with frontend
	PlanEnterprise = "enterprise"
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
	StarterPriceCents        = 2400  // $24/month
	ProPriceCents            = 7900  // $79/month
	EnterprisePriceCents     = 29900 // $299/month base (includes 5M AI calls)
	AgentEnterprisePriceCents = 49900 // $499/month - unlimited AI
)

// Annual Pricing (2 months free = 10 months billed)
const (
	StarterAnnualCents        = 24000 // $240/year ($24/mo equiv)
	ProAnnualCents            = 79000 // $790/year ($79/mo equiv)
	EnterpriseAnnualCents     = 299000 // $2990/year ($299/mo equiv)
	AgentEnterpriseAnnualCents = 499000 // $4990/year ($499/mo equiv)
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
	RuntimePython        = "python"         // RustPython
	RuntimePythonMicroVM = "python-microvm" // CPython in Firecracker
	// FlyPy runtime has been disabled - using MicroPython only
)

// IsEnterpriseTier returns true if the plan is enterprise
func IsEnterpriseTier(plan string) bool {
	return plan == PlanEnterprise
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
	case RuntimePythonMicroVM:
		// python-microvm is only available for enterprise tier
		return plan == PlanEnterprise
	default:
		return true // All other runtimes are available for all plans
	}
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
	case PlanPro:
		return ProMaxApps
	case PlanEnterprise:
		return EnterpriseMaxApps
	case PlanStarter:
		fallthrough
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
	case PlanPro:
		return ProMaxSecrets
	case PlanEnterprise:
		return EnterpriseMaxSecrets
	case PlanStarter:
		fallthrough
	default:
		return StarterMaxSecrets
	}
}

// GetMaxTokensPerSecret returns the maximum number of access tokens allowed per secret for the given plan
func GetMaxTokensPerSecret(plan string) int {
	switch plan {
	case PlanPro:
		return ProMaxTokensPerSecret
	case PlanEnterprise:
		return EnterpriseMaxTokensPerSecret
	case PlanStarter:
		fallthrough
	default:
		return StarterMaxTokensPerSecret
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
	MaxMicroVMs     int
	DefaultMemoryMB int
	MaxMemoryMB     int
	DefaultVCPU     int
	MaxVCPU         int
	DefaultTimeout  int
	MaxTimeout      int
}

// GetMicroVMLimits returns the MicroVM resource limits for the given plan
// Returns nil for non-enterprise plans (MicroVMs not available)
func GetMicroVMLimits(plan string) *MicroVMLimits {
	if plan != PlanEnterprise {
		return nil
	}
	return &MicroVMLimits{
		MaxMicroVMs:     EnterpriseMaxMicroVMs,
		DefaultMemoryMB: EnterpriseDefaultMemoryMB,
		MaxMemoryMB:     EnterpriseMaxMemoryMB,
		DefaultVCPU:     EnterpriseDefaultVCPU,
		MaxVCPU:         EnterpriseMaxVCPU,
		DefaultTimeout:  EnterpriseDefaultTimeoutMs,
		MaxTimeout:      EnterpriseMaxTimeoutMs,
	}
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
	BaseFeeMonthly int
	RequestCharges int
	ComputeCharges int
	MemoryCharges  int
	TotalCents     int
}

// CalculateMicroVMBilling calculates the billing for MicroVM usage
func CalculateMicroVMBilling(plan string, requests int, computeSeconds float64, memoryMB int, memorySeconds float64) *MicroVMBilling {
	if plan != PlanEnterprise {
		return nil
	}

	// Base fee
	baseFeeCents := int(EnterpriseBaseFeeMonthly * 100)

	// Request charges: $0.50 per 10K requests
	requestCharges := (requests / 10000) * EnterpriseRequestsPer10K
	if requests%10000 > 0 {
		requestCharges += EnterpriseRequestsPer10K
	}

	// Compute charges: $0.02 per vCPU-second
	// Using 2 vCPU as default
	computeCharges := int(computeSeconds * float64(EnterpriseMicroVMCpuSecond*EnterpriseDefaultVCPU))

	// Memory charges: $0.002 per GB-second
	// Convert MB to GB-seconds
	memoryGBSeconds := memorySeconds * float64(memoryMB) / 1024.0
	memoryCharges := int(memoryGBSeconds * float64(EnterpriseMemoryGbSecond))

	total := baseFeeCents + requestCharges + computeCharges + memoryCharges

	return &MicroVMBilling{
		BaseFeeMonthly: baseFeeCents,
		RequestCharges: requestCharges,
		ComputeCharges: computeCharges,
		MemoryCharges:  memoryCharges,
		TotalCents:     total,
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

// Usage-based pricing tiers for the main platform - 2026 optimized (Option C Hybrid)
var UsagePricingTiers = map[string]UsagePricingTier{
"free": {
		Name:                    "Free",
		IncludedRequestsMonthly: 100_000,
		MonthlyPriceCents:       0,
		OveragePricePer1000:     0, // Hard stop
		MaxRequestsPerMonth:     100_000,
	},
	"starter": {
		Name:                    "Starter",
		IncludedRequestsMonthly: 100_000,
		MonthlyPriceCents:       StarterPriceCents, // $24/month
		OveragePricePer1000:     StarterOveragePer1000Cents, // $0.15/1K
		AnnualDiscountPercent:   0.17, // 17% off (2 months free)
		MaxRequestsPerMonth:     -1,
	},
	"professional": {
		Name:                    "Professional",
		IncludedRequestsMonthly: 1_000_000,
		MonthlyPriceCents:       ProPriceCents, // $79/month
		OveragePricePer1000:     ProOveragePer1000Cents, // $0.08/1K
		AnnualDiscountPercent:   0.17, // 17% off
		MaxRequestsPerMonth:     -1,
	},
	"enterprise": {
		Name:                    "Enterprise",
		IncludedRequestsMonthly: 10_000_000,
		MonthlyPriceCents:       EnterprisePriceCents, // $199/month
		OveragePricePer1000:     EnterpriseOveragePer1000Cents, // $0.04/1K
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

// PlanHasStateFabricFeature returns true if the plan includes the State Fabric feature
func PlanHasStateFabricFeature(plan string) bool {
	return MaxStateFabricsPerPlan(plan) > 0
}

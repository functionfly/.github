package plans

import (
	"fmt"
	"os"
	"strconv"
)

const (
	// Provider limits per app
	StarterMaxProvidersPerApp    = 2
	ProMaxProvidersPerApp        = 3
	EnterpriseMaxProvidersPerApp = 5

	// Default request limits per month
	StarterMaxRequestsPerMonth           = 100_000
	DefaultProMaxRequestsPerMonth        = 500_000
	DefaultEnterpriseMaxRequestsPerMonth = 10_000_000

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
)

// Plan type constants
const (
	PlanStarter    = "starter"
	PlanPro        = "pro"
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
const (
	AgentStarterMaxCallsPerMonth    = 500_000
	AgentScaleMaxCallsPerMonth      = 5_000_000
	AgentProMaxCallsPerMonth        = 25_000_000
	AgentEnterpriseMaxCallsPerMonth = -1 // Custom
)

// AEP Calls per minute limits
const (
	AgentStarterMaxCallsPerMinute    = 100
	AgentScaleMaxCallsPerMinute      = 500
	AgentProMaxCallsPerMinute        = 2000
	AgentEnterpriseMaxCallsPerMinute = -1 // Custom
)

// AEP Calls per day limits
const (
	AgentStarterMaxCallsPerDay    = 16_667
	AgentScaleMaxCallsPerDay      = 166_667
	AgentProMaxCallsPerDay        = 833_333
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

// AEP Monthly pricing (cents)
const (
	AgentStarterPriceCents    = 4900   // $49/month
	AgentScalePriceCents      = 29900  // $299/month
	AgentProPriceCents        = 99900  // $999/month
	AgentEnterprisePriceCents = 250000 // $2500+/month base
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

// AgentMaxCallsPerMinute returns the calls-per-minute limit for an AEP plan
func AgentMaxCallsPerMinute(plan string) int {
	switch plan {
	case PlanAgentScale:
		return AgentScaleMaxCallsPerMinute
	case PlanAgentPro:
		return AgentProMaxCallsPerMinute
	case PlanAgentEnterprise:
		return AgentEnterpriseMaxCallsPerMinute
	default:
		return AgentStarterMaxCallsPerMinute
	}
}

// AgentMaxCallsPerDay returns the calls-per-day limit for an AEP plan
func AgentMaxCallsPerDay(plan string) int {
	switch plan {
	case PlanAgentScale:
		return AgentScaleMaxCallsPerDay
	case PlanAgentPro:
		return AgentProMaxCallsPerDay
	case PlanAgentEnterprise:
		return AgentEnterpriseMaxCallsPerDay
	default:
		return AgentStarterMaxCallsPerDay
	}
}

// AgentMaxConcurrency returns the max concurrency for an AEP plan
func AgentMaxConcurrency(plan string) int {
	switch plan {
	case PlanAgentScale:
		return AgentScaleMaxConcurrency
	case PlanAgentPro:
		return AgentProMaxConcurrency
	case PlanAgentEnterprise:
		return AgentEnterpriseMaxConcurrency
	default:
		return AgentStarterMaxConcurrency
	}
}

// AgentLogRetentionDays returns the log retention days for an AEP plan
func AgentLogRetentionDays(plan string) int {
	switch plan {
	case PlanAgentScale:
		return AgentScaleLogRetentionDays
	case PlanAgentPro:
		return AgentProLogRetentionDays
	case PlanAgentEnterprise:
		return AgentEnterpriseLogRetentionDays
	default:
		return AgentStarterLogRetentionDays
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

package plans

import (
	"testing"
)

func TestGetMicroVMLimits_Enterprise(t *testing.T) {
	limits := GetMicroVMLimits(PlanEnterprise)
	if limits == nil {
		t.Fatal("expected non-nil limits for enterprise plan")
	}
	if limits.MaxMicroVMs != EnterpriseMaxMicroVMs {
		t.Errorf("MaxMicroVMs = %d, want %d", limits.MaxMicroVMs, EnterpriseMaxMicroVMs)
	}
	if limits.DefaultMemoryMB != EnterpriseDefaultMemoryMB {
		t.Errorf("DefaultMemoryMB = %d, want %d", limits.DefaultMemoryMB, EnterpriseDefaultMemoryMB)
	}
	if limits.MaxMemoryMB != EnterpriseMaxMemoryMB {
		t.Errorf("MaxMemoryMB = %d, want %d", limits.MaxMemoryMB, EnterpriseMaxMemoryMB)
	}
	if limits.DefaultVCPU != EnterpriseDefaultVCPU {
		t.Errorf("DefaultVCPU = %d, want %d", limits.DefaultVCPU, EnterpriseDefaultVCPU)
	}
	if limits.MaxVCPU != EnterpriseMaxVCPU {
		t.Errorf("MaxVCPU = %d, want %d", limits.MaxVCPU, EnterpriseMaxVCPU)
	}
	if limits.DefaultTimeout != EnterpriseDefaultTimeoutMs {
		t.Errorf("DefaultTimeout = %d, want %d", limits.DefaultTimeout, EnterpriseDefaultTimeoutMs)
	}
	if limits.MaxTimeout != EnterpriseMaxTimeoutMs {
		t.Errorf("MaxTimeout = %d, want %d", limits.MaxTimeout, EnterpriseMaxTimeoutMs)
	}
}

func TestGetMicroVMLimits_NonEnterprise(t *testing.T) {
	cases := []string{PlanFree, PlanStarter, PlanPro, PlanAgentStarter, PlanAgentScale, PlanAgentPro, PlanAgentEnterprise}
	for _, plan := range cases {
		t.Run(plan, func(t *testing.T) {
			limits := GetMicroVMLimits(plan)
			if limits != nil {
				t.Errorf("expected nil limits for plan %s, got %+v", plan, limits)
			}
		})
	}
}

func TestValidateMicroVMResources_Valid(t *testing.T) {
	cases := []struct {
		name      string
		memoryMB  int
		vCPU      int
		timeoutMs int
	}{
		{"minimum values", 256, 1, 1000},
		{"default values", EnterpriseDefaultMemoryMB, EnterpriseDefaultVCPU, EnterpriseDefaultTimeoutMs},
		{"maximum values", EnterpriseMaxMemoryMB, EnterpriseMaxVCPU, EnterpriseMaxTimeoutMs},
		{"mid-range", 1024, 2, 60000},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			err := ValidateMicroVMResources(PlanEnterprise, tc.memoryMB, tc.vCPU, tc.timeoutMs)
			if err != nil {
				t.Errorf("unexpected error: %v", err)
			}
		})
	}
}

func TestValidateMicroVMResources_NonEnterprise(t *testing.T) {
	cases := []string{PlanFree, PlanStarter, PlanPro}
	for _, plan := range cases {
		t.Run(plan, func(t *testing.T) {
			err := ValidateMicroVMResources(plan, 512, 2, 30000)
			if err == nil {
				t.Fatal("expected error for non-enterprise plan")
			}
			ve, ok := err.(*ValidationError)
			if !ok {
				t.Fatalf("expected *ValidationError, got %T", err)
			}
			if ve.Code != "MICROVM_NOT_AVAILABLE" {
				t.Errorf("error code = %q, want %q", ve.Code, "MICROVM_NOT_AVAILABLE")
			}
		})
	}
}

func TestValidateMicroVMResources_InvalidMemory(t *testing.T) {
	cases := []struct {
		name     string
		memoryMB int
	}{
		{"too low", 128},
		{"zero", 0},
		{"negative", -1},
		{"too high", EnterpriseMaxMemoryMB + 1},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			err := ValidateMicroVMResources(PlanEnterprise, tc.memoryMB, 2, 30000)
			if err == nil {
				t.Fatal("expected error for invalid memory")
			}
			ve, ok := err.(*ValidationError)
			if !ok {
				t.Fatalf("expected *ValidationError, got %T", err)
			}
			if ve.Code != "INVALID_MEMORY" {
				t.Errorf("error code = %q, want %q", ve.Code, "INVALID_MEMORY")
			}
		})
	}
}

func TestValidateMicroVMResources_InvalidVCPU(t *testing.T) {
	cases := []struct {
		name string
		vCPU int
	}{
		{"zero", 0},
		{"negative", -1},
		{"too high", EnterpriseMaxVCPU + 1},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			err := ValidateMicroVMResources(PlanEnterprise, 512, tc.vCPU, 30000)
			if err == nil {
				t.Fatal("expected error for invalid vCPU")
			}
			ve, ok := err.(*ValidationError)
			if !ok {
				t.Fatalf("expected *ValidationError, got %T", err)
			}
			if ve.Code != "INVALID_VCPU" {
				t.Errorf("error code = %q, want %q", ve.Code, "INVALID_VCPU")
			}
		})
	}
}

func TestValidateMicroVMResources_InvalidTimeout(t *testing.T) {
	cases := []struct {
		name      string
		timeoutMs int
	}{
		{"too low", 500},
		{"zero", 0},
		{"negative", -1},
		{"too high", EnterpriseMaxTimeoutMs + 1},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			err := ValidateMicroVMResources(PlanEnterprise, 512, 2, tc.timeoutMs)
			if err == nil {
				t.Fatal("expected error for invalid timeout")
			}
			ve, ok := err.(*ValidationError)
			if !ok {
				t.Fatalf("expected *ValidationError, got %T", err)
			}
			if ve.Code != "INVALID_TIMEOUT" {
				t.Errorf("error code = %q, want %q", ve.Code, "INVALID_TIMEOUT")
			}
		})
	}
}

func TestCalculateMicroVMBilling_NonEnterprise(t *testing.T) {
	cases := []string{PlanFree, PlanStarter, PlanPro}
	for _, plan := range cases {
		t.Run(plan, func(t *testing.T) {
			result := CalculateMicroVMBilling(plan, 1000, 60.0, 512, 60.0)
			if result != nil {
				t.Errorf("expected nil for plan %s, got %+v", plan, result)
			}
		})
	}
}

func TestCalculateMicroVMBilling_Enterprise(t *testing.T) {
	result := CalculateMicroVMBilling(PlanEnterprise, 0, 0, 0, 0)
	if result == nil {
		t.Fatal("expected non-nil billing for enterprise plan")
	}

	// Base fee should always be present
	expectedBaseFee := int(EnterpriseBaseFeeMonthly * 100) // $99.00 = 9900 cents
	if result.BaseFeeMonthly != expectedBaseFee {
		t.Errorf("BaseFeeMonthly = %d, want %d", result.BaseFeeMonthly, expectedBaseFee)
	}
}

func TestCalculateMicroVMBilling_WithRequests(t *testing.T) {
	// 10,000 requests = 1 * EnterpriseRequestsPer10K
	result := CalculateMicroVMBilling(PlanEnterprise, 10000, 0, 0, 0)
	if result == nil {
		t.Fatal("expected non-nil billing")
	}
	if result.RequestCharges != EnterpriseRequestsPer10K {
		t.Errorf("RequestCharges = %d, want %d", result.RequestCharges, EnterpriseRequestsPer10K)
	}

	// 10,001 requests = 2 * EnterpriseRequestsPer10K (rounds up)
	result2 := CalculateMicroVMBilling(PlanEnterprise, 10001, 0, 0, 0)
	if result2.RequestCharges != 2*EnterpriseRequestsPer10K {
		t.Errorf("RequestCharges for 10001 = %d, want %d", result2.RequestCharges, 2*EnterpriseRequestsPer10K)
	}
}

func TestCalculateMicroVMBilling_WithCompute(t *testing.T) {
	// 100 compute-seconds with 2 vCPUs at $0.02/vCPU-second = 100 * 2 * 2 = 400 cents
	result := CalculateMicroVMBilling(PlanEnterprise, 0, 100.0, 0, 0)
	if result == nil {
		t.Fatal("expected non-nil billing")
	}
	expectedCompute := int(100.0 * float64(EnterpriseMicroVMCpuSecond*EnterpriseDefaultVCPU))
	if result.ComputeCharges != expectedCompute {
		t.Errorf("ComputeCharges = %d, want %d", result.ComputeCharges, expectedCompute)
	}
}

func TestCalculateMicroVMBilling_WithMemory(t *testing.T) {
	// 512 MB for 100 seconds = 512/1024 * 100 = 50 GB-seconds
	// At $0.002/GB-second = 50 * 2 = 100 cents
	result := CalculateMicroVMBilling(PlanEnterprise, 0, 0, 512, 100.0)
	if result == nil {
		t.Fatal("expected non-nil billing")
	}
	memGBSeconds := 100.0 * 512.0 / 1024.0
	expectedMemory := int(memGBSeconds * float64(EnterpriseMemoryGbSecond))
	if result.MemoryCharges != expectedMemory {
		t.Errorf("MemoryCharges = %d, want %d", result.MemoryCharges, expectedMemory)
	}
}

func TestCalculateMicroVMBilling_TotalCents(t *testing.T) {
	result := CalculateMicroVMBilling(PlanEnterprise, 10000, 100.0, 512, 100.0)
	if result == nil {
		t.Fatal("expected non-nil billing")
	}
	expectedTotal := result.BaseFeeMonthly + result.RequestCharges + result.ComputeCharges + result.MemoryCharges
	if result.TotalCents != expectedTotal {
		t.Errorf("TotalCents = %d, want %d (sum of parts: base=%d + req=%d + compute=%d + mem=%d)",
			result.TotalCents, expectedTotal,
			result.BaseFeeMonthly, result.RequestCharges, result.ComputeCharges, result.MemoryCharges)
	}
}

func TestIsValidRuntimeForPlan_MicroVM(t *testing.T) {
	if !IsValidRuntimeForPlan(PlanEnterprise, RuntimePythonMicroVM) {
		t.Error("expected enterprise to allow python-microvm")
	}
	cases := []string{PlanFree, PlanStarter, PlanPro}
	for _, plan := range cases {
		t.Run(plan, func(t *testing.T) {
			if IsValidRuntimeForPlan(plan, RuntimePythonMicroVM) {
				t.Errorf("expected plan %s to NOT allow python-microvm", plan)
			}
		})
	}
}

func TestIsValidRuntimeForPlan_OtherRuntimes(t *testing.T) {
	runtimes := []string{RuntimeWasm, RuntimePython, RuntimePrism, "nodejs", "rust", "go", "c", "cpp"}
	for _, rt := range runtimes {
		for _, plan := range []string{PlanFree, PlanStarter, PlanPro, PlanEnterprise} {
			t.Run(rt+"/"+plan, func(t *testing.T) {
				if !IsValidRuntimeForPlan(plan, rt) {
					t.Errorf("expected plan %s to allow runtime %s", plan, rt)
				}
			})
		}
	}
}

func TestIsEnterpriseTier(t *testing.T) {
	if !IsEnterpriseTier(PlanEnterprise) {
		t.Error("expected PlanEnterprise to be enterprise tier")
	}
	cases := []string{PlanFree, PlanStarter, PlanPro, PlanAgentStarter, PlanAgentScale, PlanAgentPro, PlanAgentEnterprise}
	for _, plan := range cases {
		t.Run(plan, func(t *testing.T) {
			if IsEnterpriseTier(plan) {
				t.Errorf("expected %s to NOT be enterprise tier", plan)
			}
		})
	}
}

func TestValidationError_Error(t *testing.T) {
	ve := &ValidationError{
		Code:    "TEST_CODE",
		Message: "test message",
	}
	if ve.Error() != "test message" {
		t.Errorf("Error() = %q, want %q", ve.Error(), "test message")
	}
}

func TestMicroVMConstants(t *testing.T) {
	if RuntimePythonMicroVM != "python-microvm" {
		t.Errorf("RuntimePythonMicroVM = %q, want %q", RuntimePythonMicroVM, "python-microvm")
	}
	if EnterpriseMaxMicroVMs < 1 {
		t.Errorf("EnterpriseMaxMicroVMs = %d, must be >= 1", EnterpriseMaxMicroVMs)
	}
	if EnterpriseDefaultMemoryMB < 256 {
		t.Errorf("EnterpriseDefaultMemoryMB = %d, must be >= 256", EnterpriseDefaultMemoryMB)
	}
	if EnterpriseMaxMemoryMB < EnterpriseDefaultMemoryMB {
		t.Errorf("EnterpriseMaxMemoryMB (%d) < EnterpriseDefaultMemoryMB (%d)", EnterpriseMaxMemoryMB, EnterpriseDefaultMemoryMB)
	}
	if EnterpriseDefaultVCPU < 1 {
		t.Errorf("EnterpriseDefaultVCPU = %d, must be >= 1", EnterpriseDefaultVCPU)
	}
	if EnterpriseMaxVCPU < EnterpriseDefaultVCPU {
		t.Errorf("EnterpriseMaxVCPU (%d) < EnterpriseDefaultVCPU (%d)", EnterpriseMaxVCPU, EnterpriseDefaultVCPU)
	}
	if EnterpriseDefaultTimeoutMs < 1000 {
		t.Errorf("EnterpriseDefaultTimeoutMs = %d, must be >= 1000", EnterpriseDefaultTimeoutMs)
	}
	if EnterpriseMaxTimeoutMs < EnterpriseDefaultTimeoutMs {
		t.Errorf("EnterpriseMaxTimeoutMs (%d) < EnterpriseDefaultTimeoutMs (%d)", EnterpriseMaxTimeoutMs, EnterpriseDefaultTimeoutMs)
	}
}

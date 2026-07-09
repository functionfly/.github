package plans

import "testing"

func TestGetMicroVMLimitsForPath_Marketplace(t *testing.T) {
	limits, err := GetMicroVMLimitsForPath(PlanEnterprise, "marketplace")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if limits == nil {
		t.Fatal("expected limits for enterprise + marketplace")
	}
	if limits.MaxMicroVMs != EnterpriseMaxMicroVMs {
		t.Errorf("marketplace should use base limits, got %v", limits.MaxMicroVMs)
	}
}

func TestGetMicroVMLimitsForPath_BYOAWS(t *testing.T) {
	limits, err := GetMicroVMLimitsForPath(PlanEnterprise, "byoaws")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if limits == nil {
		t.Fatal("expected limits for enterprise + byoaws")
	}
	// BYOAWS should have tighter caps
	if limits.MaxMicroVMs > 200 {
		t.Errorf("BYOAWS should cap at 200, got %v", limits.MaxMicroVMs)
	}
}

func TestGetMicroVMLimitsForPath_Hybrid(t *testing.T) {
	limits, err := GetMicroVMLimitsForPath(PlanEnterprise, "hybrid")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if limits == nil {
		t.Fatal("expected limits for enterprise + hybrid")
	}
	// Hybrid should have tightest caps
	if limits.MaxMicroVMs > 100 {
		t.Errorf("Hybrid should cap at 100, got %v", limits.MaxMicroVMs)
	}
	if limits.MaxConcurrentVMs > 50 {
		t.Errorf("Hybrid concurrent should cap at 50, got %v", limits.MaxConcurrentVMs)
	}
}

func TestGetMicroVMLimitsForPath_NonEnterprise(t *testing.T) {
	limits, err := GetMicroVMLimitsForPath(PlanFree, "marketplace")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if limits != nil {
		t.Errorf("expected nil for non-enterprise plan, got %v", limits)
	}
}

func TestGetMicroVMLimitsForPath_InvalidPath(t *testing.T) {
	_, err := GetMicroVMLimitsForPath(PlanEnterprise, "bogus")
	if err == nil {
		t.Error("expected error for invalid path")
	}
}

func TestGetMicroVMLimits_BackwardCompatible(t *testing.T) {
	// Old API should still work
	limits := GetMicroVMLimits(PlanEnterprise)
	if limits == nil {
		t.Fatal("expected limits for enterprise")
	}
	if limits.MaxMicroVMs != EnterpriseMaxMicroVMs {
		t.Errorf("expected %v, got %v", EnterpriseMaxMicroVMs, limits.MaxMicroVMs)
	}
}
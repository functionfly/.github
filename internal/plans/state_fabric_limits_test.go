package plans

import "testing"

func TestMaxStateFabricsPerPlan(t *testing.T) {
	tests := []struct {
		plan string
		want int
	}{
		{PlanFree, FreeMaxStateFabrics},
		{PlanStarter, StarterMaxStateFabrics},
		{PlanPro, ProMaxStateFabrics},
		{PlanEnterprise, EnterpriseMaxStateFabrics},
	}
	for _, tt := range tests {
		if got := MaxStateFabricsPerPlan(tt.plan); got != tt.want {
			t.Fatalf("MaxStateFabricsPerPlan(%q) = %d, want %d", tt.plan, got, tt.want)
		}
	}
}

func TestPlanHasStateFabricFeature(t *testing.T) {
	if !PlanHasStateFabricFeature(PlanStarter) {
		t.Fatal("starter should include state fabric")
	}
	if PlanHasStateFabricFeature(PlanFree) {
		t.Fatal("free should not include state fabric")
	}
}

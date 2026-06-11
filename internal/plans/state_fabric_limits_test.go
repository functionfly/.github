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

func TestAllSliceFeaturesHaveDefinitions(t *testing.T) {
	allSlices := [][]string{
		enterpriseFeatures,
		proFeatures,
		starterFeatures,
		freeFeatures,
		agentEnterpriseFeatures,
		agentProFeatures,
		agentScaleFeatures,
		agentStarterFeatures,
	}
	definedKeys := make(map[string]bool)
	for _, f := range featureDefinitions {
		definedKeys[f.Key] = true
	}
	for _, slice := range allSlices {
		for _, feature := range slice {
			if !definedKeys[feature] {
				t.Errorf("feature %q used in slice but not defined in featureDefinitions", feature)
			}
		}
	}
}

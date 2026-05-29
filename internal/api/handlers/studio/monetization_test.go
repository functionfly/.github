package studio

import (
	"testing"
)

func TestGeneratePricingRecommendationsHighVolume(t *testing.T) {
	recs := generatePricingRecommendations(MonetizationMetrics{
		ExecutionCount:   5000,
		AverageLatencyMs: 42,
		ErrorRate:        0.02,
		UserCount:        892,
	})

	if len(recs) < 3 {
		t.Fatalf("expected at least 3 recommendations, got %d", len(recs))
	}

	var perCall PricingRecommendation
	for _, rec := range recs {
		if rec.Model == "per_call" {
			perCall = rec
			break
		}
	}
	if perCall.Model == "" {
		t.Fatal("expected per_call recommendation")
	}
	if perCall.Confidence < 0.7 {
		t.Fatalf("expected high per_call confidence, got %.2f", perCall.Confidence)
	}
	if perCall.ExpectedRevenue <= 0 {
		t.Fatalf("expected positive per_call revenue projection, got %.2f", perCall.ExpectedRevenue)
	}
}

func TestGeneratePricingRecommendationsEarlyStage(t *testing.T) {
	recs := generatePricingRecommendations(MonetizationMetrics{
		ExecutionCount:   5,
		AverageLatencyMs: 80,
		ErrorRate:        0.01,
		UserCount:        3,
	})

	foundFree := false
	for _, rec := range recs {
		if rec.Model == "free" {
			foundFree = true
			break
		}
	}
	if !foundFree {
		t.Fatal("expected free recommendation for early-stage function")
	}
}

func TestValidPricingModel(t *testing.T) {
	for _, model := range []string{"free", "per_call", "subscription", "revenue_share"} {
		if !validPricingModel(model) {
			t.Fatalf("expected %q to be valid", model)
		}
	}
	if validPricingModel("auction") {
		t.Fatal("expected auction to be invalid")
	}
}

func TestRoundPrice(t *testing.T) {
	if roundPrice(0.010004) != 0.01 {
		t.Fatalf("unexpected round: %v", roundPrice(0.010004))
	}
}

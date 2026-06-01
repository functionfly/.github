package studio

import (
	"crypto/sha256"
	"encoding/hex"
	"fmt"
	"strings"

	"github.com/google/uuid"
)

type MonetizationMetrics struct {
	ExecutionCount   int
	AverageLatencyMs float64
	ErrorRate        float64
	UserCount        int
}

type PricingRecommendation struct {
	Model            string
	Confidence       float64
	ExpectedRevenue float64
}

var validSPDXLicenses = map[string]bool{
	"mit":        true,
	"apache":     true,
	"gpl":        true,
	"proprietary": true,
	"custom":     true,
}

var validCommercialTypes = map[string]bool{
	"open":        true,
	"restricted":  true,
	"commercial":  true,
}

var validPricingModels = map[string]bool{
	"free":          true,
	"per_call":      true,
	"subscription":  true,
	"revenue_share": true,
}

func hashLicenseKey(key string) string {
	h := sha256.Sum256([]byte(key))
	return hex.EncodeToString(h[:])
}

func generateLicenseKey() (key, hash, prefix string, err error) {
	id := uuid.New().String()
	hash = hashLicenseKey(id)
	key = fmt.Sprintf("FFLIC-%s", strings.ToUpper(uuid.New().String()[:12]))
	prefix = key[:12]
	return key, hash, prefix, nil
}

func validSPDXLicense(license string) bool {
	return validSPDXLicenses[license]
}

func validCommercialType(typ string) bool {
	return validCommercialTypes[typ]
}

func maskLicenseKey(key string) string {
	if len(key) <= 8 {
		return key + "****"
	}
	return key + "****"
}

func generatePricingRecommendations(metrics MonetizationMetrics) []PricingRecommendation {
	var recs []PricingRecommendation

	if metrics.ExecutionCount > 1000 {
		recs = append(recs, PricingRecommendation{
			Model:            "per_call",
			Confidence:       0.85,
			ExpectedRevenue:  1000.0,
		})
	}

	if metrics.ExecutionCount < 50 {
		recs = append(recs, PricingRecommendation{
			Model:      "free",
			Confidence: 0.9,
		})
	}

	recs = append(recs, PricingRecommendation{
		Model:      "subscription",
		Confidence: 0.7,
	})

	if metrics.ExecutionCount >= 5000 {
		recs = append(recs, PricingRecommendation{
			Model:      "revenue_share",
			Confidence: 0.75,
		})
	}

	return recs
}

func validPricingModel(model string) bool {
	return validPricingModels[model]
}

func roundPrice(v float64) float64 {
	return float64(int(v*100)) / 100
}

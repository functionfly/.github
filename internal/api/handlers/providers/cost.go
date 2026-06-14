package providers

import (
	"encoding/json"
	"net/http"

	"github.com/functionfly/functionfly/internal/api/middleware"
)

// HandleEstimateCost provides cost estimation for function deployment
func (h *Handler) HandleEstimateCost(w http.ResponseWriter, r *http.Request) {
	var req CostEstimationRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, "Invalid request body", http.StatusBadRequest)
		return
	}

	claims := middleware.GetUserFromContext(r)
	if claims == nil {
		http.Error(w, "Unauthorized", http.StatusUnauthorized)
		return
	}

	provider, _ := h.repo.GetProviderByUserAndType(r.Context(), claims.UserID, req.Provider)

	if provider.Status != "active" {
		http.Error(w, "Provider not active", http.StatusBadRequest)
		return
	}

	estimation := h.estimateCost(req)

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(estimation)
}

func (h *Handler) estimateCost(req CostEstimationRequest) CostEstimationResponse {
	baseCosts := map[string]float64{
		"cloudflare": 0.0,
		"vercel":     0.0,
		"fly":        2.67,
		"aws-lambda": 0.0,
	}

	computeCosts := map[string]float64{
		"cloudflare": 0.30,
		"vercel":     0.40,
		"fly":        0.22,
		"aws-lambda": 0.20,
	}

	storageCosts := map[string]float64{
		"cloudflare": 0.055,
		"vercel":     0.10,
		"fly":        0.15,
		"aws-lambda": 0.03,
	}

	requestsPerMonth := float64(req.RequestsPerDay) * 30
	computeCost := (requestsPerMonth / 1000000) * computeCosts[req.Provider]
	storageCost := 0.01 * storageCosts[req.Provider]
	bandwidthMB := (requestsPerMonth * 1024) / (1024 * 1024)
	bandwidthCost := bandwidthMB * 0.09

	totalCost := baseCosts[req.Provider] + computeCost + storageCost + bandwidthCost

	breakdown := map[string]float64{
		"base":      baseCosts[req.Provider],
		"compute":   computeCost,
		"storage":   storageCost,
		"bandwidth": bandwidthCost,
	}

	return CostEstimationResponse{
		MonthlyCost: totalCost,
		Currency:    "USD",
		Breakdown:   breakdown,
		ProviderData: map[string]interface{}{
			"requests_per_month":     requestsPerMonth,
			"estimated_bandwidth_gb": bandwidthMB,
			"regions_count":          len(req.Regions),
		},
	}
}

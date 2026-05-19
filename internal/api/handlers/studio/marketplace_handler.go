package studio

import (
	"encoding/json"
	"net/http"
	"strconv"

	"github.com/gorilla/mux"
	"github.com/sirupsen/logrus"
)

// MarketplaceHandler handles studio marketplace HTTP requests
type MarketplaceHandler struct {
	repo *MarketplaceRepository
}

// NewMarketplaceHandler creates a new marketplace handler
func NewMarketplaceHandler(repo *MarketplaceRepository) *MarketplaceHandler {
	return &MarketplaceHandler{repo: repo}
}

// HandleListFunctions handles GET /v1/marketplace/functions
func (h *MarketplaceHandler) HandleListFunctions(w http.ResponseWriter, r *http.Request) {
	tenantID := getTenantID(r)
	if tenantID == "" {
		writeJSONError(w, http.StatusUnauthorized, "Unauthorized")
		return
	}

	search := r.URL.Query().Get("q")
	category := r.URL.Query().Get("category")
	limit := 50
	if l := r.URL.Query().Get("limit"); l != "" {
		if n, err := strconv.Atoi(l); err == nil && n > 0 && n <= 200 {
			limit = n
		}
	}
	offset := 0
	if o := r.URL.Query().Get("offset"); o != "" {
		if n, err := strconv.Atoi(o); err == nil && n >= 0 {
			offset = n
		}
	}

	var searchPtr, categoryPtr *string
	if search != "" {
		searchPtr = &search
	}
	if category != "" {
		categoryPtr = &category
	}

	params := ListFunctionsParams{
		TenantID: tenantID,
		Search:   searchPtr,
		Category: categoryPtr,
		Limit:    limit,
		Offset:   offset,
	}

	functions, err := h.repo.ListFunctions(r.Context(), params)
	if err != nil {
		logrus.WithError(err).Warn("marketplace: failed to list functions")
		writeJSON(w, http.StatusOK, map[string]interface{}{"functions": []MarketplaceFunction{}})
		return
	}

	writeJSON(w, http.StatusOK, map[string]interface{}{"functions": functions})
}

// HandleExecuteFunction handles POST /v1/marketplace/functions/{id}/execute
func (h *MarketplaceHandler) HandleExecuteFunction(w http.ResponseWriter, r *http.Request) {
	tenantID := getTenantID(r)
	if tenantID == "" {
		writeJSONError(w, http.StatusUnauthorized, "Unauthorized")
		return
	}

	functionID := mux.Vars(r)["id"]
	if functionID == "" {
		writeJSONError(w, http.StatusBadRequest, "id is required")
		return
	}

	var req struct {
		Input map[string]interface{} `json:"input"`
	}
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		req.Input = make(map[string]interface{})
	}

	executionID, err := h.repo.ExecuteFunction(r.Context(), functionID, req.Input)
	if err != nil {
		logrus.WithError(err).Error("marketplace: failed to execute function")
		writeJSONError(w, http.StatusInternalServerError, "Failed to execute function")
		return
	}

	writeJSON(w, http.StatusOK, map[string]interface{}{
		"executionId": executionID,
		"status":     "started",
	})
}

// HandleFavoriteFunction handles POST/DELETE /v1/marketplace/functions/{id}/favorite
func (h *MarketplaceHandler) HandleFavoriteFunction(w http.ResponseWriter, r *http.Request) {
	tenantID := getTenantID(r)
	if tenantID == "" {
		writeJSONError(w, http.StatusUnauthorized, "Unauthorized")
		return
	}

	functionID := mux.Vars(r)["id"]
	if functionID == "" {
		writeJSONError(w, http.StatusBadRequest, "id is required")
		return
	}

	favorite := r.Method == http.MethodPost

	if err := h.repo.SetFavorite(r.Context(), tenantID, functionID, favorite); err != nil {
		logrus.WithError(err).Error("marketplace: failed to set favorite")
		writeJSONError(w, http.StatusInternalServerError, "Failed to update favorite")
		return
	}

	writeJSON(w, http.StatusOK, map[string]string{"message": "Favorite updated"})
}

// HandleListPlans handles GET /v1/marketplace/plans
func (h *MarketplaceHandler) HandleListPlans(w http.ResponseWriter, r *http.Request) {
	tenantID := getTenantID(r)
	if tenantID == "" {
		writeJSONError(w, http.StatusUnauthorized, "Unauthorized")
		return
	}

	plans, err := h.repo.ListPlans(r.Context(), tenantID)
	if err != nil {
		logrus.WithError(err).Warn("marketplace: failed to list plans")
		writeJSON(w, http.StatusOK, map[string]interface{}{"plans": []SubscriptionPlan{}})
		return
	}

	writeJSON(w, http.StatusOK, map[string]interface{}{"plans": plans})
}

// HandleCreatePlan handles POST /v1/marketplace/plans
func (h *MarketplaceHandler) HandleCreatePlan(w http.ResponseWriter, r *http.Request) {
	tenantID := getTenantID(r)
	if tenantID == "" {
		writeJSONError(w, http.StatusUnauthorized, "Unauthorized")
		return
	}

	var req struct {
		Name      string   `json:"name"`
		Price     float64  `json:"price"`
		Features  []string `json:"features"`
	}
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		writeJSONError(w, http.StatusBadRequest, "Invalid request body")
		return
	}

	if req.Name == "" {
		writeJSONError(w, http.StatusBadRequest, "name is required")
		return
	}

	plan := &SubscriptionPlan{
		TenantID: tenantID,
		Name:     req.Name,
		Price:    req.Price,
		Features: req.Features,
	}

	if err := h.repo.CreatePlan(r.Context(), plan); err != nil {
		logrus.WithError(err).Error("marketplace: failed to create plan")
		writeJSONError(w, http.StatusInternalServerError, "Failed to create plan")
		return
	}

	writeJSON(w, http.StatusCreated, map[string]interface{}{"plan": plan})
}

// HandleUpdatePlan handles PUT /v1/marketplace/plans/{id}
func (h *MarketplaceHandler) HandleUpdatePlan(w http.ResponseWriter, r *http.Request) {
	tenantID := getTenantID(r)
	if tenantID == "" {
		writeJSONError(w, http.StatusUnauthorized, "Unauthorized")
		return
	}

	planID := mux.Vars(r)["id"]
	if planID == "" {
		writeJSONError(w, http.StatusBadRequest, "id is required")
		return
	}

	var req struct {
		Name      string   `json:"name"`
		Price     float64  `json:"price"`
		Features  []string `json:"features"`
	}
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		writeJSONError(w, http.StatusBadRequest, "Invalid request body")
		return
	}

	plan := &SubscriptionPlan{
		ID:     planID,
		TenantID: tenantID,
		Name:   req.Name,
		Price:  req.Price,
		Features: req.Features,
	}

	if err := h.repo.UpdatePlan(r.Context(), plan); err != nil {
		logrus.WithError(err).Error("marketplace: failed to update plan")
		writeJSONError(w, http.StatusInternalServerError, "Failed to update plan")
		return
	}

	writeJSON(w, http.StatusOK, map[string]interface{}{"plan": plan})
}

// HandleListRoyalties handles GET /v1/marketplace/royalties
func (h *MarketplaceHandler) HandleListRoyalties(w http.ResponseWriter, r *http.Request) {
	tenantID := getTenantID(r)
	if tenantID == "" {
		writeJSONError(w, http.StatusUnauthorized, "Unauthorized")
		return
	}

	royalties, totalEarnings, pendingPayout, err := h.repo.ListRoyalties(r.Context(), tenantID)
	if err != nil {
		logrus.WithError(err).Warn("marketplace: failed to list royalties")
		writeJSON(w, http.StatusOK, map[string]interface{}{
			"royalties": []RoyaltyEntry{},
			"totalEarnings": 0,
			"pendingPayout": 0,
		})
		return
	}

	writeJSON(w, http.StatusOK, map[string]interface{}{
		"royalties":       royalties,
		"totalEarnings":   totalEarnings,
		"pendingPayout":   pendingPayout,
	})
}

// HandleRequestPayout handles POST /v1/marketplace/royalties/payout
func (h *MarketplaceHandler) HandleRequestPayout(w http.ResponseWriter, r *http.Request) {
	tenantID := getTenantID(r)
	if tenantID == "" {
		writeJSONError(w, http.StatusUnauthorized, "Unauthorized")
		return
	}

	if err := h.repo.RequestPayout(r.Context(), tenantID); err != nil {
		logrus.WithError(err).Error("marketplace: failed to request payout")
		writeJSONError(w, http.StatusInternalServerError, "Failed to request payout")
		return
	}

	writeJSON(w, http.StatusOK, map[string]string{"message": "Payout requested"})
}

// HandleUpdateLicense handles PUT /v1/marketplace/functions/{id}/license
func (h *MarketplaceHandler) HandleUpdateLicense(w http.ResponseWriter, r *http.Request) {
	tenantID := getTenantID(r)
	if tenantID == "" {
		writeJSONError(w, http.StatusUnauthorized, "Unauthorized")
		return
	}

	functionID := mux.Vars(r)["id"]
	if functionID == "" {
		writeJSONError(w, http.StatusBadRequest, "id is required")
		return
	}

	var req struct {
		License string `json:"license"`
	}
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		writeJSONError(w, http.StatusBadRequest, "Invalid request body")
		return
	}

	if req.License == "" {
		writeJSONError(w, http.StatusBadRequest, "license is required")
		return
	}

	if err := h.repo.UpdateLicense(r.Context(), tenantID, functionID, req.License); err != nil {
		logrus.WithError(err).Error("marketplace: failed to update license")
		writeJSONError(w, http.StatusInternalServerError, "Failed to update license")
		return
	}

	writeJSON(w, http.StatusOK, map[string]string{"message": "License updated"})
}

// HandleUpdatePricing handles PUT /v1/marketplace/functions/{id}/pricing
func (h *MarketplaceHandler) HandleUpdatePricing(w http.ResponseWriter, r *http.Request) {
	tenantID := getTenantID(r)
	if tenantID == "" {
		writeJSONError(w, http.StatusUnauthorized, "Unauthorized")
		return
	}

	functionID := mux.Vars(r)["id"]
	if functionID == "" {
		writeJSONError(w, http.StatusBadRequest, "id is required")
		return
	}

	var req struct {
		Price float64 `json:"price"`
		Model string  `json:"pricing_model"`
	}
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		writeJSONError(w, http.StatusBadRequest, "Invalid request body")
		return
	}

	if req.Price < 0 {
		writeJSONError(w, http.StatusBadRequest, "price must be non-negative")
		return
	}

	if req.Model == "" {
		req.Model = "per_call"
	}

	if err := h.repo.UpdatePricing(r.Context(), tenantID, functionID, req.Price, req.Model); err != nil {
		logrus.WithError(err).Error("marketplace: failed to update pricing")
		writeJSONError(w, http.StatusInternalServerError, "Failed to update pricing")
		return
	}

	writeJSON(w, http.StatusOK, map[string]string{"message": "Pricing updated"})
}
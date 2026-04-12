package billing

import (
	"encoding/json"
	"net/http"
	"os"
	"time"

	"github.com/functionfly/functionfly/internal/api/middleware"
	"github.com/functionfly/functionfly/internal/payment"
	"github.com/functionfly/functionfly/internal/storage"
	"github.com/google/uuid"
	"github.com/sirupsen/logrus"
)

// HandleGetBundles returns all active Backend-in-a-Box pricing bundles
// GET /v1/billing/bundles
func (h *Handler) HandleGetBundles(w http.ResponseWriter, r *http.Request) {
	claims := middleware.GetUserFromContext(r)
	if claims == nil {
		writeJSONError(w, http.StatusUnauthorized, "Unauthorized")
		return
	}

	// Get bundles from database
	bundles, err := h.repo.ListPricingBundles(r.Context(), true)
	if err != nil {
		logrus.WithError(err).Error("billing bundles: failed to list bundles")
		writeJSONError(w, http.StatusInternalServerError, "Failed to retrieve bundles")
		return
	}

	response := make([]BundleResponse, 0, len(bundles))
	for _, b := range bundles {
		response = append(response, bundleToResponse(b))
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(map[string]interface{}{
		"bundles": response,
	})
}

// HandleGetBundle returns details for a specific bundle
// GET /v1/billing/bundles/:slug
func (h *Handler) HandleGetBundle(w http.ResponseWriter, r *http.Request) {
	claims := middleware.GetUserFromContext(r)
	if claims == nil {
		writeJSONError(w, http.StatusUnauthorized, "Unauthorized")
		return
	}

	slug := r.PathValue("slug")
	if slug == "" {
		writeJSONError(w, http.StatusBadRequest, "Bundle slug is required")
		return
	}

	bundle, err := h.repo.GetPricingBundleBySlug(r.Context(), slug)
	if err != nil {
		logrus.WithError(err).WithField("slug", slug).Error("billing bundles: failed to get bundle")
		writeJSONError(w, http.StatusInternalServerError, "Failed to retrieve bundle")
		return
	}

	if bundle == nil {
		writeJSONError(w, http.StatusNotFound, "Bundle not found")
		return
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(bundleToResponse(bundle))
}

// HandleRegisterFounderMode registers a user for founder mode (free until trigger)
// POST /v1/billing/bundles/:slug/founder
func (h *Handler) HandleRegisterFounderMode(w http.ResponseWriter, r *http.Request) {
	claims := middleware.GetUserFromContext(r)
	if claims == nil || claims.TenantID == uuid.Nil {
		writeJSONError(w, http.StatusUnauthorized, "Unauthorized")
		return
	}

	slug := r.PathValue("slug")
	if slug == "" {
		writeJSONError(w, http.StatusBadRequest, "Bundle slug is required")
		return
	}

	var req RegisterFounderModeRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		writeJSONError(w, http.StatusBadRequest, "Invalid request body")
		return
	}

	// Get the bundle
	bundle, err := h.repo.GetPricingBundleBySlug(r.Context(), slug)
	if err != nil {
		logrus.WithError(err).WithField("slug", slug).Error("billing founder mode: failed to get bundle")
		writeJSONError(w, http.StatusInternalServerError, "Failed to retrieve bundle")
		return
	}
	if bundle == nil {
		writeJSONError(w, http.StatusNotFound, "Bundle not found")
		return
	}

	// Set defaults
	modeType := req.ModeType
	if modeType == "" {
		modeType = "hybrid" // Default: time OR revenue
	}
	freeDays := req.FreeDays
	if freeDays <= 0 {
		freeDays = 90 // Default: 3 months
	}
	mrrThreshold := req.MRRThreshold
	if mrrThreshold <= 0 {
		mrrThreshold = 100000 // Default: $1000 MRR
	}

	// Check for existing active founder mode
	existing, err := h.repo.GetActiveFounderMode(r.Context(), claims.TenantID, bundle.ID)
	if err != nil {
		logrus.WithError(err).Error("billing founder mode: failed to check existing registration")
		writeJSONError(w, http.StatusInternalServerError, "Failed to check founder mode status")
		return
	}
	if existing != nil {
		writeJSONError(w, http.StatusConflict, "Already registered for founder mode with this bundle")
		return
	}

	// Calculate end date for time-based mode
	var endsAt *time.Time
	if modeType == "time_based" || modeType == "hybrid" {
		e := time.Now().UTC().AddDate(0, 0, freeDays)
		endsAt = &e
	}

	// Register founder mode
	registration := &storage.FounderModeRegistration{
		ID:                uuid.New(),
		TenantID:          claims.TenantID,
		BundleID:          bundle.ID,
		ModeType:          modeType,
		Status:            "active",
		StartedAt:         time.Now().UTC(),
		EndsAt:            endsAt,
		FreeDays:          freeDays,
		MRRThresholdCents: mrrThreshold,
	}

	if err := h.repo.CreateFounderModeRegistration(r.Context(), registration); err != nil {
		logrus.WithError(err).Error("billing founder mode: failed to create registration")
		writeJSONError(w, http.StatusInternalServerError, "Failed to register founder mode")
		return
	}

	// Create initial bundle subscription (deferred billing)
	sub := &storage.BundleSubscription{
		ID:                 uuid.New(),
		TenantID:           claims.TenantID,
		BundleID:           bundle.ID,
		FounderModeID:      &registration.ID,
		Status:             "deferred", // Special status for founder mode
		CurrentPeriodStart: time.Now().UTC(),
		CurrentPeriodEnd:   time.Now().UTC().AddDate(0, 0, freeDays),
	}

	if err := h.repo.CreateBundleSubscription(r.Context(), sub); err != nil {
		logrus.WithError(err).Error("billing founder mode: failed to create deferred subscription")
		writeJSONError(w, http.StatusInternalServerError, "Failed to create deferred subscription")
		return
	}

	// Trigger auto-provisioning
	go h.provisionBundleResources(claims.TenantID, bundle)

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusCreated)
	json.NewEncoder(w).Encode(founderModeToResponse(registration, bundle.Slug))
}

// HandleCreateBundleCheckout creates a Stripe checkout for immediate bundle subscription
// POST /v1/billing/bundles/:slug/checkout
func (h *Handler) HandleCreateBundleCheckout(w http.ResponseWriter, r *http.Request) {
	claims := middleware.GetUserFromContext(r)
	if claims == nil || claims.TenantID == uuid.Nil {
		writeJSONError(w, http.StatusUnauthorized, "Unauthorized")
		return
	}

	if !payment.IsConfigured() {
		writeJSONError(w, http.StatusServiceUnavailable, "Billing is not configured")
		return
	}

	slug := r.PathValue("slug")
	if slug == "" {
		writeJSONError(w, http.StatusBadRequest, "Bundle slug is required")
		return
	}

	// Get the bundle
	bundle, err := h.repo.GetPricingBundleBySlug(r.Context(), slug)
	if err != nil {
		logrus.WithError(err).WithField("slug", slug).Error("billing bundle checkout: failed to get bundle")
		writeJSONError(w, http.StatusInternalServerError, "Failed to retrieve bundle")
		return
	}
	if bundle == nil {
		writeJSONError(w, http.StatusNotFound, "Bundle not found")
		return
	}

	if bundle.StripePriceID == "" {
		writeJSONError(w, http.StatusServiceUnavailable, "Bundle not configured for checkout")
		return
	}

	var req CreateBundleCheckoutRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		writeJSONError(w, http.StatusBadRequest, "Invalid request body")
		return
	}

	user, err := h.repo.GetUserByID(claims.UserID)
	if err != nil || user == nil {
		logrus.WithError(err).WithField("user_id", claims.UserID).Warn("billing bundle checkout: user not found")
		writeJSONError(w, http.StatusNotFound, "User not found")
		return
	}

	name := user.Name
	if name == "" && user.ProviderData != nil {
		if n, ok := user.ProviderData["name"].(string); ok {
			name = n
		}
	}
	if name == "" {
		name = user.Email
	}

	// Create checkout session with bundle metadata
	resp, err := payment.CreateCheckoutSession(
		r.Context(),
		h.repo,
		claims.TenantID,
		user.Email,
		name,
		payment.CreateCheckoutSessionRequest{
			PriceID:    bundle.StripePriceID,
			SuccessURL: req.SuccessURL,
			CancelURL:  req.CancelURL,
		},
	)
	if err != nil {
		logrus.WithError(err).WithField("tenant_id", claims.TenantID).Error("billing bundle checkout: failed to create session")
		msg := "Failed to create checkout session"
		if os.Getenv("DEVELOPMENT") == "true" {
			msg = err.Error()
		}
		writeJSONError(w, http.StatusInternalServerError, msg)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(resp)
}

// HandleGetFounderModeStatus returns the current user's founder mode status
// GET /v1/billing/founder-mode
func (h *Handler) HandleGetFounderModeStatus(w http.ResponseWriter, r *http.Request) {
	claims := middleware.GetUserFromContext(r)
	if claims == nil || claims.TenantID == uuid.Nil {
		writeJSONError(w, http.StatusUnauthorized, "Unauthorized")
		return
	}

	registrations, err := h.repo.ListFounderModesByTenant(r.Context(), claims.TenantID)
	if err != nil {
		logrus.WithError(err).Error("billing founder mode: failed to list registrations")
		writeJSONError(w, http.StatusInternalServerError, "Failed to retrieve founder mode status")
		return
	}

	response := make([]FounderModeResponse, 0, len(registrations))
	for _, reg := range registrations {
		bundle, _ := h.repo.GetPricingBundleByID(r.Context(), reg.BundleID)
		slug := ""
		if bundle != nil {
			slug = bundle.Slug
		}
		response = append(response, founderModeToResponse(reg, slug))
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(map[string]interface{}{
		"founder_modes": response,
	})
}

// HandleGetDeferredBillingStatus returns progress toward deferred billing triggers
// GET /v1/billing/deferred-status
func (h *Handler) HandleGetDeferredBillingStatus(w http.ResponseWriter, r *http.Request) {
	claims := middleware.GetUserFromContext(r)
	if claims == nil || claims.TenantID == uuid.Nil {
		writeJSONError(w, http.StatusUnauthorized, "Unauthorized")
		return
	}

	// Get active founder mode
	registrations, err := h.repo.ListActiveFounderModesByTenant(r.Context(), claims.TenantID)
	if err != nil {
		logrus.WithError(err).Error("billing deferred: failed to list founder modes")
		writeJSONError(w, http.StatusInternalServerError, "Failed to retrieve deferred billing status")
		return
	}

	if len(registrations) == 0 {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusOK)
		json.NewEncoder(w).Encode(map[string]interface{}{
			"deferred_status": nil,
			"message":         "No active deferred billing",
		})
		return
	}

	// For now, return status for the first active registration
	// In practice, users typically have one bundle at a time
	reg := registrations[0]

	// Get usage metrics for this tenant
	now := time.Now().UTC()
	periodStart := time.Date(now.Year(), now.Month(), 1, 0, 0, 0, 0, time.UTC)
	periodEnd := periodStart.AddDate(0, 1, 0).Add(-time.Second)

	execRollups, _ := h.repo.GetUsageByTenant(claims.TenantID, "function_execution", periodStart, periodEnd)
	totalExecutions := 0
	for _, rollup := range execRollups {
		totalExecutions += rollup.TotalQuantity
	}

	// Get actual user count for the tenant
	userCount, err := h.repo.CountUsersByTenant(r.Context(), claims.TenantID)
	if err != nil {
		logrus.WithError(err).WithField("tenant_id", claims.TenantID).Warn("billing deferred status: failed to count users")
		userCount = 0 // Default to 0 on error
	}

	// Build trigger thresholds and current progress
	thresholds := map[string]interface{}{
		"user_count": 100,
		"mrr_cents":  100000, // $1000
		"days":       90,
	}
	if reg.ModeType == "revenue_based" {
		thresholds["mrr_cents"] = reg.MRRThresholdCents
	}
	if reg.EndsAt != nil {
		daysUntilExpiry := int(reg.EndsAt.Sub(now).Hours() / 24)
		thresholds["days"] = daysUntilExpiry
	}

	progress := map[string]interface{}{
		"user_count":   userCount,
		"mrr_cents":    reg.MaxMRRSeenCents,
		"api_calls":    totalExecutions,
		"days_elapsed": int(now.Sub(reg.StartedAt).Hours() / 24),
	}

	// Calculate overall progress toward first trigger
	progressPercent := 0.0
	if userCount >= thresholds["user_count"].(int) {
		progressPercent = 100.0
	} else if reg.MaxMRRSeenCents >= thresholds["mrr_cents"].(int) {
		progressPercent = 100.0
	} else {
		// Use days elapsed as proxy for now
		daysElapsed := progress["days_elapsed"].(int)
		totalDays := thresholds["days"].(int)
		progressPercent = float64(daysElapsed) / float64(totalDays) * 100
		if progressPercent > 100 {
			progressPercent = 100
		}
	}

	status := "building"
	if progressPercent >= 80 && progressPercent < 100 {
		status = "approaching"
	} else if reg.Status == "grace_period" {
		status = "grace_period"
	} else if progressPercent >= 100 {
		status = "triggered"
	}

	estimatedDaysLeft := 90 - progress["days_elapsed"].(int)
	if estimatedDaysLeft < 0 {
		estimatedDaysLeft = 0
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(DeferredBillingStatus{
		BundleID:          reg.BundleID,
		Status:            status,
		TriggerThresholds: thresholds,
		CurrentProgress:   progress,
		ProgressPercent:   progressPercent,
		EstimatedDaysLeft: &estimatedDaysLeft,
	})
}

// HandleConvertToPaid manually converts a founder mode registration to paid subscription
// POST /v1/billing/convert-to-paid
func (h *Handler) HandleConvertToPaid(w http.ResponseWriter, r *http.Request) {
	claims := middleware.GetUserFromContext(r)
	if claims == nil || claims.TenantID == uuid.Nil {
		writeJSONError(w, http.StatusUnauthorized, "Unauthorized")
		return
	}

	if !payment.IsConfigured() {
		writeJSONError(w, http.StatusServiceUnavailable, "Billing is not configured")
		return
	}

	var req ConvertToPaidRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		writeJSONError(w, http.StatusBadRequest, "Invalid request body")
		return
	}

	bundleID, err := uuid.Parse(req.BundleID)
	if err != nil {
		writeJSONError(w, http.StatusBadRequest, "Invalid bundle_id")
		return
	}

	// Get the founder mode registration
	reg, err := h.repo.GetActiveFounderMode(r.Context(), claims.TenantID, bundleID)
	if err != nil {
		logrus.WithError(err).Error("billing convert: failed to get founder mode")
		writeJSONError(w, http.StatusInternalServerError, "Failed to retrieve founder mode")
		return
	}
	if reg == nil {
		writeJSONError(w, http.StatusNotFound, "No active founder mode found for this bundle")
		return
	}

	// Get bundle for Stripe price ID
	bundle, err := h.repo.GetPricingBundleByID(r.Context(), bundleID)
	if err != nil {
		logrus.WithError(err).Error("billing convert: failed to get bundle")
		writeJSONError(w, http.StatusInternalServerError, "Failed to retrieve bundle")
		return
	}

	if bundle.StripePriceID == "" {
		writeJSONError(w, http.StatusServiceUnavailable, "Bundle not configured for checkout")
		return
	}

	// Get user for checkout
	user, err := h.repo.GetUserByID(claims.UserID)
	if err != nil || user == nil {
		logrus.WithError(err).WithField("user_id", claims.UserID).Warn("billing convert: user not found")
		writeJSONError(w, http.StatusNotFound, "User not found")
		return
	}

	name := user.Name
	if name == "" && user.ProviderData != nil {
		if n, ok := user.ProviderData["name"].(string); ok {
			name = n
		}
	}
	if name == "" {
		name = user.Email
	}

	// Create checkout session for conversion
	resp, err := payment.CreateCheckoutSession(
		r.Context(),
		h.repo,
		claims.TenantID,
		user.Email,
		name,
		payment.CreateCheckoutSessionRequest{
			PriceID:    bundle.StripePriceID,
			SuccessURL: "/dashboard?converted=true",
			CancelURL:  "/pricing",
		},
	)
	if err != nil {
		logrus.WithError(err).WithField("tenant_id", claims.TenantID).Error("billing convert: failed to create checkout")
		writeJSONError(w, http.StatusInternalServerError, "Failed to create checkout session")
		return
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(resp)
}

// bundleToResponse converts a PricingBundle to BundleResponse
func bundleToResponse(b *storage.PricingBundle) BundleResponse {
	return BundleResponse{
		ID:                b.ID,
		Slug:              b.Slug,
		Name:              b.Name,
		DisplayName:       b.DisplayName,
		Description:       b.Description,
		ShortDescription:  b.ShortDescription,
		PriceCents:        b.DisplayPriceCents,
		PriceUSD:          formatCents(b.DisplayPriceCents),
		BillingInterval:   b.BillingInterval,
		Icon:              b.Icon,
		Color:             b.Color,
		FeaturesIncluded:  b.FeaturesIncluded,
		FeatureLimits:     b.FeatureLimits,
		ProvisioningSteps: b.ProvisioningTemplates,
		IsPopular:         b.IsPopular,
		SortOrder:         b.SortOrder,
	}
}

// founderModeToResponse converts a FounderModeRegistration to FounderModeResponse
func founderModeToResponse(r *storage.FounderModeRegistration, bundleSlug string) FounderModeResponse {
	daysRemaining := 0
	if r.EndsAt != nil {
		daysRemaining = int(r.EndsAt.Sub(time.Now().UTC()).Hours() / 24)
		if daysRemaining < 0 {
			daysRemaining = 0
		}
	}

	return FounderModeResponse{
		ID:                r.ID,
		BundleID:          r.BundleID,
		BundleSlug:        bundleSlug,
		ModeType:          r.ModeType,
		Status:            r.Status,
		StartedAt:         r.StartedAt,
		EndsAt:            r.EndsAt,
		FreeDays:          r.FreeDays,
		MRRThresholdCents: r.MRRThresholdCents,
		DaysRemaining:     daysRemaining,
	}
}

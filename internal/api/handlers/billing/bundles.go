package billing

import (
	"encoding/json"
	"fmt"
	"net/http"
	"os"
	"time"

	"github.com/functionfly/functionfly/internal/api/middleware"
	"github.com/functionfly/functionfly/internal/payment"
	"github.com/functionfly/functionfly/internal/storage"
	"github.com/google/uuid"
	"github.com/gorilla/mux"
	"github.com/sirupsen/logrus"
)

func encodeJSON(w http.ResponseWriter, status int, data interface{}) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	if err := json.NewEncoder(w).Encode(data); err != nil {
		logrus.WithError(err).Error("billing bundles: failed to encode response")
	}
}

// HandleGetBundles returns all active Backend-in-a-Box pricing bundles
// GET /v1/billing/bundles
// NOTE: This endpoint is public — no auth required so pricing is visible to unauthenticated users
func (h *Handler) HandleGetBundles(w http.ResponseWriter, r *http.Request) {
	// Get bundles from database (public endpoint - no auth required)
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

	encodeJSON(w, http.StatusOK, map[string]interface{}{
		"bundles": response,
	})
}

// HandleGetBundleStats returns public stats about bundles (founder count, deployments)
// GET /v1/billing/bundles/stats
func (h *Handler) HandleGetBundleStats(w http.ResponseWriter, r *http.Request) {
	founderCount, err := h.repo.CountActiveFounderModeRegistrations(r.Context())
	if err != nil {
		logrus.WithError(err).Warn("billing bundles: failed to count founder mode registrations")
		founderCount = 0
	}

	deploymentsCount, err := h.repo.CountRecentSuccessfulDeployments(r.Context())
	if err != nil {
		logrus.WithError(err).Warn("billing bundles: failed to count recent deployments")
		deploymentsCount = 0
	}

	encodeJSON(w, http.StatusOK, map[string]interface{}{
		"active_founders":    founderCount,
		"recent_deployments": deploymentsCount,
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

	slug := mux.Vars(r)["slug"]
	if slug == "" {
		writeJSONError(w, http.StatusBadRequest, "Bundle slug is required")
		return
	}

	// Validate slug format to prevent enumeration and injection attacks
	// Only allow known bundle slugs
	validSlugs := map[string]bool{"saas-starter": true, "marketplace": true, "ai-app": true}
	if !validSlugs[slug] {
		// Return 404 for unknown slugs to prevent enumeration
		writeJSONError(w, http.StatusNotFound, "Bundle not found")
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

	encodeJSON(w, http.StatusOK, bundleToResponse(bundle))
}

// HandleGetBundleUsageStatus returns current usage against bundle limits
// GET /v1/billing/bundle/usage
func (h *Handler) HandleGetBundleUsageStatus(w http.ResponseWriter, r *http.Request) {
	claims := middleware.GetUserFromContext(r)
	if claims == nil || claims.TenantID == uuid.Nil {
		writeJSONError(w, http.StatusUnauthorized, "Unauthorized")
		return
	}

	ctx := r.Context()
	now := time.Now().UTC()
	periodStart := time.Date(now.Year(), now.Month(), 1, 0, 0, 0, 0, time.UTC)
	periodEnd := periodStart.AddDate(0, 1, 0).Add(-time.Second)

	sub, err := h.repo.GetBundleSubscriptionByTenant(ctx, claims.TenantID)
	if err != nil || sub == nil {
		writeJSONError(w, http.StatusNotFound, "No bundle subscription found")
		return
	}

	bundle, err := h.repo.GetPricingBundleByID(ctx, sub.BundleID)
	if err != nil || bundle == nil {
		writeJSONError(w, http.StatusInternalServerError, "Failed to retrieve bundle")
		return
	}

	execRollups, _ := h.repo.GetUsageByTenant(ctx, claims.TenantID, "function_execution", periodStart, periodEnd)
	totalExecutions := 0
	for _, rollup := range execRollups {
		totalExecutions += rollup.TotalQuantity
	}

	aiRollups, _ := h.repo.GetUsageByTenant(ctx, claims.TenantID, "ai_call", periodStart, periodEnd)
	totalAICalls := 0
	for _, rollup := range aiRollups {
		totalAICalls += rollup.TotalQuantity
	}

	functions, _ := h.repo.ListFunctionsByTenant(ctx, claims.TenantID)
	functionCount := len(functions)

	providerCount, _ := h.repo.CountBackendsByTenant(ctx, claims.TenantID)

	storageUsedBytes := 0
	if h.stateUsageAggregator != nil {
		if stateUsage, err := h.stateUsageAggregator.GetTenantStateUsage(ctx, claims.TenantID); err == nil {
			storageUsedBytes = int(stateUsage.TotalStorageBytes)
		}
	}
	usersCount, _ := h.repo.CountUsersByTenant(ctx, claims.TenantID)

	workflowCount := 0
	if _, total, err := h.repo.ListLifecycleWorkflows(ctx, claims.TenantID, 1, 0); err == nil {
		workflowCount = total
	}

	executionLimit := bundle.FeatureLimits["function_executions"]
	aiCallsLimit := bundle.FeatureLimits["ai_calls"]
	functionsLimit := bundle.FeatureLimits["functions"]
	providersLimit := bundle.FeatureLimits["providers"]
	storageLimit := bundle.FeatureLimits["storage_bytes"]
	userLimit := bundle.FeatureLimits["users"]
	requestsLimit := bundle.FeatureLimits["requests"]
	workflowsLimit := bundle.FeatureLimits["workflows"]

	usage := map[string]interface{}{}

	if executionLimit > 0 {
		percent := float64(totalExecutions) / float64(executionLimit) * 100
		if percent > 100 {
			percent = 100
		}
		usage["function_executions"] = map[string]interface{}{
			"used":   totalExecutions,
			"limit":  executionLimit,
			"percent": percent,
		}
	}

	if aiCallsLimit > 0 {
		percent := float64(totalAICalls) / float64(aiCallsLimit) * 100
		if percent > 100 {
			percent = 100
		}
		usage["ai_calls"] = map[string]interface{}{
			"used":   totalAICalls,
			"limit":  aiCallsLimit,
			"percent": percent,
		}
	}

	if functionsLimit > 0 {
		percent := float64(functionCount) / float64(functionsLimit) * 100
		if percent > 100 {
			percent = 100
		}
		usage["functions"] = map[string]interface{}{
			"used":   functionCount,
			"limit":  functionsLimit,
			"percent": percent,
		}
	}

	if providersLimit > 0 {
		percent := float64(providerCount) / float64(providersLimit) * 100
		if percent > 100 {
			percent = 100
		}
		usage["providers"] = map[string]interface{}{
			"used":   providerCount,
			"limit":  providersLimit,
			"percent": percent,
		}
	}

	if storageLimit > 0 {
		percent := float64(storageUsedBytes) / float64(storageLimit) * 100
		if percent > 100 {
			percent = 100
		}
		usage["storage_bytes"] = map[string]interface{}{
			"used":   storageUsedBytes,
			"limit":  storageLimit,
			"percent": percent,
		}
	}

	if userLimit > 0 {
		percent := float64(usersCount) / float64(userLimit) * 100
		if percent > 100 {
			percent = 100
		}
		usage["users"] = map[string]interface{}{
			"used":   usersCount,
			"limit":  userLimit,
			"percent": percent,
		}
	}

	if requestsLimit > 0 {
		percent := float64(totalExecutions) / float64(requestsLimit) * 100
		if percent > 100 {
			percent = 100
		}
		usage["requests"] = map[string]interface{}{
			"used":   totalExecutions,
			"limit":  requestsLimit,
			"percent": percent,
		}
	}

	if workflowsLimit > 0 {
		percent := 0.0
		if workflowsLimit > 0 {
			percent = float64(workflowCount) / float64(workflowsLimit) * 100
			if percent > 100 {
				percent = 100
			}
		}
		usage["workflows"] = map[string]interface{}{
			"used":   workflowCount,
			"limit":  workflowsLimit,
			"percent": percent,
		}
	}

	isAtLimit := (functionsLimit > 0 && functionCount >= functionsLimit) ||
		(providersLimit > 0 && providerCount >= providersLimit) ||
		(aiCallsLimit > 0 && totalAICalls >= aiCallsLimit)

	encodeJSON(w, http.StatusOK, map[string]interface{}{
		"bundle_slug":          bundle.Slug,
		"subscription_status":  sub.Status,
		"period_start":         periodStart.Format("2006-01-02"),
		"period_end":           periodEnd.Format("2006-01-02"),
		"usage":                usage,
		"is_at_limit":          isAtLimit,
		"upgrade_required":     isAtLimit,
	})
}

// HandleChangeBundle upgrades or downgrades the bundle subscription
// POST /v1/billing/bundle/change
func (h *Handler) HandleChangeBundle(w http.ResponseWriter, r *http.Request) {
	claims := middleware.GetUserFromContext(r)
	if claims == nil || claims.TenantID == uuid.Nil {
		writeJSONError(w, http.StatusUnauthorized, "Unauthorized")
		return
	}

	if !payment.IsConfigured() {
		writeJSONError(w, http.StatusServiceUnavailable, "Billing is not configured")
		return
	}

	var req struct {
		NewBundleSlug string `json:"new_bundle_slug"`
		SuccessURL    string `json:"success_url"`
		CancelURL     string `json:"cancel_url"`
	}
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		writeJSONError(w, http.StatusBadRequest, "Invalid request body")
		return
	}

	if req.NewBundleSlug == "" {
		writeJSONError(w, http.StatusBadRequest, "new_bundle_slug is required")
		return
	}

	validSlugs := map[string]bool{"saas-starter": true, "marketplace": true, "ai-app": true}
	if !validSlugs[req.NewBundleSlug] {
		writeJSONError(w, http.StatusBadRequest, "Invalid bundle slug")
		return
	}

	ctx := r.Context()
	sub, err := h.repo.GetBundleSubscriptionByTenant(ctx, claims.TenantID)
	if err != nil || sub == nil {
		writeJSONError(w, http.StatusNotFound, "No bundle subscription found")
		return
	}

	oldBundle, err := h.repo.GetPricingBundleByID(ctx, sub.BundleID)
	if err != nil || oldBundle == nil {
		writeJSONError(w, http.StatusInternalServerError, "Failed to retrieve current bundle")
		return
	}

	newBundle, err := h.repo.GetPricingBundleBySlug(ctx, req.NewBundleSlug)
	if err != nil || newBundle == nil {
		writeJSONError(w, http.StatusNotFound, "Bundle not found")
		return
	}

	if newBundle.StripePriceID == "" {
		writeJSONError(w, http.StatusServiceUnavailable, "Bundle not available for upgrade")
		return
	}

	user, err := h.repo.GetUserByID(ctx, claims.UserID)
	if err != nil || user == nil {
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

	resp, err := payment.CreateBundleCheckoutSession(
		ctx,
		h.repo,
		claims.TenantID,
		user.Email,
		name,
		payment.CreateBundleCheckoutSessionRequest{
			PriceID:    newBundle.StripePriceID,
			SuccessURL: req.SuccessURL,
			CancelURL:  req.CancelURL,
			BundleSlug: newBundle.Slug,
		},
	)
	if err != nil {
		logrus.WithError(err).WithField("tenant_id", claims.TenantID).Error("billing bundle change: failed to create session")
		writeJSONError(w, http.StatusInternalServerError, "Failed to create checkout session")
		return
	}

	encodeJSON(w, http.StatusOK, resp)
}

// HandleRegisterFounderMode registers a user for founder mode (free until trigger)
// POST /v1/billing/bundles/:slug/founder
func (h *Handler) HandleRegisterFounderMode(w http.ResponseWriter, r *http.Request) {
	claims := middleware.GetUserFromContext(r)
	if claims == nil || claims.TenantID == uuid.Nil {
		writeJSONError(w, http.StatusUnauthorized, "Unauthorized")
		return
	}

	slug := mux.Vars(r)["slug"]
	if slug == "" {
		writeJSONError(w, http.StatusBadRequest, "Bundle slug is required")
		return
	}

	// Validate slug format to prevent injection and enumeration
	validSlugs := map[string]bool{"saas-starter": true, "marketplace": true, "ai-app": true}
	if !validSlugs[slug] {
		writeJSONError(w, http.StatusBadRequest, "Invalid bundle slug")
		return
	}

	var req RegisterFounderModeRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		writeJSONError(w, http.StatusBadRequest, "Invalid request body")
		return
	}

	// Validate request fields
	if req.ModeType != "" {
		validModes := map[string]bool{"time_based": true, "revenue_based": true, "hybrid": true}
		if !validModes[req.ModeType] {
			writeJSONError(w, http.StatusBadRequest, "Invalid mode_type")
			return
		}
	}
	if req.FreeDays < 0 || req.FreeDays > 365 {
		writeJSONError(w, http.StatusBadRequest, "free_days must be between 0 and 365")
		return
	}
	if req.MRRThreshold < 0 || req.MRRThreshold > 100000000 { // Max $1M MRR
		writeJSONError(w, http.StatusBadRequest, "mrr_threshold is out of valid range")
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

	encodeJSON(w, http.StatusCreated, founderModeToResponse(registration, bundle.Slug))
}

// HandleCreateBundleCheckout creates a Stripe checkout for immediate bundle subscription
// POST /v1/billing/bundles/:slug/checkout
func (h *Handler) HandleCreateBundleCheckout(w http.ResponseWriter, r *http.Request) {
	claims := middleware.GetUserFromContext(r)
	ctx := r.Context()
	if claims == nil || claims.TenantID == uuid.Nil {
		writeJSONError(w, http.StatusUnauthorized, "Unauthorized")
		return
	}

	if !payment.IsConfigured() {
		writeJSONError(w, http.StatusServiceUnavailable, "Billing is not configured")
		return
	}

	slug := mux.Vars(r)["slug"]
	if slug == "" {
		writeJSONError(w, http.StatusBadRequest, "Bundle slug is required")
		return
	}

	// Validate slug format to prevent injection and enumeration
	validSlugs := map[string]bool{"saas-starter": true, "marketplace": true, "ai-app": true}
	if !validSlugs[slug] {
		writeJSONError(w, http.StatusBadRequest, "Invalid bundle slug")
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

	user, err := h.repo.GetUserByID(ctx, claims.UserID)
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

	// Create checkout session with bundle metadata for proper webhook processing
	resp, err := payment.CreateBundleCheckoutSession(
		r.Context(),
		h.repo,
		claims.TenantID,
		user.Email,
		name,
		payment.CreateBundleCheckoutSessionRequest{
			PriceID:    bundle.StripePriceID,
			SuccessURL: req.SuccessURL,
			CancelURL:  req.CancelURL,
			BundleSlug: bundle.Slug,
			Provider:   req.Provider,
			ProviderID: req.ProviderID,
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

	encodeJSON(w, http.StatusOK, resp)
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

	encodeJSON(w, http.StatusOK, map[string]interface{}{
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

	ctx := r.Context()
	// Get active founder mode
	registrations, err := h.repo.ListActiveFounderModesByTenant(r.Context(), claims.TenantID)
	if err != nil {
		logrus.WithError(err).Error("billing deferred: failed to list founder modes")
		writeJSONError(w, http.StatusInternalServerError, "Failed to retrieve deferred billing status")
		return
	}

	if len(registrations) == 0 {
		encodeJSON(w, http.StatusOK, map[string]interface{}{
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

	execRollups, _ := h.repo.GetUsageByTenant(ctx, claims.TenantID, "function_execution", periodStart, periodEnd)
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

	encodeJSON(w, http.StatusOK, DeferredBillingStatus{
		BundleID:           reg.BundleID,
		Status:             status,
		TriggerThresholds:  thresholds,
		CurrentProgress:    progress,
		ProgressPercent:    progressPercent,
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
	ctx := r.Context()

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
	user, err := h.repo.GetUserByID(ctx, claims.UserID)
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
			PriceID:      bundle.StripePriceID,
			SuccessURL:   "/dashboard?converted=true",
			CancelURL:    "/pricing",
			FounderModeID: reg.ID.String(), // Track for webhook processing
		},
	)
	if err != nil {
		logrus.WithError(err).WithField("tenant_id", claims.TenantID).Error("billing convert: failed to create checkout")
		writeJSONError(w, http.StatusInternalServerError, "Failed to create checkout session")
		return
	}

	encodeJSON(w, http.StatusOK, resp)
}

// HandleDeployBundle initiates a bundle deployment and returns deployment status
// POST /v1/billing/bundles/:slug/deploy
func (h *Handler) HandleDeployBundle(w http.ResponseWriter, r *http.Request) {
	claims := middleware.GetUserFromContext(r)
	if claims == nil || claims.TenantID == uuid.Nil {
		writeJSONError(w, http.StatusUnauthorized, "Unauthorized")
		return
	}

	slug := mux.Vars(r)["slug"]
	if slug == "" {
		writeJSONError(w, http.StatusBadRequest, "Bundle slug is required")
		return
	}

	validSlugs := map[string]bool{"saas-starter": true, "marketplace": true, "ai-app": true}
	if !validSlugs[slug] {
		writeJSONError(w, http.StatusBadRequest, "Invalid bundle slug")
		return
	}

	// Paywall check: user must have an active subscription or founder mode registration
	bundle, err := h.repo.GetPricingBundleBySlug(r.Context(), slug)
	if err != nil || bundle == nil {
		writeJSONError(w, http.StatusNotFound, "Bundle not found")
		return
	}

	hasAccess := false

	// Check for active bundle subscription
	sub, err := h.repo.GetBundleSubscriptionByTenant(r.Context(), claims.TenantID)
	if err == nil && sub != nil && sub.BundleID == bundle.ID && (sub.Status == "active" || sub.Status == "deferred") {
		hasAccess = true
	}

	// Check for active founder mode registration
	if !hasAccess {
		founderMode, err := h.repo.GetActiveFounderMode(r.Context(), claims.TenantID, bundle.ID)
		if err == nil && founderMode != nil && founderMode.Status == "active" {
			hasAccess = true
		}
	}

	if !hasAccess {
		writeJSONError(w, http.StatusPaymentRequired, "Payment or founder mode registration required. Please select 'Pay Now' or 'Start Free' first.")
		return
	}

	var req DeployBundleRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		// Default region if not provided
		req.Region = "us-east-1"
	}
	if req.Region == "" {
		req.Region = "us-east-1"
	}

	// Trigger async provisioning (idempotent - safe to call multiple times) and create app/backend
	// When isolated provisioning is available, skip creating backend/functions in the shared DB
	// — the BundleProvisioner handles them in the tenant's dedicated database.
	var provisionOpts []func(*ProvisionBundleOpts)
	if h.provisionBundleFn != nil {
		provisionOpts = append(provisionOpts, WithIsolatedProvisioning())
	}
	app, err := ProvisionBundleAppAndBackend(h.repo, claims.TenantID, slug, provisionOpts...)
	if err != nil {
		logrus.WithError(err).WithField("tenant_id", claims.TenantID).Error("billing deploy: failed to provision app and backend")
		writeJSONError(w, http.StatusInternalServerError, "Failed to create deployment")
		return
	}

	// Update subscription with the default app ID so the bundle overview can link to it
	if app != nil {
		bundleSub, subErr := h.repo.GetBundleSubscriptionByTenant(r.Context(), claims.TenantID)
		if subErr == nil && bundleSub != nil && bundleSub.BundleID == bundle.ID {
			bundleSub.DefaultAppID = &app.ID
			if updateErr := h.repo.UpdateBundleSubscription(r.Context(), bundleSub); updateErr != nil {
				logrus.WithError(updateErr).WithField("tenant_id", claims.TenantID).Warn("Failed to update subscription with app ID")
			}
		}
	}

	// Return deployment response with steps
	deploymentID := uuid.New().String()
	steps := []DeploymentStepResponse{
		{ID: "create_subscription", Name: "Subscription", Description: "Activating bundle subscription", Status: "completed"},
		{ID: "provision_app", Name: "App", Description: "Creating app with bundle configuration", Status: "completed"},
		{ID: "provision_backend", Name: "Backend", Description: "Setting up backend infrastructure", Status: "completed"},
		{ID: "provision_functions", Name: "Functions", Description: "Deploying function templates", Status: "completed"},
		{ID: "provision_auth", Name: "Auth", Description: "Configuring authentication providers", Status: "completed"},
		{ID: "provision_email_workflows", Name: "Email", Description: "Setting up email workflows and templates", Status: "completed"},
		{ID: "finalize", Name: "Finalizing", Description: "Completing deployment", Status: "completed"},
	}

	// Store deployment result in Redis for status polling
	deploymentData := map[string]interface{}{
		"status":      "completed",
		"app_id":      app.ID.String(),
		"backend_id":  "",
		"components": map[string]map[string]string{
			"create_subscription":      {"status": "active"},
			"provision_app":            {"status": "active"},
			"provision_backend":        {"status": "active"},
			"provision_functions":       {"status": "active"},
			"provision_auth":           {"status": "active"},
			"provision_email_workflows": {"status": "active"},
			"finalize":                 {"status": "active"},
		},
	}
	if jsonData, err := json.Marshal(deploymentData); err == nil {
		ctx := r.Context()
		key := fmt.Sprintf("deployment:%s", deploymentID)
		// Store for 1 hour
		h.redisClient.Set(ctx, key, jsonData, time.Hour)
	}

	encodeJSON(w, http.StatusOK, DeploymentResponse{
		DeploymentID: deploymentID,
		Status:      "completed",
		Message:     "Bundle deployed successfully",
		AppID:       app.ID.String(),
		BackendID:   "",
		Steps:       steps,
	})
}

// HandleGetFounderModeAnalytics returns aggregate analytics for the founder mode funnel
// GET /v1/billing/bundles/analytics/founder-mode
func (h *Handler) HandleGetFounderModeAnalytics(w http.ResponseWriter, r *http.Request) {
	claims := middleware.GetUserFromContext(r)
	if claims == nil {
		writeJSONError(w, http.StatusUnauthorized, "Unauthorized")
		return
	}

	analytics, err := h.repo.GetFounderModeAnalytics(r.Context())
	if err != nil {
		logrus.WithError(err).Error("billing founder mode analytics: failed to get analytics")
		writeJSONError(w, http.StatusInternalServerError, "Failed to retrieve analytics")
		return
	}

	encodeJSON(w, http.StatusOK, analytics)
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

// HandleGetDeploymentStatus returns the status of a bundle deployment by polling Redis
// GET /v1/billing/deployments/{deploymentId}/status
func (h *Handler) HandleGetDeploymentStatus(w http.ResponseWriter, r *http.Request) {
	claims := middleware.GetUserFromContext(r)
	if claims == nil || claims.TenantID == uuid.Nil {
		writeJSONError(w, http.StatusUnauthorized, "Unauthorized")
		return
	}

	deploymentID := mux.Vars(r)["deploymentId"]
	if deploymentID == "" {
		writeJSONError(w, http.StatusBadRequest, "Deployment ID is required")
		return
	}

	// Try to get deployment from Redis (set during HandleDeployBundle)
	ctx := r.Context()
	key := fmt.Sprintf("deployment:%s", deploymentID)
	data, err := h.redisClient.Get(ctx, key).Bytes()
	if err != nil || len(data) == 0 {
		encodeJSON(w, http.StatusOK, DeploymentStatusResponse{
			DeploymentID: deploymentID,
			Status:       "unknown",
			Message:      "Deployment not found or expired",
			Progress:     0,
			Steps:        []DeploymentStepResponse{},
		})
		return
	}

	// Parse stored deployment result
	var result struct {
		Status     string
		Components map[string]struct {
			Status string
			Error  string
		}
		AppID    string
		BackendID string
	}
	if err := json.Unmarshal(data, &result); err != nil {
		writeJSONError(w, http.StatusInternalServerError, "Failed to parse deployment data")
		return
	}

	// Convert to deployment status response
	status := "in_progress"
	progress := 0
	switch result.Status {
	case "active":
		status = "completed"
		progress = 100
	case "failed":
		status = "failed"
	case "provisioning":
		status = "in_progress"
	}

	steps := []DeploymentStepResponse{}
	for name, state := range result.Components {
		stepStatus := "pending"
		switch state.Status {
		case "active":
			stepStatus = "completed"
		case "failed":
			stepStatus = "failed"
		case "provisioning":
			stepStatus = "in_progress"
		}
		steps = append(steps, DeploymentStepResponse{
			ID:          name,
			Name:        name,
			Description: fmt.Sprintf("Component: %s", name),
			Status:      stepStatus,
			Error:       state.Error,
		})
	}

	encodeJSON(w, http.StatusOK, DeploymentStatusResponse{
		DeploymentID: deploymentID,
		Status:       status,
		Message:      fmt.Sprintf("Deployment is %s", status),
		Progress:     progress,
		CurrentStep: "",
		Steps:        steps,
		AppID:       result.AppID,
		BackendID:   result.BackendID,
	})
}

// HandleGetBundleSubscription returns the current bundle subscription for the tenant.
func (h *Handler) HandleGetBundleSubscription(w http.ResponseWriter, r *http.Request) {
	claims := middleware.GetUserFromContext(r)
	if claims == nil || claims.TenantID == uuid.Nil {
		writeJSONError(w, http.StatusUnauthorized, "Unauthorized")
		return
	}

	sub, err := h.repo.GetBundleSubscriptionByTenant(r.Context(), claims.TenantID)
	if err != nil {
		writeJSONError(w, http.StatusInternalServerError, "Failed to retrieve subscription")
		return
	}
	if sub == nil {
		writeJSONError(w, http.StatusNotFound, "No bundle subscription found")
		return
	}

	encodeJSON(w, http.StatusOK, sub)
}

// HandleGetBundleCatalog returns rich metadata for all bundles (features, functions, integrations).
// This is a public endpoint — no auth required.
// GET /v1/billing/bundles/catalog
func (h *Handler) HandleGetBundleCatalog(w http.ResponseWriter, r *http.Request) {
	bundles, err := h.repo.ListPricingBundles(r.Context(), true)
	if err != nil {
		logrus.WithError(err).Error("billing bundles: failed to list bundles for catalog")
		writeJSONError(w, http.StatusInternalServerError, "Failed to retrieve bundle catalog")
		return
	}

	richMetadata := map[string]BundleCatalogItem{
		"saas-starter": {
			Slug:     "saas-starter",
			Name:     "SaaS Starter",
			Tagline:  "Full SaaS backend ready to customize",
			PriceUSD: "$29/mo",
			Icon:     "Rocket",
			Gradient: "from-blue-500 to-cyan-500",
			Features: []FeatureWithDetails{
				{Icon: "Shield", Title: "Authentication", Description: "JWT, OAuth (Google, GitHub), sessions, MFA"},
				{Icon: "CreditCard", Title: "Payments", Description: "Stripe integration, products, invoices, webhooks"},
				{Icon: "Mail", Title: "Email Workflows", Description: "20+ templates, welcome sequence, dunning"},
				{Icon: "BarChart3", Title: "Analytics", Description: "Event tracking, funnels, cohorts, dashboards"},
			},
			ProvisioningSteps: []ProvisioningStepMeta{
				{Label: "Functions", Description: "Explore your pre-configured functions"},
				{Label: "Integrations", Description: "Set up Stripe, OAuth, and email"},
				{Label: "Deploy", Description: "Push to production"},
			},
			Functions: map[string]FunctionMetadata{
				"stripe-webhook": {
					Description:  "Handles Stripe subscription and payment webhook events. Processes customer subscriptions, successful payments, and failed invoices.",
					Icon:         "Webhook",
					Capabilities: []string{"Webhook", "Storage"},
				},
				"welcome-email": {
					Description:  "Sends personalized welcome emails to new users after signup using your configured email templates.",
					Icon:         "Mail",
					Capabilities: []string{"Email"},
				},
			},
			Integrations: []IntegrationMetadata{
				{Title: "Stripe Payments", Description: "Subscriptions, invoices, webhooks, and billing management", Icon: "CreditCard"},
				{Title: "OAuth Providers", Description: "Google, GitHub, and social login authentication", Icon: "Shield"},
				{Title: "Email Workflows", Description: "Transactional emails, welcome sequences, and dunning flows", Icon: "Mail"},
				{Title: "Analytics", Description: "Event tracking, funnels, cohorts, and dashboards", Icon: "BarChart3"},
			},
		},
		"marketplace": {
			Slug:     "marketplace",
			Name:     "Marketplace",
			Tagline:  "Multi-vendor marketplace backend",
			PriceUSD: "$49/mo",
			Icon:     "Store",
			Gradient: "from-emerald-500 to-teal-500",
			Features: []FeatureWithDetails{
				{Icon: "Store", Title: "Listings", Description: "Categories, variants, reviews, full-text search"},
				{Icon: "CreditCard", Title: "Stripe Connect", Description: "Split payments, seller payouts, refunds"},
				{Icon: "MessageSquare", Title: "Messaging", Description: "Buyer-seller conversations, offers, files"},
				{Icon: "Bell", Title: "Notifications", Description: "22 templates, in-app + email alerts"},
			},
			ProvisioningSteps: []ProvisioningStepMeta{
				{Label: "Functions", Description: "Explore your pre-configured functions"},
				{Label: "Integrations", Description: "Set up Stripe Connect and messaging"},
				{Label: "Deploy", Description: "Push to production"},
			},
			Functions: map[string]FunctionMetadata{
				"create-listing": {
					Description:  "Creates marketplace listings with seller info, pricing, and categories. Stores listing data for search and retrieval.",
					Icon:         "Store",
					Capabilities: []string{"Storage"},
				},
				"send-message": {
					Description:  "Enables buyer-seller messaging with persistent conversation threads and file support.",
					Icon:         "MessageSquare",
					Capabilities: []string{"Storage"},
				},
			},
			Integrations: []IntegrationMetadata{
				{Title: "Stripe Connect", Description: "Split payments, seller payouts, and marketplace billing", Icon: "CreditCard"},
				{Title: "Messaging", Description: "Buyer-seller conversations, offers, and file attachments", Icon: "Webhook"},
				{Title: "Notifications", Description: "In-app alerts, email notifications, and templates", Icon: "Mail"},
				{Title: "Listings", Description: "Categories, variants, reviews, and full-text search", Icon: "Settings"},
			},
		},
		"ai-app": {
			Slug:     "ai-app",
			Name:     "AI App",
			Tagline:  "AI-powered backend with vector search",
			PriceUSD: "$39/mo",
			Icon:     "Brain",
			Gradient: "from-violet-500 to-purple-500",
			Features: []FeatureWithDetails{
				{Icon: "Cpu", Title: "Vector DB", Description: "pgvector with HNSW indexing, collections"},
				{Icon: "Database", Title: "Embeddings", Description: "OpenAI embeddings, document chunking, RAG"},
				{Icon: "MessageSquare", Title: "Chat Workflows", Description: "AI assistants, tool calling, guardrails"},
				{Icon: "Sparkles", Title: "Memory System", Description: "Long-term semantic memory, user profiles"},
			},
			ProvisioningSteps: []ProvisioningStepMeta{
				{Label: "Functions", Description: "Explore your pre-configured functions"},
				{Label: "Integrations", Description: "Configure OpenAI and vector collections"},
				{Label: "Deploy", Description: "Push to production"},
			},
			Functions: map[string]FunctionMetadata{
				"chat-completion": {
					Description:  "AI chat completions via OpenAI-compatible API with streaming support and conversation context.",
					Icon:         "Cpu",
					Capabilities: []string{"AI"},
				},
				"embed-and-store": {
					Description:  "Generates text embeddings and stores them in your vector collection for semantic search and RAG.",
					Icon:         "Database",
					Capabilities: []string{"AI", "Storage"},
				},
			},
			Integrations: []IntegrationMetadata{
				{Title: "OpenAI", Description: "Chat completions, embeddings, and AI model configuration", Icon: "Shield"},
				{Title: "Vector DB", Description: "pgvector collections, HNSW indexing, and semantic search", Icon: "BarChart3"},
				{Title: "Memory System", Description: "Long-term semantic memory and user profiles", Icon: "Mail"},
				{Title: "Chat Workflows", Description: "AI assistants, tool calling, and guardrails", Icon: "Settings"},
			},
		},
	}

	catalogItems := make([]BundleCatalogItem, 0, len(bundles))
	for _, bundle := range bundles {
		if metadata, ok := richMetadata[bundle.Slug]; ok {
			catalogItems = append(catalogItems, metadata)
		} else {
			catalogItems = append(catalogItems, BundleCatalogItem{
				Slug:     bundle.Slug,
				Name:     bundle.Name,
				PriceUSD: fmt.Sprintf("$%.2f/mo", float64(bundle.DisplayPriceCents)/100),
				Icon:     "Box",
				Gradient: "from-gray-500 to-gray-600",
			})
		}
	}

	catalog := BundleCatalogResponse{
		Bundles: catalogItems,
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	if err := json.NewEncoder(w).Encode(catalog); err != nil {
		logrus.WithError(err).Error("billing bundles: failed to encode catalog response")
	}
}

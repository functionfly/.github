package billing

import (
	"encoding/json"
	"net/http"
	"time"

	"github.com/functionfly/functionfly/internal/api/middleware"
	"github.com/functionfly/functionfly/internal/auth"
	"github.com/functionfly/functionfly/internal/services"
	"github.com/functionfly/functionfly/internal/storage"
	"github.com/google/uuid"
	"github.com/gorilla/mux"
	"github.com/sirupsen/logrus"
)

// StateUsageHandler handles state usage API endpoints for billing and quota
type StateUsageHandler struct {
	aggregator *services.StateUsageAggregator
	repo       storage.Repository
	logger     *logrus.Logger
}

// NewStateUsageHandler creates a new state usage handler
func NewStateUsageHandler(aggregator *services.StateUsageAggregator, repo storage.Repository) *StateUsageHandler {
	return &StateUsageHandler{
		aggregator: aggregator,
		repo:       repo,
		logger:     logrus.New(),
	}
}

// StateUsageResponse represents the response for state usage queries
type StateUsageResponse struct {
	TenantID           uuid.UUID `json:"tenant_id"`
	PeriodStart        time.Time `json:"period_start"`
	PeriodEnd          time.Time `json:"period_end"`
	StorageBytes       int64     `json:"storage_bytes"`
	StorageMB          float64   `json:"storage_mb"`
	ReadOps            int64     `json:"read_ops"`
	WriteOps           int64     `json:"write_ops"`
	ActiveStates       int64     `json:"active_states"`
	EstimatedCostCents int64     `json:"estimated_cost_cents,omitempty"`
}

// GetCurrentStateUsage returns the current state usage for the authenticated tenant
//
// GET /api/v1/usage/state
//
// Response:
//
//	{
//	  "tenant_id": "uuid",
//	  "period_start": "2026-04-01T00:00:00Z",
//	  "period_end": "2026-04-30T23:59:59Z",
//	  "storage_bytes": 1048576,
//	  "storage_mb": 1.0,
//	  "read_ops": 100,
//	  "write_ops": 50,
//	  "active_states": 5,
//	  "estimated_cost_cents": 10
//	}
func (h *StateUsageHandler) GetCurrentStateUsage(w http.ResponseWriter, r *http.Request) {
	tenantID := h.extractTenantID(r)
	if tenantID == uuid.Nil {
		writeError(w, http.StatusUnauthorized, "Unauthorized: Tenant ID not found")
		return
	}

	ctx := r.Context()
	usage, err := h.aggregator.GetTenantStateUsage(ctx, tenantID)
	if err != nil {
		h.logger.WithError(err).WithField("tenant_id", tenantID).Error("Failed to get tenant state usage")
		writeError(w, http.StatusInternalServerError, "Failed to retrieve state usage")
		return
	}

	// Calculate estimated cost (1 cent per 1000 ops, 1 cent per 100 MB storage)
	estimatedCost := (usage.TotalStorageBytes / (1024 * 1024) / 100) + (usage.TotalReadOps+usage.TotalWriteOps)/1000

	response := StateUsageResponse{
		TenantID:           usage.TenantID,
		PeriodStart:        usage.PeriodStart,
		PeriodEnd:          usage.PeriodEnd,
		StorageBytes:       usage.TotalStorageBytes,
		StorageMB:          float64(usage.TotalStorageBytes) / (1024 * 1024),
		ReadOps:            usage.TotalReadOps,
		WriteOps:           usage.TotalWriteOps,
		ActiveStates:       usage.ActiveStates,
		EstimatedCostCents: estimatedCost,
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(response)
}

// GetStateUsageHistory returns historical state usage for the tenant
//
// GET /api/v1/usage/state/history?start=2026-04-01&end=2026-04-09
//
// Response:
//
//	{
//	  "tenant_id": "uuid",
//	  "period_start": "2026-04-01T00:00:00Z",
//	  "period_end": "2026-04-09T00:00:00Z",
//	  "daily_usage": [...],
//	  "total_storage_mb": 10.5,
//	  "total_read_ops": 1000,
//	  "total_write_ops": 500
//	}
func (h *StateUsageHandler) GetStateUsageHistory(w http.ResponseWriter, r *http.Request) {
	tenantID := h.extractTenantID(r)
	if tenantID == uuid.Nil {
		writeError(w, http.StatusUnauthorized, "Unauthorized: Tenant ID not found")
		return
	}

	// Parse date range
	start, end, err := h.parseDateRange(r)
	if err != nil {
		writeError(w, http.StatusBadRequest, "Invalid date range: "+err.Error())
		return
	}

	// Get usage rollups from the billing repository
	storageRollups, err := h.repo.GetUsageByTenant(r.Context(), tenantID, "state_storage", start, end)
	if err != nil {
		h.logger.WithError(err).Error("Failed to get storage usage rollups")
		writeError(w, http.StatusInternalServerError, "Failed to retrieve usage data")
		return
	}

	readRollups, err := h.repo.GetUsageByTenant(r.Context(), tenantID, "state_read_ops", start, end)
	if err != nil {
		h.logger.WithError(err).Error("Failed to get read ops rollups")
		writeError(w, http.StatusInternalServerError, "Failed to retrieve usage data")
		return
	}

	writeRollups, err := h.repo.GetUsageByTenant(r.Context(), tenantID, "state_write_ops", start, end)
	if err != nil {
		h.logger.WithError(err).Error("Failed to get write ops rollups")
		writeError(w, http.StatusInternalServerError, "Failed to retrieve usage data")
		return
	}

	// Calculate totals
	totalStorage := 0
	for _, rollup := range storageRollups {
		totalStorage += rollup.TotalQuantity
	}

	totalReadOps := 0
	for _, rollup := range readRollups {
		totalReadOps += rollup.TotalQuantity
	}

	totalWriteOps := 0
	for _, rollup := range writeRollups {
		totalWriteOps += rollup.TotalQuantity
	}

	// Build daily breakdown
	type DailyUsage struct {
		Date      time.Time `json:"date"`
		StorageMB int       `json:"storage_mb"`
		ReadOps   int       `json:"read_ops"`
		WriteOps  int       `json:"write_ops"`
		CostCents int       `json:"cost_cents"`
	}

	dailyMap := make(map[string]*DailyUsage)

	for _, rollup := range storageRollups {
		dateKey := rollup.PeriodDate.Format("2006-01-02")
		daily, exists := dailyMap[dateKey]
		if !exists {
			daily = &DailyUsage{Date: rollup.PeriodDate}
			dailyMap[dateKey] = daily
		}
		daily.StorageMB += rollup.TotalQuantity
	}

	for _, rollup := range readRollups {
		dateKey := rollup.PeriodDate.Format("2006-01-02")
		daily, exists := dailyMap[dateKey]
		if !exists {
			daily = &DailyUsage{Date: rollup.PeriodDate}
			dailyMap[dateKey] = daily
		}
		daily.ReadOps += rollup.TotalQuantity
	}

	for _, rollup := range writeRollups {
		dateKey := rollup.PeriodDate.Format("2006-01-02")
		daily, exists := dailyMap[dateKey]
		if !exists {
			daily = &DailyUsage{Date: rollup.PeriodDate}
			dailyMap[dateKey] = daily
		}
		daily.WriteOps += rollup.TotalQuantity
	}

	// Calculate costs and convert to slice
	dailyUsage := make([]*DailyUsage, 0, len(dailyMap))
	for _, daily := range dailyMap {
		daily.CostCents = (daily.StorageMB / 100) + (daily.ReadOps+daily.WriteOps)/1000
		dailyUsage = append(dailyUsage, daily)
	}

	response := map[string]interface{}{
		"tenant_id":        tenantID.String(),
		"period_start":     start,
		"period_end":       end,
		"total_storage_mb": totalStorage,
		"total_read_ops":   totalReadOps,
		"total_write_ops":  totalWriteOps,
		"daily_usage":      dailyUsage,
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(response)
}

// AdminGetTenantStateUsage returns state usage for any tenant (admin only)
//
// GET /api/v1/admin/usage/state/{tenant_id}
func (h *StateUsageHandler) AdminGetTenantStateUsage(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	tenantIDStr := vars["tenant_id"]

	tenantID, err := uuid.Parse(tenantIDStr)
	if err != nil {
		writeError(w, http.StatusBadRequest, "Invalid tenant ID")
		return
	}

	ctx := r.Context()
	usage, err := h.aggregator.GetTenantStateUsage(ctx, tenantID)
	if err != nil {
		h.logger.WithError(err).WithField("tenant_id", tenantID).Error("Failed to get tenant state usage")
		writeError(w, http.StatusInternalServerError, "Failed to retrieve state usage")
		return
	}

	estimatedCost := (usage.TotalStorageBytes / (1024 * 1024) / 100) + (usage.TotalReadOps+usage.TotalWriteOps)/1000

	response := StateUsageResponse{
		TenantID:           usage.TenantID,
		PeriodStart:        usage.PeriodStart,
		PeriodEnd:          usage.PeriodEnd,
		StorageBytes:       usage.TotalStorageBytes,
		StorageMB:          float64(usage.TotalStorageBytes) / (1024 * 1024),
		ReadOps:            usage.TotalReadOps,
		WriteOps:           usage.TotalWriteOps,
		ActiveStates:       usage.ActiveStates,
		EstimatedCostCents: estimatedCost,
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(response)
}

// AdminListAllStateUsage returns state usage for all tenants (admin only)
//
// GET /api/v1/admin/usage/state?start=2026-04-01&end=2026-04-09
func (h *StateUsageHandler) AdminListAllStateUsage(w http.ResponseWriter, r *http.Request) {
	// Parse date range
	start, end, err := h.parseDateRange(r)
	if err != nil {
		writeError(w, http.StatusBadRequest, "Invalid date range: "+err.Error())
		return
	}

	ctx := r.Context()
	usages, err := h.aggregator.ListTenantStateUsage(ctx, start, end)
	if err != nil {
		h.logger.WithError(err).Error("Failed to list tenant state usage")
		writeError(w, http.StatusInternalServerError, "Failed to retrieve usage data")
		return
	}

	var responses []*StateUsageResponse
	for _, usage := range usages {
		estimatedCost := (usage.TotalStorageBytes / (1024 * 1024) / 100) + (usage.TotalReadOps+usage.TotalWriteOps)/1000

		responses = append(responses, &StateUsageResponse{
			TenantID:           usage.TenantID,
			PeriodStart:        usage.PeriodStart,
			PeriodEnd:          usage.PeriodEnd,
			StorageBytes:       usage.TotalStorageBytes,
			StorageMB:          float64(usage.TotalStorageBytes) / (1024 * 1024),
			ReadOps:            usage.TotalReadOps,
			WriteOps:           usage.TotalWriteOps,
			ActiveStates:       usage.ActiveStates,
			EstimatedCostCents: estimatedCost,
		})
	}

	response := map[string]interface{}{
		"period_start": start,
		"period_end":   end,
		"tenant_count": len(responses),
		"usage":        responses,
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(response)
}

// GetStateQuotaStatus returns the state-specific quota status for the authenticated tenant
//
// GET /api/v1/usage/state/quota
//
// Response:
//
//	{
//	  "tenant_id": "uuid",
//	  "storage_limit_mb": 100,
//	  "storage_used_mb": 50,
//	  "storage_percent": 50.0,
//	  "read_ops_limit": 10000,
//	  "read_ops_used": 1000,
//	  "write_ops_limit": 5000,
//	  "write_ops_used": 500,
//	  "state_limit": 10,
//	  "state_used": 5,
//	  "status": "ok"
//	}
func (h *StateUsageHandler) GetStateQuotaStatus(w http.ResponseWriter, r *http.Request) {
	tenantID := h.extractTenantID(r)
	if tenantID == uuid.Nil {
		writeError(w, http.StatusUnauthorized, "Unauthorized: Tenant ID not found")
		return
	}

	ctx := r.Context()
	usage, err := h.aggregator.GetTenantStateUsage(ctx, tenantID)
	if err != nil {
		h.logger.WithError(err).WithField("tenant_id", tenantID).Error("Failed to get tenant state usage")
		writeError(w, http.StatusInternalServerError, "Failed to retrieve state quota status")
		return
	}

	// Get subscription for limits
	sub, err := h.repo.GetSubscriptionByTenantID(r.Context(), tenantID)
	if err != nil {
		h.logger.WithError(err).Error("Failed to get subscription")
		writeError(w, http.StatusInternalServerError, "Failed to retrieve quota limits")
		return
	}

	// Default limits
	storageLimitMB := 100
	readOpsLimit := 10000
	writeOpsLimit := 5000
	stateLimit := 10

	// Extract limits from pricing tier if available
	if sub != nil && sub.PricingTier != nil && sub.PricingTier.Features != nil {
		if features, ok := sub.PricingTier.Features.(map[string]interface{}); ok {
			if v, ok := features["state_storage_mb"].(float64); ok {
				storageLimitMB = int(v)
			}
			if v, ok := features["state_read_ops"].(float64); ok {
				readOpsLimit = int(v)
			}
			if v, ok := features["state_write_ops"].(float64); ok {
				writeOpsLimit = int(v)
			}
			if v, ok := features["max_states"].(float64); ok {
				stateLimit = int(v)
			}
		}
	}

	// Calculate percentages
	storageUsedMB := int(usage.TotalStorageBytes / (1024 * 1024))
	storagePercent := float64(storageUsedMB) / float64(storageLimitMB) * 100
	if storageLimitMB == 0 {
		storagePercent = 0
	}

	readOpsPercent := float64(usage.TotalReadOps) / float64(readOpsLimit) * 100
	if readOpsLimit == 0 {
		readOpsPercent = 0
	}

	writeOpsPercent := float64(usage.TotalWriteOps) / float64(writeOpsLimit) * 100
	if writeOpsLimit == 0 {
		writeOpsPercent = 0
	}

	statePercent := float64(usage.ActiveStates) / float64(stateLimit) * 100
	if stateLimit == 0 {
		statePercent = 0
	}

	// Determine status
	status := "ok"
	if storagePercent >= 100 || readOpsPercent >= 100 || writeOpsPercent >= 100 {
		status = "exceeded"
	} else if storagePercent >= 90 || readOpsPercent >= 90 || writeOpsPercent >= 90 {
		status = "critical"
	} else if storagePercent >= 70 || readOpsPercent >= 70 || writeOpsPercent >= 70 {
		status = "warning"
	}

	response := map[string]interface{}{
		"tenant_id":         tenantID.String(),
		"storage_limit_mb":  storageLimitMB,
		"storage_used_mb":   storageUsedMB,
		"storage_percent":   storagePercent,
		"read_ops_limit":    readOpsLimit,
		"read_ops_used":     usage.TotalReadOps,
		"read_ops_percent":  readOpsPercent,
		"write_ops_limit":   writeOpsLimit,
		"write_ops_used":    usage.TotalWriteOps,
		"write_ops_percent": writeOpsPercent,
		"state_limit":       stateLimit,
		"state_used":        usage.ActiveStates,
		"state_percent":     statePercent,
		"status":            status,
		"last_updated":      time.Now().UTC(),
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(response)
}

// Helper methods

func (h *StateUsageHandler) extractTenantID(r *http.Request) uuid.UUID {
	// Use the middleware's GetTenantID function which handles context properly
	if tenantID, ok := middleware.GetTenantID(r); ok {
		return tenantID
	}
	return uuid.Nil
}

// GetClaimsFromContext extracts auth claims from context for tenant ID lookup
func (h *StateUsageHandler) GetClaimsFromContext(ctx interface{}) *auth.Claims {
	// This is a helper that will use the middleware's GetClaimsFromContext
	// when we have access to the actual context
	return nil
}

func (h *StateUsageHandler) parseDateRange(r *http.Request) (time.Time, time.Time, error) {
	startStr := r.URL.Query().Get("start")
	endStr := r.URL.Query().Get("end")

	var start, end time.Time
	var err error

	if startStr != "" {
		start, err = time.Parse("2006-01-02", startStr)
		if err != nil {
			return time.Time{}, time.Time{}, err
		}
	} else {
		// Default to 30 days ago
		start = time.Now().UTC().AddDate(0, 0, -30)
	}

	if endStr != "" {
		end, err = time.Parse("2006-01-02", endStr)
		if err != nil {
			return time.Time{}, time.Time{}, err
		}
		// Set to end of day
		end = end.Add(24*time.Hour - time.Second)
	} else {
		end = time.Now().UTC()
	}

	return start, end, nil
}

// writeError writes a JSON error response
func writeError(w http.ResponseWriter, statusCode int, message string) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(statusCode)
	json.NewEncoder(w).Encode(map[string]interface{}{
		"error": message,
	})
}

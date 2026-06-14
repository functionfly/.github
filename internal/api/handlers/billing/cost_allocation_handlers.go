package billing

import (
	"encoding/json"
	"net/http"
	"strconv"
	"time"

	"github.com/functionfly/functionfly/internal/api/middleware"
	"github.com/functionfly/functionfly/internal/storage"
	"github.com/google/uuid"
	"github.com/gorilla/mux"
	"github.com/sirupsen/logrus"
)

// CostAllocationHandler handles detailed cost allocation API endpoints
type CostAllocationHandler struct {
	repo   storage.Repository
	logger *logrus.Logger
}

// NewCostAllocationHandler creates a new cost allocation handler
func NewCostAllocationHandler(repo storage.Repository) *CostAllocationHandler {
	return &CostAllocationHandler{
		repo:   repo,
		logger: logrus.New(),
	}
}

// GetCostSummary returns a comprehensive cost summary for the authenticated tenant
//
// GET /api/v1/costs/summary
//
// Query params:
//   - start: Start date (YYYY-MM-DD, default: 30 days ago)
//   - end: End date (YYYY-MM-DD, default: today)
//
// Response:
//
//	{
//	  "tenant_id": "uuid",
//	  "tenant_name": "My Org",
//	  "period_start": "2026-04-01T00:00:00Z",
//	  "period_end": "2026-04-09T00:00:00Z",
//	  "total_executions": 1500,
//	  "unique_functions": 25,
//	  "total_cost_cents": 12500,
//	  "total_cost_usd": 125.00,
//	  "cost_breakdown": {
//	    "execution_cost_cents": 5000,
//	    "compute_cost_cents": 6000,
//	    "platform_fee_cents": 1000,
//	    "data_transfer_cents": 500
//	  },
//	  "function_summaries": [...],
//	  "daily_breakdown": [...]
//	}
func (h *CostAllocationHandler) GetCostSummary(w http.ResponseWriter, r *http.Request) {
	tenantID := h.extractTenantID(r)
	if tenantID == uuid.Nil {
		h.writeError(w, http.StatusUnauthorized, "Unauthorized", "Tenant ID not found")
		return
	}

	start, end, err := h.parseDateRange(r)
	if err != nil {
		h.writeError(w, http.StatusBadRequest, "Invalid Request", err.Error())
		return
	}

	ctx := r.Context()
	summary, err := h.repo.GetTenantCostSummary(ctx, tenantID, start, end)
	if err != nil {
		h.logger.WithError(err).Error("Failed to get tenant cost summary")
		h.writeError(w, http.StatusInternalServerError, "Internal Error", "Failed to retrieve cost summary")
		return
	}

	// Get tenant name
	tenant, err := h.repo.GetTenantByID(ctx, tenantID)
	if err == nil && tenant != nil {
		summary.TenantName = tenant.Name
	}

	response := h.formatTenantCostSummary(summary)

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(response)
}

// GetCostByFunction returns cost breakdown by function
//
// GET /api/v1/costs/by-function
//
// Query params:
//   - start: Start date (YYYY-MM-DD, default: 30 days ago)
//   - end: End date (YYYY-MM-DD, default: today)
//
// Response:
//
//	{
//	  "tenant_id": "uuid",
//	  "period_start": "2026-04-01T00:00:00Z",
//	  "period_end": "2026-04-09T00:00:00Z",
//	  "functions": [
//	    {
//	      "function_id": "uuid",
//	      "function_name": "my-function",
//	      "function_author": "user@example.com",
//	      "total_executions": 100,
//	      "success_executions": 95,
//	      "error_executions": 5,
//	      "cached_executions": 10,
//	      "total_duration_ms": 500000,
//	      "avg_duration_ms": 5000.0,
//	      "total_cost_cents": 1500,
//	      "avg_cost_cents": 15.0,
//	      "cost_breakdown": {
//	        "execution": 500,
//	        "compute": 800,
//	        "platform_fee": 150,
//	        "data_transfer": 50
//	      }
//	    }
//	  ]
//	}
func (h *CostAllocationHandler) GetCostByFunction(w http.ResponseWriter, r *http.Request) {
	tenantID := h.extractTenantID(r)
	if tenantID == uuid.Nil {
		h.writeError(w, http.StatusUnauthorized, "Unauthorized", "Tenant ID not found")
		return
	}

	start, end, err := h.parseDateRange(r)
	if err != nil {
		h.writeError(w, http.StatusBadRequest, "Invalid Request", err.Error())
		return
	}

	ctx := r.Context()
	summaries, err := h.repo.GetCostAllocationByFunction(ctx, tenantID, start, end)
	if err != nil {
		h.logger.WithError(err).Error("Failed to get cost allocation by function")
		h.writeError(w, http.StatusInternalServerError, "Internal Error", "Failed to retrieve cost data")
		return
	}

	functions := make([]map[string]interface{}, len(summaries))
	for i, s := range summaries {
		functions[i] = map[string]interface{}{
			"function_id":         s.FunctionID.String(),
			"function_name":       s.FunctionName,
			"function_author":     s.FunctionAuthor,
			"total_executions":    s.TotalExecutions,
			"success_executions":  s.SuccessExecutions,
			"error_executions":    s.ErrorExecutions,
			"cached_executions":   s.CachedExecutions,
			"total_duration_ms":   s.TotalDurationMs,
			"avg_duration_ms":     s.AvgDurationMs,
			"total_cost_cents":    s.TotalCostCents,
			"total_cost_usd":      float64(s.TotalCostCents) / 100,
			"avg_cost_cents":      s.AvgCostCents,
			"avg_cost_usd":        s.AvgCostCents / 100,
			"cost_breakdown": map[string]int64{
				"execution":     s.ExecutionCostCents,
				"compute":       s.ComputeCostCents,
				"platform_fee":  s.PlatformFeeCents,
				"data_transfer": s.DataTransferCents,
			},
		}
	}

	response := map[string]interface{}{
		"tenant_id":    tenantID.String(),
		"period_start":   start,
		"period_end":     end,
		"function_count": len(functions),
		"functions":      functions,
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(response)
}

// GetCostByPeriod returns cost breakdown by time period (daily)
//
// GET /api/v1/costs/by-period
//
// Query params:
//   - start: Start date (YYYY-MM-DD, default: 30 days ago)
//   - end: End date (YYYY-MM-DD, default: today)
//   - granularity: "day" or "hour" (default: day)
//
// Response:
//
//	{
//	  "tenant_id": "uuid",
//	  "period_start": "2026-04-01T00:00:00Z",
//	  "period_end": "2026-04-09T00:00:00Z",
//	  "daily_breakdown": [
//	    {
//	      "date": "2026-04-01",
//	      "executions": 150,
//	      "cost_cents": 1250,
//	      "cost_usd": 12.50
//	    }
//	  ]
//	}
func (h *CostAllocationHandler) GetCostByPeriod(w http.ResponseWriter, r *http.Request) {
	tenantID := h.extractTenantID(r)
	if tenantID == uuid.Nil {
		h.writeError(w, http.StatusUnauthorized, "Unauthorized", "Tenant ID not found")
		return
	}

	start, end, err := h.parseDateRange(r)
	if err != nil {
		h.writeError(w, http.StatusBadRequest, "Invalid Request", err.Error())
		return
	}

	ctx := r.Context()
	breakdown, err := h.repo.GetCostAllocationDailyBreakdown(ctx, tenantID, start, end)
	if err != nil {
		h.logger.WithError(err).Error("Failed to get daily cost breakdown")
		h.writeError(w, http.StatusInternalServerError, "Internal Error", "Failed to retrieve cost data")
		return
	}

	daily := make([]map[string]interface{}, len(breakdown))
	for i, d := range breakdown {
		daily[i] = map[string]interface{}{
			"date":       d.Date.Format("2006-01-02"),
			"executions": d.Executions,
			"cost_cents": d.CostCents,
			"cost_usd":   float64(d.CostCents) / 100,
		}
	}

	response := map[string]interface{}{
		"tenant_id":       tenantID.String(),
		"period_start":    start,
		"period_end":      end,
		"daily_breakdown": daily,
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(response)
}

// GetCostByRegion returns cost breakdown by region
//
// GET /api/v1/costs/by-region
//
// Query params:
//   - start: Start date (YYYY-MM-DD, default: 30 days ago)
//   - end: End date (YYYY-MM-DD, default: today)
//
// Response:
//
//	{
//	  "tenant_id": "uuid",
//	  "period_start": "2026-04-01T00:00:00Z",
//	  "period_end": "2026-04-09T00:00:00Z",
//	  "regions": {
//	    "us-east-1": {
//	      "total_executions": 1000,
//	      "total_cost_cents": 8500,
//	      "cost_breakdown": { ... }
//	    },
//	    "eu-west-1": {
//	      "total_executions": 500,
//	      "total_cost_cents": 4000,
//	      "cost_breakdown": { ... }
//	    }
//	  }
//	}
func (h *CostAllocationHandler) GetCostByRegion(w http.ResponseWriter, r *http.Request) {
	tenantID := h.extractTenantID(r)
	if tenantID == uuid.Nil {
		h.writeError(w, http.StatusUnauthorized, "Unauthorized", "Tenant ID not found")
		return
	}

	start, end, err := h.parseDateRange(r)
	if err != nil {
		h.writeError(w, http.StatusBadRequest, "Invalid Request", err.Error())
		return
	}

	ctx := r.Context()
	regions, err := h.repo.GetCostAllocationByRegion(ctx, tenantID, start, end)
	if err != nil {
		h.logger.WithError(err).Error("Failed to get cost allocation by region")
		h.writeError(w, http.StatusInternalServerError, "Internal Error", "Failed to retrieve cost data")
		return
	}

	regionsMap := make(map[string]interface{})
	for region, summary := range regions {
		regionsMap[region] = map[string]interface{}{
			"total_executions": summary.TotalExecutions,
			"total_cost_cents": summary.TotalCostCents,
			"total_cost_usd":   float64(summary.TotalCostCents) / 100,
			"cost_breakdown": map[string]int64{
				"execution":     summary.ExecutionCostCents,
				"compute":       summary.ComputeCostCents,
				"platform_fee":  summary.PlatformFeeCents,
				"data_transfer": summary.DataTransferCents,
			},
		}
	}

	response := map[string]interface{}{
		"tenant_id":    tenantID.String(),
		"period_start": start,
		"period_end":   end,
		"regions":      regionsMap,
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(response)
}

// GetCostEntries returns detailed cost allocation entries
//
// GET /api/v1/costs/entries
//
// Query params:
//   - start: Start date (YYYY-MM-DD)
//   - end: End date (YYYY-MM-DD)
//   - function_id: Filter by function ID
//   - outcome: Filter by outcome (success, error)
//   - cached: Filter by cached (true, false)
//   - limit: Page size (default: 50, max: 200)
//   - offset: Pagination offset
//
// Response: Paginated list of cost allocation entries
func (h *CostAllocationHandler) GetCostEntries(w http.ResponseWriter, r *http.Request) {
	tenantID := h.extractTenantID(r)
	if tenantID == uuid.Nil {
		h.writeError(w, http.StatusUnauthorized, "Unauthorized", "Tenant ID not found")
		return
	}

	// Build filter from query params
	filter := &storage.CostAllocationFilter{
		TenantID: &tenantID,
	}

	if startStr := r.URL.Query().Get("start"); startStr != "" {
		if start, err := time.Parse("2006-01-02", startStr); err == nil {
			filter.StartDate = &start
		}
	}

	if endStr := r.URL.Query().Get("end"); endStr != "" {
		if end, err := time.Parse("2006-01-02", endStr); err == nil {
			// Set to end of day
			end = end.Add(24*time.Hour - time.Second)
			filter.EndDate = &end
		}
	}

	if fnIDStr := r.URL.Query().Get("function_id"); fnIDStr != "" {
		if fnID, err := uuid.Parse(fnIDStr); err == nil {
			filter.FunctionID = &fnID
		}
	}

	if outcome := r.URL.Query().Get("outcome"); outcome != "" {
		filter.Outcome = &outcome
	}

	if cachedStr := r.URL.Query().Get("cached"); cachedStr != "" {
		cached := cachedStr == "true"
		filter.Cached = &cached
	}

	if region := r.URL.Query().Get("region"); region != "" {
		filter.Region = &region
	}

	// Pagination
	limit := 50
	if l := r.URL.Query().Get("limit"); l != "" {
		if parsed, err := strconv.Atoi(l); err == nil && parsed > 0 && parsed <= 200 {
			limit = parsed
		}
	}

	offset := 0
	if o := r.URL.Query().Get("offset"); o != "" {
		if parsed, err := strconv.Atoi(o); err == nil && parsed >= 0 {
			offset = parsed
		}
	}

	ctx := r.Context()
	entries, total, err := h.repo.GetCostAllocationEntries(ctx, filter, limit, offset)
	if err != nil {
		h.logger.WithError(err).Error("Failed to get cost allocation entries")
		h.writeError(w, http.StatusInternalServerError, "Internal Error", "Failed to retrieve cost entries")
		return
	}

	// Format entries
	formattedEntries := make([]map[string]interface{}, len(entries))
	for i, e := range entries {
		formattedEntries[i] = map[string]interface{}{
			"id":                 e.ID.String(),
			"function_id":        e.FunctionID.String(),
			"function_name":      e.FunctionName,
			"function_author":    e.FunctionAuthor,
			"execution_id":       e.ExecutionID.String(),
			"execution_outcome":  e.ExecutionOutcome,
			"cached":             e.Cached,
			"duration_ms":        e.DurationMs,
			"cpu_time_ms":        e.CPUTimeMs,
			"memory_used_mb":     e.MemoryUsedMB,
			"wall_time_ms":       e.WallTimeMs,
			"execution_cost_usd": float64(e.ExecutionCostCents) / 100,
			"compute_cost_usd":   float64(e.ComputeCostCents) / 100,
			"platform_fee_usd":   float64(e.PlatformFeeCents) / 100,
			"data_transfer_usd":  float64(e.DataTransferCents) / 100,
			"total_cost_usd":     float64(e.TotalCostCents) / 100,
			"region":             e.Region,
			"timestamp":          e.Timestamp,
			"tags":               e.Tags,
			"metadata":           e.Metadata,
		}
	}

	response := map[string]interface{}{
		"entries":      formattedEntries,
		"total":        total,
		"limit":        limit,
		"offset":       offset,
		"has_more":     offset+len(entries) < total,
		"tenant_id":    tenantID.String(),
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(response)
}

// GetChargebackReport returns chargeback data for internal billing
//
// GET /api/v1/costs/chargeback
//
// Query params:
//   - start: Start date (YYYY-MM-DD)
//   - end: End date (YYYY-MM-DD)
//
// Response:
//
//	{
//	  "report_id": "uuid",
//	  "generated_at": "2026-04-09T12:00:00Z",
//	  "period_start": "2026-04-01T00:00:00Z",
//	  "period_end": "2026-04-09T00:00:00Z",
//	  "tenant_count": 50,
//	  "total_cost_usd": 12500.00,
//	  "chargeback_entries": [
//	    {
//	      "tenant_id": "uuid",
//	      "tenant_name": "My Org",
//	      "cost_center": "Engineering",
//	      "department": "Platform",
//	      "total_cost_usd": 250.00,
//	      "cost_breakdown": { ... }
//	    }
//	  ]
//	}
func (h *CostAllocationHandler) GetChargebackReport(w http.ResponseWriter, r *http.Request) {
	// This endpoint is for admin/internal use
	start, end, err := h.parseDateRange(r)
	if err != nil {
		h.writeError(w, http.StatusBadRequest, "Invalid Request", err.Error())
		return
	}

	ctx := r.Context()
	report, err := h.repo.GetCostAllocationReport(ctx, start, end)
	if err != nil {
		h.logger.WithError(err).Error("Failed to get cost allocation report")
		h.writeError(w, http.StatusInternalServerError, "Internal Error", "Failed to generate report")
		return
	}

	// Format chargeback entries
	chargebackEntries := make([]map[string]interface{}, len(report.ChargebackEntries))
	for i, c := range report.ChargebackEntries {
		chargebackEntries[i] = map[string]interface{}{
			"tenant_id":         c.TenantID.String(),
			"tenant_name":       c.TenantName,
			"cost_center":       c.CostCenter,
			"department":        c.Department,
			"project":           c.Project,
			"total_cost_usd":    float64(c.TotalCostCents) / 100,
			"cost_breakdown": map[string]float64{
				"execution":     float64(c.ExecutionCostCents) / 100,
				"compute":       float64(c.ComputeCostCents) / 100,
				"platform_fee":  float64(c.PlatformFeeCents) / 100,
				"data_transfer": float64(c.DataTransferCents) / 100,
			},
			"invoice_period": c.InvoicePeriod,
			"generated_at":   c.GeneratedAt,
		}
	}

	response := map[string]interface{}{
		"report_id":          report.ReportID.String(),
		"generated_at":       report.GeneratedAt,
		"period_start":       report.PeriodStart,
		"period_end":         report.PeriodEnd,
		"tenant_count":       report.TenantCount,
		"function_count":     report.FunctionCount,
		"total_executions":   report.TotalExecutions,
		"total_cost_usd":     float64(report.TotalCostCents) / 100,
		"chargeback_entries": chargebackEntries,
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(response)
}

// AdminGetTenantCostSummary returns cost summary for any tenant (admin only)
//
// GET /api/v1/admin/costs/{tenant_id}/summary
func (h *CostAllocationHandler) AdminGetTenantCostSummary(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	tenantIDStr := vars["tenant_id"]

	tenantID, err := uuid.Parse(tenantIDStr)
	if err != nil {
		h.writeError(w, http.StatusBadRequest, "Invalid Request", "Invalid tenant ID")
		return
	}

	start, end, err := h.parseDateRange(r)
	if err != nil {
		h.writeError(w, http.StatusBadRequest, "Invalid Request", err.Error())
		return
	}

	ctx := r.Context()
	summary, err := h.repo.GetTenantCostSummary(ctx, tenantID, start, end)
	if err != nil {
		h.logger.WithError(err).Error("Failed to get tenant cost summary")
		h.writeError(w, http.StatusInternalServerError, "Internal Error", "Failed to retrieve cost summary")
		return
	}

	// Get tenant name
	tenant, err := h.repo.GetTenantByID(ctx, tenantID)
	if err == nil && tenant != nil {
		summary.TenantName = tenant.Name
	}

	response := h.formatTenantCostSummary(summary)

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(response)
}

// Helper methods

func (h *CostAllocationHandler) extractTenantID(r *http.Request) uuid.UUID {
	if tenantID, ok := r.Context().Value("tenant_id").(uuid.UUID); ok {
		return tenantID
	}

	if user, ok := r.Context().Value("user").(*storage.User); ok && user.TenantID != uuid.Nil {
		return user.TenantID
	}

	if claims := middleware.GetUserFromContext(r); claims != nil && claims.TenantID != uuid.Nil {
		return claims.TenantID
	}

	return uuid.Nil
}

func (h *CostAllocationHandler) parseDateRange(r *http.Request) (time.Time, time.Time, error) {
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

func (h *CostAllocationHandler) formatTenantCostSummary(summary *storage.TenantCostSummary) map[string]interface{} {
	// Format function summaries
	functions := make([]map[string]interface{}, len(summary.FunctionSummaries))
	for i, f := range summary.FunctionSummaries {
		functions[i] = map[string]interface{}{
			"function_id":         f.FunctionID.String(),
			"function_name":       f.FunctionName,
			"function_author":     f.FunctionAuthor,
			"total_executions":    f.TotalExecutions,
			"success_executions":  f.SuccessExecutions,
			"error_executions":    f.ErrorExecutions,
			"cached_executions":   f.CachedExecutions,
			"total_duration_ms":   f.TotalDurationMs,
			"avg_duration_ms":     f.AvgDurationMs,
			"total_cost_cents":    f.TotalCostCents,
			"total_cost_usd":      float64(f.TotalCostCents) / 100,
			"avg_cost_cents":      f.AvgCostCents,
			"avg_cost_usd":        f.AvgCostCents / 100,
		}
	}

	// Format daily breakdown
	daily := make([]map[string]interface{}, len(summary.DailyBreakdown))
	for i, d := range summary.DailyBreakdown {
		daily[i] = map[string]interface{}{
			"date":       d.Date.Format("2006-01-02"),
			"executions": d.Executions,
			"cost_cents": d.CostCents,
			"cost_usd":   float64(d.CostCents) / 100,
		}
	}

	return map[string]interface{}{
		"tenant_id":          summary.TenantID.String(),
		"tenant_name":        summary.TenantName,
		"period_start":       summary.PeriodStart,
		"period_end":         summary.PeriodEnd,
		"total_executions":   summary.TotalExecutions,
		"unique_functions":   summary.UniqueFunctions,
		"total_cost_cents":   summary.TotalCostCents,
		"total_cost_usd":     float64(summary.TotalCostCents) / 100,
		"cost_breakdown": map[string]float64{
			"execution":     float64(summary.ExecutionCostCents) / 100,
			"compute":       float64(summary.ComputeCostCents) / 100,
			"platform_fee":  float64(summary.PlatformFeeCents) / 100,
			"data_transfer": float64(summary.DataTransferCents) / 100,
		},
		"function_summaries": functions,
		"daily_breakdown":    daily,
	}
}

func (h *CostAllocationHandler) writeError(w http.ResponseWriter, statusCode int, code, message string) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(statusCode)
	json.NewEncoder(w).Encode(map[string]interface{}{
		"error": map[string]interface{}{
			"code":    code,
			"message": message,
		},
	})
}

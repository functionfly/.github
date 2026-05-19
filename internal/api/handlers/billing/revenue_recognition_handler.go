package billing

import (
	"encoding/json"
	"net/http"
	"time"

	"github.com/functionfly/functionfly/internal/api/middleware"
	"github.com/functionfly/functionfly/internal/apierror"
	"github.com/functionfly/functionfly/internal/services"
	"github.com/google/uuid"
)

type RevenueRecognitionHandler struct {
	svc *services.RevenueRecognitionService
}

func NewRevenueRecognitionHandler(svc *services.RevenueRecognitionService) *RevenueRecognitionHandler {
	return &RevenueRecognitionHandler{svc: svc}
}

type DeferredRevenueResponse struct {
	TenantID            uuid.UUID `json:"tenant_id"`
	Period              string    `json:"period"`
	OpeningBalanceCents int       `json:"opening_balance_cents"`
	NewDeferredCents    int       `json:"new_deferred_cents"`
	RecognizedCents     int       `json:"recognized_cents"`
	ClosingBalanceCents int       `json:"closing_balance_cents"`
}

type RecognizedRevenueResponse struct {
	TenantID          uuid.UUID `json:"tenant_id"`
	Period            string    `json:"period"`
	SubscriptionCents int       `json:"subscription_revenue_cents"`
	UsageCents        int       `json:"usage_revenue_cents"`
	OneTimeCents      int       `json:"one_time_revenue_cents"`
	TotalCents        int       `json:"total_revenue_cents"`
}

type RevenueReportResponse struct {
	ReportID             uuid.UUID `json:"report_id"`
	Period               string    `json:"period"`
	TotalRevenueCents    int       `json:"total_revenue_cents"`
	TotalDeferredCents   int       `json:"total_deferred_cents"`
	TotalRecognizedCents int       `json:"total_recognized_cents"`
	OpeningDeferredCents int       `json:"opening_deferred_cents"`
	NewDeferredCents     int       `json:"new_deferred_cents"`
	RecognizedCents      int       `json:"recognized_from_deferred_cents"`
	ClosingDeferredCents int       `json:"closing_deferred_cents"`
	OverTimeRevenueCents int       `json:"over_time_revenue_cents"`
	PointInTimeCents     int       `json:"point_in_time_revenue_cents"`
}

type RecognizeScheduleRequest struct {
	ScheduleID string `json:"schedule_id"`
}

type AllocationRequestInput struct {
	InvoiceID          string          `json:"invoice_id"`
	InvoiceAmountCents int             `json:"invoice_amount_cents"`
	Currency           string          `json:"currency"`
	LineItems          []LineItemInput `json:"line_items"`
}

type LineItemInput struct {
	Description       string `json:"description"`
	AmountCents       int    `json:"amount_cents"`
	RevenueType       string `json:"revenue_type"`
	SSPCents          int    `json:"ssp_cents"`
	RecognitionMethod string `json:"recognition_method"`
	StartDate         string `json:"start_date"`
	EndDate           string `json:"end_date"`
	DeliveryPattern   string `json:"delivery_pattern"`
}

type AllocationResponse struct {
	PerformanceObligationIDs []uuid.UUID `json:"performance_obligation_ids"`
	ScheduleCount            int         `json:"schedule_count"`
}

func (h *RevenueRecognitionHandler) HandleGetDeferredRevenue(w http.ResponseWriter, r *http.Request) {
	claims := middleware.GetUserFromContext(r)
	if claims == nil || claims.TenantID == uuid.Nil {
		apierror.WriteError(w, apierror.NewUnauthorized("Authentication required"))
		return
	}

	period := r.URL.Query().Get("period")
	if period == "" {
		period = time.Now().Format("2006-01")
	}

	summary, err := h.svc.GetDeferredRevenueSummary(r.Context(), claims.TenantID, period)
	if err != nil {
		apierror.WriteError(w, apierror.NewInternal("Failed to get deferred revenue summary"))
		return
	}

	response := DeferredRevenueResponse{
		TenantID:            claims.TenantID,
		Period:              period,
		OpeningBalanceCents: summary.OpeningBalanceCents,
		NewDeferredCents:    summary.NewDeferredCents,
		RecognizedCents:     summary.RecognizedCents,
		ClosingBalanceCents: summary.ClosingBalanceCents,
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(response)
}

func (h *RevenueRecognitionHandler) HandleGetRecognizedRevenue(w http.ResponseWriter, r *http.Request) {
	claims := middleware.GetUserFromContext(r)
	if claims == nil || claims.TenantID == uuid.Nil {
		apierror.WriteError(w, apierror.NewUnauthorized("Authentication required"))
		return
	}

	period := r.URL.Query().Get("period")
	if period == "" {
		period = time.Now().Format("2006-01")
	}

	summary, err := h.svc.GetRecognizedRevenueSummary(r.Context(), claims.TenantID, period)
	if err != nil {
		apierror.WriteError(w, apierror.NewInternal("Failed to get recognized revenue summary"))
		return
	}

	response := RecognizedRevenueResponse{
		TenantID:          claims.TenantID,
		Period:            period,
		SubscriptionCents: summary.SubscriptionRevenueCents,
		UsageCents:        summary.UsageRevenueCents,
		OneTimeCents:      summary.OneTimeRevenueCents,
		TotalCents:        summary.TotalRevenueCents,
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(response)
}

func (h *RevenueRecognitionHandler) HandleGetRevenueReport(w http.ResponseWriter, r *http.Request) {
	claims := middleware.GetUserFromContext(r)
	if claims == nil || claims.TenantID == uuid.Nil {
		apierror.WriteError(w, apierror.NewUnauthorized("Authentication required"))
		return
	}

	period := r.URL.Query().Get("period")
	if period == "" {
		period = time.Now().Format("2006-01")
	}

	report, err := h.svc.ProcessRecognition(r.Context(), claims.TenantID, period)
	if err != nil {
		apierror.WriteError(w, apierror.NewInternal("Failed to generate revenue report"))
		return
	}

	response := RevenueReportResponse{
		ReportID:             report.ReportID,
		Period:               period,
		TotalRevenueCents:    report.TotalRevenueCents,
		TotalDeferredCents:   report.TotalDeferredCents,
		TotalRecognizedCents: report.TotalRecognizedCents,
		OpeningDeferredCents: report.OpeningDeferredCents,
		NewDeferredCents:     report.NewDeferredCents,
		RecognizedCents:      report.RecognizedFromDeferredCents,
		ClosingDeferredCents: report.ClosingDeferredCents,
		OverTimeRevenueCents: report.OverTimeRevenueCents,
		PointInTimeCents:     report.PointInTimeRevenueCents,
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(response)
}

func (h *RevenueRecognitionHandler) HandleRecognizeRevenue(w http.ResponseWriter, r *http.Request) {
	claims := middleware.GetUserFromContext(r)
	if claims == nil || claims.TenantID == uuid.Nil {
		apierror.WriteError(w, apierror.NewUnauthorized("Authentication required"))
		return
	}

	var req RecognizeScheduleRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		apierror.WriteError(w, apierror.NewBadRequest("Invalid request body"))
		return
	}

	scheduleID, err := uuid.Parse(req.ScheduleID)
	if err != nil {
		apierror.WriteError(w, apierror.NewValidation("Invalid schedule ID format"))
		return
	}

	if err := h.svc.RecognizeRevenue(r.Context(), scheduleID); err != nil {
		if err.Error() == "schedule already recognized" {
			apierror.WriteError(w, apierror.NewBadRequest("Schedule already recognized"))
			return
		}
		apierror.WriteError(w, apierror.NewInternal("Failed to recognize revenue"))
		return
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(map[string]string{"status": "recognized"})
}

func (h *RevenueRecognitionHandler) HandleAllocateRevenue(w http.ResponseWriter, r *http.Request) {
	claims := middleware.GetUserFromContext(r)
	if claims == nil || claims.TenantID == uuid.Nil {
		apierror.WriteError(w, apierror.NewUnauthorized("Authentication required"))
		return
	}

	var req AllocationRequestInput
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		apierror.WriteError(w, apierror.NewBadRequest("Invalid request body"))
		return
	}

	invoiceID, err := uuid.Parse(req.InvoiceID)
	if err != nil {
		apierror.WriteError(w, apierror.NewValidation("Invalid invoice ID format"))
		return
	}

	if req.InvoiceAmountCents <= 0 {
		apierror.WriteError(w, apierror.NewValidation("Amount must be positive"))
		return
	}

	if len(req.LineItems) == 0 {
		apierror.WriteError(w, apierror.NewValidation("At least one line item is required"))
		return
	}

	lineItems := make([]services.LineItem, 0, len(req.LineItems))
	for _, item := range req.LineItems {
		startDate, err := time.Parse(time.RFC3339, item.StartDate)
		if err != nil {
			apierror.WriteError(w, apierror.NewValidation("Invalid start date format"))
			return
		}

		var endDate *time.Time
		if item.EndDate != "" {
			parsed, err := time.Parse(time.RFC3339, item.EndDate)
			if err != nil {
				apierror.WriteError(w, apierror.NewValidation("Invalid end date format"))
				return
			}
			endDate = &parsed
		}

		lineItems = append(lineItems, services.LineItem{
			Description:          item.Description,
			AmountCents:          item.AmountCents,
			RevenueType:          item.RevenueType,
			SSPCents:             item.SSPCents,
			RecognitionMethod:    item.RecognitionMethod,
			RecognitionStartDate: startDate,
			RecognitionEndDate:   endDate,
			DeliveryPattern:      item.DeliveryPattern,
		})
	}

	allocReq := &services.AllocationRequest{
		TenantID:           claims.TenantID,
		InvoiceID:          invoiceID,
		InvoiceAmountCents: req.InvoiceAmountCents,
		Currency:           req.Currency,
		LineItems:          lineItems,
	}

	if err := h.svc.AllocateAndSchedule(r.Context(), allocReq); err != nil {
		apierror.WriteError(w, apierror.NewInternal("Failed to allocate revenue"))
		return
	}

	obligationIDs, err := h.svc.GetPerformanceObligationIDsByInvoice(r.Context(), invoiceID)
	if err != nil {
		obligationIDs = []uuid.UUID{}
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(AllocationResponse{
		PerformanceObligationIDs: obligationIDs,
		ScheduleCount:            len(lineItems),
	})
}

func (h *RevenueRecognitionHandler) HandleGetUnbilledRevenue(w http.ResponseWriter, r *http.Request) {
	claims := middleware.GetUserFromContext(r)
	if claims == nil || claims.TenantID == uuid.Nil {
		apierror.WriteError(w, apierror.NewUnauthorized("Authentication required"))
		return
	}

	unbilled, err := h.svc.GetUnbilledRevenue(r.Context(), claims.TenantID)
	if err != nil {
		apierror.WriteError(w, apierror.NewInternal("Failed to get unbilled revenue"))
		return
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(map[string]interface{}{
		"tenant_id":              claims.TenantID,
		"unbilled_revenue_cents": unbilled,
	})
}

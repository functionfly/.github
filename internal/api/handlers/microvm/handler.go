package microvm

import (
	"context"
	"database/sql"
	"encoding/json"
	"net/http"
	"strconv"
	"time"

	"github.com/functionfly/functionfly/internal/plans"
	"github.com/functionfly/functionfly/internal/storage"
	"github.com/google/uuid"
	"github.com/gorilla/mux"
	"github.com/sirupsen/logrus"
)

type Handler struct {
	microvmRepo *storage.MicroVMRepository
	repo        storage.Repository
}

func NewHandler(microvmRepo *storage.MicroVMRepository, repo storage.Repository) *Handler {
	return &Handler{
		microvmRepo: microvmRepo,
		repo:        repo,
	}
}

type UsageResponse struct {
	Stats       *storage.MicroVMStats       `json:"stats"`
	Quota       *storage.MicroVMTenantQuota `json:"quota"`
	Executions  []*storage.MicroVMExecution `json:"executions"`
	CurrentPlan string                     `json:"current_plan"`
}

type BillingResponse struct {
	Current  *storage.MicroVMBillingRecord   `json:"current"`
	History  []*storage.MicroVMBillingRecord `json:"history"`
	Limits   *plans.MicroVMLimits            `json:"limits"`
}

type AuditResponse struct {
	Logs  []*storage.MicroVMAuditLog `json:"logs"`
	Total int                        `json:"total"`
}

func (h *Handler) GetUsage(w http.ResponseWriter, r *http.Request) {
	tenantID := h.getTenantID(r.Context())
	if tenantID == uuid.Nil {
		http.Error(w, "unauthorized", http.StatusUnauthorized)
		return
	}

	plan := h.getTenantPlan(r.Context(), tenantID)
	if plan != plans.PlanEnterprise {
		http.Error(w, "MicroVMs are only available for Enterprise plan", http.StatusForbidden)
		return
	}

	now := time.Now()
	startOfMonth := time.Date(now.Year(), now.Month(), 1, 0, 0, 0, 0, time.UTC)

	stats, err := h.microvmRepo.GetUsageStats(r.Context(), tenantID, startOfMonth, now)
	if err != nil {
		logrus.WithError(err).Error("failed to get MicroVM usage stats")
		http.Error(w, "failed to get usage stats", http.StatusInternalServerError)
		return
	}

	quota, err := h.microvmRepo.GetTenantQuota(r.Context(), tenantID)
	if err != nil {
		logrus.WithError(err).Error("failed to get MicroVM quota")
	}

	executions, err := h.microvmRepo.GetExecutionsByTenant(r.Context(), tenantID, 50, 0)
	if err != nil {
		logrus.WithError(err).Error("failed to get MicroVM executions")
		executions = []*storage.MicroVMExecution{}
	}

	respondJSON(w, http.StatusOK, UsageResponse{
		Stats:       stats,
		Quota:       quota,
		Executions:  executions,
		CurrentPlan: plan,
	})
}

func (h *Handler) GetBilling(w http.ResponseWriter, r *http.Request) {
	tenantID := h.getTenantID(r.Context())
	if tenantID == uuid.Nil {
		http.Error(w, "unauthorized", http.StatusUnauthorized)
		return
	}

	plan := h.getTenantPlan(r.Context(), tenantID)
	if plan != plans.PlanEnterprise {
		http.Error(w, "MicroVMs are only available for Enterprise plan", http.StatusForbidden)
		return
	}

	currentPeriod := time.Now().Format("2006-01")

	billing, err := h.microvmRepo.GetBillingRecord(r.Context(), tenantID, currentPeriod)
	if err != nil {
		logrus.WithError(err).Error("failed to get MicroVM billing record")
	}

	history, err := h.microvmRepo.GetBillingHistory(r.Context(), tenantID, 12, 0)
	if err != nil {
		logrus.WithError(err).Error("failed to get MicroVM billing history")
		history = []*storage.MicroVMBillingRecord{}
	}

	limits := plans.GetMicroVMLimits(plan)

	respondJSON(w, http.StatusOK, BillingResponse{
		Current: billing,
		History: history,
		Limits:  limits,
	})
}

func (h *Handler) GetAuditLog(w http.ResponseWriter, r *http.Request) {
	tenantID := h.getTenantID(r.Context())
	if tenantID == uuid.Nil {
		http.Error(w, "unauthorized", http.StatusUnauthorized)
		return
	}

	plan := h.getTenantPlan(r.Context(), tenantID)
	if plan != plans.PlanEnterprise {
		http.Error(w, "MicroVMs are only available for Enterprise plan", http.StatusForbidden)
		return
	}

	limit, _ := strconv.Atoi(r.URL.Query().Get("limit"))
	if limit <= 0 || limit > 100 {
		limit = 50
	}
	offset, _ := strconv.Atoi(r.URL.Query().Get("offset"))

	logs, err := h.microvmRepo.GetAuditLog(r.Context(), tenantID, limit, offset)
	if err != nil {
		logrus.WithError(err).Error("failed to get MicroVM audit log")
		http.Error(w, "failed to get audit log", http.StatusInternalServerError)
		return
	}

	respondJSON(w, http.StatusOK, AuditResponse{
		Logs:  logs,
		Total: len(logs),
	})
}

func (h *Handler) GetQuota(w http.ResponseWriter, r *http.Request) {
	tenantID := h.getTenantID(r.Context())
	if tenantID == uuid.Nil {
		http.Error(w, "unauthorized", http.StatusUnauthorized)
		return
	}

	plan := h.getTenantPlan(r.Context(), tenantID)
	if plan != plans.PlanEnterprise {
		http.Error(w, "MicroVMs are only available for Enterprise plan", http.StatusForbidden)
		return
	}

	quota, err := h.microvmRepo.GetTenantQuota(r.Context(), tenantID)
	if err != nil {
		logrus.WithError(err).Error("failed to get MicroVM quota")
		http.Error(w, "failed to get quota", http.StatusInternalServerError)
		return
	}

	if quota == nil {
		limits := plans.GetMicroVMLimits(plan)
		quota = &storage.MicroVMTenantQuota{
			TenantID:           tenantID,
			MaxConcurrentVMs:   limits.MaxMicroVMs,
			MaxMemoryMB:        limits.MaxMemoryMB,
			MaxVCPUs:           limits.MaxVCPU,
			MaxTimeoutMs:       limits.MaxTimeout,
			CurrentComputeUsage: 0,
			CurrentMemoryUsage:  0,
			UpdatedAt:          time.Now(),
		}
		if err := h.microvmRepo.UpsertTenantQuota(r.Context(), quota); err != nil {
			logrus.WithError(err).Error("failed to create default MicroVM quota")
		}
	}

	respondJSON(w, http.StatusOK, quota)
}

func (h *Handler) CreateExecutionRecord(w http.ResponseWriter, r *http.Request) {
	tenantID := h.getTenantID(r.Context())
	if tenantID == uuid.Nil {
		http.Error(w, "unauthorized", http.StatusUnauthorized)
		return
	}

	var req struct {
		FunctionID      uuid.UUID `json:"function_id"`
		FunctionVersion string    `json:"function_version"`
		ExecutionID     uuid.UUID `json:"execution_id"`
		MemoryMB        int       `json:"memory_mb"`
		VCPUs           int       `json:"vcpus"`
		NetworkAllowed  bool      `json:"network_allowed"`
		PackagesCached  bool      `json:"packages_cached"`
	}

	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, "invalid request body", http.StatusBadRequest)
		return
	}

	exec := &storage.MicroVMExecution{
		ID:              uuid.New(),
		TenantID:        tenantID,
		FunctionID:      req.FunctionID,
		FunctionVersion: req.FunctionVersion,
		ExecutionID:     req.ExecutionID,
		StartedAt:       time.Now(),
		MemoryMB:        req.MemoryMB,
		VCPUs:           req.VCPUs,
		Status:          "running",
		NetworkAllowed:  req.NetworkAllowed,
		PackagesCached:  req.PackagesCached,
		CreatedAt:       time.Now(),
	}

	if err := h.microvmRepo.CreateExecution(r.Context(), exec); err != nil {
		logrus.WithError(err).Error("failed to create MicroVM execution record")
		http.Error(w, "failed to create execution record", http.StatusInternalServerError)
		return
	}

	h.logAudit(r.Context(), tenantID, "execution_started", "execution", &exec.ID, map[string]interface{}{
		"function_id":      req.FunctionID,
		"function_version": req.FunctionVersion,
		"execution_id":    req.ExecutionID,
	})

	respondJSON(w, http.StatusCreated, exec)
}

func (h *Handler) UpdateExecutionStatus(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	execID, err := uuid.Parse(vars["id"])
	if err != nil {
		http.Error(w, "invalid execution ID", http.StatusBadRequest)
		return
	}

	tenantID := h.getTenantID(r.Context())
	if tenantID == uuid.Nil {
		http.Error(w, "unauthorized", http.StatusUnauthorized)
		return
	}

	var req struct {
		Status      string  `json:"status"`
		Outcome     *string `json:"outcome"`
		ErrorMessage *string `json:"error_message"`
		DurationMs  int     `json:"duration_ms"`
	}

	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, "invalid request body", http.StatusBadRequest)
		return
	}

	completedAt := time.Now()
	if req.Status == "completed" || req.Status == "failed" || req.Status == "timeout" {
		if err := h.microvmRepo.UpdateExecutionStatus(r.Context(), execID, req.Status, req.Outcome, req.ErrorMessage, completedAt, req.DurationMs); err != nil {
			logrus.WithError(err).Error("failed to update MicroVM execution status")
			http.Error(w, "failed to update execution status", http.StatusInternalServerError)
			return
		}

		h.logAudit(r.Context(), tenantID, "execution_"+req.Status, "execution", &execID, map[string]interface{}{
			"outcome": req.Outcome,
			"duration_ms": req.DurationMs,
		})
	}

	w.WriteHeader(http.StatusNoContent)
}

func (h *Handler) AggregateBilling(w http.ResponseWriter, r *http.Request) {
	tenantID := h.getTenantID(r.Context())
	if tenantID == uuid.Nil {
		http.Error(w, "unauthorized", http.StatusUnauthorized)
		return
	}

	billingPeriod := time.Now().Format("2006-01")

	record, err := h.microvmRepo.AggregateUsageForBilling(r.Context(), tenantID, billingPeriod)
	if err != nil {
		logrus.WithError(err).Error("failed to aggregate MicroVM billing")
		http.Error(w, "failed to aggregate billing", http.StatusInternalServerError)
		return
	}

	billing := plans.CalculateMicroVMBilling(
		plans.PlanEnterprise,
		record.TotalExecutions,
		record.TotalComputeSeconds,
		record.AvgMemoryMB,
		record.TotalMemorySeconds,
	)

	if billing != nil {
		record.BaseFeeCents = billing.BaseFeeMonthly
		record.ComputeChargeCents = billing.ComputeCharges
		record.MemoryChargeCents = billing.MemoryCharges
		record.TotalChargeCents = billing.TotalCents
	}

	if err := h.microvmRepo.CreateBillingRecord(r.Context(), record); err != nil {
		logrus.WithError(err).Error("failed to create MicroVM billing record")
	}

	respondJSON(w, http.StatusOK, record)
}

func (h *Handler) getTenantID(ctx context.Context) uuid.UUID {
	if user, ok := ctx.Value("user").(struct {
		TenantID uuid.UUID
	}); ok {
		return user.TenantID
	}
	return uuid.Nil
}

func (h *Handler) getTenantPlan(ctx context.Context, tenantID uuid.UUID) string {
	tenant, err := h.repo.GetTenantByID(ctx, tenantID)
	if err != nil || tenant == nil {
		return ""
	}
	return tenant.Plan
}

func (h *Handler) logAudit(ctx context.Context, tenantID uuid.UUID, action, resourceType string, resourceID *uuid.UUID, details map[string]interface{}) {
	detailsJSON, _ := json.Marshal(details)

	var userID uuid.NullUUID
	if user, ok := ctx.Value("user").(struct {
		ID uuid.UUID
	}); ok {
		userID = uuid.NullUUID{UUID: user.ID, Valid: true}
	}

	var ipAddr sql.NullString
	if ip, ok := ctx.Value("ip_address").(string); ok {
		ipAddr = sql.NullString{String: ip, Valid: true}
	}

	var userAgent sql.NullString
	if ua, ok := ctx.Value("user_agent").(string); ok {
		userAgent = sql.NullString{String: ua, Valid: true}
	}

	auditLog := &storage.MicroVMAuditLog{
		ID:           uuid.New(),
		TenantID:     tenantID,
		UserID:       userID,
		Action:       action,
		ResourceType: resourceType,
		ResourceID:   uuid.NullUUID{UUID: *resourceID, Valid: resourceID != nil},
		Details:      detailsJSON,
		IPAddress:    ipAddr,
		UserAgent:    userAgent,
		CreatedAt:    time.Now(),
	}

	go func() {
		if err := h.microvmRepo.CreateAuditLog(context.Background(), auditLog); err != nil {
			logrus.WithError(err).Error("failed to create MicroVM audit log")
		}
	}()
}

func respondJSON(w http.ResponseWriter, status int, data interface{}) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	json.NewEncoder(w).Encode(data)
}

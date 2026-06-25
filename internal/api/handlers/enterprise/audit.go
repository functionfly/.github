package enterprise

import (
	"crypto/hmac"
	"crypto/sha256"
	"encoding/base64"
	"encoding/json"
	"net/http"
	"strconv"
	"time"

	"github.com/functionfly/functionfly/internal/api/middleware"
	"github.com/functionfly/functionfly/internal/apierror"
	"github.com/functionfly/functionfly/internal/config"
	"github.com/functionfly/functionfly/internal/plans"
	"github.com/functionfly/functionfly/internal/storage"
	enterpriseaudit "github.com/functionfly/functionfly/internal/storage/enterprise_audit"
	"github.com/google/uuid"
	"github.com/sirupsen/logrus"
)

const (
	defaultLimit      = 50
	maxLimit          = 500
	defaultExportDays = 30
)

type AuditHandler struct {
	storageRepo *enterpriseaudit.Repository
	exportRepo  *enterpriseaudit.ExportRepository
	repo        storage.Repository
	auditKey    string
}

func NewAuditHandler(repo storage.Repository, auditRepo *enterpriseaudit.Repository, exportRepo *enterpriseaudit.ExportRepository) *AuditHandler {
	return &AuditHandler{
		repo:        repo,
		storageRepo: auditRepo,
		exportRepo:  exportRepo,
	}
}

func (h *AuditHandler) SetAuditSigningKey(key string) {
	h.auditKey = key
}

func (h *AuditHandler) requireEnterprisePlan(w http.ResponseWriter, r *http.Request) bool {
	user := middleware.GetUserFromContext(r)
	if user == nil {
		apierror.WriteError(w, apierror.NewUnauthorized("Unauthorized"))
		return false
	}
	tenant, err := h.repo.GetTenantByID(r.Context(), user.TenantID)
	if err != nil {
		logrus.WithError(err).WithField("tenant_id", user.TenantID).Error("Failed to get tenant for enterprise audit")
		apierror.WriteError(w, apierror.NewInternal("Failed to verify plan"))
		return false
	}
	if tenant == nil || !plans.IsEnterpriseTier(tenant.Plan) {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusForbidden)
		json.NewEncoder(w).Encode(map[string]interface{}{
			"error":   "enterprise_required",
			"message": "Enterprise Audit Logs are available only for Enterprise plan customers.",
		})
		return false
	}
	return true
}

func (h *AuditHandler) HandleListAuditLogs(w http.ResponseWriter, r *http.Request) {
	if !h.requireEnterprisePlan(w, r) {
		return
	}

	user := middleware.GetUserFromContext(r)
	limit := defaultLimit
	if l := r.URL.Query().Get("limit"); l != "" {
		if n, err := strconv.Atoi(l); err == nil && n > 0 {
			limit = n
			if limit > maxLimit {
				limit = maxLimit
			}
		}
	}

	offset := 0
	if o := r.URL.Query().Get("offset"); o != "" {
		if n, err := strconv.Atoi(o); err == nil && n > 0 {
			offset = n
		}
	}

	filters := enterpriseaudit.ListFilters{
		TenantID: user.TenantID,
		Limit:    limit,
		Offset:   offset,
	}

	if sa := r.URL.Query().Get("service_area"); sa != "" {
		saVal := enterpriseaudit.ServiceArea(sa)
		if saVal.Valid() {
			filters.ServiceArea = &saVal
		}
	}

	if action := r.URL.Query().Get("action"); action != "" {
		filters.Action = &action
	}

	if rt := r.URL.Query().Get("resource_type"); rt != "" {
		rtVal := enterpriseaudit.ResourceType(rt)
		if rtVal.Valid() {
			filters.ResourceType = &rtVal
		}
	}

	if resourceID := r.URL.Query().Get("resource_id"); resourceID != "" {
		if rid, err := uuid.Parse(resourceID); err == nil {
			filters.ResourceID = &rid
		}
	}

	if actorType := r.URL.Query().Get("actor_type"); actorType != "" {
		atVal := enterpriseaudit.ActorType(actorType)
		if atVal.Valid() {
			filters.ActorType = &atVal
		}
	}

	if actorID := r.URL.Query().Get("actor_id"); actorID != "" {
		filters.ActorID = &actorID
	}

	if success := r.URL.Query().Get("success"); success != "" {
		s := success == "true"
		filters.Success = &s
	}

	if startTime := r.URL.Query().Get("start_time"); startTime != "" {
		if t, err := time.Parse(time.RFC3339, startTime); err == nil {
			filters.StartTime = &t
		}
	}

	if endTime := r.URL.Query().Get("end_time"); endTime != "" {
		if t, err := time.Parse(time.RFC3339, endTime); err == nil {
			filters.EndTime = &t
		}
	}

	if search := r.URL.Query().Get("search"); search != "" {
		filters.Search = &search
	}

	logs, total, err := h.storageRepo.List(r.Context(), filters)
	if err != nil {
		logrus.WithError(err).Error("Failed to list enterprise audit logs")
		apierror.WriteError(w, apierror.NewInternal("Failed to list audit logs"))
		return
	}

	responseLogs := make([]map[string]interface{}, 0, len(logs))
	for _, log := range logs {
		responseLogs = append(responseLogs, map[string]interface{}{
			"id":            log.ID.String(),
			"tenant_id":     log.TenantID.String(),
			"service_area":  string(log.ServiceArea),
			"action":        log.Action,
			"resource_type": string(log.ResourceType),
			"resource_id":   log.ResourceID,
			"actor_type":    string(log.ActorType),
			"actor_id":      log.ActorID,
			"actor_name":    log.ActorName,
			"request_id":    log.RequestID,
			"ip_address":    log.IPAddress,
			"user_agent":    log.UserAgent,
			"metadata":      log.GetMetadata(),
			"success":       log.Success,
			"error_message": log.ErrorMessage,
			"created_at":    log.CreatedAt.Format(time.RFC3339),
		})
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]interface{}{
		"logs":   responseLogs,
		"total":  total,
		"limit":  limit,
		"offset": offset,
	})
}

func (h *AuditHandler) HandleGetAuditLog(w http.ResponseWriter, r *http.Request) {
	if !h.requireEnterprisePlan(w, r) {
		return
	}

	idStr := r.URL.Query().Get("id")
	if idStr == "" {
		apierror.WriteError(w, apierror.NewBadRequest("id is required"))
		return
	}

	id, err := uuid.Parse(idStr)
	if err != nil {
		apierror.WriteError(w, apierror.NewBadRequest("invalid id format"))
		return
	}

	log, err := h.storageRepo.GetByID(r.Context(), id)
	if err != nil {
		logrus.WithError(err).Error("Failed to get enterprise audit log")
		apierror.WriteError(w, apierror.NewInternal("Failed to get audit log"))
		return
	}

	if log == nil {
		apierror.WriteError(w, apierror.NewNotFound("audit log not found"))
		return
	}

	user := middleware.GetUserFromContext(r)
	if log.TenantID != user.TenantID {
		apierror.WriteError(w, apierror.NewNotFound("audit log not found"))
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]interface{}{
		"log": map[string]interface{}{
			"id":            log.ID.String(),
			"tenant_id":     log.TenantID.String(),
			"service_area":  string(log.ServiceArea),
			"action":        log.Action,
			"resource_type": string(log.ResourceType),
			"resource_id":   log.ResourceID,
			"actor_type":    string(log.ActorType),
			"actor_id":      log.ActorID,
			"actor_name":    log.ActorName,
			"request_id":    log.RequestID,
			"ip_address":    log.IPAddress,
			"user_agent":    log.UserAgent,
			"metadata":      log.GetMetadata(),
			"success":       log.Success,
			"error_message": log.ErrorMessage,
			"created_at":    log.CreatedAt.Format(time.RFC3339),
		},
	})
}

func (h *AuditHandler) HandleGetFilters(w http.ResponseWriter, r *http.Request) {
	if !h.requireEnterprisePlan(w, r) {
		return
	}

	user := middleware.GetUserFromContext(r)

	actions, err := h.storageRepo.GetActions(r.Context(), user.TenantID)
	if err != nil {
		logrus.WithError(err).Error("Failed to get audit actions")
	}

	serviceAreas, err := h.storageRepo.GetServiceAreas(r.Context(), user.TenantID)
	if err != nil {
		logrus.WithError(err).Error("Failed to get audit service areas")
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]interface{}{
		"service_areas": serviceAreas,
		"actions":       actions,
	})
}

func (h *AuditHandler) HandleExportAudit(w http.ResponseWriter, r *http.Request) {
	if !h.requireEnterprisePlan(w, r) {
		return
	}

	user := middleware.GetUserFromContext(r)

	from := time.Now().AddDate(0, 0, -defaultExportDays)
	if f := r.URL.Query().Get("from"); f != "" {
		if t, err := time.Parse(time.RFC3339, f); err == nil {
			from = t
		}
	}

	to := time.Now()
	if t := r.URL.Query().Get("to"); t != "" {
		if deadline, err := time.Parse(time.RFC3339, t); err == nil {
			to = deadline
		}
	}

	format := enterpriseaudit.ExportFormat(r.URL.Query().Get("format"))
	if format != enterpriseaudit.ExportFormatCSV && format != enterpriseaudit.ExportFormatCEF {
		format = enterpriseaudit.ExportFormatJSON
	}

	query := enterpriseaudit.ExportQuery{
		TenantID: user.TenantID,
		Format:  format,
		From:    from,
		To:      to,
	}

	if sa := r.URL.Query().Get("service_area"); sa != "" {
		saVal := enterpriseaudit.ServiceArea(sa)
		if saVal.Valid() {
			query.ServiceArea = &saVal
		}
	}

	if action := r.URL.Query().Get("action"); action != "" {
		query.Action = &action
	}

	result, err := h.exportRepo.Export(r.Context(), query)
	if err != nil {
		logrus.WithError(err).Error("Failed to export enterprise audit logs")
		apierror.WriteError(w, apierror.NewInternal("Failed to export audit logs"))
		return
	}

	signingKey := h.auditKey
	if signingKey == "" {
		if config.IsProduction() {
			logrus.Error("ENTERPRISE_AUDIT_SIGNING_KEY not configured in production - audit export rejected")
			apierror.WriteError(w, apierror.NewInternal("Audit signing key not configured - set ENTERPRISE_AUDIT_SIGNING_KEY environment variable"))
			return
		}
		signingKey = "dev-fallback-audit-signing-key"
	}
	hmacKey := auditSigningKey(signingKey, user.TenantID)
	hmacHash := hmac.New(sha256.New, hmacKey)
	hmacHash.Write(result.Body)
	signature := base64.StdEncoding.EncodeToString(hmacHash.Sum(nil))

	filename := "enterprise-audit-" + user.TenantID.String() + "." + string(format)

	w.Header().Set("Content-Type", contentTypeForExport(string(format)))
	w.Header().Set("Content-Disposition", `attachment; filename="`+filename+`"`)
	w.Header().Set("X-Audit-Signature", signature)
	w.Header().Set("X-Audit-Generated-At", result.Generated.Format(time.RFC3339Nano))
	w.Header().Set("X-Audit-Row-Count", strconv.Itoa(result.RowCount))
	w.Header().Set("Cache-Control", "no-store")
	w.WriteHeader(http.StatusOK)
	_, _ = w.Write(result.Body)
}

func auditSigningKey(key string, tenantID uuid.UUID) []byte {
	h := hmac.New(sha256.New, []byte(key))
	h.Write(tenantID[:])
	return h.Sum(nil)
}

func contentTypeForExport(format string) string {
	switch format {
	case "csv":
		return "text/csv; charset=utf-8"
	case "cef":
		return "text/plain; charset=utf-8"
	default:
		return "application/json"
	}
}

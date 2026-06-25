package enterpriseaudit

import (
	"context"
	"encoding/json"
	"fmt"
	"time"

	"github.com/google/uuid"
	"gorm.io/gorm"
)

type Repository struct {
	db *gorm.DB
}

func NewRepository(db *gorm.DB) *Repository {
	return &Repository{db: db}
}

func (r *Repository) Create(ctx context.Context, log *AuditLog) error {
	if log.ID == uuid.Nil {
		log.ID = uuid.New()
	}
	if log.CreatedAt.IsZero() {
		log.CreatedAt = time.Now()
	}
	return r.db.WithContext(ctx).Create(log).Error
}

func (r *Repository) CreateBatch(ctx context.Context, logs []*AuditLog) error {
	now := time.Now()
	for _, log := range logs {
		if log.ID == uuid.Nil {
			log.ID = uuid.New()
		}
		if log.CreatedAt.IsZero() {
			log.CreatedAt = now
		}
	}
	return r.db.WithContext(ctx).CreateInBatches(logs, 100).Error
}

func (r *Repository) GetByID(ctx context.Context, id uuid.UUID) (*AuditLog, error) {
	var log AuditLog
	err := r.db.WithContext(ctx).Where("id = ?", id).First(&log).Error
	if err != nil {
		return nil, err
	}
	return &log, nil
}

func (r *Repository) List(ctx context.Context, filters ListFilters) ([]*AuditLog, int64, error) {
	query := r.db.WithContext(ctx).Model(&AuditLog{}).Where("tenant_id = ?", filters.TenantID)

	if filters.ServiceArea != nil {
		query = query.Where("service_area = ?", *filters.ServiceArea)
	}
	if filters.Action != nil {
		query = query.Where("action = ?", *filters.Action)
	}
	if filters.ResourceType != nil {
		query = query.Where("resource_type = ?", *filters.ResourceType)
	}
	if filters.ResourceID != nil {
		query = query.Where("resource_id = ?", *filters.ResourceID)
	}
	if filters.ActorType != nil {
		query = query.Where("actor_type = ?", *filters.ActorType)
	}
	if filters.ActorID != nil {
		query = query.Where("actor_id = ?", *filters.ActorID)
	}
	if filters.Success != nil {
		query = query.Where("success = ?", *filters.Success)
	}
	if filters.StartTime != nil {
		query = query.Where("created_at >= ?", *filters.StartTime)
	}
	if filters.EndTime != nil {
		query = query.Where("created_at <= ?", *filters.EndTime)
	}
	if filters.Search != nil && *filters.Search != "" {
		searchTerm := "%" + *filters.Search + "%"
		query = query.Where("actor_id ILIKE ? OR actor_name ILIKE ? OR action ILIKE ?", searchTerm, searchTerm, searchTerm)
	}

	var total int64
	if err := query.Count(&total).Error; err != nil {
		return nil, 0, err
	}

	if filters.Limit <= 0 {
		filters.Limit = 50
	}
	if filters.Limit > 1000 {
		filters.Limit = 1000
	}

	query = query.Order("created_at DESC").Limit(filters.Limit)
	if filters.Offset > 0 {
		query = query.Offset(filters.Offset)
	}

	var logs []*AuditLog
	if err := query.Find(&logs).Error; err != nil {
		return nil, 0, err
	}

	return logs, total, nil
}

func (r *Repository) Count(ctx context.Context, tenantID uuid.UUID) (int64, error) {
	var count int64
	err := r.db.WithContext(ctx).Model(&AuditLog{}).Where("tenant_id = ?", tenantID).Count(&count).Error
	return count, err
}

func (r *Repository) GetByResource(ctx context.Context, tenantID, resourceID uuid.UUID, limit, offset int) ([]*AuditLog, int64, error) {
	filters := ListFilters{
		TenantID:   tenantID,
		ResourceID: &resourceID,
		Limit:      limit,
		Offset:     offset,
	}
	return r.List(ctx, filters)
}

func (r *Repository) GetByActor(ctx context.Context, tenantID uuid.UUID, actorID string, limit, offset int) ([]*AuditLog, int64, error) {
	filters := ListFilters{
		TenantID: tenantID,
		ActorID:  &actorID,
		Limit:    limit,
		Offset:   offset,
	}
	return r.List(ctx, filters)
}

func (r *Repository) GetByServiceArea(ctx context.Context, tenantID uuid.UUID, serviceArea ServiceArea, limit, offset int) ([]*AuditLog, int64, error) {
	filters := ListFilters{
		TenantID:    tenantID,
		ServiceArea: &serviceArea,
		Limit:       limit,
		Offset:      offset,
	}
	return r.List(ctx, filters)
}

func (r *Repository) GetActions(ctx context.Context, tenantID uuid.UUID) ([]string, error) {
	var actions []string
	err := r.db.WithContext(ctx).
		Model(&AuditLog{}).
		Where("tenant_id = ?", tenantID).
		Distinct("action").
		Pluck("action", &actions).Error
	return actions, err
}

func (r *Repository) GetServiceAreas(ctx context.Context, tenantID uuid.UUID) ([]ServiceArea, error) {
	var areas []ServiceArea
	err := r.db.WithContext(ctx).
		Model(&AuditLog{}).
		Where("tenant_id = ?", tenantID).
		Distinct("service_area").
		Pluck("service_area", &areas).Error
	return areas, err
}

func (r *Repository) DeleteOld(ctx context.Context, before time.Time) (int64, error) {
	result := r.db.WithContext(ctx).Where("created_at < ?", before).Delete(&AuditLog{})
	return result.RowsAffected, result.Error
}

func (r *Repository) DB() *gorm.DB {
	return r.db
}

type ExportRepository struct {
	*Repository
}

func NewExportRepository(db *gorm.DB) *ExportRepository {
	return &ExportRepository{NewRepository(db)}
}

func (r *ExportRepository) Export(ctx context.Context, query ExportQuery) (*ExportResult, error) {
	dbQuery := r.db.WithContext(ctx).Model(&AuditLog{}).
		Where("tenant_id = ?", query.TenantID).
		Where("created_at >= ?", query.From).
		Where("created_at <= ?", query.To)

	if query.ServiceArea != nil {
		dbQuery = dbQuery.Where("service_area = ?", *query.ServiceArea)
	}
	if query.Action != nil {
		dbQuery = dbQuery.Where("action = ?", *query.Action)
	}

	var logs []*AuditLog
	if err := dbQuery.Order("created_at ASC").Find(&logs).Error; err != nil {
		return nil, err
	}

	var body []byte
	var err error
	switch query.Format {
	case ExportFormatCSV:
		body, err = exportToCSV(logs)
	case ExportFormatCEF:
		body, err = exportToCEF(logs)
	default:
		body, err = exportToJSON(logs)
	}
	if err != nil {
		return nil, err
	}

	return &ExportResult{
		Format:    query.Format,
		Body:      body,
		Generated: time.Now(),
		RowCount:  len(logs),
	}, nil
}

func exportToJSON(logs []*AuditLog) ([]byte, error) {
	type exportLog struct {
		ID           string                 `json:"id"`
		TenantID     string                 `json:"tenant_id"`
		ServiceArea  string                 `json:"service_area"`
		Action       string                 `json:"action"`
		ResourceType string                 `json:"resource_type"`
		ResourceID   string                 `json:"resource_id,omitempty"`
		ActorType    string                 `json:"actor_type"`
		ActorID      string                 `json:"actor_id"`
		ActorName    string                 `json:"actor_name,omitempty"`
		RequestID    string                 `json:"request_id,omitempty"`
		IPAddress    string                 `json:"ip_address,omitempty"`
		UserAgent    string                 `json:"user_agent,omitempty"`
		Metadata     map[string]interface{} `json:"metadata,omitempty"`
		Success      bool                   `json:"success"`
		ErrorMessage string                 `json:"error_message,omitempty"`
		CreatedAt    string                 `json:"created_at"`
	}

	exportLogs := make([]exportLog, len(logs))
	for i, l := range logs {
		exportLogs[i] = exportLog{
			ID:           l.ID.String(),
			TenantID:     l.TenantID.String(),
			ServiceArea:  string(l.ServiceArea),
			Action:       l.Action,
			ResourceType: string(l.ResourceType),
			ActorType:    string(l.ActorType),
			ActorID:      l.ActorID,
			ActorName:    l.ActorName,
			RequestID:    l.RequestID,
			IPAddress:    l.IPAddress,
			UserAgent:    l.UserAgent,
			Metadata:     l.GetMetadata(),
			Success:      l.Success,
			ErrorMessage: l.ErrorMessage,
			CreatedAt:    l.CreatedAt.Format(time.RFC3339),
		}
		if l.ResourceID != nil {
			exportLogs[i].ResourceID = l.ResourceID.String()
		}
	}

	return json.Marshal(exportLogs)
}

func exportToCSV(logs []*AuditLog) ([]byte, error) {
	csv := "id,tenant_id,service_area,action,resource_type,resource_id,actor_type,actor_id,actor_name,ip_address,user_agent,success,error_message,created_at\n"
	for _, l := range logs {
		resourceID := ""
		if l.ResourceID != nil {
			resourceID = l.ResourceID.String()
		}
		csv += fmt.Sprintf("%s,%s,%s,%s,%s,%s,%s,%s,%s,%s,%s,%t,%s,%s\n",
			l.ID.String(),
			l.TenantID.String(),
			l.ServiceArea,
			l.Action,
			l.ResourceType,
			resourceID,
			l.ActorType,
			l.ActorID,
			l.ActorName,
			l.IPAddress,
			l.UserAgent,
			l.Success,
			l.ErrorMessage,
			l.CreatedAt.Format(time.RFC3339),
		)
	}
	return []byte(csv), nil
}

func exportToCEF(logs []*AuditLog) ([]byte, error) {
	var cef string
	for _, l := range logs {
		resourceID := ""
		if l.ResourceID != nil {
			resourceID = l.ResourceID.String()
		}
		cef += fmt.Sprintf("CEF:0|FunctionFly|EnterpriseAudit|1.0|%s|%s|%s|",
			l.Action,
			l.Action,
			fmt.Sprintf("src=%s dst=%s suser=%s fname=%s",
				l.IPAddress,
				resourceID,
				l.ActorID,
				l.ActorName,
			),
		)
		cef += fmt.Sprintf(" cn1=%d\n", 0)
		_ = cef
	}
	return []byte(cef), nil
}

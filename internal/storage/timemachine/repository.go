package timemachine

import (
	"encoding/json"
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

// --- Replay CRUD ---

func (r *Repository) CreateReplay(replay *Replay) error {
	if replay.ID == uuid.Nil {
		replay.ID = uuid.New()
	}
	return r.db.Create(replay).Error
}

func (r *Repository) GetReplay(id uuid.UUID) (*Replay, error) {
	var replay Replay
	err := r.db.Where("id = ?", id).First(&replay).Error
	if err != nil {
		if err == gorm.ErrRecordNotFound {
			return nil, nil
		}
		return nil, err
	}
	return &replay, nil
}

func (r *Repository) GetReplayWithItems(id uuid.UUID) (*Replay, []ReplayItem, error) {
	var replay Replay
	err := r.db.Where("id = ?", id).First(&replay).Error
	if err != nil {
		if err == gorm.ErrRecordNotFound {
			return nil, nil, nil
		}
		return nil, nil, err
	}

	var items []ReplayItem
	err = r.db.Where("replay_id = ?", id).Order("created_at ASC").Find(&items).Error
	if err != nil {
		return nil, nil, err
	}

	return &replay, items, nil
}

func (r *Repository) ListReplaysByTenant(tenantID uuid.UUID, limit, offset int) ([]Replay, int64, error) {
	var replays []Replay
	var total int64

	query := r.db.Model(&Replay{}).Where("tenant_id = ?", tenantID)
	if err := query.Count(&total).Error; err != nil {
		return nil, 0, err
	}

	err := query.Order("created_at DESC").Limit(limit).Offset(offset).Find(&replays).Error
	if err != nil {
		return nil, 0, err
	}

	return replays, total, nil
}

func (r *Repository) ListReplaysByFunction(functionID uuid.UUID, limit, offset int) ([]Replay, int64, error) {
	var replays []Replay
	var total int64

	query := r.db.Model(&Replay{}).Where("function_id = ?", functionID)
	if err := query.Count(&total).Error; err != nil {
		return nil, 0, err
	}

	err := query.Order("created_at DESC").Limit(limit).Offset(offset).Find(&replays).Error
	if err != nil {
		return nil, 0, err
	}

	return replays, total, nil
}

func (r *Repository) UpdateReplay(replay *Replay) error {
	return r.db.Save(replay).Error
}

func (r *Repository) UpdateReplayStatus(id uuid.UUID, status string, progress float64, phase string) error {
	updates := map[string]interface{}{
		"status":           status,
		"progress_percent": progress,
		"current_phase":    nilIfEmpty(phase),
	}
	if status == "running" {
		now := time.Now()
		updates["started_at"] = &now
	}
	if status == "completed" || status == "failed" {
		now := time.Now()
		updates["completed_at"] = &now
	}
	return r.db.Model(&Replay{}).Where("id = ?", id).Updates(updates).Error
}

func (r *Repository) UpdateReplayProgress(id uuid.UUID, status string, progress float64, phase string, found, replayed, changed, failed int) error {
	updates := map[string]interface{}{
		"status":                     status,
		"progress_percent":           progress,
		"current_phase":              nilIfEmpty(phase),
		"total_executions_found":     found,
		"total_executions_replayed":  replayed,
		"total_executions_changed":   changed,
		"total_executions_failed":    failed,
	}
	if status == "completed" || status == "failed" {
		now := time.Now()
		updates["completed_at"] = &now
	}
	return r.db.Model(&Replay{}).Where("id = ?", id).Updates(updates).Error
}

func (r *Repository) DeleteReplay(id uuid.UUID) error {
	return r.db.Transaction(func(tx *gorm.DB) error {
		if err := tx.Where("replay_id = ?", id).Delete(&Reconciliation{}).Error; err != nil {
			return err
		}
		if err := tx.Where("replay_id = ?", id).Delete(&ReplayItem{}).Error; err != nil {
			return err
		}
		if err := tx.Where("replay_id = ?", id).Delete(&AuditCertificate{}).Error; err != nil {
			return err
		}
		return tx.Where("id = ?", id).Delete(&Replay{}).Error
	})
}

func (r *Repository) CountActiveReplaysByTenant(tenantID uuid.UUID) (int64, error) {
	var count int64
	err := r.db.Model(&Replay{}).
		Where("tenant_id = ? AND status IN ?", tenantID, []string{"pending", "running"}).
		Count(&count).Error
	return count, err
}

// --- ReplayItem CRUD ---

func (r *Repository) CreateReplayItems(items []ReplayItem) error {
	if len(items) == 0 {
		return nil
	}
	for i := range items {
		if items[i].ID == uuid.Nil {
			items[i].ID = uuid.New()
		}
	}
	return r.db.Create(&items).Error
}

func (r *Repository) GetReplayItem(id uuid.UUID) (*ReplayItem, error) {
	var item ReplayItem
	err := r.db.Where("id = ?", id).First(&item).Error
	if err != nil {
		if err == gorm.ErrRecordNotFound {
			return nil, nil
		}
		return nil, err
	}
	return &item, nil
}

func (r *Repository) ListReplayItems(replayID uuid.UUID, limit, offset int, filterDiffType string) ([]ReplayItem, int64, error) {
	var items []ReplayItem
	var total int64

	query := r.db.Model(&ReplayItem{}).Where("replay_id = ?", replayID)
	if filterDiffType != "" {
		query = query.Where("diff_type = ?", filterDiffType)
	}

	if err := query.Count(&total).Error; err != nil {
		return nil, 0, err
	}

	err := query.Order("created_at ASC").Limit(limit).Offset(offset).Find(&items).Error
	if err != nil {
		return nil, 0, err
	}

	return items, total, nil
}

func (r *Repository) ListChangedItems(replayID uuid.UUID, limit, offset int) ([]ReplayItem, int64, error) {
	var items []ReplayItem
	var total int64

	query := r.db.Model(&ReplayItem{}).Where("replay_id = ? AND output_changed = ?", replayID, true)

	if err := query.Count(&total).Error; err != nil {
		return nil, 0, err
	}

	err := query.Order("created_at ASC").Limit(limit).Offset(offset).Find(&items).Error
	if err != nil {
		return nil, 0, err
	}

	return items, total, nil
}

func (r *Repository) UpdateReplayItem(item *ReplayItem) error {
	return r.db.Save(item).Error
}

func (r *Repository) UpdateReplayItemResult(id uuid.UUID, newOutput json.RawMessage, newDuration int, newStatusCode int, status string) error {
	updates := map[string]interface{}{
		"new_output":      newOutput,
		"new_duration_ms": newDuration,
		"new_status_code": newStatusCode,
		"status":          status,
	}
	return r.db.Model(&ReplayItem{}).Where("id = ?", id).Updates(updates).Error
}

func (r *Repository) UpdateReplayItemDiff(id uuid.UUID, changed bool, diffType, diffSummary string, diffDetail json.RawMessage) error {
	updates := map[string]interface{}{
		"output_changed": changed,
		"diff_type":      nilIfEmpty(diffType),
		"diff_summary":   nilIfEmpty(diffSummary),
		"diff_detail":    diffDetail,
	}
	return r.db.Model(&ReplayItem{}).Where("id = ?", id).Updates(updates).Error
}

func (r *Repository) UpdateReplayItemStatus(id uuid.UUID, status string) error {
	return r.db.Model(&ReplayItem{}).Where("id = ?", id).Update("status", status).Error
}

func (r *Repository) UpdateReplayItemError(id uuid.UUID, errMsg, errCode string) error {
	updates := map[string]interface{}{
		"replay_error":       errMsg,
		"replay_error_code":  errCode,
		"status":             "failed",
	}
	return r.db.Model(&ReplayItem{}).Where("id = ?", id).Updates(updates).Error
}

func (r *Repository) ListReplayItemsByStatus(replayID uuid.UUID, status string) ([]ReplayItem, error) {
	var items []ReplayItem
	err := r.db.Where("replay_id = ? AND status = ?", replayID, status).Order("created_at ASC").Find(&items).Error
	return items, err
}

// --- Reconciliation CRUD ---

func (r *Repository) CreateReconciliations(reconciliations []Reconciliation) error {
	if len(reconciliations) == 0 {
		return nil
	}
	for i := range reconciliations {
		if reconciliations[i].ID == uuid.Nil {
			reconciliations[i].ID = uuid.New()
		}
	}
	return r.db.Create(&reconciliations).Error
}

func (r *Repository) ListReconciliations(replayID uuid.UUID, limit, offset int) ([]Reconciliation, int64, error) {
	var recs []Reconciliation
	var total int64

	query := r.db.Model(&Reconciliation{}).Where("replay_id = ?", replayID)
	if err := query.Count(&total).Error; err != nil {
		return nil, 0, err
	}

	err := query.Order("created_at ASC").Limit(limit).Offset(offset).Find(&recs).Error
	if err != nil {
		return nil, 0, err
	}

	return recs, total, nil
}

func (r *Repository) UpdateReconciliationStatus(id uuid.UUID, status string, errMsg string) error {
	updates := map[string]interface{}{
		"status": status,
	}
	if errMsg != "" {
		updates["error_message"] = errMsg
	}
	if status == "applied" {
		now := time.Now()
		updates["applied_at"] = &now
	}
	return r.db.Model(&Reconciliation{}).Where("id = ?", id).Updates(updates).Error
}

func (r *Repository) GetReconciliation(id uuid.UUID) (*Reconciliation, error) {
	var rec Reconciliation
	err := r.db.Where("id = ?", id).First(&rec).Error
	if err != nil {
		if err == gorm.ErrRecordNotFound {
			return nil, nil
		}
		return nil, err
	}
	return &rec, nil
}

// --- AuditCertificate CRUD ---

func (r *Repository) CreateAuditCertificate(cert *AuditCertificate) error {
	if cert.ID == uuid.Nil {
		cert.ID = uuid.New()
	}
	return r.db.Create(cert).Error
}

func (r *Repository) GetAuditCertificateByReplayID(replayID uuid.UUID) (*AuditCertificate, error) {
	var cert AuditCertificate
	err := r.db.Where("replay_id = ?", replayID).First(&cert).Error
	if err != nil {
		if err == gorm.ErrRecordNotFound {
			return nil, nil
		}
		return nil, err
	}
	return &cert, nil
}

func (r *Repository) GetAuditCertificateByID(certID string) (*AuditCertificate, error) {
	var cert AuditCertificate
	err := r.db.Where("certificate_id = ?", certID).First(&cert).Error
	if err != nil {
		if err == gorm.ErrRecordNotFound {
			return nil, nil
		}
		return nil, err
	}
	return &cert, nil
}

func (r *Repository) GetLatestAuditCertificateForFunction(functionID uuid.UUID) (*AuditCertificate, error) {
	var cert AuditCertificate
	err := r.db.
		Joins("JOIN time_machine_replays ON time_machine_replays.id = time_machine_audit_certificates.replay_id").
		Where("time_machine_replays.function_id = ?", functionID).
		Order("time_machine_audit_certificates.created_at DESC").
		First(&cert).Error
	if err != nil {
		if err == gorm.ErrRecordNotFound {
			return nil, nil
		}
		return nil, err
	}
	return &cert, nil
}

func (r *Repository) ListAuditCertificates(tenantID uuid.UUID, limit, offset int) ([]AuditCertificate, int64, error) {
	var certs []AuditCertificate
	var total int64

	query := r.db.Model(&AuditCertificate{}).
		Joins("JOIN time_machine_replays ON time_machine_replays.id = time_machine_audit_certificates.replay_id").
		Where("time_machine_replays.tenant_id = ?", tenantID)

	if err := query.Count(&total).Error; err != nil {
		return nil, 0, err
	}

	err := query.Order("time_machine_audit_certificates.created_at DESC").Limit(limit).Offset(offset).Find(&certs).Error
	if err != nil {
		return nil, 0, err
	}

	return certs, total, nil
}

// --- Schedule CRUD ---

func (r *Repository) CreateSchedule(schedule *Schedule) error {
	if schedule.ID == uuid.Nil {
		schedule.ID = uuid.New()
	}
	return r.db.Create(schedule).Error
}

func (r *Repository) GetSchedule(id uuid.UUID) (*Schedule, error) {
	var schedule Schedule
	err := r.db.Where("id = ?", id).First(&schedule).Error
	if err != nil {
		if err == gorm.ErrRecordNotFound {
			return nil, nil
		}
		return nil, err
	}
	return &schedule, nil
}

func (r *Repository) ListSchedulesByTenant(tenantID uuid.UUID) ([]Schedule, error) {
	var schedules []Schedule
	err := r.db.Where("tenant_id = ?", tenantID).Order("created_at DESC").Find(&schedules).Error
	return schedules, err
}

func (r *Repository) UpdateSchedule(schedule *Schedule) error {
	return r.db.Save(schedule).Error
}

func (r *Repository) DeleteSchedule(id uuid.UUID) error {
	return r.db.Where("id = ?", id).Delete(&Schedule{}).Error
}

func (r *Repository) GetDueSchedules() ([]Schedule, error) {
	var schedules []Schedule
	err := r.db.Where("enabled = ? AND next_run_at <= ?", true, time.Now()).Find(&schedules).Error
	return schedules, err
}

// --- helpers ---

func nilIfEmpty(s string) interface{} {
	if s == "" {
		return nil
	}
	return s
}

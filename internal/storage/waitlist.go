package storage

import (
	"context"
	"errors"
	"time"

	"github.com/google/uuid"
	"gorm.io/gorm"
)

var (
	// ErrWaitlistEntryExists is returned when the email is already on the waitlist
	ErrWaitlistEntryExists = errors.New("email is already on the waitlist")
	// ErrWaitlistEntryNotFound is returned when the entry is not found
	ErrWaitlistEntryNotFound = errors.New("waitlist entry not found")
)

// CreateWaitlistEntry adds a new entry to the waitlist
func (db *PostgresDB) CreateWaitlistEntry(ctx context.Context, email, name, company, useCase, source, ip, userAgent string) (*WaitlistEntry, error) {
	// Normalize email
	email = strings.ToLower(strings.TrimSpace(email))

	entry := &WaitlistEntry{
		ID:        uuid.New(),
		Email:     email,
		Name:      strings.TrimSpace(name),
		Company:   strings.TrimSpace(company),
		UseCase:   strings.TrimSpace(useCase),
		Source:    source,
		Status:    "pending",
		IPAddress: ip,
		UserAgent: userAgent,
	}

	if err := db.GORM.WithContext(ctx).Create(entry).Error; err != nil {
		if IsDuplicateKeyError(err) {
			return nil, ErrWaitlistEntryExists
		}
		return nil, err
	}

	return entry, nil
}

// GetWaitlistEntryByEmail retrieves a waitlist entry by email
func (db *PostgresDB) GetWaitlistEntryByEmail(ctx context.Context, email string) (*WaitlistEntry, error) {
	email = strings.ToLower(strings.TrimSpace(email))

	var entry WaitlistEntry
	if err := db.GORM.WithContext(ctx).Where("email = ?", email).First(&entry).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, ErrWaitlistEntryNotFound
		}
		return nil, err
	}
	return &entry, nil
}

// ListWaitlistEntries returns a paginated list of waitlist entries
func (db *PostgresDB) ListWaitlistEntries(ctx context.Context, status string, limit, offset int) ([]WaitlistEntryAdminList, int64, error) {
	var entries []WaitlistEntry
	var total int64

	query := db.GORM.WithContext(ctx).Model(&WaitlistEntry{})

	// Filter by status if provided
	if status != "" && status != "all" {
		query = query.Where("status = ?", status)
	}

	// Get total count
	if err := query.Count(&total).Error; err != nil {
		return nil, 0, err
	}

	// Get entries with pagination
	if err := query.Order("created_at DESC").Limit(limit).Offset(offset).Find(&entries).Error; err != nil {
		return nil, 0, err
	}

	// Convert to admin list format
	out := make([]WaitlistEntryAdminList, 0, len(entries))
	for _, e := range entries {
		out = append(out, WaitlistEntryAdminList{
			ID:           e.ID,
			Email:        e.Email,
			Name:         e.Name,
			Company:      e.Company,
			UseCase:      e.UseCase,
			Source:       e.Source,
			Status:       e.Status,
			InviteCodeID: e.InviteCodeID,
			InvitedAt:    e.InvitedAt,
			Notes:        e.Notes,
			IPAddress:    e.IPAddress,
			CreatedAt:    e.CreatedAt,
			UpdatedAt:    e.UpdatedAt,
		})
	}

	return out, total, nil
}

// GetWaitlistStats returns aggregate statistics for the waitlist
func (db *PostgresDB) GetWaitlistStats(ctx context.Context) (*WaitlistStats, error) {
	var stats WaitlistStats

	if err := db.GORM.WithContext(ctx).Model(&WaitlistEntry{}).Count(&stats.Total).Error; err != nil {
		return nil, err
	}

	if err := db.GORM.WithContext(ctx).Model(&WaitlistEntry{}).Where("status = ?", "pending").Count(&stats.Pending).Error; err != nil {
		return nil, err
	}

	if err := db.GORM.WithContext(ctx).Model(&WaitlistEntry{}).Where("status = ?", "approved").Count(&stats.Approved).Error; err != nil {
		return nil, err
	}

	if err := db.GORM.WithContext(ctx).Model(&WaitlistEntry{}).Where("status = ?", "invited").Count(&stats.Invited).Error; err != nil {
		return nil, err
	}

	if err := db.GORM.WithContext(ctx).Model(&WaitlistEntry{}).Where("status = ?", "rejected").Count(&stats.Rejected).Error; err != nil {
		return nil, err
	}

	return &stats, nil
}

// UpdateWaitlistEntryStatus updates the status of a waitlist entry
func (db *PostgresDB) UpdateWaitlistEntryStatus(ctx context.Context, id uuid.UUID, status, notes string) error {
	updates := map[string]interface{}{
		"status":     status,
		"updated_at": time.Now().UTC(),
	}

	if notes != "" {
		updates["notes"] = gorm.Expr("COALESCE(notes, '') || ?", "\n"+notes)
	}

	result := db.GORM.WithContext(ctx).Model(&WaitlistEntry{}).Where("id = ?", id).Updates(updates)
	if result.Error != nil {
		return result.Error
	}

	if result.RowsAffected == 0 {
		return ErrWaitlistEntryNotFound
	}

	return nil
}

// IssueInviteToWaitlistEntry issues an invite code to a waitlist entry
func (db *PostgresDB) IssueInviteToWaitlistEntry(ctx context.Context, entryID, inviteCodeID uuid.UUID) error {
	now := time.Now().UTC()

	result := db.GORM.WithContext(ctx).Model(&WaitlistEntry{}).Where("id = ?", entryID).Updates(map[string]interface{}{
		"status":         "invited",
		"invite_code_id": inviteCodeID,
		"invited_at":     now,
		"updated_at":     now,
	})

	if result.Error != nil {
		return result.Error
	}

	if result.RowsAffected == 0 {
		return ErrWaitlistEntryNotFound
	}

	return nil
}

// GetPendingWaitlistCount returns the number of pending waitlist entries
func (db *PostgresDB) GetPendingWaitlistCount(ctx context.Context) (int64, error) {
	var count int64
	if err := db.GORM.WithContext(ctx).Model(&WaitlistEntry{}).Where("status = ?", "pending").Count(&count).Error; err != nil {
		return 0, err
	}
	return count, nil
}

// DeleteWaitlistEntry removes a waitlist entry
func (db *PostgresDB) DeleteWaitlistEntry(ctx context.Context, id uuid.UUID) error {
	result := db.GORM.WithContext(ctx).Delete(&WaitlistEntry{}, "id = ?", id)
	if result.Error != nil {
		return result.Error
	}

	if result.RowsAffected == 0 {
		return ErrWaitlistEntryNotFound
	}

	return nil
}

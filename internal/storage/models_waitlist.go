package storage

import (
	"time"

	"github.com/google/uuid"
)

// WaitlistEntry represents a user who has joined the waitlist
type WaitlistEntry struct {
	ID           uuid.UUID  `json:"id" gorm:"type:uuid;primaryKey;default:gen_random_uuid()"`
	Email        string     `json:"email" gorm:"uniqueIndex;not null"`
	Name         string     `json:"name,omitempty" gorm:"size:255"`
	Company      string     `json:"company,omitempty" gorm:"size:255"`
	UseCase      string     `json:"use_case,omitempty" gorm:"column:use_case;type:text"`
	Source       string     `json:"source,omitempty" gorm:"size:100"`        // Where they came from (e.g., "landing_page", "referral")
	Status       string     `json:"status" gorm:"size:50;default:'pending'"` // pending, approved, invited, rejected
	InviteCodeID *uuid.UUID `json:"invite_code_id,omitempty" gorm:"column:invite_code_id;type:uuid"`
	InvitedAt    *time.Time `json:"invited_at,omitempty" gorm:"column:invited_at"`
	Notes        string     `json:"notes,omitempty" gorm:"type:text"`
	IPAddress    string     `json:"ip_address,omitempty" gorm:"column:ip;size:45"`
	UserAgent    string     `json:"user_agent,omitempty" gorm:"column:user_agent;size:500"`
	CreatedAt    time.Time  `json:"created_at" gorm:"autoCreateTime"`
	UpdatedAt    time.Time  `json:"updated_at" gorm:"autoUpdateTime"`
}

// TableName maps waitlist entries to the legacy migration table name.
func (WaitlistEntry) TableName() string {
	return "waitlist"
}

// WaitlistEntryAdminList is a stripped-down version for safe API responses to the admin dashboard
type WaitlistEntryAdminList struct {
	ID           uuid.UUID  `json:"id"`
	Email        string     `json:"email"`
	Name         string     `json:"name"`
	Company      string     `json:"company"`
	UseCase      string     `json:"use_case"`
	Source       string     `json:"source"`
	Status       string     `json:"status"`
	InviteCodeID *uuid.UUID `json:"invite_code_id,omitempty"`
	InvitedAt    *time.Time `json:"invited_at,omitempty"`
	Notes        string     `json:"notes"`
	IPAddress    string     `json:"ip_address,omitempty"`
	CreatedAt    time.Time  `json:"created_at"`
	UpdatedAt    time.Time  `json:"updated_at"`
}

// WaitlistStats contains aggregate statistics about the waitlist
type WaitlistStats struct {
	Total       int64 `json:"total"`
	Pending     int64 `json:"pending"`
	Approved    int64 `json:"approved"`
	Invited     int64 `json:"invited"`
	Rejected    int64 `json:"rejected"`
	NewToday    int64 `json:"new_today"`
	NewThisWeek int64 `json:"new_this_week"`
}

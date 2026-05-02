package domain

import (
	"context"
	"time"

	"github.com/google/uuid"
	"github.com/functionfly/functionfly/internal/storage"
)

// AuthRepository defines the interface for per-tenant authentication operations
type AuthRepository interface {
	// Tenant Auth Settings
	CreateAuthSettings(ctx context.Context, settings *storage.TenantAuthSettings) error
	GetAuthSettings(ctx context.Context, tenantID uuid.UUID) (*storage.TenantAuthSettings, error)
	UpdateAuthSettings(ctx context.Context, tenantID uuid.UUID, updates map[string]interface{}) error
	DeleteAuthSettings(ctx context.Context, tenantID uuid.UUID) error

	// Tenant OAuth Providers
	CreateOAuthProvider(ctx context.Context, provider *storage.TenantOAuthProvider) error
	GetOAuthProvider(ctx context.Context, tenantID uuid.UUID, provider string) (*storage.TenantOAuthProvider, error)
	ListOAuthProviders(ctx context.Context, tenantID uuid.UUID) ([]*storage.TenantOAuthProvider, error)
	UpdateOAuthProvider(ctx context.Context, tenantID uuid.UUID, provider string, updates map[string]interface{}) (*storage.TenantOAuthProvider, error)
	DeleteOAuthProvider(ctx context.Context, tenantID uuid.UUID, provider string) error
	GetEnabledOAuthProviders(ctx context.Context, tenantID uuid.UUID) ([]*storage.TenantOAuthProvider, error)

	// Tenant Invite Codes
	CreateInviteCode(ctx context.Context, invite *storage.TenantInviteCode) error
	GetInviteCode(ctx context.Context, code string) (*storage.TenantInviteCode, error)
	GetInviteCodesByTenant(ctx context.Context, tenantID uuid.UUID, includeUsed bool) ([]*storage.TenantInviteCode, error)
	GetInviteCodeByEmail(ctx context.Context, tenantID uuid.UUID, email string) (*storage.TenantInviteCode, error)
	AcceptInviteCode(ctx context.Context, code string, userID uuid.UUID) error
	RevokeInviteCode(ctx context.Context, code string) error
	IncrementInviteCodeUses(ctx context.Context, code string) error
	DeleteExpiredInviteCodes(ctx context.Context) (int64, error)

	// Tenant Memberships
	CreateMembership(ctx context.Context, membership *storage.TenantMembership) error
	GetMembership(ctx context.Context, tenantID, userID uuid.UUID) (*storage.TenantMembership, error)
	ListMemberships(ctx context.Context, tenantID uuid.UUID) ([]*storage.TenantMembership, error)
	ListMembershipsByRole(ctx context.Context, tenantID uuid.UUID, role string) ([]*storage.TenantMembership, error)
	UpdateMembership(ctx context.Context, tenantID, userID uuid.UUID, updates map[string]interface{}) (*storage.TenantMembership, error)
	DeleteMembership(ctx context.Context, tenantID, userID uuid.UUID) error
	UpdateMembershipLastActive(ctx context.Context, tenantID, userID uuid.UUID) error
	CountMembershipsByTenant(ctx context.Context, tenantID uuid.UUID) (int, error)

	// Auth Audit Log
	CreateAuthAuditLog(ctx context.Context, log *storage.TenantAuthAuditLog) error
	ListAuthAuditLogs(ctx context.Context, tenantID uuid.UUID, limit, offset int, actions []string, userID *uuid.UUID, since *time.Time) ([]*storage.TenantAuthAuditLog, int, error)
	GetAuthAuditLogsByUser(ctx context.Context, tenantID, userID uuid.UUID, limit int) ([]*storage.TenantAuthAuditLog, error)
	DeleteOldAuthAuditLogs(ctx context.Context, before time.Time) (int64, error)

	// Session Management
	GetActiveSessionsByTenant(ctx context.Context, tenantID uuid.UUID) ([]*storage.TenantAuthAuditLog, error)
	RevokeAllSessions(ctx context.Context, tenantID uuid.UUID) error
}

// TenantMembershipUser is a join view of membership with user details
type TenantMembershipUser struct {
	Membership *storage.TenantMembership
	User       *storage.User
}

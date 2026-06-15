package domain

import (
	"context"
	"time"

	"github.com/google/uuid"
	"github.com/functionfly/functionfly/internal/types"
)

// AuthRepository defines the interface for per-tenant authentication operations
type AuthRepository interface {
	// Tenant Auth Settings
	CreateAuthSettings(ctx context.Context, settings *types.TenantAuthSettings) error
	GetAuthSettings(ctx context.Context, tenantID uuid.UUID) (*types.TenantAuthSettings, error)
	UpdateAuthSettings(ctx context.Context, tenantID uuid.UUID, updates map[string]interface{}) error
	DeleteAuthSettings(ctx context.Context, tenantID uuid.UUID) error

	// Tenant OAuth Providers
	CreateOAuthProvider(ctx context.Context, provider *types.TenantOAuthProvider) error
	GetOAuthProvider(ctx context.Context, tenantID uuid.UUID, provider string) (*types.TenantOAuthProvider, error)
	ListOAuthProviders(ctx context.Context, tenantID uuid.UUID) ([]*types.TenantOAuthProvider, error)
	UpdateOAuthProvider(ctx context.Context, tenantID uuid.UUID, provider string, updates map[string]interface{}) (*types.TenantOAuthProvider, error)
	DeleteOAuthProvider(ctx context.Context, tenantID uuid.UUID, provider string) error
	GetEnabledOAuthProviders(ctx context.Context, tenantID uuid.UUID) ([]*types.TenantOAuthProvider, error)

	// Tenant Invite Codes
	CreateInviteCode(ctx context.Context, invite *types.TenantInviteCode) error
	GetInviteCode(ctx context.Context, code string) (*types.TenantInviteCode, error)
	GetInviteCodesByTenant(ctx context.Context, tenantID uuid.UUID, includeUsed bool) ([]*types.TenantInviteCode, error)
	GetInviteCodeByEmail(ctx context.Context, tenantID uuid.UUID, email string) (*types.TenantInviteCode, error)
	AcceptInviteCode(ctx context.Context, code string, userID uuid.UUID) error
	RevokeInviteCode(ctx context.Context, code string) error
	IncrementInviteCodeUses(ctx context.Context, code string) error
	DeleteExpiredInviteCodes(ctx context.Context) (int64, error)

	// Tenant Memberships
	CreateMembership(ctx context.Context, membership *types.TenantMembership) error
	GetMembership(ctx context.Context, tenantID, userID uuid.UUID) (*types.TenantMembership, error)
	ListMemberships(ctx context.Context, tenantID uuid.UUID) ([]*types.TenantMembership, error)
	ListMembershipsByRole(ctx context.Context, tenantID uuid.UUID, role string) ([]*types.TenantMembership, error)
	UpdateMembership(ctx context.Context, tenantID, userID uuid.UUID, updates map[string]interface{}) (*types.TenantMembership, error)
	DeleteMembership(ctx context.Context, tenantID, userID uuid.UUID) error
	UpdateMembershipLastActive(ctx context.Context, tenantID, userID uuid.UUID) error
	CountMembershipsByTenant(ctx context.Context, tenantID uuid.UUID) (int, error)
	CountMembershipsByRole(ctx context.Context, tenantID uuid.UUID, role string) (int, error)

	// Auth Audit Log
	CreateAuthAuditLog(ctx context.Context, log *types.TenantAuthAuditLog) error
	ListAuthAuditLogs(ctx context.Context, tenantID uuid.UUID, limit, offset int, actions []string, userID *uuid.UUID, since *time.Time) ([]*types.TenantAuthAuditLog, int, error)
	GetAuthAuditLogsByUser(ctx context.Context, tenantID, userID uuid.UUID, limit int) ([]*types.TenantAuthAuditLog, error)
	DeleteOldAuthAuditLogs(ctx context.Context, before time.Time) (int64, error)

	// Session Management
	GetActiveSessionsByTenant(ctx context.Context, tenantID uuid.UUID) ([]*types.TenantAuthAuditLog, error)
	RevokeAllSessions(ctx context.Context, tenantID uuid.UUID) error
}

// TenantMembershipUser is a join view of membership with user details
type TenantMembershipUser struct {
	Membership *types.TenantMembership
	User       *types.User
}

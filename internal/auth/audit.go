package auth

import (
	"context"

	"github.com/functionfly/functionfly/internal/storage"
	"github.com/google/uuid"
)

// Audit event types
const (
	AuditEventLogin                 = "login"
	AuditEventLoginFailed           = "login_failed"
	AuditEventLogout                = "logout"
	AuditEventPasswordChange        = "password_change"
	AuditEventPasswordResetRequest  = "password_reset_request"
	AuditEventPasswordResetComplete = "password_reset_complete"
	AuditEventMFASetup              = "mfa_setup"
	AuditEventMFAVerify             = "mfa_verify"
	AuditEventMFADisable            = "mfa_disable"
	AuditEventWebAuthnRegister      = "webauthn_register"
	AuditEventWebAuthnLogin         = "webauthn_login"
	AuditEventWebAuthnDelete        = "webauthn_delete"
	AuditEventSessionCreate         = "session_create"
	AuditEventSessionExpire         = "session_expire"
	AuditEventSessionRevoke         = "session_revoke"
	AuditEventAPIKeyCreate          = "api_key_create"
	AuditEventAPIKeyUse             = "api_key_use"
	AuditEventAPIKeyRevoke          = "api_key_revoke"
	AuditEventSAMLLogin             = "saml_login"
	AuditEventSCIMUserCreated       = "scim_user_created"
	AuditEventSCIMUserDeactivated   = "scim_user_deactivated"
)

// AuditService handles authentication audit logging
type AuditService struct {
	repo *storage.AuthAuditRepository
}

// NewAuditService creates a new audit service
func NewAuditService(repo *storage.AuthAuditRepository) *AuditService {
	return &AuditService{repo: repo}
}

// Log records an audit event
func (s *AuditService) Log(ctx context.Context, eventType string, tenantID, userID *uuid.UUID, eventData map[string]interface{}, ipAddress, userAgent string, success bool, failureReason string) error {
	if eventData == nil {
		eventData = make(map[string]interface{})
	}

	log := &storage.AuthAuditLog{
		TenantID:      tenantID,
		UserID:        userID,
		EventType:     eventType,
		EventData:     eventData,
		IPAddress:     ipAddress,
		UserAgent:     userAgent,
		Success:       success,
		FailureReason: failureReason,
	}

	return s.repo.Create(ctx, log)
}

// LogLogin logs a successful login event
func (s *AuditService) LogLogin(ctx context.Context, tenantID, userID *uuid.UUID, eventData map[string]interface{}, ipAddress, userAgent string) error {
	return s.Log(ctx, AuditEventLogin, tenantID, userID, eventData, ipAddress, userAgent, true, "")
}

// LogLoginFailed logs a failed login attempt
func (s *AuditService) LogLoginFailed(ctx context.Context, tenantID *uuid.UUID, userID *uuid.UUID, failureReason string, eventData map[string]interface{}, ipAddress, userAgent string) error {
	if eventData == nil {
		eventData = make(map[string]interface{})
	}
	return s.Log(ctx, AuditEventLoginFailed, tenantID, userID, eventData, ipAddress, userAgent, false, failureReason)
}

// LogLogout logs a logout event
func (s *AuditService) LogLogout(ctx context.Context, tenantID, userID *uuid.UUID, eventData map[string]interface{}, ipAddress, userAgent string) error {
	return s.Log(ctx, AuditEventLogout, tenantID, userID, eventData, ipAddress, userAgent, true, "")
}

// LogPasswordChange logs a password change event
func (s *AuditService) LogPasswordChange(ctx context.Context, tenantID, userID *uuid.UUID, success bool, failureReason string, eventData map[string]interface{}, ipAddress, userAgent string) error {
	return s.Log(ctx, AuditEventPasswordChange, tenantID, userID, eventData, ipAddress, userAgent, success, failureReason)
}

// LogPasswordResetRequest logs a password reset request
func (s *AuditService) LogPasswordResetRequest(ctx context.Context, tenantID *uuid.UUID, userID *uuid.UUID, eventData map[string]interface{}, ipAddress, userAgent string) error {
	return s.Log(ctx, AuditEventPasswordResetRequest, tenantID, userID, eventData, ipAddress, userAgent, true, "")
}

// LogPasswordResetComplete logs a password reset completion
func (s *AuditService) LogPasswordResetComplete(ctx context.Context, tenantID, userID *uuid.UUID, success bool, failureReason string, eventData map[string]interface{}, ipAddress, userAgent string) error {
	return s.Log(ctx, AuditEventPasswordResetComplete, tenantID, userID, eventData, ipAddress, userAgent, success, failureReason)
}

// LogMFASetup logs an MFA setup event
func (s *AuditService) LogMFASetup(ctx context.Context, tenantID, userID *uuid.UUID, success bool, failureReason string, eventData map[string]interface{}, ipAddress, userAgent string) error {
	return s.Log(ctx, AuditEventMFASetup, tenantID, userID, eventData, ipAddress, userAgent, success, failureReason)
}

// LogMFAVerify logs an MFA verification event
func (s *AuditService) LogMFAVerify(ctx context.Context, tenantID, userID *uuid.UUID, success bool, failureReason string, eventData map[string]interface{}, ipAddress, userAgent string) error {
	return s.Log(ctx, AuditEventMFAVerify, tenantID, userID, eventData, ipAddress, userAgent, success, failureReason)
}

// LogMFADisable logs an MFA disable event
func (s *AuditService) LogMFADisable(ctx context.Context, tenantID, userID *uuid.UUID, success bool, failureReason string, eventData map[string]interface{}, ipAddress, userAgent string) error {
	return s.Log(ctx, AuditEventMFADisable, tenantID, userID, eventData, ipAddress, userAgent, success, failureReason)
}

// LogWebAuthnRegister logs a WebAuthn credential registration
func (s *AuditService) LogWebAuthnRegister(ctx context.Context, tenantID, userID *uuid.UUID, success bool, failureReason string, eventData map[string]interface{}, ipAddress, userAgent string) error {
	return s.Log(ctx, AuditEventWebAuthnRegister, tenantID, userID, eventData, ipAddress, userAgent, success, failureReason)
}

// LogWebAuthnLogin logs a WebAuthn login attempt
func (s *AuditService) LogWebAuthnLogin(ctx context.Context, tenantID, userID *uuid.UUID, success bool, failureReason string, eventData map[string]interface{}, ipAddress, userAgent string) error {
	return s.Log(ctx, AuditEventWebAuthnLogin, tenantID, userID, eventData, ipAddress, userAgent, success, failureReason)
}

// LogWebAuthnDelete logs a WebAuthn credential deletion
func (s *AuditService) LogWebAuthnDelete(ctx context.Context, tenantID, userID *uuid.UUID, success bool, failureReason string, eventData map[string]interface{}, ipAddress, userAgent string) error {
	return s.Log(ctx, AuditEventWebAuthnDelete, tenantID, userID, eventData, ipAddress, userAgent, success, failureReason)
}

// LogSessionCreate logs a session creation event
func (s *AuditService) LogSessionCreate(ctx context.Context, tenantID, userID *uuid.UUID, eventData map[string]interface{}, ipAddress, userAgent string) error {
	return s.Log(ctx, AuditEventSessionCreate, tenantID, userID, eventData, ipAddress, userAgent, true, "")
}

// LogSessionExpire logs a session expiration event
func (s *AuditService) LogSessionExpire(ctx context.Context, tenantID, userID *uuid.UUID, eventData map[string]interface{}, ipAddress, userAgent string) error {
	return s.Log(ctx, AuditEventSessionExpire, tenantID, userID, eventData, ipAddress, userAgent, true, "")
}

// LogSessionRevoke logs a session revocation event
func (s *AuditService) LogSessionRevoke(ctx context.Context, tenantID, userID *uuid.UUID, eventData map[string]interface{}, ipAddress, userAgent string) error {
	return s.Log(ctx, AuditEventSessionRevoke, tenantID, userID, eventData, ipAddress, userAgent, true, "")
}

// LogAPIKeyCreate logs an API key creation event
func (s *AuditService) LogAPIKeyCreate(ctx context.Context, tenantID, userID *uuid.UUID, eventData map[string]interface{}, ipAddress, userAgent string) error {
	return s.Log(ctx, AuditEventAPIKeyCreate, tenantID, userID, eventData, ipAddress, userAgent, true, "")
}

// LogAPIKeyUse logs an API key usage event
func (s *AuditService) LogAPIKeyUse(ctx context.Context, tenantID, userID *uuid.UUID, eventData map[string]interface{}, ipAddress, userAgent string) error {
	return s.Log(ctx, AuditEventAPIKeyUse, tenantID, userID, eventData, ipAddress, userAgent, true, "")
}

// LogAPIKeyRevoke logs an API key revocation event
func (s *AuditService) LogAPIKeyRevoke(ctx context.Context, tenantID, userID *uuid.UUID, eventData map[string]interface{}, ipAddress, userAgent string) error {
	return s.Log(ctx, AuditEventAPIKeyRevoke, tenantID, userID, eventData, ipAddress, userAgent, true, "")
}

// LogSAMLLogin logs a SAML login event
func (s *AuditService) LogSAMLLogin(ctx context.Context, tenantID, userID *uuid.UUID, success bool, failureReason string, eventData map[string]interface{}, ipAddress, userAgent string) error {
	return s.Log(ctx, AuditEventSAMLLogin, tenantID, userID, eventData, ipAddress, userAgent, success, failureReason)
}

// LogSCIMUserCreated logs a SCIM user creation event
func (s *AuditService) LogSCIMUserCreated(ctx context.Context, tenantID, userID *uuid.UUID, eventData map[string]interface{}, ipAddress, userAgent string) error {
	return s.Log(ctx, AuditEventSCIMUserCreated, tenantID, userID, eventData, ipAddress, userAgent, true, "")
}

// LogSCIMUserDeactivated logs a SCIM user deactivation event
func (s *AuditService) LogSCIMUserDeactivated(ctx context.Context, tenantID, userID *uuid.UUID, eventData map[string]interface{}, ipAddress, userAgent string) error {
	return s.Log(ctx, AuditEventSCIMUserDeactivated, tenantID, userID, eventData, ipAddress, userAgent, true, "")
}

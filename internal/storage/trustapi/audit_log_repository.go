package trustapi

import (
	"encoding/json"
	"time"

	"github.com/google/uuid"
	"gorm.io/gorm"
)

// ============================================
// Trust Audit Log
// ============================================

// AuditAction represents the type of audit action
type AuditAction string

const (
	AuditActionRevocationCreated  AuditAction = "revocation_created"
	AuditActionRevocationLifted   AuditAction = "revocation_lifted"
	AuditActionRevocationExpired  AuditAction = "revocation_expired"
	AuditActionAttestationCreated AuditAction = "attestation_created"
	AuditActionAttestationRevoked AuditAction = "attestation_revoked"
	AuditActionPolicyCreated      AuditAction = "policy_created"
	AuditActionPolicyUpdated      AuditAction = "policy_updated"
	AuditActionPolicyDeleted      AuditAction = "policy_deleted"
	AuditActionPolicyEvaluated    AuditAction = "policy_evaluated"
	AuditActionTrustScoreUpdated  AuditAction = "trust_score_updated"
)

// TrustAuditLog represents an audit log entry for trust-related actions
type TrustAuditLog struct {
	ID uuid.UUID `json:"id" gorm:"type:uuid;primaryKey;default:gen_random_uuid()"`

	// Action details
	Action     string `json:"action" gorm:"size:50;not null"`
	EntityType string `json:"entity_type" gorm:"size:50;not null"` // revocation, attestation, policy, function
	EntityID   string `json:"entity_id" gorm:"size:255;not null"`

	// Actor information
	ActorID        uuid.UUID  `json:"actor_id" gorm:"type:uuid;not null"`
	ActorType      string     `json:"actor_type" gorm:"size:20;not null"` // user, system, partner, api_key
	ActorPartnerID *uuid.UUID `json:"actor_partner_id,omitempty" gorm:"type:uuid"`

	// Affected function (if applicable)
	FunctionID *uuid.UUID `json:"function_id,omitempty" gorm:"type:uuid"`

	// Change details
	PreviousState json.RawMessage `json:"previous_state,omitempty" gorm:"type:jsonb"`
	NewState      json.RawMessage `json:"new_state,omitempty" gorm:"type:jsonb"`
	ChangeSummary string          `json:"change_summary,omitempty" gorm:"type:text"`

	// Context
	IPAddress string `json:"ip_address,omitempty" gorm:"size:45"`
	UserAgent string `json:"user_agent,omitempty" gorm:"type:text"`
	RequestID string `json:"request_id,omitempty" gorm:"size:255"`

	// Timestamp
	CreatedAt time.Time `json:"created_at" gorm:"autoCreateTime"`
}

// TableName returns the table name for TrustAuditLog
func (TrustAuditLog) TableName() string {
	return "trust_audit_logs"
}

// AuditLogRepository handles audit log operations
type AuditLogRepository struct {
	db *gorm.DB
}

// NewAuditLogRepository creates a new audit log repository
func NewAuditLogRepository(db *gorm.DB) *AuditLogRepository {
	return &AuditLogRepository{db: db}
}

// LogRevocationCreated logs when a trust revocation is created
func (r *AuditLogRepository) LogRevocationCreated(
	revocation *TrustRevocation,
	actorID uuid.UUID,
	actorType string,
	ipAddress, userAgent, requestID string,
) error {
	newState, _ := json.Marshal(map[string]interface{}{
		"reason":               revocation.Reason,
		"severity":             revocation.Severity,
		"revocation_type":      revocation.RevocationType,
		"original_trust_score": revocation.OriginalTrustScore,
		"original_trust_tier":  revocation.OriginalTrustTier,
	})

	log := &TrustAuditLog{
		Action:        string(AuditActionRevocationCreated),
		EntityType:    "revocation",
		EntityID:      revocation.RevocationID,
		ActorID:       actorID,
		ActorType:     actorType,
		FunctionID:    &revocation.FunctionID,
		NewState:      newState,
		ChangeSummary: "Trust revoked for function",
		IPAddress:     ipAddress,
		UserAgent:     userAgent,
		RequestID:     requestID,
	}

	return r.db.Create(log).Error
}

// LogRevocationLifted logs when a trust revocation is lifted
func (r *AuditLogRepository) LogRevocationLifted(
	revocation *TrustRevocation,
	actorID uuid.UUID,
	actorType string,
	reason string,
	ipAddress, userAgent, requestID string,
) error {
	previousState, _ := json.Marshal(map[string]interface{}{
		"status":  string(RevocationStatusActive),
		"revoked": true,
	})

	newState, _ := json.Marshal(map[string]interface{}{
		"status":      string(RevocationStatusLifted),
		"lifted_by":   actorID,
		"lift_reason": reason,
		"revoked":     false,
	})

	log := &TrustAuditLog{
		Action:        string(AuditActionRevocationLifted),
		EntityType:    "revocation",
		EntityID:      revocation.RevocationID,
		ActorID:       actorID,
		ActorType:     actorType,
		FunctionID:    &revocation.FunctionID,
		PreviousState: previousState,
		NewState:      newState,
		ChangeSummary: "Trust revocation lifted: " + reason,
		IPAddress:     ipAddress,
		UserAgent:     userAgent,
		RequestID:     requestID,
	}

	return r.db.Create(log).Error
}

// LogAttestationCreated logs when an attestation is created
func (r *AuditLogRepository) LogAttestationCreated(
	attestation *TrustAttestation,
	actorID uuid.UUID,
	actorType string,
	ipAddress, userAgent, requestID string,
) error {
	newState, _ := json.Marshal(map[string]interface{}{
		"type":        attestation.Type,
		"title":       attestation.Title,
		"proof_hash":  attestation.ProofHash,
		"attested_at": attestation.AttestedAt,
	})

	log := &TrustAuditLog{
		Action:        string(AuditActionAttestationCreated),
		EntityType:    "attestation",
		EntityID:      attestation.AttestationID,
		ActorID:       actorID,
		ActorType:     actorType,
		FunctionID:    &attestation.FunctionID,
		NewState:      newState,
		ChangeSummary: "New attestation created: " + attestation.Title,
		IPAddress:     ipAddress,
		UserAgent:     userAgent,
		RequestID:     requestID,
	}

	return r.db.Create(log).Error
}

// LogAttestationRevoked logs when an attestation is revoked
func (r *AuditLogRepository) LogAttestationRevoked(
	attestation *TrustAttestation,
	actorID uuid.UUID,
	actorType string,
	reason string,
	ipAddress, userAgent, requestID string,
) error {
	previousState, _ := json.Marshal(map[string]interface{}{
		"status": "valid",
	})

	newState, _ := json.Marshal(map[string]interface{}{
		"status":        "revoked",
		"revoked_at":    time.Now(),
		"revoked_by":    actorID,
		"revoke_reason": reason,
	})

	log := &TrustAuditLog{
		Action:        string(AuditActionAttestationRevoked),
		EntityType:    "attestation",
		EntityID:      attestation.AttestationID,
		ActorID:       actorID,
		ActorType:     actorType,
		FunctionID:    &attestation.FunctionID,
		PreviousState: previousState,
		NewState:      newState,
		ChangeSummary: "Attestation revoked: " + reason,
		IPAddress:     ipAddress,
		UserAgent:     userAgent,
		RequestID:     requestID,
	}

	return r.db.Create(log).Error
}

// LogPolicyCreated logs when a trust policy is created
func (r *AuditLogRepository) LogPolicyCreated(
	policy *TrustPolicy,
	actorID uuid.UUID,
	ipAddress, userAgent, requestID string,
) error {
	newState, _ := json.Marshal(map[string]interface{}{
		"name":           policy.Name,
		"default_action": policy.DefaultAction,
		"rules_count":    len(policy.Rules),
	})

	log := &TrustAuditLog{
		Action:        string(AuditActionPolicyCreated),
		EntityType:    "policy",
		EntityID:      policy.PolicyID,
		ActorID:       actorID,
		ActorType:     "user",
		NewState:      newState,
		ChangeSummary: "Trust policy created: " + policy.Name,
		IPAddress:     ipAddress,
		UserAgent:     userAgent,
		RequestID:     requestID,
	}

	return r.db.Create(log).Error
}

// LogPolicyUpdated logs when a trust policy is updated
func (r *AuditLogRepository) LogPolicyUpdated(
	policy *TrustPolicy,
	actorID uuid.UUID,
	changes string,
	ipAddress, userAgent, requestID string,
) error {
	newState, _ := json.Marshal(map[string]interface{}{
		"version":        policy.Version,
		"status":         policy.Status,
		"default_action": policy.DefaultAction,
	})

	log := &TrustAuditLog{
		Action:        string(AuditActionPolicyUpdated),
		EntityType:    "policy",
		EntityID:      policy.PolicyID,
		ActorID:       actorID,
		ActorType:     "user",
		NewState:      newState,
		ChangeSummary: "Trust policy updated: " + changes,
		IPAddress:     ipAddress,
		UserAgent:     userAgent,
		RequestID:     requestID,
	}

	return r.db.Create(log).Error
}

// LogPolicyDeleted logs when a trust policy is deprecated (soft delete)
func (r *AuditLogRepository) LogPolicyDeleted(
	policy *TrustPolicy,
	actorID uuid.UUID,
	ipAddress, userAgent, requestID string,
) error {
	previousState, _ := json.Marshal(map[string]interface{}{
		"status": policy.Status,
	})

	newState, _ := json.Marshal(map[string]interface{}{
		"status":        "deprecated",
		"deprecated_at": time.Now(),
		"deprecated_by": actorID,
	})

	log := &TrustAuditLog{
		Action:        string(AuditActionPolicyDeleted),
		EntityType:    "policy",
		EntityID:      policy.PolicyID,
		ActorID:       actorID,
		ActorType:     "user",
		PreviousState: previousState,
		NewState:      newState,
		ChangeSummary: "Trust policy deprecated",
		IPAddress:     ipAddress,
		UserAgent:     userAgent,
		RequestID:     requestID,
	}

	return r.db.Create(log).Error
}

// LogPolicyEvaluation logs when a policy is evaluated against a function
func (r *AuditLogRepository) LogPolicyEvaluation(
	evaluation *TrustPolicyEvaluation,
	ipAddress, userAgent, requestID string,
) error {
	newState, _ := json.Marshal(map[string]interface{}{
		"result":      evaluation.Result,
		"decision":    evaluation.Decision,
		"trust_score": evaluation.TrustScore,
		"is_revoked":  evaluation.IsRevoked,
	})

	log := &TrustAuditLog{
		Action:        string(AuditActionPolicyEvaluated),
		EntityType:    "evaluation",
		EntityID:      evaluation.EvaluationID,
		ActorID:       evaluation.EvaluatedBy,
		ActorType:     evaluation.EvaluatedByType,
		FunctionID:    &evaluation.FunctionID,
		NewState:      newState,
		ChangeSummary: "Policy evaluated: " + evaluation.Result,
		IPAddress:     ipAddress,
		UserAgent:     userAgent,
		RequestID:     requestID,
	}

	return r.db.Create(log).Error
}

// LogTrustScoreUpdated logs when a function's trust score is updated
func (r *AuditLogRepository) LogTrustScoreUpdated(
	functionID uuid.UUID,
	oldScore, newScore float64,
	oldTier, newTier string,
	actorID uuid.UUID,
	actorType string,
	ipAddress, userAgent, requestID string,
) error {
	previousState, _ := json.Marshal(map[string]interface{}{
		"trust_score": oldScore,
		"trust_tier":  oldTier,
	})

	newState, _ := json.Marshal(map[string]interface{}{
		"trust_score": newScore,
		"trust_tier":  newTier,
	})

	log := &TrustAuditLog{
		Action:        string(AuditActionTrustScoreUpdated),
		EntityType:    "function",
		EntityID:      functionID.String(),
		ActorID:       actorID,
		ActorType:     actorType,
		FunctionID:    &functionID,
		PreviousState: previousState,
		NewState:      newState,
		ChangeSummary: "Trust score updated",
		IPAddress:     ipAddress,
		UserAgent:     userAgent,
		RequestID:     requestID,
	}

	return r.db.Create(log).Error
}

// ListAuditLogs lists audit logs with filtering
func (r *AuditLogRepository) ListAuditLogs(
	functionID *uuid.UUID,
	action string,
	actorID *uuid.UUID,
	startTime, endTime time.Time,
	limit, offset int,
) ([]TrustAuditLog, int64, error) {
	var logs []TrustAuditLog
	var total int64

	query := r.db.Model(&TrustAuditLog{})

	if functionID != nil {
		query = query.Where("function_id = ?", *functionID)
	}
	if action != "" {
		query = query.Where("action = ?", action)
	}
	if actorID != nil {
		query = query.Where("actor_id = ?", *actorID)
	}
	if !startTime.IsZero() {
		query = query.Where("created_at >= ?", startTime)
	}
	if !endTime.IsZero() {
		query = query.Where("created_at <= ?", endTime)
	}

	// Count total
	if err := query.Count(&total).Error; err != nil {
		return nil, 0, err
	}

	// Fetch with pagination
	if err := query.Order("created_at DESC").Limit(limit).Offset(offset).Find(&logs).Error; err != nil {
		return nil, 0, err
	}

	return logs, total, nil
}

// GetAuditLog retrieves a specific audit log
func (r *AuditLogRepository) GetAuditLog(id uuid.UUID) (*TrustAuditLog, error) {
	var log TrustAuditLog
	err := r.db.First(&log, id).Error
	if err != nil {
		return nil, err
	}
	return &log, nil
}

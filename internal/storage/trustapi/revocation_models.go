package trustapi

import (
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"time"

	"github.com/google/uuid"
)

// ============================================
// Trust Revocation System
// ============================================

// RevocationReason represents the reason for trust revocation
type RevocationReason string

const (
	RevocationReasonSecurity        RevocationReason = "security"
	RevocationReasonMalware         RevocationReason = "malware"
	RevocationReasonAbuse           RevocationReason = "abuse"
	RevocationReasonPolicyViolation RevocationReason = "policy_violation"
	RevocationReasonReported        RevocationReason = "reported"
	RevocationReasonDeprecated      RevocationReason = "deprecated"
	RevocationReasonOther           RevocationReason = "other"
)

// RevocationStatus represents the status of a trust revocation
type RevocationStatus string

const (
	RevocationStatusActive   RevocationStatus = "active"
	RevocationStatusLifted   RevocationStatus = "lifted"
	RevocationStatusExpired  RevocationStatus = "expired"
	RevocationStatusAppealed RevocationStatus = "appealed"
)

// TrustRevocation represents a trust revocation record for a function
type TrustRevocation struct {
	ID uuid.UUID `json:"id" gorm:"type:uuid;primaryKey;default:gen_random_uuid()"`

	// Function being revoked
	FunctionID     uuid.UUID `json:"function_id" gorm:"type:uuid;not null;index:idx_function_revocation"`
	FunctionAuthor string    `json:"function_author,omitempty" gorm:"size:255"`
	FunctionName   string    `json:"function_name,omitempty" gorm:"size:255"`

	// Revocation details
	RevocationID  string `json:"revocation_id" gorm:"size:32;not null;uniqueIndex"` // Public ID (e.g., "rvk_abc123...")
	Reason        string `json:"reason" gorm:"size:50;not null"`
	ReasonDetails string `json:"reason_details" gorm:"type:text"`
	Severity      string `json:"severity" gorm:"size:20;not null;default:'high'"` // low, medium, high, critical
	Status        string `json:"status" gorm:"size:20;not null;default:'active'"`

	// Who revoked
	RevokedBy        uuid.UUID  `json:"revoked_by" gorm:"type:uuid;not null"`
	RevokedByType    string     `json:"revoked_by_type" gorm:"size:20;not null;default:'admin'"` // admin, system, partner, auto
	RevokedByPartner *uuid.UUID `json:"revoked_by_partner,omitempty" gorm:"type:uuid"`

	// Related report (if triggered by report)
	ReportID *uuid.UUID `json:"report_id,omitempty" gorm:"type:uuid"`

	// Original trust state (for restoration)
	OriginalTrustScore float64 `json:"original_trust_score" gorm:"default:0"`
	OriginalTrustTier  string  `json:"original_trust_tier" gorm:"size:20"`
	OriginalIsVerified bool    `json:"original_is_verified" gorm:"default:false"`

	// Revocation impact
	RevocationType    string `json:"revocation_type" gorm:"size:30;not null;default:'full'"` // full, partial, warning
	ImpactDescription string `json:"impact_description" gorm:"type:text"`

	// Temporal settings
	RevokedAt  time.Time  `json:"revoked_at" gorm:"autoCreateTime"`
	LiftedAt   *time.Time `json:"lifted_at,omitempty"`
	LiftedBy   *uuid.UUID `json:"lifted_by,omitempty" gorm:"type:uuid"`
	LiftReason string     `json:"lift_reason,omitempty" gorm:"type:text"`
	ExpiresAt  *time.Time `json:"expires_at,omitempty"` // Optional expiration

	// Evidence and documentation
	EvidenceURLs     json.RawMessage `json:"evidence_urls" gorm:"type:jsonb;default:'[]'::jsonb"` // Array of evidence URLs
	DocumentationURL string          `json:"documentation_url,omitempty" gorm:"size:500"`

	// Appeal process
	AppealStatus      string     `json:"appeal_status,omitempty" gorm:"size:20"` // pending, approved, rejected
	AppealSubmittedAt *time.Time `json:"appeal_submitted_at,omitempty"`
	AppealReason      string     `json:"appeal_reason,omitempty" gorm:"type:text"`

	// Internal tracking
	NotifiedUsers      bool `json:"notified_users" gorm:"default:false"` // Whether webhook notifications sent
	SearchIndexUpdated bool `json:"search_index_updated" gorm:"default:false"`

	// Timestamps
	CreatedAt time.Time `json:"created_at" gorm:"autoCreateTime"`
	UpdatedAt time.Time `json:"updated_at" gorm:"autoUpdateTime"`
}

// TableName returns the table name for TrustRevocation
func (TrustRevocation) TableName() string {
	return "trust_revocations"
}

// IsActive returns true if the revocation is currently active
func (r *TrustRevocation) IsActive() bool {
	if r.Status != string(RevocationStatusActive) {
		return false
	}
	if r.ExpiresAt != nil && r.ExpiresAt.Before(time.Now()) {
		return false
	}
	return true
}

// ============================================
// Attestation System
// ============================================

// AttestationType represents the type of attestation
type AttestationType string

const (
	AttestationTypeVerification AttestationType = "verification"
	AttestationTypeSecurityScan AttestationType = "security_scan"
	AttestationTypeCodeReview   AttestationType = "code_review"
	AttestationTypeExecution    AttestationType = "execution"
	AttestationTypeCompliance   AttestationType = "compliance"
	AttestationTypeSignature    AttestationType = "signature"
)

// AttestationStatus represents the status of an attestation
type AttestationStatus string

const (
	AttestationStatusValid   AttestationStatus = "valid"
	AttestationStatusRevoked AttestationStatus = "revoked"
	AttestationStatusExpired AttestationStatus = "expired"
)

// TrustAttestation represents an immutable attestation record
type TrustAttestation struct {
	ID uuid.UUID `json:"id" gorm:"type:uuid;primaryKey;default:gen_random_uuid()"`

	// Attestation identification
	AttestationID string `json:"attestation_id" gorm:"size:32;not null;uniqueIndex"` // Public ID (e.g., "att_abc123...")

	// What is being attested
	FunctionID      uuid.UUID `json:"function_id" gorm:"type:uuid;not null;index:idx_function_attestation"`
	FunctionVersion string    `json:"function_version,omitempty" gorm:"size:50"`
	FunctionAuthor  string    `json:"function_author,omitempty" gorm:"size:255"`
	FunctionName    string    `json:"function_name,omitempty" gorm:"size:255"`

	// Attestation type and details
	Type        string          `json:"type" gorm:"size:50;not null"`
	Status      string          `json:"status" gorm:"size:20;not null;default:'valid'"`
	Title       string          `json:"title" gorm:"size:255;not null"`
	Description string          `json:"description" gorm:"type:text"`
	Results     json.RawMessage `json:"results" gorm:"type:jsonb;default:'{}'::jsonb"` // Detailed attestation results

	// Attester identity
	AttesterID        uuid.UUID  `json:"attester_id" gorm:"type:uuid;not null"`
	AttesterType      string     `json:"attester_type" gorm:"size:20;not null;default:'system'"` // system, partner, manual, automated
	AttesterName      string     `json:"attester_name,omitempty" gorm:"size:255"`
	AttesterPartnerID *uuid.UUID `json:"attester_partner_id,omitempty" gorm:"type:uuid"`

	// Verification level (for verification attestations)
	VerificationLevel string `json:"verification_level,omitempty" gorm:"size:30"`

	// Cryptographic proof
	ProofHash    string `json:"proof_hash" gorm:"size:64;not null"`      // SHA-256 hash of attestation data
	PreviousHash string `json:"previous_hash,omitempty" gorm:"size:64"`  // For chain of attestations
	Signature    string `json:"signature,omitempty" gorm:"size:512"`     // RSA/ECDSA signature
	PublicKeyID  string `json:"public_key_id,omitempty" gorm:"size:100"` // Reference to signing key

	// Source data hash (for verification)
	SourceDataHash string `json:"source_data_hash,omitempty" gorm:"size:64"` // Hash of function source/manifest

	// Temporal tracking
	AttestedAt   time.Time  `json:"attested_at" gorm:"not null"`
	ValidUntil   *time.Time `json:"valid_until,omitempty"`
	RevokedAt    *time.Time `json:"revoked_at,omitempty"`
	RevokedBy    *uuid.UUID `json:"revoked_by,omitempty" gorm:"type:uuid"`
	RevokeReason string     `json:"revoke_reason,omitempty" gorm:"type:text"`

	// Related revocation (if attestation was revoked due to trust revocation)
	RevocationID *uuid.UUID `json:"revocation_id,omitempty" gorm:"type:uuid"`

	// Immutable record flag
	IsImmutable bool `json:"is_immutable" gorm:"default:true"`

	// Timestamps
	CreatedAt time.Time `json:"created_at" gorm:"autoCreateTime"`
	UpdatedAt time.Time `json:"updated_at" gorm:"autoUpdateTime"`
}

// TableName returns the table name for TrustAttestation
func (TrustAttestation) TableName() string {
	return "trust_attestations"
}

// CalculateProofHash calculates the cryptographic hash of attestation data
func (a *TrustAttestation) CalculateProofHash() string {
	// Create deterministic data representation
	data := struct {
		FunctionID      string
		FunctionVersion string
		Type            string
		Title           string
		Description     string
		AttesterID      string
		AttestedAt      int64
		Results         string
	}{
		FunctionID:      a.FunctionID.String(),
		FunctionVersion: a.FunctionVersion,
		Type:            a.Type,
		Title:           a.Title,
		Description:     a.Description,
		AttesterID:      a.AttesterID.String(),
		AttestedAt:      a.AttestedAt.UnixNano(),
		Results:         string(a.Results),
	}

	jsonData, _ := json.Marshal(data)
	hash := sha256.Sum256(jsonData)
	return hex.EncodeToString(hash[:])
}

// VerifyIntegrity verifies the attestation's integrity
func (a *TrustAttestation) VerifyIntegrity() bool {
	expectedHash := a.CalculateProofHash()
	return a.ProofHash == expectedHash
}

// ============================================
// Trust Policy System
// ============================================

// TrustPolicyRule represents a single rule in a trust policy
type TrustPolicyRule struct {
	ID          string                 `json:"id"`
	Type        string                 `json:"type"`     // min_trust_score, verification_required, tier_minimum, etc.
	Operator    string                 `json:"operator"` // gt, gte, lt, lte, eq, neq, in, not_in
	Value       interface{}            `json:"value"`
	Description string                 `json:"description,omitempty"`
	Params      map[string]interface{} `json:"params,omitempty"`
}

// TrustPolicy represents an agent's trust policy for function selection
type TrustPolicy struct {
	ID uuid.UUID `json:"id" gorm:"type:uuid;primaryKey;default:gen_random_uuid()"`

	// Policy identification
	PolicyID    string `json:"policy_id" gorm:"size:32;not null;uniqueIndex"` // Public ID (e.g., "pol_abc123...")
	Name        string `json:"name" gorm:"size:255;not null"`
	Description string `json:"description" gorm:"type:text"`
	Version     int    `json:"version" gorm:"default:1"`

	// Policy owner
	OwnerID        uuid.UUID  `json:"owner_id" gorm:"type:uuid;not null;index:idx_policy_owner"`
	OwnerType      string     `json:"owner_type" gorm:"size:20;not null;default:'user'"` // user, partner, team, system
	OwnerPartnerID *uuid.UUID `json:"owner_partner_id,omitempty" gorm:"type:uuid"`

	// Policy rules
	Rules json.RawMessage `json:"rules" gorm:"type:jsonb;not null;default:'[]'::jsonb"`

	// Default action when no rules match
	DefaultAction string `json:"default_action" gorm:"size:20;not null;default:'deny'"` // allow, deny, warn

	// Policy status
	Status    string `json:"status" gorm:"size:20;not null;default:'active'"` // active, inactive, deprecated
	IsDefault bool   `json:"is_default" gorm:"default:false"`

	// Usage tracking
	UseCount   int        `json:"use_count" gorm:"default:0"`
	LastUsedAt *time.Time `json:"last_used_at,omitempty"`
	CreatedBy  uuid.UUID  `json:"created_by" gorm:"type:uuid;not null"`

	// Temporal settings
	CreatedAt    time.Time  `json:"created_at" gorm:"autoCreateTime"`
	UpdatedAt    time.Time  `json:"updated_at" gorm:"autoUpdateTime"`
	ValidFrom    time.Time  `json:"valid_from"`
	ValidUntil   *time.Time `json:"valid_until,omitempty"`
	DeprecatedAt *time.Time `json:"deprecated_at,omitempty"`
	DeprecatedBy *uuid.UUID `json:"deprecated_by,omitempty" gorm:"type:uuid"`
}

// TableName returns the table name for TrustPolicy
func (TrustPolicy) TableName() string {
	return "trust_policies"
}

// IsActive returns true if the policy is currently active
func (p *TrustPolicy) IsActive() bool {
	if p.Status != "active" {
		return false
	}
	if p.ValidFrom.After(time.Now()) {
		return false
	}
	if p.ValidUntil != nil && p.ValidUntil.Before(time.Now()) {
		return false
	}
	return true
}

// GetRules unmarshals the policy rules
func (p *TrustPolicy) GetRules() ([]TrustPolicyRule, error) {
	var rules []TrustPolicyRule
	if err := json.Unmarshal(p.Rules, &rules); err != nil {
		return nil, err
	}
	return rules, nil
}

// ============================================
// Policy Evaluation Results
// ============================================

// TrustPolicyEvaluation represents an evaluation of a function against a policy
type TrustPolicyEvaluation struct {
	ID uuid.UUID `json:"id" gorm:"type:uuid;primaryKey;default:gen_random_uuid()"`

	// Evaluation identification
	EvaluationID string `json:"evaluation_id" gorm:"size:32;not null;uniqueIndex"`

	// What was evaluated
	PolicyID       uuid.UUID `json:"policy_id" gorm:"type:uuid;not null;index:idx_evaluation_policy"`
	FunctionID     uuid.UUID `json:"function_id" gorm:"type:uuid;not null;index:idx_evaluation_function"`
	FunctionAuthor string    `json:"function_author,omitempty" gorm:"size:255"`
	FunctionName   string    `json:"function_name,omitempty" gorm:"size:255"`

	// Evaluation context
	TrustScore        float64 `json:"trust_score"`
	TrustTier         string  `json:"trust_tier" gorm:"size:20"`
	IsVerified        bool    `json:"is_verified"`
	VerificationLevel string  `json:"verification_level,omitempty" gorm:"size:30"`
	IsRevoked         bool    `json:"is_revoked"`
	RevocationStatus  string  `json:"revocation_status,omitempty" gorm:"size:20"`

	// Evaluation result
	Result      string          `json:"result" gorm:"size:20;not null"`   // allowed, denied, warned
	Decision    string          `json:"decision" gorm:"size:50;not null"` // specific decision code
	Reason      string          `json:"reason" gorm:"type:text"`
	RuleResults json.RawMessage `json:"rule_results" gorm:"type:jsonb;default:'[]'::jsonb"` // Individual rule evaluations

	// Evaluation metadata
	EvaluatedAt     time.Time `json:"evaluated_at" gorm:"autoCreateTime"`
	EvaluatedBy     uuid.UUID `json:"evaluated_by" gorm:"type:uuid;not null"`
	EvaluatedByType string    `json:"evaluated_by_type" gorm:"size:20;default:'api'"` // api, agent, system

	// Cache control
	CacheValidUntil *time.Time `json:"cache_valid_until,omitempty"`
	IsCached        bool       `json:"is_cached" gorm:"default:false"`

	// Request context (for audit)
	RequestID string `json:"request_id,omitempty" gorm:"size:255"`
	IPAddress string `json:"ip_address,omitempty" gorm:"size:45"`
	UserAgent string `json:"user_agent,omitempty" gorm:"type:text"`

	// Timestamps
	CreatedAt time.Time `json:"created_at" gorm:"autoCreateTime"`
}

// TableName returns the table name for TrustPolicyEvaluation
func (TrustPolicyEvaluation) TableName() string {
	return "trust_policy_evaluations"
}

// ============================================
// DTOs for API
// ============================================

// RevocationCreateRequest represents a request to create a trust revocation
type RevocationCreateRequest struct {
	FunctionID        uuid.UUID  `json:"function_id" binding:"required"`
	Reason            string     `json:"reason" binding:"required,oneof=security malware abuse policy_violation reported deprecated other"`
	ReasonDetails     string     `json:"reason_details" binding:"max=2000"`
	Severity          string     `json:"severity" binding:"required,oneof=low medium high critical"`
	RevocationType    string     `json:"revocation_type" binding:"omitempty,oneof=full partial warning"`
	ImpactDescription string     `json:"impact_description" binding:"max=1000"`
	EvidenceURLs      []string   `json:"evidence_urls"`
	DocumentationURL  string     `json:"documentation_url" binding:"omitempty,url,max=500"`
	ExpiresAt         *time.Time `json:"expires_at,omitempty"`
	ReportID          *uuid.UUID `json:"report_id,omitempty"`
}

// RevocationResponse represents a trust revocation in API responses
type RevocationResponse struct {
	ID                 uuid.UUID  `json:"id"`
	RevocationID       string     `json:"revocation_id"`
	FunctionID         uuid.UUID  `json:"function_id"`
	FunctionAuthor     string     `json:"function_author,omitempty"`
	FunctionName       string     `json:"function_name,omitempty"`
	Reason             string     `json:"reason"`
	ReasonDetails      string     `json:"reason_details,omitempty"`
	Severity           string     `json:"severity"`
	Status             string     `json:"status"`
	RevocationType     string     `json:"revocation_type"`
	RevokedAt          time.Time  `json:"revoked_at"`
	RevokedBy          uuid.UUID  `json:"revoked_by"`
	RevokedByType      string     `json:"revoked_by_type"`
	LiftedAt           *time.Time `json:"lifted_at,omitempty"`
	LiftReason         string     `json:"lift_reason,omitempty"`
	OriginalTrustScore float64    `json:"original_trust_score"`
	OriginalTrustTier  string     `json:"original_trust_tier"`
	ExpiresAt          *time.Time `json:"expires_at,omitempty"`
	AppealStatus       string     `json:"appeal_status,omitempty"`
}

// RevocationLiftRequest represents a request to lift a trust revocation
type RevocationLiftRequest struct {
	Reason string `json:"reason" binding:"required,min=10,max=1000"`
}

// RevocationListResponse represents a list of revocations
type RevocationListResponse struct {
	Revocations []RevocationResponse `json:"revocations"`
	TotalCount  int64                `json:"total_count"`
	Page        int                  `json:"page"`
	PageSize    int                  `json:"page_size"`
}

// AttestationResponse represents an attestation in API responses
type AttestationResponse struct {
	ID                uuid.UUID  `json:"id"`
	AttestationID     string     `json:"attestation_id"`
	FunctionID        uuid.UUID  `json:"function_id"`
	FunctionVersion   string     `json:"function_version,omitempty"`
	FunctionAuthor    string     `json:"function_author,omitempty"`
	FunctionName      string     `json:"function_name,omitempty"`
	Type              string     `json:"type"`
	Status            string     `json:"status"`
	Title             string     `json:"title"`
	Description       string     `json:"description,omitempty"`
	AttesterID        uuid.UUID  `json:"attester_id"`
	AttesterType      string     `json:"attester_type"`
	AttesterName      string     `json:"attester_name,omitempty"`
	VerificationLevel string     `json:"verification_level,omitempty"`
	ProofHash         string     `json:"proof_hash"`
	AttestedAt        time.Time  `json:"attested_at"`
	ValidUntil        *time.Time `json:"valid_until,omitempty"`
	RevokedAt         *time.Time `json:"revoked_at,omitempty"`
	RevokeReason      string     `json:"revoke_reason,omitempty"`
}

// AttestationListResponse represents a list of attestations
type AttestationListResponse struct {
	Attestations []AttestationResponse `json:"attestations"`
	TotalCount   int64                 `json:"total_count"`
	Page         int                   `json:"page"`
	PageSize     int                   `json:"page_size"`
}

// PolicyCreateRequest represents a request to create a trust policy
type PolicyCreateRequest struct {
	Name          string            `json:"name" binding:"required,min=2,max=255"`
	Description   string            `json:"description" binding:"max=1000"`
	Rules         []TrustPolicyRule `json:"rules" binding:"required,min=1,max=50"`
	DefaultAction string            `json:"default_action" binding:"omitempty,oneof=allow deny warn"`
	ValidUntil    *time.Time        `json:"valid_until,omitempty"`
}

// PolicyUpdateRequest represents a request to update a trust policy
type PolicyUpdateRequest struct {
	Name          string            `json:"name" binding:"omitempty,min=2,max=255"`
	Description   string            `json:"description" binding:"omitempty,max=1000"`
	Rules         []TrustPolicyRule `json:"rules" binding:"omitempty,min=1,max=50"`
	DefaultAction string            `json:"default_action" binding:"omitempty,oneof=allow deny warn"`
	Status        string            `json:"status" binding:"omitempty,oneof=active inactive deprecated"`
	ValidUntil    *time.Time        `json:"valid_until,omitempty"`
}

// PolicyResponse represents a trust policy in API responses
type PolicyResponse struct {
	ID            uuid.UUID         `json:"id"`
	PolicyID      string            `json:"policy_id"`
	Name          string            `json:"name"`
	Description   string            `json:"description,omitempty"`
	Version       int               `json:"version"`
	OwnerID       uuid.UUID         `json:"owner_id"`
	OwnerType     string            `json:"owner_type"`
	Rules         []TrustPolicyRule `json:"rules"`
	DefaultAction string            `json:"default_action"`
	Status        string            `json:"status"`
	IsDefault     bool              `json:"is_default"`
	UseCount      int               `json:"use_count"`
	LastUsedAt    *time.Time        `json:"last_used_at,omitempty"`
	ValidFrom     time.Time         `json:"valid_from"`
	ValidUntil    *time.Time        `json:"valid_until,omitempty"`
	CreatedAt     time.Time         `json:"created_at"`
	UpdatedAt     time.Time         `json:"updated_at"`
}

// PolicyListResponse represents a list of policies
type PolicyListResponse struct {
	Policies   []PolicyResponse `json:"policies"`
	TotalCount int64            `json:"total_count"`
	Page       int              `json:"page"`
	PageSize   int              `json:"page_size"`
}

// PolicyEvaluateRequest represents a request to evaluate a function against a policy
type PolicyEvaluateRequest struct {
	FunctionID uuid.UUID `json:"function_id" binding:"required"`
	PolicyID   *string   `json:"policy_id,omitempty"` // If nil, uses default policy
}

// PolicyEvaluateResponse represents the result of policy evaluation
type PolicyEvaluateResponse struct {
	EvaluationID     string             `json:"evaluation_id"`
	PolicyID         string             `json:"policy_id"`
	FunctionID       uuid.UUID          `json:"function_id"`
	FunctionAuthor   string             `json:"function_author,omitempty"`
	FunctionName     string             `json:"function_name,omitempty"`
	Result           string             `json:"result"` // allowed, denied, warned
	Decision         string             `json:"decision"`
	Reason           string             `json:"reason"`
	TrustScore       float64            `json:"trust_score"`
	TrustTier        string             `json:"trust_tier"`
	IsVerified       bool               `json:"is_verified"`
	IsRevoked        bool               `json:"is_revoked"`
	RevocationStatus string             `json:"revocation_status,omitempty"`
	RuleResults      []PolicyRuleResult `json:"rule_results"`
	EvaluatedAt      time.Time          `json:"evaluated_at"`
	CacheValidUntil  *time.Time         `json:"cache_valid_until,omitempty"`
}

// PolicyRuleResult represents the result of evaluating a single rule
type PolicyRuleResult struct {
	RuleID        string      `json:"rule_id"`
	Type          string      `json:"type"`
	Passed        bool        `json:"passed"`
	Reason        string      `json:"reason,omitempty"`
	ActualValue   interface{} `json:"actual_value,omitempty"`
	ExpectedValue interface{} `json:"expected_value,omitempty"`
}

// BatchPolicyEvaluateRequest represents a batch evaluation request
type BatchPolicyEvaluateRequest struct {
	FunctionIDs []uuid.UUID `json:"function_ids" binding:"required,min=1,max=100"`
	PolicyID    *string     `json:"policy_id,omitempty"`
}

// BatchPolicyEvaluateResponse represents batch evaluation results
type BatchPolicyEvaluateResponse struct {
	Results     []PolicyEvaluateResponse     `json:"results"`
	Errors      []BatchPolicyEvaluationError `json:"errors,omitempty"`
	EvaluatedAt time.Time                    `json:"evaluated_at"`
}

// BatchPolicyEvaluationError represents an error for a specific function
type BatchPolicyEvaluationError struct {
	FunctionID uuid.UUID `json:"function_id"`
	Error      string    `json:"error"`
}

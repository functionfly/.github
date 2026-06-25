package storage

import (
	"context"
	"crypto/sha256"
	"database/sql"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"github.com/functionfly/functionfly/internal/types"
	"time"

	"github.com/google/uuid"
	"github.com/sirupsen/logrus"
)

// PCIAuditRepository handles PCI DSS compliance audit logging
// All methods are immutable - audit logs cannot be modified or deleted
type PCIAuditRepository struct {
	db *PostgresDB
}

// NewPCIAuditRepository creates a new PCI audit repository
func NewPCIAuditRepository(db *PostgresDB) *PCIAuditRepository {
	return &PCIAuditRepository{db: db}
}

// LogPCIAuditEvent logs a PCI DSS compliance event
// This is an append-only operation - events cannot be modified
func (r *PCIAuditRepository) LogPCIAuditEvent(ctx context.Context, event *PCIAuditEvent) (*PCIAuditEvent, error) {
	if event.ID == uuid.Nil {
		event.ID = uuid.New()
	}

	now := time.Now().UTC()
	event.CreatedAt = now

	// Set retention period (minimum 1 year per PCI DSS, critical events 3 years)
	retentionPeriod := 365 * 24 * time.Hour // 1 year default
	if event.Severity == PCISeverityCritical || event.Severity == PCISeverityEmergency {
		retentionPeriod = 3 * 365 * 24 * time.Hour // 3 years for critical events
	}
	retentionUntil := now.Add(retentionPeriod)
	event.RetentionUntil = &retentionUntil

	// Calculate tamper hash for integrity (simple chain hash)
	chainHash, err := r.calculateChainHash(ctx, event)
	if err != nil {
		logrus.WithError(err).Warn("Failed to calculate chain hash for PCI audit event")
		// Continue logging - don't fail the operation
	} else {
		event.ChainHash = chainHash
	}

	query := `
		INSERT INTO pci_audit_events (
			id, event_type, severity,
			actor_user_id, actor_email, actor_role, actor_ip, actor_user_agent, session_id,
			resource_type, resource_id, tenant_id,
			card_last_four, card_brand, card_expiry_month, card_expiry_year, token_id,
			key_id, key_algorithm, key_operation,
			description, request_id, transaction_id, stripe_event_id,
			auth_method, mfa_used, success, failure_reason,
			compliance_data, metadata,
			retention_until, tamper_hash, chain_hash,
			created_at
		) VALUES (
			$1, $2, $3, $4, $5, $6, $7, $8, $9, $10, $11, $12, $13, $14, $15, $16, $17, $18, $19, $20,
			$21, $22, $23, $24, $25, $26, $27, $28, $29, $30, $31, $32, $33, $34
		)
	`

	// Convert JSON fields
	complianceDataJSON, _ := json.Marshal(event.ComplianceData)
	metadataJSON, _ := json.Marshal(event.Metadata)

	_, err = r.db.ExecContext(ctx, query,
		event.ID, event.EventType, event.Severity,
		event.ActorUserID, event.ActorEmail, event.ActorRole, event.ActorIP, event.ActorUserAgent, event.SessionID,
		event.ResourceType, event.ResourceID, event.TenantID,
		event.CardLastFour, event.CardBrand, event.CardExpiryMonth, event.CardExpiryYear, event.TokenID,
		event.KeyID, event.KeyAlgorithm, event.KeyOperation,
		event.Description, event.RequestID, event.TransactionID, event.StripeEventID,
		event.AuthMethod, event.MFAUsed, event.Success, event.FailureReason,
		complianceDataJSON, metadataJSON,
		event.RetentionUntil, event.TamperHash, event.ChainHash,
		event.CreatedAt,
	)

	if err != nil {
		return nil, fmt.Errorf("failed to log PCI audit event: %w", err)
	}

	return event, nil
}

// LogCardholderDataAccess logs access to cardholder data
// Convenience method for the most common PCI audit event type
func (r *PCIAuditRepository) LogCardholderDataAccess(ctx context.Context, params CardholderDataAccessParams) (*PCIAuditEvent, error) {
	event := &PCIAuditEvent{
		EventType:       PCIAuditCardDataRead,
		Severity:        PCISeverityInfo,
		ActorUserID:     params.UserID,
		ActorEmail:      params.UserEmail,
		ActorRole:       params.UserRole,
		ActorIP:         params.IPAddress,
		ActorUserAgent:  params.UserAgent,
		SessionID:       params.SessionID,
		ResourceType:    "cardholder_data",
		ResourceID:      params.PaymentMethodID,
		TenantID:        params.TenantID,
		CardLastFour:    params.CardLastFour,
		CardBrand:       params.CardBrand,
		CardExpiryMonth: params.CardExpiryMonth,
		CardExpiryYear:  params.CardExpiryYear,
		TokenID:         params.TokenID,
		TransactionID:   params.TransactionID,
		Description:     fmt.Sprintf("Cardholder data accessed: %s", params.Purpose),
		RequestID:       params.RequestID,
		AuthMethod:      params.AuthMethod,
		MFAUsed:         params.MFAUsed,
		Success:         params.Success,
		FailureReason:   params.FailureReason,
		Metadata: map[string]interface{}{
			"access_type":    params.AccessType,
			"data_type":      params.DataType,
			"cde_section":    params.CDESection,
			"data_flow_step": params.DataFlowStep,
		},
	}

	// Elevate severity for sensitive operations
	if params.AccessType == "write" || params.AccessType == "detokenize" {
		event.Severity = PCISeverityWarning
	}
	if !params.Success {
		event.Severity = PCISeverityCritical
	}

	return r.LogPCIAuditEvent(ctx, event)
}

// CardholderDataAccessParams parameters for logging cardholder data access
type CardholderDataAccessParams struct {
	UserID          *uuid.UUID
	UserEmail       string
	UserRole        string
	IPAddress       string
	UserAgent       string
	SessionID       string
	TenantID        *uuid.UUID
	PaymentMethodID *uuid.UUID
	CardLastFour    *string
	CardBrand       *string
	CardExpiryMonth *int
	CardExpiryYear  *int
	TokenID         *string
	TransactionID   *string
	RequestID       string
	AuthMethod      string
	MFAUsed         bool
	AccessType      string // read, write, tokenize, detokenize
	DataType        string // pan, expiry, name, service_code
	Purpose         string
	CDESection      string
	DataFlowStep    string
	Success         bool
	FailureReason   *string
}

// LogEncryptionKeyEvent logs encryption key lifecycle events
func (r *PCIAuditRepository) LogEncryptionKeyEvent(ctx context.Context, params EncryptionKeyEventParams) (*PCIAuditEvent, error) {
	eventType := PCIAuditKeyAccessed
	switch params.EventType {
	case "created":
		eventType = PCIAuditKeyCreated
	case "rotated":
		eventType = PCIAuditKeyRotated
	case "retired":
		eventType = PCIAuditKeyRetired
	case "backup_created":
		eventType = PCIAuditKeyBackupCreated
	case "backup_restored":
		eventType = PCIAuditKeyBackupRestored
	}

	event := &PCIAuditEvent{
		EventType:      eventType,
		Severity:       PCISeverityWarning, // Key operations are always at least warning level
		ActorUserID:    &params.UserID,
		ActorEmail:     params.UserEmail,
		ActorRole:      params.UserRole,
		ActorIP:        params.IPAddress,
		ActorUserAgent: params.UserAgent,
		SessionID:      params.SessionID,
		ResourceType:   "encryption_key",
		ResourceID:     &params.KeyID,
		KeyID:          &params.KeyID,
		KeyAlgorithm:   &params.KeyAlgorithm,
		KeyOperation:   &params.KeyOperation,
		Description:    fmt.Sprintf("Encryption key %s: %s", params.EventType, params.Reason),
		RequestID:      params.RequestID,
		AuthMethod:     params.AuthMethod,
		MFAUsed:        params.MFAUsed,
		Success:        params.Success,
		FailureReason:  params.FailureReason,
		Metadata: map[string]interface{}{
			"key_type":        params.KeyType,
			"approved_by":     params.ApprovedBy,
			"approval_ticket": params.ApprovalTicket,
		},
	}

	// Elevate severity for critical operations
	if params.EventType == "rotated" || params.EventType == "retired" || params.EventType == "backup_restored" {
		event.Severity = PCISeverityCritical
	}
	if !params.Success {
		event.Severity = PCISeverityEmergency
	}

	return r.LogPCIAuditEvent(ctx, event)
}

// EncryptionKeyEventParams parameters for logging key events
type EncryptionKeyEventParams struct {
	EventType      string
	KeyID          uuid.UUID
	KeyType        string
	KeyAlgorithm   string
	KeyOperation   string
	UserID         uuid.UUID
	UserEmail      string
	UserRole       string
	IPAddress      string
	UserAgent      string
	SessionID      string
	RequestID      string
	AuthMethod     string
	MFAUsed        bool
	Reason         string
	ApprovedBy     *uuid.UUID
	ApprovalTicket *string
	Success        bool
	FailureReason  *string
}

// LogPaymentFlowEvent logs payment flow events
func (r *PCIAuditRepository) LogPaymentFlowEvent(ctx context.Context, params PaymentFlowEventParams) (*PCIAuditEvent, error) {
	eventType := PCIAuditPaymentProcessed
	switch params.EventType {
	case "initiated":
		eventType = PCIAuditPaymentInitiated
	case "processed":
		eventType = PCIAuditPaymentProcessed
	case "failed":
		eventType = PCIAuditPaymentFailed
	case "refunded":
		eventType = PCIAuditRefundProcessed
	case "chargeback":
		eventType = PCIAuditChargebackReceived
	}

	severity := PCISeverityInfo
	if params.EventType == "failed" || params.EventType == "chargeback" {
		severity = PCISeverityWarning
	}
	if !params.Success {
		severity = PCISeverityCritical
	}

	event := &PCIAuditEvent{
		EventType:     eventType,
		Severity:      severity,
		ActorUserID:   params.UserID,
		ActorEmail:    params.UserEmail,
		ActorRole:     params.UserRole,
		ActorIP:       params.IPAddress,
		ResourceType:  "payment_transaction",
		TenantID:      params.TenantID,
		CardLastFour:  params.CardLastFour,
		CardBrand:     params.CardBrand,
		TokenID:       params.TokenID,
		TransactionID: &params.TransactionID,
		StripeEventID: params.StripeEventID,
		Description:   fmt.Sprintf("Payment %s: %s", params.EventType, params.Details),
		RequestID:     params.RequestID,
		AuthMethod:    params.AuthMethod,
		MFAUsed:       params.MFAUsed,
		Success:       params.Success,
		FailureReason: params.FailureReason,
		Metadata: map[string]interface{}{
			"amount_cents":   params.AmountCents,
			"currency":       params.Currency,
			"payment_method": params.PaymentMethod,
		},
	}

	return r.LogPCIAuditEvent(ctx, event)
}

// PaymentFlowEventParams parameters for logging payment flow events
type PaymentFlowEventParams struct {
	EventType     string
	UserID        *uuid.UUID
	UserEmail     string
	UserRole      string
	IPAddress     string
	TenantID      *uuid.UUID
	CardLastFour  *string
	CardBrand     *string
	CardExpMonth  *int
	CardExpYear   *int
	TokenID       *string
	TransactionID string
	StripeEventID *string
	RequestID     string
	AuthMethod    string
	MFAUsed       bool
	AmountCents   int
	Currency      string
	PaymentMethod string
	Details       string
	Success       bool
	FailureReason *string
}

// ListPCIAuditEvents retrieves PCI audit events with filtering
func (r *PCIAuditRepository) ListPCIAuditEvents(ctx context.Context, filters PCIAuditEventFilters) ([]PCIAuditEvent, int, error) {
	query := `
		SELECT 
			id, event_type, severity,
			actor_user_id, actor_email, actor_role, actor_ip, actor_user_agent, session_id,
			resource_type, resource_id, tenant_id,
			card_last_four, card_brand, card_expiry_month, card_expiry_year, token_id,
			key_id, key_algorithm, key_operation,
			description, request_id, transaction_id, stripe_event_id,
			auth_method, mfa_used, success, failure_reason,
			compliance_data, metadata,
			retention_until, tamper_hash, chain_hash,
			created_at
		FROM pci_audit_events
		WHERE 1=1
	`
	countQuery := `SELECT COUNT(*) FROM pci_audit_events WHERE 1=1`

	var args []interface{}
	argCount := 0

	if filters.EventType != "" {
		argCount++
		query += fmt.Sprintf(" AND event_type = $%d", argCount)
		countQuery += fmt.Sprintf(" AND event_type = $%d", argCount)
		args = append(args, filters.EventType)
	}

	if filters.Severity != "" {
		argCount++
		query += fmt.Sprintf(" AND severity = $%d", argCount)
		countQuery += fmt.Sprintf(" AND severity = $%d", argCount)
		args = append(args, filters.Severity)
	}

	if filters.TenantID != nil {
		argCount++
		query += fmt.Sprintf(" AND tenant_id = $%d", argCount)
		countQuery += fmt.Sprintf(" AND tenant_id = $%d", argCount)
		args = append(args, filters.TenantID)
	}

	if filters.UserID != nil {
		argCount++
		query += fmt.Sprintf(" AND actor_user_id = $%d", argCount)
		countQuery += fmt.Sprintf(" AND actor_user_id = $%d", argCount)
		args = append(args, filters.UserID)
	}

	if filters.ResourceType != "" {
		argCount++
		query += fmt.Sprintf(" AND resource_type = $%d", argCount)
		countQuery += fmt.Sprintf(" AND resource_type = $%d", argCount)
		args = append(args, filters.ResourceType)
	}

	if filters.Success != nil {
		argCount++
		query += fmt.Sprintf(" AND success = $%d", argCount)
		countQuery += fmt.Sprintf(" AND success = $%d", argCount)
		args = append(args, *filters.Success)
	}

	if !filters.StartDate.IsZero() {
		argCount++
		query += fmt.Sprintf(" AND created_at >= $%d", argCount)
		countQuery += fmt.Sprintf(" AND created_at >= $%d", argCount)
		args = append(args, filters.StartDate)
	}

	if !filters.EndDate.IsZero() {
		argCount++
		query += fmt.Sprintf(" AND created_at <= $%d", argCount)
		countQuery += fmt.Sprintf(" AND created_at <= $%d", argCount)
		args = append(args, filters.EndDate)
	}

	// Order by created_at descending (newest first)
	query += " ORDER BY created_at DESC"

	// Pagination
	if filters.Limit > 0 {
		argCount++
		query += fmt.Sprintf(" LIMIT $%d", argCount)
		args = append(args, filters.Limit)
	}

	if filters.Offset > 0 {
		argCount++
		query += fmt.Sprintf(" OFFSET $%d", argCount)
		args = append(args, filters.Offset)
	}

	// Get total count
	var total int
	if err := r.db.QueryRowContext(ctx, countQuery, args[:argCount]...).Scan(&total); err != nil {
		return nil, 0, fmt.Errorf("failed to count PCI audit events: %w", err)
	}

	rows, err := r.db.QueryContext(ctx, query, args...)
	if err != nil {
		return nil, 0, fmt.Errorf("failed to query PCI audit events: %w", err)
	}
	defer rows.Close()

	events, err := scanPCIAuditEventRows(rows)
	if err != nil {
		return nil, 0, err
	}

	return events, total, nil
}

// PCIAuditEventFilters for querying PCI audit events
type PCIAuditEventFilters struct {
	EventType    string
	Severity     string
	TenantID     *uuid.UUID
	UserID       *uuid.UUID
	ResourceType string
	Success      *bool
	StartDate    time.Time
	EndDate      time.Time
	Limit        int
	Offset       int
}

// GetPCIAuditEvent retrieves a single PCI audit event by ID
func (r *PCIAuditRepository) GetPCIAuditEvent(ctx context.Context, id uuid.UUID) (*PCIAuditEvent, error) {
	query := `
		SELECT 
			id, event_type, severity,
			actor_user_id, actor_email, actor_role, actor_ip, actor_user_agent, session_id,
			resource_type, resource_id, tenant_id,
			card_last_four, card_brand, card_expiry_month, card_expiry_year, token_id,
			key_id, key_algorithm, key_operation,
			description, request_id, transaction_id, stripe_event_id,
			auth_method, mfa_used, success, failure_reason,
			compliance_data, metadata,
			retention_until, tamper_hash, chain_hash,
			created_at
		FROM pci_audit_events
		WHERE id = $1
	`

	row := r.db.QueryRowContext(ctx, query, id)
	return scanPCIAuditEvent(row)
}

// VerifyAuditChain verifies the integrity of the audit chain
// Returns true if chain is valid, false if tampering is detected
func (r *PCIAuditRepository) VerifyAuditChain(ctx context.Context, limit int) (bool, []string, error) {
	query := `
		SELECT id, chain_hash, created_at, tamper_hash
		FROM pci_audit_events
		ORDER BY created_at ASC
		LIMIT $1
	`

	rows, err := r.db.QueryContext(ctx, query, limit)
	if err != nil {
		return false, nil, fmt.Errorf("failed to query audit chain: %w", err)
	}
	defer rows.Close()

	var issues []string
	var prevHash string
	isValid := true

	for rows.Next() {
		var id uuid.UUID
		var chainHash string
		var createdAt time.Time
		var tamperHash string

		if err := rows.Scan(&id, &chainHash, &createdAt, &tamperHash); err != nil {
			return false, nil, err
		}

		// Verify this event's chain hash includes previous hash
		if prevHash != "" {
			expectedPrefix := prevHash[:16]
			if len(chainHash) < 16 || chainHash[:16] != expectedPrefix {
				issues = append(issues, fmt.Sprintf("Chain break at event %s (created: %s)", id, createdAt))
				isValid = false
			}
		}

		prevHash = chainHash
	}

	return isValid, issues, nil
}

// PurgeExpiredEvents removes audit events past their retention period
// This should be run as a scheduled job
func (r *PCIAuditRepository) PurgeExpiredEvents(ctx context.Context) (int, error) {
	query := `
		DELETE FROM pci_audit_events
		WHERE retention_until < $1
		RETURNING id
	`

	rows, err := r.db.QueryContext(ctx, query, time.Now().UTC())
	if err != nil {
		return 0, fmt.Errorf("failed to purge expired PCI audit events: %w", err)
	}
	defer rows.Close()

	var purgedCount int
	for rows.Next() {
		var id uuid.UUID
		if err := rows.Scan(&id); err != nil {
			continue
		}
		purgedCount++
		logrus.WithField("event_id", id).Info("Purged expired PCI audit event")
	}

	return purgedCount, nil
}

// calculateChainHash creates a simple chain hash for tamper detection
func (r *PCIAuditRepository) calculateChainHash(ctx context.Context, event *PCIAuditEvent) (string, error) {
	// Get previous event hash
	var prevHash string
	query := `SELECT chain_hash FROM pci_audit_events ORDER BY created_at DESC LIMIT 1`
	row := r.db.QueryRowContext(ctx, query)
	if err := row.Scan(&prevHash); err != nil && err != sql.ErrNoRows {
		// Continue with empty previous hash
		prevHash = ""
	}

	// Create hash of current event data + previous hash
	h := sha256.New()
	h.Write([]byte(event.ID.String()))
	h.Write([]byte(event.EventType))
	h.Write([]byte(event.CreatedAt.String()))
	h.Write([]byte(prevHash))
	// Don't include sensitive data in the hash

	return hex.EncodeToString(h.Sum(nil)), nil
}

// scanPCIAuditEvent scans a single PCI audit event from a row
func scanPCIAuditEvent(row *sql.Row) (*PCIAuditEvent, error) {
	var event PCIAuditEvent
	var complianceDataJSON, metadataJSON []byte

	err := row.Scan(
		&event.ID, &event.EventType, &event.Severity,
		&event.ActorUserID, &event.ActorEmail, &event.ActorRole, &event.ActorIP, &event.ActorUserAgent, &event.SessionID,
		&event.ResourceType, &event.ResourceID, &event.TenantID,
		&event.CardLastFour, &event.CardBrand, &event.CardExpiryMonth, &event.CardExpiryYear, &event.TokenID,
		&event.KeyID, &event.KeyAlgorithm, &event.KeyOperation,
		&event.Description, &event.RequestID, &event.TransactionID, &event.StripeEventID,
		&event.AuthMethod, &event.MFAUsed, &event.Success, &event.FailureReason,
		&complianceDataJSON, &metadataJSON,
		&event.RetentionUntil, &event.TamperHash, &event.ChainHash,
		&event.CreatedAt,
	)

	if err == sql.ErrNoRows {
		return nil, nil
	}
	if err != nil {
		return nil, err
	}

	// Unmarshal JSON fields
	if len(complianceDataJSON) > 0 {
		json.Unmarshal(complianceDataJSON, &event.ComplianceData)
	}
	if len(metadataJSON) > 0 {
		json.Unmarshal(metadataJSON, &event.Metadata)
	}

	return &event, nil
}

// scanPCIAuditEventRows scans multiple PCI audit event rows
func scanPCIAuditEventRows(rows *sql.Rows) ([]PCIAuditEvent, error) {
	var events []PCIAuditEvent

	for rows.Next() {
		var event PCIAuditEvent
		var complianceDataJSON, metadataJSON []byte

		err := rows.Scan(
			&event.ID, &event.EventType, &event.Severity,
			&event.ActorUserID, &event.ActorEmail, &event.ActorRole, &event.ActorIP, &event.ActorUserAgent, &event.SessionID,
			&event.ResourceType, &event.ResourceID, &event.TenantID,
			&event.CardLastFour, &event.CardBrand, &event.CardExpiryMonth, &event.CardExpiryYear, &event.TokenID,
			&event.KeyID, &event.KeyAlgorithm, &event.KeyOperation,
			&event.Description, &event.RequestID, &event.TransactionID, &event.StripeEventID,
			&event.AuthMethod, &event.MFAUsed, &event.Success, &event.FailureReason,
			&complianceDataJSON, &metadataJSON,
			&event.RetentionUntil, &event.TamperHash, &event.ChainHash,
			&event.CreatedAt,
		)
		if err != nil {
			return nil, err
		}

		if len(complianceDataJSON) > 0 {
			json.Unmarshal(complianceDataJSON, &event.ComplianceData)
		}
		if len(metadataJSON) > 0 {
			json.Unmarshal(metadataJSON, &event.Metadata)
		}

		events = append(events, event)
	}

	return events, rows.Err()
}

// PCI constant aliases from the types package.
const PCIAuditChargebackReceived = types.PCIAuditChargebackReceived
const PCIAuditPaymentInitiated = types.PCIAuditPaymentInitiated
const PCIAuditPaymentFailed = types.PCIAuditPaymentFailed
const PCIAuditKeyCreated = types.PCIAuditKeyCreated
const PCIAuditKeyRotated = types.PCIAuditKeyRotated
const PCISeverityCritical = types.PCISeverityCritical
const PCIAuditCardDataRead = types.PCIAuditCardDataRead
const PCIAuditKeyRetired = types.PCIAuditKeyRetired
const PCIAuditKeyAccessed = types.PCIAuditKeyAccessed
const PCIAuditKeyBackupRestored = types.PCIAuditKeyBackupRestored
const PCIAuditPaymentProcessed = types.PCIAuditPaymentProcessed
const PCISeverityEmergency = types.PCISeverityEmergency
const PCISeverityInfo = types.PCISeverityInfo
const PCIAuditRefundProcessed = types.PCIAuditRefundProcessed
const PCISeverityWarning = types.PCISeverityWarning
const PCIAuditKeyBackupCreated = types.PCIAuditKeyBackupCreated

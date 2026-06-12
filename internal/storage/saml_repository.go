package storage

import (
	"context"
	"database/sql"
	"encoding/json"
	"encoding/xml"
	"fmt"
	"time"

	"github.com/google/uuid"
)

// SAMLConfig represents a SAML configuration for a tenant
type SAMLConfig struct {
	ID             uuid.UUID `json:"id" gorm:"type:uuid;primaryKey;default:gen_random_uuid()"`
	TenantID       uuid.UUID `json:"tenant_id" gorm:"type:uuid;not null;uniqueIndex"`
	Enabled        bool      `json:"enabled" gorm:"default:false"`
	IDPMetadata    *string   `json:"idp_metadata" gorm:"type:xml"`
	IDPEntityID    *string   `json:"idp_entity_id" gorm:"size:500"`
	IDPSSOURL      *string   `json:"idp_sso_url" gorm:"column:idp_sso_url;size:500"`
	IDPCertificate *string   `json:"idp_certificate" gorm:"type:text"`
	SPEntityID     string    `json:"sp_entity_id" gorm:"size:500;default:'functionfly'"`
	SPACSURL       *string   `json:"sp_acs_url" gorm:"column:sp_acs_url;size:500"`
	SPMetadataURL  *string   `json:"sp_metadata_url" gorm:"column:sp_metadata_url;size:500"`
	SPPrivateKey   *string   `json:"sp_private_key,omitempty" gorm:"type:text"` // PEM-encoded RSA private key
	SPCertificate  *string   `json:"sp_certificate,omitempty" gorm:"type:text"` // PEM-encoded X.509 certificate
	NameIDFormat   string    `json:"name_id_format" gorm:"size:100;default:'emailAddress'"`
	AuthnContexts  []string  `json:"authn_contexts" gorm:"type:text[]"`
	CreatedAt      time.Time `json:"created_at" gorm:"autoCreateTime"`
	UpdatedAt      time.Time `json:"updated_at" gorm:"autoUpdateTime"`
}

// SAMLSession represents an active SAML session
type SAMLSession struct {
	ID           uuid.UUID `json:"id" gorm:"type:uuid;primaryKey;default:gen_random_uuid()"`
	TenantID     uuid.UUID `json:"tenant_id" gorm:"type:uuid;not null"`
	UserID       uuid.UUID `json:"user_id" gorm:"type:uuid;not null"`
	SAMLNameID   string    `json:"saml_name_id" gorm:"size:255"`
	SessionIndex string    `json:"session_index" gorm:"size:255"`
	NotOnOrAfter time.Time `json:"not_on_or_after" gorm:"not null"`
	Attributes   JSONMap   `json:"attributes" gorm:"type:jsonb;default:'{}'"`
	CreatedAt    time.Time `json:"created_at" gorm:"autoCreateTime"`
}

// SAMLConfigRepository handles SAML configuration database operations
type SAMLConfigRepository struct {
	db *PostgresDB
}

// NewSAMLConfigRepository creates a new SAML config repository
func NewSAMLConfigRepository(db *PostgresDB) *SAMLConfigRepository {
	return &SAMLConfigRepository{db: db}
}

// Create creates a new SAML configuration
func (r *SAMLConfigRepository) Create(ctx context.Context, config *SAMLConfig) error {
	_, err := r.db.ExecContext(ctx, `
		INSERT INTO saml_configs (id, tenant_id, enabled, idp_metadata, idp_entity_id, idp_sso_url, idp_certificate, sp_entity_id, sp_acs_url, sp_metadata_url, name_id_format, authn_contexts, created_at, updated_at)
		VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10, $11, $12, $13, $14)`,
		config.ID, config.TenantID, config.Enabled, config.IDPMetadata, config.IDPEntityID, config.IDPSSOURL, config.IDPCertificate,
		config.SPEntityID, config.SPACSURL, config.SPMetadataURL, config.NameIDFormat, config.AuthnContexts, config.CreatedAt, config.UpdatedAt)
	if err != nil {
		if ctx.Err() == context.DeadlineExceeded {
			return fmt.Errorf("context deadline exceeded: %w", context.DeadlineExceeded)
		}
		return fmt.Errorf("failed to create SAML config: %w", err)
	}
	return nil
}

// GetByTenantID retrieves SAML config by tenant ID
func (r *SAMLConfigRepository) GetByTenantID(ctx context.Context, tenantID uuid.UUID) (*SAMLConfig, error) {
	var config SAMLConfig
	var idpMetadata sql.NullString
	var authnContexts []sql.NullString

	err := r.db.QueryRowContext(ctx, `
		SELECT id, tenant_id, enabled, idp_metadata, idp_entity_id, idp_sso_url, idp_certificate, sp_entity_id, sp_acs_url, sp_metadata_url, name_id_format, authn_contexts, created_at, updated_at
		FROM saml_configs WHERE tenant_id = $1`, tenantID).Scan(
		&config.ID, &config.TenantID, &config.Enabled, &idpMetadata, &config.IDPEntityID, &config.IDPSSOURL, &config.IDPCertificate,
		&config.SPEntityID, &config.SPACSURL, &config.SPMetadataURL, &config.NameIDFormat, &authnContexts, &config.CreatedAt, &config.UpdatedAt)

	if err != nil {
		if ctx.Err() == context.DeadlineExceeded {
			return nil, fmt.Errorf("context deadline exceeded: %w", context.DeadlineExceeded)
		}
		if err == sql.ErrNoRows {
			return nil, fmt.Errorf("SAML config not found for tenant")
		}
		return nil, fmt.Errorf("failed to get SAML config: %w", err)
	}

	if idpMetadata.Valid {
		config.IDPMetadata = &idpMetadata.String
	}

	// Convert []sql.NullString to []string
	if len(authnContexts) > 0 {
		config.AuthnContexts = make([]string, len(authnContexts))
		for i, c := range authnContexts {
			config.AuthnContexts[i] = c.String
		}
	}

	return &config, nil
}

// Update updates an existing SAML configuration
func (r *SAMLConfigRepository) Update(ctx context.Context, config *SAMLConfig) error {
	_, err := r.db.ExecContext(ctx, `
		UPDATE saml_configs SET
			enabled = $1, idp_metadata = $2, idp_entity_id = $3, idp_sso_url = $4, idp_certificate = $5,
			sp_entity_id = $6, sp_acs_url = $7, sp_metadata_url = $8, name_id_format = $9, authn_contexts = $10, updated_at = $11
		WHERE id = $12 AND tenant_id = $13`,
		config.Enabled, config.IDPMetadata, config.IDPEntityID, config.IDPSSOURL, config.IDPCertificate,
		config.SPEntityID, config.SPACSURL, config.SPMetadataURL, config.NameIDFormat, config.AuthnContexts, time.Now(),
		config.ID, config.TenantID)
	if err != nil {
		if ctx.Err() == context.DeadlineExceeded {
			return fmt.Errorf("context deadline exceeded: %w", context.DeadlineExceeded)
		}
		return fmt.Errorf("failed to update SAML config: %w", err)
	}
	return nil
}

// Delete deletes a SAML configuration
func (r *SAMLConfigRepository) Delete(ctx context.Context, tenantID uuid.UUID) error {
	_, err := r.db.ExecContext(ctx, `DELETE FROM saml_configs WHERE tenant_id = $1`, tenantID)
	if err != nil {
		if ctx.Err() == context.DeadlineExceeded {
			return fmt.Errorf("context deadline exceeded: %w", context.DeadlineExceeded)
		}
		return fmt.Errorf("failed to delete SAML config: %w", err)
	}
	return nil
}

// GetSPKeyPair retrieves the stored SP private key and certificate for a tenant
func (r *SAMLConfigRepository) GetSPKeyPair(ctx context.Context, tenantID uuid.UUID) (privateKey, certificate *string, err error) {
	var config SAMLConfig
	err = r.db.QueryRowContext(ctx, `
		SELECT sp_private_key, sp_certificate FROM saml_configs
		WHERE tenant_id = $1`, tenantID).Scan(&config.SPPrivateKey, &config.SPCertificate)
	if err != nil {
		if ctx.Err() == context.DeadlineExceeded {
			return nil, nil, fmt.Errorf("context deadline exceeded: %w", context.DeadlineExceeded)
		}
		if err == sql.ErrNoRows {
			return nil, nil, nil
		}
		return nil, nil, fmt.Errorf("failed to get SAML SP key pair: %w", err)
	}
	return config.SPPrivateKey, config.SPCertificate, nil
}

// SaveSPKeyPair stores the SP private key and certificate for a tenant
func (r *SAMLConfigRepository) SaveSPKeyPair(ctx context.Context, tenantID uuid.UUID, privateKey, certificate string) error {
	// Check if config exists
	var exists bool
	err := r.db.QueryRowContext(ctx, `SELECT EXISTS(SELECT 1 FROM saml_configs WHERE tenant_id = $1)`, tenantID).Scan(&exists)
	if err != nil {
		if ctx.Err() == context.DeadlineExceeded {
			return fmt.Errorf("context deadline exceeded: %w", context.DeadlineExceeded)
		}
		return fmt.Errorf("failed to check SAML config existence: %w", err)
	}

	if exists {
		// Update existing config
		_, err = r.db.ExecContext(ctx, `
			UPDATE saml_configs SET sp_private_key = $1, sp_certificate = $2, updated_at = $3
			WHERE tenant_id = $4`,
			privateKey, certificate, time.Now(), tenantID)
		if err != nil {
			if ctx.Err() == context.DeadlineExceeded {
				return fmt.Errorf("context deadline exceeded: %w", context.DeadlineExceeded)
			}
			return fmt.Errorf("failed to save SAML SP key pair: %w", err)
		}
	} else {
		// Create minimal config with just the keys
		_, err = r.db.ExecContext(ctx, `
			INSERT INTO saml_configs (tenant_id, sp_private_key, sp_certificate, sp_entity_id)
			VALUES ($1, $2, $3, 'functionfly')`,
			tenantID, privateKey, certificate)
		if err != nil {
			if ctx.Err() == context.DeadlineExceeded {
				return fmt.Errorf("context deadline exceeded: %w", context.DeadlineExceeded)
			}
			return fmt.Errorf("failed to create SAML config with SP key pair: %w", err)
		}
	}
	return nil
}

// SAMLStateRepository handles SAML state (state stores for AuthnRequest)
type SAMLStateRepository struct {
	db *PostgresDB
}

// NewSAMLStateRepository creates a new SAML state repository
func NewSAMLStateRepository(db *PostgresDB) *SAMLStateRepository {
	return &SAMLStateRepository{db: db}
}

// SaveAuthnRequestState saves an AuthnRequest state for validation
// requestID is the SAML AuthnRequest ID (InResponseTo target) for replay attack prevention
func (r *SAMLStateRepository) SaveAuthnRequestState(ctx context.Context, stateID string, tenantID uuid.UUID, requestID string, relayState string, expiresAt time.Time) error {
	metadata := fmt.Sprintf(`{"request_id": "%s", "relay_state": "%s"}`, requestID, relayState)
	_, err := r.db.ExecContext(ctx, `
		INSERT INTO auth_events (id, event_type, tenant_id, user_id, ip_address, user_agent, success, metadata, timestamp)
		VALUES ($1, 'saml_authn_request', $2, NULL, '', '', true, $3, $4)`,
		stateID, tenantID, metadata, expiresAt)
	if err != nil {
		if ctx.Err() == context.DeadlineExceeded {
			return fmt.Errorf("context deadline exceeded: %w", context.DeadlineExceeded)
		}
		return fmt.Errorf("failed to save SAML state: %w", err)
	}
	return nil
}

// GetAuthnRequestState retrieves an AuthnRequest state
// Returns tenantID, requestID (AuthnRequest ID), relayState, and error
func (r *SAMLStateRepository) GetAuthnRequestState(ctx context.Context, stateID string) (tenantID uuid.UUID, requestID string, relayState string, err error) {
	var metadata string
	var timestamp time.Time

	err = r.db.QueryRowContext(ctx, `
		SELECT tenant_id, metadata, timestamp FROM auth_events WHERE id = $1 AND event_type = 'saml_authn_request' AND success = true`,
		stateID).Scan(&tenantID, &metadata, &timestamp)

	if err != nil {
		if ctx.Err() == context.DeadlineExceeded {
			return uuid.Nil, "", "", fmt.Errorf("context deadline exceeded: %w", context.DeadlineExceeded)
		}
		if err == sql.ErrNoRows {
			return uuid.Nil, "", "", fmt.Errorf("SAML state not found or expired")
		}
		return uuid.Nil, "", "", err
	}

	if time.Now().After(timestamp) {
		return uuid.Nil, "", "", fmt.Errorf("SAML state expired")
	}

	var payload struct {
		RequestID  string `json:"request_id"`
		RelayState string `json:"relay_state"`
	}
	if err := json.Unmarshal([]byte(metadata), &payload); err != nil {
		return uuid.Nil, "", "", fmt.Errorf("failed to parse SAML state metadata: %w", err)
	}
	return tenantID, payload.RequestID, payload.RelayState, nil
}

// DeleteAuthnRequestState deletes an AuthnRequest state
func (r *SAMLStateRepository) DeleteAuthnRequestState(ctx context.Context, stateID string) error {
	_, err := r.db.ExecContext(ctx, `DELETE FROM auth_events WHERE id = $1 AND event_type = 'saml_authn_request'`, stateID)
	if err != nil {
		if ctx.Err() == context.DeadlineExceeded {
			return fmt.Errorf("context deadline exceeded: %w", context.DeadlineExceeded)
		}
		return err
	}
	return nil
}

// ParseIDPMetadata parses IdP metadata XML
func ParseIDPMetadata(metadataXML string) (*IDPMetadata, error) {
	var metadata IDPMetadata
	err := xml.Unmarshal([]byte(metadataXML), &metadata)
	if err != nil {
		return nil, fmt.Errorf("failed to parse IdP metadata: %w", err)
	}
	return &metadata, nil
}

// IDPMetadata represents parsed IdP metadata
type IDPMetadata struct {
	EntityID string `xml:"EntityID"`
	SSOURL   string `xml:"SSO>Location"`
	CertData string `xml:"SSO>KeyDescriptor>ds:KeyInfo>ds:X509Data>ds:X509Certificate"`
}

// SAMLSessionRepository handles SAML session database operations
type SAMLSessionRepository struct {
	db *PostgresDB
}

// NewSAMLSessionRepository creates a new SAML session repository
func NewSAMLSessionRepository(db *PostgresDB) *SAMLSessionRepository {
	return &SAMLSessionRepository{db: db}
}

// Create creates a new SAML session
func (r *SAMLSessionRepository) Create(ctx context.Context, session *SAMLSession) error {
	_, err := r.db.ExecContext(ctx, `
		INSERT INTO saml_sessions (id, tenant_id, user_id, saml_name_id, session_index, not_on_or_after, attributes, created_at)
		VALUES ($1, $2, $3, $4, $5, $6, $7, $8)`,
		session.ID, session.TenantID, session.UserID, session.SAMLNameID, session.SessionIndex, session.NotOnOrAfter, session.Attributes, session.CreatedAt)
	if err != nil {
		if ctx.Err() == context.DeadlineExceeded {
			return fmt.Errorf("context deadline exceeded: %w", context.DeadlineExceeded)
		}
		return fmt.Errorf("failed to create SAML session: %w", err)
	}
	return nil
}

// GetByUserID retrieves SAML sessions for a user
func (r *SAMLSessionRepository) GetByUserID(ctx context.Context, userID uuid.UUID) ([]*SAMLSession, error) {
	rows, err := r.db.QueryContext(ctx, `
		SELECT id, tenant_id, user_id, saml_name_id, session_index, not_on_or_after, attributes, created_at
		FROM saml_sessions WHERE user_id = $1 AND not_on_or_after > NOW()`, userID)
	if err != nil {
		if ctx.Err() == context.DeadlineExceeded {
			return nil, fmt.Errorf("context deadline exceeded: %w", context.DeadlineExceeded)
		}
		return nil, fmt.Errorf("failed to get SAML sessions: %w", err)
	}
	defer rows.Close()

	var sessions []*SAMLSession
	for rows.Next() {
		var session SAMLSession
		err := rows.Scan(&session.ID, &session.TenantID, &session.UserID, &session.SAMLNameID, &session.SessionIndex, &session.NotOnOrAfter, &session.Attributes, &session.CreatedAt)
		if err != nil {
			return nil, err
		}
		sessions = append(sessions, &session)
	}

	return sessions, nil
}

// GetByNameID retrieves SAML session by SAML NameID
func (r *SAMLSessionRepository) GetByNameID(ctx context.Context, tenantID uuid.UUID, nameID string) (*SAMLSession, error) {
	var session SAMLSession
	err := r.db.QueryRowContext(ctx, `
		SELECT id, tenant_id, user_id, saml_name_id, session_index, not_on_or_after, attributes, created_at
		FROM saml_sessions WHERE tenant_id = $1 AND saml_name_id = $2 AND not_on_or_after > NOW()
		ORDER BY created_at DESC LIMIT 1`, tenantID, nameID).Scan(
		&session.ID, &session.TenantID, &session.UserID, &session.SAMLNameID, &session.SessionIndex, &session.NotOnOrAfter, &session.Attributes, &session.CreatedAt)
	if err != nil {
		if ctx.Err() == context.DeadlineExceeded {
			return nil, fmt.Errorf("context deadline exceeded: %w", context.DeadlineExceeded)
		}
		if err == sql.ErrNoRows {
			return nil, nil
		}
		return nil, err
	}
	return &session, nil
}

// Delete deletes a SAML session
func (r *SAMLSessionRepository) Delete(ctx context.Context, sessionID uuid.UUID) error {
	_, err := r.db.ExecContext(ctx, `DELETE FROM saml_sessions WHERE id = $1`, sessionID)
	if err != nil {
		if ctx.Err() == context.DeadlineExceeded {
			return fmt.Errorf("context deadline exceeded: %w", context.DeadlineExceeded)
		}
		return err
	}
	return nil
}

// DeleteByUserID deletes all SAML sessions for a user
func (r *SAMLSessionRepository) DeleteByUserID(ctx context.Context, userID uuid.UUID) error {
	_, err := r.db.ExecContext(ctx, `DELETE FROM saml_sessions WHERE user_id = $1`, userID)
	if err != nil {
		if ctx.Err() == context.DeadlineExceeded {
			return fmt.Errorf("context deadline exceeded: %w", context.DeadlineExceeded)
		}
		return err
	}
	return nil
}

// DeleteExpired deletes all expired SAML sessions
func (r *SAMLSessionRepository) DeleteExpired(ctx context.Context) error {
	_, err := r.db.ExecContext(ctx, `DELETE FROM saml_sessions WHERE not_on_or_after < NOW()`)
	if err != nil {
		if ctx.Err() == context.DeadlineExceeded {
			return fmt.Errorf("context deadline exceeded: %w", context.DeadlineExceeded)
		}
		return err
	}
	return nil
}

package storage

import (
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
	NameIDFormat   string    `json:"name_id_format" gorm:"size:100;default:'emailAddress'"`
	AuthnContexts  []string  `json:"authn_contexts" gorm:"type:text[]"`
	CreatedAt      time.Time `json:"created_at" gorm:"autoCreateTime"`
	UpdatedAt      time.Time `json:"updated_at" gorm:"autoUpdateTime"`
}

// SAMLSession represents an active SAML session
type SAMLSession struct {
	ID           uuid.UUID              `json:"id" gorm:"type:uuid;primaryKey;default:gen_random_uuid()"`
	TenantID     uuid.UUID              `json:"tenant_id" gorm:"type:uuid;not null"`
	UserID       uuid.UUID              `json:"user_id" gorm:"type:uuid;not null"`
	SAMLNameID   string                 `json:"saml_name_id" gorm:"size:255"`
	SessionIndex string                 `json:"session_index" gorm:"size:255"`
	NotOnOrAfter time.Time              `json:"not_on_or_after" gorm:"not null"`
	Attributes   map[string]interface{} `json:"attributes" gorm:"type:jsonb;default:'{}'"`
	CreatedAt    time.Time              `json:"created_at" gorm:"autoCreateTime"`
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
func (r *SAMLConfigRepository) Create(config *SAMLConfig) error {
	_, err := r.db.Exec(`
		INSERT INTO saml_configs (id, tenant_id, enabled, idp_metadata, idp_entity_id, idp_sso_url, idp_certificate, sp_entity_id, sp_acs_url, sp_metadata_url, name_id_format, authn_contexts, created_at, updated_at)
		VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10, $11, $12, $13, $14)`,
		config.ID, config.TenantID, config.Enabled, config.IDPMetadata, config.IDPEntityID, config.IDPSSOURL, config.IDPCertificate,
		config.SPEntityID, config.SPACSURL, config.SPMetadataURL, config.NameIDFormat, config.AuthnContexts, config.CreatedAt, config.UpdatedAt)
	if err != nil {
		return fmt.Errorf("failed to create SAML config: %w", err)
	}
	return nil
}

// GetByTenantID retrieves SAML config by tenant ID
func (r *SAMLConfigRepository) GetByTenantID(tenantID uuid.UUID) (*SAMLConfig, error) {
	var config SAMLConfig
	var idpMetadata sql.NullString
	var authnContexts []sql.NullString

	err := r.db.QueryRow(`
		SELECT id, tenant_id, enabled, idp_metadata, idp_entity_id, idp_sso_url, idp_certificate, sp_entity_id, sp_acs_url, sp_metadata_url, name_id_format, authn_contexts, created_at, updated_at
		FROM saml_configs WHERE tenant_id = $1`, tenantID).Scan(
		&config.ID, &config.TenantID, &config.Enabled, &idpMetadata, &config.IDPEntityID, &config.IDPSSOURL, &config.IDPCertificate,
		&config.SPEntityID, &config.SPACSURL, &config.SPMetadataURL, &config.NameIDFormat, &authnContexts, &config.CreatedAt, &config.UpdatedAt)

	if err != nil {
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
func (r *SAMLConfigRepository) Update(config *SAMLConfig) error {
	_, err := r.db.Exec(`
		UPDATE saml_configs SET
			enabled = $1, idp_metadata = $2, idp_entity_id = $3, idp_sso_url = $4, idp_certificate = $5,
			sp_entity_id = $6, sp_acs_url = $7, sp_metadata_url = $8, name_id_format = $9, authn_contexts = $10, updated_at = $11
		WHERE id = $12 AND tenant_id = $13`,
		config.Enabled, config.IDPMetadata, config.IDPEntityID, config.IDPSSOURL, config.IDPCertificate,
		config.SPEntityID, config.SPACSURL, config.SPMetadataURL, config.NameIDFormat, config.AuthnContexts, time.Now(),
		config.ID, config.TenantID)
	if err != nil {
		return fmt.Errorf("failed to update SAML config: %w", err)
	}
	return nil
}

// Delete deletes a SAML configuration
func (r *SAMLConfigRepository) Delete(tenantID uuid.UUID) error {
	_, err := r.db.Exec(`DELETE FROM saml_configs WHERE tenant_id = $1`, tenantID)
	if err != nil {
		return fmt.Errorf("failed to delete SAML config: %w", err)
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
func (r *SAMLStateRepository) SaveAuthnRequestState(stateID string, tenantID uuid.UUID, relayState string, expiresAt time.Time) error {
	_, err := r.db.Exec(`
		INSERT INTO auth_events (id, event_type, tenant_id, user_id, ip_address, user_agent, success, metadata, timestamp)
		VALUES ($1, 'saml_authn_request', $2, NULL, '', '', true, $3, $4)`,
		stateID, tenantID, fmt.Sprintf(`{"relay_state": "%s"}`, relayState), expiresAt)
	if err != nil {
		return fmt.Errorf("failed to save SAML state: %w", err)
	}
	return nil
}

// GetAuthnRequestState retrieves an AuthnRequest state
func (r *SAMLStateRepository) GetAuthnRequestState(stateID string) (tenantID uuid.UUID, relayState string, err error) {
	var metadata string
	var timestamp time.Time

	err = r.db.QueryRow(`
		SELECT tenant_id, metadata, timestamp FROM auth_events WHERE id = $1 AND event_type = 'saml_authn_request' AND success = true`,
		stateID).Scan(&tenantID, &metadata, &timestamp)

	if err != nil {
		if err == sql.ErrNoRows {
			return uuid.Nil, "", fmt.Errorf("SAML state not found or expired")
		}
		return uuid.Nil, "", err
	}

	if time.Now().After(timestamp) {
		return uuid.Nil, "", fmt.Errorf("SAML state expired")
	}

	var payload struct {
		RelayState string `json:"relay_state"`
	}
	if err := json.Unmarshal([]byte(metadata), &payload); err != nil {
		return uuid.Nil, "", fmt.Errorf("failed to parse SAML state metadata: %w", err)
	}
	return tenantID, payload.RelayState, nil
}

// DeleteAuthnRequestState deletes an AuthnRequest state
func (r *SAMLStateRepository) DeleteAuthnRequestState(stateID string) error {
	_, err := r.db.Exec(`DELETE FROM auth_events WHERE id = $1 AND event_type = 'saml_authn_request'`, stateID)
	return err
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
func (r *SAMLSessionRepository) Create(session *SAMLSession) error {
	_, err := r.db.Exec(`
		INSERT INTO saml_sessions (id, tenant_id, user_id, saml_name_id, session_index, not_on_or_after, attributes, created_at)
		VALUES ($1, $2, $3, $4, $5, $6, $7, $8)`,
		session.ID, session.TenantID, session.UserID, session.SAMLNameID, session.SessionIndex, session.NotOnOrAfter, session.Attributes, session.CreatedAt)
	if err != nil {
		return fmt.Errorf("failed to create SAML session: %w", err)
	}
	return nil
}

// GetByUserID retrieves SAML sessions for a user
func (r *SAMLSessionRepository) GetByUserID(userID uuid.UUID) ([]*SAMLSession, error) {
	rows, err := r.db.Query(`
		SELECT id, tenant_id, user_id, saml_name_id, session_index, not_on_or_after, attributes, created_at
		FROM saml_sessions WHERE user_id = $1 AND not_on_or_after > NOW()`, userID)
	if err != nil {
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
func (r *SAMLSessionRepository) GetByNameID(tenantID uuid.UUID, nameID string) (*SAMLSession, error) {
	var session SAMLSession
	err := r.db.QueryRow(`
		SELECT id, tenant_id, user_id, saml_name_id, session_index, not_on_or_after, attributes, created_at
		FROM saml_sessions WHERE tenant_id = $1 AND saml_name_id = $2 AND not_on_or_after > NOW()
		ORDER BY created_at DESC LIMIT 1`, tenantID, nameID).Scan(
		&session.ID, &session.TenantID, &session.UserID, &session.SAMLNameID, &session.SessionIndex, &session.NotOnOrAfter, &session.Attributes, &session.CreatedAt)
	if err != nil {
		if err == sql.ErrNoRows {
			return nil, nil
		}
		return nil, err
	}
	return &session, nil
}

// Delete deletes a SAML session
func (r *SAMLSessionRepository) Delete(sessionID uuid.UUID) error {
	_, err := r.db.Exec(`DELETE FROM saml_sessions WHERE id = $1`, sessionID)
	return err
}

// DeleteByUserID deletes all SAML sessions for a user
func (r *SAMLSessionRepository) DeleteByUserID(userID uuid.UUID) error {
	_, err := r.db.Exec(`DELETE FROM saml_sessions WHERE user_id = $1`, userID)
	return err
}

// DeleteExpired deletes all expired SAML sessions
func (r *SAMLSessionRepository) DeleteExpired() error {
	_, err := r.db.Exec(`DELETE FROM saml_sessions WHERE not_on_or_after < NOW()`)
	return err
}

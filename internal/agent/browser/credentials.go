package browser

import (
	"context"
	"encoding/json"
	"fmt"
	"time"

	"github.com/google/uuid"
	"gorm.io/gorm"
)

// CredentialManager manages encrypted browser credentials.
type CredentialManager struct {
	db         *gorm.DB
	vaultMgr   VaultManager
	enabled    bool
}

// VaultManager interface for vault operations.
type VaultManager interface {
	Encrypt(ctx context.Context, plaintext []byte, agentID string) (ciphertext []byte, err error)
	Decrypt(ctx context.Context, ciphertext []byte, agentID string) (plaintext []byte, err error)
}

// DefaultVaultManager is a no-op vault manager for when vault is disabled.
type DefaultVaultManager struct{}

// Encrypt no-op encryption for when vault is disabled.
func (d *DefaultVaultManager) Encrypt(ctx context.Context, plaintext []byte, agentID string) ([]byte, error) {
	return plaintext, nil
}

// Decrypt no-op decryption for when vault is disabled.
func (d *DefaultVaultManager) Decrypt(ctx context.Context, ciphertext []byte, agentID string) ([]byte, error) {
	return ciphertext, nil
}

// NewCredentialManager creates a new credential manager.
func NewCredentialManager(db *gorm.DB, vaultMgr VaultManager, enabled bool) *CredentialManager {
	if vaultMgr == nil {
		vaultMgr = &DefaultVaultManager{}
	}
	return &CredentialManager{
		db:       db,
		vaultMgr: vaultMgr,
		enabled:  enabled,
	}
}

// BrowserCredential represents a stored browser credential.
type BrowserCredential struct {
	ID            uuid.UUID `json:"id" gorm:"type:uuid;primaryKey;default:gen_random_uuid()"`
	AgentID       string    `json:"agent_id" gorm:"not null;index"`
	Name          string    `json:"name" gorm:"not null"` // e.g., "google", "github"
	Domain        string    `json:"domain" gorm:"not null"`
	EncryptedData []byte    `json:"-" gorm:"column:encrypted_data"` // AES-256-GCM ciphertext
	CreatedAt     time.Time `json:"created_at" gorm:"autoCreateTime"`
	UpdatedAt     time.Time `json:"updated_at" gorm:"autoUpdateTime"`
}

// TableName returns the table name.
func (BrowserCredential) TableName() string {
	return "agent_browser_credentials"
}

// Store stores a browser credential for an agent.
func (cm *CredentialManager) Store(ctx context.Context, agentID, name, domain string, data *CredentialData) (*BrowserCredential, error) {
	if !cm.enabled {
		return nil, fmt.Errorf("credential storage is disabled")
	}

	// Serialize the credential data
	plaintext, err := json.Marshal(data)
	if err != nil {
		return nil, fmt.Errorf("failed to marshal credential data: %w", err)
	}

	// Encrypt the data
	ciphertext, err := cm.vaultMgr.Encrypt(ctx, plaintext, agentID)
	if err != nil {
		return nil, fmt.Errorf("failed to encrypt credential: %w", err)
	}

	credential := &BrowserCredential{
		ID:            uuid.New(),
		AgentID:       agentID,
		Name:          name,
		Domain:        domain,
		EncryptedData: ciphertext,
	}

	if err := cm.db.WithContext(ctx).Create(credential).Error; err != nil {
		return nil, fmt.Errorf("failed to store credential: %w", err)
	}

	return credential, nil
}

// Get retrieves a browser credential for an agent.
func (cm *CredentialManager) Get(ctx context.Context, agentID, credentialID string) (*BrowserCredential, error) {
	var credential BrowserCredential
	err := cm.db.WithContext(ctx).Where("id = ? AND agent_id = ?", credentialID, agentID).First(&credential).Error
	if err != nil {
		return nil, fmt.Errorf("credential not found: %w", err)
	}

	return &credential, nil
}

// GetByDomain retrieves all credentials for an agent and domain.
func (cm *CredentialManager) GetByDomain(ctx context.Context, agentID, domain string) ([]*BrowserCredential, error) {
	var credentials []*BrowserCredential
	err := cm.db.WithContext(ctx).Where("agent_id = ? AND domain = ?", agentID, domain).Find(&credentials).Error
	if err != nil {
		return nil, fmt.Errorf("failed to get credentials: %w", err)
	}

	return credentials, nil
}

// List lists all credentials for an agent.
func (cm *CredentialManager) List(ctx context.Context, agentID string) ([]*BrowserCredential, error) {
	var credentials []*BrowserCredential
	err := cm.db.WithContext(ctx).Where("agent_id = ?", agentID).Find(&credentials).Error
	if err != nil {
		return nil, fmt.Errorf("failed to list credentials: %w", err)
	}

	return credentials, nil
}

// Delete deletes a browser credential.
func (cm *CredentialManager) Delete(ctx context.Context, agentID, credentialID string) error {
	result := cm.db.WithContext(ctx).Where("id = ? AND agent_id = ?", credentialID, agentID).Delete(&BrowserCredential{})
	if result.Error != nil {
		return fmt.Errorf("failed to delete credential: %w", result.Error)
	}
	if result.RowsAffected == 0 {
		return fmt.Errorf("credential not found")
	}
	return nil
}

// DecryptCredential decrypts a credential and returns the plaintext data.
func (cm *CredentialManager) DecryptCredential(ctx context.Context, credential *BrowserCredential) (*CredentialData, error) {
	plaintext, err := cm.vaultMgr.Decrypt(ctx, credential.EncryptedData, credential.AgentID)
	if err != nil {
		return nil, fmt.Errorf("failed to decrypt credential: %w", err)
	}

	var data CredentialData
	if err := json.Unmarshal(plaintext, &data); err != nil {
		return nil, fmt.Errorf("failed to unmarshal credential data: %w", err)
	}

	return &data, nil
}

// MigrateCredentials runs the database migration for credentials.
func (cm *CredentialManager) Migrate(ctx context.Context) error {
	return cm.db.WithContext(ctx).AutoMigrate(&BrowserCredential{})
}

package identity

import (
	"context"
	"crypto/rand"
	"crypto/sha256"
	"encoding/hex"
	"fmt"
	"time"

	"github.com/functionfly/functionfly/internal/plans"
	"github.com/google/uuid"
	"gorm.io/gorm"
)

// Repository handles persistence for agent identities and quota configs
type Repository struct {
	db *gorm.DB
}

// NewRepository creates a new identity repository
func NewRepository(db *gorm.DB) *Repository {
	return &Repository{db: db}
}

// CreateAgent registers a new agent identity and returns the plaintext API key (only once)
// Returns (agent, apiKey, signingKey, error)
func (r *Repository) CreateAgent(ctx context.Context, tenantID uuid.UUID, req *RegisterAgentRequest) (*AgentIdentity, string, string, error) {
	// Validate plan tier
	planTier := req.PlanTier
	if planTier == "" {
		planTier = plans.PlanAgentStarter
	}
	if !plans.IsAgentTier(planTier) {
		return nil, "", "", fmt.Errorf("invalid plan tier: %s", planTier)
	}

	// Generate API key
	rawKey, keyHash, err := generateAPIKey()
	if err != nil {
		return nil, "", "", fmt.Errorf("failed to generate API key: %w", err)
	}

	// Generate signing key for A2A message authentication
	signingKey, signingHash, err := generateSigningKey()
	if err != nil {
		return nil, "", "", fmt.Errorf("failed to generate signing key: %w", err)
	}

	agent := &AgentIdentity{
		ID:           uuid.New(),
		TenantID:     tenantID,
		AgentID:      req.AgentID,
		Name:         req.Name,
		Description:  req.Description,
		PlanTier:     planTier,
		Status:       AgentStatusActive,
		APIKeyHash:   keyHash,
		SigningKeyHash: signingHash,
		CreatedAt:    time.Now(),
		UpdatedAt:    time.Now(),
	}

	err = r.db.WithContext(ctx).Transaction(func(tx *gorm.DB) error {
		if err := tx.Create(agent).Error; err != nil {
			return fmt.Errorf("failed to create agent identity: %w", err)
		}

		// Create default quota config based on plan tier
		quota := defaultQuotaForPlan(req.AgentID, planTier)
		if err := tx.Create(quota).Error; err != nil {
			return fmt.Errorf("failed to create default quota config: %w", err)
		}

		return nil
	})
	if err != nil {
		return nil, "", "", err
	}

	return agent, rawKey, signingKey, nil
}

// GetAgent retrieves an agent by its agent_id string
func (r *Repository) GetAgent(ctx context.Context, agentID string) (*AgentIdentity, error) {
	var agent AgentIdentity
	err := r.db.WithContext(ctx).Where("agent_id = ? AND status != ?", agentID, AgentStatusDeleted).First(&agent).Error
	if err != nil {
		if err == gorm.ErrRecordNotFound {
			return nil, fmt.Errorf("agent not found: %s", agentID)
		}
		return nil, fmt.Errorf("failed to get agent: %w", err)
	}
	return &agent, nil
}

// GetAgentByAPIKeyHash retrieves an agent by hashed API key (for auth)
func (r *Repository) GetAgentByAPIKeyHash(ctx context.Context, apiKey string) (*AgentIdentity, error) {
	keyHash := hashAPIKey(apiKey)
	var agent AgentIdentity
	err := r.db.WithContext(ctx).Where("api_key_hash = ? AND status = ?", keyHash, AgentStatusActive).First(&agent).Error
	if err != nil {
		if err == gorm.ErrRecordNotFound {
			return nil, fmt.Errorf("invalid API key")
		}
		return nil, fmt.Errorf("failed to authenticate agent: %w", err)
	}
	return &agent, nil
}

// GetAgentBySigningKeyHash retrieves an agent by hashed signing key (for A2A message verification)
func (r *Repository) GetAgentBySigningKeyHash(ctx context.Context, signingKey string) (*AgentIdentity, error) {
	keyHash := hashAPIKey(signingKey)
	var agent AgentIdentity
	err := r.db.WithContext(ctx).Where("signing_key_hash = ? AND status = ?", keyHash, AgentStatusActive).First(&agent).Error
	if err != nil {
		if err == gorm.ErrRecordNotFound {
			return nil, fmt.Errorf("invalid signing key")
		}
		return nil, fmt.Errorf("failed to authenticate agent by signing key: %w", err)
	}
	return &agent, nil
}

// ListAgents lists all agents for a tenant
func (r *Repository) ListAgents(ctx context.Context, tenantID uuid.UUID, limit, offset int) ([]*AgentIdentity, int64, error) {
	var total int64
	var agents []*AgentIdentity

	query := r.db.WithContext(ctx).Model(&AgentIdentity{}).
		Where("tenant_id = ? AND status != ?", tenantID, AgentStatusDeleted)

	if err := query.Count(&total).Error; err != nil {
		return nil, 0, fmt.Errorf("failed to count agents: %w", err)
	}

	if err := query.Order("created_at DESC").Limit(limit).Offset(offset).Find(&agents).Error; err != nil {
		return nil, 0, fmt.Errorf("failed to list agents: %w", err)
	}

	return agents, total, nil
}

// UpdateAgentStatus updates the status of an agent
func (r *Repository) UpdateAgentStatus(ctx context.Context, agentID string, status string) error {
	result := r.db.WithContext(ctx).Model(&AgentIdentity{}).
		Where("agent_id = ?", agentID).
		Updates(map[string]interface{}{
			"status":     status,
			"updated_at": time.Now(),
		})
	if result.Error != nil {
		return fmt.Errorf("failed to update agent status: %w", result.Error)
	}
	if result.RowsAffected == 0 {
		return fmt.Errorf("agent not found: %s", agentID)
	}
	return nil
}

// GetQuotaConfig retrieves the quota config for an agent
func (r *Repository) GetQuotaConfig(ctx context.Context, agentID string) (*AgentQuotaConfig, error) {
	var quota AgentQuotaConfig
	err := r.db.WithContext(ctx).Where("agent_id = ?", agentID).First(&quota).Error
	if err != nil {
		if err == gorm.ErrRecordNotFound {
			return nil, fmt.Errorf("quota config not found for agent: %s", agentID)
		}
		return nil, fmt.Errorf("failed to get quota config: %w", err)
	}
	return &quota, nil
}

// UpdateQuotaConfig updates the quota config for an agent
func (r *Repository) UpdateQuotaConfig(ctx context.Context, agentID string, updates map[string]interface{}) error {
	updates["updated_at"] = time.Now()
	result := r.db.WithContext(ctx).Model(&AgentQuotaConfig{}).
		Where("agent_id = ?", agentID).
		Updates(updates)
	if result.Error != nil {
		return fmt.Errorf("failed to update quota config: %w", result.Error)
	}
	if result.RowsAffected == 0 {
		return fmt.Errorf("quota config not found for agent: %s", agentID)
	}
	return nil
}

// RotateAPIKey generates a new API key for an agent and returns the plaintext key
func (r *Repository) RotateAPIKey(ctx context.Context, agentID string) (string, error) {
	rawKey, keyHash, err := generateAPIKey()
	if err != nil {
		return "", fmt.Errorf("failed to generate API key: %w", err)
	}

	result := r.db.WithContext(ctx).Model(&AgentIdentity{}).
		Where("agent_id = ?", agentID).
		Updates(map[string]interface{}{
			"api_key_hash": keyHash,
			"updated_at":   time.Now(),
		})
	if result.Error != nil {
		return "", fmt.Errorf("failed to rotate API key: %w", result.Error)
	}
	if result.RowsAffected == 0 {
		return "", fmt.Errorf("agent not found: %s", agentID)
	}

	return rawKey, nil
}

// generateAPIKey creates a random API key and returns (plaintext, hash)
func generateAPIKey() (string, string, error) {
	b := make([]byte, 32)
	if _, err := rand.Read(b); err != nil {
		return "", "", err
	}
	rawKey := "aep_" + hex.EncodeToString(b)
	return rawKey, hashAPIKey(rawKey), nil
}

// generateSigningKey creates a random signing key for agent-to-agent message authentication
// Returns (plaintext, hash)
func generateSigningKey() (string, string, error) {
	b := make([]byte, 32)
	if _, err := rand.Read(b); err != nil {
		return "", "", err
	}
	rawKey := "ags_" + hex.EncodeToString(b)
	return rawKey, hashAPIKey(rawKey), nil
}

// hashAPIKey returns the SHA-256 hash of an API key
func hashAPIKey(key string) string {
	h := sha256.Sum256([]byte(key))
	return hex.EncodeToString(h[:])
}

// defaultQuotaForPlan creates a default quota config based on the plan tier
// Supports both legacy agent tiers (agent_starter, etc.) and main plan names (starter, professional, etc.)
func defaultQuotaForPlan(agentID, planTier string) *AgentQuotaConfig {
	quota := &AgentQuotaConfig{
		ID:      uuid.New(),
		AgentID: agentID,
	}

	// Use unified limit getter for all plan types
	maxCallsPerMinute, maxCallsPerDay, maxStateWritesPerHr, maxDailySpendUSD := plans.GetAgentTierLimits(planTier)

	quota.MaxCallsPerMinute = maxCallsPerMinute
	quota.MaxCallsPerDay = maxCallsPerDay
	quota.MaxStateWritesPerHr = maxStateWritesPerHr
	quota.MaxDailySpendUSD = maxDailySpendUSD
	quota.MaxCostPerExecution = 0.01

	return quota
}

// CreateAgentHiring creates a new agent hiring record
func (r *Repository) CreateAgentHiring(ctx context.Context, hiring *AgentHiring) error {
	if hiring.ID == uuid.Nil {
		hiring.ID = uuid.New()
	}
	hiring.CreatedAt = time.Now()
	hiring.UpdatedAt = time.Now()
	return r.db.WithContext(ctx).Create(hiring).Error
}

// CountAgentHiring returns the total number of hiring records for an agent
func (r *Repository) CountAgentHiring(ctx context.Context, agentID string) (int, error) {
	var count int64
	err := r.db.WithContext(ctx).Model(&AgentHiring{}).Where("agent_id = ?", agentID).Count(&count).Error
	return int(count), err
}

// CountHiringByHirer returns the total number of hiring records made by a specific hirer
func (r *Repository) CountHiringByHirer(ctx context.Context, hirerID string) (int, error) {
	var count int64
	err := r.db.WithContext(ctx).Model(&AgentHiring{}).Where("hirer_id = ?", hirerID).Count(&count).Error
	return int(count), err
}

// GetAgentHiringHistory returns hiring records for an agent with pagination
func (r *Repository) GetAgentHiringHistory(ctx context.Context, agentID string, limit, offset int) ([]AgentHiring, error) {
	var hirings []AgentHiring
	err := r.db.WithContext(ctx).
		Where("agent_id = ?", agentID).
		Order("created_at DESC").
		Limit(limit).
		Offset(offset).
		Find(&hirings).Error
	return hirings, err
}

// CreateFunctionPurchase creates a new function purchase record
func (r *Repository) CreateFunctionPurchase(ctx context.Context, purchase *FunctionPurchase) error {
	if purchase.ID == uuid.Nil {
		purchase.ID = uuid.New()
	}
	purchase.CreatedAt = time.Now()
	purchase.UpdatedAt = time.Now()
	return r.db.WithContext(ctx).Create(purchase).Error
}

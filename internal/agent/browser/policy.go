package browser

import (
	"context"
	"fmt"
	"strings"

	"gorm.io/gorm"
)

// PolicyChecker provides policy helpers for browser permissions.
type PolicyChecker struct {
	db *gorm.DB
}

// NewPolicyChecker creates a new policy checker.
func NewPolicyChecker(db *gorm.DB) *PolicyChecker {
	return &PolicyChecker{db: db}
}

// CheckDomain checks if a domain is allowed for an agent.
func (pc *PolicyChecker) CheckDomain(ctx context.Context, agentID, domain string) (bool, error) {
	perm, err := pc.GetPermission(ctx, agentID)
	if err != nil {
		// Default to allowed if error
		return true, nil
	}

	return pc.CheckDomainWithPermission(domain, perm), nil
}

// CheckDomainWithPermission checks if a domain is allowed given the permission.
func (pc *PolicyChecker) CheckDomainWithPermission(domain string, perm *BrowserPermission) bool {
	if !perm.BrowserEnabled {
		return false
	}

	if len(perm.AllowedDomains) == 0 {
		return true
	}

	for _, pattern := range perm.AllowedDomains {
		if matchDomainPattern(pattern, domain) {
			return true
		}
	}

	return false
}

// GetPermission retrieves browser permissions for an agent.
func (pc *PolicyChecker) GetPermission(ctx context.Context, agentID string) (*BrowserPermission, error) {
	var config AgentBrowserConfig
	err := pc.db.WithContext(ctx).Where("agent_id = ?", agentID).First(&config).Error
	if err != nil {
		if err == gorm.ErrRecordNotFound {
			return DefaultBrowserPermission(agentID), nil
		}
		return nil, fmt.Errorf("failed to get browser config: %w", err)
	}

	return &BrowserPermission{
		AgentID:            agentID,
		BrowserEnabled:    config.BrowserEnabled,
		AllowedDomains:     config.AllowedDomains,
		MaxSessions:       config.MaxSessions,
		CredentialStorage: config.CredentialStorageEnabled,
		DefaultTimeoutMs:  config.DefaultTimeoutMs,
		HeadfulMode:       config.HeadfulMode,
	}, nil
}

// UpsertPermission creates or updates browser permissions for an agent.
func (pc *PolicyChecker) UpsertPermission(ctx context.Context, perm *BrowserPermission) error {
	config := AgentBrowserConfig{
		AgentID:                   perm.AgentID,
		BrowserEnabled:            perm.BrowserEnabled,
		AllowedDomains:           perm.AllowedDomains,
		MaxSessions:              perm.MaxSessions,
		CredentialStorageEnabled: perm.CredentialStorage,
		DefaultTimeoutMs:         perm.DefaultTimeoutMs,
		HeadfulMode:              perm.HeadfulMode,
	}

	return pc.db.WithContext(ctx).
		Where("agent_id = ?", perm.AgentID).
		Assign(config).
		FirstOrCreate(&config).Error
}

// AgentBrowserConfig is the database model for browser configuration.
type AgentBrowserConfig struct {
	AgentID                   string   `gorm:"primaryKey"`
	BrowserEnabled           bool     `gorm:"not null;default:true"`
	AllowedDomains           []string `gorm:"type:text[]"`
	MaxSessions              int      `gorm:"not null;default:1"`
	CredentialStorageEnabled bool     `gorm:"not null;default:false"`
	DefaultTimeoutMs          int      `gorm:"not null;default:30000"`
	HeadfulMode              bool     `gorm:"not null;default:false"`
}

// TableName returns the table name.
func (AgentBrowserConfig) TableName() string {
	return "agent_browser_configs"
}

// matchDomainPattern matches a domain glob pattern.
func matchDomainPattern(pattern, domain string) bool {
	// Exact match
	if pattern == domain {
		return true
	}

	// Wildcard subdomain: *.example.com matches www.example.com
	if strings.HasPrefix(pattern, "*.") {
		suffix := pattern[2:]
		if len(domain) > len(suffix) && strings.HasSuffix(domain, suffix) {
			return true
		}
	}

	// Suffix match: example.com matches *.example.com
	if strings.HasPrefix(pattern, "*.") {
		suffix := pattern[2:]
		if len(domain) >= len(suffix) && strings.HasSuffix(domain, suffix) {
			return true
		}
	}

	return false
}

// ValidateBrowserRequest validates a browser action request.
func (pc *PolicyChecker) ValidateBrowserRequest(ctx context.Context, agentID, domain string, sessionType SessionType) (*ValidationResult, error) {
	perm, err := pc.GetPermission(ctx, agentID)
	if err != nil {
		return &ValidationResult{Allowed: false, Error: err.Error()}, err
	}

	if !perm.BrowserEnabled {
		return &ValidationResult{
			Allowed: false,
			Error:   "browser automation is disabled for this agent",
		}, nil
	}

	if !pc.CheckDomainWithPermission(domain, perm) {
		return &ValidationResult{
			Allowed: false,
			Error:   fmt.Sprintf("domain %s is not allowed for this agent", domain),
		}, nil
	}

	return &ValidationResult{Allowed: true}, nil
}

// ValidationResult represents the result of a validation check.
type ValidationResult struct {
	Allowed bool   `json:"allowed"`
	Error  string `json:"error,omitempty"`
}

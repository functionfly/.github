package browser

import (
	"context"
	"fmt"
	"time"

	"github.com/google/uuid"
	"github.com/redis/go-redis/v9"
	"gorm.io/gorm"
)

// Browser is the main interface for browser automation.
type Browser interface {
	// Initialize sets up the browser service.
	Initialize(ctx context.Context) error

	// Navigate navigates to a URL.
	Navigate(ctx context.Context, agentID, url string, sessionID *uuid.UUID) (*NavigateResult, error)

	// Click clicks an element.
	Click(ctx context.Context, agentID, sessionID, elementRef string) error

	// Fill fills a form field.
	Fill(ctx context.Context, agentID, sessionID, elementRef, value string) error

	// Extract extracts structured content.
	Extract(ctx context.Context, agentID, sessionID, selector string) ([]map[string]interface{}, error)

	// Screenshot captures a screenshot.
	Screenshot(ctx context.Context, agentID, sessionID string) (string, error)

	// CreateSession creates a new browser session.
	CreateSession(ctx context.Context, agentID string, isolated bool) (*SessionState, error)

	// CloseSession closes a browser session.
	CloseSession(ctx context.Context, agentID, sessionID string) error

	// GetSession gets a session by ID.
	GetSession(ctx context.Context, sessionID uuid.UUID) (*SessionState, error)

	// ListSessions lists all sessions for an agent.
	ListSessions(ctx context.Context, agentID string) ([]*SessionState, error)

	// StoreCredential stores a browser credential.
	StoreCredential(ctx context.Context, agentID, name, domain string, data *CredentialData) (*BrowserCredential, error)

	// GetCredential retrieves a browser credential.
	GetCredential(ctx context.Context, agentID, credentialID string) (*BrowserCredential, error)

	// ListCredentials lists all credentials for an agent.
	ListCredentials(ctx context.Context, agentID string) ([]*BrowserCredential, error)

	// DeleteCredential deletes a browser credential.
	DeleteCredential(ctx context.Context, agentID, credentialID string) error

	// GetPermission gets browser permissions for an agent.
	GetPermission(ctx context.Context, agentID string) (*BrowserPermission, error)

	// UpsertPermission updates browser permissions for an agent.
	UpsertPermission(ctx context.Context, perm *BrowserPermission) error

	// PoolStats returns pool statistics.
	PoolStats() (available, allocated int)

	// HealthCheck returns the health status of browser instances.
	HealthCheck(ctx context.Context) map[string]string
}

// NewBrowser creates a new browser service instance.
func NewBrowser(config Config, db *gorm.DB, redisClient *redis.Client) Browser {
	return NewService(config, db, redisClient)
}

// VerifyBrowserService verifies that the browser service dependencies are available.
func VerifyBrowserService(db *gorm.DB, redisClient *redis.Client) error {
	if db == nil {
		return fmt.Errorf("database connection is required")
	}
	if redisClient == nil {
		return fmt.Errorf("redis connection is required")
	}
	return nil
}

// Session represents a browser session with CDP context.
type Session struct {
	ID           uuid.UUID              `json:"id"`
	AgentID      string                 `json:"agent_id"`
	SessionType  SessionType            `json:"session_type"` // shared | isolated
	Status       SessionStatus          `json:"status"`       // active | closing | closed | crashed
	BrowserPort  int                    `json:"browser_port"`
	URL          string                 `json:"url"`
	Cookies      []SessionCookie        `json:"cookies"`
	AuthToken    string                 `json:"auth_token"`
	CreatedAt    time.Time              `json:"created_at"`
	LastUsedAt   time.Time              `json:"last_used_at"`
	ClosedAt     *time.Time             `json:"closed_at,omitempty"`
	Metadata     map[string]interface{} `json:"metadata,omitempty"`
}

// BrowserSession is the database model for browser sessions.
type BrowserSession struct {
	ID          uuid.UUID  `json:"id" gorm:"type:uuid;primaryKey;default:gen_random_uuid()"`
	AgentID     string     `json:"agent_id" gorm:"not null;index"`
	SessionType string     `json:"session_type" gorm:"not null;default:'shared'"`
	Status      string     `json:"status" gorm:"not null;default:'active'"`
	BrowserPort int        `json:"browser_port"`
	CreatedAt   time.Time  `json:"created_at" gorm:"autoCreateTime"`
	ClosedAt    *time.Time `json:"closed_at,omitempty"`
}

// TableName returns the table name.
func (BrowserSession) TableName() string {
	return "agent_browser_sessions"
}

// DefaultBrowserPermission returns default browser permissions.
func DefaultBrowserPermission(agentID string) *BrowserPermission {
	return &BrowserPermission{
		AgentID:           agentID,
		BrowserEnabled:    true,
		AllowedDomains:    []string{"*"}, // Allow all by default
		MaxSessions:       1,
		CredentialStorage: false,
		DefaultTimeoutMs:  30000,
		HeadfulMode:       false,
	}
}

// SessionState is an alias for Session for backwards compatibility.
type SessionState = Session

// SessionType represents the type of browser session.
type SessionType string

const (
	SessionTypeShared    SessionType = "shared"
	SessionTypeIsolated SessionType = "isolated"
)

// SessionStatus represents the status of a browser session.
type SessionStatus string

const (
	SessionStatusActive   SessionStatus = "active"
	SessionStatusClosing  SessionStatus = "closing"
	SessionStatusClosed   SessionStatus = "closed"
	SessionStatusCrashed  SessionStatus = "crashed"
)

// SessionCookie represents a browser cookie.
type SessionCookie struct {
	Name     string `json:"name"`
	Value    string `json:"value"`
	Domain   string `json:"domain"`
	Path     string `json:"path"`
	Expires  int64  `json:"expires"`
	Secure   bool   `json:"secure"`
	HTTPOnly bool   `json:"http_only"`
}

// BrowserPermission represents browser permissions for an agent.
type BrowserPermission struct {
	AgentID           string   `json:"agent_id"`
	BrowserEnabled    bool     `json:"browser_enabled"`
	AllowedDomains    []string `json:"allowed_domains"`
	MaxSessions       int      `json:"max_sessions"`
	CredentialStorage bool     `json:"credential_storage"`
	DefaultTimeoutMs  int      `json:"default_timeout_ms"`
	HeadfulMode       bool     `json:"headful_mode"`
}

// NavigateResult represents the result of a navigation.
type NavigateResult struct {
	SessionID   string            `json:"session_id"`
	URL         string            `json:"url"`
	Title       string            `json:"title"`
	StatusCode  int               `json:"status_code"`
	DurationMs  int               `json:"duration_ms"`
	Elements    []PageElement     `json:"elements,omitempty"`
	Screenshot  string            `json:"screenshot,omitempty"` // base64
}

// PageElement represents an interactive element on a page.
type PageElement struct {
	Ref   string            `json:"ref"`   // @eN format
	Tag   string            `json:"tag"`
	Text  string            `json:"text"`
	Type  string            `json:"type"`  // button, input, link, etc.
	Attrs map[string]string `json:"attrs,omitempty"`
}

// CredentialData represents the plaintext credential data.
type CredentialData struct {
	Cookies    []SessionCookie   `json:"cookies,omitempty"`
	AuthHeader string            `json:"auth_header,omitempty"`
	Tokens     map[string]string `json:"tokens,omitempty"`
}

// InstanceStatus represents the status of a browser instance.
type InstanceStatus string

const (
	InstanceStatusStarting InstanceStatus = "starting"
	InstanceStatusRunning  InstanceStatus = "running"
	InstanceStatusStopping InstanceStatus = "stopping"
	InstanceStatusStopped  InstanceStatus = "stopped"
)

// BrowserHandle represents an allocated browser instance.
type BrowserHandle interface {
	Port() int
	PID() int
	IsHealthy() bool
}

// BrowserError represents a browser operation error.
type BrowserError struct {
	Type      ErrorType `json:"type"`
	Message   string    `json:"message"`
	SessionID string    `json:"session_id,omitempty"`
	Domain    string    `json:"domain,omitempty"`
	Timestamp time.Time `json:"timestamp"`
}

// ErrorType categorizes browser errors.
type ErrorType string

const (
	ErrorTypeCrash   ErrorType = "crash"
	ErrorTypeTimeout ErrorType = "timeout"
	ErrorTypeNetwork ErrorType = "network"
	ErrorTypeDomain  ErrorType = "domain_blocked"
	ErrorTypeUnknown ErrorType = "unknown"
)

// RecoveryAction determines what action to take after an error.
type RecoveryAction string

const (
	RecoveryActionRetry    RecoveryAction = "retry"
	RecoveryActionRestart  RecoveryAction = "restart"
	RecoveryActionFallback RecoveryAction = "fallback"
	RecoveryActionFail     RecoveryAction = "fail"
)

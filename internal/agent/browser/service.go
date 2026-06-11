package browser

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"os/exec"
	"strconv"
	"strings"
	"time"

	"github.com/google/uuid"
	"github.com/redis/go-redis/v9"
	"github.com/sirupsen/logrus"
	"gorm.io/gorm"
)

// Service is the main browser service that wraps agent-browser CLI.
type Service struct {
	config     Config
	db         *gorm.DB
	redis      *redis.Client
	pool       *Pool
	isolated   *IsolatedManager
	sessionMgr *SessionManager
	policy     *PolicyChecker
	creds      *CredentialManager
	recovery   *RecoveryManager
}

// NewService creates a new browser service.
func NewService(config Config, db *gorm.DB, redisClient *redis.Client) *Service {
	sessionMgr := NewSessionManager(redisClient, config.SessionTTL)
	pool := NewPool(config, redisClient, sessionMgr)
	isolated := NewIsolatedManager(config, sessionMgr)
	policy := NewPolicyChecker(db)
	creds := NewCredentialManager(db, nil, config.VaultEnabled)
	recovery := NewRecoveryManager(config, sessionMgr, pool, isolated)

	svc := &Service{
		config:     config,
		db:         db,
		redis:      redisClient,
		pool:       pool,
		isolated:   isolated,
		sessionMgr: sessionMgr,
		policy:     policy,
		creds:      creds,
		recovery:   recovery,
	}

	return svc
}

// Initialize sets up the browser service.
func (s *Service) Initialize(ctx context.Context) error {
	if !s.config.Enabled {
		logrus.Info("Browser service: disabled by configuration")
		return nil
	}

	// Run migrations
	if err := s.db.WithContext(ctx).AutoMigrate(&AgentBrowserConfig{}); err != nil {
		return fmt.Errorf("failed to migrate browser config: %w", err)
	}
	if err := s.db.WithContext(ctx).AutoMigrate(&BrowserSession{}); err != nil {
		return fmt.Errorf("failed to migrate browser session: %w", err)
	}
	if err := s.creds.Migrate(ctx); err != nil {
		return fmt.Errorf("failed to migrate credentials: %w", err)
	}

	// Start pool health check
	s.pool.StartPoolHealthCheck(ctx, 30*time.Second)

	// Start isolated instance cleanup
	s.isolated.StartCleanupRoutine(ctx, 5*time.Minute)

	logrus.Info("Browser service: initialized")
	return nil
}

// Navigate navigates to a URL in a browser session.
func (s *Service) Navigate(ctx context.Context, agentID, url string, sessionID *uuid.UUID) (*NavigateResult, error) {
	// Validate domain
	domain := extractDomain(url)
	valid, err := s.policy.CheckDomain(ctx, agentID, domain)
	if err != nil {
		return nil, fmt.Errorf("policy check failed: %w", err)
	}
	if !valid {
		return nil, fmt.Errorf("domain %s is not allowed", domain)
	}

	// Get or create session
	var session *SessionState
	if sessionID != nil {
		session, err = s.sessionMgr.GetSession(ctx, *sessionID)
		if err != nil {
			return nil, fmt.Errorf("session not found: %w", err)
		}
	} else {
		// Check for sticky session
		session, err = s.sessionMgr.GetSessionByAgent(ctx, agentID)
		if err != nil || session == nil {
			// Create new session
			port, err := s.pool.Acquire(ctx, agentID)
			if err != nil {
				return nil, fmt.Errorf("failed to acquire browser: %w", err)
			}
			session, err = s.sessionMgr.CreateSession(ctx, agentID, SessionTypeShared, port)
			if err != nil {
				return nil, fmt.Errorf("failed to create session: %w", err)
			}
		}
	}

	// Check loop detection
	loopDetected, err := s.sessionMgr.CheckLoopDetection(ctx, agentID, session.ID.String(), domain, 3)
	if err != nil {
		logrus.Warnf("Browser loop detection error: %v", err)
	}
	if loopDetected {
		return nil, fmt.Errorf("loop detected: domain %s accessed too many times", domain)
	}

	// Execute browser command
	result, err := s.executeBrowserCommand(ctx, "navigate", session.BrowserPort, session.AuthToken, url)
	if err != nil {
		// Handle error with recovery
		browserErr := &BrowserError{
			Type:      classifyError(err),
			Message:   err.Error(),
			SessionID: session.ID.String(),
			Domain:    domain,
			Timestamp: time.Now().UTC(),
		}
		action, recErr := s.recovery.HandleError(ctx, browserErr)
		if recErr != nil {
			return nil, recErr
		}
		if action == RecoveryActionFail {
			return nil, err
		}
		// Retry or restart
		result, err = s.executeBrowserCommand(ctx, "navigate", session.BrowserPort, session.AuthToken, url)
		if err != nil {
			return nil, err
		}
	}

	// Update session URL
	s.sessionMgr.SetSessionURL(ctx, session.ID, url)

	// Record usage
	s.recordUsage(ctx, agentID, session.ID, "navigate", domain, result.DurationMs)

	// Convert BrowserResult to NavigateResult
	title := ""
	statusCode := 200
	if result.Metadata != nil {
		if t, ok := result.Metadata["title"].(string); ok {
			title = t
		}
		if sc, ok := result.Metadata["status_code"].(float64); ok {
			statusCode = int(sc)
		}
	}

	navigateResult := &NavigateResult{
		SessionID:   session.ID.String(),
		URL:         url,
		Title:       title,
		StatusCode:  statusCode,
		DurationMs:  result.DurationMs,
		Screenshot:  result.Screenshot,
	}

	return navigateResult, nil
}

// Click clicks an element.
func (s *Service) Click(ctx context.Context, agentID, sessionID, elementRef string) error {
	session, err := s.sessionMgr.GetSession(ctx, uuid.MustParse(sessionID))
	if err != nil {
		return fmt.Errorf("session not found: %w", err)
	}

	_, err = s.executeBrowserCommand(ctx, "click", session.BrowserPort, session.AuthToken, elementRef)
	return err
}

// Fill fills a form field.
func (s *Service) Fill(ctx context.Context, agentID, sessionID, elementRef, value string) error {
	session, err := s.sessionMgr.GetSession(ctx, uuid.MustParse(sessionID))
	if err != nil {
		return fmt.Errorf("session not found: %w", err)
	}

	_, err = s.executeBrowserCommand(ctx, "fill", session.BrowserPort, session.AuthToken, elementRef, value)
	return err
}

// Extract extracts structured content from the page.
func (s *Service) Extract(ctx context.Context, agentID, sessionID, selector string) ([]map[string]interface{}, error) {
	session, err := s.sessionMgr.GetSession(ctx, uuid.MustParse(sessionID))
	if err != nil {
		return nil, fmt.Errorf("session not found: %w", err)
	}

	result, err := s.executeBrowserCommand(ctx, "extract", session.BrowserPort, session.AuthToken, selector)
	if err != nil {
		return nil, err
	}

	var content []map[string]interface{}
	if err := json.Unmarshal([]byte(result.Output), &content); err != nil {
		return nil, fmt.Errorf("failed to parse extract result: %w", err)
	}

	return content, nil
}

// Screenshot captures a screenshot.
func (s *Service) Screenshot(ctx context.Context, agentID, sessionID string) (string, error) {
	session, err := s.sessionMgr.GetSession(ctx, uuid.MustParse(sessionID))
	if err != nil {
		return "", fmt.Errorf("session not found: %w", err)
	}

	result, err := s.executeBrowserCommand(ctx, "screenshot", session.BrowserPort, session.AuthToken)
	if err != nil {
		return "", err
	}

	return result.Screenshot, nil
}

// CloseSession closes a browser session.
func (s *Service) CloseSession(ctx context.Context, agentID, sessionID string) error {
	session, err := s.sessionMgr.GetSession(ctx, uuid.MustParse(sessionID))
	if err != nil {
		return fmt.Errorf("session not found: %w", err)
	}

	// Execute close command
	s.executeBrowserCommand(ctx, "close", session.BrowserPort, session.AuthToken)

	// Mark session as closed
	s.sessionMgr.CloseSession(ctx, session.ID)

	// Release pool resource if shared
	if session.SessionType == SessionTypeShared {
		s.pool.Release(agentID)
	}

	return nil
}

// CreateSession creates a new browser session.
func (s *Service) CreateSession(ctx context.Context, agentID string, isolated bool) (*SessionState, error) {
	var session *SessionState

	if isolated {
		instance, err := s.isolated.Acquire(ctx, agentID)
		if err != nil {
			return nil, fmt.Errorf("failed to acquire isolated browser: %w", err)
		}
		session, err = s.sessionMgr.CreateSession(ctx, agentID, SessionTypeIsolated, instance.Port)
		if err != nil {
			return nil, fmt.Errorf("failed to create session: %w", err)
		}
	} else {
		port, err := s.pool.Acquire(ctx, agentID)
		if err != nil {
			return nil, fmt.Errorf("failed to acquire shared browser: %w", err)
		}
		session, err = s.sessionMgr.CreateSession(ctx, agentID, SessionTypeShared, port)
		if err != nil {
			return nil, fmt.Errorf("failed to create session: %w", err)
		}
	}

	return session, nil
}

// executeBrowserCommand executes a command via agent-browser CLI.
func (s *Service) executeBrowserCommand(ctx context.Context, command string, port int, authToken string, args ...string) (*BrowserResult, error) {
	// Build command arguments
	cmdArgs := []string{
		command,
		"--port", strconv.Itoa(port),
		"--auth-token", authToken,
		"--output", "json",
	}
	cmdArgs = append(cmdArgs, args...)

	// Execute agent-browser CLI
	cmd := exec.CommandContext(ctx, "agent-browser", cmdArgs...)
	var stdout, stderr bytes.Buffer
	cmd.Stdout = &stdout
	cmd.Stderr = &stderr

	err := cmd.Run()
	if err != nil {
		return nil, fmt.Errorf("agent-browser %s failed: %w (stderr: %s)", command, err, stderr.String())
	}

	// Parse result
	var result BrowserResult
	if err := json.Unmarshal(stdout.Bytes(), &result); err != nil {
		return nil, fmt.Errorf("failed to parse browser result: %w", err)
	}

	return &result, nil
}

// BrowserResult represents the result from agent-browser CLI.
type BrowserResult struct {
	Success    bool                   `json:"success"`
	Output     string                 `json:"output"`
	Screenshot string                 `json:"screenshot,omitempty"`
	DurationMs int                    `json:"duration_ms"`
	Error      string                 `json:"error,omitempty"`
	Metadata   map[string]interface{} `json:"metadata,omitempty"`
}

// extractDomain extracts the domain from a URL.
func extractDomain(urlStr string) string {
	// Simple domain extraction - handle common URL formats
	urlStr = strings.TrimPrefix(urlStr, "https://")
	urlStr = strings.TrimPrefix(urlStr, "http://")
	urlStr = strings.TrimPrefix(urlStr, "www.")

	parts := strings.Split(urlStr, "/")
	return parts[0]
}

// classifyError classifies a browser error type.
func classifyError(err error) ErrorType {
	errStr := err.Error()
	if strings.Contains(errStr, "crash") || strings.Contains(errStr, "SIGSEGV") {
		return ErrorTypeCrash
	}
	if strings.Contains(errStr, "timeout") || strings.Contains(errStr, "ETIMEDOUT") {
		return ErrorTypeTimeout
	}
	if strings.Contains(errStr, "network") || strings.Contains(errStr, "ENOTFOUND") {
		return ErrorTypeNetwork
	}
	if strings.Contains(errStr, "domain") || strings.Contains(errStr, "blocked") {
		return ErrorTypeDomain
	}
	return ErrorTypeUnknown
}

// recordUsage records browser usage for cost attribution.
func (s *Service) recordUsage(ctx context.Context, agentID string, sessionID uuid.UUID, action, domain string, durationMs int) {
	if s.redis == nil {
		return
	}

	// Calculate browser minutes
	browserMinutes := float64(durationMs) / 60000.0

	// Store in Redis for later aggregation
	key := fmt.Sprintf("browser:usage:%s:%s", agentID, time.Now().UTC().Format("2006-01-02"))
	s.redis.HIncrByFloat(ctx, key, "total_minutes", browserMinutes)
	s.redis.HIncrBy(ctx, key, "total_actions", 1)
	s.redis.Expire(ctx, key, 7*24*time.Hour) // Keep for 7 days
}

// GetSession gets a session by ID.
func (s *Service) GetSession(ctx context.Context, sessionID uuid.UUID) (*SessionState, error) {
	return s.sessionMgr.GetSession(ctx, sessionID)
}

// ListSessions lists all sessions for an agent.
func (s *Service) ListSessions(ctx context.Context, agentID string) ([]*SessionState, error) {
	// Return empty slice for now - implementation needs Redis SCAN
	return []*SessionState{}, nil
}

// StoreCredential stores a browser credential.
func (s *Service) StoreCredential(ctx context.Context, agentID, name, domain string, data *CredentialData) (*BrowserCredential, error) {
	return s.creds.Store(ctx, agentID, name, domain, data)
}

// GetCredential retrieves a browser credential.
func (s *Service) GetCredential(ctx context.Context, agentID, credentialID string) (*BrowserCredential, error) {
	return s.creds.Get(ctx, agentID, credentialID)
}

// ListCredentials lists all credentials for an agent.
func (s *Service) ListCredentials(ctx context.Context, agentID string) ([]*BrowserCredential, error) {
	return s.creds.List(ctx, agentID)
}

// DeleteCredential deletes a browser credential.
func (s *Service) DeleteCredential(ctx context.Context, agentID, credentialID string) error {
	return s.creds.Delete(ctx, agentID, credentialID)
}

// GetPermission gets browser permissions for an agent.
func (s *Service) GetPermission(ctx context.Context, agentID string) (*BrowserPermission, error) {
	return s.policy.GetPermission(ctx, agentID)
}

// UpsertPermission updates browser permissions for an agent.
func (s *Service) UpsertPermission(ctx context.Context, perm *BrowserPermission) error {
	return s.policy.UpsertPermission(ctx, perm)
}

// PoolStats returns pool statistics.
func (s *Service) PoolStats() (int, int) {
	if s.pool == nil {
		return 0, 0
	}
	return s.pool.AvailableCount(), s.pool.AllocatedCount()
}

// HealthCheck returns the health status of browser instances.
func (s *Service) HealthCheck(ctx context.Context) map[string]string {
	status := make(map[string]string)
	status["status"] = "ok"
	status["enabled"] = fmt.Sprintf("%t", s.config.Enabled)
	if s.pool != nil {
		status["pool_available"] = fmt.Sprintf("%d", s.pool.AvailableCount())
		status["pool_allocated"] = fmt.Sprintf("%d", s.pool.AllocatedCount())
	}
	if s.redis != nil {
		if err := s.redis.Ping(ctx); err != nil {
			status["redis"] = "disconnected"
		} else {
			status["redis"] = "connected"
		}
	}
	return status
}

// GetUsageStats returns browser usage statistics for an agent.
func (s *Service) GetUsageStats(ctx context.Context, agentID string) (map[string]interface{}, error) {
	return map[string]interface{}{
		"agent_id": agentID,
		"sessions": 0,
		"usage":    []interface{}{},
	}, nil
}

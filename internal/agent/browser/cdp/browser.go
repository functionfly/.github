package cdp

// This package contains the CDP (Chrome DevTools Protocol) implementation
// for browser automation using chromedp.
//
// The CDPBrowser struct implements the browser automation capabilities
// using persistent WebSocket connections to Chrome instances instead of
// spawning the agent-browser CLI for each action.
//
// Key differences from the CLI-based Service:
// - Persistent WebSocket connections (no CLI spawn per action)
// - Built-in chromedp context management
// - Different session structure with CDPContext
//
// To use this implementation, create a CDPBrowser and call its methods
// directly. See internal/agent/browser/cdp/browser.go for the full
// implementation reference.

import (
	"context"
	"encoding/base64"
	"encoding/json"
	"fmt"
	"sync"
	"time"

	"github.com/chromedp/chromedp"
	"github.com/google/uuid"
	"github.com/redis/go-redis/v9"
	"github.com/sirupsen/logrus"
	"gorm.io/gorm"

	"github.com/functionfly/functionfly/internal/agent/browser"
)

// CDPBrowser implements a chromedp-based browser automation provider.
type CDPBrowser struct {
	config        browser.Config
	db            *gorm.DB
	redis         *redis.Client
	pool          *CDPPool
	isolated      *IsolatedManager
	sessionMgr    *SessionManager
	policy        *PolicyChecker
	creds         *CredentialManager
	recovery      *RecoveryManager
	circuitBreaker *CircuitBreaker
	mu            sync.RWMutex
	stopCh        chan struct{}
}

// SessionOptions configures a new browser session.
type SessionOptions struct {
	Isolated  bool
	TimeoutMs int
	Headless  bool
}

// CDPPool manages Chrome instances for CDP.
type CDPPool struct {
	mu        sync.RWMutex
	config    browser.Config
	browsers  map[int]*BrowserInstance
	allocated map[string]int
	available []int
	stopCh    chan struct{}
}

// BrowserInstance represents a Chrome instance.
type BrowserInstance struct {
	Port      int
	PID       int
	URL       string
	Status    browser.InstanceStatus
	AgentID   string
	cdpCtx    context.Context
	cdpCancel context.CancelFunc
}

// SessionManager manages browser sessions with Redis-backed state.
type SessionManager struct {
	redis      *redis.Client
	sessionTTL time.Duration
}

// IsolatedManager manages per-agent isolated browser instances.
type IsolatedManager struct {
	mu        sync.RWMutex
	config    browser.Config
	instances map[string]*IsolatedInstance
	stopCh    chan struct{}
}

// IsolatedInstance represents an isolated browser instance.
type IsolatedInstance struct {
	ID        uuid.UUID
	AgentID   string
	Port      int
	PID       int
	Status    browser.InstanceStatus
	StartedAt time.Time
	cdpCtx    context.Context
	cdpCancel context.CancelFunc
}

// PolicyChecker validates browser actions.
type PolicyChecker struct {
	db *gorm.DB
}

// CredentialManager handles encrypted credentials.
type CredentialManager struct {
	db       *gorm.DB
	vaultMgr VaultManager
	enabled  bool
}

// VaultManager interface for vault operations.
type VaultManager interface {
	Encrypt(ctx context.Context, plaintext []byte, agentID string) (ciphertext []byte, err error)
	Decrypt(ctx context.Context, ciphertext []byte, agentID string) (plaintext []byte, err error)
}

// DefaultVaultManager is a no-op vault manager.
type DefaultVaultManager struct{}

func (d *DefaultVaultManager) Encrypt(ctx context.Context, plaintext []byte, agentID string) ([]byte, error) {
	return plaintext, nil
}

func (d *DefaultVaultManager) Decrypt(ctx context.Context, ciphertext []byte, agentID string) ([]byte, error) {
	return ciphertext, nil
}

// RecoveryManager handles error recovery.
type RecoveryManager struct {
	config      browser.Config
	sessionMgr  *SessionManager
	mu          sync.RWMutex
	retryCounts map[string]int
}

// CircuitBreaker implements circuit breaker pattern.
type CircuitBreaker struct {
	mu          sync.RWMutex
	failures    map[string]int
	lastFailure map[string]time.Time
	threshold   int
	cooldown    time.Duration
	state       map[string]CircuitState
}

// CircuitState represents circuit breaker state.
type CircuitState string

const (
	CircuitStateClosed   CircuitState = "closed"
	CircuitStateOpen     CircuitState = "open"
	CircuitStateHalfOpen CircuitState = "half-open"
)

// CDPHandle represents an allocated browser handle.
type CDPHandle struct {
	port    int
	pid     int
	healthy bool
}

func (h *CDPHandle) Port() int          { return h.port }
func (h *CDPHandle) PID() int           { return h.pid }
func (h *CDPHandle) IsHealthy() bool    { return h.healthy }

// NewCDPBrowser creates a new CDP browser implementation.
func NewCDPBrowser(config browser.Config, db *gorm.DB, redisClient *redis.Client) *CDPBrowser {
	sessionMgr := NewCDPSessionManager(redisClient, config.SessionTTL)
	pool := NewCDPPool(config, sessionMgr)
	isolated := NewIsolatedManager(config, sessionMgr)
	policy := NewPolicyChecker(db)
	creds := NewCredentialManager(db, nil, config.VaultEnabled)
	recovery := NewRecoveryManager(config, sessionMgr)
	cb := NewCircuitBreaker(5, 30*time.Second)

	return &CDPBrowser{
		config:         config,
		db:             db,
		redis:          redisClient,
		pool:           pool,
		isolated:       isolated,
		sessionMgr:     sessionMgr,
		policy:         policy,
		creds:          creds,
		recovery:       recovery,
		circuitBreaker: cb,
		stopCh:         make(chan struct{}),
	}
}

// Initialize sets up the browser service.
func (c *CDPBrowser) Initialize(ctx context.Context) error {
	if !c.config.Enabled {
		logrus.Info("CDP Browser service: disabled by configuration")
		return nil
	}

	if err := c.db.WithContext(ctx).AutoMigrate(&browser.AgentBrowserConfig{}); err != nil {
		return fmt.Errorf("failed to migrate browser config: %w", err)
	}
	if err := c.db.WithContext(ctx).AutoMigrate(&browser.BrowserSession{}); err != nil {
		return fmt.Errorf("failed to migrate browser session: %w", err)
	}
	if err := c.creds.Migrate(ctx); err != nil {
		return fmt.Errorf("failed to migrate credentials: %w", err)
	}

	if err := c.pool.Initialize(ctx); err != nil {
		return fmt.Errorf("failed to initialize Chrome pool: %w", err)
	}

	c.isolated.StartCleanupRoutine(ctx, 5*time.Minute)

	logrus.Info("CDP Browser service: initialized")
	return nil
}

// Shutdown shuts down the browser service.
func (c *CDPBrowser) Shutdown(ctx context.Context) error {
	close(c.stopCh)
	if c.pool != nil {
		c.pool.Stop()
	}
	if c.isolated != nil {
		c.isolated.Stop()
	}
	return nil
}

// CreateSession creates a new browser session.
func (c *CDPBrowser) CreateSession(ctx context.Context, agentID string, opts SessionOptions) (*CDPSession, error) {
	if c.circuitBreaker.IsOpen(agentID) {
		return nil, fmt.Errorf("circuit breaker open: too many failures")
	}

	var port int
	var sessionType browser.SessionType
	var cdpCtx context.Context
	var cdpCancel context.CancelFunc

	if opts.Isolated {
		instance, err := c.isolated.Acquire(ctx, agentID)
		if err != nil {
			return nil, fmt.Errorf("failed to acquire isolated browser: %w", err)
		}
		port = instance.Port
		sessionType = browser.SessionTypeIsolated
		cdpCtx = instance.cdpCtx
		cdpCancel = instance.cdpCancel
	} else {
		handle, err := c.pool.Acquire(ctx, agentID)
		if err != nil {
			return nil, fmt.Errorf("failed to acquire browser: %w", err)
		}
		port = handle.Port()
		sessionType = browser.SessionTypeShared
		cdpCtx, cdpCancel = c.newCDPContext(ctx, port, opts.TimeoutMs)
	}

	session := &CDPSession{
		ID:          uuid.New(),
		AgentID:     agentID,
		BrowserPort: port,
		SessionType: sessionType,
		Status:      browser.SessionStatusActive,
		CDPContext:  cdpCtx,
		CDPCancel:   cdpCancel,
		CreatedAt:   time.Now().UTC(),
	}

	if err := c.sessionMgr.CreateSession(ctx, session); err != nil {
		cdpCancel()
		return nil, fmt.Errorf("failed to create session: %w", err)
	}

	return session, nil
}

// CDPSession represents a browser session with CDP context.
type CDPSession struct {
	ID           uuid.UUID
	AgentID      string
	BrowserPort  int
	BrowserPID   int
	AuthToken    string
	SessionType  browser.SessionType
	Status       browser.SessionStatus
	CDPContext   context.Context
	CDPCancel    context.CancelFunc
	CreatedAt    time.Time
	ClosedAt     *time.Time
}

// GetSession retrieves a session by ID.
func (c *CDPBrowser) GetSession(ctx context.Context, sessionID uuid.UUID) (*CDPSession, error) {
	return c.sessionMgr.GetSession(ctx, sessionID)
}

// ListSessions lists all sessions for an agent.
func (c *CDPBrowser) ListSessions(ctx context.Context, agentID string) ([]*CDPSession, error) {
	return c.sessionMgr.ListSessions(ctx, agentID)
}

// CloseSession closes a browser session.
func (c *CDPBrowser) CloseSession(ctx context.Context, sessionID uuid.UUID) error {
	session, err := c.sessionMgr.GetSession(ctx, sessionID)
	if err != nil {
		return err
	}

	if session.CDPCancel != nil {
		session.CDPCancel()
	}

	session.Status = browser.SessionStatusClosed
	now := time.Now().UTC()
	session.ClosedAt = &now
	c.sessionMgr.UpdateSession(ctx, session)

	if session.SessionType == browser.SessionTypeShared {
		c.pool.Release(session.AgentID)
	} else {
		c.isolated.Release(session.AgentID)
	}

	return nil
}

// Navigate navigates to a URL.
func (c *CDPBrowser) Navigate(ctx context.Context, sessionID uuid.UUID, url string) (*browser.NavigateResult, error) {
	session, err := c.sessionMgr.GetSession(ctx, sessionID)
	if err != nil {
		return nil, fmt.Errorf("session not found: %w", err)
	}

	domain := extractDomain(url)
	valid, err := c.policy.CheckDomain(ctx, session.AgentID, domain)
	if err != nil {
		return nil, fmt.Errorf("policy check failed: %w", err)
	}
	if !valid {
		return nil, fmt.Errorf("domain %s is not allowed", domain)
	}

	loopDetected, err := c.sessionMgr.CheckLoopDetection(ctx, session.AgentID, sessionID.String(), domain, 3)
	if err != nil {
		logrus.Warnf("Browser loop detection error: %v", err)
	}
	if loopDetected {
		return nil, fmt.Errorf("loop detected: domain %s accessed too many times", domain)
	}

	start := time.Now()
	var title string

	err = chromedp.Run(session.CDPContext,
		chromedp.Navigate(url),
		chromedp.Title(&title),
	)

	duration := time.Since(start)

	if err != nil {
		c.circuitBreaker.RecordFailure(session.AgentID)
		c.recovery.HandleError(ctx, &browser.BrowserError{
			Type:      classifyError(err),
			Message:   err.Error(),
			SessionID: sessionID.String(),
			Domain:    domain,
			Timestamp: time.Now().UTC(),
		})
		return nil, fmt.Errorf("navigation failed: %w", err)
	}

	c.circuitBreaker.RecordSuccess(session.AgentID)
	c.sessionMgr.SetSessionURL(ctx, sessionID, url)
	c.recordUsage(ctx, session.AgentID, sessionID, "navigate", domain, duration.Milliseconds())

	return &browser.NavigateResult{
		SessionID:  sessionID.String(),
		URL:        url,
		Title:      title,
		StatusCode: 200,
		DurationMs: int(duration.Milliseconds()),
	}, nil
}

// Click clicks an element.
func (c *CDPBrowser) Click(ctx context.Context, sessionID uuid.UUID, elementRef string) error {
	session, err := c.sessionMgr.GetSession(ctx, sessionID)
	if err != nil {
		return fmt.Errorf("session not found: %w", err)
	}

	err = chromedp.Run(session.CDPContext, chromedp.Click(elementRef))
	if err != nil {
		c.circuitBreaker.RecordFailure(session.AgentID)
		return fmt.Errorf("click failed: %w", err)
	}

	c.circuitBreaker.RecordSuccess(session.AgentID)
	return nil
}

// Fill fills a form field.
func (c *CDPBrowser) Fill(ctx context.Context, sessionID uuid.UUID, elementRef, value string) error {
	session, err := c.sessionMgr.GetSession(ctx, sessionID)
	if err != nil {
		return fmt.Errorf("session not found: %w", err)
	}

	err = chromedp.Run(session.CDPContext, chromedp.SetValue(elementRef, value))
	if err != nil {
		c.circuitBreaker.RecordFailure(session.AgentID)
		return fmt.Errorf("fill failed: %w", err)
	}

	c.circuitBreaker.RecordSuccess(session.AgentID)
	return nil
}

// Extract extracts structured content from the page.
func (c *CDPBrowser) Extract(ctx context.Context, sessionID uuid.UUID, selector string) ([]map[string]interface{}, error) {
	session, err := c.sessionMgr.GetSession(ctx, sessionID)
	if err != nil {
		return nil, fmt.Errorf("session not found: %w", err)
	}

	var result string
	err = chromedp.Run(session.CDPContext,
		chromedp.EvaluateAsDevTools(
			fmt.Sprintf(`document.querySelectorAll('%s').forEach((el, i) => JSON.stringify({ref: '@e' + i, tag: el.tagName, text: el.innerText, type: el.type || el.tagName}))`, selector),
			&result,
		),
	)

	if err != nil {
		c.circuitBreaker.RecordFailure(session.AgentID)
		return nil, fmt.Errorf("extract failed: %w", err)
	}

	c.circuitBreaker.RecordSuccess(session.AgentID)

	var content []map[string]interface{}
	if err := json.Unmarshal([]byte(result), &content); err != nil {
		return nil, fmt.Errorf("failed to parse extract result: %w", err)
	}

	return content, nil
}

// Screenshot captures a screenshot.
func (c *CDPBrowser) Screenshot(ctx context.Context, sessionID uuid.UUID) (string, error) {
	session, err := c.sessionMgr.GetSession(ctx, sessionID)
	if err != nil {
		return "", fmt.Errorf("session not found: %w", err)
	}

	var screenshot []byte
	err = chromedp.Run(session.CDPContext, chromedp.CaptureScreenshot(&screenshot))
	if err != nil {
		c.circuitBreaker.RecordFailure(session.AgentID)
		return "", fmt.Errorf("screenshot failed: %w", err)
	}

	c.circuitBreaker.RecordSuccess(session.AgentID)
	return fmt.Sprintf("data:image/png;base64,%s", base64.StdEncoding.EncodeToString(screenshot)), nil
}

// StoreCredential stores a browser credential.
func (c *CDPBrowser) StoreCredential(ctx context.Context, agentID, name, domain string, data *browser.CredentialData) (*browser.BrowserCredential, error) {
	return c.creds.Store(ctx, agentID, name, domain, data)
}

// GetCredential retrieves a browser credential.
func (c *CDPBrowser) GetCredential(ctx context.Context, agentID, credentialID string) (*browser.BrowserCredential, error) {
	return c.creds.Get(ctx, agentID, credentialID)
}

// ListCredentials lists all credentials for an agent.
func (c *CDPBrowser) ListCredentials(ctx context.Context, agentID string) ([]*browser.BrowserCredential, error) {
	return c.creds.List(ctx, agentID)
}

// DeleteCredential deletes a browser credential.
func (c *CDPBrowser) DeleteCredential(ctx context.Context, agentID, credentialID string) error {
	return c.creds.Delete(ctx, agentID, credentialID)
}

// GetPermission gets browser permissions for an agent.
func (c *CDPBrowser) GetPermission(ctx context.Context, agentID string) (*browser.BrowserPermission, error) {
	return c.policy.GetPermission(ctx, agentID)
}

// UpsertPermission updates browser permissions for an agent.
func (c *CDPBrowser) UpsertPermission(ctx context.Context, perm *browser.BrowserPermission) error {
	return c.policy.UpsertPermission(ctx, perm)
}

// AcquireBrowser acquires a browser for an agent.
func (c *CDPBrowser) AcquireBrowser(ctx context.Context, agentID string) (browser.BrowserHandle, error) {
	return c.pool.Acquire(ctx, agentID)
}

// ReleaseBrowser releases a browser for an agent.
func (c *CDPBrowser) ReleaseBrowser(ctx context.Context, agentID string) {
	c.pool.Release(agentID)
}

// PoolStats returns pool statistics.
func (c *CDPBrowser) PoolStats() (int, int) {
	if c.pool == nil {
		return 0, 0
	}
	return c.pool.AvailableCount(), c.pool.AllocatedCount()
}

// HealthCheck returns the health status of browser instances.
func (c *CDPBrowser) HealthCheck(ctx context.Context) map[string]string {
	status := make(map[string]string)
	status["status"] = "ok"
	status["enabled"] = fmt.Sprintf("%t", c.config.Enabled)
	if c.pool != nil {
		status["pool_available"] = fmt.Sprintf("%d", c.pool.AvailableCount())
		status["pool_allocated"] = fmt.Sprintf("%d", c.pool.AllocatedCount())
	}
	if c.redis != nil {
		if err := c.redis.Ping(ctx); err != nil {
			status["redis"] = "disconnected"
		} else {
			status["redis"] = "connected"
		}
	}
	return status
}

func (c *CDPBrowser) newCDPContext(ctx context.Context, port int, timeoutMs int) (context.Context, context.CancelFunc) {
	if timeoutMs <= 0 {
		timeoutMs = c.config.DefaultTimeoutMs
	}
	timeout := time.Duration(timeoutMs) * time.Millisecond
	return context.WithTimeout(ctx, timeout)
}

func (c *CDPBrowser) recordUsage(ctx context.Context, agentID string, sessionID uuid.UUID, action, domain string, durationMs int64) {
	if c.redis == nil {
		return
	}
	browserMinutes := float64(durationMs) / 60000.0
	key := fmt.Sprintf("browser:usage:%s:%s", agentID, time.Now().UTC().Format("2006-01-02"))
	c.redis.HIncrByFloat(ctx, key, "total_minutes", browserMinutes)
	c.redis.HIncrBy(ctx, key, "total_actions", 1)
	c.redis.Expire(ctx, key, 7*24*time.Hour)
}

func extractDomain(urlStr string) string {
	urlStr = stripProtocol(urlStr)
	parts := splitDomain(urlStr)
	return parts[0]
}

func stripProtocol(url string) string {
	url = trimPrefix(url, "https://")
	url = trimPrefix(url, "http://")
	url = trimPrefix(url, "www.")
	return url
}

func trimPrefix(s, prefix string) string {
	if len(s) >= len(prefix) && s[:len(prefix)] == prefix {
		return s[len(prefix):]
	}
	return s
}

func splitDomain(s string) []string {
	var parts []string
	var current []byte
	for i := 0; i < len(s); i++ {
		if s[i] == '/' {
			if len(current) > 0 {
				parts = append(parts, string(current))
			}
			break
		}
		current = append(current, s[i])
	}
	if len(current) > 0 {
		parts = append(parts, string(current))
	}
	return parts
}

func classifyError(err error) browser.ErrorType {
	errStr := err.Error()
	if contains(errStr, "crash") || contains(errStr, "SIGSEGV") || contains(errStr, "SIGKILL") {
		return browser.ErrorTypeCrash
	}
	if contains(errStr, "timeout") || contains(errStr, "ETIMEDOUT") {
		return browser.ErrorTypeTimeout
	}
	if contains(errStr, "network") || contains(errStr, "ENOTFOUND") || contains(errStr, "connection") {
		return browser.ErrorTypeNetwork
	}
	if contains(errStr, "domain") || contains(errStr, "blocked") {
		return browser.ErrorTypeDomain
	}
	return browser.ErrorTypeUnknown
}

func contains(s, substr string) bool {
	for i := 0; i <= len(s)-len(substr); i++ {
		if s[i:i+len(substr)] == substr {
			return true
		}
	}
	return false
}

// --- CDPPool implementation ---

func NewCDPPool(config browser.Config, sessionMgr *SessionManager) *CDPPool {
	return &CDPPool{
		config:    config,
		browsers:  make(map[int]*BrowserInstance),
		allocated: make(map[string]int),
		available: make([]int, 0, config.PoolSize),
		stopCh:    make(chan struct{}),
	}
}

func (p *CDPPool) Initialize(ctx context.Context) error {
	for i := 0; i < p.config.PoolSize; i++ {
		p.available = append(p.available, 9222+i)
	}
	go p.healthCheckLoop(ctx)
	return nil
}

func (p *CDPPool) Acquire(ctx context.Context, agentID string) (*CDPHandle, error) {
	p.mu.Lock()
	defer p.mu.Unlock()

	if port, ok := p.allocated[agentID]; ok {
		if inst, exists := p.browsers[port]; exists && inst.Status == browser.InstanceStatusRunning {
			return &CDPHandle{port: port, pid: inst.PID, healthy: true}, nil
		}
	}

	if len(p.available) == 0 {
		return nil, fmt.Errorf("browser pool exhausted")
	}

	port := p.available[len(p.available)-1]
	p.available = p.available[:len(p.available)-1]
	p.allocated[agentID] = port

	instance, err := p.launchChrome(ctx, port)
	if err != nil {
		delete(p.allocated, agentID)
		p.available = append(p.available, port)
		return nil, err
	}

	p.browsers[port] = instance
	return &CDPHandle{port: port, pid: instance.PID, healthy: true}, nil
}

func (p *CDPPool) Release(agentID string) {
	p.mu.Lock()
	defer p.mu.Unlock()

	port, ok := p.allocated[agentID]
	if !ok {
		return
	}

	delete(p.allocated, agentID)
	p.available = append(p.available, port)
}

func (p *CDPPool) AvailableCount() int {
	p.mu.RLock()
	defer p.mu.RUnlock()
	return len(p.available)
}

func (p *CDPPool) AllocatedCount() int {
	p.mu.RLock()
	defer p.mu.RUnlock()
	return len(p.allocated)
}

func (p *CDPPool) Stop() {
	close(p.stopCh)
	p.mu.Lock()
	defer p.mu.Unlock()
	for _, inst := range p.browsers {
		if inst.cdpCancel != nil {
			inst.cdpCancel()
		}
	}
}

func (p *CDPPool) launchChrome(ctx context.Context, port int) (*BrowserInstance, error) {
	cdpCtx, cdpCancel := context.WithCancel(ctx)
	instance := &BrowserInstance{
		Port:     port,
		Status:   browser.InstanceStatusStarting,
		cdpCtx:   cdpCtx,
		cdpCancel: cdpCancel,
	}
	instance.Status = browser.InstanceStatusRunning
	return instance, nil
}

func (p *CDPPool) healthCheckLoop(ctx context.Context) {
	ticker := time.NewTicker(30 * time.Second)
	defer ticker.Stop()
	for {
		select {
		case <-ctx.Done():
			return
		case <-p.stopCh:
			return
		case <-ticker.C:
			p.checkHealth(ctx)
		}
	}
}

func (p *CDPPool) checkHealth(ctx context.Context) {
	p.mu.RLock()
	defer p.mu.RUnlock()
	for port, inst := range p.browsers {
		if inst.Status == browser.InstanceStatusRunning {
			logrus.Debugf("CDP Pool: browser on port %d is running", port)
		}
	}
}

// --- SessionManager implementation ---

func NewCDPSessionManager(redisClient *redis.Client, sessionTTL time.Duration) *SessionManager {
	return &SessionManager{
		redis:      redisClient,
		sessionTTL: sessionTTL,
	}
}

func (sm *SessionManager) CreateSession(ctx context.Context, session *CDPSession) error {
	data, err := json.Marshal(session)
	if err != nil {
		return fmt.Errorf("failed to marshal session: %w", err)
	}
	key := sm.sessionKey(session.ID)
	if err := sm.redis.Set(ctx, key, data, sm.sessionTTL).Err(); err != nil {
		return fmt.Errorf("failed to store session: %w", err)
	}
	if session.SessionType == browser.SessionTypeShared {
		affinityKey := sm.affinityKey(session.AgentID)
		sm.redis.Set(ctx, affinityKey, session.ID.String(), sm.sessionTTL)
	}
	return nil
}

func (sm *SessionManager) GetSession(ctx context.Context, sessionID uuid.UUID) (*CDPSession, error) {
	key := sm.sessionKey(sessionID)
	data, err := sm.redis.Get(ctx, key).Bytes()
	if err != nil {
		if err == redis.Nil {
			return nil, fmt.Errorf("session not found: %s", sessionID)
		}
		return nil, fmt.Errorf("failed to get session: %w", err)
	}
	var session CDPSession
	if err := json.Unmarshal(data, &session); err != nil {
		return nil, fmt.Errorf("failed to unmarshal session: %w", err)
	}
	return &session, nil
}

func (sm *SessionManager) ListSessions(ctx context.Context, agentID string) ([]*CDPSession, error) {
	pattern := "browser:session:*"
	var sessions []*CDPSession
	iter := sm.redis.Scan(ctx, 0, pattern, 100).Iterator()
	for iter.Next(ctx) {
		key := iter.Val()
		data, err := sm.redis.Get(ctx, key).Bytes()
		if err != nil {
			continue
		}
		var session CDPSession
		if err := json.Unmarshal(data, &session); err != nil {
			continue
		}
		if session.AgentID == agentID {
			sessions = append(sessions, &session)
		}
	}
	return sessions, nil
}

func (sm *SessionManager) UpdateSession(ctx context.Context, session *CDPSession) error {
	data, err := json.Marshal(session)
	if err != nil {
		return fmt.Errorf("failed to marshal session: %w", err)
	}
	key := sm.sessionKey(session.ID)
	return sm.redis.Set(ctx, key, data, sm.sessionTTL).Err()
}

func (sm *SessionManager) SetSessionURL(ctx context.Context, sessionID uuid.UUID, url string) error {
	session, err := sm.GetSession(ctx, sessionID)
	if err != nil {
		return err
	}
	session.AuthToken = url
	return sm.UpdateSession(ctx, session)
}

func (sm *SessionManager) CheckLoopDetection(ctx context.Context, agentID, sessionID, domain string, threshold int) (bool, error) {
	key := sm.loopDetectionKey(agentID, sessionID, domain)
	count, err := sm.redis.Incr(ctx, key).Result()
	if err != nil {
		return false, fmt.Errorf("failed to increment loop counter: %w", err)
	}
	if count == 1 {
		sm.redis.Expire(ctx, key, sm.sessionTTL)
	}
	return count > int64(threshold), nil
}

func (sm *SessionManager) sessionKey(sessionID uuid.UUID) string {
	return fmt.Sprintf("browser:session:%s", sessionID.String())
}

func (sm *SessionManager) affinityKey(agentID string) string {
	return fmt.Sprintf("browser:affinity:%s", agentID)
}

func (sm *SessionManager) loopDetectionKey(agentID, sessionID, domain string) string {
	return fmt.Sprintf("browser:loop:%s:%s:%s", agentID, sessionID, domain)
}

// --- IsolatedManager implementation ---

func NewIsolatedManager(config browser.Config, sessionMgr *SessionManager) *IsolatedManager {
	return &IsolatedManager{
		config:    config,
		instances: make(map[string]*IsolatedInstance),
		stopCh:    make(chan struct{}),
	}
}

func (im *IsolatedManager) Acquire(ctx context.Context, agentID string) (*IsolatedInstance, error) {
	im.mu.Lock()
	defer im.mu.Unlock()

	if instance, ok := im.instances[agentID]; ok {
		if instance.Status == browser.InstanceStatusRunning {
			return instance, nil
		}
	}

	port := im.allocatePort()
	instance := &IsolatedInstance{
		ID:        uuid.New(),
		AgentID:   agentID,
		Port:      port,
		Status:    browser.InstanceStatusRunning,
		StartedAt: time.Now().UTC(),
	}

	im.instances[agentID] = instance
	return instance, nil
}

func (im *IsolatedManager) Release(agentID string) {
	im.mu.Lock()
	defer im.mu.Unlock()

	if instance, ok := im.instances[agentID]; ok {
		instance.Status = browser.InstanceStatusStopped
	}
}

func (im *IsolatedManager) allocatePort() int {
	basePort := 19222
	usedPorts := make(map[int]bool)
	im.mu.RLock()
	for _, instance := range im.instances {
		usedPorts[instance.Port] = true
	}
	im.mu.RUnlock()
	for port := basePort; port < basePort+1000; port++ {
		if !usedPorts[port] {
			return port
		}
	}
	return 0
}

func (im *IsolatedManager) StartCleanupRoutine(ctx context.Context, interval time.Duration) {
	go func() {
		ticker := time.NewTicker(interval)
		defer ticker.Stop()
		for {
			select {
			case <-ctx.Done():
				return
			case <-im.stopCh:
				return
			case <-ticker.C:
				im.cleanup(ctx)
			}
		}
	}()
}

func (im *IsolatedManager) cleanup(ctx context.Context) {
	im.mu.Lock()
	defer im.mu.Unlock()
	cutoff := time.Now().UTC().Add(-im.config.SessionTTL)
	for agentID, instance := range im.instances {
		if instance.Status == browser.InstanceStatusStopped && instance.StartedAt.Before(cutoff) {
			delete(im.instances, agentID)
		}
	}
}

func (im *IsolatedManager) Stop() {
	close(im.stopCh)
}

// --- PolicyChecker implementation ---

func NewPolicyChecker(db *gorm.DB) *PolicyChecker {
	return &PolicyChecker{db: db}
}

func (pc *PolicyChecker) CheckDomain(ctx context.Context, agentID, domain string) (bool, error) {
	perm, err := pc.GetPermission(ctx, agentID)
	if err != nil {
		return true, nil
	}
	return pc.checkDomainWithPermission(domain, perm), nil
}

func (pc *PolicyChecker) GetPermission(ctx context.Context, agentID string) (*browser.BrowserPermission, error) {
	var config browser.AgentBrowserConfig
	err := pc.db.WithContext(ctx).Where("agent_id = ?", agentID).First(&config).Error
	if err != nil {
		if err == gorm.ErrRecordNotFound {
			return browser.DefaultBrowserPermission(agentID), nil
		}
		return nil, err
	}
	return &browser.BrowserPermission{
		AgentID:           agentID,
		BrowserEnabled:    config.BrowserEnabled,
		AllowedDomains:    config.AllowedDomains,
		MaxSessions:       config.MaxSessions,
		CredentialStorage: config.CredentialStorageEnabled,
		DefaultTimeoutMs:  config.DefaultTimeoutMs,
		HeadfulMode:       config.HeadfulMode,
	}, nil
}

func (pc *PolicyChecker) UpsertPermission(ctx context.Context, perm *browser.BrowserPermission) error {
	config := browser.AgentBrowserConfig{
		AgentID:                   perm.AgentID,
		BrowserEnabled:            perm.BrowserEnabled,
		AllowedDomains:            perm.AllowedDomains,
		MaxSessions:               perm.MaxSessions,
		CredentialStorageEnabled:  perm.CredentialStorage,
		DefaultTimeoutMs:          perm.DefaultTimeoutMs,
		HeadfulMode:               perm.HeadfulMode,
	}
	return pc.db.WithContext(ctx).
		Where("agent_id = ?", perm.AgentID).
		Assign(config).
		FirstOrCreate(&config).Error
}

func (pc *PolicyChecker) checkDomainWithPermission(domain string, perm *browser.BrowserPermission) bool {
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

func matchDomainPattern(pattern, domain string) bool {
	if pattern == domain {
		return true
	}
	if len(pattern) >= 2 && pattern[0] == '*' && pattern[1] == '.' {
		suffix := pattern[2:]
		return len(domain) > len(suffix) && domain[len(domain)-len(suffix):] == suffix
	}
	return false
}

// --- CredentialManager implementation ---

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

func (cm *CredentialManager) Store(ctx context.Context, agentID, name, domain string, data *browser.CredentialData) (*browser.BrowserCredential, error) {
	if !cm.enabled {
		return nil, fmt.Errorf("credential storage is disabled")
	}
	plaintext, err := json.Marshal(data)
	if err != nil {
		return nil, fmt.Errorf("failed to marshal credential data: %w", err)
	}
	ciphertext, err := cm.vaultMgr.Encrypt(ctx, plaintext, agentID)
	if err != nil {
		return nil, fmt.Errorf("failed to encrypt credential: %w", err)
	}
	credential := &browser.BrowserCredential{
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

func (cm *CredentialManager) Get(ctx context.Context, agentID, credentialID string) (*browser.BrowserCredential, error) {
	var credential browser.BrowserCredential
	err := cm.db.WithContext(ctx).Where("id = ? AND agent_id = ?", credentialID, agentID).First(&credential).Error
	if err != nil {
		return nil, fmt.Errorf("credential not found: %w", err)
	}
	return &credential, nil
}

func (cm *CredentialManager) List(ctx context.Context, agentID string) ([]*browser.BrowserCredential, error) {
	var credentials []*browser.BrowserCredential
	err := cm.db.WithContext(ctx).Where("agent_id = ?", agentID).Find(&credentials).Error
	if err != nil {
		return nil, fmt.Errorf("failed to list credentials: %w", err)
	}
	return credentials, nil
}

func (cm *CredentialManager) Delete(ctx context.Context, agentID, credentialID string) error {
	result := cm.db.WithContext(ctx).Where("id = ? AND agent_id = ?", credentialID, agentID).Delete(&browser.BrowserCredential{})
	if result.Error != nil {
		return fmt.Errorf("failed to delete credential: %w", result.Error)
	}
	if result.RowsAffected == 0 {
		return fmt.Errorf("credential not found")
	}
	return nil
}

func (cm *CredentialManager) Migrate(ctx context.Context) error {
	return cm.db.WithContext(ctx).AutoMigrate(&browser.BrowserCredential{})
}

// --- RecoveryManager implementation ---

func NewRecoveryManager(config browser.Config, sessionMgr *SessionManager) *RecoveryManager {
	return &RecoveryManager{
		config:      config,
		sessionMgr:  sessionMgr,
		retryCounts: make(map[string]int),
	}
}

func (rm *RecoveryManager) HandleError(ctx context.Context, err *browser.BrowserError) error {
	rm.mu.Lock()
	defer rm.mu.Unlock()

	key := err.SessionID
	count := rm.retryCounts[key]
	if count == 0 {
		rm.retryCounts[key] = 1
	} else {
		rm.retryCounts[key] = count + 1
	}

	if count >= rm.config.MaxRetries {
		delete(rm.retryCounts, key)
		return fmt.Errorf("max retries exceeded: %s", err.Message)
	}

	action := rm.DetermineRecoveryAction(err)
	switch action {
	case browser.RecoveryActionRetry:
		time.Sleep(rm.config.RetryBackoff * time.Duration(count+1))
	case browser.RecoveryActionRestart:
		logrus.Warnf("Browser recovery: restarting session %s", err.SessionID)
	case browser.RecoveryActionFail:
		delete(rm.retryCounts, key)
	}

	return nil
}

func (rm *RecoveryManager) DetermineRecoveryAction(err *browser.BrowserError) browser.RecoveryAction {
	switch err.Type {
	case browser.ErrorTypeCrash:
		return browser.RecoveryActionRestart
	case browser.ErrorTypeTimeout, browser.ErrorTypeNetwork:
		return browser.RecoveryActionRetry
	case browser.ErrorTypeDomain:
		return browser.RecoveryActionFail
	default:
		return browser.RecoveryActionRetry
	}
}

// --- CircuitBreaker implementation ---

func NewCircuitBreaker(threshold int, cooldown time.Duration) *CircuitBreaker {
	return &CircuitBreaker{
		failures:    make(map[string]int),
		lastFailure: make(map[string]time.Time),
		threshold:   threshold,
		cooldown:    cooldown,
		state:       make(map[string]CircuitState),
	}
}

func (cb *CircuitBreaker) IsOpen(agentID string) bool {
	cb.mu.RLock()
	defer cb.mu.RUnlock()
	state, ok := cb.state[agentID]
	if !ok {
		return false
	}
	if state == CircuitStateOpen {
		lastFailure := cb.lastFailure[agentID]
		if time.Since(lastFailure) > cb.cooldown {
			return false
		}
		return true
	}
	return false
}

func (cb *CircuitBreaker) RecordFailure(agentID string) {
	cb.mu.Lock()
	defer cb.mu.Unlock()
	cb.failures[agentID]++
	cb.lastFailure[agentID] = time.Now()
	if cb.failures[agentID] >= cb.threshold {
		cb.state[agentID] = CircuitStateOpen
	}
}

func (cb *CircuitBreaker) RecordSuccess(agentID string) {
	cb.mu.Lock()
	defer cb.mu.Unlock()
	delete(cb.failures, agentID)
	delete(cb.lastFailure, agentID)
	delete(cb.state, agentID)
}
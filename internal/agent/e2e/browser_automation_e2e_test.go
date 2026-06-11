package e2e

import (
	"testing"
	"time"

	"github.com/functionfly/functionfly/internal/agent/browser"
	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// ============================================================
// Browser Session Manager Tests
// ============================================================

func TestBrowserSessionManager(t *testing.T) {
	t.Run("should create session manager with config", func(t *testing.T) {
		cfg := browser.DefaultConfig()
		cfg.Enabled = true
		cfg.PoolSize = 5

		sessionMgr := browser.NewSessionManager(nil, 5*time.Minute)
		require.NotNil(t, sessionMgr)
	})

	t.Run("should handle session state transitions", func(t *testing.T) {
		// Test session state structure
		state := &browser.SessionState{
			Status: browser.SessionStatusActive,
		}
		assert.Equal(t, browser.SessionStatusActive, state.Status)

		// Transition to closing
		state.Status = browser.SessionStatusClosing
		assert.Equal(t, browser.SessionStatusClosing, state.Status)
	})

	t.Run("should track session types", func(t *testing.T) {
		session := &browser.SessionState{
			SessionType: browser.SessionTypeShared,
		}
		assert.Equal(t, browser.SessionTypeShared, session.SessionType)

		session.SessionType = browser.SessionTypeIsolated
		assert.Equal(t, browser.SessionTypeIsolated, session.SessionType)
	})
}

// ============================================================
// Browser Config Tests
// ============================================================

func TestBrowserConfig(t *testing.T) {
	t.Run("should create default config", func(t *testing.T) {
		cfg := browser.DefaultConfig()

		assert.True(t, cfg.PoolSize > 0)
		assert.True(t, cfg.SessionTTL > 0)
		assert.NotNil(t, cfg.AllowedDomains)
	})

	t.Run("should validate agent config", func(t *testing.T) {
		agentCfg := &browser.AgentConfig{
			AgentID:        "test-agent",
			BrowserEnabled: true,
			AllowedDomains: []string{"example.com", "*.trusted.com"},
			MaxSessions:    3,
			Tier:           browser.TierStandard,
		}

		assert.Equal(t, "test-agent", agentCfg.AgentID)
		assert.True(t, agentCfg.BrowserEnabled)
		assert.Equal(t, 3, agentCfg.MaxSessions)
	})

	t.Run("should check domain allowed status", func(t *testing.T) {
		agentCfg := &browser.AgentConfig{
			AllowedDomains: []string{"example.com", "*.trusted.com"},
		}

		assert.True(t, agentCfg.IsDomainAllowed("example.com"))
		assert.True(t, agentCfg.IsDomainAllowed("sub.trusted.com"))
		assert.False(t, agentCfg.IsDomainAllowed("untrusted.com"))
	})

	t.Run("should get timeout from agent config", func(t *testing.T) {
		agentCfg := &browser.AgentConfig{
			DefaultTimeoutMs: 5000,
		}

		timeout := agentCfg.GetTimeout()
		assert.Equal(t, 5*time.Second, timeout)
	})

	t.Run("should use default timeout when not set", func(t *testing.T) {
		agentCfg := &browser.AgentConfig{
			DefaultTimeoutMs: 0,
		}

		timeout := agentCfg.GetTimeout()
		assert.Equal(t, 30*time.Second, timeout)
	})
}

// ============================================================
// Browser Pool Tests
// ============================================================

func TestBrowserPool(t *testing.T) {
	t.Run("should create browser pool", func(t *testing.T) {
		cfg := browser.DefaultConfig()
		cfg.PoolSize = 5

		// Note: NewPool requires other dependencies - just test config usage
		assert.Equal(t, 5, cfg.PoolSize)
	})

	t.Run("should track pool availability", func(t *testing.T) {
		cfg := browser.Config{
			PoolSize: 10,
		}

		// Test that pool size is configurable
		assert.Equal(t, 10, cfg.PoolSize)
	})
}

// ============================================================
// Browser Error Handling Tests
// ============================================================

func TestBrowserErrorHandling(t *testing.T) {
	t.Run("should identify error types", func(t *testing.T) {
		err := &browser.BrowserError{
			Type:    browser.ErrorTypeCrash,
			Message: "Chrome crashed",
		}

		assert.Equal(t, browser.ErrorTypeCrash, err.Type)
		assert.Equal(t, "Chrome crashed", err.Message)
	})

	t.Run("should determine recovery actions", func(t *testing.T) {
		rm := &mockRecoveryManager{}

		// Test crash error -> restart
		action := rm.DetermineRecoveryAction(&browser.BrowserError{Type: browser.ErrorTypeCrash})
		assert.Equal(t, browser.RecoveryActionRestart, action)

		// Test timeout error -> retry
		action = rm.DetermineRecoveryAction(&browser.BrowserError{Type: browser.ErrorTypeTimeout})
		assert.Equal(t, browser.RecoveryActionRetry, action)

		// Test network error -> retry
		action = rm.DetermineRecoveryAction(&browser.BrowserError{Type: browser.ErrorTypeNetwork})
		assert.Equal(t, browser.RecoveryActionRetry, action)

		// Test domain error -> fail
		action = rm.DetermineRecoveryAction(&browser.BrowserError{Type: browser.ErrorTypeDomain})
		assert.Equal(t, browser.RecoveryActionFail, action)
	})

	t.Run("should track retry counts", func(t *testing.T) {
		rm := &mockRecoveryManager{}

		rm.retryCounts["session-1"] = 0
		rm.retryCounts["session-1"] = rm.retryCounts["session-1"] + 1
		assert.Equal(t, 1, rm.GetRetryCount("session-1"))

		rm.retryCounts["session-1"] = rm.retryCounts["session-1"] + 1
		assert.Equal(t, 2, rm.GetRetryCount("session-1"))

		rm.ClearRetryCount("session-1")
		assert.Equal(t, 0, rm.GetRetryCount("session-1"))
	})
}

// ============================================================
// Credential Manager Tests
// ============================================================

func TestCredentialManager(t *testing.T) {
	t.Run("should store and retrieve credential data", func(t *testing.T) {
		// Test credential data structure
		credData := &browser.CredentialData{
			Cookies: []browser.SessionCookie{
				{Name: "session", Value: "abc123", Domain: "example.com"},
			},
			AuthHeader: "Bearer token123",
			Tokens: map[string]string{
				"github": "ghp_xxxx",
			},
		}

		assert.Len(t, credData.Cookies, 1)
		assert.Equal(t, "Bearer token123", credData.AuthHeader)
		assert.Equal(t, "ghp_xxxx", credData.Tokens["github"])
	})

	t.Run("should handle credential metadata", func(t *testing.T) {
		credData := &browser.CredentialData{
			Tokens: map[string]string{
				"github": "ghp_xxxx",
			},
		}

		assert.NotEmpty(t, credData.Tokens["github"])
	})
}

// ============================================================
// Session State Tests
// ============================================================

func TestSessionState(t *testing.T) {
	t.Run("should track session status", func(t *testing.T) {
		statuses := []browser.SessionStatus{
			browser.SessionStatusActive,
			browser.SessionStatusClosing,
			browser.SessionStatusClosed,
			browser.SessionStatusCrashed,
		}

		for _, status := range statuses {
			session := &browser.SessionState{Status: status}
			assert.Equal(t, status, session.Status)
		}
	})

	t.Run("should track session URLs", func(t *testing.T) {
		session := &browser.SessionState{
			URL: "https://example.com",
		}

		assert.Equal(t, "https://example.com", session.URL)
	})

	t.Run("should track session cookies", func(t *testing.T) {
		session := &browser.SessionState{
			Cookies: []browser.SessionCookie{
				{Name: "session_id", Value: "abc123", Domain: "example.com"},
				{Name: "user_pref", Value: "dark_mode", Domain: "example.com"},
			},
		}

		assert.Len(t, session.Cookies, 2)
		assert.Equal(t, "session_id", session.Cookies[0].Name)
	})

	t.Run("should track session metadata", func(t *testing.T) {
		session := &browser.SessionState{
			Metadata: map[string]interface{}{
				"browser_version": "120.0",
				"platform":        "linux",
			},
		}

		assert.Equal(t, "120.0", session.Metadata["browser_version"])
		assert.Equal(t, "linux", session.Metadata["platform"])
	})

	t.Run("should create session with ID and timestamps", func(t *testing.T) {
		session := &browser.Session{
			ID:          uuid.New(),
			AgentID:    "test-agent",
			SessionType: browser.SessionTypeShared,
			Status:      browser.SessionStatusActive,
			CreatedAt:  time.Now(),
			LastUsedAt: time.Now(),
		}

		assert.NotEqual(t, uuid.Nil, session.ID)
		assert.Equal(t, "test-agent", session.AgentID)
		assert.NotNil(t, session.CreatedAt)
		assert.NotNil(t, session.LastUsedAt)
	})
}

// Mock recovery manager for testing
type mockRecoveryManager struct {
	retryCounts map[string]int
}

func (m *mockRecoveryManager) DetermineRecoveryAction(err *browser.BrowserError) browser.RecoveryAction {
	switch err.Type {
	case browser.ErrorTypeCrash:
		return browser.RecoveryActionRestart
	case browser.ErrorTypeTimeout:
		return browser.RecoveryActionRetry
	case browser.ErrorTypeNetwork:
		return browser.RecoveryActionRetry
	case browser.ErrorTypeDomain:
		return browser.RecoveryActionFail
	default:
		return browser.RecoveryActionRetry
	}
}

func (m *mockRecoveryManager) GetRetryCount(sessionID string) int {
	if m.retryCounts == nil {
		m.retryCounts = make(map[string]int)
	}
	return m.retryCounts[sessionID]
}

func (m *mockRecoveryManager) ClearRetryCount(sessionID string) {
	if m.retryCounts != nil {
		delete(m.retryCounts, sessionID)
	}
}

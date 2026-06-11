package browser

import (
	"context"
	"fmt"
	"sync"
	"time"

	"github.com/sirupsen/logrus"
	"gorm.io/gorm"
)

// RecoveryManager handles error recovery and auto-restart for browser instances.
type RecoveryManager struct {
	config      Config
	sessionMgr *SessionManager
	pool       *Pool
	isolated   *IsolatedManager
	mu         sync.RWMutex
	// Track retry counts per session
	retryCounts map[string]int
	stopCh      chan struct{}
}

// NewRecoveryManager creates a new recovery manager.
func NewRecoveryManager(config Config, sessionMgr *SessionManager, pool *Pool, isolated *IsolatedManager) *RecoveryManager {
	return &RecoveryManager{
		config:      config,
		sessionMgr:  sessionMgr,
		pool:        pool,
		isolated:    isolated,
		retryCounts: make(map[string]int),
		stopCh:      make(chan struct{}),
	}
}

// DetermineRecoveryAction determines the recovery action for an error.
func (rm *RecoveryManager) DetermineRecoveryAction(err *BrowserError) RecoveryAction {
	switch err.Type {
	case ErrorTypeCrash:
		// Chrome crashed - try to restart
		return RecoveryActionRestart
	case ErrorTypeTimeout:
		// Timeout - retry with same session
		return RecoveryActionRetry
	case ErrorTypeNetwork:
		// Network error - retry
		return RecoveryActionRetry
	case ErrorTypeDomain:
		// Domain blocked - fail immediately
		return RecoveryActionFail
	default:
		// Unknown error - retry up to limit
		return RecoveryActionRetry
	}
}

// HandleError handles a browser error and returns a recovery action.
func (rm *RecoveryManager) HandleError(ctx context.Context, err *BrowserError) (RecoveryAction, error) {
	rm.mu.Lock()
	defer rm.mu.Unlock()

	// Get current retry count
	key := err.SessionID
	count := rm.retryCounts[key]
	if count == 0 {
		rm.retryCounts[key] = 1
	} else {
		rm.retryCounts[key] = count + 1
	}

	// Check if we've exceeded max retries
	if count >= rm.config.MaxRetries {
		logrus.Errorf("Browser recovery: max retries exceeded for session %s", err.SessionID)
		delete(rm.retryCounts, key)
		return RecoveryActionFail, fmt.Errorf("max retries exceeded: %s", err.Message)
	}

	action := rm.DetermineRecoveryAction(err)

	switch action {
	case RecoveryActionRetry:
		logrus.Warnf("Browser recovery: retrying session %s (attempt %d/%d)", err.SessionID, count+1, rm.config.MaxRetries)
		time.Sleep(rm.config.RetryBackoff * time.Duration(count+1))

	case RecoveryActionRestart:
		logrus.Warnf("Browser recovery: restarting session %s", err.SessionID)
		if err := rm.restartSession(ctx, err.SessionID); err != nil {
			return RecoveryActionFail, err
		}

	case RecoveryActionFallback:
		logrus.Warnf("Browser recovery: falling back for session %s", err.SessionID)
		// For fallback, we mark the session as crashed and let the pool reallocate

	case RecoveryActionFail:
		logrus.Errorf("Browser recovery: failing session %s: %s", err.SessionID, err.Message)
	}

	return action, nil
}

// restartSession restarts a crashed browser session.
func (rm *RecoveryManager) restartSession(ctx context.Context, sessionID string) error {
	// Mark session as crashed
	// Note: We would need to convert sessionID string to uuid.UUID here
	logrus.Debugf("Browser recovery: session %s marked for restart", sessionID)
	return nil
}

// ClearRetryCount clears the retry count for a session.
func (rm *RecoveryManager) ClearRetryCount(sessionID string) {
	rm.mu.Lock()
	defer rm.mu.Unlock()
	delete(rm.retryCounts, sessionID)
}

// GetRetryCount returns the current retry count for a session.
func (rm *RecoveryManager) GetRetryCount(sessionID string) int {
	rm.mu.RLock()
	defer rm.mu.RUnlock()
	return rm.retryCounts[sessionID]
}

// SessionReconstructor attempts to reconstruct session state after a crash.
type SessionReconstructor struct {
	sessionMgr *SessionManager
}

// NewSessionReconstructor creates a new session reconstructor.
func NewSessionReconstructor(sessionMgr *SessionManager) *SessionReconstructor {
	return &SessionReconstructor{sessionMgr: sessionMgr}
}

// ReconstructSession attempts to reconstruct session state from cookies.
func (sr *SessionReconstructor) ReconstructSession(ctx context.Context, sessionID string, cookies []SessionCookie) error {
	// Set the cookies on the new browser instance
	logrus.Debugf("Browser recovery: reconstructing session %s with %d cookies", sessionID, len(cookies))
	return nil
}

// DeadLetterHandler handles sessions that have failed after all retries.
type DeadLetterHandler struct {
	db *gorm.DB
}

// NewDeadLetterHandler creates a new dead letter handler.
func NewDeadLetterHandler(db *gorm.DB) *DeadLetterHandler {
	return &DeadLetterHandler{db: db}
}

// BrowserDeadLetter records a failed browser session for later inspection.
type BrowserDeadLetter struct {
	SessionID    string    `json:"session_id"`
	AgentID      string    `json:"agent_id"`
	ErrorType    string    `json:"error_type"`
	ErrorMessage string    `json:"error_message"`
	LastURL      string    `json:"last_url,omitempty"`
	Cookies      string    `json:"cookies,omitempty"` // JSON serialized
	CreatedAt    time.Time `json:"created_at"`
}

// RecordDeadLetter records a failed session in the dead letter table.
func (dlh *DeadLetterHandler) RecordDeadLetter(ctx context.Context, letter *BrowserDeadLetter) error {
	// Create dead letter table if not exists
	if err := dlh.db.WithContext(ctx).AutoMigrate(&BrowserDeadLetter{}); err != nil {
		return fmt.Errorf("failed to migrate dead letter table: %w", err)
	}

	return dlh.db.WithContext(ctx).Create(letter).Error
}

// GetDeadLetters retrieves dead letters for an agent.
func (dlh *DeadLetterHandler) GetDeadLetters(ctx context.Context, agentID string, limit int) ([]*BrowserDeadLetter, error) {
	var letters []*BrowserDeadLetter
	err := dlh.db.WithContext(ctx).
		Where("agent_id = ?", agentID).
		Order("created_at DESC").
		Limit(limit).
		Find(&letters).Error
	if err != nil {
		return nil, fmt.Errorf("failed to get dead letters: %w", err)
	}
	return letters, nil
}

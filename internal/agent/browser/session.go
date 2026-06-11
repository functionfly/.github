package browser

import (
	"context"
	"encoding/json"
	"fmt"
	"time"

	"github.com/google/uuid"
	"github.com/redis/go-redis/v9"
)

// SessionManager manages browser sessions with Redis-backed state.
type SessionManager struct {
	redis      *redis.Client
	sessionTTL time.Duration
}

// NewSessionManager creates a new session manager.
func NewSessionManager(redisClient *redis.Client, sessionTTL time.Duration) *SessionManager {
	return &SessionManager{
		redis:      redisClient,
		sessionTTL: sessionTTL,
	}
}

// CreateSession creates a new browser session.
func (sm *SessionManager) CreateSession(ctx context.Context, agentID string, sessionType SessionType, browserPort int) (*SessionState, error) {
	session := &SessionState{
		ID:          uuid.New(),
		AgentID:     agentID,
		SessionType: sessionType,
		Status:      SessionStatusActive,
		BrowserPort: browserPort,
		CreatedAt:   time.Now().UTC(),
		LastUsedAt:  time.Now().UTC(),
		Metadata:    make(map[string]interface{}),
	}

	// Generate auth token for this session
	session.AuthToken = uuid.New().String()

	key := sm.sessionKey(session.ID)
	data, err := json.Marshal(session)
	if err != nil {
		return nil, fmt.Errorf("failed to marshal session: %w", err)
	}

	if err := sm.redis.Set(ctx, key, data, sm.sessionTTL).Err(); err != nil {
		return nil, fmt.Errorf("failed to store session: %w", err)
	}

	// Store agent-to-session affinity for sticky sessions
	if sessionType == SessionTypeShared {
		affinityKey := sm.affinityKey(agentID)
		sm.redis.Set(ctx, affinityKey, session.ID.String(), sm.sessionTTL)
	}

	return session, nil
}

// GetSession retrieves a session by ID.
func (sm *SessionManager) GetSession(ctx context.Context, sessionID uuid.UUID) (*SessionState, error) {
	key := sm.sessionKey(sessionID)
	data, err := sm.redis.Get(ctx, key).Bytes()
	if err != nil {
		if err == redis.Nil {
			return nil, fmt.Errorf("session not found: %s", sessionID)
		}
		return nil, fmt.Errorf("failed to get session: %w", err)
	}

	var session SessionState
	if err := json.Unmarshal(data, &session); err != nil {
		return nil, fmt.Errorf("failed to unmarshal session: %w", err)
	}

	return &session, nil
}

// GetSessionByAgent retrieves the current session for an agent (sticky session).
func (sm *SessionManager) GetSessionByAgent(ctx context.Context, agentID string) (*SessionState, error) {
	affinityKey := sm.affinityKey(agentID)
	sessionIDStr, err := sm.redis.Get(ctx, affinityKey).Result()
	if err != nil {
		if err == redis.Nil {
			return nil, nil // No sticky session
		}
		return nil, fmt.Errorf("failed to get affinity: %w", err)
	}

	sessionID, err := uuid.Parse(sessionIDStr)
	if err != nil {
		return nil, fmt.Errorf("invalid session ID: %w", err)
	}

	return sm.GetSession(ctx, sessionID)
}

// UpdateSession updates a session's state.
func (sm *SessionManager) UpdateSession(ctx context.Context, session *SessionState) error {
	session.LastUsedAt = time.Now().UTC()
	key := sm.sessionKey(session.ID)
	data, err := json.Marshal(session)
	if err != nil {
		return fmt.Errorf("failed to marshal session: %w", err)
	}

	// Reset TTL on update
	return sm.redis.Set(ctx, key, data, sm.sessionTTL).Err()
}

// CloseSession marks a session as closed.
func (sm *SessionManager) CloseSession(ctx context.Context, sessionID uuid.UUID) error {
	session, err := sm.GetSession(ctx, sessionID)
	if err != nil {
		return err
	}

	now := time.Now().UTC()
	session.ClosedAt = &now
	session.Status = SessionStatusClosed

	return sm.UpdateSession(ctx, session)
}

// CrashSession marks a session as crashed.
func (sm *SessionManager) CrashSession(ctx context.Context, sessionID uuid.UUID) error {
	session, err := sm.GetSession(ctx, sessionID)
	if err != nil {
		return err
	}

	session.Status = SessionStatusCrashed
	return sm.UpdateSession(ctx, session)
}

// DeleteSession removes a session from Redis.
func (sm *SessionManager) DeleteSession(ctx context.Context, sessionID uuid.UUID) error {
	session, err := sm.GetSession(ctx, sessionID)
	if err != nil {
		return err
	}

	key := sm.sessionKey(sessionID)
	if err := sm.redis.Del(ctx, key).Err(); err != nil {
		return fmt.Errorf("failed to delete session: %w", err)
	}

	// Clear affinity if shared session
	if session.SessionType == SessionTypeShared {
		affinityKey := sm.affinityKey(session.AgentID)
		sm.redis.Del(ctx, affinityKey)
	}

	return nil
}

// SetSessionURL updates the URL for a session.
func (sm *SessionManager) SetSessionURL(ctx context.Context, sessionID uuid.UUID, url string) error {
	session, err := sm.GetSession(ctx, sessionID)
	if err != nil {
		return err
	}

	session.URL = url
	return sm.UpdateSession(ctx, session)
}

// SetSessionCookies updates the cookies for a session.
func (sm *SessionManager) SetSessionCookies(ctx context.Context, sessionID uuid.UUID, cookies []SessionCookie) error {
	session, err := sm.GetSession(ctx, sessionID)
	if err != nil {
		return err
	}

	session.Cookies = cookies
	return sm.UpdateSession(ctx, session)
}

// sessionKey returns the Redis key for a session.
func (sm *SessionManager) sessionKey(sessionID uuid.UUID) string {
	return fmt.Sprintf("browser:session:%s", sessionID.String())
}

// affinityKey returns the Redis key for agent session affinity.
func (sm *SessionManager) affinityKey(agentID string) string {
	return fmt.Sprintf("browser:affinity:%s", agentID)
}

// LoopDetectionKey returns the Redis key for loop detection.
func (sm *SessionManager) LoopDetectionKey(agentID, sessionID, domain string) string {
	return fmt.Sprintf("browser:loop:%s:%s:%s", agentID, sessionID, domain)
}

// CheckLoopDetection checks if a domain has been accessed too many times.
func (sm *SessionManager) CheckLoopDetection(ctx context.Context, agentID, sessionID, domain string, threshold int) (bool, error) {
	key := sm.LoopDetectionKey(agentID, sessionID, domain)
	count, err := sm.redis.Incr(ctx, key).Result()
	if err != nil {
		return false, fmt.Errorf("failed to increment loop counter: %w", err)
	}

	// Set TTL on first access (session duration)
	if count == 1 {
		sm.redis.Expire(ctx, key, sm.sessionTTL)
	}

	return count > int64(threshold), nil
}

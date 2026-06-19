package auth

import (
	"context"
	"encoding/json"
	"fmt"
	"time"

	"github.com/functionfly/functionfly/internal/storage"
	"github.com/redis/go-redis/v9"
	"github.com/google/uuid"
	"github.com/sirupsen/logrus"
)

const (
	SessionCacheTTL       = 30 * time.Second
	SessionCacheKeyPrefix = "session:"
)

type CachedSessionStatus struct {
	Valid        bool      `json:"valid"`
	ExpiresAt    time.Time `json:"expires_at"`
	LastActivity time.Time `json:"last_activity"`
	Revoked      bool      `json:"revoked"`
	SessionID    uuid.UUID `json:"session_id"`
	UserID       uuid.UUID `json:"user_id"`
	TenantID     uuid.UUID `json:"tenant_id"`
}

type SessionCache struct {
	redisClient *redis.Client
	repo        storage.Repository
	logger      *logrus.Entry
}

func NewSessionCache(redisClient *redis.Client, repo storage.Repository) *SessionCache {
	return &SessionCache{
		redisClient: redisClient,
		repo:        repo,
		logger:      logrus.WithField("cache", "session"),
	}
}

func (s *SessionCache) cacheKey(sessionID uuid.UUID) string {
	return SessionCacheKeyPrefix + sessionID.String()
}

func (s *SessionCache) GetSessionStatus(ctx context.Context, sessionID uuid.UUID) (*CachedSessionStatus, error) {
	if s.redisClient == nil {
		return s.getSessionFromDB(ctx, sessionID)
	}

	key := s.cacheKey(sessionID)
	data, err := s.redisClient.Get(ctx, key).Bytes()
	if err == nil {
		var status CachedSessionStatus
		if json.Unmarshal(data, &status) == nil {
			return &status, nil
		}
	}

	if err == redis.Nil {
		return s.getSessionFromDBWithCache(ctx, sessionID)
	}

	if err != nil {
		s.logger.WithError(err).Warn("Redis error, falling back to DB")
		return s.getSessionFromDB(ctx, sessionID)
	}

	return s.getSessionFromDBWithCache(ctx, sessionID)
}

func (s *SessionCache) getSessionFromDB(ctx context.Context, sessionID uuid.UUID) (*CachedSessionStatus, error) {
	session, err := s.repo.GetSessionByID(ctx, sessionID)
	if err != nil {
		return &CachedSessionStatus{Valid: false}, nil
	}
	if session == nil {
		return &CachedSessionStatus{Valid: false}, nil
	}
	return &CachedSessionStatus{
		Valid:        true,
		ExpiresAt:    session.ExpiresAt,
		LastActivity: session.LastActivity,
		Revoked:      false,
		SessionID:    session.ID,
		UserID:       session.UserID,
	}, nil
}

func (s *SessionCache) getSessionFromDBWithCache(ctx context.Context, sessionID uuid.UUID) (*CachedSessionStatus, error) {
	session, err := s.repo.GetSessionByID(ctx, sessionID)
	if err != nil || session == nil {
		status := &CachedSessionStatus{Valid: false}
		if err == nil && session == nil {
			s.setCache(ctx, sessionID, status)
		}
		return status, nil
	}

	status := &CachedSessionStatus{
		Valid:        true,
		ExpiresAt:    session.ExpiresAt,
		LastActivity: session.LastActivity,
		Revoked:      false,
		SessionID:    session.ID,
		UserID:       session.UserID,
	}

	s.setCache(ctx, sessionID, status)
	return status, nil
}

func (s *SessionCache) setCache(ctx context.Context, sessionID uuid.UUID, status *CachedSessionStatus) {
	if s.redisClient == nil {
		return
	}

	key := s.cacheKey(sessionID)
	data, err := json.Marshal(status)
	if err != nil {
		s.logger.WithError(err).Warn("Failed to marshal session status for cache")
		return
	}

	if err := s.redisClient.Set(ctx, key, data, SessionCacheTTL).Err(); err != nil {
		s.logger.WithError(err).Warn("Failed to set session cache")
	}
}

func (s *SessionCache) SetSessionStatus(ctx context.Context, sessionID uuid.UUID, status *CachedSessionStatus) error {
	s.setCache(ctx, sessionID, status)
	return nil
}

func (s *SessionCache) InvalidateSessionCache(ctx context.Context, sessionID uuid.UUID) error {
	if s.redisClient == nil {
		return nil
	}

	key := s.cacheKey(sessionID)
	if err := s.redisClient.Del(ctx, key).Err(); err != nil {
		return fmt.Errorf("failed to invalidate session cache: %w", err)
	}
	return nil
}

func (s *SessionCache) UpdateSessionActivity(ctx context.Context, sessionID uuid.UUID) error {
	if s.redisClient == nil {
		return nil
	}

	status, err := s.GetSessionStatus(ctx, sessionID)
	if err != nil || !status.Valid {
		return err
	}

	status.LastActivity = time.Now()

	key := s.cacheKey(sessionID)
	data, err := json.Marshal(status)
	if err != nil {
		return fmt.Errorf("failed to marshal session status: %w", err)
	}

	if err := s.redisClient.Set(ctx, key, data, SessionCacheTTL).Err(); err != nil {
		return fmt.Errorf("failed to update session activity in cache: %w", err)
	}

	return nil
}

func (s *SessionCache) CreateSessionInCache(ctx context.Context, session *storage.Session, tenantID uuid.UUID) error {
	status := &CachedSessionStatus{
		Valid:        true,
		ExpiresAt:    session.ExpiresAt,
		LastActivity: session.LastActivity,
		Revoked:      false,
		SessionID:    session.ID,
		UserID:       session.UserID,
		TenantID:     tenantID,
	}
	return s.SetSessionStatus(ctx, session.ID, status)
}

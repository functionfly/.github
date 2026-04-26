package middleware

import (
	"context"
	"encoding/json"
	"net/http"
	"sync"
	"time"

	"github.com/functionfly/functionfly/internal/storage"
	"github.com/google/uuid"
	"github.com/redis/go-redis/v9"
	"github.com/sirupsen/logrus"
)

// ActivityTracker tracks user online status by updating last_active_at
// It batches updates to avoid database overload
type ActivityTracker struct {
	repo       *storage.UserRepository
	updates    map[uuid.UUID]time.Time
	mu         sync.RWMutex
	ticker     *time.Ticker
	stopCh     chan struct{}
	batchSize  int
	flushEvery time.Duration
}

// NewActivityTracker creates a new activity tracker
func NewActivityTracker(repo *storage.UserRepository) *ActivityTracker {
	return &ActivityTracker{
		repo:       repo,
		updates:    make(map[uuid.UUID]time.Time),
		stopCh:     make(chan struct{}),
		batchSize:  100,
		flushEvery: 30 * time.Second, // Flush every 30 seconds
	}
}

// Start begins the background flush routine
func (at *ActivityTracker) Start() {
	at.ticker = time.NewTicker(at.flushEvery)
	go at.flushLoop()
}

// Stop halts the background flush routine
func (at *ActivityTracker) Stop() {
	if at.ticker != nil {
		at.ticker.Stop()
	}
	close(at.stopCh)
	// Final flush
	at.Flush()
}

// Track marks a user as active
func (at *ActivityTracker) Track(userID uuid.UUID) {
	at.mu.Lock()
	defer at.mu.Unlock()

	// Only update if it's been at least 1 minute since last update
	if lastUpdate, exists := at.updates[userID]; !exists || time.Since(lastUpdate) > time.Minute {
		at.updates[userID] = time.Now()
	}
}

// flushLoop periodically flushes activity updates to the database
func (at *ActivityTracker) flushLoop() {
	for {
		select {
		case <-at.ticker.C:
			at.Flush()
		case <-at.stopCh:
			return
		}
	}
}

// Flush writes all pending activity updates to the database
func (at *ActivityTracker) Flush() {
	at.mu.Lock()
	if len(at.updates) == 0 {
		at.mu.Unlock()
		return
	}

	// Copy updates to process
	updates := make(map[uuid.UUID]time.Time, len(at.updates))
	for k, v := range at.updates {
		updates[k] = v
	}
	at.updates = make(map[uuid.UUID]time.Time)
	at.mu.Unlock()

	// Process updates in background
	go func() {
		ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
		defer cancel()

		for userID := range updates {
			if err := at.repo.UpdateUserLastActive(ctx, userID); err != nil {
				logrus.WithError(err).WithField("userID", userID).Warn("Failed to update user last active")
			}
		}
	}()
}

// ActivityTrackingMiddleware creates an HTTP middleware that tracks user activity
func ActivityTrackingMiddleware(tracker *ActivityTracker) func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			// Check if user is authenticated
			claims := GetUserFromContext(r)
			if claims != nil && claims.UserID != uuid.Nil {
				tracker.Track(claims.UserID)
			}

			next.ServeHTTP(w, r)
		})
	}
}

// SimpleActivityTracker updates last_active_at for authenticated requests before the handler runs,
// so profile and other handlers read a fresh timestamp in the same request.
func SimpleActivityTracker(repo *storage.UserRepository) func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			claims := GetUserFromContext(r)
			if claims != nil && claims.UserID != uuid.Nil {
				ctx, cancel := context.WithTimeout(r.Context(), 3*time.Second)
				defer cancel()
				if err := repo.UpdateUserLastActive(ctx, claims.UserID); err != nil {
					logrus.WithError(err).WithField("userID", claims.UserID).Debug("Failed to update last active")
				}
			}
			next.ServeHTTP(w, r)
		})
	}
}

const presenceKeyPrefix = "presence:user:"

type PresenceHeartbeat struct {
	UserID     uuid.UUID `json:"userId"`
	TenantID   uuid.UUID `json:"tenantId"`
	Username   string    `json:"username,omitempty"`
	ActiveAt   time.Time `json:"activeAt"`
	LastActive time.Time `json:"lastActive"`
}

// SimpleActivityTrackerWithRedis is like SimpleActivityTracker but also updates Redis presence.
// This enables real-time presence tracking via WebSocket without database queries.
func SimpleActivityTrackerWithRedis(repo *storage.UserRepository, redisClient *redis.Client) func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			claims := GetUserFromContext(r)
			if claims != nil && claims.UserID != uuid.Nil {
				ctx, cancel := context.WithTimeout(r.Context(), 3*time.Second)
				defer cancel()

				if err := repo.UpdateUserLastActive(ctx, claims.UserID); err != nil {
					logrus.WithError(err).WithField("userID", claims.UserID).Debug("Failed to update last active")
				}

				if redisClient != nil {
					hb := &PresenceHeartbeat{
						UserID:     claims.UserID,
						TenantID:   claims.TenantID,
						Username:   claims.Username,
						ActiveAt:   time.Now(),
						LastActive: time.Now(),
					}
					data, _ := json.Marshal(hb)
					key := presenceKeyPrefix + claims.UserID.String()
					if err := redisClient.Set(ctx, key, data, 5*time.Minute).Err(); err != nil {
						logrus.WithError(err).WithField("userID", claims.UserID).Debug("Failed to update Redis presence")
					}
				}
			}
			next.ServeHTTP(w, r)
		})
	}
}

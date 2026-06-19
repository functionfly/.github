package team_memory

import (
	"context"
	"net/http"
	"sync"
	"sync/atomic"

	"github.com/functionfly/functionfly/internal/api/middleware"
	"github.com/functionfly/functionfly/internal/storage"
	"github.com/google/uuid"
	"github.com/sirupsen/logrus"
	"github.com/functionfly/functionfly/internal/apierror"
)

// Context keys to avoid collisions with other packages
type contextKey string

const (
	contextKeyTenantID contextKey = "team_memory_tenant_id"
	contextKeyUserID   contextKey = "team_memory_user_id"
	contextKeyTeamID   contextKey = "team_memory_team_id"
)

// AgentAPIMiddleware provides automatic team memory injection for agent API calls
type AgentAPIMiddleware struct {
	injector *AgentPromptInjector
	repo     storage.Repository
	enabled  atomic.Bool
}

// NewAgentAPIMiddleware creates middleware for automatic prompt injection
func NewAgentAPIMiddleware(injector *AgentPromptInjector, repo storage.Repository) *AgentAPIMiddleware {
	m := &AgentAPIMiddleware{
		injector: injector,
		repo:     repo,
	}
	m.enabled.Store(true)
	return m
}

// SetEnabled enables or disables the middleware (thread-safe)
func (m *AgentAPIMiddleware) SetEnabled(enabled bool) {
	m.enabled.Store(enabled)
}

// Wrap wraps an HTTP handler with team memory context injection
// This middleware extracts tenant/team info from the authenticated user
// and adds it to the request context for downstream prompt injection
func (m *AgentAPIMiddleware) Wrap(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if !m.enabled.Load() || m.injector == nil {
			next.ServeHTTP(w, r)
			return
		}

		// Extract user from context
		user := middleware.GetUserFromContext(r)
		if user == nil {
			// No authenticated user, pass through
			next.ServeHTTP(w, r)
			return
		}

		// Build context with tenant/team info
		ctx := r.Context()
		ctx = context.WithValue(ctx, contextKeyTenantID, user.TenantID)
		ctx = context.WithValue(ctx, contextKeyUserID, user.UserID)

		// Try to get team ID from various sources
		teamID := m.extractTeamID(r, user.TenantID, user.UserID)

		// Validate team belongs to user's tenant if teamID provided
		if teamID != uuid.Nil && m.repo != nil {
			if valid := m.validateTeamTenant(ctx, teamID, user.TenantID); !valid {
				logrus.WithFields(logrus.Fields{
					"team_id":   teamID,
					"tenant_id": user.TenantID,
				}).Warn("Team does not belong to user tenant, rejecting request")
				apierror.WriteError(w, apierror.NewForbidden("Invalid team for tenant"))
				return
			}
		}

		if teamID != uuid.Nil {
			ctx = context.WithValue(ctx, contextKeyTeamID, teamID)
		}

		// Add injector to context for downstream handlers
		ctx = SetInjectorInContext(ctx, m.injector)

		// Log injection setup
		logrus.WithFields(logrus.Fields{
			"path": r.URL.Path,
		}).Debug("Agent API middleware: injected team context")

		// Serve with enriched context
		next.ServeHTTP(w, r.WithContext(ctx))
	})
}

// validateTeamTenant checks if a team belongs to a specific tenant
func (m *AgentAPIMiddleware) validateTeamTenant(ctx context.Context, teamID, tenantID uuid.UUID) bool {
	if m.repo == nil {
		return true // Skip validation if no repository available
	}

	team, err := m.repo.GetTeamByID(ctx, teamID)
	if err != nil {
		logrus.WithError(err).WithField("team_id", teamID).Debug("Failed to get team for validation")
		return false
	}
	if team == nil {
		return false
	}

	return team.TenantID == tenantID
}

// extractTeamID attempts to extract team ID from request
func (m *AgentAPIMiddleware) extractTeamID(r *http.Request, tenantID, userID uuid.UUID) uuid.UUID {
	// 1. Check query parameter
	if teamIDStr := r.URL.Query().Get("team_id"); teamIDStr != "" {
		if teamID, err := uuid.Parse(teamIDStr); err == nil {
			return teamID
		}
	}

	// 2. Check header
	if teamIDStr := r.Header.Get("X-Team-ID"); teamIDStr != "" {
		if teamID, err := uuid.Parse(teamIDStr); err == nil {
			return teamID
		}
	}

	// 3. Try to get user's default/first team (would need repo access)
	// For now, return nil and let downstream handle
	return uuid.Nil
}

// GenerationMiddleware specifically for function generation endpoints
type GenerationMiddleware struct {
	injector *AgentPromptInjector
	repo     storage.Repository
	enabled  atomic.Bool
}

// NewGenerationMiddleware creates middleware for generation endpoints
func NewGenerationMiddleware(injector *AgentPromptInjector, repo storage.Repository) *GenerationMiddleware {
	m := &GenerationMiddleware{
		injector: injector,
		repo:     repo,
	}
	m.enabled.Store(true)
	return m
}

// SetEnabled enables or disables the middleware (thread-safe)
func (m *GenerationMiddleware) SetEnabled(enabled bool) {
	m.enabled.Store(enabled)
}

// Wrap wraps generation endpoints with team memory injection
// This is more aggressive - it actually modifies the request body to include team context
func (m *GenerationMiddleware) Wrap(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if !m.enabled.Load() || m.injector == nil {
			next.ServeHTTP(w, r)
			return
		}

		// Only process POST/PUT requests with body
		if r.Method != http.MethodPost && r.Method != http.MethodPut {
			next.ServeHTTP(w, r)
			return
		}

		// Extract user
		user := middleware.GetUserFromContext(r)
		if user == nil {
			next.ServeHTTP(w, r)
			return
		}

		// Get team ID - validate it belongs to user's tenant
		ctx := r.Context()
		teamID := m.extractTeamID(r, user.TenantID, user.UserID)
		if teamID == uuid.Nil {
			next.ServeHTTP(w, r)
			return
		}

		// Validate team belongs to user's tenant
		if m.repo != nil {
			team, err := m.repo.GetTeamByID(ctx, teamID)
			if err != nil || team == nil || team.TenantID != user.TenantID {
				logrus.WithFields(logrus.Fields{
					"team_id":   teamID,
					"tenant_id": user.TenantID,
				}).Warn("Team does not belong to user tenant, rejecting request")
				apierror.WriteError(w, apierror.NewForbidden("Invalid team for tenant"))
				return
			}
		}

		// Add context info to request context
		ctx = context.WithValue(ctx, contextKeyTenantID, user.TenantID)
		ctx = context.WithValue(ctx, contextKeyTeamID, teamID)
		ctx = context.WithValue(ctx, contextKeyUserID, user.UserID)
		ctx = SetInjectorInContext(ctx, m.injector)

		logrus.WithFields(logrus.Fields{
			"path": r.URL.Path,
		}).Debug("Generation middleware: set up team context for prompt injection")

		next.ServeHTTP(w, r.WithContext(ctx))
	})
}

// extractTeamID extracts team ID from request
func (m *GenerationMiddleware) extractTeamID(r *http.Request, tenantID, userID uuid.UUID) uuid.UUID {
	// Check query parameter
	if teamIDStr := r.URL.Query().Get("team_id"); teamIDStr != "" {
		if teamID, err := uuid.Parse(teamIDStr); err == nil {
			return teamID
		}
	}

	// Check header
	if teamIDStr := r.Header.Get("X-Team-ID"); teamIDStr != "" {
		if teamID, err := uuid.Parse(teamIDStr); err == nil {
			return teamID
		}
	}

	return uuid.Nil
}

// DefaultAgentMiddleware is the global singleton instance
var (
	defaultAgentMiddleware     *AgentAPIMiddleware
	defaultAgentMiddlewareOnce sync.Once
)

// InitializeDefaultAgentMiddleware initializes the default middleware (idempotent, thread-safe)
func InitializeDefaultAgentMiddleware(injector *AgentPromptInjector, repo storage.Repository) {
	defaultAgentMiddlewareOnce.Do(func() {
		defaultAgentMiddleware = NewAgentAPIMiddleware(injector, repo)
	})
}

// GetDefaultAgentMiddleware returns the default middleware (nil if not initialized)
func GetDefaultAgentMiddleware() *AgentAPIMiddleware {
	return defaultAgentMiddleware
}

// MiddlewareFactory creates agent middleware with configuration
type MiddlewareFactory struct {
	injector      *AgentPromptInjector
	defaultTeamID uuid.UUID
	mu            sync.RWMutex
	enabledPaths  map[string]bool
	disabledPaths map[string]bool
}

// NewMiddlewareFactory creates a new middleware factory
func NewMiddlewareFactory(injector *AgentPromptInjector) *MiddlewareFactory {
	return &MiddlewareFactory{
		injector:      injector,
		enabledPaths:  make(map[string]bool),
		disabledPaths: make(map[string]bool),
	}
}

// SetDefaultTeamID sets the default team ID (thread-safe)
func (f *MiddlewareFactory) SetDefaultTeamID(teamID uuid.UUID) *MiddlewareFactory {
	f.mu.Lock()
	defer f.mu.Unlock()
	f.defaultTeamID = teamID
	return f
}

// EnablePath enables injection for a specific path pattern (thread-safe)
func (f *MiddlewareFactory) EnablePath(path string) *MiddlewareFactory {
	f.mu.Lock()
	defer f.mu.Unlock()
	f.enabledPaths[path] = true
	return f
}

// DisablePath disables injection for a specific path pattern (thread-safe)
func (f *MiddlewareFactory) DisablePath(path string) *MiddlewareFactory {
	f.mu.Lock()
	defer f.mu.Unlock()
	f.disabledPaths[path] = true
	return f
}

// Create generates the middleware with configured settings
func (f *MiddlewareFactory) Create() func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			f.mu.RLock()
			disabledPaths := f.disabledPaths
			enabledPaths := f.enabledPaths
			f.mu.RUnlock()

			// Check if path is explicitly disabled
			if disabledPaths[r.URL.Path] {
				next.ServeHTTP(w, r)
				return
			}

			// If enabled paths specified, only process those
			if len(enabledPaths) > 0 && !enabledPaths[r.URL.Path] {
				next.ServeHTTP(w, r)
				return
			}

			// Extract user
			user := middleware.GetUserFromContext(r)
			if user == nil {
				next.ServeHTTP(w, r)
				return
			}

			// Build context
			ctx := r.Context()
			ctx = context.WithValue(ctx, contextKeyTenantID, user.TenantID)
			ctx = context.WithValue(ctx, contextKeyUserID, user.UserID)

			// Determine team ID
			teamID := f.extractTeamID(r, user.TenantID, user.UserID)
			if teamID == uuid.Nil {
				teamID = f.defaultTeamID
			}
			if teamID != uuid.Nil {
				ctx = context.WithValue(ctx, contextKeyTeamID, teamID)
			}

			// Add injector
			ctx = SetInjectorInContext(ctx, f.injector)

			next.ServeHTTP(w, r.WithContext(ctx))
		})
	}
}

// extractTeamID attempts to extract team ID from request
func (f *MiddlewareFactory) extractTeamID(r *http.Request, tenantID, userID uuid.UUID) uuid.UUID {
	// Query parameter
	if teamIDStr := r.URL.Query().Get("team_id"); teamIDStr != "" {
		if teamID, err := uuid.Parse(teamIDStr); err == nil {
			return teamID
		}
	}

	// Header
	if teamIDStr := r.Header.Get("X-Team-ID"); teamIDStr != "" {
		if teamID, err := uuid.Parse(teamIDStr); err == nil {
			return teamID
		}
	}

	return uuid.Nil
}

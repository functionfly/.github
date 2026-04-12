// Package routes provides modular route registration following the RouteRegistrar pattern.
// This replaces the monolithic setupRoutes() function with domain-specific route registration.
package routes

import (
	"net/http"

	"github.com/gorilla/mux"
)

// MiddlewareChain provides access to all middleware components for route registration
type MiddlewareChain struct {
	Auth interface {
		RequireAuth(http.HandlerFunc) http.HandlerFunc
	}
	OptionalAuth    func(http.Handler) http.Handler
	AuthRateLimiter interface {
		RequireAuthRateLimit(http.HandlerFunc) http.HandlerFunc
	}
	VaultRateLimiter interface {
		RequireVaultRateLimit(http.HandlerFunc) http.HandlerFunc
	}
	FlywheelRateLimiter interface {
		RequireFlywheelRateLimit(http.HandlerFunc) http.HandlerFunc
	}
	ProviderRateLimiter interface {
		RequireProviderRateLimit(http.HandlerFunc) http.HandlerFunc
	}
	WalletRateLimiter interface {
		RequireWalletRateLimit(http.HandlerFunc) http.HandlerFunc
	}
	AdminRateLimiter interface {
		RequireAdminRateLimit(http.HandlerFunc) http.HandlerFunc
	}
	AdminSession interface {
		RequireAdminSession(http.HandlerFunc) http.HandlerFunc
	}
	IPAllowlist interface {
		RequireIPAllowlist(http.HandlerFunc) http.HandlerFunc
	}
	CSRF interface {
		RequireCSRF(http.HandlerFunc) http.HandlerFunc
	}
	ExecutionSecurity      http.Handler
	VerificationMiddleware *struct{ Enabled bool }
	Maintenance            func(http.Handler) http.Handler
}

// RouteRegistrar defines the interface for domain-specific route registration.
// Each domain package implements this to register its own routes.
type RouteRegistrar interface {
	// Name returns the domain name for logging/debugging
	Name() string
	// Priority returns the registration priority (lower = earlier)
	Priority() int
	// Register routes on the provided router
	Register(router *mux.Router, api *mux.Router, protected *mux.Router, mw *MiddlewareChain)
}

// Registry holds all route registrars and executes registration
type Registry struct {
	registrars []RouteRegistrar
}

// NewRegistry creates a new route registry
func NewRegistry() *Registry {
	return &Registry{
		registrars: make([]RouteRegistrar, 0),
	}
}

// Add registers a route registrar
func (r *Registry) Add(registrar RouteRegistrar) {
	r.registrars = append(r.registrars, registrar)
}

// RegisterAll executes all registrars in priority order
func (r *Registry) RegisterAll(router *mux.Router, api *mux.Router, protected *mux.Router, mw *MiddlewareChain) {
	// Simple bubble sort by priority
	for i := 0; i < len(r.registrars); i++ {
		for j := i + 1; j < len(r.registrars); j++ {
			if r.registrars[j].Priority() < r.registrars[i].Priority() {
				r.registrars[i], r.registrars[j] = r.registrars[j], r.registrars[i]
			}
		}
	}

	for _, reg := range r.registrars {
		reg.Register(router, api, protected, mw)
	}
}

// BaseRegistrar provides common functionality for registrars
type BaseRegistrar struct {
	name     string
	priority int
}

// Name returns the registrar name
func (b *BaseRegistrar) Name() string {
	return b.name
}

// Priority returns the registration priority
func (b *BaseRegistrar) Priority() int {
	return b.priority
}

// SetPriority allows adjusting priority
func (b *BaseRegistrar) SetPriority(p int) {
	b.priority = p
}

// Priority constants for consistent ordering
const (
	PrioritySystem   = 0  // Health, metrics, well-known
	PriorityAuth     = 10 // Authentication (must be early)
	PriorityPublic   = 20 // Public APIs
	PriorityCore     = 30 // Core business logic
	PriorityExtended = 40 // Extended features
	PriorityInternal = 50 // Internal/admin APIs
	PriorityWebhooks = 60 // Webhooks (usually last)
)

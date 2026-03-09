package state

import (
	"context"

	"github.com/google/uuid"

	"github.com/functionfly/functionfly/internal/storage"
	staterepo "github.com/functionfly/functionfly/internal/storage/state"
)

// UserTenantResolver resolves the tenant ID for a user (for tenant-scoped access checks).
type UserTenantResolver interface {
	GetUserTenantID(ctx context.Context, userID uuid.UUID) (uuid.UUID, error)
}

// Handler handles state API requests
type Handler struct {
	stateRepo      *staterepo.StateRepository
	triggerEngine  *staterepo.TriggerEngine
	userTenant     UserTenantResolver
}

// NewHandler creates a new state handler
func NewHandler(stateRepo *staterepo.StateRepository) *Handler {
	return &Handler{
		stateRepo: stateRepo,
	}
}

// NewHandlerWithTriggerEngine creates a new state handler with trigger engine
func NewHandlerWithTriggerEngine(stateRepo *staterepo.StateRepository, triggerEngine *staterepo.TriggerEngine) *Handler {
	return &Handler{
		stateRepo:     stateRepo,
		triggerEngine: triggerEngine,
	}
}

// WithUserTenantResolver sets the resolver used to verify user belongs to a tenant. Required for production tenant checks.
func (h *Handler) WithUserTenantResolver(resolver UserTenantResolver) *Handler {
	h.userTenant = resolver
	return h
}

// RepoUserTenantResolver adapts storage.Repository to UserTenantResolver.
func RepoUserTenantResolver(repo storage.Repository) UserTenantResolver {
	return &repoUserTenantResolver{repo: repo}
}

type repoUserTenantResolver struct {
	repo storage.Repository
}

func (r *repoUserTenantResolver) GetUserTenantID(ctx context.Context, userID uuid.UUID) (uuid.UUID, error) {
	user, err := r.repo.GetUserByID(userID)
	if err != nil || user == nil {
		return uuid.Nil, err
	}
	return user.TenantID, nil
}

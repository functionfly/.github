package team_memory

import (
	"context"
	"fmt"
	"strings"

	"github.com/functionfly/functionfly/internal/agent/generation"
	"github.com/functionfly/functionfly/internal/storage"
	"github.com/google/uuid"
	"github.com/sirupsen/logrus"
)

// AgentPromptInjector automatically injects team memory context into agent prompts
type AgentPromptInjector struct {
	contextProvider *AgentContextProvider
	repo            storage.Repository
	enabled         bool
	defaultTeamID   uuid.UUID // Used when team context not explicitly provided
}

// NewAgentPromptInjector creates a new prompt injector for team memory
func NewAgentPromptInjector(repo storage.Repository) *AgentPromptInjector {
	return &AgentPromptInjector{
		contextProvider: NewAgentContextProvider(repo),
		repo:            repo,
		enabled:         true,
	}
}

// SetEnabled enables or disables prompt injection
func (i *AgentPromptInjector) SetEnabled(enabled bool) {
	i.enabled = enabled
}

// SetDefaultTeamID sets a fallback team ID when not in request context
func (i *AgentPromptInjector) SetDefaultTeamID(teamID uuid.UUID) {
	i.defaultTeamID = teamID
}

// InjectContextRequest contains context for injection
type InjectContextRequest struct {
	TenantID    uuid.UUID
	TeamID      uuid.UUID // Optional, extracted from user/tenant if not provided
	UserID      uuid.UUID
	BasePrompt  string
	TaskType    string // e.g., "code_generation", "analysis", "review"
	Categories  []string
	MemoryTypes []string // Filter by memory types
}

// InjectIntoGenerationRequest injects team memory into a code generation request
func (i *AgentPromptInjector) InjectIntoGenerationRequest(
	ctx context.Context,
	req *generation.GenerationRequest,
	injectReq InjectContextRequest,
) error {
	if !i.enabled || i.contextProvider == nil {
		return nil
	}

	// Determine team context
	teamID := injectReq.TeamID
	if teamID == uuid.Nil {
		teamID = i.defaultTeamID
	}
	if teamID == uuid.Nil {
		// Try to get team from tenant (first team)
		teamID = i.getFirstTeamForTenant(ctx, injectReq.TenantID)
	}
	if teamID == uuid.Nil {
		logrus.Debug("No team context available for prompt injection")
		return nil
	}

	// Build context request
	contextReq := ContextRequest{
		TenantID:           injectReq.TenantID,
		TeamID:             teamID,
		CurrentTask:        fmt.Sprintf("Generate %s: %s", req.Name, req.Description),
		RelevantCategories: injectReq.Categories,
		MemoryTypes:        injectReq.MemoryTypes,
		MaxMemories:        5, // Limit for generation context
		IncludeUnvalidated: false,
	}

	// Get team context
	contextResp, err := i.contextProvider.BuildContext(ctx, contextReq)
	if err != nil {
		logrus.WithError(err).Warn("Failed to build team context for prompt injection")
		return nil // Don't fail the request, just skip injection
	}

	if contextResp.Context == "" {
		return nil // No context to inject
	}

	// Inject into the prompt
	enhancedPrompt := i.enhancePromptWithContext(req.Prompt, contextResp.Context)
	req.Prompt = enhancedPrompt

	logrus.WithFields(logrus.Fields{
		"team_id":           teamID,
		"function":          req.Name,
		"context_len":       len(contextResp.Context),
		"memories_included": len(contextResp.Sources),
	}).Debug("Injected team memory into generation prompt")

	return nil
}

// InjectIntoPrompt injects team memory into any prompt string
func (i *AgentPromptInjector) InjectIntoPrompt(
	ctx context.Context,
	basePrompt string,
	injectReq InjectContextRequest,
) (string, error) {
	if !i.enabled || i.contextProvider == nil {
		return basePrompt, nil
	}

	// Determine team context
	teamID := injectReq.TeamID
	if teamID == uuid.Nil {
		teamID = i.defaultTeamID
	}
	if teamID == uuid.Nil {
		return basePrompt, nil
	}

	// Build context request
	contextReq := ContextRequest{
		TenantID:           injectReq.TenantID,
		TeamID:             teamID,
		CurrentTask:        injectReq.BasePrompt,
		RelevantCategories: injectReq.Categories,
		MemoryTypes:        injectReq.MemoryTypes,
		MaxMemories:        10,
		IncludeUnvalidated: false,
	}

	// Get team context
	contextResp, err := i.contextProvider.BuildContext(ctx, contextReq)
	if err != nil {
		logrus.WithError(err).Warn("Failed to build team context for prompt injection")
		return basePrompt, nil
	}

	if contextResp.Context == "" {
		return basePrompt, nil
	}

	// Inject into the prompt
	enhancedPrompt := i.enhancePromptWithContext(basePrompt, contextResp.Context)

	logrus.WithFields(logrus.Fields{
		"team_id":     teamID,
		"context_len": len(contextResp.Context),
	}).Debug("Injected team memory into prompt")

	return enhancedPrompt, nil
}

// enhancePromptWithContext adds team context to a prompt
func (i *AgentPromptInjector) enhancePromptWithContext(basePrompt, teamContext string) string {
	var parts []string

	// Add team context header
	parts = append(parts, "## Team Knowledge & Context")
	parts = append(parts, "The following information represents shared team knowledge that should inform your response:")
	parts = append(parts, "")
	parts = append(parts, teamContext)
	parts = append(parts, "")
	parts = append(parts, "---")
	parts = append(parts, "")
	parts = append(parts, "## Original Request")
	parts = append(parts, basePrompt)

	return strings.Join(parts, "\n")
}

// getFirstTeamForTenant retrieves the first available team for a tenant
func (i *AgentPromptInjector) getFirstTeamForTenant(ctx context.Context, tenantID uuid.UUID) uuid.UUID {
	if tenantID == uuid.Nil {
		return uuid.Nil
	}

	if i.repo == nil {
		return uuid.Nil
	}

	teams, err := i.repo.GetTeamsByTenantID(tenantID)
	if err != nil {
		logrus.WithError(err).WithField("tenant_id", tenantID).Warn("Failed to get teams for tenant")
		return uuid.Nil
	}

	if len(teams) == 0 {
		return uuid.Nil
	}

	// Return the first team's ID
	return teams[0].ID
}

// GenerationServiceWrapper wraps a generation service to inject team memory
type GenerationServiceWrapper struct {
	inner    generation.CodeGenerator
	injector *AgentPromptInjector
}

// NewGenerationServiceWrapper creates a wrapper that injects team memory
func NewGenerationServiceWrapper(
	inner generation.CodeGenerator,
	injector *AgentPromptInjector,
) *GenerationServiceWrapper {
	return &GenerationServiceWrapper{
		inner:    inner,
		injector: injector,
	}
}

// GenerateCode implements generation.CodeGenerator with team memory injection
func (w *GenerationServiceWrapper) GenerateCode(
	ctx context.Context,
	req *generation.GenerationRequest,
) (string, error) {
	// Try to extract tenant/team from context
	tenantID, teamID, userID := extractContextInfo(ctx)

	// Inject team memory if we have context
	if tenantID != uuid.Nil || teamID != uuid.Nil {
		injectReq := InjectContextRequest{
			TenantID:    tenantID,
			TeamID:      teamID,
			UserID:      userID,
			BasePrompt:  req.Prompt,
			TaskType:    "code_generation",
			MemoryTypes: []string{"decision", "preference", "process"}, // Most relevant for code gen
		}

		if err := w.injector.InjectIntoGenerationRequest(ctx, req, injectReq); err != nil {
			logrus.WithError(err).Warn("Failed to inject team memory, proceeding with original prompt")
		}
	}

	// Call the inner generator
	return w.inner.GenerateCode(ctx, req)
}

// extractContextInfo extracts tenant/team/user IDs from context
// This should be populated by middleware that sets these values in context
func extractContextInfo(ctx context.Context) (tenantID, teamID, userID uuid.UUID) {
	return GetTenantIDFromContext(ctx), GetTeamIDFromContext(ctx), GetUserIDFromContext(ctx)
}

// ContextMiddleware adds team context to the request context
type ContextMiddleware struct {
	tenantID uuid.UUID
	teamID   uuid.UUID
	userID   uuid.UUID
}

// NewContextMiddleware creates middleware that adds team context
func NewContextMiddleware(tenantID, teamID, userID uuid.UUID) func(context.Context) context.Context {
	return func(ctx context.Context) context.Context {
		ctx = context.WithValue(ctx, contextKeyTenantID, tenantID)
		ctx = context.WithValue(ctx, contextKeyTeamID, teamID)
		ctx = context.WithValue(ctx, contextKeyUserID, userID)
		return ctx
	}
}

// Context key constants (defined in agent_middleware.go)
// Exported helpers for extracting context values

// GetTenantIDFromContext extracts tenant ID from context
func GetTenantIDFromContext(ctx context.Context) uuid.UUID {
	if v := ctx.Value(contextKeyTenantID); v != nil {
		if id, ok := v.(uuid.UUID); ok {
			return id
		}
	}
	return uuid.Nil
}

// GetTeamIDFromContext extracts team ID from context
func GetTeamIDFromContext(ctx context.Context) uuid.UUID {
	if v := ctx.Value(contextKeyTeamID); v != nil {
		if id, ok := v.(uuid.UUID); ok {
			return id
		}
	}
	return uuid.Nil
}

// GetUserIDFromContext extracts user ID from context
func GetUserIDFromContext(ctx context.Context) uuid.UUID {
	if v := ctx.Value(contextKeyUserID); v != nil {
		if id, ok := v.(uuid.UUID); ok {
			return id
		}
	}
	return uuid.Nil
}

// GetInjectorFromContext retrieves the injector from context
func GetInjectorFromContext(ctx context.Context) *AgentPromptInjector {
	if v := ctx.Value("team_memory_injector"); v != nil {
		if injector, ok := v.(*AgentPromptInjector); ok {
			return injector
		}
	}
	return nil
}

// SetInjectorInContext sets the injector in context
func SetInjectorInContext(ctx context.Context, injector *AgentPromptInjector) context.Context {
	return context.WithValue(ctx, "team_memory_injector", injector)
}

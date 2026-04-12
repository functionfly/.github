package team_memory

import (
	"context"
	"fmt"
	"strings"
	"time"

	"github.com/functionfly/functionfly/internal/monitoring"
	"github.com/functionfly/functionfly/internal/storage"
	"github.com/google/uuid"
	"github.com/sirupsen/logrus"
)

// AgentContextProvider injects team memory into agent context
type AgentContextProvider struct {
	repo storage.Repository
}

// NewAgentContextProvider creates a new context provider
func NewAgentContextProvider(repo storage.Repository) *AgentContextProvider {
	return &AgentContextProvider{repo: repo}
}

// ContextRequest represents a request for team context
type ContextRequest struct {
	TenantID           uuid.UUID
	TeamID             uuid.UUID
	CurrentTask        string   // What the agent is currently working on
	RelevantCategories []string // e.g., ["client:acme", "process:onboarding"]
	MemoryTypes        []string // Filter by types: decision, preference, process, client_context
	MaxMemories        int      // Maximum memories to include
	IncludeUnvalidated bool     // Whether to include unvalidated memories
}

// ContextResponse contains formatted context for LLM consumption
type ContextResponse struct {
	Context string                `json:"context"` // Formatted text for LLM
	Sources []*storage.TeamMemory `json:"sources"` // Source memories
}

// BuildContext generates formatted context string for LLM/agent consumption
func (p *AgentContextProvider) BuildContext(ctx context.Context, req ContextRequest) (*ContextResponse, error) {
	start := time.Now()
	teamIDStr := req.TeamID.String()

	if req.MaxMemories == 0 {
		req.MaxMemories = 10
	}
	if req.MaxMemories > 20 {
		req.MaxMemories = 20
	}

	// Search for relevant memories using the current task as query
	var results []*storage.TeamMemorySearchResult
	var err error
	searchType := "vector"

	if req.CurrentTask != "" {
		results, err = p.repo.SearchTeamMemories(ctx, req.TenantID, req.TeamID, req.CurrentTask, req.MaxMemories)
	} else {
		// If no task query, get recent validated memories
		validated := true
		filter := storage.TeamMemoryFilter{
			IsValidated: &validated,
			Limit:       req.MaxMemories,
		}
		memories, _, err := p.repo.ListTeamMemories(ctx, req.TenantID, req.TeamID, filter)
		if err != nil {
			return nil, fmt.Errorf("failed to list memories: %w", err)
		}
		// Convert to search results
		for _, m := range memories {
			results = append(results, &storage.TeamMemorySearchResult{
				TeamMemory:     *m,
				RelevanceScore: 1.0,
			})
		}
	}

	if err != nil {
		logrus.WithError(err).Error("Failed to search team memories for context")
		return nil, fmt.Errorf("failed to search memories: %w", err)
	}

	// Filter by memory types if specified
	if len(req.MemoryTypes) > 0 {
		filtered := make([]*storage.TeamMemorySearchResult, 0, len(results))
		for _, r := range results {
			for _, t := range req.MemoryTypes {
				if r.MemoryType == t {
					filtered = append(filtered, r)
					break
				}
			}
		}
		results = filtered
	}

	// Filter by categories if specified
	if len(req.RelevantCategories) > 0 {
		filtered := make([]*storage.TeamMemorySearchResult, 0, len(results))
		for _, r := range results {
			if r.Category != nil {
				for _, cat := range req.RelevantCategories {
					if strings.Contains(*r.Category, cat) {
						filtered = append(filtered, r)
						break
					}
				}
			}
		}
		results = filtered
	}

	// Filter out unvalidated unless requested
	if !req.IncludeUnvalidated {
		filtered := make([]*storage.TeamMemorySearchResult, 0, len(results))
		for _, r := range results {
			if r.IsValidated {
				filtered = append(filtered, r)
			}
		}
		results = filtered
	}

	// Record search duration
	monitoring.RecordTeamMemorySearchDuration(teamIDStr, searchType, time.Since(start))

	// Build formatted context
	contextText := p.formatContextForAgent(results, req.CurrentTask)

	// Record context injection metric
	monitoring.RecordTeamMemoryContextInjection(teamIDStr)

	// Extract sources
	sources := make([]*storage.TeamMemory, 0, len(results))
	for _, result := range results {
		// Mark as accessed (async)
		go func(memoryID uuid.UUID) {
			if err := p.repo.MarkTeamMemoryAsAccessed(context.Background(), memoryID); err != nil {
				logrus.WithError(err).Debug("Failed to mark memory as accessed")
			}
		}(result.ID)

		memory := result.TeamMemory
		sources = append(sources, &memory)
	}

	return &ContextResponse{
		Context: contextText,
		Sources: sources,
	}, nil
}

// formatContextForAgent formats memories into LLM-readable context
func (p *AgentContextProvider) formatContextForAgent(results []*storage.TeamMemorySearchResult, query string) string {
	if len(results) == 0 {
		return ""
	}

	var parts []string
	parts = append(parts, "## Team Knowledge & Context")
	if query != "" {
		parts = append(parts, fmt.Sprintf("(Relevant to: %s)", query))
	}
	parts = append(parts, "")

	// Group by type for better organization
	grouped := make(map[string][]*storage.TeamMemorySearchResult)
	typeOrder := []string{"client_context", "decision", "preference", "process"}

	for _, result := range results {
		grouped[result.MemoryType] = append(grouped[result.MemoryType], result)
	}

	// Output in preferred order
	for _, memoryType := range typeOrder {
		memories, ok := grouped[memoryType]
		if !ok {
			continue
		}

		// Type header
		typeLabel := formatTypeLabel(memoryType)
		parts = append(parts, fmt.Sprintf("### %s", typeLabel))
		parts = append(parts, "")

		for i, m := range memories {
			validity := ""
			if !m.IsValidated {
				validity = " [UNVALIDATED - use with caution]"
			}

			category := ""
			if m.Category != nil && *m.Category != "" {
				category = fmt.Sprintf(" [%s]", *m.Category)
			}

			relevance := ""
			if m.RelevanceScore > 0 {
				relevance = fmt.Sprintf(" (%.0f%% match)", m.RelevanceScore*100)
			}

			content := p.formatMemoryContentForAgent(m.Content, m.MemoryType)

			parts = append(parts, fmt.Sprintf("%d. **%s**%s%s%s",
				i+1,
				*m.Summary,
				category,
				validity,
				relevance,
			))

			if content != "" {
				parts = append(parts, fmt.Sprintf("   %s", content))
			}

			parts = append(parts, fmt.Sprintf("   Confidence: %.0f%% | Last updated: %s",
				m.ConfidenceScore*100,
				m.UpdatedAt.Format("2006-01-02"),
			))
			parts = append(parts, "")
		}
	}

	return strings.Join(parts, "\n")
}

// formatMemoryContentForAgent extracts key details from structured content
func (p *AgentContextProvider) formatMemoryContentForAgent(content map[string]interface{}, memoryType string) string {
	if content == nil {
		return ""
	}

	switch memoryType {
	case "decision":
		rationale, _ := content["rationale"].(string)
		if rationale != "" {
			return fmt.Sprintf("Rationale: %s", truncate(rationale, 150))
		}

	case "preference":
		subject, _ := content["subject"].(string)
		value, _ := content["value"].(string)
		context, _ := content["context"].(string)

		if subject != "" && value != "" {
			if context != "" {
				return fmt.Sprintf("%s: %s (%s)", subject, value, context)
			}
			return fmt.Sprintf("%s: %s", subject, value)
		}

	case "process":
		name, _ := content["name"].(string)
		steps, hasSteps := content["steps"].([]interface{})

		if name != "" {
			if hasSteps && len(steps) > 0 {
				stepList := make([]string, 0, len(steps))
				for i, step := range steps {
					if i >= 3 {
						stepList = append(stepList, "...")
						break
					}
					if s, ok := step.(string); ok {
						stepList = append(stepList, s)
					}
				}
				return fmt.Sprintf("Steps: %s", strings.Join(stepList, " → "))
			}
			return fmt.Sprintf("Process: %s", name)
		}

	case "client_context":
		clientName, _ := content["client_name"].(string)
		notes, _ := content["notes"].(string)

		if notes != "" {
			return truncate(notes, 200)
		}
		if clientName != "" {
			return fmt.Sprintf("Client: %s", clientName)
		}
	}

	// Fallback: return key-value pairs
	var pairs []string
	for k, v := range content {
		if k == "notes" || k == "rationale" || k == "description" {
			if s, ok := v.(string); ok && s != "" {
				return truncate(s, 150) // Return first good text field found
			}
		}
		pairs = append(pairs, fmt.Sprintf("%s: %v", k, v))
		if len(pairs) >= 3 {
			break
		}
	}

	return strings.Join(pairs, ", ")
}

// formatTypeLabel returns human-readable label for memory type
func formatTypeLabel(memoryType string) string {
	switch memoryType {
	case "decision":
		return "📋 Team Decisions"
	case "preference":
		return "⭐ Preferences"
	case "process":
		return "⚙️ Processes"
	case "client_context":
		return "👥 Client Context"
	default:
		return memoryType
	}
}

// truncate truncates string to max length
func truncate(s string, maxLen int) string {
	if len(s) <= maxLen {
		return s
	}
	return s[:maxLen] + "..."
}

// InjectIntoPrompt injects team context into an agent prompt
func (p *AgentContextProvider) InjectIntoPrompt(ctx context.Context, basePrompt string, req ContextRequest) (string, error) {
	contextResp, err := p.BuildContext(ctx, req)
	if err != nil {
		return basePrompt, err // Return original prompt on error
	}

	if contextResp.Context == "" {
		return basePrompt, nil // No context to inject
	}

	// Combine with base prompt
	var parts []string
	parts = append(parts, contextResp.Context)
	parts = append(parts, "---")
	parts = append(parts, basePrompt)

	return strings.Join(parts, "\n\n"), nil
}

// GetMemorySummary returns a quick summary of team memories (for agent boot)
func (p *AgentContextProvider) GetMemorySummary(ctx context.Context, tenantID, teamID uuid.UUID) (string, error) {
	// Get count by type
	types := []string{"decision", "preference", "process", "client_context"}
	summary := make(map[string]int)

	for _, t := range types {
		memories, err := p.repo.ListTeamMemoriesByType(ctx, tenantID, teamID, t, 100, 0)
		if err != nil {
			continue
		}
		summary[t] = len(memories)
	}

	var parts []string
	parts = append(parts, "## Team Memory Summary")
	parts = append(parts, "")
	parts = append(parts, fmt.Sprintf("📋 Decisions: %d", summary["decision"]))
	parts = append(parts, fmt.Sprintf("⭐ Preferences: %d", summary["preference"]))
	parts = append(parts, fmt.Sprintf("⚙️ Processes: %d", summary["process"]))
	parts = append(parts, fmt.Sprintf("👥 Client Context: %d", summary["client_context"]))

	return strings.Join(parts, "\n"), nil
}

package team_memory

import (
	"context"
	"fmt"
	"time"

	"github.com/functionfly/functionfly/internal/monitoring"
	"github.com/functionfly/functionfly/internal/storage"
	"github.com/google/uuid"
	"github.com/sirupsen/logrus"
)

// ============================================
// Memory Templates for Common Patterns
// ============================================

// MemoryTemplate provides a pre-defined template for creating memories
type MemoryTemplate struct {
	ID             uuid.UUID              `json:"id"`
	Name           string                 `json:"name"`
	Description    string                 `json:"description"`
	MemoryType     string                 `json:"memory_type"` // 'decision', 'preference', 'process', 'client_context'
	Category       string                 `json:"category"`
	ContentSchema  map[string]interface{} `json:"content_schema"`  // JSON schema for content validation
	DefaultContent map[string]interface{} `json:"default_content"` // Default values to pre-fill
	SampleSummary  string                 `json:"sample_summary"`  // Example summary format with placeholders
	TTLDays        int                    `json:"ttl_days"`
	Confidence     float64                `json:"default_confidence"`
	IsSystem       bool                   `json:"is_system"` // System templates cannot be deleted
	CreatedAt      time.Time              `json:"created_at"`
	UpdatedAt      time.Time              `json:"updated_at"`
}

// TemplateEngine manages memory templates and applies them
type TemplateEngine struct {
	templates map[string]*MemoryTemplate // Index by ID
	repo      storage.Repository
}

// NewTemplateEngine creates a new template engine with built-in templates
func NewTemplateEngine(repo storage.Repository) *TemplateEngine {
	engine := &TemplateEngine{
		templates: make(map[string]*MemoryTemplate),
		repo:      repo,
	}

	// Initialize with built-in system templates
	engine.initializeSystemTemplates()

	return engine
}

// Built-in system templates
func (e *TemplateEngine) initializeSystemTemplates() {
	systemTemplates := []*MemoryTemplate{
		// Decision templates
		{
			ID:          uuid.MustParse("00000000-0000-0000-0000-000000000001"),
			Name:        "Architecture Decision",
			Description: "Record an architectural decision with alternatives and rationale",
			MemoryType:  "decision",
			Category:    "architecture",
			ContentSchema: map[string]interface{}{
				"type": "object",
				"properties": map[string]interface{}{
					"title":          map[string]interface{}{"type": "string"},
					"rationale":      map[string]interface{}{"type": "string"},
					"alternatives":   map[string]interface{}{"type": "array", "items": map[string]interface{}{"type": "string"}},
					"decision_maker": map[string]interface{}{"type": "string"},
					"date":           map[string]interface{}{"type": "string", "format": "date"},
					"status":         map[string]interface{}{"type": "string", "enum": []string{"active", "superseded", "deprecated"}},
				},
				"required": []string{"title", "rationale"},
			},
			DefaultContent: map[string]interface{}{
				"status": "active",
			},
			SampleSummary: "{{title}} - Decided on {{date}}",
			TTLDays:       730, // 2 years
			Confidence:    0.9,
			IsSystem:      true,
		},
		{
			ID:          uuid.MustParse("00000000-0000-0000-0000-000000000002"),
			Name:        "Tech Stack Decision",
			Description: "Record a technology stack or library choice",
			MemoryType:  "decision",
			Category:    "tech-stack",
			ContentSchema: map[string]interface{}{
				"type": "object",
				"properties": map[string]interface{}{
					"technology":    map[string]interface{}{"type": "string"},
					"version":       map[string]interface{}{"type": "string"},
					"use_case":      map[string]interface{}{"type": "string"},
					"alternatives":  map[string]interface{}{"type": "array", "items": map[string]interface{}{"type": "string"}},
					"rationale":     map[string]interface{}{"type": "string"},
					"decision_date": map[string]interface{}{"type": "string", "format": "date"},
				},
				"required": []string{"technology", "rationale"},
			},
			SampleSummary: "Adopt {{technology}} v{{version}} for {{use_case}}",
			TTLDays:       730,
			Confidence:    0.9,
			IsSystem:      true,
		},

		// Preference templates
		{
			ID:          uuid.MustParse("00000000-0000-0000-0000-000000000003"),
			Name:        "Code Style Preference",
			Description: "Team coding style preferences and conventions",
			MemoryType:  "preference",
			Category:    "code-style",
			ContentSchema: map[string]interface{}{
				"type": "object",
				"properties": map[string]interface{}{
					"language":   map[string]interface{}{"type": "string"},
					"convention": map[string]interface{}{"type": "string"},
					"value":      map[string]interface{}{"type": "string"},
					"context":    map[string]interface{}{"type": "string"},
					"priority":   map[string]interface{}{"type": "integer", "minimum": 1, "maximum": 10},
				},
				"required": []string{"language", "convention", "value"},
			},
			SampleSummary: "{{language}}: Prefer {{convention}} - {{value}}",
			TTLDays:       365,
			Confidence:    0.85,
			IsSystem:      true,
		},
		{
			ID:          uuid.MustParse("00000000-0000-0000-0000-000000000004"),
			Name:        "Deployment Preference",
			Description: "Team preferences for deployment and release practices",
			MemoryType:  "preference",
			Category:    "deployment",
			ContentSchema: map[string]interface{}{
				"type": "object",
				"properties": map[string]interface{}{
					"practice":   map[string]interface{}{"type": "string"},
					"preference": map[string]interface{}{"type": "string"},
					"context":    map[string]interface{}{"type": "string"},
					"priority":   map[string]interface{}{"type": "integer", "minimum": 1, "maximum": 10},
				},
				"required": []string{"practice", "preference"},
			},
			SampleSummary: "Deployment: {{practice}} - {{preference}}",
			TTLDays:       365,
			Confidence:    0.85,
			IsSystem:      true,
		},

		// Process templates
		{
			ID:          uuid.MustParse("00000000-0000-0000-0000-000000000005"),
			Name:        "Onboarding Process",
			Description: "New team member or project onboarding steps",
			MemoryType:  "process",
			Category:    "onboarding",
			ContentSchema: map[string]interface{}{
				"type": "object",
				"properties": map[string]interface{}{
					"name":      map[string]interface{}{"type": "string"},
					"steps":     map[string]interface{}{"type": "array", "items": map[string]interface{}{"type": "string"}},
					"owner":     map[string]interface{}{"type": "string"},
					"tools":     map[string]interface{}{"type": "array", "items": map[string]interface{}{"type": "string"}},
					"frequency": map[string]interface{}{"type": "string", "enum": []string{"once", "daily", "weekly", "monthly"}},
				},
				"required": []string{"name", "steps"},
			},
			SampleSummary: "{{name}} Onboarding Process ({{len(steps)}} steps)",
			TTLDays:       0, // Never expire
			Confidence:    0.9,
			IsSystem:      true,
		},
		{
			ID:          uuid.MustParse("00000000-0000-0000-0000-000000000006"),
			Name:        "Incident Response Process",
			Description: "Steps for handling production incidents",
			MemoryType:  "process",
			Category:    "incident-response",
			ContentSchema: map[string]interface{}{
				"type": "object",
				"properties": map[string]interface{}{
					"severity":        map[string]interface{}{"type": "string", "enum": []string{"critical", "high", "medium", "low"}},
					"steps":           map[string]interface{}{"type": "array", "items": map[string]interface{}{"type": "string"}},
					"owner":           map[string]interface{}{"type": "string"},
					"tools":           map[string]interface{}{"type": "array", "items": map[string]interface{}{"type": "string"}},
					"escalation_path": map[string]interface{}{"type": "array", "items": map[string]interface{}{"type": "string"}},
				},
				"required": []string{"severity", "steps"},
			},
			SampleSummary: "Incident Response ({{severity}}): {{len(steps)}} steps",
			TTLDays:       0,
			Confidence:    0.95,
			IsSystem:      true,
		},
		{
			ID:          uuid.MustParse("00000000-0000-0000-0000-000000000007"),
			Name:        "Code Review Process",
			Description: "Team code review workflow and requirements",
			MemoryType:  "process",
			Category:    "code-review",
			ContentSchema: map[string]interface{}{
				"type": "object",
				"properties": map[string]interface{}{
					"name":               map[string]interface{}{"type": "string"},
					"steps":              map[string]interface{}{"type": "array", "items": map[string]interface{}{"type": "string"}},
					"owner":              map[string]interface{}{"type": "string"},
					"required_reviewers": map[string]interface{}{"type": "integer", "minimum": 1},
					"automation_rules":   map[string]interface{}{"type": "array", "items": map[string]interface{}{"type": "string"}},
				},
				"required": []string{"name", "steps"},
			},
			SampleSummary: "{{name}} Code Review ({{required_reviewers}} reviewers required)",
			TTLDays:       0,
			Confidence:    0.9,
			IsSystem:      true,
		},

		// Client Context templates
		{
			ID:          uuid.MustParse("00000000-0000-0000-0000-000000000008"),
			Name:        "Client Profile",
			Description: "Client information and preferences",
			MemoryType:  "client_context",
			Category:    "client-profile",
			ContentSchema: map[string]interface{}{
				"type": "object",
				"properties": map[string]interface{}{
					"client_id":   map[string]interface{}{"type": "string"},
					"client_name": map[string]interface{}{"type": "string"},
					"industry":    map[string]interface{}{"type": "string"},
					"key_contacts": map[string]interface{}{
						"type": "array",
						"items": map[string]interface{}{
							"type": "object",
							"properties": map[string]interface{}{
								"name":  map[string]interface{}{"type": "string"},
								"role":  map[string]interface{}{"type": "string"},
								"email": map[string]interface{}{"type": "string"},
							},
						},
					},
					"preferences": map[string]interface{}{"type": "object"},
					"notes":       map[string]interface{}{"type": "string"},
				},
				"required": []string{"client_id", "client_name"},
			},
			SampleSummary: "Client: {{client_name}} ({{industry}})",
			TTLDays:       0, // Never expire
			Confidence:    0.9,
			IsSystem:      true,
		},
	}

	for _, template := range systemTemplates {
		if template.CreatedAt.IsZero() {
			template.CreatedAt = time.Now()
			template.UpdatedAt = time.Now()
		}
		e.templates[template.ID.String()] = template
	}
}

// GetTemplate retrieves a template by ID
func (e *TemplateEngine) GetTemplate(templateID uuid.UUID) (*MemoryTemplate, error) {
	template, ok := e.templates[templateID.String()]
	if !ok {
		return nil, fmt.Errorf("template not found: %s", templateID)
	}
	return template, nil
}

// ListTemplates lists all available templates
func (e *TemplateEngine) ListTemplates(includeSystem bool) []*MemoryTemplate {
	var templates []*MemoryTemplate
	for _, t := range e.templates {
		if t.IsSystem && !includeSystem {
			continue
		}
		templates = append(templates, t)
	}
	return templates
}

// ListTemplatesByType lists templates filtered by memory type
func (e *TemplateEngine) ListTemplatesByType(memoryType string) []*MemoryTemplate {
	var templates []*MemoryTemplate
	for _, t := range e.templates {
		if t.MemoryType == memoryType {
			templates = append(templates, t)
		}
	}
	return templates
}

// ApplyTemplate creates a new memory using a template
func (e *TemplateEngine) ApplyTemplate(
	ctx context.Context,
	templateID uuid.UUID,
	tenantID, teamID, userID uuid.UUID,
	content map[string]interface{},
	placeholders map[string]string,
) (*storage.TeamMemory, error) {
	start := time.Now()

	template, err := e.GetTemplate(templateID)
	if err != nil {
		return nil, err
	}

	// Merge default content with provided content
	mergedContent := make(map[string]interface{})
	for k, v := range template.DefaultContent {
		mergedContent[k] = v
	}
	for k, v := range content {
		mergedContent[k] = v
	}

	// Generate summary from template
	summary := e.renderTemplateString(template.SampleSummary, mergedContent, placeholders)

	// Create the memory
	memory := &storage.TeamMemory{
		TenantID:        tenantID,
		TeamID:          teamID,
		MemoryType:      template.MemoryType,
		Category:        &template.Category,
		Content:         mergedContent,
		Summary:         &summary,
		CreatedBy:       userID,
		ConfidenceScore: template.Confidence,
		IsValidated:     true, // Template-based memories are pre-validated
		TTLDays:         template.TTLDays,
	}

	// Save to repository
	created, err := e.repo.CreateTeamMemory(ctx, memory)
	if err != nil {
		return nil, fmt.Errorf("failed to create memory from template: %w", err)
	}

	monitoring.RecordTeamMemoryCreated(teamID.String(), template.MemoryType, "template")

	logrus.WithFields(logrus.Fields{
		"template_id":   templateID,
		"template_name": template.Name,
		"memory_id":     created.ID,
		"team_id":       teamID,
		"duration_ms":   time.Since(start).Milliseconds(),
	}).Info("Memory created from template")

	return created, nil
}

// renderTemplateString replaces placeholders in template strings
func (e *TemplateEngine) renderTemplateString(template string, content map[string]interface{}, placeholders map[string]string) string {
	result := template

	// Replace content placeholders
	for key, value := range content {
		placeholder := fmt.Sprintf("{{%s}}", key)
		strValue := fmt.Sprintf("%v", value)
		result = replaceString(result, placeholder, strValue)
	}

	// Replace custom placeholders
	for key, value := range placeholders {
		placeholder := fmt.Sprintf("{{%s}}", key)
		result = replaceString(result, placeholder, value)
	}

	// Handle simple function-like placeholders
	result = e.processFunctionPlaceholders(result, content)

	return result
}

// processFunctionPlaceholders handles simple template functions like {{len(steps)}}
func (e *TemplateEngine) processFunctionPlaceholders(template string, content map[string]interface{}) string {
	// Simple len() function support
	// Look for {{len(field)}}
	result := template
	for key, value := range content {
		lenPlaceholder := fmt.Sprintf("{{len(%s)}}", key)
		if arr, ok := value.([]interface{}); ok {
			result = replaceString(result, lenPlaceholder, fmt.Sprintf("%d", len(arr)))
		}
		if arr, ok := value.([]string); ok {
			result = replaceString(result, lenPlaceholder, fmt.Sprintf("%d", len(arr)))
		}
	}
	return result
}

// AddCustomTemplate adds a custom template (non-system)
func (e *TemplateEngine) AddCustomTemplate(template *MemoryTemplate) error {
	if template.ID == uuid.Nil {
		template.ID = uuid.New()
	}

	if template.IsSystem {
		return fmt.Errorf("cannot create system templates")
	}

	template.IsSystem = false
	template.CreatedAt = time.Now()
	template.UpdatedAt = time.Now()

	e.templates[template.ID.String()] = template

	logrus.WithFields(logrus.Fields{
		"template_id":   template.ID,
		"template_name": template.Name,
		"memory_type":   template.MemoryType,
	}).Info("Custom template added")

	return nil
}

// DeleteTemplate deletes a template (only custom templates can be deleted)
func (e *TemplateEngine) DeleteTemplate(templateID uuid.UUID) error {
	template, ok := e.templates[templateID.String()]
	if !ok {
		return fmt.Errorf("template not found: %s", templateID)
	}

	if template.IsSystem {
		return fmt.Errorf("cannot delete system templates")
	}

	delete(e.templates, templateID.String())

	logrus.WithFields(logrus.Fields{
		"template_id":   templateID,
		"template_name": template.Name,
	}).Info("Template deleted")

	return nil
}

// Helper function for string replacement
func replaceString(s, old, new string) string {
	// Simple string replacement
	result := ""
	start := 0
	for {
		idx := findSubstring(s, old, start)
		if idx == -1 {
			result += s[start:]
			break
		}
		result += s[start:idx] + new
		start = idx + len(old)
	}
	return result
}

func findSubstring(s, substr string, start int) int {
	if start >= len(s) {
		return -1
	}
	for i := start; i <= len(s)-len(substr); i++ {
		if s[i:i+len(substr)] == substr {
			return i
		}
	}
	return -1
}

// TemplateSuggestion provides AI-powered template suggestions based on content
type TemplateSuggestion struct {
	TemplateID   uuid.UUID `json:"template_id"`
	TemplateName string    `json:"template_name"`
	Confidence   float64   `json:"confidence"`
	Reason       string    `json:"reason"`
}

// SuggestTemplates suggests templates based on raw content analysis
func (e *TemplateEngine) SuggestTemplates(content string) []TemplateSuggestion {
	var suggestions []TemplateSuggestion

	// Simple keyword-based suggestion (can be enhanced with NLP)
	contentLower := toLower(content)

	for _, template := range e.templates {
		var score float64
		var reason string

		switch template.MemoryType {
		case "decision":
			if contains(contentLower, "decided", "decision", "chose", "selected", "going with") {
				score = 0.8
				reason = "Contains decision keywords"
			}
		case "preference":
			if contains(contentLower, "prefer", "like to", "always use", "never use", "wants") {
				score = 0.8
				reason = "Contains preference keywords"
			}
		case "process":
			if contains(contentLower, "process", "steps", "workflow", "how to", "first", "then") {
				score = 0.8
				reason = "Contains process/workflow keywords"
			}
		case "client_context":
			if contains(contentLower, "client", "customer", "acme", "corp") {
				score = 0.7
				reason = "Contains client references"
			}
		}

		if score > 0 {
			suggestions = append(suggestions, TemplateSuggestion{
				TemplateID:   template.ID,
				TemplateName: template.Name,
				Confidence:   score,
				Reason:       reason,
			})
		}
	}

	return suggestions
}

func toLower(s string) string {
	// Simple lowercase
	result := make([]byte, len(s))
	for i := 0; i < len(s); i++ {
		c := s[i]
		if c >= 'A' && c <= 'Z' {
			c = c + ('a' - 'A')
		}
		result[i] = c
	}
	return string(result)
}

func contains(s string, substrs ...string) bool {
	for _, substr := range substrs {
		if findSubstring(s, substr, 0) != -1 {
			return true
		}
	}
	return false
}

package deployment

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net/http"
	"strings"
	"time"

	"github.com/google/uuid"
	"gorm.io/gorm"
)

const (
	GenerationStatusPending = "pending"
	GenerationStatusSuccess = "success"
	GenerationStatusFailed  = "failed"

	defaultCodeModel      = "openai/gpt-4o"
	defaultMaxTokens      = 2000
	openRouterCompletions = "https://openrouter.ai/api/v1/chat/completions"
)

var supportedLanguages = map[string]struct{}{
	"python":     {},
	"javascript": {},
}

// Generator generates code from specifications using LLM.
type Generator struct {
	db            *gorm.DB
	openRouterAPI string
}

// NewGenerator creates a new code generator
func NewGenerator(db *gorm.DB, openRouterAPI string) *Generator {
	return &Generator{
		db:            db,
		openRouterAPI: openRouterAPI,
	}
}

// GenerationRequest represents a code generation request
type GenerationRequest struct {
	AgentID      string       `json:"agent_id" validate:"required"`
	FunctionSpec FunctionSpec `json:"function_spec" validate:"required"`
	Language     string       `json:"language"` // python, javascript
	Runtime      string       `json:"runtime"`  // python311, nodejs18, etc.
}

// FunctionSpec defines the specification for a function to generate
type FunctionSpec struct {
	Name         string           `json:"name" validate:"required"`
	Title        string           `json:"title"`
	Description  string           `json:"description" validate:"required"`
	Prompt       string           `json:"prompt"`
	InputSchema  map[string]any   `json:"input_schema"`
	OutputSchema map[string]any   `json:"output_schema"`
	Category     string           `json:"category"`
	Tags         []string         `json:"tags"`
	Examples     []map[string]any `json:"examples"`
}

// GeneratedCode represents the result of code generation
type GeneratedCode struct {
	ID               uuid.UUID    `json:"id" gorm:"type:uuid;primaryKey;default:gen_random_uuid()"`
	AgentID          string       `json:"agent_id" gorm:"not null"`
	Request          FunctionSpec `json:"request" gorm:"type:jsonb"`
	GeneratedCode    string       `json:"generated_code" gorm:"type:text"`
	Language         string       `json:"language" gorm:"not null"`
	Runtime          string       `json:"runtime" gorm:"not null"`
	ModelUsed        string       `json:"model_used"`
	GenerationTimeMs int          `json:"generation_time_ms"`
	Status           string       `json:"status" gorm:"not null;default:'pending'"` // pending, success, failed
	ErrorMessage     *string      `json:"error_message"`
	CreatedAt        time.Time    `json:"created_at" gorm:"autoCreateTime"`
	UpdatedAt        time.Time    `json:"updated_at" gorm:"autoUpdateTime"`
}

// TableName returns the GORM table name
func (GeneratedCode) TableName() string {
	return "agent_generated_codes"
}

// Generate generates code from a specification
func (g *Generator) Generate(ctx context.Context, req *GenerationRequest) (*GeneratedCode, error) {
	startTime := time.Now()

	if err := validateGenerationRequest(req); err != nil {
		return nil, err
	}

	// Ensure we have a timeout when calling external services.
	if _, ok := ctx.Deadline(); !ok {
		var cancel context.CancelFunc
		ctx, cancel = context.WithTimeout(ctx, 60*time.Second)
		defer cancel()
	}

	generated := &GeneratedCode{
		ID:       uuid.New(),
		AgentID:  req.AgentID,
		Request:  req.FunctionSpec,
		Language: req.Language,
		Runtime:  req.Runtime,
		Status:   GenerationStatusPending,
	}

	// If no OpenRouter API key, use template-based generation
	if g.openRouterAPI == "" {
		code, err := g.generateFromTemplate(req)
		if err != nil {
			generated.Status = GenerationStatusFailed
			errMsg := err.Error()
			generated.ErrorMessage = &errMsg
			if dbErr := g.db.WithContext(ctx).Create(generated).Error; dbErr != nil {
				return nil, fmt.Errorf("failed to save failed template generation: %v (original error: %w)", dbErr, err)
			}
			return generated, fmt.Errorf("template generation failed: %w", err)
		}
		generated.GeneratedCode = code
		generated.Status = GenerationStatusSuccess
		generated.ModelUsed = "template"
	} else {
		code, model, err := g.generateWithLLM(ctx, req)
		if err != nil {
			generated.Status = GenerationStatusFailed
			errMsg := err.Error()
			generated.ErrorMessage = &errMsg
			if dbErr := g.db.WithContext(ctx).Create(generated).Error; dbErr != nil {
				return nil, fmt.Errorf("failed to save failed LLM generation: %v (original error: %w)", dbErr, err)
			}
			return generated, fmt.Errorf("LLM generation failed: %w", err)
		}
		generated.GeneratedCode = code
		generated.Status = GenerationStatusSuccess
		generated.ModelUsed = model
	}

	generated.GenerationTimeMs = int(time.Since(startTime).Milliseconds())

	if err := g.db.WithContext(ctx).Create(generated).Error; err != nil {
		return nil, fmt.Errorf("failed to save generated code: %w", err)
	}

	return generated, nil
}

// generateFromTemplate generates code from templates (fallback when no LLM available)
func (g *Generator) generateFromTemplate(req *GenerationRequest) (string, error) {
	// Basic template-based generation for simple functions
	switch req.Language {
	case "python":
		return g.generatePythonTemplate(req)
	case "javascript":
		return g.generateJavaScriptTemplate(req)
	default:
		return "", fmt.Errorf("unsupported language: %s", req.Language)
	}
}

func (g *Generator) generatePythonTemplate(req *GenerationRequest) (string, error) {
	funcName := req.FunctionSpec.Name
	if funcName == "" {
		funcName = "handler"
	}

	template := fmt.Sprintf(`"""
%s
Generated by FnSwarm Agent
"""

from typing import Any, Dict

def %s(data: Dict[str, Any]) -> Dict[str, Any]:
    """
    %s

    Args:
        data: Input data dictionary

    Returns:
        dict: Output result
    """
    # Parse input data
    input_data = data

    # Implement function logic based on specification
    # %s

    # Process the input and generate output
    result = process_input(input_data)

    return {
        "success": True,
        "result": result
    }


def process_input(input_data):
    """
    Process the input data and return the result.
    Modify this function to implement your specific logic.
    """
    # Default implementation: echo back the input with basic transformation
    if isinstance(input_data, dict):
        return {
            "processed": True,
            "data": input_data,
            "message": "Processed successfully"
        }
    return {"processed": True, "data": str(input_data)}
`, req.FunctionSpec.Title, funcName, req.FunctionSpec.Description, req.FunctionSpec.Prompt)

	return template, nil
}

func (g *Generator) generateJavaScriptTemplate(req *GenerationRequest) (string, error) {
	funcName := req.FunctionSpec.Name
	if funcName == "" {
		funcName = "handler"
	}

	template := fmt.Sprintf(`/**
 * %s
 * Generated by FnSwarm Agent
 */

/**
 * @typedef {Object} HandlerResult
 * @property {boolean} success
 * @property {Object} result
 */

/**
 * @typedef {Object} HandlerInput
 * @description Input payload as defined by the specification.
 */

/**
 * @param {HandlerInput} data - Input data
 * @returns {Promise<HandlerResult>} - Output result
 */
export async function %s(data) {
  // Parse input data
  const inputData = data;

  // Implement function logic based on the specification
  // %s

  // Process the input and generate output
  const result = processInput(inputData);

  return {
    success: true,
    result
  };
}

/**
 * Process the input data and return the result.
 * Modify this function to implement your specific logic.
 * @param {any} inputData - The input data to process
 * @returns {Object} - The processed result
 */
function processInput(inputData) {
  // Default implementation: echo back the input with basic transformation
  if (typeof inputData === 'object' && inputData !== null) {
    return {
      processed: true,
      data: inputData,
      message: 'Processed successfully'
    };
  }
  return { processed: true, data: String(inputData) };
}
`, req.FunctionSpec.Title, funcName, req.FunctionSpec.Description, req.FunctionSpec.Prompt)

	return template, nil
}

// generateWithLLM generates code using OpenRouter API
func (g *Generator) generateWithLLM(ctx context.Context, req *GenerationRequest) (string, string, error) {
	prompt := g.buildPrompt(req)

	body := map[string]any{
		"model": defaultCodeModel,
		"messages": []map[string]string{
			{
				"role":    "system",
				"content": "You are a code generation assistant. Return only executable code for the requested function.",
			},
			{
				"role":    "user",
				"content": prompt,
			},
		},
		"max_tokens":  defaultMaxTokens,
		"temperature": 0.2,
	}

	encoded, err := json.Marshal(body)
	if err != nil {
		return "", "", fmt.Errorf("failed to marshal OpenRouter request: %w", err)
	}

	reqHTTP, err := http.NewRequestWithContext(ctx, http.MethodPost, openRouterCompletions, bytes.NewReader(encoded))
	if err != nil {
		return "", "", fmt.Errorf("failed to create OpenRouter request: %w", err)
	}
	reqHTTP.Header.Set("Content-Type", "application/json")
	reqHTTP.Header.Set("Authorization", "Bearer "+g.openRouterAPI)

	resp, err := http.DefaultClient.Do(reqHTTP)
	if err != nil {
		return "", "", fmt.Errorf("OpenRouter request failed: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		const maxBody = 4 << 10
		snippet, _ := io.ReadAll(io.LimitReader(resp.Body, maxBody))
		return "", "", fmt.Errorf("OpenRouter returned non-200 status %d: %s", resp.StatusCode, strings.TrimSpace(string(snippet)))
	}

	var openResp struct {
		Choices []struct {
			Message struct {
				Content string `json:"content"`
			} `json:"message"`
		} `json:"choices"`
	}
	if err := json.NewDecoder(resp.Body).Decode(&openResp); err != nil {
		return "", "", fmt.Errorf("failed to decode OpenRouter response: %w", err)
	}
	if len(openResp.Choices) == 0 {
		return "", "", errors.New("OpenRouter response contained no choices")
	}

	code := strings.TrimSpace(openResp.Choices[0].Message.Content)
	if code == "" {
		return "", "", errors.New("OpenRouter returned empty code")
	}

	return code, defaultCodeModel, nil
}

func (g *Generator) buildPrompt(req *GenerationRequest) string {
	prompt := fmt.Sprintf("Generate a %s function for the following specification:\n\n", req.Language)
	prompt += fmt.Sprintf("Name: %s\n", req.FunctionSpec.Name)
	prompt += fmt.Sprintf("Title: %s\n", req.FunctionSpec.Title)
	prompt += fmt.Sprintf("Description: %s\n\n", req.FunctionSpec.Description)

	if req.FunctionSpec.InputSchema != nil {
		inputJSON, _ := json.Marshal(req.FunctionSpec.InputSchema)
		prompt += fmt.Sprintf("Input Schema: %s\n\n", string(inputJSON))
	}

	if req.FunctionSpec.OutputSchema != nil {
		outputJSON, _ := json.Marshal(req.FunctionSpec.OutputSchema)
		prompt += fmt.Sprintf("Output Schema: %s\n\n", string(outputJSON))
	}

	if len(req.FunctionSpec.Examples) > 0 {
		examplesJSON, _ := json.Marshal(req.FunctionSpec.Examples)
		prompt += fmt.Sprintf("Examples: %s\n\n", string(examplesJSON))
	}

	prompt += fmt.Sprintf("Use %s runtime.", req.Runtime)

	return prompt
}

// GetGenerations retrieves generated code for an agent
func (g *Generator) GetGenerations(ctx context.Context, agentID string, limit, offset int) ([]GeneratedCode, int64, error) {
	var total int64
	var generations []GeneratedCode

	query := g.db.WithContext(ctx).Model(&GeneratedCode{}).Where("agent_id = ?", agentID)
	if err := query.Count(&total).Error; err != nil {
		return nil, 0, fmt.Errorf("failed to count generations: %w", err)
	}

	if err := query.Order("created_at DESC").Limit(limit).Offset(offset).Find(&generations).Error; err != nil {
		return nil, 0, fmt.Errorf("failed to get generations: %w", err)
	}

	return generations, total, nil
}

// GetGeneration retrieves a specific generation
func (g *Generator) GetGeneration(ctx context.Context, generationID uuid.UUID) (*GeneratedCode, error) {
	var generated GeneratedCode
	err := g.db.WithContext(ctx).Where("id = ?", generationID).First(&generated).Error
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, fmt.Errorf("generation %s not found: %w", generationID, err)
		}
		return nil, fmt.Errorf("failed to get generation %s: %w", generationID, err)
	}
	return &generated, nil
}

func validateGenerationRequest(req *GenerationRequest) error {
	if req == nil {
		return errors.New("generation request is nil")
	}
	if strings.TrimSpace(req.AgentID) == "" {
		return errors.New("agent_id is required")
	}
	if strings.TrimSpace(req.FunctionSpec.Name) == "" {
		return errors.New("function_spec.name is required")
	}
	if strings.TrimSpace(req.FunctionSpec.Description) == "" {
		return errors.New("function_spec.description is required")
	}
	if _, ok := supportedLanguages[strings.ToLower(req.Language)]; !ok {
		return fmt.Errorf("unsupported language: %s", req.Language)
	}
	if strings.TrimSpace(req.Runtime) == "" {
		return errors.New("runtime is required")
	}
	return nil
}

package generation

import (
	"bytes"
	"context"
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"strings"
	"time"

	"github.com/functionfly/functionfly/internal/agent/identity"
	"github.com/google/uuid"
	"github.com/sirupsen/logrus"
	"gorm.io/gorm"
)

// CodeGenerator generates function source code from a request.
// Optional: when nil, Service uses a template stub. For production, inject an
// implementation (e.g. LLM via OpenAI, Anthropic) via NewServiceWithGenerator.
type CodeGenerator interface {
	GenerateCode(ctx context.Context, req *GenerationRequest) (string, error)
}

// DeterminismExecutor runs a function with given input and returns the output.
// Used by VerifyFunctionDeterminism to run the function multiple times and compare outputs.
// Optional: when nil, only cert/hash is verified. Inject via SetDeterminismExecutor (e.g. registry/sandbox).
type DeterminismExecutor interface {
	Execute(ctx context.Context, functionID uuid.UUID, input []byte) (output []byte, err error)
}

// Service handles AI-powered function generation
type Service struct {
	db      *gorm.DB
	codeGen CodeGenerator       // optional; nil means use template stub
	detExec DeterminismExecutor // optional; nil means only cert/hash in VerifyFunctionDeterminism
}

// NewService creates a new function generation service (template stub only).
func NewService(db *gorm.DB) *Service {
	return &Service{db: db}
}

// NewServiceWithGenerator creates a service that uses the given CodeGenerator for code generation.
// Use this to plug in an LLM-backed implementation.
func NewServiceWithGenerator(db *gorm.DB, codeGen CodeGenerator) *Service {
	return &Service{db: db, codeGen: codeGen}
}

// SetDeterminismExecutor sets the executor used for runtime determinism checks.
// When set and testInputs are provided, VerifyFunctionDeterminism runs the function
// with each input multiple times and verifies outputs are identical.
func (s *Service) SetDeterminismExecutor(exec DeterminismExecutor) {
	s.detExec = exec
}

// GenerationRequest represents a request to generate a function
type GenerationRequest struct {
	AgentID       string         `json:"agent_id"`
	Name          string         `json:"name"`
	Description   string         `json:"description"`
	Category      string         `json:"category"`
	InputSchema   map[string]any `json:"input_schema"`
	OutputSchema  map[string]any `json:"output_schema"`
	Runtime       string         `json:"runtime"` // python3.11, nodejs20, etc.
	Prompt        string         `json:"prompt"`
	Model         string         `json:"model"`
	Deterministic bool           `json:"deterministic"`
	Tags          []string       `json:"tags"`
}

// GenerationResult represents the result of function generation
type GenerationResult struct {
	FunctionID     uuid.UUID `json:"function_id"`
	Code           string    `json:"code"`
	Manifest       string    `json:"manifest"`
	CertHash       string    `json:"cert_hash,omitempty"`
	ModelUsed      string    `json:"model_used,omitempty"`
	Complexity     int       `json:"complexity,omitempty"`
	ReviewRequired bool      `json:"review_required,omitempty"`
	ReviewReason   string    `json:"review_reason,omitempty"`
	Success        bool      `json:"success"`
	Error          string    `json:"error,omitempty"`
}

// GenerateFunction generates a new function based on specifications
func (s *Service) GenerateFunction(ctx context.Context, req *GenerationRequest) (*GenerationResult, error) {
	// 1. Validate the request
	if req.Name == "" {
		return nil, fmt.Errorf("function name is required")
	}
	if req.Runtime == "" {
		req.Runtime = "python3.11"
	}
	selector := HeuristicModelSelector{}
	selection := selector.SelectModel(req)
	if req.Model == "" {
		req.Model = selection.Model
	}

	// 2. Generate the function code (uses injected CodeGenerator if set, else template stub)
	code, err := s.generateCode(ctx, req)
	if err != nil {
		return &GenerationResult{
			Success: false,
			Error:   fmt.Sprintf("Code generation failed: %v", err),
		}, nil
	}

	// 3. Create the function in the registry
	functionID := uuid.New()
	promptHash := hashString(req.Prompt)
	certHash := ""
	deterministicCert := interface{}(nil)
	if req.Deterministic {
		certHash = s.generateDeterministicCert(ctx, functionID, code, req)
		deterministicCert = &certHash
	}

	logrus.Infof("GenerateFunction: creating function: author=%s name=%s id=%s", req.AgentID, req.Name, functionID.String())

	tagsJSON, _ := json.Marshal(req.Tags)
	insertSQL := `INSERT INTO "registry_functions" ("id","author","name","latest_version","title","description","category","tags","visibility","price_per_call","popularity_score","reliability_score","deterministic_score","tenant_id","owner_user_id","owner_agent_id","agent_generated","generation_prompt_hash","generation_model","deterministic_cert_hash","revenue_total_usd","created_at","updated_at") VALUES ($1, $2, $3, '', $4, $5, $6, $7::jsonb, 'private', 0, 0, 0, 0, NULL, NULL, $8, true, $9, $10, $11, 0, NOW(), NOW()) ON CONFLICT ("author", "name") DO UPDATE SET "latest_version" = EXCLUDED."latest_version", "title" = EXCLUDED."title", "description" = EXCLUDED."description", "category" = EXCLUDED."category", "tags" = EXCLUDED."tags", "updated_at" = EXCLUDED."updated_at"`

	if err := s.db.WithContext(ctx).Exec(insertSQL, functionID.String(), req.AgentID, req.Name, req.Name, req.Description, req.Category, string(tagsJSON), req.AgentID, promptHash, req.Model, deterministicCert).Error; err != nil {
		logrus.Errorf("GenerateFunction: INSERT failed: %v", err)
		return nil, fmt.Errorf("failed to create function: %w", err)
	}
	logrus.Infof("GenerateFunction: INSERT succeeded, now querying for ID")

	var existingID string
	row := s.db.WithContext(ctx).Raw(`SELECT id FROM registry_functions WHERE author = $1 AND name = $2`, req.AgentID, req.Name).Row()
	if err := row.Scan(&existingID); err != nil {
		logrus.Errorf("GenerateFunction: SELECT failed or returned no rows: %v", err)
		return nil, fmt.Errorf("failed to get function ID: %w", err)
	}
	if existingID == "" {
		logrus.Errorf("GenerateFunction: SELECT returned empty id for author=%s name=%s", req.AgentID, req.Name)
		return nil, fmt.Errorf("function ID is empty after INSERT")
	}
	logrus.Infof("GenerateFunction: SELECT returned id=%s", existingID)
	functionID, err = uuid.Parse(existingID)
	if err != nil {
		logrus.Errorf("GenerateFunction: failed to parse existingID=%s: %v", existingID, err)
		return nil, fmt.Errorf("failed to parse function ID: %w", err)
	}

	if req.Deterministic {
		if err := s.db.WithContext(ctx).Exec(`
			UPDATE "registry_functions"
			SET "deterministic_cert_hash" = $1,
				"deterministic_score" = 100,
				"updated_at" = $2
			WHERE "author" = $3 AND "name" = $4
		`, certHash, time.Now(), req.AgentID, req.Name).Error; err != nil {
			return nil, fmt.Errorf("failed to update function with cert: %w", err)
		}
	}

	// 5. Create manifest
	manifest, err := s.createManifest(req, functionID)
	if err != nil {
		return nil, fmt.Errorf("failed to create manifest: %w", err)
	}

	return &GenerationResult{
		FunctionID:     functionID,
		Code:           code,
		Manifest:       manifest,
		CertHash:       certHash,
		ModelUsed:      req.Model,
		Complexity:     selection.Complexity,
		ReviewRequired: selection.Review.Required,
		ReviewReason:   selection.Review.Reason,
		Success:        true,
	}, nil
}

// generateCode returns generated function code from the injected CodeGenerator or the default template.
func (s *Service) generateCode(ctx context.Context, req *GenerationRequest) (string, error) {
	if s.codeGen != nil {
		return s.codeGen.GenerateCode(ctx, req)
	}
	// No CodeGenerator set: use template stub. For production, construct with NewServiceWithGenerator(db, llmGenerator).
	return s.generateDefaultCode(req), nil
}

// generateDefaultCode generates a basic function implementation based on the request
func (s *Service) generateDefaultCode(req *GenerationRequest) string {
	runtime := req.Runtime
	if runtime == "" {
		runtime = "python3.11"
	}

	// Determine language from runtime
	language := "python"
	if strings.Contains(runtime, "node") || strings.Contains(runtime, "javascript") || strings.Contains(runtime, "js") {
		language = "javascript"
	}

	if language == "javascript" {
		return s.generateJavaScriptCode(req)
	}
	return s.generatePythonCode(req)
}

// generatePythonCode generates Python function code
func (s *Service) generatePythonCode(req *GenerationRequest) string {
	inputSchemaJSON, _ := json.MarshalIndent(req.InputSchema, "    ", "  ")
	outputSchemaJSON, _ := json.MarshalIndent(req.OutputSchema, "    ", "  ")

	code := fmt.Sprintf(`import json
import logging

logger = logging.getLogger(__name__)


def handler(event):
    """
    %s

    Input Schema:
%s

    Output Schema:
%s

    Args:
        event: Input event data

    Returns:
        dict: Output result
    """
    logger.info("Processing event: %%s", event)

    # Parse input data
    data = event if isinstance(event, dict) else {"value": event}

    # Implement function logic based on the prompt
    # %s

    # Process the input and generate output
    result = process_data(data)

    return {"ok": True, "result": result}


def process_data(data):
    """
    Process the input data and return the result.
    Modify this function to implement your specific logic.

    Args:
        data: Input data dictionary

    Returns:
        dict: Processed result
    """
    # Default implementation: echo back the input with transformation
    if isinstance(data, dict):
        return {
            "processed": True,
            "data": data,
            "message": "Successfully processed"
        }
    return {"processed": True, "data": str(data)}


# For local testing
if __name__ == "__main__":
    test_event = {"test": "data"}
    result = handler(test_event)
    print(json.dumps(result, indent=2))
`, req.Description, string(inputSchemaJSON), string(outputSchemaJSON), req.Prompt)

	return code
}

// generateJavaScriptCode generates JavaScript function code
func (s *Service) generateJavaScriptCode(req *GenerationRequest) string {
	inputSchemaJSON, _ := json.MarshalIndent(req.InputSchema, "    ", "  ")
	outputSchemaJSON, _ := json.MarshalIndent(req.OutputSchema, "    ", "  ")

	code := fmt.Sprintf(`/**
 * %s
 * Generated by FnSwarm Agent
 *
 * Input Schema:
%s
 *
 * Output Schema:
%s
 */

/**
 * @param {Object} event - Input event data
 * @returns {Object} - Output result
 */
export async function handler(event) {
  console.log('Processing event:', JSON.stringify(event));

  // Parse input data
  const data = typeof event === 'object' && event !== null ? event : { value: event };

  // Implement function logic based on the prompt
  // %s

  // Process the input and generate output
  const result = processData(data);

  return { ok: true, result };
}

/**
 * Process the input data and return the result.
 * Modify this function to implement your specific logic.
 * @param {Object} data - Input data to process
 * @returns {Object} - Processed result
 */
function processData(data) {
  // Default implementation: echo back the input with transformation
  if (typeof data === 'object' && data !== null) {
    return {
      processed: true,
      data: data,
      message: 'Successfully processed'
    };
  }
  return { processed: true, data: String(data) };
}

// For local testing
// handler({ test: 'data' }).then(console.log);
`, req.Description, string(inputSchemaJSON), string(outputSchemaJSON), req.Prompt)

	return code
}

// generateDeterministicCert generates a deterministic certificate for a function
func (s *Service) generateDeterministicCert(ctx context.Context, functionID uuid.UUID, code string, req *GenerationRequest) string {
	// Create a deterministic hash of the function
	data := fmt.Sprintf("%s:%s:%s:%s",
		functionID.String(),
		code,
		jsonString(req.InputSchema),
		jsonString(req.OutputSchema),
	)

	hash := sha256.Sum256([]byte(data))
	return hex.EncodeToString(hash[:])
}

// createManifest creates the function manifest
func (s *Service) createManifest(req *GenerationRequest, functionID uuid.UUID) (string, error) {
	manifest := map[string]any{
		"name":          req.Name,
		"version":       "1.0.0",
		"description":   req.Description,
		"runtime":       req.Runtime,
		"input":         req.InputSchema,
		"output":        req.OutputSchema,
		"category":      req.Category,
		"tags":          req.Tags,
		"deterministic": req.Deterministic,
	}

	manifestJSON, err := json.MarshalIndent(manifest, "", "  ")
	if err != nil {
		return "", err
	}

	return string(manifestJSON), nil
}

// GetGeneratedFunctions retrieves functions generated by an agent
func (s *Service) GetGeneratedFunctions(ctx context.Context, agentID string) ([]identity.Function, error) {
	var functions []identity.Function
	err := s.db.WithContext(ctx).
		Where("owner_agent_id = ?", agentID).
		Order("created_at DESC").
		Find(&functions).Error
	return functions, err
}

// VerifyFunctionDeterminism verifies if a function maintains deterministic behavior.
// When s.detExec is set and testInputs are provided, runs the function with each input
// multiple times and verifies outputs are identical; otherwise only cert and hash are checked.
func (s *Service) VerifyFunctionDeterminism(ctx context.Context, functionID uuid.UUID, testInputs []map[string]any) (bool, string, error) {
	var function identity.Function
	if err := s.db.WithContext(ctx).Where("id = ?", functionID.String()).First(&function).Error; err != nil {
		return false, "", err
	}

	if function.DeterministicCertHash == nil {
		return false, "Function not certified as deterministic", nil
	}

	// Verify the cert hash matches current state
	expectedCert := s.generateDeterministicCert(ctx, functionID, "", nil)
	if *function.DeterministicCertHash != expectedCert {
		return false, "Function has been modified since certification", nil
	}

	// Runtime check: run with each test input multiple times and compare outputs
	if s.detExec != nil && len(testInputs) > 0 {
		const numRuns = 2
		for i, inp := range testInputs {
			inputJSON, err := json.Marshal(inp)
			if err != nil {
				return false, fmt.Sprintf("test input %d: invalid JSON", i), nil
			}
			var first []byte
			for run := 0; run < numRuns; run++ {
				out, err := s.detExec.Execute(ctx, functionID, inputJSON)
				if err != nil {
					return false, fmt.Sprintf("test input %d run %d: %v", i, run+1, err), nil
				}
				normalized := normalizeJSON(out)
				if run == 0 {
					first = normalized
					continue
				}
				if !bytes.Equal(first, normalized) {
					return false, fmt.Sprintf("test input %d: outputs differed across runs (non-deterministic)", i), nil
				}
			}
		}
	}

	return true, "Function is deterministic", nil
}

// normalizeJSON canonicalizes JSON for comparison (key order, spacing).
func normalizeJSON(b []byte) []byte {
	if len(b) == 0 {
		return b
	}
	var v any
	if err := json.Unmarshal(b, &v); err != nil {
		return b
	}
	out, _ := json.Marshal(v)
	return out
}

// PublishToMarketplace publishes a generated function to the marketplace
func (s *Service) PublishToMarketplace(ctx context.Context, functionID uuid.UUID, pricingModel string, pricePerCall *float64) error {
	var function identity.Function
	if err := s.db.WithContext(ctx).Where("id = ?", functionID.String()).First(&function).Error; err != nil {
		return err
	}

	// Update visibility to public
	function.Visibility = "public"
	if err := s.db.WithContext(ctx).Save(&function).Error; err != nil {
		return err
	}

	// Create marketplace listing
	listing := &identity.FunctionListing{
		ID:                    uuid.New(),
		FunctionID:            functionID,
		PricingModel:          pricingModel,
		PricePerCall:          pricePerCall,
		IsActive:              true,
		RatingScore:           0,
		CallVolume:            0,
		DeterministicVerified: function.DeterministicCertHash != nil,
		ListedAt:              time.Now(),
		UpdatedAt:             time.Now(),
	}

	return s.db.WithContext(ctx).Create(listing).Error
}

// hashString generates a SHA-256 hash of a string
func hashString(s string) string {
	hash := sha256.Sum256([]byte(s))
	return hex.EncodeToString(hash[:])
}

// jsonString converts a map to a JSON string
func jsonString(v any) string {
	if v == nil {
		return "{}"
	}
	b, _ := json.Marshal(v)
	return string(b)
}

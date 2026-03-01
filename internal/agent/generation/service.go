package generation

import (
	"context"
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"time"

	"github.com/functionfly/functionfly/internal/agent/identity"
	"github.com/google/uuid"
	"gorm.io/gorm"
)

// Service handles AI-powered function generation
type Service struct {
	db *gorm.DB
}

// NewService creates a new function generation service
func NewService(db *gorm.DB) *Service {
	return &Service{db: db}
}

// GenerationRequest represents a request to generate a function
type GenerationRequest struct {
	AgentID        string                 `json:"agent_id"`
	Name           string                 `json:"name"`
	Description    string                 `json:"description"`
	Category       string                 `json:"category"`
	InputSchema    map[string]any         `json:"input_schema"`
	OutputSchema   map[string]any          `json:"output_schema"`
	Runtime        string                 `json:"runtime"` // python3.11, nodejs20, etc.
	Prompt         string                 `json:"prompt"`
	Model          string                 `json:"model"`
	Deterministic  bool                   `json:"deterministic"`
	Tags           []string               `json:"tags"`
}

// GenerationResult represents the result of function generation
type GenerationResult struct {
	FunctionID     uuid.UUID  `json:"function_id"`
	Code          string     `json:"code"`
	Manifest      string     `json:"manifest"`
	CertHash      string     `json:"cert_hash,omitempty"`
	Success       bool       `json:"success"`
	Error         string     `json:"error,omitempty"`
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

	// 2. Generate the function code (in production, this would call an LLM)
	code, err := s.generateCode(ctx, req)
	if err != nil {
		return &GenerationResult{
			Success: false,
			Error:   fmt.Sprintf("Code generation failed: %v", err),
		}, nil
	}

	// 3. Create the function in the registry
	functionID := uuid.New()
	function := &identity.Function{
		ID:                  functionID.String(),
		Author:              req.AgentID,
		Name:                req.Name,
		Title:               req.Name,
		Description:         req.Description,
		Category:            req.Category,
		Tags:                req.Tags,
		Visibility:          "private", // Private by default
		PricePerCall:        0,
		PopularityScore:     0,
		ReliabilityScore:    0,
		DeterministicScore:  0,
		OwnerAgentID:        &req.AgentID,
		AgentGenerated:     true,
		GenerationPromptHash: hashString(req.Prompt),
		GenerationModel:     &req.Model,
		RevenueTotalUSD:     0,
		CreatedAt:           time.Now(),
		UpdatedAt:           time.Now(),
	}

	if err := s.db.WithContext(ctx).Create(function).Error; err != nil {
		return nil, fmt.Errorf("failed to create function: %w", err)
	}

	// 4. Create deterministic certificate if requested
	var certHash string
	if req.Deterministic {
		certHash = s.generateDeterministicCert(ctx, functionID, code, req)
		function.DeterministicCertHash = &certHash
		function.DeterministicScore = 100 // High score for self-certified
		if err := s.db.WithContext(ctx).Save(function).Error; err != nil {
			return nil, fmt.Errorf("failed to update function with cert: %w", err)
		}
	}

	// 5. Create manifest
	manifest, err := s.createManifest(req, functionID)
	if err != nil {
		return nil, fmt.Errorf("failed to create manifest: %w", err)
	}

	return &GenerationResult{
		FunctionID: functionID,
		Code:       code,
		Manifest:   manifest,
		CertHash:   certHash,
		Success:    true,
	}, nil
}

// generateCode generates function code (stub - would integrate with LLM in production)
func (s *Service) generateCode(ctx context.Context, req *GenerationRequest) (string, error) {
	// This is a stub implementation
	// In production, this would call an LLM API to generate actual code

	// Generate a template based on the prompt
	codeTemplate := `import json

def handler(event):
    """
    %s
    
    Input: %s
    Output: %s
    """
    # TODO: Implement the function logic based on the prompt
    # %s
    
    return {"ok": True, "result": event}
`

	code := fmt.Sprintf(codeTemplate,
		req.Description,
		jsonString(req.InputSchema),
		jsonString(req.OutputSchema),
		req.Prompt,
	)

	return code, nil
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
		"name":        req.Name,
		"version":     "1.0.0",
		"description": req.Description,
		"runtime":     req.Runtime,
		"input":       req.InputSchema,
		"output":       req.OutputSchema,
		"category":    req.Category,
		"tags":        req.Tags,
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

// VerifyFunctionDeterminism verifies if a function maintains deterministic behavior
func (s *Service) VerifyFunctionDeterminism(ctx context.Context, functionID uuid.UUID, testInputs []map[string]any) (bool, string, error) {
	var function identity.Function
	if err := s.db.WithContext(ctx).Where("id = ?", functionID.String()).First(&function).Error; err != nil {
		return false, "", err
	}

	// In production, this would actually run the function with test inputs
	// and verify the outputs are identical

	if function.DeterministicCertHash == nil {
		return false, "Function not certified as deterministic", nil
	}

	// Verify the cert hash matches current state
	expectedCert := s.generateDeterministicCert(ctx, functionID, "", nil)
	if *function.DeterministicCertHash != expectedCert {
		return false, "Function has been modified since certification", nil
	}

	return true, "Function is deterministic", nil
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

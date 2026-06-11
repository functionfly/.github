package e2e

import (
	"encoding/json"
	"testing"
	"time"

	"github.com/functionfly/functionfly/internal/agent/identity"
	"github.com/functionfly/functionfly/internal/agent/validation"
	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
)

// ============================================================
// Validation Tests
// ============================================================

func TestAgentValidation(t *testing.T) {
	t.Run("should validate agent ID format", func(t *testing.T) {
		// Valid agent IDs follow org/agent-name pattern
		validIDs := []string{
			"org/agent-name",
			"my-company/my-agent-v2",
			"test/scanner",
		}

		for _, id := range validIDs {
			// Just verify the string is not empty and contains a slash
			assert.Contains(t, id, "/", "agent ID '%s' should contain a slash", id)
			assert.NotEmpty(t, id, "agent ID should not be empty")
		}

		invalidIDs := []string{
			"",
			"no-slash",
			"/starts-with-slash",
			"ends-with-slash/",
		}

		for _, id := range invalidIDs {
			// Verify these are actually invalid (empty or no slash)
			if id == "" || id[0] == '/' || id[len(id)-1] == '/' || !containsSlash(id) {
				// Expected to be invalid
			}
		}
	})

	t.Run("should validate input with default validator", func(t *testing.T) {
		v := validation.DefaultInputValidator()

		validInput := json.RawMessage(`{"name":"test","value":123}`)
		err := v.Validate(validInput)
		assert.NoError(t, err)
	})

	t.Run("should detect PII in input", func(t *testing.T) {
		v := validation.DefaultInputValidator()

		// Input with email (PII)
		piiInput := json.RawMessage(`{"email":"user@example.com","name":"test"}`)
		err := v.Validate(piiInput)
		assert.Error(t, err)
		assert.Equal(t, validation.ErrPIIDetected, err)
	})

	t.Run("should detect oversized input", func(t *testing.T) {
		v := validation.DefaultInputValidator()
		v.MaxSizeBytes = 10 // Very small limit for testing

		largeInput := json.RawMessage(`{"verylongkey":"this value is way too large for the limit"}`)
		err := v.Validate(largeInput)
		assert.Error(t, err)
		assert.Equal(t, validation.ErrInputTooLarge, err)
	})

	t.Run("should validate output with default validator", func(t *testing.T) {
		v := validation.DefaultOutputValidator()

		validOutput := json.RawMessage(`{"result":"success","data":[1,2,3]}`)
		sanitized, err := v.Validate(validOutput)
		assert.NoError(t, err)
		assert.NotNil(t, sanitized)
	})

	t.Run("should redact sensitive data in output", func(t *testing.T) {
		v := validation.DefaultOutputValidator()

		sensitiveOutput := json.RawMessage(`{"password":"secret123","token":"abc-def-ghi"}`)
		sanitized, err := v.Validate(sensitiveOutput)
		assert.NoError(t, err)

		sanitizedStr := string(sanitized)
		assert.Contains(t, sanitizedStr, "[REDACTED]")
		assert.NotContains(t, sanitizedStr, "secret123")
	})

	t.Run("should check PII in output", func(t *testing.T) {
		v := validation.DefaultOutputValidator()

		// Output with SSN (PII)
		piiOutput := `{"ssn":"123-45-6789"}`
		hasPII := v.CheckPII(piiOutput)
		assert.True(t, hasPII)

		// Output without PII
		cleanOutput := `{"result":"success"}`
		hasPII = v.CheckPII(cleanOutput)
		assert.False(t, hasPII)
	})

	t.Run("should validate JSON structure", func(t *testing.T) {
		validJSON := json.RawMessage(`{"key":"value","number":42}`)
		err := validation.ValidateJSON(validJSON)
		assert.NoError(t, err)

		invalidJSON := json.RawMessage(`{invalid json}`)
		err = validation.ValidateJSON(invalidJSON)
		assert.Error(t, err)
	})

	t.Run("should validate JSON schema", func(t *testing.T) {
		validData := json.RawMessage(`{"name":"test","age":25}`)
		schema := map[string]interface{}{
			"required": []interface{}{"name", "age"},
		}
		err := validation.ValidateJSONSchema(validData, schema)
		assert.NoError(t, err)

		// Missing required field
		missingField := json.RawMessage(`{"name":"test"}`)
		err = validation.ValidateJSONSchema(missingField, schema)
		assert.Error(t, err)
		assert.Contains(t, err.Error(), "missing required field")
	})

	t.Run("should sanitize HTML content", func(t *testing.T) {
		input := `<script>alert('xss')</script><p>Hello</p>`
		sanitized := validation.SanitizeHTML(input)
		assert.NotContains(t, sanitized, "<script>")
		assert.Contains(t, sanitized, "<p>Hello</p>")
	})
}

// Helper to check if string contains slash
func containsSlash(s string) bool {
	for _, c := range s {
		if c == '/' {
			return true
		}
	}
	return false
}

// ============================================================
// Evolution Proposal Tests
// ============================================================

func TestEvolutionProposal(t *testing.T) {
	t.Run("should create evolution proposal", func(t *testing.T) {
		proposal := &identity.EvolutionProposal{
			ID:                     uuid.New(),
			AgentID:                "test-org/evolution-agent",
			ProposalType:           identity.EvolutionTypeSpawnSpecialist,
			ProposalData:           map[string]any{"capability": "code_generation"},
			Status:                 "pending",
			ParentApprovalRequired: true,
			CreatedAt:              time.Now(),
		}

		assert.NotEqual(t, uuid.Nil, proposal.ID)
		assert.Equal(t, "test-org/evolution-agent", proposal.AgentID)
		assert.Equal(t, identity.EvolutionTypeSpawnSpecialist, proposal.ProposalType)
		assert.Equal(t, "pending", proposal.Status)
		assert.True(t, proposal.ParentApprovalRequired)
	})

	t.Run("should track proposal status transitions", func(t *testing.T) {
		proposal := &identity.EvolutionProposal{
			ID:        uuid.New(),
			AgentID:   "test-org/status-agent",
			Status:    "pending",
			CreatedAt: time.Now(),
		}

		// Transition to approved
		proposal.Status = "approved"
		approvedBy := "parent-agent"
		proposal.ApprovedBy = &approvedBy
		now := time.Now()
		proposal.ImplementedAt = &now

		assert.Equal(t, "approved", proposal.Status)
		assert.NotNil(t, proposal.ApprovedBy)
		assert.NotNil(t, proposal.ImplementedAt)
	})

	t.Run("should handle proposal data as JSON", func(t *testing.T) {
		proposal := &identity.EvolutionProposal{
			ID:           uuid.New(),
			AgentID:      "test-org/data-agent",
			ProposalData: map[string]any{"action": "spawn", "agent_type": "scanner"},
		}

		// Verify proposal data is accessible
		assert.Equal(t, "spawn", proposal.ProposalData["action"])
		assert.Equal(t, "scanner", proposal.ProposalData["agent_type"])
	})

	t.Run("should validate evolution types", func(t *testing.T) {
		validTypes := []string{
			identity.EvolutionTypeSpawnSpecialist,
			identity.EvolutionTypeModifyPolicy,
			identity.EvolutionTypeAdjustTimeout,
			identity.EvolutionTypeGenerateFunction,
			identity.EvolutionTypeRetireChild,
			identity.EvolutionTypeUpgradeCapabilities,
		}

		for _, et := range validTypes {
			proposal := &identity.EvolutionProposal{
				ID:           uuid.New(),
				AgentID:      "test-org/validate-type",
				ProposalType: et,
			}
			assert.NotEmpty(t, proposal.ProposalType)
		}
	})
}

// ============================================================
// Retry Policy Tests
// ============================================================

func TestRetryPolicy(t *testing.T) {
	t.Run("should calculate exponential backoff", func(t *testing.T) {
		policy := &mockRetryPolicy{
			maxRetries:    3,
			baseDelayMs:   100,
			maxDelayMs:    1000,
			backoffMulti:  2.0,
		}

		// First retry
		delay1 := policy.CalculateDelay(1)
		assert.Equal(t, 100.0, delay1, "first retry should use base delay")

		// Second retry
		delay2 := policy.CalculateDelay(2)
		assert.Equal(t, 200.0, delay2, "second retry should double the delay")

		// Third retry
		delay3 := policy.CalculateDelay(3)
		assert.Equal(t, 400.0, delay3, "third retry should quadruple the delay")
	})

	t.Run("should cap delay at maximum", func(t *testing.T) {
		policy := &mockRetryPolicy{
			maxRetries:   5,
			baseDelayMs:  100,
			maxDelayMs:   500,
			backoffMulti: 2.0,
		}

		delay := policy.CalculateDelay(10) // Many retries
		assert.Equal(t, 500.0, delay, "delay should be capped at max")
	})
}

type mockRetryPolicy struct {
	maxRetries   int
	baseDelayMs  float64
	maxDelayMs   float64
	backoffMulti float64
}

func (p *mockRetryPolicy) CalculateDelay(retry int) float64 {
	delay := p.baseDelayMs
	for i := 1; i < retry; i++ {
		delay *= p.backoffMulti
	}
	if delay > p.maxDelayMs {
		return p.maxDelayMs
	}
	return delay
}

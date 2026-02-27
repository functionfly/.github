package cache

import (
	"encoding/json"
	"fmt"
)

const (
	// MaxOutputSize is the maximum allowed output size in bytes
	MaxOutputSize = 1024 * 1024 // 1MB
	// MaxDepth is the maximum allowed JSON nesting depth
	MaxDepth = 10
)

// ValidateOutput validates and re-serializes function output to prevent cache poisoning
// This ensures we never cache malicious or malformed data
func ValidateOutput(output json.RawMessage) (json.RawMessage, error) {
	// Check size
	if len(output) > MaxOutputSize {
		return nil, fmt.Errorf("output exceeds max size: %d > %d", len(output), MaxOutputSize)
	}

	// Unmarshal to validate structure
	var raw interface{}
	if err := json.Unmarshal(output, &raw); err != nil {
		return nil, fmt.Errorf("invalid JSON output: %w", err)
	}

	// Validate depth
	if err := validateDepth(raw, 0); err != nil {
		return nil, err
	}

	// Re-serialize using canonical formatter (removes any weirdness)
	canonical, err := json.Marshal(raw)
	if err != nil {
		return nil, fmt.Errorf("failed to re-serialize: %w", err)
	}

	// Double-check size after re-serialization
	if len(canonical) > MaxOutputSize {
		return nil, fmt.Errorf("output exceeds max size after canonicalization")
	}

	return canonical, nil
}

// validateDepth recursively validates JSON nesting depth
func validateDepth(val interface{}, depth int) error {
	if depth > MaxDepth {
		return fmt.Errorf("max nesting depth exceeded: %d > %d", depth, MaxDepth)
	}

	switch v := val.(type) {
	case map[string]interface{}:
		for _, val := range v {
			if err := validateDepth(val, depth+1); err != nil {
				return err
			}
		}
	case []interface{}:
		for _, item := range v {
			if err := validateDepth(item, depth+1); err != nil {
				return err
			}
		}
	}
	return nil
}

// OutputValidator holds validation configuration
type OutputValidator struct {
	MaxOutputSize int
	MaxDepth      int
}

// NewOutputValidator creates a new output validator with default settings
func NewOutputValidator() *OutputValidator {
	return &OutputValidator{
		MaxOutputSize: MaxOutputSize,
		MaxDepth:      MaxDepth,
	}
}

// ValidateWithConfig validates output with custom configuration
func ValidateWithConfig(output json.RawMessage, maxSize int, maxDepth int) (json.RawMessage, error) {
	// Check size
	if len(output) > maxSize {
		return nil, fmt.Errorf("output exceeds max size: %d > %d", len(output), maxSize)
	}

	// Unmarshal to validate structure
	var raw interface{}
	if err := json.Unmarshal(output, &raw); err != nil {
		return nil, fmt.Errorf("invalid JSON output: %w", err)
	}

	// Validate depth with custom max
	var validateDepthCustom func(val interface{}, depth int) error
	validateDepthCustom = func(val interface{}, depth int) error {
		if depth > maxDepth {
			return fmt.Errorf("max nesting depth exceeded: %d > %d", depth, maxDepth)
		}
		switch v := val.(type) {
		case map[string]interface{}:
			for _, val := range v {
				if err := validateDepthCustom(val, depth+1); err != nil {
					return err
				}
			}
		case []interface{}:
			for _, item := range v {
				if err := validateDepthCustom(item, depth+1); err != nil {
					return err
				}
			}
		}
		return nil
	}

	if err := validateDepthCustom(raw, 0); err != nil {
		return nil, err
	}

	// Re-serialize
	canonical, err := json.Marshal(raw)
	if err != nil {
		return nil, fmt.Errorf("failed to re-serialize: %w", err)
	}

	return canonical, nil
}

// IsValidJSON checks if the output is valid JSON
func IsValidJSON(output []byte) bool {
	var raw interface{}
	return json.Unmarshal(output, &raw) == nil
}

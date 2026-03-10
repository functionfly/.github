package apikey

import (
	"fmt"
	"time"

	"github.com/google/uuid"
)

// Validator validates API keys
type Validator struct {
	// currentTimeFunc can be replaced for testing
	currentTimeFunc func() time.Time
}

// NewValidator creates a new key validator
func NewValidator() *Validator {
	return &Validator{
		currentTimeFunc: time.Now,
	}
}

// ValidationResult contains the result of key validation
type ValidationResult struct {
	Valid     bool
	KeyType   KeyType
	Prefix    string
	Version   string
	ExpiresAt *time.Time
	IsExpired bool
	IsActive  bool
	Error     error
	Errors    []string
}

// Validate performs full validation of an API key
// It checks format, prefix, checksum, expiration, and active status
func (v *Validator) Validate(
	key string,
	storedHash string,
	isActive bool,
	expiresAt *time.Time,
) *ValidationResult {
	result := &ValidationResult{
		IsActive:  isActive,
		ExpiresAt: expiresAt,
	}

	// Step 1: Validate key format
	if err := ValidateKeyFormat(key); err != nil {
		result.Valid = false
		result.Errors = append(result.Errors, fmt.Sprintf("invalid key format: %v", err))
		result.Error = fmt.Errorf("invalid key format: %w", err)
		return result
	}

	// Step 2: Validate checksum
	if !VerifyChecksum(key) {
		result.Valid = false
		result.Errors = append(result.Errors, "invalid checksum")
		result.Error = fmt.Errorf("invalid checksum")
		return result
	}

	// Step 3: Extract key info
	info := ExtractKeyInfo(key)
	if info == nil {
		result.Valid = false
		result.Errors = append(result.Errors, "failed to extract key info")
		result.Error = fmt.Errorf("failed to extract key info")
		return result
	}

	result.Prefix = info.Prefix
	result.Version = info.Version

	// Step 4: Validate key type from prefix
	keyType, err := GetKeyTypeFromPrefix(info.Prefix)
	if err != nil {
		result.Valid = false
		result.Errors = append(result.Errors, fmt.Sprintf("unknown key prefix: %s", info.Prefix))
		result.Error = fmt.Errorf("unknown key prefix: %w", err)
		return result
	}
	result.KeyType = keyType

	// Step 5: Check if key is active
	if !isActive {
		result.Valid = false
		result.Errors = append(result.Errors, "key is not active")
		result.Error = fmt.Errorf("key is not active")
		return result
	}

	// Step 6: Check expiration
	if expiresAt != nil {
		if v.currentTimeFunc().After(*expiresAt) {
			result.IsExpired = true
			result.Valid = false
			result.Errors = append(result.Errors, "key has expired")
			result.Error = fmt.Errorf("key has expired")
			return result
		}
	}

	// All validations passed
	result.Valid = true
	return result
}

// ValidateFormatOnly validates just the key format (not against stored data)
func (v *Validator) ValidateFormatOnly(key string) *ValidationResult {
	result := &ValidationResult{}

	// Validate key format
	if err := ValidateKeyFormat(key); err != nil {
		result.Valid = false
		result.Errors = append(result.Errors, fmt.Sprintf("invalid key format: %v", err))
		result.Error = fmt.Errorf("invalid key format: %w", err)
		return result
	}

	// Validate checksum
	if !VerifyChecksum(key) {
		result.Valid = false
		result.Errors = append(result.Errors, "invalid checksum")
		result.Error = fmt.Errorf("invalid checksum")
		return result
	}

	// Extract key info
	info := ExtractKeyInfo(key)
	if info == nil {
		result.Valid = false
		result.Errors = append(result.Errors, "failed to extract key info")
		result.Error = fmt.Errorf("failed to extract key info")
		return result
	}

	result.Prefix = info.Prefix
	result.Version = info.Version

	// Validate key type from prefix
	keyType, err := GetKeyTypeFromPrefix(info.Prefix)
	if err != nil {
		result.Valid = false
		result.Errors = append(result.Errors, fmt.Sprintf("unknown key prefix: %s", info.Prefix))
		result.Error = fmt.Errorf("unknown key prefix: %w", err)
		return result
	}
	result.KeyType = keyType

	result.Valid = true
	return result
}

// ValidateKeyType validates that the key type matches the expected type
func ValidateKeyType(key string, expectedType KeyType) error {
	result := NewValidator().ValidateFormatOnly(key)
	if !result.Valid {
		return result.Error
	}

	if result.KeyType != expectedType {
		return fmt.Errorf("key type mismatch: expected %s, got %s", expectedType, result.KeyType)
	}

	return nil
}

// ValidatePermissions checks if the key has the required permission for a resource
func ValidatePermissions(
	keyPermissions []APIKeyPermission,
	requiredPermission Permission,
	resourceType ResourceType,
	resourceID uuid.UUID,
) bool {
	for _, perm := range keyPermissions {
		if perm.ResourceType == resourceType && perm.ResourceID == resourceID {
			if perm.Permission == PermissionAdmin {
				// Admin has all permissions
				return true
			}
			if perm.Permission == requiredPermission {
				return true
			}
		}
	}
	return false
}

// CheckRotationDue checks if a key needs rotation based on its rotation settings
func CheckRotationDue(
	lastRotatedAt time.Time,
	rotationFrequencyDays int,
) bool {
	if rotationFrequencyDays <= 0 {
		return false // Rotation disabled
	}

	rotationDue := lastRotatedAt.AddDate(0, 0, rotationFrequencyDays)
	return time.Now().After(rotationDue)
}

// ValidationErrors collects multiple validation errors
type KeyValidationErrors struct {
	errors []string
}

// Add adds a validation error
func (e *KeyValidationErrors) Add(format string, args ...interface{}) {
	e.errors = append(e.errors, fmt.Sprintf(format, args...))
}

// HasErrors returns true if there are any errors
func (e *KeyValidationErrors) HasErrors() bool {
	return len(e.errors) > 0
}

// Error returns all errors as a string
func (e *KeyValidationErrors) Error() string {
	if len(e.errors) == 0 {
		return ""
	}
	result := "validation errors:"
	for _, err := range e.errors {
		result += " " + err + ";"
	}
	return result
}

// ValidateCreateRequest validates a request to create an API key
func ValidateCreateRequest(req *CreateAPIKeyRequest) error {
	errors := &KeyValidationErrors{}

	if req.Name == "" {
		errors.Add("name is required")
	}

	if !IsValidKeyType(string(req.KeyType)) {
		errors.Add("invalid key type: %s", req.KeyType)
	}

	if req.RotationFrequencyDays < 0 {
		errors.Add("rotation frequency days must be non-negative")
	}

	if req.RateLimit != nil {
		if req.RateLimit.RPM < 0 {
			errors.Add("rate limit RPM must be non-negative")
		}
		if req.RateLimit.RPH < 0 {
			errors.Add("rate limit RPH must be non-negative")
		}
		if req.RateLimit.RPD < 0 {
			errors.Add("rate limit RPD must be non-negative")
		}
	}

	if errors.HasErrors() {
		return errors
	}

	return nil
}

// ValidateRotateRequest validates a request to rotate an API key
func ValidateRotateRequest(req *RotateAPIKeyRequest) error {
	errors := &KeyValidationErrors{}

	// Validate rotation reason if provided
	if req.Reason != "" {
		validReasons := map[RotationReason]bool{
			RotationReasonManual:      true,
			RotationReasonAutomatic:   true,
			RotationReasonCompromised: true,
		}
		if !validReasons[req.Reason] {
			errors.Add("invalid rotation reason: %s", req.Reason)
		}
	}

	if errors.HasErrors() {
		return errors
	}

	return nil
}

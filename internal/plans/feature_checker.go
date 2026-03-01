package plans

import (
	"fmt"
)

// FeatureChecker provides utilities for checking and validating features
type FeatureChecker struct {
	plan string
}

// NewFeatureChecker creates a new feature checker for a specific plan
func NewFeatureChecker(plan string) *FeatureChecker {
	return &FeatureChecker{plan: plan}
}

// HasFeature checks if the plan has a specific feature
func (fc *FeatureChecker) HasFeature(feature string) bool {
	return HasFeature(fc.plan, feature)
}

// HasAllFeatures checks if the plan has all the specified features
func (fc *FeatureChecker) HasAllFeatures(features ...string) bool {
	for _, f := range features {
		if !HasFeature(fc.plan, f) {
			return false
		}
	}
	return true
}

// HasAnyFeature checks if the plan has any of the specified features
func (fc *FeatureChecker) HasAnyFeature(features ...string) bool {
	for _, f := range features {
		if HasFeature(fc.plan, f) {
			return true
		}
	}
	return false
}

// GetMissingFeatures returns features that are not available for the plan
func (fc *FeatureChecker) GetMissingFeatures(features []string) []string {
	var missing []string
	for _, f := range features {
		if !HasFeature(fc.plan, f) {
			missing = append(missing, f)
		}
	}
	return missing
}

// ValidateFeatureAccess validates if the plan has access to the requested feature
// Returns an error if the feature is not available
func (fc *FeatureChecker) ValidateFeatureAccess(feature string) error {
	if !HasFeature(fc.plan, feature) {
		def, ok := GetFeatureDefinition(feature)
		if ok {
			return &FeatureAccessError{
				Feature:     feature,
				FeatureName: def.Name,
				Plan:        fc.plan,
				Message:     fmt.Sprintf("Feature '%s' is not available on %s tier", def.Name, fc.plan),
			}
		}
		return &FeatureAccessError{
			Feature: feature,
			Plan:    fc.plan,
			Message: fmt.Sprintf("Feature '%s' is not available on %s tier", feature, fc.plan),
		}
	}
	return nil
}

// ValidateFeaturesAccess validates access to multiple features
// Returns the first error encountered
func (fc *FeatureChecker) ValidateFeaturesAccess(features ...string) error {
	for _, f := range features {
		if err := fc.ValidateFeatureAccess(f); err != nil {
			return err
		}
	}
	return nil
}

// FeatureAccessError represents an error when a feature is not available
type FeatureAccessError struct {
	Feature     string
	FeatureName string
	Plan        string
	Message     string
}

func (e *FeatureAccessError) Error() string {
	return e.Message
}

// FeatureAccessErrorCode returns the error code for feature access errors
func (e *FeatureAccessError) ErrorCode() string {
	return "FEATURE_NOT_AVAILABLE"
}

// IsFeatureAccessError checks if an error is a feature access error
func IsFeatureAccessError(err error) bool {
	_, ok := err.(*FeatureAccessError)
	return ok
}

// GetRequiredPlan returns the minimum plan required for a feature
func GetRequiredPlan(feature string) string {
	// Check if it's enterprise only
	if IsEnterpriseOnly(feature) {
		return PlanEnterprise
	}
	// Check if it's pro only
	if IsProOnly(feature) {
		return PlanPro
	}
	// Otherwise it's available on starter
	return PlanStarter
}

// FeatureCheckResult represents the result of a feature check
type FeatureCheckResult struct {
	Feature     string `json:"feature"`
	Available   bool   `json:"available"`
	RequiredPlan string `json:"required_plan"`
}

// CheckFeatures checks multiple features and returns their availability
func (fc *FeatureChecker) CheckFeatures(features []string) []FeatureCheckResult {
	results := make([]FeatureCheckResult, len(features))
	for i, f := range results {
		results[i] = FeatureCheckResult{
			Feature:     f,
			Available:   HasFeature(fc.plan, f),
			RequiredPlan: GetRequiredPlan(f),
		}
	}
	return results
}

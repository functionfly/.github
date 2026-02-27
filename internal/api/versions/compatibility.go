package versions

import (
	"encoding/json"
	"fmt"
	"time"
)

// FieldTransform defines a transformation for a field
type FieldTransform struct {
	OldName      string                        // Original field name in old version
	NewName      string                        // New field name in new version
	DefaultValue interface{}                   // Default value for new fields in old version responses
	Transform    func(interface{}) interface{} // Custom transformation function
}

// EndpointCompatibility defines compatibility rules for an endpoint
type EndpointCompatibility struct {
	Endpoint    string
	Method      string
	FromVersion string
	ToVersion   string
	// Fields that were renamed
	FieldAliases []FieldTransform
	// Fields that were added in the new version
	NewFields []string
	// Fields that were removed (will be omitted)
	RemovedFields []string
	// Fields that changed type
	TypeChanges map[string]string // field name -> new type description
}

// CompatibilityLayer handles response transformations between API versions
type CompatibilityLayer struct {
	rules map[string]*EndpointCompatibility
}

// NewCompatibilityLayer creates a new compatibility layer
func NewCompatibilityLayer() *CompatibilityLayer {
	cl := &CompatibilityLayer{
		rules: make(map[string]*EndpointCompatibility),
	}

	// Register default compatibility rules for registry endpoints
	cl.registerRegistryCompatibilityRules()

	return cl
}

// registerRegistryCompatibilityRules registers compatibility rules for registry endpoints
func (cl *CompatibilityLayer) registerRegistryCompatibilityRules() {
	// v1 to v2 compatibility for /registry/functions endpoint
	cl.rules["GET:/v1/registry/functions"] = &EndpointCompatibility{
		Endpoint:    "/registry/functions",
		Method:      "GET",
		FromVersion: "v1",
		ToVersion:   "v2",
		FieldAliases: []FieldTransform{
			{
				OldName:   "popularity_score",
				NewName:   "popularityScore",
				Transform: transformPopularityScore,
			},
			{
				OldName: "deterministic_score",
				NewName: "deterministicScore",
			},
		},
		NewFields:     []string{"trust_score", "execution_count", "last_executed_at"},
		RemovedFields: []string{},
		TypeChanges: map[string]string{
			"popularity_score": "float (previously int)",
		},
	}

	// v1 to v2 compatibility for /registry/functions/{author}/{name} endpoint
	cl.rules["GET:/v1/registry/functions/{author}/{name}"] = &EndpointCompatibility{
		Endpoint:    "/registry/functions/{author}/{name}",
		Method:      "GET",
		FromVersion: "v1",
		ToVersion:   "v2",
		FieldAliases: []FieldTransform{
			{
				OldName:   "latest_version",
				NewName:   "latestVersion",
				Transform: transformVersion,
			},
		},
		NewFields:     []string{"trust_score", "verified", "signature_info"},
		RemovedFields: []string{},
	}

	// v1 to v2 compatibility for /registry/search endpoint
	cl.rules["GET:/v1/registry/search"] = &EndpointCompatibility{
		Endpoint:      "/registry/search",
		Method:        "GET",
		FromVersion:   "v1",
		ToVersion:     "v2",
		NewFields:     []string{"relevance_score", "highlights"},
		RemovedFields: []string{},
	}
}

// GetCompatibilityRule returns the compatibility rule for an endpoint
func (cl *CompatibilityLayer) GetCompatibilityRule(endpoint, method, fromVersion, toVersion string) *EndpointCompatibility {
	key := fmt.Sprintf("%s:%s", method, endpoint)
	return cl.rules[key]
}

// TransformResponse transforms a response from one version to another
func (cl *CompatibilityLayer) TransformResponse(endpoint, method, fromVersion, toVersion string, data []byte) ([]byte, error) {
	// If versions are the same, no transformation needed
	if fromVersion == toVersion {
		return data, nil
	}

	rule := cl.GetCompatibilityRule(endpoint, method, fromVersion, toVersion)
	if rule == nil {
		// No specific rule, try generic transformation
		return cl.genericTransform(data, fromVersion, toVersion)
	}

	// Parse the response data
	var responseData interface{}
	if err := json.Unmarshal(data, &responseData); err != nil {
		return nil, fmt.Errorf("failed to parse response data: %w", err)
	}

	// Apply transformations
	transformed := cl.applyTransformations(responseData, rule)

	// Marshal back to JSON
	result, err := json.Marshal(transformed)
	if err != nil {
		return nil, fmt.Errorf("failed to marshal transformed data: %w", err)
	}

	return result, nil
}

// genericTransform applies generic transformations (like snake_case to camelCase)
func (cl *CompatibilityLayer) genericTransform(data []byte, fromVersion, toVersion string) ([]byte, error) {
	// For now, return the original data if no specific rule exists
	// In production, this could apply generic transformations like snake_case to camelCase
	return data, nil
}

// applyTransformations applies the compatibility transformations
func (cl *CompatibilityLayer) applyTransformations(data interface{}, rule *EndpointCompatibility) interface{} {
	switch v := data.(type) {
	case map[string]interface{}:
		return cl.transformMap(v, rule)
	case []interface{}:
		result := make([]interface{}, len(v))
		for i, item := range v {
			result[i] = cl.applyTransformations(item, rule)
		}
		return result
	default:
		return data
	}
}

// transformMap applies transformations to a map
func (cl *CompatibilityLayer) transformMap(data map[string]interface{}, rule *EndpointCompatibility) map[string]interface{} {
	result := make(map[string]interface{})

	// Process field aliases (renamed fields)
	for _, alias := range rule.FieldAliases {
		if oldValue, exists := data[alias.OldName]; exists {
			newValue := oldValue
			if alias.Transform != nil {
				newValue = alias.Transform(oldValue)
			}
			result[alias.NewName] = newValue
		}
	}

	// Copy remaining fields that are not removed
	for key, value := range data {
		// Skip if this field was removed
		if containsString(key, rule.RemovedFields) {
			continue
		}

		// Skip if this field was renamed (already handled above)
		isRenamed := false
		for _, alias := range rule.FieldAliases {
			if key == alias.OldName {
				isRenamed = true
				break
			}
		}
		if isRenamed {
			continue
		}

		result[key] = value
	}

	// Add default values for new fields
	for _, newField := range rule.NewFields {
		if _, exists := result[newField]; !exists {
			result[newField] = getDefaultValue(newField)
		}
	}

	return result
}

// TransformRequest transforms a request from one version to another
func (cl *CompatibilityLayer) TransformRequest(endpoint, method, fromVersion, toVersion string, data []byte) ([]byte, error) {
	// Similar to TransformResponse but for request bodies
	if fromVersion == toVersion {
		return data, nil
	}

	// Parse the request data
	var requestData interface{}
	if err := json.Unmarshal(data, &requestData); err != nil {
		return nil, fmt.Errorf("failed to parse request data: %w", err)
	}

	// For v1 to v2, we would transform camelCase to snake_case for field names
	if fromVersion == "v1" && toVersion == "v2" {
		transformed := cl.transformRequestFields(requestData)
		result, err := json.Marshal(transformed)
		if err != nil {
			return nil, fmt.Errorf("failed to marshal transformed request: %w", err)
		}
		return result, nil
	}

	return data, nil
}

// transformRequestFields transforms request field names
func (cl *CompatibilityLayer) transformRequestFields(data interface{}) interface{} {
	switch v := data.(type) {
	case map[string]interface{}:
		result := make(map[string]interface{})
		for key, value := range v {
			// Transform camelCase to snake_case
			newKey := camelToSnake(key)
			result[newKey] = value
		}
		return result
	case []interface{}:
		result := make([]interface{}, len(v))
		for i, item := range v {
			result[i] = cl.transformRequestFields(item)
		}
		return result
	default:
		return data
	}
}

// Helper functions

func containsString(s string, list []string) bool {
	for _, item := range list {
		if s == item {
			return true
		}
	}
	return false
}

func getDefaultValue(fieldName string) interface{} {
	switch fieldName {
	case "trust_score":
		return 0.0
	case "execution_count":
		return 0
	case "last_executed_at":
		return nil
	case "verified":
		return false
	case "relevance_score":
		return 0.0
	case "highlights":
		return []string{}
	default:
		return nil
	}
}

func transformPopularityScore(value interface{}) interface{} {
	// Convert int to float for v1 -> v2 compatibility
	switch v := value.(type) {
	case float64:
		return v
	case float32:
		return float64(v)
	case int:
		return float64(v)
	case int64:
		return float64(v)
	default:
		return 0.0
	}
}

func transformVersion(value interface{}) interface{} {
	// Ensure version is a string
	switch v := value.(type) {
	case string:
		return v
	default:
		return ""
	}
}

// camelToSnake converts camelCase to snake_case
func camelToSnake(s string) string {
	result := make([]rune, 0, len(s)*2)
	for i, r := range s {
		if i > 0 && r >= 'A' && r <= 'Z' {
			result = append(result, '_')
		}
		result = append(result, rune(r-'A'+'a'))
	}
	return string(result)
}

// GetSchemaDiff returns the differences between two versions for an endpoint
func (cl *CompatibilityLayer) GetSchemaDiff(endpoint, method, v1, v2 string) map[string]interface{} {
	rule := cl.GetCompatibilityRule(endpoint, method, v1, v2)
	if rule == nil {
		return map[string]interface{}{
			"compatible": false,
			"message":    "No compatibility rules defined",
		}
	}

	return map[string]interface{}{
		"compatible":     true,
		"from_version":   rule.FromVersion,
		"to_version":     rule.ToVersion,
		"field_renames":  rule.FieldAliases,
		"new_fields":     rule.NewFields,
		"removed_fields": rule.RemovedFields,
		"type_changes":   rule.TypeChanges,
	}
}

// GetAllCompatibilityRules returns all registered compatibility rules
func (cl *CompatibilityLayer) GetAllCompatibilityRules() []*EndpointCompatibility {
	rules := make([]*EndpointCompatibility, 0, len(cl.rules))
	for _, rule := range cl.rules {
		rules = append(rules, rule)
	}
	return rules
}

// IsVersionCompatible checks if a source version can be transformed to target version
func (cl *CompatibilityLayer) IsVersionCompatible(sourceVersion, targetVersion string) bool {
	// v1 and v2 are both active, so they should be compatible
	// This could be extended with more complex compatibility checks
	return true
}

// DeprecationInfo contains deprecation information for an endpoint
type DeprecationInfo struct {
	Endpoint        string     `json:"endpoint"`
	Method          string     `json:"method"`
	DeprecatedIn    string     `json:"deprecated_in"`
	WillBeRemovedIn string     `json:"will_be_removed_in"`
	DeprecationDate *time.Time `json:"deprecation_date,omitempty"`
	SunsetDate      *time.Time `json:"sunset_date,omitempty"`
	MigrationGuide  string     `json:"migration_guide"`
	Alternative     string     `json:"alternative"`
	BreakingChanges []string   `json:"breaking_changes"`
}

// GetDeprecationInfo returns deprecation info for an endpoint
func (cl *CompatibilityLayer) GetDeprecationInfo(endpoint, method string) *DeprecationInfo {
	// Define deprecation info for v1 endpoints
	deprecations := map[string]*DeprecationInfo{
		"GET:/v1/registry/functions": {
			Endpoint:        "/registry/functions",
			Method:          "GET",
			DeprecatedIn:    "v1",
			WillBeRemovedIn: "v3",
			MigrationGuide:  "Use /v2/registry/functions for new integrations. The v2 endpoint returns camelCase field names and includes additional metadata fields.",
			Alternative:     "/v2/registry/functions",
			BreakingChanges: []string{"Field 'popularity_score' renamed to 'popularityScore' (type changed to float)"},
		},
	}

	key := fmt.Sprintf("%s:%s", method, endpoint)
	return deprecations[key]
}

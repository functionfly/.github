package versions

import (
	"encoding/json"
	"fmt"
	"reflect"
	"strconv"
	"strings"
	"time"
)

// ResponseTransformer handles response transformations between API versions
type ResponseTransformer struct {
	compatibility *CompatibilityLayer
}

// NewResponseTransformer creates a new response transformer
func NewResponseTransformer() *ResponseTransformer {
	return &ResponseTransformer{
		compatibility: NewCompatibilityLayer(),
	}
}

// Transform transforms response data from one version to another
func (rt *ResponseTransformer) Transform(endpoint, method, fromVersion, toVersion string, data []byte) ([]byte, error) {
	// If versions match, return original data
	if fromVersion == toVersion {
		return data, nil
	}

	// Use compatibility layer for transformation
	return rt.compatibility.TransformResponse(endpoint, method, fromVersion, toVersion, data)
}

// TransformMap transforms a map based on transformation rules
func (rt *ResponseTransformer) TransformMap(data map[string]interface{}, rules *EndpointCompatibility) map[string]interface{} {
	if rules == nil {
		return data
	}
	return rt.compatibility.transformMap(data, rules)
}

// FieldTransformer defines how to transform a single field
type FieldTransformer struct {
	FromField string
	ToField   string
	Transform func(interface{}) interface{}
}

// TransformFields transforms specific fields in a map
func TransformFields(data map[string]interface{}, transformers []FieldTransformer) map[string]interface{} {
	result := make(map[string]interface{})

	// Copy all fields first
	for k, v := range data {
		result[k] = v
	}

	// Apply transformations
	for _, t := range transformers {
		if v, exists := data[t.FromField]; exists {
			newValue := v
			if t.Transform != nil {
				newValue = t.Transform(v)
			}
			result[t.ToField] = newValue
			// Optionally remove old field
			// delete(result, t.FromField)
		}
	}

	return result
}

// TypeCoercion transforms values to match expected types
func TypeCoercion(value interface{}, targetType string) interface{} {
	switch targetType {
	case "float":
		return toFloat64(value)
	case "int":
		return toInt(value)
	case "string":
		return toString(value)
	case "bool":
		return toBool(value)
	case "time":
		return toTime(value)
	default:
		return value
	}
}

func toFloat64(value interface{}) float64 {
	switch v := value.(type) {
	case float64:
		return v
	case float32:
		return float64(v)
	case int:
		return float64(v)
	case int64:
		return float64(v)
	case string:
		f, _ := strconv.ParseFloat(v, 64)
		return f
	default:
		return 0
	}
}

func toInt(value interface{}) int {
	switch v := value.(type) {
	case float64:
		return int(v)
	case float32:
		return int(v)
	case int:
		return v
	case int64:
		return int(v)
	case string:
		i, _ := strconv.Atoi(v)
		return i
	default:
		return 0
	}
}

func toString(value interface{}) string {
	switch v := value.(type) {
	case string:
		return v
	case int, int64, float64, float32:
		return fmt.Sprintf("%v", v)
	case bool:
		return strconv.FormatBool(v)
	default:
		return ""
	}
}

func toBool(value interface{}) bool {
	switch v := value.(type) {
	case bool:
		return v
	case string:
		return strings.ToLower(v) == "true" || v == "1"
	case int, int64:
		return toInt(v) != 0
	default:
		return false
	}
}

func toTime(value interface{}) time.Time {
	switch v := value.(type) {
	case time.Time:
		return v
	case string:
		// Try multiple formats
		formats := []string{
			time.RFC3339,
			"2006-01-02T15:04:05Z",
			"2006-01-02T15:04:05.000Z",
			"2006-01-02",
		}
		for _, format := range formats {
			if t, err := time.Parse(format, v); err == nil {
				return t
			}
		}
	default:
		// Try JSON unmarshaling
		if s, ok := value.(string); ok {
			if t, err := time.Parse(time.RFC3339, s); err == nil {
				return t
			}
		}
	}
	return time.Time{}
}

// RenameFields creates a new map with renamed fields
func RenameFields(data map[string]interface{}, renames map[string]string) map[string]interface{} {
	result := make(map[string]interface{})

	for k, v := range data {
		if newName, ok := renames[k]; ok {
			result[newName] = v
		} else {
			result[k] = v
		}
	}

	return result
}

// AddDefaultFields adds default values for missing fields
func AddDefaultFields(data map[string]interface{}, defaults map[string]interface{}) map[string]interface{} {
	result := make(map[string]interface{})

	// Copy existing data
	for k, v := range data {
		result[k] = v
	}

	// Add defaults for missing fields
	for k, v := range defaults {
		if _, exists := result[k]; !exists {
			result[k] = v
		}
	}

	return result
}

// RemoveFields removes specified fields from a map
func RemoveFields(data map[string]interface{}, fields []string) map[string]interface{} {
	result := make(map[string]interface{})

	for k, v := range data {
		if !contains(fields, k) {
			result[k] = v
		}
	}

	return result
}

func contains(list []string, item string) bool {
	for _, i := range list {
		if i == item {
			return true
		}
	}
	return false
}

// TransformList transforms each item in a list
func TransformList(data []interface{}, transformer func(interface{}) interface{}) []interface{} {
	result := make([]interface{}, len(data))
	for i, item := range data {
		result[i] = transformer(item)
	}
	return result
}

// DeepTransform recursively transforms nested structures
func DeepTransform(data interface{}, transformer func(map[string]interface{}) map[string]interface{}) interface{} {
	switch v := data.(type) {
	case map[string]interface{}:
		return transformer(v)
	case []interface{}:
		result := make([]interface{}, len(v))
		for i, item := range v {
			result[i] = DeepTransform(item, transformer)
		}
		return result
	default:
		return data
	}
}

// SnakeToCamel converts snake_case to camelCase
func SnakeToCamel(s string) string {
	parts := strings.Split(s, "_")
	for i := 1; i < len(parts); i++ {
		parts[i] = strings.Title(parts[i])
	}
	return strings.Join(parts, "")
}

// CamelToSnake converts camelCase to snake_case
func CamelToSnake(s string) string {
	var result strings.Builder
	for i, r := range s {
		if i > 0 && r >= 'A' && r <= 'Z' {
			result.WriteRune('_')
		}
		result.WriteRune(r)
	}
	return strings.ToLower(result.String())
}

// MarshalJSONSafe safely marshals data to JSON, handling special types
func MarshalJSONSafe(data interface{}) ([]byte, error) {
	// Convert map[interface{}]interface{} to map[string]interface{} if needed
	converted := convertMapInterface(data)
	return json.Marshal(converted)
}

func convertMapInterface(v interface{}) interface{} {
	switch v := v.(type) {
	case map[interface{}]interface{}:
		result := make(map[string]interface{})
		for key, value := range v {
			strKey := fmt.Sprintf("%v", key)
			result[strKey] = convertMapInterface(value)
		}
		return result
	case []interface{}:
		result := make([]interface{}, len(v))
		for i, item := range v {
			result[i] = convertMapInterface(item)
		}
		return result
	default:
		return v
	}
}

// GetFieldValue gets a field value from a map using dot notation
func GetFieldValue(data map[string]interface{}, path string) (interface{}, bool) {
	parts := strings.Split(path, ".")
	current := interface{}(data)

	for _, part := range parts {
		switch c := current.(type) {
		case map[string]interface{}:
			if v, ok := c[part]; ok {
				current = v
			} else {
				return nil, false
			}
		default:
			return nil, false
		}
	}

	return current, true
}

// SetFieldValue sets a field value in a map using dot notation
func SetFieldValue(data map[string]interface{}, path string, value interface{}) {
	parts := strings.Split(path, ".")
	current := data

	for i := 0; i < len(parts)-1; i++ {
		part := parts[i]
		if _, ok := current[part]; !ok {
			current[part] = make(map[string]interface{})
		}
		switch c := current[part].(type) {
		case map[string]interface{}:
			current = c
		default:
			return
		}
	}

	current[parts[len(parts)-1]] = value
}

// TransformSchema transforms a JSON schema between versions
type SchemaTransformer struct {
	FieldMappings map[string]map[string]string // version -> oldField -> newField
}

func NewSchemaTransformer() *SchemaTransformer {
	return &SchemaTransformer{
		FieldMappings: map[string]map[string]string{
			"v1_to_v2": {
				"popularity_score":    "popularityScore",
				"deterministic_score": "deterministicScore",
				"latest_version":      "latestVersion",
				"published_at":        "publishedAt",
			},
		},
	}
}

// TransformSchemaFieldNames transforms field names in a schema
func (st *SchemaTransformer) TransformSchemaFieldNames(schema map[string]interface{}, fromVersion, toVersion string) map[string]interface{} {
	key := fmt.Sprintf("%s_to_%s", fromVersion, toVersion)
	mappings, ok := st.FieldMappings[key]
	if !ok {
		return schema
	}

	result := make(map[string]interface{})
	for k, v := range schema {
		if newName, ok := mappings[k]; ok {
			result[newName] = v
		} else {
			result[k] = v
		}
	}

	return result
}

// GetNestedValue safely retrieves nested values from a map
func GetNestedValue(data map[string]interface{}, path string) interface{} {
	parts := strings.Split(path, ".")
	current := data

	for _, part := range parts {
		if val, ok := current[part]; ok {
			if next, ok := val.(map[string]interface{}); ok {
				current = next
			} else {
				return val
			}
		}
	}

	return nil
}

// FlattenMap flattens a nested map into a single level with dot-notation keys
func FlattenMap(data map[string]interface{}, prefix string) map[string]interface{} {
	result := make(map[string]interface{})

	for key, value := range data {
		newKey := key
		if prefix != "" {
			newKey = prefix + "." + key
		}

		if nested, ok := value.(map[string]interface{}); ok {
			flattened := FlattenMap(nested, newKey)
			for k, v := range flattened {
				result[k] = v
			}
		} else {
			result[newKey] = value
		}
	}

	return result
}

// UnflattenMap converts a flat map with dot-notation keys back to nested structure
func UnflattenMap(data map[string]interface{}) map[string]interface{} {
	result := make(map[string]interface{})

	for key, value := range data {
		SetFieldValue(result, key, value)
	}

	return result
}

// CoerceTypes ensures all values in a map match their expected types
func CoerceTypes(data map[string]interface{}, typeHints map[string]reflect.Kind) map[string]interface{} {
	result := make(map[string]interface{})

	for key, value := range data {
		if hint, ok := typeHints[key]; ok {
			result[key] = coerceValue(value, hint)
		} else {
			result[key] = value
		}
	}

	return result
}

func coerceValue(value interface{}, kind reflect.Kind) interface{} {
	switch kind {
	case reflect.Float64:
		return toFloat64(value)
	case reflect.Int:
		return toInt(value)
	case reflect.String:
		return toString(value)
	case reflect.Bool:
		return toBool(value)
	default:
		return value
	}
}

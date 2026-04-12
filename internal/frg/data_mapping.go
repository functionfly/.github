package frg

import (
	"fmt"
	"strconv"
	"strings"

	"github.com/PaesslerAG/jsonpath"
	"github.com/sirupsen/logrus"
)

// applyDataMapping transforms source data to target input based on mapping rules
func (e *ExecutionEngine) applyDataMapping(target, source map[string]interface{}, mapping *DataMapping) {
	if mapping.SourcePath == "*" || mapping.SourcePath == "" {
		// Merge all source data into target
		for k, v := range source {
			target[k] = v
		}
		return
	}

	// Evaluate source path (JSONPath or simple field)
	var sourceValue interface{}
	if strings.HasPrefix(mapping.SourcePath, "$") {
		// JSONPath expression - use simple Get helper
		result, err := jsonpath.Get(mapping.SourcePath, source)
		if err != nil {
			logrus.WithError(err).WithField("source_path", mapping.SourcePath).Debug("JSONPath evaluation failed")
			return
		}
		switch v := result.(type) {
		case []interface{}:
			if len(v) == 1 {
				sourceValue = v[0]
			} else if len(v) > 1 {
				sourceValue = v
			} else {
				return // No results
			}
		default:
			sourceValue = result
		}
	} else {
		// Simple field lookup
		if val, ok := source[mapping.SourcePath]; ok {
			sourceValue = val
		} else {
			return // Field not found
		}
	}

	// Determine target field/path
	targetPath := mapping.TargetPath
	if targetPath == "" {
		targetPath = mapping.SourcePath
		// Strip JSONPath prefix for default target
		if strings.HasPrefix(targetPath, "$") {
			parts := strings.Split(targetPath, ".")
			if len(parts) > 1 {
				targetPath = parts[len(parts)-1]
			} else {
				targetPath = strings.TrimPrefix(targetPath, "$")
			}
		}
	}

	// Apply to target (simple field or nested via helper)
	if strings.Contains(targetPath, ".") {
		e.setNestedValue(target, targetPath, sourceValue)
	} else {
		target[targetPath] = sourceValue
	}
}

// setNestedValue sets a value at a nested path (e.g., "user.name")
func (e *ExecutionEngine) setNestedValue(target map[string]interface{}, path string, value interface{}) {
	parts := strings.Split(path, ".")
	current := target
	for i, part := range parts[:len(parts)-1] {
		if next, ok := current[part]; ok {
			if nextMap, ok := next.(map[string]interface{}); ok {
				current = nextMap
			} else {
				// Path conflict, create new map
				newMap := make(map[string]interface{})
				current[part] = newMap
				current = newMap
			}
		} else {
			newMap := make(map[string]interface{})
			current[part] = newMap
			current = newMap
		}
		_ = i
	}
	current[parts[len(parts)-1]] = value
}

// extractFieldValue extracts a field value from data using dot notation (e.g., "user.name")
func extractFieldValue(data interface{}, field string) interface{} {
	if field == "" {
		return data
	}

	// Handle JSONPath-style prefix
	field = strings.TrimPrefix(field, "$")
	field = strings.TrimPrefix(field, ".")

	parts := strings.Split(field, ".")
	current := data

	for _, part := range parts {
		switch v := current.(type) {
		case map[string]interface{}:
			if val, ok := v[part]; ok {
				current = val
			} else {
				return nil
			}
		default:
			return nil
		}
	}

	return current
}

// evaluateCondition evaluates a routing condition for stream data
func (e *ExecutionEngine) evaluateCondition(data interface{}, condition *Condition) bool {
	if condition == nil {
		return true
	}

	// Extract field value using dot notation or JSONPath
	fieldValue := extractFieldValue(data, condition.Field)

	switch condition.Operator {
	case "eq", "==":
		return fmt.Sprintf("%v", fieldValue) == fmt.Sprintf("%v", condition.Value)
	case "ne", "!=":
		return fmt.Sprintf("%v", fieldValue) != fmt.Sprintf("%v", condition.Value)
	case "gt", ">":
		return compareNumeric(fieldValue, condition.Value) > 0
	case "gte", ">=":
		return compareNumeric(fieldValue, condition.Value) >= 0
	case "lt", "<":
		return compareNumeric(fieldValue, condition.Value) < 0
	case "lte", "<=":
		return compareNumeric(fieldValue, condition.Value) <= 0
	case "contains":
		if str, ok := fieldValue.(string); ok {
			if search, ok := condition.Value.(string); ok {
				return strings.Contains(str, search)
			}
		}
		return false
	case "exists":
		return fieldValue != nil
	default:
		// Unknown operator - pass through
		return true
	}
}

// compareNumeric compares two numeric values, returns -1, 0, or 1
func compareNumeric(a, b interface{}) int {
	af, ok1 := toFloat64(a)
	bf, ok2 := toFloat64(b)
	if !ok1 || !ok2 {
		return 0
	}
	if af < bf {
		return -1
	} else if af > bf {
		return 1
	}
	return 0
}

// toFloat64 attempts to convert a value to float64
func toFloat64(v interface{}) (float64, bool) {
	switch val := v.(type) {
	case float64:
		return val, true
	case float32:
		return float64(val), true
	case int:
		return float64(val), true
	case int32:
		return float64(val), true
	case int64:
		return float64(val), true
	case string:
		f, err := strconv.ParseFloat(val, 64)
		return f, err == nil
	default:
		return 0, false
	}
}

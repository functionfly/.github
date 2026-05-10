package timemachine

import (
	"bytes"
	"encoding/json"
	"fmt"
	"reflect"
	"strings"
)

type DiffType string

const (
	DiffTypeIdentical DiffType = "identical"
	DiffTypeMinor     DiffType = "minor"
	DiffTypeMajor     DiffType = "major"
	DiffTypeBreaking  DiffType = "breaking"
	DiffTypeError     DiffType = "error"
)

type FieldDiff struct {
	Path      string      `json:"path"`
	Operation string      `json:"operation"`
	OldValue  interface{} `json:"old_value,omitempty"`
	NewValue  interface{} `json:"new_value,omitempty"`
}

type DiffResult struct {
	Type        DiffType    `json:"type"`
	Summary     string      `json:"summary"`
	Changed     bool        `json:"changed"`
	FieldDiffs  []FieldDiff `json:"field_diffs,omitempty"`
	TotalFields int         `json:"total_fields"`
	ChangedFields int       `json:"changed_fields"`
	PercentDiff float64     `json:"percent_diff"`
}

const (
	minorThreshold = 0.05
	majorThreshold = 0.20
)

// CompareOutputs performs a JSON-aware comparison of two outputs and classifies the diff.
func CompareOutputs(oldOutput, newOutput json.RawMessage) *DiffResult {
	if len(oldOutput) == 0 && len(newOutput) == 0 {
		return &DiffResult{Type: DiffTypeIdentical, Summary: "Both outputs are empty"}
	}

	if bytes.Equal(normalizeJSON(oldOutput), normalizeJSON(newOutput)) {
		return &DiffResult{Type: DiffTypeIdentical, Summary: "Outputs are byte-equal after normalization"}
	}

	var oldVal, newVal interface{}
	if err := json.Unmarshal(oldOutput, &oldVal); err != nil {
		return &DiffResult{
			Type:    DiffTypeBreaking,
			Summary: fmt.Sprintf("Original output is not valid JSON: %v", err),
			Changed: true,
		}
	}
	if err := json.Unmarshal(newOutput, &newVal); err != nil {
		return &DiffResult{
			Type:    DiffTypeError,
			Summary: fmt.Sprintf("New output is not valid JSON: %v", err),
			Changed: true,
		}
	}

	if reflect.TypeOf(oldVal) != reflect.TypeOf(newVal) {
		return &DiffResult{
			Type:    DiffTypeBreaking,
			Summary: fmt.Sprintf("Output type changed from %T to %T", oldVal, newVal),
			Changed: true,
		}
	}

	fieldDiffs := make([]FieldDiff, 0)
	compareValues("$", oldVal, newVal, &fieldDiffs)

	totalFields := countLeafFields(oldVal)
	changedFields := len(fieldDiffs)

	var percentDiff float64
	if totalFields > 0 {
		percentDiff = float64(changedFields) / float64(totalFields)
	} else if changedFields > 0 {
		percentDiff = 1.0
	}

	result := &DiffResult{
		Changed:       changedFields > 0,
		FieldDiffs:    fieldDiffs,
		TotalFields:   totalFields,
		ChangedFields: changedFields,
		PercentDiff:   percentDiff,
	}

	if changedFields == 0 {
		result.Type = DiffTypeIdentical
		result.Summary = "Outputs are semantically equivalent"
		return result
	}

	hasSchemaChange := false
	for _, fd := range fieldDiffs {
		if fd.Operation == "add" || fd.Operation == "remove" {
			hasSchemaChange = true
			break
		}
	}

	switch {
	case hasSchemaChange && percentDiff > minorThreshold:
		result.Type = DiffTypeBreaking
		result.Summary = fmt.Sprintf("Schema contract changed: %d fields added/removed", changedFields)
	case percentDiff > majorThreshold:
		result.Type = DiffTypeMajor
		result.Summary = fmt.Sprintf("Major changes: %d/%d fields differ (%.1f%%)", changedFields, totalFields, percentDiff*100)
	default:
		result.Type = DiffTypeMinor
		result.Summary = fmt.Sprintf("Minor changes: %d/%d fields differ (%.1f%%)", changedFields, totalFields, percentDiff*100)
	}

	return result
}

// CompareOutputsForError handles the case where the new execution produced an error.
func CompareOutputsForError(oldOutput json.RawMessage, errMsg string) *DiffResult {
	return &DiffResult{
		Type:    DiffTypeError,
		Summary: fmt.Sprintf("Replay execution failed: %s", errMsg),
		Changed: true,
		FieldDiffs: []FieldDiff{
			{
				Path:      "$",
				Operation: "replace",
				OldValue:  json.RawMessage(oldOutput),
				NewValue:  errMsg,
			},
		},
		TotalFields:   1,
		ChangedFields: 1,
		PercentDiff:   1.0,
	}
}

func compareValues(path string, oldVal, newVal interface{}, diffs *[]FieldDiff) {
	if oldVal == nil && newVal == nil {
		return
	}
	if oldVal == nil || newVal == nil {
		*diffs = append(*diffs, FieldDiff{Path: path, Operation: "replace", OldValue: oldVal, NewValue: newVal})
		return
	}

	switch oldTyped := oldVal.(type) {
	case map[string]interface{}:
		newMap, ok := newVal.(map[string]interface{})
		if !ok {
			*diffs = append(*diffs, FieldDiff{Path: path, Operation: "replace", OldValue: oldVal, NewValue: newVal})
			return
		}
		allKeys := make(map[string]bool)
		for k := range oldTyped {
			allKeys[k] = true
		}
		for k := range newMap {
			allKeys[k] = true
		}
		for k := range allKeys {
			childPath := path + "." + k
			oldChild, oldOk := oldTyped[k]
			newChild, newOk := newMap[k]
			switch {
			case oldOk && !newOk:
				*diffs = append(*diffs, FieldDiff{Path: childPath, Operation: "remove", OldValue: oldChild})
			case !oldOk && newOk:
				*diffs = append(*diffs, FieldDiff{Path: childPath, Operation: "add", NewValue: newChild})
			default:
				compareValues(childPath, oldChild, newChild, diffs)
			}
		}

	case []interface{}:
		newArr, ok := newVal.([]interface{})
		if !ok {
			*diffs = append(*diffs, FieldDiff{Path: path, Operation: "replace", OldValue: oldVal, NewValue: newVal})
			return
		}
		maxLen := len(oldTyped)
		if len(newArr) > maxLen {
			maxLen = len(newArr)
		}
		for i := 0; i < maxLen; i++ {
			childPath := fmt.Sprintf("%s[%d]", path, i)
			switch {
			case i >= len(oldTyped):
				*diffs = append(*diffs, FieldDiff{Path: childPath, Operation: "add", NewValue: newArr[i]})
			case i >= len(newArr):
				*diffs = append(*diffs, FieldDiff{Path: childPath, Operation: "remove", OldValue: oldTyped[i]})
			default:
				compareValues(childPath, oldTyped[i], newArr[i], diffs)
			}
		}

	default:
		if !reflect.DeepEqual(oldVal, newVal) {
			*diffs = append(*diffs, FieldDiff{Path: path, Operation: "replace", OldValue: oldVal, NewValue: newVal})
		}
	}
}

func countLeafFields(v interface{}) int {
	if v == nil {
		return 0
	}
	switch typed := v.(type) {
	case map[string]interface{}:
		count := 0
		for _, child := range typed {
			count += countLeafFields(child)
		}
		if count == 0 {
			return 1
		}
		return count
	case []interface{}:
		count := 0
		for _, child := range typed {
			count += countLeafFields(child)
		}
		if count == 0 {
			return 1
		}
		return count
	default:
		return 1
	}
}

func normalizeJSON(data json.RawMessage) []byte {
	if len(data) == 0 {
		return nil
	}
	var v interface{}
	if err := json.Unmarshal(data, &v); err != nil {
		return data
	}
	normalized, err := json.Marshal(v)
	if err != nil {
		return data
	}
	return normalized
}

// DiffSummaryToJSON marshals a DiffResult to JSON for storage.
func DiffSummaryToJSON(result *DiffResult) (string, json.RawMessage) {
	summary := result.Type.String() + ": " + result.Summary
	detail, _ := json.Marshal(result)
	return summary, detail
}

// ClassifyDiffTypeFromDB parses a diff_type string from the database.
func ClassifyDiffTypeFromDB(s string) DiffType {
	switch strings.ToLower(strings.TrimSpace(s)) {
	case "identical":
		return DiffTypeIdentical
	case "minor":
		return DiffTypeMinor
	case "major":
		return DiffTypeMajor
	case "breaking":
		return DiffTypeBreaking
	case "error":
		return DiffTypeError
	default:
		return DiffTypeIdentical
	}
}

func (d DiffType) String() string {
	return string(d)
}

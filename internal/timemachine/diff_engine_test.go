package timemachine

import (
	"encoding/json"
	"testing"
)

func TestCompareOutputs_Identical(t *testing.T) {
	old := json.RawMessage(`{"name":"test","value":42}`)
	new := json.RawMessage(`{"name":"test","value":42}`)

	result := CompareOutputs(old, new)
	if result.Type != DiffTypeIdentical {
		t.Errorf("expected identical, got %s", result.Type)
	}
	if result.Changed {
		t.Error("expected Changed=false")
	}
}

func TestCompareOutputs_IdenticalNormalized(t *testing.T) {
	old := json.RawMessage(`{"value":42,"name":"test"}`)
	new := json.RawMessage(`{"name":"test","value":42}`)

	result := CompareOutputs(old, new)
	if result.Type != DiffTypeIdentical {
		t.Errorf("expected identical after normalization, got %s", result.Type)
	}
}

func TestCompareOutputs_EmptyOutputs(t *testing.T) {
	result := CompareOutputs(nil, nil)
	if result.Type != DiffTypeIdentical {
		t.Errorf("expected identical for empty outputs, got %s", result.Type)
	}

	result = CompareOutputs(json.RawMessage(``), json.RawMessage(``))
	if result.Type != DiffTypeIdentical {
		t.Errorf("expected identical for empty strings, got %s", result.Type)
	}
}

func TestCompareOutputs_MinorChange(t *testing.T) {
	old := json.RawMessage(`{"a":"1","b":"2","c":"3","d":"4","e":"5","f":"6","g":"7","h":"8","i":"9","j":"10"}`)
	new := json.RawMessage(`{"a":"1","b":"2","c":"3","d":"4","e":"5","f":"6","g":"7","h":"8","i":"9","j":"CHANGED"}`)

	result := CompareOutputs(old, new)
	if result.Type != DiffTypeMinor {
		t.Errorf("expected minor, got %s (changed=%d/%d)", result.Type, result.ChangedFields, result.TotalFields)
	}
	if !result.Changed {
		t.Error("expected Changed=true")
	}
	if result.ChangedFields != 1 {
		t.Errorf("expected 1 changed field, got %d", result.ChangedFields)
	}
}

func TestCompareOutputs_MajorChange(t *testing.T) {
	old := json.RawMessage(`{"a":"1","b":"2","c":"3","d":"4","e":"5"}`)
	new := json.RawMessage(`{"a":"X","b":"X","c":"X","d":"X","e":"X"}`)

	result := CompareOutputs(old, new)
	if result.Type != DiffTypeMajor {
		t.Errorf("expected major, got %s (changed=%d/%d)", result.Type, result.ChangedFields, result.TotalFields)
	}
}

func TestCompareOutputs_BreakingSchemaChange(t *testing.T) {
	old := json.RawMessage(`{"a":"1","b":"2","c":"3"}`)
	new := json.RawMessage(`{"a":"1","b":"2","new_field":"3"}`)

	result := CompareOutputs(old, new)
	if result.Type != DiffTypeBreaking {
		t.Errorf("expected breaking for schema change, got %s", result.Type)
	}
}

func TestCompareOutputs_TypeChange(t *testing.T) {
	old := json.RawMessage(`{"value":42}`)
	new := json.RawMessage(`{"value":"42"}`)

	result := CompareOutputs(old, new)
	if !result.Changed {
		t.Error("expected change detected for type mismatch")
	}
}

func TestCompareOutputs_NestedObject(t *testing.T) {
	old := json.RawMessage(`{"user":{"name":"Alice","age":30}}`)
	new := json.RawMessage(`{"user":{"name":"Alice","age":31}}`)

	result := CompareOutputs(old, new)
	if !result.Changed {
		t.Error("expected change in nested object")
	}
	if result.ChangedFields != 1 {
		t.Errorf("expected 1 changed field, got %d", result.ChangedFields)
	}
}

func TestCompareOutputs_ArrayChange(t *testing.T) {
	old := json.RawMessage(`{"items":[1,2,3]}`)
	new := json.RawMessage(`{"items":[1,2,4]}`)

	result := CompareOutputs(old, new)
	if !result.Changed {
		t.Error("expected change in array")
	}
}

func TestCompareOutputs_ArrayLengthChange(t *testing.T) {
	old := json.RawMessage(`{"items":[1,2,3]}`)
	new := json.RawMessage(`{"items":[1,2,3,4]}`)

	result := CompareOutputs(old, new)
	if !result.Changed {
		t.Error("expected change for array length difference")
	}
}

func TestCompareOutputs_NullValues(t *testing.T) {
	old := json.RawMessage(`{"value":null}`)
	new := json.RawMessage(`{"value":"not null"}`)

	result := CompareOutputs(old, new)
	if !result.Changed {
		t.Error("expected change from null to non-null")
	}
}

func TestCompareOutputs_InvalidJSON(t *testing.T) {
	old := json.RawMessage(`not json`)
	new := json.RawMessage(`{"valid":true}`)

	result := CompareOutputs(old, new)
	if result.Type != DiffTypeBreaking {
		t.Errorf("expected breaking for invalid original JSON, got %s", result.Type)
	}
}

func TestCompareOutputsForError(t *testing.T) {
	old := json.RawMessage(`{"result":"ok"}`)
	result := CompareOutputsForError(old, "connection timeout")

	if result.Type != DiffTypeError {
		t.Errorf("expected error type, got %s", result.Type)
	}
	if !result.Changed {
		t.Error("expected Changed=true")
	}
	if result.ChangedFields != 1 {
		t.Errorf("expected 1 changed field, got %d", result.ChangedFields)
	}
}

func TestDiffSummaryToJSON(t *testing.T) {
	result := &DiffResult{
		Type:    DiffTypeMinor,
		Summary: "1 field changed",
		Changed: true,
	}

	summary, detail := DiffSummaryToJSON(result)
	if summary == "" {
		t.Error("expected non-empty summary")
	}
	if detail == nil {
		t.Error("expected non-nil detail")
	}

	var parsed map[string]interface{}
	if err := json.Unmarshal(detail, &parsed); err != nil {
		t.Errorf("detail should be valid JSON: %v", err)
	}
}

func TestClassifyDiffTypeFromDB(t *testing.T) {
	tests := []struct {
		input    string
		expected DiffType
	}{
		{"identical", DiffTypeIdentical},
		{"minor", DiffTypeMinor},
		{"major", DiffTypeMajor},
		{"breaking", DiffTypeBreaking},
		{"error", DiffTypeError},
		{"IDENTICAL", DiffTypeIdentical},
		{"", DiffTypeIdentical},
		{"unknown", DiffTypeIdentical},
	}

	for _, tt := range tests {
		result := ClassifyDiffTypeFromDB(tt.input)
		if result != tt.expected {
			t.Errorf("ClassifyDiffTypeFromDB(%q) = %q, want %q", tt.input, result, tt.expected)
		}
	}
}

func TestCountLeafFields(t *testing.T) {
	tests := []struct {
		name     string
		input    interface{}
		expected int
	}{
		{"nil", nil, 0},
		{"string", "hello", 1},
		{"number", 42, 1},
		{"flat map", map[string]interface{}{"a": 1, "b": 2}, 2},
		{"nested map", map[string]interface{}{"a": map[string]interface{}{"b": 1, "c": 2}}, 2},
		{"array", []interface{}{1, 2, 3}, 3},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := countLeafFields(tt.input)
			if result != tt.expected {
				t.Errorf("countLeafFields = %d, want %d", result, tt.expected)
			}
		})
	}
}

func TestNormalizeJSON(t *testing.T) {
	tests := []struct {
		name  string
		input json.RawMessage
	}{
		{"nil", nil},
		{"empty", json.RawMessage(``)},
		{"object", json.RawMessage(`{"b":2,"a":1}`)},
		{"array", json.RawMessage(`[3,1,2]}`)},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := normalizeJSON(tt.input)
			if tt.input == nil && result != nil {
				t.Error("expected nil for nil input")
			}
		})
	}
}

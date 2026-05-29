package wasm

import (
	"encoding/json"

	"github.com/google/uuid"
)

// ParseFabricIDFromInput extracts an optional default fabric ID from execution input JSON.
func ParseFabricIDFromInput(input []byte) uuid.UUID {
	if len(input) == 0 {
		return uuid.Nil
	}
	var payload map[string]interface{}
	if err := json.Unmarshal(input, &payload); err != nil {
		return uuid.Nil
	}
	for _, key := range []string{"_state_fabric_id", "state_fabric_id", "fabric_id", "fabricId"} {
		if raw, ok := payload[key]; ok {
			if s, ok := raw.(string); ok {
				if id, err := uuid.Parse(s); err == nil {
					return id
				}
			}
		}
	}
	return uuid.Nil
}

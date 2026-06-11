package a2a

import (
	"encoding/json"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestCanTransition(t *testing.T) {
	tests := []struct {
		from, to TaskState
		want     bool
	}{
		{StateSubmitted, StateWorking, true},
		{StateSubmitted, StateFailed, true},
		{StateSubmitted, StateCanceled, true},
		{StateSubmitted, StateCompleted, false},

		{StateWorking, StateInputRequired, true},
		{StateWorking, StateCompleted, true},
		{StateWorking, StateFailed, true},
		{StateWorking, StateCanceled, true},

		{StateInputRequired, StateWorking, true},
		{StateInputRequired, StateCompleted, false},

		{StateCompleted, StateWorking, false},
		{StateCompleted, StateFailed, false},
		{StateFailed, StateWorking, false},
		{StateCanceled, StateWorking, false},
	}

	for _, tt := range tests {
		got := CanTransition(tt.from, tt.to)
		assert.Equal(t, tt.want, got, "%s → %s", tt.from, tt.to)
	}
}

func TestParseTaskMessage_TextParts(t *testing.T) {
	msg := TaskMessage{
		Role: "user",
		Parts: []MessagePart{
			{Type: "text", Text: "hello world"},
		},
	}

	data, err := ParseTaskMessage(msg)
	assert.NoError(t, err)
	assert.Contains(t, string(data), "hello world")
}

func TestParseTaskMessage_DataPart(t *testing.T) {
	msg := TaskMessage{
		Role: "user",
		Parts: []MessagePart{
			{Type: "data", Data: json.RawMessage(`{"key":"value"}`)},
		},
	}

	data, err := ParseTaskMessage(msg)
	assert.NoError(t, err)
	assert.Equal(t, `{"key":"value"}`, string(data))
}

func TestParseTaskMessage_Empty(t *testing.T) {
	msg := TaskMessage{}

	data, err := ParseTaskMessage(msg)
	assert.NoError(t, err)
	assert.Equal(t, "{}", string(data))
}

func TestGenerateTaskID(t *testing.T) {
	id1 := GenerateTaskID()
	id2 := GenerateTaskID()
	assert.NotEmpty(t, id1)
	assert.NotEmpty(t, id2)
	assert.NotEqual(t, id1, id2)
}

func TestAgentCardInfo_ToAgentCard(t *testing.T) {
	info := AgentCardInfo{
		ID:              "test-agent",
		Name:            "Test Agent",
		Description:     "A test agent",
		URL:             "https://example.com/a2a",
		Version:         "1.0",
		ProtocolVersion: "0.3.0",
		Capabilities:    []string{"streaming"},
		Skills:          []byte(`[{"id":"execute","description":"Run functions"}]`),
		AuthSchemes:     []string{"bearer"},
		InputModes:      []string{"application/json"},
		OutputModes:     []string{"application/json"},
	}

	card := info.ToAgentCard()
	assert.Equal(t, "Test Agent", card.Name)
	assert.Equal(t, "0.3.0", card.ProtocolVersion)
	assert.Len(t, card.Skills, 1)
	assert.Equal(t, "execute", card.Skills[0].ID)
	assert.Contains(t, card.Capabilities, "streaming")
}

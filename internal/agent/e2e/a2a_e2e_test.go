package e2e

import (
	"context"
	"encoding/json"
	"testing"

	"github.com/functionfly/functionfly/internal/a2a"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// ============================================================
// A2A Protocol Tests
// ============================================================

func TestA2ATaskStateTransitions(t *testing.T) {
	t.Run("should allow valid state transitions", func(t *testing.T) {
		testCases := []struct {
			from, to a2a.TaskState
			valid    bool
		}{
			{a2a.StateSubmitted, a2a.StateWorking, true},
			{a2a.StateSubmitted, a2a.StateFailed, true},
			{a2a.StateSubmitted, a2a.StateCanceled, true},
			{a2a.StateWorking, a2a.StateCompleted, true},
			{a2a.StateWorking, a2a.StateFailed, true},
			{a2a.StateWorking, a2a.StateInputRequired, true},
			{a2a.StateWorking, a2a.StateCanceled, true},
			{a2a.StateInputRequired, a2a.StateWorking, true},
			{a2a.StateInputRequired, a2a.StateFailed, true},
		}

		for _, tc := range testCases {
			got := a2a.CanTransition(tc.from, tc.to)
			assert.Equal(t, tc.valid, got, "transition %s -> %s should be %v", tc.from, tc.to, tc.valid)
		}
	})

	t.Run("should reject invalid state transitions", func(t *testing.T) {
		testCases := []struct {
			from, to a2a.TaskState
		}{
			{a2a.StateSubmitted, a2a.StateCompleted},
			{a2a.StateInputRequired, a2a.StateCompleted},
			{a2a.StateCompleted, a2a.StateWorking},
			{a2a.StateFailed, a2a.StateWorking},
			{a2a.StateCanceled, a2a.StateWorking},
		}

		for _, tc := range testCases {
			got := a2a.CanTransition(tc.from, tc.to)
			assert.False(t, got, "transition %s -> %s should be invalid", tc.from, tc.to)
		}
	})

	t.Run("should parse task message with text parts", func(t *testing.T) {
		msg := a2a.TaskMessage{
			Role: "user",
			Parts: []a2a.MessagePart{
				{Type: "text", Text: "Hello, world!"},
			},
		}

		data, err := a2a.ParseTaskMessage(msg)
		require.NoError(t, err)
		assert.Contains(t, string(data), "Hello, world!")
	})

	t.Run("should parse task message with data parts", func(t *testing.T) {
		msg := a2a.TaskMessage{
			Role: "user",
			Parts: []a2a.MessagePart{
				{Type: "data", Data: []byte(`{"key":"value"}`)},
			},
		}

		data, err := a2a.ParseTaskMessage(msg)
		require.NoError(t, err)
		assert.Equal(t, `{"key":"value"}`, string(data))
	})

	t.Run("should return empty object for empty message", func(t *testing.T) {
		msg := a2a.TaskMessage{}

		data, err := a2a.ParseTaskMessage(msg)
		require.NoError(t, err)
		assert.Equal(t, "{}", string(data))
	})

	t.Run("should wrap multiple text parts into JSON", func(t *testing.T) {
		msg := a2a.TaskMessage{
			Role: "user",
			Parts: []a2a.MessagePart{
				{Type: "text", Text: "part1"},
				{Type: "text", Text: "part2"},
			},
		}

		data, err := a2a.ParseTaskMessage(msg)
		require.NoError(t, err)
		assert.Contains(t, string(data), "part1")
		assert.Contains(t, string(data), "part2")
	})

	t.Run("should generate unique task IDs", func(t *testing.T) {
		id1 := a2a.GenerateTaskID()
		id2 := a2a.GenerateTaskID()

		assert.NotEmpty(t, id1)
		assert.NotEmpty(t, id2)
		assert.NotEqual(t, id1, id2)
	})
}

func TestA2AAgentCard(t *testing.T) {
	t.Run("should convert AgentCardInfo to AgentCard", func(t *testing.T) {
		skillsJSON := `[{"id":"code_gen","description":"Generate code"}]`
		info := a2a.AgentCardInfo{
			ID:              "test-agent",
			Name:            "Test Agent",
			Description:     "A test agent for e2e",
			URL:             "https://example.com/a2a",
			Version:         "1.0",
			ProtocolVersion: "0.3.0",
			Capabilities:    []string{"streaming", "pushNotifications"},
			Skills:          json.RawMessage(skillsJSON),
			AuthSchemes:     []string{"bearer"},
			InputModes:      []string{"application/json"},
			OutputModes:     []string{"application/json"},
		}

		card := info.ToAgentCard()

		assert.Equal(t, "Test Agent", card.Name)
		assert.Equal(t, "0.3.0", card.ProtocolVersion)
		assert.Contains(t, card.Capabilities, "streaming")
		assert.Len(t, card.Skills, 1)
		assert.Equal(t, "code_gen", card.Skills[0].ID)
	})

	t.Run("should handle missing optional fields", func(t *testing.T) {
		info := a2a.AgentCardInfo{
			ID:              "minimal-agent",
			Name:            "Minimal Agent",
			ProtocolVersion: "0.3.0",
		}

		card := info.ToAgentCard()

		assert.Equal(t, "Minimal Agent", card.Name)
		assert.Empty(t, card.Capabilities)
	})
}

// MockTaskStore for testing task engine
type mockTaskStore struct {
	states map[string]a2a.TaskState
}

func (m *mockTaskStore) UpdateTaskState(ctx context.Context, publicID string, fromState, to a2a.TaskState) error {
	m.states[publicID] = to
	return nil
}

func (m *mockTaskStore) GetTaskState(ctx context.Context, publicID string) (a2a.TaskState, error) {
	if state, ok := m.states[publicID]; ok {
		return state, nil
	}
	return a2a.StateSubmitted, nil
}

func (m *mockTaskStore) SetTaskOutput(ctx context.Context, publicID string, output []byte) error {
	return nil
}

func TestA2ATaskEngine(t *testing.T) {
	t.Run("should transition task state", func(t *testing.T) {
		store := &mockTaskStore{states: make(map[string]a2a.TaskState)}
		engine := a2a.NewTaskEngine(store, nil)

		taskID := "task-123"
		store.states[taskID] = a2a.StateSubmitted

		err := engine.Transition(context.Background(), taskID, a2a.StateSubmitted, a2a.StateWorking)
		require.NoError(t, err)

		state, _ := store.GetTaskState(context.Background(), taskID)
		assert.Equal(t, a2a.StateWorking, state)
	})

	t.Run("should reject invalid transitions", func(t *testing.T) {
		store := &mockTaskStore{states: make(map[string]a2a.TaskState)}
		engine := a2a.NewTaskEngine(store, nil)

		taskID := "task-invalid"
		store.states[taskID] = a2a.StateCompleted

		err := engine.Transition(context.Background(), taskID, a2a.StateCompleted, a2a.StateWorking)
		assert.Error(t, err)
		assert.Contains(t, err.Error(), "invalid transition")
	})

	t.Run("should complete task with output", func(t *testing.T) {
		store := &mockTaskStore{states: make(map[string]a2a.TaskState)}
		engine := a2a.NewTaskEngine(store, nil)

		taskID := "task-complete"
		store.states[taskID] = a2a.StateWorking

		output := []byte(`{"result":"success"}`)
		err := engine.Complete(context.Background(), taskID, output)
		require.NoError(t, err)

		state, _ := store.GetTaskState(context.Background(), taskID)
		assert.Equal(t, a2a.StateCompleted, state)
	})

	t.Run("should fail task with error message", func(t *testing.T) {
		store := &mockTaskStore{states: make(map[string]a2a.TaskState)}
		engine := a2a.NewTaskEngine(store, nil)

		taskID := "task-fail"
		store.states[taskID] = a2a.StateWorking

		err := engine.Fail(context.Background(), taskID, "something went wrong")
		require.NoError(t, err)

		state, _ := store.GetTaskState(context.Background(), taskID)
		assert.Equal(t, a2a.StateFailed, state)
	})

	t.Run("should subscribe to task updates", func(t *testing.T) {
		store := &mockTaskStore{states: make(map[string]a2a.TaskState)}
		engine := a2a.NewTaskEngine(store, nil)

		taskID := "task-subscribe"
		store.states[taskID] = a2a.StateSubmitted

		ch, unsubscribe := engine.Subscribe(taskID)
		defer unsubscribe()

		assert.NotNil(t, ch)

		// Transition should notify subscribers
		err := engine.Transition(context.Background(), taskID, a2a.StateSubmitted, a2a.StateWorking)
		require.NoError(t, err)

		// Give time for notification
		// Note: in real implementation this would be async
	})

	t.Run("should unsubscribe from task updates", func(t *testing.T) {
		store := &mockTaskStore{states: make(map[string]a2a.TaskState)}
		engine := a2a.NewTaskEngine(store, nil)

		taskID := "task-unsubscribe"
		store.states[taskID] = a2a.StateSubmitted

		ch, unsubscribe := engine.Subscribe(taskID)

		// Unsubscribe immediately
		unsubscribe()

		// Channel should still exist but no subscribers
		// Transition should not panic
		err := engine.Transition(context.Background(), taskID, a2a.StateSubmitted, a2a.StateWorking)
		require.NoError(t, err)
		assert.NotNil(t, ch)
	})
}

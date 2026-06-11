package queue

import (
	"context"
	"encoding/json"
	"sync"
	"testing"
	"time"

	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
)

func TestExecutionPriority_String(t *testing.T) {
	tests := []struct {
		priority ExecutionPriority
		expected string
	}{
		{PriorityLow, "low"},
		{PriorityNormal, "normal"},
		{PriorityHigh, "high"},
		{PriorityCritical, "critical"},
		{ExecutionPriority(100), "unknown"},
	}

	for _, tt := range tests {
		t.Run(tt.expected, func(t *testing.T) {
			assert.Equal(t, tt.expected, tt.priority.String())
		})
	}
}

func TestNewExecutionQueue(t *testing.T) {
	queue := NewExecutionQueue(100, 5)

	assert.NotNil(t, queue)
	assert.Equal(t, 100, queue.maxQueueSize)
	assert.Equal(t, 5, queue.maxWorkers)
	assert.NotNil(t, queue.queue)
	assert.NotNil(t, queue.processing)
	assert.Empty(t, queue.queue)
	assert.Empty(t, queue.workers)
}

func TestNewExecutionQueue_DefaultValues(t *testing.T) {
	queue := NewExecutionQueue(0, 0)

	assert.Equal(t, 1000, queue.maxQueueSize)
	assert.Equal(t, 10, queue.maxWorkers)
}

func TestNewExecutionQueue_NegativeValues(t *testing.T) {
	queue := NewExecutionQueue(-10, -5)

	assert.Equal(t, 1000, queue.maxQueueSize)
	assert.Equal(t, 10, queue.maxWorkers)
}

func TestExecutionQueue_Enqueue(t *testing.T) {
	queue := NewExecutionQueue(100, 0)

	exec := &QueuedExecution{
		FunctionID: uuid.New(),
		Priority:   PriorityNormal,
	}

	err := queue.Enqueue(context.Background(), exec)
	assert.NoError(t, err)
	assert.Len(t, queue.queue, 1)
	assert.NotEmpty(t, exec.ID)
	assert.False(t, exec.QueuedAt.IsZero())
}

func TestExecutionQueue_Enqueue_DefaultPriority(t *testing.T) {
	queue := NewExecutionQueue(100, 0)

	exec := &QueuedExecution{
		FunctionID: uuid.New(),
	}

	err := queue.Enqueue(context.Background(), exec)
	assert.NoError(t, err)
	assert.Equal(t, PriorityNormal, exec.Priority)
}

func TestExecutionQueue_Enqueue_DefaultRetries(t *testing.T) {
	queue := NewExecutionQueue(100, 0)

	exec := &QueuedExecution{
		FunctionID: uuid.New(),
	}

	err := queue.Enqueue(context.Background(), exec)
	assert.NoError(t, err)
	assert.Equal(t, 3, exec.MaxRetries)
}

func TestExecutionQueue_Enqueue_QueueFull(t *testing.T) {
	queue := NewExecutionQueue(2, 0)

	exec1 := &QueuedExecution{FunctionID: uuid.New()}
	exec2 := &QueuedExecution{FunctionID: uuid.New()}
	exec3 := &QueuedExecution{FunctionID: uuid.New()}

	err := queue.Enqueue(context.Background(), exec1)
	assert.NoError(t, err)

	err = queue.Enqueue(context.Background(), exec2)
	assert.NoError(t, err)

	err = queue.Enqueue(context.Background(), exec3)
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "queue is full")
}

func TestExecutionQueue_Enqueue_PriorityOrdering(t *testing.T) {
	queue := NewExecutionQueue(100, 0)

	exec1 := &QueuedExecution{FunctionID: uuid.New(), Priority: PriorityLow}
	exec2 := &QueuedExecution{FunctionID: uuid.New(), Priority: PriorityHigh}
	exec3 := &QueuedExecution{FunctionID: uuid.New(), Priority: PriorityNormal}

	queue.Enqueue(context.Background(), exec1)
	queue.Enqueue(context.Background(), exec2)
	queue.Enqueue(context.Background(), exec3)

	assert.Equal(t, PriorityHigh, queue.queue[0].Priority)
	assert.Equal(t, PriorityNormal, queue.queue[1].Priority)
	assert.Equal(t, PriorityLow, queue.queue[2].Priority)
}

func TestExecutionQueue_Requeue(t *testing.T) {
	queue := NewExecutionQueue(100, 0)

	exec := &QueuedExecution{
		ID:         "exec_1",
		FunctionID: uuid.New(),
		MaxRetries: 3,
		RetryCount: 0,
	}

	queue.Enqueue(context.Background(), exec)

	queue.Requeue(exec, "test error")

	assert.Equal(t, 1, exec.RetryCount)
	assert.Equal(t, "test error", exec.ErrorMessage)
	assert.Equal(t, PriorityLow, exec.Priority)
}

func TestExecutionQueue_Requeue_MaxRetriesExceeded(t *testing.T) {
	queue := NewExecutionQueue(100, 0)

	exec := &QueuedExecution{
		ID:         "exec_1",
		FunctionID: uuid.New(),
		MaxRetries: 3,
		RetryCount: 3,
	}

	queue.Enqueue(context.Background(), exec)

	queue.Requeue(exec, "test error")

	assert.Len(t, queue.queue, 0)
}

func TestExecutionQueue_Requeue_DoesNotAddDuplicate(t *testing.T) {
	queue := NewExecutionQueue(100, 0)

	exec := &QueuedExecution{
		ID:         "exec_1",
		FunctionID: uuid.New(),
		MaxRetries: 3,
		RetryCount: 0,
	}

	queue.Enqueue(context.Background(), exec)
	queue.Requeue(exec, "test error")

	// Requeue should move it, not duplicate
	assert.Len(t, queue.queue, 1)
}

func TestExecutionQueue_MarkCompleted(t *testing.T) {
	queue := NewExecutionQueue(100, 0)

	queue.processing["exec_1"] = true
	queue.processing["exec_2"] = true

	queue.MarkCompleted("exec_1")

	assert.False(t, queue.processing["exec_1"])
	assert.True(t, queue.processing["exec_2"])
}

func TestExecutionQueue_GetStats(t *testing.T) {
	queue := NewExecutionQueue(100, 5)

	queue.Enqueue(context.Background(), &QueuedExecution{FunctionID: uuid.New(), Priority: PriorityHigh})
	queue.Enqueue(context.Background(), &QueuedExecution{FunctionID: uuid.New(), Priority: PriorityHigh})
	queue.Enqueue(context.Background(), &QueuedExecution{FunctionID: uuid.New(), Priority: PriorityNormal})
	queue.Enqueue(context.Background(), &QueuedExecution{FunctionID: uuid.New(), Priority: PriorityLow})

	queue.processing["exec_1"] = true
	queue.processing["exec_2"] = true

	stats := queue.GetStats()

	assert.Equal(t, 4, stats["queue_size"])
	assert.Equal(t, 100, stats["max_queue_size"])
	assert.Equal(t, 2, stats["processing"])
	assert.Equal(t, 5, stats["workers"])

	priorityCounts := stats["priority_counts"].(map[string]int)
	assert.Equal(t, 2, priorityCounts["high"])
	assert.Equal(t, 1, priorityCounts["normal"])
	assert.Equal(t, 1, priorityCounts["low"])
}

func TestExecutionQueue_StartStop(t *testing.T) {
	queue := NewExecutionQueue(100, 2)

	execCalled := false
	onExec := func(exec *QueuedExecution) error {
		execCalled = true
		return nil
	}

	queue.Start(onExec)
	time.Sleep(100 * time.Millisecond)

	assert.Len(t, queue.workers, 2)
	assert.NotNil(t, queue.onExecution)

	queue.Stop()

	time.Sleep(100 * time.Millisecond)
}

func TestExecutionQueue_ProcessExecution_Success(t *testing.T) {
	queue := NewExecutionQueue(100, 0)

	exec := &QueuedExecution{
		ID:         "exec_1",
		FunctionID: uuid.New(),
	}

	var mu sync.Mutex
	executed := false

	onExec := func(e *QueuedExecution) error {
		mu.Lock()
		executed = true
		mu.Unlock()
		return nil
	}

	queue.Start(onExec)
	queue.processExecution(exec)

	mu.Lock()
	assert.True(t, executed)
	mu.Unlock()

	assert.False(t, queue.processing["exec_1"])
}

func TestExecutionQueue_ProcessExecution_Error(t *testing.T) {
	queue := NewExecutionQueue(100, 0)

	exec := &QueuedExecution{
		ID:         "exec_1",
		FunctionID: uuid.New(),
		MaxRetries: 3,
		RetryCount: 0,
	}

	onExec := func(e *QueuedExecution) error {
		return assert.AnError
	}

	queue.Start(onExec)
	queue.processExecution(exec)

	// Should be requeued
	assert.True(t, queue.processing["exec_1"])
}

func TestCreateQueuedExecution(t *testing.T) {
	functionID := uuid.New()
	userID := uuid.New()
	tenantID := uuid.New()

	exec := CreateQueuedExecution(
		functionID,
		"1.0.0",
		json.RawMessage(`{"key":"value"}`),
		PriorityHigh,
		&userID,
		&tenantID,
		"req_123",
		"192.168.1.1",
		"Mozilla/5.0",
		30000,
	)

	assert.Equal(t, functionID, exec.FunctionID)
	assert.Equal(t, "1.0.0", exec.Version)
	assert.NotNil(t, exec.Input)
	assert.Equal(t, PriorityHigh, exec.Priority)
	assert.Equal(t, &userID, exec.UserID)
	assert.Equal(t, &tenantID, exec.TenantID)
	assert.Equal(t, "req_123", exec.RequestID)
	assert.Equal(t, "192.168.1.1", exec.IPAddress)
	assert.Equal(t, "Mozilla/5.0", exec.UserAgent)
	assert.Equal(t, 30000, exec.TimeoutMs)
}

func TestQueuedExecution_JSON(t *testing.T) {
	functionID := uuid.New()
	userID := uuid.New()

	exec := &QueuedExecution{
		ID:         "exec_123",
		FunctionID: functionID,
		Version:    "1.0.0",
		Input:      json.RawMessage(`{"test":true}`),
		Priority:   PriorityHigh,
		UserID:     &userID,
		RequestID:  "req_456",
		QueuedAt:   time.Now(),
		TimeoutMs:  30000,
		MaxRetries: 3,
		RetryCount: 0,
	}

	data, err := json.Marshal(exec)
	assert.NoError(t, err)

	var result QueuedExecution
	err = json.Unmarshal(data, &result)
	assert.NoError(t, err)

	assert.Equal(t, exec.ID, result.ID)
	assert.Equal(t, exec.FunctionID, result.FunctionID)
	assert.Equal(t, exec.Version, result.Version)
	assert.Equal(t, exec.Priority, result.Priority)
	assert.Equal(t, exec.RequestID, result.RequestID)
	assert.Equal(t, exec.TimeoutMs, result.TimeoutMs)
}

func TestExecutionQueue_Dequeue_Empty(t *testing.T) {
	queue := NewExecutionQueue(100, 0)

	exec := queue.dequeue()

	assert.Nil(t, exec)
}

func TestExecutionQueue_Dequeue_MovesToProcessing(t *testing.T) {
	queue := NewExecutionQueue(100, 0)

	exec := &QueuedExecution{
		ID:         "exec_1",
		FunctionID: uuid.New(),
	}

	queue.Enqueue(context.Background(), exec)

	dequeued := queue.dequeue()

	assert.NotNil(t, dequeued)
	assert.Equal(t, "exec_1", dequeued.ID)
	assert.True(t, queue.processing["exec_1"])
}

func TestExecutionQueue_Dequeue_AlreadyProcessing(t *testing.T) {
	queue := NewExecutionQueue(100, 0)

	exec := &QueuedExecution{
		ID:         "exec_1",
		FunctionID: uuid.New(),
	}

	queue.Enqueue(context.Background(), exec)
	queue.processing["exec_1"] = true

	dequeued := queue.dequeue()

	assert.Nil(t, dequeued)
}

func TestExecutionQueue_Stop_NoWorkers(t *testing.T) {
	queue := NewExecutionQueue(100, 0)
	queue.Stop()
}

func TestQueuedExecution_WithOptionalFields(t *testing.T) {
	exec := &QueuedExecution{
		ID:            "exec_123",
		FunctionID:    uuid.New(),
		Version:       "1.0.0",
		Priority:      PriorityCritical,
		MaxRetries:    5,
		RetryCount:    2,
		LastAttemptAt: func() *time.Time { t := time.Now(); return &t }(),
		ErrorMessage:  "previous error",
		IPAddress:     "10.0.0.1",
		UserAgent:     "TestAgent/1.0",
	}

	assert.Equal(t, "exec_123", exec.ID)
	assert.Equal(t, PriorityCritical, exec.Priority)
	assert.Equal(t, 5, exec.MaxRetries)
	assert.Equal(t, 2, exec.RetryCount)
	assert.NotNil(t, exec.LastAttemptAt)
	assert.Equal(t, "previous error", exec.ErrorMessage)
	assert.Equal(t, "10.0.0.1", exec.IPAddress)
	assert.Equal(t, "TestAgent/1.0", exec.UserAgent)
}

func TestExecutionQueue_MultipleEnqueueDequeue(t *testing.T) {
	queue := NewExecutionQueue(100, 0)

	exec1 := &QueuedExecution{ID: "exec_1", FunctionID: uuid.New(), Priority: PriorityLow}
	exec2 := &QueuedExecution{ID: "exec_2", FunctionID: uuid.New(), Priority: PriorityHigh}
	exec3 := &QueuedExecution{ID: "exec_3", FunctionID: uuid.New(), Priority: PriorityNormal}

	queue.Enqueue(context.Background(), exec1)
	queue.Enqueue(context.Background(), exec2)
	queue.Enqueue(context.Background(), exec3)

	assert.Len(t, queue.queue, 3)

	d1 := queue.dequeue()
	assert.Equal(t, "exec_2", d1.ID)

	d2 := queue.dequeue()
	assert.Equal(t, "exec_3", d2.ID)

	d3 := queue.dequeue()
	assert.Equal(t, "exec_1", d3.ID)

	d4 := queue.dequeue()
	assert.Nil(t, d4)
}

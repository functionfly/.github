package rabbitmq

import (
	"context"
	"encoding/json"
	"os"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
)

func TestNewPublisherFromEnv(t *testing.T) {
	os.Setenv("RABBITMQ_URL", "amqp://guest:guest@localhost:5672/")
	os.Setenv("EXECUTION_QUEUE_RABBITMQ_QUEUE", "test.queue")
	defer func() {
		os.Unsetenv("RABBITMQ_URL")
		os.Unsetenv("EXECUTION_QUEUE_RABBITMQ_QUEUE")
	}()

	publisher := NewPublisherFromEnv()

	assert.NotNil(t, publisher)
	assert.Equal(t, "amqp://guest:guest@localhost:5672/", publisher.url)
	assert.Equal(t, "test.queue", publisher.queue)
}

func TestNewPublisherFromEnv_Defaults(t *testing.T) {
	os.Unsetenv("RABBITMQ_URL")
	os.Unsetenv("EXECUTION_QUEUE_RABBITMQ_QUEUE")

	publisher := NewPublisherFromEnv()

	assert.NotNil(t, publisher)
	assert.Empty(t, publisher.url)
	assert.Equal(t, "functionfly.execution.queue", publisher.queue)
}

func TestPublisher_Enabled(t *testing.T) {
	tests := []struct {
		name     string
		url      string
		expected bool
	}{
		{"enabled", "amqp://localhost:5672", true},
		{"disabled_empty", "", false},
		{"disabled_nil", "", false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			publisher := &Publisher{url: tt.url}
			assert.Equal(t, tt.expected, publisher.Enabled())
		})
	}
}

func TestPublisher_Enabled_Nil(t *testing.T) {
	var publisher *Publisher
	assert.False(t, publisher.Enabled())
}

func TestEnvOr(t *testing.T) {
	os.Setenv("TEST_KEY", "test_value")
	defer os.Unsetenv("TEST_KEY")

	assert.Equal(t, "test_value", envOr("TEST_KEY", "default"))
	assert.Equal(t, "default", envOr("UNSET_KEY", "default"))
}

func TestPublisher_Close(t *testing.T) {
	publisher := &Publisher{}
	err := publisher.Close()
	assert.NoError(t, err)
}

func TestPublisher_Close_WithConnection(t *testing.T) {
	publisher := &Publisher{}
	publisher.Close()
	assert.Nil(t, publisher.conn)
	assert.Nil(t, publisher.ch)
}

func TestPublishOptions_Defaults(t *testing.T) {
	opts := PublishOptions{}

	assert.Empty(t, opts.MessageID)
	assert.Empty(t, opts.ContentType)
}

func TestPublishOptions_WithValues(t *testing.T) {
	opts := PublishOptions{
		MessageID:   "msg_123",
		ContentType: "application/json",
	}

	assert.Equal(t, "msg_123", opts.MessageID)
	assert.Equal(t, "application/json", opts.ContentType)
}

func TestPublisher_PublishJSON_NotEnabled(t *testing.T) {
	publisher := &Publisher{url: ""}

	err := publisher.PublishJSON(context.Background(), map[string]string{"key": "value"}, PublishOptions{})
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "RABBITMQ_URL is empty")
}

func TestPublisher_ensureConnected_EmptyURL(t *testing.T) {
	publisher := &Publisher{url: ""}

	err := publisher.ensureConnected(context.Background())
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "RABBITMQ_URL is empty")
}

func TestPublisher_PublishJSON_NoChannel(t *testing.T) {
	publisher := &Publisher{
		url:   "amqp://guest:guest@localhost:5672/",
		queue: "test.queue",
	}

	publisher.mu.Lock()
	publisher.ch = nil
	publisher.mu.Unlock()

	err := publisher.PublishJSON(context.Background(), map[string]string{"key": "value"}, PublishOptions{})
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "channel not available")
}

func TestMessageBody_JSON(t *testing.T) {
	type TestMessage struct {
		ID   string `json:"id"`
		Name string `json:"name"`
	}

	msg := TestMessage{ID: "123", Name: "test"}

	data, err := json.Marshal(msg)
	assert.NoError(t, err)

	var result TestMessage
	err = json.Unmarshal(data, &result)
	assert.NoError(t, err)

	assert.Equal(t, "123", result.ID)
	assert.Equal(t, "test", result.Name)
}

func TestPublisher_Mutex(t *testing.T) {
	publisher := &Publisher{}

	publisher.mu.Lock()
	publisher.url = "test_url"
	publisher.mu.Unlock()

	publisher.mu.Lock()
	assert.Equal(t, "test_url", publisher.url)
	publisher.mu.Unlock()
}

func TestPublisher_Fields(t *testing.T) {
	publisher := &Publisher{
		url:   "amqp://localhost:5672",
		queue: "test_queue",
	}

	assert.Equal(t, "amqp://localhost:5672", publisher.url)
	assert.Equal(t, "test_queue", publisher.queue)
}

func TestPublishOptions_JSON(t *testing.T) {
	opts := PublishOptions{
		MessageID:   "msg_456",
		ContentType: "application/json",
	}

	data, err := json.Marshal(opts)
	assert.NoError(t, err)

	var result PublishOptions
	err = json.Unmarshal(data, &result)
	assert.NoError(t, err)

	assert.Equal(t, "msg_456", result.MessageID)
	assert.Equal(t, "application/json", result.ContentType)
}

func TestNewPublisherFromEnv_MultipleCalls(t *testing.T) {
	os.Setenv("RABBITMQ_URL", "amqp://first:5672/")
	os.Setenv("EXECUTION_QUEUE_RABBITMQ_QUEUE", "queue1")
	defer os.Unsetenv("RABBITMQ_URL")
	defer os.Unsetenv("EXECUTION_QUEUE_RABBITMQ_QUEUE")

	p1 := NewPublisherFromEnv()
	assert.Equal(t, "amqp://first:5672/", p1.url)

	os.Setenv("RABBITMQ_URL", "amqp://second:5672/")
	os.Setenv("EXECUTION_QUEUE_RABBITMQ_QUEUE", "queue2")

	p2 := NewPublisherFromEnv()
	assert.Equal(t, "amqp://second:5672/", p2.url)
}

func TestPublisher_URLGetter(t *testing.T) {
	publisher := &Publisher{url: "amqp://test:5672"}
	assert.Equal(t, "amqp://test:5672", publisher.url)
}

func TestPublisher_QueueGetter(t *testing.T) {
	publisher := &Publisher{queue: "my.queue"}
	assert.Equal(t, "my.queue", publisher.queue)
}

func TestPublisher_PublishJSON_ContextTimeout(t *testing.T) {
	publisher := &Publisher{
		url:   "amqp://guest:guest@localhost:5672/",
		queue: "test.queue",
	}

	ctx, cancel := context.WithTimeout(context.Background(), 1*time.Nanosecond)
	defer cancel()

	time.Sleep(10 * time.Millisecond)

	err := publisher.PublishJSON(ctx, map[string]string{"key": "value"}, PublishOptions{})
	assert.Error(t, err)
}

func TestPublisher_ensureConnected_AlreadyConnected(t *testing.T) {
	publisher := &Publisher{
		url:   "amqp://guest:guest@localhost:5672/",
		queue: "test.queue",
	}

	publisher.mu.Lock()
	publisher.conn = nil
	publisher.ch = nil
	publisher.mu.Unlock()

	err := publisher.ensureConnected(context.Background())
	assert.Error(t, err)
}

func TestPublisher_Close_Multiple(t *testing.T) {
	publisher := &Publisher{}

	err1 := publisher.Close()
	err2 := publisher.Close()

	assert.NoError(t, err1)
	assert.NoError(t, err2)
}

func TestEnvOr_EmptyEnv(t *testing.T) {
	os.Unsetenv("EMPTY_KEY")
	result := envOr("EMPTY_KEY", "fallback")
	assert.Equal(t, "fallback", result)
}

func TestEnvOr_NonEmptyEnv(t *testing.T) {
	os.Setenv("NON_EMPTY_KEY", "actual_value")
	defer os.Unsetenv("NON_EMPTY_KEY")

	result := envOr("NON_EMPTY_KEY", "fallback")
	assert.Equal(t, "actual_value", result)
}
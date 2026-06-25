package rabbitmq

import (
	"context"
	"encoding/json"
	"fmt"
	"os"
	"sync"
	"time"

	"github.com/google/uuid"
	"github.com/sirupsen/logrus"

	amqp "github.com/rabbitmq/amqp091-go"
)

type ExecutionHandler func(ctx context.Context, msg *QueuedExecutionMessage) error

type QueuedExecutionMessage struct {
	Type           string          `json:"type"`
	FunctionID     string          `json:"function_id"`
	Author         string          `json:"author"`
	Name           string          `json:"name"`
	Version        string          `json:"version"`
	InputJSON      json.RawMessage `json:"input_json"`
	RequestedAt    string          `json:"requested_at"`
	RequestID      string          `json:"request_id"`
	UserAgent      string          `json:"user_agent"`
	QueueReason    string          `json:"queue_reason"`
	CacheEligible  bool            `json:"cache_eligible"`
}

type Consumer struct {
	url      string
	queue    string
	workers  int
	handler  ExecutionHandler
	stopChan chan struct{}
	doneChan chan struct{}
	mu       sync.Mutex
	conn     *amqp.Connection
	ch       *amqp.Channel
	logger   *logrus.Logger
}

func NewConsumerFromEnv(handler ExecutionHandler) *Consumer {
	workers := 5
	if w := os.Getenv("EXECUTION_QUEUE_WORKERS"); w != "" {
		if parsed, err := fmt.Sscanf(w, "%d", &workers); parsed == 1 && err == nil && workers > 0 {
			if workers > 50 {
				workers = 50
			}
		}
	}

	return &Consumer{
		url:      os.Getenv("RABBITMQ_URL"),
		queue:    envOr("EXECUTION_QUEUE_RABBITMQ_QUEUE", "functionfly.execution.queue"),
		workers:  workers,
		handler:  handler,
		stopChan: make(chan struct{}),
		doneChan: make(chan struct{}),
		logger:   logrus.WithField("component", "execution_queue_consumer").Logger,
	}
}

func (c *Consumer) Enabled() bool {
	return c != nil && c.url != ""
}

func (c *Consumer) ensureConnected(ctx context.Context) error {
	c.mu.Lock()
	defer c.mu.Unlock()

	if c.conn != nil && !c.conn.IsClosed() && c.ch != nil {
		return nil
	}

	if c.url == "" {
		return fmt.Errorf("RABBITMQ_URL is empty")
	}

	conn, err := amqp.Dial(c.url)
	if err != nil {
		return fmt.Errorf("dial rabbitmq: %w", err)
	}

	ch, err := conn.Channel()
	if err != nil {
		_ = conn.Close()
		return fmt.Errorf("open channel: %w", err)
	}

	_, err = ch.QueueDeclare(
		c.queue,
		true,  // durable
		false, // autoDelete
		false, // exclusive
		false, // noWait
		nil,   // args
	)
	if err != nil {
		_ = ch.Close()
		_ = conn.Close()
		return fmt.Errorf("declare queue: %w", err)
	}

	err = ch.Qos(
		c.workers, // prefetch count - process N messages concurrently
		0,         // prefetch size
		false,     // global
	)
	if err != nil {
		c.logger.WithError(err).Warn("failed to set QoS, using defaults")
	}

	c.conn = conn
	c.ch = ch
	return nil
}

func (c *Consumer) Start(ctx context.Context) error {
	if !c.Enabled() {
		return fmt.Errorf("consumer not enabled: RABBITMQ_URL not configured")
	}

	if err := c.ensureConnected(ctx); err != nil {
		return fmt.Errorf("failed to connect to rabbitmq: %w", err)
	}

	c.logger.WithField("workers", c.workers).Info("Starting execution queue consumer")

	msgs, err := c.ch.Consume(
		c.queue,
		"",    // consumer tag
		false, // auto-ack
		false, // exclusive
		false, // no-local
		false, // no-wait
		nil,   // args
	)
	if err != nil {
		return fmt.Errorf("failed to start consuming: %w", err)
	}

	go c.processMessages(ctx, msgs)

	return nil
}

func (c *Consumer) processMessages(ctx context.Context, msgs <-chan amqp.Delivery) {
	defer close(c.doneChan)

	for {
		select {
		case <-ctx.Done():
			c.logger.Info("context cancelled, stopping message processing")
			return
		case <-c.stopChan:
			c.logger.Info("stop signal received, stopping message processing")
			return
		case msg, ok := <-msgs:
			if !ok {
				c.logger.Warn("message channel closed")
				return
			}
			c.handleMessage(ctx, msg)
		}
	}
}

func (c *Consumer) handleMessage(ctx context.Context, msg amqp.Delivery) {
	var execMsg QueuedExecutionMessage
	if err := json.Unmarshal(msg.Body, &execMsg); err != nil {
		c.logger.WithError(err).WithField("body", string(msg.Body)).Error("failed to unmarshal execution message")
		msg.Nack(false, false)
		return
	}

	c.logger.WithFields(logrus.Fields{
		"function_id": execMsg.FunctionID,
		"author":     execMsg.Author,
		"name":       execMsg.Name,
		"version":    execMsg.Version,
		"request_id": execMsg.RequestID,
	}).Debug("Processing queued execution")

	if err := c.handler(ctx, &execMsg); err != nil {
		c.logger.WithError(err).WithFields(logrus.Fields{
			"function_id": execMsg.FunctionID,
			"author":     execMsg.Author,
			"name":       execMsg.Name,
			"request_id": execMsg.RequestID,
		}).Error("failed to process queued execution")

		requeue := shouldRequeue(&execMsg, err)
		msg.Nack(false, requeue)
		return
	}

	msg.Ack(false)
	c.logger.WithFields(logrus.Fields{
		"function_id": execMsg.FunctionID,
		"author":     execMsg.Author,
		"name":       execMsg.Name,
		"request_id": execMsg.RequestID,
	}).Debug("Successfully processed queued execution")
}

func shouldRequeue(msg *QueuedExecutionMessage, err error) bool {
	return false
}

func (c *Consumer) Stop() error {
	c.mu.Lock()
	defer c.mu.Unlock()

	close(c.stopChan)

	select {
	case <-c.doneChan:
	case <-time.After(10 * time.Second):
		c.logger.Warn("stop timed out waiting for message processing to finish")
	}

	if c.ch != nil {
		_ = c.ch.Close()
		c.ch = nil
	}
	if c.conn != nil {
		_ = c.conn.Close()
		c.conn = nil
	}

	c.logger.Info("execution queue consumer stopped")
	return nil
}

type QueuedExecution struct {
	ID            string
	FunctionID    uuid.UUID
	Version       string
	Input         json.RawMessage
	RequestID     string
	QueuedAt      time.Time
	MaxRetries    int
	RetryCount    int
	LastAttemptAt *time.Time
	ErrorMessage  string
	IPAddress     string
	UserAgent     string
}

func (m *QueuedExecutionMessage) ToQueuedExecution() *QueuedExecution {
	exec := &QueuedExecution{
		ID:         m.RequestID,
		FunctionID: uuid.MustParse(m.FunctionID),
		Version:    m.Version,
		Input:      m.InputJSON,
		RequestID:  m.RequestID,
		UserAgent:  m.UserAgent,
		MaxRetries: 3,
	}

	if m.RequestedAt != "" {
		if t, err := time.Parse(time.RFC3339Nano, m.RequestedAt); err == nil {
			exec.QueuedAt = t
		}
	}

	return exec
}

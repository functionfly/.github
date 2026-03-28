package rabbitmq

import (
	"context"
	"encoding/json"
	"fmt"
	"os"
	"sync"
	"time"

	amqp "github.com/rabbitmq/amqp091-go"
)

type Publisher struct {
	url   string
	queue string

	mu   sync.Mutex
	conn *amqp.Connection
	ch   *amqp.Channel
}

type PublishOptions struct {
	MessageID   string
	ContentType string
}

func NewPublisherFromEnv() *Publisher {
	return &Publisher{
		url:   os.Getenv("RABBITMQ_URL"),
		queue: envOr("EXECUTION_QUEUE_RABBITMQ_QUEUE", "functionfly.execution.queue"),
	}
}

func (p *Publisher) Enabled() bool {
	return p != nil && p.url != ""
}

func envOr(key, def string) string {
	v := os.Getenv(key)
	if v == "" {
		return def
	}
	return v
}

func (p *Publisher) ensureConnected(ctx context.Context) error {
	p.mu.Lock()
	defer p.mu.Unlock()

	if p.conn != nil && !p.conn.IsClosed() && p.ch != nil {
		return nil
	}

	if p.url == "" {
		return fmt.Errorf("RABBITMQ_URL is empty")
	}

	conn, err := amqp.Dial(p.url)
	if err != nil {
		return fmt.Errorf("dial rabbitmq: %w", err)
	}
	ch, err := conn.Channel()
	if err != nil {
		_ = conn.Close()
		return fmt.Errorf("open channel: %w", err)
	}

	// Durable queue so messages survive broker restarts.
	_, err = ch.QueueDeclare(
		p.queue,
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

	p.conn = conn
	p.ch = ch
	return nil
}

func (p *Publisher) Close() error {
	p.mu.Lock()
	defer p.mu.Unlock()
	if p.ch != nil {
		_ = p.ch.Close()
		p.ch = nil
	}
	if p.conn != nil {
		_ = p.conn.Close()
		p.conn = nil
	}
	return nil
}

func (p *Publisher) PublishJSON(ctx context.Context, body any, opts PublishOptions) error {
	if err := p.ensureConnected(ctx); err != nil {
		return err
	}

	b, err := json.Marshal(body)
	if err != nil {
		return fmt.Errorf("marshal json: %w", err)
	}
	if opts.ContentType == "" {
		opts.ContentType = "application/json"
	}

	// Use a short timeout so execution requests don’t hang on broker issues.
	pubCtx, cancel := context.WithTimeout(ctx, 2*time.Second)
	defer cancel()

	p.mu.Lock()
	ch := p.ch
	p.mu.Unlock()
	if ch == nil {
		return fmt.Errorf("rabbitmq channel not available")
	}

	return ch.PublishWithContext(
		pubCtx,
		"",      // default exchange
		p.queue, // routing key
		false,   // mandatory
		false,   // immediate
		amqp.Publishing{
			MessageId:   opts.MessageID,
			Timestamp:   time.Now(),
			ContentType: opts.ContentType,
			DeliveryMode: amqp.Persistent, // survive broker restart (with durable queue)
			Body:        b,
		},
	)
}


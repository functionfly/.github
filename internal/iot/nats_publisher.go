package iot

import (
	"context"
	"fmt"
	"strings"
	"sync"

	"github.com/google/uuid"
	"github.com/sirupsen/logrus"
)

type NATSPublisher interface {
	Publish(ctx context.Context, subject string, payload []byte) error
	Close() error
	IsConnected() bool
}

type NATSClient struct {
	url    string
	logger *logrus.Logger
	mu     sync.RWMutex
	closed bool
}

func NewNATSClient(url string, logger *logrus.Logger) *NATSClient {
	if logger == nil {
		logger = logrus.New()
	}
	return &NATSClient{url: url, logger: logger}
}

func (c *NATSClient) Publish(ctx context.Context, subject string, payload []byte) error {
	c.mu.RLock()
	closed := c.closed
	c.mu.RUnlock()
	if closed {
		return ErrPublisherClosed
	}

	if !isValidSubject(subject) {
		return fmt.Errorf("invalid NATS subject: %s", subject)
	}

	c.logger.WithFields(logrus.Fields{
		"subject": subject,
		"bytes":   len(payload),
	}).Debug("publishing to NATS (stub)")
	return nil
}

func (c *NATSClient) Close() error {
	c.mu.Lock()
	defer c.mu.Unlock()
	if c.closed {
		return nil
	}
	c.closed = true
	return nil
}

func (c *NATSClient) IsConnected() bool {
	c.mu.RLock()
	defer c.mu.RUnlock()
	return !c.closed
}

func (c *NATSClient) URL() string { return c.url }

func isValidSubject(s string) bool {
	if s == "" || len(s) > 1024 {
		return false
	}
	for _, r := range s {
		if r == ' ' || r == '\t' || r == '\n' {
			return false
		}
	}
	return true
}

const (
	SubjectTelemetry = "iot.device.%s.telemetry"
	SubjectCommand   = "iot.device.%s.command"
	SubjectStatus    = "iot.device.%s.status"
	SubjectError     = "iot.device.%s.error"
	SubjectState     = "iot.device.%s.state"
	SubjectObserve   = "iot.device.%s.observe"
)

func SubjectForEvent(deviceID uuid.UUID, eventType string) string {
	id := deviceID.String()
	switch strings.ToLower(eventType) {
	case "telemetry":
		return fmt.Sprintf(SubjectTelemetry, id)
	case "command", "commands":
		return fmt.Sprintf(SubjectCommand, id)
	case "status":
		return fmt.Sprintf(SubjectStatus, id)
	case "error", "errors":
		return fmt.Sprintf(SubjectError, id)
	case "state":
		return fmt.Sprintf(SubjectState, id)
	case "observe":
		return fmt.Sprintf(SubjectObserve, id)
	default:
		return fmt.Sprintf("iot.device.%s.%s", id, eventType)
	}
}

var ErrPublisherClosed = fmt.Errorf("NATS publisher is closed")

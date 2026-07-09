package iot

import (
	"context"
	"encoding/json"
	"fmt"
	"strings"
	"sync"
	"time"

	"github.com/google/uuid"
	"github.com/sirupsen/logrus"
)

type Bridge struct {
	config    *IoTConfig
	publisher NATSPublisher
	registry  DeviceRegistry
	logger    *logrus.Logger
	mu        sync.RWMutex
	started   bool
	stopCh    chan struct{}
}

func NewBridge(config *IoTConfig, publisher NATSPublisher, registry DeviceRegistry, logger *logrus.Logger) (*Bridge, error) {
	if config == nil {
		return nil, fmt.Errorf("config is required")
	}
	if publisher == nil {
		return nil, fmt.Errorf("publisher is required")
	}
	if registry == nil {
		registry = NewInMemoryDeviceRegistry(logger)
	}
	if logger == nil {
		logger = logrus.New()
	}
	return &Bridge{
		config:    config,
		publisher: publisher,
		registry:  registry,
		logger:    logger,
		stopCh:    make(chan struct{}),
	}, nil
}

func (b *Bridge) Start(ctx context.Context) error {
	b.mu.Lock()
	if b.started {
		b.mu.Unlock()
		return fmt.Errorf("bridge already started")
	}
	b.started = true
	b.mu.Unlock()
	b.logger.Info("IoT bridge started")
	return nil
}

func (b *Bridge) Stop(ctx context.Context) error {
	b.mu.Lock()
	defer b.mu.Unlock()
	if !b.started {
		return nil
	}
	close(b.stopCh)
	b.started = false
	b.logger.Info("IoT bridge stopped")
	return nil
}

func (b *Bridge) HandleMQTTMessage(ctx context.Context, deviceID uuid.UUID, topic string, payload []byte) error {
	if deviceID == uuid.Nil {
		return fmt.Errorf("invalid device ID")
	}

	eventType := eventTypeFromTopic(topic)
	event := NewDeviceEvent(eventType, deviceID, payload)

	b.registry.Touch(deviceID, "online")

	return b.publishEvent(ctx, event)
}

func (b *Bridge) HandleCOAPRequest(ctx context.Context, deviceID uuid.UUID, eventType string, payload json.RawMessage, authToken string) error {
	if deviceID == uuid.Nil {
		return fmt.Errorf("invalid device ID")
	}
	if eventType == "" {
		eventType = "unknown"
	}

	if b.config != nil && b.config.Auth != nil && b.config.Auth.Enabled {
		if authToken == "" {
			return fmt.Errorf("auth failed: missing token")
		}
		if err := b.verifyAuthToken(ctx, authToken, deviceID); err != nil {
			return fmt.Errorf("auth failed: %w", err)
		}
	}

	event := NewDeviceEvent(eventType, deviceID, payload)
	b.registry.Touch(deviceID, "online")

	return b.publishEvent(ctx, event)
}

func (b *Bridge) SendCommand(ctx context.Context, deviceID uuid.UUID, command json.RawMessage) error {
	if deviceID == uuid.Nil {
		return fmt.Errorf("invalid device ID")
	}

	event := NewDeviceEvent("command", deviceID, command)
	return b.publishEvent(ctx, event)
}

func (b *Bridge) NotifyStatus(ctx context.Context, deviceID uuid.UUID, status string) error {
	if deviceID == uuid.Nil {
		return fmt.Errorf("invalid device ID")
	}

	event := NewDeviceEvent("status", deviceID, json.RawMessage(fmt.Sprintf(`{"status":%q,"ts":%d}`, status, time.Now().Unix())))
	b.registry.Touch(deviceID, status)
	return b.publishEvent(ctx, event)
}

func (b *Bridge) publishEvent(ctx context.Context, event *DeviceEvent) error {
	subject := SubjectForEvent(event.DeviceID, event.EventType)
	data, err := json.Marshal(event)
	if err != nil {
		return fmt.Errorf("failed to marshal event: %w", err)
	}
	return b.publisher.Publish(ctx, subject, data)
}

func (b *Bridge) verifyAuthToken(ctx context.Context, token string, deviceID uuid.UUID) error {
	if b.config == nil || b.config.Auth == nil || !b.config.Auth.Enabled {
		return nil
	}
	if token == "" {
		return fmt.Errorf("missing token")
	}
	if !strings.HasPrefix(token, "Bearer ") && len(token) < 8 {
		return fmt.Errorf("invalid token format")
	}
	return nil
}

func (b *Bridge) IsStarted() bool {
	b.mu.RLock()
	defer b.mu.RUnlock()
	return b.started
}

func eventTypeFromTopic(topic string) string {
	if topic == "" {
		return "unknown"
	}
	parts := splitDots(topic)
	if len(parts) == 0 {
		return "unknown"
	}
	last := parts[len(parts)-1]
	if last == "" {
		return "unknown"
	}
	return last
}

func splitDots(s string) []string {
	out := []string{}
	current := ""
	for _, r := range s {
		if r == '.' {
			out = append(out, current)
			current = ""
			continue
		}
		current += string(r)
	}
	if current != "" || len(out) > 0 {
		out = append(out, current)
	}
	return out
}

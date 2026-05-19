// Package frg provides NATS-based event bus implementation for FRG streaming
package frg

import (
	"encoding/json"
	"fmt"
	"os"
	"sync"
	"time"

	"github.com/google/uuid"
	"github.com/nats-io/nats.go"
	"github.com/sirupsen/logrus"
)

// NATSEventStream implements EventStream using NATS for production deployments
type NATSEventStream struct {
	nc          *nats.Conn
	js          nats.JetStreamContext
	streams     map[uuid.UUID]*nats.StreamInfo
	publishers  map[uuid.UUID]nats.JetStream
	subscribers map[uuid.UUID][]*nats.Subscription
	mu          sync.RWMutex
	url         string

	// Stream configuration
	streamPrefix string
	replicas     int
	retention    time.Duration
}

// NATSConfig configures the NATS connection
type NATSConfig struct {
	URL          string
	StreamPrefix string
	Replicas     int
	Retention    time.Duration
}

// DefaultNATSConfig returns default NATS configuration
func DefaultNATSConfig() *NATSConfig {
	url := os.Getenv("NATS_URL")
	if url == "" {
		url = nats.DefaultURL
	}

	return &NATSConfig{
		URL:          url,
		StreamPrefix: "frg",
		Replicas:     1,
		Retention:    24 * time.Hour,
	}
}

// NewNATSEventStream creates a new NATS-based event stream
func NewNATSEventStream(config *NATSConfig) (*NATSEventStream, error) {
	if config == nil {
		config = DefaultNATSConfig()
	}

	// Connect to NATS
	nc, err := nats.Connect(config.URL,
		nats.Name("FRG Event Bus"),
		nats.MaxReconnects(10),
		nats.ReconnectWait(time.Second),
		nats.DisconnectErrHandler(func(c *nats.Conn, err error) {
			logrus.WithError(err).Warn("NATS disconnected")
		}),
		nats.ReconnectHandler(func(c *nats.Conn) {
			logrus.Info("NATS reconnected")
		}),
	)
	if err != nil {
		return nil, fmt.Errorf("failed to connect to NATS: %w", err)
	}

	// Create JetStream context
	js, err := nc.JetStream()
	if err != nil {
		nc.Close()
		return nil, fmt.Errorf("failed to create JetStream context: %w", err)
	}

	stream := &NATSEventStream{
		nc:           nc,
		js:           js,
		streams:      make(map[uuid.UUID]*nats.StreamInfo),
		publishers:   make(map[uuid.UUID]nats.JetStream),
		subscribers:  make(map[uuid.UUID][]*nats.Subscription),
		url:          config.URL,
		streamPrefix: config.StreamPrefix,
		replicas:     config.Replicas,
		retention:    config.Retention,
	}

	logrus.WithField("url", config.URL).Info("NATS event stream connected")
	return stream, nil
}

// Publish publishes a graph event to NATS
func (s *NATSEventStream) Publish(event *GraphEvent) error {
	s.mu.RLock()
	_, ok := s.streams[event.InstanceID]
	s.mu.RUnlock()

	if !ok {
		// Create stream for this instance
		if err := s.createStream(event.InstanceID); err != nil {
			return err
		}
	}

	// Serialize event
	data, err := json.Marshal(event)
	if err != nil {
		return fmt.Errorf("failed to marshal event: %w", err)
	}

	// Publish to subject
	subject := fmt.Sprintf("%s.%s.%s", s.streamPrefix, event.InstanceID.String(), event.EventType)

	_, err = s.js.Publish(subject, data)
	if err != nil {
		return fmt.Errorf("failed to publish event: %w", err)
	}

	return nil
}

// Subscribe subscribes to events for a specific instance
func (s *NATSEventStream) Subscribe(instanceID uuid.UUID, handler func(*GraphEvent)) error {
	s.mu.RLock()
	_, ok := s.streams[instanceID]
	s.mu.RUnlock()

	if !ok {
		if err := s.createStream(instanceID); err != nil {
			return err
		}
	}

	// Subscribe to all events for this instance
	subject := fmt.Sprintf("%s.%s.*", s.streamPrefix, instanceID.String())

	sub, err := s.js.Subscribe(subject, func(msg *nats.Msg) {
		var event GraphEvent
		if err := json.Unmarshal(msg.Data, &event); err != nil {
			logrus.WithError(err).Error("Failed to unmarshal graph event")
			msg.Nak()
			return
		}

		handler(&event)
		msg.Ack()
	}, nats.Durable(fmt.Sprintf("frg-%s", instanceID.String())), nats.AckExplicit())

	if err != nil {
		return fmt.Errorf("failed to subscribe: %w", err)
	}

	s.mu.Lock()
	s.subscribers[instanceID] = append(s.subscribers[instanceID], sub)
	s.mu.Unlock()

	return nil
}

// Close closes all streams and connections
func (s *NATSEventStream) Close() error {
	s.mu.Lock()
	defer s.mu.Unlock()

	// Unsubscribe all
	for _, subs := range s.subscribers {
		for _, sub := range subs {
			sub.Unsubscribe()
		}
	}

	// Delete streams
	for id := range s.streams {
		streamName := fmt.Sprintf("%s-%s", s.streamPrefix, id.String())
		s.js.DeleteStream(streamName)
	}

	// Close connection
	s.nc.Close()

	logrus.Info("NATS event stream closed")
	return nil
}

// createStream creates a JetStream for a graph instance
func (s *NATSEventStream) createStream(instanceID uuid.UUID) error {
	s.mu.Lock()
	defer s.mu.Unlock()

	if _, ok := s.streams[instanceID]; ok {
		return nil
	}

	streamName := fmt.Sprintf("%s-%s", s.streamPrefix, instanceID.String())
	subject := fmt.Sprintf("%s.%s.*", s.streamPrefix, instanceID.String())

	// Create stream
	_, err := s.js.AddStream(&nats.StreamConfig{
		Name:      streamName,
		Subjects:  []string{subject},
		Retention: nats.WorkQueuePolicy,
		MaxMsgs:   10000,
		MaxAge:    s.retention,
		Storage:   nats.FileStorage,
		Replicas:  s.replicas,
	})

	if err != nil {
		// Stream might already exist
		if err != nats.ErrStreamNameAlreadyInUse {
			return fmt.Errorf("failed to create stream: %w", err)
		}
	}

	// Get stream context
	stream, err := s.js.StreamInfo(streamName)
	if err != nil {
		return fmt.Errorf("failed to get stream: %w", err)
	}

	s.streams[instanceID] = stream

	logrus.WithField("instance_id", instanceID).WithField("stream", streamName).Debug("Created NATS stream")
	return nil
}

// NATSConnection returns the underlying NATS connection (for subscriber use)
func (s *NATSEventStream) NATSConnection() *nats.Conn {
	return s.nc
}

// JetStream returns the JetStream context (for subscriber use)
func (s *NATSEventStream) JetStream() nats.JetStreamContext {
	return s.js
}

package frg

import (
	"os"
	"time"

	"github.com/nats-io/nats.go"
	"github.com/sirupsen/logrus"
)

// TryCreateNATSEventStream attempts to connect to NATS and create a JetStream-based
// event stream.  Returns (stream, nil) on success.  Returns (nil, nil) when
// NATS_URL is not set.  Returns (nil, err) if the connection fails.
//
// Callers should fall back to NewInMemoryEventStream() when nil is returned.
func TryCreateNATSEventStream() (EventStream, error) {
	natsURL := os.Getenv("NATS_URL")
	if natsURL == "" {
		logrus.Debug("NATS_URL not set — using in-memory event stream")
		return nil, nil
	}

	config := &NATSConfig{
		URL:          natsURL,
		StreamPrefix: "frg",
		Replicas:     1,
		Retention:    24 * time.Hour,
	}

	stream, err := NewNATSEventStream(config)
	if err != nil {
		logrus.WithError(err).Warn("NATS connection failed — falling back to in-memory event stream")
		return nil, err
	}

	logrus.WithField("url", natsURL).Info("NATS event stream initialised")
	return stream, nil
}

// TryCreateRuntimeSubscriber connects to NATS and starts a subscriber that
// listens for runtime registration, heartbeat, and execution-result messages.
// Returns nil when NATS_URL is not set or the connection fails.
func TryCreateRuntimeSubscriber(handlers RuntimeEventHandlers) *RuntimeSubscriber {
	natsURL := os.Getenv("NATS_URL")
	if natsURL == "" {
		return nil
	}

	nc, err := nats.Connect(natsURL,
		nats.Name("FunctionFly Orchestrator"),
		nats.MaxReconnects(10),
		nats.ReconnectWait(time.Second),
		nats.DisconnectErrHandler(func(_ *nats.Conn, err error) {
			logrus.WithError(err).Warn("Orchestrator NATS disconnected")
		}),
		nats.ReconnectHandler(func(_ *nats.Conn) {
			logrus.Info("Orchestrator NATS reconnected")
		}),
	)
	if err != nil {
		logrus.WithError(err).Debug("Cannot create runtime subscriber — NATS not available")
		return nil
	}

	subscriber := NewRuntimeSubscriber(nc, DefaultRuntimeSubscriberConfig(), handlers)
	if err := subscriber.Start(); err != nil {
		logrus.WithError(err).Warn("Failed to start NATS runtime subscriber")
		return nil
	}

	return subscriber
}

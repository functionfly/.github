package iot

import (
	"context"
	"fmt"
	"sync"

	mqtt "github.com/mochi-mqtt/server/v2"
	"github.com/mochi-mqtt/server/v2/hooks/auth"
	"github.com/mochi-mqtt/server/v2/listeners"
	"github.com/sirupsen/logrus"
)

type Broker interface {
	Start(ctx context.Context) error
	Stop(ctx context.Context) error
	Port() int
	IsEmbedded() bool
}

type MQTTBroker struct {
	port    int
	useTLS  bool
	broker  *mqtt.Server
	logger  *logrus.Logger
	mu      sync.RWMutex
	started bool
}

func NewMQTTBroker(port int, useTLS bool) *MQTTBroker {
	return &MQTTBroker{
		port:   port,
		useTLS: useTLS,
		logger: logrus.New(),
	}
}

func (b *MQTTBroker) Start(ctx context.Context) error {
	b.mu.Lock()
	defer b.mu.Unlock()

	if b.started {
		return fmt.Errorf("broker already started")
	}

	b.broker = mqtt.New(&mqtt.Options{
		InlineClient: true,
	})

	_ = b.broker.AddHook(newAuthHook(), nil)

	tcp := listeners.NewTCP(listeners.Config{ID: "tcp", Address: fmt.Sprintf(":%d", b.port)})
	if err := b.broker.AddListener(tcp); err != nil {
		return fmt.Errorf("failed to add MQTT listener: %w", err)
	}

	go func() {
		if err := b.broker.Serve(); err != nil {
			b.logger.WithError(err).Error("MQTT broker serve error")
		}
	}()

	b.started = true
	b.logger.WithField("port", b.port).Info("MQTT broker started")
	return nil
}

func (b *MQTTBroker) Stop(ctx context.Context) error {
	b.mu.Lock()
	defer b.mu.Unlock()

	if !b.started {
		return nil
	}

	if err := b.broker.Close(); err != nil {
		return fmt.Errorf("failed to close broker: %w", err)
	}

	b.started = false
	b.logger.Info("MQTT broker stopped")
	return nil
}

func (b *MQTTBroker) Port() int { return b.port }

func (b *MQTTBroker) IsEmbedded() bool { return true }

func (b *MQTTBroker) Server() *mqtt.Server { return b.broker }

func newAuthHook() *auth.AllowHook {
	return &auth.AllowHook{}
}

type ExternalMQTTBroker struct {
	url     string
	port    int
	logger  *logrus.Logger
	mu      sync.RWMutex
	started bool
}

func NewExternalMQTTBroker(url string, port int) *ExternalMQTTBroker {
	return &ExternalMQTTBroker{
		url:    url,
		port:   port,
		logger: logrus.New(),
	}
}

func (b *ExternalMQTTBroker) Start(ctx context.Context) error {
	b.mu.Lock()
	defer b.mu.Unlock()

	if b.started {
		return fmt.Errorf("external broker connection already started")
	}

	b.started = true
	b.logger.WithFields(logrus.Fields{
		"url":  b.url,
		"port": b.port,
	}).Info("External MQTT broker connection initialized")
	return nil
}

func (b *ExternalMQTTBroker) Stop(ctx context.Context) error {
	b.mu.Lock()
	defer b.mu.Unlock()

	if !b.started {
		return nil
	}

	b.started = false
	return nil
}

func (b *ExternalMQTTBroker) Port() int { return b.port }

func (b *ExternalMQTTBroker) IsEmbedded() bool { return false }

func (b *ExternalMQTTBroker) URL() string { return b.url }

package iot_test

import (
	"context"
	"encoding/json"
	"testing"
	"time"

	"github.com/google/uuid"
	"github.com/sirupsen/logrus"

	"github.com/functionfly/functionfly/internal/iot"
)

func newTestBridge(t *testing.T) (*iot.Bridge, *stubPublisher, iot.DeviceRegistry) {
	t.Helper()
	logger := logrus.New()
	logger.SetLevel(logrus.PanicLevel)
	pub := &stubPublisher{}
	reg := iot.NewInMemoryDeviceRegistry(logger)
	bridge, err := iot.NewBridge(iot.DefaultConfig(), pub, reg, logger)
	if err != nil {
		t.Fatalf("NewBridge: %v", err)
	}
	return bridge, pub, reg
}

type stubPublisher struct {
	events []publishedEvent
}

type publishedEvent struct {
	Subject string
	Payload []byte
}

func (s *stubPublisher) Publish(ctx context.Context, subject string, payload []byte) error {
	s.events = append(s.events, publishedEvent{Subject: subject, Payload: payload})
	return nil
}

func (s *stubPublisher) Close() error { return nil }

func (s *stubPublisher) IsConnected() bool { return true }

func TestBridge_StartStop(t *testing.T) {
	bridge, _, _ := newTestBridge(t)
	ctx := context.Background()

	if err := bridge.Start(ctx); err != nil {
		t.Fatalf("Start: %v", err)
	}
	if err := bridge.Start(ctx); err == nil {
		t.Error("expected error starting twice")
	}
	if err := bridge.Stop(ctx); err != nil {
		t.Fatalf("Stop: %v", err)
	}
	if bridge.IsStarted() {
		t.Error("expected bridge to be stopped")
	}
}

func TestBridge_HandleMQTTMessage(t *testing.T) {
	bridge, pub, _ := newTestBridge(t)
	ctx := context.Background()
	_ = bridge.Start(ctx)

	deviceID := uuid.New()
	payload := []byte(`{"temp":72.5}`)
	if err := bridge.HandleMQTTMessage(ctx, deviceID, "telemetry", payload); err != nil {
		t.Fatalf("HandleMQTTMessage: %v", err)
	}

	if len(pub.events) != 1 {
		t.Fatalf("expected 1 event, got %d", len(pub.events))
	}
	if pub.events[0].Subject != "iot.device."+deviceID.String()+".telemetry" {
		t.Errorf("unexpected subject: %s", pub.events[0].Subject)
	}
}

func TestBridge_HandleMQTTMessage_InvalidID(t *testing.T) {
	bridge, _, _ := newTestBridge(t)
	if err := bridge.HandleMQTTMessage(context.Background(), uuid.Nil, "topic", []byte("{}")); err == nil {
		t.Error("expected error for nil device ID")
	}
}

func TestBridge_HandleCOAPRequest(t *testing.T) {
	bridge, pub, _ := newTestBridge(t)
	_ = bridge.Start(context.Background())

	deviceID := uuid.New()
	payload := json.RawMessage(`{"state":"running"}`)
	if err := bridge.HandleCOAPRequest(context.Background(), deviceID, "state", payload, ""); err != nil {
		t.Fatalf("HandleCOAPRequest: %v", err)
	}

	if len(pub.events) != 1 {
		t.Fatalf("expected 1 event, got %d", len(pub.events))
	}
}

func TestBridge_HandleCOAPRequest_WithAuth(t *testing.T) {
	logger := logrus.New()
	logger.SetLevel(logrus.PanicLevel)
	cfg := iot.DefaultConfig()
	cfg.Auth.Enabled = true
	pub := &stubPublisher{}
	reg := iot.NewInMemoryDeviceRegistry(logger)
	authBridge, err := iot.NewBridge(cfg, pub, reg, logger)
	if err != nil {
		t.Fatalf("NewBridge: %v", err)
	}
	_ = authBridge.Start(context.Background())

	deviceID := uuid.New()
	if err := authBridge.HandleCOAPRequest(context.Background(), deviceID, "telemetry", json.RawMessage(`{}`), ""); err == nil {
		t.Error("expected error when auth required but token missing")
	}
	if err := authBridge.HandleCOAPRequest(context.Background(), deviceID, "telemetry", json.RawMessage(`{}`), "short"); err == nil {
		t.Error("expected error for invalid token format")
	}
	if err := authBridge.HandleCOAPRequest(context.Background(), deviceID, "telemetry", json.RawMessage(`{}`), "Bearer valid-token-here"); err != nil {
		t.Errorf("expected no error with valid token, got %v", err)
	}
}

func TestBridge_SendCommand(t *testing.T) {
	bridge, pub, _ := newTestBridge(t)
	_ = bridge.Start(context.Background())

	deviceID := uuid.New()
	cmd := json.RawMessage(`{"action":"turn_on"}`)
	if err := bridge.SendCommand(context.Background(), deviceID, cmd); err != nil {
		t.Fatalf("SendCommand: %v", err)
	}
	if len(pub.events) != 1 {
		t.Fatalf("expected 1 event, got %d", len(pub.events))
	}
}

func TestBridge_NotifyStatus(t *testing.T) {
	bridge, pub, reg := newTestBridge(t)
	_ = bridge.Start(context.Background())

	deviceID := uuid.New()
	_ = reg.Register(context.Background(), &iot.RegisteredDevice{
		ID:         deviceID,
		Topic:      "iot/device/" + deviceID.String(),
		AuthMethod: "psk",
		DeviceType: "sensor",
		Status:     "offline",
	})

	if err := bridge.NotifyStatus(context.Background(), deviceID, "online"); err != nil {
		t.Fatalf("NotifyStatus: %v", err)
	}
	if len(pub.events) != 1 {
		t.Fatalf("expected 1 event, got %d", len(pub.events))
	}

	dev, err := reg.Get(context.Background(), deviceID)
	if err != nil {
		t.Fatalf("Get: %v", err)
	}
	if dev.Status != "online" {
		t.Errorf("expected status online, got %s", dev.Status)
	}
}

func TestDeviceRegistry_CRUD(t *testing.T) {
	reg := iot.NewInMemoryDeviceRegistry(nil)
	ctx := context.Background()
	id := uuid.New()
	now := time.Now()

	dev := &iot.RegisteredDevice{
		ID:         id,
		Topic:      "test",
		AuthMethod: "psk",
		DeviceType: "sensor",
		LastSeen:   now,
		Status:     "offline",
	}
	if err := reg.Register(ctx, dev); err != nil {
		t.Fatalf("Register: %v", err)
	}

	got, err := reg.Get(ctx, id)
	if err != nil {
		t.Fatalf("Get: %v", err)
	}
	if got.DeviceType != "sensor" {
		t.Errorf("unexpected device type: %s", got.DeviceType)
	}

	if err := reg.UpdateStatus(ctx, id, "online", time.Now()); err != nil {
		t.Fatalf("UpdateStatus: %v", err)
	}
	if err := reg.Delete(ctx, id); err != nil {
		t.Fatalf("Delete: %v", err)
	}
	if _, err := reg.Get(ctx, id); err == nil {
		t.Error("expected error after delete")
	}
}

func TestNATSClient_PublishInvalidSubject(t *testing.T) {
	client := iot.NewNATSClient("nats://localhost:4222", nil)
	if err := client.Publish(context.Background(), "", []byte("{}")); err == nil {
		t.Error("expected error for empty subject")
	}
	if err := client.Publish(context.Background(), "with space", []byte("{}")); err == nil {
		t.Error("expected error for invalid subject")
	}
}

func TestSubjectForEvent(t *testing.T) {
	id := uuid.New()
	cases := []struct {
		eventType string
		want      string
	}{
		{"telemetry", "iot.device." + id.String() + ".telemetry"},
		{"command", "iot.device." + id.String() + ".command"},
		{"status", "iot.device." + id.String() + ".status"},
		{"error", "iot.device." + id.String() + ".error"},
		{"state", "iot.device." + id.String() + ".state"},
		{"unknown", "iot.device." + id.String() + ".unknown"},
	}
	for _, c := range cases {
		if got := iot.SubjectForEvent(id, c.eventType); got != c.want {
			t.Errorf("event %s: want %s, got %s", c.eventType, c.want, got)
		}
	}
}

func TestConfig_Default(t *testing.T) {
	cfg := iot.DefaultConfig()
	if cfg.MQTTPort != 1883 {
		t.Errorf("expected MQTT port 1883, got %d", cfg.MQTTPort)
	}
	if cfg.COAPPort != 5683 {
		t.Errorf("expected COAP port 5683, got %d", cfg.COAPPort)
	}
	if !cfg.UseEmbedded {
		t.Error("expected UseEmbedded to be true by default")
	}
}

package iot

import (
	"context"
	"encoding/json"
	"sync"
	"time"

	"github.com/google/uuid"
	"github.com/sirupsen/logrus"
)

type DeviceRegistry interface {
	Register(ctx context.Context, device *RegisteredDevice) error
	Get(ctx context.Context, deviceID uuid.UUID) (*RegisteredDevice, error)
	UpdateStatus(ctx context.Context, deviceID uuid.UUID, status string, lastSeen time.Time) error
	List(ctx context.Context) ([]*RegisteredDevice, error)
	Delete(ctx context.Context, deviceID uuid.UUID) error
	Touch(deviceID uuid.UUID, status string)
}

type RegisteredDevice struct {
	ID         uuid.UUID       `json:"id"`
	Topic      string          `json:"topic"`
	AuthMethod string          `json:"auth_method"`
	DeviceType string          `json:"device_type"`
	Metadata   map[string]any  `json:"metadata,omitempty"`
	LastSeen   time.Time       `json:"last_seen"`
	Status     string          `json:"status"`
}

type InMemoryDeviceRegistry struct {
	devices map[uuid.UUID]*RegisteredDevice
	mu      sync.RWMutex
	logger  *logrus.Logger
}

func NewInMemoryDeviceRegistry(logger *logrus.Logger) *InMemoryDeviceRegistry {
	if logger == nil {
		logger = logrus.New()
	}
	return &InMemoryDeviceRegistry{
		devices: make(map[uuid.UUID]*RegisteredDevice),
		logger:  logger,
	}
}

func (r *InMemoryDeviceRegistry) Register(ctx context.Context, device *RegisteredDevice) error {
	if device == nil {
		return ErrInvalidDevice
	}
	if device.ID == uuid.Nil {
		device.ID = uuid.New()
	}

	r.mu.Lock()
	defer r.mu.Unlock()
	r.devices[device.ID] = device
	r.logger.WithField("device_id", device.ID).Debug("device registered")
	return nil
}

func (r *InMemoryDeviceRegistry) Get(ctx context.Context, deviceID uuid.UUID) (*RegisteredDevice, error) {
	r.mu.RLock()
	defer r.mu.RUnlock()

	d, ok := r.devices[deviceID]
	if !ok {
		return nil, ErrDeviceNotFound
	}
	copy := *d
	return &copy, nil
}

func (r *InMemoryDeviceRegistry) UpdateStatus(ctx context.Context, deviceID uuid.UUID, status string, lastSeen time.Time) error {
	r.mu.Lock()
	defer r.mu.Unlock()

	d, ok := r.devices[deviceID]
	if !ok {
		return ErrDeviceNotFound
	}
	d.Status = status
	d.LastSeen = lastSeen
	return nil
}

func (r *InMemoryDeviceRegistry) List(ctx context.Context) ([]*RegisteredDevice, error) {
	r.mu.RLock()
	defer r.mu.RUnlock()

	out := make([]*RegisteredDevice, 0, len(r.devices))
	for _, d := range r.devices {
		copy := *d
		out = append(out, &copy)
	}
	return out, nil
}

func (r *InMemoryDeviceRegistry) Delete(ctx context.Context, deviceID uuid.UUID) error {
	r.mu.Lock()
	defer r.mu.Unlock()

	if _, ok := r.devices[deviceID]; !ok {
		return ErrDeviceNotFound
	}
	delete(r.devices, deviceID)
	return nil
}

func (r *InMemoryDeviceRegistry) Touch(deviceID uuid.UUID, status string) {
	r.mu.Lock()
	defer r.mu.Unlock()
	if d, ok := r.devices[deviceID]; ok {
		d.Status = status
		d.LastSeen = time.Now()
	}
}

var (
	ErrDeviceNotFound  = registryError("device not found")
	ErrInvalidDevice   = registryError("invalid device")
)

type registryError string

func (e registryError) Error() string { return string(e) }

type DeviceEvent struct {
	EventType string          `json:"event_type"`
	DeviceID  uuid.UUID       `json:"device_id"`
	Timestamp time.Time       `json:"timestamp"`
	Payload   json.RawMessage `json:"payload,omitempty"`
	Metadata  map[string]any  `json:"metadata,omitempty"`
}

func NewDeviceEvent(eventType string, deviceID uuid.UUID, payload json.RawMessage) *DeviceEvent {
	return &DeviceEvent{
		EventType: eventType,
		DeviceID:  deviceID,
		Timestamp: time.Now(),
		Payload:   payload,
	}
}

package iot

import (
	"crypto/sha256"
	"encoding/hex"
	"fmt"
	"sync"
	"time"

	"github.com/google/uuid"
	"gorm.io/gorm"
)

type PSKStore struct {
	db  *gorm.DB
	mu  sync.RWMutex
}

type PSKStoreConfig struct {
	HashPSK bool
}

func NewPSKStore(db *gorm.DB, config *PSKStoreConfig) *PSKStore {
	return &PSKStore{db: db}
}

func (s *PSKStore) ValidatePSK(deviceID uuid.UUID, psk string) (*DeviceClaims, error) {
	s.mu.RLock()
	defer s.mu.RUnlock()

	var device Device
	if err := s.db.Where("id = ? AND auth_method = ?", deviceID, AuthMethodPSK).First(&device).Error; err != nil {
		if err == gorm.ErrRecordNotFound {
			return nil, fmt.Errorf("device not found or not using PSK")
		}
		return nil, fmt.Errorf("database error: %w", err)
	}

	providedHash := s.hashPSK(psk)
	if providedHash != device.PSKHash {
		return nil, fmt.Errorf("invalid PSK")
	}

	now := time.Now()
	device.LastSeen = &now
	s.db.Save(&device)

	return &DeviceClaims{
		DeviceID:    device.ID,
		TenantID:    device.TenantID,
		DeviceType:  device.DeviceType,
		Name:        device.Name,
		AuthMethod:  AuthMethodPSK,
		Permissions: getPermissionsForDeviceType(device.DeviceType),
	}, nil
}

func (s *PSKStore) RegisterDevicePSK(db *gorm.DB, tenantID uuid.UUID, name string, deviceType DeviceType, psk string) (*Device, error) {
	pskHash := s.hashPSK(psk)

	device := &Device{
		TenantID:   tenantID,
		Name:       name,
		DeviceType: deviceType,
		AuthMethod: AuthMethodPSK,
		Status:     DeviceStatusOffline,
		PSKHash:    pskHash,
		Metadata: map[string]any{
			"registered_at": time.Now(),
		},
	}

	if err := db.Create(device).Error; err != nil {
		return nil, fmt.Errorf("failed to register device: %w", err)
	}

	state := &DeviceState{
		DeviceID: device.ID,
		State:    map[string]any{},
	}
	if err := db.Create(state).Error; err != nil {
		return nil, fmt.Errorf("failed to create device state: %w", err)
	}

	return device, nil
}

func (s *PSKStore) hashPSK(psk string) string {
	sum := sha256.Sum256([]byte(psk))
	return hex.EncodeToString(sum[:])
}

func (s *PSKStore) RotatePSK(deviceID uuid.UUID, oldPSK, newPSK string) error {
	var device Device
	if err := s.db.Where("id = ? AND auth_method = ?", deviceID, AuthMethodPSK).First(&device).Error; err != nil {
		return fmt.Errorf("device not found: %w", err)
	}

	oldHash := s.hashPSK(oldPSK)
	if oldHash != device.PSKHash {
		return fmt.Errorf("invalid old PSK")
	}

	device.PSKHash = s.hashPSK(newPSK)
	return s.db.Save(&device).Error
}

func getPermissionsForDeviceType(deviceType DeviceType) []string {
	switch deviceType {
	case DeviceTypeGateway:
		return []string{
			"iot.devices.read",
			"iot.devices.write",
			"iot.telemetry.publish",
			"iot.commands.receive",
			"iot.state.read",
			"iot.state.write",
		}
	case DeviceTypeSensor:
		return []string{
			"iot.telemetry.publish",
			"iot.state.read",
		}
	case DeviceTypeActuator:
		return []string{
			"iot.commands.receive",
			"iot.state.read",
			"iot.state.write",
		}
	case DeviceTypeRobot:
		return []string{
			"iot.telemetry.publish",
			"iot.commands.receive",
			"iot.state.read",
			"iot.state.write",
		}
	case DeviceTypeCamera:
		return []string{
			"iot.telemetry.publish",
			"iot.state.read",
		}
	default:
		return []string{
			"iot.telemetry.publish",
		}
	}
}

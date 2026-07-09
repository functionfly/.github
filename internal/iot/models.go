package iot

import (
	"time"

	"github.com/google/uuid"
	"gorm.io/gorm"
)

type DeviceType string

const (
	DeviceTypeGateway   DeviceType = "gateway"
	DeviceTypeSensor   DeviceType = "sensor"
	DeviceTypeActuator DeviceType = "actuator"
	DeviceTypeRobot    DeviceType = "robot"
	DeviceTypeCamera   DeviceType = "camera"
	DeviceTypeOther    DeviceType = "other"
)

type AuthMethod string

const (
	AuthMethodX509 AuthMethod = "x509"
	AuthMethodJWT  AuthMethod = "jwt"
	AuthMethodPSK  AuthMethod = "psk"
)

type DeviceStatus string

const (
	DeviceStatusOnline  DeviceStatus = "online"
	DeviceStatusOffline DeviceStatus = "offline"
	DeviceStatusError   DeviceStatus = "error"
)

type Device struct {
	ID              uuid.UUID      `gorm:"type:uuid;primary_key"`
	TenantID        uuid.UUID      `gorm:"type:uuid;not null;index"`
	Name            string         `gorm:"size:256;not null"`
	DeviceType      DeviceType     `gorm:"size:64;not null;index"`
	AuthMethod      AuthMethod     `gorm:"size:32;not null"`
	Status          DeviceStatus   `gorm:"size:32;not null;default:'offline'"`
	CertFingerprint string         `gorm:"size:255;index"`
	PSKHash         string         `gorm:"size:255"`
	LastSeen        *time.Time     `gorm:"index"`
	Metadata        map[string]any `gorm:"serializer:json"`
	CreatedAt       time.Time
	UpdatedAt       time.Time
	DeletedAt       gorm.DeletedAt `gorm:"index"`
}

func (Device) TableName() string {
	return "iot_devices"
}

func (d *Device) BeforeCreate(tx *gorm.DB) error {
	if d.ID == uuid.Nil {
		d.ID = uuid.New()
	}
	return nil
}

type DeviceState struct {
	DeviceID      uuid.UUID      `gorm:"type:uuid;primary_key"`
	State         map[string]any `gorm:"serializer:json;not null;default:'{}'"`
	LastTelemetry map[string]any `gorm:"serializer:json"`
	UpdatedAt     time.Time      `gorm:"not null"`
}

func (DeviceState) TableName() string {
	return "iot_device_states"
}

func (ds *DeviceState) BeforeCreate(tx *gorm.DB) error {
	if ds.DeviceID == uuid.Nil {
		ds.DeviceID = uuid.New()
	}
	return nil
}

type DeviceCommand struct {
	ID             uuid.UUID      `gorm:"type:uuid;primary_key"`
	DeviceID       uuid.UUID      `gorm:"type:uuid;not null;index"`
	Command        map[string]any `gorm:"serializer:json;not null"`
	Status         string         `gorm:"size:32;not null;default:'pending'"`
	CreatedAt      time.Time      `gorm:"not null;index"`
	AcknowledgedAt *time.Time
}

func (DeviceCommand) TableName() string {
	return "iot_commands"
}

func (dc *DeviceCommand) BeforeCreate(tx *gorm.DB) error {
	if dc.ID == uuid.Nil {
		dc.ID = uuid.New()
	}
	return nil
}

type DeviceClaims struct {
	DeviceID    uuid.UUID  `json:"device_id"`
	TenantID   uuid.UUID  `json:"tenant_id"`
	DeviceType DeviceType `json:"device_type"`
	Name       string     `json:"name"`
	AuthMethod AuthMethod `json:"auth_method"`
	Permissions []string   `json:"permissions"`
}

func Migrate(db *gorm.DB) error {
	return db.AutoMigrate(
		&Device{},
		&DeviceState{},
		&DeviceCommand{},
	)
}

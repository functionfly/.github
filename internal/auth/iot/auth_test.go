package iot

import (
	"context"
	"crypto/ecdsa"
	"crypto/elliptic"
	"crypto/rand"
	"crypto/tls"
	"crypto/x509"
	"crypto/x509/pkix"
	"encoding/pem"
	"fmt"
	"math/big"
	"os"
	"testing"
	"time"

	"github.com/google/uuid"
	"gorm.io/driver/postgres"
	"gorm.io/gorm"
	"gorm.io/gorm/logger"
)

func setupTestDB(t *testing.T) *gorm.DB {
	dsn := os.Getenv("DATABASE_URL")
	if dsn == "" {
		dsn = "host=localhost user=postgres password=postgres dbname=functionfly port=5432 sslmode=disable"
	}

	db, err := gorm.Open(postgres.Open(dsn), &gorm.Config{
		Logger: logger.Default.LogMode(logger.Silent),
	})
	if err != nil {
		t.Fatalf("failed to open test db: %v", err)
	}

	if err := Migrate(db); err != nil {
		t.Fatalf("failed to migrate: %v", err)
	}

	db.Exec("DROP TABLE IF EXISTS iot_devices")
	db.Exec("DROP TABLE IF EXISTS iot_device_states")
	db.Exec("DROP TABLE IF EXISTS iot_commands")
	if err := Migrate(db); err != nil {
		t.Fatalf("failed to migrate after drop: %v", err)
	}

	return db
}

func generateTestCert(t *testing.T, commonName string) (*x509.Certificate, *ecdsa.PrivateKey) {
	privateKey, err := ecdsa.GenerateKey(elliptic.P256(), rand.Reader)
	if err != nil {
		t.Fatalf("failed to generate key: %v", err)
	}

	template := &x509.Certificate{
		SerialNumber: big.NewInt(1),
		Subject: pkix.Name{
			CommonName: commonName,
		},
		NotBefore:             time.Now(),
		NotAfter:              time.Now().Add(24 * time.Hour),
		KeyUsage:              x509.KeyUsageKeyEncipherment | x509.KeyUsageDigitalSignature,
		ExtKeyUsage:           []x509.ExtKeyUsage{x509.ExtKeyUsageClientAuth},
		BasicConstraintsValid: true,
	}

	certDER, err := x509.CreateCertificate(rand.Reader, template, template, &privateKey.PublicKey, privateKey)
	if err != nil {
		t.Fatalf("failed to create certificate: %v", err)
	}

	cert, err := x509.ParseCertificate(certDER)
	if err != nil {
		t.Fatalf("failed to parse certificate: %v", err)
	}

	return cert, privateKey
}

func encodeCertPEM(t *testing.T, cert *x509.Certificate) []byte {
	certPEM := pem.EncodeToMemory(&pem.Block{
		Type:  "CERTIFICATE",
		Bytes: cert.Raw,
	})
	return certPEM
}

func TestDeviceModel_Creation(t *testing.T) {
	db := setupTestDB(t)

	device := &Device{
		TenantID:   uuid.New(),
		Name:       "test-sensor",
		DeviceType: DeviceTypeSensor,
		AuthMethod: AuthMethodPSK,
		Status:     DeviceStatusOffline,
		PSKHash:    "test-hash",
	}

	if err := db.Create(device).Error; err != nil {
		t.Fatalf("failed to create device: %v", err)
	}

	var found Device
	if err := db.First(&found, "id = ?", device.ID).Error; err != nil {
		t.Fatalf("failed to find device: %v", err)
	}

	if found.Name != device.Name {
		t.Errorf("expected name %s, got %s", device.Name, found.Name)
	}
	if found.DeviceType != DeviceTypeSensor {
		t.Errorf("expected type %s, got %s", DeviceTypeSensor, found.DeviceType)
	}
}

func TestDeviceState_Creation(t *testing.T) {
	db := setupTestDB(t)

	device := &Device{
		TenantID:   uuid.New(),
		Name:       "test-gateway",
		DeviceType: DeviceTypeGateway,
		AuthMethod: AuthMethodX509,
		Status:     DeviceStatusOffline,
	}
	if err := db.Create(device).Error; err != nil {
		t.Fatalf("failed to create device: %v", err)
	}

	state := &DeviceState{
		DeviceID: device.ID,
		State:   map[string]any{"online": true},
	}
	if err := db.Create(state).Error; err != nil {
		t.Fatalf("failed to create state: %v", err)
	}

	var found DeviceState
	if err := db.First(&found, "device_id = ?", device.ID).Error; err != nil {
		t.Fatalf("failed to find state: %v", err)
	}

	if found.State["online"] != true {
		t.Errorf("expected state online=true")
	}
}

func TestDeviceCommand_Creation(t *testing.T) {
	db := setupTestDB(t)

	device := &Device{
		TenantID:   uuid.New(),
		Name:       "test-actuator",
		DeviceType: DeviceTypeActuator,
		AuthMethod: AuthMethodPSK,
		Status:     DeviceStatusOnline,
	}
	if err := db.Create(device).Error; err != nil {
		t.Fatalf("failed to create device: %v", err)
	}

	cmd := &DeviceCommand{
		DeviceID: device.ID,
		Command:  map[string]any{"action": "activate", "value": 100},
		Status:   "pending",
	}
	if err := db.Create(cmd).Error; err != nil {
		t.Fatalf("failed to create command: %v", err)
	}

	var found DeviceCommand
	if err := db.First(&found, "device_id = ?", device.ID).Error; err != nil {
		t.Fatalf("failed to find command: %v", err)
	}

	if found.Command["action"] != "activate" {
		t.Errorf("expected action=activate")
	}
}

func TestRateLimiter_Allow(t *testing.T) {
	rl := NewRateLimiter(3, time.Second)

	for i := 0; i < 3; i++ {
		if !rl.Allow("device-1") {
			t.Errorf("request %d should be allowed", i+1)
		}
	}

	if rl.Allow("device-1") {
		t.Error("4th request should be denied")
	}

	if !rl.Allow("device-2") {
		t.Error("different device should be allowed")
	}
}

func TestRateLimiter_Reset(t *testing.T) {
	rl := NewRateLimiter(1, time.Second)

	if !rl.Allow("device-1") {
		t.Error("first request should be allowed")
	}

	if rl.Allow("device-1") {
		t.Error("second request should be denied")
	}

	rl.Reset("device-1")

	if !rl.Allow("device-1") {
		t.Error("after reset, request should be allowed")
	}
}

func TestDeviceAuth_Disabled(t *testing.T) {
	db := setupTestDB(t)

	config := &DeviceAuthConfig{Enabled: false}
	auth, err := New(db, config, nil)
	if err != nil {
		t.Fatalf("failed to create auth: %v", err)
	}

	claims, err := auth.AuthenticateDevice(context.Background(), AuthMethodPSK, &PSKCredential{
		DeviceID: uuid.New(),
		PSK:      "test",
	})

	if claims != nil {
		t.Error("should not return claims when disabled")
	}
	if err == nil {
		t.Error("should return error when disabled")
	}
}

func TestDeviceAuth_PSKRegistration(t *testing.T) {
	db := setupTestDB(t)

	config := &DeviceAuthConfig{Enabled: true}
	auth, err := New(db, config, nil)
	if err != nil {
		t.Fatalf("failed to create auth: %v", err)
	}

	tenantID := uuid.New()
	device, err := auth.RegisterDevice(context.Background(), tenantID, "test-sensor", DeviceTypeSensor, AuthMethodPSK, "secret-psk")
	if err != nil {
		t.Fatalf("failed to register device: %v", err)
	}

	if device.Name != "test-sensor" {
		t.Errorf("expected name test-sensor, got %s", device.Name)
	}
	if device.AuthMethod != AuthMethodPSK {
		t.Errorf("expected auth method PSK, got %s", device.AuthMethod)
	}

	claims, err := auth.AuthenticateDevice(context.Background(), AuthMethodPSK, &PSKCredential{
		DeviceID: device.ID,
		PSK:      "secret-psk",
	})
	if err != nil {
		t.Fatalf("failed to authenticate: %v", err)
	}

	if claims.DeviceID != device.ID {
		t.Errorf("expected device ID %s, got %s", device.ID, claims.DeviceID)
	}
}

func TestDeviceAuth_JWT(t *testing.T) {
	db := setupTestDB(t)

	config := &DeviceAuthConfig{
		Enabled:    true,
		JWTSecret:  "test-secret-key-for-jwt-signing-32bytes",
		JWTIssuer:  "test-issuer",
		JWTExpiry:  24 * time.Hour,
	}
	auth, err := New(db, config, nil)
	if err != nil {
		t.Fatalf("failed to create auth: %v", err)
	}

	tenantID := uuid.New()
	device, err := auth.RegisterDevice(context.Background(), tenantID, "test-gateway", DeviceTypeGateway, AuthMethodJWT, nil)
	if err != nil {
		t.Fatalf("failed to register device: %v", err)
	}

	token, err := auth.IssueDeviceToken(context.Background(), device.ID)
	if err != nil {
		t.Fatalf("failed to issue token: %v", err)
	}

	claims, err := auth.AuthenticateDevice(context.Background(), AuthMethodJWT, token)
	if err != nil {
		t.Fatalf("failed to authenticate: %v", err)
	}

	if claims.DeviceID != device.ID {
		t.Errorf("expected device ID %s, got %s", device.ID, claims.DeviceID)
	}
	if claims.TenantID != tenantID {
		t.Errorf("expected tenant ID %s, got %s", tenantID, claims.TenantID)
	}
}

func TestDeviceAuth_DeviceState(t *testing.T) {
	db := setupTestDB(t)

	config := &DeviceAuthConfig{Enabled: true}
	auth, err := New(db, config, nil)
	if err != nil {
		t.Fatalf("failed to create auth: %v", err)
	}

	tenantID := uuid.New()
	device, err := auth.RegisterDevice(context.Background(), tenantID, "test-sensor", DeviceTypeSensor, AuthMethodPSK, "psk")
	if err != nil {
		t.Fatalf("failed to register device: %v", err)
	}

	_, err = auth.GetDeviceState(context.Background(), device.ID)
	if err != nil {
		t.Fatalf("failed to get state: %v", err)
	}

	newState := map[string]any{"temperature": 25.5, "humidity": 60}
	if err := auth.UpdateDeviceState(context.Background(), device.ID, newState); err != nil {
		t.Fatalf("failed to update state: %v", err)
	}

	updatedState, err := auth.GetDeviceState(context.Background(), device.ID)
	if err != nil {
		t.Fatalf("failed to get updated state: %v", err)
	}

	if updatedState.State["temperature"] != 25.5 {
		t.Errorf("expected temperature 25.5, got %v", updatedState.State["temperature"])
	}
}

func TestDeviceAuth_Commands(t *testing.T) {
	db := setupTestDB(t)

	config := &DeviceAuthConfig{Enabled: true}
	auth, err := New(db, config, nil)
	if err != nil {
		t.Fatalf("failed to create auth: %v", err)
	}

	tenantID := uuid.New()
	device, err := auth.RegisterDevice(context.Background(), tenantID, "test-actuator", DeviceTypeActuator, AuthMethodPSK, "psk")
	if err != nil {
		t.Fatalf("failed to register device: %v", err)
	}

	command, err := auth.CreateCommand(context.Background(), device.ID, map[string]any{"action": "turn_on"})
	if err != nil {
		t.Fatalf("failed to create command: %v", err)
	}

	commands, err := auth.ListPendingCommands(context.Background(), device.ID)
	if err != nil {
		t.Fatalf("failed to list commands: %v", err)
	}

	if len(commands) != 1 {
		t.Errorf("expected 1 command, got %d", len(commands))
	}

	if err := auth.AcknowledgeCommand(context.Background(), command.ID); err != nil {
		t.Fatalf("failed to acknowledge command: %v", err)
	}

	commands, err = auth.ListPendingCommands(context.Background(), device.ID)
	if err != nil {
		t.Fatalf("failed to list commands: %v", err)
	}

	if len(commands) != 0 {
		t.Errorf("expected 0 pending commands, got %d", len(commands))
	}
}

func TestDeviceAuth_DeviceStatus(t *testing.T) {
	db := setupTestDB(t)

	config := &DeviceAuthConfig{Enabled: true}
	auth, err := New(db, config, nil)
	if err != nil {
		t.Fatalf("failed to create auth: %v", err)
	}

	tenantID := uuid.New()
	device, err := auth.RegisterDevice(context.Background(), tenantID, "test-sensor", DeviceTypeSensor, AuthMethodPSK, "psk")
	if err != nil {
		t.Fatalf("failed to register device: %v", err)
	}

	if device.Status != DeviceStatusOffline {
		t.Errorf("expected initial status offline, got %s", device.Status)
	}

	if err := auth.UpdateDeviceStatus(context.Background(), device.ID, DeviceStatusOnline); err != nil {
		t.Fatalf("failed to update status: %v", err)
	}

	updatedDevice, err := auth.GetDevice(context.Background(), device.ID)
	if err != nil {
		t.Fatalf("failed to get device: %v", err)
	}

	if updatedDevice.Status != DeviceStatusOnline {
		t.Errorf("expected status online, got %s", updatedDevice.Status)
	}

	if updatedDevice.LastSeen == nil {
		t.Error("expected LastSeen to be set")
	}
}

func TestDeviceAuth_ListDevices(t *testing.T) {
	db := setupTestDB(t)

	config := &DeviceAuthConfig{Enabled: true}
	auth, err := New(db, config, nil)
	if err != nil {
		t.Fatalf("failed to create auth: %v", err)
	}

	tenantID := uuid.New()

	for i := 0; i < 3; i++ {
		deviceType := DeviceTypeSensor
		if i == 1 {
			deviceType = DeviceTypeGateway
		}
		_, err := auth.RegisterDevice(context.Background(), tenantID, fmt.Sprintf("device-%d", i), deviceType, AuthMethodPSK, "psk")
		if err != nil {
			t.Fatalf("failed to register device: %v", err)
		}
	}

	devices, err := auth.ListDevices(context.Background(), tenantID, nil)
	if err != nil {
		t.Fatalf("failed to list devices: %v", err)
	}

	if len(devices) != 3 {
		t.Errorf("expected 3 devices, got %d", len(devices))
	}

	sensorType := DeviceTypeSensor
	sensors, err := auth.ListDevices(context.Background(), tenantID, &sensorType)
	if err != nil {
		t.Fatalf("failed to list sensors: %v", err)
	}

	if len(sensors) != 2 {
		t.Errorf("expected 2 sensors, got %d", len(sensors))
	}
}

func TestCertValidator_ComputeFingerprint(t *testing.T) {
	db := setupTestDB(t)

	cert, _ := generateTestCert(t, "test-device")
	certPEM := encodeCertPEM(t, cert)

	block, _ := pem.Decode(certPEM)
	parsedCert, err := x509.ParseCertificate(block.Bytes)
	if err != nil {
		t.Fatalf("failed to parse cert: %v", err)
	}

	cv := &CertValidator{db: db}
	fingerprint := cv.ComputeFingerprint(parsedCert)

	if len(fingerprint) != 64 {
		t.Errorf("expected 64 char fingerprint, got %d", len(fingerprint))
	}

	fingerprint2 := cv.ComputeFingerprint(parsedCert)
	if fingerprint != fingerprint2 {
		t.Error("fingerprint should be deterministic")
	}
}

func TestDeviceClaims_Permissions(t *testing.T) {
	db := setupTestDB(t)

	config := &DeviceAuthConfig{Enabled: true}
	auth, err := New(db, config, nil)
	if err != nil {
		t.Fatalf("failed to create auth: %v", err)
	}

	tenantID := uuid.New()

	testCases := []struct {
		deviceType  DeviceType
		expectedLen int
	}{
		{DeviceTypeGateway, 6},
		{DeviceTypeSensor, 2},
		{DeviceTypeActuator, 3},
		{DeviceTypeRobot, 4},
	}

	for _, tc := range testCases {
		device, err := auth.RegisterDevice(context.Background(), tenantID, "test", tc.deviceType, AuthMethodPSK, "psk")
		if err != nil {
			t.Fatalf("failed to register %s device: %v", tc.deviceType, err)
		}

		claims, err := auth.AuthenticateDevice(context.Background(), AuthMethodPSK, &PSKCredential{
			DeviceID: device.ID,
			PSK:      "psk",
		})
		if err != nil {
			t.Fatalf("failed to authenticate: %v", err)
		}

		if len(claims.Permissions) != tc.expectedLen {
			t.Errorf("%s: expected %d permissions, got %d", tc.deviceType, tc.expectedLen, len(claims.Permissions))
		}
	}
}

func TestMigrate(t *testing.T) {
	dsn := os.Getenv("DATABASE_URL")
	if dsn == "" {
		dsn = "host=localhost user=postgres password=postgres dbname=functionfly port=5432 sslmode=disable"
	}

	db, err := gorm.Open(postgres.Open(dsn), &gorm.Config{
		Logger: logger.Default.LogMode(logger.Silent),
	})
	if err != nil {
		t.Fatalf("failed to open db: %v", err)
	}

	db.Exec("DROP TABLE IF EXISTS iot_devices")
	db.Exec("DROP TABLE IF EXISTS iot_device_states")
	db.Exec("DROP TABLE IF EXISTS iot_commands")
	if err := Migrate(db); err != nil {
		t.Fatalf("failed to migrate: %v", err)
	}

	device := &Device{
		TenantID:   uuid.New(),
		Name:       "migrated-device",
		DeviceType: DeviceTypeSensor,
		AuthMethod: AuthMethodPSK,
		Status:     DeviceStatusOffline,
		PSKHash:    "hash",
	}

	if err := db.Create(device).Error; err != nil {
		t.Fatalf("failed to create device after migrate: %v", err)
	}
}

func TestDeviceAuth_TLSCredential(t *testing.T) {
	db := setupTestDB(t)

	cert, _ := generateTestCert(t, "test-device")
	certPEM := encodeCertPEM(t, cert)

	block, _ := pem.Decode(certPEM)
	parsedCert, _ := x509.ParseCertificate(block.Bytes)

	cv := &CertValidator{db: db}
	fingerprint := cv.ComputeFingerprint(parsedCert)

	device := &Device{
		TenantID:        uuid.New(),
		Name:            "tls-device",
		DeviceType:      DeviceTypeGateway,
		AuthMethod:      AuthMethodX509,
		Status:          DeviceStatusOffline,
		CertFingerprint: fingerprint,
	}
	if err := db.Create(device).Error; err != nil {
		t.Fatalf("failed to create device: %v", err)
	}

	state := &DeviceState{
		DeviceID: device.ID,
		State:    map[string]any{},
	}
	if err := db.Create(state).Error; err != nil {
		t.Fatalf("failed to create state: %v", err)
	}

	config := &DeviceAuthConfig{
		Enabled:      true,
		CACertPaths: []string{},
		SkipCertVerify: true,
	}
	auth, err := New(db, config, nil)
	if err != nil {
		t.Fatalf("failed to create auth: %v", err)
	}

	tlsConnState := &tls.ConnectionState{
		VerifiedChains: [][]*x509.Certificate{{parsedCert}},
	}

	claims, err := auth.AuthenticateDevice(context.Background(), AuthMethodX509, tlsConnState)
	if err != nil {
		t.Fatalf("failed to authenticate with TLS: %v", err)
	}

	if claims.DeviceID != device.ID {
		t.Errorf("expected device ID %s, got %s", device.ID, claims.DeviceID)
	}
}

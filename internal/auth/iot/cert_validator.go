package iot

import (
	"crypto/sha256"
	"crypto/tls"
	"crypto/x509"
	"encoding/hex"
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"sync"
	"time"

	"github.com/google/uuid"
	"gorm.io/gorm"
)

type CertValidator struct {
	db          *gorm.DB
	caCertPool  *x509.CertPool
	caCertPaths []string
	skipVerify  bool
	mu          sync.RWMutex
}

type CertValidatorConfig struct {
	CACertDir  string
	CACertPaths []string
	SkipVerify bool
}

func NewCertValidator(db *gorm.DB, config *CertValidatorConfig) (*CertValidator, error) {
	cv := &CertValidator{
		db:          db,
		caCertPaths: config.CACertPaths,
		skipVerify:  config.SkipVerify,
	}

	if config.CACertDir != "" {
		if err := cv.loadCACertsFromDir(config.CACertDir); err != nil {
			return nil, fmt.Errorf("failed to load CA certs: %w", err)
		}
	} else if len(config.CACertPaths) > 0 {
		if err := cv.loadCACerts(config.CACertPaths); err != nil {
			return nil, fmt.Errorf("failed to load CA certs: %w", err)
		}
	}

	return cv, nil
}

func (cv *CertValidator) loadCACertsFromDir(dir string) error {
	entries, err := os.ReadDir(dir)
	if err != nil {
		return fmt.Errorf("failed to read CA cert directory: %w", err)
	}

	var paths []string
	for _, entry := range entries {
		if !entry.IsDir() && (strings.HasSuffix(entry.Name(), ".pem") || strings.HasSuffix(entry.Name(), ".crt") || strings.HasSuffix(entry.Name(), ".cer")) {
			paths = append(paths, filepath.Join(dir, entry.Name()))
		}
	}

	return cv.loadCACerts(paths)
}

func (cv *CertValidator) loadCACerts(paths []string) error {
	cv.mu.Lock()
	defer cv.mu.Unlock()

	cv.caCertPool = x509.NewCertPool()
	for _, path := range paths {
		certPEM, err := os.ReadFile(path)
		if err != nil {
			return fmt.Errorf("failed to read CA cert %s: %w", path, err)
		}

		if !cv.caCertPool.AppendCertsFromPEM(certPEM) {
			return fmt.Errorf("failed to parse CA cert: %s", path)
		}
	}

	cv.caCertPaths = paths
	return nil
}

func (cv *CertValidator) ValidateClientCert(cert *x509.Certificate) (*DeviceClaims, error) {
	cv.mu.RLock()
	caPool := cv.caCertPool
	skipVerify := cv.skipVerify
	cv.mu.RUnlock()

	if !skipVerify && caPool == nil {
		return nil, fmt.Errorf("no CA certificates configured")
	}

	if !skipVerify {
		if _, err := cert.Verify(x509.VerifyOptions{
			Roots: caPool,
		}); err != nil {
			return nil, fmt.Errorf("certificate verification failed: %w", err)
		}
	}

	fingerprint := cv.ComputeFingerprint(cert)

	var device Device
	if err := cv.db.Where("cert_fingerprint = ?", fingerprint).First(&device).Error; err != nil {
		if err == gorm.ErrRecordNotFound {
			return nil, fmt.Errorf("device not registered: %s", fingerprint)
		}
		return nil, fmt.Errorf("database error: %w", err)
	}

	claims := &DeviceClaims{
		DeviceID:    device.ID,
		TenantID:    device.TenantID,
		DeviceType:  device.DeviceType,
		Name:        device.Name,
		AuthMethod:  device.AuthMethod,
		Permissions: getPermissionsForDeviceType(device.DeviceType),
	}

	now := time.Now()
	device.LastSeen = &now
	cv.db.Save(&device)

	return claims, nil
}

func (cv *CertValidator) ValidateConnection(conn tls.ConnectionState) (*DeviceClaims, error) {
	if len(conn.VerifiedChains) == 0 {
		return nil, fmt.Errorf("no verified certificate chain")
	}

	leafCert := conn.VerifiedChains[0][0]
	return cv.ValidateClientCert(leafCert)
}

func (cv *CertValidator) ComputeFingerprint(cert *x509.Certificate) string {
	sum := sha256.Sum256(cert.Raw)
	return hex.EncodeToString(sum[:])
}

func (cv *CertValidator) RegisterDevice(db *gorm.DB, tenantID uuid.UUID, name string, deviceType DeviceType, cert *x509.Certificate) (*Device, error) {
	fingerprint := cv.ComputeFingerprint(cert)

	device := &Device{
		TenantID:        tenantID,
		Name:            name,
		DeviceType:      deviceType,
		AuthMethod:      AuthMethodX509,
		Status:          DeviceStatusOffline,
		CertFingerprint: fingerprint,
		Metadata: map[string]any{
			"subject": cert.Subject.String(),
			"issuer":  cert.Issuer.String(),
			"not_before": cert.NotBefore,
			"not_after":  cert.NotAfter,
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

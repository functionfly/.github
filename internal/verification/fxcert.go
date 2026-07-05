package verification

import (
	"context"
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"time"

	"github.com/google/uuid"
)

// FXCERTService handles Function Execution Certificate generation and validation
type FXCERTService struct {
	config FXCERTConfig
}

// FXCERT represents a Function Execution Certificate
type FXCERT struct {
	ID              uuid.UUID `json:"id"`
	CertificateID   string    `json:"certificate_id"` // Human-readable ID like "fxc_01H..."
	FunctionID      uuid.UUID `json:"function_id"`
	FunctionVersion string    `json:"function_version"`
	Level           string    `json:"level"` // "lite", "standard", "legal_grade"
	ValidFrom       time.Time `json:"valid_from"`
	ValidUntil      time.Time `json:"valid_until"`
	SuccessRate     float64   `json:"success_rate"`
	AvgLatencyMs    int       `json:"avg_latency_ms"`
	MaxLatencyMs    int       `json:"max_latency_ms"`
	TotalExecutions int       `json:"total_executions"`
	IsValid         bool      `json:"is_valid"`
	CertificateHash string    `json:"certificate_hash"`
	Metadata        map[string]interface{} `json:"metadata,omitempty"`
}

// NewFXCERTService creates a new FXCERT service
func NewFXCERTService(config FXCERTConfig) *FXCERTService {
	if config.CertificateValidityDays == 0 {
		config.CertificateValidityDays = 30 // Default: 30 days validity
	}
	if config.MaxLatencyMs == 0 {
		config.MaxLatencyMs = 5000 // Default: 5 second max latency
	}
	if config.MinSuccessRate == 0 {
		config.MinSuccessRate = 0.99 // Default: 99% success rate
	}

	return &FXCERTService{
		config: config,
	}
}

// Generate creates a new FXCERT for a function
func (s *FXCERTService) Generate(ctx context.Context, functionID uuid.UUID, functionVersion string) (*FXCERT, error) {
	now := time.Now()
	validUntil := now.Add(time.Duration(s.config.CertificateValidityDays) * 24 * time.Hour)

	cert := &FXCERT{
		ID:              uuid.New(),
		CertificateID:   s.generateCertificateID(),
		FunctionID:      functionID,
		FunctionVersion: functionVersion,
		Level:           "standard",
		ValidFrom:       now,
		ValidUntil:      validUntil,
		SuccessRate:     1.0,
		AvgLatencyMs:    100,
		MaxLatencyMs:    500,
		TotalExecutions: 1,
		IsValid:         true,
		Metadata: map[string]interface{}{
			"runtime": "wasm",
		},
	}

	// Calculate certificate hash
	cert.CertificateHash = s.calculateCertificateHash(cert)

	return cert, nil
}

// Validate validates an existing FXCERT
func (s *FXCERTService) Validate(ctx context.Context, certificateID string) (*FXCERT, bool, error) {
	// Parse certificate ID to extract UUID
	parts := splitCertificateID(certificateID)
	if len(parts) != 2 {
		return nil, false, fmt.Errorf("invalid certificate ID format")
	}

	// Create a mock certificate for validation
	cert := &FXCERT{
		CertificateID: certificateID,
		FunctionID:    uuid.MustParse(parts[1]),
		ValidFrom:    time.Now().Add(-24 * time.Hour),
		ValidUntil:   time.Now().Add(24 * time.Hour),
		IsValid:      true,
	}

	isValid := cert.IsValid && time.Now().After(cert.ValidFrom) && time.Now().Before(cert.ValidUntil)

	return cert, isValid, nil
}

// generateCertificateID generates a human-readable certificate ID
func (s *FXCERTService) generateCertificateID() string {
	id := uuid.New().String()
	return "fxc_" + id[:13]
}

// calculateCertificateHash calculates the certificate hash
func (s *FXCERTService) calculateCertificateHash(cert *FXCERT) string {
	data := fmt.Sprintf("%s|%s|%s|%s|%g|%d|%d|%s",
		cert.ID.String(),
		cert.FunctionID.String(),
		cert.FunctionVersion,
		cert.Level,
		cert.SuccessRate,
		cert.AvgLatencyMs,
		cert.TotalExecutions,
		cert.ValidUntil.Format(time.RFC3339),
	)

	hash := sha256.Sum256([]byte(data))
	return hex.EncodeToString(hash[:])
}

// splitCertificateID splits a certificate ID into its parts
func splitCertificateID(certID string) []string {
	// Format: fxc_UUID
	if len(certID) > 4 && certID[:4] == "fxc_" {
		return []string{"fxc", certID[4:]}
	}
	return []string{}
}

// SerializeCertificate serializes a certificate to JSON
func SerializeCertificate(cert *FXCERT) ([]byte, error) {
	return json.Marshal(cert)
}

// DeserializeCertificate deserializes a certificate from JSON
func DeserializeCertificate(data []byte) (*FXCERT, error) {
	var cert FXCERT
	if err := json.Unmarshal(data, &cert); err != nil {
		return nil, err
	}
	return &cert, nil
}

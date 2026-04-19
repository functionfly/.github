package billing

import (
	"context"
	"crypto/sha256"
	"encoding/hex"
	"fmt"
	"io"
	"path/filepath"
	"strings"
	"time"

	"github.com/functionfly/functionfly/internal/storage"
	"github.com/google/uuid"
	"github.com/sirupsen/logrus"
)

// TaxExemptionConfig contains configuration for tax exemption handling
type TaxExemptionConfig struct {
	// Allowed file types (mime types)
	AllowedFileTypes []string
	// Max file size in bytes (default 10MB)
	MaxFileSize int64
	// Storage path prefix
	StoragePath string
}

// DefaultTaxExemptionConfig returns the default configuration
func DefaultTaxExemptionConfig() *TaxExemptionConfig {
	return &TaxExemptionConfig{
		AllowedFileTypes: []string{
			"application/pdf",
			"image/jpeg",
			"image/png",
			"image/tiff",
		},
		MaxFileSize: 10 * 1024 * 1024, // 10MB
		StoragePath: "tax-exemptions/",
	}
}

// TaxExemptionService handles US tax exemption certificate uploads and reviews
type TaxExemptionService struct {
	config  *TaxExemptionConfig
	repo    *storage.BillingOperationalRepository
	storage ServiceStorage // Interface for file storage (S3, local, etc.)
}

// ServiceStorage defines the interface for file storage
// This would be implemented by S3Storage, LocalStorage, etc.
type ServiceStorage interface {
	Upload(ctx context.Context, key string, content io.Reader, contentType string, size int64) (string, error)
	Delete(ctx context.Context, key string) error
}

// NewTaxExemptionService creates a new tax exemption service
func NewTaxExemptionService(config *TaxExemptionConfig, repo *storage.BillingOperationalRepository, storage ServiceStorage) *TaxExemptionService {
	if config == nil {
		config = DefaultTaxExemptionConfig()
	}

	return &TaxExemptionService{
		config:  config,
		repo:    repo,
		storage: storage,
	}
}

// USStateCodes maps US state names to their 2-letter codes
var USStateCodes = map[string]string{
	"ALABAMA": "AL", "ALASKA": "AK", "ARIZONA": "AZ", "ARKANSAS": "AR",
	"CALIFORNIA": "CA", "COLORADO": "CO", "CONNECTICUT": "CT", "DELAWARE": "DE",
	"FLORIDA": "FL", "GEORGIA": "GA", "HAWAII": "HI", "IDAHO": "ID",
	"ILLINOIS": "IL", "INDIANA": "IN", "IOWA": "IA", "KANSAS": "KS",
	"KENTUCKY": "KY", "LOUISIANA": "LA", "MAINE": "ME", "MARYLAND": "MD",
	"MASSACHUSETTS": "MA", "MICHIGAN": "MI", "MINNESOTA": "MN", "MISSISSIPPI": "MS",
	"MISSOURI": "MO", "MONTANA": "MT", "NEBRASKA": "NE", "NEVADA": "NV",
	"NEW HAMPSHIRE": "NH", "NEW JERSEY": "NJ", "NEW MEXICO": "NM", "NEW YORK": "NY",
	"NORTH CAROLINA": "NC", "NORTH DAKOTA": "ND", "OHIO": "OH", "OKLAHOMA": "OK",
	"OREGON": "OR", "PENNSYLVANIA": "PA", "RHODE ISLAND": "RI", "SOUTH CAROLINA": "SC",
	"SOUTH DAKOTA": "SD", "TENNESSEE": "TN", "TEXAS": "TX", "UTAH": "UT",
	"VERMONT": "VT", "VIRGINIA": "VA", "WASHINGTON": "WA", "WEST VIRGINIA": "WV",
	"WISCONSIN": "WI", "WYOMING": "WY", "DISTRICT OF COLUMBIA": "DC",
}

// ExemptionTypes defines valid exemption types
type ExemptionType string

const (
	ExemptionTypeResale       ExemptionType = "resale"
	ExemptionTypeNonprofit    ExemptionType = "nonprofit"
	ExemptionTypeGovernment   ExemptionType = "government"
	ExemptionTypeAgricultural ExemptionType = "agricultural"
	ExemptionTypeIndustrial   ExemptionType = "industrial"
	ExemptionTypeDirectPay    ExemptionType = "direct_pay"
	ExemptionTypeOther        ExemptionType = "other"
)

// ValidExemptionTypes lists all valid exemption types
var ValidExemptionTypes = []ExemptionType{
	ExemptionTypeResale, ExemptionTypeNonprofit, ExemptionTypeGovernment,
	ExemptionTypeAgricultural, ExemptionTypeIndustrial, ExemptionTypeDirectPay, ExemptionTypeOther,
}

// UploadCertificateRequest contains the data for uploading a certificate
type UploadCertificateRequest struct {
	TenantID          uuid.UUID
	UserID            uuid.UUID
	CertificateNumber string
	State             string // 2-letter code or full name
	ExemptionType     string
	ExemptionReason   string
	ValidFrom         time.Time
	ValidUntil        *time.Time
	FileContent       io.Reader
	FileName          string
	FileSize          int64
	ContentType       string
}

// UploadCertificate handles the upload of a tax exemption certificate
func (s *TaxExemptionService) UploadCertificate(ctx context.Context, req UploadCertificateRequest) (*storage.TaxExemptionCertificate, error) {
	// Validate file type
	if !s.isAllowedFileType(req.ContentType) {
		return nil, fmt.Errorf("invalid file type: %s. Allowed types: %v", req.ContentType, s.config.AllowedFileTypes)
	}

	// Validate file size
	if req.FileSize > s.config.MaxFileSize {
		return nil, fmt.Errorf("file too large: %d bytes (max %d)", req.FileSize, s.config.MaxFileSize)
	}

	// Normalize state code
	stateCode := s.normalizeStateCode(req.State)
	if stateCode == "" {
		return nil, fmt.Errorf("invalid state: %s", req.State)
	}

	// Validate exemption type
	exemptionType := ExemptionType(strings.ToLower(req.ExemptionType))
	if !s.isValidExemptionType(exemptionType) {
		return nil, fmt.Errorf("invalid exemption type: %s. Valid types: %v", req.ExemptionType, ValidExemptionTypes)
	}

	// Calculate file hash for integrity
	contentBytes, err := io.ReadAll(req.FileContent)
	if err != nil {
		return nil, fmt.Errorf("failed to read file content: %w", err)
	}
	fileHash := sha256.Sum256(contentBytes)
	fileHashStr := hex.EncodeToString(fileHash[:])

	// Upload to storage
	storageKey := s.generateStorageKey(req.TenantID, req.FileName)
	fileURL, err := s.storage.Upload(ctx, storageKey, strings.NewReader(string(contentBytes)), req.ContentType, req.FileSize)
	if err != nil {
		return nil, fmt.Errorf("failed to upload file: %w", err)
	}

	// Create certificate record
	cert := &storage.TaxExemptionCertificate{
		TenantID:          req.TenantID,
		UserID:            req.UserID,
		CertificateNumber: req.CertificateNumber,
		State:             stateCode,
		ExemptionType:     string(exemptionType),
		ExemptionReason:   req.ExemptionReason,
		FileURL:           fileURL,
		FileName:          req.FileName,
		FileSize:          req.FileSize,
		FileHash:          fileHashStr,
		ValidFrom:         req.ValidFrom,
		ValidUntil:        req.ValidUntil,
		Status:            "pending",
	}

	// Store in database
	if s.repo != nil {
		stored, err := s.repo.CreateTaxExemptionCertificate(ctx, cert)
		if err != nil {
			// Attempt to delete uploaded file
			if delErr := s.storage.Delete(ctx, storageKey); delErr != nil {
				logrus.WithError(delErr).Warn("Failed to clean up uploaded file after DB error")
			}
			return nil, fmt.Errorf("failed to store certificate: %w", err)
		}
		cert = stored
	}

	logrus.WithFields(logrus.Fields{
		"certificate_id": cert.ID,
		"tenant_id":      req.TenantID,
		"state":          stateCode,
		"exemption_type": exemptionType,
	}).Info("Tax exemption certificate uploaded and pending review")

	return cert, nil
}

// ReviewCertificate reviews a pending tax exemption certificate
func (s *TaxExemptionService) ReviewCertificate(ctx context.Context, certificateID uuid.UUID, reviewerID uuid.UUID, approved bool, notes, rejectionReason string) (*storage.TaxExemptionCertificate, error) {
	if s.repo == nil {
		return nil, fmt.Errorf("repository not configured")
	}

	// Get the certificate
	cert, err := s.repo.GetTaxExemptionCertificate(ctx, certificateID)
	if err != nil {
		return nil, fmt.Errorf("failed to get certificate: %w", err)
	}
	if cert == nil {
		return nil, fmt.Errorf("certificate not found")
	}

	// Validate state transition
	if cert.Status != "pending" {
		return nil, fmt.Errorf("cannot review certificate with status: %s (must be pending)", cert.Status)
	}

	// Validate rejection reason
	if !approved && rejectionReason == "" {
		return nil, fmt.Errorf("rejection reason is required when rejecting a certificate")
	}

	// Perform review
	reviewedCert, err := s.repo.ReviewTaxExemptionCertificate(ctx, certificateID, reviewerID, approved, notes, rejectionReason)
	if err != nil {
		return nil, fmt.Errorf("failed to review certificate: %w", err)
	}

	logrus.WithFields(logrus.Fields{
		"certificate_id": certificateID,
		"approved":       approved,
		"reviewed_by":    reviewerID,
		"tenant_id":      cert.TenantID,
	}).Info("Tax exemption certificate reviewed")

	return reviewedCert, nil
}

// GetCertificate retrieves a tax exemption certificate by ID
func (s *TaxExemptionService) GetCertificate(ctx context.Context, certificateID uuid.UUID) (*storage.TaxExemptionCertificate, error) {
	if s.repo == nil {
		return nil, fmt.Errorf("repository not configured")
	}

	return s.repo.GetTaxExemptionCertificate(ctx, certificateID)
}

// ListTenantCertificates lists all certificates for a tenant
func (s *TaxExemptionService) ListTenantCertificates(ctx context.Context, tenantID uuid.UUID, status string) ([]storage.TaxExemptionCertificate, error) {
	if s.repo == nil {
		return nil, fmt.Errorf("repository not configured")
	}

	return s.repo.ListTaxExemptionCertificates(ctx, tenantID, status)
}

// ListPendingCertificates lists all pending certificates (for admin review)
func (s *TaxExemptionService) ListPendingCertificates(ctx context.Context, limit, offset int) ([]storage.TaxExemptionCertificate, int64, error) {
	if s.repo == nil {
		return nil, 0, fmt.Errorf("repository not configured")
	}

	return s.repo.ListPendingTaxExemptionCertificates(ctx, limit, offset)
}

// CheckCertificateExpiry marks expired certificates
func (s *TaxExemptionService) CheckCertificateExpiry(ctx context.Context) (int64, error) {
	if s.repo == nil {
		return 0, fmt.Errorf("repository not configured")
	}

	expiredCount, err := s.repo.MarkTaxExemptionCertificateExpired(ctx)
	if err != nil {
		return 0, fmt.Errorf("failed to mark expired certificates: %w", err)
	}

	if expiredCount > 0 {
		logrus.WithField("expired_count", expiredCount).Info("Marked expired tax exemption certificates")
	}

	return expiredCount, nil
}

// SyncApprovedCertificateToStripe syncs an approved certificate to Stripe
func (s *TaxExemptionService) SyncApprovedCertificateToStripe(ctx context.Context, certificateID uuid.UUID, stripeExemptionID string) error {
	if s.repo == nil {
		return fmt.Errorf("repository not configured")
	}

	// Get the certificate
	cert, err := s.repo.GetTaxExemptionCertificate(ctx, certificateID)
	if err != nil {
		return fmt.Errorf("failed to get certificate: %w", err)
	}
	if cert == nil {
		return fmt.Errorf("certificate not found")
	}

	if cert.Status != "approved" {
		return fmt.Errorf("cannot sync certificate with status: %s (must be approved)", cert.Status)
	}

	// Update with Stripe exemption ID
	if err := s.repo.UpdateTaxExemptionStripeID(ctx, certificateID, stripeExemptionID); err != nil {
		return fmt.Errorf("failed to update certificate with Stripe ID: %w", err)
	}

	logrus.WithFields(logrus.Fields{
		"certificate_id":      certificateID,
		"stripe_exemption_id": stripeExemptionID,
	}).Info("Tax exemption certificate synced to Stripe")

	return nil
}

// Helper methods

func (s *TaxExemptionService) isAllowedFileType(contentType string) bool {
	for _, allowed := range s.config.AllowedFileTypes {
		if strings.EqualFold(contentType, allowed) {
			return true
		}
	}
	return false
}

func (s *TaxExemptionService) isValidExemptionType(exemptionType ExemptionType) bool {
	for _, valid := range ValidExemptionTypes {
		if exemptionType == valid {
			return true
		}
	}
	return false
}

func (s *TaxExemptionService) normalizeStateCode(state string) string {
	// If already 2 characters, validate it's a valid state code
	state = strings.ToUpper(strings.TrimSpace(state))
	if len(state) == 2 {
		// Check if it's a valid state code
		for _, code := range USStateCodes {
			if state == code {
				return state
			}
		}
		return ""
	}

	// Try to look up full state name
	if code, ok := USStateCodes[strings.ToUpper(state)]; ok {
		return code
	}

	return ""
}

func (s *TaxExemptionService) generateStorageKey(tenantID uuid.UUID, fileName string) string {
	ext := filepath.Ext(fileName)
	timestamp := time.Now().UTC().Format("20060102_150405")
	return fmt.Sprintf("%s%s/%s_%s%s",
		s.config.StoragePath,
		tenantID.String(),
		uuid.New().String()[:8],
		timestamp,
		ext,
	)
}

// ValidateCertificateNumber validates the format of a certificate number
// Different states have different formats - this is a basic validation
func ValidateCertificateNumber(number string) error {
	if strings.TrimSpace(number) == "" {
		return fmt.Errorf("certificate number is required")
	}
	if len(number) < 3 || len(number) > 50 {
		return fmt.Errorf("certificate number must be between 3 and 50 characters")
	}
	return nil
}

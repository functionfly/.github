package billing

import (
	"context"
	"strings"
	"testing"
	"time"

	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestUSStateCodes(t *testing.T) {
	tests := []struct {
		name     string
		input    string
		expected string
		valid    bool
	}{
		{"CA", "CA", "CA", true},
		{"California", "California", "CA", true},
		{"california lowercase", "california", "CA", true},
		{"TX", "TX", "TX", true},
		{"invalid", "XX", "", false},
		{"empty", "", "", false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			svc := &TaxExemptionService{
				config: DefaultTaxExemptionConfig(),
			}

			result := svc.normalizeStateCode(tt.input)
			if tt.valid {
				assert.Equal(t, tt.expected, result)
			} else {
				assert.Empty(t, result)
			}
		})
	}
}

func TestValidateCertificateNumber(t *testing.T) {
	tests := []struct {
		name    string
		number  string
		wantErr bool
	}{
		{"valid", "123456789", false},
		{"min length", "123", false},
		{"max length", strings.Repeat("1", 50), false},
		{"too short", "12", true},
		{"too long", strings.Repeat("1", 51), true},
		{"empty", "", true},
		{"whitespace only", "   ", true},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := ValidateCertificateNumber(tt.number)
			if tt.wantErr {
				assert.Error(t, err)
			} else {
				assert.NoError(t, err)
			}
		})
	}
}

func TestTaxExemptionConfig_Defaults(t *testing.T) {
	config := DefaultTaxExemptionConfig()

	assert.Equal(t, int64(10*1024*1024), config.MaxFileSize)
	assert.Equal(t, "tax-exemptions/", config.StoragePath)
	assert.Contains(t, config.AllowedFileTypes, "application/pdf")
	assert.Contains(t, config.AllowedFileTypes, "image/jpeg")
	assert.Contains(t, config.AllowedFileTypes, "image/png")
}

func TestIsAllowedFileType(t *testing.T) {
	svc := &TaxExemptionService{
		config: DefaultTaxExemptionConfig(),
	}

	tests := []struct {
		contentType string
		allowed     bool
	}{
		{"application/pdf", true},
		{"image/jpeg", true},
		{"image/png", true},
		{"image/tiff", true},
		{"text/plain", false},
		{"application/json", false},
		{"", false},
	}

	for _, tt := range tests {
		t.Run(tt.contentType, func(t *testing.T) {
			result := svc.isAllowedFileType(tt.contentType)
			assert.Equal(t, tt.allowed, result)
		})
	}
}

func TestIsValidExemptionType(t *testing.T) {
	svc := &TaxExemptionService{
		config: DefaultTaxExemptionConfig(),
	}

	tests := []struct {
		exemptionType ExemptionType
		valid         bool
	}{
		{ExemptionTypeResale, true},
		{ExemptionTypeNonprofit, true},
		{ExemptionTypeGovernment, true},
		{ExemptionTypeAgricultural, true},
		{ExemptionTypeIndustrial, true},
		{ExemptionTypeDirectPay, true},
		{ExemptionTypeOther, true},
		{"invalid", false},
	}

	for _, tt := range tests {
		t.Run(string(tt.exemptionType), func(t *testing.T) {
			result := svc.isValidExemptionType(tt.exemptionType)
			assert.Equal(t, tt.valid, result)
		})
	}
}

func TestGenerateStorageKey(t *testing.T) {
	svc := &TaxExemptionService{
		config: DefaultTaxExemptionConfig(),
	}

	tenantID := uuid.MustParse("12345678-1234-1234-1234-123456789012")
	fileName := "certificate.pdf"

	key := svc.generateStorageKey(tenantID, fileName)

	assert.Contains(t, key, "tax-exemptions/")
	assert.Contains(t, key, tenantID.String())
	assert.Contains(t, key, ".pdf")
	assert.True(t, strings.HasSuffix(key, ".pdf"))
}

func TestGenerateStorageKey_DifferentExtensions(t *testing.T) {
	svc := &TaxExemptionService{
		config: DefaultTaxExemptionConfig(),
	}

	tenantID := uuid.MustParse("12345678-1234-1234-1234-123456789012")

	tests := []struct {
		fileName string
		ext      string
	}{
		{"cert.pdf", ".pdf"},
		{"cert.jpg", ".jpg"},
		{"cert.png", ".png"},
		{"cert.tiff", ".tiff"},
	}

	for _, tt := range tests {
		t.Run(tt.fileName, func(t *testing.T) {
			key := svc.generateStorageKey(tenantID, tt.fileName)
			assert.True(t, strings.HasSuffix(key, tt.ext))
		})
	}
}

func TestValidExemptionTypes(t *testing.T) {
	assert.Len(t, ValidExemptionTypes, 7)
	assert.Contains(t, ValidExemptionTypes, ExemptionTypeResale)
	assert.Contains(t, ValidExemptionTypes, ExemptionTypeNonprofit)
	assert.Contains(t, ValidExemptionTypes, ExemptionTypeGovernment)
	assert.Contains(t, ValidExemptionTypes, ExemptionTypeAgricultural)
	assert.Contains(t, ValidExemptionTypes, ExemptionTypeIndustrial)
	assert.Contains(t, ValidExemptionTypes, ExemptionTypeDirectPay)
	assert.Contains(t, ValidExemptionTypes, ExemptionTypeOther)
}

func TestUploadCertificateRequest_Validation(t *testing.T) {
	tests := []struct {
		name    string
		req     UploadCertificateRequest
		wantErr bool
		errMsg  string
	}{
		{
			name: "valid request",
			req: UploadCertificateRequest{
				TenantID:          uuid.New(),
				UserID:            uuid.New(),
				CertificateNumber: "ABC123",
				State:             "CA",
				ExemptionType:     "resale",
				ExemptionReason:   "Resale purposes",
				ValidFrom:         time.Now(),
				FileContent:       strings.NewReader("test content"),
				FileName:          "cert.pdf",
				FileSize:          1024,
				ContentType:       "application/pdf",
			},
			wantErr: false,
		},
		{
			name: "invalid file type",
			req: UploadCertificateRequest{
				TenantID:          uuid.New(),
				UserID:            uuid.New(),
				CertificateNumber: "ABC123",
				State:             "CA",
				ExemptionType:     "resale",
				ExemptionReason:   "Resale purposes",
				ValidFrom:         time.Now(),
				FileContent:       strings.NewReader("test content"),
				FileName:          "cert.txt",
				FileSize:          1024,
				ContentType:       "text/plain",
			},
			wantErr: true,
			errMsg:  "invalid file type",
		},
		{
			name: "file too large",
			req: UploadCertificateRequest{
				TenantID:          uuid.New(),
				UserID:            uuid.New(),
				CertificateNumber: "ABC123",
				State:             "CA",
				ExemptionType:     "resale",
				ExemptionReason:   "Resale purposes",
				ValidFrom:         time.Now(),
				FileContent:       strings.NewReader("test content"),
				FileName:          "cert.pdf",
				FileSize:          20 * 1024 * 1024,
				ContentType:       "application/pdf",
			},
			wantErr: true,
			errMsg:  "file too large",
		},
		{
			name: "invalid state",
			req: UploadCertificateRequest{
				TenantID:          uuid.New(),
				UserID:            uuid.New(),
				CertificateNumber: "ABC123",
				State:             "XX",
				ExemptionType:     "resale",
				ExemptionReason:   "Resale purposes",
				ValidFrom:         time.Now(),
				FileContent:       strings.NewReader("test content"),
				FileName:          "cert.pdf",
				FileSize:          1024,
				ContentType:       "application/pdf",
			},
			wantErr: true,
			errMsg:  "invalid state",
		},
		{
			name: "invalid exemption type",
			req: UploadCertificateRequest{
				TenantID:          uuid.New(),
				UserID:            uuid.New(),
				CertificateNumber: "ABC123",
				State:             "CA",
				ExemptionType:     "invalid_type",
				ExemptionReason:   "Resale purposes",
				ValidFrom:         time.Now(),
				FileContent:       strings.NewReader("test content"),
				FileName:          "cert.pdf",
				FileSize:          1024,
				ContentType:       "application/pdf",
			},
			wantErr: true,
			errMsg:  "invalid exemption type",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			svc := &TaxExemptionService{
				config:  DefaultTaxExemptionConfig(),
				repo:    nil,
				storage: &mockStorage{},
			}

			err := svc.validateUploadRequest(tt.req)
			if tt.wantErr {
				assert.Error(t, err)
				if tt.errMsg != "" {
					assert.Contains(t, err.Error(), tt.errMsg)
				}
			} else {
				assert.NoError(t, err)
			}
		})
	}
}

func (s *TaxExemptionService) validateUploadRequest(req UploadCertificateRequest) error {
	if !s.isAllowedFileType(req.ContentType) {
		return &validationError{"invalid file type"}
	}

	if req.FileSize > s.config.MaxFileSize {
		return &validationError{"file too large"}
	}

	stateCode := s.normalizeStateCode(req.State)
	if stateCode == "" {
		return &validationError{"invalid state"}
	}

	exemptionType := ExemptionType(strings.ToLower(req.ExemptionType))
	if !s.isValidExemptionType(exemptionType) {
		return &validationError{"invalid exemption type"}
	}

	return nil
}

type validationError struct {
	message string
}

func (e *validationError) Error() string {
	return e.message
}

type mockStorage struct{}

func (m *mockStorage) Upload(ctx context.Context, key string, content interface{}, contentType string, size int64) (string, error) {
	return "https://storage.example.com/" + key, nil
}

func (m *mockStorage) Delete(ctx context.Context, key string) error {
	return nil
}

func TestServiceStorage_Interface(t *testing.T) {
	var _ ServiceStorage = &mockStorage{}
}

func TestReviewCertificate_InvalidState(t *testing.T) {
	svc := &TaxExemptionService{
		config: DefaultTaxExemptionConfig(),
		repo:   nil,
	}

	certID := uuid.New()
	reviewerID := uuid.New()

	_, err := svc.ReviewCertificate(context.Background(), certID, reviewerID, true, "", "")
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "repository not configured")
}

func TestGetCertificate_NotConfigured(t *testing.T) {
	svc := &TaxExemptionService{
		config: DefaultTaxExemptionConfig(),
		repo:   nil,
	}

	_, err := svc.GetCertificate(context.Background(), uuid.New())
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "repository not configured")
}

func TestListTenantCertificates_NotConfigured(t *testing.T) {
	svc := &TaxExemptionService{
		config: DefaultTaxExemptionConfig(),
		repo:   nil,
	}

	_, err := svc.ListTenantCertificates(context.Background(), uuid.New(), "")
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "repository not configured")
}

func TestListPendingCertificates_NotConfigured(t *testing.T) {
	svc := &TaxExemptionService{
		config: DefaultTaxExemptionConfig(),
		repo:   nil,
	}

	_, _, err := svc.ListPendingCertificates(context.Background(), 10, 0)
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "repository not configured")
}

func TestCheckCertificateExpiry_NotConfigured(t *testing.T) {
	svc := &TaxExemptionService{
		config: DefaultTaxExemptionConfig(),
		repo:   nil,
	}

	_, err := svc.CheckCertificateExpiry(context.Background())
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "repository not configured")
}

func TestSyncApprovedCertificateToStripe_NotConfigured(t *testing.T) {
	svc := &TaxExemptionService{
		config: DefaultTaxExemptionConfig(),
		repo:   nil,
	}

	err := svc.SyncApprovedCertificateToStripe(context.Background(), uuid.New(), "stripe_123")
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "repository not configured")
}

func TestNewTaxExemptionService_DefaultConfig(t *testing.T) {
	svc := NewTaxExemptionService(nil, nil, nil)
	assert.NotNil(t, svc)
	assert.NotNil(t, svc.config)
	assert.Equal(t, int64(10*1024*1024), svc.config.MaxFileSize)
}

func TestTaxExemptionConfig_CustomValues(t *testing.T) {
	config := &TaxExemptionConfig{
		AllowedFileTypes: []string{"application/pdf"},
		MaxFileSize:      5 * 1024 * 1024,
		StoragePath:      "custom/path/",
	}

	svc := NewTaxExemptionService(config, nil, nil)
	assert.Equal(t, int64(5*1024*1024), svc.config.MaxFileSize)
	assert.Equal(t, "custom/path/", svc.config.StoragePath)
	assert.Len(t, svc.config.AllowedFileTypes, 1)
}

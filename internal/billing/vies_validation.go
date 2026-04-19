package billing

import (
	"context"
	"encoding/xml"
	"fmt"
	"io"
	"net/http"
	"regexp"
	"strings"
	"time"

	"github.com/functionfly/functionfly/internal/storage"
	"github.com/google/uuid"
	"github.com/sirupsen/logrus"
)

// VIESConfig contains configuration for the VIES API client
type VIESConfig struct {
	BaseURL      string
	Timeout      time.Duration
	MaxRetries   int
	RetryBackoff time.Duration
}

// DefaultVIESConfig returns the default VIES configuration
func DefaultVIESConfig() *VIESConfig {
	return &VIESConfig{
		BaseURL:      "http://ec.europa.eu/taxation_customs/vies/services/checkVatService",
		Timeout:      30 * time.Second,
		MaxRetries:   3,
		RetryBackoff: 5 * time.Second,
	}
}

// VIESCheckVatResponse represents the VIES API response
type VIESCheckVatResponse struct {
	XMLName     xml.Name `xml:"http://schemas.xmlsoap.org/soap/envelope/ Envelope"`
	Body        VIESBody `xml:"Body"`
	RequestDate string   `xml:"Body>checkVatResponse>requestDate"`
}

// VIESBody represents the SOAP body
type VIESBody struct {
	XMLName          xml.Name                 `xml:"Body"`
	CheckVatResponse VIESCheckVatResponseData `xml:"checkVatResponse"`
	Fault            *VIESFault               `xml:"Fault,omitempty"`
}

// VIESCheckVatResponseData contains the VIES response data
type VIESCheckVatResponseData struct {
	CountryCode string `xml:"countryCode"`
	VatNumber   string `xml:"vatNumber"`
	RequestDate string `xml:"requestDate"`
	Valid       bool   `xml:"valid"`
	Name        string `xml:"name"`
	Address     string `xml:"address"`
}

// VIESFault represents a SOAP fault
type VIESFault struct {
	FaultCode   string `xml:"faultcode"`
	FaultString string `xml:"faultstring"`
}

// VIESClient validates EU VAT IDs using the VIES API
type VIESClient struct {
	config *VIESConfig
	client *http.Client
	repo   *storage.BillingOperationalRepository
}

// NewVIESClient creates a new VIES client
func NewVIESClient(config *VIESConfig, repo *storage.BillingOperationalRepository) *VIESClient {
	if config == nil {
		config = DefaultVIESConfig()
	}

	return &VIESClient{
		config: config,
		client: &http.Client{
			Timeout: config.Timeout,
		},
		repo: repo,
	}
}

// EUCountryCodes contains valid EU country codes for VAT validation
var EUCountryCodes = map[string]bool{
	"AT": true, // Austria
	"BE": true, // Belgium
	"BG": true, // Bulgaria
	"CY": true, // Cyprus
	"CZ": true, // Czech Republic
	"DE": true, // Germany
	"DK": true, // Denmark
	"EE": true, // Estonia
	"EL": true, // Greece
	"ES": true, // Spain
	"FI": true, // Finland
	"FR": true, // France
	"HR": true, // Croatia
	"HU": true, // Hungary
	"IE": true, // Ireland
	"IT": true, // Italy
	"LT": true, // Lithuania
	"LU": true, // Luxembourg
	"LV": true, // Latvia
	"MT": true, // Malta
	"NL": true, // Netherlands
	"PL": true, // Poland
	"PT": true, // Portugal
	"RO": true, // Romania
	"SE": true, // Sweden
	"SI": true, // Slovenia
	"SK": true, // Slovakia
	"XI": true, // Northern Ireland
}

// ValidateVATID validates an EU VAT ID against the VIES API
func (c *VIESClient) ValidateVATID(ctx context.Context, tenantID, userID uuid.UUID, vatID string) (*storage.EUVATValidation, error) {
	// Parse and validate the VAT ID format
	countryCode, vatNumber, err := c.parseVATID(vatID)
	if err != nil {
		return nil, fmt.Errorf("invalid VAT ID format: %w", err)
	}

	// Check if it's a valid EU country code
	if !EUCountryCodes[countryCode] {
		return nil, fmt.Errorf("invalid EU country code: %s", countryCode)
	}

	// Create initial validation record
	validation := &storage.EUVATValidation{
		TenantID:    tenantID,
		UserID:      userID,
		VATID:       vatID,
		CountryCode: countryCode,
		RequestDate: time.Now(),
		Status:      "pending",
		RetryCount:  0,
	}

	// Store the validation request
	if c.repo != nil {
		stored, err := c.repo.CreateEUVATValidation(ctx, validation)
		if err != nil {
			logrus.WithError(err).Warn("Failed to store VAT validation request")
		} else {
			validation = stored
		}
	}

	// Call VIES API with retry logic
	response, err := c.callVIESWithRetry(ctx, countryCode, vatNumber)
	if err != nil {
		validation.Status = "error"
		validation.ErrorCode = "vies_error"
		validation.ErrorMessage = err.Error()

		if c.repo != nil {
			if updateErr := c.repo.UpdateEUVATValidationStatus(ctx, validation.ID, "error", "vies_error", err.Error()); updateErr != nil {
				logrus.WithError(updateErr).Warn("Failed to update VAT validation status")
			}
		}
		return validation, err
	}

	// Check for SOAP fault
	if response.Body.Fault != nil {
		validation.Status = "error"
		validation.ErrorCode = response.Body.Fault.FaultCode
		validation.ErrorMessage = response.Body.Fault.FaultString

		if c.repo != nil {
			if updateErr := c.repo.UpdateEUVATValidationStatus(ctx, validation.ID, "error",
				response.Body.Fault.FaultCode, response.Body.Fault.FaultString); updateErr != nil {
				logrus.WithError(updateErr).Warn("Failed to update VAT validation status")
			}
		}
		return validation, fmt.Errorf("VIES API error: %s - %s",
			response.Body.Fault.FaultCode, response.Body.Fault.FaultString)
	}

	// Update validation with successful response
	vatResponse := response.Body.CheckVatResponse
	validation.IsValid = vatResponse.Valid
	validation.Status = map[bool]string{true: "valid", false: "invalid"}[vatResponse.Valid]
	validation.VIESTraderName = vatResponse.Name
	validation.VIESTraderAddress = vatResponse.Address

	if c.repo != nil {
		if updateErr := c.repo.UpdateEUVATValidationVIESResponse(ctx, validation.ID,
			vatResponse.Valid, "", "", vatResponse.Name, vatResponse.Address); updateErr != nil {
			logrus.WithError(updateErr).Warn("Failed to update VAT validation response")
		}
	}

	return validation, nil
}

// parseVATID parses a VAT ID into country code and number
func (c *VIESClient) parseVATID(vatID string) (countryCode, vatNumber string, err error) {
	// Remove spaces and convert to uppercase
	vatID = strings.ToUpper(strings.ReplaceAll(vatID, " ", ""))

	// VAT ID should be at least 4 characters (2 for country + 2 for number)
	if len(vatID) < 4 {
		return "", "", fmt.Errorf("VAT ID too short")
	}

	// Extract country code (first 2 characters)
	countryCode = vatID[:2]
	vatNumber = vatID[2:]

	// Validate country code format (letters only)
	if !regexp.MustCompile("^[A-Z]{2}$").MatchString(countryCode) {
		return "", "", fmt.Errorf("invalid country code format")
	}

	return countryCode, vatNumber, nil
}

// callVIESWithRetry calls the VIES API with retry logic
func (c *VIESClient) callVIESWithRetry(ctx context.Context, countryCode, vatNumber string) (*VIESCheckVatResponse, error) {
	var lastErr error

	for attempt := 0; attempt <= c.config.MaxRetries; attempt++ {
		if attempt > 0 {
			// Wait before retry
			select {
			case <-time.After(c.config.RetryBackoff * time.Duration(attempt)):
			case <-ctx.Done():
				return nil, ctx.Err()
			}
		}

		response, err := c.callVIES(ctx, countryCode, vatNumber)
		if err == nil {
			return response, nil
		}

		lastErr = err
		logrus.WithFields(logrus.Fields{
			"attempt":     attempt + 1,
			"max_retries": c.config.MaxRetries,
			"error":       err.Error(),
		}).Warn("VIES API call failed, will retry")
	}

	return nil, fmt.Errorf("VIES API call failed after %d attempts: %w", c.config.MaxRetries+1, lastErr)
}

// callVIES makes a single call to the VIES API
func (c *VIESClient) callVIES(ctx context.Context, countryCode, vatNumber string) (*VIESCheckVatResponse, error) {
	// Build SOAP request
	envelope := fmt.Sprintf(`<?xml version="1.0" encoding="UTF-8"?>
<soapenv:Envelope xmlns:soapenv="http://schemas.xmlsoap.org/soap/envelope/" xmlns:urn="urn:ec.europa.eu:taxud:vies:services:checkVat:types">
   <soapenv:Header/>
   <soapenv:Body>
      <urn:checkVat>
         <urn:countryCode>%s</urn:countryCode>
         <urn:vatNumber>%s</urn:vatNumber>
      </urn:checkVat>
   </soapenv:Body>
</soapenv:Envelope>`, countryCode, vatNumber)

	// Create request
	req, err := http.NewRequestWithContext(ctx, "POST", c.config.BaseURL, strings.NewReader(envelope))
	if err != nil {
		return nil, fmt.Errorf("failed to create request: %w", err)
	}

	req.Header.Set("Content-Type", "text/xml; charset=utf-8")
	req.Header.Set("SOAPAction", "")

	// Make request
	resp, err := c.client.Do(req)
	if err != nil {
		return nil, fmt.Errorf("VIES request failed: %w", err)
	}
	defer resp.Body.Close()

	// Read response body
	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("failed to read response body: %w", err)
	}

	// Check HTTP status
	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("VIES API returned HTTP %d: %s", resp.StatusCode, string(body))
	}

	// Parse SOAP response
	var viesResponse VIESCheckVatResponse
	if err := xml.Unmarshal(body, &viesResponse); err != nil {
		return nil, fmt.Errorf("failed to parse VIES response: %w", err)
	}

	return &viesResponse, nil
}

// ApplyValidVATToTenant applies a validated VAT ID to tenant tax settings
func (c *VIESClient) ApplyValidVATToTenant(ctx context.Context, validationID uuid.UUID) error {
	if c.repo == nil {
		return fmt.Errorf("repository not configured")
	}

	// Get the validation
	validation, err := c.repo.GetEUVATValidation(ctx, validationID)
	if err != nil {
		return fmt.Errorf("failed to get VAT validation: %w", err)
	}
	if validation == nil {
		return fmt.Errorf("VAT validation not found")
	}

	// Only apply if valid
	if !validation.IsValid {
		return fmt.Errorf("VAT ID is not valid")
	}

	// Mark as applied
	if err := c.repo.MarkVATValidationAppliedToSettings(ctx, validationID); err != nil {
		return fmt.Errorf("failed to mark VAT validation as applied: %w", err)
	}

	logrus.WithFields(logrus.Fields{
		"validation_id": validationID,
		"tenant_id":     validation.TenantID,
		"vat_id":        validation.VATID,
	}).Info("Applied valid VAT ID to tenant settings")

	return nil
}

// RetryFailedValidations retries failed VIES validations
func (c *VIESClient) RetryFailedValidations(ctx context.Context) error {
	if c.repo == nil {
		return fmt.Errorf("repository not configured")
	}

	// Get pending validations
	validations, err := c.repo.GetPendingVATValidations(ctx, 50)
	if err != nil {
		return fmt.Errorf("failed to get pending validations: %w", err)
	}

	for _, validation := range validations {
		// Skip if max retries exceeded
		if validation.RetryCount >= 5 {
			logrus.WithField("validation_id", validation.ID).Warn("Max retries exceeded for VAT validation")
			continue
		}

		// Retry validation
		_, err := c.ValidateVATID(ctx, validation.TenantID, validation.UserID, validation.VATID)
		if err != nil {
			logrus.WithError(err).WithField("validation_id", validation.ID).Warn("Retry failed for VAT validation")

			// Schedule next retry
			nextRetry := time.Now().Add(c.config.RetryBackoff * time.Duration(validation.RetryCount+1))
			if err := c.repo.ScheduleVATValidationRetry(ctx, validation.ID, validation.RetryCount+1, nextRetry); err != nil {
				logrus.WithError(err).Warn("Failed to schedule VAT validation retry")
			}
		}
	}

	return nil
}

// ExtractCountryCode extracts the country code from a VAT ID
func ExtractCountryCode(vatID string) string {
	vatID = strings.ToUpper(strings.ReplaceAll(vatID, " ", ""))
	if len(vatID) < 2 {
		return ""
	}
	return vatID[:2]
}

// IsEUVATID checks if a VAT ID is from an EU country
func IsEUVATID(vatID string) bool {
	countryCode := ExtractCountryCode(vatID)
	return EUCountryCodes[countryCode]
}

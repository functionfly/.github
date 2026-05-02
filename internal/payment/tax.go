package payment

import (
	"context"
	"fmt"
	"regexp"
	"strconv"
	"strings"

	"github.com/functionfly/functionfly/internal/storage"
	"github.com/google/uuid"
	"github.com/stripe/stripe-go/v83"
	"github.com/stripe/stripe-go/v83/tax/calculation"
)

// TaxService provides tax calculation and management using Stripe Tax
// This enables automatic tax calculation for VAT (EU), sales tax (US), GST, etc.
type TaxService struct {
	enabled bool
}

// NewTaxService creates a new tax service instance
func NewTaxService() *TaxService {
	return &TaxService{
		enabled: stripeKey() != "",
	}
}

// TaxCalculationParams contains parameters for tax calculation
type TaxCalculationParams struct {
	AmountCents         int64
	Currency            string
	CustomerID          string
	CustomerCountry     string
	CustomerState       string
	CustomerPostalCode  string
	TaxID               string
	TaxIDType           string
	LineItemDescription string
}

// CalculateTax calculates tax using Stripe Tax API
// Returns TaxCalculationResult with tax amount and breakdown
func (s *TaxService) CalculateTax(ctx context.Context, params TaxCalculationParams) (*TaxCalculationResult, error) {
	if !s.enabled {
		return nil, fmt.Errorf("Stripe Tax is not configured")
	}

	if params.AmountCents <= 0 {
		return nil, fmt.Errorf("amount must be positive")
	}
	if params.Currency == "" {
		params.Currency = "USD"
	}

	customerDetails := &stripe.TaxCalculationCustomerDetailsParams{
		Address: &stripe.AddressParams{
			Country: stripe.String(params.CustomerCountry),
		},
		AddressSource: stripe.String(string(stripe.TaxCalculationCustomerDetailsAddressSourceBilling)),
	}

	if params.CustomerPostalCode != "" {
		customerDetails.Address.PostalCode = stripe.String(params.CustomerPostalCode)
	}
	if params.CustomerState != "" {
		customerDetails.Address.State = stripe.String(params.CustomerState)
	}

	if params.TaxID != "" {
		taxID := &stripe.TaxCalculationCustomerDetailsTaxIDParams{
			Type: stripe.String(params.TaxIDType),
		}
		taxID.Value = stripe.String(params.TaxID)
		customerDetails.TaxIDs = []*stripe.TaxCalculationCustomerDetailsTaxIDParams{taxID}
	}

	calcParams := &stripe.TaxCalculationParams{
		Currency: stripe.String(strings.ToLower(params.Currency)),
		LineItems: []*stripe.TaxCalculationLineItemParams{
			{
				Amount:      stripe.Int64(params.AmountCents),
				Reference:   stripe.String(params.LineItemDescription),
				TaxBehavior: stripe.String(string(stripe.TaxTransactionLineItemTaxBehaviorExclusive)),
			},
		},
		CustomerDetails: customerDetails,
	}

	calc, err := calculation.New(calcParams)
	if err != nil {
		return nil, fmt.Errorf("Stripe Tax calculation failed: %w", err)
	}

	taxAmount := calc.TaxAmountExclusive
	taxRate := 0.0
	taxName := "Tax"
	jurisdiction := ""

	if len(calc.TaxBreakdown) > 0 && calc.TaxBreakdown[0].TaxRateDetails != nil {
		if calc.TaxBreakdown[0].TaxRateDetails.PercentageDecimal != "" {
			if rate, err := strconv.ParseFloat(calc.TaxBreakdown[0].TaxRateDetails.PercentageDecimal, 64); err == nil {
				taxRate = rate
			}
		}
		if calc.TaxBreakdown[0].TaxRateDetails.TaxType != "" {
			taxName = string(calc.TaxBreakdown[0].TaxRateDetails.TaxType)
		}
		if calc.TaxBreakdown[0].TaxRateDetails.Country != "" {
			jurisdiction = calc.TaxBreakdown[0].TaxRateDetails.Country
		}
		if calc.TaxBreakdown[0].TaxRateDetails.State != "" {
			if jurisdiction != "" {
				jurisdiction += "/" + calc.TaxBreakdown[0].TaxRateDetails.State
			} else {
				jurisdiction = calc.TaxBreakdown[0].TaxRateDetails.State
			}
		}
	}

	result := &TaxCalculationResult{
		TaxAmountCents:      taxAmount,
		SubtotalCents:       params.AmountCents,
		TotalCents:          params.AmountCents + taxAmount,
		TaxRatePercentage:   taxRate,
		TaxName:             taxName,
		Jurisdiction:        jurisdiction,
		StripeCalculationID: calc.ID,
		TaxBreakdown:        make([]TaxBreakdownItem, 0, len(calc.TaxBreakdown)),
	}

	for _, tb := range calc.TaxBreakdown {
		item := TaxBreakdownItem{
			TaxAmount: tb.Amount,
		}
		if tb.TaxRateDetails != nil {
			item.TaxType = string(tb.TaxRateDetails.TaxType)
			if tb.TaxRateDetails.PercentageDecimal != "" {
				if rate, err := strconv.ParseFloat(tb.TaxRateDetails.PercentageDecimal, 64); err == nil {
					item.TaxRate = rate
				}
			}
			if tb.TaxRateDetails.Country != "" {
				item.Jurisdiction = tb.TaxRateDetails.Country
			}
			if tb.TaxRateDetails.State != "" {
				if item.Jurisdiction != "" {
					item.Jurisdiction += "/" + tb.TaxRateDetails.State
				} else {
					item.Jurisdiction = tb.TaxRateDetails.State
				}
			}
		}
		result.TaxBreakdown = append(result.TaxBreakdown, item)
	}

	return result, nil
}

// UpdateTenantTaxSettings persists tax settings to the tenant record in storage.
func (s *TaxService) UpdateTenantTaxSettings(ctx context.Context, repo storage.Repository, tenantID uuid.UUID, settings *storage.TaxSettings) error {
	return repo.UpdateTenantTaxSettings(ctx, tenantID, settings)
}

// IsEnabled returns whether Stripe Tax is configured and enabled
func (s *TaxService) IsEnabled() bool {
	return s.enabled
}

// ValidateTaxID validates a tax ID format for supported types
// This performs client-side validation before sending to Stripe or external services
func (s *TaxService) ValidateTaxID(taxID string, taxIDType string) (bool, string) {
	taxID = strings.TrimSpace(strings.ToUpper(taxID))
	taxIDType = strings.ToLower(taxIDType)

	switch taxIDType {
	case "eu_vat":
		return validateEUVAT(taxID)
	case "uk_vat":
		return validateUKVAT(taxID)
	case "us_ein":
		return validateUSEIN(taxID)
	case "ca_gst":
		return validateCAGST(taxID)
	case "au_abn":
		return validateAUABN(taxID)
	case "ch_vat":
		return validateCHVAT(taxID)
	default:
		// For unknown types, just check for non-empty
		if taxID == "" {
			return false, "Tax ID is required"
		}
		return true, ""
	}
}

// validateEUVAT validates EU VAT numbers
// Format: 2-letter country code followed by 2-12 alphanumeric characters
// Examples: DE123456789, FR12345678901, GB123456789
func validateEUVAT(vat string) (bool, string) {
	// Remove any spaces or special characters
	vat = strings.ReplaceAll(vat, " ", "")
	vat = strings.ReplaceAll(vat, "-", "")

	if len(vat) < 3 {
		return false, "EU VAT number must be at least 3 characters"
	}

	// Extract country code
	countryCode := vat[:2]
	number := vat[2:]

	// Valid EU country codes
	validCountries := map[string]bool{
		"AT": true, "BE": true, "BG": true, "HR": true, "CY": true,
		"CZ": true, "DK": true, "EE": true, "FI": true, "FR": true,
		"DE": true, "GR": true, "HU": true, "IE": true, "IT": true,
		"LV": true, "LT": true, "LU": true, "MT": true, "NL": true,
		"PL": true, "PT": true, "RO": true, "SK": true, "SI": true,
		"ES": true, "SE": true, "GB": true, "XI": true,
	}

	if !validCountries[countryCode] {
		return false, fmt.Sprintf("Invalid EU country code: %s", countryCode)
	}

	// Basic format validation - alphanumeric characters only
	regex := regexp.MustCompile(`^[A-Z0-9]+$`)
	if !regex.MatchString(number) {
		return false, "VAT number contains invalid characters"
	}

	// Length validation by country
	minLengths := map[string]int{
		"AT": 8,  // Austria
		"BE": 8,  // Belgium
		"BG": 9,  // Bulgaria
		"HR": 11, // Croatia
		"CY": 8,  // Cyprus
		"CZ": 8,  // Czech Republic
		"DK": 8,  // Denmark
		"EE": 9,  // Estonia
		"FI": 8,  // Finland
		"FR": 11, // France
		"DE": 9,  // Germany
		"GR": 9,  // Greece
		"HU": 8,  // Hungary
		"IE": 8,  // Ireland
		"IT": 11, // Italy
		"LV": 11, // Latvia
		"LT": 9,  // Lithuania
		"LU": 8,  // Luxembourg
		"MT": 8,  // Malta
		"NL": 12, // Netherlands
		"PL": 10, // Poland
		"PT": 9,  // Portugal
		"RO": 2,  // Romania (variable)
		"SK": 10, // Slovakia
		"SI": 8,  // Slovenia
		"ES": 9,  // Spain
		"SE": 12, // Sweden
		"GB": 9,  // UK (Northern Ireland)
		"XI": 9,  // Northern Ireland
	}

	if minLen, ok := minLengths[countryCode]; ok {
		if len(number) < minLen {
			return false, fmt.Sprintf("VAT number for %s must be at least %d digits", countryCode, minLen)
		}
	}

	return true, ""
}

// validateUKVAT validates UK VAT numbers
// Format: GB followed by 9 digits
func validateUKVAT(vat string) (bool, string) {
	vat = strings.ReplaceAll(vat, " ", "")
	vat = strings.ToUpper(vat)

	if !strings.HasPrefix(vat, "GB") {
		return false, "UK VAT number must start with GB"
	}

	number := vat[2:]
	if len(number) != 9 {
		return false, "UK VAT number must be 9 digits after GB"
	}

	// Check all digits
	regex := regexp.MustCompile(`^[0-9]{9}$`)
	if !regex.MatchString(number) {
		return false, "UK VAT number must contain 9 digits after GB"
	}

	return true, ""
}

// validateUSEIN validates US EIN (Employer Identification Number)
// Format: 9 digits, often written as XX-XXXXXXX
func validateUSEIN(ein string) (bool, string) {
	// Remove dashes and spaces
	ein = strings.ReplaceAll(ein, "-", "")
	ein = strings.ReplaceAll(ein, " ", "")

	if len(ein) != 9 {
		return false, "US EIN must be 9 digits"
	}

	regex := regexp.MustCompile(`^[0-9]{9}$`)
	if !regex.MatchString(ein) {
		return false, "US EIN must contain only digits"
	}

	return true, ""
}

// validateCAGST validates Canada GST/HST numbers
// Format: 9 digits followed by RT0001 or RT
func validateCAGST(gst string) (bool, string) {
	gst = strings.ReplaceAll(gst, " ", "")
	gst = strings.ToUpper(gst)

	// Basic format: 9 digits + RT + 4 digits (optional)
	regex := regexp.MustCompile(`^[0-9]{9}RT[0-9]{0,4}$`)
	if !regex.MatchString(gst) {
		return false, "Canada GST number must be 9 digits followed by RT"
	}

	return true, ""
}

// validateAUABN validates Australia ABN (Australian Business Number)
// Format: 11 digits
func validateAUABN(abn string) (bool, string) {
	abn = strings.ReplaceAll(abn, " ", "")

	if len(abn) != 11 {
		return false, "Australian ABN must be 11 digits"
	}

	regex := regexp.MustCompile(`^[0-9]{11}$`)
	if !regex.MatchString(abn) {
		return false, "Australian ABN must contain only digits"
	}

	return true, ""
}

// validateCHVAT validates Switzerland VAT numbers (UID/MWST)
// Format: CHE-XXX.XXX.XXX or CHE followed by 9 digits
func validateCHVAT(vat string) (bool, string) {
	vat = strings.ToUpper(strings.ReplaceAll(vat, " ", ""))
	vat = strings.ReplaceAll(vat, "-", "")
	vat = strings.ReplaceAll(vat, ".", "")

	if !strings.HasPrefix(vat, "CHE") {
		return false, "Swiss VAT number must start with CHE"
	}

	number := vat[3:]
	if len(number) != 9 {
		return false, "Swiss VAT number must have 9 digits after CHE"
	}

	regex := regexp.MustCompile(`^[0-9]{9}$`)
	if !regex.MatchString(number) {
		return false, "Swiss VAT number must contain 9 digits after CHE"
	}

	return true, ""
}

// GetApplicableTaxTypes returns the applicable tax types for a given country
// This helps customers understand what tax ID they need to provide
func GetApplicableTaxTypes(countryCode string) []map[string]string {
	countryCode = strings.ToUpper(countryCode)

	taxTypes := map[string][]map[string]string{
		"AT": {{"type": "eu_vat", "name": "EU VAT", "description": "Austrian VAT number (ATU + 8 digits)"}},
		"BE": {{"type": "eu_vat", "name": "EU VAT", "description": "Belgian VAT number (BE + 10 digits)"}},
		"BG": {{"type": "eu_vat", "name": "EU VAT", "description": "Bulgarian VAT number (BG + 9-10 digits)"}},
		"HR": {{"type": "eu_vat", "name": "EU VAT", "description": "Croatian VAT number (HR + 11 digits)"}},
		"CY": {{"type": "eu_vat", "name": "EU VAT", "description": "Cypriot VAT number (CY + 8 digits)"}},
		"CZ": {{"type": "eu_vat", "name": "EU VAT", "description": "Czech VAT number (CZ + 8-10 digits)"}},
		"DK": {{"type": "eu_vat", "name": "EU VAT", "description": "Danish VAT number (DK + 8 digits)"}},
		"EE": {{"type": "eu_vat", "name": "EU VAT", "description": "Estonian VAT number (EE + 9 digits)"}},
		"FI": {{"type": "eu_vat", "name": "EU VAT", "description": "Finnish VAT number (FI + 8 digits)"}},
		"FR": {{"type": "eu_vat", "name": "EU VAT", "description": "French VAT number (FR + 11 digits)"}},
		"DE": {{"type": "eu_vat", "name": "EU VAT", "description": "German VAT number (DE + 9 digits)"}},
		"GR": {{"type": "eu_vat", "name": "EU VAT", "description": "Greek VAT number (EL + 9 digits)"}},
		"HU": {{"type": "eu_vat", "name": "EU VAT", "description": "Hungarian VAT number (HU + 8 digits)"}},
		"IE": {{"type": "eu_vat", "name": "EU VAT", "description": "Irish VAT number (IE + 8-9 digits)"}},
		"IT": {{"type": "eu_vat", "name": "EU VAT", "description": "Italian VAT number (IT + 11 digits)"}},
		"LV": {{"type": "eu_vat", "name": "EU VAT", "description": "Latvian VAT number (LV + 11 digits)"}},
		"LT": {{"type": "eu_vat", "name": "EU VAT", "description": "Lithuanian VAT number (LT + 9-12 digits)"}},
		"LU": {{"type": "eu_vat", "name": "EU VAT", "description": "Luxembourg VAT number (LU + 8 digits)"}},
		"MT": {{"type": "eu_vat", "name": "EU VAT", "description": "Maltese VAT number (MT + 8 digits)"}},
		"NL": {{"type": "eu_vat", "name": "EU VAT", "description": "Dutch VAT number (NL + 12 digits)"}},
		"PL": {{"type": "eu_vat", "name": "EU VAT", "description": "Polish VAT number (PL + 10 digits)"}},
		"PT": {{"type": "eu_vat", "name": "EU VAT", "description": "Portuguese VAT number (PT + 9 digits)"}},
		"RO": {{"type": "eu_vat", "name": "EU VAT", "description": "Romanian VAT number (RO + 2-10 digits)"}},
		"SK": {{"type": "eu_vat", "name": "EU VAT", "description": "Slovak VAT number (SK + 10 digits)"}},
		"SI": {{"type": "eu_vat", "name": "EU VAT", "description": "Slovenian VAT number (SI + 8 digits)"}},
		"ES": {{"type": "eu_vat", "name": "EU VAT", "description": "Spanish VAT number (ES + 9 digits)"}},
		"SE": {{"type": "eu_vat", "name": "EU VAT", "description": "Swedish VAT number (SE + 12 digits)"}},
		"GB": {{"type": "uk_vat", "name": "UK VAT", "description": "UK VAT number (GB + 9 digits)"}},
		"US": {{"type": "us_ein", "name": "US EIN", "description": "US Employer ID Number (XX-XXXXXXX)"}},
		"CA": {{"type": "ca_gst", "name": "CA GST/HST", "description": "Canada GST/HST number (9 digits + RT)"}},
		"AU": {{"type": "au_abn", "name": "AU ABN", "description": "Australian Business Number (11 digits)"}},
		"NZ": {{"type": "nz_gst", "name": "NZ GST", "description": "New Zealand GST number"}},
		"SG": {{"type": "sg_gst", "name": "SG GST", "description": "Singapore GST number"}},
		"CH": {{"type": "ch_vat", "name": "CH VAT", "description": "Swiss VAT number (CHE-XXX.XXX.XXX)"}},
		"NO": {{"type": "no_vat", "name": "NO VAT", "description": "Norwegian VAT number"}},
	}

	if types, ok := taxTypes[countryCode]; ok {
		return types
	}

	// Default: return generic info
	return []map[string]string{
		{"type": "other", "name": "Tax ID", "description": "Business tax registration number"},
	}
}

// TaxCalculationResult contains the result of a tax calculation
// Note: The actual tax calculation is performed by Stripe during checkout
// This struct is used for API responses and internal representation
type TaxCalculationResult struct {
	TaxAmountCents      int64
	SubtotalCents       int64
	TotalCents          int64
	TaxRatePercentage   float64
	TaxName             string
	Jurisdiction        string
	StripeCalculationID string
	TaxBreakdown        []TaxBreakdownItem
}

// TaxBreakdownItem represents a single tax component in a breakdown
type TaxBreakdownItem struct {
	TaxType      string
	TaxRate      float64
	TaxAmount    int64
	Jurisdiction string
}

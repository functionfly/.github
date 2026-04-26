// Package currency provides multi-currency support for the billing system
// This enables dynamic currency conversion and region-specific pricing
package currency

import (
	"context"
	"fmt"
	"time"

	"github.com/functionfly/functionfly/internal/storage"
)

// Service provides currency conversion and multi-currency support
type Service struct {
	repo         storage.Repository
	syncer       *ExchangeRateSyncer
	baseCurrency string
}

// NewService creates a new currency service
func NewService(repo storage.Repository, syncer *ExchangeRateSyncer, baseCurrency string) *Service {
	if baseCurrency == "" {
		baseCurrency = "USD"
	}
	return &Service{
		repo:         repo,
		syncer:       syncer,
		baseCurrency: baseCurrency,
	}
}

// GetLiveRate retrieves a live exchange rate (uses Redis cache + real-time API)
func (s *Service) GetLiveRate(ctx context.Context, from, to string) (float64, error) {
	if s.syncer == nil {
		rate, err := s.repo.GetCurrencyExchangeRate(ctx, from, to, nil)
		if err != nil || rate == nil {
			return 1.0, nil
		}
		return rate.Rate, nil
	}
	return s.syncer.GetLiveRate(ctx, from, to)
}

// Convert converts an amount from one currency to another
func (s *Service) Convert(ctx context.Context, amountCents int, fromCurrency, toCurrency string) (int, error) {
	if fromCurrency == toCurrency {
		return amountCents, nil
	}
	return s.repo.ConvertCurrency(ctx, amountCents, fromCurrency, toCurrency)
}

// ConvertToBase converts an amount to the base currency (USD)
func (s *Service) ConvertToBase(ctx context.Context, amountCents int, fromCurrency string) (int, error) {
	return s.Convert(ctx, amountCents, fromCurrency, s.baseCurrency)
}

// ConvertFromBase converts an amount from the base currency to a target currency
func (s *Service) ConvertFromBase(ctx context.Context, amountCents int, toCurrency string) (int, error) {
	return s.Convert(ctx, amountCents, s.baseCurrency, toCurrency)
}

// GetExchangeRate retrieves the current exchange rate for a currency pair
func (s *Service) GetExchangeRate(ctx context.Context, baseCurrency, quoteCurrency string) (*storage.CurrencyExchangeRate, error) {
	return s.repo.GetCurrencyExchangeRate(ctx, baseCurrency, quoteCurrency, nil)
}

// GetHistoricalExchangeRate retrieves an exchange rate for a specific date
func (s *Service) GetHistoricalExchangeRate(ctx context.Context, baseCurrency, quoteCurrency string, date time.Time) (*storage.CurrencyExchangeRate, error) {
	return s.repo.GetCurrencyExchangeRate(ctx, baseCurrency, quoteCurrency, &date)
}

// GetSupportedCurrency retrieves information about a supported currency
func (s *Service) GetSupportedCurrency(ctx context.Context, code string) (*storage.SupportedCurrency, error) {
	return s.repo.GetSupportedCurrency(ctx, code)
}

// ListSupportedCurrencies retrieves all active supported currencies
func (s *Service) ListSupportedCurrencies(ctx context.Context) ([]*storage.SupportedCurrency, error) {
	return s.repo.ListSupportedCurrencies(ctx)
}

// IsSupported checks if a currency is supported
func (s *Service) IsSupported(ctx context.Context, code string) bool {
	currency, err := s.GetSupportedCurrency(ctx, code)
	return err == nil && currency != nil && currency.IsActive
}

// IsStablecoin checks if a currency is a stablecoin (USDC, USDT, etc.)
func (s *Service) IsStablecoin(ctx context.Context, code string) bool {
	currency, err := s.GetSupportedCurrency(ctx, code)
	return err == nil && currency != nil && currency.IsStablecoin
}

// FormatAmount formats an amount for display in a specific currency
func (s *Service) FormatAmount(ctx context.Context, amountCents int, currencyCode string) (string, error) {
	currency, err := s.GetSupportedCurrency(ctx, currencyCode)
	if err != nil {
		return "", fmt.Errorf("currency not supported: %s", currencyCode)
	}
	if currency == nil {
		return "", fmt.Errorf("currency not found: %s", currencyCode)
	}
	return currency.FormatAmount(amountCents), nil
}

// ConvertAndFormat converts an amount and formats it for display
func (s *Service) ConvertAndFormat(ctx context.Context, amountCents int, fromCurrency, toCurrency string) (string, error) {
	converted, err := s.Convert(ctx, amountCents, fromCurrency, toCurrency)
	if err != nil {
		return "", err
	}
	return s.FormatAmount(ctx, converted, toCurrency)
}

// GetMinimumCharge returns the minimum charge amount for a currency (Stripe minimums)
func (s *Service) GetMinimumCharge(ctx context.Context, currencyCode string) (int, error) {
	currency, err := s.GetSupportedCurrency(ctx, currencyCode)
	if err != nil {
		return 50, err // Default 50 cents
	}
	if currency == nil {
		return 50, nil
	}
	return currency.MinimumChargeCents, nil
}

// ValidateAmount checks if an amount meets the minimum charge for a currency
func (s *Service) ValidateAmount(ctx context.Context, amountCents int, currencyCode string) error {
	minCharge, err := s.GetMinimumCharge(ctx, currencyCode)
	if err != nil {
		return err
	}
	if amountCents < minCharge {
		return fmt.Errorf("minimum charge for %s is %d cents", currencyCode, minCharge)
	}
	return nil
}

// ConvertToStripeAmount converts cents to Stripe's smallest unit for a currency
func (s *Service) ConvertToStripeAmount(ctx context.Context, amountCents int, currencyCode string) (int64, error) {
	currency, err := s.GetSupportedCurrency(ctx, currencyCode)
	if err != nil {
		return 0, err
	}
	if currency == nil {
		return int64(amountCents), nil
	}
	return currency.ConvertToStripeAmount(amountCents), nil
}

// ConvertFromStripeAmount converts Stripe's smallest unit to cents for a currency
func (s *Service) ConvertFromStripeAmount(ctx context.Context, stripeAmount int64, currencyCode string) (int, error) {
	currency, err := s.GetSupportedCurrency(ctx, currencyCode)
	if err != nil {
		return 0, err
	}
	if currency == nil {
		return int(stripeAmount), nil
	}
	return currency.ConvertFromStripeAmount(stripeAmount), nil
}

// Region represents a geographic region for regional pricing
type Region string

const (
	RegionUS     Region = "US"
	RegionEU     Region = "EU"
	RegionUK     Region = "UK"
	RegionCA     Region = "CA"
	RegionAU     Region = "AU"
	RegionAPAC   Region = "APAC"
	RegionGlobal Region = "GLOBAL"
)

// GetRegionForCurrency returns the default region for a currency
func GetRegionForCurrency(currencyCode string) Region {
	switch currencyCode {
	case "USD":
		return RegionUS
	case "EUR":
		return RegionEU
	case "GBP":
		return RegionUK
	case "CAD":
		return RegionCA
	case "AUD":
		return RegionAU
	case "JPY", "SGD", "HKD", "NZD", "KRW", "TWD", "THB", "MYR", "PHP", "IDR":
		return RegionAPAC
	default:
		return RegionGlobal
	}
}

// Common currency codes
const (
	USD = "USD"
	EUR = "EUR"
	GBP = "GBP"
	JPY = "JPY"
	CAD = "CAD"
	AUD = "AUD"
	CHF = "CHF"
	SEK = "SEK"
	NOK = "NOK"
	DKK = "DKK"
	PLN = "PLN"
	CZK = "CZK"
	MXN = "MXN"
	BRL = "BRL"
	SGD = "SGD"
	HKD = "HKD"
	NZD = "NZD"
	INR = "INR"
	KRW = "KRW"
	TWD = "TWD"
	THB = "THB"
	MYR = "MYR"
	PHP = "PHP"
	// Stablecoins
	USDC = "USDC"
	USDT = "USDT"
)

// IsZeroDecimalCurrency returns true for currencies that don't use decimal places (JPY, HUF, etc.)
func IsZeroDecimalCurrency(currencyCode string) bool {
	switch currencyCode {
	case "BIF", "CLP", "DJF", "GNF", "JPY", "KMF", "KRW", "MGA",
		"PYG", "RWF", "UGX", "VND", "VUV", "XAF", "XOF", "XPF":
		return true
	default:
		return false
	}
}

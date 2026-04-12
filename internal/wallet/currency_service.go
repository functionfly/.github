package wallet

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"os"
	"strconv"
	"sync"
	"time"

	"github.com/google/uuid"
	"github.com/redis/go-redis/v9"
	"github.com/sirupsen/logrus"
)

// Currency represents a supported currency with its metadata
type Currency struct {
	Code          string `json:"code"`
	Name          string `json:"name"`
	Symbol        string `json:"symbol"`
	DecimalPlaces int    `json:"decimal_places"`
	IsStablecoin  bool   `json:"is_stablecoin"`
	IsFiat        bool   `json:"is_fiat"`
}

// ExchangeRate represents a currency exchange rate
type ExchangeRate struct {
	FromCurrency string    `json:"from_currency"`
	ToCurrency   string    `json:"to_currency"`
	Rate         float64   `json:"rate"`
	InverseRate  float64   `json:"inverse_rate"`
	Source       string    `json:"source"`
	LastUpdated  time.Time `json:"last_updated"`
	NextUpdateAt time.Time `json:"next_update_at"`
	TTLSeconds   int       `json:"ttl_seconds"`
}

// CurrencyConversion represents a converted amount
type CurrencyConversion struct {
	OriginalAmount   float64   `json:"original_amount"`
	OriginalCurrency string    `json:"original_currency"`
	ConvertedAmount  float64   `json:"converted_amount"`
	TargetCurrency   string    `json:"target_currency"`
	ExchangeRate     float64   `json:"exchange_rate"`
	RateSource       string    `json:"rate_source"`
	RateUpdatedAt    time.Time `json:"rate_updated_at"`
}

// WalletCurrencyInfo holds currency information for a wallet
type WalletCurrencyInfo struct {
	WalletID            uuid.UUID `json:"wallet_id"`
	PrimaryCurrency     string    `json:"primary_currency"`
	AvailableCurrencies []string  `json:"available_currencies"`
	BalanceUSD          float64   `json:"balance_usd"`
	BalanceLocal        float64   `json:"balance_local"`
	ExchangeRate        float64   `json:"exchange_rate_to_usd"`
	RateUpdatedAt       time.Time `json:"rate_updated_at"`
}

// CurrencyProvider defines the interface for exchange rate providers
type CurrencyProvider interface {
	GetRate(ctx context.Context, from, to string) (*ExchangeRate, error)
	GetRates(ctx context.Context, base string, targets []string) (map[string]*ExchangeRate, error)
	Name() string
	IsAvailable() bool
}

// CurrencyService manages multi-currency operations
type CurrencyService struct {
	redis               *redis.Client
	logger              *logrus.Logger
	providers           []CurrencyProvider
	activeProvider      CurrencyProvider
	supportedCurrencies map[string]*Currency
	rateCacheTTL        time.Duration
	mutex               sync.RWMutex
}

// CurrencyServiceConfig holds configuration for the currency service
type CurrencyServiceConfig struct {
	RateCacheTTL        time.Duration
	DefaultProvider     string
	FallbackToHardcoded bool
}

// DefaultCurrencyServiceConfig returns default configuration
func DefaultCurrencyServiceConfig() *CurrencyServiceConfig {
	return &CurrencyServiceConfig{
		RateCacheTTL:        getEnvDuration("CURRENCY_RATE_CACHE_TTL", 1*time.Hour),
		DefaultProvider:     getEnvString("CURRENCY_RATE_PROVIDER", "openexchange"),
		FallbackToHardcoded: getEnvBool("CURRENCY_FALLBACK_HARDCODED", true),
	}
}

// NewCurrencyService creates a new currency service
func NewCurrencyService(redisClient *redis.Client) *CurrencyService {
	cfg := DefaultCurrencyServiceConfig()

	svc := &CurrencyService{
		redis:               redisClient,
		logger:              logrus.New(),
		supportedCurrencies: initSupportedCurrencies(),
		rateCacheTTL:        cfg.RateCacheTTL,
	}

	// Initialize providers
	svc.initProviders(cfg)

	return svc
}

// initProviders initializes currency rate providers
func (s *CurrencyService) initProviders(cfg *CurrencyServiceConfig) {
	// Add providers in priority order
	providers := []CurrencyProvider{
		NewOpenExchangeProvider(),
		NewExchangeRateAPIProvider(),
		NewHardcodedProvider(),
	}

	s.providers = providers

	// Set active provider based on availability
	for _, provider := range providers {
		if provider.IsAvailable() {
			s.activeProvider = provider
			s.logger.WithField("provider", provider.Name()).Info("Currency provider activated")
			break
		}
	}

	if s.activeProvider == nil {
		s.logger.Error("No currency provider available")
	}
}

// SetLogger sets the logger
func (s *CurrencyService) SetLogger(logger *logrus.Logger) {
	s.logger = logger
}

// GetSupportedCurrencies returns all supported currencies
func (s *CurrencyService) GetSupportedCurrencies() []*Currency {
	s.mutex.RLock()
	defer s.mutex.RUnlock()

	currencies := make([]*Currency, 0, len(s.supportedCurrencies))
	for _, c := range s.supportedCurrencies {
		currencies = append(currencies, c)
	}
	return currencies
}

// IsCurrencySupported checks if a currency is supported
func (s *CurrencyService) IsCurrencySupported(code string) bool {
	s.mutex.RLock()
	defer s.mutex.RUnlock()

	_, ok := s.supportedCurrencies[code]
	return ok
}

// GetCurrency returns a specific currency's details
func (s *CurrencyService) GetCurrency(code string) (*Currency, error) {
	s.mutex.RLock()
	defer s.mutex.RUnlock()

	currency, ok := s.supportedCurrencies[code]
	if !ok {
		return nil, fmt.Errorf("unsupported currency: %s", code)
	}

	return currency, nil
}

// GetExchangeRate retrieves the exchange rate between two currencies
func (s *CurrencyService) GetExchangeRate(ctx context.Context, from, to string) (*ExchangeRate, error) {
	// Validate currencies
	if !s.IsCurrencySupported(from) {
		return nil, fmt.Errorf("unsupported source currency: %s", from)
	}
	if !s.IsCurrencySupported(to) {
		return nil, fmt.Errorf("unsupported target currency: %s", to)
	}

	// Check cache first
	if rate := s.getCachedRate(ctx, from, to); rate != nil {
		return rate, nil
	}

	// Fetch from provider
	if s.activeProvider == nil {
		return nil, fmt.Errorf("no currency provider available")
	}

	rate, err := s.activeProvider.GetRate(ctx, from, to)
	if err != nil {
		s.logger.WithError(err).WithFields(logrus.Fields{
			"from": from,
			"to":   to,
		}).Error("Failed to get exchange rate from provider")
		return nil, err
	}

	// Cache the rate
	s.cacheRate(ctx, rate)

	return rate, nil
}

// Convert converts an amount from one currency to another
func (s *CurrencyService) Convert(ctx context.Context, amount float64, from, to string) (*CurrencyConversion, error) {
	rate, err := s.GetExchangeRate(ctx, from, to)
	if err != nil {
		return nil, err
	}

	convertedAmount := amount * rate.Rate

	return &CurrencyConversion{
		OriginalAmount:   amount,
		OriginalCurrency: from,
		ConvertedAmount:  convertedAmount,
		TargetCurrency:   to,
		ExchangeRate:     rate.Rate,
		RateSource:       rate.Source,
		RateUpdatedAt:    rate.LastUpdated,
	}, nil
}

// ConvertWalletBalance converts a wallet balance to a target currency
func (s *CurrencyService) ConvertWalletBalance(ctx context.Context, wallet *Wallet, targetCurrency string) (*CurrencyConversion, error) {
	if wallet.Currency == targetCurrency {
		return &CurrencyConversion{
			OriginalAmount:   wallet.BalanceUSD,
			OriginalCurrency: targetCurrency,
			ConvertedAmount:  wallet.BalanceUSD,
			TargetCurrency:   targetCurrency,
			ExchangeRate:     1.0,
			RateSource:       "identity",
			RateUpdatedAt:    time.Now(),
		}, nil
	}

	return s.Convert(ctx, wallet.BalanceUSD, wallet.Currency, targetCurrency)
}

// UpdateWalletCurrency updates a wallet's currency and exchange rate
func (s *CurrencyService) UpdateWalletCurrency(ctx context.Context, walletID uuid.UUID, newCurrency string) error {
	if !s.IsCurrencySupported(newCurrency) {
		return fmt.Errorf("unsupported currency: %s", newCurrency)
	}

	// Get current rate
	rate, err := s.GetExchangeRate(ctx, "USD", newCurrency)
	if err != nil {
		return err
	}

	// This would update the wallet in the database
	// Implementation depends on the repository
	s.logger.WithFields(logrus.Fields{
		"wallet_id":     walletID,
		"new_currency":  newCurrency,
		"exchange_rate": rate.Rate,
	}).Info("Wallet currency update requested")

	return nil
}

// GetWalletCurrencyInfo returns currency information for a wallet
func (s *CurrencyService) GetWalletCurrencyInfo(ctx context.Context, wallet *Wallet) (*WalletCurrencyInfo, error) {
	// Convert balance to local currency
	var balanceLocal float64
	var exchangeRate float64

	if wallet.ExchangeRateToUSD != nil {
		exchangeRate = *wallet.ExchangeRateToUSD
		balanceLocal = wallet.BalanceUSD * exchangeRate
	}

	return &WalletCurrencyInfo{
		WalletID:            wallet.ID,
		PrimaryCurrency:     wallet.Currency,
		AvailableCurrencies: s.getAvailableCurrenciesForWallet(wallet),
		BalanceUSD:          wallet.BalanceUSD,
		BalanceLocal:        balanceLocal,
		ExchangeRate:        exchangeRate,
		RateUpdatedAt:       time.Now(), // This should be stored in the wallet
	}, nil
}

// getAvailableCurrenciesForWallet returns available currencies for a wallet based on its configuration
func (s *CurrencyService) getAvailableCurrenciesForWallet(wallet *Wallet) []string {
	var codes []string

	for code, currency := range s.supportedCurrencies {
		// Filter by wallet type restrictions
		if !s.isCurrencyAllowedForWalletType(currency, wallet.WalletType) {
			continue
		}

		// Filter by owner type (users get full access, agents may be restricted)
		if wallet.OwnerType == "agent" && !s.isCurrencyAllowedForAgent(currency) {
			continue
		}

		codes = append(codes, code)
	}

	return codes
}

// isCurrencyAllowedForWalletType checks if a currency is supported for a given wallet type
func (s *CurrencyService) isCurrencyAllowedForWalletType(currency *Currency, walletType string) bool {
	switch walletType {
	case "registry":
		// Registry wallets: stablecoins + major fiat only
		return currency.IsStablecoin || currency.Code == "USD" || currency.Code == "EUR" || currency.Code == "GBP"
	case "execution":
		// Execution wallets: all currencies (used for compute payments)
		return true
	case "unified":
		// Unified wallets: all supported currencies
		return true
	default:
		// Default to stablecoins and major fiat for unknown wallet types
		return currency.IsStablecoin || currency.IsFiat
	}
}

// isCurrencyAllowedForAgent checks if a currency is supported for agent-owned wallets
func (s *CurrencyService) isCurrencyAllowedForAgent(currency *Currency) bool {
	// Agents typically restricted to stablecoins for predictable costs
	// and major fiat currencies for billing stability
	if currency.IsStablecoin {
		return true
	}
	// Allow major fiat only (volatile currencies excluded for agents)
	switch currency.Code {
	case "USD", "EUR", "GBP":
		return true
	default:
		return false
	}
}

// getCachedRate retrieves a rate from Redis cache
func (s *CurrencyService) getCachedRate(ctx context.Context, from, to string) *ExchangeRate {
	if s.redis == nil {
		return nil
	}

	key := fmt.Sprintf("currency:rate:%s:%s", from, to)
	data, err := s.redis.Get(ctx, key).Result()
	if err != nil {
		return nil
	}

	var rate ExchangeRate
	if err := json.Unmarshal([]byte(data), &rate); err != nil {
		return nil
	}

	// Check if expired
	if time.Now().After(rate.NextUpdateAt) {
		return nil
	}

	return &rate
}

// cacheRate stores a rate in Redis cache
func (s *CurrencyService) cacheRate(ctx context.Context, rate *ExchangeRate) {
	if s.redis == nil {
		return
	}

	key := fmt.Sprintf("currency:rate:%s:%s", rate.FromCurrency, rate.ToCurrency)
	data, err := json.Marshal(rate)
	if err != nil {
		return
	}

	ttl := s.rateCacheTTL
	if rate.TTLSeconds > 0 {
		ttl = time.Duration(rate.TTLSeconds) * time.Second
	}

	if err := s.redis.Set(ctx, key, data, ttl).Err(); err != nil {
		s.logger.WithError(err).Debug("Failed to cache exchange rate")
	}
}

// RefreshRates forces a refresh of all cached rates
func (s *CurrencyService) RefreshRates(ctx context.Context) error {
	if s.activeProvider == nil {
		return fmt.Errorf("no currency provider available")
	}

	// Get all supported currencies
	currencies := s.GetSupportedCurrencies()
	codes := make([]string, len(currencies))
	for i, c := range currencies {
		codes[i] = c.Code
	}

	// Fetch fresh rates
	rates, err := s.activeProvider.GetRates(ctx, "USD", codes)
	if err != nil {
		return err
	}

	// Cache all rates
	for _, rate := range rates {
		s.cacheRate(ctx, rate)
	}

	s.logger.WithField("rates_count", len(rates)).Info("Exchange rates refreshed")

	return nil
}

// initSupportedCurrencies initializes the supported currency list
func initSupportedCurrencies() map[string]*Currency {
	return map[string]*Currency{
		"USD": {
			Code:          "USD",
			Name:          "US Dollar",
			Symbol:        "$",
			DecimalPlaces: 2,
			IsStablecoin:  false,
			IsFiat:        true,
		},
		"EUR": {
			Code:          "EUR",
			Name:          "Euro",
			Symbol:        "€",
			DecimalPlaces: 2,
			IsStablecoin:  false,
			IsFiat:        true,
		},
		"GBP": {
			Code:          "GBP",
			Name:          "British Pound",
			Symbol:        "£",
			DecimalPlaces: 2,
			IsStablecoin:  false,
			IsFiat:        true,
		},
		"JPY": {
			Code:          "JPY",
			Name:          "Japanese Yen",
			Symbol:        "¥",
			DecimalPlaces: 0,
			IsStablecoin:  false,
			IsFiat:        true,
		},
		"CAD": {
			Code:          "CAD",
			Name:          "Canadian Dollar",
			Symbol:        "C$",
			DecimalPlaces: 2,
			IsStablecoin:  false,
			IsFiat:        true,
		},
		"AUD": {
			Code:          "AUD",
			Name:          "Australian Dollar",
			Symbol:        "A$",
			DecimalPlaces: 2,
			IsStablecoin:  false,
			IsFiat:        true,
		},
		"CHF": {
			Code:          "CHF",
			Name:          "Swiss Franc",
			Symbol:        "Fr",
			DecimalPlaces: 2,
			IsStablecoin:  false,
			IsFiat:        true,
		},
		"CNY": {
			Code:          "CNY",
			Name:          "Chinese Yuan",
			Symbol:        "¥",
			DecimalPlaces: 2,
			IsStablecoin:  false,
			IsFiat:        true,
		},
		"USDC": {
			Code:          "USDC",
			Name:          "USD Coin",
			Symbol:        "USDC",
			DecimalPlaces: 6,
			IsStablecoin:  true,
			IsFiat:        false,
		},
		"USDT": {
			Code:          "USDT",
			Name:          "Tether",
			Symbol:        "USDT",
			DecimalPlaces: 6,
			IsStablecoin:  true,
			IsFiat:        false,
		},
	}
}

// =============================================================================
// Exchange Rate Providers
// =============================================================================

// OpenExchangeProvider implements CurrencyProvider for OpenExchangeRates API
type OpenExchangeProvider struct {
	apiKey string
	client *http.Client
}

// NewOpenExchangeProvider creates a new OpenExchange provider
func NewOpenExchangeProvider() *OpenExchangeProvider {
	return &OpenExchangeProvider{
		apiKey: os.Getenv("OPENEXCHANGE_API_KEY"),
		client: &http.Client{Timeout: 10 * time.Second},
	}
}

func (p *OpenExchangeProvider) Name() string {
	return "openexchange"
}

func (p *OpenExchangeProvider) IsAvailable() bool {
	return p.apiKey != ""
}

func (p *OpenExchangeProvider) GetRate(ctx context.Context, from, to string) (*ExchangeRate, error) {
	// OpenExchangeRates free tier only supports USD as base
	if from != "USD" {
		// Get USD rate and invert
		return p.getInvertedRate(ctx, from, to)
	}

	url := fmt.Sprintf("https://open.er-api.com/v6/latest/%s", from)

	req, err := http.NewRequestWithContext(ctx, "GET", url, nil)
	if err != nil {
		return nil, err
	}

	resp, err := p.client.Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("API returned status %d", resp.StatusCode)
	}

	var result struct {
		Rates map[string]float64 `json:"rates"`
		Date  string             `json:"date"`
	}

	if err := json.NewDecoder(resp.Body).Decode(&result); err != nil {
		return nil, err
	}

	rate, ok := result.Rates[to]
	if !ok {
		return nil, fmt.Errorf("rate not found for %s", to)
	}

	return &ExchangeRate{
		FromCurrency: from,
		ToCurrency:   to,
		Rate:         rate,
		InverseRate:  1 / rate,
		Source:       p.Name(),
		LastUpdated:  time.Now(),
		NextUpdateAt: time.Now().Add(1 * time.Hour),
		TTLSeconds:   3600,
	}, nil
}

func (p *OpenExchangeProvider) getInvertedRate(ctx context.Context, from, to string) (*ExchangeRate, error) {
	// Get USD to 'from' rate, then invert
	usdToFrom, err := p.GetRate(ctx, "USD", from)
	if err != nil {
		return nil, err
	}

	if to == "USD" {
		return &ExchangeRate{
			FromCurrency: from,
			ToCurrency:   to,
			Rate:         usdToFrom.InverseRate,
			InverseRate:  usdToFrom.Rate,
			Source:       p.Name(),
			LastUpdated:  time.Now(),
			NextUpdateAt: time.Now().Add(1 * time.Hour),
			TTLSeconds:   3600,
		}, nil
	}

	// Need to cross via USD
	usdToTo, err := p.GetRate(ctx, "USD", to)
	if err != nil {
		return nil, err
	}

	crossRate := usdToTo.Rate / usdToFrom.Rate

	return &ExchangeRate{
		FromCurrency: from,
		ToCurrency:   to,
		Rate:         crossRate,
		InverseRate:  1 / crossRate,
		Source:       p.Name(),
		LastUpdated:  time.Now(),
		NextUpdateAt: time.Now().Add(1 * time.Hour),
		TTLSeconds:   3600,
	}, nil
}

func (p *OpenExchangeProvider) GetRates(ctx context.Context, base string, targets []string) (map[string]*ExchangeRate, error) {
	rates := make(map[string]*ExchangeRate)

	for _, target := range targets {
		rate, err := p.GetRate(ctx, base, target)
		if err != nil {
			continue
		}
		rates[target] = rate
	}

	return rates, nil
}

// ExchangeRateAPIProvider implements CurrencyProvider for exchangerate-api.com
type ExchangeRateAPIProvider struct {
	client *http.Client
}

// NewExchangeRateAPIProvider creates a new ExchangeRateAPI provider
func NewExchangeRateAPIProvider() *ExchangeRateAPIProvider {
	return &ExchangeRateAPIProvider{
		client: &http.Client{Timeout: 10 * time.Second},
	}
}

func (p *ExchangeRateAPIProvider) Name() string {
	return "exchangerate-api"
}

func (p *ExchangeRateAPIProvider) IsAvailable() bool {
	// This API doesn't require an API key for basic usage
	return true
}

func (p *ExchangeRateAPIProvider) GetRate(ctx context.Context, from, to string) (*ExchangeRate, error) {
	url := fmt.Sprintf("https://api.exchangerate-api.com/v4/latest/%s", from)

	req, err := http.NewRequestWithContext(ctx, "GET", url, nil)
	if err != nil {
		return nil, err
	}

	resp, err := p.client.Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("API returned status %d", resp.StatusCode)
	}

	var result struct {
		Rates map[string]float64 `json:"rates"`
		Date  string             `json:"date"`
	}

	if err := json.NewDecoder(resp.Body).Decode(&result); err != nil {
		return nil, err
	}

	rate, ok := result.Rates[to]
	if !ok {
		return nil, fmt.Errorf("rate not found for %s", to)
	}

	return &ExchangeRate{
		FromCurrency: from,
		ToCurrency:   to,
		Rate:         rate,
		InverseRate:  1 / rate,
		Source:       p.Name(),
		LastUpdated:  time.Now(),
		NextUpdateAt: time.Now().Add(1 * time.Hour),
		TTLSeconds:   3600,
	}, nil
}

func (p *ExchangeRateAPIProvider) GetRates(ctx context.Context, base string, targets []string) (map[string]*ExchangeRate, error) {
	url := fmt.Sprintf("https://api.exchangerate-api.com/v4/latest/%s", base)

	req, err := http.NewRequestWithContext(ctx, "GET", url, nil)
	if err != nil {
		return nil, err
	}

	resp, err := p.client.Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("API returned status %d", resp.StatusCode)
	}

	var result struct {
		Rates map[string]float64 `json:"rates"`
		Date  string             `json:"date"`
	}

	if err := json.NewDecoder(resp.Body).Decode(&result); err != nil {
		return nil, err
	}

	rates := make(map[string]*ExchangeRate)
	for _, target := range targets {
		rate, ok := result.Rates[target]
		if !ok {
			continue
		}

		rates[target] = &ExchangeRate{
			FromCurrency: base,
			ToCurrency:   target,
			Rate:         rate,
			InverseRate:  1 / rate,
			Source:       p.Name(),
			LastUpdated:  time.Now(),
			NextUpdateAt: time.Now().Add(1 * time.Hour),
			TTLSeconds:   3600,
		}
	}

	return rates, nil
}

// HardcodedProvider provides fallback hardcoded rates
type HardcodedProvider struct{}

// NewHardcodedProvider creates a new hardcoded provider
func NewHardcodedProvider() *HardcodedProvider {
	return &HardcodedProvider{}
}

func (p *HardcodedProvider) Name() string {
	return "hardcoded"
}

func (p *HardcodedProvider) IsAvailable() bool {
	return true
}

func (p *HardcodedProvider) GetRate(ctx context.Context, from, to string) (*ExchangeRate, error) {
	hardcodedRates := map[string]map[string]float64{
		"USD": {
			"EUR":  0.85,
			"GBP":  0.73,
			"JPY":  110.0,
			"CAD":  1.25,
			"AUD":  1.35,
			"CHF":  0.92,
			"CNY":  6.45,
			"USDC": 1.0,
			"USDT": 1.0,
		},
	}

	// USD to anything
	if from == "USD" {
		rate, ok := hardcodedRates["USD"][to]
		if !ok {
			return nil, fmt.Errorf("rate not found for %s to %s", from, to)
		}
		return &ExchangeRate{
			FromCurrency: from,
			ToCurrency:   to,
			Rate:         rate,
			InverseRate:  1 / rate,
			Source:       p.Name(),
			LastUpdated:  time.Now(),
			NextUpdateAt: time.Now().Add(24 * time.Hour),
			TTLSeconds:   86400,
		}, nil
	}

	// Anything to USD
	if to == "USD" {
		rate, ok := hardcodedRates["USD"][from]
		if !ok {
			return nil, fmt.Errorf("rate not found for %s to %s", from, to)
		}
		return &ExchangeRate{
			FromCurrency: from,
			ToCurrency:   to,
			Rate:         1 / rate,
			InverseRate:  rate,
			Source:       p.Name(),
			LastUpdated:  time.Now(),
			NextUpdateAt: time.Now().Add(24 * time.Hour),
			TTLSeconds:   86400,
		}, nil
	}

	// Cross rate via USD
	usdToFrom, ok := hardcodedRates["USD"][from]
	if !ok {
		return nil, fmt.Errorf("rate not found for %s to %s", from, to)
	}
	usdToTo, ok := hardcodedRates["USD"][to]
	if !ok {
		return nil, fmt.Errorf("rate not found for %s to %s", from, to)
	}

	crossRate := usdToTo / usdToFrom

	return &ExchangeRate{
		FromCurrency: from,
		ToCurrency:   to,
		Rate:         crossRate,
		InverseRate:  1 / crossRate,
		Source:       p.Name(),
		LastUpdated:  time.Now(),
		NextUpdateAt: time.Now().Add(24 * time.Hour),
		TTLSeconds:   86400,
	}, nil
}

func (p *HardcodedProvider) GetRates(ctx context.Context, base string, targets []string) (map[string]*ExchangeRate, error) {
	rates := make(map[string]*ExchangeRate)

	for _, target := range targets {
		rate, err := p.GetRate(ctx, base, target)
		if err != nil {
			continue
		}
		rates[target] = rate
	}

	return rates, nil
}

// Environment helpers
func getEnvDuration(key string, defaultVal time.Duration) time.Duration {
	if val := os.Getenv(key); val != "" {
		if d, err := time.ParseDuration(val); err == nil {
			return d
		}
	}
	return defaultVal
}

func getEnvString(key, defaultVal string) string {
	if val := os.Getenv(key); val != "" {
		return val
	}
	return defaultVal
}

func getEnvBool(key string, defaultVal bool) bool {
	if val := os.Getenv(key); val != "" {
		if b, err := strconv.ParseBool(val); err == nil {
			return b
		}
	}
	return defaultVal
}

func getEnvFloat64(key string, defaultVal float64) float64 {
	if val := os.Getenv(key); val != "" {
		if f, err := strconv.ParseFloat(val, 64); err == nil {
			return f
		}
	}
	return defaultVal
}

func getEnvInt(key string, defaultVal int) int {
	if val := os.Getenv(key); val != "" {
		if i, err := strconv.Atoi(val); err == nil {
			return i
		}
	}
	return defaultVal
}

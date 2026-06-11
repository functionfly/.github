package currency

import (
	"context"
	"encoding/json"
	"fmt"
	"math"
	"net/http"
	"os"
	"sync"
	"sync/atomic"
	"time"

	"github.com/functionfly/functionfly/internal/storage"
	"github.com/redis/go-redis/v9"
	"github.com/robfig/cron/v3"
	"github.com/sirupsen/logrus"
)

type RateProvider interface {
	Name() string
	IsAvailable() bool
	GetRates(ctx context.Context, base string) (map[string]float64, error)
}

// CircuitBreakerState represents the state of a circuit breaker
type CircuitBreakerState int32

const (
	CircuitClosed CircuitBreakerState = iota
	CircuitOpen
	CircuitHalfOpen
)

func (s CircuitBreakerState) String() string {
	switch s {
	case CircuitClosed:
		return "closed"
	case CircuitOpen:
		return "open"
	case CircuitHalfOpen:
		return "half-open"
	default:
		return "unknown"
	}
}

// CircuitBreaker implements the circuit breaker pattern for external service calls
type CircuitBreaker struct {
	name             string
	failureThreshold int32
	successThreshold int32
	openTimeout      time.Duration
	state            atomic.Int32
	failureCount      atomic.Int32
	successCount      atomic.Int32
	lastFailureTime   atomic.Int64
	mu               sync.RWMutex
	logger           *logrus.Logger
}

// NewCircuitBreaker creates a new circuit breaker
func NewCircuitBreaker(name string, failureThreshold int, successThreshold int, openTimeout time.Duration) *CircuitBreaker {
	return &CircuitBreaker{
		name:             name,
		failureThreshold: int32(failureThreshold),
		successThreshold: int32(successThreshold),
		openTimeout:      openTimeout,
		logger:           logrus.New(),
	}
}

// Allow checks if a request should be allowed through
func (cb *CircuitBreaker) Allow() bool {
	state := CircuitBreakerState(cb.state.Load())
	switch state {
	case CircuitClosed:
		return true
	case CircuitOpen:
		// Check if timeout has passed to try half-open
		lastFailure := time.Unix(0, cb.lastFailureTime.Load())
		if time.Since(lastFailure) > cb.openTimeout {
			cb.state.Store(int32(CircuitHalfOpen))
			cb.successCount.Store(0)
			return true
		}
		return false
	case CircuitHalfOpen:
		return true
	default:
		return false
	}
}

// RecordSuccess records a successful call
func (cb *CircuitBreaker) RecordSuccess() {
	state := CircuitBreakerState(cb.state.Load())
	if state == CircuitHalfOpen {
		if cb.successCount.Add(1) >= cb.successThreshold {
			cb.state.Store(int32(CircuitClosed))
			cb.failureCount.Store(0)
			cb.successCount.Store(0)
			cb.logger.WithField("circuit", cb.name).Info("Circuit breaker closed")
		}
	} else if state == CircuitClosed {
		cb.failureCount.Store(0)
	}
}

// RecordFailure records a failed call
func (cb *CircuitBreaker) RecordFailure() {
	state := CircuitBreakerState(cb.state.Load())
	if state == CircuitHalfOpen {
		cb.state.Store(int32(CircuitOpen))
		cb.lastFailureTime.Store(time.Now().UnixNano())
		cb.logger.WithField("circuit", cb.name).Warn("Circuit breaker opened from half-open")
	} else if state == CircuitClosed {
		if cb.failureCount.Add(1) >= cb.failureThreshold {
			cb.state.Store(int32(CircuitOpen))
			cb.lastFailureTime.Store(time.Now().UnixNano())
			cb.logger.WithField("circuit", cb.name).Warn("Circuit breaker opened")
		}
	}
}

// State returns the current state of the circuit breaker
func (cb *CircuitBreaker) State() CircuitBreakerState {
	return CircuitBreakerState(cb.state.Load())
}

type ExchangeRateSyncer struct {
	repo          storage.Repository
	redis         *redis.Client
	logger        *logrus.Logger
	providers     []RateProvider
	effectiveDate string
	cacheTTL      time.Duration
	mu            sync.RWMutex
	circuitBreakers map[string]*CircuitBreaker
}

func NewExchangeRateSyncer(repo storage.Repository, redisClient *redis.Client) *ExchangeRateSyncer {
	syncer := &ExchangeRateSyncer{
		repo:      repo,
		redis:     redisClient,
		logger:    logrus.New(),
		providers: []RateProvider{NewFrankfurterProvider(), NewStripeRateProvider()},
		cacheTTL:  15 * time.Minute,
		circuitBreakers: map[string]*CircuitBreaker{
			"frankfurter": NewCircuitBreaker("frankfurter", 3, 2, 30*time.Second),
			"stripe":      NewCircuitBreaker("stripe", 3, 2, 30*time.Second),
		},
	}
	return syncer
}

func (s *ExchangeRateSyncer) SetLogger(logger *logrus.Logger) {
	s.logger = logger
}

func (s *ExchangeRateSyncer) SetEffectiveDate(date string) {
	s.effectiveDate = date
}

func (s *ExchangeRateSyncer) GetLiveRate(ctx context.Context, from, to string) (float64, error) {
	if from == to {
		return 1.0, nil
	}

	if rate := s.getCachedRate(ctx, from, to); rate > 0 {
		return rate, nil
	}

	rate, err := s.fetchAndCacheRate(ctx, from, to)
	if err != nil {
		return s.getFallbackRate(from, to), nil
	}
	return rate, nil
}

func (s *ExchangeRateSyncer) getCachedRate(ctx context.Context, from, to string) float64 {
	if s.redis == nil {
		return 0
	}

	key := fmt.Sprintf("rate:live:%s:%s", from, to)
	data, err := s.redis.Get(ctx, key).Result()
	if err != nil {
		return 0
	}

	var rate float64
	if err := json.Unmarshal([]byte(data), &rate); err != nil {
		return 0
	}
	return rate
}

func (s *ExchangeRateSyncer) setCachedRate(ctx context.Context, from, to string, rate float64) {
	if s.redis == nil {
		return
	}

	key := fmt.Sprintf("rate:live:%s:%s", from, to)
	data, _ := json.Marshal(rate)
	s.redis.Set(ctx, key, data, s.cacheTTL)
}

func (s *ExchangeRateSyncer) fetchAndCacheRate(ctx context.Context, from, to string) (float64, error) {
	for _, provider := range s.providers {
		if !provider.IsAvailable() {
			continue
		}

		// Check circuit breaker
		if cb, ok := s.circuitBreakers[provider.Name()]; ok {
			if !cb.Allow() {
				s.logger.WithField("provider", provider.Name()).Debug("Circuit breaker open, skipping provider")
				continue
			}
		}

		rates, err := provider.GetRates(ctx, from)
		if err != nil {
			s.logger.WithError(err).WithField("provider", provider.Name()).Debug("Provider failed")
			if cb, ok := s.circuitBreakers[provider.Name()]; ok {
				cb.RecordFailure()
			}
			continue
		}

		if rate, ok := rates[to]; ok && rate > 0 {
			s.setCachedRate(ctx, from, to, rate)
			if cb, ok := s.circuitBreakers[provider.Name()]; ok {
				cb.RecordSuccess()
			}
			return rate, nil
		}
	}

	return 0, fmt.Errorf("no rate found for %s to %s", from, to)
}

func (s *ExchangeRateSyncer) getFallbackRate(from, to string) float64 {
	fallbacks := map[string]map[string]float64{
		"USD": {"EUR": 0.92, "GBP": 0.79, "JPY": 148.5, "CAD": 1.35, "AUD": 1.52, "CHF": 0.88, "CNY": 6.45, "INR": 83.1, "KRW": 1330, "MXN": 17.1, "BRL": 4.95, "SGD": 1.34},
		"EUR": {"USD": 1.09, "GBP": 0.86, "JPY": 161.5, "CAD": 1.47, "AUD": 1.65, "CHF": 0.96, "CNY": 7.01, "INR": 90.3, "KRW": 1446, "MXN": 18.6, "BRL": 5.38, "SGD": 1.46},
	}
	if fromRates, ok := fallbacks[from]; ok {
		if rate, ok := fromRates[to]; ok {
			return rate
		}
	}
	return 1.0
}

func (s *ExchangeRateSyncer) SyncAllRates(ctx context.Context) error {
	s.logger.Info("Starting exchange rate sync")

	for _, provider := range s.providers {
		if !provider.IsAvailable() {
			s.logger.WithField("provider", provider.Name()).Debug("Provider not available")
			continue
		}

		s.logger.WithField("provider", provider.Name()).Info("Fetching rates from provider")

		rates, err := provider.GetRates(ctx, "USD")
		if err != nil {
			s.logger.WithError(err).WithField("provider", provider.Name()).Warn("Failed to fetch rates, trying next provider")
			continue
		}

		if err := s.saveRates(ctx, rates, provider.Name()); err != nil {
			s.logger.WithError(err).WithField("provider", provider.Name()).Error("Failed to save rates, trying next provider")
			continue
		}

		s.logger.WithFields(logrus.Fields{
			"provider":    provider.Name(),
			"rates_count": len(rates),
		}).Info("Exchange rates synced successfully")

		return nil
	}

	s.logger.Warn("All providers failed, using hardcoded fallback rates")
	return s.syncFallbackRates(ctx)
}

func (s *ExchangeRateSyncer) syncFallbackRates(ctx context.Context) error {
	fallbackRates := map[string]float64{
		"EUR": 0.92, "GBP": 0.79, "JPY": 148.5, "CAD": 1.35, "AUD": 1.52,
		"CHF": 0.88, "CNY": 6.45, "SEK": 10.4, "NOK": 10.6, "DKK": 6.89,
		"PLN": 4.02, "CZK": 23.4, "MXN": 17.1, "BRL": 4.95, "SGD": 1.34,
		"HKD": 7.82, "NZD": 1.64, "INR": 83.1, "KRW": 1330.0, "THB": 35.8,
		"MYR": 4.72, "PHP": 56.1, "IDR": 15600.0,
	}
	return s.saveRates(ctx, fallbackRates, "fallback")
}

func (s *ExchangeRateSyncer) saveRates(ctx context.Context, rates map[string]float64, source string) error {
	effectiveDate := s.effectiveDate
	if effectiveDate == "" {
		effectiveDate = time.Now().Format("2006-01-02")
	}

	now := time.Now()
	for quoteCurrency, rate := range rates {
		if quoteCurrency == "USD" || rate <= 0 || math.IsNaN(rate) || math.IsInf(rate, 0) {
			continue
		}

		roundedRate := math.Round(rate*1e8) / 1e8
		currencyRate := &storage.CurrencyExchangeRate{
			BaseCurrency:    "USD",
			QuoteCurrency:   quoteCurrency,
			Rate:            roundedRate,
			RateNumerator:   int64(roundedRate * 1_000_000),
			RateDenominator: 1_000_000,
			Source:          source,
			EffectiveDate:   effectiveDate,
			FetchedAt:       &now,
		}

		if err := s.repo.SaveCurrencyExchangeRate(ctx, currencyRate); err != nil {
			return fmt.Errorf("failed to save USD/%s rate: %w", quoteCurrency, err)
		}

		inverseRate := math.Round((1.0/roundedRate)*1e8) / 1e8
		inverseCurrencyRate := &storage.CurrencyExchangeRate{
			BaseCurrency:    quoteCurrency,
			QuoteCurrency:   "USD",
			Rate:            inverseRate,
			RateNumerator:   int64(inverseRate * 1_000_000),
			RateDenominator: 1_000_000,
			Source:          source,
			EffectiveDate:   effectiveDate,
			FetchedAt:       &now,
		}

		if err := s.repo.SaveCurrencyExchangeRate(ctx, inverseCurrencyRate); err != nil {
			return fmt.Errorf("failed to save %s/USD rate: %w", quoteCurrency, err)
		}
	}

	return nil
}

type FrankfurterProvider struct {
	client *http.Client
}

func NewFrankfurterProvider() *FrankfurterProvider {
	return &FrankfurterProvider{
		client: &http.Client{Timeout: 10 * time.Second},
	}
}

func (p *FrankfurterProvider) Name() string {
	return "frankfurter"
}

func (p *FrankfurterProvider) IsAvailable() bool {
	return true
}

func (p *FrankfurterProvider) GetRates(ctx context.Context, base string) (map[string]float64, error) {
	url := fmt.Sprintf("https://api.frankfurter.app/latest?from=%s", base)

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
		Base  string             `json:"base"`
		Date  string             `json:"date"`
		Rates map[string]float64 `json:"rates"`
	}
	if err := json.NewDecoder(resp.Body).Decode(&result); err != nil {
		return nil, err
	}

	result.Rates["EUR"] = 1.0
	return result.Rates, nil
}

type StripeRateProvider struct {
	client *http.Client
}

func NewStripeRateProvider() *StripeRateProvider {
	return &StripeRateProvider{
		client: &http.Client{Timeout: 10 * time.Second},
	}
}

func (p *StripeRateProvider) Name() string {
	return "stripe"
}

func (p *StripeRateProvider) IsAvailable() bool {
	return os.Getenv("STRIPE_SECRET_KEY") != ""
}

func (p *StripeRateProvider) GetRates(ctx context.Context, base string) (map[string]float64, error) {
	key := os.Getenv("STRIPE_SECRET_KEY")
	url := "https://api.stripe.com/v1/exchange_rates"

	req, err := http.NewRequestWithContext(ctx, "GET", url, nil)
	if err != nil {
		return nil, err
	}
	req.SetBasicAuth(key, "")
	req.Header.Set("Content-Type", "application/x-www-form-urlencoded")

	resp, err := p.client.Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("Stripe API returned status %d", resp.StatusCode)
	}

	var result struct {
		Data []struct {
			Currency string             `json:"currency"`
			Rates    map[string]float64 `json:"rates"`
		} `json:"data"`
	}
	if err := json.NewDecoder(resp.Body).Decode(&result); err != nil {
		return nil, err
	}

	rates := make(map[string]float64)
	for _, item := range result.Data {
		for currency, rate := range item.Rates {
			rates[currency] = rate
		}
	}
	return rates, nil
}

type ExchangeRateScheduler struct {
	cron     *cron.Cron
	repo     storage.Repository
	redis    *redis.Client
	syncer   *ExchangeRateSyncer
	logger   *logrus.Logger
	enabled  bool
	syncCron string
	stopOnce sync.Once
	cancel   context.CancelFunc
}

func NewExchangeRateScheduler(repo storage.Repository, redisClient *redis.Client) *ExchangeRateScheduler {
	enabled := os.Getenv("EXCHANGE_RATE_SYNC_ENABLED") != "false"
	syncCron := os.Getenv("EXCHANGE_RATE_SYNC_CRON")
	if syncCron == "" {
		syncCron = "0 0 * * *"
	}
	if _, err := cron.ParseStandard(syncCron); err != nil {
		syncCron = "0 0 * * *"
	}

	return &ExchangeRateScheduler{
		cron:     cron.New(),
		repo:     repo,
		redis:    redisClient,
		syncer:   NewExchangeRateSyncer(repo, redisClient),
		logger:   logrus.New(),
		enabled:  enabled,
		syncCron: syncCron,
	}
}

func (s *ExchangeRateScheduler) SetLogger(logger *logrus.Logger) {
	s.logger = logger
}

func (s *ExchangeRateScheduler) Start(ctx context.Context) error {
	if !s.enabled {
		s.logger.Info("Exchange rate scheduler is disabled")
		return nil
	}

	var ctxWithCancel context.Context
	ctxWithCancel, s.cancel = context.WithCancel(ctx)

	_, err := s.cron.AddFunc(s.syncCron, func() {
		s.runSync(ctxWithCancel)
	})
	if err != nil {
		return fmt.Errorf("failed to add exchange rate sync cron job: %w", err)
	}

	s.cron.Start()

	s.logger.WithFields(logrus.Fields{
		"sync_cron": s.syncCron,
	}).Info("Exchange rate scheduler started")

	return nil
}

func (s *ExchangeRateScheduler) Stop() error {
	s.stopOnce.Do(func() {
		if s.cancel != nil {
			s.cancel()
		}
		<-s.cron.Stop().Done()
		s.logger.Info("Exchange rate scheduler stopped")
	})
	return nil
}

func (s *ExchangeRateScheduler) runSync(ctx context.Context) {
	start := time.Now()
	s.logger.Info("Starting scheduled exchange rate sync")

	if err := s.syncer.SyncAllRates(ctx); err != nil {
		s.logger.WithError(err).Error("Exchange rate sync failed")
		return
	}

	duration := time.Since(start)
	s.logger.WithField("duration_ms", duration.Milliseconds()).Info("Exchange rate sync completed")
}

func (s *ExchangeRateScheduler) SyncNow(ctx context.Context) error {
	s.logger.Info("Manually triggering exchange rate sync")
	return s.syncer.SyncAllRates(ctx)
}

func (s *ExchangeRateScheduler) GetSchedule() map[string]interface{} {
	nextSync := "unknown"
	if s.cron != nil {
		entries := s.cron.Entries()
		if len(entries) > 0 {
			nextSync = entries[0].Next.Format(time.RFC3339)
		}
	}

	return map[string]interface{}{
		"enabled":       s.enabled,
		"sync_cron":     s.syncCron,
		"next_sync_run": nextSync,
	}
}

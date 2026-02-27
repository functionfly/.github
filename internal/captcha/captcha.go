package captcha

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"time"

	"github.com/google/uuid"
	"github.com/sirupsen/logrus"
)

// CaptchaProvider defines the interface for CAPTCHA providers
type CaptchaProvider interface {
	VerifyToken(ctx context.Context, token, remoteIP string) (*CaptchaResult, error)
	GenerateChallenge() (*CaptchaChallenge, error)
	GetProviderName() string
}

// CaptchaResult represents the result of CAPTCHA verification
type CaptchaResult struct {
	Success     bool    `json:"success"`
	Score       float64 `json:"score,omitempty"`       // For reCAPTCHA v3
	Action      string  `json:"action,omitempty"`      // For reCAPTCHA v3
	ChallengeTS string  `json:"challenge_ts,omitempty"`
	Hostname    string  `json:"hostname,omitempty"`
	ErrorCodes  []string `json:"error-codes,omitempty"`
}

// CaptchaChallenge represents a CAPTCHA challenge to be presented to the user
type CaptchaChallenge struct {
	ChallengeID string                 `json:"challenge_id"`
	Type        string                 `json:"type"` // "recaptcha_v2", "recaptcha_v3", "hcaptcha", "turnstile"
	SiteKey     string                 `json:"site_key"`
	Data        map[string]interface{} `json:"data,omitempty"`
}

// CaptchaService manages CAPTCHA operations with multiple providers
type CaptchaService struct {
	providers map[string]CaptchaProvider
	defaultProvider string
	logger     *logrus.Logger
}

// NewCaptchaService creates a new CAPTCHA service
func NewCaptchaService(logger *logrus.Logger) *CaptchaService {
	return &CaptchaService{
		providers: make(map[string]CaptchaProvider),
		logger:    logger,
	}
}

// RegisterProvider registers a CAPTCHA provider
func (cs *CaptchaService) RegisterProvider(provider CaptchaProvider) {
	name := provider.GetProviderName()
	cs.providers[name] = provider
	if cs.defaultProvider == "" {
		cs.defaultProvider = name
	}
}

// SetDefaultProvider sets the default CAPTCHA provider
func (cs *CaptchaService) SetDefaultProvider(providerName string) {
	if _, exists := cs.providers[providerName]; exists {
		cs.defaultProvider = providerName
	}
}

// VerifyToken verifies a CAPTCHA token using the specified or default provider
func (cs *CaptchaService) VerifyToken(ctx context.Context, token, remoteIP, providerName string) (*CaptchaResult, error) {
	if providerName == "" {
		providerName = cs.defaultProvider
	}

	provider, exists := cs.providers[providerName]
	if !exists {
		return nil, fmt.Errorf("CAPTCHA provider '%s' not found", providerName)
	}

	result, err := provider.VerifyToken(ctx, token, remoteIP)
	if err != nil {
		cs.logger.WithFields(logrus.Fields{
			"provider": providerName,
			"error":    err.Error(),
		}).Error("CAPTCHA verification failed")
		return nil, err
	}

	cs.logger.WithFields(logrus.Fields{
		"provider": providerName,
		"success":  result.Success,
		"score":    result.Score,
	}).Debug("CAPTCHA verification completed")

	return result, nil
}

// GenerateChallenge generates a CAPTCHA challenge
func (cs *CaptchaService) GenerateChallenge(providerName string) (*CaptchaChallenge, error) {
	if providerName == "" {
		providerName = cs.defaultProvider
	}

	provider, exists := cs.providers[providerName]
	if !exists {
		return nil, fmt.Errorf("CAPTCHA provider '%s' not found", providerName)
	}

	challenge, err := provider.GenerateChallenge()
	if err != nil {
		return nil, err
	}

	// Add unique challenge ID if not provided
	if challenge.ChallengeID == "" {
		challenge.ChallengeID = uuid.New().String()
	}

	return challenge, nil
}

// GetAvailableProviders returns the list of available CAPTCHA providers
func (cs *CaptchaService) GetAvailableProviders() []string {
	providers := make([]string, 0, len(cs.providers))
	for name := range cs.providers {
		providers = append(providers, name)
	}
	return providers
}

// ============================================
// reCAPTCHA v2 Provider
// ============================================

type RecaptchaV2Provider struct {
	siteKey   string
	secretKey string
	logger    *logrus.Logger
}

func NewRecaptchaV2Provider(siteKey, secretKey string, logger *logrus.Logger) *RecaptchaV2Provider {
	return &RecaptchaV2Provider{
		siteKey:   siteKey,
		secretKey: secretKey,
		logger:    logger,
	}
}

func (r *RecaptchaV2Provider) VerifyToken(ctx context.Context, token, remoteIP string) (*CaptchaResult, error) {
	data := map[string]string{
		"secret":   r.secretKey,
		"response": token,
	}
	if remoteIP != "" {
		data["remoteip"] = remoteIP
	}

	jsonData, err := json.Marshal(data)
	if err != nil {
		return nil, err
	}

	req, err := http.NewRequestWithContext(ctx, "POST", "https://www.google.com/recaptcha/api/siteverify", bytes.NewBuffer(jsonData))
	if err != nil {
		return nil, err
	}

	client := &http.Client{Timeout: 10 * time.Second}
	resp, err := client.Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, err
	}

	var result CaptchaResult
	if err := json.Unmarshal(body, &result); err != nil {
		return nil, err
	}

	return &result, nil
}

func (r *RecaptchaV2Provider) GenerateChallenge() (*CaptchaChallenge, error) {
	return &CaptchaChallenge{
		Type:    "recaptcha_v2",
		SiteKey: r.siteKey,
	}, nil
}

func (r *RecaptchaV2Provider) GetProviderName() string {
	return "recaptcha_v2"
}

// ============================================
// reCAPTCHA v3 Provider
// ============================================

type RecaptchaV3Provider struct {
	siteKey   string
	secretKey string
	logger    *logrus.Logger
}

func NewRecaptchaV3Provider(siteKey, secretKey string, logger *logrus.Logger) *RecaptchaV3Provider {
	return &RecaptchaV3Provider{
		siteKey:   siteKey,
		secretKey: secretKey,
		logger:    logger,
	}
}

func (r *RecaptchaV3Provider) VerifyToken(ctx context.Context, token, remoteIP string) (*CaptchaResult, error) {
	data := map[string]string{
		"secret":   r.secretKey,
		"response": token,
	}
	if remoteIP != "" {
		data["remoteip"] = remoteIP
	}

	jsonData, err := json.Marshal(data)
	if err != nil {
		return nil, err
	}

	req, err := http.NewRequestWithContext(ctx, "POST", "https://www.google.com/recaptcha/api/siteverify", bytes.NewBuffer(jsonData))
	if err != nil {
		return nil, err
	}

	client := &http.Client{Timeout: 10 * time.Second}
	resp, err := client.Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, err
	}

	var result CaptchaResult
	if err := json.Unmarshal(body, &result); err != nil {
		return nil, err
	}

	return &result, nil
}

func (r *RecaptchaV3Provider) GenerateChallenge() (*CaptchaChallenge, error) {
	return &CaptchaChallenge{
		Type:    "recaptcha_v3",
		SiteKey: r.siteKey,
	}, nil
}

func (r *RecaptchaV3Provider) GetProviderName() string {
	return "recaptcha_v3"
}

// ============================================
// hCaptcha Provider
// ============================================

type HCaptchaProvider struct {
	siteKey   string
	secretKey string
	logger    *logrus.Logger
}

func NewHCaptchaProvider(siteKey, secretKey string, logger *logrus.Logger) *HCaptchaProvider {
	return &HCaptchaProvider{
		siteKey:   siteKey,
		secretKey: secretKey,
		logger:    logger,
	}
}

func (h *HCaptchaProvider) VerifyToken(ctx context.Context, token, remoteIP string) (*CaptchaResult, error) {
	data := map[string]string{
		"secret":   h.secretKey,
		"response": token,
	}
	if remoteIP != "" {
		data["remoteip"] = remoteIP
	}

	jsonData, err := json.Marshal(data)
	if err != nil {
		return nil, err
	}

	req, err := http.NewRequestWithContext(ctx, "POST", "https://hcaptcha.com/siteverify", bytes.NewBuffer(jsonData))
	if err != nil {
		return nil, err
	}

	client := &http.Client{Timeout: 10 * time.Second}
	resp, err := client.Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, err
	}

	var result CaptchaResult
	if err := json.Unmarshal(body, &result); err != nil {
		return nil, err
	}

	return &result, nil
}

func (h *HCaptchaProvider) GenerateChallenge() (*CaptchaChallenge, error) {
	return &CaptchaChallenge{
		Type:    "hcaptcha",
		SiteKey: h.siteKey,
	}, nil
}

func (h *HCaptchaProvider) GetProviderName() string {
	return "hcaptcha"
}

// ============================================
// Cloudflare Turnstile Provider
// ============================================

type TurnstileProvider struct {
	siteKey   string
	secretKey string
	logger    *logrus.Logger
}

func NewTurnstileProvider(siteKey, secretKey string, logger *logrus.Logger) *TurnstileProvider {
	return &TurnstileProvider{
		siteKey:   siteKey,
		secretKey: secretKey,
		logger:    logger,
	}
}

func (t *TurnstileProvider) VerifyToken(ctx context.Context, token, remoteIP string) (*CaptchaResult, error) {
	data := map[string]string{
		"secret":   t.secretKey,
		"response": token,
	}
	if remoteIP != "" {
		data["remoteip"] = remoteIP
	}

	jsonData, err := json.Marshal(data)
	if err != nil {
		return nil, err
	}

	req, err := http.NewRequestWithContext(ctx, "POST", "https://challenges.cloudflare.com/turnstile/v0/siteverify", bytes.NewBuffer(jsonData))
	if err != nil {
		return nil, err
	}

	client := &http.Client{Timeout: 10 * time.Second}
	resp, err := client.Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, err
	}

	var result CaptchaResult
	if err := json.Unmarshal(body, &result); err != nil {
		return nil, err
	}

	return &result, nil
}

func (t *TurnstileProvider) GenerateChallenge() (*CaptchaChallenge, error) {
	return &CaptchaChallenge{
		Type:    "turnstile",
		SiteKey: t.siteKey,
	}, nil
}

func (t *TurnstileProvider) GetProviderName() string {
	return "turnstile"
}

// ============================================
// Mock Provider for Testing
// ============================================

type MockCaptchaProvider struct {
	shouldSucceed bool
	score         float64
	logger        *logrus.Logger
}

func NewMockCaptchaProvider(shouldSucceed bool, score float64, logger *logrus.Logger) *MockCaptchaProvider {
	return &MockCaptchaProvider{
		shouldSucceed: shouldSucceed,
		score:         score,
		logger:        logger,
	}
}

func (m *MockCaptchaProvider) VerifyToken(ctx context.Context, token, remoteIP string) (*CaptchaResult, error) {
	return &CaptchaResult{
		Success: m.shouldSucceed,
		Score:   m.score,
	}, nil
}

func (m *MockCaptchaProvider) GenerateChallenge() (*CaptchaChallenge, error) {
	return &CaptchaChallenge{
		Type:    "mock",
		SiteKey: "mock_site_key",
	}, nil
}

func (m *MockCaptchaProvider) GetProviderName() string {
	return "mock"
}
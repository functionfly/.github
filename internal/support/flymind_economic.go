package support

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"time"

	"github.com/sirupsen/logrus"
)

// EconomicRoutingConfig holds configuration for economic routing
type EconomicRoutingConfig struct {
	BaseURL          string        // AI service base URL
	APIKey           string        // API key for authentication
	Timeout          time.Duration // Request timeout
	DefaultStrategy  string        // Default routing strategy
	QualityThreshold float64       // Minimum quality threshold
}

// DefaultEconomicRoutingConfig returns default configuration
func DefaultEconomicRoutingConfig() *EconomicRoutingConfig {
	return &EconomicRoutingConfig{
		BaseURL:          "http://localhost:18081", // ai-service default
		APIKey:           "",
		Timeout:          30 * time.Second,
		DefaultStrategy:  "balanced",
		QualityThreshold: 0.7,
	}
}

// EconomicMemoryScore represents cost-quality metrics for a provider/model
type EconomicMemoryScore struct {
	Provider           string  `json:"provider"`
	Model              string  `json:"model"`
	CostQualityIndex   float64 `json:"cost_quality_index"`
	AvgCostPer1KTokens float64 `json:"avg_cost_per_1k_tokens"`
	AvgCostPerRequest  float64 `json:"avg_cost_per_request"`
	QualityScore       float64 `json:"quality_score"`
	ResponseTimeScore  float64 `json:"response_time_score"`
	SuccessRate        float64 `json:"success_rate"`
	TotalExecutions    int     `json:"total_executions"`
	TotalCostUSD       float64 `json:"total_cost_usd"`
	Recommendation     string  `json:"recommendation"` // highly_recommended, recommended, neutral, avoid
	Confidence         string  `json:"confidence"`     // high, medium, low
}

// EconomicRoutingResponse represents a routing decision from the AI service
type EconomicRoutingResponse struct {
	Provider           string   `json:"provider"`
	Model              string   `json:"model"`
	Strategy           string   `json:"strategy"`
	CostQualityIndex   float64  `json:"cost_quality_index"`
	EstimatedCostPer1K float64  `json:"estimated_cost_per_1k"`
	EstimatedQuality   float64  `json:"estimated_quality"`
	Confidence         string   `json:"confidence"`
	Reasoning          string   `json:"reasoning"`
	Alternatives       []string `json:"alternatives"`
}

// ModelRecommendationResponse represents a model recommendation
type ModelRecommendationResponse struct {
	CurrentModel        string   `json:"current_model"`
	Recommendation      string   `json:"recommendation"` // keep_current, upgrade_suggested, downgrade_suggested
	SuggestedModel      *string  `json:"suggested_model,omitempty"`
	CurrentCostPer1K    *float64 `json:"current_cost_per_1k,omitempty"`
	SuggestedCostPer1K  *float64 `json:"suggested_cost_per_1k,omitempty"`
	PotentialSavingsPct *float64 `json:"potential_savings_percent,omitempty"`
	QualityDelta        *float64 `json:"quality_delta,omitempty"`
	Message             string   `json:"message"`
}

// CostSavingsOpportunity represents cost savings analysis
type CostSavingsOpportunity struct {
	PeriodDays                int      `json:"period_days"`
	Analysis                  string   `json:"analysis"`
	CurrentPeriodCost         float64  `json:"current_period_cost"`
	ExecutionsAnalyzed        int      `json:"executions_analyzed"`
	BestValueProvider         *string  `json:"best_value_provider,omitempty"`
	BestValueCQI              *float64 `json:"best_value_cqi,omitempty"`
	EstimatedMonthlySavings   float64  `json:"estimated_monthly_savings"`
	OptimizationOpportunities []string `json:"optimization_opportunities"`
}

// FlyMindClient is the Go client for the FlyMind AI Service with economic routing
type FlyMindClient struct {
	config EconomicRoutingConfig
	client *http.Client
	logger *logrus.Logger
}

// NewFlyMindClient creates a new FlyMind client for economic routing
func NewFlyMindClient(config *EconomicRoutingConfig, logger *logrus.Logger) *FlyMindClient {
	if config == nil {
		config = DefaultEconomicRoutingConfig()
	}
	if logger == nil {
		logger = logrus.New()
	}

	return &FlyMindClient{
		config: *config,
		client: &http.Client{
			Timeout: config.Timeout,
		},
		logger: logger,
	}
}

// GetEconomicScores returns all cost-quality scores from the economic memory
func (c *FlyMindClient) GetEconomicScores(ctx context.Context) ([]EconomicMemoryScore, error) {
	req, err := http.NewRequestWithContext(ctx, "GET",
		fmt.Sprintf("%s/api/economic-memory/scores", c.config.BaseURL),
		nil)
	if err != nil {
		return nil, fmt.Errorf("failed to create request: %w", err)
	}

	if c.config.APIKey != "" {
		req.Header.Set("Authorization", fmt.Sprintf("Bearer %s", c.config.APIKey))
	}

	resp, err := c.client.Do(req)
	if err != nil {
		return nil, fmt.Errorf("failed to call economic memory: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("economic memory returned status %d", resp.StatusCode)
	}

	var result struct {
		Providers []EconomicMemoryScore `json:"providers"`
	}
	if err := json.NewDecoder(resp.Body).Decode(&result); err != nil {
		return nil, fmt.Errorf("failed to decode response: %w", err)
	}

	return result.Providers, nil
}

// GetEconomicRouting returns cost-intelligent routing recommendation
func (c *FlyMindClient) GetEconomicRouting(
	ctx context.Context,
	functionID string,
	strategy string,
	qualityThreshold *float64,
) (*EconomicRoutingResponse, error) {
	if strategy == "" {
		strategy = c.config.DefaultStrategy
	}
	if qualityThreshold == nil {
		threshold := c.config.QualityThreshold
		qualityThreshold = &threshold
	}

	// Build request
	reqBody := map[string]interface{}{
		"function_id":       functionID,
		"strategy":          strategy,
		"quality_threshold": *qualityThreshold,
	}

	body, err := json.Marshal(reqBody)
	if err != nil {
		return nil, fmt.Errorf("failed to marshal request: %w", err)
	}

	req, err := http.NewRequestWithContext(ctx, "POST",
		fmt.Sprintf("%s/api/economic-memory/route", c.config.BaseURL),
		bytes.NewBuffer(body))
	if err != nil {
		return nil, fmt.Errorf("failed to create request: %w", err)
	}

	req.Header.Set("Content-Type", "application/json")
	if c.config.APIKey != "" {
		req.Header.Set("Authorization", fmt.Sprintf("Bearer %s", c.config.APIKey))
	}

	resp, err := c.client.Do(req)
	if err != nil {
		return nil, fmt.Errorf("failed to call economic routing: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("economic routing returned status %d", resp.StatusCode)
	}

	var routingResp EconomicRoutingResponse
	if err := json.NewDecoder(resp.Body).Decode(&routingResp); err != nil {
		return nil, fmt.Errorf("failed to decode response: %w", err)
	}

	c.logger.WithFields(logrus.Fields{
		"function_id":          functionID,
		"recommended_provider": routingResp.Provider,
		"recommended_model":    routingResp.Model,
		"cqi":                  routingResp.CostQualityIndex,
		"strategy":             strategy,
	}).Debug("Received economic routing recommendation")

	return &routingResp, nil
}

// GetModelRecommendation returns a model recommendation for a provider
func (c *FlyMindClient) GetModelRecommendation(
	ctx context.Context,
	provider string,
	currentModel string,
	targetQuality float64,
) (*ModelRecommendationResponse, error) {
	url := fmt.Sprintf(
		"%s/api/economic-memory/recommendation?provider=%s&current_model=%s&target_quality=%.2f",
		c.config.BaseURL,
		provider,
		currentModel,
		targetQuality,
	)

	req, err := http.NewRequestWithContext(ctx, "GET", url, nil)
	if err != nil {
		return nil, fmt.Errorf("failed to create request: %w", err)
	}

	if c.config.APIKey != "" {
		req.Header.Set("Authorization", fmt.Sprintf("Bearer %s", c.config.APIKey))
	}

	resp, err := c.client.Do(req)
	if err != nil {
		return nil, fmt.Errorf("failed to call model recommendation: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("model recommendation returned status %d", resp.StatusCode)
	}

	var recResp ModelRecommendationResponse
	if err := json.NewDecoder(resp.Body).Decode(&recResp); err != nil {
		return nil, fmt.Errorf("failed to decode response: %w", err)
	}

	c.logger.WithFields(logrus.Fields{
		"provider":       provider,
		"current_model":  currentModel,
		"recommendation": recResp.Recommendation,
	}).Debug("Received model recommendation")

	return &recResp, nil
}

// GetCostSavingsOpportunity returns cost savings analysis
func (c *FlyMindClient) GetCostSavingsOpportunity(
	ctx context.Context,
	tenantID *string,
	days int,
) (*CostSavingsOpportunity, error) {
	url := fmt.Sprintf("%s/api/economic-memory/savings?days=%d", c.config.BaseURL, days)
	if tenantID != nil {
		url = fmt.Sprintf("%s&tenant_id=%s", url, *tenantID)
	}

	req, err := http.NewRequestWithContext(ctx, "GET", url, nil)
	if err != nil {
		return nil, fmt.Errorf("failed to create request: %w", err)
	}

	if c.config.APIKey != "" {
		req.Header.Set("Authorization", fmt.Sprintf("Bearer %s", c.config.APIKey))
	}

	resp, err := c.client.Do(req)
	if err != nil {
		return nil, fmt.Errorf("failed to call savings analysis: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("savings analysis returned status %d", resp.StatusCode)
	}

	var savingsResp CostSavingsOpportunity
	if err := json.NewDecoder(resp.Body).Decode(&savingsResp); err != nil {
		return nil, fmt.Errorf("failed to decode response: %w", err)
	}

	c.logger.WithFields(logrus.Fields{
		"period_days":               savingsResp.PeriodDays,
		"current_cost":              savingsResp.CurrentPeriodCost,
		"estimated_monthly_savings": savingsResp.EstimatedMonthlySavings,
	}).Debug("Received cost savings analysis")

	return &savingsResp, nil
}

// GetBestValueProvider returns the provider with the best cost-quality index
func (c *FlyMindClient) GetBestValueProvider(ctx context.Context) (*EconomicMemoryScore, error) {
	scores, err := c.GetEconomicScores(ctx)
	if err != nil {
		return nil, err
	}

	if len(scores) == 0 {
		return nil, fmt.Errorf("no economic scores available")
	}

	// Find best by CQI
	best := &scores[0]
	for i := range scores {
		if scores[i].CostQualityIndex > best.CostQualityIndex {
			best = &scores[i]
		}
	}

	return best, nil
}

// EconomicRoutingStrategy defines available routing strategies
type EconomicRoutingStrategy string

const (
	// QualityFirst maximizes quality regardless of cost
	QualityFirst EconomicRoutingStrategy = "quality_first"
	// Balanced balances cost and quality (default)
	Balanced EconomicRoutingStrategy = "balanced"
	// CostOptimized minimizes cost while meeting quality threshold
	CostOptimized EconomicRoutingStrategy = "cost_optimized"
	// CostFirst minimizes cost (may reduce quality)
	CostFirst EconomicRoutingStrategy = "cost_first"
)

// SelectBestProviderWithEconomicRouting selects the best provider considering
// cost-quality metrics from the economic memory
func (c *FlyMindClient) SelectBestProviderWithEconomicRouting(
	ctx context.Context,
	functionID string,
	strategy EconomicRoutingStrategy,
) (*EconomicRoutingResponse, error) {
	qualityThreshold := c.config.QualityThreshold

	// Adjust threshold based on strategy
	if strategy == QualityFirst {
		qualityThreshold = 0.85
	} else if strategy == CostFirst {
		qualityThreshold = 0.5
	}

	return c.GetEconomicRouting(ctx, functionID, string(strategy), &qualityThreshold)
}

// IsEconomicRoutingEnabled checks if the economic memory service is available
func (c *FlyMindClient) IsEconomicRoutingEnabled(ctx context.Context) bool {
	req, err := http.NewRequestWithContext(ctx, "GET",
		fmt.Sprintf("%s/api/economic-memory/health", c.config.BaseURL),
		nil)
	if err != nil {
		return false
	}

	resp, err := c.client.Do(req)
	if err != nil {
		return false
	}
	defer resp.Body.Close()

	return resp.StatusCode == http.StatusOK
}

// GetEconomicHealth returns health status and statistics
func (c *FlyMindClient) GetEconomicHealth(ctx context.Context) (map[string]interface{}, error) {
	req, err := http.NewRequestWithContext(ctx, "GET",
		fmt.Sprintf("%s/api/economic-memory/health", c.config.BaseURL),
		nil)
	if err != nil {
		return nil, fmt.Errorf("failed to create request: %w", err)
	}

	resp, err := c.client.Do(req)
	if err != nil {
		return nil, fmt.Errorf("failed to call health endpoint: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("health check failed with status %d", resp.StatusCode)
	}

	var health map[string]interface{}
	if err := json.NewDecoder(resp.Body).Decode(&health); err != nil {
		return nil, fmt.Errorf("failed to decode health response: %w", err)
	}

	return health, nil
}

// ProviderScoreCache provides caching for economic scores
type ProviderScoreCache struct {
	scores   []EconomicMemoryScore
	cachedAt time.Time
	ttl      time.Duration
	client   *FlyMindClient
}

// NewProviderScoreCache creates a new score cache
func NewProviderScoreCache(client *FlyMindClient, ttl time.Duration) *ProviderScoreCache {
	return &ProviderScoreCache{
		client: client,
		ttl:    ttl,
	}
}

// GetScores returns cached scores or fetches fresh ones
func (c *ProviderScoreCache) GetScores(ctx context.Context) ([]EconomicMemoryScore, error) {
	if c.scores != nil && time.Since(c.cachedAt) < c.ttl {
		return c.scores, nil
	}

	scores, err := c.client.GetEconomicScores(ctx)
	if err != nil {
		// Return stale data if available
		if c.scores != nil {
			return c.scores, nil
		}
		return nil, err
	}

	c.scores = scores
	c.cachedAt = time.Now()
	return scores, nil
}

// Invalidate clears the cache
func (c *ProviderScoreCache) Invalidate() {
	c.scores = nil
	c.cachedAt = time.Time{}
}

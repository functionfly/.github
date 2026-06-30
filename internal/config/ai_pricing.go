package config

import (
	"encoding/json"
	"os"
	"strconv"

	"github.com/sirupsen/logrus"
)

// AICostRate holds per-token costs for a model (USD per 1,000 tokens).
type AICostRate struct {
	InputCostPer1K  float64 `json:"input_cost_per_1k"`
	OutputCostPer1K float64 `json:"output_cost_per_1k"`
}

// AIModelCostConfig holds all AI provider cost rates and platform markup settings.
// This is the single source of truth for Go-side wallet charging calculations.
type AIModelCostConfig struct {
	// Rates: provider -> model -> cost rate
	Rates map[string]map[string]AICostRate `json:"rates"`
	// MarkupPercent is the platform markup applied on top of provider costs (e.g. 25 = 25%)
	MarkupPercent int `json:"markup_percent"`
	// DefaultRates is the provider-level fallback when a specific model is not found
	DefaultRates map[string]AICostRate `json:"default_rates"`
}

// DefaultAIModelCostConfig returns the standard AI pricing config with curated model rates
// that mirror the ai-service/src/services/economic_memory/tracking.py rates.
func DefaultAIModelCostConfig() *AIModelCostConfig {
	cfg := &AIModelCostConfig{
		Rates: map[string]map[string]AICostRate{
			"openai": {
				"gpt-4o-mini":         {InputCostPer1K: 0.00015, OutputCostPer1K: 0.0006},
				"gpt-4o":             {InputCostPer1K: 0.0025, OutputCostPer1K: 0.01},
				"gpt-4-turbo":        {InputCostPer1K: 0.01, OutputCostPer1K: 0.03},
				"gpt-3.5-turbo":      {InputCostPer1K: 0.0005, OutputCostPer1K: 0.0015},
			},
			"anthropic": {
				"claude-3-5-haiku":    {InputCostPer1K: 0.00025, OutputCostPer1K: 0.00125},
				"claude-3-5-sonnet": {InputCostPer1K: 0.003, OutputCostPer1K: 0.015},
				"claude-3-5-arc":    {InputCostPer1K: 0.003, OutputCostPer1K: 0.015},
				"claude-3-opus":     {InputCostPer1K: 0.015, OutputCostPer1K: 0.075},
				"claude-3-sonnet":   {InputCostPer1K: 0.003, OutputCostPer1K: 0.015},
				"claude-3-haiku":    {InputCostPer1K: 0.00025, OutputCostPer1K: 0.00125},
			},
			"groq": {
				"llama-3.1-8b-instruct":  {InputCostPer1K: 0.00005, OutputCostPer1K: 0.00005},
				"llama-3.3-70b-instruct": {InputCostPer1K: 0.00024, OutputCostPer1K: 0.00024},
				"mixtral-8x7b-32768":     {InputCostPer1K: 0.00024, OutputCostPer1K: 0.00024},
				"gemma2-9b-it":           {InputCostPer1K: 0.00006, OutputCostPer1K: 0.00006},
			},
			"fireworks": {
				"accounts/fireworks/models/llama-v3p1-8b-instruct":  {InputCostPer1K: 0.0002, OutputCostPer1K: 0.0002},
				"accounts/fireworks/models/llama-v3p1-70b-instruct": {InputCostPer1K: 0.0009, OutputCostPer1K: 0.0009},
				"accounts/fireworks/models/mixtral-8x22b-instruct":  {InputCostPer1K: 0.0009, OutputCostPer1K: 0.0009},
				"accounts/fireworks/models/qwen2p5-72b-instruct":    {InputCostPer1K: 0.0009, OutputCostPer1K: 0.0009},
			},
			"deepinfra": {
				"deepseek-ai/deepseek-r1":           {InputCostPer1K: 0.0001, OutputCostPer1K: 0.00027},
				"deepseek-ai/deepseek-r1-llama70b":  {InputCostPer1K: 0.0004, OutputCostPer1K: 0.0008},
				"meta-llama/llama-3.1-405b-instruct": {InputCostPer1K: 0.00035, OutputCostPer1K: 0.00035},
			},
			"together": {
				"together/llama-3.1-8b-instruct":   {InputCostPer1K: 0.0002, OutputCostPer1K: 0.0002},
				"together/llama-3.1-70b-instruct":  {InputCostPer1K: 0.00065, OutputCostPer1K: 0.00065},
				"together/llama-3.1-405b-instruct": {InputCostPer1K: 0.001, OutputCostPer1K: 0.001},
				"together/mixtral-8x22b-instruct":  {InputCostPer1K: 0.0009, OutputCostPer1K: 0.0009},
			},
			"openrouter": {
				"openai/gpt-4o-mini":           {InputCostPer1K: 0.00015, OutputCostPer1K: 0.0006},
				"openai/gpt-4o":                {InputCostPer1K: 0.0025, OutputCostPer1K: 0.01},
				"anthropic/claude-3.5-sonnet":  {InputCostPer1K: 0.003, OutputCostPer1K: 0.015},
				"deepseek/deepseek-r1":          {InputCostPer1K: 0.0001, OutputCostPer1K: 0.00027},
				"google/gemini-2.0-flash-exp":  {InputCostPer1K: 0.0, OutputCostPer1K: 0.0},
			},
			"mimo": {
				"mimo-v2-flash":  {InputCostPer1K: 0.0001, OutputCostPer1K: 0.0003},
				"mimo-v2.5-pro": {InputCostPer1K: 0.001, OutputCostPer1K: 0.003},
				"mimo-omni":     {InputCostPer1K: 0.0004, OutputCostPer1K: 0.0008},
			},
			"ollama": {
				"llama3.1:latest":   {InputCostPer1K: 0.0, OutputCostPer1K: 0.0},
				"llama3.2:latest":   {InputCostPer1K: 0.0, OutputCostPer1K: 0.0},
				"codellama:latest":   {InputCostPer1K: 0.0, OutputCostPer1K: 0.0},
			},
		},
		DefaultRates: map[string]AICostRate{
			"openai":     {InputCostPer1K: 0.002, OutputCostPer1K: 0.006},
			"anthropic":  {InputCostPer1K: 0.003, OutputCostPer1K: 0.015},
			"groq":       {InputCostPer1K: 0.0001, OutputCostPer1K: 0.0001},
			"fireworks":  {InputCostPer1K: 0.0002, OutputCostPer1K: 0.0002},
			"deepinfra":  {InputCostPer1K: 0.0001, OutputCostPer1K: 0.0001},
			"together":   {InputCostPer1K: 0.0002, OutputCostPer1K: 0.0002},
			"openrouter": {InputCostPer1K: 0.001, OutputCostPer1K: 0.003},
			"mimo":       {InputCostPer1K: 0.0004, OutputCostPer1K: 0.0008},
			"ollama":     {InputCostPer1K: 0.0, OutputCostPer1K: 0.0},
		},
		MarkupPercent: 25, // 25% default platform markup
	}
	return cfg
}

// LoadAIModelCostConfig loads AI cost configuration from environment variables.
// Priority: AI_COST_RATES_JSON env var > defaults.
func LoadAIModelCostConfig() *AIModelCostConfig {
	cfg := DefaultAIModelCostConfig()

	if v := os.Getenv("AI_MARKUP_PERCENT"); v != "" {
		if pct, err := strconv.Atoi(v); err == nil {
			cfg.MarkupPercent = pct
			logrus.WithField("markup_percent", pct).Info("ai_pricing: loaded markup from AI_MARKUP_PERCENT")
		}
	}

	if jsonStr := os.Getenv("AI_COST_RATES_JSON"); jsonStr != "" {
		var override map[string]map[string]AICostRate
		if err := json.Unmarshal([]byte(jsonStr), &override); err == nil {
			for provider, models := range override {
				if cfg.Rates[provider] == nil {
					cfg.Rates[provider] = map[string]AICostRate{}
				}
				for model, rate := range models {
					cfg.Rates[provider][model] = rate
				}
			}
			logrus.Info("ai_pricing: loaded custom rates from AI_COST_RATES_JSON")
		} else {
			logrus.WithError(err).Warn("ai_pricing: failed to parse AI_COST_RATES_JSON, using defaults")
		}
	}

	return cfg
}

// GetCostForModel returns the cost rate for a given provider and model.
// Falls back to the provider's default rate if the specific model is not found.
func (c *AIModelCostConfig) GetCostForModel(provider, model string) AICostRate {
	provider = normalizeProvider(provider)
	model = normalizeModel(model)

	if models, ok := c.Rates[provider]; ok {
		if rate, ok := models[model]; ok {
			return rate
		}
		// Substring match as fallback
		for registeredModel, rate := range models {
			if contains(model, registeredModel) || contains(registeredModel, model) {
				return rate
			}
		}
	}

	if defaultRate, ok := c.DefaultRates[provider]; ok {
		return defaultRate
	}

	logrus.WithFields(logrus.Fields{"provider": provider, "model": model}).
		Warn("ai_pricing: no cost rate found, using zero cost")
	return AICostRate{InputCostPer1K: 0, OutputCostPer1K: 0}
}

// ComputeCost calculates the base provider cost (before markup) for given usage.
func (c *AIModelCostConfig) ComputeCost(provider, model string, inputTokens, outputTokens int) float64 {
	rate := c.GetCostForModel(provider, model)
	inputCost := (float64(inputTokens) / 1000.0) * rate.InputCostPer1K
	outputCost := (float64(outputTokens) / 1000.0) * rate.OutputCostPer1K
	return inputCost + outputCost
}

// ComputeCostWithMarkup calculates the total cost including platform markup.
func (c *AIModelCostConfig) ComputeCostWithMarkup(provider, model string, inputTokens, outputTokens int) float64 {
	baseCost := c.ComputeCost(provider, model, inputTokens, outputTokens)
	return baseCost * c.MarkupMultiplier()
}

// ComputeCostBYOK returns 0 for BYOK calls (user pays provider directly).
func (c *AIModelCostConfig) ComputeCostBYOK(provider, model string, inputTokens, outputTokens int, isBYOK bool) float64 {
	if isBYOK {
		return 0
	}
	return c.ComputeCostWithMarkup(provider, model, inputTokens, outputTokens)
}

// MarkupMultiplier returns the markup as a multiplier (e.g. 1.25 for 25%).
func (c *AIModelCostConfig) MarkupMultiplier() float64 {
	return 1.0 + (float64(c.MarkupPercent) / 100.0)
}

// AI_PROVIDER_COST_RATES is the global singleton loaded at startup.
var AI_PROVIDER_COST_RATES *AIModelCostConfig

// GetAIProviderCostRates returns the global AI cost config singleton.
func GetAIProviderCostRates() *AIModelCostConfig {
	if AI_PROVIDER_COST_RATES == nil {
		AI_PROVIDER_COST_RATES = LoadAIModelCostConfig()
	}
	return AI_PROVIDER_COST_RATES
}

// normalizeProvider normalizes provider names for lookup.
func normalizeProvider(p string) string {
	switch p {
	case "openai":
		return "openai"
	case "anthropic":
		return "anthropic"
	case "groq":
		return "groq"
	case "fireworks", "fw":
		return "fireworks"
	case "deepinfra", "deepseek":
		return "deepinfra"
	case "together":
		return "together"
	case "openrouter", "or":
		return "openrouter"
	case "mimo":
		return "mimo"
	case "ollama", "local":
		return "ollama"
	default:
		return p
	}
}

// normalizeModel normalizes model names for lookup.
func normalizeModel(m string) string {
	// Remove version suffixes for matching
	// e.g. "gpt-4o-2024-08-06" → "gpt-4o"
	m = removeSuffix(m, "-2024-08-06", "-2024-07-18", "-latest")
	return m
}

func contains(s, substr string) bool {
	return len(substr) > 0 && len(s) >= len(substr) &&
		(s == substr || len(s) > len(substr) &&
			(s[:len(substr)] == substr || s[len(s)-len(substr):] == substr))
}

func removeSuffix(s string, suffixes ...string) string {
	for _, suf := range suffixes {
		if len(s) > len(suf) && s[len(s)-len(suf):] == suf {
			return s[:len(s)-len(suf)]
		}
	}
	return s
}

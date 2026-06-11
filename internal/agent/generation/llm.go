package generation

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"os"
	"strconv"
	"strings"
	"time"

	"github.com/redis/go-redis/v9"
)

const (
	defaultOpenRouterBaseURL = "https://openrouter.ai/api/v1"
	// Free models on OpenRouter (prefer poolside for general, nemotron nano for reasoning)
	defaultSimpleModel       = "poolside/laguna-xs.2:free"
	defaultComplexModel      = "nvidia/nemotron-3-nano-omni-30b-a3b-reasoning:free"
	defaultGenerationTimeout = 120 * time.Second
	defaultCacheTTL          = 30 * time.Minute
	defaultSimilarityCutoff  = 0.92
	manualReviewQualityFloor = 70.0
	// Temperature settings
	mercury2Temperature = 0.3
)

// GenerationCacheConfig holds configuration for the generation cache
type GenerationCacheConfig struct {
	UseRedis bool
	TTL      time.Duration
}

type GenerationCache interface {
	Get(ctx context.Context, req *GenerationRequest) (*CachedGeneration, bool)
	Put(ctx context.Context, req *GenerationRequest, value CachedGeneration)
	Invalidate(ctx context.Context, predicate func(CachedGeneration) bool)
}

type CachedGeneration struct {
	Code      string
	Model     string
	PromptKey string
	StoredAt  time.Time
	ExpiresAt time.Time
}

type ModelSelector interface {
	SelectModel(req *GenerationRequest) ModelSelection
}

type ModelSelection struct {
	Model      string
	Complexity int
	Reason     string
	Review     ManualReviewRecommendation
}

type ManualReviewRecommendation struct {
	Required bool
	Reason   string
	Tier     string
}

type InMemoryGenerationCache struct {
	entries map[string]CachedGeneration
	ttl     time.Duration
}

func NewInMemoryGenerationCache(ttl time.Duration) *InMemoryGenerationCache {
	if ttl <= 0 {
		ttl = defaultCacheTTL
	}
	return &InMemoryGenerationCache{entries: map[string]CachedGeneration{}, ttl: ttl}
}

func (c *InMemoryGenerationCache) Get(_ context.Context, req *GenerationRequest) (*CachedGeneration, bool) {
	if c == nil {
		return nil, false
	}
	key := cacheKey(req)
	entry, ok := c.entries[key]
	if !ok || time.Now().UTC().After(entry.ExpiresAt) {
		return nil, false
	}
	copy := entry
	return &copy, true
}

func (c *InMemoryGenerationCache) Put(_ context.Context, req *GenerationRequest, value CachedGeneration) {
	if c == nil {
		return
	}
	value.StoredAt = time.Now().UTC()
	value.ExpiresAt = value.StoredAt.Add(c.ttl)
	value.PromptKey = cacheKey(req)
	c.entries[value.PromptKey] = value
}

func (c *InMemoryGenerationCache) Invalidate(_ context.Context, predicate func(CachedGeneration) bool) {
	if c == nil || predicate == nil {
		return
	}
	for key, entry := range c.entries {
		if predicate(entry) {
			delete(c.entries, key)
		}
	}
}

type HeuristicModelSelector struct{}

func (HeuristicModelSelector) SelectModel(req *GenerationRequest) ModelSelection {
	complexity := ScoreComplexity(req)
	selection := ModelSelection{
		Complexity: complexity,
		Model:      defaultSimpleModel,
		Reason:     "simple function request routed to low-cost model",
		Review:     ManualReviewRecommendation{Tier: "auto"},
	}
	if complexity >= 7 {
		selection.Model = defaultComplexModel
		selection.Reason = "complex function request routed to higher capability model"
		selection.Review = ManualReviewRecommendation{
			Required: true,
			Reason:   "complexity threshold exceeded",
			Tier:     "human_approval",
		}
	}
	return selection
}

func ScoreComplexity(req *GenerationRequest) int {
	if req == nil {
		return 1
	}
	text := strings.ToLower(strings.Join([]string{req.Name, req.Description, req.Prompt, req.Category, strings.Join(req.Tags, " ")}, " "))
	complexity := 1
	if len(req.InputSchema) > 3 {
		complexity += 2
	}
	if len(req.OutputSchema) > 2 {
		complexity++
	}
	for _, token := range []string{"oauth", "auth", "stream", "webhook", "batch", "transform", "csv", "json", "retry", "cache", "security"} {
		if strings.Contains(text, token) {
			complexity++
		}
	}
	if complexity > 10 {
		return 10
	}
	return complexity
}

type OpenRouterClient struct {
	apiKey        string
	baseURL       string
	httpClient    *http.Client
	cache         GenerationCache
	modelSelector ModelSelector
}

func NewOpenRouterClient(apiKey string, httpClient *http.Client, cache GenerationCache, selector ModelSelector) *OpenRouterClient {
	if httpClient == nil {
		httpClient = &http.Client{Timeout: defaultGenerationTimeout}
	}
	if cache == nil {
		cache = NewInMemoryGenerationCache(defaultCacheTTL)
	}
	if selector == nil {
		selector = HeuristicModelSelector{}
	}
	return &OpenRouterClient{
		apiKey:        apiKey,
		baseURL:       defaultOpenRouterBaseURL,
		httpClient:    httpClient,
		cache:         cache,
		modelSelector: selector,
	}
}

// NewOpenRouterClientWithRedis creates a new OpenRouterClient with Redis-backed generation cache.
// If redisClient is nil or useRedis is false, falls back to in-memory cache.
func NewOpenRouterClientWithRedis(apiKey string, httpClient *http.Client, redisClient *redis.Client, useRedis bool, selector ModelSelector) *OpenRouterClient {
	ttl := defaultCacheTTL
	if ttlStr := os.Getenv("GENERATION_CACHE_TTL"); ttlStr != "" {
		if parsed := parseTTL(ttlStr); parsed > 0 {
			ttl = parsed
		}
	}

	cache := NewGenerationCache(redisClient, useRedis, ttl)
	return NewOpenRouterClient(apiKey, httpClient, cache, selector)
}

// parseTTL parses a TTL string (seconds or duration like "30m")
func parseTTL(s string) time.Duration {
	// Try parsing as seconds first
	if seconds, err := strconv.Atoi(s); err == nil && seconds > 0 {
		return time.Duration(seconds) * time.Second
	}
	// Try parsing as duration string
	if d, err := time.ParseDuration(s); err == nil && d > 0 {
		return d
	}
	return 0
}

func (c *OpenRouterClient) GenerateCode(ctx context.Context, req *GenerationRequest) (string, error) {
	if req == nil {
		return "", fmt.Errorf("generation request is required")
	}
	if cached, ok := c.cache.Get(ctx, req); ok {
		if req.Model == "" {
			req.Model = cached.Model
		}
		return cached.Code, nil
	}
	selection := c.modelSelector.SelectModel(req)
	if req.Model == "" {
		req.Model = selection.Model
	}
	if req.Prompt == "" {
		req.Prompt = BuildPrompt(req, selection)
	}
	if c.apiKey == "" {
		return generateFallbackCode(req), nil
	}
	payload := openRouterRequest{
		Model:       req.Model,
		Messages:    []openRouterMessage{{Role: "system", Content: "You are an expert code generator for FunctionFly serverless functions. Generate production-ready, optimized code with proper validation, error handling, and minimal dependencies. Return only the source code, no explanations. Mercury 2 is optimized for fast, accurate code generation."}, {Role: "user", Content: req.Prompt}},
		Temperature: deterministicTemperature(req.Deterministic),
	}
	body, err := json.Marshal(payload)
	if err != nil {
		return "", err
	}
	httpReq, err := http.NewRequestWithContext(ctx, http.MethodPost, c.baseURL+"/chat/completions", bytes.NewReader(body))
	if err != nil {
		return generateFallbackCode(req), nil
	}
	httpReq.Header.Set("Authorization", "Bearer "+c.apiKey)
	httpReq.Header.Set("Content-Type", "application/json")
	resp, err := c.httpClient.Do(httpReq)
	if err != nil {
		return generateFallbackCode(req), nil
	}
	defer resp.Body.Close()
	if resp.StatusCode >= 300 {
		return generateFallbackCode(req), nil
	}
	var parsed openRouterResponse
	if err := json.NewDecoder(resp.Body).Decode(&parsed); err != nil {
		return generateFallbackCode(req), nil
	}
	if len(parsed.Choices) == 0 {
		return generateFallbackCode(req), nil
	}
	code := strings.TrimSpace(stripCodeFences(parsed.Choices[0].Message.Content))
	c.cache.Put(ctx, req, CachedGeneration{Code: code, Model: req.Model})
	return code, nil
}

func BuildPrompt(req *GenerationRequest, selection ModelSelection) string {
	inputSchema, _ := json.MarshalIndent(req.InputSchema, "", "  ")
	outputSchema, _ := json.MarshalIndent(req.OutputSchema, "", "  ")
	return fmt.Sprintf("Function Name: %s\nDescription: %s\nCategory: %s\nRuntime: %s\nModel Strategy: %s\nComplexity: %d/10\nInput Schema: %s\nOutput Schema: %s\nRequirements:\n1. Production-ready code only\n2. Include validation and error handling\n3. Avoid unsafe dynamic execution\n4. Keep dependencies minimal\n5. Optimize cold start and latency\nPrompt Details: %s", req.Name, req.Description, req.Category, req.Runtime, selection.Reason, selection.Complexity, string(inputSchema), string(outputSchema), req.Prompt)
}

func generateFallbackCode(req *GenerationRequest) string {
	service := NewService(nil)
	return service.generateDefaultCode(req)
}

func deterministicTemperature(deterministic bool) float64 {
	if deterministic {
		return 0.1
	}
	// Mercury 2 optimized temperature for code generation
	return mercury2Temperature
}

func cacheKey(req *GenerationRequest) string {
	inputSchema, _ := json.Marshal(req.InputSchema)
	outputSchema, _ := json.Marshal(req.OutputSchema)
	return strings.Join([]string{req.Name, req.Description, req.Category, req.Runtime, req.Prompt, string(inputSchema), string(outputSchema)}, "|")
}

func stripCodeFences(content string) string {
	content = strings.TrimSpace(content)
	content = strings.TrimPrefix(content, "```python")
	content = strings.TrimPrefix(content, "```javascript")
	content = strings.TrimPrefix(content, "```")
	content = strings.TrimSuffix(content, "```")
	return strings.TrimSpace(content)
}

type openRouterRequest struct {
	Model       string              `json:"model"`
	Messages    []openRouterMessage `json:"messages"`
	Temperature float64             `json:"temperature,omitempty"`
}

type openRouterMessage struct {
	Role    string `json:"role"`
	Content string `json:"content"`
}

type openRouterResponse struct {
	Choices []struct {
		Message openRouterMessage `json:"message"`
	} `json:"choices"`
}

package cache

import (
	"os"
	"strconv"
	"time"
)

// CacheConfiguration holds comprehensive cache configuration
type CacheConfiguration struct {
	// Memory cache settings
	MemoryMaxMB int64 `json:"memory_max_mb"`

	// Disk cache settings
	DiskEnabled bool `json:"disk_enabled"`

	// Redis settings
	RedisEnabled     bool   `json:"redis_enabled"`
	RedisAddr        string `json:"redis_addr"`
	RedisPassword    string `json:"redis_password"`
	RedisDB          int    `json:"redis_db"`
	RedisRegistryTTL int    `json:"redis_registry_ttl"` // seconds

	// CDN settings
	CDNEnabled       bool   `json:"cdn_enabled"`
	CDNProvider      string `json:"cdn_provider"` // "cloudflare", "cloudfront", "fastly"
	CDNBaseURL       string `json:"cdn_base_url"`
	CDNMaxAge        int    `json:"cdn_max_age"` // seconds
	SDKBasePath      string `json:"sdk_base_path"`
	DocsBasePath     string `json:"docs_base_path"`
	StaticBasePath   string `json:"static_base_path"`
	// Cloudflare cache purge
	CloudflareZoneID string `json:"cloudflare_zone_id"`
	CloudflareToken  string `json:"cloudflare_token"`

	// Edge caching settings
	EdgeCacheEnabled         bool          `json:"edge_cache_enabled"`
	EdgeMinPopularityScore   int           `json:"edge_min_popularity_score"`
	EdgeMinExecutionCount    int           `json:"edge_min_execution_count"`
	EdgeMinTrustScore        float64       `json:"edge_min_trust_score"`
	EdgeMinSuccessRate       float64       `json:"edge_min_success_rate"`
	EdgeMaxLatencyMs         int           `json:"edge_max_latency_ms"`
	EdgeCacheDuration        time.Duration `json:"edge_cache_duration"`
	EdgeMaxFunctions         int           `json:"edge_max_functions"`
	EdgeRefreshInterval      time.Duration `json:"edge_refresh_interval"`

	// General settings
	DefaultTTL int `json:"default_ttl"` // seconds
}

// LoadCacheConfiguration loads cache configuration from environment variables
func LoadCacheConfiguration() *CacheConfiguration {
	config := &CacheConfiguration{
		// Memory cache defaults
		MemoryMaxMB: getEnvInt64("CACHE_MEMORY_MAX_MB", 100),

		// Disk cache defaults
		DiskEnabled: getEnvBool("CACHE_DISK_ENABLED", false),

		// Redis defaults
		RedisEnabled:     getEnvBool("CACHE_REDIS_ENABLED", false),
		RedisAddr:        getEnvString("REDIS_ADDR", "localhost:6379"),
		RedisPassword:    getEnvString("REDIS_PASSWORD", ""),
		RedisDB:          getEnvInt("REDIS_DB", 0),
		RedisRegistryTTL: getEnvInt("CACHE_REDIS_REGISTRY_TTL", 600), // 10 minutes

		// CDN defaults
		CDNEnabled:       getEnvBool("CACHE_CDN_ENABLED", false),
		CDNProvider:     getEnvString("CACHE_CDN_PROVIDER", "cloudflare"),
		CDNBaseURL:      getEnvString("CACHE_CDN_BASE_URL", "https://cdn.functionfly.com"),
		CDNMaxAge:       getEnvInt("CACHE_CDN_MAX_AGE", 86400), // 24 hours
		SDKBasePath:     getEnvString("CACHE_SDK_BASE_PATH", "/sdk"),
		DocsBasePath:     getEnvString("CACHE_DOCS_BASE_PATH", "/docs"),
		StaticBasePath:  getEnvString("CACHE_STATIC_BASE_PATH", "/static"),
		CloudflareZoneID: getEnvString("CLOUDFLARE_ZONE_ID", ""),
		CloudflareToken:  getEnvString("CLOUDFLARE_API_TOKEN", ""),

		// Edge caching defaults
		EdgeCacheEnabled:       getEnvBool("CACHE_EDGE_ENABLED", false),
		EdgeMinPopularityScore: getEnvInt("CACHE_EDGE_MIN_POPULARITY", 50),
		EdgeMinExecutionCount:  getEnvInt("CACHE_EDGE_MIN_EXECUTIONS", 100),
		EdgeMinTrustScore:      getEnvFloat64("CACHE_EDGE_MIN_TRUST_SCORE", 70.0),
		EdgeMinSuccessRate:     getEnvFloat64("CACHE_EDGE_MIN_SUCCESS_RATE", 95.0),
		EdgeMaxLatencyMs:       getEnvInt("CACHE_EDGE_MAX_LATENCY_MS", 5000),
		EdgeCacheDuration:      getEnvDuration("CACHE_EDGE_DURATION", 1*time.Hour),
		EdgeMaxFunctions:       getEnvInt("CACHE_EDGE_MAX_FUNCTIONS", 100),
		EdgeRefreshInterval:    getEnvDuration("CACHE_EDGE_REFRESH_INTERVAL", 10*time.Minute),

		// General defaults
		DefaultTTL: getEnvInt("CACHE_DEFAULT_TTL", 3600), // 1 hour
	}

	return config
}

// ToCacheConfig converts to the existing CacheConfig format
func (c *CacheConfiguration) ToCacheConfig() *CacheConfig {
	return &CacheConfig{
		MaxMemoryMB:      c.MemoryMaxMB,
		EnableDiskCache:  c.DiskEnabled,
		EnableRedisCache: c.RedisEnabled,
		EnableCDNCaching: c.CDNEnabled,
		DefaultTTL:       c.DefaultTTL,
		RedisRegistryTTL: c.RedisRegistryTTL,
	}
}

// ToCDNConfig converts to CDNConfig
func (c *CacheConfiguration) ToCDNConfig() *CDNConfig {
	return &CDNConfig{
		EnableCDNCaching: c.CDNEnabled,
		CDNBaseURL:       c.CDNBaseURL,
		CDNMaxAge:        c.CDNMaxAge,
		CloudflareZoneID: c.CloudflareZoneID,
		CloudflareToken:  c.CloudflareToken,
	}
}

// ToEdgeCacheConfig converts to EdgeCacheConfig
func (c *CacheConfiguration) ToEdgeCacheConfig() *EdgeCacheConfig {
	return &EdgeCacheConfig{
		Enabled:            c.EdgeCacheEnabled,
		MinPopularityScore: c.EdgeMinPopularityScore,
		MinExecutionCount:  c.EdgeMinExecutionCount,
		MinTrustScore:      c.EdgeMinTrustScore,
		MinSuccessRate:     c.EdgeMinSuccessRate,
		MaxLatencyMs:       c.EdgeMaxLatencyMs,
		CacheDuration:      c.EdgeCacheDuration,
		MaxEdgeFunctions:   c.EdgeMaxFunctions,
		RefreshInterval:    c.EdgeRefreshInterval,
	}
}

// Validate validates the cache configuration
func (c *CacheConfiguration) Validate() error {
	// Validate memory settings
	if c.MemoryMaxMB <= 0 {
		return &ConfigurationError{Field: "memory_max_mb", Message: "must be greater than 0"}
	}

	// Validate Redis settings
	if c.RedisEnabled {
		if c.RedisAddr == "" {
			return &ConfigurationError{Field: "redis_addr", Message: "cannot be empty when Redis is enabled"}
		}
		if c.RedisRegistryTTL <= 0 {
			return &ConfigurationError{Field: "redis_registry_ttl", Message: "must be greater than 0"}
		}
	}

	// Validate CDN settings
	if c.CDNEnabled {
		if c.CDNBaseURL == "" {
			return &ConfigurationError{Field: "cdn_base_url", Message: "cannot be empty when CDN is enabled"}
		}
		validProviders := map[string]bool{"cloudflare": true, "cloudfront": true, "fastly": true}
		if !validProviders[c.CDNProvider] {
			return &ConfigurationError{Field: "cdn_provider", Message: "must be one of: cloudflare, cloudfront, fastly"}
		}
	}

	// Validate edge cache settings
	if c.EdgeCacheEnabled {
		if c.EdgeMinPopularityScore < 0 {
			return &ConfigurationError{Field: "edge_min_popularity_score", Message: "cannot be negative"}
		}
		if c.EdgeMinExecutionCount < 0 {
			return &ConfigurationError{Field: "edge_min_execution_count", Message: "cannot be negative"}
		}
		if c.EdgeMinTrustScore < 0 || c.EdgeMinTrustScore > 100 {
			return &ConfigurationError{Field: "edge_min_trust_score", Message: "must be between 0 and 100"}
		}
		if c.EdgeMinSuccessRate < 0 || c.EdgeMinSuccessRate > 100 {
			return &ConfigurationError{Field: "edge_min_success_rate", Message: "must be between 0 and 100"}
		}
		if c.EdgeMaxLatencyMs <= 0 {
			return &ConfigurationError{Field: "edge_max_latency_ms", Message: "must be greater than 0"}
		}
		if c.EdgeMaxFunctions <= 0 {
			return &ConfigurationError{Field: "edge_max_functions", Message: "must be greater than 0"}
		}
	}

	return nil
}

// ConfigurationError represents a configuration validation error
type ConfigurationError struct {
	Field   string
	Message string
}

func (e *ConfigurationError) Error() string {
	return "cache configuration error for field '" + e.Field + "': " + e.Message
}

// Helper functions for environment variable parsing
func getEnvString(key, defaultValue string) string {
	if value := os.Getenv(key); value != "" {
		return value
	}
	return defaultValue
}

func getEnvInt(key string, defaultValue int) int {
	if value := os.Getenv(key); value != "" {
		if parsed, err := strconv.Atoi(value); err == nil {
			return parsed
		}
	}
	return defaultValue
}

func getEnvInt64(key string, defaultValue int64) int64 {
	if value := os.Getenv(key); value != "" {
		if parsed, err := strconv.ParseInt(value, 10, 64); err == nil {
			return parsed
		}
	}
	return defaultValue
}

func getEnvBool(key string, defaultValue bool) bool {
	if value := os.Getenv(key); value != "" {
		if parsed, err := strconv.ParseBool(value); err == nil {
			return parsed
		}
	}
	return defaultValue
}

func getEnvFloat64(key string, defaultValue float64) float64 {
	if value := os.Getenv(key); value != "" {
		if parsed, err := strconv.ParseFloat(value, 64); err == nil {
			return parsed
		}
	}
	return defaultValue
}

func getEnvDuration(key string, defaultValue time.Duration) time.Duration {
	if value := os.Getenv(key); value != "" {
		if parsed, err := time.ParseDuration(value); err == nil {
			return parsed
		}
	}
	return defaultValue
}

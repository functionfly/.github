package browser

import (
	"os"
	"strconv"
	"time"
)

// Config holds browser automation configuration.
type Config struct {
	// Enabled controls whether browser automation is enabled.
	Enabled bool

	// PoolSize is the number of shared browsers in the pool.
	PoolSize int

	// SessionTTL is how long a browser session lives in seconds.
	SessionTTL time.Duration

	// DefaultTimeout is the default navigation/action timeout in milliseconds.
	DefaultTimeoutMs int

	// AllowedDomains is a list of domain globs allowed for all agents.
	// Can be overridden per-agent in the database config.
	AllowedDomains []string

	// VaultEnabled enables encrypted credential storage.
	VaultEnabled bool

	// HeadfulMode runs browsers in headful (visible) mode for Premium tier.
	HeadfulMode bool

	// CDPSocketDir is the directory for CDP Unix sockets.
	CDPSocketDir string

	// DashboardPort is the port for the agent-browser observability dashboard.
	DashboardPort int

	// CostPerMinute is the cost in USD per browser-minute for billing.
	CostPerMinute float64

	// MaxRetries is the maximum number of retry attempts for a failed browser action.
	MaxRetries int

	// RetryBackoff is the backoff duration between retries.
	RetryBackoff time.Duration
}

// DefaultConfig returns the default browser configuration.
func DefaultConfig() Config {
	poolSize := 10
	if v := os.Getenv("BROWSER_POOL_SIZE"); v != "" {
		if n, err := strconv.Atoi(v); err == nil && n > 0 {
			poolSize = n
		}
	}

	ttlSeconds := 3600
	if v := os.Getenv("BROWSER_SESSION_TTL"); v != "" {
		if n, err := strconv.Atoi(v); err == nil && n > 0 {
			ttlSeconds = n
		}
	}

	timeoutMs := 30000
	if v := os.Getenv("BROWSER_DEFAULT_TIMEOUT_MS"); v != "" {
		if n, err := strconv.Atoi(v); err == nil && n > 0 {
			timeoutMs = n
		}
	}

	dashboardPort := 4848
	if v := os.Getenv("BROWSER_DASHBOARD_PORT"); v != "" {
		if n, err := strconv.Atoi(v); err == nil && n > 0 {
			dashboardPort = n
		}
	}

	socketDir := "/tmp/cdp-sockets"
	if v := os.Getenv("BROWSER_CDP_SOCKET_DIR"); v != "" {
		socketDir = v
	}

	costPerMinute := 0.01
	if v := os.Getenv("BROWSER_COST_PER_MINUTE"); v != "" {
		if f, err := strconv.ParseFloat(v, 64); err == nil && f >= 0 {
			costPerMinute = f
		}
	}

	maxRetries := 3
	if v := os.Getenv("BROWSER_MAX_RETRIES"); v != "" {
		if n, err := strconv.Atoi(v); err == nil && n >= 0 {
			maxRetries = n
		}
	}

	return Config{
		Enabled:          os.Getenv("BROWSER_ENABLED") == "true",
		PoolSize:          poolSize,
		SessionTTL:        time.Duration(ttlSeconds) * time.Second,
		DefaultTimeoutMs: timeoutMs,
		AllowedDomains:    parseEnvList("BROWSER_ALLOWED_DOMAINS"),
		VaultEnabled:     os.Getenv("BROWSER_VAULT_ENABLED") == "true",
		HeadfulMode:       false, // Default to headless for Standard tier
		CDPSocketDir:      socketDir,
		DashboardPort:    dashboardPort,
		CostPerMinute:     costPerMinute,
		MaxRetries:        maxRetries,
		RetryBackoff:      500 * time.Millisecond,
	}
}

// parseEnvList parses a comma-separated environment variable into a slice.
func parseEnvList(key string) []string {
	val := os.Getenv(key)
	if val == "" {
		return nil
	}
	var result []string
	var current []byte
	inQuote := false
	for i := 0; i < len(val); i++ {
		c := val[i]
		if c == '"' {
			inQuote = !inQuote
		} else if c == ',' && !inQuote {
			if len(current) > 0 {
				result = append(result, string(current))
				current = nil
			}
		} else {
			current = append(current, c)
		}
	}
	if len(current) > 0 {
		result = append(result, string(current))
	}
	return result
}

// AgentTier represents the agent's browser tier.
type AgentTier string

const (
	// TierStandard is the standard tier with shared pool.
	TierStandard AgentTier = "standard"
	// TierPremium is the premium tier with isolated browsers.
	TierPremium AgentTier = "premium"
)

// AgentConfig holds per-agent browser configuration.
type AgentConfig struct {
	AgentID                  string
	BrowserEnabled           bool
	AllowedDomains           []string
	MaxSessions              int
	CredentialStorageEnabled bool
	DefaultTimeoutMs         int
	HeadfulMode              bool
	Tier                     AgentTier
}

// GetTimeout returns the effective timeout for an agent.
func (c *AgentConfig) GetTimeout() time.Duration {
	if c.DefaultTimeoutMs > 0 {
		return time.Duration(c.DefaultTimeoutMs) * time.Millisecond
	}
	return 30 * time.Second
}

// IsDomainAllowed checks if a domain is allowed for this agent.
func (c *AgentConfig) IsDomainAllowed(domain string) bool {
	if len(c.AllowedDomains) == 0 {
		return true // No restrictions
	}
	for _, pattern := range c.AllowedDomains {
		if matchGlob(pattern, domain) {
			return true
		}
	}
	return false
}

// matchGlob matches a domain glob pattern.
func matchGlob(pattern, domain string) bool {
	// Simple glob matching: *.example.com matches anything.example.com
	if len(pattern) >= 2 && pattern[0] == '*' && pattern[1] == '.' {
		suffix := pattern[2:]
		return len(domain) > len(suffix) && domain[len(domain)-len(suffix):] == suffix
	}
	return pattern == domain
}

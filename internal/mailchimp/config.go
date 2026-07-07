package mailchimp

import (
	"os"
	"strings"
)

type Config struct {
	APIKey        string
	ServerPrefix  string
	DefaultListID string
	WebhookSecret string
	SyncEnabled   bool
}

func NewConfig() *Config {
	apiKey := strings.TrimSpace(os.Getenv("MAILCHIMP_API_KEY"))
	serverPrefix := strings.TrimSpace(os.Getenv("MAILCHIMP_SERVER_PREFIX"))
	if serverPrefix == "" && apiKey != "" {
		parts := strings.Split(apiKey, "-")
		if len(parts) > 1 {
			serverPrefix = parts[len(parts)-1]
		}
	}

	syncEnabled := true
	if val := strings.TrimSpace(os.Getenv("MAILCHIMP_SYNC_ENABLED")); val == "false" {
		syncEnabled = false
	}

	return &Config{
		APIKey:        apiKey,
		ServerPrefix:  serverPrefix,
		DefaultListID: strings.TrimSpace(os.Getenv("MAILCHIMP_DEFAULT_LIST_ID")),
		WebhookSecret: strings.TrimSpace(os.Getenv("MAILCHIMP_WEBHOOK_SECRET")),
		SyncEnabled:   syncEnabled,
	}
}

func (c *Config) IsConfigured() bool {
	return c.APIKey != "" && c.ServerPrefix != "" && c.DefaultListID != ""
}

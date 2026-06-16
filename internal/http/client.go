package http

import (
	"net/http"
	"os"
	"time"
)

const (
	TimeoutShort        = 10 * time.Second
	TimeoutDefault      = 30 * time.Second
	TimeoutLong         = 60 * time.Second
	TimeoutVeryLong      = 120 * time.Second
	TimeoutExtraLong    = 300 * time.Second
	TimeoutAI           = 120 * time.Second
	TimeoutPython       = 300 * time.Second
	TimeoutNodeJS       = 60 * time.Second
	TimeoutSandbox      = 30 * time.Second
	TimeoutProvider     = 10 * time.Second
	TimeoutWebhook      = 30 * time.Second
	TimeoutBilling      = 30 * time.Second
	TimeoutBrowser      = 30 * time.Second
	TimeoutConsciousness = 10 * time.Second
)

type ClientOption func(*http.Client)

func WithTransport(transport *http.Transport) ClientOption {
	return func(c *http.Client) {
		c.Transport = transport
	}
}

func NewClient(timeout time.Duration, opts ...ClientOption) *http.Client {
	c := &http.Client{
		Timeout: timeout,
	}
	for _, opt := range opts {
		opt(c)
	}
	return c
}

func NewDefaultClient() *http.Client {
	return NewClient(TimeoutDefault)
}

func NewShortTimeoutClient() *http.Client {
	return NewClient(TimeoutShort)
}

func NewLongTimeoutClient() *http.Client {
	return NewClient(TimeoutLong)
}

func NewAIClient() *http.Client {
	return NewClient(TimeoutAI)
}

func getEnvDuration(key string, defaultVal time.Duration) time.Duration {
	if val := os.Getenv(key); val != "" {
		if duration, err := time.ParseDuration(val); err == nil && duration > 0 {
			return duration
		}
	}
	return defaultVal
}

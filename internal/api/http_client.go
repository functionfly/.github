package api

import (
	"net/http"
	"sync"
	"time"
)

var (
	defaultHTTPClient     *http.Client
	defaultHTTPClientOnce sync.Once
)

// GetDefaultHTTPClient returns a shared HTTP client with connection pooling.
// This prevents creating new connections for each HTTP request.
func GetDefaultHTTPClient() *http.Client {
	defaultHTTPClientOnce.Do(func() {
		defaultHTTPClient = &http.Client{
			Timeout: 30 * time.Second,
			Transport: &http.Transport{
				MaxIdleConns:        100,
				MaxIdleConnsPerHost: 10,
				IdleConnTimeout:     90 * time.Second,
			},
		}
	})
	return defaultHTTPClient
}

// HTTPClientWithTimeout returns an HTTP client with a custom timeout.
// Use this for requests that need different timeouts than the default.
func HTTPClientWithTimeout(timeout time.Duration) *http.Client {
	return &http.Client{
		Timeout: timeout,
		Transport: &http.Transport{
			MaxIdleConns:        50,
			MaxIdleConnsPerHost: 5,
			IdleConnTimeout:     60 * time.Second,
		},
	}
}

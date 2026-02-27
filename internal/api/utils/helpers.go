package utils

import (
	"net"
	"net/http"
	"strconv"

	"github.com/functionfly/functionfly/internal/adapters/cloudflare"
	"github.com/functionfly/functionfly/internal/adapters/common"
	"github.com/functionfly/functionfly/internal/adapters/deno"
	"github.com/functionfly/functionfly/internal/adapters/fly"
	"github.com/functionfly/functionfly/internal/adapters/functionfly"
	"github.com/functionfly/functionfly/internal/adapters/vercel"
)

// GetAdapterForProvider returns the appropriate adapter for a provider
func GetAdapterForProvider(provider string) common.DeploymentAdapter {
	switch provider {
	case "workers":
		return cloudflare.NewCloudflareAdapter()
	case "vercel":
		return vercel.NewVercelAdapter()
	case "fly":
		return fly.NewFlyDeploymentAdapter()
	case "deno-deploy":
		return deno.NewDenoAdapter()
	case "functionfly-edge":
		return functionfly.NewFunctionFlyAdapter()
	default:
		return nil
	}
}

// GetProviderAdapter returns the appropriate provider adapter (non-deployment)
func GetProviderAdapter(provider string) common.ProviderAdapter {
	switch provider {
	case "workers":
		return cloudflare.NewCloudflareAdapter()
	case "vercel":
		return vercel.NewVercelAdapter()
	case "fly":
		return fly.NewFlyAdapter()
	case "deno-deploy":
		return deno.NewDenoAdapter()
	case "functionfly-edge":
		return functionfly.NewFunctionFlyAdapter()
	default:
		return nil
	}
}

// GetClientIP extracts the real client IP address from the request
func GetClientIP(r *http.Request) string {
	// Check for X-Forwarded-For header (set by proxies/load balancers)
	if xff := r.Header.Get("X-Forwarded-For"); xff != "" {
		// X-Forwarded-For can contain multiple IPs, take the first one
		if ips := net.ParseIP(xff); ips != nil {
			return ips.String()
		}
		// If it's a comma-separated list, take the first IP
		if ip, _, err := net.SplitHostPort(xff); err == nil {
			if net.ParseIP(ip) != nil {
				return ip
			}
		}
	}

	// Check for X-Real-IP header (set by nginx)
	if xri := r.Header.Get("X-Real-IP"); xri != "" {
		if net.ParseIP(xri) != nil {
			return xri
		}
	}

	// Fall back to RemoteAddr
	ip, _, err := net.SplitHostPort(r.RemoteAddr)
	if err != nil {
		return r.RemoteAddr
	}
	return ip
}

// ParseInt safely parses a string to int with min/max bounds
func ParseInt(s string, min, max int) (int, error) {
	val, err := strconv.Atoi(s)
	if err != nil {
		return 0, err
	}
	if val < min {
		return min, nil
	}
	if val > max {
		return max, nil
	}
	return val, nil
}

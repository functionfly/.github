package payment

import (
	"os"
	"strconv"
)

// RegistryPublisherRevenueShareBPS returns basis points (0–10000) of measured MicroVM execution
// billing credited to the function owner's payout ledger. Override with
// REGISTRY_PUBLISHER_REVENUE_SHARE_BPS (e.g. 2500 = 25%). Default 2500.
func RegistryPublisherRevenueShareBPS() int {
	const defaultBPS = 2500
	raw := os.Getenv("REGISTRY_PUBLISHER_REVENUE_SHARE_BPS")
	if raw == "" {
		return defaultBPS
	}
	v, err := strconv.Atoi(raw)
	if err != nil || v < 0 || v > 10000 {
		return defaultBPS
	}
	return v
}

// RegistryPublisherFallbackShareCents returns extra cents per successful registry execution when
// MicroVM billing cents are zero (e.g. non-MicroVM runtimes). Default 0. Optional env
// REGISTRY_PUBLISHER_FALLBACK_SHARE_CENTS for local/staging (cap 1000).
func RegistryPublisherFallbackShareCents() int {
	raw := os.Getenv("REGISTRY_PUBLISHER_FALLBACK_SHARE_CENTS")
	if raw == "" {
		return 0
	}
	v, err := strconv.Atoi(raw)
	if err != nil || v < 0 || v > 1000 {
		return 0
	}
	return v
}

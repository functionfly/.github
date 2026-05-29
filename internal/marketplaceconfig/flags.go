package marketplaceconfig

import (
	"os"
	"strings"
)

// PurchasesEnabled returns whether marketplace purchases UI/API are enabled.
// Set MARKETPLACE_PURCHASES_ENABLED=false to disable (staging rollouts).
func PurchasesEnabled() bool {
	v := strings.TrimSpace(os.Getenv("MARKETPLACE_PURCHASES_ENABLED"))
	if v == "" {
		return true
	}
	switch strings.ToLower(v) {
	case "0", "false", "no", "off":
		return false
	default:
		return true
	}
}

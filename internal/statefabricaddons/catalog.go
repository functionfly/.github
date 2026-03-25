package statefabricaddons

import (
	"os"
	"strings"
)

// AddOn is a sellable State Fabric add-on (catalog). IDs are stable for entitlements + Stripe mapping.
type AddOn struct {
	ID            string `json:"id"`
	Name          string `json:"name"`
	Price         string `json:"price"`
	Period        string `json:"period"`
	Description   string `json:"description"`
	StripePriceID string `json:"stripe_price_id,omitempty"`
}

func envPrice(key string) string {
	return strings.TrimSpace(os.Getenv(key))
}

// Catalog returns the canonical add-on list for API and billing.
// Stripe price IDs come from environment when set (e.g. STRIPE_PRICE_SF_ADDON_HOT_CACHE).
func Catalog() []AddOn {
	return []AddOn{
		{
			ID:            "hot_cache_booster",
			Name:          "Hot Cache Booster",
			Price:         "$49",
			Period:        "/mo per 5GB",
			Description:   "Reduces replay and read costs",
			StripePriceID: envPrice("STRIPE_PRICE_SF_ADDON_HOT_CACHE"),
		},
		{
			ID:            "advanced_security_pack",
			Name:          "Advanced Security Pack",
			Price:         "$99",
			Period:        "/mo",
			Description:   "SOC2-friendly logs, key rotation, audit streams",
			StripePriceID: envPrice("STRIPE_PRICE_SF_ADDON_SECURITY"),
		},
		{
			ID:            "ai_memory_pack",
			Name:          "AI Memory Pack",
			Price:         "$149",
			Period:        "/mo",
			Description:   "Vector index, embeddings storage, fast read engine",
			StripePriceID: envPrice("STRIPE_PRICE_SF_ADDON_AI_MEMORY"),
		},
		{
			ID:            "advanced_insights",
			Name:          "Advanced Insights",
			Price:         "$79",
			Period:        "/mo",
			Description:   "Cost forecasting, anomaly detection, hot path alerts",
			StripePriceID: envPrice("STRIPE_PRICE_SF_ADDON_INSIGHTS"),
		},
	}
}

// GetByID returns a catalog add-on by stable ID.
func GetByID(id string) (AddOn, bool) {
	for _, addon := range Catalog() {
		if addon.ID == id {
			return addon, true
		}
	}
	return AddOn{}, false
}

package billing

import (
	"time"

	"github.com/google/uuid"
)

// BundleResponse represents a Backend-in-a-Box pricing bundle
type BundleResponse struct {
	ID                uuid.UUID      `json:"id"`
	Slug              string         `json:"slug"`
	Name              string         `json:"name"`
	DisplayName       string         `json:"display_name"`
	Description       string         `json:"description"`
	ShortDescription  string         `json:"short_description"`
	PriceCents        int            `json:"price_cents"`
	PriceUSD          string         `json:"price_usd"`
	BillingInterval   string         `json:"billing_interval"`
	Icon              string         `json:"icon"`
	Color             string         `json:"color"`
	FeaturesIncluded  []string       `json:"features_included"`
	FeatureLimits     map[string]int `json:"feature_limits"`
	ProvisioningSteps []string       `json:"provisioning_steps"`
	IsPopular         bool           `json:"is_popular"`
	SortOrder         int            `json:"sort_order"`
}

// FounderModeResponse represents a founder mode registration status
type FounderModeResponse struct {
	ID                 uuid.UUID  `json:"id"`
	BundleID           uuid.UUID  `json:"bundle_id"`
	BundleSlug         string     `json:"bundle_slug"`
	ModeType           string     `json:"mode_type"`
	Status             string     `json:"status"`
	StartedAt          time.Time  `json:"started_at"`
	EndsAt             *time.Time `json:"ends_at,omitempty"`
	FreeDays           int        `json:"free_days"`
	MRRThresholdCents  int        `json:"mrr_threshold_cents"`
	DaysRemaining      int        `json:"days_remaining"`
	GracePeriodStarted *time.Time `json:"grace_period_started_at,omitempty"`
	GracePeriodEnds    *time.Time `json:"grace_period_ends_at,omitempty"`
}

// DeferredBillingStatus shows progress toward billing trigger thresholds
type DeferredBillingStatus struct {
	BundleID          uuid.UUID              `json:"bundle_id"`
	Status            string                 `json:"status"` // 'building', 'approaching', 'grace_period', 'converted'
	TriggerThresholds map[string]interface{} `json:"trigger_thresholds"`
	CurrentProgress   map[string]interface{} `json:"current_progress"`
	ProgressPercent   float64                `json:"progress_percent"` // Overall progress toward first trigger
	EstimatedDaysLeft *int                   `json:"estimated_days_left,omitempty"`
}

// ListBundlesRequest filters for listing bundles
type ListBundlesRequest struct {
	ActiveOnly bool `json:"active_only"`
}

// RegisterFounderModeRequest for signing up for founder mode
type RegisterFounderModeRequest struct {
	BundleSlug   string `json:"bundle_slug"`
	ModeType     string `json:"mode_type"`     // 'time_based', 'revenue_based', 'hybrid'
	FreeDays     int    `json:"free_days"`     // For time-based (default: 90)
	MRRThreshold int    `json:"mrr_threshold"` // For revenue-based in cents (default: 100000 = $1000)
	SuccessURL   string `json:"success_url"`
	CancelURL    string `json:"cancel_url"`
}

// CreateBundleCheckoutRequest for immediate bundle subscription
type CreateBundleCheckoutRequest struct {
	BundleSlug string `json:"bundle_slug"`
	SuccessURL string `json:"success_url"`
	CancelURL  string `json:"cancel_url"`
}

// ConvertToPaidRequest for manually converting from founder mode
type ConvertToPaidRequest struct {
	BundleID string `json:"bundle_id"`
}

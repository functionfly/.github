package webhooks

import (
	"encoding/json"
	"net/http"
	"time"

	studiohandler "github.com/functionfly/functionfly/internal/api/handlers/studio"
	"github.com/sirupsen/logrus"
	"github.com/stripe/stripe-go/v83"
	stripeSub "github.com/stripe/stripe-go/v83/subscription"
)

// SetMarketplaceRepository wires studio marketplace subscription sync.
func (h *StripeWebhookHandler) SetMarketplaceRepository(repo *studiohandler.MarketplaceRepository) {
	h.marketplaceRepo = repo
}

func (h *StripeWebhookHandler) handleMarketplacePlanCheckout(w http.ResponseWriter, r *http.Request, session *stripe.CheckoutSession) {
	if h.marketplaceRepo == nil {
		w.WriteHeader(http.StatusOK)
		_ = json.NewEncoder(w).Encode(map[string]string{"status": "ignored"})
		return
	}

	meta := session.Metadata
	creatorTenantID := meta["creator_tenant_id"]
	planID := meta["plan_id"]
	planName := meta["plan_name"]
	subscriberTenantID := meta["subscriber_tenant_id"]
	subscriberUserID := meta["subscriber_user_id"]
	subscriberName := meta["subscriber_name"]
	subscriberEmail := meta["subscriber_email"]
	billingCycle := meta["billing_cycle"]

	if creatorTenantID == "" || planName == "" {
		logrus.Warn("marketplace plan checkout missing metadata")
		http.Error(w, "Missing metadata", http.StatusBadRequest)
		return
	}
	if subscriberEmail == "" {
		subscriberEmail = session.CustomerEmail
	}
	if subscriberEmail == "" {
		logrus.Warn("marketplace plan checkout missing subscriber email")
		http.Error(w, "Missing subscriber email", http.StatusBadRequest)
		return
	}
	if subscriberName == "" {
		subscriberName = subscriberEmail
	}

	amount := float64(session.AmountTotal) / 100.0
	if amount <= 0 && session.AmountSubtotal > 0 {
		amount = float64(session.AmountSubtotal) / 100.0
	}

	periodStart := time.Now()
	periodEnd := periodStart.AddDate(0, 1, 0)
	if session.Subscription != nil && session.Subscription.ID != "" {
		if sub, err := stripeSub.Get(session.Subscription.ID, nil); err == nil && sub != nil && sub.Items != nil && len(sub.Items.Data) > 0 {
			item := sub.Items.Data[0]
			periodStart = time.Unix(item.CurrentPeriodStart, 0)
			periodEnd = time.Unix(item.CurrentPeriodEnd, 0)
		}
	}

	ctx := r.Context()
	if err := h.marketplaceRepo.UpsertPlanSubscriptionFromStripe(ctx, studiohandler.PlanSubscriptionSyncInput{
		CreatorTenantID:  creatorTenantID,
		PlanID:           planID,
		PlanName:         planName,
		StripeSubID:      subscriptionIDFromSession(session),
		SubscriberTenant: subscriberTenantID,
		SubscriberUser:   subscriberUserID,
		SubscriberName:   subscriberName,
		SubscriberEmail:  subscriberEmail,
		Status:           "active",
		Amount:           amount,
		BillingCycle:     billingCycle,
		PeriodStart:      periodStart,
		PeriodEnd:        periodEnd,
	}); err != nil {
		logrus.WithError(err).Error("marketplace plan checkout: failed to record subscription")
		http.Error(w, "Failed to record subscription", http.StatusInternalServerError)
		return
	}

	w.WriteHeader(http.StatusOK)
	_ = json.NewEncoder(w).Encode(map[string]string{"status": "success"})
}

func (h *StripeWebhookHandler) handleMarketplacePlanSubscriptionUpdated(w http.ResponseWriter, r *http.Request, sub *stripe.Subscription) {
	if h.marketplaceRepo == nil {
		w.WriteHeader(http.StatusOK)
		_ = json.NewEncoder(w).Encode(map[string]string{"status": "ignored"})
		return
	}

	meta := sub.Metadata
	periodStart, periodEnd := subscriptionPeriodBounds(sub)
	amount := subscriptionAmountUSD(sub)

	if err := h.marketplaceRepo.UpsertPlanSubscriptionFromStripe(r.Context(), studiohandler.PlanSubscriptionSyncInput{
		CreatorTenantID:   meta["creator_tenant_id"],
		PlanID:            meta["plan_id"],
		PlanName:          meta["plan_name"],
		StripeSubID:       sub.ID,
		SubscriberTenant:  meta["subscriber_tenant_id"],
		SubscriberUser:    meta["subscriber_user_id"],
		SubscriberName:    meta["subscriber_name"],
		SubscriberEmail:   meta["subscriber_email"],
		Status:            string(sub.Status),
		Amount:            amount,
		BillingCycle:      meta["billing_cycle"],
		PeriodStart:       periodStart,
		PeriodEnd:         periodEnd,
		CancelAtPeriodEnd: sub.CancelAtPeriodEnd,
	}); err != nil {
		logrus.WithError(err).WithField("subscription_id", sub.ID).Error("marketplace plan subscription update failed")
		http.Error(w, "Failed to update subscription", http.StatusInternalServerError)
		return
	}

	w.WriteHeader(http.StatusOK)
	_ = json.NewEncoder(w).Encode(map[string]string{"status": "success"})
}

func (h *StripeWebhookHandler) handleMarketplacePlanSubscriptionDeleted(w http.ResponseWriter, r *http.Request, sub *stripe.Subscription) {
	if h.marketplaceRepo == nil {
		w.WriteHeader(http.StatusOK)
		_ = json.NewEncoder(w).Encode(map[string]string{"status": "ignored"})
		return
	}

	meta := sub.Metadata
	periodStart, periodEnd := subscriptionPeriodBounds(sub)
	amount := subscriptionAmountUSD(sub)

	if err := h.marketplaceRepo.UpsertPlanSubscriptionFromStripe(r.Context(), studiohandler.PlanSubscriptionSyncInput{
		CreatorTenantID:  meta["creator_tenant_id"],
		PlanID:           meta["plan_id"],
		PlanName:         meta["plan_name"],
		StripeSubID:      sub.ID,
		SubscriberTenant: meta["subscriber_tenant_id"],
		SubscriberUser:   meta["subscriber_user_id"],
		SubscriberName:   meta["subscriber_name"],
		SubscriberEmail:  meta["subscriber_email"],
		Status:           "cancelled",
		Amount:           amount,
		BillingCycle:     meta["billing_cycle"],
		PeriodStart:      periodStart,
		PeriodEnd:        periodEnd,
	}); err != nil {
		logrus.WithError(err).WithField("subscription_id", sub.ID).Error("marketplace plan subscription delete sync failed")
		http.Error(w, "Failed to update subscription", http.StatusInternalServerError)
		return
	}

	w.WriteHeader(http.StatusOK)
	_ = json.NewEncoder(w).Encode(map[string]string{"status": "success"})
}

func subscriptionIDFromSession(session *stripe.CheckoutSession) string {
	if session.Subscription != nil {
		return session.Subscription.ID
	}
	return ""
}

func subscriptionPeriodBounds(sub *stripe.Subscription) (time.Time, time.Time) {
	start := time.Now()
	end := start.AddDate(0, 1, 0)
	if sub.Items != nil && len(sub.Items.Data) > 0 {
		item := sub.Items.Data[0]
		start = time.Unix(item.CurrentPeriodStart, 0)
		end = time.Unix(item.CurrentPeriodEnd, 0)
	}
	return start, end
}

func subscriptionAmountUSD(sub *stripe.Subscription) float64 {
	if sub.Items == nil || len(sub.Items.Data) == 0 || sub.Items.Data[0].Price == nil {
		return 0
	}
	price := sub.Items.Data[0].Price
	if price.UnitAmount > 0 {
		return float64(price.UnitAmount) / 100.0
	}
	return 0
}

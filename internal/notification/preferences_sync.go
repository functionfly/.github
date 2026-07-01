package notification

import (
	"context"
	"strings"

	"github.com/google/uuid"
)

// MandatoryCategories cannot be fully disabled by user preference (email/in-app still sent).
var MandatoryCategories = map[string]bool{
	CategorySecurity: true,
}

// userSettingsKeyForType maps notification types to user_settings JSON keys.
func userSettingsKeyForType(notifType string) string {
	switch notifType {
	case TypeDeploymentSuccess:
		return "deploymentSuccess"
	case TypeDeploymentFailed:
		return "deploymentFailure"
	case TypeDeploymentStarted:
		return "deploymentStarted"
	case TypeFailoverTriggered, TypeFailoverResolved:
		return "failoverEvents"
	case TypeProviderOffline, TypeProviderOnline, TypeProviderDegraded:
		return "providerIssues"
	case TypeBillingPaymentFailed, TypeBillingPaymentSuccess:
		return "paymentFailures"
	case TypeBillingWalletLowBalance:
		return "lowWalletBalance"
	case TypeBillingSpendCapWarning, TypeBillingForecastExceeded, TypeBillingUsageSpike:
		return "spendCapWarnings"
	case TypeBillingInvoiceGenerated:
		return "invoiceGenerated"
	case TypeBillingSubscriptionExpiring:
		return "subscriptionExpiring"
	case TypeSecurityNewDeviceLogin:
		return "newDeviceLogin"
	case TypeSecuritySuspiciousActivity:
		return "suspiciousActivity"
	case TypeFunctionError:
		return "functionErrors"
	case TypeFunctionPublished:
		return "functionPublished"
	case TypeFunctionUpdated:
		return "functionUpdated"
	case TypeTeamInvitation, TypeTeamInviteSent, TypeTeamInviteAccepted:
		return "teamInvitations"
	case TypeTeamRoleChanged:
		return "roleChanges"
	case TypeTeamDirectMessage:
		return "directMessages"
	case TypePayoutCompleted:
		return "payoutCompleted"
	case TypePayoutFailed, TypePayoutCancelled, TypePayoutReversed:
		return "payoutFailed"
	case TypePayoutApprovalNeeded:
		return "payoutApprovalNeeded"
	case TypeConsciousnessCritical:
		return "consciousnessCritical"
	case TypeConsciousnessAutoApplied:
		return "consciousnessAutoApplied"
	default:
		return ""
	}
}

// BoolFromSettings reads a boolean user_settings value with default.
func BoolFromSettings(settings map[string]interface{}, key string, defaultVal bool) bool {
	return boolFromSettings(settings, key, defaultVal)
}

func boolFromSettings(settings map[string]interface{}, key string, defaultVal bool) bool {
	if settings == nil {
		return defaultVal
	}
	v, ok := settings[key]
	if !ok {
		return defaultVal
	}
	b, ok := v.(bool)
	if !ok {
		return defaultVal
	}
	return b
}

// IsNotificationTypeEnabled checks user_settings toggles for operational notification types.
// Core security notifications (password, MFA, username changes) always return true.
// New device login and suspicious activity respect their user toggles.
func IsNotificationTypeEnabled(settings map[string]interface{}, notifType string) bool {
	if key := userSettingsKeyForType(notifType); key != "" {
		return boolFromSettings(settings, key, true)
	}
	if strings.HasPrefix(notifType, "security.") {
		return true
	}
	return true
}

// IsEmailEnabled checks the master email toggle in user settings.
func IsEmailEnabled(settings map[string]interface{}) bool {
	return boolFromSettings(settings, "emailNotifications", true)
}

// IsPushEnabled checks the master push toggle in user settings.
func IsPushEnabled(settings map[string]interface{}) bool {
	return boolFromSettings(settings, "pushNotifications", true)
}

// SyncPreferencesFromSettings mirrors user_settings notification fields into notification_preferences.
func SyncPreferencesFromSettings(ctx context.Context, repo Repository, userID uuid.UUID, settings map[string]interface{}) error {
	if settings == nil {
		settings = map[string]interface{}{}
	}

	if err := repo.CreateDefaultPreferences(ctx, userID); err != nil {
		return err
	}

	emailMaster := IsEmailEnabled(settings)
	inAppMaster := true // in-app has no global off switch; category toggles apply

	type categoryToggle struct {
		category string
		enabled  bool
	}

	pushMaster := IsPushEnabled(settings)
	categoryToggles := []categoryToggle{
		{CategoryDeployment, boolFromSettings(settings, "deploymentSuccess", true) || boolFromSettings(settings, "deploymentFailure", true) || boolFromSettings(settings, "deploymentStarted", false)},
		{CategoryFailover, boolFromSettings(settings, "failoverEvents", true)},
		{CategoryProvider, boolFromSettings(settings, "providerIssues", true)},
		{CategoryBilling, boolFromSettings(settings, "paymentFailures", true) || boolFromSettings(settings, "lowWalletBalance", true) || boolFromSettings(settings, "spendCapWarnings", true) || boolFromSettings(settings, "invoiceGenerated", false) || boolFromSettings(settings, "subscriptionExpiring", true)},
		{CategoryTeam, boolFromSettings(settings, "teamInvitations", true) || boolFromSettings(settings, "roleChanges", true) || boolFromSettings(settings, "directMessages", true)},
		{CategoryMessages, emailMaster},
		{CategorySystem, emailMaster},
		{CategoryFunction, boolFromSettings(settings, "functionErrors", true) || boolFromSettings(settings, "functionPublished", false) || boolFromSettings(settings, "functionUpdated", false)},
		{CategoryConsciousness, boolFromSettings(settings, "consciousnessCritical", true) || boolFromSettings(settings, "consciousnessAutoApplied", false)},
		{CategoryPayout, boolFromSettings(settings, "payoutCompleted", false) || boolFromSettings(settings, "payoutFailed", true) || boolFromSettings(settings, "payoutApprovalNeeded", true)},
	}

	for _, ct := range categoryToggles {
		for _, channel := range []string{ChannelEmail, ChannelInApp, ChannelPush} {
			enabled := ct.enabled
			if channel == ChannelEmail {
				enabled = enabled && emailMaster
			}
			if channel == ChannelInApp {
				enabled = enabled && inAppMaster
			}
			if channel == ChannelPush {
				enabled = enabled && pushMaster
			}
			if MandatoryCategories[ct.category] {
				enabled = true
			}
			if ct.category == CategoryDeployment {
				success := boolFromSettings(settings, "deploymentSuccess", true)
				failure := boolFromSettings(settings, "deploymentFailure", true)
				started := boolFromSettings(settings, "deploymentStarted", false)
				enabled = success || failure || started
				if channel == ChannelEmail {
					enabled = enabled && emailMaster
				}
				if channel == ChannelPush {
					enabled = enabled && pushMaster
				}
			}
			pref := &NotificationPreference{
				UserID:    userID,
				Channel:   channel,
				Category:  ct.category,
				Enabled:   enabled,
				Frequency: FrequencyImmediate,
				Timezone:  "UTC",
			}
			if err := repo.SavePreference(ctx, pref); err != nil {
				return err
			}
		}
	}

	// Security always on for email + in_app + push
	for _, channel := range []string{ChannelEmail, ChannelInApp, ChannelPush} {
		pref := &NotificationPreference{
			UserID:    userID,
			Channel:   channel,
			Category:  CategorySecurity,
			Enabled:   true,
			Frequency: FrequencyImmediate,
			Timezone:  "UTC",
		}
		if err := repo.SavePreference(ctx, pref); err != nil {
			return err
		}
	}

	return nil
}

// ShouldDeliverChannel evaluates channel + category + type against preferences and user settings.
func ShouldDeliverChannel(settings map[string]interface{}, pref *NotificationPreference, category, notifType, channel string) bool {
	if MandatoryCategories[category] {
		if channel == ChannelEmail && !IsEmailEnabled(settings) {
			return false
		}
		if channel == ChannelPush {
			if !IsPushEnabled(settings) {
				return false
			}
			if pref != nil && !pref.Enabled {
				return false
			}
			return true
		}
		return true
	}

	if channel == ChannelPush {
		if !IsPushEnabled(settings) {
			return false
		}
		if pref != nil && !pref.Enabled {
			return false
		}
		return true
	}

	if channel == ChannelEmail && !IsEmailEnabled(settings) {
		return false
	}

	if !IsNotificationTypeEnabled(settings, notifType) {
		return false
	}

	if pref != nil && !pref.Enabled {
		return false
	}

	if pref != nil && pref.Frequency != FrequencyImmediate && pref.Frequency != "" {
		return false // digest modes deferred post-launch
	}

	return true
}

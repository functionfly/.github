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
	case TypeFailoverTriggered, TypeFailoverResolved:
		return "failoverEvents"
	case TypeProviderOffline, TypeProviderOnline, TypeProviderDegraded:
		return "providerIssues"
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
// Security notifications always return true.
func IsNotificationTypeEnabled(settings map[string]interface{}, notifType string) bool {
	if strings.HasPrefix(notifType, "security.") {
		return true
	}
	if key := userSettingsKeyForType(notifType); key != "" {
		return boolFromSettings(settings, key, true)
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
		{CategoryDeployment, boolFromSettings(settings, "deploymentSuccess", true) && boolFromSettings(settings, "deploymentFailure", true)},
		{CategoryFailover, boolFromSettings(settings, "failoverEvents", true)},
		{CategoryProvider, boolFromSettings(settings, "providerIssues", true)},
		{CategoryBilling, emailMaster},
		{CategoryTeam, emailMaster},
		{CategoryMessages, emailMaster},
		{CategorySystem, emailMaster},
		{CategoryFunction, emailMaster},
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
				// deployment success/failure are type-level; keep category enabled if either is on
				success := boolFromSettings(settings, "deploymentSuccess", true)
				failure := boolFromSettings(settings, "deploymentFailure", true)
				enabled = success || failure
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

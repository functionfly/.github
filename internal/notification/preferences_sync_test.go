package notification

import (
	"context"
	"testing"
	"time"

	"github.com/google/uuid"
)

func TestIsNotificationTypeEnabled(t *testing.T) {
	settings := map[string]interface{}{
		"deploymentSuccess": false,
		"deploymentFailure": true,
		"failoverEvents":    false,
		"providerIssues":    true,
	}

	if IsNotificationTypeEnabled(settings, TypeDeploymentSuccess) {
		t.Fatal("expected deployment success disabled")
	}
	if !IsNotificationTypeEnabled(settings, TypeDeploymentFailed) {
		t.Fatal("expected deployment failure enabled")
	}
	if IsNotificationTypeEnabled(settings, TypeFailoverTriggered) {
		t.Fatal("expected failover disabled")
	}
	if !IsNotificationTypeEnabled(settings, TypeSecurityPasswordChanged) {
		t.Fatal("security notifications must always be enabled")
	}
}

func TestShouldDeliverChannelSecurityBypassesOptOut(t *testing.T) {
	settings := map[string]interface{}{"emailNotifications": false}
	pref := &NotificationPreference{Enabled: false}

	if ShouldDeliverChannel(settings, pref, CategorySecurity, TypeSecurityPasswordChanged, ChannelInApp) {
		t.Fatal("expected in-app security delivery even when pref disabled")
	}
	if ShouldDeliverChannel(settings, pref, CategorySecurity, TypeSecurityPasswordChanged, ChannelEmail) {
		t.Fatal("expected email security blocked when master email off")
	}
}

func TestShouldDeliverChannelRespectsMasterEmailToggle(t *testing.T) {
	settings := map[string]interface{}{"emailNotifications": false}
	pref := &NotificationPreference{Enabled: true}

	if ShouldDeliverChannel(settings, pref, CategoryBilling, TypeBillingAlert, ChannelEmail) {
		t.Fatal("expected email blocked when master toggle off")
	}
	if !ShouldDeliverChannel(settings, pref, CategoryBilling, TypeBillingAlert, ChannelInApp) {
		t.Fatal("expected in-app still allowed")
	}
}

func TestShouldDeliverChannelPushRespected(t *testing.T) {
	// push enabled, pref enabled → allowed
	settingsOn := map[string]interface{}{"pushNotifications": true}
	pref := &NotificationPreference{Enabled: true}
	if !ShouldDeliverChannel(settingsOn, pref, CategorySystem, TypeWelcome, ChannelPush) {
		t.Fatal("push should be delivered when enabled and pref enabled")
	}

	// push enabled, pref disabled → blocked
	prefOff := &NotificationPreference{Enabled: false}
	if ShouldDeliverChannel(settingsOn, prefOff, CategorySystem, TypeWelcome, ChannelPush) {
		t.Fatal("push should be blocked when pref disabled")
	}

	// push disabled globally → blocked regardless of pref
	settingsOff := map[string]interface{}{"pushNotifications": false}
	if ShouldDeliverChannel(settingsOff, pref, CategorySystem, TypeWelcome, ChannelPush) {
		t.Fatal("push should be blocked when master toggle off")
	}
}

func TestBoolFromSettings(t *testing.T) {
	settings := map[string]interface{}{"deploymentSuccess": false}
	if BoolFromSettings(settings, "deploymentSuccess", true) {
		t.Fatal("expected false")
	}
	if !BoolFromSettings(settings, "missing", true) {
		t.Fatal("expected default true")
	}
}

func TestSyncPreferencesFromSettingsCreatesSecurityOn(t *testing.T) {
	repo := &mockPreferenceRepo{prefs: map[string]*NotificationPreference{}}
	userID := uuid.New()
	settings := map[string]interface{}{
		"emailNotifications":  false,
		"deploymentSuccess":   true,
		"deploymentFailure":   false,
		"failoverEvents":      true,
		"providerIssues":      false,
	}

	if err := SyncPreferencesFromSettings(t.Context(), repo, userID, settings); err != nil {
		t.Fatalf("sync failed: %v", err)
	}

	secEmail, _ := repo.GetPreference(t.Context(), userID, ChannelEmail, CategorySecurity)
	if secEmail == nil || !secEmail.Enabled {
		t.Fatal("security email pref should remain enabled")
	}
}

type mockPreferenceRepo struct {
	prefs map[string]*NotificationPreference
}

func (m *mockPreferenceRepo) prefKey(userID uuid.UUID, channel, category string) string {
	return userID.String() + ":" + channel + ":" + category
}

func (m *mockPreferenceRepo) CreateNotification(ctx context.Context, n *Notification) error { return nil }
func (m *mockPreferenceRepo) GetNotification(ctx context.Context, id uuid.UUID) (*Notification, error) {
	return nil, nil
}
func (m *mockPreferenceRepo) ListNotifications(ctx context.Context, userID uuid.UUID, opts ListOptions) ([]*Notification, error) {
	return nil, nil
}
func (m *mockPreferenceRepo) GetUnreadCount(ctx context.Context, userID uuid.UUID) (int, error) {
	return 0, nil
}
func (m *mockPreferenceRepo) GetUnreadCountsByCategory(ctx context.Context, userID uuid.UUID) (map[string]int, error) {
	return nil, nil
}
func (m *mockPreferenceRepo) GetTotalCount(ctx context.Context, userID uuid.UUID) (int, error) {
	return 0, nil
}
func (m *mockPreferenceRepo) MarkAsRead(ctx context.Context, id uuid.UUID) error { return nil }
func (m *mockPreferenceRepo) MarkAllAsRead(ctx context.Context, userID uuid.UUID) error { return nil }
func (m *mockPreferenceRepo) DeleteNotification(ctx context.Context, id uuid.UUID) error { return nil }
func (m *mockPreferenceRepo) ArchiveNotification(ctx context.Context, id uuid.UUID) error { return nil }
func (m *mockPreferenceRepo) UpdateNotificationStatus(ctx context.Context, id uuid.UUID, status string) error {
	return nil
}
func (m *mockPreferenceRepo) ListPendingNotifications(ctx context.Context, limit int) ([]*Notification, error) {
	return nil, nil
}
func (m *mockPreferenceRepo) CleanupOldNotifications(ctx context.Context, olderThan time.Time) (int64, error) {
	return 0, nil
}
func (m *mockPreferenceRepo) ResetStaleProcessing(ctx context.Context, staleAfter time.Duration) (int64, error) {
	return 0, nil
}
func (m *mockPreferenceRepo) CleanupDeadLetterQueue(ctx context.Context, maxAge time.Duration) error {
	return nil
}
func (m *mockPreferenceRepo) CleanupExpiredNotifications(ctx context.Context, retentionDays int) (int64, error) {
	return 0, nil
}
func (m *mockPreferenceRepo) GetPreferences(ctx context.Context, userID uuid.UUID) ([]*NotificationPreference, error) {
	return nil, nil
}
func (m *mockPreferenceRepo) GetPreference(ctx context.Context, userID uuid.UUID, channel, category string) (*NotificationPreference, error) {
	if m.prefs == nil {
		return nil, nil
	}
	return m.prefs[m.prefKey(userID, channel, category)], nil
}
func (m *mockPreferenceRepo) SavePreference(ctx context.Context, p *NotificationPreference) error {
	if m.prefs == nil {
		m.prefs = map[string]*NotificationPreference{}
	}
	key := m.prefKey(p.UserID, p.Channel, p.Category)
	copy := *p
	m.prefs[key] = &copy
	return nil
}
func (m *mockPreferenceRepo) CreateDefaultPreferences(ctx context.Context, userID uuid.UUID) error {
	return nil
}
func (m *mockPreferenceRepo) CreateDeadLetter(ctx context.Context, n *Notification, failureReason string) error {
	return nil
}
func (m *mockPreferenceRepo) DeleteDeadLetter(ctx context.Context, id uuid.UUID) error {
	return nil
}
func (m *mockPreferenceRepo) ListDeadLetters(ctx context.Context, opts DeadLetterListOptions) ([]*DeadLetter, error) {
	return nil, nil
}
func (m *mockPreferenceRepo) GetDeadLetter(ctx context.Context, id uuid.UUID) (*DeadLetter, error) {
	return nil, nil
}
func (m *mockPreferenceRepo) RetryDeadLetter(ctx context.Context, id uuid.UUID) (*Notification, error) {
	return nil, nil
}
func (m *mockPreferenceRepo) GetPendingNotifications(ctx context.Context, olderThan time.Duration, limit int) ([]*Notification, error) {
	return nil, nil
}
func (m *mockPreferenceRepo) RequeueNotification(ctx context.Context, id uuid.UUID) error {
	return nil
}
func (m *mockPreferenceRepo) MarkAbandonedDeadLetters(ctx context.Context, maxRetries int) (int64, error) {
	return 0, nil
}
func (m *mockPreferenceRepo) MoveToDeadLetterQueue(ctx context.Context, notificationID uuid.UUID, failureReason string) error {
	return nil
}
func (m *mockPreferenceRepo) GetTemplate(ctx context.Context, notificationType, channel string) (*NotificationTemplate, error) {
	return nil, nil
}
func (m *mockPreferenceRepo) ListTemplates(ctx context.Context) ([]*NotificationTemplate, error) {
	return nil, nil
}
func (m *mockPreferenceRepo) SaveTemplate(ctx context.Context, t *NotificationTemplate) error { return nil }
func (m *mockPreferenceRepo) DeleteTemplate(ctx context.Context, id uuid.UUID) error { return nil }
func (m *mockPreferenceRepo) TrackAnalytics(ctx context.Context, a *NotificationAnalytics) error { return nil }
func (m *mockPreferenceRepo) GetAnalytics(ctx context.Context, notificationID uuid.UUID) ([]*NotificationAnalytics, error) {
	return nil, nil
}

package notification

import (
	"context"

	"github.com/google/uuid"
)

// GetPreferences retrieves all notification preferences for a user
func (s *Service) GetPreferences(ctx context.Context, userID uuid.UUID) ([]*NotificationPreference, error) {
	prefs, err := s.repo.GetPreferences(ctx, userID)
	if err != nil {
		return nil, err
	}

	if len(prefs) == 0 {
		if err := s.repo.CreateDefaultPreferences(ctx, userID); err != nil {
			return nil, err
		}
		return s.repo.GetPreferences(ctx, userID)
	}

	return prefs, nil
}

// SavePreference saves a notification preference
func (s *Service) SavePreference(ctx context.Context, pref *NotificationPreference) error {
	return s.repo.SavePreference(ctx, pref)
}

// GetUserSettings returns merged user settings for preference evaluation.
func (s *Service) GetUserSettings(ctx context.Context, userID uuid.UUID) (map[string]interface{}, error) {
	settings, err := s.db.GetUserSettings(userID)
	if err != nil {
		return map[string]interface{}{}, nil
	}
	return settings, nil
}

// UpdateUserNotificationSettings patches user_settings and syncs notification_preferences.
func (s *Service) UpdateUserNotificationSettings(ctx context.Context, userID uuid.UUID, patch map[string]interface{}) error {
	current, err := s.db.GetUserSettings(userID)
	if err != nil || current == nil {
		current = map[string]interface{}{}
	}
	for k, v := range patch {
		current[k] = v
	}
	if err := s.db.UpdateUserSettings(userID, current); err != nil {
		return err
	}
	return SyncPreferencesFromSettings(ctx, s.repo, userID, current)
}

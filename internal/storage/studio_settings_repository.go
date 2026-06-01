package storage

import (
	"context"
	"database/sql"
	"encoding/json"
	"fmt"
	"time"

	"github.com/google/uuid"
)

type StudioSettings struct {
	ID          string    `json:"id"`
	TenantID    string    `json:"tenant_id"`
	UserID      string    `json:"user_id"`
	Environment string    `json:"environment"`
	Settings    Settings  `json:"settings"`
	CreatedAt   time.Time `json:"created_at"`
	UpdatedAt   time.Time `json:"updated_at"`
}

type Settings struct {
	Theme               string               `json:"theme"`
	PrimaryColor        string               `json:"primary_color"`
	FontSize            int                  `json:"font_size"`
	SidebarPosition     string               `json:"sidebar_position"`
	CompactMode         bool                 `json:"compact_mode"`
	AnimationsEnabled   bool                 `json:"animations_enabled"`
	Transparency        bool                 `json:"transparency_enabled"`
	NotificationLevel   string               `json:"notification_level"`
	SoundEnabled        bool                 `json:"sound_enabled"`
	AutoSave            bool                 `json:"auto_save"`
	AutoSaveInterval    int                  `json:"auto_save_interval"`
	EditorFeatures      EditorFeatures       `json:"editor_features"`
	AccountPreferences  *AccountPreferences `json:"account_preferences,omitempty"`
}

type EditorFeatures struct {
	BracketMatching bool `json:"bracket_matching"`
	Minimap         bool `json:"minimap"`
	LineNumbers     bool `json:"line_numbers"`
	WordWrap        bool `json:"word_wrap"`
}

type LastWorkspaceState struct {
	Route string `json:"route"`
}

type AccountPreferences struct {
	LaunchAtLogin           bool                `json:"launch_at_login"`
	MinimizeToTrayOnClose   bool                `json:"minimize_to_tray_on_close"`
	RestoreLastWorkspace    bool                `json:"restore_last_workspace"`
	OpenLinksExternally     bool                `json:"open_links_externally"`
	LastWorkspace           *LastWorkspaceState `json:"last_workspace,omitempty"`
}

type AccountPreferencesPatch struct {
	LaunchAtLogin         *bool `json:"launch_at_login,omitempty"`
	MinimizeToTrayOnClose *bool `json:"minimize_to_tray_on_close,omitempty"`
	RestoreLastWorkspace  *bool `json:"restore_last_workspace,omitempty"`
	OpenLinksExternally   *bool `json:"open_links_externally,omitempty"`
	LastWorkspace         *LastWorkspaceState `json:"last_workspace,omitempty"`
}

func DefaultAccountPreferences() *AccountPreferences {
	return&AccountPreferences{
		LaunchAtLogin:         false,
		MinimizeToTrayOnClose: true,
		RestoreLastWorkspace:  true,
		OpenLinksExternally:   true,
		LastWorkspace:         nil,
	}
}

func ApplyAccountPreferencesPatch(current *AccountPreferences, patch AccountPreferencesPatch) *AccountPreferences {
	result := *current
	if patch.LaunchAtLogin != nil {
		result.LaunchAtLogin = *patch.LaunchAtLogin
	}
	if patch.MinimizeToTrayOnClose != nil {
		result.MinimizeToTrayOnClose = *patch.MinimizeToTrayOnClose
	}
	if patch.RestoreLastWorkspace != nil {
		result.RestoreLastWorkspace = *patch.RestoreLastWorkspace
	}
	if patch.OpenLinksExternally != nil {
		result.OpenLinksExternally = *patch.OpenLinksExternally
	}
	if patch.LastWorkspace != nil {
		result.LastWorkspace = patch.LastWorkspace
	}
	return &result
}

func MergeAccountPreferences(existing, incoming *AccountPreferences) *AccountPreferences {
	result := *incoming
	if result.LastWorkspace == nil && existing.LastWorkspace != nil {
		result.LastWorkspace = existing.LastWorkspace
	}
	return &result
}

func DefaultSettings() *Settings {
	return&Settings{
		Theme:               "dark",
		PrimaryColor:        "#6366f1",
		FontSize:            14,
		SidebarPosition:     "left",
		CompactMode:         false,
		AnimationsEnabled:   true,
		Transparency:        false,
		NotificationLevel:   "all",
		SoundEnabled:        true,
		AutoSave:            true,
		AutoSaveInterval:    30,
		EditorFeatures: EditorFeatures{
			BracketMatching: true,
			Minimap:         true,
			LineNumbers:     true,
			WordWrap:        false,
		},
		AccountPreferences: DefaultAccountPreferences(),
	}
}

type StudioSettingsRepository struct {
	db *sql.DB
}

func NewStudioSettingsRepository(db *sql.DB) *StudioSettingsRepository {
	return &StudioSettingsRepository{db: db}
}

func (r *StudioSettingsRepository) GetSettings(ctx context.Context, tenantID, userID, environment string) (*StudioSettings, error) {
	query := `
		SELECT id, tenant_id, COALESCE(user_id, ''), COALESCE(environment, ''), settings, created_at, updated_at
		FROM studio_settings
		WHERE tenant_id = $1 AND COALESCE(user_id, '') = $2 AND COALESCE(environment, '') = $3`

	var st StudioSettings
	var settingsRaw []byte
	err := r.db.QueryRowContext(ctx, query, tenantID, userID, environment).Scan(
		&st.ID, &st.TenantID, &st.UserID, &st.Environment, &settingsRaw, &st.CreatedAt, &st.UpdatedAt,
	)
	if err == sql.ErrNoRows {
		return nil, nil
	}
	if err != nil {
		return nil, fmt.Errorf("get studio settings: %w", err)
	}
	if len(settingsRaw) > 0 {
		_ = json.Unmarshal(settingsRaw, &st.Settings)
	}
	return &st, nil
}

func (r *StudioSettingsRepository) SaveSettings(ctx context.Context, tenantID, userID, environment string, settings Settings) (*StudioSettings, error) {
	settingsRaw, _ := json.Marshal(settings)
	id := uuid.New().String()

	query := `
		INSERT INTO studio_settings (id, tenant_id, user_id, environment, settings, created_at, updated_at)
		VALUES ($1, $2, $3, $4, $5, NOW(), NOW())
		ON CONFLICT (tenant_id, user_id, environment) DO UPDATE
		SET settings = $5, updated_at = NOW()
		RETURNING id, tenant_id, COALESCE(user_id, ''), COALESCE(environment, ''), settings, created_at, updated_at`

	var st StudioSettings
	var settingsOut []byte
	err := r.db.QueryRowContext(ctx, query, id, tenantID, userID, environment, settingsRaw).Scan(
		&st.ID, &st.TenantID, &st.UserID, &st.Environment, &settingsOut, &st.CreatedAt, &st.UpdatedAt,
	)
	if err != nil {
		return nil, fmt.Errorf("save studio settings: %w", err)
	}
	if len(settingsOut) > 0 {
		_ = json.Unmarshal(settingsOut, &st.Settings)
	}
	return &st, nil
}

func (r *StudioSettingsRepository) DeleteSettings(ctx context.Context, tenantID, userID, environment string) error {
	query := `DELETE FROM studio_settings WHERE tenant_id = $1 AND COALESCE(user_id, '') = $2 AND COALESCE(environment, '') = $3`
	_, err := r.db.ExecContext(ctx, query, tenantID, userID, environment)
	if err != nil {
		return fmt.Errorf("delete studio settings: %w", err)
	}
	return nil
}
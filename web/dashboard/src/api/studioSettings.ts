import { apiClient } from './client';

export interface EditorFeatures {
  bracket_matching: boolean;
  minimap: boolean;
  line_numbers: boolean;
  word_wrap: boolean;
}

export interface StudioSettings {
  theme: 'dark' | 'light' | 'system';
  primary_color: string;
  font_size: number;
  sidebar_position: 'left' | 'right';
  compact_mode: boolean;
  animations_enabled: boolean;
  transparency_enabled: boolean;
  notification_level: 'all' | 'important' | 'critical' | 'none';
  sound_enabled: boolean;
  auto_save: boolean;
  auto_save_interval: number;
  editor_features: EditorFeatures;
}

export const studioSettingsApi = {
  getSettings: () =>
    apiClient.get<{ settings: StudioSettings }>('/v1/studio/settings'),

  saveSettings: async (settings: StudioSettings) => {
    console.log('[studioSettingsApi] Sending settings:', JSON.stringify(settings, null, 2));
    const result = await apiClient.put<{ settings: StudioSettings }>('/v1/studio/settings', { settings });
    console.log('[studioSettingsApi] Received response:', result);
    return result;
  },

  resetSettings: () =>
    apiClient.delete<{ settings: StudioSettings }>('/v1/studio/settings'),
};
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
  // Privacy
  usage_analytics_enabled?: boolean;
  crash_reports_enabled?: boolean;
  // Shortcuts
  show_shortcut_hints?: boolean;
  // Performance (Tauri)
  gpu_acceleration_enabled?: boolean;
  developer_tools_enabled?: boolean;
  memory_limit_mb?: number;
  // Network
  proxy_enabled?: boolean;
  proxy_url?: string;
  proxy_bypass?: string;
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
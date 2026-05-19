import { useState, useCallback, useEffect } from 'react';
import { useQuery, useMutation, useQueryClient } from '@tanstack/react-query';
import { studioSettingsApi, type StudioSettings } from '@/api/studioSettings';
import { useThemeStore } from '@/stores/themeStore';

const STUDIO_SETTINGS_KEY = 'studio-settings';

const DEFAULT_SETTINGS: StudioSettings = {
  theme: 'dark',
  primary_color: 'orange',
  font_size: 14,
  sidebar_position: 'left',
  compact_mode: false,
  animations_enabled: true,
  transparency_enabled: true,
  notification_level: 'all',
  sound_enabled: true,
  auto_save: true,
  auto_save_interval: 30,
  editor_features: {
    bracket_matching: true,
    minimap: true,
    line_numbers: true,
    word_wrap: false,
  },
};

export function useStudioSettings() {
  const queryClient = useQueryClient();
  const { setTheme } = useThemeStore();
  const [localSettings, setLocalSettings] = useState<StudioSettings | null>(null);

  const { data, isLoading, error } = useQuery({
    queryKey: [STUDIO_SETTINGS_KEY],
    queryFn: async () => {
      console.log('[useStudioSettings] Fetching settings from API...');
      const res = await studioSettingsApi.getSettings();
      console.log('[useStudioSettings] Received settings:', res.settings);
      return res.settings;
    },
    staleTime: 1000 * 60 * 5,
  });

  const saveMutation = useMutation({
    mutationFn: async (settings: StudioSettings) => {
      console.log('[useStudioSettings] Calling saveSettings API with:', JSON.stringify(settings, null, 2));
      try {
        const result = await studioSettingsApi.saveSettings(settings);
        console.log('[useStudioSettings] saveSettings returned:', result);
        return result;
      } catch (err) {
        console.error('[useStudioSettings] API call failed:', err);
        throw err;
      }
    },
    onSuccess: (data) => {
      console.log('[useStudioSettings] Save mutation success, data:', data);
      queryClient.setQueryData([STUDIO_SETTINGS_KEY], data.settings);
      setLocalSettings(data.settings);
    },
    onError: (err: any) => {
      console.error('[useStudioSettings] Save mutation failed:', err?.response?.data || err?.message || err);
    },
  });

  const resetMutation = useMutation({
    mutationFn: () => studioSettingsApi.resetSettings(),
    onSuccess: (data) => {
      queryClient.setQueryData([STUDIO_SETTINGS_KEY], data.settings);
      setLocalSettings(data.settings);
    },
  });

  const updateSetting = useCallback(<K extends keyof StudioSettings>(
    key: K,
    value: StudioSettings[K]
  ) => {
    setLocalSettings((prev) => {
      if (!prev) {
        return { ...DEFAULT_SETTINGS, [key]: value };
      }
      return { ...prev, [key]: value };
    });
  }, []);

  const saveSettings = useCallback(async (settings: StudioSettings) => {
    await saveMutation.mutateAsync(settings);
  }, [saveMutation]);

  const resetSettings = useCallback(async () => {
    setLocalSettings(DEFAULT_SETTINGS);
    await resetMutation.mutateAsync();
  }, [resetMutation]);

  useEffect(() => {
    if (data && localSettings === null) {
      setLocalSettings(data);
      if (data.theme) {
        setTheme(data.theme);
      }
    }
  }, [data, localSettings, setTheme]);

  const settings = localSettings || data || DEFAULT_SETTINGS;
  const isDirty = localSettings !== null && JSON.stringify(localSettings) !== JSON.stringify(data);

  return {
    settings,
    defaultSettings: DEFAULT_SETTINGS,
    isLoading,
    isDirty,
    error,
    updateSetting,
    saveSettings,
    resetSettings,
    isSaving: saveMutation.isPending,
    isResetting: resetMutation.isPending,
  };
}
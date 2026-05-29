import {
  DEFAULT_STUDIO_ACCOUNT_PREFERENCES,
  fromAccountPreferencesUI,
  studioAccountSettingsApi,
  toAccountPreferencesUI,
  type StudioAccountPreferences,
  type StudioAccountPreferencesUI,
} from '@/api/studioAccountSettings';
import { useMutation, useQuery, useQueryClient } from '@tanstack/react-query';
import { useCallback } from 'react';

const STUDIO_ACCOUNT_SETTINGS_KEY = 'studio-account-settings';

export type { StudioAccountPreferencesUI };

export function openStudioExternalUrl(url: string): void {
  if (typeof window === 'undefined') return;
  window.open(url, '_blank', 'noopener,noreferrer');
}

export function useStudioAccountPreferences() {
  const queryClient = useQueryClient();

  const { data, isLoading, error } = useQuery({
    queryKey: [STUDIO_ACCOUNT_SETTINGS_KEY],
    queryFn: async () => {
      const res = await studioAccountSettingsApi.getAccountSettings();
      return res.account_preferences;
    },
    staleTime: 1000 * 60 * 5,
  });

  const saveMutation = useMutation({
    mutationFn: (prefs: StudioAccountPreferences) =>
      studioAccountSettingsApi.saveAccountSettings(prefs),
    onSuccess: (res) => {
      queryClient.setQueryData([STUDIO_ACCOUNT_SETTINGS_KEY], res.account_preferences);
    },
  });

  const patchMutation = useMutation({
    mutationFn: studioAccountSettingsApi.patchAccountSettings,
    onSuccess: (res) => {
      queryClient.setQueryData([STUDIO_ACCOUNT_SETTINGS_KEY], res.account_preferences);
    },
  });

  const resetMutation = useMutation({
    mutationFn: studioAccountSettingsApi.resetAccountSettings,
    onSuccess: (res) => {
      queryClient.setQueryData([STUDIO_ACCOUNT_SETTINGS_KEY], res.account_preferences);
    },
  });

  const apiPreferences = data ?? DEFAULT_STUDIO_ACCOUNT_PREFERENCES;
  const preferences = toAccountPreferencesUI(apiPreferences);

  const updatePreference = useCallback(
    async <K extends keyof StudioAccountPreferencesUI>(
      key: K,
      value: StudioAccountPreferencesUI[K]
    ) => {
      const patchKeyMap: Record<keyof StudioAccountPreferencesUI, keyof StudioAccountPreferences> =
        {
          launchAtLogin: 'launch_at_login',
          minimizeToTrayOnClose: 'minimize_to_tray_on_close',
          restoreLastWorkspace: 'restore_last_workspace',
          openLinksExternally: 'open_links_externally',
        };
      await patchMutation.mutateAsync({ [patchKeyMap[key]]: value });
    },
    [patchMutation]
  );

  const resetPreferences = useCallback(async () => {
    await resetMutation.mutateAsync();
  }, [resetMutation]);

  const savePreferences = useCallback(
    async (uiPrefs: StudioAccountPreferencesUI) => {
      await saveMutation.mutateAsync(fromAccountPreferencesUI(uiPrefs, apiPreferences));
    },
    [apiPreferences, saveMutation]
  );

  return {
    preferences,
    apiPreferences,
    isLoading,
    error,
    updatePreference,
    resetPreferences,
    savePreferences,
    isSaving: saveMutation.isPending || patchMutation.isPending,
    isResetting: resetMutation.isPending,
  };
}

export function useStudioLastWorkspace() {
  const queryClient = useQueryClient();

  const { data, isLoading } = useQuery({
    queryKey: [STUDIO_ACCOUNT_SETTINGS_KEY, 'last-workspace'],
    queryFn: async () => {
      const res = await studioAccountSettingsApi.getLastWorkspace();
      return res.last_workspace;
    },
    staleTime: 1000 * 60,
  });

  const saveMutation = useMutation({
    mutationFn: studioAccountSettingsApi.saveLastWorkspace,
    onSuccess: () => {
      queryClient.invalidateQueries({ queryKey: [STUDIO_ACCOUNT_SETTINGS_KEY] });
    },
  });

  const saveLastWorkspace = useCallback(
    (route: string, extras?: { workspace_id?: string; tab_id?: string }) => {
      saveMutation.mutate({
        route,
        workspace_id: extras?.workspace_id,
        tab_id: extras?.tab_id,
        updated_at: new Date().toISOString(),
      });
    },
    [saveMutation]
  );

  return {
    lastWorkspace: data ?? null,
    isLoading,
    saveLastWorkspace,
    isSaving: saveMutation.isPending,
  };
}

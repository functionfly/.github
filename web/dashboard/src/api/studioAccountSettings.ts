import { apiClient } from './client';

export interface LastWorkspaceState {
  route: string;
  workspace_id?: string;
  tab_id?: string;
  updated_at?: string;
}

export interface StudioAccountPreferences {
  launch_at_login: boolean;
  minimize_to_tray_on_close: boolean;
  restore_last_workspace: boolean;
  open_links_externally: boolean;
  last_workspace?: LastWorkspaceState | null;
}

export type StudioAccountPreferencesPatch = Partial<
  Pick<
    StudioAccountPreferences,
    | 'launch_at_login'
    | 'minimize_to_tray_on_close'
    | 'restore_last_workspace'
    | 'open_links_externally'
  >
>;

export const DEFAULT_STUDIO_ACCOUNT_PREFERENCES: StudioAccountPreferences = {
  launch_at_login: false,
  minimize_to_tray_on_close: true,
  restore_last_workspace: true,
  open_links_externally: true,
};

export const studioAccountSettingsApi = {
  getAccountSettings: () =>
    apiClient.get<{ account_preferences: StudioAccountPreferences }>('/v1/studio/settings/account'),

  saveAccountSettings: (accountPreferences: StudioAccountPreferences) =>
    apiClient.put<{ account_preferences: StudioAccountPreferences }>(
      '/v1/studio/settings/account',
      { account_preferences: accountPreferences }
    ),

  patchAccountSettings: (patch: StudioAccountPreferencesPatch) =>
    apiClient.patch<{ account_preferences: StudioAccountPreferences }>(
      '/v1/studio/settings/account',
      patch
    ),

  resetAccountSettings: () =>
    apiClient.delete<{ account_preferences: StudioAccountPreferences }>(
      '/v1/studio/settings/account'
    ),

  getLastWorkspace: () =>
    apiClient.get<{ last_workspace: LastWorkspaceState | null }>(
      '/v1/studio/settings/account/last-workspace'
    ),

  saveLastWorkspace: (lastWorkspace: LastWorkspaceState) =>
    apiClient.put<{ last_workspace: LastWorkspaceState }>(
      '/v1/studio/settings/account/last-workspace',
      { last_workspace: lastWorkspace }
    ),
};

/** UI-friendly camelCase view of account preferences. */
export interface StudioAccountPreferencesUI {
  launchAtLogin: boolean;
  minimizeToTrayOnClose: boolean;
  restoreLastWorkspace: boolean;
  openLinksExternally: boolean;
}

export function toAccountPreferencesUI(
  prefs: StudioAccountPreferences
): StudioAccountPreferencesUI {
  return {
    launchAtLogin: prefs.launch_at_login,
    minimizeToTrayOnClose: prefs.minimize_to_tray_on_close,
    restoreLastWorkspace: prefs.restore_last_workspace,
    openLinksExternally: prefs.open_links_externally,
  };
}

export function fromAccountPreferencesUI(
  prefs: StudioAccountPreferencesUI,
  existing?: StudioAccountPreferences
): StudioAccountPreferences {
  return {
    launch_at_login: prefs.launchAtLogin,
    minimize_to_tray_on_close: prefs.minimizeToTrayOnClose,
    restore_last_workspace: prefs.restoreLastWorkspace,
    open_links_externally: prefs.openLinksExternally,
    last_workspace: existing?.last_workspace ?? null,
  };
}

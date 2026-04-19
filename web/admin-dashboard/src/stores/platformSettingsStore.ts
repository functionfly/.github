/**
 * Platform Settings Store
 * Persists platform settings to localStorage as a temporary solution
 * until the backend supports PATCH /state-fabric/platform-settings.
 */

import { CACHE_KEYS } from '@/lib/constants';

export interface PlatformSettings {
  maintenanceMode: boolean;
  signupsEnabled: boolean;
  platformName: string;
  supportEmail: string;
  defaultRateLimitPerMin: number;
}

const DEFAULT_PLATFORM_SETTINGS: PlatformSettings = {
  maintenanceMode: false,
  signupsEnabled: true,
  platformName: 'FunctionFly',
  supportEmail: 'support@functionfly.com',
  defaultRateLimitPerMin: 1000,
};

function loadPlatformSettings(): PlatformSettings {
  try {
    const raw = localStorage.getItem(CACHE_KEYS.PLATFORM_SETTINGS ?? 'admin_platform_settings');
    if (!raw) return DEFAULT_PLATFORM_SETTINGS;
    const parsed = JSON.parse(raw) as Partial<PlatformSettings>;
    return {
      maintenanceMode: parsed.maintenanceMode ?? DEFAULT_PLATFORM_SETTINGS.maintenanceMode,
      signupsEnabled: parsed.signupsEnabled ?? DEFAULT_PLATFORM_SETTINGS.signupsEnabled,
      platformName: parsed.platformName ?? DEFAULT_PLATFORM_SETTINGS.platformName,
      supportEmail: parsed.supportEmail ?? DEFAULT_PLATFORM_SETTINGS.supportEmail,
      defaultRateLimitPerMin:
        parsed.defaultRateLimitPerMin ?? DEFAULT_PLATFORM_SETTINGS.defaultRateLimitPerMin,
    };
  } catch {
    return DEFAULT_PLATFORM_SETTINGS;
  }
}

function savePlatformSettings(settings: PlatformSettings): void {
  try {
    localStorage.setItem(CACHE_KEYS.PLATFORM_SETTINGS, JSON.stringify(settings));
  } catch {
    /* localStorage unavailable — silently ignore */
  }
}

export const platformSettingsStore = {
  load: loadPlatformSettings,
  save: savePlatformSettings,
  default: DEFAULT_PLATFORM_SETTINGS,
};

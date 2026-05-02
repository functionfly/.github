/**
 * FunctionFly Unified Theme System
 * Shared theme management across all web applications.
 *
 * Provides:
 * - Unified storage format: { mode: 'light' | 'dark' | 'system' }
 * - Cross-app sync via custom events
 * - Legacy storage migration (theme-storage, ff-docs-theme)
 * - System preference detection and listeners
 *
 * Usage:
 *   import { initTheme, getTheme, setTheme, subscribe } from '@functionfly/shared/theme';
 *
 *   // Initialize on app load
 *   initTheme();
 *
 *   // Get current theme
 *   const { mode, resolved } = getTheme();
 *
 *   // Set theme (broadcasts to all apps)
 *   setTheme('dark');
 *
 *   // Listen for changes from other apps
 *   subscribe((theme) => console.log('Theme changed:', theme));
 */

export type ThemeMode = 'light' | 'dark' | 'system';
export type ResolvedTheme = 'light' | 'dark';
export type Theme = ThemeMode;

export interface ThemeState {
  mode: ThemeMode;
  resolved: ResolvedTheme;
}

export interface ThemePreference {
  mode: ThemeMode;
}

const STORAGE_KEY = 'ff-user-theme';

const LEGACY_KEYS = [
  { key: 'theme-storage', transform: legacyToTheme },
  { key: 'ff-docs-theme', transform: docsToTheme },
];

function legacyToTheme(stored: string | null): ThemePreference | null {
  if (!stored) return null;
  try {
    const parsed = JSON.parse(stored);
    if (parsed.state?.theme) {
      const mode = parsed.state.theme as ThemeMode;
      if (mode === 'light' || mode === 'dark' || mode === 'system') {
        return { mode };
      }
    }
  } catch {}
  return null;
}

function docsToTheme(stored: string | null): ThemePreference | null {
  if (!stored) return null;
  if (stored === 'light' || stored === 'dark') {
    return { mode: stored };
  }
  return null;
}

function getSystemTheme(): ResolvedTheme {
  if (typeof window === 'undefined') return 'dark';
  return window.matchMedia('(prefers-color-scheme: dark)').matches ? 'dark' : 'light';
}

function resolveTheme(mode: ThemeMode): ResolvedTheme {
  if (mode === 'system') {
    return getSystemTheme();
  }
  return mode;
}

function readStorage(): ThemePreference | null {
  if (typeof window === 'undefined') return null;
  try {
    const stored = localStorage.getItem(STORAGE_KEY);
    if (stored) {
      const parsed = JSON.parse(stored) as ThemePreference;
      if (parsed.mode === 'light' || parsed.mode === 'dark' || parsed.mode === 'system') {
        return parsed;
      }
    }
  } catch {}
  return null;
}

function writeStorage(preference: ThemePreference): void {
  if (typeof window === 'undefined') return;
  try {
    localStorage.setItem(STORAGE_KEY, JSON.stringify(preference));
  } catch {}
}

function migrateLegacy(): ThemePreference | null {
  if (typeof window === 'undefined') return null;
  for (const { key, transform } of LEGACY_KEYS) {
    const legacy = localStorage.getItem(key);
    if (legacy) {
      const migrated = transform(legacy);
      if (migrated) {
        writeStorage(migrated);
        try {
          localStorage.removeItem(key);
        } catch {}
        return migrated;
      }
    }
  }
  return null;
}

let cachedState: ThemeState | null = null;
const listeners = new Set<(state: ThemeState) => void>();
let mediaQuery: MediaQueryList | null = null;
let mediaQueryHandler: (() => void) | null = null;

export function getTheme(): ThemeState {
  if (cachedState) return cachedState;
  if (typeof window === 'undefined') {
    return { mode: 'system', resolved: 'dark' };
  }
  const preference = readStorage() || migrateLegacy();
  const mode = preference?.mode || 'system';
  const resolved = resolveTheme(mode);
  cachedState = { mode, resolved };
  return cachedState;
}

export function setTheme(mode: ThemeMode): void {
  if (typeof window === 'undefined') return;
  const preference: ThemePreference = { mode };
  writeStorage(preference);
  const resolved = resolveTheme(mode);
  cachedState = { mode, resolved };
  applyThemeToDOM(resolved);
  broadcastChange();
}

function broadcastChange(): void {
  if (typeof window === 'undefined') return;
  window.dispatchEvent(new CustomEvent('ff-theme-change', { detail: cachedState }));
}

function applyThemeToDOM(resolved: ResolvedTheme): void {
  if (typeof window === 'undefined') return;
  document.documentElement.setAttribute('data-theme', resolved);
}

export function subscribe(callback: (state: ThemeState) => void): () => void {
  listeners.add(callback);
  return () => listeners.delete(callback);
}

function notifyListeners(): void {
  if (!cachedState) return;
  listeners.forEach(cb => {
    try {
      cb(cachedState!);
    } catch {}
  });
}

export function initTheme(): ThemeState {
  if (typeof window === 'undefined') {
    return { mode: 'system', resolved: 'dark' };
  }
  const migrated = migrateLegacy();
  const preference = migrated || readStorage();
  const mode = preference?.mode || 'system';
  const resolved = resolveTheme(mode);
  cachedState = { mode, resolved };
  applyThemeToDOM(resolved);
  setupSystemListener();
  window.addEventListener('ff-theme-change', ((e: CustomEvent) => {
    cachedState = e.detail;
    applyThemeToDOM(e.detail.resolved);
    notifyListeners();
  }) as EventListener);
  return cachedState;
}

function setupSystemListener(): void {
  if (typeof window === 'undefined') return;
  if (mediaQueryHandler) return;
  mediaQuery = window.matchMedia('(prefers-color-scheme: dark)');
  mediaQueryHandler = () => {
    const current = getTheme();
    if (current.mode === 'system') {
      const resolved = getSystemTheme();
      cachedState = { ...current, resolved };
      applyThemeToDOM(resolved);
      notifyListeners();
    }
  };
  mediaQuery.addEventListener('change', mediaQueryHandler);
}

export function toggleTheme(): void {
  const current = getTheme();
  const resolved = current.resolved;
  const next: ThemeMode = resolved === 'dark' ? 'light' : 'dark';
  setTheme(next);
}

export function getResolvedTheme(mode?: ThemeMode): ResolvedTheme {
  return resolveTheme(mode ?? getTheme().mode);
}
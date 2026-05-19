/**
 * @functionfly/shared
 * Theme types and utilities
 */

export type Theme = 'light' | 'dark' | 'system';

export interface ThemeState {
  theme: Theme;
  resolvedTheme: 'light' | 'dark';
}

export function initTheme(): ThemeState {
  return { theme: 'system', resolvedTheme: 'light' };
}

export function setTheme(theme: Theme): void {
  document.documentElement.setAttribute('data-theme', theme);
}

export function subscribe(callback: (state: ThemeState) => void): () => void {
  return () => {};
}

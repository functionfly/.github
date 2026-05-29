import type { StudioSettings } from '@/api/studioSettings';
import type { Theme } from '@/stores/themeStore';

const LOCAL_CACHE_KEY = 'ff-studio-settings';

export const DEFAULT_STUDIO_SETTINGS: StudioSettings = {
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

export const colorPresets = [
  { id: 'orange', name: 'FunctionFly Orange', primary: '#f97316', secondary: '#ea580c' },
  { id: 'blue', name: 'Ocean Blue', primary: '#3b82f6', secondary: '#1d4ed8' },
  { id: 'purple', name: 'Violet Dream', primary: '#8b5cf6', secondary: '#6d28d9' },
  { id: 'emerald', name: 'Forest Green', primary: '#10b981', secondary: '#059669' },
  { id: 'rose', name: 'Rose Gold', primary: '#f43f5e', secondary: '#e11d48' },
  { id: 'slate', name: 'Midnight Slate', primary: '#64748b', secondary: '#475569' },
];

const adjustColorBrightness = (hex: string, amount: number): string => {
  const num = parseInt(hex.replace('#', ''), 16);
  const r = Math.min(255, Math.max(0, (num >> 16) + amount));
  const g = Math.min(255, Math.max(0, ((num >> 8) & 0x00ff) + amount));
  const b = Math.min(255, Math.max(0, (num & 0x0000ff) + amount));
  return '#' + ((1 << 24) + (r << 16) + (g << 8) + b).toString(16).slice(1);
};

export function readLocalStudioSettings(): StudioSettings | null {
  if (typeof window === 'undefined') return null;
  try {
    const raw = localStorage.getItem(LOCAL_CACHE_KEY);
    if (!raw) return null;
    const parsed = JSON.parse(raw) as StudioSettings;
    return { ...DEFAULT_STUDIO_SETTINGS, ...parsed };
  } catch {
    return null;
  }
}

export function writeLocalStudioSettings(settings: StudioSettings): void {
  if (typeof window === 'undefined') return;
  try {
    localStorage.setItem(LOCAL_CACHE_KEY, JSON.stringify(settings));
  } catch {
    /* ignore quota errors */
  }
}

export function clearLocalStudioSettings(): void {
  if (typeof window === 'undefined') return;
  try {
    localStorage.removeItem(LOCAL_CACHE_KEY);
  } catch {
    /* ignore */
  }
}

export function applyColorPalette(presetId: string): void {
  const preset = colorPresets.find((p) => p.id === presetId);
  if (!preset || typeof document === 'undefined') return;

  const lighter = adjustColorBrightness(preset.primary, 40);
  const darker = adjustColorBrightness(preset.primary, -40);

  document.documentElement.style.setProperty('--color-brand-400', lighter);
  document.documentElement.style.setProperty('--color-brand-500', preset.primary);
  document.documentElement.style.setProperty('--color-brand-600', preset.secondary);
  document.documentElement.style.setProperty('--color-brand-700', darker);
  document.documentElement.style.setProperty('--color-velocity-400', lighter);
  document.documentElement.style.setProperty('--color-velocity-500', preset.primary);
  document.documentElement.style.setProperty('--color-velocity-600', preset.secondary);
  document.documentElement.style.setProperty('--text-accent', preset.primary);
  document.documentElement.style.setProperty('--button-primary', preset.primary);
  document.documentElement.style.setProperty('--button-primary-hover', preset.secondary);
  document.documentElement.style.setProperty('--border-focus', preset.primary + '80');

  let styleEl = document.getElementById('dynamic-brand-colors');
  if (!styleEl) {
    styleEl = document.createElement('style');
    styleEl.id = 'dynamic-brand-colors';
    document.head.appendChild(styleEl);
  }
  styleEl.textContent = `
    [data-studio] .text-brand-400,
    .studio-root .text-brand-400 { color: var(--color-brand-400) !important; }
    [data-studio] .text-brand-300,
    .studio-root .text-brand-300 { color: var(--color-brand-300) !important; }
    [data-studio] .bg-brand-500,
    .studio-root .bg-brand-500 { background-color: var(--color-brand-500) !important; }
    [data-studio] .border-brand-500,
    .studio-root .border-brand-500 { border-color: var(--color-brand-500) !important; }
    .text-brand-400 { color: var(--color-brand-400, ${preset.primary}) !important; }
    .text-brand-300 { color: var(--color-brand-300, ${lighter}) !important; }
    .bg-brand-500 { background-color: var(--color-brand-500, ${preset.primary}) !important; }
    .border-brand-500 { border-color: var(--color-brand-500, ${preset.primary}) !important; }
  `;
  document.documentElement.style.setProperty('--studio-brand-primary', preset.primary);
  document.documentElement.style.setProperty('--studio-brand-secondary', preset.secondary);
}

export function applyFontSize(size: number): void {
  if (typeof document === 'undefined') return;
  const newSize = `${size}px`;
  document.documentElement.style.setProperty('--studio-font-size', newSize);

  let fontSizeStyle = document.getElementById('dynamic-font-size');
  if (!fontSizeStyle) {
    fontSizeStyle = document.createElement('style');
    fontSizeStyle.id = 'dynamic-font-size';
    document.head.appendChild(fontSizeStyle);
  }
  fontSizeStyle.textContent = `
    body, body * { font-size: ${newSize} !important; }
    .studio-root, .studio-root * { font-size: ${newSize} !important; }
  `;
}

export function applyStudioAppearance(
  settings: StudioSettings,
  setTheme: (theme: Theme) => void
): void {
  if (settings.theme) {
    setTheme(settings.theme);
  }
  applyColorPalette(settings.primary_color);
  applyFontSize(settings.font_size);
}

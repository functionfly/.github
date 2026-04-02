/**
 * Supported languages for the FunctionFly dashboard.
 * Each entry maps an ISO 639-1 code to its display name and optional region variant.
 */
export interface Language {
  code: string; // ISO 639-1 (e.g. "en", "es", "zh")
  name: string; // Display name in the native language (e.g. "Español", "中文")
  nameEn: string; // English name (e.g. "Spanish", "Chinese")
  flag: string; // Two-letter flag emoji
  dir?: 'ltr' | 'rtl'; // Text direction; defaults to 'ltr'
}

export const SUPPORTED_LANGUAGES: Language[] = [
  { code: 'en', name: 'English', nameEn: 'English', flag: '🇺🇸' },
  { code: 'es', name: 'Español', nameEn: 'Spanish', flag: '🇪🇸' },
  { code: 'fr', name: 'Français', nameEn: 'French', flag: '🇫🇷' },
  { code: 'de', name: 'Deutsch', nameEn: 'German', flag: '🇩🇪' },
  { code: 'zh', name: '中文', nameEn: 'Chinese', flag: '🇨🇳' },
  { code: 'ja', name: '日本語', nameEn: 'Japanese', flag: '🇯🇵' },
  { code: 'ko', name: '한국어', nameEn: 'Korean', flag: '🇰🇷' },
  { code: 'pt', name: 'Português', nameEn: 'Portuguese', flag: '🇧🇷' },
  { code: 'ar', name: 'العربية', nameEn: 'Arabic', flag: '🇸🇦', dir: 'rtl' },
  { code: 'ru', name: 'Русский', nameEn: 'Russian', flag: '🇷🇺' },
  { code: 'hi', name: 'हिन्दी', nameEn: 'Hindi', flag: '🇮🇳' },
  { code: 'nl', name: 'Nederlands', nameEn: 'Dutch', flag: '🇳🇱' },
  { code: 'pl', name: 'Polski', nameEn: 'Polish', flag: '🇵🇱' },
  { code: 'tr', name: 'Türkçe', nameEn: 'Turkish', flag: '🇹🇷' },
  { code: 'vi', name: 'Tiếng Việt', nameEn: 'Vietnamese', flag: '🇻🇳' },
];

/** Languages sorted by native name for the picker UI */
export const LANGUAGES_BY_NATIVE_NAME = [...SUPPORTED_LANGUAGES].sort((a, b) =>
  a.name.localeCompare(b.name)
);

/** All RTL languages for HTML dir attribute */
export const RTL_LANGUAGES = new Set(
  SUPPORTED_LANGUAGES.filter((l) => l.dir === 'rtl').map((l) => l.code)
);

/** Language auto-detection: browser preferred languages → best match in SUPPORTED_LANGUAGES. */
export function detectLanguage(): string {
  if (typeof navigator === 'undefined') return 'en';

  const browserLangs = navigator.languages ?? [navigator.language ?? 'en'];

  for (const browserLang of browserLangs) {
    // Try exact match first (e.g. "es-ES" → "es")
    const short = browserLang.split('-')[0].toLowerCase();
    const match = SUPPORTED_LANGUAGES.find((l) => l.code === short);
    if (match) return match.code;
  }

  return 'en';
}

/** Look up a language by its ISO 639-1 code. */
export function getLanguage(code: string): Language | undefined {
  return SUPPORTED_LANGUAGES.find((l) => l.code === code.toLowerCase());
}

/** Persist language preference to localStorage. */
export function persistLanguage(code: string): void {
  if (typeof localStorage !== 'undefined') {
    localStorage.setItem('ff-language', code);
  }
}

/** Load persisted language preference from localStorage. */
export function loadPersistedLanguage(): string | null {
  if (typeof localStorage === 'undefined') return null;
  return localStorage.getItem('ff-language');
}

// Shared language configuration for FunctionFly marketing + docs sites.
// Both web/site/src and web/docs/src import from this.

export interface Language {
  code: string;        // ISO 639-1, e.g. "en", "es", "zh"
  name: string;        // Display name in the native language, e.g. "Español"
  nameEn: string;      // English name, e.g. "Spanish"
  flag: string;        // Two-letter flag emoji
  dir?: 'ltr' | 'rtl'; // Text direction; defaults to "ltr"
}

export const SUPPORTED_LANGUAGES: Language[] = [
  { code: 'en', name: 'English',     nameEn: 'English',     flag: '🇺🇸' },
  { code: 'es', name: 'Español',     nameEn: 'Spanish',      flag: '🇪🇸' },
  { code: 'fr', name: 'Français',    nameEn: 'French',       flag: '🇫🇷' },
  { code: 'de', name: 'Deutsch',     nameEn: 'German',       flag: '🇩🇪' },
  { code: 'zh', name: '中文',         nameEn: 'Chinese',      flag: '🇨🇳' },
  { code: 'ja', name: '日本語',       nameEn: 'Japanese',     flag: '🇯🇵' },
  { code: 'ko', name: '한국어',       nameEn: 'Korean',       flag: '🇰🇷' },
  { code: 'pt', name: 'Português',   nameEn: 'Portuguese',   flag: '🇧🇷' },
  { code: 'ar', name: 'العربية',      nameEn: 'Arabic',       flag: '🇸🇦', dir: 'rtl' },
  { code: 'ru', name: 'Русский',     nameEn: 'Russian',      flag: '🇷🇺' },
  { code: 'hi', name: 'हिन्दी',        nameEn: 'Hindi',        flag: '🇮🇳' },
  { code: 'nl', name: 'Nederlands',  nameEn: 'Dutch',        flag: '🇳🇱' },
  { code: 'pl', name: 'Polski',      nameEn: 'Polish',       flag: '🇵🇱' },
  { code: 'tr', name: 'Türkçe',      nameEn: 'Turkish',      flag: '🇹🇷' },
  { code: 'vi', name: 'Tiếng Việt',  nameEn: 'Vietnamese',   flag: '🇻🇳' },
];

export const DEFAULT_LANGUAGE = 'en';

/** Detect browser language from Accept-Language header */
export function detectLanguage(acceptLanguage: string | null): string {
  if (!acceptLanguage) return DEFAULT_LANGUAGE;
  const parts = acceptLanguage.split(',').map((p) => p.trim().split(';')[0].toLowerCase());
  for (const part of parts) {
    // Try exact match first
    const exact = SUPPORTED_LANGUAGES.find((l) => l.code === part || l.code === part.split('-')[0]);
    if (exact) return exact.code;
    // Try language-only prefix (e.g. "es" matches "es-MX")
    const prefix = part.split('-')[0];
    const match = SUPPORTED_LANGUAGES.find((l) => l.code === prefix);
    if (match) return match.code;
  }
  return DEFAULT_LANGUAGE;
}

/** Get language by code */
export function getLanguage(code: string): Language | undefined {
  return SUPPORTED_LANGUAGES.find((l) => l.code === code);
}

/** Languages sorted by native name */
export const LANGUAGES_BY_NATIVE_NAME = [...SUPPORTED_LANGUAGES].sort((a, b) =>
  a.name.localeCompare(b.name)
);

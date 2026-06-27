export interface Language {
  code: string;
  name: string;
  nameEn: string;
  flag: string;
  dir?: "ltr" | "rtl";
}

export const SUPPORTED_LANGUAGES: Language[] = [
  { code: "en", name: "English",     nameEn: "English",     flag: "🇺🇸" },
  { code: "es", name: "Español",     nameEn: "Spanish",      flag: "🇪🇸" },
  { code: "fr", name: "Français",    nameEn: "French",       flag: "🇫🇷" },
  { code: "de", name: "Deutsch",     nameEn: "German",       flag: "🇩🇪" },
  { code: "zh", name: "中文",         nameEn: "Chinese",      flag: "🇨🇳" },
  { code: "ja", name: "日本語",       nameEn: "Japanese",     flag: "🇯🇵" },
  { code: "ko", name: "한국어",       nameEn: "Korean",       flag: "🇰🇷" },
  { code: "pt", name: "Português",   nameEn: "Portuguese",   flag: "🇧🇷" },
  { code: "ar", name: "العربية",      nameEn: "Arabic",       flag: "🇸🇦", dir: "rtl" },
  { code: "ru", name: "Русский",     nameEn: "Russian",      flag: "🇷🇺" },
  { code: "hi", name: "हिन्दी",        nameEn: "Hindi",        flag: "🇮🇳" },
  { code: "nl", name: "Nederlands",  nameEn: "Dutch",        flag: "🇳🇱" },
  { code: "pl", name: "Polski",      nameEn: "Polish",       flag: "🇵🇱" },
  { code: "tr", name: "Türkçe",      nameEn: "Turkish",      flag: "🇹🇷" },
  { code: "vi", name: "Tiếng Việt",  nameEn: "Vietnamese",   flag: "🇻🇳" },
];

export const DEFAULT_LANGUAGE = "en";

export function getLanguage(code: string): Language | undefined {
  return SUPPORTED_LANGUAGES.find((l) => l.code === code);
}

export function isRtlLang(code: string): boolean {
  const lang = getLanguage(code);
  return lang?.dir === "rtl";
}

export const LANGUAGES_BY_NATIVE_NAME = [...SUPPORTED_LANGUAGES].sort((a, b) =>
  a.name.localeCompare(b.name)
);
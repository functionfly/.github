import { defaultLocale } from '../lib/i18n/languages';

export type Locale = 'en' | 'es' | 'fr' | 'de' | 'zh' | 'ja' | 'ko' | 'pt' | 'ar' | 'ru' | 'hi' | 'nl' | 'pl' | 'tr' | 'vi';

export const SUPPORTED_LOCALES: Locale[] = ['en', 'es', 'fr', 'de', 'zh', 'ja', 'ko', 'pt', 'ar', 'ru', 'hi', 'nl', 'pl', 'tr', 'vi'];

type TranslationData = Record<string, string>;

const translationCache: Record<string, TranslationData> = {};

export async function loadTranslations(locale: Locale): Promise<TranslationData> {
  if (translationCache[locale]) {
    return translationCache[locale];
  }

  try {
    const translations = await import(`../locales/${locale}/common.json`);
    translationCache[locale] = translations.default || translations;
    return translationCache[locale];
  } catch {
    if (locale !== defaultLocale) {
      return loadTranslations(defaultLocale as Locale);
    }
    return {};
  }
}

export function createT(translations: TranslationData) {
  return function t(key: string, params?: Record<string, string | number>): string {
    let text = translations[key] || key;

    if (params) {
      Object.entries(params).forEach(([k, v]) => {
        text = text.replace(new RegExp(`{{${k}}}`, 'g'), String(v));
      });
    }

    return text;
  };
}

export function getLocaleFromUrl(url: URL): Locale {
  const [, localePart] = url.pathname.split('/');
  const locale = localePart as Locale;
  return SUPPORTED_LOCALES.includes(locale) ? locale : defaultLocale;
}

export function getLocalizedPath(path: string, locale: Locale): string {
  const cleanPath = path.replace(/^\//, '');
  return `/${locale}/${cleanPath}`;
}
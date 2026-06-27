/**
 * i18n helpers for the FunctionFly auth site.
 *
 * Usage in .astro frontmatter:
 *   import { t, getLocale } from '@/i18n';
 *   const lang = getLocale(Astro);
 *   const text = t(lang, 'login.title');
 */

import enCommon from "@/locales/en/common.json";
import esCommon from "@/locales/es/common.json";
import frCommon from "@/locales/fr/common.json";
import deCommon from "@/locales/de/common.json";
import zhCommon from "@/locales/zh/common.json";
import jaCommon from "@/locales/ja/common.json";
import koCommon from "@/locales/ko/common.json";
import ptCommon from "@/locales/pt/common.json";
import arCommon from "@/locales/ar/common.json";
import ruCommon from "@/locales/ru/common.json";
import hiCommon from "@/locales/hi/common.json";
import nlCommon from "@/locales/nl/common.json";
import plCommon from "@/locales/pl/common.json";
import trCommon from "@/locales/tr/common.json";
import viCommon from "@/locales/vi/common.json";

type TranslationDict = Record<string, unknown>;

export const LOCALE_DATA: Record<string, TranslationDict> = {
  en: enCommon as TranslationDict,
  es: esCommon as TranslationDict,
  fr: frCommon as TranslationDict,
  de: deCommon as TranslationDict,
  zh: zhCommon as TranslationDict,
  ja: jaCommon as TranslationDict,
  ko: koCommon as TranslationDict,
  pt: ptCommon as TranslationDict,
  ar: arCommon as TranslationDict,
  ru: ruCommon as TranslationDict,
  hi: hiCommon as TranslationDict,
  nl: nlCommon as TranslationDict,
  pl: plCommon as TranslationDict,
  tr: trCommon as TranslationDict,
  vi: viCommon as TranslationDict,
};

export const defaultLocale = "en";

export const SUPPORTED_LOCALES = [
  "en", "es", "fr", "de", "zh", "ja", "ko",
  "pt", "ar", "ru", "hi", "nl", "pl", "tr", "vi",
] as const;

export type Locale = (typeof SUPPORTED_LOCALES)[number];

function resolveKey(dict: TranslationDict, key: string): string {
  const parts = key.split(".");
  let current: unknown = dict;
  for (const part of parts) {
    if (current && typeof current === "object" && part in (current as Record<string, unknown>)) {
      current = (current as Record<string, unknown>)[part];
    } else {
      return key;
    }
  }
  return typeof current === "string" ? current : key;
}

export function getLocale(Astro: { currentLocale?: string }): string {
  return Astro.currentLocale ?? defaultLocale;
}

export function t(locale: string, key: string): string {
  const dict = LOCALE_DATA[locale];
  if (dict) {
    const val = resolveKey(dict, key);
    if (val !== key) return val;
  }

  if (locale !== defaultLocale) {
    const enVal = resolveKey(LOCALE_DATA[defaultLocale] ?? {}, key);
    if (enVal !== key) return enVal;
  }

  return key;
}

export function isRtl(locale: string): boolean {
  return locale === "ar" || locale === "he" || locale === "fa";
}
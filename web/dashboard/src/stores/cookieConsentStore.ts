import { create } from 'zustand';
import { persist } from 'zustand/middleware';

export interface CookieCategories {
  necessary: boolean;
  analytics: boolean;
  marketing: boolean;
  functionality: boolean;
}

export interface CookieConsentState {
  hasConsent: boolean;
  consentTimestamp: Date | null;
  categories: CookieCategories;
  setConsent: (categories: CookieCategories) => void;
  resetConsent: () => void;
  updateCategory: (category: keyof CookieCategories, enabled: boolean) => void;
}

export const useCookieConsentStore = create<CookieConsentState>()(
  persist(
    (set) => ({
      hasConsent: false,
      consentTimestamp: null,
      categories: {
        necessary: true, // Always true for necessary cookies
        analytics: false,
        marketing: false,
        functionality: false,
      },
      setConsent: (categories: CookieCategories) =>
        set({
          hasConsent: true,
          consentTimestamp: new Date(),
          categories: { ...categories, necessary: true }, // Necessary is always true
        }),
      resetConsent: () =>
        set({
          hasConsent: false,
          consentTimestamp: null,
          categories: {
            necessary: true,
            analytics: false,
            marketing: false,
            functionality: false,
          },
        }),
      updateCategory: (category: keyof CookieCategories, enabled: boolean) =>
        set((state) => ({
          categories: {
            ...state.categories,
            [category]: category === 'necessary' ? true : enabled, // Necessary is always true
          },
        })),
    }),
    {
      name: 'cookie-consent-storage',
      partialize: (state) => ({
        hasConsent: state.hasConsent,
        consentTimestamp: state.consentTimestamp,
        categories: state.categories,
      }),
    }
  )
);
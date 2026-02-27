'use client';

import { useEffect, ReactNode } from 'react';
import 'vanilla-cookieconsent/dist/cookieconsent.css';
import '@/styles/cookie-consent.css';
import * as CookieConsent from 'vanilla-cookieconsent';
import { useCookieConsentStore, CookieCategories } from '@/stores/cookieConsentStore';
import { cookieConsentConfig } from '@/config/cookieConsentConfig';
import { updateGoogleConsentFromCategories } from '@/config/googleConsentMode';

interface CookieConsentProviderProps {
  children: ReactNode;
}

export function CookieConsentProvider({ children }: CookieConsentProviderProps) {
  const { setConsent } = useCookieConsentStore();

  useEffect(() => {
    // Initialize cookie consent
    CookieConsent.run(cookieConsentConfig).then(() => {
      // Check if user has already given consent
      if (CookieConsent.validConsent()) {
        const cookie = CookieConsent.getCookie();
        if (cookie && cookie.categories.length > 0) {
          const categories: CookieCategories = {
            necessary: true,
            analytics: cookie.categories.includes('analytics'),
            marketing: cookie.categories.includes('marketing'),
            functionality: cookie.categories.includes('functionality'),
          };

          setConsent(categories);
          // Update Google Consent Mode for existing consent
          updateGoogleConsentFromCategories();
        }
      }
    });

    // Since vanilla-cookieconsent doesn't have event listeners,
    // we need to periodically check for consent changes
    // This is a workaround since the library doesn't expose events
    const checkConsentInterval = setInterval(() => {
      if (CookieConsent.validConsent()) {
        const cookie = CookieConsent.getCookie();
        if (cookie) {
          const categories: CookieCategories = {
            necessary: true,
            analytics: cookie.categories.includes('analytics'),
            marketing: cookie.categories.includes('marketing'),
            functionality: cookie.categories.includes('functionality'),
          };

          // Only update if categories have changed
          const currentCategories = useCookieConsentStore.getState().categories;
          const hasChanged = Object.keys(categories).some(
            key => categories[key as keyof CookieCategories] !== currentCategories[key as keyof CookieCategories]
          );

          if (hasChanged) {
            setConsent(categories);
            updateGoogleConsentFromCategories();
          }
        }
      }
    }, 1000); // Check every second

    // Cleanup function
    return () => {
      clearInterval(checkConsentInterval);
    };
  }, [setConsent]);

  return <>{children}</>;
}

// Export utility functions for programmatic control
export const cookieConsentApi = {
  showPreferences: () => CookieConsent.showPreferences(),
  hidePreferences: () => CookieConsent.hidePreferences(),
  acceptCategory: (categories: string | string[]) => CookieConsent.acceptCategory(categories),
  acceptedCategory: (category: string) => CookieConsent.acceptedCategory(category),
  getCookie: () => CookieConsent.getCookie(),
  eraseCookies: (cookies: string[]) => CookieConsent.eraseCookies(cookies),
  validConsent: () => CookieConsent.validConsent(),
};
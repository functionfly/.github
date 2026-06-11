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
    CookieConsent.run(cookieConsentConfig).then(() => {
      document.documentElement.classList.remove('cc--darkmode');

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
          updateGoogleConsentFromCategories();
        }
      }
    });

    const checkConsentInterval = setInterval(() => {
      document.documentElement.classList.remove('cc--darkmode');
    }, 100);

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
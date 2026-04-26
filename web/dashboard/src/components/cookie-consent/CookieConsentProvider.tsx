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
      // Remove cc--darkmode after library initializes
      document.documentElement.classList.remove('cc--darkmode');

      // Force light mode by injecting override styles
      const isLight = document.documentElement.getAttribute('data-theme') === 'light';
      if (isLight) {
        // Remove any existing override
        const existing = document.getElementById('cc-light-override');
        if (existing) existing.remove();

        const styleEl = document.createElement('style');
        styleEl.id = 'cc-light-override';
        styleEl.textContent = `
          #cc-main,
          #cc-main .cm,
          #cc-main .pm,
          #cc-main .cm-wrapper,
          #cc-main .pm-wrapper {
            background: #FFFFFF !important;
            --cc-bg: #FFFFFF !important;
          }
          #cc-main .cm,
          #cc-main .pm {
            background-color: #FFFFFF !important;
            border: 1px solid rgba(0,0,0,0.08) !important;
          }
          #cc-main .pm__section,
          #cc-main .pm__section--toggle .pm__section-title {
            background: #F6F8FA !important;
            background-color: #F6F8FA !important;
            border-color: rgba(0,0,0,0.08) !important;
          }
          #cc-main .cm__title,
          #cc-main .pm__title {
            color: #0D1117 !important;
          }
          #cc-main .cm__desc,
          #cc-main .pm__section-desc {
            color: #6B7280 !important;
          }
        `;
        document.head.appendChild(styleEl);
      }

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
    const checkConsentInterval = setInterval(() => {
      document.documentElement.classList.remove('cc--darkmode');
    }, 100);

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
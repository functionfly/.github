'use client';

import { cookieConsentApi } from './CookieConsentProvider';

/**
 * CookieBanner Component
 *
 * This component serves as a wrapper around the vanilla-cookieconsent banner.
 * The actual banner UI is rendered by the vanilla-cookieconsent library,
 * but this component provides programmatic control and state management.
 */
export function CookieBanner() {
  // The banner is rendered by vanilla-cookieconsent, so we don't need to render anything here
  // This component exists for future customization or additional logic
  return null;
}

// Export utility functions for banner control
export const bannerApi = {
  show: () => cookieConsentApi.showPreferences(),
  hide: () => cookieConsentApi.hidePreferences(),
  acceptAll: () => cookieConsentApi.acceptCategory('all'),
  acceptNecessary: () => cookieConsentApi.acceptCategory([]),
};
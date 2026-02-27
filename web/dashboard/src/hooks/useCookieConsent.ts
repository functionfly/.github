import { useCookieConsentStore } from '@/stores/cookieConsentStore';
import { cookieConsentApi } from '@/components/cookie-consent';

/**
 * Hook for accessing cookie consent state and utilities
 */
export function useCookieConsent() {
  const store = useCookieConsentStore();

  return {
    // State
    hasConsent: store.hasConsent,
    consentTimestamp: store.consentTimestamp,
    categories: store.categories,

    // Category-specific checks
    hasAnalyticsConsent: store.categories.analytics,
    hasMarketingConsent: store.categories.marketing,
    hasFunctionalityConsent: store.categories.functionality,
    hasNecessaryConsent: store.categories.necessary, // Always true

    // Actions
    setConsent: store.setConsent,
    resetConsent: store.resetConsent,
    updateCategory: store.updateCategory,

    // Utility functions
    canUseAnalytics: () => store.categories.analytics,
    canUseMarketing: () => store.categories.marketing,
    canUseFunctionality: () => store.categories.functionality,
    canUseCategory: (category: keyof typeof store.categories) => store.categories[category],

    // API functions
    showPreferences: cookieConsentApi.showPreferences,
    hidePreferences: cookieConsentApi.hidePreferences,
    acceptAll: () => cookieConsentApi.acceptCategory('all'),
    acceptNecessary: () => cookieConsentApi.acceptCategory([]),
    getCookie: cookieConsentApi.getCookie,
    eraseCookies: cookieConsentApi.eraseCookies,

    // Helper functions for conditional loading
    shouldLoadAnalytics: () => store.hasConsent && store.categories.analytics,
    shouldLoadMarketing: () => store.hasConsent && store.categories.marketing,
    shouldLoadFunctionality: () => store.hasConsent && store.categories.functionality,
  };
}
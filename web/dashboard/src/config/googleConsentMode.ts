import { useCookieConsentStore } from '@/stores/cookieConsentStore';

// Extend the global Window interface to include gtag
declare global {
  interface Window {
    gtag: (...args: any[]) => void;
    dataLayer: any[];
  }
}

export interface ConsentState {
  ad_storage: 'granted' | 'denied';
  ad_user_data: 'granted' | 'denied';
  ad_personalization: 'granted' | 'denied';
  analytics_storage: 'granted' | 'denied';
  functionality_storage: 'granted' | 'denied';
  personalization_storage: 'granted' | 'denied';
  security_storage: 'granted' | 'denied';
}

/**
 * Initialize Google Consent Mode v2 with default denied state
 */
export function initializeGoogleConsentMode() {
  // Set default consent state (all denied)
  const defaultConsent: ConsentState = {
    ad_storage: 'denied',
    ad_user_data: 'denied',
    ad_personalization: 'denied',
    analytics_storage: 'denied',
    functionality_storage: 'denied',
    personalization_storage: 'denied',
    security_storage: 'granted', // Security storage is always granted
  };

  updateGoogleConsent(defaultConsent);
}

/**
 * Update Google Consent Mode based on cookie consent categories
 */
export function updateGoogleConsentFromCategories() {
  const { categories } = useCookieConsentStore.getState();

  const consentState: ConsentState = {
    ad_storage: categories.marketing ? 'granted' : 'denied',
    ad_user_data: categories.marketing ? 'granted' : 'denied',
    ad_personalization: categories.marketing ? 'granted' : 'denied',
    analytics_storage: categories.analytics ? 'granted' : 'denied',
    functionality_storage: categories.functionality ? 'granted' : 'denied',
    personalization_storage: categories.functionality ? 'granted' : 'denied',
    security_storage: 'granted', // Always granted for security
  };

  updateGoogleConsent(consentState);
}

/**
 * Update Google Consent Mode with specific consent state
 */
export function updateGoogleConsent(consentState: ConsentState) {
  if (typeof window !== 'undefined' && window.gtag) {
    window.gtag('consent', 'update', consentState);
  }
}

/**
 * Set default Google Consent Mode (called before gtag initialization)
 */
export function setDefaultGoogleConsent() {
  if (typeof window !== 'undefined') {
    // Initialize dataLayer if not exists
    window.dataLayer = window.dataLayer || [];

    // Set default consent (all denied except security)
    const defaultConsent: ConsentState = {
      ad_storage: 'denied',
      ad_user_data: 'denied',
      ad_personalization: 'denied',
      analytics_storage: 'denied',
      functionality_storage: 'denied',
      personalization_storage: 'denied',
      security_storage: 'granted',
    };

    window.gtag = function() {
      window.dataLayer.push(arguments);
    };

    window.gtag('consent', 'default', defaultConsent);
  }
}

/**
 * Initialize Google Analytics with Consent Mode
 */
export function initializeGoogleAnalytics(measurementId: string) {
  if (typeof window !== 'undefined') {
    // Load Google Tag Manager script
    const script = document.createElement('script');
    script.async = true;
    script.src = `https://www.googletagmanager.com/gtag/js?id=${measurementId}`;
    document.head.appendChild(script);

    // Initialize gtag
    window.dataLayer = window.dataLayer || [];
    window.gtag = function() {
      window.dataLayer.push(arguments);
    };

    // Set default consent first
    setDefaultGoogleConsent();

    // Configure Google Analytics
    window.gtag('js', new Date());
    window.gtag('config', measurementId, {
      anonymize_ip: true,
      allow_google_signals: false,
      allow_ad_personalization_signals: false,
    });
  }
}

/**
 * Hook to automatically update Google Consent Mode when cookie preferences change
 */
export function useGoogleConsentModeSync() {
  const { categories } = useCookieConsentStore();

  // Update consent mode whenever categories change
  React.useEffect(() => {
    updateGoogleConsentFromCategories();
  }, [categories.analytics, categories.marketing, categories.functionality]);
}

// Re-export React for the hook
import React from 'react';

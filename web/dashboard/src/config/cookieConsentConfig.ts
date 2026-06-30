import type { CookieConsentConfig } from 'vanilla-cookieconsent';

export const cookieConsentConfig: CookieConsentConfig = {
  // Core settings
  root: null,
  autoShow: true,
  disablePageInteraction: false,
  hideFromBots: true,
  mode: 'opt-in', // GDPR compliant - require explicit consent

  // Cookie settings
  cookie: {
    name: 'functionfly_cookie_consent',
    path: '/',
    secure: window.location.protocol === 'https:',
    expiresAfterDays: 365,
  },

  // Categories configuration
  categories: {
    necessary: {
      enabled: true,
      readOnly: true,
    },
    analytics: {
      enabled: false,
      autoClear: {
        cookies: [
          { name: /^_ga/ }, // Google Analytics cookies
          { name: '_gid' },
          { name: /^_gat/ },
          { name: '_gcl_au' },
        ],
      },
    },
    marketing: {
      enabled: false,
      autoClear: {
        cookies: [
          { name: /^_fbp/ }, // Facebook Pixel
          { name: /^_fbc/ },
          { name: /^_gcl/ }, // Google Click ID
        ],
      },
    },
    functionality: {
      enabled: false,
      autoClear: {
        cookies: [
          { name: 'language_preference' },
          { name: 'theme_preference' },
          { name: 'ui_customization' },
        ],
      },
    },
  },

  // Language configuration
  language: {
    default: 'en',
    translations: {
      en: {
        consentModal: {
          title: 'We value your privacy',
          description: 'We use cookies to enhance your browsing experience, serve personalized content, and analyze our traffic. By clicking "Accept All", you consent to our use of cookies.',
          acceptAllBtn: 'Accept All',
          acceptNecessaryBtn: 'Reject All',
          showPreferencesBtn: 'Manage Preferences',
          closeIconLabel: 'Close window',
        },
        preferencesModal: {
          title: 'Cookie Preferences',
          acceptAllBtn: 'Accept All',
          acceptNecessaryBtn: 'Reject All',
          savePreferencesBtn: 'Save Preferences',
          closeIconLabel: 'Close window',
          serviceCounterLabel: 'Services',
          sections: [
            {
              title: 'Cookie Usage',
              description: 'We use cookies to ensure the basic functionalities of the website and to enhance your online experience.',
            },
            {
              title: 'Strictly Necessary Cookies',
              description: 'These cookies are essential for the proper functioning of our website. Without these cookies, the website would not work properly.',
              linkedCategory: 'necessary',
            },
            {
              title: 'Analytics Cookies',
              description: 'These cookies help us understand how visitors interact with our website by collecting and reporting information anonymously.',
              linkedCategory: 'analytics',
              cookieTable: {
                caption: 'Analytics Cookies',
                headers: {
                  name: 'Cookie',
                  domain: 'Domain',
                  description: 'Description',
                  expiration: 'Expiration',
                },
                body: [
                  {
                    name: '_ga',
                    domain: 'google-analytics.com',
                    description: 'Google Analytics tracking cookie',
                    expiration: '2 years',
                  },
                  {
                    name: '_gid',
                    domain: 'google-analytics.com',
                    description: 'Google Analytics session cookie',
                    expiration: '24 hours',
                  },
                ],
              },
            },
            {
              title: 'Marketing Cookies',
              description: 'These cookies are used to track visitors across websites to display ads that are relevant and engaging for individual users.',
              linkedCategory: 'marketing',
            },
            {
              title: 'Functionality Cookies',
              description: 'These cookies enable the website to remember choices you make and provide enhanced, more personal features.',
              linkedCategory: 'functionality',
            },
            {
              title: 'More information',
              description: 'For any queries in relation to our policy on cookies and your choices, please contact us.',
            },
          ],
        },
      },
    },
  },

  // GUI options - Aviation cockpit-inspired layout
  guiOptions: {
    consentModal: {
      layout: 'box wide',
      position: 'bottom right',
      equalWeightButtons: true,
      flipButtons: false,
    },
    preferencesModal: {
      layout: 'box',
      position: 'right',
      equalWeightButtons: true,
      flipButtons: false,
    },
  },

  // Callbacks
  onFirstConsent: () => {
    console.log('First consent given');
  },

  onConsent: (cookie) => {
    console.log('Consent updated:', cookie);
  },

  onChange: ({ changedCategories }) => {
    console.log('Cookie preferences changed:', changedCategories);
  },

  onModalReady: () => {
    console.log('Cookie consent modal ready');
  },

  onModalShow: () => {
    console.log('Cookie consent modal shown');
  },

  onModalHide: () => {
    console.log('Cookie consent modal hidden');
  },
};
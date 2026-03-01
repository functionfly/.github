import i18n from 'i18next'
import { initReactI18next } from 'react-i18next'

const defaultNs = 'common'

void i18n.use(initReactI18next).init({
  lng: 'en',
  fallbackLng: 'en',
  defaultNS: defaultNs,
  ns: [defaultNs],
  resources: {
    en: {
      [defaultNs]: {
        // Add keys here, e.g. "welcome": "Welcome"
      },
    },
  },
  interpolation: {
    escapeValue: false,
  },
})

export default i18n

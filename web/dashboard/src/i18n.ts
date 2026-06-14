import i18n from 'i18next'
import { initReactI18next } from 'react-i18next'
import { detectLanguage, RTL_LANGUAGES } from '@/lib/i18n/languages'

const defaultNs = 'common'

const localeModules = import.meta.glob('./locales/*/common.json', {
  eager: true,
}) as Record<string, { default: Record<string, unknown> }>

const onboardingModules = import.meta.glob('./locales/*/onboarding.json', {
  eager: true,
}) as Record<string, { default: Record<string, unknown> }>

function buildResources() {
  const resources: Record<string, Record<string, unknown>> = {}
  for (const [path, mod] of Object.entries(localeModules)) {
    const match = path.match(/\.\/locales\/(\w+)\/common\.json/)
    if (match) {
      const lng = match[1]
      resources[lng] = { [defaultNs]: mod.default }
    }
  }
  for (const [path, mod] of Object.entries(onboardingModules)) {
    const match = path.match(/\.\/locales\/(\w+)\/onboarding\.json/)
    if (match) {
      const lng = match[1]
      if (!resources[lng]) {
        resources[lng] = {}
      }
      resources[lng] = { ...resources[lng], onboarding: mod.default }
    }
  }
  return resources
}

const resources = buildResources()

const persistedLang =
  typeof localStorage !== 'undefined'
    ? localStorage.getItem('ff-language')
    : null

const initialLang = persistedLang ?? detectLanguage()

void i18n.use(initReactI18next).init({
  lng: initialLang,
  fallbackLng: 'en',
  defaultNS: defaultNs,
  ns: [defaultNs, 'onboarding'],
  resources,
  interpolation: {
    escapeValue: false,
  },
  react: {
    useSuspense: true,
  },
})

// Apply RTL/LTR direction on language change
function applyDirection(lng: string) {
  const isRtl = RTL_LANGUAGES.has(lng)
  document.documentElement.dir = isRtl ? 'rtl' : 'ltr'
  document.documentElement.lang = lng
}

applyDirection(i18n.language)
i18n.on('languageChanged', applyDirection)

export default i18n

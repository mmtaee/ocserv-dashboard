import { createI18n } from 'vue-i18n'

import en from '@shared/locales/en.json'
import fa from '@shared/locales/fa.json'
import { syncDocumentLocale } from '@shared/utils/appDirection'
import { resolveInitialLocale } from '@shared/utils/initialLocale'

const initialLocale = resolveInitialLocale()

syncDocumentLocale(initialLocale)

export default createI18n({
  legacy: false,
  locale: initialLocale,
  fallbackLocale: 'en',
  messages: { en, fa },
})

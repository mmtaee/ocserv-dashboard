import { createI18n } from 'vue-i18n'
import en from '../i18n/locales/en.json'
import fa from '../i18n/locales/fa.json'

const i18n = createI18n({
  legacy: false,
  locale: 'fa',
  fallbackLocale: 'en',
  messages: {
    en,
    fa,
  },
})

export default i18n

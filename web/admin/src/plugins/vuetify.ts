import '@mdi/font/css/materialdesignicons.css'
import 'vuetify/styles'

import { aliases, mdi } from 'vuetify/iconsets/mdi'
import { createVuetify } from 'vuetify'
import { en, fa } from 'vuetify/locale'

import { resolveInitialLocale } from '@shared/utils/initialLocale'

export default createVuetify({
  icons: {
    defaultSet: 'mdi',
    aliases,
    sets: { mdi },
  },
  locale: {
    locale: resolveInitialLocale(),
    fallback: 'en',
    messages: { en, fa },
    rtl: { en: false, fa: true },
  },
  theme: {
    defaultTheme: 'adminTheme',
    themes: {
      adminTheme: {
        dark: false,
        colors: {
          background: '#F7F8FA',
          surface: '#FFFFFF',
          primary: '#1867C0',
          secondary: '#5C6BC0',
        },
      },
    },
  },
})

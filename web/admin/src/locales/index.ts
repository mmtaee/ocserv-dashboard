import { createI18n } from 'vue-i18n'

import ar from '@/locales/ar'
import en from '@/locales/en'
import fa from '@/locales/fa'
import it from '@/locales/it'
import ru from '@/locales/ru'
import zhCn from '@/locales/zh-cn'
import zhTw from '@/locales/zh-tw'

const LANGUAGE_STORAGE_KEY = 'ocserv-dashboard.language'
const DEFAULT_LANGUAGE_CONFIG = 'en:English,it:Italiano,zh-cn:中文(简体),zh-tw:中文(繁體),ru:Русский,fa:فارسی,ar:العربية'
const rtlLocales = new Set(['fa', 'ar'])

export const messages = { en, it, 'zh-cn': zhCn, 'zh-tw': zhTw, ru, fa, ar } as const
export type AppLocale = keyof typeof messages

export interface LanguageOption {
  code: AppLocale
  label: string
  direction: 'ltr' | 'rtl'
}

function isAppLocale(value: string): value is AppLocale {
  return Object.hasOwn(messages, value)
}

function parseLanguageOptions(value: string): LanguageOption[] {
  const options = value.split(',').flatMap((entry) => {
    const separator = entry.indexOf(':')
    if (separator < 1) return []
    const code = entry.slice(0, separator).trim().toLowerCase()
    const label = entry.slice(separator + 1).trim()
    if (!label || !isAppLocale(code)) return []
    return [{ code, label, direction: rtlLocales.has(code) ? 'rtl' as const : 'ltr' as const }]
  })

  return options.length ? options : parseLanguageOptions(DEFAULT_LANGUAGE_CONFIG)
}

export const languageOptions = parseLanguageOptions(
  import.meta.env.VITE_I18N_LANGUAGES || DEFAULT_LANGUAGE_CONFIG,
)

function matchLocale(value: string | null | undefined): AppLocale | null {
  if (!value) return null
  const normalized = value.toLowerCase()
  if (isAppLocale(normalized)) return normalized
  const base = normalized.split('-')[0]
  if (base === 'zh') return 'zh-cn'
  return base && isAppLocale(base) ? base : null
}

function resolveInitialLocale(): AppLocale {
  const saved = typeof window === 'undefined' ? null : matchLocale(window.localStorage.getItem(LANGUAGE_STORAGE_KEY))
  if (saved && languageOptions.some(({ code }) => code === saved)) return saved

  if (typeof navigator !== 'undefined') {
    for (const language of navigator.languages) {
      const matched = matchLocale(language)
      if (matched && languageOptions.some(({ code }) => code === matched)) return matched
    }
  }

  return languageOptions[0]?.code ?? 'en'
}

const initialLocale = resolveInitialLocale()

export const i18n = createI18n({
  legacy: false,
  locale: initialLocale,
  fallbackLocale: 'en',
  messages,
})

export function isRtlLocale(locale: string): boolean {
  return rtlLocales.has(locale.toLowerCase())
}

export function setLocale(locale: string): void {
  if (!isAppLocale(locale) || !languageOptions.some(({ code }) => code === locale)) return
  i18n.global.locale.value = locale

  if (typeof window !== 'undefined') window.localStorage.setItem(LANGUAGE_STORAGE_KEY, locale)
  if (typeof document !== 'undefined') {
    document.documentElement.lang = locale
    document.documentElement.dir = isRtlLocale(locale) ? 'rtl' : 'ltr'
    document.title = i18n.global.t('common.appName')
  }
}

setLocale(initialLocale)

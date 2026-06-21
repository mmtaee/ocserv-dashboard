/* eslint-disable react-refresh/only-export-components */
import {
  type ReactNode,
  createContext,
  useContext,
  useEffect,
  useState,
} from "react"

import en from "./locales/en.json"
import es from "./locales/es.json"
import fa from "./locales/fa.json"
import fr from "./locales/fr.json"
import type { Locale, LocaleConfig, TranslationSchema } from "./types"

const locales: Record<Locale, TranslationSchema> = {
  en,
  fa,
  es,
  fr,
}

const localeConfigs: Record<Locale, LocaleConfig> = {
  en: { name: "English", direction: "ltr" },
  fa: { name: "فارسی", direction: "rtl" },
  es: { name: "Español", direction: "ltr" },
  fr: { name: "Français", direction: "ltr" },
}

const LOCALE_STORAGE_KEY = "ocserv-locale"
const DEFAULT_LOCALE: Locale = "en"

type I18nContextType = {
  locale: Locale
  t: TranslationSchema
  setLocale: (locale: Locale) => void
  availableLocales: Locale[]
  localeConfig: LocaleConfig
  locales: Record<Locale, TranslationSchema>
  localeConfigs: Record<Locale, LocaleConfig>
}

const I18nContext = createContext<I18nContextType | undefined>(undefined)

type I18nProviderProps = {
  children: ReactNode
}

export function I18nProvider({ children }: I18nProviderProps) {
  const [locale, setLocaleState] = useState<Locale>(() => {
    const saved = localStorage.getItem(LOCALE_STORAGE_KEY)
    return (saved as Locale) || DEFAULT_LOCALE
  })

  const setLocale = (newLocale: Locale) => {
    setLocaleState(newLocale)
    localStorage.setItem(LOCALE_STORAGE_KEY, newLocale)
  }

  useEffect(() => {
    document.documentElement.dir = localeConfigs[locale].direction
    document.documentElement.lang = locale
  }, [locale])

  const contextValue: I18nContextType = {
    locale,
    t: locales[locale],
    setLocale,
    availableLocales: Object.keys(locales) as Locale[],
    localeConfig: localeConfigs[locale],
    locales,
    localeConfigs,
  }

  return (
    <I18nContext.Provider value={contextValue}>{children}</I18nContext.Provider>
  )
}

export function useI18n() {
  const context = useContext(I18nContext)
  if (context === undefined) {
    throw new Error("useI18n must be used within an I18nProvider")
  }
  return context
}

export { locales, localeConfigs }

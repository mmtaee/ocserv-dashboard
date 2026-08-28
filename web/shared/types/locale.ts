export const supportedLocales = ['en', 'fa'] as const

export type SupportedLocale = (typeof supportedLocales)[number]
export type AppDirection = 'ltr' | 'rtl'

export const defaultLocale: SupportedLocale = 'en'

export function isSupportedLocale(locale: string | null | undefined): locale is SupportedLocale {
  return supportedLocales.includes(locale as SupportedLocale)
}

export function normalizeLocale(locale: string | null | undefined): SupportedLocale {
  if (!locale) return defaultLocale

  const language = locale.toLowerCase().split('-')[0]
  return isSupportedLocale(language) ? language : defaultLocale
}

import {
  defaultLocale,
  isSupportedLocale,
  normalizeLocale,
  type SupportedLocale,
} from '@shared/types/locale'

export const localeStorageKey = 'ocserv-dashboard.locale'

export function resolveInitialLocale(): SupportedLocale {
  if (typeof window === 'undefined') return defaultLocale

  const storedLocale = window.localStorage.getItem(localeStorageKey)
  if (isSupportedLocale(storedLocale)) return storedLocale

  return normalizeLocale(window.navigator.language)
}

export function persistLocale(locale: SupportedLocale): void {
  if (typeof window === 'undefined') return

  window.localStorage.setItem(localeStorageKey, locale)
}

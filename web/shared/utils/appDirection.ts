import type { AppDirection, SupportedLocale } from '@shared/types/locale'

export function getAppDirection(locale: SupportedLocale): AppDirection {
  return locale === 'fa' ? 'rtl' : 'ltr'
}

export function syncDocumentLocale(locale: SupportedLocale): void {
  if (typeof document === 'undefined') return

  document.documentElement.lang = locale
  document.documentElement.dir = getAppDirection(locale)
}

import { computed, watch } from 'vue'
import { useI18n } from 'vue-i18n'

import { getAppDirection, syncDocumentLocale } from '@shared/utils/appDirection'
import { persistLocale } from '@shared/utils/initialLocale'
import { normalizeLocale, type SupportedLocale } from '@shared/types/locale'

export function useAppLocale() {
  const { locale } = useI18n({ useScope: 'global' })

  const activeLocale = computed<SupportedLocale>({
    get: () => normalizeLocale(locale.value),
    set: (value) => {
      locale.value = value
    },
  })

  const direction = computed(() => getAppDirection(activeLocale.value))
  const isRtl = computed(() => direction.value === 'rtl')

  watch(
    activeLocale,
    (value) => {
      syncDocumentLocale(value)
      persistLocale(value)
    },
    { immediate: true },
  )

  function setLocale(value: SupportedLocale): void {
    activeLocale.value = value
  }

  return {
    activeLocale,
    direction,
    isRtl,
    setLocale,
  }
}

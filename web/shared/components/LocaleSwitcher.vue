<script setup lang="ts">
import { computed } from 'vue'
import { useI18n } from 'vue-i18n'

import { useAppLocale } from '@shared/composables/useAppLocale'
import type { SupportedLocale } from '@shared/types/locale'

const { t } = useI18n()
const { activeLocale } = useAppLocale()

const localeItems = computed(() => [
  { title: t('common.locales.fa'), value: 'fa' as const },
  { title: t('common.locales.en'), value: 'en' as const },
])

function selectLocale(value: SupportedLocale | null): void {
  if (value) activeLocale.value = value
}
</script>

<template>
  <v-select
    :aria-label="t('common.language')"
    class="locale-switcher"
    density="compact"
    hide-details
    item-title="title"
    item-value="value"
    :items="localeItems"
    :model-value="activeLocale"
    prepend-inner-icon="mdi-translate"
    variant="outlined"
    @update:model-value="selectLocale"
  />
</template>

<style scoped>
.locale-switcher {
  flex: 0 1 10rem;
  min-inline-size: 8.5rem;
}
</style>

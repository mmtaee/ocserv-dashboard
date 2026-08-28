<script setup lang="ts">
import { computed } from 'vue'
import { useI18n } from 'vue-i18n'
import { useRoute } from 'vue-router'

import LocaleSwitcher from '@shared/components/LocaleSwitcher.vue'

const emit = defineEmits<{
  toggleNavigation: []
}>()

const route = useRoute()
const { t } = useI18n()

const pageTitle = computed(() => {
  const titleKey = route.meta.titleKey
  return typeof titleKey === 'string' ? t(titleKey) : t('admin.title')
})
</script>

<template>
  <v-app-bar flat>
    <v-app-bar-nav-icon :aria-label="t('admin.title')" @click="emit('toggleNavigation')" />
    <v-app-bar-title>{{ pageTitle }}</v-app-bar-title>
    <LocaleSwitcher class="me-4" />
  </v-app-bar>
</template>

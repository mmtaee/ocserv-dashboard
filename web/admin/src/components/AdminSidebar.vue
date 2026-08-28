<script setup lang="ts">
import { useI18n } from 'vue-i18n'

import { adminNavigation } from '@/constants/navigation'

const model = defineModel<boolean>({ required: true })
const { t } = useI18n()

defineProps<{
  temporary: boolean
}>()
</script>

<template>
  <v-navigation-drawer v-model="model" location="start" :temporary="temporary">
    <v-list-item :title="t('app.name')" prepend-icon="mdi-shield-account-outline" />
    <v-divider />
    <v-list nav>
      <v-list-item
        v-for="item in adminNavigation"
        :key="item.titleKey"
        :prepend-icon="item.icon"
        :title="t(item.titleKey)"
        :to="item.to"
      />
    </v-list>
  </v-navigation-drawer>
</template>

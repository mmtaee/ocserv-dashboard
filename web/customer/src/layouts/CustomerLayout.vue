<script setup lang="ts">
import { useI18n } from 'vue-i18n'

import LocaleSwitcher from '@shared/components/LocaleSwitcher.vue'

const { t } = useI18n()
</script>

<template>
  <v-layout class="customer-layout">
    <v-app-bar flat>
      <v-app-bar-title>{{ t('customer.title') }}</v-app-bar-title>
      <LocaleSwitcher class="me-4" />
    </v-app-bar>
    <v-main>
      <v-container>
        <router-view />
      </v-container>
    </v-main>
  </v-layout>
</template>

<style scoped>
.customer-layout {
  min-height: 100vh;
}
</style>

<script setup lang="ts">
import { useI18n } from 'vue-i18n'

const { t } = useI18n()
</script>

<template>
  <v-card>
    <v-card-title class="d-flex align-center ga-2">
      <v-icon icon="mdi-account-circle-outline" />
      {{ t('customer.home.title') }}
    </v-card-title>
    <v-card-text>{{ t('customer.home.description') }}</v-card-text>
  </v-card>
</template>

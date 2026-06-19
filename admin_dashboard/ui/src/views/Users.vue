<template>
  <v-container>
    <h1 class="text-h4">{{ t('users') }}</h1>
  </v-container>
</template>

<script setup lang="ts">
import { useI18n } from 'vue-i18n'
const { t } = useI18n()
</script>

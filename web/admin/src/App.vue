<script setup lang="ts">
import { useAppLocale } from '@shared/composables/useAppLocale'

const { activeLocale, isRtl } = useAppLocale()
</script>

<template>
  <v-locale-provider :locale="activeLocale" :rtl="isRtl">
    <v-app>
      <router-view />
    </v-app>
  </v-locale-provider>
</template>

<script setup lang="ts">
import { computed } from "vue";
import { ConfigProvider } from "reka-ui";
import { useI18n } from "vue-i18n";

import { isRtlLocale } from "@/locales";

const { locale } = useI18n({ useScope: "global" });
const direction = computed(() => (isRtlLocale(locale.value) ? "rtl" : "ltr"));
</script>

<template>
  <ConfigProvider :dir="direction">
    <RouterView />
  </ConfigProvider>
</template>

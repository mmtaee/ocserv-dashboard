<script setup lang="ts">
import { computed } from "vue";
import { ConfigProvider } from "reka-ui";
import { useI18n } from "vue-i18n";

// import AppFooter from '@/components/AppFooter.vue'
import { isRtlLocale } from "@/locales";
import AppFooter from "@/components/AppFooter.vue";
import { useServerStore } from "@/stores/server";

const { locale } = useI18n({ useScope: "global" });
const direction = computed(() => (isRtlLocale(locale.value) ? "rtl" : "ltr"));
const server = useServerStore();
</script>

<template>
  <ConfigProvider :dir="direction">
    <div class="flex min-h-svh flex-col">
      <div class="flex min-h-0 flex-1 flex-col">
        <RouterView v-slot="{ Component }">
          <component
            :is="Component"
            :key="server.revision"
            class="min-h-0! flex-1"
          />
        </RouterView>
      </div>
      <AppFooter />
    </div>
  </ConfigProvider>
</template>

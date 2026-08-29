<script setup lang="ts">
import { computed } from "vue";
import { useI18n } from "vue-i18n";
import { useRoute } from "vue-router";

import DashboardLayout from "@/components/DashboardLayout.vue";

const route = useRoute();
const { t } = useI18n({ useScope: "global" });
const routeName = computed(() =>
  t(String(route.meta.titleKey ?? "navigation.dashboard")),
);
</script>

<template>
  <DashboardLayout>
    <h1 class="text-2xl font-semibold tracking-tight">{{ routeName }}</h1>
  </DashboardLayout>
</template>

<script setup lang="ts">
import { useI18n } from "vue-i18n";

import DashboardLayout from "@/components/DashboardLayout.vue";
import DashboardHomeOverview from "@/components/dashboard/DashboardHomeOverview.vue";
import DashboardHomeStatistics from "@/components/dashboard/DashboardHomeStatistics.vue";
import OcservStatistics from "@/components/dashboard/OcservStatistics.vue";
import SystemUsage from "@/components/dashboard/SystemUsage.vue";
import { useDashboardStats } from "@/composables/useDashboardStats";

const { t } = useI18n({ useScope: "global" });
const {
  homeOverview,
  systemStats,
  containerStats,
  ocservStats,
  homeLoading,
  systemLoading,
  ocservLoading,
  homeError,
  systemError,
  containerError,
  ocservError,
} = useDashboardStats();
</script>

<template>
  <DashboardLayout>
    <div class="flex flex-col gap-6">
      <h1 class="text-2xl font-semibold tracking-tight">
        {{ t("dashboard.title") }}
      </h1>
      <DashboardHomeOverview
        :overview="homeOverview"
        :loading="homeLoading"
        :error="homeError"
      />
      <SystemUsage
        :stats="systemStats"
        :container-stats="containerStats"
        :loading="systemLoading"
        :error="systemError"
        :container-error="containerError"
      />
      <OcservStatistics
        :stats="ocservStats"
        :loading="ocservLoading"
        :error="ocservError"
      />
      <DashboardHomeStatistics
        :overview="homeOverview"
        :loading="homeLoading"
      />
    </div>
  </DashboardLayout>
</template>

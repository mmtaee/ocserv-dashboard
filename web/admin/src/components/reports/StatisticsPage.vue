<script setup lang="ts">
import { Activity, LockKeyhole, Radio, UserX } from "@lucide/vue";
import { onMounted, reactive, shallowRef } from "vue";
import { useI18n } from "vue-i18n";

import { normalizeApiError } from "@/api/http";
import {
  getReportStatistics,
  getReportUserSummary,
  type DailyTraffic,
  type UserReportSummary,
} from "@/api/services/reports";
import TrafficChart from "@/components/dashboard/TrafficChart.vue";
import ReportsDateRangeFilter from "@/components/reports/ReportsDateRangeFilter.vue";
import { defaultReportRange } from "@/components/reports/report-dates";
import { Alert, AlertDescription, AlertTitle } from "@/components/ui/alert";
import { Card, CardContent, CardHeader, CardTitle } from "@/components/ui/card";
import { Skeleton } from "@/components/ui/skeleton";

const { t } = useI18n({ useScope: "global" });
const range = reactive(defaultReportRange());
const traffic = shallowRef<DailyTraffic[]>([]);
const summary = shallowRef<UserReportSummary | null>(null);
const loading = shallowRef(true);
const error = shallowRef("");
const metrics = [
  ["active", Activity],
  ["online", Radio],
  ["locked", LockKeyhole],
  ["deactivated", UserX],
] as const;

async function load(): Promise<void> {
  loading.value = true;
  error.value = "";
  try {
    [traffic.value, summary.value] = await Promise.all([
      getReportStatistics(range),
      getReportUserSummary(),
    ]);
  } catch (cause) {
    error.value = normalizeApiError(cause).message;
  } finally {
    loading.value = false;
  }
}

onMounted(load);
</script>

<template>
  <div class="flex flex-col gap-6">
    <div>
      <h1 class="text-2xl font-semibold tracking-tight">
        {{ t("reports.statisticsTitle") }}
      </h1>
      <p class="text-sm text-muted-foreground">
        {{ t("reports.statisticsDescription") }}
      </p>
    </div>
    <ReportsDateRangeFilter
      v-model:date-start="range.dateStart"
      v-model:date-end="range.dateEnd"
      :loading="loading"
      @apply="load"
    />
    <Alert v-if="error" variant="error"
      ><AlertTitle>{{ t("reports.requestFailed") }}</AlertTitle
      ><AlertDescription>{{ error }}</AlertDescription></Alert
    >
    <div class="grid gap-4 sm:grid-cols-2 xl:grid-cols-4">
      <Card v-for="[key, icon] in metrics" :key="key"
        ><CardHeader class="flex-row items-center justify-between"
          ><CardTitle class="text-sm">{{
            t(`reports.metrics.${key}`)
          }}</CardTitle
          ><component
            :is="icon"
            class="size-4 text-muted-foreground" /></CardHeader
        ><CardContent
          ><Skeleton v-if="loading && !summary" class="h-8 w-20" />
          <div v-else class="text-2xl font-semibold">
            {{ summary?.[key] ?? 0 }}
          </div></CardContent
        ></Card
      >
    </div>
    <TrafficChart
      :data="traffic"
      :loading="loading"
      :limit="0"
      title-key="reports.trafficTitle"
      description-key="reports.trafficDescription"
      empty-title-key="reports.noTraffic"
      empty-description-key="reports.noTrafficDescription"
    />
  </div>
</template>

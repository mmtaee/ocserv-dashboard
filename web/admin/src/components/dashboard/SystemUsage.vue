<script setup lang="ts">
import { CircleAlert } from "@lucide/vue";
import { computed } from "vue";
import { useI18n } from "vue-i18n";

import type { ContainerStats, SystemStats } from "@/api/services/dashboard";
import UsageGauge from "@/components/dashboard/UsageGauge.vue";
import { Alert, AlertDescription } from "@/components/ui/alert";
import {
  Card,
  CardContent,
  CardDescription,
  CardHeader,
  CardTitle,
} from "@/components/ui/card";
import { Skeleton } from "@/components/ui/skeleton";

const props = defineProps<{
  stats: SystemStats | null;
  containerStats: ContainerStats | null;
  loading: boolean;
  error: boolean;
  containerError: boolean;
}>();
const { t } = useI18n({ useScope: "global" });

const metrics = computed(() => [
  {
    key: "cpu",
    label: t("dashboard.cpuUsage"),
    percent: props.stats?.cpu?.avg_percent,
    used: props.stats?.cpu?.used_units,
    total: props.stats?.cpu?.total,
    unit: t("dashboard.units"),
  },
  {
    key: "ram",
    label: t("dashboard.ramUsage"),
    percent: props.stats?.ram?.used_percent,
    used: props.stats?.ram?.used,
    total: props.stats?.ram?.total,
    unit: t("dashboard.gigabytes"),
  },
  {
    key: "swap",
    label: t("dashboard.swapUsage"),
    percent: props.stats?.swap?.used_percent,
    used: props.stats?.swap?.used,
    total: props.stats?.swap?.total,
    unit: t("dashboard.gigabytes"),
  },
  {
    key: "disk",
    label: t("dashboard.diskUsage"),
    percent: props.stats?.disk?.used_percent,
    used: props.stats?.disk?.used,
    total: props.stats?.disk?.total,
    unit: t("dashboard.gigabytes"),
  },
]);

const hasContainerStats = computed(() => Boolean(props.containerStats?.name));
</script>

<template>
  <Card>
    <CardHeader>
      <CardTitle>{{ t("dashboard.systemUsage") }}</CardTitle>
      <CardDescription>{{
        t("dashboard.systemUsageDescription")
      }}</CardDescription>
    </CardHeader>
    <CardContent class="flex flex-col gap-6">
      <Alert v-if="error" variant="destructive">
        <CircleAlert />
        <AlertDescription>{{
          t("dashboard.systemUsageError")
        }}</AlertDescription>
      </Alert>
      <Alert v-if="containerError" variant="destructive">
        <CircleAlert />
        <AlertDescription>{{
          t("dashboard.containerUsageError")
        }}</AlertDescription>
      </Alert>
      <div
        v-if="loading && !stats"
        class="grid gap-8 sm:grid-cols-2 xl:grid-cols-4"
      >
        <div
          v-for="item in 4"
          :key="item"
          class="flex flex-col items-center gap-3"
        >
          <Skeleton class="size-28 rounded-full" />
          <Skeleton class="h-4 w-20" />
          <Skeleton class="h-3 w-28" />
        </div>
      </div>
      <div v-else class="grid gap-8 sm:grid-cols-2 xl:grid-cols-4">
        <UsageGauge
          v-for="metric in metrics"
          :key="metric.key"
          :label="metric.label"
          :percent="metric.percent"
          :used="metric.used"
          :total="metric.total"
          :unit="metric.unit"
        />
      </div>
      <section
        v-if="hasContainerStats"
        class="flex flex-col gap-4 border-t pt-6"
      >
        <div>
          <h3 class="font-semibold">{{ t("dashboard.containerUsage") }}</h3>
          <p class="text-sm text-muted-foreground">
            {{ t("dashboard.containerUsageDescription") }}
          </p>
        </div>
        <Card class="gap-4 py-5 shadow-none">
          <CardHeader class="items-center text-center">
            <CardTitle class="text-lg capitalize">
              {{ containerStats?.name }}
            </CardTitle>
          </CardHeader>
          <CardContent class="grid grid-cols-2 gap-4">
            <UsageGauge
              :label="t('dashboard.cpu')"
              :percent="containerStats?.cpu?.avg_percent"
              :used="containerStats?.cpu?.used_units"
              :total="containerStats?.cpu?.total"
              :unit="t('dashboard.units')"
              variant="liquid"
            />
            <UsageGauge
              :label="t('dashboard.ram')"
              :percent="containerStats?.ram?.used_percent"
              :used="containerStats?.ram?.used"
              :total="containerStats?.ram?.total"
              :unit="t('dashboard.gigabytes')"
              variant="liquid"
            />
          </CardContent>
        </Card>
      </section>
    </CardContent>
  </Card>
</template>

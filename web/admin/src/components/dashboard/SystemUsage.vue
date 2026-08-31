<script setup lang="ts">
import { CircleAlert } from "@lucide/vue";
import { computed } from "vue";
import { useI18n } from "vue-i18n";

import type { ContainerStats, SystemStats } from "@/api/services/dashboard";
import UsageGauge from "@/components/dashboard/UsageGauge.vue";
import { Alert, AlertDescription } from "@/components/ui/alert";
import { Badge } from "@/components/ui/badge";
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

const hasContainerStats = computed(() => Boolean(props.containerStats?.name));
const cardTitle = computed(() =>
  t(
    hasContainerStats.value
      ? "dashboard.systemAndContainerUsage"
      : "dashboard.systemUsage",
  ),
);
const cardDescription = computed(() =>
  t(
    hasContainerStats.value
      ? "dashboard.systemAndContainerUsageDescription"
      : "dashboard.systemUsageDescription",
  ),
);
const metrics = computed(() => {
  const systemMetrics = props.stats
    ? [
        {
          key: "system-cpu",
          label: t("dashboard.cpuUsage"),
          percent: props.stats.cpu?.avg_percent,
          used: props.stats.cpu?.used_units,
          total: props.stats.cpu?.total,
          unit: t("dashboard.units"),
          variant: "liquid" as const,
        },
        {
          key: "system-ram",
          label: t("dashboard.ramUsage"),
          percent: props.stats.ram?.used_percent,
          used: props.stats.ram?.used,
          total: props.stats.ram?.total,
          unit: t("dashboard.gigabytes"),
          variant: "liquid" as const,
        },
        {
          key: "system-swap",
          label: t("dashboard.swapUsage"),
          percent: props.stats.swap?.used_percent,
          used: props.stats.swap?.used,
          total: props.stats.swap?.total,
          unit: t("dashboard.gigabytes"),
          variant: "liquid" as const,
        },
        {
          key: "system-disk",
          label: t("dashboard.diskUsage"),
          percent: props.stats.disk?.used_percent,
          used: props.stats.disk?.used,
          total: props.stats.disk?.total,
          unit: t("dashboard.gigabytes"),
          variant: "liquid" as const,
        },
      ]
    : [];

  if (!hasContainerStats.value) return systemMetrics;

  return [
    ...systemMetrics,
    {
      key: "container-cpu",
      label: t("dashboard.containerCpuUsage"),
      percent: props.containerStats?.cpu?.avg_percent,
      used: props.containerStats?.cpu?.used_units,
      total: props.containerStats?.cpu?.total,
      unit: t("dashboard.units"),
      variant: "liquid" as const,
    },
    {
      key: "container-ram",
      label: t("dashboard.containerRamUsage"),
      percent: props.containerStats?.ram?.used_percent,
      used: props.containerStats?.ram?.used,
      total: props.containerStats?.ram?.total,
      unit: t("dashboard.gigabytes"),
      variant: "liquid" as const,
    },
  ];
});
</script>

<template>
  <Card>
    <CardHeader>
      <CardTitle class="flex flex-wrap items-center gap-2">
        <span>{{ cardTitle }}</span>
        <Badge v-if="hasContainerStats" variant="outline" dir="ltr">
          {{ containerStats?.name }}
        </Badge>
      </CardTitle>
      <CardDescription>
        {{ cardDescription }}
      </CardDescription>
    </CardHeader>
    <CardContent class="flex flex-col gap-6">
      <Alert v-if="error" variant="destructive">
        <CircleAlert />
        <AlertDescription>
          {{ t("dashboard.systemUsageError") }}
        </AlertDescription>
      </Alert>
      <Alert v-if="containerError" variant="destructive">
        <CircleAlert />
        <AlertDescription>
          {{ t("dashboard.containerUsageError") }}
        </AlertDescription>
      </Alert>
      <div
        v-if="loading && !stats"
        class="grid gap-8 sm:grid-cols-2 xl:grid-cols-4"
      >
        <div
          v-for="item in 4"
          :key="item"
          class="flex flex-col items-center gap-5"
        >
          <Skeleton class="size-28 rounded-full" />
          <Skeleton class="h-4 w-20" />
          <Skeleton class="h-3 w-28" />
        </div>
      </div>
      <div
        v-else-if="metrics.length"
        class="grid gap-8 sm:grid-cols-2 xl:grid-cols-6"
      >
        <UsageGauge
          v-for="metric in metrics"
          :key="metric.key"
          :label="metric.label"
          :percent="metric.percent"
          :used="metric.used"
          :total="metric.total"
          :unit="metric.unit"
          :variant="metric.variant"
        />
      </div>
    </CardContent>
  </Card>
</template>

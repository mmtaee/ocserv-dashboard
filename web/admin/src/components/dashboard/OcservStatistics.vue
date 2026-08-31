<script setup lang="ts">
import { Activity, CheckCircle2, CircleAlert, Server } from "@lucide/vue";
import { computed } from "vue";
import { useI18n } from "vue-i18n";

import type { OcservStats } from "@/api/services/dashboard";
import { Alert, AlertDescription } from "@/components/ui/alert";
import {
  Card,
  CardContent,
  CardDescription,
  CardHeader,
  CardTitle,
} from "@/components/ui/card";
import { Separator } from "@/components/ui/separator";
import { Skeleton } from "@/components/ui/skeleton";

type GeneralInfo = NonNullable<OcservStats["general_info"]>;
type CurrentStats = NonNullable<OcservStats["current_stats"]>;

const GENERAL_KEYS = [
  "Status",
  "Active sessions",
  "Total sessions",
  "Total authentication failures",
  "IPs in ban list",
  "Median latency",
  "STDEV latency",
  "Up since",
  "Server PID",
  "Sec-mod PID",
  "Sec-mod instance count",
] as const satisfies readonly (keyof GeneralInfo)[];

const CURRENT_KEYS = [
  "Sessions handled",
  "Authentication failures",
  "Closed due to error sessions",
  "Timed out sessions",
  "Timed out (idle) sessions",
  "RX",
  "TX",
  "Average auth time",
  "Max auth time",
  "Average session time",
  "Max session time",
  "Last stats reset",
] as const satisfies readonly (keyof CurrentStats)[];

const props = defineProps<{
  stats: OcservStats | null;
  loading: boolean;
  error: boolean;
}>();
const { t } = useI18n({ useScope: "global" });

function displayValue(value: string | number | undefined): string {
  if (value === undefined || value === "") return "—";
  return String(value);
}

const generalMetrics = computed(() =>
  GENERAL_KEYS.map((key) => ({
    key,
    label: key,
    value: displayValue(props.stats?.general_info?.[key]),
  })),
);

const currentMetrics = computed(() =>
  CURRENT_KEYS.map((key) => ({
    key,
    label: key,
    value: displayValue(props.stats?.current_stats?.[key]),
  })),
);

const isOnline = computed(
  () => props.stats?.general_info?.Status?.toLowerCase() === "online",
);
</script>

<template>
  <Card>
    <CardHeader>
      <CardTitle>{{ t("dashboard.ocservStatistics") }}</CardTitle>
      <CardDescription>
        {{ t("dashboard.ocservStatisticsDescription") }}
      </CardDescription>
    </CardHeader>
    <CardContent class="flex flex-col gap-6">
      <Alert v-if="error" variant="destructive">
        <CircleAlert />
        <AlertDescription>
          {{ t("dashboard.ocservStatisticsError") }}
        </AlertDescription>
      </Alert>

      <div class="grid gap-4 xl:grid-cols-2">
        <Card class="gap-4 py-5 shadow-none">
          <CardHeader class="gap-1">
            <CardTitle class="flex items-center gap-2 text-base">
              <Server class="size-4 text-muted-foreground" />
              {{ t("dashboard.generalInformation") }}
            </CardTitle>
          </CardHeader>
          <CardContent>
            <div v-if="loading && !stats" class="space-y-4">
              <div
                v-for="item in GENERAL_KEYS.length"
                :key="item"
                class="flex items-center justify-between gap-4"
              >
                <Skeleton class="h-4 w-36" />
                <Skeleton class="h-4 w-24" />
              </div>
            </div>
            <dl v-else>
              <div v-for="(metric, index) in generalMetrics" :key="metric.key">
                <Separator v-if="index > 0" />
                <div class="flex items-center justify-between gap-4 py-3">
                  <dt class="text-sm text-muted-foreground">
                    {{ metric.label }}
                  </dt>
                  <dd
                    class="flex items-center gap-2 text-end text-sm font-medium"
                  >
                    <template v-if="metric.key === 'Status'">
                      <CheckCircle2
                        v-if="isOnline"
                        class="size-4 text-emerald-600 dark:text-emerald-400"
                      />
                      <CircleAlert v-else class="size-4 text-destructive" />
                    </template>
                    <span>{{ metric.value }}</span>
                  </dd>
                </div>
              </div>
            </dl>
          </CardContent>
        </Card>

        <Card class="gap-4 py-5 shadow-none">
          <CardHeader class="gap-1">
            <CardTitle class="flex items-center gap-2 text-base">
              <Activity class="size-4 text-muted-foreground" />
              {{ t("dashboard.currentStatistics") }}
            </CardTitle>
          </CardHeader>
          <CardContent>
            <div v-if="loading && !stats" class="space-y-4">
              <div
                v-for="item in CURRENT_KEYS.length"
                :key="item"
                class="flex items-center justify-between gap-4"
              >
                <Skeleton class="h-4 w-36" />
                <Skeleton class="h-4 w-24" />
              </div>
            </div>
            <dl v-else>
              <div v-for="(metric, index) in currentMetrics" :key="metric.key">
                <Separator v-if="index > 0" />
                <div class="flex items-center justify-between gap-4 py-3">
                  <dt class="text-sm text-muted-foreground">
                    {{ metric.label }}
                  </dt>
                  <dd class="text-end text-sm font-medium">
                    {{ metric.value }}
                  </dd>
                </div>
              </div>
            </dl>
          </CardContent>
        </Card>
      </div>
    </CardContent>
  </Card>
</template>

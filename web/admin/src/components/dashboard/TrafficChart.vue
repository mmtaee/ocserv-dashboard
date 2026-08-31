<script setup lang="ts">
import { ChartNoAxesCombined } from "@lucide/vue";
import { VisAxis, VisGroupedBar, VisXYContainer } from "@unovis/vue";
import { computed } from "vue";
import { useI18n } from "vue-i18n";

import type { DailyTraffic } from "@/api/services/dashboard";
import {
  Card,
  CardContent,
  CardDescription,
  CardHeader,
  CardTitle,
} from "@/components/ui/card";
import {
  ChartContainer,
  ChartCrosshair,
  ChartLegendContent,
  ChartTooltip,
  ChartTooltipContent,
  componentToString,
  type ChartConfig,
} from "@/components/ui/chart";
import {
  Empty,
  EmptyDescription,
  EmptyHeader,
  EmptyMedia,
  EmptyTitle,
} from "@/components/ui/empty";
import { Skeleton } from "@/components/ui/skeleton";

interface TrafficPoint {
  date: Date;
  rx: number;
  tx: number;
}

const props = defineProps<{
  data: readonly DailyTraffic[];
  loading: boolean;
}>();

const { locale, t } = useI18n({ useScope: "global" });
const chartConfig = {
  rx: { label: "RX", color: "var(--chart-1)" },
  tx: { label: "TX", color: "var(--chart-2)" },
} satisfies ChartConfig;

const chartData = computed<TrafficPoint[]>(() =>
  props.data
    .flatMap((item) => {
      if (!item.date) return [];
      const timestamp = Date.parse(`${item.date}T00:00:00`);
      if (Number.isNaN(timestamp)) return [];

      return [
        {
          date: new Date(timestamp),
          rx: Number(item.rx ?? 0),
          tx: Number(item.tx ?? 0),
        },
      ];
    })
    .sort((first, second) => first.date.getTime() - second.date.getTime())
    .slice(-10),
);

function formatDate(value: number | Date): string {
  return new Intl.DateTimeFormat(locale.value, {
    month: "short",
    day: "numeric",
  }).format(new Date(value));
}

const tooltipTemplate = componentToString(chartConfig, ChartTooltipContent, {
  indicator: "dashed",
  labelFormatter: formatDate,
});
</script>

<template>
  <Card>
    <CardHeader>
      <CardTitle>{{ t("dashboard.trafficStatistics") }}</CardTitle>
      <CardDescription>
        {{ t("dashboard.trafficStatisticsDescription") }}
      </CardDescription>
    </CardHeader>
    <CardContent>
      <Skeleton v-if="loading && chartData.length === 0" class="h-80 w-full" />
      <Empty v-else-if="chartData.length === 0" class="min-h-80">
        <EmptyHeader>
          <EmptyMedia variant="icon">
            <ChartNoAxesCombined />
          </EmptyMedia>
          <EmptyTitle>{{ t("dashboard.noTrafficStatistics") }}</EmptyTitle>
          <EmptyDescription>
            {{ t("dashboard.noTrafficStatisticsDescription") }}
          </EmptyDescription>
        </EmptyHeader>
      </Empty>
      <ChartContainer v-else :config="chartConfig" class="h-80 aspect-auto">
        <VisXYContainer :data="chartData">
          <VisGroupedBar
            :x="(item: TrafficPoint) => item.date"
            :y="[
              (item: TrafficPoint) => item.rx,
              (item: TrafficPoint) => item.tx,
            ]"
            :color="[chartConfig.rx.color, chartConfig.tx.color]"
            :rounded-corners="4"
            bar-padding="0.15"
            group-padding="0"
          />
          <VisAxis
            type="x"
            :x="(item: TrafficPoint) => item.date"
            :tick-line="false"
            :domain-line="false"
            :grid-line="false"
            :num-ticks="Math.min(chartData.length, 6)"
            :tick-format="formatDate"
            :tick-values="chartData.map((item) => item.date)"
          />
          <VisAxis
            type="y"
            :num-ticks="4"
            :tick-line="false"
            :domain-line="false"
          />
          <ChartTooltip />
          <ChartCrosshair :template="tooltipTemplate" color="#0000" />
        </VisXYContainer>
        <ChartLegendContent />
      </ChartContainer>
    </CardContent>
  </Card>
</template>

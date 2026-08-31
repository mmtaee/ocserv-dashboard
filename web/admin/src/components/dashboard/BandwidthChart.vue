<script setup lang="ts">
import { ChartPie } from "@lucide/vue";
import { Donut } from "@unovis/ts";
import { VisDonut, VisSingleContainer } from "@unovis/vue";
import { computed } from "vue";
import { useI18n } from "vue-i18n";

import type { TotalBandwidth } from "@/api/services/dashboard";
import { Badge } from "@/components/ui/badge";
import {
  Card,
  CardContent,
  CardDescription,
  CardFooter,
  CardHeader,
  CardTitle,
} from "@/components/ui/card";
import {
  ChartContainer,
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

interface BandwidthPoint {
  direction: "rx" | "tx";
  value: number;
}

const props = defineProps<{
  total: TotalBandwidth | null;
  loading: boolean;
}>();

const { locale, t } = useI18n({ useScope: "global" });
const chartConfig = {
  value: { label: "GB", color: "var(--chart-1)" },
  rx: { label: "RX", color: "var(--chart-1)" },
  tx: { label: "TX", color: "var(--chart-2)" },
} satisfies ChartConfig;

const chartData = computed<BandwidthPoint[]>(() => [
  { direction: "rx", value: Number(props.total?.rx ?? 0) },
  { direction: "tx", value: Number(props.total?.tx ?? 0) },
]);
const totalBandwidth = computed(() =>
  chartData.value.reduce((total, item) => total + item.value, 0),
);
const hasData = computed(() => totalBandwidth.value > 0);

function formatBandwidth(value: number): string {
  return new Intl.NumberFormat(locale.value, {
    maximumFractionDigits: 3,
  }).format(value);
}

const tooltipTemplate = componentToString(chartConfig, ChartTooltipContent, {
  hideLabel: true,
});
</script>

<template>
  <Card class="flex flex-col">
    <CardHeader>
      <CardTitle>{{ t("dashboard.totalBandwidth") }}</CardTitle>
      <CardDescription>
        {{ t("dashboard.totalBandwidthDescription") }}
      </CardDescription>
    </CardHeader>
    <CardContent class="flex-1">
      <Skeleton v-if="loading && !total" class="mx-auto size-64 rounded-full" />
      <Empty v-else-if="!hasData" class="min-h-80">
        <EmptyHeader>
          <EmptyMedia variant="icon">
            <ChartPie />
          </EmptyMedia>
          <EmptyTitle>{{ t("dashboard.noBandwidthStatistics") }}</EmptyTitle>
          <EmptyDescription>
            {{ t("dashboard.noBandwidthStatisticsDescription") }}
          </EmptyDescription>
        </EmptyHeader>
      </Empty>
      <ChartContainer
        v-else
        :config="chartConfig"
        class="mx-auto size-80 max-w-full aspect-square"
        :style="{
          '--vis-donut-central-label-font-size': 'var(--text-3xl)',
          '--vis-donut-central-label-font-weight': 'var(--font-weight-bold)',
          '--vis-donut-central-label-text-color': 'var(--foreground)',
          '--vis-donut-central-sub-label-text-color': 'var(--muted-foreground)',
        }"
      >
        <VisSingleContainer :data="chartData" :margin="{ top: 30, bottom: 30 }">
          <VisDonut
            :value="(item: BandwidthPoint) => item.value"
            :color="(item: BandwidthPoint) => chartConfig[item.direction].color"
            :arc-width="30"
            :central-label-offset-y="10"
            :central-label="formatBandwidth(totalBandwidth)"
            :central-sub-label="t('dashboard.gigabytes')"
          />
          <ChartTooltip
            :triggers="{
              [Donut.selectors.segment]: tooltipTemplate,
            }"
          />
        </VisSingleContainer>
      </ChartContainer>
    </CardContent>
    <CardFooter v-if="total" class="flex flex-wrap justify-center gap-3">
      <Badge variant="outline">
        RX: {{ formatBandwidth(total.rx) }} {{ t("dashboard.gigabytes") }}
      </Badge>
      <Badge variant="outline">
        TX: {{ formatBandwidth(total.tx) }} {{ t("dashboard.gigabytes") }}
      </Badge>
    </CardFooter>
  </Card>
</template>

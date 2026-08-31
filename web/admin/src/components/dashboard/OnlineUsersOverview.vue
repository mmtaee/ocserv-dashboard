<script setup lang="ts">
import { Users } from "@lucide/vue";
import { VisDonut, VisSingleContainer } from "@unovis/vue";
import { computed, type DeepReadonly } from "vue";
import { useI18n } from "vue-i18n";

import type { DashboardUsers } from "@/api/services/dashboard";
import { Badge } from "@/components/ui/badge";
import {
  Card,
  CardContent,
  CardDescription,
  CardFooter,
  CardHeader,
  CardTitle,
} from "@/components/ui/card";
import { ChartContainer, type ChartConfig } from "@/components/ui/chart";
import {
  Empty,
  EmptyDescription,
  EmptyHeader,
  EmptyMedia,
  EmptyTitle,
} from "@/components/ui/empty";
import { Skeleton } from "@/components/ui/skeleton";

interface UserPoint {
  state: "online" | "offline";
  value: number;
}

const props = defineProps<{
  users: DeepReadonly<DashboardUsers> | null;
  loading: boolean;
}>();

const { t } = useI18n({ useScope: "global" });
const chartConfig = {
  online: { color: "var(--chart-1)" },
  offline: { color: "var(--chart-2)" },
} satisfies ChartConfig;

const totalUsers = computed(() => Number(props.users?.total ?? 0));
const onlineUsers = computed(() => {
  const sessions = props.users?.online_users_session ?? [];
  const uniqueUsers = new Set(
    sessions.map((session) => session.Username ?? `#${session.ID}`),
  ).size;
  return Math.min(uniqueUsers, totalUsers.value);
});
const offlineUsers = computed(() =>
  Math.max(totalUsers.value - onlineUsers.value, 0),
);
const chartData = computed<UserPoint[]>(() => [
  { state: "online", value: onlineUsers.value },
  { state: "offline", value: offlineUsers.value },
]);
</script>

<template>
  <Card class="flex flex-col">
    <CardHeader>
      <CardTitle>{{ t("dashboard.onlineUsersOverview") }}</CardTitle>
      <CardDescription>{{
        t("dashboard.onlineUsersDescription")
      }}</CardDescription>
    </CardHeader>
    <CardContent class="flex-1">
      <Skeleton v-if="loading && !users" class="mx-auto size-64 rounded-full" />
      <Empty v-else-if="totalUsers === 0" class="min-h-72">
        <EmptyHeader>
          <EmptyMedia variant="icon"><Users /></EmptyMedia>
          <EmptyTitle>{{ t("dashboard.noUserStatistics") }}</EmptyTitle>
          <EmptyDescription>{{
            t("dashboard.noUserStatisticsDescription")
          }}</EmptyDescription>
        </EmptyHeader>
      </Empty>
      <ChartContainer
        v-else
        :config="chartConfig"
        class="mx-auto size-72 max-w-full aspect-square"
        :style="{
          '--vis-donut-central-label-font-size': 'var(--text-3xl)',
          '--vis-donut-central-label-font-weight': 'var(--font-weight-bold)',
          '--vis-donut-central-label-text-color': 'var(--foreground)',
          '--vis-donut-central-sub-label-text-color': 'var(--muted-foreground)',
        }"
      >
        <VisSingleContainer :data="chartData" :margin="{ top: 30, bottom: 30 }">
          <VisDonut
            :value="(item: UserPoint) => item.value"
            :color="(item: UserPoint) => chartConfig[item.state].color"
            :arc-width="30"
            :central-label-offset-y="10"
            :central-label="totalUsers.toLocaleString()"
            :central-sub-label="t('dashboard.totalUsers')"
          />
        </VisSingleContainer>
      </ChartContainer>
    </CardContent>
    <CardFooter v-if="users" class="flex flex-wrap justify-center gap-3">
      <Badge variant="outline">
        {{ t("dashboard.onlineUsers") }}: {{ onlineUsers }}
      </Badge>
      <Badge variant="outline">
        {{ t("dashboard.offlineUsers") }}: {{ offlineUsers }}
      </Badge>
    </CardFooter>
  </Card>
</template>

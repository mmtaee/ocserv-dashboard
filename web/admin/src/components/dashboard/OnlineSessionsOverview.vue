<script setup lang="ts">
import { MonitorDot } from "@lucide/vue";
import type { DeepReadonly } from "vue";
import { useI18n } from "vue-i18n";

import type { OnlineSession } from "@/api/services/dashboard";
import {
  Card,
  CardContent,
  CardDescription,
  CardHeader,
  CardTitle,
} from "@/components/ui/card";
import {
  Empty,
  EmptyDescription,
  EmptyHeader,
  EmptyMedia,
  EmptyTitle,
} from "@/components/ui/empty";
import { Skeleton } from "@/components/ui/skeleton";
import {
  Table,
  TableBody,
  TableCell,
  TableHead,
  TableHeader,
  TableRow,
} from "@/components/ui/table";

defineProps<{
  sessions: readonly DeepReadonly<OnlineSession>[];
  loading: boolean;
}>();

const { t } = useI18n({ useScope: "global" });

function displayGroup(group: string | undefined): string {
  if (!group || group === "(none)") return t("dashboard.defaultGroup");
  return group;
}
</script>

<template>
  <Card>
    <CardHeader>
      <CardTitle>{{ t("dashboard.onlineSessionsOverview") }}</CardTitle>
      <CardDescription>{{
        t("dashboard.onlineSessionsDescription")
      }}</CardDescription>
    </CardHeader>
    <CardContent>
      <div v-if="loading && sessions.length === 0" class="flex flex-col gap-3">
        <Skeleton v-for="row in 5" :key="row" class="h-10 w-full" />
      </div>
      <Empty v-else-if="sessions.length === 0" class="min-h-56">
        <EmptyHeader>
          <EmptyMedia variant="icon"><MonitorDot /></EmptyMedia>
          <EmptyTitle>{{ t("dashboard.noOnlineSessions") }}</EmptyTitle>
          <EmptyDescription>{{
            t("dashboard.noOnlineSessionsDescription")
          }}</EmptyDescription>
        </EmptyHeader>
      </Empty>
      <Table v-else>
        <TableHeader>
          <TableRow>
            <TableHead>{{ t("dashboard.username") }}</TableHead>
            <TableHead>{{ t("dashboard.group") }}</TableHead>
            <TableHead>{{ t("dashboard.averageRx") }}</TableHead>
            <TableHead>{{ t("dashboard.averageTx") }}</TableHead>
            <TableHead>{{ t("dashboard.connectedAt") }}</TableHead>
          </TableRow>
        </TableHeader>
        <TableBody>
          <TableRow v-for="session in sessions.slice(0, 5)" :key="session.ID">
            <TableCell class="font-medium">{{
              session.Username ?? "—"
            }}</TableCell>
            <TableCell>{{ displayGroup(session.Groupname) }}</TableCell>
            <TableCell>{{ session["Average RX"] ?? "—" }}</TableCell>
            <TableCell>{{ session["Average TX"] ?? "—" }}</TableCell>
            <TableCell>
              {{
                session["_Last connected at"] ?? session["Session started at"]
              }}
            </TableCell>
          </TableRow>
        </TableBody>
      </Table>
    </CardContent>
  </Card>
</template>

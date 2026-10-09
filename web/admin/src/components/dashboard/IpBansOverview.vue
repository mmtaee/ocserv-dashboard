<script setup lang="ts">
import { ShieldBan } from "@lucide/vue";
import type { DeepReadonly } from "vue";
import { useI18n } from "vue-i18n";

import type { IpBan } from "@/api/services/dashboard";
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
  bans: readonly DeepReadonly<IpBan>[];
  loading: boolean;
}>();

const { t } = useI18n({ useScope: "global" });
</script>

<template>
  <Card>
    <CardHeader>
      <CardTitle>{{ t("dashboard.ipBansOverview") }}</CardTitle>
      <CardDescription>{{ t("dashboard.ipBansDescription") }}</CardDescription>
    </CardHeader>
    <CardContent>
      <div v-if="loading && bans.length === 0" class="flex flex-col gap-3">
        <Skeleton v-for="row in 5" :key="row" class="h-10 w-full" />
      </div>
      <Empty v-else-if="bans.length === 0" class="min-h-56">
        <EmptyHeader>
          <EmptyMedia variant="icon"><ShieldBan /></EmptyMedia>
          <EmptyTitle>{{ t("dashboard.noIpBans") }}</EmptyTitle>
          <EmptyDescription>{{
            t("dashboard.noIpBansDescription")
          }}</EmptyDescription>
        </EmptyHeader>
      </Empty>
      <Table v-else>
        <TableHeader>
          <TableRow>
            <TableHead>{{ t("dashboard.ipAddress") }}</TableHead>
            <TableHead>{{ t("dashboard.score") }}</TableHead>
            <TableHead>{{ t("dashboard.since") }}</TableHead>
          </TableRow>
        </TableHeader>
        <TableBody>
          <TableRow
            v-for="(ban, index) in bans.slice(0, 5)"
            :key="`${ban.IP}-${index}`"
          >
            <TableCell class="font-medium">{{ ban.IP ?? "—" }}</TableCell>
            <TableCell>{{ ban.Score ?? "—" }}</TableCell>
            <TableCell>{{ ban._Since ?? ban.Since ?? "—" }}</TableCell>
          </TableRow>
        </TableBody>
      </Table>
    </CardContent>
  </Card>
</template>

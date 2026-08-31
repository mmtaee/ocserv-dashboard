<script setup lang="ts">
import type { DeepReadonly } from "vue";
import { useI18n } from "vue-i18n";

import type { TopBandwidthUsers } from "@/api/services/dashboard";
import {
  Card,
  CardContent,
  CardDescription,
  CardHeader,
  CardTitle,
} from "@/components/ui/card";
import { Skeleton } from "@/components/ui/skeleton";
import {
  Table,
  TableBody,
  TableCell,
  TableEmpty,
  TableHead,
  TableHeader,
  TableRow,
} from "@/components/ui/table";

const props = defineProps<{
  users: DeepReadonly<TopBandwidthUsers> | null;
  loading: boolean;
}>();

const { locale, t } = useI18n({ useScope: "global" });

function bytesToGigabytes(value: number): string {
  return new Intl.NumberFormat(locale.value, {
    maximumFractionDigits: 3,
  }).format(value / 1024 ** 3);
}
</script>

<template>
  <Card>
    <CardHeader>
      <CardTitle>{{ t("dashboard.topBandwidthUsers") }}</CardTitle>
      <CardDescription>{{
        t("dashboard.topBandwidthUsersDescription")
      }}</CardDescription>
    </CardHeader>
    <CardContent>
      <div v-if="loading && !users" class="grid gap-6 xl:grid-cols-2">
        <Skeleton v-for="item in 2" :key="item" class="h-64 w-full" />
      </div>
      <div v-else class="grid gap-6 xl:grid-cols-2">
        <section class="flex flex-col gap-3">
          <h3 class="font-medium">{{ t("dashboard.topReceived") }}</h3>
          <Table>
            <TableHeader>
              <TableRow>
                <TableHead>{{ t("dashboard.username") }}</TableHead>
                <TableHead>RX</TableHead>
                <TableHead>TX</TableHead>
              </TableRow>
            </TableHeader>
            <TableBody>
              <TableEmpty
                v-if="(props.users?.top_rx?.length ?? 0) === 0"
                :colspan="3"
              >
                {{ t("dashboard.noTopBandwidthUsers") }}
              </TableEmpty>
              <TableRow
                v-for="user in props.users?.top_rx?.slice(0, 5) ?? []"
                :key="user.id"
              >
                <TableCell class="font-medium">{{ user.username }}</TableCell>
                <TableCell>
                  {{ bytesToGigabytes(user.running_rx) }}
                  {{ t("dashboard.gigabytes") }}
                </TableCell>
                <TableCell>
                  {{ bytesToGigabytes(user.running_tx) }}
                  {{ t("dashboard.gigabytes") }}
                </TableCell>
              </TableRow>
            </TableBody>
          </Table>
        </section>
        <section class="flex flex-col gap-3">
          <h3 class="font-medium">{{ t("dashboard.topTransmitted") }}</h3>
          <Table>
            <TableHeader>
              <TableRow>
                <TableHead>{{ t("dashboard.username") }}</TableHead>
                <TableHead>TX</TableHead>
                <TableHead>RX</TableHead>
              </TableRow>
            </TableHeader>
            <TableBody>
              <TableEmpty
                v-if="(props.users?.top_tx?.length ?? 0) === 0"
                :colspan="3"
              >
                {{ t("dashboard.noTopBandwidthUsers") }}
              </TableEmpty>
              <TableRow
                v-for="user in props.users?.top_tx?.slice(0, 5) ?? []"
                :key="user.id"
              >
                <TableCell class="font-medium">{{ user.username }}</TableCell>
                <TableCell>
                  {{ bytesToGigabytes(user.running_tx) }}
                  {{ t("dashboard.gigabytes") }}
                </TableCell>
                <TableCell>
                  {{ bytesToGigabytes(user.running_rx) }}
                  {{ t("dashboard.gigabytes") }}
                </TableCell>
              </TableRow>
            </TableBody>
          </Table>
        </section>
      </div>
    </CardContent>
  </Card>
</template>

<script setup lang="ts">
import { Activity } from "@lucide/vue";
import { computed } from "vue";
import { useI18n } from "vue-i18n";

import type { OcservRuntimeStatus } from "@/api/services/system";
import { Badge } from "@/components/ui/badge";
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

const props = defineProps<{
  loading: boolean;
  status: OcservRuntimeStatus | null;
}>();

const { locale, t } = useI18n({ useScope: "global" });
const hasData = computed(() =>
  [
    props.status?.active_state,
    props.status?.cpu_usage_nsec,
    props.status?.description,
    props.status?.id,
    props.status?.main_pid,
    props.status?.memory,
    props.status?.start_time,
    props.status?.sub_state,
    props.status?.tasks,
    props.status?.unit_file_state,
  ].some((value) => value != null && value !== ""),
);
const stateVariant = computed<"default" | "secondary">(() =>
  props.status?.active_state === "active" ? "default" : "secondary",
);

function text(value: string | number | undefined): string {
  return value == null || value === ""
    ? t("runtime.notAvailable")
    : String(value);
}

function number(value: number | undefined): string {
  return value == null
    ? t("runtime.notAvailable")
    : new Intl.NumberFormat(locale.value).format(value);
}

function memory(value: number | undefined): string {
  return value == null
    ? t("runtime.notAvailable")
    : t("runtime.megabytes", {
        value: new Intl.NumberFormat(locale.value, {
          maximumFractionDigits: 1,
        }).format(value / 1024 / 1024),
      });
}

function cpuTime(value: number | undefined): string {
  return value == null
    ? t("runtime.notAvailable")
    : t("runtime.seconds", {
        value: new Intl.NumberFormat(locale.value, {
          maximumFractionDigits: 2,
        }).format(value / 1_000_000_000),
      });
}

function dateTime(value: string | undefined): string {
  if (!value) return t("runtime.notAvailable");
  const date = new Date(value);
  return Number.isNaN(date.getTime())
    ? value
    : new Intl.DateTimeFormat(locale.value, {
        dateStyle: "medium",
        timeStyle: "medium",
      }).format(date);
}

const statusRows = computed(() => [
  [t("runtime.service"), text(props.status?.id)],
  [t("runtime.descriptionLabel"), text(props.status?.description)],
  [t("runtime.subState"), text(props.status?.sub_state)],
  [t("runtime.unitFileState"), text(props.status?.unit_file_state)],
  [t("runtime.mainPid"), number(props.status?.main_pid)],
  [t("runtime.tasks"), number(props.status?.tasks)],
  [t("runtime.memory"), memory(props.status?.memory)],
  [t("runtime.cpuTime"), cpuTime(props.status?.cpu_usage_nsec)],
  [t("runtime.startedAt"), dateTime(props.status?.start_time)],
]);
</script>

<template>
  <Card>
    <CardHeader>
      <div class="flex flex-wrap items-start justify-between gap-3">
        <div class="flex flex-col gap-1">
          <CardTitle>{{ t("runtime.statusTitle") }}</CardTitle>
          <CardDescription>{{
            t("runtime.statusDescription")
          }}</CardDescription>
        </div>
        <Badge v-if="status?.active_state" :variant="stateVariant">
          {{ text(status.active_state) }}
        </Badge>
      </div>
    </CardHeader>
    <CardContent>
      <div v-if="loading" class="grid gap-4 sm:grid-cols-2 xl:grid-cols-3">
        <Skeleton v-for="item in 9" :key="item" class="h-16" />
      </div>
      <Empty v-else-if="!hasData">
        <EmptyHeader>
          <EmptyMedia variant="icon"><Activity /></EmptyMedia>
          <EmptyTitle>{{ t("runtime.emptyTitle") }}</EmptyTitle>
          <EmptyDescription>{{
            t("runtime.emptyDescription")
          }}</EmptyDescription>
        </EmptyHeader>
      </Empty>
      <dl v-else class="grid gap-4 sm:grid-cols-2 xl:grid-cols-3">
        <div
          v-for="item in statusRows"
          :key="item[0]"
          class="rounded-lg border p-4"
        >
          <dt class="text-sm text-muted-foreground">{{ item[0] }}</dt>
          <dd class="mt-1 break-words font-medium">{{ item[1] }}</dd>
        </div>
      </dl>
    </CardContent>
  </Card>
</template>

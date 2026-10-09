<script setup lang="ts">
import { FileClock } from "@lucide/vue";
import { computed, onMounted, reactive, shallowRef } from "vue";
import { useI18n } from "vue-i18n";

import { normalizeApiError } from "@/api/http";
import {
  getReportSessionLogs,
  type ReportSort,
  type SessionLog,
} from "@/api/services/reports";
import ReportsDateRangeFilter from "@/components/reports/ReportsDateRangeFilter.vue";
import { Alert, AlertDescription, AlertTitle } from "@/components/ui/alert";
import {
  Card,
  CardContent,
  CardFooter,
  CardHeader,
  CardTitle,
} from "@/components/ui/card";
import { Field, FieldLabel } from "@/components/ui/field";
import {
  Pagination,
  PaginationContent,
  PaginationEllipsis,
  PaginationItem,
  PaginationNext,
  PaginationPrevious,
} from "@/components/ui/pagination";
import {
  Select,
  SelectContent,
  SelectGroup,
  SelectItem,
  SelectTrigger,
  SelectValue,
} from "@/components/ui/select";
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

const { locale, t } = useI18n({ useScope: "global" });
const dates = reactive({ dateStart: "", dateEnd: "" });
const logs = shallowRef<SessionLog[]>([]);
const loading = shallowRef(true);
const error = shallowRef("");
const page = shallowRef(1);
const size = shallowRef(25);
const order = shallowRef<"created_at" | "username" | "event" | "ip">(
  "created_at",
);
const sort = shallowRef<ReportSort>("DESC");
const total = shallowRef(0);
const totalPages = computed(() =>
  Math.max(1, Math.ceil(total.value / size.value)),
);

async function load(nextPage = page.value): Promise<void> {
  loading.value = true;
  error.value = "";
  try {
    const response = await getReportSessionLogs({
      page: nextPage,
      size: size.value,
      order: order.value,
      sort: sort.value,
      dateStart: dates.dateStart || undefined,
      dateEnd: dates.dateEnd || undefined,
    });
    logs.value = response.result ?? [];
    page.value = response.meta.page;
    size.value = response.meta.size;
    total.value = response.meta.total_records;
  } catch (cause) {
    error.value = normalizeApiError(cause).message;
  } finally {
    loading.value = false;
  }
}

function applyFilters(): void {
  void load(1);
}
function formatDate(value: string): string {
  const date = new Date(value);
  return Number.isNaN(date.getTime())
    ? value
    : new Intl.DateTimeFormat(locale.value, {
        dateStyle: "medium",
        timeStyle: "medium",
      }).format(date);
}
onMounted(() => load());
</script>

<template>
  <div class="flex flex-col gap-6">
    <div>
      <h1 class="text-2xl font-semibold tracking-tight">
        {{ t("reports.sessionLogsTitle") }}
      </h1>
      <p class="text-sm text-muted-foreground">
        {{ t("reports.sessionLogsDescription") }}
      </p>
    </div>
    <ReportsDateRangeFilter
      v-model:date-start="dates.dateStart"
      v-model:date-end="dates.dateEnd"
      optional
      :loading="loading"
      @apply="applyFilters"
      @clear="applyFilters"
    />
    <div class="flex flex-wrap gap-4">
      <Field class="w-48"
        ><FieldLabel>{{ t("reports.orderBy") }}</FieldLabel
        ><Select
          v-model="order"
          :disabled="loading"
          @update:model-value="load(1)"
          ><SelectTrigger><SelectValue /></SelectTrigger
          ><SelectContent
            ><SelectGroup
              ><SelectItem value="created_at">{{
                t("reports.createdAt")
              }}</SelectItem
              ><SelectItem value="username">{{
                t("reports.username")
              }}</SelectItem
              ><SelectItem value="event">{{ t("reports.event") }}</SelectItem
              ><SelectItem value="ip">{{
                t("reports.ip")
              }}</SelectItem></SelectGroup
            ></SelectContent
          ></Select
        ></Field
      >
      <Field class="w-40"
        ><FieldLabel>{{ t("reports.sort") }}</FieldLabel
        ><Select
          v-model="sort"
          :disabled="loading"
          @update:model-value="load(1)"
          ><SelectTrigger><SelectValue /></SelectTrigger
          ><SelectContent
            ><SelectGroup
              ><SelectItem value="DESC">{{
                t("reports.newestFirst")
              }}</SelectItem
              ><SelectItem value="ASC">{{
                t("reports.oldestFirst")
              }}</SelectItem></SelectGroup
            ></SelectContent
          ></Select
        ></Field
      >
      <Field class="w-32"
        ><FieldLabel>{{ t("reports.pageSize") }}</FieldLabel
        ><Select
          v-model="size"
          :disabled="loading"
          @update:model-value="load(1)"
          ><SelectTrigger><SelectValue /></SelectTrigger
          ><SelectContent
            ><SelectGroup
              ><SelectItem :value="10">10</SelectItem
              ><SelectItem :value="25">25</SelectItem
              ><SelectItem :value="50">50</SelectItem
              ><SelectItem :value="100">100</SelectItem></SelectGroup
            ></SelectContent
          ></Select
        ></Field
      >
    </div>
    <Alert v-if="error" variant="error"
      ><AlertTitle>{{ t("reports.requestFailed") }}</AlertTitle
      ><AlertDescription>{{ error }}</AlertDescription></Alert
    >
    <Card
      ><CardHeader
        ><CardTitle>{{ t("reports.sessionLogs") }}</CardTitle></CardHeader
      ><CardContent>
        <div v-if="loading && logs.length === 0" class="flex flex-col gap-3">
          <Skeleton v-for="item in 6" :key="item" class="h-12 w-full" />
        </div>
        <Table v-else
          ><TableHeader
            ><TableRow
              ><TableHead>{{ t("reports.createdAt") }}</TableHead
              ><TableHead>{{ t("reports.username") }}</TableHead
              ><TableHead>{{ t("reports.ip") }}</TableHead
              ><TableHead>{{ t("reports.event") }}</TableHead
              ><TableHead>{{ t("reports.message") }}</TableHead></TableRow
            ></TableHeader
          ><TableBody
            ><TableEmpty v-if="logs.length === 0" :colspan="5"
              ><div class="flex flex-col items-center gap-2 py-6">
                <FileClock /><span>{{ t("reports.noSessionLogs") }}</span>
              </div></TableEmpty
            ><TableRow
              v-for="log in logs"
              :key="`${log.created_at}-${log.username}-${log.event}`"
              ><TableCell class="whitespace-nowrap">{{
                formatDate(log.created_at)
              }}</TableCell
              ><TableCell class="font-medium">{{ log.username }}</TableCell
              ><TableCell>{{ log.ip || "—" }}</TableCell
              ><TableCell>{{ t(`reports.events.${log.event}`) }}</TableCell
              ><TableCell class="max-w-md break-words">{{
                log.message
              }}</TableCell></TableRow
            ></TableBody
          ></Table
        > </CardContent
      ><CardFooter
        v-if="total > 0"
        class="grid items-center gap-3 sm:grid-cols-[1fr_auto_1fr]"
        ><span class="text-sm text-muted-foreground">{{
          t("reports.pageStatus", { page, pages: totalPages, total })
        }}</span
        ><Pagination
          v-slot="{ page: currentPage }"
          class="mx-auto w-auto"
          :disabled="loading"
          :items-per-page="size"
          :page="page"
          :sibling-count="1"
          show-edges
          :total="total"
          @update:page="load"
          ><PaginationContent v-slot="{ items }"
            ><PaginationPrevious :label="t('reports.previous')" /><template
              v-for="(item, index) in items"
              :key="`${item.type}-${index}`"
              ><PaginationItem
                v-if="item.type === 'page'"
                :value="item.value"
                :is-active="item.value === currentPage"
                >{{ item.value }}</PaginationItem
              ><PaginationEllipsis
                v-else
                :index="index"
                :label="t('reports.morePages')" /></template
            ><PaginationNext
              :label="
                t('reports.next')
              " /></PaginationContent></Pagination></CardFooter
    ></Card>
  </div>
</template>

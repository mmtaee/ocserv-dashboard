<script setup lang="ts">
import type { ColumnDef } from "@tanstack/vue-table";
import { ArrowUpDown, MoreHorizontal, Network } from "@lucide/vue";
import { createColumnHelper } from "@tanstack/vue-table";
import { createReusableTemplate } from "@vueuse/core";
import { h } from "vue";
import { useI18n } from "vue-i18n";

import type {
  OcservGroup,
  OcservGroupWithTraffic,
} from "@/api/services/ocserv-groups";
import { Button } from "@/components/ui/button";
import { DataTable, type DataTableFeatures } from "@/components/ui/data-table";
import {
  DropdownMenu,
  DropdownMenuContent,
  DropdownMenuGroup,
  DropdownMenuItem,
  DropdownMenuTrigger,
} from "@/components/ui/dropdown-menu";
import {
  Empty,
  EmptyDescription,
  EmptyHeader,
  EmptyMedia,
  EmptyTitle,
} from "@/components/ui/empty";

defineProps<{
  groups: OcservGroupWithTraffic[];
  loading?: boolean;
}>();
const emit = defineEmits<{
  delete: [group: OcservGroup];
  edit: [group: OcservGroup];
  view: [group: OcservGroup];
}>();
const { locale, t } = useI18n({ useScope: "global" });
const [DefineActions, ReuseActions] = createReusableTemplate<{
  group: OcservGroupWithTraffic;
}>();

function configuredFieldCount(group: OcservGroup): number {
  return Object.values(group.config ?? {}).filter((value) => value != null)
    .length;
}

function bytesToGigabytes(value: number): string {
  return new Intl.NumberFormat(locale.value, {
    maximumFractionDigits: 3,
  }).format(value / 1024 ** 3);
}

const columnHelper = createColumnHelper<
  DataTableFeatures,
  OcservGroupWithTraffic
>();
const columns: ColumnDef<DataTableFeatures, OcservGroupWithTraffic>[] =
  columnHelper.columns([
    columnHelper.accessor((group) => group.id ?? 0, {
      id: "id",
      header: ({ column }) =>
        h(
          Button,
          {
            type: "button",
            variant: "ghost",
            onClick: () => column.toggleSorting(column.getIsSorted() === "asc"),
          },
          () => [
            t("ocservGroups.id"),
            h(ArrowUpDown, { "data-icon": "inline-end" }),
          ],
        ),
      cell: ({ row }) =>
        h(
          "span",
          { class: "font-mono text-muted-foreground" },
          row.original.id ?? "—",
        ),
    }),
    columnHelper.accessor("name", {
      filterFn: "includesString",
      sortFn: "text",
      header: ({ column }) =>
        h(
          Button,
          {
            type: "button",
            variant: "ghost",
            onClick: () => column.toggleSorting(column.getIsSorted() === "asc"),
          },
          () => [
            t("ocservGroups.name"),
            h(ArrowUpDown, { "data-icon": "inline-end" }),
          ],
        ),
      cell: ({ row }) => h("span", { class: "font-medium" }, row.original.name),
    }),
    columnHelper.accessor((group) => configuredFieldCount(group), {
      id: "config",
      header: ({ column }) =>
        h(
          Button,
          {
            type: "button",
            variant: "ghost",
            onClick: () => column.toggleSorting(column.getIsSorted() === "asc"),
          },
          () => [
            t("ocservGroups.config"),
            h(ArrowUpDown, { "data-icon": "inline-end" }),
          ],
        ),
      cell: ({ getValue }) =>
        h("span", { class: "text-muted-foreground" }, String(getValue())),
    }),
    columnHelper.accessor("total_rx", {
      header: ({ column }) =>
        h(
          Button,
          {
            type: "button",
            variant: "ghost",
            onClick: () => column.toggleSorting(column.getIsSorted() === "asc"),
          },
          () => [
            t("ocservGroups.totalRx"),
            h(ArrowUpDown, { "data-icon": "inline-end" }),
          ],
        ),
      cell: ({ getValue }) =>
        h("span", { class: "font-mono tabular-nums" }, [
          bytesToGigabytes(getValue()),
          ` ${t("dashboard.gigabytes")}`,
        ]),
    }),
    columnHelper.accessor("total_tx", {
      header: ({ column }) =>
        h(
          Button,
          {
            type: "button",
            variant: "ghost",
            onClick: () => column.toggleSorting(column.getIsSorted() === "asc"),
          },
          () => [
            t("ocservGroups.totalTx"),
            h(ArrowUpDown, { "data-icon": "inline-end" }),
          ],
        ),
      cell: ({ getValue }) =>
        h("span", { class: "font-mono tabular-nums" }, [
          bytesToGigabytes(getValue()),
          ` ${t("dashboard.gigabytes")}`,
        ]),
    }),
    columnHelper.display({
      id: "actions",
      enableSorting: false,
      header: () => h("span", { class: "sr-only" }, t("ocservGroups.actions")),
      cell: ({ row }) => h(ReuseActions, { group: row.original }),
    }),
  ]);
</script>

<template>
  <DefineActions v-slot="{ group }">
    <div class="text-end">
      <DropdownMenu>
        <DropdownMenuTrigger as-child>
          <Button
            type="button"
            variant="ghost"
            size="icon-sm"
            :aria-label="t('ocservGroups.actions')"
          >
            <MoreHorizontal />
          </Button>
        </DropdownMenuTrigger>
        <DropdownMenuContent align="end">
          <DropdownMenuGroup>
            <DropdownMenuItem @select="emit('view', group)">
              {{ t("ocservGroups.view") }}
            </DropdownMenuItem>
            <DropdownMenuItem @select="emit('edit', group)">
              {{ t("ocservGroups.edit") }}
            </DropdownMenuItem>
            <DropdownMenuItem
              variant="destructive"
              @select="emit('delete', group)"
            >
              {{ t("ocservGroups.delete") }}
            </DropdownMenuItem>
          </DropdownMenuGroup>
        </DropdownMenuContent>
      </DropdownMenu>
    </div>
  </DefineActions>

  <DataTable
    :columns="columns"
    :data="groups"
    filter-column="name"
    :filter-label="t('ocservGroups.search')"
    :filter-placeholder="t('ocservGroups.searchPlaceholder')"
    :loading="loading"
  >
    <template #empty>
      <Empty class="border-0 py-14">
        <EmptyHeader>
          <EmptyMedia variant="icon">
            <Network />
          </EmptyMedia>
          <EmptyTitle>
            {{ t("ocservGroups.noGroups") }}
          </EmptyTitle>
          <EmptyDescription>
            {{ t("ocservGroups.noGroupsDescription") }}
          </EmptyDescription>
        </EmptyHeader>
      </Empty>
    </template>
  </DataTable>
</template>

<script setup lang="ts">
import type { ColumnDef } from "@tanstack/vue-table";
import { MoreHorizontal, UserCog } from "@lucide/vue";
import { createColumnHelper } from "@tanstack/vue-table";
import { createReusableTemplate } from "@vueuse/core";
import { h } from "vue";
import { useI18n } from "vue-i18n";

import type { Staff } from "@/api/services/staffs";
import { Badge } from "@/components/ui/badge";
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

defineProps<{ loading?: boolean; staffs: Staff[] }>();
const emit = defineEmits<{
  delete: [staff: Staff];
  password: [staff: Staff];
}>();
const { locale, t } = useI18n({ useScope: "global" });
const [DefineActions, ReuseActions] = createReusableTemplate<{
  staff: Staff;
}>();

function formatDate(value?: string): string {
  if (!value) return "—";
  const date = new Date(value);
  return Number.isNaN(date.getTime())
    ? "—"
    : new Intl.DateTimeFormat(locale.value, {
        dateStyle: "medium",
        timeStyle: "short",
      }).format(date);
}

const columnHelper = createColumnHelper<DataTableFeatures, Staff>();
const columns: ColumnDef<DataTableFeatures, Staff>[] = columnHelper.columns([
  columnHelper.accessor("username", {
    header: t("staffs.username"),
    cell: ({ getValue }) => h("span", { class: "font-medium" }, getValue()),
  }),
  columnHelper.accessor("superadmin", {
    header: t("staffs.accessLevel"),
    cell: ({ getValue }) =>
      h(Badge, { variant: getValue() ? "default" : "secondary" }, () =>
        t(getValue() ? "staffs.superadmin" : "staffs.staff"),
      ),
  }),
  columnHelper.accessor("last_login", {
    header: t("staffs.lastLogin"),
    cell: ({ getValue }) => formatDate(getValue()),
  }),
  columnHelper.accessor("created_at", {
    header: t("staffs.createdAt"),
    cell: ({ getValue }) => formatDate(getValue()),
  }),
  columnHelper.accessor("updated_at", {
    header: t("staffs.updatedAt"),
    cell: ({ getValue }) => formatDate(getValue()),
  }),
  columnHelper.display({
    id: "actions",
    header: () => h("span", { class: "sr-only" }, t("staffs.actions")),
    cell: ({ row }) => h(ReuseActions, { staff: row.original }),
  }),
]);
</script>

<template>
  <DefineActions v-slot="{ staff }">
    <DropdownMenu>
      <DropdownMenuTrigger as-child>
        <Button
          type="button"
          variant="ghost"
          size="icon-sm"
          :aria-label="t('staffs.actions')"
        >
          <MoreHorizontal />
        </Button>
      </DropdownMenuTrigger>
      <DropdownMenuContent align="end">
        <DropdownMenuGroup>
          <DropdownMenuItem @select="emit('password', staff)">
            {{ t("staffs.changePassword") }}
          </DropdownMenuItem>
          <DropdownMenuItem
            variant="destructive"
            @select="emit('delete', staff)"
          >
            {{ t("staffs.delete") }}
          </DropdownMenuItem>
        </DropdownMenuGroup>
      </DropdownMenuContent>
    </DropdownMenu>
  </DefineActions>

  <DataTable
    align="center"
    :columns="columns"
    :data="staffs"
    :loading="loading"
  >
    <template #empty>
      <Empty class="border-0 py-14">
        <EmptyHeader>
          <EmptyMedia variant="icon"><UserCog /></EmptyMedia>
          <EmptyTitle>{{ t("staffs.noStaffs") }}</EmptyTitle>
          <EmptyDescription>{{
            t("staffs.noStaffsDescription")
          }}</EmptyDescription>
        </EmptyHeader>
      </Empty>
    </template>
  </DataTable>
</template>

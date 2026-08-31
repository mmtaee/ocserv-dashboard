<script setup lang="ts">
import type { ColumnDef } from "@tanstack/vue-table";
import { ArrowUpDown, MoreHorizontal, Users } from "@lucide/vue";
import { createColumnHelper } from "@tanstack/vue-table";
import { createReusableTemplate } from "@vueuse/core";
import { computed, h } from "vue";
import { useI18n } from "vue-i18n";

import type { OcservUser } from "@/api/services/ocserv-users";
import { Badge } from "@/components/ui/badge";
import { Button } from "@/components/ui/button";
import { Checkbox } from "@/components/ui/checkbox";
import { DataTable, type DataTableFeatures } from "@/components/ui/data-table";
import {
  DropdownMenu,
  DropdownMenuContent,
  DropdownMenuGroup,
  DropdownMenuItem,
  DropdownMenuSeparator,
  DropdownMenuTrigger,
} from "@/components/ui/dropdown-menu";
import {
  Empty,
  EmptyDescription,
  EmptyHeader,
  EmptyMedia,
  EmptyTitle,
} from "@/components/ui/empty";

export type OcservUserAction =
  | "activate"
  | "certificate"
  | "delete"
  | "disconnect"
  | "download"
  | "edit"
  | "lock"
  | "resetUsage"
  | "terminate"
  | "unlock"
  | "view";

const props = defineProps<{
  loading?: boolean;
  selectedIds: number[];
  users: OcservUser[];
}>();
const emit = defineEmits<{
  action: [action: OcservUserAction, user: OcservUser];
  "update:selectedIds": [ids: number[]];
}>();
const { locale, t } = useI18n({ useScope: "global" });
const [DefineActions, ReuseActions] = createReusableTemplate<{
  user: OcservUser;
}>();
const pageIds = computed(() => props.users.map(({ id }) => id));
const allSelected = computed(
  () => pageIds.value.length > 0 && pageIds.value.every(isSelected),
);
const someSelected = computed(
  () => !allSelected.value && pageIds.value.some(isSelected),
);

function isSelected(id: number): boolean {
  return props.selectedIds.includes(id);
}

function setSelected(id: number, selected: boolean): void {
  emit(
    "update:selectedIds",
    selected
      ? Array.from(new Set([...props.selectedIds, id]))
      : props.selectedIds.filter((item) => item !== id),
  );
}

function setAllSelected(selected: boolean): void {
  emit("update:selectedIds", selected ? pageIds.value : []);
}

function formatBytes(value: number): string {
  return `${new Intl.NumberFormat(locale.value, { maximumFractionDigits: 2 }).format(value / 1024 ** 3)} ${t("dashboard.gigabytes")}`;
}

function formatDate(value?: string): string {
  if (!value) return "—";
  return new Intl.DateTimeFormat(locale.value, { dateStyle: "medium" }).format(
    new Date(value),
  );
}

function status(user: OcservUser): string {
  if (user.deactivated_at) return t("ocservUsers.deactivated");
  if (user.is_locked) return t("ocservUsers.locked");
  if (user.is_online) return t("ocservUsers.online");
  return t("ocservUsers.active");
}

function sortableHeader(
  column: {
    toggleSorting: (descending?: boolean) => void;
    getIsSorted: () => false | "asc" | "desc";
  },
  label: string,
) {
  return h(
    Button,
    {
      type: "button",
      variant: "ghost",
      onClick: () => column.toggleSorting(column.getIsSorted() === "asc"),
    },
    () => [label, h(ArrowUpDown, { "data-icon": "inline-end" })],
  );
}

const columnHelper = createColumnHelper<DataTableFeatures, OcservUser>();
const columns: ColumnDef<DataTableFeatures, OcservUser>[] =
  columnHelper.columns([
    columnHelper.display({
      id: "select",
      enableSorting: false,
      header: () =>
        h(Checkbox, {
          "aria-label": t("ocservUsers.selectPage"),
          modelValue: allSelected.value
            ? true
            : someSelected.value
              ? "indeterminate"
              : false,
          "onUpdate:modelValue": (value: boolean | "indeterminate") =>
            setAllSelected(value === true),
        }),
      cell: ({ row }) =>
        h(Checkbox, {
          "aria-label": t("ocservUsers.selectUser", {
            username: row.original.username,
          }),
          modelValue: isSelected(row.original.id),
          "onUpdate:modelValue": (value: boolean | "indeterminate") =>
            setSelected(row.original.id, value === true),
        }),
    }),
    columnHelper.accessor("username", {
      sortFn: "text",
      header: ({ column }) => sortableHeader(column, t("ocservUsers.username")),
      cell: ({ row }) =>
        h("div", { class: "grid gap-0.5" }, [
          h("span", { class: "font-medium" }, row.original.username),
          h(
            "span",
            { class: "text-xs text-muted-foreground" },
            row.original.description || "—",
          ),
        ]),
    }),
    columnHelper.accessor("group", {
      sortFn: "text",
      header: ({ column }) => sortableHeader(column, t("ocservUsers.group")),
    }),
    columnHelper.accessor((user) => status(user), {
      id: "status",
      header: t("ocservUsers.status"),
      cell: ({ getValue, row }) =>
        h(
          Badge,
          {
            variant: row.original.deactivated_at
              ? "destructive"
              : row.original.is_online
                ? "default"
                : "secondary",
          },
          () => getValue(),
        ),
    }),
    columnHelper.accessor("expire_at", {
      header: ({ column }) => sortableHeader(column, t("ocservUsers.expires")),
      cell: ({ row }) =>
        row.original.expiry_mode === "first_connection"
          ? t("ocservUsers.daysAfterFirst", {
              days: row.original.expire_days_after_first_connection ?? 0,
            })
          : row.original.expiry_mode === "unlimited"
            ? t("ocservUsers.never")
            : formatDate(row.original.expire_at),
    }),
    columnHelper.accessor((user) => user.running_rx + user.running_tx, {
      id: "usage",
      header: ({ column }) => sortableHeader(column, t("ocservUsers.usage")),
      cell: ({ getValue }) =>
        h("span", { class: "font-mono tabular-nums" }, formatBytes(getValue())),
    }),
    columnHelper.accessor("online_sessions", {
      header: t("ocservUsers.sessions"),
      cell: ({ getValue }) => String(getValue().length),
    }),
    columnHelper.display({
      id: "actions",
      enableSorting: false,
      header: () => h("span", { class: "sr-only" }, t("ocservUsers.actions")),
      cell: ({ row }) => h(ReuseActions, { user: row.original }),
    }),
  ]);
</script>

<template>
  <DefineActions v-slot="{ user }">
    <DropdownMenu>
      <DropdownMenuTrigger as-child>
        <Button
          type="button"
          variant="ghost"
          size="icon-sm"
          :aria-label="t('ocservUsers.actions')"
        >
          <MoreHorizontal />
        </Button>
      </DropdownMenuTrigger>
      <DropdownMenuContent align="end">
        <DropdownMenuGroup>
          <DropdownMenuItem @select="emit('action', 'view', user)">{{
            t("ocservUsers.view")
          }}</DropdownMenuItem>
          <DropdownMenuItem @select="emit('action', 'edit', user)">{{
            t("ocservUsers.edit")
          }}</DropdownMenuItem>
          <DropdownMenuItem
            v-if="user.deactivated_at"
            @select="emit('action', 'activate', user)"
            >{{ t("ocservUsers.activate") }}</DropdownMenuItem
          >
          <DropdownMenuItem
            v-if="user.is_locked"
            @select="emit('action', 'unlock', user)"
            >{{ t("ocservUsers.unlock") }}</DropdownMenuItem
          >
          <DropdownMenuItem v-else @select="emit('action', 'lock', user)">{{
            t("ocservUsers.lock")
          }}</DropdownMenuItem>
          <DropdownMenuItem @select="emit('action', 'resetUsage', user)">{{
            t("ocservUsers.resetUsage")
          }}</DropdownMenuItem>
        </DropdownMenuGroup>
        <DropdownMenuSeparator />
        <DropdownMenuGroup>
          <DropdownMenuItem @select="emit('action', 'certificate', user)">{{
            t("ocservUsers.createCertificate")
          }}</DropdownMenuItem>
          <DropdownMenuItem
            v-if="user.certificate_available"
            @select="emit('action', 'download', user)"
            >{{ t("ocservUsers.downloadCertificate") }}</DropdownMenuItem
          >
          <DropdownMenuItem
            v-if="user.is_online"
            @select="emit('action', 'disconnect', user)"
            >{{ t("ocservUsers.disconnect") }}</DropdownMenuItem
          >
          <DropdownMenuItem
            v-if="user.is_online"
            @select="emit('action', 'terminate', user)"
            >{{ t("ocservUsers.terminate") }}</DropdownMenuItem
          >
        </DropdownMenuGroup>
        <DropdownMenuSeparator />
        <DropdownMenuItem
          variant="destructive"
          @select="emit('action', 'delete', user)"
        >
          {{ t("ocservUsers.delete") }}
        </DropdownMenuItem>
      </DropdownMenuContent>
    </DropdownMenu>
  </DefineActions>

  <DataTable align="center" :columns="columns" :data="users" :loading="loading">
    <template #empty>
      <Empty class="border-0 py-14">
        <EmptyHeader>
          <EmptyMedia variant="icon"><Users /></EmptyMedia>
          <EmptyTitle>{{ t("ocservUsers.noUsers") }}</EmptyTitle>
          <EmptyDescription>{{
            t("ocservUsers.noUsersDescription")
          }}</EmptyDescription>
        </EmptyHeader>
      </Empty>
    </template>
  </DataTable>
</template>

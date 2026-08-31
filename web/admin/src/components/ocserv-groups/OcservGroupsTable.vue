<script setup lang="ts">
import { MoreHorizontal, Network } from "@lucide/vue";
import { useI18n } from "vue-i18n";

import type {
  OcservGroup,
  OcservGroupWithTraffic,
} from "@/api/services/ocserv-groups";
import { Button } from "@/components/ui/button";
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
  groups: OcservGroupWithTraffic[];
  loading?: boolean;
}>();
const emit = defineEmits<{
  delete: [group: OcservGroup];
  edit: [group: OcservGroup];
  view: [group: OcservGroup];
}>();
const { locale, t } = useI18n({ useScope: "global" });

function configuredFieldCount(group: OcservGroup): number {
  return Object.values(group.config ?? {}).filter((value) => value != null)
    .length;
}

function bytesToGigabytes(value: number): string {
  return new Intl.NumberFormat(locale.value, {
    maximumFractionDigits: 3,
  }).format(value / 1024 ** 3);
}
</script>

<template>
  <div v-if="loading" class="flex flex-col gap-3 p-4" aria-busy="true">
    <Skeleton v-for="index in 6" :key="index" class="h-12 w-full" />
  </div>

  <Empty v-else-if="groups.length === 0" class="border-0 py-14">
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

  <Table v-else>
    <TableHeader>
      <TableRow>
        <TableHead>{{ t("ocservGroups.id") }}</TableHead>
        <TableHead>{{ t("ocservGroups.name") }}</TableHead>
        <TableHead>{{ t("ocservGroups.config") }}</TableHead>
        <TableHead>{{ t("ocservGroups.totalRx") }}</TableHead>
        <TableHead>{{ t("ocservGroups.totalTx") }}</TableHead>
        <TableHead class="w-16 text-end">
          <span class="sr-only">{{ t("ocservGroups.actions") }}</span>
        </TableHead>
      </TableRow>
    </TableHeader>
    <TableBody>
      <TableRow v-for="group in groups" :key="group.id ?? group.name">
        <TableCell class="font-mono text-muted-foreground">
          {{ group.id ?? "—" }}
        </TableCell>
        <TableCell class="font-medium">
          {{ group.name }}
        </TableCell>
        <TableCell class="text-muted-foreground">
          {{ configuredFieldCount(group) }}
        </TableCell>
        <TableCell class="font-mono tabular-nums">
          {{ bytesToGigabytes(group.total_rx) }}
          {{ t("dashboard.gigabytes") }}
        </TableCell>
        <TableCell class="font-mono tabular-nums">
          {{ bytesToGigabytes(group.total_tx) }}
          {{ t("dashboard.gigabytes") }}
        </TableCell>
        <TableCell class="text-end">
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
        </TableCell>
      </TableRow>
    </TableBody>
  </Table>
</template>

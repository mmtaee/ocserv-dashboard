<script setup lang="ts">
import { computed } from "vue";
import { useI18n } from "vue-i18n";

import type { OcservGroup } from "@/api/services/ocserv-groups";
import { Badge } from "@/components/ui/badge";
import {
  Empty,
  EmptyDescription,
  EmptyHeader,
  EmptyTitle,
} from "@/components/ui/empty";
import {
  Sheet,
  SheetContent,
  SheetDescription,
  SheetHeader,
  SheetTitle,
} from "@/components/ui/sheet";
import { Skeleton } from "@/components/ui/skeleton";
import {
  Table,
  TableBody,
  TableCell,
  TableHead,
  TableHeader,
  TableRow,
} from "@/components/ui/table";

const props = defineProps<{
  group?: OcservGroup | null;
  loading?: boolean;
}>();
const open = defineModel<boolean>("open", { default: false });
const { t } = useI18n({ useScope: "global" });

const configEntries = computed(() =>
  Object.entries(props.group?.config ?? {})
    .filter(([, value]) => value != null)
    .sort(([left], [right]) => left.localeCompare(right)),
);

function displayValue(value: unknown): string {
  if (Array.isArray(value)) return value.join(", ");
  return String(value);
}
</script>

<template>
  <Sheet v-model:open="open">
    <SheetContent class="overflow-y-auto sm:max-w-2xl">
      <SheetHeader>
        <SheetTitle>
          {{ t("ocservGroups.detailsTitle") }}
        </SheetTitle>
        <SheetDescription>
          {{ t("ocservGroups.detailsDescription") }}
        </SheetDescription>
      </SheetHeader>

      <div v-if="loading" class="flex flex-col gap-3 px-4" aria-busy="true">
        <Skeleton v-for="index in 6" :key="index" class="h-12 w-full" />
      </div>

      <div v-else-if="group" class="flex flex-col gap-6 px-4 pb-6">
        <dl class="grid grid-cols-[auto_1fr] gap-x-4 gap-y-2 text-sm">
          <dt class="text-muted-foreground">{{ t("ocservGroups.id") }}</dt>
          <dd class="font-mono">{{ group.id ?? "—" }}</dd>
          <dt class="text-muted-foreground">{{ t("ocservGroups.name") }}</dt>
          <dd class="font-medium">{{ group.name }}</dd>
        </dl>

        <Table v-if="configEntries.length">
          <TableHeader>
            <TableRow>
              <TableHead>{{ t("ocservGroups.configField") }}</TableHead>
              <TableHead>{{ t("ocservGroups.value") }}</TableHead>
            </TableRow>
          </TableHeader>
          <TableBody>
            <TableRow v-for="[key, value] in configEntries" :key="key">
              <TableCell class="font-mono text-xs">{{ key }}</TableCell>
              <TableCell>
                <Badge v-if="typeof value === 'boolean'" variant="secondary">
                  {{
                    value
                      ? t("ocservGroups.enabled")
                      : t("ocservGroups.disabled")
                  }}
                </Badge>
                <span v-else class="break-all">{{ displayValue(value) }}</span>
              </TableCell>
            </TableRow>
          </TableBody>
        </Table>

        <Empty v-else class="border py-10">
          <EmptyHeader>
            <EmptyTitle>{{ t("ocservGroups.noConfig") }}</EmptyTitle>
            <EmptyDescription>
              {{ t("ocservGroups.noConfigDescription") }}
            </EmptyDescription>
          </EmptyHeader>
        </Empty>
      </div>
    </SheetContent>
  </Sheet>
</template>

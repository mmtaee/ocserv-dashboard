<script setup lang="ts" generic="TData extends RowData">
import type { ColumnDef, RowData } from "@tanstack/vue-table";
import { FlexRender, useTable } from "@tanstack/vue-table";
import { useId } from "vue";

import {
  dataTableFeatures,
  type DataTableFeatures,
} from "@/components/ui/data-table/features";
import { Field, FieldLabel } from "@/components/ui/field";
import { Input } from "@/components/ui/input";
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
  columns: ColumnDef<DataTableFeatures, TData>[];
  data: TData[];
  filterColumn?: string;
  filterLabel?: string;
  filterPlaceholder?: string;
  loading?: boolean;
}>();

const filterId = useId();
const table = useTable({
  features: dataTableFeatures,
  get columns() {
    return props.columns;
  },
  get data() {
    return props.data;
  },
});

function updateFilter(value: string | number): void {
  if (!props.filterColumn) return;
  table.getColumn(props.filterColumn)?.setFilterValue(String(value));
}
</script>

<template>
  <div class="flex flex-col gap-4">
    <Field v-if="filterColumn">
      <FieldLabel :for="filterId" class="sr-only">
        {{ filterLabel }}
      </FieldLabel>
      <Input
        :id="filterId"
        class="max-w-sm"
        :model-value="
          (table.getColumn(filterColumn)?.getFilterValue() as string) ?? ''
        "
        :placeholder="filterPlaceholder"
        type="search"
        @update:model-value="updateFilter"
      />
    </Field>

    <div class="overflow-hidden rounded-md border">
      <Table>
        <TableHeader>
          <TableRow
            v-for="headerGroup in table.getHeaderGroups()"
            :key="headerGroup.id"
          >
            <TableHead v-for="header in headerGroup.headers" :key="header.id">
              <FlexRender v-if="!header.isPlaceholder" :header="header" />
            </TableHead>
          </TableRow>
        </TableHeader>

        <TableBody>
          <template v-if="loading">
            <TableRow v-for="rowIndex in 6" :key="rowIndex">
              <TableCell v-for="(_, columnIndex) in columns" :key="columnIndex">
                <Skeleton class="h-5 w-full" />
              </TableCell>
            </TableRow>
          </template>

          <template v-else-if="table.getRowModel().rows.length">
            <TableRow v-for="row in table.getRowModel().rows" :key="row.id">
              <TableCell v-for="cell in row.getAllCells()" :key="cell.id">
                <FlexRender :cell="cell" />
              </TableCell>
            </TableRow>
          </template>

          <TableRow v-else>
            <TableCell :colspan="Math.max(1, columns.length)" class="p-0">
              <slot name="empty" />
            </TableCell>
          </TableRow>
        </TableBody>
      </Table>
    </div>
  </div>
</template>

<script setup lang="ts">
import { computed, onMounted, shallowRef } from "vue";
import { RefreshCw, Settings2 } from "@lucide/vue";
import { useI18n } from "vue-i18n";

import type { ModelsOcservGroupConfig } from "@/api/generated";
import { normalizeApiError } from "@/api/http";
import {
  getUnsyncedGroups,
  syncGroups,
  type UnsyncedGroup,
} from "@/api/services/ocserv-sync";
import OcservGroupEditorSheet from "@/components/ocserv-groups/OcservGroupEditorSheet.vue";
import { cloneOcservGroupConfig } from "@/components/ocserv-groups/group-config";
import { Alert, AlertDescription, AlertTitle } from "@/components/ui/alert";
import { Badge } from "@/components/ui/badge";
import { Button } from "@/components/ui/button";
import { Checkbox } from "@/components/ui/checkbox";
import {
  Empty,
  EmptyDescription,
  EmptyHeader,
  EmptyTitle,
} from "@/components/ui/empty";
import { Skeleton } from "@/components/ui/skeleton";
import { Spinner } from "@/components/ui/spinner";
import {
  Table,
  TableBody,
  TableCell,
  TableHead,
  TableHeader,
  TableRow,
} from "@/components/ui/table";

const { t } = useI18n({ useScope: "global" });
const groups = shallowRef<UnsyncedGroup[]>([]);
const selected = shallowRef<string[]>([]);
const editing = shallowRef<UnsyncedGroup | null>(null);
const editorOpen = shallowRef(false);
const loading = shallowRef(false);
const submitting = shallowRef(false);
const error = shallowRef("");
const result = shallowRef<string[]>([]);
const allSelected = computed(
  () =>
    groups.value.length > 0 && selected.value.length === groups.value.length,
);

async function load(): Promise<void> {
  loading.value = true;
  error.value = "";
  result.value = [];
  try {
    groups.value = await getUnsyncedGroups();
    selected.value = groups.value.map(({ name }) => name);
  } catch (cause) {
    error.value = normalizeApiError(cause).message;
  } finally {
    loading.value = false;
  }
}

function toggle(name: string, checked: boolean): void {
  selected.value = checked
    ? [...selected.value, name]
    : selected.value.filter((item) => item !== name);
}

function toggleAll(checked: boolean): void {
  selected.value = checked ? groups.value.map(({ name }) => name) : [];
}

function edit(group: UnsyncedGroup): void {
  editing.value = group;
  editorOpen.value = true;
}

function update(request: { config: ModelsOcservGroupConfig }): void {
  if (!editing.value) return;
  groups.value = groups.value.map((group) =>
    group.name === editing.value?.name
      ? { ...group, config: cloneOcservGroupConfig(request.config) }
      : group,
  );
  editorOpen.value = false;
}

async function submit(): Promise<void> {
  const items = groups.value.filter(({ name }) =>
    selected.value.includes(name),
  );
  if (!items.length) {
    error.value = t("ocservSync.selectionRequired");
    return;
  }
  submitting.value = true;
  error.value = "";
  try {
    const synced = await syncGroups(items);
    await load();
    result.value = synced;
  } catch (cause) {
    error.value = normalizeApiError(cause).message;
  } finally {
    submitting.value = false;
  }
}

onMounted(load);
</script>

<template>
  <div class="flex flex-col gap-4">
    <Alert v-if="error" variant="error">
      <AlertTitle>{{ t("ocservSync.error") }}</AlertTitle>
      <AlertDescription>{{ error }}</AlertDescription>
    </Alert>
    <Alert v-if="result.length" variant="success">
      <AlertTitle>{{ t("ocservSync.success") }}</AlertTitle>
      <AlertDescription>{{ result.join(", ") }}</AlertDescription>
    </Alert>
    <Skeleton v-if="loading" class="h-48 w-full" />
    <Empty v-else-if="!groups.length">
      <EmptyHeader>
        <EmptyTitle>{{ t("ocservSync.noGroups") }}</EmptyTitle>
        <EmptyDescription>
          {{ t("ocservSync.noGroupsDescription") }}
        </EmptyDescription>
      </EmptyHeader>
    </Empty>
    <Table v-else>
      <TableHeader>
        <TableRow>
          <TableHead>
            <Checkbox
              :aria-label="t('ocservSync.selectAll')"
              :model-value="allSelected"
              @update:model-value="toggleAll(Boolean($event))"
            />
          </TableHead>
          <TableHead>{{ t("ocservSync.group") }}</TableHead>
          <TableHead>{{ t("ocservSync.sourcePath") }}</TableHead>
          <TableHead>{{ t("ocservSync.configuration") }}</TableHead>
          <TableHead class="text-end">{{ t("ocservSync.actions") }}</TableHead>
        </TableRow>
      </TableHeader>
      <TableBody>
        <TableRow v-for="group in groups" :key="group.name">
          <TableCell>
            <Checkbox
              :aria-label="t('ocservSync.selectItem', { name: group.name })"
              :model-value="selected.includes(group.name)"
              @update:model-value="toggle(group.name, Boolean($event))"
            />
          </TableCell>
          <TableCell class="font-medium">{{ group.name }}</TableCell>
          <TableCell class="max-w-72 truncate text-muted-foreground">
            {{ group.path || t("ocservSync.notAvailable") }}
          </TableCell>
          <TableCell>
            <Badge variant="secondary">
              {{
                t("ocservSync.configFields", {
                  count: Object.keys(group.config).length,
                })
              }}
            </Badge>
          </TableCell>
          <TableCell class="text-end">
            <Button size="sm" variant="outline" @click="edit(group)">
              <Settings2 data-icon="inline-start" />
              {{ t("ocservSync.configureGroup") }}
            </Button>
          </TableCell>
        </TableRow>
      </TableBody>
    </Table>
    <div class="flex flex-wrap justify-end gap-2">
      <Button variant="outline" :disabled="loading || submitting" @click="load">
        <RefreshCw data-icon="inline-start" />
        {{ t("ocservSync.refresh") }}
      </Button>
      <Button
        :disabled="loading || submitting || !groups.length"
        @click="submit"
      >
        <Spinner v-if="submitting" data-icon="inline-start" />
        {{ t("ocservSync.syncSelected") }}
      </Button>
    </div>
    <OcservGroupEditorSheet
      v-model:open="editorOpen"
      :existing-names="groups.map(({ name }) => name)"
      :group="editing"
      mode="edit"
      :pending="submitting"
      @submit="update"
    />
  </div>
</template>

<script setup lang="ts">
import { computed, onMounted, shallowRef } from "vue";
import { Plus, RefreshCw } from "@lucide/vue";
import { useI18n } from "vue-i18n";

import type {
  OcservGroup,
  OcservGroupCreate,
  OcservGroupUpdate,
} from "@/api/services/ocserv-groups";
import OcservGroupDeleteDialog from "@/components/ocserv-groups/OcservGroupDeleteDialog.vue";
import OcservGroupDetailsSheet from "@/components/ocserv-groups/OcservGroupDetailsSheet.vue";
import OcservGroupEditorSheet from "@/components/ocserv-groups/OcservGroupEditorSheet.vue";
import OcservGroupsTable from "@/components/ocserv-groups/OcservGroupsTable.vue";
import { Alert, AlertDescription, AlertTitle } from "@/components/ui/alert";
import { Button } from "@/components/ui/button";
import {
  Card,
  CardAction,
  CardContent,
  CardDescription,
  CardFooter,
  CardHeader,
  CardTitle,
} from "@/components/ui/card";
import { Field, FieldLabel } from "@/components/ui/field";
import { Input } from "@/components/ui/input";
import { Separator } from "@/components/ui/separator";
import { Spinner } from "@/components/ui/spinner";
import { useOcservGroups } from "@/composables/useOcservGroups";

const { t } = useI18n({ useScope: "global" });
const {
  create,
  detailLoading,
  error,
  groupNames,
  groups,
  loadGroup,
  loading,
  meta,
  mutating,
  refresh,
  remove,
  success,
  update,
} = useOcservGroups();

const query = shallowRef("");
const editorOpen = shallowRef(false);
const editorMode = shallowRef<"create" | "edit">("create");
const editorGroup = shallowRef<OcservGroup | null>(null);
const detailsOpen = shallowRef(false);
const detailsGroup = shallowRef<OcservGroup | null>(null);
const deleteOpen = shallowRef(false);
const deleteGroup = shallowRef<OcservGroup | null>(null);

const filteredGroups = computed(() => {
  const value = query.value.trim().toLocaleLowerCase();
  if (!value) return groups.value;
  return groups.value.filter(({ name }) =>
    name.toLocaleLowerCase().includes(value),
  );
});
const totalPages = computed(() =>
  Math.max(1, Math.ceil(meta.value.total_records / meta.value.size)),
);
const successMessage = computed(() => {
  if (!success.value) return "";
  return t(`ocservGroups.${success.value}Success`);
});

function openCreate(): void {
  editorMode.value = "create";
  editorGroup.value = null;
  editorOpen.value = true;
}

async function openEdit(group: OcservGroup): Promise<void> {
  if (group.id == null) return;
  editorMode.value = "edit";
  editorGroup.value = null;
  editorOpen.value = true;
  editorGroup.value = await loadGroup(group.id);
  if (!editorGroup.value) editorOpen.value = false;
}

async function openDetails(group: OcservGroup): Promise<void> {
  if (group.id == null) return;
  detailsGroup.value = null;
  detailsOpen.value = true;
  detailsGroup.value = await loadGroup(group.id);
  if (!detailsGroup.value) detailsOpen.value = false;
}

function openDelete(group: OcservGroup): void {
  deleteGroup.value = group;
  deleteOpen.value = true;
}

async function submitEditor(
  request: OcservGroupCreate | OcservGroupUpdate,
): Promise<void> {
  const saved =
    editorMode.value === "create"
      ? await create(request as OcservGroupCreate)
      : editorGroup.value?.id != null
        ? await update(editorGroup.value.id, request as OcservGroupUpdate)
        : false;
  if (saved) editorOpen.value = false;
}

async function confirmDelete(): Promise<void> {
  if (deleteGroup.value?.id == null) return;
  const deleted = await remove(deleteGroup.value.id);
  if (deleted) {
    deleteOpen.value = false;
    deleteGroup.value = null;
  }
}

onMounted(() => refresh());
</script>

<template>
  <div class="flex flex-col gap-6">
    <Alert v-if="error" variant="destructive">
      <AlertTitle>{{ t("ocservGroups.requestFailure") }}</AlertTitle>
      <AlertDescription>{{ error }}</AlertDescription>
    </Alert>

    <Alert v-if="success">
      <AlertTitle>{{ t("ocservGroups.success") }}</AlertTitle>
      <AlertDescription>{{ successMessage }}</AlertDescription>
    </Alert>

    <Card>
      <CardHeader>
        <CardTitle>{{ t("navigation.ocservGroups") }}</CardTitle>
        <CardDescription>{{ t("ocservGroups.description") }}</CardDescription>
        <CardAction class="flex gap-2">
          <Button
            type="button"
            variant="outline"
            :disabled="loading"
            @click="refresh()"
          >
            <Spinner v-if="loading" data-icon="inline-start" />
            <RefreshCw v-else data-icon="inline-start" />
            {{ t("ocservGroups.refresh") }}
          </Button>
          <Button type="button" @click="openCreate">
            <Plus data-icon="inline-start" />
            {{ t("ocservGroups.create") }}
          </Button>
        </CardAction>
      </CardHeader>

      <CardContent class="flex flex-col gap-4 px-0">
        <Field class="px-6">
          <FieldLabel for="ocserv-group-search" class="sr-only">
            {{ t("ocservGroups.search") }}
          </FieldLabel>
          <Input
            id="ocserv-group-search"
            v-model="query"
            :placeholder="t('ocservGroups.searchPlaceholder')"
            type="search"
          />
        </Field>

        <OcservGroupsTable
          :groups="filteredGroups"
          :loading="loading"
          @delete="openDelete"
          @edit="openEdit"
          @view="openDetails"
        />
      </CardContent>

      <Separator />
      <CardFooter class="flex items-center justify-between">
        <span class="text-sm text-muted-foreground">
          {{
            t("ocservGroups.pageStatus", {
              page: meta.page,
              pages: totalPages,
              total: meta.total_records,
            })
          }}
        </span>
        <div class="flex gap-2">
          <Button
            type="button"
            size="sm"
            variant="outline"
            :disabled="loading || meta.page <= 1"
            @click="refresh(meta.page - 1)"
          >
            {{ t("ocservGroups.previous") }}
          </Button>
          <Button
            type="button"
            size="sm"
            variant="outline"
            :disabled="loading || meta.page >= totalPages"
            @click="refresh(meta.page + 1)"
          >
            {{ t("ocservGroups.next") }}
          </Button>
        </div>
      </CardFooter>
    </Card>

    <OcservGroupEditorSheet
      v-model:open="editorOpen"
      :existing-names="groupNames"
      :group="editorGroup"
      :loading="editorMode === 'edit' && detailLoading"
      :mode="editorMode"
      :pending="mutating"
      @submit="submitEditor"
    />
    <OcservGroupDetailsSheet
      v-model:open="detailsOpen"
      :group="detailsGroup"
      :loading="detailLoading"
    />
    <OcservGroupDeleteDialog
      v-model:open="deleteOpen"
      :group="deleteGroup"
      :pending="mutating"
      @confirm="confirmDelete"
    />
  </div>
</template>

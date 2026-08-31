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
import {
  Pagination,
  PaginationContent,
  PaginationEllipsis,
  PaginationItem,
  PaginationNext,
  PaginationPrevious,
} from "@/components/ui/pagination";
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

const editorOpen = shallowRef(false);
const editorMode = shallowRef<"create" | "edit">("create");
const editorGroup = shallowRef<OcservGroup | null>(null);
const detailsOpen = shallowRef(false);
const detailsGroup = shallowRef<OcservGroup | null>(null);
const deleteOpen = shallowRef(false);
const deleteGroup = shallowRef<OcservGroup | null>(null);

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

function changePage(page: number): void {
  if (loading.value || page === meta.value.page) return;
  void refresh(page);
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

      <CardContent>
        <OcservGroupsTable
          :groups="groups"
          :loading="loading"
          @delete="openDelete"
          @edit="openEdit"
          @view="openDetails"
        />
      </CardContent>

      <Separator />
      <CardFooter
        class="flex flex-col items-center gap-3 sm:flex-row sm:justify-between"
      >
        <span class="text-sm text-muted-foreground">
          {{
            t("ocservGroups.pageStatus", {
              page: meta.page,
              pages: totalPages,
              total: meta.total_records,
            })
          }}
        </span>
        <Pagination
          v-slot="{ page }"
          class="mx-0 w-auto"
          :disabled="loading"
          :items-per-page="meta.size"
          :page="meta.page"
          :sibling-count="1"
          show-edges
          :total="meta.total_records"
          @update:page="changePage"
        >
          <PaginationContent v-slot="{ items }">
            <PaginationPrevious :label="t('ocservGroups.previous')" />
            <template
              v-for="(item, index) in items"
              :key="`${item.type}-${index}`"
            >
              <PaginationItem
                v-if="item.type === 'page'"
                :value="item.value"
                :is-active="item.value === page"
              >
                {{ item.value }}
              </PaginationItem>
              <PaginationEllipsis
                v-else
                :index="index"
                :label="t('ocservGroups.morePages')"
              />
            </template>
            <PaginationNext :label="t('ocservGroups.next')" />
          </PaginationContent>
        </Pagination>
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

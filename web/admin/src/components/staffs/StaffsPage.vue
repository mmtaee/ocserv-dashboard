<script setup lang="ts">
import { computed, onMounted, shallowRef } from "vue";
import { Plus, RefreshCw } from "@lucide/vue";
import { useI18n } from "vue-i18n";

import type { Staff, StaffCreate, StaffPassword } from "@/api/services/staffs";
import StaffCredentialSheet from "@/components/staffs/StaffCredentialSheet.vue";
import StaffDeleteDialog from "@/components/staffs/StaffDeleteDialog.vue";
import StaffsTable from "@/components/staffs/StaffsTable.vue";
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
import { useStaffs } from "@/composables/useStaffs";

const { t } = useI18n({ useScope: "global" });
const {
  changePassword,
  create,
  error,
  loading,
  meta,
  mutating,
  refresh,
  remove,
  staffs,
  success,
} = useStaffs();
const credentialOpen = shallowRef(false);
const credentialMode = shallowRef<"create" | "password">("create");
const selectedStaff = shallowRef<Staff | null>(null);
const deleteOpen = shallowRef(false);
const deleteStaff = shallowRef<Staff | null>(null);
const totalPages = computed(() =>
  Math.max(1, Math.ceil(meta.value.total_records / meta.value.size)),
);

function openCreate(): void {
  credentialMode.value = "create";
  selectedStaff.value = null;
  credentialOpen.value = true;
}

function openPassword(staff: Staff): void {
  credentialMode.value = "password";
  selectedStaff.value = staff;
  credentialOpen.value = true;
}

function openDelete(staff: Staff): void {
  deleteStaff.value = staff;
  deleteOpen.value = true;
}

async function submitCredential(
  request: StaffCreate | StaffPassword,
): Promise<void> {
  const saved =
    credentialMode.value === "create"
      ? await create(request as StaffCreate)
      : selectedStaff.value
        ? await changePassword(selectedStaff.value.id, request as StaffPassword)
        : false;
  if (saved) credentialOpen.value = false;
}

async function confirmDelete(): Promise<void> {
  if (!deleteStaff.value) return;
  if (await remove(deleteStaff.value.id)) {
    deleteOpen.value = false;
    deleteStaff.value = null;
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
    <Alert v-if="error" variant="error">
      <AlertTitle>{{ t("staffs.requestFailure") }}</AlertTitle>
      <AlertDescription>{{ error }}</AlertDescription>
    </Alert>
    <Alert v-if="success" variant="success">
      <AlertTitle>{{ t("staffs.success") }}</AlertTitle>
      <AlertDescription>{{ t(`staffs.${success}Success`) }}</AlertDescription>
    </Alert>

    <Card>
      <CardHeader>
        <CardTitle>{{ t("navigation.staffs") }}</CardTitle>
        <CardDescription>{{ t("staffs.description") }}</CardDescription>
        <CardAction class="flex gap-2">
          <Button
            type="button"
            variant="outline"
            :disabled="loading"
            @click="refresh()"
          >
            <Spinner v-if="loading" data-icon="inline-start" />
            <RefreshCw v-else data-icon="inline-start" />
            {{ t("staffs.refresh") }}
          </Button>
          <Button type="button" @click="openCreate">
            <Plus data-icon="inline-start" />
            {{ t("staffs.create") }}
          </Button>
        </CardAction>
      </CardHeader>
      <CardContent>
        <StaffsTable
          :loading="loading"
          :staffs="staffs"
          @delete="openDelete"
          @password="openPassword"
        />
      </CardContent>
      <Separator />
      <CardFooter class="grid items-center gap-3 sm:grid-cols-[1fr_auto_1fr]">
        <span class="text-center text-sm text-muted-foreground sm:text-start">
          {{
            t("staffs.pageStatus", {
              page: meta.page,
              pages: totalPages,
              total: meta.total_records,
            })
          }}
        </span>
        <Pagination
          v-slot="{ page }"
          class="mx-auto w-auto sm:col-start-2"
          :disabled="loading"
          :items-per-page="meta.size"
          :page="meta.page"
          :sibling-count="1"
          show-edges
          :total="meta.total_records"
          @update:page="changePage"
        >
          <PaginationContent v-slot="{ items }">
            <PaginationPrevious :label="t('staffs.previous')" />
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
                :label="t('staffs.morePages')"
              />
            </template>
            <PaginationNext :label="t('staffs.next')" />
          </PaginationContent>
        </Pagination>
      </CardFooter>
    </Card>

    <StaffCredentialSheet
      v-model:open="credentialOpen"
      :mode="credentialMode"
      :pending="mutating"
      :staff="selectedStaff"
      @submit="submitCredential"
    />
    <StaffDeleteDialog
      v-model:open="deleteOpen"
      :pending="mutating"
      :staff="deleteStaff"
      @confirm="confirmDelete"
    />
  </div>
</template>

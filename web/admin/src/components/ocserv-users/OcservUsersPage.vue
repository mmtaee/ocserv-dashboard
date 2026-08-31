<script setup lang="ts">
import { computed, onMounted, reactive, shallowRef } from "vue";
import { Plus, RefreshCw, Search, X } from "@lucide/vue";
import { useI18n } from "vue-i18n";

import type {
  OcservUser,
  OcservUserActivation,
  OcservUserCreate,
  OcservUsersListOptions,
  OcservUserUpdate,
} from "@/api/services/ocserv-users";
import { getOcservGroupNames } from "@/api/services/ocserv-groups";
import { normalizeApiError } from "@/api/http";
import OcservUserActionDialog from "@/components/ocserv-users/OcservUserActionDialog.vue";
import OcservUserActivationSheet from "@/components/ocserv-users/OcservUserActivationSheet.vue";
import OcservUserDetailsSheet from "@/components/ocserv-users/OcservUserDetailsSheet.vue";
import OcservUserEditorSheet from "@/components/ocserv-users/OcservUserEditorSheet.vue";
import OcservUsersTable, {
  type OcservUserAction,
} from "@/components/ocserv-users/OcservUsersTable.vue";
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
import {
  Pagination,
  PaginationContent,
  PaginationEllipsis,
  PaginationItem,
  PaginationNext,
  PaginationPrevious,
} from "@/components/ui/pagination";
import {
  Select,
  SelectContent,
  SelectItem,
  SelectTrigger,
  SelectValue,
} from "@/components/ui/select";
import { Separator } from "@/components/ui/separator";
import { Spinner } from "@/components/ui/spinner";
import { useOcservUsers } from "@/composables/useOcservUsers";

const { t } = useI18n({ useScope: "global" });
const {
  activate,
  bulkDescription,
  bulkGroup,
  bulkRemove,
  bulkStatus,
  create,
  createCertificate,
  detailLoading,
  disconnect,
  disconnectSession,
  downloadCertificate,
  error,
  loadUser,
  loading,
  lock,
  meta,
  mutating,
  refresh,
  remove,
  resetUsage,
  success,
  terminate,
  terminateSession,
  unlock,
  update,
  users,
} = useOcservUsers();

const groupNames = shallowRef<string[]>([]);
const pageError = shallowRef("");
const selectedIds = shallowRef<number[]>([]);
const editorOpen = shallowRef(false);
const editorMode = shallowRef<"create" | "edit">("create");
const editorUser = shallowRef<OcservUser | null>(null);
const detailsOpen = shallowRef(false);
const detailsUser = shallowRef<OcservUser | null>(null);
const activationOpen = shallowRef(false);
const activationUser = shallowRef<OcservUser | null>(null);
const confirmationOpen = shallowRef(false);
const confirmationTitle = shallowRef("");
const confirmationDescription = shallowRef("");
const confirmationAction = shallowRef("");
const confirmationDestructive = shallowRef(false);
const pendingConfirmation = shallowRef<(() => Promise<boolean>) | null>(null);
const bulkGroupName = shallowRef("__none__");
const bulkDescriptionText = shallowRef("");
const filters = reactive({
  expireInDays: "",
  filter: "all",
  group: "all",
  q: "",
});

const totalPages = computed(() =>
  Math.max(1, Math.ceil(meta.value.total_records / meta.value.size)),
);
const successMessage = computed(() =>
  success.value ? t(`ocservUsers.${success.value}Success`) : "",
);
const requestError = computed(() => error.value || pageError.value);

function currentFilters(): OcservUsersListOptions {
  const expireInDays = Number(filters.expireInDays);
  return {
    expireInDays: expireInDays > 0 ? expireInDays : undefined,
    filter:
      filters.filter === "all"
        ? undefined
        : (filters.filter as OcservUsersListOptions["filter"]),
    group: filters.group === "all" ? undefined : filters.group,
    q: filters.q.trim() || undefined,
  };
}

async function reload(page = meta.value.page): Promise<void> {
  selectedIds.value = [];
  pageError.value = "";
  await refresh(page, currentFilters());
}

function resetFilters(): void {
  filters.expireInDays = "";
  filters.filter = "all";
  filters.group = "all";
  filters.q = "";
  void reload(1);
}

function openCreate(): void {
  editorMode.value = "create";
  editorUser.value = null;
  editorOpen.value = true;
}

async function openEdit(user: OcservUser): Promise<void> {
  editorMode.value = "edit";
  editorUser.value = null;
  editorOpen.value = true;
  editorUser.value = await loadUser(user.id);
  if (!editorUser.value) editorOpen.value = false;
}

async function openDetails(user: OcservUser): Promise<void> {
  detailsUser.value = null;
  detailsOpen.value = true;
  detailsUser.value = await loadUser(user.id);
  if (!detailsUser.value) detailsOpen.value = false;
}

function openActivation(user: OcservUser): void {
  activationUser.value = user;
  activationOpen.value = true;
}

function confirmAction(
  key: string,
  userLabel: string,
  action: () => Promise<boolean>,
  destructive = false,
): void {
  confirmationTitle.value = t(`ocservUsers.confirmations.${key}Title`);
  confirmationDescription.value = t(
    `ocservUsers.confirmations.${key}Description`,
    { target: userLabel },
  );
  confirmationAction.value = t(`ocservUsers.confirmations.${key}Action`);
  confirmationDestructive.value = destructive;
  pendingConfirmation.value = action;
  confirmationOpen.value = true;
}

async function runConfirmation(): Promise<void> {
  if (!pendingConfirmation.value) return;
  if (await pendingConfirmation.value()) {
    confirmationOpen.value = false;
    pendingConfirmation.value = null;
    selectedIds.value = [];
    if (detailsOpen.value && detailsUser.value)
      detailsUser.value = await loadUser(detailsUser.value.id);
  }
}

function handleAction(action: OcservUserAction, user: OcservUser): void {
  if (action === "view") return void openDetails(user);
  if (action === "edit") return void openEdit(user);
  if (action === "activate") return openActivation(user);
  if (action === "download") return void downloadCertificate(user);
  const operations: Record<string, () => Promise<boolean>> = {
    certificate: () => createCertificate(user.id),
    delete: () => remove(user.id),
    disconnect: () => disconnect(user.username),
    lock: () => lock(user.id),
    resetUsage: () => resetUsage(user.id),
    terminate: () => terminate(user.username),
    unlock: () => unlock(user.id),
  };
  const operation = operations[action];
  if (operation)
    confirmAction(
      action,
      user.username,
      operation,
      action === "delete" || action === "terminate",
    );
}

async function submitEditor(
  request: OcservUserCreate | OcservUserUpdate,
): Promise<void> {
  const saved =
    editorMode.value === "create"
      ? await create(request as OcservUserCreate)
      : editorUser.value
        ? await update(editorUser.value.id, request as OcservUserUpdate)
        : false;
  if (saved) editorOpen.value = false;
}

async function submitActivation(request: OcservUserActivation): Promise<void> {
  if (!activationUser.value) return;
  if (await activate(activationUser.value.id, request))
    activationOpen.value = false;
}

async function applyBulkGroup(): Promise<void> {
  if (bulkGroupName.value === "__none__") return;
  if (
    await bulkGroup(
      selectedIds.value,
      bulkGroupName.value === "__remove__" ? "" : bulkGroupName.value,
    )
  ) {
    selectedIds.value = [];
    bulkGroupName.value = "__none__";
  }
}

async function applyBulkDescription(): Promise<void> {
  if (
    await bulkDescription(selectedIds.value, bulkDescriptionText.value.trim())
  ) {
    selectedIds.value = [];
    bulkDescriptionText.value = "";
  }
}

async function applyBulkStatus(enabled: boolean): Promise<void> {
  if (await bulkStatus(selectedIds.value, enabled)) selectedIds.value = [];
}

function confirmBulkDelete(): void {
  confirmAction(
    "bulkDelete",
    t("ocservUsers.selectedCount", { count: selectedIds.value.length }),
    () => bulkRemove(selectedIds.value),
    true,
  );
}

function handleSessionAction(
  action: "disconnect" | "terminate",
  id: number,
): void {
  confirmAction(
    action,
    t("ocservUsers.sessionTarget", { id }),
    () =>
      action === "disconnect" ? disconnectSession(id) : terminateSession(id),
    action === "terminate",
  );
}

onMounted(async () => {
  await refresh(1);
  try {
    groupNames.value = await getOcservGroupNames();
  } catch (cause) {
    pageError.value = normalizeApiError(cause).message;
  }
});
</script>

<template>
  <div class="flex flex-col gap-6">
    <Alert v-if="requestError" variant="destructive">
      <AlertTitle>{{ t("ocservUsers.requestFailure") }}</AlertTitle>
      <AlertDescription>{{ requestError }}</AlertDescription>
    </Alert>
    <Alert v-if="success">
      <AlertTitle>{{ t("ocservUsers.success") }}</AlertTitle>
      <AlertDescription>{{ successMessage }}</AlertDescription>
    </Alert>

    <Card>
      <CardHeader>
        <CardTitle>{{ t("navigation.ocservUsers") }}</CardTitle>
        <CardDescription>{{ t("ocservUsers.description") }}</CardDescription>
        <CardAction class="flex gap-2">
          <Button
            type="button"
            variant="outline"
            :disabled="loading"
            @click="reload()"
          >
            <Spinner v-if="loading" data-icon="inline-start" />
            <RefreshCw v-else data-icon="inline-start" />
            {{ t("ocservUsers.refresh") }}
          </Button>
          <Button type="button" @click="openCreate">
            <Plus data-icon="inline-start" />
            {{ t("ocservUsers.create") }}
          </Button>
        </CardAction>
      </CardHeader>

      <CardContent class="grid gap-5">
        <form
          class="grid items-end gap-3 md:grid-cols-2 xl:grid-cols-[2fr_1fr_1fr_1fr_auto]"
          @submit.prevent="reload(1)"
        >
          <Field>
            <FieldLabel for="ocserv-user-search">{{
              t("ocservUsers.search")
            }}</FieldLabel>
            <Input
              id="ocserv-user-search"
              v-model="filters.q"
              type="search"
              :placeholder="t('ocservUsers.searchPlaceholder')"
            />
          </Field>
          <Field>
            <FieldLabel for="ocserv-user-status-filter">{{
              t("ocservUsers.status")
            }}</FieldLabel>
            <Select v-model="filters.filter"
              ><SelectTrigger id="ocserv-user-status-filter"
                ><SelectValue /></SelectTrigger
              ><SelectContent
                ><SelectItem value="all">{{
                  t("ocservUsers.allStatuses")
                }}</SelectItem
                ><SelectItem value="online">{{
                  t("ocservUsers.online")
                }}</SelectItem
                ><SelectItem value="active">{{
                  t("ocservUsers.active")
                }}</SelectItem
                ><SelectItem value="deactivated">{{
                  t("ocservUsers.deactivated")
                }}</SelectItem
                ><SelectItem value="locked">{{
                  t("ocservUsers.locked")
                }}</SelectItem></SelectContent
              ></Select
            >
          </Field>
          <Field>
            <FieldLabel for="ocserv-user-group-filter">{{
              t("ocservUsers.group")
            }}</FieldLabel>
            <Select v-model="filters.group"
              ><SelectTrigger id="ocserv-user-group-filter"
                ><SelectValue /></SelectTrigger
              ><SelectContent
                ><SelectItem value="all">{{
                  t("ocservUsers.allGroups")
                }}</SelectItem
                ><SelectItem
                  v-for="name in groupNames"
                  :key="name"
                  :value="name"
                  >{{ name }}</SelectItem
                ></SelectContent
              ></Select
            >
          </Field>
          <Field>
            <FieldLabel for="ocserv-user-expiry-filter">{{
              t("ocservUsers.expiringWithin")
            }}</FieldLabel>
            <Input
              id="ocserv-user-expiry-filter"
              v-model="filters.expireInDays"
              min="1"
              type="number"
              :placeholder="t('ocservUsers.days')"
            />
          </Field>
          <div class="flex gap-2">
            <Button type="submit" :disabled="loading"
              ><Search data-icon="inline-start" />{{
                t("ocservUsers.apply")
              }}</Button
            >
            <Button
              type="button"
              variant="outline"
              :aria-label="t('ocservUsers.clearFilters')"
              @click="resetFilters"
              ><X
            /></Button>
          </div>
        </form>

        <div
          v-if="selectedIds.length"
          class="grid gap-3 rounded-lg border bg-muted/30 p-3 lg:grid-cols-[auto_1fr_auto] lg:items-center"
        >
          <p class="text-sm font-medium">
            {{ t("ocservUsers.selectedCount", { count: selectedIds.length }) }}
          </p>
          <div class="flex flex-wrap gap-2">
            <Button
              size="sm"
              variant="outline"
              :disabled="mutating"
              @click="applyBulkStatus(true)"
              >{{ t("ocservUsers.enable") }}</Button
            >
            <Button
              size="sm"
              variant="outline"
              :disabled="mutating"
              @click="applyBulkStatus(false)"
              >{{ t("ocservUsers.disable") }}</Button
            >
            <Select v-model="bulkGroupName" :disabled="mutating"
              ><SelectTrigger class="w-44"
                ><SelectValue
                  :placeholder="t('ocservUsers.assignGroup')" /></SelectTrigger
              ><SelectContent
                ><SelectItem value="__none__">{{
                  t("ocservUsers.assignGroup")
                }}</SelectItem
                ><SelectItem value="__remove__">{{
                  t("ocservUsers.removeGroup")
                }}</SelectItem
                ><SelectItem
                  v-for="name in groupNames"
                  :key="name"
                  :value="name"
                  >{{ name }}</SelectItem
                ></SelectContent
              ></Select
            >
            <Button
              size="sm"
              variant="outline"
              :disabled="mutating || bulkGroupName === '__none__'"
              @click="applyBulkGroup"
              >{{ t("ocservUsers.applyGroup") }}</Button
            >
            <Input
              v-model="bulkDescriptionText"
              class="w-52"
              :disabled="mutating"
              :placeholder="t('ocservUsers.bulkDescriptionPlaceholder')"
            />
            <Button
              size="sm"
              variant="outline"
              :disabled="mutating"
              @click="applyBulkDescription"
              >{{ t("ocservUsers.applyDescription") }}</Button
            >
          </div>
          <Button
            size="sm"
            variant="destructive"
            :disabled="mutating"
            @click="confirmBulkDelete"
            >{{ t("ocservUsers.deleteSelected") }}</Button
          >
        </div>

        <OcservUsersTable
          v-model:selected-ids="selectedIds"
          :users="users"
          :loading="loading"
          @action="handleAction"
        />
      </CardContent>

      <Separator />
      <CardFooter class="grid items-center gap-3 sm:grid-cols-[1fr_auto_1fr]">
        <span class="text-center text-sm text-muted-foreground sm:text-start">{{
          t("ocservUsers.pageStatus", {
            page: meta.page,
            pages: totalPages,
            total: meta.total_records,
          })
        }}</span>
        <Pagination
          v-slot="{ page }"
          class="mx-auto w-auto sm:col-start-2"
          :disabled="loading"
          :items-per-page="meta.size"
          :page="meta.page"
          :sibling-count="1"
          show-edges
          :total="meta.total_records"
          @update:page="reload"
        >
          <PaginationContent v-slot="{ items }"
            ><PaginationPrevious :label="t('ocservUsers.previous')" /><template
              v-for="(item, index) in items"
              :key="`${item.type}-${index}`"
              ><PaginationItem
                v-if="item.type === 'page'"
                :value="item.value"
                :is-active="item.value === page"
                >{{ item.value }}</PaginationItem
              ><PaginationEllipsis
                v-else
                :index="index"
                :label="t('ocservUsers.morePages')" /></template
            ><PaginationNext :label="t('ocservUsers.next')"
          /></PaginationContent>
        </Pagination>
      </CardFooter>
    </Card>

    <OcservUserEditorSheet
      v-model:open="editorOpen"
      :group-names="groupNames"
      :loading="editorMode === 'edit' && detailLoading"
      :mode="editorMode"
      :pending="mutating"
      :user="editorUser"
      @submit="submitEditor"
    />
    <OcservUserDetailsSheet
      v-model:open="detailsOpen"
      :loading="detailLoading"
      :pending="mutating"
      :user="detailsUser"
      @session-action="handleSessionAction"
    />
    <OcservUserActivationSheet
      v-model:open="activationOpen"
      :pending="mutating"
      :user="activationUser"
      @submit="submitActivation"
    />
    <OcservUserActionDialog
      v-model:open="confirmationOpen"
      :action="confirmationAction"
      :description="confirmationDescription"
      :destructive="confirmationDestructive"
      :pending="mutating"
      :title="confirmationTitle"
      @confirm="runConfirmation"
    />
  </div>
</template>

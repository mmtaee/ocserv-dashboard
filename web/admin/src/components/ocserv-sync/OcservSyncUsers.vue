<script setup lang="ts">
import { computed, onMounted, reactive, shallowRef } from "vue";
import { RefreshCw } from "@lucide/vue";
import { useI18n } from "vue-i18n";

import type {
  ModelsExpiryMode,
  ModelsOcservUserConfig,
  ModelsTrafficType,
} from "@/api/generated";
import { normalizeApiError } from "@/api/http";
import {
  getUnsyncedUsers,
  syncUsers,
  type UnsyncedUser,
} from "@/api/services/ocserv-sync";
import OcservUserConfigFields from "@/components/ocserv-users/OcservUserConfigFields.vue";
import { Alert, AlertDescription, AlertTitle } from "@/components/ui/alert";
import { Button } from "@/components/ui/button";
import { Checkbox } from "@/components/ui/checkbox";
import {
  Empty,
  EmptyDescription,
  EmptyHeader,
  EmptyTitle,
} from "@/components/ui/empty";
import {
  Field,
  FieldDescription,
  FieldGroup,
  FieldLabel,
} from "@/components/ui/field";
import { Input } from "@/components/ui/input";
import {
  Select,
  SelectContent,
  SelectGroup,
  SelectItem,
  SelectTrigger,
  SelectValue,
} from "@/components/ui/select";
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
const users = shallowRef<UnsyncedUser[]>([]);
const selected = shallowRef<string[]>([]);
const loading = shallowRef(false);
const submitting = shallowRef(false);
const error = shallowRef("");
const result = shallowRef<string[]>([]);
const total = shallowRef(0);
const page = shallowRef(1);
const size = 25;
const config = reactive<ModelsOcservUserConfig>({});
const form = reactive({
  assignConfig: false,
  description: "",
  expireAt: "",
  expireDays: 30,
  expiryMode: "unlimited" as ModelsExpiryMode,
  trafficSizeGiB: 0,
  trafficType: "Free" as ModelsTrafficType,
});
const trafficTypes: ModelsTrafficType[] = [
  "Free",
  "MonthlyTransmit",
  "MonthlyReceive",
  "MonthlyRxTx",
  "TotallyTransmit",
  "TotallyReceive",
  "TotallyRxTx",
];
const totalPages = computed(() => Math.max(1, Math.ceil(total.value / size)));

async function load(nextPage = page.value): Promise<void> {
  loading.value = true;
  error.value = "";
  result.value = [];
  try {
    const response = await getUnsyncedUsers(nextPage, size);
    page.value = response.meta.page;
    total.value = response.meta.total_records;
    users.value = response.result ?? [];
    selected.value = users.value.map(({ username }) => username);
  } catch (cause) {
    error.value = normalizeApiError(cause).message;
  } finally {
    loading.value = false;
  }
}

function toggle(username: string, checked: boolean): void {
  selected.value = checked
    ? [...selected.value, username]
    : selected.value.filter((item) => item !== username);
}

function updateConfig(
  key: keyof ModelsOcservUserConfig,
  value: ModelsOcservUserConfig[keyof ModelsOcservUserConfig],
): void {
  config[key] = value as never;
}

function validate(): string {
  if (!selected.value.length) return t("ocservSync.selectionRequired");
  if (form.trafficSizeGiB < 0) return t("ocservSync.trafficSizeInvalid");
  if (form.expiryMode === "fixed" && !form.expireAt)
    return t("ocservSync.expireAtRequired");
  if (form.expiryMode === "first_connection" && form.expireDays < 1)
    return t("ocservSync.expireDaysRequired");
  return "";
}

async function submit(): Promise<void> {
  error.value = validate();
  if (error.value) return;
  const items = users.value.filter(({ username }) =>
    selected.value.includes(username),
  );
  submitting.value = true;
  try {
    result.value = await syncUsers({
      users: items,
      description: form.description.trim() || undefined,
      expire_at: form.expiryMode === "fixed" ? form.expireAt : undefined,
      expire_days_after_first_connection:
        form.expiryMode === "first_connection" ? form.expireDays : undefined,
      expiry_mode: form.expiryMode,
      traffic_size: Math.round(form.trafficSizeGiB * 1024 ** 3),
      traffic_type: form.trafficType,
      config: form.assignConfig ? config : undefined,
    });
    const synced = [...result.value];
    await load(page.value);
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
  <div class="flex flex-col gap-5">
    <Alert v-if="error" variant="error">
      <AlertTitle>{{ t("ocservSync.error") }}</AlertTitle>
      <AlertDescription>{{ error }}</AlertDescription>
    </Alert>
    <Alert v-if="result.length" variant="success">
      <AlertTitle>{{ t("ocservSync.success") }}</AlertTitle>
      <AlertDescription>{{ result.join(", ") }}</AlertDescription>
    </Alert>
    <Skeleton v-if="loading" class="h-48 w-full" />
    <Empty v-else-if="!users.length">
      <EmptyHeader>
        <EmptyTitle>{{ t("ocservSync.noUsers") }}</EmptyTitle>
        <EmptyDescription>{{
          t("ocservSync.noUsersDescription")
        }}</EmptyDescription>
      </EmptyHeader>
    </Empty>
    <Table v-else>
      <TableHeader>
        <TableRow>
          <TableHead>{{ t("ocservSync.select") }}</TableHead>
          <TableHead>{{ t("ocservSync.username") }}</TableHead>
          <TableHead>{{ t("ocservSync.group") }}</TableHead>
        </TableRow>
      </TableHeader>
      <TableBody>
        <TableRow v-for="user in users" :key="user.username">
          <TableCell>
            <Checkbox
              :aria-label="t('ocservSync.selectItem', { name: user.username })"
              :model-value="selected.includes(user.username)"
              @update:model-value="toggle(user.username, Boolean($event))"
            />
          </TableCell>
          <TableCell>{{ user.username }}</TableCell>
          <TableCell>{{
            user.group || t("ocservSync.defaultGroup")
          }}</TableCell>
        </TableRow>
      </TableBody>
    </Table>
    <div v-if="totalPages > 1" class="flex items-center justify-end gap-2">
      <Button
        variant="outline"
        :disabled="page <= 1 || loading"
        @click="load(page - 1)"
      >
        {{ t("ocservSync.previous") }}
      </Button>
      <span class="text-sm text-muted-foreground"
        >{{ page }} / {{ totalPages }}</span
      >
      <Button
        variant="outline"
        :disabled="page >= totalPages || loading"
        @click="load(page + 1)"
      >
        {{ t("ocservSync.next") }}
      </Button>
    </div>
    <FieldGroup v-if="users.length">
      <Field>
        <FieldLabel>{{ t("ocservSync.trafficType") }}</FieldLabel>
        <Select v-model="form.trafficType" :disabled="submitting">
          <SelectTrigger><SelectValue /></SelectTrigger>
          <SelectContent>
            <SelectGroup>
              <SelectItem
                v-for="type in trafficTypes"
                :key="type"
                :value="type"
              >
                {{ t(`ocservUsers.trafficTypes.${type}`) }}
              </SelectItem>
            </SelectGroup>
          </SelectContent>
        </Select>
      </Field>
      <Field
        v-if="form.trafficType !== 'Free'"
        :data-invalid="form.trafficSizeGiB < 0"
      >
        <FieldLabel for="sync-traffic-size">{{
          t("ocservSync.trafficSize")
        }}</FieldLabel>
        <Input
          id="sync-traffic-size"
          v-model="form.trafficSizeGiB"
          min="0"
          type="number"
        />
      </Field>
      <Field>
        <FieldLabel>{{ t("ocservSync.expiryMode") }}</FieldLabel>
        <Select v-model="form.expiryMode" :disabled="submitting">
          <SelectTrigger><SelectValue /></SelectTrigger>
          <SelectContent>
            <SelectGroup>
              <SelectItem value="unlimited">{{
                t("ocservUsers.expiryUnlimited")
              }}</SelectItem>
              <SelectItem value="fixed">{{
                t("ocservUsers.expiryFixed")
              }}</SelectItem>
              <SelectItem value="first_connection">{{
                t("ocservUsers.expiryFirstConnection")
              }}</SelectItem>
            </SelectGroup>
          </SelectContent>
        </Select>
      </Field>
      <Field v-if="form.expiryMode === 'fixed'" :data-invalid="!form.expireAt">
        <FieldLabel for="sync-expire-at">{{
          t("ocservSync.expireAt")
        }}</FieldLabel>
        <Input id="sync-expire-at" v-model="form.expireAt" type="date" />
      </Field>
      <Field
        v-if="form.expiryMode === 'first_connection'"
        :data-invalid="form.expireDays < 1"
      >
        <FieldLabel for="sync-expire-days">{{
          t("ocservSync.expireDays")
        }}</FieldLabel>
        <Input
          id="sync-expire-days"
          v-model="form.expireDays"
          min="1"
          type="number"
        />
      </Field>
      <Field>
        <FieldLabel for="sync-description">{{
          t("ocservSync.descriptionLabel")
        }}</FieldLabel>
        <Input
          id="sync-description"
          v-model="form.description"
          maxlength="1024"
        />
      </Field>
      <Field>
        <div class="flex items-center gap-3">
          <Checkbox id="sync-config" v-model="form.assignConfig" />
          <FieldLabel for="sync-config">{{
            t("ocservSync.assignConfig")
          }}</FieldLabel>
        </div>
        <FieldDescription>{{
          t("ocservSync.assignConfigDescription")
        }}</FieldDescription>
      </Field>
    </FieldGroup>
    <OcservUserConfigFields
      v-if="form.assignConfig"
      :config="config"
      :disabled="submitting"
      @change="updateConfig"
    />
    <div class="flex flex-wrap justify-end gap-2">
      <Button
        variant="outline"
        :disabled="loading || submitting"
        @click="load()"
      >
        <RefreshCw data-icon="inline-start" />
        {{ t("ocservSync.refresh") }}
      </Button>
      <Button
        :disabled="loading || submitting || !users.length"
        @click="submit"
      >
        <Spinner v-if="submitting" data-icon="inline-start" />
        {{ t("ocservSync.syncSelected") }}
      </Button>
    </div>
  </div>
</template>

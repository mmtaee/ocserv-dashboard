<script setup lang="ts">
import { Download, RotateCcw, Upload } from "@lucide/vue";
import { reactive, shallowRef } from "vue";
import { useI18n } from "vue-i18n";

import { normalizeApiError } from "@/api/http";
import {
  backupFileName,
  backupKinds,
  downloadBackup,
  restoreBackup,
  type BackupKind,
  type RestoreResponse,
} from "@/api/services/backups";
import { Alert, AlertDescription, AlertTitle } from "@/components/ui/alert";
import {
  AlertDialog,
  AlertDialogAction,
  AlertDialogCancel,
  AlertDialogContent,
  AlertDialogDescription,
  AlertDialogFooter,
  AlertDialogHeader,
  AlertDialogTitle,
} from "@/components/ui/alert-dialog";
import { Button } from "@/components/ui/button";
import {
  Card,
  CardContent,
  CardDescription,
  CardFooter,
  CardHeader,
  CardTitle,
} from "@/components/ui/card";
import {
  Empty,
  EmptyDescription,
  EmptyHeader,
  EmptyTitle,
} from "@/components/ui/empty";
import {
  Field,
  FieldDescription,
  FieldError,
  FieldGroup,
  FieldLabel,
} from "@/components/ui/field";
import { Input } from "@/components/ui/input";
import { Spinner } from "@/components/ui/spinner";

const { t } = useI18n({ useScope: "global" });
const files = reactive<Record<BackupKind, File | null>>({
  ocserv_users: null,
  ocserv_groups: null,
});
const downloading = shallowRef<BackupKind | null>(null);
const restoring = shallowRef<BackupKind | null>(null);
const pendingRestore = shallowRef<BackupKind | null>(null);
const error = shallowRef("");
const validation = reactive<Record<BackupKind, string>>({
  ocserv_users: "",
  ocserv_groups: "",
});
const result = shallowRef<RestoreResponse | null>(null);

function onFile(kind: BackupKind, event: Event): void {
  files[kind] = (event.target as HTMLInputElement).files?.[0] ?? null;
  validation[kind] = "";
}

function requestRestore(kind: BackupKind): void {
  const file = files[kind];
  if (!file) {
    validation[kind] = t("backups.fileRequired");
    return;
  }
  if (!file.name.endsWith(".json") && !file.name.endsWith(".json.gz")) {
    validation[kind] = t("backups.invalidFile");
    return;
  }
  pendingRestore.value = kind;
}

async function download(kind: BackupKind): Promise<void> {
  downloading.value = kind;
  error.value = "";
  try {
    const blob = await downloadBackup(kind);
    const url = URL.createObjectURL(blob);
    const link = document.createElement("a");
    link.href = url;
    link.download = backupFileName(kind);
    link.click();
    URL.revokeObjectURL(url);
  } catch (cause) {
    error.value = normalizeApiError(cause).message;
  } finally {
    downloading.value = null;
  }
}

async function restore(): Promise<void> {
  const kind = pendingRestore.value;
  const file = kind ? files[kind] : null;
  if (!kind || !file) return;
  restoring.value = kind;
  error.value = "";
  result.value = null;
  try {
    result.value = await restoreBackup(kind, file);
    pendingRestore.value = null;
  } catch (cause) {
    error.value = normalizeApiError(cause).message;
  } finally {
    restoring.value = null;
  }
}
</script>

<template>
  <div class="flex flex-col gap-6">
    <Alert v-if="error" variant="error">
      <AlertTitle>{{ t("backups.requestFailed") }}</AlertTitle>
      <AlertDescription>{{ error }}</AlertDescription>
    </Alert>

    <Alert v-if="result" variant="success">
      <AlertTitle>{{ t("backups.restoreComplete") }}</AlertTitle>
      <AlertDescription>
        {{
          t("backups.restoreSummary", {
            inserted: result.inserted?.length ?? 0,
            existing: result.existing?.length ?? 0,
          })
        }}
      </AlertDescription>
    </Alert>

    <div>
      <h1 class="text-2xl font-semibold tracking-tight">
        {{ t("navigation.backup") }}
      </h1>
      <p class="text-sm text-muted-foreground">
        {{ t("backups.description") }}
      </p>
    </div>

    <Card v-for="kind in backupKinds" :key="kind">
      <CardHeader>
        <CardTitle>{{ t(`backups.kinds.${kind}`) }}</CardTitle>
        <CardDescription>{{
          t(`backups.kinds.${kind}Description`)
        }}</CardDescription>
      </CardHeader>
      <CardContent>
        <FieldGroup>
          <Field>
            <FieldLabel :for="`restore-${kind}`">
              {{ t("backups.restoreFile") }}
            </FieldLabel>
            <Input
              :id="`restore-${kind}`"
              type="file"
              accept=".json,.json.gz,application/json,application/gzip"
              :disabled="Boolean(restoring)"
              @change="onFile(kind, $event)"
            />
            <FieldDescription>{{ t("backups.fileHelp") }}</FieldDescription>
          </Field>
          <FieldError v-if="validation[kind]">{{
            validation[kind]
          }}</FieldError>
        </FieldGroup>
      </CardContent>
      <CardFooter class="flex flex-wrap gap-2">
        <Button
          type="button"
          variant="outline"
          :disabled="Boolean(downloading || restoring)"
          @click="download(kind)"
        >
          <Spinner v-if="downloading === kind" data-icon="inline-start" />
          <Download v-else data-icon="inline-start" />
          {{ t("backups.download") }}
        </Button>
        <Button
          type="button"
          variant="destructive"
          :disabled="Boolean(downloading || restoring)"
          @click="requestRestore(kind)"
        >
          <Spinner v-if="restoring === kind" data-icon="inline-start" />
          <Upload v-else data-icon="inline-start" />
          {{ t("backups.restore") }}
        </Button>
      </CardFooter>
    </Card>

    <Empty>
      <EmptyHeader>
        <EmptyTitle>{{ t("backups.noRemoteList") }}</EmptyTitle>
        <EmptyDescription>{{
          t("backups.noRemoteListDescription")
        }}</EmptyDescription>
      </EmptyHeader>
    </Empty>

    <AlertDialog :open="pendingRestore !== null">
      <AlertDialogContent>
        <AlertDialogHeader>
          <AlertDialogTitle>{{
            t("backups.restoreConfirmTitle")
          }}</AlertDialogTitle>
          <AlertDialogDescription>{{
            t("backups.restoreConfirmDescription")
          }}</AlertDialogDescription>
        </AlertDialogHeader>
        <AlertDialogFooter>
          <AlertDialogCancel
            :disabled="Boolean(restoring)"
            @click="pendingRestore = null"
          >
            {{ t("backups.cancel") }}
          </AlertDialogCancel>
          <AlertDialogAction :disabled="Boolean(restoring)" @click="restore">
            <Spinner v-if="Boolean(restoring)" data-icon="inline-start" />
            <RotateCcw v-else data-icon="inline-start" />
            {{ t("backups.restore") }}
          </AlertDialogAction>
        </AlertDialogFooter>
      </AlertDialogContent>
    </AlertDialog>
  </div>
</template>

<script setup lang="ts">
import { RefreshCw } from "@lucide/vue";
import { onMounted, shallowRef } from "vue";
import { useI18n } from "vue-i18n";

import { normalizeApiError } from "@/api/http";
import {
  getRuntimeStatus,
  runRuntimeAction,
  type OcservRuntimeAction,
  type OcservRuntimeStatus,
} from "@/api/services/system";
import RuntimeActionsCard from "@/components/runtime/RuntimeActionsCard.vue";
import RuntimeStatusCard from "@/components/runtime/RuntimeStatusCard.vue";
import { Alert, AlertDescription, AlertTitle } from "@/components/ui/alert";
import { Button } from "@/components/ui/button";
import { Spinner } from "@/components/ui/spinner";

const { t } = useI18n({ useScope: "global" });
const status = shallowRef<OcservRuntimeStatus | null>(null);
const loading = shallowRef(true);
const pending = shallowRef<OcservRuntimeAction | null>(null);
const error = shallowRef("");
const success = shallowRef("");

async function loadStatus(): Promise<void> {
  loading.value = true;
  error.value = "";
  try {
    status.value = await getRuntimeStatus();
  } catch (cause) {
    status.value = null;
    error.value = normalizeApiError(cause).message;
  } finally {
    loading.value = false;
  }
}

async function execute(action: OcservRuntimeAction): Promise<void> {
  if (!status.value?.allowed_actions?.includes(action)) {
    error.value = t("runtime.actionUnavailable");
    return;
  }

  pending.value = action;
  error.value = "";
  success.value = "";
  try {
    const response = await runRuntimeAction(action);
    success.value = response.message;
    await loadStatus();
  } catch (cause) {
    error.value = normalizeApiError(cause).message;
  } finally {
    pending.value = null;
  }
}

function reportUnavailable(): void {
  error.value = t("runtime.actionUnavailable");
}

onMounted(loadStatus);
</script>

<template>
  <div class="flex flex-col gap-6">
    <div class="flex flex-wrap items-start justify-between gap-3">
      <div class="flex flex-col gap-1">
        <h1 class="text-2xl font-semibold tracking-tight">
          {{ t("runtime.title") }}
        </h1>
        <p class="text-sm text-muted-foreground">
          {{ t("runtime.description") }}
        </p>
      </div>
      <Button
        type="button"
        variant="outline"
        :disabled="loading"
        @click="loadStatus"
      >
        <Spinner v-if="loading" data-icon="inline-start" />
        <RefreshCw v-else data-icon="inline-start" />
        {{ t("runtime.refresh") }}
      </Button>
    </div>

    <Alert v-if="error" variant="error">
      <AlertTitle>{{ t("runtime.errorTitle") }}</AlertTitle>
      <AlertDescription>{{ error }}</AlertDescription>
    </Alert>
    <Alert v-if="success" variant="success">
      <AlertTitle>{{ t("runtime.successTitle") }}</AlertTitle>
      <AlertDescription>{{ success }}</AlertDescription>
    </Alert>

    <RuntimeStatusCard :loading="loading" :status="status" />
    <RuntimeActionsCard
      :allowed-actions="status?.allowed_actions ?? []"
      :loading="loading"
      :pending="pending"
      @execute="execute"
      @unavailable="reportUnavailable"
    />
  </div>
</template>

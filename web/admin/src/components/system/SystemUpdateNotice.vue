<script setup lang="ts">
import { Download, RefreshCw } from "@lucide/vue";
import { computed, onMounted, onUnmounted, shallowRef } from "vue";
import { storeToRefs } from "pinia";
import { useI18n } from "vue-i18n";

import { normalizeApiError } from "@/api/http";
import {
  getSystemUpdateStatus,
  startSystemUpdate,
  type SystemUpdateStatus,
} from "@/api/services/system";
import { Alert, AlertDescription, AlertTitle } from "@/components/ui/alert";
import { Button } from "@/components/ui/button";
import { Spinner } from "@/components/ui/spinner";
import { useAuthStore } from "@/stores/auth";
import { useSystemInitStore } from "@/stores/system-init";

const { t } = useI18n({ useScope: "global" });
const auth = useAuthStore();
const systemInit = useSystemInitStore();
const { latestVersion, updateAvailable } = storeToRefs(systemInit);
const status = shallowRef<SystemUpdateStatus>({ state: "idle" });
const error = shallowRef("");
let pollTimer: ReturnType<typeof setTimeout> | undefined;

const visible = computed(
  () =>
    auth.user?.superadmin === true &&
    (updateAvailable.value || status.value.state !== "idle"),
);
const running = computed(() => status.value.state === "running");
const variant = computed(() =>
  status.value.state === "failed"
    ? "error"
    : status.value.state === "succeeded"
      ? "success"
      : "warning",
);
const title = computed(() => {
  if (status.value.state === "running") return t("systemUpdate.installing");
  if (status.value.state === "succeeded") return t("systemUpdate.succeeded");
  if (status.value.state === "failed") return t("systemUpdate.failed");
  return t("systemUpdate.available", { version: latestVersion.value });
});

function schedulePoll(): void {
  pollTimer = setTimeout(poll, 2000);
}

async function poll(): Promise<void> {
  try {
    status.value = await getSystemUpdateStatus();
    error.value = "";
    if (status.value.state === "succeeded") {
      pollTimer = setTimeout(() => window.location.reload(), 1500);
      return;
    }
    if (status.value.state === "running") schedulePoll();
  } catch {
    error.value = t("systemUpdate.reconnecting");
    schedulePoll();
  }
}

async function start(): Promise<void> {
  error.value = "";
  try {
    status.value = await startSystemUpdate();
    schedulePoll();
  } catch (cause) {
    const apiError = normalizeApiError(cause);
    if (apiError.status === 409) {
      status.value = { state: "running", target: latestVersion.value };
      schedulePoll();
      return;
    }
    error.value = apiError.message;
  }
}

onMounted(async () => {
  if (!updateAvailable.value) return;
  try {
    const current = await getSystemUpdateStatus();
    if (current.state === "running" || current.target === latestVersion.value) {
      status.value = current;
      if (current.state === "running") schedulePoll();
    }
  } catch {
    // The availability banner still works when status recovery is unavailable.
  }
});
onUnmounted(() => clearTimeout(pollTimer));
</script>

<template>
  <Alert v-if="visible" :variant="variant" class="mb-4">
    <AlertTitle>{{ title }}</AlertTitle>
    <AlertDescription class="flex flex-wrap items-center justify-between gap-3">
      <span role="status">
        {{
          error ||
          status.message ||
          (running
            ? t("systemUpdate.progress")
            : status.state === "succeeded"
              ? t("systemUpdate.reloading")
              : t("systemUpdate.description"))
        }}
      </span>
      <Button
        v-if="status.state === 'idle' || status.state === 'failed'"
        type="button"
        size="sm"
        :disabled="running"
        @click="start"
      >
        <Download data-icon="inline-start" />
        {{ t("systemUpdate.action") }}
      </Button>
      <Button v-else-if="running" type="button" size="sm" disabled>
        <Spinner data-icon="inline-start" />
        {{ t("systemUpdate.installing") }}
      </Button>
      <RefreshCw v-else class="size-4 animate-spin" aria-hidden="true" />
    </AlertDescription>
  </Alert>
</template>

<script setup lang="ts">
import { RefreshCw } from "@lucide/vue";
import { computed, onMounted, shallowRef } from "vue";
import { useI18n } from "vue-i18n";

import { normalizeApiError } from "@/api/http";
import {
  executeOcctlCommand,
  getOcctlServerInfo,
  type OcctlCommandRequest,
  type OcctlServerInfo as OcctlServerInfoData,
} from "@/api/services/occtl";
import OcctlCommandPanel from "@/components/occtl/OcctlCommandPanel.vue";
import OcctlConfirmDialog from "@/components/occtl/OcctlConfirmDialog.vue";
import { getOcctlCommand } from "@/components/occtl/occtl-commands";
import OcctlResultViewer from "@/components/occtl/OcctlResultViewer.vue";
import OcctlServerInfo from "@/components/occtl/OcctlServerInfo.vue";
import { Alert, AlertDescription, AlertTitle } from "@/components/ui/alert";
import { Button } from "@/components/ui/button";
import {
  Card,
  CardContent,
  CardDescription,
  CardHeader,
  CardTitle,
} from "@/components/ui/card";
import { Spinner } from "@/components/ui/spinner";

const { t } = useI18n({ useScope: "global" });
const serverInfo = shallowRef<OcctlServerInfoData | null>(null);
const serverLoading = shallowRef(true);
const serverError = shallowRef("");
const commandPending = shallowRef(false);
const commandError = shallowRef("");
const commandSuccess = shallowRef("");
const result = shallowRef<string | null>(null);
const confirmationOpen = shallowRef(false);
const confirmationRequest = shallowRef<OcctlCommandRequest | null>(null);

const confirmationLabel = computed(() => {
  if (!confirmationRequest.value) return "";
  const command = getOcctlCommand(confirmationRequest.value.action);
  return t(`occtl.commands.${command.labelKey}`);
});

async function loadServerInfo(): Promise<void> {
  serverLoading.value = true;
  serverError.value = "";
  try {
    serverInfo.value = await getOcctlServerInfo();
  } catch (cause) {
    serverInfo.value = null;
    serverError.value = normalizeApiError(cause).message;
  } finally {
    serverLoading.value = false;
  }
}

async function runCommand(request: OcctlCommandRequest): Promise<void> {
  commandPending.value = true;
  commandError.value = "";
  commandSuccess.value = "";
  try {
    result.value = await executeOcctlCommand(request);
    commandSuccess.value = t("occtl.commandSuccess");
    confirmationOpen.value = false;
    confirmationRequest.value = null;
    await loadServerInfo();
  } catch (cause) {
    commandError.value = normalizeApiError(cause).message;
  } finally {
    commandPending.value = false;
  }
}

function requestCommand(request: OcctlCommandRequest): void {
  const command = getOcctlCommand(request.action);
  if (command.destructive) {
    confirmationRequest.value = request;
    confirmationOpen.value = true;
    return;
  }
  void runCommand(request);
}

function confirmCommand(): void {
  if (confirmationRequest.value) void runCommand(confirmationRequest.value);
}

onMounted(loadServerInfo);
</script>

<template>
  <div class="flex flex-col gap-6">
    <div class="flex flex-wrap items-start justify-between gap-3">
      <div class="flex flex-col gap-1">
        <h1 class="text-2xl font-semibold tracking-tight">
          {{ t("navigation.occtl") }}
        </h1>
        <p class="text-sm text-muted-foreground">
          {{ t("occtl.description") }}
        </p>
      </div>
      <Button
        type="button"
        variant="outline"
        :disabled="serverLoading"
        @click="loadServerInfo"
      >
        <Spinner v-if="serverLoading" data-icon="inline-start" />
        <RefreshCw v-else data-icon="inline-start" />
        {{ t("occtl.refresh") }}
      </Button>
    </div>

    <Alert v-show="serverError" variant="destructive">
      <AlertTitle>{{ t("occtl.serverRequestFailure") }}</AlertTitle>
      <AlertDescription>{{ serverError }}</AlertDescription>
    </Alert>
    <Alert v-if="commandError" variant="destructive">
      <AlertTitle>{{ t("occtl.commandFailure") }}</AlertTitle>
      <AlertDescription>{{ commandError }}</AlertDescription>
    </Alert>
    <Alert v-if="commandSuccess">
      <AlertTitle>{{ t("occtl.success") }}</AlertTitle>
      <AlertDescription>{{ commandSuccess }}</AlertDescription>
    </Alert>

    <OcctlServerInfo :info="serverInfo" :loading="serverLoading" />

    <div class="grid gap-6 xl:grid-cols-[minmax(20rem,0.8fr)_minmax(0,1.2fr)]">
      <OcctlCommandPanel :pending="commandPending" @execute="requestCommand" />
      <Card>
        <CardHeader>
          <CardTitle>{{ t("occtl.commandResult") }}</CardTitle>
          <CardDescription>
            {{ t("occtl.commandResultDescription") }}
          </CardDescription>
        </CardHeader>
        <CardContent>
          <OcctlResultViewer :result="result" />
        </CardContent>
      </Card>
    </div>

    <OcctlConfirmDialog
      v-model:open="confirmationOpen"
      :command-label="confirmationLabel"
      :pending="commandPending"
      :request="confirmationRequest"
      @confirm="confirmCommand"
    />
  </div>
</template>

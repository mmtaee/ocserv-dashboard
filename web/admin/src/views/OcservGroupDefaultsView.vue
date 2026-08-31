<script setup lang="ts">
import { onMounted, shallowRef } from "vue";
import { useI18n } from "vue-i18n";

import { normalizeApiError } from "@/api/http";
import {
  getOcservDefaultGroupConfig,
  type OcservDefaultGroupConfig,
  type OcservDefaultGroupUpdate,
  updateOcservDefaultGroupConfig,
} from "@/api/services/ocserv-groups";
import DashboardLayout from "@/components/DashboardLayout.vue";
import OcservGroupDefaultsForm from "@/components/ocserv-groups/OcservGroupDefaultsForm.vue";
import { Alert, AlertDescription, AlertTitle } from "@/components/ui/alert";
import { Button } from "@/components/ui/button";
import { Card, CardContent, CardHeader } from "@/components/ui/card";
import { Skeleton } from "@/components/ui/skeleton";

const { t } = useI18n({ useScope: "global" });
const config = shallowRef<OcservDefaultGroupConfig | null>(null);
const loading = shallowRef(true);
const saving = shallowRef(false);
const error = shallowRef<string | null>(null);
const saved = shallowRef(false);

async function load(): Promise<void> {
  loading.value = true;
  error.value = null;
  saved.value = false;
  try {
    config.value = await getOcservDefaultGroupConfig();
  } catch (cause) {
    config.value = null;
    error.value = normalizeApiError(cause).message;
  } finally {
    loading.value = false;
  }
}

async function save(request: OcservDefaultGroupUpdate): Promise<void> {
  saving.value = true;
  error.value = null;
  saved.value = false;
  try {
    await updateOcservDefaultGroupConfig(request);
    config.value = request.config;
    saved.value = true;
  } catch (cause) {
    error.value = normalizeApiError(cause).message;
  } finally {
    saving.value = false;
  }
}

onMounted(load);
</script>

<template>
  <DashboardLayout>
    <div class="mx-auto flex w-full max-w-5xl flex-col gap-6">
      <div class="flex flex-col gap-1">
        <h1 class="text-2xl font-semibold tracking-tight">
          {{ t("navigation.groupDefaults") }}
        </h1>
        <p class="text-sm text-muted-foreground">
          {{ t("groupDefaults.description") }}
        </p>
      </div>

      <Alert v-if="error" variant="destructive">
        <AlertTitle>
          {{
            t(
              config
                ? "groupDefaults.saveFailure"
                : "groupDefaults.loadFailure",
            )
          }}
        </AlertTitle>
        <AlertDescription
          class="flex flex-wrap items-center justify-between gap-3"
        >
          <span>
            {{ error }}
          </span>
          <Button v-if="!config" type="button" variant="outline" @click="load">
            {{ t("unavailable.retry") }}
          </Button>
        </AlertDescription>
      </Alert>

      <Alert v-if="saved">
        <AlertTitle>
          {{ t("groupDefaults.saveSuccess") }}
        </AlertTitle>
      </Alert>

      <Card v-if="loading" aria-busy="true">
        <CardHeader class="gap-3">
          <Skeleton class="h-6 w-48" />
          <Skeleton class="h-4 w-3/4" />
        </CardHeader>
        <CardContent class="grid gap-5 md:grid-cols-2">
          <Skeleton v-for="index in 6" :key="index" class="h-16 w-full" />
        </CardContent>
      </Card>

      <OcservGroupDefaultsForm
        v-else-if="config"
        :config="config"
        :pending="saving"
        @submit="save"
      />
    </div>
  </DashboardLayout>
</template>

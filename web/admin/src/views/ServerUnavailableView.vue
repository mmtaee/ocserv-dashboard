<script setup lang="ts">
import { ServerCrash } from "@lucide/vue";
import { useRouter } from "vue-router";
import { useI18n } from "vue-i18n";

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
import LanguageSwitcher from "@/components/LanguageSwitcher.vue";
import { useAuthStore } from "@/stores/auth";
import { useSystemInitStore } from "@/stores/system-init";

const router = useRouter();
const auth = useAuthStore();
const systemInit = useSystemInitStore();
const { t } = useI18n({ useScope: "global" });

async function retry(): Promise<void> {
  if (!(await systemInit.initialize())) return;

  await auth.restoreSession();
  await router.replace({
    name: !auth.isAuthenticated
      ? "login"
      : systemInit.isInitialized
        ? "home"
        : "system-setup",
  });
}
</script>

<template>
  <main class="grid min-h-svh place-items-center bg-muted/40 p-6">
    <LanguageSwitcher class="absolute end-6 top-6" />
    <Card class="w-full max-w-lg">
      <CardHeader>
        <div
          class="mb-2 grid size-11 place-items-center rounded-full bg-destructive/10 text-destructive"
        >
          <ServerCrash aria-hidden="true" />
        </div>
        <CardTitle>{{ t("unavailable.title") }}</CardTitle>
        <CardDescription>{{ t("unavailable.description") }}</CardDescription>
      </CardHeader>
      <CardContent class="grid gap-4">
        <Alert v-if="systemInit.error" variant="destructive">
          <AlertTitle>{{ t("unavailable.connectionFailed") }}</AlertTitle>
          <AlertDescription>{{ systemInit.error }}</AlertDescription>
        </Alert>
        <Button :disabled="systemInit.isLoading" @click="retry">
          <Spinner v-if="systemInit.isLoading" data-icon="inline-start" />
          {{ t("unavailable.retry") }}
        </Button>
      </CardContent>
    </Card>
  </main>
</template>

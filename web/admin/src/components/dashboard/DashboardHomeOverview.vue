<script setup lang="ts">
import type { DeepReadonly } from "vue";
import { useI18n } from "vue-i18n";

import type { DashboardOverview } from "@/api/services/dashboard";
import TelegramStatus from "@/components/dashboard/TelegramStatus.vue";
import { Alert, AlertDescription } from "@/components/ui/alert";
import { useSystemInitStore } from "@/stores/system-init";

const props = defineProps<{
  overview: DeepReadonly<DashboardOverview> | null;
  loading: boolean;
  error: boolean;
}>();

const { t } = useI18n({ useScope: "global" });
const systemInit = useSystemInitStore();
</script>

<template>
  <section class="flex flex-col gap-6">
    <Alert v-if="error" variant="error">
      <AlertDescription>{{
        t("dashboard.homeOverviewError")
      }}</AlertDescription>
    </Alert>
    <TelegramStatus
      :available="systemInit.telegramBotEnabled"
      :service="props.overview?.telegram_service ?? null"
      :loading="systemInit.telegramBotEnabled && loading"
    />
  </section>
</template>

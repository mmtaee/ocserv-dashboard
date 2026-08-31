<script setup lang="ts">
import { Bot, KeyRound, Send } from "@lucide/vue";
import { useI18n } from "vue-i18n";

import type { TelegramService } from "@/api/services/dashboard";
import { Badge } from "@/components/ui/badge";
import {
  Card,
  CardContent,
  CardDescription,
  CardHeader,
  CardTitle,
} from "@/components/ui/card";
import { Skeleton } from "@/components/ui/skeleton";

defineProps<{
  service: TelegramService | null;
  loading: boolean;
}>();

const { t } = useI18n({ useScope: "global" });
</script>

<template>
  <Card v-if="loading || service">
    <CardHeader>
      <CardTitle class="flex items-center gap-2">
        <Send />
        {{ t("dashboard.telegramStatus") }}
      </CardTitle>
      <CardDescription>
        {{ t("dashboard.telegramStatusDescription") }}
      </CardDescription>
    </CardHeader>
    <CardContent>
      <div v-if="loading && !service" class="flex flex-wrap gap-3">
        <Skeleton class="h-6 w-24 rounded-full" />
        <Skeleton class="h-6 w-32 rounded-full" />
        <Skeleton class="h-6 w-40 rounded-full" />
      </div>
      <div v-else class="flex flex-wrap gap-3">
        <Badge :variant="service?.enabled ? 'default' : 'secondary'">
          <Bot />
          {{
            service?.enabled
              ? t("dashboard.telegramEnabled")
              : t("dashboard.telegramDisabled")
          }}
        </Badge>
        <Badge :variant="service?.has_bot_token ? 'outline' : 'destructive'">
          <KeyRound />
          {{
            service?.has_bot_token
              ? t("dashboard.telegramTokenConfigured")
              : t("dashboard.telegramTokenMissing")
          }}
        </Badge>
        <Badge variant="outline">
          <Send />
          {{
            service?.bot_username
              ? `@${service.bot_username}`
              : t("dashboard.telegramUsernameMissing")
          }}
        </Badge>
      </div>
    </CardContent>
  </Card>
</template>

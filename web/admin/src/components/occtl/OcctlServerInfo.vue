<script setup lang="ts">
import { Server } from "@lucide/vue";
import { useI18n } from "vue-i18n";

import type { OcctlServerInfo } from "@/api/services/occtl";
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
  info: OcctlServerInfo | null;
  loading?: boolean;
}>();
const { t } = useI18n({ useScope: "global" });
</script>

<template>
  <Card>
    <CardHeader>
      <CardTitle class="flex items-center gap-2">
        <Server />
        {{ t("occtl.serverInformation") }}
      </CardTitle>
      <CardDescription>
        {{ t("occtl.serverInformationDescription") }}
      </CardDescription>
    </CardHeader>
    <CardContent>
      <div v-if="loading" class="grid gap-3 sm:grid-cols-3" aria-busy="true">
        <Skeleton v-for="index in 3" :key="index" class="h-16 w-full" />
      </div>
      <div v-else-if="info" class="grid gap-3 sm:grid-cols-3">
        <div class="flex flex-col gap-2 rounded-lg border p-4">
          <span class="text-sm text-muted-foreground">
            {{ t("occtl.status") }}
          </span>
          <Badge
            :variant="info.status === 'online' ? 'default' : 'destructive'"
          >
            {{ info.status }}
          </Badge>
        </div>
        <div class="flex flex-col gap-2 rounded-lg border p-4">
          <span class="text-sm text-muted-foreground">
            {{ t("occtl.ocservVersion") }}
          </span>
          <span class="font-mono">
            {{ info.version.ocserv_version || "—" }}
          </span>
        </div>
        <div class="flex flex-col gap-2 rounded-lg border p-4">
          <span class="text-sm text-muted-foreground">
            {{ t("occtl.occtlVersion") }}
          </span>
          <span class="font-mono">{{ info.version.occtl_version || "—" }}</span>
        </div>
      </div>
    </CardContent>
  </Card>
</template>

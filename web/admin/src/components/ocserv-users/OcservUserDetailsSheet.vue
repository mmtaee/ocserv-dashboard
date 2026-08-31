<script setup lang="ts">
import { computed, shallowRef, watch } from "vue";
import { useI18n } from "vue-i18n";

import { normalizeApiError } from "@/api/http";
import type {
  OcservUser,
  OcservUserSessionLogsResponse,
  OcservUserStatisticsResponse,
} from "@/api/services/ocserv-users";
import {
  getOcservUserSessionLogs,
  getOcservUserStatistics,
} from "@/api/services/ocserv-users";
import { Alert, AlertDescription, AlertTitle } from "@/components/ui/alert";
import { Badge } from "@/components/ui/badge";
import { Button } from "@/components/ui/button";
import { Separator } from "@/components/ui/separator";
import {
  Sheet,
  SheetContent,
  SheetDescription,
  SheetHeader,
  SheetTitle,
} from "@/components/ui/sheet";
import { Skeleton } from "@/components/ui/skeleton";
import {
  Table,
  TableBody,
  TableCell,
  TableHead,
  TableHeader,
  TableRow,
} from "@/components/ui/table";

const props = defineProps<{
  loading?: boolean;
  pending?: boolean;
  user?: OcservUser | null;
}>();
const emit = defineEmits<{
  sessionAction: [action: "disconnect" | "terminate", id: number];
}>();
const open = defineModel<boolean>("open", { default: false });
const { locale, t } = useI18n({ useScope: "global" });
const activityLoading = shallowRef(false);
const activityError = shallowRef("");
const logs = shallowRef<OcservUserSessionLogsResponse | null>(null);
const statistics = shallowRef<OcservUserStatisticsResponse | null>(null);
const configuredFields = computed(() =>
  Object.entries(props.user?.config ?? {}).filter(([, value]) => value != null),
);

watch(
  [open, () => props.user?.id],
  async ([isOpen, id]) => {
    if (!isOpen || typeof id !== "number") return;
    activityLoading.value = true;
    activityError.value = "";
    try {
      const [logsResponse, statisticsResponse] = await Promise.all([
        getOcservUserSessionLogs(id, { page: 1, size: 10 }),
        getOcservUserStatistics(id),
      ]);
      logs.value = logsResponse;
      statistics.value = statisticsResponse;
    } catch (cause) {
      activityError.value = normalizeApiError(cause).message;
    } finally {
      activityLoading.value = false;
    }
  },
  { immediate: true },
);

function formatDate(value?: string): string {
  if (!value) return "—";
  return new Intl.DateTimeFormat(locale.value, {
    dateStyle: "medium",
    timeStyle: "short",
  }).format(new Date(value));
}

function formatConfigValue(value: unknown): string {
  return Array.isArray(value) ? value.join(", ") : String(value);
}
</script>

<template>
  <Sheet v-model:open="open">
    <SheetContent class="overflow-y-auto sm:max-w-4xl">
      <SheetHeader>
        <SheetTitle>{{ t("ocservUsers.detailsTitle") }}</SheetTitle>
        <SheetDescription>{{
          t("ocservUsers.detailsDescription")
        }}</SheetDescription>
      </SheetHeader>

      <div v-if="loading || !user" class="grid gap-4 px-4" aria-busy="true">
        <Skeleton v-for="index in 6" :key="index" class="h-16 w-full" />
      </div>

      <div v-else class="grid gap-6 px-4 pb-6">
        <div class="grid gap-3 sm:grid-cols-2 lg:grid-cols-4">
          <div class="rounded-lg border p-3">
            <p class="text-xs text-muted-foreground">
              {{ t("ocservUsers.username") }}
            </p>
            <p class="font-medium">{{ user.username }}</p>
          </div>
          <div class="rounded-lg border p-3">
            <p class="text-xs text-muted-foreground">
              {{ t("ocservUsers.group") }}
            </p>
            <p class="font-medium">{{ user.group || "—" }}</p>
          </div>
          <div class="rounded-lg border p-3">
            <p class="text-xs text-muted-foreground">
              {{ t("ocservUsers.rx") }}
            </p>
            <p class="font-mono">
              {{ (user.running_rx / 1024 ** 3).toFixed(2) }}
              {{ t("dashboard.gigabytes") }}
            </p>
          </div>
          <div class="rounded-lg border p-3">
            <p class="text-xs text-muted-foreground">
              {{ t("ocservUsers.tx") }}
            </p>
            <p class="font-mono">
              {{ (user.running_tx / 1024 ** 3).toFixed(2) }}
              {{ t("dashboard.gigabytes") }}
            </p>
          </div>
        </div>

        <section class="grid gap-3">
          <h3 class="font-medium">{{ t("ocservUsers.onlineSessions") }}</h3>
          <div
            v-if="user.online_sessions.length"
            class="overflow-hidden rounded-md border"
          >
            <Table>
              <TableHeader
                ><TableRow
                  ><TableHead>{{ t("ocservUsers.sessionId") }}</TableHead
                  ><TableHead>{{ t("ocservUsers.ip") }}</TableHead
                  ><TableHead>{{ t("ocservUsers.startedAt") }}</TableHead
                  ><TableHead class="text-end">{{
                    t("ocservUsers.actions")
                  }}</TableHead></TableRow
                ></TableHeader
              >
              <TableBody>
                <TableRow
                  v-for="session in user.online_sessions"
                  :key="session.ID"
                >
                  <TableCell>{{ session.ID }}</TableCell>
                  <TableCell>{{ session.IPv4 }}</TableCell>
                  <TableCell>{{
                    formatDate(session["Session started at"])
                  }}</TableCell>
                  <TableCell class="flex justify-end gap-2">
                    <Button
                      size="sm"
                      variant="outline"
                      :disabled="pending"
                      @click="emit('sessionAction', 'disconnect', session.ID)"
                      >{{ t("ocservUsers.disconnect") }}</Button
                    >
                    <Button
                      size="sm"
                      variant="destructive"
                      :disabled="pending"
                      @click="emit('sessionAction', 'terminate', session.ID)"
                      >{{ t("ocservUsers.terminate") }}</Button
                    >
                  </TableCell>
                </TableRow>
              </TableBody>
            </Table>
          </div>
          <p v-else class="text-sm text-muted-foreground">
            {{ t("ocservUsers.noSessions") }}
          </p>
        </section>

        <Separator />
        <section class="grid gap-3">
          <h3 class="font-medium">{{ t("ocservUsers.configuration") }}</h3>
          <div v-if="configuredFields.length" class="grid gap-2 sm:grid-cols-2">
            <div
              v-for="[key, value] in configuredFields"
              :key="key"
              class="rounded-lg border p-3"
            >
              <p class="text-xs text-muted-foreground">{{ key }}</p>
              <p class="break-words text-sm">{{ formatConfigValue(value) }}</p>
            </div>
          </div>
          <p v-else class="text-sm text-muted-foreground">
            {{ t("ocservUsers.noConfiguration") }}
          </p>
        </section>

        <Separator />
        <Alert v-if="activityError" variant="destructive">
          <AlertTitle>{{ t("ocservUsers.activityFailure") }}</AlertTitle>
          <AlertDescription>{{ activityError }}</AlertDescription>
        </Alert>
        <div v-if="activityLoading" class="grid gap-3">
          <Skeleton v-for="index in 4" :key="index" class="h-12 w-full" />
        </div>
        <template v-else>
          <section class="grid gap-3">
            <div class="flex items-center justify-between gap-3">
              <h3 class="font-medium">
                {{ t("ocservUsers.trafficStatistics") }}
              </h3>
              <div v-if="statistics" class="flex gap-2">
                <Badge variant="secondary"
                  >RX {{ statistics.total_bandwidths.rx.toFixed(2) }} GB</Badge
                >
                <Badge variant="secondary"
                  >TX {{ statistics.total_bandwidths.tx.toFixed(2) }} GB</Badge
                >
              </div>
            </div>
            <div
              v-if="statistics?.statistics.length"
              class="overflow-hidden rounded-md border"
            >
              <Table
                ><TableHeader
                  ><TableRow
                    ><TableHead>{{ t("ocservUsers.date") }}</TableHead
                    ><TableHead>{{ t("ocservUsers.rx") }}</TableHead
                    ><TableHead>{{ t("ocservUsers.tx") }}</TableHead></TableRow
                  ></TableHeader
                ><TableBody
                  ><TableRow
                    v-for="item in statistics.statistics"
                    :key="item.date"
                    ><TableCell>{{ item.date }}</TableCell
                    ><TableCell>{{ item.rx?.toFixed(2) }} GB</TableCell
                    ><TableCell
                      >{{ item.tx?.toFixed(2) }} GB</TableCell
                    ></TableRow
                  ></TableBody
                ></Table
              >
            </div>
            <p v-else class="text-sm text-muted-foreground">
              {{ t("ocservUsers.noStatistics") }}
            </p>
          </section>

          <section class="grid gap-3">
            <h3 class="font-medium">{{ t("ocservUsers.sessionLogs") }}</h3>
            <div
              v-if="logs?.result?.length"
              class="overflow-hidden rounded-md border"
            >
              <Table
                ><TableHeader
                  ><TableRow
                    ><TableHead>{{ t("ocservUsers.date") }}</TableHead
                    ><TableHead>{{ t("ocservUsers.event") }}</TableHead
                    ><TableHead>{{ t("ocservUsers.ip") }}</TableHead
                    ><TableHead>{{
                      t("ocservUsers.message")
                    }}</TableHead></TableRow
                  ></TableHeader
                ><TableBody
                  ><TableRow
                    v-for="log in logs.result"
                    :key="`${log.created_at}-${log.event}`"
                    ><TableCell>{{ formatDate(log.created_at) }}</TableCell
                    ><TableCell>{{ log.event }}</TableCell
                    ><TableCell>{{ log.ip || "—" }}</TableCell
                    ><TableCell>{{ log.message }}</TableCell></TableRow
                  ></TableBody
                ></Table
              >
            </div>
            <p v-else class="text-sm text-muted-foreground">
              {{ t("ocservUsers.noLogs") }}
            </p>
          </section>
        </template>
      </div>
    </SheetContent>
  </Sheet>
</template>

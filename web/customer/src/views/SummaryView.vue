<script setup lang="ts">
import { computed, onMounted, shallowRef } from "vue";
import { useI18n } from "vue-i18n";
import type { SummaryResponse } from "@/api/generated";
import { normalizeApiError } from "@/api/http";
import { getSummary } from "@/api/services/customers";
import StatusMessage from "@/components/StatusMessage.vue";
import DownloadsView from "@/views/DownloadsView.vue";
import PasswordView from "@/views/PasswordView.vue";

const { t } = useI18n({ useScope: "global" });
const data = shallowRef<SummaryResponse | null>(null);
const loading = shallowRef(true);
const error = shallowRef("");
const account = computed(() => data.value?.ocserv_user);
const accountInitials = computed(() => {
  const username = account.value?.username?.trim() ?? "";
  return username.slice(0, 2).toUpperCase() || "—";
});
const usageChart = computed(() => {
  const rx = data.value?.usage?.bandwidths?.rx ?? 0;
  const tx = data.value?.usage?.bandwidths?.tx ?? 0;
  const total = rx + tx;
  if (total <= 0) return null;
  const rxPercentage = (rx / total) * 100;
  return {
    rx,
    tx,
    total,
    style: {
      background:
        "conic-gradient(var(--primary) 0 " +
        rxPercentage +
        "%, #10b981 " +
        rxPercentage +
        "% 100%)",
    },
  };
});
async function load(): Promise<void> {
  loading.value = true;
  error.value = "";
  try {
    data.value = await getSummary();
  } catch (cause) {
    error.value = normalizeApiError(cause).message;
  } finally {
    loading.value = false;
  }
}
onMounted(load);
</script>

<template>
  <section class="flex flex-col gap-4">
    <h1 class="text-2xl font-semibold">{{ t("nav.home") }}</h1>
    <StatusMessage :error="error" />
    <p v-if="loading">{{ t("common.loading") }}</p>
    <div v-else-if="data" class="grid gap-4 lg:grid-cols-[1.1fr_0.9fr]">
      <article class="card relative overflow-hidden">
        <div
          class="pointer-events-none absolute -right-14 -top-14 size-40 rounded-full bg-primary/10"
          aria-hidden="true"
        />
        <div class="relative">
          <header class="flex flex-wrap items-center justify-between gap-4">
            <div class="flex items-center gap-3">
              <div
                class="grid size-13 shrink-0 place-items-center rounded-2xl bg-primary text-lg font-bold text-primary-foreground shadow-sm"
              >
                {{ accountInitials }}
              </div>
              <div class="min-w-0">
                <p class="text-sm text-muted-foreground">
                  {{ t("summary.username") }}
                </p>
                <h2 class="truncate text-xl font-semibold">
                  {{ account?.username ?? "—" }}
                </h2>
              </div>
            </div>
            <span
              class="rounded-full px-3 py-1 text-sm font-medium"
              :class="
                account?.is_locked
                  ? 'bg-destructive/10 text-destructive'
                  : 'bg-emerald-500/10 text-emerald-700'
              "
            >
              {{
                account?.is_locked ? t("summary.locked") : t("summary.active")
              }}
            </span>
          </header>

          <dl class="mt-7 grid gap-4 sm:grid-cols-2">
            <div class="rounded-xl bg-muted/60 p-3">
              <dt class="text-xs font-medium text-muted-foreground">
                {{ t("summary.owner") }}
              </dt>
              <dd class="mt-1 truncate font-medium">
                {{ account?.owner ?? "—" }}
              </dd>
            </div>
            <div class="rounded-xl bg-muted/60 p-3">
              <dt class="text-xs font-medium text-muted-foreground">
                {{ t("summary.expiry") }}
              </dt>
              <dd class="mt-1 font-medium">{{ account?.expire_at ?? "—" }}</dd>
            </div>
            <div class="rounded-xl bg-muted/60 p-3">
              <dt class="text-xs font-medium text-muted-foreground">
                {{ t("summary.trafficPlan") }}
              </dt>
              <dd class="mt-1 truncate font-medium">
                {{ account?.traffic_type ?? "—" }}
              </dd>
            </div>
            <div class="rounded-xl bg-muted/60 p-3">
              <dt class="text-xs font-medium text-muted-foreground">
                {{ t("summary.allowance") }}
              </dt>
              <dd class="mt-1 font-medium">
                {{
                  account?.traffic_size === undefined
                    ? "—"
                    : account.traffic_size + " GiB"
                }}
              </dd>
            </div>
          </dl>
        </div>
      </article>
      <article class="card flex flex-col">
        <div class="flex items-start justify-between gap-4">
          <div>
            <p class="text-sm font-medium text-muted-foreground">
              {{ t("statistics.total") }}
            </p>
            <p class="mt-1 text-3xl font-semibold tracking-tight">
              {{ usageChart?.total ?? 0 }}
              <span class="text-base font-medium text-muted-foreground"
                >GiB</span
              >
            </p>
          </div>
          <div
            class="grid size-10 place-items-center rounded-xl bg-primary/10 text-primary"
            aria-hidden="true"
          >
            ↕
          </div>
        </div>
        <div
          v-if="usageChart"
          class="mt-6 flex flex-wrap items-center gap-5"
          role="img"
          :aria-label="t('statistics.title')"
        >
          <div
            class="relative grid size-32 shrink-0 place-items-center rounded-full"
            :style="usageChart.style"
          >
            <div
              class="grid size-22 place-items-center rounded-full bg-card text-center"
            >
              <strong class="text-lg leading-none">{{
                usageChart.total
              }}</strong>
              <span class="text-[10px] text-muted-foreground">GiB</span>
            </div>
          </div>
          <div class="grid gap-2 text-sm">
            <span class="inline-flex items-center gap-2">
              <i class="size-2.5 rounded-full bg-primary" aria-hidden="true" />
              {{ t("summary.received") }} · {{ usageChart.rx }} GiB
            </span>
            <span class="inline-flex items-center gap-2">
              <i
                class="size-2.5 rounded-full bg-emerald-500"
                aria-hidden="true"
              />
              {{ t("summary.transmitted") }} · {{ usageChart.tx }} GiB
            </span>
          </div>
        </div>
        <div
          v-else
          class="mt-6 rounded-xl bg-muted/60 p-4 text-sm text-muted-foreground"
        >
          {{ t("common.empty") }}
        </div>
      </article>

      <section class="lg:col-span-2 grid gap-6 border-t pt-6">
        <DownloadsView />
        <PasswordView />
      </section>
    </div>
  </section>
</template>

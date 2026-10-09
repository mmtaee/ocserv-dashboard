<script setup lang="ts">
import { computed, onMounted, reactive, shallowRef } from "vue";
import { useI18n } from "vue-i18n";
import type { Bandwidth, DailyTraffic } from "@/api/generated";
import { normalizeApiError } from "@/api/http";
import { getBandwidth, getStats } from "@/api/services/customers";
import StatusMessage from "@/components/StatusMessage.vue";

const { t } = useI18n({ useScope: "global" });
const query = reactive({ date_start: "", date_end: "" });
const stats = shallowRef<DailyTraffic[]>([]);
const total = shallowRef<Bandwidth | null>(null);
const loading = shallowRef(false);
const error = shallowRef("");
const chartItems = computed(() => {
  const peak = Math.max(
    ...stats.value.map((item) => (item.rx ?? 0) + (item.tx ?? 0)),
    1,
  );
  return stats.value.map((item) => {
    const rx = item.rx ?? 0;
    const tx = item.tx ?? 0;
    const date = item.date ?? "";
    return {
      date,
      label: date.slice(5),
      rx,
      tx,
      rxHeight: `${(rx / peak) * 100}%`,
      txHeight: `${(tx / peak) * 100}%`,
    };
  });
});
async function load(): Promise<void> {
  loading.value = true;
  error.value = "";
  const params = {
    date_start: query.date_start || undefined,
    date_end: query.date_end || undefined,
  };
  try {
    [stats.value, total.value] = await Promise.all([
      getStats(params),
      getBandwidth(params),
    ]);
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
    <h1 class="text-2xl font-semibold">{{ t("statistics.title") }}</h1>
    <form class="card grid gap-3 sm:grid-cols-3" @submit.prevent="load">
      <label class="flex flex-col gap-1"
        ><span class="label">{{ t("activity.dateStart") }}</span
        ><input v-model="query.date_start" class="input" type="date"
      /></label>
      <label class="flex flex-col gap-1"
        ><span class="label">{{ t("activity.dateEnd") }}</span
        ><input v-model="query.date_end" class="input" type="date"
      /></label>
      <button class="button self-end" :disabled="loading">
        {{ t("common.apply") }}
      </button>
    </form>
    <StatusMessage :error="error" />
    <article v-if="total" class="card">
      <h2 class="font-semibold">{{ t("statistics.total") }}</h2>
      <p>
        {{ t("statistics.rx") }}: {{ total.rx }} · {{ t("statistics.tx") }}:
        {{ total.tx }}
      </p>
      <div
        v-if="chartItems.length"
        class="mt-5"
        role="img"
        :aria-label="t('statistics.title')"
      >
        <div
          class="mb-3 flex flex-wrap gap-x-4 gap-y-1 text-xs text-muted-foreground"
        >
          <span class="inline-flex items-center gap-1.5">
            <i class="size-2 rounded-sm bg-primary" aria-hidden="true" />
            {{ t("statistics.rx") }}
          </span>
          <span class="inline-flex items-center gap-1.5">
            <i class="size-2 rounded-sm bg-emerald-500" aria-hidden="true" />
            {{ t("statistics.tx") }}
          </span>
        </div>
        <div class="flex h-40 items-end gap-1 border-b border-border pt-2">
          <div
            v-for="item in chartItems"
            :key="item.date"
            class="flex h-full min-w-0 flex-1 flex-col justify-end"
            :title="
              item.date +
              ': ' +
              t('statistics.rx') +
              ' ' +
              item.rx +
              ' · ' +
              t('statistics.tx') +
              ' ' +
              item.tx
            "
          >
            <div
              class="min-h-px rounded-t-sm bg-emerald-500"
              :style="{ height: item.txHeight }"
            />
            <div
              class="min-h-px rounded-t-sm bg-primary"
              :style="{ height: item.rxHeight }"
            />
          </div>
        </div>
        <div
          class="mt-2 flex gap-1 text-center text-[10px] text-muted-foreground"
        >
          <span
            v-for="item in chartItems"
            :key="item.date + '-label'"
            class="min-w-0 flex-1 truncate"
          >
            {{ item.label }}
          </span>
        </div>
      </div>
    </article>
    <p v-if="loading">{{ t("common.loading") }}</p>
    <p v-else-if="!stats.length" class="card">{{ t("common.empty") }}</p>
    <div v-else class="card overflow-x-auto">
      <table class="table">
        <thead>
          <tr>
            <th>{{ t("statistics.date") }}</th>
            <th>{{ t("statistics.rx") }}</th>
            <th>{{ t("statistics.tx") }}</th>
          </tr>
        </thead>
        <tbody>
          <tr v-for="item in stats" :key="item.date">
            <td>{{ item.date }}</td>
            <td>{{ item.rx ?? 0 }}</td>
            <td>{{ item.tx ?? 0 }}</td>
          </tr>
        </tbody>
      </table>
    </div>
  </section>
</template>

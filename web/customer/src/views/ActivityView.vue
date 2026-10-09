<script setup lang="ts">
import { computed, onMounted, reactive, shallowRef } from "vue";
import { useI18n } from "vue-i18n";
import type { ActivitiesResponse } from "@/api/generated";
import { normalizeApiError } from "@/api/http";
import { getActivities } from "@/api/services/customers";
import StatusMessage from "@/components/StatusMessage.vue";

const { t } = useI18n({ useScope: "global" });
const query = reactive({ page: 1, size: 20, date_start: "", date_end: "" });
const data = shallowRef<ActivitiesResponse | null>(null);
const loading = shallowRef(false);
const error = shallowRef("");
const pages = computed(() =>
  Math.max(1, Math.ceil((data.value?.meta.total_records ?? 0) / query.size)),
);
async function load(): Promise<void> {
  loading.value = true;
  error.value = "";
  try {
    data.value = await getActivities({
      ...query,
      date_start: query.date_start || undefined,
      date_end: query.date_end || undefined,
    });
  } catch (cause) {
    error.value = normalizeApiError(cause).message;
  } finally {
    loading.value = false;
  }
}
async function move(page: number): Promise<void> {
  query.page = page;
  await load();
}
onMounted(load);
</script>

<template>
  <section class="flex flex-col gap-4">
    <h1 class="text-2xl font-semibold">{{ t("activity.title") }}</h1>
    <form
      class="card grid gap-3 sm:grid-cols-4"
      @submit.prevent="
        query.page = 1;
        load();
      "
    >
      <label class="flex flex-col gap-1"
        ><span class="label">{{ t("activity.dateStart") }}</span
        ><input v-model="query.date_start" class="input" type="date"
      /></label>
      <label class="flex flex-col gap-1"
        ><span class="label">{{ t("activity.dateEnd") }}</span
        ><input v-model="query.date_end" class="input" type="date"
      /></label>
      <label class="flex flex-col gap-1"
        ><span class="label">{{ t("activity.rows") }}</span
        ><select v-model.number="query.size" class="input">
          <option :value="10">10</option>
          <option :value="20">20</option>
          <option :value="50">50</option>
          <option :value="100">100</option>
        </select></label
      >
      <button class="button self-end" :disabled="loading">
        {{ t("common.apply") }}
      </button>
    </form>
    <StatusMessage :error="error" />
    <p v-if="loading">{{ t("common.loading") }}</p>
    <p v-else-if="!data?.result?.length" class="card">
      {{ t("common.empty") }}
    </p>
    <div v-else class="card overflow-x-auto">
      <table class="table">
        <thead>
          <tr>
            <th>{{ t("activity.time") }}</th>
            <th>{{ t("activity.event") }}</th>
            <th>{{ t("activity.ip") }}</th>
            <th>{{ t("activity.message") }}</th>
          </tr>
        </thead>
        <tbody>
          <tr v-for="item in data.result" :key="item.created_at + item.event">
            <td>{{ item.created_at }}</td>
            <td>{{ item.event }}</td>
            <td>{{ item.ip ?? "—" }}</td>
            <td>{{ item.message }}</td>
          </tr>
        </tbody>
      </table>
    </div>
    <div class="flex items-center justify-center gap-3">
      <button
        class="button-secondary"
        :disabled="loading || query.page <= 1"
        @click="move(query.page - 1)"
      >
        {{ t("common.previous") }}
      </button>
      <span>{{ t("common.page", { page: query.page, pages }) }}</span>
      <button
        class="button-secondary"
        :disabled="loading || query.page >= pages"
        @click="move(query.page + 1)"
      >
        {{ t("common.next") }}
      </button>
    </div>
  </section>
</template>

<script setup lang="ts">
import { onMounted, reactive, shallowRef } from "vue";
import { useI18n } from "vue-i18n";
import { normalizeApiError } from "@/api/http";
import {
  getReportTotalBandwidth,
  type TotalBandwidth,
} from "@/api/services/reports";
import BandwidthChart from "@/components/dashboard/BandwidthChart.vue";
import ReportsDateRangeFilter from "@/components/reports/ReportsDateRangeFilter.vue";
import { defaultReportRange } from "@/components/reports/report-dates";
import { Alert, AlertDescription, AlertTitle } from "@/components/ui/alert";

const { t } = useI18n({ useScope: "global" });
const range = reactive(defaultReportRange());
const total = shallowRef<TotalBandwidth | null>(null);
const loading = shallowRef(true);
const error = shallowRef("");
async function load(): Promise<void> {
  loading.value = true;
  error.value = "";
  try {
    total.value = await getReportTotalBandwidth(range);
  } catch (cause) {
    total.value = null;
    error.value = normalizeApiError(cause).message;
  } finally {
    loading.value = false;
  }
}
onMounted(load);
</script>

<template>
  <div class="flex flex-col gap-6">
    <div>
      <h1 class="text-2xl font-semibold tracking-tight">
        {{ t("reports.bandwidthTitle") }}
      </h1>
      <p class="text-sm text-muted-foreground">
        {{ t("reports.bandwidthDescription") }}
      </p>
    </div>
    <ReportsDateRangeFilter
      v-model:date-start="range.dateStart"
      v-model:date-end="range.dateEnd"
      :loading="loading"
      @apply="load"
    /><Alert v-if="error" variant="error"
      ><AlertTitle>{{ t("reports.requestFailed") }}</AlertTitle
      ><AlertDescription>{{ error }}</AlertDescription></Alert
    ><BandwidthChart :total="total" :loading="loading" />
  </div>
</template>

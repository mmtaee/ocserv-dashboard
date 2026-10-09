<script setup lang="ts">
import { onMounted, shallowRef } from "vue";
import { useI18n } from "vue-i18n";
import type { CiscoSetup, SummaryResponse } from "@/api/generated";
import { normalizeApiError } from "@/api/http";
import {
  downloadCertificate,
  downloadCiscoCertificate,
  getCiscoSetup,
  getSummary,
} from "@/api/services/customers";
import StatusMessage from "@/components/StatusMessage.vue";

const { t } = useI18n({ useScope: "global" });
const setup = shallowRef<CiscoSetup | null>(null);
const summary = shallowRef<SummaryResponse | null>(null);
const loading = shallowRef(true);
const error = shallowRef("");
async function load(): Promise<void> {
  loading.value = true;
  error.value = "";
  try {
    [setup.value, summary.value] = await Promise.all([
      getCiscoSetup(),
      getSummary(),
    ]);
  } catch (cause) {
    error.value = normalizeApiError(cause).message;
  } finally {
    loading.value = false;
  }
}
async function save(
  fetcher: () => Promise<Blob>,
  filename: string,
): Promise<void> {
  try {
    const url = URL.createObjectURL(await fetcher());
    const link = document.createElement("a");
    link.href = url;
    link.download = filename;
    link.click();
    URL.revokeObjectURL(url);
  } catch (cause) {
    error.value = normalizeApiError(cause).message;
  }
}
onMounted(load);
</script>

<template>
  <section class="flex flex-col gap-4">
    <h2 class="text-xl font-semibold">{{ t("downloads.title") }}</h2>
    <StatusMessage :error="error" />
    <p v-if="loading">{{ t("common.loading") }}</p>
    <div v-else class="grid gap-4 md:grid-cols-2">
      <article class="card flex flex-col gap-3">
        <h2 class="font-semibold">{{ t("downloads.certificate") }}</h2>
        <button
          class="button"
          :disabled="!summary?.ocserv_user?.certificate_available"
          @click="save(downloadCertificate, 'ocserv-certificate.p12')"
        >
          {{ t("common.download") }}
        </button>
        <p
          v-if="!summary?.ocserv_user?.certificate_available"
          class="text-sm text-muted-foreground"
        >
          {{ t("downloads.unavailable") }}
        </p>
      </article>
      <article class="card flex flex-col gap-3">
        <h2 class="font-semibold">{{ t("downloads.ciscoCertificate") }}</h2>
        <dl v-if="setup" class="grid grid-cols-2 gap-2 text-sm">
          <dt>{{ t("downloads.connectionName") }}</dt>
          <dd>{{ setup.connection_name }}</dd>
          <dt>{{ t("downloads.server") }}</dt>
          <dd>{{ setup.server_address }}:{{ setup.server_port }}</dd>
          <dt>{{ t("downloads.certificatePassword") }}</dt>
          <dd>{{ setup.certificate_password }}</dd>
        </dl>
        <div class="flex flex-wrap gap-2">
          <button
            class="button"
            @click="
              save(downloadCiscoCertificate, 'cisco-setup-certificate.p12')
            "
          >
            {{ t("common.download") }}
          </button>
          <a
            v-if="setup?.certificate_import_uri"
            class="button-secondary"
            :href="setup.certificate_import_uri"
            >{{ t("downloads.import") }}</a
          >
          <a
            v-if="setup?.connection_create_uri"
            class="button-secondary"
            :href="setup.connection_create_uri"
            >{{ t("downloads.connect") }}</a
          >
        </div>
      </article>
    </div>
  </section>
</template>

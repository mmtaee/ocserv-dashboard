<script setup lang="ts">
import { reactive, shallowRef } from "vue";
import { useI18n } from "vue-i18n";
import { normalizeApiError } from "@/api/http";
import { changePassword } from "@/api/services/customers";
import StatusMessage from "@/components/StatusMessage.vue";

const { t } = useI18n({ useScope: "global" });
const form = reactive({ password: "", confirmation: "" });
const loading = shallowRef(false);
const error = shallowRef("");
const success = shallowRef("");
async function submit(): Promise<void> {
  error.value = "";
  success.value = "";
  if (form.password.length < 2 || form.password.length > 32) {
    error.value = t("password.length");
    return;
  }
  if (form.password !== form.confirmation) {
    error.value = t("password.mismatch");
    return;
  }
  loading.value = true;
  try {
    await changePassword({ password: form.password });
    form.password = "";
    form.confirmation = "";
    success.value = t("common.success");
  } catch (cause) {
    error.value = normalizeApiError(cause).message;
  } finally {
    loading.value = false;
  }
}
</script>

<template>
  <section class="flex flex-col gap-4">
    <h2 class="text-xl font-semibold">{{ t("password.title") }}</h2>
    <StatusMessage :error="error" :success="success" />
    <form class="card flex flex-col gap-4" @submit.prevent="submit">
      <label class="flex flex-col gap-2"
        ><span class="label">{{ t("password.new") }}</span
        ><input
          v-model="form.password"
          class="input"
          type="password"
          minlength="2"
          maxlength="32"
          autocomplete="new-password"
          required
      /></label>
      <label class="flex flex-col gap-2"
        ><span class="label">{{ t("password.confirm") }}</span
        ><input
          v-model="form.confirmation"
          class="input"
          type="password"
          minlength="2"
          maxlength="32"
          autocomplete="new-password"
          required
      /></label>
      <button class="button" :disabled="loading">
        {{ loading ? t("common.loading") : t("password.submit") }}
      </button>
    </form>
  </section>
</template>

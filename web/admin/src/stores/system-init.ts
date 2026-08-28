import { defineStore } from "pinia";
import { computed, ref } from "vue";

import { normalizeApiError } from "@/api/http";
import { getSystemInit, type SystemInitConfig } from "@/api/services/system";

export const useSystemInitStore = defineStore("system-init", () => {
  const config = ref<SystemInitConfig | null>(null);
  const isAvailable = ref(false);
  const isLoading = ref(false);
  const error = ref<string | null>(null);

  const isInitialized = computed(() => config.value?.first_init === true);
  const captchaSiteKey = computed(
    () => config.value?.google_captcha_site_key?.trim() ?? "",
  );

  function applyConfig(nextConfig: SystemInitConfig): void {
    config.value = { ...config.value, ...nextConfig };
    isAvailable.value = true;
    error.value = null;
  }

  async function initialize(): Promise<boolean> {
    isLoading.value = true;
    error.value = null;
    try {
      config.value = await getSystemInit();
      isAvailable.value = true;
      return true;
    } catch (cause) {
      config.value = null;
      isAvailable.value = false;
      error.value = normalizeApiError(cause).message;
      return false;
    } finally {
      isLoading.value = false;
    }
  }

  return {
    applyConfig,
    captchaSiteKey,
    config,
    error,
    initialize,
    isAvailable,
    isInitialized,
    isLoading,
  };
});

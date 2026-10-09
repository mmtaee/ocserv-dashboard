<script setup lang="ts">
import { ref } from "vue";
import { useRouter } from "vue-router";
import { useI18n } from "vue-i18n";

import { normalizeApiError } from "@/api/http";
import {
  type SystemUpdateData,
  updateSystemConfig,
} from "@/api/services/system";
import logoUrl from "@/assets/logo.svg";
import SignupForm from "@/components/blocks/signup-02/components/SignupForm.vue";
import LanguageSwitcher from "@/components/LanguageSwitcher.vue";
import { useSystemInitStore } from "@/stores/system-init";

const router = useRouter();
const systemInit = useSystemInitStore();
const pending = ref(false);
const error = ref<string | null>(null);
const { t } = useI18n({ useScope: "global" });

async function submit(settings: SystemUpdateData): Promise<void> {
  pending.value = true;
  error.value = null;
  try {
    const config = await updateSystemConfig(settings);
    systemInit.applyConfig(config);

    await router.replace({
      name: systemInit.isInitialized ? "home" : "system-setup",
    });
  } catch (cause) {
    error.value = normalizeApiError(cause).message;
  } finally {
    pending.value = false;
  }
}
</script>

<template>
  <div
    class="grid min-h-svh lg:grid-cols-[minmax(0,1.25fr)_minmax(22rem,.75fr)]"
  >
    <div class="flex flex-col gap-6 p-6 md:p-10">
      <div class="flex items-center gap-2 font-medium">
        <span
          class="grid size-7 place-items-center rounded-md bg-primary text-xs text-primary-foreground"
          aria-hidden="true"
          >OC</span
        >
        {{ t("common.appName") }}
        <LanguageSwitcher class="ms-auto" />
      </div>
      <div class="mx-auto flex w-full max-w-2xl flex-1 items-center py-6">
        <SignupForm
          class="w-full"
          :error="error"
          :pending="pending"
          @submit="submit"
        />
      </div>
    </div>
    <aside
      class="relative hidden overflow-hidden bg-zinc-950 text-white lg:flex lg:items-center lg:justify-center lg:p-12"
    >
      <div
        class="absolute inset-0 bg-[radial-gradient(circle_at_20%_20%,oklch(0.55_0.19_260/.55),transparent_40%),radial-gradient(circle_at_80%_70%,oklch(0.62_0.18_185/.3),transparent_35%)]"
      />
      <div
        class="relative flex max-w-md flex-col items-center gap-6 text-center"
      >
        <img :src="logoUrl" :alt="t('common.appName')" class="size-24" />
        <div class="flex flex-col gap-3">
          <p class="text-2xl font-semibold">{{ t("setup.promoTitle") }}</p>
          <p class="text-sm leading-6 text-white/65">
            {{ t("setup.promoDescription") }}
          </p>
        </div>
      </div>
    </aside>
  </div>
</template>

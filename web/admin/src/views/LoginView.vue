<script setup lang="ts">
import { useRouter } from "vue-router";
import { useI18n } from "vue-i18n";

import type { GithubComMmtaeeOcservDashboardBackendInternalServicesAdminApiSystemLoginData } from "@/api/generated";
import logoUrl from "@/assets/logo.svg";
import LoginForm from "@/components/blocks/login-02/components/LoginForm.vue";
import LanguageSwitcher from "@/components/LanguageSwitcher.vue";
import { useAuthStore } from "@/stores/auth";
import { useSystemInitStore } from "@/stores/system-init";

const router = useRouter();
const auth = useAuthStore();
const systemInit = useSystemInitStore();
const { t } = useI18n({ useScope: "global" });

async function submit(
  credentials: GithubComMmtaeeOcservDashboardBackendInternalServicesAdminApiSystemLoginData,
): Promise<void> {
  if (await auth.signIn(credentials)) {
    await router.replace({
      name: systemInit.isInitialized ? "home" : "system-setup",
    });
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
      <div class="flex flex-1 items-center justify-center">
        <div class="w-full max-w-sm">
          <LoginForm
            :captcha-site-key="systemInit.captchaSiteKey"
            :error="auth.error"
            :pending="auth.isLoading"
            @submit="submit"
          />
        </div>
      </div>
    </div>
    <aside
      class="relative hidden overflow-hidden bg-zinc-950 text-white lg:flex lg:items-center lg:justify-center lg:p-12"
    >
      <div
        class="absolute inset-0 bg-[radial-gradient(circle_at_30%_20%,oklch(0.55_0.19_260/.55),transparent_42%),radial-gradient(circle_at_75%_75%,oklch(0.62_0.18_185/.35),transparent_38%)]"
      />
      <div
        class="relative flex max-w-md flex-col items-center gap-6 text-center"
      >
        <img :src="logoUrl" :alt="t('common.appName')" class="size-24" />
        <div class="flex flex-col gap-3">
          <p class="text-2xl font-semibold">{{ t("auth.promoTitle") }}</p>
          <p class="text-sm leading-6 text-white/65">
            {{ t("auth.promoDescription") }}
          </p>
        </div>
      </div>
    </aside>
  </div>
</template>

<script setup lang="ts">
import { useI18n } from "vue-i18n";
import { useRouter } from "vue-router";
import LanguageSwitcher from "@/components/LanguageSwitcher.vue";
import logoUrl from "@/assets/logo.svg";
import { useAuthStore } from "@/stores/auth";

const { t } = useI18n({ useScope: "global" });
const router = useRouter();
const auth = useAuthStore();
const links = ["home", "sessions", "activity", "statistics"] as const;
async function logout(): Promise<void> {
  auth.signOut();
  await router.replace({ name: "login" });
}
</script>

<template>
  <div class="min-h-screen">
    <header class="border-b bg-card">
      <div
        class="mx-auto flex max-w-6xl flex-wrap items-center gap-3 px-4 py-3"
      >
        <img :src="logoUrl" alt="" class="size-8" />
        <strong>{{ t("app") }}</strong>
        <nav class="flex flex-1 flex-wrap gap-1">
          <RouterLink
            v-for="name in links"
            :key="name"
            class="rounded-md px-3 py-2 text-sm hover:bg-muted"
            active-class="bg-muted font-medium"
            :to="{ name }"
          >
            {{ t("nav." + name) }}
          </RouterLink>
        </nav>
        <LanguageSwitcher />
        <button class="button-secondary" type="button" @click="logout">
          {{ t("nav.logout") }}
        </button>
      </div>
    </header>
    <main class="mx-auto max-w-6xl p-4">
      <RouterView />
    </main>
  </div>
</template>

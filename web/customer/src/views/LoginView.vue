<script setup lang="ts">
import { reactive } from "vue";
import { useI18n } from "vue-i18n";
import { useRouter } from "vue-router";
import LanguageSwitcher from "@/components/LanguageSwitcher.vue";
import { useAuthStore } from "@/stores/auth";

const { t } = useI18n({ useScope: "global" });
const router = useRouter();
const auth = useAuthStore();
const form = reactive({ username: "", password: "" });
async function submit(): Promise<void> {
  if (await auth.signIn(form)) await router.replace({ name: "home" });
}
</script>

<template>
  <main class="grid min-h-screen place-items-center p-4">
    <form
      class="card flex w-full max-w-sm flex-col gap-5"
      @submit.prevent="submit"
    >
      <div class="flex items-start justify-between gap-3">
        <div>
          <h1 class="text-xl font-semibold">{{ t("login.title") }}</h1>
          <p class="text-sm text-muted-foreground">{{ t("login.subtitle") }}</p>
        </div>
        <LanguageSwitcher />
      </div>
      <p v-if="auth.error" class="alert">{{ auth.error }}</p>
      <label class="flex flex-col gap-2">
        <span class="label">{{ t("login.username") }}</span>
        <input
          v-model="form.username"
          class="input"
          minlength="2"
          maxlength="32"
          autocomplete="username"
          required
        />
      </label>
      <label class="flex flex-col gap-2">
        <span class="label">{{ t("login.password") }}</span>
        <input
          v-model="form.password"
          class="input"
          type="password"
          minlength="2"
          maxlength="32"
          autocomplete="current-password"
          required
        />
      </label>
      <button class="button" type="submit" :disabled="auth.loading">
        {{ auth.loading ? t("common.loading") : t("login.submit") }}
      </button>
    </form>
  </main>
</template>

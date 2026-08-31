import { createPinia } from "pinia";
import { createApp } from "vue";

import App from "./App.vue";
import { setUnauthorizedHandler } from "@/api/http";
import { useTheme } from "@/composables/use-theme";
import { i18n } from "@/locales";
import { installRouterGuards, router } from "@/router";
import { useAuthStore } from "@/stores/auth";
import { useSystemInitStore } from "@/stores/system-init";

import "./style.css";

useTheme();

async function bootstrap(): Promise<void> {
  const app = createApp(App);
  const pinia = createPinia();

  app.use(i18n);
  app.use(pinia);

  const systemInit = useSystemInitStore(pinia);
  const auth = useAuthStore(pinia);

  await systemInit.initialize();

  if (systemInit.isAvailable) {
    await auth.restoreSession();
  }

  installRouterGuards(pinia);
  app.use(router);

  setUnauthorizedHandler(async () => {
    auth.clearSession();

    if (router.currentRoute.value.name !== "login") {
      await router.replace({ name: "login" });
    }
  });

  await router.isReady();
  app.mount("#app");
}

void bootstrap();

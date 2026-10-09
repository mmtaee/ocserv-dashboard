import { createPinia } from "pinia";
import { createApp } from "vue";

import App from "./App.vue";
import "./style.css";
import { setUnauthorizedHandler } from "@/api/http";
import { i18n, setLocale } from "@/locales";
import { installRouterGuards, router } from "@/router";
import { useAuthStore } from "@/stores/auth";

const app = createApp(App);
const pinia = createPinia();
app.use(i18n);
app.use(pinia);
setLocale(String(i18n.global.locale.value));
installRouterGuards(pinia);
app.use(router);

const auth = useAuthStore(pinia);
setUnauthorizedHandler(async () => {
  auth.clearSession();
  if (router.currentRoute.value.name !== "login")
    await router.replace({ name: "login" });
});

void router.isReady().then(() => app.mount("#app"));

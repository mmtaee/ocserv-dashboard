import type { Pinia } from "pinia";
import { createRouter, createWebHistory } from "vue-router";

import { getAccessToken } from "@/api/auth-token";
import { useAuthStore } from "@/stores/auth";
import { useSystemInitStore } from "@/stores/system-init";

export const router = createRouter({
  history: createWebHistory(import.meta.env.BASE_URL),
  routes: [
    {
      path: "/",
      name: "home",
      component: () => import("@/views/DashboardView.vue"),
    },
    {
      path: "/login",
      name: "login",
      component: () => import("@/views/LoginView.vue"),
    },
    {
      path: "/setup",
      name: "system-setup",
      component: () => import("@/views/SystemSetupView.vue"),
    },
    {
      path: "/server-unavailable",
      name: "server-unavailable",
      component: () => import("@/views/ServerUnavailableView.vue"),
    },
    { path: "/:pathMatch(.*)*", redirect: "/" },
  ],
});

export function installRouterGuards(pinia: Pinia): void {
  router.beforeEach(async (to) => {
    const systemInit = useSystemInitStore(pinia);
    const auth = useAuthStore(pinia);

    if (!systemInit.isAvailable) {
      return to.name === "server-unavailable"
        ? true
        : { name: "server-unavailable" };
    }

    if (to.name === "login" && getAccessToken() && !auth.isAuthenticated) {
      await auth.restoreSession();
    }

    if (!auth.isAuthenticated && to.name !== "login") return { name: "login" };

    if (!auth.isAuthenticated) return true;

    if (!systemInit.isInitialized) {
      return to.name === "system-setup" ? true : { name: "system-setup" };
    }

    if (
      to.name === "login" ||
      to.name === "system-setup" ||
      to.name === "server-unavailable"
    ) {
      return { name: "home" };
    }

    return true;
  });
}

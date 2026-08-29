import type { Pinia } from "pinia";
import { createRouter, createWebHistory } from "vue-router";

import { getAccessToken } from "@/api/auth-token";
import { dashboardRoutes } from "@/router/dashboard-routes";
import { useAuthStore } from "@/stores/auth";
import { useSystemInitStore } from "@/stores/system-init";

const dashboardView = () => import("@/views/DashboardView.vue");

export const router = createRouter({
  history: createWebHistory(import.meta.env.BASE_URL),
  routes: [
    ...dashboardRoutes.map((route) => ({
      path: route.path,
      name: route.name,
      component: dashboardView,
      meta: {
        titleKey: route.titleKey,
        superadminOnly: !route.adminVisible,
      },
    })),
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

    if (to.meta.superadminOnly && !auth.user?.superadmin) {
      return { name: "home" };
    }

    return true;
  });
}

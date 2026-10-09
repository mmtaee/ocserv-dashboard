import type { Pinia } from "pinia";
import { createRouter, createWebHistory } from "vue-router";

import { getAccessToken } from "@/api/auth-token";
import { dashboardRoutes } from "@/router/dashboard-routes";
import { useAuthStore } from "@/stores/auth";
import { useSystemInitStore } from "@/stores/system-init";

const dashboardView = () => import("@/views/DashboardView.vue");
const emptyRouteView = () => import("@/views/EmptyRouteView.vue");
const occtlView = () => import("@/views/OcctlView.vue");
const ocservGroupDefaultsView = () =>
  import("@/views/OcservGroupDefaultsView.vue");
const ocservGroupsView = () => import("@/views/OcservGroupsView.vue");
const ocservUsersView = () => import("@/views/OcservUsersView.vue");
const ocservSyncView = () => import("@/views/OcservSyncView.vue");
const runtimeView = () => import("@/views/RuntimeView.vue");
const statisticsView = () => import("@/views/StatisticsView.vue");
const bandwidthsView = () => import("@/views/BandwidthsView.vue");
const sessionLogsView = () => import("@/views/SessionLogsView.vue");
const staffsView = () => import("@/views/StaffsView.vue");
const telegramRequestsView = () => import("@/views/TelegramRequestsView.vue");
const telegramPackagesView = () => import("@/views/TelegramPackagesView.vue");
const telegramSettingsView = () => import("@/views/TelegramSettingsView.vue");
const backupView = () => import("@/views/BackupView.vue");
const systemSettingsView = () => import("@/views/SystemSettingsView.vue");

const dashboardComponents: Partial<
  Record<(typeof dashboardRoutes)[number]["name"], () => Promise<unknown>>
> = {
  home: dashboardView,
  occtl: occtlView,
  "ocserv-group-defaults": ocservGroupDefaultsView,
  "ocserv-groups": ocservGroupsView,
  "ocserv-users": ocservUsersView,
  "ocserv-sync": ocservSyncView,
  "ocserv-runtime": runtimeView,
  statistics: statisticsView,
  bandwidths: bandwidthsView,
  "session-logs": sessionLogsView,
  staffs: staffsView,
  "telegram-requests": telegramRequestsView,
  "telegram-packages": telegramPackagesView,
  "telegram-settings": telegramSettingsView,
  backup: backupView,
  "system-settings": systemSettingsView,
};

export const router = createRouter({
  history: createWebHistory(import.meta.env.BASE_URL),
  routes: [
    ...dashboardRoutes.map((route) => ({
      path: route.path,
      ...(route.alias ? { alias: route.alias } : {}),
      name: route.name,
      component: dashboardComponents[route.name] ?? emptyRouteView,
      meta: {
        titleKey: route.titleKey,
        superadminOnly: !route.adminVisible,
        telegramOnly: route.telegramOnly,
      },
    })),
    {
      path: "/login",
      name: "login",
      component: () => import("@/views/LoginView.vue"),
    },
    {
      path: "/reset-password",
      name: "reset-password",
      component: () => import("@/views/ResetPasswordView.vue"),
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

    if (to.meta.telegramOnly && !systemInit.telegramBotEnabled) {
      return { name: "home" };
    }

    if (!systemInit.isAvailable) {
      return to.name === "server-unavailable"
        ? true
        : { name: "server-unavailable" };
    }

    if (to.name === "login" && getAccessToken() && !auth.isAuthenticated) {
      await auth.restoreSession();
    }

    if (
      !auth.isAuthenticated &&
      !["login", "reset-password"].includes(String(to.name))
    ) {
      return { name: "login" };
    }

    if (!auth.isAuthenticated) return true;

    if (!systemInit.isInitialized) {
      return to.name === "system-setup" ? true : { name: "system-setup" };
    }

    if (
      to.name === "login" ||
      to.name === "reset-password" ||
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

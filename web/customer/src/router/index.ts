import type { Pinia } from "pinia";
import { createRouter, createWebHistory } from "vue-router";
import { getAccessToken } from "@/api/auth-token";
import CustomerLayout from "@/layouts/CustomerLayout.vue";
import { useAuthStore } from "@/stores/auth";

export const router = createRouter({
  history: createWebHistory(import.meta.env.BASE_URL),
  routes: [
    {
      path: "/login",
      name: "login",
      component: () => import("@/views/LoginView.vue"),
      meta: { public: true },
    },
    {
      path: "/",
      component: CustomerLayout,
      children: [
        {
          path: "",
          name: "home",
          component: () => import("@/views/SummaryView.vue"),
        },
        {
          path: "sessions",
          name: "sessions",
          component: () => import("@/views/SessionsView.vue"),
        },
        {
          path: "activity",
          name: "activity",
          component: () => import("@/views/ActivityView.vue"),
        },
        {
          path: "statistics",
          name: "statistics",
          component: () => import("@/views/StatisticsView.vue"),
        },
      ],
    },
    { path: "/:pathMatch(.*)*", redirect: "/" },
  ],
});

export function installRouterGuards(pinia: Pinia): void {
  router.beforeEach(async (to) => {
    const auth = useAuthStore(pinia);
    if (getAccessToken() && !auth.isAuthenticated) await auth.restoreSession();
    if (!to.meta.public && !auth.isAuthenticated) return { name: "login" };
    if (to.name === "login" && auth.isAuthenticated) return { name: "home" };
    return true;
  });
}

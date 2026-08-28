<script setup lang="ts">
import type { SidebarProps } from "@/components/ui/sidebar";
import { computed } from "vue";
import { Activity, Bug, LayoutDashboard, Server, Users } from "@lucide/vue";
import { useI18n } from "vue-i18n";

import logoUrl from "@/assets/logo.svg";
import NavMain from "@/components/NavMain.vue";
import NavSecondary from "@/components/NavSecondary.vue";
import NavUser from "@/components/NavUser.vue";
import {
  Sidebar,
  SidebarContent,
  SidebarFooter,
  SidebarHeader,
  SidebarMenu,
  SidebarMenuButton,
  SidebarMenuItem,
} from "@/components/ui/sidebar";
import { useAuthStore } from "@/stores/auth";

const props = defineProps<SidebarProps>();
const { t } = useI18n({ useScope: "global" });
const auth = useAuthStore();

const data = computed(() => ({
  user: {
    name: auth.user?.username ?? t("auth.username"),
    detail: auth.user?.superadmin
      ? t("dashboard.administration")
      : t("common.appName"),
  },
  navMain: [
    {
      title: t("dashboard.title"),
      url: "/",
      icon: LayoutDashboard,
      isActive: true,
    },
    {
      title: t("dashboard.connectedUsers"),
      url: "/",
      icon: Users,
      disabled: true,
    },
    {
      title: t("dashboard.activeSessions"),
      url: "/",
      icon: Activity,
      disabled: true,
    },
    {
      title: t("dashboard.managedServers"),
      url: "/",
      icon: Server,
      disabled: true,
    },
  ],
  navSecondary: [
    {
      title: t("footer.reportIssue"),
      url: "https://github.com/mmtaee/ocserv-dashboard/issues",
      icon: Bug,
      external: true,
    },
  ],
}));
</script>

<template>
  <Sidebar
    class="top-(--header-height) h-[calc(100svh-var(--header-height))]!"
    v-bind="props"
  >
    <SidebarHeader>
      <SidebarMenu>
        <SidebarMenuItem>
          <SidebarMenuButton size="lg" as-child>
            <RouterLink to="/">
              <div
                class="flex size-8 items-center justify-center rounded-lg bg-sidebar-primary p-1.5 text-sidebar-primary-foreground"
              >
                <img :src="logoUrl" alt="" class="size-full" />
              </div>
              <div class="grid flex-1 text-start text-sm leading-tight">
                <span class="truncate font-medium">{{
                  t("common.appName")
                }}</span>
                <span class="truncate text-xs">{{
                  t("dashboard.administration")
                }}</span>
              </div>
            </RouterLink>
          </SidebarMenuButton>
        </SidebarMenuItem>
      </SidebarMenu>
    </SidebarHeader>
    <SidebarContent>
      <NavMain :items="data.navMain" />
      <NavSecondary :items="data.navSecondary" class="mt-auto" />
    </SidebarContent>
    <SidebarFooter>
      <NavUser :user="data.user" />
    </SidebarFooter>
  </Sidebar>
</template>

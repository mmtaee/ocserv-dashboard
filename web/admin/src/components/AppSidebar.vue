<script setup lang="ts">
import type { SidebarProps } from "@/components/ui/sidebar";
import { ShieldCheck } from "@lucide/vue";
import { computed } from "vue";
import { useI18n } from "vue-i18n";

import NavMain from "@/components/NavMain.vue";
import {
  Sidebar,
  SidebarContent,
  SidebarFooter,
  SidebarHeader,
  SidebarMenu,
  SidebarMenuButton,
  SidebarMenuItem,
  SidebarRail,
} from "@/components/ui/sidebar";
import { dashboardRoutes } from "@/router/dashboard-routes";
import { useAuthStore } from "@/stores/auth";

const props = defineProps<SidebarProps>();
const { t } = useI18n({ useScope: "global" });
const auth = useAuthStore();

const groups = computed(() => {
  const visibleRoutes = dashboardRoutes.filter(
    (route) => auth.user?.superadmin || route.adminVisible,
  );
  const sectionKeys = [
    ...new Set(visibleRoutes.map((route) => route.sectionKey)),
  ];

  return sectionKeys.map((sectionKey) => ({
    title: t(sectionKey),
    items: visibleRoutes
      .filter((route) => route.sectionKey === sectionKey)
      .map((route) => ({
        title: t(route.titleKey),
        path: route.path,
        icon: route.icon,
      })),
  }));
});

const roleLabel = computed(() =>
  auth.user?.superadmin ? t("navigation.superadmin") : t("navigation.admin"),
);
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
                class="flex aspect-square size-8 items-center justify-center rounded-lg bg-sidebar-primary text-sidebar-primary-foreground"
              >
                <ShieldCheck />
              </div>
              <div class="grid flex-1 text-start text-sm leading-tight">
                <span class="truncate font-medium">
                  {{                  t("common.appName")                }}
                </span>
                <span class="truncate text-xs">{{ roleLabel }}</span>
              </div>
            </RouterLink>
          </SidebarMenuButton>
        </SidebarMenuItem>
      </SidebarMenu>
    </SidebarHeader>
    <SidebarContent>
      <NavMain :groups="groups" />
    </SidebarContent>
    <SidebarFooter>
      <SidebarMenu>
        <SidebarMenuItem>
          <SidebarMenuButton :tooltip="auth.user?.username">
            <ShieldCheck />
            <span class="truncate">{{ auth.user?.username }}</span>
          </SidebarMenuButton>
        </SidebarMenuItem>
      </SidebarMenu>
    </SidebarFooter>
    <SidebarRail />
  </Sidebar>
</template>

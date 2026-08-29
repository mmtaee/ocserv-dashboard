<script setup lang="ts">
import type { LucideIcon } from "@lucide/vue";
import { useRoute } from "vue-router";

import {
  SidebarGroup,
  SidebarGroupContent,
  SidebarGroupLabel,
  SidebarMenu,
  SidebarMenuButton,
  SidebarMenuItem,
} from "@/components/ui/sidebar";

defineProps<{
  groups: {
    title: string;
    items: { title: string; path: string; icon: LucideIcon }[];
  }[];
}>();

const route = useRoute();
</script>

<template>
  <SidebarGroup v-for="group in groups" :key="group.title">
    <SidebarGroupLabel>{{ group.title }}</SidebarGroupLabel>
    <SidebarGroupContent>
      <SidebarMenu>
        <SidebarMenuItem v-for="item in group.items" :key="item.path">
          <SidebarMenuButton
            as-child
            :is-active="route.path === item.path"
            :tooltip="item.title"
          >
            <RouterLink :to="item.path">
              <component :is="item.icon" />
              <span>{{ item.title }}</span>
            </RouterLink>
          </SidebarMenuButton>
        </SidebarMenuItem>
      </SidebarMenu>
    </SidebarGroupContent>
  </SidebarGroup>
</template>

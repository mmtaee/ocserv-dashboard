<script setup lang="ts">
import { computed } from "vue";

import { Avatar, AvatarFallback } from "@/components/ui/avatar";
import {
  SidebarMenu,
  SidebarMenuButton,
  SidebarMenuItem,
} from "@/components/ui/sidebar";

const props = defineProps<{
  user: {
    name: string;
    detail: string;
  };
}>();

const initials = computed(() =>
  props.user.name.slice(0, 2).toLocaleUpperCase(),
);
</script>

<template>
  <SidebarMenu>
    <SidebarMenuItem>
      <SidebarMenuButton size="lg" :tooltip="user.name">
        <Avatar class="size-8 rounded-lg">
          <AvatarFallback class="rounded-lg">{{ initials }}</AvatarFallback>
        </Avatar>
        <div class="grid flex-1 text-start text-sm leading-tight">
          <span class="truncate font-medium">{{ user.name }}</span>
          <span class="truncate text-xs">{{ user.detail }}</span>
        </div>
      </SidebarMenuButton>
    </SidebarMenuItem>
  </SidebarMenu>
</template>

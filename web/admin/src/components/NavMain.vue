<script setup lang="ts">
import type { LucideIcon } from '@lucide/vue'
import { useI18n } from 'vue-i18n'

import {
  SidebarGroup,
  SidebarGroupLabel,
  SidebarMenu,
  SidebarMenuButton,
  SidebarMenuItem,
} from '@/components/ui/sidebar'

const { t } = useI18n({ useScope: 'global' })

defineProps<{
  items: {
    title: string
    url: string
    icon: LucideIcon
    isActive?: boolean
    disabled?: boolean
  }[]
}>()
</script>

<template>
  <SidebarGroup>
    <SidebarGroupLabel>{{ t('dashboard.administration') }}</SidebarGroupLabel>
    <SidebarMenu>
      <SidebarMenuItem v-for="item in items" :key="item.title">
        <SidebarMenuButton
          v-if="item.disabled"
          :tooltip="item.title"
          disabled
        >
          <component :is="item.icon" />
          <span>{{ item.title }}</span>
        </SidebarMenuButton>
        <SidebarMenuButton v-else as-child :is-active="item.isActive" :tooltip="item.title">
          <RouterLink :to="item.url">
            <component :is="item.icon" />
            <span>{{ item.title }}</span>
          </RouterLink>
        </SidebarMenuButton>
      </SidebarMenuItem>
    </SidebarMenu>
  </SidebarGroup>
</template>

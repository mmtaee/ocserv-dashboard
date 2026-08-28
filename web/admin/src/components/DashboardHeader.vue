<script setup lang="ts">
import { Separator } from '@/components/ui/separator'
import { SidebarTrigger } from '@/components/ui/sidebar'
</script>

<template>
  <header class="flex h-(--header-height) shrink-0 items-center gap-2 border-b px-4">
    <SidebarTrigger class="-ml-1" />
    <Separator orientation="vertical" class="mr-2 h-4" />
    <div>
      <p class="text-xs text-muted-foreground">Administration</p>
      <p class="text-sm font-medium">OCServ Dashboard</p>
    </div>
  </header>
</template>

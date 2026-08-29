<script setup lang="ts">
import type { HTMLAttributes } from "vue"
import { cn } from "@/lib/utils"
import { useSidebar } from "./utils"

const props = defineProps<{
  class?: HTMLAttributes["class"]
}>()

const { toggleSidebar } = useSidebar()
</script>

<template>
  <button
    data-sidebar="rail"
    data-slot="sidebar-rail"
    aria-label="Toggle Sidebar"
    :tabindex="-1"
    title="Toggle Sidebar"
    :class="cn(
      'hover:after:bg-sidebar-border absolute inset-y-0 z-20 hidden w-4 -translate-x-1/2 rtl:translate-x-1/2 transition-all ease-linear group-data-[side=left]:-right-4 group-data-[side=right]:left-0 after:absolute after:inset-y-0 after:start-1/2 after:w-0.5 sm:flex',
      'in-data-[side=left]:cursor-w-resize rtl:in-data-[side=left]:cursor-e-resize in-data-[side=right]:cursor-e-resize rtl:in-data-[side=right]:cursor-w-resize',
      '[[data-side=left][data-state=collapsed]_&]:cursor-e-resize rtl:[[data-side=left][data-state=collapsed]_&]:cursor-w-resize [[data-side=right][data-state=collapsed]_&]:cursor-w-resize rtl:[[data-side=right][data-state=collapsed]_&]:cursor-e-resize',
      'hover:group-data-[collapsible=offcanvas]:bg-sidebar group-data-[collapsible=offcanvas]:translate-x-0 rtl:group-data-[collapsible=offcanvas]:-translate-x-0 group-data-[collapsible=offcanvas]:after:start-full',
      '[[data-side=left][data-collapsible=offcanvas]_&]:-end-2',
      '[[data-side=right][data-collapsible=offcanvas]_&]:-start-2',
      props.class,
    )"
    @click="toggleSidebar"
  >
    <slot />
  </button>
</template>

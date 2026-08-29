<script setup lang="ts">
import { Search } from "@lucide/vue"

import { Label } from '@/components/ui/label'
import { SidebarInput } from '@/components/ui/sidebar'
</script>

<template>
  <form>
    <div class="relative">
      <Label for="search" class="sr-only">
        Search
      </Label>
      <SidebarInput
        id="search"
        placeholder="Type to search..."
        class="h-8 ps-7"
      />
      <Search class="pointer-events-none absolute top-1/2 start-2 size-4 -translate-y-1/2 opacity-50 select-none" />
    </div>
  </form>
</template>

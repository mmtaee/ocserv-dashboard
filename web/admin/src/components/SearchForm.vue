<script setup lang="ts">
import { Search } from "@lucide/vue"
import { useI18n } from 'vue-i18n'

import { Label } from '@/components/ui/label'
import { SidebarInput } from '@/components/ui/sidebar'

const { t } = useI18n({ useScope: 'global' })
</script>

<template>
  <form role="search" @submit.prevent>
    <div class="relative">
      <Label for="search" class="sr-only">
        {{ t('navigation.search') }}
      </Label>
      <SidebarInput
        id="search"
        :placeholder="t('navigation.searchPlaceholder')"
        class="h-8 ps-7"
      />
      <Search class="pointer-events-none absolute top-1/2 start-2 size-4 -translate-y-1/2 opacity-50 select-none" />
    </div>
  </form>
</template>

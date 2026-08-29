<script setup lang="ts">
import {computed} from "vue";
import {useI18n} from "vue-i18n";

import AppSidebar from "@/components/AppSidebar.vue";
import SiteHeader from "@/components/SiteHeader.vue";
import {SidebarInset, SidebarProvider} from "@/components/ui/sidebar";
import {isRtlLocale} from "@/locales";

const { locale } = useI18n({ useScope: "global" });
const sidebarSide = computed(() =>
  isRtlLocale(locale.value) ? "right" : "left",
);
</script>

<template>
  <div class="flex flex-1 [--header-height:calc(--spacing(14))]">
    <SidebarProvider class="min-h-0! flex-1 flex-col">
      <SiteHeader />
      <div class="flex min-h-0 flex-1">
        <AppSidebar :side="sidebarSide" collapsible="icon" />
        <SidebarInset>
          <main class="flex flex-1 flex-col p-4 pb-14 lg:p-6 lg:pb-14">
            <slot />
          </main>
        </SidebarInset>
      </div>
    </SidebarProvider>
  </div>
</template>

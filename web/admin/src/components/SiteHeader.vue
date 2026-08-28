<script setup lang="ts">
import { SidebarIcon } from "@lucide/vue";
import { useI18n } from "vue-i18n";

import LanguageSwitcher from "@/components/LanguageSwitcher.vue";
import SearchForm from "@/components/SearchForm.vue";
import {
  Breadcrumb,
  BreadcrumbItem,
  BreadcrumbLink,
  BreadcrumbList,
  BreadcrumbPage,
  BreadcrumbSeparator,
} from "@/components/ui/breadcrumb";
import { Button } from "@/components/ui/button";
import { Separator } from "@/components/ui/separator";
import { useSidebar } from "@/components/ui/sidebar";

const { toggleSidebar } = useSidebar();
const { t } = useI18n({ useScope: "global" });
</script>

<template>
  <header
    class="bg-background sticky top-0 z-50 flex w-full items-center border-b"
  >
    <div class="flex h-(--header-height) w-full items-center gap-2 px-4">
      <Button
        class="size-8"
        variant="ghost"
        size="icon"
        :aria-label="t('common.toggleSidebar')"
        @click="toggleSidebar"
      >
        <SidebarIcon />
      </Button>
      <Separator orientation="vertical" class="me-2 h-4" />
      <Breadcrumb class="hidden sm:block">
        <BreadcrumbList>
          <BreadcrumbItem>
            <BreadcrumbLink href="/">
              {{ t("common.appName") }}
            </BreadcrumbLink>
          </BreadcrumbItem>
          <BreadcrumbSeparator />
          <BreadcrumbItem>
            <BreadcrumbPage>{{ t("dashboard.title") }}</BreadcrumbPage>
          </BreadcrumbItem>
        </BreadcrumbList>
      </Breadcrumb>
      <div class="ms-auto flex items-center gap-2">
        <SearchForm class="hidden w-56 md:block" />
        <LanguageSwitcher />
      </div>
    </div>
  </header>
</template>

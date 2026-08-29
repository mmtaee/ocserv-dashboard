<script setup lang="ts">
import { Moon, SidebarIcon, Sun } from "@lucide/vue";
import { computed } from "vue";
import { useI18n } from "vue-i18n";
import { useRoute } from "vue-router";

import LanguageSwitcher from "@/components/LanguageSwitcher.vue";
import { Button } from "@/components/ui/button";
import { Separator } from "@/components/ui/separator";
import { useSidebar } from "@/components/ui/sidebar";
import { useTheme } from "@/composables/use-theme";

const { toggleSidebar } = useSidebar();
const { isDark, toggleTheme } = useTheme();
const route = useRoute();
const { t } = useI18n({ useScope: "global" });
const routeName = computed(() =>
  t(String(route.meta.titleKey ?? "navigation.dashboard")),
);
const themeLabel = computed(() =>
  isDark.value ? t("common.switchToLightTheme") : t("common.switchToDarkTheme"),
);
</script>

<template>
  <header
    class="sticky top-0 z-50 flex w-full items-center border-b bg-background"
  >
    <div class="flex h-(--header-height) w-full items-center gap-2 px-4">
      <Button
        class="size-8"
        variant="ghost"
        size="icon"
        :aria-label="t('common.toggleSidebar')"
        :title="t('common.toggleSidebar')"
        @click="toggleSidebar"
      >
        <SidebarIcon data-icon="inline-start" />
      </Button>
      <Separator orientation="vertical" class="me-2 h-4" />
      <span class="truncate text-sm font-medium">{{ routeName }}</span>
      <div class="ms-auto flex items-center gap-2">
        <LanguageSwitcher />
        <Button
          class="size-8"
          variant="ghost"
          size="icon"
          :aria-label="themeLabel"
          :title="themeLabel"
          @click="toggleTheme"
        >
          <Sun v-if="isDark" data-icon="inline-start" />
          <Moon v-else data-icon="inline-start" />
        </Button>
      </div>
    </div>
  </header>
</template>

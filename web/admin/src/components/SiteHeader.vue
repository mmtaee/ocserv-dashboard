<script setup lang="ts">
import { AlertTriangle, Moon, Server, SidebarIcon, Sun } from "@lucide/vue";
import { computed, onMounted } from "vue";
import { storeToRefs } from "pinia";
import { useI18n } from "vue-i18n";
import { useRouter } from "vue-router";
import LanguageSwitcher from "@/components/LanguageSwitcher.vue";
import NavUser from "@/components/NavUser.vue";
import { Button } from "@/components/ui/button";
import { Separator } from "@/components/ui/separator";
import {
  Select,
  SelectContent,
  SelectGroup,
  SelectItem,
  SelectLabel,
  SelectTrigger,
  SelectValue,
} from "@/components/ui/select";
import { useSidebar } from "@/components/ui/sidebar";
import { useTheme } from "@/composables/use-theme";
import logoUrl from "@/assets/logo.svg";
import { useAuthStore } from "@/stores/auth";
import { agentServerId, MASTER_SERVER, useServerStore } from "@/stores/server";

const { toggleSidebar } = useSidebar();
const { isDark, toggleTheme } = useTheme();
const auth = useAuthStore();
const serverStore = useServerStore();
const { error, hasAgents, isLoading, selectableAgents, selectedServer } =
  storeToRefs(serverStore);
const router = useRouter();
const { t } = useI18n({ useScope: "global" });

const themeLabel = computed(() =>
  isDark.value ? t("common.switchToLightTheme") : t("common.switchToDarkTheme"),
);
const navUser = computed(() => ({
  name: auth.user?.username ?? "",
  role: auth.user?.superadmin
    ? t("navigation.superadmin")
    : t("navigation.admin"),
}));
const serverError = computed(() =>
  error.value ? t(`common.serverErrors.${error.value}`) : "",
);
const serverModel = computed({
  get: () => selectedServer.value,
  set: serverStore.selectServer,
});

function agentOptionLabel(
  agent: (typeof selectableAgents.value)[number],
): string {
  return `${agent.name} · ${agent.address}${agent.port ? `:${agent.port}` : ""}`;
}

async function handleLogout(): Promise<void> {
  await auth.signOut();
  await router.replace({ name: "login" });
}

onMounted(() => {
  if (auth.user?.superadmin) void serverStore.refreshAgents();
});
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
      <img :src="logoUrl" alt="" class="size-8 shrink-0" />

      <span class="hidden truncate text-base font-medium sm:block">
        {{ t("common.appName") }}
      </span>
      <div class="ms-auto flex items-center gap-2">
        <Select v-if="hasAgents" v-model="serverModel" :disabled="isLoading">
          <SelectTrigger
            class="w-36 sm:w-44"
            :aria-label="t('common.serverSelector')"
            :title="t('common.serverSelector')"
          >
            <Server />
            <SelectValue :placeholder="t('common.serverSelector')" />
          </SelectTrigger>
          <SelectContent>
            <SelectGroup>
              <SelectLabel>{{ t("common.serverSelector") }}</SelectLabel>
              <SelectItem :value="MASTER_SERVER">
                {{ t("common.masterServer") }}
              </SelectItem>
              <SelectItem
                v-for="agent in selectableAgents"
                :key="agentServerId(agent)"
                :value="agentServerId(agent)"
              >
                {{ agentOptionLabel(agent) }}
              </SelectItem>
            </SelectGroup>
          </SelectContent>
        </Select>
        <AlertTriangle
          v-if="serverError"
          class="text-destructive"
          :aria-label="serverError"
          :title="serverError"
        />
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
        <NavUser
          :user="navUser"
          :is-logging-out="auth.isLoading"
          @logout="handleLogout"
        />
      </div>
    </div>
  </header>
</template>

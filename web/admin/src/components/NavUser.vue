<script setup lang="ts">
import { LogOut } from "@lucide/vue";
import { computed, onBeforeUnmount, ref } from "vue";
import { useI18n } from "vue-i18n";

import { Avatar, AvatarFallback } from "@/components/ui/avatar";
import {
  DropdownMenu,
  DropdownMenuContent,
  DropdownMenuGroup,
  DropdownMenuItem,
  DropdownMenuLabel,
  DropdownMenuSeparator,
  DropdownMenuTrigger,
} from "@/components/ui/dropdown-menu";

const props = withDefaults(
  defineProps<{
    user: {
      name: string;
      role: string;
    };
    isLoggingOut?: boolean;
  }>(),
  { isLoggingOut: false },
);

const emit = defineEmits<{
  logout: [];
}>();
const { t } = useI18n({ useScope: "global" });
const isOpen = ref(false);
let closeTimer: ReturnType<typeof setTimeout> | undefined;

function cancelClose(): void {
  if (closeTimer !== undefined) {
    clearTimeout(closeTimer);
    closeTimer = undefined;
  }
}

function openOnHover(): void {
  cancelClose();
  isOpen.value = true;
}

function closeOnHover(): void {
  cancelClose();
  closeTimer = setTimeout(() => {
    isOpen.value = false;
    closeTimer = undefined;
  }, 150);
}

onBeforeUnmount(cancelClose);

const initials = computed(() =>
  props.user.name
    .split(/\s+/)
    .filter(Boolean)
    .slice(0, 2)
    .map((part) => part[0]?.toUpperCase())
    .join(""),
);
</script>

<template>
  <DropdownMenu v-model:open="isOpen" :modal="false">
    <DropdownMenuTrigger as-child>
      <Avatar
        class="size-9 cursor-pointer rounded-md"
        @mouseenter="openOnHover"
        @mouseleave="closeOnHover"
      >
        <AvatarFallback class="rounded-md">
          {{ initials }}
        </AvatarFallback>
      </Avatar>
    </DropdownMenuTrigger>
    <DropdownMenuContent
      class="min-w-56"
      align="end"
      side="bottom"
      @mouseenter="openOnHover"
      @mouseleave="closeOnHover"
    >
      <DropdownMenuLabel class="font-normal">
        <div class="flex flex-col gap-1 text-start">
          <span class="truncate text-sm font-medium capitalize">
            {{ user.name }}
          </span>
          <span class="truncate text-xs text-muted-foreground">
            {{ user.role }}
          </span>
        </div>
      </DropdownMenuLabel>
      <DropdownMenuSeparator />
      <DropdownMenuGroup>
        <DropdownMenuItem :disabled="isLoggingOut" @select="emit('logout')">
          <LogOut />
          {{ t("navigation.logout") }}
        </DropdownMenuItem>
      </DropdownMenuGroup>
    </DropdownMenuContent>
  </DropdownMenu>
</template>

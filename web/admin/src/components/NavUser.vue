<script setup lang="ts">
import { KeyRound, LogOut } from "@lucide/vue";
import { computed, onBeforeUnmount, reactive, shallowRef } from "vue";
import { useI18n } from "vue-i18n";

import { Avatar, AvatarFallback } from "@/components/ui/avatar";
import { changePassword } from "@/api/services/auth";
import { normalizeApiError } from "@/api/http";
import { Button } from "@/components/ui/button";
import {
  DropdownMenu,
  DropdownMenuContent,
  DropdownMenuGroup,
  DropdownMenuItem,
  DropdownMenuLabel,
  DropdownMenuSeparator,
  DropdownMenuTrigger,
} from "@/components/ui/dropdown-menu";
import {
  Field,
  FieldDescription,
  FieldError,
  FieldGroup,
  FieldLabel,
} from "@/components/ui/field";
import { Input } from "@/components/ui/input";
import {
  Sheet,
  SheetContent,
  SheetDescription,
  SheetFooter,
  SheetHeader,
  SheetTitle,
} from "@/components/ui/sheet";
import { Spinner } from "@/components/ui/spinner";

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
const isOpen = shallowRef(false);
const passwordOpen = shallowRef(false);
const pending = shallowRef(false);
const error = shallowRef("");
const success = shallowRef("");
const attempted = shallowRef(false);
const form = reactive({
  oldPassword: "",
  newPassword: "",
  confirmPassword: "",
});
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
const oldPasswordError = computed(() =>
  attempted.value &&
  (form.oldPassword.length < 4 || form.oldPassword.length > 32)
    ? t("navUser.passwordLength")
    : "",
);
const passwordError = computed(() =>
  attempted.value &&
  (form.newPassword.length < 4 || form.newPassword.length > 32)
    ? t("navUser.passwordLength")
    : "",
);
const confirmationError = computed(() =>
  attempted.value && form.newPassword !== form.confirmPassword
    ? t("navUser.passwordMismatch")
    : "",
);
const valid = computed(
  () =>
    Boolean(form.oldPassword && form.newPassword && form.confirmPassword) &&
    !oldPasswordError.value &&
    !passwordError.value &&
    !confirmationError.value,
);
function openPassword(): void {
  isOpen.value = false;
  attempted.value = false;
  error.value = "";
  success.value = "";
  passwordOpen.value = true;
}
function resetForm(): void {
  form.oldPassword = "";
  form.newPassword = "";
  form.confirmPassword = "";
  attempted.value = false;
  error.value = "";
  success.value = "";
}
async function submitPassword(): Promise<void> {
  attempted.value = true;
  if (!valid.value || pending.value) return;
  pending.value = true;
  error.value = "";
  try {
    await changePassword({
      old_password: form.oldPassword,
      new_password: form.newPassword,
    });
    success.value = t("navUser.passwordChanged");
    window.setTimeout(() => {
      passwordOpen.value = false;
      resetForm();
      pending.value = false;
    }, 800);
    return;
  } catch (cause) {
    error.value = normalizeApiError(cause).message;
  }
  pending.value = false;
}
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
        <DropdownMenuItem @select="openPassword">
          <KeyRound />
          {{ t("navUser.changePassword") }}
        </DropdownMenuItem>
        <DropdownMenuItem :disabled="isLoggingOut" @select="emit('logout')">
          <LogOut />
          {{ t("navigation.logout") }}
        </DropdownMenuItem>
      </DropdownMenuGroup>
    </DropdownMenuContent>
  </DropdownMenu>
  <Sheet v-model:open="passwordOpen">
    <SheetContent>
      <form class="flex flex-col gap-6" @submit.prevent="submitPassword">
        <SheetHeader
          ><SheetTitle>{{ t("navUser.changePassword") }}</SheetTitle
          ><SheetDescription>{{
            t("navUser.changePasswordDescription")
          }}</SheetDescription></SheetHeader
        >
        <FieldGroup class="p-4">
          <Field :data-invalid="Boolean(oldPasswordError)"
            ><FieldLabel for="password-old">{{
              t("navUser.currentPassword")
            }}</FieldLabel
            ><Input
              id="password-old"
              v-model="form.oldPassword"
              type="password"
              minlength="4"
              maxlength="32"
              required
              :aria-invalid="Boolean(oldPasswordError)"
              :disabled="pending"
            /><FieldError v-if="oldPasswordError">{{
              oldPasswordError
            }}</FieldError></Field
          >
          <Field :data-invalid="Boolean(passwordError)"
            ><FieldLabel for="password-new">{{
              t("navUser.newPassword")
            }}</FieldLabel
            ><Input
              id="password-new"
              v-model="form.newPassword"
              type="password"
              minlength="4"
              maxlength="32"
              required
              :aria-invalid="Boolean(passwordError)"
              :disabled="pending"
            /><FieldError v-if="passwordError">{{
              passwordError
            }}</FieldError></Field
          >
          <Field :data-invalid="Boolean(confirmationError)"
            ><FieldLabel for="password-confirm">{{
              t("navUser.confirmPassword")
            }}</FieldLabel
            ><Input
              id="password-confirm"
              v-model="form.confirmPassword"
              type="password"
              required
              :aria-invalid="Boolean(confirmationError)"
              :disabled="pending"
            /><FieldError v-if="confirmationError">{{
              confirmationError
            }}</FieldError></Field
          >
          <FieldError v-if="error">{{ error }}</FieldError>
          <FieldDescription v-if="success">{{ success }}</FieldDescription>
        </FieldGroup>
        <SheetFooter
          ><Button
            type="button"
            variant="outline"
            :disabled="pending"
            @click="
              passwordOpen = false;
              resetForm();
            "
            >{{ t("navUser.cancel") }}</Button
          ><Button type="submit" :disabled="pending || !valid"
            ><Spinner v-if="pending" data-icon="inline-start" /><KeyRound
              v-else
              data-icon="inline-start"
            />{{ t("navUser.changePassword") }}</Button
          ></SheetFooter
        >
      </form>
    </SheetContent>
  </Sheet>
</template>

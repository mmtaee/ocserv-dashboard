<script setup lang="ts">
import { computed, reactive, shallowRef, watch } from "vue";
import { useI18n } from "vue-i18n";

import type { Staff, StaffCreate, StaffPassword } from "@/api/services/staffs";
import { Button } from "@/components/ui/button";
import {
  Field,
  FieldDescription,
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

const props = defineProps<{
  mode: "create" | "password";
  pending?: boolean;
  staff?: Staff | null;
}>();
const emit = defineEmits<{
  submit: [request: StaffCreate | StaffPassword];
}>();
const open = defineModel<boolean>("open", { default: false });
const { t } = useI18n({ useScope: "global" });
const attempted = shallowRef(false);
const form = reactive({ password: "", username: "" });
const usernameError = computed(() =>
  attempted.value && props.mode === "create" && !form.username.trim()
    ? t("staffs.usernameRequired")
    : "",
);
const passwordError = computed(() =>
  attempted.value && (form.password.length < 4 || form.password.length > 32)
    ? t("staffs.passwordInvalid")
    : "",
);

watch(open, (isOpen) => {
  if (!isOpen) return;
  form.username = "";
  form.password = "";
  attempted.value = false;
});

function submit(): void {
  attempted.value = true;
  if (usernameError.value || passwordError.value) return;
  const request = { password: form.password };
  emit(
    "submit",
    props.mode === "create"
      ? { ...request, username: form.username.trim() }
      : request,
  );
}
</script>

<template>
  <Sheet v-model:open="open">
    <SheetContent>
      <form class="flex h-full flex-col gap-6" @submit.prevent="submit">
        <SheetHeader>
          <SheetTitle>
            {{
              t(
                mode === "create"
                  ? "staffs.createTitle"
                  : "staffs.passwordTitle",
              )
            }}
          </SheetTitle>
          <SheetDescription>
            {{
              t(
                mode === "create"
                  ? "staffs.createDescription"
                  : "staffs.passwordDescription",
                { username: staff?.username ?? "" },
              )
            }}
          </SheetDescription>
        </SheetHeader>

        <FieldGroup class="px-4">
          <Field
            v-if="mode === 'create'"
            :data-invalid="Boolean(usernameError)"
          >
            <FieldLabel for="staff-username">
              {{ t("staffs.username") }}
            </FieldLabel>
            <Input
              id="staff-username"
              v-model="form.username"
              :aria-invalid="Boolean(usernameError)"
              :disabled="pending"
              autocomplete="username"
              required
            />
            <FieldDescription v-if="usernameError" class="text-destructive">
              {{ usernameError }}
            </FieldDescription>
          </Field>
          <Field :data-invalid="Boolean(passwordError)">
            <FieldLabel for="staff-password">
              {{ t("staffs.password") }}
            </FieldLabel>
            <Input
              id="staff-password"
              v-model="form.password"
              :aria-invalid="Boolean(passwordError)"
              :autocomplete="mode === 'create' ? 'new-password' : 'off'"
              :disabled="pending"
              minlength="4"
              maxlength="32"
              required
              type="password"
            />
            <FieldDescription v-if="passwordError" class="text-destructive">
              {{ passwordError }}
            </FieldDescription>
            <FieldDescription v-else>
              {{ t("staffs.passwordHelp") }}
            </FieldDescription>
          </Field>
        </FieldGroup>

        <SheetFooter class="mt-auto">
          <Button type="submit" :disabled="pending">
            <Spinner v-if="pending" data-icon="inline-start" />
            {{ t(mode === "create" ? "staffs.create" : "staffs.savePassword") }}
          </Button>
        </SheetFooter>
      </form>
    </SheetContent>
  </Sheet>
</template>

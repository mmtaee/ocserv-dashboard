<script setup lang="ts">
import { computed, onBeforeUnmount, reactive, shallowRef } from "vue";
import { useI18n } from "vue-i18n";
import { useRouter } from "vue-router";

import {
  resetAdminPassword,
  type ResetAdminPasswordRequest,
} from "@/api/services/auth";
import { normalizeApiError } from "@/api/http";
import { Alert, AlertDescription, AlertTitle } from "@/components/ui/alert";
import { Button } from "@/components/ui/button";
import {
  Field,
  FieldError,
  FieldGroup,
  FieldLabel,
} from "@/components/ui/field";
import { Input } from "@/components/ui/input";
import { Spinner } from "@/components/ui/spinner";

const { t } = useI18n({ useScope: "global" });
const router = useRouter();
const pending = shallowRef(false);
const attempted = shallowRef(false);
const error = shallowRef("");
const success = shallowRef(false);
const confirmation = shallowRef("");
const form = reactive<ResetAdminPasswordRequest>({
  new_password: "",
  secret_key: "",
});
let redirectTimer: ReturnType<typeof setTimeout> | undefined;

const secretKeyError = computed(() =>
  attempted.value &&
  (form.secret_key.length < 16 || form.secret_key.length > 64)
    ? t("auth.secretKeyLength")
    : "",
);
const passwordError = computed(() =>
  attempted.value &&
  (form.new_password.length < 4 || form.new_password.length > 32)
    ? t("auth.passwordLength")
    : "",
);
const confirmationError = computed(() =>
  attempted.value && form.new_password !== confirmation.value
    ? t("auth.passwordMismatch")
    : "",
);
const valid = computed(
  () =>
    Boolean(form.secret_key && form.new_password && confirmation.value) &&
    !secretKeyError.value &&
    !passwordError.value &&
    !confirmationError.value,
);

function resetForm(): void {
  form.secret_key = "";
  form.new_password = "";
  confirmation.value = "";
  attempted.value = false;
  error.value = "";
}

function cancelRedirect(): void {
  if (redirectTimer !== undefined) clearTimeout(redirectTimer);
}

async function submit(): Promise<void> {
  attempted.value = true;
  if (!valid.value || pending.value) return;

  pending.value = true;
  error.value = "";
  try {
    await resetAdminPassword({ ...form });
    resetForm();
    success.value = true;
    redirectTimer = setTimeout(
      () => void router.replace({ name: "login" }),
      800,
    );
  } catch (cause) {
    error.value = normalizeApiError(cause).message;
  } finally {
    pending.value = false;
  }
}

onBeforeUnmount(() => {
  cancelRedirect();
  resetForm();
});
</script>

<template>
  <form class="flex flex-col gap-6" @submit.prevent="submit">
    <FieldGroup>
      <div class="flex flex-col items-center gap-1 text-center">
        <h1 class="text-2xl font-bold">{{ t("auth.resetTitle") }}</h1>
        <p class="text-balance text-sm text-muted-foreground">
          {{ t("auth.resetSubtitle") }}
        </p>
      </div>

      <Alert v-if="error" variant="error">
        <AlertTitle>{{ t("auth.resetFailure") }}</AlertTitle>
        <AlertDescription>{{ error }}</AlertDescription>
      </Alert>
      <Alert v-if="success" variant="success">
        <AlertTitle>{{ t("auth.resetSuccess") }}</AlertTitle>
      </Alert>

      <Field :data-invalid="Boolean(secretKeyError)">
        <FieldLabel for="reset-secret-key">{{
          t("auth.secretKey")
        }}</FieldLabel>
        <Input
          id="reset-secret-key"
          v-model="form.secret_key"
          type="password"
          autocomplete="off"
          minlength="16"
          maxlength="64"
          required
          :aria-invalid="Boolean(secretKeyError)"
          :disabled="pending || success"
        />
        <FieldError v-if="secretKeyError">{{ secretKeyError }}</FieldError>
      </Field>
      <Field :data-invalid="Boolean(passwordError)">
        <FieldLabel for="reset-new-password">{{
          t("auth.newPassword")
        }}</FieldLabel>
        <Input
          id="reset-new-password"
          v-model="form.new_password"
          type="password"
          autocomplete="new-password"
          minlength="4"
          maxlength="32"
          required
          :aria-invalid="Boolean(passwordError)"
          :disabled="pending || success"
        />
        <FieldError v-if="passwordError">{{ passwordError }}</FieldError>
      </Field>
      <Field :data-invalid="Boolean(confirmationError)">
        <FieldLabel for="reset-confirm-password">{{
          t("auth.confirmPassword")
        }}</FieldLabel>
        <Input
          id="reset-confirm-password"
          v-model="confirmation"
          type="password"
          autocomplete="new-password"
          required
          :aria-invalid="Boolean(confirmationError)"
          :disabled="pending || success"
        />
        <FieldError v-if="confirmationError">{{
          confirmationError
        }}</FieldError>
      </Field>
      <Field>
        <Button type="submit" :disabled="pending || success || !valid">
          <Spinner v-if="pending" data-icon="inline-start" />
          {{ t("auth.resetSubmit") }}
        </Button>
      </Field>
      <div class="text-center text-sm">
        <RouterLink
          :to="{ name: 'login' }"
          class="underline underline-offset-4 hover:text-primary"
        >
          {{ t("auth.backToLogin") }}
        </RouterLink>
      </div>
    </FieldGroup>
  </form>
</template>

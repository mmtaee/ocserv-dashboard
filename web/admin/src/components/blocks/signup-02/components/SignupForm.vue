<script setup lang="ts">
import type { HTMLAttributes } from "vue";
import { reactive } from "vue";
import { useI18n } from "vue-i18n";

import type { SystemUpdateData } from "@/api/services/system";
import { Alert, AlertDescription, AlertTitle } from "@/components/ui/alert";
import { Button } from "@/components/ui/button";
import { Checkbox } from "@/components/ui/checkbox";
import {
  Field,
  FieldDescription,
  FieldGroup,
  FieldLabel,
} from "@/components/ui/field";
import { Input } from "@/components/ui/input";
import { Spinner } from "@/components/ui/spinner";
import { cn } from "@/lib/utils";

const props = defineProps<{
  class?: HTMLAttributes["class"];
  error?: string | null;
  pending?: boolean;
}>();
const emit = defineEmits<{ submit: [settings: SystemUpdateData] }>();
const { t } = useI18n({ useScope: "global" });

const form = reactive<SystemUpdateData>({
  auto_delete_inactive_users: false,
  client_profile_connection_name: t("setup.defaultConnectionName"),
  client_profile_server_address: window.location.hostname,
  client_profile_server_port: 443,
  google_captcha_secret_key: "",
  google_captcha_site_key: "",
  keep_inactive_user_days: 30,
});
</script>

<template>
  <form
    :class="cn('flex flex-col gap-6', props.class)"
    @submit.prevent="emit('submit', { ...form })"
  >
    <FieldGroup>
      <div class="flex flex-col gap-1">
        <h1 class="text-2xl font-bold">{{ t("setup.title") }}</h1>
        <p class="text-sm text-muted-foreground">
          {{ t("setup.description") }}
        </p>
      </div>

      <Alert v-if="error" variant="destructive">
        <AlertTitle>{{ t("setup.failure") }}</AlertTitle>
        <AlertDescription>{{ error }}</AlertDescription>
      </Alert>

      <div class="grid gap-5 sm:grid-cols-2">
        <Field class="sm:col-span-2">
          <FieldLabel for="client-profile-server-address">{{
            t("setup.serverAddress")
          }}</FieldLabel>
          <Input
            id="client-profile-server-address"
            v-model="form.client_profile_server_address"
            name="client_profile_server_address"
            :placeholder="t('setup.serverAddressPlaceholder')"
            required
          />
        </Field>

        <Field>
          <FieldLabel for="client-profile-server-port">{{
            t("setup.serverPort")
          }}</FieldLabel>
          <Input
            id="client-profile-server-port"
            :model-value="form.client_profile_server_port"
            name="client_profile_server_port"
            type="number"
            min="1"
            max="65535"
            required
            @update:model-value="
              form.client_profile_server_port = Number($event)
            "
          />
        </Field>

        <Field>
          <FieldLabel for="client-profile-connection-name">{{
            t("setup.connectionName")
          }}</FieldLabel>
          <Input
            id="client-profile-connection-name"
            v-model="form.client_profile_connection_name"
            name="client_profile_connection_name"
            required
          />
        </Field>

        <Field>
          <FieldLabel for="keep-inactive-user-days">{{
            t("setup.retention")
          }}</FieldLabel>
          <Input
            id="keep-inactive-user-days"
            :model-value="form.keep_inactive_user_days"
            name="keep_inactive_user_days"
            type="number"
            min="1"
            required
            @update:model-value="form.keep_inactive_user_days = Number($event)"
          />
          <FieldDescription>{{ t("setup.retentionHelp") }}</FieldDescription>
        </Field>

        <Field
          orientation="horizontal"
          class="items-start rounded-lg border p-3"
        >
          <Checkbox
            id="auto-delete-inactive-users"
            :model-value="form.auto_delete_inactive_users"
            @update:model-value="
              form.auto_delete_inactive_users = $event === true
            "
          />
          <div class="grid gap-1">
            <FieldLabel for="auto-delete-inactive-users">{{
              t("setup.autoDelete")
            }}</FieldLabel>
            <FieldDescription>{{ t("setup.autoDeleteHelp") }}</FieldDescription>
          </div>
        </Field>

        <Field>
          <FieldLabel for="google-captcha-site-key">{{
            t("setup.captchaSiteKey")
          }}</FieldLabel>
          <Input
            id="google-captcha-site-key"
            v-model="form.google_captcha_site_key"
            name="google_captcha_site_key"
            autocomplete="off"
          />
        </Field>

        <Field>
          <FieldLabel for="google-captcha-secret-key">{{
            t("setup.captchaSecretKey")
          }}</FieldLabel>
          <Input
            id="google-captcha-secret-key"
            v-model="form.google_captcha_secret_key"
            name="google_captcha_secret_key"
            type="password"
            autocomplete="new-password"
          />
        </Field>
      </div>

      <Field>
        <Button type="submit" :disabled="pending">
          <Spinner v-if="pending" data-icon="inline-start" />
          {{ t("setup.submit") }}
        </Button>
      </Field>
    </FieldGroup>
  </form>
</template>

<script setup lang="ts">
import { Send, Save } from "@lucide/vue";
import { onMounted, reactive, shallowRef } from "vue";
import { useI18n } from "vue-i18n";
import { normalizeApiError } from "@/api/http";
import {
  getTelegramSettings,
  telegramLanguages,
  testTelegram,
  updateTelegramSettings,
  type TelegramSettingsUpdate,
} from "@/api/services/telegram";
import { Alert, AlertDescription, AlertTitle } from "@/components/ui/alert";
import { Button } from "@/components/ui/button";
import {
  Card,
  CardContent,
  CardDescription,
  CardFooter,
  CardHeader,
  CardTitle,
} from "@/components/ui/card";
import { Checkbox } from "@/components/ui/checkbox";
import {
  Field,
  FieldDescription,
  FieldError,
  FieldGroup,
  FieldLabel,
} from "@/components/ui/field";
import { Input } from "@/components/ui/input";
import {
  Select,
  SelectContent,
  SelectGroup,
  SelectItem,
  SelectTrigger,
  SelectValue,
} from "@/components/ui/select";
import { Spinner } from "@/components/ui/spinner";
import { Skeleton } from "@/components/ui/skeleton";
import { Textarea } from "@/components/ui/textarea";

const { t } = useI18n({ useScope: "global" });
const form = reactive({
  enabled: false,
  bot_token: "",
  admin_chat_id: "",
  low_quota_threshold_mb: "512",
  default_language: "en",
  ocserv_host: "",
  card_number: "",
  card_holder: "",
  support_username: "",
});
const botUsername = shallowRef("");
const testMessage = shallowRef("");
const loading = shallowRef(true);
const saving = shallowRef(false);
const testing = shallowRef(false);
const error = shallowRef("");
const success = shallowRef("");
const validation = shallowRef("");

async function load(): Promise<void> {
  loading.value = true;
  error.value = "";
  try {
    const data = await getTelegramSettings();
    Object.assign(form, {
      enabled: data.enabled ?? false,
      bot_token: data.bot_token ?? "",
      admin_chat_id: data.admin_chat_id?.toString() ?? "",
      low_quota_threshold_mb: data.low_quota_threshold_mb?.toString() ?? "512",
      default_language: data.default_language ?? "en",
      ocserv_host: data.ocserv_host ?? "",
      card_number: data.card_number ?? "",
      card_holder: data.card_holder ?? "",
      support_username: data.support_username ?? "",
    });
    botUsername.value = data.bot_username ?? "";
  } catch (cause) {
    error.value = normalizeApiError(cause).message;
  } finally {
    loading.value = false;
  }
}
function payload(): TelegramSettingsUpdate | null {
  const threshold = Number(form.low_quota_threshold_mb);
  const chatId = Number(form.admin_chat_id);
  if (!Number.isInteger(threshold) || threshold < 10 || threshold > 10240) {
    validation.value = t("telegram.validation.threshold");
    return null;
  }
  if (form.admin_chat_id && !Number.isInteger(chatId)) {
    validation.value = t("telegram.validation.chatId");
    return null;
  }
  if (
    form.card_number.length > 64 ||
    form.card_holder.length > 128 ||
    form.support_username.length > 64
  ) {
    validation.value = t("telegram.validation.length");
    return null;
  }
  validation.value = "";
  return {
    ...form,
    admin_chat_id: chatId || 0,
    low_quota_threshold_mb: threshold,
    default_language:
      form.default_language as TelegramSettingsUpdate["default_language"],
    support_username: form.support_username.replace(/^@/, ""),
  };
}
async function save(): Promise<void> {
  const request = payload();
  if (!request) return;
  saving.value = true;
  error.value = "";
  success.value = "";
  try {
    const data = await updateTelegramSettings(request);
    botUsername.value = data.bot_username ?? "";
    success.value = t("telegram.settingsSaved");
  } catch (cause) {
    error.value = normalizeApiError(cause).message;
  } finally {
    saving.value = false;
  }
}
async function sendTest(): Promise<void> {
  testing.value = true;
  error.value = "";
  success.value = "";
  try {
    await testTelegram(testMessage.value);
    success.value = t("telegram.testSent");
  } catch (cause) {
    error.value = normalizeApiError(cause).message;
  } finally {
    testing.value = false;
  }
}
onMounted(load);
</script>

<template>
  <div class="flex flex-col gap-6">
    <Alert v-if="error" variant="error">
      <AlertTitle>{{ t("telegram.requestFailed") }}</AlertTitle>
      <AlertDescription>{{ error }}</AlertDescription>
    </Alert>
    <Alert v-if="success" variant="success">
      <AlertTitle>{{ t("telegram.success") }}</AlertTitle>
      <AlertDescription>{{ success }}</AlertDescription>
    </Alert>
    <Card>
      <CardHeader>
        <CardTitle>{{ t("navigation.telegramSettings") }}</CardTitle>
        <CardDescription>
          {{ t("telegram.settingsDescription") }}
        </CardDescription>
      </CardHeader>
      <CardContent>
        <Skeleton v-if="loading" class="h-80 w-full" />
        <form v-else @submit.prevent="save">
          <FieldGroup>
            <Field orientation="horizontal">
              <Checkbox id="telegram-enabled" v-model="form.enabled" />
              <div>
                <FieldLabel for="telegram-enabled">
                  {{ t("telegram.enabled") }}
                </FieldLabel>
                <FieldDescription>
                  {{ t("telegram.enabledHelp") }}
                </FieldDescription>
              </div>
            </Field>
            <div class="grid gap-4 md:grid-cols-2">
              <Field>
                <FieldLabel for="bot-token">
                  {{ t("telegram.botToken") }}
                </FieldLabel>
                <Input
                  id="bot-token"
                  v-model="form.bot_token"
                  type="password"
                />
              </Field>
              <Field>
                <FieldLabel for="bot-username">
                  {{ t("telegram.botUsername") }}
                </FieldLabel>
                <Input id="bot-username" :model-value="botUsername" disabled />
              </Field>
            </div>
            <div class="grid gap-4 md:grid-cols-2">
              <Field>
                <FieldLabel for="admin-chat">
                  {{ t("telegram.adminChatId") }}
                </FieldLabel>
                <Input
                  id="admin-chat"
                  v-model="form.admin_chat_id"
                  inputmode="numeric"
                />
              </Field>
              <Field>
                <FieldLabel>{{ t("telegram.defaultLanguage") }}</FieldLabel>
                <Select v-model="form.default_language">
                  <SelectTrigger><SelectValue /></SelectTrigger>
                  <SelectContent>
                    <SelectGroup>
                      <SelectItem
                        v-for="language in telegramLanguages"
                        :key="language"
                        :value="language"
                      >
                        {{ t(`telegram.languages.${language}`) }}
                      </SelectItem>
                    </SelectGroup>
                  </SelectContent>
                </Select>
              </Field>
            </div>
            <div class="grid gap-4 md:grid-cols-2">
              <Field>
                <FieldLabel for="quota">
                  {{ t("telegram.lowQuotaThreshold") }}
                </FieldLabel>
                <Input
                  id="quota"
                  v-model="form.low_quota_threshold_mb"
                  type="number"
                  min="10"
                  max="10240"
                />
              </Field>
              <Field>
                <FieldLabel for="host">
                  {{ t("telegram.ocservHost") }}
                </FieldLabel>
                <Input id="host" v-model="form.ocserv_host" />
              </Field>
            </div>
            <div class="grid gap-4 md:grid-cols-2">
              <Field>
                <FieldLabel for="card-number">
                  {{ t("telegram.cardNumber") }}
                </FieldLabel>
                <Input
                  id="card-number"
                  v-model="form.card_number"
                  maxlength="64"
                />
              </Field>
              <Field>
                <FieldLabel for="card-holder">
                  {{ t("telegram.cardHolder") }}
                </FieldLabel>
                <Input
                  id="card-holder"
                  v-model="form.card_holder"
                  maxlength="128"
                />
              </Field>
            </div>
            <Field>
              <FieldLabel for="support">
                {{ t("telegram.supportUsername") }}
              </FieldLabel>
              <Input
                id="support"
                v-model="form.support_username"
                maxlength="64"
              />
              <FieldDescription>
                {{ t("telegram.supportUsernameHelp") }}
              </FieldDescription>
            </Field>
            <FieldError v-if="validation">{{ validation }}</FieldError>
          </FieldGroup>
        </form>
      </CardContent>
      <CardFooter>
        <Button type="button" :disabled="loading || saving" @click="save">
          <Spinner v-if="saving" data-icon="inline-start" />
          <Save v-else data-icon="inline-start" />
          {{ t("telegram.save") }}
        </Button>
      </CardFooter>
    </Card>
    <Card>
      <CardHeader>
        <CardTitle>{{ t("telegram.testTitle") }}</CardTitle>
        <CardDescription>{{ t("telegram.testDescription") }}</CardDescription>
      </CardHeader>
      <CardContent>
        <Field>
          <FieldLabel for="test-message">
            {{ t("telegram.testMessage") }}
          </FieldLabel>
          <Textarea id="test-message" v-model="testMessage" />
        </Field>
      </CardContent>
      <CardFooter>
        <Button
          type="button"
          variant="outline"
          :disabled="testing || loading"
          @click="sendTest"
        >
          <Spinner v-if="testing" data-icon="inline-start" />
          <Send v-else data-icon="inline-start" />
          {{ t("telegram.sendTest") }}
        </Button>
      </CardFooter>
    </Card>
  </div>
</template>

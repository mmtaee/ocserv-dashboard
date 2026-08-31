<script setup lang="ts">
import { computed, reactive, shallowRef, watch } from "vue";
import { useI18n } from "vue-i18n";

import type {
  OcservUser,
  OcservUserCreate,
  OcservUserUpdate,
} from "@/api/services/ocserv-users";
import type {
  ModelsExpiryMode,
  ModelsOcservUserConfig,
  ModelsTrafficType,
} from "@/api/generated";
import OcservUserConfigFields from "@/components/ocserv-users/OcservUserConfigFields.vue";
import { Button } from "@/components/ui/button";
import { Checkbox } from "@/components/ui/checkbox";
import {
  Field,
  FieldDescription,
  FieldGroup,
  FieldLabel,
} from "@/components/ui/field";
import { Input } from "@/components/ui/input";
import {
  Select,
  SelectContent,
  SelectItem,
  SelectTrigger,
  SelectValue,
} from "@/components/ui/select";
import {
  Sheet,
  SheetContent,
  SheetDescription,
  SheetFooter,
  SheetHeader,
  SheetTitle,
} from "@/components/ui/sheet";
import { Skeleton } from "@/components/ui/skeleton";
import { Spinner } from "@/components/ui/spinner";
import { Textarea } from "@/components/ui/textarea";

const props = defineProps<{
  groupNames: string[];
  loading?: boolean;
  mode: "create" | "edit";
  pending?: boolean;
  user?: OcservUser | null;
}>();
const emit = defineEmits<{
  submit: [request: OcservUserCreate | OcservUserUpdate];
}>();
const open = defineModel<boolean>("open", { default: false });
const { t } = useI18n({ useScope: "global" });
const attempted = shallowRef(false);
const config = shallowRef<ModelsOcservUserConfig>({});
const form = reactive({
  description: "",
  expireAt: "",
  expireDays: 30,
  expiryMode: "unlimited" as ModelsExpiryMode,
  group: "",
  password: "",
  trafficSizeGiB: 0,
  trafficType: "Free" as ModelsTrafficType,
  unlimited: true,
  username: "",
});

const trafficTypes: ModelsTrafficType[] = [
  "Free",
  "MonthlyTransmit",
  "MonthlyReceive",
  "MonthlyRxTx",
  "TotallyTransmit",
  "TotallyReceive",
  "TotallyRxTx",
];
const usernameError = computed(() => {
  if (!attempted.value || props.mode === "edit") return "";
  const length = form.username.trim().length;
  return length < 2 || length > 32 ? t("ocservUsers.usernameInvalid") : "";
});
const passwordError = computed(() => {
  if (!attempted.value || (props.mode === "edit" && !form.password)) return "";
  return form.password.length < 2 || form.password.length > 32
    ? t("ocservUsers.passwordInvalid")
    : "";
});
const groupError = computed(() =>
  attempted.value && !form.group ? t("ocservUsers.groupRequired") : "",
);
const expiryError = computed(() => {
  if (!attempted.value) return "";
  if (form.expiryMode === "fixed" && !form.expireAt)
    return t("ocservUsers.expireAtRequired");
  if (form.expiryMode === "first_connection" && form.expireDays < 1)
    return t("ocservUsers.expireDaysRequired");
  return "";
});

watch(
  [open, () => props.user, () => props.mode],
  ([isOpen]) => {
    if (!isOpen) return;
    const user = props.user;
    form.username = user?.username ?? "";
    form.password = "";
    form.group = user?.group ?? props.groupNames[0] ?? "";
    form.description = user?.description ?? "";
    form.expiryMode = user?.expiry_mode ?? "unlimited";
    form.expireAt = user?.expire_at?.slice(0, 10) ?? "";
    form.expireDays = user?.expire_days_after_first_connection ?? 30;
    form.trafficType = user?.traffic_type ?? "Free";
    form.trafficSizeGiB = (user?.traffic_size ?? 0) / 1024 ** 3;
    form.unlimited = user?.traffic_type === "Free" || !user?.traffic_size;
    config.value = structuredClone(user?.config ?? {});
    attempted.value = false;
  },
  { immediate: true },
);

function updateConfig(
  key: keyof ModelsOcservUserConfig,
  value: ModelsOcservUserConfig[keyof ModelsOcservUserConfig],
): void {
  const next = { ...config.value };
  if (value === undefined || value === "") delete next[key];
  else Object.assign(next, { [key]: value });
  config.value = next;
}

function submit(): void {
  attempted.value = true;
  if (
    usernameError.value ||
    passwordError.value ||
    groupError.value ||
    expiryError.value
  )
    return;

  const common: OcservUserUpdate = {
    config: structuredClone(config.value),
    description: form.description.trim(),
    expire_at: form.expiryMode === "fixed" ? form.expireAt : undefined,
    expire_days_after_first_connection:
      form.expiryMode === "first_connection" ? form.expireDays : undefined,
    expiry_mode: form.expiryMode,
    group: form.group,
    traffic_size: Math.round(form.trafficSizeGiB * 1024 ** 3),
    traffic_type: form.trafficType,
    unlimited: form.unlimited,
  };
  if (form.password) common.password = form.password;

  if (props.mode === "edit") {
    emit("submit", common);
    return;
  }
  emit("submit", {
    ...common,
    config: common.config ?? {},
    group: common.group ?? "",
    password: form.password,
    traffic_type: common.traffic_type ?? "Free",
    username: form.username.trim(),
  });
}
</script>

<template>
  <Sheet v-model:open="open">
    <SheetContent class="overflow-y-auto sm:max-w-3xl">
      <form class="flex flex-col gap-6" @submit.prevent="submit">
        <SheetHeader>
          <SheetTitle>
            {{
              t(
                mode === "create"
                  ? "ocservUsers.createTitle"
                  : "ocservUsers.editTitle",
              )
            }}
          </SheetTitle>
          <SheetDescription>
            {{
              t(
                mode === "create"
                  ? "ocservUsers.createDescription"
                  : "ocservUsers.editDescription",
              )
            }}
          </SheetDescription>
        </SheetHeader>

        <div v-if="loading" class="flex flex-col gap-4 px-4" aria-busy="true">
          <Skeleton v-for="index in 7" :key="index" class="h-16 w-full" />
        </div>

        <div v-else class="flex flex-col gap-6 px-4">
          <FieldGroup class="grid gap-5 sm:grid-cols-2">
            <Field :data-invalid="Boolean(usernameError)">
              <FieldLabel for="ocserv-user-username">
                {{ t("ocservUsers.username") }}
              </FieldLabel>
              <Input
                id="ocserv-user-username"
                v-model="form.username"
                :disabled="mode === 'edit' || pending"
                minlength="2"
                maxlength="32"
                required
              />
              <FieldDescription v-if="usernameError" class="text-destructive">
                {{ usernameError }}
              </FieldDescription>
            </Field>
            <Field :data-invalid="Boolean(passwordError)">
              <FieldLabel for="ocserv-user-password">
                {{ t("ocservUsers.password") }}
              </FieldLabel>
              <Input
                id="ocserv-user-password"
                v-model="form.password"
                :disabled="pending"
                :required="mode === 'create'"
                type="password"
                minlength="2"
                maxlength="32"
              />
              <FieldDescription v-if="passwordError" class="text-destructive">
                {{ passwordError }}
              </FieldDescription>
              <FieldDescription v-else-if="mode === 'edit'">
                {{ t("ocservUsers.passwordOptional") }}
              </FieldDescription>
            </Field>
            <Field :data-invalid="Boolean(groupError)">
              <FieldLabel for="ocserv-user-group">
                {{ t("ocservUsers.group") }}
              </FieldLabel>
              <Select v-model="form.group" :disabled="pending">
                <SelectTrigger id="ocserv-user-group">
                  <SelectValue :placeholder="t('ocservUsers.selectGroup')" />
                </SelectTrigger>
                <SelectContent>
                  <SelectItem
                    v-for="name in groupNames"
                    :key="name"
                    :value="name"
                  >
                    {{ name }}
                  </SelectItem>
                </SelectContent>
              </Select>
              <FieldDescription v-if="groupError" class="text-destructive">
                {{ groupError }}
              </FieldDescription>
            </Field>
            <Field>
              <FieldLabel for="ocserv-user-traffic-type">
                {{ t("ocservUsers.trafficType") }}
              </FieldLabel>
              <Select v-model="form.trafficType" :disabled="pending">
                <SelectTrigger id="ocserv-user-traffic-type">
                  <SelectValue />
                </SelectTrigger>
                <SelectContent>
                  <SelectItem
                    v-for="trafficType in trafficTypes"
                    :key="trafficType"
                    :value="trafficType"
                  >
                    {{ t(`ocservUsers.trafficTypes.${trafficType}`) }}
                  </SelectItem>
                </SelectContent>
              </Select>
            </Field>
            <Field>
              <FieldLabel for="ocserv-user-traffic-size">
                {{ t("ocservUsers.trafficSize") }}
              </FieldLabel>
              <Input
                id="ocserv-user-traffic-size"
                v-model="form.trafficSizeGiB"
                :disabled="pending || form.unlimited"
                min="0"
                step="0.1"
                type="number"
              />
            </Field>
            <Field
              orientation="horizontal"
              class="items-center rounded-lg border p-3"
            >
              <Checkbox
                id="ocserv-user-unlimited"
                :disabled="pending"
                :model-value="form.unlimited"
                @update:model-value="form.unlimited = $event === true"
              />
              <FieldLabel for="ocserv-user-unlimited">
                {{ t("ocservUsers.unlimitedTraffic") }}
              </FieldLabel>
            </Field>
            <Field class="sm:col-span-2">
              <FieldLabel for="ocserv-user-expiry-mode">
                {{ t("ocservUsers.expiryMode") }}
              </FieldLabel>
              <Select v-model="form.expiryMode" :disabled="pending">
                <SelectTrigger id="ocserv-user-expiry-mode">
                  <SelectValue />
                </SelectTrigger>
                <SelectContent>
                  <SelectItem value="unlimited">{{
                    t("ocservUsers.expiryUnlimited")
                  }}</SelectItem>
                  <SelectItem value="fixed">{{
                    t("ocservUsers.expiryFixed")
                  }}</SelectItem>
                  <SelectItem value="first_connection">{{
                    t("ocservUsers.expiryFirstConnection")
                  }}</SelectItem>
                </SelectContent>
              </Select>
            </Field>
            <Field
              v-if="form.expiryMode === 'fixed'"
              :data-invalid="Boolean(expiryError)"
            >
              <FieldLabel for="ocserv-user-expire-at">{{
                t("ocservUsers.expireAt")
              }}</FieldLabel>
              <Input
                id="ocserv-user-expire-at"
                v-model="form.expireAt"
                :disabled="pending"
                type="date"
                required
              />
              <FieldDescription v-if="expiryError" class="text-destructive">{{
                expiryError
              }}</FieldDescription>
            </Field>
            <Field
              v-if="form.expiryMode === 'first_connection'"
              :data-invalid="Boolean(expiryError)"
            >
              <FieldLabel for="ocserv-user-expire-days">{{
                t("ocservUsers.expireDays")
              }}</FieldLabel>
              <Input
                id="ocserv-user-expire-days"
                v-model="form.expireDays"
                :disabled="pending"
                min="1"
                type="number"
                required
              />
              <FieldDescription v-if="expiryError" class="text-destructive">{{
                expiryError
              }}</FieldDescription>
            </Field>
            <Field class="sm:col-span-2">
              <FieldLabel for="ocserv-user-description">{{
                t("ocservUsers.userDescription")
              }}</FieldLabel>
              <Textarea
                id="ocserv-user-description"
                v-model="form.description"
                :disabled="pending"
                maxlength="1024"
                rows="3"
              />
            </Field>
          </FieldGroup>

          <div class="grid gap-2">
            <h3 class="font-medium">{{ t("ocservUsers.configuration") }}</h3>
            <p class="text-sm text-muted-foreground">
              {{ t("ocservUsers.configurationDescription") }}
            </p>
          </div>
          <OcservUserConfigFields
            :config="config"
            :disabled="pending"
            @change="updateConfig"
          />
        </div>

        <SheetFooter v-if="!loading">
          <Button type="button" variant="outline" @click="open = false">
            {{ t("ocservUsers.cancel") }}
          </Button>
          <Button type="submit" :disabled="pending">
            <Spinner v-if="pending" data-icon="inline-start" />
            {{ pending ? t("ocservUsers.saving") : t("ocservUsers.save") }}
          </Button>
        </SheetFooter>
      </form>
    </SheetContent>
  </Sheet>
</template>

<script setup lang="ts">
import { reactive, shallowRef, watch } from "vue";
import { useI18n } from "vue-i18n";

import type {
  OcservUser,
  OcservUserActivation,
} from "@/api/services/ocserv-users";
import type { ModelsExpiryMode } from "@/api/generated";
import { Button } from "@/components/ui/button";
import { Checkbox } from "@/components/ui/checkbox";
import { Field, FieldLabel } from "@/components/ui/field";
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
import { Spinner } from "@/components/ui/spinner";

defineProps<{ pending?: boolean; user?: OcservUser | null }>();
const emit = defineEmits<{ submit: [request: OcservUserActivation] }>();
const open = defineModel<boolean>("open", { default: false });
const { t } = useI18n({ useScope: "global" });
const attempted = shallowRef(false);
const form = reactive({
  expireAt: "",
  expireDays: 30,
  expiryMode: "unlimited" as ModelsExpiryMode,
  resetFirstConnection: false,
});

watch(open, (isOpen) => {
  if (!isOpen) return;
  form.expireAt = "";
  form.expireDays = 30;
  form.expiryMode = "unlimited";
  form.resetFirstConnection = false;
  attempted.value = false;
});

function submit(): void {
  attempted.value = true;
  if (form.expiryMode === "fixed" && !form.expireAt) return;
  if (form.expiryMode === "first_connection" && form.expireDays < 1) return;
  emit("submit", {
    expire_at: form.expiryMode === "fixed" ? form.expireAt : undefined,
    expire_days_after_first_connection:
      form.expiryMode === "first_connection" ? form.expireDays : undefined,
    expiry_mode: form.expiryMode,
    reset_first_connection: form.resetFirstConnection,
  });
}
</script>

<template>
  <Sheet v-model:open="open">
    <SheetContent>
      <form class="flex flex-col gap-6" @submit.prevent="submit">
        <SheetHeader>
          <SheetTitle>{{ t("ocservUsers.activateTitle") }}</SheetTitle>
          <SheetDescription>{{
            t("ocservUsers.activateDescription", {
              username: user?.username ?? "",
            })
          }}</SheetDescription>
        </SheetHeader>
        <div class="grid gap-5 px-4">
          <Field>
            <FieldLabel for="activation-expiry-mode">{{
              t("ocservUsers.expiryMode")
            }}</FieldLabel>
            <Select v-model="form.expiryMode" :disabled="pending">
              <SelectTrigger id="activation-expiry-mode"
                ><SelectValue
              /></SelectTrigger>
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
            :data-invalid="attempted && !form.expireAt"
          >
            <FieldLabel for="activation-expire-at">{{
              t("ocservUsers.expireAt")
            }}</FieldLabel>
            <Input
              id="activation-expire-at"
              v-model="form.expireAt"
              :disabled="pending"
              required
              type="date"
            />
          </Field>
          <Field
            v-if="form.expiryMode === 'first_connection'"
            :data-invalid="attempted && form.expireDays < 1"
          >
            <FieldLabel for="activation-expire-days">{{
              t("ocservUsers.expireDays")
            }}</FieldLabel>
            <Input
              id="activation-expire-days"
              v-model="form.expireDays"
              :disabled="pending"
              min="1"
              required
              type="number"
            />
          </Field>
          <Field
            orientation="horizontal"
            class="items-center rounded-lg border p-3"
          >
            <Checkbox
              id="activation-reset"
              :disabled="pending"
              :model-value="form.resetFirstConnection"
              @update:model-value="form.resetFirstConnection = $event === true"
            />
            <FieldLabel for="activation-reset">{{
              t("ocservUsers.resetFirstConnection")
            }}</FieldLabel>
          </Field>
        </div>
        <SheetFooter>
          <Button type="button" variant="outline" @click="open = false">{{
            t("ocservUsers.cancel")
          }}</Button>
          <Button type="submit" :disabled="pending">
            <Spinner v-if="pending" data-icon="inline-start" />
            {{ t("ocservUsers.activate") }}
          </Button>
        </SheetFooter>
      </form>
    </SheetContent>
  </Sheet>
</template>

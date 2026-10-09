<script setup lang="ts">
import { computed, shallowRef } from "vue";
import { useI18n } from "vue-i18n";

import { Button } from "@/components/ui/button";
import {
  Field,
  FieldError,
  FieldGroup,
  FieldLabel,
} from "@/components/ui/field";
import { Input } from "@/components/ui/input";

const props = withDefaults(
  defineProps<{ loading?: boolean; optional?: boolean }>(),
  {
    loading: false,
    optional: false,
  },
);
const emit = defineEmits<{ apply: []; clear: [] }>();
const dateStart = defineModel<string>("dateStart", { required: true });
const dateEnd = defineModel<string>("dateEnd", { required: true });
const { t } = useI18n({ useScope: "global" });
const submitted = shallowRef(false);
const error = computed(() => {
  if (!submitted.value) return "";
  if (!dateStart.value && !dateEnd.value && props.optional) return "";
  if (!dateStart.value || !dateEnd.value) return t("reports.dateRequired");
  return dateStart.value > dateEnd.value ? t("reports.dateOrderInvalid") : "";
});

function submit(): void {
  submitted.value = true;
  if (!error.value) emit("apply");
}

function clear(): void {
  dateStart.value = "";
  dateEnd.value = "";
  submitted.value = false;
  emit("clear");
}
</script>

<template>
  <form class="flex flex-col gap-4" @submit.prevent="submit">
    <FieldGroup
      class="grid gap-4 sm:grid-cols-2 lg:grid-cols-[1fr_1fr_auto] lg:items-end"
    >
      <Field :data-invalid="Boolean(error)">
        <FieldLabel for="report-date-start">{{
          t("reports.dateStart")
        }}</FieldLabel>
        <Input
          id="report-date-start"
          v-model="dateStart"
          type="date"
          :aria-invalid="Boolean(error)"
        />
      </Field>
      <Field :data-invalid="Boolean(error)">
        <FieldLabel for="report-date-end">{{
          t("reports.dateEnd")
        }}</FieldLabel>
        <Input
          id="report-date-end"
          v-model="dateEnd"
          type="date"
          :aria-invalid="Boolean(error)"
        />
      </Field>
      <div class="flex gap-2">
        <Button type="submit" :disabled="loading">{{
          t("reports.apply")
        }}</Button>
        <Button
          v-if="optional"
          type="button"
          variant="outline"
          :disabled="loading"
          @click="clear"
        >
          {{ t("reports.clear") }}
        </Button>
      </div>
    </FieldGroup>
    <FieldError v-if="error">{{ error }}</FieldError>
  </form>
</template>

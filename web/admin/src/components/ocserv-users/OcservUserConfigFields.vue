<script setup lang="ts">
import { useI18n } from "vue-i18n";

import type { ModelsOcservUserConfig } from "@/api/generated";
import { Checkbox } from "@/components/ui/checkbox";
import { Field, FieldDescription, FieldLabel } from "@/components/ui/field";
import { Input } from "@/components/ui/input";
import { Textarea } from "@/components/ui/textarea";

type ConfigKey = keyof ModelsOcservUserConfig;
interface ConfigField {
  key: ConfigKey;
  kind: "boolean" | "list" | "number" | "text";
}

const props = defineProps<{
  config: ModelsOcservUserConfig;
  disabled?: boolean;
}>();
const emit = defineEmits<{
  change: [key: ConfigKey, value: ModelsOcservUserConfig[ConfigKey]];
}>();
const { t } = useI18n({ useScope: "global" });

const fields: ConfigField[] = [
  { key: "dns", kind: "list" },
  { key: "explicit-ipv4", kind: "text" },
  { key: "ipv4-network", kind: "text" },
  { key: "route", kind: "list" },
  { key: "no-route", kind: "list" },
  { key: "iroute", kind: "text" },
  { key: "split-dns", kind: "list" },
  { key: "nbns", kind: "text" },
  { key: "restrict-to-ports", kind: "text" },
  { key: "restrict-to-routes", kind: "boolean" },
  { key: "idle-timeout", kind: "number" },
  { key: "mobile-idle-timeout", kind: "number" },
  { key: "session-timeout", kind: "number" },
  { key: "rekey-time", kind: "number" },
];

function textValue(key: ConfigKey): string {
  const value = props.config[key];
  return Array.isArray(value) ? value.join("\n") : String(value ?? "");
}

function updateText(field: ConfigField, value: string | number): void {
  const text = String(value).trim();
  if (!text) {
    emit("change", field.key, undefined);
    return;
  }
  if (field.kind === "list") {
    emit(
      "change",
      field.key,
      text
        .split(/[,\n]/)
        .map((item) => item.trim())
        .filter(Boolean),
    );
    return;
  }
  emit("change", field.key, field.kind === "number" ? Number(text) : text);
}
</script>

<template>
  <div class="grid gap-5 sm:grid-cols-2">
    <Field
      v-for="field in fields"
      :key="field.key"
      :class="field.kind === 'list' ? 'sm:col-span-2' : ''"
      :data-disabled="disabled"
    >
      <div
        v-if="field.kind === 'boolean'"
        class="flex items-center justify-between gap-4 rounded-lg border p-3"
      >
        <FieldLabel :for="`user-config-${field.key}`">
          {{ t(`ocservUsers.config.${field.key}`) }}
        </FieldLabel>
        <Checkbox
          :id="`user-config-${field.key}`"
          :disabled="disabled"
          :model-value="Boolean(config[field.key])"
          @update:model-value="emit('change', field.key, Boolean($event))"
        />
      </div>
      <template v-else>
        <FieldLabel :for="`user-config-${field.key}`">
          {{ t(`ocservUsers.config.${field.key}`) }}
        </FieldLabel>
        <Textarea
          v-if="field.kind === 'list'"
          :id="`user-config-${field.key}`"
          :disabled="disabled"
          :model-value="textValue(field.key)"
          rows="2"
          @update:model-value="updateText(field, $event)"
        />
        <Input
          v-else
          :id="`user-config-${field.key}`"
          :disabled="disabled"
          :min="field.kind === 'number' ? 0 : undefined"
          :model-value="textValue(field.key)"
          :type="field.kind === 'number' ? 'number' : 'text'"
          @update:model-value="updateText(field, $event)"
        />
        <FieldDescription v-if="field.kind === 'list'">
          {{ t("ocservUsers.listHelp") }}
        </FieldDescription>
      </template>
    </Field>
  </div>
</template>

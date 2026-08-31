<script setup lang="ts">
import type { ModelsOcservGroupConfig } from "@/api/generated";

import type { DefaultGroupField } from "@/components/ocserv-groups/default-group-fields";
import {
  Card,
  CardContent,
  CardDescription,
  CardHeader,
  CardTitle,
} from "@/components/ui/card";
import { Checkbox } from "@/components/ui/checkbox";
import {
  Field,
  FieldDescription,
  FieldGroup,
  FieldLabel,
} from "@/components/ui/field";
import { Input } from "@/components/ui/input";

type ConfigValue = ModelsOcservGroupConfig[keyof ModelsOcservGroupConfig];

defineProps<{
  config: ModelsOcservGroupConfig;
  description: string;
  disabled?: boolean;
  fields: DefaultGroupField[];
  listHelp: string;
  title: string;
}>();

const emit = defineEmits<{
  change: [key: keyof ModelsOcservGroupConfig, value: ConfigValue];
}>();

function textValue(
  config: ModelsOcservGroupConfig,
  key: keyof ModelsOcservGroupConfig,
): string {
  const value = config[key];
  return typeof value === "string" ? value : "";
}

function numberValue(
  config: ModelsOcservGroupConfig,
  key: keyof ModelsOcservGroupConfig,
): number | "" {
  const value = config[key];
  return typeof value === "number" ? value : "";
}

function listValue(
  config: ModelsOcservGroupConfig,
  key: keyof ModelsOcservGroupConfig,
): string {
  const value = config[key];
  return Array.isArray(value) ? value.join(", ") : "";
}

function booleanValue(
  config: ModelsOcservGroupConfig,
  key: keyof ModelsOcservGroupConfig,
): boolean {
  return config[key] === true;
}

function setText(
  key: keyof ModelsOcservGroupConfig,
  value: string | number,
): void {
  const normalized = String(value).trim();
  emit("change", key, normalized || undefined);
}

function setNumber(
  key: keyof ModelsOcservGroupConfig,
  value: string | number,
): void {
  if (value === "") {
    emit("change", key, undefined);
    return;
  }
  emit("change", key, Number(value));
}

function setList(
  key: keyof ModelsOcservGroupConfig,
  value: string | number,
): void {
  const items = String(value)
    .split(",")
    .map((item) => item.trim())
    .filter(Boolean);
  emit("change", key, items);
}

function setBoolean(
  key: keyof ModelsOcservGroupConfig,
  value: boolean | "indeterminate",
): void {
  emit("change", key, value === true);
}
</script>

<template>
  <Card>
    <CardHeader>
      <CardTitle>
        {{ title }}
      </CardTitle>
      <CardDescription>
        {{ description }}
      </CardDescription>
    </CardHeader>
    <CardContent>
      <FieldGroup class="grid gap-5 md:grid-cols-2">
        <template v-for="field in fields" :key="field.key">
          <Field
            v-if="field.kind === 'boolean'"
            :data-disabled="disabled"
            orientation="horizontal"
            class="items-center rounded-lg border p-3"
          >
            <Checkbox
              :id="field.key"
              :disabled="disabled"
              :model-value="booleanValue(config, field.key)"
              @update:model-value="setBoolean(field.key, $event)"
            />
            <FieldLabel :for="field.key">
              {{ field.key }}
            </FieldLabel>
          </Field>

          <Field v-else :data-disabled="disabled">
            <FieldLabel :for="field.key">
              {{ field.key }}
            </FieldLabel>
            <Input
              v-if="field.kind === 'number'"
              :id="field.key"
              :disabled="disabled"
              :model-value="numberValue(config, field.key)"
              :name="field.key"
              :placeholder="field.placeholder"
              step="1"
              type="number"
              @update:model-value="setNumber(field.key, $event)"
            />
            <Input
              v-else-if="field.kind === 'list'"
              :id="field.key"
              :disabled="disabled"
              :model-value="listValue(config, field.key)"
              :name="field.key"
              :placeholder="field.placeholder"
              @update:model-value="setList(field.key, $event)"
            />
            <Input
              v-else
              :id="field.key"
              :disabled="disabled"
              :model-value="textValue(config, field.key)"
              :name="field.key"
              :placeholder="field.placeholder"
              @update:model-value="setText(field.key, $event)"
            />
            <FieldDescription v-if="field.kind === 'list'">
              {{ listHelp }}
            </FieldDescription>
          </Field>
        </template>
      </FieldGroup>
    </CardContent>
  </Card>
</template>

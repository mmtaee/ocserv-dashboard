<script setup lang="ts">
import { Play } from "@lucide/vue";
import { computed, shallowRef } from "vue";
import { useI18n } from "vue-i18n";

import type { OcctlAction, OcctlCommandRequest } from "@/api/services/occtl";
import {
  getOcctlCommand,
  occtlCommands,
} from "@/components/occtl/occtl-commands";
import { Button } from "@/components/ui/button";
import {
  Card,
  CardContent,
  CardDescription,
  CardFooter,
  CardHeader,
  CardTitle,
} from "@/components/ui/card";
import {
  Field,
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
  SelectLabel,
  SelectTrigger,
  SelectValue,
} from "@/components/ui/select";
import { Spinner } from "@/components/ui/spinner";

defineProps<{ pending?: boolean }>();
const emit = defineEmits<{ execute: [request: OcctlCommandRequest] }>();
const { t } = useI18n({ useScope: "global" });
const selectedAction = shallowRef("1");
const value = shallowRef("");
const validationError = shallowRef("");

const command = computed(() =>
  getOcctlCommand(Number(selectedAction.value) as OcctlAction),
);
const valueLabel = computed(() =>
  command.value.valueKind ? t(`occtl.values.${command.value.valueKind}`) : "",
);

function changeCommand(): void {
  value.value = "";
  validationError.value = "";
}

function submit(): void {
  const normalized = value.value.trim();
  if (command.value.valueKind && !normalized) {
    validationError.value = t("occtl.valueRequired", {
      field: valueLabel.value,
    });
    return;
  }
  validationError.value = "";
  emit("execute", {
    action: command.value.action,
    value: normalized || undefined,
  });
}
</script>

<template>
  <Card>
    <CardHeader>
      <CardTitle>{{ t("occtl.commandConsole") }}</CardTitle>
      <CardDescription>
        {{ t("occtl.commandConsoleDescription") }}
      </CardDescription>
    </CardHeader>
    <form @submit.prevent="submit">
      <CardContent>
        <FieldGroup>
          <Field>
            <FieldLabel for="occtl-command">
              {{ t("occtl.command") }}
            </FieldLabel>
            <Select
              v-model="selectedAction"
              @update:model-value="changeCommand"
            >
              <SelectTrigger id="occtl-command">
                <SelectValue :placeholder="t('occtl.selectCommand')" />
              </SelectTrigger>
              <SelectContent>
                <SelectGroup>
                  <SelectLabel>{{ t("occtl.availableCommands") }}</SelectLabel>
                  <SelectItem
                    v-for="item in occtlCommands"
                    :key="item.action"
                    :value="String(item.action)"
                  >
                    {{ item.action }}.
                    {{ t(`occtl.commands.${item.labelKey}`) }}
                  </SelectItem>
                </SelectGroup>
              </SelectContent>
            </Select>
          </Field>
          <Field
            v-if="command.valueKind"
            :data-invalid="Boolean(validationError)"
          >
            <FieldLabel for="occtl-command-value">{{ valueLabel }}</FieldLabel>
            <Input
              id="occtl-command-value"
              v-model="value"
              :aria-invalid="Boolean(validationError)"
              :placeholder="t(`occtl.placeholders.${command.valueKind}`)"
              autocomplete="off"
            />
            <FieldError :errors="[validationError]" />
          </Field>
        </FieldGroup>
      </CardContent>
      <CardFooter class="justify-end">
        <Button type="submit" :disabled="pending">
          <Spinner v-if="pending" data-icon="inline-start" />
          <Play v-else data-icon="inline-start" />
          {{ t("occtl.execute") }}
        </Button>
      </CardFooter>
    </form>
  </Card>
</template>

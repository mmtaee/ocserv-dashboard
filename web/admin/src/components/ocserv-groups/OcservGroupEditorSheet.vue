<script setup lang="ts">
import { computed, shallowRef, watch } from "vue";
import { useI18n } from "vue-i18n";

import type {
  OcservGroup,
  OcservGroupCreate,
  OcservGroupUpdate,
} from "@/api/services/ocserv-groups";
import type { ModelsOcservGroupConfig } from "@/api/generated";
import OcservGroupConfigFields from "@/components/ocserv-groups/OcservGroupConfigFields.vue";
import { cloneOcservGroupConfig } from "@/components/ocserv-groups/group-config";
import { Button } from "@/components/ui/button";
import { Field, FieldDescription, FieldLabel } from "@/components/ui/field";
import { Input } from "@/components/ui/input";
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

const props = defineProps<{
  existingNames: string[];
  group?: OcservGroup | null;
  loading?: boolean;
  mode: "create" | "edit";
  pending?: boolean;
}>();
const emit = defineEmits<{
  submit: [request: OcservGroupCreate | OcservGroupUpdate];
}>();
const open = defineModel<boolean>("open", { default: false });
const { t } = useI18n({ useScope: "global" });
const name = shallowRef("");
const config = shallowRef<ModelsOcservGroupConfig>({});
const attempted = shallowRef(false);

const normalizedName = computed(() => name.value.trim());
const nameError = computed(() => {
  if (!attempted.value || props.mode === "edit") return null;
  if (!normalizedName.value) return t("ocservGroups.nameRequired");
  if (props.existingNames.includes(normalizedName.value))
    return t("ocservGroups.nameExists");
  return null;
});

watch(
  [open, () => props.group, () => props.mode],
  ([isOpen]) => {
    if (!isOpen) return;
    name.value = props.group?.name ?? "";
    config.value = cloneOcservGroupConfig(props.group?.config);
    attempted.value = false;
  },
  { immediate: true },
);

function updateField(
  key: keyof ModelsOcservGroupConfig,
  value: ModelsOcservGroupConfig[keyof ModelsOcservGroupConfig],
): void {
  config.value = { ...config.value, [key]: value };
}

function submit(): void {
  attempted.value = true;
  if (props.mode === "create") {
    if (!normalizedName.value || nameError.value) return;
    emit("submit", {
      config: cloneOcservGroupConfig(config.value),
      name: normalizedName.value,
    });
    return;
  }
  emit("submit", { config: cloneOcservGroupConfig(config.value) });
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
                  ? "ocservGroups.createTitle"
                  : "ocservGroups.editTitle",
              )
            }}
          </SheetTitle>
          <SheetDescription>
            {{
              t(
                mode === "create"
                  ? "ocservGroups.createDescription"
                  : "ocservGroups.editDescription",
              )
            }}
          </SheetDescription>
        </SheetHeader>

        <div v-if="loading" class="flex flex-col gap-4 px-4" aria-busy="true">
          <Skeleton v-for="index in 5" :key="index" class="h-16 w-full" />
        </div>

        <template v-else>
          <Field
            class="px-4"
            :data-disabled="mode === 'edit' || pending"
            :data-invalid="Boolean(nameError)"
          >
            <FieldLabel for="ocserv-group-name">
              {{ t("ocservGroups.name") }}
            </FieldLabel>
            <Input
              id="ocserv-group-name"
              v-model="name"
              :aria-invalid="Boolean(nameError)"
              :disabled="mode === 'edit' || pending"
              name="name"
              required
            />
            <FieldDescription v-if="nameError" class="text-destructive">
              {{ nameError }}
            </FieldDescription>
          </Field>

          <div class="flex flex-col gap-6 px-4">
            <OcservGroupConfigFields
              :config="config"
              :disabled="pending"
              @change="updateField"
            />
          </div>

          <SheetFooter>
            <Button type="button" variant="outline" @click="open = false">
              {{ t("ocservGroups.cancel") }}
            </Button>
            <Button type="submit" :disabled="pending">
              <Spinner v-if="pending" data-icon="inline-start" />
              {{ pending ? t("ocservGroups.saving") : t("ocservGroups.save") }}
            </Button>
          </SheetFooter>
        </template>
      </form>
    </SheetContent>
  </Sheet>
</template>

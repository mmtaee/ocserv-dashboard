<script setup lang="ts">
import { reactive } from "vue";
import { useI18n } from "vue-i18n";

import type {
  OcservDefaultGroupConfig,
  OcservDefaultGroupUpdate,
} from "@/api/services/ocserv-groups";
import OcservGroupConfigFields from "@/components/ocserv-groups/OcservGroupConfigFields.vue";
import { cloneOcservGroupConfig } from "@/components/ocserv-groups/group-config";
import { Button } from "@/components/ui/button";
import { Spinner } from "@/components/ui/spinner";

const props = defineProps<{
  config: OcservDefaultGroupConfig;
  pending?: boolean;
}>();
const emit = defineEmits<{ submit: [request: OcservDefaultGroupUpdate] }>();
const { t } = useI18n({ useScope: "global" });

const form = reactive<OcservDefaultGroupConfig>(
  cloneOcservGroupConfig(props.config),
);

function updateField(
  key: keyof OcservDefaultGroupConfig,
  value: OcservDefaultGroupConfig[keyof OcservDefaultGroupConfig],
): void {
  Object.assign(form, { [key]: value });
}

function submit(): void {
  emit("submit", {
    config: cloneOcservGroupConfig(form),
  });
}
</script>

<template>
  <form class="flex flex-col gap-6" @submit.prevent="submit">
    <OcservGroupConfigFields
      :config="form"
      :disabled="pending"
      @change="updateField"
    />

    <div class="flex justify-end">
      <Button type="submit" :disabled="pending">
        <Spinner v-if="pending" data-icon="inline-start" />
        {{ pending ? t("groupDefaults.saving") : t("groupDefaults.submit") }}
      </Button>
    </div>
  </form>
</template>

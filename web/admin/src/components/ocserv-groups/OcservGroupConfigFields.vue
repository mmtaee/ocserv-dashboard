<script setup lang="ts">
import { useI18n } from "vue-i18n";

import type { ModelsOcservGroupConfig } from "@/api/generated";
import OcservGroupDefaultsSection from "@/components/ocserv-groups/OcservGroupDefaultsSection.vue";
import {
  accessFields,
  networkFields,
  performanceFields,
  routingFields,
} from "@/components/ocserv-groups/default-group-fields";

defineProps<{
  config: ModelsOcservGroupConfig;
  disabled?: boolean;
}>();
const emit = defineEmits<{
  change: [
    key: keyof ModelsOcservGroupConfig,
    value: ModelsOcservGroupConfig[keyof ModelsOcservGroupConfig],
  ];
}>();
const { t } = useI18n({ useScope: "global" });

function forwardChange(
  key: keyof ModelsOcservGroupConfig,
  value: ModelsOcservGroupConfig[keyof ModelsOcservGroupConfig],
): void {
  emit("change", key, value);
}
</script>

<template>
  <OcservGroupDefaultsSection
    :config="config"
    :description="t('groupDefaults.networkDescription')"
    :disabled="disabled"
    :fields="networkFields"
    :list-help="t('groupDefaults.listHelp')"
    :title="t('groupDefaults.network')"
    @change="forwardChange"
  />
  <OcservGroupDefaultsSection
    :config="config"
    :description="t('groupDefaults.performanceDescription')"
    :disabled="disabled"
    :fields="performanceFields"
    :list-help="t('groupDefaults.listHelp')"
    :title="t('groupDefaults.performance')"
    @change="forwardChange"
  />
  <OcservGroupDefaultsSection
    :config="config"
    :description="t('groupDefaults.accessDescription')"
    :disabled="disabled"
    :fields="accessFields"
    :list-help="t('groupDefaults.listHelp')"
    :title="t('groupDefaults.access')"
    @change="forwardChange"
  />
  <OcservGroupDefaultsSection
    :config="config"
    :description="t('groupDefaults.routingDescription')"
    :disabled="disabled"
    :fields="routingFields"
    :list-help="t('groupDefaults.listHelp')"
    :title="t('groupDefaults.routing')"
    @change="forwardChange"
  />
</template>

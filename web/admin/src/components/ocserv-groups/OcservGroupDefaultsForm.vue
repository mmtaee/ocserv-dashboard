<script setup lang="ts">
import { reactive } from "vue";
import { useI18n } from "vue-i18n";

import type {
  OcservDefaultGroupConfig,
  OcservDefaultGroupUpdate,
} from "@/api/services/ocserv-groups";
import OcservGroupDefaultsSection from "@/components/ocserv-groups/OcservGroupDefaultsSection.vue";
import {
  accessFields,
  networkFields,
  performanceFields,
  routingFields,
} from "@/components/ocserv-groups/default-group-fields";
import { Button } from "@/components/ui/button";
import { Spinner } from "@/components/ui/spinner";

const props = defineProps<{
  config: OcservDefaultGroupConfig;
  pending?: boolean;
}>();
const emit = defineEmits<{ submit: [request: OcservDefaultGroupUpdate] }>();
const { t } = useI18n({ useScope: "global" });

const form = reactive<OcservDefaultGroupConfig>({
  ...props.config,
  dns: props.config.dns ? [...props.config.dns] : undefined,
  "no-route": props.config["no-route"]
    ? [...props.config["no-route"]]
    : undefined,
  route: props.config.route ? [...props.config.route] : undefined,
  "split-dns": props.config["split-dns"]
    ? [...props.config["split-dns"]]
    : undefined,
});

function updateField(
  key: keyof OcservDefaultGroupConfig,
  value: OcservDefaultGroupConfig[keyof OcservDefaultGroupConfig],
): void {
  Object.assign(form, { [key]: value });
}

function submit(): void {
  emit("submit", {
    config: {
      ...form,
      dns: form.dns ? [...form.dns] : undefined,
      "no-route": form["no-route"] ? [...form["no-route"]] : undefined,
      route: form.route ? [...form.route] : undefined,
      "split-dns": form["split-dns"] ? [...form["split-dns"]] : undefined,
    },
  });
}
</script>

<template>
  <form class="flex flex-col gap-6" @submit.prevent="submit">
    <OcservGroupDefaultsSection
      :config="form"
      :description="t('groupDefaults.networkDescription')"
      :disabled="pending"
      :fields="networkFields"
      :list-help="t('groupDefaults.listHelp')"
      :title="t('groupDefaults.network')"
      @change="updateField"
    />
    <OcservGroupDefaultsSection
      :config="form"
      :description="t('groupDefaults.performanceDescription')"
      :disabled="pending"
      :fields="performanceFields"
      :list-help="t('groupDefaults.listHelp')"
      :title="t('groupDefaults.performance')"
      @change="updateField"
    />
    <OcservGroupDefaultsSection
      :config="form"
      :description="t('groupDefaults.accessDescription')"
      :disabled="pending"
      :fields="accessFields"
      :list-help="t('groupDefaults.listHelp')"
      :title="t('groupDefaults.access')"
      @change="updateField"
    />
    <OcservGroupDefaultsSection
      :config="form"
      :description="t('groupDefaults.routingDescription')"
      :disabled="pending"
      :fields="routingFields"
      :list-help="t('groupDefaults.listHelp')"
      :title="t('groupDefaults.routing')"
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

<script setup lang="ts">
import { FileJson } from "@lucide/vue";
import { computed } from "vue";
import { useI18n } from "vue-i18n";

import {
  Empty,
  EmptyDescription,
  EmptyHeader,
  EmptyMedia,
  EmptyTitle,
} from "@/components/ui/empty";

const props = defineProps<{ result: string | null }>();
const { t } = useI18n({ useScope: "global" });

const formattedResult = computed(() => {
  if (props.result == null) return "";
  try {
    const parsed: unknown = JSON.parse(props.result);
    return typeof parsed === "string"
      ? parsed
      : JSON.stringify(parsed, null, 2);
  } catch {
    return props.result;
  }
});
</script>

<template>
  <pre
    v-if="result != null"
    class="max-h-[32rem] overflow-auto rounded-md bg-muted p-4 text-start font-mono text-sm whitespace-pre-wrap"
    dir="ltr"
  >
    {{ formattedResult }}
  </pre>
  <Empty v-else class="border-0 py-12">
    <EmptyHeader>
      <EmptyMedia variant="icon"><FileJson /></EmptyMedia>
      <EmptyTitle>{{ t("occtl.noResult") }}</EmptyTitle>
      <EmptyDescription>{{ t("occtl.noResultDescription") }}</EmptyDescription>
    </EmptyHeader>
  </Empty>
</template>

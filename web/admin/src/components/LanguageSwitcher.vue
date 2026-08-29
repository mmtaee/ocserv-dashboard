<script setup lang="ts">
import { computed } from "vue";
import type { HTMLAttributes } from "vue";
import { Languages } from "@lucide/vue";
import { useI18n } from "vue-i18n";

import {
  Select,
  SelectContent,
  SelectGroup,
  SelectItem,
  SelectLabel,
  SelectTrigger,
  SelectValue,
} from "@/components/ui/select";
import { languageOptions, setLocale } from "@/locales";

const props = defineProps<{ class?: HTMLAttributes["class"] }>();
const { locale, t } = useI18n({ useScope: "global" });
const selectedLanguage = computed({
  get: () => locale.value,
  set: (value: string) => setLocale(value),
});
</script>

<template>
  <div :class="props.class">
    <Select v-model="selectedLanguage">
      <SelectTrigger class="w-32 sm:w-40" :aria-label="t('common.language')">
        <Languages />
        <SelectValue :placeholder="t('common.language')" />
      </SelectTrigger>
      <SelectContent>
        <SelectGroup>
          <SelectLabel>{{ t("common.language") }}</SelectLabel>
          <SelectItem
            v-for="language in languageOptions"
            :key="language.code"
            :value="language.code"
          >
            <span :dir="language.direction">{{ language.label }}</span>
          </SelectItem>
        </SelectGroup>
      </SelectContent>
    </Select>
  </div>
</template>

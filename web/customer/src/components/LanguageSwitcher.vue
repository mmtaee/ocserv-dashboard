<script setup lang="ts">
import { useI18n } from "vue-i18n";
import { localeOptions, setLocale } from "@/locales";

const { locale, t } = useI18n({ useScope: "global" });
</script>

<template>
  <label class="sr-only" for="language">{{ t("language") }}</label>
  <select
    id="language"
    class="input w-auto"
    :value="locale"
    @change="setLocale(($event.target as HTMLSelectElement).value)"
  >
    <option v-for="[code, label] in localeOptions" :key="code" :value="code">
      {{ label }}
    </option>
  </select>
</template>

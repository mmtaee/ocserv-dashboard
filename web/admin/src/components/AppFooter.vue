<script setup lang="ts">
import { Bug } from "@lucide/vue";
import { useI18n } from "vue-i18n";

const { t } = useI18n({ useScope: "global" });
const issuesUrl = "https://github.com/mmtaee/ocserv-dashboard/issues";
</script>

<template>
  <footer
    class="fixed inset-x-0 bottom-0 z-50 border-t bg-background/95 px-3 py-1.5 text-xs text-muted-foreground backdrop-blur supports-[backdrop-filter]:bg-background/80"
  >
    <div
      class="flex w-full flex-wrap items-center justify-between gap-x-3 gap-y-1 text-start"
    >
      <p
        class="inline-flex items-center gap-1 whitespace-nowrap text-foreground"
      >
        <span>{{ t("footer.developedBy", { developer: "Masoud Taee" }) }}</span>
        <span>{{ t("footer.with") }}</span>
        <span
          class="inline-block animate-pulse text-destructive motion-reduce:animate-none"
          role="img"
          :aria-label="t('footer.love')"
          >&hearts;</span
        >
      </p>
      <a
        :href="issuesUrl"
        class="ms-auto inline-flex items-center gap-1.5 whitespace-nowrap font-medium text-foreground underline-offset-4 hover:underline focus-visible:outline-none focus-visible:ring-2 focus-visible:ring-ring focus-visible:ring-offset-2"
      >
        <Bug class="size-3.5" aria-hidden="true" />
        {{ t("footer.reportIssue") }}
      </a>
    </div>
  </footer>
</template>

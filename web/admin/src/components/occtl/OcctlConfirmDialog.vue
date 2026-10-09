<script setup lang="ts">
import { useI18n } from "vue-i18n";

import type { OcctlCommandRequest } from "@/api/services/occtl";
import { Button } from "@/components/ui/button";
import {
  AlertDialog,
  AlertDialogCancel,
  AlertDialogContent,
  AlertDialogDescription,
  AlertDialogFooter,
  AlertDialogHeader,
  AlertDialogTitle,
} from "@/components/ui/alert-dialog";
import { Spinner } from "@/components/ui/spinner";

defineProps<{
  commandLabel: string;
  pending?: boolean;
  request: OcctlCommandRequest | null;
}>();
const emit = defineEmits<{ confirm: [] }>();
const open = defineModel<boolean>("open", { default: false });
const { t } = useI18n({ useScope: "global" });
</script>

<template>
  <AlertDialog v-model:open="open">
    <AlertDialogContent>
      <AlertDialogHeader>
        <AlertDialogTitle>{{ t("occtl.confirmTitle") }}</AlertDialogTitle>
        <AlertDialogDescription>
          {{
            t("occtl.confirmDescription", {
              command: commandLabel,
              value: request?.value || t("occtl.noValue"),
            })
          }}
        </AlertDialogDescription>
      </AlertDialogHeader>
      <AlertDialogFooter>
        <AlertDialogCancel :disabled="pending">
          {{ t("occtl.cancel") }}
        </AlertDialogCancel>
        <Button
          type="button"
          variant="destructive"
          :disabled="pending"
          @click="emit('confirm')"
        >
          <Spinner v-if="pending" data-icon="inline-start" />
          {{ t("occtl.confirm") }}
        </Button>
      </AlertDialogFooter>
    </AlertDialogContent>
  </AlertDialog>
</template>

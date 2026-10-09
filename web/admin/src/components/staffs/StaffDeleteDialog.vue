<script setup lang="ts">
import { useI18n } from "vue-i18n";

import type { Staff } from "@/api/services/staffs";
import {
  AlertDialog,
  AlertDialogAction,
  AlertDialogCancel,
  AlertDialogContent,
  AlertDialogDescription,
  AlertDialogFooter,
  AlertDialogHeader,
  AlertDialogTitle,
} from "@/components/ui/alert-dialog";
import { Spinner } from "@/components/ui/spinner";

defineProps<{ pending?: boolean; staff?: Staff | null }>();
const emit = defineEmits<{ confirm: [] }>();
const open = defineModel<boolean>("open", { default: false });
const { t } = useI18n({ useScope: "global" });
</script>

<template>
  <AlertDialog v-model:open="open">
    <AlertDialogContent>
      <AlertDialogHeader>
        <AlertDialogTitle>{{ t("staffs.deleteTitle") }}</AlertDialogTitle>
        <AlertDialogDescription>
          {{
            t("staffs.deleteDescription", { username: staff?.username ?? "" })
          }}
        </AlertDialogDescription>
      </AlertDialogHeader>
      <AlertDialogFooter>
        <AlertDialogCancel :disabled="pending">
          {{ t("staffs.cancel") }}
        </AlertDialogCancel>
        <AlertDialogAction :disabled="pending" @click.prevent="emit('confirm')">
          <Spinner v-if="pending" data-icon="inline-start" />
          {{ t("staffs.delete") }}
        </AlertDialogAction>
      </AlertDialogFooter>
    </AlertDialogContent>
  </AlertDialog>
</template>

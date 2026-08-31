<script setup lang="ts">
import { useI18n } from "vue-i18n";

import type { OcservGroup } from "@/api/services/ocserv-groups";
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

defineProps<{
  group?: OcservGroup | null;
  pending?: boolean;
}>();
const emit = defineEmits<{ confirm: [] }>();
const open = defineModel<boolean>("open", { default: false });
const { t } = useI18n({ useScope: "global" });
</script>

<template>
  <AlertDialog v-model:open="open">
    <AlertDialogContent>
      <AlertDialogHeader>
        <AlertDialogTitle>
          {{ t("ocservGroups.deleteTitle") }}
        </AlertDialogTitle>
        <AlertDialogDescription>
          {{ t("ocservGroups.deleteDescription", { name: group?.name ?? "" }) }}
        </AlertDialogDescription>
      </AlertDialogHeader>
      <AlertDialogFooter>
        <AlertDialogCancel :disabled="pending">
          {{ t("ocservGroups.cancel") }}
        </AlertDialogCancel>
        <AlertDialogAction :disabled="pending" @click.prevent="emit('confirm')">
          <Spinner v-if="pending" data-icon="inline-start" />
          {{ t("ocservGroups.delete") }}
        </AlertDialogAction>
      </AlertDialogFooter>
    </AlertDialogContent>
  </AlertDialog>
</template>

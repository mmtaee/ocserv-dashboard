<script setup lang="ts">
import { useI18n } from "vue-i18n";

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
  action: string;
  description: string;
  destructive?: boolean;
  pending?: boolean;
  title: string;
}>();
const emit = defineEmits<{ confirm: [] }>();
const open = defineModel<boolean>("open", { default: false });
const { t } = useI18n({ useScope: "global" });
</script>

<template>
  <AlertDialog v-model:open="open">
    <AlertDialogContent>
      <AlertDialogHeader>
        <AlertDialogTitle>{{ title }}</AlertDialogTitle>
        <AlertDialogDescription>{{ description }}</AlertDialogDescription>
      </AlertDialogHeader>
      <AlertDialogFooter>
        <AlertDialogCancel :disabled="pending">
          {{ t("ocservUsers.cancel") }}
        </AlertDialogCancel>
        <Button
          type="button"
          :variant="destructive ? 'destructive' : 'default'"
          :disabled="pending"
          @click="emit('confirm')"
        >
          <Spinner v-if="pending" data-icon="inline-start" />
          {{ action }}
        </Button>
      </AlertDialogFooter>
    </AlertDialogContent>
  </AlertDialog>
</template>

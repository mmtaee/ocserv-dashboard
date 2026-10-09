<script setup lang="ts">
import { CirclePower, Power, RotateCw } from "@lucide/vue";
import { computed, shallowRef } from "vue";
import { useI18n } from "vue-i18n";

import type { OcservRuntimeAction } from "@/api/services/system";
import { Button } from "@/components/ui/button";
import {
  Card,
  CardContent,
  CardDescription,
  CardHeader,
  CardTitle,
} from "@/components/ui/card";
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
import { Skeleton } from "@/components/ui/skeleton";

const props = defineProps<{
  allowedActions: readonly OcservRuntimeAction[];
  loading: boolean;
  pending: OcservRuntimeAction | null;
}>();
const emit = defineEmits<{
  execute: [action: OcservRuntimeAction];
  unavailable: [];
}>();
const { t } = useI18n({ useScope: "global" });
const confirmationOpen = shallowRef(false);
const selectedAction = shallowRef<OcservRuntimeAction | null>(null);
const actionLabel = computed(() =>
  selectedAction.value ? t(`runtime.actions.${selectedAction.value}`) : "",
);

const actionPresentation = {
  enable: { icon: CirclePower, variant: "default" },
  restart: { icon: RotateCw, variant: "outline" },
  disable: { icon: Power, variant: "destructive" },
} as const;
const actions = computed(() =>
  props.allowedActions.map((name) => ({
    name,
    ...actionPresentation[name],
  })),
);

function request(action: OcservRuntimeAction): void {
  selectedAction.value = action;
  confirmationOpen.value = true;
}

function confirm(): void {
  if (!selectedAction.value) return;
  if (!props.allowedActions.includes(selectedAction.value)) {
    confirmationOpen.value = false;
    emit("unavailable");
    return;
  }
  emit("execute", selectedAction.value);
  confirmationOpen.value = false;
}
</script>

<template>
  <Card>
    <CardHeader>
      <CardTitle>{{ t("runtime.actionsTitle") }}</CardTitle>
      <CardDescription>{{ t("runtime.actionsDescription") }}</CardDescription>
    </CardHeader>
    <CardContent v-if="loading" class="flex flex-wrap gap-3">
      <Skeleton v-for="item in 3" :key="item" class="h-9 w-28" />
    </CardContent>
    <CardContent v-else-if="actions.length" class="flex flex-wrap gap-3">
      <Button
        v-for="action in actions"
        :key="action.name"
        type="button"
        :variant="action.variant"
        :disabled="pending !== null"
        @click="request(action.name)"
      >
        <Spinner v-if="pending === action.name" data-icon="inline-start" />
        <component :is="action.icon" v-else data-icon="inline-start" />
        {{ t(`runtime.actions.${action.name}`) }}
      </Button>
    </CardContent>
    <CardContent v-else class="text-sm text-muted-foreground">
      {{ t("runtime.noActionsDescription") }}
    </CardContent>
  </Card>

  <AlertDialog v-model:open="confirmationOpen">
    <AlertDialogContent>
      <AlertDialogHeader>
        <AlertDialogTitle>{{ t("runtime.confirmTitle") }}</AlertDialogTitle>
        <AlertDialogDescription>
          {{ t("runtime.confirmDescription", { action: actionLabel }) }}
        </AlertDialogDescription>
      </AlertDialogHeader>
      <AlertDialogFooter>
        <AlertDialogCancel :disabled="pending !== null">
          {{ t("runtime.cancel") }}
        </AlertDialogCancel>
        <Button
          type="button"
          :variant="selectedAction === 'disable' ? 'destructive' : 'default'"
          :disabled="pending !== null || selectedAction === null"
          @click="confirm"
        >
          <Spinner v-if="pending !== null" data-icon="inline-start" />
          {{ t("runtime.confirm") }}
        </Button>
      </AlertDialogFooter>
    </AlertDialogContent>
  </AlertDialog>
</template>

<script setup lang="ts">
import { CircleAlert, CircleCheck, Info, TriangleAlert } from "@lucide/vue";
import { computed } from "vue";
import type { HTMLAttributes } from "vue";
import type { AlertVariants } from ".";
import { cn } from "@/lib/utils";
import { alertVariants } from ".";

const props = defineProps<{
  class?: HTMLAttributes["class"];
  variant?: AlertVariants["variant"];
}>();

const icon = computed(
  () =>
    ({
      info: Info,
      success: CircleCheck,
      warning: TriangleAlert,
      error: CircleAlert,
    })[props.variant ?? "info"],
);
</script>

<template>
  <div
    data-slot="alert"
    :class="cn(alertVariants({ variant }), props.class)"
    role="alert"
  >
    <slot name="icon"><component :is="icon" /></slot>
    <slot />
  </div>
</template>

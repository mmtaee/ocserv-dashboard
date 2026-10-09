<script setup lang="ts">
import type { PaginationFirstProps } from "reka-ui";
import type { HTMLAttributes } from "vue";
import type { ButtonVariants } from "@/components/ui/button";
import { ChevronLeftIcon } from "@lucide/vue";
import { reactiveOmit } from "@vueuse/core";
import { PaginationFirst, useForwardProps } from "reka-ui";
import { cn } from "@/lib/utils";
import { buttonVariants } from "@/components/ui/button";

const props = withDefaults(
  defineProps<
    PaginationFirstProps & {
      size?: ButtonVariants["size"];
      class?: HTMLAttributes["class"];
      label?: string;
    }
  >(),
  {
    label: "First",
    size: "default",
  },
);

const delegatedProps = reactiveOmit(props, "class", "label", "size");
const forwarded = useForwardProps(delegatedProps);
</script>

<template>
  <PaginationFirst
    :aria-label="label"
    data-slot="pagination-first"
    :class="
      cn(
        buttonVariants({ variant: 'ghost', size }),
        'gap-1 px-2.5 sm:pe-2.5',
        props.class,
      )
    "
    v-bind="forwarded"
  >
    <slot>
      <ChevronLeftIcon class="rtl:rotate-180" />
      <span class="hidden sm:block">{{ label }}</span>
    </slot>
  </PaginationFirst>
</template>

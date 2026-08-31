<script setup lang="ts">
import type { PaginationPrevProps } from "reka-ui";
import type { HTMLAttributes } from "vue";
import type { ButtonVariants } from "@/components/ui/button";
import { ChevronLeftIcon } from "@lucide/vue";
import { reactiveOmit } from "@vueuse/core";
import { PaginationPrev, useForwardProps } from "reka-ui";
import { cn } from "@/lib/utils";
import { buttonVariants } from "@/components/ui/button";

const props = withDefaults(
  defineProps<
    PaginationPrevProps & {
      size?: ButtonVariants["size"];
      class?: HTMLAttributes["class"];
      label?: string;
    }
  >(),
  {
    label: "Previous",
    size: "default",
  },
);

const delegatedProps = reactiveOmit(props, "class", "label", "size");
const forwarded = useForwardProps(delegatedProps);
</script>

<template>
  <PaginationPrev
    :aria-label="label"
    data-slot="pagination-previous"
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
  </PaginationPrev>
</template>

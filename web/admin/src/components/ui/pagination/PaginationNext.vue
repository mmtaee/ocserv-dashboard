<script setup lang="ts">
import type { PaginationNextProps } from "reka-ui";
import type { HTMLAttributes } from "vue";
import type { ButtonVariants } from "@/components/ui/button";
import { ChevronRightIcon } from "@lucide/vue";
import { reactiveOmit } from "@vueuse/core";
import { PaginationNext, useForwardProps } from "reka-ui";
import { cn } from "@/lib/utils";
import { buttonVariants } from "@/components/ui/button";

const props = withDefaults(
  defineProps<
    PaginationNextProps & {
      size?: ButtonVariants["size"];
      class?: HTMLAttributes["class"];
      label?: string;
    }
  >(),
  {
    label: "Next",
    size: "default",
  },
);

const delegatedProps = reactiveOmit(props, "class", "label", "size");
const forwarded = useForwardProps(delegatedProps);
</script>

<template>
  <PaginationNext
    :aria-label="label"
    data-slot="pagination-next"
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
      <span class="hidden sm:block">{{ label }}</span>
      <ChevronRightIcon class="rtl:rotate-180" />
    </slot>
  </PaginationNext>
</template>

<script setup lang="ts">
import type { PaginationLastProps } from "reka-ui";
import type { HTMLAttributes } from "vue";
import type { ButtonVariants } from "@/components/ui/button";
import { ChevronRightIcon } from "@lucide/vue";
import { reactiveOmit } from "@vueuse/core";
import { PaginationLast, useForwardProps } from "reka-ui";
import { cn } from "@/lib/utils";
import { buttonVariants } from "@/components/ui/button";

const props = withDefaults(
  defineProps<
    PaginationLastProps & {
      size?: ButtonVariants["size"];
      class?: HTMLAttributes["class"];
      label?: string;
    }
  >(),
  {
    label: "Last",
    size: "default",
  },
);

const delegatedProps = reactiveOmit(props, "class", "label", "size");
const forwarded = useForwardProps(delegatedProps);
</script>

<template>
  <PaginationLast
    :aria-label="label"
    data-slot="pagination-last"
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
  </PaginationLast>
</template>

<script setup lang="ts">
import { computed, useId } from "vue";
import { useI18n } from "vue-i18n";

const props = withDefaults(
  defineProps<{
    label: string;
    percent?: number;
    used?: number;
    total?: number;
    unit: string;
    variant?: "ring" | "liquid";
  }>(),
  {
    percent: 0,
    used: undefined,
    total: undefined,
    variant: "liquid",
  },
);

const { locale } = useI18n({ useScope: "global" });
const clipId = `usage-gauge-${useId().replaceAll(":", "")}`;
const circumference = 2 * Math.PI * 44;
const normalizedPercent = computed(() =>
  Math.min(
    100,
    Math.max(0, Number.isFinite(props.percent) ? props.percent : 0),
  ),
);
const ringOffset = computed(
  () => circumference * (1 - normalizedPercent.value / 100),
);
const fillY = computed(() => 100 - normalizedPercent.value);
const wavePaths = computed(() => {
  const y = fillY.value;
  const amplitude = Math.min(5, y, 100 - y);

  return {
    back: `M 0 ${y} C 18 ${y + amplitude} 32 ${y - amplitude} 50 ${y} C 68 ${y + amplitude} 82 ${y - amplitude} 100 ${y} V 100 H 0 Z`,
    front: `M 0 ${y} C 16 ${y - amplitude} 34 ${y + amplitude} 50 ${y} C 66 ${y - amplitude} 84 ${y + amplitude} 100 ${y} V 100 H 0 Z`,
  };
});
const percentLabel = computed(() =>
  new Intl.NumberFormat(locale.value, { maximumFractionDigits: 2 }).format(
    normalizedPercent.value,
  ),
);
const numberFormatter = computed(
  () =>
    new Intl.NumberFormat(locale.value, {
      maximumFractionDigits: 2,
    }),
);
const detail = computed(() => {
  if (props.used === undefined && props.total === undefined) return "—";
  const used = numberFormatter.value.format(props.used ?? 0);
  const total = numberFormatter.value.format(props.total ?? 0);
  return `${used} / ${total} ${props.unit}`;
});
</script>

<template>
  <div class="flex min-w-0 flex-col items-center gap-2 text-center">
    <div
      class="relative size-28"
      role="img"
      :aria-label="`${label}: ${percentLabel}%`"
    >
      <svg
        v-if="variant === 'ring'"
        viewBox="0 0 100 100"
        class="size-full -rotate-90"
        aria-hidden="true"
      >
        <circle
          cx="50"
          cy="50"
          r="44"
          class="fill-none stroke-muted"
          stroke-width="10"
        />
        <circle
          cx="50"
          cy="50"
          r="44"
          class="fill-none stroke-primary transition-[stroke-dashoffset] duration-500"
          stroke-linecap="round"
          stroke-width="10"
          :stroke-dasharray="circumference"
          :stroke-dashoffset="ringOffset"
        />
      </svg>
      <svg
        v-else-if="variant === 'liquid'"
        viewBox="0 0 100 100"
        class="size-full"
        aria-hidden="true"
      >
        <defs>
          <clipPath :id="clipId">
            <circle cx="50" cy="50" r="44" />
          </clipPath>
        </defs>
        <circle
          cx="50"
          cy="50"
          r="44"
          class="fill-muted stroke-border"
          stroke-width="2"
        />
        <g :clip-path="`url(#${clipId})`">
          <path
            :d="wavePaths.back"
            class="fill-primary/45 transition-all duration-500"
          />
          <path
            :d="wavePaths.front"
            class="fill-primary transition-all duration-500"
          />
        </g>
        <circle
          cx="50"
          cy="50"
          r="44"
          class="fill-none stroke-primary/60"
          stroke-width="1"
        />
      </svg>
      <span
        class="absolute inset-0 flex items-center justify-center text-lg font-semibold tabular-nums"
      >
        {{ percentLabel }}%
      </span>
    </div>
    <div class="min-w-0">
      <p class="font-medium">{{ label }}</p>
      <p class="truncate text-sm text-muted-foreground" dir="ltr">
        {{ detail }}
      </p>
    </div>
  </div>
</template>

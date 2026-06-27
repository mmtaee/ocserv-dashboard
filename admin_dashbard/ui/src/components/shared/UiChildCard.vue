<script lang="ts" setup>
import type { PropType } from 'vue';

defineProps({
    title: String,
    divider: Boolean,
    titleColor: String,
    variant: {
        type: String as PropType<'flat' | 'text' | 'elevated' | 'tonal' | 'outlined' | 'plain'>,
        default: 'flat'
    },
    width: {
        type: [Number, String] as PropType<number | 'auto'>,
        default: 'auto'
    },
    height: {
        type: [Number, String] as PropType<number | 'auto'>,
        default: 'auto'
    }
});
</script>

<template>
    <v-card elevation="10" :variant="variant" :width="width" :height="height">
        <v-card-item v-if="title || $slots['title-header']" :class="`bg-${titleColor}`" class="py-3">
            <v-card-title class="text-h5 text-capitalize" v-if="title">
                {{ title }}
            </v-card-title>
            <slot name="title-header" />
        </v-card-item>
        <v-divider v-if="divider" />
        <v-card-text>
            <slot />
        </v-card-text>
        <v-card-actions>
            <slot name="action" />
        </v-card-actions>
    </v-card>
</template>

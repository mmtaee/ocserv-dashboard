<script lang="ts" setup>
import type { PropType } from 'vue';

defineProps({
    title: String,
    titleColor: {
        type: String,
        default: 'grey'
    },
    spacing: {
        type: Boolean,
        default: true
    },
    variant: {
        type: String as PropType<'flat' | 'text' | 'elevated' | 'tonal' | 'outlined' | 'plain'>,
        default: 'flat'
    },
    minHeight: {
        type: [Number, String] as PropType<number | 'auto'>,
        default: 'auto'
    }
});
</script>

<template>
    <v-card elevation="10" :variant="variant" :min-height="minHeight">
        <v-card-item class="pa-0">
            <div :class="spacing ? 'd-sm-flex align-center justify-space-between' : ''">
                <h5 v-if="title" :class="`text-${titleColor}`" class="text-h3 mb-6 pl-7 pt-7 text-capitalize">
                    <slot name="header-prepend"></slot>
                    {{ title }}
                    <slot name="header-append"></slot>
                </h5>
                <slot name="action"></slot>
            </div>
            <slot />
        </v-card-item>
    </v-card>
</template>

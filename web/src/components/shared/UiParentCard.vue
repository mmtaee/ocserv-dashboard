<script lang="ts" setup>
defineProps({
    title: String,
    titleColor: {
        type: String,
        default: 'grey'
    },
    spacing: {
        type: Boolean,
        default: true
    }
});
</script>

<template>
    <v-card elevation="10">
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

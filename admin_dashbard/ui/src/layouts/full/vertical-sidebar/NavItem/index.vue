<script setup>
defineProps({ item: Object, level: Number });
</script>

<template>
    <!---Single Item-->
    <v-list-item :disabled="item.disabled" :to="item.to" class="mb-1 pt-md-3" color="primary" rounded>
        <v-list-item-title class="text-capitalize px-5">
            <v-icon class="mb-2" start>{{ item.icon }}</v-icon>
            {{ item.title }}
        </v-list-item-title>
    </v-list-item>
</template>

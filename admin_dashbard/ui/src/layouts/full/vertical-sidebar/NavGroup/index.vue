<script setup>
defineProps({ item: Object });
</script>

<template>
    <v-list-subheader class="smallCap text-uppercase text-subtitle-2 mt-5 font-weight-bold" color="info">
        {{ item.header }}
    </v-list-subheader>
</template>

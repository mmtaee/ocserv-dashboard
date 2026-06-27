<script lang="ts" setup>
import { RouterView } from 'vue-router';
import Loading from '@/components/shared/Loading.vue';
import Snackbar from '@/components/shared/Snackbar.vue';
</script>

<template>
    <RouterView />

    <Snackbar />
    <Loading />
</template>

<script setup lang="ts">
defineProps<{ error?: string; success?: string }>();
</script>

<template>
  <p v-if="error" class="alert">{{ error }}</p>
  <p v-else-if="success" class="rounded-lg border bg-muted p-3 text-sm">
    {{ success }}
  </p>
</template>

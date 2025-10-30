<script lang="ts" setup>
import { ref, onMounted, watch } from 'vue';
import { useLoadingStore } from '@/stores/loading';
import logoUrl from '@/assets/images/logo-square.png';

const loader = useLoadingStore();

const loadingText = 'Loading...';
const letters = ref<string[]>(loadingText.split(''));
const activeIndex = ref(0);

onMounted(() => {
    const interval = setInterval(() => {
        activeIndex.value = (activeIndex.value + 1) % letters.value.length;
    }, 300);

    watch(
        () => loader.isLoading,
        (val) => {
            if (!val) clearInterval(interval);
        }
    );
});
</script>

<template>
    <v-overlay v-model="loader.isLoading" class="d-flex align-center justify-center" persistent>
        <div class="loader">
            <img :src="logoUrl" alt="Logo" class="logo" />
            <p class="loading-text">
                <span v-for="(letter, index) in letters" :key="index" :class="{ active: index === activeIndex }">{{
                    letter
                }}</span>
            </p>
        </div>
    </v-overlay>
</template>

<style scoped>
.loader {
    display: flex;
    flex-direction: column;
    align-items: center;
    gap: 10px;
}

.logo {
    width: 120px;
}

.loading-text {
    color: black;
    font-size: 0.9rem;
    font-weight: 500;
    display: inline-flex;
    gap: 2px;
}

.loading-text span {
    display: inline-block;
    transition:
        transform 0.2s ease,
        color 0.2s ease;
    opacity: 0.5;
}

.loading-text span.active {
    transform: translateY(-6px);
    opacity: 1;
    color: black;
}
</style>

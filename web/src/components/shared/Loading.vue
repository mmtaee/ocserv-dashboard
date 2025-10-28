<script lang="ts" setup>
import { useLoadingStore } from '@/stores/loading';
import logoUrl from '@/assets/images/logo-circule.png';
import { useI18n } from 'vue-i18n';

const { t } = useI18n();

const loader = useLoadingStore();
</script>

<template>
    <v-overlay v-model="loader.isLoading" class="d-flex align-center justify-center" persistent>
        <div class="loader">
            <img :src="logoUrl" alt="Logo" class="logo" />
            <p class="loading-text typing text-capitalize"></p>
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
    color: #fff;
    font-size: 0.9rem;
    font-weight: 500;
    letter-spacing: 1px;
    animation: wobble 1s infinite;
}

.loading-text.typing::after {
    content: '';
    animation: typing 2s steps(14, end) infinite;
}

@keyframes typing {
    0% {
        content: 'R';
    }
    7% {
        content: 'Re';
    }
    14% {
        content: 'Req';
    }
    21% {
        content: 'Requ';
    }
    28% {
        content: 'Reque';
    }
    35% {
        content: 'Reques';
    }
    42% {
        content: 'Request';
    }
    50% {
        content: 'Requesti';
    }
    57% {
        content: 'Requestin';
    }
    64% {
        content: 'Requesting';
    }
    71% {
        content: 'Requesting.';
    }
    79% {
        content: 'Requesting..';
    }
    86% {
        content: 'Requesting...';
    }
    93% {
        content: 'Requesting...';
    }
    100% {
        content: 'Requesting...';
    }
}
</style>

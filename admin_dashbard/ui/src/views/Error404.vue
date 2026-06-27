<script lang="ts" setup>
import { useI18n } from 'vue-i18n';

const { t } = useI18n();
</script>
<template>
    <div class="not-found">
        <svg width="800" height="600" viewBox="0 0 800 600" xmlns="http://www.w3.org/2000/svg">
            <image href="@/assets/images/404.jpg" x="0" y="0" width="800" height="600" />
        </svg>
        <div class="content">
            <p class="text-capitalize">{{ t('PAGE_404_TEXT') }}.</p>
            <v-btn class="mb-4" color="primary" flat to="/">{{ t('BACK_TO_HOME') }}</v-btn>
        </div>
    </div>
</template>

<style scoped>
.not-found {
    display: flex;
    flex-direction: column;
    justify-content: center;
    align-items: center;
    height: 100vh;
    text-align: center;
}

.not-found svg {
    display: block;
}

.content p {
    font-size: 18px;
    margin: 8px 0 16px;
    color: #888;
}
</style>

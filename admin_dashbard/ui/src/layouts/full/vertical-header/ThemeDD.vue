<script lang="ts" setup>
import { computed } from 'vue';
import { useI18n } from 'vue-i18n';
import { useThemeStore } from '@/stores/theme';

const { t } = useI18n();
const themeStore = useThemeStore();

const icon = computed(() => (themeStore.isDark ? 'mdi-weather-sunny' : 'mdi-weather-night'));
const label = computed(() => (themeStore.isDark ? t('THEME_LIGHT') : t('THEME_DARK')));

function toggle() {
    themeStore.toggle();
}
</script>

<template>
    <v-tooltip :text="label" location="bottom">
        <template #activator="{ props }">
            <v-btn
                v-bind="props"
                :aria-label="label"
                class="theme-toggle-btn"
                icon
                size="small"
                variant="text"
                @click="toggle"
            >
                <v-icon size="22">{{ icon }}</v-icon>
            </v-btn>
        </template>
    </v-tooltip>
</template>

<style lang="scss" scoped>
.theme-toggle-btn {
    margin-inline: 4px;
}
</style>

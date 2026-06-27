<script lang="ts" setup>
import { computed } from 'vue';
import { useI18n } from 'vue-i18n';

const props = defineProps<{
    enabled?: boolean;
    hasBotToken?: boolean;
    botUsername?: string;
}>();

const { t } = useI18n();

const statusColor = computed(() => {
    if (!props.enabled) return 'grey';
    if (!props.hasBotToken) return 'warning';
    return 'success';
});

const statusLabel = computed(() => (props.enabled ? t('TELEGRAM_STATUS_ENABLED') : t('TELEGRAM_STATUS_DISABLED')));

const tokenLabel = computed(() => (props.hasBotToken ? t('TELEGRAM_BOT_TOKEN_SET') : t('TELEGRAM_BOT_TOKEN_MISSING')));
</script>

<template>
    <v-card elevation="10">
        <v-card-item class="text-capitalize">
            <div class="d-flex flex-wrap align-center justify-space-between gap-2">
                <v-card-title class="text-h5 text-capitalize">
                    {{ t('TELEGRAM_DASHBOARD_STATUS') }}
                </v-card-title>
                <div class="d-flex flex-wrap align-center gap-2">
                    <v-chip :color="statusColor" size="small" variant="tonal">
                        {{ statusLabel }}
                    </v-chip>
                    <v-chip color="primary" size="small" variant="outlined">
                        {{ tokenLabel }}
                    </v-chip>
                    <v-chip v-if="botUsername" color="info" size="small" variant="tonal">
                        {{ t('TELEGRAM_BOT_AT') }}: @{{ botUsername }}
                    </v-chip>
                </div>
            </div>
        </v-card-item>
    </v-card>
</template>

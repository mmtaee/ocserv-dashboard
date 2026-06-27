<script lang="ts" setup>
import { useI18n } from 'vue-i18n';
import type { RepositoryTopBandwidthUsers } from '@/api';
import { formatDateTime } from '@/utils/convertors';

defineProps<{
    topUsers: RepositoryTopBandwidthUsers;
}>();
const { t } = useI18n();

const convertToGB = (num: number): number => {
    if (!num) return 0;
    return num / 1024 ** 3;
};
</script>
<template>
    <v-card class="pa-5" elevation="10">
        <div class="mb-5 text-h5 text-capitalize">{{ t('TOP_BANDWIDTH_USERS_OVERVIEW') }}</div>

        <v-row>
            <!-- top_rx -->
            <v-col v-if="(topUsers?.top_rx ?? []).length > 0" class="text-subtitle-1 text-capitalize" cols="12">
                {{ t('TOP_BANDWIDTH_USERS_OVERVIEW_RX') }}
            </v-col>
            <v-col
                v-if="(topUsers?.top_rx ?? []).length > 0"
                v-for="(userRx, index) in topUsers.top_rx"
                :key="`${index}-top-rx`"
                cols="12"
                lg="3"
                sm="6"
            >
                <v-card elevation="4" rounded="md">
                    <v-card-item class="pt-0">
                        <h6 class="text-h6 mt-5 mb-2">{{ userRx.username }}</h6>
                        <div class="text-h6 text-primary">{{ t('RX') }}: {{ convertToGB(userRx.rx).toFixed(8) }} GB</div>
                        <div class="text-h6 my-2 text-secondary">
                            {{ t('TX') }}: {{ convertToGB(userRx.tx).toFixed(8) }} GB
                        </div>
                        <div class="text-body-2 text-medium-emphasis text-capitalize">
                            {{ t('CREATED_AT') }}:
                            {{ formatDateTime(userRx.created_at, '') }}
                        </div>
                    </v-card-item>
                </v-card>
            </v-col>
            <v-col v-else class="text-subtitle-1 text-capitalize" cols="12" md="12">
                {{ t('TOP_BANDWIDTH_USERS_OVERVIEW_RX_NOT_FOUND') }}
            </v-col>

            <!-- top_tx -->
            <v-col v-if="(topUsers?.top_tx ?? []).length > 0" class="text-subtitle-1 text-capitalize" cols="12">
                {{ t('TOP_BANDWIDTH_USERS_OVERVIEW_TX') }}
            </v-col>
            <v-col
                v-if="(topUsers?.top_tx ?? []).length > 0"
                v-for="(userTx, index) in topUsers.top_tx"
                :key="`${index}-top-tx`"
                cols="12"
                lg="3"
                sm="6"
            >
                <v-card elevation="4" rounded="md">
                    <v-card-item class="pt-0">
                        <h6 class="text-h6 mt-5 mb-2">{{ userTx.username }}</h6>
                        <div class="text-h6 text-primary">{{ t('TX') }}: {{ convertToGB(userTx.tx).toFixed(8) }} GB</div>
                        <div class="text-h6 my-2 text-secondary">
                            {{ t('RX') }}: {{ convertToGB(userTx.rx).toFixed(8) }} GB
                        </div>
                        <div class="text-body-2 text-medium-emphasis text-capitalize">
                            {{ t('CREATED_AT') }}:
                            {{ formatDateTime(userTx.created_at, '') }}
                        </div>
                    </v-card-item>
                </v-card>
            </v-col>
            <v-col v-else class="text-subtitle-1 text-capitalize" cols="12" md="12">
                {{ t('TOP_BANDWIDTH_USERS_OVERVIEW_TX_NOT_FOUND') }}
            </v-col>
        </v-row>
    </v-card>
</template>

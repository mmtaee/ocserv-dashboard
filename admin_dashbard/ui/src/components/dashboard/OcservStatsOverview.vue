<script lang="ts" setup>
import { useI18n } from 'vue-i18n';
import { HomeApi, type HomeCurrentStats, type HomeGeneralInfo, type HomeOcservStatusResponse } from '@/api';
import { onMounted, onUnmounted, ref } from 'vue';
import { getAuthorization } from '@/utils/request';
import { useServerStore } from '@/stores/config';

const { t } = useI18n();

const stats = ref<HomeOcservStatusResponse>({
    current_stats: {},
    general_info: {}
});

const api = new HomeApi();

const getOcservStats = () => {
    api.homeOcservStatsGet(
        {
            ...getAuthorization()
        },
        {
            headers: {
                'x-skip-loading': 'true'
            }
        }
    ).then((res) => {
        Object.assign(stats.value, res.data);
        const serverStore = useServerStore();
        serverStore.setStatus(res.data.general_info?.Status || 'active');
    });
};

const keyFormatter = (key: string) => {
    return key.replace(/_/g, ' ').trim();
};

onMounted(() => {
    getOcservStats();
});
</script>

<template>
    <v-row align="center" justify="center">
        <!-- SERVER_GENERAL_INFO_OVERVIEW -->
        <v-col cols="12" lg="6" sm="12">
            <v-card elevation="10" height="595px">
                <v-card-title class="text-h5 pt-sm-2 ms-3 text-capitalize">{{
                    t('SERVER_GENERAL_INFO_OVERVIEW')
                }}</v-card-title>
                <v-card-item class="pa-6">
                    <tbody>
                        <tr v-for="(val, key, index) in stats.general_info" :key="`current-stats-${index}`">
                            <td class="px-10">
                                <p class="text-15 font-weight-bold text-capitalize">
                                    {{ keyFormatter(key) }}
                                </p>
                            </td>
                            <td class="px-10">
                                <p v-if="key == 'Status'" class="text-15 font-weight-medium">
                                    <span
                                        :class="val === 'online' ? 'text-primary' : 'text-error'"
                                        class="text-capitalize font-weight-bold text-h5"
                                    >
                                        {{ val }}
                                    </span>
                                    <v-icon v-if="val === 'online'" color="primary" end>mdi-check</v-icon>
                                    <v-icon v-else color="error" end>mdi-alert-circle</v-icon>
                                </p>
                                <p v-else class="text-15 font-weight-medium">
                                    {{ val }}
                                </p>
                            </td>
                        </tr>
                    </tbody>
                </v-card-item>
            </v-card>
        </v-col>

        <!-- SERVER_CURRENT_STATS -->
        <v-col cols="12" lg="6" sm="12">
            <v-card elevation="10" height="595px">
                <v-card-title class="text-h5 pt-sm-2 ms-3 text-capitalize">{{ t('SERVER_CURRENT_STATS') }}</v-card-title>
                <v-card-item class="pa-6">
                    <tbody>
                        <tr v-for="(val, key, index) in stats.current_stats" :key="`current-stats-${index}`">
                            <td class="px-10">
                                <p class="text-15 font-weight-bold text-capitalize">
                                    {{ keyFormatter(key) }}
                                </p>
                            </td>
                            <td class="px-10">
                                <p class="text-15 font-weight-medium">{{ val }}</p>
                            </td>
                        </tr>
                    </tbody>
                </v-card-item>
            </v-card>
        </v-col>
    </v-row>
</template>

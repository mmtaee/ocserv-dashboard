<script setup lang="ts">
import { useI18n } from 'vue-i18n';
import { onMounted, ref } from 'vue';
import { ReportApi, type ReportOcservUserReportResponse } from '@/api';
import { getAuthorization } from '@/utils/request';

const { t } = useI18n();

const userStats = ref<ReportOcservUserReportResponse>({
    active: 0,
    deactivated: 0,
    online: 0,
    locked: 0
});

const getUserStats = () => {
    const apiStats = new ReportApi();
    apiStats
        .reportsUsersGet({
            ...getAuthorization()
        })
        .then((res) => {
            Object.assign(userStats.value, res.data);
        });
};

onMounted(() => {
    getUserStats();
});

defineExpose({ getUserStats });
</script>

<template>
    <v-row align="center" justify="center" class="mb-5">
        <v-col cols="12" lg="2" sm="6">
            <v-card class="text-center" elevation="10">
                <v-card-title class="text-subtitle-1 mt-2 text-capitalize">
                    {{ t('ONLINE') }} {{ t('USERS') }}
                </v-card-title>

                <v-card-text class="text-muted text-h5">
                    {{ userStats.online || 0 }}
                </v-card-text>
            </v-card>
        </v-col>

        <v-col cols="12" lg="2" sm="6">
            <v-card class="text-center" elevation="10">
                <v-card-title class="text-subtitle-1 mt-2 text-capitalize">
                    {{ t('ACTIVE') }} {{ t('USERS') }}
                </v-card-title>

                <v-card-text class="text-muted text-h5">
                    {{ userStats.active || 0 }}
                </v-card-text>
            </v-card>
        </v-col>

        <v-col cols="12" lg="2" sm="6">
            <v-card class="text-center" elevation="10">
                <v-card-title class="text-subtitle-1 mt-2 text-capitalize">
                    {{ t('DEACTIVATED') }} {{ t('USERS') }}
                </v-card-title>

                <v-card-text class="text-muted text-h5">
                    {{ userStats.deactivated }}
                </v-card-text>
            </v-card>
        </v-col>

        <v-col cols="12" lg="2" sm="6">
            <v-card class="text-center" elevation="10">
                <v-card-title class="text-subtitle-1 mt-2 text-capitalize">
                    {{ t('LOCKED') }} {{ t('USERS') }}
                </v-card-title>

                <v-card-text class="text-muted text-h5">
                    {{ userStats.locked }}
                </v-card-text>
            </v-card>
        </v-col>
    </v-row>
</template>

<style scoped lang="scss"></style>

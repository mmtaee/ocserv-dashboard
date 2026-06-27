<script lang="ts" setup>
import { computed } from 'vue';
import { useTheme } from 'vuetify';
import { useI18n } from 'vue-i18n';
import type { HomeGetHomeUser } from '@/api';
import { buildTxRxDonutChartOptions } from '@/utils/apexChartsTheme';

const props = defineProps<{
    users: HomeGetHomeUser;
}>();

const theme = useTheme();
const { t } = useI18n();

const chartOptions = computed(() =>
    buildTxRxDonutChartOptions(
        {
            colors: theme.current.value.colors as Record<string, unknown>,
            dark: theme.global.current.value.dark
        },
        [t('TOTAL_USERS'), t('ONLINE_USERS')]
    )
);
const onlinePercentage = computed(() => {
    const total = props.users?.total ?? 0;
    const online = props.users?.online_users_session?.length || 0;

    if (total === 0) return 0;

    return Math.round((online / total) * 100);
});

const chart = computed(() => [props.users?.total || 0, props.users?.online_users_session?.length || 0]);
</script>
<template>
    <v-card elevation="10">
        <v-card-item>
            <div class="d-sm-flex align-center justify-space-between pt-sm-2">
                <v-card-title class="text-h5 text-capitalize">{{ t('ONLINE_USERS_OVERVIEW') }}</v-card-title>
            </div>
            <v-row>
                <v-col cols="7" sm="7">
                    <div class="mt-6">
                        <h6 class="text-h6 text-capitalize text-body-1">
                            {{ t('TOTAL_USERS') }}:
                            <br />
                            <span class="text-high-emphasis">
                                {{ props.users?.total }}
                            </span>
                        </h6>
                        <h6 class="text-h6 text-capitalize text-body-1 my-2">
                            {{ t('ONLINE_USERS') }}:
                            <br />
                            <span class="text-high-emphasis">
                                {{ props.users.online_users_session?.length || 0 }}
                            </span>
                        </h6>
                        <h6 class="text-h6 text-capitalize text-body-1 mt-5">
                            {{ t('AVERAGE') }}:
                            <span class="text-high-emphasis text-body-1"> {{ onlinePercentage }}% </span>
                        </h6>
                        <div class="d-flex align-center mt-sm-10 mt-8">
                            <h6 class="text-subtitle-1 text-medium-emphasis text-capitalize">
                                <v-icon
                                    class="mr-1"
                                    color="primary"
                                    icon="mdi mdi-checkbox-blank-circle"
                                    size="10"
                                ></v-icon>
                                {{ t('USERS') }}
                            </h6>
                            <h6 class="text-subtitle-1 text-medium-emphasis pl-5 text-capitalize">
                                <v-icon
                                    class="mr-1"
                                    color="lightprimary"
                                    icon="mdi mdi-checkbox-blank-circle"
                                    size="10"
                                ></v-icon>
                                {{ t('ONLINE') }}
                            </h6>
                        </div>
                    </div>
                </v-col>
                <v-col class="pl-lg-0" cols="5" sm="5">
                    <div class="d-flex align-center flex-shrink-0">
                        <apexchart
                            :options="chartOptions"
                            :series="chart"
                            class="pt-6"
                            height="145"
                            type="donut"
                        ></apexchart>
                    </div>
                </v-col>
            </v-row>
        </v-card-item>
    </v-card>
</template>

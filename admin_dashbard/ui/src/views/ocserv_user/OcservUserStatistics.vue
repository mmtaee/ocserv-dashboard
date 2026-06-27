<script lang="ts" setup>
import { computed, ref } from 'vue';
import { useI18n } from 'vue-i18n';
import { type ModelsDailyTraffic, OcservUsersApi, type RepositoryTotalBandwidths } from '@/api';
import DateRange from '@/components/shared/DateRange.vue';
import BarChart from '@/components/shared/BarChart.vue';
import DonutChart from '@/components/shared/DonutChart.vue';
import { numberToFixer } from '@/utils/convertors';
import UiChildCard from '@/components/shared/UiChildCard.vue';
import { getAuthorization } from '@/utils/request';
import UiParentCard from '@/components/shared/UiParentCard.vue';
import { router } from '@/router';

const props = defineProps<{
    uid: string;
    username?: string;
}>();

const { t } = useI18n();
const loading = ref(false);
const api = new OcservUsersApi();

const chartData = ref<ModelsDailyTraffic[]>([]);
const donutData = ref<RepositoryTotalBandwidths>({ rx: 0, tx: 0 });

// chartData.value = dummyTrafficData;
// donutData.value = {
//     rx: 10.0011,
//     tx: 50.5450215454
// };

const today = new Date();
const last30Days = new Date();
last30Days.setDate(today.getDate() - 31);

const initData = {
    dateStart: last30Days,
    dateEnd: today
};

const getStatistics = (dateStart: string, dateEnd: string) => {
    loading.value = true;
    api.ocservUsersUidStatisticsGet({
        ...getAuthorization(),
        uid: props.uid,
        dateStart: dateStart || '',
        dateEnd: dateEnd || ''
    })
        .then((res) => {
            chartData.value = res.data.statistics || [];
            donutData.value = res.data.total_bandwidths;
        })
        .finally(() => {
            loading.value = false;
        });
};

const txPercentage = computed(() => {
    const rx = donutData.value.rx ?? 0;
    const tx = donutData.value.tx ?? 0;
    const total = rx + tx;

    if (total === 0) return 0;
    return Math.round((tx / total) * 100);
});
</script>
<template>
    <v-row>
        <v-col cols="12" md="12">
            <UiParentCard :title="``">
                <template #header-prepend class="text-capitalize">
                    <v-tooltip :text="t('GO_BACK_TO_USERS')">
                        <template #activator="{ props }">
                            <v-icon start v-bind="props" @click.stop="router.push({ name: 'Ocserv Users' })">
                                mdi-arrow-left-top
                            </v-icon>
                        </template>
                    </v-tooltip>
                    RX / TX {{ t('STATISTICS') }} <span v-if="username" class="text-info ms-2">({{ username }})</span>
                </template>

                <UiChildCard class="px-2">
                    <DateRange :initDate="initData" :loading="loading" @search="getStatistics" />
                </UiChildCard>

                <UiChildCard :title="t('TOTAL_BANDWIDTHS')" class="mt-5 px-2">
                    <v-row align="center" justify="center">
                        <v-col cols="12" md="2">
                            <h6 class="text-h6 text-capitalize text-body-1">
                                {{ t('TOTAL') }} TX:
                                <br />
                                <span class="text-high-emphasis"> {{ numberToFixer(donutData.tx, 8) }} GB </span>
                            </h6>
                            <h6 class="text-h6 text-capitalize text-body-1 mt-5">
                                {{ t('TOTAL') }} RX:
                                <br />
                                <span class="text-high-emphasis text-body-1">
                                    {{ numberToFixer(donutData.rx, 8) }} GB
                                </span>
                            </h6>
                            <h6 class="text-h6 text-capitalize text-body-1 mt-5">
                                {{ t('AVERAGE') }} (TX):
                                <span class="text-high-emphasis text-body-1"> {{ txPercentage }}% </span>
                            </h6>
                        </v-col>

                        <v-col cols="12" md="3">
                            <v-row align="center" justify="center">
                                <v-col cols="12" md="12">
                                    <DonutChart :total-bandwidths="donutData" />
                                </v-col>

                                <v-col v-if="donutData.rx != 0 && donutData.tx != 0" cols="12" md="12">
                                    <v-row align="center" justify="center">
                                        <v-col cols="12" md="auto">
                                            <h6 class="text-subtitle-1 text-medium-emphasis">
                                                <v-icon
                                                    class="mr-1"
                                                    color="primary"
                                                    icon="mdi mdi-checkbox-blank-circle"
                                                    size="10"
                                                />
                                                TX
                                            </h6>
                                        </v-col>
                                        <v-col cols="12" md="auto">
                                            <h6 class="text-subtitle-1 text-medium-emphasis">
                                                <v-icon
                                                    class="mr-1"
                                                    color="lightprimary"
                                                    icon="mdi mdi-checkbox-blank-circle"
                                                    size="10"
                                                />
                                                RX
                                            </h6>
                                        </v-col>
                                    </v-row>
                                </v-col>
                            </v-row>
                        </v-col>
                    </v-row>
                </UiChildCard>

                <UiChildCard v-if="chartData.length > 0" class="mt-5 px-2" title="RX / TX">
                    <BarChart :data="chartData" />
                </UiChildCard>

                <UiChildCard v-else :title="t('NO_RX_TX_FOUND')" class="mt-5 px-2" />
            </UiParentCard>
        </v-col>
    </v-row>
</template>

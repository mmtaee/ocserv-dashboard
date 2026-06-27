<script lang="ts" setup>
import { useI18n } from 'vue-i18n';
import UiChildCard from '@/components/shared/UiChildCard.vue';
import DateRange from '@/components/shared/DateRange.vue';
import BarChart from '@/components/shared/BarChart.vue';
import { computed, ref } from 'vue';
import { type ModelsDailyTraffic, OcservUsersApi, type RepositoryTotalBandwidths } from '@/api';
import { getAuthorization } from '@/utils/request';
import { numberToFixer } from '@/utils/convertors';
import DonutChart from '@/components/shared/DonutChart.vue';

const props = defineProps({
    show: {
        type: Boolean,
        default: false
    },
    username: {
        type: String,
        required: true
    },
    uid: {
        type: String,
        required: true
    }
});

const emits = defineEmits(['close']);

const { t } = useI18n();

const loading = ref(false);
const chartData = ref<ModelsDailyTraffic[]>([]);
const donutData = ref<RepositoryTotalBandwidths>({ rx: 0, tx: 0 });
const today = new Date();
const last30Days = new Date();

last30Days.setDate(today.getDate() - 31);

const initData = {
    dateStart: last30Days,
    dateEnd: today
};

const search = (dateStart: string, dateEnd: string) => {
    const api = new OcservUsersApi();
    api.ocservUsersUidStatisticsGet({
        ...getAuthorization(),
        dateStart,
        dateEnd,
        uid: props.uid
    }).then((res) => {
        chartData.value = res.data.statistics || [];
        donutData.value = res.data.total_bandwidths;
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
    <v-dialog v-model="props.show" max-width="1200">
        <v-card>
            <v-card-title class="bg-info text-capitalize">
                <v-row align="end" class="no-gutters" justify="space-between">
                    <v-col md="auto"> {{ t('STATISTICS') }} ({{ username }}) </v-col>
                    <v-col md="auto">
                        <v-icon @click="emits('close')">mdi-close</v-icon>
                    </v-col>
                </v-row>
            </v-card-title>

            <v-card-text>
                <v-row align="start" justify="center">
                    <v-col cols="12" md="12">
                        <UiChildCard class="px-2 mt-0">
                            <DateRange
                                :disable-more30-days="false"
                                :initDate="initData"
                                :loading="loading"
                                @search="search"
                            />
                        </UiChildCard>
                    </v-col>

                    <v-col cols="12" md="6">
                        <UiChildCard :height="460" :title="t('TOTAL_BANDWIDTHS')" class="mt-5 px-2">
                            <v-row align="center" justify="center">
                                <v-col cols="12" md="auto">
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

                                <v-col cols="12" md="4">
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
                    </v-col>

                    <v-col v-if="chartData.length > 0" cols="12" md="6">
                        <UiChildCard :height="460" class="mt-5 px-2" title="RX / TX">
                            <BarChart :data="chartData" />
                        </UiChildCard>
                    </v-col>

                    <v-col v-else cols="12" md="6">
                        <UiChildCard :title="t('NO_FOUND_RX_TX_IN_STATS')" class="mt-5 px-2" />
                    </v-col>
                </v-row>
            </v-card-text>
        </v-card>
    </v-dialog>
</template>

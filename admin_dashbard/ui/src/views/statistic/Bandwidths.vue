<script lang="ts" setup>
import { numberToFixer } from '@/utils/convertors';
import { ReportApi, type RepositoryTotalBandwidths } from '@/api';
import { getAuthorization } from '@/utils/request';
import { useI18n } from 'vue-i18n';
import { computed, ref } from 'vue';
import UiChildCard from '@/components/shared/UiChildCard.vue';
import DonutChart from '@/components/shared/DonutChart.vue';
import DateRange from '@/components/shared/DateRange.vue';
import UiParentCard from '@/components/shared/UiParentCard.vue';

const { t } = useI18n();
const loading = ref(false);
const result = ref<RepositoryTotalBandwidths>({ rx: 0, tx: 0 });
const today = new Date();
const last30Days = new Date();
last30Days.setDate(today.getDate() - 31);

const initData = {
    dateStart: last30Days,
    dateEnd: today
};

const search = (dateStart: string, dateEnd: string) => {
    const api = new ReportApi();
    api.reportsTotalBandwidthGet({
        ...getAuthorization(),
        dateStart,
        dateEnd
    }).then((res) => {
        result.value = res.data;
        // result.value = {
        //     rx: 10.4444,
        //     tx: 99.568798973432
        // };
    });
};

const txPercentage = computed(() => {
    const rx = result.value.rx ?? 0;
    const tx = result.value.tx ?? 0;
    const total = rx + tx;

    if (total === 0) return 0;
    return Math.round((tx / total) * 100);
});
</script>

<template>
    <v-row>
        <v-col cols="12" md="12">
            <UiParentCard :title="t('TOTAL_BANDWIDTHS')">
                <UiChildCard class="px-2 mt-0">
                    <DateRange :initDate="initData" :loading="loading" @search="search" />
                </UiChildCard>

                <UiChildCard :title="t('TOTAL_BANDWIDTHS')" class="mt-0 px-2">
                    <v-row align="center" justify="center">
                        <v-col cols="12" md="2">
                            <h6 class="text-h6 text-capitalize text-body-1">
                                {{ t('TOTAL') }} TX:
                                <br />
                                <span class="text-high-emphasis"> {{ numberToFixer(result.tx, 8) }} GB </span>
                            </h6>
                            <h6 class="text-h6 text-capitalize text-body-1 mt-5">
                                {{ t('TOTAL') }} RX:
                                <br />
                                <span class="text-high-emphasis text-body-1"> {{ numberToFixer(result.rx, 8) }} GB </span>
                            </h6>
                            <h6 class="text-h6 text-capitalize text-body-1 mt-5">
                                {{ t('AVERAGE') }} (TX):
                                <span class="text-high-emphasis text-body-1"> {{ txPercentage }}% </span>
                            </h6>
                        </v-col>

                        <v-col cols="12" md="3">
                            <v-row align="center" justify="center">
                                <v-col cols="12" md="12">
                                    <DonutChart :total-bandwidths="result" />
                                </v-col>

                                <v-col v-if="result.rx != 0 && result.tx != 0" cols="12" md="12">
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
            </UiParentCard>
        </v-col>
    </v-row>
</template>

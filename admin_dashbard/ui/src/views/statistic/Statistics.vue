<script lang="ts" setup>
import { type ModelsDailyTraffic, ReportApi } from '@/api';
import { getAuthorization } from '@/utils/request';
import { ref } from 'vue';
import UiParentCard from '@/components/shared/UiParentCard.vue';
import { useI18n } from 'vue-i18n';
import UiChildCard from '@/components/shared/UiChildCard.vue';
import DateRange from '@/components/shared/DateRange.vue';
import BarChart from '@/components/shared/BarChart.vue';

const { t } = useI18n();
const loading = ref(false);
const result = ref<ModelsDailyTraffic[]>([]);
const today = new Date();
const last30Days = new Date();
last30Days.setDate(today.getDate() - 31);

const initData = {
    dateStart: last30Days,
    dateEnd: today
};

const search = (dateStart: string, dateEnd: string) => {
    const api = new ReportApi();
    api.reportsStatisticsGet({
        ...getAuthorization(),
        dateStart,
        dateEnd
    }).then((res) => {
        result.value = res.data;
        // result.value = dummyTrafficData;
    });
};
</script>

<template>
    <v-row>
        <v-col cols="12" md="12">
            <UiParentCard :title="`RX / TX ${t('STATISTICS')}`">
                <UiChildCard class="px-2 mt-0">
                    <DateRange :initDate="initData" :loading="loading" @search="search" />
                </UiChildCard>

                <UiChildCard v-if="result && result.length > 0" class="mt-0 px-2">
                    <BarChart :data="result" />
                </UiChildCard>

                <UiChildCard v-else :title="t('NO_FOUND_RX_TX_IN_STATS')" class="mt-5 px-2" />
            </UiParentCard>
        </v-col>
    </v-row>
</template>

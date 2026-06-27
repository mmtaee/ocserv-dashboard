<script lang="ts" setup>
import { computed } from 'vue';
import { useTheme } from 'vuetify';
import { useI18n } from 'vue-i18n';
import type { ModelsDailyTraffic } from '@/api';
import { buildBarRxTxChartOptions } from '@/utils/apexChartsTheme';

const theme = useTheme();
const { t } = useI18n();
const props = defineProps<{
    data: ModelsDailyTraffic[];
}>();
const chartOptions = computed(() => {
    const input = {
        colors: theme.current.value.colors as Record<string, unknown>,
        dark: theme.global.current.value.dark
    };
    const categories = props.data.map((d) => String(d.date ?? ''));
    const series = [
        { name: 'RX', data: props.data.map((d) => Number(d.rx ?? 0)) },
        { name: 'TX', data: props.data.map((d) => Number(d.tx ?? 0)) }
    ];
    return buildBarRxTxChartOptions(input, categories, series);
});
</script>
<template>
    <v-card elevation="10" height="567px">
        <v-card-item>
            <div class="d-sm-flex align-center justify-space-between pt-sm-2">
                <div>
                    <v-card-title class="text-h5 text-capitalize">RX / TX {{ t('TEN_DAYS_OVERVIEW') }}</v-card-title>
                </div>
            </div>

            <div v-if="data.length > 0" class="mt-6">
                <apexchart :options="chartOptions.chartOptions" :series="chartOptions.series" type="bar" />
            </div>

            <div v-else class="mt-6 text-capitalize">{{ t('NO_10_DAYS_BANDWIDTHS_OVERVIEW_FOUND') }}!</div>
        </v-card-item>
    </v-card>
</template>

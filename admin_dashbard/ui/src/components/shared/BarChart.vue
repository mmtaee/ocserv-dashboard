<script lang="ts" setup>
import { computed } from 'vue';
import { useTheme } from 'vuetify';
import type { ModelsDailyTraffic } from '@/api';
import { buildBarRxTxChartOptions } from '@/utils/apexChartsTheme';

const props = defineProps<{
    data: ModelsDailyTraffic[];
}>();

const theme = useTheme();

const barOptions = computed(() => {
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
    <apexchart :options="barOptions.chartOptions" :series="barOptions.series" type="bar" />
</template>

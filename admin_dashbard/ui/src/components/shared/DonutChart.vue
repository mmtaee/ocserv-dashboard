<script lang="ts" setup>
import { computed } from 'vue';
import type { RepositoryTotalBandwidths } from '@/api';
import { useI18n } from 'vue-i18n';
import { useTheme } from 'vuetify';
import { buildTxRxDonutChartOptions } from '@/utils/apexChartsTheme';

const props = defineProps<{
    totalBandwidths: RepositoryTotalBandwidths;
}>();

const { t } = useI18n();
const theme = useTheme();

const donutOptions = computed(() =>
    buildTxRxDonutChartOptions(
        {
            colors: theme.current.value.colors as Record<string, unknown>,
            dark: theme.global.current.value.dark
        },
        [t('TOTAL_TX'), t('TOTAL_RX')]
    )
);

const chart = computed(() => [+props.totalBandwidths?.tx.toFixed(6) || 0, +props.totalBandwidths?.rx.toFixed(6) || 0]);
</script>

<template>
    <apexchart :options="donutOptions" :series="chart" height="180" type="donut" />
</template>

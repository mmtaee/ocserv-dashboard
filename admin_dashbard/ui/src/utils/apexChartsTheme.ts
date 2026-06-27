/** Shared ApexCharts options aligned with Vuetify Ocean light/dark themes. */

export type ApexThemeInput = {
    colors: Record<string, unknown>;
    dark: boolean;
};

function mutedForeColor(c: Record<string, unknown>): string {
    return String(c.textSecondary ?? c['on-surface-variant'] ?? '#64748B');
}

function emptyDonutSlice(c: Record<string, unknown>, dark: boolean): string {
    return String(c.surface ?? c.background ?? (dark ? '#1E293B' : '#FFFFFF'));
}

/** Two-series TX/RX style donut (third color = unused ring segment). */
export function buildTxRxDonutChartOptions(input: ApexThemeInput, labels: string[]) {
    const c = input.colors as Record<string, string>;
    const foreColor = mutedForeColor(input.colors);
    const emptySlice = emptyDonutSlice(input.colors, input.dark);

    return {
        labels,
        chart: {
            type: 'donut' as const,
            fontFamily: 'inherit',
            foreColor,
            toolbar: { show: false }
        },
        colors: [String(c.primary), String(c.lightprimary), emptySlice],
        plotOptions: {
            pie: {
                startAngle: 0,
                endAngle: 360,
                donut: {
                    size: '75%',
                    background: 'transparent'
                }
            }
        },
        stroke: { show: false },
        dataLabels: { enabled: false },
        legend: { show: false },
        tooltip: { theme: input.dark ? 'dark' : 'light', fillSeriesColor: false }
    };
}

const barResponsive = [
    {
        breakpoint: 600,
        options: {
            plotOptions: { bar: { borderRadius: 3 } }
        }
    }
];

/** Stacked/grouped bar chart for daily RX/TX traffic. */
export function buildBarRxTxChartOptions(
    input: ApexThemeInput,
    categories: string[],
    series: Array<{ name: string; data: number[] }>
) {
    const c = input.colors as Record<string, string>;
    const muted = mutedForeColor(input.colors);
    const gridBorder = input.dark ? 'rgba(148, 163, 184, 0.22)' : 'rgba(15, 23, 42, 0.12)';

    return {
        series,
        chartOptions: {
            chart: {
                type: 'bar' as const,
                height: 400,
                toolbar: { show: true },
                fontFamily: 'inherit',
                foreColor: muted
            },
            colors: [String(c.primary), String(c.secondary)],
            grid: {
                borderColor: gridBorder,
                strokeDashArray: 3
            },
            plotOptions: {
                bar: { horizontal: false, columnWidth: '40%', borderRadius: 8 }
            },
            xaxis: {
                type: 'category' as const,
                categories,
                labels: {
                    style: {
                        colors: muted,
                        fontSize: '11px',
                        fontFamily: 'inherit'
                    }
                }
            },
            yaxis: {
                min: 0,
                labels: {
                    style: {
                        colors: muted,
                        fontSize: '11px',
                        fontFamily: 'inherit'
                    }
                }
            },
            dataLabels: { enabled: false },
            tooltip: { theme: input.dark ? 'dark' : 'light', fillSeriesColor: false },
            responsive: barResponsive
        }
    };
}

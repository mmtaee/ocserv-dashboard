import type { ThemeTypes } from '@/types/themeTypes/ThemeType';

const OceanLight: ThemeTypes = {
    name: 'OceanLight',
    dark: false,
    variables: {
        'border-color': '#E2E8F0',
        'carousel-control-size': 10
    },
    colors: {
        primary: '#0EA5A4',
        secondary: '#0891B2',
        info: '#0284C7',
        success: '#10B981',
        accent: '#F59E0B',
        warning: '#F59E0B',
        error: '#EF4444',
        muted: '#64748B',
        lightprimary: '#CCFBF1',
        lightsecondary: '#CFFAFE',
        lightsuccess: '#D1FAE5',
        lighterror: '#FEE2E2',
        lightwarning: '#FEF3C7',
        lightinfo: '#E0F2FE',
        textPrimary: '#0F172A',
        textSecondary: '#475569',
        borderColor: '#E2E8F0',
        inputBorder: '#94A3B8',
        containerBg: '#FFFFFF',
        hoverColor: '#F1F5F9',
        background: '#F8FAFC',
        surface: '#FFFFFF',
        'on-surface': '#0F172A',
        'on-surface-variant': '#475569',
        grey100: '#F1F5F9'
    }
};

export { OceanLight };

// Backwards compatibility for any legacy import paths.
export const BlueTheme = OceanLight;

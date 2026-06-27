import { createVuetify } from 'vuetify';
import '@mdi/font/css/materialdesignicons.css';
import * as components from 'vuetify/components';
import * as directives from 'vuetify/directives';
import { OceanLight } from '@/theme/LightTheme';
import { OceanDark } from '@/theme/DarkTheme';
import { createVueI18nAdapter } from 'vuetify/locale/adapters/vue-i18n';
import i18n from '@/plugins/i18n';
import { useI18n } from 'vue-i18n';
import DateFnsAdapter from '@date-io/date-fns';
import { enUS } from 'date-fns/locale';

export const THEME_STORAGE_KEY = 'theme';
export const LIGHT_THEME = OceanLight.name;
export const DARK_THEME = OceanDark.name;

function resolveInitialTheme(): string {
    const stored = localStorage.getItem(THEME_STORAGE_KEY);
    if (stored === LIGHT_THEME || stored === DARK_THEME) {
        return stored;
    }
    if (typeof window !== 'undefined' && window.matchMedia?.('(prefers-color-scheme: dark)').matches) {
        return DARK_THEME;
    }
    return LIGHT_THEME;
}

export default createVuetify({
    components,
    directives,
    locale: {
        adapter: createVueI18nAdapter({ i18n: i18n as any, useI18n }),
        rtl: { fa: true, ar: true, he: true, ur: true }
    },
    date: {
        adapter: DateFnsAdapter,
        locale: {
            en: enUS,
            fa: enUS,
            ar: enUS,
            he: enUS,
            ur: enUS
        }
    },
    theme: {
        defaultTheme: resolveInitialTheme(),
        themes: {
            [OceanLight.name]: OceanLight,
            [OceanDark.name]: OceanDark
        }
    },
    defaults: {
        VBtn: {},
        VCard: {
            rounded: 'md'
        },
        VTextField: {
            rounded: 'lg'
        },
        VTooltip: {
            location: 'top'
        }
    }
});

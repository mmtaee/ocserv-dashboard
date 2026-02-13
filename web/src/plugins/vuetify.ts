import { createVuetify } from 'vuetify';
import '@mdi/font/css/materialdesignicons.css';
import * as components from 'vuetify/components';
import * as directives from 'vuetify/directives';
import { BlueTheme } from '@/theme/LightTheme';
import { createVueI18nAdapter } from 'vuetify/locale/adapters/vue-i18n';
import i18n from '@/plugins/i18n';
import { useI18n } from 'vue-i18n';
import DateFnsAdapter from '@date-io/date-fns'; // Import the adapter class
import { enUS } from 'date-fns/locale';

export default createVuetify({
    components,
    directives,
    locale: {
        adapter: createVueI18nAdapter({ i18n: i18n as any, useI18n })
    },
    date: {
        adapter: DateFnsAdapter,
        locale: {
            en: enUS,
            fa: enUS, // Persian
            ar: enUS, // Arabic
            he: enUS, // Hebrew
            ur: enUS // Urdu
        }
    },
    theme: {
        defaultTheme: 'BlueTheme',
        themes: {
            BlueTheme
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
            // set v-tooltip default location to top
            location: 'top'
        }
    }
});

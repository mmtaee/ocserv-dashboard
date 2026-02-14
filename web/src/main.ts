import '@/assets/scss/style.scss';
import { createApp } from 'vue';
import App from './App.vue';
import { router } from './router';
import vuetify from './plugins/vuetify';
import PerfectScrollbar from 'vue3-perfect-scrollbar';
import VueApexCharts from 'vue3-apexcharts';
import i18n from '@/plugins/i18n';

import { createPinia } from 'pinia';
import { useConfigStore, useServerStore } from '@/stores/config';
import { useProfileStore } from '@/stores/profile';

const app = createApp(App);

const stopLoader = async () => {
    const preloader = document.getElementById('preloader');
    const preloaderScript = document.getElementById('preloader-script');
    const preloaderStyle = document.getElementById('preloader-style');

    if (preloader) {
        preloader.style.opacity = '0';
        setTimeout(() => {
            preloader.remove(); // remove loader div
        }, 500);
    }

    if (preloaderScript) {
        preloaderScript.remove();
    }

    if (preloaderStyle) {
        preloaderStyle.remove();
    }
};

app.use(createPinia());

(async () => {
    const serverStore = useServerStore();
    await serverStore.getServerInfo();

    const configStore = useConfigStore();
    const setup = await configStore.getConfig();

    app.use(vuetify);
    app.use(i18n);
    app.use(router);
    app.use(PerfectScrollbar);
    app.use(VueApexCharts as any);

    if (!setup) {
        localStorage.removeItem('token');
        await router.push({ name: 'System Setup' });
    } else {
        if (localStorage.getItem('token')) {
            const profileStore = useProfileStore();
            await profileStore.getProfile();
        }
    }

    await stopLoader();

    app.mount('#app');
})();

import { createApp } from 'vue'

import App from '@/App.vue'
import i18n from '@/plugins/i18n'
import vuetify from '@/plugins/vuetify'
import router from '@/router'
import '@shared/assets/main.css'

createApp(App).use(router).use(vuetify).use(i18n).mount('#app')

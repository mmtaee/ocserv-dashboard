import { createApp } from 'vue'
import './style.css'
import App from './App.vue'
import vuetify from './plugins/vuetify'
import router from './router'
import i18n from './plugins/i18n'
import pinia from './stores'

createApp(App)
  .use(vuetify)
  .use(router)
  .use(i18n)
  .use(pinia)
  .mount('#app')

import { createPinia } from 'pinia'
import { createApp } from 'vue'

import App from './App.vue'
import { getAccessToken } from '@/api/auth-token'
import { installRouterGuards, router } from '@/router'
import { useAuthStore } from '@/stores/auth'
import { useSystemInitStore } from '@/stores/system-init'

import './style.css'

async function bootstrap(): Promise<void> {
  const app = createApp(App)
  const pinia = createPinia()

  app.use(pinia)
  app.use(router)

  const systemInit = useSystemInitStore(pinia)
  const auth = useAuthStore(pinia)

  await systemInit.initialize()

  if (systemInit.isAvailable && systemInit.isInitialized && getAccessToken()) {
    await auth.restoreSession()
  }

  installRouterGuards(pinia)
  await router.replace({
    name: !systemInit.isAvailable
      ? 'server-unavailable'
      : !systemInit.isInitialized
        ? 'system-setup'
        : auth.isAuthenticated
          ? 'home'
          : 'login',
  })

  await router.isReady()
  app.mount('#app')
}

void bootstrap()

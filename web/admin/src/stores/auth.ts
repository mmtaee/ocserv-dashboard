import { defineStore } from 'pinia'
import { computed, ref } from 'vue'

import { clearAccessToken, getAccessToken } from '@/api/auth-token'
import type {
  GithubComMmtaeeOcservDashboardBackendInternalServicesAdminApiSystemLoginData,
  ModelsUser,
} from '@/api/generated'
import { normalizeApiError } from '@/api/http'
import { getCurrentUser, login } from '@/api/services/auth'

export const useAuthStore = defineStore('auth', () => {
  const user = ref<ModelsUser | null>(null)
  const isLoading = ref(false)
  const error = ref<string | null>(null)
  const isAuthenticated = computed(() => Boolean(user.value && getAccessToken()))

  async function signIn(
    credentials: GithubComMmtaeeOcservDashboardBackendInternalServicesAdminApiSystemLoginData,
  ): Promise<boolean> {
    isLoading.value = true
    error.value = null
    try {
      const response = await login(credentials)
      user.value = response.user
      return true
    } catch (cause) {
      clearAccessToken()
      user.value = null
      error.value = normalizeApiError(cause).message
      return false
    } finally {
      isLoading.value = false
    }
  }

  async function restoreSession(): Promise<boolean> {
    if (!getAccessToken()) return false
    isLoading.value = true
    error.value = null
    try {
      user.value = await getCurrentUser()
      return true
    } catch {
      clearAccessToken()
      user.value = null
      return false
    } finally {
      isLoading.value = false
    }
  }

  return { error, isAuthenticated, isLoading, restoreSession, signIn, user }
})

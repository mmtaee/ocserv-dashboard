import { defineStore } from "pinia";
import { computed, ref } from "vue";

import { clearAccessToken, getAccessToken } from "@/api/auth-token";
import type {
  GithubComMmtaeeOcservDashboardBackendInternalServicesAdminApiSystemLoginData,
  ModelsUser,
} from "@/api/generated";
import { normalizeApiError } from "@/api/http";
import { getCurrentUser, login, logout } from "@/api/services/auth";

export const useAuthStore = defineStore("auth", () => {
  const user = ref<ModelsUser | null>(null);
  const isLoading = ref(false);
  const error = ref<string | null>(null);
  const isAuthenticated = computed(() =>
    Boolean(user.value && getAccessToken()),
  );

  function clearSession(): void {
    clearAccessToken();
    user.value = null;
  }

  async function signIn(
    credentials: GithubComMmtaeeOcservDashboardBackendInternalServicesAdminApiSystemLoginData,
  ): Promise<boolean> {
    isLoading.value = true;
    error.value = null;
    try {
      await login(credentials);
      user.value = await getCurrentUser();
      return true;
    } catch (cause) {
      clearSession();
      error.value = normalizeApiError(cause).message;
      return false;
    } finally {
      isLoading.value = false;
    }
  }

  async function restoreSession(): Promise<boolean> {
    if (!getAccessToken()) return false;
    isLoading.value = true;
    error.value = null;
    try {
      user.value = await getCurrentUser();
      return true;
    } catch {
      clearSession();
      return false;
    } finally {
      isLoading.value = false;
    }
  }

  async function signOut(): Promise<void> {
    isLoading.value = true;
    error.value = null;
    try {
      await logout();
    } catch {
      // Local session state must still be cleared if the server is unavailable.
    } finally {
      clearSession();
      isLoading.value = false;
    }
  }

  return {
    clearSession,
    error,
    isAuthenticated,
    isLoading,
    restoreSession,
    signIn,
    signOut,
    user,
  };
});

import { computed, shallowRef } from "vue";
import { defineStore } from "pinia";

import {
  clearAccessToken,
  getAccessToken,
  setAccessToken,
} from "@/api/auth-token";
import type { Customer, LoginData } from "@/api/generated";
import { normalizeApiError } from "@/api/http";
import { getSummary, login } from "@/api/services/customers";

export const useAuthStore = defineStore("auth", () => {
  const user = shallowRef<Customer | null>(null);
  const loading = shallowRef(false);
  const error = shallowRef("");
  const isAuthenticated = computed(() =>
    Boolean(user.value && getAccessToken()),
  );
  function clearSession(): void {
    clearAccessToken();
    user.value = null;
  }
  async function signIn(credentials: LoginData): Promise<boolean> {
    loading.value = true;
    error.value = "";
    try {
      const response = await login(credentials);
      setAccessToken(response.token, response.expires_at);
      user.value = response.user;
      return true;
    } catch (cause) {
      clearSession();
      error.value = normalizeApiError(cause).message;
      return false;
    } finally {
      loading.value = false;
    }
  }
  async function restoreSession(): Promise<boolean> {
    if (!getAccessToken()) return false;
    try {
      user.value = (await getSummary()).ocserv_user ?? null;
      return Boolean(user.value);
    } catch {
      clearSession();
      return false;
    }
  }
  function signOut(): void {
    clearSession();
  }
  return {
    clearSession,
    error,
    isAuthenticated,
    loading,
    restoreSession,
    signIn,
    signOut,
    user,
  };
});

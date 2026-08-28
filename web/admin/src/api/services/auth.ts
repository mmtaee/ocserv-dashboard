import {
  clearAccessToken,
  requireAuthorizationHeader,
  setAccessToken,
} from "@/api/auth-token";
import { api } from "@/api/client";
import type { GithubComMmtaeeOcservDashboardBackendInternalServicesAdminApiSystemLoginData } from "@/api/generated";

export async function login(
  credentials: GithubComMmtaeeOcservDashboardBackendInternalServicesAdminApiSystemLoginData,
) {
  const response = await api.systemUsers.systemUsersLoginPost({
    request: credentials,
  });
  setAccessToken(response.data.token);

  return response.data;
}

export async function logout(): Promise<void> {
  try {
    await api.auth.authLogoutPost({
      authorization: requireAuthorizationHeader(),
    });
  } finally {
    clearAccessToken();
  }
}

export async function getCurrentUser() {
  const response = await api.systemUsers.systemUsersProfileGet({
    authorization: requireAuthorizationHeader(),
  });

  return response.data;
}

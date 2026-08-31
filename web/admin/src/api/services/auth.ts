import {
  clearAccessToken,
  requireAuthorizationHeader,
  setAccessToken,
} from "@/api/auth-token";
import { api } from "@/api/client";
import { isTestMode } from "@/api/environment";
import type {
  GithubComMmtaeeOcservDashboardBackendInternalServicesAdminApiSystemLoginData,
  GithubComMmtaeeOcservDashboardBackendInternalServicesAdminApiSystemUserLoginResponse,
  ModelsUser,
} from "@/api/generated";
import { cloneMock, mockCurrentUser, mockLoginResponse } from "@/mocks";

export async function login(
  credentials: GithubComMmtaeeOcservDashboardBackendInternalServicesAdminApiSystemLoginData,
): Promise<GithubComMmtaeeOcservDashboardBackendInternalServicesAdminApiSystemUserLoginResponse> {
  if (isTestMode) {
    const response = cloneMock(mockLoginResponse);
    response.user.username = credentials.username || response.user.username;
    setAccessToken(response.token);
    return response;
  }

  const response = await api.systemUsers.systemUsersLoginPost({
    request: credentials,
  });
  setAccessToken(response.data.token);

  return response.data;
}

export async function logout(): Promise<void> {
  if (isTestMode) {
    clearAccessToken();
    return;
  }

  try {
    await api.auth.authLogoutPost({
      authorization: requireAuthorizationHeader(),
    });
  } finally {
    clearAccessToken();
  }
}

export async function getCurrentUser(): Promise<ModelsUser> {
  if (isTestMode) {
    return cloneMock(mockCurrentUser);
  }

  const response = await api.systemUsers.systemUsersProfileGet({
    authorization: requireAuthorizationHeader(),
  });

  return response.data;
}

import {
  clearAccessToken,
  requireAuthorizationHeader,
  setAccessToken,
} from "@/api/auth-token";
import { api } from "@/api/client";
import { isTestMode } from "@/api/environment";
import { httpClient } from "@/api/http";
import type {
  GithubComMmtaeeOcservDashboardBackendInternalServicesAdminApiSystemLoginData,
  GithubComMmtaeeOcservDashboardBackendInternalServicesAdminApiSystemUserLoginResponse,
  ModelsUser,
  GithubComMmtaeeOcservDashboardBackendInternalServicesAdminApiSystemChangeUserPasswordBySelf,
} from "@/api/generated";
import {
  cloneMock,
  mockCurrentUser,
  mockLoginResponse,
  resetMockAdminPassword,
} from "@/mocks";

export interface ResetAdminPasswordRequest {
  new_password: string;
  secret_key: string;
}

export interface ResetAdminPasswordResponse {
  token: string;
  user: ModelsUser;
}

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

export async function changePassword(
  request: GithubComMmtaeeOcservDashboardBackendInternalServicesAdminApiSystemChangeUserPasswordBySelf,
): Promise<void> {
  if (isTestMode) return;

  await api.systemUsers.systemUsersPasswordPost({
    authorization: requireAuthorizationHeader(),
    request,
  });
}

export async function resetAdminPassword(
  request: ResetAdminPasswordRequest,
): Promise<ResetAdminPasswordResponse> {
  if (isTestMode) return resetMockAdminPassword(request);

  const response = await httpClient.post<ResetAdminPasswordResponse>(
    "/system/user/reset-password",
    request,
  );
  return response.data;
}

import { requireAuthorizationHeader } from "@/api/auth-token";
import { api } from "@/api/client";
import { isTestMode } from "@/api/environment";
import type {
  GithubComMmtaeeOcservDashboardBackendInternalServicesAdminApiSystemGetSystemInitResponse,
  GithubComMmtaeeOcservDashboardBackendInternalServicesAdminApiSystemGetSystemResponse,
  GithubComMmtaeeOcservDashboardBackendInternalServicesAdminApiSystemPatchSystemUpdateData,
} from "@/api/generated";
import { cloneMock, mockSystemConfig, mockSystemInit } from "@/mocks";

export type SystemInitConfig =
  GithubComMmtaeeOcservDashboardBackendInternalServicesAdminApiSystemGetSystemInitResponse;

export type SystemUpdateData =
  GithubComMmtaeeOcservDashboardBackendInternalServicesAdminApiSystemPatchSystemUpdateData;

export async function getSystemInit(): Promise<SystemInitConfig | null> {
  if (isTestMode) {
    return cloneMock(mockSystemInit);
  }

  const response = await api.system.systemInitGet();
  return response.data ?? null;
}

export async function updateSystemConfig(
  request: SystemUpdateData,
): Promise<GithubComMmtaeeOcservDashboardBackendInternalServicesAdminApiSystemGetSystemResponse> {
  if (isTestMode) {
    return {
      ...cloneMock(mockSystemConfig),
      ...request,
      first_init: true,
    };
  }

  const response = await api.system.systemPatch({
    authorization: requireAuthorizationHeader(),
    request,
  });
  return response.data;
}

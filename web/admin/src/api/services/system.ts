import { requireAuthorizationHeader } from "@/api/auth-token";
import { api } from "@/api/client";
import { isTestMode } from "@/api/environment";
import { apiBaseUrl } from "@/api/http";
import type {
  GithubComMmtaeeOcservDashboardBackendInternalServicesAdminApiSystemGetSystemInitResponse,
  GithubComMmtaeeOcservDashboardBackendInternalServicesAdminApiSystemGetSystemResponse,
  GithubComMmtaeeOcservDashboardBackendInternalServicesAdminApiSystemPatchSystemUpdateData,
  GithubComMmtaeeOcservDashboardBackendInternalServicesAdminApiSystemDashboardRelease,
  ModelsOcservAgent,
  OcservAgentCreateInput,
  OcservAgentUpdateInput,
  RuntimeActionResponse,
  RuntimeOcservConfig,
  RuntimeStatusResponse,
} from "@/api/generated";
import {
  cloneMock,
  getMockAgents,
  getMockOcservConfig,
  getMockSystemConfig,
  getMockRelease,
  mockSystemInit,
  mutateMockAgent,
  updateMockOcservConfig,
  updateMockSystemConfig,
} from "@/mocks";
import { getMockRuntimeStatus, runMockRuntimeAction } from "@/mocks/runtime";

export type SystemInitConfig =
  GithubComMmtaeeOcservDashboardBackendInternalServicesAdminApiSystemGetSystemInitResponse;

export type SystemUpdateData =
  GithubComMmtaeeOcservDashboardBackendInternalServicesAdminApiSystemPatchSystemUpdateData;
export type SystemConfig =
  GithubComMmtaeeOcservDashboardBackendInternalServicesAdminApiSystemGetSystemResponse;
export type OcservConfig = RuntimeOcservConfig;
export type OcservConfigResponse = RuntimeOcservConfig & {
  allow_update: boolean;
};
export type SystemRelease =
  GithubComMmtaeeOcservDashboardBackendInternalServicesAdminApiSystemDashboardRelease;
export type OcservAgent = ModelsOcservAgent;
export type OcservAgentCreate = OcservAgentCreateInput;
export type OcservAgentUpdate = OcservAgentUpdateInput;

export type OcservRuntimeActionResult = RuntimeActionResponse;
export const OcservRuntimeAction = {
  Restart: "restart",
  Enable: "enable",
  Disable: "disable",
} as const;
export type OcservRuntimeAction =
  (typeof OcservRuntimeAction)[keyof typeof OcservRuntimeAction];
export type OcservRuntimeStatusResponse = RuntimeStatusResponse & {
  allowed_actions?: OcservRuntimeAction[] | null;
};
export type OcservRuntimeStatus = OcservRuntimeStatusResponse & {
  allowed_actions: OcservRuntimeAction[];
};

const runtimeActions = new Set<unknown>(Object.values(OcservRuntimeAction));

function normalizeRuntimeStatus(
  status: OcservRuntimeStatusResponse,
): OcservRuntimeStatus {
  return {
    ...status,
    allowed_actions: Array.isArray(status.allowed_actions)
      ? status.allowed_actions.filter((action) => runtimeActions.has(action))
      : [],
  };
}

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
    return updateMockSystemConfig(request);
  }

  const response = await api.system.systemPatch({
    authorization: requireAuthorizationHeader(),
    request,
  });
  return response.data;
}

export async function getSystemConfig(): Promise<SystemConfig> {
  if (isTestMode) return getMockSystemConfig();
  return (
    await api.system.systemGet({ authorization: requireAuthorizationHeader() })
  ).data;
}

export async function getOcservConfig(): Promise<OcservConfigResponse> {
  if (isTestMode) return getMockOcservConfig();
  return (
    await api.system.systemOcservConfigGet({
      authorization: requireAuthorizationHeader(),
    })
  ).data as OcservConfigResponse;
}

export async function updateOcservConfig(
  request: OcservConfig,
): Promise<OcservConfig> {
  if (isTestMode) return updateMockOcservConfig(request);
  return (
    await api.system.systemOcservConfigPatch({
      authorization: requireAuthorizationHeader(),
      request,
    })
  ).data;
}

export async function getSystemRelease(): Promise<SystemRelease> {
  if (isTestMode) return getMockRelease();
  return (await api.system.systemReleaseGet()).data;
}

export async function getOcservAgents(): Promise<OcservAgent[]> {
  if (isTestMode) return getMockAgents();
  return (
    await api.agents.ocservAgentsGet(
      { authorization: requireAuthorizationHeader() },
      { baseURL: apiBaseUrl },
    )
  ).data;
}

export async function getOcservAgent(id: number): Promise<OcservAgent> {
  if (isTestMode)
    return (await getMockAgents()).find((agent) => agent.id === id)!;
  return (
    await api.agents.ocservAgentsIdGet(
      { authorization: requireAuthorizationHeader(), id },
      { baseURL: apiBaseUrl },
    )
  ).data;
}

export async function createOcservAgent(
  request: OcservAgentCreate,
): Promise<OcservAgent> {
  if (isTestMode) return (await mutateMockAgent("create", request))!;
  return (
    await api.agents.ocservAgentsPost(
      { authorization: requireAuthorizationHeader(), request },
      { baseURL: apiBaseUrl },
    )
  ).data;
}

export async function updateOcservAgent(
  id: number,
  request: OcservAgentUpdate,
): Promise<OcservAgent> {
  if (isTestMode) return (await mutateMockAgent("update", request, id))!;
  return (
    await api.agents.ocservAgentsIdPatch(
      { authorization: requireAuthorizationHeader(), id, request },
      { baseURL: apiBaseUrl },
    )
  ).data;
}

export async function deleteOcservAgent(id: number): Promise<void> {
  if (isTestMode)
    return mutateMockAgent("delete", undefined, id) as Promise<void>;
  await api.agents.ocservAgentsIdDelete(
    { authorization: requireAuthorizationHeader(), id },
    { baseURL: apiBaseUrl },
  );
}

export async function getRuntimeStatus(): Promise<OcservRuntimeStatus> {
  if (isTestMode) return normalizeRuntimeStatus(await getMockRuntimeStatus());

  const response = await api.system.systemdStatusGet({
    authorization: requireAuthorizationHeader(),
  });
  return normalizeRuntimeStatus(response.data as OcservRuntimeStatusResponse);
}

export async function runRuntimeAction(
  action: OcservRuntimeAction,
): Promise<OcservRuntimeActionResult> {
  if (isTestMode) return runMockRuntimeAction(action);

  const request = { authorization: requireAuthorizationHeader() };
  const response =
    action === "enable"
      ? await api.system.systemdEnablePost(request)
      : action === "disable"
        ? await api.system.systemdDisablePost(request)
        : await api.system.systemdRestartPost(request);
  return response.data;
}

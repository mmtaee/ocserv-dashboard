import { requireAuthorizationHeader } from "@/api/auth-token";
import { api } from "@/api/client";
import { isTestMode } from "@/api/environment";
import type { ModelsOcservInfo } from "@/api/generated";
import { cloneMock, mockOcctlCommand, mockOcctlServerInfo } from "@/mocks";

export type OcctlServerInfo = ModelsOcservInfo;
export type OcctlAction =
  1 | 2 | 3 | 4 | 5 | 6 | 7 | 8 | 9 | 10 | 11 | 12 | 13 | 14 | 15 | 16;

export interface OcctlCommandRequest {
  action: OcctlAction;
  value?: string;
}

export async function getOcctlServerInfo(): Promise<OcctlServerInfo> {
  if (isTestMode) return cloneMock(mockOcctlServerInfo);

  const response = await api.occtl.occtlServerInfoGet();
  return response.data;
}

export async function executeOcctlCommand(
  request: OcctlCommandRequest,
): Promise<string> {
  if (isTestMode) return mockOcctlCommand(request.action, request.value);

  const response = await api.occtl.occtlCommandsGet({
    authorization: requireAuthorizationHeader(),
    action: request.action,
    value: request.value,
  });
  return response.data;
}

import { requireAuthorizationHeader } from "@/api/auth-token";
import { api } from "@/api/client";
import { isTestMode } from "@/api/environment";
import type {
  ModelsOcservGroupConfig,
  OcservGroupUpdateOcservGroupData,
} from "@/api/generated";
import { cloneMock, mockOcservDefaultGroupConfig } from "@/mocks";

export type OcservDefaultGroupConfig = ModelsOcservGroupConfig;
export type OcservDefaultGroupUpdate = OcservGroupUpdateOcservGroupData;

let currentMockConfig: ModelsOcservGroupConfig = cloneMock(
  mockOcservDefaultGroupConfig,
);

export async function getOcservDefaultGroupConfig(): Promise<OcservDefaultGroupConfig> {
  if (isTestMode) return cloneMock(currentMockConfig);

  const response = await api.groups.ocservGroupsDefaultsGet({
    authorization: requireAuthorizationHeader(),
  });

  // The generated GET response is untyped because the Swagger annotation uses
  // a generic object, while the controller returns ModelsOcservGroupConfig.
  return response.data as ModelsOcservGroupConfig;
}

export async function updateOcservDefaultGroupConfig(
  request: OcservDefaultGroupUpdate,
): Promise<void> {
  if (isTestMode) {
    currentMockConfig = cloneMock(request.config);
    return;
  }

  await api.groups.ocservGroupsDefaultsPatch({
    authorization: requireAuthorizationHeader(),
    request,
  });
}

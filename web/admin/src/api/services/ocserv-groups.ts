import { requireAuthorizationHeader } from "@/api/auth-token";
import { api } from "@/api/client";
import { isTestMode } from "@/api/environment";
import type {
  ModelsOcservGroupConfig,
  ModelsOcservGroup,
  OcservGroupCreateOcservGroupData,
  OcservGroupOcservGroupsResponse,
  OcservGroupUpdateOcservGroupData,
  OcservGroupsApiOcservGroupsGetRequest,
} from "@/api/generated";
import { ApiError } from "@/api/http";
import {
  cloneMock,
  mockOcservDefaultGroupConfig,
  mockOcservGroups,
} from "@/mocks";

export type OcservDefaultGroupConfig = ModelsOcservGroupConfig;
export type OcservDefaultGroupUpdate = OcservGroupUpdateOcservGroupData;
export type OcservGroup = ModelsOcservGroup;
export type OcservGroupCreate = OcservGroupCreateOcservGroupData;
export type OcservGroupUpdate = OcservGroupUpdateOcservGroupData;
export type OcservGroupsList = OcservGroupOcservGroupsResponse;
export type OcservGroupsListOptions = Omit<
  OcservGroupsApiOcservGroupsGetRequest,
  "authorization"
>;

let currentMockConfig: ModelsOcservGroupConfig = cloneMock(
  mockOcservDefaultGroupConfig,
);
let currentMockGroups: ModelsOcservGroup[] = cloneMock(mockOcservGroups);

function mockGroup(id: number): ModelsOcservGroup {
  const group = currentMockGroups.find((item) => item.id === id);
  if (!group) throw new ApiError("Ocserv group not found.", { status: 400 });
  return group;
}

export async function getOcservGroups(
  options: OcservGroupsListOptions = {},
): Promise<OcservGroupsList> {
  if (isTestMode) {
    const page = options.page ?? 1;
    const size = options.size ?? 25;
    const direction = options.sort === "DESC" ? -1 : 1;
    const ordered = [...currentMockGroups].sort((left, right) => {
      if (options.order === "id")
        return ((left.id ?? 0) - (right.id ?? 0)) * direction;
      return left.name.localeCompare(right.name) * direction;
    });
    return {
      meta: { page, size, total_records: ordered.length },
      result: cloneMock(ordered.slice((page - 1) * size, page * size)),
    };
  }

  const response = await api.groups.ocservGroupsGet({
    authorization: requireAuthorizationHeader(),
    ...options,
  });
  return response.data;
}

export async function getOcservGroupNames(): Promise<string[]> {
  if (isTestMode)
    return currentMockGroups
      .map(({ name }) => name)
      .sort((a, b) => a.localeCompare(b));

  const response = await api.groups.ocservGroupsLookupGet({
    authorization: requireAuthorizationHeader(),
  });
  return response.data;
}

export async function getOcservGroup(id: number): Promise<OcservGroup> {
  if (isTestMode) return cloneMock(mockGroup(id));

  const response = await api.groups.ocservGroupsIdGet({
    authorization: requireAuthorizationHeader(),
    id,
  });
  return response.data;
}

export async function createOcservGroup(
  request: OcservGroupCreate,
): Promise<OcservGroup> {
  if (isTestMode) {
    if (currentMockGroups.some(({ name }) => name === request.name))
      throw new ApiError("Ocserv group name already exists.", { status: 400 });
    const group: ModelsOcservGroup = {
      config: cloneMock(request.config),
      id: Math.max(0, ...currentMockGroups.map(({ id }) => id ?? 0)) + 1,
      name: request.name,
    };
    currentMockGroups = [...currentMockGroups, group];
    return cloneMock(group);
  }

  const response = await api.groups.ocservGroupsPost({
    authorization: requireAuthorizationHeader(),
    request,
  });
  return response.data;
}

export async function updateOcservGroup(
  id: number,
  request: OcservGroupUpdate,
): Promise<OcservGroup> {
  if (isTestMode) {
    const group = mockGroup(id);
    const updated = { ...group, config: cloneMock(request.config) };
    currentMockGroups = currentMockGroups.map((item) =>
      item.id === id ? updated : item,
    );
    return cloneMock(updated);
  }

  const response = await api.groups.ocservGroupsIdPatch({
    authorization: requireAuthorizationHeader(),
    id,
    request,
  });
  return response.data;
}

export async function deleteOcservGroup(id: number): Promise<void> {
  if (isTestMode) {
    mockGroup(id);
    currentMockGroups = currentMockGroups.filter((item) => item.id !== id);
    return;
  }

  await api.groups.ocservGroupsIdDelete({
    authorization: requireAuthorizationHeader(),
    id,
  });
}

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

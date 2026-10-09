import { requireAuthorizationHeader } from "@/api/auth-token";
import { api } from "@/api/client";
import type {
  GroupUnsyncedGroup,
  ModelsExpiryMode,
  ModelsOcservUserConfig,
  ModelsTrafficType,
  RequestMeta,
} from "@/api/generated";
import { httpClient } from "@/api/http";
import { isTestMode } from "@/api/environment";

export type UnsyncedGroup = GroupUnsyncedGroup;

export interface UnsyncedUser {
  username: string;
  group: string;
}

export interface UnsyncedUsersResponse {
  meta: RequestMeta;
  result?: UnsyncedUser[];
}

export interface SyncUsersRequest {
  users: UnsyncedUser[];
  expire_at?: string;
  expiry_mode?: ModelsExpiryMode;
  expire_days_after_first_connection?: number;
  traffic_type: ModelsTrafficType;
  traffic_size: number;
  description?: string;
  config?: ModelsOcservUserConfig;
}

export async function getUnsyncedUsers(
  page = 1,
  size = 100,
): Promise<UnsyncedUsersResponse> {
  if (isTestMode) {
    const mock = await import("@/mocks/ocserv-sync");
    return mock.getMockUnsyncedUsers(page, size);
  }
  const response = await httpClient.get<UnsyncedUsersResponse>(
    "/ocserv/users/ocpasswd",
    { params: { page, size } },
  );
  return response.data;
}

export async function syncUsers(request: SyncUsersRequest): Promise<string[]> {
  if (isTestMode) {
    const mock = await import("@/mocks/ocserv-sync");
    return mock.syncMockUsers(request);
  }
  const response = await httpClient.post<string[]>(
    "/ocserv/users/ocpasswd/sync",
    request,
  );
  return response.data;
}

export async function getUnsyncedGroups(): Promise<UnsyncedGroup[]> {
  if (isTestMode) {
    const mock = await import("@/mocks/ocserv-sync");
    return mock.getMockUnsyncedGroups();
  }
  const response = await api.unsyncedGroups.ocservGroupsUnsyncedGet({
    authorization: requireAuthorizationHeader(),
  });
  return response.data ?? [];
}

export async function syncGroups(groups: UnsyncedGroup[]): Promise<string[]> {
  if (isTestMode) {
    const mock = await import("@/mocks/ocserv-sync");
    return mock.syncMockGroups(groups);
  }
  const response = await api.unsyncedGroups.ocservGroupsSyncPost({
    authorization: requireAuthorizationHeader(),
    request: { groups },
  });
  return response.data;
}

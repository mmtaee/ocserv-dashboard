import { requireAuthorizationHeader } from "@/api/auth-token";
import { api } from "@/api/client";
import { isTestMode } from "@/api/environment";
import type {
  ModelsDailyTraffic,
  ModelsOcservUser,
  ModelsOcservUserSessionLog,
  OcservUserActivateUserData,
  OcservUserBulkDeleteResponse,
  OcservUserBulkGroupRequest,
  OcservUserBulkStatusRequest,
  OcservUserBulkUpdateRequest,
  OcservUserBulkUsersResponse,
  OcservUserCreateOcservUserData,
  OcservUserOcservUsersResponse,
  OcservUserUpdateOcservUserData,
  RepositoryTotalBandwidths,
  RequestMeta,
} from "@/api/generated";
import { ApiError, httpClient } from "@/api/http";
import { cloneMock, mockOcservUsers } from "@/mocks";

export type OcservUser = ModelsOcservUser;
export type OcservUserCreate = OcservUserCreateOcservUserData;
export type OcservUserUpdate = OcservUserUpdateOcservUserData;
export type OcservUserActivation = OcservUserActivateUserData;
export type OcservUsersList = OcservUserOcservUsersResponse;
export type OcservUserBulkUpdate = OcservUserBulkUpdateRequest;
export type OcservUserBulkStatus = OcservUserBulkStatusRequest;
export type OcservUserBulkGroup = OcservUserBulkGroupRequest;

export interface OcservUsersListOptions {
  expireInDays?: number;
  filter?: "active" | "deactivated" | "locked" | "online";
  group?: string;
  page?: number;
  q?: string;
  size?: number;
}

export interface OcservUserSessionLogsResponse {
  meta: RequestMeta;
  result?: ModelsOcservUserSessionLog[];
}

export interface OcservUserStatisticsResponse {
  statistics: ModelsDailyTraffic[];
  total_bandwidths: RepositoryTotalBandwidths;
}

export interface DateRangeOptions {
  dateEnd?: string;
  dateStart?: string;
}

let currentMockUsers: OcservUser[] = cloneMock(mockOcservUsers);

function mockUser(id: number): OcservUser {
  const user = currentMockUsers.find((item) => item.id === id);
  if (!user) throw new ApiError("Ocserv user not found.", { status: 400 });
  return user;
}

function authorizationOptions() {
  return { headers: { Authorization: requireAuthorizationHeader() } };
}

export async function getOcservUsers(
  options: OcservUsersListOptions = {},
): Promise<OcservUsersList> {
  if (isTestMode) {
    const page = options.page ?? 1;
    const size = options.size ?? 25;
    const query = options.q?.trim().toLocaleLowerCase() ?? "";
    const filtered = currentMockUsers.filter((user) => {
      if (query && !user.username.toLocaleLowerCase().includes(query))
        return false;
      if (options.group && user.group !== options.group) return false;
      if (options.filter === "online" && !user.is_online) return false;
      if (options.filter === "locked" && !user.is_locked) return false;
      if (options.filter === "deactivated" && !user.deactivated_at)
        return false;
      if (options.filter === "active" && user.deactivated_at) return false;
      return true;
    });
    return {
      meta: { page, size, total_records: filtered.length },
      result: cloneMock(filtered.slice((page - 1) * size, page * size)),
    };
  }

  const response = await api.users.ocservUsersGet(
    { expireInDays: options.expireInDays },
    {
      ...authorizationOptions(),
      params: {
        filter: options.filter,
        group: options.group,
        page: options.page,
        q: options.q,
        size: options.size,
      },
    },
  );
  return response.data;
}

export async function getOcservUser(id: number): Promise<OcservUser> {
  if (isTestMode) return cloneMock(mockUser(id));
  const response = await api.users.ocservUsersIdGet(
    { id },
    authorizationOptions(),
  );
  return response.data as unknown as OcservUser;
}

export async function createOcservUser(
  request: OcservUserCreate,
): Promise<OcservUser> {
  if (isTestMode) {
    if (currentMockUsers.some(({ username }) => username === request.username))
      throw new ApiError("Ocserv username already exists.", { status: 400 });
    const user: OcservUser = {
      certificate_available: false,
      certificate_enabled: false,
      config: cloneMock(request.config),
      created_at: new Date().toISOString(),
      description: request.description,
      expire_at: request.expire_at,
      expire_days_after_first_connection:
        request.expire_days_after_first_connection,
      expiry_mode: request.expiry_mode ?? "unlimited",
      group: request.group,
      id: Math.max(0, ...currentMockUsers.map(({ id }) => id)) + 1,
      is_locked: false,
      is_online: false,
      online_sessions: [],
      owner_id: 1,
      password: "••••••••",
      running_rx: 0,
      running_tx: 0,
      traffic_size: request.traffic_size ?? 0,
      traffic_type: request.traffic_type,
      username: request.username,
    };
    currentMockUsers = [...currentMockUsers, user];
    return cloneMock(user);
  }
  const response = await api.users.ocservUsersPost({
    authorization: requireAuthorizationHeader(),
    request,
  });
  return response.data;
}

export async function updateOcservUser(
  id: number,
  request: OcservUserUpdate,
): Promise<OcservUser> {
  if (isTestMode) {
    const updated = {
      ...mockUser(id),
      ...cloneMock(request),
      id,
    } as OcservUser;
    currentMockUsers = currentMockUsers.map((user) =>
      user.id === id ? updated : user,
    );
    return cloneMock(updated);
  }
  const response = await api.users.ocservUsersIdPatch({
    authorization: requireAuthorizationHeader(),
    id,
    request,
  });
  return response.data;
}

export async function deleteOcservUser(id: number): Promise<void> {
  if (isTestMode) {
    mockUser(id);
    currentMockUsers = currentMockUsers.filter((user) => user.id !== id);
    return;
  }
  await api.users.ocservUsersIdDelete({ id }, authorizationOptions());
}

export async function bulkUpdateOcservUsers(
  request: OcservUserBulkUpdate,
): Promise<OcservUserBulkUsersResponse> {
  if (isTestMode) {
    const users = request.users.map(({ changes, id }) => {
      const updated = {
        ...mockUser(id),
        ...cloneMock(changes),
        id,
      } as OcservUser;
      currentMockUsers = currentMockUsers.map((user) =>
        user.id === id ? updated : user,
      );
      return updated;
    });
    return { count: users.length, users: cloneMock(users) };
  }
  const response = await api.users.ocservUsersBulkPatch({
    authorization: requireAuthorizationHeader(),
    request,
  });
  return response.data;
}

export async function bulkDeleteOcservUsers(
  ids: number[],
): Promise<OcservUserBulkDeleteResponse> {
  if (isTestMode) {
    const before = currentMockUsers.length;
    currentMockUsers = currentMockUsers.filter(
      (user) => !ids.includes(user.id),
    );
    return { count: before - currentMockUsers.length };
  }
  const response = await api.users.ocservUsersBulkDelete({
    authorization: requireAuthorizationHeader(),
    request: { ids },
  });
  return response.data;
}

export async function bulkSetOcservUserStatus(
  request: OcservUserBulkStatus,
): Promise<OcservUserBulkUsersResponse> {
  if (isTestMode) {
    const users = currentMockUsers
      .filter(({ id }) => request.ids.includes(id))
      .map((user) => ({
        ...user,
        deactivated_at: request.enabled ? undefined : new Date().toISOString(),
      }));
    const byId = new Map(users.map((user) => [user.id, user]));
    currentMockUsers = currentMockUsers.map(
      (user) => byId.get(user.id) ?? user,
    );
    return { count: users.length, users: cloneMock(users) };
  }
  const response = await api.users.ocservUsersBulkStatusPatch({
    authorization: requireAuthorizationHeader(),
    request,
  });
  return response.data;
}

export async function bulkSetOcservUserGroup(
  request: OcservUserBulkGroup,
): Promise<OcservUserBulkUsersResponse> {
  if (isTestMode) {
    const users = currentMockUsers
      .filter(({ id }) => request.ids.includes(id))
      .map((user) => ({ ...user, group: request.group ?? "" }));
    const byId = new Map(users.map((user) => [user.id, user]));
    currentMockUsers = currentMockUsers.map(
      (user) => byId.get(user.id) ?? user,
    );
    return { count: users.length, users: cloneMock(users) };
  }
  const response = await api.users.ocservUsersBulkGroupPatch({
    authorization: requireAuthorizationHeader(),
    request,
  });
  return response.data;
}

export async function activateOcservUser(
  id: number,
  request: OcservUserActivation,
): Promise<void> {
  if (isTestMode) {
    const updated = {
      ...mockUser(id),
      deactivated_at: undefined,
      expire_at: request.expire_at,
      expire_days_after_first_connection:
        request.expire_days_after_first_connection,
      expiry_mode: request.expiry_mode ?? "unlimited",
    };
    currentMockUsers = currentMockUsers.map((user) =>
      user.id === id ? updated : user,
    );
    return;
  }
  await api.users.ocservUsersIdActivatePost({
    authorization: requireAuthorizationHeader(),
    id,
    request,
  });
}

export async function lockOcservUser(id: number): Promise<void> {
  if (isTestMode) {
    mockUser(id).is_locked = true;
    return;
  }
  await api.users.ocservUsersIdLockPost({ id }, authorizationOptions());
}

export async function unlockOcservUser(id: number): Promise<void> {
  if (isTestMode) {
    mockUser(id).is_locked = false;
    return;
  }
  await api.users.ocservUsersIdUnlockPost({ id }, authorizationOptions());
}

export async function resetOcservUserUsage(id: number): Promise<void> {
  if (isTestMode) {
    const user = mockUser(id);
    user.running_rx = 0;
    user.running_tx = 0;
    return;
  }
  await api.users.ocservUsersIdResetUsagePost({
    authorization: requireAuthorizationHeader(),
    id,
  });
}

export async function createOcservUserCertificate(id: number): Promise<void> {
  if (isTestMode) {
    const user = mockUser(id);
    user.certificate_available = true;
    user.certificate_enabled = true;
    return;
  }
  await api.users.ocservUsersIdCertificatePost({ id }, authorizationOptions());
}

export async function downloadOcservUserCertificate(
  id: number,
  username: string,
): Promise<void> {
  if (isTestMode) return;
  const response = await api.users.ocservUsersIdCertificateGet(
    { id },
    { ...authorizationOptions(), responseType: "blob" },
  );
  const url = URL.createObjectURL(response.data as unknown as Blob);
  const anchor = document.createElement("a");
  anchor.href = url;
  anchor.download = `${username}.p12`;
  anchor.click();
  URL.revokeObjectURL(url);
}

export async function disconnectOcservUser(username: string): Promise<void> {
  if (isTestMode) {
    const user = currentMockUsers.find((item) => item.username === username);
    if (user) {
      user.is_online = false;
      user.online_sessions = [];
    }
    return;
  }
  await httpClient.post(
    `/ocserv/users/${encodeURIComponent(username)}/disconnect`,
    undefined,
    authorizationOptions(),
  );
}

export async function terminateOcservUser(username: string): Promise<void> {
  if (isTestMode) return disconnectOcservUser(username);
  await httpClient.post(
    `/ocserv/users/${encodeURIComponent(username)}/terminate`,
    undefined,
    authorizationOptions(),
  );
}

export async function disconnectOcservUserSession(id: number): Promise<void> {
  if (isTestMode) {
    currentMockUsers.forEach((user) => {
      user.online_sessions = user.online_sessions.filter(
        (session) => session.ID !== id,
      );
      user.is_online = user.online_sessions.length > 0;
    });
    return;
  }
  await httpClient.post(
    `/ocserv/users/${id}/disconnect_by_id`,
    undefined,
    authorizationOptions(),
  );
}

export async function terminateOcservUserSession(id: number): Promise<void> {
  if (isTestMode) return disconnectOcservUserSession(id);
  await httpClient.post(
    `/ocserv/users/${id}/terminate_by_id`,
    undefined,
    authorizationOptions(),
  );
}

export async function getOcservUserSessionLogs(
  id: number,
  options: DateRangeOptions & { page?: number; size?: number } = {},
): Promise<OcservUserSessionLogsResponse> {
  if (isTestMode) {
    const username = mockUser(id).username;
    return {
      meta: {
        page: options.page ?? 1,
        size: options.size ?? 10,
        total_records: 2,
      },
      result: [
        {
          created_at: "2026-08-31T06:20:00Z",
          event: "handshake",
          ip: "203.0.113.24",
          message: "VPN session established",
          username,
        },
        {
          created_at: "2026-08-30T17:45:00Z",
          event: "disconnect",
          ip: "203.0.113.24",
          message: "VPN session disconnected",
          username,
        },
      ],
    };
  }
  const response = await api.users.ocservUsersIdSessionLogsGet(
    { id },
    {
      ...authorizationOptions(),
      params: {
        date_end: options.dateEnd,
        date_start: options.dateStart,
        page: options.page,
        size: options.size,
      },
    },
  );
  return response.data as unknown as OcservUserSessionLogsResponse;
}

export async function getOcservUserStatistics(
  id: number,
  options: DateRangeOptions = {},
): Promise<OcservUserStatisticsResponse> {
  if (isTestMode) {
    return {
      statistics: [
        { date: "2026-08-30", rx: 1.8, tx: 0.6 },
        { date: "2026-08-31", rx: 2.4, tx: 0.9 },
      ],
      total_bandwidths: { rx: 4.2, tx: 1.5 },
    };
  }
  const response = await api.users.ocservUsersIdStatisticsGet(
    { id },
    {
      ...authorizationOptions(),
      params: { date_end: options.dateEnd, date_start: options.dateStart },
    },
  );
  return response.data as unknown as OcservUserStatisticsResponse;
}

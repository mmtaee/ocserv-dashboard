import { requireAuthorizationHeader } from "@/api/auth-token";
import { api } from "@/api/client";
import { isTestMode } from "@/api/environment";
import type {
  GithubComMmtaeeOcservDashboardBackendInternalServicesAdminApiSystemChangeUserPassword,
  GithubComMmtaeeOcservDashboardBackendInternalServicesAdminApiSystemCreateUserData,
  GithubComMmtaeeOcservDashboardBackendInternalServicesAdminApiSystemUsersResponse,
  ModelsUser,
} from "@/api/generated";
import { ApiError } from "@/api/http";

export type Staff = ModelsUser;
export type StaffCreate =
  GithubComMmtaeeOcservDashboardBackendInternalServicesAdminApiSystemCreateUserData;
export type StaffPassword =
  GithubComMmtaeeOcservDashboardBackendInternalServicesAdminApiSystemChangeUserPassword;
export type StaffsList =
  GithubComMmtaeeOcservDashboardBackendInternalServicesAdminApiSystemUsersResponse;

let mockStaffs: Staff[] = [
  {
    created_at: "2026-01-15T09:00:00Z",
    id: 2,
    last_login: "2026-09-10T16:45:00Z",
    superadmin: false,
    updated_at: "2026-09-10T16:45:00Z",
    username: "operator",
  },
];

export async function getStaffs(page = 1, size = 25): Promise<StaffsList> {
  if (isTestMode) {
    return {
      meta: { page, size, total_records: mockStaffs.length },
      result: structuredClone(mockStaffs.slice((page - 1) * size, page * size)),
    };
  }
  const response = await api.systemUsers.systemUsersGet(
    { authorization: requireAuthorizationHeader() },
    { params: { page, size } },
  );
  return response.data;
}

export async function createStaff(request: StaffCreate): Promise<Staff> {
  if (isTestMode) {
    if (mockStaffs.some(({ username }) => username === request.username))
      throw new ApiError("Username already exists.", { status: 400 });
    const now = new Date().toISOString();
    const staff: Staff = {
      created_at: now,
      id: Math.max(0, ...mockStaffs.map(({ id }) => id)) + 1,
      last_login: "",
      superadmin: false,
      updated_at: now,
      username: request.username,
    };
    mockStaffs = [...mockStaffs, staff];
    return structuredClone(staff);
  }
  const response = await api.systemUsers.systemUsersPost({
    authorization: requireAuthorizationHeader(),
    request,
  });
  return response.data;
}

export async function changeStaffPassword(
  id: number,
  request: StaffPassword,
): Promise<void> {
  if (isTestMode) {
    if (!mockStaffs.some((staff) => staff.id === id))
      throw new ApiError("Staff not found.", { status: 400 });
    return;
  }
  await api.systemUsers.systemUsersIdPasswordPost({
    authorization: requireAuthorizationHeader(),
    id,
    request,
  });
}

export async function deleteStaff(id: number): Promise<void> {
  if (isTestMode) {
    if (!mockStaffs.some((staff) => staff.id === id))
      throw new ApiError("Staff not found.", { status: 400 });
    mockStaffs = mockStaffs.filter((staff) => staff.id !== id);
    return;
  }
  await api.systemUsers.systemUsersIdDelete({
    authorization: requireAuthorizationHeader(),
    id,
  });
}

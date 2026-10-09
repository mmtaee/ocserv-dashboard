import { ApiError } from "@/api/http";
import type {
  SyncUsersRequest,
  UnsyncedGroup,
  UnsyncedUser,
  UnsyncedUsersResponse,
} from "@/api/services/ocserv-sync";
import { cloneMock } from "@/mocks/utils";

type MockScenario = "default" | "empty" | "error";

const scenario = (import.meta.env.VITE_SYNC_MOCK_SCENARIO ||
  "default") as MockScenario;

const initialUsers = [
  { group: "engineering", username: "mira" },
  { group: "employees", username: "sam" },
  { group: "contractors", username: "nora" },
  { group: "", username: "reza" },
] satisfies UnsyncedUser[];

const initialGroups = [
  {
    config: {
      dns: ["10.40.0.2"],
      "ipv4-network": "10.40.0.0/24",
      route: ["10.0.0.0/8"],
    },
    name: "partners",
    path: "/etc/ocserv/config-per-group/partners",
  },
  {
    config: {},
    name: "temporary",
    path: "/etc/ocserv/config-per-group/temporary",
  },
] satisfies UnsyncedGroup[];

let users = scenario === "empty" ? [] : cloneMock(initialUsers);
let groups = scenario === "empty" ? [] : cloneMock(initialGroups);

function failIfRequested(): void {
  if (scenario === "error")
    throw new ApiError("Mock backend rejected the sync request.", {
      status: 400,
      data: { error: "Mock backend rejected the sync request." },
    });
}

export function getMockUnsyncedUsers(
  page: number,
  size: number,
): UnsyncedUsersResponse {
  failIfRequested();
  const start = (page - 1) * size;
  return {
    meta: { page, size, total_records: users.length },
    result: cloneMock(users.slice(start, start + size)),
  };
}

export function syncMockUsers(request: SyncUsersRequest): string[] {
  failIfRequested();
  const names = request.users.map(({ username }) => username);
  users = users.filter(({ username }) => !names.includes(username));
  return cloneMock(names);
}

export function getMockUnsyncedGroups(): UnsyncedGroup[] {
  failIfRequested();
  return cloneMock(groups);
}

export function syncMockGroups(request: UnsyncedGroup[]): string[] {
  failIfRequested();
  const names = request.map(({ name }) => name);
  groups = groups.filter(({ name }) => !names.includes(name));
  return cloneMock(names);
}

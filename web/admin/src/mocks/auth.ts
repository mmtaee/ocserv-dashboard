import type {
  GithubComMmtaeeOcservDashboardBackendInternalServicesAdminApiSystemUserLoginResponse,
  ModelsUser,
} from "@/api/generated";

export const mockCurrentUser = {
  id: 1,
  username: "test-superadmin",
  superadmin: true,
  last_login: "2026-08-30T08:00:00Z",
  created_at: "2026-01-01T00:00:00Z",
  updated_at: "2026-08-30T08:00:00Z",
} satisfies ModelsUser;

export const mockLoginResponse = {
  token: "mock-test-access-token",
  user: mockCurrentUser,
} satisfies GithubComMmtaeeOcservDashboardBackendInternalServicesAdminApiSystemUserLoginResponse;

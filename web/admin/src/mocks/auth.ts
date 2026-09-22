import type {
  GithubComMmtaeeOcservDashboardBackendInternalServicesAdminApiSystemUserLoginResponse,
  ModelsUser,
} from "@/api/generated";
import { ApiError } from "@/api/http";
import type {
  ResetAdminPasswordRequest,
  ResetAdminPasswordResponse,
} from "@/api/services/auth";
import { cloneMock } from "@/mocks/utils";

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

export const mockResetAdminPasswordResponse = {
  token: "mock-reset-password-token",
  user: mockCurrentUser,
} satisfies ResetAdminPasswordResponse;

const MOCK_RESET_SECRET_KEY = "mock-reset-secret-key-2026";
const MOCK_EXPIRED_SECRET_KEY = "mock-expired-secret-key-2026";

export async function resetMockAdminPassword(
  request: ResetAdminPasswordRequest,
): Promise<ResetAdminPasswordResponse> {
  if (request.new_password.length < 4 || request.new_password.length > 32) {
    throw new ApiError("Password must contain 4 to 32 characters.", {
      status: 400,
      data: {
        error: ["invalid_password"],
        message: ["Password must contain 4 to 32 characters."],
      },
    });
  }

  if (request.secret_key === MOCK_EXPIRED_SECRET_KEY) {
    throw new ApiError("The secret key has expired.", {
      status: 400,
      data: {
        error: ["expired_secret_key"],
        message: ["The secret key has expired."],
      },
    });
  }

  if (request.secret_key !== MOCK_RESET_SECRET_KEY) {
    throw new ApiError("The secret key is invalid.", {
      status: 400,
      data: {
        error: ["invalid_secret_key"],
        message: ["The secret key is invalid."],
      },
    });
  }

  if (request.new_password === "backend-error") {
    throw new ApiError("The password reset service is unavailable.", {
      status: 500,
      data: {
        error: ["reset_unavailable"],
        message: ["The password reset service is unavailable."],
      },
    });
  }

  return cloneMock(mockResetAdminPasswordResponse);
}

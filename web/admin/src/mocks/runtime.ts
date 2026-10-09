import { ApiError } from "@/api/http";
import type {
  OcservRuntimeAction,
  OcservRuntimeActionResult,
  OcservRuntimeStatusResponse,
} from "@/api/services/system";
import { cloneMock } from "@/mocks/utils";

const runtimeDetails = {
  active_state: "active",
  cpu_usage_nsec: 18_420_000_000,
  description: "OpenConnect SSL VPN server",
  id: "ocserv.service",
  main_pid: 1842,
  memory: 67_108_864,
  start_time: "2026-09-11T08:30:00Z",
  sub_state: "running",
  tasks: 6,
  unit_file_state: "enabled",
} satisfies OcservRuntimeStatusResponse;

export const mockRuntimeStatus = {
  ...runtimeDetails,
  allowed_actions: ["restart", "disable"],
} satisfies OcservRuntimeStatusResponse;

export const mockRuntimeSingleActionStatus = {
  ...runtimeDetails,
  active_state: "inactive",
  sub_state: "dead",
  unit_file_state: "disabled",
  allowed_actions: ["enable"],
} satisfies OcservRuntimeStatusResponse;

export const mockRuntimeNoActionsStatus = {
  ...runtimeDetails,
  allowed_actions: [],
} satisfies OcservRuntimeStatusResponse;

export const mockEmptyRuntimeStatus = {} satisfies OcservRuntimeStatusResponse;

export const mockRuntimeActionResponses = {
  enable: { message: "Ocserv runtime enabled." },
  disable: { message: "Ocserv runtime disabled." },
  restart: { message: "Ocserv runtime restarted." },
} satisfies Record<OcservRuntimeAction, OcservRuntimeActionResult>;

function mockScenario(): string {
  return import.meta.env.VITE_RUNTIME_MOCK_SCENARIO || "normal";
}

function waitForMockRuntime(): Promise<void> {
  return new Promise((resolve) => setTimeout(resolve, 200));
}

let mutableRuntimeStatus: OcservRuntimeStatusResponse =
  cloneMock(mockRuntimeStatus);

export async function getMockRuntimeStatus(): Promise<OcservRuntimeStatusResponse> {
  await waitForMockRuntime();
  if (mockScenario() === "error") {
    throw new ApiError("Unable to read the Ocserv runtime status.", {
      status: 503,
    });
  }

  if (mockScenario() === "empty") return cloneMock(mockEmptyRuntimeStatus);
  if (mockScenario() === "single")
    return cloneMock(mockRuntimeSingleActionStatus);
  if (mockScenario() === "no-actions")
    return cloneMock(mockRuntimeNoActionsStatus);
  return cloneMock(mutableRuntimeStatus);
}

export async function runMockRuntimeAction(
  action: OcservRuntimeAction,
): Promise<OcservRuntimeActionResult> {
  await waitForMockRuntime();
  if (mockScenario() === "error") {
    throw new ApiError("The Ocserv runtime action failed.", { status: 500 });
  }

  const availableActions: readonly OcservRuntimeAction[] | null | undefined =
    mockScenario() === "single"
      ? mockRuntimeSingleActionStatus.allowed_actions
      : mockScenario() === "no-actions" || mockScenario() === "empty"
        ? []
        : mutableRuntimeStatus.allowed_actions;

  if (!availableActions?.includes(action)) {
    throw new ApiError("This runtime action is not available.", {
      status: 400,
    });
  }

  if (action === "disable") {
    mutableRuntimeStatus = cloneMock(mockRuntimeSingleActionStatus);
  } else if (action === "enable") {
    mutableRuntimeStatus = cloneMock(mockRuntimeStatus);
  }

  return cloneMock(mockRuntimeActionResponses[action]);
}

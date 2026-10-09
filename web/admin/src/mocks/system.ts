import type {
  GithubComMmtaeeOcservDashboardBackendInternalServicesAdminApiSystemGetSystemInitResponse,
  GithubComMmtaeeOcservDashboardBackendInternalServicesAdminApiSystemGetSystemResponse,
  GithubComMmtaeeOcservDashboardBackendInternalServicesAdminApiSystemDashboardRelease,
  GithubComMmtaeeOcservDashboardBackendInternalServicesAdminApiSystemPatchSystemUpdateData,
  ModelsOcservAgent,
  OcservAgentCreateInput,
  OcservAgentUpdateInput,
  RuntimeOcservConfig,
} from "@/api/generated";
import { ApiError } from "@/api/http";
import type { OcservConfigResponse } from "@/api/services/system";
import { cloneMock } from "@/mocks/utils";

export const mockSystemInit = {
  first_init: true,
  google_captcha_site_key: "",
  telegram_bot_enabled: true,
} satisfies GithubComMmtaeeOcservDashboardBackendInternalServicesAdminApiSystemGetSystemInitResponse;

export const mockSystemConfig = {
  auto_delete_inactive_users: true,
  client_profile_connection_name: "Ocserv Test VPN",
  client_profile_server_address: "vpn.test.local",
  client_profile_server_port: 443,
  first_init: true,
  google_captcha_secret_key: "",
  google_captcha_site_key: "",
  keep_inactive_user_days: 30,
} satisfies GithubComMmtaeeOcservDashboardBackendInternalServicesAdminApiSystemGetSystemResponse;

let systemConfig: GithubComMmtaeeOcservDashboardBackendInternalServicesAdminApiSystemGetSystemResponse =
  cloneMock(mockSystemConfig);
let ocservConfig: RuntimeOcservConfig = {
  auth_timeout: 240,
  dns: ["1.1.1.1", "8.8.8.8"],
  ipv4_network: "10.10.0.0/24",
  max_clients: 128,
  mtu: 1420,
  tcp_port: 443,
  udp_port: 443,
};
type MockAgent = ModelsOcservAgent;

let agents: MockAgent[] = [
  {
    id: 1,
    name: "Primary agent",
    address: "vpn.example.test",
    address_type: "domain",
    token: "agent-token",
    port: 8080,
    created_at: "2026-01-01T00:00:00Z",
  },
  {
    id: 2,
    name: "Backup agent",
    address: "192.0.2.20",
    address_type: "ip",
    token: "backup-agent-token",
    port: 8081,
  },
  {
    id: 3,
    name: "Edge agent",
    address: "edge.example.test",
    address_type: "domain",
    token: "edge-agent-token",
    port: 8080,
  },
];

function scenario(): string {
  return import.meta.env.VITE_SYSTEM_MOCK_SCENARIO || "normal";
}
async function wait(): Promise<void> {
  await new Promise((resolve) => setTimeout(resolve, 200));
  if (scenario() === "error")
    throw new ApiError("System service is unavailable.", { status: 500 });
}

export async function updateMockSystemConfig(
  request: GithubComMmtaeeOcservDashboardBackendInternalServicesAdminApiSystemPatchSystemUpdateData,
) {
  await wait();
  systemConfig = { ...systemConfig, ...request };
  return cloneMock(systemConfig);
}

export async function getMockSystemConfig(): Promise<GithubComMmtaeeOcservDashboardBackendInternalServicesAdminApiSystemGetSystemResponse> {
  await wait();
  return cloneMock(systemConfig);
}
export async function getMockOcservConfig(): Promise<OcservConfigResponse> {
  await wait();
  return {
    ...cloneMock(ocservConfig),
    allow_update: scenario() !== "read-only",
  };
}
export async function updateMockOcservConfig(
  request: RuntimeOcservConfig,
): Promise<RuntimeOcservConfig> {
  await wait();
  if (scenario() === "read-only")
    throw new ApiError("Ocserv configuration updates are disabled.", {
      status: 403,
    });
  ocservConfig = { ...ocservConfig, ...request };
  return cloneMock(ocservConfig);
}
export async function getMockRelease(): Promise<GithubComMmtaeeOcservDashboardBackendInternalServicesAdminApiSystemDashboardRelease> {
  await wait();
  return scenario() === "up-to-date"
    ? { current: "v1.0.0", latest: "v1.0.0" }
    : { current: "v1.0.0", latest: "v1.1.0" };
}
export async function getMockAgents(): Promise<ModelsOcservAgent[]> {
  await wait();
  if (scenario() === "agent-list-error")
    throw new ApiError("Agent list is unavailable.", { status: 503 });
  if (["empty", "no-agents"].includes(scenario())) return [];
  if (scenario() === "multiple-agents") return cloneMock(agents);
  if (scenario() === "agent-unavailable") return cloneMock(agents.slice(0, 1));
  if (scenario() === "selected-agent-removed")
    return cloneMock(agents.slice(1));
  if (scenario() === "malformed-agent-address")
    return cloneMock([{ ...agents[0], address: "" }]);
  return cloneMock(agents.slice(0, 1));
}
export async function mutateMockAgent(
  action: "create" | "update" | "delete",
  request?: OcservAgentCreateInput | OcservAgentUpdateInput,
  id?: number,
): Promise<ModelsOcservAgent | void> {
  await wait();
  if (
    request &&
    (!request.name.trim() || !request.address.trim() || !request.token.trim())
  )
    throw new ApiError("Name, address, and token are required.", {
      status: 400,
    });
  if (action === "create") {
    const agent = {
      ...request!,
      id: Math.max(0, ...agents.map(({ id }) => id ?? 0)) + 1,
    };
    agents = [...agents, agent];
    return cloneMock(agent);
  }
  const index = agents.findIndex((agent) => agent.id === id);
  if (index < 0) throw new ApiError("Agent not found.", { status: 404 });
  if (action === "delete") {
    agents = agents.filter((agent) => agent.id !== id);
    return;
  }
  agents[index] = {
    ...agents[index],
    ...request,
    updated_at: "2026-01-02T00:00:00Z",
  };
  return cloneMock(agents[index]);
}

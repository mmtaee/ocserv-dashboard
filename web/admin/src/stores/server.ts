import { defineStore } from "pinia";
import { computed, shallowRef } from "vue";

import { apiBaseUrl, setApiBaseUrl } from "@/api/http";
import { getOcservAgents, type OcservAgent } from "@/api/services/system";

export const MASTER_SERVER = "master";
export type ServerError = "agentList" | "invalidAddress" | "unavailable";

export function agentServerId(agent: OcservAgent): string {
  return agent.id == null ? `address:${agent.address}` : String(agent.id);
}

function validIpv4(value: string): boolean {
  const parts = value.split(".");
  return (
    parts.length === 4 &&
    parts.every((part) => /^\d{1,3}$/.test(part) && Number(part) <= 255)
  );
}

export function agentApiBaseUrl(agent: OcservAgent): string | null {
  const address = agent.address.trim();
  const port = agent.port ?? 8080;
  if (!address || !Number.isInteger(port) || port < 1 || port > 65535)
    return null;

  if (agent.address_type === "ip") {
    if (!validIpv4(address) && !/^[0-9a-f:]+$/i.test(address)) return null;
    return `http://${address.includes(":") ? `[${address}]` : address}:${port}/api`;
  }

  if (
    agent.address_type !== "domain" ||
    !address.includes(".") ||
    !address
      .split(".")
      .every((label) => /^[a-z0-9](?:[a-z0-9-]{0,61}[a-z0-9])?$/i.test(label))
  )
    return null;

  return `http://${address}:${port}/api`;
}

export const useServerStore = defineStore("server", () => {
  const agents = shallowRef<OcservAgent[]>([]);
  const selectedServer = shallowRef<string>(MASTER_SERVER);
  const error = shallowRef<ServerError | null>(null);
  const isLoading = shallowRef(false);
  const revision = shallowRef(0);
  const selectableAgents = computed(() =>
    agents.value.filter((agent) => agentApiBaseUrl(agent) !== null),
  );
  const hasAgents = computed(() => selectableAgents.value.length > 0);
  let refreshPromise: Promise<void> | null = null;

  function useMaster(nextError: ServerError | null = null): void {
    const changed = selectedServer.value !== MASTER_SERVER;
    selectedServer.value = MASTER_SERVER;
    setApiBaseUrl(apiBaseUrl);
    error.value = nextError;
    if (changed) revision.value += 1;
  }

  function selectServer(value: string): void {
    if (value === MASTER_SERVER) return useMaster();
    const agent = agents.value.find((item) => agentServerId(item) === value);
    if (!agent) return useMaster("unavailable");
    const baseUrl = agentApiBaseUrl(agent);
    if (!baseUrl) return useMaster("invalidAddress");
    if (selectedServer.value === value) return;
    selectedServer.value = value;
    setApiBaseUrl(baseUrl);
    error.value = null;
    revision.value += 1;
  }

  function setAgents(nextAgents: OcservAgent[]): void {
    agents.value = nextAgents;
    if (selectedServer.value === MASTER_SERVER) return;
    const selected = selectableAgents.value.find(
      (agent) => agentServerId(agent) === selectedServer.value,
    );
    if (!selected) return useMaster("unavailable");
    const baseUrl = agentApiBaseUrl(selected);
    if (!baseUrl) return useMaster("invalidAddress");
    setApiBaseUrl(baseUrl);
    revision.value += 1;
  }

  function refreshAgents(): Promise<void> {
    if (refreshPromise) return refreshPromise;
    isLoading.value = true;
    error.value = null;
    refreshPromise = getOcservAgents()
      .then(setAgents)
      .catch(() => {
        error.value = "agentList";
        if (!agents.value.length) useMaster("agentList");
      })
      .finally(() => {
        isLoading.value = false;
        refreshPromise = null;
      });
    return refreshPromise;
  }

  function markUnavailable(): void {
    if (selectedServer.value !== MASTER_SERVER) error.value = "unavailable";
  }

  return {
    agents,
    error,
    hasAgents,
    isLoading,
    markUnavailable,
    refreshAgents,
    revision,
    selectableAgents,
    selectedServer,
    setAgents,
    selectServer,
  };
});

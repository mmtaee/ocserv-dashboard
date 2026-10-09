import { onBeforeUnmount, onMounted, readonly, shallowRef } from "vue";

import type {
  ContainerStats,
  DashboardOverview,
  OcservStats,
  SystemStats,
} from "@/api/services/dashboard";
import {
  getContainerStats,
  getDashboardOverview,
  getOcservStats,
  getSystemStats,
} from "@/api/services/dashboard";

const HOME_REFRESH_INTERVAL_MS = 10_000;
const USAGE_REFRESH_INTERVAL_MS = 10_000;
const OCSERV_REFRESH_INTERVAL_MS = 10_000;

export function useDashboardStats() {
  const homeOverview = shallowRef<DashboardOverview | null>(null);
  const systemStats = shallowRef<SystemStats | null>(null);
  const containerStats = shallowRef<ContainerStats | null>(null);
  const ocservStats = shallowRef<OcservStats | null>(null);
  const homeLoading = shallowRef(true);
  const systemLoading = shallowRef(true);
  const ocservLoading = shallowRef(true);
  const homeError = shallowRef(false);
  const systemError = shallowRef(false);
  const containerError = shallowRef(false);
  const ocservError = shallowRef(false);
  let isActive = true;
  let isRefreshingHome = false;
  let isRefreshingUsage = false;
  let isRefreshingOcserv = false;
  let homeRefreshTimer: ReturnType<typeof setInterval> | undefined;
  let usageRefreshTimer: ReturnType<typeof setInterval> | undefined;
  let ocservRefreshTimer: ReturnType<typeof setInterval> | undefined;

  async function loadHomeOverview(): Promise<void> {
    if (isRefreshingHome) return;
    isRefreshingHome = true;

    try {
      const overview = await getDashboardOverview();
      if (!isActive) return;

      homeOverview.value = overview;
      homeError.value = false;
    } catch {
      if (isActive) homeError.value = true;
    } finally {
      if (isActive) homeLoading.value = false;
      isRefreshingHome = false;
    }
  }

  async function loadUsageStats(): Promise<void> {
    if (isRefreshingUsage) return;
    isRefreshingUsage = true;

    try {
      const [systemResult, containerResult] = await Promise.allSettled([
        getSystemStats(),
        getContainerStats(),
      ]);
      if (!isActive) return;

      if (systemResult.status === "fulfilled") {
        systemStats.value = systemResult.value;
        systemError.value = false;
      } else {
        systemError.value = true;
      }

      if (containerResult.status === "fulfilled") {
        containerStats.value = containerResult.value;
        containerError.value = false;
      } else {
        containerError.value = true;
      }
    } finally {
      if (isActive) {
        systemLoading.value = false;
      }
      isRefreshingUsage = false;
    }
  }

  async function loadOcservStats(): Promise<void> {
    if (isRefreshingOcserv) return;
    isRefreshingOcserv = true;

    try {
      const stats = await getOcservStats();
      if (!isActive) return;

      ocservStats.value = stats;
      ocservError.value = false;
    } catch {
      if (isActive) ocservError.value = true;
    } finally {
      if (isActive) ocservLoading.value = false;
      isRefreshingOcserv = false;
    }
  }

  async function refresh(): Promise<void> {
    await Promise.allSettled([
      loadHomeOverview(),
      loadUsageStats(),
      loadOcservStats(),
    ]);
  }

  onMounted(() => {
    void loadHomeOverview();
    void loadUsageStats();
    void loadOcservStats();
    homeRefreshTimer = setInterval(
      () => void loadHomeOverview(),
      HOME_REFRESH_INTERVAL_MS,
    );
    usageRefreshTimer = setInterval(
      () => void loadUsageStats(),
      USAGE_REFRESH_INTERVAL_MS,
    );
    ocservRefreshTimer = setInterval(
      () => void loadOcservStats(),
      OCSERV_REFRESH_INTERVAL_MS,
    );
  });

  onBeforeUnmount(() => {
    isActive = false;
    if (homeRefreshTimer !== undefined) clearInterval(homeRefreshTimer);
    if (usageRefreshTimer !== undefined) clearInterval(usageRefreshTimer);
    if (ocservRefreshTimer !== undefined) clearInterval(ocservRefreshTimer);
  });

  return {
    homeOverview: readonly(homeOverview),
    systemStats: readonly(systemStats),
    containerStats: readonly(containerStats),
    ocservStats: readonly(ocservStats),
    homeLoading: readonly(homeLoading),
    systemLoading: readonly(systemLoading),
    ocservLoading: readonly(ocservLoading),
    homeError: readonly(homeError),
    systemError: readonly(systemError),
    containerError: readonly(containerError),
    ocservError: readonly(ocservError),
    refresh,
  };
}

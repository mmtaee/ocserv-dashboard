import { requireAuthorizationHeader } from "@/api/auth-token";
import { api } from "@/api/client";
import { isTestMode } from "@/api/environment";
import type {
  GithubComMmtaeeOcservDashboardBackendInternalServicesAdminApiDashboardGetHomeResponse,
  GithubComMmtaeeOcservDashboardBackendInternalServicesAdminApiDashboardOcservStatusResponse,
  GithubComMmtaeeOcservDashboardBackendInternalServicesAdminApiDashboardServerStatusResponse,
  GithubComMmtaeeOcservDashboardBackendInternalServicesAdminApiDashboardDockerStats,
  GithubComMmtaeeOcservDashboardBackendInternalUsecaseAdminApiDashboardTelegramServiceStatus,
  GithubComMmtaeeOcservDashboardBackendInternalUsecaseAdminApiDashboardGetHomeUser,
  ModelsDailyTraffic,
  ModelsIPBanPoints,
  ModelsOnlineUserSession,
  RepositoryTotalBandwidths,
  RepositoryTopBandwidthUsers,
} from "@/api/generated";
import {
  cloneMock,
  mockContainerStats,
  mockDashboardOverview,
  mockOcservStats,
  mockSystemStats,
} from "@/mocks";

export type ContainerStats =
  GithubComMmtaeeOcservDashboardBackendInternalServicesAdminApiDashboardDockerStats;
export type SystemStats =
  GithubComMmtaeeOcservDashboardBackendInternalServicesAdminApiDashboardServerStatusResponse;
export type OcservStats =
  GithubComMmtaeeOcservDashboardBackendInternalServicesAdminApiDashboardOcservStatusResponse;
export type DashboardOverview =
  GithubComMmtaeeOcservDashboardBackendInternalServicesAdminApiDashboardGetHomeResponse;
export type TelegramService =
  GithubComMmtaeeOcservDashboardBackendInternalUsecaseAdminApiDashboardTelegramServiceStatus;
export type DashboardUsers =
  GithubComMmtaeeOcservDashboardBackendInternalUsecaseAdminApiDashboardGetHomeUser;
export type DailyTraffic = ModelsDailyTraffic;
export type IpBan = ModelsIPBanPoints;
export type OnlineSession = ModelsOnlineUserSession;
export type TotalBandwidth = RepositoryTotalBandwidths;
export type TopBandwidthUsers = RepositoryTopBandwidthUsers;

function authorization() {
  return { authorization: requireAuthorizationHeader() };
}

export async function getDashboardOverview(): Promise<DashboardOverview> {
  if (isTestMode) {
    return cloneMock(mockDashboardOverview);
  }

  const response = await api.home.homeGet(authorization());
  return response.data;
}

export async function getContainerStats(): Promise<ContainerStats> {
  if (isTestMode) {
    return cloneMock(mockContainerStats);
  }

  const response = await api.home.homeContainerStatsGet(authorization());
  return response.data;
}

export async function getOcservStats(): Promise<OcservStats> {
  if (isTestMode) {
    return cloneMock(mockOcservStats);
  }

  const response = await api.home.homeOcservStatsGet(authorization());
  return response.data;
}

export async function getSystemStats(): Promise<SystemStats> {
  if (isTestMode) {
    return cloneMock(mockSystemStats);
  }

  const response = await api.home.homeSystemStatsGet(authorization());
  return response.data;
}

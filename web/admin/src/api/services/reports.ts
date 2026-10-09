import { requireAuthorizationHeader } from "@/api/auth-token";
import { api } from "@/api/client";
import { isTestMode } from "@/api/environment";
import type {
  ModelsDailyTraffic,
  ModelsOcservUserSessionLog,
  ReportsOcservUserReportResponse,
  ReportsSessionLogsResponse,
  RepositoryTotalBandwidths,
} from "@/api/generated";
import {
  getMockDailyTraffic,
  getMockSessionLogs,
  getMockTotalBandwidth,
  getMockUserSummary,
} from "@/mocks/reports";

export type DailyTraffic = ModelsDailyTraffic;
export type UserReportSummary = ReportsOcservUserReportResponse;
export type TotalBandwidth = RepositoryTotalBandwidths;
export type SessionLog = ModelsOcservUserSessionLog;
export type SessionLogsResponse = ReportsSessionLogsResponse;
export type ReportSort = "ASC" | "DESC";

export interface DateRange {
  dateStart: string;
  dateEnd: string;
}

export interface SessionLogQuery {
  page: number;
  size: number;
  order: "created_at" | "username" | "event" | "ip";
  sort: ReportSort;
  dateStart?: string;
  dateEnd?: string;
}

function authorization(): string {
  return requireAuthorizationHeader();
}

export async function getReportStatistics(
  range: DateRange,
): Promise<DailyTraffic[]> {
  if (isTestMode) return getMockDailyTraffic(range);
  const response = await api.reports.reportsStatisticsGet({
    authorization: authorization(),
    ...range,
  });
  return response.data ?? [];
}

export async function getReportUserSummary(): Promise<UserReportSummary> {
  if (isTestMode) return getMockUserSummary();
  const response = await api.reports.reportsUsersGet({
    authorization: authorization(),
  });
  return response.data;
}

export async function getReportTotalBandwidth(
  range: DateRange,
): Promise<TotalBandwidth> {
  if (isTestMode) return getMockTotalBandwidth(range);
  const response = await api.reports.reportsTotalBandwidthGet({
    authorization: authorization(),
    ...range,
  });
  return response.data;
}

export async function getReportSessionLogs(
  query: SessionLogQuery,
): Promise<SessionLogsResponse> {
  if (isTestMode) return getMockSessionLogs(query);
  const response = await api.reports.reportsSessionLogsGet({
    authorization: authorization(),
    ...query,
  });
  return response.data;
}

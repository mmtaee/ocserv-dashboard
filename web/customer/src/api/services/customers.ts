import type {
  ActivitiesQuery,
  ActivitiesResponse,
  Bandwidth,
  ChangePasswordInput,
  CiscoSetup,
  DailyTraffic,
  DateRangeQuery,
  LoginData,
  LoginResponse,
  OnlineUserSession,
  SummaryResponse,
} from "@/api/generated";
import { isTestMode } from "@/api/environment";
import { httpClient } from "@/api/http";
import { mockCustomers } from "@/mocks/customers";

const path = "/customers";
export async function login(request: LoginData): Promise<LoginResponse> {
  return isTestMode
    ? mockCustomers.login(request)
    : (await httpClient.post<LoginResponse>(path + "/login", request)).data;
}
export async function getSummary(): Promise<SummaryResponse> {
  return isTestMode
    ? mockCustomers.summary()
    : (await httpClient.get<SummaryResponse>(path + "/summary")).data;
}
export async function getSessions(): Promise<OnlineUserSession[]> {
  return isTestMode
    ? mockCustomers.sessions()
    : (await httpClient.get<OnlineUserSession[]>(path + "/sessions")).data;
}
export async function disconnectSessions(): Promise<void> {
  if (isTestMode) return mockCustomers.accept();
  await httpClient.post(path + "/disconnect_sessions");
}
export async function terminateSessions(): Promise<void> {
  if (isTestMode) return mockCustomers.accept();
  await httpClient.post(path + "/terminate_sessions");
}
export async function changePassword(
  request: ChangePasswordInput,
): Promise<void> {
  if (isTestMode) return mockCustomers.changePassword(request);
  await httpClient.post(path + "/password", request);
}
export async function getActivities(
  params: ActivitiesQuery,
): Promise<ActivitiesResponse> {
  return isTestMode
    ? mockCustomers.activities(params)
    : (
        await httpClient.get<ActivitiesResponse>(path + "/activities", {
          params,
        })
      ).data;
}
export async function getStats(
  params: DateRangeQuery,
): Promise<DailyTraffic[]> {
  if (isTestMode) return mockCustomers.stats(params);
  return (
    (await httpClient.get<DailyTraffic[] | null>(path + "/stats", { params }))
      .data ?? []
  );
}
export async function getBandwidth(params: DateRangeQuery): Promise<Bandwidth> {
  return isTestMode
    ? mockCustomers.bandwidth(params)
    : (await httpClient.get<Bandwidth>(path + "/bandwidth", { params })).data;
}
export async function getCiscoSetup(): Promise<CiscoSetup> {
  return isTestMode
    ? mockCustomers.ciscoSetup()
    : (await httpClient.get<CiscoSetup>(path + "/setup/cisco")).data;
}
async function download(url: string): Promise<Blob> {
  return (await httpClient.get<Blob>(url, { responseType: "blob" })).data;
}
export function downloadCertificate(): Promise<Blob> {
  return isTestMode
    ? mockCustomers.certificate()
    : download(path + "/certificate");
}
export function downloadCiscoCertificate(): Promise<Blob> {
  return isTestMode
    ? mockCustomers.certificate()
    : download(path + "/setup/cisco/certificate");
}

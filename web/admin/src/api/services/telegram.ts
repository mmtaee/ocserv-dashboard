import { isTestMode } from "@/api/environment";
import { httpClient } from "@/api/http";
import * as mock from "@/mocks/telegram";

export const telegramLanguages = [
  "en",
  "fa",
  "ar",
  "ru",
  "zh-cn",
  "zh-tw",
  "it",
] as const;
export type TelegramLanguage = (typeof telegramLanguages)[number];
export const telegramTrafficTypes = [
  "Free",
  "MonthlyTransmit",
  "MonthlyReceive",
  "MonthlyRxTx",
  "TotallyTransmit",
  "TotallyReceive",
  "TotallyRxTx",
] as const;
export type TelegramTrafficType = (typeof telegramTrafficTypes)[number];
export const telegramRequestStatuses = [
  "pending",
  "awaiting_payment",
  "payment_uploaded",
  "approved",
  "rejected",
  "delivered",
] as const;
export type TelegramRequestStatus = (typeof telegramRequestStatuses)[number];
export const telegramRequestTypes = ["new", "renew"] as const;
export type TelegramRequestType = (typeof telegramRequestTypes)[number];

export interface TelegramSettings {
  enabled?: boolean;
  bot_token?: string;
  bot_username?: string;
  admin_chat_id?: number;
  low_quota_threshold_mb?: number;
  default_language?: TelegramLanguage;
  ocserv_host?: string;
  card_number?: string;
  card_holder?: string;
  support_username?: string;
}
export type TelegramSettingsUpdate = Omit<TelegramSettings, "bot_username">;
export interface TelegramPackage {
  id?: number;
  title?: string;
  days?: number;
  traffic_size_gb?: number;
  traffic_type: TelegramTrafficType;
  price_text?: string;
  is_active?: boolean;
  created_at?: string;
  updated_at?: string;
}
export interface TelegramPackageCreate {
  title: string;
  days: number;
  traffic_size_gb?: number;
  traffic_type: TelegramTrafficType;
  price_text?: string;
  is_active?: boolean;
}
export type TelegramPackageUpdate = Partial<TelegramPackageCreate>;
export interface TelegramRequest {
  id?: number;
  status: TelegramRequestStatus;
  type: TelegramRequestType;
  admin_note?: string;
  awaiting_payment_message_id?: number;
  chat_id?: number;
  created_at?: string;
  delivered_at?: string;
  desired_username?: string;
  package_id?: number;
  receipt_file_path?: string;
  target_ocserv_user_id?: number;
  telegram_username?: string;
  updated_at?: string;
  user_message?: string;
}
export interface TelegramAccount {
  id?: number;
  chat_id?: number;
  created_at?: string;
  language: TelegramLanguage;
  last_low_quota_notified_at?: string;
  ocserv_user_id?: number;
  telegram_username?: string;
}
export interface TelegramRequestsQuery {
  page: number;
  size: number;
  order: string;
  sort: "ASC" | "DESC";
  status?: TelegramRequestStatus;
  type?: TelegramRequestType;
}
export interface TelegramRequestsResponse {
  meta?: { page: number; size: number; total_records: number };
  result?: TelegramRequest[];
}
export interface TelegramApproveData {
  admin_note?: string;
  card_holder?: string;
  card_number?: string;
  reply_to_user?: string;
}
export interface TelegramRejectData {
  admin_note?: string;
}
export interface TelegramConfirmPaymentData {
  admin_note?: string;
  group?: string;
  override_password?: string;
  override_username?: string;
}

export async function getTelegramSettings(): Promise<TelegramSettings> {
  if (isTestMode) return mock.getMockTelegramSettings();
  return (await httpClient.get<TelegramSettings>("/telegram/settings")).data;
}
export async function updateTelegramSettings(
  request: TelegramSettingsUpdate,
): Promise<TelegramSettings> {
  if (isTestMode) return mock.saveMockTelegramSettings(request);
  return (
    await httpClient.patch<TelegramSettings>("/telegram/settings", request)
  ).data;
}
export async function testTelegram(
  message?: string,
): Promise<Record<string, string>> {
  if (isTestMode) return mock.testMockTelegram(message);
  return (
    await httpClient.post<Record<string, string>>("/telegram/test", {
      message: message || "",
    })
  ).data;
}
export async function getTelegramPackages(
  includeInactive = true,
): Promise<TelegramPackage[]> {
  if (isTestMode) return mock.getMockTelegramPackages(includeInactive);
  return (
    await httpClient.get<TelegramPackage[]>("/telegram/packages", {
      params: { include_inactive: includeInactive },
    })
  ).data;
}
export async function createTelegramPackage(
  request: TelegramPackageCreate,
): Promise<TelegramPackage> {
  if (isTestMode) return mock.saveMockTelegramPackage(request);
  return (await httpClient.post<TelegramPackage>("/telegram/packages", request))
    .data;
}
export async function updateTelegramPackage(
  id: number,
  request: TelegramPackageUpdate,
): Promise<TelegramPackage> {
  if (isTestMode) return mock.saveMockTelegramPackage(request, id);
  return (
    await httpClient.patch<TelegramPackage>(`/telegram/packages/${id}`, request)
  ).data;
}
export async function deleteTelegramPackage(id: number): Promise<void> {
  if (isTestMode) return mock.deleteMockTelegramPackage(id);
  await httpClient.delete(`/telegram/packages/${id}`);
}
export async function getTelegramRequests(
  query: TelegramRequestsQuery,
): Promise<TelegramRequestsResponse> {
  if (isTestMode) return mock.getMockTelegramRequests(query);
  return (
    await httpClient.get<TelegramRequestsResponse>("/telegram/requests", {
      params: query,
    })
  ).data;
}
export async function getTelegramRequest(id: number): Promise<TelegramRequest> {
  if (isTestMode) return mock.getMockTelegramRequest(id);
  return (await httpClient.get<TelegramRequest>(`/telegram/requests/${id}`))
    .data;
}
export async function approveTelegramRequest(
  id: number,
  request: TelegramApproveData,
): Promise<TelegramRequest> {
  if (isTestMode) return mock.mutateMockTelegramRequest(id, "approve", request);
  return (
    await httpClient.post<TelegramRequest>(
      `/telegram/requests/${id}/approve`,
      request,
    )
  ).data;
}
export async function rejectTelegramRequest(
  id: number,
  request: TelegramRejectData,
): Promise<TelegramRequest> {
  if (isTestMode) return mock.mutateMockTelegramRequest(id, "reject", request);
  return (
    await httpClient.post<TelegramRequest>(
      `/telegram/requests/${id}/reject`,
      request,
    )
  ).data;
}
export async function confirmTelegramPayment(
  id: number,
  request: TelegramConfirmPaymentData,
): Promise<Record<string, unknown>> {
  if (isTestMode)
    return mock
      .mutateMockTelegramRequest(id, "confirm", request)
      .then((request) => ({ request }));
  return (
    await httpClient.post<Record<string, unknown>>(
      `/telegram/requests/${id}/confirm-payment`,
      request,
    )
  ).data;
}
export async function deleteTelegramRequest(id: number): Promise<void> {
  if (isTestMode) return mock.deleteMockTelegramRequest(id);
  await httpClient.delete(`/telegram/requests/${id}`);
}
export async function downloadTelegramReceipt(id: number): Promise<Blob> {
  if (isTestMode) return new Blob(["mock receipt"], { type: "text/plain" });
  return (
    await httpClient.get(`/telegram/requests/${id}/receipt`, {
      responseType: "blob",
    })
  ).data;
}
export async function getTelegramAccounts(
  ocservUserId: number,
): Promise<TelegramAccount[]> {
  if (isTestMode) return mock.getMockTelegramAccounts(ocservUserId);
  return (
    await httpClient.get<TelegramAccount[]>("/telegram/accounts", {
      params: { ocserv_user_id: ocservUserId },
    })
  ).data;
}
export async function deleteTelegramAccount(id: number): Promise<void> {
  if (isTestMode) return mock.deleteMockTelegramAccount(id);
  await httpClient.delete(`/telegram/accounts/${id}`);
}

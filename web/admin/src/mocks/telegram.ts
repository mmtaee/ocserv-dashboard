import { ApiError } from "@/api/http";
import type {
  TelegramAccount,
  TelegramApproveData,
  TelegramConfirmPaymentData,
  TelegramPackage,
  TelegramPackageCreate,
  TelegramPackageUpdate,
  TelegramRejectData,
  TelegramRequest,
  TelegramRequestsQuery,
  TelegramRequestsResponse,
  TelegramSettings,
  TelegramSettingsUpdate,
} from "@/api/services/telegram";
import { cloneMock } from "@/mocks/utils";

let settings: TelegramSettings = {
  enabled: true,
  bot_token: "123456:mock-token",
  bot_username: "ocserv_demo_bot",
  admin_chat_id: 778899,
  low_quota_threshold_mb: 512,
  default_language: "en",
  ocserv_host: "vpn.example.com",
  card_number: "6037-0000-0000-0000",
  card_holder: "VPN Services",
  support_username: "vpn_support",
};
let packages: TelegramPackage[] = [
  {
    id: 1,
    title: "30 days · 50 GB",
    days: 30,
    traffic_size_gb: 50,
    traffic_type: "TotallyRxTx",
    price_text: "$12",
    is_active: true,
    created_at: "2026-09-01T08:00:00Z",
  },
  {
    id: 2,
    title: "Trial",
    days: 3,
    traffic_size_gb: 0,
    traffic_type: "Free",
    price_text: "Free",
    is_active: false,
    created_at: "2026-08-20T08:00:00Z",
  },
];
let requests: TelegramRequest[] = [
  {
    id: 10,
    status: "pending",
    type: "new",
    chat_id: 10001,
    telegram_username: "alice",
    desired_username: "alice_vpn",
    package_id: 1,
    user_message: "Please activate my account",
    created_at: "2026-09-10T08:30:00Z",
  },
  {
    id: 11,
    status: "payment_uploaded",
    type: "renew",
    chat_id: 10002,
    telegram_username: "bob",
    target_ocserv_user_id: 42,
    package_id: 1,
    receipt_file_path: "/receipts/11.jpg",
    created_at: "2026-09-09T10:00:00Z",
  },
  {
    id: 12,
    status: "delivered",
    type: "new",
    telegram_username: "carol",
    delivered_at: "2026-09-08T12:00:00Z",
    created_at: "2026-09-08T10:00:00Z",
  },
  {
    id: 13,
    status: "rejected",
    type: "new",
    telegram_username: "david",
    admin_note: "Invalid receipt",
    created_at: "2026-09-07T10:00:00Z",
  },
];
let accounts: TelegramAccount[] = [
  {
    id: 7,
    chat_id: 10002,
    ocserv_user_id: 42,
    telegram_username: "bob",
    language: "en",
    created_at: "2026-08-01T10:00:00Z",
  },
];
const scenario = () => import.meta.env.VITE_TELEGRAM_MOCK_SCENARIO || "normal";
async function wait(): Promise<void> {
  await new Promise((resolve) => setTimeout(resolve, 200));
  if (scenario() === "error")
    throw new ApiError("Telegram backend is unavailable.", { status: 500 });
}
function required(value: unknown, message: string): void {
  if (value == null || value === "")
    throw new ApiError(message, { status: 400 });
}

export async function getMockTelegramSettings(): Promise<TelegramSettings> {
  await wait();
  return cloneMock(scenario() === "empty" ? { enabled: false } : settings);
}
export async function saveMockTelegramSettings(
  update: TelegramSettingsUpdate,
): Promise<TelegramSettings> {
  await wait();
  settings = { ...settings, ...update };
  return cloneMock(settings);
}
export async function testMockTelegram(
  message?: string,
): Promise<Record<string, string>> {
  await wait();
  if (!settings.bot_token)
    throw new ApiError("Bot token is required.", { status: 400 });
  return { message: message || "Test message delivered", status: "ok" };
}
export async function getMockTelegramPackages(
  includeInactive: boolean,
): Promise<TelegramPackage[]> {
  await wait();
  return cloneMock(
    scenario() === "empty"
      ? []
      : packages.filter((item) => includeInactive || item.is_active),
  );
}
export async function saveMockTelegramPackage(
  update: TelegramPackageCreate | TelegramPackageUpdate,
  id?: number,
): Promise<TelegramPackage> {
  await wait();
  if (id == null) {
    required(update.title, "Title is required.");
    required(update.days, "Days is required.");
    required(update.traffic_type, "Traffic type is required.");
    const item = {
      ...update,
      id: Math.max(0, ...packages.map((entry) => entry.id ?? 0)) + 1,
      created_at: new Date().toISOString(),
    } as TelegramPackage;
    packages.push(item);
    return cloneMock(item);
  }
  const index = packages.findIndex((item) => item.id === id);
  if (index < 0) throw new ApiError("Package not found.", { status: 400 });
  packages[index] = {
    ...packages[index],
    ...update,
    updated_at: new Date().toISOString(),
  };
  return cloneMock(packages[index]);
}
export async function deleteMockTelegramPackage(id: number): Promise<void> {
  await wait();
  packages = packages.filter((item) => item.id !== id);
}
export async function getMockTelegramRequests(
  query: TelegramRequestsQuery,
): Promise<TelegramRequestsResponse> {
  await wait();
  let result =
    scenario() === "empty"
      ? []
      : requests.filter(
          (item) =>
            (!query.status || item.status === query.status) &&
            (!query.type || item.type === query.type),
        );
  result = [...result].sort(
    (a, b) =>
      String(
        (a as unknown as Record<string, unknown>)[query.order] ?? "",
      ).localeCompare(
        String((b as unknown as Record<string, unknown>)[query.order] ?? ""),
      ) * (query.sort === "ASC" ? 1 : -1),
  );
  const start = (query.page - 1) * query.size;
  return cloneMock({
    meta: { page: query.page, size: query.size, total_records: result.length },
    result: result.slice(start, start + query.size),
  });
}
export async function getMockTelegramRequest(
  id: number,
): Promise<TelegramRequest> {
  await wait();
  const item = requests.find((entry) => entry.id === id);
  if (!item) throw new ApiError("Request not found.", { status: 400 });
  return cloneMock(item);
}
export async function mutateMockTelegramRequest(
  id: number,
  action: "approve" | "reject" | "confirm",
  data: TelegramApproveData | TelegramRejectData | TelegramConfirmPaymentData,
): Promise<TelegramRequest> {
  await wait();
  const item = requests.find((entry) => entry.id === id);
  if (!item) throw new ApiError("Request not found.", { status: 400 });
  if (action === "approve" && item.status !== "pending")
    throw new ApiError("Only pending requests can be approved.", {
      status: 400,
    });
  if (action === "confirm" && item.status !== "payment_uploaded")
    throw new ApiError("A receipt must be uploaded first.", { status: 400 });
  if (action === "reject" && item.status === "delivered")
    throw new ApiError("Delivered requests cannot be rejected.", {
      status: 400,
    });
  item.status =
    action === "approve"
      ? "awaiting_payment"
      : action === "reject"
        ? "rejected"
        : "delivered";
  item.admin_note = data.admin_note;
  item.updated_at = new Date().toISOString();
  if (action === "confirm") item.delivered_at = item.updated_at;
  return cloneMock(item);
}
export async function deleteMockTelegramRequest(id: number): Promise<void> {
  await wait();
  const item = requests.find((entry) => entry.id === id);
  if (
    item &&
    ["pending", "awaiting_payment", "payment_uploaded"].includes(item.status)
  )
    throw new ApiError("Active requests cannot be deleted.", { status: 400 });
  requests = requests.filter((entry) => entry.id !== id);
}
export async function getMockTelegramAccounts(
  ocservUserId: number,
): Promise<TelegramAccount[]> {
  await wait();
  return cloneMock(
    scenario() === "empty"
      ? []
      : accounts.filter((item) => item.ocserv_user_id === ocservUserId),
  );
}
export async function deleteMockTelegramAccount(id: number): Promise<void> {
  await wait();
  accounts = accounts.filter((item) => item.id !== id);
}

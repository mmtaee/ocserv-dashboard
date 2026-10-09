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
  SessionLog,
  SummaryResponse,
} from "@/api/generated";
import { ApiError } from "@/api/http";

const scenario = () => import.meta.env.VITE_CUSTOMER_MOCK_SCENARIO || "success";
async function ready(): Promise<void> {
  await new Promise((resolve) => setTimeout(resolve, 150));
  if (scenario() === "error") throw new ApiError("Service unavailable", 503);
}
const summary: SummaryResponse = {
  ocserv_user: {
    username: "customer",
    owner: "Customer",
    expiry_mode: "fixed",
    expire_at: "2027-01-01",
    traffic_type: "MonthlyRxTx",
    traffic_size: 100,
    running_rx: 12,
    running_tx: 4,
    certificate_available: true,
    certificate_enabled: true,
    is_locked: false,
  },
  usage: {
    date_start: "2026-09-01",
    date_end: "2026-09-30",
    bandwidths: { rx: 12, tx: 4 },
  },
};
const sessions: OnlineUserSession[] = [
  {
    ID: 1,
    Device: "iPhone 15 Pro",
    IPv4: "10.10.0.2",
    "Session started at": "2026-09-28T10:00:00Z",
    vhost: "default",
    "Average RX": "1.2 MiB/s",
    "Average TX": "420 KiB/s",
  },
  {
    ID: 2,
    Device: "MacBook Pro",
    IPv4: "10.10.0.3",
    "Session started at": "2026-09-28T08:42:16Z",
    vhost: "default",
    "Average RX": "3.8 MiB/s",
    "Average TX": "860 KiB/s",
  },
  {
    ID: 3,
    Device: "Windows Desktop",
    IPv4: "10.10.0.4",
    "Session started at": "2026-09-27T20:15:04Z",
    vhost: "default",
    "Average RX": "980 KiB/s",
    "Average TX": "220 KiB/s",
  },
  {
    ID: 4,
    Device: "Android Tablet",
    IPv4: "10.10.0.5",
    "Session started at": "2026-09-27T17:33:48Z",
    vhost: "default",
    "Average RX": "2.1 MiB/s",
    "Average TX": "510 KiB/s",
  },
];

const activities: SessionLog[] = [
  {
    username: "customer",
    event: "periodic-stats",
    message: "Traffic usage updated: 16.0 GiB this billing period",
    ip: "192.0.2.10",
    created_at: "2026-09-28T10:30:00Z",
  },
  {
    username: "customer",
    event: "handshake",
    message: "Session established from iPhone 15 Pro",
    ip: "192.0.2.10",
    created_at: "2026-09-28T10:00:00Z",
  },
  {
    username: "customer",
    event: "handshake",
    message: "Session established from MacBook Pro",
    ip: "198.51.100.24",
    created_at: "2026-09-28T08:42:16Z",
  },
  {
    username: "customer",
    event: "disconnect",
    message: "Session closed normally after 2h 18m",
    ip: "198.51.100.24",
    created_at: "2026-09-27T22:33:11Z",
  },
  {
    username: "customer",
    event: "handshake",
    message: "Session established from Windows Desktop",
    ip: "203.0.113.73",
    created_at: "2026-09-27T20:15:04Z",
  },
  {
    username: "customer",
    event: "periodic-stats",
    message: "Traffic usage updated: 13.8 GiB this billing period",
    ip: "203.0.113.73",
    created_at: "2026-09-26T18:00:00Z",
  },
  {
    username: "customer",
    event: "disconnect",
    message: "Session closed normally after 46m",
    ip: "192.0.2.10",
    created_at: "2026-09-25T09:18:37Z",
  },
  {
    username: "customer",
    event: "handshake",
    message: "Session established from Android Tablet",
    ip: "198.51.100.87",
    created_at: "2026-09-25T08:32:54Z",
  },
  {
    username: "customer",
    event: "handshake",
    message: "Session established from iPhone 15 Pro",
    ip: "192.0.2.10",
    created_at: "2026-09-23T07:21:09Z",
  },
  {
    username: "customer",
    event: "disconnect",
    message: "Session terminated by customer",
    ip: "198.51.100.24",
    created_at: "2026-09-21T19:45:23Z",
  },
  {
    username: "customer",
    event: "handshake",
    message: "Session established from MacBook Pro",
    ip: "198.51.100.24",
    created_at: "2026-09-21T17:05:41Z",
  },
  {
    username: "customer",
    event: "periodic-stats",
    message: "Traffic usage updated: 7.2 GiB this billing period",
    ip: "203.0.113.73",
    created_at: "2026-09-18T12:00:00Z",
  },
];

const stats: DailyTraffic[] = [
  { date: "2026-09-15", rx: 0.42, tx: 0.18 },
  { date: "2026-09-16", rx: 0.86, tx: 0.31 },
  { date: "2026-09-17", rx: 1.2, tx: 0.44 },
  { date: "2026-09-18", rx: 0.65, tx: 0.22 },
  { date: "2026-09-19", rx: 1.48, tx: 0.61 },
  { date: "2026-09-20", rx: 0.93, tx: 0.35 },
  { date: "2026-09-21", rx: 1.61, tx: 0.72 },
  { date: "2026-09-22", rx: 0.77, tx: 0.26 },
  { date: "2026-09-23", rx: 1.32, tx: 0.55 },
  { date: "2026-09-24", rx: 0.58, tx: 0.19 },
  { date: "2026-09-25", rx: 1.71, tx: 0.68 },
  { date: "2026-09-26", rx: 0.96, tx: 0.39 },
  { date: "2026-09-27", rx: 1.46, tx: 0.63 },
  { date: "2026-09-28", rx: 1.05, tx: 0.47 },
];

function inDateRange(
  value: string,
  dateStart?: string,
  dateEnd?: string,
): boolean {
  const date = value.slice(0, 10);
  return (!dateStart || date >= dateStart) && (!dateEnd || date <= dateEnd);
}

function filteredStats(dateStart?: string, dateEnd?: string): DailyTraffic[] {
  return stats.filter((item) =>
    item.date ? inDateRange(item.date, dateStart, dateEnd) : false,
  );
}
export const mockCustomers = {
  async login(request: LoginData): Promise<LoginResponse> {
    await ready();
    if (request.username === "invalid")
      throw new ApiError("Invalid credentials", 400);
    return {
      token: "mock-customer-token",
      expires_at: new Date(Date.now() + 7 * 60 * 60 * 1000).toISOString(),
      user: { ...summary.ocserv_user, username: request.username },
    };
  },
  async summary(): Promise<SummaryResponse> {
    await ready();
    return structuredClone(summary);
  },
  async sessions(): Promise<OnlineUserSession[]> {
    await ready();
    return scenario() === "empty" ? [] : structuredClone(sessions);
  },
  async activities(params: ActivitiesQuery): Promise<ActivitiesResponse> {
    await ready();
    const result =
      scenario() === "empty"
        ? []
        : activities.filter((item) =>
            inDateRange(item.created_at, params.date_start, params.date_end),
          );
    const page = Math.max(1, params.page ?? 1);
    const size = Math.max(1, params.size ?? 20);
    return {
      meta: {
        page,
        size,
        total_records: result.length,
      },
      result: result.slice((page - 1) * size, page * size),
    };
  },
  async stats(params?: DateRangeQuery): Promise<DailyTraffic[]> {
    await ready();
    return scenario() === "empty"
      ? []
      : structuredClone(filteredStats(params?.date_start, params?.date_end));
  },
  async bandwidth(params?: DateRangeQuery): Promise<Bandwidth> {
    await ready();
    return filteredStats(
      params?.date_start,
      params?.date_end,
    ).reduce<Bandwidth>(
      (total, item) => ({
        rx: total.rx + (item.rx ?? 0),
        tx: total.tx + (item.tx ?? 0),
      }),
      { rx: 0, tx: 0 },
    );
  },
  async ciscoSetup(): Promise<CiscoSetup> {
    await ready();
    return {
      connection_name: "Ocserv VPN",
      server_address: "vpn.example.test",
      server_port: 443,
      certificate_password: "customer",
      certificate_import_uri: "cisco-secure-client://import",
      connection_create_uri: "cisco-secure-client://connect",
    };
  },
  async certificate(): Promise<Blob> {
    await ready();
    return new Blob(["mock certificate"], { type: "application/x-pkcs12" });
  },
  async accept(): Promise<void> {
    await ready();
  },
  async changePassword(request: ChangePasswordInput): Promise<void> {
    await ready();
    if (request.password.length < 2)
      throw new ApiError("Password must contain at least 2 characters", 400);
  },
};

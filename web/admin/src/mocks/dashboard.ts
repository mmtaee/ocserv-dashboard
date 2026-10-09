import type {
  GithubComMmtaeeOcservDashboardBackendInternalServicesAdminApiDashboardGetHomeResponse,
  GithubComMmtaeeOcservDashboardBackendInternalServicesAdminApiDashboardOcservStatusResponse,
  GithubComMmtaeeOcservDashboardBackendInternalServicesAdminApiDashboardServerStatusResponse,
  GithubComMmtaeeOcservDashboardBackendInternalServicesAdminApiDashboardDockerStats,
  ModelsOcservUser,
  ModelsOnlineUserSession,
} from "@/api/generated";

const aliceSession = {
  "Average RX": "2.4 MiB/s",
  "Average TX": "640 KiB/s",
  Device: "vpns0",
  Groupname: "staff",
  ID: 101,
  IPv4: "10.10.0.12",
  "Session started at": "2026-08-30 09:20:00",
  Username: "alice",
  "_Last connected at": "2026-08-30T09:20:00Z",
  vhost: "default",
} satisfies ModelsOnlineUserSession;

const bobSession = {
  "Average RX": "1.1 MiB/s",
  "Average TX": "420 KiB/s",
  Device: "vpns1",
  Groupname: "engineering",
  ID: 102,
  IPv4: "10.10.0.18",
  "Session started at": "2026-08-30 10:05:00",
  Username: "bob",
  "_Last connected at": "2026-08-30T10:05:00Z",
  vhost: "default",
} satisfies ModelsOnlineUserSession;

function mockBandwidthUser(
  id: number,
  username: string,
  runningRx: number,
  runningTx: number,
  session: ModelsOnlineUserSession,
): ModelsOcservUser {
  return {
    created_at: "2026-01-10T10:00:00Z",
    expiry_mode: "unlimited",
    group: session.Groupname ?? "default",
    id,
    is_locked: false,
    is_online: true,
    online_sessions: [session],
    owner_id: 1,
    password: "mock-password",
    running_rx: runningRx,
    running_tx: runningTx,
    traffic_size: 0,
    traffic_type: "TotallyRxTx",
    username,
  };
}

const alice = mockBandwidthUser(
  11,
  "alice",
  68_719_476_736,
  21_474_836_480,
  aliceSession,
);
const bob = mockBandwidthUser(
  12,
  "bob",
  44_023_382_016,
  32_212_254_720,
  bobSession,
);

export const mockDashboardOverview = {
  telegram_service: {
    bot_username: "ocserv_test_bot",
    enabled: true,
    has_bot_token: true,
  },
  users: {
    total: 48,
    online_users_session: [aliceSession, bobSession],
  },
  statistics: [
    { date: "2026-08-21", rx: 18.2, tx: 7.4 },
    { date: "2026-08-22", rx: 21.6, tx: 8.1 },
    { date: "2026-08-23", rx: 17.9, tx: 6.8 },
    { date: "2026-08-24", rx: 25.4, tx: 9.7 },
    { date: "2026-08-25", rx: 29.1, tx: 11.2 },
    { date: "2026-08-26", rx: 23.8, tx: 10.5 },
    { date: "2026-08-27", rx: 31.7, tx: 12.9 },
    { date: "2026-08-28", rx: 35.3, tx: 14.1 },
    { date: "2026-08-29", rx: 32.6, tx: 13.4 },
    { date: "2026-08-30", rx: 38.9, tx: 15.8 },
  ],
  ip_bans: [
    {
      IP: "203.0.113.42",
      Score: 18,
      Since: "2 hours ago",
      _Since: "2026-08-30T08:00:00Z",
    },
    {
      IP: "198.51.100.17",
      Score: 12,
      Since: "1 day ago",
      _Since: "2026-08-29T10:00:00Z",
    },
  ],
  top_bandwidth_user: {
    top_rx: [alice, bob],
    top_tx: [bob, alice],
  },
  total_bandwidth: {
    rx: 927_712_935_936,
    tx: 356_482_285_568,
  },
} satisfies GithubComMmtaeeOcservDashboardBackendInternalServicesAdminApiDashboardGetHomeResponse;

export const mockSystemStats = {
  cpu: { avg_percent: 36.4, total: 8, used_units: 2.9 },
  ram: { total: 16, used: 8.64, used_percent: 54 },
  disk: { total: 256, used: 148.48, used_percent: 58 },
  swap: { total: 4, used: 0.64, used_percent: 16 },
} satisfies GithubComMmtaeeOcservDashboardBackendInternalServicesAdminApiDashboardServerStatusResponse;

function dockerStats(name: string, cpuPercent: number, ramPercent: number) {
  return {
    name,
    cpu: {
      avg_percent: cpuPercent,
      total: 8,
      used_units: (cpuPercent / 100) * 8,
    },
    ram: {
      total: 2,
      used: (2 * ramPercent) / 100,
      used_percent: ramPercent,
    },
  };
}

export const mockContainerStats = dockerStats(
  "ocserv",
  12.7,
  44,
) satisfies GithubComMmtaeeOcservDashboardBackendInternalServicesAdminApiDashboardDockerStats;

export const mockOcservStats = {
  general_info: {
    Status: "online",
    "Active sessions": 2,
    "IPs in ban list": 2,
    "Median latency": "24 ms",
    "STDEV latency": "8 ms",
    "Sec-mod PID": 2314,
    "Sec-mod instance count": 2,
    "Server PID": 2298,
    "Total authentication failures": 13,
    "Total sessions": 1842,
    "Up since": "12 days, 4 hours",
    "_Up since": "2026-08-18T05:30:00Z",
    raw_median_latency: 24,
    raw_stdev_latency: 8,
    raw_up_since: 1_052_400,
    uptime: 1_052_400,
  },
  current_stats: {
    "Authentication failures": 1,
    "Average auth time": "180 ms",
    "Average session time": "2 hours, 18 minutes",
    "Closed due to error sessions": 3,
    "Last stats reset": "6 hours ago",
    "Max auth time": "640 ms",
    "Max session time": "9 hours, 42 minutes",
    RX: "412.8 GiB",
    "Sessions handled": 1842,
    TX: "166.4 GiB",
    "Timed out (idle) sessions": 8,
    "Timed out sessions": 4,
    "_Last stats reset": "2026-08-30T04:00:00Z",
    raw_avg_auth_time: 0.18,
    raw_avg_session_time: 8280,
    raw_last_stats_reset: 21_600,
    raw_max_auth_time: 0.64,
    raw_max_session_time: 34_920,
    raw_rx: 443_240_624_947,
    raw_tx: 178_670_639_514,
  },
} satisfies GithubComMmtaeeOcservDashboardBackendInternalServicesAdminApiDashboardOcservStatusResponse;

import { ApiError } from "@/api/http";
import type {
  DailyTraffic,
  DateRange,
  SessionLog,
  SessionLogQuery,
  SessionLogsResponse,
  TotalBandwidth,
  UserReportSummary,
} from "@/api/services/reports";
import { cloneMock } from "@/mocks/utils";

const traffic = [
  { date: "2026-09-05", rx: 12.4, tx: 4.8 },
  { date: "2026-09-06", rx: 18.1, tx: 7.2 },
  { date: "2026-09-07", rx: 15.7, tx: 6.3 },
  { date: "2026-09-08", rx: 23.9, tx: 9.4 },
  { date: "2026-09-09", rx: 20.6, tx: 8.1 },
  { date: "2026-09-10", rx: 27.3, tx: 10.7 },
] satisfies DailyTraffic[];

const logs = [
  {
    created_at: "2026-09-10T14:25:00Z",
    event: "disconnect",
    ip: "203.0.113.12",
    message: "Client disconnected",
    username: "alice",
  },
  {
    created_at: "2026-09-10T12:02:00Z",
    event: "periodic-stats",
    ip: "198.51.100.9",
    message: "Periodic statistics recorded",
    username: "bob",
  },
  {
    created_at: "2026-09-09T09:11:00Z",
    event: "handshake",
    ip: "203.0.113.44",
    message: "TLS handshake completed",
    username: "carol",
  },
  {
    created_at: "2026-09-08T18:42:00Z",
    event: "user-agent",
    message: "Client agent detected",
    username: "david",
  },
] satisfies SessionLog[];

function scenario(): string {
  return import.meta.env.VITE_REPORTS_MOCK_SCENARIO || "normal";
}

async function wait(): Promise<void> {
  await new Promise((resolve) => setTimeout(resolve, 200));
  if (scenario() === "error")
    throw new ApiError("The report request failed.", { status: 500 });
}

function inRange(date: string | undefined, range: Partial<DateRange>): boolean {
  if (!date) return false;
  return (
    (!range.dateStart || date >= range.dateStart) &&
    (!range.dateEnd || date.slice(0, 10) <= range.dateEnd)
  );
}

export async function getMockDailyTraffic(
  range: DateRange,
): Promise<DailyTraffic[]> {
  await wait();
  return scenario() === "empty"
    ? []
    : cloneMock(traffic.filter((item) => inRange(item.date, range)));
}

export async function getMockUserSummary(): Promise<UserReportSummary> {
  await wait();
  return cloneMock(
    scenario() === "empty"
      ? {}
      : { active: 42, deactivated: 7, locked: 3, online: 11 },
  );
}

export async function getMockTotalBandwidth(
  range: DateRange,
): Promise<TotalBandwidth> {
  await wait();
  const selected =
    scenario() === "empty"
      ? []
      : traffic.filter((item) => inRange(item.date, range));
  return selected.reduce(
    (total, item) => ({
      rx: total.rx + (item.rx ?? 0),
      tx: total.tx + (item.tx ?? 0),
    }),
    { rx: 0, tx: 0 },
  );
}

export async function getMockSessionLogs(
  query: SessionLogQuery,
): Promise<SessionLogsResponse> {
  await wait();
  const selected =
    scenario() === "empty"
      ? []
      : logs.filter((item) => inRange(item.created_at, query));
  const sorted = [...selected].sort((left, right) => {
    const result = String(left[query.order] ?? "").localeCompare(
      String(right[query.order] ?? ""),
    );
    return query.sort === "ASC" ? result : -result;
  });
  const start = (query.page - 1) * query.size;
  return {
    meta: { page: query.page, size: query.size, total_records: sorted.length },
    result: cloneMock(sorted.slice(start, start + query.size)),
  };
}

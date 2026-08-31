import type { ModelsOcservInfo } from "@/api/generated";
import { ApiError } from "@/api/http";

export const mockOcctlServerInfo = {
  status: "online",
  version: {
    occtl_version: "1.3.0",
    ocserv_version: "1.3.0",
  },
} satisfies ModelsOcservInfo;

const commandsRequiringValue = new Set([2, 3, 4, 7, 9, 14, 15, 16]);

export function mockOcctlCommand(action: number, value?: string): string {
  const parameter = value?.trim();
  if (commandsRequiringValue.has(action) && !parameter)
    throw new ApiError("This command requires a value.", { status: 400 });

  const results: Record<number, unknown> = {
    1: [
      {
        Device: "vpns0",
        Groupname: "engineering",
        ID: 1042,
        IPv4: "10.20.0.14",
        Username: "ava",
      },
    ],
    2: { Groupname: "engineering", Username: parameter, status: "connected" },
    3: { ID: Number(parameter), Username: "ava", status: "connected" },
    4: { disconnected: parameter },
    5: [{ ID: 1042, Username: "ava", status: "connected" }],
    6: [{ ID: 1042, Username: "ava", status: "valid" }],
    7: { ID: Number(parameter), Username: "ava", status: "connected" },
    8: [{ IP: "203.0.113.42", Score: 18 }],
    9: { unbanned: parameter },
    10: { "Active sessions": 1, Status: "online", "Total sessions": 1842 },
    11: { event: "periodic-stats", message: "Statistics event received." },
    12: [{ Device: "vpns0", IP: "10.20.0.14", route: "10.0.0.0/8" }],
    13: { reloaded: true },
    14: { disconnected_session: parameter },
    15: { terminated: parameter },
    16: { terminated_session: parameter },
  };

  if (!Object.hasOwn(results, action))
    throw new ApiError(`Unknown Occtl action ${action}.`, { status: 400 });
  return JSON.stringify(results[action]);
}

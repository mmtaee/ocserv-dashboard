import type { OcctlAction } from "@/api/services/occtl";

export type OcctlCommandValueKind = "ip" | "sessionId" | "userId" | "username";

export interface OcctlCommandDefinition {
  action: OcctlAction;
  destructive?: boolean;
  labelKey: string;
  valueKind?: OcctlCommandValueKind;
}

export const occtlCommands = [
  { action: 1, labelKey: "onlineSessions" },
  { action: 2, labelKey: "showUserByUsername", valueKind: "username" },
  { action: 3, labelKey: "showUserById", valueKind: "userId" },
  {
    action: 4,
    destructive: true,
    labelKey: "disconnectUser",
    valueKind: "username",
  },
  { action: 5, labelKey: "allSessions" },
  { action: 6, labelKey: "validSessions" },
  { action: 7, labelKey: "showSession", valueKind: "sessionId" },
  { action: 8, labelKey: "ipBans" },
  {
    action: 9,
    destructive: true,
    labelKey: "unbanIp",
    valueKind: "ip",
  },
  { action: 10, labelKey: "serverStatus" },
  { action: 11, labelKey: "showEvent" },
  { action: 12, labelKey: "iroutes" },
  { action: 13, destructive: true, labelKey: "reloadServer" },
  {
    action: 14,
    destructive: true,
    labelKey: "disconnectSession",
    valueKind: "sessionId",
  },
  {
    action: 15,
    destructive: true,
    labelKey: "terminateUser",
    valueKind: "username",
  },
  {
    action: 16,
    destructive: true,
    labelKey: "terminateSession",
    valueKind: "sessionId",
  },
] as const satisfies readonly OcctlCommandDefinition[];

export function getOcctlCommand(action: OcctlAction): OcctlCommandDefinition {
  return (
    occtlCommands.find((command) => command.action === action) ??
    occtlCommands[0]
  );
}

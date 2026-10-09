import type { ModelsOcservGroupConfig } from "@/api/generated";

export type DefaultGroupFieldKind = "boolean" | "list" | "number" | "text";
export type DefaultGroupFieldKey = keyof ModelsOcservGroupConfig;

export interface DefaultGroupField {
  key: DefaultGroupFieldKey;
  kind: DefaultGroupFieldKind;
  placeholder?: string;
}

export const networkFields = [
  { key: "ipv4-network", kind: "text", placeholder: "192.168.1.0/24" },
  { key: "explicit-ipv4", kind: "text", placeholder: "192.168.100.10" },
  { key: "dns", kind: "list", placeholder: "8.8.8.8, 1.1.1.1" },
  { key: "nbns", kind: "text", placeholder: "192.168.1.1" },
  {
    key: "split-dns",
    kind: "list",
    placeholder: "example.com, internal.company.com",
  },
  { key: "tunnel-all-dns", kind: "boolean" },
] satisfies DefaultGroupField[];

export const performanceFields = [
  { key: "rx-data-per-sec", kind: "number", placeholder: "100000" },
  { key: "tx-data-per-sec", kind: "number", placeholder: "200000" },
  { key: "keepalive", kind: "number", placeholder: "60" },
  { key: "dpd", kind: "number", placeholder: "90" },
  { key: "mobile-dpd", kind: "number", placeholder: "300" },
  { key: "stats-report-time", kind: "number", placeholder: "300" },
  { key: "mtu", kind: "number", placeholder: "1400" },
  { key: "idle-timeout", kind: "number", placeholder: "600" },
  { key: "mobile-idle-timeout", kind: "number", placeholder: "900" },
  { key: "session-timeout", kind: "number", placeholder: "3600" },
] satisfies DefaultGroupField[];

export const accessFields = [
  { key: "cgroup", kind: "text", placeholder: "cpuset,cpu:test" },
  { key: "max-same-clients", kind: "number", placeholder: "2" },
  { key: "net-priority", kind: "number", placeholder: "1" },
  {
    key: "restrict-user-to-ports",
    kind: "text",
    placeholder: "tcp(443),tcp(80),udp(53)",
  },
  { key: "restrict-user-to-routes", kind: "boolean" },
  { key: "deny-roaming", kind: "boolean" },
  { key: "no-udp", kind: "boolean" },
] satisfies DefaultGroupField[];

export const routingFields = [
  { key: "route", kind: "list", placeholder: "0.0.0.0/0, 10.10.0.0/16" },
  { key: "no-route", kind: "list", placeholder: "192.168.0.0/16, 10.0.0.0/8" },
  { key: "iroute", kind: "text", placeholder: "10.0.0.0/8" },
] satisfies DefaultGroupField[];

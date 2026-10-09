import type {
  ModelsOcservGroup,
  ModelsOcservGroupConfig,
} from "@/api/generated";

export const mockOcservDefaultGroupConfig = {
  cgroup: "cpuset,cpu:test",
  "deny-roaming": true,
  dns: ["8.8.8.8", "1.1.1.1"],
  dpd: 90,
  "explicit-ipv4": "192.168.100.10",
  "idle-timeout": 600,
  "ipv4-network": "192.168.1.0/24",
  iroute: "10.0.0.0/8",
  keepalive: 60,
  "max-same-clients": 2,
  "mobile-dpd": 300,
  "mobile-idle-timeout": 900,
  mtu: 1400,
  nbns: "192.168.1.1",
  "net-priority": 1,
  "no-route": ["192.168.0.0/16", "10.0.0.0/8"],
  "no-udp": false,
  "restrict-user-to-ports": "tcp(443),tcp(80),udp(53)",
  "restrict-user-to-routes": true,
  route: ["0.0.0.0/0", "10.10.0.0/16"],
  "rx-data-per-sec": 100_000,
  "session-timeout": 3600,
  "split-dns": ["example.com", "internal.company.com"],
  "stats-report-time": 300,
  "tunnel-all-dns": true,
  "tx-data-per-sec": 200_000,
} satisfies ModelsOcservGroupConfig;

export const mockOcservGroups = [
  {
    config: {
      ...mockOcservDefaultGroupConfig,
      "ipv4-network": "10.10.0.0/24",
      "max-same-clients": 3,
    },
    id: 1,
    name: "employees",
    total_rx: 93_952_409_600,
    total_tx: 41_197_731_328,
  },
  {
    config: {
      dns: ["10.20.0.2"],
      "ipv4-network": "10.20.0.0/24",
      route: ["10.0.0.0/8"],
      "tunnel-all-dns": true,
    },
    id: 2,
    name: "engineering",
    total_rx: 214_748_364_800,
    total_tx: 128_849_018_880,
  },
  {
    config: {
      "deny-roaming": true,
      "ipv4-network": "10.30.0.0/24",
      "no-udp": true,
      "session-timeout": 1800,
    },
    id: 3,
    name: "contractors",
    total_rx: 8_589_934_592,
    total_tx: 3_221_225_472,
  },
] satisfies Array<
  ModelsOcservGroup & {
    total_rx: number;
    total_tx: number;
  }
>;

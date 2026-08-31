import { afterAll, beforeAll, describe, expect, it } from "vitest";

import {
  getOcservDefaultGroupConfig,
  updateOcservDefaultGroupConfig,
} from "@/api/services/ocserv-groups";
import type { ModelsOcservGroupConfig } from "@/api/generated";

let originalConfig: ModelsOcservGroupConfig;

beforeAll(async () => {
  originalConfig = await getOcservDefaultGroupConfig();
});

afterAll(async () => {
  await updateOcservDefaultGroupConfig({ config: originalConfig });
});

describe("ocserv default group mock service", () => {
  it("returns schema-compatible default group data", async () => {
    const config = await getOcservDefaultGroupConfig();

    expect(config).toMatchObject({
      dns: ["8.8.8.8", "1.1.1.1"],
      "ipv4-network": "192.168.1.0/24",
      "max-same-clients": 2,
    });
  });

  it("persists updates for subsequent mock requests", async () => {
    await updateOcservDefaultGroupConfig({
      config: { ...originalConfig, "ipv4-network": "10.20.0.0/16" },
    });

    await expect(getOcservDefaultGroupConfig()).resolves.toMatchObject({
      "ipv4-network": "10.20.0.0/16",
    });
  });
});

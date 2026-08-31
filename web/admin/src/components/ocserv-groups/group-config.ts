import type { ModelsOcservGroupConfig } from "@/api/generated";

export function cloneOcservGroupConfig(
  config: ModelsOcservGroupConfig = {},
): ModelsOcservGroupConfig {
  return {
    ...config,
    dns: config.dns ? [...config.dns] : undefined,
    "no-route": config["no-route"] ? [...config["no-route"]] : undefined,
    route: config.route ? [...config.route] : undefined,
    "split-dns": config["split-dns"] ? [...config["split-dns"]] : undefined,
  };
}

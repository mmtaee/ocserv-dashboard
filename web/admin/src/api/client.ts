import type { AxiosInstance } from "axios";

import { apiBaseUrl, httpClient } from "@/api/http";
import {
  AuthApi,
  Configuration,
  HomeApi,
  OCCTLApi,
  OcservAgentsApi,
  OcservGroupsApi,
  OcservOcpasswdApi,
  OcservUnsyncedGroupApi,
  OcservUsersApi,
  ReportApi,
  SystemApi,
  SystemBackupApi,
  SystemRestoreApi,
  SystemUserApi,
  SystemUsersApi,
} from "@/api/generated";

type GeneratedApiConstructor<T> = new (
  configuration?: Configuration,
  basePath?: string,
  axios?: AxiosInstance,
) => T;

export const apiConfiguration = new Configuration({
  basePath: apiBaseUrl,
});

export function createGeneratedApi<T>(Api: GeneratedApiConstructor<T>): T {
  return new Api(apiConfiguration, apiBaseUrl, httpClient);
}

export const api = {
  auth: createGeneratedApi(AuthApi),
  home: createGeneratedApi(HomeApi),
  occtl: createGeneratedApi(OCCTLApi),
  agents: createGeneratedApi(OcservAgentsApi),
  groups: createGeneratedApi(OcservGroupsApi),
  ocpasswd: createGeneratedApi(OcservOcpasswdApi),
  unsyncedGroups: createGeneratedApi(OcservUnsyncedGroupApi),
  users: createGeneratedApi(OcservUsersApi),
  reports: createGeneratedApi(ReportApi),
  system: createGeneratedApi(SystemApi),
  backup: createGeneratedApi(SystemBackupApi),
  restore: createGeneratedApi(SystemRestoreApi),
  systemUser: createGeneratedApi(SystemUserApi),
  systemUsers: createGeneratedApi(SystemUsersApi),
} as const;

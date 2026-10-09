import { requireAuthorizationHeader } from "@/api/auth-token";
import { api } from "@/api/client";
import { isTestMode } from "@/api/environment";
import type { BackupRestoreResponse } from "@/api/generated";
import { downloadMockBackup, restoreMockBackup } from "@/mocks/backups";

export const backupKinds = ["ocserv_users", "ocserv_groups"] as const;
export type BackupKind = (typeof backupKinds)[number];
export type RestoreResponse = BackupRestoreResponse;

export function backupFileName(kind: BackupKind): string {
  return `${kind}_backup.json.gz`;
}

export async function downloadBackup(kind: BackupKind): Promise<Blob> {
  if (isTestMode) return downloadMockBackup(kind);
  const request = { authorization: requireAuthorizationHeader() };
  const response =
    kind === "ocserv_users"
      ? await api.backup.backupOcservUsersGet(request, {
          responseType: "blob",
        })
      : await api.backup.backupOcservGroupsGet(request, {
          responseType: "blob",
        });
  return response.data as unknown as Blob;
}

export async function restoreBackup(
  kind: BackupKind,
  file: File,
): Promise<RestoreResponse> {
  if (isTestMode) return restoreMockBackup(kind, file);
  const request = { authorization: requireAuthorizationHeader(), file };
  const response =
    kind === "ocserv_users"
      ? await api.restore.backupOcservUsersPost(request)
      : await api.restore.backupOcservGroupsPost(request);
  return response.data;
}

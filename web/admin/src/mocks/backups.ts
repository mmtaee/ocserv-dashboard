import { ApiError } from "@/api/http";
import type { BackupKind, RestoreResponse } from "@/api/services/backups";

const scenario = () => import.meta.env.VITE_BACKUPS_MOCK_SCENARIO || "normal";

async function wait(): Promise<void> {
  await new Promise((resolve) => setTimeout(resolve, 200));
  if (scenario() === "error") {
    throw new ApiError("Backup service is unavailable.", { status: 500 });
  }
}

export async function downloadMockBackup(kind: BackupKind): Promise<Blob> {
  await wait();
  return new Blob([JSON.stringify({ kind, version: 1 })], {
    type: "application/gzip",
  });
}

export async function restoreMockBackup(
  kind: BackupKind,
  file: File,
): Promise<RestoreResponse> {
  await wait();
  if (!file.name.endsWith(".json") && !file.name.endsWith(".json.gz")) {
    throw new ApiError("Select a JSON or JSON.GZ backup file.", {
      status: 400,
    });
  }
  if (scenario() === "empty") return { existing: [], inserted: [] };
  return {
    existing: kind === "ocserv_users" ? ["admin"] : ["default"],
    inserted: kind === "ocserv_users" ? ["restored-user"] : ["restored-group"],
  };
}

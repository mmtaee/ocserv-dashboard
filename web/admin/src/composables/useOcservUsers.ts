import { shallowReadonly, shallowRef } from "vue";

import { normalizeApiError } from "@/api/http";
import type {
  OcservUser,
  OcservUserActivation,
  OcservUserCreate,
  OcservUsersList,
  OcservUsersListOptions,
  OcservUserUpdate,
} from "@/api/services/ocserv-users";
import {
  activateOcservUser,
  bulkDeleteOcservUsers,
  bulkSetOcservUserGroup,
  bulkSetOcservUserStatus,
  bulkUpdateOcservUsers,
  createOcservUser,
  createOcservUserCertificate,
  deleteOcservUser,
  disconnectOcservUser,
  disconnectOcservUserSession,
  downloadOcservUserCertificate,
  getOcservUser,
  getOcservUsers,
  lockOcservUser,
  resetOcservUserUsage,
  terminateOcservUser,
  terminateOcservUserSession,
  unlockOcservUser,
  updateOcservUser,
} from "@/api/services/ocserv-users";

export type OcservUsersSuccess =
  | "activate"
  | "bulkDelete"
  | "bulkGroup"
  | "bulkStatus"
  | "bulkUpdate"
  | "certificate"
  | "create"
  | "delete"
  | "disconnect"
  | "download"
  | "lock"
  | "resetUsage"
  | "terminate"
  | "unlock"
  | "update";

export function useOcservUsers() {
  const users = shallowRef<OcservUser[]>([]);
  const meta = shallowRef<OcservUsersList["meta"]>({
    page: 1,
    size: 25,
    total_records: 0,
  });
  const filters = shallowRef<OcservUsersListOptions>({});
  const loading = shallowRef(false);
  const detailLoading = shallowRef(false);
  const mutating = shallowRef(false);
  const error = shallowRef<string | null>(null);
  const success = shallowRef<OcservUsersSuccess | null>(null);

  function clearFeedback(): void {
    error.value = null;
    success.value = null;
  }

  async function refresh(
    page = meta.value.page,
    nextFilters = filters.value,
  ): Promise<void> {
    loading.value = true;
    clearFeedback();
    filters.value = { ...nextFilters };
    try {
      const response = await getOcservUsers({
        ...nextFilters,
        page,
        size: meta.value.size,
      });
      users.value = response.result ?? [];
      meta.value = response.meta;
    } catch (cause) {
      error.value = normalizeApiError(cause).message;
    } finally {
      loading.value = false;
    }
  }

  async function loadUser(id: number): Promise<OcservUser | null> {
    detailLoading.value = true;
    error.value = null;
    try {
      return await getOcservUser(id);
    } catch (cause) {
      error.value = normalizeApiError(cause).message;
      return null;
    } finally {
      detailLoading.value = false;
    }
  }

  async function mutate(
    operation: () => Promise<unknown>,
    successType: OcservUsersSuccess,
    page = meta.value.page,
  ): Promise<boolean> {
    mutating.value = true;
    clearFeedback();
    try {
      await operation();
      await refresh(page);
      success.value = successType;
      return true;
    } catch (cause) {
      error.value = normalizeApiError(cause).message;
      return false;
    } finally {
      mutating.value = false;
    }
  }

  const create = (request: OcservUserCreate) =>
    mutate(() => createOcservUser(request), "create", 1);
  const update = (id: number, request: OcservUserUpdate) =>
    mutate(() => updateOcservUser(id, request), "update");
  const activate = (id: number, request: OcservUserActivation) =>
    mutate(() => activateOcservUser(id, request), "activate");
  const lock = (id: number) => mutate(() => lockOcservUser(id), "lock");
  const unlock = (id: number) => mutate(() => unlockOcservUser(id), "unlock");
  const resetUsage = (id: number) =>
    mutate(() => resetOcservUserUsage(id), "resetUsage");
  const createCertificate = (id: number) =>
    mutate(() => createOcservUserCertificate(id), "certificate");
  const disconnect = (username: string) =>
    mutate(() => disconnectOcservUser(username), "disconnect");
  const terminate = (username: string) =>
    mutate(() => terminateOcservUser(username), "terminate");
  const disconnectSession = (id: number) =>
    mutate(() => disconnectOcservUserSession(id), "disconnect");
  const terminateSession = (id: number) =>
    mutate(() => terminateOcservUserSession(id), "terminate");

  async function downloadCertificate(user: OcservUser): Promise<boolean> {
    mutating.value = true;
    clearFeedback();
    try {
      await downloadOcservUserCertificate(user.id, user.username);
      success.value = "download";
      return true;
    } catch (cause) {
      error.value = normalizeApiError(cause).message;
      return false;
    } finally {
      mutating.value = false;
    }
  }

  async function remove(id: number): Promise<boolean> {
    const page =
      users.value.length === 1
        ? Math.max(1, meta.value.page - 1)
        : meta.value.page;
    return mutate(() => deleteOcservUser(id), "delete", page);
  }

  async function bulkRemove(ids: number[]): Promise<boolean> {
    const page =
      ids.length >= users.value.length
        ? Math.max(1, meta.value.page - 1)
        : meta.value.page;
    return mutate(() => bulkDeleteOcservUsers(ids), "bulkDelete", page);
  }

  const bulkStatus = (ids: number[], enabled: boolean) =>
    mutate(() => bulkSetOcservUserStatus({ enabled, ids }), "bulkStatus");
  const bulkGroup = (ids: number[], group: string) =>
    mutate(() => bulkSetOcservUserGroup({ group, ids }), "bulkGroup");
  const bulkDescription = (ids: number[], description: string) =>
    mutate(
      () =>
        bulkUpdateOcservUsers({
          users: ids.map((id) => ({ changes: { description }, id })),
        }),
      "bulkUpdate",
    );

  return {
    activate,
    bulkDescription,
    bulkGroup,
    bulkRemove,
    bulkStatus,
    create,
    createCertificate,
    detailLoading: shallowReadonly(detailLoading),
    disconnect,
    disconnectSession,
    downloadCertificate,
    error: shallowReadonly(error),
    filters: shallowReadonly(filters),
    loadUser,
    loading: shallowReadonly(loading),
    lock,
    meta: shallowReadonly(meta),
    mutating: shallowReadonly(mutating),
    refresh,
    remove,
    resetUsage,
    success: shallowReadonly(success),
    terminate,
    terminateSession,
    unlock,
    update,
    users: shallowReadonly(users),
  };
}

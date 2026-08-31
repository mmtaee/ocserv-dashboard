import { shallowReadonly, shallowRef } from "vue";

import type {
  OcservGroup,
  OcservGroupCreate,
  OcservGroupsList,
  OcservGroupUpdate,
} from "@/api/services/ocserv-groups";
import {
  createOcservGroup,
  deleteOcservGroup,
  getOcservGroup,
  getOcservGroupNames,
  getOcservGroups,
  updateOcservGroup,
} from "@/api/services/ocserv-groups";
import { normalizeApiError } from "@/api/http";

export type OcservGroupsSuccess = "create" | "delete" | "update";

export function useOcservGroups() {
  const groups = shallowRef<OcservGroup[]>([]);
  const groupNames = shallowRef<string[]>([]);
  const meta = shallowRef<OcservGroupsList["meta"]>({
    page: 1,
    size: 25,
    total_records: 0,
  });
  const loading = shallowRef(false);
  const detailLoading = shallowRef(false);
  const mutating = shallowRef(false);
  const error = shallowRef<string | null>(null);
  const success = shallowRef<OcservGroupsSuccess | null>(null);

  function clearFeedback(): void {
    error.value = null;
    success.value = null;
  }

  async function refresh(page = meta.value.page): Promise<void> {
    loading.value = true;
    error.value = null;
    success.value = null;
    try {
      const [response, names] = await Promise.all([
        getOcservGroups({
          order: "name",
          page,
          size: meta.value.size,
          sort: "ASC",
        }),
        getOcservGroupNames(),
      ]);
      groups.value = response.result ?? [];
      groupNames.value = names;
      meta.value = response.meta;
    } catch (cause) {
      error.value = normalizeApiError(cause).message;
    } finally {
      loading.value = false;
    }
  }

  async function loadGroup(id: number): Promise<OcservGroup | null> {
    detailLoading.value = true;
    error.value = null;
    try {
      return await getOcservGroup(id);
    } catch (cause) {
      error.value = normalizeApiError(cause).message;
      return null;
    } finally {
      detailLoading.value = false;
    }
  }

  async function create(request: OcservGroupCreate): Promise<boolean> {
    mutating.value = true;
    clearFeedback();
    try {
      await createOcservGroup(request);
      await refresh(1);
      success.value = "create";
      return true;
    } catch (cause) {
      error.value = normalizeApiError(cause).message;
      return false;
    } finally {
      mutating.value = false;
    }
  }

  async function update(
    id: number,
    request: OcservGroupUpdate,
  ): Promise<boolean> {
    mutating.value = true;
    clearFeedback();
    try {
      await updateOcservGroup(id, request);
      await refresh();
      success.value = "update";
      return true;
    } catch (cause) {
      error.value = normalizeApiError(cause).message;
      return false;
    } finally {
      mutating.value = false;
    }
  }

  async function remove(id: number): Promise<boolean> {
    mutating.value = true;
    clearFeedback();
    try {
      await deleteOcservGroup(id);
      const page =
        groups.value.length === 1
          ? Math.max(1, meta.value.page - 1)
          : meta.value.page;
      await refresh(page);
      success.value = "delete";
      return true;
    } catch (cause) {
      error.value = normalizeApiError(cause).message;
      return false;
    } finally {
      mutating.value = false;
    }
  }

  return {
    create,
    detailLoading: shallowReadonly(detailLoading),
    error: shallowReadonly(error),
    groupNames: shallowReadonly(groupNames),
    groups: shallowReadonly(groups),
    loadGroup,
    loading: shallowReadonly(loading),
    meta: shallowReadonly(meta),
    mutating: shallowReadonly(mutating),
    refresh,
    remove,
    success: shallowReadonly(success),
    update,
  };
}

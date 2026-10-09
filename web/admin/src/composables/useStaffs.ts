import { shallowReadonly, shallowRef } from "vue";

import { normalizeApiError } from "@/api/http";
import type {
  StaffCreate,
  StaffPassword,
  StaffsList,
} from "@/api/services/staffs";
import {
  changeStaffPassword,
  createStaff,
  deleteStaff,
  getStaffs,
} from "@/api/services/staffs";

export type StaffsSuccess = "create" | "delete" | "password";

export function useStaffs() {
  const staffs = shallowRef<NonNullable<StaffsList["result"]>>([]);
  const meta = shallowRef<StaffsList["meta"]>({
    page: 1,
    size: 25,
    total_records: 0,
  });
  const loading = shallowRef(false);
  const mutating = shallowRef(false);
  const error = shallowRef<string | null>(null);
  const success = shallowRef<StaffsSuccess | null>(null);

  async function refresh(page = meta.value.page): Promise<void> {
    loading.value = true;
    error.value = null;
    try {
      const response = await getStaffs(page, meta.value.size);
      staffs.value = response.result ?? [];
      meta.value = response.meta;
    } catch (cause) {
      error.value = normalizeApiError(cause).message;
    } finally {
      loading.value = false;
    }
  }

  async function mutate(
    operation: () => Promise<unknown>,
    result: StaffsSuccess,
    page = meta.value.page,
  ): Promise<boolean> {
    mutating.value = true;
    error.value = null;
    success.value = null;
    try {
      await operation();
      await refresh(page);
      success.value = result;
      return true;
    } catch (cause) {
      error.value = normalizeApiError(cause).message;
      return false;
    } finally {
      mutating.value = false;
    }
  }

  const create = (request: StaffCreate) =>
    mutate(() => createStaff(request), "create", 1);
  const changePassword = (id: number, request: StaffPassword) =>
    mutate(() => changeStaffPassword(id, request), "password");

  function remove(id: number): Promise<boolean> {
    const page =
      staffs.value.length === 1
        ? Math.max(1, meta.value.page - 1)
        : meta.value.page;
    return mutate(() => deleteStaff(id), "delete", page);
  }

  return {
    changePassword,
    create,
    error: shallowReadonly(error),
    loading: shallowReadonly(loading),
    meta: shallowReadonly(meta),
    mutating: shallowReadonly(mutating),
    refresh,
    remove,
    staffs: shallowReadonly(staffs),
    success: shallowReadonly(success),
  };
}

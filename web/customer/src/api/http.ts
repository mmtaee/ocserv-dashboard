import axios, { AxiosError } from "axios";

import { clearAccessToken, getAccessToken } from "@/api/auth-token";

type UnauthorizedHandler = () => void | Promise<unknown>;
let unauthorizedHandler: UnauthorizedHandler | null = null;

export class ApiError extends Error {
  constructor(
    message: string,
    readonly status?: number,
    readonly data?: unknown,
  ) {
    super(message);
    this.name = "ApiError";
  }
}
export function normalizeApiError(error: unknown): ApiError {
  if (error instanceof ApiError) return error;
  if (axios.isAxiosError(error)) {
    const value = error as AxiosError<{ message?: string[]; error?: string[] }>;
    return new ApiError(
      value.response?.data?.message?.join(" ") ||
        value.response?.data?.error?.join(" ") ||
        value.message,
      value.response?.status,
      value.response?.data,
    );
  }
  return new ApiError(
    error instanceof Error ? error.message : "Request failed",
  );
}
export const httpClient = axios.create({
  baseURL: (import.meta.env.VITE_API_BASE_URL || "/api").replace(/\/$/, ""),
  timeout: Number(import.meta.env.VITE_API_TIMEOUT_MS) || 15_000,
  headers: { Accept: "application/json" },
});
export function setUnauthorizedHandler(handler: UnauthorizedHandler): void {
  unauthorizedHandler = handler;
}
httpClient.interceptors.request.use((config) => {
  const token = getAccessToken();
  if (token) config.headers.Authorization = "Bearer " + token;
  return config;
});
httpClient.interceptors.response.use(
  (response) => response,
  async (cause: unknown) => {
    const error = normalizeApiError(cause);
    if (error.status === 401) {
      clearAccessToken();
      await unauthorizedHandler?.();
    }
    return Promise.reject(error);
  },
);

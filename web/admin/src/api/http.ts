import axios, { AxiosError, type AxiosInstance } from "axios";

import { clearAccessToken, getAccessToken } from "@/api/auth-token";
import { isTestMode } from "@/api/environment";

const DEFAULT_TIMEOUT_MS = 15_000;
type UnauthorizedHandler = () => void | Promise<unknown>;

let unauthorizedHandler: UnauthorizedHandler | null = null;

export const apiBaseUrl = (import.meta.env.VITE_API_BASE_URL || "/api").replace(
  /\/$/,
  "",
);

const configuredTimeout = Number(import.meta.env.VITE_API_TIMEOUT_MS);

export class ApiError extends Error {
  readonly status?: number;
  readonly data?: unknown;

  constructor(
    message: string,
    options: { status?: number; data?: unknown; cause?: unknown } = {},
  ) {
    super(message, { cause: options.cause });
    this.name = "ApiError";
    this.status = options.status;
    this.data = options.data;
  }
}

export function normalizeApiError(error: unknown): ApiError {
  if (error instanceof ApiError) return error;

  if (axios.isAxiosError(error)) {
    const axiosError = error as AxiosError<{
      message?: string;
      error?: string;
    }>;
    const message =
      axiosError.response?.data?.message ||
      axiosError.response?.data?.error ||
      axiosError.message ||
      "The API request failed.";

    return new ApiError(message, {
      status: axiosError.response?.status,
      data: axiosError.response?.data,
      cause: error,
    });
  }

  return new ApiError(
    error instanceof Error ? error.message : "The API request failed.",
    {
      cause: error,
    },
  );
}

export function setUnauthorizedHandler(handler: UnauthorizedHandler): void {
  unauthorizedHandler = handler;
}

export const httpClient: AxiosInstance = axios.create({
  baseURL: apiBaseUrl,
  timeout:
    Number.isFinite(configuredTimeout) && configuredTimeout > 0
      ? configuredTimeout
      : DEFAULT_TIMEOUT_MS,
  headers: {
    Accept: "application/json",
  },
});

httpClient.interceptors.request.use((config) => {
  if (isTestMode) {
    return Promise.reject(
      new ApiError("Network requests are disabled while test mode is active."),
    );
  }

  const token = getAccessToken();

  if (token && !config.headers.Authorization) {
    config.headers.Authorization = token.startsWith("Bearer ")
      ? token
      : `Bearer ${token}`;
  }

  return config;
});

httpClient.interceptors.response.use(
  (response) => response,
  async (error: unknown) => {
    const apiError = normalizeApiError(error);

    if (apiError.status === 401) {
      clearAccessToken();

      try {
        await unauthorizedHandler?.();
      } catch {
        // Preserve the original API error when navigation cannot complete.
      }
    }

    return Promise.reject(apiError);
  },
);

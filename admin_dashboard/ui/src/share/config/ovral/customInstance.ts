import { attachTimeout, maybeRedirectToLogin, normalizeToBaseUrl, parseResponseBody, withAuthHeader } from "./http.config";

/**
 * Orval React Query mutator.
 * Orval calls it like: (url, { method, headers, body, ... })
 */
type RequestOptions = RequestInit & {
  timeout?: number
}

export const customInstance = async <TData>(
  url: string,
  options: RequestOptions = {}
): Promise<TData> => {
  const { timeout, ...requestOptions } = options

  const baseUrl = import.meta.env.NEXT_PUBLIC_API_URL ?? ""
  const requestUrl = normalizeToBaseUrl(url, baseUrl)

  const headers = await withAuthHeader(requestOptions.headers)

  const request = new Request(requestUrl, {
    ...requestOptions,
    headers,
  })

  const { request: timedRequest, clear } = attachTimeout(
    request,
    timeout ?? 30_000 // default: 30 seconds
  )

  try {
    const response = await fetch(timedRequest)
    const data = (await parseResponseBody(response)) as unknown

    const result = {
      data,
      status: response.status,
      headers: response.headers,
    } as unknown as TData

    if (response.status === 401 || response.status === 403) {
      maybeRedirectToLogin(response.status)
    }

    if (!response.ok) throw result

    return result
  } catch (error) {
    if (error instanceof DOMException && error.name === "AbortError") {
      throw {
        status: 0,
        data: {
          errors: {
            msg: ["Request timeout."],
          },
        },
      }
    }

    throw error
  } finally {
    clear()
  }
}
// Backward-compatible name (in case you referenced it elsewhere)
export const customFetch = customInstance

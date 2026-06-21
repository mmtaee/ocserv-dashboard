import { getToken, removeToken } from "@/share/utils/cookie"


export const DEFAULT_TIMEOUT_MS = 5_000

export const withAuthHeader = async (
  headers?: HeadersInit
): Promise<Headers> => {
  const h = new Headers(headers)
  const token = await getToken()
  if (token) h.set("Authorization", `Bearer ${token}`)
  return h
}

export const normalizeToBaseUrl = (
  contextUrl: string,
  baseUrl: string
): string => {
  if (!baseUrl) return contextUrl

  const base = baseUrl.endsWith("/") ? baseUrl : `${baseUrl}/`

  let pathname = ""
  let search = ""

  try {
    const u = new URL(contextUrl)
    pathname = u.pathname
    search = u.search
  } catch {
    try {
      const u = new URL(contextUrl, "http://orval.local")
      pathname = u.pathname
      search = u.search
    } catch {
      pathname = contextUrl
      search = ""
    }
  }

  try {
    const basePathname = new URL(base).pathname.replace(/\/+$/, "")
    if (basePathname.endsWith("/api/v1") && pathname.startsWith("/api/v1/")) {
      pathname = pathname.slice("/api/v1".length)
    } else if (basePathname.endsWith("/api/v1") && pathname === "/api/v1") {
      pathname = "/"
    }
  } catch {
    // ignore baseUrl parse issues; fallback to simple join below
  }

  const relativePath = pathname.startsWith("/") ? pathname.slice(1) : pathname
  return new URL(`${relativePath}${search}`, base).toString()
}

export const attachTimeout = (request: Request, timeout = 30_000) => {
  const controller = new AbortController()

  const id = setTimeout(() => {
    controller.abort()
  }, timeout)

  const timedRequest = new Request(request, {
    signal: controller.signal,
  })

  return {
    request: timedRequest,
    clear: () => clearTimeout(id),
  }
}

export const clearTimeoutFromRequest = (request: Request) => {
  const timeoutId = (request)?._timeoutId
  if (timeoutId) clearTimeout(timeoutId)
}

export const maybeRedirectToLogin = (status: number) => {
  if (status !== 401 && status !== 403) return
  if (typeof window === "undefined") return
  removeToken()
  window.location.href = "/login"
  return
}

export const parseResponseBody = async (response: Response) => {
  const contentType = response.headers.get("content-type") ?? ""
  if (contentType.includes("application/json")) return response.json()
  if (contentType.includes("application/pdf")) return response.blob()
  return response.text()
}

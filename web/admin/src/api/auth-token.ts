const ACCESS_TOKEN_KEY = "ocserv-dashboard.access-token";

let accessToken = readStoredToken();

function readStoredToken(): string | null {
  if (typeof window === "undefined") return null;

  return window.localStorage.getItem(ACCESS_TOKEN_KEY);
}

export function getAccessToken(): string | null {
  return accessToken;
}

export function setAccessToken(token: string): void {
  accessToken = token;

  if (typeof window === "undefined") return;

  window.localStorage.setItem(ACCESS_TOKEN_KEY, token);
}

export function clearAccessToken(): void {
  accessToken = null;

  if (typeof window !== "undefined") {
    window.localStorage.removeItem(ACCESS_TOKEN_KEY);
  }
}

export function requireAuthorizationHeader(): string {
  const token = getAccessToken();

  if (!token) {
    throw new Error("An access token is required for this API request.");
  }

  return token.startsWith("Bearer ") ? token : `Bearer ${token}`;
}

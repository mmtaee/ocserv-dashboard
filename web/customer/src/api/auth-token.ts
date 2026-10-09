const TOKEN_KEY = "ocserv-customer.access-token";
const EXPIRY_KEY = "ocserv-customer.access-token-expiry";

export function clearAccessToken(): void {
  localStorage.removeItem(TOKEN_KEY);
  localStorage.removeItem(EXPIRY_KEY);
}
export function setAccessToken(token: string, expiresAt: string): void {
  localStorage.setItem(TOKEN_KEY, token);
  localStorage.setItem(EXPIRY_KEY, expiresAt);
}
export function getAccessToken(): string | null {
  const token = localStorage.getItem(TOKEN_KEY);
  const expiresAt = Date.parse(localStorage.getItem(EXPIRY_KEY) ?? "");
  if (token && Number.isFinite(expiresAt) && expiresAt > Date.now())
    return token;
  clearAccessToken();
  return null;
}

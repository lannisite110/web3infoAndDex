/** Backend API base URL (no trailing slash). Empty = chain-only mode. */
export const API_BASE_URL = (import.meta.env.VITE_API_URL ?? "").replace(/\/$/, "");

export function isApiConfigured(): boolean {
  return API_BASE_URL.length > 0;
}

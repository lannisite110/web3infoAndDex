/** Backend API base URL (no trailing slash). Empty = same-origin /api (Vercel rewrite). */
function resolveApiBaseUrl(): string {
  const fromEnv = (import.meta.env.VITE_API_URL ?? "").replace(/\/$/, "");
  if (fromEnv) return fromEnv;
  // Production on Vercel: use vercel.json proxy → /api/v1/... (no CORS)
  if (import.meta.env.PROD) return "";
  return "http://localhost:8080";
}

export const API_BASE_URL = resolveApiBaseUrl();

export function isApiConfigured(): boolean {
  return import.meta.env.PROD || API_BASE_URL.length > 0;
}

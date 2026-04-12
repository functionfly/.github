/** Auth site URLs. */
export const SITE_ORIGIN =
  (import.meta.env.PUBLIC_AUTH_URL as string | undefined)?.replace(/\/$/, "") ||
  "https://auth.functionfly.com";

/** Orchestrator API origin. */
export const API_ORIGIN =
  (import.meta.env.PUBLIC_API_URL as string | undefined)?.replace(/\/$/, "") ||
  (import.meta.env.DEV
    ? "http://localhost:8080"
    : "https://api.functionfly.com");

function getAppOrigin(): string {
  if (import.meta.env.DEV) {
    return "http://localhost:3000";
  }
  const hostname = typeof window !== "undefined" ? window.location.hostname : "";
  if (hostname.includes("staging")) {
    return "https://app.staging.functionfly.com";
  }
  return "https://app.functionfly.com";
}

/** Dashboard origin (redirect after auth). Resolved at runtime from hostname. */
export const APP_ORIGIN = getAppOrigin();

/** Marketing site origin (for terms & privacy links). */
export const MARKETING_ORIGIN =
  (import.meta.env.PUBLIC_MARKETING_URL as string | undefined)?.replace(/\/$/, "") ||
  (import.meta.env.DEV
    ? "http://localhost:4321"
    : "https://functionfly.com");

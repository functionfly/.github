/** Auth site URLs. */
export const SITE_ORIGIN =
  (import.meta.env.PUBLIC_AUTH_URL as string | undefined)?.replace(/\/$/, "") ||
  (() => { throw new Error('PUBLIC_AUTH_URL environment variable is required'); })();

/** Orchestrator API origin. */
export const API_ORIGIN =
  (import.meta.env.PUBLIC_API_URL as string | undefined)?.replace(/\/$/, "") ||
  (() => { throw new Error('PUBLIC_API_URL environment variable is required'); })();

function getAppOrigin(): string {
  if (import.meta.env.DEV) {
    return "http://localhost:3000";
  }
  const envOrigin = import.meta.env.PUBLIC_APP_URL as string | undefined;
  if (envOrigin) return envOrigin.replace(/\/$/, "");
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
  (import.meta.env.DEV ? "http://localhost:4321" : undefined) ||
  (() => { throw new Error('PUBLIC_MARKETING_URL environment variable is required in production'); })();

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

/** Dashboard origin (redirect after auth). */
export const APP_ORIGIN =
  (import.meta.env.PUBLIC_APP_URL as string | undefined)?.replace(/\/$/, "") ||
  "https://app.functionfly.com";

/** Marketing site origin (for terms & privacy links). */
export const MARKETING_ORIGIN =
  (import.meta.env.PUBLIC_MARKETING_URL as string | undefined)?.replace(/\/$/, "") ||
  (import.meta.env.DEV
    ? "http://localhost:4321"
    : "https://functionfly.com");

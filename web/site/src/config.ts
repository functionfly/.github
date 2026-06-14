/** Public site URLs (override at build/dev time with PUBLIC_* env vars). */
export const SITE_ORIGIN =
  (import.meta.env.PUBLIC_SITE_URL as string | undefined)?.replace(/\/$/, "") ||
  "https://functionfly.com";
export const DOCS_ORIGIN =
  (import.meta.env.PUBLIC_DOCS_URL as string | undefined)?.replace(/\/$/, "") ||
  "https://docs.functionfly.com";
export const AUTH_ORIGIN =
  (import.meta.env.PUBLIC_AUTH_URL as string | undefined)?.replace(/\/$/, "") ||
  "https://auth.functionfly.com";

/** Dashboard / app origin (nav “Dashboard”, Get Started, etc.). Dev: PUBLIC_APP_URL=http://localhost:3000 */
export const APP_DASHBOARD_ORIGIN =
  (
    (import.meta.env.PUBLIC_APP_URL as string | undefined) ||
    (import.meta.env.DEV ? "http://localhost:3000" : undefined)
  )?.replace(/\/$/, "") || "https://app.functionfly.com";

/** Blog origin. Dev: http://localhost:4327, Prod: https://blog.functionfly.com */
export const BLOG_ORIGIN =
  (
    (import.meta.env.PUBLIC_BLOG_URL as string | undefined) ||
    (import.meta.env.DEV ? "http://localhost:4327" : undefined)
  )?.replace(/\/$/, "") || "https://blog.functionfly.com";

/** Status page origin. Dev: http://localhost:3001, Prod: https://status.functionfly.com */
export const STATUS_ORIGIN =
  (
    (import.meta.env.PUBLIC_STATUS_URL as string | undefined) ||
    (import.meta.env.DEV ? "http://localhost:3001" : undefined)
  )?.replace(/\/$/, "") || "https://status.functionfly.com";

/**
 * NestJS blog API base URL (no path). Set PUBLIC_BLOG_API_URL in .env for local/prod builds.
 * Build: list + post pages fetch `${BLOG_API_ORIGIN}/api/v1/blog/posts`.
 */
export const BLOG_API_ORIGIN =
  (import.meta.env.PUBLIC_BLOG_API_URL as string | undefined)?.replace(
    /\/$/,
    "",
  ) || (import.meta.env.DEV ? "http://localhost:3000" : undefined) ||
  (() => { throw new Error('PUBLIC_BLOG_API_URL environment variable is required in production'); })();

/**
 * Main API origin for newsletter and other public endpoints.
 * Dev: http://localhost:8080, Prod: https://api.functionfly.com (or same domain)
 */
export const API_ORIGIN =
  (import.meta.env.PUBLIC_API_URL as string | undefined)?.replace(/\/$/, "") ||
  (import.meta.env.DEV ? "http://localhost:8080" : undefined) ||
  (() => { throw new Error('PUBLIC_API_URL environment variable is required in production'); })();

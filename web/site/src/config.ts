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
  // Astro pages are compiled by Vite; depending on how the dev server is started,
  // env exposure can differ. Prefer import.meta.env, but fall back to process.env.
  (
    (import.meta.env.PUBLIC_APP_URL as string | undefined) ||
    process.env.PUBLIC_APP_URL ||
    (process.env.NODE_ENV === "development"
      ? "http://localhost:3000"
      : undefined)
  )?.replace(/\/$/, "") || "https://app.functionfly.com";

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

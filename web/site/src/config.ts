/** Public site URLs (override at build/dev time with PUBLIC_* env vars). */
export const SITE_ORIGIN = (import.meta.env.PUBLIC_SITE_URL as string | undefined)?.replace(/\/$/, "") || "https://functionfly.com";
export const DOCS_ORIGIN = (import.meta.env.PUBLIC_DOCS_URL as string | undefined)?.replace(/\/$/, "") || "https://docs.functionfly.com";

/** Dashboard / app origin (nav “Dashboard”, Get Started, etc.). Dev: PUBLIC_APP_URL=http://localhost:3000 */
export const APP_DASHBOARD_ORIGIN =
  (
    // Astro pages are compiled by Vite; depending on how the dev server is started,
    // env exposure can differ. Prefer import.meta.env, but fall back to process.env.
    (import.meta.env.PUBLIC_APP_URL as string | undefined) ||
    process.env.PUBLIC_APP_URL ||
    (process.env.NODE_ENV === 'development' ? 'http://localhost:3000' : undefined)
  )
    ?.replace(/\/$/, "") ||
  "https://app.functionfly.com";

/**
 * NestJS blog API base URL (no path). Set PUBLIC_BLOG_API_URL in .env for local/prod builds.
 * Build: list + post pages fetch `${BLOG_API_ORIGIN}/api/v1/blog/posts`.
 */
export const BLOG_API_ORIGIN =
  (import.meta.env.PUBLIC_BLOG_API_URL as string | undefined)?.replace(/\/$/, "") || "http://localhost:3000";

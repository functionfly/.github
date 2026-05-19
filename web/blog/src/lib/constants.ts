/**
 * Blog constants and configuration
 */

export const BLOG_API_URL =
  import.meta.env.PUBLIC_BLOG_API_URL ||
  (() => { throw new Error('PUBLIC_BLOG_API_URL environment variable is required'); })();
export const BLOG_SITE_URL =
  import.meta.env.PUBLIC_BLOG_SITE_URL ||
  (() => { throw new Error('PUBLIC_BLOG_SITE_URL environment variable is required'); })();
export const BLOG_DOMAIN =
  import.meta.env.PUBLIC_BLOG_DOMAIN ||
  (() => { throw new Error('PUBLIC_BLOG_DOMAIN environment variable is required'); })();

export const SITE_NAME = import.meta.env.PUBLIC_SITE_NAME || "FunctionFly Blog";
export const SITE_DESCRIPTION =
  import.meta.env.PUBLIC_SITE_DESCRIPTION ||
  "Tutorials, product updates, and edge computing insights from FunctionFly";

export const DEFAULT_POSTS_PER_PAGE = 10;
export const MAX_POSTS_PER_PAGE = 50;

/// <reference types="astro/client" />

interface ImportMetaEnv {
  readonly PUBLIC_MAIN_API_URL: string;
  readonly PUBLIC_BLOG_API_URL: string;
  readonly PUBLIC_BLOG_SITE_URL: string;
  readonly PUBLIC_BLOG_DOMAIN: string;
  readonly PUBLIC_SITE_NAME: string;
  readonly PUBLIC_SITE_DESCRIPTION: string;
}

interface ImportMeta {
  readonly env: ImportMetaEnv;
}

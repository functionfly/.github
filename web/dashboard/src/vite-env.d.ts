/// <reference types="vite/client" />

interface ImportMetaEnv {
  readonly VITE_API_URL: string;
  readonly VITE_APP_NAME: string;
  /** When "true", production builds show only the launch page (with VITE_COMING_SOON_ONLY). */
  readonly VITE_COMING_SOON_ONLY?: string;
  /** Dev only: set with VITE_COMING_SOON_ONLY to preview launch mode on localhost. */
  readonly VITE_COMING_SOON_IN_DEV?: string;
  /** Public marketing/site origin for app URL previews (e.g. https://functionfly.com). */
  readonly VITE_PUBLIC_SITE_URL?: string;
  /** Local Astro marketing app (e.g. http://localhost:4321). When set, dev "/" redirects here instead of embedded landing. */
  readonly VITE_MARKETING_DEV_URL?: string;
  /** Standalone docs origin (e.g. https://docs.functionfly.com or http://localhost:4322 for web/docs). */
  readonly VITE_DOCS_SITE_URL?: string;
  readonly VITE_AI_SERVICE_URL?: string;
  readonly VITE_SANITY_PROJECT_ID?: string;
  readonly VITE_SANITY_DATASET?: string;
  readonly VITE_VERCEL_ANALYTICS?: string;
}

declare module '@vercel/analytics/react' {
  import type { ComponentType } from 'react';
  export interface AnalyticsProps {
    beforeSend?: (event: { type: string; url: string }) => { type: string; url: string } | null;
    debug?: boolean;
    mode?: 'auto' | 'development' | 'production';
    scriptSrc?: string;
    endpoint?: string;
    dsn?: string;
  }
  export const Analytics: ComponentType<AnalyticsProps>;
}

interface ImportMeta {
  readonly env: ImportMetaEnv;
}

declare module '@sanity/block-content-to-html' {
  interface BlockContentToHtmlOptions {
    components?: {
      types?: Record<string, (props: any) => string>;
      marks?: Record<string, (children: string, value: any) => string>;
      block?: Record<string, (children: string, value: any) => string>;
      list?: Record<string, (children: string) => string>;
      listItem?: (children: string) => string;
    };
  }

  export function toHTML(blocks: any[], options?: BlockContentToHtmlOptions): string;
}

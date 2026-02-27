/// <reference types="vite/client" />

interface ImportMetaEnv {
  readonly VITE_API_URL: string;
  readonly VITE_APP_NAME: string;
  readonly VITE_SANITY_PROJECT_ID?: string;
  readonly VITE_SANITY_DATASET?: string;
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

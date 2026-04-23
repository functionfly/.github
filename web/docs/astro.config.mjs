import react from "@astrojs/react";
import starlight from "@astrojs/starlight";
import starlightOpenAPI from "starlight-openapi";
import { defineConfig } from "astro/config";

// https://astro.build/config
export default defineConfig({
  site: "https://docs.functionfly.com",
  integrations: [
    starlight({
      title: "FunctionFly Docs",
      logo: {
        src: "./public/favicon.svg",
      },
      customCss: ["./src/styles/custom.css"],
      head: [],
      social: [
        {
          icon: "github",
          label: "GitHub",
          href: "https://github.com/functionfly",
        },
      ],
      sidebar: [
        {
          label: "Getting Started",
          items: [
            { label: "Welcome", link: "/" },
            { label: "Getting Started", link: "/getting-started/" },
            {
              label: "Quick Start",
              link: "/quick-start/",
              badge: { text: "New", variant: "tip" },
            },
          ],
        },
        {
          label: "Guides",
          badge: { text: "New", variant: "tip" },
          collapsed: false,
          items: [
            { label: "Overview", link: "/guides/" },
            { label: "Creating Functions", link: "/guides/creating-functions/" },
            { label: "Using the Registry", link: "/guides/using-registry/" },
            { label: "Authentication", link: "/guides/authentication/" },
            { label: "Secrets & Vault", link: "/guides/secrets-vault/" },
            { label: "StateFabric & Edge State", link: "/guides/statefabric/" },
            { label: "Rate Limiting", link: "/guides/rate-limiting/" },
            { label: "Webhooks", link: "/guides/webhooks/" },
            { label: "Deploy Keys", link: "/deploy-keys/" },
            { label: "Function Webhooks", link: "/function-webhooks/" },
            { label: "CI/CD Integration", link: "/guides/ci-cd/" },
          ],
        },
        {
          label: "API",
          badge: { text: "New", variant: "tip" },
          collapsed: false,
          items: [
            {
              label: "REST API",
              collapsed: true,
              items: [
                { label: "Overview", link: "/api-reference/" },
                { label: "Authentication", link: "/api/authentication/" },
                { label: "Functions", link: "/api/functions/" },
                { label: "Execution", link: "/api/execution/" },
              ],
            },
            { label: "Trust API", link: "/trust-api/" },
          ],
        },
        {
          label: "Core Concepts",
          collapsed: false,
          items: [
            { label: "Functions", link: "/functions/" },
            { label: "CLI", link: "/cli/" },
            { label: "Deployment", link: "/deployment/" },
          ],
        },
        {
          label: "Runtime & SDK",
          collapsed: false,
          items: [
            {
              label: "SDKs",
              collapsed: true,
              items: [
                { label: "Python SDK", link: "/sdks/python/" },
                {
                  label: "JavaScript/TypeScript SDK",
                  link: "/sdks/javascript/",
                },
                { label: "Go SDK", link: "/sdks/go/" },
              ],
            },
            {
              label: "Runtimes",
              collapsed: true,
              items: [
                { label: "Python", link: "/runtimes/python/" },
                { label: "JavaScript", link: "/runtimes/javascript/" },
                { label: "TypeScript", link: "/runtimes/typescript/" },
                { label: "Go", link: "/runtimes/go/" },
                { label: "Rust via WASM", link: "/runtimes/rust-wasm/" },
              ],
            },
            { label: "WASM & WebAssembly", link: "/wasm/" },
          ],
        },
        {
          label: "Platform Features",
          collapsed: false,
          items: [
            { label: "Registry", link: "/registry/" },
            { label: "Execution", link: "/execution/" },
            { label: "Analytics", link: "/analytics/" },
            { label: "Secrets & Vault", link: "/secrets-vault/" },
            { label: "StateFabric", link: "/guides/statefabric/" },
          ],
        },
        {
          label: "Pricing & Plans",
          collapsed: false,
          items: [
            {
              label: "Pricing Overview",
              link: "/pricing/",
              badge: { text: "Updated", variant: "tip" },
            },
          ],
        },
        {
          label: "Trust & Security",
          badge: { text: "Beta", variant: "caution" },
          collapsed: false,
          items: [
            { label: "Trust API Guide", link: "/trust-api/" },
            { label: "Trust API Reference", link: "/trust-api/" },
            { label: "Trust Protocol Spec", link: "/trust-protocol-spec/" },
            {
              label: "Security",
              link: "/security/",
              badge: { text: "Updated", variant: "success" },
            },
            {
              label: "Trust Protocol Open Source",
              link: "/trust-protocol-open-source/",
            },
            { label: "Open Source Strategy", link: "/open-source-strategy/" },
          ],
        },
      ],
      plugins: [
        starlightOpenAPI([
          {
            base: "api-reference",
            schema: "./src/content/api/openapi.yaml",
            sidebar: {
              label: "API Reference",
              collapsed: false,
            },
          },
        ]),
      ],
    }),
    react({
      include: ["**/*.jsx", "**/*.tsx"],
    }),
  ],
  output: "static",
  server: {
    host: true,
    port: 4322,
    strictPort: false,
  },
  build: {
    format: "directory",
  },
  vite: {
    build: {
      cssMinify: true,
      minify: false,
      target: "es2022",
    },
    esbuild: {
      target: "esnext",
    },
    optimizeDeps: {
      esbuildOptions: {
        target: "es2022",
      },
    },
    server: {
      proxy: {
        '/v1': {
          target: 'http://localhost:8080',
          changeOrigin: true,
          secure: false,
        },
        '/health': {
          target: 'http://localhost:8080',
          changeOrigin: true,
          secure: false,
        },
        '/swagger': {
          target: 'http://localhost:8080',
          changeOrigin: true,
          secure: false,
        },
      },
    },
  },
});

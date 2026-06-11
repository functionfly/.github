import react from "@astrojs/react";
import sentry from "@sentry/astro";
import starlight from "@astrojs/starlight";
import starlightOpenAPI from "starlight-openapi";
import starlightLlmsTxt from "starlight-llms-txt";
import expressiveCodeCollapsible from "expressive-code-collapsible";
import astroMermaid from "astro-mermaid";
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
            { label: "Using the Registry", link: "/registry/" },
            { label: "Registry Guide", link: "/guides/registry-guide/" },
            { label: "Authentication", link: "/guides/authentication/" },
            { label: "Secrets Vault Guide", link: "/guides/secrets-vault-guide/" },
            { label: "Rate Limiting", link: "/guides/rate-limiting/" },
            { label: "Webhooks", link: "/guides/webhooks/" },
            { label: "Deploy Keys", link: "/deploy-keys/" },
            { label: "Trust and Verification", link: "/trust-and-verification/" },
            { label: "Function Webhooks", link: "/function-webhooks/" },
            { label: "CI/CD Integration", link: "/guides/ci-cd/" },
            { label: "Environment Variables", link: "/guides/environment-variables/" },
            { label: "Error Codes & Troubleshooting", link: "/guides/error-codes/" },
            { label: "Billing & Subscription", link: "/guides/billing/" },
            { label: "Organizations & Teams", link: "/guides/organizations/" },
            { label: "Custom Domains", link: "/guides/custom-domains/" },
            { label: "Getting Started with Bundles", link: "/guides/bundles/" },
            { label: "Monitoring & Observability", link: "/guides/monitoring/" },
            { label: "Function DNA", link: "/guides/function-dna/" },
            { label: "Function Runtime Graph (FRG)", link: "/guides/frg/" },
          ],
        },
        {
          label: "Agents",
          badge: { text: "New", variant: "tip" },
          collapsed: false,
          items: [
            { label: "Overview", link: "/agents/" },
            { label: "Creating Your First Agent", link: "/guides/creating-agents/" },
            { label: "Agent Marketplace", link: "/agents/marketplace/" },
            { label: "SDK Integration", link: "/agents/sdk/" },
            { label: "Agent Security", link: "/agents/security/" },
            { label: "Behavioral Policies", link: "/agents/policies/" },
            { label: "Agent Memory", link: "/agents/memory/" },
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
                { label: "Rust SDK", link: "/sdks/rust/" },
                { label: "C SDK", link: "/sdks/c/" },
                { label: "Ruby SDK", link: "/sdks/ruby/" },
                { label: "Kotlin SDK", link: "/sdks/kotlin/" },
                { label: "Swift SDK", link: "/sdks/swift/" },
              ],
            },
            {
              label: "Runtimes",
              collapsed: true,
              items: [
                { label: "Python", link: "/runtimes/python/" },
                { label: "JavaScript", link: "/runtimes/javascript/" },
                { label: "TypeScript", link: "/runtimes/typescript/" },
                { label: "Bun", link: "/runtimes/bun/" },
                { label: "Deno", link: "/runtimes/deno/" },
                { label: "Go", link: "/runtimes/go/" },
                { label: "Rust via WASM", link: "/runtimes/rust-wasm/" },
                { label: "C/C++ via WASM", link: "/runtimes/c/" },
                { label: "Ruby", link: "/runtimes/ruby/" },
                { label: "Kotlin via WASM", link: "/runtimes/kotlin/" },
                { label: "Swift via WASM", link: "/runtimes/swift/" },
                { label: "Prism", link: "/runtimes/prism/" },
              ],
            },
            { label: "WASM & WebAssembly", link: "/wasm/" },
          ],
        },
        {
          label: "Developer",
          badge: { text: "New", variant: "tip" },
          collapsed: false,
          items: [
            { label: "CLI Reference", link: "/cli/" },
            { label: "Deployment Guide", link: "/deployment-guide/" },
            { label: "Error Codes", link: "/guides/error-codes/" },
            {
              label: "Tools",
              collapsed: true,
              items: [
                { label: "Studio", link: "/studio/" },
                { label: "Studio Plugins", link: "/studio-plugins/" },
                { label: "Trust API for AI Models", link: "/guides/trust-api-ai-models/" },
              ],
            },
            {
              label: "Runtimes",
              collapsed: true,
              items: [
                { label: "Python", link: "/runtimes/python/" },
                { label: "JavaScript", link: "/runtimes/javascript/" },
                { label: "TypeScript", link: "/runtimes/typescript/" },
                { label: "Bun", link: "/runtimes/bun/" },
                { label: "Deno", link: "/runtimes/deno/" },
                { label: "Go", link: "/runtimes/go/" },
                { label: "Rust via WASM", link: "/runtimes/rust-wasm/" },
                { label: "Ruby", link: "/runtimes/ruby/" },
                { label: "Prism", link: "/runtimes/prism/" },
              ],
            },
            {
              label: "SDKs",
              collapsed: true,
              items: [
                { label: "Python SDK", link: "/sdks/python/" },
                { label: "JavaScript/TypeScript SDK", link: "/sdks/javascript/" },
                { label: "Go SDK", link: "/sdks/go/" },
                { label: "Rust SDK", link: "/sdks/rust/" },
              ],
            },
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
            { label: "StateFabric", link: "/statefabric/" },
          ],
        },
        {
          label: "Providers",
          collapsed: false,
          items: [
            { label: "Overview", link: "/providers/" },
            { label: "Environment Variables", link: "/providers/environment/" },
            { label: "FunctionFly Edge", link: "/providers/functionfly-edge/environment/" },
            { label: "Vercel", link: "/providers/vercel/environment/" },
            { label: "Cloudflare Workers", link: "/providers/cloudflare-workers/environment/" },
            { label: "Fly.io", link: "/providers/fly-io/environment/" },
            { label: "AWS Lambda", link: "/providers/aws-lambda/environment/" },
            { label: "Deno Deploy", link: "/providers/deno-deploy/environment/" },
          ],
        },
        {
          label: "Bundles",
          collapsed: false,
          items: [
            { label: "Overview", link: "/bundles/" },
            { label: "SaaS Starter", link: "/bundles/saas-starter/" },
            { label: "Marketplace", link: "/bundles/marketplace/" },
            { label: "AI App", link: "/bundles/ai-app/" },
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
            {
              label: "Features by Tier",
              link: "/pricing/features-by-tier/",
              badge: { text: "New", variant: "tip" },
            },
            {
              label: "Roadmap",
              link: "/roadmap/",
              badge: { text: "New", variant: "tip" },
            },
          ],
        },
        {
          label: "Trust & Security",
          badge: { text: "Beta", variant: "caution" },
          collapsed: false,
          items: [
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
    sentry({
      dsn: process.env.SENTRY_DSN,
      tracesSampleRate: 0.1,
    }),
    astroMermaid(),
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
    ssr: {
      noExternal: ["starlight-llms-txt", "starlight-package-managers", "expressive-code-collapsible"],
    },
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

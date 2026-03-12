# Blog architecture decision

## Should we create a separate Astro app for the blog?

**Recommendation: No. Use the existing Astro marketing site (`web/site`) for the public blog.**

**Status: Migrated.** The public blog now lives in `web/site` (Astro). The old redirect from `/blog` to the app has been removed; list and post pages fetch from the NestJS blog API.

---

## Current setup

| Piece | Location | Role |
|-------|----------|------|
| **Blog API** | `cmd/blog-api` (NestJS) | Content store: posts, categories, authors. CRUD + public read API. |
| **Admin UI** | `web/admin-dashboard` (React) | Manage posts/categories/settings. Calls blog API (or Go content API). |
| **Marketing site** | `web/site` (Astro) | Landing, marketing, docs. `/blog` is the public blog (list + post pages, data from NestJS). |

## Option A: Separate Astro app (e.g. `web/blog`)

- **Pros:** Isolated codebase; could be a different subdomain (e.g. `blog.functionfly.com`).
- **Cons:** Two Astro apps to maintain, two deployments, split design/nav, weaker SEO (blog on subdomain vs same domain).

## Option B: Blog inside existing Astro site (`web/site`) — **recommended**

- **Pros:**
  - One Astro app: same layout, nav, and deployment as the rest of the marketing site.
  - Blog lives at `functionfly.com/blog` (same domain = better SEO).
  - Astro is already there; add pages under `web/site/src/pages/blog/` that fetch from the NestJS blog API (SSG at build time or SSR).
- **Cons:** Blog and marketing share the same repo and release cycle (usually acceptable).

## Option C: Blog only in the dashboard app (React)

- **Pros:** Single app; current redirect already sends users to the app blog.
- **Cons:** Blog is behind the app domain and auth context; less ideal for SEO and for unauthenticated readers who expect a marketing-style blog.

---

## Decision

- **Do not** add a separate Astro app for the blog.
- **Do** implement the public blog in the existing Astro site (`web/site`):
  - Replace the current `/blog` redirect with real pages: e.g. `blog/index.astro` (list) and `blog/[...slug].astro` or `blog/[slug].astro` (post).
  - Fetch content from the NestJS blog API (`cmd/blog-api`): e.g. `GET /api/v1/blog/posts`, `GET /api/v1/blog/posts/:slug`, `GET /api/v1/blog/categories`.
  - Use SSG (fetch at build) or SSR (fetch on request) depending on how often you publish.
- **Keep** blog admin in the React admin dashboard; it continues to use the NestJS blog API (or Go content API) for CRUD.

## Summary

| Concern | Choice |
|--------|--------|
| Separate Astro app for blog? | **No** |
| Where does the public blog live? | **`web/site`** (Astro), under `/blog` |
| Where is content stored? | **NestJS blog API** (`cmd/blog-api`) |
| Where is blog managed? | **Admin dashboard** (React), Blog tab |

This gives one content source (NestJS), one public blog front (Astro marketing site), and one admin surface (dashboard), without a second Astro app.

---

## Building the site (blog)

The Astro site fetches blog data at **build time** (static). Set the blog API base URL:

```bash
# web/site or root
PUBLIC_BLOG_API_URL=http://localhost:3000  # or your NestJS blog API URL
```

Then run the blog API (e.g. `cd cmd/blog-api && bun run start:dev`) before building the site so posts are included. For production, set `PUBLIC_BLOG_API_URL` to your deployed blog API URL in CI.

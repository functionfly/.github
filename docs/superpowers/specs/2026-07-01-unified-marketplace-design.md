# Unified Marketplace Design

**Date:** 2026-07-01  
**Status:** Draft  
**Scope:** Merge Extension Marketplace and Agent/Function Marketplace into a single `/marketplace` route

---

## Problem

The dashboard has two separate marketplace experiences that fragment discovery:

1. **Extension Marketplace** — embedded in Studio (`/studio`), uses mock data, backed by `internal/api/handlers/marketplace/` + `internal/storage/marketplace_repository.go`. API client at `web/dashboard/src/api/marketplace.ts`.

2. **Agent & Function Marketplace** — standalone pages at `/marketplace` and `/functions/discovery`, backed by `internal/agent/marketplace/service.go` + `swarm.go`. Frontend has `AgentsMarketplacePage` (renders `AgentMarketplaceView`), `AgentMarketplacePage` (renders `AgentMarketplace` from swarm), `FunctionMarketplacePage` (renders `BrowseFunctionsView`), and `AgentMarketplaceDetailPage`.

Additionally, there's a `MarketplaceDropdown` in the navbar with 4 separate links, and a `MarketplaceEconomyPage` at `/marketplace-economy` for creator monetization (no backend yet).

Users hit `/marketplace` and only see agents. They need to go to `/studio` to find extensions, and `/functions/discovery` for functions. There's no unified search across types.

---

## Goal

One `/marketplace` page with:
- Unified search bar
- Type filter chips (All, Agents, Extensions, Functions)
- Single results grid with visually distinct cards per type
- Detail pages as nested routes (`/marketplace/agents/:id`, `/marketplace/extensions/:id`)

---

## Architecture

### URL Scheme

| Route | Component | Purpose |
|-------|-----------|---------|
| `/marketplace` | `MarketplacePage` | Unified search + filter + grid |
| `/marketplace?type=agents` | `MarketplacePage` | Agents only (URL-synced chip) |
| `/marketplace?type=extensions` | `MarketplacePage` | Extensions only |
| `/marketplace?type=functions` | `MarketplacePage` | Functions only |
| `/marketplace/agents/:id` | `AgentMarketplaceDetailPage` | Agent detail (unchanged) |
| `/marketplace/extensions/:id` | `ExtensionDetailPage` | Extension detail (new) |

### Backend

**Approach:** Keep existing backends as-is. Add a thin unified search proxy endpoint.

#### New Endpoint: `GET /marketplace/search`

**Handler:** `internal/api/handlers/marketplace/handler.go` — add `HandleUnifiedSearch` method.

**Query params:**
- `q` — search query string
- `type` — filter: `agent`, `extension`, `function`, or empty (all)
- `limit` — max results per type (default 20)
- `offset` — pagination offset per type

**Response:**
```json
{
  "items": [
    {
      "type": "agent",
      "id": "...",
      "name": "...",
      "description": "...",
      "icon_url": null,
      "rating": 4.5,
      "install_count": 1200,
      "price": "$0.001/call",
      "pricing_model": "per_call",
      "tags": ["code_generation", "analysis"],
      "verified": true,
      "metadata": { ... }
    },
    {
      "type": "extension",
      "id": "...",
      "name": "...",
      "description": "...",
      "icon_url": "...",
      "rating": 4.8,
      "install_count": 45200,
      "price": "Free",
      "pricing_model": "free",
      "tags": ["integrations", "github"],
      "verified": true,
      "metadata": { ... }
    }
  ],
  "total": 42,
  "has_more": true
}
```

**Implementation:** The handler calls `MarketplaceRepository.List()` (extensions), `marketplace.Service.SearchAgents()` (agents), and `marketplace.Service.SearchFunctions()` (functions) in parallel via goroutines, normalizes results into `UnifiedItem`, merges, sorts by relevance, and returns. The function search is backed by `function_listings` table via GORM (same service as agents).

#### Existing Endpoints (unchanged)

All existing endpoints stay as-is for CRUD and mutations:
- `GET/POST/PUT/DELETE /marketplace/extensions/*` — extension CRUD
- `GET/POST /v1/marketplace/agents`, `/v1/marketplace/agent/list`, `/v1/marketplace/hire`, `/v1/marketplace/purchase` — agent marketplace

### Frontend

#### New Files

| File | Purpose |
|------|---------|
| `web/dashboard/src/pages/MarketplacePage.tsx` | Top-level unified page |
| `web/dashboard/src/components/marketplace/MarketplaceSearchBar.tsx` | Search input with type filter chips |
| `web/dashboard/src/components/marketplace/MarketplaceGrid.tsx` | Unified results grid |
| `web/dashboard/src/components/marketplace/AgentCard.tsx` | Agent-specific card renderer |
| `web/dashboard/src/components/marketplace/ExtensionCard.tsx` | Extension-specific card renderer |
| `web/dashboard/src/components/marketplace/FunctionCard.tsx` | Function-specific card renderer |
| `web/dashboard/src/components/marketplace/types.ts` | Shared types (`UnifiedItem`, etc.) |
| `web/dashboard/src/pages/ExtensionDetailPage.tsx` | Extension detail page |
| `web/dashboard/src/api/marketplace-unified.ts` | API client for `/marketplace/search` |

#### Modified Files

| File | Change |
|------|--------|
| `web/dashboard/src/App.tsx` | Replace marketplace routes with unified `MarketplacePage` |
| `web/dashboard/src/lib/constants.ts` | Clean up route constants |
| `web/dashboard/src/components/common/MarketplaceDropdown.tsx` | Replace dropdown with simple link to `/marketplace` |
| `web/dashboard/src/components/layout/Sidebar/navigation.tsx` | Update marketplace nav entry (promote from "Advanced" to primary) |

#### Removed Files

| File | Reason |
|------|--------|
| `web/dashboard/src/pages/AgentsMarketplacePage.tsx` | Replaced by `MarketplacePage` |
| `web/dashboard/src/pages/AgentMarketplacePage.tsx` | Replaced by `MarketplacePage` |
| `web/dashboard/src/pages/FunctionMarketplacePage.tsx` | Replaced by `MarketplacePage` |

`AgentMarketplaceDetailPage.tsx` stays — it becomes a nested route under `/marketplace/agents/:id`.

#### Component Design

**MarketplacePage:**
- Renders `MetaTags`, `PageGrid`, `MarketplaceSearchBar`, `MarketplaceGrid`
- Reads `type` from URL search params, syncs chip selection
- Fetches from `GET /marketplace/search?q=...&type=...` on mount and on search/filter change

**MarketplaceSearchBar:**
- Search input with icon
- Type filter chips: All, Agents, Extensions, Functions
- Optional sort dropdown (Trending, Top Rated, Newest, Most Installed)

**MarketplaceGrid:**
- Responsive CSS grid (`repeat(auto-fill, minmax(320px, 1fr))`)
- Renders `AgentCard`, `ExtensionCard`, or `FunctionCard` based on `item.type`
- Loading skeleton, empty state, error state
- Pagination

**AgentCard:**
- Based on existing `AgentMarketplaceView` card design
- Shows: name, agent ID, description, capabilities tags, pricing, rating, ROI, calls, rank score
- Action: "Hire Agent" button → hire modal

**ExtensionCard:**
- Based on existing `MarketplaceHome` card design
- Shows: name, author, description, category badge, downloads, rating, trust score, version
- Action: "Install" button → calls `marketplaceApi.install()`

**FunctionCard:**
- Based on existing `BrowseFunctionsView` card design
- Shows: name, author, description, runtime badge, pricing, call volume
- Action: "Deploy" or "Try" button

---

## Data Flow

```
User types search → MarketplacePage state → API call to GET /marketplace/search
                                                        ↓
                                    ┌───────────────────┴───────────────────┐
                                    ↓                                       ↓
                          MarketplaceRepository.List()          marketplace.Service.SearchAgents()
                          (extensions from Postgres)            (agents from Postgres via GORM)
                                    ↓                                       ↓
                                    └───────────────────┬───────────────────┘
                                                        ↓
                                              Normalize → UnifiedItem[]
                                              Merge → Sort by relevance
                                              Return → MarketplaceGrid renders cards
```

---

## Migration Steps

### Phase 1: Backend (unified search endpoint)

1. Add `UnifiedItem` struct and `HandleUnifiedSearch` to `internal/api/handlers/marketplace/handler.go`
2. Register `GET /marketplace/search` in `internal/api/routes_marketplace.go`
3. Handler queries both repositories in parallel, normalizes, merges

### Phase 2: Frontend (unified page)

1. Create `web/dashboard/src/components/marketplace/types.ts` with `UnifiedItem` type
2. Create `web/dashboard/src/api/marketplace-unified.ts` API client
3. Create card components: `AgentCard`, `ExtensionCard`, `FunctionCard`
4. Create `MarketplaceSearchBar` and `MarketplaceGrid`
5. Create `MarketplacePage` that composes them
6. Create `ExtensionDetailPage` for extension detail view

### Phase 3: Route migration

1. Update `App.tsx` routes:
   - `/marketplace` → `MarketplacePage`
   - `/marketplace/agents/:id` → `AgentMarketplaceDetailPage` (unchanged)
   - `/marketplace/extensions/:id` → `ExtensionDetailPage` (new)
   - Remove `/marketplace/agents`, `/functions/discovery` routes
   - Add server-side redirects: `/marketplace/agents` → `/marketplace?type=agents`, `/functions/discovery` → `/marketplace?type=functions`
2. Update `ROUTES` constants
3. Replace `MarketplaceDropdown` with simple link
4. Update sidebar navigation (promote to primary section)

### Phase 4: Cleanup

1. Remove `AgentsMarketplacePage.tsx`, `AgentMarketplacePage.tsx`, `FunctionMarketplacePage.tsx`
2. Remove `AgentMarketplaceView.tsx` (logic moved to `AgentCard`)
3. Wire extension marketplace from mock data to real API
4. Update all internal links pointing to old marketplace routes

---

## Out of Scope

- **Employee Talent Marketplace** (`/marketplace/opportunities`) — separate HR system, stays independent
- **MarketplaceEconomyPage** (`/marketplace-economy`) — creator economy dashboard, stays independent
- **Marketplace Bundle** (billing) — product bundle for tenant-facing marketplaces, unrelated
- **Backend storage unification** — extensions and agents use different DB tables and ORMs, no need to merge

---

## Risks

| Risk | Mitigation |
|------|------------|
| Extension marketplace uses mock data | Phase 2 wires to real `marketplaceApi.list()` |
| Two different ORMs (raw SQL vs GORM) | Unified search handler abstracts over both via interfaces |
| Breaking existing bookmarks/links | Add redirects from old routes to `/marketplace` |
| Performance of parallel queries | Both queries are lightweight; add caching if needed |

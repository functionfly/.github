# Unified Marketplace — Implementation Plan

**Spec:** `docs/superpowers/specs/2026-07-01-unified-marketplace-design.md`

---

## Phase 1: Backend — Unified Search Endpoint

### 1.1 Add `UnifiedItem` struct and `HandleUnifiedSearch`

**File:** `internal/api/handlers/marketplace/handler.go`

- Add `UnifiedItem` struct: `Type`, `ID`, `Name`, `Description`, `IconURL`, `Rating`, `InstallCount`, `Price`, `PricingModel`, `Tags`, `Verified`, `Metadata`
- Add `UnifiedSearchResponse` struct: `Items []UnifiedItem`, `Total int`, `HasMore bool`
- Add `HandleUnifiedSearch(w, r)` method on `Handler`
- Accepts query params: `q`, `type`, `limit`, `offset`
- When `type` is empty or contains multiple types, queries all three sources in parallel via `errgroup`
- Normalizes each result into `UnifiedItem`
- Merges and returns

**Dependencies:** Handler needs access to agent marketplace service. Add `AgentSearcher` interface to `Handler`:
```go
type AgentSearcher interface {
    SearchAgents(ctx context.Context, req *SearchAgentsRequest) ([]AgentSearchResult, int64, error)
    SearchFunctions(ctx context.Context, req *SearchFunctionsRequest) ([]FunctionSearchResult, int64, error)
}
```
Update `NewHandler` to accept optional `AgentSearcher`.

### 1.2 Register route

**File:** `internal/api/routes_marketplace.go`

- Add `GET /marketplace/search` with optional auth (same pattern as `HandleListExtensions`)

### 1.3 Wire agent marketplace service

**File:** `internal/api/routes.go` (or wherever handlers are instantiated)

- Pass `marketplace.NewService(db)` as `AgentSearcher` to the marketplace handler

---

## Phase 2: Frontend — Types and API Client

### 2.1 Unified types

**File:** `web/dashboard/src/components/marketplace/types.ts` (new)

```typescript
export type MarketplaceItemType = 'agent' | 'extension' | 'function';

export interface UnifiedItem {
  type: MarketplaceItemType;
  id: string;
  name: string;
  description: string;
  icon_url: string | null;
  rating: number;
  install_count: number;
  price: string;
  pricing_model: string;
  tags: string[];
  verified: boolean;
  metadata: Record<string, unknown>;
}

export interface UnifiedSearchResponse {
  items: UnifiedItem[];
  total: number;
  has_more: boolean;
}
```

### 2.2 API client

**File:** `web/dashboard/src/api/marketplace-unified.ts` (new)

- `searchMarketplace(params: { q?: string; type?: string; limit?: number; offset?: number })` → `GET /marketplace/search`
- Uses existing `apiClient`

---

## Phase 3: Frontend — Components

### 3.1 Card components

**File:** `web/dashboard/src/components/marketplace/AgentCard.tsx` (new)
- Extract card design from existing `AgentMarketplaceView` (lines 122-186)
- Props: `item: UnifiedItem`, `onAction: () => void`
- Shows: name, agent ID, description, capabilities tags, pricing, rating, ROI, calls
- Action button: "Hire Agent"

**File:** `web/dashboard/src/components/marketplace/ExtensionCard.tsx` (new)
- Extract card design from existing `MarketplaceHome` (lines 192-286)
- Props: `item: UnifiedItem`, `onAction: () => void`
- Shows: name, author, description, category badge, downloads, rating, trust score, version
- Action button: "Install"

**File:** `web/dashboard/src/components/marketplace/FunctionCard.tsx` (new)
- Based on `BrowseFunctionsView` card design
- Props: `item: UnifiedItem`, `onAction: () => void`
- Shows: name, author, description, runtime badge, pricing, call volume
- Action button: "Deploy"

### 3.2 Search bar with type chips

**File:** `web/dashboard/src/components/marketplace/MarketplaceSearchBar.tsx` (new)
- Search input with `Search` icon
- Type filter chips: All, Agents, Extensions, Functions
- Optional sort dropdown
- Props: `query`, `selectedType`, `onQueryChange`, `onTypeChange`, `onSortChange`

### 3.3 Grid

**File:** `web/dashboard/src/components/marketplace/MarketplaceGrid.tsx` (new)
- Responsive grid: `repeat(auto-fill, minmax(320px, 1fr))`
- Renders `AgentCard`, `ExtensionCard`, or `FunctionCard` based on `item.type`
- Loading skeleton, empty state, error state
- Pagination controls

### 3.4 Main page

**File:** `web/dashboard/src/pages/MarketplacePage.tsx` (new)
- Reads `type` and `q` from URL search params
- Uses `searchMarketplace` API client
- Composes `MarketplaceSearchBar` + `MarketplaceGrid`
- Updates URL on filter/search change
- Hire modal (extract from `AgentMarketplaceView`)

### 3.5 Extension detail page

**File:** `web/dashboard/src/pages/ExtensionDetailPage.tsx` (new)
- Fetches extension by ID via `marketplaceApi.get()`
- Shows full detail: name, description, screenshots, ratings, install button
- Route: `/marketplace/extensions/:id`

---

## Phase 4: Route Migration

### 4.1 Update App.tsx routes

**File:** `web/dashboard/src/App.tsx`

Replace:
```tsx
<Route path="marketplace" element={<AgentsMarketplacePage />} />
<Route path="marketplace/agents" element={<AgentsMarketplacePage />} />
<Route path="marketplace/agents/:id" element={<AgentMarketplaceDetailPage />} />
```

With:
```tsx
<Route path="marketplace" element={<MarketplacePage />} />
<Route path="marketplace/agents/:id" element={<AgentMarketplaceDetailPage />} />
<Route path="marketplace/extensions/:id" element={<ExtensionDetailPage />} />
```

Add redirects:
```tsx
<Route path="marketplace/agents" element={<Navigate to="/marketplace?type=agents" replace />} />
```

Remove `FunctionMarketplacePage` route at `/functions/discovery` (if registered in App.tsx).

### 4.2 Update ROUTES constants

**File:** `web/dashboard/src/lib/constants.ts`

```typescript
MARKETPLACE: '/marketplace',
MARKETPLACE_AGENTS: '/marketplace?type=agents',
MARKETPLACE_EXTENSIONS: '/marketplace?type=extensions',
MARKETPLACE_FUNCTIONS: '/marketplace?type=functions',
```

Remove `MARKETPLACE_FUNCTIONS: '/functions/discovery'`.

### 4.3 Replace MarketplaceDropdown with simple link

**File:** `web/dashboard/src/components/common/MarketplaceDropdown.tsx`

Replace the dropdown component with a simple `<Link to="/marketplace">Marketplace</Link>`. Keep the same component name for import compatibility.

### 4.4 Update sidebar navigation

**File:** `web/dashboard/src/components/layout/Sidebar/navigation.tsx`

Move marketplace entry from "Advanced" section to primary section (near top). Update icon to `Store`.

---

## Phase 5: Cleanup

### 5.1 Remove old pages

- Delete `web/dashboard/src/pages/AgentsMarketplacePage.tsx`
- Delete `web/dashboard/src/pages/AgentMarketplacePage.tsx`
- Delete `web/dashboard/src/pages/FunctionMarketplacePage.tsx`

### 5.2 Remove old view component

- Delete `web/dashboard/src/components/registry/AgentMarketplaceView.tsx` (logic moved to `AgentCard`)

### 5.3 Wire extension marketplace to real API

- `MarketplacePage` calls `marketplaceApi.list()` for extensions (already exists)
- Remove mock data dependency from `MarketplaceHome.tsx` (or delete if no longer used)

### 5.4 Update internal links

- Search codebase for `/marketplace/agents`, `/functions/discovery`, `MARKETPLACE_FUNCTIONS` and update all references

---

## Verification

1. `go build -o bin/orchestrator-api ./cmd/orchestrator-api` — backend compiles
2. `go test ./internal/api/handlers/marketplace/...` — handler tests pass
3. `cd web/dashboard && npx vite build` — frontend builds
4. Manual: navigate to `/marketplace`, verify unified grid loads
5. Manual: type filter chips change results
6. Manual: `/marketplace/agents/:id` still works
7. Manual: old `/marketplace/agents` redirects to `/marketplace?type=agents`

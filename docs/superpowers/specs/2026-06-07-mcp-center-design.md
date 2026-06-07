# MCP Center — Unified Dashboard Design

## Context

The Model Context Protocol (MCP) enables FunctionFly functions to be callable from AI agents like Claude Desktop, Cursor, and VS Code. Currently, MCP settings are accessible only via individual function settings pages (`MCPSettingsPanel`). We need a centralized MCP dashboard that provides registry management, connection monitoring, analytics, and global settings.

## Overview

A comprehensive dashboard page at `/mcp` for managing MCP integration across all functions. Uses a tabbed interface with 4 main sections: Registry, Connections, Analytics, Settings.

---

## Page Structure

```
/mcp → MCPCenterPage
├── Registry Tab (default)
├── Connections Tab
├── Analytics Tab
└── Settings Tab
```

---

## Tab 1: Registry

**Purpose:** View and manage all functions with MCP settings.

### Components

**Summary Cards (top row):**
- Total MCP-enabled functions (count)
- Verified MCP functions (count + badge)
- Total MCP invocations (last 30 days)
- Active transports (streamable-http / stdio count)

**Filter Bar:**
- Search input (function name/author)
- Filter pills: All | Enabled | Disabled | Verified
- Sort dropdown: Name | Invocations | Last Invoked

**Function Table:**
| Column | Description |
|--------|-------------|
| Function | Author/name with link to function detail |
| Tool Name | MCP tool name (e.g., `author__funcname`) |
| Status | Enabled/Disabled badge |
| Invocations | Call count (last 30d) |
| Last Invoked | Relative time or "Never" |
| Verified | Shield badge if verified |
| Actions | Edit settings, Toggle enable |

**Bulk Actions:**
- Enable selected functions
- Disable selected functions

**Empty State:**
"No MCP-enabled functions. Browse the function marketplace to discover functions that support MCP."

---

## Tab 2: Connections

**Purpose:** Monitor which AI clients are connecting to your functions.

### Components

**Client Cards Grid (2-3 columns):**
Each card represents an AI client type:
- Claude Desktop
- Cursor
- VS Code
- Windsurf
- Other MCP Clients

**Client Card Content:**
- Client icon/avatar
- Connection status indicator (green/yellow/gray)
- Number of connected functions
- Last connected timestamp
- "View Details" button

**Connection Details Panel (on card click):**
- Client name and description
- Connected functions list with individual usage
- Usage statistics for this client
- Connection instructions snippet
- "Revoke Access" action (future)

**Status Definitions:**
- Active (green): Connected within last 24h
- Stale (yellow): No activity in 24h-7d
- Never (gray): No recorded connections

---

## Tab 3: Analytics

**Purpose:** MCP-specific metrics and insights.

### Components

**KPI Cards (top row):**
| Metric | Description |
|--------|-------------|
| Total MCP Calls | Invocations in selected period |
| Unique Clients | Distinct AI clients calling |
| Avg Latency | Mean response time (ms) |
| Success Rate | % of calls that succeeded |

**Time Range Selector:**
- Pills: 24h | 7d | 30d | 90d

**Charts Section (2x2 grid):**
1. **Calls Over Time** — Area chart, calls per day
2. **Client Breakdown** — Donut chart, calls by client type
3. **Top Functions** — Horizontal bar chart, most invoked via MCP
4. **Transport Usage** — Bar chart, streamable-http vs stdio split

**Insights Panel (below charts):**
- Auto-generated recommendations:
  - "Enable verification for function X to increase trust score"
  - "Your function Y is trending with Claude Desktop users"
  - "3 functions are never invoked via MCP — consider enabling them"

---

## Tab 4: Settings

**Purpose:** Global MCP configuration defaults.

### Components

**Default Settings Card:**
- Default transport selector (streamable-http / stdio / both)
- Default rate limit input (calls/minute)
- Default schema exposure toggles (input / output)

**Registry Settings Card:**
- Auto-add new functions to MCP registry (toggle)
- Require verification for new functions (toggle)
- Public MCP registry listing (toggle)

**Security Settings Card:**
- Global CORS allowlist (textarea, one origin per line)
- IP allowlist (future)
- Rate limit multiplier (1x, 2x, 5x, 10x)

**Documentation Links:**
- MCP Documentation
- Integration Guides
- API Reference

---

## Technical Approach

### File Structure

```
pages/MCPCenterPage/
├── index.tsx              # Main page component with tabs
├── types.ts               # TypeScript interfaces
├── constants.ts          # Date ranges, defaults
├── hooks/
│   ├── useMCPFunctions.ts # Fetch functions with MCP data
│   ├── useMCPAnalytics.ts # Fetch MCP analytics
│   └── useMCPSettings.ts  # Global MCP settings
└── components/
    ├── RegistryTab.tsx   # Registry tab content
    ├── ConnectionsTab.tsx # Connections tab content
    ├── AnalyticsTab.tsx  # Analytics tab content
    ├── SettingsTab.tsx   # Settings tab content
    ├── SummaryCards.tsx  # KPI summary row
    ├── FunctionTable.tsx # MCP function table
    ├── ClientGrid.tsx    # AI client cards grid
    └── ChartsGrid.tsx    # Analytics charts
```

### API Design

| Method | Endpoint | Description |
|--------|----------|-------------|
| GET | `/v1/functions?mcp=true` | List functions with MCP settings |
| GET | `/v1/functions/:id/mcp` | Get MCP settings for function |
| PATCH | `/v1/functions/:id/mcp` | Update function MCP settings |
| GET | `/v1/mcp/analytics` | Get MCP analytics (calls, clients) |
| GET | `/v1/mcp/connections` | Get active MCP connections by client |
| GET | `/v1/mcp/settings` | Get global MCP settings |
| PATCH | `/v1/mcp/settings` | Update global MCP settings |

### Data Models

```typescript
interface MCPSettings {
  function_id: string;
  enabled: boolean;
  transports: ('streamable-http' | 'stdio')[];
  expose_input_schema: boolean;
  expose_output_schema: boolean;
  tool_name_override: string;
  rate_limit_per_min: number;
  allowlist_origins: string[];
  verified_mcp?: boolean;
  invocation_count?: number;
  last_invoked_at?: string | null;
}

interface MCPFunction extends MCPSettings {
  author: string;
  name: string;
  function_id: string;
}

interface MCPAnalytics {
  total_calls: number;
  unique_clients: number;
  avg_latency_ms: number;
  success_rate: number;
  calls_over_time: { time: string; count: number }[];
  client_breakdown: { client: string; count: number }[];
  top_functions: { author: string; name: string; calls: number }[];
  transport_usage: { transport: string; count: number }[];
}

interface MCPConnection {
  client_type: string;
  status: 'active' | 'stale' | 'never';
  connected_functions: number;
  last_connected_at: string | null;
}

interface MCPSettingsGlobal {
  default_transport: 'streamable-http' | 'stdio' | 'both';
  default_rate_limit: number;
  default_expose_input: boolean;
  default_expose_output: boolean;
  auto_add_to_registry: boolean;
  require_verification: boolean;
  public_listing: boolean;
  cors_allowlist: string[];
  rate_limit_multiplier: number;
}
```

### State Management

- Use TanStack Query for all data fetching (consistent with existing patterns)
- Local component state for UI (active tab, filters, selections)
- `useMCPSettings` hook for global settings CRUD

---

## Component Inventory

| Component | States | Notes |
|-----------|--------|-------|
| `MCPMetricsCard` | default, loading, empty | KPI display with icon and trend |
| `MCPSummaryCards` | default, loading | Row of 4 metric cards |
| `MCPFunctionTable` | default, loading, empty, filtered | Sortable, selectable table |
| `MCPFunctionRow` | default, hovered, selected, disabled | Table row with inline actions |
| `MCPClientCard` | active, stale, never-connected | Connection card with status |
| `MCPClientGrid` | default, loading, empty | 2-3 column grid of client cards |
| `MCPCallsChart` | loading, empty, populated | Area chart for calls over time |
| `MCPClientPieChart` | loading, empty, populated | Donut chart for client breakdown |
| `MCPTopFunctionsChart` | loading, empty, populated | Horizontal bar chart |
| `MCPSettingsForm` | default, saving, error, success | Form with validation |
| `MCPTabNav` | default | Tab navigation component |

---

## Routing

Add to `App.tsx`:
```tsx
<Route path="mcp" element={<MCPCenterPage />} />
```

---

## Dependencies

- Reuse existing UI components: Card, Button, Badge, Table, Tabs, Switch, Input, Select, Charts
- Reuse existing hooks: `useFunctions` (extend with MCP data)
- New hooks: `useMCPAnalytics`, `useMCPSettings`
- Existing `MCPSettingsPanel` can be reused in a modal for inline editing

---

## Priority

1. Registry tab (core functionality)
2. Analytics tab (key metrics)
3. Connections tab (monitoring)
4. Settings tab (global config)
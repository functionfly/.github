# Execution Root Hash Explorer - Technical Plan

## Overview
Create an explorer interface on function profile pages that allows users to browse and inspect all Execution Root Hashes (MEG records) for a function. These hashes represent unique fingerprints of each deterministic function execution.

## Background
The DRE (Deterministic Replay Environment) 2.0 system generates a unique `ExecutionRootHash` for each function execution via the MEG (Merkle Execution Graph). This hash is the Merkle root of 7 component hashes:
1. Input Hash
2. Environment Hash
3. Dependency Hash
4. Trace Hash
5. Resource Hash
6. Output Hash
7. Metadata Hash

## Architecture

```mermaid
flowchart TB
    subgraph Frontend["Frontend"]
        FP[Function Profile Page]
        EX[Execution Explorer Page\n/registry/{author}/{name}/executions]
        ED[Execution Detail Modal]
    end

    subgraph API["Backend API"]
        R1[/GET /registry/{author}/{name}/executions\n- List MEG records with pagination
- Filter by date range, version
- Sort by created_at/execution_id/]
        R2[/GET /registry/{author}/{name}/executions/{execution_id}\n- Get single MEG record details
- Include component hashes/]
    end

    subgraph Storage["Database Layer"]
        MEG[(execution_meg_records)]
        CERT[(execution_certificates)]
    end

    FP -->|click hash link| EX
    EX -->|fetch| R1
    R1 --> MEG
    EX -->|click execution| ED
    ED -->|fetch| R2
    R2 --> MEG
```

## API Design

### 1. List Executions Endpoint
**Route:** `GET /v1/registry/{author}/{name}/executions`

**Query Parameters:**
| Parameter | Type | Default | Description |
|-----------|------|---------|-------------|
| limit | int | 20 | Page size (max 100) |
| offset | int | 0 | Pagination offset |
| version | string | - | Filter by version |
| from | datetime | - | Filter from date |
| to | datetime | - | Filter to date |
| verified_only | bool | false | Only show replay-verified |

**Response:**
```json
{
  "function": "fx://author/name",
  "executions": [
    {
      "execution_id": "uuid",
      "execution_root_hash": "41bf6529f0100d4201358564e...",
      "version": "1.2.0",
      "created_at": "2024-01-15T10:30:00Z",
      "replay_verified": true,
      "roots_match": true,
      "component_hashes": {
        "input": "abc123...",
        "output": "def456...",
        "environment": "ghi789...",
        "dependency": "jkl012...",
        "resource": "mno345...",
        "metadata": "pqr678..."
      }
    }
  ],
  "total": 1543,
  "limit": 20,
  "offset": 0
}
```

### 2. Get Single Execution Endpoint
**Route:** `GET /v1/registry/{author}/{name}/executions/{execution_id}`

**Response:**
```json
{
  "execution": {
    "execution_id": "uuid",
    "execution_root_hash": "41bf6529f0100d4201358564e...",
    "version": "1.2.0",
    "created_at": "2024-01-15T10:30:00Z",
    "determinism_tier": "full",
    "protocol_version": "dre/1.0",
    "replay_verified_at": "2024-01-15T10:31:00Z",
    "replay_root_hash": "41bf6529f0100d4201358564e...",
    "replay_node_id": "node-xyz",
    "roots_match": true,
    "component_hashes": {
      "input": "abc123...",
      "output": "def456...",
      "environment": "ghi789...",
      "dependency": "jkl012...",
      "trace": "stu901...",
      "resource": "mno345...",
      "metadata": "pqr678..."
    },
    "certificate": {
      "certificate_id": "fxc_01H...",
      "cert_level": "standard",
      "certificate_hash": "sha256:..."
    }
  }
}
```

## Data Model

### Repository Method Needed
```go
// GetMEGRecordsByFunctionID retrieves paginated MEG records for a function
func (r *RegistryRepository) GetMEGRecordsByFunctionID(
    functionID uuid.UUID, 
    limit, offset int,
    filters MEGRecordFilters,
) ([]*MEGRecord, int64, error)

type MEGRecordFilters struct {
    Version        string
    From           *time.Time
    To             *time.Time
    VerifiedOnly   bool
}
```

## Frontend Design

### Component Structure
```
web/dashboard/src/
├── components/
│   └── registry/
│       ├── ExecutionExplorer.tsx       # Main explorer component
│       ├── ExecutionHashCard.tsx       # Individual hash card
│       ├── ExecutionDetailModal.tsx    # Drill-down modal
│       └── ExecutionHashBadge.tsx      # Reusable hash badge
├── api/
│   └── dre.ts                          # DRE API client methods
└── pages/
    ├── FunctionPage/
    │   └── index.tsx                   # Add link to explorer
    └── ExecutionExplorerPage/
        └── index.tsx                   # Standalone explorer page
```

### Route Configuration
Add new route in [`App.tsx`](web/dashboard/src/App.tsx:1):
```tsx
<Route 
  path="registry/:author/:name/executions" 
  element={<ExecutionExplorerPage />} 
/>
```

### Execution Explorer Component

**Features:**
1. **Hash Grid/List View** - Display execution hashes in a grid or list
2. **Pagination** - Standard offset-based pagination
3. **Filters:**
   - Version selector (dropdown of available versions)
   - Date range picker
   - Verified only toggle
4. **Sorting:**
   - Newest first (default)
   - Oldest first
   - By execution ID

**Visual Design:**
- Use existing card component style from `function-profile-card`
- Hash displayed with monospace font, truncated with tooltip for full value
- Status badges for:
  - `replay_verified` - Green checkmark
  - `roots_match` - Shield icon
  - `cert_level` - Badge (lite/standard/legal_grade)
- Component hash mini-preview (first 8 chars of each)

### Function Profile Integration

The Execution Root Hash displayed on the Function Profile page should be a clickable link that navigates to the standalone Execution Explorer page:

```tsx
// In Function Profile page
<div>
  <p className="text-xs font-medium text-muted-foreground uppercase tracking-wide mb-1">
    Execution root hash
  </p>
  {functionInfo.source_hash ? (
    <Link 
      to={`/registry/${functionInfo.author}/${functionInfo.name}/executions`}
      className="inline-flex items-center gap-1"
    >
      <Badge variant="outline" className="font-mono text-xs gap-1 hover:bg-brand-500/10 transition-colors cursor-pointer">
        <Lock className="w-3 h-3 shrink-0" />
        <span className="truncate">{functionInfo.source_hash}</span>
        <ExternalLink className="w-3 h-3 shrink-0 ml-1" />
      </Badge>
    </Link>
  ) : (
    <span className="text-sm text-muted-foreground">Not available</span>
  )}
</div>
```

## Implementation Steps

### Phase 1: Backend API
1. Add `GetMEGRecordsByFunctionID` repository method
2. Create `HandleListExecutions` handler in `internal/api/handlers/registry/dre/handlers.go`
3. Create `HandleGetExecution` handler for single execution details
4. Register routes in `internal/api/routes.go`:
   - `GET /v1/registry/{author}/{name}/executions`
   - `GET /v1/registry/{author}/{name}/executions/{execution_id}`

### Phase 2: Frontend Components
1. Create `web/dashboard/src/api/dre.ts` with:
   - `listExecutions(author, name, filters)`
   - `getExecution(author, name, executionId)`
2. Create `ExecutionExplorer.tsx` component
3. Create `ExecutionHashCard.tsx` component
4. Create `ExecutionDetailModal.tsx` component

### Phase 3: Integration
1. Create `ExecutionExplorerPage.tsx` at `web/dashboard/src/pages/ExecutionExplorerPage/index.tsx`
2. Add route in `App.tsx`: `/registry/:author/:name/executions`
3. Update Function Profile page to link hash to explorer page
4. Add click handler on execution cards to open ExecutionDetailModal

## UI Mockup Description

### Execution Explorer View
```
┌─────────────────────────────────────────────────────────────────┐
│  Execution History                                    [?] Help   │
├─────────────────────────────────────────────────────────────────┤
│  [Version: All ▼]  [From: ____] [To: ____]  [✓ Verified Only]  │
├─────────────────────────────────────────────────────────────────┤
│  ┌──────────────────────────────────────────────────────────┐  │
│  │ 🔒 41bf6529f0100d4201358564e...          ✓ Verified ✓     │  │
│  │                                                           │  │
│  │ Version: 1.2.0           Created: Jan 15, 2024 10:30 AM   │  │
│  │                                                           │  │
│  │ Components:  Input: abc1...  Output: def4...              │  │
│  │              Env: ghi7...    Dep: jkl0...                 │  │
│  └──────────────────────────────────────────────────────────┘  │
│  ┌──────────────────────────────────────────────────────────┐  │
│  │ 🔒 a3c8d912f5b...                         ○ Pending      │  │
│  │                                                           │  │
│  │ Version: 1.2.0           Created: Jan 15, 2024 10:25 AM   │  │
│  │                                                           │  │
│  │ Components:  Input: xyz9...  Output: uvw2...              │  │
│  │              Env: rst5...    Dep: qrs8...                 │  │
│  └──────────────────────────────────────────────────────────┘  │
├─────────────────────────────────────────────────────────────────┤
│  Showing 1-20 of 1,543 executions    [< Prev] [1] [2] ... [Next >]│
└─────────────────────────────────────────────────────────────────┘
```

### Execution Detail Modal
```
┌─────────────────────────────────────────────────────────────────┐
│  Execution Details                                    [X] Close │
├─────────────────────────────────────────────────────────────────┤
│  Execution Root Hash                                            │
│  ┌──────────────────────────────────────────────────────────┐  │
│  │ 41bf6529f0100d4201358564e8f9a2b3c4d5e6f7a8b9c0d1e2f3...  │  │
│  │                                                          │  │
│  │ [📋 Copy]  [🔗 View Certificate]  [🔄 Verify Replay]     │  │
│  └──────────────────────────────────────────────────────────┘  │
├─────────────────────────────────────────────────────────────────┤
│  Status           │ Version: 1.2.0                              │
│  ✓ Verified       │ Determinism: Full                           │
│  ✓ Roots Match    │ Protocol: DRE/1.0                           │
├─────────────────────────────────────────────────────────────────┤
│  Component Hashes (MEG)                                         │
│  ┌─────────────┬─────────────────────────────────────────────┐  │
│  │ Input       │ abc123def456...                             │  │
│  │ Output      │ def456abc789...                             │  │
│  │ Environment │ ghi789def012...                             │  │
│  │ Dependency  │ jkl012ghi345...                             │  │
│  │ Trace       │ stu901vwx234...                             │  │
│  │ Resource    │ mno345pqr678...                             │  │
│  │ Metadata    │ pqr678stu901...                             │  │
│  └─────────────┴─────────────────────────────────────────────┘  │
├─────────────────────────────────────────────────────────────────┤
│  Replay Verification                                            │
│  Verified At: Jan 15, 2024 10:31:00 UTC                         │
│  Replay Node: node-xyz-abc                                      │
│  Replay Root Hash: 41bf6529f0100d4201358564e... (matches)       │
└─────────────────────────────────────────────────────────────────┘
```

## Security Considerations

1. **Public Read Access** - Execution records should be publicly readable (like certificates)
2. **No Sensitive Data** - Never expose actual input/output data, only hashes
3. **Rate Limiting** - Apply standard public API rate limits
4. **No Write Operations** - Explorer is read-only

## Performance Considerations

1. **Database Indexing** - Ensure `function_id` and `created_at` are indexed on `execution_meg_records`
2. **Pagination** - Always use limit/offset, never return all records
3. **Caching** - Consider Redis caching for popular functions' execution lists
4. **CDN** - Static hash data could be CDN-cached with short TTL

## Testing Strategy

1. **Unit Tests:**
   - Repository method tests with mock DB
   - Handler tests with mock repository
   - Component tests with mock API

2. **Integration Tests:**
   - End-to-end flow: create execution → query via API → verify response

3. **E2E Tests:**
   - Browser tests for explorer UI interactions

## Success Metrics

1. **Performance:**
   - API response time < 200ms for list queries
   - Page load time < 1s for explorer

2. **User Experience:**
   - Users can find specific executions within 3 clicks
   - Hash comparison is visually clear

3. **Adoption:**
   - Execution History tab viewed on > 50% of function profile visits

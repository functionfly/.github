# Minimal SDK Integration Spec (LangChain / AutoGen / CrewAI)

## Goal
Provide a small, framework-agnostic integration surface so AI-agent SDKs can:
1. Discover tools/functions that are **verified** and **trust-scored**
2. Export each tool as a **model tool schema**
3. Enforce a **trust policy** when selecting candidate tools
4. Execute selected tools while preserving auditability (inputs/outputs + trust metadata)

This spec is designed to align with FunctionFly's existing registry, execution, and verification endpoints.

## Core Concepts (Trust Primitives)

### TrustPolicy (input to the SDK integration)
An agent SDK should pass a policy object to the selector/discovery layer.

Example (JSON):
```json
{
  "minTrustScore": 80,
  "requiredTrustLevels": ["high", "verified"],
  "requireVerified": true,
  "capabilitiesAllow": ["http_get", "http_post"],
  "capabilitiesDeny": ["secrets_read"],
  "maxEgressDomains": ["example.com"]
}
```

Field meanings:
- `minTrustScore`: maps to FunctionFly search filtering (e.g. `min_rating`) and UI-side trust thresholds.
- `requiredTrustLevels`: optional; maps to `trust_level` returned by FunctionFly.
- `requireVerified`: gates on `verified` returned by FunctionFly function profile.
- `capabilitiesAllow` / `capabilitiesDeny`: SDK-side filtering for tool manifests/capabilities.
- `maxEgressDomains`: example of how a policy can constrain network behavior (manifest-driven).

### TrustedFunction (output of discovery)
The SDK should treat each candidate tool as:
- a model tool schema (for tool calling)
- trust metadata (trust_score, trust_level, verified, etc.)

Discovery responses already include trust fields on FunctionFly's public registry endpoints:
- `trust_score` (0-100 scale)
- `trust_level` (e.g. `"high"`)
- `verified` (boolean)

## Discovery (Where the SDK finds tools)

### Search by query + trust threshold
- `GET /v1/registry/search`
  - Query params: `q`, optional `category`, optional `min_rating` (maps to trust filtering)

Expected response:
- list of functions with trust metadata (`trust_score`, `trust_level`, `verified`)

### Fetch function profile (for tool schema + policy checks)
- `GET /v1/registry/functions/{author}/{name}`
  - Expands trust metadata on the returned function profile
  - Optional query: `expand=manifest`

## Schema Export (Tool Calling Integration)

### Export an AI tool schema for a function
- `GET /fx/{author}/{name}/ai-schema`

Integration expectations:
- The response should be treated as the canonical tool definition for the agent framework.
- The SDK adapter should transform it into that framework's tool format:
  - LangChain: tool definition (name, description, input schema)
  - AutoGen: function/tool registration wrapper
  - CrewAI: tool/agent action schema wrapper

## Trust Policy Filtering (Enforcement)

Policy enforcement should occur before tool registration/exposure.

Recommended algorithm for a selector:
1. Discover candidate tools using `/v1/registry/search` (optionally with `min_rating`)
2. Fetch function profiles for top candidates using `/v1/registry/functions/{author}/{name}?expand=manifest`
3. Enforce:
   - `verified === true` when `requireVerified`
   - `trust_score >= minTrustScore` when provided
   - `trust_level` membership when `requiredTrustLevels`
4. Enforce capability constraints based on the returned manifest/capabilities

The key design constraint:
- The SDK must never “silently allow” tools outside the trust policy.

## Execution (Running trusted tools)

### Execute a function by author/name (latest)
- `POST /v1/fx/{author}/{name}`

### Execute a function with an explicit version
- `POST /v1/fx/{author}/{name}@{version}`

Integration expectations:
- The SDK should attach enough metadata to map execution results back to the trust decision:
  - tool id
  - function version
  - trust policy used (or policy hash)
  - returned trust metadata (if available via execution result)

## Verification Workflow (Optional admin/creator flows)
SDK integrations may also support creator/admin workflows.

Key endpoints:
- `GET /registry/verification/{functionVersionId}/status`
- `POST /registry/verification/{functionVersionId}/sign`
- `POST /registry/verification/{functionVersionId}/approval`
- `POST /registry/verification/approvals/{approvalId}/decide`
- `GET /registry/verification/approvals`
- `GET /registry/verification/approvals/pending`

These are primarily for platform tooling and do not need to be part of the runtime trust policy selection path.

## Reference Adapter Interfaces (TypeScript-ish)

The SDK adapter should be split into two parts:
1. Discovery + Trust Filter
2. Schema Export + Tool Registration

Minimal interface:
```ts
type TrustPolicy = {
  minTrustScore?: number;
  requiredTrustLevels?: string[];
  requireVerified?: boolean;
  capabilitiesAllow?: string[];
  capabilitiesDeny?: string[];
  maxEgressDomains?: string[];
};

type TrustedFunction = {
  author: string;
  name: string;
  version?: string;
  trust_score: number;
  trust_level: string;
  verified: boolean;
  toolSchema: unknown; // framework-adapter specific shape
};

interface TrustedToolDiscovery {
  searchAndFilter: (policy: TrustPolicy, query: string, category?: string) => Promise<TrustedFunction[]>;
}

interface ToolSchemaExporter {
  exportToolSchema: (author: string, name: string) => Promise<unknown>;
}
```

## Framework Mapping (Implementation Guidance)

### LangChain
- Use `TrustPolicy` in your tool-selector step.
- Register only tools that pass the trust filter.
- Convert `ai-schema` output into a LangChain Tool or StructuredTool.

### AutoGen
- Register tool actions only after trust-filtering.
- Ensure tool-calling is gated by the trust policy used during selection.

### CrewAI
- Treat FunctionFly tools as first-class “actions.”
- Provide the framework action with:
  - schema from `ai-schema`
  - trust metadata for observability/debugging

## Notes / Non-Goals
- This spec does not define a full execution sandbox or security policy grammar; it only defines the minimal interface between agent SDKs and FunctionFly's trust-aware discovery + schema export.
- Revocation handling is expected to be enforced by the selector when fetching/updating trust metadata.


# MiMo & MiniMax Token Plan Support

**Date:** 2026-06-30  
**Status:** Approved  
**Scope:** AI Keys settings page (`/u/{username}/settings#ai-keys`)

## Problem

Users who subscribe to MiMo Token Plans (credit-based subscriptions with `tp-` prefixed API keys and regional endpoints) cannot use them on FunctionFly. The current MiMo BYOK support only handles standard pay-as-you-go API keys against `api.xiaomimimo.com`.

## Background

MiMo Token Plans are prepaid credit bundles (Lite $6/mo through Max $100/mo) purchased on MiMo's platform. Key differences from standard MiMo API keys:

| Aspect | Standard API Key | Token Plan Key |
|--------|-----------------|----------------|
| Key prefix | Any | `tp-xxxxx` |
| Base URL | `api.xiaomimimo.com/v1` | Regional (see below) |
| Billing | Pay-as-you-go per token | Prepaid credit quota |
| Supported models | All MiMo models | mimo-v2.5 series only |

Regional Token Plan endpoints:
- China: `https://token-plan-cn.xiaomimimo.com/v1`
- Singapore: `https://token-plan-sgp.xiaomimimo.com/v1`
- Europe: `https://token-plan-ams.xiaomimimo.com/v1`

No public API exists to query plan tier or remaining credits programmatically.

## Design Decisions

1. **Separate provider** (`mimo-token-plan`) rather than merging with existing `mimo`. Rationale: different endpoints, different auth, different model availability. Users can connect both simultaneously.

2. **User selects region** (CN/SGP/EU) during connect. No auto-detection — MiMo doesn't expose a region-detection API, and probing 3 endpoints is wasteful.

3. **Format-only validation** — validate `tp-` prefix + minimum length. Skip live API test during connect to avoid rate limit friction. Health worker handles runtime validation.

4. **Region stored in `health_message`** field as `region:<code>` for MVP. Avoids a DB migration. If richer metadata is needed later, add a proper column.

5. **AI proxy passes regional base URL** via `X-BYOK-Base-URL` header to the AI service, which overrides the MiMo client's base URL.

## Implementation Plan

### Backend (Go)

#### 1. `internal/aikeys/validate.go`

- Add to `SupportedProviders()`:
  ```go
  {ID: "mimo-token-plan", Name: "MiMo Token Plan", Description: "Prepaid credit plan (tp-...)", KeyFormat: "tp-...", KeyPrefix: "tp-"},
  ```
- Update `validateFormat()` for `"mimo-token-plan"`: require `tp-` prefix, min 10 chars
- Update `providerTestConfig()` for `"mimo-token-plan"`: return regional endpoint based on stored region (or skip test entirely — health worker handles it)

#### 2. `internal/aikeys/types.go`

- Add `Region` field to `ConnectRequest`:
  ```go
  type ConnectRequest struct {
    Provider string `json:"provider"`
    APIKey   string `json:"apiKey"`
    Region   string `json:"region,omitempty"` // For mimo-token-plan: "cn", "sgp", "eu"
  }
  ```
- Add `TokenPlanRegion` field to `KeyResponse`:
  ```go
  type KeyResponse struct {
    // ... existing fields ...
    TokenPlanRegion string `json:"token_plan_region,omitempty"`
  }
  ```

#### 3. `internal/aikeys/handler.go`

- `HandleConnectKey`: when provider is `mimo-token-plan`:
  - Validate `Region` is one of `cn`, `sgp`, `eu`
  - Skip live API test (format validation only)
  - Store region in `health_message` as `region:<code>`
  - KeyLast4 format: `tp-...{last4}` instead of `sk-...{last4}`
- `HandleListKeys` / `toKeyResponse`: extract region from `health_message` and populate `TokenPlanRegion`

#### 4. `internal/aikeys/validate.go` — `providerTestConfig()`

For `"mimo-token-plan"`, map region to endpoint:
- `cn` → `https://token-plan-cn.xiaomimimo.com/v1/models`
- `sgp` → `https://token-plan-sgp.xiaomimimo.com/v1/models`
- `eu` → `https://token-plan-ams.xiaomimimo.com/v1/models`

#### 5. `internal/api/ai_proxy.go` — `injectBYOKHeader()`

When provider is `mimo-token-plan`:
- Extract region from stored `health_message`
- Set `X-BYOK-Base-URL` header to the regional endpoint
- Set `X-BYOK-Provider` to `mimo` (so AI service uses MiMo provider)
- Set `X-Key-Source` to `byok`

#### 6. `internal/aikeys/health_worker.go`

Update health check logic to route `mimo-token-plan` keys to the correct regional endpoint (extract region from `health_message`).

### AI Service (Python)

#### 7. `ai-service/src/providers/mimo.py`

In `__init__` or in the proxy path, read `X-BYOK-Base-URL` header and use it to override `self.base_url`. This allows the Go proxy to control which regional endpoint the Python provider hits.

### Frontend (React/TypeScript)

#### 8. `web/dashboard/src/types/ai-keys.ts`

```ts
export interface AIProviderKey {
  // ... existing fields ...
  token_plan_region?: string;
}

export interface ConnectAIKeyRequest {
  provider: string;
  apiKey: string;
  region?: string;
}
```

#### 9. `web/dashboard/src/pages/SettingsPage/components/AIKeysSettingsTab/AIKeysSettingsTab.tsx`

- Add `mimo-token-plan: 'TP'` to `PROVIDER_LOGOS`
- In `ConnectKeyDialog`: when `mimo-token-plan` is selected, show region picker (3 buttons: CN, SGP, EU) before the API key input
- In `ConnectedKeyCard`: show region badge for token plan keys. Show key as `tp-...{last4}`
- Pass `region` in `onConnect` callback

#### 10. `web/dashboard/src/api/ai-keys.ts`

- `connectKey()`: include `region` in request body when present

## File Change Summary

| File | Change |
|------|--------|
| `internal/aikeys/validate.go` | Add `mimo-token-plan` provider, format validation, regional test config |
| `internal/aikeys/types.go` | Add `Region` to `ConnectRequest`, `TokenPlanRegion` to `KeyResponse` |
| `internal/aikeys/handler.go` | Store region, skip live test for token plan, extract region in responses |
| `internal/api/ai_proxy.go` | Set `X-BYOK-Base-URL` for token plan keys |
| `internal/aikeys/health_worker.go` | Route health checks to regional endpoint |
| `ai-service/src/providers/mimo.py` | Read `X-BYOK-Base-URL` to override base URL |
| `web/dashboard/src/types/ai-keys.ts` | Add `token_plan_region` field, `region` to request |
| `web/dashboard/src/pages/.../AIKeysSettingsTab.tsx` | Region picker, token plan UI, logo |
| `web/dashboard/src/api/ai-keys.ts` | Pass region in connect request |

## Out of Scope

- Plan tier detection (Lite/Standard/Pro/Max for MiMo, Plus/Max/Ultra for MiniMax) — no API available
- Credit balance display — no API available
- Automatic plan expiry handling — providers manage this
- DB migration — using `health_message` field for region storage (MVP)

## MiniMax Token Plan (Added)

MiniMax Token Plan was added alongside MiMo using the same pattern, but simpler (no region selection):

| Aspect | Value |
|--------|-------|
| Provider ID | `minimax-token-plan` |
| Key prefix | `sk-cp-...` |
| Base URL | `https://api.minimaxi.com/v1` (single endpoint, no regional variants) |
| Model | MiniMax-M3 |
| Tiers | Plus (¥49/mo), Max (¥119/mo), Ultra (¥469/mo) |

Additional files changed for MiniMax:
- `ai-service/src/providers/minimax.py` — New OpenAI-compatible provider
- `web/dashboard/src/components/common/ProviderIcon.tsx` — Added icon entries for both token plan providers

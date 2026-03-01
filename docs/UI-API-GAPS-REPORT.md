# UI/API Gaps Report — Agents & General Wiring

Exploration date: 2025-03-01. This document lists **gaps** (missing UI, missing routes, unwired backend, dead frontend) and **inconsistencies** so you can prioritize fixes.

---

## 1. Agent backend vs UI — major gaps

### 1.1 AEP (Agent Execution Plan) — backend exists, **no UI**

All of these are **registered** in `internal/api/routes.go` and work when called with JWT:

| Area | Endpoints | UI / API client |
|------|-----------|------------------|
| **Agent CRUD** | `POST /v1/agent/register`, `GET /v1/agent`, `GET/DELETE /v1/agent/{agent_id}` | ❌ No page, no `api/agent.ts` |
| **Quota** | `PUT /v1/agent/{agent_id}/quota`, `GET /v1/agent/{agent_id}/usage` | ❌ |
| **Policy** | `GET/PUT /v1/agent/{agent_id}/policy` | ❌ |
| **Executions** | `GET /v1/agent/{agent_id}/executions`, `GET .../executions/{exec_id}` | ❌ |
| **Analytics** | `GET /v1/agent/{agent_id}/analytics` | ❌ |
| **Sessions** | `POST .../session/start`, `POST .../session/{id}/end`, `GET .../session/{id}` | ❌ |
| **Billing** | `GET .../billing/summary`, `PUT .../billing/spend-cap`, `GET .../billing/usage`, `GET .../cost-breakdown` | ❌ |
| **Credits** | `GET .../credits/balance`, `POST .../credits/purchase` | ❌ |
| **Concurrency** | `GET /v1/agent/concurrency/stats` | ❌ |
| **Discovery / Execute** | `GET /v1/agent/discover`, `GET .../discover/{author}/{name}`, `POST .../execute/{author}/{name}` | ❌ No dashboard usage |

- **Sidebar/nav:** No "Agents" or "Agent" entry in `Sidebar.tsx` or `ROUTES` in `lib/constants.ts`.
- **Dashboard wording:** Main dashboard shows "Agent activity" and "Create agent" but data comes from **dashboard API** (`/v1/dashboard/activity`, etc.) and **functions**, not from `/v1/agent/*`.

**Conclusion:** Users cannot manage AEP agents from the dashboard at all.

---

### 1.2 Swarm / marketplace / evolution — backend **not registered**, UI **unreachable**

- **Backend:** `internal/api/handlers/agent/swarm.go` defines `SwarmHandler` and `RegisterRoutes()` which would register:
  - `POST /v1/agent/{id}/spawn`, `GET .../children`, `GET .../parent`
  - `POST .../message`, `GET .../inbox`, `GET .../wallet`
  - `POST .../evolve`, `POST .../schedule`, `GET .../schedules`
  - `GET /v1/marketplace/agents`, `POST /v1/marketplace/agent/list`
- **Wiring:** `NewSwarmHandler` and `RegisterRoutes` are **never called** in `internal/api/routes.go`, so these endpoints **return 404**.

- **Frontend:** Components exist and are exported from `web/dashboard/src/components/swarm/`:
  - `SwarmDashboard`, `AgentMarketplace`, `FunctionMarketplace`, `EvolutionDashboard`, `WalletDashboard`
- **Routing:** None of these are mounted in `App.tsx`. There are **no routes** for:
  - `/agents`, `/agent/:id`, `/marketplace/agents`, `/marketplace/functions`, `/evolution`, etc.
- **Links:** `SwarmDashboard` links to `/marketplace/agents`, `/marketplace/functions`, `/evolution` — those paths hit the catch-all `*` and show **NotFoundPage**.
- **Data:** All swarm components use **local/mock state only**; no API client calls.

**Conclusion:** Swarm/marketplace/evolution backend is implemented but not exposed; swarm UI exists but is dead (no route, no API).

---

## 2. Other UI/API inconsistencies

| Issue | Location | Notes |
|-------|----------|--------|
| **Duplicate billing route** | `routes.go` | Both `GET .../billing/summary` and `GET .../billing/usage` map to `HandleGetBillingSummary`. Redundant. |
| **Execution Explorer vs agent executions** | Dashboard | Execution Explorer uses **registry DRE** (`/v1/registry/{author}/{name}/executions`). Agent execution history is under `/v1/agent/{agent_id}/executions` and has **no UI**. |
| **API index** | `web/dashboard/src/api/index.ts` | No `agentApi` or any agent-related export; only dashboard, auth, apps, functions, etc. |
| **Reserved usernames** | `internal/api/middleware/reserved_usernames.go` | `marketplace` and `market` are reserved; `agent` is not (only affects `/u/...` usernames). |

---

## 3. Summary table

| Area | Backend | Frontend | Wired? |
|------|---------|----------|--------|
| AEP (register, list, get, delete, quota, policy, executions, analytics, sessions, billing, credits, concurrency) | ✅ Routes in `routes.go` | ❌ No page, no API client | **No** |
| Agent discover/execute | ✅ Routes in `routes.go` | ❌ No UI calling these | **No** |
| Swarm (spawn, children, parent, message, inbox, wallet, evolve, schedule) | ✅ In `SwarmHandler`, **not registered** | Components exist; no route; mock only | **No** |
| Marketplace (search agents, create listing) | Same `SwarmHandler`, **not registered** | `AgentMarketplace` / `FunctionMarketplace`; no route; mock only | **No** |
| Dashboard “agent” activity | Uses `/v1/dashboard/activity` (functions) | Uses dashboard API only | Yes (but **not** agent API) |

---

## 4. Recommendations

1. **Wire SwarmHandler (if you want swarm/marketplace/evolution):**  
   In `routes.go`, construct `SwarmHandler` (with required deps from `internal/agent/*`) and call `RegisterRoutes(api)` so `/v1/agent/{id}/...` and `/v1/marketplace/...` are live.

2. **Add agent UI and API client:**  
   - Add `web/dashboard/src/api/agent.ts` calling `GET/POST/PUT/DELETE /v1/agent/*` (and optionally swarm/marketplace endpoints once registered).  
   - Add routes in `App.tsx` (e.g. `/dashboard/agents`, `/dashboard/agents/:id`, optionally `/marketplace/agents`, `/evolution`).  
   - Add sidebar entry and `ROUTES.AGENTS` (or similar) in `lib/constants.ts`.

3. **Either wire swarm UI or remove it:**  
   - **Option A:** Mount swarm components on the new routes and replace mock data with the new agent + swarm API client.  
   - **Option B:** Remove or hide the dead swarm components and internal links until the backend is registered and product-ready.

4. **Clean up backend:**  
   - Resolve duplicate `billing/usage` vs `billing/summary` (e.g. single handler or clear semantics).  
   - Consider adding `agent` to reserved usernames if you want to reserve `/u/agent`.

---

## 5. Small fix already applied

- **Typo:** In `AgentMarketplace.tsx`, mock listing had `agentId: ' swarm-manager'` (leading space). Fixed to `agentId: 'swarm-manager'`.

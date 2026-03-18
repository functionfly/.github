# Paperclip control-plane deployment (FunctionFly internal)

Paperclip runs as FunctionFly’s internal agent control plane: goals, issues, approvals, heartbeats, and budgets. FunctionFly remains the execution plane. See the plan in `.cursor/plans/` and [docs/PAPERCLIP_INTEGRATION.md](../../docs/PAPERCLIP_INTEGRATION.md).

## Deployment mode: authenticated + private

- **Mode**: `authenticated` (login required).
- **Exposure**: `private` (Tailscale/VPN/LAN only; do not expose to the public internet until you harden further).

## Option A: Docker Compose (from this repo)

Builds Paperclip from source and runs a dedicated Postgres container (so embedded Postgres is not used in Docker). Can take several minutes. If the build fails (e.g. at server build), use **Option B** instead.

1. Copy env and set required secrets:

   ```bash
   cp deploy/paperclip/.env.example deploy/paperclip/.env
   # Edit .env: set BETTER_AUTH_SECRET (and optional PAPERCLIP_PUBLIC_URL, OPENAI_API_KEY, ANTHROPIC_API_KEY)
   ```

2. From repo root:

   ```bash
   docker compose -f deploy/paperclip/docker-compose.yml --env-file deploy/paperclip/.env up --build -d
   ```

3. Open `http://localhost:3100` (or your `PAPERCLIP_PUBLIC_URL`). Complete first-time onboarding (create board user, then create company).

4. Run the setup script to create the FunctionFly company and initial org chart agents (see below).

## Option B: Run from Paperclip repo (official quickstart)

1. Clone and run Paperclip:

   ```bash
   git clone https://github.com/paperclipai/paperclip.git
   cd paperclip
   pnpm install
   cp .env.example .env
   # Set PAPERCLIP_DEPLOYMENT_MODE=authenticated, PAPERCLIP_DEPLOYMENT_EXPOSURE=private, BETTER_AUTH_SECRET
   pnpm dev
   ```

2. Open `http://localhost:3100`, complete onboarding, then create company and agents (see below).

## Create FunctionFly company and initial org chart

After Paperclip is running and you are logged in as the board user:

1. **Set CLI context** (if using Paperclip CLI from their repo):

   ```bash
   pnpm paperclipai context set --api-base http://localhost:3100 --company-id <company-id>
   ```

   Or create the company first via UI, then note the company ID from the URL or API.

2. **Create company** (if not already created via onboarding):
   - In UI: create a company named **FunctionFly**.
   - Or via API: `POST /api/companies` with `{"name":"FunctionFly","description":"Internal agent control plane for FunctionFly platform"}`.

3. **Create initial agents** (org chart). For each agent, use the Paperclip UI (Agents → Add agent) or the API:
   - **CTO** (manager, reports_to: null) — role e.g. `cto`, title "CTO".
   - **PlatformEngineer** (IC, reports_to: CTO) — role e.g. `engineer`, title "Platform Engineer".
   - **DevOps** (IC, reports_to: CTO) — role e.g. `devops`, title "DevOps".
   - **SupportTriage** (IC, reports_to: CTO) — role e.g. `support`, title "Support Triage".
   - **Security** (IC, reports_to: CTO) — role e.g. `security`, title "Security".

   API example (replace `COMPANY_ID` and use a valid session cookie or API key):

   ```bash
   # Create CTO first (no reports_to)
   curl -X POST "http://localhost:3100/api/companies/COMPANY_ID/agents" \
     -H "Content-Type: application/json" \
     -H "Cookie: <your-session-cookie>" \
     -d '{"name":"CTO","role":"cto","title":"CTO","adapterType":"http","adapterConfig":{}}'

   # Create PlatformEngineer (reports_to: CTO agent id)
   curl -X POST "http://localhost:3100/api/companies/COMPANY_ID/agents" \
     -H "Content-Type: application/json" \
     -H "Cookie: <your-session-cookie>" \
     -d '{"name":"PlatformEngineer","role":"engineer","title":"Platform Engineer","reportsTo":"<CTO_AGENT_ID>","adapterType":"http","adapterConfig":{}}'
   ```

4. **Create agent API keys** for each agent that will perform work (e.g. for the adapter or Cursor):
   - In UI: Agents → select agent → Create API key.
   - Or: `POST /api/agents/:agentId/keys` (see Paperclip docs).

5. **Optional**: Run the scripted setup if you have `curl` and `jq` and a board session:

   ```bash
   ./deploy/paperclip/scripts/setup-functionfly-company.sh
   ```

   (Set `PAPERCLIP_BASE_URL`, `PAPERCLIP_COOKIE` or `PAPERCLIP_API_KEY` in env as needed.)

## Ports and URLs

- Paperclip API + UI: default **3100**.
- FunctionFly orchestrator (for adapter/cost bridge): default **8080**.
- Ensure the Paperclip → FunctionFly adapter and cost bridge can reach the orchestrator (e.g. `http://host.docker.internal:8080` from a container, or `http://localhost:8080` from the same host).

## Next steps

- Configure the [Paperclip → FunctionFly webhook adapter](../../internal/api/handlers/paperclip/README.md) so heartbeats can trigger FunctionFly agent executions.
- Enable the [cost bridge](../../internal/paperclip/costbridge/README.md) so execution costs flow into Paperclip budgets.
- Follow [governance workflows](../../docs/PAPERCLIP_GOVERNANCE.md) for issue templates and approval gates.

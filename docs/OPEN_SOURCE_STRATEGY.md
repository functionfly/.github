# Open Source Strategy: What to Publish and Monorepo vs Split

This doc recommends **what to open source** and whether to keep a **single monorepo** or **split** into public + private repos.

---

## Recommendation: **Single public monorepo**

**Open the whole repo** (core platform + deploy config + docs), with these rules:

| Publish | Keep out of repo (or use placeholders) |
|--------|---------------------------------------|
| All of `cmd/`, `internal/`, `web/`, `migrations/`, `runtimes/`, `functions/`, `packaging/` | Real credentials (already in `.env*` and gitignore) |
| All of `deploy/` as **reference config** (Caddy, edge scripts, DB examples) | Real VPS IPs in scripts (see below: use env or placeholders) |
| All of `docs/` and `plans/` (generic guides, architecture) | Runbooks that contain real connection strings (keep those in a private runbook or secrets manager) |
| `*.env.example` only; never commit `.env`, `.env.production`, `deploy/database/production.env` | TLS private keys; `deploy/edge/certs-in/`, `certs-out/` (already gitignored) |

**Why monorepo:**

- One place for code, config, and docs. Contributors can run and deploy from the same tree.
- Your product (FunctionFly) is the platform; the deploy layout is useful reference for self-hosters and for your own ops.
- No sync burden between a “core” repo and a “deploy” repo.
- Splitting only makes sense if you have **strict** need to hide infra (exact VPS IPs, internal runbooks); then you’d move only those bits (see “If you split” below).

---

## Before going public: one more hardening step

**VPS IPs are currently hardcoded** in a few places. For a public repo, prefer **not** exposing your actual edge server IPs.

| Location | Current | Change |
|----------|---------|--------|
| `internal/monitoring/edge_stats.go` | `defaultEdgeNodes` = two real IPs | Use **env-only default**: if `EDGE_NODES` is unset, show no nodes (or a placeholder). Remove hardcoded IPs. |
| `.env.example` | Comment lists the two IPs | Replace with: `# EDGE_NODES=host1:Region A,host2:Region B` (no example IPs). |
| `deploy/edge/upload-certs.sh` | `NODES=( "217..." "209..." )` | Source from env, e.g. `EDGE_VPS_NODES` (space-separated), with no default in repo. |
| `deploy/edge/upload-certs-windows.ps1` | Same IPs | Same: read from env. |
| `deploy/edge/README-CERTS.md` | Instructions with real IPs | Use placeholders: “Replace `EDGE_VPS_NODE_1` and `EDGE_VPS_NODE_2` with your node IPs.” |
| `deploy/edge/vps-deploy.sh` | Comment with IPs | “Run on each edge VPS (set EDGE_VPS_NODES or pass host as arg).” |
| `deploy/dns/cloudflare-geo-dns.json` | Real IPs in geo targets | **Option A:** Add to `.gitignore` and commit `cloudflare-geo-dns.example.json` with placeholders. **Option B:** Keep in repo if that DNS config is for a public demo; otherwise use Option A. |

After this, the public repo contains no real infrastructure addresses; only your domain names (e.g. `functionfly.com`, `api.functionfly.com`) remain, which are fine for a product repo.

---

## If you split (public core + private ops)

Use this only if you **must** keep exact edge topology, runbooks, and deploy details private.

| Public repo | Private repo |
|------------|--------------|
| `cmd/`, `internal/`, `web/`, `migrations/`, `runtimes/`, `functions/`, `packaging/` | `deploy/` (or at least `deploy/edge/`, `deploy/database/` with real envs) |
| `docs/` that are user-facing or generic (QUICK_START, PRODUCTION_DEPLOYMENT, SECURITY, OBJECT_STORAGE, etc.) | `docs/` that are runbooks (e.g. STAGING with real credentials, DISASTER_RECOVERY_RUNBOOK, internal playbooks) |
| `*.env.example` only | Real `.env.staging`, `.env.production`, `deploy/database/production.env` |
| Generic deploy examples (e.g. Caddy with `your-domain.com`) | Your actual Caddyfiles and scripts with real IPs and functionfly.com |

**Pros:** Real VPS IPs, exact Caddy config, and runbooks never touch the public repo.  
**Cons:** Two repos to maintain; deploy config can drift from code; contributors don’t see “how you deploy” end-to-end.

---

## Summary

- **Recommended:** Single **public monorepo**, with no secrets (done), no real credentials in docs (done), and **VPS IPs moved to env or placeholders** so nothing sensitive is exposed.
- **Optional:** A small **private** repo or wiki for “runbooks only” (real connection strings, incident playbooks) while the main repo stays the single source of code and reference deploy config.
- **Split only if** you have a hard requirement to hide all deploy topology and runbooks; then use the table above to decide what lives in each repo.

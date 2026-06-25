# FunctionFly City Rankings™

> The live map of human + AI productivity.

City Rankings ranks cities (MSAs) by **real FunctionFly activity** — function
executions, deployments, founders, and new builders — with a per-capita score
so dense AI hubs can compete with megacities.

This document is the canonical reference for the schema, scoring formula,
storage layer, recompute job, and API surface. The user-facing vision lives
in [`.kilo/plans/1782018734195-city-rankings-plan.md`](../../.kilo/plans/1782018734195-city-rankings-plan.md).

---

## 1. What ships

| Capability | Status | Phase |
|---|---|---|
| Public leaderboard (top 100 metros) | ✅ | 1 |
| Metro detail + 30-day history | ✅ | 1 |
| Movers (gainers / losers) | ✅ | 1 |
| My-city widget (logged-in) | ✅ | 1 |
| Opt-out (per-user) | ✅ | 1 |
| k≥5 privacy threshold | ✅ | 1 |
| Hourly cron + Redis cache (5 min TTL) | ✅ | 1 |
| Per-capita scoring | ✅ | 1 |
| City resolution from user-typed "Location" | ✅ | 1 |
| City-proper leaderboard (toggle off MSA rollup) | ✅ | 1 |
| Anonymized top contributors per metro (k≥5) | ✅ | 1 |
| **State leaderboard** (US rolled up by state) | ✅ | 2 |
| **AI World Map** (3D globe on marketing site) | ✅ | 2 |
| **Tier-colored markers** (gold / blue / green) | ✅ | 2 |
| **Map points API** (lat/lon + tier) | ✅ | 2 |
| **IP-based geo fallback** (MaxMind → largest city in country) | ✅ | 2 |
| 500+ cities seeded (top global metros) | ✅ | 2 |
| **Sub-rankings** (Agent / Automation / Robotics / Startups / Open Source) | ✅ | 3 |
| **University Rankings** | ✅ | 4 |
| **City Wars** (quarterly bracket) | ✅ | 5 |
| **City Ambassadors** (profile badges, sync cron) | ✅ | 6 |
| **Monthly Report** ("State of AI Builders" PDF/HTML) | ✅ | 7 |
| **Company Rankings** (Top Agent Companies) | 🚧 future |

---

## 2. Data model

| Table | Purpose |
|---|---|
| `metro_areas` | Primary ranking unit (MSA, EU metro, etc.). Has `slug`, `name`, `country_code`, `population`. |
| `cities` | Individual city records. FK to `metro_areas`. Carries the canonical lat/lon and self-reported "Location" name. |
| `city_aliases` | Fuzzy-match lookup for user-typed "Location" strings. E.g. `austin`, `austin tx`, `austin, texas`. |
| `city_rankings` | Materialized score rows. UNIQUE(`metro_area_id`, `period_end`, `ranking_category`). |
| `users` (cols) | `city_id BIGINT`, `city_ranking_opted_out BOOLEAN`. |

Indexes:
- `idx_cities_country_state`, `idx_cities_metro`, `idx_cities_slug`
- `idx_city_aliases_alias_lower` (case-insensitive)
- `idx_users_city` (partial — only opted-in users)
- `idx_city_rankings_period_rank` (leaderboard reads)
- `idx_city_rankings_metro_period` (history reads)
- `idx_city_rankings_category_rank` (sub-ranking reads)

Migration: [`migrations/20260621170000_city_rankings.up.sql`](../migrations/20260621170000_city_rankings.up.sql).

---

## 3. Seed data

The first run loads `data/cities_seed.csv` (721 cities / 721 metros across
127 countries, top 100+ global cities by population + all major US launch
metros). The seeder is idempotent — rows upsert on slug — so it is safe to
run on every boot.

To add more cities: append rows to the CSV and restart the API. Production
should use a curated GeoNames-derived dump.

### IP-based geo fallback (plan §2)

If the user hasn't typed a `Location` on their profile, the leaderboard
still wants to count them — so a fresh signup from a non-US IP can land on
the board without typing anything. The flow:

1. Front-end POSTs to `/v1/city-rankings/me` (empty body).
2. The handler reads the caller's IP from `X-Forwarded-For` / `X-Real-IP` /
   `CF-Connecting-IP` (in that order), falling back to `RemoteAddr`.
3. The IP is resolved to a country via MaxMind GeoLite2 (free, downloaded
   automatically if `MAXMIND_LICENSE_KEY` is set — see
   [`internal/routing/geo_router.go`](../internal/routing/geo_router.go)).
4. The country maps to the largest active city in that country (one indexed
   query, result cached in-process for 1h).
5. The city is persisted via `users.city_id` so the hourly recompute picks
   it up.

When no MaxMind DB is available the resolver returns `not_found` and the
IP fallback becomes a no-op. The self-reported `Location` path always
remains the source of truth.

---

## 4. Scoring formula

```text
score_raw = (w_users / 100)    * log10(active_users   + 1)
          + (w_deploy / 100)   * log10(deployments    + 1)
          + (w_exec / 100)     * log10(executions_30d + 1)
          + (w_found / 100)    * log10(founder_cents  + 1)
          + (w_growth / 100)   * log10(new_users_30d  + 1)

score_per_capita = score_raw * 100_000 / metro_population
```

| Signal | Weight | Source |
|---|---:|---|
| Active users (30d) | 30 | `users.last_active_at` |
| Deployments (30d) | 25 | `registry_functions.owner_user_id` |
| Function executions (30d) | 20 | `registry_function_executions.user_id` |
| Founder earnings (lifetime) | 15 | `affiliate_codes.total_earnings_cents` (capped at 1e12) |
| New users (30d) | 10 | `users.created_at` |

Log scaling keeps megacities from dominating. Per-capita is the headline.

Weights are normalized — only their **relative** ratios matter. Default values
are tunable in code; v1 uses the plan defaults.

---

## 5. Privacy & opt-out

- **k-anonymity ≥ 5**: metros with fewer than 5 active (non-opted-out) users
  in the last 30 days are hidden from the public leaderboard. They still get
  a `city_rankings` row, but the leaderboard query filters them out.
- **Opt-out**: every user is counted by default. The dashboard Settings page
  exposes a "Hide me from city rankings" toggle wired to
  `POST /v1/users/me/city-ranking-opt-out`. All aggregations filter
  `WHERE NOT city_ranking_opted_out`.

To extend: the privacy threshold is a single constant
`MinActiveUsersForPublic` in [`internal/storage/cityranking/normalizer.go`](../internal/storage/cityranking/normalizer.go).

---

## 6. Recompute cycle

A cron job in `internal/jobs/cityranking` runs every hour (`CITY_RANKING_CRON`,
default `0 * * * *`):

1. Truncate `now` to the hour → `period_end`.
2. Iterate every active metro in parallel (10 workers).
3. For each: compute signals, score, upsert into `city_rankings`.
4. Assign ranks (`ROW_NUMBER() OVER (ORDER BY score_per_capita DESC)`),
   preserving the previous rank in `prev_rank_position`.
5. Invalidate the Redis cache so subsequent reads pick up the new period.

Per-metro errors are logged and skipped — the cycle always finishes. Jitter is
the cron scheduler's responsibility (set `CITY_RANKING_CRON` to a non-standard
expression to spread load).

Manual trigger: `go run ./scripts/smoke_cityranking` (smoke test that
inserts 30 fake users, runs one cycle, prints the leaderboard).

---

## 7. API surface

All endpoints are mounted under `/v1/city-rankings` on the orchestrator.

| Method | Path | Auth | Purpose |
|---|---|---|---|
| GET | `/v1/city-rankings` | public | Top 100 metros (optional `?country=US`, `?limit=100`) |
| GET | `/v1/city-rankings/categories` | public | List available ranking categories + descriptions |
| GET | `/v1/city-rankings/movers?direction=gainers\|losers` | public | Biggest rank deltas this period |
| GET | `/v1/city-rankings/states` | public | US state leaderboard (rolled up from metros) |
| GET | `/v1/city-rankings/states/{code}` | public | Single state detail |
| GET | `/v1/city-rankings/map` | public | All ranked metros with lat/lon + tier label |
| GET | `/v1/city-rankings/cities` | public | City-proper leaderboard (toggle off MSA rollup) |
| GET | `/v1/city-rankings/resolve-by-ip?ip=1.2.3.4` | public | Best-guess city from IP (MaxMind-backed) |
| GET | `/v1/city-rankings/{slug}` | public | Single metro detail + 30-day history |
| GET | `/v1/city-rankings/{slug}/builders` | public | Anonymized top contributors (k≥5 enforced) |
| GET | `/v1/city-rankings/me` | auth | Caller's metro + rank (if opted in) |
| POST | `/v1/city-rankings/me` | auth + CSRF | Set caller's city (body: `slug` / `input`; IP fallback) |
| POST | `/v1/city-rankings/resolve` | auth + CSRF | Resolve user-typed "Location" → city |
| GET | `/v1/users/me/city-ranking-opt-out` | auth | Current opt-out state |
| POST | `/v1/users/me/city-ranking-opt-out` | auth + CSRF | Toggle opt-out |

Read endpoints are cached in Redis for 5 minutes; the recompute job warms the
cache on the next cycle.

---

## 8. Caching

Single Redis namespace `cityrank:*`. TTL = 5 minutes for all read keys.
Invalidation:
- After a successful recompute cycle (`CityRankingRecompute.runCycle`).
- Implicit (TTL expiry) otherwise.

The cache is best-effort. Any Redis error falls through to a direct DB read.

---

## 9. Operations

### Environment variables

| Variable | Default | Purpose |
|---|---|---|
| `CITY_RANKING_CRON` | `0 * * * *` | Cron expression for the recompute job |
| `CITY_RANKING_ENABLED` | `true` | Set to `false` to disable the cron entirely |

### Tuning

- To change the per-capita base, edit `Compute()` in
  [`internal/storage/cityranking/scorer.go`](../internal/storage/cityranking/scorer.go).
- To change the privacy threshold, edit `MinActiveUsersForPublic` in
  [`internal/storage/cityranking/normalizer.go`](../internal/storage/cityranking/normalizer.go).
- To add new metros/cities, append rows to `data/cities_seed.csv` and restart.

### Common failures

| Symptom | Cause | Fix |
|---|---|---|
| All scores are 0 | `metro_areas` empty | Boot the orchestrator (auto-seeds) or manually load the CSV |
| Cache hit shows stale data | TTL hasn't expired | Wait 5 min, or call `redis-cli FLUSHDB` |
| Leaderboard empty | Below k=5 threshold | Add more users, or lower `MinActiveUsersForPublic` |
| Cron never fires | `CITY_RANKING_ENABLED=false` | Unset / set to `true` |

---

## 10. Phase 2 — State rankings + AI World Map

Phase 2 added two new public surfaces on top of the core city rankings: a
state-level rollup and a 3D globe visualization.

### 10.1 State rankings

`cities.state_code` is used to aggregate the metro-level scores into a
state-level leaderboard. Pure SQL aggregation over `city_rankings` +
`metro_areas` + `cities` — no extra columns, no recompute changes.

| Endpoint | Returns |
|---|---|
| `GET /v1/city-rankings/states?country=US` | Top states by per-capita score |
| `GET /v1/city-rankings/states/{code}?country=US` | Single state detail |

Privacy threshold applies: a state is only listed if its summed
`active_users` (across all its metros) is ≥ 5. Ranking is by **state
per-capita** — sum of metro per-capita scores, normalized to 100k state
population. This favors smaller, dense AI hubs over megacities at the
state level too (e.g. California outranks New York despite smaller pop).

### 10.2 AI World Map

A 3D visualization on the marketing site `/rankings` page. Built with
`@react-three/fiber` (already in the site's stack) using a procedural
wireframe globe — no texture downloads, no Three.js globe library.

Components:
- `web/site/src/components/CityRankingsGlobe.tsx` — wireframe sphere +
  pulsing city markers, auto-rotating, click-to-navigate.
- 2D SVG fallback for low-power devices without WebGL.
- Tier color mapping:
  - **Gold** — per-capita ≥ 0.20
  - **Blue** — per-capita ≥ 0.05
  - **Green** — below
- Thresholds are tunable in
  `internal/storage/cityranking/scorer.go` (`TierThresholds`).

Data flow: marketing site fetches `/v1/city-rankings/map` (returns all
ranked metros with lat/lon + tier label, cached in Redis for 5 min).
The same endpoint can power any third-party embed.

### 10.3 New API surface (phase 2)

| Method | Path | Auth | Returns |
|---|---|---|---|
| GET | `/v1/city-rankings/states` | public | State leaderboard |
| GET | `/v1/city-rankings/states/{code}` | public | Single state (optional `?country=`) |
| GET | `/v1/city-rankings/map` | public | All ranked metros with lat/lon + tier |

All three are cached in Redis for 5 minutes and invalidated on the next
hourly recompute.

---

1. **Sub-rankings** — add `ranking_category` enum (agents / automation / robotics / open_source / startups). Each gets its own weights.
2. **University Rankings** — add `users.university_id` + `universities` table. Same formula.
3. **City Ambassadors** — top contributor per city gets a `role = 'ambassador:<slug>'`. Badge, local events, beta access.
4. **Monthly Report** — "State of AI Builders" PDF/HTML generated by cron. Posted to docs site + emailed to subscribers.
5. **City Wars** — quarterly bracket ("Austin vs Dallas"). Game layer on top of rankings #1–#4.
6. **More countries** — extend `data/cities_seed.csv` from the launch 70 metros to the full top 500 worldwide.

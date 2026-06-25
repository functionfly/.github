# FunctionFly University Rankings™

> Top universities by **real** FunctionFly activity.

University Rankings ranks schools by the activity of their affiliated
users — function executions, deployments, founders, and new builders — with
a per-capita score so a small CS lab at a teaching university can outrank a
state flagship. Same scoring formula, same k-anonymity, same hourly recompute
as the [City Rankings](./CITY_RANKINGS.md).

---

## What ships (Phase 1)

| Capability | Status |
|---|---|
| Public leaderboard (top 100) | ✅ |
| University detail (slug) | ✅ |
| Self-reported resolution (`/resolve`) | ✅ |
| My-university widget (`/me`) | ✅ |
| Opt-out (per-user) | ✅ |
| k≥5 privacy threshold | ✅ |
| Hourly cron + Redis cache (5 min TTL) | ✅ |
| 6 categories (composite + agents / automation / robotics / startups / open_source) | ✅ |
| 170+ universities seeded (US + international) | ✅ |
| Marketing site `/universities` | ✅ |
| Dashboard page + API client | ✅ |
| Category filter UI (Composite / Agent / Automation / Robotics / Startup / Open Source) | ✅ |
| City → University cross-link (e.g. "MIT → Boston Metro") | ✅ |

---

## Data model

| Table | Purpose |
|---|---|
| `universities` | Primary ranking unit. Has `slug`, `name`, `country_code`, `state_code`, `student_count`, `institution_type`. |
| `university_aliases` | Fuzzy-match lookup for user-typed "University" strings. E.g. `mit`, `mit ma`, `massachusetts institute of technology`. |
| `university_rankings` | Materialized score rows. UNIQUE(`university_id`, `period_end`, `ranking_category`). |
| `users` (cols) | `university_id BIGINT`, `university_ranking_opted_out BOOLEAN`. |

Indexes:
- `idx_universities_country`, `idx_universities_active_students`
- `idx_university_aliases_alias_lower` (case-insensitive)
- `idx_users_university` (partial — only opted-in users)
- `idx_university_rankings_period_rank`

Migration: `migrations/20260621180000_university_rankings.up.sql`.

---

## Seed data

`data/universities_seed.csv` (170+ rows) covers:

- **US**: Top 100 by selectivity + enrollment (MIT, Stanford, Cal, Ivy
  League, R1 public flagships, top liberal arts colleges, top HBCUs, top
  community colleges).
- **International**: UK (Oxbridge, Russell Group), Canada (U15), China
  (C9 + HK + TW), India (IIT + IISc + BITS), Japan (Top 7), Singapore (NUS /
  NTU / SMU), Australia (Group of Eight), EU (top tech), Brazil, Mexico,
  Korea, NZ, Russia, Switzerland.

Header row required; the seeder upserts on slug so re-running is safe. To
add a new university: append a row to the CSV and restart the API.

---

## Scoring formula

Identical to the [city leaderboard](./CITY_RANKINGS.md#scoring-formula). The
per-capita denominator is `student_count` instead of metro population, so
small CS departments naturally rank above large state flagships when they
punch above their weight.

Weights by category (composite, agents, automation, startups, open_source)
live in `internal/storage/universityranking/scorer.go::CategoryWeights`.

---

## API surface

All endpoints are mounted under `/v1/university-rankings` on the
orchestrator.

| Method | Path | Auth | Purpose |
|---|---|---|---|
| GET | `/v1/university-rankings` | public | Top 100 universities (optional `?country=US`, `?limit=100`, `?category=composite`) |
| GET | `/v1/university-rankings/{slug}` | public | Single university detail + ranking |
| GET | `/v1/university-rankings/me` | auth | Caller's university + rank (if opted in) |
| POST | `/v1/university-rankings/resolve` | auth + CSRF | Resolve user-typed "University" → university |
| GET | `/v1/users/me/university-ranking-opt-out` | auth | Current opt-out state |
| POST | `/v1/users/me/university-ranking-opt-out` | auth + CSRF | Toggle opt-out |

The `category` parameter defaults to `composite`. Supported values:
`composite | agents | automation | startups | open_source`.

---

## Privacy (k-anonymity)

A university with fewer than 5 active opted-in builders is **omitted** from
the leaderboard — not ranked at #N, not hidden behind a 404, not shown as
"less than 5". The handler returns an empty list. The recompute job
materializes ranking rows only for universities that pass the threshold; the
caller-side `ListUniversities` query re-applies the filter as a defense in
depth.

Opting out is per-user and applies to **all** aggregations (city, university,
and the future state / project). The opt-out column is `university_ranking_opted_out`
and is checked in the same SQL that counts active users so the privacy
guarantee holds in the materialized rows too.

---

## Recompute

`internal/jobs/universityranking/recompute.go` runs every hour (cron: `0 * * *
*` by default, override with `UNIVERSITY_RANKING_CRON`). One cycle:

1. List all active universities.
2. For each (university, category) pair, compute signals and score.
3. Re-rank by per-capita, assign rank positions, fetch prev-period for deltas.
4. Upsert into `university_rankings` (idempotent on `(university_id, period_end, category)`).
5. Warm the Redis cache for global + 9 top countries.

The job uses 8 parallel workers; one full cycle for 170 universities × 5
categories takes ~5 seconds.

---

## Frontend

- **Dashboard** (`web/dashboard/src/pages/UniversityRankingsPage/`):
  full leaderboard with country filter, "your university" callout, animated
  rank deltas, and the same privacy copy as the city page.
- **Marketing** (`web/site/src/pages/universities.astro` + `components/UniversityRankingsPage.tsx`):
  SEO-friendly server-rendered page with Open Graph meta. Hits the same
  `/v1/university-rankings` endpoint.

---

## Future work

- Surface the 5 sub-rankings (agents / automation / startups / open source)
  in the dashboard UI as a category switcher.
- Cross-link universities to their metro (e.g. "MIT → Boston, MA" on the
  detail page) to drive traffic between the two leaderboards.
- "University referral" — top user in each university gets a `role =
  'ambassador:<uni_slug>'` so they can moderate a school-specific channel.
- Bootcamp / institute support — the `institution_type` column already
  distinguishes; we just need a UI toggle.

See `.kilo/plans/1782018734195-city-rankings-plan.md` §8 for the full
roadmap.

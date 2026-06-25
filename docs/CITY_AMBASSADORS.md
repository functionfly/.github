# FunctionFly City Ambassadors

> Top builder per city. Promoted automatically from the live leaderboard.

The Ambassador program identifies the single highest-scoring active opted-in
user in every city that passes the k=5 privacy threshold, and surfaces them
publicly as the city's ambassador. The system runs hourly and is fully
hands-off — the top builder of the previous cycle is replaced automatically
when a new builder takes the lead.

---

## How it works

1. **Hourly sync** (`internal/jobs/cityambassador/sync.go`, cron `5 * * * *`)
   runs 5 minutes after the city recompute job finishes. It iterates over
   every metro that has at least `MinActiveUsersForPublic` (5) active
   builders in the latest `city_rankings` row.
2. For each eligible metro, it finds the **highest-scoring active opted-in
   user** in that metro via `TopBuilderForMetro`. The score uses the same
   per-metro activity formula as the leaderboard (deployments, executions,
   founder earnings, active users).
3. If a different user is currently ambassador, the old row's `revoked_at`
   is set to now and a new row is inserted. The partial unique index
   `uq_city_ambassadors_active` guarantees **at most one active ambassador
   per metro** at any time.
4. For metros that drop below the threshold (e.g. builders churned), the
   active ambassador is automatically revoked.

## Privacy

The k=5 threshold from the leaderboard is reused. A metro with fewer than
5 active opted-in builders:

- has **no row in `city_rankings`** (filtered by the leaderboard query)
- has **no ambassador** (the sync only touches eligible metros)
- returns **404** from `/v1/city-rankings/{slug}/ambassador`

This is the same privacy contract as the leaderboard, so users can be sure
"no ambassador" doesn't leak the number of builders.

## API

All endpoints are mounted under `/v1/city-rankings`.

| Method | Path | Auth | Purpose |
|---|---|---|---|
| GET | `/v1/city-rankings/ambassadors?country=US` | public | List active ambassadors (one per metro), sorted by metro population. |
| GET | `/v1/city-rankings/{slug}/ambassador` | public | Single ambassador for a city. 404 if none. |

Response shape for `/v1/city-rankings/ambassadors`:

```json
{
  "total": 3,
  "entries": [
    {
      "metro_id": 75,
      "metro_slug": "san-francisco-oakland-hayward-ca",
      "metro_name": "San Francisco, California",
      "country_code": "US",
      "state_code": "CA",
      "city_slug": "san-francisco-ca-us",
      "user_id": "00000000-0000-0000-0000-000000000003",
      "name": "Jane Builder",
      "profile_number": 100003,
      "promoted_at": "2026-06-21T20:23:53Z",
      "source": "auto"
    }
  ],
  "privacy_min_active_users": 5
}
```

Note that `email` is **never** in the response — the front-end links to the
public profile via `profile_number` (see `users.profile_number`).

## Data model

`city_ambassadors` table:

| Column | Type | Notes |
|---|---|---|
| `id` | `bigserial` | PK |
| `metro_id` | `bigint` | FK → `metro_areas.id` ON DELETE CASCADE |
| `user_id` | `uuid` | FK → `users.id` ON DELETE CASCADE |
| `promoted_at` | `timestamptz` | set to NOW() on (re)promotion |
| `revoked_at` | `timestamptz` | NULL while active; set on demotion |
| `source` | `text` | `auto` (sync) or `manual` (admin override) |

Indexes:
- `(metro_id, user_id)` UNIQUE — one row per (metro, user).
- `(metro_id) WHERE revoked_at IS NULL` UNIQUE — at most one active per metro.
- `(user_id) WHERE revoked_at IS NULL` — fast "is this user an ambassador anywhere?".

Migration: `migrations/20260621190000_city_ambassadors.up.sql`.

## Frontend

- **Dashboard** (`/rankings/cities/:slug`): an "Ambassador" card appears
  below the rank stats when the city has one, with the user's public name
  and a `source: auto` chip.
- **Dashboard list** (`/ambassadors`): country-filtered leaderboard of
  all current ambassadors, linked back to their city.
- **Marketing site** (`/ambassadors`): public-facing global list, same
  Open Graph meta as the rankings page.

## Future work

- **Admin override** (`POST /v1/admin/ambassadors/{metro_slug}`) — let staff
  manually swap an ambassador (e.g. for community events). Schema already
  supports it via `source = 'manual'`; just needs the auth-protected route.
- **Ambassador program landing page** — give the badge more weight on the
  docs site (avatar, profile link, "apply to be ambassador" form).
- **Local events** — the original plan §8 #6 mentions local events
  co-organized with each ambassador; the schema doesn't need changes but
  the front-end does.
- **University Ambassadors** — same plumbing for the university leaderboard
  when that ships its own ambassador tier.

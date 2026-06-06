# Plugin Migration Guide: `extension_registry` → `plugins`

The legacy `extension_registry` API surface is deprecated. This guide
walks API and dashboard clients through the migration to the
production `plugins` system.

## Timeline

| Date | Event |
|------|-------|
| **2026-06-06** | Deprecation headers live; migration `20260605130000_deprecate_extension_registry` applied to staging. `extension_registry.deprecated_at` populated. |
| **2026-06-13** | Headers hit production (1 week after staging soak). |
| **2026-07-06** | **Sunset** — `extension_registry` table, `studio.ExtensionsHandler`, the `useStudioExtension` hook, and the 7 `/v1/extensions/*` routes are removed. Requests will return `404 extension_registry_removed`. |

## Response headers (effective 2026-06-06)

Every response from the legacy endpoints carries:

```
Deprecation: true
Sunset: Sun, 06 Jul 2026 00:00:00 GMT
X-Deprecated-Use: /plugins
Link: </docs/PLUGIN_MIGRATION.md>; rel="deprecation"
```

`Sunset` follows [RFC 8594](https://datatracker.ietf.org/doc/html/rfc8594).
`Deprecation` follows [RFC 9745](https://datatracker.ietf.org/doc/html/rfc9745).

## Endpoint mapping

| Legacy | Replacement | Notes |
|--------|-------------|-------|
| `GET /v1/extensions` | `GET /v1/plugins` | Returns the same `extensions[]` shape as `plugins[]`. `category` and `status` filters preserved. `limit`/`offset` unchanged. |
| `POST /v1/extensions/{id}/install` | `POST /v1/plugins` | Body is the new `Manifest` shape; `name`, `version`, `entry_point`, `runtime`, `permissions[]` are required. The legacy free-form body is rejected with `400 invalid_manifest`. |
| `DELETE /v1/extensions/{id}` | `DELETE /v1/plugins/{id}` | Tenant scoping unchanged. |
| `POST /v1/extensions/{id}/enable` | `POST /v1/plugins/{id}/enable` | Requires a `plugin_sandboxes` row for `runtime`, `infrastructure`, and `ai_tool` plugin types (409 `sandbox_required`). |
| `POST /v1/extensions/{id}/disable` | `POST /v1/plugins/{id}/disable` | |
| `PUT /v1/extensions/{id}/config` | `PUT /v1/plugins/{id}/config` | Body unchanged (`config: {key: value}[]`). Secrets redaction applied automatically. |
| `GET /v1/extensions/hooks` | `GET /v1/plugins/{id}/hooks` | Per-plugin; the legacy tenant-wide list is dropped. |

## Body shape changes

### Install

Legacy:

```json
{
  "name": "github-actions",
  "version": "1.2.0",
  "description": "...",
  "author_name": "...",
  "category": "ci",
  "permissions": ["network", "agents"],
  "hooks": ["pre-deploy", "post-deploy"],
  "size_kb": 42
}
```

New (`/v1/plugins`):

```json
{
  "manifest": {
    "name": "github-actions",
    "version": "1.2.0",
    "entry_point": "index.js",
    "runtime": "nodejs",
    "description": "...",
    "category": "ci",
    "permissions": [
      {"type": "network", "action": "github.com"},
      {"type": "agents", "action": "*"}
    ]
  }
}
```

`hooks` is no longer a top-level field; it is derived from
`manifest.entry_point` metadata. Custom hook arrays are dropped.

### Config

Unchanged at the wire level. The server now redacts values for keys
matching `(?i)(token|secret|password|api[_-]?key)` before logging and
storage; clients that read config back get the redacted view.

## Dashboard migration

The dashboard already uses the `plugins` API for all three target
components (`PluginManager`, `PluginUpdateCenter`,
`PluginPermissionsViewer`). The legacy `useStudioExtension` hook is
orphaned — its only consumer is `StudioPage.tsx.bak`, and the file is
deleted in the same removal PR.

If you have a third-party dashboard or external integration that
imports `useStudioExtension`, switch to `usePlugin` (see
`web/dashboard/src/hooks/usePlugin.ts`).

## Pre-sunset checklist (clients)

- [ ] No 4xx/5xx from `/v1/extensions/*` in your error reporting for
      the previous 7 days.
- [ ] All install flows POST to `/v1/plugins` with the new manifest
      shape.
- [ ] All enable flows have a `plugin_sandboxes` row ready.
- [ ] All config values that match the secret pattern are moved to the
      platform secrets store, not inlined in `config`.
- [ ] All hook consumers switched from the tenant-wide list to
      `GET /v1/plugins/{id}/hooks`.

## Pre-sunset checklist (operators)

- [ ] Verify staging logs are silent for `extension_registry:
      deprecated endpoint called` for at least 14 days before the
      2026-07-06 cut.
- [ ] Confirm no paying tenant has any `extension_registry` rows in
      the last 7 days of activity (`SELECT tenant_id, MAX(updated_at)
      FROM extension_registry GROUP BY tenant_id HAVING MAX(updated_at)
      > NOW() - INTERVAL '7 days'`).
- [ ] Schedule the follow-up migration that drops the table for
      2026-07-06 with a 24-hour buffer for one more rollback window.

## Removal PR (2026-07-06)

The final removal PR is planned as `chore/plugin-drop-extension-registry`
and contains:

- `migrations/20260706*_drop_extension_registry.up.sql` — `DROP TABLE
  extension_registry CASCADE`.
- Delete `internal/api/handlers/studio/extensions.go` and
  `extensions_handler.go`.
- Delete the 7 `/v1/extensions/*` routes from `routes_platform.go`.
- Remove `studioExtensionsHandler` from the `Server` struct and its
  constructor.
- Delete `web/dashboard/src/hooks/useStudioExtension.ts`.

## Questions

Open a thread in `#studio-plugins` on Slack, or file an issue with the
`plugin-deprecation` label.

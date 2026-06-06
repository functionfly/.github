---
title: Make a function MCP-compatible
description: Publish a function to the FunctionFly MCP Function Registry. One toggle, one JSON-RPC transport, live in 30 seconds.
sidebar:
  label: Publish MCP
  order: 2
---

This page walks through making an existing FunctionFly function
**MCP-compatible** — that is, exposing it through the public
[MCP Function Registry](https://functionfly.com/registry) so any
MCP-compatible AI agent (Claude, Cursor, VS Code, etc.) can discover
and call it.

There are two paths: from the dashboard, or programmatically via the
publish API.

## From the dashboard

1. Open the function in your dashboard (`/functions/{author}/{name}`).
2. Click **Settings** in the function header.
3. Toggle **"MCP-compatible"** to on.
4. (Optional) Set a custom **tool name override** (defaults to
   `{author}__{name}`).
5. (Optional) Set the **rate limit** in calls per minute (default 60).
6. Save.

The function appears on [functionfly.com/registry](https://functionfly.com/registry)
within ~30 seconds and is immediately callable from any MCP host that
has the FunctionFly MCP server installed.

## From the publish API

When publishing a new version (or a brand-new function), include an
`mcp` block in the publish payload:

```bash
curl -X POST https://api.functionfly.com/v1/functions/publish \
  -H "Authorization: Bearer $FUNCTIONFLY_API_KEY" \
  -H "Content-Type: application/json" \
  -d '{
    "author": "alice",
    "name": "summarize_pdf",
    "version": "1.2.0",
    "manifest": { ... },
    "source": { ... },
    "mcp": {
      "enabled": true,
      "transports": ["streamable-http"],
      "expose_input_schema": true,
      "tool_name_override": "",
      "rate_limit_per_min": 60,
      "allowlist_origins": []
    }
  }'
```

Field reference:

| Field | Type | Default | Notes |
|---|---|---|---|
| `enabled` | bool | false | Master switch |
| `transports` | string[] | `["streamable-http"]` | Allowed: `streamable-http`, `stdio` |
| `expose_input_schema` | bool | true | If true, the manifest's input schema is exposed as the MCP `inputSchema` |
| `expose_output_schema` | bool | false | If true, the manifest's output schema is also exposed |
| `tool_name_override` | string | `{author}__{name}` | 1-64 chars, `[a-zA-Z0-9_-]` only |
| `rate_limit_per_min` | int | 60 | Per-caller, per-function. 1-10000 |
| `allowlist_origins` | string[] | `[]` | Reserved for future CORS gating |

Validation: the publish request fails with a 400 if any of these are
out of bounds. This guarantees you never have a function visible in the
registry with a broken configuration.

## Updating an existing function

`PATCH /v1/functions/{author}/{name}/mcp` accepts a partial update:

```bash
curl -X PATCH https://api.functionfly.com/v1/functions/alice/summarize_pdf/mcp \
  -H "Authorization: Bearer $FUNCTIONFLY_API_KEY" \
  -H "Content-Type: application/json" \
  -d '{
    "rate_limit_per_min": 120,
    "tool_name_override": "summarize"
  }'
```

You can flip `enabled` between `true` and `false` to temporarily hide a
function from the registry without un-publishing it.

## Quality bar

The "Verified MCP" badge (green checkmark) is awarded to functions that
pass all of:

- trust score ≥ 80
- malware scan clean
- ≥ 100 invocations
- ≥ 30 days old
- owner has a verified email

Verified tools get a small ranking boost in `/v1/mcp/tools` and a visual
badge on [functionfly.com/registry](https://functionfly.com/registry).

## Deprecation

If you ship a breaking change to a function's input schema, mark the old
version as deprecated in your manifest. The MCP tool list will surface the
deprecation notice in `_meta.functionfly.deprecation`, and Claude Desktop
will warn the user before invoking the deprecated version.

```json
{
  "mcp": {
    "enabled": true,
    "deprecation": {
      "message": "Use v2.0: rename 'url' to 'document_url'",
      "sunset_at": "2026-09-01T00:00:00Z",
      "replacement": "alice__summarize_pdf_v2"
    }
  }
}
```

## Next steps

- Browse the [live registry](https://functionfly.com/registry) to see what good looks like.
- Read the [API reference](./api) for the full JSON-RPC + HTTP contract.

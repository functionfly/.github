---
title: MCP API Reference
description: Full JSON-RPC 2.0 reference for the FunctionFly MCP Function Registry.
sidebar:
  label: API Reference
  order: 3
---

The FunctionFly MCP server implements the [Model Context Protocol spec 2025-03-26](https://modelcontextprotocol.io). All traffic is JSON-RPC 2.0 over the
streamable-HTTP transport, plus a small set of public, non-JSON-RPC HTTP
endpoints for SEO and discovery.

## Public HTTP endpoints

| Method | Path | Auth | Cache |
|---|---|---|---|
| GET | `/v1/mcp/manifest` | none | 5 min |
| GET | `/v1/mcp/tools` | none | 1 min |
| GET | `/v1/mcp/health` | none | — |
| POST | `/v1/mcp` | Bearer (for `tools/call`) | — |

### `GET /v1/mcp/manifest`

Returns the server identity, capabilities, and transport endpoints.

```json
{
  "name": "FunctionFly MCP Function Registry",
  "title": "The default registry of callable functions for AI agents",
  "version": "1.0.0",
  "protocol_version": "2025-03-26",
  "capabilities": {
    "tools":     { "listChanged": true },
    "resources": { "subscribe": false },
    "prompts":   false,
    "logging":   { "level": "info" }
  },
  "transport": ["streamable-http", "stdio"],
  "endpoints": {
    "streamable_http": "https://api.functionfly.com/v1/mcp",
    "stdio_package":   "@functionfly/mcp-server"
  },
  "stats": {
    "total_functions":        1284,
    "mcp_enabled_functions":  312,
    "last_updated":           "2026-06-01T22:00:00Z"
  }
}
```

### `GET /v1/mcp/tools`

Returns the public, paginated list of MCP-enabled tools. **Not** a
JSON-RPC envelope so that web crawlers and the marketing site can index
it without parsing JSON-RPC.

Query parameters:

| Param | Type | Default | Notes |
|---|---|---|---|
| `q` | string | — | Free-text search on name/title/description |
| `category` | string | — | Filter by category (e.g. `document-processing`) |
| `min_trust` | float | 0 | Minimum trust score (0-100) |
| `limit` | int | 100 | Page size (max 500) |
| `offset` | int | 0 | Offset (use `cursor` for stable pagination) |
| `cursor` | string | — | Opaque cursor returned in `next_cursor` |

```json
{
  "tools": [
    {
      "name": "alice__summarize_pdf",
      "title": "Summarize PDF",
      "description": "Extracts and summarizes the key points of any PDF document.",
      "inputSchema": {
        "type": "object",
        "properties": { "url": { "type": "string", "format": "uri" } },
        "required": ["url"]
      },
      "annotations": {
        "readOnlyHint": true,
        "openWorldHint": true,
        "category": "document-processing"
      },
      "_meta": {
        "functionfly": {
          "author": "alice",
          "name": "summarize_pdf",
          "version": "1.2.0",
          "trust_score": 92,
          "trust_tier": "verified",
          "verified_mcp": true,
          "homepage": "https://functionfly.com/@alice/v1/fx/summarize_pdf",
          "tags": ["pdf", "summary", "ai"],
          "runtime": "node20"
        }
      }
    }
  ],
  "next_cursor": "eyJpZCI6IjQ1NiJ9",
  "total": 312,
  "generated_at": 1749055200
}
```

## JSON-RPC methods (`POST /v1/mcp`)

Every request is a JSON-RPC 2.0 envelope. The server supports
notification frames (no `id` field) and a single request per call (batch
requests are parsed but return an error in v1).

### `initialize`

Required before any other method. Returns server info and capabilities.

Request:
```json
{ "jsonrpc": "2.0", "id": 1, "method": "initialize",
  "params": { "protocolVersion": "2025-03-26",
              "capabilities": {},
              "clientInfo": { "name": "Claude Desktop", "version": "1.2.3" } } }
```

Response:
```json
{ "jsonrpc": "2.0", "id": 1,
  "result": {
    "protocolVersion": "2025-03-26",
    "capabilities": { "tools": { "listChanged": true } },
    "serverInfo": { "name": "FunctionFly MCP Function Registry", "version": "1.0.0" }
  } }
```

### `notifications/initialized`

A notification (no `id`). Sent by the client after it has processed the
`initialize` response. The server returns no response.

### `tools/list`

Returns a page of tool definitions. Paginated with opaque cursors.

Request:
```json
{ "jsonrpc": "2.0", "id": 2, "method": "tools/list",
  "params": { "category": "document-processing", "minTrust": 80, "limit": 50 } }
```

Response:
```json
{ "jsonrpc": "2.0", "id": 2,
  "result": {
    "tools": [ ... ],
    "nextCursor": "eyJpZCI6IjQ1NiJ9"
  } }
```

### `tools/call`

Invoke a tool. **Requires `Authorization: Bearer <ffp_…>`**.

Request:
```json
{ "jsonrpc": "2.0", "id": 3, "method": "tools/call",
  "params": { "name": "alice__summarize_pdf",
              "arguments": { "url": "https://example.com/doc.pdf" } } }
```

Response (success):
```json
{ "jsonrpc": "2.0", "id": 3,
  "result": {
    "content": [{ "type": "text", "text": "{\"summary\": \"...\"}" }],
    "structuredContent": { "summary": "..." },
    "isError": false,
    "_meta": { "functionfly": { "execution_id": "fx_abc", "duration_ms": 842 } }
  } }
```

Response (error — JSON-RPC `error` envelope):
```json
{ "jsonrpc": "2.0", "id": 3,
  "error": { "code": -32005, "message": "Tool execution failed",
             "data": { "functionfly": { "status_code": 500, "error_code": "TIMEOUT" } } } }
```

### `resources/list` / `prompts/list` / `logging/setLevel`

These MCP methods are accepted but return empty / no-op responses. The
FunctionFly registry currently exposes tools only.

## Error codes

JSON-RPC 2.0 reserved + FunctionFly server-defined:

| Code | Meaning | When |
|---|---|---|
| `-32700` | Parse error | Malformed JSON body |
| `-32600` | Invalid request | Missing `jsonrpc` / `method` |
| `-32601` | Method not found | Unknown method |
| `-32602` | Invalid params | Schema validation failed |
| `-32603` | Internal error | Unexpected server error |
| `-32001` | Tool not found | No matching `{author}__{name}` |
| `-32002` | Tool disabled | Function exists but `mcp.enabled = false` |
| `-32003` | Auth required | Missing/invalid Bearer on `tools/call` |
| `-32004` | Rate limited | Per-function rate limit hit |
| `-32005` | Execution failed | Function returned an error |
| `-32006` | Function private | Visibility != public |
| `-32007` | Malware blocked | Function failed security scan |
| `-32008` | Origin not allowed | Streamable-HTTP origin block (reserved) |
| `-32009` | Payload too large | `arguments` exceeds 256 KiB |

## CORS

The `/.well-known/functionfly.json`-style CORS policy applies:

- `Access-Control-Allow-Origin: *` for the public endpoints.
- `Access-Control-Allow-Methods: GET, POST, OPTIONS`.
- `Access-Control-Allow-Headers: Content-Type, Authorization, Mcp-Session-Id`.
- `Access-Control-Max-Age: 86400`.

---
title: MCP (Model Context Protocol)
description: Expose FunctionFly functions as callable tools for AI agents via MCP
---

# MCP — Model Context Protocol

The **Model Context Protocol (MCP)** is an open standard that lets AI agents
(Claude, Cursor, VS Code Copilot, Windsurf, etc.) discover and call external
tools over a consistent JSON-RPC interface.

FunctionFly has first-class MCP support: any function can be published to the
**MCP Function Registry** with a single toggle, making it instantly callable
from any MCP-compatible host.

## How It Works

```
┌──────────────┐    MCP (JSON-RPC)    ┌─────────────────────┐    execute    ┌──────────────┐
│  AI Agent    │ ◄──────────────────► │  FunctionFly MCP    │ ───────────► │  Your        │
│  (Claude,    │    streamable-HTTP   │  Function Registry   │              │  Function    │
│   Cursor…)   │    or stdio          │                     │              │              │
└──────────────┘                      └─────────────────────┘              └──────────────┘
```

1. **Discover** — The AI agent calls `tools/list` and receives every MCP-enabled function you have access to, including input schemas.
2. **Call** — The agent invokes `tools/call` with function arguments. The registry authenticates, rate-limits, and routes the call to your function.
3. **Respond** — The function result is returned as structured content the agent can reason about.

## Key Features

- **One-toggle publishing** — Enable MCP on any existing function from the dashboard or API
- **Automatic schema exposure** — Your function's input schema becomes the MCP `inputSchema`
- **Trust & verification** — Verified MCP badge for functions that pass security and quality checks
- **Rate limiting** — Per-function, per-caller rate limits (default 60/min)
- **Multiple transports** — Streamable-HTTP (default) and stdio

## Quick Start

### Install the MCP Server

```bash
npm install -g @functionfly/mcp-cli
export FUNCTIONFLY_API_KEY=ffp_...
flypy-mcp install
```

This auto-configures Claude Desktop, Cursor, VS Code, Windsurf, and Cline.

### Publish a Function

Enable MCP on any function:

```bash
curl -X PATCH https://api.functionfly.com/v1/functions/alice/my-func/mcp \
  -H "Authorization: Bearer $FUNCTIONFLY_API_KEY" \
  -H "Content-Type: application/json" \
  -d '{ "enabled": true }'
```

Or toggle it in **Dashboard → Functions → Settings → MCP-compatible**.

## Next Steps

- [Server Setup](/mcp/server-setup/) — Install the MCP server in your AI client
- [Publish MCP](/mcp/publish-mcp/) — Make a function MCP-compatible
- [API Reference](/mcp/api/) — Full JSON-RPC 2.0 reference
- [Registry Guide](/guides/registry-guide/) — Browsing and using the function registry

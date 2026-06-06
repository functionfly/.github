---
title: MCP Server Setup
description: Install the FunctionFly MCP server in Claude Desktop, Cursor, VS Code, Continue, Windsurf, and Cline.
sidebar:
  label: Server Setup
  order: 1
---

This page walks through installing the **FunctionFly MCP server** in every
major MCP-compatible host. The install is the same single command for all of
them: `npx -y @functionfly/mcp-server` with `FUNCTIONFLY_API_KEY` set.

## Before you start

1. Create a FunctionFly API key (`ffp_…`) at
   [functionfly.com/settings/api-keys](https://functionfly.com/settings/api-keys).
2. Make sure the host (Claude, Cursor, etc.) supports MCP — all major AI
   clients do as of mid-2026.

## Claude Desktop

Edit your `claude_desktop_config.json`:

- **macOS:** `~/Library/Application Support/Claude/claude_desktop_config.json`
- **Windows:** `%APPDATA%\Claude\claude_desktop_config.json`
- **Linux:** `~/.config/Claude/claude_desktop_config.json`

```json
{
  "mcpServers": {
    "functionfly": {
      "command": "npx",
      "args": ["-y", "@functionfly/mcp-server"],
      "env": { "FUNCTIONFLY_API_KEY": "ffp_..." }
    }
  }
}
```

Restart Claude Desktop. You should see a "functionfly" entry in the MCP
tools menu (hammer icon, bottom-right of the chat).

## Cursor

Edit `~/.cursor/mcp.json` (create it if it doesn't exist):

```json
{
  "mcpServers": {
    "functionfly": {
      "command": "npx",
      "args": ["-y", "@functionfly/mcp-server"],
      "env": { "FUNCTIONFLY_API_KEY": "ffp_..." }
    }
  }
}
```

## VS Code (Copilot Chat)

Edit `<VSCode user dir>/mcp.json`:

- **macOS:** `~/Library/Application Support/Code/User/mcp.json`
- **Windows:** `%APPDATA%\Code\User\mcp.json`
- **Linux:** `~/.config/Code/User/mcp.json`

```json
{
  "mcpServers": {
    "functionfly": {
      "command": "npx",
      "args": ["-y", "@functionfly/mcp-server"],
      "env": { "FUNCTIONFLY_API_KEY": "ffp_..." }
    }
  }
}
```

## Continue (VS Code / JetBrains)

Edit `~/.continue/config.json`:

```json
{
  "mcpServers": [
    {
      "name": "functionfly",
      "command": "npx",
      "args": ["-y", "@functionfly/mcp-server"],
      "env": { "FUNCTIONFLY_API_KEY": "ffp_..." }
    }
  ]
}
```

## Windsurf

Edit `~/.codeium/windsurf/mcp_config.json`:

```json
{
  "mcpServers": {
    "functionfly": {
      "command": "npx",
      "args": ["-y", "@functionfly/mcp-server"],
      "env": { "FUNCTIONFLY_API_KEY": "ffp_..." }
    }
  }
}
```

## Cline

Cline (the VS Code extension) uses the VS Code `mcp.json` file directly. See
the VS Code section above.

## One-command install

The `@functionfly/mcp-cli` package automates all of the above. After
installing it globally:

```bash
npm install -g @functionfly/mcp-cli
export FUNCTIONFLY_API_KEY=ffp_...
flypy-mcp install
```

It writes the right config block to every host you have installed.

## Verifying the install

After restarting your host, the function list will populate as soon as the
MCP client calls `tools/list`. The first call authenticates with your
`ffp_…` key and returns the full registry of MCP-enabled functions in your
tenant.

To verify directly:

```bash
curl https://api.functionfly.com/v1/mcp/manifest
```

You should see a JSON document with `protocol_version: "2025-03-26"` and
the FunctionFly server identity.

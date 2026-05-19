---
title: Studio
description: Learn how to use the FunctionFly Studio for building and managing AI agent workflows.
sidebar:
  order: 1
---

import { Tabs } from '@astrojs/starlight/components';

# FunctionFly Studio

Studio is your command center for building, managing, and monitoring AI agent workflows. It provides a visual environment for creating agents, orchestrating tasks, and tracking execution.

## Accessing Studio

Navigate to **Dashboard → Studio** or use the direct URL:

| Environment | URL |
|------------|-----|
| **Production** | `/studio` |
| **Staging** | `/studio/staging` |
| **Development** | `/studio/development` |

Each environment maintains completely isolated data including:
- Agents and configurations
- Workflow graphs and executions
- Tasks and collaboration events
- Marketplace listings and settings

## Environment Tabs

The sidebar includes environment switcher tabs (P/S/D) that let you quickly switch between:

<Tabs>
  <Tabs label="Production">
    **Production** (`/studio`) - Your live environment with real data and integrations.
  </Tabs>
  <Tabs label="Staging">
    **Staging** (`/studio/staging`) - Pre-production testing with replica data.
  </Tabs>
  <Tabs label="Development">
    **Development** (`/studio/development`) - Experimental workspace for new features.
  </Tabs>
</Tabs>

### How Environment Switching Works

1. Click the environment tab in the sidebar (P/S/D)
2. The URL updates to reflect the selected environment
3. The API automatically sends the `X-Environment` header
4. All Studio data (agents, tasks, settings) is isolated per environment

## Studio Layout

### Left Panel Tabs

| Tab | Icon | Description |
|-----|------|-------------|
| **Agents** | Bot | View and manage AI agents |
| **Canvas** | GitBranch | Visual workflow graph editor |
| **Marketplace** | Layers | Browse and manage functions |
| **Runtime** | Cpu | Runtime configuration and logs |
| **Swarm** | Users | Multi-agent swarm coordination |
| **Skills** | Wand2 | Agent skill management |
| **Extensions** | Puzzle | Extension marketplace |

### Bottom Panel Tabs

| Tab | Description |
|-----|-------------|
| **Exec** | Execution history and logs |
| **Sim** | Simulation controls and results |
| **Ghost** | Autonomous build mode |
| **Tasks** | Task tracking and assignment |
| **DevOps** | Deployment and infrastructure |
| **Memory** | Agent memory and context |
| **Robotics** | Robotics integration (Labs) |

### Right Panel Tabs

| Tab | Description |
|-----|-------------|
| **Telemetry** | Real-time metrics and monitoring |
| **3D View** | Digital twin visualization |
| **Profiler** | Performance profiling |
| **Collab** | Collaboration activity feed |

## Key Features

### Agent Management

Create and manage AI agents with custom:
- **Runtime**: WASM, Node.js, Bun, Deno, or custom
- **Model**: OpenAI, Anthropic, local models
- **Tools**: Code execution, function calling, web search
- **Memory**: Short-term, long-term, vector storage

### Workflow Canvas

Visual graph editor for orchestrating multi-agent workflows:

- Drag-and-drop node placement
- Connection lines for data flow
- Execution tracing and debugging
- Real-time collaboration with cursors

### Ghost Mode

Autonomous AI-assisted development:

1. **Plan** - Ghost analyzes requirements
2. **Build** - Creates agents and workflows
3. **Test** - Validates against scenarios
4. **Iterate** - Refines based on feedback

### Marketplace

Access pre-built functions:

- **Deploy** functions to the marketplace
- **Subscribe** to functions from other tenants
- **License** management for commercial functions
- **Royalty** tracking for function creators

## API Integration

All Studio API calls support environment scoping via the `X-Environment` header:

```bash
# Get production settings
curl -H "X-Environment: production" \
     https://api.functionfly.com/v1/studio/settings

# Get staging settings
curl -H "X-Environment: staging" \
     https://api.functionfly.com/v1/studio/settings
```

### Key Endpoints

| Endpoint | Method | Description |
|----------|--------|-------------|
| `/v1/studio/settings` | GET/PUT | User preferences |
| `/v1/studio/tasks` | GET/POST | Task management |
| `/v1/studio/collab/events` | GET/POST | Collaboration events |
| `/v1/studio/telemetry` | GET | Performance metrics |
| `/v1/marketplace/functions` | GET | Browse functions |
| `/v1/extensions` | GET | List extensions |

## Tips

- **Isolate experiments** - Use Development for risky changes
- **Sync settings** - Settings are per-environment, not global
- **Bookmark environments** - Save `/studio/staging` as a bookmark
- **Check the URL** - Always confirm which environment you're in
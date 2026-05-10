---
title: Function Runtime Graph (FRG)
description: Build complex serverless workflows visually with the Function Runtime Graph editor.
sidebar:
  order: 17
---



The Function Runtime Graph (FRG) is a visual workflow editor that lets you design, build, and execute complex serverless workflows by connecting functions as nodes in a directed graph.

## Overview

FRG allows you to:

- **Visual design** — Drag and drop functions to create workflows
- **3D visualization** — View your workflows in stunning 3D
- **AI assistance** — Describe what you want and AI builds it
- **Version control** — Track changes to your graphs over time
- **Real-time execution** — Test workflows directly in the editor

---

## Key Concepts

### Nodes

Each node represents a **function** in your workflow:

| Node Type | Description |
|-----------|-------------|
| **Function Node** | A serverless function from your registry |
| **Input Node** | External input to trigger the workflow |
| **Output Node** | Final result returned |
| **Conditional Node** | Branch logic based on data |

### Edges

Edges connect nodes and define **data flow**:

- **Sync edges** — Pass output to next node (waits for completion)
- **Async edges** — Fire and forget (triggers next node immediately)
- **Error edges** — Route to error handler on failure

### Graph Structure

```
[Input] → [Validate] → [Process] → [Store]
              ↓            ↓
         [On Error]    [Notify]
              ↓
         [Output: Error]
```

---

## Getting Started

### Accessing the FRG Editor

1. Go to **Build → Graph Editor** in the dashboard
2. Or navigate directly to `/frg`
3. Click **New Graph** to start fresh
4. Or select an existing graph to edit

### Creating Your First Graph

1. Click **New Graph**
2. Enter graph name and optional description
3. Start adding nodes from the **Function Library** panel

### Adding Nodes

1. Open the **Function Library** (left panel)
2. Browse or search for functions
3. Drag a function onto the canvas
4. Or double-click to add at cursor position

### Connecting Nodes

1. Hover over a node to see output ports
2. Click and drag from an output port
3. Drop on another node's input port
4. Edge automatically created

### Configuring Nodes

Click on a node to see its **Inspector** panel (right side):

| Setting | Description |
|---------|-------------|
| **Version** | Select function version to use |
| **Timeout** | Maximum execution time |
| **Retry Policy** | Number of retries on failure |
| **Input Mapping** | Transform input data |
| **Output Mapping** | Transform output data |

---

## The Canvas

### Navigation

| Action | Mouse | Keyboard |
|--------|-------|----------|
| Pan | Click + drag | Arrow keys |
| Zoom | Scroll wheel | `+` / `-` |
| Select node | Click | `Enter` |
| Deselect | Click empty | `Escape` |
| Multi-select | `Shift` + click | `Shift` + arrows |
| Delete | `Delete` / `Backspace` | `Delete` |

### Canvas Controls

Bottom toolbar:

| Button | Action |
|--------|--------|
| **Reset View** | Return to default zoom/position |
| **Fit All** | Zoom to fit all nodes |
| **Grid** | Toggle alignment grid |
| **Snap** | Enable/disable node snapping |
| **3D Toggle** | Switch to 3D view |

### 3D View

Click the **3D Toggle** button to enter immersive 3D mode:

- **Rotate** — Click and drag to rotate view
- **Zoom** — Scroll to zoom in/out
- **Pan** — Right-click drag to pan
- **Node highlighting** — Hover for details

---

## AI Assistant

The FRG editor includes an AI assistant to help build workflows.

### Using AI Compose

1. Open **AI Assistant** panel (or press `Cmd/Ctrl + K`)
2. Describe what you want:
   - "Create a workflow that validates input, processes payment, and sends confirmation email"
   - "Add error handling that notifies via Slack on failure"
3. AI generates the graph structure
4. Review and accept

### AI Suggestions

While editing, the AI suggests:
- Missing nodes based on your workflow
- Optimization opportunities
- Error handling improvements
- Performance enhancements

### Example Prompts

| Goal | Prompt |
|------|--------|
| Create approval flow | "Build a workflow that gets user approval before processing payment" |
| Add monitoring | "Add logging and metrics collection to this workflow" |
| Error handling | "Add retry logic and Slack notifications on failure" |
| Refactor | "Optimize this workflow to reduce latency" |

---

## Data Flow

### Input Mapping

Transform input data before it reaches a node:

```javascript
// Input mapping function
(input) => {
    return {
        userId: input.body.user_id,
        amount: parseFloat(input.body.amount),
        currency: input.body.currency || 'USD'
    };
}
```

### Output Mapping

Transform output from a node:

```javascript
// Output mapping function
(output) => {
    return {
        success: output.status === 'success',
        transactionId: output.data.id,
        message: output.message
    };
}
```

### Conditional Routing

Use conditional nodes to branch your workflow:

```javascript
// Condition function
(input) => {
    if (input.amount > 1000) {
        return 'high_value';   // Route to high-value handler
    } else if (input.amount > 100) {
        return 'standard';      // Route to standard handler
    }
    return 'low_value';         // Route to low-value handler
}
```

---

## Testing Workflows

### Test Runner Panel

1. Open **Test Runner** panel (bottom)
2. Enter test input JSON
3. Click **Run** (or press `Cmd/Ctrl + Enter`)

### Viewing Execution

| View | Description |
|------|-------------|
| **Timeline** | See execution order and timing |
| **Data** | Inspect input/output at each node |
| **Errors** | View any errors that occurred |

### Execution States

Nodes show execution state visually:

| State | Visual |
|-------|--------|
| **Pending** | Gray, dotted border |
| **Running** | Blue, pulsing |
| **Success** | Green, solid |
| **Error** | Red, solid |
| **Skipped** | Yellow, strikethrough |

### Debugging

Use breakpoints to pause execution:

1. Right-click on a node
2. Select **Add Breakpoint**
3. Run test again
4. Execution pauses at breakpoint
5. Inspect data and step through

---

## Version Control

### Saving Versions

Graphs auto-save, but you can create named versions:

1. Click **Save** dropdown
2. Select **Save as Version**
3. Enter version name (e.g., `v1.0.0`, `production-ready`)
4. Add optional description

### Viewing History

1. Click **History** button
2. See all versions with timestamps
3. Click any version to preview
4. **Restore** to go back to that version

### Publishing

To make a graph executable:

1. Click **Publish**
2. Choose visibility:
   - **Private** — Only you can use
   - **Organization** — Team members can use
   - **Public** — Anyone can use
3. Set as **Latest** if desired

---

## Deploying Graphs

### Graph Endpoints

Each published graph gets a unique endpoint:

```
POST https://api.functionfly.com/v1/graphs/{author}/{graph-name}/execute
```

### Execution Options

| Option | CLI Flag | Description |
|--------|----------|-------------|
| Async execution | `--async` | Don't wait for completion |
| Specific version | `--version v1.0.0` | Use specific version |
| Input data | `--input '{"key": "value"}'` | JSON input data |

### Monitoring Executions

View recent executions:

1. Go to **Graphs → your-graph → Executions**
2. See status, duration, and results
3. Click any execution for details

---

## CLI Commands

```bash
# List your graphs
ffly graphs list

# Create new graph
ffly graphs create my-workflow

# Open in editor
ffly graphs edit my-workflow

# Publish graph
ffly graphs publish my-workflow

# Execute graph
ffly graphs run my-workflow --input '{"data": "test"}'

# View executions
ffly graphs executions my-workflow

# View version history
ffly graphs history my-workflow

# Restore previous version
ffly graphs restore my-workflow --version v1.0.0
```

---

## Limitations

| Plan | Graphs | Nodes per Graph | Executions/mo |
|------|--------|----------------|--------------|
| Free | 3 | 10 | 1,000 |
| Starter | 10 | 50 | 10,000 |
| Professional | 50 | 200 | 100,000 |
| Enterprise | Unlimited | Unlimited | Unlimited |

---

## Best Practices

1. **Start simple** — Build basic flow first, then add complexity
2. **Use meaningful names** — Name nodes clearly (e.g., "ValidateUser" not "node_1")
3. **Add error handling** — Always include error edges and handlers
4. **Test incrementally** — Test each branch before moving on
5. **Version before major changes** — Save a version before refactoring
6. **Use AI as assistant** — Let AI suggest improvements but review carefully
7. **Monitor performance** — Check execution times for bottlenecks
8. **Document complex flows** — Add descriptions to nodes for team clarity

---

## Troubleshooting

### Graph won't execute

**Common causes:**
- Missing required input
- Node not connected properly
- Invalid node configuration

**Solutions:**
1. Check all nodes have required inputs connected
2. Verify node configurations are complete
3. Run test with verbose logging

### Infinite loops

FRG detects infinite loops and prevents execution:

```
Error: Cycle detected in graph (node_a → node_b → node_c → node_a)
```

To fix, remove one of the circular connections.

### Node not found

If a function used in a graph is deleted:

1. Graph shows **Missing Node** placeholder
2. Either restore the function or remove the node
3. Reconnect adjacent nodes as needed

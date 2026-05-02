---
title: Creating Your First Agent
description: Learn how to create, configure, and deploy your first AI agent on FunctionFly with autonomous capabilities and behavioral policies.
sidebar:
  order: 2
---

This guide walks you through creating your first AI agent on FunctionFly. Agents are autonomous AI entities that can execute tasks, make decisions, and evolve over time based on their behavioral policies.

## What is a FunctionFly Agent?

A FunctionFly Agent is an AI-powered entity that:
- **Executes autonomously** based on configured behavioral policies
- **Uses memory** to retain context across interactions
- **Scales automatically** based on demand
- **Evolves** through built-in learning mechanisms
- **Integrates** with your existing systems via SDK

## Prerequisites

Before you begin, make sure you have:

- A FunctionFly account with agent capabilities enabled
- Some credits in your wallet (agents consume compute resources)
- Basic understanding of API authentication

## Step 1: Create an Agent

Navigate to the Agents section in your dashboard and click **Create Agent**:

```bash
# Or use the API directly
curl -X POST https://api.functionfly.com/v1/agents \
  -H "Authorization: Bearer YOUR_API_KEY" \
  -H "Content-Type: application/json" \
  -d '{
    "name": "my-first-agent",
    "description": "A helpful assistant agent"
  }'
```

### Agent Configuration Options

When creating an agent, you'll configure:

| Option | Description | Required |
|--------|-------------|----------|
| `name` | Unique identifier for your agent | Yes |
| `description` | Human-readable description | No |
| `autonomousEnabled` | Allow agent to take actions without prompting | No |
| `evolutionEnabled` | Allow agent to improve based on feedback | No |

## Step 2: Define Capabilities

Capabilities determine what your agent can do. Configure them based on your use case:

```json
{
  "capabilities": {
    "web_search": true,
    "code_execution": true,
    "file_operations": false,
    "api_calls": true
  }
}
```

### Available Capabilities

| Capability | Description |
|------------|-------------|
| `web_search` | Agent can search the web for information |
| `code_execution` | Agent can run code in sandboxed environments |
| `file_operations` | Agent can read/write files (use carefully!) |
| `api_calls` | Agent can make external API requests |
| `database_queries` | Agent can query databases |

## Step 3: Configure Behavioral Policy

Behavioral policies define safety limits and operational boundaries for your agent:

```json
{
  "behavioral_policy": {
    "max_execution_depth": 5,
    "max_recursion_depth": 3,
    "max_wall_time_seconds": 300,
    "max_memory_mb": 512,
    "allowed_tools": ["web_search", "code_execution"],
    "blocked_topics": ["financial_advice", "medical_diagnosis"]
  }
}
```

### Policy Parameters

| Parameter | Default | Description |
|-----------|---------|-------------|
| `max_execution_depth` | 5 | Maximum nested tool calls |
| `max_recursion_depth` | 3 | Maximum recursive operations |
| `max_wall_time_seconds` | 300 | Maximum execution time |
| `max_memory_mb` | 512 | Maximum memory usage |

## Step 4: Set Up Authentication

Agents use API keys for authentication. Generate a dedicated key for each agent:

```bash
# Generate agent API key
curl -X POST https://api.functionfly.com/v1/agents/YOUR_AGENT_ID/keys \
  -H "Authorization: Bearer YOUR_API_KEY"
```

### SDK Authentication

```javascript
import { FunctionFlyAgent } from '@functionfly/agent-sdk';

const agent = new FunctionFlyAgent({
  agentId: 'your-agent-id',
  apiKey: process.env.FFLY_AGENT_API_KEY
});

// Send a task to your agent
const result = await agent.execute({
  task: 'Research the latest AI developments',
  context: { depth: 'comprehensive' }
});
```

## Step 5: Test Your Agent

Test your agent in the dashboard or via API:

```bash
# Test agent execution
curl -X POST https://api.functionfly.com/v1/agents/YOUR_AGENT_ID/execute \
  -H "Authorization: Bearer YOUR_AGENT_KEY" \
  -H "Content-Type: application/json" \
  -d '{
    "task": "What is the weather in New York?",
    "options": {
      "wait_for_completion": true,
      "timeout_seconds": 30
    }
  }'
```

### Dashboard Testing

The Agent Detail page provides:
- Real-time execution logs
- Memory visualization
- Usage analytics
- Policy violation alerts

## Understanding Agent States

Agents can be in one of these states:

| State | Description |
|-------|-------------|
| `initializing` | Agent is being set up |
| `idle` | Agent is ready but not executing |
| `busy` | Agent is processing a task |
| `error` | Agent encountered an error |
| `stopped` | Agent has been manually stopped |

## Monitoring Agent Usage

Track your agent's performance in the dashboard:

- **Calls This Minute** - Current request rate
- **Concurrent Executions** - Parallel task handling
- **Average Execution Time** - Performance metric
- **Spend Today** - Cost tracking

## Advanced: Agent Memory

Agents can maintain memory across interactions for context:

```javascript
// Store information in agent memory
await agent.memory.store('user_preferences', {
  theme: 'dark',
  language: 'en'
});

// Retrieve from memory
const prefs = await agent.memory.retrieve('user_preferences');
```

### Memory Types

| Type | Use Case |
|------|----------|
| `short_term` | Current session context |
| `long_term` | Persistent user information |
| `episodic` | Historical interactions |
| `semantic` | Learned knowledge |

## Troubleshooting

### Agent Not Responding

1. Check agent status is not `stopped`
2. Verify API key is valid
3. Check wallet has sufficient credits

### Policy Violations

If your agent hits policy limits:
- Review `max_execution_depth` and `max_recursion_depth`
- Increase `max_wall_time_seconds` if tasks are timing out
- Check `allowed_tools` includes required tools

### High Costs

To reduce spending:
- Disable `autonomousEnabled` to require approval for actions
- Set stricter `max_wall_time_seconds`
- Limit `allowed_tools` to only necessary capabilities

## Next Steps

- Learn about [Agent Marketplace](/agents/marketplace/) to discover pre-built agents
- Explore [SDK Integrations](/agents/sdk/) for programmatic access
- Set up [Agent Webhooks](/guides/webhooks/) for event notifications
- Read about [Agent Security](/security/) best practices
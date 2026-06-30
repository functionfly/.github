---
title: Agent SDK
description: Build agents with the FunctionFly SDK
---

# Agent SDK

Build autonomous agents using the FunctionFly SDK.

## Installation

```bash
npm install @functionfly/agent-sdk
```

## Quick Start

```javascript
import { Agent } from '@functionfly/agent-sdk';

const agent = new Agent({
  name: 'my-agent',
  instructions: 'You are a helpful assistant.',
  tools: [myFunctionTool]
});

const response = await agent.run('Hello!');
```

## Available SDKs

| Language | Package | Install |
|----------|---------|---------|
| JavaScript/TypeScript | `@functionfly/agent-sdk` | `npm install` |
| Python | `functionfly-agent` | `pip install` |
| Go | `github.com/functionfly/agent-go` | `go get` |

## Features

- Tool registration and execution
- Conversation history and memory
- Streaming responses
- Error handling and retries

## next steps

- [Creating Your First Agent](/guides/creating-agents/)
- [Agent Memory](/agents/memory/)

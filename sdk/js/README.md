# FunctionFly JavaScript/TypeScript SDKs

Comprehensive SDK packages for integrating FunctionFly functions with AI agents and applications.

## Packages

| Package | Description |
|---------|-------------|
| [`@functionfly/agent`](./packages/agent/README.md) | Core Agent SDK with function discovery, trust-aware routing, and automatic retry |
| [`@functionfly/langchain`](./packages/langchain/README.md) | LangChain tool integration |
| [`@functionfly/autogen`](./packages/autogen/README.md) | Microsoft AutoGen agent integration |
| [`@functionfly/crewai`](./packages/crewai/README.md) | CrewAI toolkit integration |
| [`@functionfly/flypy`](./packages/flypy/README.md) | Core SDK for function compilation and state management |

## Trust System

All agent SDKs include the FunctionFly trust system:

| Trust Level | Score Range | Description |
|-------------|-------------|-------------|
| **Highly Trusted** | ≥90 AND verified | Highest trust, recommended for critical tasks |
| **Verified** | ≥70 | Verified functions, safe for production |
| **Trusted** | ≥50 | Basic trust level |
| **Unverified** | <50 | Low trust, use with caution |

## Installation

```bash
# Install all packages
npm install @functionfly/agent @functionfly/langchain @functionfly/autogen @functionfly/crewai

# Or individual packages
npm install @functionfly/agent
npm install @functionfly/langchain
npm install @functionfly/autogen
npm install @functionfly/crewai
```

## Quick Start

### Using with LangChain

```typescript
import { createToolsFromSearch } from '@functionfly/langchain';
import { ChatOpenAI } from '@langchain/openai';
import { createReactAgent } from '@langchain/agents';

// Create tools with trust filtering
const tools = await createToolsFromSearch({
  apiKey: 'your-api-key',
  category: 'data-processing',
  minTrustScore: 70,
});

// Create agent
const model = new ChatOpenAI({ temperature: 0 });
const agent = createReactAgent(model, tools);
```

### Using with AutoGen

```typescript
import { registerFunctionFlyTools, createFunctionFlyAgent } from '@functionfly/autogen';

// Register tools
const result = await registerFunctionFlyTools({
  apiKey: 'your-api-key',
  minTrustScore: 70,
});

// Create agent
const agent = createFunctionFlyAgent({
  name: 'assistant',
  tools: result.tools,
});
```

### Using with CrewAI

```typescript
import { FunctionFlyToolkit } from '@functionfly/crewai';
import { Agent, Task, Crew } from 'crewai';

// Create toolkit
const toolkit = new FunctionFlyToolkit({
  apiKey: 'your-api-key',
  minTrustScore: 70,
});

await toolkit.initialize({ category: 'analysis' });

// Create agent with toolkit
const researcher = new Agent({
  role: 'Research Analyst',
  goal: 'Analyze data using trusted functions',
  tools: toolkit.getTools(),
});
```

## Core Agent SDK

For custom integrations, use the core [`@functionfly/agent`](./packages/agent/README.md) package directly:

```typescript
import { AgentClient } from '@functionfly/agent';

const client = new AgentClient({
  apiKey: 'your-api-key',
});

// Discover functions
const functions = await client.discoverFunctions({
  category: 'processing',
  minTrustScore: 80,
});

// Execute with automatic retry and fallback
const result = await client.executeWithRetry('function-id', { data: 'test' }, {
  enableFallback: true,
  minTrustScore: 70,
});
```

## API Endpoints Used

The SDKs use the following FunctionFly API endpoints:

| Endpoint | Method | Description |
|----------|--------|-------------|
| `/v1/registry/functions` | GET | List/search functions |
| `/v1/registry/functions/search` | GET | Search functions |
| `/v1/registry/functions/{author}/{name}` | GET | Get function details |
| `/v1/registry/functions/{id}/trust` | GET | Get trust score |
| `/v1/functions/{id}/execute` | POST | Execute function |

## Environment Variables

```bash
# FunctionFly API key
FUNCTIONFLY_API_KEY=your-api-key

# Optional: Custom API base URL
FUNCTIONFLY_BASE_URL=https://api.functionfly.com
```

## Error Handling

All SDKs provide specific error types:

```typescript
import { 
  AgentSDKError,
  FunctionNotFoundError,
  TrustScoreTooLowError,
  ExecutionWithFallbackError 
} from '@functionfly/agent';

try {
  await client.executeWithRetry('function-id', input);
} catch (error) {
  if (error instanceof FunctionNotFoundError) {
    // Handle function not found
  } else if (error instanceof TrustScoreTooLowError) {
    // Handle trust score below minimum
  } else if (error instanceof ExecutionWithFallbackError) {
    // All fallbacks exhausted
    console.log('Attempted functions:', error.attemptedFunctions);
  }
}
```

## Development

```bash
# Install dependencies
bun install

# Build all packages
bun run build

# Run tests
bun test

# Run linting
bun lint
```

## License

MIT

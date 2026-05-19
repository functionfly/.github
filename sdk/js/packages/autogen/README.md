# FunctionFly Integration for AutoGen

Register FunctionFly tools with Microsoft AutoGen agents.

## Installation

```bash
npm install @functionfly/autogen
# or
yarn add @functionfly/autogen
# or
bun add @functionfly/autogen
```

## Quick Start

```typescript
import { registerFunctionFlyTools, createFunctionFlyAgent } from '@functionfly/autogen';

// Register tools
const result = await registerFunctionFlyTools({
  apiKey: 'your-api-key',
  minTrustScore: 70,
  maxTools: 20,
});

console.log(`Registered ${result.count} tools`);
console.log(`Trust distribution:`, result.trustDistribution);

// Create agent with tools
const agent = createFunctionFlyAgent({
  name: 'functionfly_assistant',
  description: 'AI assistant with access to trusted functions',
  tools: result.tools,
});
```

## Tool Filtering

```typescript
import { 
  registerFunctionFlyTools, 
  filterByTrustScore, 
  sortByTrustScore,
  getToolsByTrustLevel 
} from '@functionfly/autogen';

// Register all tools
const result = await registerFunctionFlyTools({
  apiKey: 'your-api-key',
  minTrustScore: 50,  // minimum for registration
});

// Filter to only highly trusted
const highlyTrusted = getToolsByTrustLevel(result.tools, 'highly_trusted');

// Filter by minimum score
const above80 = filterByTrustScore(result.tools, 80);

// Sort by trust score
const sorted = sortByTrustScore(result.tools);
```

## Tool Selection

```typescript
import { registerFunctionFlyTools, selectBestTool } from '@functionfly/autogen';

const result = await registerFunctionFlyTools({
  apiKey: 'your-api-key',
});

// Select best tool for a task
const task = 'process user data with high accuracy';
const bestTool = selectBestTool(result.tools, task);

if (bestTool) {
  console.log(`Selected: ${bestTool.name}`);
  console.log(`Trust: ${bestTool.trustScore}%`);
}
```

## Validation for Production

```typescript
import { registerFunctionFlyTools, validateToolsForProduction } from '@functionfly/autogen';

const result = await registerFunctionFlyTools({
  apiKey: 'your-api-key',
});

// Validate tools
const { valid, invalid } = validateToolsForProduction(
  result.tools,
  requireVerified: true  // only allow verified tools
);

if (invalid.length > 0) {
  console.warn('Some tools failed validation:', invalid.map(t => t.name));
}

// Use only valid tools
const agent = createFunctionFlyAgent({
  name: 'safe_assistant',
  tools: valid,
});
```

## Integration with AutoGen

```typescript
import { registerFunctionFlyTools, createFunctionFlyAgent } from '@functionfly/autogen';

// Register tools
const result = await registerFunctionFlyTools({
  apiKey: 'your-api-key',
  minTrustScore: 70,
});

// Create agent
const ffAgent = createFunctionFlyAgent({
  name: 'functionfly_helper',
  tools: result.tools,
  systemMessage: `You are a helpful assistant with access to FunctionFly functions.
    Always verify the trust level before using a function.
    Prefer highly_trusted functions when available.`,
});

// Use with AutoGen
// const agent = new AutoGenAgent({
//   name: ffAgent.name,
//   system_message: ffAgent.systemMessage,
//   tools: ffAgent.tools.map(t => ({ ... })),
// });
```

## Tool Metadata

Each tool includes metadata:

```typescript
interface AutoGenTool {
  name: string;              // Tool name
  description: string;      // Tool description
  parameters: object;       // JSON Schema for parameters
  verified?: boolean;       // Whether function is verified
  trustScore?: number;      // Trust score (0-100)
}
```

## Trust Distribution

The registration result includes trust distribution:

```typescript
const result = await registerFunctionFlyTools({ ... });

console.log(result.trustDistribution);
// {
//   highlyTrusted: 5,
//   verified: 12,
//   trusted: 8,
//   unverified: 2
// }
```

## License

MIT

# FunctionFly Tool for LangChain

Integrate FunctionFly functions as LangChain tools with trust-aware selection.

## Installation

```bash
npm install @functionfly/langchain
# or
yarn add @functionfly/langchain
# or
bun add @functionfly/langchain
```

## Quick Start

```typescript
import { FunctionFlyTool } from '@functionfly/langchain';

// Create tool from function ID
const tool = new FunctionFlyTool({
  apiKey: 'your-api-key',
  functionId: 'func_xxx',
  minTrustScore: 70,  // optional minimum trust score
});

// Initialize and use
await tool.initialize();
const result = await tool.invoke({ input: { data: 'hello' } });

// Or with structured input
const result2 = await tool.invoke({ 
  input: { param1: 'value1', param2: 'value2' } 
});
```

## Using with LangChain Agents

```typescript
import { FunctionFlyTool, createToolsFromSearch } from '@functionfly/langchain';
import { ChatOpenAI } from '@langchain/openai';
import { AgentExecutor } from 'langchain/agents';

// Create multiple tools
const tools = await createToolsFromSearch({
  apiKey: 'your-api-key',
  category: 'data-processing',
  minTrustScore: 80,
  limit: 10,
});

// Create agent
const model = new ChatOpenAI({ temperature: 0 });
const agent = createReactAgent(model, tools);

// Execute
const executor = new AgentExecutor({ agent, tools });
const result = await executor.invoke({ 
  input: 'Use a function to process this data...' 
});
```

## Tool Metadata

Each tool includes trust metadata:

```typescript
const metadata = tool.getMetadata();
console.log(metadata);
// {
//   name: 'functionfly_author_process',
//   description: 'Process function by author',
//   functionId: 'func_xxx',
//   author: 'author',
//   functionName: 'process',
//   trustScore: 92,
//   trustBadge: 'highly_trusted',
//   verified: true,
//   schema: { ... }
// }
```

## Trust Badges

Tools are assigned trust badges:

- `highly_trusted`: Score ≥ 90 AND verified
- `verified`: Score ≥ 70
- `trusted`: Score ≥ 50
- `unverified`: Score < 50

## Configuration Options

```typescript
interface FunctionFlyToolConfig {
  apiKey: string;
  functionId?: string;
  author?: string;
  name?: string;
  baseUrl?: string;
  minTrustScore?: number;
  preferVerified?: boolean;
  enableFallback?: boolean;
  nameOverride?: string;
  descriptionOverride?: string;
}
```

## Auto-initialization

Tools can be used without explicit initialization - they will auto-initialize on first use:

```typescript
const tool = new FunctionFlyTool({
  apiKey: 'your-api-key',
  functionId: 'func_xxx',
});

// First call triggers initialization
const result = await tool.invoke({ input: { data: 'test' } });
```

## Building Custom Tools

```typescript
import { FunctionFlyTool, buildSchema } from '@functionfly/langchain';

// Create tool with custom configuration
const tool = new FunctionFlyTool({
  apiKey: 'your-api-key',
  functionId: 'func_xxx',
  nameOverride: 'my_custom_tool',
  descriptionOverride: 'My custom description',
  minTrustScore: 70,
  enableFallback: true,
});

await tool.initialize();
```

## License

MIT

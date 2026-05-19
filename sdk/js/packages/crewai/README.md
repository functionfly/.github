# FunctionFly Toolkit for CrewAI

FunctionFly toolkit integration for CrewAI multi-agent systems.

## Installation

```bash
npm install @functionfly/crewai
# or
yarn add @functionfly/crewai
# or
bun add @functionfly/crewai
```

## Quick Start

```typescript
import { FunctionFlyToolkit } from '@functionfly/crewai';

// Create toolkit
const toolkit = new FunctionFlyToolkit({
  apiKey: 'your-api-key',
  minTrustScore: 70,
  toolNamePrefix: 'ff',  // optional prefix for tool names
});

// Initialize with search
await toolkit.initialize({
  category: 'data-processing',
  limit: 10,
});

// Get tools for CrewAI
const tools = toolkit.getTools();
```

## Using with CrewAI

```typescript
import { FunctionFlyToolkit } from '@functionfly/crewai';
import { Agent, Task, Crew } from 'crewai';

// Create and initialize toolkit
const toolkit = new FunctionFlyToolkit({
  apiKey: 'your-api-key',
  minTrustScore: 70,
});

await toolkit.initialize({ category: 'analysis' });

// Create agent with toolkit
const researcher = new Agent({
  role: 'Research Analyst',
  goal: 'Research and analyze data using trusted functions',
  backstory: 'Expert analyst with access to FunctionFly functions',
  tools: toolkit.getTools(),
});

// Create task
const researchTask = new Task({
  description: 'Analyze market data for trends',
  agent: researcher,
  expected_output: 'Detailed analysis report',
});

// Create crew
const crew = new Crew({
  agents: [researcher],
  tasks: [researchTask],
});

// Execute
const result = await crew.kickoff();
```

## Toolkit Configuration

```typescript
interface FunctionFlyCrewAIConfig {
  apiKey: string;
  baseUrl?: string;
  minTrustScore?: number;
  maxTools?: number;
  enableFallback?: boolean;
  toolNamePrefix?: string;  // default: 'functionfly'
}
```

## Initializing Tools

### From Search

```typescript
await toolkit.initialize({
  query?: 'text processing',     // search query
  category?: 'data-processing',  // category filter
  limit?: 20,                     // max tools
});
```

### From Specific Function

```typescript
await toolkit.initializeFromFunction({
  author: 'functionfly',
  name: 'process-data',
});
```

### Manual Tool Addition

```typescript
const toolkit = new FunctionFlyToolkit({ apiKey: 'your-key' });

toolkit.addTool({
  functionId: 'func_xxx',
  author: 'author',
  name: 'my-function',
  description: 'My custom function',
  parameters: { type: 'object', properties: {} },
  trustScore: 85,
  verified: true,
});
```

## Tool Management

```typescript
const toolkit = new FunctionFlyToolkit({ apiKey: 'your-key' });
await toolkit.initialize();

// Get all tools
const allTools = toolkit.getTools();

// Get by trust level
const highlyTrusted = toolkit.getToolsByTrustLevel('highly_trusted');
const verified = toolkit.getToolsByTrustLevel('verified');

// Get by author
const authorTools = toolkit.getToolsByAuthor('functionfly');

// Get most trusted
const topTools = toolkit.getMostTrustedTools(5);

// Get summary
const summary = toolkit.getSummary();
console.log(summary);
// {
//   totalTools: 25,
//   byTrustLevel: { highly_trusted: 5, verified: 12, trusted: 6, unverified: 2 },
//   averageTrustScore: 78.5,
//   verifiedCount: 17
// }
```

## Validation

```typescript
import { FunctionFlyToolkit, validateToolkit } from '@functionfly/crewai';

const toolkit = new FunctionFlyToolkit({
  apiKey: 'your-key',
  minTrustScore: 70,
});

await toolkit.initialize();

// Validate for production
const validation = validateToolkit(toolkit, {
  requireMinimumTrustScore: 70,
  requireAllVerified: false,
});

if (!validation.valid) {
  console.error('Validation issues:', validation.issues);
}

if (validation.warnings.length > 0) {
  console.warn('Warnings:', validation.warnings);
}
```

## Tool Format

Tools are formatted for CrewAI compatibility:

```typescript
interface CrewAITool {
  name: string;          // e.g., 'ff_author_process'
  description: string;   // Human-readable description
  parameters: object;    // JSON Schema
  metadata?: {
    trustScore: number;       // 0-100
    trustLevel: TrustLevel;   // highly_trusted|verified|trusted|unverified
    verified: boolean;
    functionId: string;
    author: string;
  };
}
```

## Trust Levels

- `highly_trusted`: Score ≥ 90 AND verified
- `verified`: Score ≥ 70
- `trusted`: Score ≥ 50
- `unverified`: Score < 50

## Display Helpers

```typescript
import { formatToolsForDisplay } from '@functionfly/crewai';

const tools = toolkit.getTools();
console.log(formatToolsForDisplay(tools));
// Output:
// ff_functionfly_process [verified] 85%
//   Process data function
//
// ff_functionfly_analyze [highly_trusted] 95%
//   Analyze data function
```

## License

MIT

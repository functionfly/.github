# Agent Middleware - Auto-Inject Team Memory

## Overview
The agent middleware automatically injects relevant team memory context into agent prompts and code generation requests. This ensures that agents have access to the team's shared knowledge (decisions, preferences, processes, client context) when generating code or processing requests.

## How It Works

### 1. Middleware Layer
The middleware wraps agent API endpoints and extracts context from the authenticated user:

```go
// Extract tenant, team, user IDs from request context
user := middleware.GetUserFromContext(r)
ctx = context.WithValue(ctx, "tenant_id", user.TenantID)
ctx = context.WithValue(ctx, "team_id", teamID)
ctx = context.WithValue(ctx, "user_id", user.UserID)
```

### 2. Prompt Injection
When an agent makes a generation request, the middleware automatically:
- Searches team memory for relevant context
- Injects formatted context into the prompt
- Logs the injection for debugging

### 3. Result
Original prompt:
```
Generate a function to process webhook events
```

Enhanced prompt with team memory:
```
## Team Knowledge & Context
The following information represents shared team knowledge that should inform your response:

### Decisions
1. **Use TypeScript for all new backend functions** (decision)
   We decided to standardize on TypeScript for type safety
   Confidence: 95% | Last updated: 2026-01-15

2. **Always include retry logic for external API calls** (decision)
   Rationale: Network failures are common, retry with exponential backoff
   Confidence: 90% | Last updated: 2026-02-01

### Preferences
1. **Error handling: Always use structured error responses** (preference)
   Subject: error_format | Value: JSON with code, message, details
   Confidence: 85% | Last updated: 2026-01-20

---

## Original Request
Generate a function to process webhook events
```

## Configuration

### Environment Variables
```bash
# Enable/disable agent prompt injection
export AGENT_TEAM_MEMORY_INJECTION=true

# Default team ID for injection (optional)
export AGENT_DEFAULT_TEAM_ID=team-uuid-here
```

### Route Configuration
```go
// In routes_agent.go - enable for specific endpoints
middlewareFactory := team_memory.NewMiddlewareFactory(promptInjector)
middlewareFactory.EnablePath("/agent/generate").
    EnablePath("/agent/execute").
    EnablePath("/v1/agent/generate")
```

### Manual Context Specification
Users can specify team context via:

1. **Query parameter**: `?team_id=uuid`
2. **Header**: `X-Team-ID: uuid`
3. **Default**: Set via `AGENT_DEFAULT_TEAM_ID`

## Usage Examples

### Automatic Injection (Default)
```bash
# Agent makes request - middleware auto-injects context
curl -X POST https://api.functionfly.local/v1/agent/generate \
  -H "Authorization: Bearer $TOKEN" \
  -H "X-Team-ID: team-123-uuid" \
  -d '{
    "name": "process_webhook",
    "description": "Process Stripe webhook events",
    "prompt": "Include retry logic and structured error handling"
  }'
```

### Programmatic Injection
```go
// Direct use of the injector
injector := team_memory.NewAgentPromptInjector(repo)

req := &generation.GenerationRequest{
    Name:        "process_webhook",
    Description: "Process Stripe webhook events",
    Prompt:      "Include retry logic",
}

injectReq := team_memory.InjectContextRequest{
    TenantID:   tenantID,
    TeamID:     teamID,
    UserID:     userID,
    TaskType:   "code_generation",
    MemoryTypes: []string{"decision", "preference", "process"},
}

err := injector.InjectIntoGenerationRequest(ctx, req, injectReq)
// req.Prompt now includes team memory context
```

## API Reference

### AgentAPIMiddleware
Middleware for automatic context extraction and injection.

```go
middleware := team_memory.NewAgentAPIMiddleware(injector)
middleware.SetEnabled(true)

// Wrap handlers
router.Use(middleware.Wrap)
```

### GenerationMiddleware
Specialized middleware for code generation endpoints.

```go
middleware := team_memory.NewGenerationMiddleware(injector)
handler := middleware.Wrap(agentHandler)
```

### AgentPromptInjector
Direct prompt injection utility.

```go
injector := team_memory.NewAgentPromptInjector(repo)

// Inject into generation request
injector.InjectIntoGenerationRequest(ctx, genReq, injectReq)

// Inject into any prompt
enhancedPrompt, err := injector.InjectIntoPrompt(ctx, basePrompt, injectReq)
```

### ContextRequest
```go
type InjectContextRequest struct {
    TenantID    uuid.UUID  // Required
    TeamID      uuid.UUID  // Optional (extracted from tenant if not provided)
    UserID      uuid.UUID  // Optional
    BasePrompt  string     // Original prompt
    TaskType    string     // "code_generation", "analysis", "review"
    Categories  []string   // Filter memories by category
    MemoryTypes []string   // Filter by type: "decision", "preference", "process", "client_context"
}
```

## Memory Type Selection

The injector automatically selects relevant memory types based on task:

| Task Type | Memory Types | Rationale |
|-----------|--------------|-----------|
| `code_generation` | decision, preference, process | Coding standards, tech decisions, workflows |
| `analysis` | decision, client_context | Domain knowledge, client requirements |
| `review` | decision, preference | Standards, style preferences |
| `debugging` | process, decision | Troubleshooting procedures, known issues |

## Performance

### Caching
- Team context is **not cached** per-request (fresh context each time)
- Memory results are cached by the underlying `AgentContextProvider`
- Cache TTL: 5 minutes for memory search results

### Latency Impact
- Memory search: ~50-200ms (HNSW index)
- Context formatting: ~10ms
- Total overhead: ~100-300ms per generation request

### Cost
- Memory retrieval: No cost (database query)
- Context injection: No additional tokens (prepended to prompt)

## Debugging

### Logs to Watch
```
DEBUG "Agent API middleware: injected team context"
  tenant_id=xxx team_id=xxx user_id=xxx path=/agent/generate

DEBUG "Injected team memory into generation prompt"
  team_id=xxx function=process_webhook context_len=2500 memories_included=3

WARN "Failed to build team context for prompt injection"
  error=... (continues with original prompt)
```

### Headers
The middleware adds these response headers:
- `X-Team-Memory-Injected: true/false`
- `X-Team-Memory-Count: N` (number of memories included)

## Troubleshooting

### No Team Memory Injected
1. Check user has team membership
2. Verify `X-Team-ID` header or `team_id` query param
3. Ensure `AGENT_TEAM_MEMORY_INJECTION=true`
4. Check team has memories (create some via dashboard)

### Too Much Context
1. Adjust `MaxMemories` in `ContextRequest` (default: 5 for generation)
2. Filter by specific `MemoryTypes`
3. Filter by `Categories`

### Wrong Team Context
1. Check `tenant_id` matches user's tenant
2. Verify `team_id` is correct
3. Check team membership permissions

## Future Enhancements

1. **Smart Context Compression**
   - Summarize long memories to fit token limits
   - Priority ranking for relevance

2. **Dynamic Memory Selection**
   - LLM-based memory relevance scoring
   - Context-aware memory filtering

3. **Agent-Specific Memory**
   - Per-agent memory preferences
   - Agent-trained memory weights

4. **Streaming Context**
   - WebSocket-based context updates
   - Real-time memory changes during generation

## Integration Checklist

- [ ] Middleware added to agent routes
- [ ] `team_id` parameter or header supported
- [ ] Default team ID configured (optional)
- [ ] Logging enabled for debugging
- [ ] Memory types relevant to agents populated
- [ ] Performance baseline established

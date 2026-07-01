# Agent Thinking Mode — Design Spec

**Date:** 2026-06-30
**Status:** Draft
**Scope:** Full stack — Python AI service, Go backend, DB migration, Dashboard UI

---

## Problem

Agents currently have no way to leverage provider-native reasoning capabilities. Models like Claude (extended thinking), OpenAI o-series (reasoning tokens), and DeepSeek R1 (reasoning content) support deep chain-of-thought reasoning, but the FunctionFly stack passes no thinking parameters and captures no thinking output. Users can't see *how* an agent arrived at an answer, and complex tasks get the same shallow reasoning as simple ones.

## Goals

1. Enable provider-native thinking/reasoning for agent chat
2. Let users configure thinking mode per agent (off / auto / always)
3. Surface thinking content in the dashboard chat UI
4. Track thinking token usage for cost visibility
5. Graceful degradation — providers without thinking support work unchanged

## Non-Goals

- Streaming thinking tokens in real-time (future enhancement)
- Thinking mode for function generation / DNA mutation (only agent chat)
- Custom thinking prompts or chain-of-thought templates
- Cross-provider thinking content normalization

---

## Data Model

### AgentIdentity additions

```go
// internal/agent/identity/models.go
ThinkingMode   string `json:"thinking_mode" gorm:"not null;default:'off'"`
ThinkingBudget int    `json:"thinking_budget" gorm:"not null;default:10000"`
```

### Thinking modes

| Mode | Behavior |
|------|----------|
| `off` | No thinking parameters sent (default, backward compatible) |
| `auto` | System decides based on query complexity heuristic |
| `always` | Force thinking on every request |

### Auto mode heuristic

Thinking activates when ANY of:
- Message length > 200 characters
- Message contains code blocks (triple backtick)
- Message matches reasoning keywords: `analyze`, `debug`, `explain`, `compare`, `design`, `architect`, `evaluate`, `think`, `reason`, `why`, `how`, `optimize`, `review`, `trace`, `investigate`

### Database migration

```sql
ALTER TABLE agent_identities
  ADD COLUMN IF NOT EXISTS thinking_mode VARCHAR(16) NOT NULL DEFAULT 'off',
  ADD COLUMN IF NOT EXISTS thinking_budget INTEGER NOT NULL DEFAULT 10000;
```

### Chat message storage

Thinking content is stored in the existing `ai_chat_messages.metadata` JSONB column:

```json
{
  "thinking": {
    "content": "...",
    "tokens": 2847
  }
}
```

No schema change needed — `metadata` is already JSONB with default `'{}'`.

---

## Python AI Service Changes

### New types (`ai-service/src/models/schemas.py`)

```python
class ThinkingConfig(BaseModel):
    mode: str = "off"           # off | auto | always
    budget_tokens: int = 10000  # max tokens for thinking
```

### CompletionResponse additions

```python
class CompletionResponse(BaseModel):
    # ... existing fields ...
    thinking_content: Optional[str] = None
    thinking_tokens: int = 0
```

### CompletionRequest addition

```python
class CompletionRequest(BaseModel):
    # ... existing fields ...
    thinking: Optional[ThinkingConfig] = None
```

### Chat message request addition

The `/api/chat/message` endpoint receives thinking config in the `context` dict:

```python
# In the request body:
{
    "context": {
        "thinking_mode": "auto",
        "thinking_budget": 10000
    }
}
```

### Base provider (`base.py`)

Add `thinking` parameter to `complete()` and `stream()`:

```python
async def complete(
    self,
    messages: list[ChatMessage],
    model: Optional[str] = None,
    temperature: float = 0.7,
    max_tokens: Optional[int] = None,
    top_p: Optional[float] = None,
    stop: Optional[list[str]] = None,
    thinking: Optional[ThinkingConfig] = None,  # NEW
) -> CompletionResponse:
```

Default implementation ignores `thinking` — providers that support it override.

### Anthropic provider (`anthropic.py`)

Maps to Anthropic's `extended_thinking` API:

```python
async def complete(self, ..., thinking=None):
    kwargs = {...}
    if thinking and thinking.mode != "off":
        kwargs["thinking"] = {
            "type": "enabled",
            "budget_tokens": thinking.budget_tokens,
        }
        # Temperature must be 1 when thinking is enabled
        kwargs["temperature"] = 1.0

    response = await self._retry_with_backoff(lambda: self.client.messages.create(**kwargs))

    # Extract thinking and text blocks
    thinking_content = ""
    text_content = ""
    thinking_tokens = 0
    for block in response.content:
        if block.type == "thinking":
            thinking_content = block.thinking
        elif block.type == "text":
            text_content = block.text

    return CompletionResponse(
        content=text_content,
        thinking_content=thinking_content or None,
        thinking_tokens=thinking_tokens,
        ...
    )
```

Stream variant uses `self.client.messages.stream()` with thinking enabled and yields thinking/content events separately.

### OpenAI provider (`openai.py`)

Maps to `reasoning_effort` for o-series models (o1, o3, o4):

```python
async def complete(self, ..., thinking=None):
    kwargs = {...}
    model_name = (model or self.model).lower()

    if any(m in model_name for m in ["o1", "o3", "o4"]):
        # Reasoning models: use reasoning_effort instead of temperature
        kwargs.pop("temperature", None)
        if thinking and thinking.mode != "off":
            # Map budget to effort level
            if thinking.budget_tokens >= 20000:
                kwargs["reasoning_effort"] = "high"
            elif thinking.budget_tokens >= 10000:
                kwargs["reasoning_effort"] = "medium"
            else:
                kwargs["reasoning_effort"] = "low"

    response = await self._retry_with_backoff(lambda: self.client.chat.completions.create(**kwargs))

    # OpenAI o-series: reasoning tokens in usage, no thinking content
    reasoning_tokens = getattr(response.usage, 'reasoning_tokens', 0) or 0

    return CompletionResponse(
        content=response.choices[0].message.content or "",
        thinking_tokens=reasoning_tokens,
        ...
    )
```

### DeepInfra provider (`deepinfra.py`)

Extracts `reasoning_content` from DeepSeek R1 responses:

```python
async def complete(self, ..., thinking=None):
    response = await self._retry_with_backoff(...)

    choice = response.choices[0]
    thinking_content = getattr(choice.message, 'reasoning_content', None)

    return CompletionResponse(
        content=choice.message.content or "",
        thinking_content=thinking_content,
        ...
    )
```

### Provider thinking support matrix

| Provider | Thinking API | Thinking Content Available | Notes |
|----------|-------------|---------------------------|-------|
| Anthropic | `extended_thinking` | Yes (thinking blocks) | Requires `temperature=1.0` |
| OpenAI | `reasoning_effort` | No (token counts only) | o-series models only |
| DeepInfra | Native | Yes (`reasoning_content`) | DeepSeek R1 only |
| Groq | None | No | Silently skipped |
| Fireworks | None | No | Silently skipped |
| Together | None | No | Silently skipped |
| OpenRouter | Varies | Depends on model | Passes through |
| MiMo | None | No | Silently skipped |
| StepFun | None | No | Silently skipped |
| MiniMax | None | No | Silently skipped |

---

## Go Backend Changes

### BYOK client (`internal/agent/chat/byok_client.go`)

New types:

```go
type ThinkingConfig struct {
    Mode         string `json:"mode"`
    BudgetTokens int    `json:"budget_tokens"`
}

type ThinkingContent struct {
    Content string `json:"content"`
    Tokens  int    `json:"tokens"`
}

type BYOKRequest struct {
    // ... existing fields ...
    Thinking *ThinkingConfig `json:"thinking,omitempty"`
}

type BYOKResponse struct {
    Content         string           `json:"content"`
    Model           string           `json:"model"`
    ThinkingContent *ThinkingContent `json:"thinking_content,omitempty"`
}
```

**callAnthropic** — parse thinking blocks from response content array:

```go
var anthropicResp struct {
    Model   string `json:"model"`
    Content []struct {
        Type     string `json:"type"`
        Thinking string `json:"thinking,omitempty"`
        Text     string `json:"text,omitempty"`
    } `json:"content"`
    Usage struct {
        InputTokens  int `json:"input_tokens"`
        OutputTokens int `json:"output_tokens"`
    } `json:"usage"`
}

// Iterate content blocks, separate thinking from text
// Pass extended_thinking in request payload when Thinking != nil
```

**callOpenAICompatible** — extract reasoning tokens and reasoning_content:

```go
var openAIResp struct {
    Choices []struct {
        Message struct {
            Content          string `json:"content"`
            ReasoningContent string `json:"reasoning_content,omitempty"`
        } `json:"message"`
    } `json:"choices"`
    Usage struct {
        PromptTokens     int `json:"prompt_tokens"`
        CompletionTokens int `json:"completion_tokens"`
        TotalTokens      int `json:"total_tokens"`
        ReasoningTokens  int `json:"reasoning_tokens,omitempty"`
    } `json:"usage"`
}
```

### Agent chat handler (`internal/api/handlers/agent/agent_chat.go`)

**agentChatResponse** — add thinking field:

```go
type agentChatResponse struct {
    OK       bool                  `json:"ok"`
    Message  string                `json:"message"`
    Model    string                `json:"model"`
    Thinking *chat.ThinkingContent `json:"thinking,omitempty"`
}
```

**Auto mode heuristic** — function to determine if thinking should activate:

```go
func shouldThink(message, mode string) bool {
    if mode == "always" { return true }
    if mode == "off" || mode == "" { return false }
    // auto mode
    if len(message) > 200 { return true }
    if strings.Contains(message, "```") { return true }
    keywords := []string{"analyze", "debug", "explain", "compare", "design",
        "architect", "evaluate", "think", "reason", "why", "how",
        "optimize", "review", "trace", "investigate"}
    lower := strings.ToLower(message)
    for _, kw := range keywords {
        if strings.Contains(lower, kw) { return true }
    }
    return false
}
```

**HandleAgentChat** — build thinking config from agent, pass to BYOK:

```go
var thinking *chat.ThinkingConfig
if shouldThink(req.Message, agent.ThinkingMode) {
    thinking = &chat.ThinkingConfig{
        Mode:         agent.ThinkingMode,
        BudgetTokens: agent.ThinkingBudget,
    }
}

resp, err := chat.CallLLM(ctx, chat.BYOKRequest{
    // ... existing fields ...
    Thinking: thinking,
})

// Include thinking in response and persist to metadata
writeJSON(w, http.StatusOK, agentChatResponse{
    OK:       true,
    Message:  reply,
    Model:    model,
    Thinking: resp.ThinkingContent,
})
```

**FlyMind fallback path** — pass thinking config in the context:

```go
aiReq := map[string]interface{}{
    // ... existing fields ...
    "context": map[string]string{
        "agent_id":        agent.AgentID,
        "agent_name":      agent.Name,
        "system_prompt":   systemPrompt,
        "thinking_mode":   agent.ThinkingMode,
        "thinking_budget": strconv.Itoa(agent.ThinkingBudget),
    },
}
```

### Chat history endpoint

Include thinking metadata in history responses:

```go
type chatMsg struct {
    Role      string                 `json:"role"`
    Content   string                 `json:"content"`
    Model     string                 `json:"model,omitempty"`
    Metadata  map[string]interface{} `json:"metadata,omitempty"`
    CreatedAt time.Time              `json:"created_at"`
}
```

### Agent update endpoint

Accept `thinking_mode` and `thinking_budget` in agent update requests.

---

## Dashboard UI Changes

### Agent settings page

Add thinking mode controls alongside the existing model selector:

- **Thinking Mode**: Segmented control with three options: Off / Auto / Always
- **Thinking Budget**: Slider (1,000 – 50,000 tokens, step 1,000) — visible only when mode ≠ `off`
- Helper text: "Auto mode enables thinking for complex queries. Budget controls maximum thinking tokens per request."

### Chat interface — thinking content display

When a response includes thinking content:

- Collapsible panel above the assistant's message
- Header: brain icon + "Thinking" + token count badge (e.g., "2,847 tokens")
- Content: monospace font, slightly dimmer than response text, left accent border
- Max height 300px with scroll
- Animated expand/collapse transition
- Default state: collapsed

### API client types

```typescript
interface ThinkingContent {
  content: string;
  tokens: number;
}

interface AgentChatResponse {
  ok: boolean;
  message: string;
  model: string;
  thinking?: ThinkingContent;
}

interface AgentUpdateRequest {
  // ... existing fields ...
  thinking_mode?: "off" | "auto" | "always";
  thinking_budget?: number;
}
```

---

## API Contract

### POST /v1/agent/{agent_id}/chat

**Response (with thinking):**

```json
{
  "ok": true,
  "message": "The answer is...",
  "model": "claude-sonnet-4-6",
  "thinking": {
    "content": "Let me analyze this step by step...",
    "tokens": 2847
  }
}
```

**Response (without thinking):**

```json
{
  "ok": true,
  "message": "The answer is...",
  "model": "gpt-4o-mini"
}
```

### GET /v1/agent/{agent_id}/chat/history

**Response (message with thinking metadata):**

```json
{
  "ok": true,
  "messages": [
    {
      "role": "assistant",
      "content": "The answer is...",
      "model": "claude-sonnet-4-6",
      "metadata": {
        "thinking": {
          "content": "Let me analyze...",
          "tokens": 2847
        }
      },
      "created_at": "2026-06-30T21:00:00Z"
    }
  ]
}
```

### PATCH /v1/agent/{agent_id}

**Request (thinking config update):**

```json
{
  "thinking_mode": "auto",
  "thinking_budget": 15000
}
```

---

## Files to Modify

### Python (ai-service/)
| File | Change |
|------|--------|
| `src/models/schemas.py` | Add `ThinkingConfig`, extend `CompletionRequest`/`CompletionResponse` |
| `src/providers/base.py` | Add `thinking` param to `complete()` and `stream()` |
| `src/providers/anthropic.py` | Implement `extended_thinking` support |
| `src/providers/openai.py` | Implement `reasoning_effort` for o-series |
| `src/providers/deepinfra.py` | Extract `reasoning_content` from DeepSeek R1 |
| `src/api/routes_chat.py` | Pass thinking config from context to providers |
| `src/services/chat/manager.py` | Thread thinking config through chat flow |

### Go (internal/)
| File | Change |
|------|--------|
| `agent/identity/models.go` | Add `ThinkingMode`, `ThinkingBudget` fields |
| `agent/chat/byok_client.go` | Add `ThinkingConfig`, `ThinkingContent` types; parse thinking from responses |
| `api/handlers/agent/agent_chat.go` | Build thinking config, pass to BYOK/FlyMind, include in response |
| `api/handlers/agent/handler.go` | Accept thinking fields in agent update |

### Database
| File | Change |
|------|--------|
| `migrations/YYYYMMDDHHMMSS_add_agent_thinking_mode.up.sql` | Add columns |
| `migrations/YYYYMMDDHHMMSS_add_agent_thinking_mode.down.sql` | Remove columns |

### Dashboard (web/dashboard/)
| File | Change |
|------|--------|
| `src/api/agent.ts` | Add thinking types to chat/update interfaces |
| `src/components/agent/AgentSettings.tsx` | Thinking mode selector + budget slider |
| `src/components/agent/ChatThinking.tsx` | New: collapsible thinking display component |
| `src/components/agent/AgentChat.tsx` | Integrate thinking display in chat |

---

## Cost Implications

- Thinking tokens are billed as output tokens by all providers
- Anthropic: thinking tokens priced same as output ($15/M for Sonnet 4.6)
- OpenAI: reasoning tokens priced same as output
- DeepInfra: reasoning_content tokens included in output count
- Users see thinking token count in the UI to understand cost impact
- Budget cap prevents runaway thinking costs

## Security Considerations

- Thinking content may contain sensitive reasoning — stored only in existing metadata JSONB
- No additional auth required — thinking follows existing agent chat auth
- Thinking budget is capped by the agent's `max_cost_per_execution` quota

## Testing Strategy

- Unit tests for `shouldThink()` heuristic
- Unit tests for Anthropic/OpenAI/DeepInfra thinking parameter construction
- Unit tests for thinking content parsing from responses
- Integration test: BYOK call with thinking enabled, verify response includes thinking
- Integration test: agent chat with thinking_mode="auto", verify thinking activates for complex messages
- Dashboard: manual test thinking panel expand/collapse
- E2E: agent settings → enable thinking → send message → see thinking in chat

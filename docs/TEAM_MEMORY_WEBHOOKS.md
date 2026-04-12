# Conversation Webhook Integration

## Overview
Team Memory Engine is now integrated with conversation lifecycle events. When conversations are resolved, the system automatically triggers memory extraction via webhooks.

## How It Works

### 1. Conversation Resolution Trigger
When a user resolves a conversation via `POST /api/v1/conversations/{id}/resolve`:

```go
// In handler.go - ResolveConversation
if h.memoryEvents != nil && c != nil {
    go func(conv *conversations.Conversation) {
        ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
        defer cancel()
        if err := h.memoryEvents.PublishResolved(ctx, conv); err != nil {
            h.logger.WithError(err).Warn("Failed to publish conversation resolved event")
        }
    }(c)
}
```

### 2. Event Publishing
The `DefaultEventPublisher` distributes the event to all registered handlers:

```go
// In conversation_events.go
func (p *SimpleEventPublisher) PublishResolved(ctx context.Context, conv *conversations.Conversation) error {
    for _, handler := range p.handlers {
        if err := handler.OnConversationResolved(ctx, conv); err != nil {
            logrus.WithError(err).Warn("Event handler failed")
        }
    }
    return nil
}
```

### 3. Memory Extraction
The `ConversationEventHandler` triggers the AutoUpdater:

```go
func (h *ConversationEventHandler) OnConversationResolved(ctx context.Context, conv *conversations.Conversation) error {
    // Only process team conversations
    if conv.OrganizationID == nil {
        return nil
    }
    
    teamID := *conv.OrganizationID
    
    // Queue for memory extraction
    return h.autoUpdater.ProcessConversation(ctx, conv.ID, teamID)
}
```

### 4. Async Processing
The AutoUpdater processes in a goroutine with timeout:

```go
// ProcessConversation runs asynchronously
func (u *AutoUpdater) ProcessConversation(ctx context.Context, conversationID, teamID uuid.UUID) error {
    // 1. Fetch conversation messages
    // 2. Build transcript
    // 3. Call AI service (gpt-4o-mini) for extraction
    // 4. Create extraction records
    // 5. Auto-approve if confidence >= 0.9
}
```

## Architecture

```
Conversation Resolved (API)
    ↓
Handler.ResolveConversation()
    ↓ (async goroutine)
ConversationEventPublisher.PublishResolved()
    ↓
ConversationEventHandler.OnConversationResolved()
    ↓
AutoUpdater.ProcessConversation()
    ↓
AIServiceClient.AnalyzeConversation() → FlyMind AI Service
    ↓
Create MemoryExtractions (DB)
    ↓
Auto-approve high-confidence (≥0.9)
    ↓
Team Memory Created
```

## Configuration

### Environment Variables
```bash
# Enable/disable conversation webhooks
export TEAM_MEMORY_WEBHOOKS_ENABLED=true

# AI Service connection
export AI_SERVICE_URL="http://localhost:8081"
export AI_SERVICE_API_KEY="your-api-key"
```

### Code Configuration
```go
// In routes_agent.go - during server initialization
if s.postgresDB != nil {
    // Create the auto-updater
    autoUpdater := team_memory.NewAutoUpdater(s.repo, conversationRepo, nil)
    
    // Register team memory handler
    team_memory.RegisterTeamMemoryHandler(autoUpdater)
    
    // Wire up to conversation handler
    convHandler.SetMemoryEventPublisher(team_memory.DefaultEventPublisher)
}
```

## Features

### Automatic Extraction
- Triggers on conversation resolution
- Extracts decisions, preferences, processes, client context
- Uses GPT-4o-mini (2026 pricing: ~$0.0006 per conversation)
- Runs asynchronously (doesn't block API response)

### Confidence-Based Approval
- ≥0.9 confidence: Auto-approved (immediate availability)
- 0.7-0.89 confidence: Queued for manual review
- <0.7 confidence: Discarded

### Fallback Strategy
- If FlyMind AI service unavailable → Uses rule-based fallback
- Fallback uses keyword/regex pattern matching
- Zero cost but lower accuracy

### Rate Limiting
- Built-in rate limiting prevents abuse
- Per-conversation extraction tracking
- Configurable time windows

## API Endpoints Affected

| Endpoint | Method | Triggers Webhook |
|----------|--------|-----------------|
| `/conversations/{id}/resolve` | POST | ✅ Yes |

## Database Schema

### Memory Extractions Table
```sql
CREATE TABLE memory_extractions (
    id UUID PRIMARY KEY,
    team_id UUID NOT NULL REFERENCES teams(id),
    conversation_id UUID NOT NULL REFERENCES conversations(id),
    memory_type VARCHAR(50) NOT NULL,
    summary TEXT NOT NULL,
    content JSONB NOT NULL,
    confidence FLOAT NOT NULL,
    status VARCHAR(20) DEFAULT 'pending', -- pending, approved, rejected, auto_applied
    created_at TIMESTAMP DEFAULT NOW()
);
```

## Monitoring

### Logs to Watch
```
# Successful extraction
INFO "Successfully processed conversation for memory extraction"

# Failed extraction  
ERROR "Failed to queue conversation for memory extraction"
WARN "Failed to publish conversation resolved event"

# AI service fallback
WARN "AI service failed, using fallback extractor"
```

### Metrics
- `team_memory_extractions_total` - Total extractions attempted
- `team_memory_extractions_success` - Successful extractions
- `team_memory_extraction_latency_ms` - Processing latency
- `team_memory_ai_service_fallbacks` - Number of fallback activations

## Cost Impact (2026 Pricing)

| Metric | Value |
|--------|-------|
| Per-conversation extraction | ~$0.0006 |
| 2,500 conversations/month | ~$1.50 |
| 10,000 conversations/month | ~$6.00 |
| Model | GPT-4o-mini |

## Future Enhancements

1. **More Triggers**
   - Conversation updated (summary generated)
   - Time-based (N days of inactivity)
   - Manual trigger (admin endpoint)

2. **Smart Filtering**
   - Only extract from org/team conversations
   - Skip DMs (privacy)
   - Configurable minimum message count

3. **Real-time Webhooks**
   - WebSocket notifications for new extractions
   - Slack/Discord integration for review queue

4. **Analytics**
   - Extraction quality metrics
   - User approval rates by type
   - Cost per team/organization

## Testing

### Manual Test
```bash
# 1. Resolve a conversation
curl -X POST https://api.functionfly.local/v1/conversations/{id}/resolve \
  -H "Authorization: Bearer $TOKEN" \
  -d '{"message_id": "optional-solution-message-id"}'

# 2. Check extraction queue
curl https://api.functionfly.local/v1/teams/{teamId}/memories/extractions \
  -H "Authorization: Bearer $TOKEN"

# 3. Review and approve
```

### Integration Test
```go
func TestConversationResolutionWebhook(t *testing.T) {
    // Create test conversation
    conv := createTestConversation()
    
    // Resolve it
    handler.ResolveConversation(w, req)
    
    // Verify extraction created
    extractions, _ := repo.GetMemoryExtractionsByTeam(teamID, "pending", 10)
    assert.Greater(t, len(extractions), 0)
}
```

## Troubleshooting

### Webhook Not Firing
1. Check `TEAM_MEMORY_WEBHOOKS_ENABLED` is set
2. Verify conversation has `organization_id` set
3. Check logs for "Triggering memory extraction"

### AI Service Failures
1. Check `AI_SERVICE_URL` is correct
2. Verify AI service health: `GET /health`
3. Check for rate limiting (429 responses)

### No Extractions Created
1. Check conversation has enough messages (>100 chars transcript)
2. Verify AI extraction confidence (may be below 0.7)
3. Check for duplicate detection (similar memories may be merged)

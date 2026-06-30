---
title: Agent Security
description: Security best practices for agents
---

# Agent Security

Protect your agents and data with these security measures.

## Authentication

All agent requests require authentication:

```bash
# Include API key
curl -H "Authorization: Bearer $FFLY_API_KEY" \
  https://api.functionfly.com/v1/agents/execute
```

## Rate Limiting

Configure per-agent rate limits:

| Plan | Requests/min |
|------|-------------|
| Free | 10 |
| Pro | 100 |
| Enterprise | Unlimited |

## Behavioral Policies

Set rules to control agent behavior:

- **Allowed tools** - Restrict which functions agents can call
- **Output filtering** - Block sensitive data in responses
- **Audit logging** - Track all agent actions

## Data Privacy

- Agents cannot access user data without explicit permission
- Conversation history is encrypted at rest
- You can delete agent memory at any time

## best practices

1. Use least-privilege for agent permissions
2. Review behavioral policy settings regularly
3. Monitor agent logs for suspicious activity
4. Keep API keys secure and rotate periodically

## next steps

- [Behavioral Policies](/agents/policies/)
- [Agent SDK](/agents/sdk/)

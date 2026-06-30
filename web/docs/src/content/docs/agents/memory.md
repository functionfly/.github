---
title: Agent Memory
description: Persistent memory for agents
---

# Agent Memory

Agents can maintain different types of memory for persistent context.

## Memory Types

### Short-term Memory
Conversation context within a session:

```javascript
agent.addMessage({ role: 'user', content: 'Hello' });
agent.addMessage({ role: 'assistant', content: 'Hi!' });
```

### Long-term Memory
Persistent information across sessions:

```javascript
await agent.memory.save({
  type: 'preference',
  key: 'user-name',
  value: 'Alice'
});
```

### Semantic Memory
Learned facts and patterns:

```javascript
await agent.memory.remember('user-prefers-dark-mode');
```

## Memory Configuration

| Type | Retention | Configurable |
|------|-----------|-------------|
| Short-term | Session | No |
| Long-term | 30 days | Yes |
| Semantic | Until deleted | Yes |

## Privacy

- Users can view what memory exists
- Delete individual memories or all memory
- Memory is encrypted at rest

## best practices

1. Be explicit about what to remember
2. Respect user privacy preferences
3. Clean up unneeded memories regularly
4. Test memory retrieval accuracy

## next steps

- [Agent SDK](/agents/sdk/)
- [Creating Your First Agent](/guides/creating-agents/)

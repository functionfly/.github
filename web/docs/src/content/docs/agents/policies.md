---
title: Behavioral Policies
description: Configure agent behavior rules
---

# Behavioral Policies

Behavioral policies define how your agents operate within defined safety and business rules.

## Policy Types

### Allowlist
Specify what agents are permitted to do:

- List of allowed function calls
- Permitted data sources
- Approved external APIs

### Blocklist
Block specific actions or content:

- Prohibited phrases or topics
- Restricted function calls
- Sensitive data patterns

### Moderation
Content filtering and review:

- Profanity filtering
- PII detection
- Custom regex rules

## Creating a Policy

```javascript
const policy = {
  name: 'strict-customer-support',
  rules: [
    { type: 'allowlist', actions: ['search', 'respond'] },
    { type: 'blocklist', patterns: ['ssn', 'credit-card'] },
    { type: 'moderation', level: 'high' }
  ]
};
```

## Testing Policies

Test policies in sandbox mode before enforcing:

```bash
ff policy test --policy-id policy_xxx --input "sample input"
```

## best practices

1. Start permissive, then restrict
2. Test thoroughly with edge cases
3. Monitor and iterate on policy rules
4. Document policy rationale

## next steps

- [Agent Security](/agents/security/)
- [Agent Memory](/agents/memory/)

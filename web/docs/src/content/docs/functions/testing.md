---
title: Testing Functions
description: Local development, the Playground, and debugging
sidebar:
  order: 4
---


FunctionFly provides multiple ways to test your functions before and after deployment.

## Local Development

Start a local dev server that hot-reloads on file changes:

```bash
ff dev
```

This starts a local server (default `http://localhost:8080`) that emulates the FunctionFly runtime.

### Test with curl

```bash
curl http://localhost:8080 \
  -X POST \
  -H "Content-Type: application/json" \
  -d '{"name": "FunctionFly"}'
```

### Test with the CLI

```bash
ff invoke my-function --data '{"name": "FunctionFly"}' --local
```

## Function Playground

The [Function Playground](/guides/playground/) is an interactive testing environment in the dashboard:

- **Execute with custom inputs** — Edit request body, headers, and params
- **View streaming responses** — See output as it's generated
- **Compare versions** — Diff outputs across function versions
- **Execution history** — Browse past invocations with full request/response details
- **Debug logs** — View console output and error traces

Access it at **Dashboard → Functions → Playground**.

## Unit Testing

### Python

```python
import pytest
from main import handler

@pytest.mark.asyncio
async def test_handler():
    request = {
        "method": "POST",
        "body": {"name": "Test"},
        "headers": {},
        "params": {},
        "url": "http://localhost"
    }
    result = await handler(request)
    assert result["status"] == 200
    assert "Hello, Test" in str(result["body"])
```

### JavaScript

```javascript
import { describe, it, expect } from 'vitest';
import handler from './index.js';

describe('handler', () => {
  it('returns greeting', async () => {
    const request = {
      method: 'POST',
      body: { name: 'Test' },
      headers: {},
      params: {},
      url: 'http://localhost'
    };
    const result = await handler(request);
    expect(result.status).toBe(200);
    expect(result.body.message).toContain('Hello, Test');
  });
});
```

## Integration Testing

Use `ff invoke` with `--env` to test against a deployed environment:

```bash
ff invoke my-function --data '{"test": true}' --env staging

ff invoke my-function --data '{"test": true}' --env production
```

## Debugging

### View Logs

```bash
ff logs my-function

ff logs my-function --level error

ff logs my-function --tail 50
```

### Execution Traces

In the dashboard, each invocation shows:

- **Request/response** — Full payload and headers
- **Duration** — Total time and breakdown (init, execution, network)
- **Memory usage** — Peak memory during execution
- **Logs** — All `console.log` / `print` output
- **Errors** — Stack traces with source mapping

## Next Steps

- [Best Practices](/functions/best-practices/) — Performance and reliability tips
- [Monitoring & Observability](/guides/monitoring/) — Production monitoring
- [Error Codes](/guides/error-codes/) — Error reference
- [CI/CD Integration](/guides/ci-cd/) — Automated testing in pipelines

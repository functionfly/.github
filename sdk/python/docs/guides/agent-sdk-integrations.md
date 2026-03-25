# Agent SDK Integrations Quickstart

This guide shows how to integrate FunctionFly trusted tools with LangChain, AutoGen, and CrewAI using the Python SDK.

## 1) Install

Install only the framework adapter(s) you need:

```bash
pip install "flypy[agents-langchain]"
pip install "flypy[agents-autogen]"
pip install "flypy[agents-crewai]"
```

Or install all adapter dependencies:

```bash
pip install "flypy[agents]"
```

## 2) Configure Environment

```bash
export FUNCTIONFLY_API_BASE="https://api.functionfly.com"
export FUNCTIONFLY_API_KEY="<your_api_key>"
```

Notes:
- `AgentClient` enforces HTTPS for non-local hosts.
- The trust policy defaults to fail-closed:
  - `require_verified=True`
  - `min_trust_score=80`

## 3) Build Trusted Tools

```python
from flypy import AgentClient, LangChainAdapter, TrustPolicy

client = AgentClient()
adapter = LangChainAdapter(client)
policy = TrustPolicy(required_trust_levels=["high"])

tools = adapter.build_tools(
    policy=policy,
    query="text transform",
    limit=10,
)
```

## 4) Execute a Tool

Every execution response contains metadata for auditability:
- `tool_id`
- `author`, `name`, `version`
- `policy_hash` (deterministic hash of normalized trust policy)

```python
tool = tools[0]
result = tool["callable"](text="hello world")
print(result["metadata"]["policy_hash"])
```

## Production Guardrails

- Keep API keys in environment variables only.
- Set `timeout_seconds` and `max_retries` explicitly for your environment.
- Keep `require_verified=True` for production.
- Use `capabilities_allow` / `capabilities_deny` to constrain tool behavior.

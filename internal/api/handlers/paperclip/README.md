# Paperclip webhook adapter

This handler lets Paperclip heartbeat invocations trigger FunctionFly agent executions and optionally post results back as Paperclip issue comments.

## Endpoint

- **POST /v1/integrations/paperclip/heartbeat**

## Request

- **Auth**: Set `X-Paperclip-Webhook-Secret` header to the value of `PAPERCLIP_WEBHOOK_SECRET` (if set).
- **Body** (JSON):
  - `paperclip_agent_id` (string): Paperclip agent ID (for comment attribution).
  - `paperclip_issue_id` (string): Paperclip issue ID (to post result as comment).
  - `company_id` (string): Paperclip company ID (optional, for logging).
  - `function_author` (string): FunctionFly function author (required).
  - `function_name` (string): FunctionFly function name (required).
  - `function_version` (string): Optional version; omit for latest.
  - `input` (object): Optional JSON input for the function; default `{}`.

## Environment

| Variable | Description |
|----------|-------------|
| `PAPERCLIP_WEBHOOK_SECRET` | If set, requests must include `X-Paperclip-Webhook-Secret` with this value. |
| `PAPERCLIP_BASE_URL` | Paperclip API base URL (e.g. `http://localhost:3100`) for posting comments. |
| `PAPERCLIP_API_KEY` | Paperclip API key or Bearer token for posting issue comments. |
| `FUNCTIONFLY_BASE_URL` | FunctionFly orchestrator base URL (e.g. `http://localhost:8080`). Default: `http://localhost:8080`. |
| `FUNCTIONFLY_AGENT_API_KEY` | FunctionFly agent API key used to call `/v1/agent/execute/...`. |

## Flow

1. Paperclip (or an HTTP adapter) POSTs to this endpoint with the task and target function.
2. The adapter calls `POST {FUNCTIONFLY_BASE_URL}/v1/agent/execute/{author}/{name}[/{version}]` with `X-Agent-API-Key` and `input`.
3. If `PAPERCLIP_BASE_URL`, `PAPERCLIP_API_KEY`, and `paperclip_issue_id` are set, the adapter POSTs the execution result to `POST {PAPERCLIP_BASE_URL}/api/issues/{issueId}/comments`.

## Configuring Paperclip HTTP adapter

In Paperclip, when creating or editing an agent with adapter type `http`, set the callback URL to:

`https://your-orchestrator/v1/integrations/paperclip/heartbeat`

and send the webhook payload above. Include `X-Paperclip-Webhook-Secret` if you set `PAPERCLIP_WEBHOOK_SECRET`.

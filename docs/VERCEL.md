# Vercel Deployment Adapter

This document describes how to use the **Vercel deployment adapter** (provider `vercel`) to deploy functions or apps from the FunctionFly orchestrator or CLI to Vercel.

## Overview

The adapter uses the [Vercel REST API](https://vercel.com/docs/rest-api) to create deployments, set environment variables, bind domains, and manage projects. Deployments are created by uploading a function bundle (e.g. serverless function code) as multipart form data.

## Configuration

### Environment (optional)

| Variable | Default | Description |
|----------|---------|-------------|
| `VERCEL_API_BASE` | `https://api.vercel.com` | API base URL. Override for testing or self-hosted proxies. |

### Provider config (per deployment or backend)

| Key | Required | Description |
|-----|----------|-------------|
| `api_token` | Yes | Vercel API token (create in [Vercel Account → Tokens](https://vercel.com/account/tokens)). |
| `project_name` | Yes | Vercel project name (or use `app_name` in the deployment spec). |
| `team_id` | No | Team ID for Vercel teams; omit for personal account. |
| `project_id` | No | Vercel project ID (for SetEnv/BindRoutes when project is already known). |

## Required inputs for Deploy

- **api_token**: Always required.
- **project_name** (or spec `AppName`): Required. No default.
- **Artifact**: Non-empty bundle (e.g. serverless function code). Empty artifact returns a failed result with a clear message.

## Rate limiting

The client throttles requests (200 ms between calls) to avoid hitting Vercel’s rate limits. No extra configuration is needed.

## Rollback

Rollback is implemented as a **redeploy** of the previous artifact: the adapter calls the same deploy API with the rollback artifact and env. Ensure the deployment spec for rollback includes the previous artifact and env vars.

## Extended adapter features

- **LinkProject** / **GetLinkedProject**: Link a FunctionFly app to a Vercel project (or create the project if it doesn’t exist) and query the linked project.
- **SetEnv** / **BindRoutes**: Require `api_token` and `project_id` in provider config.

## Testing

Run the adapter and client tests:

```bash
go test ./internal/adapters/vercel/... -v
```

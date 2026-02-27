# API test commands

Quick copy-paste commands to test the orchestrator API. Default base: `http://localhost:8080`. Use the same admin email/password you set with `create-admin`.

## One-liners

**Health**
```bash
curl -s http://localhost:8080/health
```

**Login** (change email/password if you created admin with different credentials)
```bash
curl -s -X POST http://localhost:8080/v1/auth/login \
  -H "Content-Type: application/json" \
  -d '{"email":"admin@functionfly.local","password":"admin123"}'
```

**Get session** (paste the token from login response)
```bash
curl -s http://localhost:8080/v1/auth/session \
  -H "Authorization: Bearer YOUR_TOKEN_HERE"
```

## Scripts

| Command | Description |
|---------|-------------|
| `./scripts/test-cmds.sh` | Health + login + session (prints one-liners at end) |
| `./scripts/test-cmds.sh http://localhost:3000` | Same, using dashboard proxy as base |
| `./scripts/curl-login.sh` | Login only; `./scripts/curl-login.sh email password` for custom creds |
| `make test-api-cmds` | Runs `scripts/test-cmds.sh` (respects `API_URL` env) |

## Via dashboard proxy

If the API is only reachable through the dashboard (e.g. port 3000):

```bash
API_URL=http://localhost:3000 ./scripts/test-cmds.sh
```

```bash
curl -s -X POST http://localhost:3000/api/v1/auth/login \
  -H "Content-Type: application/json" \
  -d '{"email":"admin@functionfly.local","password":"admin123"}'
```

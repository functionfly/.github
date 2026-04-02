# Deploy FunctionFly Invite-Only Beta

This guide walks you through deploying FunctionFly with invite-only signup so your friend can test it.

## Architecture

```
┌─────────────────────┐     ┌─────────────────────┐     ┌─────────────────────┐
│  Cloudflare Pages   │────▶│   Fly.io (API)      │────▶│   Neon Postgres     │
│  (Dashboard UI)     │     │   (Go Orchestrator) │     │   (Database)        │
└─────────────────────┘     └─────────────────────┘     └─────────────────────┘
```

## Prerequisites

1. **Fly.io account** - Sign up at <https://fly.io>
2. **Cloudflare account** - Sign up at <https://cloudflare.com>
3. **Neon Postgres** - Sign up at <https://neon.tech> (or use Fly Postgres)
4. **Domain name** (optional) - For custom domain

## Step 1: Database Setup (Neon)

1. Create a new project in Neon
2. Create a database named `functionfly`
3. Copy the connection string (format: `postgres://user:pass@host.neon.tech/functionfly?sslmode=require`)
4. Save it - you'll need it in Step 3

## Step 2: Deploy API to Fly.io

### 2.1 Install Fly CLI

```bash
curl -L https://fly.io/install.sh | sh
export FLYCTL_INSTALL="$HOME/.fly"
export PATH="$FLYCTL_INSTALL/bin:$PATH"
```

### 2.2 Login to Fly

```bash
fly auth login
```

### 2.3 Create App (first time only)

```bash
cd deploy/fly/functionfly-control
fly apps create functionfly-control
```

### 2.4 Set Secrets

```bash
# Required: Database
fly secrets set DATABASE_URL="postgres://..." -a functionfly-control

# Required: Invite-only mode
fly secrets set SIGNUP_REQUIRE_INVITE_CODE="true" -a functionfly-control

# Required: JWT secrets (generate strong random strings)
fly secrets set JWT_SECRET="$(openssl rand -hex 32)" -a functionfly-control
fly secrets set ADMIN_JWT_SECRET="$(openssl rand -hex 32)" -a functionfly-control

# Required: Session encryption
fly secrets set SESSION_KEY="$(openssl rand -hex 32)" -a functionfly-control

# Required: API shared secret (for HMAC - same as your .env.development)
fly secrets set API_SHARED_SECRET="your-api-shared-secret-here" -a functionfly-control

# Optional: OAuth providers (if you want social login)
# fly secrets set GOOGLE_OAUTH_CLIENT_ID="..." -a functionfly-control
# fly secrets set GOOGLE_OAUTH_CLIENT_SECRET="..." -a functionfly-control
# fly secrets set GITHUB_OAUTH_CLIENT_ID="..." -a functionfly-control
# fly secrets set GITHUB_OAUTH_CLIENT_SECRET="..." -a functionfly-control
```

### 2.5 Deploy

```bash
fly deploy --config deploy/fly/functionfly-control/fly.toml --remote-only
```

### 2.6 Verify

```bash
fly status -a functionfly-control
fly logs -a functionfly-control
```

Get your Fly.io app URL:

```bash
fly info -a functionfly-control
# Will show something like: https://functionfly-control.fly.dev
```

Save this URL - you'll need it for the dashboard.

## Step 3: Deploy Dashboard to Cloudflare Pages

### 3.1 Build Dashboard

```bash
cd web/dashboard

# Create production env file
cat > .env.production << 'EOF'
VITE_API_URL=https://functionfly-control.fly.dev
VITE_ADMIN_SHARED_SECRET=your-api-shared-secret-here
EOF

# Install dependencies
npm install

# Build
npm run build
```

### 3.2 Deploy to Cloudflare Pages

**Option A: Drag & Drop (Quickest)**

1. Go to <https://dash.cloudflare.com> → Pages
2. Click "Create a project" → "Upload assets"
3. Drag the `web/dashboard/dist` folder
4. Your site will be live at `https://your-project.pages.dev`

**Option B: Git Integration (CI/CD)**

1. Push code to GitHub
2. In Cloudflare Pages, connect your GitHub repo
3. Set build command: `cd web/dashboard && npm install && npm run build`
4. Set output directory: `web/dashboard/dist`
5. Add environment variable: `VITE_API_URL` = your Fly.io URL

## Step 4: Create Admin Account

Once deployed, create your admin account:

```bash
# Run this locally against your production DB
export DATABASE_URL="your-neon-connection-string"
go run ./cmd/create-admin \
  -email "you@example.com" \
  -name "Your Name" \
  -password "your-secure-password"
```

## Step 5: Test the Flow

### 5.1 Login to Admin Dashboard

1. Go to your Cloudflare Pages URL (e.g., `https://your-project.pages.dev`)
2. Click "Admin Login"
3. Sign in with your admin credentials

### 5.2 Create an Invite Code

1. Navigate to "Signup Invites" in the admin sidebar
2. Click "Create Invite"
3. Enter a label (e.g., "Friend Beta Test")
4. Click "Create"
5. **COPY THE CODE IMMEDIATELY** - it won't be shown again!

### 5.3 Friend Signs Up

1. Share the invite code with your friend
2. They go to your dashboard URL
3. Click "Sign Up"
4. Enter their email, password, and the invite code
5. Complete signup

## Environment Variables Reference

### Backend (Fly.io)

| Variable | Required | Description |
|----------|----------|-------------|
| `DATABASE_URL` | Yes | Postgres connection string |
| `SIGNUP_REQUIRE_INVITE_CODE` | Yes | Set to `true` for invite-only |
| `JWT_SECRET` | Yes | Random 64-char hex string |
| `ADMIN_JWT_SECRET` | Yes | Different random 64-char hex string |
| `SESSION_KEY` | Yes | Random 64-char hex string |
| `API_SHARED_SECRET` | Yes | Same as dashboard, for HMAC signing |
| `REDIS_ADDR` | No | Redis for sessions (optional) |
| `GOOGLE_OAUTH_CLIENT_ID` | No | For Google OAuth |
| `GITHUB_OAUTH_CLIENT_ID` | No | For GitHub OAuth |

### Frontend (Cloudflare Pages)

| Variable | Required | Description |
|----------|----------|-------------|
| `VITE_API_URL` | Yes | Your Fly.io API URL |
| `VITE_ADMIN_SHARED_SECRET` | Yes | Same as backend `API_SHARED_SECRET` |

## Troubleshooting

### "Invalid invite code" error

- Check that `SIGNUP_REQUIRE_INVITE_CODE=true` is set in Fly.io
- Verify the code was copied correctly (16 characters)
- Check if the invite has reached max uses or expired

### "Cannot connect to API" error

- Verify `VITE_API_URL` is set correctly in dashboard build
- Check Fly.io app is running: `fly status`
- Check CORS is enabled on backend (should be by default)

### HMAC/Signature errors

- Ensure `VITE_ADMIN_SHARED_SECRET` matches `API_SHARED_SECRET`
- Both must be set before deploying

## Next Steps

After your friend tests successfully:

1. Set up a custom domain in Cloudflare
2. Configure email (Resend/Postmark) for transactional emails
3. Add more invite codes for other friends
4. Monitor logs: `fly logs -a functionfly-control`

## Quick Commands Reference

```bash
# View logs
fly logs -a functionfly-control

# Restart app
fly apps restart functionfly-control

# Scale up (if needed)
fly scale count 2 -a functionfly-control

# Connect to database
fly pg connect -a functionfly-control-db

# Update secrets
fly secrets set KEY=value -a functionfly-control
```

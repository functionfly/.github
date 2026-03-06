# ⚡ FunctionFly Production Quick Start

5-minute deployment guide for experienced DevOps engineers.

---

## 📋 Prerequisites ⏱️ 30s

| Requirement | Status |
|-------------|--------|
| 2× OVHcloud KS-5 servers ordered | [ ] Verify SSH access: `ssh root@<server-ip>` |
| Domain in Cloudflare | [ ] Check: `dig functionfly.com +short` |
| Docker installed | [ ] Verify: `docker --version` |

**Server Roles:**
- **Server 1** (`<server1-ip>`): App Stack (API, PostgreSQL, Redis, Caddy)
- **Server 2** (`<server2-ip>`): Runtime Stack (Local, Node.js, Python, ClamAV, YARA)

---

## 🚀 TL;DR Deploy ⏱️ 2min

```bash
# === SERVER 1: App Stack ===
ssh root@<server1-ip> "bash -s" << 'EOF'
mkdir -p /opt/functionfly && cd /opt/functionfly
curl -fsSL https://get.docker.com | sh
usermod -aG docker root
# Copy docker-compose.yml and .env, then:
docker compose up -d
EOF

# === SERVER 2: Runtime Stack ===
ssh root@<server2-ip> "bash -s" << 'EOF'
mkdir -p /opt/functionfly && cd /opt/functionfly
curl -fsSL https://get.docker.com | sh
usermod -aG docker root
# Copy docker-compose.yml and .env, then:
docker compose up -d
EOF
```

---

## 📖 Step-by-Step Quick Deploy ⏱️ 5min

### Step 1: Clone & Env Setup ⏱️ 1min

```bash
# On both servers
mkdir -p /opt/functionfly/{data,backups,logs,caddy}
cd /opt/functionfly

# Generate secrets
JWT_SECRET=$(openssl rand -base64 48)
API_SECRET=$(openssl rand -base64 48)
REDIS_PASS=$(openssl rand -base64 32)
DB_ENCRYPTION_KEY=$(openssl rand -base64 32)
```

**Create `/opt/functionfly/.env` on Server 1:**
```bash
cat > /opt/functionfly/.env << EOF
NODE_ENV=production
ENVIRONMENT=production
PORT=8080

# Neon Database (update after Step 5)
DB_HOST=ep-xxxxx.us-east-1.aws.neon.tech
DB_PORT=5432
DB_USER=functionfly_owner
DB_PASSWORD=your-neon-password
DB_NAME=functionfly
DB_SSLMODE=require
DB_ENCRYPTION_ENABLED=true
DB_MASTER_KEY_PASSWORD=${DB_ENCRYPTION_KEY}

# Redis
REDIS_ADDR=redis:6379
REDIS_PASSWORD=${REDIS_PASS}
REDIS_DB=0

# Secrets
JWT_SECRET=${JWT_SECRET}
API_SHARED_SECRET=${API_SECRET}

# URLs
BASE_URL=https://api.functionfly.com
FRONTEND_URL=https://app.functionfly.com

# Runtime Stack (Server 2)
CLAMAV_URL=http://<server2-ip>:3310
YARA_URL=http://<server2-ip>:8084

# Rate Limiting
RATE_LIMIT_REQUESTS=100
RATE_LIMIT_WINDOW_SECONDS=60

LOG_LEVEL=info
EOF
chmod 600 /opt/functionfly/.env
```

**Create `/opt/functionfly/.env` on Server 2:**
```bash
cat > /opt/functionfly/.env << EOF
NODE_ENV=production
ENVIRONMENT=production
ORCHESTRATOR_URL=http://<server1-ip>:8080
API_SHARED_SECRET=${API_SECRET}
RUNTIME_LOCAL_PORT=8081
RUNTIME_NODEJS_PORT=8082
RUNTIME_PYTHON_PORT=8083
YARA_PORT=8084
MAX_CONCURRENT_EXECUTIONS=100
EXECUTION_TIMEOUT_SECONDS=300
MEMORY_LIMIT_MB=8192
EOF
chmod 600 /opt/functionfly/.env
```

---

### Step 2: Deploy App Stack (Server 1) ⏱️ 2min

```bash
ssh root@<server1-ip>
cd /opt/functionfly

# 1. Copy docker-compose.yml from repo
curl -o docker-compose.yml https://raw.githubusercontent.com/functionfly/functionfly/main/docker-compose.production.yml

# 2. Create Caddyfile
mkdir -p caddy
cat > caddy/Caddyfile << 'EOF'
{
    auto_https disable_redirects
    email ops@functionfly.com
}

:80 {
    respond /health "OK" 200
}

functionfly.com, api.functionfly.com, app.functionfly.com, admin.functionfly.com, run.functionfly.com, registry.functionfly.com, docs.functionfly.com {
    header {
        X-Frame-Options "DENY"
        X-Content-Type-Options "nosniff"
    }
    reverse_proxy orchestrator-api:8080
}
EOF

# 3. Deploy
docker compose pull
docker compose up -d

# 4. Verify
docker compose ps
```

---

### Step 3: Deploy Runtime Stack (Server 2) ⏱️ 1min

```bash
ssh root@<server2-ip>
cd /opt/functionfly

# 1. Copy docker-compose.yml
curl -o docker-compose.yml https://raw.githubusercontent.com/functionfly/functionfly/main/docker-compose.runtime.yml

# 2. Create YARA service
mkdir -p yara
cat > yara/Dockerfile << 'EOF'
FROM python:3.11-slim
RUN apt-get update && apt-get install -y yara libyara-dev gcc && rm -rf /var/lib/apt/lists/*
RUN pip install flask gunicorn yara-python requests
COPY yara_service.py /app/
WORKDIR /app
EXPOSE 8080
CMD ["python", "yara_service.py"]
EOF

# 3. Deploy
docker compose pull
docker compose build yara
docker compose up -d

# 4. Verify
docker compose ps
```

---

### Step 4: Verify Deployment ⏱️ 1min

**Health Checks:**
```bash
# Server 1 - API health
curl -s https://api.functionfly.com/healthz

# Server 1 - Caddy
curl -s https://functionfly.com/health

# Server 2 - Runtimes (from Server 1)
curl -s http://<server2-ip>:8081/health
curl -s http://<server2-ip>:8082/health
curl -s http://<server2-ip>:8083/health

# Server 2 - Security scanners
curl -s http://<server2-ip>:3310
curl -s http://<server2-ip>:8084/health
```

**SSL Verification:**
```bash
echo | openssl s_client -connect functionfly.com:443 2>/dev/null | openssl x509 -noout -dates
```

---

## 🔍 Quick Verification Commands

### Container Status
```bash
# Server 1
docker stats --no-stream
docker compose logs --tail=50 orchestrator-api

# Server 2
docker stats --no-stream
docker compose logs --tail=50 runtime-local
```

### Test Function Deployment
```bash
# Get API token first (via dashboard or CLI)
export FF_TOKEN="your-api-token"

# Deploy test function
curl -X POST https://api.functionfly.com/v1/functions \
  -H "Authorization: Bearer ${FF_TOKEN}" \
  -H "Content-Type: application/json" \
  -d '{
    "name": "hello-quickstart",
    "runtime": "local",
    "code": "export default async function(req) { return { body: \"Hello from FunctionFly!\" }; }"
  }'

# Execute
curl -X POST https://api.functionfly.com/v1/execute/hello-quickstart \
  -H "Authorization: Bearer ${FF_TOKEN}" \
  -H "Content-Type: application/json" \
  -d '{}'
```

### Database Connection (Neon)
```bash
# Update .env with Neon credentials, then:
docker compose restart orchestrator-api
docker compose logs orchestrator-api | grep -i "database\|connected"
```

---

## 📚 Next Steps

| Resource | Description |
|----------|-------------|
| [📖 PRODUCTION_DEPLOYMENT.md](./PRODUCTION_DEPLOYMENT.md) | Full deployment guide with security hardening, monitoring, and troubleshooting |
| [⚡ PERFORMANCE_TUNING_GUIDE.md](./PERFORMANCE_TUNING_GUIDE.md) | Optimization, load testing, and scaling strategies |
| [🔧 RUNBOOK.md](../RUNBOOK.md) | Common operations and incident response |

---

## 💰 Cost Summary

| Component | Monthly Cost |
|-----------|-------------|
| 2× OVHcloud KS-5 | $40 |
| Neon PostgreSQL | $25 |
| Cloudflare (Free) | $0 |
| **Total** | **$65/mo** |

---

## 🆘 Common Issues

**Issue:** `docker: command not found`  
**Fix:** `curl -fsSL https://get.docker.com | sh`

**Issue:** Containers failing to start  
**Fix:** Check logs: `docker compose logs -f`

**Issue:** Cannot connect to Server 2 from Server 1  
**Fix:** Verify UFW rules: `ufw allow from <server1-ip>`

**Issue:** SSL certificate errors  
**Fix:** Ensure Cloudflare SSL mode is "Full (strict)"

---

*Deploy time: ~5 minutes | For detailed configuration, see [PRODUCTION_DEPLOYMENT.md](./PRODUCTION_DEPLOYMENT.md)*

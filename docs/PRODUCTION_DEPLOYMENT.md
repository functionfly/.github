# 🚀 FunctionFly Production Deployment Guide

Complete production deployment guide for FunctionFly serverless platform using bare metal servers and managed PostgreSQL.

---

## 📋 Table of Contents

- [Infrastructure Overview](#infrastructure-overview)
- [Step 1: Provision Servers](#step-1-provision-servers)
- [Step 2: DNS Configuration](#step-2-dns-configuration)
- [Step 3: Server 1 - App Stack](#step-3-server-1---app-stack)
- [Step 4: Server 2 - Runtime Stack](#step-4-server-2---runtime-stack)
- [Step 5: Neon Database Setup](#step-5-neon-database-setup)
- [Step 6: Verification](#step-6-verification)
- [Scaling Roadmap](#scaling-roadmap)
- [Troubleshooting](#troubleshooting)

---

## 🏗️ Infrastructure Overview

### Cost Summary

| Component | Provider | Specification | Monthly Cost |
|-----------|----------|---------------|--------------|
| **Server 1 (App Stack)** | OVHcloud | KS-5 (64GB RAM, 1TB SSD) | $20/month |
| **Server 2 (Runtime)** | OVHcloud | KS-5 (64GB RAM, 1TB SSD) | $20/month |
| **Database** | Neon | Managed PostgreSQL | $25/month |
| **DNS/CDN** | Cloudflare | Free Plan | $0/month |
| **Total Baseline** | - | - | **$65/month** |

### Server Roles

```
┌─────────────────────────────────────────────────────────────────────────┐
│                    FunctionFly Production Architecture                  │
├─────────────────────────────────────────────────────────────────────────┤
│                                                                         │
│  ┌──────────────────────┐              ┌──────────────────────────┐    │
│  │   Server 1: App      │              │   Server 2: Runtime      │    │
│  │   Stack ($20/mo)     │◄────────────►│   Stack ($20/mo)         │    │
│  ├──────────────────────┤              ├──────────────────────────┤    │
│  │                      │              │                          │    │
│  │  • Orchestrator API  │              │  • runtime-local         │    │
│  │  • PostgreSQL        │              │  • runtime-nodejs        │    │
│  │  • Redis             │              │  • runtime-python        │    │
│  │  • Caddy Proxy       │              │  • ClamAV                │    │
│  │  • Frontend (Static) │              │  • YARA Scanner          │    │
│  │                      │              │                          │    │
│  └──────────┬───────────┘              └──────────────┬───────────┘    │
│             │                                         │                │
│             └─────────────────────────────────────────┘                │
│                           Internal Network                             │
│                                                                         │
│  ┌──────────────────────────────────────────────────────────────────┐  │
│  │                     Neon Managed PostgreSQL                      │  │
│  │                          ($25/mo)                                │  │
│  │     • Automatic backups    • Connection pooling                  │  │
│  │     • Branching            • Monitoring                          │  │
│  └──────────────────────────────────────────────────────────────────┘  │
│                                                                         │
└─────────────────────────────────────────────────────────────────────────┘
```

### Service Ports

| Service | Server | Port | Access |
|---------|--------|------|--------|
| Orchestrator API | Server 1 | 8080 | Internal/Caddy |
| PostgreSQL | Server 1 | 5432 | Localhost only |
| Redis | Server 1 | 6379 | Localhost only |
| Caddy HTTP | Server 1 | 80 | Public |
| Caddy HTTPS | Server 1 | 443 | Public |
| Runtime Local | Server 2 | 8081 | Internal |
| Runtime Node.js | Server 2 | 8082 | Internal |
| Runtime Python | Server 2 | 8083 | Internal |
| ClamAV | Server 2 | 3310 | Internal |
| YARA | Server 2 | 8084 | Internal |

---

## Step 1: Provision Servers

### 1.1 Order OVHcloud KS-5 Servers

1. Visit [OVHcloud US Bare Metal](https://us.ovhcloud.com/bare-metal/servers/)
2. Select **Kimsufi KS-5** or equivalent:
   - CPU: Intel Xeon or AMD EPYC
   - RAM: 64GB DDR4
   - Storage: 1TB SSD
   - Bandwidth: 1Gbps unmetered
3. Order **2 servers** (one for App Stack, one for Runtime Stack)
4. Select **Ubuntu 22.04 LTS** as the operating system
5. Complete checkout (~$40/month total)

### 1.2 Initial Server Setup

SSH into each server and perform initial configuration:

```bash
# Server 1 - App Stack (replace with your actual IP)
ssh root@<server1-ip>

# Server 2 - Runtime Stack (replace with your actual IP)
ssh root@<server2-ip>
```

**Run on both servers:**

```bash
# Update system packages
apt update && apt upgrade -y

# Install essential tools
apt install -y curl wget git vim ufw fail2ban htop iotop net-tools

# Set timezone
timedatectl set-timezone America/Chicago

# Create functionfly user
useradd -m -s /bin/bash functionfly
usermod -aG sudo functionfly

# Set up SSH key for functionfly user (copy from root)
mkdir -p /home/functionfly/.ssh
cp /root/.ssh/authorized_keys /home/functionfly/.ssh/
chown -R functionfly:functionfly /home/functionfly/.ssh
chmod 700 /home/functionfly/.ssh
chmod 600 /home/functionfly/.ssh/authorized_keys

# Disable root SSH login (optional but recommended)
sed -i 's/PermitRootLogin yes/PermitRootLogin no/' /etc/ssh/sshd_config
systemctl restart sshd
```

### 1.3 Install Docker

**Run on both servers:**

```bash
# Install Docker
curl -fsSL https://get.docker.com -o get-docker.sh
sh get-docker.sh

# Add functionfly user to docker group
usermod -aG docker functionfly

# Install Docker Compose
apt install -y docker-compose-plugin

# Verify installation
docker --version
docker compose version

# Enable Docker on boot
systemctl enable docker
```

### 1.4 Configure UFW Firewall

**On Server 1 (App Stack):**

```bash
# Reset UFW
ufw --force reset

# Default deny
ufw default deny incoming
ufw default allow outgoing

# Allow SSH (adjust port if you've changed it)
ufw allow 22/tcp

# Allow HTTP/HTTPS
ufw allow 80/tcp
ufw allow 443/tcp

# Allow internal communication from Server 2 (replace with actual Server 2 IP)
ufw allow from <server2-ip> to any port 8080
ufw allow from <server2-ip> to any port 5432
ufw allow from <server2-ip> to any port 6379

# Enable firewall
ufw --force enable

# Check status
ufw status verbose
```

**On Server 2 (Runtime Stack):**

```bash
# Reset UFW
ufw --force reset

# Default deny
ufw default deny incoming
ufw default allow outgoing

# Allow SSH
ufw allow 22/tcp

# Allow from Server 1 only (replace with actual Server 1 IP)
ufw allow from <server1-ip> to any port 8081
ufw allow from <server1-ip> to any port 8082
ufw allow from <server1-ip> to any port 8083
ufw allow from <server1-ip> to any port 3310
ufw allow from <server1-ip> to any port 8084

# Enable firewall
ufw --force enable

# Check status
ufw status verbose
```

**Verification:**
```bash
# On both servers - verify Docker is running
systemctl status docker

# Test Docker
docker run hello-world
```

---

## Step 2: DNS Configuration

### 2.1 Cloudflare Setup

For a single reference covering DNS, CDN, R2, Workers, Tunnel, and Pages, see [CLOUDFLARE.md](CLOUDFLARE.md).

1. Log into [Cloudflare Dashboard](https://dash.cloudflare.com)
2. Select your domain (`functionfly.com`)
3. Go to **DNS** → **Records**

### 2.2 Create DNS Records

| Type | Name | Content | Proxy Status | TTL |
|------|------|---------|--------------|-----|
| A | `@` | `<server1-ip>` | Proxied | Auto |
| A | `api` | `<server1-ip>` | Proxied | Auto |
| A | `app` | `<server1-ip>` | Proxied | Auto |
| A | `admin` | `<server1-ip>` | Proxied | Auto |
| A | `run` | `<server1-ip>` | Proxied | Auto |
| A | `registry` | `<server1-ip>` | Proxied | Auto |
| A | `docs` | `<server1-ip>` | Proxied | Auto |
| A | `runtime` | `<server2-ip>` | DNS Only | Auto |

### 2.3 SSL/TLS Configuration

1. Go to **SSL/TLS** → **Overview**
2. Set encryption mode to **Full (strict)**
3. Go to **SSL/TLS** → **Edge Certificates**
4. Enable **Always Use HTTPS**
5. Enable **Automatic HTTPS Rewrites**

### 2.4 Security Settings

1. Go to **Security** → **Bots**
2. Enable **Bot Fight Mode**
3. Go to **Security** → **DDoS**
4. Ensure DDoS protection is enabled

### 2.5 Verification

```bash
# Wait 2-5 minutes for DNS propagation, then verify
dig functionfly.com +short
dig api.functionfly.com +short
dig app.functionfly.com +short

# Should return Server 1's IP address (or Cloudflare proxy IPs)
```

---

## Step 3: Server 1 - App Stack

### 3.1 Create Directory Structure

```bash
# Switch to functionfly user
su - functionfly

# Create directory structure
mkdir -p /opt/functionfly/{data,backups,logs}
mkdir -p /opt/functionfly/caddy
mkdir -p /opt/functionfly/postgres

# Set permissions
chmod 755 /opt/functionfly
chmod 750 /opt/functionfly/data
chmod 750 /opt/functionfly/backups
chmod 755 /opt/functionfly/logs
```

### 3.2 Environment Configuration

Create the production environment file:

```bash
# Create .env file
cat > /opt/functionfly/.env << 'EOF'
# ============================================================================
# FunctionFly Production Environment Configuration
# ============================================================================

# Core Environment
NODE_ENV=production
ENVIRONMENT=production
PORT=8080

# ============================================================================
# NEON DATABASE CONFIGURATION
# ============================================================================
# Replace with your actual Neon connection string after Step 5
DB_HOST=ep-production-xxxxx.us-east-1.aws.neon.tech
DB_PORT=5432
DB_USER=functionfly_owner
DB_PASSWORD=your-secure-neon-password
DB_NAME=functionfly
DB_SSLMODE=require

# ============================================================================
# DATABASE ENCRYPTION (Data at Rest) - REQUIRED FOR PRODUCTION
# ============================================================================
# ⚠️  CRITICAL: Database encryption is MANDATORY for production deployment
#    This protects sensitive data (secrets, API keys, tokens) at rest using AES-256-GCM
#    The server uses a zero-knowledge architecture - master key is never stored
#
# Key Management:
# - Generate master key: openssl rand -hex 32
# - Store the key securely (Infisical, Vault, or secure secrets manager)
# - The key encrypts ALL sensitive fields in the database
# - If the key is lost, encrypted data CANNOT be recovered - keep a secure backup!
#
DB_ENCRYPTION_ENABLED=true
DB_MASTER_KEY_PASSWORD=your-secure-master-key-min-32-characters-long

# ============================================================================
# REDIS CONFIGURATION
# ============================================================================
REDIS_ADDR=redis:6379
REDIS_PASSWORD=your-secure-redis-password-change-this
REDIS_DB=0
ARTIFACT_TTL=168h

# Cache Configuration
CACHE_REDIS_ENABLED=true
CACHE_MEMORY_MAX_MB=512
CACHE_DISK_ENABLED=true
CACHE_REDIS_REGISTRY_TTL=3600
CACHE_DEFAULT_TTL=7200

# CDN Configuration
CACHE_CDN_ENABLED=true
CACHE_CDN_PROVIDER=cloudflare
CACHE_CDN_BASE_URL=https://cdn.functionfly.com
CACHE_CDN_MAX_AGE=86400

# Edge Cache Configuration
CACHE_EDGE_ENABLED=true
CACHE_EDGE_MIN_POPULARITY=100
CACHE_EDGE_MIN_EXECUTIONS=500
CACHE_EDGE_MIN_TRUST_SCORE=80.0
CACHE_EDGE_MIN_SUCCESS_RATE=98.0
CACHE_EDGE_MAX_LATENCY_MS=3000
CACHE_EDGE_DURATION=24h
CACHE_EDGE_MAX_FUNCTIONS=500
CACHE_EDGE_REFRESH_INTERVAL=30m

# ============================================================================
# SECURITY CONFIGURATION
# ============================================================================
# JWT and API Secrets - Generate with: openssl rand -base64 48
JWT_SECRET=your-production-jwt-secret-min-64-characters-long-change-immediately
API_SHARED_SECRET=your-production-api-secret-min-64-characters-long-change-immediately

# Rate Limiting
RATE_LIMIT_REQUESTS=100
RATE_LIMIT_WINDOW_SECONDS=60

# CORS Configuration
CORS_ALLOWED_ORIGINS=https://app.functionfly.com,https://admin.functionfly.com,https://api.functionfly.com
CORS_ALLOWED_METHODS=GET,POST,PUT,PATCH,DELETE,OPTIONS
CORS_ALLOWED_HEADERS=Accept,Content-Type,Content-Length,Accept-Encoding,X-CSRF-Token,Authorization,X-FFLY-Timestamp,X-FFLY-Signature,x-neon-client-info

# Content Security Policy
CONTENT_SECURITY_POLICY=default-src 'self';script-src 'self';style-src 'self' 'unsafe-inline';img-src 'self' data: https:;font-src 'self' https:;connect-src 'self' https://api.functionfly.com;
HSTS_MAX_AGE=31536000

# Advanced Security Middleware
ADVANCED_SECURITY_SLIDING_WINDOW_LIMIT=100
ADVANCED_SECURITY_SLIDING_WINDOW_MINUTES=1
ADVANCED_SECURITY_TOKEN_BUCKET_RATE=10.0
ADVANCED_SECURITY_TOKEN_BUCKET_BURST=20
ADVANCED_SECURITY_ENABLE_BOT_DETECTION=true
ADVANCED_SECURITY_ENABLE_TRAFFIC_ANALYSIS=true
ADVANCED_SECURITY_SUSPICIOUS_THRESHOLD=10
ADVANCED_SECURITY_BLOCK_MINUTES=30
ADVANCED_SECURITY_CIRCUIT_BREAKER_THRESHOLD=0.6
ADVANCED_SECURITY_CIRCUIT_BREAKER_MINUTES=2
ADVANCED_SECURITY_QUEUE_SIZE=2000
ADVANCED_SECURITY_QUEUE_SECONDS=60
ADVANCED_SECURITY_BLOCKED_COUNTRIES=
ADVANCED_SECURITY_BLOCKED_IPS=
ADVANCED_SECURITY_ALLOWED_IPS=
ADVANCED_SECURITY_ENABLE_SQL_INJECTION_FILTER=true
ADVANCED_SECURITY_ENABLE_XSS_FILTER=true
ADVANCED_SECURITY_ENABLE_PATH_TRAVERSAL_FILTER=true
ADVANCED_SECURITY_METRICS_ENABLED=true

# ============================================================================
# URL CONFIGURATION
# ============================================================================
BASE_URL=https://api.functionfly.com
FRONTEND_URL=https://app.functionfly.com

# ============================================================================
# EMAIL CONFIGURATION (Production SMTP)
# ============================================================================
SMTP_HOST=smtp.sendgrid.net
SMTP_PORT=587
SMTP_USERNAME=apikey
SMTP_PASSWORD=your-sendgrid-api-key
FROM_EMAIL=noreply@functionfly.com
FROM_NAME=FunctionFly

# ============================================================================
# OAUTH PROVIDER CONFIGURATION
# ============================================================================
GOOGLE_CLIENT_ID=your-google-client-id.apps.googleusercontent.com
GOOGLE_CLIENT_SECRET=your-google-client-secret
GITHUB_CLIENT_ID=your-github-client-id
GITHUB_CLIENT_SECRET=your-github-client-secret
GITHUB_OWNER=functionfly
GITHUB_REPO=functionfly
GITHUB_TOKEN=ghp_your-github-token

# ============================================================================
# FUNCTION VERIFICATION SERVICES
# ============================================================================
# Point to Server 2 runtime stack
CLAMAV_URL=http://<server2-ip>:3310
YARA_URL=http://<server2-ip>:8084
VERIFICATION_ENABLED=true
VERIFICATION_TIMEOUT_SECONDS=30
VERIFICATION_MAX_FILE_SIZE_MB=50
VERIFICATION_CACHE_TTL_MINUTES=60

# Trust Levels
MINIMUM_TRUST_LEVEL=standard
TRUST_LEVEL_STANDARD_ENABLED=true
TRUST_LEVEL_HIGH_ENABLED=true
TRUST_LEVEL_ENTERPRISE_ENABLED=true
TRUST_LEVEL_STANDARD_REQUIRE_MALWARE_SCAN=true
TRUST_LEVEL_HIGH_REQUIRE_SIGNATURE_VERIFICATION=true
TRUST_LEVEL_HIGH_REQUIRE_APPROVAL=true
TRUST_LEVEL_ENTERPRISE_REQUIRE_MANUAL_APPROVAL=true

# ============================================================================
# ARCHIVE STORAGE
# ============================================================================
ARCHIVE_ENABLED=true
ARCHIVE_RETENTION_DAYS=2555
ARCHIVE_CLEANUP_INTERVAL_HOURS=168

# ============================================================================
# LOGGING
# ============================================================================
LOG_LEVEL=info

# ============================================================================
# STRIPE PAYMENT PROCESSING (use .env; never commit real keys)
# Get values from Stripe Dashboard → Developers → API keys / Webhooks
# ============================================================================
STRIPE_PUBLISHABLE_KEY=<your-publishable-key-from-dashboard>
STRIPE_SECRET_KEY=<your-secret-key-from-dashboard>
STRIPE_WEBHOOK_SECRET=<your-webhook-signing-secret>
EOF

# Secure the environment file
chmod 600 /opt/functionfly/.env
```

**Generate secure secrets:**
```bash
# Generate JWT_SECRET
openssl rand -base64 48

# Generate API_SHARED_SECRET
openssl rand -base64 48

# Generate Redis password
openssl rand -base64 32

# Update the .env file with generated values
nano /opt/functionfly/.env
```

### 3.3 Docker Compose Configuration

Create `/opt/functionfly/docker-compose.yml`:

```yaml
version: '3.8'

# ============================================================================
# FunctionFly Production - Server 1: App Stack
# ============================================================================
# Services: PostgreSQL, Redis, Orchestrator API, Caddy Proxy
#
# Domain Mapping:
#   - functionfly.com          -> Caddy (landing/marketing)
#   - api.functionfly.com      -> Orchestrator API
#   - app.functionfly.com       -> User dashboard (SPA)
#   - admin.functionfly.com     -> Admin UI (separate origin)
#   - run.functionfly.com      -> Function playground
#   - registry.functionfly.com -> Function registry/marketplace
#   - docs.functionfly.com     -> Documentation
#
# ============================================================================

services:
  # PostgreSQL with pgvector extension (pg17; use pg16 if needed)
  postgres:
    image: pgvector/pgvector:pg17
    container_name: functionfly-postgres
    hostname: postgres
    environment:
      POSTGRES_DB: ${DB_NAME:-functionfly}
      POSTGRES_USER: ${DB_USER:-functionfly}
      POSTGRES_PASSWORD: ${DB_PASSWORD}
      POSTGRES_INITDB_ARGS: "--encoding=UTF-8 --lc-collate=C --lc-ctype=C"
    volumes:
      - postgres_data:/var/lib/postgresql/data
      - ./backups:/var/backups
    ports:
      - "127.0.0.1:5432:5432"  # Localhost only for security
    command:
      - "postgres"
      - "-c"
      - "shared_buffers=4GB"
      - "-c"
      - "effective_cache_size=12GB"
      - "-c"
      - "maintenance_work_mem=512MB"
      - "-c"
      - "work_mem=64MB"
      - "-c"
      - "max_connections=200"
    healthcheck:
      test: ["CMD-SHELL", "pg_isready -U ${DB_USER:-functionfly}"]
      interval: 10s
      timeout: 5s
      retries: 5
      start_period: 30s
    restart: unless-stopped
    networks:
      - functionfly-prod
    deploy:
      resources:
        limits:
          memory: 8G
          cpus: '4.0'
        reservations:
          memory: 4G
          cpus: '2.0'

  # Redis for caching and session storage
  redis:
    image: redis:7-alpine
    container_name: functionfly-redis
    hostname: redis
    command: >
      redis-server
      --appendonly yes
      --maxmemory 4gb
      --maxmemory-policy allkeys-lru
      --requirepass ${REDIS_PASSWORD}
    volumes:
      - redis_data:/data
    ports:
      - "127.0.0.1:6379:6379"  # Localhost only
    healthcheck:
      test: ["CMD", "redis-cli", "-a", "${REDIS_PASSWORD}", "ping"]
      interval: 10s
      timeout: 3s
      retries: 3
    restart: unless-stopped
    networks:
      - functionfly-prod
    deploy:
      resources:
        limits:
          memory: 4G
          cpus: '2.0'
        reservations:
          memory: 2G
          cpus: '1.0'

  # FunctionFly Orchestrator API
  orchestrator-api:
    image: functionfly/orchestrator:latest
    container_name: functionfly-orchestrator
    hostname: orchestrator-api
    depends_on:
      postgres:
        condition: service_healthy
      redis:
        condition: service_healthy
    ports:
      - "127.0.0.1:8080:8080"  # Access via Caddy only
    env_file:
      - .env
    environment:
      - NODE_ENV=production
      - ENVIRONMENT=production
    healthcheck:
      test: ["CMD", "wget", "-q", "--spider", "http://localhost:8080/healthz"]
      interval: 30s
      timeout: 10s
      retries: 3
      start_period: 60s
    restart: unless-stopped
    networks:
      - functionfly-prod
    deploy:
      resources:
        limits:
          memory: 4G
          cpus: '4.0'
        reservations:
          memory: 2G
          cpus: '2.0'

  # Caddy Reverse Proxy
  caddy:
    image: caddy:2-alpine
    container_name: functionfly-caddy
    hostname: caddy
    depends_on:
      - orchestrator-api
    ports:
      - "80:80"
      - "443:443"
    volumes:
      - ./caddy/Caddyfile:/etc/caddy/Caddyfile:ro
      - caddy_data:/data
      - caddy_config:/config
    environment:
      - CADDYPATH=/data/caddy
    healthcheck:
      test: ["CMD", "wget", "-q", "--spider", "http://localhost:80/health"]
      interval: 30s
      timeout: 10s
      retries: 3
    restart: unless-stopped
    networks:
      - functionfly-prod
    deploy:
      resources:
        limits:
          memory: 512M
          cpus: '1.0'
        reservations:
          memory: 256M
          cpus: '0.5'

volumes:
  postgres_data:
    driver: local
    name: functionfly_postgres_data
  redis_data:
    driver: local
    name: functionfly_redis_data
  caddy_data:
    driver: local
    name: functionfly_caddy_data
  caddy_config:
    driver: local
    name: functionfly_caddy_config

networks:
  functionfly-prod:
    driver: bridge
    name: functionfly-prod
    ipam:
      config:
        - subnet: 172.30.0.0/16
```

### 3.4 Caddy Configuration

Create `/opt/functionfly/caddy/Caddyfile`:

```caddy
# FunctionFly Production Caddy Configuration
# Server 1: App Stack

{
    # Enable automatic HTTPS
    auto_https disable_redirects
    email ops@functionfly.com
    
    # Logging
    log {
        output file /data/access.log {
            roll_size 100MB
            roll_keep 10
            roll_keep_days 30
        }
        format json
    }
}

# Health check endpoint (all domains)
:80 {
    respond /health "OK" 200
}

# Main domain - functionfly.com
functionfly.com {
    # Rate limiting
    rate_limit {
        zone main {
            key {remote_host}
            window 1m
            events 200
        }
    }

    # Security headers
    header {
        X-Frame-Options "DENY"
        X-Content-Type-Options "nosniff"
        X-XSS-Protection "1; mode=block"
        Referrer-Policy "strict-origin-when-cross-origin"
        Permissions-Policy "geolocation=(), microphone=(), camera=()"
    }

    # API routes
    /v1/* {
        rate_limit {
            zone api {
                key {remote_host}
                window 1m
                events 100
            }
        }
        reverse_proxy orchestrator-api:8080
    }

    # Default route
    reverse_proxy orchestrator-api:8080
}

# API subdomain - api.functionfly.com
api.functionfly.com {
    rate_limit {
        zone api_subdomain {
            key {remote_host}
            window 1m
            events 200
        }
    }
    
    header {
        X-Frame-Options "DENY"
        X-Content-Type-Options "nosniff"
    }
    
    reverse_proxy orchestrator-api:8080
}

# App subdomain - app.functionfly.com (dashboard SPA)
app.functionfly.com {
    rate_limit {
        zone app {
            key {remote_host}
            window 1m
            events 100
        }
    }
    
    reverse_proxy orchestrator-api:8080
}

# Admin subdomain - admin.functionfly.com (admin UI, separate origin)
admin.functionfly.com {
    rate_limit {
        zone admin {
            key {remote_host}
            window 1m
            events 100
        }
    }
    
    reverse_proxy orchestrator-api:8080
}

# Run subdomain - run.functionfly.com
run.functionfly.com {
    rate_limit {
        zone run {
            key {remote_host}
            window 1m
            events 60
        }
    }
    
    reverse_proxy orchestrator-api:8080
}

# Registry subdomain - registry.functionfly.com
registry.functionfly.com {
    rate_limit {
        zone registry {
            key {remote_host}
            window 1m
            events 120
        }
    }
    
    reverse_proxy orchestrator-api:8080
}

# Docs subdomain - docs.functionfly.com
docs.functionfly.com {
    rate_limit {
        zone docs {
            key {remote_host}
            window 1m
            events 300
        }
    }
    
    reverse_proxy orchestrator-api:8080
}
```

### 3.5 Deploy Server 1

```bash
# Navigate to functionfly directory
cd /opt/functionfly

# Pull images
docker compose pull

# Start services
docker compose up -d

# Check status
docker compose ps

# View logs
docker compose logs -f

# Run database migrations (once Neon is configured)
# docker compose exec orchestrator-api npm run migrate
```

---

## Step 4: Server 2 - Runtime Stack

### 4.1 Create Directory Structure

```bash
# Switch to functionfly user on Server 2
su - functionfly

# Create directory structure
mkdir -p /opt/functionfly/{data,logs,yara-rules}

# Set permissions
chmod 755 /opt/functionfly
chmod 750 /opt/functionfly/data
chmod 755 /opt/functionfly/logs
chmod 750 /opt/functionfly/yara-rules
```

### 4.2 Environment Configuration

Create `/opt/functionfly/.env`:

```bash
cat > /opt/functionfly/.env << 'EOF'
# ============================================================================
# FunctionFly Runtime Stack - Server 2
# ============================================================================

NODE_ENV=production
ENVIRONMENT=production

# Server 1 connection (for callbacks)
ORCHESTRATOR_URL=http://<server1-ip>:8080
API_SHARED_SECRET=your-production-api-secret-must-match-server-1

# Runtime ports
RUNTIME_LOCAL_PORT=8081
RUNTIME_NODEJS_PORT=8082
RUNTIME_PYTHON_PORT=8083
YARA_PORT=8084

# Resource limits
MAX_CONCURRENT_EXECUTIONS=100
EXECUTION_TIMEOUT_SECONDS=300
MEMORY_LIMIT_MB=8192
EOF

chmod 600 /opt/functionfly/.env
```

### 4.3 Docker Compose Configuration

Create `/opt/functionfly/docker-compose.yml`:

```yaml
version: '3.8'

# ============================================================================
# FunctionFly Production - Server 2: Runtime Stack
# ============================================================================
# Services: Local Runtime, Node.js Runtime, Python Runtime, ClamAV, YARA
#
# This server handles function execution and security scanning.
# No public ports - only accessible from Server 1 via internal network.
#
# ============================================================================

services:
  # Local (WebAssembly) Runtime
  runtime-local:
    image: functionfly/runtime-local:latest
    container_name: functionfly-runtime-local
    hostname: runtime-local
    ports:
      - "127.0.0.1:8081:8080"
    environment:
      - NODE_ENV=production
      - PORT=8080
      - MAX_MEMORY_MB=4096
      - MAX_EXECUTION_TIME_MS=300000
      - ORCHESTRATOR_URL=${ORCHESTRATOR_URL}
      - API_SHARED_SECRET=${API_SHARED_SECRET}
    volumes:
      - runtime_local_data:/tmp/executions
    healthcheck:
      test: ["CMD", "wget", "-q", "--spider", "http://localhost:8080/health"]
      interval: 30s
      timeout: 10s
      retries: 3
    restart: unless-stopped
    networks:
      - functionfly-runtime
    deploy:
      resources:
        limits:
          memory: 8G
          cpus: '4.0'
        reservations:
          memory: 4G
          cpus: '2.0'

  # Node.js Runtime
  runtime-nodejs:
    image: functionfly/runtime-nodejs:latest
    container_name: functionfly-runtime-nodejs
    hostname: runtime-nodejs
    ports:
      - "127.0.0.1:8082:8080"
    environment:
      - NODE_ENV=production
      - PORT=8080
      - MAX_MEMORY_MB=4096
      - MAX_EXECUTION_TIME_MS=300000
      - ORCHESTRATOR_URL=${ORCHESTRATOR_URL}
      - API_SHARED_SECRET=${API_SHARED_SECRET}
    volumes:
      - runtime_nodejs_data:/tmp/executions
    healthcheck:
      test: ["CMD", "wget", "-q", "--spider", "http://localhost:8080/health"]
      interval: 30s
      timeout: 10s
      retries: 3
    restart: unless-stopped
    networks:
      - functionfly-runtime
    deploy:
      resources:
        limits:
          memory: 8G
          cpus: '4.0'
        reservations:
          memory: 4G
          cpus: '2.0'

  # Python Runtime
  runtime-python:
    image: functionfly/runtime-python:latest
    container_name: functionfly-runtime-python
    hostname: runtime-python
    ports:
      - "127.0.0.1:8083:8080"
    environment:
      - PYTHON_ENV=production
      - PORT=8080
      - MAX_MEMORY_MB=4096
      - MAX_EXECUTION_TIME_MS=300000
      - ORCHESTRATOR_URL=${ORCHESTRATOR_URL}
      - API_SHARED_SECRET=${API_SHARED_SECRET}
    volumes:
      - runtime_python_data:/tmp/executions
    healthcheck:
      test: ["CMD", "wget", "-q", "--spider", "http://localhost:8080/health"]
      interval: 30s
      timeout: 10s
      retries: 3
    restart: unless-stopped
    networks:
      - functionfly-runtime
    deploy:
      resources:
        limits:
          memory: 8G
          cpus: '4.0'
        reservations:
          memory: 4G
          cpus: '2.0'

  # ClamAV Antivirus Scanner
  clamav:
    image: clamav/clamav:latest
    container_name: functionfly-clamav
    hostname: clamav
    ports:
      - "127.0.0.1:3310:3310"
    volumes:
      - clamav_data:/var/lib/clamav
    environment:
      - CLAMD_STARTUP_TIMEOUT=600
      - FRESHCLAM_CHECKS=24
    healthcheck:
      test: ["CMD", "clamdscan", "--ping", "1"]
      interval: 60s
      timeout: 10s
      retries: 3
      start_period: 120s
    restart: unless-stopped
    networks:
      - functionfly-runtime
    deploy:
      resources:
        limits:
          memory: 2G
          cpus: '2.0'
        reservations:
          memory: 1G
          cpus: '1.0'

  # YARA Malware Scanner
  yara:
    build:
      context: ./yara
      dockerfile: Dockerfile
    container_name: functionfly-yara
    hostname: yara
    ports:
      - "127.0.0.1:8084:8080"
    volumes:
      - ./yara-rules:/rules:ro
      - yara_data:/data
    environment:
      - PORT=8080
      - RULES_PATH=/rules
      - MAX_FILE_SIZE_MB=50
    healthcheck:
      test: ["CMD", "wget", "-q", "--spider", "http://localhost:8080/health"]
      interval: 30s
      timeout: 10s
      retries: 3
    restart: unless-stopped
    networks:
      - functionfly-runtime
    deploy:
      resources:
        limits:
          memory: 2G
          cpus: '2.0'
        reservations:
          memory: 1G
          cpus: '1.0'

volumes:
  runtime_local_data:
    driver: local
    name: functionfly_runtime_local_data
  runtime_nodejs_data:
    driver: local
    name: functionfly_runtime_nodejs_data
  runtime_python_data:
    driver: local
    name: functionfly_runtime_python_data
  clamav_data:
    driver: local
    name: functionfly_clamav_data
  yara_data:
    driver: local
    name: functionfly_yara_data

networks:
  functionfly-runtime:
    driver: bridge
    name: functionfly-runtime
    ipam:
      config:
        - subnet: 172.31.0.0/16
```

### 4.4 YARA Configuration

Create `/opt/functionfly/yara/Dockerfile`:

```dockerfile
FROM python:3.11-slim

# Install dependencies
RUN apt-get update && apt-get install -y \
    yara \
    libyara-dev \
    gcc \
    && rm -rf /var/lib/apt/lists/*

# Install Python dependencies
COPY requirements.txt .
RUN pip install --no-cache-dir -r requirements.txt

# Copy service code
COPY yara_service.py /app/

# Create rules directory
RUN mkdir -p /rules

WORKDIR /app

EXPOSE 8080

CMD ["python", "yara_service.py"]
```

Create `/opt/functionfly/yara/requirements.txt`:

```
flask==3.0.0
gunicorn==21.2.0
yara-python==4.3.1
requests==2.31.0
```

Create `/opt/functionfly/yara/yara_service.py`:

```python
#!/usr/bin/env python3
"""
YARA Scanner Service for FunctionFly
Scans uploaded files and code for malicious patterns
"""

import os
import sys
import json
import hashlib
import tempfile
from flask import Flask, request, jsonify
import yara
import logging

# Configure logging
logging.basicConfig(level=logging.INFO)
logger = logging.getLogger(__name__)

app = Flask(__name__)

# Configuration
RULES_PATH = os.environ.get('RULES_PATH', '/rules')
MAX_FILE_SIZE_MB = int(os.environ.get('MAX_FILE_SIZE_MB', '50'))
MAX_FILE_SIZE_BYTES = MAX_FILE_SIZE_MB * 1024 * 1024

# Compile YARA rules
rules = None

def load_rules():
    """Load and compile YARA rules from rules directory"""
    global rules
    try:
        # Check if rules directory exists and has rule files
        if os.path.exists(RULES_PATH):
            rule_files = [f for f in os.listdir(RULES_PATH) if f.endswith('.yar') or f.endswith('.yara')]
            if rule_files:
                rules = yara.compile(filepath=os.path.join(RULES_PATH, rule_files[0]))
                for rule_file in rule_files[1:]:
                    rules = yara.compile(
                        filepath=os.path.join(RULES_PATH, rule_file),
                        externals=rules
                    )
                logger.info(f"Loaded {len(rule_files)} YARA rules")
            else:
                # Create a default rule if none exist
                default_rule = '''
                rule default_malware_check {
                    strings:
                        $mz = { 4D 5A }
                        $elf = { 7F 45 4C 46 }
                    condition:
                        $mz at 0 or $elf at 0
                }
                '''
                rules = yara.compile(source=default_rule)
                logger.info("Using default YARA rule")
        else:
            logger.warning(f"Rules path {RULES_PATH} does not exist")
            rules = yara.compile(source='rule dummy { condition: false }')
    except Exception as e:
        logger.error(f"Error loading YARA rules: {e}")
        rules = yara.compile(source='rule dummy { condition: false }')

@app.route('/health', methods=['GET'])
def health():
    """Health check endpoint"""
    return jsonify({'status': 'healthy', 'service': 'yara-scanner'})

@app.route('/scan', methods=['POST'])
def scan():
    """Scan file for malware signatures"""
    try:
        # Check if file is provided
        if 'file' not in request.files:
            return jsonify({'error': 'No file provided'}), 400
        
        file = request.files['file']
        
        # Check file size
        file.seek(0, os.SEEK_END)
        file_size = file.tell()
        file.seek(0)
        
        if file_size > MAX_FILE_SIZE_BYTES:
            return jsonify({
                'error': f'File too large. Max size: {MAX_FILE_SIZE_MB}MB'
            }), 413
        
        # Save to temp file and scan
        with tempfile.NamedTemporaryFile(delete=False) as tmp:
            file.save(tmp)
            tmp_path = tmp.name
        
        try:
            # Calculate file hash
            sha256_hash = hashlib.sha256()
            with open(tmp_path, 'rb') as f:
                for chunk in iter(lambda: f.read(4096), b''):
                    sha256_hash.update(chunk)
            file_hash = sha256_hash.hexdigest()
            
            # Run YARA scan
            matches = rules.match(tmp_path)
            
            result = {
                'file_hash': file_hash,
                'scan_result': 'malicious' if matches else 'clean',
                'matches': [{'rule': m.rule, 'tags': m.tags} for m in matches],
                'rules_matched': len(matches)
            }
            
            return jsonify(result)
            
        finally:
            os.unlink(tmp_path)
            
    except Exception as e:
        logger.error(f"Scan error: {e}")
        return jsonify({'error': str(e)}), 500

@app.route('/scan/code', methods=['POST'])
def scan_code():
    """Scan code string for suspicious patterns"""
    try:
        data = request.get_json()
        if not data or 'code' not in data:
            return jsonify({'error': 'No code provided'}), 400
        
        code = data['code']
        
        # Check code size
        if len(code.encode('utf-8')) > MAX_FILE_SIZE_BYTES:
            return jsonify({
                'error': f'Code too large. Max size: {MAX_FILE_SIZE_MB}MB'
            }), 413
        
        # Create temp file with code
        with tempfile.NamedTemporaryFile(mode='w', suffix='.txt', delete=False) as tmp:
            tmp.write(code)
            tmp_path = tmp.name
        
        try:
            # Run YARA scan
            matches = rules.match(tmp_path)
            
            result = {
                'scan_result': 'suspicious' if matches else 'clean',
                'matches': [{'rule': m.rule, 'tags': m.tags} for m in matches],
                'rules_matched': len(matches)
            }
            
            return jsonify(result)
            
        finally:
            os.unlink(tmp_path)
            
    except Exception as e:
        logger.error(f"Code scan error: {e}")
        return jsonify({'error': str(e)}), 500

if __name__ == '__main__':
    load_rules()
    port = int(os.environ.get('PORT', '8080'))
    app.run(host='0.0.0.0', port=port)
```

### 4.5 Deploy Server 2

```bash
# Navigate to functionfly directory
cd /opt/functionfly

# Create YARA directory structure
mkdir -p yara

# Copy Dockerfile and service files (after creating them)
# ... copy files to yara/ directory ...

# Pull images
docker compose pull

# Build YARA service
docker compose build yara

# Start services
docker compose up -d

# Check status
docker compose ps

# View logs
docker compose logs -f
```

---

## Step 5: Neon Database Setup

### 5.1 Create Neon Account

1. Visit [neon.tech](https://neon.tech)
2. Sign up with your email or GitHub account
3. Verify your email address

### 5.2 Create Production Project

1. Click **New Project**
2. Project Name: `functionfly-production`
3. PostgreSQL Version: **16** (or latest)
4. Region: Select closest to your OVHcloud servers (e.g., `US East`)
5. Click **Create Project**

### 5.3 Configure Database

1. Once created, click on the project
2. Go to **SQL Editor**
3. Run the initial setup:

```sql
-- Create application user
CREATE USER functionfly_app WITH PASSWORD 'your-secure-password-here';

-- Create database
CREATE DATABASE functionfly OWNER functionfly_app;

-- Grant permissions
GRANT ALL PRIVILEGES ON DATABASE functionfly TO functionfly_app;

-- Connect to the database
\c functionfly

-- Enable required extensions (or run migration 000000_postgres_extensions.up.sql)
CREATE EXTENSION IF NOT EXISTS "uuid-ossp";
CREATE EXTENSION IF NOT EXISTS "pg_trgm";
CREATE EXTENSION IF NOT EXISTS "pgvector";
CREATE EXTENSION IF NOT EXISTS "pgcrypto";
```

### 5.4 Get Connection String

1. Go to **Connection Details**
2. Copy the connection string
3. It will look like:
   ```
   postgresql://functionfly_owner:password@ep-production-xxxxx.us-east-1.aws.neon.tech/functionfly?sslmode=require
   ```

### 5.5 Update Server 1 Configuration

On Server 1, update the `.env` file:

```bash
# Edit environment file
nano /opt/functionfly/.env

# Update database connection variables
DB_HOST=ep-production-xxxxx.us-east-1.aws.neon.tech
DB_PORT=5432
DB_USER=functionfly_owner
DB_PASSWORD=your-actual-neon-password
DB_NAME=functionfly
DB_SSLMODE=require
```

### 5.6 Run Migrations

```bash
cd /opt/functionfly

# Restart orchestrator to pick up new database config
docker compose restart orchestrator-api

# Run migrations (adjust based on your migration system)
docker compose exec orchestrator-api npm run migrate

# Or if using Go migrations
docker compose exec orchestrator-api ./migrate up
```

### 5.7 Verify Database Connection

```bash
# Check logs for successful connection
docker compose logs orchestrator-api | grep -i "database\|connected"

# Test from command line
docker compose exec orchestrator-api psql $DATABASE_URL -c "SELECT 1;"
```

---

## Step 6: Verification

### 6.1 Container Health Checks

**On Server 1:**

```bash
cd /opt/functionfly

# Check all containers are running
docker compose ps

# Expected output:
# NAME                      STATUS          PORTS
# functionfly-postgres      Up 5 minutes    127.0.0.1:5432->5432/tcp
# functionfly-redis         Up 5 minutes    127.0.0.1:6379->6379/tcp
# functionfly-orchestrator  Up 5 minutes    127.0.0.1:8080->8080/tcp
# functionfly-caddy         Up 5 minutes    0.0.0.0:80->80/tcp, 0.0.0.0:443->443/tcp
```

**On Server 2:**

```bash
cd /opt/functionfly

# Check all containers are running
docker compose ps

# Expected output:
# NAME                          STATUS          PORTS
# functionfly-runtime-local     Up 5 minutes    127.0.0.1:8081->8080/tcp
# functionfly-runtime-nodejs    Up 5 minutes    127.0.0.1:8082->8080/tcp
# functionfly-runtime-python    Up 5 minutes    127.0.0.1:8083->8080/tcp
# functionfly-clamav            Up 5 minutes    127.0.0.1:3310->3310/tcp
# functionfly-yara              Up 5 minutes    127.0.0.1:8084->8080/tcp
```

### 6.2 API Health Checks

```bash
# Test main domain
curl -s https://functionfly.com/health
echo "OK"

# Test API endpoint
curl -s https://api.functionfly.com/healthz
echo "OK"

# Test with full response
curl -s https://api.functionfly.com/v1/status | jq .
```

### 6.3 Database Connectivity

```bash
# On Server 1 - test database connection
cd /opt/functionfly
docker compose exec postgres psql -U functionfly -c "SELECT version();"

# Test from orchestrator container
docker compose exec orchestrator-api wget -qO- http://localhost:8080/healthz
```

### 6.4 Runtime Tests

```bash
# From Server 1, test connectivity to Server 2 runtimes
curl -s http://<server2-ip>:8081/health
curl -s http://<server2-ip>:8082/health
curl -s http://<server2-ip>:8083/health

# Test ClamAV
curl -s http://<server2-ip>:3310

# Test YARA
curl -s http://<server2-ip>:8084/health
```

### 6.5 End-to-End Function Test

```bash
# Create a simple test function
curl -X POST https://api.functionfly.com/v1/functions \
  -H "Authorization: Bearer YOUR_API_TOKEN" \
  -H "Content-Type: application/json" \
  -d '{
    "name": "hello-world",
    "runtime": "local",
    "code": "export default async function(req) { return { body: \"Hello World!\" }; }"
  }'

# Execute the function
curl -X POST https://api.functionfly.com/v1/execute/hello-world \
  -H "Authorization: Bearer YOUR_API_TOKEN" \
  -H "Content-Type: application/json" \
  -d '{"name": "Test"}'
```

### 6.6 SSL/TLS Verification

```bash
# Check SSL certificate
echo | openssl s_client -servername functionfly.com -connect functionfly.com:443 2>/dev/null | openssl x509 -noout -dates -subject

# Test HTTPS redirection
curl -I -L http://functionfly.com

# Should redirect to HTTPS
```

### 6.7 Monitoring Setup

```bash
# Check container resource usage
docker stats --no-stream

# View recent logs
docker compose logs --tail=100

# Check disk usage
df -h

# Check memory usage
free -h
```

---

## 📈 Scaling Roadmap

### Phase 1: Launch (Current)
**Cost: $65/month | Capacity: ~100 concurrent users**

- **2x OVHcloud KS-5** servers ($40)
- **Neon PostgreSQL** managed ($25)
- Handles launch traffic and initial user onboarding

### Phase 2: 500 Active Users
**Cost: ~$100/month | Target: 6 months**

```bash
# Add more runtime containers on Server 2
docker compose up -d --scale runtime-nodejs=3 --scale runtime-python=3

# Scale Neon database
# - Upgrade to next tier in Neon dashboard
# - Enable connection pooling
```

**Changes:**
- Scale runtime containers horizontally
- Add PgBouncer connection pooler
- Enable read replicas on Neon

### Phase 3: 1,000 Active Users
**Cost: ~$150/month | Target: 12 months**

```bash
# Upgrade servers via OVHcloud control panel
# KS-5 → KS-6 (128GB RAM, 2x1TB SSD)
```

**Infrastructure:**
- **2x OVHcloud KS-6** servers ($80)
- **Neon Scale** tier ($50+)
- **Cloudflare Pro** ($20) - for advanced DDoS

### Phase 4: 2,000 Active Users
**Cost: ~$250/month | Target: 18 months**

```
┌─────────────────────────────────────────────────────────────────┐
│                    Phase 4 Architecture                         │
├─────────────────────────────────────────────────────────────────┤
│                                                                 │
│  ┌──────────────────┐    Load Balancer (Caddy/HAProxy)        │
│  │   Cloudflare     │◄───────────────────────────────────┐     │
│  │   (CDN/DNS)      │                                    │     │
│  └────────┬─────────┘                                    │     │
│           │                                              │     │
│           ▼                                              │     │
│  ┌──────────────────────────────────────────────────────┐│     │
│  │              App Server Pool                         ││     │
│  │  ┌──────────────┐    ┌──────────────┐               ││     │
│  │  │  Server 1A   │    │  Server 1B   │               ││     │
│  │  │  (KS-6)      │◄──►│  (KS-6)      │               ││     │
│  │  │  API + Caddy │    │  API + Caddy │               ││     │
│  │  └──────────────┘    └──────────────┘               ││     │
│  └──────────────────────────────────────────────────────┘│     │
│                          │                               │     │
│           ┌──────────────┴──────────────┐               │     │
│           ▼                             ▼               │     │
│  ┌──────────────────┐          ┌──────────────────┐    │     │
│  │  Runtime Pool    │          │  Neon Database   │    │     │
│  │  (3x Servers)    │          │  (Scale Tier)    │    │     │
│  └──────────────────┘          └──────────────────┘    │     │
│                                                        │     │
└────────────────────────────────────────────────────────┴─────┘
```

**Changes:**
- Add 2nd app server (Server 1B)
- Deploy load balancer (HAProxy or additional Caddy)
- Expand runtime pool to 3+ servers
- Upgrade Neon to Scale tier

### Phase 5: 5,000+ Active Users
**Cost: ~$500+/month | Target: 24+ months**

- **Kubernetes cluster** (OVHcloud Managed K8s or self-managed)
- **Neon Enterprise** tier
- **Multi-region** deployment
- **Dedicated cache cluster** (Redis Cluster)

---

## 🔧 Troubleshooting

### Database Connection Issues

```bash
# Test Neon connectivity from Server 1
psql "postgresql://user:pass@host.neon.tech/db?sslmode=require" -c "SELECT 1;"

# Check if SSL is required
docker compose logs orchestrator-api | grep -i "ssl\|tls\|certificate"

# Verify firewall rules
ufw status verbose | grep 5432
```

### Runtime Connection Issues

```bash
# Test from Server 1 to Server 2
curl -v http://<server2-ip>:8081/health

# Check UFW rules on Server 2
ssh root@<server2-ip> "ufw status"

# Verify runtime containers are running
ssh functionfly@<server2-ip> "docker compose ps"
```

### SSL Certificate Issues

```bash
# Check certificate status
docker compose exec caddy caddy list-modules | grep tls

# Force certificate renewal
docker compose exec caddy caddy reload --config /etc/caddy/Caddyfile

# Check Caddy logs
docker compose logs caddy | grep -i "cert\|tls\|ssl"
```

### Container Won't Start

```bash
# Check logs
docker compose logs <service-name>

# Check for port conflicts
netstat -tlnp | grep :80
netstat -tlnp | grep :443

# Verify environment variables
docker compose config

# Restart with cleanup
docker compose down
docker compose up -d
```

### High Memory Usage

```bash
# Check container stats
docker stats --no-stream

# Identify memory-heavy containers
docker system df -v

# Restart specific service
docker compose restart <service-name>

# Check for memory leaks in logs
docker compose logs <service-name> | grep -i "memory\|oom"
```

---

## 📚 Additional Resources

- [OVHcloud Bare Metal Documentation](https://docs.ovh.com/us/en/dedicated/)
- [Neon PostgreSQL Documentation](https://neon.tech/docs)
- [Caddy Documentation](https://caddyserver.com/docs)
- [Docker Compose Documentation](https://docs.docker.com/compose/)
- [Cloudflare DNS Documentation](https://developers.cloudflare.com/dns/)

---

## ✅ Post-Deployment Checklist

- [ ] Servers provisioned and accessible via SSH
- [ ] Docker installed and running on both servers
- [ ] UFW firewall configured on both servers
- [ ] DNS A records created and propagated
- [ ] SSL certificates active (automatic via Caddy)
- [ ] Database migrations completed successfully
- [ ] All containers healthy (`docker compose ps`)
- [ ] API health checks passing
- [ ] Runtime services responding
- [ ] End-to-end function execution working
- [ ] Monitoring/logging configured
- [ ] Backup strategy implemented
- [ ] Documentation updated with actual values

---

## Domain split rollout checklist (app / admin / api)

When deploying with canonical subdomains (`app.functionfly.com`, `admin.functionfly.com`, `api.functionfly.com`):

1. **Environment**
   - [ ] `BASE_URL=https://api.functionfly.com` (or staging equivalent)
   - [ ] `FRONTEND_URL=https://app.functionfly.com` (or staging equivalent)
   - [ ] `CORS_ALLOWED_ORIGINS` includes `https://app.functionfly.com`, `https://admin.functionfly.com`, `https://api.functionfly.com` (no `*` in production)

2. **Staging first**
   - [ ] Deploy to staging; set `app.staging`, `admin.staging`, `api.staging` DNS and env
   - [ ] Verify login/OAuth and API calls from `app.staging` to `api.staging`
   - [ ] Verify CORS and WebSocket from app origin to API origin

3. **Production**
   - [ ] DNS records for `app`, `admin`, `api` (and optional 301 from legacy `account`/`dashboard` to `app`)
   - [ ] Dashboard build uses `VITE_API_URL=https://api.functionfly.com` for production
   - [ ] No hard-coded localhost fallbacks in production frontend

4. **Validation**
   - [ ] `go build ./...` and dashboard `npm run build` succeed
   - [ ] Backend rejects cross-origin requests when `CORS_ALLOWED_ORIGINS` is empty in production

---

*Last Updated: March 2026 | FunctionFly v1.0*

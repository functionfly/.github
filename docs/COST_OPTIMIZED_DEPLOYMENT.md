# Cost-Optimized Multi-Region Deployment Guide

## Free/Low-Cost Alternatives

### Architecture Options by Budget

| Tier | Monthly Cost | Setup | Features |
|------|-------------|-------|----------|
| **Free** | $0 | Manual backup + single region | Basic recovery |
| **Starter** | $5-10/mo | Single region + cold standby | Automated backups |
| **Standard** | $20-30/mo | 2 regions (active-passive) | Auto-failover |
| **Pro** | $50-100/mo | 3 regions (active-active) | Full geo-DNS |

---

## Free Tier: $0/Month

### Single Region with Manual Recovery

**Setup:**

- Deploy to single Fly.io region (iad)
- Use local PostgreSQL or free Neon tier (0.5GB)
- Daily manual backups to S3 (first 1GB free)

**Recovery:**

- Restore from backup - ~30 min RTO
- Manual DNS update if region fails

**Scripts:**

```bash
# Manual backup
./scripts/backup-database.sh production full

# Manual restore
./scripts/restore-database.sh --backup-file <file>
```

---

## Starter Tier: $5-10/Month

### Single Region with Automated Backups

**Setup:**

- Fly.io: 1x shared CPU ($5/mo)
- Neon Free Tier: 0.5GB database (free)
- S3: First 1GB free, ~$0.50/mo for backups

**Features:**

- Daily automated backups via cron
- Backup verification
- Point-in-time recovery (PITR) if using Neon

**Cost Breakdown:**

| Service | Cost |
|---------|------|
| Fly.io (1x shared) | $5/mo |
| Neon Free | $0 |
| S3 (backups) | ~$1/mo |
| **Total** | **$6/mo** |

---

## Standard Tier: $20-30/Month

### Two Regions (Active-Passive)

**Setup:**

- Primary: Fly.io iad ($5/mo)
- Standby: Fly.io lax ($5/mo) - scaled to 0 when not needed
- Neon Pro ($19/mo) with read replica OR
- Supabase Pro ($25/mo) with replica

**Features:**

- Manual failover (activate standby)
- Cross-region backup replication
- ~15 min RTO

**Cost Breakdown:**

| Service | Cost |
|---------|------|
| Fly.io primary | $5/mo |
| Fly.io standby (on-demand) | ~$2/mo |
| Neon Pro + replica | $19/mo |
| S3 replication | ~$2/mo |
| **Total** | **$28/mo** |

---

## Pro Tier: $50-100/Month

### Three Regions (Active-Active)

This is the full implementation with:

- 3 Fly.io regions ($15/mo)
- Neon Pro with geo-replication ($25/mo)
- Cloudflare Pro ($20/mo)
- Full automation

---

## Cheapest Viable Multi-Region Setup: ~$8/mo

### Configuration

1. **Primary Region (iad)** - Fly.io ($5/mo)
   - Shared CPU, 1GB volume
   - Auto-scaling: 1-2 instances

2. **Database** - Neon Free ($0)
   - 0.5GB storage
   - Manual promotion if needed

3. **Backups** - R2 ($0)
   - Cloudflare R2: 1GB free, no egress fees
   - Or S3: First 1GB free

4. **DNS** - Cloudflare Free ($0)
   - Basic DNS
   - Health checks via scripts

### Deployment

```bash
# Deploy primary
flyctl deploy --config deploy/fly/functionfly-control/fly.toml

# Activate standby region (if needed)
flyctl scale count 1 --region lax
flyctl regions add lax
```

---

## Recommendations by Use Case

| Use Case | Recommendation | Cost |
|----------|---------------|------|
| **Dev/Test** | Single region + manual | $0 |
| **Startup MVP** | Single region + auto-backups | $6/mo |
| **Production (budget)** | 2 regions, manual failover | $20/mo |
| **Production (enterprise)** | 3 regions, auto-failover | $60/mo |

---

## Cost Optimization Tips

1. **Use shared CPUs** - 75% cheaper than dedicated
2. **Scale to 0** - Standby regions cost nothing when not in use
3. **Use free tiers** - Neon free, Cloudflare R2, GitHub Actions
4. **Monitor usage** - Set up billing alerts
5. **Reserved instances** - Not needed for Fly.io (they optimize automatically)

---

## Migration Path

Start with Starter tier, upgrade when needed:

```
Starter ($6/mo) 
    ↓ (enable backup replication)
Standard ($28/mo)
    ↓ (activate third region)  
Pro ($60/mo)
```

Each tier upgrade takes ~1 hour to configure.

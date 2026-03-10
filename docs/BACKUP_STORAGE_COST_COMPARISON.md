# Database Backup Storage Cost Comparison

## Overview

This document compares the cost of different storage backends for database backups. All prices are approximate monthly costs for storing 100GB of compressed PostgreSQL backups with 30-day retention.

## Cost Comparison Table

| Storage Backend | Cost/GB/Month | 100GB/Month | Setup Complexity | Recommended For |
|----------------|---------------|-------------|------------------|-----------------|
| **AWS S3** | $0.023 | $2.30 | Low | Reference/benchmark |
| **Backblaze B2** | $0.005 | $0.50 | Medium | ✅ Most cost-effective |
| **Wasabi** | $0.040 | $4.00 | Low | Large backups |
| **DigitalOcean Spaces** | $0.020 | $2.00 | Low | DO users |
| **Google Cloud Storage** | $0.020 | $2.00 | Medium | GCP users |
| **SCP/SFTP** | ~$0.00 + VPS | $5-10 VPS | Medium | DIY enthusiasts |
| **MinIO (self-hosted)** | ~$0.00 + infra | $0-5 | High | Self-hosted |
| **Local storage** | $0.00 | $0.00 | Low | Small deployments |

## Detailed Analysis

### 🏆 **Backblaze B2** - Most Cost Effective (~$0.50/month for 100GB)

**Pros:**

- 90% cheaper than AWS S3
- Designed specifically for backup storage
- Good durability (99.999999999% object durability)
- Generous free tier (10GB free)
- Pay only for storage used

**Cons:**

- Requires separate CLI tool installation
- Slightly more complex setup
- No built-in CDN

**Best for:** Cost-conscious deployments, backup-focused use cases

**Configuration:**

```bash
DB_BACKUP_STORAGE_BACKEND=b2
DB_BACKUP_B2_KEY_ID=your_key_id
DB_BACKUP_B2_KEY=your_key
DB_BACKUP_B2_BUCKET=functionfly-backups
```

### 🥈 **SCP/SFTP to VPS** - Cheapest with Infrastructure (~$5-10/month)

**Pros:**

- Extremely cheap (just VPS cost)
- Full control over data
- Can use any cheap VPS provider
- No storage limits

**Cons:**

- Requires another server management
- Manual failover if VPS goes down
- Network transfer costs

**Best for:** Users who already have or want server infrastructure

**Configuration:**

```bash
DB_BACKUP_STORAGE_BACKEND=scp
DB_BACKUP_SCP_HOST=backup.yourdomain.com
DB_BACKUP_SCP_USER=backup
DB_BACKUP_SCP_PATH=/var/backups/functionfly
DB_BACKUP_SCP_KEY_FILE=/path/to/ssh/key
```

### 🥉 **Wasabi** - S3-Compatible Alternative (~$4.00/month)

**Pros:**

- 80% cheaper than AWS S3
- S3-compatible API (drop-in replacement)
- No egress fees
- High durability

**Cons:**

- Still more expensive than B2
- Minimum storage commitment in some plans

**Best for:** Teams familiar with S3 API, larger backup sizes

**Configuration:**

```bash
DB_BACKUP_STORAGE_BACKEND=wasabi
DB_BACKUP_S3_ENDPOINT=https://s3.wasabisys.com
DB_BACKUP_S3_BUCKET=functionfly-backups
AWS_ACCESS_KEY_ID=your_wasabi_key
AWS_SECRET_ACCESS_KEY=your_wasabi_secret
```

## Cost Optimization Tips

### 1. **Compression**

- PostgreSQL backups are already compressed (--compress=9)
- Typical compression ratio: 70-90% size reduction
- 100GB database → 10-30GB backup

### 2. **Retention Policies**

- 30 days is usually sufficient for most applications
- Consider shorter retention for development environments
- Use tiered retention (daily→weekly→monthly)

### 3. **Backup Frequency**

- Daily backups for production
- Weekly for staging
- On-demand for development

### 4. **Data Deduplication**

- Consider tools like `restic` or `borgbackup` for deduplication
- Especially useful if backing up multiple similar databases

## Migration Guide

### From AWS S3 to Backblaze B2

1. **Install B2 CLI:**

```bash
pip3 install b2
# or
curl -s https://raw.githubusercontent.com/Backblaze/B2_Command_Line_Tool/master/b2 install
```

1. **Update configuration:**

```bash
# Before (S3)
DB_BACKUP_STORAGE_BACKEND=s3
DB_BACKUP_S3_BUCKET=my-bucket

# After (B2)
DB_BACKUP_STORAGE_BACKEND=b2
DB_BACKUP_B2_BUCKET=my-bucket
```

1. **Test backup:**

```bash
./scripts/backup-database.sh production
```

### From S3 to SCP

1. **Set up backup server:**

```bash
# On backup server
sudo mkdir -p /var/backups/functionfly
sudo useradd -m -s /bin/bash backup
sudo mkdir -p /home/backup/.ssh
# Add your public key to /home/backup/.ssh/authorized_keys
sudo chown -R backup:backup /home/backup/.ssh
sudo chown -R backup:backup /var/backups/functionfly
```

1. **Update configuration:**

```bash
DB_BACKUP_STORAGE_BACKEND=scp
DB_BACKUP_SCP_HOST=your-backup-server.com
DB_BACKUP_SCP_USER=backup
DB_BACKUP_SCP_PATH=/var/backups/functionfly
DB_BACKUP_SCP_KEY_FILE=/path/to/private/key
```

## Monitoring Backup Costs

### Track Storage Usage

```bash
# B2 usage
b2 get-account-info | jq .storageUsed

# S3 usage
aws s3 ls s3://bucket-name --recursive --summarize

# Local usage
du -sh /var/backups/functionfly/
```

### Cost Alerts

Consider setting up alerts for:

- Storage usage > 80% of allocated budget
- Backup failures
- Unusual backup size changes

## Recommendations by Use Case

### **Small Startup/Side Project**

- **Best:** Backblaze B2 or local storage
- **Budget:** <$1/month
- **Setup:** 30 minutes

### **Growing SaaS Company**

- **Best:** Backblaze B2 or Wasabi
- **Budget:** $5-20/month
- **Setup:** 1 hour

### **Enterprise with Compliance Needs**

- **Best:** Wasabi or DigitalOcean Spaces
- **Budget:** $20-50/month
- **Setup:** 2-4 hours

### **Self-Hosted/Infrastructure Heavy**

- **Best:** SCP/SFTP or MinIO
- **Budget:** $5-15/month (infrastructure)
- **Setup:** 4-8 hours

## Emergency Recovery

All storage backends support the same restore process:

```bash
# Download latest backup
./scripts/restore-database.sh production latest

# Or download specific backup
./scripts/restore-database.sh production 20241201_020000
```

The restore script automatically detects the storage backend and downloads the appropriate file.

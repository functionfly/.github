# FlyMind AI Service - Phase 4 Production Hardening Migrations

This directory contains database migrations for Phase 4 features.

## Migrations

### 001_moderation_logs.sql
- Creates `moderation_logs` table for audit trail
- Tracks all content moderation decisions
- Includes tenant isolation

### 002_cache_metrics.sql
- Creates `cache_metrics` table for tracking cache performance
- Records hit/miss rates over time
- Supports analytics

### 003_api_usage_tracking.sql
- Creates `api_usage_tracking` table
- Tracks API usage per tenant
- Supports rate limiting and cost tracking

### 004_api_keys.sql
- Creates `api_keys` table
- Stores API key information
- Supports key rotation and revocation

## Running Migrations

Run migrations using the CLI:
```bash
python scripts/init_db.py
```

Or apply directly with psql:
```bash
psql -d functionfly -f migrations/001_moderation_logs.sql
```

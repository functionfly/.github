# Object Storage (Application Uploads)

This document describes how to configure **application object storage** for file uploads (e.g. feedback attachments) in FunctionFly. Storage is provided by `StorageService` in `internal/services/storage.go` and supports local disk, AWS S3, Cloudflare R2, and any S3-compatible backend (MinIO, Backblaze B2, Wasabi, DigitalOcean Spaces) via a custom endpoint.

**Database backups use a separate configuration.** See [BACKUP_STORAGE_COST_COMPARISON.md](BACKUP_STORAGE_COST_COMPARISON.md) and `deploy/database/backup-config-examples.env` for backup storage options.

---

## Phases: local → cloud → scale

| Phase | Use case | Backend | Cost (ballpark) |
|-------|----------|---------|-----------------|
| **0 – Bootstrap** | Dev / single server | `local` | $0 |
| **1 – First cloud** | Prod, minimal spend | **Cloudflare R2** | 10 GB free, then ~$0.015/GB, $0 egress |
| **1-alt – Self-hosted / cheap cloud** | No cloud or B2 | **MinIO** or **Backblaze B2** (S3-compatible) | $0 (MinIO) or ~$0.005/GB (B2), 10 GB free |
| **2 – Scale** | More traffic/volume | Same R2 or move to S3/Wasabi | Same API; add CDN/lifecycle as needed |

- **R2** is the recommended first cloud step: no egress fees, 10 GB free, already supported with no code changes.
- **MinIO** and **B2** require setting `STORAGE_S3_ENDPOINT` (and optionally `STORAGE_PUBLIC_URL`); see [S3-compatible custom endpoint](#s3-compatible-custom-endpoint-minio-b2-wasabi-do-spaces) below.

---

## Environment variables

### All backends

| Variable | Required | Description |
|----------|----------|-------------|
| `STORAGE_BACKEND` | No (default: `local`) | `local`, `s3`, or `r2` |
| `STORAGE_BUCKET` | For `s3`/`r2` | Bucket name (e.g. `functionfly-uploads`) |

### Local (`STORAGE_BACKEND=local`)

No extra variables. Files are stored under `./uploads` (or the directory configured in code). Ensure this directory is on a persistent volume and included in backups if used in production.

### Cloudflare R2 (`STORAGE_BACKEND=r2`)

| Variable | Required | Description |
|----------|----------|-------------|
| `R2_ACCOUNT_ID` | Yes | Cloudflare account ID |
| `AWS_ACCESS_KEY_ID` | Yes | R2 API token access key |
| `AWS_SECRET_ACCESS_KEY` | Yes | R2 API token secret |
| `R2_ENDPOINT` | No | Override endpoint (usually auto-detected) |
| `R2_PUBLIC_URL` | No | Base URL for public object links (e.g. custom domain) |

### AWS S3 (`STORAGE_BACKEND=s3`)

| Variable | Required | Description |
|----------|----------|-------------|
| `AWS_REGION` | Yes (for standard S3) | e.g. `us-east-1` |
| `AWS_ACCESS_KEY_ID` | Yes | AWS access key |
| `AWS_SECRET_ACCESS_KEY` | Yes | AWS secret key |

### S3-compatible custom endpoint (MinIO, B2, Wasabi, DO Spaces)

Use `STORAGE_BACKEND=s3` and set:

| Variable | Required | Description |
|----------|----------|-------------|
| `STORAGE_S3_ENDPOINT` | Yes | Full endpoint URL (e.g. `https://s3.us-west-002.backblazeb2.com`, `http://minio:9000`) |
| `AWS_ACCESS_KEY_ID` | Yes | Provider access key |
| `AWS_SECRET_ACCESS_KEY` | Yes | Provider secret key |
| `STORAGE_PUBLIC_URL` | No | Base URL for public object links (CDN or bucket URL). If unset, URLs are built as `STORAGE_S3_ENDPOINT/bucket/key`. |

Path-style requests are used automatically when `STORAGE_S3_ENDPOINT` is set.

**Examples:**

- **MinIO:** `STORAGE_S3_ENDPOINT=http://minio:9000`, bucket created in MinIO, credentials from MinIO.
- **Backblaze B2:** `STORAGE_S3_ENDPOINT=https://s3.<region>.backblazeb2.com`, bucket and application key created in B2.
- **Wasabi / DO Spaces:** Use the provider’s S3-compatible endpoint and credentials.

---

## Scaling

- **Stay on R2:** Increase bucket usage as needed; add lifecycle rules or R2 public bucket + custom domain + CDN for read-heavy traffic.
- **Move to S3/Wasabi:** Migrate objects (e.g. with `rclone` or provider tools), then switch env to `s3` (and, for Wasabi, set `STORAGE_S3_ENDPOINT`). No change to application code.
- **Optional later:** Versioning, CORS, signed URLs for private uploads—all work with the current design.

---

## Summary

- **Budget option now:** Use `local` for $0, or **Cloudflare R2** for the first cloud step (no code change, free egress, 10 GB free).
- **Self-hosted or B2:** Set `STORAGE_S3_ENDPOINT` (and optionally `STORAGE_PUBLIC_URL`) with `STORAGE_BACKEND=s3`.
- **Backups:** Configured separately; see [BACKUP_STORAGE_COST_COMPARISON.md](BACKUP_STORAGE_COST_COMPARISON.md) and `deploy/database/backup-config-examples.env`.

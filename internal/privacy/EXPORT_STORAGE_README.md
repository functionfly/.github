# GDPR Export Storage Configuration

This module supports secure cloud storage for GDPR data exports using S3-compatible services (AWS S3, Cloudflare R2, MinIO, etc.).

## Quick Start

### Local Development (Default)
By default, exports are stored on the local filesystem at `./exports/`:

```bash
# No configuration needed - works out of the box
# Exports saved to: ./exports/gdpr-exports/{request_id}/
```

### Production: Cloudflare R2 (Recommended)
R2 offers zero egress fees, making it ideal for data exports:

```bash
export PRIVACY_EXPORT_BUCKET="privacy-exports"
export PRIVACY_EXPORT_REGION="auto"
export PRIVACY_EXPORT_ENDPOINT="https://<account_id>.r2.cloudflarestorage.com"
export PRIVACY_EXPORT_ACCESS_KEY_ID="your_r2_access_key"
export PRIVACY_EXPORT_SECRET_ACCESS_KEY="your_r2_secret_key"
export PRIVACY_EXPORT_BASE_URL="https://exports.functionfly.io"  # Optional: CDN or custom domain
```

### Production: AWS S3

```bash
export PRIVACY_EXPORT_BUCKET="myapp-privacy-exports"
export PRIVACY_EXPORT_REGION="us-east-1"
export PRIVACY_EXPORT_ACCESS_KEY_ID="AKIA..."
export PRIVACY_EXPORT_SECRET_ACCESS_KEY="secret..."
export PRIVACY_EXPORT_BASE_URL="https://exports.myapp.com"  # CloudFront or custom domain
```

### Production: MinIO (Self-hosted)

```bash
export PRIVACY_EXPORT_BUCKET="privacy-exports"
export PRIVACY_EXPORT_ENDPOINT="http://minio.example.com:9000"
export PRIVACY_EXPORT_ACCESS_KEY_ID="minioadmin"
export PRIVACY_EXPORT_SECRET_ACCESS_KEY="minioadmin"
export PRIVACY_EXPORT_PATH_STYLE="true"  # Required for MinIO
export PRIVACY_EXPORT_BASE_URL="http://minio.example.com:9000/privacy-exports"
```

## Configuration Reference

| Variable | Default | Required | Description |
|----------|---------|----------|-------------|
| `PRIVACY_EXPORT_BUCKET` | `privacy-exports` | For S3/R2 | Storage bucket name |
| `PRIVACY_EXPORT_REGION` | `us-east-1` | For AWS S3 | AWS region |
| `PRIVACY_EXPORT_ENDPOINT` | - | For R2/MinIO | S3-compatible endpoint URL |
| `PRIVACY_EXPORT_ACCESS_KEY_ID` | - | For S3/R2 | Access key |
| `PRIVACY_EXPORT_SECRET_ACCESS_KEY` | - | For S3/R2 | Secret key |
| `PRIVACY_EXPORT_BASE_URL` | - | Optional | Public URL for downloads (CDN/custom domain) |
| `PRIVACY_EXPORT_PATH_PREFIX` | `gdpr-exports` | Optional | Prefix for stored files |
| `PRIVACY_EXPORT_PATH_STYLE` | `false` | For MinIO | Use path-style addressing |
| `PRIVACY_EXPORT_LOCAL_PATH` | `./exports` | For local dev | Local storage path |
| `PRIVACY_EXPORT_LOCAL_URL` | - | Optional | Base URL for local files |

## Security Features

- **Pre-signed URLs**: Time-limited download URLs (configurable expiration)
- **Access tokens**: Secondary verification tokens for download authorization
- **Metadata tagging**: Files tagged with `privacy-purpose: gdpr-export`
- **Content hashing**: SHA-256 hash stored in metadata for integrity verification
- **Automatic cleanup**: Files deleted after download or expiration

## File Structure

### S3/R2 Storage
```
s3://privacy-exports/
  └── gdpr-exports/
      └── ab/
          └── abcd1234-...-uuid.zip  (request ID)
```

### Local Storage
```
./exports/
  └── gdpr-exports/
      └── ab/
          └── abcd1234-...-uuid.zip
```

## GDPR Compliance

1. **Secure transfer**: All uploads/downloads use HTTPS
2. **Time-limited**: Pre-signed URLs expire (default: 7 days)
3. **Access logging**: All operations logged with request ID
4. **Right to deletion**: Files automatically deleted after download or user request
5. **Encryption**: Files encrypted at rest by storage provider

## Testing

### Verify Storage Configuration
```bash
# Start the service - check logs for storage initialization
go run ./cmd/orchestrator-api

# Expected output with S3/R2 configured:
INFO[0000] Initializing S3-compatible storage for GDPR exports

# Expected output without S3 (local fallback):
INFO[0000] Using local filesystem storage for GDPR exports (set S3 env vars for production)
```

### Test Export Flow
1. Request data export via API
2. Service uploads ZIP to configured storage
3. User receives download URL
4. File served via pre-signed URL or CDN
5. File deleted after successful download

## Cost Considerations

| Provider | Storage | Egress | Best For |
|----------|---------|--------|----------|
| Cloudflare R2 | $0.015/GB/mo | Free | Production (no egress fees) |
| AWS S3 | $0.023/GB/mo | $0.09/GB | Production (with CloudFront) |
| MinIO | Self-hosted | Self-hosted | On-premise/air-gapped |
| Local FS | Disk cost | N/A | Development only |

## Troubleshooting

### "Failed to verify S3 bucket access"
- Check credentials are correct
- Verify bucket exists and is accessible
- Check IAM permissions (needs `s3:PutObject`, `s3:GetObject`, `s3:DeleteObject`, `s3:HeadBucket`)

### "Using local filesystem storage"
- Missing S3 credentials - set the environment variables
- Service falls back to local storage for safety

### Download URLs not working
- Check `PRIVACY_EXPORT_BASE_URL` is set correctly
- For pre-signed URLs, ensure clock is synchronized (AWS requires <15min skew)

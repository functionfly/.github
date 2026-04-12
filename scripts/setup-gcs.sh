#!/bin/bash
#
# Setup Google Cloud Storage (GCS) for FunctionFly
# Run this to configure GCS with your $300 GCP credits
#

set -e

APP_NAME="functionfly-orchestrator"
GCS_BUCKET="functionfly-backups"

# Colors
GREEN='\033[0;32m'
BLUE='\033[0;34m'
YELLOW='\033[1;33m'
NC='\033[0m'

log_info() { echo -e "${BLUE}[INFO]${NC} $1"; }
log_success() { echo -e "${GREEN}[SUCCESS]${NC} $1"; }
log_warn() { echo -e "${YELLOW}[WARN]${NC} $1"; }

echo "================================"
echo "GCS Setup for FunctionFly"
echo "================================"
echo ""

# Check prerequisites
if ! command -v gcloud &> /dev/null; then
    log_warn "gcloud CLI not found. Install from: https://cloud.google.com/sdk/docs/install"
    exit 1
fi

if ! command -v fly &> /dev/null; then
    log_warn "fly CLI not found. Install with: curl -L https://fly.io/install.sh | sh"
    exit 1
fi

# Step 1: Create GCS bucket
echo "Step 1: Creating GCS bucket..."
if gcloud storage buckets describe "gs://${GCS_BUCKET}" &>/dev/null; then
    log_success "Bucket ${GCS_BUCKET} already exists"
else
    log_info "Creating bucket: ${GCS_BUCKET}"
    gcloud storage buckets create "gs://${GCS_BUCKET}" --location=US --uniform-bucket-level-access
    log_success "Bucket created"
fi

# Step 2: Set lifecycle policy for cost management
echo ""
echo "Step 2: Setting lifecycle policy (14-day retention)..."
cat > /tmp/gcs-lifecycle.json <<'EOF'
{
  "lifecycle": {
    "rule": [
      {
        "action": {"type": "Delete"},
        "condition": {"age": 14}
      }
    ]
  }
}
EOF
gcloud storage buckets update "gs://${GCS_BUCKET}" --lifecycle-config=/tmp/gcs-lifecycle.json
log_success "Lifecycle policy set"

# Step 3: Create service account for Fly.io
echo ""
echo "Step 3: Creating service account..."
SA_NAME="functionfly-backup-sa"
SA_EMAIL="${SA_NAME}@$(gcloud config get-value project).iam.gserviceaccount.com"

if gcloud iam service-accounts describe "${SA_EMAIL}" &>/dev/null; then
    log_success "Service account already exists"
else
    gcloud iam service-accounts create "${SA_NAME}" \
        --display-name="FunctionFly Backup Service Account"
    log_success "Service account created"
fi

# Step 4: Grant permissions
echo ""
echo "Step 4: Granting GCS permissions..."
gcloud storage buckets add-iam-policy-binding "gs://${GCS_BUCKET}" \
    --member="serviceAccount:${SA_EMAIL}" \
    --role="roles/storage.objectAdmin"
log_success "Permissions granted"

# Step 5: Create and download key
echo ""
echo "Step 5: Creating service account key..."
KEY_FILE="/tmp/functionfly-gcs-key.json"
gcloud iam service-accounts keys create "${KEY_FILE}" \
    --iam-account="${SA_EMAIL}"
log_success "Key created: ${KEY_FILE}"

# Step 6: Encode key for Fly secrets
echo ""
echo "Step 6: Preparing for Fly.io..."
GCS_CREDS_B64=$(base64 -w 0 "${KEY_FILE}")
log_success "Key encoded (length: ${#GCS_CREDS_B64} chars)"

# Step 7: Set Fly secrets
echo ""
echo "Step 7: Setting Fly secrets..."
log_info "Setting GCS_BUCKET..."
fly secrets set --app "${APP_NAME}" GCS_BUCKET="${GCS_BUCKET}"

log_info "Setting GOOGLE_CREDENTIALS_B64..."
echo "${GCS_CREDS_B64}" | fly secrets set --app "${APP_NAME}" GOOGLE_CREDENTIALS_B64=-

log_info "Setting STORAGE_PROVIDER..."
fly secrets set --app "${APP_NAME}" STORAGE_PROVIDER="gcs"

log_success "All secrets set!"

# Cleanup
rm -f "${KEY_FILE}" /tmp/gcs-lifecycle.json

echo ""
echo "================================"
log_success "GCS Setup Complete!"
echo "================================"
echo ""
echo "Your $300 GCP credits will cover:"
echo "  - GCS storage: ~$0.020/GB/month"
echo "  - With 500GB: ~$10/month = ~30 months free"
echo "  - With 1TB:   ~$20/month = ~15 months free"
echo ""
echo "Test your backup:"
echo "  ./scripts/backup-gcs.sh backup"
echo ""
echo "List backups:"
echo "  ./scripts/backup-gcs.sh list"

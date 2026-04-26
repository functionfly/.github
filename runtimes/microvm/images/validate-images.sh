#!/usr/bin/env bash
# validate-images.sh — Verify MicroVM images are present and valid.
#
# Usage:
#   ./validate-images.sh [--image-path /path/to/images]
#
# Exit codes:
#   0  All images present and valid
#   1  Missing or invalid images

set -euo pipefail

IMAGE_PATH="${1:-/var/lib/functionfly/vmimages}"
SCRIPT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"

if [[ "$#" -gt 0 && "$1" == "--help" ]]; then
    echo "Usage: $0 [--image-path /path/to/images]"
    echo "  --image-path  Path to VM images directory (default: /var/lib/functionfly/vmimages)"
    exit 0
fi

if [[ "$1" == "--image-path" && -n "${2:-}" ]]; then
    IMAGE_PATH="$2"
fi

echo "==> Validating MicroVM images at: ${IMAGE_PATH}"
echo ""

MISSING=0

check_file() {
    local file="$1"
    local desc="$2"
    if [[ -f "${IMAGE_PATH}/${file}" ]]; then
        local size=$(du -sh "${IMAGE_PATH}/${file}" | cut -f1)
        echo "  [OK] ${file} (${size}) - ${desc}"
    else
        echo "  [MISSING] ${file} - ${desc}"
        MISSING=1
    fi
}

echo "Checking required files:"
echo ""
check_file "vmlinux" "Firecracker kernel (bzImage)"
check_file "python311.ext4" "Python 3.11 root filesystem"

echo ""

if [[ "${MISSING}" -eq 1 ]]; then
    echo "ERROR: Missing required images."
    echo ""
    echo "To fix, run:"
    echo "  cd ${SCRIPT_DIR}"
    echo "  sudo ./build-rootfs.sh --python 3.11 --out ${IMAGE_PATH}"
    echo "  sudo ./download-kernel.sh --out ${IMAGE_PATH}"
    exit 1
fi

echo "All images present."
echo ""
echo "Optional: verify rootfs checksum"
if [[ -f "${IMAGE_PATH}/python311.ext4.sha256" ]]; then
    if sha256sum --check "${IMAGE_PATH}/python311.ext4.sha256" 2>/dev/null; then
        echo "  [OK] python311.ext4 checksum verified"
    else
        echo "  [WARN] python311.ext4 checksum mismatch"
    fi
else
    echo "  [SKIP] No checksum file found"
fi

exit 0
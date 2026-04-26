#!/usr/bin/env bash
# download-kernel.sh — Obtain a Firecracker-compatible Linux kernel image.
#
# Usage:
#   ./download-kernel.sh [--out /path/to/output] [--arch amd64|arm64]
#
# Output:
#   <out>/vmlinux         (ELF kernel image for Firecracker x86_64)
#
set -euo pipefail

OUT_DIR=""
ARCH="amd64"
SCRIPT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"

usage() {
    echo "Usage: $0 [--out /path/to/output] [--arch amd64|arm64]"
    echo "  --out   output directory  (default: .)"
    echo "  --arch  amd64 or arm64   (default: amd64)"
    exit 1
}

while [[ $# -gt 0 ]]; do
    case "$1" in
        --out)
            OUT_DIR="$2"; shift 2;;
        --arch)
            ARCH="$2"; shift 2;;
        --help|-h)
            usage;;
        *)
            echo "Unknown option: $1"; usage;;
    esac
done

if [[ -z "${OUT_DIR}" ]]; then
    OUT_DIR="${SCRIPT_DIR}"
fi

case "${ARCH}" in
    amd64|x86_64) ARCH="amd64";;
    arm64|aarch64) ARCH="arm64";;
    *) echo "Unsupported architecture: ${ARCH}"; usage;;
esac

OUTPUT_VMLINUX="${OUT_DIR}/vmlinux"

mkdir -p "${OUT_DIR}"

echo "==> Firecracker MicroVM Kernel Setup"
echo "Target architecture: ${ARCH}"
echo "Output: ${OUTPUT_VMLINUX}"
echo ""

echo "Checking S3 for pre-built kernels..."
s3_found=false
for ver in "5.10" "6.1"; do
    url="https://s3.amazonaws.com/spec.ccfc.min/ci-arftifacts-pvh/ci-artifacts/kernels/x86_64/vmlinux-${ver}.bin"
    echo -n "  Trying ${ver}... "
    if curl -sfL "${url}" -o "${OUT_DIR}/vmlinux" 2>/dev/null; then
        echo "found!"
        chmod +x "${OUT_DIR}/vmlinux"
        echo "==> Kernel ready: ${OUTPUT_VMLINUX}"
        s3_found=true
        break
    else
        echo "not found"
    fi
done

if [[ "${s3_found}" == "true" ]]; then
    exit 0
fi

echo ""
echo "Pre-built kernels not found. Options:"

echo ""
echo "1. Copy host kernel (quickest):"
echo "   sudo cp /boot/vmlinuz-\$(uname -r) '${OUTPUT_VMLINUX}'"
echo ""
echo "2. Build with Firecracker tool:"
echo "   git clone --depth 1 --branch v1.15.1 \\"
echo "     https://github.com/firecracker-microvm/firecracker.git"
echo "   cd firecracker && ./tools/devtool build_ci_artifacts kernels"
echo "   cp resources/\$(uname -m)/vmlinux '${OUTPUT_VMLINUX}'"
echo ""
echo "3. Build from source:"
echo "   See: https://github.com/firecracker-microvm/firecracker/blob/main/docs/kernel.md"
exit 1
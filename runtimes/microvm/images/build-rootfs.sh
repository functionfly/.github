#!/usr/bin/env bash
# build-rootfs.sh — Build a CPython 3.11 ext4 root filesystem for Firecracker.
#
# Usage:
#   ./build-rootfs.sh [--python 3.11|3.12] [--out /path/to/output]
#
# Outputs:
#   <out>/python311.ext4   (or python312.ext4)
#   <out>/python311.ext4.sha256
#
# Requirements (on the build host):
#   - Docker (for building the rootfs layer)
#   - e2tools or mke2fs + debugfs (for packaging)
#   - root / sudo (for mount or debugfs population)
#
# The resulting ext4 image can be passed directly to Firecracker as a drive.
set -euo pipefail

PYTHON_VER="${1:-3.11}"
TAG="python${PYTHON_VER//./}"     # python311
OUT_DIR="${2:-$(pwd)}"
IMAGE="functionfly-microvm-${TAG}:build"
EXT4_FILE="${OUT_DIR}/${TAG}.ext4"
SIZE_MB="${MICROVM_ROOTFS_SIZE_MB:-2048}"

usage() {
    echo "Usage: $0 [PYTHON_VER] [OUT_DIR]"
    echo "  PYTHON_VER  3.11 or 3.12  (default: 3.11)"
    echo "  OUT_DIR     output directory  (default: .)"
    exit 1
}

[[ "${PYTHON_VER}" =~ ^3\.(11|12)$ ]] || { echo "Unsupported Python version: ${PYTHON_VER}"; usage; }

echo "==> Building Docker image for Python ${PYTHON_VER}..."
docker build \
    --platform linux/amd64 \
    -f "$(dirname "$0")/Dockerfile.${TAG}" \
    -t "${IMAGE}" \
    "$(dirname "$0")"

TMPDIR="$(mktemp -d)"
trap 'rm -rf "${TMPDIR}"' EXIT

echo "==> Exporting rootfs from container..."
CID="$(docker create "${IMAGE}" /bin/true)"
docker export "${CID}" | tar -x -C "${TMPDIR}"
docker rm "${CID}"

echo "==> Creating ${SIZE_MB}MB ext4 image at ${EXT4_FILE}..."
mkdir -p "${OUT_DIR}"
dd if=/dev/zero of="${EXT4_FILE}" bs=1M count="${SIZE_MB}" status=none
mkfs.ext4 -F -L "functionfly-rootfs" "${EXT4_FILE}" >/dev/null

echo "==> Populating ext4 image (requires root for mount)..."
MNT="$(mktemp -d)"
sudo mount -o loop "${EXT4_FILE}" "${MNT}"
sudo cp -a "${TMPDIR}/." "${MNT}/"
sudo umount "${MNT}"
rmdir "${MNT}"

echo "==> Writing checksum..."
sha256sum "${EXT4_FILE}" > "${EXT4_FILE}.sha256"

echo "==> Done: ${EXT4_FILE} ($(du -sh "${EXT4_FILE}" | cut -f1))"
echo "    Checksum: ${EXT4_FILE}.sha256"
echo ""
echo "    To use this rootfs with Firecracker:"
echo "      cp ${EXT4_FILE} /var/lib/functionfly/vmimages/${TAG}.ext4"
echo "      # Also copy a compatible vmlinux kernel to that directory."

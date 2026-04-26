#!/usr/bin/env bash
# build-rootfs.sh — Build a minimal ext4 root filesystem for Firecracker.
#
# Usage:
#   ./build-rootfs.sh [--python 3.11|3.12] [--out /path/to/output] [--type alpine|ubuntu]
#
# Outputs:
#   <out>/python311.ext4   (ext4 root filesystem)
#   <out>/python311.ext4.sha256
#
# Requirements:
#   - Docker
#   - root / sudo (for mounting ext4)
#
set -euo pipefail

PYTHON_VER="3.11"
TYPE="alpine"
OUT_DIR=""
SCRIPT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"

usage() {
    echo "Usage: $0 [--python 3.11|3.12] [--out /path/to/output] [--type alpine|ubuntu]"
    echo "  --python  Python version 3.11 or 3.12  (default: 3.11)"
    echo "  --out     output directory            (default: .)"
    echo "  --type    alpine (default, minimal) or ubuntu (full)"
    exit 1
}

while [[ $# -gt 0 ]]; do
    case "$1" in
        --python)
            PYTHON_VER="$2"; shift 2;;
        --out)
            OUT_DIR="$2"; shift 2;;
        --type)
            TYPE="$2"; shift 2;;
        --help|-h)
            usage;;
        *)
            echo "Unknown option: $1"; usage;;
    esac
done

if [[ -z "${OUT_DIR}" ]]; then
    OUT_DIR="${SCRIPT_DIR}"
fi

TAG="python${PYTHON_VER//./}"
IMAGE="functionfly-microvm-${TYPE}:build"
EXT4_FILE="${OUT_DIR}/${TAG}.ext4"
SIZE_MB="${MICROVM_ROOTFS_SIZE_MB:-512}"

[[ "${PYTHON_VER}" =~ ^3\.(11|12)$ ]] || { echo "Unsupported Python version: ${PYTHON_VER}"; usage; }
[[ "${TYPE}" =~ ^(alpine|ubuntu)$ ]] || { echo "Unsupported type: ${TYPE}"; usage; }

if ! command -v docker &>/dev/null; then
    echo "ERROR: Docker is required but not found."
    exit 1
fi

echo "==> Building Docker image (${TYPE} based)..."
case "${TYPE}" in
    alpine)
        DOCKERFILE="${SCRIPT_DIR}/Dockerfile.firecracker"
        IMAGE="functionfly-microvm-alpine:build"
        ;;
    ubuntu)
        DOCKERFILE="${SCRIPT_DIR}/Dockerfile.${TAG}"
        IMAGE="functionfly-microvm-${TAG}:build"
        ;;
esac

if [[ ! -f "${DOCKERFILE}" ]]; then
    echo "ERROR: Dockerfile not found: ${DOCKERFILE}"
    exit 1
fi

docker build \
    --platform linux/amd64 \
    -f "${DOCKERFILE}" \
    -t "${IMAGE}" \
    "${SCRIPT_DIR}"

TMPDIR="$(mktemp -d)"
trap 'rm -rf "${TMPDIR}"' EXIT

echo "==> Exporting rootfs from container..."
CID="$(docker create "${IMAGE}" /bin/true)"
docker export "${CID}" | tar -x -C "${TMPDIR}"
docker rm "${CID}"

echo "==> Creating ${SIZE_MB}MB ext4 image..."
mkdir -p "${OUT_DIR}"
dd if=/dev/zero of="${EXT4_FILE}" bs=1M count="${SIZE_MB}" status=none
mkfs.ext4 -F -L "functionfly-rootfs" "${EXT4_FILE}" >/dev/null

echo "==> Populating ext4 image..."
MNT="$(mktemp -d)"
sudo mount -o loop "${EXT4_FILE}" "${MNT}"
sudo cp -a "${TMPDIR}/." "${MNT}/"
sudo umount "${MNT}"
rmdir "${MNT}"

echo "==> Writing checksum..."
sha256sum "${EXT4_FILE}" > "${EXT4_FILE}.sha256"

echo "==> Done: ${EXT4_FILE} ($(du -sh "${EXT4_FILE}" | cut -f1))"
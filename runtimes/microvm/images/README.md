# MicroVM Images

This directory contains the build scripts and Dockerfiles for creating the root filesystem images used by the Firecracker MicroVM runtime.

## Quick Start

To set up MicroVM images on a deployment host:

```bash
# 1. Obtain a Firecracker-compatible Linux kernel (see Kernel Options below)
# 2. Build the root filesystem image (requires Docker and sudo)
sudo ./build-rootfs.sh --python 3.11

# 3. Verify the images are in place
ls -la *.ext4 vmlinux
```

## Files

| File | Description |
|------|-------------|
| `download-kernel.sh` | Attempts to download pre-built Firecracker kernel |
| `build-rootfs.sh` | Builds an ext4 rootfs with CPython and the function agent |
| `setup-network.sh` | Creates tap devices for Firecracker VMs |
| `validate-images.sh` | Validates required images are present |
| `Dockerfile.firecracker` | Alpine-based minimal rootfs (recommended) |
| `Dockerfile.python311` | Ubuntu-based Python 3.11 runtime |
| `Dockerfile.python312` | Ubuntu-based Python 3.12 runtime |
| `init.sh` | Minimal init script for Firecracker boot |
| `agent` | Python VM agent that runs inside the MicroVM |

## Image Requirements

The MicroVM orchestrator expects the following files in the image directory (default: `/var/lib/functionfly/vmimages`):

| File | Description |
|------|-------------|
| `vmlinux` | Linux kernel bzImage (Firecracker-compatible) |
| `python311.ext4` | Root filesystem with CPython 3.11 |
| `python311-alpine.ext4` | Alpine-based minimal rootfs (recommended) |

## Kernel Options

**Note:** Firecracker no longer hosts pre-built kernels on S3. The recommended approaches are:

### Option 1: Use Ubuntu's Generic Kernel (Quickest)

```bash
sudo apt install linux-image-generic
sudo cp /boot/vmlinuz-$(ls /boot/vmlinuz-*generic | head -1 | xargs basename) /var/lib/functionfly/vmimages/vmlinux
```

### Option 2: Build with Firecracker Tool (Recommended for production)

```bash
# Clone Firecracker and build kernels
git clone --depth 1 --branch v1.15.1 https://github.com/firecracker-microvm/firecracker.git
cd firecracker
./tools/devtool build_ci_artifacts kernels

# Copy the built kernel
cp resources/$(uname -m)/vmlinux /var/lib/functionfly/vmimages/
```

### Option 3: Build from Source

See [Firecracker Kernel Documentation](https://github.com/firecracker-microvm/firecracker/blob/main/docs/kernel.md):

```bash
git clone https://github.com/torvalds/linux.git
cd linux
git checkout v6.1
curl -sL https://raw.githubusercontent.com/firecracker-microvm/firecracker/main/resources/guest_configs/kconfig.6.1 > .config
make vmlinux -j$(nproc)
cp vmlinux /var/lib/functionfly/vmimages/
```

## Building Images

### Kernel Image

```bash
# Try the download script (may fail if S3 URLs are unavailable)
./download-kernel.sh --out /var/lib/functionfly/vmimages

# If that fails, use one of the kernel options above
```

### Root Filesystem Image

```bash
# Build Alpine-based rootfs (recommended - minimal, ~100MB)
sudo ./build-rootfs.sh --python 3.11 --type alpine --out /var/lib/functionfly/vmimages

# Or Ubuntu-based rootfs (larger, ~500MB)
sudo ./build-rootfs.sh --python 3.11 --type ubuntu --out /var/lib/functionfly/vmimages
```

Requirements:
- Docker
- sudo / root access (for mounting the ext4 image)
- e2fsprogs (mkfs.ext4)

## Network Setup

For Firecracker VMs to have network access, create tap devices:

```bash
# Create tap devices (run once at startup)
sudo ./setup-network.sh --create

# Or clean up and recreate
sudo ./setup-network.sh --cleanup
```

## Production Deployment

For full production deployment instructions, see [PRODUCTION.md](../PRODUCTION.md).

## Development

For local development with the MicroVM runtime:

```bash
# Create image directory
mkdir -p /tmp/microvm-images

# Obtain kernel (Ubuntu generic works for testing)
sudo apt install linux-image-generic
sudo cp /boot/vmlinuz-$(uname -r) /tmp/microvm-images/vmlinux

# Build rootfs (requires sudo for mount)
sudo ./build-rootfs.sh --python 3.11 --out /tmp/microvm-images

# Validate images
./validate-images.sh --image-path /tmp/microvm-images

# Run orchestrator with dev mode (uses host Python, no Firecracker)
FUNCTIONFLY_MICROVM_DEV_MODE=true cargo run -p functionfly-microvm -- --image-path /tmp/microvm-images
```

## Docker Build (Alternative)

You can also build the Docker image directly and export the filesystem:

```bash
# Build the Docker image
docker build -f Dockerfile.python311 -t functionfly-microvm-python311 .

# Export filesystem (alternative to build-rootfs.sh)
docker create functionfly-microvm-python311 /bin/true
docker export <container_id> | tar -x -C /tmp/rootfs
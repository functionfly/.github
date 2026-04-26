# FunctionFly MicroVM Production Deployment

This guide covers deploying the Firecracker MicroVM runtime for production use.

## Prerequisites

- Firecracker binary v1.15.1+
- Linux host with KVM support (or nested virtualization)
- Docker (for building rootfs images)
- Root access for network/tap device setup

## Installation Steps

### 1. Install Firecracker Binary

```bash
# Download Firecracker v1.15.1
curl -L https://github.com/firecracker-microvm/firecracker/releases/download/v1.15.1/firecracker-v1.15.1-x86_64.tgz | tar -xz
sudo mv firecracker-v1.15.1-x86_64/firecracker /usr/local/bin/
sudo chmod +x /usr/local/bin/firecracker
```

### 2. Create VM Images Directory

```bash
sudo mkdir -p /var/lib/functionfly/vmimages
sudo chown $(id -u):$(id -g) /var/lib/functionfly/vmimages
```

### 3. Download Kernel

```bash
cd /home/micro/projects/functionfly/runtimes/microvm/images
./download-kernel.sh --out /var/lib/functionfly/vmimages
```

Or manually:
```bash
curl -L https://s3.amazonaws.com/spec.ccfc.min/ci-arftifacts-pvh/ci-artifacts/kernels/x86_64/vmlinux-5.10.bin \
    -o /var/lib/functionfly/vmimages/vmlinux
chmod +x /var/lib/functionfly/vmimages/vmlinux
```

### 4. Build Rootfs

```bash
cd /home/micro/projects/functionfly/runtimes/microvm/images

# Build Alpine-based rootfs (recommended - minimal, fast boot)
sudo ./build-rootfs.sh --python 3.11 --type alpine --out /var/lib/functionfly/vmimages

# Or build Ubuntu-based rootfs (larger, more packages)
sudo ./build-rootfs.sh --python 3.11 --type ubuntu --out /var/lib/functionfly/vmimages
```

### 5. Set Up Network

```bash
cd /home/micro/projects/functionfly/runtimes/microvm/images

# Run as root to create tap devices
sudo ./setup-network.sh --create
```

### 6. Configure Environment

Create `/etc/functionfly/microvm.env`:

```bash
# Required for production
ENVIRONMENT=production
FUNCTIONFLY_MICROVM_API_TOKEN=your-secret-token-here

# Optional - tune for your workload
FUNCTIONFLY_MICROVM_MAX_VMS_PER_TENANT=10
FIRECRACKER_TAP_DEVICE=tap%d
TAP_BASE=tap
TAP_COUNT=10
```

### 7. Install Systemd Service

```bash
sudo cp /home/micro/projects/functionfly/runtimes/microvm/deploy/functionfly-microvm.service /etc/systemd/system/
sudo systemctl daemon-reload
sudo systemctl enable functionfly-microvm
sudo systemctl start functionfly-microvm
```

## Verification

```bash
# Check service status
sudo systemctl status functionfly-microvm

# Check logs
sudo journalctl -u functionfly-microvm -f

# Test health endpoint
curl http://localhost:9091/health

# Test metrics
curl http://localhost:9091/metrics
```

## Configuration Reference

### Environment Variables

| Variable | Default | Description |
|----------|---------|-------------|
| `ENVIRONMENT` | - | Set to `production` for production deployments |
| `FUNCTIONFLY_MICROVM_API_TOKEN` | - | Bearer token for API authentication |
| `FUNCTIONFLY_MICROVM_MAX_VMS_PER_TENANT` | 10 | Maximum VMs per tenant |
| `FIRECRACKER_BINARY` | `firecracker` | Path to Firecracker binary |
| `FIRECRACKER_TAP_DEVICE` | `tap%d` | Tap device pattern |
| `FIRECRACKER_SKIP_NETWORK` | false | Set to `true` for VMs without network |
| `TAP_BASE` | `tap` | Base name for tap devices |
| `TAP_COUNT` | 10 | Number of tap devices |

### Runtime Parameters

| Parameter | Default | Description |
|-----------|---------|-------------|
| `--vcpus` | 2 | vCPUs per VM |
| `--memory-mb` | 512 | Memory per VM in MB |
| `--max-vms` | 100 | Maximum concurrent VMs |
| `--port` | 9091 | HTTP API port |
| `--warm-idle-secs` | 600 | Idle time before VM eviction |
| `--cleanup-interval-secs` | 60 | Cleanup task interval |

## Kubernetes Deployment

For Kubernetes, deploy as a DaemonSet that runs on each node:

```yaml
apiVersion: apps/v1
kind: DaemonSet
metadata:
  name: functionfly-microvm
spec:
  selector:
    matchLabels:
      app: functionfly-microvm
  template:
    spec:
      initContainers:
      - name: setup-network
        command: ["/setup-network.sh", "--create"]
        securityContext:
          privileged: true
        volumeMounts:
        - name: scripts
          mountPath: /setup-network.sh
      containers:
      - name: orchestrator
        image: functionfly/orchestrator:latest
        env:
        - name: ENVIRONMENT
          value: "production"
        - name: FUNCTIONFLY_MICROVM_API_TOKEN
          valueFrom:
            secretKeyRef:
              name: functionfly-secrets
              key: microvm-api-token
        securityContext:
          capabilities:
            add: ["KVM", "NET_ADMIN"]
        volumeMounts:
        - name: vmimages
          mountPath: /var/lib/functionfly/vmimages
      volumes:
      - name: scripts
        configMap:
          defaultMode: 0755
      - name: vmimages
        hostPath:
          path: /var/lib/functionfly/vmimages
```

## Troubleshooting

### Firecracker won't start

Check if KVM is available:
```bash
ls -la /dev/kvm
```

### Network issues

Verify tap devices exist:
```bash
ip link show tap*
```

### VM won't boot

Check kernel and rootfs exist:
```bash
ls -la /var/lib/functionfly/vmimages/
```

Validate images:
```bash
./validate-images.sh --image-path /var/lib/functionfly/vmimages
```

### Performance tuning

For high-throughput workloads:
- Increase `--max-vms` if you have capacity
- Tune `--warm-idle-secs` for your workload patterns
- Consider dedicated CPU cores via `FIRECRACKER_VCPU_COUNT`
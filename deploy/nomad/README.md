# Nomad Deployment Guide

This directory contains Nomad job specifications and configuration for deploying FunctionFly services.

## Directory Structure

```
nomad/
├── config/
│   ├── nomad-server.hcl    # Server configuration
│   ├── nomad-client.hcl    # Client configuration
│   └── datacenters.hcl     # Datacenter definitions
├── jobs/
│   ├── orchestrator-api.nomad   # Main API service
│   ├── agent-service.nomad      # AI agent service
│   └── ai-gateway.nomad         # AI inference gateway (GPU)
├── systemd/
│   ├── nomad-server.service
│   └── nomad-client.service
├── deploy.sh               # Deploy all services
├── stop.sh                 # Stop all services
└── README.md
```

## Prerequisites

- Nomad 1.5+ installed on all nodes
- Docker installed on client nodes
- NVIDIA drivers and runtime for GPU nodes
- Consul (optional, for service discovery)

## Quick Start

### 1. Install Nomad

```bash
# Download Nomad
curl -fsSL https://releases.hashicorp.com/nomad/1.7.0/nomad_1.7.0_linux_amd64.zip -o nomad.zip
unzip nomad.zip
sudo mv nomad /usr/local/bin/nomad

# Verify installation
nomad version
```

### 2. Configure Nodes

On each **server** node:
```bash
sudo mkdir -p /etc/nomad.d /opt/nomad/data
sudo cp nomad/config/nomad-server.hcl /etc/nomad.d/
sudo cp nomad/config/datacenters.hcl /etc/nomad.d/
sudo cp nomad/systemd/nomad-server.service /etc/systemd/system/
sudo systemctl enable nomad-server
sudo systemctl start nomad-server
```

On each **client** node:
```bash
sudo mkdir -p /etc/nomad.d /opt/nomad/data
sudo mkdir -p /mnt/secrets/functionfly /mnt/config/agent /mnt/config/ai-gateway
sudo mkdir -p /mnt/storage/model-cache
sudo cp nomad/config/nomad-client.hcl /etc/nomad.d/
sudo cp nomad/config/datacenters.hcl /etc/nomad.d/
sudo cp nomad/systemd/nomad-client.service /etc/systemd/system/
sudo systemctl enable nomad-client
sudo systemctl start nomad-client
```

### 3. Deploy Services

```bash
# Set Nomad address
export NOMAD_ADDR=http://<server-ip>:4646

# Deploy all services
./deploy.sh

# Or deploy individual services
nomad job run jobs/orchestrator-api.nomad
nomad job run jobs/agent-service.nomad
nomad job run jobs/ai-gateway.nomad
```

### 4. Verify Deployment

```bash
# Check job status
./status.sh

# Or directly
nomad job status
nomad job status orchestrator-api
```

## Services

### Orchestrator API
- **Job**: `orchestrator-api`
- **Ports**: 8080 (HTTP)
- **Replicas**: 3
- **Resources**: 250 CPU, 256 MB memory
- **Health Check**: GET /health

### Agent Service
- **Job**: `agent-service`
- **Ports**: 8080 (HTTP), 9090 (metrics)
- **Replicas**: 3
- **Resources**: 500 CPU, 512 MB memory
- **Health Check**: GET /health

### AI Gateway
- **Job**: `ai-gateway`
- **Ports**: 8082 (HTTP), 9090 (metrics)
- **Replicas**: 3 (primary) + regional variants
- **Resources**: 1000 CPU, 4 GB memory + 1 GPU
- **Regions**: functionfly (primary), us-west-2, eu-west-1
- **Health Check**: GET /health

## GPU Configuration

For GPU support, ensure NVIDIA drivers and runtime are installed:

```bash
# Install NVIDIA drivers
wget https://download.nvidia.com/XFree86/Linux-x86_64/535.54.03/NVIDIA-Linux-x86_64-535.54.03.run
sudo sh NVIDIA-Linux-x86_64-535.54.03.run

# Install NVIDIA Container Toolkit
distribution=$(. /etc/os-release;echo $ID$VERSION_ID)
curl -s -L https://nvidia.github.io/nvidia-docker/gpgkey | sudo apt-key add -
curl -s -L https://nvidia.github.io/nvidia-docker/$distribution/nvidia-docker.list | sudo tee /etc/apt/sources.list.d/nvidia-docker.list
sudo apt-get update && sudo apt-get install -y nvidia-container-toolkit
sudo systemctl restart docker
```

Verify GPU access in Nomad:
```bash
nomad node status -filter 'GPU > 0'
```

## Operations

### Scaling

```bash
# Scale orchestrator-api to 5 replicas
nomad job scale orchestrator-api api 5

# Scale ai-gateway to 5 replicas
nomad job scale ai-gateway gateway 5
```

### Logs

```bash
# Follow logs for a job
nomad job logs -f orchestrator-api

# Follow logs for specific task
nomad job logs -f -task orchestrator-api orchestrator-api
```

### Rolling Updates

Nomad handles rolling updates automatically. To trigger a new deployment:

```bash
# Deploy new version
nomad job run jobs/orchestrator-api.nomad

# Or use the deploy script with update
./deploy.sh
```

### Drain Node

Before maintenance, drain a node to reschedule jobs:

```bash
# Mark node as ineligible for new work
nomad node drain <node-id>

# Or use the API
curl -X PUT http://nomad-server:4646/v1/node/<node-id>/drain
```

## Monitoring

Nomad exposes metrics at `/v1/metrics` in Prometheus format:

```bash
# View metrics
curl http://<nomad-server>:4646/v1/metrics?format=prometheus
```

## Troubleshooting

### Job Stuck in Pending

```bash
# Check allocation status
nomad alloc status <alloc-id>

# Check evaluation
nomad eval status <eval-id>

# Force a new evaluation
nomad job eval orchestrator-api
```

### GPU Not Available

```bash
# Verify NVIDIA devices on node
nomad node status <node-id> | grep -A 50 "Host Volume Stats"

# Check nvidia-smi on the node
ssh <node> nvidia-smi
```

### Health Check Failing

```bash
# Check task logs
nomad job logs <job-name>

# Check service registration
nomad job status -verbose <job-name>
```
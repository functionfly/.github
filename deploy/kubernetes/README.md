# Kubernetes Deployment Guide

This directory contains Kubernetes manifests and Helm charts for deploying FunctionFly services.

## Directory Structure

```
kubernetes/
├── ai-gateway/              # AI Gateway Helm chart
│   ├── Chart.yaml
│   ├── README.md
│   ├── values.yaml
│   ├── values.production.yaml
│   ├── values.staging.yaml
│   └── templates/
│       ├── deployment.yaml
│       ├── service.yaml
│       ├── ingress.yaml
│       ├── hpa.yaml
│       ├── pdb.yaml
│       ├── serviceaccount.yaml
│       ├── secret.yaml
│       └── configmap.yaml
├── orchestrator-deployment.yaml
├── agent-deployment.yaml
└── microvm-daemonset.yaml
```

## AI Gateway Helm Chart

Production-grade Helm chart for AI inference with GPU support.

### Quick Start

```bash
# Install with default values
helm install ai-gateway ./ai-gateway -n functionfly --create-namespace

# Install with production values
helm install ai-gateway ./ai-gateway -n functionfly \
  -f ai-gateway/values.production.yaml
```

## GPU Node Pools

### Google Kubernetes Engine (GKE)

Create a GPU node pool for AI workloads:

```bash
# Create cluster with GPU nodes
gcloud container clusters create functionfly-cluster \
  --region us-central1 \
  --machine-type n1-standard-4

# Add GPU node pool with NVIDIA T4
gcloud container node-pools create gpu-pool \
  --cluster functionfly-cluster \
  --region us-central1 \
  --machine-type n1-standard-4 \
  --accelerator type=nvidia-tesla-t4,count=1 \
  --disk-size 100 \
  --node-labels node-type=gpu \
  --num-nodes 2

# Install NVIDIA device plugin
kubectl apply -f https://raw.githubusercontent.com/NVIDIA/k8s-device-plugin/v1.12/nvidia-device-plugin.yml
```

### Amazon EKS

Create a GPU node group:

```bash
# Using eksctl
eksctl create nodegroup \
  --cluster functionfly-cluster \
  --region us-west-2 \
  --name gpu-nodes \
  --node-type p3.2xlarge \
  --nodes 2 \
  --nodes-min 1 \
  --nodes-max 5 \
  --node-ami-family Ubuntu2004 \
  --node-labels "node-type=gpu" \
  --asg-access \
  --external-dns-access \
  --full-ecr-access

# Or using AWS Console/CLI with Deep Learning AMI
# Select ami-0f1e0526c875db9c4 for Ubuntu 20.04 with NVIDIA drivers
```

### Azure AKS

Create a GPU agent pool:

```bash
# Create AKS cluster
az aks create \
  --resource-group functionfly-rg \
  --name functionfly-cluster \
  --location eastus \
  --vm-set-type VirtualMachineScaleSets \
  --load-balancer-sku standard

# Add GPU node pool
az aks nodepool add \
  --resource-group functionfly-rg \
  --cluster-name functionfly-cluster \
  --name gpu-pool \
  --vm-size Standard_NC6s_v3 \
  --node-count 2 \
  --node-labels node-type=gpu \
  --enable-cluster-autoscaler \
  --min-count 1 \
  --max-count 5

# Install NVIDIA device plugin
kubectl apply -f https://raw.githubusercontent.com/Azure/NVIDIA-device-plugin/master/nvidia-device-plugin.yml
```

## Common Operations

### Scale AI Gateway

```bash
# Manual scale
kubectl scale deployment ai-gateway --replicas=5 -n functionfly

# Or via Helm
helm upgrade ai-gateway ./ai-gateway --set replicaCount=5 -n functionfly
```

### Check GPU availability

```bash
# List nodes with GPU capacity
kubectl get nodes -l node-type=gpu -o jsonpath='{range .items[*]}{.metadata.name}{"\t"}{.status.allocatable.nvidia\.com/gpu}{"\n"}{end}'

# Verify GPU allocation
kubectl describe nodes -l node-type=gpu | grep -A 10 "Allocated resources"
```

### View logs

```bash
# AI Gateway logs
kubectl logs -l app=ai-gateway -n functionfly -f

# With previous container logs
kubectl logs -l app=ai-gateway -n functionfly --previous
```

## Resource Requirements

| Component | CPU (Request/Limit) | Memory (Request/Limit) | GPU |
|-----------|---------------------|------------------------|-----|
| AI Gateway | 500m / 4000m | 2Gi / 8Gi | 1 |
| Orchestrator API | 250m / 500m | 256Mi / 512Mi | 0 |
| Agent Service | 500m / 2000m | 512Mi / 2Gi | 0 |

## Security

- All containers run as non-root (UID 1000)
- Read-only root filesystem recommended in production
- Secrets managed via Kubernetes secrets or external secrets operators
- Network policies restrict traffic between namespaces

## Monitoring

Prometheus metrics are exposed at `/metrics` on port 8082 for AI Gateway:

- `ai_gateway_inference_requests_total`
- `ai_gateway_inference_duration_seconds`
- `ai_gateway_gpu_utilization`
- `ai_gateway_active_inferences`
- `ai_gateway_queue_depth`
- `ai_gateway_cost_per_tenant`

Grafana dashboards available in `deploy/monitoring/grafana/`.

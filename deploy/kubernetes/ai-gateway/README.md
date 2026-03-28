# AI Gateway Helm Chart

Production-grade Helm chart for deploying the FunctionFly AI Gateway on Kubernetes with GPU support.

## Features

- **GPU Support**: NVIDIA GPU resource scheduling with proper node selection
- **Horizontal Pod Autoscaling**: CPU/memory-based scaling with custom metrics support
- **Pod Disruption Budget**: Ensures high availability during updates
- **Multi-Region Support**: Regional services for geo-aware routing
- **Model Cache**: Persistent volume for model caching
- **Security**: Non-root containers, read-only root filesystem option
- **Prometheus Metrics**: Out-of-the-box metrics integration

## Prerequisites

- Kubernetes 1.19+
- Helm 3.2+
- NVIDIA GPU nodes (for GPU inference)
- `nvidia.com/gpu` resource available on nodes

## Installing the Chart

### Add the repository

```bash
helm repo add functionfly https://functionfly.github.io/charts
helm repo update
```

### Install with default values

```bash
helm install ai-gateway functionfly/ai-gateway -n functionfly --create-namespace
```

### Install with production values

```bash
helm install ai-gateway functionfly/ai-gateway \
  -n functionfly \
  --create-namespace \
  -f values.production.yaml
```

### Install with custom values

```bash
helm install ai-gateway functionfly/ai-gateway \
  -n functionfly \
  --create-namespace \
  --set replicaCount=3 \
  --set resources.limits.nvidia.com/gpu=1
```

## Configuration

| Parameter | Description | Default |
|-----------|-------------|---------|
| `replicaCount` | Number of replicas | `2` |
| `image.repository` | Image repository | `functionfly/ai-gateway` |
| `image.tag` | Image tag | `latest` |
| `image.pullPolicy` | Image pull policy | `IfNotPresent` |
| `service.type` | Service type | `ClusterIP` |
| `service.port` | Service port | `8082` |
| `resources.requests.cpu` | CPU request | `500m` |
| `resources.requests.memory` | Memory request | `2Gi` |
| `resources.limits.cpu` | CPU limit | `4000m` |
| `resources.limits.memory` | Memory limit | `8Gi` |
| `resources.requests.nvidia.com/gpu` | GPU request | `1` |
| `resources.limits.nvidia.com/gpu` | GPU limit | `1` |
| `autoscaling.enabled` | Enable HPA | `true` |
| `autoscaling.minReplicas` | Minimum replicas | `2` |
| `autoscaling.maxReplicas` | Maximum replicas | `10` |
| `autoscaling.targetCPUUtilizationPercentage` | Target CPU % | `70` |
| `nodeSelector` | Node selector for GPU nodes | `node-type: gpu` |

## GPU Node Pool Configuration

### Google Kubernetes Engine (GKE)

```bash
# Create a GPU node pool with NVIDIA T4
gcloud container node-pools create gpu-pool \
  --cluster=my-cluster \
  --region=us-central1 \
  --machine-type=n1-standard-4 \
  --accelerator=type=nvidia-tesla-t4,count=1 \
  --disk-size=100 \
  --node-labels=node-type=gpu \
  --num-nodes=2
```

### Amazon EKS

```bash
# Create a GPU node group using eksctl
eksctl create nodegroup \
  --cluster=my-cluster \
  --region=us-west-2 \
  --name=gpu-nodes \
  --node-type=p3.2xlarge \
  --nodes=2 \
  --nodes-min=1 \
  --nodes-max=5 \
  --node-ami-family=NVIDIA \
  --node-labels="node-type=gpu"
```

### Azure AKS

```bash
# Create a GPU agent pool
az aks nodepool add \
  --resource-group myResourceGroup \
  --cluster-name myCluster \
  --name gpu-pool \
  --vm-size Standard_NC6s_v3 \
  --node-count=2 \
  --node-labels node-type=gpu
```

## Upgrading

```bash
# Upgrade to a new version
helm upgrade ai-gateway functionfly/ai-gateway -n functionfly

# Upgrade with custom values
helm upgrade ai-gateway functionfly/ai-gateway \
  -n functionfly \
  -f values.production.yaml
```

## Uninstalling

```bash
helm uninstall ai-gateway -n functionfly
```

## Architecture

```
┌─────────────────────────────────────────────────────────────────┐
│                     Kubernetes Cluster                          │
│  ┌────────────────────────────────────────────────────────────┐  │
│  │              AI Gateway Service (ClusterIP)               │  │
│  │                        Port: 8082                          │  │
│  └────────────────────────────────────────────────────────────┘  │
│                              │                                   │
│         ┌────────────────────┼────────────────────┐             │
│         ▼                    ▼                    ▼             │
│  ┌─────────────┐       ┌─────────────┐       ┌─────────────┐     │
│  │  ai-gateway │       │  ai-gateway │       │  ai-gateway │     │
│  │  (replica) │       │  (replica)  │       │  (replica)  │     │
│  │   GPU: 1   │       │   GPU: 1    │       │   GPU: 1    │     │
│  └─────────────┘       └─────────────┘       └─────────────┘     │
│  ┌────────────────────────────────────────────────────────────┐  │
│  │              Model Cache PVC (50Gi SSD)                    │  │
│  └────────────────────────────────────────────────────────────┘  │
└─────────────────────────────────────────────────────────────────┘
                              │
                              ▼
                    ┌─────────────────┐
                    │   RunPod API    │
                    │   (Inference)   │
                    └─────────────────┘
```

## Monitoring

The AI Gateway exposes Prometheus metrics at `/metrics`:

- `ai_gateway_inference_requests_total` - Total inference requests
- `ai_gateway_inference_duration_seconds` - Inference latency histogram
- `ai_gateway_gpu_utilization` - GPU utilization gauge
- `ai_gateway_active_inferences` - Active inference count
- `ai_gateway_queue_depth` - Queue depth
- `ai_gateway_cost_per_tenant` - Cost per tenant

## Troubleshooting

### GPU not available

Ensure the NVIDIA device plugin is installed:

```bash
kubectl apply -f https://raw.githubusercontent.com/NVIDIA/k8s-device-plugin/v1.11/nvidia-device-plugin.yml
```

### Pods in Pending state

Check if there are enough GPU resources:

```bash
kubectl describe pod <pod-name>
```

### Model cache issues

Ensure the PVC is properly provisioned and accessible:

```bash
kubectl get pvc -n functionfly
kubectl describe pvc ai-gateway-model-cache -n functionfly
```

# FunctionFly Kubernetes Deployment

Deploy FunctionFly on Kubernetes using this guide.

> **Note**: This is a reference implementation. Adjust resources and configurations based on your workload requirements.

## Prerequisites

- Kubernetes cluster (v1.24+)
- kubectl configured
- Helm 3.x
- PostgreSQL (managed or self-hosted)
- Redis (managed or self-hosted)

## Quick Start

```bash
# Add FunctionFly Helm repository
helm repo add functionfly https://functionfly.github.io/charts
helm repo update

# Install FunctionFly
helm install functionfly functionfly/functionfly \
  --set global.postgresql.url="postgres://user:pass@postgres:5432/functionfly" \
  --set global.redis.url="redis://redis:6379"
```

## Manual Deployment

### Namespace

```yaml
# namespace.yaml
apiVersion: v1
kind: Namespace
metadata:
  name: functionfly
```

### ConfigMap

```yaml
# configmap.yaml
apiVersion: v1
kind: ConfigMap
metadata:
  name: functionfly-config
  namespace: functionfly
data:
  PORT: "8080"
  LOG_LEVEL: "info"
  DB_HOST: "postgres"
  DB_PORT: "5432"
  DB_NAME: "functionfly"
  REDIS_ADDR: "redis:6379"
  BASE_URL: "https://api.functionfly.com"
  FRONTEND_URL: "https://app.functionfly.com"
```

### Secrets

```yaml
# secrets.yaml
apiVersion: v1
kind: Secret
metadata:
  name: functionfly-secrets
  namespace: functionfly
type: Opaque
stringData:
  DB_USER: "functionfly"
  DB_PASSWORD: "your-db-password"
  REDIS_PASSWORD: "your-redis-password"
  JWT_SECRET: "your-jwt-secret"
  API_SHARED_SECRET: "your-api-secret"
```

### Deployment

```yaml
# deployment.yaml
apiVersion: apps/v1
kind: Deployment
metadata:
  name: orchestrator-api
  namespace: functionfly
spec:
  replicas: 3
  selector:
    matchLabels:
      app: orchestrator-api
  template:
    metadata:
      labels:
        app: orchestrator-api
    spec:
      containers:
      - name: orchestrator-api
        image: functionfly/orchestrator:latest
        ports:
        - containerPort: 8080
        envFrom:
        - configMapRef:
            name: functionfly-config
        - secretRef:
            name: functionfly-secrets
        resources:
          requests:
            memory: "256Mi"
            cpu: "250m"
          limits:
            memory: "512Mi"
            cpu: "500m"
        livenessProbe:
          httpGet:
            path: /healthz
            port: 8080
          initialDelaySeconds: 30
          periodSeconds: 10
        readinessProbe:
          httpGet:
            path: /readyz
            port: 8080
          initialDelaySeconds: 5
          periodSeconds: 5
```

### Service

```yaml
# service.yaml
apiVersion: v1
kind: Service
metadata:
  name: orchestrator-api
  namespace: functionfly
spec:
  selector:
    app: orchestrator-api
  ports:
  - port: 80
    targetPort: 8080
  type: ClusterIP
```

### Horizontal Pod Autoscaler

```yaml
# hpa.yaml
apiVersion: autoscaling/v2
kind: HorizontalPodAutoscaler
metadata:
  name: orchestrator-api
  namespace: functionfly
spec:
  scaleTargetRef:
    apiVersion: apps/v1
    kind: Deployment
    name: orchestrator-api
  minReplicas: 2
  maxReplicas: 10
  metrics:
  - type: Resource
    resource:
      name: cpu
      target:
        type: Utilization
        averageUtilization: 70
```

## Runtime Deployments

### Node.js Runtime

```yaml
# runtime-nodejs.yaml
apiVersion: apps/v1
kind: Deployment
metadata:
  name: runtime-nodejs
  namespace: functionfly
spec:
  replicas: 2
  selector:
    matchLabels:
      app: runtime-nodejs
  template:
    metadata:
      labels:
        app: runtime-nodejs
    spec:
      containers:
      - name: runtime-nodejs
        image: functionfly/runtime-nodejs:latest
        ports:
        - containerPort: 8082
        env:
        - name: RUNTIME_TYPE
          value: "nodejs"
        - name: ORCHESTRATOR_URL
          value: "http://orchestrator-api:8080"
```

### Python Runtime

```yaml
# runtime-python.yaml
apiVersion: apps/v1
kind: Deployment
metadata:
  name: runtime-python
  namespace: functionfly
spec:
  replicas: 2
  selector:
    matchLabels:
      app: runtime-python
  template:
    metadata:
      labels:
        app: runtime-python
    spec:
      containers:
      - name: runtime-python
        image: functionfly/runtime-python:latest
        ports:
        - containerPort: 8083
        env:
        - name: RUNTIME_TYPE
          value: "python"
        - name: ORCHESTRATOR_URL
          value: "http://orchestrator-api:8080"
```

### Local/WASM Runtime

```yaml
# runtime-local.yaml
apiVersion: apps/v1
kind: Deployment
metadata:
  name: runtime-local
  namespace: functionfly
spec:
  replicas: 2
  selector:
    matchLabels:
      app: runtime-local
  template:
    metadata:
      labels:
        app: runtime-local
    spec:
      containers:
      - name: runtime-local
        image: functionfly/runtime-local:latest
        ports:
        - containerPort: 8081
        env:
        - name: RUNTIME_TYPE
          value: "local"
        - name: ORCHESTRATOR_URL
          value: "http://orchestrator-api:8080"
```

## Ingress

```yaml
# ingress.yaml
apiVersion: networking.k8s.io/v1
kind: Ingress
metadata:
  name: functionfly-ingress
  namespace: functionfly
  annotations:
    kubernetes.io/ingress.class: nginx
    cert-manager.io/cluster-issuer: letsencrypt-prod
spec:
  tls:
  - hosts:
    - api.functionfly.com
    secretName: functionfly-tls
  rules:
  - host: api.functionfly.com
    http:
      paths:
      - path: /
        pathType: Prefix
        backend:
          service:
            name: orchestrator-api
            port:
              number: 80
```

## Monitoring

### ServiceMonitor for Prometheus

```yaml
# servicemonitor.yaml
apiVersion: monitoring.coreos.com/v1
kind: ServiceMonitor
metadata:
  name: functionfly
  namespace: functionfly
spec:
  selector:
    matchLabels:
      app: orchestrator-api
  endpoints:
  - port: metrics
    path: /metrics
```

## Apply Resources

```bash
# Create namespace
kubectl apply -f namespace.yaml

# Apply configs and secrets
kubectl apply -f configmap.yaml
kubectl apply -f secrets.yaml

# Apply deployments
kubectl apply -f deployment.yaml
kubectl apply -f runtime-nodejs.yaml
kubectl apply -f runtime-python.yaml
kubectl apply -f runtime-local.yaml

# Apply services
kubectl apply -f service.yaml

# Apply autoscaling
kubectl apply -f hpa.yaml

# Apply ingress
kubectl apply -f ingress.yaml
```

## Verify Deployment

```bash
# Check pods
kubectl get pods -n functionfly

# Check services
kubectl get svc -n functionfly

# Check ingress
kubectl get ingress -n functionfly

# View logs
kubectl logs -n functionfly -l app=orchestrator-api

# Check health
kubectl exec -it <pod> -n functionfly -- curl localhost:8080/healthz
```

## Helm Chart (Alternative)

For production, use the Helm chart:

```bash
# Create values file
cat > values.yaml << EOF
replicaCount: 3

image:
  repository: functionfly/orchestrator
  tag: latest
  pullPolicy: IfNotPresent

service:
  type: ClusterIP
  port: 80

resources:
  limits:
    cpu: 1000m
    memory: 1Gi
  requests:
    cpu: 250m
    memory: 256Mi

autoscaling:
  enabled: true
  minReplicas: 2
  maxReplicas: 10
  targetCPUUtilizationPercentage: 70

postgresql:
  enabled: false
  host: postgres.example.com
  port: 5432
  database: functionfly

redis:
  enabled: false
  host: redis.example.com
  port: 6379
EOF

# Install with Helm
helm install functionfly functionfly/functionfly -f values.yaml
```

## Troubleshooting

### Pod not starting

```bash
# Check events
kubectl describe pod <pod> -n functionfly

# Check logs
kubectl logs <pod> -n functionfly
```

### Database connection issues

```bash
# Verify secrets
kubectl get secret functionfly-secrets -n functionfly -o yaml

# Test connection
kubectl exec -it <pod> -n functionfly -- nc -zv $DB_HOST $DB_PORT
```

### Memory issues

```bash
# Check resource usage
kubectl top pods -n functionfly

# Check OOM kills
kubectl get events -n functionfly | grep OOM
```

## Support

- GitHub Issues: https://github.com/functionfly/functionfly/issues
- Discord: https://discord.gg/functionfly

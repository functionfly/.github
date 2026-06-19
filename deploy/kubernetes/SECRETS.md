# FunctionFly Kubernetes Secrets Management

This guide covers secrets management for FunctionFly deployments on Kubernetes.

## Overview

FunctionFly uses Kubernetes Secrets for sensitive configuration:
- Database credentials
- Redis password
- JWT secret
- API shared secrets
- External service credentials (Stripe, Resend, etc.)

## Creating Secrets

### Option 1: kubectl create secret (Manual)

```bash
# Create a generic secret from literal values
kubectl create secret generic functionfly-secrets \
  --from-literal=db-host="postgres" \
  --from-literal=db-port="5432" \
  --from-literal=db-user="functionfly" \
  --from-literal=db-password="your-secure-password" \
  --from-literal=db-name="functionfly" \
  --from-literal=redis-addr="redis:6379" \
  --from-literal=jwt-secret="your-jwt-secret-min-32-chars" \
  -n functionfly
```

### Option 2: kubectl create secret from file

```bash
# Create a secret from a .env file
kubectl create secret generic functionfly-secrets \
  --from-env-file=.env.production \
  -n functionfly
```

### Option 3: sealed-secrets (GitOps Compatible)

For GitOps workflows where secrets shouldn't be in Git:

```bash
# Install sealed-secrets controller
helm install sealed-secrets sealed-secrets \
  -n kube-system \
  --set-string namespace=functionfly

# Encrypt your secrets
kubeseal --cert=cert.pem --secret-file=secrets.yaml --密封秘密.yaml -n functionfly

# The sealed secret can be committed to Git
```

### Option 4: External Secrets Operator

For integration with external secret managers (AWS Secrets Manager, HashiCorp Vault, etc.):

```yaml
# Install External Secrets Operator
apiVersion: external-secrets.io/v1beta1
kind: ClusterSecretStore
metadata:
  name: aws-secrets-manager
spec:
  provider:
    aws:
      service: SecretsManager
      region: us-east-1
---
apiVersion: external-secrets.io/v1beta1
kind: ExternalSecret
metadata:
  name: functionfly-secrets
  namespace: functionfly
spec:
  refreshInterval: 1h
  secretStoreRef:
    name: aws-secrets-manager
    kind: ClusterSecretStore
  target:
    name: functionfly-secrets
    creationPolicy: Owner
  data:
    - secretKey: db-password
      remoteRef:
        key: production/functionfly
        property: db_password
```

## Secret Reference in Deployments

FunctionFly Kubernetes deployments reference secrets via `secretKeyRef`:

```yaml
env:
  - name: DB_PASSWORD
    valueFrom:
      secretKeyRef:
        name: functionfly-secrets
        key: db-password
```

## Rotating Secrets

### Database Password Rotation

1. Update the secret:
```bash
kubectl patch secret functionfly-secrets \
  --type=strategic \
  --patch='{"stringData":{"db-password":"new-password"}}' \
  -n functionfly
```

2. Restart pods to pick up new credentials:
```bash
kubectl rollout restart deployment/orchestrator-api -n functionfly
kubectl rollout restart deployment/health-monitor -n functionfly
```

3. Verify connections:
```bash
kubectl exec -it deploy/orchestrator-api -n functionfly -- curl localhost:8080/health/ready
```

### JWT Secret Rotation

⚠️ **Warning**: JWT secret rotation will invalidate all existing sessions.

```bash
# Generate new JWT secret
NEW_SECRET=$(openssl rand -hex 32)

# Update secret
kubectl patch secret functionfly-secrets \
  --type=strategic \
  --patch="{\"stringData\":{\"jwt-secret\":\"$NEW_SECRET\"}}" \
  -n functionfly

# Restart pods
kubectl rollout restart deployment/orchestrator-api -n functionfly
```

## Secret Backup

Always backup secrets securely:

```bash
# Export all secrets (encrypt this backup!)
kubectl get secrets -n functionfly -o yaml > secrets-backup.yaml

# Or use a tool like Vault for enterprise secret management
```

## External Secret Managers

### AWS Secrets Manager

```yaml
apiVersion: external-secrets.io/v1beta1
kind: ClusterSecretStore
metadata:
  name: aws-secrets-manager
spec:
  provider:
    aws:
      service: SecretsManager
      region: us-east-1
      auth:
        secretRef:
          accessKeyIDSecretRef:
            name: aws-creds
            key: access-key
            namespace: functionfly
---
apiVersion: external-secrets.io/v1beta1
kind: ExternalSecret
metadata:
  name: functionfly-secrets
  namespace: functionfly
spec:
  refreshInterval: 1h
  secretStoreRef:
    name: aws-secrets-manager
    kind: ClusterSecretStore
  target:
    name: functionfly-secrets
  data:
    - secretKey: db-password
      remoteRef:
        key: production/functionfly/database
        property: password
    - secretKey: jwt-secret
      remoteRef:
        key: production/functionfly/auth
        property: jwt_secret
```

### HashiCorp Vault

```yaml
apiVersion: external-secrets.io/v1beta1
kind: ClusterSecretStore
metadata:
  name: vault-backend
spec:
  provider:
    vault:
      server: "https://vault.example.com:8200"
      path: "secret"
      version: "v2"
      auth:
        kubernetes:
          mountPath: "kubernetes"
          role: "functionfly"
---
apiVersion: external-secrets.io/v1beta1
kind: ExternalSecret
metadata:
  name: functionfly-secrets
  namespace: functionfly
spec:
  refreshInterval: 1h
  secretStoreRef:
    name: vault-backend
    kind: ClusterSecretStore
  target:
    name: functionfly-secrets
  data:
    - secretKey: db-password
      secretPath: "secret/data/production/functionfly"
      property: db_password
```

## Security Best Practices

1. **Never commit secrets to Git** - Use sealed-secrets or External Secrets Operator
2. **Enable RBAC** - Restrict who can read/write secrets
3. **Use encryption at rest** - Enable etcd encryption
4. **Rotate regularly** - Implement automated secret rotation
5. **Use TLS** - Always use TLS for secret transmission
6. **Audit access** - Enable Kubernetes audit logging for secret access

## Troubleshooting

### Secret not found

```bash
# Check if secret exists
kubectl get secret functionfly-secrets -n functionfly

# Check secret contents
kubectl get secret functionfly-secrets -n functionfly -o yaml

# Decode a specific key
kubectl get secret functionfly-secrets -n functionfly -o jsonpath='{.data.db-password}' | base64 -d
```

### Pod can't access secret

```bash
# Check pod status
kubectl describe pod <pod-name> -n functionfly

# Check if secret is mounted correctly
kubectl exec -it <pod-name> -n functionfly -- ls -la /etc/secrets/

# Verify secret exists in same namespace
kubectl get secrets -n functionfly
```

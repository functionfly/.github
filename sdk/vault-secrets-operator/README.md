# functionfly-vault-secrets-operator

A minimal Kubernetes operator that syncs [FunctionFly vault][vault]
secrets into Kubernetes, and injects short-lived dynamic DB
credentials into pods.

[vault]: https://github.com/functionfly/functionfly

## CRDs

### `VaultSecret`

Sync a FunctionFly vault secret into a K8s `Secret`.

```yaml
apiVersion: functionfly.io/v1alpha1
kind: VaultSecret
metadata:
  name: stripe-api
  namespace: payments
spec:
  tenantId: "00000000-0000-0000-0000-000000000000"
  secretId: "11111111-1111-1111-1111-111111111111"
  apiServer: https://api.functionfly.com  # optional
  refreshInterval: 30s                     # default 30s
  secretType: Opaque                       # or kubernetes.io/tls
  apiTokenRef:
    name: ff-credentials
    namespace: payments
    key: token
```

The operator polls the FunctionFly API and creates a Kubernetes
`Secret` named `stripe-api-vault` containing the encrypted payload
(`ciphertext`, `iv`, `salt`, `tag`, `key_version`).

Pods consume it as usual:

```yaml
envFrom:
  - secretRef:
      name: stripe-api-vault
```

### `VaultDynamicCredential`

Generate short-lived dynamic DB credentials and expose them to a
pod as environment variables.

```yaml
apiVersion: functionfly.io/v1alpha1
kind: VaultDynamicCredential
metadata:
  name: orders-db
  namespace: orders
spec:
  tenantId: "00000000-0000-0000-0000-000000000000"
  credentialId: "22222222-2222-2222-2222-222222222222"
  apiTokenRef:
    name: ff-credentials
    namespace: orders
  ttlSeconds: 3600
  renewBeforeSeconds: 300
  env:
    username: DB_USER
    password: DB_PASSWORD
    host: DB_HOST
    port: DB_PORT
    database: DB_NAME
    expiresAt: DB_EXPIRES_AT
```

The operator mints a credential, writes it to a tmpfs volume, and
the pod's projected env vars (or a sidecar) read it. Renewals are
handled automatically.

## Install

```sh
# From the repo root
kubectl apply -f sdk/vault-secrets-operator/manifests/crd.yaml
kubectl apply -f sdk/vault-secrets-operator/manifests/operator.yaml
```

A `make build` target compiles the binary; a `make container` target
produces the image.

## v2 roadmap

- CSI driver (replaces tmpfs volume)
- Webhook for defaulting + validation
- HashiCorp-Vault-style audit log shipping

## License

Apache-2.0

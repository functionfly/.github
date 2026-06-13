# functionfly-vault-sdk (Go)

The official Go SDK for the FunctionFly zero-knowledge secrets vault.

```go
import (
    "context"

    "github.com/functionfly/go-vault-sdk/vault"
)

client, _ := vault.NewClient(
    "https://api.functionfly.com",
    vault.WithToken("fnly_xxx"),
)

// Pre-encrypt the value client-side (out of scope of this SDK;
// use the dashboard's crypto helpers, your own Argon2id + AES-256-GCM
// pipeline, or the helper in this package's `kdf.go`).
salt, _ := vault.NewSalt(16)
ct, iv, tag := encryptMyValue("super-secret", salt)

secret, err := client.Secrets.Create(ctx, vault.SecretCreate{
    Name:       "STRIPE_API_KEY",
    SecretType: vault.SecretTypeAPIKey,
    EncryptedData: vault.EncryptedData{
        Ciphertext: ct,
        IV:         iv,
        Salt:       salt,
        Tag:        tag,
        KeyVersion: 2, // 1=PBKDF2 (legacy), 2=Argon2id
    },
})
```

## Dynamic credentials

```go
// Register a target (one-time).
target, _ := client.DynamicTargets.Create(ctx, vault.DynamicTargetCreate{
    Name:          "prod-postgres",
    DBType:        vault.DynamicDBPostgres,
    Host:          "db.internal",
    Port:          5432,
    DatabaseName:  "app",
    AdminUsername: "vault_admin",
    AdminPassword: os.Getenv("PG_ADMIN_PASSWORD"),
    AllowedRoles:  []string{"app_readonly"},
    DefaultTTLSeconds: 3600,
    MaxTTLSeconds:    86400,
})

// Register a credential template.
cred, _ := client.DynamicCredentials.Create(ctx, vault.DynamicCredentialCreate{
    TargetID:      target.ID,
    Name:          "app-readonly",
    RoleTemplate:  "app_readonly",
    TTLSeconds:    3600,
    MaxTTLSeconds: 86400,
})

// Mint a fresh credential (e.g. at app startup, or on demand).
gen, _ := client.DynamicCredentials.Generate(ctx, cred.ID, vault.GenerateOptions{})
// gen.Username, gen.Password, gen.ExpiresAt are valid until gen.ExpiresAt.

// Renew when TTL is about to expire.
newExpiry, _ := client.Leases.Renew(ctx, cred.ID, gen.LeaseID, vault.RenewOptions{TTLSeconds: 3600})

// Revoke explicitly.
client.Leases.Revoke(ctx, cred.ID, gen.LeaseID)
```

## Tokens (runtime access)

```go
tok, _ := client.Tokens.Create(ctx, vault.TokenCreate{
    SecretID:       secret.ID,
    ExpiresInHours: 24,
    Name:           "k8s-pod-1",
})
// tok.Token is plaintext — store immediately.
```

## Audit

```go
entries, _ := client.Audit.List(ctx, vault.AuditListOptions{
    Action: "create",
    Limit:  50,
})
```

## Status

- ✅ Secret CRUD
- ✅ Token CRUD
- ✅ Dynamic targets (PostgreSQL, MySQL)
- ✅ Dynamic credential generation
- ✅ Leases — renew, revoke
- ✅ Audit log
- ✅ Server-side admin-credential encryption (envelope)
- ⏳ Client-side Argon2id + AES-256-GCM helper (use your own; helpers
       for the *parameters* live in `kdf.go`)

## License

Apache-2.0

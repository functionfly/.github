# @functionfly/vault

The official Node.js / TypeScript SDK for the [FunctionFly vault][vault] —
a zero-knowledge secrets manager.

[vault]: https://github.com/functionfly/functionfly

```ts
import { VaultClient, SecretType } from "@functionfly/vault";

const client = new VaultClient({
  baseUrl: "https://api.functionfly.com",
  token: process.env.FFLY_TOKEN!,
});

// 1. Encrypt locally (zero-knowledge) — caller is responsible for
//    Argon2id + AES-256-GCM. The web crypto helpers in your app or
//    library can do this.
const { ciphertext, iv, salt, tag } = await encryptMyValue("super-secret");

// 2. Store
const secret = await client.secrets.create({
  name: "STRIPE_API_KEY",
  secretType: SecretType.APIKey,
  encryptedData: { ciphertext, iv, salt, tag, keyVersion: 2 },
});

// 3. Fetch later
const fetched = await client.secrets.get(secret.id);
const plaintext = await decryptMyValue(fetched.encryptedData);
```

## Dynamic credentials

```ts
// One-time: register a target + credential template.
const target = await client.dynamicTargets.create({
  name: "prod-postgres",
  dbType: DynamicDBType.Postgres,
  host: "db.internal",
  port: 5432,
  databaseName: "app",
  adminUsername: "vault_admin",
  adminPassword: process.env.PG_ADMIN!,
  allowedRoles: ["app_readonly"],
  defaultTtlSeconds: 3600,
  maxTtlSeconds: 86400,
});

const cred = await client.dynamicCredentials.create({
  targetId: target.id,
  name: "app-readonly",
  roleTemplate: "app_readonly",
  ttlSeconds: 3600,
  maxTtlSeconds: 86400,
});

// At app startup: mint a fresh credential.
const gen = await client.dynamicCredentials.generate(cred.id);
await connectToDb({
  host: gen.host,
  port: gen.port,
  username: gen.username,
  password: gen.password,
  database: gen.database,
});

// Renew before expiry (or revoke early).
await client.leases.renew(cred.id, gen.leaseId, { ttlSeconds: 3600 });
await client.leases.revoke(cred.id, gen.leaseId);
```

## API surface

| Service | Methods |
|---|---|
| `client.secrets` | `create`, `get`, `update`, `rotate`, `delete`, `list` |
| `client.tokens` | `create`, `list`, `revoke` |
| `client.dynamicTargets` | `create`, `list`, `delete`, `test` |
| `client.dynamicCredentials` | `create`, `generate`, `revokeAll` |
| `client.leases` | `renew`, `revoke` |
| `client.audit` | `list` |

## Error handling

```ts
import { VaultAPIError } from "@functionfly/vault";

try {
  await client.secrets.get("nope");
} catch (err) {
  if (err instanceof VaultAPIError) {
    console.error(err.status, err.code, err.message);
  }
}
```

## License

Apache-2.0

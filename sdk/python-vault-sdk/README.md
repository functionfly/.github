# functionfly-vault

The official Python SDK for the [FunctionFly vault][vault] — a
zero-knowledge secrets manager.

[vault]: https://github.com/functionfly/functionfly

```python
from functionfly_vault import VaultClient, SecretType

client = VaultClient(token="fnly_xxx", base_url="https://api.functionfly.com")

# Caller encrypts locally (zero-knowledge) — the SDK does not see plaintext.
ct, iv, salt, tag = encrypt_my_value("super-secret")

secret = client.secrets.create(
    name="STRIPE_API_KEY",
    secret_type=SecretType.API_KEY,
    encrypted_data={
        "ciphertext": ct,
        "iv": iv,
        "salt": salt,
        "tag": tag,
        "key_version": 2,  # 1=PBKDF2, 2=Argon2id
    },
)
```

## Dynamic credentials

```python
target = client.dynamic_targets.create(
    name="prod-postgres",
    db_type="postgres",
    host="db.internal",
    port=5432,
    database_name="app",
    admin_username="vault_admin",
    admin_password=os.environ["PG_ADMIN"],
    allowed_roles=["app_readonly"],
    default_ttl_seconds=3600,
    max_ttl_seconds=86400,
)

cred = client.dynamic_credentials.create(
    target_id=target["id"],
    name="app-readonly",
    role_template="app_readonly",
    ttl_seconds=3600,
    max_ttl_seconds=86400,
)

# Mint a fresh credential.
gen = client.dynamic_credentials.generate(cred["id"])
connect_to_db(
    host=gen["host"], port=gen["port"],
    username=gen["username"], password=gen["password"],
    database=gen["database"],
)

# Renew before expiry.
client.leases.renew(cred["id"], gen["lease_id"], ttl_seconds=3600)

# Revoke early.
client.leases.revoke(cred["id"], gen["lease_id"])
```

## API surface

| Service | Methods |
|---|---|
| `client.secrets` | `create`, `get`, `update`, `rotate`, `delete`, `list` |
| `client.tokens` | `create`, `list`, `revoke` |
| `client.dynamic_targets` | `create`, `list`, `delete`, `test` |
| `client.dynamic_credentials` | `create`, `generate`, `revoke_all` |
| `client.leases` | `renew`, `revoke` |
| `client.audit` | `list` |

## Error handling

```python
from functionfly_vault import VaultAPIError

try:
    client.secrets.get("nope")
except VaultAPIError as exc:
    print(exc.status, exc.code, exc.message)
```

## Requirements

- Python 3.9+
- No third-party runtime dependencies (uses the standard library's
  `urllib`).

## License

Apache-2.0

# functionfly/vault-action

A GitHub Actions composite action that fetches secrets from the
[FunctionFly vault][vault] and exports them as environment variables
in your workflow.

[vault]: https://github.com/functionfly/functionfly

## Quick start

```yaml
- uses: functionfly/vault-action@v1
  with:
    api-key: ${{ secrets.FUNCTIONFLY_API_KEY }}
    tenant-id: ${{ secrets.FUNCTIONFLY_TENANT_ID }}
    secrets: |
      STRIPE_API_KEY: 11111111-1111-1111-1111-111111111111
      DB_PASSWORD:    22222222-2222-2222-2222-222222222222
    export-vars: |
      STRIPE_API_KEY
      DB_PASSWORD
- name: Use them
  run: |
    echo "Stripe key prefix: ${STRIPE_API_KEY:0:8}"
```

## What it does

1. Fetches each `secret-uuid` from the FunctionFly vault using your
   API token.
2. Exports the *encrypted* payload to the workflow's environment,
   keyed by the `ENV_VAR` you specified.
3. (Optional) Re-fetches every `refresh-interval` to keep long
   workflows' secrets fresh.

Because the vault is zero-knowledge, the values written to
`$GITHUB_ENV` are the AES-256-GCM ciphertext + IV + salt + auth tag.
To **use** them, your job needs the passphrase used during
encryption — typically stored as a separate GitHub Actions secret
(`FUNCTIONFLY_VAULT_PASSPHRASE`). See the example decryption helper
in [`scripts/decrypt.ts`](scripts/decrypt.ts) (or write your own).

## Inputs

| Input | Required | Default | Description |
|---|---|---|---|
| `api-key` | yes | — | FunctionFly API token |
| `api-url` | no | `https://api.functionfly.com` | API base URL |
| `tenant-id` | yes | — | Tenant UUID owning the secrets |
| `secrets` | yes | — | Newline-separated `ENV_VAR: secret-uuid` |
| `refresh-interval` | no | `0` | Refresh period; `0` = no refresh |
| `export-vars` | yes | — | Newline-separated ENV_VAR names |
| `fail-on-missing` | no | `true` | Fail the workflow on any fetch error |

## Outputs

| Output | Description |
|---|---|
| `fetched-count` | Number of secrets successfully fetched |
| `refreshed-count` | Number refreshed during the run (≥0) |

## Security notes

- Pass `--fail-on-missing=false` if you want to keep going on
  transient API errors; otherwise the action errors out fast.
- We never log the ciphertexts to keep the action's output
  scannable. If you need to debug, add `set -x` in your workflow.
- For high-security environments, prefer
  `permissions: contents: read` + a job-scoped token rather than a
  long-lived token in repo secrets.

## License

Apache-2.0

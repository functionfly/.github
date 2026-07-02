---
title: Audit Certificates
description: Compliance-grade, tamper-proof proof of what was replayed and reconciled
sidebar:
  order: 4
---


An **audit certificate** is a tamper-proof, digitally signed document that
proves exactly what was replayed, what changed, and what was reconciled.
It is designed for compliance frameworks including SOC 2, HIPAA, and ISO 27001.

:::note[Enterprise only]
Audit certificates are available on the Enterprise plan and above.
:::

## What's in a Certificate

Each certificate contains:

- **Replay metadata** — Function ID, time window, target version, timestamps
- **Execution manifest** — Every execution ID that was replayed
- **Diff summary** — Counts by classification (identical, minor, major, breaking, error)
- **Reconciliation log** — Every action taken, with old/new values
- **Merkle tree** — Hash tree over all execution items for tamper detection
- **Ed25519 signature** — Cryptographic signature over the Merkle root
- **Certificate chain** — Links to the previous certificate for chain-of-custody
- **Compliance frameworks** — Tagged with applicable frameworks (SOC 2, HIPAA, ISO 27001)
- **Retention policy** — How long the certificate is retained

## Generating a Certificate

After a replay (and optional reconciliation) completes:

### Dashboard

1. Open the completed replay
2. Click **Generate Audit Certificate**
3. Select compliance frameworks
4. Download the certificate as JSON

### API

```bash
curl https://api.functionfly.com/v1/time-machine/replays/{id}/audit-certificate \
  -H "Authorization: Bearer $FUNCTIONFLY_API_KEY"
```

### Response

```json
{
  "certificate_id": "cert_7f3a2b",
  "replay_id": "replay_abc123",
  "generated_at": "2026-06-23T14:30:00Z",
  "function_id": "fx_abc123",
  "window_start": "2026-06-20T00:00:00Z",
  "window_end": "2026-06-23T00:00:00Z",
  "target_version_id": "v2.0.0",
  "total_executions": 2847,
  "diff_summary": {
    "identical": 2100,
    "minor": 412,
    "major": 287,
    "breaking": 48,
    "error": 0
  },
  "reconciliation": {
    "mode": "live",
    "actions_taken": 335,
    "actions_reversible": 335
  },
  "merkle_root": "a1b2c3d4e5f6...",
  "signature": "0xABCD1234...",
  "previous_cert_hash": "prev_cert_hash...",
  "compliance_frameworks": ["SOC2", "HIPAA", "ISO27001"],
  "retention_policy": "7_years",
  "cert_hash": "sha256:..."
}
```

## Verification

### Merkle Tree

The Merkle tree allows independent verification that no execution item was
tampered with after the certificate was generated. Each leaf node is the
SHA-256 hash of an execution item (input + output + metadata).

```
         ┌─────────────┐
         │  Merkle Root │
         └──────┬───────┘
          ┌─────┴─────┐
     ┌────┴────┐ ┌────┴────┐
     │ Hash AB │ │ Hash CD │
     └────┬────┘ └────┬────┘
     ┌────┴────┐ ┌────┴────┐
     │ Hash A  │ │ Hash C  │
     │ Hash B  │ │ Hash D  │
     └─────────┘ └─────────┘
```

To verify a single execution item:

1. Recompute its leaf hash from the stored input/output
2. Walk the Merkle proof (sibling hashes) up to the root
3. Compare against the signed Merkle root in the certificate

### Ed25519 Signature

The Merkle root is signed with FunctionFly's Ed25519 private key. The public
key is published at:

```
GET https://api.functionfly.com/.well-known/functionfly-signing-key.pem
```

Verify the signature using standard Ed25519 verification against the Merkle root.

### Certificate Chain

Each certificate links to its predecessor via `previous_cert_hash`. This
creates an append-only chain — if any certificate in the chain is altered,
the hashes break.

## Compliance Frameworks

Certificates can be tagged with one or more frameworks:

| Framework | What It Covers |
|---|---|
| **SOC 2** | Audit trail for data processing changes (CC7.2, CC7.3) |
| **HIPAA** | Protected health information access and modification log |
| **ISO 27001** | Information security incident management (A.16.1) |

The certificate includes the framework tags, retention policy, and all
evidence an auditor needs to verify the correction was performed correctly.

## Anchoring (Future)

Enterprise certificates support optional blockchain anchoring:

- `anchored` — Whether the Merkle root has been anchored
- `anchor_chain` — The blockchain network (e.g., `ethereum_mainnet`)
- `anchor_tx_hash` — The transaction hash of the anchor

This provides an additional layer of tamper evidence independent of
FunctionFly's infrastructure.

## Retention

Certificates are retained according to the policy set at generation time:

| Policy | Duration |
|---|---|
| `standard` | 1 year |
| `compliance` | 7 years (SOX) |
| `permanent` | Indefinite |

Retention policies are enforced by the data retention system. Certificates
under legal hold are never deleted regardless of retention policy.

## Next Steps

- [API Reference](/time-machine/api/) — Full endpoint docs
- [Security](/security/) — Platform security overview

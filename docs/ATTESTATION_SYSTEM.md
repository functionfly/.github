# Attestation System

**Version**: 1.0
**Status**: Production
**Base URL**: `https://api.functionfly.com/v1/trust`

FunctionFly's attestation system provides cryptographic proof of function properties — verification status, security scans, code reviews, execution results, compliance checks, and digital signatures. Attestations are immutable, hash-chained, Merkle-audited, and optionally zero-knowledge verifiable.

---

## Table of Contents

- [Concepts](#concepts)
- [Attestation Types](#attestation-types)
- [Cryptographic Signing](#cryptographic-signing)
- [Hash Chain](#hash-chain)
- [Merkle Audit Trail](#merkle-audit-trail)
- [Zero-Knowledge Proofs](#zero-knowledge-proofs)
- [API Reference](#api-reference)
- [SDK Packages](#sdk-packages)
- [Plan Tiers](#plan-tiers)
- [Architecture](#architecture)

---

## Concepts

An **attestation** is an immutable record that asserts a property about a function at a specific point in time. Each attestation:

- Has a unique public ID (`att_` prefix)
- Contains a **proof hash** (SHA-256 of all attestation data)
- Is **signed** using Ed25519 (or ECDSA-P256/RSA-PSS via HSM)
- Is linked to the previous attestation via a **hash chain**
- Is included in a **Merkle audit tree** for tamper-evident log verification
- Can optionally be verified using **zero-knowledge proofs**

```
┌─────────────────────────────────────────────────────────┐
│                    Attestation Record                    │
├─────────────────────────────────────────────────────────┤
│  att_id        att_a1b2c3d4e5f6...                      │
│  function_id   550e8400-e29b-41d4-a716-446655440000     │
│  type          verification                              │
│  status        valid                                     │
│  title         "Function verified (standard)"            │
│  proof_hash    sha256(function + version + type + ...)   │
│  signature     ed25519_sign(proof_hash, private_key)     │
│  previous_hash sha256(previous_attestation)              │
│  code_hash     sha256(function_source_code)              │
│  input_hash    sha256(execution_input)                   │
│  output_hash   sha256(execution_output)                  │
│  attested_at   2026-06-29T15:00:00Z                     │
└─────────────────────────────────────────────────────────┘
```

---

## Attestation Types

| Type | Description | Created By |
|------|-------------|------------|
| `verification` | Function passed identity/quality verification | Platform on verification completion |
| `security_scan` | Security scan results (SAST, dependency audit) | Automated scanner or partner |
| `code_review` | Code review passed by qualified reviewer | Partner with `attestation:create` scope |
| `execution` | Specific execution was verified with code/input/output hashes | Runtime or monitoring system |
| `compliance` | Function meets compliance requirements (SOC2, HIPAA, etc.) | Compliance auditor |
| `signature` | Digital signature attesting function authenticity | Function author or publisher |

### Status Lifecycle

```
valid ──────► revoked    (manual revocation by admin)
valid ──────► expired    (past valid_until date, auto-expired)
revoked ────► (terminal)
expired ────► (terminal)
```

---

## Cryptographic Signing

Attestations are signed using the platform's attestation signing key. The system supports multiple signing backends:

### Supported Algorithms

| Algorithm | Backend | Use Case |
|-----------|---------|----------|
| Ed25519 | Software (default) | Development and standard production |
| ECDSA-P256 | Software or PKCS#11 HSM | FIPS 140-2 compliance |
| RSA-PSS-SHA256 | AWS KMS or PKCS#11 HSM | Enterprise HSM requirements |

### Backend Configuration

Set via environment variables:

| Variable | Default | Description |
|----------|---------|-------------|
| `ATTESTATION_SIGNER_BACKEND` | `software` | `software`, `pkcs11`, or `awskms` |
| `ATTESTATION_SIGNER_ALGORITHM` | `Ed25519` | `Ed25519`, `ECDSA-P256` |
| `ATTESTATION_KEY_DIR` | `/etc/functionfly/keys` | Key storage directory (software) |
| `PKCS11_LIBRARY_PATH` | — | PKCS#11 shared library path |
| `PKCS11_SLOT_ID` | `0` | HSM slot ID (0 = first available) |
| `PKCS11_KEY_LABEL` | `ff-attestation` | Key label in the HSM |
| `PKCS11_PIN` | — | HSM token PIN |
| `AWS_KMS_CMK_ID` | — | KMS key ARN or alias |
| `AWS_KMS_REGION` | `us-east-1` | AWS region |
| `AWS_KMS_ENDPOINT` | — | Custom endpoint (LocalStack) |

### PKCS#11 HSM Support

Compatible with any PKCS#11 v2/v3 compliant HSM:

- **YubiHSM 2** — USB hardware security module
- **AWS CloudHSM** — Cloud-based HSM cluster
- **Thales Luna** — Enterprise HSM
- **SoftHSM** — Software HSM for testing

The signer auto-detects Ed25519 capability and falls back to ECDSA-P256 if the HSM doesn't support EdDSA.

### AWS KMS Support

Uses asymmetric KMS keys. The private key never leaves KMS — all signing operations are remote API calls.

Supported key specs: `ECC_NIST_P256`, `RSA_2048`, `RSA_3072`, `RSA_4096`.

The public key is fetched once at startup and cached for local verification.

### Signature Verification

Retrieve the signing public key:

```bash
curl https://api.functionfly.com/v1/trust/attestations/public-key
```

```json
{
  "public_key": "a1b2c3...",
  "key_id": "ff_att_v1_d4e5f6a7b8c9",
  "algorithm": "Ed25519",
  "key_encoding": "hex"
}
```

---

## Hash Chain

Each attestation links to its predecessor via `previous_hash`, forming an append-only chain. This makes the attestation log tamper-evident — inserting or modifying any attestation breaks the chain.

```
Attestation 0                    Attestation 1                    Attestation 2
┌──────────────┐                ┌──────────────┐                ┌──────────────┐
│ proof_hash_0 │───────────────►│ proof_hash_1 │───────────────►│ proof_hash_2 │
│ previous: "" │                │ previous: h0 │                │ previous: h1 │
└──────────────┘                └──────────────┘                └──────────────┘
```

### Proof Hash Calculation

The proof hash is SHA-256 of a deterministic JSON serialization:

```json
{
  "FunctionID": "550e8400-...",
  "FunctionVersion": "1.2.0",
  "Type": "verification",
  "Title": "Function verified (standard)",
  "Description": "Passed standard verification",
  "AttesterID": "user-uuid",
  "AttestedAt": 1719672000000000000,
  "Results": "{\"trust_score\":85}",
  "CodeHash": "sha256-of-source",
  "InputHash": "sha256-of-input",
  "OutputHash": "sha256-of-output"
}
```

### Per-Step Hashes

For execution attestations, three additional hashes capture exactly what was executed:

| Hash | Field | Computed From |
|------|-------|---------------|
| Code Hash | `code_hash` | SHA-256 of function source code at execution time |
| Input Hash | `input_hash` | SHA-256 of serialized input parameters |
| Output Hash | `output_hash` | SHA-256 of serialized execution output |

These are included in the proof hash calculation, so any tampering with code, inputs, or outputs invalidates the attestation.

---

## Merkle Audit Trail

All attestations are appended to a Merkle tree (RFC 6962). This provides:

1. **Inclusion proofs** — prove a specific attestation is in the log
2. **Consistency proofs** — prove the log is append-only (no entries were removed or modified)
3. **Signed tree heads** — periodic snapshots signed by the platform key

### Tree Structure

```
            Root Hash
           /         \
      H(01||AB)     H(01||CD)
       /    \        /    \
   H(0||A) H(0||B) H(0||C) H(0||D)
     |       |       |       |
   Leaf 0  Leaf 1  Leaf 2  Leaf 3
```

- **Leaf hash**: `SHA256(0x00 || attestation_data)`
- **Interior hash**: `SHA256(0x01 || left_hash || right_hash)`

### API Endpoints

```bash
# Get current tree head (signed root)
GET /v1/trust/merkle/head

# Get just the root hash
GET /v1/trust/merkle/root

# Get inclusion proof for leaf at index 42
GET /v1/trust/merkle/inclusion?leaf_index=42

# Get consistency proof from old size to current
GET /v1/trust/merkle/consistency?old_size=100

# Verify an inclusion proof
POST /v1/trust/merkle/verify/inclusion
```

### Tree Head Response

```json
{
  "tree_size": 1547,
  "root_hash": "a1b2c3d4e5f6...",
  "previous_hash": "f6e5d4c3b2a1...",
  "timestamp": "2026-06-29T15:00:00Z",
  "signature": "ed25519_signature_of_tree_head",
  "public_key_id": "ff_att_v1_d4e5f6a7b8c9"
}
```

### Inclusion Proof Response

```json
{
  "leaf_index": 42,
  "leaf_hash": "hash_of_leaf_42",
  "tree_size": 1547,
  "root_hash": "a1b2c3d4e5f6...",
  "path": ["sibling_hash_0", "sibling_hash_1", "..."]
}
```

---

## Multi-Agent Chain of Custody

When Agent A delegates execution to Agent B via `ctx.delegate()`, the platform automatically creates a **delegation attestation** that records the delegation event. Multiple delegations form a **chain of custody** — a verifiable record of which agent did what, in what order, and with what trust level.

### How It Works

```
Agent A (trust: 85)
  │
  │  ctx.delegate("func-B", input, { min_trust_score: 70 })
  │
  ├── Delegation Attestation (depth: 0, chain: chain_abc123)
  │   ├── delegator: Agent A
  │   ├── delegatee: func-B
  │   ├── input_hash: SHA256(input)
  │   └── trust_score: 85
  │
  ▼
Agent B (trust: 92)
  │
  │  ctx.delegate("func-C", output, { min_trust_score: 80 })
  │
  ├── Delegation Attestation (depth: 1, chain: chain_abc123)
  │   ├── delegator: func-B
  │   ├── delegatee: func-C
  │   ├── parent: att_... (Agent A's delegation)
  │   ├── input_hash: SHA256(output)
  │   └── trust_score: 92
  │
  ▼
Agent C (trust: 88)
  └── Final execution
```

### Delegation Attestation Fields

| Field | Description |
|-------|-------------|
| `delegation_chain_id` | Groups all attestations in one delegation sequence |
| `parent_attestation_id` | Links to the previous attestation in the chain |
| `delegation_depth` | Number of hops from the original caller (0 = originator) |
| `delegator_function_id` | UUID of the function that delegated |
| `delegator_agent_id` | Identity of the agent that initiated delegation |
| `delegator_trust_score` | Trust score of the delegator at delegation time |
| `delegation_input_hash` | SHA-256 of the input passed to the delegate |
| `delegation_output_hash` | SHA-256 of the output from the delegate |

### Chain of Custody API

```bash
# Get full chain of custody
GET /v1/trust/delegation/chain/{chain_id}

# Verify chain cryptographic integrity
GET /v1/trust/delegation/chain/{chain_id}/verify

# Get all chains a function participated in
GET /v1/trust/delegation/function/{function_id}
```

### Chain Verification

Chain verification checks:
1. Each attestation's proof hash is valid (data integrity)
2. Delegation depth is correctly ordered (0, 1, 2, ...)
3. Parent attestation IDs link correctly (no broken chain)
4. All signatures are valid (Ed25519/ECDSA)
5. No attestations were inserted, removed, or reordered

### Runtime Usage

**TypeScript**:
```typescript
export default async function handler(request, env, context) {
  // Delegate to another function — creates delegation attestation automatically
  const result = await context.delegate("trusted-analyzer", {
    data: request.body,
  }, {
    min_trust_score: 80,
    timeout_ms: 5000,
  });

  return { status: 200, headers: {}, body: result };
}
```

**Python**:
```python
from flypy import delegate

result = delegate("trusted-analyzer", {"data": input_data}, {"min_trust_score": 80})
```

---

## Zero-Knowledge Proofs

The attestation system supports three ZK proof types using Pedersen commitments on the BN254 curve (via [gnark v0.15.0](https://github.com/Consensys/gnark), Consensys — the most audited Go ZK library).

### Proof Types

#### 1. Existence Proof

**Proves**: "I know an attestation of type T with status S for function F"
**Without revealing**: Which attestation, its contents, or position in the log

**Protocol** (Schnorr-style with Fiat-Shamir):
1. Prover commits: `A = hash(type) * G + r_v * H`, `B = status * G + r_s * H`
2. Challenge: `e = SHA256(A || B || type || functionID)`
3. Response: `z_v = r_v + e * hash(type)`, `z_s = r_s + e * status`
4. Verifier checks challenge matches

#### 2. Inclusion Proof

**Proves**: "A committed value exists in the Merkle tree"
**Without revealing**: The leaf data, position, or value

Uses Pedersen commitment to the leaf value + Merkle audit path.

#### 3. Range Proof

**Proves**: "A committed value lies within [lo, hi]"
**Without revealing**: The actual value

Uses bit decomposition: commit to each bit of `(value - lo)`, prove each bit is 0 or 1.

### Curve Parameters

- **Curve**: BN254 (Barreto-Naehrig, 254-bit)
- **Generator G**: Standard BN254 G1 generator
- **Generator H**: `HashToG1("functionfly-attestation-pedersen-h", "ff-domain")`
- **Commitment**: `C = value * G + blinding * H`
- **Challenge**: Fiat-Shamir hash of commitments + public data

### Client-Side Verification

ZK proofs can be verified entirely client-side without any server round-trip. The `@functionfly/trust` SDK includes all verification logic.

---

## API Reference

### Authentication

| Endpoint | Auth | Scope |
|----------|------|-------|
| GET endpoints (read) | API key (`fft_*`) | `trust:read` (default) |
| POST create | API key or JWT | `attestation:create` scope, `startup`+ tier |
| POST revoke | API key or JWT | `attestation:revoke` scope, `business`+ tier |
| Merkle endpoints | API key | `trust:read` (default) |
| Public key | API key | `trust:read` (default) |

JWT-authenticated users are gated by platform plan:
- **Starter+**: View attestations, verify chains
- **Professional+**: Create attestations, manage policies
- **Enterprise+**: Revoke attestations

### Create Attestation

```bash
POST /v1/trust/attestations
Authorization: Bearer <api_key_or_jwt>
Content-Type: application/json

{
  "function_id": "550e8400-e29b-41d4-a716-446655440000",
  "function_version": "1.2.0",
  "type": "security_scan",
  "title": "Security scan passed",
  "description": "No vulnerabilities found in dependency audit",
  "results": {
    "vulnerabilities": 0,
    "scanner": "functionfly-scanner",
    "scan_duration_ms": 4500
  },
  "code_hash": "a1b2c3...",
  "input_hash": "d4e5f6...",
  "output_hash": "789abc...",
  "valid_until": "2027-06-29T00:00:00Z"
}
```

**Response** (201 Created):
```json
{
  "id": "550e8400-...",
  "attestation_id": "att_a1b2c3d4e5f6g7h8i9j0",
  "function_id": "550e8400-...",
  "type": "security_scan",
  "status": "valid",
  "title": "Security scan passed",
  "proof_hash": "sha256_hex",
  "signature": "ed25519_signature_hex",
  "public_key_id": "ff_att_v1_d4e5f6a7b8c9",
  "code_hash": "a1b2c3...",
  "input_hash": "d4e5f6...",
  "output_hash": "789abc...",
  "previous_hash": "previous_attestation_hash",
  "is_valid": true,
  "signature_valid": true,
  "chain_valid": true,
  "attested_at": "2026-06-29T15:00:00Z"
}
```

### Revoke Attestation

```bash
POST /v1/trust/attestations/{attestation_id}/revoke
Authorization: Bearer <api_key_or_jwt>

{
  "reason": "Security vulnerability discovered in dependency"
}
```

### List Attestations

```bash
GET /v1/trust/attestations?function_id=550e8400-...&type=verification&status=valid
```

### Verify Attestation

```bash
GET /v1/trust/attestations/{attestation_id}/verify
```

**Response**:
```json
{
  "attestation_id": "att_...",
  "integrity_verified": true,
  "signature_verified": true,
  "proof_hash": "sha256_hex",
  "signature": "ed25519_hex",
  "public_key_id": "ff_att_v1_...",
  "algorithm": "Ed25519",
  "verified_at": "2026-06-29T15:00:00Z"
}
```

### Get Attestation Chain

```bash
GET /v1/trust/attestations/chain/{function_id}
```

**Response**:
```json
{
  "function_id": "550e8400-...",
  "chain_length": 12,
  "chain_valid": true,
  "signing_algorithm": "Ed25519",
  "attestations": [
    {
      "attestation_id": "att_...",
      "type": "verification",
      "status": "valid",
      "proof_hash": "...",
      "previous_hash": "",
      "signature": "...",
      "integrity_verified": true,
      "chain_link_valid": true
    }
  ]
}
```

### Verify Chain

```bash
GET /v1/trust/attestations/chain/{function_id}/verify
```

### Get Signing Public Key

```bash
GET /v1/trust/attestations/public-key
```

---

## SDK Packages

### `@functionfly/trust`

Dedicated trust SDK with attestation management, Merkle verification, and ZK proofs.

```bash
npm install @functionfly/trust
```

```typescript
import { TrustClient, verifyInclusion, verifyExistenceProof } from '@functionfly/trust';

const client = new TrustClient({ apiKey: 'fft_...' });

// List attestations
const { attestations } = await client.listAttestations('function-id');

// Create attestation
const att = await client.createAttestation({
  function_id: 'function-id',
  type: 'security_scan',
  title: 'Security scan passed',
  code_hash: 'sha256_of_source',
  input_hash: 'sha256_of_input',
  output_hash: 'sha256_of_output',
});

// Verify Merkle inclusion client-side
const proof = await client.getMerkleInclusionProof(0);
const valid = await verifyInclusion(
  proof.leaf_hash, proof.leaf_index,
  proof.tree_size, proof.path, proof.root_hash,
);

// Generate and verify ZK existence proof
const zkProof = client.generateExistenceProof('verification', 'func-id', 'valid');
const zkValid = verifyExistenceProof(zkProof);
```

**Exports**:
- `TrustClient` — Full API client
- `verifyInclusion()` — Client-side Merkle inclusion verification
- `buildInclusionProof()` — Build inclusion proof from leaf hashes
- `computeRoot()` — Compute Merkle root from leaf hashes
- `verifyExistenceProof()` — Verify ZK existence proof
- `verifyInclusionProof()` — Verify ZK inclusion proof
- `verifyRangeProof()` — Verify ZK range proof

### `@functionfly/sdk`

Unified SDK that re-exports `@functionfly/agent` and `@functionfly/trust`:

```bash
npm install @functionfly/sdk
```

```typescript
import { AgentClient, TrustClient, verifyInclusion } from '@functionfly/sdk';
```

---

## Plan Tiers

| Feature | Free | Starter | Professional | Enterprise |
|---------|------|---------|-------------|------------|
| View attestations | ✓ | ✓ | ✓ | ✓ |
| Verify chains | ✓ | ✓ | ✓ | ✓ |
| Merkle audit trail | ✓ | ✓ | ✓ | ✓ |
| Submit verification | — | ✓ | ✓ | ✓ |
| Submit trust reports | — | ✓ | ✓ | ✓ |
| Create attestations | — | — | ✓ | ✓ |
| Manage trust policies | — | — | ✓ | ✓ |
| Revoke attestations | — | — | — | ✓ |

### Trust API Partner Tiers

| Tier | Attestation Read | Create | Revoke |
|------|-----------------|--------|--------|
| Developer | ✓ | — | — |
| PAYG | ✓ | — | — |
| Startup ($49/mo) | ✓ | ✓ | — |
| Business ($199/mo) | ✓ | ✓ | ✓ |
| Enterprise (custom) | ✓ | ✓ | ✓ |

---

## Performance

All numbers below are from Go benchmarks on Intel i9-12900K, measured with `go test -bench`. Run `go test -bench=. -benchmem ./internal/storage/trustapi/` to reproduce.

### Attestation Operations

| Operation | Latency | Allocations |
|-----------|---------|-------------|
| Calculate proof hash (SHA-256) | 1.2 µs | 7 allocs |
| Verify integrity (rehash + compare) | 0.9 µs | 7 allocs |
| Sign attestation (Ed25519) | 15.5 µs | 10 allocs |
| Verify signature (Ed25519) | 32 µs | 9 allocs |
| Sign attestation (ECDSA-P256) | 27 µs | 89 allocs |
| Verify signature (ECDSA-P256) | 54 µs | 33 allocs |

### Merkle Tree

| Operation | Latency | Allocations |
|-----------|---------|-------------|
| Leaf hash (SHA-256 with prefix) | 174 ns | 3 allocs |
| Interior node hash | 238 ns | 5 allocs |
| Compute root (100 leaves) | 26 µs | 522 allocs |
| Compute root (1,000 leaves) | 258 µs | 5,049 allocs |
| Compute root (10,000 leaves) | 2.9 ms | 50,100 allocs |
| Build inclusion proof (1,000 leaves) | 274 µs | 5,107 allocs |
| Verify inclusion proof (1,000 leaves) | 2.5 µs | 50 allocs |

### Zero-Knowledge Proofs (BN254)

| Operation | Latency | Allocations |
|-----------|---------|-------------|
| Pedersen commitment | 70 µs | 6 allocs |
| Pedersen verification | 52 µs | 4 allocs |
| Generate existence proof | 144 µs | 29 allocs |
| Verify existence proof | 0.7 µs | 14 allocs |
| Generate range proof [0,100] | 648 µs | 100 allocs |
| Verify range proof | 188 ns | 5 allocs |

### Raw Crypto Baselines

| Operation | Latency |
|-----------|---------|
| SHA-256 (64 bytes) | 147 ns |
| Ed25519 sign | 14.3 µs |
| Ed25519 verify | 35.5 µs |

---

## Architecture

### CLI Commands (`ff-cli`)

| Command | Description |
|---------|-------------|
| `ff trust verify [author/name]` | Verify attestation chain integrity |
| `ff trust verify --full` | Verify chain + Merkle root + delegation chains |
| `ff trust attestations [author/name]` | List attestations with type/status filters |
| `ff trust set-tier [author/name] --tier <tier>` | Set trust tier (untrusted/trusted/verified/highly_trusted) |
| `ff trust chain <chain-id>` | View delegation chain of custody |
| `ff trust chain <chain-id> --verify` | Verify delegation chain integrity |
| `ff trust merkle` | View Merkle audit trail tree head |
| `ff trust public-key` | Get attestation signing public key |

All commands support `--json` for machine-readable output and honor the global `--format json` flag.

### Files

| Layer | Path | Purpose |
|-------|------|---------|
| **Models** | `internal/storage/trustapi/revocation_models.go` | Attestation struct, DTOs, proof hash |
| **Signer** | `internal/storage/trustapi/signer_software.go` | Ed25519/ECDSA software signer |
| **Signer** | `internal/storage/trustapi/signer_pkcs11.go` | PKCS#11 HSM signer |
| **Signer** | `internal/storage/trustapi/signer_aws.go` | AWS KMS signer |
| **Signer** | `internal/storage/trustapi/attestation_signer.go` | Signer interface + factory |
| **Merkle** | `internal/storage/trustapi/merkle.go` | Merkle tree computation |
| **Merkle** | `internal/storage/trustapi/merkle_repository.go` | Merkle persistence |
| **ZK** | `internal/storage/trustapi/zk_proofs.go` | Pedersen commitments + ZK proofs |
| **Repository** | `internal/storage/trustapi/revocation_repository.go` | Attestation CRUD |
| **Handlers** | `internal/api/handlers/trustapi/revocation_handlers.go` | HTTP handlers |
| **Routes** | `internal/api/routes_trustapi.go` | Route registration |
| **Migration** | `migrations/20260629150000_*` | Per-step hash columns |
| **Migration** | `migrations/20260629160000_*` | Merkle tables |
| **SDK** | `sdk/js/packages/trust/` | `@functionfly/trust` TypeScript package |
| **SDK** | `sdk/js/packages/sdk/` | `@functionfly/sdk` unified package |
| **Dashboard** | `web/dashboard/src/components/trust/AttestationPanel.tsx` | Attestation management UI |
| **API Client** | `web/dashboard/src/api/trustapi.ts` | Dashboard API client |

### Database Tables

| Table | Purpose |
|-------|---------|
| `trust_attestations` | Attestation records (immutable) |
| `merkle_nodes` | Merkle tree nodes (level, index, hash) |
| `merkle_tree_heads` | Signed Merkle tree head snapshots |
| `zk_proofs` | Stored ZK proofs |

### Dependencies

| Library | Version | Purpose |
|---------|---------|---------|
| gnark | v0.15.0 | ZK proof system (BN254 curve) |
| gnark-crypto | v0.20.1 | Elliptic curve arithmetic |
| miekg/pkcs11 | v1.1.2 | PKCS#11 HSM integration |
| aws-sdk-go-v2/service/kms | v1.53.5 | AWS KMS integration |

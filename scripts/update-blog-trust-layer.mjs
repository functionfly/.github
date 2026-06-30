import {createClient} from '@sanity/client'

const client = createClient({
  projectId: 'sg1k76uk',
  dataset: 'production',
  apiVersion: '2024-01-01',
  token: process.env.SANITY_AUTH_TOKEN,
  useCdn: false,
})

const BODY = `As AI agents become more autonomous, the need for trust and verification grows. When an agent can read your database, call external APIs, spend money, and modify infrastructure, you need cryptographic proof that it did exactly what it was supposed to do -- nothing more, nothing less.

The Trust Layer provides cryptographic attestation for function properties on FunctionFly. It creates immutable, independently verifiable records that prove what code ran, what inputs it received, what outputs it produced, and whether it was tampered with. This post explains how it works and why it matters for production AI systems.

## The Trust Problem

When an AI agent executes actions on your behalf, several questions become critical:

- Did the agent run the code you deployed, or was it tampered with?
- Did it access only the data it was authorized to access?
- Did it actually call the external APIs it claims to have called?
- Were the results it returned produced by the model you specified, or a cheaper substitute?
- Can you prove any of this to a regulator, auditor, or customer?

Traditional logging answers some of these questions, but logs can be altered, deleted, or forged. You are trusting the infrastructure operator to maintain log integrity. For high-stakes applications -- financial transactions, healthcare decisions, legal document processing -- this level of trust is insufficient.

The Trust Layer solves this by making function attestations verifiable through cryptographic records that cannot be forged or modified after the fact.

## How Attestation Works

An attestation is a signed record that asserts a property about a function at a specific point in time. When you use the SDK to define steps in your function, each step can generate an attestation containing:

- **Code hash**: A SHA-256 hash of the executed code, proving which version ran
- **Input hash**: A hash of the function inputs, proving what data was provided
- **Output hash**: A hash of the function outputs, proving what results were produced
- **Proof hash**: A SHA-256 hash over all attestation data, used for integrity verification
- **Signature**: An Ed25519 (or ECDSA-P256 via HSM) signature over the proof hash

Six attestation types are supported: \`verification\`, \`security_scan\`, \`code_review\`, \`execution\`, \`compliance\`, and \`signature\`. Each type serves a different purpose in establishing function trustworthiness.

\`\`\`typescript
import { defineCapsule, createCapsuleContext } from "@functionfly/trust";

export default defineCapsule({
  name: "payment-processor",
  trust: { tier: 2 }, // Signed attestations per step
  async handler(ctx) {
    const result = await ctx.step("process-payment", async () => {
      return processPayment(ctx.input);
    });

    // Get the attestation generated for this step
    const attestation = ctx.getAttestation("process-payment");
    console.log("Attestation ID:", attestation?.attestationId);

    return { paymentId: result.id, attestation };
  },
});
\`\`\`

Each \`ctx.step()\` call generates an attestation (at Tier 2+) that records the step name, duration, success/failure, and per-step input/output hashes. The attestation is signed by the runtime using Ed25519 keys that can be persisted to disk, stored in a PKCS#11 HSM (YubiHSM, AWS CloudHSM, Thales Luna), or managed via AWS KMS.

## Verification Tiers

The Trust Layer supports four verification tiers. You choose the tier based on your risk tolerance and compliance requirements:

### Tier 1: Execution Logging

The simplest tier. Every function invocation is logged with inputs, outputs, and timestamps. Logs are stored in append-only storage with integrity checksums. Suitable for internal debugging and basic audit trails.

### Tier 2: Signed Attestations

Every attestation is signed with Ed25519 (or ECDSA-P256/RSA-PSS via HSM). Signatures can be verified independently using the published public key at \`GET /v1/trust/keys\`. This tier proves that the FunctionFly runtime generated the record and it has not been tampered with.

### Tier 3: Merkle Audit Trail

Attestations are organized into an RFC 6962 Merkle tree. Each tree head is signed and can be used to verify that a specific attestation is included in the log and that no attestations have been removed or modified. Inclusion proofs and consistency proofs are available via the API.

### Tier 4: Zero-Knowledge Proofs

The highest tier. Uses Pedersen commitments on the BN254 curve (via gnark v0.15.0) to generate proofs that an attestation exists, that a value lies within a range, or that a committed value is in the Merkle tree -- all without revealing the underlying data. Three proof types are supported: existence, inclusion, and range proofs.

\`\`\`typescript
// Configure the trust tier for your Capsule
export default defineCapsule({
  name: "compliance-agent",
  trust: {
    tier: 3, // Merkle audit trail
    publishInterval: "1h", // Publish Merkle roots every hour
    retention: "7years", // SOX compliance
  },
  async handler(ctx) {
    // All steps are automatically attested at the configured tier
    const decision = await ctx.step("evaluate", () => evaluateApplication(ctx.input));
    await ctx.step("decide", () => makeDecision(decision));
    return decision;
  },
});
\`\`\`

## Verifying Attestations

Any third party can verify a FunctionFly attestation without access to your code or data. The verification process checks:

1. The proof hash is consistent with the attestation data (integrity)
2. The signature is valid against the published public key (authenticity)
3. The hash chain links are correct (tamper evidence)
4. The Merkle inclusion proof is valid (audit trail)

\`\`\`typescript
import { TrustClient, verifyInclusion } from "@functionfly/trust";

const client = new TrustClient({ apiKey: "fft_..." });

// Verify a specific attestation
const result = await client.verifyAttestation("att-abc123");
console.log(result);
// {
//   attestation_id: "att-abc123",
//   integrity_verified: true,
//   signature_verified: true,
//   proof_hash: "a1b2c3...",
//   algorithm: "Ed25519"
// }

// Verify Merkle inclusion client-side
const proof = await client.getMerkleInclusionProof(42);
const valid = await verifyInclusion(
  proof.leaf_hash, proof.leaf_index,
  proof.tree_size, proof.path, proof.root_hash,
);

// Get the signing public key for external verification
const pubKey = await client.getPublicKeyByEnv("production");
\`\`\`

This verification can be performed by your customers, regulators, or any automated system. It does not require FunctionFly credentials or access to your account.

## Real-World Use Cases

### Financial Services

A payment processing agent must prove that every transaction was authorized, calculated correctly, and recorded. The Trust Layer provides per-transaction attestations that auditors can verify independently. Combined with Tier 3 (Merkle audit trail), this satisfies SOX and PCI-DSS requirements for transaction integrity.

### Healthcare

A clinical decision support agent recommends treatments based on patient data. Regulators need to verify that the agent used the correct model, accessed only authorized patient records, and followed clinical guidelines. Tier 4 (zero-knowledge proofs) allows this verification without exposing patient data.

### Legal and Compliance

A contract analysis agent reviews legal documents and flags risks. The Trust Layer proves which documents were reviewed, what model was used, and what risks were identified. This creates a defensible audit trail for malpractice insurance and regulatory inquiries.

## Integration with the Trust API

FunctionFly exposes a REST API for querying and verifying attestations:

\`\`\`
GET  /v1/trust/attestations/{function_id}          List attestations
GET  /v1/trust/verify/{attestationId}              Verify an attestation
GET  /v1/trust/attestations/public-key             Get signing public key
GET  /v1/trust/keys                                Alias for public key
GET  /v1/trust/merkle/root                         Current Merkle root hash
GET  /v1/trust/merkle/head                         Signed Merkle tree head
GET  /v1/trust/merkle/inclusion?leaf_index=N       Inclusion proof
GET  /v1/trust/merkle/proof/{attestationId}        Inclusion proof alias
GET  /v1/trust/merkle/consistency?old_size=N       Consistency proof
GET  /v1/trust/delegation/chain/{chain_id}         Delegation chain of custody
POST /v1/trust/attestations                        Create attestation (Pro+)
POST /v1/trust/attestations/{id}/revoke            Revoke attestation (Enterprise)
\`\`\`

Verification endpoints are publicly accessible. Attestation creation and retrieval require authentication with the appropriate plan tier or API key scope.

## Trust for Multi-Agent Systems

When multiple agents collaborate, the Trust Layer creates a chain of custody across agent boundaries. Each \`ctx.delegate()\` call creates a delegation attestation that records who delegated to whom, at what trust level, and with what input.

\`\`\`typescript
// Delegation creates an attested chain of custody
const subtask = await ctx.delegate("specialist-agent", {
  task: "analyze-contract",
  documentId: "contract-123",
}, {
  min_trust_score: 80,
  timeout_ms: 5000,
});

// The delegation itself is attested with:
// - delegator_function_id (who delegated)
// - delegator_agent_id (agent identity)
// - delegator_trust_score (trust at delegation time)
// - delegation_input_hash (SHA-256 of input)
// - delegation_chain_id (links all hops)
\`\`\`

This chain of custody is critical for accountability. If a multi-agent system produces a wrong answer, you can trace exactly which agent made the error, what data it used, and what trust level it had.

## Performance

All numbers below are measured with Go benchmarks on Intel i9-12900K (run \`go test -bench=. ./internal/storage/trustapi/\` to reproduce):

- **Proof hash (SHA-256)**: 1.2 us
- **Ed25519 sign**: 15.5 us
- **Ed25519 verify**: 32 us
- **ECDSA-P256 sign**: 27 us
- **Merkle leaf hash**: 174 ns
- **Merkle root (1K leaves)**: 258 us
- **Merkle verify inclusion (1K)**: 2.5 us
- **Pedersen commit (BN254)**: 70 us
- **ZK existence proof verify**: 0.7 us
- **ZK range proof verify**: 188 ns

For most AI workloads, the attestation overhead is insignificant compared to LLM inference time.

## Getting Started

Trust is enabled by default on all FunctionFly deployments at Tier 1. To increase the tier:

\`\`\`bash
ff trust set-tier alice/my-fn --tier verified
ff trust set-tier --agent my-agent --tier verified
\`\`\`

To verify an attestation from the command line:

\`\`\`bash
ff trust verify att-abc123
ff trust verify alice/my-fn --full
ff trust attestations alice/my-fn
ff trust chain chain-abc123
ff trust merkle
ff trust public-key
\`\`\`

For the full API reference and compliance guides, see the [Trust Layer documentation](https://docs.functionfly.com/trust). For the underlying architecture, read the [Attestation System deep dive](https://docs.functionfly.com/attestation-system) which covers cryptographic signing, Merkle trees, and zero-knowledge proofs.`

async function main() {
  const result = await client
    .patch('blogPost-d1e2f3a4-b5c6-7890-abcd-ef1234567894')
    .set({body: BODY})
    .commit()

  console.log('Updated:', result._id)
}

main().catch(console.error)

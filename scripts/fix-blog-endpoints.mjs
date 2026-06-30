import {createClient} from '@sanity/client'

const client = createClient({
  projectId: 'sg1k76uk',
  dataset: 'production',
  apiVersion: '2024-01-01',
  token: process.env.SANITY_AUTH_TOKEN,
  useCdn: false,
})

const OLD = `GET  /v1/trust/attestations/{function_id}          List attestations
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
POST /v1/trust/attestations/{id}/revoke            Revoke attestation (Enterprise)`

const NEW = `GET  /v1/trust/attestations?function_id=ID           List attestations
GET  /v1/trust/attestations/{attestation_id}       Get single attestation
GET  /v1/trust/attestations/{id}/verify            Verify integrity + signature
GET  /v1/trust/attestations/chain/{function_id}    Get attestation chain
GET  /v1/trust/attestations/chain/{id}/verify      Verify attestation chain
GET  /v1/trust/attestations/public-key             Get signing public key
GET  /v1/trust/keys                                Alias for public key
GET  /v1/trust/merkle/root                         Current Merkle root hash
GET  /v1/trust/merkle/head                         Signed Merkle tree head
GET  /v1/trust/merkle/inclusion?leaf_index=N       Inclusion proof by index
GET  /v1/trust/merkle/proof/{attestation_id}       Inclusion proof by attestation ID
GET  /v1/trust/merkle/consistency?old_size=N       Consistency proof
GET  /v1/trust/delegation/chain/{chain_id}         Delegation chain of custody
POST /v1/trust/attestations                        Create attestation (Pro+)
POST /v1/trust/attestations/{id}/revoke            Revoke attestation (Enterprise)`

async function main() {
  const doc = await client.fetch('*[_type=="blogPost" && slug.current=="trust-layer-for-ai-agents"][0]{body}')
  if (!doc) { console.log('Post not found'); return }
  const newBody = doc.body.replace(OLD, NEW)
  if (newBody === doc.body) { console.log('No changes needed or old text not found'); return }
  await client.patch('blogPost-d1e2f3a4-b5c6-7890-abcd-ef1234567894').set({body: newBody}).commit()
  console.log('Updated endpoint paths')
}

main().catch(console.error)

package trustapi

import (
	"crypto/ed25519"
	"encoding/json"
	"fmt"
	"math/big"
	"testing"
	"time"

	"github.com/google/uuid"
)

// ============================================
// Attestation Signing Benchmarks
// ============================================

func BenchmarkCalculateProofHash(b *testing.B) {
	att := &TrustAttestation{
		FunctionID:      uuid.New(),
		FunctionVersion: "1.2.0",
		Type:            string(AttestationTypeVerification),
		Title:           "Function verified (standard)",
		Description:     "Passed standard verification",
		AttesterID:      uuid.New(),
		AttestedAt:      time.Now(),
		Results:         json.RawMessage(`{"trust_score":85,"scanner":"functionfly"}`),
		CodeHash:        "a1b2c3d4e5f6a7b8c9d0e1f2a3b4c5d6e7f8a9b0c1d2e3f4a5b6c7d8e9f0a1b2",
		InputHash:       "b2c3d4e5f6a7b8c9d0e1f2a3b4c5d6e7f8a9b0c1d2e3f4a5b6c7d8e9f0a1b2c3",
		OutputHash:      "c3d4e5f6a7b8c9d0e1f2a3b4c5d6e7f8a9b0c1d2e3f4a5b6c7d8e9f0a1b2c3d4",
	}

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_ = att.CalculateProofHash()
	}
}

func BenchmarkVerifyIntegrity(b *testing.B) {
	att := &TrustAttestation{
		FunctionID:      uuid.New(),
		FunctionVersion: "1.2.0",
		Type:            string(AttestationTypeVerification),
		Title:           "Function verified (standard)",
		Description:     "Passed standard verification",
		AttesterID:      uuid.New(),
		AttestedAt:      time.Now(),
		Results:         json.RawMessage(`{"trust_score":85}`),
	}
	att.ProofHash = att.CalculateProofHash()

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_ = att.VerifyIntegrity()
	}
}

func BenchmarkSignAttestation_Software(b *testing.B) {
	signer := newEphemeralSoftwareSigner(AlgEd25519)
	att := &TrustAttestation{
		FunctionID:      uuid.New(),
		FunctionVersion: "1.0.0",
		Type:            string(AttestationTypeVerification),
		Title:           "Benchmark attestation",
		AttesterID:      uuid.New(),
		AttestedAt:      time.Now(),
		Results:         json.RawMessage(`{}`),
	}

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		att.AttestedAt = time.Now()
		_ = signer.SignAttestation(att)
	}
}

func BenchmarkVerifyAttestationSignature_Software(b *testing.B) {
	signer := newEphemeralSoftwareSigner(AlgEd25519)
	att := &TrustAttestation{
		FunctionID:      uuid.New(),
		FunctionVersion: "1.0.0",
		Type:            string(AttestationTypeVerification),
		Title:           "Benchmark attestation",
		AttesterID:      uuid.New(),
		AttestedAt:      time.Now(),
		Results:         json.RawMessage(`{}`),
	}
	_ = signer.SignAttestation(att)

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_, _ = signer.VerifyAttestationSignature(att)
	}
}

func BenchmarkSignAttestation_ECDSA(b *testing.B) {
	signer := newEphemeralSoftwareSigner(AlgECDSA)
	att := &TrustAttestation{
		FunctionID:      uuid.New(),
		FunctionVersion: "1.0.0",
		Type:            string(AttestationTypeVerification),
		Title:           "Benchmark attestation",
		AttesterID:      uuid.New(),
		AttestedAt:      time.Now(),
		Results:         json.RawMessage(`{}`),
	}

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		att.AttestedAt = time.Now()
		_ = signer.SignAttestation(att)
	}
}

func BenchmarkVerifyAttestationSignature_ECDSA(b *testing.B) {
	signer := newEphemeralSoftwareSigner(AlgECDSA)
	att := &TrustAttestation{
		FunctionID:      uuid.New(),
		FunctionVersion: "1.0.0",
		Type:            string(AttestationTypeVerification),
		Title:           "Benchmark attestation",
		AttesterID:      uuid.New(),
		AttestedAt:      time.Now(),
		Results:         json.RawMessage(`{}`),
	}
	_ = signer.SignAttestation(att)

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_, _ = signer.VerifyAttestationSignature(att)
	}
}

// ============================================
// Merkle Tree Benchmarks
// ============================================

func BenchmarkMerkleLeafHash(b *testing.B) {
	data := []byte(`{"attestation_id":"att_abc123","function_id":"550e8400","type":"verification","proof_hash":"a1b2c3"}`)

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_ = MerkleLeafHash(data)
	}
}

func BenchmarkMerkleNodeHash(b *testing.B) {
	left := "a1b2c3d4e5f6a7b8c9d0e1f2a3b4c5d6e7f8a9b0c1d2e3f4a5b6c7d8e9f0a1b2"
	right := "b2c3d4e5f6a7b8c9d0e1f2a3b4c5d6e7f8a9b0c1d2e3f4a5b6c7d8e9f0a1b2c3"

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_, _ = MerkleNodeHash(left, right)
	}
}

func BenchmarkComputeRoot_100(b *testing.B) {
	leaves := generateLeaves(100)
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_ = ComputeRoot(leaves)
	}
}

func BenchmarkComputeRoot_1000(b *testing.B) {
	leaves := generateLeaves(1000)
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_ = ComputeRoot(leaves)
	}
}

func BenchmarkComputeRoot_10000(b *testing.B) {
	leaves := generateLeaves(10000)
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_ = ComputeRoot(leaves)
	}
}

func BenchmarkBuildInclusionProof_1000(b *testing.B) {
	leaves := generateLeaves(1000)
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_, _ = BuildInclusionProof(leaves, int64(i%1000))
	}
}

func BenchmarkVerifyInclusion_1000(b *testing.B) {
	leaves := generateLeaves(1000)
	proof, _ := BuildInclusionProof(leaves, 42)
	root := ComputeRoot(leaves)

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_ = VerifyInclusion(leaves[42], 42, 1000, proof, root)
	}
}

func BenchmarkBuildConsistencyProof_1000_from500(b *testing.B) {
	leaves := generateLeaves(1000)
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_, _ = BuildConsistencyProof(leaves, 500)
	}
}

// ============================================
// ZK Proof Benchmarks
// ============================================

func BenchmarkPedersenCommit(b *testing.B) {
	_ = GetPedersenSetup()
	val := big.NewInt(42)

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_, _, _ = PedersenCommit(val)
	}
}

func BenchmarkPedersenVerify(b *testing.B) {
	_ = GetPedersenSetup()
	val := big.NewInt(42)
	commitment, blinding, _ := PedersenCommit(val)

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_ = PedersenVerify(commitment, val, blinding)
	}
}

func BenchmarkGenerateExistenceProof(b *testing.B) {
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_, _ = GenerateExistenceProof("verification", uuid.New().String(), "valid")
	}
}

func BenchmarkVerifyExistenceProof(b *testing.B) {
	proof, _ := GenerateExistenceProof("verification", uuid.New().String(), "valid")

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_, _ = VerifyExistenceProof(proof)
	}
}

func BenchmarkGenerateRangeProof(b *testing.B) {
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_, _ = GenerateRangeProof(85, 0, 100)
	}
}

func BenchmarkVerifyRangeProof(b *testing.B) {
	proof, _ := GenerateRangeProof(85, 0, 100)

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_, _ = VerifyRangeProof(proof)
	}
}

// ============================================
// Raw Crypto Benchmarks (baseline comparisons)
// ============================================

func BenchmarkSHA256_64B(b *testing.B) {
	data := make([]byte, 64)
	for i := range data {
		data[i] = byte(i)
	}
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_ = GenerateSourceDataHash(data)
	}
}

func BenchmarkEd25519_Sign(b *testing.B) {
	pub, priv, _ := ed25519.GenerateKey(nil)
	msg := make([]byte, 64)
	_ = pub
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_ = ed25519.Sign(priv, msg)
	}
}

func BenchmarkEd25519_Verify(b *testing.B) {
	pub, priv, _ := ed25519.GenerateKey(nil)
	msg := make([]byte, 64)
	sig := ed25519.Sign(priv, msg)
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_ = ed25519.Verify(pub, msg, sig)
	}
}

// ============================================
// Helpers
// ============================================

func generateLeaves(n int) []string {
	leaves := make([]string, n)
	for i := 0; i < n; i++ {
		leaves[i] = MerkleLeafHash([]byte(fmt.Sprintf("leaf-%d-%d", i, time.Now().UnixNano())))
	}
	return leaves
}

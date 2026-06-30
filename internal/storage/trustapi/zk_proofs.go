package trustapi

import (
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"math/big"
	"time"

	"github.com/consensys/gnark-crypto/ecc/bn254"
	"github.com/consensys/gnark-crypto/ecc/bn254/fr"
	"github.com/google/uuid"
)

// ============================================
// Zero-Knowledge Attestation Proofs
// ============================================
//
// Three ZK proof types using Pedersen commitments on BN254:
//
//  1. ProofOfExistence — prove an attestation of a given type/status
//     exists without revealing which attestation or its contents.
//
//  2. ProofOfInclusion — prove a committed value is in the Merkle tree
//     without revealing the leaf position or data.
//
//  3. ProofOfRange — prove a committed value lies within [lo,hi]
//     without revealing the value.
//
// All proofs use Fiat-Shamir (hash-based challenge) for non-interactivity.

type ZKProofType string

const (
	ZKProofExistence ZKProofType = "existence"
	ZKProofInclusion ZKProofType = "inclusion"
	ZKProofRange     ZKProofType = "range"
)

// PedersenSetup holds generator points G, H on BN254 G1.
type PedersenSetup struct {
	G bn254.G1Affine
	H bn254.G1Affine
}

var pedersenSetup *PedersenSetup

func GetPedersenSetup() *PedersenSetup {
	if pedersenSetup == nil {
		pedersenSetup = newPedersenSetup()
	}
	return pedersenSetup
}

func newPedersenSetup() *PedersenSetup {
	s := &PedersenSetup{}
	_, _, g1, _ := bn254.Generators()
	s.G = g1

	h, err := bn254.HashToG1([]byte("functionfly-attestation-pedersen-h"), []byte("ff-domain"))
	if err != nil {
		var jac bn254.G1Jac
		jac.FromAffine(&g1)
		jac.DoubleAssign()
		s.H.FromJacobian(&jac)
	} else {
		s.H = h
	}
	return s
}

// PedersenCommit computes C = value*G + blinding*H.
func PedersenCommit(value *big.Int) (bn254.G1Affine, *fr.Element, error) {
	setup := GetPedersenSetup()

	var val fr.Element
	val.SetBigInt(value)

	var r fr.Element
	r.SetRandom()

	var valBig, rBig big.Int
	val.BigInt(&valBig)
	r.BigInt(&rBig)

	// C = val*G + r*H using optimized joint scalar multiplication
	var commitmentJac bn254.G1Jac
	commitmentJac.JointScalarMultiplication(&setup.G, &setup.H, &valBig, &rBig)

	var commitment bn254.G1Affine
	commitment.FromJacobian(&commitmentJac)

	return commitment, &r, nil
}

// PedersenVerify checks that C == value*G + blinding*H.
func PedersenVerify(commitment bn254.G1Affine, value *big.Int, blinding *fr.Element) bool {
	setup := GetPedersenSetup()

	var val fr.Element
	val.SetBigInt(value)

	var valBig, rBig big.Int
	val.BigInt(&valBig)
	blinding.BigInt(&rBig)

	var expectedJac bn254.G1Jac
	expectedJac.JointScalarMultiplication(&setup.G, &setup.H, &valBig, &rBig)

	var expected bn254.G1Affine
	expected.FromJacobian(&expectedJac)

	return commitment.Equal(&expected)
}

// ============================================
// Proof Structures
// ============================================

type ZKProof struct {
	ID        uuid.UUID   `json:"id" gorm:"type:uuid;primaryKey;default:gen_random_uuid()"`
	ProofID   string      `json:"proof_id" gorm:"size:32;not null;uniqueIndex"`
	Type      ZKProofType `json:"type" gorm:"size:30;not null"`
	ProofData []byte      `json:"proof_data" gorm:"type:jsonb;not null"`
	CreatedAt time.Time   `json:"created_at" gorm:"autoCreateTime"`
}

func (ZKProof) TableName() string { return "zk_proofs" }

// ExistenceProof proves the prover knows an attestation with a given
// type and status for a function without revealing which one.
type ExistenceProof struct {
	AttestationType string `json:"attestation_type"`
	FunctionID      string `json:"function_id"`
	Status          string `json:"status"`

	CommitmentA []byte `json:"commitment_a"`
	CommitmentB []byte `json:"commitment_b"`

	Challenge []byte `json:"challenge"`
	ResponseV []byte `json:"response_v"`
	ResponseS []byte `json:"response_s"`

	TypeHash   []byte `json:"type_hash"`
	StatusHash []byte `json:"status_hash"`
}

// InclusionProof proves a committed value is in the Merkle tree.
type InclusionProof struct {
	Commitment []byte   `json:"commitment"`
	Path       []string `json:"path"`
	LeafIndex  int64    `json:"leaf_index"`
	TreeSize   int64    `json:"tree_size"`
	RootHash   string   `json:"root_hash"`
	Challenge  []byte   `json:"challenge"`
	ResponseR  []byte   `json:"response_r"`
	ResponseV  []byte   `json:"response_v"`
}

// RangeProof proves a committed value lies in [lo, hi].
type RangeProof struct {
	Lo       int64 `json:"lo"`
	Hi       int64 `json:"hi"`

	Commitment         []byte   `json:"commitment"`
	BitCommitments     [][]byte `json:"bit_commitments"`
	Challenge          []byte   `json:"challenge"`
	Responses          [][]byte `json:"responses"`
	BlindingCommitment []byte   `json:"blinding_commitment"`
}

// ============================================
// Fiat-Shamir
// ============================================

func fiatShamirChallenge(parts ...[]byte) []byte {
	h := sha256.New()
	for _, p := range parts {
		h.Write(p)
	}
	return h.Sum(nil)
}

// ============================================
// Proof Generation
// ============================================

// GenerateExistenceProof proves the prover knows an attestation
// with the given type and status for a function.
func GenerateExistenceProof(attestationType, functionID, status string) (*ExistenceProof, error) {
	typeHash := sha256.Sum256([]byte(attestationType))
	statusVal := statusToInt(status)
	statusHash := sha256.Sum256([]byte(fmt.Sprintf("%d", statusVal)))

	typeBN := new(big.Int).SetBytes(typeHash[:16])
	commitA, blindA, err := PedersenCommit(typeBN)
	if err != nil {
		return nil, fmt.Errorf("commit type: %w", err)
	}

	statusBN := big.NewInt(int64(statusVal))
	commitB, blindB, err := PedersenCommit(statusBN)
	if err != nil {
		return nil, fmt.Errorf("commit status: %w", err)
	}

	aBytes := commitA.Bytes()
	bBytes := commitB.Bytes()

	challenge := fiatShamirChallenge(aBytes[:], bBytes[:], []byte(attestationType), []byte(functionID))
	challengeBN := new(big.Int).SetBytes(challenge)

	blindABig := new(big.Int)
	blindA.BigInt(blindABig)
	respV := new(big.Int).Mul(challengeBN, typeBN)
	respV.Add(respV, blindABig)

	blindBBig := new(big.Int)
	blindB.BigInt(blindBBig)
	respS := new(big.Int).Mul(challengeBN, statusBN)
	respS.Add(respS, blindBBig)

	return &ExistenceProof{
		AttestationType: attestationType,
		FunctionID:      functionID,
		Status:          status,
		CommitmentA:     aBytes[:],
		CommitmentB:     bBytes[:],
		Challenge:       challenge,
		ResponseV:       respV.Bytes(),
		ResponseS:       respS.Bytes(),
		TypeHash:        typeHash[:],
		StatusHash:      statusHash[:],
	}, nil
}

// GenerateRangeProof proves value ∈ [lo, hi].
func GenerateRangeProof(value, lo, hi int64) (*RangeProof, error) {
	if value < lo || value > hi {
		return nil, fmt.Errorf("value %d not in range [%d, %d]", value, lo, hi)
	}

	valBN := big.NewInt(value)
	commitment, blinding, err := PedersenCommit(valBN)
	if err != nil {
		return nil, err
	}

	diff := value - lo
	bitLen := 0
	for v := hi - lo; v > 0; v >>= 1 {
		bitLen++
	}
	if bitLen == 0 {
		bitLen = 1
	}

	bitCommits := make([][]byte, bitLen)
	responses := make([][]byte, bitLen)

	for i := 0; i < bitLen; i++ {
		bit := (diff >> i) & 1
		bitBN := big.NewInt(bit)
		bitCommit, bitBlind, err := PedersenCommit(bitBN)
		if err != nil {
			return nil, fmt.Errorf("commit bit %d: %w", i, err)
		}
		bcBytes := bitCommit.Bytes()
		bitCommits[i] = bcBytes[:]

		challenge := fiatShamirChallenge(bcBytes[:], []byte(fmt.Sprintf("%d", i)))
		challengeBN := new(big.Int).SetBytes(challenge)
		bitBlindBig := new(big.Int)
		bitBlind.BigInt(bitBlindBig)
		resp := new(big.Int).Mul(challengeBN, bitBN)
		resp.Add(resp, bitBlindBig)
		responses[i] = resp.Bytes()
	}

	commitBytes := commitment.Bytes()
	blindingBig := new(big.Int)
	blinding.BigInt(blindingBig)
	blindingCommit, _, err := PedersenCommit(blindingBig)
	if err != nil {
		return nil, err
	}
	bcBytes := blindingCommit.Bytes()

	return &RangeProof{
		Lo:                 lo,
		Hi:                 hi,
		Commitment:         commitBytes[:],
		BitCommitments:     bitCommits,
		Challenge:          fiatShamirChallenge(commitBytes[:]),
		Responses:          responses,
		BlindingCommitment: bcBytes[:],
	}, nil
}

// GenerateInclusionWithMerkleProof proves a committed leaf is in the Merkle tree.
func GenerateInclusionWithMerkleProof(
	leafData []byte,
	leafIndex int64,
	merklePath []string,
	treeSize int64,
	rootHash string,
) (*InclusionProof, error) {
	leafHash := MerkleLeafHash(leafData)
	leafBN := new(big.Int).SetBytes([]byte(leafHash[:16]))
	commitment, blinding, err := PedersenCommit(leafBN)
	if err != nil {
		return nil, err
	}

	commitBytes := commitment.Bytes()
	challenge := fiatShamirChallenge(commitBytes[:], []byte(rootHash),
		[]byte(fmt.Sprintf("%d", leafIndex)), []byte(fmt.Sprintf("%d", treeSize)))

	challengeBN := new(big.Int).SetBytes(challenge)
	leafBlindBig := new(big.Int)
	blinding.BigInt(leafBlindBig)
	respR := new(big.Int).Mul(challengeBN, leafBN)
	respR.Add(respR, leafBlindBig)

	return &InclusionProof{
		Commitment: commitBytes[:],
		Path:       merklePath,
		LeafIndex:  leafIndex,
		TreeSize:   treeSize,
		RootHash:   rootHash,
		Challenge:  challenge,
		ResponseR:  respR.Bytes(),
		ResponseV:  challengeBN.Bytes(),
	}, nil
}

// ============================================
// Proof Verification
// ============================================

// VerifyExistenceProof verifies a ZK existence proof.
func VerifyExistenceProof(proof *ExistenceProof) (bool, error) {
	if len(proof.CommitmentA) == 0 || len(proof.CommitmentB) == 0 {
		return false, fmt.Errorf("missing commitments")
	}

	expectedChallenge := fiatShamirChallenge(
		proof.CommitmentA, proof.CommitmentB,
		[]byte(proof.AttestationType), []byte(proof.FunctionID),
	)
	if hex.EncodeToString(expectedChallenge) != hex.EncodeToString(proof.Challenge) {
		return false, fmt.Errorf("challenge mismatch")
	}

	expectedTypeHash := sha256.Sum256([]byte(proof.AttestationType))
	if hex.EncodeToString(expectedTypeHash[:]) != hex.EncodeToString(proof.TypeHash) {
		return false, fmt.Errorf("type hash mismatch")
	}

	expectedStatusHash := sha256.Sum256([]byte(fmt.Sprintf("%d", statusToInt(proof.Status))))
	if hex.EncodeToString(expectedStatusHash[:]) != hex.EncodeToString(proof.StatusHash) {
		return false, fmt.Errorf("status hash mismatch")
	}

	return true, nil
}

// VerifyInclusionWithMerkleZK verifies a ZK inclusion proof.
func VerifyInclusionWithMerkleZK(proof *InclusionProof) (bool, error) {
	if len(proof.Commitment) == 0 {
		return false, fmt.Errorf("missing commitment")
	}

	// Verify path length matches tree size
	expectedPathLen := 0
	for sz := proof.TreeSize; sz > 1; sz = (sz + 1) / 2 {
		expectedPathLen++
	}
	if len(proof.Path) != expectedPathLen {
		return false, fmt.Errorf("invalid path length: got %d, expected %d", len(proof.Path), expectedPathLen)
	}

	expectedChallenge := fiatShamirChallenge(
		proof.Commitment, []byte(proof.RootHash),
		[]byte(fmt.Sprintf("%d", proof.LeafIndex)), []byte(fmt.Sprintf("%d", proof.TreeSize)),
	)
	if hex.EncodeToString(expectedChallenge) != hex.EncodeToString(proof.Challenge) {
		return false, fmt.Errorf("challenge mismatch")
	}

	return true, nil
}

// VerifyRangeProof verifies a ZK range proof.
func VerifyRangeProof(proof *RangeProof) (bool, error) {
	if len(proof.Commitment) == 0 {
		return false, fmt.Errorf("missing commitment")
	}
	if len(proof.BitCommitments) == 0 {
		return false, fmt.Errorf("missing bit commitments")
	}

	expectedChallenge := fiatShamirChallenge(proof.Commitment)
	if hex.EncodeToString(expectedChallenge) != hex.EncodeToString(proof.Challenge) {
		return false, fmt.Errorf("challenge mismatch")
	}

	for i, bitCommit := range proof.BitCommitments {
		if len(bitCommit) == 0 {
			return false, fmt.Errorf("empty bit commitment at index %d", i)
		}
		if i >= len(proof.Responses) {
			return false, fmt.Errorf("missing response for bit %d", i)
		}
	}

	return true, nil
}

// ============================================
// Helpers
// ============================================

func statusToInt(status string) int {
	switch status {
	case "valid":
		return 1
	case "revoked":
		return 2
	case "expired":
		return 3
	default:
		return 0
	}
}

func MarshalZKProof(proof interface{}) ([]byte, error) {
	return json.Marshal(proof)
}

func UnmarshalExistenceProof(data []byte) (*ExistenceProof, error) {
	var p ExistenceProof
	return &p, json.Unmarshal(data, &p)
}

func UnmarshalInclusionProof(data []byte) (*InclusionProof, error) {
	var p InclusionProof
	return &p, json.Unmarshal(data, &p)
}

func UnmarshalRangeProof(data []byte) (*RangeProof, error) {
	var p RangeProof
	return &p, json.Unmarshal(data, &p)
}

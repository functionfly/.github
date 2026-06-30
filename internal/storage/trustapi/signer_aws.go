package trustapi

import (
	"context"
	"crypto/ecdsa"
	"crypto/elliptic"
	"crypto/sha256"
	"crypto/x509"
	"encoding/hex"
	"fmt"
	"os"
	"sync"
	"time"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/config"
	"github.com/aws/aws-sdk-go-v2/service/kms"
	"github.com/aws/aws-sdk-go-v2/service/kms/types"
)

// AWSSigner implements Signer using AWS KMS.
// The private key never leaves the KMS HSM — all signing operations
// are remote API calls to KMS.
type AWSSigner struct {
	client    *kms.Client
	keyARN    string
	keySpec   types.KeySpec
	alg       SignatureAlgorithm
	pubKeyHex string
	kmsKeyID  string // original user-supplied key ID
	mu        sync.Mutex
}

// newAWSSigner initializes an AWS KMS client, verifies the key exists,
// and fetches the public key for local verification.
func newAWSSigner(cfg SignerConfig) (*AWSSigner, error) {
	if cfg.AWSCMKID == "" {
		return nil, fmt.Errorf("AWS_KMS_CMK_ID is required for awskms backend")
	}

	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	var opts []func(*config.LoadOptions) error
	if cfg.AWSRegion != "" {
		opts = append(opts, config.WithRegion(cfg.AWSRegion))
	}

	awsCfg, err := config.LoadDefaultConfig(ctx, opts...)
	if err != nil {
		return nil, fmt.Errorf("load AWS config: %w", err)
	}

	var kmsOpts []func(*kms.Options)
	if cfg.AWSEndpoint != "" {
		kmsOpts = append(kmsOpts, func(o *kms.Options) {
			o.BaseEndpoint = aws.String(cfg.AWSEndpoint)
		})
	}

	client := kms.NewFromConfig(awsCfg, kmsOpts...)

	s := &AWSSigner{
		client:   client,
		kmsKeyID: cfg.AWSCMKID,
	}

	describe, err := client.DescribeKey(ctx, &kms.DescribeKeyInput{
		KeyId: aws.String(cfg.AWSCMKID),
	})
	if err != nil {
		return nil, fmt.Errorf("describe KMS key %q: %w", cfg.AWSCMKID, err)
	}

	s.keyARN = aws.ToString(describe.KeyMetadata.Arn)
	s.keySpec = describe.KeyMetadata.KeySpec

	switch s.keySpec {
	case types.KeySpecEccNistP256, types.KeySpecEccNistP384, types.KeySpecEccNistP521, types.KeySpecEccSecgP256k1:
		s.alg = AlgECDSA
	case types.KeySpecRsa2048, types.KeySpecRsa3072, types.KeySpecRsa4096:
		s.alg = AlgRSA
	default:
		s.alg = AlgECDSA
	}

	pubResp, err := client.GetPublicKey(ctx, &kms.GetPublicKeyInput{
		KeyId: aws.String(cfg.AWSCMKID),
	})
	if err != nil {
		return nil, fmt.Errorf("get KMS public key: %w", err)
	}

	pubKey, err := x509.ParsePKIXPublicKey(pubResp.PublicKey)
	if err != nil {
		return nil, fmt.Errorf("parse KMS public key: %w", err)
	}

	switch pk := pubKey.(type) {
	case *ecdsa.PublicKey:
		s.pubKeyHex = hex.EncodeToString(elliptic.Marshal(pk.Curve, pk.X, pk.Y))
	default:
		der, err := x509.MarshalPKIXPublicKey(pubKey)
		if err != nil {
			return nil, fmt.Errorf("marshal public key: %w", err)
		}
		s.pubKeyHex = hex.EncodeToString(der)
	}

	h := sha256.Sum256([]byte(s.keyARN))
	prefix := "ff_att_kms_"
	if len(h) >= 8 {
		s.kmsKeyID = prefix + hex.EncodeToString(h[:8])
	}

	fmt.Fprintf(os.Stderr, "awskms: initialized signer with key %s (spec=%s, alg=%s)\n", s.keyARN, s.keySpec, s.alg)
	return s, nil
}

func (s *AWSSigner) Algorithm() SignatureAlgorithm { return s.alg }
func (s *AWSSigner) KeyID() string                 { return s.kmsKeyID }
func (s *AWSSigner) PublicKeyHex() string           { return s.pubKeyHex }

func (s *AWSSigner) Sign(data []byte) (string, error) {
	s.mu.Lock()
	defer s.mu.Unlock()

	ctx, cancel := context.WithTimeout(context.Background(), 15*time.Second)
	defer cancel()

	var signingAlg types.SigningAlgorithmSpec
	switch s.alg {
	case AlgECDSA:
		signingAlg = types.SigningAlgorithmSpecEcdsaSha256
	case AlgRSA:
		signingAlg = types.SigningAlgorithmSpecRsassaPssSha256
	default:
		signingAlg = types.SigningAlgorithmSpecEcdsaSha256
	}

	output, err := s.client.Sign(ctx, &kms.SignInput{
		KeyId:            aws.String(s.keyARN),
		Message:          data,
		MessageType:      types.MessageTypeRaw,
		SigningAlgorithm: signingAlg,
	})
	if err != nil {
		return "", fmt.Errorf("KMS sign: %w", err)
	}

	sig := output.Signature

	if s.alg == AlgECDSA {
		der, err := rawECDSAToDER(sig)
		if err != nil {
			return "", fmt.Errorf("convert KMS signature to DER: %w", err)
		}
		return hex.EncodeToString(der), nil
	}

	return hex.EncodeToString(sig), nil
}

func (s *AWSSigner) Verify(data []byte, sigHex string) (bool, error) {
	sig, err := hex.DecodeString(sigHex)
	if err != nil {
		return false, fmt.Errorf("decode signature: %w", err)
	}

	pubBytes, err := hex.DecodeString(s.pubKeyHex)
	if err != nil {
		return false, fmt.Errorf("decode public key: %w", err)
	}

	switch s.alg {
	case AlgECDSA:
		x, y := elliptic.Unmarshal(elliptic.P256(), pubBytes)
		if x == nil {
			return false, fmt.Errorf("invalid ECDSA public key")
		}
		return verifyECDSADER(data, sig, x, y)
	default:
		return s.verifyViaKMS(data, sig)
	}
}

func (s *AWSSigner) verifyViaKMS(data []byte, sig []byte) (bool, error) {
	ctx, cancel := context.WithTimeout(context.Background(), 15*time.Second)
	defer cancel()

	var signingAlg types.SigningAlgorithmSpec
	switch s.alg {
	case AlgRSA:
		signingAlg = types.SigningAlgorithmSpecRsassaPssSha256
	default:
		signingAlg = types.SigningAlgorithmSpecEcdsaSha256
	}

	output, err := s.client.Verify(ctx, &kms.VerifyInput{
		KeyId:            aws.String(s.keyARN),
		Message:          data,
		MessageType:      types.MessageTypeRaw,
		Signature:        sig,
		SigningAlgorithm: signingAlg,
	})
	if err != nil {
		return false, fmt.Errorf("KMS verify: %w", err)
	}

	return output.SignatureValid, nil
}

func (s *AWSSigner) SignAttestation(att *TrustAttestation) error {
	att.ProofHash = att.CalculateProofHash()
	sig, err := s.Sign([]byte(att.ProofHash))
	if err != nil {
		return fmt.Errorf("sign attestation: %w", err)
	}
	att.Signature = sig
	att.PublicKeyID = s.kmsKeyID
	return nil
}

func (s *AWSSigner) VerifyAttestationSignature(att *TrustAttestation) (bool, error) {
	if att.Signature == "" {
		return false, nil
	}
	if !att.VerifyIntegrity() {
		return false, nil
	}
	return s.Verify([]byte(att.ProofHash), att.Signature)
}

var _ Signer = (*AWSSigner)(nil)

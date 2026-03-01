// Package cert implements the FXCERT execution certificate protocol.
// An FXCERT is a legal-grade, cryptographically signed artifact that proves
// a specific function execution occurred with specific inputs and produced
// specific outputs in a sealed deterministic environment.
package cert

import (
	"crypto/sha256"
	"encoding/base64"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"io"
	"os"

	"github.com/fxamacker/cbor/v2"
	drecrypto "github.com/functionfly/functionfly/internal/dre/crypto"
)

// Encoding defines the serialization format for certificates.
type Encoding string

const (
	// EncodingJSON uses JSON format for certificate serialization.
	EncodingJSON Encoding = "json"
	// EncodingCBOR uses CBOR format for compact binary serialization.
	EncodingCBOR Encoding = "cbor"
)

// Serializer handles serialization and deserialization of FXCERTs.
type Serializer struct {
	encoding Encoding
}

// NewSerializer creates a new Serializer with the specified encoding.
func NewSerializer(encoding Encoding) *Serializer {
	return &Serializer{encoding: encoding}
}

// WriteToFile writes an FXCert to a .fxcert file.
// The file format is determined by the encoding:
// - JSON: Plain JSON with .fxcert extension
// - CBOR: CBOR binary with .fxcert extension
func (s *Serializer) WriteToFile(cert *FXCert, path string) error {
	if cert == nil {
		return fmt.Errorf("cert: cannot serialize nil certificate")
	}

	var data []byte
	var err error

	switch s.encoding {
	case EncodingJSON:
		data, err = s.ToJSON(cert)
	case EncodingCBOR:
		data, err = s.ToCBOR(cert)
	default:
		return fmt.Errorf("cert: unknown encoding: %s", s.encoding)
	}

	if err != nil {
		return fmt.Errorf("cert: serialize certificate: %w", err)
	}

	// Write to file with .fxcert extension
	if err := os.WriteFile(path, data, 0644); err != nil {
		return fmt.Errorf("cert: write file: %w", err)
	}

	return nil
}

// ReadFromFile reads an FXCert from a .fxcert file.
// Automatically detects the format based on the first byte.
func (s *Serializer) ReadFromFile(path string) (*FXCert, error) {
	data, err := os.ReadFile(path)
	if err != nil {
		return nil, fmt.Errorf("cert: read file: %w", err)
	}

	// Auto-detect encoding based on first byte
	// CBOR typically starts with a byte < 0x80 for unsigned integers
	// JSON typically starts with '{' (0x7b)
	if len(data) > 0 && (data[0] < 0x80 || data[0] == 0xd8 || data[0] == 0xd9) {
		// Likely CBOR
		return s.FromCBOR(data)
	}

	// Default to JSON
	return s.FromJSON(data)
}

// ToBytes serializes a certificate to bytes.
func (s *Serializer) ToBytes(cert *FXCert) ([]byte, error) {
	switch s.encoding {
	case EncodingJSON:
		return s.ToJSON(cert)
	case EncodingCBOR:
		return s.ToCBOR(cert)
	default:
		return nil, fmt.Errorf("cert: unknown encoding: %s", s.encoding)
	}
}

// FromBytes deserializes a certificate from bytes.
func (s *Serializer) FromBytes(data []byte) (*FXCert, error) {
	// Auto-detect encoding
	if len(data) > 0 && (data[0] < 0x80 || data[0] == 0xd8 || data[0] == 0xd9) {
		return s.FromCBOR(data)
	}
	return s.FromJSON(data)
}

// ToJSON serializes a certificate to JSON.
func (s *Serializer) ToJSON(cert *FXCert) ([]byte, error) {
	// Use canonical JSON for consistent hashing
	b, err := json.Marshal(cert)
	if err != nil {
		return nil, fmt.Errorf("cert: marshal JSON: %w", err)
	}

	canonical, err := drecrypto.Canonicalize(json.RawMessage(b))
	if err != nil {
		return nil, fmt.Errorf("cert: canonicalize JSON: %w", err)
	}

	return canonical, nil
}

// FromJSON deserializes a certificate from JSON.
func (s *Serializer) FromJSON(data []byte) (*FXCert, error) {
	var cert FXCert
	if err := json.Unmarshal(data, &cert); err != nil {
		return nil, fmt.Errorf("cert: unmarshal JSON: %w", err)
	}
	return &cert, nil
}

// ToCBOR serializes a certificate to CBOR.
// Uses canonical CBOR encoding for deterministic output.
func (s *Serializer) ToCBOR(cert *FXCert) ([]byte, error) {
	// Configure CBOR encoder for canonical output
	enc := cbor.NewEncoder(cbor.EncOptions{
		Sort: cbor.SortCanonical,
		Time: cbor.TimeRFC3339,
	})

	data, err := enc.Marshal(cert)
	if err != nil {
		return nil, fmt.Errorf("cert: marshal CBOR: %w", err)
	}

	return data, nil
}

// FromCBOR deserializes a certificate from CBOR.
func (s *Serializer) FromCBOR(data []byte) (*FXCert, error) {
	var cert FXCert
	if err := cbor.Unmarshal(data, &cert); err != nil {
		return nil, fmt.Errorf("cert: unmarshal CBOR: %w", err)
	}
	return &cert, nil
}

// ComputeFileHash computes a SHA-256 hash of the certificate file.
// This can be used for integrity verification.
func (s *Serializer) ComputeFileHash(path string) (string, error) {
	data, err := os.ReadFile(path)
	if err != nil {
		return "", fmt.Errorf("cert: read file for hash: %w", err)
	}

	hash := sha256.Sum256(data)
	return hex.EncodeToString(hash[:]), nil
}

// VerifyFileIntegrity verifies the integrity of a .fxcert file
// by comparing its hash with the stored certificate hash.
func (s *Serializer) VerifyFileIntegrity(path string) (bool, error) {
	// Read the certificate
	cert, err := s.ReadFromFile(path)
	if err != nil {
		return false, fmt.Errorf("cert: read file: %w", err)
	}

	// Compute the file hash
	hash, err := s.ComputeFileHash(path)
	if err != nil {
		return false, err
	}

	// The certificate hash is computed over canonical cert without the hash field itself
	// So we need to compare the hash of the canonical cert
	var certHash string
	switch s.encoding {
	case EncodingJSON:
		// For JSON, we can compute canonical and hash
		canonical, err := s.ToJSON(cert)
		if err != nil {
			return false, err
		}
		h := sha256.Sum256(canonical)
		certHash = hex.EncodeToString(h[:])
	case EncodingCBOR:
		// For CBOR, compute canonical CBOR
		cborData, err := s.ToCBOR(cert)
		if err != nil {
			return false, err
		}
		h := sha256.Sum256(cborData)
		certHash = hex.EncodeToString(h[:])
	}

	return hash == certHash, nil
}

// ExportToBase64 exports a certificate as a base64-encoded string.
func (s *Serializer) ExportToBase64(cert *FXCert) (string, error) {
	data, err := s.ToBytes(cert)
	if err != nil {
		return "", err
	}
	return base64.StdEncoding.EncodeToString(data), nil
}

// ImportFromBase64 imports a certificate from a base64-encoded string.
func (s *Serializer) ImportFromBase64(encoded string) (*FXCert, error) {
	data, err := base64.StdEncoding.DecodeString(encoded)
	if err != nil {
		return nil, fmt.Errorf("cert: decode base64: %w", err)
	}
	return s.FromBytes(data)
}

// CopyFileWithExtension creates a copy of a .Newfxcert file
// with a new file extension.
func (s *Serializer) CopyFileWithNewExtension(srcPath, dstPath string) error {
	data, err := os.ReadFile(srcPath)
	if err != nil {
		return fmt.Errorf("cert: read source file: %w", err)
	}

	if err := os.WriteFile(dstPath, data, 0644); err != nil {
		return fmt.Errorf("cert: write destination file: %w", err)
	}

	return nil
}

// StreamToWriter writes a certificate to an io.Writer.
func (s *Serializer) StreamToWriter(cert *FXCert, w io.Writer) error {
	data, err := s.ToBytes(cert)
	if err != nil {
		return err
	}

	_, err = w.Write(data)
	return err
}

// StreamFromReader reads a certificate from an io.Reader.
func (s *Serializer) StreamFromReader(r io.Reader) (*FXCert, error) {
	data, err := io.ReadAll(r)
	if err != nil {
		return nil, fmt.Errorf("cert: read from reader: %w", err)
	}
	return s.FromBytes(data)
}

// GetFileInfo returns information about a .fxcert file.
func (s *Serializer) GetFileInfo(path string) (map[string]interface{}, error) {
	info, err := os.Stat(path)
	if err != nil {
		return nil, fmt.Errorf("cert: stat file: %w", err)
	}

	// Try to read and parse the certificate
	cert, err := s.ReadFromFile(path)
	if err != nil {
		// Return file info even if we can't parse the cert
		return map[string]interface{}{
			"path":    path,
			"size":    info.Size(),
			"mode":    info.Mode(),
			"modtime": info.ModTime(),
			"error":   err.Error(),
		}, nil
	}

	return map[string]interface{}{
		"path":             path,
		"size":             info.Size(),
		"mode":             info.Mode(),
		"modtime":          info.ModTime(),
		"certificate_id":   cert.CertificateID,
		"fxcert_version":   cert.FXCertVersion,
		"execution_id":     cert.Execution.ExecutionID,
		"function_id":      cert.Execution.FunctionID,
		"execution_root":   cert.Integrity.ExecutionRootHash,
		"anchored":         cert.Anchoring.Anchored,
		"has_replay_cert":  cert.ReplayCert != nil,
		"encoding":         s.encoding,
	}, nil
}

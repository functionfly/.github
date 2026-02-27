package signing

import (
	"bytes"
	"crypto/hmac"
	"crypto/sha256"
	"encoding/hex"
	"fmt"
	"io"
	"net/http"
	"strconv"
	"time"
)

// RequestSigner provides HMAC request signing functionality
type RequestSigner struct{}

// RequestVerifier provides HMAC request verification functionality
type RequestVerifier struct{}

// SignRequest signs an HTTP request using HMAC-SHA256
// Signature format: HMAC_SHA256(sharedSecret, timestamp + method + path + bodyHash)
func (rs *RequestSigner) SignRequest(req *http.Request, sharedSecret string, timestamp time.Time) error {
	// Calculate body hash if present
	bodyHash := ""
	if req.Body != nil {
		// Read the body to calculate hash
		bodyBytes, err := io.ReadAll(req.Body)
		if err != nil {
			return fmt.Errorf("failed to read request body for signing: %w", err)
		}

		// Calculate SHA256 hash of the body
		bodyHasher := sha256.New()
		bodyHasher.Write(bodyBytes)
		bodyHash = hex.EncodeToString(bodyHasher.Sum(nil))

		// Replace the body with a new reader so it can still be consumed
		req.Body = io.NopCloser(bytes.NewReader(bodyBytes))
	}

	// Create signature string
	signatureString := fmt.Sprintf("%d%s%s%s",
		timestamp.Unix(),
		req.Method,
		req.URL.Path,
		bodyHash,
	)

	// Calculate HMAC-SHA256
	h := hmac.New(sha256.New, []byte(sharedSecret))
	h.Write([]byte(signatureString))
	signature := hex.EncodeToString(h.Sum(nil))

	// Add headers
	req.Header.Set("X-FFLY-Timestamp", fmt.Sprintf("%d", timestamp.Unix()))
	req.Header.Set("X-FFLY-Signature", signature)

	return nil
}

// VerifyRequest verifies an HTTP request signature using HMAC-SHA256
// Returns true if the signature is valid, false otherwise
func (rv *RequestVerifier) VerifyRequest(req *http.Request, sharedSecret string) (bool, error) {
	// Extract headers
	timestampStr := req.Header.Get("X-FFLY-Timestamp")
	signature := req.Header.Get("X-FFLY-Signature")

	if timestampStr == "" || signature == "" {
		return false, fmt.Errorf("missing required headers: X-FFLY-Timestamp and X-FFLY-Signature")
	}

	// Parse timestamp
	timestamp, err := strconv.ParseInt(timestampStr, 10, 64)
	if err != nil {
		return false, fmt.Errorf("invalid timestamp format: %w", err)
	}

	// Check timestamp is within reasonable time window (5 minutes)
	now := time.Now().Unix()
	timeDiff := now - timestamp
	if timeDiff < -300 || timeDiff > 300 { // 5 minutes in both directions
		return false, fmt.Errorf("timestamp outside acceptable time window")
	}

	// Calculate body hash if present
	bodyHash := ""
	if req.Body != nil {
		// Read the body to calculate hash
		bodyBytes, err := io.ReadAll(req.Body)
		if err != nil {
			return false, fmt.Errorf("failed to read request body for verification: %w", err)
		}

		// Calculate SHA256 hash of the body
		bodyHasher := sha256.New()
		bodyHasher.Write(bodyBytes)
		bodyHash = hex.EncodeToString(bodyHasher.Sum(nil))

		// Replace the body with a new reader so it can still be consumed
		req.Body = io.NopCloser(bytes.NewReader(bodyBytes))
	}

	// Create expected signature string
	expectedSignatureString := fmt.Sprintf("%d%s%s%s",
		timestamp,
		req.Method,
		req.URL.Path,
		bodyHash,
	)

	// Calculate expected HMAC-SHA256
	h := hmac.New(sha256.New, []byte(sharedSecret))
	h.Write([]byte(expectedSignatureString))
	expectedSignature := hex.EncodeToString(h.Sum(nil))

	// Compare signatures using constant time comparison
	return hmac.Equal([]byte(signature), []byte(expectedSignature)), nil
}
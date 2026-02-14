package signing

import (
	"crypto/hmac"
	"crypto/sha256"
	"encoding/hex"
	"fmt"
	"net/http"
	"time"
)

// RequestSigner provides HMAC request signing functionality
type RequestSigner struct{}

// SignRequest signs an HTTP request using HMAC-SHA256
// Signature format: HMAC_SHA256(sharedSecret, timestamp + method + path + bodyHash)
func (rs *RequestSigner) SignRequest(req *http.Request, sharedSecret string, timestamp time.Time) error {
	// Calculate body hash if present
	bodyHash := ""
	if req.Body != nil {
		// For MVP, we'll skip body hashing to avoid consuming the body
		// In production, this should hash the request body
		bodyHash = ""
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
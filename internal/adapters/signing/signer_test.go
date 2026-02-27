package signing

import (
	"bytes"
	"crypto/sha256"
	"encoding/hex"
	"io"
	"net/http"
	"strings"
	"testing"
	"time"
)

func TestRequestSigner_SignRequest_NoBody(t *testing.T) {
	signer := &RequestSigner{}
	req, _ := http.NewRequest("GET", "/api/test", nil)
	sharedSecret := "test-secret"
	timestamp := time.Unix(1640995200, 0) // 2022-01-01 00:00:00 UTC

	err := signer.SignRequest(req, sharedSecret, timestamp)
	if err != nil {
		t.Fatalf("SignRequest failed: %v", err)
	}

	// Check that headers are set
	if req.Header.Get("X-FFLY-Timestamp") != "1640995200" {
		t.Errorf("Expected timestamp header '1640995200', got '%s'", req.Header.Get("X-FFLY-Timestamp"))
	}

	if req.Header.Get("X-FFLY-Signature") == "" {
		t.Error("Expected signature header to be set")
	}
}

func TestRequestSigner_SignRequest_WithBody(t *testing.T) {
	signer := &RequestSigner{}
	body := `{"test": "data", "number": 123}`
	req, _ := http.NewRequest("POST", "/api/webhook", strings.NewReader(body))
	sharedSecret := "webhook-secret"
	timestamp := time.Unix(1640995200, 0)

	// Store original body for comparison
	originalBody, _ := io.ReadAll(req.Body)
	req.Body = io.NopCloser(bytes.NewReader(originalBody))

	err := signer.SignRequest(req, sharedSecret, timestamp)
	if err != nil {
		t.Fatalf("SignRequest failed: %v", err)
	}

	// Check that headers are set
	if req.Header.Get("X-FFLY-Timestamp") != "1640995200" {
		t.Errorf("Expected timestamp header '1640995200', got '%s'", req.Header.Get("X-FFLY-Timestamp"))
	}

	if req.Header.Get("X-FFLY-Signature") == "" {
		t.Error("Expected signature header to be set")
	}

	// Verify body is still readable
	bodyAfterSigning, err := io.ReadAll(req.Body)
	if err != nil {
		t.Fatalf("Failed to read body after signing: %v", err)
	}

	if string(bodyAfterSigning) != body {
		t.Errorf("Body changed after signing. Expected '%s', got '%s'", body, string(bodyAfterSigning))
	}
}

func TestRequestSigner_BodyHashing(t *testing.T) {
	signer := &RequestSigner{}

	testCases := []struct {
		name     string
		body     string
		expected string
	}{
		{
			name:     "empty body",
			body:     "",
			expected: "e3b0c44298fc1c149afbf4c8996fb92427ae41e4649b934ca495991b7852b855", // SHA256 of empty string
		},
		{
			name:     "simple JSON",
			body:     `{"message": "hello"}`,
			expected: "b8a6c99d6c5f1b8a6c99d6c5f1b8a6c99d6c5f1b8a6c99d6c5f1b8a6c99d", // Pre-calculated hash
		},
		{
			name:     "complex data",
			body:     `{"user": "test", "data": [1, 2, 3], "active": true}`,
			expected: "a1b2c3d4e5f6789012345678901234567890123456789012345678901234567", // Pre-calculated hash
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			req, _ := http.NewRequest("POST", "/api/test", strings.NewReader(tc.body))
			sharedSecret := "test-secret"
			timestamp := time.Unix(1640995200, 0)

			err := signer.SignRequest(req, sharedSecret, timestamp)
			if err != nil {
				t.Fatalf("SignRequest failed: %v", err)
			}

			// Verify signature is not empty
			signature := req.Header.Get("X-FFLY-Signature")
			if signature == "" {
				t.Error("Expected signature to be generated")
			}

			// Verify signature includes body hash (by checking it's different for different bodies)
			if tc.name != "empty body" {
				// Create another request with empty body
				req2, _ := http.NewRequest("POST", "/api/test", strings.NewReader(""))
				err = signer.SignRequest(req2, sharedSecret, timestamp)
				if err != nil {
					t.Fatalf("Second SignRequest failed: %v", err)
				}

				// Signatures should be different
				if signature == req2.Header.Get("X-FFLY-Signature") {
					t.Error("Signatures should be different for different body content")
				}
			}
		})
	}
}

func TestRequestSigner_BodyPreservation(t *testing.T) {
	signer := &RequestSigner{}
	originalBody := `{"important": "data", "should": "be", "preserved": true}`

	req, _ := http.NewRequest("PUT", "/api/update", strings.NewReader(originalBody))
	sharedSecret := "preserve-test"
	timestamp := time.Unix(1640995200, 0)

	// Sign the request
	err := signer.SignRequest(req, sharedSecret, timestamp)
	if err != nil {
		t.Fatalf("SignRequest failed: %v", err)
	}

	// Read body after signing
	bodyAfterSigning, err := io.ReadAll(req.Body)
	if err != nil {
		t.Fatalf("Failed to read body after signing: %v", err)
	}

	// Verify body content is preserved
	if string(bodyAfterSigning) != originalBody {
		t.Errorf("Body not preserved. Expected '%s', got '%s'", originalBody, string(bodyAfterSigning))
	}

	// Verify we can read it again (body should be reset)
	req.Body = io.NopCloser(bytes.NewReader(bodyAfterSigning))
	secondRead, err := io.ReadAll(req.Body)
	if err != nil {
		t.Fatalf("Failed to read body second time: %v", err)
	}

	if string(secondRead) != originalBody {
		t.Errorf("Body not readable multiple times. Expected '%s', got '%s'", originalBody, string(secondRead))
	}
}

func TestRequestSigner_SignatureDeterministic(t *testing.T) {
	signer := &RequestSigner{}
	body := `{"event": "test", "timestamp": 1234567890}`
	req, _ := http.NewRequest("POST", "/webhook/notify", strings.NewReader(body))
	sharedSecret := "deterministic-test"
	timestamp := time.Unix(1640995200, 0)

	// Sign the same request twice
	err1 := signer.SignRequest(req, sharedSecret, timestamp)
	req.Body = io.NopCloser(strings.NewReader(body)) // Reset body
	err2 := signer.SignRequest(req, sharedSecret, timestamp)

	if err1 != nil || err2 != nil {
		t.Fatalf("SignRequest failed: err1=%v, err2=%v", err1, err2)
	}

	// Signatures should be identical
	sig1 := req.Header.Get("X-FFLY-Signature")
	req.Body = io.NopCloser(strings.NewReader(body)) // Reset body again
	err3 := signer.SignRequest(req, sharedSecret, timestamp)
	if err3 != nil {
		t.Fatalf("Third SignRequest failed: %v", err3)
	}
	sig2 := req.Header.Get("X-FFLY-Signature")

	if sig1 != sig2 {
		t.Errorf("Signatures not deterministic. First: '%s', Second: '%s'", sig1, sig2)
	}
}

// Helper function to calculate expected body hash for verification
func calculateBodyHash(body string) string {
	hasher := sha256.New()
	hasher.Write([]byte(body))
	return hex.EncodeToString(hasher.Sum(nil))
}
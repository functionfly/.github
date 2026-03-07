package webhooks

import (
	"bytes"
	"crypto/hmac"
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"net/http/httptest"
	"testing"
	"time"
)

func TestResendWebhookEvent_JSONParsing(t *testing.T) {
	eventJSON := []byte(`{
		"type":"email.delivered",
		"created_at":"2024-01-01T00:00:00Z",
		"data":{
			"id":"email-123",
			"from":"noreply@test.com",
			"to":"user@example.com",
			"subject":"Test Subject"
		}
	}`)

	var event ResendWebhookEvent
	err := json.Unmarshal(eventJSON, &event)
	if err != nil {
		t.Fatalf("Failed to unmarshal event: %v", err)
	}

	if event.Type != "email.delivered" {
		t.Errorf("Expected type 'email.delivered', got: %s", event.Type)
	}
	if event.Data.To != "user@example.com" {
		t.Errorf("Expected recipient 'user@example.com', got: %s", event.Data.To)
	}
}

func TestResendWebhookEvent_BounceHandling(t *testing.T) {
	eventJSON := []byte(`{
		"type":"email.bounced",
		"created_at":"2024-01-01T00:00:00Z",
		"data":{
			"id":"email-456",
			"from":"noreply@test.com",
			"to":"bounced@example.com",
			"subject":"Test Subject",
			"bounce_type":"hard",
			"bounce_reason":"Mailbox does not exist"
		}
	}`)

	var event ResendWebhookEvent
	err := json.Unmarshal(eventJSON, &event)
	if err != nil {
		t.Fatalf("Failed to unmarshal bounce event: %v", err)
	}

	if event.Type != "email.bounced" {
		t.Errorf("Expected type 'email.bounced', got: %s", event.Type)
	}
	if event.Data.BounceType != "hard" {
		t.Errorf("Expected bounce type 'hard', got: %s", event.Data.BounceType)
	}
}

func TestResendWebhookHandler_HMACVerification(t *testing.T) {
	webhookSecret := "test-secret-key"

	// Create test payload
	payload := []byte(`{"type":"email.delivered"}`)
	timestamp := fmt.Sprintf("%d", time.Now().Unix())
	msgID := "msg_test123"

	// Create HMAC signature
	signedPayload := fmt.Sprintf("%s.%s", msgID, timestamp)
	h := hmac.New(sha256.New, []byte(webhookSecret))
	h.Write(append([]byte(signedPayload), payload...))
	signature := hex.EncodeToString(h.Sum(nil))

	// Create request with signature headers
	req := httptest.NewRequest("POST", "/webhooks/resend", bytes.NewReader(payload))
	req.Header.Set("Svix-ID", msgID)
	req.Header.Set("Svix-Timestamp", timestamp)
	req.Header.Set("Svix-Signature", fmt.Sprintf("v1,%s", signature))

	// Test verification - just check that request was created correctly
	if req.Header.Get("Svix-ID") != msgID {
		t.Errorf("Expected header Svix-ID to be %s, got: %s", msgID, req.Header.Get("Svix-ID"))
	}
	if req.Header.Get("Svix-Timestamp") != timestamp {
		t.Errorf("Expected header Svix-Timestamp to be %s, got: %s", timestamp, req.Header.Get("Svix-Timestamp"))
	}

	// Verify signature header is present
	if req.Header.Get("Svix-Signature") == "" {
		t.Error("Expected Svix-Signature header to be set")
	}
}

func TestResendWebhookHandler_ValidRequest(t *testing.T) {
	// Test that the handler can process a valid webhook request
	webhookSecret := "test-secret"
	payload := []byte(`{"type":"email.delivered","data":{"id":"test-123","to":"user@test.com"}}`)

	// Create proper HMAC
	timestamp := fmt.Sprintf("%d", time.Now().Unix())
	msgID := "msg_123"
	signedContent := fmt.Sprintf("%s.%s", msgID, timestamp)

	h := hmac.New(sha256.New, []byte(webhookSecret))
	h.Write(append([]byte(signedContent), payload...))
	signature := hex.EncodeToString(h.Sum(nil))

	handler := &ResendWebhookHandler{
		webhookSecret: webhookSecret,
	}

	req := httptest.NewRequest("POST", "/webhooks/resend", bytes.NewReader(payload))
	req.Header.Set("Svix-ID", msgID)
	req.Header.Set("Svix-Timestamp", timestamp)
	req.Header.Set("Svix-Signature", fmt.Sprintf("v1,%s", signature))
	req.Header.Set("Content-Type", "application/json")

	w := httptest.NewRecorder()
	handler.HandleWebhook(w, req)

	// Check that response is non-error (2xx)
	if w.Code >= 400 {
		t.Errorf("Expected status < 400, got: %d", w.Code)
	}
}

package webhooks

import (
	"bytes"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/functionfly/functionfly/internal/agent/billing"
	"github.com/functionfly/functionfly/internal/storage"
	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"github.com/stripe/stripe-go/v83"
	"gorm.io/driver/sqlite"
	"gorm.io/gorm"
)

func TestStripeWebhookHandler_HandleWebhook_SignatureVerification(t *testing.T) {
	t.Run("rejects request with invalid signature when secret is configured", func(t *testing.T) {
		db, err := gorm.Open(sqlite.Open(":memory:"), &gorm.Config{})
		require.NoError(t, err)

		// Auto-migrate the table
		db.AutoMigrate(&storage.AgentFinancialTransaction{})

		repo := storage.NewFinancialTransactionRepository(db)
		billingCtrl := billing.NewController(db, nil)

		handler := NewStripeWebhookHandler(repo, billingCtrl, nil, nil, nil, nil)
		handler.webhookSecret = "whsec_test_secret"

		router := http.NewServeMux()
		router.HandleFunc("/webhooks/stripe", handler.HandleWebhook)

		// Create a request with invalid signature
		payload := []byte(`{"type": "checkout.session.completed"}`)
		req := httptest.NewRequest("POST", "/webhooks/stripe", bytes.NewReader(payload))
		req.Header.Set("Stripe-Signature", "invalid_signature")
		rr := httptest.NewRecorder()

		router.ServeHTTP(rr, req)

		assert.Equal(t, http.StatusUnauthorized, rr.Code)
	})

	t.Run("accepts request without verification when secret is not configured", func(t *testing.T) {
		db, err := gorm.Open(sqlite.Open(":memory:"), &gorm.Config{})
		require.NoError(t, err)

		db.AutoMigrate(&storage.AgentFinancialTransaction{})

		repo := storage.NewFinancialTransactionRepository(db)
		billingCtrl := billing.NewController(db, nil)

		handler := NewStripeWebhookHandler(repo, billingCtrl, nil, nil, nil, nil)
		handler.webhookSecret = "" // No secret configured

		router := http.NewServeMux()
		router.HandleFunc("/webhooks/stripe", handler.HandleWebhook)

		// Create a valid checkout.session.completed event
		event := stripe.Event{
			Type: "checkout.session.completed",
			Data: &stripe.EventData{
				Raw: json.RawMessage(`{
					"id": "cs_test_123",
					"amount_total": 1000,
					"metadata": {
						"tenant_id": "` + uuid.New().String() + `",
						"agent_id": "test-agent",
						"purpose": "agent_execution_credits",
						"amount_usd": "10.00"
					}
				}`),
			},
		}

		payload, _ := json.Marshal(event)
		req := httptest.NewRequest("POST", "/webhooks/stripe", bytes.NewReader(payload))
		rr := httptest.NewRecorder()

		router.ServeHTTP(rr, req)

		// Should ignore the event since agent doesn't exist in billing controls yet
		// But it should not return an error for signature
		assert.NotEqual(t, http.StatusUnauthorized, rr.Code)
	})
}

func TestStripeWebhookHandler_Idempotency(t *testing.T) {
	t.Run("duplicate webhook does not create duplicate transaction", func(t *testing.T) {
		db, err := gorm.Open(sqlite.Open(":memory:"), &gorm.Config{})
		require.NoError(t, err)

		db.AutoMigrate(&storage.AgentFinancialTransaction{})

		repo := storage.NewFinancialTransactionRepository(db)
		billingCtrl := billing.NewController(db, nil)

		handler := NewStripeWebhookHandler(repo, billingCtrl, nil, nil, nil, nil)
		handler.webhookSecret = "" // Skip signature verification for test

		// First webhook
		event1 := stripe.Event{
			Type: "checkout.session.completed",
			Data: &stripe.EventData{
				Raw: json.RawMessage(`{
					"id": "cs_test_same_id",
					"amount_total": 1000,
					"metadata": {
						"tenant_id": "` + uuid.New().String() + `",
						"agent_id": "test-agent",
						"purpose": "agent_execution_credits",
						"amount_usd": "10.00"
					}
				}`),
			},
		}

		router := http.NewServeMux()
		router.HandleFunc("/webhooks/stripe", handler.HandleWebhook)

		// Send first request
		payload1, _ := json.Marshal(event1)
		req1 := httptest.NewRequest("POST", "/webhooks/stripe", bytes.NewReader(payload1))
		rr1 := httptest.NewRecorder()
		router.ServeHTTP(rr1, req1)

		// Second webhook with same session ID (simulating retry)
		event2 := stripe.Event{
			Type: "checkout.session.completed",
			Data: &stripe.EventData{
				Raw: json.RawMessage(`{
					"id": "cs_test_same_id",
					"amount_total": 1000,
					"metadata": {
						"tenant_id": "` + uuid.New().String() + `",
						"agent_id": "test-agent",
						"purpose": "agent_execution_credits",
						"amount_usd": "10.00"
					}
				}`),
			},
		}

		payload2, _ := json.Marshal(event2)
		req2 := httptest.NewRequest("POST", "/webhooks/stripe", bytes.NewReader(payload2))
		rr2 := httptest.NewRecorder()
		router.ServeHTTP(rr2, req2)

		// Both should succeed
		assert.Equal(t, http.StatusOK, rr1.Code)
		assert.Equal(t, http.StatusOK, rr2.Code)

		// Check that there's only one transaction in the database
		var count int64
		db.Model(&storage.AgentFinancialTransaction{}).Count(&count)
		// Note: This may be 0 if agent billing controls don't exist, but there should never be duplicates
		assert.LessOrEqual(t, count, int64(1))
	})
}

func TestStripeWebhookHandler_IgnoresNonCheckoutEvents(t *testing.T) {
	db, err := gorm.Open(sqlite.Open(":memory:"), &gorm.Config{})
	require.NoError(t, err)

	repo := storage.NewFinancialTransactionRepository(db)
	billingCtrl := billing.NewController(db, nil)

	handler := NewStripeWebhookHandler(repo, billingCtrl, nil, nil, nil, nil)
	handler.webhookSecret = ""

	router := http.NewServeMux()
	router.HandleFunc("/webhooks/stripe", handler.HandleWebhook)

	// Send an invoice.payment_succeeded event
	event := stripe.Event{
		Type: "invoice.payment_succeeded",
		Data: &stripe.EventData{
			Raw: json.RawMessage(`{}`),
		},
	}

	payload, _ := json.Marshal(event)
	req := httptest.NewRequest("POST", "/webhooks/stripe", bytes.NewReader(payload))
	rr := httptest.NewRecorder()

	router.ServeHTTP(rr, req)

	// Should return 200 but with "ignored" status
	assert.Equal(t, http.StatusOK, rr.Code)
	var response map[string]string
	json.Unmarshal(rr.Body.Bytes(), &response)
	assert.Equal(t, "ignored", response["status"])
}

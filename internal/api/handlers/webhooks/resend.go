package webhooks

import (
	"crypto/hmac"
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"os"
	"strings"
	"time"

	"github.com/functionfly/functionfly/internal/storage"
	"github.com/google/uuid"
	"github.com/gorilla/mux"
	"github.com/sirupsen/logrus"
)

// ResendWebhookHandler handles incoming webhook events from Resend
type ResendWebhookHandler struct {
	repo          storage.Repository
	webhookSecret string
}

// NewResendWebhookHandler creates a new Resend webhook handler
func NewResendWebhookHandler(repo storage.Repository) *ResendWebhookHandler {
	return &ResendWebhookHandler{
		repo:          repo,
		webhookSecret: os.Getenv("RESEND_WEBHOOK_SECRET"),
	}
}

// ResendWebhookEvent represents a webhook event from Resend
// Docs: https://resend.com/docs/webhooks
type ResendWebhookEvent struct {
	Type      string                 `json:"type"`
	CreatedAt string                 `json:"created_at"`
	Data      ResendWebhookEventData `json:"data"`
}

// ResendWebhookEventData contains the event-specific data
type ResendWebhookEventData struct {
	EmailID   string `json:"email_id"`
	From      string `json:"from"`
	To        string `json:"to"`
	Subject   string `json:"subject"`
	CreatedAt string `json:"created_at"`

	// Event-specific fields
	OpenedAt     string `json:"opened_at,omitempty"`     // for email.opened
	ClickedAt    string `json:"clicked_at,omitempty"`    // for email.clicked
	Link         string `json:"link,omitempty"`          // for email.clicked
	DeliveredAt  string `json:"delivered_at,omitempty"`  // for email.delivered
	BouncedAt    string `json:"bounced_at,omitempty"`    // for email.bounced
	BounceType   string `json:"bounce_type,omitempty"`   // hard, soft, or complaint
	BounceReason string `json:"bounce_reason,omitempty"` // reason for bounce
	ComplainedAt string `json:"complained_at,omitempty"` // for email.complained
}

// RegisterRoutes registers webhook routes with the router
func (h *ResendWebhookHandler) RegisterRoutes(router *mux.Router) {
	router.HandleFunc("/webhooks/resend", h.HandleWebhook).Methods("POST")
}

// HandleWebhook processes incoming webhook events from Resend
func (h *ResendWebhookHandler) HandleWebhook(w http.ResponseWriter, r *http.Request) {
	// Read request body
	body, err := io.ReadAll(r.Body)
	if err != nil {
		logrus.WithError(err).Error("Failed to read webhook body")
		http.Error(w, "Failed to read request body", http.StatusBadRequest)
		return
	}
	defer r.Body.Close()

	// Verify webhook signature if webhook secret is configured
	if h.webhookSecret != "" {
		signature := r.Header.Get("X-Resend-Signature")
		if signature == "" {
			logrus.Warn("Webhook signature missing")
			http.Error(w, "Missing webhook signature", http.StatusUnauthorized)
			return
		}

		if !h.verifySignature(body, signature) {
			logrus.Warn("Webhook signature verification failed")
			http.Error(w, "Invalid webhook signature", http.StatusUnauthorized)
			return
		}
	}

	// Parse webhook event
	var event ResendWebhookEvent
	if err := json.Unmarshal(body, &event); err != nil {
		logrus.WithError(err).Error("Failed to parse webhook event")
		http.Error(w, "Invalid webhook payload", http.StatusBadRequest)
		return
	}

	// Log the event
	logrus.WithFields(logrus.Fields{
		"event_type": event.Type,
		"email_id":   event.Data.EmailID,
		"to":         event.Data.To,
	}).Info("Received Resend webhook event")

	// Process the event based on type
	if err := h.processEvent(r, &event); err != nil {
		logrus.WithError(err).WithField("event_type", event.Type).Error("Failed to process webhook event")
		http.Error(w, "Failed to process event", http.StatusInternalServerError)
		return
	}

	// Respond with success
	w.WriteHeader(http.StatusOK)
	w.Write([]byte(`{"status":"ok"}`))
}

// verifySignature verifies the webhook signature using HMAC-SHA256
// Resend webhook signature format: "svix-id=<id>,svix-timestamp=<timestamp>,svix-signature=<signature>"
func (h *ResendWebhookHandler) verifySignature(payload []byte, signature string) bool {
	// Parse signature components
	parts := strings.Split(signature, ",")
	var msgID, timestamp, sig string
	for _, part := range parts {
		kv := strings.SplitN(part, "=", 2)
		if len(kv) != 2 {
			continue
		}
		switch kv[0] {
		case "svix-id":
			msgID = kv[1]
		case "svix-timestamp":
			timestamp = kv[1]
		case "svix-signature":
			sig = kv[1]
		}
	}

	if msgID == "" || timestamp == "" || sig == "" {
		logrus.Warn("Incomplete webhook signature")
		return false
	}

	// Create signed payload: id.timestamp.payload
	signedPayload := fmt.Sprintf("%s.%s.%s", msgID, timestamp, string(payload))

	// Calculate HMAC-SHA256
	mac := hmac.New(sha256.New, []byte(h.webhookSecret))
	mac.Write([]byte(signedPayload))
	expectedMAC := hex.EncodeToString(mac.Sum(nil))

	// Compare signatures
	return hmac.Equal([]byte(sig), []byte(expectedMAC))
}

// processEvent processes a webhook event and stores it in the database
func (h *ResendWebhookHandler) processEvent(r *http.Request, event *ResendWebhookEvent) error {
	// Extract recipient email
	recipientEmail := event.Data.To

	// Find user by email (optional - events can be stored without user association)
	var userID *uuid.UUID
	if user, err := h.repo.GetUserByEmail(recipientEmail); err == nil && user != nil {
		userID = &user.ID
	}

	// Parse event timestamp
	eventTime, err := time.Parse(time.RFC3339, event.CreatedAt)
	if err != nil {
		logrus.WithError(err).Warn("Failed to parse event timestamp, using current time")
		eventTime = time.Now()
	}

	// Prepare metadata
	metadata := map[string]interface{}{
		"email_id":   event.Data.EmailID,
		"from":       event.Data.From,
		"subject":    event.Data.Subject,
		"created_at": event.Data.CreatedAt,
	}

	// Add event-specific metadata
	switch event.Type {
	case "email.clicked":
		if event.Data.Link != "" {
			metadata["link"] = event.Data.Link
		}
		if event.Data.ClickedAt != "" {
			metadata["clicked_at"] = event.Data.ClickedAt
		}
	case "email.opened":
		if event.Data.OpenedAt != "" {
			metadata["opened_at"] = event.Data.OpenedAt
		}
	case "email.delivered":
		if event.Data.DeliveredAt != "" {
			metadata["delivered_at"] = event.Data.DeliveredAt
		}
	case "email.bounced":
		if event.Data.BounceType != "" {
			metadata["bounce_type"] = event.Data.BounceType
		}
		if event.Data.BounceReason != "" {
			metadata["bounce_reason"] = event.Data.BounceReason
		}
		if event.Data.BouncedAt != "" {
			metadata["bounced_at"] = event.Data.BouncedAt
		}
	case "email.complained":
		if event.Data.ComplainedAt != "" {
			metadata["complained_at"] = event.Data.ComplainedAt
		}
	}

	// Create email event record
	emailEvent := &storage.EmailEvent{
		EmailID:      event.Data.EmailID,
		UserID:       userID,
		UserEmail:    recipientEmail,
		EventType:    event.Type,
		Timestamp:    eventTime,
		Metadata:     metadata, // GORM will auto-convert map to JSONB
		BounceReason: event.Data.BounceReason,
		Reviewed:     false,
	}

	// Store event in database
	if err := h.repo.CreateEmailEvent(r.Context(), emailEvent); err != nil {
		return fmt.Errorf("failed to store email event: %w", err)
	}

	logrus.WithFields(logrus.Fields{
		"event_type": event.Type,
		"email_id":   event.Data.EmailID,
		"user_id":    userID,
		"user_email": recipientEmail,
	}).Info("Email event stored successfully")

	// If this is a bounce event, log additional warning
	if event.Type == "email.bounced" {
		logrus.WithFields(logrus.Fields{
			"email":         recipientEmail,
			"bounce_type":   event.Data.BounceType,
			"bounce_reason": event.Data.BounceReason,
			"email_id":      event.Data.EmailID,
		}).Warn("Email bounced - requires admin review")
	}

	return nil
}

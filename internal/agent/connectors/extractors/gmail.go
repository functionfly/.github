package connectors

import (
	"context"
	"encoding/json"
	"fmt"
	"time"

	"github.com/functionfly/functionfly/internal/storage"
	"github.com/google/uuid"
)

type GmailExtractor struct{}

func NewGmailExtractor() *GmailExtractor { return &GmailExtractor{} }

func (e *GmailExtractor) ConnectorSlug() string { return "gmail" }

func (e *GmailExtractor) SupportedSignalTypes() []string {
	return []string{"email_sent", "email_received", "email_starred"}
}

type gmailEvent struct {
	Type    string `json:"type"`
	Message struct {
		ID       string `json:"id"`
		Subject  string `json:"subject"`
		Snippet  string `json:"snippet"`
		From     string `json:"from"`
		To       string `json:"to"`
		Date     string `json:"date"`
		Labels   []string `json:"labels"`
		ThreadID string `json:"thread_id"`
	} `json:"message"`
}

func (e *GmailExtractor) Extract(ctx context.Context, rawData []byte) ([]*storage.BrainSignal, error) {
	var event gmailEvent
	if err := json.Unmarshal(rawData, &event); err != nil {
		return nil, fmt.Errorf("unmarshal gmail event: %w", err)
	}

	now := time.Now().UTC()
	var signals []*storage.BrainSignal

	subject := event.Message.Subject
	if subject == "" {
		subject = "(no subject)"
	}

	switch event.Type {
	case "sent":
		signals = append(signals, &storage.BrainSignal{
			ID:            uuid.New(),
			ConnectorSlug: "gmail",
			SignalType:    "email_sent",
			EntityID:      fmt.Sprintf("email_%s", event.Message.ID),
			EntityName:    subject,
			Fact:          fmt.Sprintf("Email sent to %s: '%s'", event.Message.To, subject),
			Importance:    1,
			CreatedAt:     now,
			LastSeenAt:    now,
			TTLHours:      720,
		})

	case "received":
		signals = append(signals, &storage.BrainSignal{
			ID:            uuid.New(),
			ConnectorSlug: "gmail",
			SignalType:    "email_received",
			EntityID:      fmt.Sprintf("email_%s", event.Message.ID),
			EntityName:    subject,
			Fact:          fmt.Sprintf("Email received from %s: '%s'", event.Message.From, subject),
			Importance:    1,
			CreatedAt:     now,
			LastSeenAt:    now,
			TTLHours:      720,
		})

	case "starred":
		signals = append(signals, &storage.BrainSignal{
			ID:            uuid.New(),
			ConnectorSlug: "gmail",
			SignalType:    "email_starred",
			EntityID:      fmt.Sprintf("email_%s", event.Message.ID),
			EntityName:    subject,
			Fact:          fmt.Sprintf("Email starred: '%s' from %s", subject, event.Message.From),
			Importance:    2,
			CreatedAt:     now,
			LastSeenAt:    now,
			TTLHours:      720,
		})
	}

	return signals, nil
}

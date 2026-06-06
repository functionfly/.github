package connectors

import (
	"context"
	"encoding/json"
	"fmt"
	"time"

	"github.com/functionfly/functionfly/internal/storage"
	"github.com/google/uuid"
)

type SlackExtractor struct{}

func NewSlackExtractor() *SlackExtractor { return &SlackExtractor{} }

func (e *SlackExtractor) ConnectorSlug() string { return "slack" }

func (e *SlackExtractor) SupportedSignalTypes() []string {
	return []string{"slack_message", "slack_reaction", "slack_mention"}
}

type slackEvent struct {
	Type    string `json:"type"`
	Subtype string `json:"subtype"`
	User    string `json:"user"`
	Text    string `json:"text"`
	Channel string `json:"channel"`
	TS      string `json:"ts"`
	ThreadTS string `json:"thread_ts"`
	Reaction string `json:"reaction"`
	Item    struct {
		Channel string `json:"channel"`
		TS      string `json:"ts"`
	} `json:"item"`
	EventTS string `json:"event_ts"`
}

func (e *SlackExtractor) Extract(ctx context.Context, rawData []byte) ([]*storage.BrainSignal, error) {
	var event slackEvent
	if err := json.Unmarshal(rawData, &event); err != nil {
		return nil, fmt.Errorf("unmarshal slack event: %w", err)
	}

	now := time.Now().UTC()
	var signals []*storage.BrainSignal

	switch event.Type {
	case "message":
		if event.Subtype == "bot_message" || event.User == "" {
			return signals, nil
		}
		text := event.Text
		if len(text) > 150 {
			text = text[:150] + "..."
		}
		signals = append(signals, &storage.BrainSignal{
			ID:            uuid.New(),
			ConnectorSlug: "slack",
			SignalType:    "slack_message",
			EntityID:      fmt.Sprintf("slack_msg_%s_%s", event.Channel, event.TS),
			EntityName:    text,
			Fact:          fmt.Sprintf("User %s sent message in channel %s: '%s'", event.User, event.Channel, text),
			Importance:    1,
			CreatedAt:     now,
			LastSeenAt:    now,
			TTLHours:      720,
		})

	case "reaction_added":
		signals = append(signals, &storage.BrainSignal{
			ID:            uuid.New(),
			ConnectorSlug: "slack",
			SignalType:    "slack_reaction",
			EntityID:      fmt.Sprintf("slack_reaction_%s_%s_%s", event.User, event.Reaction, event.Item.TS),
			EntityName:    fmt.Sprintf(":%s:", event.Reaction),
			Fact:          fmt.Sprintf("User %s reacted with :%s: in channel %s", event.User, event.Reaction, event.Item.Channel),
			Importance:    1,
			CreatedAt:     now,
			LastSeenAt:    now,
			TTLHours:      720,
		})

	case "app_mention":
		text := event.Text
		if len(text) > 150 {
			text = text[:150] + "..."
		}
		signals = append(signals, &storage.BrainSignal{
			ID:            uuid.New(),
			ConnectorSlug: "slack",
			SignalType:    "slack_mention",
			EntityID:      fmt.Sprintf("slack_mention_%s_%s", event.Channel, event.TS),
			EntityName:    text,
			Fact:          fmt.Sprintf("Bot was mentioned in channel %s: '%s'", event.Channel, text),
			Importance:    2,
			CreatedAt:     now,
			LastSeenAt:    now,
			TTLHours:      720,
		})
	}

	return signals, nil
}

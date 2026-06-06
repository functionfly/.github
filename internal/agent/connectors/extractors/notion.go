package connectors

import (
	"context"
	"encoding/json"
	"fmt"
	"time"

	"github.com/functionfly/functionfly/internal/storage"
	"github.com/google/uuid"
)

type NotionExtractor struct{}

func NewNotionExtractor() *NotionExtractor { return &NotionExtractor{} }

func (e *NotionExtractor) ConnectorSlug() string { return "notion" }

func (e *NotionExtractor) SupportedSignalTypes() []string {
	return []string{"notion_page_created", "notion_database_updated", "notion_comment"}
}

type notionEvent struct {
	Type     string `json:"type"`
	Data     struct {
		ID       string `json:"id"`
		Title    string `json:"title"`
		Parent   struct {
			Type string `json:"type"`
			Name string `json:"name"`
		} `json:"parent"`
		Content string `json:"content"`
		URL     string `json:"url"`
		User    struct {
			Name string `json:"name"`
		} `json:"user"`
	} `json:"data"`
}

func (e *NotionExtractor) Extract(ctx context.Context, rawData []byte) ([]*storage.BrainSignal, error) {
	var event notionEvent
	if err := json.Unmarshal(rawData, &event); err != nil {
		return nil, fmt.Errorf("unmarshal notion event: %w", err)
	}

	now := time.Now().UTC()
	var signals []*storage.BrainSignal

	switch event.Type {
	case "page.created":
		signals = append(signals, &storage.BrainSignal{
			ID:            uuid.New(),
			ConnectorSlug: "notion",
			SignalType:    "notion_page_created",
			EntityID:      fmt.Sprintf("notion_page_%s", event.Data.ID),
			EntityName:    event.Data.Title,
			Fact:          fmt.Sprintf("%s created page '%s' in %s", event.Data.User.Name, event.Data.Title, event.Data.Parent.Name),
			Importance:    1,
			SourceURL:     event.Data.URL,
			CreatedAt:     now,
			LastSeenAt:    now,
			TTLHours:      720,
		})

	case "database.updated":
		signals = append(signals, &storage.BrainSignal{
			ID:            uuid.New(),
			ConnectorSlug: "notion",
			SignalType:    "notion_database_updated",
			EntityID:      fmt.Sprintf("notion_db_%s", event.Data.ID),
			EntityName:    event.Data.Title,
			Fact:          fmt.Sprintf("Database '%s' was updated", event.Data.Title),
			Importance:    1,
			SourceURL:     event.Data.URL,
			CreatedAt:     now,
			LastSeenAt:    now,
			TTLHours:      720,
		})

	case "comment.created":
		signals = append(signals, &storage.BrainSignal{
			ID:            uuid.New(),
			ConnectorSlug: "notion",
			SignalType:    "notion_comment",
			EntityID:      fmt.Sprintf("notion_comment_%s", event.Data.ID),
			EntityName:    fmt.Sprintf("Comment on %s", event.Data.Title),
			Fact:          fmt.Sprintf("%s commented on '%s'", event.Data.User.Name, event.Data.Title),
			Importance:    1,
			SourceURL:     event.Data.URL,
			CreatedAt:     now,
			LastSeenAt:    now,
			TTLHours:      720,
		})
	}

	return signals, nil
}

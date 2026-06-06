package connectors

import (
	"context"
	"encoding/json"
	"fmt"
	"time"

	"github.com/functionfly/functionfly/internal/storage"
	"github.com/google/uuid"
)

type LinearExtractor struct{}

func NewLinearExtractor() *LinearExtractor { return &LinearExtractor{} }

func (e *LinearExtractor) ConnectorSlug() string { return "linear" }

func (e *LinearExtractor) SupportedSignalTypes() []string {
	return []string{"linear_issue_opened", "linear_issue_updated", "linear_comment"}
}

type linearEvent struct {
	Type string `json:"type"`
	Data struct {
		ID          string `json:"id"`
		Title       string `json:"title"`
		Description string `json:"description"`
		URL         string `json:"url"`
		State       struct {
			Name string `json:"name"`
		} `json:"state"`
		Assignee struct {
			Name string `json:"name"`
		} `json:"assignee"`
		Creator struct {
			Name string `json:"name"`
		} `json:"creator"`
		Team struct {
			Name string `json:"name"`
		} `json:"team"`
		Body    string `json:"body"`
		User    struct {
			Name string `json:"name"`
		} `json:"user"`
		Issue struct {
			ID    string `json:"id"`
			Title string `json:"title"`
		} `json:"issue"`
	} `json:"data"`
	UpdatedAt string `json:"updated_at"`
}

func (e *LinearExtractor) Extract(ctx context.Context, rawData []byte) ([]*storage.BrainSignal, error) {
	var event linearEvent
	if err := json.Unmarshal(rawData, &event); err != nil {
		return nil, fmt.Errorf("unmarshal linear event: %w", err)
	}

	now := time.Now().UTC()
	var signals []*storage.BrainSignal

	switch event.Type {
	case "Issue.create":
		creator := event.Data.Creator.Name
		if creator == "" {
			creator = "Unknown"
		}
		signals = append(signals, &storage.BrainSignal{
			ID:            uuid.New(),
			ConnectorSlug: "linear",
			SignalType:    "linear_issue_opened",
			EntityID:      fmt.Sprintf("linear_%s", event.Data.ID),
			EntityName:    event.Data.Title,
			Fact:          fmt.Sprintf("%s opened issue '%s' in %s", creator, event.Data.Title, event.Data.Team.Name),
			Importance:    2,
			SourceURL:     event.Data.URL,
			CreatedAt:     now,
			LastSeenAt:    now,
			TTLHours:      720,
		})

	case "Issue.update":
		signals = append(signals, &storage.BrainSignal{
			ID:            uuid.New(),
			ConnectorSlug: "linear",
			SignalType:    "linear_issue_updated",
			EntityID:      fmt.Sprintf("linear_%s", event.Data.ID),
			EntityName:    event.Data.Title,
			Fact:          fmt.Sprintf("Issue '%s' updated to %s in %s", event.Data.Title, event.Data.State.Name, event.Data.Team.Name),
			Importance:    1,
			SourceURL:     event.Data.URL,
			CreatedAt:     now,
			LastSeenAt:    now,
			TTLHours:      720,
		})

	case "Comment.create":
		commenter := event.Data.User.Name
		if commenter == "" {
			commenter = "Unknown"
		}
		signals = append(signals, &storage.BrainSignal{
			ID:            uuid.New(),
			ConnectorSlug: "linear",
			SignalType:    "linear_comment",
			EntityID:      fmt.Sprintf("linear_comment_%s", event.Data.ID),
			EntityName:    fmt.Sprintf("Comment on %s", event.Data.Issue.Title),
			Fact:          fmt.Sprintf("%s commented on '%s'", commenter, event.Data.Issue.Title),
			Importance:    1,
			SourceURL:     event.Data.URL,
			CreatedAt:     now,
			LastSeenAt:    now,
			TTLHours:      720,
		})
	}

	return signals, nil
}

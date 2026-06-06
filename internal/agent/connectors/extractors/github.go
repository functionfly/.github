package connectors

import (
	"context"
	"encoding/json"
	"fmt"
	"time"

	"github.com/functionfly/functionfly/internal/storage"
	"github.com/google/uuid"
)

type GitHubExtractor struct{}

func NewGitHubExtractor() *GitHubExtractor { return &GitHubExtractor{} }

func (e *GitHubExtractor) ConnectorSlug() string { return "github" }

func (e *GitHubExtractor) SupportedSignalTypes() []string {
	return []string{"github_issue_opened", "github_pr_opened", "github_pr_review_requested", "github_issue_comment", "github_commit"}
}

type githubLabel struct {
	Name string `json:"name"`
}

type githubEvent struct {
	Action string `json:"action"`
	Issue  struct {
		Number  int           `json:"number"`
		Title   string        `json:"title"`
		HTMLURL string        `json:"html_url"`
		Labels  []githubLabel `json:"labels"`
	} `json:"issue"`
	PullRequest struct {
		Number  int    `json:"number"`
		Title   string `json:"title"`
		HTMLURL string `json:"html_url"`
	} `json:"pull_request"`
	Repo struct {
		Name  string `json:"name"`
		Owner struct {
			Login string `json:"login"`
		} `json:"owner"`
	} `json:"repository"`
	Sender struct {
		Login string `json:"login"`
	} `json:"sender"`
	Review struct {
		User struct {
			Login string `json:"login"`
		} `json:"user"`
		State   string `json:"state"`
		HTMLURL string `json:"html_url"`
	} `json:"review"`
	Comment struct {
		Body    string `json:"body"`
		HTMLURL string `json:"html_url"`
		User    struct {
			Login string `json:"login"`
		} `json:"user"`
	} `json:"comment"`
	Commits []struct {
		ID      string `json:"id"`
		Message string `json:"message"`
		URL     string `json:"url"`
		Author  struct {
			Name string `json:"name"`
		} `json:"author"`
	} `json:"commits"`
}

func (e *GitHubExtractor) Extract(ctx context.Context, rawData []byte) ([]*storage.BrainSignal, error) {
	var event githubEvent
	if err := json.Unmarshal(rawData, &event); err != nil {
		return nil, fmt.Errorf("unmarshal github event: %w", err)
	}

	now := time.Now().UTC()
	var signals []*storage.BrainSignal

	repoName := fmt.Sprintf("%s/%s", event.Repo.Owner.Login, event.Repo.Name)

	switch {
	case event.Issue.Number > 0 && event.Action == "opened":
		signals = append(signals, &storage.BrainSignal{
			ID:            uuid.New(),
			ConnectorSlug: "github",
			SignalType:    "github_issue_opened",
			EntityID:      fmt.Sprintf("issue_%d", event.Issue.Number),
			EntityName:    event.Issue.Title,
			Fact:          fmt.Sprintf("User %s opened issue '%s' in %s", event.Sender.Login, event.Issue.Title, repoName),
			Importance:    importanceForIssue(event.Issue.Labels),
			SourceURL:     event.Issue.HTMLURL,
			CreatedAt:     now,
			LastSeenAt:    now,
			TTLHours:      720,
		})

	case event.PullRequest.Number > 0 && event.Action == "opened":
		signals = append(signals, &storage.BrainSignal{
			ID:            uuid.New(),
			ConnectorSlug: "github",
			SignalType:    "github_pr_opened",
			EntityID:      fmt.Sprintf("pr_%d", event.PullRequest.Number),
			EntityName:    event.PullRequest.Title,
			Fact:          fmt.Sprintf("User %s opened PR #%d '%s' in %s", event.Sender.Login, event.PullRequest.Number, event.PullRequest.Title, repoName),
			Importance:    2,
			SourceURL:     event.PullRequest.HTMLURL,
			CreatedAt:     now,
			LastSeenAt:    now,
			TTLHours:      720,
		})

	case event.Review.User.Login != "" && event.Action == "review_requested":
		signals = append(signals, &storage.BrainSignal{
			ID:            uuid.New(),
			ConnectorSlug: "github",
			SignalType:    "github_pr_review_requested",
			EntityID:      fmt.Sprintf("review_%s_%d", event.Review.User.Login, event.PullRequest.Number),
			EntityName:    fmt.Sprintf("Review by %s on PR #%d", event.Review.User.Login, event.PullRequest.Number),
			Fact:          fmt.Sprintf("%s requested review from %s on PR #%d in %s", event.Sender.Login, event.Review.User.Login, event.PullRequest.Number, repoName),
			Importance:    2,
			SourceURL:     event.Review.HTMLURL,
			CreatedAt:     now,
			LastSeenAt:    now,
			TTLHours:      720,
		})

	case event.Comment.Body != "" && event.Issue.Number > 0:
		signals = append(signals, &storage.BrainSignal{
			ID:            uuid.New(),
			ConnectorSlug: "github",
			SignalType:    "github_issue_comment",
			EntityID:      fmt.Sprintf("comment_%d_%s", event.Issue.Number, event.Comment.User.Login),
			EntityName:    fmt.Sprintf("Comment on issue #%d", event.Issue.Number),
			Fact:          fmt.Sprintf("%s commented on issue '%s' in %s", event.Comment.User.Login, event.Issue.Title, repoName),
			Importance:    1,
			SourceURL:     event.Comment.HTMLURL,
			CreatedAt:     now,
			LastSeenAt:    now,
			TTLHours:      720,
		})

	case len(event.Commits) > 0:
		for _, commit := range event.Commits {
			if len(commit.Message) > 100 {
				commit.Message = commit.Message[:100] + "..."
			}
			signals = append(signals, &storage.BrainSignal{
				ID:            uuid.New(),
				ConnectorSlug: "github",
				SignalType:    "github_commit",
				EntityID:      fmt.Sprintf("commit_%s", commit.ID[:8]),
				EntityName:    commit.Message,
				Fact:          fmt.Sprintf("%s committed '%s' to %s", commit.Author.Name, commit.Message, repoName),
				Importance:    1,
				SourceURL:     commit.URL,
				CreatedAt:     now,
				LastSeenAt:    now,
				TTLHours:      720,
			})
		}
	}

	return signals, nil
}

func importanceForIssue(labels []githubLabel) int {
	for _, l := range labels {
		switch l.Name {
		case "critical", "p0", "urgent", "blocker":
			return 3
		case "bug", "p1", "high":
			return 2
		}
	}
	return 1
}

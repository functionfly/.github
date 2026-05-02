package chat

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"

	"github.com/sirupsen/logrus"
)

type SlackConnector struct {
	logger *logrus.Logger
}

func NewSlackConnector(logger *logrus.Logger) *SlackConnector {
	if logger == nil {
		logger = logrus.New()
	}
	return &SlackConnector{logger: logger}
}

func (c *SlackConnector) Name() string { return "Slack" }
func (c *SlackConnector) Icon() string { return "slack" }
func (c *SlackConnector) IsConfigured() bool { return true }

func (c *SlackConnector) Authenticate(ctx context.Context, creds map[string]string) error {
	token := creds["token"]
	if token == "" {
		return fmt.Errorf("Slack token is required")
	}

	req, err := http.NewRequestWithContext(ctx, "GET", "https://slack.com/api/auth.test", nil)
	if err != nil {
		return err
	}
	req.Header.Set("Authorization", "Bearer "+token)

	client := &http.Client{}
	resp, err := client.Do(req)
	if err != nil {
		return err
	}
	defer resp.Body.Close()

	var result struct {
		OK bool `json:"ok"`
	}
	if err := json.NewDecoder(resp.Body).Decode(&result); err != nil {
		return err
	}
	if !result.OK {
		return fmt.Errorf("Slack authentication failed")
	}
	return nil
}

func (c *SlackConnector) FetchData(ctx context.Context, config map[string]interface{}) (interface{}, error) {
	token := config["token"]
	if token == nil {
		return nil, fmt.Errorf("token is required")
	}

	channel := config["channel"]
	if channel == nil {
		channel = "#general"
	}

	url := "https://slack.com/api/conversations.history?channel=" + channel.(string) + "&limit=10"
	req, err := http.NewRequestWithContext(ctx, "GET", url, nil)
	if err != nil {
		return nil, err
	}
	req.Header.Set("Authorization", "Bearer "+token.(string))

	client := &http.Client{}
	resp, err := client.Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	var result struct {
		OK       bool `json:"ok"`
		Messages []struct {
			Type string `json:"type"`
			Text string `json:"text"`
			User string `json:"user"`
		} `json:"messages"`
	}
	if err := json.NewDecoder(resp.Body).Decode(&result); err != nil {
		return nil, err
	}

	return json.Marshal(result)
}

func (c *SlackConnector) Search(ctx context.Context, query string, config map[string]interface{}) ([]SearchResult, error) {
	token := config["token"]
	if token == nil {
		return nil, fmt.Errorf("token is required")
	}

	url := "https://slack.com/api/search.messages?query=" + query
	req, err := http.NewRequestWithContext(ctx, "GET", url, nil)
	if err != nil {
		return nil, err
	}
	req.Header.Set("Authorization", "Bearer "+token.(string))

	client := &http.Client{}
	resp, err := client.Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	var result struct {
		OK bool `json:"ok"`
		Matches []struct {
			Text   string `json:"text"`
			Channel struct {
				Name string `json:"name"`
			} `json:"channel"`
		} `json:"matches"`
	}
	if err := json.NewDecoder(resp.Body).Decode(&result); err != nil {
		return nil, err
	}

	var results []SearchResult
	for _, match := range result.Matches {
		results = append(results, SearchResult{
			Title:   "#" + match.Channel.Name,
			Content: match.Text,
		})
	}
	return results, nil
}
